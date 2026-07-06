package dynatracedatasource

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"errors"
	"math/big"
	"reflect"
	"strings"
	"testing"
	"time"

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

// testCACertPEM generates a valid self-signed certificate PEM at runtime so the
// tests exercise the real x509.AppendCertsFromPEM path in Validate without
// committing a certificate blob to the repository.
func testCACertPEM(t *testing.T) string {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "dynatrace-test-ca"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create certificate: %v", err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}))
}

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name              string
		example           string // schema.go SettingsExamples key
		useSettings       bool
		settings          backend.DataSourceInstanceSettings
		wantErr           error
		wantAPIType       APIType
		wantEnvID         string
		wantDomain        string
		wantTimeout       int
		wantSecureKeys    SecureJsonDataConfig
		wantAPIToken      string
		wantPlatformToken string
	}{
		{
			// The default schema example has no environmentId and an empty
			// apiToken placeholder, so LoadConfig's Validate step rejects it.
			name:    "default example fails validation (empty environmentId and token)",
			example: "",
			wantErr: errors.New("environment ID"),
		},
		{
			name:           "saas api token",
			example:        "saasApiToken",
			wantAPIType:    APITypeSaaS,
			wantEnvID:      "abc12345",
			wantTimeout:    DefaultHTTPClientTimeout,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyAPIToken},
			wantAPIToken:   "<your-dynatrace-api-token>",
		},
		{
			name:              "saas platform token",
			example:           "saasPlatformToken",
			wantAPIType:       APITypeSaaS,
			wantEnvID:         "abc12345",
			wantSecureKeys:    SecureJsonDataConfig{SecureJsonDataKeyPlatformToken},
			wantPlatformToken: "<your-dynatrace-platform-token>",
		},
		{
			name:              "saas both tokens",
			example:           "saasBothTokens",
			wantAPIType:       APITypeSaaS,
			wantEnvID:         "abc12345",
			wantSecureKeys:    SecureJsonDataConfig{SecureJsonDataKeyAPIToken, SecureJsonDataKeyPlatformToken},
			wantAPIToken:      "<your-dynatrace-api-token>",
			wantPlatformToken: "<your-dynatrace-platform-token>",
		},
		{
			name:           "managed",
			example:        "managed",
			wantAPIType:    APITypeManaged,
			wantEnvID:      "abc12345",
			wantDomain:     "dynatrace.example.com",
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyAPIToken},
			wantAPIToken:   "<your-dynatrace-api-token>",
		},
		{
			name:           "raw url",
			example:        "rawUrl",
			wantAPIType:    APITypeURL,
			wantEnvID:      "https://abc12345.live.dynatrace.com",
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyAPIToken},
			wantAPIToken:   "<your-dynatrace-api-token>",
		},
		{
			// The CA cert example uses an obviously-fake PEM placeholder, so the
			// real x509 parse in Validate rejects it (documents the placeholder
			// limitation, like the empty-token default example above).
			name:    "tls ca cert example fails validation (placeholder PEM)",
			example: "tlsCACert",
			wantErr: errors.New("failed to parse TLS CA PEM certificate"),
		},
		{
			// Empty JSONData is a parse error upstream — LoadSettings
			// unconditionally json.Unmarshal(nil, &settings) which fails.
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
			name:        "missing environmentId errors",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"apiType":"saas"}`),
				DecryptedSecureJSONData: map[string]string{"apiToken": "tok"},
			},
			wantErr: errors.New("environment ID"),
		},
		{
			name:        "managed without domain errors",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"apiType":"managed","environmentId":"abc12345"}`),
				DecryptedSecureJSONData: map[string]string{"apiToken": "tok"},
			},
			wantErr: errors.New("domain"),
		},
		{
			name:        "no tokens errors",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"apiType":"saas","environmentId":"abc12345"}`),
			},
			wantErr: errors.New("API token"),
		},
		{
			name:        "empty apiType defaults to saas",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"environmentId":"abc12345"}`),
				DecryptedSecureJSONData: map[string]string{"apiToken": "tok"},
			},
			wantAPIType:    APITypeSaaS,
			wantEnvID:      "abc12345",
			wantTimeout:    DefaultHTTPClientTimeout,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyAPIToken},
			wantAPIToken:   "tok",
		},
		{
			name:        "explicit timeout preserved",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"apiType":"saas","environmentId":"abc12345","httpClientTimeout":260}`),
				DecryptedSecureJSONData: map[string]string{"platformToken": "tok"},
			},
			wantAPIType:       APITypeSaaS,
			wantEnvID:         "abc12345",
			wantTimeout:       260,
			wantSecureKeys:    SecureJsonDataConfig{SecureJsonDataKeyPlatformToken},
			wantPlatformToken: "tok",
		},
		{
			// enableSecureSocksProxy is intentionally omitted from both the
			// dsconfig schema and the Go Config struct; json unmarshal silently
			// ignores unknown fields.
			name:        "unknown enableSecureSocksProxy field is ignored",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"apiType":"saas","environmentId":"abc12345","enableSecureSocksProxy":true}`),
				DecryptedSecureJSONData: map[string]string{"apiToken": "tok"},
			},
			wantAPIType:  APITypeSaaS,
			wantEnvID:    "abc12345",
			wantAPIToken: "tok",
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

			if tt.wantAPIType != "" && cfg.APIType != tt.wantAPIType {
				t.Errorf("APIType = %q, want %q", cfg.APIType, tt.wantAPIType)
			}
			if tt.wantEnvID != "" && cfg.EnvironmentID != tt.wantEnvID {
				t.Errorf("EnvironmentID = %q, want %q", cfg.EnvironmentID, tt.wantEnvID)
			}
			if tt.wantDomain != "" && cfg.Domain != tt.wantDomain {
				t.Errorf("Domain = %q, want %q", cfg.Domain, tt.wantDomain)
			}
			if tt.wantTimeout != 0 && cfg.HTTPClientTimeout != tt.wantTimeout {
				t.Errorf("HTTPClientTimeout = %d, want %d", cfg.HTTPClientTimeout, tt.wantTimeout)
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
			if tt.wantAPIToken != "" && cfg.DecryptedSecureJSONData[SecureJsonDataKeyAPIToken] != tt.wantAPIToken {
				t.Errorf("DecryptedSecureJSONData[apiToken] = %q, want %q", cfg.DecryptedSecureJSONData[SecureJsonDataKeyAPIToken], tt.wantAPIToken)
			}
			if tt.wantPlatformToken != "" && cfg.DecryptedSecureJSONData[SecureJsonDataKeyPlatformToken] != tt.wantPlatformToken {
				t.Errorf("DecryptedSecureJSONData[platformToken] = %q, want %q", cfg.DecryptedSecureJSONData[SecureJsonDataKeyPlatformToken], tt.wantPlatformToken)
			}
		})
	}
}

