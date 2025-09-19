package functions

import (
	"fmt"
	"regexp"

	"github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/ast"
)

// regexMatchFunction checks if the text matches the regular expression pattern
func regexMatchFunction(args []any) (any, error) {
	pattern, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("pattern must be a string")
	}
	text, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("text must be a string")
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %w", err)
	}

	return re.MatchString(text), nil
}

// regexFindFunction finds the first match of the regular expression in the text
func regexFindFunction(args []any) (any, error) {
	pattern, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("pattern must be a string")
	}
	text, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("text must be a string")
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %w", err)
	}

	match := re.FindString(text)
	if match == "" {
		return nil, nil // Return null for no match
	}
	return match, nil
}

// regexFindAllFunction finds all matches of the regular expression in the text
func regexFindAllFunction(args []any) (any, error) {
	pattern, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("pattern must be a string")
	}
	text, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("text must be a string")
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %w", err)
	}

	matches := re.FindAllString(text, -1)
	if matches == nil {
		return []any{}, nil // Return empty array for no matches
	}

	// Convert []string to []any for Jsonnet compatibility
	result := make([]any, len(matches))
	for i, match := range matches {
		result[i] = match
	}
	return result, nil
}

// regexReplaceFunction replaces all matches of the regular expression with the replacement string
func regexReplaceFunction(args []any) (any, error) {
	pattern, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("pattern must be a string")
	}
	replacement, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("replacement must be a string")
	}
	text, ok := args[2].(string)
	if !ok {
		return nil, fmt.Errorf("text must be a string")
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %w", err)
	}

	return re.ReplaceAllString(text, replacement), nil
}

// regexSplitFunction splits the text by the regular expression pattern
func regexSplitFunction(args []any) (any, error) {
	pattern, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("pattern must be a string")
	}
	text, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("text must be a string")
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %w", err)
	}

	parts := re.Split(text, -1)

	// Convert []string to []any for Jsonnet compatibility
	result := make([]any, len(parts))
	for i, part := range parts {
		result[i] = part
	}
	return result, nil
}

var RegexpFunctions = map[string]*jsonnet.NativeFunction{
	"regex_match": {
		Params: []ast.Identifier{"pattern", "text"},
		Func:   regexMatchFunction,
	},
	"regex_find": {
		Params: []ast.Identifier{"pattern", "text"},
		Func:   regexFindFunction,
	},
	"regex_find_all": {
		Params: []ast.Identifier{"pattern", "text"},
		Func:   regexFindAllFunction,
	},
	"regex_replace": {
		Params: []ast.Identifier{"pattern", "replacement", "text"},
		Func:   regexReplaceFunction,
	},
	"regex_split": {
		Params: []ast.Identifier{"pattern", "text"},
		Func:   regexSplitFunction,
	},
}

func init() {
	initializeFunctionMap(RegexpFunctions)
}
