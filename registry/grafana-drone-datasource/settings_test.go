package dronedatasource

import (
	"strings"
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

func TestLoadConfig(t *testing.T) {
	t.Run("default example fails validation (empty token+url)", func(t *testing.T) {
		if _, err := LoadConfig(t.Context(), backend.DataSourceInstanceSettings{
			JSONData: []byte(`{"services":{"drone":{"auth":{"id":"auth_bearer"}}}}`),
		}); err == nil || !strings.Contains(err.Error(), "is required") {
			t.Fatalf("expected required error, got %v", err)
		}
	})
	t.Run("valid config loads", func(t *testing.T) {
		cfg, err := LoadConfig(t.Context(), backend.DataSourceInstanceSettings{
			JSONData:                []byte(`{"services":{"drone":{"auth":{"id":"auth_bearer"}}},"variables":{"url":"https://drone.example.com"}}`),
			DecryptedSecureJSONData: map[string]string{"drone.token": "tok"},
		})
		if err != nil {
			t.Fatalf("LoadConfig: %v", err)
		}
		if cfg.Services.Drone.Auth.Id != AuthMethodAuthBearer {
			t.Errorf("auth.id = %q, want %q", cfg.Services.Drone.Auth.Id, AuthMethodAuthBearer)
		}
		if cfg.Variables.URL != "https://drone.example.com" {
			t.Errorf("url = %q", cfg.Variables.URL)
		}
		if cfg.DecryptedSecureJSONData[SecureJsonDataKeyToken] != "tok" {
			t.Errorf("token = %q, want tok", cfg.DecryptedSecureJSONData[SecureJsonDataKeyToken])
		}
	})
	t.Run("invalid jsonData errors", func(t *testing.T) {
		if _, err := LoadConfig(t.Context(), backend.DataSourceInstanceSettings{JSONData: []byte(`{`)}); err == nil || !strings.Contains(err.Error(), "parse jsonData") {
			t.Fatalf("expected parse error, got %v", err)
		}
	})
}

func TestValidate(t *testing.T) {
	ok := Config{
		Services:                ServicesConfig{Drone: ServiceConfig{Auth: AuthConfig{Id: AuthMethodAuthBearer}}},
		Variables:               VariablesConfig{URL: "https://drone.example.com"},
		DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyToken: "tok"},
	}
	if err := ok.Validate(); err != nil {
		t.Fatalf("Validate ok: %v", err)
	}
	noURL := ok
	noURL.Variables.URL = ""
	if err := noURL.Validate(); err == nil || !strings.Contains(err.Error(), "url is required") {
		t.Fatalf("Validate no url: %v", err)
	}
}
