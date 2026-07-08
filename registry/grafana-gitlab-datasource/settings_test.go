package gitlabdatasource

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// settingsFromExample converts a SettingsExamples entry (a full instance
// settings object with a root url, jsonData, and secureJsonData) into the
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
	settings := backend.DataSourceInstanceSettings{
		JSONData:                jsonData,
		DecryptedSecureJSONData: secure,
	}
	if s, ok := value["url"].(string); ok {
		settings.URL = s
	}
	return settings
}

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name            string
		useExample      bool
		example         string // schema.go SettingsExamples key
		settings        backend.DataSourceInstanceSettings
		wantErr         error
		wantURL         string
		wantPageLimit   int
		wantSecureKeys  SecureJsonDataConfig
		wantAccessToken string
	}{
		{
			// The default example intentionally has an empty accessToken
			// placeholder, so LoadConfig's Validate step is expected to reject it.
			name:       "default example fails validation (empty accessToken placeholder)",
			useExample: true,
			example:    "",
			wantErr:    errors.New("access token (secureJsonData.accessToken) can not be blank"),
		},
		{
			name:            "gitlab saas",
			useExample:      true,
			example:         "gitlabSaaS",
			wantURL:         DefaultURL,
			wantPageLimit:   DefaultPageLimit,
			wantSecureKeys:  SecureJsonDataConfig{SecureJsonDataKeyAccessToken},
			wantAccessToken: "<gitlab-personal-access-token>",
		},
		{
			name:            "self-hosted",
			useExample:      true,
			example:         "selfHosted",
			wantURL:         "https://gitlab.example.com",
			wantPageLimit:   DefaultPageLimit,
			wantSecureKeys:  SecureJsonDataConfig{SecureJsonDataKeyAccessToken},
			wantAccessToken: "<gitlab-personal-access-token>",
		},
		{
			name:            "self-hosted with explicit api/v4 and custom page limit",
			useExample:      true,
			example:         "selfHostedApiV4CustomPageLimit",
			wantURL:         "https://gitlab.example.com/api/v4",
			wantPageLimit:   10,
			wantSecureKeys:  SecureJsonDataConfig{SecureJsonDataKeyAccessToken},
			wantAccessToken: "<gitlab-personal-access-token>",
		},
		{
			name: "empty root url defaults to gitlab.com api base",
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{}`),
				DecryptedSecureJSONData: map[string]string{"accessToken": "tok"},
			},
			wantURL:       DefaultURL,
			wantPageLimit: DefaultPageLimit,
		},
		{
			name: "explicit url and pageLimit are preserved",
			settings: backend.DataSourceInstanceSettings{
				URL:                     "https://gitlab.example.com/api/v4",
				JSONData:                []byte(`{"pageLimit":25}`),
				DecryptedSecureJSONData: map[string]string{"accessToken": "tok"},
			},
			wantURL:       "https://gitlab.example.com/api/v4",
			wantPageLimit: 25,
		},
		{
			name: "pageLimit 0 defaults to 5",
			settings: backend.DataSourceInstanceSettings{
				URL:                     "https://gitlab.example.com",
				JSONData:                []byte(`{"pageLimit":0}`),
				DecryptedSecureJSONData: map[string]string{"accessToken": "tok"},
			},
			wantURL:       "https://gitlab.example.com",
			wantPageLimit: DefaultPageLimit,
		},
		{
			name: "invalid jsonData errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{`),
			},
			wantErr: errors.New("parse jsonData"),
		},
		{
			// Upstream LoadSettings unmarshals config.JSONData unconditionally
			// (pkg/models/settings.go:30), so a truly-empty payload is a parse
			// error. Grafana always sends at least "{}".
			name:     "empty JSONData is a parse error (mirrors upstream)",
			settings: backend.DataSourceInstanceSettings{},
			wantErr:  errors.New("parse jsonData"),
		},
		{
			name: "missing access token errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://gitlab.example.com",
				JSONData: []byte(`{}`),
			},
			wantErr: errors.New("access token (secureJsonData.accessToken) can not be blank"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings := tt.settings
			if tt.useExample {
				settings = settingsFromExample(t, tt.example)
			}

			cfg, err := LoadConfig(t.Context(), settings)
			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("LoadConfig: expected error %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr.Error()) {
					t.Fatalf("LoadConfig: error = %q, want it to contain %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("LoadConfig: %v", err)
			}

			if tt.wantURL != "" && cfg.URL != tt.wantURL {
				t.Errorf("URL = %q, want %q", cfg.URL, tt.wantURL)
			}
			if tt.wantPageLimit != 0 && cfg.PageLimit != tt.wantPageLimit {
				t.Errorf("PageLimit = %d, want %d", cfg.PageLimit, tt.wantPageLimit)
			}
			if tt.wantSecureKeys != nil {
				gotKeys := SecureJsonDataConfig{}
				for _, key := range SecureJsonDataKeys {
					if _, ok := cfg.DecryptedSecureJSONData[key]; ok {
						gotKeys = append(gotKeys, key)
					}
				}
				if !reflect.DeepEqual(gotKeys, tt.wantSecureKeys) {
					t.Errorf("configured secure keys = %v, want %v", gotKeys, tt.wantSecureKeys)
				}
			}
			if tt.wantAccessToken != "" && cfg.DecryptedSecureJSONData[SecureJsonDataKeyAccessToken] != tt.wantAccessToken {
				t.Errorf("DecryptedSecureJSONData[accessToken] = %q, want %q", cfg.DecryptedSecureJSONData[SecureJsonDataKeyAccessToken], tt.wantAccessToken)
			}
		})
	}
}

