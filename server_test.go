package armed_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	armed "github.com/fujiwara/jsonnet-armed"
	"github.com/google/go-cmp/cmp"
)

func TestServerHandler(t *testing.T) {
	tests := []struct {
		name            string
		serveCmd        *armed.ServeCmd
		method          string
		target          string
		wantStatus      int
		wantJSON        string // expected JSON body (compared structurally)
		wantContains    string // substring expected in the body
		wantNotContains string // substring that must not appear in the body
		wantAllow       string // expected Allow header
	}{
		{
			name:       "static file",
			method:     http.MethodGet,
			target:     "/static.jsonnet",
			wantStatus: http.StatusOK,
			wantJSON:   `{"ok": true}`,
		},
		{
			name:       "ext_str via query",
			method:     http.MethodGet,
			target:     "/hello.jsonnet?name=world",
			wantStatus: http.StatusOK,
			wantJSON:   `{"message": "hello, world"}`,
		},
		{
			name:       "numeric parameter parsed on the jsonnet side",
			method:     http.MethodGet,
			target:     "/limit.jsonnet?limit=10",
			wantStatus: http.StatusOK,
			wantJSON:   `{"limit": 10, "doubled": 20}`,
		},
		{
			name:       "native functions available",
			method:     http.MethodGet,
			target:     "/native.jsonnet",
			wantStatus: http.StatusOK,
			wantJSON:   `{"hash": "2d711642b726b04401627ca9fbac32f5c8530fb1903cc4db02258717921a4881"}`,
		},
		{
			name:       "nested path",
			method:     http.MethodGet,
			target:     "/sub/nested.jsonnet",
			wantStatus: http.StatusOK,
			wantJSON:   `{"nested": true}`,
		},
		{
			name: "server-wide default ext_str",
			serveCmd: &armed.ServeCmd{
				Dir:    "testdata/server",
				ExtStr: map[string]string{"name": "default"},
			},
			method:     http.MethodGet,
			target:     "/hello.jsonnet",
			wantStatus: http.StatusOK,
			wantJSON:   `{"message": "hello, default"}`,
		},
		{
			name: "query overrides server-wide default",
			serveCmd: &armed.ServeCmd{
				Dir:    "testdata/server",
				ExtStr: map[string]string{"name": "default"},
			},
			method:     http.MethodGet,
			target:     "/hello.jsonnet?name=query",
			wantStatus: http.StatusOK,
			wantJSON:   `{"message": "hello, query"}`,
		},
		{
			name:       "missing file",
			method:     http.MethodGet,
			target:     "/missing.jsonnet",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "non-jsonnet extension",
			method:     http.MethodGet,
			target:     "/static.json",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "root path",
			method:     http.MethodGet,
			target:     "/",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "directory path",
			method:     http.MethodGet,
			target:     "/sub",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "method not allowed",
			method:     http.MethodPost,
			target:     "/static.jsonnet",
			wantStatus: http.StatusMethodNotAllowed,
			wantAllow:  http.MethodGet,
		},
		{
			name:         "evaluation error",
			method:       http.MethodGet,
			target:       "/error.jsonnet",
			wantStatus:   http.StatusInternalServerError,
			wantContains: "boom",
		},
		{
			name:            "path traversal",
			method:          http.MethodGet,
			target:          "/../secret.jsonnet",
			wantStatus:      http.StatusNotFound,
			wantNotContains: "top-secret-value",
		},
		{
			name:            "path traversal percent-encoded",
			method:          http.MethodGet,
			target:          "/%2e%2e/secret.jsonnet",
			wantStatus:      http.StatusNotFound,
			wantNotContains: "top-secret-value",
		},
		{
			name:            "path traversal via subdirectory",
			method:          http.MethodGet,
			target:          "/sub/../../secret.jsonnet",
			wantStatus:      http.StatusNotFound,
			wantNotContains: "top-secret-value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.serveCmd
			if s == nil {
				s = &armed.ServeCmd{Dir: "testdata/server"}
			}
			req := httptest.NewRequest(tt.method, tt.target, nil)
			rec := httptest.NewRecorder()
			s.Handler().ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status: got %d, want %d (body: %s)", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
				t.Errorf("Content-Type: got %q, want %q", ct, "application/json")
			}
			if tt.wantAllow != "" {
				if allow := rec.Header().Get("Allow"); allow != tt.wantAllow {
					t.Errorf("Allow: got %q, want %q", allow, tt.wantAllow)
				}
			}
			body := rec.Body.String()
			if tt.wantJSON != "" {
				var got, want any
				if err := json.Unmarshal([]byte(body), &got); err != nil {
					t.Fatalf("failed to unmarshal body %q: %v", body, err)
				}
				if err := json.Unmarshal([]byte(tt.wantJSON), &want); err != nil {
					t.Fatalf("failed to unmarshal expected %q: %v", tt.wantJSON, err)
				}
				if diff := cmp.Diff(want, got); diff != "" {
					t.Errorf("body mismatch (-want +got):\n%s", diff)
				}
			}
			if tt.wantContains != "" && !strings.Contains(body, tt.wantContains) {
				t.Errorf("body %q does not contain %q", body, tt.wantContains)
			}
			if tt.wantNotContains != "" && strings.Contains(body, tt.wantNotContains) {
				t.Errorf("body %q must not contain %q", body, tt.wantNotContains)
			}
		})
	}
}

