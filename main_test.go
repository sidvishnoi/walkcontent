package main

import (
	"reflect"
	"runtime"
	"testing"

	walkcontent "github.com/sidvishnoi/walkcontent/src"
)

func TestParseArgs(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantDirs   []walkcontent.Dir
		wantOutput string
		wantOpts   walkcontent.Options
		wantErr    bool
	}{
		{
			name:     "single dir no patterns",
			args:     []string{"-d", "content"},
			wantDirs: []walkcontent.Dir{{Directory: "content"}},
		},
		{
			name:     "include-content flag",
			args:     []string{"-d", "content", "--include-content"},
			wantDirs: []walkcontent.Dir{{Directory: "content"}},
			wantOpts: walkcontent.Options{IncludeContent: true},
		},
		{
			name: "single dir with patterns",
			args: []string{"-d", "content", "drafts/**", "*.tmp.md"},
			wantDirs: []walkcontent.Dir{
				{Directory: "content", IgnorePatterns: []string{"drafts/**", "*.tmp.md"}},
			},
		},
		{
			name: "multiple dirs and output",
			args: []string{
				"-d", "content", "drafts/**",
				"-d", "blog", "unpublished/**",
				"-o", "out.json",
			},
			wantDirs: []walkcontent.Dir{
				{Directory: "content", IgnorePatterns: []string{"drafts/**"}},
				{Directory: "blog", IgnorePatterns: []string{"unpublished/**"}},
			},
			wantOutput: "out.json",
		},
		{
			name:    "no -d is an error",
			args:    []string{"-o", "out.json"},
			wantErr: true,
		},
		{
			name:    "-d without a directory is an error",
			args:    []string{"-d"},
			wantErr: true,
		},
		{
			name:    "-o without a value is an error",
			args:    []string{"-d", "content", "-o"},
			wantErr: true,
		},
		{
			name:    "-o twice is an error",
			args:    []string{"-d", "content", "-o", "a.json", "-o", "b.json"},
			wantErr: true,
		},
		{
			name:    "argument before any flag is an error",
			args:    []string{"content"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dirs, output, opts, err := parseArgs(tt.args)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if !reflect.DeepEqual(dirs, tt.wantDirs) {
				t.Errorf("dirs = %#v, want %#v", dirs, tt.wantDirs)
			}
			if output != tt.wantOutput {
				t.Errorf("output = %q, want %q", output, tt.wantOutput)
			}
			if opts != tt.wantOpts {
				t.Errorf("opts = %#v, want %#v", opts, tt.wantOpts)
			}
		})
	}
}

func TestParseArgsConvertsWindowsPathsToSlash(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("path separator conversion only applies on windows")
	}

	dirs, _, _, err := parseArgs([]string{"-d", `content\blog`, `drafts\**`})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []walkcontent.Dir{{Directory: "content/blog", IgnorePatterns: []string{"drafts/**"}}}
	if !reflect.DeepEqual(dirs, want) {
		t.Errorf("dirs = %#v, want %#v", dirs, want)
	}
}

func TestHasHelpFlag(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{name: "no args", args: nil, want: false},
		{name: "-h alone", args: []string{"-h"}, want: true},
		{name: "--help alone", args: []string{"--help"}, want: true},
		{name: "-h among other args", args: []string{"-d", "content", "-h"}, want: true},
		{name: "no help flag", args: []string{"-d", "content", "-o", "out.json"}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasHelpFlag(tt.args); got != tt.want {
				t.Errorf("hasHelpFlag(%v) = %v, want %v", tt.args, got, tt.want)
			}
		})
	}
}
