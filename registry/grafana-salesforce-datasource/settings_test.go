package salesforcedatasource

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
		wantAuthType   AuthType
		wantUser       string
		wantTokenURL   string
		wantSandbox    bool
		wantSecureKeys SecureJsonDataConfig
	}{
		{
			// The default schema example has an empty user and empty secret
			// placeholders, so LoadConfig's Validate step rejects it.
			name:    "default example fails validation (empty user and secrets)",
			example: "",
			wantErr: errors.New("invalid or empty username"),
		},
		{
			name:         "user credentials production",
			example:      "userCredentials",
			wantAuthType: AuthTypeUser,
			wantUser:     "user@example.com",
			wantTokenURL: TokenURLProd,
			wantSecureKeys: SecureJsonDataConfig{
				SecureJsonDataKeyPassword,
				SecureJsonDataKeySecurityToken,
				SecureJsonDataKeyClientID,
				SecureJsonDataKeyClientSecret,
			},
		},
		{
			name:         "user credentials sandbox",
			example:      "userCredentialsSandbox",
			wantAuthType: AuthTypeUser,
			wantUser:     "user@example.com.sandbox",
			wantTokenURL: TokenURLSandbox,
			wantSecureKeys: SecureJsonDataConfig{
				SecureJsonDataKeyPassword,
				SecureJsonDataKeySecurityToken,
				SecureJsonDataKeyClientID,
				SecureJsonDataKeyClientSecret,
			},
		},
		{
			name:         "jwt production",
			example:      "jwt",
			wantAuthType: AuthTypeJWT,
			wantUser:     "user@example.com",
			wantTokenURL: TokenURLProd,
			wantSecureKeys: SecureJsonDataConfig{
				SecureJsonDataKeyClientID,
				SecureJsonDataKeyCert,
				SecureJsonDataKeyPrivateKey,
			},
		},
		{
			name:         "jwt sandbox",
			example:      "jwtSandbox",
			wantAuthType: AuthTypeJWT,
			wantUser:     "user@example.com.sandbox",
			wantTokenURL: TokenURLSandbox,
			wantSecureKeys: SecureJsonDataConfig{
				SecureJsonDataKeyClientID,
				SecureJsonDataKeyCert,
				SecureJsonDataKeyPrivateKey,
			},
		},
		{
			// No authType -> defaults to user; no tokenUrl but sandbox=true ->
			// derives the sandbox host.
			name:         "legacy sandbox flag defaults to user and sandbox host",
			example:      "legacySandboxFlag",
			wantAuthType: AuthTypeUser,
			wantUser:     "user@example.com.sandbox",
			wantTokenURL: TokenURLSandbox,
			wantSandbox:  true,
			wantSecureKeys: SecureJsonDataConfig{
				SecureJsonDataKeyPassword,
				SecureJsonDataKeyClientID,
				SecureJsonDataKeyClientSecret,
			},
		},
		{
			// Empty JSONData is a parse error upstream — pkg/models/settings.go:43
			// json.Unmarshal(nil, settings) fails.
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
			name:        "user auth missing password errors",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"authType":"user","user":"u"}`),
				DecryptedSecureJSONData: map[string]string{
					"clientID":     "id",
					"clientSecret": "secret",
				},
			},
			wantErr: errors.New("invalid or empty password"),
		},
		{
			name:        "jwt auth missing certificate errors",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"authType":"jwt","user":"u"}`),
				DecryptedSecureJSONData: map[string]string{"privateKey": "key"},
			},
			wantErr: errors.New("invalid or empty certificate"),
		},
		{
			name:        "explicit tokenUrl overrides sandbox flag",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"authType":"user","user":"u","sandbox":true,"tokenUrl":"https://login.salesforce.com"}`),
				DecryptedSecureJSONData: map[string]string{
					"password":     "p",
					"clientID":     "id",
					"clientSecret": "secret",
				},
			},
			wantAuthType: AuthTypeUser,
			wantUser:     "u",
			wantTokenURL: TokenURLProd,
			wantSandbox:  true,
		},
		{
			// enableSecureSocksProxy is intentionally omitted from both the
			// dsconfig schema and the Go Config struct; json unmarshal silently
			// ignores unknown fields.
			name:        "unknown enableSecureSocksProxy field is ignored",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"authType":"user","user":"u","enableSecureSocksProxy":true}`),
				DecryptedSecureJSONData: map[string]string{
					"password":     "p",
					"clientID":     "id",
					"clientSecret": "secret",
				},
			},
			wantAuthType: AuthTypeUser,
			wantUser:     "u",
			wantTokenURL: TokenURLProd,
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

			if tt.wantAuthType != "" && cfg.AuthType != tt.wantAuthType {
				t.Errorf("AuthType = %q, want %q", cfg.AuthType, tt.wantAuthType)
			}
			if tt.wantUser != "" && cfg.User != tt.wantUser {
				t.Errorf("User = %q, want %q", cfg.User, tt.wantUser)
			}
			if tt.wantTokenURL != "" && cfg.TokenURL != tt.wantTokenURL {
				t.Errorf("TokenURL = %q, want %q", cfg.TokenURL, tt.wantTokenURL)
			}
			if cfg.Sandbox != tt.wantSandbox {
				t.Errorf("Sandbox = %v, want %v", cfg.Sandbox, tt.wantSandbox)
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
		name         string
		in           Config
		wantAuthType AuthType
		wantTokenURL string
	}{
		{
			name:         "empty config defaults to user + production",
			in:           Config{},
			wantAuthType: AuthTypeUser,
			wantTokenURL: TokenURLProd,
		},
		{
			name:         "empty tokenUrl with sandbox flag defaults to sandbox host",
			in:           Config{Sandbox: true},
			wantAuthType: AuthTypeUser,
			wantTokenURL: TokenURLSandbox,
		},
		{
			name:         "explicit auth type is preserved",
			in:           Config{AuthType: AuthTypeJWT},
			wantAuthType: AuthTypeJWT,
			wantTokenURL: TokenURLProd,
		},
		{
			name:         "explicit tokenUrl is preserved over sandbox flag",
			in:           Config{Sandbox: true, TokenURL: TokenURLProd},
			wantAuthType: AuthTypeUser,
			wantTokenURL: TokenURLProd,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.in
			got.ApplyDefaults()
			if got.AuthType != tt.wantAuthType {
				t.Errorf("AuthType = %q, want %q", got.AuthType, tt.wantAuthType)
			}
			if got.TokenURL != tt.wantTokenURL {
				t.Errorf("TokenURL = %q, want %q", got.TokenURL, tt.wantTokenURL)
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
			name: "user auth happy path",
			cfg: Config{
				AuthType: AuthTypeUser,
				User:     "u",
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{
					SecureJsonDataKeyPassword:     "p",
					SecureJsonDataKeyClientID:     "id",
					SecureJsonDataKeyClientSecret: "secret",
				},
			},
		},
		{
			name: "user auth does not require security token",
			cfg: Config{
				AuthType: AuthTypeUser,
				User:     "u",
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{
					SecureJsonDataKeyPassword:     "p",
					SecureJsonDataKeyClientID:     "id",
					SecureJsonDataKeyClientSecret: "secret",
				},
			},
		},
		{
			name:    "user auth missing everything joins errors",
			cfg:     Config{AuthType: AuthTypeUser},
			wantErr: "invalid or empty username",
		},
		{
			name: "jwt auth happy path",
			cfg: Config{
				AuthType: AuthTypeJWT,
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{
					SecureJsonDataKeyCert:       "cert",
					SecureJsonDataKeyPrivateKey: "key",
				},
			},
		},
		{
			name:    "jwt auth missing certificate errors",
			cfg:     Config{AuthType: AuthTypeJWT, DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyPrivateKey: "key"}},
			wantErr: "invalid or empty certificate",
		},
		{
			name:    "jwt auth missing private key errors",
			cfg:     Config{AuthType: AuthTypeJWT, DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyCert: "cert"}},
			wantErr: "invalid or empty private key",
		},
		{
			// Upstream treats any non-jwt authType as user-auth validation.
			name:    "empty auth type validates as user auth",
			cfg:     Config{},
			wantErr: "invalid or empty username",
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
