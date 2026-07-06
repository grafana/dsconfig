package auroradatasource

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
// truth for the Aurora datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the declarative dsconfig schema
// (single source of truth) for the Aurora datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the full k8s-style SDK plugin schema for the Aurora
// datasource: the settings (configuration) spec derived from dsconfig.json,
// the secure values, and example configurations, stamped with
// TargetAPIVersion. Grafana's datasource API server serves this bundle as
// {TargetAPIVersion}.json.
func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

// SettingsExamples returns k8s-style example configurations for the Aurora
// datasource, covering the default configuration and each editor-selectable
// AWS authentication provider crossed with each Aurora engine plus the
// split-auth-endpoint variant. Each example value is a full instance
// settings object with the plugin configuration under jsonData and the
// relevant write-only secrets under secureJsonData (placeholder values —
// replace them with real secrets).
func SettingsExamples() *sdkschema.SettingsExamples {
	return &sdkschema.SettingsExamples{
		Examples: map[string]*spec3.Example{
			"": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Default configuration",
					Description: "The defaults a fresh datasource starts with: AWS SDK Default credentials and the Aurora Postgres engine. The Aurora selectors (defaultRegion, dbHost, dbPort, dbUser) still need to be filled in before queries will run; secureJsonData is empty because the default provider reads credentials from the AWS SDK chain and Aurora authenticates to the database with a generated RDS IAM auth token rather than a password.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType": string(AWSAuthTypeDefault),
							"engine":   string(AuroraEnginePostgres),
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAccessKey): "",
						},
					},
				},
			},
			"awsSdkDefaultPostgres": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "AWS SDK Default + Aurora (PostgreSQL Compatible)",
					Description: "Uses the AWS SDK default credential chain (env vars, shared config, EC2/ECS/EKS metadata) against an Aurora PostgreSQL-compatible cluster on port 5432. No secrets are set in secureJsonData; the placeholder accessKey is empty to signal that.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType":      string(AWSAuthTypeDefault),
							"defaultRegion": "us-east-1",
							"engine":        string(AuroraEnginePostgres),
							"dbHost":        "my-cluster.cluster-xxxxxxxxxxxx.us-east-1.rds.amazonaws.com",
							"dbPort":        5432,
							"dbName":        "mydb",
							"dbUser":        "iam_user",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAccessKey): "",
						},
					},
				},
			},
			"accessAndSecretKeyMysql": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Access & secret key + Aurora (MySQL Compatible)",
					Description: "IAM user credentials supplied directly via secureJsonData against an Aurora MySQL-compatible cluster on port 3306. Optionally include sessionToken for temporary STS credentials.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType":      string(AWSAuthTypeKeys),
							"defaultRegion": "us-east-1",
							"engine":        string(AuroraEngineMySQL),
							"dbHost":        "my-cluster.cluster-xxxxxxxxxxxx.us-east-1.rds.amazonaws.com",
							"dbPort":        3306,
							"dbName":        "mydb",
							"dbUser":        "iam_user",
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
							"profile":       "my-aurora-profile",
							"defaultRegion": "us-east-1",
							"engine":        string(AuroraEnginePostgres),
							"dbHost":        "my-cluster.cluster-xxxxxxxxxxxx.us-east-1.rds.amazonaws.com",
							"dbPort":        5432,
							"dbName":        "mydb",
							"dbUser":        "iam_user",
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
					Description: "Uses the IAM role attached to the Grafana workload (EC2 instance profile / ECS task role / EKS IRSA). Editor label is 'Workspace IAM Role'; storage value is 'ec2_iam_role'. The role must have `rds-db:connect` on the target database user for the RDS auth token to be accepted.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType":      string(AWSAuthTypeEC2IAMRole),
							"defaultRegion": "us-east-1",
							"engine":        string(AuroraEnginePostgres),
							"dbHost":        "my-cluster.cluster-xxxxxxxxxxxx.us-east-1.rds.amazonaws.com",
							"dbPort":        5432,
							"dbName":        "mydb",
							"dbUser":        "iam_user",
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
					Description: "Grafana Cloud's temporary-credentials broker. Feature-gated on the awsDatasourcesTempCredentials toggle; the editor hides assumeRoleArn/externalId/endpoint when this provider is selected because Grafana derives them itself.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType":      string(AWSAuthTypeGrafanaAssumeRole),
							"defaultRegion": "us-east-1",
							"engine":        string(AuroraEnginePostgres),
							"dbHost":        "my-cluster.cluster-xxxxxxxxxxxx.us-east-1.rds.amazonaws.com",
							"dbPort":        5432,
							"dbName":        "mydb",
							"dbUser":        "iam_user",
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
					Description: "IAM user credentials that then assume a cross-account IAM role via STS. `externalId` is optional and only meaningful when `assumeRoleArn` is set. The assumed role must be granted `rds-db:connect` on the target database user.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType":      string(AWSAuthTypeKeys),
							"assumeRoleArn": "arn:aws:iam::123456789012:role/GrafanaAurora",
							"externalId":    "external-id-abc123",
							"defaultRegion": "us-east-1",
							"engine":        string(AuroraEnginePostgres),
							"dbHost":        "my-cluster.cluster-xxxxxxxxxxxx.us-east-1.rds.amazonaws.com",
							"dbPort":        5432,
							"dbName":        "mydb",
							"dbUser":        "iam_user",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAccessKey): "AKIAIOSFODNN7EXAMPLE",
							string(SecureJsonDataKeySecretKey): "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
						},
					},
				},
			},
			"splitAuthEndpoint": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Split auth-token endpoint (behind a load balancer)",
					Description: "Grafana connects to Aurora through a load balancer. jsonData.dbHost/dbPort is the LB address used for SQL traffic; jsonData.dbHostAuth (and optionally dbPortAuth) points at the primary cluster endpoint so the RDS `generate-db-auth-token` call is signed with the endpoint AWS recognises. Backend picks up the override at pkg/plugin/connect.go:59-66.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType":      string(AWSAuthTypeEC2IAMRole),
							"defaultRegion": "us-east-1",
							"engine":        string(AuroraEnginePostgres),
							"dbHost":        "aurora-lb.internal",
							"dbPort":        5432,
							"dbHostAuth":    "my-cluster.cluster-xxxxxxxxxxxx.us-east-1.rds.amazonaws.com",
							"dbPortAuth":    5432,
							"dbName":        "mydb",
							"dbUser":        "iam_user",
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
					Description: "Datasources provisioned before the 'arn' value was renamed still store `authType: \"arn\"`. The shared awsds layer (awsds.AuthType.UnmarshalJSON, awsds/settings.go:87-88) maps this to the modern 'default' at load time.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType":      string(AWSAuthTypeARN),
							"defaultRegion": "us-east-1",
							"engine":        string(AuroraEnginePostgres),
							"dbHost":        "my-cluster.cluster-xxxxxxxxxxxx.us-east-1.rds.amazonaws.com",
							"dbPort":        5432,
							"dbName":        "mydb",
							"dbUser":        "iam_user",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAccessKey): "",
						},
					},
				},
			},
			"legacyMissingEngine": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Legacy: missing engine (pre-multi-engine beta)",
					Description: "Aurora's original beta shipped without an engine field. The backend falls back to aurora-postgres at connect time (pkg/plugin/connect.go:83-85, 135-138) to keep those datasources working. New configurations should always set jsonData.engine explicitly; LoadConfig's ApplyDefaults sets it to aurora-postgres for the same reason.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType":      string(AWSAuthTypeDefault),
							"defaultRegion": "us-east-1",
							"dbHost":        "my-cluster.cluster-xxxxxxxxxxxx.us-east-1.rds.amazonaws.com",
							"dbPort":        5432,
							"dbName":        "mydb",
							"dbUser":        "iam_user",
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
