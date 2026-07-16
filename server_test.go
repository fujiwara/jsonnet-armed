package armed_test

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
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
