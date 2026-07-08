package azuredevopsdatasource

import (
	"encoding/json"
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
		fromExample       bool
		example           string // schema.go SettingsExamples key
		settings          backend.DataSourceInstanceSettings
		wantErr           string // empty = expect no error; otherwise substring match
		wantURL           string
		wantAuthType      AuthType
		wantProjectsLimit int
		wantUsername      string
		wantSecureKeys    SecureJsonDataConfig
		wantPatToken      string
	}{
		{
			// The default schema example intentionally omits url and has an empty
			// patToken placeholder, so LoadConfig's Validate step rejects it.
			name:        "default example fails validation (missing url + empty patToken)",
			fromExample: true,
			example:     "",
			wantErr:     "url (jsonData.url) is required",
		},
		{
			name:              "personal access token (services)",
			fromExample:       true,
			example:           "patToken",
			wantURL:           "https://dev.azure.com/your-organization",
			wantAuthType:      AuthTypePAT,
			wantProjectsLimit: DefaultProjectsLimit,
			wantSecureKeys:    SecureJsonDataConfig{SecureJsonDataKeyPATToken},
			wantPatToken:      "<your-azure-devops-pat>",
		},
		{
			name:              "azure devops server with username",
			fromExample:       true,
			example:           "azureDevOpsServer",
			wantURL:           "https://azuredevops.example.com/DefaultCollection",
			wantAuthType:      AuthTypePAT,
			wantUsername:      "ado",
			wantProjectsLimit: DefaultProjectsLimit,
			wantSecureKeys:    SecureJsonDataConfig{SecureJsonDataKeyPATToken},
			wantPatToken:      "<your-azure-devops-pat>",
		},
		{
			name: "authType and projectsLimit default when unset",
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"url":"https://dev.azure.com/org"}`),
				DecryptedSecureJSONData: map[string]string{"patToken": "tok"},
			},
			wantURL:           "https://dev.azure.com/org",
			wantAuthType:      AuthTypePAT,
			wantProjectsLimit: DefaultProjectsLimit,
			wantSecureKeys:    SecureJsonDataConfig{SecureJsonDataKeyPATToken},
			wantPatToken:      "tok",
		},
		{
			name: "explicit projectsLimit is preserved",
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"url":"https://dev.azure.com/org","authType":"patToken","projectsLimit":50}`),
				DecryptedSecureJSONData: map[string]string{"patToken": "tok"},
			},
			wantProjectsLimit: 50,
		},
		{
			name: "zero projectsLimit coerced to default",
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"url":"https://dev.azure.com/org","projectsLimit":0}`),
				DecryptedSecureJSONData: map[string]string{"patToken": "tok"},
			},
			wantProjectsLimit: DefaultProjectsLimit,
		},
		{
			name: "negative projectsLimit coerced to default",
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"url":"https://dev.azure.com/org","projectsLimit":-5}`),
				DecryptedSecureJSONData: map[string]string{"patToken": "tok"},
			},
			wantProjectsLimit: DefaultProjectsLimit,
		},
		{
			name: "missing url fails validation",
			settings: backend.DataSourceInstanceSettings{
				JSONData:                []byte(`{"authType":"patToken"}`),
				DecryptedSecureJSONData: map[string]string{"patToken": "tok"},
			},
			wantErr: "url (jsonData.url) is required",
		},
		{
			name: "missing patToken fails validation",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"url":"https://dev.azure.com/org"}`),
			},
			wantErr: "personal access token (secureJsonData.patToken) is required",
		},
		{
			name: "invalid jsonData errors",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{`),
			},
			wantErr: "parse jsonData",
		},
		{
			// Empty settings carry nil JSONData; json.Unmarshal rejects it,
			// mirroring the upstream unconditional json.Unmarshal.
			name:     "empty settings fail with parse error",
			settings: backend.DataSourceInstanceSettings{},
			wantErr:  "parse jsonData",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings := tt.settings
			if tt.fromExample {
				settings = settingsFromExample(t, tt.example)
			}

			cfg, err := LoadConfig(t.Context(), settings)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("LoadConfig: expected error %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
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
			if tt.wantAuthType != "" && cfg.AuthType != tt.wantAuthType {
				t.Errorf("AuthType = %q, want %q", cfg.AuthType, tt.wantAuthType)
			}
			if tt.wantProjectsLimit != 0 && cfg.ProjectsLimit != tt.wantProjectsLimit {
				t.Errorf("ProjectsLimit = %d, want %d", cfg.ProjectsLimit, tt.wantProjectsLimit)
			}
			if tt.wantUsername != "" && cfg.Username != tt.wantUsername {
				t.Errorf("Username = %q, want %q", cfg.Username, tt.wantUsername)
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
			if tt.wantPatToken != "" && cfg.DecryptedSecureJSONData[SecureJsonDataKeyPATToken] != tt.wantPatToken {
				t.Errorf("Secrets[patToken] = %q, want %q", cfg.DecryptedSecureJSONData[SecureJsonDataKeyPATToken], tt.wantPatToken)
			}
		})
	}
}

func TestApplyDefaults(t *testing.T) {
	tests := []struct {
		name              string
		in                Config
		wantAuthType      AuthType
		wantProjectsLimit int
	}{
		{
			name:              "empty config gets patToken + 100",
			in:                Config{},
			wantAuthType:      AuthTypePAT,
			wantProjectsLimit: DefaultProjectsLimit,
		},
		{
			name:              "existing projects limit is preserved",
			in:                Config{ProjectsLimit: 25},
			wantAuthType:      AuthTypePAT,
			wantProjectsLimit: 25,
		},
		{
			name:              "negative projects limit is coerced",
			in:                Config{ProjectsLimit: -1},
			wantAuthType:      AuthTypePAT,
			wantProjectsLimit: DefaultProjectsLimit,
		},
		{
			name:              "existing auth type is preserved",
			in:                Config{AuthType: AuthTypePAT, ProjectsLimit: 100},
			wantAuthType:      AuthTypePAT,
			wantProjectsLimit: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.in
			got.ApplyDefaults()
			if got.AuthType != tt.wantAuthType {
				t.Errorf("AuthType = %q, want %q", got.AuthType, tt.wantAuthType)
			}
			if got.ProjectsLimit != tt.wantProjectsLimit {
				t.Errorf("ProjectsLimit = %d, want %d", got.ProjectsLimit, tt.wantProjectsLimit)
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
			name: "url and patToken present",
			cfg: Config{
				URL:                     "https://dev.azure.com/org",
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyPATToken: "tok"},
			},
		},
		{
			name: "missing url errors",
			cfg: Config{
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{SecureJsonDataKeyPATToken: "tok"},
			},
			wantErr: "url (jsonData.url) is required",
		},
		{
			name:    "missing patToken errors",
			cfg:     Config{URL: "https://dev.azure.com/org"},
			wantErr: "personal access token (secureJsonData.patToken) is required",
		},
		{
			name:    "missing everything reports both",
			cfg:     Config{},
			wantErr: "personal access token (secureJsonData.patToken) is required",
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
