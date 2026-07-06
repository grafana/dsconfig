/**
 * Configuration models for the Jaeger datasource plugin (`jaeger`).
 *
 * Sources of truth (https://github.com/grafana/grafana-jaeger-datasource @ 7014ae8):
 * - `src/plugin.json` — plugin id (`"jaeger"`), name (`"Jaeger"`), docs URL
 * - `src/configuration/ConfigEditor.tsx:20-70` — outermost editor (composes
 *   `DataSourceDescription`, `ConnectionSettings` with `urlPlaceholder="http://localhost:16686"`,
 *   `Auth` (via `convertLegacyAuthProps`), `TraceToLogsSection`, `TraceToMetricsSection`,
 *   and a collapsible "Additional settings" section containing `AdvancedHttpSettings`,
 *   `SecureSocksProxySettings` (excluded), `NodeGraphSection`, `SpanBarSection`,
 *   and plugin-local `TraceIdTimeParams`)
 * - `src/configuration/TraceIdTimeParams.tsx:11-45` — plugin-local component;
 *   heading `"Query Trace by ID with Time Params"`, `InlineField label="Enable Time Parameters"`,
 *   tooltip `"pass time parameters when querying trace by ID"`; writes
 *   `jsonData.traceIdTimeParams.enabled: boolean`
 * - `src/types.ts:38-59` — frontend `JaegerQuery` (query-level, not config)
 * - `pkg/jaeger/jaeger.go:21-55` — backend `NewDatasource` reads
 *   `settings.URL` directly and unmarshals `jsonData.traceIdTimeParams` via
 *   the ad-hoc `datasourceJSONData` struct; `settings.HTTPClientOptions(ctx)`
 *   builds the HTTP client
 * - `pkg/jaeger/jaeger.go:57-78` — `CheckHealth` calls `Services` (HTTP) or
 *   `GrpcServices` when the Grafana `jaegerEnableGrpcEndpoint` feature toggle
 *   is enabled
 * - `pkg/jaeger/client.go:26-34, 242-266` — HTTP client honours
 *   `settings.URL` and consumes `jsonData.traceIdTimeParams.enabled` when
 *   building the trace-by-ID request
 * - `pkg/jaeger/types/types.go:11-15` — `SettingsJSONData` — the backend's
 *   only jsonData reader; a minimal `{ traceIdTimeParams: { enabled: bool } }`
 *
 * External components consulted at their pinned versions:
 * - `@grafana/plugin-ui@0.13.1` — `ConnectionSettings` (URL input; default
 *   `urlLabel="URL"`, placeholder from plugin), `Auth` +
 *   `convertLegacyAuthProps` (visible methods `[BasicAuth, OAuthForward, NoAuth]`;
 *   option labels `Basic authentication` / `Forward OAuth Identity` /
 *   `No Authentication`; description "Choose an authentication method to
 *   access the data source"), `AdvancedHttpSettings` (Allowed cookies,
 *   Timeout), `DataSourceDescription`, `ConfigSection`
 * - `@grafana/o11y-ds-frontend@13.1.0-25027462778` — `TraceToLogsSection`
 *   (writes `jsonData.tracesToLogsV2` v2 shape and clears legacy
 *   `jsonData.tracesToLogs`), `TraceToMetricsSection` (writes
 *   `jsonData.tracesToMetrics`), `NodeGraphSection` (writes
 *   `jsonData.nodeGraph.enabled`), `SpanBarSection` (writes
 *   `jsonData.spanBar.{type, tag}`)
 * - `@grafana/ui@13.1.0-24716567714` — `InlineField`, `InlineFieldRow`,
 *   `InlineSwitch`, `Divider`, `Stack`, `SecureSocksProxySettings` (excluded)
 * - `@grafana/data@^13.1.0-24716567714` — `DataSourceJsonData`,
 *   `DataSourcePluginOptionsEditorProps`,
 *   `updateDatasourcePluginJsonDataOption`
 * - `@grafana/runtime@^12.4.2` — `config` (reads
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

/** Written by the plugin-local `TraceIdTimeParams` component (`src/configuration/TraceIdTimeParams.tsx:11-45`). */
export type TraceIdTimeParamsConfig = {
  /** When true, append `start` and `end` query parameters to `GET /api/traces/{traceID}`. */
  enabled?: boolean;
};

