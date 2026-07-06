package verceldatasource

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

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
	return backend.DataSourceInstanceSettings{JSONData: jsonData, DecryptedSecureJSONData: secure}
}

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name      string
		example   string
		settings  backend.DataSourceInstanceSettings
		wantErr   string
		wantAuth  AuthMethodID
		wantToken string
		wantTeam  string
	}{
		{name: "default example fails validation (empty token)", example: "", wantErr: "is required"},
		{name: "access token example loads", example: "accessToken", wantAuth: AuthMethodVercelAPIKey, wantToken: "<vercel-access-token>", wantTeam: "team_1a2b3c4d5e6f7g8h9i0j1k2l"},
		{name: "empty settings default and fail validation", settings: backend.DataSourceInstanceSettings{}, wantErr: "is required"},
		{name: "invalid jsonData errors", settings: backend.DataSourceInstanceSettings{JSONData: []byte(`{`)}, wantErr: "parse jsonData"},
		{
			name: "explicit config loads without team",
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"services":{"vercel":{"auth":{"id":"vercelApiKey"}}}}`),
				DecryptedSecureJSONData: map[string]string{"vercel.token": "tok"},
			},
			wantAuth: AuthMethodVercelAPIKey, wantToken: "tok",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings := tt.settings
			if tt.example != "" || (tt.settings.JSONData == nil && tt.settings.DecryptedSecureJSONData == nil && tt.wantErr == "") {
				if _, ok := SettingsExamples().Examples[tt.example]; ok {
					settings = settingsFromExample(t, tt.example)
				}
			}
			cfg, err := LoadConfig(t.Context(), settings)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("LoadConfig error = %v, want contains %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("LoadConfig: %v", err)
			}
			if tt.wantAuth != "" && cfg.Services.Vercel.Auth.Id != tt.wantAuth {
				t.Errorf("auth.id = %q, want %q", cfg.Services.Vercel.Auth.Id, tt.wantAuth)
			}
			if tt.wantToken != "" && cfg.DecryptedSecureJSONData[SecureJsonDataKeyToken] != tt.wantToken {
				t.Errorf("token = %q, want %q", cfg.DecryptedSecureJSONData[SecureJsonDataKeyToken], tt.wantToken)
			}
			if tt.wantTeam != "" && cfg.Variables.TeamID != tt.wantTeam {
				t.Errorf("team_id = %q, want %q", cfg.Variables.TeamID, tt.wantTeam)
			}
		})
	}
}

func TestApplyDefaultsAndValidate(t *testing.T) {
	c := Config{}
	c.ApplyDefaults()
	if c.Services.Vercel.Auth.Id != AuthMethodVercelAPIKey {
		t.Fatalf("auth.id = %q, want %q", c.Services.Vercel.Auth.Id, AuthMethodVercelAPIKey)
	}
	if err := c.Validate(); err == nil || !strings.Contains(err.Error(), "is required") {
		t.Fatalf("Validate empty token: %v, want token required", err)
	}
	c.DecryptedSecureJSONData = map[SecureJsonDataKey]string{SecureJsonDataKeyToken: "tok"}
	if err := c.Validate(); err != nil {
		t.Fatalf("Validate with token: %v", err)
	}
	c.Services.Vercel.Auth.Id = "bogus"
	if err := c.Validate(); err == nil || !strings.Contains(err.Error(), "unknown auth method") {
		t.Fatalf("Validate bogus auth: %v", err)
	}
}
