/**
 * Configuration models for the Falcon LogScale datasource plugin (`grafana-falconlogscale-datasource`).
 *
 * Sources of truth (https://github.com/grafana/falconlogscale-datasource @ 1e83b294390e6d93865156cec0c72ed252e791c0):
 * - `src/plugin.json:5` — plugin id (`grafana-falconlogscale-datasource`), name ("Falcon LogScale")
 * - `src/types.ts:5-29` — frontend `DataSourceMode`, `LogScaleOptions`, `SecretLogScaleOptions`
 * - `src/components/ConfigEditor/ConfigEditor.tsx` — outermost editor (composes @grafana/plugin-ui
 *   `DataSourceDescription`, `ConnectionSettings`, `Auth` via `convertLegacyAuthProps`,
 *   `AdvancedHttpSettings`, `ConfigSection`, plus plugin-local `OAuth2Component`,
 *   `DefaultRepository`, and `DataLinks`)
 * - `src/components/ConfigEditor/OAuth2Component.tsx` — Client ID / Client Secret fields
 * - `src/components/ConfigEditor/DefaultRepository.tsx` — Default Repository select + Load button
 * - `src/components/DataLinks/DataLinks.tsx` + `DataLink.tsx` + `types.ts` — data links list
 * - `pkg/plugin/settings.go:11-63` — backend `Settings` struct and `LoadSettings`
 * - `pkg/plugin/plugin.go:52-70` — how each setting is consumed by the HTTP client
 * - `pkg/plugin/healthcheck.go:26-48` — mode-conditional health check behavior
 *
 * External components consulted at their pinned versions:
 * - `@grafana/plugin-ui@^0.13.0` — `ConnectionSettings` (URL input), `Auth` with
 *   `convertLegacyAuthProps`, `AdvancedHttpSettings` (Allowed cookies, Timeout),
 *   `DataSourceDescription`, `ConfigSection`
 * - `@grafana/ui@^11.3.2` — `Field`, `Input`, `SecretInput`, `Select`, `Switch`, `Button`,
 *   `DataLinkInput`
 * - `@grafana/runtime@^11.6.0` — `config` (reads `config.secureSocksDSProxyEnabled` for the
 *   excluded Secure Socks Proxy toggle), `getBackendSrv`
 * - `@grafana/data@^11.6.0` — `DataSourceJsonData` base interface, `isValidDuration`
 */

/**
 * Data source mode discriminator. Selects between the LogScale product (default) and
 * CrowdStrike NGSIEM, which only supports OAuth2 client credentials authentication and
 * automatically pins the default repository to `search-all`.
 *
 * Defined in `src/types.ts:5-8`.
 */
export type DataSourceMode = 'LogScale' | 'NGSIEM';

/**
 * A single data-link configuration entry, as rendered by `DataLink.tsx` and consumed by
 * `src/logs.ts` when transforming backend results into log-row data links. When
 * `datasourceUid` is set the entry becomes an internal link (the URL is interpolated as a
 * Query on that data source); otherwise it becomes an external URL template.
 *
 * Defined in `src/components/DataLinks/types.ts:1-7`.
 */
export type DataLinkConfig = {
  /** Field name (or regex pattern that matches on a field name). Required. */
  field: string;
  /** Human-readable label shown on the derived data link. */
  label: string;
  /** Regex applied to the field value; captured groups are usable in the URL template. Required. */
  matcherRegex: string;
  /** URL template (external) or Query text (internal). Supports interpolation like `${__value.raw}`. */
  url: string;
  /** UID of a Grafana data source; when set the data link becomes an internal link. */
  datasourceUid?: string;
};

/**
 * Root (top-level datasource settings) fields the Falcon LogScale plugin cares about.
 *
 * `url` is read directly by the backend (`pkg/plugin/settings.go:38`) — the backend hard-fails
 * with "URL can not be blank" when empty. `basicAuth` and `basicAuthUser` are populated by
 * @grafana/plugin-ui's `Auth` component when the Basic authentication method is selected; the
 * plugin's backend reads `config.BasicAuthUser` at `pkg/plugin/settings.go:59` and pairs it
 * with the decrypted `basicAuthPassword` secret.
 */
