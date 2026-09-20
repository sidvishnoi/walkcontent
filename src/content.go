package walkcontent

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"

	"gopkg.in/yaml.v3"
)

var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

const frontmatterDelim = "---"

func parseFrontmatterAndH1(r io.Reader) (map[string]any, string, int, error) {
	lines := newLineScanner(r)
	fm, contentStart, line, ok, err := parseFrontmatter(lines)
	if err != nil {
		return nil, "", 0, err
	}
	h1, _ := scanH1(lines, line, ok)
	return fm, h1, contentStart, nil
}

func parseFrontmatterAndContent(r io.Reader) (map[string]any, int, []byte, error) {
	lines := newLineScanner(r)
	fm, contentStart, line, ok, err := parseFrontmatter(lines)
	if err != nil {
		return nil, 0, nil, err
	}

	var body bytes.Buffer
	if ok {
		body.WriteString(line)
	}
	if _, err := body.ReadFrom(lines.r); err != nil {
		return nil, 0, nil, fmt.Errorf("reading content: %w", err)
	}

	return fm, contentStart, body.Bytes(), nil
}

func parseFrontmatter(lines *lineScanner) (fm map[string]any, contentStart int, line string, ok bool, err error) {
	line, ok = lines.next()
	line = strings.TrimPrefix(line, string(utf8BOM))

	contentStart = 1
	if ok && strings.TrimSpace(line) == frontmatterDelim {
		var block strings.Builder
		closed := false
		for {
			l, more := lines.next()
			if !more {
				break
			}
			if strings.TrimSpace(l) == frontmatterDelim {
				closed = true
				break
			}
			block.WriteString(l)
		}
		if closed {
			if err := yaml.Unmarshal([]byte(block.String()), &fm); err != nil {
				return nil, 0, "", false, fmt.Errorf("invalid yaml frontmatter: %w", err)
			}
			contentStart = lines.lineNo + 1
		}
		line, ok = lines.next()
	}

	return fm, contentStart, line, ok, nil
}

type lineScanner struct {
	r      *bufio.Reader
	done   bool
	lineNo int
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
	s.lineNo++
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
