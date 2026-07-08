package helloworlddatasource

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// settingsFromExample converts a SettingsExamples entry (a full instance
// settings object with jsonData and secureJsonData) into the
// backend.DataSourceInstanceSettings shape LoadConfig expects.
func settingsFromExample(t *testing.T, exampleKey string) backend.DataSourceInstanceSettings {
	t.Helper()
	ex, ok := SettingsExamples().Examples[exampleKey]
	if !ok {
		t.Fatalf("unknown example %q", exampleKey)
	}
	value, ok := ex.Value.(map[string]any)
	if !ok {
		t.Fatalf("example %q value is not an object", exampleKey)
	}
	jsonData, err := json.Marshal(value["jsonData"])
	if err != nil {
		t.Fatalf("marshal jsonData: %v", err)
	}
	secure := map[string]string{}
	if raw, ok := value["secureJsonData"].(map[string]any); ok {
		for k, v := range raw {
			s, _ := v.(string)
			secure[k] = s
		}
	}
	return backend.DataSourceInstanceSettings{
		JSONData:                jsonData,
		DecryptedSecureJSONData: secure,
	}
}

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name        string
		example     string
		settings    backend.DataSourceInstanceSettings
		useSettings bool
		wantErr     string // empty = expect no error; otherwise substring match
		wantAPIKey  string // expected value stored under the placeholder apiKey (if any)
		wantHasKey  bool   // whether the placeholder key should be present
	}{
		{
			name:       "default example loads",
			example:    "",
			wantHasKey: true, // the "" example carries an (empty) apiKey secret
			wantAPIKey: "",
		},
		{
			name:        "empty jsonData is accepted",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{}`),
			},
		},
		{
			name:        "nil jsonData is accepted",
			useSettings: true,
			settings:    backend.DataSourceInstanceSettings{},
		},
		{
			name:        "unknown jsonData keys are ignored",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"foo":"bar","nested":{"a":1}}`),
			},
		},
		{
			name:        "placeholder secret is copied through",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{}`),
				DecryptedSecureJSONData: map[string]string{
					"apiKey": "<placeholder-not-read-by-plugin>",
				},
			},
			wantHasKey: true,
			wantAPIKey: "<placeholder-not-read-by-plugin>",
		},
		{
			name:        "unknown secret keys are ignored",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{}`),
				DecryptedSecureJSONData: map[string]string{
					"somethingElse": "<value>",
				},
			},
		},
		{
			name:        "malformed jsonData errors",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{`),
			},
			wantErr: "parse jsonData",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings := tt.settings
			if !tt.useSettings {
				settings = settingsFromExample(t, tt.example)
			}

			cfg, err := LoadConfig(t.Context(), settings)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("LoadConfig: expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("LoadConfig: error = %q, want it to contain %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("LoadConfig: %v", err)
			}
			if cfg.DecryptedSecureJSONData == nil {
				t.Fatalf("DecryptedSecureJSONData must be initialized, got nil")
			}

			got, present := cfg.DecryptedSecureJSONData[SecureJsonDataKeyAPIKey]
			if present != tt.wantHasKey {
				t.Errorf("apiKey present = %v, want %v", present, tt.wantHasKey)
			}
			if tt.wantHasKey && got != tt.wantAPIKey {
				t.Errorf("apiKey = %q, want %q", got, tt.wantAPIKey)
			}
		})
	}
}

func TestApplyDefaults(t *testing.T) {
	// ApplyDefaults is intentionally a no-op for Hello World — the editor
	// persists no fields and writes no defaults. This test guards that no
	// field is silently defaulted.
	in := Config{
		DecryptedSecureJSONData: map[SecureJsonDataKey]string{
			SecureJsonDataKeyAPIKey: "<value>",
		},
	}
	got := Config{
		DecryptedSecureJSONData: map[SecureJsonDataKey]string{
			SecureJsonDataKeyAPIKey: "<value>",
		},
	}
	got.ApplyDefaults()
	if !reflect.DeepEqual(in, got) {
		t.Errorf("ApplyDefaults mutated Config: %#v -> %#v", in, got)
	}
}

func TestValidate(t *testing.T) {
	// Hello World requires no configuration, so Validate always succeeds —
	// including for a zero-value Config.
	cases := []Config{
		{},
		{DecryptedSecureJSONData: map[SecureJsonDataKey]string{}},
		{DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyAPIKey: "<value>"}},
	}
	for i, c := range cases {
		if err := c.Validate(); err != nil {
			t.Errorf("case %d: Validate() = %v, want nil", i, err)
		}
	}
}

// TestPluginID guards that the entry's plugin id matches the schema's
// pluginType and the upstream plugin.json id / backend PluginId constant.
func TestPluginID(t *testing.T) {
	if PluginID != "grafana-helloworld-datasource" {
		t.Fatalf("PluginID = %q, want %q", PluginID, "grafana-helloworld-datasource")
	}
	cfg, err := ConfigSchema()
	if err != nil {
		t.Fatalf("ConfigSchema: %v", err)
	}
	if cfg.PluginType != PluginID {
		t.Errorf("schema pluginType = %q, want %q", cfg.PluginType, PluginID)
	}
}
