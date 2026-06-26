package packs

import (
	_ "embed"

	"github.com/grafana/dsconfig/dsconfig"
)

// google_sdk_settings.json defines the fields for grafana-google-sdk-go:
// Google credential fields and related settings.
// Field IDs are namespaced with the "google_sdk_settings." prefix.
//
//go:embed google_sdk_settings.json
var googleSDKSettingsJSON []byte

func init() {
	mustLoadPack(dsconfig.PackGoogleSDKSettings, googleSDKSettingsJSON)
}
