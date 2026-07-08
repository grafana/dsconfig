/**
 * Configuration models for the Yugabyte datasource plugin (`grafana-yugabyte-datasource`).
 *
 * Sources of truth (https://github.com/grafana/yugabyte-datasource @ d2c8a440):
 * - `src/plugin.json` — plugin `id`, `name`
 * - `src/types.ts` — `YugabyteOptions extends SQLOptions`, `YugabyteSecureJsonData`
 * - `src/components/ConfigEditor.tsx` — the configuration editor (4 fields)
 * - `pkg/settings.go` — backend `Settings` + `LoadSettings` + `BuildConnectionString`
 * - `pkg/driver.go` — how the settings are consumed (pgx + Secure Socks Proxy)
 * - External:
 *   - `@grafana/plugin-ui` 0.13.1: `SQLOptions`, `DataSourceDescription`,
 *     `ConfigSection`, `SecureSocksProxyToggle`
 *   - `@grafana/ui` 13.1.0: `Field`, `Input`, `SecretInput`
 */

/**
 * Root (top-level datasource settings) fields the Yugabyte backend actually reads.
 * Both are consumed from `backend.DataSourceInstanceSettings` in `pkg/settings.go:25,38`.
 */
export type RootConfig = {
  /** Yugabyte YSQL host+port, e.g. `"localhost:5433"`. Split via `net.SplitHostPort` at `pkg/settings.go:25`. */
  url?: string;
  /** Database user. Read at `pkg/settings.go:38`. */
  user?: string;
};

/**
 * Fields stored in `jsonData`.
 *
 * `YugabyteOptions extends SQLOptions` at `src/types.ts:11`, so any SQLOptions field
 * MAY appear in jsonData for legacy or provisioning payloads — but the backend
 * ignores everything except `database`. Only `database` and the (excluded)
 * `enableSecureSocksProxy` are recognized by this plugin.
 */
export type JsonDataConfig = {
  /**
   * Yugabyte database name. Populated on the backend via `json.Unmarshal` into
   * `Settings.Database` (`pkg/settings.go:15,43`) and interpolated into the
   * connection string at `pkg/settings.go:57`.
   */
  database?: string;
  /**
   * Secure Socks Proxy toggle written by `SecureSocksProxyToggle`
   * (`@grafana/plugin-ui/SecureSocksProxyToggle.tsx:19`) and consumed at
   * `pkg/driver.go:41-53`. Excluded from the dsconfig registry entry per
   * repository policy — declared here for type completeness only.
   */
  enableSecureSocksProxy?: boolean;
};

/**
 * Secret key names stored in `secureJsonData` (write-only).
 * Only `password` is used — read at `pkg/settings.go:39` and interpolated into
 * the connection string at `pkg/settings.go:56`.
 */
export type SecureJsonDataConfig = Array<'password'>;
