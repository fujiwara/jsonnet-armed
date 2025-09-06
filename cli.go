package armed

import (
	"time"

	"github.com/alecthomas/kong"
)

type CLI struct {
	OutputFile string            `short:"o" name:"output-file" help:"Write to the output file rather than stdout" type:"path"`
	ExtStr     map[string]string `short:"V" name:"ext-str" help:"Set external string variable (can be repeated)."`
	ExtCode    map[string]string `name:"ext-code" help:"Set external code variable (can be repeated)."`
	Timeout    time.Duration     `short:"t" name:"timeout" help:"Timeout for evaluation (e.g., 30s, 5m, 1h)"`
	Version    kong.VersionFlag  `short:"v" help:"Show version and exit."`

	Filename string `arg:"" name:"filename" help:"Filename or code to execute" type:"path"`
}
