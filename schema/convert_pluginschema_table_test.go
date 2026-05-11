package schema_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grafana/dsconfig/schema"
)

// TestToPluginSettings_Table runs file-based table tests.
// Each subdirectory under testdata/convert/ contains:
//   - input.json:  DatasourceConfigSchema
//   - output.json: expected PluginSettings
//   - config.json: example Grafana storage model (documentation only, not asserted)
func TestToPluginSettings_Table(t *testing.T) {
	testdataDir := filepath.Join("testdata", "convert")

	entries, err := os.ReadDir(testdataDir)
	require.NoError(t, err, "failed to read testdata/convert directory")

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		dir := filepath.Join(testdataDir, name)

		// Skip directories without fixture files (shouldn't happen, but be safe).
		if _, err := os.Stat(filepath.Join(dir, "input.json")); os.IsNotExist(err) {
			continue
		}

		t.Run(name, func(t *testing.T) {
			inputBytes, err := os.ReadFile(filepath.Join(dir, "input.json"))
			require.NoError(t, err)

			expectedBytes, err := os.ReadFile(filepath.Join(dir, "output.json"))
			require.NoError(t, err)

			// Parse input schema
			var inputSchema schema.DatasourceConfigSchema
			require.NoError(t, json.Unmarshal(inputBytes, &inputSchema))

			// Convert
			got, err := inputSchema.ToPluginSettings()
			require.NoError(t, err, "ToPluginSettings() returned error")

			// Parse expected output
			var expected schema.PluginSettings
			require.NoError(t, json.Unmarshal(expectedBytes, &expected))

			// Marshal both to JSON for comparison (normalises field ordering).
			gotJSON, err := json.Marshal(got)
			require.NoError(t, err)
			expectedJSON, err := json.Marshal(expected)
			require.NoError(t, err)

			// Unmarshal back into generic maps so the diff is readable.
			var gotMap, expectedMap map[string]any
			require.NoError(t, json.Unmarshal(gotJSON, &gotMap))
			require.NoError(t, json.Unmarshal(expectedJSON, &expectedMap))

			assert.Equal(t, expectedMap, gotMap)
		})
	}
}
