package schema_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/grafana/dsconfig/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// examplesDir returns the path to the examples directory relative to the
// test file. This allows `go test` to find examples from the schema/ dir.
func examplesDir() string {
	return filepath.Join("examples")
}

// loadExample reads a JSON example file, unmarshals it into a
// DatasourceConfigSchema, and returns it. Fails the test on any error.
func loadExample(t *testing.T, filename string) *schema.DatasourceConfigSchema {
	t.Helper()
	path := filepath.Join(examplesDir(), filename)
	data, err := os.ReadFile(path)
	require.NoError(t, err, "failed to read example file: %s", filename)

	var s schema.DatasourceConfigSchema
	require.NoError(t, json.Unmarshal(data, &s), "failed to unmarshal: %s", filename)
	return &s
}

// TestExampleFiles_GoValidation loads each example JSON file, unmarshals
// it into the Go schema struct, and runs Validate(). This ensures the
// examples are valid according to the Go contract and that Go's JSON
// tags correctly map to the schema field names.
func TestExampleFiles_GoValidation(t *testing.T) {
	examples := []struct {
		file        string
		description string
		fieldCount  int // expected number of field IDs (including item fields)
	}{
		{
			file:        "simple-url.schema.json",
			description: "Minimal schema: single URL field with pattern validation",
			fieldCount:  1,
		},
		{
			file:        "bearer-token.schema.json",
			description: "Auth schema: URL + auth method select + secure bearer token with conditional visibility",
			fieldCount:  3,
		},
		{
			file:        "indexed-headers.schema.json",
			description: "Storage mapping: array of header objects with indexedPair mapping to legacy jsonData/secureJsonData keys",
			fieldCount:  4, // url + httpHeaders + 2 item fields
		},
		{
			file:        "virtual-auth.schema.json",
			description: "Virtual fields: basic auth with computed virtual field and pair relationship",
			fieldCount:  5,
		},
		{
			file:        "array-of-objects.schema.json",
			description: "Array of objects: trace-to-metrics queries with item fields and textarea UI",
			fieldCount:  4, // url + tracesToMetrics + 2 item fields
		},
	}

	for _, tc := range examples {
		t.Run(tc.file, func(t *testing.T) {
			s := loadExample(t, tc.file)

			// Validate the full schema
			require.NoError(t, s.Validate(), "%s: %s", tc.file, tc.description)

			// Verify field ID count matches expectation
			ids, err := s.FieldIDs()
			require.NoError(t, err)
			assert.Len(t, ids, tc.fieldCount, "%s: expected %d field IDs", tc.file, tc.fieldCount)

			// Verify root fields were populated
			assert.NotEmpty(t, s.SchemaVersion)
			assert.NotEmpty(t, s.PluginType)
			assert.NotEmpty(t, s.PluginName)
		})
	}
}

// TestExampleFiles_GoRoundTrip verifies that each example survives
// Go unmarshal -> marshal -> unmarshal and still validates. This catches
// serialization issues where Go omitempty or field naming mismatches
// could silently drop data.
func TestExampleFiles_GoRoundTrip(t *testing.T) {
	files, err := filepath.Glob(filepath.Join(examplesDir(), "*.schema.json"))
	require.NoError(t, err)
	require.NotEmpty(t, files, "no example files found")

	for _, path := range files {
		name := filepath.Base(path)
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(path)
			require.NoError(t, err)

			// First pass: unmarshal + validate
			var original schema.DatasourceConfigSchema
			require.NoError(t, json.Unmarshal(data, &original))
			require.NoError(t, original.Validate())

			// Round-trip: marshal + unmarshal
			roundTripped, err := json.Marshal(&original)
			require.NoError(t, err)

			var decoded schema.DatasourceConfigSchema
			require.NoError(t, json.Unmarshal(roundTripped, &decoded))
			require.NoError(t, decoded.Validate())

			// Verify key properties survived
			assert.Equal(t, original.SchemaVersion, decoded.SchemaVersion)
			assert.Equal(t, original.PluginType, decoded.PluginType)
			assert.Len(t, decoded.Fields, len(original.Fields))
		})
	}
}
