package walkcontent

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildIncludeContent(t *testing.T) {
	dir := t.TempDir()
	content := "---\ntitle: Hello\n---\n# Hello\n\nSee [world](/world) and [home][ref].\n\n## Section\n\n[ref]: /home\n"
	if err := os.WriteFile(filepath.Join(dir, "post.md"), []byte(content), 0o644); err != nil {
		t.Fatalf("writing fixture: %v", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	entries, err := Build([]Dir{{Directory: "."}}, Options{IncludeContent: true})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	entry, ok := entries["post.md"]
	if !ok {
		t.Fatalf("missing entry for post.md, got %#v", entries)
	}
	if entry.H1 != "Hello" {
		t.Errorf("H1 = %q, want %q", entry.H1, "Hello")
	}
	wantHeadings := []Heading{
		{Level: 1, Text: "Hello", RawText: "Hello"},
		{Level: 2, Text: "Section", RawText: "Section"},
	}
	if len(entry.Headings) != len(wantHeadings) || entry.Headings[0] != wantHeadings[0] || entry.Headings[1] != wantHeadings[1] {
		t.Errorf("headings = %#v, want %#v", entry.Headings, wantHeadings)
	}
	wantLinks := []Link{
		{Text: "world", RawText: "world", Href: "/world"},
		{Text: "home", RawText: "home", Href: "/home"},
	}
	if len(entry.Links) != len(wantLinks) || entry.Links[0] != wantLinks[0] || entry.Links[1] != wantLinks[1] {
		t.Errorf("links = %#v, want %#v", entry.Links, wantLinks)
	}
}

func TestBuildIncludeContentH1IgnoresPositionRestriction(t *testing.T) {
	dir := t.TempDir()
	// unlike the fast path, a heading here isn't required to be the first
	// thing in the file to count as $h1 -- it's derived purely from the
	// parsed heading list.
	content := "intro text\n\n# Heading\n"
	if err := os.WriteFile(filepath.Join(dir, "post.md"), []byte(content), 0o644); err != nil {
		t.Fatalf("writing fixture: %v", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	entries, err := Build([]Dir{{Directory: "."}}, Options{IncludeContent: true})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	entry, ok := entries["post.md"]
	if !ok {
		t.Fatalf("missing entry for post.md, got %#v", entries)
	}
	if entry.H1 != "Heading" {
		t.Errorf("H1 = %q, want %q", entry.H1, "Heading")
	}
}

func TestBuildWithoutIncludeContentOmitsHeadingsAndLinks(t *testing.T) {
	dir := t.TempDir()
	content := "# Hello\n\nSee [world](/world).\n"
	if err := os.WriteFile(filepath.Join(dir, "post.md"), []byte(content), 0o644); err != nil {
		t.Fatalf("writing fixture: %v", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	entries, err := Build([]Dir{{Directory: "."}}, Options{})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	entry, ok := entries["post.md"]
	if !ok {
		t.Fatalf("missing entry for post.md, got %#v", entries)
	}
	if entry.Headings != nil {
		t.Errorf("headings = %#v, want nil", entry.Headings)
	}
	if entry.Links != nil {
		t.Errorf("links = %#v, want nil", entry.Links)
	}
}
