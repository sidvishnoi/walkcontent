package walkcontent

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestParseHeadingsAndLinks(t *testing.T) {
	tests := []struct {
		name         string
		in           string
		wantHeadings []Heading
		wantLinks    []Link
	}{
		{
			name: "atx headings of various levels",
			in:   "# Title\n\n## Section\n\n### Sub\n",
			wantHeadings: []Heading{
				{Level: 1, Text: "Title", RawText: "Title"},
				{Level: 2, Text: "Section", RawText: "Section"},
				{Level: 3, Text: "Sub", RawText: "Sub"},
			},
		},
		{
			name: "setext headings",
			in:   "Title\n=====\n\nSection\n-------\n",
			wantHeadings: []Heading{
				{Level: 1, Text: "Title", RawText: "Title"},
				{Level: 2, Text: "Section", RawText: "Section"},
			},
		},
		{
			name: "heading with inline formatting preserves raw markdown",
			in:   "# Hello **world**\n",
			wantHeadings: []Heading{
				{Level: 1, Text: "Hello world", RawText: "Hello **world**"},
			},
		},
		{
			name:      "inline link",
			in:        "See [my post](/posts/hello) for more.\n",
			wantLinks: []Link{{Text: "my post", RawText: "my post", Href: "/posts/hello"}},
		},
		{
			name:      "reference link",
			in:        "See [my post][ref] for more.\n\n[ref]: /posts/hello\n",
			wantLinks: []Link{{Text: "my post", RawText: "my post", Href: "/posts/hello"}},
		},
		{
			name:      "autolink",
			in:        "Visit <https://example.com>.\n",
			wantLinks: []Link{{Text: "https://example.com", RawText: "https://example.com", Href: "https://example.com"}},
		},
		{
			name: "link inside a heading is collected as both",
			in:   "# [Home](/)\n",
			wantHeadings: []Heading{
				{Level: 1, Text: "Home", RawText: "[Home](/)"},
			},
			wantLinks: []Link{{Text: "Home", RawText: "Home", Href: "/"}},
		},
		{
			name: "headings and links inside a fenced code block are ignored",
			in:   "# Real Heading\n\n```md\n# Not a heading\n[not a link](/nope)\n```\n",
			wantHeadings: []Heading{
				{Level: 1, Text: "Real Heading", RawText: "Real Heading"},
			},
		},
		{
			name:      "link text with a code span preserves backticks",
			in:        "[`code`](/code)\n",
			wantLinks: []Link{{Text: "code", RawText: "`code`", Href: "/code"}},
		},
		{
			name:      "link text with emphasis preserves markdown",
			in:        "[**bold** text](/x)\n",
			wantLinks: []Link{{Text: "bold text", RawText: "**bold** text", Href: "/x"}},
		},
		{
			name: "plain text has no headings or links",
			in:   "just some text\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headings, links := parseHeadingsAndLinks([]byte(tt.in))
			if !reflect.DeepEqual(headings, tt.wantHeadings) {
				t.Errorf("headings = %#v, want %#v", headings, tt.wantHeadings)
			}
			if !reflect.DeepEqual(links, tt.wantLinks) {
				t.Errorf("links = %#v, want %#v", links, tt.wantLinks)
			}
		})
	}
}

func TestHeadingMarshalJSONOmitsRawTextWhenSame(t *testing.T) {
	tests := []struct {
		name string
		h    Heading
		want string
	}{
		{
			name: "plain text: rawText omitted",
			h:    Heading{Level: 1, Text: "Hello", RawText: "Hello"},
			want: `{"level":1,"text":"Hello"}`,
		},
		{
			name: "formatted text: rawText included",
			h:    Heading{Level: 1, Text: "Hello world", RawText: "Hello **world**"},
			want: `{"level":1,"rawText":"Hello **world**","text":"Hello world"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.h)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if string(got) != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}

func TestLinkMarshalJSONOmitsRawTextWhenSame(t *testing.T) {
	tests := []struct {
		name string
		l    Link
		want string
	}{
		{
			name: "plain text: rawText omitted",
			l:    Link{Text: "home", Href: "/home", RawText: "home"},
			want: `{"href":"/home","text":"home"}`,
		},
		{
			name: "formatted text: rawText included",
			l:    Link{Text: "code", Href: "/code", RawText: "`code`"},
			want: `{"href":"/code","rawText":"` + "`code`" + `","text":"code"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.l)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if string(got) != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}
