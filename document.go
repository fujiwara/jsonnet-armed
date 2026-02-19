package armed

import (
	_ "embed"
	"strings"
)

//go:embed README.md
var readmeContent string

// extractTOC extracts table of contents from markdown content.
// It finds all heading lines (starting with #) and formats them with indentation.
// Lines inside fenced code blocks (```) are skipped.
func extractTOC(content string) string {
	var b strings.Builder
	inCodeBlock := false
	for line := range strings.SplitSeq(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			inCodeBlock = !inCodeBlock
			continue
		}
		if inCodeBlock {
			continue
		}
		if !strings.HasPrefix(trimmed, "#") {
			continue
		}
		// Count the number of # characters
		level := 0
		for _, c := range trimmed {
			if c == '#' {
				level++
			} else {
				break
			}
		}
		if level == 0 {
			continue
		}
		// Extract the heading text
		title := strings.TrimSpace(trimmed[level:])
		if title == "" {
			continue
		}
		// Add indentation based on level (level 1 = no indent, level 2 = 2 spaces, etc.)
		indent := strings.Repeat("  ", level-1)
		b.WriteString(indent)
		b.WriteString(title)
		b.WriteString("\n")
	}
	return b.String()
}

// searchSections searches for sections in markdown content that match the keyword.
// It splits content by heading lines, then returns sections
// where the keyword appears (case-insensitive) in the heading or body.
// Lines inside fenced code blocks (```) are not treated as headings.
func searchSections(content string, keyword string) string {
	keyword = strings.ToLower(keyword)

	lines := strings.Split(content, "\n")
	type section struct {
		heading string
		body    strings.Builder
	}

	var sections []section
	var current *section
	inCodeBlock := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			inCodeBlock = !inCodeBlock
			if current != nil {
				current.body.WriteString(line)
				current.body.WriteString("\n")
			}
			continue
		}
		if !inCodeBlock && strings.HasPrefix(trimmed, "#") {
			sections = append(sections, section{heading: line})
			current = &sections[len(sections)-1]
		} else if current != nil {
			current.body.WriteString(line)
			current.body.WriteString("\n")
		}
	}

	var result strings.Builder
	for _, sec := range sections {
		text := strings.ToLower(sec.heading) + strings.ToLower(sec.body.String())
		if strings.Contains(text, keyword) {
			if result.Len() > 0 {
				result.WriteString("\n")
			}
			result.WriteString(sec.heading)
			result.WriteString("\n")
			result.WriteString(sec.body.String())
		}
	}
	return result.String()
}
