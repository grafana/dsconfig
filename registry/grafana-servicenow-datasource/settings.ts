/**
 * Configuration models for the ServiceNow datasource plugin
 * (plugin id: `grafana-servicenow-datasource`).
 *
 * Sources of truth (github.com/grafana/plugins-private @ 267f4937, plugin path
 * `plugins/grafana-servicenow-datasource`):
 * - `src/plugin.json:3,4,38` — plugin name (`"ServiceNow"`), id
 *   (`"grafana-servicenow-datasource"`), docs link
 *   (`"https://grafana.com/docs/plugins/grafana-servicenow-datasource"`).
 * - `src/types/index.ts:5-27` — the frontend config types `ServiceNowOptions`
 *   (jsonData), `ServiceNowAuthMethod`, `authOptions`, `ServiceNowSecureOptions`.
 * - `src/components/ConfigEditor.tsx:44-302` — the configuration editor:
 *   - `<ConfigSection title="ServiceNow Instance Settings">` (`:135`).
 *   - Authentication Type radio -> jsonData.authMethod (`:136-143`), value
 *     derived on load by `initAuthMethod` (`:25-38`) from authMethod ??
 *     (oauthEnabled ? serviceNowOAuth : basicAuth).
 *   - URL -> root `options.url` (`:145-158`), Username -> root
 *     `options.basicAuthUser` (`:160-169`), Password -> secureJsonData.basicAuthPassword
 *     (`:171-181`).
 *   - authMethod==='serviceNowOAuth' block (`:183-212`): Client ID ->
 *     jsonData.oauthClientID (`:185-194`), Client Secret ->
 *     secureJsonData.oauthClientSecret (`:196-210`).
 *   - Permissions help modal (`:214-216`, `PermissionsHelp.tsx`), Use Sys Tables?
 *     switch -> jsonData.useSysTables (`:218-232`), Query Timeout number ->
 *     jsonData.queryTimeoutSeconds (`:234-254`), and CustomHeadersSettings
 *     (`:256`) which writes dynamic httpHeaderName<N>/httpHeaderValue<N> pairs.
 *   - Secure Socks Proxy switch -> jsonData.enableSecureSocksProxy
 *     (`:259-299`, excluded here).
 * - `src/selectors.ts:28-64` — the `Components.ConfigEditor` label/placeholder/
 *   tooltip/id map the editor binds to (labels and tooltips live here).
 * - `pkg/models/settings.go:18-135` — backend `Settings` struct, `LoadSettings`,
 *   and `IsValid`.
 * - `pkg/models/auth_method.go:4-23` — `AuthMethod` alias, its two constants, and
 *   `GetAuthMethod` (legacy `oauthEnabled` fallback).
 * - `pkg/httputil/client.go:37-82` / `pkg/httputil/auth.go:23-116` — how each
 *   field is consumed: Basic auth via `SetBasicAuth(username, password)` vs the
 *   OAuth2 password grant (client_id/client_secret/username/password).
 *
 * External components consulted at their pinned versions (plugin `package.json`
 * -> monorepo `.yarnrc.yml` catalog):
 * - `@grafana/ui@^11.6.7` — `RadioButtonGroup`, `Input`, `Switch`, `SecretInput`,
 *   `InlineField`, `InlineFormLabel`, `useTheme2`.
 * - `@grafana/data@^11.6.7` — `DataSourceJsonData` (the base interface
 *   `ServiceNowOptions` extends), `DataSourcePluginOptionsEditorProps`,
 *   `FeatureToggles`.
 * - `@grafana/runtime@^11.6.7` — `config` (read to gate the Secure Socks Proxy
 *   switch).
 * - `@grafana/plugin-ui@^0.13.1` — `CustomHeadersSettings` (dynamic HTTP header
 *   pairs) and `ConfigSection`.
 *
 * The Secure Socks Proxy switch (`jsonData.enableSecureSocksProxy`) and the
 * dynamic custom-header pairs are deliberately excluded from this registry entry
 * (AGENTS.md exclusion / dynamic indexed keys).
 */

/** Authentication method discriminator (`src/types/index.ts:17`). */
export type ServiceNowAuthMethod = 'basicAuth' | 'serviceNowOAuth';

