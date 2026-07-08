/**
 * Configuration models for the Amazon CloudWatch datasource plugin
 * (plugin id `cloudwatch`).
 *
 * Sources of truth (all read at pinned versions):
 * - Plugin (https://github.com/grafana/grafana-cloudwatch-datasource@6e21d10):
 *   - `src/plugin.json` — plugin id (`cloudwatch`), name, docs link
 *   - `src/components/ConfigEditor/ConfigEditor.tsx` — every editor-visible
 *     field (Namespaces of Custom Metrics, Cloudwatch Logs section with
 *     Query Result Timeout + Default Log Groups, Application Signals trace
 *     link section) plus the embedded `@grafana/aws-sdk` ConnectionConfig
 *     called with `showHttpProxySettings` — CloudWatch is one of the few
 *     AWS datasources that opts in to the ConnectionConfig proxy subsection.
 *   - `src/components/ConfigEditor/XrayLinkConfig.tsx` — writes
 *     `jsonData.tracingDatasourceUid` via a DataSourcePicker.
 *   - `src/components/ConfigEditor/SecureSocksProxySettingsNewStyling.tsx` —
 *     writes `jsonData.enableSecureSocksProxy` (excluded per AGENTS.md).
 *   - `src/components/shared/LogGroups/LogGroupsField.tsx` — writes both
 *     `jsonData.logGroups` (new, LogGroup[]) and `jsonData.defaultLogGroups`
 *     (deprecated, string[]). Gated on the `cloudWatchCrossAccountQuerying`
 *     feature toggle for the modern selector.
 *   - `src/types.ts` — `CloudWatchJsonData extends AwsAuthDataSourceJsonData`
 *     (adds customMetricsNamespaces, logsTimeout, logGroups, defaultLogGroups,
 *     tracingDatasourceUid, timeField, database) and
 *     `CloudWatchSecureJsonData extends AwsAuthDataSourceSecureJsonData` (adds
 *     nothing).
 *   - `src/dataquery.ts` — `LogGroup` shape (arn, name, accountId?,
 *     accountLabel?).
 * - `@grafana/aws-sdk` `0.10.2` (grafana/grafana-aws-sdk-react tag `v0.10.2`,
 *   SHA `fe0c4d8`):
 *   - `src/types.ts` — `AwsAuthType`, `AwsAuthDataSourceJsonData`,
 *     `AwsAuthDataSourceSecureJsonData`, `ConnectionConfigProps`
 *   - `src/components/ConnectionConfig.tsx` — every AWS field's label,
 *     placeholder, description, and conditional render; the Proxy
 *     Configuration subsection (only visible when the calling plugin passes
 *     `showHttpProxySettings` AND the `awsPerDatasourceHTTPProxyEnabled`
 *     runtime toggle is on)
 *   - `src/providers.ts` — the Select options for `authType`
 * - Backend `grafana-aws-sdk` `v1.4.4` (`pkg/awsds/settings.go`):
 *   - `AWSDatasourceSettings` — Go struct that receives jsonData; note the
 *     `assumeRoleARN` (uppercase RN) json tag versus the frontend's
 *     `assumeRoleArn` (lowercase arn). Go's case-insensitive Unmarshal makes
 *     both work.
 * - Backend plugin (`pkg/cloudwatch/models/settings.go:16-24`):
 *   - `CloudWatchSettings` — embeds `awsds.AWSDatasourceSettings` and adds
 *     `Namespace string \`json:"customMetricsNamespaces"\``,
 *     `SecureSocksProxyEnabled bool \`json:"enableSecureSocksProxy"\`` (excluded),
 *     and `LogsTimeout Duration \`json:"logsTimeout"\`` (custom UnmarshalJSON
 *     accepts string durations and raw nanosecond numbers).
 */

/**
 * The AWS authentication provider values persisted to `jsonData.authType`.
 *
 * Five values are editor-selectable via the ConnectionConfig `Select`
 * (`providers.ts:4-25`): `ec2_iam_role`, `grafana_assume_role`, `default`,
 * `keys`, `credentials`. A sixth value, `arn`, is a legacy stored value the
 * backend maps to `default` (see `awsds.AuthType.UnmarshalJSON` /
 * `awsds/settings.go:87-88`).
 */
export type AwsAuthType =
  | 'default'
  | 'keys'
  | 'credentials'
  | 'ec2_iam_role'
  | 'grafana_assume_role'
  | 'arn';

/**
 * A single default log group entry stored in `jsonData.logGroups`. Mirrors
 * `LogGroup` from `src/dataquery.ts:326-343` in the CloudWatch plugin.
 */
export type LogGroup = {
  /** ARN of the log group. Required. */
  arn: string;
  /** Name of the log group. Required. */
  name: string;
  /** AccountId of the log group (only populated when cross-account querying is used). */
  accountId?: string;
  /** Label of the log group. */
  accountLabel?: string;
};

/**
 * Root (top-level datasource settings) fields.
 *
 * The CloudWatch datasource stores no plugin-specific fields at the root
 * level (`url`, `basicAuth`, etc. are unused; the root `database` field is
 * touched only by `awsds.Load` as a legacy profile fallback and is not
 * something the config editor writes), so this is a blank object rather than
 * null.
 */
export type RootConfig = Record<string, never>;

