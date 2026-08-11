package packs

import (
	"testing"

	"github.com/grafana/dsconfig/dsconfig"
)

// TestSqlengSettingsPackValidates ensures the embedded sqleng_settings.json
// produces a valid Schema after ResolveBaseFields(). mustLoadPack only checks
// that the JSON parses; field-level validation only runs when the pack is
// resolved into a Schema, so this test guards against authoring errors
// (unknown roles, bad UI components, invalid validation rules, effect
// references to unknown field IDs, etc.).
func TestSqlengSettingsPackValidates(t *testing.T) {
	s := &dsconfig.Schema{
		SchemaVersion: "1.0.0",
		PluginType:    "test-sqleng-datasource",
		PluginName:    "Test sqleng Datasource",
		BaseFields: []dsconfig.BaseFieldRef{
			{From: dsconfig.PackSqlengSettings},
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
		t.Fatalf("sqleng_settings pack failed schema validation: %v", err)
	}

	// Sanity-check that the expected canonical field IDs are present.
	// These mirror the common fields in the shared `sqleng.JsonData` struct
	// (pkg/.../sqleng/sql_engine.go in postgres/mysql/mssql) and the root-
	// level + secure fields rendered by each plugin's ConfigurationEditor.tsx.
	want := []string{
		"sqleng_settings.url",
		"sqleng_settings.user",
		"sqleng_settings.password",
		"sqleng_settings.database",
		"sqleng_settings.timeInterval",
		"sqleng_settings.connectionTimeout",
		"sqleng_settings.maxOpenConns",
		"sqleng_settings.maxIdleConns",
		"sqleng_settings.connMaxLifetime",
		"sqleng_settings.tlsSkipVerify",
		"sqleng_settings.tlsConfigurationMethod",
		"sqleng_settings.sslRootCertFile",
		"sqleng_settings.sslCertFile",
		"sqleng_settings.sslKeyFile",
		"sqleng_settings.servername",
		"sqleng_settings.tlsCACert",
		"sqleng_settings.tlsClientCert",
		"sqleng_settings.tlsClientKey",
		"sqleng_settings.enableSecureSocksProxy",
		"sqleng_settings.secureSocksProxyUsername",
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
			t.Errorf("sqleng_settings pack is missing expected field %q", id)
		}
	}
}
