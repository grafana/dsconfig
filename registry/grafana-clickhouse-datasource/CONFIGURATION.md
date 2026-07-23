# ClickHouse configuration

Configuration reference for the **ClickHouse** data source (`grafana-clickhouse-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/grafana/plugins/grafana-clickhouse-datasource/).

> Generated from [`dsconfig.json`](dsconfig.json). Do not edit by hand — run `go generate ./...` to refresh.

## Fields

| Field | Type | Target | Required | Description |
|---|---|---|---|---|
| `jsonData.host` | string | jsonData | yes | Enter server address |
| `jsonData.port` | number | jsonData | yes | ClickHouse server port |
| `jsonData.protocol` | enum (native, http) | jsonData |  | Native or HTTP for server protocol |
| `jsonData.secure` | boolean | jsonData |  | Toggle on if the connection is secure |
| `jsonData.path` | string | jsonData |  | Additional URL path for HTTP requests |
| `jsonData.username` | string | jsonData |  | We recommend configuring a read-only user. |
| `secureJsonData.password` 🔒 | string | secureJsonData |  | ClickHouse password |
| `jsonData.tlsSkipVerify` | boolean | jsonData |  | Skip TLS Verify |
| `jsonData.tlsAuth` | boolean | jsonData |  | TLS Client Auth |
| `jsonData.tlsAuthWithCACert` | boolean | jsonData |  | Needed for verifying self-signed TLS Certs |
| `secureJsonData.tlsCACert` 🔒 | string (multiline) | secureJsonData |  | CA Cert |
| `secureJsonData.tlsClientCert` 🔒 | string (multiline) | secureJsonData |  | Client Cert |
| `secureJsonData.tlsClientKey` 🔒 | string (multiline) | secureJsonData |  | Client Key |
| `jsonData.configMode` | enum (classic, single-table) | jsonData |  | Choose how this datasource is used. 'Single table' provides a focused, compact query editor for one table. 'All databases' gives full access to explore any database and table. |
| `jsonData.signalType` | enum (logs, traces) | jsonData |  | What kind of data does this table contain? |
| `jsonData.defaultDatabase` | string | jsonData |  | the default database used by the query builder |
| `jsonData.defaultTable` | string | jsonData |  | the default table used by the query builder |
| `jsonData.dialTimeout` | string | jsonData |  | Timeout in seconds for connection |
| `jsonData.queryTimeout` | string | jsonData |  | Timeout in seconds for read queries |
| `jsonData.connMaxLifetime` | string | jsonData |  | Maximum lifetime of a connection in minutes |
| `jsonData.maxIdleConns` | string | jsonData |  | Maximum number of idle connections |
| `jsonData.maxOpenConns` | string | jsonData |  | Maximum number of open connections |
| `jsonData.validateSql` | boolean | jsonData |  | Validate SQL in the editor. |
| `jsonData.enableMapKeysDiscovery` | boolean | jsonData |  | When enabled, the filter editor probes Map(...) columns for distinct keys to populate the key-suggestion dropdown. On large tables with high-cardinality maps this probe can scan billions of rows. Disable to suppress the probe — operators can still type Map keys manually. Defaults to enabled. |
| `jsonData.logs.defaultDatabase` | string | jsonData |  | the default database used by the logs query builder |
| `jsonData.logs.defaultTable` | string | jsonData |  | the default table used by the logs query builder |
| `jsonData.logs.otelEnabled` | boolean | jsonData |  | Enables Open Telemetry schema versioning |
| `jsonData.logs.otelVersion` | enum | jsonData |  | OTel version |
| `jsonData.logs.filterTimeColumn` | string | jsonData |  | A lower precision column for filtering logs by timestamp |
| `jsonData.logs.timeColumn` | string | jsonData |  | Column for the log timestamp, used for high precision sorting |
| `jsonData.logs.levelColumn` | string | jsonData |  | Column for the log level |
| `jsonData.logs.messageColumn` | string | jsonData |  | Column for log message |
| `jsonData.logs.selectContextColumns` | boolean | jsonData |  | When enabled, will always include context columns in log queries |
| `jsonData.logs.contextColumns` | list | jsonData |  | Comma separated list of column names to use for identifying a log's source |
| `jsonData.logs.showLogLinks` | boolean | jsonData |  | Show "View logs" links on trace_id/traceid fields. |
| `jsonData.traces.defaultDatabase` | string | jsonData |  | the default database used by the trace query builder |
| `jsonData.traces.defaultTable` | string | jsonData |  | the default table used by the trace query builder |
| `jsonData.traces.otelEnabled` | boolean | jsonData |  | Enables Open Telemetry schema versioning |
| `jsonData.traces.otelVersion` | enum | jsonData |  | OTel version |
| `jsonData.traces.traceIdColumn` | string | jsonData |  | Column for the trace ID |
| `jsonData.traces.spanIdColumn` | string | jsonData |  | Column for the span ID |
| `jsonData.traces.operationNameColumn` | string | jsonData |  | Column for the operation name |
| `jsonData.traces.parentSpanIdColumn` | string | jsonData |  | Column for the parent span ID |
| `jsonData.traces.serviceNameColumn` | string | jsonData |  | Column for the service name |
| `jsonData.traces.durationColumn` | string | jsonData |  | Column for the duration time |
| `jsonData.traces.durationUnit` | enum (nanoseconds, microseconds, milliseconds, seconds) | jsonData |  | Unit used by your Duration column. OTel stores nanoseconds; other schemas often use milliseconds or seconds. |
| `jsonData.traces.startTimeColumn` | string | jsonData |  | Column for the start time |
| `jsonData.traces.tagsColumn` | string | jsonData |  | Column for the trace tags |
| `jsonData.traces.serviceTagsColumn` | string | jsonData |  | Column for the service tags |
| `jsonData.traces.kindColumn` | string | jsonData |  | Column for the trace kind |
| `jsonData.traces.statusCodeColumn` | string | jsonData |  | Column for the trace status code |
| `jsonData.traces.statusMessageColumn` | string | jsonData |  | Column for the trace status message |
| `jsonData.traces.stateColumn` | string | jsonData |  | Column for the trace state |
| `jsonData.traces.instrumentationLibraryNameColumn` | string | jsonData |  | Column for the instrumentation library name |
| `jsonData.traces.instrumentationLibraryVersionColumn` | string | jsonData |  | Column for the instrumentation library version |
| `jsonData.traces.flattenNested` | boolean | jsonData |  | Enable if your traces table was created with flatten_nested=1 |
| `jsonData.traces.traceEventsColumnPrefix` | string | jsonData |  | Prefix for the events column (Events.Timestamp, Events.Name, etc.) |
| `jsonData.traces.traceLinksColumnPrefix` | string | jsonData |  | Prefix for the trace references column (Links.TraceId, Links.TraceState, etc.) |
| `jsonData.traces.showTraceLinks` | boolean | jsonData |  | Show "View trace" links on trace_id/traceid fields. |
| `jsonData.traces.traceTimestampTableSuffix` | string | jsonData |  | Suffix appended to the traces table name to locate a companion index keyed by TraceId with Start/End columns. When such a table exists, trace ID lookups narrow the main query to a small time window instead of scanning the whole table. Leave blank to use the OTel default (_trace_id_ts). |
| `jsonData.httpHeaders` | list | jsonData |  | Add Custom HTTP headers when querying the database |
| `jsonData.httpHeaders[].name` | string | jsonData |  | Header Name |
| `jsonData.httpHeaders[].value` | string | jsonData |  | Empty when the header is stored securely; the encrypted value lives in secureJsonData under secureHttpHeaders.<Header Name>. |
| `jsonData.httpHeaders[].secure` | boolean | jsonData |  | Secure |
| `jsonData.forwardGrafanaHeaders` | boolean | jsonData |  | Forward Grafana HTTP Headers to datasource. |
| `jsonData.customSettings` | list | jsonData |  | Additional ClickHouse settings sent with every query as SETTINGS name=value. |
| `jsonData.customSettings[].setting` | string | jsonData |  | Setting |
| `jsonData.customSettings[].value` | string | jsonData |  | Value |
| `jsonData.aliasTables` | list | jsonData |  | Provide alias tables with a (`alias` String, `select` String, `type` String) schema to use as a source for column selection. |
| `jsonData.aliasTables[].targetDatabase` | string | jsonData |  | Target Database |
| `jsonData.aliasTables[].targetTable` | string | jsonData |  | Target Table |
| `jsonData.aliasTables[].aliasDatabase` | string | jsonData |  | Alias Database |
| `jsonData.aliasTables[].aliasTable` | string | jsonData |  | Alias Table |
| `jsonData.enableRowLimit` | boolean | jsonData |  | Enable using the Grafana row limit setting to limit the number of rows returned from Clickhouse. Ensure the appropriate permissions are set for your user. Only supported for Grafana >= 11.0.0. Defaults to false. |
| `jsonData.rowLimit` | number | jsonData |  | Row Limit |
| `jsonData.hideTableNameInAdhocFilters` | boolean | jsonData |  | Show only column names in ad hoc filter keys instead of the full "table.column" format. This simplifies the filter interface when working with schemas that have many tables. Defaults to false. |
| `jsonData.version` | string | jsonData |  | Plugin version that last wrote this configuration. Stamped by the config editor's useConfigDefaults hook on every save; frontend-only. |
| `jsonData.enableSchemaCache` | boolean | jsonData |  | Gates the in-process cache that memoizes system.tables / system.columns / DISTINCT column-value lookups used by the query builder. Not exposed in the configuration editor; backend defaults it to true when unset. |
| `jsonData.schemaCacheTTLSeconds` | number | jsonData |  | Controls how long schema-introspection results are considered fresh. Not exposed in the configuration editor; backend defaults it to 60 when unset or <=0. |

## Provisioning examples

Each scenario below shows how to provision the data source in Grafana using a YAML file (loaded by Grafana's [file provisioner](https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources)) and using the [Grafana Terraform provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source).

Placeholders like `<YOUR_TOKEN>` must be replaced with real values before use.

### Default configuration

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: ClickHouse
    type: grafana-clickhouse-datasource
    access: proxy
    jsonData:
      configMode: classic
      connMaxLifetime: "5"
      dialTimeout: "10"
      enableMapKeysDiscovery: true
      enableRowLimit: false
      enableSchemaCache: true
      forwardGrafanaHeaders: false
      hideTableNameInAdhocFilters: false
      host: Server address
      logs:
        defaultTable: otel_logs
        otelVersion: latest
        selectContextColumns: true
      maxIdleConns: "25"
      maxOpenConns: "50"
      port: 9000
      protocol: native
      queryTimeout: "60"
      schemaCacheTTLSeconds: 60
      secure: false
      signalType: logs
      tlsAuth: false
      tlsAuthWithCACert: false
      tlsSkipVerify: false
      traces:
        defaultTable: otel_traces
        durationUnit: "nanoseconds"
        flattenNested: false
        otelVersion: latest
      validateSql: false
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_clickhouse_datasource" {
  type = "grafana-clickhouse-datasource"
  name = "ClickHouse"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    configMode = "classic"
    connMaxLifetime = "5"
    dialTimeout = "10"
    enableMapKeysDiscovery = true
    enableRowLimit = false
    enableSchemaCache = true
    forwardGrafanaHeaders = false
    hideTableNameInAdhocFilters = false
    host = "Server address"
    logs = {
      defaultTable = "otel_logs"
      otelVersion = "latest"
      selectContextColumns = true
    }
    maxIdleConns = "25"
    maxOpenConns = "50"
    port = 9000
    protocol = "native"
    queryTimeout = "60"
    schemaCacheTTLSeconds = 60
    secure = false
    signalType = "logs"
    tlsAuth = false
    tlsAuthWithCACert = false
    tlsSkipVerify = false
    traces = {
      defaultTable = "otel_traces"
      durationUnit = "nanoseconds"
      flattenNested = false
      otelVersion = "latest"
    }
    validateSql = false
  })
}
```

