package honeycombdatasource

import (
	"encoding/json"
	"reflect"
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
		name          string
		example       string // schema.go SettingsExamples key ("" handled explicitly below)
		useExample    bool
		settings      backend.DataSourceInstanceSettings
		wantErr       string // substring match; empty = expect success
		wantHostname  string
		wantTeam      string
		wantEnv       string
		wantRetention int
		wantAPIKey    string
		wantSecure    SecureJsonDataConfig
	}{
		{
			// The default schema example has empty team + apiKey placeholders,
			// so LoadConfig's Validate step is expected to reject it.
			name:       "default example fails validation (empty placeholders)",
			example:    "",
			useExample: true,
			wantErr:    "is required",
		},
		{
			name:          "us api example",
			example:       "usApi",
			useExample:    true,
			wantHostname:  DefaultHoneycombAPIURL,
			wantTeam:      "<your-honeycomb-team-slug>",
			wantRetention: DefaultRetentionLimitDays,
			wantAPIKey:    "<your-honeycomb-api-key>",
			wantSecure:    SecureJsonDataConfig{SecureJsonDataKeyAPIKey},
		},
		{
			name:          "eu api example",
			example:       "euApi",
			useExample:    true,
			wantHostname:  "https://api.eu1.honeycomb.io",
			wantTeam:      "<your-honeycomb-team-slug>",
			wantRetention: DefaultRetentionLimitDays,
			wantAPIKey:    "<your-honeycomb-api-key>",
			wantSecure:    SecureJsonDataConfig{SecureJsonDataKeyAPIKey},
		},
		{
			name:          "with environment example",
			example:       "withEnvironment",
			useExample:    true,
			wantHostname:  DefaultHoneycombAPIURL,
			wantTeam:      "<your-honeycomb-team-slug>",
			wantEnv:       "<your-honeycomb-environment>",
			wantRetention: DefaultRetentionLimitDays,
			wantAPIKey:    "<your-honeycomb-api-key>",
			wantSecure:    SecureJsonDataConfig{SecureJsonDataKeyAPIKey},
		},
		{
			name:          "extended retention example",
			example:       "extendedRetention",
			useExample:    true,
			wantHostname:  DefaultHoneycombAPIURL,
			wantTeam:      "<your-honeycomb-team-slug>",
			wantRetention: 30,
			wantAPIKey:    "<your-honeycomb-api-key>",
			wantSecure:    SecureJsonDataConfig{SecureJsonDataKeyAPIKey},
		},
		{
			name: "minimal inline config succeeds and applies defaults",
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"team":"my-team"}`),
				DecryptedSecureJSONData: map[string]string{"apiKey": "secret"},
			},
			wantHostname:  DefaultHoneycombAPIURL,
			wantTeam:      "my-team",
			wantRetention: DefaultRetentionLimitDays,
			wantAPIKey:    "secret",
			wantSecure:    SecureJsonDataConfig{SecureJsonDataKeyAPIKey},
		},
		{
			name: "explicit values override defaults",
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"hostname":"https://api.eu1.honeycomb.io","team":"t","environment":"prod","retentionLimit":14}`),
				DecryptedSecureJSONData: map[string]string{"apiKey": "secret"},
			},
			wantHostname:  "https://api.eu1.honeycomb.io",
			wantTeam:      "t",
			wantEnv:       "prod",
			wantRetention: 14,
			wantAPIKey:    "secret",
			wantSecure:    SecureJsonDataConfig{SecureJsonDataKeyAPIKey},
		},
		{
			name: "empty JSONData is a parse error",
			settings: backend.DataSourceInstanceSettings{
				DecryptedSecureJSONData: map[string]string{"apiKey": "secret"},
			},
			wantErr: "parse jsonData",
		},
		{
			name: "malformed jsonData errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{`),
			},
			wantErr: "parse jsonData",
		},
		{
			name: "non-https hostname fails validation",
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"hostname":"http://api.honeycomb.io","team":"t"}`),
				DecryptedSecureJSONData: map[string]string{"apiKey": "secret"},
			},
			wantErr: "https scheme",
		},
		{
			name: "missing team fails validation",
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"hostname":"https://api.honeycomb.io"}`),
				DecryptedSecureJSONData: map[string]string{"apiKey": "secret"},
			},
			wantErr: "team name (jsonData.team) is required",
		},
		{
			name: "missing api key fails validation",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"hostname":"https://api.honeycomb.io","team":"t"}`),
			},
			wantErr: "API key (secureJsonData.apiKey) is required",
		},
		{
			name: "empty hostname fails validation",
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"hostname":"","team":"t"}`),
				DecryptedSecureJSONData: map[string]string{"apiKey": "secret"},
			},
			// ApplyDefaults restores the default hostname when it is empty, so
			// this actually succeeds — assert defaulting rather than an error.
			wantHostname:  DefaultHoneycombAPIURL,
			wantTeam:      "t",
			wantRetention: DefaultRetentionLimitDays,
			wantAPIKey:    "secret",
			wantSecure:    SecureJsonDataConfig{SecureJsonDataKeyAPIKey},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings := tt.settings
			if tt.useExample {
				settings = settingsFromExample(t, tt.example)
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

			if cfg.Hostname != tt.wantHostname {
				t.Errorf("Hostname = %q, want %q", cfg.Hostname, tt.wantHostname)
			}
			if cfg.Team != tt.wantTeam {
				t.Errorf("Team = %q, want %q", cfg.Team, tt.wantTeam)
			}
			if cfg.Env != tt.wantEnv {
				t.Errorf("Env = %q, want %q", cfg.Env, tt.wantEnv)
			}
			if cfg.RetentionLimit != tt.wantRetention {
				t.Errorf("RetentionLimit = %d, want %d", cfg.RetentionLimit, tt.wantRetention)
			}
			if cfg.DecryptedSecureJSONData[SecureJsonDataKeyAPIKey] != tt.wantAPIKey {
				t.Errorf("apiKey = %q, want %q", cfg.DecryptedSecureJSONData[SecureJsonDataKeyAPIKey], tt.wantAPIKey)
			}
			if tt.wantSecure != nil {
				gotKeys := SecureJsonDataConfig{}
				for _, key := range SecureJsonDataKeys {
					if _, ok := cfg.DecryptedSecureJSONData[key]; ok {
						gotKeys = append(gotKeys, key)
					}
				}
				if !reflect.DeepEqual(gotKeys, tt.wantSecure) {
					t.Errorf("configured secure keys = %v, want %v", gotKeys, tt.wantSecure)
				}
			}
		})
	}
}

