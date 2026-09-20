package walkcontent

import (
	"strings"
	"testing"
)

func TestParseFrontmatterAndH1(t *testing.T) {
	tests := []struct {
		name             string
		in               string
		wantFM           map[string]any
		wantHeading      string
		wantContentStart int
	}{
		{
			name:             "frontmatter then atx heading",
			in:               "---\ntitle: Hello\n---\n# Heading\n\nbody\n",
			wantFM:           map[string]any{"title": "Hello"},
			wantHeading:      "Heading",
			wantContentStart: 4,
		},
		{
			name:             "blank lines between frontmatter and heading are skipped",
			in:               "---\ntitle: Hello\n---\n\n\n# Heading\n",
			wantFM:           map[string]any{"title": "Hello"},
			wantHeading:      "Heading",
			wantContentStart: 4,
		},
		{
			name:             "content before heading rules it out",
			in:               "---\ntitle: Hello\n---\nsome intro text\n\n# Heading\n",
			wantFM:           map[string]any{"title": "Hello"},
			wantContentStart: 4,
		},
		{
			name:             "setext heading",
			in:               "---\ntitle: Hello\n---\nHeading\n===\n",
			wantFM:           map[string]any{"title": "Hello"},
			wantHeading:      "Heading",
			wantContentStart: 4,
		},
		{
			name:             "h2 is not h1",
			in:               "---\ntitle: Hello\n---\n## Heading\n",
			wantFM:           map[string]any{"title": "Hello"},
			wantContentStart: 4,
		},
		{
			name:             "no frontmatter, heading first",
			in:               "# Heading\n\nbody\n",
			wantHeading:      "Heading",
			wantContentStart: 1,
		},
		{
			name:             "no frontmatter, no heading",
			in:               "just text\n",
			wantContentStart: 1,
		},
		{
			name:             "unterminated frontmatter delimiter rules out both",
			in:               "---\ntitle: Hello\n# Heading\n",
			wantContentStart: 1,
		},
		{
			name:             "empty file",
			in:               "",
			wantContentStart: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm, h1, contentStart, err := parseFrontmatterAndH1(strings.NewReader(tt.in))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if h1 != tt.wantHeading {
				t.Errorf("heading = %q, want %q", h1, tt.wantHeading)
			}
			if contentStart != tt.wantContentStart {
				t.Errorf("contentStart = %d, want %d", contentStart, tt.wantContentStart)
			}
			if len(fm) != len(tt.wantFM) {
				t.Fatalf("frontmatter = %#v, want %#v", fm, tt.wantFM)
			}
			for k, want := range tt.wantFM {
				if got := fm[k]; got != want {
					t.Errorf("frontmatter[%q] = %#v, want %#v", k, got, want)
				}
			}
		})
	}
}