/**
 * Fields stored in `jsonData`. Union of:
 *  - the AWS-shared fields from `AwsAuthDataSourceJsonData`
 *    (`@grafana/aws-sdk@0.10.2`, `src/types.ts:15-25`), including the proxy
 *    fields because CloudWatch opts into `showHttpProxySettings`, and
 *  - the CloudWatch-specific fields from `CloudWatchJsonData`
 *    (`src/types.ts:36-51` of grafana-cloudwatch-datasource).
 *
 * `timeField` and `database` are declared in the plugin's TypeScript type but
 * are never written by the config editor and never read by the CloudWatch
 * backend (`pkg/cloudwatch/models/settings.go`), so they are omitted from the
 * schema and this model.
 */
export type JsonDataConfig = {
  // ---- AWS SDK ConnectionConfig fields (`@grafana/aws-sdk@0.10.2`) ----

  /** AWS credentials chain to use. Discriminator for accessKey/secretKey/profile/assumeRoleArn. */
  authType?: AwsAuthType;
  /** Credentials profile name from `~/.aws/credentials`. Editor-visible when `authType === 'credentials'`. */
  profile?: string;
  /**
   * ARN of an IAM role to assume via STS. Editor-visible when the caller
   * doesn't pass `hideAssumeRoleArn` (CloudWatch does not). Note: the
   * backend's json tag is `assumeRoleARN` (uppercase RN); Go's case-
   * insensitive Unmarshal makes the frontend spelling work too.
   */
  assumeRoleArn?: string;
  /** External ID passed to STS AssumeRole. Editor-visible when `authType !== 'grafana_assume_role'`. */
  externalId?: string;

  /**
   * HTTP proxy type used by the AWS SDK client. Editor-visible when
   * `showHttpProxySettings` (CloudWatch passes it) AND the
   * `awsPerDatasourceHTTPProxyEnabled` runtime toggle is on. Defaults to `env`.
   */
  proxyType?: 'none' | 'env' | 'url';
  /** Proxy URL. Editor-visible when `proxyType === 'url'`. */
  proxyUrl?: string;
  /** Proxy username. Editor-visible when `proxyType === 'url'`. */
  proxyUsername?: string;

  /** Optional custom AWS service endpoint. Editor-visible when `authType !== 'grafana_assume_role'`. */
  endpoint?: string;
  /** Default AWS region, e.g. `us-east-1`. Also used to seed the runtime `region` (backend). */
  defaultRegion?: string;

  // ---- CloudWatch-specific (`CloudWatchJsonData`) ----

  /**
   * Comma-separated list of custom-metric namespaces to expose in the query
   * editor. Editor label: "Namespaces of Custom Metrics"; placeholder:
   * `Namespace1,Namespace2`. Backend field name is `Namespace` with json
   * tag `customMetricsNamespaces` (`pkg/cloudwatch/models/settings.go:18`).
   */
  customMetricsNamespaces?: string;
  /**
   * Duration string used to bound how long the backend polls CloudWatch
   * Logs before returning a timeout error. Backend parses it as a
   * `time.Duration` (see the custom `UnmarshalJSON` on `Duration` at
   * `pkg/cloudwatch/models/settings.go:52-77`) — string values ("30m",
   * "2000ms", "1.5s") and raw nanosecond numbers are both accepted. Empty
   * / unset defaults to `30m` (`settings.go:42-44`). Editor renders it as a
   * text input; validated on the frontend via `rangeUtil.describeInterval`.
   */
  logsTimeout?: string;
  /**
   * Log groups selected as defaults for CloudWatch Logs queries. Written
   * by `LogGroupsField` when the `cloudWatchCrossAccountQuerying` feature
   * toggle is on. Not read by CloudWatchSettings — consumed at query time
   * by the logs query builder. `logGroups` supersedes `defaultLogGroups`
   * (the editor migrates the latter into the former on first open).
   */
  logGroups?: LogGroup[];
  /**
   * @deprecated Use `logGroups`. Legacy storage shape: an array of log
   * group name strings written by `LegacyLogGroupSelection` when the
   * `cloudWatchCrossAccountQuerying` feature toggle is off. Kept for round-
   * trip fidelity with older provisioned configs.
   */
  defaultLogGroups?: string[];
  /**
   * UID of a `grafana-x-ray-datasource` instance used to link log entries
   * that contain an `@xrayTraceId` field to a trace in Application Signals
   * (formerly X-Ray). Written by `XrayLinkConfig`; not read by the
   * CloudWatch backend — powers a frontend link only.
   */
  tracingDatasourceUid?: string;
};

/**
 * Secret key names stored in `secureJsonData` (write-only; read existing
 * config via `secureJsonFields`).
 *
 * - `accessKey` / `secretKey` — set when `authType === 'keys'`
 * - `sessionToken` — backend-only (`awsds/settings.go:137`); no editor UI
 *   writes it, provisioning-only
 * - `proxyPassword` — editor-visible only when both `showHttpProxySettings`
 *   AND `awsPerDatasourceHTTPProxyEnabled` are true; backend copies it from
 *   decrypted secure data (`awsds/settings.go:138`)
 */
export type SecureJsonDataConfig = Array<
  'accessKey' | 'secretKey' | 'sessionToken' | 'proxyPassword'
>;
