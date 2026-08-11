package packs

import (
	"testing"

	"github.com/grafana/dsconfig/dsconfig"
)

// TestAWSSDKSettingsPackValidates ensures the embedded aws_sdk_settings.json
// produces a valid Schema after ResolveBaseFields(). mustLoadPack only checks
// that the JSON parses; field-level validation only runs when the pack is
// resolved into a Schema, so this test guards against authoring errors
// (unknown roles, bad UI components, invalid validation rules, effect
// references to unknown field IDs, etc.).
func TestAWSSDKSettingsPackValidates(t *testing.T) {
	s := &dsconfig.Schema{
		SchemaVersion: "1.0.0",
		PluginType:    "test-aws-datasource",
		PluginName:    "Test AWS Datasource",
		BaseFields: []dsconfig.BaseFieldRef{
			{From: dsconfig.PackAWSSDKSettings},
		},
		// Plugins must declare at least one field of their own; provide a
		// minimal stub so Validate() does not fail on len(Fields)==0.
		Fields: []dsconfig.ConfigField{
			{
				ID:        "stub",
				Key:       "stub",
				ValueType: dsconfig.StringType,
				Target: func() *dsconfig.TargetLocation {
					t := dsconfig.JSONDataTarget
					return &t
				}(),
			},
		},
	}

	resolved, err := s.ResolveBaseFields()
	if err != nil {
		t.Fatalf("ResolveBaseFields failed: %v", err)
	}

	if err := resolved.Validate(); err != nil {
		t.Fatalf("aws_sdk_settings pack failed schema validation: %v", err)
	}

	// Sanity-check that the expected canonical field IDs are present.
	want := []string{
		"aws_sdk_settings.authType",
		"aws_sdk_settings.profile",
		"aws_sdk_settings.accessKey",
		"aws_sdk_settings.secretKey",
		"aws_sdk_settings.sessionToken",
		"aws_sdk_settings.assumeRoleArn",
		"aws_sdk_settings.externalId",
		"aws_sdk_settings.endpoint",
		"aws_sdk_settings.defaultRegion",
		"aws_sdk_settings.proxyType",
		"aws_sdk_settings.proxyUrl",
		"aws_sdk_settings.proxyUsername",
		"aws_sdk_settings.proxyPassword",
	}

	have := map[string]bool{}
	for _, f := range resolved.Fields {
		have[f.ID] = true
	}
	for _, id := range want {
		if !have[id] {
			t.Errorf("aws_sdk_settings pack is missing expected field %q", id)
		}
	}
}
