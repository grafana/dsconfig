/**
 * Configuration models for the Elasticsearch datasource plugin (`elasticsearch`).
 *
 * Sources of truth (https://github.com/grafana/grafana-elasticsearch-datasource @ 265a51e):
 * - `src/plugin.json:1-53` — plugin id ("elasticsearch"), name ("Elasticsearch"),
 *   docs URL (info.links[2].url = "https://grafana.com/docs/grafana/latest/datasources/elasticsearch/")
 * - `src/configuration/ConfigEditor.tsx` — outer editor; composes @grafana/plugin-ui's
 *   `ConnectionSettings`, `Auth` (with custom-api-key and optional custom-sigv4 methods),
 *   `AdvancedHttpSettings`, plus plugin-owned `ElasticDetails`, `LogsConfig`, and `DataLinks`.
 * - `src/configuration/ElasticDetails.tsx` — index name, pattern, time field, max shard
 *   requests, min time interval, include frozen, default query mode
 * - `src/configuration/LogsConfig.tsx` — log message field, log level field
 * - `src/configuration/DataLinks.tsx`, `src/configuration/DataLink.tsx` — dataLinks entries
 * - `src/configuration/ApiKeyConfig.tsx` — custom API Key auth method component
 * - `src/configuration/utils.ts` — `coerceOptions` writes editor-parity defaults on mount:
 *   `timeField='@timestamp'`, `maxConcurrentShardRequests=5`, `logMessageField=''`,
 *   `logLevelField=''`, `includeFrozen=false`, `defaultQueryMode='metrics'`
 * - `src/types.ts:60-82` — `ElasticsearchOptions`, `Interval`, `QueryType`,
 *   `ElasticsearchSecureJsonData`, `DataLinkConfig`
 * - `pkg/elasticsearch/elasticsearch.go:106-243` — `NewDatasource` reads jsonData
 *   verbatim as a `map[string]any` (no plugin-owned settings struct), pulls
 *   apiKey from decrypted secrets, hard-fails on empty timeField, defaults
 *   maxConcurrentShardRequests to 5, falls back to `settings.Database` when
 *   jsonData.index is empty.
 *
 * External components consulted at their pinned versions:
 * - `@grafana/plugin-ui@0.13.1` — `ConnectionSettings` (URL input, placeholder
 *   overridden to "http://localhost:9200"), `Auth` (`AuthMethodSettings`,
 *   `BasicAuth`, `TLSSettings`/`SelfSignedCertificate`/`TLSClientAuth`/
 *   `SkipTLSVerification`, `CustomHeaders`), `AdvancedHttpSettings` (Allowed
 *   cookies TagsInput + Timeout input), `DataSourceDescription`, `ConfigSection`,
 *   `ConfigSubSection`, `ConfigDescriptionLink`, `convertLegacyAuthProps`
 * - `@grafana/ui@13.1.0` — `Alert`, `Divider`, `SecureSocksProxySettings`
 *   (rendered conditionally, excluded per AGENTS.md), `Input`, `InlineField`,
 *   `InlineSwitch`, `Select`, `SecretInput`, `Button`, `DataLinkInput`
 * - `@grafana/data@13.1.0` — `DataSourceJsonData` base interface,
 *   `DataSourcePluginOptionsEditorProps`, `updateDatasourcePluginResetOption`,
 *   `onUpdateDatasourceSecureJsonDataOption`
 * - `@grafana/aws-sdk@0.10.0` — `SIGV4ConnectionConfig` (contributes the sigV4*
 *   jsonData fields when the Grafana instance has `config.sigV4AuthEnabled`)
 * - `@grafana/runtime@13.1.0` — `config.sigV4AuthEnabled`, `config.secureSocksDSProxyEnabled`
 */

/** Time-based index pattern selector; empty / undefined means "No pattern" (`src/types.ts:60`). */
export type ElasticsearchInterval = 'Hourly' | 'Daily' | 'Weekly' | 'Monthly' | 'Yearly';

