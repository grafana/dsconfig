package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// The test binary runs in this package's directory, so the dsconfig module
// root (with packs/ and schema.json) is two levels up.
var (
	repoPacksDir   = filepath.Join("..", "..", defaultPacksSubdir)
	repoSchemaPath = filepath.Join("..", "..", defaultSchemaFile)
)

// TestSchemaJSONInSync fails when the committed schema.json no longer matches
// what the generator would produce from the committed pack files, for example
// after a pack gains a field without regenerating schema.json.
func TestSchemaJSONInSync(t *testing.T) {
	packs, err := loadPacks(repoPacksDir)
	if err != nil {
		t.Fatalf("loadPacks: %v", err)
	}
	original, err := os.ReadFile(repoSchemaPath)
	if err != nil {
		t.Fatalf("read schema: %v", err)
	}
	updated, err := updateSchema(original, packs)
	if err != nil {
		t.Fatalf("updateSchema: %v", err)
	}
	if string(updated) != string(original) {
		t.Fatal("schema.json is out of date with the pack files; run `go generate ./...` from the dsconfig module root")
	}
}

// TestSchemaJSONCoversAllPacks fails when a pack file has no matching entry in
// schema.json. The generator skips such packs silently, so a new pack would
// otherwise never get editor autocomplete or validation for exclude/patch.
func TestSchemaJSONCoversAllPacks(t *testing.T) {
	packs, err := loadPacks(repoPacksDir)
	if err != nil {
		t.Fatalf("loadPacks: %v", err)
	}
	raw, err := os.ReadFile(repoSchemaPath)
	if err != nil {
		t.Fatalf("read schema: %v", err)
	}
	var doc struct {
		Definitions struct {
			BaseFieldRef struct {
				Properties struct {
					From struct {
						Enum []string `json:"enum"`
					} `json:"from"`
				} `json:"properties"`
			} `json:"BaseFieldRef"`
		} `json:"definitions"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("parse schema: %v", err)
	}
	fromEnum := doc.Definitions.BaseFieldRef.Properties.From.Enum

	for _, packID := range sortedKeys(packs) {
		if !slices.Contains(fromEnum, packID) {
			t.Errorf("pack %q is missing from BaseFieldRef.from.enum in schema.json", packID)
		}
		if !strings.Contains(string(raw), `"const": "`+packID+`"`) {
			t.Errorf("pack %q has no allOf branch in schema.json; add one so the generator can fill its enums", packID)
		}
	}
}
