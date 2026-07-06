/**
 * Configuration models for the Amazon Aurora datasource plugin
 * (`grafana-aurora-datasource`).
 *
 * Sources of truth (all read at pinned versions):
 * - Plugin (https://github.com/grafana/grafana-aurora-datasource@c7452e8):
 *   - `src/plugin.json` — plugin id (`grafana-aurora-datasource`), name (`Amazon Aurora`), docs link
 *   - `src/components/ConfigEditor.tsx` — Database Settings + Advanced auth-endpoint subsections
 *     plus the embedded `@grafana/aws-sdk` ConnectionConfig. Aurora does NOT pass
 *     `showHttpProxySettings`, so the AWS proxy fields are not editor-visible.
 *   - `src/types.ts` — `AuroraConfigOptions extends AwsAuthDataSourceJsonData` and
 *     `AuroraSecureConfigOptions extends AwsAuthDataSourceSecureJsonData`, plus
 *     `SupportedEngines` enum (`aurora-mysql`, `aurora-postgres`).
 * - `@grafana/aws-sdk` `0.10.2` (github.com/grafana/grafana-aws-sdk-react v0.10.2):
 *   - `src/types.ts` — `AwsAuthType`, `AwsAuthDataSourceJsonData`,
 *     `AwsAuthDataSourceSecureJsonData`
 *   - `src/components/ConnectionConfig.tsx` — every AWS field's label,
 *     placeholder, description, and conditional render
 *   - `src/providers.ts` — the Select options for `authType`
 * - Backend `grafana-aws-sdk` `v1.4.6` (`pkg/awsds/settings.go`):
 *   - `AWSDatasourceSettings` — the Go struct the backend embeds. Its custom
 *     `AuthType.UnmarshalJSON` maps legacy `sharedCreds` and `arn` to the
 *     modern values, which is why the schema's allowedValues includes `arn`.
 * - Backend plugin (`pkg/plugin/driver.go:90-116`):
 *   - `AuroraConfigSettings` — embeds `awsds.AWSDatasourceSettings` and adds
 *     `Engine`, `DBUser`, `DBName`, `DBHost`, `DBPort`, `DefaultRegion`,
 *     `DBHostAuth`, `DBPortAuth`. `LoadSettings` reads accessKey / secretKey /
 *     sessionToken directly from `settings.DecryptedSecureJSONData`.
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
 * The Aurora engine values persisted to `jsonData.engine`, matching
 * `SupportedEngines` in `src/types.ts:44-47`.
 *
 * - `aurora-postgres` — Aurora (PostgreSQL Compatible); default port 5432
 * - `aurora-mysql` — Aurora (MySQL Compatible); default port 3306
 */
export type AuroraEngine = 'aurora-postgres' | 'aurora-mysql';

/**
 * Root (top-level datasource settings) fields.
 *
 * The Aurora datasource stores no plugin-specific fields at the root level
 * (`url`, `basicAuth`, etc. are unused by the plugin's backend; the RDS
 * endpoint lives in `jsonData.dbHost` / `jsonData.dbPort`). This is a blank
 * object rather than null.
 */
export type RootConfig = Record<string, never>;

/**
 * Fields stored in `jsonData`. Union of the AWS-shared fields from
 * `AwsAuthDataSourceJsonData` (`types.ts:15-25` of `@grafana/aws-sdk@0.10.2`)
 * and the Aurora-specific fields from `AuroraConfigOptions`
 * (`src/types.ts:32-40`).
 *
 * The Aurora config editor does not render the AWS proxy fields
 * (`proxyType`, `proxyUrl`, `proxyUsername`) because `ConfigEditor.tsx:31-34`
 * calls `<ConnectionConfig>` without `showHttpProxySettings`, so those
 * fields are omitted from this shape.
 */
export type JsonDataConfig = {
  // ---- AWS SDK ConnectionConfig fields (`@grafana/aws-sdk@0.10.2`) ----

  /** AWS credentials chain to use. Discriminator for accessKey/secretKey/profile/assumeRoleArn. */
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
  /** Default AWS region, e.g. `us-east-1`. Also passed to the RDS `generate-db-auth-token` call (pkg/plugin/connect.go:68). */
  defaultRegion?: string;

  // ---- Aurora-specific (`AuroraConfigOptions`) ----

  /**
   * Aurora engine. Default `aurora-postgres` (`ConfigEditor.tsx:47`). If empty
   * or unrecognized, the backend falls back to `aurora-postgres` at connect
   * time (`pkg/plugin/connect.go:83-85, 135-138`) to keep legacy beta
   * customers working.
   */
  engine?: AuroraEngine;
  /** Aurora database name — used to build the SQL driver DSN. */
  dbName?: string;
  /**
   * Aurora database user — the DB principal the RDS IAM auth token will
   * impersonate. Required by the config editor (`ConfigEditor.tsx:62-75`)
   * and by the backend to construct the DSN.
   */
  dbUser?: string;
  /**
   * Aurora cluster endpoint (host portion only). Required by the config
   * editor (`ConfigEditor.tsx:76-100`). The backend uses this for both
   * query traffic and the RDS `generate-db-auth-token` call unless
   * `dbHostAuth` overrides the latter.
   */
  dbHost?: string;
  /**
   * Aurora cluster port. Required by the config editor
   * (`ConfigEditor.tsx:101-116`). Editor shows `3306` as the placeholder
   * when `engine === 'aurora-mysql'`, otherwise `5432`. Stored as a JSON
   * number (the input type is `number` and `onChange` casts to `Number`).
   * Nullable in the frontend type because a cleared input writes `null`.
   */
  dbPort?: number | null;
  /**
   * Optional. Separate host for generating the RDS IAM auth token — useful
   * when Grafana connects through a load balancer that hides the primary
   * cluster endpoint. Backend falls back to `dbHost` when empty
   * (`pkg/plugin/connect.go:59-62`).
   */
  dbHostAuth?: string;
  /**
   * Optional. Separate port for generating the RDS IAM auth token. Backend
   * falls back to `dbPort` when zero (`pkg/plugin/connect.go:63-66`).
   * Nullable in the frontend type because a cleared input writes `null`.
   */
  dbPortAuth?: number | null;
};

/**
 * Secret key names stored in `secureJsonData` (write-only; read existing
 * config via `secureJsonFields`).
 *
 * - `accessKey` / `secretKey` — set when `authType === 'keys'`
 * - `sessionToken` — backend-only (`pkg/plugin/driver.go:112`); no editor UI
 *   writes it, provisioning-only
 *
 * There is no password field: Aurora authenticates with an RDS IAM auth
 * token generated at connect time from the resolved AWS credentials.
 */
export type SecureJsonDataConfig = Array<'accessKey' | 'secretKey' | 'sessionToken'>;
