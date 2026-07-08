package postgresqldatasource

import (
	"testing"

	"github.com/grafana/dsconfig/schema"
)

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
