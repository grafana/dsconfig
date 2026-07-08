/**
 * Configuration models for the MongoDB datasource plugin (`grafana-mongodb-datasource`).
 *
 * Sources of truth (github.com/grafana/plugins-private @ 267f4937806ed6404b6628d13ae358a5d308e376,
 * plugins/grafana-mongodb-datasource):
 * - `src/types.ts:29-46` — `MongoDbJsonData` (frontend jsonData shape)
 * - `src/types.ts:52-57` — `SecureSettings` (frontend secureJsonData shape)
 * - `src/components/ConfigEditor.tsx` — the configuration editor (Connection, Additional Settings)
 * - `src/components/AuthConfig.tsx` — authentication section (method picker + Credentials)
 * - `src/components/KerberosAuthConfig.tsx` — Kerberos credential inputs
 * - `pkg/models/settings.go:15-45` — backend `Settings`; `LoadSettings` (`:50-162`) reads/migrates
 *   the fields and `pkg/datasource/client.go` consumes them
 * - External: `@grafana/plugin-ui` ^0.13.1 (catalog) — `Auth` + `convertLegacyAuthProps`
 *   (`dist/esm/components/ConfigEditor/Auth/utils.js`) store the Credentials username at the ROOT
 *   `basicAuthUser` field and the password at `secureJsonData.basicAuthPassword`; `AuthMethod`
 *   supplies the `NoAuth` / `BasicAuth` values.
 */

/** Authentication method stored in `jsonData.authType` (`AuthConfig.tsx:13`, validated `settings.go:66`). */
export type AuthType = 'NoAuth' | 'BasicAuth' | 'custom-Kerberos';

/**
 * Root (top-level datasource settings) fields the MongoDB backend actually reads.
 *
 * The Credentials method is wired through `@grafana/plugin-ui`'s `Auth` component, which stores the
 * username at the root `basicAuthUser` field. The backend reads `config.BasicAuthUser` and
 * `config.BasicAuthEnabled` (`settings.go:98,111-121`).
 */
export type RootConfig = {
  /**
   * Standard Grafana basic-auth enabled flag. Read as `BasicAuthEnabled` (`settings.go:98`) and
   * forced on when a username/password is present. Set by provisioning and the editor's legacy
   * migration (`ConfigEditor.tsx:299-307`), not by selecting the Credentials method.
   */
  basicAuth?: boolean;
  /** Credentials username. Written by `Auth`/`convertLegacyAuthProps`; read at `settings.go:111-119`. */
  basicAuthUser?: string;
};

/**
 * Fields stored in `jsonData`. Matches the frontend `MongoDbJsonData` (`src/types.ts:29-46`) and the
 * json-tagged fields of the backend `Settings` (`pkg/models/settings.go:15-45`).
 *
 * Excluded: `enableSecureSocksProxy` (Secure Socks Proxy is out of scope for registry entries).
 */
export type JsonDataConfig = {
  /** MongoDB connection string. Required. Read at `client.go:134`, `settings.go:17`. */
  connection?: string;
  /** Authentication discriminator. Default 'BasicAuth' (`ConfigEditor.tsx:18`, `settings.go:66-71`). */
  authType?: AuthType;
  /** Kerberos client principal username (`KerberosAuthConfig.tsx:65-70`, `settings.go:23`). */
  kerberosUser?: string;
  /** Kerberos keytab file path (`KerberosAuthConfig.tsx:90-96`, `settings.go:25`). */
  keyTabFilePath?: string;
  /** Kerberos global ccache file path (`KerberosAuthConfig.tsx:98-104`, `settings.go:26`). */
  globalCcacheFilePath?: string;
  /** Kerberos ccache lookup file path (`KerberosAuthConfig.tsx:106-112`, `settings.go:27`). */
  ccacheLookupFile?: string;
  /**
   * Backend-only: TLS server name used to verify the returned certificate when `tlsAuth` is set
   * (`client.go:428-429`, `settings.go:29`). Not rendered by the config editor; set via provisioning.
   */
  serverName?: string;
  /**
   * Backend-only: enables TLS client-certificate auth (`client.go:407`, `settings.go:36`).
   * Not rendered by the config editor; set via provisioning.
   */
  tlsAuth?: boolean;
  /**
   * Backend-only: verify the server certificate against a custom CA (`client.go:391`, `settings.go:35`).
   * Not rendered by the config editor; set via provisioning.
   */
  tlsAuthWithCACert?: boolean;
  /**
   * Skip TLS certificate verification (`client.go:401`, `settings.go:34`). Set by the editor's legacy
   * migration from `skipTLSValidation` (`ConfigEditor.tsx:308-310`) or via provisioning.
   */
  tlsSkipVerify?: boolean;
  /**
   * Frontend-only: enables real-time BSON syntax validation in the query editor
   * (`ConfigEditor.tsx:192-206`). Never read by the backend.
   */
  validate?: boolean;
  /**
   * Maximum rows returned by a query, stored as a string (`ConfigEditor.tsx:270-289`, `settings.go:37`).
   * The editor displays a fallback of 100000; the backend defaults an empty value to "10000"
   * (`settings.go:56-58`).
   */
  responseRowsLimit?: string;
  /**
   * Legacy: pre-v1.9.0 username. Migrated to root `basicAuthUser` by the backend
   * (`settings.go:42,102-107`). New configs use root `basicAuthUser`.
   */
  user?: string;
  /**
   * Legacy: pre-v1.9.0 skip-TLS flag copied into `tlsSkipVerify` by the backend
   * (`settings.go:44,143-148`). New configs use `tlsSkipVerify`.
   */
  skipTLSValidation?: boolean;
  /**
   * Legacy frontend-only flag; the editor migration reads it to detect old basic-auth datasources
   * (`ConfigEditor.tsx:299`, `src/types.ts:32`). Never read by the backend.
   */
  credentials?: boolean;
};

/**
 * Secret key names stored in `secureJsonData` (write-only; read existing config via `secureJsonFields`):
 * - `basicAuthPassword` — Credentials password (`AuthConfig.tsx:44`, `settings.go:122`).
 * - `kerberosPassword` — Kerberos password (`KerberosAuthConfig.tsx:46-54`, `settings.go:140`).
 * - `tlsCertificateKeyFilePassword` — password for an encrypted TLS client key (`ConfigEditor.tsx:248-268`, `client.go:419`).
 * - `tlsCACert`, `tlsClientCert`, `tlsClientKey` — TLS PEM material read from secureJsonData via
 *   `mapstructure.Decode` (`settings.go:154`); set via provisioning.
 * - `password` — legacy pre-v1.9.0 password (`settings.go:105`).
 */
export type SecureJsonDataConfig = Array<
  | 'basicAuthPassword'
  | 'kerberosPassword'
  | 'tlsCertificateKeyFilePassword'
  | 'tlsCACert'
  | 'tlsClientCert'
  | 'tlsClientKey'
  | 'password'
>;
