# tempo

Declarative configuration schema for the [Tempo datasource plugin](https://github.com/grafana/grafana-tempo-datasource) (`tempo`).

## Upstream researched

- **Repo**: `github.com/grafana/grafana-tempo-datasource`
- **Ref**: `main`
- **Commit SHA**: `485876240cd54aecac9a1f02eca8ec1d55dd2137`

Every value in [`dsconfig.json`](dsconfig.json) — labels, placeholders,
tooltips, option labels/values, section titles, help markdown, defaults,
validations, dependency and required-when expressions, storage keys, storage
targets, value types, group titles, and instructions — is traceable to a
specific `file:line` in either the plugin repo (at this SHA) or the pinned
external editor libraries. See [Field provenance](#field-provenance) below.

To reproduce this research:

```bash
git clone https://github.com/grafana/grafana-tempo-datasource
cd grafana-tempo-datasource
git checkout 485876240cd54aecac9a1f02eca8ec1d55dd2137
```

If upstream `main` has advanced past this SHA, re-diff the sources listed under
[Sources researched](#sources-researched) before merging any changes to this
entry.

## Files

| File | Purpose |
| --- | --- |
| [`dsconfig.json`](dsconfig.json) | dsconfig v1 schema — single source of truth for all config fields, groups, effects, and instructions |
| [`settings.ts`](settings.ts) | TypeScript models: `RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig`, plus the nested config types |
| [`settings.go`](settings.go) | Go `Config` model (flat: root URL/basicAuth/basicAuthUser/withCredentials tagged `json:"-"`, jsonData fields as value-typed nested structs, `DecryptedSecureJSONData`), `PluginID`, typed enums (`SpanBarType`, `TraceqlSearchScope`, `SecureJsonDataKey`, `TimeRangeForTags*` constants), a lenient `UnmarshalJSON` for `tagLimit`, and the `LoadConfig` utility |
| [`schema.go`](schema.go) | k8s-style SDK plugin schema: embeds `dsconfig.json`, converts it via `dsconfig.NewSDKSchema`, and defines `SettingsExamples` for each auth/TLS/feature variant |
| [`settings_test.go`](settings_test.go) | Table tests for `LoadConfig`, `ApplyDefaults`, and `Validate` |
| [`conformance_test.go`](conformance_test.go) | Runs `schema.RunPluginTests` — the shared dsconfig conformance suite — against `dsconfig.json` |
| [`schema.gen.json`](schema.gen.json), [`settings.gen.json`](settings.gen.json), [`settings.examples.gen.json`](settings.examples.gen.json) | Committed schema artifacts (regenerate with `go generate ./...` inside this directory; the `SchemaArtifactInSync` conformance subtest guards drift) |

There is no per-entry `go.mod` — every registry entry is a subpackage of the
shared [`registry/`](..) module (`github.com/grafana/dsconfig/registry`).

## Sources researched

Every source below was read at the pinned upstream SHA (`4858762`), plus
external editor components from the `@grafana/o11y-ds-frontend` and
`@grafana/plugin-ui` packages at the versions the plugin's `package.json`
pins.

### Plugin repo (`github.com/grafana/grafana-tempo-datasource@4858762`)

| File | What was read |
| --- | --- |
| `src/plugin.json:1-50` | `pluginType` (`id` = `"tempo"`), `pluginName` (`name` = `"Tempo"`), docs URL (`info.links[2].url` = `"https://grafana.com/docs/grafana/latest/datasources/tempo/"`) |
| `src/configuration/ConfigEditor.tsx:33-141` | Top-level editor — composes `DataSourceDescription`, `ConnectionSettings` (with `urlPlaceholder="http://localhost:3200"`), `Auth` (via `convertLegacyAuthProps`), `StreamingSection` (top-level), `TraceToLogsSection` / `TraceToMetricsSection` / `TraceToProfilesSection` (top-level), and a collapsible "Additional settings" section containing `AdvancedHttpSettings`, `SecureSocksProxySettings` (excluded), `ServiceGraphSettings`, `NodeGraphSection`, `TraceQLSearchSettings`, `QuerySettings`, `TagsTimeRangeSettings`, `TagLimitSection`, `SpanBarSection` |
| `src/configuration/StreamingSection.tsx:13-80` | `ConfigSection title="Streaming"`, `Alert severity="info" title="Streaming and self-managed Tempo instances"`; two `InlineSwitch`es labelled "Search queries" / "Metrics queries", each tooltip `` `Enable streaming for X queries. Minimum required version for Tempo: ${featuresToTempoVersion[FeatureName.searchStreaming\|metricsStreaming]}.` `` (`featuresToTempoVersion` at `src/datasource.ts:79-82` = search 2.2.0, metrics 2.7.0) |
| `src/configuration/ServiceGraphSettings.tsx:15-56` | `InlineField label="Data source" tooltip="The Prometheus data source with the service graph data"`, `DataSourcePicker` filtered to `pluginId: 'prometheus'`; writes `serviceMap.datasourceUid` |
| `src/configuration/QuerySettings.tsx:15-74` | `InlineField label="Use time range in query"` tooltip (`:29-30`), IntervalInput labels `"Time shift for start of search"`/`"Time shift for end of search"` (`getLabel`), tooltip `"Shifts the ${type} of the time range when searching by TraceID. Searching can return traces that do not fully fall into the search time range, so we recommend using higher time shifts for longer traces. Default: 30m (Time units can be used here, for example: 5s, 1m, 3h"` (`getTooltip:22-24`); writes `traceQuery.{timeShiftEnabled, spanStartTimeShift, spanEndTimeShift}` |
| `src/configuration/TagsTimeRangeSettings.tsx:10-45` | `InlineField label="Time range in query" tooltip="Time range in tags and tag value queries"`, `Combobox` with placeholder `"Time range for tags"` and five options (1800/10800/86400/259200/604800 seconds); default `DEFAULT_TIME_RANGE_FOR_TAGS = 1800` |
| `src/configuration/TagLimitSettings.tsx:12-59` | `ConfigSubSection title="Tag limit"`, description via `ConfigDescriptionLink` (`:47-52`); `InlineField label="Max tags and tag values"` tooltip `"Specify the max number of tags and tag values to display in the Tempo editor. Default: 5000"`; `Input type="number" placeholder="5000"` bound to `jsonData.tagLimit` — writes `v.currentTarget.value` (string) |
| `src/configuration/TraceQLSearchSettings.tsx:16-48` | `InlineField label="Hide search" tooltip="Removes the search tab from the query editor"`, `InlineField label="Static filters" tooltip="Configures which fields are available in the UI"`; writes `search.hide` and delegates `search.filters` to `TraceQLSearchTags` |
| `src/dataquery.ts:105-142` | `TraceqlSearchScope` enum and `TraceqlFilter` shape |
| `src/types.ts:6-29` | Frontend `TempoJsonData extends DataSourceJsonData` — the declared shape (some fields intentionally use out-of-date/older sub-shapes, see [Upstream findings](#upstream-findings)) |
| `src/datasource.ts:70-159` | `FeatureName`, `featuresToTempoVersion` (streaming version gates); constructor seeds `search.filters` with two default entries (`service.name` in `resource`/`span.name` in `span`) on first load |
| `pkg/tempo/tempo.go:52-90` | `NewDatasource` — reads `settings.URL` directly, calls `settings.HTTPClientOptions(ctx)` and `newGrpcClient`; forces `opts.ForwardHTTPHeaders = true` |
| `pkg/tempo/tempo.go:149-253` | `CheckHealth` — decodes `settings.JSONData` as `map[string]interface{}`, probes gRPC streaming when `streamingEnabled.search` is `true`, otherwise `GET /api/echo` on the HTTP client |
| `pkg/tempo/tempo.go:255-377` | Resource proxy — reads `dsInfo.URL` and proxies `/tags` and `/tag-values` requests to Tempo |
| `pkg/tempo/grpc.go:85-137` | `newGrpcClient` — parses `settings.URL`, appends `:80` (http) or `:443` (https) if no port, honors `settings.BasicAuthEnabled` (per-RPC creds) and secure socks proxy (`settings.ProxyClient`) |
| `pkg/tempo/grpc.go:178-197` | Basic auth over gRPC (`basicAuth` credentials) and TLS transport wiring (`credentials.NewTLS(tls)`) |
| `package.json:26-33` | External component versions (see next table) |

Notably absent: no `pkg/models/settings.go`, no upstream typed `LoadSettings`
— the Tempo plugin does not own a backend jsonData settings model. Server-side
reads use `settings.URL` directly and one ad-hoc `map[string]interface{}` unmarshal
in `CheckHealth`. Everything else is delegated to the SDK.

### External editor components

Read at the versions the plugin's `package.json` pins.

| Component | Version | Source consulted | What was read |
| --- | --- | --- | --- |
| `ConnectionSettings` | `@grafana/plugin-ui@0.15.1` | `github.com/grafana/plugin-ui/blob/main/src/components/ConfigEditor/Connection/ConnectionSettings.tsx` | URL label defaults to `"URL"`, placeholder passed by plugin (`ConfigEditor.tsx:45` — `urlPlaceholder="http://localhost:3200"`); required + built-in URL regex validation |
| `Auth`, `AuthMethodSettings`, `BasicAuth` | `@grafana/plugin-ui@0.15.1` | `src/components/ConfigEditor/Auth/{Auth,auth-method/AuthMethodSettings,auth-method/BasicAuth}.tsx` | Default `visibleMethods = [BasicAuth, OAuthForward, NoAuth]`; option labels `Basic authentication` / `Forward OAuth Identity` / `No Authentication`; BasicAuth `User`/`Password` labels + placeholders + tooltips |
| `convertLegacyAuthProps` | `@grafana/plugin-ui@0.15.1` | `src/components/ConfigEditor/Auth/utils.ts` | Maps `basicAuth` / `withCredentials` / `jsonData.oauthPassThru` ↔ AuthMethod enum; `onAuthMethodSelect` writes basicAuth+withCredentials+oauthPassThru in one shot |
| TLS pack (`SelfSignedCertificate`, `TLSClientAuth`, `SkipTLSVerification`) | `@grafana/plugin-ui@0.15.1` | `src/components/ConfigEditor/Auth/tls/*.tsx` | All TLS labels/placeholders/rows come verbatim from these files (see [Field provenance](#field-provenance)); note the RSA private key placeholder typo `Begins with --- RSA PRIVATE KEY CERTIFICATE ---` (shared across plugins) |
| `CustomHeaders`, `CustomHeader` | `@grafana/plugin-ui@0.15.1` | `src/components/ConfigEditor/Auth/custom-headers/*.tsx` | Indexed `httpHeaderName<N>` / `httpHeaderValue<N>` storage pattern; **not modeled** in this schema (see [Modeling decisions](#modeling-decisions)) |
| `AdvancedHttpSettings` | `@grafana/plugin-ui@0.15.1` | `src/components/ConfigEditor/AdvancedSettings/AdvancedHttpSettings.tsx` | `Allowed cookies` and `Timeout` labels/tooltips/placeholders |
| `DataSourceDescription`, `ConfigSection`, `ConfigSubSection`, `ConfigDescriptionLink` | `@grafana/plugin-ui@0.15.1` | `src/components/ConfigEditor/*` | Section title/description props (no storage keys — layout only) |
| `TraceToLogsSection`, `TraceToLogsSettings` (writes `tracesToLogsV2`; migrates legacy `tracesToLogs` via `getTraceToLogsOptions`) | `@grafana/o11y-ds-frontend@13.1.0-canary` (`13.1.0-27316398859`) | `github.com/grafana/grafana/blob/main/packages/grafana-o11y-ds-frontend/src/TraceToLogs/TraceToLogsSettings.tsx` | Section title `"Trace to logs"`; supported target datasource types (`loki`, `elasticsearch`, `grafana-splunk-datasource`, `grafana-opensearch-datasource`, `grafana-falconlogscale-datasource`, `googlecloud-logging-datasource`, `victoriametrics-logs-datasource`); v1→v2 migration in `getTraceToLogsOptions:55-73`; field labels (`Data source`, `Span start time shift`, `Span end time shift`, `Tags`, `Filter by trace ID`, `Filter by span ID`, `Use custom query`, `Query`); default tags list in tooltip |
| `TraceToMetricsSection`, `TraceToMetricsSettings` (writes `tracesToMetrics`) | `@grafana/o11y-ds-frontend@13.1.0-canary` | `packages/grafana-o11y-ds-frontend/src/TraceToMetrics/TraceToMetricsSettings.tsx` | Section title `"Trace to metrics"`; supported target datasources (`prometheus`, `victoriametrics-metrics-datasource`); placeholder time shifts `-2m` / `2m`; queries array shape `{ name?, query? }` |
| `TraceToProfilesSection`, `TraceToProfilesSettings` (writes `tracesToProfiles`) | `@grafana/o11y-ds-frontend@13.1.0-canary` | `packages/grafana-o11y-ds-frontend/src/TraceToProfiles/TraceToProfilesSettings.tsx` | Section title `"Trace to profiles"`; single supported target datasource (`grafana-pyroscope-datasource`); default tags `service.name`, `service.namespace`; `profileTypeId` populated from the target datasource's `profileTypes` resource |
| `NodeGraphSection`, `NodeGraphSettings` (writes `nodeGraph.enabled`) | `@grafana/o11y-ds-frontend@13.1.0-canary` | `packages/grafana-o11y-ds-frontend/src/NodeGraph/NodeGraphSettings.tsx` | `ConfigSubSection title="Node graph"`; `InlineField label="Enable node graph"` tooltip `"Displays the node graph above the trace view. Default: disabled"` |
| `SpanBarSection`, `SpanBarSettings` (writes `spanBar.{type, tag}`) | `@grafana/o11y-ds-frontend@13.1.0-canary` | `packages/grafana-o11y-ds-frontend/src/SpanBar/SpanBarSettings.tsx` | `ConfigSubSection title="Span bar"`; `Select` with three options (`None`, `Duration`, `Tag`), placeholder `"Duration"`, tooltip `"Default: duration"`; conditional `Tag key` input |
| `IntervalInput`, `TagMappingInput`, `ProfileTypesCascader` | `@grafana/o11y-ds-frontend@13.1.0-canary` | `packages/grafana-o11y-ds-frontend/src/{IntervalInput,TraceToLogs/TagMappingInput,pyroscope/ProfileTypesCascader}.tsx` | Time-shift/tag/profile-type input widgets; label & tooltip helpers in `getTimeShiftLabel`/`getTimeShiftTooltip` |
| `SecureSocksProxySettings` (excluded) | `@grafana/ui@13.1.0-canary` | grafana/grafana `packages/grafana-ui/src/components/DataSourceSettings/SecureSocksProxySettings.tsx` | Storage key it writes (`jsonData.enableSecureSocksProxy`) — confirmed and excluded |
| `DataSourcePicker` | `@grafana/runtime@13.1.0-canary` | grafana/grafana `packages/grafana-runtime/src/components/DataSourcePicker.tsx` | Used inside `ServiceGraphSettings`, `TraceToLogsSettings`, `TraceToMetricsSettings`, `TraceToProfilesSettings`; writes `<sectionKey>.datasourceUid` |
| `Input`, `InlineField`, `InlineFieldRow`, `InlineSwitch`, `Combobox`, `Select`, `Divider`, `Stack`, `TextLink`, `Alert` | `@grafana/ui@13.1.0-canary` | grafana/grafana `packages/grafana-ui/src/components/` | Prop names — needed to know which UI attributes to record |
| `DataSourceJsonData`, `DataSourcePluginOptionsEditorProps`, `updateDatasourcePluginJsonDataOption` | `@grafana/data@13.1.0-canary` | grafana/grafana `packages/grafana-data/src/` | Base jsonData interface and per-section option updater |

## Field provenance

Every dsconfig field, traced from its schema `id` to the upstream `file:line`
where each of its label, placeholder, tooltip, default, storage key, and value
type is defined.

### Root and shared HTTP/TLS fields

| Schema `id` | Storage key | Target | Label / provenance | Placeholder / options / default | Notes |
| --- | --- | --- | --- | --- | --- |
| `root_url` | `url` | `root` | `ConnectionSettings.tsx` (default `urlLabel = "URL"`) | `ConfigEditor.tsx:45` (`urlPlaceholder="http://localhost:3200"`) | Required per `pkg/tempo/tempo.go:78` + `pkg/tempo/grpc.go:86` |
| `virtual_authMethod` | — (virtual) | virtual | Default `AuthMethodSettings.tsx` label `"Authentication method"` | Options from `AuthMethodSettings.tsx`; default `'NoAuth'` mirrors `getSelectedMethod` fallthrough | `storage.computed.read` mirrors `getSelectedMethod` minus `CrossSiteCredentials`; `effects` mirror `onAuthMethodSelect` |
| `root_basicAuth` | `basicAuth` | `root` | — (managed by virtual) | Written by `utils.ts` on method select | Tagged `managed-by:virtual_authMethod` |
| `root_basicAuthUser` | `basicAuthUser` | `root` | `BasicAuth.tsx` (`userLabel = "User"`) | Placeholder `"User"`, tooltip `"The username of the data source account"` | `dependsOn: virtual_authMethod == 'BasicAuth'`; `requiredWhen: root_basicAuth == true` |
| `secureJsonData_basicAuthPassword` | `basicAuthPassword` | `secureJsonData` | `BasicAuth.tsx` (`passwordLabel = "Password"`) | Placeholder `"Password"`, tooltip `"The password of the data source account"` | Same conditional/required as `basicAuthUser` |
| `jsonData_oauthPassThru` | `oauthPassThru` | `jsonData` | — (managed by virtual) | Written by `utils.ts` on method select | Tagged `managed-by:virtual_authMethod` |
| `jsonData_tlsAuthWithCACert` | `tlsAuthWithCACert` | `jsonData` | `SelfSignedCertificate.tsx` (`label="Add self-signed certificate"`) | Default `false` | — |
| `secureJsonData_tlsCACert` | `tlsCACert` | `secureJsonData` | `SelfSignedCertificate.tsx` (`label="CA Certificate"`) | `placeholder="Begins with --- BEGIN CERTIFICATE ---"`, `rows: 6` | `dependsOn` / `requiredWhen`: `jsonData_tlsAuthWithCACert == true` |
| `jsonData_tlsAuth` | `tlsAuth` | `jsonData` | `TLSClientAuth.tsx` (`label="TLS Client Authentication"`) | Default `false` | — |
| `jsonData_serverName` | `serverName` | `jsonData` | `TLSClientAuth.tsx` (`label="ServerName"`) | `placeholder="domain.example.com"` | `dependsOn: jsonData_tlsAuth == true`; required for mTLS contract |
| `secureJsonData_tlsClientCert` | `tlsClientCert` | `secureJsonData` | `TLSClientAuth.tsx` (`label="Client Certificate"`) | `placeholder="Begins with --- BEGIN CERTIFICATE ---"`, `rows: 6` | Required when `tlsAuth` is `true` |
| `secureJsonData_tlsClientKey` | `tlsClientKey` | `secureJsonData` | `TLSClientAuth.tsx` (`label="Client Key"`) | ``placeholder="Begins with --- RSA PRIVATE KEY CERTIFICATE ---"`` — upstream typo preserved | Required when `tlsAuth` is `true`; see [Upstream findings](#upstream-findings) #2 |
| `jsonData_tlsSkipVerify` | `tlsSkipVerify` | `jsonData` | `SkipTLSVerification.tsx` (`label="Skip TLS certificate validation"`) | Default `false` | Role `transport.tlsSkipVerify` |
| `jsonData_keepCookies` | `keepCookies` | `jsonData` | `AdvancedHttpSettings.tsx` (`label="Allowed cookies"`) | Placeholder `"New cookie (hit enter to add)"` | — |
| `jsonData_timeout` | `timeout` | `jsonData` | `AdvancedHttpSettings.tsx` (`label="Timeout"`) | Placeholder `"Timeout in seconds"` | Role `transport.timeoutSeconds` |

### Tempo-specific fields

| Schema `id` | Storage key | Target | Label / provenance | Storage shape | Notes |
| --- | --- | --- | --- | --- | --- |
| `jsonData_streamingEnabled` | `streamingEnabled` | `jsonData` | `StreamingSection.tsx:24` (`ConfigSection title="Streaming"`) | `{ search?: boolean, metrics?: boolean }` | Only `streamingEnabled.search` is read server-side (`pkg/tempo/tempo.go:150-208`); `metrics` is frontend-only. Min Tempo versions: 2.2.0 (search) / 2.7.0 (metrics) from `src/datasource.ts:79-82` |
| `jsonData_tracesToLogsV2` | `tracesToLogsV2` | `jsonData` | `TraceToLogsSection.tsx` (`ConfigSection title="Trace to logs"`) | `{ datasourceUid?, tags?, spanStartTimeShift?, spanEndTimeShift?, filterByTraceID?, filterBySpanID?, query?, customQuery }` | The editor clears legacy `tracesToLogs` on every write to `tracesToLogsV2` |
| `jsonData_tracesToLogs` | `tracesToLogs` | `jsonData` | — (no editor UI; legacy) | `{ datasourceUid?, tags?, mappedTags?, mapTagNamesEnabled?, spanStartTimeShift?, spanEndTimeShift?, filterByTraceID?, filterBySpanID?, lokiSearch? }` | Migrated to v2 on read via `getTraceToLogsOptions:55-73`; kept for round-trip parity |
| `jsonData_tracesToMetrics` | `tracesToMetrics` | `jsonData` | `TraceToMetricsSection.tsx` (`ConfigSection title="Trace to metrics"`) | `{ datasourceUid?, tags?, queries?, spanStartTimeShift?, spanEndTimeShift? }` | `queries[]` is a list of `{ name?, query? }`; time shift placeholders `-2m` / `2m` |
| `jsonData_tracesToProfiles` | `tracesToProfiles` | `jsonData` | `TraceToProfilesSection.tsx` (`ConfigSection title="Trace to profiles"`) | `{ datasourceUid?, tags?, profileTypeId?, query?, customQuery }` | Only targets `grafana-pyroscope-datasource`; `profileTypeId` sourced from the target's `profileTypes` resource |
| `jsonData_serviceMap` | `serviceMap` | `jsonData` | `ConfigEditor.tsx:82` (`ConfigSubSection title="Service graph"`); `ServiceGraphSettings.tsx:23` (`InlineField label="Data source" tooltip="The Prometheus data source with the service graph data"`) | `{ datasourceUid?: string }` | Filters to `pluginId: 'prometheus'` |
| `jsonData_nodeGraph` | `nodeGraph` | `jsonData` | `NodeGraphSection.tsx` (`ConfigSubSection title="Node graph"`); `NodeGraphSettings.tsx:32` (`InlineField label="Enable node graph"`) | `{ enabled?: boolean }` | Default disabled |
| `jsonData_search` | `search` | `jsonData` | `ConfigEditor.tsx:97` (`ConfigSubSection title="Tempo search"`); `TraceQLSearchSettings.tsx:28,42` | `{ hide?: boolean, filters?: TraceqlFilter[] }` | The frontend seeds `filters` with two defaults on first load (`src/datasource.ts:147-159`) |
| `jsonData_traceQuery` | `traceQuery` | `jsonData` | `ConfigEditor.tsx:110` (`ConfigSubSection title="TraceID query"`); `QuerySettings.tsx:28-71` | `{ timeShiftEnabled?, spanStartTimeShift?, spanEndTimeShift? }` | Inputs disabled unless `timeShiftEnabled` is true |
| `jsonData_timeRangeForTags` | `timeRangeForTags` | `jsonData` | `ConfigEditor.tsx:123` (`ConfigSubSection title="Tags time range"`); `TagsTimeRangeSettings.tsx:26` (`InlineField label="Time range in query" tooltip="Time range in tags and tag value queries"`) | number (int seconds) | Only five values allowed (1800/10800/86400/259200/604800); default 1800; `ApplyDefaults` writes this default |
| `jsonData_tagLimit` | `tagLimit` | `jsonData` | `TagLimitSettings.tsx:25` (`InlineField label="Max tags and tag values"`) | number (accepts numeric string as well — see [Upstream findings](#upstream-findings) #3) | Placeholder `"5000"`; not persisted as a default |
| `jsonData_spanBar` | `spanBar` | `jsonData` | `SpanBarSection.tsx` (`ConfigSubSection title="Span bar"`); `SpanBarSettings.tsx:36,60` | `{ type?: 'None'\|'Duration'\|'Tag', tag?: string }` | `tag` required only when `type === 'Tag'` |

## Field inventory summary

| Schema field | Storage key | Target | Read by backend? |
| --- | --- | --- | --- |
| `root_url` | `url` | `root` | Yes — direct (`pkg/tempo/tempo.go:78` HTTP client, `pkg/tempo/grpc.go:86` gRPC client, `pkg/tempo/tempo.go:211` health check) |
| `virtual_authMethod` | — (virtual) | — | — (editor-local selector) |
| `root_basicAuth` | `basicAuth` | `root` | Yes (SDK `HTTPClientOptions`; gRPC per-RPC creds at `grpc.go:178`) |
| `root_basicAuthUser` | `basicAuthUser` | `root` | Yes (SDK + `grpc.go:181` via `opts.BasicAuth.User`) |
| `secureJsonData_basicAuthPassword` | `basicAuthPassword` | `secureJsonData` | Yes (SDK + `grpc.go:181` via `opts.BasicAuth.Password`) |
| `jsonData_oauthPassThru` | `oauthPassThru` | `jsonData` | Yes (SDK) |
| `jsonData_tlsAuthWithCACert` | `tlsAuthWithCACert` | `jsonData` | Yes (SDK, also picked up by `grpc.go:188` `httpclient.GetTLSConfig`) |
| `secureJsonData_tlsCACert` | `tlsCACert` | `secureJsonData` | Yes (SDK) |
| `jsonData_tlsAuth` | `tlsAuth` | `jsonData` | Yes (SDK + `grpc.go:186-197` for gRPC TLS) |
| `jsonData_serverName` | `serverName` | `jsonData` | Yes (SDK) |
| `secureJsonData_tlsClientCert` | `tlsClientCert` | `secureJsonData` | Yes (SDK) |
| `secureJsonData_tlsClientKey` | `tlsClientKey` | `secureJsonData` | Yes (SDK) |
| `jsonData_tlsSkipVerify` | `tlsSkipVerify` | `jsonData` | Yes (SDK) |
| `jsonData_keepCookies` | `keepCookies` | `jsonData` | Yes (SDK) |
| `jsonData_timeout` | `timeout` | `jsonData` | Yes (SDK) |
| `jsonData_streamingEnabled` | `streamingEnabled` | `jsonData` | Partial — `search` read by `CheckHealth` (`tempo.go:164-170`); `metrics` is frontend-only |
| `jsonData_tracesToLogsV2` | `tracesToLogsV2` | `jsonData` | No — frontend-only |
| `jsonData_tracesToLogs` | `tracesToLogs` | `jsonData` | No — legacy, migrated on read then cleared on write |
| `jsonData_tracesToMetrics` | `tracesToMetrics` | `jsonData` | No — frontend-only |
| `jsonData_tracesToProfiles` | `tracesToProfiles` | `jsonData` | No — frontend-only |
| `jsonData_serviceMap` | `serviceMap` | `jsonData` | No — frontend-only |
| `jsonData_nodeGraph` | `nodeGraph` | `jsonData` | No — frontend-only |
| `jsonData_search` | `search` | `jsonData` | No — frontend-only |
| `jsonData_traceQuery` | `traceQuery` | `jsonData` | No — frontend-only |
| `jsonData_timeRangeForTags` | `timeRangeForTags` | `jsonData` | No — frontend-only |
| `jsonData_tagLimit` | `tagLimit` | `jsonData` | No — frontend-only |
| `jsonData_spanBar` | `spanBar` | `jsonData` | No — frontend-only |

### Frontend-only settings

Almost every Tempo-specific jsonData field is frontend-only. The only settings
read server-side are `settings.URL` (from `backend.DataSourceInstanceSettings`,
not from jsonData) and `jsonData.streamingEnabled.search` (via a bare
`map[string]interface{}` unmarshal in `CheckHealth`). All others — trace-to-X
mappings, service graph, node graph, span bar, TraceQL search config,
TraceID-query time shifts, tag limit, tags time range — are consumed by the
frontend datasource class (`src/datasource.ts`) and the trace view.

### Backend-only settings

None. The Tempo backend does not accept jsonData fields the editor doesn't
also expose.

### Excluded settings

- **Secure Socks Proxy** (`jsonData.enableSecureSocksProxy` and associated
  socks-proxy fields) — rendered conditionally at `ConfigEditor.tsx:77-79`
  when `config.secureSocksDSProxyEnabled` is set on the Grafana instance.
  Deliberately omitted per `AGENTS.md`. Note: the Tempo gRPC client honors
  this via `settings.ProxyClient(ctx).SecureSocksProxyEnabled()`
  (`pkg/tempo/grpc.go:200-226`), so provisioning payloads that include it
  will still work at runtime.
- **Custom HTTP headers** (`@grafana/plugin-ui`'s `CustomHeaders`) — the
  editor writes indexed pairs `jsonData.httpHeaderName<N>` /
  `secureJsonData.httpHeaderValue<N>` starting at index 1. Not modeled as a
  first-class field because the storage keys are dynamic. Downstream tools
  should walk `jsonData` for the `httpHeaderName` prefix and pair up matching
  `httpHeaderValue<N>` secrets; the SDK's `HTTPClientOptions` already does
  this and forwards the resulting headers to Tempo on both HTTP and gRPC
  transports (`pkg/tempo/grpc.go:234-246`).

## Where the types are defined

### Frontend (TypeScript)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `TempoJsonData`, `TempoQuery`, `TempoQueryType`, `TraceqlFilter`, `TraceqlSearchScope` | `src/types.ts`, `src/dataquery.ts` | plugin ([grafana/grafana-tempo-datasource](https://github.com/grafana/grafana-tempo-datasource)) |
| `FeatureName`, `featuresToTempoVersion`, `DEFAULT_TIME_RANGE_FOR_TAGS` | `src/datasource.ts:70-82`, `src/configuration/TagsTimeRangeSettings.tsx:10` | plugin |
| `TagLimitOptions` | `src/configuration/TagLimitSettings.tsx:12-14` | plugin |
| `TraceToLogsOptions` (v1), `TraceToLogsOptionsV2`, `TraceToLogsTag`, `getTraceToLogsOptions`, `TraceToLogsSection` | `packages/grafana-o11y-ds-frontend/src/TraceToLogs/TraceToLogsSettings.tsx` | `@grafana/o11y-ds-frontend` `13.1.0-canary` |
| `TraceToMetricsOptions`, `TraceToMetricQuery`, `TraceToMetricsSection` | `packages/grafana-o11y-ds-frontend/src/TraceToMetrics/TraceToMetricsSettings.tsx` | `@grafana/o11y-ds-frontend` `13.1.0-canary` |
| `TraceToProfilesOptions`, `TraceToProfilesSection` | `packages/grafana-o11y-ds-frontend/src/TraceToProfiles/TraceToProfilesSettings.tsx` | `@grafana/o11y-ds-frontend` `13.1.0-canary` |
| `NodeGraphOptions`, `NodeGraphSection` | `packages/grafana-o11y-ds-frontend/src/NodeGraph/NodeGraphSettings.tsx` | `@grafana/o11y-ds-frontend` `13.1.0-canary` |
| `SpanBarOptions`, `SpanBarSection` | `packages/grafana-o11y-ds-frontend/src/SpanBar/SpanBarSettings.tsx` | `@grafana/o11y-ds-frontend` `13.1.0-canary` |
| `ConnectionSettings`, `Auth`, `AuthMethod`, `convertLegacyAuthProps`, `AdvancedHttpSettings`, `DataSourceDescription`, `ConfigSection`, `ConfigSubSection`, `ConfigDescriptionLink` | `src/components/ConfigEditor/…` | `@grafana/plugin-ui` `0.15.1` |
| `DataSourceJsonData` (base interface), `DataSourcePluginOptionsEditorProps`, `updateDatasourcePluginJsonDataOption` | `packages/grafana-data/src/` | `@grafana/data` `13.1.0-canary` |
| `SecureSocksProxySettings` (excluded), `Input`, `InlineField`, `InlineSwitch`, `Combobox`, `Select`, `Divider`, `Stack`, `TextLink`, `Alert` | `packages/grafana-ui/src/components/` | `@grafana/ui` `13.1.0-canary` |
| `DataSourcePicker`, `config` (`config.secureSocksDSProxyEnabled`) | `packages/grafana-runtime/src/` | `@grafana/runtime` `13.1.0-canary` |

### Backend (Go)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `DataSource`, `DatasourceInfo`, `NewDatasource`, `QueryData`, `CheckHealth`, `handleTags`, `handleTagValues`, `proxyToTempo` | `pkg/tempo/tempo.go:36-393` | plugin |
| `newGrpcClient`, `getDialOpts`, `basicAuth`, streaming interceptors (metrics/tracing/custom-headers/user-agent) | `pkg/tempo/grpc.go:79-346` | plugin |
| `stream_handler.go`, `search_stream.go`, `metrics_stream.go`, `search.go`, `trace.go`, `traceql_query.go`, `trace_transform.go`, `protospan_translation.go` | `pkg/tempo/*.go` | plugin |
| `backend.DataSourceInstanceSettings` (carries `URL`, `BasicAuthEnabled`, `BasicAuthUser`, `JSONData`, `DecryptedSecureJSONData`), `HTTPClientOptions(ctx)`, `ProxyClient` | `backend/common.go`, `backend/httpclient/`, `backend/proxy/` | `github.com/grafana/grafana-plugin-sdk-go` |

The models in this entry flatten the above into a single Go `Config` type
(root `URL`/`BasicAuth`/`BasicAuthUser`/`WithCredentials` tagged `json:"-"`,
plus every jsonData field as a value-typed nested struct, plus
`DecryptedSecureJSONData`) and a `SecureJsonDataKey` typed constant list.
`settings.ts` keeps the three canonical TypeScript types (`RootConfig`,
`JsonDataConfig`, `SecureJsonDataConfig`) alongside the nested config types.

## Modeling decisions

- **Virtual auth method**: `convertLegacyAuthProps`'s `onAuthMethodSelect`
  (`@grafana/plugin-ui utils.ts`) writes three storage fields in one shot —
  `root.basicAuth`, `root.withCredentials`, and `jsonData.oauthPassThru`.
  Tempo's editor default `visibleMethods` is `[BasicAuth, OAuthForward, NoAuth]`,
  so the virtual field's effects only write `basicAuth` and `oauthPassThru`.
  If a provisioning payload writes `withCredentials=true` directly, the SDK
  still honors it — the virtual's `storage.computed.read` doesn't preserve
  that state, but the underlying root storage does.
- **`requiredWhen` on `basicAuthUser` / `basicAuthPassword`**: keyed on the
  underlying storage field (`root_basicAuth == true`), not the virtual
  selector. The backend contract is "if basicAuth is on, we need a username
  and password".
- **TLS pair requirements**: `TLSClientAuth` and `SelfSignedCertificate` mark
  every field with `required` in the UI, but they only require the paired
  fields when the parent switch is on. Encoded as `dependsOn` + `requiredWhen`
  on each field.
- **Nested trace-to-X / node-graph / span-bar / search / traceQuery / streaming
  as `valueType: "object"` at the top level**: per `AGENTS.md`, complex nested
  objects are best modeled as opaque `object` fields with a `help` markdown
  block documenting the shape, rather than exploded into per-leaf schema
  entries via `section`. This keeps the schema focused (16 top-level fields
  instead of 60+) and defers detailed validation to the Go `Config` /
  `Validate` code — which mirrors the shapes the editor writes.
- **Legacy `tracesToLogs` (v1) retained**: `getTraceToLogsOptions`
  (`grafana-o11y-ds-frontend`) migrates v1 to v2 on read and the editor
  clears v1 on the next write, but datasources stored before the migration
  landed can still carry `tracesToLogs`. Kept as an editor-invisible field
  with `tags: ["legacy"]` so provisioning payloads round-trip. See the
  `legacyTracesToLogs` example in `SettingsExamples`.
- **`tagLimit` accepts number or numeric string**: `TagLimitSettings.tsx:32-34`
  binds `v.currentTarget.value` from an `Input type="number"`, so the
  persisted value is a string. `Config.UnmarshalJSON` accepts both shapes and
  coerces to `int64` for validation and Go-side use. See [Upstream findings](#upstream-findings) #3.
- **`timeRangeForTags` `allowedValues` validation**: enforced in the schema
  (`allowedValues` on the field) and again in `Config.Validate`. The default
  is applied by `ApplyDefaults` (the one hard default in Tempo's editor —
  fallback in `TagsTimeRangeSettings.tsx:30-36`).
- **`spanBar.type` closed-set validation**: `Select` in `SpanBarSettings.tsx:36`
  offers only three literals (`None`, `Duration`, `Tag`); `Config.Validate`
  rejects unknown values. `spanBar.tag` is required only when `type === 'Tag'`
  (matches the editor's conditional field rendering).
- **`streamingEnabled` metrics is frontend-consumable only**: `CheckHealth`
  only inspects `streamingEnabled.search`. If `streamingEnabled.metrics` is
  true but `search` is false, CheckHealth still probes the HTTP echo endpoint
  — the metrics-only path is exercised at query time from the frontend
  datasource class. Both keys are still schema-first-class because the
  editor writes both.
- **`search.filters` shape**: modeled as `SearchConfig.Filters
  []TraceqlFilter` where `TraceqlFilter.Value` is `any` (`interface{}`)
  because the editor stores either a `string` or a `[]string` depending on
  the operator (`dataquery.ts:139`). Matching schema value type is `object`
  at the parent field; the `help` markdown documents the sub-shape.
- **Custom HTTP headers and Secure Socks Proxy excluded**: see
  [Excluded settings](#excluded-settings) above.
- **Flat `Config` in Go**: `settings.go` collapses jsonData fields and
  decrypted secrets onto a single `Config` struct. Root-level fields the
  editor and SDK both use (`URL`, `BasicAuth`, `BasicAuthUser`,
  `WithCredentials`) are carried with `json:"-"` tags so `LoadConfig` returns
  them alongside the jsonData shape. Nested config types are declared as
  value-typed fields (not pointers) so the shared conformance walker
  (`schema.RunPluginTests`, `JSONDataTypesMatchStruct`) can match their Go
  `reflect.Struct` kind against the schema's `object` valueType.
- **`ApplyDefaults` writes exactly one field**: `timeRangeForTags` defaults
  to 1800 (30 minutes) to match `TagsTimeRangeSettings.tsx:30`. Everything
  else the editor visualises as a default (`tagLimit` "5000",
  `spanBar.type` "Duration", initial `search.filters`) is a render-time
  fallback or seeded by the frontend datasource constructor and never
  persisted; we mirror that by leaving those fields untouched.
- **`SecureJsonDataConfig` is a key list**: secure values are write-only, so
  the secure type is just the array of secret key names (`basicAuthPassword`,
  `tlsCACert`, `tlsClientCert`, `tlsClientKey`); consumers read
  `secureJsonFields` to see what is configured.

## SDK plugin schema and k8s-style examples (`schema.go`)

`NewSchema()` assembles the `grafana-plugin-sdk-go` `pluginschema.PluginSchema`
bundle (the k8s-style schema Grafana's datasource API server serves as
`{apiVersion}.json`, `v0alpha1` today) from the embedded `dsconfig.json`:
root fields plus a nested `jsonData` object become the OpenAPI settings
`spec`, secure fields become `secureValues`, and virtual fields are skipped.

`SettingsExamples()` provides the default configuration plus one k8s-style
example per authentication method, TLS variant, and Tempo-specific feature.
Each example is a full instance-settings object with the plugin configuration
nested under `jsonData` and the relevant write-only secrets under
`secureJsonData` (placeholder values to be replaced with real secrets):

| Example | Auth | TLS | Tempo extras | `secureJsonData` |
| --- | --- | --- | --- | --- |
| `""` (default) | None | — | — | `basicAuthPassword` (empty) |
| `noAuth` | None | — | `nodeGraph.enabled=true` | `basicAuthPassword` (empty) |
| `basicAuth` | Basic | — | — | `basicAuthPassword` |
| `oauthForward` | OAuth Identity | — | — | `basicAuthPassword` (empty) |
| `tlsMutualAuth` | None | mTLS (serverName + client cert/key) | — | `tlsClientCert`, `tlsClientKey` |
| `tlsSelfSignedCA` | None | Custom CA | — | `tlsCACert` |
| `streaming` | None | — | `streamingEnabled.search=true`, `.metrics=true` | `basicAuthPassword` (empty) |
| `fullObservability` | None | — | `nodeGraph`, `spanBar='Duration'`, `serviceMap→prometheus`, `tracesToLogsV2→loki`, `tracesToMetrics→prometheus` (2 named queries), `tracesToProfiles→pyroscope` | `basicAuthPassword` (empty) |
| `traceQLSearchAndTraceID` | None | — | `timeRangeForTags=10800`, `tagLimit=10000`, `search.filters=[service.name, span.name]`, `traceQuery.timeShiftEnabled=true` | `basicAuthPassword` (empty) |
| `legacyTracesToLogs` | None | — | Legacy `tracesToLogs` v1 payload | `basicAuthPassword` (empty) |

## `LoadConfig` utility (`settings.go`)

`LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error)`
runs the full three-phase load flow on a datasource instance's settings and
returns a fully-defaulted, validated `Config`:

1. **Parse** — copy `settings.URL`, `settings.BasicAuthEnabled`,
   `settings.BasicAuthUser` into `Config`, unmarshal `settings.JSONData` into
   the jsonData portion of the same struct via `Config.UnmarshalJSON` (which
   tolerantly parses `tagLimit` as either a number or a numeric string),
   and copy the four decrypted secrets into `DecryptedSecureJSONData`. The
   Tempo plugin has no upstream `LoadSettings` to mirror — `pkg/tempo/tempo.go:52-90`
   is the only server-side read of settings and it just uses `settings.URL`
   + `settings.HTTPClientOptions` + `newGrpcClient`.
2. **`ApplyDefaults`** — writes `TimeRangeForTags = 1800` if zero (see
   [Modeling decisions](#modeling-decisions)).
3. **`Validate`** — enforce the runtime contract: URL is required, Basic auth
   requires a username, mTLS requires serverName + client cert + client key,
   custom-CA requires the CA PEM, `timeout` must be non-negative, `spanBar.type`
   must be one of the three allowed literals when set (`Tag` requires
   `spanBar.tag`), `timeRangeForTags` must be one of the five allowed
   second-counts, and `tagLimit` must be non-negative. Errors are joined so
   every problem surfaces at once.

Everything is logged via `backend.Logger.FromContext(ctx)` with
`datasource_uid`, `datasource_name`, and `plugin` labels so log lines carry
request context.

### Direct access to individual phases

`(*Config).ApplyDefaults()` and `(Config).Validate() error` are still exported
for callers that want to compose them themselves (e.g. provisioning preview,
schema-example round-trip, tests that need to distinguish parse-level from
policy-level errors). Skip them by never calling `LoadConfig` in those flows —
assemble a `Config` directly.

## Upstream findings

Potential bugs, misleading UX, and consistency issues discovered while
researching upstream. All preserved verbatim in the schema — the schema
records what the plugin **does**, not what it **should** do.

1. **Frontend `TempoJsonData` type is out of date**: `src/types.ts:6-29`
   still declares `tracesToLogs?: TraceToLogsOptions` (v1) and
   `streamingEnabled?: { search?: boolean }` (missing `metrics`). The actual
   editor writes `tracesToLogsV2` and both `streamingEnabled.search` and
   `streamingEnabled.metrics`. The type also does not declare
   `tracesToMetrics` or `tracesToProfiles` at all — those keys are written
   by the shared sections but only reachable through the untyped
   `updateDatasourcePluginJsonDataOption` path. Downstream tooling that
   trusts `TempoJsonData` will miss half of what the editor persists.
2. **Upstream typo preserved**: `TLSClientAuth.tsx` sets the client key
   placeholder to `` `Begins with --- RSA PRIVATE KEY CERTIFICATE ---` `` —
   an RSA private key is not a "certificate". Preserved verbatim in
   `secureJsonData_tlsClientKey.ui.placeholder`. Shared across every plugin
   using `@grafana/plugin-ui`.
3. **`tagLimit` typed as number but persisted as string**: `TagLimitSettings.tsx:32`
   binds `v.currentTarget.value` (a string) directly, so on save the value
   round-trips through storage as a string. The frontend `?? ''` fallback at
   read time treats it as either. Provisioning APIs that specify `tagLimit`
   as a real JSON number are equally valid — the Config's `UnmarshalJSON`
   accepts both.
4. **`timeRangeForTags` is written as a raw number by the editor but never
   as a default**: `TagsTimeRangeSettings.tsx:30-36` falls back to
   `DEFAULT_TIME_RANGE_FOR_TAGS` (1800) at render time, but only writes when
   the user picks an option. A datasource that has always followed the
   default has `timeRangeForTags` undefined in storage. We surface this by
   writing the default in `Config.ApplyDefaults` so callers see 1800
   post-load; the schema also carries `defaultValue: 1800` so provisioning
   tooling can round-trip cleanly.
5. **`streamingEnabled.metrics` is not read by CheckHealth**: `tempo.go:164-170`
   only checks `streamingEnabled.search` when deciding whether to probe the
   gRPC streaming endpoint. A datasource with `metrics` streaming enabled
   but `search` streaming disabled will still pass a purely HTTP health
   check — even if metrics streaming happens to be broken.
6. **`serviceMap.datasourceUid` picker is filtered to `prometheus`**:
   `ServiceGraphSettings.tsx:28` hard-codes `pluginId="prometheus"` in the
   `DataSourcePicker`, so `victoriametrics-metrics-datasource` and other
   Prometheus-compatible datasources are not selectable from the UI even
   though they can theoretically serve `traces_service_graph_*` metrics.
   Provisioning payloads that point at a non-Prometheus UID will work as
   long as the target datasource exposes the required metrics.
7. **Legacy `tracesToLogs` is cleared on every write to v2**: any editor
   save wipes the v1 key. Provisioning tools that write both — expecting
   the backend to prefer v2 — will find their v1 payload silently deleted
   the moment a user opens the config editor.
8. **`spanBar.type` empty vs `'Duration'`**: `SpanBarSettings.tsx:44` uses
   `placeholder="Duration"` and `isClearable` — clearing the Select removes
   the persisted `type`, so a datasource with no `spanBar.type` set is
   effectively "Duration" via the placeholder. Our schema allows the empty
   string alongside `'None'`, `'Duration'`, `'Tag'`; `Validate` treats
   empty as OK.
9. **`opts.ForwardHTTPHeaders = true` is forced at connect time**: Tempo
   overrides the SDK default at `pkg/tempo/tempo.go:60`. There is no
   editor UI for it, and provisioning payloads can't override the
   override.
10. **The `TempoQuery` type mixes datasource-level and query-level fields**
    (`src/types.ts:31-35` — `serviceMapUseNativeHistograms`,
    `overrideStreamingEnabled`). These are per-query knobs, not config, and
    are intentionally excluded from `Config`.
11. **`SecureSocksProxySettings` also writes to jsonData**: the excluded
    Secure Socks Proxy widget writes `jsonData.enableSecureSocksProxy` and
    related fields. Provisioning payloads that include those keys will
    round-trip through the SDK but are not represented in `Config` or
    `SettingsExamples`.

## Validation performed

- `dsconfig.ParseAndResolveSchemaJSON` + `Schema.Validate()` (Go validator in
  this repo) — passes (via `TestSchemaConformance/ConfigSchemaValid`).
- `schema.RunPluginTests` conformance suite — passes (schema round-trip,
  artifact drift, spec/secure separation, `jsonData`/struct key parity,
  `jsonData`/struct type parity, secure-key parity).
- `go test ./...` on this entry — passes (`LoadConfig` incl. every auth
  method, streaming variant, TLS variant, legacy `tracesToLogs` shape,
  malformed jsonData, `tagLimit` as number and string, `spanBar` type
  guards, `timeRangeForTags` allowed-values guard).
- `settings.go` / `schema.go` / test files: `go build`, `go vet`, `gofmt` —
  clean across the whole `registry/` module.
- `settings.ts`: exports the three canonical types (`RootConfig`,
  `JsonDataConfig`, `SecureJsonDataConfig`) plus supporting nested types —
  reviewed by hand against the frontend sources; no `tsc` runner is wired
  into the registry module.
