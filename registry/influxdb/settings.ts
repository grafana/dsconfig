/**
 * Configuration models for the InfluxDB datasource plugin (`influxdb`).
 *
 * Sources of truth (https://github.com/grafana/grafana-influxdb-datasource @ a3e5fe3):
 * - `src/plugin.json:1-40` — plugin id (`"influxdb"`), name (`"InfluxDB"`),
 *   docs URL (`info.links[1].url` = `"https://grafana.com/docs/grafana/latest/datasources/influxdb/"`)
 * - `src/module.ts:10-14` — chooses between `ConfigEditorV1` (default) and
 *   `ConfigEditorV2` (feature toggle `newInfluxDSConfigPageDesign`) as the
 *   registered config editor. Both editors write into the same jsonData +
 *   secureJsonData shape — this schema captures the union of what either
 *   writes to storage.
 * - `src/types.ts:5-47` — `InfluxVersion` enum ('InfluxQL' | 'Flux' | 'SQL'),
 *   `InfluxOptions extends DataSourceJsonData` (version, timeInterval, httpMode,
 *   showTagTime, dbName, product, pdcInjected, oauthPassThru, organization,
 *   defaultBucket, maxSeries, metadata, insecureGrpc), deprecated
 *   `InfluxOptionsV1` (user, database), `InfluxSecureJsonData` (token, password)
 * - `src/components/editor/config/ConfigEditor.tsx:20-142` — v1 outer editor:
 *   query-language Select + `DataSourceHttpSettings` from @grafana/ui +
 *   InfluxDB Details FieldSet composing InfluxInfluxQLConfig /
 *   InfluxFluxConfig / InfluxSqlConfig + Max series input. Also enforces
 *   proxy access and forces basicAuth=true when Flux is selected.
 * - `src/components/editor/config/InfluxInfluxQLConfig.tsx:32-178` — v1
 *   InfluxQL tab: Database (jsonData.dbName), User (root.user), Password
 *   (secureJsonData.password), HTTP Method (jsonData.httpMode; GET/POST), Min
 *   time interval (jsonData.timeInterval), Autocomplete range
 *   (jsonData.showTagTime)
 * - `src/components/editor/config/InfluxFluxConfig.tsx:22-84` — v1 Flux tab:
 *   Organization (jsonData.organization), Token (secureJsonData.token),
 *   Default Bucket (jsonData.defaultBucket), Min time interval
 *   (jsonData.timeInterval)
 * - `src/components/editor/config/InfluxSQLConfig.tsx:19-90` — v1 SQL tab:
 *   Database (jsonData.dbName), Token (secureJsonData.token), Insecure
 *   Connection (jsonData.insecureGrpc)
 * - `src/components/editor/config-v2/*` — v2 editor (feature-flagged):
 *   `UrlAndAuthenticationSection.tsx`, `AuthSettings.tsx`,
 *   `DatabaseConnectionSection.tsx`, `AdvancedHttpSettings.tsx`,
 *   `AdvancedDBConnectionSettings.tsx`. Introduces jsonData.product +
 *   jsonData.pdcInjected + jsonData.timeout + jsonData.keepCookies and swaps
 *   the v1 root.user/secureJsonData.password pair for root.basicAuthUser +
 *   secureJsonData.basicAuthPassword under BasicAuth radio selection.
 * - `src/datasource.ts:48-102` — `InfluxDatasource` constructor: reads
 *   settings.url, settings.username, settings.password, settings.basicAuth,
 *   settings.withCredentials, settings.access, jsonData.dbName (fallback
 *   settings.database), jsonData.timeInterval, jsonData.showTagTime,
 *   jsonData.httpMode (default 'GET'), jsonData.version (default 'InfluxQL')
 * - `pkg/influxdb/influxdb.go:26-87` — backend `NewDatasource`:
 *   `settings.HTTPClientOptions(ctx)` for the HTTP client, then unmarshals
 *   `settings.JSONData` into `DatasourceInfo` (dbName, version, httpMode,
 *   timeInterval, defaultBucket, organization, maxSeries, insecureGrpc);
 *   falls back to settings.Database (root) when jsonData.dbName is empty;
 *   defaults httpMode='GET', maxSeries=1000, version='InfluxQL'
 * - `pkg/influxdb/models/datasource_info.go:11-32` — backend jsonData shape
 * - `pkg/influxdb/influxql/influxql.go:140-194` — createRequest reads
 *   dsInfo.URL, dsInfo.DbName (writes ?db=…), dsInfo.HTTPMode
 *
 * External components consulted at their pinned versions:
 * - `@grafana/ui@13.1.0` — `DataSourceHttpSettings` (URL input, Access help,
 *   Allowed cookies TagsInput, Timeout input, Basic auth switch, With
 *   Credentials switch), `BasicAuthSettings`, `HttpProxySettings` (TLS Client
 *   Auth, With CA Cert, Skip TLS Verify, Forward OAuth Identity switches),
 *   `TLSAuthSettings` + `CertificationKey` (ServerName, CA Cert, Client Cert,
 *   Client Key textareas), `CustomHeadersSettings` (dynamic httpHeaderName<N>
 *   / secureJsonData httpHeaderValue<N> — NOT modeled here),
 *   `SecureSocksProxySettings` (rendered when `config.secureSocksDSProxyEnabled`
 *   — excluded per AGENTS.md), plus `Field`, `FieldSet`, `Select`, `Input`,
 *   `SecretInput`, `InlineField`, `InlineLabel`, `InlineSwitch`, `Combobox`,
 *   `Alert`, `Box`, `Stack`, `Text`, `TextLink`, `TagsInput`, `Checkbox`,
 *   `Button`, `Space`
 * - `@grafana/plugin-ui@0.13.1` — `AuthMethod` enum, `convertLegacyAuthProps`
 *   (used ONLY by v2 editor's `AuthSettings.tsx` — v1 editor uses @grafana/ui's
 *   `DataSourceHttpSettings` directly)
 * - `@grafana/data@13.1.0` — `DataSourceJsonData` base interface,
 *   `DataSourcePluginOptionsEditorProps`, `updateDatasourcePluginJsonDataOption`,
 *   `onUpdateDatasourceOption`, `onUpdateDatasourceJsonDataOption`,
 *   `onUpdateDatasourceJsonDataOptionSelect`,
 *   `onUpdateDatasourceSecureJsonDataOption`,
 *   `updateDatasourcePluginResetOption`
 * - `@grafana/runtime@13.1.0` — `config.secureSocksDSProxyEnabled`,
 *   `config.featureToggles.newInfluxDSConfigPageDesign`,
 *   `config.featureToggles.influxDBConfigValidation`
 */

