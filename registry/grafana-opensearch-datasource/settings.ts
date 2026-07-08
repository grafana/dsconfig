/**
 * Configuration models for the OpenSearch datasource plugin (`grafana-opensearch-datasource`).
 *
 * Sources of truth (https://github.com/grafana/opensearch-datasource @ 6881bb4218ee9924f8ca6330c5b558b913ab5f19):
 * - `src/plugin.json:1-41` — plugin id ("grafana-opensearch-datasource"), name ("OpenSearch"),
 *   docs URL (info.links[0].url = "https://github.com/grafana/opensearch-datasource")
 * - `src/configuration/ConfigEditor.tsx:1-112` — outer editor; composes @grafana/ui's
 *   `DataSourceHttpSettings` (with `showAccessOptions=true`, `sigV4AuthToggleEnabled` gated on
 *   `config.sigV4AuthEnabled`, and a `renderSigV4Editor` slot filled by @grafana/aws-sdk's
 *   `SIGV4ConnectionConfig`), plus the conditionally-rendered `SecureSocksProxySettings`,
 *   `OpenSearchDetails`, `LogsConfig`, and `DataLinks`.
 * - `src/configuration/OpenSearchDetails.tsx:1-327` — Index name (`jsonData.database`), Pattern
 *   (`jsonData.interval`), Time field name (`jsonData.timeField`), Serverless (`jsonData.serverless`),
 *   Version (`jsonData.version` + `jsonData.flavor` + `jsonData.versionLabel`), Max concurrent
 *   Shard Requests (`jsonData.maxConcurrentShardRequests`), Min time interval (`jsonData.timeInterval`),
 *   PPL enabled (`jsonData.pplEnabled`). The intervalHandler mistakenly writes to root.database
 *   (line 269) — a latent upstream bug preserved in the notes only.
 * - `src/configuration/LogsConfig.tsx:1-48` — log message field / log level field
 * - `src/configuration/DataLinks.tsx:1-84`, `src/configuration/DataLink.tsx:1-159` — dataLinks entries
 *   with fields `field`, `title`, `url`, `datasourceUid`
 * - `src/configuration/utils.ts:6-48` — `coerceOptions` writes editor-parity defaults on mount:
 *   `timeField='@timestamp'`, `maxConcurrentShardRequests` from `defaultMaxConcurrentShardRequests`,
 *   `logMessageField=''`, `logLevelField=''`, `pplEnabled=true`
 * - `src/types.ts:12-28` — `OpenSearchOptions` shape (jsonData), `Flavor` enum, `DataLinkConfig`
 * - `pkg/opensearch/opensearch.go:45-174` — `CheckHealth` reads jsonData.flavor, jsonData.version,
 *   jsonData.timeField, jsonData.database, jsonData.interval; hard-fails on missing flavor/version
 *   or missing timeField.
 * - `pkg/opensearch/client/client.go:30-63` — `NewDatasourceHttpClient` reads jsonData.serverless
 *   and jsonData.oauthPassThru, wires forward-HTTP-headers, forces `SigV4.Service='aoss'` when
 *   serverless is true otherwise 'es'.
 * - `pkg/opensearch/client/client.go:96-153` — `NewClient` reads jsonData.version, jsonData.flavor,
 *   jsonData.timeField, jsonData.logLevelField, jsonData.logMessageField (defaulting to `_source`),
 *   jsonData.database, jsonData.interval.
 * - `pkg/opensearch/client/client.go:288-302,520-534` — `secureJsonData.basicAuthPassword` and
 *   legacy `secureJsonData.password` (used when root.user is set and root.basicAuth is false).
 *
 * External components consulted at their pinned versions:
 * - `@grafana/ui@12.4.2` — `DataSourceHttpSettings`, `BasicAuthSettings`, `HttpProxySettings`,
 *   `TLSAuthSettings`, `CustomHeadersSettings`, `SecureSocksProxySettings` (excluded per
 *   AGENTS.md), `LegacyForms.FormField`/`Input`/`Select`/`Switch`, `Alert`, `Button`, `Divider`,
 *   `TagsInput`, `RadioButtonGroup`, `DataLinkInput`
 * - `@grafana/data@12.4.2` — `DataSourceJsonData` base, `DataSourcePluginOptionsEditorProps`,
 *   `DataSourceSettings`, `DataLinkBuiltInVars`
 * - `@grafana/aws-sdk@0.10.2` — `SIGV4ConnectionConfig` (contributes the sigV4* jsonData fields
 *   plus sigV4AccessKey / sigV4SecretKey secrets when `config.sigV4AuthEnabled` is true)
 * - `@grafana/runtime@12.4.2` — `config.sigV4AuthEnabled`, `config.secureSocksDSProxyEnabled`
 * - `semver@7.7.4` — used by both editor and backend to validate `jsonData.version`
 */

