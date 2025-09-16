package armed

import (
	"context"
	"fmt"
	"io"
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
	// Apply timeout if specified
	if cli.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cli.Timeout)
		defer cancel()
	}

	// Create a channel to signal completion
	resultCh := make(chan result, 1)

	// Run evaluation and output in goroutine to enable timeout
	go func() {
		jsonStr, err := cli.evaluate(ctx)
		if err != nil {
			resultCh <- result{jsonStr: "", err: err}
			return
		}

		// Write output within the timeout scope
		err = cli.writeOutput(jsonStr)
		resultCh <- result{jsonStr: jsonStr, err: err}
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

func (cli *CLI) evaluate(ctx context.Context) (string, error) {
	vm := jsonnet.MakeVM()

	// Register native functions
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

	if cli.Filename == "-" {
		// Read from stdin
		input, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("failed to read from stdin: %w", err)
		}
		jsonStr, err = vm.EvaluateAnonymousSnippet("stdin", string(input))
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
		return writeFileAtomic(cli.OutputFile, []byte(jsonStr), 0644)
	}
	_, err := io.WriteString(cli.writer, jsonStr)
	return err
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
