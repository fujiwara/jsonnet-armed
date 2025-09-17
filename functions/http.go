package functions

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/ast"
)

const versionKey = "version"

var (
	// DefaultHttpTimeout is the default timeout for HTTP requests
	DefaultHttpTimeout = 30 * time.Second
)

// setDefaultUserAgent sets the default User-Agent header if not already present
func setDefaultUserAgent(req *http.Request, version string) {
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", fmt.Sprintf("jsonnet-armed/%s", version))
	}
}

// makeHttpRequest is the shared implementation for HTTP requests
func makeHttpRequest(method, url string, headers map[string]any, body string, version string) (any, error) {
	var bodyReader io.Reader
	if body != "" {
		bodyReader = bytes.NewReader([]byte(body))
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("http request: failed to create request: %w", err)
	}

	// Set headers if provided
	for key, value := range headers {
		if valueStr, ok := value.(string); ok {
			req.Header.Set(key, valueStr)
		} else {
			return nil, fmt.Errorf("http request: header value for %s must be a string", key)
		}
	}

	// Set default User-Agent if not specified
	setDefaultUserAgent(req, version)

	client := &http.Client{
		Timeout: DefaultHttpTimeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: request failed: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("http request: failed to read response body: %w", err)
	}

	// Convert response headers to map[string]any
	responseHeaders := make(map[string]any)
	for key, values := range resp.Header {
		if len(values) == 1 {
			responseHeaders[key] = values[0]
		} else {
			responseHeaders[key] = values
		}
	}

	result := map[string]any{
		"status_code": resp.StatusCode,
		"status":      resp.Status,
		"headers":     responseHeaders,
		"body":        string(responseBody),
	}

	return result, nil
}

func GenerateHttpFunctions(ctx context.Context) map[string]*jsonnet.NativeFunction {
	version, _ := ctx.Value(versionKey).(string)
	if version == "" {
		version = "unknown"
	}

	funcs := map[string]*jsonnet.NativeFunction{
		"http_request": {
			Params: []ast.Identifier{"method", "url", "headers", "body"},
			Func: func(args []any) (any, error) {
				method, ok := args[0].(string)
				if !ok {
					return nil, fmt.Errorf("http_request: method must be a string")
				}

				url, ok := args[1].(string)
				if !ok {
					return nil, fmt.Errorf("http_request: url must be a string")
				}

				var headers map[string]any
				if args[2] != nil {
					headersMap, ok := args[2].(map[string]any)
					if !ok {
						return nil, fmt.Errorf("http_request: headers must be an object or null")
					}
					headers = headersMap
				}

				var body string
				if args[3] != nil {
					bodyStr, ok := args[3].(string)
					if !ok {
						return nil, fmt.Errorf("http_request: body must be a string or null")
					}
					body = bodyStr
				}

				return makeHttpRequest(method, url, headers, body, version)
			},
		},
		"http_get": {
			Params: []ast.Identifier{"url", "headers"},
			Func: func(args []any) (any, error) {
				// Validate URL argument
				url, ok := args[0].(string)
				if !ok {
					return nil, fmt.Errorf("http_get: url must be a string")
				}

				var headers map[string]any
				if args[1] != nil {
					headersMap, ok := args[1].(map[string]any)
					if !ok {
						return nil, fmt.Errorf("http_get: headers must be an object or null")
					}
					headers = headersMap
				}

				// Call shared implementation with GET method and no body
				return makeHttpRequest("GET", url, headers, "", version)
			},
		},
	}

	// Initialize function names
	initializeFunctionMap(funcs)
	return funcs
}