/**
 * OpenSearch flavor. `flavor='opensearch'` uses `_plugins/_ppl` for PPL queries;
 * `flavor='elasticsearch'` uses `_opendistro/_ppl` for compatibility with legacy
 * Elasticsearch clusters running the Open Distro plugins (client.go:557-560).
 */
export type OpenSearchFlavor = 'opensearch' | 'elasticsearch';

/**
 * Time-based index pattern selector. Empty / undefined means "No pattern"
 * (`OpenSearchDetails.tsx:10-17`). Selecting a value triggers `intervalHandler`
 * which auto-fills `root.database` (upstream inconsistency — see README) with a
 * bracketed template like `[logstash-]YYYY.MM.DD`.
 */
export type OpenSearchInterval = 'Hourly' | 'Daily' | 'Weekly' | 'Monthly' | 'Yearly';

/**
 * Datasource root `access` mode. `'proxy'` (Server, default) means requests
 * flow through the Grafana backend; `'direct'` (Browser) means requests are
 * made from the browser and are subject to CORS. OpenSearch supports both
 * because `DataSourceHttpSettings` is instantiated with `showAccessOptions=true`
 * (ConfigEditor.tsx:58), but `'direct'` disables the plugin's Additional-HTTP
 * (`keepCookies`, `timeout`), forward-OAuth, and custom-headers panels
 * (DataSourceHttpSettings.mjs:204,341,357).
 */
export type OpenSearchAccessMode = 'proxy' | 'direct';

/**
 * A single "Data links" entry configured by `DataLinks.tsx` / `DataLink.tsx`
 * (`src/types.ts:101-106`). When `datasourceUid` is set the editor treats
 * `url` as a query for that internal data source; otherwise `url` is an
 * external URL template.
 *
 * NOTE: OpenSearch's DataLinkConfig uses `title` (not `urlDisplayLabel` as in
 * the Elasticsearch plugin) and does not have a `urlDisplayLabel` field.
 */
export type OpenSearchDataLinkConfig = {
  field: string;
  url: string;
  datasourceUid?: string;
  title?: string;
};

/**
 * Root (top-level datasource settings) fields the OpenSearch editor writes.
 *
 * All of these are populated by @grafana/ui's `DataSourceHttpSettings` +
 * `BasicAuthSettings` + `HttpProxySettings` (invoked from `ConfigEditor.tsx:55-62`)
 * and consumed by the SDK's `settings.HTTPClientOptions(ctx)` call in
 * `pkg/opensearch/client/client.go:41` when building the HTTP client. The plugin's
 * own Go code reads `settings.URL` (opensearch.go:99, client.go:245,479) directly;
 * `settings.User` (legacy) and `settings.BasicAuthEnabled` / `settings.BasicAuthUser`
 * are consumed at client.go:288-298 and client.go:520-530 when setting per-request
 * basic-auth credentials.
 */
export type RootConfig = {
  /** Complete HTTP URL of the OpenSearch node/cluster. */
  url?: string;
  /**
   * `'proxy'` (Server, default) or `'direct'` (Browser). The OpenSearch editor
   * exposes this as a radio group (`DataSourceHttpSettings` prop
   * `showAccessOptions=true`).
   */
  access?: OpenSearchAccessMode;
  /** True when HTTP Basic authentication is enabled. Toggled by the editor's Basic auth switch. */
  basicAuth?: boolean;
  /** Basic-auth username. Only meaningful when `basicAuth === true`. */
  basicAuthUser?: string;
  /**
   * Cross-site cookie/header forwarding toggle for browser mode. Consumed
   * only by the frontend browser fetch — the backend HTTP client never reads
   * it. Independent from `basicAuth`.
   */
  withCredentials?: boolean;
  /**
   * Legacy root-level user field. Read by the backend at
   * `client.go:294-298,526-530` when `basicAuth` is false, pairing with
   * `secureJsonData.password` to send Basic-auth credentials. Never written by
   * the current editor.
   */
  user?: string;
  /**
   * Legacy root-level database. Reset to '' by the editor's
   * `intervalHandler` (OpenSearchDetails.tsx:256-283) but never read by the
   * backend, which pulls the index from `jsonData.database` instead
   * (opensearch.go:77, client.go:120).
   */
  database?: string;
};

