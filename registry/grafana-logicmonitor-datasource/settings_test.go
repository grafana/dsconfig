package logicmonitordatasource

import (
	"strings"
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

func TestLoadConfig(t *testing.T) {
	t.Run("default example fails validation (empty token+account)", func(t *testing.T) {
		if _, err := LoadConfig(t.Context(), backend.DataSourceInstanceSettings{
			JSONData: []byte(`{"services":{"logicmonitor":{"auth":{"id":"auth_bearer"}}}}`),
		}); err == nil || !strings.Contains(err.Error(), "is required") {
			t.Fatalf("expected required error, got %v", err)
		}
	})
	t.Run("valid config loads", func(t *testing.T) {
		cfg, err := LoadConfig(t.Context(), backend.DataSourceInstanceSettings{
			JSONData:                []byte(`{"services":{"logicmonitor":{"auth":{"id":"auth_bearer"}}},"variables":{"account_name":"foo"}}`),
			DecryptedSecureJSONData: map[string]string{"logicmonitor.token": "tok"},
		})
		if err != nil {
			t.Fatalf("LoadConfig: %v", err)
		}
		if cfg.Variables.AccountName != "foo" {
			t.Errorf("account_name = %q", cfg.Variables.AccountName)
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
		Services:                ServicesConfig{LogicMonitor: ServiceConfig{Auth: AuthConfig{Id: AuthMethodAuthBearer}}},
		Variables:               VariablesConfig{AccountName: "foo"},
		DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyToken: "tok"},
	}
	if err := ok.Validate(); err != nil {
		t.Fatalf("Validate ok: %v", err)
	}
	noAcct := ok
	noAcct.Variables.AccountName = ""
	if err := noAcct.Validate(); err == nil || !strings.Contains(err.Error(), "account_name is required") {
		t.Fatalf("Validate no account_name: %v", err)
	}
}
