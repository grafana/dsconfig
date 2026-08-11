package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type packFile struct {
	ID     string `json:"id"`
	Fields []struct {
		ID string `json:"id"`
	} `json:"fields"`
}

// loadPacks reads every *_settings.json file in dir and returns a map of
// pack ID -> sorted field IDs.
func loadPacks(dir string) (map[string][]string, error) {
	matches, err := filepath.Glob(filepath.Join(dir, "*_settings.json"))
	if err != nil {
		return nil, err
	}
	out := map[string][]string{}
	for _, path := range matches {
		data, err := os.ReadFile(path) // #nosec G304 -- glob-controlled path
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", path, err)
		}
		var p packFile
		if err := json.Unmarshal(data, &p); err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
		if p.ID == "" {
			return nil, fmt.Errorf("%s: missing \"id\"", path)
		}
		ids := make([]string, 0, len(p.Fields))
		for _, f := range p.Fields {
			if f.ID == "" {
				return nil, fmt.Errorf("%s: field with empty \"id\"", path)
			}
			ids = append(ids, f.ID)
		}
		sort.Strings(ids)
		out[p.ID] = ids
	}
	return out, nil
}

// updateSchema rewrites the `enum` arrays inside each pack-specific
// BaseFieldRef allOf branch. It operates on the raw JSON bytes to preserve
// the file's hand-curated key ordering, indentation, and comments-style
// layout that round-tripping through encoding/json would destroy.
func updateSchema(src []byte, packs map[string][]string) ([]byte, error) {
	out := string(src)
	for _, packID := range sortedKeys(packs) {
		fieldIDs := packs[packID]
		var err error
		out, err = rewritePackBranch(out, packID, fieldIDs)
		if err != nil {
			return nil, err
		}
	}
	return []byte(out), nil
}

// rewritePackBranch locates the `allOf` branch for a single pack ID and
// replaces the two `enum` arrays inside it (exclude.items.enum first,
// patch.propertyNames.enum second) with the supplied field IDs.
func rewritePackBranch(src, packID string, fieldIDs []string) (string, error) {
	marker := fmt.Sprintf(`"const": %q`, packID)
	markerIdx := strings.Index(src, marker)
	if markerIdx == -1 {
		// Pack has no branch in schema.json yet — not an error; nothing to do.
		return src, nil
	}

	// Two consecutive `"enum": [...]` arrays sit after the marker inside the
	// matching `then` block: the first is exclude.items.enum, the second is
	// patch.propertyNames.enum.
	cursor := markerIdx
	for i := 0; i < 2; i++ {
		next, replaced, err := replaceNextEnum(src, cursor, fieldIDs)
		if err != nil {
			return "", fmt.Errorf("pack %q: %w", packID, err)
		}
		src = next
		cursor = replaced
	}
	return src, nil
}

// enumRe matches `"enum": [ ... ]` allowing any single-line or multi-line
// body. The non-greedy `[\s\S]*?` keeps it bounded to the nearest closing
// bracket, which works because pack-field-ID enums never contain nested
// arrays.
var enumRe = regexp.MustCompile(`"enum":\s*\[[\s\S]*?\]`)

// replaceNextEnum finds the next `"enum": [...]` occurrence at or after
// `from` and replaces it with a new enum array built from values. It
// returns the modified string and the index immediately past the
// replacement, so callers can chain multiple replacements without
// re-matching already-processed regions.
func replaceNextEnum(src string, from int, values []string) (string, int, error) {
	loc := enumRe.FindStringIndex(src[from:])
	if loc == nil {
		return "", 0, fmt.Errorf(`expected "enum": [...] after offset %d`, from)
	}
	start := from + loc[0]
	end := from + loc[1]

	indent := lineIndent(src, start)
	replacement := renderEnum(values, indent)
	return src[:start] + replacement + src[end:], start + len(replacement), nil
}

// lineIndent returns the whitespace prefix of the line containing pos.
func lineIndent(src string, pos int) string {
	lineStart := strings.LastIndexByte(src[:pos], '\n') + 1
	end := lineStart
	for end < len(src) && (src[end] == ' ' || src[end] == '\t') {
		end++
	}
	return src[lineStart:end]
}

// renderEnum produces a deterministic, pretty-printed `"enum": [...]`
// literal. Empty enums collapse to a single line to match the existing
// schema.json convention.
func renderEnum(values []string, indent string) string {
	if len(values) == 0 {
		return `"enum": []`
	}
	itemIndent := indent + "    "
	var b strings.Builder
	b.WriteString(`"enum": [`)
	for i, v := range values {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteByte('\n')
		b.WriteString(itemIndent)
		// Use json.Marshal to get correctly escaped string literals.
		raw, _ := json.Marshal(v)
		b.Write(raw)
	}
	b.WriteByte('\n')
	b.WriteString(indent)
	b.WriteByte(']')
	return b.String()
}

func sortedKeys(m map[string][]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func pathExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

// moduleRoot walks up from this source file's directory to find the
// dsconfig module root (the directory containing packs/ and schema.json).
// Returns ("", false) when the executable was built away from the source
// tree (e.g. installed via `go install`).
func moduleRoot() (string, bool) {
	exe, err := os.Executable()
	if err != nil {
		return "", false
	}
	// When run via `go run`, the executable lives in a temp dir. Fall back
	// to walking up from CWD instead.
	candidates := []string{filepath.Dir(exe)}
	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates, wd)
	}
	for _, start := range candidates {
		dir := start
		for i := 0; i < 8; i++ {
			if pathExists(filepath.Join(dir, "schema.json")) && pathExists(filepath.Join(dir, "packs")) {
				return dir, true
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	return "", false
}
