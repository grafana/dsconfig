package adobeanalyticsdatasource

import (
	"strings"
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

func TestLoadConfig(t *testing.T) {
	t.Run("default example fails validation (empty fields)", func(t *testing.T) {
		if _, err := LoadConfig(t.Context(), backend.DataSourceInstanceSettings{
			JSONData: []byte(`{"services":{"adobe_analytics":{"auth":{"id":"oauth2_m2m"}}}}`),
		}); err == nil || !strings.Contains(err.Error(), "is required") {
			t.Fatalf("expected required error, got %v", err)
		}
	})
	t.Run("valid config loads", func(t *testing.T) {
		cfg, err := LoadConfig(t.Context(), backend.DataSourceInstanceSettings{
			JSONData:                []byte(`{"services":{"adobe_analytics":{"auth":{"id":"oauth2_m2m","clientId":"cid"}}},"variables":{"global_company_id":"gcid"}}`),
			DecryptedSecureJSONData: map[string]string{"adobe_analytics.clientSecret": "secret"},
		})
		if err != nil {
			t.Fatalf("LoadConfig: %v", err)
		}
		if cfg.Services.AdobeAnalytics.Auth.ClientId != "cid" {
			t.Errorf("clientId = %q", cfg.Services.AdobeAnalytics.Auth.ClientId)
		}
		if cfg.Variables.GlobalCompanyID != "gcid" {
			t.Errorf("global_company_id = %q", cfg.Variables.GlobalCompanyID)
		}
		if cfg.DecryptedSecureJSONData[SecureJsonDataKeyClientSecret] != "secret" {
			t.Errorf("clientSecret = %q", cfg.DecryptedSecureJSONData[SecureJsonDataKeyClientSecret])
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
		Services:                ServicesConfig{AdobeAnalytics: ServiceConfig{Auth: AuthConfig{Id: AuthMethodOAuth2M2M, ClientId: "cid"}}},
		Variables:               VariablesConfig{GlobalCompanyID: "gcid"},
		DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyClientSecret: "secret"},
	}
	if err := ok.Validate(); err != nil {
		t.Fatalf("Validate ok: %v", err)
	}
	noCID := ok
	noCID.Services.AdobeAnalytics.Auth.ClientId = ""
	if err := noCID.Validate(); err == nil || !strings.Contains(err.Error(), "clientId is required") {
		t.Fatalf("Validate no clientId: %v", err)
	}
	noGCID := ok
	noGCID.Variables.GlobalCompanyID = ""
	if err := noGCID.Validate(); err == nil || !strings.Contains(err.Error(), "global_company_id is required") {
		t.Fatalf("Validate no global_company_id: %v", err)
	}
}