export type RootConfig = {
  /** Complete HTTP URL of the LogScale (or NGSIEM) server. Required by the backend. */
  url?: string;
  /** True when HTTP Basic authentication is enabled. Written by @grafana/plugin-ui's Auth component. */
  basicAuth?: boolean;
  /** Basic-auth username. Only meaningful when `basicAuth === true`. */
  basicAuthUser?: string;
};

/**
 * Fields stored in `jsonData`. Combines the plugin's `LogScaleOptions` (`src/types.ts:10-23`) —
 * excluding the Secure Socks Proxy toggle (`enableSecureSocksProxy`, deliberately excluded from
 * registry entries) and the dead `basicAuthUser` declared on the frontend type but never written
 * or read from jsonData — with the HTTP fields written by @grafana/plugin-ui's
 * `AdvancedHttpSettings`.
 */
export type JsonDataConfig = {
  /**
   * Frontend-only snapshot of `root.url` written by the LogScale token authentication component
   * (`ConfigEditor.tsx:155`). Never read by the backend, which always reads `settings.URL`
   * (`pkg/plugin/settings.go:38`). The upstream backend struct declares `json:"baseURL"` (Go
   * case-insensitive unmarshal happens to match) but its value is unconditionally overwritten
   * by `settings.BaseURL = config.URL` before use.
   */
  baseUrl?: string;
  /**
   * True when authenticating via a LogScale personal token stored in `secureJsonData.accessToken`.
   * When true, the backend derives `<url>/humio/graphql` (GraphQL) and `<url>/humio` (REST) as the
   * client endpoints (`pkg/plugin/settings.go:44-50`); when false, the endpoints are `<url>/graphql`
   * and `<url>` respectively.
   */
  authenticateWithToken?: boolean;
  /** True when the "Forward OAuth Identity" auth method is selected. */
  oauthPassThru?: boolean;
  /** True when authenticating via OAuth2 client credentials (required in NGSIEM mode). */
  oauth2?: boolean;
  /** OAuth2 client ID. Only meaningful when `oauth2 === true`. */
  oauth2ClientId?: string;
  /**
   * Data source mode discriminator. Defaults to `LogScale` in the editor
   * (`ConfigEditor.tsx:59`). In `NGSIEM` mode the backend appends `/humio` to the base URL
   * (`pkg/plugin/plugin.go:54-56`) and health check switches to OAuth2 client credentials
   * verification (`pkg/plugin/healthcheck.go:26-36`).
   */
  mode?: DataSourceMode;
  /**
   * Default repository / view name used when a query has no explicit repository set.
   * In NGSIEM mode this is auto-set to `search-all` and restricted to
   * `['search-all', 'investigate_view', 'third-party']` (`src/types.ts:52`).
   */
  defaultRepository?: string;
  /** Derived data links applied to log rows by the frontend result transformer (`src/logs.ts`). */
  dataLinks?: DataLinkConfig[];
  /**
   * Enable incremental querying: on auto-refresh, only fetch the newest window and merge it
   * with the cached previous result. Consumed by `src/DataSource.ts:99-105`; the backend is
   * unaffected. Marked "experimental" by the editor.
   */
  incrementalQuerying?: boolean;
  /**
   * Duration window re-fetched on each incremental query (e.g. `10m`, `30s`, `1h`). Frontend-only;
   * default `10m` (`src/incrementalQuery.ts` `DEFAULT_OVERLAP_WINDOW`).
   */
  incrementalQueryOverlapWindow?: string;
  /**
   * Cookies to forward to the datasource, by name. Written by @grafana/plugin-ui's
   * `AdvancedHttpSettings`, consumed by the SDK's `HTTPClientOptions`.
   */
  keepCookies?: string[];
  /** HTTP request timeout in seconds, written by @grafana/plugin-ui's `AdvancedHttpSettings`. */
  timeout?: number;
};

/**
 * Secret key names stored in `secureJsonData` (write-only; read existing config via
 * `secureJsonFields`):
 * - `accessToken` — LogScale personal token; set when `jsonData.authenticateWithToken` is true
 * - `oauth2ClientSecret` — OAuth2 client secret; set when `jsonData.oauth2` is true
 * - `basicAuthPassword` — HTTP Basic-auth password; set when `root.basicAuth` is true
 */
export type SecureJsonDataConfig = Array<'accessToken' | 'oauth2ClientSecret' | 'basicAuthPassword'>;
