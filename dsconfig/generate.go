package dsconfig

// Regenerate the per-pack `if/then` enum arrays in schema.json from the
// pack JSON files in packs/. Run with `go generate ./...` from the
// dsconfig module root.
//
//go:generate go run ./cmd/gen-schema-json
