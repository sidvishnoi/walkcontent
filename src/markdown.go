package walkcontent

import (
	"bytes"
	"encoding/json"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// Heading is a markdown heading (ATX or Setext) found in a file's content.
type Heading struct {
	Level int    `json:"level"`
	Text  string `json:"text"`
	// RawText is Text as written, without markdown formatting stripped
	// (e.g. "**bold**" rather than "bold").
	RawText string `json:"rawText"`
}

func (h Heading) MarshalJSON() ([]byte, error) {
	out := map[string]any{"level": h.Level, "text": h.Text}
	if h.RawText != h.Text {
		out["rawText"] = h.RawText
	}
	return json.Marshal(out)
}

// Link is a markdown link (inline, reference, or autolink) found in a
// file's content.
type Link struct {
	Text string `json:"text"`
	Href string `json:"href"`
	// RawText is Text as written, without markdown formatting stripped
	// (e.g. "**bold**" rather than "bold").
	RawText string `json:"rawText"`
}

func (l Link) MarshalJSON() ([]byte, error) {
	out := map[string]any{"text": l.Text, "href": l.Href}
	if l.RawText != l.Text {
		out["rawText"] = l.RawText
	}
	return json.Marshal(out)
}

// parseHeadingsAndLinks parses source as markdown and collects every heading
// and link in document order.
func parseHeadingsAndLinks(source []byte) (headings []Heading, links []Link) {
	root := goldmark.DefaultParser().Parse(text.NewReader(source))

	_ = ast.Walk(root, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch node := n.(type) {
		case *ast.Heading:
			headings = append(headings, Heading{
				Level:   node.Level,
				Text:    nodeText(node, source),
				RawText: headingRawText(node, source),
			})
		case *ast.Link:
			links = append(links, Link{
				Text:    nodeText(node, source),
				Href:    string(node.Destination),
				RawText: linkRawText(node, source),
			})
		case *ast.AutoLink:
			url := string(node.URL(source))
			label := string(node.Label(source))
			links = append(links, Link{Text: label, Href: url, RawText: label})
		}
		return ast.WalkContinue, nil
	})

	return headings, links
}

// headingRawText returns a heading's content exactly as written, i.e.
// before inline parsing splits it into Text/Emphasis/etc. children -- so
// markdown formatting like "**bold**" is preserved as-is.
func headingRawText(h *ast.Heading, source []byte) string {
	lines := h.Lines()
	var buf bytes.Buffer
	for i := 0; i < lines.Len(); i++ {
		if i > 0 {
			buf.WriteByte('\n')
		}
		seg := lines.At(i)
		buf.Write(seg.Value(source))
	}
	return buf.String()
}

// linkRawText returns a link's label exactly as written, with any markdown
// formatting inside it (e.g. "**bold**", “ `code` “) preserved as-is.
func linkRawText(n ast.Node, source []byte) string {
	if start, end, ok := spanOfChildren(n, source); ok {
		return string(source[start:end])
	}
	// n's label has no plain-text leaves to anchor a raw span to (e.g. it
	// consists solely of an autolink); fall back to the rendered text.
	return nodeText(n, source)
}

// spanOf returns the source byte range spanned by n as originally written,
// including any of its own markdown delimiters (e.g. the surrounding "**"
// of an *ast.Emphasis, or the backticks of an *ast.CodeSpan).
func spanOf(n ast.Node, source []byte) (start, end int, ok bool) {
	switch node := n.(type) {
	case *ast.Text:
		return node.Segment.Start, node.Segment.Stop, true
	case *ast.Emphasis:
		start, end, ok := spanOfChildren(node, source)
		if !ok {
			return 0, 0, false
		}
		return start - node.Level, end + node.Level, true
	case *ast.CodeSpan:
		start, end, ok := spanOfChildren(node, source)
		if !ok {
			return 0, 0, false
		}
		for start > 0 && source[start-1] == '`' {
			start--
		}
		for end < len(source) && source[end] == '`' {
			end++
		}
		return start, end, true
	default:
		return spanOfChildren(n, source)
	}
}

// spanOfChildren returns the source byte range spanning all of n's
// children, as written.
func spanOfChildren(n ast.Node, source []byte) (start, end int, ok bool) {
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		cs, ce, cok := spanOf(c, source)
		if !cok {
			continue
		}
		if !ok || cs < start {
			start = cs
		}
		if !ok || ce > end {
			end = ce
		}
		ok = true
	}
	return start, end, ok
}

// nodeText concatenates the plain-text content of n's inline children,
// e.g. resolving `**bold** text` to "bold text".
func nodeText(n ast.Node, source []byte) string {
	var buf bytes.Buffer
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		writeNodeText(&buf, c, source)
	}
	return buf.String()
}

func writeNodeText(buf *bytes.Buffer, n ast.Node, source []byte) {
	switch node := n.(type) {
	case *ast.Text:
		buf.Write(node.Value(source))
		if node.SoftLineBreak() || node.HardLineBreak() {
			buf.WriteByte(' ')
		}
	case *ast.AutoLink:
		buf.Write(node.Label(source))
	default:
		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			writeNodeText(buf, c, source)
		}
	}
}
