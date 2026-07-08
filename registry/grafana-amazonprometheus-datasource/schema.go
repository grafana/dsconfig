package amazonprometheusdatasource

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
// truth for the Amazon Managed Service for Prometheus datasource
// configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the declarative dsconfig schema
// for the Amazon Managed Service for Prometheus datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the k8s-style SDK plugin schema bundle Grafana's
// datasource API server serves as {TargetAPIVersion}.json.
func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

// SettingsExamples returns k8s-style example configurations for the Amazon
// Managed Service for Prometheus datasource, covering the default
// (schema-defaults) configuration plus one example per SigV4 authentication
// type the config editor supports and one legacy provisioning example.
//
// Each example value is a full instance-settings object with the plugin
// configuration nested under `jsonData` and the relevant write-only secrets
// under `secureJsonData` (placeholder values — replace them with real
// secrets).
func SettingsExamples() *sdkschema.SettingsExamples {
	return &sdkschema.SettingsExamples{
		Examples: map[string]*spec3.Example{
			"": {
				ExampleProps: spec3.ExampleProps{
					Summary: "Default configuration",
					Description: "The defaults a new datasource starts with. `sigV4Auth` is forced to true and " +
						"`sigv4Service` defaults to \"aps\" (Amazon Managed Prometheus). `httpMethod` defaults to " +
						"POST. The user must still pick a SigV4 auth provider and (when applicable) supply " +
						"credentials before the datasource is usable.",
					Value: map[string]any{
						"url": "https://aps-workspaces.<region>.amazonaws.com/workspaces/<workspace-id>",
						"jsonData": map[string]any{
							"sigV4Auth":    true,
							"sigv4Service": DefaultSigV4Service,
							"httpMethod":   string(HTTPMethodPOST),
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeySigV4AccessKey): "",
							string(SecureJsonDataKeySigV4SecretKey): "",
						},
					},
				},
			},
			"ec2IamRole": {
				ExampleProps: spec3.ExampleProps{
					Summary: "Workspace IAM Role",
					Description: "Sign requests using the IAM role attached to the current EC2 instance / ECS " +
						"task / EKS pod. No secret material required — the AWS SDK reads temporary credentials " +
						"from the instance metadata service.",
					Value: map[string]any{
						"url": "https://aps-workspaces.us-east-1.amazonaws.com/workspaces/ws-12345678-1234-1234-1234-123456789012",
						"jsonData": map[string]any{
							"sigV4Auth":     true,
							"sigV4AuthType": string(SigV4AuthTypeEC2IAMRole),
							"sigV4Region":   "us-east-1",
							"sigv4Service":  DefaultSigV4Service,
							"httpMethod":    string(HTTPMethodPOST),
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeySigV4AccessKey): "",
							string(SecureJsonDataKeySigV4SecretKey): "",
						},
					},
				},
			},
			"accessKeys": {
				ExampleProps: spec3.ExampleProps{
					Summary: "Access & secret key",
					Description: "Sign requests using a static AWS IAM user's access key + secret key pair. " +
						"Both secrets are stored write-only in `secureJsonData` (`sigV4AccessKey`, " +
						"`sigV4SecretKey`).",
					Value: map[string]any{
						"url": "https://aps-workspaces.us-east-1.amazonaws.com/workspaces/ws-12345678-1234-1234-1234-123456789012",
						"jsonData": map[string]any{
							"sigV4Auth":     true,
							"sigV4AuthType": string(SigV4AuthTypeKeys),
							"sigV4Region":   "us-east-1",
							"sigv4Service":  DefaultSigV4Service,
							"httpMethod":    string(HTTPMethodPOST),
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeySigV4AccessKey): "AKIAIOSFODNN7EXAMPLE",
							string(SecureJsonDataKeySigV4SecretKey): "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
						},
					},
				},
			},
			"credentialsFile": {
				ExampleProps: spec3.ExampleProps{
					Summary: "Credentials file",
					Description: "Read credentials from a named profile in `~/.aws/credentials` on the Grafana " +
						"host. Leave `sigV4Profile` empty to use the default profile.",
					Value: map[string]any{
						"url": "https://aps-workspaces.us-east-1.amazonaws.com/workspaces/ws-12345678-1234-1234-1234-123456789012",
						"jsonData": map[string]any{
							"sigV4Auth":     true,
							"sigV4AuthType": string(SigV4AuthTypeCredentials),
							"sigV4Profile":  "aps-workspace-1",
							"sigV4Region":   "us-east-1",
							"sigv4Service":  DefaultSigV4Service,
							"httpMethod":    string(HTTPMethodPOST),
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeySigV4AccessKey): "",
							string(SecureJsonDataKeySigV4SecretKey): "",
						},
					},
				},
			},
			"awsSdkDefault": {
				ExampleProps: spec3.ExampleProps{
					Summary: "AWS SDK Default credential chain",
					Description: "Delegate credential lookup to the AWS SDK default chain (env vars, shared " +
						"config, EC2/ECS metadata, ...). No secret material required.",
					Value: map[string]any{
						"url": "https://aps-workspaces.us-east-1.amazonaws.com/workspaces/ws-12345678-1234-1234-1234-123456789012",
						"jsonData": map[string]any{
							"sigV4Auth":     true,
							"sigV4AuthType": string(SigV4AuthTypeDefault),
							"sigV4Region":   "us-east-1",
							"sigv4Service":  DefaultSigV4Service,
							"httpMethod":    string(HTTPMethodPOST),
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeySigV4AccessKey): "",
							string(SecureJsonDataKeySigV4SecretKey): "",
						},
					},
				},
			},
			"grafanaAssumeRole": {
				ExampleProps: spec3.ExampleProps{
					Summary: "Grafana Assume Role (Grafana Cloud)",
					Description: "Delegate to Grafana Cloud's temporary credentials broker. Feature-gated on " +
						"the Grafana instance's `awsDatasourcesTempCredentials` toggle. The external ID field " +
						"hides itself when this auth type is selected because Grafana Cloud injects the " +
						"external ID automatically (`ConnectionConfig.tsx:274`).",
					Value: map[string]any{
						"url": "https://aps-workspaces.us-east-1.amazonaws.com/workspaces/ws-12345678-1234-1234-1234-123456789012",
						"jsonData": map[string]any{
							"sigV4Auth":          true,
							"sigV4AuthType":      string(SigV4AuthTypeGrafanaAssumeRole),
							"sigV4AssumeRoleArn": "arn:aws:iam::123456789012:role/GrafanaCloudPrometheus",
							"sigV4Region":        "us-east-1",
							"sigv4Service":       DefaultSigV4Service,
							"httpMethod":         string(HTTPMethodPOST),
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeySigV4AccessKey): "",
							string(SecureJsonDataKeySigV4SecretKey): "",
						},
					},
				},
			},
			"assumeRoleCrossAccount": {
				ExampleProps: spec3.ExampleProps{
					Summary: "Access keys + Assume Role ARN (cross-account)",
					Description: "Use static access keys to call STS AssumeRole against a role in another AWS " +
						"account, with an external ID for the cross-account trust policy.",
					Value: map[string]any{
						"url": "https://aps-workspaces.us-east-1.amazonaws.com/workspaces/ws-12345678-1234-1234-1234-123456789012",
						"jsonData": map[string]any{
							"sigV4Auth":          true,
							"sigV4AuthType":      string(SigV4AuthTypeKeys),
							"sigV4AssumeRoleArn": "arn:aws:iam::987654321098:role/PrometheusReader",
							"sigV4ExternalId":    "my-external-id",
							"sigV4Region":        "us-east-1",
							"sigv4Service":       DefaultSigV4Service,
							"httpMethod":         string(HTTPMethodPOST),
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeySigV4AccessKey): "AKIAIOSFODNN7EXAMPLE",
							string(SecureJsonDataKeySigV4SecretKey): "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
						},
					},
				},
			},
			"forwardGrafanaUserHeader": {
				ExampleProps: spec3.ExampleProps{
					Summary: "Forward Grafana User HTTP Header",
					Description: "Enable `forwardGrafanaUserHeader` so the plugin adds the logged-in Grafana " +
						"user's `X-Grafana-User` header on every upstream request. Requires " +
						"`send_user_header = true` in the Grafana server configuration.",
					Value: map[string]any{
						"url": "https://aps-workspaces.us-east-1.amazonaws.com/workspaces/ws-12345678-1234-1234-1234-123456789012",
						"jsonData": map[string]any{
							"sigV4Auth":                true,
							"sigV4AuthType":            string(SigV4AuthTypeEC2IAMRole),
							"sigV4Region":              "us-east-1",
							"sigv4Service":             DefaultSigV4Service,
							"forwardGrafanaUserHeader": true,
							"httpMethod":               string(HTTPMethodPOST),
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeySigV4AccessKey): "",
							string(SecureJsonDataKeySigV4SecretKey): "",
						},
					},
				},
			},
			"migratedFromPrometheus": {
				ExampleProps: spec3.ExampleProps{
					Summary: "Migrated from vanilla Prometheus",
					Description: "Datasources migrated from the vanilla Prometheus plugin carry the sentinel " +
						"flag `jsonData['prometheus-type-migration'] = true`, which triggers the " +
						"'Data source migrated' banner at `ConfigEditor.tsx:37-48`. The banner is purely " +
						"informational — nothing in the runtime depends on the flag.",
					Value: map[string]any{
						"url": "https://aps-workspaces.us-east-1.amazonaws.com/workspaces/ws-12345678-1234-1234-1234-123456789012",
						"jsonData": map[string]any{
							"sigV4Auth":                 true,
							"sigV4AuthType":             string(SigV4AuthTypeEC2IAMRole),
							"sigV4Region":               "us-east-1",
							"sigv4Service":              DefaultSigV4Service,
							"httpMethod":                string(HTTPMethodPOST),
							"prometheus-type-migration": true,
							"prometheusType":            string(PromApplicationPrometheus),
							"prometheusVersion":         "2.50.1",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeySigV4AccessKey): "",
							string(SecureJsonDataKeySigV4SecretKey): "",
						},
					},
				},
			},
		},
	}
}