/** Default query mode for the query editor (`src/types.ts:84`). Defaults to `'metrics'`. */
export type ElasticsearchQueryType = 'metrics' | 'logs' | 'raw_data' | 'raw_document';

/**
 * Editor-local authentication method identifier. Not a stored field — it is
 * derived by @grafana/plugin-ui's `convertLegacyAuthProps` from
 * `root.basicAuth`, `root.withCredentials`, `jsonData.oauthPassThru`,
 * `jsonData.apiKeyAuth`, and `jsonData.sigV4Auth` (`ConfigEditor.tsx:48-61`).
 */
export type ElasticsearchAuthMethod =
  | 'NoAuth'
  | 'BasicAuth'
  | 'OAuthForward'
  | 'CrossSiteCredentials'
  | 'custom-api-key'
  | 'custom-sigv4';

/**
 * A single "Data links" entry configured by `DataLinks.tsx` / `DataLink.tsx`
 * (`src/types.ts:135-140`). When `datasourceUid` is set the editor treats
 * `url` as a query for that internal data source; otherwise `url` is an
 * external URL template.
 */
export type ElasticsearchDataLinkConfig = {
  field: string;
  url: string;
  urlDisplayLabel?: string;
  datasourceUid?: string;
};

/**
 * Root (top-level datasource settings) fields the Elasticsearch editor writes.
 *
 * `url`, `basicAuth`, `basicAuthUser`, and `withCredentials` are populated by
 * @grafana/plugin-ui's `ConnectionSettings` + `Auth` components and consumed
 * by the SDK's `settings.HTTPClientOptions(ctx)` call in
 * `pkg/elasticsearch/elasticsearch.go:112` when building the HTTP client.
 * The plugin's own Go code reads `settings.URL` (elasticsearch.go:196, 213)
 * and — for legacy datasources — `settings.Database` (elasticsearch.go:169)
 * directly. `access` is legacy: the editor renders a persistent error Alert
 * whenever `options.access === 'direct'` (ConfigEditor.tsx:65-69) because
 * browser mode is no longer supported.
 */
export type RootConfig = {
  /** Complete HTTP URL of the Elasticsearch node/cluster. */
  url?: string;
  /**
   * `'proxy'` (Server, default) or `'direct'` (Browser). The Elasticsearch
   * editor does not offer an Access control, so new datasources always end
   * up with `access: 'proxy'`. Legacy `'direct'` datasources trigger the
   * "Browser access mode ... is no longer available" error banner.
   */
  access?: 'proxy' | 'direct';
  /** True when HTTP Basic authentication is enabled. Written by the editor's onAuthMethodSelect. */
  basicAuth?: boolean;
  /** Basic-auth username. Only meaningful when `basicAuth === true`. */
  basicAuthUser?: string;
  /**
   * Legacy cross-site access-control toggle. @grafana/plugin-ui's
   * `AuthMethodSettings` default `visibleMethods` does not include
   * `CrossSiteCredentials`, so the Elasticsearch editor never writes this
   * field; provisioned datasources that carry `withCredentials: true` from
   * older Grafana versions display as "No Authentication" in the picker.
   */
  withCredentials?: boolean;
  /**
   * Legacy — pre-jsonData storage location for the index name. The
   * backend falls back to `settings.Database` when `jsonData.index` is empty
   * (`elasticsearch.go:164-170`); the editor's `indexChangeHandler` always
   * clears `database` to an empty string when the user edits the index name
   * (`ElasticDetails.tsx:162`).
   */
  database?: string;
};

/**
 * Fields stored in `jsonData`. Combines the plugin's `ElasticsearchOptions`
 * (`src/types.ts:62-78`) with the TLS/HTTP fields written by @grafana/plugin-ui,
 * and the SigV4 fields contributed by @grafana/aws-sdk when
 * `config.sigV4AuthEnabled` is true.
 */
