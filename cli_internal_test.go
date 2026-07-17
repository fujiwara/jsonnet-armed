package armed

import (
	"testing"

	"github.com/alecthomas/kong"
)

func TestRootCLIDispatch(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantCmd string
	}{
		{"filename only", []string{"testdata/server/static.jsonnet"}, "eval <filename>"},
		{"flag before filename", []string{"-c", "testdata/server/static.jsonnet"}, "eval <filename>"},
		{"stdin", []string{"-"}, "eval <filename>"},
		{"document flag only", []string{"--document"}, "eval"},
		{"serve", []string{"serve", "testdata/server"}, "serve <dir>"},
		{"serve with listen", []string{"serve", "--listen", "127.0.0.1:0", "testdata/server"}, "serve <dir>"},
		{"serve with log-format", []string{"serve", "--log-format", "json", "testdata/server"}, "serve <dir>"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := &rootCLI{}
			parser, err := kong.New(root, kong.Vars{"version": "test"})
			if err != nil {
				t.Fatal(err)
			}
			kctx, err := parser.Parse(tt.args)
			if err != nil {
				t.Fatal(err)
			}
			if kctx.Command() != tt.wantCmd {
				t.Errorf("command: got %q, want %q", kctx.Command(), tt.wantCmd)
			}
		})
	}
}

func TestServeLogFormatEnum(t *testing.T) {
	root := &rootCLI{}
	parser, err := kong.New(root, kong.Vars{"version": "test"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := parser.Parse([]string{"serve", "--log-format", "xml", "testdata/server"}); err == nil {
		t.Error("invalid --log-format value must be rejected")
	}
}
