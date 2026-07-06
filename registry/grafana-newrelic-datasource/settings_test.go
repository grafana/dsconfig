package newrelicdatasource

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
		name               string
		exampleKey         string // SettingsExamples key when useExample is true
		useExample         bool
		settings           backend.DataSourceInstanceSettings
		wantErr            error // substring match; nil = expect success
		wantRegion         Region
		wantTimeout        int64
		checkAccountID     bool
		wantAccountID      int
		wantSecureKeys     SecureJsonDataConfig
		wantPersonalAPIKey string
	}{
		{
			// The default schema example intentionally has empty secret
			// placeholders, so LoadConfig's Validate step rejects it.
			name:       "default example fails validation (empty secrets)",
			exampleKey: "",
			useExample: true,
			wantErr:    errors.New("personal API key (secureJsonData.personalApiKey) is required"),
		},
		{
			name:               "us region",
			exampleKey:         "usRegion",
			useExample:         true,
			wantRegion:         RegionUS,
			wantTimeout:        300, // omitted in the example -> ApplyDefaults fills 300
			checkAccountID:     true,
			wantAccountID:      1234567,
			wantSecureKeys:     SecureJsonDataConfig{SecureJsonDataKeyPersonalAPIKey, SecureJsonDataKeyAccountID},
			wantPersonalAPIKey: "<your-newrelic-api-key>",
		},
		{
			name:           "eu region with custom timeout",
			exampleKey:     "euRegion",
			useExample:     true,
			wantRegion:     RegionEU,
			wantTimeout:    600,
			checkAccountID: true,
			wantAccountID:  1234567,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyPersonalAPIKey, SecureJsonDataKeyAccountID},
		},
		{
			// Legacy shape: accountId lives in jsonData, not secureJsonData, so
			// the backend never reads it -> AccountID stays 0 -> Validate fails.
			name:       "legacy accountId in jsonData fails validation",
			exampleKey: "legacyAccountIdInJsonData",
			useExample: true,
			wantErr:    errors.New("account ID (secureJsonData.accountId) is required"),
		},
		{
			name: "invalid jsonData errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{`),
			},
			wantErr: errors.New("parse jsonData"),
		},
		{
			// Upstream LoadSettings unmarshals JSONData unconditionally, so
			// empty/nil JSONData is a parse error (pkg/models/settings.go:30-32).
			name:     "empty JSONData is a parse error",
			settings: backend.DataSourceInstanceSettings{},
			wantErr:  errors.New("parse jsonData"),
		},
		{
			name: "non-numeric accountId leaves AccountID zero and fails",
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{}`),
				DecryptedSecureJSONData: map[string]string{"personalApiKey": "key", "accountId": "not-a-number"},
			},
			wantErr: errors.New("account ID (secureJsonData.accountId) is required"),
		},
		{
			name: "whitespace personalApiKey fails",
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{}`),
				DecryptedSecureJSONData: map[string]string{"personalApiKey": "   ", "accountId": "5"},
			},
			wantErr: errors.New("personal API key (secureJsonData.personalApiKey) is required"),
		},
		{
			name: "timeout below one defaults to 300",
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"timeoutInSeconds":0}`),
				DecryptedSecureJSONData: map[string]string{"personalApiKey": "key", "accountId": "5"},
			},
			wantTimeout:    300,
			checkAccountID: true,
			wantAccountID:  5,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyPersonalAPIKey, SecureJsonDataKeyAccountID},
		},
		{
			name: "explicit timeout preserved",
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"region":"US","timeoutInSeconds":42}`),
				DecryptedSecureJSONData: map[string]string{"personalApiKey": "key", "accountId": "5"},
			},
			wantRegion:  RegionUS,
			wantTimeout: 42,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings := tt.settings
			if tt.useExample {
				settings = settingsFromExample(t, tt.exampleKey)
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

			if tt.wantRegion != "" && cfg.Region != tt.wantRegion {
				t.Errorf("Region = %q, want %q", cfg.Region, tt.wantRegion)
			}
			if tt.wantTimeout != 0 && cfg.TimeoutInSeconds != tt.wantTimeout {
				t.Errorf("TimeoutInSeconds = %d, want %d", cfg.TimeoutInSeconds, tt.wantTimeout)
			}
			if tt.checkAccountID && cfg.AccountID != tt.wantAccountID {
				t.Errorf("AccountID = %d, want %d", cfg.AccountID, tt.wantAccountID)
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
			if tt.wantPersonalAPIKey != "" && cfg.DecryptedSecureJSONData[SecureJsonDataKeyPersonalAPIKey] != tt.wantPersonalAPIKey {
				t.Errorf("DecryptedSecureJSONData[personalApiKey] = %q, want %q",
					cfg.DecryptedSecureJSONData[SecureJsonDataKeyPersonalAPIKey], tt.wantPersonalAPIKey)
			}
		})
	}
}

func TestApplyDefaults(t *testing.T) {
	tests := []struct {
		name        string
		in          Config
		wantTimeout int64
		wantRegion  Region
	}{
		{
			name:        "empty config gets 300s timeout",
			in:          Config{},
			wantTimeout: 300,
		},
		{
			name:        "zero timeout defaults to 300",
			in:          Config{TimeoutInSeconds: 0},
			wantTimeout: 300,
		},
		{
			name:        "negative timeout defaults to 300",
			in:          Config{TimeoutInSeconds: -5},
			wantTimeout: 300,
		},
		{
			name:        "existing timeout preserved",
			in:          Config{TimeoutInSeconds: 60},
			wantTimeout: 60,
		},
		{
			name:        "region is left untouched",
			in:          Config{Region: RegionEU},
			wantTimeout: 300,
			wantRegion:  RegionEU,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.in
			got.ApplyDefaults()
			if got.TimeoutInSeconds != tt.wantTimeout {
				t.Errorf("TimeoutInSeconds = %d, want %d", got.TimeoutInSeconds, tt.wantTimeout)
			}
			if got.Region != tt.wantRegion {
				t.Errorf("Region = %q, want %q", got.Region, tt.wantRegion)
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
			name: "valid config",
			cfg: Config{
				AccountID:               5,
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyPersonalAPIKey: "key"},
			},
		},
		{
			name:    "missing personalApiKey",
			cfg:     Config{AccountID: 5, DecryptedSecureJSONData: map[SecureJsonDataKey]string{}},
			wantErr: "personal API key (secureJsonData.personalApiKey) is required",
		},
		{
			name: "whitespace personalApiKey",
			cfg: Config{
				AccountID:               5,
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyPersonalAPIKey: "  "},
			},
			wantErr: "personal API key (secureJsonData.personalApiKey) is required",
		},
		{
			name: "zero account ID",
			cfg: Config{
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyPersonalAPIKey: "key"},
			},
			wantErr: "account ID (secureJsonData.accountId) is required",
		},
		{
			name:    "both missing",
			cfg:     Config{DecryptedSecureJSONData: map[SecureJsonDataKey]string{}},
			wantErr: "personal API key (secureJsonData.personalApiKey) is required",
		},
		{
			// Upstream CheckSettings only checks AccountID == 0
			// (pkg/datasource/handler_checkhealth.go:143), so a negative account
			// ID passes even though the message says "positive". Documented as
			// an upstream quirk in the README.
			name: "negative account ID passes (upstream only checks == 0)",
			cfg: Config{
				AccountID:               -5,
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyPersonalAPIKey: "key"},
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
