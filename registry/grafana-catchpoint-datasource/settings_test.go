package catchpointdatasource

import (
	"strings"
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

func TestLoadConfig(t *testing.T) {
	t.Run("default example fails validation (empty key)", func(t *testing.T) {
		if _, err := LoadConfig(t.Context(), backend.DataSourceInstanceSettings{
			JSONData: []byte(`{"services":{"catchpoint":{"auth":{"id":"bearer_token"}}}}`),
		}); err == nil || !strings.Contains(err.Error(), "is required") {
			t.Fatalf("expected key required error, got %v", err)
		}
	})
	t.Run("valid config loads and defaults auth id", func(t *testing.T) {
		cfg, err := LoadConfig(t.Context(), backend.DataSourceInstanceSettings{
			DecryptedSecureJSONData: map[string]string{"catchpoint.token": "tok"},
		})
		if err != nil {
			t.Fatalf("LoadConfig: %v", err)
		}
		if cfg.Services.Catchpoint.Auth.Id != AuthMethodBearerToken {
			t.Errorf("auth.id = %q, want %q", cfg.Services.Catchpoint.Auth.Id, AuthMethodBearerToken)
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
		Services:                ServicesConfig{Catchpoint: ServiceConfig{Auth: AuthConfig{Id: AuthMethodBearerToken}}},
		DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyToken: "tok"},
	}
	if err := ok.Validate(); err != nil {
		t.Fatalf("Validate ok: %v", err)
	}
	if err := (Config{Services: ServicesConfig{Catchpoint: ServiceConfig{Auth: AuthConfig{Id: "bogus"}}}}).Validate(); err == nil || !strings.Contains(err.Error(), "unknown auth method") {
		t.Fatalf("Validate bogus: %v", err)
	}
}
