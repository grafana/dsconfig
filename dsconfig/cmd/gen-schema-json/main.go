// Command gen-schema-json regenerates the per-pack `if/then` enum arrays
// inside dsconfig/schema.json from the on-disk field pack JSON files.
//
// For each pack JSON file in dsconfig/packs/*_settings.json, the generator
// collects the field IDs and writes them into the matching BaseFieldRef
// allOf branch in schema.json:
//
//	allOf[i].then.properties.exclude.items.enum         <- pack field IDs
//	allOf[i].then.properties.patch.propertyNames.enum   <- pack field IDs
//
// This keeps `exclude` and `patch` autocomplete in editors (and JSON Schema
// validation) in sync with the actual pack contents without hand-editing
// schema.json. Run with:
//
//	go generate ./...
//
// or directly:
//
//	go run ./cmd/gen-schema-json
//
// The command is intentionally dependency-free (stdlib only) so it can run
// in any environment that has the `go` toolchain.
//
// Helper functions and their tests live in utils.go / utils_test.go.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

// Default paths are resolved relative to the dsconfig module root, which is
// discovered via moduleRoot() at runtime. Callers can still override them
// with explicit flags.
const (
	defaultPacksSubdir  = "packs"
	defaultSchemaFile   = "schema.json"
)

func main() {
	// Resolve defaults relative to the dsconfig module root so the command
	// works regardless of the caller's CWD (e.g. when run via `go generate`
	// from a parent module, or directly via `go run`).
	defaultPacks := defaultPacksSubdir
	defaultSchema := defaultSchemaFile
	if root, ok := moduleRoot(); ok {
		defaultPacks = filepath.Join(root, defaultPacksSubdir)
		defaultSchema = filepath.Join(root, defaultSchemaFile)
	}

	packsDir := flag.String("packs", defaultPacks, "path to the dsconfig packs directory")
	schemaPath := flag.String("schema", defaultSchema, "path to dsconfig/schema.json")
	flag.Parse()

	packs, err := loadPacks(*packsDir)
	if err != nil {
		fail("load packs: %v", err)
	}
	if len(packs) == 0 {
		fail("no packs found in %s", *packsDir)
	}

	original, err := os.ReadFile(*schemaPath) // #nosec G304 -- flag-controlled path
	if err != nil {
		fail("read schema: %v", err)
	}

	updated, err := updateSchema(original, packs)
	if err != nil {
		fail("update schema: %v", err)
	}

	if string(updated) == string(original) {
		fmt.Fprintln(os.Stderr, "gen-schema-json: schema.json already up to date")
		return
	}

	if err := os.WriteFile(*schemaPath, updated, 0o644); err != nil { // #nosec G306 -- committed source file
		fail("write schema: %v", err)
	}
	fmt.Fprintf(os.Stderr, "gen-schema-json: updated %s\n", *schemaPath)
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "gen-schema-json: "+format+"\n", args...)
	os.Exit(1)
}