export type JsonDataConfig = {
  // --- Auth discriminators (managed by the virtual authMethod selector) ---
  /** True when the "Forward OAuth Identity" auth method is selected. */
  oauthPassThru?: boolean;
  /** True when the "API Key" custom auth method is selected; consumed by the backend to add `Authorization: ApiKey <key>` (elasticsearch.go:125-131). */
  apiKeyAuth?: boolean;
  /** True when the "SigV4 auth" custom method is selected. Only offered by the editor when `config.sigV4AuthEnabled` is true (ConfigEditor.tsx:50-61). */
  sigV4Auth?: boolean;

  // --- TLS (written by @grafana/plugin-ui Auth/TLSSettings) ---
  /** Enable TLS client authentication (mTLS). Requires `serverName`, `tlsClientCert`, `tlsClientKey`. */
  tlsAuth?: boolean;
  /** Enable custom CA verification. Requires `tlsCACert`. */
  tlsAuthWithCACert?: boolean;
  /** Skip TLS certificate validation. Not recommended outside testing. */
  tlsSkipVerify?: boolean;
  /** TLS SNI / cert-verification server name (`TLSClientAuth.tsx:37`). */
  serverName?: string;

  // --- HTTP (written by @grafana/plugin-ui AdvancedHttpSettings) ---
  /** HTTP request timeout in seconds (`AdvancedHttpSettings.tsx:60-71`). */
  timeout?: number;
  /** Cookies to forward to the datasource, by name (`AdvancedHttpSettings.tsx:38-52`). */
  keepCookies?: string[];

  // --- Elasticsearch-specific storage (ElasticDetails.tsx / LogsConfig.tsx) ---
  /** Concrete index name or index prefix (required at load time). May contain a wildcard or time template. */
  index?: string;
  /** Time-pattern selector; undefined = "No pattern". */
  interval?: ElasticsearchInterval;
  /** Name of the timestamp field (required; defaults to `'@timestamp'` via `coerceOptions`). */
  timeField?: string;
  /** Maximum concurrent shard requests per node; defaults to 5 (editor `coerceOptions` + backend `elasticsearch.go:187-189`). */
  maxConcurrentShardRequests?: number;
  /** Lower bound for the auto group-by time interval; free-form duration string (`10s`, `1m`, …). */
  timeInterval?: string;
  /** Include frozen indices in searches (adds `ignore_throttled=false` to the msearch query string). */
  includeFrozen?: boolean;
  /** Default query mode surfaced in the query editor. */
  defaultQueryMode?: ElasticsearchQueryType;

  // --- Logs sub-section ---
  /** Field to render as the log message. */
  logMessageField?: string;
  /** Field determining the log level of each row. */
  logLevelField?: string;

  // --- Data links ---
  /** Array of link definitions rendered next to matching fields in log row details. */
  dataLinks?: ElasticsearchDataLinkConfig[];
};

/**
 * Secret key names stored in `secureJsonData` (write-only; read existing config
 * via `secureJsonFields`):
 * - `basicAuthPassword` — user password when Basic auth is enabled.
 * - `apiKey` — Elasticsearch API key (sent as `Authorization: ApiKey <value>`).
 * - `tlsCACert` — custom CA PEM when `tlsAuthWithCACert` is true.
 * - `tlsClientCert`, `tlsClientKey` — mTLS client credentials when `tlsAuth` is true.
 *
 * The editor also writes:
 * - Dynamic `httpHeaderValue<N>` secrets via @grafana/plugin-ui's `CustomHeaders`.
 * - `sigV4AccessKey` / `sigV4SecretKey` when the SigV4 auth method is selected
 *   (contributed by @grafana/aws-sdk's `SIGV4ConnectionConfig`, not by this plugin).
 * Neither category is modeled as a first-class secret here.
 */
export type SecureJsonDataConfig = Array<
  'basicAuthPassword' | 'apiKey' | 'tlsCACert' | 'tlsClientCert' | 'tlsClientKey'
>;
