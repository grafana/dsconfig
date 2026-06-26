package packs

import (
	_ "embed"

	"github.com/grafana/dsconfig/dsconfig"
)

// aws_sdk_settings.json defines the fields for grafana-aws-sdk-go:
// SigV4 authentication, region, profile, and related settings.
// Field IDs are namespaced with the "aws_sdk_settings." prefix.
//
//go:embed aws_sdk_settings.json
var awsSDKSettingsJSON []byte

func init() {
	mustLoadPack(dsconfig.PackAWSSDKSettings, awsSDKSettingsJSON)
}