func TestApplyDefaults(t *testing.T) {
	tests := []struct {
		name          string
		in            Config
		wantURL       string
		wantPageLimit int
	}{
		{
			name:          "empty config gets default url + page limit",
			in:            Config{},
			wantURL:       DefaultURL,
			wantPageLimit: DefaultPageLimit,
		},
		{
			name:          "existing url is preserved",
			in:            Config{URL: "https://gitlab.example.com/api/v4"},
			wantURL:       "https://gitlab.example.com/api/v4",
			wantPageLimit: DefaultPageLimit,
		},
		{
			name:          "existing page limit is preserved",
			in:            Config{PageLimit: 42},
			wantURL:       DefaultURL,
			wantPageLimit: 42,
		},
		{
			name:          "zero page limit defaults to 5",
			in:            Config{URL: "https://gitlab.example.com", PageLimit: 0},
			wantURL:       "https://gitlab.example.com",
			wantPageLimit: DefaultPageLimit,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.in
			got.ApplyDefaults()
			if got.URL != tt.wantURL {
				t.Errorf("URL = %q, want %q", got.URL, tt.wantURL)
			}
			if got.PageLimit != tt.wantPageLimit {
				t.Errorf("PageLimit = %d, want %d", got.PageLimit, tt.wantPageLimit)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr string // empty = expect no error; otherwise substring match
	}{
		{
			name: "access token present",
			cfg: Config{
				URL:                     DefaultURL,
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyAccessToken: "tok"},
			},
		},
		{
			name:    "empty access token errors",
			cfg:     Config{URL: DefaultURL, DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyAccessToken: ""}},
			wantErr: "access token (secureJsonData.accessToken) can not be blank",
		},
		{
			name:    "missing access token errors",
			cfg:     Config{URL: DefaultURL},
			wantErr: "access token (secureJsonData.accessToken) can not be blank",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("Validate: unexpected error %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("Validate: expected error containing %q, got nil", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Validate: error = %q, want it to contain %q", err, tt.wantErr)
			}
		})
	}
}

// TestSettingsExamples guards the AGENTS.md example requirements: the default
// example keyed by "" exists, and every example carries a jsonData object and a
// non-empty secureJsonData object whose keys are all known secret keys.
func TestSettingsExamples(t *testing.T) {
	examples := SettingsExamples().Examples
	if _, ok := examples[""]; !ok {
		t.Fatalf(`missing default example keyed by ""`)
	}

	known := map[string]bool{}
	for _, k := range SecureJsonDataKeys {
		known[string(k)] = true
	}

	for key, ex := range examples {
		value, ok := ex.Value.(map[string]any)
		if !ok {
			t.Fatalf("example %q value is not an object", key)
		}
		if _, ok := value["jsonData"].(map[string]any); !ok {
			t.Errorf("example %q missing jsonData object", key)
		}
		secure, ok := value["secureJsonData"].(map[string]any)
		if !ok || len(secure) == 0 {
			t.Errorf("example %q missing non-empty secureJsonData object", key)
			continue
		}
		for secretKey := range secure {
			if !known[secretKey] {
				t.Errorf("example %q references unknown secret key %q", key, secretKey)
			}
		}
	}
}
