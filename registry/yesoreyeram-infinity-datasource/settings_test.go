package infinitydatasource

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// settingsFromExample converts a SettingsExamples entry (a full instance
// settings object with root fields, jsonData, and secureJsonData) into
// the backend.DataSourceInstanceSettings shape LoadConfig expects.
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
	if b, ok := value["basicAuth"].(bool); ok {
		settings.BasicAuthEnabled = b
	}
	if s, ok := value["basicAuthUser"].(string); ok {
		settings.BasicAuthUser = s
	}
	return settings
}

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name             string
		example          string
		useSettings      bool
		settings         backend.DataSourceInstanceSettings
		wantErr          error
		wantAuthMethod   AuthType
		wantURL          string
		wantProxyType    ProxyType
		wantTimeout      int64
		wantUnsecured    UnsecuredQueryHandlingMode
		wantAPIKeyType   APIKeyType
		wantOAuth2Type   OAuth2Type
		wantAzureCloud   AzureBlobCloudType
		wantAzureBlobURL string
		wantSecureKeys   SecureJsonDataConfig
		wantHeaderCount  int
		wantQueryCount   int
	}{
		{
			name:           "default example loads",
			example:        "",
			wantAuthMethod: AuthTypeNone,
			wantProxyType:  ProxyTypeEnv,
			wantTimeout:    60,
			wantUnsecured:  UnsecuredQueryHandlingWarn,
			wantAPIKeyType: APIKeyTypeHeader,
		},
		{
			name:           "basic auth",
			example:        "basicAuth",
			wantAuthMethod: AuthTypeBasic,
			wantURL:        "https://api.example.com",
			wantProxyType:  ProxyTypeEnv,
			wantTimeout:    60,
			wantUnsecured:  UnsecuredQueryHandlingWarn,
			wantAPIKeyType: APIKeyTypeHeader,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyBasicAuthPassword},
		},
		{
			name:           "bearer token",
			example:        "bearerToken",
			wantAuthMethod: AuthTypeBearerToken,
			wantURL:        "https://api.example.com",
			wantProxyType:  ProxyTypeEnv,
			wantTimeout:    60,
			wantUnsecured:  UnsecuredQueryHandlingWarn,
			wantAPIKeyType: APIKeyTypeHeader,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyBearerToken},
		},
		{
			name:           "api key header",
			example:        "apiKeyHeader",
			wantAuthMethod: AuthTypeAPIKey,
			wantURL:        "https://api.example.com",
			wantAPIKeyType: APIKeyTypeHeader,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyAPIKeyValue},
		},
		{
			name:           "api key query",
			example:        "apiKeyQuery",
			wantAuthMethod: AuthTypeAPIKey,
			wantURL:        "https://api.example.com",
			wantAPIKeyType: APIKeyTypeQuery,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyAPIKeyValue},
		},
		{
			name:           "digest auth",
			example:        "digestAuth",
			wantAuthMethod: AuthTypeDigest,
			wantURL:        "https://api.example.com",
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyBasicAuthPassword},
		},
		{
			name:           "forward oauth identity",
			example:        "forwardOAuth",
			wantAuthMethod: AuthTypeForwardOAuth,
			wantURL:        "https://api.example.com",
		},
		{
			name:           "oauth2 client credentials",
			example:        "oauth2ClientCredentials",
			wantAuthMethod: AuthTypeOAuth2,
			wantOAuth2Type: OAuth2TypeClientCredentials,
			wantURL:        "https://api.example.com",
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyOAuth2ClientSecret},
		},
		{
			name:           "oauth2 jwt",
			example:        "oauth2JWT",
			wantAuthMethod: AuthTypeOAuth2,
			wantOAuth2Type: OAuth2TypeJWT,
			wantURL:        "https://monitoring.googleapis.com",
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyOAuth2JWTPrivateKey},
		},
		{
			name:           "aws sigv4",
			example:        "awsSigV4",
			wantAuthMethod: AuthTypeAWS,
			wantURL:        "https://monitoring.us-east-1.amazonaws.com",
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyAWSAccessKey, SecureJsonDataKeyAWSSecretKey},
		},
		{
			name:             "azure blob",
			example:          "azureBlob",
			wantAuthMethod:   AuthTypeAzureBlob,
			wantAzureCloud:   AzureBlobCloudTypeAzureCloud,
			wantAzureBlobURL: "https://%s.blob.core.windows.net/",
			wantSecureKeys:   SecureJsonDataConfig{SecureJsonDataKeyAzureBlobAccountKey},
		},
		{
			name:           "tls mutual auth",
			example:        "tlsMutualAuth",
			wantAuthMethod: AuthTypeNone,
			wantURL:        "https://api.internal.corp",
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyTLSClientCert, SecureJsonDataKeyTLSClientKey},
		},
		{
			name:           "tls custom CA",
			example:        "tlsCustomCA",
			wantAuthMethod: AuthTypeNone,
			wantURL:        "https://api.internal.corp",
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyTLSCACert},
		},
		{
			name:           "custom proxy",
			example:        "customProxy",
			wantAuthMethod: AuthTypeNone,
			wantURL:        "https://api.example.com",
			wantProxyType:  ProxyTypeURL,
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyProxyUserPassword},
		},
		{
			name:            "custom headers and query params",
			example:         "customHeadersAndQueryParams",
			wantAuthMethod:  AuthTypeNone,
			wantURL:         "https://api.example.com",
			wantHeaderCount: 1,
			wantQueryCount:  1,
		},
		{
			name:           "reference data",
			example:        "referenceData",
			wantAuthMethod: AuthTypeNone,
			wantProxyType:  ProxyTypeEnv,
			wantTimeout:    60,
		},
		{
			name:           "custom health check",
			example:        "customHealthCheck",
			wantAuthMethod: AuthTypeNone,
			wantURL:        "https://api.example.com",
		},
		{
			name:           "legacy basicAuth without auth_method",
			example:        "legacyBasicAuthWithoutMethod",
			wantAuthMethod: AuthTypeBasic,
			wantURL:        "https://api.example.com",
			wantSecureKeys: SecureJsonDataConfig{SecureJsonDataKeyBasicAuthPassword},
		},
		{
			name:        "IGNORE_URL sentinel normalizes to empty",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				URL:      IgnoreURLSentinel,
				JSONData: []byte(`{"auth_method":"none"}`),
			},
			wantAuthMethod: AuthTypeNone,
			wantURL:        "",
		},
		{
			name:        "empty auth_method with basicAuth=true defaults to basicAuth",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				URL:              "https://api.example.com",
				BasicAuthEnabled: true,
				BasicAuthUser:    "grafana",
				JSONData:         []byte(`{}`),
				DecryptedSecureJSONData: map[string]string{
					"basicAuthPassword": "pw",
				},
			},
			wantAuthMethod: AuthTypeBasic,
			wantURL:        "https://api.example.com",
		},
		{
			name:        "empty auth_method with oauthPassThru=true defaults to oauthPassThru",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://api.example.com",
				JSONData: []byte(`{"oauthPassThru":true}`),
			},
			wantAuthMethod: AuthTypeForwardOAuth,
		},
		{
			name:        "oauth2 with missing oauth2_type defaults to client_credentials",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://api.example.com",
				JSONData: []byte(`{"auth_method":"oauth2","allowedHosts":["https://api.example.com"],"oauth2":{"client_id":"cid","token_url":"https://auth"}}`),
				DecryptedSecureJSONData: map[string]string{
					"oauth2ClientSecret": "s",
				},
			},
			wantAuthMethod: AuthTypeOAuth2,
			wantOAuth2Type: OAuth2TypeClientCredentials,
		},
		{
			name:        "invalid jsonData errors",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://api.example.com",
				JSONData: []byte(`{`),
			},
			wantErr: errors.New("parse jsonData"),
		},
		{
			name:        "basicAuth without password errors",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				URL:              "https://api.example.com",
				BasicAuthEnabled: true,
				BasicAuthUser:    "grafana",
				JSONData:         []byte(`{"auth_method":"basicAuth"}`),
			},
			wantErr: errors.New("basicAuthPassword is required"),
		},
		{
			name:        "apiKey without key errors",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://api.example.com",
				JSONData: []byte(`{"auth_method":"apiKey"}`),
				DecryptedSecureJSONData: map[string]string{
					"apiKeyValue": "v",
				},
			},
			wantErr: errors.New("apiKeyKey is required"),
		},
		{
			name:        "apiKey without value errors",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://api.example.com",
				JSONData: []byte(`{"auth_method":"apiKey","apiKeyKey":"X-K"}`),
			},
			wantErr: errors.New("apiKeyValue is required"),
		},
		{
			name:        "bearerToken without token errors",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://api.example.com",
				JSONData: []byte(`{"auth_method":"bearerToken"}`),
			},
			wantErr: errors.New("bearerToken is required"),
		},
		{
			name:        "aws keys without access key errors",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://api.example.com",
				JSONData: []byte(`{"auth_method":"aws","aws":{"authType":"keys","region":"us-east-1","service":"monitoring"}}`),
				DecryptedSecureJSONData: map[string]string{
					"awsSecretKey": "s",
				},
			},
			wantErr: errors.New("awsAccessKey is required"),
		},
		{
			name:        "aws keys without secret key errors",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://api.example.com",
				JSONData: []byte(`{"auth_method":"aws","aws":{"authType":"keys","region":"us-east-1","service":"monitoring"}}`),
				DecryptedSecureJSONData: map[string]string{
					"awsAccessKey": "a",
				},
			},
			wantErr: errors.New("awsSecretKey is required"),
		},
		{
			name:        "azureBlob without account name errors",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"auth_method":"azureBlob"}`),
				DecryptedSecureJSONData: map[string]string{
					"azureBlobAccountKey": "k",
				},
			},
			wantErr: errors.New("azureBlobAccountName is required"),
		},
		{
			name:        "azureBlob without account key errors",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"auth_method":"azureBlob","azureBlobAccountName":"acct"}`),
			},
			wantErr: errors.New("azureBlobAccountKey is required"),
		},
		{
			name:        "empty URL with bearerToken requires allowedHosts",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"auth_method":"bearerToken"}`),
				DecryptedSecureJSONData: map[string]string{
					"bearerToken": "t",
				},
			},
			wantErr: errors.New("allowedHosts must contain at least one host"),
		},
		{
			name:        "empty URL with bearerToken and allowedHosts is ok",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"auth_method":"bearerToken","allowedHosts":["https://api.example.com"]}`),
				DecryptedSecureJSONData: map[string]string{
					"bearerToken": "t",
				},
			},
			wantAuthMethod: AuthTypeBearerToken,
		},
		{
			name:        "allowedHosts entry without protocol errors",
			useSettings: true,
			settings: backend.DataSourceInstanceSettings{
				URL:      "https://api.example.com",
				JSONData: []byte(`{"auth_method":"none","allowedHosts":["example.com"]}`),
			},
			wantErr: errors.New("invalid url in allowed list"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings := tt.settings
			if !tt.useSettings {
				settings = settingsFromExample(t, tt.example)
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

			if tt.wantAuthMethod != "" && cfg.AuthenticationMethod != tt.wantAuthMethod {
				t.Errorf("AuthenticationMethod = %q, want %q", cfg.AuthenticationMethod, tt.wantAuthMethod)
			}
			if tt.wantURL != "" && cfg.URL != tt.wantURL {
				t.Errorf("URL = %q, want %q", cfg.URL, tt.wantURL)
			}
			if tt.name == "IGNORE_URL sentinel normalizes to empty" && cfg.URL != "" {
				t.Errorf("URL = %q, want empty", cfg.URL)
			}
			if tt.wantProxyType != "" && cfg.ProxyType != tt.wantProxyType {
				t.Errorf("ProxyType = %q, want %q", cfg.ProxyType, tt.wantProxyType)
			}
			if tt.wantTimeout != 0 && cfg.TimeoutInSeconds != tt.wantTimeout {
				t.Errorf("TimeoutInSeconds = %d, want %d", cfg.TimeoutInSeconds, tt.wantTimeout)
			}
			if tt.wantUnsecured != "" && cfg.UnsecuredQueryHandling != tt.wantUnsecured {
				t.Errorf("UnsecuredQueryHandling = %q, want %q", cfg.UnsecuredQueryHandling, tt.wantUnsecured)
			}
			if tt.wantAPIKeyType != "" && cfg.APIKeyType != tt.wantAPIKeyType {
				t.Errorf("APIKeyType = %q, want %q", cfg.APIKeyType, tt.wantAPIKeyType)
			}
			if tt.wantOAuth2Type != "" && cfg.OAuth2Settings.OAuth2Type != tt.wantOAuth2Type {
				t.Errorf("OAuth2Type = %q, want %q", cfg.OAuth2Settings.OAuth2Type, tt.wantOAuth2Type)
			}
			if tt.wantAzureCloud != "" && cfg.AzureBlobCloudType != tt.wantAzureCloud {
				t.Errorf("AzureBlobCloudType = %q, want %q", cfg.AzureBlobCloudType, tt.wantAzureCloud)
			}
			if tt.wantAzureBlobURL != "" && cfg.AzureBlobAccountURL != tt.wantAzureBlobURL {
				t.Errorf("AzureBlobAccountURL = %q, want %q", cfg.AzureBlobAccountURL, tt.wantAzureBlobURL)
			}
			if tt.wantHeaderCount != 0 && len(cfg.CustomHeaders) != tt.wantHeaderCount {
				t.Errorf("CustomHeaders count = %d, want %d (got %v)", len(cfg.CustomHeaders), tt.wantHeaderCount, cfg.CustomHeaders)
			}
			if tt.wantQueryCount != 0 && len(cfg.SecureQueryFields) != tt.wantQueryCount {
				t.Errorf("SecureQueryFields count = %d, want %d (got %v)", len(cfg.SecureQueryFields), tt.wantQueryCount, cfg.SecureQueryFields)
			}
			if tt.wantSecureKeys != nil {
				gotKeys := SecureJsonDataConfig{}
				for _, key := range SecureJsonDataKeys {
					if v, ok := cfg.DecryptedSecureJSONData[key]; ok && v != "" {
						gotKeys = append(gotKeys, key)
					}
				}
				if !reflect.DeepEqual(gotKeys, tt.wantSecureKeys) {
					t.Errorf("configured secure keys = %v, want %v", gotKeys, tt.wantSecureKeys)
				}
			}
		})
	}
}

