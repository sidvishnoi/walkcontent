package walkcontent

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/bmatcuk/doublestar/v4"
)

type Dir struct {
	Directory string
	// doublestar glob patterns (e.g. "drafts/**", "*.draft.md")
	IgnorePatterns []string
}

var contentExts = map[string]bool{
	".md":       true,
	".mdx":      true,
	".markdown": true,
}

func walkFiles(dirs []normalizedDir) ([]string, error) {
	results := make([][]string, len(dirs))
	errs := make([]error, len(dirs))

	sem := make(chan struct{}, runtime.NumCPU())
	var wg sync.WaitGroup
	for i, dir := range dirs {
		wg.Add(1)
		sem <- struct{}{}
		go func(i int, dir normalizedDir) {
			defer wg.Done()
			defer func() { <-sem }()
			results[i], errs[i] = walkDir(dir)
		}(i, dir)
	}
	wg.Wait()

	for _, err := range errs {
		if err != nil {
			return nil, err
		}
	}

	var all []string
	for _, files := range results {
		all = append(all, files...)
	}
	return all, nil
}

func walkDir(dir normalizedDir) ([]string, error) {
	var files []string
	err := filepath.WalkDir(dir.abs, func(path string, de fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == dir.abs {
			return nil
		}

		rel, err := filepath.Rel(dir.abs, path)
		if err != nil {
			return fmt.Errorf("resolving relative path for %s: %w", path, err)
		}
		relSlash := filepath.ToSlash(rel)

		if de.IsDir() {
			if strings.HasPrefix(de.Name(), ".") || matchesAny(dir.ignorePatterns, relSlash) {
				return filepath.SkipDir
			}
			return nil
		}

		if !contentExts[strings.ToLower(filepath.Ext(path))] {
			return nil
		}
		if matchesAny(dir.ignorePatterns, relSlash) {
			return nil
		}

		files = append(files, path)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walking %s: %w", dir.original, err)
	}
	return files, nil
}

func matchesAny(patterns []string, relSlash string) bool {
	for _, pattern := range patterns {
		if ok, _ := doublestar.Match(pattern, relSlash); ok {
			return true
		}
	}
	return false
}

type normalizedDir struct {
	original       string
	abs            string
	resolved       string
	ignorePatterns []string
}

func normalizeDirs(dirs []Dir) ([]normalizedDir, error) {
	if len(dirs) == 0 {
		return nil, fmt.Errorf("no directories provided")
	}

	normalized := make([]normalizedDir, len(dirs))
	for i, d := range dirs {
		abs, err := filepath.Abs(d.Directory)
		if err != nil {
			return nil, fmt.Errorf("resolving %s: %w", d.Directory, err)
		}
		info, err := os.Stat(abs)
		if err != nil {
			return nil, fmt.Errorf("checking %s: %w", d.Directory, err)
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("%s is not a directory", d.Directory)
		}
		resolved, err := filepath.EvalSymlinks(abs)
		if err != nil {
			return nil, fmt.Errorf("resolving symlinks for %s: %w", d.Directory, err)
		}
		for _, pattern := range d.IgnorePatterns {
			if !doublestar.ValidatePattern(pattern) {
				return nil, fmt.Errorf("invalid ignore pattern %q for %s", pattern, d.Directory)
			}
		}

		normalized[i] = normalizedDir{
			original:       d.Directory,
			abs:            filepath.Clean(abs),
			resolved:       resolved,
			ignorePatterns: d.IgnorePatterns,
		}
	}

	for i := range normalized {
		for j := i + 1; j < len(normalized); j++ {
			if normalized[i].resolved == normalized[j].resolved {
				return nil, fmt.Errorf("%s and %s resolve to the same directory", normalized[i].original, normalized[j].original)
			}
			if isAncestor(normalized[i].resolved, normalized[j].resolved) {
				return nil, fmt.Errorf("%s is a subdirectory of %s", normalized[j].original, normalized[i].original)
			}
			if isAncestor(normalized[j].resolved, normalized[i].resolved) {
				return nil, fmt.Errorf("%s is a subdirectory of %s", normalized[i].original, normalized[j].original)
			}
		}
	}

	return normalized, nil
}

func isAncestor(parent, child string) bool {
	rel, err := filepath.Rel(parent, child)
	if err != nil {
		return false
	}
	return rel != "." && !strings.HasPrefix(rel, "..")
}
