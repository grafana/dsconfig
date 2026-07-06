/**
 * Configuration models for the MySQL datasource plugin (`mysql`).
 *
 * Sources of truth (https://github.com/grafana/grafana-mysql-datasource @ 98f55a8e):
 * - `src/types.ts` — `MySQLOptions extends SQLOptions` (adds `allowCleartextPasswords`)
 * - `src/configuration/ConfigurationEditor.tsx` — the configuration editor
 * - `pkg/mysql/mysql.go` — backend Datasource setup (reads root URL/User/Database + secret password)
 * - `pkg/mysql/sqleng/sql_engine.go:40-61` — `sqleng.JsonData` (backend jsonData shape)
 * - External:
 *   - `@grafana/sql` 13.0.1: `SQLOptions` (`types.ts:37-46`), `ConnectionLimits`
 *     (`components/configuration/ConnectionLimits.tsx`), `TLSSecretsConfig`
 *     (`components/configuration/TLSSecretsConfig.tsx`), `useMigrateDatabaseFields`
 *     (migrates root `database` -> `jsonData.database` and sets connection-pool defaults)
 */

/**
 * Root (top-level datasource settings) fields the MySQL backend actually reads. Populated
 * by the editor via direct `options.url` / `options.user` writes, not through jsonData.
 */
export type RootConfig = {
  /** MySQL host+port (or a Unix socket path starting with '/'). Read at `pkg/mysql/mysql.go:65-77`. */
  url?: string;
  /** Database user. Read at `pkg/mysql/mysql.go:66,91,99`. */
  user?: string;
  /**
   * Legacy: older datasources stored the database name at root level. `useMigrateDatabaseFields`
   * (packages/grafana-sql) moves it into `jsonData.database` on first render; the backend
   * (`pkg/mysql/mysql.go:58-61`) falls back to root.database when `jsonData.database` is empty.
   */
  database?: string;
};

/**
 * Fields stored in `jsonData`. Combines the plugin's `MySQLOptions` (`src/types.ts:3-5`)
 * with the shared `SQLOptions` (`@grafana/sql/src/types.ts:37-46`) and the connection-pool
 * fields from `SQLConnectionLimits` (`@grafana/sql/src/types.ts:30-35`).
 */
export type JsonDataConfig = {
  /** Preferred way to store the database name. */
  database?: string;
  /**
   * TLS switches. `tlsAuth` enables mTLS with a client cert/key; `tlsAuthWithCACert` enables
   * server-cert verification against a custom CA. Both may be enabled independently
   * (`ConfigurationEditor.tsx:123-141`).
   */
  tlsAuth?: boolean;
  tlsAuthWithCACert?: boolean;
  tlsSkipVerify?: boolean;
  /**
   * MySQL cleartext client-side plugin toggle — required by some accounts.
   * Consumed at `pkg/mysql/mysql.go:106-108`.
   */
  allowCleartextPasswords?: boolean;
  /**
   * Timezone the backend sets on the session with `SET time_zone='...'`. Empty = inherit
   * database default. Consumed at `pkg/mysql/mysql.go:130-132`.
   */
  timezone?: string;
  /** Auto group-by lower bound, e.g. `"1m"`. */
  timeInterval?: string;
  /** Connection pool: consumed at `pkg/mysql/mysql.go:155-157`. */
  maxOpenConns?: number;
  maxIdleConns?: number;
  maxIdleConnsAuto?: boolean;
  connMaxLifetime?: number;
};

/**
 * Secret key names stored in `secureJsonData` (write-only; read existing config via
 * `secureJsonFields`):
 * - `password` — MySQL user password (consumed at `pkg/mysql/mysql.go:100`).
 * - `tlsClientCert`, `tlsClientKey` — supplied together when `tlsAuth` is enabled.
 * - `tlsCACert` — supplied when `tlsAuthWithCACert` is enabled.
 */
export type SecureJsonDataConfig = Array<'password' | 'tlsClientCert' | 'tlsClientKey' | 'tlsCACert'>;
