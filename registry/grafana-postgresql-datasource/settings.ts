/**
 * Configuration models for the PostgreSQL datasource plugin (`grafana-postgresql-datasource`).
 *
 * Sources of truth (https://github.com/grafana/grafana-postgresql-datasource @ c5d28c45):
 * - `src/types.ts` — `PostgresOptions extends SQLOptions`, `PostgresTLSModes`, `PostgresTLSMethods`, `SecureJsonData`
 * - `src/configuration/ConfigurationEditor.tsx` — the configuration editor
 * - `pkg/postgresql/postgres.go` — backend Datasource setup (root URL/User/Database + jsonData)
 * - `pkg/postgresql/sqleng/sql_engine.go` — `sqleng.JsonData` (shared jsonData shape)
 * - External:
 *   - `@grafana/sql` 13.0.1: `SQLOptions`, `MaxOpenConnectionsField`, `MaxLifetimeField`,
 *     `TLSSecretsConfig`, `useMigrateDatabaseFields`
 */

export type PostgresTLSMode = 'disable' | 'require' | 'verify-ca' | 'verify-full';
export type PostgresTLSMethod = 'file-path' | 'file-content';

/**
 * Root (top-level datasource settings) fields the PostgreSQL backend actually reads.
 */
export type RootConfig = {
  /** PostgreSQL host+port. Read at `pkg/postgresql/postgres.go:108`. */
  url?: string;
  /** Database user. Read at `pkg/postgresql/postgres.go:109`. */
  user?: string;
  /**
   * Legacy: older datasources stored the database name at root level.
   * `useMigrateDatabaseFields` moves it into `jsonData.database`; the backend
   * (`pkg/postgresql/postgres.go:101-104`) falls back to root when jsonData.database is empty.
   */
  database?: string;
};

/**
 * Fields stored in `jsonData`. Extends the plugin's `PostgresOptions` (`src/types.ts:13-22`)
 * with the shared `SQLOptions` (`@grafana/sql`).
 */
export type JsonDataConfig = {
  /** Preferred way to store the database name. */
  database?: string;
  /** TLS negotiation mode. Default `'require'` (`ConfigurationEditor.tsx:197`). */
  sslmode?: PostgresTLSMode;
  /**
   * How TLS certificates are supplied. Default `'file-path'` (backend default at
   * `pkg/postgresql/postgres.go:92`). When `'file-content'`, the backend reads
   * `secureJsonData.tls*` inline and writes them to disk at connection time.
   */
  tlsConfigurationMethod?: PostgresTLSMethod;
  /** File-path TLS credentials (only used when tlsConfigurationMethod is 'file-path'). */
  sslRootCertFile?: string;
  sslCertFile?: string;
  sslKeyFile?: string;
  /**
   * Numeric PostgreSQL version, e.g. `903` = 9.3, `1500` = 15. Controls query-builder UI
   * only — the wire protocol is autodetected.
   */
  postgresVersion?: number;
  /** Enable TimescaleDB extension features in the query builder. */
  timescaledb?: boolean;
  /** Auto group-by lower bound, e.g. `"1m"`. */
  timeInterval?: string;
  /** Connection pool: consumed at `pkg/postgresql/postgres.go:351-354`. */
  maxOpenConns?: number;
  connMaxLifetime?: number;
};

/**
 * Secret key names stored in `secureJsonData` (write-only):
 * - `password` — PostgreSQL user password.
 * - `tlsCACert`, `tlsClientCert`, `tlsClientKey` — supplied when
 *   `tlsConfigurationMethod === 'file-content'`.
 */
export type SecureJsonDataConfig = Array<'password' | 'tlsCACert' | 'tlsClientCert' | 'tlsClientKey'>;
