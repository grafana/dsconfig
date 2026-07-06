package cloudwatchdatasource

import (
	"encoding/json"
	"errors"
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

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name               string
		example            string
		settings           backend.DataSourceInstanceSettings
		wantErr            error
		wantAuthType       AWSAuthType
		wantDefaultReg     string
		wantProxyType      AWSProxyType
		wantProxyURL       string
		wantProxyUsername  string
		wantEndpoint       string
		wantAssumeARN      string
		wantExternalID     string
		wantNamespaces     string
		wantLogsTimeout    time.Duration
		wantLogGroups      int
		wantDefaultLGCount int
		wantTracingUID     string
		wantSecureKeys     SecureJsonDataConfig
		wantAccessKey      string
		wantSecretKey      string
		wantSessionToken   string
		wantProxyPassword  string
	}{
		{
			// The default schema example intentionally leaves defaultRegion
			// blank so LoadConfig's Validate step rejects it.
			name:    "default example fails validation (missing defaultRegion)",
			example: "",
			wantErr: errors.New("defaultRegion is required"),
		},
		{
			name:            "aws sdk default",
			example:         "awsSdkDefault",
			wantAuthType:    AWSAuthTypeDefault,
			wantDefaultReg:  "us-east-1",
			wantProxyType:   AWSProxyTypeEnv,
			wantLogsTimeout: 30 * time.Minute,
			wantNamespaces:  "AWS/EC2,AWS/ELB",
		},
		{
			name:            "access and secret key",
			example:         "accessAndSecretKey",
			wantAuthType:    AWSAuthTypeKeys,
			wantDefaultReg:  "us-east-1",
			wantProxyType:   AWSProxyTypeEnv,
			wantLogsTimeout: 30 * time.Minute,
			wantSecureKeys:  SecureJsonDataConfig{SecureJsonDataKeyAccessKey, SecureJsonDataKeySecretKey},
			wantAccessKey:   "AKIAIOSFODNN7EXAMPLE",
			wantSecretKey:   "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		},
		{
			name:            "credentials file",
			example:         "credentialsFile",
			wantAuthType:    AWSAuthTypeCredentials,
			wantDefaultReg:  "us-east-1",
			wantProxyType:   AWSProxyTypeEnv,
			wantLogsTimeout: 30 * time.Minute,
		},
		{
			name:            "workspace iam role",
			example:         "workspaceIamRole",
			wantAuthType:    AWSAuthTypeEC2IAMRole,
			wantDefaultReg:  "us-east-1",
			wantProxyType:   AWSProxyTypeEnv,
			wantLogsTimeout: 30 * time.Minute,
		},
		{
			name:            "grafana assume role",
			example:         "grafanaAssumeRole",
			wantAuthType:    AWSAuthTypeGrafanaAssumeRole,
			wantDefaultReg:  "us-east-1",
			wantProxyType:   AWSProxyTypeEnv,
			wantLogsTimeout: 30 * time.Minute,
		},
		{
			name:            "keys with sts assume role",
			example:         "assumeRoleFromKeys",
			wantAuthType:    AWSAuthTypeKeys,
			wantDefaultReg:  "us-east-1",
			wantProxyType:   AWSProxyTypeEnv,
			wantLogsTimeout: 30 * time.Minute,
			wantAssumeARN:   "arn:aws:iam::123456789012:role/GrafanaCloudWatch",
			wantExternalID:  "external-id-abc123",
			wantSecureKeys:  SecureJsonDataConfig{SecureJsonDataKeyAccessKey, SecureJsonDataKeySecretKey},
			wantAccessKey:   "AKIAIOSFODNN7EXAMPLE",
			wantSecretKey:   "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		},
		{
			name:              "url proxy",
			example:           "urlProxy",
			wantAuthType:      AWSAuthTypeDefault,
			wantDefaultReg:    "us-east-1",
			wantProxyType:     AWSProxyTypeURL,
			wantProxyURL:      "https://proxy.internal:3128",
			wantProxyUsername: "grafana",
			wantLogsTimeout:   30 * time.Minute,
			wantSecureKeys:    SecureJsonDataConfig{SecureJsonDataKeyProxyPassword},
			wantProxyPassword: "proxy-password-placeholder",
		},
		{
			name:            "cloudwatch logs defaults",
			example:         "cloudwatchLogsDefaults",
			wantAuthType:    AWSAuthTypeDefault,
			wantDefaultReg:  "us-east-1",
			wantProxyType:   AWSProxyTypeEnv,
			wantLogsTimeout: 10 * time.Minute,
			wantNamespaces:  "AWS/EC2,AWS/RDS,MyApp/Custom",
			wantLogGroups:   2,
			wantTracingUID:  "aws-xray-uid",
		},
		{
			name:            "legacy arn auth type",
			example:         "legacyArnAuthType",
			wantAuthType:    AWSAuthTypeARN,
			wantDefaultReg:  "us-east-1",
			wantProxyType:   AWSProxyTypeEnv,
			wantLogsTimeout: 30 * time.Minute,
		},
		{
			name:               "legacy defaultLogGroups shape",
			example:            "legacyDefaultLogGroups",
			wantAuthType:       AWSAuthTypeDefault,
			wantDefaultReg:     "us-east-1",
			wantProxyType:      AWSProxyTypeEnv,
			wantLogsTimeout:    30 * time.Minute,
			wantDefaultLGCount: 2,
		},
		{
			name: "malformed jsonData errors",
			// Use a longer malformed payload than a single `{`. Upstream
			// LoadCloudWatchSettings gates parsing on `len > 1`
			// (pkg/cloudwatch/models/settings.go:29), so single-byte inputs
			// are silently ignored — LoadConfig mirrors that quirk verbatim.
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"authType":`),
			},
			wantErr: errors.New("parse jsonData"),
		},
		{
			name: "single-byte jsonData is tolerated (upstream len>1 gate)",
			// Mirrors LoadCloudWatchSettings quirk: `{` (or any 1-byte body)
			// bypasses json.Unmarshal, so the config takes its
			// ApplyDefaults-filled defaults and then fails Validate on the
			// runtime-required defaultRegion.
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{`),
			},
			wantErr: errors.New("defaultRegion is required"),
		},
		{
			name: "logsTimeout parses numeric nanoseconds",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"authType":"default","defaultRegion":"us-east-1","logsTimeout":1500000000}`),
			},
			wantAuthType:    AWSAuthTypeDefault,
			wantDefaultReg:  "us-east-1",
			wantProxyType:   AWSProxyTypeEnv,
			wantLogsTimeout: 1500 * time.Millisecond,
		},
		{
			name: "logsTimeout empty string falls back to default",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"authType":"default","defaultRegion":"us-east-1","logsTimeout":""}`),
			},
			wantAuthType:    AWSAuthTypeDefault,
			wantDefaultReg:  "us-east-1",
			wantProxyType:   AWSProxyTypeEnv,
			wantLogsTimeout: 30 * time.Minute,
		},
		{
			name: "logsTimeout invalid duration returns downstream error",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"authType":"default","defaultRegion":"us-east-1","logsTimeout":"10mm"}`),
			},
			wantErr: errors.New("parse jsonData"),
		},
		{
			name: "session token loaded from secure data",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"authType":"keys","defaultRegion":"us-east-1"}`),
				DecryptedSecureJSONData: map[string]string{
					"accessKey":    "AKID",
					"secretKey":    "SECRET",
					"sessionToken": "STS-TOKEN",
				},
			},
			wantAuthType:     AWSAuthTypeKeys,
			wantDefaultReg:   "us-east-1",
			wantProxyType:    AWSProxyTypeEnv,
			wantLogsTimeout:  30 * time.Minute,
			wantSecureKeys:   SecureJsonDataConfig{SecureJsonDataKeyAccessKey, SecureJsonDataKeySecretKey, SecureJsonDataKeySessionToken},
			wantAccessKey:    "AKID",
			wantSecretKey:    "SECRET",
			wantSessionToken: "STS-TOKEN",
		},
		{
			name: "pascalcase assumeRoleARN (backend spelling) still decodes",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{
                    "authType":"default",
                    "defaultRegion":"us-east-1",
                    "assumeRoleARN":"arn:aws:iam::123456789012:role/Legacy"
                }`),
			},
			wantAuthType:    AWSAuthTypeDefault,
			wantDefaultReg:  "us-east-1",
			wantProxyType:   AWSProxyTypeEnv,
			wantLogsTimeout: 30 * time.Minute,
			wantAssumeARN:   "arn:aws:iam::123456789012:role/Legacy",
		},
		{
			name:     "empty settings default to AWS SDK default and fail validation",
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
			name: "url proxy without proxyUrl errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"authType":"default","defaultRegion":"us-east-1","proxyType":"url"}`),
			},
			wantErr: errors.New("proxyUrl is required when proxyType is 'url'"),
		},
		{
			name: "unknown proxy type errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"authType":"default","defaultRegion":"us-east-1","proxyType":"weird"}`),
			},
			wantErr: errors.New(`unknown proxyType "weird"`),
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
			if tt.wantProxyType != "" && cfg.ProxyType != tt.wantProxyType {
				t.Errorf("ProxyType = %q, want %q", cfg.ProxyType, tt.wantProxyType)
			}
			if tt.wantProxyURL != "" && cfg.ProxyURL != tt.wantProxyURL {
				t.Errorf("ProxyURL = %q, want %q", cfg.ProxyURL, tt.wantProxyURL)
			}
			if tt.wantProxyUsername != "" && cfg.ProxyUsername != tt.wantProxyUsername {
				t.Errorf("ProxyUsername = %q, want %q", cfg.ProxyUsername, tt.wantProxyUsername)
			}
			if tt.wantEndpoint != "" && cfg.Endpoint != tt.wantEndpoint {
				t.Errorf("Endpoint = %q, want %q", cfg.Endpoint, tt.wantEndpoint)
			}
			if tt.wantAssumeARN != "" && cfg.AssumeRoleARN != tt.wantAssumeARN {
				t.Errorf("AssumeRoleARN = %q, want %q", cfg.AssumeRoleARN, tt.wantAssumeARN)
			}
			if tt.wantExternalID != "" && cfg.ExternalID != tt.wantExternalID {
				t.Errorf("ExternalID = %q, want %q", cfg.ExternalID, tt.wantExternalID)
			}
			if tt.wantNamespaces != "" && cfg.CustomMetricsNamespaces != tt.wantNamespaces {
				t.Errorf("CustomMetricsNamespaces = %q, want %q", cfg.CustomMetricsNamespaces, tt.wantNamespaces)
			}
			if tt.wantLogsTimeout != 0 && cfg.LogsTimeout.Duration != tt.wantLogsTimeout {
				t.Errorf("LogsTimeout = %v, want %v", cfg.LogsTimeout.Duration, tt.wantLogsTimeout)
			}
			if tt.wantLogGroups != 0 && len(cfg.LogGroups) != tt.wantLogGroups {
				t.Errorf("len(LogGroups) = %d, want %d", len(cfg.LogGroups), tt.wantLogGroups)
			}
			if tt.wantDefaultLGCount != 0 && len(cfg.DefaultLogGroups) != tt.wantDefaultLGCount {
				t.Errorf("len(DefaultLogGroups) = %d, want %d", len(cfg.DefaultLogGroups), tt.wantDefaultLGCount)
			}
			if tt.wantTracingUID != "" && cfg.TracingDatasourceUID != tt.wantTracingUID {
				t.Errorf("TracingDatasourceUID = %q, want %q", cfg.TracingDatasourceUID, tt.wantTracingUID)
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
			if tt.wantProxyPassword != "" && cfg.DecryptedSecureJSONData[SecureJsonDataKeyProxyPassword] != tt.wantProxyPassword {
				t.Errorf("Secrets[proxyPassword] = %q, want %q", cfg.DecryptedSecureJSONData[SecureJsonDataKeyProxyPassword], tt.wantProxyPassword)
			}
		})
	}
}

