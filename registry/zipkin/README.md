# zipkin

Declarative configuration schema for the [Zipkin datasource plugin](https://github.com/grafana/grafana-zipkin-datasource) (`zipkin`).

## Upstream researched

- **Repo**: `github.com/grafana/grafana-zipkin-datasource`
- **Ref**: `main`
- **Commit SHA**: `1982b15b27b2f7cc3cfaea9d1842c6c1c957bb32`

Every value in [`dsconfig.json`](dsconfig.json) — labels, placeholders,
tooltips, option labels/values, section titles, help markdown, defaults,
validations, dependency and required-when expressions, storage keys, storage
targets, value types, group titles, and instructions — is traceable to a
specific `file:line` in either the plugin repo (at this SHA) or the pinned
external editor libraries. See [Field provenance](#field-provenance) below.

To reproduce this research:

```bash
git clone --depth 20 https://github.com/grafana/grafana-zipkin-datasource
cd grafana-zipkin-datasource
git checkout 1982b15b27b2f7cc3cfaea9d1842c6c1c957bb32
```

If upstream `main` has advanced past this SHA, re-diff the sources listed
under [Sources researched](#sources-researched) before merging any changes to
this entry.

## Files

| File | Purpose |
| --- | --- |
| [`dsconfig.json`](dsconfig.json) | dsconfig v1 schema — single source of truth for all config fields, groups, effects, and instructions |
| [`settings.ts`](settings.ts) | TypeScript models: `RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig`, plus the nested config types |
| [`settings.go`](settings.go) | Go `Config` model (flat: root URL/basicAuth/basicAuthUser/withCredentials tagged `json:"-"`, jsonData fields as value-typed nested structs, `DecryptedSecureJSONData`), `PluginID`, typed enums (`SpanBarType`, `SecureJsonDataKey`), plus the `LoadConfig` utility (parse → `ApplyDefaults` → `Validate`) |
| [`schema.go`](schema.go) | k8s-style SDK plugin schema: embeds `dsconfig.json`, converts it via `dsconfig.NewSDKSchema`, and defines `SettingsExamples` for each auth/TLS/feature variant |
| [`settings_test.go`](settings_test.go) | Table tests for `LoadConfig`, `ApplyDefaults`, and `Validate` |
| [`conformance_test.go`](conformance_test.go) | Runs `schema.RunPluginTests` — the shared dsconfig conformance suite — against `dsconfig.json` |
| [`schema.gen.json`](schema.gen.json), [`settings.gen.json`](settings.gen.json), [`settings.examples.gen.json`](settings.examples.gen.json) | Committed schema artifacts (regenerate with `go generate ./...` inside this directory; the `SchemaArtifactInSync` conformance subtest guards drift) |

There is no per-entry `go.mod` — every registry entry is a subpackage of the
shared [`registry/`](..) module (`github.com/grafana/dsconfig/registry`).

## Sources researched

Every source below was read at the pinned upstream SHA (`1982b15`), plus
external editor components from the `@grafana/o11y-ds-frontend` and
`@grafana/plugin-ui` packages at the versions the plugin's `package.json` and
`yarn.lock` pin.

### Plugin repo (`github.com/grafana/grafana-zipkin-datasource@1982b15`)

| File | What was read |
| --- | --- |
| `src/plugin.json:1-47` | `pluginType` (`id` = `"zipkin"`), `pluginName` (`name` = `"Zipkin"`), docs URL (`info.links[2].url` = `"https://grafana.com/docs/grafana/latest/datasources/zipkin/"`) |
| `src/ConfigEditor.tsx:19-68` | Top-level editor — composes `DataSourceDescription` (`dataSourceName="Zipkin"`, `hasRequiredFields={false}`), `ConnectionSettings` (`urlPlaceholder="http://localhost:9411"`), `Auth` (via `convertLegacyAuthProps`), `TraceToLogsSection` (top-level), `TraceToMetricsSection` (top-level), plus a collapsible "Additional settings" `ConfigSection` (title `"Additional settings"`, description `"Additional settings are optional settings that can be configured for more control over your data source."`, `isCollapsible={true}`, `isInitiallyOpen={false}`) containing `AdvancedHttpSettings`, conditional `SecureSocksProxySettings` (excluded), `NodeGraphSection`, and `SpanBarSection` |
| `src/datasource.ts:21-35, 46-60` | Frontend `ZipkinJsonData` extends `DataSourceJsonData` with `nodeGraph?: NodeGraphOptions`; the constructor pulls `instanceSettings.jsonData.nodeGraph` and the query pipeline uses `nodeGraph?.enabled` to attach node-graph frames on the client side |
| `src/types.ts:1-34` | Frontend `ZipkinQuery`, `ZipkinSpan`, `ZipkinEndpoint`, `ZipkinAnnotation`, `ZipkinQueryType` (query-level, not config) — captured to confirm no additional jsonData fields |
| `src/plugin.json:6-13` | `backend: true`, `tracing: true`, `metrics: true`, `alerting/annotations/logs/streaming: false` — used to shape the doc/help text |
| `pkg/zipkin/zipkin.go:22-44` | `NewDatasource` — reads `settings.URL` directly, calls `settings.HTTPClientOptions(ctx)`, builds an HTTP client via `httpclient.NewProvider().New(...)`; **never unmarshals `settings.JSONData`** |
| `pkg/zipkin/zipkin.go:46-65` | `CheckHealth` — calls `Services()` (HTTP `/api/v2/services`) via the shared client |
| `pkg/zipkin/zipkin.go:67-77` | `CallResource` and `QueryData` — build the client with `NewZipkinClient(ctx, ds.info, logger)`; no jsonData read here either |
| `pkg/zipkin/client.go:16-38` | `DatasourceInfo` — stores `url`, `httpClient`, `logger`; `NewZipkinClient` factory |
| `pkg/zipkin/client.go:48-68` | `Services()` — GET `<url>/api/v2/services`, response is `[]string` |
| `pkg/zipkin/client.go:72-98` | `Spans(serviceName)` — GET `<url>/api/v2/spans?serviceName=…` |
| `pkg/zipkin/client.go:102-130` | `Traces(serviceName, spanName)` — GET `<url>/api/v2/traces?serviceName=…&spanName=…` |
| `pkg/zipkin/client.go:134-173` | `Trace(traceId)` — GET `<url>/api/v2/trace/{url.QueryEscape(traceId)}`; explicitly maps non-2xx statuses to `backend.DownstreamError` |
| `pkg/zipkin/client.go:175-197` | `createZipkinURL` helper — used by `Spans`/`Traces` to build URLs with query parameters |
| `pkg/zipkin/handler_callresource.go:11-71` | Resource routes: GET `/services`, `/spans`, `/traces`, `/trace/{traceId}` |
| `pkg/zipkin/handler_querydata.go:14-81` | `queryData` — decodes per-query JSON as `zipkinQuery { query, queryType }`; `traceID` calls `zc.Trace(query.Query)`; `upload` returns an unsupported-in-backend error (frontend-only). No datasource-level jsonData is read here |
| `pkg/zipkin/handler_querydata.go:62-72` | `zipkinQueryType` constants `traceID` / `upload` — mirrors the frontend `ZipkinQueryType` |
| `package.json:68-79` | External component versions (see next table) |

Notably absent: no `pkg/models/settings.go`, no typed `LoadSettings`, no
ad-hoc `datasourceJSONData` struct inside `zipkin.go` (unlike Jaeger). The
Zipkin backend does **not** read any jsonData field — every jsonData knob is
frontend-only or SDK-honoured via `settings.HTTPClientOptions`.

### External editor components

Read at the versions the plugin pins (see `package.json` / `yarn.lock`).

| Component | Version | Source consulted | What was read |
| --- | --- | --- | --- |
| `ConnectionSettings` | `@grafana/plugin-ui@0.13.1` | `src/components/ConfigEditor/Connection/ConnectionSettings.tsx` | URL label defaults to `"URL"`, placeholder passed by plugin (`ConfigEditor.tsx:32` — `urlPlaceholder="http://localhost:9411"`); required + built-in URL regex validation |
| `Auth`, `AuthMethodSettings`, `BasicAuth` | `@grafana/plugin-ui@0.13.1` | `src/components/ConfigEditor/Auth/{Auth,auth-method/AuthMethodSettings,auth-method/BasicAuth}.tsx` | Default `visibleMethods = [BasicAuth, OAuthForward, NoAuth]`; option labels `Basic authentication` / `Forward OAuth Identity` / `No Authentication`; description `"Choose an authentication method to access the data source"`; BasicAuth `User`/`Password` labels + placeholders + tooltips |
| `convertLegacyAuthProps` | `@grafana/plugin-ui@0.13.1` | `src/components/ConfigEditor/Auth/utils.ts` | Maps `basicAuth` / `withCredentials` / `jsonData.oauthPassThru` ↔ AuthMethod enum; `onAuthMethodSelect` writes basicAuth+withCredentials+oauthPassThru in one shot |
| TLS pack (`SelfSignedCertificate`, `TLSClientAuth`, `SkipTLSVerification`) | `@grafana/plugin-ui@0.13.1` | `src/components/ConfigEditor/Auth/tls/*.tsx` | All TLS labels/placeholders/rows come verbatim from these files (see [Field provenance](#field-provenance)); note the RSA private key placeholder typo `Begins with --- RSA PRIVATE KEY CERTIFICATE ---` (shared across plugins) |
| `CustomHeaders`, `CustomHeader` | `@grafana/plugin-ui@0.13.1` | `src/components/ConfigEditor/Auth/custom-headers/*.tsx` | Indexed `httpHeaderName<N>` / `httpHeaderValue<N>` storage pattern; **not modeled** in this schema (see [Modeling decisions](#modeling-decisions)) |
| `AdvancedHttpSettings` | `@grafana/plugin-ui@0.13.1` | `src/components/ConfigEditor/AdvancedSettings/AdvancedHttpSettings.tsx` | `Allowed cookies` and `Timeout` labels/tooltips/placeholders |
| `DataSourceDescription`, `ConfigSection`, `ConfigSubSection` | `@grafana/plugin-ui@0.13.1` | `src/components/ConfigEditor/*` | Section title/description props (no storage keys — layout only) |
| `TraceToLogsSection`, `TraceToLogsSettings` (writes `tracesToLogsV2`; migrates legacy `tracesToLogs` via `getTraceToLogsOptions`) | `@grafana/o11y-ds-frontend@13.0.1` | `github.com/grafana/grafana/blob/main/packages/grafana-o11y-ds-frontend/src/TraceToLogs/TraceToLogsSettings.tsx` | Section title `"Trace to logs"`; supported target datasource types (`loki`, `elasticsearch`, `grafana-splunk-datasource`, `grafana-opensearch-datasource`, `grafana-falconlogscale-datasource`, `googlecloud-logging-datasource`, `victoriametrics-logs-datasource`); v1→v2 migration in `getTraceToLogsOptions`; field labels (`Data source`, `Span start time shift`, `Span end time shift`, `Tags`, `Filter by trace ID`, `Filter by span ID`, `Use custom query`, `Query`) |
| `TraceToMetricsSection`, `TraceToMetricsSettings` (writes `tracesToMetrics`) | `@grafana/o11y-ds-frontend@13.0.1` | `packages/grafana-o11y-ds-frontend/src/TraceToMetrics/TraceToMetricsSettings.tsx` | Section title `"Trace to metrics"`; supported target datasources (`prometheus`, `victoriametrics-metrics-datasource`); placeholder time shifts `-2m` / `2m`; queries array shape `{ name?, query? }` |
| `NodeGraphSection`, `NodeGraphSettings` (writes `nodeGraph.enabled`) | `@grafana/o11y-ds-frontend@13.0.1` | `packages/grafana-o11y-ds-frontend/src/NodeGraph/NodeGraphSettings.tsx` | `ConfigSubSection title="Node graph"`; `InlineField label="Enable node graph"` tooltip `"Displays the node graph above the trace view. Default: disabled"` |
| `SpanBarSection`, `SpanBarSettings` (writes `spanBar.{type, tag}`) | `@grafana/o11y-ds-frontend@13.0.1` | `packages/grafana-o11y-ds-frontend/src/SpanBar/SpanBarSettings.tsx` | `ConfigSubSection title="Span bar"`; `Select` with three options (`None`, `Duration`, `Tag`), placeholder `"Duration"`, tooltip `"Default: duration"`; conditional `Tag key` input |
| `createNodeGraphFrames` | `@grafana/o11y-ds-frontend@13.0.1` | `packages/grafana-o11y-ds-frontend/src/NodeGraph/…` | Client-side helper the Zipkin datasource calls in `addNodeGraphFramesToResponse` (`src/datasource.ts:111-124`) when `jsonData.nodeGraph.enabled` is true |
| `SecureSocksProxySettings` (excluded) | `@grafana/ui@13.0.1` | grafana/grafana `packages/grafana-ui/src/components/DataSourceSettings/SecureSocksProxySettings.tsx` | Storage key it writes (`jsonData.enableSecureSocksProxy`) — confirmed and excluded |
| `Divider`, `Stack`, `InlineField`, `InlineFieldRow`, `InlineSwitch`, `useStyles2` | `@grafana/ui@13.0.1` | grafana/grafana `packages/grafana-ui/src/components/` | Prop names — needed to know which UI attributes to record |
| `DataSourceJsonData`, `DataSourcePluginOptionsEditorProps` | `@grafana/data@13.0.1` | grafana/grafana `packages/grafana-data/src/` | Base jsonData interface and editor props |
| `config` (`config.secureSocksDSProxyEnabled`), `DataSourceWithBackend`, `getTemplateSrv`, `TemplateSrv` | `@grafana/runtime@13.0.1` | grafana/grafana `packages/grafana-runtime/src/` | Feature-flag switch that gates rendering of the excluded proxy widget; backend-datasource plumbing |

## Field provenance

Every dsconfig field, traced from its schema `id` to the upstream `file:line`
where each of its label, placeholder, tooltip, default, storage key, and value
type is defined.

### Root and shared HTTP/TLS fields

| Schema `id` | Storage key | Target | Label / provenance | Placeholder / options / default | Notes |
| --- | --- | --- | --- | --- | --- |
| `root_url` | `url` | `root` | `ConnectionSettings.tsx` (default `urlLabel = "URL"`) | `ConfigEditor.tsx:32` (`urlPlaceholder="http://localhost:9411"`) | Required per `pkg/zipkin/zipkin.go:33-35` + `pkg/zipkin/client.go:50` |
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

### Zipkin-specific fields

| Schema `id` | Storage key | Target | Label / provenance | Storage shape | Notes |
| --- | --- | --- | --- | --- | --- |
| `jsonData_tracesToLogsV2` | `tracesToLogsV2` | `jsonData` | `TraceToLogsSection.tsx` (`ConfigSection title="Trace to logs"`) | `{ datasourceUid?, tags?, spanStartTimeShift?, spanEndTimeShift?, filterByTraceID?, filterBySpanID?, query?, customQuery }` | The editor clears legacy `tracesToLogs` on every write to `tracesToLogsV2` |
| `jsonData_tracesToLogs` | `tracesToLogs` | `jsonData` | — (no editor UI; legacy) | `{ datasourceUid?, tags?, mappedTags?, mapTagNamesEnabled?, spanStartTimeShift?, spanEndTimeShift?, filterByTraceID?, filterBySpanID?, lokiSearch? }` | Migrated to v2 on read via `getTraceToLogsOptions`; kept for round-trip parity |
| `jsonData_tracesToMetrics` | `tracesToMetrics` | `jsonData` | `TraceToMetricsSection.tsx` (`ConfigSection title="Trace to metrics"`) | `{ datasourceUid?, tags?, queries?, spanStartTimeShift?, spanEndTimeShift? }` | `queries[]` is a list of `{ name?, query? }`; time shift placeholders `-2m` / `2m` |
| `jsonData_nodeGraph` | `nodeGraph` | `jsonData` | `NodeGraphSection.tsx` (`ConfigSubSection title="Node graph"`); `NodeGraphSettings.tsx` (`InlineField label="Enable node graph"`) | `{ enabled?: boolean }` | Default disabled. Also read on the frontend by the Zipkin datasource class (`src/datasource.ts:34,55`) to attach node-graph frames |
| `jsonData_spanBar` | `spanBar` | `jsonData` | `SpanBarSection.tsx` (`ConfigSubSection title="Span bar"`); `SpanBarSettings.tsx` | `{ type?: 'None'\|'Duration'\|'Tag', tag?: string }` | `tag` required only when `type === 'Tag'` |

## Field inventory summary

| Schema field | Storage key | Target | Read by backend? |
| --- | --- | --- | --- |
| `root_url` | `url` | `root` | Yes — direct (`pkg/zipkin/zipkin.go:33-42`, every client method in `pkg/zipkin/client.go`) |
| `virtual_authMethod` | — (virtual) | — | — (editor-local selector) |
| `root_basicAuth` | `basicAuth` | `root` | Yes (SDK `HTTPClientOptions`) |
| `root_basicAuthUser` | `basicAuthUser` | `root` | Yes (SDK) |
| `secureJsonData_basicAuthPassword` | `basicAuthPassword` | `secureJsonData` | Yes (SDK) |
| `jsonData_oauthPassThru` | `oauthPassThru` | `jsonData` | Yes (SDK) |
| `jsonData_tlsAuthWithCACert` | `tlsAuthWithCACert` | `jsonData` | Yes (SDK) |
| `secureJsonData_tlsCACert` | `tlsCACert` | `secureJsonData` | Yes (SDK) |
| `jsonData_tlsAuth` | `tlsAuth` | `jsonData` | Yes (SDK) |
| `jsonData_serverName` | `serverName` | `jsonData` | Yes (SDK) |
| `secureJsonData_tlsClientCert` | `tlsClientCert` | `secureJsonData` | Yes (SDK) |
| `secureJsonData_tlsClientKey` | `tlsClientKey` | `secureJsonData` | Yes (SDK) |
| `jsonData_tlsSkipVerify` | `tlsSkipVerify` | `jsonData` | Yes (SDK) |
| `jsonData_keepCookies` | `keepCookies` | `jsonData` | Yes (SDK) |
| `jsonData_timeout` | `timeout` | `jsonData` | Yes (SDK) |
| `jsonData_tracesToLogsV2` | `tracesToLogsV2` | `jsonData` | No — frontend-only |
| `jsonData_tracesToLogs` | `tracesToLogs` | `jsonData` | No — legacy, migrated on read then cleared on write |
| `jsonData_tracesToMetrics` | `tracesToMetrics` | `jsonData` | No — frontend-only |
| `jsonData_nodeGraph` | `nodeGraph` | `jsonData` | No — frontend-only (consumed by `src/datasource.ts:34,55` on the client) |
| `jsonData_spanBar` | `spanBar` | `jsonData` | No — frontend-only |

### Frontend-only settings

Every Zipkin-specific jsonData field is frontend-only. The only settings read
server-side are `settings.URL` (from `backend.DataSourceInstanceSettings`, not
from jsonData) and the SDK-managed HTTP settings. There is no equivalent to
Jaeger's `traceIdTimeParams.enabled` — nothing in `pkg/zipkin` unmarshals
`settings.JSONData` at all.

### Backend-only settings

None. The Zipkin backend does not accept jsonData fields the editor doesn't
also expose.

### Excluded settings

- **Secure Socks Proxy** (`jsonData.enableSecureSocksProxy` and associated
  socks-proxy fields) — rendered conditionally at `ConfigEditor.tsx:58-60`
  when `config.secureSocksDSProxyEnabled` is set on the Grafana instance.
  Deliberately omitted per `AGENTS.md`. Provisioning payloads that include it
  will still work at runtime because the SDK's `HTTPClientOptions` honours
  the flag transparently.
- **Custom HTTP headers** (`@grafana/plugin-ui`'s `CustomHeaders`) — the
  editor writes indexed pairs `jsonData.httpHeaderName<N>` /
  `secureJsonData.httpHeaderValue<N>` starting at index 1. Not modeled as a
  first-class field because the storage keys are dynamic. Downstream tools
  should walk `jsonData` for the `httpHeaderName` prefix and pair up matching
  `httpHeaderValue<N>` secrets; the SDK's `HTTPClientOptions` already does
  this and forwards the resulting headers to Zipkin over HTTP.

## Where the types are defined

### Frontend (TypeScript)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `ZipkinQuery`, `ZipkinQueryType`, `ZipkinSpan`, `ZipkinEndpoint`, `ZipkinAnnotation` | `src/types.ts` | plugin ([grafana/grafana-zipkin-datasource](https://github.com/grafana/grafana-zipkin-datasource)) |
| `ZipkinJsonData`, `ZipkinDatasource`, `addNodeGraphFramesToResponse` | `src/datasource.ts:21-124` | plugin |
| `TraceToLogsOptions` (v1), `TraceToLogsOptionsV2`, `TraceToLogsTag`, `getTraceToLogsOptions`, `TraceToLogsSection` | `packages/grafana-o11y-ds-frontend/src/TraceToLogs/TraceToLogsSettings.tsx` | `@grafana/o11y-ds-frontend@13.0.1` |
| `TraceToMetricsOptions`, `TraceToMetricQuery`, `TraceToMetricsSection` | `packages/grafana-o11y-ds-frontend/src/TraceToMetrics/TraceToMetricsSettings.tsx` | `@grafana/o11y-ds-frontend@13.0.1` |
| `NodeGraphOptions`, `NodeGraphSection`, `createNodeGraphFrames` | `packages/grafana-o11y-ds-frontend/src/NodeGraph/…` | `@grafana/o11y-ds-frontend@13.0.1` |
| `SpanBarOptions`, `SpanBarSection` | `packages/grafana-o11y-ds-frontend/src/SpanBar/SpanBarSettings.tsx` | `@grafana/o11y-ds-frontend@13.0.1` |
| `ConnectionSettings`, `Auth`, `AuthMethod`, `convertLegacyAuthProps`, `AdvancedHttpSettings`, `DataSourceDescription`, `ConfigSection` | `src/components/ConfigEditor/…` | `@grafana/plugin-ui@0.13.1` |
| `DataSourceJsonData`, `DataSourcePluginOptionsEditorProps` | `packages/grafana-data/src/` | `@grafana/data@13.0.1` |
| `SecureSocksProxySettings` (excluded), `Divider`, `Stack`, `useStyles2` | `packages/grafana-ui/src/components/` | `@grafana/ui@13.0.1` |
| `config` (`config.secureSocksDSProxyEnabled`), `DataSourceWithBackend`, `getTemplateSrv`, `TemplateSrv` | `packages/grafana-runtime/src/` | `@grafana/runtime@13.0.1` |

### Backend (Go)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `Datasource`, `NewDatasource`, `CheckHealth`, `CallResource`, `QueryData` | `pkg/zipkin/zipkin.go:15-77` | plugin |
| `DatasourceInfo`, `Client`, `baseClientImpl`, `NewZipkinClient`, `Services`, `Spans`, `Traces`, `Trace`, `createZipkinURL` | `pkg/zipkin/client.go:16-197` | plugin |
| Resource routing (`registerResourceRoutes`, `get*Handler`) | `pkg/zipkin/handler_callresource.go:11-71` | plugin |
| `zipkinQueryType`, `zipkinQuery`, `queryData`, `loadQuery`, `TraceKeyValuePair`, `TraceLog`, `transformResponse` | `pkg/zipkin/handler_querydata.go:14-261` | plugin |
| `model.SpanModel`, `model.Annotation`, `model.Endpoint` | `github.com/openzipkin/zipkin-go/model` | `openzipkin/zipkin-go` |
| `backend.DataSourceInstanceSettings` (carries `URL`, `BasicAuthEnabled`, `BasicAuthUser`, `JSONData`, `DecryptedSecureJSONData`), `HTTPClientOptions(ctx)`, `httpclient.NewProvider().New(...)` | `backend/common.go`, `backend/httpclient/` | `github.com/grafana/grafana-plugin-sdk-go` |

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
  Zipkin's editor uses `visibleMethods` (default `[BasicAuth, OAuthForward,
  NoAuth]`), so the virtual field's effects only write `basicAuth` and
  `oauthPassThru`. If a provisioning payload writes `withCredentials=true`
  directly, the SDK still honors it — the virtual's `storage.computed.read`
  doesn't preserve that state, but the underlying root storage does.
- **`requiredWhen` on `basicAuthUser` / `basicAuthPassword`**: keyed on the
  underlying storage field (`root_basicAuth == true`), not the virtual
  selector. The backend contract is "if basicAuth is on, we need a username
  and password".
- **TLS pair requirements**: `TLSClientAuth` and `SelfSignedCertificate` mark
  every field with `required` in the UI, but they only require the paired
  fields when the parent switch is on. Encoded as `dependsOn` + `requiredWhen`
  on each field.
- **Nested trace-to-X / node-graph / span-bar as `valueType: "object"` at the
  top level**: per `AGENTS.md`, complex nested objects are best modeled as
  opaque `object` fields with a `help` markdown block documenting the shape,
  rather than exploded into per-leaf schema entries. This keeps the schema
  focused and defers detailed validation to the Go `Config` / `Validate`
  code — which mirrors the shapes the editor writes.
- **Legacy `tracesToLogs` (v1) retained**: `getTraceToLogsOptions`
  (`grafana-o11y-ds-frontend`) migrates v1 to v2 on read and the editor
  clears v1 on the next write, but datasources stored before the migration
  landed can still carry `tracesToLogs`. Kept as an editor-invisible field
  with `tags: ["legacy"]` so provisioning payloads round-trip. See the
  `legacyTracesToLogs` example in `SettingsExamples`.
- **`spanBar.type` closed-set validation**: `Select` in `SpanBarSettings.tsx`
  offers only three literals (`None`, `Duration`, `Tag`); `Config.Validate`
  rejects unknown values. `spanBar.tag` is required only when
  `type === 'Tag'` (matches the editor's conditional field rendering).
- **`ApplyDefaults` is a no-op**: unlike Tempo, the Zipkin editor writes no
  persisted defaults into jsonData. `ApplyDefaults` is kept for API symmetry
  and to satisfy the intended parse → default → validate lifecycle.
- **No Query Trace by ID / Time Params section**: Zipkin's editor does not
  include the plugin-local component Jaeger ships. There is no
  `traceIdTimeParams` storage key in Zipkin; consumers should not expect one.
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
example per authentication method, TLS variant, and Zipkin-specific feature.
Each example is a full instance-settings object with the plugin configuration
nested under `jsonData` and the relevant write-only secrets under
`secureJsonData` (placeholder values to be replaced with real secrets):

| Example | Auth | TLS | Zipkin extras | `secureJsonData` |
| --- | --- | --- | --- | --- |
| `""` (default) | None | — | — | `basicAuthPassword` (empty) |
| `noAuth` | None | — | `nodeGraph.enabled=true` | `basicAuthPassword` (empty) |
| `basicAuth` | Basic | — | — | `basicAuthPassword` |
| `oauthForward` | OAuth Identity | — | — | `basicAuthPassword` (empty) |
| `tlsMutualAuth` | None | mTLS (serverName + client cert/key) | — | `tlsClientCert`, `tlsClientKey` |
| `tlsSelfSignedCA` | None | Custom CA | — | `tlsCACert` |
| `fullObservability` | None | — | `nodeGraph`, `spanBar='Duration'`, `tracesToLogsV2→loki`, `tracesToMetrics→prometheus` (2 named queries) | `basicAuthPassword` (empty) |
| `legacyTracesToLogs` | None | — | Legacy `tracesToLogs` v1 payload | `basicAuthPassword` (empty) |

## `LoadConfig` utility (`settings.go`)

`LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error)`
runs the full three-phase load flow on a datasource instance's settings and
returns a fully-defaulted, validated `Config`:

1. **Parse** — copy `settings.URL`, `settings.BasicAuthEnabled`,
   `settings.BasicAuthUser` into `Config`, unmarshal `settings.JSONData` into
   the jsonData portion of the same struct, and copy the four decrypted
   secrets into `DecryptedSecureJSONData`. The Zipkin plugin has no upstream
   `LoadSettings` to mirror — `pkg/zipkin/zipkin.go:22-44` is the only
   server-side read of settings and it just uses `settings.URL` +
   `settings.HTTPClientOptions`; **nothing in `pkg/zipkin` unmarshals
   `settings.JSONData`**.
2. **`ApplyDefaults`** — no-op (the Zipkin editor persists no defaults).
3. **`Validate`** — enforce the runtime contract: URL is required, Basic auth
   requires a username, mTLS requires serverName + client cert + client key,
   custom-CA requires the CA PEM, `timeout` must be non-negative, and
   `spanBar.type` must be one of the three allowed literals when set (`Tag`
   requires `spanBar.tag`). Errors are joined so every problem surfaces at
   once.

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

1. **`hasRequiredFields={false}` on `DataSourceDescription` contradicts the
   backend contract**: `ConfigEditor.tsx:27` tells the description block that
   no fields are required, but `pkg/zipkin/zipkin.go:33-35` refuses to
   construct the datasource when `settings.URL` is empty. The dsconfig entry
   marks `root_url` as `requiredWhen: "true"` to reflect the true contract.
2. **Upstream typo preserved**: `@grafana/plugin-ui`'s `TLSClientAuth.tsx`
   sets the client key placeholder to `` `Begins with --- RSA PRIVATE KEY
   CERTIFICATE ---` `` — an RSA private key is not a "certificate". Preserved
   verbatim in `secureJsonData_tlsClientKey.ui.placeholder`. Shared across
   every plugin using `@grafana/plugin-ui`.
3. **`nodeGraph` is duplicated on the frontend**: `ZipkinJsonData`
   (`src/datasource.ts:21-23`) declares only `nodeGraph?: NodeGraphOptions`,
   but the datasource class also assigns `spanBar` on itself
   (`src/datasource.ts:28`) without a corresponding type — `spanBar` is only
   read by the trace view via `jsonData.spanBar`. `spanBar` is missing from
   `ZipkinJsonData` but is written to `jsonData` by the editor.
4. **`Spans()` swallows the HTTP error before deferring body close**:
   `pkg/zipkin/client.go:83-97` — the `defer` that closes `res.Body` runs
   even after the `httpClient.Get` failed (nil response), guarded by a `if
   res != nil` check. The same pattern is not used in `Services`
   (`client.go:59-63`); a nil response there would panic. Not a config
   concern but noted for reviewers.
5. **`Trace()` explicitly maps non-2xx to `DownstreamError`**:
   `pkg/zipkin/client.go:161-166` classifies HTTP failures as downstream via
   `backend.ErrorSourceFromHTTPStatus`. The other client methods do not —
   they return raw errors. Provisioning tooling should not assume every
   non-2xx surfaces the same way.
6. **Legacy `tracesToLogs` is cleared on every write to v2**: any editor
   save wipes the v1 key. Provisioning tools that write both — expecting
   the backend to prefer v2 — will find their v1 payload silently deleted
   the moment a user opens the config editor.
7. **`upload` query type is frontend-only**: `pkg/zipkin/handler_querydata.go:33-38`
   returns an error for the `upload` query type; only the frontend
   `ZipkinDatasource.query` method (`src/datasource.ts:37-63`) handles it by
   parsing `this.uploadedJson`. Provisioning payloads that seed uploaded
   traces cannot execute those queries server-side.
8. **`nodeGraph` field is redundantly consumed**: the frontend datasource
   (`src/datasource.ts:34,55`) reads `jsonData.nodeGraph.enabled` on every
   query to decide whether to attach node-graph frames on the client side;
   the trace view also reads it. There is no server-side node graph in
   Zipkin — the transformation is entirely frontend.
9. **Backend never unmarshals `settings.JSONData`**: unlike Jaeger (which
   reads `traceIdTimeParams.enabled`) or Tempo (which decodes
   `streamingEnabled.search`), no code path in `pkg/zipkin` calls
   `json.Unmarshal(settings.JSONData, …)`. The Zipkin backend is
   configuration-agnostic beyond `settings.URL` and the SDK's HTTP client.

## Validation performed

- `dsconfig.ParseAndResolveSchemaJSON` + `Schema.Validate()` (Go validator in
  this repo) — passes (via `TestSchemaConformance/ConfigSchemaValid`).
- `schema.RunPluginTests` conformance suite — passes (schema round-trip,
  artifact drift, spec/secure separation, `jsonData`/struct key parity,
  `jsonData`/struct type parity, secure-key parity).
- `go test ./...` on this entry — passes (`LoadConfig` incl. every auth
  method, TLS variant, legacy `tracesToLogs` shape, malformed jsonData,
  `spanBar` type guards).
- `settings.go` / `schema.go` / test files: `go build`, `go vet`, `gofmt` —
  clean across the whole `registry/` module.
- `settings.ts`: exports the three canonical types (`RootConfig`,
  `JsonDataConfig`, `SecureJsonDataConfig`) plus supporting nested types —
  reviewed by hand against the frontend sources; no `tsc` runner is wired
  into the registry module.