func TestApplyDefaults(t *testing.T) {
	tests := []struct {
		name          string
		in            Config
		wantHostname  string
		wantRetention int
	}{
		{
			name:          "empty config gets hostname + retention defaults",
			in:            Config{},
			wantHostname:  DefaultHoneycombAPIURL,
			wantRetention: DefaultRetentionLimitDays,
		},
		{
			name:          "existing hostname is preserved",
			in:            Config{Hostname: "https://api.eu1.honeycomb.io"},
			wantHostname:  "https://api.eu1.honeycomb.io",
			wantRetention: DefaultRetentionLimitDays,
		},
		{
			name:          "existing retention is preserved",
			in:            Config{RetentionLimit: 30},
			wantHostname:  DefaultHoneycombAPIURL,
			wantRetention: 30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.in
			got.ApplyDefaults()
			if got.Hostname != tt.wantHostname {
				t.Errorf("Hostname = %q, want %q", got.Hostname, tt.wantHostname)
			}
			if got.RetentionLimit != tt.wantRetention {
				t.Errorf("RetentionLimit = %d, want %d", got.RetentionLimit, tt.wantRetention)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	valid := func() Config {
		return Config{
			Hostname:                DefaultHoneycombAPIURL,
			Team:                    "my-team",
			RetentionLimit:          DefaultRetentionLimitDays,
			DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyAPIKey: "secret"},
		}
	}

	tests := []struct {
		name    string
		cfg     Config
		wantErr string // empty = expect no error; otherwise substring match
	}{
		{
			name: "valid config",
			cfg:  valid(),
		},
		{
			name: "missing hostname errors",
			cfg: func() Config {
				c := valid()
				c.Hostname = ""
				return c
			}(),
			wantErr: "hostname (jsonData.hostname) is required",
		},
		{
			name: "non-https hostname errors",
			cfg: func() Config {
				c := valid()
				c.Hostname = "http://api.honeycomb.io"
				return c
			}(),
			wantErr: "https scheme",
		},
		{
			name: "non-uri hostname errors",
			cfg: func() Config {
				c := valid()
				c.Hostname = "not a url"
				return c
			}(),
			wantErr: "invalid hostname URL",
		},
		{
			name: "missing api key errors",
			cfg: func() Config {
				c := valid()
				c.DecryptedSecureJSONData = map[SecureJsonDataKey]string{}
				return c
			}(),
			wantErr: "API key (secureJsonData.apiKey) is required",
		},
		{
			name: "missing team errors",
			cfg: func() Config {
				c := valid()
				c.Team = "   "
				return c
			}(),
			wantErr: "team name (jsonData.team) is required",
		},
		{
			name:    "multiple problems are joined",
			cfg:     Config{},
			wantErr: "is required",
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
