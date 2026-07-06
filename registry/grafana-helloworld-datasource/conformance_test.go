package helloworlddatasource

import (
	"testing"

	"github.com/grafana/dsconfig/schema"
)

// TestSchemaConformance runs the shared conformance suite against this entry's
// dsconfig.json using schema.RunPluginTests. Invoke with -generateArtifacts to
// (re)write the committed schema.gen.json / settings.gen.json /
// settings.examples.gen.json artifacts; without the flag it runs the full guard-
// rail suite (schema round-trip, artifact drift, spec/secure separation,
// jsonData/struct parity in both directions, secure-key parity).
//
// The Config struct is the settings model: it has no json-tagged jsonData
// fields (the Hello World plugin stores no jsonData), so the JSONDataMatchesStruct
// check compares two empty sets. Its only field, DecryptedSecureJSONData, is
// tagged json:"-" and is skipped by the conformance walker. SecureKeys carries
// the single placeholder key (apiKey) so RunPluginTests' non-empty-SecureKeys
// precondition and SchemaRoundTrip's non-empty-SecureValues assertion are
// satisfied (see the entry README for why a placeholder is required).
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
