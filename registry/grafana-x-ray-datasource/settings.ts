/**
 * Configuration models for the AWS Application Signals (X-Ray) datasource
 * plugin (`grafana-x-ray-datasource`).
 *
 * Sources of truth (all read at pinned versions):
 * - Plugin (https://github.com/grafana/x-ray-datasource@3d8a237):
 *   - `src/plugin.json` — plugin id (`grafana-x-ray-datasource`), name
 *     ("AWS Application Signals" — renamed from "X-Ray" in v2.16.0), docs
 *     link
 *   - `src/components/ConfigEditor/ConfigEditor.tsx` — the entire config
 *     editor: `<ConnectionConfig standardRegions=...>` plus the
 *     `<SecureSocksProxySettings>` (which writes
 *     `jsonData.enableSecureSocksProxy` and is excluded per AGENTS.md).
 *   - `src/components/ConfigEditor/regions.ts` — the standardRegions list
 *     passed to ConnectionConfig.
 *   - `src/types.ts:106-108` — `XrayJsonData extends AwsAuthDataSourceJsonData`
 *     with a "Can add X-Ray specific values here" placeholder — no
 *     plugin-specific fields today.
 *   - `pkg/datasource/configuration.go:8-20` — backend `getDsSettings`: uses
 *     `awsds.AWSDatasourceSettings.Load` verbatim and adds a legacy
 *     `settings.Database → Profile` fallback when Profile is empty.
 * - `@grafana/aws-sdk` `0.10.2` (grafana/grafana-aws-sdk-react tag `v0.10.2`,
 *   SHA `fe0c4d8`):
 *   - `src/types.ts` — `AwsAuthType`, `AwsAuthDataSourceJsonData`,
 *     `AwsAuthDataSourceSecureJsonData`, `ConnectionConfigProps`
 *   - `src/components/ConnectionConfig.tsx` — every AWS field's label,
 *     placeholder, description, and conditional render. Note X-Ray does NOT
 *     pass `showHttpProxySettings`, `hideAssumeRoleArn`, or `skipEndpoint`,
 *     so the Assume Role subsection and the Endpoint field are visible for
 *     every non-`grafana_assume_role` provider, and the proxy subsection is
 *     omitted from the editor entirely.
 *   - `src/providers.ts` — the Select options for `authType` (labels:
 *     Workspace IAM Role, Grafana Assume Role, AWS SDK Default, Access &
 *     secret key, Credentials file)
 *   - `src/regions.ts` — same standardRegions list that X-Ray re-exports
 *     verbatim in `src/components/ConfigEditor/regions.ts`
 * - Backend `grafana-aws-sdk` `v1.4.3` (`pkg/awsds/settings.go`):
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
 * Root (top-level datasource settings) fields.
 *
 * X-Ray does not surface a UI for any root-level datasource setting, but its
 * backend `getDsSettings` (`pkg/datasource/configuration.go:16-18`) reads the
 * top-level `database` field as a legacy fallback for `jsonData.profile` when
 * the latter is empty. Modelled here so provisioning tools can round-trip
 * `database` verbatim.
 */
export type RootConfig = {
  /**
   * Legacy fallback for `jsonData.profile`. When `jsonData.profile` is empty
   * and this is non-empty, the plugin uses it as the AWS credentials profile
   * name (`pkg/datasource/configuration.go:16-18`). No editor UI writes it;
   * datasources that never carried a Grafana `database` value can omit it.
   */
  database?: string;
};

/**
 * Fields stored in `jsonData`. Union of the AWS-shared fields from
 * `AwsAuthDataSourceJsonData` (`@grafana/aws-sdk@0.10.2`, `src/types.ts:15-25`)
 * that X-Ray's ConfigEditor actually exposes.
 *
 * X-Ray does NOT pass `showHttpProxySettings` to `ConnectionConfig`, so the
 * AWS proxy fields (`proxyType`, `proxyUrl`, `proxyUsername`) are neither
 * editor-visible nor part of the plugin's declared surface. `region` is
 * declared on `AwsAuthDataSourceJsonData` but the frontend never writes it;
 * the backend mirrors `defaultRegion` into `Region` at load time
 * (`awsds/settings.go:127-129`).
 *
 * X-Ray defines a plugin-specific `XrayJsonData` type but leaves it as a
 * marker comment — it has no plugin-specific jsonData fields today.
 */
export type JsonDataConfig = {
  /** AWS credentials chain to use. Discriminator for accessKey/secretKey/profile/assumeRoleArn. */
  authType?: AwsAuthType;
  /** Credentials profile name from `~/.aws/credentials`. Editor-visible when `authType === 'credentials'`. */
  profile?: string;
  /**
   * ARN of an IAM role to assume via STS. Editor-visible when the caller
   * doesn't pass `hideAssumeRoleArn` (X-Ray does not). Note: the backend's
   * json tag is `assumeRoleARN` (uppercase RN); Go's case-insensitive
   * Unmarshal makes the frontend spelling work too.
   */
  assumeRoleArn?: string;
  /** External ID passed to STS AssumeRole. Editor-visible when `authType !== 'grafana_assume_role'`. */
  externalId?: string;
  /** Optional custom AWS service endpoint. Editor-visible when `authType !== 'grafana_assume_role'`. */
  endpoint?: string;
  /** Default AWS region, e.g. `us-east-1`. Also used to seed the runtime `region` (backend). */
  defaultRegion?: string;
};

/**
 * Secret key names stored in `secureJsonData` (write-only; read existing
 * config via `secureJsonFields`).
 *
 * - `accessKey` / `secretKey` — set when `authType === 'keys'`
 * - `sessionToken` — backend-only (`awsds/settings.go:137`); no editor UI
 *   writes it. Used for temporary STS credentials (paired with `keys`) and,
 *   since plugin v2.17.0, to support the Grafana Assume Role flow.
 */
export type SecureJsonDataConfig = Array<'accessKey' | 'secretKey' | 'sessionToken'>;
