package mongodbdatasource

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// settingsFromExample converts a SettingsExamples entry (a full instance
// settings object with root fields, jsonData and secureJsonData) into the
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
	if user, ok := value["basicAuthUser"].(string); ok {
		settings.BasicAuthUser = user
	}
	if enabled, ok := value["basicAuth"].(bool); ok {
		settings.BasicAuthEnabled = enabled
	}
	return settings
}

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name                string
		example             string // SettingsExamples key ("" handled explicitly)
		settings            backend.DataSourceInstanceSettings
		useExample          bool
		wantErr             string
		wantAuthType        AuthType
		wantConnection      string
		wantBasicAuthUser   string
		wantBasicAuthPwd    string
		wantBasicAuthOn     bool
		wantKerberosEnabled bool
		wantTLSSkipVerify   bool
		wantSecrets         map[SecureJsonDataKey]string
	}{
		{
			// The default example intentionally has an empty connection, so
			// LoadConfig's Validate step is expected to reject it.
			name:       "default example fails validation (empty connection)",
			example:    "",
			useExample: true,
			wantErr:    "connection string",
		},
		{
			name:              "credentials example",
			example:           "credentials",
			useExample:        true,
			wantAuthType:      AuthTypeBasicAuth,
			wantConnection:    "mongodb://mongodb.example.com:27017/mydb",
			wantBasicAuthUser: "grafana_reader",
			wantBasicAuthPwd:  "<your-password>",
			wantBasicAuthOn:   true,
			wantSecrets:       map[SecureJsonDataKey]string{SecureJsonDataKeyBasicAuthPassword: "<your-password>"},
		},
		{
			name:           "no auth example",
			example:        "noAuth",
			useExample:     true,
			wantAuthType:   AuthTypeNoAuth,
			wantConnection: "mongodb://mongodb.example.com:27017/mydb",
		},
		{
			name:                "kerberos example",
			example:             "kerberos",
			useExample:          true,
			wantAuthType:        AuthTypeKerberos,
			wantConnection:      "mongodb://mongodb.example.com:27017/?authMechanism=GSSAPI",
			wantKerberosEnabled: true,
			wantSecrets:         map[SecureJsonDataKey]string{SecureJsonDataKeyKerberosPassword: "<kerberos-password>"},
		},
		{
			name:            "tls client auth example",
			example:         "tlsClientAuth",
			useExample:      true,
			wantAuthType:    AuthTypeBasicAuth,
			wantConnection:  "mongodb://mongodb.example.com:27017/?tls=true",
			wantBasicAuthOn: true,
			wantSecrets: map[SecureJsonDataKey]string{
				SecureJsonDataKeyTLSCACert:     "-----BEGIN CERTIFICATE-----\n<redacted>\n-----END CERTIFICATE-----",
				SecureJsonDataKeyTLSClientCert: "-----BEGIN CERTIFICATE-----\n<redacted>\n-----END CERTIFICATE-----",
				SecureJsonDataKeyTLSClientKey:  "-----BEGIN PRIVATE KEY-----\n<redacted>\n-----END PRIVATE KEY-----",
			},
		},
		{
			name:              "legacy credentials example migrates to basic auth",
			example:           "legacyCredentials",
			useExample:        true,
			wantAuthType:      AuthTypeBasicAuth, // defaulted (no authType in legacy config)
			wantConnection:    "mongodb://mongodb.example.com:27017/mydb",
			wantBasicAuthUser: "grafana_reader",
			wantBasicAuthPwd:  "<your-password>",
			wantBasicAuthOn:   true,
		},
		{
			name: "empty authType defaults to BasicAuth",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"authType":"","connection":"mongodb://host:27017/db"}`),
			},
			wantAuthType:   AuthTypeBasicAuth,
			wantConnection: "mongodb://host:27017/db",
		},
		{
			name: "unknown authType coerced to BasicAuth",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"authType":"bogus","connection":"mongodb://host:27017/db"}`),
			},
			wantAuthType: AuthTypeBasicAuth,
		},
		{
			name: "root basicAuthUser overrides legacy jsonData user",
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"connection":"mongodb://host:27017/db","user":"legacy-user"}`),
				BasicAuthUser:           "root-user",
				DecryptedSecureJSONData: map[string]string{"password": "legacy-pw"},
			},
			wantAuthType:      AuthTypeBasicAuth,
			wantBasicAuthUser: "root-user",
			wantBasicAuthPwd:  "legacy-pw",
			wantBasicAuthOn:   true,
		},
		{
			name: "modern basicAuthPassword overrides legacy password",
			settings: backend.DataSourceInstanceSettings{
				JSONData:      []byte(`{"connection":"mongodb://host:27017/db","authType":"BasicAuth"}`),
				BasicAuthUser: "u",
				DecryptedSecureJSONData: map[string]string{
					"password":          "legacy-pw",
					"basicAuthPassword": "new-pw",
				},
			},
			wantBasicAuthPwd: "new-pw",
			wantBasicAuthOn:  true,
		},
		{
			name: "kerberos not enabled without GSSAPI in connection",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"authType":"custom-Kerberos","connection":"mongodb://host:27017/db","kerberosUser":"u@REALM"}`),
			},
			wantAuthType:        AuthTypeKerberos,
			wantKerberosEnabled: false,
		},
		{
			name: "kerberos enabled with GSSAPI and user",
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"authType":"custom-Kerberos","connection":"mongodb://host:27017/?authMechanism=GSSAPI","kerberosUser":"u@REALM"}`),
				DecryptedSecureJSONData: map[string]string{"kerberosPassword": "pw"},
			},
			wantAuthType:        AuthTypeKerberos,
			wantKerberosEnabled: true,
		},
		{
			name: "legacy skipTLSValidation copied into tlsSkipVerify",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"connection":"mongodb://host:27017/db","authType":"NoAuth","skipTLSValidation":true}`),
			},
			wantTLSSkipVerify: true,
		},
		{
			name: "invalid jsonData errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{`),
			},
			wantErr: "parse jsonData",
		},
		{
			name: "missing connection fails",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"authType":"BasicAuth"}`),
			},
			wantErr: "connection string",
		},
		{
			name: "tlsAuth without cert+key fails",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"connection":"mongodb://host:27017/db","authType":"NoAuth","tlsAuth":true}`),
			},
			wantErr: "tlsClientCert",
		},
		{
			name: "tlsAuthWithCACert without CA fails",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"connection":"mongodb://host:27017/db","authType":"NoAuth","tlsAuthWithCACert":true}`),
			},
			wantErr: "tlsCACert",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings := tt.settings
			if tt.useExample {
				settings = settingsFromExample(t, tt.example)
			}

			cfg, err := LoadConfig(context.Background(), settings)
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
				t.Fatalf("LoadConfig: unexpected error: %v", err)
			}

			if tt.wantAuthType != "" && cfg.AuthType != tt.wantAuthType {
				t.Errorf("AuthType = %q, want %q", cfg.AuthType, tt.wantAuthType)
			}
			if tt.wantConnection != "" && cfg.Connection != tt.wantConnection {
				t.Errorf("Connection = %q, want %q", cfg.Connection, tt.wantConnection)
			}
			if tt.wantBasicAuthUser != "" && cfg.BasicAuthUser != tt.wantBasicAuthUser {
				t.Errorf("BasicAuthUser = %q, want %q", cfg.BasicAuthUser, tt.wantBasicAuthUser)
			}
			if tt.wantBasicAuthPwd != "" && cfg.BasicAuthPassword() != tt.wantBasicAuthPwd {
				t.Errorf("BasicAuthPassword() = %q, want %q", cfg.BasicAuthPassword(), tt.wantBasicAuthPwd)
			}
			if tt.wantBasicAuthOn && !cfg.BasicAuthEnabled {
				t.Errorf("BasicAuthEnabled = false, want true")
			}
			if cfg.KerberosEnabled() != tt.wantKerberosEnabled {
				t.Errorf("KerberosEnabled() = %v, want %v", cfg.KerberosEnabled(), tt.wantKerberosEnabled)
			}
			if cfg.TLSSkipVerify != tt.wantTLSSkipVerify {
				t.Errorf("TLSSkipVerify = %v, want %v", cfg.TLSSkipVerify, tt.wantTLSSkipVerify)
			}
			for k, want := range tt.wantSecrets {
				if got := cfg.DecryptedSecureJSONData[k]; got != want {
					t.Errorf("DecryptedSecureJSONData[%s] = %q, want %q", k, got, want)
				}
			}
		})
	}
}

func TestApplyDefaults(t *testing.T) {
	tests := []struct {
		name     string
		in       Config
		wantAuth AuthType
		wantRows string
	}{
		{"empty gets BasicAuth + 10000", Config{}, AuthTypeBasicAuth, "10000"},
		{"invalid auth coerced to BasicAuth", Config{AuthType: "bogus"}, AuthTypeBasicAuth, "10000"},
		{"valid auth preserved", Config{AuthType: AuthTypeKerberos}, AuthTypeKerberos, "10000"},
		{"explicit rows preserved", Config{AuthType: AuthTypeNoAuth, ResponseRowsLimit: "500"}, AuthTypeNoAuth, "500"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.in
			got.ApplyDefaults()
			if got.AuthType != tt.wantAuth {
				t.Errorf("AuthType = %q, want %q", got.AuthType, tt.wantAuth)
			}
			if got.ResponseRowsLimit != tt.wantRows {
				t.Errorf("ResponseRowsLimit = %q, want %q", got.ResponseRowsLimit, tt.wantRows)
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
			name: "basic ok",
			cfg:  Config{Connection: "mongodb://h:27017/db", AuthType: AuthTypeBasicAuth},
		},
		{
			name: "noauth ok",
			cfg:  Config{Connection: "mongodb://h:27017/db", AuthType: AuthTypeNoAuth},
		},
		{
			name:    "missing connection",
			cfg:     Config{AuthType: AuthTypeBasicAuth},
			wantErr: "connection string",
		},
		{
			name:    "empty auth type",
			cfg:     Config{Connection: "mongodb://h:27017/db"},
			wantErr: "authType is required",
		},
		{
			name:    "unknown auth type",
			cfg:     Config{Connection: "mongodb://h:27017/db", AuthType: "bogus"},
			wantErr: `unknown authType "bogus"`,
		},
		{
			name:    "tlsAuth needs cert+key",
			cfg:     Config{Connection: "mongodb://h:27017/db", AuthType: AuthTypeNoAuth, TLSAuth: true},
			wantErr: "tlsClientCert",
		},
		{
			name: "tlsAuth ok with cert+key",
			cfg: Config{
				Connection: "mongodb://h:27017/db", AuthType: AuthTypeNoAuth, TLSAuth: true,
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{
					SecureJsonDataKeyTLSClientCert: "CERT",
					SecureJsonDataKeyTLSClientKey:  "KEY",
				},
			},
		},
		{
			name:    "tlsAuthWithCACert needs CA",
			cfg:     Config{Connection: "mongodb://h:27017/db", AuthType: AuthTypeNoAuth, TLSAuthWithCACert: true},
			wantErr: "tlsCACert",
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

func TestKerberosEnabled(t *testing.T) {
	if (Config{Connection: "mongodb://h/?authMechanism=GSSAPI", KerberosUser: "u"}).KerberosEnabled() != true {
		t.Errorf("expected KerberosEnabled true")
	}
	if (Config{Connection: "mongodb://h/?authMechanism=GSSAPI"}).KerberosEnabled() != false {
		t.Errorf("expected false without kerberosUser")
	}
	if (Config{Connection: "mongodb://h/db", KerberosUser: "u"}).KerberosEnabled() != false {
		t.Errorf("expected false without GSSAPI")
	}
}

func TestBasicAuthPassword(t *testing.T) {
	if got := (Config{DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyBasicAuthPassword: "new", SecureJsonDataKeyPassword: "old"}}).BasicAuthPassword(); got != "new" {
		t.Errorf("modern precedence: got %q", got)
	}
	if got := (Config{DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyPassword: "old"}}).BasicAuthPassword(); got != "old" {
		t.Errorf("legacy fallback: got %q", got)
	}
	if got := (Config{}).BasicAuthPassword(); got != "" {
		t.Errorf("empty: got %q", got)
	}
}

// TestSettingsExamplesShape guards the example matrix: the default example keyed
// by "" must exist, every example must carry jsonData, and every secureJsonData
// key must be a known secret. All examples except the credential-free "noAuth"
// method carry at least one secret.
func TestSettingsExamplesShape(t *testing.T) {
	known := map[string]bool{}
	for _, k := range SecureJsonDataKeys {
		known[string(k)] = true
	}

	examples := SettingsExamples().Examples
	if _, ok := examples[""]; !ok {
		t.Fatalf("missing default example keyed by \"\"")
	}

	for key, ex := range examples {
		value, ok := ex.Value.(map[string]any)
		if !ok {
			t.Fatalf("example %q value is not an object", key)
		}
		if _, ok := value["jsonData"]; !ok {
			t.Errorf("example %q missing jsonData", key)
		}
		secure, ok := value["secureJsonData"].(map[string]any)
		if !ok {
			t.Errorf("example %q missing secureJsonData", key)
			continue
		}
		for sk := range secure {
			if !known[sk] {
				t.Errorf("example %q uses unknown secure key %q", key, sk)
			}
		}
		if key != "noAuth" && len(secure) == 0 {
			t.Errorf("example %q has an empty secureJsonData", key)
		}
	}
}
