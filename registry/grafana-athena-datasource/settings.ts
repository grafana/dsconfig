/**
 * Configuration models for the Amazon Athena datasource plugin
 * (`grafana-athena-datasource`).
 *
 * Sources of truth (all read at pinned versions):
 * - Plugin (https://github.com/grafana/athena-datasource@a708c50):
 *   - `src/plugin.json` — plugin id, name, docs link
 *   - `src/ConfigEditor.tsx` — Athena Details subsection (catalog, database,
 *     workgroup, outputLocation) plus embedded `@grafana/aws-sdk` ConnectionConfig
 *   - `src/types.ts` — `AthenaDataSourceOptions extends AwsAuthDataSourceJsonData`
 *     and `AthenaDataSourceSecureJsonData extends AwsAuthDataSourceSecureJsonData`
 * - `@grafana/aws-sdk` `v0.10.2`
 *   (github.com/grafana/grafana-aws-sdk-react tag `v0.10.2`):
 *   - `src/types.ts` — `AwsAuthType`, `AwsAuthDataSourceJsonData`,
 *     `AwsAuthDataSourceSecureJsonData`
 *   - `src/components/ConnectionConfig.tsx` — every AWS field's label,
 *     placeholder, description, and conditional render
 *   - `src/providers.ts` — the Select options for `authType`
 * - Backend `grafana-aws-sdk` `v1.4.3` (`pkg/awsds/settings.go`):
 *   - `AWSDatasourceSettings` — the Go struct the backend loads jsonData into.
 *     Its custom `AuthType.UnmarshalJSON` maps legacy `sharedCreds` and `arn`
 *     to the modern values, which is why the schema allows both.
 * - Backend plugin (`pkg/athena/models/settings.go`):
 *   - `AthenaDataSourceSettings` — embeds `awsds.AWSDatasourceSettings` and
 *     adds Athena-specific fields with PascalCase json tags that match the
 *     frontend camelCase via Go's case-insensitive Unmarshal.
 */

/**
 * The AWS authentication provider values persisted to `jsonData.authType`.
 *
 * Five values are editor-selectable via the ConnectionConfig `Select`:
 * `ec2_iam_role`, `grafana_assume_role`, `default`, `keys`, `credentials`.
 * A sixth value, `arn`, is a legacy stored value the backend maps to
 * `default` (see `awsds.AuthType.UnmarshalJSON`).
 */
export type AwsAuthType =
  | 'default'
  | 'keys'
  | 'credentials'
  | 'ec2_iam_role'
  | 'grafana_assume_role'
  | 'arn';

/**
 * Root (top-level datasource settings) fields.
 *
 * The Athena datasource stores no plugin-specific fields at the root level
 * (`url`, `basicAuth`, etc. are unused), so this is a blank object rather
 * than null.
 */
export type RootConfig = Record<string, never>;

/**
 * Fields stored in `jsonData`. Union of the AWS-shared fields from
 * `AwsAuthDataSourceJsonData` (types.ts:15-25 of `@grafana/aws-sdk@0.10.2`)
 * and the Athena-specific fields from `AthenaDataSourceOptions`
 * (`src/types.ts:68-73`).
 *
 * The Athena config editor does not render the AWS proxy fields
 * (`proxyType`, `proxyUrl`, `proxyUsername`) or `region` because
 * `ConfigEditor.tsx:112` calls `<ConnectionConfig>` without
 * `showHttpProxySettings`, so those fields are omitted from this shape.
 */
export type JsonDataConfig = {
  // ---- AWS SDK ConnectionConfig fields (`@grafana/aws-sdk@0.10.2`) ----

  /** AWS credentials chain to use. Discriminator for `accessKey`/`secretKey`/`profile`/`assumeRoleArn`. */
  authType?: AwsAuthType;
  /** Credentials profile name from `~/.aws/credentials`. Editor-visible when `authType === 'credentials'`. */
  profile?: string;
  /**
   * ARN of an IAM role to assume via STS. Editor-visible for every auth type
   * except `grafana_assume_role`. Note: the backend's json tag is
   * `assumeRoleARN` (uppercase RN); Go's case-insensitive Unmarshal makes the
   * frontend spelling work too.
   */
  assumeRoleArn?: string;
  /** External ID passed to STS AssumeRole. Editor-visible when `authType !== 'grafana_assume_role'`. */
  externalId?: string;
  /** Optional custom AWS service endpoint. Editor-visible when `authType !== 'grafana_assume_role'`. */
  endpoint?: string;
  /** Default AWS region, e.g. `us-east-1`. Also used to seed the runtime `region` (backend). */
  defaultRegion?: string;

  // ---- Athena-specific (`AthenaDataSourceOptions`) ----

  /** Athena data catalog. Editor labels it "Data source" (selectors.ts:18). */
  catalog?: string;
  /** Athena database within the selected catalog. */
  database?: string;
  /** Athena workgroup. */
  workgroup?: string;
  /** S3 output location (e.g. `s3://bucket/prefix/`). Optional — falls back to the workgroup's default. */
  outputLocation?: string;
};

/**
 * Secret key names stored in `secureJsonData` (write-only; read existing
 * config via `secureJsonFields`).
 *
 * - `accessKey` / `secretKey` — set for `authType === 'keys'`
 * - `sessionToken` — backend-only (`pkg/athena/models/settings.go:47`,
 *   `awsds.settings.go:137`); not rendered by ConnectionConfig, provisioning-only
 */
export type SecureJsonDataConfig = Array<'accessKey' | 'secretKey' | 'sessionToken'>;
