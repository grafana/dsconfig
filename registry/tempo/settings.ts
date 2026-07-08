/**
 * Configuration models for the Tempo datasource plugin (`tempo`).
 *
 * Sources of truth (https://github.com/grafana/grafana-tempo-datasource @ 4858762):
 * - `src/plugin.json` — plugin id ("tempo"), name ("Tempo"), docs URL
 * - `src/configuration/ConfigEditor.tsx` — the outermost editor (composes @grafana/plugin-ui
 *   `DataSourceDescription`, `ConnectionSettings`, `Auth` (via `convertLegacyAuthProps`),
 *   `AdvancedHttpSettings`, plus plugin-local `StreamingSection`, `ServiceGraphSettings`,
 *   `QuerySettings`, `TagsTimeRangeSettings`, `TagLimitSection`, `TraceQLSearchSettings`,
 *   and @grafana/o11y-ds-frontend `TraceToLogsSection`, `TraceToMetricsSection`,
 *   `TraceToProfilesSection`, `NodeGraphSection`, `SpanBarSection`)
 * - `src/configuration/StreamingSection.tsx:13-80` — `streamingEnabled.search` and
 *   `streamingEnabled.metrics` toggles (labels "Search queries", "Metrics queries";
 *   tooltips reference `featuresToTempoVersion` in `datasource.ts:79-82`)
 * - `src/configuration/ServiceGraphSettings.tsx:15-56` — `serviceMap.datasourceUid`
 *   DataSourcePicker filtered to `pluginId: 'prometheus'`
 * - `src/configuration/QuerySettings.tsx:15-74` — `traceQuery.{timeShiftEnabled,
 *   spanStartTimeShift, spanEndTimeShift}` toggle + two IntervalInputs
 * - `src/configuration/TagsTimeRangeSettings.tsx:10-45` — `timeRangeForTags` Combobox
 *   with the five second-count options; DEFAULT_TIME_RANGE_FOR_TAGS = 1800
 * - `src/configuration/TagLimitSettings.tsx:12-59` — `tagLimit` numeric input
 *   (placeholder "5000")
 * - `src/configuration/TraceQLSearchSettings.tsx:16-48` — `search.hide` switch and
 *   `search.filters` TraceQLSearchTags editor
 * - `src/types.ts:6-29` — frontend `TempoJsonData extends DataSourceJsonData`
 * - `src/datasource.ts:70-159` — frontend consumption of the trace-to-X / serviceMap
 *   / search fields; seeds `search.filters` with default TraceQL scopes on first load
 * - `pkg/tempo/tempo.go:52-90` — backend `NewDatasource` reads `settings.URL`
 *   directly; `pkg/tempo/tempo.go:150-208` `CheckHealth` reads `jsonData.streamingEnabled.search`
 *   as an untyped `map[string]interface{}` — the Tempo backend does not ship a
 *   typed `pkg/models/settings.go`.
 * - `pkg/tempo/grpc.go:85-137` — gRPC client honours `settings.URL`,
 *   `settings.BasicAuthEnabled`, and secure socks via `settings.ProxyClient`
 *
 * External components consulted at their pinned versions:
 * - `@grafana/plugin-ui@0.15.1` — `ConnectionSettings` (URL input), `Auth` with
 *   `convertLegacyAuthProps` (AuthMethodSettings, BasicAuth, TLSSettings via
 *   SelfSignedCertificate/TLSClientAuth/SkipTLSVerification, CustomHeaders),
 *   `AdvancedHttpSettings` (Allowed cookies, Timeout), `DataSourceDescription`,
 *   `ConfigSection`/`ConfigSubSection`, `ConfigDescriptionLink`
 * - `@grafana/o11y-ds-frontend@13.1.0-canary` — `TraceToLogsSection` writes
 *   `jsonData.tracesToLogsV2` (v2 shape) and clears legacy `jsonData.tracesToLogs`;
 *   `TraceToMetricsSection` writes `jsonData.tracesToMetrics`;
 *   `TraceToProfilesSection` writes `jsonData.tracesToProfiles`;
 *   `NodeGraphSection` writes `jsonData.nodeGraph.enabled`;
 *   `SpanBarSection` writes `jsonData.spanBar.{type, tag}`
 * - `@grafana/ui@13.1.0-canary` — `Input`, `Switch`, `Combobox`, `SecureSocksProxySettings`
 *   (rendered conditionally, excluded here), `DataSourcePicker` (in ServiceGraphSettings)
 * - `@grafana/data@13.1.0-canary` — `DataSourceJsonData` base interface, `DataSourcePluginOptionsEditorProps`,
 *   `updateDatasourcePluginJsonDataOption`
 */