func TestServerSymlink(t *testing.T) {
	// A symlink pointing outside the served directory must be rejected,
	// while a symlink staying inside it is served normally.
	escape := filepath.Join("testdata", "server", "escape.jsonnet")
	if err := os.Symlink(filepath.Join("..", "secret.jsonnet"), escape); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}
	defer os.Remove(escape)
	inside := filepath.Join("testdata", "server", "alias.jsonnet")
	if err := os.Symlink("static.jsonnet", inside); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}
	defer os.Remove(inside)

	s := &armed.ServeCmd{Dir: "testdata/server"}

	req := httptest.NewRequest(http.MethodGet, "/escape.jsonnet", nil)
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("escape symlink: status got %d, want %d", rec.Code, http.StatusNotFound)
	}
	if strings.Contains(rec.Body.String(), "top-secret-value") {
		t.Errorf("escape symlink: body %q must not contain secret content", rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/alias.jsonnet", nil)
	rec = httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("inside symlink: status got %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestServerTimeout(t *testing.T) {
	s := &armed.ServeCmd{Dir: "testdata/server", Timeout: 100 * time.Millisecond}
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/sleep.jsonnet")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusGatewayTimeout {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("status: got %d, want %d (body: %s)", resp.StatusCode, http.StatusGatewayTimeout, body)
	}
}

func getWithCacheStatus(t *testing.T, url string) (int, string, string) {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return resp.StatusCode, string(body), resp.Header.Get("X-Cache")
}

func TestServerCacheHitMiss(t *testing.T) {
	s := &armed.ServeCmd{Dir: "testdata/server", Cache: 500 * time.Millisecond}
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	status, body1, cache1 := getWithCacheStatus(t, ts.URL+"/uuid.jsonnet")
	if status != http.StatusOK || cache1 != "MISS" {
		t.Errorf("first request: got status=%d X-Cache=%q, want 200/MISS", status, cache1)
	}

	status, body2, cache2 := getWithCacheStatus(t, ts.URL+"/uuid.jsonnet")
	if status != http.StatusOK || cache2 != "HIT" {
		t.Errorf("second request: got status=%d X-Cache=%q, want 200/HIT", status, cache2)
	}
	if body1 != body2 {
		t.Errorf("cached response differs: %q vs %q", body1, body2)
	}

	time.Sleep(600 * time.Millisecond) // beyond ttl (stale not set)
	status, body3, cache3 := getWithCacheStatus(t, ts.URL+"/uuid.jsonnet")
	if status != http.StatusOK || cache3 != "MISS" {
		t.Errorf("after expiry: got status=%d X-Cache=%q, want 200/MISS", status, cache3)
	}
	if body1 == body3 {
		t.Errorf("response after expiry should be re-evaluated, got same body %q", body3)
	}
}

func TestServerCacheStaleFallback(t *testing.T) {
	dir := t.TempDir()
	dataFile := filepath.Join(dir, "data.txt")
	if err := os.WriteFile(dataFile, []byte("v1"), 0644); err != nil {
		t.Fatal(err)
	}
	jsonnetFile := filepath.Join(dir, "stale.jsonnet")
	code := fmt.Sprintf("{ v: std.native('file_content')('%s') }", dataFile)
	if err := os.WriteFile(jsonnetFile, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	s := &armed.ServeCmd{Dir: dir, Cache: 200 * time.Millisecond, Stale: 2 * time.Second}
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	status, body1, cache1 := getWithCacheStatus(t, ts.URL+"/stale.jsonnet")
	if status != http.StatusOK || cache1 != "MISS" {
		t.Fatalf("first request: got status=%d X-Cache=%q (body: %s), want 200/MISS", status, cache1, body1)
	}

	// Remove the data file so that re-evaluation fails. The jsonnet file
	// itself is unchanged, so the cache key stays the same.
	if err := os.Remove(dataFile); err != nil {
		t.Fatal(err)
	}

	time.Sleep(300 * time.Millisecond) // beyond ttl, within stale
	status, body2, cache2 := getWithCacheStatus(t, ts.URL+"/stale.jsonnet")
	if status != http.StatusOK || cache2 != "STALE" {
		t.Errorf("stale request: got status=%d X-Cache=%q (body: %s), want 200/STALE", status, cache2, body2)
	}
	if !strings.Contains(body2, "v1") {
		t.Errorf("stale response %q should contain original value", body2)
	}

	time.Sleep(2 * time.Second) // beyond stale
	status, body3, _ := getWithCacheStatus(t, ts.URL+"/stale.jsonnet")
	if status != http.StatusInternalServerError {
		t.Errorf("after stale expiry: got status=%d (body: %s), want 500", status, body3)
	}
}

func TestServerCacheStaleOnTimeout(t *testing.T) {
	dir := t.TempDir()
	durFile := filepath.Join(dir, "dur.txt")
	if err := os.WriteFile(durFile, []byte("0"), 0644); err != nil {
		t.Fatal(err)
	}
	jsonnetFile := filepath.Join(dir, "slow.jsonnet")
	code := fmt.Sprintf("{ r: std.native('exec')('sleep', [std.native('file_content')('%s')]).exit_code }", durFile)
	if err := os.WriteFile(jsonnetFile, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	s := &armed.ServeCmd{
		Dir:     dir,
		Timeout: 300 * time.Millisecond,
		Cache:   100 * time.Millisecond,
		Stale:   5 * time.Second,
	}
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	status, body1, cache1 := getWithCacheStatus(t, ts.URL+"/slow.jsonnet")
	if status != http.StatusOK || cache1 != "MISS" {
		t.Fatalf("first request: got status=%d X-Cache=%q (body: %s), want 200/MISS", status, cache1, body1)
	}

	// Make evaluation exceed the timeout without changing the jsonnet file.
	if err := os.WriteFile(durFile, []byte("5"), 0644); err != nil {
		t.Fatal(err)
	}

	time.Sleep(150 * time.Millisecond) // beyond ttl, within stale
	status, body2, cache2 := getWithCacheStatus(t, ts.URL+"/slow.jsonnet")
	if status != http.StatusOK || cache2 != "STALE" {
		t.Errorf("timeout request: got status=%d X-Cache=%q (body: %s), want 200/STALE", status, cache2, body2)
	}
}

func TestServerCacheQueryParams(t *testing.T) {
	s := &armed.ServeCmd{Dir: "testdata/server", Cache: 5 * time.Second}
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	_, bodyA1, cacheA1 := getWithCacheStatus(t, ts.URL+"/uuid.jsonnet?a=1")
	_, bodyB, cacheB := getWithCacheStatus(t, ts.URL+"/uuid.jsonnet?a=2")
	if cacheA1 != "MISS" || cacheB != "MISS" {
		t.Errorf("different query params must be different entries: got %q, %q", cacheA1, cacheB)
	}
	if bodyA1 == bodyB {
		t.Errorf("different query params returned the same cached body %q", bodyA1)
	}

	_, bodyA2, cacheA2 := getWithCacheStatus(t, ts.URL+"/uuid.jsonnet?a=1")
	if cacheA2 != "HIT" || bodyA1 != bodyA2 {
		t.Errorf("same query params should hit: X-Cache=%q, bodies %q vs %q", cacheA2, bodyA1, bodyA2)
	}
}

func TestServerCacheDisabled(t *testing.T) {
	s := &armed.ServeCmd{Dir: "testdata/server"}
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	_, body1, cache1 := getWithCacheStatus(t, ts.URL+"/uuid.jsonnet")
	_, body2, cache2 := getWithCacheStatus(t, ts.URL+"/uuid.jsonnet")
	if cache1 != "" || cache2 != "" {
		t.Errorf("X-Cache header must not be set when cache is disabled: %q, %q", cache1, cache2)
	}
	if body1 == body2 {
		t.Errorf("responses must be re-evaluated when cache is disabled, got same body %q", body1)
	}
}

func TestServerCacheConcurrent(t *testing.T) {
	s := &armed.ServeCmd{Dir: "testdata/server", Cache: 100 * time.Millisecond}
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	var wg sync.WaitGroup
	for i := range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp, err := http.Get(fmt.Sprintf("%s/uuid.jsonnet?i=%d", ts.URL, i%5))
			if err != nil {
				t.Error(err)
				return
			}
			defer resp.Body.Close()
			io.Copy(io.Discard, resp.Body)
			if resp.StatusCode != http.StatusOK {
				t.Errorf("status: got %d, want 200", resp.StatusCode)
			}
		}()
	}
	wg.Wait()
}

func TestServerGracefulShutdown(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	s := &armed.ServeCmd{Dir: "testdata/server"}
	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		errCh <- s.Serve(ctx, ln)
	}()

	url := "http://" + ln.Addr().String() + "/static.jsonnet"
	var resp *http.Response
	for range 10 {
		resp, err = http.Get(url)
		if err == nil {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if err != nil {
		t.Fatalf("failed to GET %s: %v", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusOK)
	}

	cancel()
	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("Serve returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("Serve did not return within 2s after context cancel")
	}
}
