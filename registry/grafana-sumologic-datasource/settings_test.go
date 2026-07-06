package sumologicdatasource

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
		example        string
		settings       backend.DataSourceInstanceSettings
		useSettings    bool
		wantErr        error
		wantAuthMethod AuthenticationMethod
		wantApiURL     string
		wantAccessID   string
		wantTimeout    int
		wantInterval   int
		wantSecure     SecureJsonDataConfig
		wantAccessKey  string
	}{
		{
			// The default example intentionally has an empty accessKey
			// placeholder and no accessId, so LoadConfig's Validate step
			// rejects it.
			name:    "default example fails validation (empty access credentials)",
			example: "",
			wantErr: errors.New("invalid access id"),
		},
		{
			name:           "access key US1 default region",
			example:        "accessKey",
			wantAuthMethod: AuthenticationMethodAccessKey,
			wantApiURL:     "https://api.sumologic.com/api/",
			wantAccessID:   "<your-access-id>",
			wantTimeout:    30,
			wantInterval:   1000,
			wantSecure:     SecureJsonDataConfig{SecureJsonDataKeyAccessKey},
			wantAccessKey:  "<your-access-key>",
		},
		{
			name:           "access key EU region",
			example:        "accessKeyEU",
			wantAuthMethod: AuthenticationMethodAccessKey,
			wantApiURL:     "https://api.eu.sumologic.com/api/",
			wantAccessID:   "<your-access-id>",
			wantSecure:     SecureJsonDataConfig{SecureJsonDataKeyAccessKey},
			wantAccessKey:  "<your-access-key>",
		},
		{
			name:           "legacy datasource without authMethod defaults to accessKey",
			example:        "legacyNoAuthMethod",
			wantAuthMethod: AuthenticationMethodAccessKey,
			wantApiURL:     "https://api.sumologic.com/api/",
			wantAccessID:   "<your-access-id>",
			wantSecure:     SecureJsonDataConfig{SecureJsonDataKeyAccessKey},
			wantAccessKey:  "<your-access-key>",
		},
		{
			name:        "custom timeout and interval are preserved",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"authMethod":"accessKey","apiUrl":"https://api.au.sumologic.com/api/","accessId":"id","timeout":15,"interval":500}`),
				DecryptedSecureJSONData: map[string]string{"accessKey": "key"},
			},
			wantAuthMethod: AuthenticationMethodAccessKey,
			wantApiURL:     "https://api.au.sumologic.com/api/",
			wantAccessID:   "id",
			wantTimeout:    15,
			wantInterval:   500,
			wantSecure:     SecureJsonDataConfig{SecureJsonDataKeyAccessKey},
			wantAccessKey:  "key",
		},
		{
			name:        "empty settings apply defaults and fail validation",
			useSettings: true,
			settings:    backend.DataSourceInstanceSettings{},
			wantErr:     errors.New("invalid access id"),
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
			name:        "access key auth missing access key secret errors",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"authMethod":"accessKey","accessId":"id"}`),
			},
			wantErr: errors.New("invalid access key"),
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

			if tt.wantAuthMethod != "" && cfg.AuthenticationMethod != tt.wantAuthMethod {
				t.Errorf("AuthenticationMethod = %q, want %q", cfg.AuthenticationMethod, tt.wantAuthMethod)
			}
			if tt.wantApiURL != "" && cfg.ApiURL != tt.wantApiURL {
				t.Errorf("ApiURL = %q, want %q", cfg.ApiURL, tt.wantApiURL)
			}
			if tt.wantAccessID != "" && cfg.AccessID != tt.wantAccessID {
				t.Errorf("AccessID = %q, want %q", cfg.AccessID, tt.wantAccessID)
			}
			if tt.wantTimeout != 0 && cfg.Timeout != tt.wantTimeout {
				t.Errorf("Timeout = %d, want %d", cfg.Timeout, tt.wantTimeout)
			}
			if tt.wantInterval != 0 && cfg.Interval != tt.wantInterval {
				t.Errorf("Interval = %d, want %d", cfg.Interval, tt.wantInterval)
			}
			if tt.wantSecure != nil {
				gotKeys := SecureJsonDataConfig{}
				for _, key := range SecureJsonDataKeys {
					if _, ok := cfg.DecryptedSecureJSONData[key]; ok {
						gotKeys = append(gotKeys, key)
					}
				}
				if !reflect.DeepEqual(gotKeys, tt.wantSecure) {
					t.Errorf("configured secure keys = %v, want %v", gotKeys, tt.wantSecure)
				}
			}
			if tt.wantAccessKey != "" && cfg.DecryptedSecureJSONData[SecureJsonDataKeyAccessKey] != tt.wantAccessKey {
				t.Errorf("DecryptedSecureJSONData[accessKey] = %q, want %q", cfg.DecryptedSecureJSONData[SecureJsonDataKeyAccessKey], tt.wantAccessKey)
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
			name: "empty config gets accessKey + url + timeout + interval",
			in:   Config{},
			want: Config{AuthenticationMethod: AuthenticationMethodAccessKey, ApiURL: DefaultApiURL, Timeout: DefaultTimeout, Interval: DefaultInterval},
		},
		{
			name: "explicit values are preserved",
			in:   Config{AuthenticationMethod: AuthenticationMethodAccessKey, ApiURL: "https://api.eu.sumologic.com/api/", Timeout: 15, Interval: 500},
			want: Config{AuthenticationMethod: AuthenticationMethodAccessKey, ApiURL: "https://api.eu.sumologic.com/api/", Timeout: 15, Interval: 500},
		},
		{
			name: "only timeout set gets other defaults",
			in:   Config{Timeout: 5},
			want: Config{AuthenticationMethod: AuthenticationMethodAccessKey, ApiURL: DefaultApiURL, Timeout: 5, Interval: DefaultInterval},
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
			name: "access key happy path",
			cfg: Config{
				AuthenticationMethod:    AuthenticationMethodAccessKey,
				ApiURL:                  DefaultApiURL,
				AccessID:                "id",
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyAccessKey: "key"},
			},
		},
		{
			name: "missing apiUrl errors",
			cfg: Config{
				AuthenticationMethod:    AuthenticationMethodAccessKey,
				AccessID:                "id",
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyAccessKey: "key"},
			},
			wantErr: "invalid API URL",
		},
		{
			name:    "empty config errors on url and auth method",
			cfg:     Config{},
			wantErr: "invalid authentication method",
		},
		{
			name: "access key auth missing access id errors",
			cfg: Config{
				AuthenticationMethod:    AuthenticationMethodAccessKey,
				ApiURL:                  DefaultApiURL,
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyAccessKey: "key"},
			},
			wantErr: "invalid access id",
		},
		{
			name: "access key auth missing access key errors",
			cfg: Config{
				AuthenticationMethod: AuthenticationMethodAccessKey,
				ApiURL:               DefaultApiURL,
				AccessID:             "id",
			},
			wantErr: "invalid access key",
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
