package amazonprometheusdatasource

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// settingsFromExample converts a SettingsExamples entry (a full instance
// settings object with root fields, jsonData, and secureJsonData) into the
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
	settings := backend.DataSourceInstanceSettings{
		JSONData:                jsonData,
		DecryptedSecureJSONData: secure,
	}
	if s, ok := value["url"].(string); ok {
		settings.URL = s
	}
	return settings
}

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name                  string
		example               string
		settings              backend.DataSourceInstanceSettings
		wantErr               error
		wantURL               string
		wantHTTPMethod        HTTPMethod
		wantSigV4AuthType     SigV4AuthType
		wantSigV4Region       string
		wantSigV4Service      string
		wantSigV4Auth         bool
		wantForwardHeader     bool
		wantConfiguredSecrets SecureJsonDataConfig
	}{
		{
			name:             "default example loads",
			example:          "",
			wantURL:          "https://aps-workspaces.<region>.amazonaws.com/workspaces/<workspace-id>",
			wantHTTPMethod:   HTTPMethodPOST,
			wantSigV4Service: DefaultSigV4Service,
			wantSigV4Auth:    true,
			// wantSigV4Region is empty — Validate would fail; not called here
			// because example URL has "<region>" placeholder, so we don't
			// require SigV4Region validation via LoadConfig for the default
			// example. Instead the test invokes it below and asserts error.
		},
		{
			name:              "workspace iam role",
			example:           "ec2IamRole",
			wantURL:           "https://aps-workspaces.us-east-1.amazonaws.com/workspaces/ws-12345678-1234-1234-1234-123456789012",
			wantHTTPMethod:    HTTPMethodPOST,
			wantSigV4AuthType: SigV4AuthTypeEC2IAMRole,
			wantSigV4Region:   "us-east-1",
			wantSigV4Service:  DefaultSigV4Service,
			wantSigV4Auth:     true,
		},
		{
			name:              "access keys",
			example:           "accessKeys",
			wantURL:           "https://aps-workspaces.us-east-1.amazonaws.com/workspaces/ws-12345678-1234-1234-1234-123456789012",
			wantHTTPMethod:    HTTPMethodPOST,
			wantSigV4AuthType: SigV4AuthTypeKeys,
			wantSigV4Region:   "us-east-1",
			wantSigV4Service:  DefaultSigV4Service,
			wantSigV4Auth:     true,
			wantConfiguredSecrets: SecureJsonDataConfig{
				SecureJsonDataKeySigV4AccessKey,
				SecureJsonDataKeySigV4SecretKey,
			},
		},
		{
			name:              "credentials file",
			example:           "credentialsFile",
			wantURL:           "https://aps-workspaces.us-east-1.amazonaws.com/workspaces/ws-12345678-1234-1234-1234-123456789012",
			wantHTTPMethod:    HTTPMethodPOST,
			wantSigV4AuthType: SigV4AuthTypeCredentials,
			wantSigV4Region:   "us-east-1",
			wantSigV4Service:  DefaultSigV4Service,
			wantSigV4Auth:     true,
		},
		{
			name:              "aws sdk default",
			example:           "awsSdkDefault",
			wantURL:           "https://aps-workspaces.us-east-1.amazonaws.com/workspaces/ws-12345678-1234-1234-1234-123456789012",
			wantHTTPMethod:    HTTPMethodPOST,
			wantSigV4AuthType: SigV4AuthTypeDefault,
			wantSigV4Region:   "us-east-1",
			wantSigV4Service:  DefaultSigV4Service,
			wantSigV4Auth:     true,
		},
		{
			name:              "grafana assume role",
			example:           "grafanaAssumeRole",
			wantURL:           "https://aps-workspaces.us-east-1.amazonaws.com/workspaces/ws-12345678-1234-1234-1234-123456789012",
			wantHTTPMethod:    HTTPMethodPOST,
			wantSigV4AuthType: SigV4AuthTypeGrafanaAssumeRole,
			wantSigV4Region:   "us-east-1",
			wantSigV4Service:  DefaultSigV4Service,
			wantSigV4Auth:     true,
		},
		{
			name:              "cross-account assume role",
			example:           "assumeRoleCrossAccount",
			wantURL:           "https://aps-workspaces.us-east-1.amazonaws.com/workspaces/ws-12345678-1234-1234-1234-123456789012",
			wantHTTPMethod:    HTTPMethodPOST,
			wantSigV4AuthType: SigV4AuthTypeKeys,
			wantSigV4Region:   "us-east-1",
			wantSigV4Service:  DefaultSigV4Service,
			wantSigV4Auth:     true,
			wantConfiguredSecrets: SecureJsonDataConfig{
				SecureJsonDataKeySigV4AccessKey,
				SecureJsonDataKeySigV4SecretKey,
			},
		},
		{
			name:              "forward grafana user header",
			example:           "forwardGrafanaUserHeader",
			wantURL:           "https://aps-workspaces.us-east-1.amazonaws.com/workspaces/ws-12345678-1234-1234-1234-123456789012",
			wantHTTPMethod:    HTTPMethodPOST,
			wantSigV4AuthType: SigV4AuthTypeEC2IAMRole,
			wantSigV4Region:   "us-east-1",
			wantSigV4Service:  DefaultSigV4Service,
			wantSigV4Auth:     true,
			wantForwardHeader: true,
		},
		{
			name:              "migrated from prometheus",
			example:           "migratedFromPrometheus",
			wantURL:           "https://aps-workspaces.us-east-1.amazonaws.com/workspaces/ws-12345678-1234-1234-1234-123456789012",
			wantHTTPMethod:    HTTPMethodPOST,
			wantSigV4AuthType: SigV4AuthTypeEC2IAMRole,
			wantSigV4Region:   "us-east-1",
			wantSigV4Service:  DefaultSigV4Service,
			wantSigV4Auth:     true,
		},
		{
			name: "missing URL errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"sigV4AuthType":"ec2_iam_role","sigV4Region":"us-east-1"}`),
			},
			wantErr: errors.New("Prometheus server URL"),
		},
		{
			name: "missing region errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://aps-workspaces.us-east-1.amazonaws.com/workspaces/ws-x",
				JSONData: []byte(`{"sigV4AuthType":"ec2_iam_role"}`),
			},
			wantErr: errors.New("jsonData.sigV4Region is required"),
		},
		{
			name: "invalid httpMethod errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://aps-workspaces.us-east-1.amazonaws.com/workspaces/ws-x",
				JSONData: []byte(`{"sigV4AuthType":"ec2_iam_role","sigV4Region":"us-east-1","httpMethod":"PUT"}`),
			},
			wantErr: errors.New(`invalid httpMethod "PUT"`),
		},
		{
			name: "lowercase httpMethod normalises to uppercase",
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://aps-workspaces.us-east-1.amazonaws.com/workspaces/ws-x",
				JSONData: []byte(`{"sigV4AuthType":"ec2_iam_role","sigV4Region":"us-east-1","httpMethod":"get"}`),
			},
			wantURL:           "https://aps-workspaces.us-east-1.amazonaws.com/workspaces/ws-x",
			wantHTTPMethod:    HTTPMethodGET,
			wantSigV4AuthType: SigV4AuthTypeEC2IAMRole,
			wantSigV4Region:   "us-east-1",
			wantSigV4Service:  DefaultSigV4Service,
			wantSigV4Auth:     true,
		},
		{
			name: "empty httpMethod defaults to POST",
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://aps-workspaces.us-east-1.amazonaws.com/workspaces/ws-x",
				JSONData: []byte(`{"sigV4AuthType":"ec2_iam_role","sigV4Region":"us-east-1"}`),
			},
			wantURL:           "https://aps-workspaces.us-east-1.amazonaws.com/workspaces/ws-x",
			wantHTTPMethod:    HTTPMethodPOST,
			wantSigV4AuthType: SigV4AuthTypeEC2IAMRole,
			wantSigV4Region:   "us-east-1",
			wantSigV4Service:  DefaultSigV4Service,
			wantSigV4Auth:     true,
		},
		{
			name: "empty sigv4Service defaults to aps",
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://aps-workspaces.us-east-1.amazonaws.com/workspaces/ws-x",
				JSONData: []byte(`{"sigV4AuthType":"ec2_iam_role","sigV4Region":"us-east-1"}`),
			},
			wantURL:           "https://aps-workspaces.us-east-1.amazonaws.com/workspaces/ws-x",
			wantHTTPMethod:    HTTPMethodPOST,
			wantSigV4AuthType: SigV4AuthTypeEC2IAMRole,
			wantSigV4Region:   "us-east-1",
			wantSigV4Service:  DefaultSigV4Service,
			wantSigV4Auth:     true,
		},
		{
			name: "explicit sigv4Service preserved",
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://aps-workspaces.us-east-1.amazonaws.com/workspaces/ws-x",
				JSONData: []byte(`{"sigV4AuthType":"ec2_iam_role","sigV4Region":"us-east-1","sigv4Service":"custom"}`),
			},
			wantURL:           "https://aps-workspaces.us-east-1.amazonaws.com/workspaces/ws-x",
			wantHTTPMethod:    HTTPMethodPOST,
			wantSigV4AuthType: SigV4AuthTypeEC2IAMRole,
			wantSigV4Region:   "us-east-1",
			wantSigV4Service:  "custom",
			wantSigV4Auth:     true,
		},
		{
			name: "sigV4Auth forced true even when stored false",
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://aps-workspaces.us-east-1.amazonaws.com/workspaces/ws-x",
				JSONData: []byte(`{"sigV4Auth":false,"sigV4AuthType":"ec2_iam_role","sigV4Region":"us-east-1"}`),
			},
			wantURL:           "https://aps-workspaces.us-east-1.amazonaws.com/workspaces/ws-x",
			wantHTTPMethod:    HTTPMethodPOST,
			wantSigV4AuthType: SigV4AuthTypeEC2IAMRole,
			wantSigV4Region:   "us-east-1",
			wantSigV4Service:  DefaultSigV4Service,
			wantSigV4Auth:     true,
		},
		{
			name: "keys auth without access key errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://aps-workspaces.us-east-1.amazonaws.com/workspaces/ws-x",
				JSONData: []byte(`{"sigV4AuthType":"keys","sigV4Region":"us-east-1"}`),
			},
			wantErr: errors.New("secureJsonData.sigV4AccessKey is required"),
		},
		{
			name: "keys auth without secret key errors",
			settings: backend.DataSourceInstanceSettings{
				URL:                     "https://aps-workspaces.us-east-1.amazonaws.com/workspaces/ws-x",
				JSONData:                []byte(`{"sigV4AuthType":"keys","sigV4Region":"us-east-1"}`),
				DecryptedSecureJSONData: map[string]string{"sigV4AccessKey": "AKIA"},
			},
			wantErr: errors.New("secureJsonData.sigV4SecretKey is required"),
		},
		{
			name: "unknown authType errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://aps-workspaces.us-east-1.amazonaws.com/workspaces/ws-x",
				JSONData: []byte(`{"sigV4AuthType":"nonsense","sigV4Region":"us-east-1"}`),
			},
			wantErr: errors.New(`unknown sigV4AuthType "nonsense"`),
		},
		{
			name: "legacy arn authType accepted",
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://aps-workspaces.us-east-1.amazonaws.com/workspaces/ws-x",
				JSONData: []byte(`{"sigV4AuthType":"arn","sigV4Region":"us-east-1"}`),
			},
			wantURL:           "https://aps-workspaces.us-east-1.amazonaws.com/workspaces/ws-x",
			wantHTTPMethod:    HTTPMethodPOST,
			wantSigV4AuthType: SigV4AuthTypeARN,
			wantSigV4Region:   "us-east-1",
			wantSigV4Service:  DefaultSigV4Service,
			wantSigV4Auth:     true,
		},
		{
			name: "invalid jsonData errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://aps-workspaces.us-east-1.amazonaws.com/workspaces/ws-x",
				JSONData: []byte(`{`),
			},
			wantErr: errors.New("parse jsonData"),
		},
		{
			name: "negative timeout errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://aps-workspaces.us-east-1.amazonaws.com/workspaces/ws-x",
				JSONData: []byte(`{"sigV4AuthType":"ec2_iam_role","sigV4Region":"us-east-1","timeout":-5}`),
			},
			wantErr: errors.New("timeout must be non-negative"),
		},
		{
			name: "negative seriesLimit errors",
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://aps-workspaces.us-east-1.amazonaws.com/workspaces/ws-x",
				JSONData: []byte(`{"sigV4AuthType":"ec2_iam_role","sigV4Region":"us-east-1","seriesLimit":-1}`),
			},
			wantErr: errors.New("seriesLimit must be non-negative"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings := tt.settings
			// If this test case is example-driven, load the example.
			if tt.example != "" {
				if _, ok := SettingsExamples().Examples[tt.example]; ok {
					settings = settingsFromExample(t, tt.example)
				}
			} else if tt.example == "" && tt.settings.JSONData == nil && tt.wantErr == nil {
				// The default (empty-key) example is a special-case: it has
				// a placeholder URL with "<region>" and empty sigV4Region,
				// so Validate would reject it. Test the default example
				// separately in TestDefaultExampleShape.
				t.Skip("default example is validated separately")
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

			if tt.wantURL != "" && cfg.URL != tt.wantURL {
				t.Errorf("URL = %q, want %q", cfg.URL, tt.wantURL)
			}
			if tt.wantHTTPMethod != "" && cfg.HTTPMethod != tt.wantHTTPMethod {
				t.Errorf("HTTPMethod = %q, want %q", cfg.HTTPMethod, tt.wantHTTPMethod)
			}
			if tt.wantSigV4AuthType != "" && cfg.SigV4AuthType != tt.wantSigV4AuthType {
				t.Errorf("SigV4AuthType = %q, want %q", cfg.SigV4AuthType, tt.wantSigV4AuthType)
			}
			if tt.wantSigV4Region != "" && cfg.SigV4Region != tt.wantSigV4Region {
				t.Errorf("SigV4Region = %q, want %q", cfg.SigV4Region, tt.wantSigV4Region)
			}
			if tt.wantSigV4Service != "" && cfg.SigV4Service != tt.wantSigV4Service {
				t.Errorf("SigV4Service = %q, want %q", cfg.SigV4Service, tt.wantSigV4Service)
			}
			if cfg.SigV4Auth != tt.wantSigV4Auth {
				t.Errorf("SigV4Auth = %v, want %v", cfg.SigV4Auth, tt.wantSigV4Auth)
			}
			if cfg.ForwardGrafanaUserHeader != tt.wantForwardHeader {
				t.Errorf("ForwardGrafanaUserHeader = %v, want %v", cfg.ForwardGrafanaUserHeader, tt.wantForwardHeader)
			}
			if tt.wantConfiguredSecrets != nil {
				gotKeys := SecureJsonDataConfig{}
				for _, key := range SecureJsonDataKeys {
					if v, ok := cfg.DecryptedSecureJSONData[key]; ok && v != "" {
						gotKeys = append(gotKeys, key)
					}
				}
				if !reflect.DeepEqual(gotKeys, tt.wantConfiguredSecrets) {
					t.Errorf("configured secure keys = %v, want %v", gotKeys, tt.wantConfiguredSecrets)
				}
			}
		})
	}
}

