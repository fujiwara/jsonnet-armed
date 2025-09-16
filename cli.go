package armed

import (
	"io"
	"time"

	"github.com/alecthomas/kong"
)

type CLI struct {
	OutputFile     string            `short:"o" name:"output-file" help:"Write to the output file rather than stdout" type:"path"`
	WriteIfChanged bool              `name:"write-if-changed" help:"Write output file only if content has changed"`
	ExtStr         map[string]string `short:"V" name:"ext-str" help:"Set external string variable (can be repeated)."`
	ExtCode        map[string]string `name:"ext-code" help:"Set external code variable (can be repeated)."`
	Timeout        time.Duration     `short:"t" name:"timeout" help:"Timeout for evaluation (e.g., 30s, 5m, 1h)"`
	Version        kong.VersionFlag  `short:"v" help:"Show version and exit."`

	Filename string `arg:"" name:"filename" help:"Filename or code to execute" type:"path"`

	// writer for output (not exposed to CLI, used internally)
	writer io.Writer `kong:"-"`
}
