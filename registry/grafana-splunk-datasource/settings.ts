/**
 * Configuration models for the Splunk datasource plugin (`grafana-splunk-datasource`).
 *
 * Sources of truth (github.com/grafana/plugins-private @ 267f4937806ed6404b6628d13ae358a5d308e376,
 * plugins/grafana-splunk-datasource):
 * - `src/plugin.json:4` — plugin id (`grafana-splunk-datasource`), name ("Splunk")
 * - `src/types.ts:96-124` — frontend `SplunkOptions` (jsonData) and `SplunkSecureJsonData`
 * - `src/components/ConfigEditor.tsx` — outermost editor (composes @grafana/plugin-ui
 *   `DataSourceDescription`, `ConnectionSettings`, `DataLinks`, plus plugin-local
 *   `SplunkAuthComponent` and `AdditionalSettingsEditor`)
 * - `src/components/SplunkAuthComponent.tsx` — auth method selector (@grafana/plugin-ui `Auth`
 *   via `convertLegacyAuthProps`) with a custom 'custom-splunk' method (authToken)
 * - `src/components/AdditionalSettingsEditor.tsx` — @grafana/plugin-ui `AdvancedHttpSettings`
 *   (keepCookies, timeout) + the plugin's "Advanced options" fields
 * - `pkg/models/settings.go:26-137` — backend `Settings` struct and `LoadSettings`
 * - `pkg/splunk/client.go:62-71,229-230` — how URL and authToken are consumed
 *
 * External components consulted at their pinned versions:
 * - `@grafana/plugin-ui@^0.13.1` (.yarnrc.yml catalog) — `ConnectionSettings` (root.url),
 *   `Auth` + `convertLegacyAuthProps` (root.basicAuthUser, secureJsonData.basicAuthPassword,
 *   jsonData.oauthPassThru, and the TLS + custom-headers sub-sections),
 *   `AdvancedHttpSettings` (jsonData.keepCookies, jsonData.timeout), `DataLinks`,
 *   `DataSourceDescription`
 * - `@grafana/ui@^11.6.7` — `SecretTextArea`, `Input`, `Select`, `InlineSwitch`
 * - `@grafana/data@^11.6.7` — `DataSourceJsonData` base interface, update helpers
 */

/**
 * Authentication method persisted in `jsonData.authType`. The value is the id of the method
 * selected in the @grafana/plugin-ui `Auth` component. An empty/missing value is treated as
 * `BasicAuth` by the backend (`pkg/models/settings.go:95`).
 *
 * Defined via `AuthMethod` from `@grafana/plugin-ui` plus the plugin's custom `custom-splunk`
 * method (`src/components/SplunkAuthComponent.tsx:11,31,47`).
 */
export type SplunkAuthType = 'BasicAuth' | 'custom-splunk' | 'OAuthForward';

/** Fields search mode (`jsonData.fieldSearchType`). `src/components/AdditionalSettingsEditor.tsx:24-27`. */
export type FieldSearchType = 'quick' | 'full';

/** Variables search mode (`jsonData.variableSearchLevel`). `src/components/AdditionalSettingsEditor.tsx:29-33`. */
export type VariableSearchLevel = 'fast' | 'smart' | 'verbose';

/**
 * A single data-link configuration entry, written by @grafana/plugin-ui's `DataLink` and
 * consumed by the plugin (frontend result transformer; the backend also compiles the regex in
 * `pkg/models/settings.go:62-85`). When `datasourceUid` is set the entry becomes an internal
 * link (the URL is interpreted as a query on that data source).
 *
 * Mirrors the backend `DataLinkConfig` (`pkg/models/settings.go:14-23`).
 */
export type DataLinkConfig = {
  /** Field name (or `/regex/` pattern that matches on a field name). Required. */
  field: string;
  /** Human-readable label shown on the derived data link. */
  label: string;
  /** Regex applied to the field value; captured groups are usable in the URL template. Required. */
  matcherRegex: string;
  /** URL template (external) or query text (internal). Supports interpolation like `${__value.raw}`. */
  url: string;
  /** UID of a Grafana data source; when set the data link becomes an internal link. */
  datasourceUid?: string;
};

/**
 * Root (top-level datasource settings) fields the Splunk plugin cares about.
 *
 * `url` is read directly by the backend as `config.URL` (`pkg/models/settings.go:90`) and used to
 * build the Splunk REST endpoints (`pkg/splunk/client.go:67-68`). `basicAuthUser` is written by
 * @grafana/plugin-ui's `Auth` component (`convertLegacyAuthProps` → root `basicAuthUser`) and used
 * by the SDK HTTP client for Basic auth; the plugin's own `LoadSettings` does not read it.
 *
 * Note: unlike most HTTP datasources, the Splunk editor does NOT write `root.basicAuth` — the
 * backend derives `BasicAuthEnabled` from `jsonData.authType` (`pkg/models/settings.go:93-95`).
 */
export type RootConfig = {
  /** Complete HTTP URL of the Splunk management/REST API (default port 8089). Required. */
  url?: string;
  /** Basic-auth username. Only meaningful when the BasicAuth method is selected. */
  basicAuthUser?: string;
};

