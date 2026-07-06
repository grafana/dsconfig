package sentrydatasource

import (
	"encoding/json"
	"errors"
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
		name           string
		example        string // schema.go SettingsExamples key
		settings       backend.DataSourceInstanceSettings
		useSettings    bool
		wantErr        error
		wantURL        string
		wantOrgSlug    string
		wantTLSSkip    bool
		wantSecureKeys SecureJsonDataConfig
		wantAuthToken  string
	}{
		{
			// The default schema example has an empty orgSlug and empty
			// authToken placeholder, so LoadConfig's Validate step is
			// expected to reject it.
			name:    "default example fails validation (empty orgSlug and authToken)",
			example: "",
			wantErr: errors.New("organization slug"),
		},
		{
			name:           "sentry saas",
			example:        "sentrySaaS",
			wantURL:        DefaultSentryURL,
			wantOrgSlug:    "example-org",
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyAuthToken},
			wantAuthToken:  "sntrys_XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
		},
		{
			name:           "self hosted",
			example:        "selfHosted",
			wantURL:        "https://sentry.example.com",
			wantOrgSlug:    "example-org",
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyAuthToken},
			wantAuthToken:  "sntrys_XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
		},
		{
			name:           "self hosted with tls skip verify",
			example:        "selfHostedTLSSkipVerify",
			wantURL:        "https://sentry.internal.corp",
			wantOrgSlug:    "example-org",
			wantTLSSkip:    true,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyAuthToken},
			wantAuthToken:  "sntrys_XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
		},
		{
			name:           "legacy missing url defaults to sentry.io",
			example:        "legacyMissingURL",
			wantURL:        DefaultSentryURL,
			wantOrgSlug:    "example-org",
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyAuthToken},
			wantAuthToken:  "sntrys_XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
		},
		{
			// Empty JSONData is a parse error upstream — pkg/plugin/settings.go
			// unconditionally json.Unmarshal(nil, cfg) which fails.
			name:        "empty settings error (empty JSONData)",
			useSettings: true,
			settings:    backend.DataSourceInstanceSettings{},
			wantErr:     errors.New("parse jsonData"),
		},
		{
			name:        "invalid jsonData errors",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{`),
			},
			wantErr: errors.New("parse jsonData"),
		},
		{
			name:        "missing org slug errors",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"url":"https://sentry.io"}`),
				DecryptedSecureJSONData: map[string]string{"authToken": "tok"},
			},
			wantErr: errors.New("organization slug"),
		},
		{
			name:        "missing auth token errors",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"url":"https://sentry.io","orgSlug":"acme"}`),
			},
			wantErr: errors.New("auth token"),
		},
		{
			name:        "empty url defaults to sentry.io",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"orgSlug":"acme"}`),
				DecryptedSecureJSONData: map[string]string{"authToken": "tok"},
			},
			wantURL:     DefaultSentryURL,
			wantOrgSlug: "acme",
		},
		{
			name:        "tls skip verify parses",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"url":"https://s","orgSlug":"acme","tlsSkipVerify":true}`),
				DecryptedSecureJSONData: map[string]string{"authToken": "tok"},
			},
			wantURL:     "https://s",
			wantOrgSlug: "acme",
			wantTLSSkip: true,
		},
		{
			// enableSecureSocksProxy is intentionally omitted from both
			// the dsconfig schema and the Go Config struct; json unmarshal
			// silently ignores unknown fields.
			name:        "unknown enableSecureSocksProxy field is ignored",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"url":"https://s","orgSlug":"acme","enableSecureSocksProxy":true}`),
				DecryptedSecureJSONData: map[string]string{"authToken": "tok"},
			},
			wantURL:     "https://s",
			wantOrgSlug: "acme",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings := tt.settings
			if !tt.useSettings {
				settings = settingsFromExample(t, tt.example)
			}

			cfg, err := LoadConfig(t.Context(), settings)
			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("LoadConfig: expected error containing %q, got nil", tt.wantErr)
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
			if tt.wantOrgSlug != "" && cfg.OrgSlug != tt.wantOrgSlug {
				t.Errorf("OrgSlug = %q, want %q", cfg.OrgSlug, tt.wantOrgSlug)
			}
			if cfg.TLSSkipVerify != tt.wantTLSSkip {
				t.Errorf("TLSSkipVerify = %v, want %v", cfg.TLSSkipVerify, tt.wantTLSSkip)
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
			if tt.wantAuthToken != "" && cfg.DecryptedSecureJSONData[SecureJsonDataKeyAuthToken] != tt.wantAuthToken {
				t.Errorf("DecryptedSecureJSONData[authToken] = %q, want %q", cfg.DecryptedSecureJSONData[SecureJsonDataKeyAuthToken], tt.wantAuthToken)
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
			name: "empty url defaults to sentry.io",
			in:   Config{},
			want: Config{URL: DefaultSentryURL},
		},
		{
			name: "explicit url is preserved",
			in:   Config{URL: "https://sentry.example.com"},
			want: Config{URL: "https://sentry.example.com"},
		},
		{
			name: "unrelated fields untouched",
			in:   Config{OrgSlug: "acme", TLSSkipVerify: true},
			want: Config{URL: DefaultSentryURL, OrgSlug: "acme", TLSSkipVerify: true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.in
			got.ApplyDefaults()
			if got.URL != tt.want.URL {
				t.Errorf("URL = %q, want %q", got.URL, tt.want.URL)
			}
			if got.OrgSlug != tt.want.OrgSlug {
				t.Errorf("OrgSlug = %q, want %q", got.OrgSlug, tt.want.OrgSlug)
			}
			if got.TLSSkipVerify != tt.want.TLSSkipVerify {
				t.Errorf("TLSSkipVerify = %v, want %v", got.TLSSkipVerify, tt.want.TLSSkipVerify)
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
			name: "happy path",
			cfg: Config{
				URL:                     DefaultSentryURL,
				OrgSlug:                 "acme",
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyAuthToken: "tok"},
			},
		},
		{
			name: "self-hosted with tls skip",
			cfg: Config{
				URL:                     "https://sentry.example.com",
				OrgSlug:                 "acme",
				TLSSkipVerify:           true,
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyAuthToken: "tok"},
			},
		},
		{
			name:    "empty url errors",
			cfg:     Config{OrgSlug: "acme", DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyAuthToken: "tok"}},
			wantErr: "sentry URL",
		},
		{
			name:    "empty org slug errors",
			cfg:     Config{URL: DefaultSentryURL, DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyAuthToken: "tok"}},
			wantErr: "organization slug",
		},
		{
			name:    "empty auth token errors",
			cfg:     Config{URL: DefaultSentryURL, OrgSlug: "acme"},
			wantErr: "auth token",
		},
		{
			name:    "everything empty joins all errors",
			cfg:     Config{},
			wantErr: "sentry URL",
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
