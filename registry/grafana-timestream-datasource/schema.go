package timestreamdatasource

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
// truth for the Amazon Timestream datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the declarative dsconfig schema
// (single source of truth) for the Timestream datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the full k8s-style SDK plugin schema for the Timestream
// datasource: the settings (configuration) spec derived from dsconfig.json,
// the secure values, and example configurations, stamped with
// TargetAPIVersion. Grafana's datasource API server serves this bundle as
// {TargetAPIVersion}.json.
func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

// SettingsExamples returns k8s-style example configurations for the
// Timestream datasource, covering the default configuration and each
// editor-selectable AWS authentication provider plus an STS assume-role
// variant, the legacy `arn` value, and a fully-macro-defaulted example.
// Each example is a full instance-settings object with the plugin
// configuration under jsonData and the relevant write-only secrets under
// secureJsonData (placeholder values — replace them with real secrets).
func SettingsExamples() *sdkschema.SettingsExamples {
	return &sdkschema.SettingsExamples{
		Examples: map[string]*spec3.Example{
			"": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Default configuration",
					Description: "The defaults a fresh datasource starts with: AWS SDK Default credentials. `defaultRegion` still needs to be set before queries will run; secureJsonData is empty because the default provider reads credentials from the AWS SDK chain.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType": string(AWSAuthTypeDefault),
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAccessKey): "",
						},
					},
				},
			},
			"awsSdkDefault": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "AWS SDK Default",
					Description: "Uses the AWS SDK default credential chain (env vars, shared config, EC2/ECS/EKS metadata). No secrets are set; the placeholder accessKey is empty to signal that.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType":      string(AWSAuthTypeDefault),
							"defaultRegion": "us-east-1",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAccessKey): "",
						},
					},
				},
			},
			"accessAndSecretKey": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Access & secret key",
					Description: "IAM user credentials supplied directly via secureJsonData. Optionally include sessionToken for temporary STS credentials.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType":      string(AWSAuthTypeKeys),
							"defaultRegion": "us-east-1",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAccessKey): "AKIAIOSFODNN7EXAMPLE",
							string(SecureJsonDataKeySecretKey): "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
						},
					},
				},
			},
			"credentialsFile": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Credentials file",
					Description: "Reads a named profile from ~/.aws/credentials on the Grafana host. Leave `profile` blank to pick the default profile.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType":      string(AWSAuthTypeCredentials),
							"profile":       "my-timestream-profile",
							"defaultRegion": "us-east-1",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAccessKey): "",
						},
					},
				},
			},
			"workspaceIamRole": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Workspace IAM Role (ec2_iam_role)",
					Description: "Uses the IAM role attached to the Grafana workload (EC2 instance profile / ECS task role / EKS IRSA). Editor label is 'Workspace IAM Role'; storage value is 'ec2_iam_role'.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType":      string(AWSAuthTypeEC2IAMRole),
							"defaultRegion": "us-east-1",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAccessKey): "",
						},
					},
				},
			},
			"grafanaAssumeRole": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Grafana Assume Role (Grafana Cloud only)",
					Description: "Grafana Cloud's temporary-credentials broker. Feature-gated on the awsDatasourcesTempCredentials toggle; the editor hides externalId/endpoint when this provider is selected because Grafana derives them itself. Timestream is in the ConnectionConfig allow-list (`DS_TYPES_THAT_SUPPORT_TEMP_CREDS`).",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType":      string(AWSAuthTypeGrafanaAssumeRole),
							"defaultRegion": "us-east-1",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAccessKey): "",
						},
					},
				},
			},
			"assumeRoleFromKeys": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Access & secret key + STS AssumeRole",
					Description: "IAM user credentials that then assume a cross-account IAM role via STS. `externalId` is optional and only meaningful when `assumeRoleArn` is set.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType":      string(AWSAuthTypeKeys),
							"assumeRoleArn": "arn:aws:iam::123456789012:role/GrafanaTimestream",
							"externalId":    "external-id-abc123",
							"defaultRegion": "us-east-1",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAccessKey): "AKIAIOSFODNN7EXAMPLE",
							string(SecureJsonDataKeySecretKey): "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
						},
					},
				},
			},
			"legacyArnAuthType": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Legacy: authType = 'arn'",
					Description: "Datasources provisioned before the 'arn' value was renamed still store `authType: \"arn\"`. The backend (awsds.AuthType.UnmarshalJSON) maps this to the modern 'default' at load time.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType":      string(AWSAuthTypeARN),
							"defaultRegion": "us-east-1",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAccessKey): "",
						},
					},
				},
			},
			"withMacroDefaults": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "AWS SDK Default with Timestream macro defaults",
					Description: "Seeds the {{database}}, {{table}}, and {{measure}} query macros so dashboards do not need to specify them per panel. The macro defaults are optional at save-time and can be overridden per query.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType":        string(AWSAuthTypeDefault),
							"defaultRegion":   "us-east-1",
							"defaultDatabase": "sampleDB",
							"defaultTable":    "IoT",
							"defaultMeasure":  "speed",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAccessKey): "",
						},
					},
				},
			},
		},
	}
}
