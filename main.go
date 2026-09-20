package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	walkcontent "github.com/sidvishnoi/walkcontent/src"
)

const usage = `Usage:
  walkcontent -d <directory> [ignorePattern ...] [-d <directory> [ignorePattern ...] ...] [-o <output>] [--include-content]

  -d <directory>       a base directory to walk (repeatable)
  [ignorePattern ...]  doublestar glob patterns to ignore within that directory
  -o <output>          write JSON output to this file (default: stdout)
  --include-content    also collect every heading and link found in each file

  -h, --help           show this help message`

func main() {
	if hasHelpFlag(os.Args[1:]) {
		fmt.Println(usage)
		return
	}

	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(1)
	}
}

func hasHelpFlag(args []string) bool {
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			return true
		}
	}
	return false
}

func run(args []string) error {
	dirs, output, opts, err := parseArgs(args)
	if err != nil {
		return err
	}

	entries, err := walkcontent.Build(dirs, opts)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding json: %w", err)
	}
	data = append(data, '\n')

	if output == "" {
		_, err := os.Stdout.Write(data)
		return err
	}
	if dir := filepath.Dir(output); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("creating directory for %s: %w", output, err)
		}
	}
	if err := os.WriteFile(output, data, 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", output, err)
	}
	return nil
}

func parseArgs(args []string) (dirs []walkcontent.Dir, output string, opts walkcontent.Options, err error) {
	c := &argCursor{args: args}
	var outputSet bool

	for {
		flag, ok := c.next()
		if !ok {
			break
		}

		switch flag {
		case "-d":
			dir, ok := c.next()
			if !ok {
				return nil, "", opts, fmt.Errorf("-d requires a directory argument")
			}
			d := walkcontent.Dir{Directory: filepath.ToSlash(dir)}
			for !c.peekIsFlag() {
				pattern, _ := c.next()
				d.IgnorePatterns = append(d.IgnorePatterns, filepath.ToSlash(pattern))
			}
			dirs = append(dirs, d)

		case "-o":
			if outputSet {
				return nil, "", opts, fmt.Errorf("-o specified more than once")
			}
			out, ok := c.next()
			if !ok {
				return nil, "", opts, fmt.Errorf("-o requires an output path argument")
			}
			output = out
			outputSet = true

		case "--include-content":
			opts.IncludeContent = true

		default:
			return nil, "", opts, fmt.Errorf("unexpected argument: %s", flag)
		}
	}

	if len(dirs) == 0 {
		return nil, "", opts, fmt.Errorf("at least one -d <directory> is required")
	}
	return dirs, output, opts, nil
}

type argCursor struct {
	args []string
	pos  int
}

func (c *argCursor) next() (string, bool) {
	if c.pos >= len(c.args) {
		return "", false
	}
	arg := c.args[c.pos]
	c.pos++
	return arg, true
}

func (c *argCursor) peekIsFlag() bool {
	return c.pos >= len(c.args) || isFlag(c.args[c.pos])
}

func isFlag(arg string) bool {
	return arg == "-d" || arg == "-o" || arg == "--include-content"
}
