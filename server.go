package armed

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/go-jsonnet"
)

const serverShutdownTimeout = 5 * time.Second

// ServeCmd runs an HTTP server that evaluates jsonnet files under Dir
// and returns the results as JSON.
type ServeCmd struct {
	Listen  string            `name:"listen" default:"localhost:9898" help:"Listen address (host:port)"`
	Timeout time.Duration     `short:"t" name:"timeout" help:"Timeout for each request's evaluation (e.g., 30s, 5m)"`
	ExtStr  map[string]string `short:"V" name:"ext-str" help:"Default external string variables (overridden by query parameters)"`
	Cache   time.Duration     `name:"cache" help:"Cache evaluation results in memory for specified duration (e.g., 5m, 1h)"`
	Stale   time.Duration     `name:"stale" help:"Maximum duration to serve stale cache when evaluation fails (e.g., 10m, 2h)"`
	Dir     string            `arg:"" name:"dir" help:"Directory containing .jsonnet files to serve" type:"existingdir"`

	// functions holds additional native functions to be added to the Jsonnet VM
	functions []*jsonnet.NativeFunction `kong:"-"`

	// cache holds the in-memory cache for evaluation results
	cache cacheStore `kong:"-"`
}

// AddFunctions adds custom native functions to the server
func (s *ServeCmd) AddFunctions(funcs ...*jsonnet.NativeFunction) {
	s.functions = append(s.functions, funcs...)
}

// Run listens on s.Listen and serves HTTP until ctx is cancelled
func (s *ServeCmd) Run(ctx context.Context) error {
	ln, err := net.Listen("tcp", s.Listen)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", s.Listen, err)
	}
	return s.Serve(ctx, ln)
}

// Serve serves HTTP on the given listener until ctx is cancelled,
// then shuts down gracefully.
func (s *ServeCmd) Serve(ctx context.Context, ln net.Listener) error {
	srv := &http.Server{Handler: s.Handler()}
	if s.cache != nil {
		go s.cleanCachePeriodically(ctx)
	}
	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Serve(ln)
	}()
	slog.Info("jsonnet-armed server starting", "addr", ln.Addr().String(), "dir", s.Dir)
	select {
	case <-ctx.Done():
		sctx, cancel := context.WithTimeout(context.Background(), serverShutdownTimeout)
		defer cancel()
		if err := srv.Shutdown(sctx); err != nil {
			srv.Close()
			return err
		}
		return nil
	case err := <-errCh:
		return err
	}
}

// Handler returns the HTTP handler of the server
func (s *ServeCmd) Handler() http.Handler {
	if s.Cache > 0 && s.cache == nil {
		s.cache = newMemoryCache(s.Cache, s.Stale)
	}
	return http.HandlerFunc(s.handleRequest)
}

// cleanCachePeriodically removes completely expired cache entries until
// ctx is cancelled.
func (s *ServeCmd) cleanCachePeriodically(ctx context.Context) {
	interval := max(s.Cache, s.Stale)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.cache.Clean()
		}
	}
}

func (s *ServeCmd) handleRequest(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	status := s.processHTTPRequest(w, r)
	slog.Info("request",
		"method", r.Method,
		"path", r.URL.Path,
		"status", status,
		"duration", time.Since(start),
		"remote", r.RemoteAddr,
	)
}

