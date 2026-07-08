package stackdriver

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
// truth for the Google Cloud Monitoring datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the declarative dsconfig schema
// (single source of truth) for the Google Cloud Monitoring datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the full k8s-style SDK plugin schema for the Google Cloud
// Monitoring datasource: the settings (configuration) spec derived from
// dsconfig.json, the secure values, and example configurations, stamped with
// TargetAPIVersion. Grafana's datasource API server serves this bundle as
// {TargetAPIVersion}.json.
func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

// SettingsExamples returns k8s-style example configurations for the Google
// Cloud Monitoring datasource, covering the default configuration, each
// authentication type the config editor supports, and the additional-settings
// knobs. Each example value is a full instance-settings object with the plugin
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
					Description: "Authenticate as a Google service account. defaultProject, clientEmail, and tokenUri come from the service-account JSON; the private key is supplied inline as secureJsonData.privateKey. Grant the SA the 'Monitoring Viewer' role on the project and enable the Cloud Monitoring API.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authenticationType": string(AuthTypeJWT),
							"defaultProject":     "my-gcp-project",
							"clientEmail":        "cloud-monitoring-reader@my-gcp-project.iam.gserviceaccount.com",
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
					Description: "Same as googleJWTFile, but the private key is read from a file on the Grafana server via jsonData.privateKeyPath — no secret is stored in secureJsonData.privateKey. The referenced file may be a raw PEM key or a service-account JSON containing a private_key field (grafana-google-sdk-go/pkg/utils/utils.go:62-89).",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authenticationType": string(AuthTypeJWT),
							"defaultProject":     "my-gcp-project",
							"clientEmail":        "cloud-monitoring-reader@my-gcp-project.iam.gserviceaccount.com",
							"tokenUri":           "https://oauth2.googleapis.com/token",
							"privateKeyPath":     "/etc/secrets/cloud-monitoring-sa.json",
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
					Description: "Grafana is running on a Google Compute Engine VM; credentials are retrieved from the GCE metadata server. No secret is stored in secureJsonData. defaultProject may be omitted — the backend resolves it via utils.GCEDefaultProject (pkg/cloudmonitoring/cloudmonitoring.go:666-675). Do not populate jsonData.gceDefaultProject in provisioning payloads — that key is a frontend-only runtime cache (src/datasource.ts:186-191).",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authenticationType": string(AuthTypeGCE),
							"defaultProject":     "my-gcp-project",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPrivateKey): "",
						},
					},
				},
			},
			"workloadIdentityFederation": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Workload Identity Federation",
					Description: "Federated identity via a WIF pool provider. workloadIdentityPoolProvider is mandatory (backend fails at pkg/cloudmonitoring/httpclient.go:87-89 if empty); wifServiceAccountEmail is optional for impersonation. defaultProject is required (pkg/cloudmonitoring/cloudmonitoring.go:121-125 in CheckHealth). Only exposed in the editor when Grafana is running as a Cloud stack (src/utils.ts:15 isCloud).",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authenticationType":           string(AuthTypeWorkloadIdentityFederation),
							"workloadIdentityPoolProvider": "projects/123/locations/global/workloadIdentityPools/my-pool/providers/my-provider",
							"wifServiceAccountEmail":       "cloud-monitoring-reader@my-gcp-project.iam.gserviceaccount.com",
							"defaultProject":               "my-gcp-project",
							"oauthPassThru":                true,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPrivateKey): "",
						},
					},
				},
			},
			"forwardOAuthIdentity": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Forward OAuth Identity",
					Description: "The caller's OAuth token is forwarded to Cloud Monitoring. No credentials are stored on the datasource. The backend sets ForwardHTTPHeaders on the HTTP client at pkg/cloudmonitoring/cloudmonitoring.go:260-262; oauthPassThru is set to true as a side-effect. defaultProject is required (pkg/cloudmonitoring/cloudmonitoring.go:121-125 in CheckHealth). Alerting rules are not supported with this method (pkg/cloudmonitoring/cloudmonitoring.go:412-418).",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authenticationType": string(AuthTypeForwardOAuthIdentity),
							"defaultProject":     "my-gcp-project",
							"oauthPassThru":      true,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPrivateKey): "",
						},
					},
				},
			},
			"impersonation": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "JWT with service account impersonation",
					Description: "JWT auth where the base service account impersonates another SA. Only usable for jwt / gce auth (ConfigEditor.tsx:41-42 controls when the AuthConfig impersonation UI renders). Backend uses NewImpersonatedJwtAccessTokenProvider (pkg/cloudmonitoring/httpclient.go:68-70).",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authenticationType":          string(AuthTypeJWT),
							"defaultProject":              "my-gcp-project",
							"clientEmail":                 "cloud-monitoring-caller@my-gcp-project.iam.gserviceaccount.com",
							"tokenUri":                    "https://oauth2.googleapis.com/token",
							"usingImpersonation":          true,
							"serviceAccountToImpersonate": "cloud-monitoring-reader@my-gcp-project.iam.gserviceaccount.com",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPrivateKey): "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----\n",
						},
					},
				},
			},
			"universeDomain": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "JWT with a custom universe domain",
					Description: "JWT auth targeting a non-default Google Cloud universe (Trusted Partner Cloud, mTLS endpoint). The backend joins each service host with universeDomain at pkg/cloudmonitoring/httpclient.go:79-83; empty is treated as 'googleapis.com'. The editor only surfaces this field when the Grafana instance has secure_socks_datasource_proxy enabled, but provisioning payloads may set it anytime.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authenticationType": string(AuthTypeJWT),
							"defaultProject":     "my-gcp-project",
							"clientEmail":        "cloud-monitoring-reader@my-gcp-project.iam.gserviceaccount.com",
							"tokenUri":           "https://oauth2.googleapis.com/token",
							"universeDomain":     "googleapis.mtls.google.com",
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
