/**
 * Configuration models for the AWS IoT SiteWise datasource plugin
 * (`grafana-iot-sitewise-datasource`).
 *
 * Sources of truth (all read at pinned versions):
 * - Plugin (https://github.com/grafana/iot-sitewise-datasource@5fed0c9):
 *   - `src/plugin.json` — plugin id, name, docs link
 *   - `src/components/ConfigEditor.tsx` — the branching editor: AWS ConnectionConfig
 *     when `defaultRegion !== 'Edge'`; Edge Kernel custom controls when it is.
 *   - `src/types.ts:264-274` — `SitewiseOptions extends AwsAuthDataSourceJsonData`
 *     (adds `edgeAuthMode`, `edgeAuthUser`) and
 *     `SitewiseSecureJsonData extends AwsAuthDataSourceSecureJsonData`
 *     (adds `edgeAuthPass`, `cert`).
 *   - `src/regions.ts:5-20` — the sitewise-specific supported regions list,
 *     which includes the sentinel value `'Edge'`.
 * - `@grafana/aws-sdk` `0.10.2` (github.com/grafana/grafana-aws-sdk-react `v0.10.2`):
 *   - `src/types.ts` — `AwsAuthType`, `AwsAuthDataSourceJsonData`,
 *     `AwsAuthDataSourceSecureJsonData`
 *   - `src/components/ConnectionConfig.tsx` — every AWS field's label,
 *     placeholder, description, and conditional render
 *   - `src/providers.ts` — the Select options for `authType`
 * - Backend `grafana-aws-sdk` `v1.4.3` (`pkg/awsds/settings.go`):
 *   - `AWSDatasourceSettings` — the Go struct the backend loads jsonData into.
 *     Its custom `AuthType.UnmarshalJSON` maps legacy `sharedCreds` and `arn`
 *     to the modern values, which is why the schema allows both.
 * - Backend plugin (`pkg/models/setting.go:16-91`):
 *   - `AWSSiteWiseDataSourceSetting` — embeds `awsds.AWSDatasourceSettings` and
 *     adds `Cert`/`EdgeAuthMode`/`EdgeAuthUser`/`EdgeAuthPass`. Secrets are
 *     copied from `DecryptedSecureJSONData` at load time (`Load`, lines 24-50).
 *   - `Validate` (lines 52-74): `Edge` region requires `endpoint` and `cert`;
 *     non-`default` edgeAuthMode also requires `edgeAuthUser`/`edgeAuthPass`.
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
 * Edge Kernel authentication modes selected by the "Authentication Mode"
 * Select when `defaultRegion === 'Edge'`. Backend constants live at
 * `pkg/models/setting.go:12-14`.
 */
export type EdgeAuthMode = 'default' | 'linux' | 'ldap';

/**
 * Root (top-level datasource settings) fields.
 *
 * The IoT SiteWise datasource stores no plugin-specific fields at the root
 * level (`url`, `basicAuth`, etc. are unused), so this is a blank object
 * rather than null.
 */
export type RootConfig = Record<string, never>;

/**
 * Fields stored in `jsonData`. Union of the AWS-shared fields from
 * `AwsAuthDataSourceJsonData` (types.ts of `@grafana/aws-sdk@0.10.2`) and the
 * SiteWise-specific fields from `SitewiseOptions` (`src/types.ts:264-268`).
 *
 * The IoT SiteWise config editor does not render the AWS proxy fields
 * (`proxyType`, `proxyUrl`, `proxyUsername`) because
 * `ConfigEditor.tsx:36,110` calls `<ConnectionConfig>` without
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
  /**
   * Optional custom AWS service endpoint. Editor-visible when
   * `authType !== 'grafana_assume_role'`, and also rendered as an explicit
   * control in Edge Kernel mode (`ConfigEditor.tsx:86-97`). Required by the
   * backend when `defaultRegion === 'Edge'` (`setting.go:57-59`).
   */
  endpoint?: string;
  /**
   * Default AWS region, e.g. `us-east-1`. The sitewise supportedRegions list
   * (`src/regions.ts`) includes the sentinel `'Edge'` which flips the editor
   * into Edge Kernel mode and requires an endpoint + SSL cert to connect.
   */
  defaultRegion?: string;

  // ---- IoT SiteWise-specific (`SitewiseOptions`, `src/types.ts:264-268`) ----

  /**
   * Edge Kernel auth mode Select. Editor-visible when
   * `defaultRegion === 'Edge'`. Backend defaults empty to `'default'` when
   * `Region === 'Edge'` (`setting.go:40-42`).
   */
  edgeAuthMode?: EdgeAuthMode;
  /**
   * Username for local Edge Kernel authentication proxy. Editor-visible when
   * `defaultRegion === 'Edge' && edgeAuthMode !== 'default'`. Required by
   * backend under the same condition (`setting.go:64-67`).
   */
  edgeAuthUser?: string;
};

/**
 * Secret key names stored in `secureJsonData` (write-only; read existing
 * config via `secureJsonFields`).
 *
 * - `accessKey` / `secretKey` — set for `authType === 'keys'`
 * - `sessionToken` — backend-only (`awsds.settings.go:137`); not rendered by
 *   ConnectionConfig, provisioning-only
 * - `edgeAuthPass` — Edge Kernel proxy password, paired with `edgeAuthUser`
 *   (editor-visible when `edgeAuthMode !== 'default'`)
 * - `cert` — Edge Kernel SSL certificate (PEM); editor-visible whenever
 *   `defaultRegion === 'Edge'`
 */
export type SecureJsonDataConfig = Array<
  'accessKey' | 'secretKey' | 'sessionToken' | 'edgeAuthPass' | 'cert'
>;