/**
 * Fields stored in `jsonData`. Combines the plugin's `OpenSearchOptions`
 * (`src/types.ts:12-28`) with the TLS/HTTP fields written by @grafana/ui's
 * `DataSourceHttpSettings` and the SigV4 fields contributed by @grafana/aws-sdk
 * when `config.sigV4AuthEnabled` is true.
 */
export type JsonDataConfig = {
  // --- Auth toggles (each is an independent boolean, no discriminator) ---
  /** True when the Forward OAuth Identity toggle is on (HttpProxySettings.mjs:75-89). Only shown when access='proxy'. */
  oauthPassThru?: boolean;
  /** True when SigV4 auth is enabled. Only shown by the editor when config.sigV4AuthEnabled is true. */
  sigV4Auth?: boolean;

  // --- TLS (written by @grafana/ui's HttpProxySettings + TLSAuthSettings) ---
  /** Enable TLS client authentication (mTLS). Requires `serverName`, `tlsClientCert`, `tlsClientKey`. */
  tlsAuth?: boolean;
  /** Enable custom CA verification for self-signed certs. Requires `tlsCACert`. */
  tlsAuthWithCACert?: boolean;
  /** Skip TLS certificate validation. Not recommended outside testing. */
  tlsSkipVerify?: boolean;
  /** TLS SNI / cert-verification server name (`TLSAuthSettings.mjs:80-86`). */
  serverName?: string;

  // --- HTTP (written by @grafana/ui's DataSourceHttpSettings; only when access='proxy') ---
  /** HTTP request timeout in seconds (`DataSourceHttpSettings.mjs:238-244`). */
  timeout?: number;
  /** Cookies to forward to the datasource, by name (`DataSourceHttpSettings.mjs:213-221`). */
  keepCookies?: string[];

  // --- OpenSearch-specific storage (OpenSearchDetails.tsx / LogsConfig.tsx) ---
  /** Concrete index name or index prefix. May contain a wildcard or time template. Editor marks it as required. */
  database?: string;
  /** Time-pattern selector; undefined = "No pattern". */
  interval?: OpenSearchInterval;
  /** Name of the timestamp field (required; defaults to `'@timestamp'` via `coerceOptions`). */
  timeField?: string;
  /** True when the datasource targets AWS OpenSearch Serverless — enables the 'aoss' SigV4 service namespace and adds x-amz-content-sha256 headers. */
  serverless?: boolean;
  /** OpenSearch flavor discriminator. Set by the 'Get Version and Save' button; required by the backend health check. */
  flavor?: OpenSearchFlavor;
  /** Semantic version string (semver) reported by the cluster. Required by the backend. */
  version?: string;
  /** Human-readable version label rendered in the disabled Version input. Editor-only. */
  versionLabel?: string;
  /** Maximum concurrent shard requests per node. Defaults: 5 (OpenSearch or Elasticsearch >=7.0.0), 256 (Elasticsearch <7.0.0). */
  maxConcurrentShardRequests?: number;
  /** Lower bound for the auto group-by time interval; free-form duration string (`10s`, `1m`, …). */
  timeInterval?: string;
  /** Enable Piped Processing Language as an alternative query syntax. Defaults to true. */
  pplEnabled?: boolean;

  // --- Logs sub-section ---
  /** Field to render as the log message. Backend defaults to `_source` when empty (client.go:118). */
  logMessageField?: string;
  /** Field determining the log level of each row. */
  logLevelField?: string;

  // --- Data links ---
  /** Array of link definitions rendered next to matching fields in log row details. Frontend-only. */
  dataLinks?: OpenSearchDataLinkConfig[];
};

/**
 * Secret key names stored in `secureJsonData` (write-only; read existing config
 * via `secureJsonFields`):
 * - `basicAuthPassword` — user password when Basic auth is enabled.
 * - `tlsCACert` — custom CA PEM when `tlsAuthWithCACert` is true.
 * - `tlsClientCert`, `tlsClientKey` — mTLS client credentials when `tlsAuth` is true.
 *
 * The editor also writes:
 * - Dynamic `httpHeaderValue<N>` secrets via @grafana/ui's `CustomHeadersSettings`.
 * - `sigV4AccessKey` / `sigV4SecretKey` when the SigV4 auth method is selected
 *   (contributed by @grafana/aws-sdk's `SIGV4ConnectionConfig`, not by this plugin).
 * The legacy `password` secret is read by the backend when `root.user` is set
 * and `root.basicAuth` is false (client.go:294-298); no current editor UI writes it.
 * None of those categories is modeled as a first-class secret here.
 */
export type SecureJsonDataConfig = Array<
  'basicAuthPassword' | 'tlsCACert' | 'tlsClientCert' | 'tlsClientKey'
>;
