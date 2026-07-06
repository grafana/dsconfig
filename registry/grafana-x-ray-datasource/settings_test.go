package xraydatasource

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
		name              string
		example           string // schema.go SettingsExamples key ("" = use inline settings)
		settings          backend.DataSourceInstanceSettings
		wantErr           error
		wantAuthType      AWSAuthType
		wantDefaultReg    string
		wantProfile       string
		wantAssumeARN     string
		wantExternalID    string
		wantDatabase      string
		wantEffectiveProf string
		wantSecureKeys    SecureJsonDataConfig
		wantAccessKey     string
		wantSecretKey     string
		wantSessionToken  string
	}{
		{
			// The default schema example intentionally leaves defaultRegion
			// blank so LoadConfig's Validate step rejects it.
			name:    "default example fails validation (missing defaultRegion)",
			example: "",
			wantErr: errors.New("defaultRegion is required"),
		},
		{
			name:           "aws sdk default",
			example:        "awsSdkDefault",
			wantAuthType:   AWSAuthTypeDefault,
			wantDefaultReg: "us-east-1",
		},
		{
			name:           "access and secret key",
			example:        "accessAndSecretKey",
			wantAuthType:   AWSAuthTypeKeys,
			wantDefaultReg: "us-east-1",
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyAccessKey, SecureJsonDataKeySecretKey},
			wantAccessKey:  "AKIAIOSFODNN7EXAMPLE",
			wantSecretKey:  "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		},
		{
			name:              "credentials file",
			example:           "credentialsFile",
			wantAuthType:      AWSAuthTypeCredentials,
			wantDefaultReg:    "us-east-1",
			wantProfile:       "my-xray-profile",
			wantEffectiveProf: "my-xray-profile",
		},
		{
			name:           "workspace iam role",
			example:        "workspaceIamRole",
			wantAuthType:   AWSAuthTypeEC2IAMRole,
			wantDefaultReg: "us-east-1",
		},
		{
			name:           "grafana assume role",
			example:        "grafanaAssumeRole",
			wantAuthType:   AWSAuthTypeGrafanaAssumeRole,
			wantDefaultReg: "us-east-1",
		},
		{
			name:           "keys with sts assume role",
			example:        "assumeRoleFromKeys",
			wantAuthType:   AWSAuthTypeKeys,
			wantDefaultReg: "us-east-1",
			wantAssumeARN:  "arn:aws:iam::123456789012:role/GrafanaXRay",
			wantExternalID: "external-id-abc123",
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyAccessKey, SecureJsonDataKeySecretKey},
			wantAccessKey:  "AKIAIOSFODNN7EXAMPLE",
			wantSecretKey:  "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		},
		{
			name:           "legacy arn auth type",
			example:        "legacyArnAuthType",
			wantAuthType:   AWSAuthTypeARN,
			wantDefaultReg: "us-east-1",
		},
		{
			name: "malformed jsonData errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{`),
			},
			wantErr: errors.New("parse jsonData"),
		},
		{
			name: "assumeRoleARN pascal case decodes",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{
                    "authType":"default",
                    "defaultRegion":"us-east-1",
                    "assumeRoleARN":"arn:aws:iam::123456789012:role/Legacy"
                }`),
			},
			wantAuthType:   AWSAuthTypeDefault,
			wantDefaultReg: "us-east-1",
			wantAssumeARN:  "arn:aws:iam::123456789012:role/Legacy",
		},
		{
			name: "legacy database becomes effective profile when profile empty",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{
                    "authType":"credentials",
                    "defaultRegion":"us-east-1"
                }`),
				Database: "legacy-profile",
			},
			wantAuthType:      AWSAuthTypeCredentials,
			wantDefaultReg:    "us-east-1",
			wantDatabase:      "legacy-profile",
			wantEffectiveProf: "legacy-profile",
		},
		{
			name: "explicit profile wins over legacy database",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{
                    "authType":"credentials",
                    "profile":"explicit",
                    "defaultRegion":"us-east-1"
                }`),
				Database: "legacy-profile",
			},
			wantAuthType:      AWSAuthTypeCredentials,
			wantDefaultReg:    "us-east-1",
			wantProfile:       "explicit",
			wantDatabase:      "legacy-profile",
			wantEffectiveProf: "explicit",
		},
		{
			name: "session token loaded from secure data",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{
                    "authType":"keys",
                    "defaultRegion":"us-east-1"
                }`),
				DecryptedSecureJSONData: map[string]string{
					"accessKey":    "AKID",
					"secretKey":    "SECRET",
					"sessionToken": "STS-TOKEN",
				},
			},
			wantAuthType:     AWSAuthTypeKeys,
			wantDefaultReg:   "us-east-1",
			wantSecureKeys:   SecureJsonDataConfig{SecureJsonDataKeyAccessKey, SecureJsonDataKeySecretKey, SecureJsonDataKeySessionToken},
			wantAccessKey:    "AKID",
			wantSecretKey:    "SECRET",
			wantSessionToken: "STS-TOKEN",
		},
		{
			name: "empty settings default to AWS SDK default and fail validation",
			// After ApplyDefaults the auth type is "default", but
			// defaultRegion is empty, so Validate rejects the config.
			settings: backend.DataSourceInstanceSettings{},
			wantErr:  errors.New("defaultRegion is required"),
		},
		{
			name: "unknown auth type errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"authType":"nonsense","defaultRegion":"us-east-1"}`),
			},
			wantErr: errors.New(`unknown authType "nonsense"`),
		},
		{
			name: "keys auth without accessKey errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"authType":"keys","defaultRegion":"us-east-1"}`),
			},
			wantErr: errors.New("accessKey is required for keys auth"),
		},
		{
			name: "keys auth with only accessKey still errors on secretKey",
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"authType":"keys","defaultRegion":"us-east-1"}`),
				DecryptedSecureJSONData: map[string]string{"accessKey": "AKID"},
			},
			wantErr: errors.New("secretKey is required for keys auth"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings := tt.settings
			if tt.example != "" || (tt.settings.JSONData == nil && tt.settings.DecryptedSecureJSONData == nil && tt.settings.Database == "" && tt.wantErr == nil) {
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
			if tt.wantProfile != "" && cfg.Profile != tt.wantProfile {
				t.Errorf("Profile = %q, want %q", cfg.Profile, tt.wantProfile)
			}
			if tt.wantAssumeARN != "" && cfg.AssumeRoleARN != tt.wantAssumeARN {
				t.Errorf("AssumeRoleARN = %q, want %q", cfg.AssumeRoleARN, tt.wantAssumeARN)
			}
			if tt.wantExternalID != "" && cfg.ExternalID != tt.wantExternalID {
				t.Errorf("ExternalID = %q, want %q", cfg.ExternalID, tt.wantExternalID)
			}
			if tt.wantDatabase != "" && cfg.Database != tt.wantDatabase {
				t.Errorf("Database = %q, want %q", cfg.Database, tt.wantDatabase)
			}
			if tt.wantEffectiveProf != "" && cfg.EffectiveProfile() != tt.wantEffectiveProf {
				t.Errorf("EffectiveProfile() = %q, want %q", cfg.EffectiveProfile(), tt.wantEffectiveProf)
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
		name string
		in   Config
		want AWSAuthType
	}{
		{
			name: "empty config gets default auth type",
			in:   Config{},
			want: AWSAuthTypeDefault,
		},
		{
			name: "existing auth type is preserved",
			in:   Config{AuthType: AWSAuthTypeKeys},
			want: AWSAuthTypeKeys,
		},
		{
			name: "legacy arn is preserved",
			in:   Config{AuthType: AWSAuthTypeARN},
			want: AWSAuthTypeARN,
		},
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
			name: "credentials auth via legacy Database happy path",
			mutate: func(c *Config) {
				c.AuthType = AWSAuthTypeCredentials
				c.Database = "legacy-profile"
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
			name:    "empty auth type errors as unknown",
			mutate:  func(c *Config) {},
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
			name: "missing defaultRegion errors",
			mutate: func(c *Config) {
				c.AuthType = AWSAuthTypeDefault
				c.DefaultRegion = ""
			},
			wantErr: "defaultRegion is required",
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

func TestEffectiveProfile(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
		want string
	}{
		{name: "both empty", cfg: Config{}, want: ""},
		{name: "profile only", cfg: Config{Profile: "p"}, want: "p"},
		{name: "database only (legacy fallback)", cfg: Config{Database: "d"}, want: "d"},
		{name: "profile beats database", cfg: Config{Profile: "p", Database: "d"}, want: "p"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cfg.EffectiveProfile(); got != tt.want {
				t.Errorf("EffectiveProfile() = %q, want %q", got, tt.want)
			}
		})
	}
}
