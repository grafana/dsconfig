package atlassianstatuspagedatasource

import (
	"strings"
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

func TestLoadConfig(t *testing.T) {
	t.Run("default example fails validation (empty url)", func(t *testing.T) {
		if _, err := LoadConfig(t.Context(), backend.DataSourceInstanceSettings{JSONData: []byte(`{"variables":{"url":""}}`)}); err == nil || !strings.Contains(err.Error(), "url is required") {
			t.Fatalf("expected url required error, got %v", err)
		}
	})
	t.Run("configured loads", func(t *testing.T) {
		cfg, err := LoadConfig(t.Context(), backend.DataSourceInstanceSettings{JSONData: []byte(`{"variables":{"url":"https://www.githubstatus.com"}}`)})
		if err != nil {
			t.Fatalf("LoadConfig: %v", err)
		}
		if cfg.Variables.URL != "https://www.githubstatus.com" {
			t.Errorf("url = %q", cfg.Variables.URL)
		}
	})
	t.Run("invalid jsonData errors", func(t *testing.T) {
		if _, err := LoadConfig(t.Context(), backend.DataSourceInstanceSettings{JSONData: []byte(`{`)}); err == nil || !strings.Contains(err.Error(), "parse jsonData") {
			t.Fatalf("expected parse error, got %v", err)
		}
	})
}