// TestLoadConfigValidCACert exercises the happy path where tlsAuthWithCACert is
// enabled together with a genuinely valid PEM CA certificate.
func TestLoadConfigValidCACert(t *testing.T) {
	settings := backend.DataSourceInstanceSettings{
		JSONData: []byte(`{"apiType":"saas","environmentId":"abc12345","tlsAuthWithCACert":true}`),
		DecryptedSecureJSONData: map[string]string{
			"apiToken":  "tok",
			"tlsCACert": testCACertPEM(t),
		},
	}
	cfg, err := LoadConfig(t.Context(), settings)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if !cfg.TLSAuthWithCACert {
		t.Errorf("TLSAuthWithCACert = false, want true")
	}
	if _, ok := cfg.DecryptedSecureJSONData[SecureJsonDataKeyTLSCACert]; !ok {
		t.Errorf("expected tlsCACert secret to be loaded")
	}
}

func TestApplyDefaults(t *testing.T) {
	tests := []struct {
		name        string
		in          Config
		wantAPIType APIType
		wantTimeout int
	}{
		{
			name:        "empty config gets saas + timeout 30",
			in:          Config{},
			wantAPIType: APITypeSaaS,
			wantTimeout: DefaultHTTPClientTimeout,
		},
		{
			name:        "existing api type preserved, zero timeout defaulted",
			in:          Config{APIType: APITypeManaged},
			wantAPIType: APITypeManaged,
			wantTimeout: DefaultHTTPClientTimeout,
		},
		{
			name:        "negative timeout defaulted",
			in:          Config{APIType: APITypeURL, HTTPClientTimeout: -260},
			wantAPIType: APITypeURL,
			wantTimeout: DefaultHTTPClientTimeout,
		},
		{
			name:        "explicit values untouched",
			in:          Config{APIType: APITypeSaaS, HTTPClientTimeout: 260},
			wantAPIType: APITypeSaaS,
			wantTimeout: 260,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.in
			got.ApplyDefaults()
			if got.APIType != tt.wantAPIType {
				t.Errorf("APIType = %q, want %q", got.APIType, tt.wantAPIType)
			}
			if got.HTTPClientTimeout != tt.wantTimeout {
				t.Errorf("HTTPClientTimeout = %d, want %d", got.HTTPClientTimeout, tt.wantTimeout)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	validCert := testCACertPEM(t)
	tests := []struct {
		name    string
		cfg     Config
		wantErr string // empty = expect no error; otherwise substring match
	}{
		{
			name: "saas with api token",
			cfg: Config{
				APIType:                 APITypeSaaS,
				EnvironmentID:           "abc12345",
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyAPIToken: "tok"},
			},
		},
		{
			name: "saas with platform token only",
			cfg: Config{
				APIType:                 APITypeSaaS,
				EnvironmentID:           "abc12345",
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyPlatformToken: "tok"},
			},
		},
		{
			name: "managed with domain and token",
			cfg: Config{
				APIType:                 APITypeManaged,
				EnvironmentID:           "abc12345",
				Domain:                  "dynatrace.example.com",
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyAPIToken: "tok"},
			},
		},
		{
			name: "valid ca cert",
			cfg: Config{
				APIType:           APITypeSaaS,
				EnvironmentID:     "abc12345",
				TLSAuthWithCACert: true,
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{
					SecureJsonDataKeyAPIToken:  "tok",
					SecureJsonDataKeyTLSCACert: validCert,
				},
			},
		},
		{
			name:    "missing environment id errors",
			cfg:     Config{APIType: APITypeSaaS, DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyAPIToken: "tok"}},
			wantErr: "environment ID",
		},
		{
			name: "managed without domain errors",
			cfg: Config{
				APIType:                 APITypeManaged,
				EnvironmentID:           "abc12345",
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyAPIToken: "tok"},
			},
			wantErr: "domain",
		},
		{
			name:    "no tokens errors",
			cfg:     Config{APIType: APITypeSaaS, EnvironmentID: "abc12345"},
			wantErr: "API token",
		},
		{
			name: "ca cert enabled but missing errors",
			cfg: Config{
				APIType:                 APITypeSaaS,
				EnvironmentID:           "abc12345",
				TLSAuthWithCACert:       true,
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyAPIToken: "tok"},
			},
			wantErr: "TLS CA certificate",
		},
		{
			name: "ca cert unparseable errors",
			cfg: Config{
				APIType:           APITypeSaaS,
				EnvironmentID:     "abc12345",
				TLSAuthWithCACert: true,
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{
					SecureJsonDataKeyAPIToken:  "tok",
					SecureJsonDataKeyTLSCACert: "hello",
				},
			},
			wantErr: "failed to parse TLS CA PEM certificate",
		},
		{
			name:    "everything empty joins all errors",
			cfg:     Config{},
			wantErr: "environment ID",
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
