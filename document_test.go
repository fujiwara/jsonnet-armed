package armed

import (
	"strings"
	"testing"
)

func TestExtractTOC(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "empty content",
			content:  "",
			expected: "",
		},
		{
			name:     "no headings",
			content:  "This is plain text\nwith no headings.",
			expected: "",
		},
		{
			name:     "single level 1 heading",
			content:  "# Title\nSome content.",
			expected: "Title\n",
		},
		{
			name:     "multiple levels",
			content:  "# Title\n## Section 1\n### Subsection 1.1\n## Section 2\n",
			expected: "Title\n  Section 1\n    Subsection 1.1\n  Section 2\n",
		},
		{
			name:     "heading with extra spaces",
			content:  "#   Spaced Title  \n##  Spaced Section  \n",
			expected: "Spaced Title\n  Spaced Section\n",
		},
		{
			name:     "level 4 heading",
			content:  "# Top\n#### Deep Heading\n",
			expected: "Top\n      Deep Heading\n",
		},
		{
			name:     "skip empty headings",
			content:  "# Title\n##\n## Valid\n",
			expected: "Title\n  Valid\n",
		},
		{
			name:     "skip headings in code blocks",
			content:  "# Title\n```bash\n# This is a comment\n## Not a heading\n```\n## Real Section\n",
			expected: "Title\n  Real Section\n",
		},
		{
			name:    "embedded README has headings",
			content: readmeContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractTOC(tt.content)
			if tt.name == "embedded README has headings" {
				// Just verify it's not empty
				if got == "" {
					t.Error("extractTOC returned empty string for embedded README")
				}
				return
			}
			if got != tt.expected {
				t.Errorf("extractTOC() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestSearchSections(t *testing.T) {
	content := `# Project Title

This is the project introduction.

## Installation

Install with go install.

## Hash Functions

Calculate hash of the given string.

### MD5

MD5 produces a 128-bit hash.

### SHA256

SHA256 produces a 256-bit hash.

## DNS Functions

Perform DNS lookups for various record types.

## License

MIT
`

	tests := []struct {
		name            string
		keyword         string
		wantEmpty       bool
		wantContains    []string
		wantNotContains []string
	}{
		{
			name:         "match section by heading",
			keyword:      "hash",
			wantContains: []string{"Hash Functions", "MD5", "SHA256"},
		},
		{
			name:         "case insensitive match",
			keyword:      "HASH",
			wantContains: []string{"Hash Functions"},
		},
		{
			name:         "match in body text",
			keyword:      "128-bit",
			wantContains: []string{"MD5"},
		},
		{
			name:      "no match",
			keyword:   "nonexistent-keyword-xyz",
			wantEmpty: true,
		},
		{
			name:            "match specific section only",
			keyword:         "dns",
			wantContains:    []string{"DNS Functions"},
			wantNotContains: []string{"Hash Functions"},
		},
		{
			name:         "match introduction",
			keyword:      "introduction",
			wantContains: []string{"Project Title"},
		},
		{
			name:         "match license",
			keyword:      "mit",
			wantContains: []string{"License"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := searchSections(content, tt.keyword)
			if tt.wantEmpty {
				if got != "" {
					t.Errorf("searchSections() returned non-empty result: %q", got)
				}
				return
			}
			if got == "" {
				t.Error("searchSections() returned empty result")
				return
			}
			for _, want := range tt.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("searchSections() result should contain %q, got:\n%s", want, got)
				}
			}
			for _, notWant := range tt.wantNotContains {
				if strings.Contains(got, notWant) {
					t.Errorf("searchSections() result should not contain %q, got:\n%s", notWant, got)
				}
			}
		})
	}
}
