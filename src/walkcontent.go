package walkcontent

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sync"
	"time"
)

type Entry struct {
	Frontmatter map[string]any
	// H1 is the file's top-level markdown heading, if any.
	H1 string
	// MTime is the file's modification time as a Unix timestamp (seconds).
	MTime int64
	// Date is the first YYYY-MM-DD date found in the file's path, if any.
	Date string
	// ContentStart is the 1-based line number at which content following
	// the frontmatter block (if any) begins.
	ContentStart int
	// Headings holds every heading in the file, in document order (requires Options.IncludeContent).
	Headings []Heading
	// Links holds every link in the file, in document order (requires Options.IncludeContent).
	Links []Link
}

func (e Entry) MarshalJSON() ([]byte, error) {
	out := make(map[string]any, len(e.Frontmatter)+5)
	for k, v := range e.Frontmatter {
		out[k] = v
	}
	if e.H1 != "" {
		out["$h1"] = e.H1
	}
	if e.Date != "" {
		out["$date"] = e.Date
	}
	out["$mtime"] = e.MTime
	out["$contentStart"] = e.ContentStart
	if len(e.Headings) > 0 {
		out["$headings"] = e.Headings
	}
	if len(e.Links) > 0 {
		out["$links"] = e.Links
	}
	return json.Marshal(out)
}

type Options struct {
	// Additionally collect every heading and link found in each file's content.
	// This requires reading and parsing the whole file, rather than just enough
	// to find the frontmatter and first heading.
	IncludeContent bool
}

func Build(dirs []Dir, opts Options) (map[string]Entry, error) {
	norm, err := normalizeDirs(dirs)
	if err != nil {
		return nil, err
	}

	files, err := walkFiles(norm)
	if err != nil {
		return nil, err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("getting working directory: %w", err)
	}

	type keyedEntry struct {
		key   string
		entry Entry
	}
	results := make([]keyedEntry, len(files))
	errs := make([]error, len(files))

	sem := make(chan struct{}, runtime.NumCPU())
	var wg sync.WaitGroup
	for i, f := range files {
		wg.Add(1)
		sem <- struct{}{}
		go func(i int, f string) {
			defer wg.Done()
			defer func() { <-sem }()
			key, entry, err := buildEntry(f, cwd, opts)
			results[i] = keyedEntry{key: key, entry: entry}
			errs[i] = err
		}(i, f)
	}
	wg.Wait()

	for _, err := range errs {
		if err != nil {
			return nil, err
		}
	}

	out := make(map[string]Entry, len(results))
	for _, r := range results {
		out[r.key] = r.entry
	}
	return out, nil
}

func buildEntry(path, cwd string, opts Options) (string, Entry, error) {
	rel, err := filepath.Rel(cwd, path)
	if err != nil {
		return "", Entry{}, fmt.Errorf("resolving relative path for %s: %w", path, err)
	}
	key := filepath.ToSlash(rel)

	f, err := os.Open(path)
	if err != nil {
		return "", Entry{}, fmt.Errorf("opening %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	info, err := f.Stat()
	if err != nil {
		return "", Entry{}, fmt.Errorf("statting %s: %w", path, err)
	}

	var fm map[string]any
	var h1 string
	var contentStart int
	var headings []Heading
	var links []Link
	if opts.IncludeContent {
		var body []byte
		fm, contentStart, body, err = parseFrontmatterAndContent(f)
		if err != nil {
			return "", Entry{}, fmt.Errorf("parsing %s: %w", path, err)
		}
		headings, links = parseHeadingsAndLinks(body)
		for _, h := range headings {
			if h.Level == 1 {
				h1 = h.Text
				break
			}
		}
	} else {
		fm, h1, contentStart, err = parseFrontmatterAndH1(f)
		if err != nil {
			return "", Entry{}, fmt.Errorf("parsing %s: %w", path, err)
		}
	}

	date, _ := extractDate(key)

	return key, Entry{
		Frontmatter:  fm,
		H1:           h1,
		MTime:        info.ModTime().Unix(),
		Date:         date,
		ContentStart: contentStart,
		Headings:     headings,
		Links:        links,
	}, nil
}

// datePattern matches a year-month-day date anywhere in a path, with either
// "-" or "/" between components, e.g. "2024-01-15", "2024/01-15", or
// "2024/01/15" — as in "posts/2024-01-15-hello.md" or "posts/2024/01/15.md".
var datePattern = regexp.MustCompile(`(\d{4})[-/](\d{2})[-/](\d{2})`)

// get the first year-month-day match in path that is also a valid calendar
// date, normalized to "YYYY-MM-DD"
func extractDate(path string) (string, bool) {
	m := datePattern.FindStringSubmatch(path)
	if m == nil {
		return "", false
	}
	date := fmt.Sprintf("%s-%s-%s", m[1], m[2], m[3])
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return "", false
	}
	return date, true
}
