/**
 * Configuration models for the Grafana CSV datasource plugin
 * (plugin id: `marcusolsson-csv-datasource`).
 *
 * Sources of truth (https://github.com/grafana/grafana-csv-datasource @ 5de0466):
 * - `src/plugin.json:1-47` — plugin id (`"marcusolsson-csv-datasource"`),
 *   name (`"CSV"`), docs URL
 *   (`info.links[0].url` = `"https://grafana.com/docs/plugins/marcusolsson-csv-datasource/latest/"`),
 *   `backend: true`, `grafanaDependency: ">=11.6.0-0"`.
 * - `src/ConfigEditor.tsx:1-136` — the config editor. Composes
 *   @grafana/plugin-ui `DataSourceDescription` (dataSourceName="CSV",
 *   docsLink hard-coded at `:58`, `hasRequiredFields={false}` at `:59`), a
 *   `<Field label="Storage Location">` RadioButtonGroup that toggles
 *   `jsonData.storage` between `'http'` and `'local'` (`:64-73`), then
 *   either:
 *     - HTTP branch (`:77-112`): `ConnectionSettings` with
 *       `urlPlaceholder="http://localhost:8080"` (`:82`), `Auth` via
 *       `convertLegacyAuthProps` (`:87-92`), and a collapsible
 *       `ConfigSection title="Additional settings"` (`:96-110`) containing
 *       `AdvancedHttpSettings` (`:97`) and a
 *       `<Field label="Custom query parameters" description="Add custom
 *       parameters to your queries.">` `Input` with placeholder `"limit=100"`
 *       (`:101-108`) that writes `jsonData.queryParams`.
 *     - Local branch (`:114-124`): `<Field label="Path">` `Input` with
 *       placeholder `"Path to the CSV file"` (`:116-122`) that writes back
 *       to `config.url` — the SAME root storage key as the HTTP URL.
 * - `src/utils.ts:4-10` — `getOptionsWithDefaults` returns the options
 *   unchanged when `jsonData.storage` is set; otherwise applies
 *   `defaultOptions = { storage: 'http' }` (`types.ts:47-49`).
 * - `src/types.ts:9-51` — the frontend types. `CSVDataSourceOptions extends
 *   DataSourceJsonData` has only `storage?: string` and `queryParams?:
 *   string` at datasource level (`:42-45`). Everything else — delimiter,
 *   schema, header, skipRows, decimalSeparator, timezone, ignoreUnknown,
 *   method, path, params, headers, body, experimental.regex — is per-query
 *   state on `CSVQuery` (`:9-28`) and never persisted on the datasource.
 * - `pkg/settings.go:1-28` — backend `PluginSettings { Storage string,
 *   QueryParams string }` and `LoadPluginSettings` unmarshaling
 *   `settings.JSONData`, then defaulting `Storage = "http"` when empty for
 *   backwards compatibility (`:22-24`).
 * - `pkg/datasource.go:34-47` — `NewDatasource` reads
 *   `os.Getenv("GF_PLUGIN_ALLOW_LOCAL_MODE") == "true"` into a private
 *   `allowLocalMode` flag; every non-HTTP storage rejects with
 *   `"local mode has been disabled by your administrator"` when the flag
 *   is off (`:158-160`).
 * - `pkg/datasource.go:109-145` — `CheckHealth` loads plugin settings, then
 *   for `Storage == "http"` requires a non-empty `settings.URL` (`:121-125`).
 *   For any storage it calls `store.Stat` (`:135-140`) which for local
 *   storage `os.Stat`s the URL as a filesystem path.
 * - `pkg/http_storage.go:73-131` — HTTP request building. Parses
 *   `settings.URL + query.Path` (`:79-93`), merges `customSettings.QueryParams`
 *   into the request query string with the admin's values OVERRIDING the
 *   query editor's own params on key collision (`:102-109`).
 * - `pkg/local_storage.go:33-51` — local storage opens `filepath.ToSlash
 *   (settings.URL)` (or `settings.URL + query.Path`), refusing paths that
 *   escape the base via `..` (`:41-43`); `Stat` `os.Stat`s the URL.
 *
 * External components consulted at their pinned versions:
 * - `@grafana/plugin-ui@0.13.1` —
 *   - `DataSourceDescription` (renders "Before you can use the {dataSourceName}
 *     data source, you must configure it below..."; `hasRequiredFields=false`
 *     suppresses the "Fields marked with * are required" note).
 *   - `ConnectionSettings` — writes `config.url`; default label `"URL"`,
 *     required=true, invalid state uses "Please enter a valid URL" error and
 *     the URL regex at `Connection/ConnectionSettings.js:16`.
 *   - `Auth` + `convertLegacyAuthProps` — the modern discriminator model.
 *     `AuthMethodSettings.js:9-30` defines the four `defaultOptions` (Basic
 *     authentication / Enable cross-site access control requests / Forward
 *     OAuth Identity / No Authentication). Default `visibleMethods` is
 *     `[BasicAuth, OAuthForward, NoAuth]` (`:47-52`) — CrossSiteCredentials
 *     is deliberately hidden. `utils.js:36-48` `getOnAuthMethodSelectHandler`
 *     writes BOTH `root.basicAuth` and `jsonData.oauthPassThru` on every
 *     selection (only one true).
 *   - `BasicAuth` (`auth-method/BasicAuth.js:6-79`) — labels default to
 *     `User` / `Password` with matching tooltips, and writes
 *     `root.basicAuthUser` + `secureJsonData.basicAuthPassword`.
 *   - `TLSSettings` composing `SelfSignedCertificate`, `TLSClientAuth`,
 *     `SkipTLSVerification` — labels: "Add self-signed certificate",
 *     "CA Certificate", "TLS Client Authentication", "ServerName",
 *     "Client Certificate", "Client Key", "Skip TLS certificate validation".
 *     Textareas use `placeholder="Begins with --- BEGIN CERTIFICATE ---"`
 *     and `placeholder="Begins with --- RSA PRIVATE KEY CERTIFICATE ---"`.
 *   - `CustomHeaders` — dynamic `httpHeaderName<N>` / `httpHeaderValue<N>`
 *     storage pattern; NOT modeled here as first-class fields (see README).
 *   - `AdvancedHttpSettings` (`AdvancedSettings/AdvancedHttpSettings.js:35-77`)
 *     — writes `jsonData.keepCookies` (TagsInput, placeholder `"New cookie
 *     (hit enter to add)"`) and `jsonData.timeout` (Input type="number",
 *     placeholder `"Timeout in seconds"`), both under a
 *     "Advanced HTTP settings" ConfigSubSection.
 * - `@grafana/ui@13.1.0-25893932881` — `Divider`, `Field`, `Input`,
 *   `RadioButtonGroup`, `useStyles2`.
 * - `@grafana/data@13.1.0-25893932881` — `DataSourceJsonData`,
 *   `DataSourcePluginOptionsEditorProps`, `GrafanaTheme2`.
 * - `@grafana/schema@13.1.0-25893932881` — `DataQuery` (base of `CSVQuery`).
 *
 * The Secure Socks Proxy widget is NOT rendered by this plugin's editor at
 * any point; there is no reference to `SecureSocksProxySettings` or the
 * `config.secureSocksDSProxyEnabled` runtime flag in `src/ConfigEditor.tsx`.
 * The corresponding `jsonData.enableSecureSocksProxy` storage key is
 * excluded per AGENTS.md regardless.
 */

