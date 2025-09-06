package armed

import (
	"context"
	"fmt"
	"io"
	"os"

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
	return run(ctx, cli)
}

// RunWithCLI runs the jsonnet evaluation with the given CLI configuration
func RunWithCLI(ctx context.Context, cli *CLI) error {
	// Set default writer if not specified
	if cli.writer == nil {
		cli.writer = os.Stdout
	}
	return run(ctx, cli)
}

func run(ctx context.Context, cli *CLI) error {
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
		jsonStr, err := evaluate(cli)
		if err != nil {
			resultCh <- result{jsonStr: "", err: err}
			return
		}

		// Write output within the timeout scope
		err = writeOutput(cli, jsonStr)
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

func evaluate(cli *CLI) (string, error) {
	vm := jsonnet.MakeVM()

	// Add importer for armed.libsonnet
	vm.Importer(&ArmedImporter{})

	// Register native functions
	for _, f := range functions.AllFunctions() {
		vm.NativeFunction(f)
	}

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

func writeOutput(cli *CLI, jsonStr string) error {
	if cli.OutputFile != "" {
		return os.WriteFile(cli.OutputFile, []byte(jsonStr), 0644)
	}
	_, err := io.WriteString(cli.writer, jsonStr)
	return err
}

// ArmedImporter provides virtual file system for armed.libsonnet
type ArmedImporter struct{}

func (ai *ArmedImporter) Import(importedFrom, importedPath string) (contents jsonnet.Contents, foundAt string, err error) {
	if importedPath == "armed.libsonnet" {
		// Generate the library content dynamically
		content := functions.GenerateArmedLib()
		return jsonnet.MakeContents(content), "armed.libsonnet", nil
	}

	// Fall back to default file system import
	importer := &jsonnet.FileImporter{}
	return importer.Import(importedFrom, importedPath)
}
