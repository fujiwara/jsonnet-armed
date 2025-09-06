package functions

import (
	"fmt"
	"time"

	"github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/ast"
)

// timeFormats maps format constant names to their actual format strings
var timeFormats = map[string]string{
	// Common time format constants
	"RFC3339":     time.RFC3339,     // "2006-01-02T15:04:05Z07:00"
	"RFC3339Nano": time.RFC3339Nano, // "2006-01-02T15:04:05.999999999Z07:00"
	"RFC1123":     time.RFC1123,     // "Mon, 02 Jan 2006 15:04:05 MST"
	"RFC1123Z":    time.RFC1123Z,    // "Mon, 02 Jan 2006 15:04:05 -0700"
	"DateTime":    time.DateTime,    // "2006-01-02 15:04:05"
	"DateOnly":    time.DateOnly,    // "2006-01-02"
	"TimeOnly":    time.TimeOnly,    // "15:04:05"
}

// getTimeFormat returns the actual time format string.
// If format matches a Go time constant name, it returns the corresponding format.
// Otherwise, it returns the format as-is.
func getTimeFormat(format string) string {
	if actualFormat, exists := timeFormats[format]; exists {
		return actualFormat
	}
	return format
}

var TimeFunctions = map[string]*jsonnet.NativeFunction{
	"now": {
		Params: []ast.Identifier{},
		Func: func(args []any) (any, error) {
			// Return Unix timestamp as float64 with nanosecond precision
			return float64(time.Now().UnixNano()) / float64(time.Second), nil
		},
	},
	"time_format": {
		Params: []ast.Identifier{"timestamp", "format"},
		Func: func(args []any) (any, error) {
			timestamp, ok := args[0].(float64)
			if !ok {
				return nil, fmt.Errorf("time_format: timestamp must be a number")
			}

			format, ok := args[1].(string)
			if !ok {
				return nil, fmt.Errorf("time_format: format must be a string")
			}

			// Convert float64 timestamp to time.Time
			seconds := int64(timestamp)
			nanos := int64((timestamp - float64(seconds)) * float64(time.Second))
			t := time.Unix(seconds, nanos).UTC()

			// Check if format is a Go time constant name
			actualFormat := getTimeFormat(format)

			// Format the time using Go's time format
			return t.Format(actualFormat), nil
		},
	},
}

func init() {
	initializeFunctionMap(TimeFunctions)
}
