/**
 * Configuration models for the CockroachDB datasource plugin (`grafana-cockroachdb-datasource`).
 *
 * Sources of truth (github.com/grafana/plugins-private @ 267f4937806ed6404b6628d13ae358a5d308e376,
 * plugins/grafana-cockroachdb-datasource):
 * - `src/plugin.json` — plugin `id`, `name`, docs link
 * - `src/types.ts` — `CockroachOptions extends SQLOptions`, `CockroachTLSModes`,
 *   `CockroachTLSMethods`, `CockroachAuthenticationType`, `CockroachSecureJsonData`
 * - `src/components/ConfigEditor/ConfigEditor.tsx` — the configuration editor
 * - `src/components/ConfigEditor/{Kerberos,ConnectionLimits,TLSSecretsConfig}.tsx` — sub-components
 * - `pkg/plugin/settings.go` — backend `Settings` + `LoadSettings` (reads url/user/database
 *   from jsonData; password from secureJsonData)
 * - `pkg/plugin/{driver,tlsmanager}.go`, `pkg/kerberos/kerberos.go` — how settings are consumed
 * - External:
 *   - `@grafana/plugin-ui` ^0.13.1 (catalog, resolved 0.13.1): `SQLOptions`,
 *     `SQLConnectionLimits`, `ConfigSection`, `ConfigSubSection`, `DataSourceDescription`
 *   - `@grafana/ui` ^11.6.7 (catalog, resolved 11.6.14): `Field`, `Input`, `Select`,
 *     `SecretInput`, `SecretTextArea`, `Switch`, `Tooltip`, `Icon`, `Label`, `Stack`
 */

/** jsonData.sslmode — `CockroachTLSModes` (`src/types.ts:7-12`). */
export type CockroachTLSMode = 'disable' | 'require' | 'verify-ca' | 'verify-full';

/** jsonData.tlsConfigurationMethod — `CockroachTLSMethods` (`src/types.ts:14-17`). */
export type CockroachTLSMethod = 'file-path' | 'file-content';

/** jsonData.authType — `CockroachAuthenticationType` (`src/types.ts:38-42`). Stored as the full label string. */
export type CockroachAuthType = 'SQL Authentication' | 'Kerberos Authentication' | 'TLS/SSL Authentication';

/**
 * Root (top-level datasource settings) fields the CockroachDB backend reads.
 *
 * The plugin stores NOTHING at the root level: `LoadSettings`
 * (`pkg/plugin/settings.go:245-274`) unmarshals `config.JSONData` for url/user/database
 * and only pulls `password` out of `config.DecryptedSecureJSONData`. Root url/user/database
 * are never read. Blank object per convention (never null).
 */
export type RootConfig = Record<string, never>;

/**
 * Fields stored in `jsonData`, keyed by their raw storage names.
 *
 * `CockroachOptions extends SQLOptions` (`src/types.ts:18-32`). The interface also declares
 * `postgresVersion?`, `timescaledb?` and `enableSecureSocksProxy?`, but the config editor never
 * writes the first two and the backend never reads them (dead/vestigial type fields — see README),
 * while `enableSecureSocksProxy` is the Secure Socks Proxy toggle excluded from this registry entry
 * per repo policy. Only the fields the editor writes AND/OR the backend reads are modeled here.
 */
export type JsonDataConfig = {
  /** CockroachDB host+port, e.g. `"localhost:26257"`. Written by the editor to jsonData (ConfigEditor.tsx:82-83); read at `pkg/plugin/settings.go:31,247`. */
  url?: string;
  /** Database name. Editor: ConfigEditor.tsx:93-94; backend: `pkg/plugin/settings.go:29`. */
  database?: string;
  /** User. Editor: ConfigEditor.tsx:118-119; backend: `pkg/plugin/settings.go:32`. */
  user?: string;
  /** Authentication discriminator. Editor: ConfigEditor.tsx:102-111; backend: `pkg/plugin/settings.go:39`. */
  authType?: CockroachAuthType;
  /** TLS negotiation mode. Editor default `'require'` (ConfigEditor.tsx:331). Only used for TLS/SSL Authentication. */
  sslmode?: CockroachTLSMode;
  /** How TLS certificates are supplied. Editor default `'file-content'` (ConfigEditor.tsx:177-181). */
  tlsConfigurationMethod?: CockroachTLSMethod;
  /** File-path TLS credentials (only when tlsConfigurationMethod === 'file-path'). Editor: ConfigEditor.tsx:216-224. */
  sslRootCertFile?: string;
  sslCertFile?: string;
  sslKeyFile?: string;
  /** Kerberos: krb5 config file path. Editor default `'/etc/krb5.conf'` (ConfigEditor.tsx:305); backend: `pkg/kerberos/kerberos.go:21`. */
  configFilePath?: string;
  /** Kerberos: credential cache path (required). Editor: Kerberos.tsx:22-37; backend: `pkg/kerberos/kerberos.go:20`. */
  credentialCache?: string;
  /** Kerberos: optional krbsrvname, default `'postgres'`. Editor: Kerberos.tsx:38-49; backend: `pkg/plugin/settings.go:38,139`. */
  kerberosServerName?: string;
  /** Connection pool max open connections. Backend default 5 (`pkg/plugin/settings.go:20,254-256`). */
  maxOpenConns?: number;
  /** Connection pool max idle connections. Backend default 2 (`pkg/plugin/settings.go:21,257-259`). */
  maxIdleConns?: number;
  /**
   * FRONTEND-ONLY: when true the editor keeps maxIdleConns in sync with maxOpenConns
   * (ConnectionLimits.tsx:70-92). Written to jsonData but never read by the backend
   * (`Settings` has no such field).
   */
  maxIdleConnsAuto?: boolean;
  /** Connection pool max lifetime (seconds). Backend default 300 (`pkg/plugin/settings.go:22,260-262`). */
  connMaxLifetime?: number;
  /** Query timeout (seconds). Backend default 30, clamped to 5-600 (`pkg/plugin/settings.go:24,263-272`). */
  queryTimeout?: number;
};

/**
 * Secret key names stored in `secureJsonData` (write-only):
 * - `password` — user password. Read at `pkg/plugin/settings.go:251`.
 * - `tlsCACert`, `tlsClientCert`, `tlsClientKey` — supplied when
 *   `authType === 'TLS/SSL Authentication'` and `tlsConfigurationMethod === 'file-content'`.
 *   Read at `pkg/plugin/tlsmanager.go:40-47,108-110`.
 */
export type SecureJsonDataConfig = Array<'password' | 'tlsCACert' | 'tlsClientCert' | 'tlsClientKey'>;
