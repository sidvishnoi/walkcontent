package walkcontent

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"gopkg.in/yaml.v3"
)

var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

const frontmatterDelim = "---"

func parseFrontmatterAndH1(r io.Reader) (map[string]any, string, error) {
	lines := newLineScanner(r)

	line, ok := lines.next()
	line = strings.TrimPrefix(line, string(utf8BOM))

	var fm map[string]any
	if ok && strings.TrimSpace(line) == frontmatterDelim {
		var block strings.Builder
		closed := false
		for {
			line, more := lines.next()
			if !more {
				break
			}
			if strings.TrimSpace(line) == frontmatterDelim {
				closed = true
				break
			}
			block.WriteString(line)
		}
		if !closed {
			// no closing delimiter: this wasn't frontmatter after all, and its first
			// line ("---") already rules out a heading directly following it
			return nil, "", nil
		}
		if err := yaml.Unmarshal([]byte(block.String()), &fm); err != nil {
			return nil, "", fmt.Errorf("invalid yaml frontmatter: %w", err)
		}
		line, ok = lines.next()
	}

	h1, _ := scanH1(lines, line, ok)
	return fm, h1, nil
}

type lineScanner struct {
	r    *bufio.Reader
	done bool
}

func newLineScanner(r io.Reader) *lineScanner {
	return &lineScanner{r: bufio.NewReader(r)}
}

func (s *lineScanner) next() (string, bool) {
	if s.done {
		return "", false
	}
	line, err := s.r.ReadString('\n')
	if err != nil {
		s.done = true
		if line == "" {
			return "", false
		}
	}
	return line, true
}

// Look for a top-level markdown heading: `# Heading` (ATX) or `Heading\n===`
// (Setext). If the first non-blank line is not a heading; then assume no
// heading.
func scanH1(lines *lineScanner, line string, ok bool) (string, bool) {
	line, ok = skipBlankLines(lines, line, ok)
	if !ok {
		return "", false
	}
	trimmed := strings.TrimSpace(line)
	if heading, isH1 := atxH1(trimmed); isH1 {
		return heading, true
	}
	next, nextOK := lines.next()
	if nextOK && isSetextH1Underline(strings.TrimSpace(next)) {
		return trimmed, true
	}
	return "", false
}

func skipBlankLines(lines *lineScanner, line string, ok bool) (string, bool) {
	for ok && strings.TrimSpace(line) == "" {
		line, ok = lines.next()
	}
	return line, ok
}

func atxH1(trimmed string) (string, bool) {
	if !strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "##") {
		return "", false
	}
	rest := strings.TrimPrefix(trimmed, "#")
	if rest != "" && !strings.HasPrefix(rest, " ") && !strings.HasPrefix(rest, "\t") {
		return "", false
	}
	heading := strings.TrimSpace(rest)
	heading = strings.TrimRight(heading, "#")
	return strings.TrimSpace(heading), true
}

func isSetextH1Underline(trimmed string) bool {
	return trimmed != "" && strings.Trim(trimmed, "=") == ""
}
