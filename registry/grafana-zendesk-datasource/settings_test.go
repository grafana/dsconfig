package zendeskdatasource

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// settingsFromExample converts a SettingsExamples entry (a full instance
// settings object with jsonData and secureJsonData) into the
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
	return backend.DataSourceInstanceSettings{
		JSONData:                jsonData,
		DecryptedSecureJSONData: secure,
	}
}

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name         string
		example      string // schema.go SettingsExamples key ("" excluded here; use inline settings)
		settings     backend.DataSourceInstanceSettings
		wantErr      string // empty = expect success; otherwise substring match
		wantAuth     AuthMethodID
		wantUsername string
		wantSubdomn  string
		wantToken    string
	}{
		{
			// The default schema example intentionally has empty username /
			// subdomain / token placeholders, so Validate is expected to reject it.
			name:    "default example fails validation (empty placeholders)",
			example: "",
			wantErr: "is required",
		},
		{
			name:         "basic auth example loads",
			example:      "basicAuth",
			wantAuth:     AuthMethodBasic,
			wantUsername: "agent@example.com",
			wantSubdomn:  "mycompany",
			wantToken:    "<zendesk-api-token>",
		},
		{
			name:     "empty settings default to basic auth and fail validation",
			settings: backend.DataSourceInstanceSettings{},
			wantErr:  "is required",
		},
		{
			name: "invalid jsonData errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{`),
			},
			wantErr: "parse jsonData",
		},
		{
			name: "explicit config loads",
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"services":{"zendesk":{"auth":{"id":"basic_auth","username":"me@example.com"}}},"variables":{"subdomain":"acme"}}`),
				DecryptedSecureJSONData: map[string]string{"zendesk.password": "tok"},
			},
			wantAuth:     AuthMethodBasic,
			wantUsername: "me@example.com",
			wantSubdomn:  "acme",
			wantToken:    "tok",
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
				if err == nil {
					t.Fatalf("LoadConfig: expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("LoadConfig: error = %q, want it to contain %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("LoadConfig: %v", err)
			}
			if tt.wantAuth != "" && cfg.Services.Zendesk.Auth.Id != tt.wantAuth {
				t.Errorf("auth.id = %q, want %q", cfg.Services.Zendesk.Auth.Id, tt.wantAuth)
			}
			if tt.wantUsername != "" && cfg.Services.Zendesk.Auth.UserName != tt.wantUsername {
				t.Errorf("auth.username = %q, want %q", cfg.Services.Zendesk.Auth.UserName, tt.wantUsername)
			}
			if tt.wantSubdomn != "" && cfg.Variables.Subdomain != tt.wantSubdomn {
				t.Errorf("variables.subdomain = %q, want %q", cfg.Variables.Subdomain, tt.wantSubdomn)
			}
			if tt.wantToken != "" && cfg.DecryptedSecureJSONData[SecureJsonDataKeyPassword] != tt.wantToken {
				t.Errorf("secret[zendesk.password] = %q, want %q", cfg.DecryptedSecureJSONData[SecureJsonDataKeyPassword], tt.wantToken)
			}
		})
	}
}

func TestApplyDefaults(t *testing.T) {
	t.Run("empty auth id defaults to basic_auth", func(t *testing.T) {
		c := Config{}
		c.ApplyDefaults()
		if c.Services.Zendesk.Auth.Id != AuthMethodBasic {
			t.Errorf("auth.id = %q, want %q", c.Services.Zendesk.Auth.Id, AuthMethodBasic)
		}
	})
	t.Run("existing auth id preserved", func(t *testing.T) {
		c := Config{Services: ServicesConfig{Zendesk: ServiceConfig{Auth: AuthConfig{Id: "other"}}}}
		c.ApplyDefaults()
		if c.Services.Zendesk.Auth.Id != "other" {
			t.Errorf("auth.id = %q, want %q", c.Services.Zendesk.Auth.Id, "other")
		}
	})
}

func TestValidate(t *testing.T) {
	base := func() Config {
		return Config{
			Services:                ServicesConfig{Zendesk: ServiceConfig{Auth: AuthConfig{Id: AuthMethodBasic, UserName: "me@example.com"}}},
			Variables:               VariablesConfig{Subdomain: "acme"},
			DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyPassword: "tok"},
		}
	}

	tests := []struct {
		name    string
		mutate  func(*Config)
		wantErr string
	}{
		{name: "happy path", mutate: func(*Config) {}},
		{
			name:    "missing username",
			mutate:  func(c *Config) { c.Services.Zendesk.Auth.UserName = "" },
			wantErr: "username (email) is required",
		},
		{
			name:    "missing token",
			mutate:  func(c *Config) { delete(c.DecryptedSecureJSONData, SecureJsonDataKeyPassword) },
			wantErr: `secureJsonData "zendesk.password"`,
		},
		{
			name:    "missing subdomain",
			mutate:  func(c *Config) { c.Variables.Subdomain = "" },
			wantErr: "subdomain is required",
		},
		{
			name:    "empty auth method",
			mutate:  func(c *Config) { c.Services.Zendesk.Auth.Id = "" },
			wantErr: "auth method is required",
		},
		{
			name:    "unknown auth method",
			mutate:  func(c *Config) { c.Services.Zendesk.Auth.Id = "bogus" },
			wantErr: `unknown auth method "bogus"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := base()
			tt.mutate(&c)
			err := c.Validate()
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
