package clickhousedatasource

import (
	"testing"

	"github.com/grafana/dsconfig/schema"
)

// TestSchemaConformance runs the shared conformance suite against this
// entry's dsconfig.json using schema.RunPluginTests. Invoke with
// -generateArtifacts to (re)write the committed schema.gen.json /
// settings.gen.json / settings.examples.gen.json artifacts; without the flag
// it runs the full guard-rail suite (schema round-trip, artifact drift,
// spec/secure separation, jsonData/struct parity in both directions,
// secure-key parity).
//
// Config itself is the settings model: its json-tagged fields are the
// jsonData shape; its non-jsonData fields (SecureHttpHeaders,
// DecryptedSecureJSONData) are tagged json:"-" and skipped by the conformance
// walker. The dsconfig schema uses `section: "logs"` / `section: "traces"`
// for the nested Logs/Traces structs so the walker descends into them.
func TestSchemaConformance(t *testing.T) {
	secureKeys := make([]string, 0, len(SecureJsonDataKeys))
	for _, k := range SecureJsonDataKeys {
		secureKeys = append(secureKeys, string(k))
	}
	schema.RunPluginTests(t, schema.PluginUnderTest{
		ID:                PluginID,
		ConfigSchemaJSON:  configSchemaJSON,
		SettingsJSONModel: Config{},
		SecureKeys:        secureKeys,
		SettingsExamples:  SettingsExamples(),
	})
}