/**
 * `jsonData.storage` discriminates between HTTP and local-file storage
 * backends. Options mirror `src/ConfigEditor.tsx:66-69`. The frontend
 * `getOptionsWithDefaults` (`src/utils.ts:4-10`) and the backend
 * `LoadPluginSettings` (`pkg/settings.go:22-24`) both default an empty
 * value to `'http'` for backwards compatibility.
 */
export type StorageMode = 'http' | 'local';

/**
 * Root (top-level datasource settings) fields the Grafana CSV plugin uses.
 *
 * `url` is DUAL-PURPOSE depending on `jsonData.storage`:
 *   - `storage === 'http'`: it is the base HTTP URL of the CSV endpoint
 *     (`pkg/http_storage.go:79`). Editor label `"URL"`, placeholder
 *     `"http://localhost:8080"`.
 *   - `storage === 'local'`: it is a filesystem path used as the base
 *     directory for CSV file loads (`pkg/local_storage.go:33-45`). Editor
 *     label `"Path"`, placeholder `"Path to the CSV file"`. Local mode is
 *     admin-gated by `GF_PLUGIN_ALLOW_LOCAL_MODE=true` on the plugin
 *     process (`pkg/datasource.go:44,158-160`).
 *
 * `basicAuth`, `basicAuthUser`, and `withCredentials` are populated by
 * @grafana/plugin-ui's `Auth` component via `convertLegacyAuthProps` and
 * consumed by the SDK's `settings.HTTPClientOptions(ctx)` call at
 * `pkg/datasource.go:35`. The CSV plugin's own Go code never touches them
 * by name. `withCredentials` is not user-selectable in the default
 * `visibleMethods` list, but a provisioned datasource may still carry it.
 */