// TestDefaultExampleShape guards the shape of the empty-string default
// example: it must exist, must carry the SigV4 defaults (sigV4Auth=true,
// sigv4Service=aps, httpMethod=POST), and every secret key must be present
// (even if empty) so downstream generators produce a complete
// `secureJsonData` template. Validate() is expected to reject the default
// because sigV4Region and sigV4AuthType are empty — that's the whole point
// of the default: force the operator to fill them in.
func TestDefaultExampleShape(t *testing.T) {
	settings := settingsFromExample(t, "")
	cfg := Config{
		URL:                     settings.URL,
		DecryptedSecureJSONData: map[SecureJsonDataKey]string{},
	}
	if err := json.Unmarshal(settings.JSONData, &cfg); err != nil {
		t.Fatalf("unmarshal default example jsonData: %v", err)
	}
	cfg.ApplyDefaults()

	if !cfg.SigV4Auth {
		t.Errorf("default example: SigV4Auth = false, want true")
	}
	if cfg.SigV4Service != DefaultSigV4Service {
		t.Errorf("default example: SigV4Service = %q, want %q", cfg.SigV4Service, DefaultSigV4Service)
	}
	if cfg.HTTPMethod != HTTPMethodPOST {
		t.Errorf("default example: HTTPMethod = %q, want %q", cfg.HTTPMethod, HTTPMethodPOST)
	}

	// Validate should reject it (no region, no authType).
	if err := cfg.Validate(); err == nil {
		t.Errorf("default example: Validate() returned nil, want error (missing sigV4Region)")
	}

	// Both secret keys must be present in the settings example (as empty
	// placeholders) so the conformance check on example completeness passes.
	for _, key := range SecureJsonDataKeys {
		if _, ok := settings.DecryptedSecureJSONData[string(key)]; !ok {
			t.Errorf("default example: secureJsonData missing key %q", key)
		}
	}
}

