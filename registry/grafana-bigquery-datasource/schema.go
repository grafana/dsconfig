package bigquerydatasource

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
// truth for the Google BigQuery datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the declarative dsconfig schema
// (single source of truth) for the Google BigQuery datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the full k8s-style SDK plugin schema for the Google
// BigQuery datasource: the settings (configuration) spec derived from
// dsconfig.json, the secure values, and example configurations, stamped with
// TargetAPIVersion. Grafana's datasource API server serves this bundle as
// {TargetAPIVersion}.json.
func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

// SettingsExamples returns k8s-style example configurations for the Google
// BigQuery datasource, covering the default configuration, each authentication
// type the config editor supports, and the additional-settings knobs. Each
// example value is a full instance-settings object with the plugin
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
			"googleJWTFile": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Google JWT File (inline private key)",
					Description: "Authenticate as a Google service account. defaultProject, clientEmail, and tokenUri come from the service-account JSON; the private key is supplied inline as secureJsonData.privateKey. Grant the SA the 'BigQuery Data Viewer' and 'Job User' roles and enable the BigQuery API for the project.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authenticationType": string(AuthTypeJWT),
							"defaultProject":     "my-gcp-project",
							"clientEmail":        "bigquery-reader@my-gcp-project.iam.gserviceaccount.com",
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
					Description: "Same as googleJWTFile, but the private key is read from a file on the Grafana server via jsonData.privateKeyPath — no secret is stored in secureJsonData.privateKey. The referenced file may be a raw PEM key or a service-account JSON containing a private_key field (grafana-google-sdk-go/pkg/utils/utils.go:62-80).",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authenticationType": string(AuthTypeJWT),
							"defaultProject":     "my-gcp-project",
							"clientEmail":        "bigquery-reader@my-gcp-project.iam.gserviceaccount.com",
							"tokenUri":           "https://oauth2.googleapis.com/token",
							"privateKeyPath":     "/etc/secrets/bigquery-sa.json",
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
					Description: "Grafana is running on a Google Compute Engine VM; credentials are retrieved from the GCE metadata server. No secret is stored in secureJsonData. Optionally set defaultProject to override the project resolved from the metadata server.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authenticationType": string(AuthTypeGCE),
							"defaultProject":     "my-gcp-project",
						},
						"secureJsonData": map[string]any{},
					},
				},
			},
			"forwardOAuthIdentity": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Forward OAuth Identity",
					Description: "The caller's OAuth token is forwarded to BigQuery. No credentials are stored on the datasource. The backend sets ForwardHTTPHeaders on the HTTP client at pkg/bigquery/http_client.go:99-107; oauthPassThru is set to true as a side-effect.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authenticationType": string(AuthTypeForwardOAuthIdentity),
							"oauthPassThru":      true,
						},
						"secureJsonData": map[string]any{},
					},
				},
			},
			"workloadIdentityFederation": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Workload Identity Federation",
					Description: "Federated identity via a WIF pool provider. workloadIdentityPoolProvider is mandatory (backend fails at pkg/bigquery/http_client.go:95 if empty); wifServiceAccountEmail is optional for impersonation. Only exposed in the editor when Grafana is running as a Cloud stack (src/types.ts:47).",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authenticationType":           string(AuthTypeWorkloadIdentityFederation),
							"workloadIdentityPoolProvider": "projects/123/locations/global/workloadIdentityPools/my-pool/providers/my-provider",
							"wifServiceAccountEmail":       "bigquery-reader@my-gcp-project.iam.gserviceaccount.com",
							"defaultProject":               "my-gcp-project",
							"oauthPassThru":                true,
						},
						"secureJsonData": map[string]any{},
					},
				},
			},
			"impersonation": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "JWT with service account impersonation",
					Description: "JWT auth where the base service account impersonates another SA. Only usable for jwt / gce auth (ConfigEditor.tsx:35-36 controls when the AuthConfig impersonation UI renders). Backend uses NewImpersonatedJwtAccessTokenProvider (pkg/bigquery/http_client.go:73-80).",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authenticationType":          string(AuthTypeJWT),
							"defaultProject":              "my-gcp-project",
							"clientEmail":                 "bigquery-caller@my-gcp-project.iam.gserviceaccount.com",
							"tokenUri":                    "https://oauth2.googleapis.com/token",
							"usingImpersonation":          true,
							"serviceAccountToImpersonate": "bigquery-reader@my-gcp-project.iam.gserviceaccount.com",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPrivateKey): "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----\n",
						},
					},
				},
			},
			"additionalSettings": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "JWT with all Additional Settings populated",
					Description: "JWT auth plus every optional 'Additional Settings' knob: processingLocation (regional), serviceEndpoint (private-service-connect override), and MaxBytesBilled cap.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authenticationType": string(AuthTypeJWT),
							"defaultProject":     "my-gcp-project",
							"clientEmail":        "bigquery-reader@my-gcp-project.iam.gserviceaccount.com",
							"tokenUri":           "https://oauth2.googleapis.com/token",
							"processingLocation": "EU",
							"serviceEndpoint":    "https://bigquery.googleapis.com/bigquery/v2/",
							"MaxBytesBilled":     int64(5_242_880),
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPrivateKey): "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----\n",
						},
					},
				},
			},
		},
	}
}