func (s *ServeCmd) processHTTPRequest(w http.ResponseWriter, r *http.Request) int {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		return writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
	}

	filename, ok := s.resolvePath(r.URL.Path)
	if !ok {
		return writeJSONError(w, http.StatusNotFound, "not found")
	}

	cli := &CLI{
		Filename:  filename,
		ExtStr:    s.mergeQueryVars(r.URL.Query()),
		functions: s.functions,
	}

	var cacheKey, staleContent string
	if s.cache != nil {
		if content, err := os.ReadFile(filename); err != nil {
			slog.Warn("Failed to read file for cache key", "error", err.Error(), "file", filename)
		} else if key, err := generateCacheKey(cli, content); err != nil {
			slog.Warn("Failed to generate cache key", "error", err.Error(), "file", filename)
		} else {
			cacheKey = key
			if cached, isStale, ok := s.cache.GetWithStale(key); ok {
				if !isStale {
					return writeJSONResponse(w, cached, "HIT")
				}
				staleContent = cached
			}
		}
	}

	ectx := r.Context()
	if s.Timeout > 0 {
		var cancel context.CancelFunc
		ectx, cancel = context.WithTimeout(ectx, s.Timeout)
		defer cancel()
	}

	// Run evaluation in a goroutine to enable timeout even if a native
	// function blocks.
	resultCh := make(chan result, 1)
	go func() {
		jsonStr, err := cli.evaluate(ectx, "", false)
		resultCh <- result{jsonStr: jsonStr, err: err}
	}()

	select {
	case res := <-resultCh:
		if res.err != nil {
			if staleContent != "" {
				slog.Warn("Evaluation failed, using stale cache", "error", res.err.Error(), "file", filename)
				return writeJSONResponse(w, staleContent, "STALE")
			}
			slog.Error("failed to evaluate", "file", filename, "error", res.err)
			return writeJSONError(w, http.StatusInternalServerError, res.err.Error())
		}
		if cacheKey != "" {
			if err := s.cache.Set(cacheKey, res.jsonStr); err != nil {
				slog.Warn("Failed to save cache", "error", err.Error(), "file", filename)
			}
		}
		var cacheStatus string
		if s.cache != nil {
			cacheStatus = "MISS"
		}
		return writeJSONResponse(w, res.jsonStr, cacheStatus)
	case <-ectx.Done():
		if ectx.Err() == context.DeadlineExceeded {
			if staleContent != "" {
				slog.Warn("Evaluation timed out, using stale cache", "timeout", s.Timeout, "file", filename)
				return writeJSONResponse(w, staleContent, "STALE")
			}
			return writeJSONError(w, http.StatusGatewayTimeout, fmt.Sprintf("evaluation timed out after %v", s.Timeout))
		}
		return writeJSONError(w, http.StatusInternalServerError, ectx.Err().Error())
	}
}

// writeJSONResponse writes a 200 JSON response. cacheStatus is set as the
// X-Cache header when non-empty.
func writeJSONResponse(w http.ResponseWriter, body, cacheStatus string) int {
	if cacheStatus != "" {
		w.Header().Set("X-Cache", cacheStatus)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, body)
	return http.StatusOK
}

// resolvePath maps a URL path to a .jsonnet file under s.Dir.
// It returns false for paths that are not .jsonnet files, escape s.Dir
// (including via symlinks), or do not exist.
func (s *ServeCmd) resolvePath(urlPath string) (string, bool) {
	// A leading slash + path.Clean prevents ".." from escaping the root
	// (e.g. path.Clean("/../x") == "/x"). urlPath is already percent-decoded.
	upath := path.Clean("/" + urlPath)
	if !strings.HasSuffix(upath, ".jsonnet") {
		return "", false
	}
	rel := strings.TrimPrefix(upath, "/")
	// os.Root rejects paths that escape s.Dir, including via symlinks.
	root, err := os.OpenRoot(s.Dir)
	if err != nil {
		return "", false
	}
	defer root.Close()
	st, err := root.Stat(rel)
	if err != nil || st.IsDir() {
		return "", false
	}
	return filepath.Join(s.Dir, filepath.FromSlash(rel)), true
}

// mergeQueryVars merges query parameters into the server-wide default
// ext string variables. Query parameters override the server-wide defaults.
func (s *ServeCmd) mergeQueryVars(q url.Values) map[string]string {
	extStr := make(map[string]string, len(s.ExtStr)+len(q))
	maps.Copy(extStr, s.ExtStr)
	for k, values := range q {
		if len(values) == 0 {
			continue
		}
		extStr[k] = values[0]
	}
	return extStr
}

func writeJSONError(w http.ResponseWriter, code int, msg string) int {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
	return code
}
