package lookerdatasource

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
		name           string
		example        string
		settings       backend.DataSourceInstanceSettings
		useSettings    bool
		wantErr        string // empty = expect no error; otherwise substring match
		wantAuthType   AuthType
		wantBaseURL    string
		wantClientID   string
		wantSecret     string
		wantSecureKeys SecureJsonDataConfig
	}{
		{
			// The default example intentionally has an empty base_url (and
			// client_id/client_secret) placeholder, so LoadConfig's Validate
			// step rejects it.
			name:    "default example fails validation (empty base_url placeholder)",
			example: "",
			wantErr: "invalid/empty Looker base url",
		},
		{
			name:           "client secret example",
			example:        "clientSecret",
			wantAuthType:   AuthTypeClientSecret,
			wantBaseURL:    "https://your-instance.looker.app",
			wantClientID:   "<your-client-id>",
			wantSecret:     "<your-client-secret>",
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyClientSecret},
		},
		{
			name:        "empty settings default to client_secret and fail validation",
			useSettings: true,
			settings:    backend.DataSourceInstanceSettings{},
			wantErr:     "invalid/empty Looker base url",
		},
		{
			name:        "invalid jsonData errors",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{`),
			},
			wantErr: "parse jsonData",
		},
		{
			name:        "base_url only reports the missing credentials",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"base_url":"https://x.looker.app"}`),
			},
			wantErr: "invalid/empty Looker client id",
		},
		{
			name:        "trailing slash and whitespace are trimmed",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"base_url":"  https://x.looker.app/  ","client_id":"  cid  "}`),
				DecryptedSecureJSONData: map[string]string{"client_secret": "  secret  "},
			},
			wantAuthType:   AuthTypeClientSecret,
			wantBaseURL:    "https://x.looker.app",
			wantClientID:   "cid",
			wantSecret:     "secret",
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyClientSecret},
		},
		{
			name:        "explicit auth_type is preserved",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"base_url":"https://x.looker.app","auth_type":"client_secret","client_id":"cid"}`),
				DecryptedSecureJSONData: map[string]string{"client_secret": "secret"},
			},
			wantAuthType: AuthTypeClientSecret,
			wantBaseURL:  "https://x.looker.app",
			wantClientID: "cid",
			wantSecret:   "secret",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings := tt.settings
			if !tt.useSettings {
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

			if tt.wantAuthType != "" && cfg.AuthType != tt.wantAuthType {
				t.Errorf("AuthType = %q, want %q", cfg.AuthType, tt.wantAuthType)
			}
			if tt.wantBaseURL != "" && cfg.BaseURL != tt.wantBaseURL {
				t.Errorf("BaseURL = %q, want %q", cfg.BaseURL, tt.wantBaseURL)
			}
			if tt.wantClientID != "" && cfg.ClientId != tt.wantClientID {
				t.Errorf("ClientId = %q, want %q", cfg.ClientId, tt.wantClientID)
			}
			if tt.wantSecret != "" && cfg.DecryptedSecureJSONData[SecureJsonDataKeyClientSecret] != tt.wantSecret {
				t.Errorf("Secrets[client_secret] = %q, want %q", cfg.DecryptedSecureJSONData[SecureJsonDataKeyClientSecret], tt.wantSecret)
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
			name: "empty config gets client_secret auth type",
			in:   Config{},
			want: Config{AuthType: AuthTypeClientSecret},
		},
		{
			name: "existing auth type is preserved",
			in:   Config{AuthType: AuthTypeClientSecret},
			want: Config{AuthType: AuthTypeClientSecret},
		},
		{
			name: "base_url trailing slash and whitespace stripped",
			in:   Config{BaseURL: "  https://x.looker.app/  "},
			want: Config{AuthType: AuthTypeClientSecret, BaseURL: "https://x.looker.app"},
		},
		{
			name: "client_id whitespace trimmed",
			in:   Config{ClientId: "  cid  "},
			want: Config{AuthType: AuthTypeClientSecret, ClientId: "cid"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.in
			got.ApplyDefaults()
			if got.AuthType != tt.want.AuthType {
				t.Errorf("AuthType = %q, want %q", got.AuthType, tt.want.AuthType)
			}
			if got.BaseURL != tt.want.BaseURL {
				t.Errorf("BaseURL = %q, want %q", got.BaseURL, tt.want.BaseURL)
			}
			if got.ClientId != tt.want.ClientId {
				t.Errorf("ClientId = %q, want %q", got.ClientId, tt.want.ClientId)
			}
		})
	}
}

func TestApplyDefaultsTrimsSecret(t *testing.T) {
	cfg := Config{
		DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyClientSecret: "  secret  "},
	}
	cfg.ApplyDefaults()
	if got := cfg.DecryptedSecureJSONData[SecureJsonDataKeyClientSecret]; got != "secret" {
		t.Errorf("client_secret = %q, want %q", got, "secret")
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr string // empty = expect no error; otherwise substring match
	}{
		{
			name: "client_secret happy path",
			cfg: Config{
				AuthType:                AuthTypeClientSecret,
				BaseURL:                 "https://x.looker.app",
				ClientId:                "cid",
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyClientSecret: "secret"},
			},
		},
		{
			name: "missing base_url errors",
			cfg: Config{
				AuthType:                AuthTypeClientSecret,
				ClientId:                "cid",
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyClientSecret: "secret"},
			},
			wantErr: "invalid/empty Looker base url",
		},
		{
			name: "missing client_id errors",
			cfg: Config{
				AuthType:                AuthTypeClientSecret,
				BaseURL:                 "https://x.looker.app",
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyClientSecret: "secret"},
			},
			wantErr: "invalid/empty Looker client id",
		},
		{
			name: "missing client_secret errors",
			cfg: Config{
				AuthType: AuthTypeClientSecret,
				BaseURL:  "https://x.looker.app",
				ClientId: "cid",
			},
			wantErr: "invalid/empty Looker client secret",
		},
		{
			name:    "all missing reports base url first",
			cfg:     Config{AuthType: AuthTypeClientSecret},
			wantErr: "invalid/empty Looker base url",
		},
		{
			name: "empty auth type only requires base url",
			cfg:  Config{BaseURL: "https://x.looker.app"},
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

// TestSettingsExamples guards the example set: the default "" example exists,
// and every example carries a jsonData object plus a non-empty secureJsonData
// using only known secret keys.
func TestSettingsExamples(t *testing.T) {
	examples := SettingsExamples().Examples
	if _, ok := examples[""]; !ok {
		t.Fatalf(`missing default "" example`)
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
			t.Errorf("example %q has no jsonData object", key)
		}
		secure, ok := value["secureJsonData"].(map[string]any)
		if !ok || len(secure) == 0 {
			t.Errorf("example %q has no secureJsonData", key)
			continue
		}
		for secretKey := range secure {
			if !known[secretKey] {
				t.Errorf("example %q references unknown secret key %q", key, secretKey)
			}
		}
	}
}
