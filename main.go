package armed

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/alecthomas/kong"
	"github.com/fujiwara/jsonnet-armed/functions"
	"github.com/google/go-jsonnet"
)

// SetOutput sets the output destination for jsonnet evaluation results (deprecated)
// Use CLI.Writer field instead for thread-safe operation
func SetOutput(w io.Writer) {
	// This function is kept for backward compatibility but should not be used
	// in concurrent tests as it modifies global state
}

// SetWriter sets the writer for CLI
func (cli *CLI) SetWriter(w io.Writer) {
	cli.writer = w
}

func Run(ctx context.Context) error {
	cli := &CLI{writer: os.Stdout}
	kong.Parse(cli, kong.Vars{"version": fmt.Sprintf("jsonnet-armed %s", Version)})
	return cli.run(ctx)
}

// Run runs the jsonnet evaluation with the CLI configuration
func (cli *CLI) Run(ctx context.Context) error {
	// Set default writer if not specified
	if cli.writer == nil {
		cli.writer = os.Stdout
	}
	return cli.run(ctx)
}

func (cli *CLI) run(ctx context.Context) error {
	// Initialize cache if enabled
	var cache *Cache
	if cli.Cache > 0 {
		cache = NewCache(cli.Cache, cli.Stale)
		// Clean expired cache entries (best effort)
		go cache.Clean()
	}

	// Apply timeout if specified
	if cli.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cli.Timeout)
		defer cancel()
	}

	// Create a channel to signal completion
	resultCh := make(chan result, 1)

	// Run all operations in goroutine to enable timeout
	go func() {
		res := cli.processRequest(ctx, cache)
		resultCh <- res
	}()

	// Wait for either completion or timeout
	select {
	case res := <-resultCh:
		return res.err

	case <-ctx.Done():
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("evaluation timed out after %v", cli.Timeout)
		}
		return ctx.Err()
	}
}

type result struct {
	jsonStr string
	err     error
}

func (cli *CLI) processRequest(ctx context.Context, cache *Cache) result {
	// Read input content and determine if it's from stdin
	var inputContent string
	var isStdin bool
	var contentBytes []byte
	var err error

	if cli.Filename == "-" {
		// Read from stdin
		contentBytes, err = io.ReadAll(os.Stdin)
		if err != nil {
			return result{jsonStr: "", err: fmt.Errorf("failed to read from stdin: %w", err)}
		}
		inputContent = string(contentBytes)
		isStdin = true
	} else {
		// For files, we need content for cache key generation
		if cache != nil {
			contentBytes, err = os.ReadFile(cli.Filename)
			if err != nil {
				return result{jsonStr: "", err: fmt.Errorf("failed to read file: %w", err)}
			}
		}
		isStdin = false
	}

	// Try to get from cache if enabled
	var staleContent string
	if cache != nil {
		cacheKey, err := cache.GenerateCacheKey(cli, contentBytes)
		if err != nil {
			slog.Warn("Failed to generate cache key",
				"error", err.Error(),
				"filename", cli.Filename)
		} else {
			if cachedResult, isStale, exists := cache.GetWithStale(cacheKey); exists {
				if !isStale {
					// Use fresh cached result
					err = cli.writeOutput(cachedResult)
					return result{jsonStr: cachedResult, err: err}
				}
				// Store stale content for potential fallback
				staleContent = cachedResult
			}
			// Store cache key for later use
			cli.cacheKey = cacheKey
		}
	}

	jsonStr, err := cli.evaluate(ctx, inputContent, isStdin)
	if err != nil {
		// If evaluation failed and we have stale cache, use it
		if staleContent != "" {
			slog.Warn("Evaluation failed, using stale cache",
				"error", err.Error(),
				"filename", cli.Filename)
			err = cli.writeOutput(staleContent)
			return result{jsonStr: staleContent, err: err}
		}
		return result{jsonStr: "", err: err}
	}

	// Cache the result if cache is enabled
	if cache != nil && cli.cacheKey != "" {
		// Store in cache (best effort, log errors)
		if err := cache.Set(cli.cacheKey, jsonStr); err != nil {
			slog.Warn("Failed to save cache",
				"error", err.Error(),
				"cache_key", cli.cacheKey[:8]+"...",
				"filename", cli.Filename)
		}
	}

	// Write output within the timeout scope
	err = cli.writeOutput(jsonStr)
	return result{jsonStr: jsonStr, err: err}
}

