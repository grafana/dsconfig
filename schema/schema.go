package schema

import (
	"github.com/grafana/grafana-plugin-sdk-go/experimental/pluginschema"
)

// NewPluginSchema assembles a full SDK PluginSchema from the declarative single
// source of truth, for the given API version and optional settings examples.
func NewPluginSchema(apiVersion string, settingsSchema *pluginschema.Settings, settingsExamples *pluginschema.SettingsExamples) (*pluginschema.PluginSchema, error) {
	out := pluginschema.PluginSchema{}
	if apiVersion != "" {
		out.TargetAPIVersion = apiVersion
	}
	if settingsSchema != nil {
		out.SettingsSchema = settingsSchema
	}
	if settingsExamples != nil {
		out.SettingsExamples = settingsExamples
	}
	return &out, nil
}
