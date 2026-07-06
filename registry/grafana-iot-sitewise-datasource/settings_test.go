package iotsitewisedatasource

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
		wantEndpoint     string
		wantAssumeARN    string
		wantExternalID   string
		wantEdgeAuthMode EdgeAuthMode
		wantEdgeAuthUser string
		wantAccessKey    string
		wantSecretKey    string
		wantSessionToken string
		wantEdgeAuthPass string
		wantCert         string
		wantSecureKeys   SecureJsonDataConfig
	}{
		{
			// The default schema example intentionally leaves defaultRegion
			// blank so LoadConfig succeeds (no Edge branch fires and AWS SDK
			// default has no runtime requirements).
			name:           "default example loads",
			example:        "",
			wantAuthType:   AWSAuthTypeDefault,
			wantDefaultReg: "",
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
			wantAccessKey:  "AKIAIOSFODNN7EXAMPLE",
			wantSecretKey:  "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyAccessKey, SecureJsonDataKeySecretKey},
		},
		{
			name:           "credentials file",
			example:        "credentialsFile",
			wantAuthType:   AWSAuthTypeCredentials,
			wantDefaultReg: "us-east-1",
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
			wantAssumeARN:  "arn:aws:iam::123456789012:role/GrafanaSiteWise",
			wantExternalID: "external-id-abc123",
			wantAccessKey:  "AKIAIOSFODNN7EXAMPLE",
			wantSecretKey:  "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyAccessKey, SecureJsonDataKeySecretKey},
		},
		{
			name:             "edge standard (default auth mode)",
			example:          "edgeStandard",
			wantAuthType:     AWSAuthTypeKeys,
			wantDefaultReg:   EdgeRegion,
			wantEndpoint:     "https://edge.example.local:8443",
			wantEdgeAuthMode: EdgeAuthModeDefault,
			wantAccessKey:    "AKIAIOSFODNN7EXAMPLE",
			wantSecretKey:    "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
			wantCert:         examplePEMCert,
			wantSecureKeys: SecureJsonDataConfig{
				SecureJsonDataKeyAccessKey,
				SecureJsonDataKeySecretKey,
				SecureJsonDataKeyCert,
			},
		},
		{
			name:             "edge linux",
			example:          "edgeLinux",
			wantAuthType:     AWSAuthTypeDefault,
			wantDefaultReg:   EdgeRegion,
			wantEndpoint:     "https://edge.example.local:8443",
			wantEdgeAuthMode: EdgeAuthModeLinux,
			wantEdgeAuthUser: "grafana",
			wantEdgeAuthPass: "example-linux-password",
			wantCert:         examplePEMCert,
			wantSecureKeys: SecureJsonDataConfig{
				SecureJsonDataKeyEdgeAuthPass,
				SecureJsonDataKeyCert,
			},
		},
		{
			name:             "edge ldap",
			example:          "edgeLdap",
			wantAuthType:     AWSAuthTypeDefault,
			wantDefaultReg:   EdgeRegion,
			wantEndpoint:     "https://edge.example.local:8443",
			wantEdgeAuthMode: EdgeAuthModeLDAP,
			wantEdgeAuthUser: "cn=grafana,ou=users,dc=example,dc=com",
			wantEdgeAuthPass: "example-ldap-password",
			wantCert:         examplePEMCert,
			wantSecureKeys: SecureJsonDataConfig{
				SecureJsonDataKeyEdgeAuthPass,
				SecureJsonDataKeyCert,
			},
		},
		{
			name:           "legacy arn auth type",
			example:        "legacyArnAuthType",
			wantAuthType:   AWSAuthTypeARN,
			wantDefaultReg: "us-east-1",
		},
		{
			// Upstream Load only unmarshals when len(JSONData) > 1
			// (setting.go:25); a two-byte broken payload must still error.
			name: "malformed jsonData errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"`),
			},
			wantErr: errors.New("parse jsonData"),
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
			wantAccessKey:    "AKID",
			wantSecretKey:    "SECRET",
			wantSessionToken: "STS-TOKEN",
			wantSecureKeys: SecureJsonDataConfig{
				SecureJsonDataKeyAccessKey,
				SecureJsonDataKeySecretKey,
				SecureJsonDataKeySessionToken,
			},
		},
		{
			name: "empty settings default to AWS SDK default",
			// Upstream Load only unmarshals when JSONData has >1 byte; empty
			// settings still succeed because ApplyDefaults fills authType
			// and no Edge branch requirements fire.
			settings:       backend.DataSourceInstanceSettings{},
			wantAuthType:   AWSAuthTypeDefault,
			wantDefaultReg: "",
		},
		{
			name: "edge region with empty edgeAuthMode defaults to default",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{
                    "authType":"default",
                    "defaultRegion":"Edge",
                    "endpoint":"https://edge.example.local:8443"
                }`),
				DecryptedSecureJSONData: map[string]string{
					"cert": examplePEMCert,
				},
			},
			wantAuthType:     AWSAuthTypeDefault,
			wantDefaultReg:   EdgeRegion,
			wantEndpoint:     "https://edge.example.local:8443",
			wantEdgeAuthMode: EdgeAuthModeDefault,
			wantCert:         examplePEMCert,
			wantSecureKeys:   SecureJsonDataConfig{SecureJsonDataKeyCert},
		},
		{
			name: "edge without endpoint errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"authType":"default","defaultRegion":"Edge"}`),
				DecryptedSecureJSONData: map[string]string{
					"cert": examplePEMCert,
				},
			},
			wantErr: errors.New("edge region requires an explicit endpoint"),
		},
		{
			name: "edge without cert errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"authType":"default","defaultRegion":"Edge","endpoint":"https://edge.local"}`),
			},
			wantErr: errors.New("edge region requires an SSL certificate"),
		},
		{
			name: "edge linux without user errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{
                    "authType":"default",
                    "defaultRegion":"Edge",
                    "endpoint":"https://edge.local",
                    "edgeAuthMode":"linux"
                }`),
				DecryptedSecureJSONData: map[string]string{
					"cert":         examplePEMCert,
					"edgeAuthPass": "pass",
				},
			},
			wantErr: errors.New("missing edge auth user"),
		},
		{
			name: "edge linux without password errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{
                    "authType":"default",
                    "defaultRegion":"Edge",
                    "endpoint":"https://edge.local",
                    "edgeAuthMode":"linux",
                    "edgeAuthUser":"grafana"
                }`),
				DecryptedSecureJSONData: map[string]string{
					"cert": examplePEMCert,
				},
			},
			wantErr: errors.New("missing edge auth password"),
		},
		{
			name: "unknown auth type errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"authType":"nonsense"}`),
			},
			wantErr: errors.New(`unknown authType "nonsense"`),
		},
		{
			name: "keys auth without accessKey errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"authType":"keys"}`),
			},
			wantErr: errors.New("accessKey is required for keys auth"),
		},
		{
			name: "keys auth with only accessKey still errors on secretKey",
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"authType":"keys"}`),
				DecryptedSecureJSONData: map[string]string{"accessKey": "AKID"},
			},
			wantErr: errors.New("secretKey is required for keys auth"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings := tt.settings
			if tt.example != "" || (tt.settings.JSONData == nil && tt.settings.DecryptedSecureJSONData == nil && tt.wantErr == nil && tt.example == "") {
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
			if cfg.DefaultRegion != tt.wantDefaultReg {
				t.Errorf("DefaultRegion = %q, want %q", cfg.DefaultRegion, tt.wantDefaultReg)
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
			if tt.wantEdgeAuthMode != "" && cfg.EdgeAuthMode != tt.wantEdgeAuthMode {
				t.Errorf("EdgeAuthMode = %q, want %q", cfg.EdgeAuthMode, tt.wantEdgeAuthMode)
			}
			if tt.wantEdgeAuthUser != "" && cfg.EdgeAuthUser != tt.wantEdgeAuthUser {
				t.Errorf("EdgeAuthUser = %q, want %q", cfg.EdgeAuthUser, tt.wantEdgeAuthUser)
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
			if tt.wantEdgeAuthPass != "" && cfg.DecryptedSecureJSONData[SecureJsonDataKeyEdgeAuthPass] != tt.wantEdgeAuthPass {
				t.Errorf("Secrets[edgeAuthPass] = %q, want %q", cfg.DecryptedSecureJSONData[SecureJsonDataKeyEdgeAuthPass], tt.wantEdgeAuthPass)
			}
			if tt.wantCert != "" && cfg.DecryptedSecureJSONData[SecureJsonDataKeyCert] != tt.wantCert {
				t.Errorf("Secrets[cert] = %q, want %q", cfg.DecryptedSecureJSONData[SecureJsonDataKeyCert], tt.wantCert)
			}
		})
	}
}

func TestApplyDefaults(t *testing.T) {
	tests := []struct {
		name             string
		in               Config
		wantAuth         AWSAuthType
		wantEdgeAuthMode EdgeAuthMode
	}{
		{
			name:     "empty config gets default auth type",
			in:       Config{},
			wantAuth: AWSAuthTypeDefault,
		},
		{
			name:     "existing auth type is preserved",
			in:       Config{AuthType: AWSAuthTypeKeys},
			wantAuth: AWSAuthTypeKeys,
		},
		{
			name:             "edge region defaults edgeAuthMode to default",
			in:               Config{DefaultRegion: EdgeRegion},
			wantAuth:         AWSAuthTypeDefault,
			wantEdgeAuthMode: EdgeAuthModeDefault,
		},
		{
			name:             "edge region preserves existing edgeAuthMode",
			in:               Config{DefaultRegion: EdgeRegion, EdgeAuthMode: EdgeAuthModeLinux},
			wantAuth:         AWSAuthTypeDefault,
			wantEdgeAuthMode: EdgeAuthModeLinux,
		},
		{
			name:             "non-edge region does not set edgeAuthMode",
			in:               Config{DefaultRegion: "us-east-1"},
			wantAuth:         AWSAuthTypeDefault,
			wantEdgeAuthMode: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.in
			got.ApplyDefaults()
			if got.AuthType != tt.wantAuth {
				t.Errorf("AuthType = %q, want %q", got.AuthType, tt.wantAuth)
			}
			if got.EdgeAuthMode != tt.wantEdgeAuthMode {
				t.Errorf("EdgeAuthMode = %q, want %q", got.EdgeAuthMode, tt.wantEdgeAuthMode)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	basic := func() Config {
		return Config{
			AuthType:                AWSAuthTypeDefault,
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
			name: "edge default (standard) happy path",
			mutate: func(c *Config) {
				c.DefaultRegion = EdgeRegion
				c.Endpoint = "https://edge.local"
				c.EdgeAuthMode = EdgeAuthModeDefault
				withSecret(SecureJsonDataKeyCert, examplePEMCert)(c)
			},
		},
		{
			name: "edge linux happy path",
			mutate: func(c *Config) {
				c.DefaultRegion = EdgeRegion
				c.Endpoint = "https://edge.local"
				c.EdgeAuthMode = EdgeAuthModeLinux
				c.EdgeAuthUser = "grafana"
				withSecret(SecureJsonDataKeyCert, examplePEMCert)(c)
				withSecret(SecureJsonDataKeyEdgeAuthPass, "pass")(c)
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
			name: "edge without endpoint errors",
			mutate: func(c *Config) {
				c.DefaultRegion = EdgeRegion
				c.EdgeAuthMode = EdgeAuthModeDefault
				withSecret(SecureJsonDataKeyCert, examplePEMCert)(c)
			},
			wantErr: "edge region requires an explicit endpoint",
		},
		{
			name: "edge without cert errors",
			mutate: func(c *Config) {
				c.DefaultRegion = EdgeRegion
				c.Endpoint = "https://edge.local"
				c.EdgeAuthMode = EdgeAuthModeDefault
			},
			wantErr: "edge region requires an SSL certificate",
		},
		{
			name: "edge unknown authMode errors",
			mutate: func(c *Config) {
				c.DefaultRegion = EdgeRegion
				c.Endpoint = "https://edge.local"
				c.EdgeAuthMode = "kerberos"
				withSecret(SecureJsonDataKeyCert, examplePEMCert)(c)
			},
			wantErr: `unknown edgeAuthMode "kerberos"`,
		},
		{
			name: "edge linux missing user errors",
			mutate: func(c *Config) {
				c.DefaultRegion = EdgeRegion
				c.Endpoint = "https://edge.local"
				c.EdgeAuthMode = EdgeAuthModeLinux
				withSecret(SecureJsonDataKeyCert, examplePEMCert)(c)
				withSecret(SecureJsonDataKeyEdgeAuthPass, "pass")(c)
			},
			wantErr: "missing edge auth user",
		},
		{
			name: "edge linux missing password errors",
			mutate: func(c *Config) {
				c.DefaultRegion = EdgeRegion
				c.Endpoint = "https://edge.local"
				c.EdgeAuthMode = EdgeAuthModeLinux
				c.EdgeAuthUser = "grafana"
				withSecret(SecureJsonDataKeyCert, examplePEMCert)(c)
			},
			wantErr: "missing edge auth password",
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
