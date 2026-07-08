package redshiftdatasource

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
// truth for the Redshift datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the declarative dsconfig schema
// (single source of truth) for the Redshift datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the full k8s-style SDK plugin schema for the Redshift
// datasource: the settings (configuration) spec derived from dsconfig.json,
// the secure values, and example configurations, stamped with
// TargetAPIVersion. Grafana's datasource API server serves this bundle as
// {TargetAPIVersion}.json.
func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

// SettingsExamples returns k8s-style example configurations for the Redshift
// datasource. It covers the default configuration and the four (provisioning
// × credential mode) matrix quadrants, plus one example per AWS auth
// provider and a legacy `arn` example. Each example is a full instance-
// settings object with the plugin configuration under jsonData and the
// relevant write-only secrets under secureJsonData (placeholder values —
// replace them with real secrets).
func SettingsExamples() *sdkschema.SettingsExamples {
	return &sdkschema.SettingsExamples{
		Examples: map[string]*spec3.Example{
			"": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Default configuration",
					Description: "The defaults a fresh datasource starts with: AWS SDK Default credentials, Provisioned shape, and temporary IAM credentials. Cluster identifier, database, and dbUser still need to be picked from the connected AWS account before queries will run; secureJsonData is empty because the default provider reads credentials from the AWS SDK chain.",
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
			"provisionedTempCredsKeys": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Provisioned + Temporary credentials + Access & secret key",
					Description: "Redshift Provisioned cluster with temporary IAM credentials minted via GetClusterCredentials. `dbUser` is required to name the database user the temporary password is issued for. AWS auth via IAM user access key + secret key.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType":          string(AWSAuthTypeKeys),
							"defaultRegion":     "us-east-1",
							"useServerless":     false,
							"useManagedSecret":  false,
							"clusterIdentifier": "my-redshift-cluster",
							"database":          "dev",
							"dbUser":            "awsuser",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAccessKey): "AKIAIOSFODNN7EXAMPLE",
							string(SecureJsonDataKeySecretKey): "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
						},
					},
				},
			},
			"provisionedManagedSecretDefault": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Provisioned + AWS Secrets Manager + AWS SDK Default",
					Description: "Redshift Provisioned cluster whose database credentials are read from an AWS Secrets Manager secret. Both dbUser and password come from the secret; the editor also stores managedSecret.name from the Select label.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType":          string(AWSAuthTypeDefault),
							"defaultRegion":     "us-east-1",
							"useServerless":     false,
							"useManagedSecret":  true,
							"clusterIdentifier": "my-redshift-cluster",
							"database":          "dev",
							"managedSecret": map[string]any{
								"arn":  "arn:aws:secretsmanager:us-east-1:123456789012:secret:redshift-1-xxxxxx",
								"name": "redshift-1",
							},
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAccessKey): "",
						},
					},
				},
			},
			"serverlessTempCredsIamRole": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Serverless + Temporary credentials + Workspace IAM Role",
					Description: "Redshift Serverless workgroup with temporary IAM credentials minted via GetCredentials. Serverless does not require dbUser because GetCredentials issues both the username and password. AWS auth via the workload's own IAM role.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType":         string(AWSAuthTypeEC2IAMRole),
							"defaultRegion":    "us-east-1",
							"useServerless":    true,
							"useManagedSecret": false,
							"workgroupName":    "default",
							"database":         "dev",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAccessKey): "",
						},
					},
				},
			},
			"serverlessManagedSecretGrafanaAssume": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Serverless + AWS Secrets Manager + Grafana Assume Role (Grafana Cloud only)",
					Description: "Redshift Serverless workgroup whose credentials are read from AWS Secrets Manager. AWS auth via Grafana Cloud's temporary-credentials broker (feature-gated on the awsDatasourcesTempCredentials toggle — the editor hides assumeRoleArn/externalId/endpoint when this provider is selected).",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType":         string(AWSAuthTypeGrafanaAssumeRole),
							"defaultRegion":    "us-east-1",
							"useServerless":    true,
							"useManagedSecret": true,
							"workgroupName":    "default",
							"database":         "dev",
							"managedSecret": map[string]any{
								"arn":  "arn:aws:secretsmanager:us-east-1:123456789012:secret:redshift-serverless-xxxxxx",
								"name": "redshift-serverless",
							},
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAccessKey): "",
						},
					},
				},
			},
			"credentialsFile": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Credentials file",
					Description: "Reads a named profile from ~/.aws/credentials on the Grafana host. Provisioned + temporary IAM credentials. Leave `profile` blank to pick the default profile.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType":          string(AWSAuthTypeCredentials),
							"profile":           "my-redshift-profile",
							"defaultRegion":     "us-east-1",
							"useServerless":     false,
							"useManagedSecret":  false,
							"clusterIdentifier": "my-redshift-cluster",
							"database":          "dev",
							"dbUser":            "awsuser",
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
					Description: "IAM user credentials that then assume a cross-account IAM role via STS to reach Redshift. `externalId` is optional and only meaningful when `assumeRoleArn` is set.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType":          string(AWSAuthTypeKeys),
							"assumeRoleArn":     "arn:aws:iam::123456789012:role/GrafanaRedshift",
							"externalId":        "external-id-abc123",
							"defaultRegion":     "us-east-1",
							"useServerless":     false,
							"useManagedSecret":  false,
							"clusterIdentifier": "my-redshift-cluster",
							"database":          "dev",
							"dbUser":            "awsuser",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAccessKey): "AKIAIOSFODNN7EXAMPLE",
							string(SecureJsonDataKeySecretKey): "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
						},
					},
				},
			},
			"withEventBridge": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Send Redshift events to Amazon EventBridge",
					Description: "Same as the Provisioned + temp-creds + keys example, with jsonData.withEvent=true so the plugin publishes query execution events to Amazon EventBridge (ConfigEditor.tsx:368-384). Requires the IAM role/user to have events:PutEvents permission.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType":          string(AWSAuthTypeKeys),
							"defaultRegion":     "us-east-1",
							"useServerless":     false,
							"useManagedSecret":  false,
							"clusterIdentifier": "my-redshift-cluster",
							"database":          "dev",
							"dbUser":            "awsuser",
							"withEvent":         true,
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
					Description: "Datasources provisioned before the 'arn' value was renamed still store `authType: \"arn\"`. The backend (awsds.AuthType.UnmarshalJSON, awsds/settings.go:87-88) maps this to the modern 'default' at load time.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType":          string(AWSAuthTypeARN),
							"defaultRegion":     "us-east-1",
							"useServerless":     false,
							"useManagedSecret":  false,
							"clusterIdentifier": "my-redshift-cluster",
							"database":          "dev",
							"dbUser":            "awsuser",
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
