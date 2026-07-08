package iottwinmakerdatasource

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
// of truth for the IoT TwinMaker datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the declarative dsconfig
// schema (single source of truth) for the IoT TwinMaker datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the full k8s-style SDK plugin schema for the IoT
// TwinMaker datasource: the settings (configuration) spec derived from
// dsconfig.json, the secure values, and example configurations, stamped
// with TargetAPIVersion. Grafana's datasource API server serves this
// bundle as {TargetAPIVersion}.json.
func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

// exampleWorkspaceID is a placeholder workspace name used in examples.
const exampleWorkspaceID = "MyTwinMakerWorkspace"

// exampleAssumeRoleARN is a placeholder STS role ARN used in examples. Not
// a real ARN — replace with the ARN of the IAM role you created following
// https://docs.aws.amazon.com/iot-twinmaker/latest/guide/dashboard-IAM-role.html.
const exampleAssumeRoleARN = "arn:aws:iam::123456789012:role/TwinMakerDashboardRole"

// exampleWriteRoleARN is a placeholder STS role ARN used in the Alarm
// Configuration Panel write examples.
const exampleWriteRoleARN = "arn:aws:iam::123456789012:role/TwinMakerAlarmWriter"

// SettingsExamples returns k8s-style example configurations for the IoT
// TwinMaker datasource, covering the default configuration, each
// editor-selectable AWS authentication provider, the AssumeRole variant,
// the Alarm Configuration write-permissions variant, and a legacy `arn`
// example. Each example value is a full instance settings object with the
// plugin configuration under jsonData and the relevant write-only secrets
// under secureJsonData (placeholder values — replace them with real
// secrets and IDs).
func SettingsExamples() *sdkschema.SettingsExamples {
	return &sdkschema.SettingsExamples{
		Examples: map[string]*spec3.Example{
			"": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Default configuration",
					Description: "The defaults a fresh datasource starts with: AWS SDK Default credentials, defaultRegion 'us-east-1' (set on mount by the editor and again by the backend Load). workspaceId and assumeRoleArn are still required before the datasource passes CheckHealth; secureJsonData is empty because the default provider reads credentials from the AWS SDK chain.",
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
			"awsSdkDefault": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "AWS SDK Default",
					Description: "Uses the AWS SDK default credential chain (env vars, shared config, EC2/ECS/EKS metadata). The IoT TwinMaker dashboard IAM role in assumeRoleArn is required at runtime by CheckHealth.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType":      string(AWSAuthTypeDefault),
							"defaultRegion": "us-east-1",
							"workspaceId":   exampleWorkspaceID,
							"assumeRoleArn": exampleAssumeRoleARN,
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
					Description: "IAM user credentials supplied directly via secureJsonData. The IoT TwinMaker dashboard IAM role in assumeRoleArn is still required.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType":      string(AWSAuthTypeKeys),
							"defaultRegion": "us-east-1",
							"workspaceId":   exampleWorkspaceID,
							"assumeRoleArn": exampleAssumeRoleARN,
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
							"profile":       "my-twinmaker-profile",
							"defaultRegion": "us-east-1",
							"workspaceId":   exampleWorkspaceID,
							"assumeRoleArn": exampleAssumeRoleARN,
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
					Description: "Uses the IAM role attached to the Grafana workload (EC2 instance profile / ECS task role / EKS IRSA). Editor label is 'Workspace IAM Role'; storage value is 'ec2_iam_role'. Editor label is 'Workspace IAM Role' — no relation to the TwinMaker Workspace itself.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType":      string(AWSAuthTypeEC2IAMRole),
							"defaultRegion": "us-east-1",
							"workspaceId":   exampleWorkspaceID,
							"assumeRoleArn": exampleAssumeRoleARN,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAccessKey): "",
						},
					},
				},
			},
			"withExternalId": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "AWS SDK Default + STS AssumeRole with External ID",
					Description: "Cross-account AssumeRole flow: the External ID is passed to STS when assuming the dashboard IAM role. External ID is only meaningful when assumeRoleArn is set.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType":      string(AWSAuthTypeDefault),
							"defaultRegion": "us-east-1",
							"workspaceId":   exampleWorkspaceID,
							"assumeRoleArn": exampleAssumeRoleARN,
							"externalId":    "external-id-abc123",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAccessKey): "",
						},
					},
				},
			},
			"withAlarmWriteRole": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Alarm Configuration Panel write permissions",
					Description: "Sets assumeRoleArnWriter so the AWS IoT TwinMaker Alarm Configuration Panel can write property values back to IoT TwinMaker. On load, the editor's 'Define write permissions for Alarm Configuration Panel' switch is derived to on because assumeRoleArnWriter is non-empty.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType":            string(AWSAuthTypeDefault),
							"defaultRegion":       "us-east-1",
							"workspaceId":         exampleWorkspaceID,
							"assumeRoleArn":       exampleAssumeRoleARN,
							"assumeRoleArnWriter": exampleWriteRoleARN,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAccessKey): "",
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
							"workspaceId":   exampleWorkspaceID,
							"assumeRoleArn": exampleAssumeRoleARN,
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
