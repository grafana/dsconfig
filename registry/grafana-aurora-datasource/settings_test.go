package auroradatasource

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
		name             string
		example          string // schema.go SettingsExamples key ("" = use inline settings)
		settings         backend.DataSourceInstanceSettings
		wantErr          error
		wantAuthType     AWSAuthType
		wantEngine       AuroraEngine
		wantDefaultReg   string
		wantDBHost       string
		wantDBPort       int
		wantDBUser       string
		wantDBName       string
		wantDBHostAuth   string
		wantDBPortAuth   int
		wantAssumeARN    string
		wantExternalID   string
		wantSecureKeys   SecureJsonDataConfig
		wantAccessKey    string
		wantSecretKey    string
		wantSessionToken string
	}{
		{
			// The default schema example intentionally leaves the Aurora
			// selectors blank so LoadConfig's Validate step rejects it.
			name:    "default example fails validation (missing selectors)",
			example: "",
			wantErr: errors.New("defaultRegion is required"),
		},
		{
			name:           "aws sdk default with postgres engine",
			example:        "awsSdkDefaultPostgres",
			wantAuthType:   AWSAuthTypeDefault,
			wantEngine:     AuroraEnginePostgres,
			wantDefaultReg: "us-east-1",
			wantDBHost:     "my-cluster.cluster-xxxxxxxxxxxx.us-east-1.rds.amazonaws.com",
			wantDBPort:     5432,
			wantDBUser:     "iam_user",
			wantDBName:     "mydb",
		},
		{
			name:           "access and secret key with mysql engine",
			example:        "accessAndSecretKeyMysql",
			wantAuthType:   AWSAuthTypeKeys,
			wantEngine:     AuroraEngineMySQL,
			wantDefaultReg: "us-east-1",
			wantDBHost:     "my-cluster.cluster-xxxxxxxxxxxx.us-east-1.rds.amazonaws.com",
			wantDBPort:     3306,
			wantDBUser:     "iam_user",
			wantDBName:     "mydb",
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyAccessKey, SecureJsonDataKeySecretKey},
			wantAccessKey:  "AKIAIOSFODNN7EXAMPLE",
			wantSecretKey:  "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		},
		{
			name:           "credentials file",
			example:        "credentialsFile",
			wantAuthType:   AWSAuthTypeCredentials,
			wantEngine:     AuroraEnginePostgres,
			wantDefaultReg: "us-east-1",
			wantDBHost:     "my-cluster.cluster-xxxxxxxxxxxx.us-east-1.rds.amazonaws.com",
			wantDBPort:     5432,
			wantDBUser:     "iam_user",
			wantDBName:     "mydb",
		},
		{
			name:           "workspace iam role",
			example:        "workspaceIamRole",
			wantAuthType:   AWSAuthTypeEC2IAMRole,
			wantEngine:     AuroraEnginePostgres,
			wantDefaultReg: "us-east-1",
			wantDBHost:     "my-cluster.cluster-xxxxxxxxxxxx.us-east-1.rds.amazonaws.com",
			wantDBPort:     5432,
			wantDBUser:     "iam_user",
			wantDBName:     "mydb",
		},
		{
			name:           "grafana assume role",
			example:        "grafanaAssumeRole",
			wantAuthType:   AWSAuthTypeGrafanaAssumeRole,
			wantEngine:     AuroraEnginePostgres,
			wantDefaultReg: "us-east-1",
			wantDBHost:     "my-cluster.cluster-xxxxxxxxxxxx.us-east-1.rds.amazonaws.com",
			wantDBPort:     5432,
			wantDBUser:     "iam_user",
			wantDBName:     "mydb",
		},
		{
			name:           "keys with sts assume role",
			example:        "assumeRoleFromKeys",
			wantAuthType:   AWSAuthTypeKeys,
			wantEngine:     AuroraEnginePostgres,
			wantDefaultReg: "us-east-1",
			wantDBHost:     "my-cluster.cluster-xxxxxxxxxxxx.us-east-1.rds.amazonaws.com",
			wantDBPort:     5432,
			wantDBUser:     "iam_user",
			wantDBName:     "mydb",
			wantAssumeARN:  "arn:aws:iam::123456789012:role/GrafanaAurora",
			wantExternalID: "external-id-abc123",
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyAccessKey, SecureJsonDataKeySecretKey},
			wantAccessKey:  "AKIAIOSFODNN7EXAMPLE",
			wantSecretKey:  "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		},
		{
			name:           "split auth endpoint",
			example:        "splitAuthEndpoint",
			wantAuthType:   AWSAuthTypeEC2IAMRole,
			wantEngine:     AuroraEnginePostgres,
			wantDefaultReg: "us-east-1",
			wantDBHost:     "aurora-lb.internal",
			wantDBPort:     5432,
			wantDBUser:     "iam_user",
			wantDBName:     "mydb",
			wantDBHostAuth: "my-cluster.cluster-xxxxxxxxxxxx.us-east-1.rds.amazonaws.com",
			wantDBPortAuth: 5432,
		},
		{
			name:           "legacy arn auth type",
			example:        "legacyArnAuthType",
			wantAuthType:   AWSAuthTypeARN,
			wantEngine:     AuroraEnginePostgres,
			wantDefaultReg: "us-east-1",
			wantDBHost:     "my-cluster.cluster-xxxxxxxxxxxx.us-east-1.rds.amazonaws.com",
			wantDBPort:     5432,
			wantDBUser:     "iam_user",
			wantDBName:     "mydb",
		},
		{
			// legacyMissingEngine: no engine field on the wire, ApplyDefaults
			// should fill in aurora-postgres.
			name:           "legacy missing engine defaults to postgres",
			example:        "legacyMissingEngine",
			wantAuthType:   AWSAuthTypeDefault,
			wantEngine:     AuroraEnginePostgres,
			wantDefaultReg: "us-east-1",
			wantDBHost:     "my-cluster.cluster-xxxxxxxxxxxx.us-east-1.rds.amazonaws.com",
			wantDBPort:     5432,
			wantDBUser:     "iam_user",
			wantDBName:     "mydb",
		},
		{
			name: "malformed jsonData errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{`),
			},
			wantErr: errors.New("parse jsonData"),
		},
		{
			name: "session token loaded from secure data",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{
                    "authType":"keys",
                    "defaultRegion":"us-east-1",
                    "engine":"aurora-postgres",
                    "dbHost":"h",
                    "dbPort":5432,
                    "dbUser":"u",
                    "dbName":"d"
                }`),
				DecryptedSecureJSONData: map[string]string{
					"accessKey":    "AKID",
					"secretKey":    "SECRET",
					"sessionToken": "STS-TOKEN",
				},
			},
			wantAuthType:     AWSAuthTypeKeys,
			wantEngine:       AuroraEnginePostgres,
			wantDefaultReg:   "us-east-1",
			wantDBHost:       "h",
			wantDBPort:       5432,
			wantDBUser:       "u",
			wantDBName:       "d",
			wantSecureKeys:   SecureJsonDataConfig{SecureJsonDataKeyAccessKey, SecureJsonDataKeySecretKey, SecureJsonDataKeySessionToken},
			wantAccessKey:    "AKID",
			wantSecretKey:    "SECRET",
			wantSessionToken: "STS-TOKEN",
		},
		{
			name: "empty settings default to AWS SDK default + postgres and fail validation",
			// After ApplyDefaults auth type is "default" and engine is
			// aurora-postgres, but the Aurora selectors are empty, so
			// Validate rejects the config.
			settings: backend.DataSourceInstanceSettings{},
			wantErr:  errors.New("defaultRegion is required"),
		},
		{
			name: "unknown auth type errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"authType":"nonsense","defaultRegion":"us-east-1","engine":"aurora-postgres","dbHost":"h","dbPort":5432,"dbUser":"u"}`),
			},
			wantErr: errors.New(`unknown authType "nonsense"`),
		},
		{
			name: "unknown engine errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"authType":"default","defaultRegion":"us-east-1","engine":"aurora-oracle","dbHost":"h","dbPort":5432,"dbUser":"u"}`),
			},
			wantErr: errors.New(`unknown engine "aurora-oracle"`),
		},
		{
			name: "keys auth without accessKey errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"authType":"keys","defaultRegion":"us-east-1","engine":"aurora-postgres","dbHost":"h","dbPort":5432,"dbUser":"u"}`),
			},
			wantErr: errors.New("accessKey is required for keys auth"),
		},
		{
			name: "keys auth with only accessKey still errors on secretKey",
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"authType":"keys","defaultRegion":"us-east-1","engine":"aurora-postgres","dbHost":"h","dbPort":5432,"dbUser":"u"}`),
				DecryptedSecureJSONData: map[string]string{"accessKey": "AKID"},
			},
			wantErr: errors.New("secretKey is required for keys auth"),
		},
		{
			name: "missing dbUser errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"authType":"default","defaultRegion":"us-east-1","engine":"aurora-postgres","dbHost":"h","dbPort":5432}`),
			},
			wantErr: errors.New("dbUser is required"),
		},
		{
			name: "missing dbHost errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"authType":"default","defaultRegion":"us-east-1","engine":"aurora-postgres","dbPort":5432,"dbUser":"u"}`),
			},
			wantErr: errors.New("dbHost is required"),
		},
		{
			name: "missing dbPort errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"authType":"default","defaultRegion":"us-east-1","engine":"aurora-postgres","dbHost":"h","dbUser":"u"}`),
			},
			wantErr: errors.New("dbPort is required"),
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
			if tt.wantEngine != "" && cfg.Engine != tt.wantEngine {
				t.Errorf("Engine = %q, want %q", cfg.Engine, tt.wantEngine)
			}
			if tt.wantDefaultReg != "" && cfg.DefaultRegion != tt.wantDefaultReg {
				t.Errorf("DefaultRegion = %q, want %q", cfg.DefaultRegion, tt.wantDefaultReg)
			}
			if tt.wantDBHost != "" && cfg.DBHost != tt.wantDBHost {
				t.Errorf("DBHost = %q, want %q", cfg.DBHost, tt.wantDBHost)
			}
			if tt.wantDBPort != 0 && cfg.DBPort != tt.wantDBPort {
				t.Errorf("DBPort = %d, want %d", cfg.DBPort, tt.wantDBPort)
			}
			if tt.wantDBUser != "" && cfg.DBUser != tt.wantDBUser {
				t.Errorf("DBUser = %q, want %q", cfg.DBUser, tt.wantDBUser)
			}
			if tt.wantDBName != "" && cfg.DBName != tt.wantDBName {
				t.Errorf("DBName = %q, want %q", cfg.DBName, tt.wantDBName)
			}
			if tt.wantDBHostAuth != "" && cfg.DBHostAuth != tt.wantDBHostAuth {
				t.Errorf("DBHostAuth = %q, want %q", cfg.DBHostAuth, tt.wantDBHostAuth)
			}
			if tt.wantDBPortAuth != 0 && cfg.DBPortAuth != tt.wantDBPortAuth {
				t.Errorf("DBPortAuth = %d, want %d", cfg.DBPortAuth, tt.wantDBPortAuth)
			}
			if tt.wantAssumeARN != "" && cfg.AssumeRoleARN != tt.wantAssumeARN {
				t.Errorf("AssumeRoleARN = %q, want %q", cfg.AssumeRoleARN, tt.wantAssumeARN)
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
			if tt.wantSessionToken != "" && cfg.DecryptedSecureJSONData[SecureJsonDataKeySessionToken] != tt.wantSessionToken {
				t.Errorf("Secrets[sessionToken] = %q, want %q", cfg.DecryptedSecureJSONData[SecureJsonDataKeySessionToken], tt.wantSessionToken)
			}
		})
	}
}

