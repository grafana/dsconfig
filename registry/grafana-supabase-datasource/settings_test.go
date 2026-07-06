package supabasedatasource

import (
	"strings"
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

func TestLoadConfig(t *testing.T) {
	t.Run("default example fails validation (empty token)", func(t *testing.T) {
		if _, err := LoadConfig(t.Context(), backend.DataSourceInstanceSettings{
			JSONData: []byte(`{"services":{"mgmt":{"auth":{"id":"mgmt_bearer"}}}}`),
		}); err == nil || !strings.Contains(err.Error(), "is required") {
			t.Fatalf("expected token required error, got %v", err)
		}
	})
	t.Run("valid config loads and defaults auth id", func(t *testing.T) {
		cfg, err := LoadConfig(t.Context(), backend.DataSourceInstanceSettings{
			DecryptedSecureJSONData: map[string]string{"mgmt.token": "tok"},
		})
		if err != nil {
			t.Fatalf("LoadConfig: %v", err)
		}
		if cfg.Services.Mgmt.Auth.Id != AuthMethodMgmtBearer {
			t.Errorf("auth.id = %q, want %q", cfg.Services.Mgmt.Auth.Id, AuthMethodMgmtBearer)
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
		Services:                ServicesConfig{Mgmt: ServiceConfig{Auth: AuthConfig{Id: AuthMethodMgmtBearer}}},
		DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyToken: "tok"},
	}
	if err := ok.Validate(); err != nil {
		t.Fatalf("Validate ok: %v", err)
	}
	if err := (Config{Services: ServicesConfig{Mgmt: ServiceConfig{Auth: AuthConfig{Id: "bogus"}}}}).Validate(); err == nil || !strings.Contains(err.Error(), "unknown auth method") {
		t.Fatalf("Validate bogus: %v", err)
	}
}
