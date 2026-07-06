package snowflakedatasource

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
		name         string
		example      string // SettingsExamples key ("" here means "use inline settings", not the default example)
		settings     backend.DataSourceInstanceSettings
		wantErr      string
		wantAuthType AuthType
		wantAccount  string
		wantSecrets  []SecureJsonDataKey
		wantSettings int // expected len(cfg.Settings); -1 to skip
	}{
		{
			// The default schema example intentionally has an empty password
			// placeholder, so LoadConfig's Validate step is expected to reject it.
			name:         "default example fails validation (empty password placeholder)",
			settings:     settingsFromExample(t, ""),
			wantErr:      "invalid password",
			wantSettings: -1,
		},
		{
			name:         "password example",
			settings:     settingsFromExample(t, "passwordAuth"),
			wantAuthType: AuthTypePassword,
			wantAccount:  "myorg-myaccount",
			wantSecrets:  []SecureJsonDataKey{SecureJsonDataKeyPassword},
			wantSettings: -1,
		},
		{
			name:         "key pair example",
			settings:     settingsFromExample(t, "keyPair"),
			wantAuthType: AuthTypeKeyPair,
			wantAccount:  "myorg-myaccount",
			wantSecrets:  []SecureJsonDataKey{SecureJsonDataKeyPrivateKey},
			wantSettings: -1,
		},
		{
			name:         "encrypted key pair example",
			settings:     settingsFromExample(t, "keyPairEncrypted"),
			wantAuthType: AuthTypeKeyPair,
			wantAccount:  "myorg-myaccount",
			wantSecrets:  []SecureJsonDataKey{SecureJsonDataKeyPrivateKey, SecureJsonDataKeyPrivateKeyPassphrase},
			wantSettings: -1,
		},
		{
			name:         "programmatic access token example",
			settings:     settingsFromExample(t, "programmaticAccessToken"),
			wantAuthType: AuthTypePAT,
			wantAccount:  "myorg-myaccount",
			wantSecrets:  []SecureJsonDataKey{SecureJsonDataKeyPATToken},
			wantSettings: -1,
		},
		{
			name:         "oauth example",
			settings:     settingsFromExample(t, "oauth"),
			wantAuthType: AuthTypeOauth,
			wantAccount:  "myorg-myaccount",
			wantSecrets:  []SecureJsonDataKey{},
			wantSettings: -1,
		},
		{
			name:         "session parameters example",
			settings:     settingsFromExample(t, "sessionParameters"),
			wantAuthType: AuthTypePassword,
			wantAccount:  "myorg-myaccount",
			wantSecrets:  []SecureJsonDataKey{SecureJsonDataKeyPassword},
			wantSettings: 1,
		},
		{
			name: "empty authType with password defaults to password",
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"account":"acc","username":"u"}`),
				DecryptedSecureJSONData: map[string]string{"password": "pw"},
			},
			wantAuthType: AuthTypePassword,
			wantAccount:  "acc",
			wantSecrets:  []SecureJsonDataKey{SecureJsonDataKeyPassword},
			wantSettings: -1,
		},
		{
			name: "invalid jsonData errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{`),
			},
			wantErr:      "parse jsonData",
			wantSettings: -1,
		},
		{
			name:         "empty settings default to password and fail validation",
			settings:     backend.DataSourceInstanceSettings{},
			wantErr:      "invalid password",
			wantSettings: -1,
		},
		{
			name: "keypair with non-PEM private key errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"authType":"keypair"}`),
				DecryptedSecureJSONData: map[string]string{"privateKey": "not-a-pem"},
			},
			wantErr:      "invalid private key",
			wantSettings: -1,
		},
		{
			name: "keypair with empty private key errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"authType":"keypair"}`),
			},
			wantErr:      "invalid private key",
			wantSettings: -1,
		},
		{
			name: "pat with empty token errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"authType":"pat"}`),
			},
			wantErr:      "invalid programmatic access token",
			wantSettings: -1,
		},
		{
			name: "oauth without passthru errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"authType":"oauth"}`),
			},
			wantErr:      "you must enable Forward OAuth Identity",
			wantSettings: -1,
		},
		{
			name: "unknown authType errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"authType":"bogus"}`),
			},
			wantErr:      `unknown authType "bogus"`,
			wantSettings: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := LoadConfig(t.Context(), tt.settings)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("LoadConfig: expected error %q, got nil", tt.wantErr)
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
			if tt.wantAccount != "" && cfg.Account != tt.wantAccount {
				t.Errorf("Account = %q, want %q", cfg.Account, tt.wantAccount)
			}
			if tt.wantSecrets != nil {
				gotKeys := []SecureJsonDataKey{}
				for _, key := range SecureJsonDataKeys {
					if _, ok := cfg.DecryptedSecureJSONData[key]; ok {
						gotKeys = append(gotKeys, key)
					}
				}
				if !reflect.DeepEqual(gotKeys, tt.wantSecrets) {
					t.Errorf("configured secure keys = %v, want %v", gotKeys, tt.wantSecrets)
				}
			}
			if tt.wantSettings >= 0 && len(cfg.Settings) != tt.wantSettings {
				t.Errorf("len(Settings) = %d, want %d", len(cfg.Settings), tt.wantSettings)
			}
		})
	}
}