/**
 * Fields stored in `jsonData`. Union of the plugin's `SplunkOptions` (`src/types.ts:96-119`) that
 * the editor actually reads/writes and the HTTP fields written by @grafana/plugin-ui components,
 * excluding:
 * - the Secure Socks Proxy toggle (`enableSecureSocksProxy`, deliberately excluded from registry entries),
 * - the dead `apiURL` and `username` fields declared on `SplunkOptions` but never read or written
 *   by the current editor or backend (see the entry README).
 */
export type JsonDataConfig = {
  /** Selected authentication method. Empty/missing is treated as `BasicAuth` by the backend. */
  authType?: SplunkAuthType;
  /** True only for the `OAuthForward` method. Written by the editor, consumed by Grafana core (not the plugin backend). */
  oauthPassThru?: boolean;

  /** Enable verification against a self-signed / custom CA certificate. Written by @grafana/plugin-ui's Auth (TLS). */
  tlsAuthWithCACert?: boolean;
  /** Enable TLS client (mutual) authentication. Written by @grafana/plugin-ui's Auth (TLS). */
  tlsAuth?: boolean;
  /** TLS server name used to verify the returned certificate. Only meaningful when `tlsAuth === true`. */
  serverName?: string;
  /** Skip TLS certificate validation (testing only). */
  tlsSkipVerify?: boolean;

  /** Cookies to forward to the datasource, by name. Written by @grafana/plugin-ui's `AdvancedHttpSettings`; read by the backend (`keepCookies`). */
  keepCookies?: string[];
  /** SDK HTTP request timeout in seconds. Written by @grafana/plugin-ui's `AdvancedHttpSettings`; consumed by the SDK, distinct from `timeoutInSeconds`. */
  timeout?: number;

  /** Result `count` limit per request; 0 means unlimited in the editor (backend resolves 0 to 10000). Backend `Count`. */
  maxResultCount?: number;
  /** Display events as soon as they are returned. Backend `PreviewMode`. */
  previewMode?: boolean;
  /** Run the query and periodically poll for results. Backend `AsyncMode` (`json:"pollSearchResult"`). */
  pollSearchResult?: boolean;
  /**
   * Lower bound (ms) of the async poll interval. Declared `number` on `SplunkOptions` but the editor
   * persists a string (`onUpdateDatasourceJsonDataOption`); frontend-only (query builder default).
   */
  minPollInterval?: string;
  /**
   * Upper bound (ms) of the async poll interval. Declared `number` on `SplunkOptions` but the editor
   * persists a string; frontend-only (query builder default).
   */
  maxPollInterval?: string;
  /** Auto-cancel timeout (seconds) applied as a query default; 0 never auto-cancels. Frontend-only. */
  autoCancel?: string;
  /** Plugin-specific request timeout (seconds); minimum 1, default 30. Backend `TimeoutInSeconds`. Distinct from `timeout`. */
  timeoutInSeconds?: number;
  /** Maximum timeline status buckets applied as a query default; 0 disables timelines. Frontend-only. */
  statusBuckets?: string;
  /** Hide fields whose names match `internalFieldPattern`. Backend `InternalFieldsFiltration`. */
  internalFieldsFiltration?: boolean;
  /** Regex for internal fields; backend defaults it to `^_.+` and clears it when filtration is off. Backend `InternalFieldPattern`. */
  internalFieldPattern?: string;
  /** Timestamp field name; backend defaults it to `_time`. Backend `TimeField` (`json:"tsField"`). */
  tsField?: string;
  /** Fields search mode. Backend `FieldSearchType`. */
  fieldSearchType?: FieldSearchType;
  /** Variables search mode. Backend `VariableSearchLevel`. */
  variableSearchLevel?: VariableSearchLevel;
  /** Earliest time for searches without a time range; applied as a query default. Frontend-only. */
  defaultEarliestTime?: string;
  /** Legacy 'stream mode' flag; the backend migrates a true value to `previewMode`. No editor UI. */
  streamMode?: boolean;

  /** Derived data links applied to result fields. Backend compiles each entry's regex. */
  dataLinks?: DataLinkConfig[];
};

/**
 * Secret key names stored in `secureJsonData` (write-only; read existing config via
 * `secureJsonFields`):
 * - `basicAuthPassword` — Basic-auth password; set when the BasicAuth method is selected
 * - `authToken` — Splunk authentication token; set when the `custom-splunk` method is selected
 *   (sent as `Authorization: Bearer`)
 * - `tlsCACert` — custom CA certificate; set when `jsonData.tlsAuthWithCACert` is true
 * - `tlsClientCert` — TLS client certificate; set when `jsonData.tlsAuth` is true
 * - `tlsClientKey` — TLS client key; set when `jsonData.tlsAuth` is true
 *
 * The backend also reads a legacy `APIKey` secret (`pkg/models/settings.go:91`) that no code path
 * consumes and the editor never writes; it is intentionally not modeled here (see the entry README).
 */
export type SecureJsonDataConfig = Array<
  'basicAuthPassword' | 'authToken' | 'tlsCACert' | 'tlsClientCert' | 'tlsClientKey'
>;
