package falconlogscaledatasource

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// settingsFromExample converts a SettingsExamples entry (a full instance
// settings object with root fields, jsonData, and secureJsonData) into the
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
	if b, ok := value["basicAuth"].(bool); ok {
		settings.BasicAuthEnabled = b
	}
	if s, ok := value["basicAuthUser"].(string); ok {
		settings.BasicAuthUser = s
	}
	return settings
}

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name            string
		example         string
		settings        backend.DataSourceInstanceSettings
		wantErr         error
		wantURL         string
		wantMode        DataSourceMode
		wantToken       bool
		wantOAuth2      bool
		wantOAuthPass   bool
		wantBasicAuth   bool
		wantClientID    string
		wantDefaultRepo string
		wantDataLinkCnt int
		wantIncremental bool
		wantOverlap     string
		wantSecureKeys  SecureJsonDataConfig
	}{
		{
			// The default example uses token auth but leaves the access token
			// empty as a placeholder, so validation fails.
			name:    "default example fails validation (empty accessToken placeholder)",
			example: "",
			wantErr: errors.New("accessToken (secureJsonData) is required"),
		},
		{
			name:            "logscale token auth",
			example:         "logscaleToken",
			wantURL:         "https://cloud.humio.com",
			wantMode:        DataSourceModeLogScale,
			wantToken:       true,
			wantDefaultRepo: "example-repo",
			wantSecureKeys:  SecureJsonDataConfig{SecureJsonDataKeyAccessToken},
		},
		{
			name:           "logscale oauth2 client credentials",
			example:        "logscaleOAuth2Client",
			wantURL:        "https://cloud.humio.com",
			wantMode:       DataSourceModeLogScale,
			wantOAuth2:     true,
			wantClientID:   "my-client-id",
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyOAuth2ClientSecret},
		},
		{
			name:           "logscale basic auth",
			example:        "logscaleBasicAuth",
			wantURL:        "https://cloud.humio.com",
			wantMode:       DataSourceModeLogScale,
			wantBasicAuth:  true,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyBasicAuthPassword},
		},
		{
			name:          "logscale oauth forward",
			example:       "logscaleOAuthForward",
			wantURL:       "https://cloud.humio.com",
			wantMode:      DataSourceModeLogScale,
			wantOAuthPass: true,
		},
		{
			name:            "ngsiem oauth2",
			example:         "ngsiemOAuth2",
			wantURL:         "https://api.us-2.crowdstrike.com",
			wantMode:        DataSourceModeNGSIEM,
			wantOAuth2:      true,
			wantClientID:    "my-ngsiem-client-id",
			wantDefaultRepo: "search-all",
			wantSecureKeys:  SecureJsonDataConfig{SecureJsonDataKeyOAuth2ClientSecret},
		},
		{
			name:            "logscale token auth with data links and incremental querying",
			example:         "logscaleWithDataLinks",
			wantURL:         "https://cloud.humio.com",
			wantMode:        DataSourceModeLogScale,
			wantToken:       true,
			wantDefaultRepo: "production",
			wantDataLinkCnt: 2,
			wantIncremental: true,
			wantOverlap:     "30s",
			wantSecureKeys:  SecureJsonDataConfig{SecureJsonDataKeyAccessToken},
		},
		{
			name: "missing URL errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"mode":"LogScale","authenticateWithToken":true}`),
				DecryptedSecureJSONData: map[string]string{
					"accessToken": "tok",
				},
			},
			wantErr: errors.New("URL (root.url) is required"),
		},
		{
			name: "invalid jsonData errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://cloud.humio.com",
				JSONData: []byte(`{`),
			},
			wantErr: errors.New("parse jsonData"),
		},
		{
			name: "token auth without token errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://cloud.humio.com",
				JSONData: []byte(`{"authenticateWithToken":true}`),
			},
			wantErr: errors.New("accessToken (secureJsonData) is required"),
		},
		{
			name: "oauth2 without client id errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://cloud.humio.com",
				JSONData: []byte(`{"oauth2":true}`),
				DecryptedSecureJSONData: map[string]string{
					"oauth2ClientSecret": "sec",
				},
			},
			wantErr: errors.New("oauth2ClientId (jsonData) is required"),
		},
		{
			name: "oauth2 without client secret errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://cloud.humio.com",
				JSONData: []byte(`{"oauth2":true,"oauth2ClientId":"cid"}`),
			},
			wantErr: errors.New("oauth2ClientSecret (secureJsonData) is required"),
		},
		{
			name: "basic auth without user errors",
			settings: backend.DataSourceInstanceSettings{
				URL:              "https://cloud.humio.com",
				BasicAuthEnabled: true,
				JSONData:         []byte(`{}`),
				DecryptedSecureJSONData: map[string]string{
					"basicAuthPassword": "pw",
				},
			},
			wantErr: errors.New("basicAuthUser (root) is required"),
		},
		{
			name: "basic auth without password errors",
			settings: backend.DataSourceInstanceSettings{
				URL:              "https://cloud.humio.com",
				BasicAuthEnabled: true,
				BasicAuthUser:    "grafana",
				JSONData:         []byte(`{}`),
			},
			wantErr: errors.New("basicAuthPassword (secureJsonData) is required"),
		},
		{
			name: "ngsiem mode without oauth2 errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://api.us-2.crowdstrike.com",
				JSONData: []byte(`{"mode":"NGSIEM","authenticateWithToken":true}`),
				DecryptedSecureJSONData: map[string]string{
					"accessToken": "tok",
				},
			},
			wantErr: errors.New("NGSIEM mode requires OAuth2 client credentials"),
		},
		{
			name: "multiple auth methods enabled errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://cloud.humio.com",
				JSONData: []byte(`{"authenticateWithToken":true,"oauth2":true,"oauth2ClientId":"cid"}`),
				DecryptedSecureJSONData: map[string]string{
					"accessToken":        "tok",
					"oauth2ClientSecret": "sec",
				},
			},
			wantErr: errors.New("only one authentication method may be enabled"),
		},
		{
			name: "unknown mode errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://cloud.humio.com",
				JSONData: []byte(`{"mode":"bogus","authenticateWithToken":true}`),
				DecryptedSecureJSONData: map[string]string{
					"accessToken": "tok",
				},
			},
			wantErr: errors.New(`unknown mode "bogus"`),
		},
		{
			name: "negative timeout errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://cloud.humio.com",
				JSONData: []byte(`{"authenticateWithToken":true,"timeout":-5}`),
				DecryptedSecureJSONData: map[string]string{
					"accessToken": "tok",
				},
			},
			wantErr: errors.New("timeout must be non-negative"),
		},
		{
			name: "data link missing field errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://cloud.humio.com",
				JSONData: []byte(`{"authenticateWithToken":true,"dataLinks":[{"matcherRegex":".*","url":"http://x"}]}`),
				DecryptedSecureJSONData: map[string]string{
					"accessToken": "tok",
				},
			},
			wantErr: errors.New("dataLinks[0].field is required"),
		},
		{
			name: "data link missing matcherRegex errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://cloud.humio.com",
				JSONData: []byte(`{"authenticateWithToken":true,"dataLinks":[{"field":"x","url":"http://x"}]}`),
				DecryptedSecureJSONData: map[string]string{
					"accessToken": "tok",
				},
			},
			wantErr: errors.New("dataLinks[0].matcherRegex is required"),
		},
		{
			name: "empty settings default to LogScale and fail validation",
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://cloud.humio.com",
				JSONData: []byte(`{}`),
			},
			// After ApplyDefaults, Mode is LogScale and no auth method is selected —
			// which is allowed by Validate (all four flags may be false). URL is set,
			// so this should succeed.
			wantMode: DataSourceModeLogScale,
			wantURL:  "https://cloud.humio.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings := tt.settings
			// If the test names an example (including the default "" example)
			// and does not provide inline settings, load from the example.
			if tt.settings.JSONData == nil && tt.settings.URL == "" && !tt.settings.BasicAuthEnabled {
				if _, ok := SettingsExamples().Examples[tt.example]; ok {
					settings = settingsFromExample(t, tt.example)
				}
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
			if tt.wantMode != "" && cfg.Mode != tt.wantMode {
				t.Errorf("Mode = %q, want %q", cfg.Mode, tt.wantMode)
			}
			if cfg.AuthenticateWithToken != tt.wantToken {
				t.Errorf("AuthenticateWithToken = %v, want %v", cfg.AuthenticateWithToken, tt.wantToken)
			}
			if cfg.OAuth2 != tt.wantOAuth2 {
				t.Errorf("OAuth2 = %v, want %v", cfg.OAuth2, tt.wantOAuth2)
			}
			if cfg.OAuthPassThru != tt.wantOAuthPass {
				t.Errorf("OAuthPassThru = %v, want %v", cfg.OAuthPassThru, tt.wantOAuthPass)
			}
			if cfg.BasicAuth != tt.wantBasicAuth {
				t.Errorf("BasicAuth = %v, want %v", cfg.BasicAuth, tt.wantBasicAuth)
			}
			if tt.wantClientID != "" && cfg.OAuth2ClientID != tt.wantClientID {
				t.Errorf("OAuth2ClientID = %q, want %q", cfg.OAuth2ClientID, tt.wantClientID)
			}
			if tt.wantDefaultRepo != "" && cfg.DefaultRepository != tt.wantDefaultRepo {
				t.Errorf("DefaultRepository = %q, want %q", cfg.DefaultRepository, tt.wantDefaultRepo)
			}
			if len(cfg.DataLinks) != tt.wantDataLinkCnt {
				t.Errorf("len(DataLinks) = %d, want %d", len(cfg.DataLinks), tt.wantDataLinkCnt)
			}
			if cfg.IncrementalQuerying != tt.wantIncremental {
				t.Errorf("IncrementalQuerying = %v, want %v", cfg.IncrementalQuerying, tt.wantIncremental)
			}
			if tt.wantOverlap != "" && cfg.IncrementalQueryOverlapWindow != tt.wantOverlap {
				t.Errorf("IncrementalQueryOverlapWindow = %q, want %q", cfg.IncrementalQueryOverlapWindow, tt.wantOverlap)
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
		})
	}
}

