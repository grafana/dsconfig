/**
 * Configuration models for the ClickHouse datasource plugin
 * (`grafana-clickhouse-datasource`).
 *
 * Sources of truth (https://github.com/grafana/clickhouse-datasource @
 * d55f9d6250023a86d49555b12fffdbeec9a1b538):
 * - `src/plugin.json` — plugin `id`, `name`, docs link
 * - `src/types/config.ts` — `CHConfig`, `CHSecureConfig`, `CHHttpHeader`,
 *   `CHCustomSetting`, `AliasTableEntry`, `CHLogsConfig`, `CHTracesConfig`,
 *   `Protocol`, `ConfigMode`, `SignalType`, `defaultCHAdditionalSettingsConfig`
 * - `src/views/CHConfigEditor.tsx` — v1 configuration editor (labels,
 *   placeholders, tooltips, section titles)
 * - `src/views/CHConfigEditorHooks.ts` — `useConfigDefaults`,
 *   `onHttpHeadersChange` (secureHttpHeaders.<Name> secret-key convention)
 * - `src/labels.ts` — the label / placeholder / tooltip catalog
 * - `src/otel.ts` — OTel column maps and `defaultLogsTable`/`defaultTraceTable`
 *   (`otel_logs`, `otel_traces`)
 * - `pkg/plugin/settings.go` — backend `Settings` struct + `LoadSettings`
 *   (v3 legacy fallbacks, string-or-number tolerance, defaulting, secret
 *   copies, secureHttpHeaders.* handling)
 *
 * External components:
 * - `@grafana/data` 12.4.2, `@grafana/ui` 12.4.2, `@grafana/runtime` 12.4.2
 *   pinned in the plugin's `package.json`.
 */

/**
 * Root (top-level datasource settings) fields.
 *
 * The ClickHouse plugin's backend (pkg/plugin/settings.go) does not read any
 * root-level datasource fields (`url`, `basicAuth`, `user`, `database`, …);
 * every setting lives in jsonData or secureJsonData. This type is therefore a
 * blank object, never null.
 */
export type RootConfig = Record<string, never>;

/**
 * Wire protocol between Grafana and ClickHouse. Determines the driver
 * (`clickhouse-go` native TCP vs HTTP) and the default port picker.
 * Stored in jsonData.protocol.
 */
export type Protocol = 'native' | 'http';

/**
 * Datasource UI layout switch: 'classic' exposes every database/table to the
 * query builder; 'single-table' pins to one table and shows a compact editor.
 * Stored in jsonData.configMode.
 */
export type ConfigMode = 'classic' | 'single-table';

/**
 * Kind of data the single-table datasource holds. Only meaningful when
 * configMode === 'single-table'. Stored in jsonData.signalType.
 */
export type SignalType = 'logs' | 'traces';

/**
 * Trace duration column unit. OTel-instrumented tables use nanoseconds;
 * hand-rolled schemas often use milliseconds or seconds. Stored in
 * jsonData.traces.durationUnit.
 */
export type TraceDurationUnit = 'nanoseconds' | 'microseconds' | 'milliseconds' | 'seconds';

/**
 * One entry of the Custom HTTP Headers editor. When `secure` is true the value
 * is moved to secureJsonData under the key `secureHttpHeaders.<name>` and the
 * `value` field here is cleared to an empty string
 * (`src/views/CHConfigEditorHooks.ts:36-52`).
 */
export interface HttpHeader {
    name: string;
    value: string;
    secure: boolean;
}

/**
 * One entry of the ClickHouse Custom Settings editor. Appended to every query
 * as `SETTINGS <setting>=<value>` by the backend.
 */
export interface CustomSetting {
    setting: string;
    value: string;
}

/**
 * One entry of the Column Alias Tables editor. Points the query builder at a
 * separate `(alias String, select String, type String)` table that supplies
 * user-friendly column choices for a real ClickHouse table.
 */
export interface AliasTableEntry {
    targetDatabase: string;
    targetTable: string;
    aliasDatabase: string;
    aliasTable: string;
}

/**
 * Nested `jsonData.logs` configuration. Mirrors `CHLogsConfig` in
 * `src/types/config.ts:101-116` verbatim. When `otelEnabled` is true the
 * column-name fields are ignored at runtime — the OTel column map drives them
 * instead (`src/otel.ts`).
 */
export interface LogsConfig {
    defaultDatabase?: string;
    /** Default `otel_logs` — the OTel exporter's default table. */
    defaultTable?: string;
    otelEnabled?: boolean;
    otelVersion?: string;
    filterTimeColumn?: string;
    timeColumn?: string;
    levelColumn?: string;
    messageColumn?: string;
    selectContextColumns?: boolean;
    contextColumns?: string[];
    showLogLinks?: boolean;
}

/**
 * Nested `jsonData.traces` configuration. Mirrors `CHTracesConfig` in
 * `src/types/config.ts:118-154` verbatim. As with logs, `otelEnabled` makes
 * the column-name fields runtime-derived from the OTel column map.
 */
