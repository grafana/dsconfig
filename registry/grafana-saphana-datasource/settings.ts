/**
 * Configuration models for the SAP HANA® datasource plugin (`grafana-saphana-datasource`).
 *
 * Sources of truth (github.com/grafana/plugins-private @ 267f4937806ed6404b6628d13ae358a5d308e376,
 * plugins/grafana-saphana-datasource):
 * - `src/types.ts:9-29` — `HANAConfig` (jsonData) and `HANASecureConfig` (secureJsonData)
 * - `src/components/ConfigEditor.tsx` — the configuration editor
 * - `src/selectors.ts:2-74` — every editor label / placeholder / tooltip (`Components.ConfigEditor`)
 * - `pkg/models/settings.go:18-83` — backend `Settings` struct + `LoadSettings`/`IsValid`
 *   (reads nothing at the root level; every field comes from jsonData + secureJsonData)
 * - `pkg/plugin/driver.go:37-107` — `GetTLSConfig` / `GetConnection` show how each setting is
 *   consumed to build the SAP HANA connection (server:port, 3<instance>13 derivation, TLS)
 * - External:
 *   - `@grafana/ui` ^11.6.7 (`catalog:` in `.yarnrc.yml`): `LegacyForms.FormField`,
 *     `LegacyForms.SecretFormField` (password), `Switch`, `InlineFormLabel`, `InlineField`
 *   - `@grafana/data` ^11.6.7 (`catalog:`): `DataSourceJsonData` (base of `HANAConfig`),
 *     `onUpdateDatasourceJsonDataOption`, `onUpdateDatasourceSecureJsonDataOption`
 *   - `@grafana/runtime` ^11.6.7 (`catalog:`): `config` (feature-toggle gate for Secure Socks Proxy)
 *   - `src/components/ui/CertificationKey.tsx` — plugin-local textarea used for the three TLS
 *     certificate/key secrets (no tooltip, only a placeholder)
 */

/**
 * Root (top-level datasource settings) fields.
 *
 * The SAP HANA backend reads nothing at the root level — `LoadSettings`
 * (`pkg/models/settings.go:54-83`) unmarshals only `config.JSONData` and reads secrets from
 * `config.DecryptedSecureJSONData`; the connection is built entirely from jsonData + secrets
 * (`pkg/plugin/driver.go:62-107`). So this is a blank object rather than null.
 */
export type RootConfig = Record<string, never>;

/**
 * Fields stored in `jsonData`. Mirrors the plugin's `HANAConfig` (`src/types.ts:9-22`) minus
 * `enableSecureSocksProxy` (Secure Socks Proxy, intentionally excluded from this entry), and
 * matches the backend `Settings` json tags (`pkg/models/settings.go:20-34`).
 */
export type JsonDataConfig = {
  /** "Server address" — SAP HANA server address (`selectors.ts:4-8`). Always required; also used as the TLS ServerName (`driver.go:40`). */
  server: string;
  /**
   * "Server port" (`selectors.ts:9-14`). The editor renders a number input and coerces via
   * `+port` (`ConfigEditor.tsx:40-48`); the backend stores it as `int64` (`settings.go:30`).
   * Optional when both `instance` and `databaseName` are provided.
   */
  port?: number;
  /**
   * "Tenant instance number" (`selectors.ts:21-25`). NOTE: typed `number` upstream but actually
   * stored as a STRING — the editor writes it with the generic `onUpdateDatasourceJsonDataOption`
   * (`ConfigEditor.tsx:269`), not the `+port` coercion, and the backend reads it as a `string`
   * (`settings.go:21`), concatenating it into the port as `3<instance>13` (`driver.go:73`).
   */
  instance?: number;
  /** "Tenant database name" (`selectors.ts:15-20`). With `instance`, replaces the explicit port. */
  databaseName?: string;
  /** "Username" — SAP HANA username (`selectors.ts:26-30`). Required unless `tlsAuth` is enabled. */
  username: string;
  /** "Default schema" to be used; can be empty (`selectors.ts:65-69`). Applied via `conn.SetDefaultSchema` (`driver.go:101-103`). */
  defaultSchema?: string;
  /** "Timeout" for connections, in seconds, stored as a string (`selectors.ts:70-73`). Backend defaults to "30" (`settings.go:72-74`). */
  timeout?: string;
  /**
   * "TLS" switch (`selectors.ts:36-40`). Stored INVERTED: the switch shows `!tlsDisabled`
   * (`ConfigEditor.tsx:173`), so switch ON = TLS enabled = `tlsDisabled` false (the default).
   * When set true the editor also clears the three switches below (`ConfigEditor.tsx:54-58`).
   */
  tlsDisabled?: boolean;
  /** "Skip TLS Verify" (`selectors.ts:41-44`). Maps to `tls.Config.InsecureSkipVerify` (`driver.go:39`). Editor-visible only when TLS is enabled. */
  tlsSkipVerify?: boolean;
  /** "TLS Client Auth" (`selectors.ts:45-48`). Enables X.509 mutual TLS with the client cert/key. Editor-visible only when TLS is enabled. */
  tlsAuth?: boolean;
  /** "With CA Cert" — needed for verifying self-signed TLS certs (`selectors.ts:49-52`). Editor-visible only when TLS is enabled. */
  tlsAuthWithCACert?: boolean;
};

/**
 * Secret key names stored in `secureJsonData` (write-only; read existing config via
 * `secureJsonFields`). Mirrors `HANASecureConfig` (`src/types.ts:24-29`):
 * - `password` — SAP HANA password (basic auth). Consumed at `driver.go:82`.
 * - `tlsClientCert`, `tlsClientKey` — X.509 client key-pair, supplied together when `tlsAuth`
 *   is enabled. Consumed at `driver.go:51,77`.
 * - `tlsCACert` — custom/self-signed CA certificate, used when `tlsAuthWithCACert` is enabled.
 *   Consumed at `driver.go:43-48`.
 */
export type SecureJsonDataConfig = Array<'password' | 'tlsClientCert' | 'tlsClientKey' | 'tlsCACert'>;
