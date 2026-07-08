# grafana-mock-datasource

Declarative configuration schema for the [Grafana Mock datasource plugin](https://github.com/grafana/mock-datasource) (plugin id `grafana-mock-datasource`).

The Mock datasource is a small, deterministic testing/debug plugin: its query engine returns synthetic data frames from either the frontend (`json_framer`, inline / URL / scenario raw frames) or the backend (default). Its config surface reflects that — a single plugin-owned override for `CheckHealth` on top of Grafana's shared HTTP-settings model.

## Upstream researched

- **Repo**: `github.com/grafana/mock-datasource`
- **Ref**: `main`
- **Commit SHA**: `090f6f2ad0ea812a5bac3c638e3eab8f432a9597`

Every value in [`dsconfig.json`](dsconfig.json) — labels, tooltips, option labels/values, section titles, defaults, storage keys, storage targets, value types, group titles, and instructions — is traceable to a specific `file:line` in the upstream repo at this SHA. See [Field provenance](#field-provenance) below.

To reproduce this research:

```bash
git clone --depth 20 https://github.com/grafana/mock-datasource
cd mock-datasource
git checkout 090f6f2ad0ea812a5bac3c638e3eab8f432a9597
```

If upstream `main` has advanced past this SHA, re-diff the sources listed under [Sources researched](#sources-researched) before merging any changes to this entry.

## Files

| File | Purpose |
| --- | --- |
| [`dsconfig.json`](dsconfig.json) | dsconfig v1 schema — single source of truth for all config fields, groups, effects, and instructions |
| [`settings.ts`](settings.ts) | TypeScript models: `RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig` |
| [`settings.go`](settings.go) | Go `Config` model (flat: root URL/basicAuth/basicAuthUser/withCredentials tagged `json:"-"`, jsonData fields, `DecryptedSecureJSONData`), `PluginID`, `CustomHealthCheckStatus` typed constants, `SecureJsonDataKey` typed constants, and the `LoadConfig` utility |
| [`schema.go`](schema.go) | k8s-style SDK plugin schema: embeds `dsconfig.json`, converts it via `dsconfig.NewSDKSchema`, and defines `SettingsExamples` for each auth / TLS / CustomHealthCheck variant |
| [`settings_test.go`](settings_test.go) | Table tests for `LoadConfig`, `ApplyDefaults`, and `Validate` |
| [`conformance_test.go`](conformance_test.go) | Runs `schema.RunPluginTests` — the shared dsconfig conformance suite — against `dsconfig.json` |
| [`schema.gen.json`](schema.gen.json), [`settings.gen.json`](settings.gen.json), [`settings.examples.gen.json`](settings.examples.gen.json) | Committed schema artifacts (regenerate with `go generate ./...` inside this directory; the `SchemaArtifactInSync` conformance subtest guards drift) |

There is no per-entry `go.mod` — every registry entry is a subpackage of the shared [`registry/`](..) module (`github.com/grafana/dsconfig/registry`). The Go package name for this entry is `mockdatasource` (a language-safe form of the plugin id).

## Sources researched

Every source below was read at the pinned upstream SHA (`090f6f2`), plus external editor components at the exact versions the plugin's `package.json` pins.

### Plugin repo (`github.com/grafana/mock-datasource@090f6f2`)

| File | What was read |
| --- | --- |
| `src/plugin.json:1-56` | `pluginType` (`id` = `"grafana-mock-datasource"`), `pluginName` (`name` = `"Mock"`). No `info.links`, no `aliasIDs`; no canonical docs URL. `dependencies.grafanaDependency = ">=12.4.3-0"`. |
| `src/editors/MockConfigEditor.tsx:1-115` | Top-level editor. Three sections: a bare `<ConfigSection title="">` wrapping `<ConnectionSettings config onChange />` (`:19-21`, no `urlPlaceholder` override — the plugin-ui default `'URL'` applies); a second `<ConfigSection title="">` wrapping `<Auth {...convertLegacyAuthProps(...)} />` (`:23-25` — default `visibleMethods = [BasicAuth, OAuthForward, NoAuth]`); a `<ConfigSection title="Custom HealthCheck">` (`:27-45`) with an `InlineSwitch` for `customHealthCheckEnabled` (`:33-36`) and a conditional `<CustomHealthCheckOptionsEditor value={jsonData.customHealthCheck \|\| { status: 0 }} onChange />` (`:38-43`). |
| `src/editors/MockConfigEditor.tsx:50-114` | `CustomHealthCheckOptionsEditor`. Renders four rows: `RadioButtonGroup<number>` for `status` with options `[{value:1,label:'OK'},{value:2,label:'ERROR'},{value:0,label:'UNKNOWN'}]` (`:69-77`), `Input` for `message` with `onBlur` persistence (`:80-88`), `CodeEditor language="json"` for `details` (`:90-102`), and `InlineSwitch` for `skipBackend` (`:105-111`). |
| `src/selectors.ts:1-26` | Verbatim labels and tooltips consumed by `MockConfigEditor.tsx`: `title="Custom HealthCheck"`, `enable = {label:"Enable custom health check", tooltip:"allow you to have custom health check messages"}`, `status = {label:"Custom status", tooltip:"custom health check status"}`, `message = {label:"Custom message", tooltip:"custom health check message. Leave blank for empty message"}`, `detail = {label:"Custom detail json", tooltip:"Leave blank to ignore. Format: { \"message\": \"your message\", \"verboseMessage\":\"detailed message\"}"}`, `skipBackend = {label:"Skip backend", tooltip:"allow you to skip backend call for health check"}`. |
| `src/types/config.types.ts:1-19` | Frontend `MockConfig = {customHealthCheckEnabled?, customHealthCheck?} & DataSourceJsonData`; `MockSecureConfig = Partial<Record<never, string>>` (empty — the plugin declares no plugin-owned secrets, `mockSecureConfigKeys = [] as const`); `CustomHealthCheck = {status, message?, details?, skipBackend?}`. |
| `src/datasource.ts:24-60` | `MockDS.testDatasource` overrides `DataSourceWithBackend.testDatasource`: when `jsonData.customHealthCheckEnabled && jsonData.customHealthCheck?.skipBackend`, it never calls the backend and returns a fake `{status, message, details?}` object synthesised from the same jsonData (`:32-58`). Details is JSON-parsed on the way through (`:48-55`) and a parse failure downgrades the response to `{status:'error', message:'invalid details json provided. Fix the custom details json message'}`. |
| `pkg/main.go:11-17` | `datasource.Manage("grafana-mock-datasource", client.New, …)` — confirms the plugin id. |
| `pkg/client/client.go:20-38` | Instance constructor. Builds the HTTP client via `setting.HTTPClientOptions(ctx)` (`:21`) then `httpclient.New` (`:25`), and calls `models.LoadSettings(ctx, setting)` (`:29`) for the plugin-owned jsonData. |
| `pkg/models/settings.go:1-28` | Backend `Config` — verbatim tags: `CustomHealthCheckEnabled bool json:"customHealthCheckEnabled"` (`:12`), `CustomHealthCheck CustomHealthCheckConfig json:"customHealthCheck"` (`:13`); `CustomHealthCheckConfig{Status int, Message string, Details string, SkipBackend bool}` (`:16-21`) with matching `json:` tags. `LoadSettings` is a straight `json.Unmarshal(settings.JSONData, &config)` (`:23-27`). |
| `pkg/client/handler_checkhealth.go:20-40` | Consumption of the CustomHealthCheck fields: `settings.CustomHealthCheck.Message` (falls back to `"health check message not specified"` on blank, `:27-29`), `Status: backend.HealthStatus(settings.CustomHealthCheck.Status)` (`:32`), and `JSONDetails: []byte(settings.CustomHealthCheck.Details)` (`:33`). `settings.CustomHealthCheck.SkipBackend` is parsed but never referenced anywhere in `pkg/`. |
| `pkg/client/handler_querydata.go`, `pkg/client/handler_callresource.go` | Additional server-side code paths — neither unmarshals `settings.JSONData` beyond what `client.New` already did. |
| `package.json` | External component versions (see next table). |

### External editor components

Read against the plugin-ui sources at HEAD in `github.com/grafana/plugin-ui`. The plugin pins `@grafana/plugin-ui@0.14.0` (an npm-only release train; the git repo tags stop at `v0.12.0` on `main`). Component contracts on the read paths below have been stable across the 0.12 → 0.13 → 0.14 line: labels, tooltips, and the storage keys they write are unchanged.

| Component | Package @ pin | Source consulted | What was read |
| --- | --- | --- | --- |
| `ConnectionSettings` | `@grafana/plugin-ui@0.14.0` | `grafana/plugin-ui` `src/components/ConfigEditor/Connection/ConnectionSettings.tsx:17-75` | URL label defaults to `"URL"` (`:41`); URL placeholder defaults to `"URL"` (`:69`) since `MockConfigEditor.tsx:20` does not override; required + built-in URL regex validation. |
| `Auth`, `AuthMethodSettings`, `BasicAuth` | `@grafana/plugin-ui@0.14.0` | `src/components/ConfigEditor/Auth/{Auth,auth-method/AuthMethodSettings,auth-method/BasicAuth}.tsx` | Default `visibleMethods = [BasicAuth, OAuthForward, NoAuth]` (`AuthMethodSettings.tsx:59-64`); option labels/descriptions from `AuthMethodSettings.tsx:9-32`; BasicAuth `User`/`Password` labels + placeholders + tooltips from `BasicAuth.tsx:24-29`. |
| `convertLegacyAuthProps` | `@grafana/plugin-ui@0.14.0` | `src/components/ConfigEditor/Auth/utils.ts:8-55` | Maps `basicAuth` / `withCredentials` / `jsonData.oauthPassThru` ↔ AuthMethod enum; `onAuthMethodSelect` writes basicAuth+withCredentials+oauthPassThru in one shot (`:44-54`). |
| `TLSSettings`, `SelfSignedCertificate`, `TLSClientAuth`, `SkipTLSVerification` | `@grafana/plugin-ui@0.14.0` | `src/components/ConfigEditor/Auth/tls/*.tsx` | All TLS labels/placeholders/rows come verbatim from these files (see [Field provenance](#field-provenance)); note the RSA private key placeholder typo `Begins with --- RSA PRIVATE KEY CERTIFICATE ---` at `TLSClientAuth.tsx:109` — shared across every plugin that uses `Auth`. |
| `CustomHeaders`, `CustomHeader` | `@grafana/plugin-ui@0.14.0` | `src/components/ConfigEditor/Auth/custom-headers/*.tsx` (via `utils.ts:188-233`) | Indexed `httpHeaderName<N>` / `httpHeaderValue<N>` storage pattern; **not modeled** in this schema (see [Excluded settings](#excluded-settings)). |
| `ConfigSection` | `@grafana/plugin-ui@0.14.0` | `src/components/ConfigEditor/ConfigSection.tsx` | Section title/description props (no storage keys — layout only). |
| `Stack`, `InlineFormLabel`, `RadioButtonGroup`, `Input`, `CodeEditor`, `InlineSwitch` | `@grafana/ui@12.4.3` | grafana/grafana `packages/grafana-ui/src/components/` | Prop shape for the CustomHealthCheck subeditor rows (no storage keys). |
| `DataSourceJsonData`, `DataSourceSettings`, `DataSourcePluginOptionsEditorProps` | `@grafana/data@12.4.3` | grafana/grafana `packages/grafana-data/src/` | Base jsonData interface; editor prop plumbing. |
| `DataSourceWithBackend` | `@grafana/runtime@12.4.3` | grafana/grafana `packages/grafana-runtime/src/` | Parent class of `MockDS`; provides `testDatasource`, `postResource`, `getResource` that the Mock overrides at `src/datasource.ts:30-146`. |
| `backend.HealthStatus` | `@grafana/grafana-plugin-sdk-go` | `backend/health.go` | Enum used to cast `CustomHealthCheck.Status` in `handler_checkhealth.go:32`. Silently ignores out-of-range int values, which is why this entry's `Validate` guards `{0,1,2}` explicitly. |

## Field provenance

Every dsconfig field, traced from its schema `id` to the upstream `file:line` where each of its label, tooltip, default, storage key, and value type is defined.

| Schema `id` | Storage key | Target | Label source | Tooltip / options / default source | Value type source | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `root_url` | `url` | `root` | `ConnectionSettings.tsx:41` (default `urlLabel = 'URL'`; Mock editor does not override) | `ConnectionSettings.tsx:69` (default placeholder `'URL'` because Mock passes no `urlPlaceholder`) | `settings.URL string` — SDK base | Not marked `requiredWhen` here: the Mock backend never dials the URL, so an empty URL is a valid config. The editor's plugin-ui component visually marks it required — surfaced via the field's presence in the connection group. |
| `virtual_authMethod` | — | virtual | Standard @grafana/plugin-ui Authentication method selector | Options from `AuthMethodSettings.tsx:9-32`; default `'NoAuth'` mirrors `getSelectedMethod` fallthrough (`utils.ts:37`) | Union of 3 strings | `storage.computed.read` mirrors `getSelectedMethod` (`utils.ts:27-38`) minus `CrossSiteCredentials` (not exposed here per default `visibleMethods`); `effects` mirror `onAuthMethodSelect` (`utils.ts:44-54`). |
| `root_basicAuth` | `basicAuth` | `root` | — (no UI; managed by `virtual_authMethod`) | Written by `utils.ts:47` | Root SDK bool | Tagged `managed-by:virtual_authMethod`. |
| `root_basicAuthUser` | `basicAuthUser` | `root` | `BasicAuth.tsx:24` (default `userLabel = 'User'`) | `BasicAuth.tsx:26` (default `userPlaceholder = 'User'`) | SDK `settings.BasicAuthUser string` | `dependsOn: virtual_authMethod == 'BasicAuth'`; `requiredWhen: root_basicAuth == true`. |
| `secureJsonData_basicAuthPassword` | `basicAuthPassword` | `secureJsonData` | `BasicAuth.tsx:27` (default `passwordLabel = 'Password'`) | `BasicAuth.tsx:29` (default `passwordPlaceholder = 'Password'`) | Role `auth.basic.password` | Same conditional/required as `basicAuthUser`. |
| `jsonData_oauthPassThru` | `oauthPassThru` | `jsonData` | — (no visible UI; controlled by `virtual_authMethod == 'OAuthForward'`) | Written by `utils.ts:51` | `bool` (@grafana/plugin-ui writes it) | Tagged `managed-by:virtual_authMethod`. |
| `jsonData_tlsAuthWithCACert` | `tlsAuthWithCACert` | `jsonData` | `SelfSignedCertificate.tsx:33` (`label="Add self-signed certificate"`) | `tooltipText` `SelfSignedCertificate.tsx:34`; default `false` | `bool` (SDK TLS pack) | — |
| `secureJsonData_tlsCACert` | `tlsCACert` | `secureJsonData` | `SelfSignedCertificate.tsx:39` (`label="CA Certificate"`) | `SelfSignedCertificate.tsx:54` (`placeholder="Begins with --- BEGIN CERTIFICATE ---"`, `rows: 6`); default tooltip `"Your self-signed certificate"` at `:41` | Role `tls.caCert` | `dependsOn` / `requiredWhen`: `jsonData_tlsAuthWithCACert == true`. |
| `jsonData_tlsAuth` | `tlsAuth` | `jsonData` | `TLSClientAuth.tsx:45` (`label="TLS Client Authentication"`) | `tooltipText` `TLSClientAuth.tsx:46` | `bool` | — |
| `jsonData_serverName` | `serverName` | `jsonData` | `TLSClientAuth.tsx:51` (`label="ServerName"`) | `TLSClientAuth.tsx:63` (`placeholder="domain.example.com"`); default tooltip at `:53` | Role `tls.serverName` | `dependsOn: jsonData_tlsAuth == true`; required for the mTLS contract. |
| `secureJsonData_tlsClientCert` | `tlsClientCert` | `secureJsonData` | `TLSClientAuth.tsx:70` (`label="Client Certificate"`) | `TLSClientAuth.tsx:88` (`placeholder="Begins with --- BEGIN CERTIFICATE ---"`, `rows: 6`); default tooltip at `:74` | Role `tls.clientCert` | `dependsOn`/`requiredWhen`: `jsonData_tlsAuth == true`. |
| `secureJsonData_tlsClientKey` | `tlsClientKey` | `secureJsonData` | `TLSClientAuth.tsx:94` (`label="Client Key"`) | `TLSClientAuth.tsx:109` (`placeholder="Begins with --- RSA PRIVATE KEY CERTIFICATE ---"` — upstream typo preserved, `rows: 6`); default tooltip at `:96` | Role `tls.clientKey` | Same conditional/required as `tlsClientCert`; see [Upstream findings](#upstream-findings) #1. |
| `jsonData_tlsSkipVerify` | `tlsSkipVerify` | `jsonData` | `SkipTLSVerification.tsx:14` (`label="Skip TLS certificate validation"`) | `tooltipText` `SkipTLSVerification.tsx:15` | Role `transport.tlsSkipVerify` | Default `false`. |
| `jsonData_customHealthCheckEnabled` | `customHealthCheckEnabled` | `jsonData` | `selectors.ts:6` (`label:"Enable custom health check"`) | `selectors.ts:7` (`tooltip:"allow you to have custom health check messages"`) | `bool` (`pkg/models/settings.go:12`) | Default `false`. Renders as `InlineSwitch` (`MockConfigEditor.tsx:33`). |
| `jsonData_customHealthCheck_status` | `status` (section `customHealthCheck`) | `jsonData` | `selectors.ts:10` (`label:"Custom status"`) | `selectors.ts:11` (`tooltip:"custom health check status"`); options `OK=1`, `ERROR=2`, `UNKNOWN=0` (`MockConfigEditor.tsx:71-73`) | `int` (`pkg/models/settings.go:17`) | Default `0` (UNKNOWN) mirrors the editor fallback (`MockConfigEditor.tsx:40,75`). `dependsOn: jsonData_customHealthCheckEnabled == true`. |
| `jsonData_customHealthCheck_message` | `message` (section `customHealthCheck`) | `jsonData` | `selectors.ts:14` (`label:"Custom message"`) | `selectors.ts:15` (`tooltip:"custom health check message. Leave blank for empty message"`) | `string` (`pkg/models/settings.go:18`) | Persisted on `onBlur` (`MockConfigEditor.tsx:86`). `dependsOn: jsonData_customHealthCheckEnabled == true`. |
| `jsonData_customHealthCheck_details` | `details` (section `customHealthCheck`) | `jsonData` | `selectors.ts:18` (`label:"Custom detail json"`) | `selectors.ts:19` (`tooltip:'Leave blank to ignore. Format: { "message": "your message", "verboseMessage":"detailed message"}'`) | `string` (`pkg/models/settings.go:19`) | Rendered by `CodeEditor language="json"` (`MockConfigEditor.tsx:94`); backend passes bytes through as `jsonDetails`. `dependsOn: jsonData_customHealthCheckEnabled == true`. |
| `jsonData_customHealthCheck_skipBackend` | `skipBackend` (section `customHealthCheck`) | `jsonData` | `selectors.ts:22` (`label:"Skip backend"`) | `selectors.ts:23` (`tooltip:"allow you to skip backend call for health check"`) | `bool` (`pkg/models/settings.go:20`) | Frontend-consumed only: `src/datasource.ts:32-58` uses it to short-circuit `testDatasource`. The backend parses it but never reads it. `dependsOn: jsonData_customHealthCheckEnabled == true`. |

## Field inventory summary

| Schema field | Storage key | Target | Editor label | Read by backend? |
| --- | --- | --- | --- | --- |
| `root_url` | `url` | `root` | URL | Indirect (via SDK `HTTPClientOptions`; the Mock code itself never dials the URL) |
| `virtual_authMethod` | — (virtual) | — | Authentication method | — (editor-local selector) |
| `root_basicAuth` | `basicAuth` | `root` | — (managed by virtual) | Indirect (SDK `HTTPClientOptions`) |
| `root_basicAuthUser` | `basicAuthUser` | `root` | User | Indirect (SDK) |
| `secureJsonData_basicAuthPassword` | `basicAuthPassword` | `secureJsonData` | Password | Indirect (SDK) |
| `jsonData_oauthPassThru` | `oauthPassThru` | `jsonData` | — (managed by virtual) | Indirect (SDK) |
| `jsonData_tlsAuthWithCACert` | `tlsAuthWithCACert` | `jsonData` | Add self-signed certificate | Indirect (SDK) |
| `secureJsonData_tlsCACert` | `tlsCACert` | `secureJsonData` | CA Certificate | Indirect (SDK) |
| `jsonData_tlsAuth` | `tlsAuth` | `jsonData` | TLS Client Authentication | Indirect (SDK) |
| `jsonData_serverName` | `serverName` | `jsonData` | ServerName | Indirect (SDK) |
| `secureJsonData_tlsClientCert` | `tlsClientCert` | `secureJsonData` | Client Certificate | Indirect (SDK) |
| `secureJsonData_tlsClientKey` | `tlsClientKey` | `secureJsonData` | Client Key | Indirect (SDK) |
| `jsonData_tlsSkipVerify` | `tlsSkipVerify` | `jsonData` | Skip TLS certificate validation | Indirect (SDK) |
| `jsonData_customHealthCheckEnabled` | `customHealthCheckEnabled` | `jsonData` | Enable custom health check | **Yes** (`pkg/client/handler_checkhealth.go:25`) |
| `jsonData_customHealthCheck_status` | `customHealthCheck.status` | `jsonData` | Custom status | **Yes** (`pkg/client/handler_checkhealth.go:32`) |
| `jsonData_customHealthCheck_message` | `customHealthCheck.message` | `jsonData` | Custom message | **Yes** (`pkg/client/handler_checkhealth.go:26-29`) |
| `jsonData_customHealthCheck_details` | `customHealthCheck.details` | `jsonData` | Custom detail json | **Yes** (`pkg/client/handler_checkhealth.go:33`) |
| `jsonData_customHealthCheck_skipBackend` | `customHealthCheck.skipBackend` | `jsonData` | Skip backend | No — **frontend-only consumer** (`src/datasource.ts:32-58`); parsed by backend but never read |

### Frontend-only settings

- `jsonData.customHealthCheck.skipBackend` — the frontend consumes it to short-circuit `testDatasource()`; the backend parses it into `CustomHealthCheckConfig.SkipBackend` but never reads that field anywhere in `pkg/`.

### Backend-only settings

None. Every field the backend reads is editor-visible.

### Excluded settings

- **Secure Socks Proxy** (`jsonData.enableSecureSocksProxy` and associated socks-proxy fields) — the Mock editor does not render `SecureSocksProxySettings` at all. Deliberately omitted per `AGENTS.md`.
- **Custom HTTP headers** (`@grafana/plugin-ui`'s `CustomHeaders`, exposed via `convertLegacyAuthProps` — see `utils.ts:188-233`) — the editor writes indexed pairs `jsonData.httpHeaderName<N>` / `secureJsonData.httpHeaderValue<N>` starting at index 1. Not modeled as a first-class field because the storage keys are dynamic. Downstream tools should walk `jsonData` for the `httpHeaderName` prefix and pair up matching `httpHeaderValue<N>` secrets; the SDK's `HTTPClientOptions` already does this.

## Where the types are defined

### Frontend (TypeScript)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `MockConfig` (`customHealthCheckEnabled?`, `customHealthCheck?` — extends `DataSourceJsonData`) | `src/types/config.types.ts:3-6` | plugin ([grafana/mock-datasource](https://github.com/grafana/mock-datasource)) |
| `CustomHealthCheck` (`status`, `message?`, `details?`, `skipBackend?`) | `src/types/config.types.ts:12-17` | plugin |
| `MockSecureConfig` (empty — derived from `mockSecureConfigKeys = [] as const`) | `src/types/config.types.ts:8-10` | plugin |
| `DataSourceJsonData` (base interface: `authType`, `defaultRegion`, `profile`, `manageAlerts`, `alertmanagerUid`, `disableGrafanaCache`) | `packages/grafana-data/src/types/datasource.ts` | `@grafana/data@12.4.3` |
| `DataSourcePluginOptionsEditorProps`, `DataSourceSettings` | `packages/grafana-data/src/` | `@grafana/data@12.4.3` |
| `ConnectionSettings`, `Auth`, `AuthMethod`, `convertLegacyAuthProps`, `ConfigSection` | `src/components/ConfigEditor/…` | `@grafana/plugin-ui@0.14.0` |
| `Stack`, `InlineFormLabel`, `RadioButtonGroup`, `Input`, `CodeEditor`, `InlineSwitch` | `packages/grafana-ui/src/components/` | `@grafana/ui@12.4.3` |
| `DataSourceWithBackend` | `packages/grafana-runtime/src/` | `@grafana/runtime@12.4.3` |

### Backend (Go)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `Config{CustomHealthCheckEnabled, CustomHealthCheck}` and `CustomHealthCheckConfig{Status, Message, Details, SkipBackend}` | `pkg/models/settings.go:11-21` | plugin |
| `LoadSettings(ctx, settings)` | `pkg/models/settings.go:23-27` | plugin |
| `Client{Config, HttpClient, StaticScenarios}` and `client.New` (calls `settings.HTTPClientOptions(ctx)` + `models.LoadSettings`) | `pkg/client/client.go:14-38` | plugin |
| `CheckHealth(ctx, settings, req)` (reads `settings.CustomHealthCheckEnabled` / `.CustomHealthCheck.*`) | `pkg/client/handler_checkhealth.go:20-40` | plugin |
| `backend.DataSourceInstanceSettings` (carries `URL`, `BasicAuthEnabled`, `BasicAuthUser`, `JSONData`, `DecryptedSecureJSONData`), `HTTPClientOptions(ctx)` | `backend/common.go`, `backend/httpclient/` | `github.com/grafana/grafana-plugin-sdk-go` |
| `backend.HealthStatus` enum (int-cast in `handler_checkhealth.go:32`) | `backend/health.go` | `github.com/grafana/grafana-plugin-sdk-go` |

The models in this entry flatten the above into a single Go `Config` type (root `URL`/`BasicAuth`/`BasicAuthUser`/`WithCredentials` tagged `json:"-"`, plus the plugin-owned CustomHealthCheck fields, plus the standard HTTP-settings jsonData fields the editor writes and the SDK reads, plus `DecryptedSecureJSONData`) and a `SecureJsonDataKey` typed constant list. `settings.ts` keeps the three canonical TypeScript types (`RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig`).

## Modeling decisions

- **Virtual auth method**: `convertLegacyAuthProps`'s `onAuthMethodSelect` (`@grafana/plugin-ui utils.ts:44-54`) writes three storage fields in one shot — `root.basicAuth`, `root.withCredentials`, and `jsonData.oauthPassThru`. That is the same virtual-selector pattern used by the Parca, Pyroscope, Loki, Tempo, and Prometheus entries. `withCredentials` is not in the Mock editor's default `visibleMethods`, so the virtual field's effects only write `basicAuth` and `oauthPassThru`.
- **Nested `customHealthCheck` object as a `section`**: the four leaves (`status`, `message`, `details`, `skipBackend`) are declared with `section: "customHealthCheck"` and their raw keys, so the generated OpenAPI settings spec renders `jsonData.customHealthCheck` as a nested object with typed sub-properties. The conformance test's section-aware struct walker matches these leaves against `CustomHealthCheckConfig`'s json tags.
- **`root.url` is not required**: the Mock backend never dials `settings.URL`. The plugin-ui `ConnectionSettings` component visually marks the field required, but there is no backend contract to enforce it, so the schema does not set `requiredWhen: "true"`. Provisioning payloads can omit the URL entirely.
- **`CustomHealthCheckStatus` typed constants**: `Status` is typed as `CustomHealthCheckStatus` (an `int` alias) with named constants `Unknown=0`, `OK=1`, `Error=2`. This mirrors the closed set of values the backend maps to `backend.HealthStatus` and the radio options the editor renders. `Validate()` guards the set explicitly because `backend.HealthStatus(other)` silently returns an unknown status without erroring.
- **`SkipBackend` retained on `Config`**: the field is frontend-consumed but the backend parses it verbatim (`pkg/models/settings.go:20`). Keeping it on the Go `Config` matches upstream parity and lets tooling detect the flag without special-casing frontend-only fields.
- **`ApplyDefaults` is a no-op**: the Mock editor writes nothing into jsonData on load. Zero values for all leaves match what the editor renders when `jsonData.customHealthCheck` is undefined (fallback `{status: 0}` at `MockConfigEditor.tsx:40`). `ApplyDefaults` intentionally does nothing so we don't clobber intentional zero values — the `TestApplyDefaults` test guards this.
- **TLS pair requirements**: `TLSClientAuth` and `SelfSignedCertificate` mark every field with `required` in the UI, but they only require the paired fields when the parent switch is on. Encoded as `dependsOn` + `requiredWhen` on each field, and enforced in `Validate()`.
- **Field ID naming convention**: IDs are prefixed with their storage target for easy discoverability — `root_`, `jsonData_`, or `secureJsonData_` (and `virtual_` for virtual fields, which have no storage target) — followed by the camelCase storage key. For section leaves, the schema uses `jsonData_<section>_<key>` (e.g. `jsonData_customHealthCheck_status`) so the ID reads left-to-right as `target → section → key`. The `key` property keeps the plugin's raw storage key at each level.
- **Flat `Config` in Go**: `settings.go` collapses jsonData fields and decrypted secrets onto a single `Config` struct. Root-level fields the editor and SDK both use (`URL`, `BasicAuth`, `BasicAuthUser`, `WithCredentials`) are carried with `json:"-"` tags so `LoadConfig` returns them alongside the jsonData shape. Base `DataSourceJsonData` fields (authType, defaultRegion, profile, alertmanagerUid, disableGrafanaCache) exist in Grafana core but are neither written by the Mock editor nor read by the Mock plugin, so they are omitted.
- **No `LegacyPluginID`**: `src/plugin.json` declares no `aliasIDs`. The plugin id is `grafana-mock-datasource` verbatim — that is the registry directory name, the `pluginType` in `dsconfig.json`, and the `PluginID` Go constant.
- **`SecureJsonDataConfig` is a key list**: secure values are write-only, so the secure type is just the array of secret key names (`basicAuthPassword`, `tlsCACert`, `tlsClientCert`, `tlsClientKey`); consumers read `secureJsonFields` to see what is configured. The Mock plugin itself declares no plugin-owned secrets (`mockSecureConfigKeys = [] as const`); the four keys tracked here are all written by @grafana/plugin-ui's `Auth`.

## SDK plugin schema and k8s-style examples (`schema.go`)

`NewSchema()` assembles the `grafana-plugin-sdk-go` `pluginschema.PluginSchema` bundle (the k8s-style schema Grafana's datasource API server serves as `{apiVersion}.json`, `v0alpha1` today) from the embedded `dsconfig.json`: root fields plus a nested `jsonData` object (including the nested `customHealthCheck` sub-object built from `section` fields) become the OpenAPI settings `spec`, secure fields become `secureValues`, and virtual fields are skipped.

`SettingsExamples()` provides the default configuration plus one k8s-style example per authentication method and each CustomHealthCheck variant. Each example is a full instance-settings object with the plugin configuration nested under `jsonData` and the relevant write-only secrets under `secureJsonData` (placeholder values to be replaced with real secrets; the default example — keyed by the empty string `""` — carries an empty `basicAuthPassword` to show that no secret is required for the default No-auth mode):

| Example | Auth | TLS | CustomHealthCheck | `secureJsonData` |
| --- | --- | --- | --- | --- |
| `""` (default) | None | — | — | `basicAuthPassword` (empty) |
| `noAuth` | None | — | — | `basicAuthPassword` (empty) |
| `basicAuth` | Basic | — | — | `basicAuthPassword` |
| `oauthForward` | OAuth Identity | — | — | `basicAuthPassword` (empty) |
| `tlsMutualAuth` | None | mTLS (serverName + client cert/key) | — | `tlsClientCert`, `tlsClientKey` |
| `tlsSelfSignedCA` | None | Custom CA | — | `tlsCACert` |
| `customHealthCheckError` | None | — | `status=2`, `message`, `details`, `skipBackend=false` | `basicAuthPassword` (empty) |
| `customHealthCheckSkipBackend` | None | — | `status=1`, `skipBackend=true` (frontend-only) | `basicAuthPassword` (empty) |

## `LoadConfig` utility (`settings.go`)

`LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error)` runs the full three-phase load flow on a datasource instance's settings and returns a fully-defaulted, validated `Config`:

1. **Parse** — copy `settings.URL`, `settings.BasicAuthEnabled`, `settings.BasicAuthUser` into `Config`, unmarshal `settings.JSONData` into the jsonData portion of the same struct (mirroring the plugin's own `LoadSettings` at `pkg/models/settings.go:23-27` — a straight `json.Unmarshal`), and copy the four decrypted secrets into `DecryptedSecureJSONData`.
2. **`ApplyDefaults`** — intentionally a no-op (see [Modeling decisions](#modeling-decisions)).
3. **`Validate`** — enforce the runtime contract: Basic auth requires a username, mTLS requires serverName + client cert + client key, custom-CA requires the CA PEM, and (when enabled) `customHealthCheck.status` must be one of {0, 1, 2}. Errors are joined so every problem surfaces at once.

Everything is logged via `backend.Logger.FromContext(ctx)` with `datasource_uid`, `datasource_name`, and `plugin` labels so log lines carry request context.

### Direct access to individual phases

`(*Config).ApplyDefaults()` and `(Config).Validate() error` are still exported for callers that want to compose them themselves (e.g. provisioning preview, schema-example round-trip, tests that need to distinguish parse-level from policy-level errors). Skip them by never calling `LoadConfig` in those flows — assemble a `Config` directly.

## Upstream findings

Potential bugs, misleading UX, and consistency issues discovered while researching upstream. All preserved verbatim in the schema — the schema records what the plugin **does**, not what it **should** do; these notes exist so reviewers can reproduce each finding and decide separately whether to fix upstream.

1. **Upstream placeholder typo preserved**: `@grafana/plugin-ui`'s `TLSClientAuth.tsx:109` sets the client key placeholder to `Begins with --- RSA PRIVATE KEY CERTIFICATE ---` — an RSA private key is not a "certificate". Preserved verbatim in `secureJsonData_tlsClientKey.ui.placeholder`. This is a plugin-ui typo shared across all data sources that use `Auth`.
2. **`skipBackend` is parsed but not read on the backend**: `pkg/models/settings.go:20` declares `SkipBackend bool json:"skipBackend"`, but nothing in `pkg/` ever references that field. The intent (short-circuit `testDatasource` when the frontend can synthesise a response) is entirely a frontend concern (`src/datasource.ts:32-58`). Dead field on the backend; kept here for parity.
3. **`backend.HealthStatus(int)` silently accepts unknown values**: `pkg/client/handler_checkhealth.go:32` casts `settings.CustomHealthCheck.Status` (a plain `int`) through `backend.HealthStatus(...)`. Out-of-range values produce an unknown health status instead of an error. This entry's `Validate` surfaces this at load time by requiring `{0, 1, 2}` explicitly, matching the three options the editor's radio group exposes.
4. **`ConfigSection title=""` wrappers around ConnectionSettings and Auth**: `MockConfigEditor.tsx:19,23` wraps the two library components in empty-titled `<ConfigSection>` elements. The library components already render their own titled section, so the outer wrappers only affect vertical spacing; they contribute no meaningful section titles. This entry uses the library-supplied section titles (`Connection`, `Authentication`) in the group model.
5. **No canonical docs URL**: `src/plugin.json` has no `info.links`, so this entry sets `docURL` to the repository URL. If Grafana ever publishes a docs page for the mock plugin, update `docURL` accordingly.
6. **Plugin-owned `MockSecureConfig` is empty**: `src/types/config.types.ts:8` declares `mockSecureConfigKeys = [] as const`, meaning the plugin itself defines no plugin-owned secrets. Every secure key modeled here comes from @grafana/plugin-ui's `Auth`. If the plugin ever adds its own secret, the `SecureJsonDataKey` list here must be extended (and the frontend `mockSecureConfigKeys` tuple in sync).
7. **npm-only plugin-ui version pin**: `package.json` pins `@grafana/plugin-ui@0.14.0`, but the plugin-ui git repository's tags stop at `v0.12.0` on `main`. The 0.13.x / 0.14.x lines are npm-only release trains. The external-component references above use plugin-ui's `main` sources; the labels/tooltips/storage keys on the code paths this plugin exercises have been stable across 0.12 → 0.14.

## Validation performed

- `dsconfig.ParseAndResolveSchemaJSON` + `Schema.Validate()` (Go validator in this repo) — passes (invoked by the conformance suite).
- JSON Schema validation against [`dsconfig/schema.json`](../../dsconfig/schema.json) (draft 2020-12, `additionalProperties: false`) — passes (invoked by the conformance suite).
- `go test ./...` inside `registry/` — passes on every entry, including the new `grafana-mock-datasource` package (schema bundle shape, secure values, examples, `LoadConfig` incl. TLS variants, CustomHealthCheck variants, and malformed input, `SchemaArtifactInSync` guard, `JSONDataMatchesStruct` and `JSONDataTypesMatchStruct` — including the nested `customHealthCheck.*` leaves).
- `settings.go` / `schema.go`: `go build`, `go vet`, `gofmt` — clean.
- `settings.ts`: exports the three canonical types (`RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig`) — reviewed by hand against the frontend sources; no `tsc` runner is wired into the registry module.
