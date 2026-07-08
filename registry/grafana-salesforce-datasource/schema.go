package salesforcedatasource

import (
	_ "embed"

	"k8s.io/kube-openapi/pkg/spec3"

	"github.com/grafana/dsconfig/dsconfig"
	sdkschema "github.com/grafana/grafana-plugin-sdk-go/experimental/pluginschema"
)

//go:generate go test -generateArtifacts -run TestSchemaConformance ./...

// TargetAPIVersion is the API version this schema applies to.
const TargetAPIVersion = dsconfig.TargetAPIVersion

// configSchemaJSON is the declarative dsconfig schema — the single source of
// truth for the Salesforce datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the declarative dsconfig schema
// (single source of truth) for the Salesforce datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the full k8s-style SDK plugin schema for the Salesforce
// datasource: the settings (configuration) spec derived from dsconfig.json, the
// secure values, and example configurations, stamped with TargetAPIVersion.
// Grafana's datasource API server serves this bundle as {TargetAPIVersion}.json.
func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

// SettingsExamples returns k8s-style example configurations for the Salesforce
// datasource, covering the default configuration and each authentication method
// and connection (environment) variant the config editor supports. Each example
// value is a full instance settings object with the plugin configuration nested
// under jsonData and the relevant write-only secrets under secureJsonData
// (obviously-fake placeholder values — replace them with real secrets).
func SettingsExamples() *sdkschema.SettingsExamples {
	return &sdkschema.SettingsExamples{
		Examples: map[string]*spec3.Example{
			"": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Default configuration",
					Description: "The defaults a new datasource starts with: username-password (Credentials) authentication against the production host (https://login.salesforce.com). jsonData.user and the secureJsonData secrets (empty here) must be filled in — pkg/models/settings.go:112-123 requires user, password, clientID and clientSecret for user auth.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType": string(AuthTypeUser),
							"tokenUrl": TokenURLProd,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword):      "",
							string(SecureJsonDataKeySecurityToken): "",
							string(SecureJsonDataKeyClientID):      "",
							string(SecureJsonDataKeyClientSecret):  "",
						},
					},
				},
			},
			"userCredentials": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Username-password (Credentials) auth, Production",
					Description: "OAuth2 password grant against https://login.salesforce.com. jsonData.user plus secureJsonData.password, clientID (consumer key) and clientSecret (consumer secret) are required; securityToken is optional and concatenated onto the password (pkg/plugin/client.go:104).",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType": string(AuthTypeUser),
							"user":     "user@example.com",
							"tokenUrl": TokenURLProd,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword):      "<your-salesforce-password>",
							string(SecureJsonDataKeySecurityToken): "<your-security-token>",
							string(SecureJsonDataKeyClientID):      "<your-connected-app-consumer-key>",
							string(SecureJsonDataKeyClientSecret):  "<your-connected-app-consumer-secret>",
						},
					},
				},
			},
			"userCredentialsSandbox": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Username-password (Credentials) auth, SandBox",
					Description: "Same as the Credentials example but pointed at the sandbox host (https://test.salesforce.com) via jsonData.tokenUrl.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType": string(AuthTypeUser),
							"user":     "user@example.com.sandbox",
							"tokenUrl": TokenURLSandbox,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword):      "<your-salesforce-password>",
							string(SecureJsonDataKeySecurityToken): "<your-security-token>",
							string(SecureJsonDataKeyClientID):      "<your-connected-app-consumer-key>",
							string(SecureJsonDataKeyClientSecret):  "<your-connected-app-consumer-secret>",
						},
					},
				},
			},
			"jwt": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "JWT bearer auth, Production",
					Description: "OAuth2 JWT bearer grant against https://login.salesforce.com. secureJsonData.cert and privateKey are validated (pkg/models/settings.go:103-110); secureJsonData.clientID (JWT issuer) and jsonData.user (JWT subject) are also required for the token request to succeed (pkg/plugin/client.go:96, pkg/jwt/jwt.go:57-64).",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType": string(AuthTypeJWT),
							"user":     "user@example.com",
							"tokenUrl": TokenURLProd,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyClientID):   "<your-connected-app-consumer-key>",
							string(SecureJsonDataKeyCert):       "-----BEGIN CERTIFICATE-----\n<redacted>\n-----END CERTIFICATE-----",
							string(SecureJsonDataKeyPrivateKey): "-----BEGIN RSA PRIVATE KEY-----\n<redacted>\n-----END RSA PRIVATE KEY-----",
						},
					},
				},
			},
			"jwtSandbox": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "JWT bearer auth, SandBox",
					Description: "Same as the JWT example but pointed at the sandbox host (https://test.salesforce.com) via jsonData.tokenUrl.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType": string(AuthTypeJWT),
							"user":     "user@example.com.sandbox",
							"tokenUrl": TokenURLSandbox,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyClientID):   "<your-connected-app-consumer-key>",
							string(SecureJsonDataKeyCert):       "-----BEGIN CERTIFICATE-----\n<redacted>\n-----END CERTIFICATE-----",
							string(SecureJsonDataKeyPrivateKey): "-----BEGIN RSA PRIVATE KEY-----\n<redacted>\n-----END RSA PRIVATE KEY-----",
						},
					},
				},
			},
			"legacySandboxFlag": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Legacy: sandbox flag, no authType or tokenUrl",
					Description: "A pre-existing datasource provisioned before authType/tokenUrl. With no authType the backend defaults to user auth (pkg/models/settings.go:82-90); with no tokenUrl but sandbox=true the token host is derived as https://test.salesforce.com (pkg/models/settings.go:92-100). LoadConfig.ApplyDefaults fills both in on load.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"user":    "user@example.com.sandbox",
							"sandbox": true,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword):     "<your-salesforce-password>",
							string(SecureJsonDataKeyClientID):     "<your-connected-app-consumer-key>",
							string(SecureJsonDataKeyClientSecret): "<your-connected-app-consumer-secret>",
						},
					},
				},
			},
		},
	}
}