export interface TracesConfig {
    defaultDatabase?: string;
    /** Default `otel_traces` — the OTel exporter's default table. */
    defaultTable?: string;
    otelEnabled?: boolean;
    otelVersion?: string;
    traceIdColumn?: string;
    spanIdColumn?: string;
    operationNameColumn?: string;
    parentSpanIdColumn?: string;
    serviceNameColumn?: string;
    durationColumn?: string;
    durationUnit?: TraceDurationUnit;
    startTimeColumn?: string;
    tagsColumn?: string;
    serviceTagsColumn?: string;
    kindColumn?: string;
    statusCodeColumn?: string;
    statusMessageColumn?: string;
    stateColumn?: string;
    instrumentationLibraryNameColumn?: string;
    instrumentationLibraryVersionColumn?: string;
    flattenNested?: boolean;
    traceEventsColumnPrefix?: string;
    traceLinksColumnPrefix?: string;
    showTraceLinks?: boolean;
    /**
     * Suffix appended to the traces table name to locate a companion trace-id
     * index table. Blank inherits the OTel convention (`_trace_id_ts`).
     */
    traceTimestampTableSuffix?: string;
}

/**
 * Fields stored in `jsonData`. Matches the plugin's `CHConfig`
 * (`src/types/config.ts:15-79`) shape plus the two backend-only cache knobs
 * (`enableSchemaCache`, `schemaCacheTTLSeconds`) declared in
 * `pkg/plugin/settings.go:52-59`. The `enableSecureSocksProxy` field is
 * deliberately excluded per registry policy.
 */
export type JsonDataConfig = {
    /** Required. Read at `pkg/plugin/settings.go:87-92`. */
    host?: string;
    /** Required. JSON number in the current storage shape; the backend also tolerates a JSON string (`pkg/plugin/settings.go:94-103`). */
    port?: number;
    /** Defaults to `'native'` in the editor and in the backend after loading. */
    protocol?: Protocol;
    /** Toggles the driver's TLS handshake; independent of the tlsAuth* switches. Defaults to false. */
    secure?: boolean;
    /** Optional URL prefix for HTTP requests; ignored under native protocol. */
    path?: string;

    /** Also known as InsecureSkipVerify on the backend (`pkg/plugin/settings.go:26`). */
    tlsSkipVerify?: boolean;
    /** Enables mTLS with a client certificate + key. */
    tlsAuth?: boolean;
    /** Verifies the server's certificate against a custom CA. */
    tlsAuthWithCACert?: boolean;

    username?: string;

    defaultDatabase?: string;
    /**
     * Frontend-only surface for the classic-mode default table selector; the
     * backend has no field for it. Kept here to preserve write-round-trip.
     */
    defaultTable?: string;

    /** Timeouts are stored as JSON strings; backend also tolerates JSON numbers as legacy. */
    connMaxLifetime?: string;
    dialTimeout?: string;
    queryTimeout?: string;
    maxIdleConns?: string;
    maxOpenConns?: string;

    /** Editor-side SQL validation toggle; not read by the backend. */
    validateSql?: boolean;

    /** Nested logs config (see `LogsConfig`). */
    logs?: LogsConfig;
    /** Nested traces config (see `TracesConfig`). */
    traces?: TracesConfig;

    aliasTables?: AliasTableEntry[];

    /** HTTP-only: additional headers to send with every query. */
    httpHeaders?: HttpHeader[];
    /** HTTP-only: forward Grafana's incoming HTTP headers to the datasource. */
    forwardGrafanaHeaders?: boolean;

    customSettings?: CustomSetting[];

    enableRowLimit?: boolean;
    hideTableNameInAdhocFilters?: boolean;
    /**
     * Controls the Map-column key discovery probe. Defaults to true in the
     * editor and in typed backends; disable on very large tables.
     */
    enableMapKeysDiscovery?: boolean;

    /**
     * Configuration mode: 'classic' (all databases) or 'single-table'
     * (focused). Written by the editor; frontend-only.
     */
    configMode?: ConfigMode;
    /** Signal type when configMode is 'single-table'. Frontend-only. */
    signalType?: SignalType;

    /**
     * Plugin version stamped by the config editor's `useConfigDefaults` hook
     * on every save (`src/views/CHConfigEditorHooks.ts:94`); frontend-only.
     */
    version?: string;

    /** Backend-only: gates the schema-introspection cache. Defaults to true. */
    enableSchemaCache?: boolean;
    /** Backend-only: schema-introspection cache TTL in seconds. Defaults to 60. */
    schemaCacheTTLSeconds?: number;
};

/**
 * Secret key names stored in `secureJsonData` (write-only; read existing
 * config via `secureJsonFields`):
 * - `password` — ClickHouse user password
 * - `tlsCACert` — CA cert PEM (when `jsonData.tlsAuthWithCACert` is true)
 * - `tlsClientCert` / `tlsClientKey` — client-cert PEM pair (when
 *   `jsonData.tlsAuth` is true)
 *
 * Custom HTTP header values marked `secure` are stored under dynamic keys of
 * the form `secureHttpHeaders.<Header Name>` (see
 * `src/views/CHConfigEditorHooks.ts:36-52` and `pkg/plugin/settings.go:319-344`).
 * The dsconfig schema only enumerates the four static keys above; the dynamic
 * secureHttpHeaders.* keys are documented in the schema's instructions.
 */
export type SecureJsonDataConfig = Array<'password' | 'tlsCACert' | 'tlsClientCert' | 'tlsClientKey'>;