/** Streaming toggles written by `StreamingSection.tsx:44-77`. */
export type StreamingEnabledConfig = {
  /** Enable streaming for TraceQL search queries. Requires Tempo >= 2.2.0. */
  search?: boolean;
  /** Enable streaming for TraceQL metrics queries. Requires Tempo >= 2.7.0. */
  metrics?: boolean;
};

/** Legacy v1 shape kept for round-trip parity — the editor writes `tracesToLogsV2` instead. */
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

/** Written by `TraceToMetricsSection` — may contain multiple named queries. */
export type TraceToMetricsConfig = {
  /** Metrics datasource UID (prometheus / victoriametrics-metrics-datasource). */
  datasourceUid?: string;
  tags?: Array<{ key: string; value: string }>;
  queries?: Array<{ name?: string; query?: string }>;
  spanStartTimeShift?: string;
  spanEndTimeShift?: string;
};

/** Written by `TraceToProfilesSection`. */
export type TraceToProfilesConfig = {
  /** Profiles datasource UID (`grafana-pyroscope-datasource`). */
  datasourceUid?: string;
  tags?: Array<{ key: string; value?: string }>;
  /** Pyroscope profile-type identifier. */
  profileTypeId?: string;
  /** Custom query template used when `customQuery` is true. */
  query?: string;
  /** Use a custom query template. */
  customQuery: boolean;
};

/** Written by `ServiceGraphSettings.tsx:32-37` (Prometheus datasource UID). */
export type ServiceMapConfig = {
  datasourceUid?: string;
};

/** Written by `NodeGraphSection` — `{ enabled?: boolean }`. */
export type NodeGraphConfig = {
  enabled?: boolean;
};

/** Written by `SpanBarSection`. */
export type SpanBarConfig = {
  /** `'None' | 'Duration' | 'Tag'` — chosen via `Select` (SpanBarSettings.tsx:32-49). */
  type?: 'None' | 'Duration' | 'Tag' | '';
  /** Required when `type === 'Tag'`. */
  tag?: string;
};

/** TraceQL scope for a static search filter (`dataquery.ts:105-113`). */
export type TraceqlSearchScope =
  | 'event'
  | 'instrumentation'
  | 'intrinsic'
  | 'link'
  | 'resource'
  | 'span'
  | 'unscoped';

/** A single TraceQL search filter (`dataquery.ts:115-142`). */
export type TraceqlFilter = {
  /** Frontend-only identifier; not used in query generation. */
  id: string;
  isCustomValue?: boolean;
  operator?: string;
  scope?: TraceqlSearchScope;
  tag?: string;
  value?: string | string[];
  valueType?: string;
};

/** Written by `TraceQLSearchSettings.tsx:28-46`. */
export type SearchConfig = {
  /** Remove the search tab from the query editor. */
  hide?: boolean;
  /** Static filters exposed in the search UI; the frontend seeds two defaults on first load (`datasource.ts:147-159`). */
  filters?: TraceqlFilter[];
};

/** Written by `QuerySettings.tsx:26-72`. */
export type TraceQueryConfig = {
  /** Enable the "Use time range in query" toggle. Default: false. */
  timeShiftEnabled?: boolean;
  /** Shift applied to the start of the search range (e.g. `30m`). */
  spanStartTimeShift?: string;
  /** Shift applied to the end of the search range (e.g. `30m`). */
  spanEndTimeShift?: string;
};

/**
 * Root (top-level datasource settings) fields the Tempo plugin actually cares about.
 *
 * `url` is read directly by both the HTTP client (`pkg/tempo/tempo.go:78`) and the
 * gRPC streaming client (`pkg/tempo/grpc.go:86`). `basicAuth`, `basicAuthUser`, and
 * `withCredentials` are populated by @grafana/plugin-ui's `Auth` component and
 * consumed by the SDK's `settings.HTTPClientOptions(ctx)` (`pkg/tempo/tempo.go:54`);
 * `basicAuth` also gates the gRPC per-RPC BasicAuth credentials
 * (`pkg/tempo/grpc.go:178-184`).
 */
