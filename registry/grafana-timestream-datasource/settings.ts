/**
 * Configuration models for the Amazon Timestream datasource plugin
 * (`grafana-timestream-datasource`).
 *
 * Sources of truth (all read at pinned versions):
 * - Plugin (https://github.com/grafana/timestream-datasource@9e34c64):
 *   - `src/plugin.json` — plugin id (`grafana-timestream-datasource`), name
 *     ("Amazon Timestream"), docs link
 *   - `src/components/ConfigEditor.tsx` — the entire config editor:
 *     `<ConnectionConfig standardRegions={standardRegions} defaultEndpoint=
 *     "https://query-{cell}.timestream.{region}.amazonaws.com">` plus a
 *     `<ConfigSection title="Timestream Details" description="Default values
 *     to be used as macros">` with three `<ConfigSelect>` fields for
 *     defaultDatabase, defaultTable, and defaultMeasure. The
 *     `<SecureSocksProxySettings>` render (`jsonData.enableSecureSocksProxy`)
 *     is excluded per AGENTS.md.
 *   - `src/components/selectors.ts` — the Field labels ("Database", "Table",
 *     "Measure") used for the Timestream Details section.
 *   - `src/regions.ts` — the 9-region standardRegions list passed to
 *     ConnectionConfig (a Timestream-specific subset — Timestream is not
 *     available in every AWS region).
 *   - `src/types.ts:94-102` — `TimestreamOptions extends AwsAuthDataSourceJsonData`
 *     with `defaultDatabase`, `defaultTable`, `defaultMeasure`;
 *     `TimestreamSecureJsonData extends AwsAuthDataSourceSecureJsonData` adds
 *     nothing (comment: "nothing for now").
 *   - `pkg/models/settings.go:12-45` — backend `DatasourceSettings` embeds
 *     `awsds.AWSDatasourceSettings` and adds `DefaultDatabase`,
 *     `DefaultTable`, `DefaultMeasure`. `Load` unmarshals jsonData, then adds
 *     two legacy fallbacks: `settings.Region → settings.DefaultRegion` when
 *     Region is empty/"default", and `settings.Profile → config.Database` when
 *     Profile is empty (comment: "legacy support (only for cloudwatch?)").
 * - `@grafana/aws-sdk` `0.10.2` (grafana/grafana-aws-sdk-react tag `v0.10.2`,
 *   SHA `fe0c4d8`):
 *   - `src/types.ts` — `AwsAuthType`, `AwsAuthDataSourceJsonData`,
 *     `AwsAuthDataSourceSecureJsonData`, `ConnectionConfigProps`
 *   - `src/components/ConnectionConfig.tsx` — every AWS field's label,
 *     placeholder, description, and conditional render. Timestream does NOT
 *     pass `showHttpProxySettings`, `hideAssumeRoleArn`, or `skipEndpoint`,
 *     so the Assume Role subsection and the Endpoint field are visible for
 *     every non-`grafana_assume_role` provider, and the proxy subsection is
 *     omitted from the editor entirely. Line 18-28 lists Timestream in
 *     `DS_TYPES_THAT_SUPPORT_TEMP_CREDS`, so Grafana Assume Role appears when
 *     the `awsDatasourcesTempCredentials` feature toggle is on.
 *   - `src/providers.ts` — the Select options for `authType`.
 *   - `src/sql/ConfigEditor/ConfigSelect.tsx` — the Select widget used for
 *     the Timestream Details fields (disables itself until `defaultRegion`
 *     is set).
 * - Backend `grafana-aws-sdk` `v1.4.4` (`pkg/awsds/settings.go`):
 *   - `AWSDatasourceSettings` — Go struct that receives jsonData; note the
 *     `assumeRoleARN` (uppercase RN) json tag versus the frontend's
 *     `assumeRoleArn` (lowercase arn). Go's case-insensitive Unmarshal makes
 *     both work.
 *   - `AuthType.MarshalJSON`/`UnmarshalJSON` — the storage⇆enum mapping that
 *     folds legacy `arn`→`default` and `sharedCreds`→`credentials`.
 */

