package armed

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/alecthomas/kong"
	"github.com/google/go-jsonnet"
)

func Run(ctx context.Context) error {
	cli := &CLI{}
	kong.Parse(cli, kong.Vars{"version": fmt.Sprintf("jsonnet-armed %s", Version)})
	return run(ctx, cli)
}

func run(ctx context.Context, cli *CLI) error {
	vm := jsonnet.MakeVM()
	/*
		for _, f := range cli.nativeFuncs {
			vm.NativeFunction(f)
		}
	*/
	for k, v := range cli.ExtStr {
		vm.ExtVar(k, v)
	}
	for k, v := range cli.ExtCode {
		vm.ExtCode(k, v)
	}
	jsonStr, err := vm.EvaluateFile(cli.Filename)
	if err != nil {
		return fmt.Errorf("failed to evaluate: %w", err)
	}

	if cli.OutputFile != "" {
		return os.WriteFile(cli.OutputFile, []byte(jsonStr), 0644)
	}
	_, err = io.WriteString(os.Stdout, jsonStr)
	return err
}