func TestApplyDefaults(t *testing.T) {
	tests := []struct {
		name           string
		in             Config
		wantHTTPMethod HTTPMethod
		wantSigV4Auth  bool
		wantService    string
	}{
		{"empty config defaults", Config{}, HTTPMethodPOST, true, DefaultSigV4Service},
		{"existing POST preserved", Config{HTTPMethod: HTTPMethodPOST}, HTTPMethodPOST, true, DefaultSigV4Service},
		{"GET preserved", Config{HTTPMethod: HTTPMethodGET}, HTTPMethodGET, true, DefaultSigV4Service},
		{"lowercase get normalises to GET", Config{HTTPMethod: "get"}, HTTPMethodGET, true, DefaultSigV4Service},
		{"whitespace stripped and uppercased", Config{HTTPMethod: "  post  "}, HTTPMethodPOST, true, DefaultSigV4Service},
		{"explicit service preserved", Config{SigV4Service: "custom"}, HTTPMethodPOST, true, "custom"},
		{"whitespace-only service defaults", Config{SigV4Service: "   "}, HTTPMethodPOST, true, DefaultSigV4Service},
		{"stored sigV4Auth=false overwritten", Config{SigV4Auth: false}, HTTPMethodPOST, true, DefaultSigV4Service},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.in
			got.ApplyDefaults()
			if got.HTTPMethod != tt.wantHTTPMethod {
				t.Errorf("HTTPMethod = %q, want %q", got.HTTPMethod, tt.wantHTTPMethod)
			}
			if got.SigV4Auth != tt.wantSigV4Auth {
				t.Errorf("SigV4Auth = %v, want %v", got.SigV4Auth, tt.wantSigV4Auth)
			}
			if got.SigV4Service != tt.wantService {
				t.Errorf("SigV4Service = %q, want %q", got.SigV4Service, tt.wantService)
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
			name: "minimal happy path",
			cfg: Config{
				URL:           "https://aps-workspaces.us-east-1.amazonaws.com/workspaces/ws-x",
				HTTPMethod:    HTTPMethodPOST,
				SigV4AuthType: SigV4AuthTypeEC2IAMRole,
				SigV4Region:   "us-east-1",
			},
		},
		{
			name:    "missing URL",
			cfg:     Config{HTTPMethod: HTTPMethodPOST, SigV4AuthType: SigV4AuthTypeEC2IAMRole, SigV4Region: "us-east-1"},
			wantErr: "Prometheus server URL (root.url) is required",
		},
		{
			name: "missing region",
			cfg: Config{
				URL:           "https://aps-workspaces.us-east-1.amazonaws.com/workspaces/ws-x",
				HTTPMethod:    HTTPMethodPOST,
				SigV4AuthType: SigV4AuthTypeEC2IAMRole,
			},
			wantErr: "jsonData.sigV4Region is required",
		},
		{
			name: "invalid method",
			cfg: Config{
				URL:           "https://aps-workspaces.us-east-1.amazonaws.com/workspaces/ws-x",
				HTTPMethod:    "PUT",
				SigV4AuthType: SigV4AuthTypeEC2IAMRole,
				SigV4Region:   "us-east-1",
			},
			wantErr: `invalid httpMethod "PUT"`,
		},
		{
			name: "keys without access key",
			cfg: Config{
				URL:           "https://aps-workspaces.us-east-1.amazonaws.com/workspaces/ws-x",
				HTTPMethod:    HTTPMethodPOST,
				SigV4AuthType: SigV4AuthTypeKeys,
				SigV4Region:   "us-east-1",
			},
			wantErr: "sigV4AccessKey is required",
		},
		{
			name: "keys with both secrets",
			cfg: Config{
				URL:           "https://aps-workspaces.us-east-1.amazonaws.com/workspaces/ws-x",
				HTTPMethod:    HTTPMethodPOST,
				SigV4AuthType: SigV4AuthTypeKeys,
				SigV4Region:   "us-east-1",
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{
					SecureJsonDataKeySigV4AccessKey: "AKIA",
					SecureJsonDataKeySigV4SecretKey: "secret",
				},
			},
		},
		{
			name: "unknown authType",
			cfg: Config{
				URL:           "https://aps-workspaces.us-east-1.amazonaws.com/workspaces/ws-x",
				HTTPMethod:    HTTPMethodPOST,
				SigV4AuthType: "nonsense",
				SigV4Region:   "us-east-1",
			},
			wantErr: `unknown sigV4AuthType "nonsense"`,
		},
		{
			name: "empty authType is tolerated (unauth chain)",
			cfg: Config{
				URL:         "https://aps-workspaces.us-east-1.amazonaws.com/workspaces/ws-x",
				HTTPMethod:  HTTPMethodPOST,
				SigV4Region: "us-east-1",
			},
		},
		{
			name: "legacy arn is accepted",
			cfg: Config{
				URL:           "https://aps-workspaces.us-east-1.amazonaws.com/workspaces/ws-x",
				HTTPMethod:    HTTPMethodPOST,
				SigV4AuthType: SigV4AuthTypeARN,
				SigV4Region:   "us-east-1",
			},
		},
		{
			name: "negative timeout",
			cfg: Config{
				URL:           "https://aps-workspaces.us-east-1.amazonaws.com/workspaces/ws-x",
				HTTPMethod:    HTTPMethodPOST,
				SigV4AuthType: SigV4AuthTypeEC2IAMRole,
				SigV4Region:   "us-east-1",
				Timeout:       -1,
			},
			wantErr: "timeout must be non-negative",
		},
		{
			name: "negative seriesLimit",
			cfg: Config{
				URL:           "https://aps-workspaces.us-east-1.amazonaws.com/workspaces/ws-x",
				HTTPMethod:    HTTPMethodPOST,
				SigV4AuthType: SigV4AuthTypeEC2IAMRole,
				SigV4Region:   "us-east-1",
				SeriesLimit:   -5,
			},
			wantErr: "seriesLimit must be non-negative",
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
