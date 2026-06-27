package packs

import (
	"testing"

	"github.com/grafana/dsconfig/dsconfig"
)

// TestAzureSDKSettingsPackValidates ensures the embedded azure_sdk_settings.json
// produces a valid Schema after ResolveBaseFields(). mustLoadPack only checks
// that the JSON parses; field-level validation only runs when the pack is
// resolved into a Schema, so this test guards against authoring errors
// (unknown roles, bad UI components, invalid validation rules, effect
// references to unknown field IDs, etc.).
func TestAzureSDKSettingsPackValidates(t *testing.T) {
	s := &dsconfig.Schema{
		SchemaVersion: "1.0.0",
		PluginType:    "test-azure-datasource",
		PluginName:    "Test Azure Datasource",
		BaseFields: []dsconfig.BaseFieldRef{
			{From: dsconfig.PackAzureSDKSettings},
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
		t.Fatalf("azure_sdk_settings pack failed schema validation: %v", err)
	}

	// Sanity-check that the expected canonical field IDs are present.
	want := []string{
		"azure_sdk_settings.authType",
		"azure_sdk_settings.azureCloud",
		"azure_sdk_settings.tenantId",
		"azure_sdk_settings.clientId",
		"azure_sdk_settings.azureClientSecret",
		"azure_sdk_settings.certificateFormat",
		"azure_sdk_settings.clientCertificate",
		"azure_sdk_settings.privateKey",
		"azure_sdk_settings.certificatePassword",
		"azure_sdk_settings.userId",
		"azure_sdk_settings.password",
		"azure_sdk_settings.serviceCredentialsEnabled",
		"azure_sdk_settings.oauthPassThru",
	}

	have := map[string]bool{}
	for _, f := range resolved.Fields {
		have[f.ID] = true
	}
	for _, id := range want {
		if !have[id] {
			t.Errorf("azure_sdk_settings pack is missing expected field %q", id)
		}
	}
}