/**
 * Query-language selector — stored as a JSON string under `jsonData.version`.
 * Value set is `InfluxVersion` from `src/types.ts:5-9` (Note the enum values
 * match the labels character-for-character in the v1 editor).
 */
export type InfluxVersion = 'InfluxQL' | 'Flux' | 'SQL';

/**
 * HTTP verb used by the InfluxQL query path (`jsonData.httpMode`). Options
 * from `InfluxInfluxQLConfig.tsx:25-28`; backend default 'GET'
 * (`pkg/influxdb/influxdb.go:43-46`).
 */
export type InfluxHTTPMode = 'GET' | 'POST';

/**
 * InfluxDB product/edition tag written by the v2 editor's Product Combobox
 * (`UrlAndAuthenticationSection.tsx:252-257`; option values from
 * `versions.ts:24-155`). Not present in v1-authored datasources.
 */
export type InfluxProduct =
  | 'InfluxDB Cloud Dedicated'
  | 'InfluxDB Cloud Serverless'
  | 'InfluxDB Clustered'
  | 'InfluxDB Enterprise 1.x'
  | 'InfluxDB Enterprise 3.x'
  | 'InfluxDB Cloud (TSM)'
  | 'InfluxDB Cloud 1'
  | 'InfluxDB OSS 1.x'
  | 'InfluxDB OSS 2.x'
  | 'InfluxDB OSS 3.x';