/**
 * The AWS authentication provider values persisted to `jsonData.authType`.
 *
 * Five values are editor-selectable via the ConnectionConfig `Select`
 * (`providers.ts:4-25`): `ec2_iam_role`, `grafana_assume_role`, `default`,
 * `keys`, `credentials`. A sixth value, `arn`, is a legacy stored value the
 * backend maps to `default` (see `awsds.AuthType.UnmarshalJSON` /
 * `awsds/settings.go`).
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
 * Timestream does not surface a UI for any root-level datasource setting, but
 * its backend `Load` (`pkg/models/settings.go:36-38`) reads the top-level
 * `database` field as a legacy fallback for `jsonData.profile` when the
 * latter is empty (with the upstream comment: "legacy support (only for
 * cloudwatch?)"). Modelled here so provisioning tools can round-trip
 * `database` verbatim.
 */
export type RootConfig = {
  /**
   * Legacy fallback for `jsonData.profile`. When `jsonData.profile` is empty
   * and this is non-empty, the plugin uses it as the AWS credentials profile
   * name (`pkg/models/settings.go:36-38`). No editor UI writes it;
   * datasources that never carried a Grafana `database` value can omit it.
   */
  database?: string;
};

/**
 * Fields stored in `jsonData`. Union of the AWS-shared fields from
 * `AwsAuthDataSourceJsonData` (`@grafana/aws-sdk@0.10.2`, `src/types.ts:15-25`)
 * that Timestream's ConfigEditor exposes, plus the Timestream-specific
 * defaults `defaultDatabase`, `defaultTable`, and `defaultMeasure` from
 * `TimestreamOptions` (`src/types.ts:94-98`).
 *
 * Timestream does NOT pass `showHttpProxySettings` to `ConnectionConfig`, so
 * the AWS proxy fields (`proxyType`, `proxyUrl`, `proxyUsername`) are neither
 * editor-visible nor part of the plugin's declared surface. `region` is
 * declared on `AwsAuthDataSourceJsonData` but the frontend never writes it;
 * the backend mirrors `defaultRegion` into `Region` at load time
 * (`pkg/models/settings.go:32-34`).
 */
export type JsonDataConfig = {
  // ---- AWS SDK ConnectionConfig fields (`@grafana/aws-sdk@0.10.2`) ----

  /** AWS credentials chain to use. Discriminator for accessKey/secretKey/profile/assumeRoleArn. */
  authType?: AwsAuthType;
  /** Credentials profile name from `~/.aws/credentials`. Editor-visible when `authType === 'credentials'`. */
  profile?: string;
  /**
   * ARN of an IAM role to assume via STS. Editor-visible when the caller
   * doesn't pass `hideAssumeRoleArn` (Timestream does not). Note: the
   * backend's json tag is `assumeRoleARN` (uppercase RN); Go's
   * case-insensitive Unmarshal makes the frontend spelling work too.
   */
  assumeRoleArn?: string;
  /** External ID passed to STS AssumeRole. Editor-visible when `authType !== 'grafana_assume_role'`. */
  externalId?: string;
  /**
   * Optional custom AWS service endpoint. Editor-visible when
   * `authType !== 'grafana_assume_role'`. Placeholder is
   * `https://query-{cell}.timestream.{region}.amazonaws.com` — Timestream
   * exposes cell-scoped endpoints, so the `{cell}` token replaces the
   * generic `{service}` used by other AWS plugins.
   */
  endpoint?: string;
  /** Default AWS region, e.g. `us-east-1`. Also used to seed the runtime `region` (backend). */
  defaultRegion?: string;

  // ---- Timestream-specific (`TimestreamOptions`) ----

  /** Default Timestream database. Feeds the {{database}} query macro. */
  defaultDatabase?: string;
  /** Default Timestream table. Feeds the {{table}} query macro. Depends on defaultDatabase. */
  defaultTable?: string;
  /** Default Timestream measure. Feeds the {{measure}} query macro. Depends on defaultDatabase + defaultTable. */
  defaultMeasure?: string;
};

/**
 * Secret key names stored in `secureJsonData` (write-only; read existing
 * config via `secureJsonFields`).
 *
 * - `accessKey` / `secretKey` — set when `authType === 'keys'`
 * - `sessionToken` — backend-only (`pkg/models/settings.go:42`); no editor
 *   UI writes it. Used for temporary STS credentials (paired with `keys`).
 */
export type SecureJsonDataConfig = Array<'accessKey' | 'secretKey' | 'sessionToken'>;
