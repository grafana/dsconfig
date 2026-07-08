package googlesheetsdatasource

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
// truth for the Google Sheets datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the declarative dsconfig schema
// (single source of truth) for the Google Sheets datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the full k8s-style SDK plugin schema for the Google
// Sheets datasource: the settings (configuration) spec derived from
// dsconfig.json, the secure values, and example configurations, stamped with
// TargetAPIVersion. Grafana's datasource API server serves this bundle as
// {TargetAPIVersion}.json.
func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

// SettingsExamples returns k8s-style example configurations for the Google
// Sheets datasource, covering the default configuration, each authentication
// type the config editor supports, and the legacy `authType` shape. Each
// example value is a full instance settings object with the plugin
// configuration nested under jsonData and the relevant write-only secrets
// under secureJsonData (placeholder values — replace them with real secrets).
func SettingsExamples() *sdkschema.SettingsExamples {
	return &sdkschema.SettingsExamples{
		Examples: map[string]*spec3.Example{
			"": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Default configuration",
					Description: "The defaults a new datasource starts with: Google JWT File authentication (jwt). The user must still supply defaultProject, clientEmail, tokenUri, and a private key (either inline or via privateKeyPath) to get a working datasource.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authenticationType": string(AuthTypeJWT),
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPrivateKey): "",
						},
					},
				},
			},
			"apiKey": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "API Key (public spreadsheets only)",
					Description: "Authenticate to the Google Sheets API with an API key. Works only for spreadsheets that are shared publicly (anyone with the link can view). Enable the Google Sheets API for the project the key belongs to.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authenticationType": string(AuthTypeAPIKey),
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAPIKey): "AIzaSyXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
						},
					},
				},
			},
			"googleJWTFile": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Google JWT File (inline private key)",
					Description: "Authenticate as a Google service account. defaultProject, clientEmail, and tokenUri come from the service-account JSON; the private key is supplied inline as secureJsonData.privateKey. Also enable both the Google Sheets and Google Drive APIs for the project.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authenticationType": string(AuthTypeJWT),
							"defaultProject":     "my-gcp-project",
							"clientEmail":        "sheets-reader@my-gcp-project.iam.gserviceaccount.com",
							"tokenUri":           "https://oauth2.googleapis.com/token",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPrivateKey): "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----\n",
						},
					},
				},
			},
			"googleJWTFilePath": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Google JWT File (private key path)",
					Description: "Same as googleJWTFile, but the private key is read from a file on the Grafana server via jsonData.privateKeyPath — no secret is stored in secureJsonData.privateKey. The referenced file may be a raw PEM key or a service-account JSON containing a private_key field.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authenticationType": string(AuthTypeJWT),
							"defaultProject":     "my-gcp-project",
							"clientEmail":        "sheets-reader@my-gcp-project.iam.gserviceaccount.com",
							"tokenUri":           "https://oauth2.googleapis.com/token",
							"privateKeyPath":     "/etc/secrets/sheets-sa.json",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPrivateKey): "",
						},
					},
				},
			},
			"gceDefaultServiceAccount": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "GCE Default Service Account",
					Description: "Grafana is running on a Google Compute Engine VM; credentials are retrieved from the GCE metadata server. No secret is stored in secureJsonData. Optionally set defaultProject.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authenticationType": string(AuthTypeGCE),
							"defaultProject":     "my-gcp-project",
						},
						"secureJsonData": map[string]any{},
					},
				},
			},
			"legacyAuthType": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Legacy: authType instead of authenticationType",
					Description: "Older provisioned datasources stored the auth type in jsonData.authType. The backend copies authType into authenticationType on load (pkg/models/settings.go:49-51). New configurations should write only authenticationType.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType":       string(AuthTypeAPIKey),
							"defaultSheetID": "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAPIKey): "AIzaSyXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
						},
					},
				},
			},
			"withDefaultSheetID": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "API Key with default spreadsheet",
					Description: "Same as apiKey but with a defaultSheetID pre-populated so new queries start with a spreadsheet selected. The backend loads defaultSheetID but does not consume it at query time; it is a UX hint.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authenticationType": string(AuthTypeAPIKey),
							"defaultSheetID":     "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAPIKey): "AIzaSyXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
						},
					},
				},
			},
		},
	}
}
