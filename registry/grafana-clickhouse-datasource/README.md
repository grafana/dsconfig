# grafana-clickhouse-datasource

Declarative configuration schema for the [ClickHouse datasource plugin](https://github.com/grafana/clickhouse-datasource) (`grafana-clickhouse-datasource`).

## Upstream researched

- **Repo**: `github.com/grafana/clickhouse-datasource`
- **Ref**: `main`
- **Commit SHA**: `d55f9d6250023a86d49555b12fffdbeec9a1b538` (2026-07-04, `Fix datasource variable refs in Data Analysis dashboard (#1896)`)

Every value in [`dsconfig.json`](dsconfig.json) — labels, placeholders, descriptions, option
labels/values, section titles, defaults, dependency and required-when expressions, storage keys,
storage targets, and value types — is traceable to a specific `file:line` in the upstream repo at
this SHA. See [Field provenance](#field-provenance).

To reproduce this research:

```bash
git clone https://github.com/grafana/clickhouse-datasource
cd clickhouse-datasource
git checkout d55f9d6250023a86d49555b12fffdbeec9a1b538
```

If upstream `main` has advanced past this SHA, re-diff the sources listed under
[Sources researched](#sources-researched) before merging any changes to this entry.

## Files

| File | Purpose |
| --- | --- |
| [`dsconfig.json`](dsconfig.json) | dsconfig v1 schema — single source of truth for all config fields, groups, and instructions |
| [`settings.ts`](settings.ts) | TypeScript models: `RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig` (plus nested `LogsConfig` / `TracesConfig`) |
| [`settings.go`](settings.go) | Go `Config` model (flat: jsonData fields + nested `Logs`/`Traces` + `DecryptedSecureJSONData`), `PluginID`, `SecureJsonDataKey` typed constants, and the `LoadConfig` utility |
| [`schema.go`](schema.go) | k8s-style SDK plugin schema: embeds `dsconfig.json`, converts it via `dsconfig.NewSDKSchema`, and defines `SettingsExamples` for each protocol / TLS / OTel variant |
| [`settings_test.go`](settings_test.go) | Table tests for `LoadConfig`, `ApplyDefaults`, and `Validate` |
| [`conformance_test.go`](conformance_test.go) | Runs `schema.RunPluginTests` — the shared dsconfig conformance suite — against `dsconfig.json` |
| [`schema.gen.json`](schema.gen.json), [`settings.gen.json`](settings.gen.json), [`settings.examples.gen.json`](settings.examples.gen.json) | Committed schema artifacts (regenerate with `go generate ./...`; the `SchemaArtifactInSync` conformance subtest guards drift) |

There is no per-entry `go.mod` — every registry entry is a subpackage of the shared
[`registry/`](..) module (`github.com/grafana/dsconfig/registry`).

## Sources researched

Every source below was read at the pinned upstream SHA
(`d55f9d6250023a86d49555b12fffdbeec9a1b538`), plus external editor components at the exact
versions the plugin's `package.json` pins.

### Plugin repo (`github.com/grafana/clickhouse-datasource@d55f9d6`)

| File | What was read |
| --- | --- |
| `src/plugin.json:3-5,61-65` | `pluginType` (`id`), `pluginName` (`name`), docs URL |
| `src/views/ConfigEditor.tsx:1-20` | v1 vs v2 editor split via the `newClickhouseConfigPageDesign` OpenFeature flag (we mirror the v1 storage contract, which is what the backend consumes) |
| `src/views/CHConfigEditor.tsx:37-291` | Server / TLS / Credentials / Configuration Mode / Additional settings section titles, field labels, placeholders, tooltips |
| `src/views/CHConfigEditor.tsx:263-270` | Default-port picker (native/http × insecure/secure → 9000/8123/9440/8443) |
| `src/views/CHConfigEditor.tsx:459-483` | Password `SecretInput` (secureJsonData.password) |
| `src/views/CHConfigEditor.tsx:485-529` | Configuration Mode / Signal type radios and the conditional signal-type field |
| `src/views/CHConfigEditor.tsx:614-912` | "Additional settings" collapsible: default DB/table, query settings, logs config, traces config, alias tables, row-limit, ad-hoc filters toggle, custom settings |
| `src/views/CHConfigEditor.tsx:853-862` | `enableSecureSocksProxy` gate (excluded per registry policy) |
| `src/views/CHConfigEditorHooks.ts:26-78` | `onHttpHeadersChange` — the `secureHttpHeaders.<Name>` write convention for secure headers |
| `src/views/CHConfigEditorHooks.ts:82-138` | `useConfigDefaults` — v3→v4 migration (`server`→`host`, `timeout`→`dialTimeout`), plugin `version` stamp, default protocol, and logs/traces default tables |
| `src/components/configEditor/DefaultDatabaseTableConfig.tsx:12-45` | Default DB/table field labels, descriptions, placeholders |
| `src/components/configEditor/QuerySettingsConfig.tsx:23-119` | Dial/query timeouts, max lifetime, max idle/open conns, validate SQL, enableMapKeysDiscovery toggle |
| `src/components/configEditor/LogsConfig.tsx:29-168` | Logs section — default DB/table, OTel toggle + version, column-role inputs, context-columns tag list, show-log-links toggle |
| `src/components/configEditor/TracesConfig.tsx:44-326` | Traces section — default DB/table, OTel toggle + version, column-role inputs, duration unit select, prefixes, timestamp table suffix, show-trace-links toggle |
| `src/components/configEditor/AliasTableConfig.tsx:14-181` | Alias tables (target/alias DB and table columns) |
| `src/components/configEditor/HttpHeadersConfig.tsx:17-193` | Custom HTTP headers editor (name / value / secure toggle) + forwardGrafanaHeaders |
| `src/labels.ts:1-353` | The full label / placeholder / tooltip catalog — most schema strings are copied verbatim from here |
| `src/otel.ts:1-81` | `defaultLogsTable = 'otel_logs'`, `defaultTraceTable = 'otel_traces'`, `traceTimestampTableSuffix = '_trace_id_ts'`, `OtelVersion` catalog (`1.29.0`, `latest`) |
| `src/types/config.ts:1-180` | Frontend types: `CHConfig`, `CHSecureConfig`, `CHHttpHeader`, `CHCustomSetting`, `AliasTableEntry`, `CHLogsConfig`, `CHTracesConfig`, `Protocol`, `ConfigMode`, `SignalType`, `defaultCHAdditionalSettingsConfig` |
| `pkg/plugin/settings.go:19-60` | Backend `Settings` struct — the wire shape the plugin actually reads |
| `pkg/plugin/settings.go:69-77` | `isValid()` — hard-fails on missing host or port (encoded as `requiredWhen: "true"` in the schema) |
| `pkg/plugin/settings.go:80-317` | `LoadSettings` — v3 legacy fallbacks (`server`, `timeout`), string-or-number tolerance, defaulting of timeouts, secret copies, EnableSchemaCache / SchemaCacheTTLSeconds defaults |
| `pkg/plugin/settings.go:319-344` | `loadHttpHeaders` — `secureHttpHeaders.<Name>` prefix scan of secureJsonData |
| `package.json`, `go.mod` | External component versions (see next table) |

### External editor components

Read at the exact versions pinned in the plugin's `package.json` / `go.mod`.

| Component | Version | What was read |
| --- | --- | --- |
| `DataSourceDescription`, `ConfigSection`, `ConfigSubSection` | vendored in `src/components/experimental/ConfigSection.tsx` at `d55f9d6` (the plugin ships its own copy under `experimental/`) | Section title / description shape, `isCollapsible` behaviour |
| `RadioButtonGroup`, `Switch`, `Input`, `SecretInput`, `Field`, `Stack`, `Alert`, `Button` | `@grafana/ui@12.4.2` | Prop names (`label`, `placeholder`, `value`, `onChange`, `isConfigured`, `onReset`) — dictate which UI attributes we record |
| `DataSourcePluginOptionsEditorProps`, `onUpdateDatasourceJsonDataOption`, `onUpdateDatasourceSecureJsonDataOption`, `KeyValue` | `@grafana/data@12.4.2` | Storage-key semantics of the update helpers |
| `OpenFeature` client | `@openfeature/web-sdk@1.6.4` (dev toggle gate — behind `newClickhouseConfigPageDesign`) | Confirms the v2 editor is behind a feature flag; v1 is the storage contract we model |
| `github.com/grafana/grafana-plugin-sdk-go` `v0.292.1` | pinned `go.mod` | `backend.DataSourceInstanceSettings`, `backend.Logger.FromContext`, `proxy.ProxyOptions`, `sdkconfig.GrafanaConfigFromContext` |
| `github.com/ClickHouse/clickhouse-go/v2` (used by the driver, not our config) | pinned `go.mod` | Consumer of the settings — `Protocol.HTTP.String()` returns `"http"` (`pkg/plugin/settings.go:289`) |

## Field provenance

Every dsconfig field, traced from its schema `id` to the upstream `file:line` where each of its
label, placeholder, description, default, storage key, and value type is defined. Where a field
draws from multiple lines the primary source is listed; the full source cross-reference is in the
per-source table above.

### Connection fields

| Schema `id` | Storage key | Target | Label source | Placeholder / default source | Value type source |
| --- | --- | --- | --- | --- | --- |
| `jsonData_host` | `host` | `jsonData` | `labels.ts:8` (`Server address`) | `labels.ts:9` (`Server address`) | `Settings.Host string`, `pkg/plugin/settings.go:20` |
| `jsonData_port` | `port` | `jsonData` | `labels.ts:14` (`Server port`) | Port picker `CHConfigEditor.tsx:263-270`; default varies by protocol/secure | `Settings.Port int64`, `pkg/plugin/settings.go:21` (also string-tolerant, `:94-103`) |
| `jsonData_protocol` | `protocol` | `jsonData` | `labels.ts:28` (`Protocol`) | Options `CHConfigEditor.tsx:54-57`; default `native`, `CHConfigEditorHooks.ts:112-114` | `Settings.Protocol string`, `pkg/plugin/settings.go:22`; TS `Protocol`, `types/config.ts:163-166` |
| `jsonData_secure` | `secure` | `jsonData` | `labels.ts:66` (`Secure Connection`) | Default `false` | `Settings.Secure bool`, `pkg/plugin/settings.go:23` |
| `jsonData_path` | `path` | `jsonData` | `labels.ts:23` (`HTTP URL Path`) | `labels.ts:25` (`additional-path`) | `Settings.Path string`, `pkg/plugin/settings.go:24` |

### Credentials

| Schema `id` | Storage key | Target | Source |
| --- | --- | --- | --- |
| `jsonData_username` | `username` | `jsonData` | Label `labels.ts:32`, placeholder `labels.ts:33` (`default`), tooltip `labels.ts:34` |
| `secureJsonData_password` | `password` | `secureJsonData` | Label `labels.ts:37`, placeholder `labels.ts:38` (`password`), backend copy `pkg/plugin/settings.go:272-275` |

### TLS

| Schema `id` | Storage key | Target | Source |
| --- | --- | --- | --- |
| `jsonData_tlsSkipVerify` | `tlsSkipVerify` | `jsonData` | `labels.ts:41-44`; `Settings.InsecureSkipVerify`, `pkg/plugin/settings.go:26` |
| `jsonData_tlsAuth` | `tlsAuth` | `jsonData` | `labels.ts:45-48`; `Settings.TlsClientAuth`, `pkg/plugin/settings.go:27` |
| `jsonData_tlsAuthWithCACert` | `tlsAuthWithCACert` | `jsonData` | `labels.ts:49-52`; `Settings.TlsAuthWithCACert`, `pkg/plugin/settings.go:28` |
| `secureJsonData_tlsCACert` | `tlsCACert` | `secureJsonData` | `labels.ts:53-56`; backend copy `pkg/plugin/settings.go:276-279`; `dependsOn tlsAuthWithCACert==true` from render `CHConfigEditor.tsx:429-437` |
| `secureJsonData_tlsClientCert` | `tlsClientCert` | `secureJsonData` | `labels.ts:57-60`; backend copy `pkg/plugin/settings.go:280-283` |
| `secureJsonData_tlsClientKey` | `tlsClientKey` | `secureJsonData` | `labels.ts:61-64`; backend copy `pkg/plugin/settings.go:284-287` |

### Configuration mode & signal type

| Schema `id` | Storage key | Target | Source |
| --- | --- | --- | --- |
| `jsonData_configMode` | `configMode` | `jsonData` | Label / description `CHConfigEditor.tsx:487-490` (`Choose how this datasource is used…`); options `CHConfigEditor.tsx:493-496` (`All databases` / `Single table`); default `classic`; type `ConfigMode`, `types/config.ts:13` |
| `jsonData_signalType` | `signalType` | `jsonData` | Label `CHConfigEditor.tsx:510` (`Signal type`); options `CHConfigEditor.tsx:512-514` with descriptions; type `SignalType`, `types/config.ts:5` |

### Default DB / table

| Schema `id` | Storage key | Target | Source |
| --- | --- | --- | --- |
| `jsonData_defaultDatabase` | `defaultDatabase` | `jsonData` | Label `labels.ts:120`, description `labels.ts:121`, placeholder `labels.ts:123` (`default`) |
| `jsonData_defaultTable` | `defaultTable` | `jsonData` | Label `labels.ts:126`, description `labels.ts:127`, placeholder `labels.ts:129` (`table`) |

### Query settings

| Schema `id` | Storage key | Target | Source |
| --- | --- | --- | --- |
| `jsonData_dialTimeout` | `dialTimeout` | `jsonData` | `labels.ts:141-145`; backend default `10`, `pkg/plugin/settings.go:255-257` |
| `jsonData_queryTimeout` | `queryTimeout` | `jsonData` | `labels.ts:158-163`; backend default `60`, `:258-260` |
| `jsonData_connMaxLifetime` | `connMaxLifetime` | `jsonData` | `labels.ts:134-139`; backend default `5`, `:261-263` |
| `jsonData_maxIdleConns` | `maxIdleConns` | `jsonData` | `labels.ts:146-151`; backend default `25`, `:264-266` |
| `jsonData_maxOpenConns` | `maxOpenConns` | `jsonData` | `labels.ts:152-157`; backend default `50`, `:267-269` |
| `jsonData_validateSql` | `validateSql` | `jsonData` | `labels.ts:164-167`; frontend-only |
| `jsonData_enableMapKeysDiscovery` | `enableMapKeysDiscovery` | `jsonData` | `labels.ts:168-174`; editor default `true` (from the `??` in `QuerySettingsConfig.tsx:113`) |

### Logs configuration (section `logs`)

All 11 fields under `jsonData.logs.*` mirror `CHLogsConfig` in `types/config.ts:101-116`. Labels
and descriptions come from `labels.ts:288-353`; the default `logs.defaultTable = 'otel_logs'` and
`selectContextColumns = true` / `contextColumns = []` come from `useConfigDefaults`
(`CHConfigEditorHooks.ts:116-123`) and `defaultCHAdditionalSettingsConfig` (`types/config.ts:168-179`).

### Traces configuration (section `traces`)

All 25 fields under `jsonData.traces.*` mirror `CHTracesConfig` in `types/config.ts:118-154`.
Labels and descriptions come from `labels.ts:176-287`. Defaults: `traces.defaultTable = 'otel_traces'`
(`useConfigDefaults`), `traces.durationUnit = 'nanoseconds'` (`onTracesConfigChange`,
`CHConfigEditor.tsx:225`). The `traceTimestampTableSuffix` placeholder defaults to `_trace_id_ts`
via `otel.ts:6`.

### Arrays

| Schema `id` | Storage key | Target | Source |
| --- | --- | --- | --- |
| `jsonData_httpHeaders` | `httpHeaders` | `jsonData` | `HttpHeadersConfig.tsx:45-81` + `CHConfigEditorHooks.ts:26-78`; each element = `CHHttpHeader { name, value, secure }` (`types/config.ts:90-94`); conditional on `protocol == 'http'` from render `CHConfigEditor.tsx:381` |
| `jsonData_customSettings` | `customSettings` | `jsonData` | `CHConfigEditor.tsx:863-909` (Custom Settings sub-section); each element = `{ setting, value }` (`pkg/plugin/settings.go:62-65`) |
| `jsonData_aliasTables` | `aliasTables` | `jsonData` | `AliasTableConfig.tsx:69-103`; each element = `AliasTableEntry` (`types/config.ts:156-161`) |

### Other

| Schema `id` | Storage key | Target | Source |
| --- | --- | --- | --- |
| `jsonData_forwardGrafanaHeaders` | `forwardGrafanaHeaders` | `jsonData` | `labels.ts:96-99`; consumed at `pkg/plugin/settings.go:201-210`; conditional on `protocol == 'http'` |
| `jsonData_enableRowLimit` | `enableRowLimit` | `jsonData` | `labels.ts:73-78`; consumed at `pkg/plugin/settings.go:212-221,306-314` |
| `jsonData_hideTableNameInAdhocFilters` | `hideTableNameInAdhocFilters` | `jsonData` | `labels.ts:79-84`; frontend-only |
| `jsonData_version` | `version` | `jsonData` | Frontend-only stamp written by `CHConfigEditorHooks.ts:94` |
| `jsonData_enableSchemaCache` | `enableSchemaCache` | `jsonData` | Backend-only; `Settings.EnableSchemaCache`, `pkg/plugin/settings.go:52-55`; defaults `true` at `:225-237` |
| `jsonData_schemaCacheTTLSeconds` | `schemaCacheTTLSeconds` | `jsonData` | Backend-only; `Settings.SchemaCacheTTLSeconds`, `:56-59`; defaults `60` at `:238-252` |

## Frontend-only settings

These are written by the editor into jsonData but never read by the backend:

- **`configMode`, `signalType`** — pure UI-mode switches for the v1 config editor and query
  editor. The backend has no knowledge of them (grep `pkg/plugin/settings.go` — no matching keys).
- **`defaultTable`** — the classic-mode "Default table" input. The backend does not consume
  a top-level `defaultTable`; it uses `defaultDatabase` and the query-builder-supplied table.
- **`validateSql`** — SQL-validation toggle for the Monaco editor. No backend equivalent.
- **`hideTableNameInAdhocFilters`** — column-picker cosmetic toggle. No backend consumer.
- **`version`** — the plugin version stamp written by `useConfigDefaults` on every save.

## Backend-only settings

These are read by the backend but not exposed in the config editor UI:

- **`enableSchemaCache`** — gates the in-process schema-introspection cache
  (`pkg/plugin/settings.go:52-55,225-237`). Defaults `true` when unset.
- **`schemaCacheTTLSeconds`** — cache freshness in seconds
  (`pkg/plugin/settings.go:56-59,238-252`). Defaults `60` when unset or `<=0`.
- Also: `pdcInjected` is stored by Grafana core when the datasource is behind a Private Data
  Connect (PDC) tunnel; it is not user-editable and is not modelled here.

## Where the types are defined

### Frontend (TypeScript)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `CHConfig`, `CHSecureConfig`, `CHHttpHeader`, `CHCustomSetting`, `AliasTableEntry`, `CHLogsConfig`, `CHTracesConfig`, `Protocol` (enum), `ConfigMode`, `SignalType`, `defaultCHAdditionalSettingsConfig` | `src/types/config.ts:1-180` | plugin ([grafana/clickhouse-datasource](https://github.com/grafana/clickhouse-datasource)) |
| `DataSourceJsonData` (base interface `CHConfig extends`), `KeyValue`, `DataSourceSettings`, `DataSourcePluginOptionsEditorProps`, `onUpdateDatasourceJsonDataOption`, `onUpdateDatasourceSecureJsonDataOption` | `packages/grafana-data/src/` | `@grafana/data` `12.4.2` (grafana/grafana `v12.4.2`) |
| `SecureSocksProxySettings` / `enableSecureSocksProxy` jsonData field (excluded from this entry) | `packages/grafana-ui/src/components/DataSourceSettings/SecureSocksProxySettings.tsx` | `@grafana/ui` `12.4.2` |
| `RadioButtonGroup`, `Switch`, `Input`, `SecretInput`, `Field`, `Stack`, `Button`, `TagsInput` (editor UI, no storage fields) | `packages/grafana-ui/src/components/` | `@grafana/ui` `12.4.2` |
| `ConfigSection`, `ConfigSubSection`, `DataSourceDescription` (layout / intro, no storage fields) | `src/components/experimental/ConfigSection.tsx` | vendored inside the plugin at `d55f9d6` |
| `OpenFeature` client (feature-flag gate for the v2 editor, not a storage field) | `@openfeature/web-sdk@1.6.4` | via `src/views/ConfigEditor.tsx:3` |

### Backend (Go)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `Settings` (jsonData + decrypted secrets), `CustomSetting`, `LoadSettings`, `loadHttpHeaders`, `secureHeaderKeyPrefix` | `pkg/plugin/settings.go:19-345` | plugin ([grafana/clickhouse-datasource](https://github.com/grafana/clickhouse-datasource)) |
| `backend.DataSourceInstanceSettings`, `backend.Logger.FromContext`, `backend.DownstreamError` | `backend/` | `github.com/grafana/grafana-plugin-sdk-go` `v0.292.1` |
| `proxy.Options`, `ProxyOptionsFromContext` | `backend/proxy` | `github.com/grafana/grafana-plugin-sdk-go` `v0.292.1` |
| `sdkconfig.GrafanaConfigFromContext`, `cfg.SQL()` — used to pull the Grafana-wide row-limit when `enableRowLimit=true` | `config/` | `github.com/grafana/grafana-plugin-sdk-go` `v0.292.1` |
| `clickhouse.HTTP` (protocol constant used to gate `loadHttpHeaders`) | — | `github.com/ClickHouse/clickhouse-go/v2` (consumer of settings, not a settings source) |

The models in this entry flatten the frontend + backend split into a single Go `Config` struct
(jsonData fields + nested `Logs`/`Traces` + `DecryptedSecureJSONData` + `SecureHttpHeaders`) plus
a `SecureJsonDataKey` typed constant list. `settings.ts` keeps the three canonical TypeScript
types (`RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig`).

## Modeling decisions

- **Nested logs / traces via `section:`**: `jsonData.logs.*` and `jsonData.traces.*` are modeled
  as individual schema fields with `section: "logs"` / `section: "traces"`. This gives editor
  parity for every column-role field while letting the conformance suite descend into the nested
  `Logs *LogsConfig` / `Traces *TracesConfig` struct fields.
- **Timeouts as strings**: `dialTimeout`, `queryTimeout`, `connMaxLifetime`, `maxIdleConns`, and
  `maxOpenConns` are typed `string` in the schema and the Go struct, matching upstream
  (`pkg/plugin/settings.go:38-42`). The custom `Config.UnmarshalJSON` also tolerates JSON numbers
  as a legacy shape, mirroring `LoadSettings`'s number/string branching.
- **Legacy v3 fields**: `server` and numeric `timeout` are absorbed by `Config.UnmarshalJSON`
  into `Host` and `DialTimeout` — matching `LoadSettings` at `pkg/plugin/settings.go:87-92,160-168`.
  New configurations should write `host` and `dialTimeout` directly.
- **Secure HTTP headers**: dynamic `secureHttpHeaders.<Header Name>` keys are not enumerated in
  `SecureJsonDataKeys`; instead `LoadConfig` prefix-scans `DecryptedSecureJSONData` into a
  `SecureHttpHeaders map[string]string` on `Config`. Instructions document the convention.
- **`requiredWhen` vs the editor**: the editor renders host and port with a `required` mark and
  inline error (`CHConfigEditor.tsx:303-339`); the backend also hard-fails on missing host / port
  (`pkg/plugin/settings.go:69-77`). Both sides are recorded (`requiredWhen: "true"` on both
  fields, plus explicit `Validate` checks in Go).
- **Protocol-conditional fields**: `path`, `httpHeaders`, and `forwardGrafanaHeaders` are gated
  on `jsonData_protocol == 'http'`, mirroring the editor's conditional renders at
  `CHConfigEditor.tsx:366-391` and the backend's HTTP-only headers load at
  `pkg/plugin/settings.go:289-291`.
- **TLS toggles are independent**: `tlsAuth` and `tlsAuthWithCACert` are separate switches, each
  gating its own set of secrets. Neither implies the other. Both may be enabled simultaneously —
  the schema mirrors this by using `dependsOn` on the individual secret fields rather than a
  discriminator.
- **Secure Socks Proxy excluded**: `jsonData.enableSecureSocksProxy` (`CHConfigEditor.tsx:853-862`)
  is deliberately omitted from this registry entry.
- **v1 vs v2 editor**: an OpenFeature flag (`newClickhouseConfigPageDesign`, `ConfigEditor.tsx:16`)
  chooses between v1 (`CHConfigEditor.tsx`) and v2 (`config-v2/CHConfigEditor.tsx`). The v2 editor
  reads and writes the same underlying storage shape — we model the v1 editor because it is the
  current default and its labels are the ones the backend documentation refers to.

## Settings examples matrix

`SettingsExamples()` (in `schema.go`) provides the default configuration plus one k8s-style
example per major protocol / TLS / OTel / single-table variant, plus a legacy v3-storage-shape
example. Each example is a full instance-settings object with the plugin configuration nested
under `jsonData` and the relevant write-only secrets under `secureJsonData` (placeholder values —
replace with real secrets). The default example (keyed `""`) carries only the schema defaults with
an empty `password` placeholder.

| Example key | Protocol | Secure | Config mode | `secureJsonData` |
| --- | --- | --- | --- | --- |
| `""` (default) | native | false | classic | `password` (empty) |
| `nativeInsecure` | native | false | classic | `password` |
| `nativeSecure` | native | true | classic | `password` |
| `httpSecureWithHeaders` | http | true | classic | `password` + `secureHttpHeaders.X-Api-Key` |
| `tlsClientAuth` | native | true | classic | `password` + `tlsClientCert` + `tlsClientKey` |
| `tlsWithCACert` | native | true | classic | `password` + `tlsCACert` |
| `otelLogsSingleTable` | native | true | single-table (logs) | `password` |
| `otelTracesSingleTable` | native | true | single-table (traces) | `password` |
| `legacyV3ServerField` | native (via legacy `server` / numeric `timeout`) | false | classic | `password` |

## `LoadConfig` utility (`settings.go`)

`LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error)`
runs the full three-phase load flow on a datasource instance's settings and returns a
fully-defaulted, validated `Config`:

1. **Parse** — unmarshal jsonData into `Config`. The custom `Config.UnmarshalJSON` mirrors
   `LoadSettings` at `pkg/plugin/settings.go:87-186` verbatim: v3 `server`/`timeout` fallbacks,
   string-or-number tolerance for `port`, `dialTimeout`, and `queryTimeout`. Copy static
   decrypted secrets into `DecryptedSecureJSONData` and strip `secureHttpHeaders.<Name>` keys
   into `SecureHttpHeaders`.
2. **`ApplyDefaults`** — fill a curated set of zero-valued fields with the same defaults the
   editor and backend write for a fresh datasource: protocol=native, configMode=classic,
   enableMapKeysDiscovery=true, dial/query/conn timeouts (10/60/5/25/50), enableSchemaCache=true,
   schemaCacheTTLSeconds=60, logs.defaultTable=`otel_logs`, traces.defaultTable=`otel_traces`,
   traces.durationUnit=`nanoseconds`.
3. **`Validate`** — enforce the runtime contract: host + port required (mirrors `isValid`
   at `pkg/plugin/settings.go:69-77`), protocol must be one of `native` / `http`, TLS toggles
   require their secrets, and `configMode='single-table'` requires a signalType. Errors are
   joined so every problem surfaces at once.

Everything is logged via `backend.Logger.FromContext(ctx)` with `datasource_uid`,
`datasource_name`, and `plugin` labels so log lines carry request context.

### Direct access to individual phases

`(*Config).ApplyDefaults()` and `(Config).Validate() error` are still exported for callers that
want to compose them themselves (provisioning preview, schema-example round-trip, tests that need
to distinguish parse-level from policy-level errors). Skip them by never calling `LoadConfig`
in those flows — assemble a `Config` directly.

## Upstream findings

Potential quirks, misleading UX, and inconsistencies discovered while researching upstream. All
preserved verbatim in the schema — the schema records what the plugin **does**, not what it
**should** do; these notes exist so reviewers can reproduce each finding and decide separately
whether to fix upstream.

1. **`Secure Connection` and TLS toggles are semantically overlapping but independent.**
   `CHConfigEditor.tsx:352-364` — `jsonData.secure` toggles the driver's TLS handshake and picks
   the default port. `jsonData.tlsAuth` / `tlsAuthWithCACert` set certificate policy on top of
   that. It is possible to check `tlsAuth` without `secure` — the plugin will still open a plain
   TCP connection with no TLS. There is no editor-side interlock.
2. **`Port` field has no true default.** The editor shows a placeholder derived from
   protocol×secure (`CHConfigEditor.tsx:263-270`) but does not populate `jsonData.port` until the
   user types. `isValid()` at `pkg/plugin/settings.go:73` rejects `port == 0`, so a datasource
   saved with the placeholder only will fail health checks — the editor's inline error
   (`CHConfigEditor.tsx:106-107`) is the only guard before Save.
3. **Timeouts stored as strings.** `pkg/plugin/settings.go:38-42` — `DialTimeout`, `QueryTimeout`,
   `ConnMaxLifetime`, `MaxIdleConns`, `MaxOpenConns` are `string`. `LoadSettings` (`:160-186`)
   also tolerates JSON numbers as a legacy shape; the editor writes strings (`<Input type="number">`
   returns a string via `onUpdateDatasourceJsonDataOption`). The mixed shape means provisioning
   YAML can use either form, but internal callers must always emit strings.
4. **`enableMapKeysDiscovery` default is asymmetric.** The editor uses `??` to default it to `true`
   at render (`QuerySettingsConfig.tsx:113`) but the storage key is only written when the user
   toggles it. Backend `LoadSettings` has no equivalent default, so unset means false in
   `Settings.EnableMapKeysDiscovery` (there is no such field on `Settings`, in fact — the flag is
   consumed purely by the frontend query builder). The schema mirrors the editor default of `true`.
5. **`traces.durationUnit` written even when not shown.** `onTracesConfigChange`
   (`CHConfigEditor.tsx:218-230`) unconditionally back-fills `durationUnit = 'nanoseconds'` on
   every change to any traces field, even fields unrelated to duration. Round-tripping a config
   through the editor will therefore always leave `traces.durationUnit` set.
6. **`server` / `timeout` are silently migrated on load, not persisted back.** The editor's
   `useConfigDefaults` hook (`CHConfigEditorHooks.ts:96-108`) reads them into `host` / `dialTimeout`
   and deletes the source keys, but only on first render; provisioned configs that never open the
   editor keep the legacy keys forever. The backend accepts either shape (`pkg/plugin/settings.go:87-92,160-168`).
7. **Secure custom HTTP headers hide the plaintext header name.** The `secureHttpHeaders.<Header Name>`
   convention puts the header **name** into a secret key (`CHConfigEditorHooks.ts:42-49`); if a
   header is later renamed, the old key remains in `secureJsonFields` and must be explicitly
   cleared by writing an empty value under the old key (`:54-60`).
8. **`enableSchemaCache` cannot actually be enabled — it's always on unless explicitly disabled.**
   `pkg/plugin/settings.go:225-237` sets `settings.EnableSchemaCache = true` before checking for
   a stored value; the stored value can only override it to `false`. This is by design (cache-on
   by default) but is not documented in the editor.
9. **`configMode` and `signalType` are frontend-only, but re-reading them requires the
   config editor's `useConfigDefaults` hook.** The backend does not know about them; provisioning
   configs that set `configMode: 'single-table'` still have to seed a compatible logs/traces
   nested block for the query builder to produce useful queries.
10. **`enableRowLimit` requires Grafana >= 11.0.0.** `pkg/plugin/settings.go:306-314` calls
    `sdkconfig.GrafanaConfigFromContext(ctx).SQL()` which is only reliable on 11+. Older Grafanas
    will return a zero `RowLimit` even when the toggle is on.
11. **Row-limit knob (`rowLimit`) is Grafana-populated, not user-set.** `Settings.RowLimit int64`
    at `pkg/plugin/settings.go:49` has a json tag, but the backend overwrites it from the Grafana
    context after loading (`:306-313`). It is neither in the schema nor in the settings model —
    modeling it would suggest end-users can configure it, which is not the case.
12. **Alias-tables editor dedupes silently.** `AliasTableConfig.tsx:23-33` uses a stringified key
    including database names to dedupe rows; two rows for the same (targetTable, aliasTable) pair
    with different (empty) database fields are considered duplicates and one is silently dropped.

## Validation performed

- `dsconfig.ParseAndResolveSchemaJSON` + `Schema.Validate()` (Go validator in this repo) — passes.
- JSON Schema validation against [`dsconfig/schema.json`](../../dsconfig/schema.json) (draft-07,
  `additionalProperties: false`) — passes.
- `go test ./...` on this module — passes (schema bundle shape, secure values, examples,
  `LoadConfig` incl. legacy fallback and string/number port tolerance, `SchemaArtifactInSync`
  guard, TLS/config-mode validation branches).
- `settings.go`/`schema.go`: `go build ./...`, `go vet ./...`, `gofmt -l .` — all clean.
- The pre-existing `dsconfig` and `schema` workspace modules and every sibling registry entry
  still build and pass their tests (full `go test ./...` inside `registry/` is clean).
