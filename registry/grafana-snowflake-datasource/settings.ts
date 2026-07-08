/**
 * Configuration models for the Snowflake datasource plugin (`grafana-snowflake-datasource`).
 *
 * Sources of truth (github.com/grafana/plugins-private @ 267f4937,
 * plugins/grafana-snowflake-datasource):
 * - `src/types.ts:5-52` — `SnowflakeDataSourceOptions` (jsonData), `SnowflakeSecureJsonData`
 *   (secret keys), `Setting` (session parameter), `AuthenticationType` enum
 * - `src/types.ts:54-70` — `InterpolationFormat` enum (frontend-only)
 * - `src/selectors.ts:4-151` — every editor label, placeholder, and tooltip
 * - `src/editors/ConfigEditor.tsx` — the configuration editor
 * - `src/components/SessionSettings.tsx` — the session-parameter (jsonData.settings) editor
 * - `pkg/settings.go:24-45` — backend `Settings` (the jsonData fields the backend reads)
 * - `pkg/driver.go:96-147` — how the settings feed the gosnowflake `sf.Config`
 */

/**
 * Authentication method, stored in `jsonData.authType`.
 * Mirrors `AuthenticationType` (`src/types.ts:47-52`) and the backend constants
 * (`pkg/constants.go:12-16`). An empty/missing value is treated as `password`.
 */
export type SnowflakeAuthType = 'password' | 'keypair' | 'pat' | 'oauth';

/**
 * Variable interpolation format, stored in `jsonData.defaultInterpolation`.
 * Frontend-only (the backend never reads it). Mirrors `InterpolationFormat`
 * (`src/types.ts:54-70`); `''` is the "None" default.
 */
export type SnowflakeInterpolationFormat =
  | ''
  | 'raw'
  | 'sqlstring'
  | 'regex'
  | 'csv'
  | 'distributed'
  | 'doublequote'
  | 'glob'
  | 'json'
  | 'lucene'
  | 'percentencode'
  | 'pipe'
  | 'singlequote'
  | 'text'
  | 'queryparam';

/**
 * A single session parameter, stored as an element of `jsonData.settings`
 * (`src/types.ts:41-45`, `src/components/SessionSettings.tsx`). When `secure` is
 * true, the value is written to `secureJsonData` under a key equal to `name`
 * rather than being stored inline.
 */
export type SnowflakeSessionSetting = {
  name: string;
  value?: string;
  secure: boolean;
};

/**
 * Root (top-level datasource settings) fields.
 *
 * The Snowflake backend authenticates entirely from `jsonData` + `secureJsonData`
 * (`pkg/settings.go:56-192`) and never reads root-level datasource fields such as
 * `url`, `user`, or `basicAuth`, so this is a blank object rather than null.
 */
export type RootConfig = Record<string, never>;

/**
 * Fields stored in `jsonData`. Matches the plugin's `SnowflakeDataSourceOptions`
 * (`src/types.ts:5-28`) minus `enableSecureSocksProxy` (Secure Socks Proxy is
 * excluded from registry entries).
 */
export type JsonDataConfig = {
  /** Snowflake account identifier (`<account>.snowflakecomputing.com`); required. */
  account?: string;
  /** Username assigned via `CREATE USER`. Read by the backend; hidden in the editor for `oauth`. */
  username?: string;
  /** Deprecated region field; prefer encoding the region in `account`. */
  region?: string;
  /** Assume a non-default role for queries; must still be granted to the user. */
  role?: string;
  /** Authentication discriminator; defaults to `password` in the editor and backend. */
  authType?: SnowflakeAuthType;
  /** Default warehouse for queries. */
  warehouse?: string;
  /** Default database for queries. */
  database?: string;
  /** Default schema for queries. */
  schema?: string;
  /** Frontend-only: seed SQL for a new panel query. Not read by the backend. */
  defaultQuery?: string;
  /** Frontend-only: seed SQL for a new variable query. Not read by the backend. */
  defaultVariableQuery?: string;
  /** Frontend-only: template variable interpolation format. Not read by the backend. */
  defaultInterpolation?: SnowflakeInterpolationFormat;
  /** Frontend-only: min interval for `$__interval` / `$__interval_ms`. Not read by the backend. */
  timeInterval?: string;
  /** Connection (login) timeout in seconds; backend default 5 when unset. */
  loginTimeout?: number;
  /** Request (query) timeout in seconds; backend default 120 when unset. */
  requestTimeout?: number;
  /** Session parameters appended to the connection (`pkg/driver.go:72-86`). */
  settings?: SnowflakeSessionSetting[];
  /** Forward the user's OAuth identity; required to be `true` for `oauth` auth. */
  oauthPassThru?: boolean;
  /** Plugin-applied cap on rows returned per query (`pkg/driver.go:67`). */
  rowLimit?: number;
};

/**
 * Secret key names stored in `secureJsonData` (write-only; read existing config
 * via `secureJsonFields`). Mirrors `SnowflakeSecureJsonData` (`src/types.ts:30-35`):
 * - `password` — password auth (`pkg/settings.go:79-85`)
 * - `privateKey` — key-pair auth PEM (`pkg/settings.go:94-125`)
 * - `privateKeyPassphrase` — passphrase for an encrypted key-pair key (`pkg/settings.go:91-92`)
 * - `patToken` — programmatic access token (`pkg/settings.go:139-145`)
 *
 * Session parameters marked secure add further dynamic keys (named after the
 * setting) which are user-defined and not enumerable here.
 */
export type SecureJsonDataConfig = Array<'password' | 'privateKey' | 'privateKeyPassphrase' | 'patToken'>;
