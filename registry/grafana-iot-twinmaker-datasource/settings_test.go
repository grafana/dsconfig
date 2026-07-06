package iottwinmakerdatasource

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
		name            string
		example         string
		settings        backend.DataSourceInstanceSettings
		wantErr         error
		wantAuthType    AWSAuthType
		wantDefaultReg  string
		wantWorkspaceID string
		wantAssumeARN   string
		wantWriteARN    string
		wantExternalID  string
		wantAccessKey   string
		wantSecretKey   string
		wantSecureKeys  SecureJsonDataConfig
	}{
		{
			// The default example intentionally leaves workspaceId and
			// assumeRoleArn blank so LoadConfig's Validate step rejects it.
			name:    "default example fails validation (missing workspaceId + assumeRoleArn)",
			example: "",
			wantErr: errors.New("workspaceId is required"),
		},
		{
			name:            "aws sdk default",
			example:         "awsSdkDefault",
			wantAuthType:    AWSAuthTypeDefault,
			wantDefaultReg:  "us-east-1",
			wantWorkspaceID: exampleWorkspaceID,
			wantAssumeARN:   exampleAssumeRoleARN,
		},
		{
			name:            "access and secret key",
			example:         "accessAndSecretKey",
			wantAuthType:    AWSAuthTypeKeys,
			wantDefaultReg:  "us-east-1",
			wantWorkspaceID: exampleWorkspaceID,
			wantAssumeARN:   exampleAssumeRoleARN,
			wantAccessKey:   "AKIAIOSFODNN7EXAMPLE",
			wantSecretKey:   "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
			wantSecureKeys:  SecureJsonDataConfig{SecureJsonDataKeyAccessKey, SecureJsonDataKeySecretKey},
		},
		{
			name:            "credentials file",
			example:         "credentialsFile",
			wantAuthType:    AWSAuthTypeCredentials,
			wantDefaultReg:  "us-east-1",
			wantWorkspaceID: exampleWorkspaceID,
			wantAssumeARN:   exampleAssumeRoleARN,
		},
		{
			name:            "workspace iam role",
			example:         "workspaceIamRole",
			wantAuthType:    AWSAuthTypeEC2IAMRole,
			wantDefaultReg:  "us-east-1",
			wantWorkspaceID: exampleWorkspaceID,
			wantAssumeARN:   exampleAssumeRoleARN,
		},
		{
			name:            "external id set",
			example:         "withExternalId",
			wantAuthType:    AWSAuthTypeDefault,
			wantDefaultReg:  "us-east-1",
			wantWorkspaceID: exampleWorkspaceID,
			wantAssumeARN:   exampleAssumeRoleARN,
			wantExternalID:  "external-id-abc123",
		},
		{
			name:            "with alarm write role",
			example:         "withAlarmWriteRole",
			wantAuthType:    AWSAuthTypeDefault,
			wantDefaultReg:  "us-east-1",
			wantWorkspaceID: exampleWorkspaceID,
			wantAssumeARN:   exampleAssumeRoleARN,
			wantWriteARN:    exampleWriteRoleARN,
		},
		{
			name:            "legacy arn auth type",
			example:         "legacyArnAuthType",
			wantAuthType:    AWSAuthTypeARN,
			wantDefaultReg:  "us-east-1",
			wantWorkspaceID: exampleWorkspaceID,
			wantAssumeARN:   exampleAssumeRoleARN,
		},
		{
			// Upstream Load only unmarshals when len(JSONData) > 1
			// (pkg/models/settings.go:21); a two-byte broken payload must
			// still error out with a parse error.
			name: "malformed jsonData errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"`),
			},
			wantErr: errors.New("parse jsonData"),
		},
		{
			name: "assumeRoleARN pascal case decodes",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{
                    "authType":"default",
                    "defaultRegion":"us-east-1",
                    "workspaceId":"WS",
                    "assumeRoleARN":"arn:aws:iam::123456789012:role/Legacy"
                }`),
			},
			wantAuthType:    AWSAuthTypeDefault,
			wantDefaultReg:  "us-east-1",
			wantWorkspaceID: "WS",
			wantAssumeARN:   "arn:aws:iam::123456789012:role/Legacy",
		},
		{
			// Confirms sessionToken is not copied into
			// DecryptedSecureJSONData by LoadConfig, matching the plugin's
			// own Load (pkg/models/settings.go:37-38).
			name: "session token in decrypted data is dropped",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{
                    "authType":"keys",
                    "defaultRegion":"us-east-1",
                    "workspaceId":"WS",
                    "assumeRoleArn":"arn:aws:iam::123456789012:role/Dash"
                }`),
				DecryptedSecureJSONData: map[string]string{
					"accessKey":    "AKID",
					"secretKey":    "SECRET",
					"sessionToken": "STS-TOKEN",
				},
			},
			wantAuthType:    AWSAuthTypeKeys,
			wantDefaultReg:  "us-east-1",
			wantWorkspaceID: "WS",
			wantAssumeARN:   "arn:aws:iam::123456789012:role/Dash",
			wantAccessKey:   "AKID",
			wantSecretKey:   "SECRET",
			// Only accessKey + secretKey are in the wantSecureKeys list —
			// sessionToken must NOT appear.
			wantSecureKeys: SecureJsonDataConfig{
				SecureJsonDataKeyAccessKey,
				SecureJsonDataKeySecretKey,
			},
		},
		{
			// Empty settings: ApplyDefaults fills authType=default and
			// defaultRegion=us-east-1, but Validate rejects the config
			// because workspaceId is missing.
			name:     "empty settings default and then fail validation",
			settings: backend.DataSourceInstanceSettings{},
			wantErr:  errors.New("workspaceId is required"),
		},
		{
			name: "unknown auth type errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"authType":"nonsense","workspaceId":"WS","assumeRoleArn":"arn:aws:iam::123456789012:role/R"}`),
			},
			wantErr: errors.New(`unknown authType "nonsense"`),
		},
		{
			name: "keys auth without accessKey errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"authType":"keys","workspaceId":"WS","assumeRoleArn":"arn:aws:iam::123456789012:role/R"}`),
			},
			wantErr: errors.New("accessKey is required for keys auth"),
		},
		{
			name: "keys auth with only accessKey still errors on secretKey",
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"authType":"keys","workspaceId":"WS","assumeRoleArn":"arn:aws:iam::123456789012:role/R"}`),
				DecryptedSecureJSONData: map[string]string{"accessKey": "AKID"},
			},
			wantErr: errors.New("secretKey is required for keys auth"),
		},
		{
			name: "missing workspaceId errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"authType":"default","assumeRoleArn":"arn:aws:iam::123456789012:role/R"}`),
			},
			wantErr: errors.New("workspaceId is required"),
		},
		{
			name: "missing assumeRoleArn errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"authType":"default","workspaceId":"WS"}`),
			},
			wantErr: errors.New("assumeRoleArn is required"),
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
			if tt.wantWorkspaceID != "" && cfg.WorkspaceID != tt.wantWorkspaceID {
				t.Errorf("WorkspaceID = %q, want %q", cfg.WorkspaceID, tt.wantWorkspaceID)
			}
			if tt.wantAssumeARN != "" && cfg.AssumeRoleARN != tt.wantAssumeARN {
				t.Errorf("AssumeRoleARN = %q, want %q", cfg.AssumeRoleARN, tt.wantAssumeARN)
			}
			if tt.wantWriteARN != "" && cfg.AssumeRoleARNWriter != tt.wantWriteARN {
				t.Errorf("AssumeRoleARNWriter = %q, want %q", cfg.AssumeRoleARNWriter, tt.wantWriteARN)
			}
			if tt.wantExternalID != "" && cfg.ExternalID != tt.wantExternalID {
				t.Errorf("ExternalID = %q, want %q", cfg.ExternalID, tt.wantExternalID)
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
		})
	}
}

