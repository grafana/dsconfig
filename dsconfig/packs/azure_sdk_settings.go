package packs

import (
	_ "embed"

	"github.com/grafana/dsconfig/dsconfig"
)

// azure_sdk_settings.json defines the fields for grafana-azure-sdk-go:
// Azure credential fields and related settings.
// Field IDs are namespaced with the "azure_sdk_settings." prefix.
//
//go:embed azure_sdk_settings.json
var azureSDKSettingsJSON []byte

func init() {
	mustLoadPack(dsconfig.PackAzureSDKSettings, azureSDKSettingsJSON)
}
