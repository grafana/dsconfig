package cloudwatchdatasource

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
// truth for the CloudWatch datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the declarative dsconfig schema
// (single source of truth) for the CloudWatch datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the full k8s-style SDK plugin schema for the CloudWatch
// datasource: the settings (configuration) spec derived from dsconfig.json,
// the secure values, and example configurations, stamped with
// TargetAPIVersion. Grafana's datasource API server serves this bundle as
// {TargetAPIVersion}.json.
func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

// SettingsExamples returns k8s-style example configurations for the
// CloudWatch datasource, covering the default configuration and each
// editor-selectable AWS authentication provider plus a Cloudwatch-Logs-
// oriented variant (default log groups, custom logsTimeout, tracing
// datasource link) and the legacy `arn` value. Each example is a full
// instance-settings object with the plugin configuration under jsonData and
// the relevant write-only secrets under secureJsonData (placeholder values —
// replace them with real secrets).
func SettingsExamples() *sdkschema.SettingsExamples {
	return &sdkschema.SettingsExamples{
		Examples: map[string]*spec3.Example{
			"": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Default configuration",
					Description: "The defaults a fresh datasource starts with: AWS SDK Default credentials, environment proxy strategy, and the 30-minute Cloudwatch Logs polling timeout. `defaultRegion` still needs to be set before queries will run; secureJsonData is empty because the default provider reads credentials from the AWS SDK chain.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType":    string(AWSAuthTypeDefault),
							"proxyType":   string(AWSProxyTypeEnv),
							"logsTimeout": "30m0s",
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
							"authType":                string(AWSAuthTypeDefault),
							"defaultRegion":           "us-east-1",
							"proxyType":               string(AWSProxyTypeEnv),
							"logsTimeout":             "30m0s",
							"customMetricsNamespaces": "AWS/EC2,AWS/ELB",
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
							"proxyType":     string(AWSProxyTypeEnv),
							"logsTimeout":   "30m0s",
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
							"profile":       "my-cloudwatch-profile",
							"defaultRegion": "us-east-1",
							"proxyType":     string(AWSProxyTypeEnv),
							"logsTimeout":   "30m0s",
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
							"proxyType":     string(AWSProxyTypeEnv),
							"logsTimeout":   "30m0s",
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
							"proxyType":     string(AWSProxyTypeEnv),
							"logsTimeout":   "30m0s",
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
							"assumeRoleArn": "arn:aws:iam::123456789012:role/GrafanaCloudWatch",
							"externalId":    "external-id-abc123",
							"defaultRegion": "us-east-1",
							"proxyType":     string(AWSProxyTypeEnv),
							"logsTimeout":   "30m0s",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAccessKey): "AKIAIOSFODNN7EXAMPLE",
							string(SecureJsonDataKeySecretKey): "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
						},
					},
				},
			},
			"urlProxy": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "URL HTTP proxy",
					Description: "Routes AWS SDK traffic through an explicit proxy URL. Editor-visible only when both `showHttpProxySettings` is passed by the plugin (CloudWatch does) AND the runtime `awsPerDatasourceHTTPProxyEnabled` feature toggle is on. Pair with authType 'keys' or any other credential provider — proxy settings apply regardless.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType":      string(AWSAuthTypeDefault),
							"defaultRegion": "us-east-1",
							"proxyType":     string(AWSProxyTypeURL),
							"proxyUrl":      "https://proxy.internal:3128",
							"proxyUsername": "grafana",
							"logsTimeout":   "30m0s",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyProxyPassword): "proxy-password-placeholder",
						},
					},
				},
			},
			"cloudwatchLogsDefaults": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Default log groups + custom logs timeout + Application Signals link",
					Description: "Configures CloudWatch Logs default log groups (modern object shape written by the LogGroupsField when the cloudWatchCrossAccountQuerying feature toggle is on), a shorter logs polling timeout, and a link to a grafana-x-ray-datasource instance for Application Signals (formerly X-Ray) trace links from log entries containing an @xrayTraceId field.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType":                string(AWSAuthTypeDefault),
							"defaultRegion":           "us-east-1",
							"proxyType":               string(AWSProxyTypeEnv),
							"customMetricsNamespaces": "AWS/EC2,AWS/RDS,MyApp/Custom",
							"logsTimeout":             "10m",
							"logGroups": []map[string]any{
								{
									"arn":  "arn:aws:logs:us-east-1:123456789012:log-group:/aws/lambda/example",
									"name": "/aws/lambda/example",
								},
								{
									"arn":          "arn:aws:logs:us-east-1:210987654321:log-group:/aws/lambda/cross-account",
									"name":         "/aws/lambda/cross-account",
									"accountId":    "210987654321",
									"accountLabel": "production",
								},
							},
							"tracingDatasourceUid": "aws-xray-uid",
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
					Description: "Datasources provisioned before the 'arn' value was renamed still store `authType: \"arn\"`. The backend (awsds.AuthType.UnmarshalJSON, awsds/settings.go:87-88) maps this to the modern 'default' at load time; the editor surfaces a deprecation warning banner (`ARN_DEPRECATION_WARNING_MESSAGE`).",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType":      string(AWSAuthTypeARN),
							"defaultRegion": "us-east-1",
							"proxyType":     string(AWSProxyTypeEnv),
							"logsTimeout":   "30m0s",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAccessKey): "",
						},
					},
				},
			},
			"legacyDefaultLogGroups": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Legacy default log groups (defaultLogGroups)",
					Description: "The pre-cross-account log group shape: a plain array of log group name strings written by `LegacyLogGroupSelection` when the `cloudWatchCrossAccountQuerying` feature toggle is off. The editor migrates this into `logGroups` on first open.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType":      string(AWSAuthTypeDefault),
							"defaultRegion": "us-east-1",
							"proxyType":     string(AWSProxyTypeEnv),
							"logsTimeout":   "30m0s",
							"defaultLogGroups": []string{
								"/aws/lambda/example",
								"/aws/lambda/other",
							},
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