func TestApplyDefaults(t *testing.T) {
	tests := []struct {
		name string
		in   Config
		want AuthType
	}{
		{"empty config gets password", Config{}, AuthTypePassword},
		{"existing auth type is preserved", Config{AuthType: AuthTypeKeyPair}, AuthTypeKeyPair},
		{"pat preserved", Config{AuthType: AuthTypePAT}, AuthTypePAT},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.in
			got.ApplyDefaults()
			if got.AuthType != tt.want {
				t.Errorf("AuthType = %q, want %q", got.AuthType, tt.want)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	withSecret := func(k SecureJsonDataKey, v string) map[SecureJsonDataKey]string {
		return map[SecureJsonDataKey]string{k: v}
	}

	tests := []struct {
		name    string
		cfg     Config
		wantErr string
	}{
		{
			name: "password with secret",
			cfg:  Config{AuthType: AuthTypePassword, DecryptedSecureJSONData: withSecret(SecureJsonDataKeyPassword, "pw")},
		},
		{
			name:    "password without secret errors",
			cfg:     Config{AuthType: AuthTypePassword, DecryptedSecureJSONData: map[SecureJsonDataKey]string{}},
			wantErr: "invalid password",
		},
		{
			name:    "empty auth type treated as password",
			cfg:     Config{AuthType: AuthTypeUnknown, DecryptedSecureJSONData: map[SecureJsonDataKey]string{}},
			wantErr: "invalid password",
		},
		{
			name: "keypair with valid PEM",
			cfg:  Config{AuthType: AuthTypeKeyPair, DecryptedSecureJSONData: withSecret(SecureJsonDataKeyPrivateKey, examplePrivateKeyPEM)},
		},
		{
			name: "keypair with valid encrypted PEM",
			cfg:  Config{AuthType: AuthTypeKeyPair, DecryptedSecureJSONData: withSecret(SecureJsonDataKeyPrivateKey, exampleEncryptedPrivateKeyPEM)},
		},
		{
			name:    "keypair with garbage errors",
			cfg:     Config{AuthType: AuthTypeKeyPair, DecryptedSecureJSONData: withSecret(SecureJsonDataKeyPrivateKey, "garbage")},
			wantErr: "invalid private key",
		},
		{
			name:    "keypair missing key errors",
			cfg:     Config{AuthType: AuthTypeKeyPair, DecryptedSecureJSONData: map[SecureJsonDataKey]string{}},
			wantErr: "invalid private key",
		},
		{
			name: "pat with token",
			cfg:  Config{AuthType: AuthTypePAT, DecryptedSecureJSONData: withSecret(SecureJsonDataKeyPATToken, "tok")},
		},
		{
			name:    "pat without token errors",
			cfg:     Config{AuthType: AuthTypePAT, DecryptedSecureJSONData: map[SecureJsonDataKey]string{}},
			wantErr: "invalid programmatic access token",
		},
		{
			name: "oauth with passthru",
			cfg:  Config{AuthType: AuthTypeOauth, OAuthPassThrough: true},
		},
		{
			name:    "oauth without passthru errors",
			cfg:     Config{AuthType: AuthTypeOauth},
			wantErr: "you must enable Forward OAuth Identity",
		},
		{
			name:    "unknown auth type errors",
			cfg:     Config{AuthType: "bogus"},
			wantErr: `unknown authType "bogus"`,
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