func (cli *CLI) evaluate(ctx context.Context, content string, isStdin bool) (string, error) {
	vm := jsonnet.MakeVM()

	// Register native functions
	ctx = context.WithValue(ctx, "version", Version)
	funcs := functions.GenerateAllFunctions(ctx)
	for _, f := range funcs {
		vm.NativeFunction(f)
	}

	// Add importer for armed.libsonnet
	vm.Importer(&ArmedImporter{funcs: funcs})

	for k, v := range cli.ExtStr {
		vm.ExtVar(k, v)
	}
	for k, v := range cli.ExtCode {
		vm.ExtCode(k, v)
	}

	var jsonStr string
	var err error

	if isStdin {
		jsonStr, err = vm.EvaluateAnonymousSnippet("stdin", content)
	} else {
		jsonStr, err = vm.EvaluateFile(cli.Filename)
	}
	if err != nil {
		return "", fmt.Errorf("failed to evaluate: %w", err)
	}

	return jsonStr, nil
}

func (cli *CLI) writeOutput(jsonStr string) error {
	if cli.OutputFile != "" {
		data := []byte(jsonStr)

		// Check if content has changed when WriteIfChanged is enabled
		if cli.WriteIfChanged && shouldSkipWrite(cli.OutputFile, data) {
			return nil
		}

		return writeFileAtomic(cli.OutputFile, data, 0644)
	}
	_, err := io.WriteString(cli.writer, jsonStr)
	return err
}

// shouldSkipWrite checks if the file write should be skipped because content hasn't changed
func shouldSkipWrite(filename string, newData []byte) bool {
	// Check if file exists
	info, err := os.Stat(filename)
	if err != nil {
		// File doesn't exist, need to write
		return false
	}

	// Compare file sizes first
	if info.Size() != int64(len(newData)) {
		// Different size, content has changed
		return false
	}

	// Size is the same, compare SHA256 hashes
	// Calculate hash of existing file using streaming
	file, err := os.Open(filename)
	if err != nil {
		// Error opening file, write anyway
		return false
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		// Error reading file, write anyway
		return false
	}
	existingHash := hasher.Sum(nil)

	// Calculate hash of new data
	newHash := sha256.Sum256(newData)

	return bytes.Equal(existingHash, newHash[:])
}

// writeFileAtomic writes data to the named file atomically.
// It writes to a temporary file first, then renames it to the target file.
func writeFileAtomic(filename string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(filename)
	base := filepath.Base(filename)

	// Create a temporary file in the same directory
	tmpfile, err := os.CreateTemp(dir, base+".tmp")
	if err != nil {
		return err
	}
	tmpname := tmpfile.Name()

	// Clean up the temporary file if something goes wrong
	defer func() {
		if tmpfile != nil {
			tmpfile.Close()
			os.Remove(tmpname)
		}
	}()

	// Write data to the temporary file
	if _, err := tmpfile.Write(data); err != nil {
		return err
	}

	// Sync to ensure data is written to disk
	if err := tmpfile.Sync(); err != nil {
		return err
	}

	// Set the correct permissions
	if err := tmpfile.Chmod(perm); err != nil {
		return err
	}

	// Close the file before renaming
	if err := tmpfile.Close(); err != nil {
		return err
	}
	tmpfile = nil // Prevent defer from removing the file

	// Atomically replace the target file
	if err := os.Rename(tmpname, filename); err != nil {
		os.Remove(tmpname)
		return err
	}

	return nil
}

// ArmedImporter provides virtual file system for armed.libsonnet
type ArmedImporter struct {
	funcs []*jsonnet.NativeFunction
}

func (ai *ArmedImporter) Import(importedFrom, importedPath string) (contents jsonnet.Contents, foundAt string, err error) {
	if importedPath == "armed.libsonnet" {
		// Generate the library content dynamically
		content := functions.GenerateArmedLib(ai.funcs)
		return jsonnet.MakeContents(content), "armed.libsonnet", nil
	}

	// Fall back to default file system import
	importer := &jsonnet.FileImporter{}
	return importer.Import(importedFrom, importedPath)
}
