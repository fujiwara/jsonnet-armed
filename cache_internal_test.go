package armed

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestMemoryCacheGetWithStale(t *testing.T) {
	tests := []struct {
		name        string
		ttl         time.Duration
		staleTTL    time.Duration
		age         time.Duration // age of the stored entry; negative means no entry
		wantContent string
		wantStale   bool
		wantExists  bool
		wantKept    bool // entry still stored after the call
	}{
		{
			name:       "miss when no entry",
			ttl:        time.Minute,
			age:        -1,
			wantExists: false,
		},
		{
			name:        "fresh within ttl",
			ttl:         time.Minute,
			age:         time.Second,
			wantContent: "result",
			wantStale:   false,
			wantExists:  true,
			wantKept:    true,
		},
		{
			name:        "stale after ttl within staleTTL",
			ttl:         time.Second,
			staleTTL:    time.Minute,
			age:         30 * time.Second,
			wantContent: "result",
			wantStale:   true,
			wantExists:  true,
			wantKept:    true,
		},
		{
			name:       "expired beyond staleTTL",
			ttl:        time.Second,
			staleTTL:   2 * time.Second,
			age:        3 * time.Second,
			wantExists: false,
			wantKept:   false,
		},
		{
			name:       "expired without staleTTL",
			ttl:        time.Second,
			staleTTL:   0,
			age:        2 * time.Second,
			wantExists: false,
			wantKept:   false,
		},
		{
			name:       "disabled when ttl is zero",
			ttl:        0,
			age:        time.Second,
			wantExists: false,
			wantKept:   true, // not removed, just ignored
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newMemoryCache(tt.ttl, tt.staleTTL)
			if tt.age >= 0 {
				c.entries["key"] = memoryCacheEntry{
					content:  "result",
					storedAt: time.Now().Add(-tt.age),
				}
			}

			content, isStale, exists := c.GetWithStale("key")
			if exists != tt.wantExists {
				t.Errorf("exists: got %v, want %v", exists, tt.wantExists)
			}
			if isStale != tt.wantStale {
				t.Errorf("isStale: got %v, want %v", isStale, tt.wantStale)
			}
			if content != tt.wantContent {
				t.Errorf("content: got %q, want %q", content, tt.wantContent)
			}
			if tt.age >= 0 {
				_, kept := c.entries["key"]
				if kept != tt.wantKept {
					t.Errorf("entry kept: got %v, want %v", kept, tt.wantKept)
				}
			}
		})
	}
}

func TestMemoryCacheSet(t *testing.T) {
	t.Run("stores and overwrites", func(t *testing.T) {
		c := newMemoryCache(time.Minute, 0)
		c.Set("key", "v1")
		if content, _, exists := c.GetWithStale("key"); !exists || content != "v1" {
			t.Errorf("got (%q, %v), want (v1, true)", content, exists)
		}
		c.Set("key", "v2")
		if content, _, exists := c.GetWithStale("key"); !exists || content != "v2" {
			t.Errorf("got (%q, %v), want (v2, true)", content, exists)
		}
	})

	t.Run("no-op when ttl is zero", func(t *testing.T) {
		c := newMemoryCache(0, 0)
		c.Set("key", "v1")
		if len(c.entries) != 0 {
			t.Errorf("entries: got %d, want 0", len(c.entries))
		}
	})
}

func TestMemoryCacheClean(t *testing.T) {
	tests := []struct {
		name     string
		ttl      time.Duration
		staleTTL time.Duration
		ages     map[string]time.Duration
		wantKeys []string
	}{
		{
			name:     "removes only entries beyond max(ttl, staleTTL)",
			ttl:      time.Minute,
			staleTTL: 5 * time.Minute,
			ages: map[string]time.Duration{
				"fresh":   time.Second,
				"stale":   2 * time.Minute,
				"expired": 6 * time.Minute,
			},
			wantKeys: []string{"fresh", "stale"},
		},
		{
			name:     "maxAge is ttl when staleTTL is zero",
			ttl:      time.Minute,
			staleTTL: 0,
			ages: map[string]time.Duration{
				"fresh":   time.Second,
				"expired": 2 * time.Minute,
			},
			wantKeys: []string{"fresh"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newMemoryCache(tt.ttl, tt.staleTTL)
			for key, age := range tt.ages {
				c.entries[key] = memoryCacheEntry{
					content:  key,
					storedAt: time.Now().Add(-age),
				}
			}

			c.Clean()

			if len(c.entries) != len(tt.wantKeys) {
				t.Errorf("entries: got %d, want %d", len(c.entries), len(tt.wantKeys))
			}
			for _, key := range tt.wantKeys {
				if _, ok := c.entries[key]; !ok {
					t.Errorf("entry %q should be kept", key)
				}
			}
		})
	}
}

func TestMemoryCacheMaxAge(t *testing.T) {
	tests := []struct {
		name     string
		ttl      time.Duration
		staleTTL time.Duration
		want     time.Duration
	}{
		{"ttl only", time.Minute, 0, time.Minute},
		{"staleTTL longer", time.Minute, 5 * time.Minute, 5 * time.Minute},
		{"ttl longer", 5 * time.Minute, time.Minute, 5 * time.Minute},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newMemoryCache(tt.ttl, tt.staleTTL)
			if got := c.maxAge(); got != tt.want {
				t.Errorf("maxAge: got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMemoryCacheConcurrent(t *testing.T) {
	c := newMemoryCache(time.Minute, 5*time.Minute)
	var wg sync.WaitGroup
	for i := range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			key := fmt.Sprintf("key%d", i%3)
			for range 100 {
				c.Set(key, "result")
				c.GetWithStale(key)
				c.Clean()
			}
		}()
	}
	wg.Wait()
}
