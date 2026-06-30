package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestLoadPacks(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "alpha_settings.json"), `{
		"id": "alpha_settings",
		"fields": [
			{"id": "alpha_settings.b"},
			{"id": "alpha_settings.a"}
		]
	}`)
	writeFile(t, filepath.Join(dir, "beta_settings.json"), `{
		"id": "beta_settings",
		"fields": []
	}`)
	// Non-matching file should be ignored.
	writeFile(t, filepath.Join(dir, "ignored.json"), `{"id":"nope"}`)

	got, err := loadPacks(dir)
	if err != nil {
		t.Fatalf("loadPacks: %v", err)
	}
	want := map[string][]string{
		"alpha_settings": {"alpha_settings.a", "alpha_settings.b"},
		"beta_settings":  {},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestLoadPacks_MissingID(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "bad_settings.json"), `{"fields":[]}`)
	if _, err := loadPacks(dir); err == nil {
		t.Fatal("expected error for missing pack id")
	}
}

func TestLoadPacks_EmptyFieldID(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "bad_settings.json"), `{"id":"bad_settings","fields":[{"id":""}]}`)
	if _, err := loadPacks(dir); err == nil {
		t.Fatal("expected error for empty field id")
	}
}

func TestRenderEnum(t *testing.T) {
	cases := []struct {
		name   string
		values []string
		indent string
		want   string
	}{
		{
			name:   "empty",
			values: nil,
			indent: "  ",
			want:   `"enum": []`,
		},
		{
			name:   "one value",
			values: []string{"a"},
			indent: "  ",
			want:   "\"enum\": [\n      \"a\"\n  ]",
		},
		{
			name:   "escaped",
			values: []string{`a"b`},
			indent: "",
			want:   "\"enum\": [\n    \"a\\\"b\"\n]",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := renderEnum(tc.values, tc.indent)
			if got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestLineIndent(t *testing.T) {
	src := "abc\n    \"enum\": []\nxyz"
	pos := strings.Index(src, `"enum"`)
	if got := lineIndent(src, pos); got != "    " {
		t.Fatalf("got %q, want 4 spaces", got)
	}
	if got := lineIndent("no-newline", 3); got != "" {
		t.Fatalf("got %q, want empty", got)
	}
}

func TestSortedKeys(t *testing.T) {
	got := sortedKeys(map[string][]string{"b": nil, "a": nil, "c": nil})
	want := []string{"a", "b", "c"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestReplaceNextEnum(t *testing.T) {
	src := `{
    "x": {
      "enum": ["old"]
    }
}`
	out, end, err := replaceNextEnum(src, 0, []string{"new1", "new2"})
	if err != nil {
		t.Fatalf("replaceNextEnum: %v", err)
	}
	if !strings.Contains(out, `"new1"`) || !strings.Contains(out, `"new2"`) {
		t.Fatalf("expected new values in output, got:\n%s", out)
	}
	if strings.Contains(out, `"old"`) {
		t.Fatalf("old value not replaced:\n%s", out)
	}
	if end <= 0 || end > len(out) {
		t.Fatalf("end index out of range: %d (len=%d)", end, len(out))
	}
}

func TestReplaceNextEnum_NotFound(t *testing.T) {
	if _, _, err := replaceNextEnum(`{"x":1}`, 0, []string{"a"}); err == nil {
		t.Fatal("expected error when no enum present")
	}
}

func TestRewritePackBranch(t *testing.T) {
	src := `{
  "allOf": [
    {
      "if": { "properties": { "from": { "const": "alpha_settings" } } },
      "then": {
        "properties": {
          "exclude": { "items": { "enum": [] } },
          "patch": { "propertyNames": { "enum": [] } }
        }
      }
    }
  ]
}`
	out, err := rewritePackBranch(src, "alpha_settings", []string{"alpha_settings.a", "alpha_settings.b"})
	if err != nil {
		t.Fatalf("rewritePackBranch: %v", err)
	}
	if strings.Count(out, `"alpha_settings.a"`) != 2 || strings.Count(out, `"alpha_settings.b"`) != 2 {
		t.Fatalf("expected both field IDs in both enum arrays, got:\n%s", out)
	}
}

func TestRewritePackBranch_MissingPackIsNoop(t *testing.T) {
	src := `{"allOf":[]}`
	out, err := rewritePackBranch(src, "missing_settings", []string{"x"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != src {
		t.Fatalf("expected unchanged input, got:\n%s", out)
	}
}

func TestUpdateSchema(t *testing.T) {
	src := []byte(`{
  "allOf": [
    {
      "if": { "properties": { "from": { "const": "alpha_settings" } } },
      "then": {
        "properties": {
          "exclude": { "items": { "enum": [] } },
          "patch": { "propertyNames": { "enum": [] } }
        }
      }
    }
  ]
}`)
	out, err := updateSchema(src, map[string][]string{
		"alpha_settings": {"alpha_settings.a"},
	})
	if err != nil {
		t.Fatalf("updateSchema: %v", err)
	}
	if !strings.Contains(string(out), `"alpha_settings.a"`) {
		t.Fatalf("expected field id in output:\n%s", out)
	}
}

func writeFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
