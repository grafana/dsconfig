package astradbdatasource

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
// truth for the AstraDB datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the declarative dsconfig schema
// (single source of truth) for the AstraDB datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the full k8s-style SDK plugin schema for the AstraDB
// datasource: the settings (configuration) spec derived from dsconfig.json,
// the secure values, and example configurations, stamped with
// TargetAPIVersion. Grafana's datasource API server serves this bundle as
// {TargetAPIVersion}.json.
func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

// SettingsExamples returns k8s-style example configurations for the AstraDB
// datasource, covering the default configuration and each authentication /
// connection variant the config editor supports. Each example value is a full
// instance settings object with the plugin configuration nested under jsonData
// and the relevant write-only secrets under secureJsonData (placeholder
// values — replace them with real secrets).
func SettingsExamples() *sdkschema.SettingsExamples {
	return &sdkschema.SettingsExamples{
		Examples: map[string]*spec3.Example{
			"": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Default configuration",
					Description: "The defaults a new datasource starts with: Token authentication against DataStax Astra Cloud (authKind=0). Both jsonData.uri and secureJsonData.token (both empty here) must be filled in to get a working datasource.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authKind": int(AuthKindToken),
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyToken): "",
						},
					},
				},
			},
			"tokenAstraCloud": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Token authentication (DataStax Astra Cloud)",
					Description: "Authenticate against DataStax Astra Cloud with an application token (starts with AstraCS:). The URI is the Astra gRPC host:port; the backend always uses TLS in this mode.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authKind": int(AuthKindToken),
							"uri":      "cluster-id-region.apps.astra.datastax.com:443",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyToken): "AstraCS:XXXXXXXXXXXXXXXXXXXXXXXX:XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
						},
					},
				},
			},
			"credentialsSelfHostedTLS": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Credentials authentication (self-hosted Stargate, TLS)",
					Description: "Authenticate against a self-hosted Stargate deployment with basic-auth username/password. secure=true uses TLS on the gRPC channel and https:// on the auth endpoint.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authKind":     int(AuthKindCredentials),
							"grpcEndpoint": "stargate.example.com:8090",
							"authEndpoint": "stargate.example.com:8081",
							"user":         "cassandra",
							"secure":       true,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword): "cassandra",
						},
					},
				},
			},
			"credentialsSelfHostedPlaintext": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Credentials authentication (self-hosted Stargate, plaintext)",
					Description: "Local/dev Stargate over plaintext: secure=false uses insecure gRPC credentials and http:// on the auth endpoint. Only appropriate for localhost or trusted networks.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authKind":     int(AuthKindCredentials),
							"grpcEndpoint": "localhost:8090",
							"authEndpoint": "localhost:8081",
							"user":         "cassandra",
							"secure":       false,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword): "cassandra",
						},
					},
				},
			},
			"legacyMissingAuthKind": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Legacy: no authKind stored",
					Description: "Datasources saved without authKind in jsonData are treated as Token mode (authKind=0) because the numeric zero value equals AuthKindToken. The editor mirrors this via `jsonData.authKind || Connection.TOKEN` (ConfigEditor.tsx:71).",
					Value: map[string]any{
						"jsonData": map[string]any{
							"uri": "cluster-id-region.apps.astra.datastax.com:443",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyToken): "AstraCS:XXXXXXXXXXXXXXXXXXXXXXXX:XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
						},
					},
				},
			},
		},
	}
}