func TestApplyDefaults(t *testing.T) {
	tests := []struct {
		name string
		in   Config
		want Config
	}{
		{
			name: "empty config gets LogScale mode",
			in:   Config{},
			want: Config{Mode: DataSourceModeLogScale},
		},
		{
			name: "existing mode is preserved",
			in:   Config{Mode: DataSourceModeNGSIEM},
			want: Config{Mode: DataSourceModeNGSIEM, DefaultRepository: "search-all"},
		},
		{
			name: "NGSIEM without defaultRepository is auto-set to search-all",
			in:   Config{Mode: DataSourceModeNGSIEM},
			want: Config{Mode: DataSourceModeNGSIEM, DefaultRepository: "search-all"},
		},
		{
			name: "NGSIEM with custom defaultRepository preserved",
			in:   Config{Mode: DataSourceModeNGSIEM, DefaultRepository: "investigate_view"},
			want: Config{Mode: DataSourceModeNGSIEM, DefaultRepository: "investigate_view"},
		},
		{
			name: "LogScale defaultRepository untouched",
			in:   Config{Mode: DataSourceModeLogScale},
			want: Config{Mode: DataSourceModeLogScale},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.in
			got.ApplyDefaults()
			if got.Mode != tt.want.Mode {
				t.Errorf("Mode = %q, want %q", got.Mode, tt.want.Mode)
			}
			if got.DefaultRepository != tt.want.DefaultRepository {
				t.Errorf("DefaultRepository = %q, want %q", got.DefaultRepository, tt.want.DefaultRepository)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr string
	}{
		{
			name: "minimal happy path (no auth selected)",
			cfg: Config{
				URL:  "https://cloud.humio.com",
				Mode: DataSourceModeLogScale,
			},
		},
		{
			name:    "missing URL",
			cfg:     Config{},
			wantErr: "URL (root.url) is required",
		},
		{
			name: "token auth happy path",
			cfg: Config{
				URL:                   "https://cloud.humio.com",
				Mode:                  DataSourceModeLogScale,
				AuthenticateWithToken: true,
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{
					SecureJsonDataKeyAccessToken: "tok",
				},
			},
		},
		{
			name: "oauth2 happy path",
			cfg: Config{
				URL:            "https://cloud.humio.com",
				Mode:           DataSourceModeLogScale,
				OAuth2:         true,
				OAuth2ClientID: "cid",
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{
					SecureJsonDataKeyOAuth2ClientSecret: "sec",
				},
			},
		},
		{
			name: "basic auth happy path",
			cfg: Config{
				URL:           "https://cloud.humio.com",
				Mode:          DataSourceModeLogScale,
				BasicAuth:     true,
				BasicAuthUser: "grafana",
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{
					SecureJsonDataKeyBasicAuthPassword: "pw",
				},
			},
		},
		{
			name: "oauth forward happy path",
			cfg: Config{
				URL:           "https://cloud.humio.com",
				Mode:          DataSourceModeLogScale,
				OAuthPassThru: true,
			},
		},
		{
			name: "ngsiem oauth2 happy path",
			cfg: Config{
				URL:            "https://api.us-2.crowdstrike.com",
				Mode:           DataSourceModeNGSIEM,
				OAuth2:         true,
				OAuth2ClientID: "cid",
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{
					SecureJsonDataKeyOAuth2ClientSecret: "sec",
				},
			},
		},
		{
			name: "ngsiem without oauth2 errors",
			cfg: Config{
				URL:  "https://api.us-2.crowdstrike.com",
				Mode: DataSourceModeNGSIEM,
			},
			wantErr: "NGSIEM mode requires OAuth2 client credentials",
		},
		{
			name: "multiple auth methods",
			cfg: Config{
				URL:                   "https://cloud.humio.com",
				Mode:                  DataSourceModeLogScale,
				AuthenticateWithToken: true,
				OAuth2:                true,
				OAuth2ClientID:        "cid",
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{
					SecureJsonDataKeyAccessToken:        "tok",
					SecureJsonDataKeyOAuth2ClientSecret: "sec",
				},
			},
			wantErr: "only one authentication method may be enabled",
		},
		{
			name: "unknown mode",
			cfg: Config{
				URL:  "https://cloud.humio.com",
				Mode: "bogus",
			},
			wantErr: `unknown mode "bogus"`,
		},
		{
			name: "negative timeout",
			cfg: Config{
				URL:     "https://cloud.humio.com",
				Mode:    DataSourceModeLogScale,
				Timeout: -1,
			},
			wantErr: "timeout must be non-negative",
		},
		{
			name: "data link happy path",
			cfg: Config{
				URL:  "https://cloud.humio.com",
				Mode: DataSourceModeLogScale,
				DataLinks: []DataLinkConfig{
					{Field: "traceId", MatcherRegex: `trace_id=(\w+)`, URL: "${__value.raw}"},
				},
			},
		},
		{
			name: "data link missing field",
			cfg: Config{
				URL:  "https://cloud.humio.com",
				Mode: DataSourceModeLogScale,
				DataLinks: []DataLinkConfig{
					{MatcherRegex: ".*"},
				},
			},
			wantErr: "dataLinks[0].field is required",
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