func TestApplyDefaults(t *testing.T) {
	tests := []struct {
		name       string
		in         Config
		wantAuth   AWSAuthType
		wantEngine AuroraEngine
	}{
		{
			name:       "empty config gets default auth type and postgres engine",
			in:         Config{},
			wantAuth:   AWSAuthTypeDefault,
			wantEngine: AuroraEnginePostgres,
		},
		{
			name:       "existing auth type is preserved",
			in:         Config{AuthType: AWSAuthTypeKeys},
			wantAuth:   AWSAuthTypeKeys,
			wantEngine: AuroraEnginePostgres,
		},
		{
			name:       "existing engine is preserved",
			in:         Config{Engine: AuroraEngineMySQL},
			wantAuth:   AWSAuthTypeDefault,
			wantEngine: AuroraEngineMySQL,
		},
		{
			name:       "legacy arn is preserved",
			in:         Config{AuthType: AWSAuthTypeARN},
			wantAuth:   AWSAuthTypeARN,
			wantEngine: AuroraEnginePostgres,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.in
			got.ApplyDefaults()
			if got.AuthType != tt.wantAuth {
				t.Errorf("AuthType = %q, want %q", got.AuthType, tt.wantAuth)
			}
			if got.Engine != tt.wantEngine {
				t.Errorf("Engine = %q, want %q", got.Engine, tt.wantEngine)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	// Every test case that expects no error starts from a config with the
	// four Aurora selectors populated; each case adjusts the auth-related
	// or engine fields to exercise its scenario.
	basic := func() Config {
		return Config{
			DefaultRegion:           "us-east-1",
			Engine:                  AuroraEnginePostgres,
			DBHost:                  "h",
			DBPort:                  5432,
			DBUser:                  "u",
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
			name: "aurora mysql engine happy path",
			mutate: func(c *Config) {
				c.AuthType = AWSAuthTypeDefault
				c.Engine = AuroraEngineMySQL
				c.DBPort = 3306
			},
		},
		{
			name: "split auth endpoint happy path",
			mutate: func(c *Config) {
				c.AuthType = AWSAuthTypeDefault
				c.DBHostAuth = "auth-host"
				c.DBPortAuth = 5433
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
			name: "unknown engine errors",
			mutate: func(c *Config) {
				c.AuthType = AWSAuthTypeDefault
				c.Engine = "aurora-oracle"
			},
			wantErr: `unknown engine "aurora-oracle"`,
		},
		{
			name: "missing defaultRegion errors",
			mutate: func(c *Config) {
				c.AuthType = AWSAuthTypeDefault
				c.DefaultRegion = ""
			},
			wantErr: "defaultRegion is required",
		},
		{
			name: "missing dbUser errors",
			mutate: func(c *Config) {
				c.AuthType = AWSAuthTypeDefault
				c.DBUser = ""
			},
			wantErr: "dbUser is required",
		},
		{
			name: "missing dbHost errors",
			mutate: func(c *Config) {
				c.AuthType = AWSAuthTypeDefault
				c.DBHost = ""
			},
			wantErr: "dbHost is required",
		},
		{
			name: "missing dbPort errors",
			mutate: func(c *Config) {
				c.AuthType = AWSAuthTypeDefault
				c.DBPort = 0
			},
			wantErr: "dbPort is required",
		},
		{
			name: "negative dbPortAuth errors",
			mutate: func(c *Config) {
				c.AuthType = AWSAuthTypeDefault
				c.DBPortAuth = -1
			},
			wantErr: "dbPortAuth must be non-negative",
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
