export type cloudWatchConfig = {
  jsonData: {
    /**
     * REQUIRED.
     *
     * AWS authentication method for connecting to CloudWatch APIs.
     *
     * Values:
     * - "default": AWS SDK default credential provider chain (env vars, ~/.aws/credentials, EC2 instance profile, etc.).
     * - "credentials": Shared credentials file (~/.aws/credentials) with an optional named profile.
     * - "keys": Static access key and secret key (provided in secureJsonData).
     * - "ec2_iam_role": EC2 IAM role (Grafana must run on an EC2 instance with an attached role).
     * - "grafana_assume_role": Grafana Cloud only — uses Grafana-managed assume-role credentials.
     * - "arn": DEPRECATED since Grafana 7.3 — falls back to "default".
     *
     * Backend behavior:
     * - Parsed via AWSDatasourceSettings.Load(); legacy "arn" is mapped to AuthTypeDefault.
     *   https://github.com/grafana/grafana-aws-sdk/blob/main/pkg/awsds/settings.go#L66-L91
     * - If authType is not in the server's allowed_auth_providers list, backend returns "trying to use non-allowed auth method".
     *   https://github.com/grafana/grafana/blob/main/pkg/tsdb/cloudwatch/cloudwatch.go#L64-L86
     *
     * UI hints:
     * - Rendered via ConnectionConfig from @grafana/aws-sdk; shows dropdown with authentication type options.
     *   https://github.com/grafana/grafana/blob/main/public/app/plugins/datasource/cloudwatch/components/ConfigEditor/ConfigEditor.tsx#L92-L127
     * - "arn" triggers a deprecation warning: "authentication type 'arn' is deprecated, falling back to default SDK provider".
     *   https://github.com/grafana/grafana/blob/main/public/app/plugins/datasource/cloudwatch/components/ConfigEditor/ConfigEditor.tsx#L23-L40
     *
     * Provisioning examples:
     * https://github.com/grafana/grafana/blob/main/docs/sources/datasources/aws-cloudwatch/configure/index.md#L293-L351
     */
    authType?: CloudWatchAuthType;

    /**
     * REQUIRED.
     *
     * Default AWS region used when queries don't specify a region or specify "default".
     *
     * Backend behavior:
     * - If region is "default" or empty, this value is used as the effective region. If both are empty, errors with ErrMissingRegion.
     *   https://github.com/grafana/grafana/blob/main/pkg/tsdb/cloudwatch/cloudwatch.go#L64-L86
     * - AWSDatasourceSettings.Load() copies DefaultRegion to Region if Region is empty or "default".
     *   https://github.com/grafana/grafana-aws-sdk/blob/main/pkg/awsds/settings.go#L118-L141
     *
     * UI hints:
     * - Shown as "Default Region" in the ConnectionConfig component; regions are loaded dynamically from AWS.
     *   https://github.com/grafana/grafana/blob/main/public/app/plugins/datasource/cloudwatch/components/ConfigEditor/ConfigEditor.tsx#L92-L127
     *
     * Provisioning examples:
     * https://github.com/grafana/grafana/blob/main/docs/sources/datasources/aws-cloudwatch/configure/index.md#L293-L322
     */
    defaultRegion?: string;

    /**
     * OPTIONAL.
     *
     * Named AWS credentials profile from ~/.aws/credentials. Used when authType is "credentials".
     *
     * Backend behavior:
     * - Falls back to the legacy `database` field of the datasource config if not set.
     *   https://github.com/grafana/grafana-aws-sdk/blob/main/pkg/awsds/settings.go#L118-L141
     *
     * Provisioning examples:
     * https://github.com/grafana/grafana/blob/main/docs/sources/datasources/aws-cloudwatch/configure/index.md#L323-L351
     */
    profile?: string;

    /**
     * OPTIONAL.
     *
     * ARN of the IAM role to assume. Used for cross-account access.
     *
     * Backend behavior:
     * - If set and assume_role_enabled is false in Grafana config, errors with "trying to use assume role but it is disabled".
     *   https://github.com/grafana/grafana/blob/main/pkg/tsdb/cloudwatch/cloudwatch.go#L64-L86
     *
     * UI hints:
     * - Shown as "Assume Role ARN" in the ConnectionConfig component.
     *   https://github.com/grafana/grafana/blob/main/public/app/plugins/datasource/cloudwatch/components/ConfigEditor/ConfigEditor.tsx#L92-L127
     *
     * Provisioning examples:
     * https://github.com/grafana/grafana/blob/main/docs/sources/datasources/aws-cloudwatch/configure/index.md#L323-L351
     */
    assumeRoleArn?: string;

    /**
     * OPTIONAL.
     *
     * External ID for STS AssumeRole. Required by some AWS accounts for cross-account access.
     *
     * Backend behavior:
     * - Passed to STS AssumeRole call via the ExternalID option.
     *   https://github.com/grafana/grafana-aws-sdk/blob/main/pkg/awsauth/settings.go#L192-L218
     *
     * UI hints:
     * - Shown as "External ID" in the ConnectionConfig component.
     *   https://github.com/grafana/grafana/blob/main/docs/sources/datasources/aws-cloudwatch/configure/index.md#L80-L88
     */
    externalId?: string;

    /**
     * OPTIONAL.
     *
     * Custom endpoint URL for the CloudWatch API. Overrides the default AWS endpoint for the region.
     *
     * Backend behavior:
     * - Used in the AWS config provider to override the service endpoint.
     *   https://github.com/grafana/grafana/blob/main/pkg/tsdb/cloudwatch/cloudwatch.go#L64-L86
     *
     * UI hints:
     * - Shown as "Endpoint" under Additional Settings.
     *   https://github.com/grafana/grafana/blob/main/docs/sources/datasources/aws-cloudwatch/configure/index.md#L80-L88
     */
    endpoint?: string;

    /**
     * OPTIONAL.
     *
     * Comma-separated list of custom metric namespaces to include in the namespace dropdown.
     * AWS custom namespaces are not automatically discoverable via the GetMetricData API.
     *
     * Backend behavior:
     * - Stored as CloudWatchSettings.Namespace (JSON tag: "customMetricsNamespaces").
     *   https://github.com/grafana/grafana/blob/main/pkg/tsdb/cloudwatch/models/settings.go#L0-L25
     *
     * UI hints:
     * - Shown as "Namespaces of Custom Metrics" with placeholder "Namespace1,Namespace2".
     *   https://github.com/grafana/grafana/blob/main/public/app/plugins/datasource/cloudwatch/components/ConfigEditor/ConfigEditor.tsx#L92-L127
     *
     * Docs:
     * https://github.com/grafana/grafana/blob/main/docs/sources/datasources/aws-cloudwatch/configure/index.md#L88-L94
     */
    customMetricsNamespaces?: string;

    /**
     * OPTIONAL.
     *
     * Timeout duration for CloudWatch Logs queries. Grafana polls every second until AWS returns
     * "Done" status or this timeout is exceeded. Must be a valid Go/Grafana duration string
     * (e.g., "30m", "30s", "2000ms").
     *
     * Backend behavior:
     * - Defaults to 30 minutes if empty or not set.
     *   https://github.com/grafana/grafana/blob/main/pkg/tsdb/cloudwatch/models/settings.go#L26-L47
     * - Supports duration strings, float seconds (e.g., "1.5s"), and nanosecond integers.
     *   https://github.com/grafana/grafana/blob/main/pkg/tsdb/cloudwatch/models/settings.go#L51-L77
     * - Invalid values (e.g., "10mm", booleans) cause downstream errors.
     *   https://github.com/grafana/grafana/blob/main/pkg/tsdb/cloudwatch/models/settings_test.go#L184-L223
     * - For alerting queries, the Grafana evaluation_timeout_seconds takes precedence (default: 30s).
     *
     * UI hints:
     * - Shown as "Query Result Timeout" under CloudWatch Logs section with placeholder "30m".
     *   https://github.com/grafana/grafana/blob/main/public/app/plugins/datasource/cloudwatch/components/ConfigEditor/ConfigEditor.tsx#L127-L144
     * - Validated client-side using rangeUtil.describeInterval; invalid values show an error.
     *   https://github.com/grafana/grafana/blob/main/public/app/plugins/datasource/cloudwatch/components/ConfigEditor/ConfigEditor.tsx#L206-L246
     *
     * Defaults:
     * - "30m" (30 minutes).
     */
    logsTimeout?: string;

    /**
     * OPTIONAL.
     *
     * Default log groups pre-selected in the CloudWatch Logs query editor. Each entry is an object
     * with `arn` and `name` fields.
     *
     * UI hints:
     * - Shown as "Default Log Groups" in the CloudWatch Logs config section. Uses a log group selector
     *   that requires the datasource to be saved before adding groups.
     *   https://github.com/grafana/grafana/blob/main/public/app/plugins/datasource/cloudwatch/components/ConfigEditor/ConfigEditor.tsx#L144-L174
     * - Pre-fills the logGroups field in new Logs Insights queries.
     *   https://github.com/grafana/grafana/blob/main/public/app/plugins/datasource/cloudwatch/defaultQueries.ts#L35-L49
     *
     * Docs:
     * https://github.com/grafana/grafana/blob/main/docs/sources/datasources/aws-cloudwatch/configure/index.md#L94-L102
     */
    logGroups?: CloudWatchLogGroup[];

    /**
     * DEPRECATED: use logGroups.
     *
     * Legacy list of default log group names (string array). Migrated to logGroups (with ARNs) when
     * the ConfigEditor is opened or when queries are executed.
     *
     * Backend behavior:
     * - Still read by the datasource constructor for backward compatibility.
     *   https://github.com/grafana/grafana/blob/main/public/app/plugins/datasource/cloudwatch/datasource.ts#L87-L93
     *
     * UI hints:
     * - When present, the LogGroupsField component migrates them to the logGroups format via backend lookups.
     *   https://github.com/grafana/grafana/blob/main/public/app/plugins/datasource/cloudwatch/components/ConfigEditor/ConfigEditor.tsx#L174-L194
     */
    defaultLogGroups?: string[];

    /**
     * OPTIONAL.
     *
     * UID of an X-Ray or Application Signals datasource. Used to create trace links from CloudWatch
     * Logs results that contain traceId fields.
     *
     * UI hints:
     * - Configured via XrayLinkConfig component under "Application Signals trace link" section.
     *   https://github.com/grafana/grafana/blob/main/public/app/plugins/datasource/cloudwatch/components/ConfigEditor/ConfigEditor.tsx#L174-L194
     *
     * Backend behavior:
     * - Read by CloudWatchLogsQueryRunner to attach trace data links to log frames.
     *   https://github.com/grafana/grafana/blob/main/public/app/plugins/datasource/cloudwatch/query-runner/CloudWatchLogsQueryRunner.ts#L73-L79
     *
     * Docs:
     * https://github.com/grafana/grafana/blob/main/docs/sources/datasources/aws-cloudwatch/configure/index.md#L94-L102
     */
    tracingDatasourceUid?: string;

    /**
     * OPTIONAL.
     *
     * Enable the Grafana secure SOCKS datasource proxy for this datasource.
     *
     * Backend behavior:
     * - Stored as CloudWatchSettings.SecureSocksProxyEnabled.
     *   https://github.com/grafana/grafana/blob/main/pkg/tsdb/cloudwatch/models/settings.go#L0-L25
     * - Only effective when the Grafana server has secureSocksDSProxyEnabled set to true.
     *   https://github.com/grafana/grafana/blob/main/pkg/tsdb/cloudwatch/cloudwatch.go#L64-L86
     *
     * UI hints:
     * - Shown only when config.secureSocksDSProxyEnabled is true on the Grafana instance.
     *   https://github.com/grafana/grafana/blob/main/public/app/plugins/datasource/cloudwatch/components/ConfigEditor/ConfigEditor.tsx#L127-L144
     */
    enableSecureSocksProxy?: boolean;

    /**
     * OPTIONAL.
     *
     * Timestamp field name used internally. Rarely configured directly.
     *
     * UI hints:
     * - Included in ConfigEditor test fixtures with default value "@timestamp".
     *   https://github.com/grafana/grafana/blob/main/public/app/plugins/datasource/cloudwatch/components/ConfigEditor/ConfigEditor.test.tsx#L68-L110
     */
    timeField?: string;

    /**
     * OPTIONAL.
     *
     * Database field, used as legacy fallback for the credentials profile name.
     *
     * Backend behavior:
     * - AWSDatasourceSettings.Load() uses config.Database as fallback for Profile if profile is empty.
     *   https://github.com/grafana/grafana-aws-sdk/blob/main/pkg/awsds/settings.go#L118-L141
     */
    database?: string;

    /**
     * OPTIONAL.
     *
     * HTTP proxy type for AWS requests.
     *
     * Values:
     * - "none": No proxy.
     * - "env": Use environment variable proxy settings (default).
     * - "url": Use a custom proxy URL specified in proxyUrl.
     *
     * Backend behavior:
     * - Parsed from AWSDatasourceSettings and forwarded to the HTTP client transport.
     *   https://github.com/grafana/grafana-aws-sdk/blob/main/pkg/awsds/settings.go#L96-L117
     */
    proxyType?: CloudWatchProxyType;

    /**
     * CONDITIONALLY REQUIRED: when proxyType is "url".
     *
     * Custom HTTP proxy URL for AWS API requests.
     *
     * Backend behavior:
     * - Used to configure a custom proxy dialer on the HTTP transport.
     *   https://github.com/grafana/grafana-aws-sdk/blob/main/pkg/awsauth/settings.go#L224-L245
     */
    proxyUrl?: string;

    /**
     * OPTIONAL.
     *
     * Username for proxy authentication (when proxyType is "url").
     *
     * Backend behavior:
     * - Read from AWSDatasourceSettings and forwarded to the proxy transport.
     *   https://github.com/grafana/grafana-aws-sdk/blob/main/pkg/awsds/settings.go#L96-L117
     */
    proxyUsername?: string;
  };

  secureJsonData: {
    /**
     * CONDITIONALLY REQUIRED: when authType is "keys".
     *
     * AWS access key ID for static credential authentication.
     *
     * Backend behavior:
     * - Read from DecryptedSecureJSONData["accessKey"] during settings load.
     *   https://github.com/grafana/grafana-aws-sdk/blob/main/pkg/awsds/settings.go#L118-L141
     *
     * UI hints:
     * - Rendered as a SecretInput in the ConnectionConfig component.
     *
     * Provisioning examples:
     * https://github.com/grafana/grafana/blob/main/docs/sources/datasources/aws-cloudwatch/configure/index.md#L323-L351
     */
    accessKey?: string;

    /**
     * CONDITIONALLY REQUIRED: when authType is "keys".
     *
     * AWS secret access key for static credential authentication.
     *
     * Backend behavior:
     * - Read from DecryptedSecureJSONData["secretKey"] during settings load.
     *   https://github.com/grafana/grafana-aws-sdk/blob/main/pkg/awsds/settings.go#L118-L141
     *
     * Provisioning examples:
     * https://github.com/grafana/grafana/blob/main/docs/sources/datasources/aws-cloudwatch/configure/index.md#L323-L351
     */
    secretKey?: string;

    /**
     * OPTIONAL.
     *
     * AWS session token for temporary credentials.
     *
     * Backend behavior:
     * - Read from DecryptedSecureJSONData["sessionToken"] during settings load.
     *   https://github.com/grafana/grafana-aws-sdk/blob/main/pkg/awsds/settings.go#L118-L141
     */
    sessionToken?: string;

    /**
     * OPTIONAL.
     *
     * Password for proxy authentication (when proxyType is "url" and proxy requires auth).
     *
     * Backend behavior:
     * - Read from DecryptedSecureJSONData["proxyPassword"] during settings load.
     *   https://github.com/grafana/grafana-aws-sdk/blob/main/pkg/awsds/settings.go#L118-L141
     */
    proxyPassword?: string;
  };
};

/**
 * AWS authentication type.
 *
 * Values map to the backend AuthType enum:
 * https://github.com/grafana/grafana-aws-sdk/blob/main/pkg/awsds/settings.go#L0-L21
 *
 * Note: "arn" is deprecated since Grafana 7.3 and falls back to "default".
 */
export type CloudWatchAuthType =
  | "default"
  | "credentials"
  | "keys"
  | "ec2_iam_role"
  | "grafana_assume_role"
  | "arn";

/**
 * HTTP proxy type for AWS requests.
 *
 * https://github.com/grafana/grafana-aws-sdk/blob/main/pkg/awsauth/settings.go#L32-L50
 */
export type CloudWatchProxyType = "none" | "env" | "url";

/**
 * Log group object with ARN and display name.
 *
 * https://github.com/grafana/grafana/blob/main/public/app/plugins/datasource/cloudwatch/dataquery.gen.ts#L0-L30
 */
export type CloudWatchLogGroup = {
  /** ARN of the CloudWatch log group. */
  arn: string;
  /** Display name of the log group. */
  name: string;
};