export type RootConfig = {
  /** Complete HTTP URL (storage='http') or filesystem path (storage='local'). Written by `ConnectionSettings` or the "Path" Field respectively. */
  url?: string;
  /** True when HTTP Basic authentication is enabled. Written by @grafana/plugin-ui `utils.js:40`. Only rendered when storage='http'. */
  basicAuth?: boolean;
  /** Basic-auth username. Only meaningful when `basicAuth === true`. */
  basicAuthUser?: string;
  /**
   * Written by @grafana/plugin-ui's `getOnAuthMethodSelectHandler` when the
   * "Cross-site access control" method is selected. The CSV editor does not
   * offer that method in its default `visibleMethods` (`AuthMethodSettings.js:47-52`),
   * so this stays `false` in practice — but a provisioning payload can
   * still set it to true.
   */
  withCredentials?: boolean;
};

/**
 * Fields stored in `jsonData`. The plugin defines just two on the
 * datasource itself (`src/types.ts:42-45`): `storage` and `queryParams`.
 * Every other jsonData field below is written by @grafana/plugin-ui's
 * `Auth` (`convertLegacyAuthProps`) or `AdvancedHttpSettings` and read by
 * the SDK's `HTTPClientOptions`, not by any CSV-plugin-owned code path.
 */
export type JsonDataConfig = {
  /** Storage backend selector. Default `'http'`. See `StorageMode`. */
  storage?: StorageMode;
  /**
   * URL-encoded key=value pairs (e.g. `'limit=100&format=csv'`) appended to
   * every outgoing HTTP request. Merged into per-query params with
   * ADMIN-CONFIGURED VALUES OVERRIDING the per-query values on key collision
   * (`pkg/http_storage.go:102-109`). Only meaningful when `storage='http'`.
   */
  queryParams?: string;

  /** Enable TLS client authentication (mTLS). Requires `serverName`, `tlsClientCert`, `tlsClientKey`. Written by @grafana/plugin-ui `utils.js:101-108`. */
  tlsAuth?: boolean;
  /** Enable custom CA verification. Requires `tlsCACert`. Written by @grafana/plugin-ui `utils.js:77-85`. */
  tlsAuthWithCACert?: boolean;
  /** Skip TLS certificate validation. Not recommended outside testing. Written by @grafana/plugin-ui `utils.js:153-157`. */
  tlsSkipVerify?: boolean;
  /** TLS SNI / cert-verification server name (@grafana/plugin-ui `TLSClientAuth.js:50`). */
  serverName?: string;
  /** HTTP request timeout in seconds (@grafana/plugin-ui `AdvancedHttpSettings.js:64-75`). */
  timeout?: number;
  /** Cookies to forward to the datasource, by name (@grafana/plugin-ui `AdvancedHttpSettings.js:44-53`). */
  keepCookies?: string[];
  /**
   * Written by the "Forward OAuth Identity" auth method selector
   * (@grafana/plugin-ui `utils.js:44`). When true, the SDK forwards the
   * signed-in user's OAuth identity to the datasource.
   */
  oauthPassThru?: boolean;
};

/**
 * Secret key names stored in `secureJsonData` (write-only; read existing
 * config via `secureJsonFields`):
 * - `basicAuthPassword` — user password when Basic auth is enabled.
 * - `tlsCACert` — custom CA PEM when `tlsAuthWithCACert` is true.
 * - `tlsClientCert`, `tlsClientKey` — mTLS client credentials when
 *   `tlsAuth` is true.
 *
 * The editor also writes dynamic `httpHeaderValue<N>` secrets when the user
 * configures custom HTTP headers via @grafana/plugin-ui's `CustomHeaders`
 * component. Those keys are indexed pairs — not modeled as first-class
 * fields in this schema; see the README.
 *
 * The CSV datasource plugin itself defines no plugin-specific secrets
 * beyond this shared HTTP-settings set.
 */
export type SecureJsonDataConfig = Array<
  'basicAuthPassword' | 'tlsCACert' | 'tlsClientCert' | 'tlsClientKey'
>;
