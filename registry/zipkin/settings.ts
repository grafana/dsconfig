/**
 * Configuration models for the Zipkin datasource plugin (`zipkin`).
 *
 * Sources of truth (https://github.com/grafana/grafana-zipkin-datasource @ 1982b15):
 * - `src/plugin.json` — plugin id (`"zipkin"`), name (`"Zipkin"`), docs URL
 *   (`info.links[2].url = "https://grafana.com/docs/grafana/latest/datasources/zipkin/"`)
 * - `src/ConfigEditor.tsx:19-68` — outermost editor (composes
 *   `DataSourceDescription` with `dataSourceName="Zipkin"`,
 *   `ConnectionSettings` with `urlPlaceholder="http://localhost:9411"`,
 *   `Auth` (via `convertLegacyAuthProps`), `TraceToLogsSection`,
 *   `TraceToMetricsSection`, and a collapsible "Additional settings"
 *   section containing `AdvancedHttpSettings`, `SecureSocksProxySettings`
 *   (excluded), `NodeGraphSection`, and `SpanBarSection`)
 * - `src/datasource.ts:21-35, 46-60` — frontend `ZipkinJsonData` extends
 *   `DataSourceJsonData` with `nodeGraph?: NodeGraphOptions`; the constructor
 *   pulls `instanceSettings.jsonData.nodeGraph` and the query pipeline uses
 *   `nodeGraph?.enabled` to attach node-graph frames on the client side
 * - `src/types.ts:1-34` — frontend `ZipkinQuery` (query-level, not config)
 * - `pkg/zipkin/zipkin.go:22-44` — backend `NewDatasource` reads
 *   `settings.URL` directly, calls `settings.HTTPClientOptions(ctx)`, and
 *   builds an HTTP client from those options; it never unmarshals
 *   `settings.JSONData`
 * - `pkg/zipkin/zipkin.go:46-65` — `CheckHealth` calls `Services` (HTTP)
 * - `pkg/zipkin/client.go:16-197` — HTTP client honours `settings.URL` and
 *   hits `/api/v2/services`, `/api/v2/spans`, `/api/v2/traces`, and
 *   `/api/v2/trace/{traceId}`
 * - `pkg/zipkin/handler_querydata.go:14-81` — jsonData is not consumed on
 *   the backend query path either; `query.QueryType` and `query.Query` are
 *   the only per-request inputs
 * - `pkg/zipkin/handler_callresource.go:11-18` — resource routes
 *   `/services`, `/spans`, `/traces`, `/trace/{traceId}`
 *
 * External components consulted at their pinned versions:
 * - `@grafana/plugin-ui@0.13.1` — `ConnectionSettings` (URL input; default
 *   `urlLabel="URL"`, placeholder from plugin), `Auth` +
 *   `convertLegacyAuthProps` (visible methods `[BasicAuth, OAuthForward, NoAuth]`;
 *   option labels `Basic authentication` / `Forward OAuth Identity` /
 *   `No Authentication`; description "Choose an authentication method to
 *   access the data source"), `AdvancedHttpSettings` (Allowed cookies,
 *   Timeout), `DataSourceDescription`, `ConfigSection`
 * - `@grafana/o11y-ds-frontend@13.0.1` — `TraceToLogsSection` (writes
 *   `jsonData.tracesToLogsV2` v2 shape and clears legacy `jsonData.tracesToLogs`),
 *   `TraceToMetricsSection` (writes `jsonData.tracesToMetrics`),
 *   `NodeGraphSection` (writes `jsonData.nodeGraph.enabled`), `SpanBarSection`
 *   (writes `jsonData.spanBar.{type, tag}`)
 * - `@grafana/ui@13.0.1` — `Divider`, `Stack`, `SecureSocksProxySettings`
 *   (excluded)
 * - `@grafana/data@13.0.1` — `DataSourceJsonData`,
 *   `DataSourcePluginOptionsEditorProps`
 * - `@grafana/runtime@13.0.1` — `config` (reads
 *   `config.secureSocksDSProxyEnabled` to render the excluded proxy widget)
 */

