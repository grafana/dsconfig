package odbcdatasource

import (
	"encoding/json"
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
		name        string
		example     string
		useExample  bool
		settings    backend.DataSourceInstanceSettings
		wantErr     string
		wantDriver  string
		wantDSN     string
		wantTimeout string
		wantSecrets map[SecureJsonDataKey]string
	}{
		{
			// The default schema example has no driver, so LoadConfig's Validate
			// step is expected to reject it.
			name:       "default example fails validation (no driver)",
			example:    "",
			useExample: true,
			wantErr:    "driver is required",
		},
		{
			name:        "driver path DB2 example",
			example:     "driverPathDB2",
			useExample:  true,
			wantDriver:  "/opt/db2/clidriver/lib/libdb2.so.1",
			wantTimeout: "10",
			wantSecrets: map[SecureJsonDataKey]string{SecureJsonDataKeyPwd: "<your-password>"},
		},
		{
			name:        "driver alias example",
			example:     "driverAlias",
			useExample:  true,
			wantDriver:  "{MySQLDB}",
			wantTimeout: "10",
			wantSecrets: map[SecureJsonDataKey]string{SecureJsonDataKeyPwd: "<your-password>"},
		},
		{
			name:        "backend-only DSN example",
			example:     "connectionStringDSN",
			useExample:  true,
			wantDriver:  "{TESTDB}",
			wantDSN:     "TESTDB",
			wantTimeout: "10",
			wantSecrets: map[SecureJsonDataKey]string{SecureJsonDataKeyPwd: "<your-password>"},
		},
		{
			name: "lowercase editor keys parse",
			settings: backend.DataSourceInstanceSettings{
				JSONData: json.RawMessage(`{"driver":"{MyDSN}","timeout":"30","settings":[{"name":"uid","value":"grafana","secure":false}]}`),
			},
			wantDriver:  "{MyDSN}",
			wantTimeout: "30",
		},
		{
			name: "capitalized upstream keys still parse (case-insensitive)",
			settings: backend.DataSourceInstanceSettings{
				JSONData:                json.RawMessage(`{"Driver":"mysql","DSN":"{MySQLDB}","Timeout":"30","Settings":[{"Name":"pwd","Value":"","Secure":true}]}`),
				DecryptedSecureJSONData: map[string]string{"pwd": "s3cr3t"},
			},
			wantDriver:  "mysql",
			wantDSN:     "{MySQLDB}",
			wantTimeout: "30",
			wantSecrets: map[SecureJsonDataKey]string{SecureJsonDataKeyPwd: "s3cr3t"},
		},
		{
			name: "secure setting missing secret errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData:                json.RawMessage(`{"driver":"{MyDSN}","settings":[{"name":"pwd","secure":true}]}`),
				DecryptedSecureJSONData: map[string]string{},
			},
			wantErr: "Missing pwd",
		},
		{
			name: "dynamic secret name other than pwd resolves",
			settings: backend.DataSourceInstanceSettings{
				JSONData:                json.RawMessage(`{"driver":"{MyDSN}","settings":[{"name":"AccessToken","secure":true}]}`),
				DecryptedSecureJSONData: map[string]string{"AccessToken": "tok"},
			},
			wantDriver:  "{MyDSN}",
			wantTimeout: "10",
			wantSecrets: map[SecureJsonDataKey]string{"AccessToken": "tok"},
		},
		{
			name: "empty timeout defaults to 10",
			settings: backend.DataSourceInstanceSettings{
				JSONData: json.RawMessage(`{"driver":"{MyDSN}"}`),
			},
			wantDriver:  "{MyDSN}",
			wantTimeout: "10",
		},
		{
			name: "invalid jsonData errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: json.RawMessage(`{`),
			},
			wantErr: "parse jsonData",
		},
		{
			name: "missing driver errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: json.RawMessage(`{"timeout":"10"}`),
			},
			wantErr: "driver is required",
		},
		{
			name: "non-integer timeout errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: json.RawMessage(`{"driver":"{MyDSN}","timeout":"soon"}`),
			},
			wantErr: "timeout must be an integer",
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

			if tt.wantDriver != "" && cfg.Driver != tt.wantDriver {
				t.Errorf("Driver = %q, want %q", cfg.Driver, tt.wantDriver)
			}
			if tt.wantDSN != "" && cfg.DSN != tt.wantDSN {
				t.Errorf("DSN = %q, want %q", cfg.DSN, tt.wantDSN)
			}
			if tt.wantTimeout != "" && cfg.Timeout != tt.wantTimeout {
				t.Errorf("Timeout = %q, want %q", cfg.Timeout, tt.wantTimeout)
			}
			for k, want := range tt.wantSecrets {
				if got := cfg.DecryptedSecureJSONData[k]; got != want {
					t.Errorf("DecryptedSecureJSONData[%s] = %q, want %q", k, got, want)
				}
			}
		})
	}
}

func TestLoadConfigResolvesSecureSettingValue(t *testing.T) {
	settings := backend.DataSourceInstanceSettings{
		JSONData:                json.RawMessage(`{"driver":"{MyDSN}","settings":[{"name":"pwd","secure":true}]}`),
		DecryptedSecureJSONData: map[string]string{"pwd": "s3cr3t"},
	}
	cfg, err := LoadConfig(t.Context(), settings)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if len(cfg.Settings) != 1 {
		t.Fatalf("Settings len = %d, want 1", len(cfg.Settings))
	}
	if cfg.Settings[0].Value != "s3cr3t" {
		t.Errorf("secure setting Value = %q, want %q (should be resolved from secureJsonData)", cfg.Settings[0].Value, "s3cr3t")
	}
}

func TestApplyDefaults(t *testing.T) {
	tests := []struct {
		name        string
		in          Config
		wantTimeout string
	}{
		{"empty timeout gets 10", Config{}, "10"},
		{"existing timeout preserved", Config{Timeout: "45"}, "45"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.in
			got.ApplyDefaults()
			if got.Timeout != tt.wantTimeout {
				t.Errorf("Timeout = %q, want %q", got.Timeout, tt.wantTimeout)
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
			name: "driver path ok",
			cfg:  Config{Driver: "/opt/db2/clidriver/lib/libdb2.so.1", Timeout: "10"},
		},
		{
			name: "driver alias ok",
			cfg:  Config{Driver: "{MySQLDB}", Timeout: "10"},
		},
		{
			name:    "missing driver errors",
			cfg:     Config{Timeout: "10"},
			wantErr: "driver is required",
		},
		{
			name:    "non-integer timeout errors",
			cfg:     Config{Driver: "{MyDSN}", Timeout: "soon"},
			wantErr: "timeout must be an integer",
		},
		{
			name: "empty timeout is allowed (defaulted elsewhere)",
			cfg:  Config{Driver: "{MyDSN}"},
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