/**
 * Root (top-level datasource settings) fields either editor writes.
 *
 * The v1 InfluxQL editor writes `user` (root) + `secureJsonData.password`
 * (`InfluxInfluxQLConfig.tsx:87-108`) — legacy 1.x direct-DB auth. The v2
 * editor's "Basic authentication" radio (AuthSettings.tsx:164-179) writes
 * `basicAuthUser` + `secureJsonData.basicAuthPassword` instead. `basicAuth`,
 * `withCredentials`, and `access` are all populated by @grafana/ui's
 * `DataSourceHttpSettings`. The plugin's Go backend reads only
 * `settings.URL` directly (`pkg/influxdb/influxdb.go:72`); the rest is
 * honored by the SDK's transport builder.
 */
export type RootConfig = {
  /** Complete HTTP URL of the InfluxDB HTTP API. v1 defaultUrl `http://localhost:8086`. */
  url?: string;
  /**
   * `'proxy'` (Server, default) or `'direct'` (Browser). The v1 editor
   * renders an inline error Alert and datasource.query rejects with
   * `BROWSER_MODE_DISABLED_MESSAGE` when access==='direct'
   * (`datasource.ts:105-108`). Always set to `'proxy'` in practice; kept
   * for legacy provisioning payloads.
   */
  access?: 'proxy' | 'direct';
  /** True when HTTP Basic authentication is enabled. Forced true by `onVersionChanged` when Flux is selected (ConfigEditor.tsx:63). */
  basicAuth?: boolean;
  /** Basic-auth username (v2 editor / DataSourceHttpSettings top-level). Distinct from `user`. */
  basicAuthUser?: string;
  /**
   * Legacy database-user field written by the v1 InfluxQL editor
   * (`InfluxInfluxQLConfig.tsx:87-90`). Distinct from `basicAuthUser`.
   * Not consumed by the current backend or SDK HTTP auth path.
   */
  user?: string;
  /**
   * Legacy database name — v1 InfluxQL editor blanks this on any dbName
   * edit (`InfluxInfluxQLConfig.tsx:64-72`). Backend falls back to it if
   * `jsonData.dbName` is empty (`pkg/influxdb/influxdb.go:58-61`).
   */
  database?: string;
  /**
   * Cross-site access-control toggle rendered as "With Credentials" by
   * DataSourceHttpSettings. Independent from `basicAuth` — both can be true.
   */
  withCredentials?: boolean;
};

/**
 * Fields stored in `jsonData`. Union of `InfluxOptions` (`src/types.ts:11-31`)
 * plus TLS/HTTP fields written by @grafana/ui's `DataSourceHttpSettings`,
 * `HttpProxySettings`, and `TLSAuthSettings` — and the v2 editor's
 * `AdvancedHttpSettings` (timeout, keepCookies) and product-detection
 * outputs (`product`, `pdcInjected`).
 *
 * The plugin's Go backend unmarshals only the InfluxDB-specific fields
 * (`pkg/influxdb/models/datasource_info.go:17-27`); TLS + cookie + timeout
 * knobs are read by the SDK's `HTTPClientOptions` when building the HTTP
 * client.
 */
