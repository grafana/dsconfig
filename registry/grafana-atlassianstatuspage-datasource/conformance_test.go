package atlassianstatuspagedatasource

import (
	"flag"
	"testing"

	"github.com/grafana/dsconfig/dsconfig"
	"github.com/grafana/dsconfig/schema"
	"github.com/stretchr/testify/require"
)

// TestSchemaConformance runs the dsconfig conformance suite against this entry's
// dsconfig.json. Invoke with -generateArtifacts to (re)write the committed
// .gen.json artifacts.
//
// This datasource has NO authentication and therefore no secureJsonData. The
// shared schema.RunPluginTests requires at least one secure key (and its
// SchemaRoundTrip subtest asserts a non-empty SecureValues), so this entry runs
// the applicable subset of the conformance checks directly and skips only the
// secure-value assertion. Every other guard rail (schema validity, artifact
// drift, spec/secure separation, jsonData⇔struct parity) still applies.
func TestSchemaConformance(t *testing.T) {
	pluginSchema := schema.MustNewSDKSchema(t, configSchemaJSON, SettingsExamples())

	if generateArtifactsFlag() {
		require.NoError(t, schema.WriteArtifacts(pluginSchema))
		return
	}

	cfg, err := dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
	require.NoError(t, err)

	p := schema.Params{
		PluginID:          PluginID,
		DSConfigSchema:    cfg,
		PluginSchema:      pluginSchema,
		SettingsJSONModel: Config{},
		SecureKeys:        nil, // no secrets
	}

	t.Run("BaseFieldsResolved", func(t *testing.T) { schema.BaseFieldsResolved(t, p) })
	t.Run("SchemaArtifactInSync", func(t *testing.T) { schema.SchemaArtifactInSync(t, p) })
	t.Run("SchemaSpecHasNoSecureJSON", func(t *testing.T) { schema.SchemaSpecHasNoSecureJSON(t, p) })
	t.Run("ConfigSchemaValid", func(t *testing.T) { schema.ConfigSchemaValid(t, p) })
	t.Run("JSONDataMatchesStruct", func(t *testing.T) { schema.JSONDataMatchesStruct(t, p) })
	t.Run("JSONDataTypesMatchStruct", func(t *testing.T) { schema.JSONDataTypesMatchStruct(t, p) })
	t.Run("SecureValuesMatchLoadSettings", func(t *testing.T) { schema.SecureValuesMatchLoadSettings(t, p) })
	// SchemaRoundTrip is intentionally skipped: it asserts a non-empty
	// SecureValues, which does not hold for a datasource with no secrets.
}

// generateArtifactsFlag reports whether `go test -generateArtifacts` was passed.
// The flag is registered by the shared schema package (imported above); this
// entry reads it via flag.Lookup instead of redefining it.
func generateArtifactsFlag() bool {
	f := flag.Lookup("generateArtifacts")
	return f != nil && f.Value.String() == "true"
}
