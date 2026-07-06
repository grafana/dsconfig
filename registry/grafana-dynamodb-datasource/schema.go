package dynamodbdatasource

import (
	_ "embed"

	"k8s.io/kube-openapi/pkg/spec3"

	"github.com/grafana/dsconfig/dsconfig"
	sdkschema "github.com/grafana/grafana-plugin-sdk-go/experimental/pluginschema"
)

//go:generate go test -generateArtifacts -run TestSchemaConformance ./...

// TargetAPIVersion is the API version this schema applies to.
const TargetAPIVersion = dsconfig.TargetAPIVersion

// configSchemaJSON is the declarative dsconfig schema — the single
// source of truth for the DynamoDB datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the declarative dsconfig
// schema (single source of truth) for the DynamoDB datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the full k8s-style SDK plugin schema for the
// DynamoDB datasource: the settings (configuration) spec derived from
// dsconfig.json, the secure values, and example configurations,
// stamped with TargetAPIVersion. Grafana's datasource API server
// serves this bundle as {TargetAPIVersion}.json.
func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

// SettingsExamples returns k8s-style example configurations for the
// DynamoDB datasource, covering the default configuration and each
// editor-selectable AWS authentication provider, plus the legacy
// `arn` value and the pre-V2 storage shape. Each example is a full
// instance-settings object with the plugin configuration under
// jsonData and the relevant write-only secrets under secureJsonData
// (placeholder values — replace them with real secrets).
func SettingsExamples() *sdkschema.SettingsExamples {
	return &sdkschema.SettingsExamples{
		Examples: map[string]*spec3.Example{
			"": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Default configuration",
					Description: "The defaults a fresh datasource starts with under ApplyDefaults: AWS SDK Default credentials, isV2 marker unset (a fresh save from the editor would flip it to true; the schema-side default keeps it minimal). `defaultRegion` still needs to be set before queries will run; secureJsonData is empty because the default provider reads credentials from the AWS SDK chain.",
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
							"isV2":          true,
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
					Description: "IAM user credentials supplied directly via secureJsonData. Optionally include sessionToken for temporary STS credentials. This is what a fresh datasource looks like after the editor's useEffect writes `authType: 'keys'` and `isV2: true` and the user fills in an access key pair.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType":      string(AWSAuthTypeKeys),
							"isV2":          true,
							"defaultRegion": "us-east-1",
							"endpoint":      "https://dynamodb.us-east-1.amazonaws.com",
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
							"isV2":          true,
							"profile":       "my-dynamodb-profile",
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
							"isV2":          true,
							"defaultRegion": "us-east-1",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAccessKey): "",
						},
					},
				},
			},
			"keysWithSessionToken": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Access & secret key + STS session token",
					Description: "Temporary AWS STS credentials. sessionToken is backend-only (no editor UI); provisioning is the only way to set it. Included here so consumers can see the payload shape.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType":      string(AWSAuthTypeKeys),
							"isV2":          true,
							"defaultRegion": "us-east-1",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAccessKey):    "AKIAIOSFODNN7EXAMPLE",
							string(SecureJsonDataKeySecretKey):    "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
							string(SecureJsonDataKeySessionToken): "FQoGZXIvYXdzEExampleSTSSessionToken",
						},
					},
				},
			},
			"driverSettings": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Driver settings overrides",
					Description: "Overrides the default query timeout / retry count / retry pause. Values are stored as strings and parsed with utils.ParseInt server-side. There is no editor UI for these — they can only be set via provisioning.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType":      string(AWSAuthTypeDefault),
							"isV2":          true,
							"defaultRegion": "us-east-1",
							"timeout":       "120",
							"retries":       "10",
							"pause":         "2",
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
					Description: "Datasources provisioned before the 'arn' value was renamed still store `authType: \"arn\"`. The backend (awsds.AuthType.UnmarshalJSON, awsds/settings.go:87-88) maps this to the modern 'default' at load time.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType":      string(AWSAuthTypeARN),
							"isV2":          true,
							"defaultRegion": "us-east-1",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAccessKey): "",
						},
					},
				},
			},
			"legacyV1Shape": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Legacy V1 storage shape (pre-migration)",
					Description: "The pre-V2 storage layout carried by datasources provisioned before src/components/ConfigEditor.tsx:28-42 started writing isV2=true. The plain-text `accessId` in jsonData is the AWS Access Key ID; the `accessKey` key inside secureJsonData is actually the AWS Secret Access Key (a V1 naming quirk fixed at V2). `region` is used as the region because `defaultRegion` isn't set. On load the backend (pkg/models/settings.go:44-55) forces authType to 'keys', folds accessId into AccessKey, and treats secureJsonData.accessKey as the secret.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"region":   "eu-north-1",
							"endpoint": "https://dynamodb.eu-north-1.amazonaws.com",
							"accessId": "AKIAIOSFODNN7EXAMPLE",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAccessKey): "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
						},
					},
				},
			},
		},
	}
}