/**
 * Root (top-level datasource settings) fields the Jaeger plugin actually cares about.
 *
 * `url` is read directly by the HTTP client (`pkg/jaeger/jaeger.go:38-40`,
 * `pkg/jaeger/client.go:29`). `basicAuth`, `basicAuthUser`, and
 * `withCredentials` are populated by @grafana/plugin-ui's `Auth` component
 * and consumed by the SDK's `settings.HTTPClientOptions(ctx)`
 * (`pkg/jaeger/jaeger.go:28`).
 */
export type RootConfig = {
  /** Complete HTTP URL of the Jaeger query server (default placeholder `http://localhost:16686`). Required — the backend fails at NewDatasource on empty URL. */
  url?: string;
  /** True when HTTP Basic authentication is enabled. Written by @grafana/plugin-ui `utils.ts:47`. */
  basicAuth?: boolean;
  /** Basic-auth username. Only meaningful when `basicAuth === true`. */
  basicAuthUser?: string;
  /**
   * Written by @grafana/plugin-ui's `getOnAuthMethodSelectHandler` when the
   * "Cross-site access control" method is selected. The Jaeger editor does
   * not offer that method in its `visibleMethods`, so this stays `false` in
   * practice.
   */
  withCredentials?: boolean;
};

/**
 * Fields stored in `jsonData`. Combines the plugin-local `traceIdTimeParams`
 * (`src/configuration/TraceIdTimeParams.tsx`, `pkg/jaeger/types/types.go:11-15`),
 * the TLS/HTTP fields written by @grafana/plugin-ui's `Auth` +
 * `AdvancedHttpSettings`, and the trace-to-X / node-graph / span-bar fields
 * written by @grafana/o11y-ds-frontend's shared sections.
 */
export type JsonDataConfig = {
  /** Enable TLS client authentication (mTLS). Requires `serverName`, `tlsClientCert`, `tlsClientKey`. Written by @grafana/plugin-ui `utils.ts:123`. */
  tlsAuth?: boolean;
  /** Enable custom CA verification. Requires `tlsCACert`. Written by @grafana/plugin-ui `utils.ts:94`. */
  tlsAuthWithCACert?: boolean;
  /** Skip TLS certificate validation. Not recommended outside testing. Written by @grafana/plugin-ui `utils.ts:182`. */
  tlsSkipVerify?: boolean;
  /** TLS SNI / cert-verification server name (`TLSClientAuth.tsx`). */
  serverName?: string;
  /** HTTP request timeout in seconds (`AdvancedHttpSettings.tsx`). */
  timeout?: number;
  /** Cookies to forward to the datasource, by name (`AdvancedHttpSettings.tsx`). */
  keepCookies?: string[];
  /**
   * Written by the "Forward OAuth Identity" auth method selector (`@grafana/plugin-ui utils.ts:51`).
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
   * (`packages/grafana-o11y-ds-frontend/src/TraceToLogs/TraceToLogsSettings.tsx:55-73`)
   * migrates this to `tracesToLogsV2` on read. Kept in the type for round-trip
   * compatibility of legacy datasources.
   */
  tracesToLogs?: TraceToLogsV1Config;

  /** Trace-to-metrics mapping. Written by `TraceToMetricsSection`. */
  tracesToMetrics?: TraceToMetricsConfig;

  /** Toggle for the node-graph view. Written by `NodeGraphSection`. */
  nodeGraph?: NodeGraphConfig;
  /** Span-bar decoration. Written by `SpanBarSection`. */
  spanBar?: SpanBarConfig;

  /**
   * Plugin-local. When `enabled` is true, the backend appends `start` and
   * `end` query parameters to `GET /api/traces/{traceID}` — see
   * `pkg/jaeger/client.go:242-266`. This is the only jsonData field the
   * Jaeger backend actually reads.
   */
  traceIdTimeParams?: TraceIdTimeParamsConfig;
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
