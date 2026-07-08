package falconlogscaledatasource

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
// truth for the Falcon LogScale datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the declarative dsconfig schema
// (single source of truth) for the Falcon LogScale datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the full k8s-style SDK plugin schema for the Falcon
// LogScale datasource: the settings (configuration) spec derived from
// dsconfig.json, the secure values, and example configurations, stamped with
// TargetAPIVersion. Grafana's datasource API server serves this bundle as
// {TargetAPIVersion}.json.
func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

// SettingsExamples returns k8s-style example configurations for the Falcon
// LogScale datasource, covering the default configuration and each
// authentication method / data-source-mode combination the config editor
// supports. Each example value is a full instance settings object with the
// plugin configuration nested under jsonData and the relevant write-only
// secrets under secureJsonData (placeholder values — replace them with real
// secrets).
func SettingsExamples() *sdkschema.SettingsExamples {
	return &sdkschema.SettingsExamples{
		Examples: map[string]*spec3.Example{
			"": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Default configuration",
					Description: "The defaults a new datasource starts with: LogScale mode, LogScale personal-token authentication. Only root.url and secureJsonData.accessToken (empty here) need to be filled in.",
					Value: map[string]any{
						"url": "https://cloud.humio.com",
						"jsonData": map[string]any{
							"mode":                  string(DataSourceModeLogScale),
							"authenticateWithToken": true,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAccessToken): "",
						},
					},
				},
			},
			"logscaleToken": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "LogScale token authentication",
					Description: "Authenticate to a LogScale server with a personal token in secureJsonData.accessToken. The backend derives '<url>/humio/graphql' and '<url>/humio' as the client endpoints (pkg/plugin/settings.go:44-46).",
					Value: map[string]any{
						"url": "https://cloud.humio.com",
						"jsonData": map[string]any{
							"mode":                  string(DataSourceModeLogScale),
							"authenticateWithToken": true,
							"defaultRepository":     "example-repo",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAccessToken): "REPLACE_WITH_LOGSCALE_TOKEN",
						},
					},
				},
			},
			"logscaleOAuth2Client": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "LogScale mode with OAuth2 client credentials",
					Description: "Use the OAuth2 client-credentials grant against a LogScale server. jsonData.oauth2ClientId + secureJsonData.oauth2ClientSecret drive the token exchange; the plugin fetches and refreshes access tokens automatically.",
					Value: map[string]any{
						"url": "https://cloud.humio.com",
						"jsonData": map[string]any{
							"mode":           string(DataSourceModeLogScale),
							"oauth2":         true,
							"oauth2ClientId": "my-client-id",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyOAuth2ClientSecret): "REPLACE_WITH_OAUTH2_CLIENT_SECRET",
						},
					},
				},
			},
			"logscaleBasicAuth": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "LogScale mode with HTTP Basic authentication",
					Description: "Send HTTP Basic credentials with each request. root.basicAuth toggles the method; root.basicAuthUser and secureJsonData.basicAuthPassword carry the credentials. The backend reads basicAuthUser at pkg/plugin/settings.go:59.",
					Value: map[string]any{
						"url":           "https://cloud.humio.com",
						"basicAuth":     true,
						"basicAuthUser": "grafana",
						"jsonData": map[string]any{
							"mode": string(DataSourceModeLogScale),
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyBasicAuthPassword): "REPLACE_WITH_PASSWORD",
						},
					},
				},
			},
			"logscaleOAuthForward": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "LogScale mode with Forward OAuth Identity",
					Description: "Forward the signed-in Grafana user's upstream OAuth identity to LogScale. Sets only jsonData.oauthPassThru; there is no secret to configure.",
					Value: map[string]any{
						"url": "https://cloud.humio.com",
						"jsonData": map[string]any{
							"mode":          string(DataSourceModeLogScale),
							"oauthPassThru": true,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAccessToken): "",
						},
					},
				},
			},
			"ngsiemOAuth2": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "NGSIEM mode with OAuth2 client credentials",
					Description: "CrowdStrike NGSIEM tenant. Only OAuth2 client credentials is accepted; defaultRepository is auto-pinned to 'search-all' by the editor (ConfigEditor.tsx:127-137) and reflected here. The backend appends '/humio' to the base URL when Mode == NGSIEM (pkg/plugin/plugin.go:54-56); do not include '/humio' in the configured URL.",
					Value: map[string]any{
						"url": "https://api.us-2.crowdstrike.com",
						"jsonData": map[string]any{
							"mode":              string(DataSourceModeNGSIEM),
							"oauth2":            true,
							"oauth2ClientId":    "my-ngsiem-client-id",
							"defaultRepository": "search-all",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyOAuth2ClientSecret): "REPLACE_WITH_NGSIEM_CLIENT_SECRET",
						},
					},
				},
			},
			"logscaleWithDataLinks": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "LogScale token auth with data links and incremental querying",
					Description: "A fully-populated LogScale instance: token auth, custom default repository, two data links (one external URL template, one internal link to a Tempo datasource), incremental querying enabled with a 30-second overlap window.",
					Value: map[string]any{
						"url": "https://cloud.humio.com",
						"jsonData": map[string]any{
							"mode":                          string(DataSourceModeLogScale),
							"authenticateWithToken":         true,
							"defaultRepository":             "production",
							"incrementalQuerying":           true,
							"incrementalQueryOverlapWindow": "30s",
							"dataLinks": []any{
								map[string]any{
									"field":        "requestId",
									"label":        "View request",
									"matcherRegex": "req_id=(\\w+)",
									"url":          "https://requests.example.com/${__value.raw}",
								},
								map[string]any{
									"field":         "traceId",
									"label":         "Trace",
									"matcherRegex":  "trace_id=(\\w+)",
									"url":           "${__value.raw}",
									"datasourceUid": "tempo",
								},
							},
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAccessToken): "REPLACE_WITH_LOGSCALE_TOKEN",
						},
					},
				},
			},
		},
	}
}
