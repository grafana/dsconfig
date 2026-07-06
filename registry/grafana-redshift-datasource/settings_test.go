package redshiftdatasource

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
		example          string
		settings         backend.DataSourceInstanceSettings
		wantErr          error
		wantAuthType     AWSAuthType
		wantDefaultReg   string
		wantServerless   bool
		wantManaged      bool
		wantCluster      string
		wantWorkgroup    string
		wantSecretARN    string
		wantSecretName   string
		wantDBUser       string
		wantDatabase     string
		wantWithEvent    bool
		wantAssumeARN    string
		wantExternalID   string
		wantSecureKeys   SecureJsonDataConfig
		wantAccessKey    string
		wantSecretKey    string
		wantSessionToken string
	}{
		{
			// The default schema example intentionally leaves the Redshift
			// selectors blank so LoadConfig's Validate step rejects it.
			name:    "default example fails validation (missing selectors)",
			example: "",
			wantErr: errors.New("defaultRegion is required"),
		},
		{
			name:           "provisioned + temp creds + keys",
			example:        "provisionedTempCredsKeys",
			wantAuthType:   AWSAuthTypeKeys,
			wantDefaultReg: "us-east-1",
			wantServerless: false,
			wantManaged:    false,
			wantCluster:    "my-redshift-cluster",
			wantDBUser:     "awsuser",
			wantDatabase:   "dev",
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyAccessKey, SecureJsonDataKeySecretKey},
			wantAccessKey:  "AKIAIOSFODNN7EXAMPLE",
			wantSecretKey:  "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		},
		{
			name:           "provisioned + managed secret + aws sdk default",
			example:        "provisionedManagedSecretDefault",
			wantAuthType:   AWSAuthTypeDefault,
			wantDefaultReg: "us-east-1",
			wantServerless: false,
			wantManaged:    true,
			wantCluster:    "my-redshift-cluster",
			wantDatabase:   "dev",
			wantSecretARN:  "arn:aws:secretsmanager:us-east-1:123456789012:secret:redshift-1-xxxxxx",
			wantSecretName: "redshift-1",
		},
		{
			name:           "serverless + temp creds + iam role",
			example:        "serverlessTempCredsIamRole",
			wantAuthType:   AWSAuthTypeEC2IAMRole,
			wantDefaultReg: "us-east-1",
			wantServerless: true,
			wantManaged:    false,
			wantWorkgroup:  "default",
			wantDatabase:   "dev",
		},
		{
			name:           "serverless + managed secret + grafana assume",
			example:        "serverlessManagedSecretGrafanaAssume",
			wantAuthType:   AWSAuthTypeGrafanaAssumeRole,
			wantDefaultReg: "us-east-1",
			wantServerless: true,
			wantManaged:    true,
			wantWorkgroup:  "default",
			wantDatabase:   "dev",
			wantSecretARN:  "arn:aws:secretsmanager:us-east-1:123456789012:secret:redshift-serverless-xxxxxx",
			wantSecretName: "redshift-serverless",
		},
		{
			name:           "credentials file",
			example:        "credentialsFile",
			wantAuthType:   AWSAuthTypeCredentials,
			wantDefaultReg: "us-east-1",
			wantServerless: false,
			wantManaged:    false,
			wantCluster:    "my-redshift-cluster",
			wantDBUser:     "awsuser",
			wantDatabase:   "dev",
		},
		{
			name:           "keys with sts assume role",
			example:        "assumeRoleFromKeys",
			wantAuthType:   AWSAuthTypeKeys,
			wantDefaultReg: "us-east-1",
			wantServerless: false,
			wantManaged:    false,
			wantCluster:    "my-redshift-cluster",
			wantDBUser:     "awsuser",
			wantDatabase:   "dev",
			wantAssumeARN:  "arn:aws:iam::123456789012:role/GrafanaRedshift",
			wantExternalID: "external-id-abc123",
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyAccessKey, SecureJsonDataKeySecretKey},
			wantAccessKey:  "AKIAIOSFODNN7EXAMPLE",
			wantSecretKey:  "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		},
		{
			name:           "with event bridge",
			example:        "withEventBridge",
			wantAuthType:   AWSAuthTypeKeys,
			wantDefaultReg: "us-east-1",
			wantCluster:    "my-redshift-cluster",
			wantDBUser:     "awsuser",
			wantDatabase:   "dev",
			wantWithEvent:  true,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyAccessKey, SecureJsonDataKeySecretKey},
			wantAccessKey:  "AKIAIOSFODNN7EXAMPLE",
			wantSecretKey:  "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		},
		{
			name:           "legacy arn auth type",
			example:        "legacyArnAuthType",
			wantAuthType:   AWSAuthTypeARN,
			wantDefaultReg: "us-east-1",
			wantCluster:    "my-redshift-cluster",
			wantDBUser:     "awsuser",
			wantDatabase:   "dev",
		},
		{
			// Upstream (pkg/redshift/models/settings.go:59) gates the
			// json.Unmarshal on `len > 1`, so a single-byte body like `{`
			// would be silently ignored. Use a body long enough to
			// actually trip the parser.
			name: "malformed jsonData errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"authType":`),
			},
			wantErr: errors.New("parse jsonData"),
		},
		{
			// Upstream backend struct (pkg/redshift/models/settings.go:51)
			// declares `ManagedSecret ManagedSecret` with NO json tag, so
			// encoding/json would emit "ManagedSecret" (PascalCase). Case-
			// insensitive Unmarshal ensures a provisioned config using the
			// upstream spelling still loads through our camelCase-tagged
			// Config.
			name: "pascalcase ManagedSecret is accepted (case-insensitive decode)",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{
                    "authType":"default",
                    "defaultRegion":"us-east-1",
                    "useServerless":false,
                    "useManagedSecret":true,
                    "clusterIdentifier":"c1",
                    "database":"dev",
                    "ManagedSecret":{"arn":"arn:aws:secretsmanager:us-east-1:123456789012:secret:legacy","name":"legacy"}
                }`),
			},
			wantAuthType:   AWSAuthTypeDefault,
			wantDefaultReg: "us-east-1",
			wantServerless: false,
			wantManaged:    true,
			wantCluster:    "c1",
			wantDatabase:   "dev",
			wantSecretARN:  "arn:aws:secretsmanager:us-east-1:123456789012:secret:legacy",
			wantSecretName: "legacy",
		},
		{
			name: "session token loaded from secure data",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{
                    "authType":"keys",
                    "defaultRegion":"us-east-1",
                    "useServerless":false,
                    "useManagedSecret":false,
                    "clusterIdentifier":"c1",
                    "database":"dev",
                    "dbUser":"u"
                }`),
				DecryptedSecureJSONData: map[string]string{
					"accessKey":    "AKID",
					"secretKey":    "SECRET",
					"sessionToken": "STS-TOKEN",
				},
			},
			wantAuthType:     AWSAuthTypeKeys,
			wantDefaultReg:   "us-east-1",
			wantCluster:      "c1",
			wantDatabase:     "dev",
			wantDBUser:       "u",
			wantSecureKeys:   SecureJsonDataConfig{SecureJsonDataKeyAccessKey, SecureJsonDataKeySecretKey, SecureJsonDataKeySessionToken},
			wantAccessKey:    "AKID",
			wantSecretKey:    "SECRET",
			wantSessionToken: "STS-TOKEN",
		},
		{
			name:     "empty settings default to AWS SDK default and fail validation",
			settings: backend.DataSourceInstanceSettings{},
			wantErr:  errors.New("defaultRegion is required"),
		},
		{
			name: "unknown auth type errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"authType":"nonsense","defaultRegion":"us-east-1","clusterIdentifier":"c","database":"d","dbUser":"u"}`),
			},
			wantErr: errors.New(`unknown authType "nonsense"`),
		},
		{
			name: "keys auth without accessKey errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"authType":"keys","defaultRegion":"us-east-1","clusterIdentifier":"c","database":"d","dbUser":"u"}`),
			},
			wantErr: errors.New("accessKey is required for keys auth"),
		},
		{
			name: "keys auth with only accessKey still errors on secretKey",
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"authType":"keys","defaultRegion":"us-east-1","clusterIdentifier":"c","database":"d","dbUser":"u"}`),
				DecryptedSecureJSONData: map[string]string{"accessKey": "AKID"},
			},
			wantErr: errors.New("secretKey is required for keys auth"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings := tt.settings
			if tt.example != "" {
				if _, ok := SettingsExamples().Examples[tt.example]; ok {
					settings = settingsFromExample(t, tt.example)
				}
			} else if tt.settings.JSONData == nil && tt.settings.DecryptedSecureJSONData == nil && tt.wantErr == nil {
				if _, ok := SettingsExamples().Examples[""]; ok {
					settings = settingsFromExample(t, "")
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
			if cfg.UseServerless != tt.wantServerless {
				t.Errorf("UseServerless = %v, want %v", cfg.UseServerless, tt.wantServerless)
			}
			if cfg.UseManagedSecret != tt.wantManaged {
				t.Errorf("UseManagedSecret = %v, want %v", cfg.UseManagedSecret, tt.wantManaged)
			}
			if tt.wantCluster != "" && cfg.ClusterIdentifier != tt.wantCluster {
				t.Errorf("ClusterIdentifier = %q, want %q", cfg.ClusterIdentifier, tt.wantCluster)
			}
			if tt.wantWorkgroup != "" && cfg.WorkgroupName != tt.wantWorkgroup {
				t.Errorf("WorkgroupName = %q, want %q", cfg.WorkgroupName, tt.wantWorkgroup)
			}
			if tt.wantSecretARN != "" && cfg.ManagedSecret.ARN != tt.wantSecretARN {
				t.Errorf("ManagedSecret.ARN = %q, want %q", cfg.ManagedSecret.ARN, tt.wantSecretARN)
			}
			if tt.wantSecretName != "" && cfg.ManagedSecret.Name != tt.wantSecretName {
				t.Errorf("ManagedSecret.Name = %q, want %q", cfg.ManagedSecret.Name, tt.wantSecretName)
			}
			if tt.wantDBUser != "" && cfg.DBUser != tt.wantDBUser {
				t.Errorf("DBUser = %q, want %q", cfg.DBUser, tt.wantDBUser)
			}
			if tt.wantDatabase != "" && cfg.Database != tt.wantDatabase {
				t.Errorf("Database = %q, want %q", cfg.Database, tt.wantDatabase)
			}
			if cfg.WithEvent != tt.wantWithEvent {
				t.Errorf("WithEvent = %v, want %v", cfg.WithEvent, tt.wantWithEvent)
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
	// Every test case that expects no error starts from a fully-populated
	// Provisioned + temp-creds config; each case adjusts the fields to
	// exercise its scenario.
	basic := func() Config {
		return Config{
			AuthType:                AWSAuthTypeDefault,
			DefaultRegion:           "us-east-1",
			UseServerless:           false,
			UseManagedSecret:        false,
			ClusterIdentifier:       "my-cluster",
			Database:                "dev",
			DBUser:                  "awsuser",
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
			name:   "provisioned temp creds happy path",
			mutate: func(c *Config) {},
		},
		{
			name: "provisioned managed secret happy path",
			mutate: func(c *Config) {
				c.UseManagedSecret = true
				c.ManagedSecret = ManagedSecret{ARN: "arn:aws:secretsmanager:us-east-1:123456789012:secret:x", Name: "x"}
				c.DBUser = "" // read from secret at runtime
			},
		},
		{
			name: "serverless temp creds happy path",
			mutate: func(c *Config) {
				c.UseServerless = true
				c.ClusterIdentifier = ""
				c.WorkgroupName = "default"
				c.DBUser = "" // GetCredentials issues the username
			},
		},
		{
			name: "serverless managed secret happy path",
			mutate: func(c *Config) {
				c.UseServerless = true
				c.UseManagedSecret = true
				c.ClusterIdentifier = ""
				c.WorkgroupName = "default"
				c.ManagedSecret = ManagedSecret{ARN: "arn:aws:secretsmanager:us-east-1:123456789012:secret:x", Name: "x"}
				c.DBUser = ""
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
			name: "workspace iam role happy path",
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
			name: "missing defaultRegion errors",
			mutate: func(c *Config) {
				c.DefaultRegion = ""
			},
			wantErr: "defaultRegion is required",
		},
		{
			name: "missing clusterIdentifier on provisioned errors",
			mutate: func(c *Config) {
				c.ClusterIdentifier = ""
			},
			wantErr: "clusterIdentifier is required when useServerless is false",
		},
		{
			name: "missing workgroupName on serverless errors",
			mutate: func(c *Config) {
				c.UseServerless = true
				c.ClusterIdentifier = ""
				c.WorkgroupName = ""
			},
			wantErr: "workgroupName is required when useServerless is true",
		},
		{
			name: "missing managedSecret.arn errors",
			mutate: func(c *Config) {
				c.UseManagedSecret = true
			},
			wantErr: "managedSecret.arn is required when useManagedSecret is true",
		},
		{
			name: "missing dbUser on provisioned temp creds errors",
			mutate: func(c *Config) {
				c.DBUser = ""
			},
			wantErr: "dbUser is required when useManagedSecret is false and useServerless is false",
		},
		{
			name: "missing database errors",
			mutate: func(c *Config) {
				c.Database = ""
			},
			wantErr: "database is required",
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
