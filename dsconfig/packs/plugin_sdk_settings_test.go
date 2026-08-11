package packs

import (
	"testing"

	"github.com/grafana/dsconfig/dsconfig"
)

// TestPluginSDKSettingsPackValidates ensures the embedded
// plugin_sdk_settings.json produces a valid Schema after ResolveBaseFields().
// mustLoadPack only checks that the JSON parses; field-level validation only
// runs when the pack is resolved into a Schema, so this test guards against
// authoring errors (unknown roles, bad UI components, invalid validation
// rules, effect references to unknown field IDs, etc.).
func TestPluginSDKSettingsPackValidates(t *testing.T) {
	s := &dsconfig.Schema{
		SchemaVersion: "1.0.0",
		PluginType:    "test-plugin-sdk-datasource",
		PluginName:    "Test Plugin SDK Datasource",
		BaseFields: []dsconfig.BaseFieldRef{
			{From: dsconfig.PackPluginSDKSettings},
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
		t.Fatalf("plugin_sdk_settings pack failed schema validation: %v", err)
	}

	// Sanity-check that the expected canonical field IDs are present.
	// These mirror the HTTP-related options from
	// grafana-plugin-sdk-go/backend/httpclient/options.go and the controls
	// rendered by @grafana/ui DataSourceHttpSettings (and its sub-editors:
	// BasicAuthSettings, HttpProxySettings, TLSAuthSettings,
	// CustomHeadersSettings).
	want := []string{
		"plugin_sdk_settings.url",
		"plugin_sdk_settings.access",
		"plugin_sdk_settings.keepCookies",
		"plugin_sdk_settings.timeout",
		"plugin_sdk_settings.basicAuth",
		"plugin_sdk_settings.basicAuthUser",
		"plugin_sdk_settings.basicAuthPassword",
		"plugin_sdk_settings.withCredentials",
		"plugin_sdk_settings.tlsAuth",
		"plugin_sdk_settings.tlsAuthWithCACert",
		"plugin_sdk_settings.tlsSkipVerify",
		"plugin_sdk_settings.serverName",
		"plugin_sdk_settings.tlsCACert",
		"plugin_sdk_settings.tlsClientCert",
		"plugin_sdk_settings.tlsClientKey",
		"plugin_sdk_settings.oauthPassThru",
		"plugin_sdk_settings.sigV4Auth",
		"plugin_sdk_settings.httpHeaders",
		"plugin_sdk_settings.httpHeaders.item.name",
		"plugin_sdk_settings.httpHeaders.item.value",
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
			t.Errorf("plugin_sdk_settings pack is missing expected field %q", id)
		}
	}
}
