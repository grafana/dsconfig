package hellodatasource

import (
	"strings"
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

func TestLoadConfig(t *testing.T) {
	t.Run("empty settings load and default auth ids to none", func(t *testing.T) {
		cfg, err := LoadConfig(t.Context(), backend.DataSourceInstanceSettings{})
		if err != nil {
			t.Fatalf("LoadConfig: %v", err)
		}
		if cfg.Services.HTTPBin.Auth.Id != AuthMethodNone || cfg.Services.PostmanEcho.Auth.Id != AuthMethodNone {
			t.Errorf("auth ids = %q/%q, want none/none", cfg.Services.HTTPBin.Auth.Id, cfg.Services.PostmanEcho.Auth.Id)
		}
	})
	t.Run("invalid jsonData errors", func(t *testing.T) {
		if _, err := LoadConfig(t.Context(), backend.DataSourceInstanceSettings{JSONData: []byte(`{`)}); err == nil || !strings.Contains(err.Error(), "parse jsonData") {
			t.Fatalf("expected parse error, got %v", err)
		}
	})
	t.Run("unknown auth method errors", func(t *testing.T) {
		if _, err := LoadConfig(t.Context(), backend.DataSourceInstanceSettings{
			JSONData: []byte(`{"services":{"httpbin":{"auth":{"id":"bogus"}}}}`),
		}); err == nil || !strings.Contains(err.Error(), "unknown auth method") {
			t.Fatalf("expected unknown auth method error, got %v", err)
		}
	})
}
