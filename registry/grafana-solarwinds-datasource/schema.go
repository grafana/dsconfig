package solarwindsdatasource

import (
	_ "embed"

	"k8s.io/kube-openapi/pkg/spec3"

	"github.com/grafana/dsconfig/dsconfig"
	sdkschema "github.com/grafana/grafana-plugin-sdk-go/experimental/pluginschema"
)

//go:generate go test -generateArtifacts -run TestSchemaConformance ./...

// TargetAPIVersion is the API version this schema applies to.
const TargetAPIVersion = dsconfig.TargetAPIVersion

// configSchemaJSON is the dsconfig schema — the single source of truth for the
// SolarWinds datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the dsconfig schema for the SolarWinds datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the full k8s-style SDK plugin schema for the SolarWinds datasource.
func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

// SettingsExamples returns k8s-style example configurations: the default
// configuration, a basic-auth example, and a basic-auth + mutual-TLS example.
func SettingsExamples() *sdkschema.SettingsExamples {
	return &sdkschema.SettingsExamples{
		Examples: map[string]*spec3.Example{
			"": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Default configuration",
					Description: "The defaults a new datasource starts with: the single service 'solarwinds' preselects its only auth method, basic_auth. Fill in the instance url (jsonData.variables.url), username, and password (secureJsonData \"solarwinds.password\", empty here) to get a working datasource.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"services": map[string]any{
								"solarwinds": map[string]any{
									"auth": map[string]any{
										"id":       string(AuthMethodBasic),
										"username": "",
									},
								},
							},
							"variables": map[string]any{"url": ""},
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword): "",
						},
					},
				},
			},
			"basicAuth": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Basic auth",
					Description: "Authenticate against {url}:17774/SolarWinds/InformationService/v3/Json with a username and password (the password is the write-only secret at secureJsonData \"solarwinds.password\").",
					Value: map[string]any{
						"jsonData": map[string]any{
							"services": map[string]any{
								"solarwinds": map[string]any{
									"auth": map[string]any{
										"id":       string(AuthMethodBasic),
										"username": "admin",
									},
								},
							},
							"variables": map[string]any{"url": "https://solarwinds.example.com"},
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword): "<solarwinds-password>",
						},
					},
				},
			},
			"basicAuthMutualTLS": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Basic auth with mutual TLS",
					Description: "Basic auth plus TLS client authentication: the client certificate and key are write-only secrets at secureJsonData \"solarwinds.tls.clientCert\" and \"solarwinds.tls.clientKey\".",
					Value: map[string]any{
						"jsonData": map[string]any{
							"services": map[string]any{
								"solarwinds": map[string]any{
									"auth": map[string]any{
										"id":       string(AuthMethodBasic),
										"username": "admin",
										"tls": map[string]any{
											"clientAuth": map[string]any{
												"enabled":    true,
												"serverName": "solarwinds.example.com",
											},
										},
									},
								},
							},
							"variables": map[string]any{"url": "https://solarwinds.example.com"},
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword):      "<solarwinds-password>",
							string(SecureJsonDataKeyTLSClientCert): "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----",
							string(SecureJsonDataKeyTLSClientKey):  "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----",
						},
					},
				},
			},
		},
	}
}
