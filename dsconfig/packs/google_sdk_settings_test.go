package packs

import (
	"testing"

	"github.com/grafana/dsconfig/dsconfig"
)

// TestGoogleSDKSettingsPackValidates ensures the embedded
// google_sdk_settings.json produces a valid Schema after ResolveBaseFields().
// mustLoadPack only checks that the JSON parses; field-level validation only
// runs when the pack is resolved into a Schema, so this test guards against
// authoring errors (unknown roles, bad UI components, invalid validation
// rules, effect references to unknown field IDs, etc.).
func TestGoogleSDKSettingsPackValidates(t *testing.T) {
	s := &dsconfig.Schema{
		SchemaVersion: "1.0.0",
		PluginType:    "test-google-sdk-datasource",
		PluginName:    "Test Google SDK Datasource",
		BaseFields: []dsconfig.BaseFieldRef{
			{From: dsconfig.PackGoogleSDKSettings},
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
		t.Fatalf("google_sdk_settings pack failed schema validation: %v", err)
	}

	// Sanity-check that the expected canonical field IDs are present.
	// These mirror the controls rendered by `ConnectionConfig` (and the
	// `AuthConfig`, `JWTForm`, `WIFConfigEditor` components it composes)
	// from grafana-google-sdk-react.
	want := []string{
		"google_sdk_settings.authenticationType",
		"google_sdk_settings.defaultProject",
		"google_sdk_settings.clientEmail",
		"google_sdk_settings.tokenUri",
		"google_sdk_settings.privateKeyPath",
		"google_sdk_settings.privateKey",
		"google_sdk_settings.workloadIdentityPoolProvider",
		"google_sdk_settings.wifServiceAccountEmail",
		"google_sdk_settings.usingImpersonation",
		"google_sdk_settings.serviceAccountToImpersonate",
		"google_sdk_settings.oauthPassThru",
	}

	have := map[string]bool{}
	var collect func(fields []dsconfig.ConfigField)
	collect = func(fields []dsconfig.ConfigField) {
		for _, f := range fields {
			have[f.ID] = true
			if f.Item != nil {
				collect(f.Item.Fields)
			}
		}
	}
	collect(resolved.Fields)
	for _, id := range want {
		if !have[id] {
			t.Errorf("google_sdk_settings pack is missing expected field %q", id)
		}
	}
}