export type JsonDataConfig = {
  /** Enable TLS client authentication (mTLS). Requires `serverName` + `tlsClientCert` + `tlsClientKey`. */
  tlsAuth?: boolean;
  /** Enable custom CA verification. Requires `tlsCACert`. */
  tlsAuthWithCACert?: boolean;
  /** Skip TLS certificate validation. Not recommended outside testing. */
  tlsSkipVerify?: boolean;
  /** TLS SNI / cert-verification server name (@grafana/ui TLSAuthSettings). */
  serverName?: string;
  /** HTTP request timeout in seconds (v2 editor `AdvancedHttpSettings.tsx:74-100`; also possible via @grafana/ui DataSourceHttpSettings). */
  timeout?: number;
  /** Cookies to forward to the datasource, by name (v2 `AdvancedHttpSettings.tsx:44-68`; also from @grafana/ui DataSourceHttpSettings). */
  keepCookies?: string[];
  /**
   * Forward OAuth Identity — written by the "Forward OAuth Identity" switch
   * inside `HttpProxySettings` (v1) or the v2 AuthSettings radio
   * (AuthSettings.tsx:110-114). When true, the SDK forwards the signed-in
   * user's OAuth identity to InfluxDB.
   */
  oauthPassThru?: boolean;
  /**
   * Query-language discriminator. Selects the backend query path
   * (`pkg/influxdb/influxdb.go:97-107`). Backend default `'InfluxQL'`
   * (`pkg/influxdb/influxdb.go:53-56`).
   */
  version?: InfluxVersion;
  /**
   * InfluxDB product/edition — written by the v2 editor only
   * (`UrlAndAuthenticationSection.tsx:91-94`). Determines which query
   * languages are available in the v2 UI (`versions.ts`). Not read by the
   * backend.
   */
  product?: InfluxProduct;
  /**
   * PDC-injected indicator — set by the Grafana backend when a Private
   * Datasource Connect proxy has been injected into this datasource. Read
   * only by the v2 editor's `LeftSideBar` (`LeftSideBar.tsx:12`) to render
   * PDC-specific section headers. Not editor-writable.
   */
  pdcInjected?: boolean;
  /**
   * Database name (InfluxQL: 1.x database, SQL v3: database). Not used by
   * the Flux query path. Backend falls back to root.database when empty
   * (`pkg/influxdb/influxdb.go:58-61`).
   */
  dbName?: string;
  /** HTTP verb for InfluxQL queries; GET (default) or POST. */
  httpMode?: InfluxHTTPMode;
  /** Minimum auto-group-by interval (InfluxQL / Flux). Duration string like `'10s'`. */
  timeInterval?: string;
  /**
   * Autocomplete range for InfluxQL tag-filter query hints
   * (`InfluxInfluxQLConfig.tsx:159-175`; consumed as `withTimeFilter` in
   * `src/datasource.ts:447`). Duration string like `'12h'`.
   */
  showTagTime?: string;
  /** Flux Organization. */
  organization?: string;
  /** Flux default bucket. */
  defaultBucket?: string;
  /** SQL / FlightSQL: disable TLS for the gRPC transport. */
  insecureGrpc?: boolean;
  /**
   * Series/tables cap the plugin applies to result frames. Default 1000
   * (backend fallback at `pkg/influxdb/influxdb.go:48-51`; v1 editor
   * placeholder `1000` at `ConfigEditor.tsx:129`).
   */
  maxSeries?: number;
};

/**
 * Secret key names stored in `secureJsonData` (write-only; read existing
 * config via `secureJsonFields`):
 * - `token` — Flux / SQL bearer token (`InfluxFluxConfig.tsx:42-53`,
 *   `InfluxSQLConfig.tsx:50-61`, v2 `InfluxFluxDBConnection.tsx:91-100` +
 *   `InfluxSQLDBConnection.tsx:59-68`).
 * - `password` — legacy v1 InfluxQL database password paired with
 *   `root.user` (`InfluxInfluxQLConfig.tsx:98-107`). Not written by the v2
 *   editor.
 * - `basicAuthPassword` — HTTP Basic password paired with
 *   `root.basicAuthUser`, written by the v2 editor's Basic auth radio
 *   (`AuthSettings.tsx:172-178`) or the v1 DataSourceHttpSettings top-level
 *   Basic auth toggle.
 * - `tlsCACert`, `tlsClientCert`, `tlsClientKey` — TLS mutual-auth / custom
 *   CA PEMs written by @grafana/ui `TLSAuthSettings`.
 *
 * The v2 editor also writes dynamic `httpHeaderValue<N>` secrets when the
 * user configures custom HTTP headers via @grafana/ui's
 * `CustomHeadersSettings` (`AdvancedHttpSettings.tsx:102`). Those keys are
 * indexed pairs — not modeled as first-class fields in this schema; see the
 * README.
 */
export type SecureJsonDataConfig = Array<
  'token' | 'password' | 'basicAuthPassword' | 'tlsCACert' | 'tlsClientCert' | 'tlsClientKey'
>;