/** v2 shape written by `TraceToLogsSection`. */
export type TraceToLogsV2Config = {
  /** Logs datasource UID (loki, elasticsearch, grafana-splunk-datasource, grafana-opensearch-datasource, grafana-falconlogscale-datasource, googlecloud-logging-datasource, victoriametrics-logs-datasource). */
  datasourceUid?: string;
  /** Key/value tag mappings used to build the log query. */
  tags?: Array<{ key: string; value?: string }>;
  /** Time-range shift applied to the start of the range (e.g. `-1m`). */
  spanStartTimeShift?: string;
  /** Time-range shift applied to the end of the range (e.g. `2m`). */
  spanEndTimeShift?: string;
  /** Filter logs by the trace ID (disabled when `customQuery` is true). */
  filterByTraceID?: boolean;
  /** Filter logs by the span ID (disabled when `customQuery` is true). */
  filterBySpanID?: boolean;
  /** Custom query template used when `customQuery` is true (interpolates `$__tags`). */
  query?: string;
  /** Use a custom query template instead of the default logfmt filter. */
  customQuery: boolean;
};

/** Legacy v1 shape retained for round-trip parity — the editor writes `tracesToLogsV2` instead. */
export type TraceToLogsV1Config = {
  datasourceUid?: string;
  tags?: string[];
  mappedTags?: Array<{ key: string; value?: string }>;
  mapTagNamesEnabled?: boolean;
  spanStartTimeShift?: string;
  spanEndTimeShift?: string;
  filterByTraceID?: boolean;
  filterBySpanID?: boolean;
  /** Pre-`tracesToLogsV2` legacy flag. */
  lokiSearch?: boolean;
};

/** Written by `TraceToMetricsSection` — may contain multiple named queries. */
export type TraceToMetricsConfig = {
  /** Metrics datasource UID (prometheus / victoriametrics-metrics-datasource). */
  datasourceUid?: string;
  tags?: Array<{ key: string; value: string }>;
  queries?: Array<{ name?: string; query?: string }>;
  spanStartTimeShift?: string;
  spanEndTimeShift?: string;
};

/** Written by `NodeGraphSection` — `{ enabled?: boolean }`. */
export type NodeGraphConfig = {
  enabled?: boolean;
};

/** Written by `SpanBarSection`. */
export type SpanBarConfig = {
  /** `'None' | 'Duration' | 'Tag'` — chosen via `Select`. */
  type?: 'None' | 'Duration' | 'Tag' | '';
  /** Required when `type === 'Tag'`. */
  tag?: string;
};

/**
 * Root (top-level datasource settings) fields the Zipkin plugin actually cares about.
 *
 * `url` is read directly by the HTTP client (`pkg/zipkin/zipkin.go:33-42`,
 * `pkg/zipkin/client.go:50,78,110,140`). `basicAuth`, `basicAuthUser`, and
 * `withCredentials` are populated by @grafana/plugin-ui's `Auth` component
 * and consumed by the SDK's `settings.HTTPClientOptions(ctx)`
 * (`pkg/zipkin/zipkin.go:23`).
 */
export type RootConfig = {
  /** Complete HTTP URL of the Zipkin v2 API root (default placeholder `http://localhost:9411`). Required — the backend fails at NewDatasource on empty URL. */
  url?: string;
  /** True when HTTP Basic authentication is enabled. Written by @grafana/plugin-ui `utils.ts` on auth-method select. */
  basicAuth?: boolean;
  /** Basic-auth username. Only meaningful when `basicAuth === true`. */
  basicAuthUser?: string;
  /**
   * Written by @grafana/plugin-ui's `getOnAuthMethodSelectHandler` when the
   * "Cross-site access control" method is selected. The Zipkin editor does
   * not offer that method in its `visibleMethods`, so this stays `false` in
   * practice.
   */
  withCredentials?: boolean;
};

