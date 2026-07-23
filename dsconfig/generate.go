package dsconfig

// Regenerate the per-pack `if/then` enum arrays in schema.json from the
// pack JSON files in packs/. Run with `go generate ./...` from the
// dsconfig module root.
//
//go:generate go run ./cmd/gen-schema-json

// Regenerate the consumer-facing CONFIGURATION.md documents for every
// data source in the registry from their dsconfig.json files.
//
//go:generate go run ./cmd/gen-docs -dir ../registry