func TestApplyDefaults(t *testing.T) {
	tests := []struct {
		name            string
		in              Config
		wantAuthType    AWSAuthType
		wantProxyType   AWSProxyType
		wantLogsTimeout time.Duration
	}{
		{
			name:            "empty config gets defaults",
			in:              Config{},
			wantAuthType:    AWSAuthTypeDefault,
			wantProxyType:   AWSProxyTypeEnv,
			wantLogsTimeout: 30 * time.Minute,
		},
		{
			name:            "existing values are preserved",
			in:              Config{AuthType: AWSAuthTypeKeys, ProxyType: AWSProxyTypeURL, LogsTimeout: Duration{5 * time.Minute}},
			wantAuthType:    AWSAuthTypeKeys,
			wantProxyType:   AWSProxyTypeURL,
			wantLogsTimeout: 5 * time.Minute,
		},
		{
			name:            "legacy arn is preserved",
			in:              Config{AuthType: AWSAuthTypeARN},
			wantAuthType:    AWSAuthTypeARN,
			wantProxyType:   AWSProxyTypeEnv,
			wantLogsTimeout: 30 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.in
			got.ApplyDefaults()
			if got.AuthType != tt.wantAuthType {
				t.Errorf("AuthType = %q, want %q", got.AuthType, tt.wantAuthType)
			}
			if got.ProxyType != tt.wantProxyType {
				t.Errorf("ProxyType = %q, want %q", got.ProxyType, tt.wantProxyType)
			}
			if got.LogsTimeout.Duration != tt.wantLogsTimeout {
				t.Errorf("LogsTimeout = %v, want %v", got.LogsTimeout.Duration, tt.wantLogsTimeout)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	// Every test case that expects no error starts from a config with the
	// runtime-required fields populated; each case adjusts the field of
	// interest.
	basic := func() Config {
		return Config{
			AuthType:                AWSAuthTypeDefault,
			ProxyType:               AWSProxyTypeEnv,
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
		wantErr string
	}{
		{name: "default auth happy path", mutate: func(c *Config) {}},
		{
			name: "keys auth happy path",
			mutate: func(c *Config) {
				c.AuthType = AWSAuthTypeKeys
				withSecret(SecureJsonDataKeyAccessKey, "AKID")(c)
				withSecret(SecureJsonDataKeySecretKey, "SECRET")(c)
			},
		},
		{name: "credentials auth happy path", mutate: func(c *Config) { c.AuthType = AWSAuthTypeCredentials }},
		{name: "ec2 iam role happy path", mutate: func(c *Config) { c.AuthType = AWSAuthTypeEC2IAMRole }},
		{name: "grafana assume role happy path", mutate: func(c *Config) { c.AuthType = AWSAuthTypeGrafanaAssumeRole }},
		{name: "legacy arn happy path", mutate: func(c *Config) { c.AuthType = AWSAuthTypeARN }},
		{
			name:    "keys without accessKey errors",
			mutate:  func(c *Config) { c.AuthType = AWSAuthTypeKeys; withSecret(SecureJsonDataKeySecretKey, "SECRET")(c) },
			wantErr: "accessKey is required for keys auth",
		},
		{
			name:    "keys without secretKey errors",
			mutate:  func(c *Config) { c.AuthType = AWSAuthTypeKeys; withSecret(SecureJsonDataKeyAccessKey, "AKID")(c) },
			wantErr: "secretKey is required for keys auth",
		},
		{
			name:    "empty auth type errors as unknown",
			mutate:  func(c *Config) { c.AuthType = "" },
			wantErr: `unknown authType ""`,
		},
		{
			name:    "unknown auth type errors",
			mutate:  func(c *Config) { c.AuthType = "totally-not-real" },
			wantErr: `unknown authType "totally-not-real"`,
		},
		{
			name:    "missing defaultRegion errors",
			mutate:  func(c *Config) { c.DefaultRegion = "" },
			wantErr: "defaultRegion is required",
		},
		{
			name:    "url proxy without url errors",
			mutate:  func(c *Config) { c.ProxyType = AWSProxyTypeURL },
			wantErr: "proxyUrl is required when proxyType is 'url'",
		},
		{
			name: "url proxy with url is fine",
			mutate: func(c *Config) {
				c.ProxyType = AWSProxyTypeURL
				c.ProxyURL = "https://proxy.internal:3128"
			},
		},
		{
			name:    "unknown proxy type errors",
			mutate:  func(c *Config) { c.ProxyType = "weird" },
			wantErr: `unknown proxyType "weird"`,
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

func TestDurationRoundTrip(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    time.Duration
		wantErr bool
	}{
		{name: "string duration", input: `"10m"`, want: 10 * time.Minute},
		{name: "float duration string", input: `"1.5s"`, want: 1500 * time.Millisecond},
		{name: "nanosecond number", input: `1500000000`, want: 1500 * time.Millisecond},
		{name: "empty string is zero", input: `""`, want: 0},
		{name: "null is zero", input: `null`, want: 0},
		{name: "invalid string errors", input: `"10mm"`, wantErr: true},
		{name: "bool errors", input: `true`, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d Duration
			err := json.Unmarshal([]byte(tt.input), &d)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil (parsed to %v)", d.Duration)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if d.Duration != tt.want {
				t.Errorf("Duration = %v, want %v", d.Duration, tt.want)
			}
		})
	}
}