export type RootConfig = {
  /** Complete HTTP URL of the Tempo server. Required — the backend fails at request time on empty URL. */
  url?: string;
  /** True when HTTP Basic authentication is enabled. Written by @grafana/plugin-ui `utils.ts:47`. */
  basicAuth?: boolean;
  /** Basic-auth username. Only meaningful when `basicAuth === true`. */
  basicAuthUser?: string;
  /**
   * Written by @grafana/plugin-ui's `getOnAuthMethodSelectHandler` when the
   * "Cross-site access control" method is selected. The Tempo editor does not
   * offer that method in its `visibleMethods`, so this stays `false` in practice.
   */
  withCredentials?: boolean;
};

/**
 * Fields stored in `jsonData`. Combines the plugin's `TempoJsonData` (`src/types.ts:6-29`),
 * the TLS/HTTP fields written by @grafana/plugin-ui's `Auth` + `AdvancedHttpSettings`,
 * and the trace-to-X / node-graph / span-bar fields written by @grafana/o11y-ds-frontend's
 * shared sections.
 */
export type JsonDataConfig = {
  /** Enable TLS client authentication (mTLS). Requires `serverName`, `tlsClientCert`, `tlsClientKey`. Written by @grafana/plugin-ui `utils.ts:123`. */
  tlsAuth?: boolean;
  /** Enable custom CA verification. Requires `tlsCACert`. Written by @grafana/plugin-ui `utils.ts:94`. */
  tlsAuthWithCACert?: boolean;
  /** Skip TLS certificate validation. Not recommended outside testing. Written by @grafana/plugin-ui `utils.ts:182`. */
  tlsSkipVerify?: boolean;
  /** TLS SNI / cert-verification server name (`TLSClientAuth.tsx:51`). */
  serverName?: string;
  /** HTTP request timeout in seconds (`AdvancedHttpSettings.tsx:63`). */
  timeout?: number;
  /** Cookies to forward to the datasource, by name (`AdvancedHttpSettings.tsx:48-58`). */
  keepCookies?: string[];
  /**
   * Written by the "Forward OAuth Identity" auth method selector (`@grafana/plugin-ui utils.ts:51`).
   * When true, the SDK forwards the signed-in user's OAuth identity to the datasource.
   */
  oauthPassThru?: boolean;

  /**
   * Streaming toggles for search and metrics queries. Written by
   * `StreamingSection.tsx:44-77`. `CheckHealth` reads `streamingEnabled.search`
   * to decide whether to probe the gRPC streaming endpoint
   * (`pkg/tempo/tempo.go:150-208`).
   */
  streamingEnabled?: StreamingEnabledConfig;

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
  /** Trace-to-profiles mapping. Written by `TraceToProfilesSection`. */
  tracesToProfiles?: TraceToProfilesConfig;

  /** Prometheus datasource providing service-graph metrics. Written by `ServiceGraphSettings.tsx`. */
  serviceMap?: ServiceMapConfig;
  /** Toggle for the node-graph view. Written by `NodeGraphSection`. */
  nodeGraph?: NodeGraphConfig;
  /** Span-bar decoration. Written by `SpanBarSection`. */
  spanBar?: SpanBarConfig;

  /** TraceQL search tab config. Written by `TraceQLSearchSettings.tsx`. */
  search?: SearchConfig;

  /** TraceID query time-range shifts. Written by `QuerySettings.tsx`. */
  traceQuery?: TraceQueryConfig;

  /**
   * Time range applied to tag/tag-value queries in the editor, in seconds.
   * Default 1800 (30 minutes). Allowed values: 1800, 10800, 86400, 259200, 604800.
   * Written by `TagsTimeRangeSettings.tsx:31-37`.
   */
  timeRangeForTags?: number;

  /**
   * Max number of tags and tag values displayed in the Tempo editor. Default 5000
   * (rendered as a placeholder — the editor never persists the default).
   * Written by `TagLimitSettings.tsx:32-34` as `v.currentTarget.value`, which
   * flows through `updateDatasourcePluginJsonDataOption` as a string. The
   * frontend expression `options.jsonData.tagLimit || ''` then treats it as
   * either a number or a numeric string.
   */
  tagLimit?: number;
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
