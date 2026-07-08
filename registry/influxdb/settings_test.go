package influxdbdatasource

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
	if s, ok := value["user"].(string); ok {
		settings.User = s
	}
	if s, ok := value["database"].(string); ok {
		settings.Database = s
	}
	return settings
}

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name           string
		example        string // schema.go SettingsExamples key ("" default example is loaded when settings and wantErr are unset)
		settings       backend.DataSourceInstanceSettings
		wantErr        error
		wantURL        string
		wantVersion    InfluxVersion
		wantHTTPMode   InfluxHTTPMode
		wantDbName     string
		wantBasicAuth  bool
		wantBasicUser  string
		wantUser       string
		wantOAuth      bool
		wantTLSAuth    bool
		wantTLSCA      bool
		wantMaxSeries  int32
		wantSecureKeys SecureJsonDataConfig
	}{
		{
			// The default example is a minimal InfluxQL config with a dbName set.
			name:          "default example loads",
			example:       "",
			wantURL:       "http://localhost:8086",
			wantVersion:   InfluxVersionInfluxQL,
			wantHTTPMode:  InfluxHTTPModeGET,
			wantDbName:    "mydb",
			wantMaxSeries: DefaultMaxSeries,
		},
		{
			name:           "influxql basic auth",
			example:        "influxqlBasicAuth",
			wantURL:        "https://influxdb.example.com:8086",
			wantVersion:    InfluxVersionInfluxQL,
			wantHTTPMode:   InfluxHTTPModePOST,
			wantDbName:     "telegraf",
			wantBasicAuth:  true,
			wantBasicUser:  "grafana",
			wantMaxSeries:  DefaultMaxSeries,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyBasicAuthPassword},
		},
		{
			name:           "influxql legacy user/password",
			example:        "influxqlLegacyUserPassword",
			wantURL:        "http://influxdb.example.com:8086",
			wantVersion:    InfluxVersionInfluxQL,
			wantHTTPMode:   InfluxHTTPModeGET,
			wantDbName:     "telegraf",
			wantUser:       "admin",
			wantMaxSeries:  DefaultMaxSeries,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyPassword},
		},
		{
			name:           "flux with token",
			example:        "fluxToken",
			wantURL:        "https://us-west-2-1.aws.cloud2.influxdata.com",
			wantVersion:    InfluxVersionFlux,
			wantHTTPMode:   InfluxHTTPModePOST,
			wantMaxSeries:  DefaultMaxSeries,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyToken},
		},
		{
			name:           "sql flightsql",
			example:        "sqlFlightSQL",
			wantURL:        "https://us-east-1-1.aws.cloud2.influxdata.com",
			wantVersion:    InfluxVersionSQL,
			wantDbName:     "metrics",
			wantMaxSeries:  DefaultMaxSeries,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyToken},
		},
		{
			name:           "tls mutual auth",
			example:        "tlsMutualAuth",
			wantURL:        "https://influxdb.example.com:8086",
			wantVersion:    InfluxVersionInfluxQL,
			wantDbName:     "telegraf",
			wantTLSAuth:    true,
			wantMaxSeries:  DefaultMaxSeries,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyTLSClientCert, SecureJsonDataKeyTLSClientKey},
		},
		{
			name:           "tls self-signed CA",
			example:        "tlsSelfSignedCA",
			wantURL:        "https://influxdb.internal.corp",
			wantVersion:    InfluxVersionInfluxQL,
			wantDbName:     "telegraf",
			wantTLSCA:      true,
			wantMaxSeries:  DefaultMaxSeries,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyTLSCACert},
		},
		{
			name:          "oauth forward",
			example:       "oauthForward",
			wantURL:       "https://influxdb.example.com",
			wantVersion:   InfluxVersionInfluxQL,
			wantDbName:    "telegraf",
			wantOAuth:     true,
			wantMaxSeries: DefaultMaxSeries,
		},
		{
			name:          "legacy root.database fallback",
			example:       "legacyRootDatabase",
			wantURL:       "http://influxdb.example.com:8086",
			wantVersion:   InfluxVersionInfluxQL,
			wantDbName:    "legacy_db",
			wantMaxSeries: DefaultMaxSeries,
		},
		{
			name: "missing URL errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"version":"InfluxQL","dbName":"telegraf"}`),
			},
			wantErr: errors.New("InfluxDB URL"),
		},
		{
			name: "invalid version errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://localhost:8086",
				JSONData: []byte(`{"version":"NotAValidLanguage"}`),
			},
			wantErr: errors.New(`invalid version "NotAValidLanguage"`),
		},
		{
			name: "invalid httpMode errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://localhost:8086",
				JSONData: []byte(`{"version":"InfluxQL","dbName":"t","httpMode":"PUT"}`),
			},
			wantErr: errors.New(`invalid httpMode "PUT"`),
		},
		{
			name: "influxql without dbName errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://localhost:8086",
				JSONData: []byte(`{"version":"InfluxQL"}`),
			},
			wantErr: errors.New("dbName (jsonData) is required when version is InfluxQL"),
		},
		{
			name: "flux without organization errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://localhost:8086",
				JSONData: []byte(`{"version":"Flux","defaultBucket":"b"}`),
				DecryptedSecureJSONData: map[string]string{
					"token": "t",
				},
			},
			wantErr: errors.New("organization (jsonData) is required when version is Flux"),
		},
		{
			name: "flux without defaultBucket errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://localhost:8086",
				JSONData: []byte(`{"version":"Flux","organization":"o"}`),
				DecryptedSecureJSONData: map[string]string{
					"token": "t",
				},
			},
			wantErr: errors.New("defaultBucket (jsonData) is required when version is Flux"),
		},
		{
			name: "flux without token errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://localhost:8086",
				JSONData: []byte(`{"version":"Flux","organization":"o","defaultBucket":"b"}`),
			},
			wantErr: errors.New("token (secureJsonData) is required when version is Flux"),
		},
		{
			name: "sql without dbName errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://localhost:8086",
				JSONData: []byte(`{"version":"SQL"}`),
				DecryptedSecureJSONData: map[string]string{
					"token": "t",
				},
			},
			wantErr: errors.New("dbName (jsonData) is required when version is SQL"),
		},
		{
			name: "sql without token errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://localhost:8086",
				JSONData: []byte(`{"version":"SQL","dbName":"metrics"}`),
			},
			wantErr: errors.New("token (secureJsonData) is required when version is SQL"),
		},
		{
			name: "basicAuth without user errors",
			settings: backend.DataSourceInstanceSettings{
				URL:              "http://localhost:8086",
				BasicAuthEnabled: true,
				JSONData:         []byte(`{"version":"InfluxQL","dbName":"telegraf"}`),
			},
			wantErr: errors.New("basicAuthUser (root) is required"),
		},
		{
			name: "tlsAuth without serverName errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://influxdb",
				JSONData: []byte(`{"version":"InfluxQL","dbName":"t","tlsAuth":true}`),
				DecryptedSecureJSONData: map[string]string{
					"tlsClientCert": "pem",
					"tlsClientKey":  "pem",
				},
			},
			wantErr: errors.New("serverName (jsonData) is required"),
		},
		{
			name: "tlsAuth without client cert errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://influxdb",
				JSONData: []byte(`{"version":"InfluxQL","dbName":"t","tlsAuth":true,"serverName":"influxdb"}`),
			},
			wantErr: errors.New("tlsClientCert (secureJsonData) is required"),
		},
		{
			name: "tlsAuthWithCACert without CA cert errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://influxdb",
				JSONData: []byte(`{"version":"InfluxQL","dbName":"t","tlsAuthWithCACert":true}`),
			},
			wantErr: errors.New("tlsCACert (secureJsonData) is required"),
		},
		{
			name: "invalid jsonData errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://localhost:8086",
				JSONData: []byte(`{`),
			},
			wantErr: errors.New("parse jsonData"),
		},
		{
			name: "negative timeout errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://localhost:8086",
				JSONData: []byte(`{"version":"InfluxQL","dbName":"t","timeout":-5}`),
			},
			wantErr: errors.New("timeout must be non-negative"),
		},
		{
			name: "negative maxSeries errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://localhost:8086",
				JSONData: []byte(`{"version":"InfluxQL","dbName":"t","maxSeries":-1}`),
			},
			wantErr: errors.New("maxSeries must be non-negative"),
		},
		{
			name: "empty jsonData is defaulted (influxql)",
			settings: backend.DataSourceInstanceSettings{
				URL:      "http://localhost:8086",
				Database: "telegraf",
				JSONData: []byte(`{}`),
			},
			wantURL:       "http://localhost:8086",
			wantVersion:   DefaultInfluxVersion,
			wantHTTPMode:  DefaultInfluxHTTPMode,
			wantDbName:    "telegraf",
			wantMaxSeries: DefaultMaxSeries,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings := tt.settings
			if tt.example != "" || (tt.settings.JSONData == nil && tt.settings.URL == "" && tt.wantErr == nil) {
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
			if tt.wantVersion != "" && cfg.Version != tt.wantVersion {
				t.Errorf("Version = %q, want %q", cfg.Version, tt.wantVersion)
			}
			if tt.wantHTTPMode != "" && cfg.HTTPMode != tt.wantHTTPMode {
				t.Errorf("HTTPMode = %q, want %q", cfg.HTTPMode, tt.wantHTTPMode)
			}
			if tt.wantDbName != "" && cfg.DbName != tt.wantDbName {
				t.Errorf("DbName = %q, want %q", cfg.DbName, tt.wantDbName)
			}
			if cfg.BasicAuth != tt.wantBasicAuth {
				t.Errorf("BasicAuth = %v, want %v", cfg.BasicAuth, tt.wantBasicAuth)
			}
			if tt.wantBasicUser != "" && cfg.BasicAuthUser != tt.wantBasicUser {
				t.Errorf("BasicAuthUser = %q, want %q", cfg.BasicAuthUser, tt.wantBasicUser)
			}
			if tt.wantUser != "" && cfg.User != tt.wantUser {
				t.Errorf("User = %q, want %q", cfg.User, tt.wantUser)
			}
			if cfg.OauthPassThru != tt.wantOAuth {
				t.Errorf("OauthPassThru = %v, want %v", cfg.OauthPassThru, tt.wantOAuth)
			}
			if cfg.TLSAuth != tt.wantTLSAuth {
				t.Errorf("TLSAuth = %v, want %v", cfg.TLSAuth, tt.wantTLSAuth)
			}
			if cfg.TLSAuthWithCACert != tt.wantTLSCA {
				t.Errorf("TLSAuthWithCACert = %v, want %v", cfg.TLSAuthWithCACert, tt.wantTLSCA)
			}
			if tt.wantMaxSeries != 0 && cfg.MaxSeries != tt.wantMaxSeries {
				t.Errorf("MaxSeries = %d, want %d", cfg.MaxSeries, tt.wantMaxSeries)
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
			name: "empty version / httpMode / maxSeries are all defaulted",
			in:   Config{URL: "http://localhost:8086"},
			want: Config{
				URL:       "http://localhost:8086",
				Version:   DefaultInfluxVersion,
				HTTPMode:  DefaultInfluxHTTPMode,
				MaxSeries: DefaultMaxSeries,
			},
		},
		{
			name: "explicit Flux version is preserved",
			in:   Config{URL: "http://localhost:8086", Version: InfluxVersionFlux},
			want: Config{
				URL:       "http://localhost:8086",
				Version:   InfluxVersionFlux,
				HTTPMode:  DefaultInfluxHTTPMode,
				MaxSeries: DefaultMaxSeries,
			},
		},
		{
			name: "explicit POST httpMode is preserved",
			in:   Config{URL: "http://localhost:8086", HTTPMode: InfluxHTTPModePOST},
			want: Config{
				URL:       "http://localhost:8086",
				Version:   DefaultInfluxVersion,
				HTTPMode:  InfluxHTTPModePOST,
				MaxSeries: DefaultMaxSeries,
			},
		},
		{
			name: "explicit maxSeries is preserved",
			in:   Config{URL: "http://localhost:8086", MaxSeries: 5000},
			want: Config{
				URL:       "http://localhost:8086",
				Version:   DefaultInfluxVersion,
				HTTPMode:  DefaultInfluxHTTPMode,
				MaxSeries: 5000,
			},
		},
		{
			name: "empty dbName falls back to root.database",
			in:   Config{URL: "http://localhost:8086", Database: "legacy"},
			want: Config{
				URL:       "http://localhost:8086",
				Database:  "legacy",
				Version:   DefaultInfluxVersion,
				HTTPMode:  DefaultInfluxHTTPMode,
				DbName:    "legacy",
				MaxSeries: DefaultMaxSeries,
			},
		},
		{
			name: "explicit dbName wins over root.database",
			in:   Config{URL: "http://localhost:8086", Database: "legacy", DbName: "new"},
			want: Config{
				URL:       "http://localhost:8086",
				Database:  "legacy",
				DbName:    "new",
				Version:   DefaultInfluxVersion,
				HTTPMode:  DefaultInfluxHTTPMode,
				MaxSeries: DefaultMaxSeries,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.in
			got.ApplyDefaults()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ApplyDefaults: got %#v, want %#v", got, tt.want)
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
			name: "minimal influxql happy path",
			cfg: Config{
				URL:      "http://localhost:8086",
				Version:  InfluxVersionInfluxQL,
				HTTPMode: InfluxHTTPModeGET,
				DbName:   "telegraf",
			},
		},
		{
			name:    "missing URL",
			cfg:     Config{Version: InfluxVersionInfluxQL, DbName: "t"},
			wantErr: "InfluxDB URL (root.url) is required",
		},
		{
			name: "invalid version",
			cfg: Config{
				URL:     "http://localhost:8086",
				Version: InfluxVersion("Bogus"),
				DbName:  "t",
			},
			wantErr: `invalid version "Bogus"`,
		},
		{
			name: "invalid httpMode",
			cfg: Config{
				URL:      "http://localhost:8086",
				Version:  InfluxVersionInfluxQL,
				HTTPMode: "DELETE",
				DbName:   "t",
			},
			wantErr: `invalid httpMode "DELETE"`,
		},
		{
			name: "flux with everything present",
			cfg: Config{
				URL:           "http://localhost:8086",
				Version:       InfluxVersionFlux,
				Organization:  "o",
				DefaultBucket: "b",
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{
					SecureJsonDataKeyToken: "t",
				},
			},
		},
		{
			name: "sql needs dbName + token",
			cfg: Config{
				URL:     "http://localhost:8086",
				Version: InfluxVersionSQL,
			},
			wantErr: "dbName (jsonData) is required when version is SQL",
		},
		{
			name: "basicAuth needs user",
			cfg: Config{
				URL:       "http://localhost:8086",
				Version:   InfluxVersionInfluxQL,
				DbName:    "t",
				BasicAuth: true,
			},
			wantErr: "basicAuthUser (root) is required",
		},
		{
			name: "tlsAuth needs serverName + client cert + client key",
			cfg: Config{
				URL:     "https://influxdb",
				Version: InfluxVersionInfluxQL,
				DbName:  "t",
				TLSAuth: true,
			},
			wantErr: "serverName (jsonData) is required",
		},
		{
			name: "tlsAuth with everything present",
			cfg: Config{
				URL:        "https://influxdb",
				Version:    InfluxVersionInfluxQL,
				DbName:     "t",
				TLSAuth:    true,
				ServerName: "influxdb",
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{
					SecureJsonDataKeyTLSClientCert: "pem",
					SecureJsonDataKeyTLSClientKey:  "pem",
				},
			},
		},
		{
			name: "tlsAuthWithCACert without CA errors",
			cfg: Config{
				URL:               "https://influxdb",
				Version:           InfluxVersionInfluxQL,
				DbName:            "t",
				TLSAuthWithCACert: true,
			},
			wantErr: "tlsCACert (secureJsonData) is required",
		},
		{
			name: "negative timeout errors",
			cfg: Config{
				URL:     "http://localhost:8086",
				Version: InfluxVersionInfluxQL,
				DbName:  "t",
				Timeout: -1,
			},
			wantErr: "timeout must be non-negative",
		},
		{
			name: "negative maxSeries errors",
			cfg: Config{
				URL:       "http://localhost:8086",
				Version:   InfluxVersionInfluxQL,
				DbName:    "t",
				MaxSeries: -1,
			},
			wantErr: "maxSeries must be non-negative",
		},
		{
			name: "influxql with root.database fallback (no jsonData.dbName)",
			cfg: Config{
				URL:      "http://localhost:8086",
				Version:  InfluxVersionInfluxQL,
				Database: "legacy",
			},
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
