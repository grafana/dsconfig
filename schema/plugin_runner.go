package schema

import (
	"errors"
	"flag"
	"strings"
	"testing"

	"github.com/grafana/dsconfig/dsconfig"
	sdkSchema "github.com/grafana/grafana-plugin-sdk-go/experimental/pluginschema"
	"github.com/stretchr/testify/require"
)

// generateArtifacts is set by `go test -generateArtifacts`, which is invoked
// via the plugin's `go generate` directive to (re)write the committed schema
// artifacts. When the flag is not set, RunPluginTests runs the conformance
// suite instead.
var generateArtifacts = flag.Bool("generateArtifacts", false, "write the schema artifacts to disk instead of running tests")

// PluginUnderTest bundles the four plugin-specific inputs needed to drive both
// the artifact generator and the conformance suite. Every field is required.
type PluginUnderTest struct {
	// ID is the datasource plugin type (for example "grafana-athena-datasource").
	ID string

	// ConfigSchemaJSON is the raw bytes of the plugin's dsconfig.json, typically
	// supplied via //go:embed.
	ConfigSchemaJSON []byte

	// SettingsJSONModel is a zero value of the Go struct that backs jsonData
	// (for example models.InfinitySettingsJson{}). Its json tags are compared
	// against the schema's jsonData fields.
	SettingsJSONModel any

	// SecureKeys are the secureJsonData keys the plugin actually reads when
	// loading settings. They are compared against the schema's secure values.
	SecureKeys []string

	// SettingsExamples is optional. If nil, an empty SettingsExamples is used.
	SettingsExamples *sdkSchema.SettingsExamples
}

func (p PluginUnderTest) validate() error {
	if strings.TrimSpace(p.ID) == "" {
		return errors.New("ID is empty")
	}
	if len(p.ConfigSchemaJSON) == 0 {
		return errors.New("ConfigSchemaJSON is empty")
	}
	if p.SettingsJSONModel == nil {
		return errors.New("SettingsJSONModel is nil")
	}
	if len(p.SecureKeys) == 0 {
		return errors.New("SecureKeys is empty")
	}
	return nil
}

// RunPluginTests is the one-call test entry point for plugins built on the
// dsconfig single source of truth. When invoked with -generateArtifacts, it
// writes the schema artifacts to disk; otherwise it runs the full conformance
// suite. Call it from a single Test function in the plugin's test package.
func RunPluginTests(t *testing.T, p PluginUnderTest) {
	t.Helper()
	require.NotNil(t, p)
	require.NoError(t, p.validate(), "invalid PluginUnderTest params")
	examples := p.SettingsExamples
	if examples == nil {
		examples = &sdkSchema.SettingsExamples{}
	}
	pluginSchema := MustNewSDKSchema(t, p.ConfigSchemaJSON, examples)

	if *generateArtifacts {
		require.NoError(t, WriteArtifacts(pluginSchema))
		return
	}

	cfg, err := dsconfig.ParseSchemaJSON(p.ConfigSchemaJSON)
	require.NoError(t, err)
	RunConformanceTests(t, Params{
		PluginID:          p.ID,
		DSConfigSchema:    cfg,
		PluginSchema:      pluginSchema,
		SettingsJSONModel: p.SettingsJSONModel,
		SecureKeys:        p.SecureKeys,
	})
}
