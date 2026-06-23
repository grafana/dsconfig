package schema

import (
	"github.com/grafana/dsconfig/dsconfig"
	"github.com/grafana/grafana-plugin-sdk-go/experimental/pluginschema"
)

// TargetAPIVersion is the default API version used when assembling a plugin schema.
//
// Deprecated: use dsconfig.TargetAPIVersion.
const TargetAPIVersion = dsconfig.TargetAPIVersion

// NewPluginSchema assembles a full SDK PluginSchema from the declarative single
// source of truth, with optional settings examples.
//
// Deprecated: use dsconfig.NewPluginSchema. The signature is preserved for
// backward compatibility; the error return is always nil.
func NewPluginSchema(settings *pluginschema.Settings, examples *pluginschema.SettingsExamples) (*pluginschema.PluginSchema, error) {
	return dsconfig.NewPluginSchema(settings, examples), nil
}
