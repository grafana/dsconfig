package infinitydatasource

import (
	_ "embed"

	"k8s.io/kube-openapi/pkg/spec3"

	"github.com/grafana/dsconfig/dsconfig"
	sdkschema "github.com/grafana/grafana-plugin-sdk-go/experimental/pluginschema"
)

//go:generate go test -generateArtifacts -run TestSchemaConformance ./...

// TargetAPIVersion is the API version this schema applies to.
const TargetAPIVersion = dsconfig.TargetAPIVersion

// configSchemaJSON is the declarative dsconfig schema — the single source
// of truth for the Grafana Infinity datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the declarative dsconfig
// schema for the Grafana Infinity datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the full k8s-style SDK plugin schema for the Grafana
// Infinity datasource: the settings spec derived from dsconfig.json, the
// secure values, and example configurations, stamped with
// TargetAPIVersion. Grafana's datasource API server serves this bundle as
// {TargetAPIVersion}.json.
func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

// SettingsExamples returns k8s-style example configurations for the
// Grafana Infinity datasource, covering the default configuration plus
// each supported authentication method and a few common
// connection/security variants. Each example is a full instance-settings
// object with plugin config nested under jsonData and the relevant
// write-only secrets under secureJsonData (placeholder values — replace
// them with real secrets).
func SettingsExamples() *sdkschema.SettingsExamples {
	return &sdkschema.SettingsExamples{
		Examples: map[string]*spec3.Example{
			"": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Default configuration",
					Description: "The defaults a new datasource starts with: no auth, no base URL, proxy_type='env', timeoutInSeconds=60, unsecuredQueryHandling='warn'. Fully functional — every query is expected to carry its own URL.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"auth_method":            string(AuthTypeNone),
							"proxy_type":             string(ProxyTypeEnv),
							"timeoutInSeconds":       60,
							"unsecuredQueryHandling": string(UnsecuredQueryHandlingWarn),
							"apiKeyType":             string(APIKeyTypeHeader),
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyBearerToken): "",
						},
					},
				},
			},
			"basicAuth": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Basic authentication",
					Description: "HTTP Basic auth. root.basicAuth=true and root.basicAuthUser live at datasource root; only basicAuthPassword is secret. auth_method='basicAuth' keeps the modern discriminator in sync with the legacy root.basicAuth flag.",
					Value: map[string]any{
						"url":           "https://api.example.com",
						"basicAuth":     true,
						"basicAuthUser": "grafana",
						"jsonData": map[string]any{
							"auth_method": string(AuthTypeBasic),
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyBasicAuthPassword): "REPLACE_WITH_PASSWORD",
						},
					},
				},
			},
			"bearerToken": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Bearer token",
					Description: "Static bearer token added as an Authorization header. The token itself is write-only in secureJsonData.bearerToken.",
					Value: map[string]any{
						"url": "https://api.example.com",
						"jsonData": map[string]any{
							"auth_method": string(AuthTypeBearerToken),
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyBearerToken): "REPLACE_WITH_TOKEN",
						},
					},
				},
			},
			"apiKeyHeader": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "API key (header)",
					Description: "API key sent as an HTTP request header (apiKeyType='header'). apiKeyKey is the header name, apiKeyValue is the secret value.",
					Value: map[string]any{
						"url": "https://api.example.com",
						"jsonData": map[string]any{
							"auth_method": string(AuthTypeAPIKey),
							"apiKeyKey":   "X-API-Key",
							"apiKeyType":  string(APIKeyTypeHeader),
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAPIKeyValue): "REPLACE_WITH_API_KEY",
						},
					},
				},
			},
			"apiKeyQuery": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "API key (query param)",
					Description: "API key sent as a URL query-string parameter (apiKeyType='query').",
					Value: map[string]any{
						"url": "https://api.example.com",
						"jsonData": map[string]any{
							"auth_method": string(AuthTypeAPIKey),
							"apiKeyKey":   "apiKey",
							"apiKeyType":  string(APIKeyTypeQuery),
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAPIKeyValue): "REPLACE_WITH_API_KEY",
						},
					},
				},
			},
			"digestAuth": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Digest authentication",
					Description: "HTTP Digest auth (RFC 7616). Same credential shape as basicAuth but discriminated by auth_method='digestAuth'; root.basicAuth stays false.",
					Value: map[string]any{
						"url":           "https://api.example.com",
						"basicAuthUser": "grafana",
						"jsonData": map[string]any{
							"auth_method": string(AuthTypeDigest),
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyBasicAuthPassword): "REPLACE_WITH_PASSWORD",
						},
					},
				},
			},
			"forwardOAuth": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Forward OAuth identity",
					Description: "Forward the signed-in Grafana user's upstream OAuth identity to the API. No secret is stored on the datasource; jsonData.oauthPassThru mirrors root.basicAuth's discriminator role.",
					Value: map[string]any{
						"url": "https://api.example.com",
						"jsonData": map[string]any{
							"auth_method":   string(AuthTypeForwardOAuth),
							"oauthPassThru": true,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyBearerToken): "",
						},
					},
				},
			},
			"oauth2ClientCredentials": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "OAuth2 client credentials",
					Description: "OAuth2 client-credentials grant. jsonData.oauth2.oauth2_type='client_credentials', client_id + token_url in jsonData, client secret in secureJsonData.oauth2ClientSecret. authStyle=0 = auto-detect where the client_id/secret are sent.",
					Value: map[string]any{
						"url": "https://api.example.com",
						"jsonData": map[string]any{
							"auth_method": string(AuthTypeOAuth2),
							"allowedHosts": []any{
								"https://api.example.com",
							},
							"oauth2": map[string]any{
								"oauth2_type": string(OAuth2TypeClientCredentials),
								"client_id":   "REPLACE_WITH_CLIENT_ID",
								"token_url":   "https://auth.example.com/oauth/token",
								"scopes":      []any{"read"},
								"authStyle":   0,
							},
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyOAuth2ClientSecret): "REPLACE_WITH_CLIENT_SECRET",
						},
					},
				},
			},
			"oauth2JWT": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "OAuth2 JWT (service account)",
					Description: "OAuth2 JWT grant (2-legged, e.g. Google service account). jsonData.oauth2.oauth2_type='jwt' with email + token_url + optional subject; the RSA private key PEM lives in secureJsonData.oauth2JWTPrivateKey.",
					Value: map[string]any{
						"url": "https://monitoring.googleapis.com",
						"jsonData": map[string]any{
							"auth_method": string(AuthTypeOAuth2),
							"allowedHosts": []any{
								"https://monitoring.googleapis.com",
							},
							"oauth2": map[string]any{
								"oauth2_type": string(OAuth2TypeJWT),
								"email":       "sa@project.iam.gserviceaccount.com",
								"token_url":   "https://oauth2.googleapis.com/token",
								"scopes":      []any{"https://www.googleapis.com/auth/monitoring.read"},
							},
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyOAuth2JWTPrivateKey): "-----BEGIN RSA PRIVATE KEY-----\n...\n-----END RSA PRIVATE KEY-----",
						},
					},
				},
			},
			"awsSigV4": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "AWS SigV4 (access key + secret)",
					Description: "AWS Signature v4 signing with static access-key + secret-key credentials. jsonData.aws.authType='keys'; the credentials live in secureJsonData. jsonData.aws.service is the AWS service name used for signing (e.g. 'monitoring' for CloudWatch).",
					Value: map[string]any{
						"url": "https://monitoring.us-east-1.amazonaws.com",
						"jsonData": map[string]any{
							"auth_method": string(AuthTypeAWS),
							"aws": map[string]any{
								"authType": string(AWSAuthTypeKeys),
								"region":   "us-east-1",
								"service":  "monitoring",
							},
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAWSAccessKey): "AKIAIOSFODNN7EXAMPLE",
							string(SecureJsonDataKeyAWSSecretKey): "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
						},
					},
				},
			},
			"azureBlob": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Azure Blob storage",
					Description: "Read data from Azure Blob storage. jsonData.azureBlobCloudType picks the cloud environment; the backend derives azureBlobAccountUrl from it on load (pkg/models/settings.go:406-415). No base URL, and jsonData.allowedHosts is not required for azureBlob auth (pkg/models/settings.go:169-177).",
					Value: map[string]any{
						"jsonData": map[string]any{
							"auth_method":          string(AuthTypeAzureBlob),
							"azureBlobCloudType":   string(AzureBlobCloudTypeAzureCloud),
							"azureBlobAccountName": "myaccount",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAzureBlobAccountKey): "REPLACE_WITH_ACCOUNT_KEY",
						},
					},
				},
			},
			"tlsMutualAuth": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "TLS mutual auth (mTLS)",
					Description: "Endpoint requiring mTLS. jsonData.tlsAuth=true triggers client-cert authentication; serverName + client cert + client key are all mandatory when tlsAuth is enabled.",
					Value: map[string]any{
						"url": "https://api.internal.corp",
						"jsonData": map[string]any{
							"auth_method": string(AuthTypeNone),
							"tlsAuth":     true,
							"serverName":  "api.internal.corp",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyTLSClientCert): "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----",
							string(SecureJsonDataKeyTLSClientKey):  "-----BEGIN RSA PRIVATE KEY-----\n...\n-----END RSA PRIVATE KEY-----",
						},
					},
				},
			},
			"tlsCustomCA": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Self-signed / custom CA",
					Description: "Endpoint behind a private CA. jsonData.tlsAuthWithCACert=true tells the SDK to verify the server certificate against the PEM in secureJsonData.tlsCACert.",
					Value: map[string]any{
						"url": "https://api.internal.corp",
						"jsonData": map[string]any{
							"auth_method":       string(AuthTypeNone),
							"tlsAuthWithCACert": true,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyTLSCACert): "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----",
						},
					},
				},
			},
			"customProxy": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Custom HTTP proxy",
					Description: "Route outgoing requests through a custom proxy URL. proxy_type='url' turns on proxy_url + optional proxy_username / proxyUserPassword. Use with caution — RFC 2396 discourages passing credentials in URL userinfo.",
					Value: map[string]any{
						"url": "https://api.example.com",
						"jsonData": map[string]any{
							"auth_method":    string(AuthTypeNone),
							"proxy_type":     string(ProxyTypeURL),
							"proxy_url":      "https://proxy.internal.corp:3128",
							"proxy_username": "grafana",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyProxyUserPassword): "REPLACE_WITH_PROXY_PASSWORD",
						},
					},
				},
			},
			"customHeadersAndQueryParams": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Custom headers and URL query params",
					Description: "Datasource with a custom HTTP header and a custom URL query param configured through the shared SecureFieldsEditor. Each editor row writes a jsonData.httpHeaderName<N>/secureJsonData.httpHeaderValue<N> or jsonData.secureQueryName<N>/secureJsonData.secureQueryValue<N> pair (src/components/config/SecureFieldsEditor.tsx:98-113).",
					Value: map[string]any{
						"url": "https://api.example.com",
						"jsonData": map[string]any{
							"auth_method":      string(AuthTypeNone),
							"httpHeaderName1":  "X-Tenant",
							"secureQueryName1": "trace",
						},
						"secureJsonData": map[string]any{
							"httpHeaderValue1":  "REPLACE_WITH_TENANT",
							"secureQueryValue1": "REPLACE_WITH_TRACE_ID",
						},
					},
				},
			},
			"referenceData": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Reference (inline) data",
					Description: "Datasource-level named inline datasets that queries can reference via source='reference'. Each entry is { name, data }; the data string is parsed by the query editor per query.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"auth_method": string(AuthTypeNone),
							"refData": []any{
								map[string]any{
									"name": "countries",
									"data": "code,name\nUS,United States\nDE,Germany",
								},
							},
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyBearerToken): "",
						},
					},
				},
			},
			"customHealthCheck": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Custom health check URL",
					Description: "Override the plugin's default health check with an HTTP GET against jsonData.customHealthCheckUrl.",
					Value: map[string]any{
						"url": "https://api.example.com",
						"jsonData": map[string]any{
							"auth_method":              string(AuthTypeNone),
							"customHealthCheckEnabled": true,
							"customHealthCheckUrl":     "https://api.example.com/health",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyBearerToken): "",
						},
					},
				},
			},
			"legacyBasicAuthWithoutMethod": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Legacy: root.basicAuth without auth_method",
					Description: "Pre-auth_method datasource: only root.basicAuth is set. LoadSettings back-fills the effective auth method to 'basicAuth' at load time (pkg/models/settings.go:393-401).",
					Value: map[string]any{
						"url":           "https://api.example.com",
						"basicAuth":     true,
						"basicAuthUser": "grafana",
						"jsonData":      map[string]any{},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyBasicAuthPassword): "REPLACE_WITH_PASSWORD",
						},
					},
				},
			},
		},
	}
}