/**
 * Root (top-level datasource settings) fields the ServiceNow plugin actually
 * reads.
 *
 * `url` is the ServiceNow instance URL, read directly by the backend
 * (`pkg/models/settings.go:82`, `pkg/datasource.go:44`). `basicAuthUser` is the
 * ServiceNow account username, read by the backend for both auth methods
 * (`pkg/models/settings.go:91,98`) and used as the HTTP Basic username and the
 * OAuth2 password-grant username (`pkg/httputil/client.go:76`,
 * `pkg/httputil/auth.go:32`). Both are written by the config editor via
 * `options.url` / `options.basicAuthUser` (`src/components/ConfigEditor.tsx:67,70`).
 *
 * The standard root `basicAuth` (enabled) boolean is intentionally NOT modeled:
 * the editor never writes it and the backend ignores its stored value, deriving
 * `BasicAuthEnabled` from `authMethod` instead (`pkg/models/settings.go:87-96`,
 * confirmed by `pkg/models/settings_test.go:151-175`).
 */
export type RootConfig = {
  /** ServiceNow instance URL, e.g. `https://<INSTANCE_ID>.service-now.com`. Required; the backend rejects an empty URL. */
  url?: string;
  /** ServiceNow account username. Used by both Basic auth and the ServiceNow OAuth password grant. */
  basicAuthUser?: string;
};

/**
 * Fields stored in `jsonData`. Matches the plugin's frontend `ServiceNowOptions`
 * (`src/types/index.ts:5-15`, minus `enableSecureSocksProxy`) and the json tags
 * the backend `LoadSettings` reads from `config.JSONData`
 * (`pkg/models/settings.go:41-51`).
 */
export type JsonDataConfig = {
  /**
   * Authentication method (`src/types/index.ts:6,17`). Radio: Basic auth=`basicAuth`,
   * ServiceNow OAuth=`serviceNowOAuth` (`src/types/index.ts:19-22`). Written directly
   * by the editor (`src/components/ConfigEditor.tsx:112-121`). The backend treats an
   * empty value as `basicAuth` (`pkg/models/auth_method.go:12-22`).
   */
  authMethod?: ServiceNowAuthMethod;
  /**
   * OAuth Client ID (`src/types/index.ts:8`). Shown only for `serviceNowOAuth`
   * (`src/components/ConfigEditor.tsx:185-194`). Sent as `client_id` in the OAuth2
   * password grant (`pkg/httputil/auth.go:25`). Stored in jsonData (not a secret).
   */
  oauthClientID?: string;
  /**
   * Legacy/deprecated (`src/types/index.ts:10-11`). Predates `authMethod`; older
   * versions set `oauthEnabled: true` to select OAuth. Not written by the current
   * editor but still read for backwards compatibility by `initAuthMethod`
   * (`src/components/ConfigEditor.tsx:25-38`) and `GetAuthMethod`
   * (`pkg/models/auth_method.go:17-20`): when `authMethod` is empty,
   * `oauthEnabled: true` resolves to `serviceNowOAuth`.
   */
  oauthEnabled?: boolean;
  /**
   * Query sys tables for schema/meta lookups (`src/types/index.ts:13`). Requires
   * elevated ServiceNow permissions. Read by the backend to skip the Schema API
   * (`pkg/datasource.go:53-55`).
   */
  useSysTables?: boolean;
  /**
   * Per-query timeout in seconds (`src/types/index.ts:14`). Editor default 30
   * (`src/components/ConfigEditor.tsx:51`); the backend clamps any value < 1 to 30
   * (`pkg/models/settings.go:73-77`) and applies it as a per-query context timeout
   * (`pkg/newyork/table_api_v2.go:519`).
   */
  queryTimeoutSeconds?: number;
  /**
   * Written by the Secure Socks Proxy switch (`src/components/ConfigEditor.tsx:287-294`)
   * and consumed transparently by the SDK's `config.HTTPClientOptions(ctx)` call in
   * `pkg/models/settings.go:68`. The plugin's own Go code never inspects it by name.
   * Deliberately excluded from the dsconfig registry entry per AGENTS.md.
   */
  enableSecureSocksProxy?: boolean;
};

/**
 * Secret key names stored in `secureJsonData` (write-only; read existing config
 * via `secureJsonFields`). Both are copied from `DecryptedSecureJSONData` in
 * `LoadSettings` (`pkg/models/settings.go:92,99,101`):
 * - `basicAuthPassword` — the ServiceNow account password. This is the standard
 *   Grafana Basic-auth secret key; used both as the HTTP Basic password and as
 *   the OAuth2 password-grant password.
 * - `oauthClientSecret` — the OAuth application's client secret (serviceNowOAuth).
 *
 * The `CustomHeadersSettings` editor also writes dynamic `httpHeaderValue<N>`
 * secrets for each custom HTTP header; those indexed keys are not modeled as
 * first-class fields here (see the README).
 */
export type SecureJsonDataConfig = Array<'basicAuthPassword' | 'oauthClientSecret'>;