/**
 * Fields stored in `jsonData`. Combines the TLS/HTTP fields written by
 * @grafana/plugin-ui's `Auth` + `AdvancedHttpSettings`, the trace-to-X /
 * node-graph / span-bar fields written by @grafana/o11y-ds-frontend's shared
 * sections, and the `nodeGraph` field also consumed by the Zipkin frontend
 * datasource at `src/datasource.ts:34`.
 *
 * None of these are read on the Go backend — the Zipkin plugin has no typed
 * backend jsonData contract; only `settings.URL` and SDK-managed HTTP
 * settings are consumed server-side.
 */
export type JsonDataConfig = {
  /** Enable TLS client authentication (mTLS). Requires `serverName`, `tlsClientCert`, `tlsClientKey`. Written by @grafana/plugin-ui TLSClientAuth. */
  tlsAuth?: boolean;
  /** Enable custom CA verification. Requires `tlsCACert`. Written by @grafana/plugin-ui SelfSignedCertificate. */
  tlsAuthWithCACert?: boolean;
  /** Skip TLS certificate validation. Not recommended outside testing. Written by @grafana/plugin-ui SkipTLSVerification. */
  tlsSkipVerify?: boolean;
  /** TLS SNI / cert-verification server name (`TLSClientAuth.tsx`). */
  serverName?: string;
  /** HTTP request timeout in seconds (`AdvancedHttpSettings.tsx`). */
  timeout?: number;
  /** Cookies to forward to the datasource, by name (`AdvancedHttpSettings.tsx`). */
  keepCookies?: string[];
  /**
   * Written by the "Forward OAuth Identity" auth method selector (@grafana/plugin-ui `utils.ts`).
   * When true, the SDK forwards the signed-in user's OAuth identity to the datasource.
   */
  oauthPassThru?: boolean;

  /**
   * v2 shape for trace-to-logs. Written by `TraceToLogsSection`; the editor
   * clears the legacy `tracesToLogs` key on any write.
   */
  tracesToLogsV2?: TraceToLogsV2Config;
  /**
   * Legacy v1 shape. `getTraceToLogsOptions`
   * (`packages/grafana-o11y-ds-frontend/src/TraceToLogs/TraceToLogsSettings.tsx`)
   * migrates this to `tracesToLogsV2` on read. Kept in the type for round-trip
   * compatibility of legacy datasources.
   */
  tracesToLogs?: TraceToLogsV1Config;

  /** Trace-to-metrics mapping. Written by `TraceToMetricsSection`. */
  tracesToMetrics?: TraceToMetricsConfig;

  /**
   * Toggle for the node-graph view. Written by `NodeGraphSection` and also
   * read by the Zipkin frontend datasource (`src/datasource.ts:34,55`) to
   * attach node-graph frames on the client side.
   */
  nodeGraph?: NodeGraphConfig;
  /** Span-bar decoration. Written by `SpanBarSection`. */
  spanBar?: SpanBarConfig;
};

/**
 * Secret key names stored in `secureJsonData` (write-only; read existing config via
 * `secureJsonFields`):
 * - `basicAuthPassword` — user password when Basic auth is enabled.
 * - `tlsCACert` — custom CA PEM when `tlsAuthWithCACert` is true.
 * - `tlsClientCert`, `tlsClientKey` — mTLS client credentials when `tlsAuth` is true.
 *
 * The editor also writes dynamic `httpHeaderValue<N>` secrets when the user configures
 * custom HTTP headers via @grafana/plugin-ui's `CustomHeaders` component. Those keys are
 * indexed pairs — not modeled as first-class fields in this schema; see the README.
 */
export type SecureJsonDataConfig = Array<
  'basicAuthPassword' | 'tlsCACert' | 'tlsClientCert' | 'tlsClientKey'
>;
