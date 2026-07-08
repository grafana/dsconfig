/**
 * Configuration models for the Oracle Database datasource plugin (`grafana-oracle-datasource`).
 *
 * Sources of truth (github.com/grafana/plugins-private @ 267f4937806ed6404b6628d13ae358a5d308e376,
 * plugins/grafana-oracle-datasource):
 * - `src/types.ts:6-23` — `OracleOptions` (jsonData) and `OracleSecureOptions` (secureJsonData)
 * - `src/components/ConfigEditor.tsx` — the configuration editor
 * - `pkg/models/settings.go:16-88` — backend `DBConnectionOptions` + `ConnectionOptions` loader
 *   (reads root `url` and secureJsonData `password`; everything else from jsonData)
 * - External:
 *   - `@grafana/ui` ^11.6.7 (`catalog:`): `LegacyForms.SecretFormField` (password field — its
 *     default `label`/`placeholder` are both "Password" since none are passed), `Select`, `Input`,
 *     `Switch`, `InlineField`
 *   - `@grafana/plugin-ui` ^0.13.1 (`catalog:`): `ConfigSection`, `ConfigSubSection`,
 *     `DataSourceDescription`, `Auth` (renders the "Oracle authentication" custom method)
 *   - `@grafana/data` ^11.6.7 (`catalog:`): `DataSourceJsonData` (base of `OracleOptions`),
 *     `onUpdateDatasourceOption` (writes root `url`), `onUpdateDatasourceJsonDataOption`,
 *     `onUpdateDatasourceSecureJsonDataOption`
 *   - `countries-and-timezones` 2.3.1: supplies the dynamic Time zone select options
 */

/**
 * Connection method the config editor selects (`src/components/ConfigEditor.tsx:531-534`,
 * `164-183`). Editor-local `Select` value derived from `jsonData.useTNSNamesBasedConnection`;
 * not stored under this name.
 */
export type OracleConnectionType = 'tcp' | 'tns';

/**
 * Authentication method the config editor selects (`src/components/ConfigEditor.tsx:535-538`,
 * `185-204`). Editor-local `Select` value derived from `jsonData.useKerberosAuthentication`;
 * not stored under this name.
 */
export type OracleAuthType = 'basic' | 'kerberos';

/**
 * Root (top-level datasource settings) fields the Oracle backend actually reads.
 *
 * Only `url` is consumed by the backend (`pkg/models/settings.go:79`, and as the legacy
 * TNSNames fallback at `:82-84`). The config editor writes it via
 * `onUpdateDatasourceOption(props, 'url')` (`src/components/ConfigEditor.tsx:240`). The plugin
 * reads no other root fields (`user`/`database`/`password` all live in jsonData/secureJsonData).
 */
export type RootConfig = {
  /**
   * "Host" — hostname or IP address with TCP port number for the "Host with TCP Port"
   * connection method (e.g. `oracle.example.com:1521`). For legacy TNSNames datasources
   * (v3.3.0) this held the tnsnames.ora entry before it moved to `jsonData.tnsNamesEntry`
   * in v3.3.2; the backend still falls back to it (`pkg/models/settings.go:82-84`).
   */
  url?: string;
};

/**
 * Fields stored in `jsonData`. Mirrors the plugin's `OracleOptions` (`src/types.ts:6-19`)
 * minus `enableSecureSocksProxy` (Secure Socks Proxy, intentionally excluded from this entry),
 * and matches the backend `DBConnectionOptions` json tags (`pkg/models/settings.go:22-32`).
 */
export type JsonDataConfig = {
  /**
   * "Time zone" select (`ConfigEditor.tsx:386-402`). Default `"UTC"` (editor `useEffect`
   * `ConfigEditor.tsx:108`; backend `pkg/models/settings.go:51-53`). The backend always
   * connects in UTC and shifts later, so this only affects display.
   */
  timezone_name?: string;
  /**
   * Backend-only: parsed into `DBConnectionOptions.DSTEnabled` (`pkg/models/settings.go:23`)
   * but never used when building the connection string. Declared here in `OracleOptions`
   * (`src/types.ts:9`) but not written by the config editor; `use_dst` is a per-query /
   * annotation option (`src/types.ts:30`, `src/datasource.ts:109`).
   */
  use_dst?: boolean;
  /** "Database" — name of database (`ConfigEditor.tsx:248-259`). Used for the "Host with TCP Port" method. */
  database?: string;
  /** "User" — an Oracle username with access to the specified database (`ConfigEditor.tsx:327-338`, basic auth). */
  user?: string;
  /** Connection-method discriminator, driven by the "Connection methods" select (`ConfigEditor.tsx:140-183`). false = Host with TCP Port, true = TNSNames Entry. */
  useTNSNamesBasedConnection?: boolean;
  /** "TNSName" — a tnsnames.ora entry (`ConfigEditor.tsx:270-284`). Used when `useTNSNamesBasedConnection` is true. */
  tnsNamesEntry?: string;
  /** Authentication discriminator, driven by the auth-type select (`ConfigEditor.tsx:153-204`). false = Basic, true = Kerberos. Editor only exposes Kerberos when `useTNSNamesBasedConnection` is true. */
  useKerberosAuthentication?: boolean;
  /** "Connection Pool size" (`ConfigEditor.tsx:450-463`). Backend default 50 when 0 (`pkg/models/settings.go:63-65`). */
  connectionPoolSize?: number;
  /** "Dataproxy Timeout" in seconds (`ConfigEditor.tsx:466-479`). Backend default 120 when 0 (`pkg/models/settings.go:68-70`). */
  dataProxyTimeout?: number;
  /** "Prefetch Row Size" (`ConfigEditor.tsx:482-495`). Only appended to the driver connection string when > 0 (`pkg/models/settings.go:130-156`). */
  prefetchRowsCount?: number;
  /** "Row Limit" (`ConfigEditor.tsx:498-511`). Backend default 1000000 when <= 0 (`pkg/models/settings.go:73-75`). */
  rowLimit?: number;
};

/**
 * Secret key names stored in `secureJsonData` (write-only; read existing config via
 * `secureJsonFields`):
 * - `password` — the Oracle user password for basic authentication. Written by
 *   `onUpdateDatasourceSecureJsonDataOption(props, 'password')` (`ConfigEditor.tsx:347`),
 *   consumed at `pkg/models/settings.go:77` (URL-query-escaped) and `pkg/oracle/utils.go:10`.
 *   Not used for Kerberos authentication.
 */
export type SecureJsonDataConfig = Array<'password'>;
