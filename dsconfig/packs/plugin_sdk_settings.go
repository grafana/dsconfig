package packs

import (
	_ "embed"

	"github.com/grafana/dsconfig/dsconfig"
)

// plugin_sdk_settings.json defines the fields for grafana-plugin-sdk-go:
// URL, basicAuth, TLS settings, timeout, and HTTP headers.
// Field IDs are namespaced with the "plugin_sdk_settings." prefix.
//
//go:embed plugin_sdk_settings.json
var pluginSDKSettingsJSON []byte

func init() {
	mustLoadPack(dsconfig.PackPluginSDKSettings, pluginSDKSettingsJSON)
}
