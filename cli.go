package armed

import (
	"io"
	"time"

	"github.com/alecthomas/kong"
	"github.com/google/go-jsonnet"
)

type CLI struct {
	Output         string            `short:"o" name:"output" help:"Write to the output file or http(s) URL rather than stdout"`
	Stdout         bool              `short:"S" name:"stdout" help:"Also write to stdout when using -o/--output" negatable:""`
	WriteIfChanged bool              `name:"write-if-changed" help:"Write output file only if content has changed"`
	ExtStr         map[string]string `short:"V" name:"ext-str" help:"Set external string variable (can be repeated)."`
	ExtCode        map[string]string `name:"ext-code" help:"Set external code variable (can be repeated)."`
	Timeout        time.Duration     `short:"t" name:"timeout" help:"Timeout for evaluation (e.g., 30s, 5m, 1h)"`
	Cache          time.Duration     `short:"c" name:"cache" help:"Cache evaluation results for specified duration (e.g., 5m, 1h)"`
	Stale          time.Duration     `name:"stale" help:"Maximum duration to use stale cache when evaluation fails (e.g., 10m, 2h)"`
	Version        kong.VersionFlag  `short:"v" help:"Show version and exit."`

	Filename string `arg:"" name:"filename" help:"Filename or code to execute" type:"path"`

	// writer for output (not exposed to CLI, used internally)
	writer io.Writer `kong:"-"`

	// cacheKey holds the generated cache key (internal use)
	cacheKey string `kong:"-"`

	// functions holds additional native functions to be added to the Jsonnet VM
	functions []*jsonnet.NativeFunction `kong:"-"`
}