func TestApplyDefaults(t *testing.T) {
	tests := []struct {
		name           string
		in             Config
		wantAuth       AWSAuthType
		wantDefaultReg string
	}{
		{
			name:           "empty config gets default auth type and region",
			in:             Config{},
			wantAuth:       AWSAuthTypeDefault,
			wantDefaultReg: "us-east-1",
		},
		{
			name:           "existing auth type is preserved",
			in:             Config{AuthType: AWSAuthTypeKeys},
			wantAuth:       AWSAuthTypeKeys,
			wantDefaultReg: "us-east-1",
		},
		{
			name:           "existing region is preserved",
			in:             Config{DefaultRegion: "eu-west-1"},
			wantAuth:       AWSAuthTypeDefault,
			wantDefaultReg: "eu-west-1",
		},
		{
			name:           "legacy arn is preserved",
			in:             Config{AuthType: AWSAuthTypeARN},
			wantAuth:       AWSAuthTypeARN,
			wantDefaultReg: "us-east-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.in
			got.ApplyDefaults()
			if got.AuthType != tt.wantAuth {
				t.Errorf("AuthType = %q, want %q", got.AuthType, tt.wantAuth)
			}
			if got.DefaultRegion != tt.wantDefaultReg {
				t.Errorf("DefaultRegion = %q, want %q", got.DefaultRegion, tt.wantDefaultReg)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	basic := func() Config {
		return Config{
			AuthType:                AWSAuthTypeDefault,
			DefaultRegion:           "us-east-1",
			WorkspaceID:             "ws",
			AssumeRoleARN:           "arn:aws:iam::123456789012:role/Dash",
			DecryptedSecureJSONData: map[SecureJsonDataKey]string{},
		}
	}
	withSecret := func(key SecureJsonDataKey, val string) func(*Config) {
		return func(c *Config) { c.DecryptedSecureJSONData[key] = val }
	}

	tests := []struct {
		name    string
		mutate  func(*Config)
		wantErr string
	}{
		{
			name:   "default auth happy path",
			mutate: func(c *Config) {},
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
			name: "grafana assume role happy path",
			mutate: func(c *Config) {
				c.AuthType = AWSAuthTypeGrafanaAssumeRole
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
			wantErr: "accessKey is required for keys auth",
		},
		{
			name: "keys without secretKey errors",
			mutate: func(c *Config) {
				c.AuthType = AWSAuthTypeKeys
				withSecret(SecureJsonDataKeyAccessKey, "AKID")(c)
			},
			wantErr: "secretKey is required for keys auth",
		},
		{
			name: "empty auth type errors as unknown",
			mutate: func(c *Config) {
				c.AuthType = ""
			},
			wantErr: `unknown authType ""`,
		},
		{
			name: "unknown auth type errors",
			mutate: func(c *Config) {
				c.AuthType = "totally-not-real"
			},
			wantErr: `unknown authType "totally-not-real"`,
		},
		{
			name: "missing workspaceId errors",
			mutate: func(c *Config) {
				c.WorkspaceID = ""
			},
			wantErr: "workspaceId is required",
		},
		{
			name: "missing assumeRoleArn errors",
			mutate: func(c *Config) {
				c.AssumeRoleARN = ""
			},
			wantErr: "assumeRoleArn is required",
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
