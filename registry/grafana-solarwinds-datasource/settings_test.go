package solarwindsdatasource

import (
	"strings"
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

func TestLoadConfig(t *testing.T) {
	t.Run("default example fails validation (empty fields)", func(t *testing.T) {
		if _, err := LoadConfig(t.Context(), backend.DataSourceInstanceSettings{
			JSONData: []byte(`{"services":{"solarwinds":{"auth":{"id":"basic_auth"}}}}`),
		}); err == nil || !strings.Contains(err.Error(), "is required") {
			t.Fatalf("expected required error, got %v", err)
		}
	})
	t.Run("basic auth loads", func(t *testing.T) {
		cfg, err := LoadConfig(t.Context(), backend.DataSourceInstanceSettings{
			JSONData:                []byte(`{"services":{"solarwinds":{"auth":{"id":"basic_auth","username":"admin"}}},"variables":{"url":"https://sw.example.com"}}`),
			DecryptedSecureJSONData: map[string]string{"solarwinds.password": "pw"},
		})
		if err != nil {
			t.Fatalf("LoadConfig: %v", err)
		}
		if cfg.Services.Solarwinds.Auth.UserName != "admin" || cfg.Variables.URL != "https://sw.example.com" {
			t.Errorf("unexpected cfg %+v", cfg)
		}
	})
	t.Run("mutual TLS requires cert and key", func(t *testing.T) {
		_, err := LoadConfig(t.Context(), backend.DataSourceInstanceSettings{
			JSONData:                []byte(`{"services":{"solarwinds":{"auth":{"id":"basic_auth","username":"admin","tls":{"clientAuth":{"enabled":true}}}}},"variables":{"url":"https://sw.example.com"}}`),
			DecryptedSecureJSONData: map[string]string{"solarwinds.password": "pw"},
		})
		if err == nil || !strings.Contains(err.Error(), "client certificate") {
			t.Fatalf("expected client cert required error, got %v", err)
		}
	})
	t.Run("invalid jsonData errors", func(t *testing.T) {
		if _, err := LoadConfig(t.Context(), backend.DataSourceInstanceSettings{JSONData: []byte(`{`)}); err == nil || !strings.Contains(err.Error(), "parse jsonData") {
			t.Fatalf("expected parse error, got %v", err)
		}
	})
}

func TestValidate(t *testing.T) {
	base := Config{
		Services:                ServicesConfig{Solarwinds: ServiceConfig{Auth: AuthConfig{Id: AuthMethodBasic, UserName: "admin"}}},
		Variables:               VariablesConfig{URL: "https://sw.example.com"},
		DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyPassword: "pw"},
	}
	if err := base.Validate(); err != nil {
		t.Fatalf("Validate ok: %v", err)
	}
	selfSigned := base
	selfSigned.Services.Solarwinds.Auth.TLS.SelfSignedCert.Enabled = true
	if err := selfSigned.Validate(); err == nil || !strings.Contains(err.Error(), "self-signed certificate") {
		t.Fatalf("Validate self-signed enabled without cert: %v", err)
	}
	noURL := base
	noURL.Variables.URL = ""
	if err := noURL.Validate(); err == nil || !strings.Contains(err.Error(), "url is required") {
		t.Fatalf("Validate no url: %v", err)
	}
}
