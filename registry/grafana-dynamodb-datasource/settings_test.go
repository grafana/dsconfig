package dynamodbdatasource

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// settingsFromExample converts a SettingsExamples entry (a full
// instance settings object with jsonData and secureJsonData) into the
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
		name             string
		example          string // schema.go SettingsExamples key ("" = use inline settings)
		settings         backend.DataSourceInstanceSettings
		wantErr          error
		wantAuthType     AWSAuthType
		wantDefaultReg   string
		wantEndpoint     string
		wantProfile      string
		wantIsV2         bool
		wantTimeout      string
		wantRetries      string
		wantPause        string
		wantLegacyReg    string
		wantLegacyAccess string
		wantSecureKeys   SecureJsonDataConfig
		wantAccessKey    string
		wantSecretKey    string
		wantSessionToken string
	}{
		{
			// The default schema example intentionally leaves
			// defaultRegion blank so LoadConfig's Validate step rejects
			// it.
			name:    "default example fails validation (missing region)",
			example: "",
			wantErr: errors.New("missing region"),
		},
		{
			name:           "aws sdk default",
			example:        "awsSdkDefault",
			wantAuthType:   AWSAuthTypeDefault,
			wantDefaultReg: "us-east-1",
			wantIsV2:       true,
			wantTimeout:    "60",
			wantRetries:    "5",
			wantPause:      "5",
		},
		{
			name:           "access and secret key",
			example:        "accessAndSecretKey",
			wantAuthType:   AWSAuthTypeKeys,
			wantDefaultReg: "us-east-1",
			wantEndpoint:   "https://dynamodb.us-east-1.amazonaws.com",
			wantIsV2:       true,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyAccessKey, SecureJsonDataKeySecretKey},
			wantAccessKey:  "AKIAIOSFODNN7EXAMPLE",
			wantSecretKey:  "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		},
		{
			name:           "credentials file",
			example:        "credentialsFile",
			wantAuthType:   AWSAuthTypeCredentials,
			wantDefaultReg: "us-east-1",
			wantProfile:    "my-dynamodb-profile",
			wantIsV2:       true,
		},
		{
			name:           "workspace iam role",
			example:        "workspaceIamRole",
			wantAuthType:   AWSAuthTypeEC2IAMRole,
			wantDefaultReg: "us-east-1",
			wantIsV2:       true,
		},
		{
			name:             "keys with session token",
			example:          "keysWithSessionToken",
			wantAuthType:     AWSAuthTypeKeys,
			wantDefaultReg:   "us-east-1",
			wantIsV2:         true,
			wantSecureKeys:   SecureJsonDataConfig{SecureJsonDataKeyAccessKey, SecureJsonDataKeySecretKey, SecureJsonDataKeySessionToken},
			wantAccessKey:    "AKIAIOSFODNN7EXAMPLE",
			wantSecretKey:    "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
			wantSessionToken: "FQoGZXIvYXdzEExampleSTSSessionToken",
		},
		{
			name:           "driver settings overrides",
			example:        "driverSettings",
			wantAuthType:   AWSAuthTypeDefault,
			wantDefaultReg: "us-east-1",
			wantIsV2:       true,
			wantTimeout:    "120",
			wantRetries:    "10",
			wantPause:      "2",
		},
		{
			name:           "legacy arn auth type",
			example:        "legacyArnAuthType",
			wantAuthType:   AWSAuthTypeARN,
			wantDefaultReg: "us-east-1",
			wantIsV2:       true,
		},
		{
			// Mirrors pkg/models/settings_test.go's testV1Migration.
			// Under V1 storage the ID is in jsonData.accessId and the
			// SECRET is in secureJsonData.accessKey. LoadConfig
			// migrates the effective view: AuthType → keys,
			// DefaultRegion ← region, accessKey ← accessId,
			// secretKey ← the V1 accessKey value.
			name:             "legacy v1 shape from example",
			example:          "legacyV1Shape",
			wantAuthType:     AWSAuthTypeKeys,
			wantDefaultReg:   "eu-north-1",
			wantEndpoint:     "https://dynamodb.eu-north-1.amazonaws.com",
			wantIsV2:         false,
			wantLegacyReg:    "eu-north-1",
			wantLegacyAccess: "AKIAIOSFODNN7EXAMPLE",
			wantSecureKeys:   SecureJsonDataConfig{SecureJsonDataKeyAccessKey, SecureJsonDataKeySecretKey},
			wantAccessKey:    "AKIAIOSFODNN7EXAMPLE",
			wantSecretKey:    "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		},
		{
			// Inline restatement of the upstream testV1Migration
			// (pkg/models/settings_test.go:26-43): raw V1 payload,
			// only accessKey in secureJsonData.
			name: "v1 migration inline (parity with upstream test)",
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"region":"eu-north-1","endpoint":"https://dynamodb.eu-north-1.amazonaws.com","accessId":"DYNAMODB-ACCESS-KEY"}`),
				DecryptedSecureJSONData: map[string]string{"accessKey": "DYNAMODB-ACCESS-SECRET-KEY"},
			},
			wantAuthType:     AWSAuthTypeKeys,
			wantDefaultReg:   "eu-north-1",
			wantEndpoint:     "https://dynamodb.eu-north-1.amazonaws.com",
			wantIsV2:         false,
			wantLegacyReg:    "eu-north-1",
			wantLegacyAccess: "DYNAMODB-ACCESS-KEY",
			wantAccessKey:    "DYNAMODB-ACCESS-KEY",
			wantSecretKey:    "DYNAMODB-ACCESS-SECRET-KEY",
		},
		{
			// Mirrors testV1MigrationWithSessionToken.
			name: "v1 migration with session token",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"region":"eu-north-1","endpoint":"https://dynamodb.eu-north-1.amazonaws.com","accessId":"DYNAMODB-ACCESS-KEY"}`),
				DecryptedSecureJSONData: map[string]string{
					"accessKey":    "DYNAMODB-ACCESS-SECRET-KEY",
					"sessionToken": "AQoDYXdzEJr//test-session-token",
				},
			},
			wantAuthType:     AWSAuthTypeKeys,
			wantDefaultReg:   "eu-north-1",
			wantEndpoint:     "https://dynamodb.eu-north-1.amazonaws.com",
			wantIsV2:         false,
			wantLegacyReg:    "eu-north-1",
			wantLegacyAccess: "DYNAMODB-ACCESS-KEY",
			wantAccessKey:    "DYNAMODB-ACCESS-KEY",
			wantSecretKey:    "DYNAMODB-ACCESS-SECRET-KEY",
			wantSessionToken: "AQoDYXdzEJr//test-session-token",
		},
		{
			// Mirrors testV2 and testV2WithSessionToken: modern shape
			// straight through, no migration.
			name: "v2 inline (parity with upstream test)",
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"isV2":true,"authType":"keys","defaultRegion":"eu-north-1","endpoint":"https://dynamodb.eu-north-1.amazonaws.com"}`),
				DecryptedSecureJSONData: map[string]string{"accessKey": "DYNAMODB-ACCESS-KEY", "secretKey": "DYNAMODB-ACCESS-SECRET-KEY"},
			},
			wantAuthType:   AWSAuthTypeKeys,
			wantDefaultReg: "eu-north-1",
			wantEndpoint:   "https://dynamodb.eu-north-1.amazonaws.com",
			wantIsV2:       true,
			wantAccessKey:  "DYNAMODB-ACCESS-KEY",
			wantSecretKey:  "DYNAMODB-ACCESS-SECRET-KEY",
		},
		{
			name: "malformed jsonData errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{`),
			},
			wantErr: errors.New("parse jsonData"),
		},
		{
			name: "empty settings default to AWS SDK default and fail validation",
			// After ApplyDefaults the auth type is "default", but
			// defaultRegion is empty (and isV2 is false so V1
			// migration flips authType to keys — but the region is
			// still missing).
			settings: backend.DataSourceInstanceSettings{},
			wantErr:  errors.New("missing region"),
		},
		{
			name: "unknown auth type errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"authType":"nonsense","isV2":true,"defaultRegion":"us-east-1"}`),
			},
			wantErr: errors.New(`unknown authType "nonsense"`),
		},
		{
			name: "grafana_assume_role rejected (not supported by dynamodb)",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"authType":"grafana_assume_role","isV2":true,"defaultRegion":"us-east-1"}`),
			},
			wantErr: errors.New(`unknown authType "grafana_assume_role"`),
		},
		{
			name: "keys auth without accessKey errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"authType":"keys","isV2":true,"defaultRegion":"us-east-1"}`),
			},
			wantErr: errors.New("missing access key"),
		},
		{
			name: "keys auth with only accessKey still errors on secretKey",
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"authType":"keys","isV2":true,"defaultRegion":"us-east-1"}`),
				DecryptedSecureJSONData: map[string]string{"accessKey": "AKID"},
			},
			wantErr: errors.New("missing secret key"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings := tt.settings
			if tt.example != "" || (tt.settings.JSONData == nil && tt.settings.DecryptedSecureJSONData == nil && tt.wantErr == nil) {
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

			if tt.wantAuthType != "" && cfg.AuthType != tt.wantAuthType {
				t.Errorf("AuthType = %q, want %q", cfg.AuthType, tt.wantAuthType)
			}
			if tt.wantDefaultReg != "" && cfg.DefaultRegion != tt.wantDefaultReg {
				t.Errorf("DefaultRegion = %q, want %q", cfg.DefaultRegion, tt.wantDefaultReg)
			}
			if tt.wantEndpoint != "" && cfg.Endpoint != tt.wantEndpoint {
				t.Errorf("Endpoint = %q, want %q", cfg.Endpoint, tt.wantEndpoint)
			}
			if tt.wantProfile != "" && cfg.Profile != tt.wantProfile {
				t.Errorf("Profile = %q, want %q", cfg.Profile, tt.wantProfile)
			}
			if cfg.IsV2 != tt.wantIsV2 {
				t.Errorf("IsV2 = %v, want %v", cfg.IsV2, tt.wantIsV2)
			}
			if tt.wantTimeout != "" && cfg.Timeout != tt.wantTimeout {
				t.Errorf("Timeout = %q, want %q", cfg.Timeout, tt.wantTimeout)
			}
			if tt.wantRetries != "" && cfg.Retries != tt.wantRetries {
				t.Errorf("Retries = %q, want %q", cfg.Retries, tt.wantRetries)
			}
			if tt.wantPause != "" && cfg.Pause != tt.wantPause {
				t.Errorf("Pause = %q, want %q", cfg.Pause, tt.wantPause)
			}
			if tt.wantLegacyReg != "" && cfg.LegacyRegion != tt.wantLegacyReg {
				t.Errorf("LegacyRegion = %q, want %q", cfg.LegacyRegion, tt.wantLegacyReg)
			}
			if tt.wantLegacyAccess != "" && cfg.LegacyAccessKey != tt.wantLegacyAccess {
				t.Errorf("LegacyAccessKey = %q, want %q", cfg.LegacyAccessKey, tt.wantLegacyAccess)
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
			if tt.wantAccessKey != "" && cfg.DecryptedSecureJSONData[SecureJsonDataKeyAccessKey] != tt.wantAccessKey {
				t.Errorf("Secrets[accessKey] = %q, want %q", cfg.DecryptedSecureJSONData[SecureJsonDataKeyAccessKey], tt.wantAccessKey)
			}
			if tt.wantSecretKey != "" && cfg.DecryptedSecureJSONData[SecureJsonDataKeySecretKey] != tt.wantSecretKey {
				t.Errorf("Secrets[secretKey] = %q, want %q", cfg.DecryptedSecureJSONData[SecureJsonDataKeySecretKey], tt.wantSecretKey)
			}
			if tt.wantSessionToken != "" && cfg.DecryptedSecureJSONData[SecureJsonDataKeySessionToken] != tt.wantSessionToken {
				t.Errorf("Secrets[sessionToken] = %q, want %q", cfg.DecryptedSecureJSONData[SecureJsonDataKeySessionToken], tt.wantSessionToken)
			}
		})
	}
}

func TestApplyDefaults(t *testing.T) {
	tests := []struct {
		name        string
		in          Config
		wantAuth    AWSAuthType
		wantTimeout string
		wantRetries string
		wantPause   string
	}{
		{
			name:        "empty config gets default auth type and driver defaults",
			in:          Config{},
			wantAuth:    AWSAuthTypeDefault,
			wantTimeout: "60",
			wantRetries: "5",
			wantPause:   "5",
		},
		{
			name:        "existing auth type is preserved",
			in:          Config{AuthType: AWSAuthTypeKeys},
			wantAuth:    AWSAuthTypeKeys,
			wantTimeout: "60",
			wantRetries: "5",
			wantPause:   "5",
		},
		{
			name:        "legacy arn is preserved",
			in:          Config{AuthType: AWSAuthTypeARN},
			wantAuth:    AWSAuthTypeARN,
			wantTimeout: "60",
			wantRetries: "5",
			wantPause:   "5",
		},
		{
			name:        "custom driver settings are preserved",
			in:          Config{AuthType: AWSAuthTypeDefault, Timeout: "120", Retries: "10", Pause: "2"},
			wantAuth:    AWSAuthTypeDefault,
			wantTimeout: "120",
			wantRetries: "10",
			wantPause:   "2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.in
			got.ApplyDefaults()
			if got.AuthType != tt.wantAuth {
				t.Errorf("AuthType = %q, want %q", got.AuthType, tt.wantAuth)
			}
			if got.Timeout != tt.wantTimeout {
				t.Errorf("Timeout = %q, want %q", got.Timeout, tt.wantTimeout)
			}
			if got.Retries != tt.wantRetries {
				t.Errorf("Retries = %q, want %q", got.Retries, tt.wantRetries)
			}
			if got.Pause != tt.wantPause {
				t.Errorf("Pause = %q, want %q", got.Pause, tt.wantPause)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	// Every happy-path case starts from a config with a region set;
	// each case adjusts the auth-related fields to exercise its
	// scenario.
	basic := func() Config {
		return Config{
			DefaultRegion:           "us-east-1",
			DecryptedSecureJSONData: map[SecureJsonDataKey]string{},
		}
	}
	withSecret := func(key SecureJsonDataKey, val string) func(*Config) {
		return func(c *Config) { c.DecryptedSecureJSONData[key] = val }
	}

	tests := []struct {
		name    string
		mutate  func(*Config)
		wantErr string // empty = expect no error; otherwise substring match
	}{
		{
			name: "default auth happy path",
			mutate: func(c *Config) {
				c.AuthType = AWSAuthTypeDefault
			},
		},
		{
			name: "keys auth happy path",
			mutate: func(c *Config) {
				c.AuthType = AWSAuthTypeKeys
				withSecret(SecureJsonDataKeyAccessKey, "AKID")(c)
				withSecret(SecureJsonDataKeySecretKey, "SECRET")(c)
			},
		},
		{
			name: "credentials auth happy path",
			mutate: func(c *Config) {
				c.AuthType = AWSAuthTypeCredentials
				c.Profile = "my-profile"
			},
		},
		{
			name: "ec2 iam role happy path",
			mutate: func(c *Config) {
				c.AuthType = AWSAuthTypeEC2IAMRole
			},
		},
		{
			name: "legacy arn happy path",
			mutate: func(c *Config) {
				c.AuthType = AWSAuthTypeARN
			},
		},
		{
			name: "keys without accessKey errors",
			mutate: func(c *Config) {
				c.AuthType = AWSAuthTypeKeys
				withSecret(SecureJsonDataKeySecretKey, "SECRET")(c)
			},
			wantErr: "missing access key",
		},
		{
			name: "keys without secretKey errors",
			mutate: func(c *Config) {
				c.AuthType = AWSAuthTypeKeys
				withSecret(SecureJsonDataKeyAccessKey, "AKID")(c)
			},
			wantErr: "missing secret key",
		},
		{
			name:    "empty auth type errors as unknown",
			mutate:  func(c *Config) {},
			wantErr: `unknown authType ""`,
		},
		{
			name: "grafana_assume_role errors (not supported)",
			mutate: func(c *Config) {
				c.AuthType = "grafana_assume_role"
			},
			wantErr: `unknown authType "grafana_assume_role"`,
		},
		{
			name: "unknown auth type errors",
			mutate: func(c *Config) {
				c.AuthType = "totally-not-real"
			},
			wantErr: `unknown authType "totally-not-real"`,
		},
		{
			name: "missing defaultRegion errors",
			mutate: func(c *Config) {
				c.AuthType = AWSAuthTypeDefault
				c.DefaultRegion = ""
			},
			wantErr: "missing region",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := basic()
			tt.mutate(&cfg)
			err := cfg.Validate()
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
