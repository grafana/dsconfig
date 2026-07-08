package astradbdatasource

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
		example        string // schema.go SettingsExamples key ("" = use inline settings below)
		settings       backend.DataSourceInstanceSettings
		wantErr        string // substring match; empty = expect no error
		wantAuthKind   AuthKind
		checkAuthKind  bool
		wantURI        string
		wantGRPC       string
		wantAuthEP     string
		wantUser       string
		wantSecure     bool
		checkSecure    bool
		wantSecureKeys SecureJsonDataConfig
		wantToken      string
		wantPassword   string
	}{
		{
			// The default schema example intentionally has an empty token
			// placeholder AND an empty uri, so Validate is expected to reject
			// with two joined errors.
			name:    "default example fails validation (empty uri and token)",
			example: "",
			wantErr: "uri is required",
		},
		{
			name:           "token astra cloud",
			example:        "tokenAstraCloud",
			checkAuthKind:  true,
			wantAuthKind:   AuthKindToken,
			wantURI:        "cluster-id-region.apps.astra.datastax.com:443",
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyToken},
			wantToken:      "AstraCS:XXXXXXXXXXXXXXXXXXXXXXXX:XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
		},
		{
			name:           "credentials self-hosted TLS",
			example:        "credentialsSelfHostedTLS",
			checkAuthKind:  true,
			wantAuthKind:   AuthKindCredentials,
			wantGRPC:       "stargate.example.com:8090",
			wantAuthEP:     "stargate.example.com:8081",
			wantUser:       "cassandra",
			wantSecure:     true,
			checkSecure:    true,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyPassword},
			wantPassword:   "cassandra",
		},
		{
			name:           "credentials self-hosted plaintext",
			example:        "credentialsSelfHostedPlaintext",
			checkAuthKind:  true,
			wantAuthKind:   AuthKindCredentials,
			wantGRPC:       "localhost:8090",
			wantAuthEP:     "localhost:8081",
			wantUser:       "cassandra",
			wantSecure:     false,
			checkSecure:    true,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyPassword},
			wantPassword:   "cassandra",
		},
		{
			name:           "legacy missing authKind defaults to token",
			example:        "legacyMissingAuthKind",
			checkAuthKind:  true,
			wantAuthKind:   AuthKindToken,
			wantURI:        "cluster-id-region.apps.astra.datastax.com:443",
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyToken},
			wantToken:      "AstraCS:XXXXXXXXXXXXXXXXXXXXXXXX:XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
		},
		{
			name: "invalid jsonData errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{`),
			},
			wantErr: "parse jsonData",
		},
		{
			name: "token mode missing uri errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"authKind":0}`),
				DecryptedSecureJSONData: map[string]string{"token": "AstraCS:tok"},
			},
			wantErr: "uri is required",
		},
		{
			name: "token mode missing token errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"authKind":0,"uri":"cluster-id-region.apps.astra.datastax.com:443"}`),
			},
			wantErr: "token is required",
		},
		{
			name: "credentials mode missing every field errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"authKind":1}`),
			},
			wantErr: "grpcEndpoint is required",
		},
		{
			name: "credentials mode missing password errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"authKind":1,"grpcEndpoint":"h:1","authEndpoint":"h:2","user":"u"}`),
			},
			wantErr: "password is required",
		},
		{
			name: "unknown authKind errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"authKind":99}`),
			},
			wantErr: "unknown authKind 99",
		},
		{
			// After ApplyDefaults, an empty config becomes authKind=0
			// (Token), which requires uri and token — Validate rejects it.
			name:     "empty settings default to token and fail validation",
			settings: backend.DataSourceInstanceSettings{},
			wantErr:  "uri is required",
		},
		{
			name: "credentials mode with all fields loads",
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"authKind":1,"grpcEndpoint":"h:1","authEndpoint":"h:2","user":"u","secure":true}`),
				DecryptedSecureJSONData: map[string]string{"password": "p"},
			},
			checkAuthKind: true,
			wantAuthKind:  AuthKindCredentials,
			wantGRPC:      "h:1",
			wantAuthEP:    "h:2",
			wantUser:      "u",
			wantSecure:    true,
			checkSecure:   true,
			wantPassword:  "p",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings := tt.settings
			if tt.example != "" {
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

			if tt.checkAuthKind && cfg.AuthKind != tt.wantAuthKind {
				t.Errorf("AuthKind = %d, want %d", cfg.AuthKind, tt.wantAuthKind)
			}
			if tt.wantURI != "" && cfg.URI != tt.wantURI {
				t.Errorf("URI = %q, want %q", cfg.URI, tt.wantURI)
			}
			if tt.wantGRPC != "" && cfg.GRPCEndpoint != tt.wantGRPC {
				t.Errorf("GRPCEndpoint = %q, want %q", cfg.GRPCEndpoint, tt.wantGRPC)
			}
			if tt.wantAuthEP != "" && cfg.AuthEndpoint != tt.wantAuthEP {
				t.Errorf("AuthEndpoint = %q, want %q", cfg.AuthEndpoint, tt.wantAuthEP)
			}
			if tt.wantUser != "" && cfg.UserName != tt.wantUser {
				t.Errorf("UserName = %q, want %q", cfg.UserName, tt.wantUser)
			}
			if tt.checkSecure && cfg.Secure != tt.wantSecure {
				t.Errorf("Secure = %v, want %v", cfg.Secure, tt.wantSecure)
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
			if tt.wantToken != "" && cfg.DecryptedSecureJSONData[SecureJsonDataKeyToken] != tt.wantToken {
				t.Errorf("Secrets[token] = %q, want %q", cfg.DecryptedSecureJSONData[SecureJsonDataKeyToken], tt.wantToken)
			}
			if tt.wantPassword != "" && cfg.DecryptedSecureJSONData[SecureJsonDataKeyPassword] != tt.wantPassword {
				t.Errorf("Secrets[password] = %q, want %q", cfg.DecryptedSecureJSONData[SecureJsonDataKeyPassword], tt.wantPassword)
			}
		})
	}
}

func TestApplyDefaults(t *testing.T) {
	tests := []struct {
		name string
		in   Config
		want AuthKind
	}{
		{
			name: "empty config gets token (zero value)",
			in:   Config{},
			want: AuthKindToken,
		},
		{
			name: "credentials preserved",
			in:   Config{AuthKind: AuthKindCredentials},
			want: AuthKindCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.in
			got.ApplyDefaults()
			if got.AuthKind != tt.want {
				t.Errorf("AuthKind = %d, want %d", got.AuthKind, tt.want)
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
			name: "token with uri and token",
			cfg: Config{
				AuthKind:                AuthKindToken,
				URI:                     "cluster-id-region.apps.astra.datastax.com:443",
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyToken: "AstraCS:tok"},
			},
		},
		{
			name:    "token missing uri errors",
			cfg:     Config{AuthKind: AuthKindToken, DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyToken: "AstraCS:tok"}},
			wantErr: "uri is required",
		},
		{
			name:    "token missing secret errors",
			cfg:     Config{AuthKind: AuthKindToken, URI: "host:443"},
			wantErr: "token is required",
		},
		{
			name: "credentials happy path",
			cfg: Config{
				AuthKind:                AuthKindCredentials,
				GRPCEndpoint:            "h:1",
				AuthEndpoint:            "h:2",
				UserName:                "u",
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyPassword: "p"},
			},
		},
		{
			name:    "credentials missing everything",
			cfg:     Config{AuthKind: AuthKindCredentials},
			wantErr: "grpcEndpoint is required",
		},
		{
			name: "credentials missing password",
			cfg: Config{
				AuthKind:     AuthKindCredentials,
				GRPCEndpoint: "h:1",
				AuthEndpoint: "h:2",
				UserName:     "u",
			},
			wantErr: "password is required",
		},
		{
			name:    "unknown auth kind errors",
			cfg:     Config{AuthKind: 42},
			wantErr: "unknown authKind 42",
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
