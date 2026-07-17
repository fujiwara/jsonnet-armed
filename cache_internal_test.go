package armed

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// cacheStoreHarness abstracts the two cacheStore implementations so that
// the same semantics tests run against both.
type cacheStoreHarness struct {
	name string
	new  func(t *testing.T, ttl, staleTTL time.Duration) cacheStore
	// seed stores an entry with the given age, bypassing Set (which is
	// a no-op when ttl == 0 and always stamps the current time)
	seed func(t *testing.T, s cacheStore, key, content string, age time.Duration)
	// exists reports whether an entry is physically stored
	exists func(t *testing.T, s cacheStore, key string) bool
}

func cacheStoreHarnesses() []cacheStoreHarness {
	return []cacheStoreHarness{
		{
			name: "memory",
			new: func(t *testing.T, ttl, staleTTL time.Duration) cacheStore {
				return newMemoryCache(ttl, staleTTL)
			},
			seed: func(t *testing.T, s cacheStore, key, content string, age time.Duration) {
				c := s.(*memoryCache)
				c.entries[key] = memoryCacheEntry{content: content, storedAt: time.Now().Add(-age)}
			},
			exists: func(t *testing.T, s cacheStore, key string) bool {
				c := s.(*memoryCache)
				_, ok := c.entries[key]
				return ok
			},
		},
		{
			name: "file",
			new: func(t *testing.T, ttl, staleTTL time.Duration) cacheStore {
				return &Cache{dir: t.TempDir(), ttl: ttl, staleTTL: staleTTL}
			},
			seed: func(t *testing.T, s cacheStore, key, content string, age time.Duration) {
				c := s.(*Cache)
				path := filepath.Join(c.dir, key+".json")
				if err := os.WriteFile(path, []byte(content), 0600); err != nil {
					t.Fatal(err)
				}
				storedAt := time.Now().Add(-age)
				if err := os.Chtimes(path, storedAt, storedAt); err != nil {
					t.Fatal(err)
				}
			},
			exists: func(t *testing.T, s cacheStore, key string) bool {
				c := s.(*Cache)
				_, err := os.Stat(filepath.Join(c.dir, key+".json"))
				return err == nil
			},
		},
	}
}

func TestCacheStoreGetWithStale(t *testing.T) {
	tests := []struct {
		name        string
		ttl         time.Duration
		staleTTL    time.Duration
		age         time.Duration // age of the seeded entry; negative means no entry
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
			age:        time.Minute,
			wantExists: false,
			wantKept:   false,
		},
		{
			name:       "expired without staleTTL",
			ttl:        time.Second,
			staleTTL:   0,
			age:        time.Minute,
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

	for _, h := range cacheStoreHarnesses() {
		for _, tt := range tests {
			t.Run(h.name+"/"+tt.name, func(t *testing.T) {
				s := h.new(t, tt.ttl, tt.staleTTL)
				if tt.age >= 0 {
					h.seed(t, s, "key", "result", tt.age)
				}

				content, isStale, exists := s.GetWithStale("key")
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
					if kept := h.exists(t, s, "key"); kept != tt.wantKept {
						t.Errorf("entry kept: got %v, want %v", kept, tt.wantKept)
					}
				}
			})
		}
	}
}

func TestCacheStoreSet(t *testing.T) {
	for _, h := range cacheStoreHarnesses() {
		t.Run(h.name+"/stores and overwrites", func(t *testing.T) {
			s := h.new(t, time.Minute, 0)
			if err := s.Set("key", "v1"); err != nil {
				t.Fatal(err)
			}
			if content, _, exists := s.GetWithStale("key"); !exists || content != "v1" {
				t.Errorf("got (%q, %v), want (v1, true)", content, exists)
			}
			if err := s.Set("key", "v2"); err != nil {
				t.Fatal(err)
			}
			if content, _, exists := s.GetWithStale("key"); !exists || content != "v2" {
				t.Errorf("got (%q, %v), want (v2, true)", content, exists)
			}
		})

		t.Run(h.name+"/no-op when ttl is zero", func(t *testing.T) {
			s := h.new(t, 0, 0)
			if err := s.Set("key", "v1"); err != nil {
				t.Fatal(err)
			}
			if h.exists(t, s, "key") {
				t.Error("entry must not be stored when ttl is zero")
			}
		})
	}
}

func TestCacheStoreClean(t *testing.T) {
	tests := []struct {
		name     string
		ttl      time.Duration
		staleTTL time.Duration
		ages     map[string]time.Duration
		wantKeys map[string]bool
	}{
		{
			name:     "removes only entries beyond max of ttl and staleTTL",
			ttl:      time.Minute,
			staleTTL: 5 * time.Minute,
			ages: map[string]time.Duration{
				"fresh":   time.Second,
				"stale":   2 * time.Minute,
				"expired": 6 * time.Minute,
			},
			wantKeys: map[string]bool{"fresh": true, "stale": true, "expired": false},
		},
		{
			name:     "maxAge is ttl when staleTTL is zero",
			ttl:      time.Minute,
			staleTTL: 0,
			ages: map[string]time.Duration{
				"fresh":   time.Second,
				"expired": 2 * time.Minute,
			},
			wantKeys: map[string]bool{"fresh": true, "expired": false},
		},
	}

	for _, h := range cacheStoreHarnesses() {
		for _, tt := range tests {
			t.Run(h.name+"/"+tt.name, func(t *testing.T) {
				s := h.new(t, tt.ttl, tt.staleTTL)
				for key, age := range tt.ages {
					h.seed(t, s, key, key, age)
				}

				if err := s.Clean(); err != nil {
					t.Fatal(err)
				}

				for key, want := range tt.wantKeys {
					if got := h.exists(t, s, key); got != want {
						t.Errorf("entry %q kept: got %v, want %v", key, got, want)
					}
				}
			})
		}
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