func TestApplyDefaults(t *testing.T) {
	tests := []struct {
		name string
		in   Config
		want Config
	}{
		{
			name: "empty config gets defaults",
			in:   Config{},
			want: Config{
				AuthenticationMethod:   AuthTypeNone,
				APIKeyType:             APIKeyTypeHeader,
				TimeoutInSeconds:       60,
				ProxyType:              ProxyTypeEnv,
				UnsecuredQueryHandling: UnsecuredQueryHandlingWarn,
			},
		},
		{
			name: "basicAuth flag back-fills auth_method",
			in:   Config{BasicAuth: true},
			want: Config{
				BasicAuth:              true,
				AuthenticationMethod:   AuthTypeBasic,
				APIKeyType:             APIKeyTypeHeader,
				TimeoutInSeconds:       60,
				ProxyType:              ProxyTypeEnv,
				UnsecuredQueryHandling: UnsecuredQueryHandlingWarn,
			},
		},
		{
			name: "oauthPassThru flag back-fills auth_method",
			in:   Config{ForwardOauthIdentity: true},
			want: Config{
				ForwardOauthIdentity:   true,
				AuthenticationMethod:   AuthTypeForwardOAuth,
				APIKeyType:             APIKeyTypeHeader,
				TimeoutInSeconds:       60,
				ProxyType:              ProxyTypeEnv,
				UnsecuredQueryHandling: UnsecuredQueryHandlingWarn,
			},
		},
		{
			name: "oauth2 without oauth2_type gets client_credentials",
			in:   Config{AuthenticationMethod: AuthTypeOAuth2},
			want: Config{
				AuthenticationMethod:   AuthTypeOAuth2,
				OAuth2Settings:         OAuth2Settings{OAuth2Type: OAuth2TypeClientCredentials},
				APIKeyType:             APIKeyTypeHeader,
				TimeoutInSeconds:       60,
				ProxyType:              ProxyTypeEnv,
				UnsecuredQueryHandling: UnsecuredQueryHandlingWarn,
			},
		},
		{
			name: "azureBlob picks default cloud and url template",
			in:   Config{AuthenticationMethod: AuthTypeAzureBlob},
			want: Config{
				AuthenticationMethod:   AuthTypeAzureBlob,
				AzureBlobCloudType:     AzureBlobCloudTypeAzureCloud,
				AzureBlobAccountURL:    "https://%s.blob.core.windows.net/",
				APIKeyType:             APIKeyTypeHeader,
				TimeoutInSeconds:       60,
				ProxyType:              ProxyTypeEnv,
				UnsecuredQueryHandling: UnsecuredQueryHandlingWarn,
			},
		},
		{
			name: "azureBlob us gov picks gov url template",
			in: Config{
				AuthenticationMethod: AuthTypeAzureBlob,
				AzureBlobCloudType:   AzureBlobCloudTypeAzureUSGovernment,
			},
			want: Config{
				AuthenticationMethod:   AuthTypeAzureBlob,
				AzureBlobCloudType:     AzureBlobCloudTypeAzureUSGovernment,
				AzureBlobAccountURL:    "https://%s.blob.core.usgovcloudapi.net/",
				APIKeyType:             APIKeyTypeHeader,
				TimeoutInSeconds:       60,
				ProxyType:              ProxyTypeEnv,
				UnsecuredQueryHandling: UnsecuredQueryHandlingWarn,
			},
		},
		{
			name: "existing values are preserved",
			in: Config{
				AuthenticationMethod:   AuthTypeBearerToken,
				APIKeyType:             APIKeyTypeQuery,
				TimeoutInSeconds:       30,
				ProxyType:              ProxyTypeNone,
				UnsecuredQueryHandling: UnsecuredQueryHandlingDeny,
			},
			want: Config{
				AuthenticationMethod:   AuthTypeBearerToken,
				APIKeyType:             APIKeyTypeQuery,
				TimeoutInSeconds:       30,
				ProxyType:              ProxyTypeNone,
				UnsecuredQueryHandling: UnsecuredQueryHandlingDeny,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.in
			got.ApplyDefaults()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ApplyDefaults:\n got  %#v\n want %#v", got, tt.want)
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
			name: "none auth happy path",
			cfg: Config{
				URL:                  "https://api.example.com",
				AuthenticationMethod: AuthTypeNone,
			},
		},
		{
			name: "basicAuth needs password",
			cfg: Config{
				URL:                  "https://api.example.com",
				BasicAuth:            true,
				AuthenticationMethod: AuthTypeBasic,
			},
			wantErr: "basicAuthPassword is required",
		},
		{
			name: "basicAuth with password ok",
			cfg: Config{
				URL:                  "https://api.example.com",
				BasicAuth:            true,
				AuthenticationMethod: AuthTypeBasic,
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{
					SecureJsonDataKeyBasicAuthPassword: "pw",
				},
			},
		},
		{
			name: "apiKey needs both key and value",
			cfg: Config{
				URL:                  "https://api.example.com",
				AuthenticationMethod: AuthTypeAPIKey,
			},
			wantErr: "apiKeyKey is required",
		},
		{
			name: "bearer needs token",
			cfg: Config{
				URL:                  "https://api.example.com",
				AuthenticationMethod: AuthTypeBearerToken,
			},
			wantErr: "bearerToken is required",
		},
		{
			name: "aws keys needs both secrets",
			cfg: Config{
				URL:                  "https://api.example.com",
				AuthenticationMethod: AuthTypeAWS,
				AWSSettings:          AWSSettings{AuthType: AWSAuthTypeKeys},
			},
			wantErr: "awsAccessKey is required",
		},
		{
			name: "azureBlob needs name and key",
			cfg: Config{
				AuthenticationMethod: AuthTypeAzureBlob,
			},
			wantErr: "azureBlobAccountName is required",
		},
		{
			name: "no url with bearer requires allowedHosts",
			cfg: Config{
				AuthenticationMethod: AuthTypeBearerToken,
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{
					SecureJsonDataKeyBearerToken: "t",
				},
			},
			wantErr: "allowedHosts must contain at least one host",
		},
		{
			name: "no url with bearer and allowedHosts ok",
			cfg: Config{
				AuthenticationMethod: AuthTypeBearerToken,
				AllowedHosts:         []string{"https://api.example.com"},
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{
					SecureJsonDataKeyBearerToken: "t",
				},
			},
		},
		{
			name: "azureBlob without url does NOT need allowedHosts",
			cfg: Config{
				AuthenticationMethod: AuthTypeAzureBlob,
				AzureBlobAccountName: "acct",
				DecryptedSecureJSONData: map[SecureJsonDataKey]string{
					SecureJsonDataKeyAzureBlobAccountKey: "k",
				},
			},
		},
		{
			name: "allowedHosts entry without scheme rejected",
			cfg: Config{
				URL:                  "https://api.example.com",
				AuthenticationMethod: AuthTypeNone,
				AllowedHosts:         []string{"example.com"},
			},
			wantErr: "invalid url in allowed list",
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
