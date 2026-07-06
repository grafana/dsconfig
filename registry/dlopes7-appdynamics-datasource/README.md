# dlopes7-appdynamics-datasource — dsconfig registry entry

Declarative configuration schema for the **AppDynamics** datasource plugin
(`dlopes7-appdynamics-datasource`). `dsconfig.json` is the single source of
truth; the Go and TypeScript models and the generated artifacts are derived
from it.

## Files

| File | Purpose |
| --- | --- |
| `dsconfig.json` | dsconfig v1 schema — the single source of truth. |
| `settings.ts` | TypeScript config models: `RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig`. |
| `settings.go` | Flat Go `Config` (jsonData + root `url`/`basicAuthUser` + `DecryptedSecureJSONData`), typed enums, `LoadConfig`/`ApplyDefaults`/`Validate`. |
| `schema.go` | Embeds `dsconfig.json`; `ConfigSchema()`, `NewSchema()`, `SettingsExamples()`. |
| `conformance_test.go` | `schema.RunPluginTests` wrapper (guard-rails + artifact generation). |
| `settings_test.go` | `LoadConfig` / `ApplyDefaults` / `Validate` / examples tests. |
| `schema.gen.json`, `settings.gen.json`, `settings.examples.gen.json` | Generated artifacts (`go generate ./...`). |

Import path: `github.com/grafana/dsconfig/registry/dlopes7-appdynamics-datasource`
(package `appdynamicsdatasource`).

## Source researched & how to reproduce

Researched against the **grafana/plugins-private** monorepo at commit
`267f4937806ed6404b6628d13ae358a5d308e376`, plugin path
`plugins/dlopes7-appdynamics-datasource/` (plugin version `3.12.5`,
`package.json:3`). All `file:line` references below are relative to that plugin
directory at that commit.

```sh
# In an existing plugins-private checkout:
git -C <plugins-private> fetch origin
git -C <plugins-private> checkout 267f4937806ed6404b6628d13ae358a5d308e376
# Plugin lives at plugins/dlopes7-appdynamics-datasource/
```

Note the provenance divergence: the plugin **id** is `dlopes7-appdynamics-datasource`
(`src/plugin.json:4`, originating from the community "dlopes7" plugin), while the
backend Go **module path** is `github.com/grafana/appdynamics-grafana-datasource`
(`go.mod`). The registry entry directory and `pluginType` use the plugin id
verbatim, as required.

### Frontend sources

- `src/plugin.json:3-4,24` — plugin `name` (`"AppDynamics"`), `id`
  (`"dlopes7-appdynamics-datasource"`), docs URL
  (`info.links[0].url` = `https://grafana.com/docs/plugins/dlopes7-appdynamics-datasource`).
  `backend: true` (`:9`); there are **no** `routes[]` (not a proxy plugin).
- `src/components/ConfigEditor.tsx`:
  - `96-105` — `DataSourceHttpSettings` (`@grafana/ui`), `defaultUrl` =
    `HTTP_URL_PLACEHOLDER`. Renders the Controller URL (root `url`), the Basic
    auth toggle (root `basicAuth`) + User (root `basicAuthUser`) + Password
    (`secureJsonData.basicAuthPassword`), and the Skip TLS Verify toggle
    (`jsonData.tlsSkipVerify`), plus generic TLS/header/withCredentials fields
    the backend ignores (see *Exclusions*).
  - `106-185` — "Metrics" `FieldSet`: **Client Name** (`107-115`,
    `jsonData.clientName`, placeholder `"Client Name"`, no tooltip), **Client
    Domain** (`116-124`, `jsonData.clientDomain`, placeholder `"Client Domain"`,
    no tooltip), **Client Secret** (`125-145`, `SecretInput`,
    `secureJsonData.clientSecret`, tooltip `129-132`, placeholder
    `"Paste the client secret here..."`), and the excluded Secure Socks Proxy
    switch (`146-184`).
  - `186-256` — "Analytics" `FieldSet`: **Analytics API URL** (`189-201`,
    `Select` `allowCustomValue isClearable`, `jsonData.analyticsURL`, tooltip
    `"The Analytics API URL"`), **Global Account Name** (`202-215`,
    `jsonData.globalAccountName`, tooltip
    `"The global account name, as shown in the Controller UI License page."`),
    **Analytics API Key** (`216-228`, `SecretInput`,
    `secureJsonData.analyticsAPIKey`, tooltip `"The Analytics API Key"`,
    placeholder `"Paste in Analytics API Key here..."`) + a collapsible **Help**
    drawer `Alert` titled `"API Key"` (`229-255`).
  - `28-32` — `ANALYTICS_URLS` (the three SaaS Events API options).
  - `40-63` — `analyticsUrlOptions`: adds `<protocol>//<host>:9080/` when the
    Controller URL has a port (on-prem Events Service), and the current
    `analyticsURL` if set.
- `src/types.ts:77-83` — `AppDOptions` (jsonData: `clientName`, `clientDomain`,
  `analyticsURL`, `globalAccountName`, `enableSecureSocksProxy`).
- `src/types.ts:85-89` — `AppDSecureJsonData` (`basicAuthPassword`,
  `clientSecret`, `analyticsAPIKey`).
- `src/components/selectors.ts:1,3` — `HTTP_URL_PLACEHOLDER = 'http://localhost:8086'`,
  `DEFAULT_ANALYTICS_URL = 'https://analytics.api.appdynamics.com'` (the latter
  is a dead constant — see *Upstream discrepancies*).

### Backend sources

- `pkg/models/settings.go`:
  - `14-33` — `Settings` struct (loaded shape). jsonData tags: `tlsSkipVerify`
    (`:15`), `clientName` (`:16`), `clientDomain` (`:17`), `analyticsURL`
    (`:26`), `globalAccountName` (`:27`). Untagged runtime/secret/root fields:
    `ClientSecret` (`:18`), `BasicAuthUsername`/`BasicAuthPassword` (`:20-21`),
    `MetricsURL`/`MetricsAuthorization` (`:23-24`), `AnalyticsAPIKey` (`:28`),
    `ProxyOptions` (`:30`), `Inputs` (`:32`).
  - `36-63` — `LoadSettings`: unmarshal jsonData (`:40`); `MetricsURL` from root
    `config.URL` (`:44`); **auth gating** — if `secureJsonData.clientSecret` is
    non-empty use it, else take `BasicAuthUsername` from root `config.BasicAuthUser`
    and `BasicAuthPassword` from `secureJsonData.basicAuthPassword` (`:46-51`);
    `AnalyticsAPIKey` from `secureJsonData.analyticsAPIKey` (`:53-55`);
    `ProxyOptions` from `config.HTTPClientOptions(ctx)` (`:57-61`).
- `pkg/appd/auth/auth_provider.go`:
  - `22-29` — untyped `iota` auth-type constants: `BasicAuth`, `APIClient`,
    `Unknown`.
  - `55-89` — `NewMetricsProvider`: **Basic auth** when
    `BasicAuthPassword != "" && BasicAuthUsername != ""` (`56-64`); **API Client**
    (OAuth2 client-credentials) when `ClientSecret != "" && MetricsURL != "" &&
    ClientName != "" && ClientDomain != ""` (`66-83`) — builds
    `client_id=clientName@clientDomain` (`:78`) and POSTs to
    `<url>/controller/api/oauth/access_token` (`:72`).
  - `174-177` — `GetBasicAuthFromUsernameAndPassword` (base64 `username:password`).
- `pkg/appd/health_diagnostics.go`:
  - `70-84` — `IsAnalyticsConfigured`: true only when `analyticsURL`,
    `globalAccountName` **and** `analyticsAPIKey` are all non-empty.
  - `87-135` — `CheckSettings` (the data contract): Controller URL required
    (`88-90`); at least one Controller auth method required (`92-95`); if any of
    `clientSecret`/`clientName`/`clientDomain` is set, **all three** are required
    (`98-114`); if either basic-auth field is set, **both** are required
    (`116-128`).
- `pkg/appd/analytics/client.go:38-56` — Analytics `Fetch`: parses
  `settings.AnalyticsURL` (`:39`) and sets headers `X-Events-API-Key` =
  `AnalyticsAPIKey` (`:53`) and `X-Events-API-AccountName` = `AccountName`
  (`:54`).
- `pkg/appd/client/client.go:29-57` — the plugin's HTTP client honors only
  `settings.TLSSkipVerify` → `InsecureSkipVerify` (`:39-41`) and
  `settings.ProxyOptions` (`:42`); all other TLS/header options are discarded.
- `pkg/appd/client/api_client.go:25-76` — Controller `Fetch`: path-joins request
  paths onto `settings.MetricsURL` and sends `Authorization` + `Content-Type`.
- `pkg/appd/datasource.go:59-64,72-76` — `GetInstance` → `LoadSettings`; `routes`
  map of Controller REST paths for the health/events/tiers/metric-names modules.
- `pkg/models/settings_test.go` — upstream `LoadSettings` tests.

### External components (versions from the workspace catalog)

The plugin's config-relevant `@grafana/*` deps use the `catalog:` protocol
(`package.json:79-84`); versions resolve from the monorepo catalog
(`.yarnrc.yml:14-26`). There are no plugin-local version pins.

- **`@grafana/ui@^11.6.7`** (`.yarnrc.yml:26`) — the config editor imports
  (`ConfigEditor.tsx:8-19`): `DataSourceHttpSettings` and its children render the
  root `url` (label `"URL"`, placeholder = `defaultUrl`), the `"Basic auth"`
  toggle (root `basicAuth`), the `"User"` field (root `basicAuthUser`,
  placeholder `"user"`) and `"Password"` field (`secureJsonData.basicAuthPassword`)
  shown only when Basic auth is enabled, and the `"Skip TLS Verify"` toggle
  (`jsonData.tlsSkipVerify`). Also `SecretInput`, `Input`, `Select`, `InlineField`,
  `InlineSwitch`, `FieldSet`, `Alert`, `Button`, `Icon`. (Labels are the standard
  `DataSourceHttpSettings` defaults at this version; the plugin does not override
  them.)
- **`@grafana/data@^11.6.7`** (`.yarnrc.yml:19`) — `DataSourceJsonData` (base of
  `AppDOptions`), `DataSourcePluginOptionsEditorProps`, `DataSourceSettings`,
  `SelectableValue`.
- **`@grafana/runtime@^11.6.7`** (`.yarnrc.yml:23`) — `config` /
  `GrafanaBootConfig` (feature-toggle read gating the excluded Secure Socks Proxy
  switch, `ConfigEditor.tsx:146,261-264`).
- **`@grafana/plugins-private-api-query`** (workspace dep, `package.json:81`,
  version `*`) — supplies the `ApiQuery` **query** type (`src/types.ts:2,40`) and
  the backend `models.Input`; it drives the Health/Events/Tiers/MetricNames query
  editors, **not** datasource configuration, so it contributes no config fields.

## Field inventory

| schema id | storage key | target | editor label | read by backend |
| --- | --- | --- | --- | --- |
| `root_url` | `url` | root | URL | yes — `MetricsURL` (`settings.go:44`) |
| `root_basicAuth` | `basicAuth` | root | Basic auth | no — editor enabler only¹ |
| `root_basicAuthUser` | `basicAuthUser` | root | User | yes — `BasicAuthUsername` (`settings.go:49`)² |
| `secureJsonData_basicAuthPassword` | `basicAuthPassword` | secureJsonData | Password | yes — `BasicAuthPassword` (`settings.go:50`)² |
| `jsonData_clientName` | `clientName` | jsonData | Client Name | yes — `ClientName` (`settings.go:16`); OAuth `client_id` (`auth_provider.go:78`) |
| `jsonData_clientDomain` | `clientDomain` | jsonData | Client Domain | yes — `ClientDomain` (`settings.go:17`) |
| `secureJsonData_clientSecret` | `clientSecret` | secureJsonData | Client Secret | yes — `ClientSecret` (`settings.go:46-47`) |
| `jsonData_analyticsURL` | `analyticsURL` | jsonData | Analytics API URL | yes — `AnalyticsURL` (`settings.go:26`) |
| `jsonData_globalAccountName` | `globalAccountName` | jsonData | Global Account Name | yes — `AccountName` (`settings.go:27`); `X-Events-API-AccountName` (`analytics/client.go:54`) |
| `secureJsonData_analyticsAPIKey` | `analyticsAPIKey` | secureJsonData | Analytics API Key | yes — `AnalyticsAPIKey` (`settings.go:53-55`); `X-Events-API-Key` (`analytics/client.go:53`) |
| `jsonData_tlsSkipVerify` | `tlsSkipVerify` | jsonData | Skip TLS Verify | yes — `TLSSkipVerify` → `InsecureSkipVerify` (`settings.go:15`, `client/client.go:40`) |

¹ `root_basicAuth` is the `DataSourceHttpSettings` enabler toggle. The backend
never reads the flag; it infers the auth method from which credentials are
present (`settings.go:46-51`, `auth_provider.go:55-89`). It is modeled because
it controls editor visibility of the User/Password fields.

² `root_basicAuthUser` / `secureJsonData_basicAuthPassword` are rendered by
`DataSourceHttpSettings` (labels `"User"` / `"Password"`). The backend reads
them **only when no `clientSecret` is set** (`settings.go:46-51`).

### Frontend-only settings

- `root_basicAuth` — written by `DataSourceHttpSettings`, never read by the
  backend (see note ¹). Modeled because it gates editor visibility.

### Backend-only settings

None. Every backend-read config field has a corresponding editor control. The
`Settings` struct's remaining fields are not configuration storage:
`MetricsAuthorization` (runtime-computed), `ProxyOptions` (derived from the
excluded Secure Socks Proxy field via `HTTPClientOptions`), and `Inputs`
(query-time data from `@grafana/plugins-private-api-query`).

## Modeling decisions

- **No auth-type discriminator; two runtime-inferred Controller methods.** Unlike
  `grafana-github-datasource` (`virtual_selectedLicense`) or
  `grafana-datadog-datasource` (`virtual_mode`), AppDynamics stores **no**
  discriminator field. The backend derives the Controller (Metrics) API auth
  method at load time from which credentials are present, with **`clientSecret`
  taking precedence over basic auth** (`settings.go:46-51`,
  `auth_provider.go:55-89`). No `virtual_*` field is introduced — there is no
  editor-local selector to model. The Go `Config.AuthMethod()` reproduces this
  precedence (`basic-auth` / `api-client` / `unknown`).
- **Root fields carried by `Config`.** The backend reads two root-level fields —
  `config.URL` → `MetricsURL` (`settings.go:44`) and `config.BasicAuthUser` →
  `BasicAuthUsername` (`settings.go:49`). Both are modeled with `target: "root"`
  and carried on the Go `Config` with `json:"-"`. `root_basicAuth` is modeled
  (editor enabler) but **not** carried on `Config`, because the backend does not
  read the flag.
- **API Client trio requiredness mirrors the health check.** `CheckSettings`
  requires **all three** of `clientName`/`clientDomain`/`clientSecret` whenever
  **any** is set (`health_diagnostics.go:98-114`). The schema encodes this
  symmetrically: `clientName.requiredWhen` and `clientDomain.requiredWhen`
  fire when the other jsonData field **or** `secureJsonData_clientSecret` is set,
  and `clientSecret.requiredWhen` fires when either jsonData field is set. Cross-
  referencing a `secureJsonData` field in `requiredWhen` follows the
  `grafana-dynatrace-datasource` precedent (`secureJsonData_apiToken` /
  `secureJsonData_platformToken`). Basic-auth requiredness is encoded as a
  `pair` — both `root_basicAuthUser` and `secureJsonData_basicAuthPassword` are
  `requiredWhen: root_basicAuth == true`.
- **`dependsOn` only on the basic-auth pair.** The editor renders every custom
  field unconditionally; only `DataSourceHttpSettings` hides User/Password until
  the Basic auth toggle is on. So `dependsOn: root_basicAuth == true` is set on
  `root_basicAuthUser` and `secureJsonData_basicAuthPassword`, and no other field
  uses `dependsOn`.
- **Analytics is optional and independent — not validated.** Analytics (Events)
  configuration (`analyticsURL` + `globalAccountName` + `analyticsAPIKey`) is
  entirely separate from Controller auth. An incomplete Analytics section is
  **silently disabled** (`IsAnalyticsConfigured`, `health_diagnostics.go:70-84`),
  not an error, so none of the three carry `requiredWhen` and the Analytics group
  is `optional: true`. `Config.IsAnalyticsConfigured()` mirrors the check but
  `Validate()` deliberately does not gate on it.
- **Help drawer → field `help`.** The collapsible Analytics API Key help `Alert`
  (`ConfigEditor.tsx:241-255`) is attached to `secureJsonData_analyticsAPIKey` as
  a `help` object (title `"API Key"`); its "Create API Key" button (which links
  to `<controller-url>/controller/#/location=ACCOUNT_ADMIN_API_CLIENTS`) is
  rendered as a markdown line.
- **`LoadConfig` = parse → `ApplyDefaults` → `Validate`.** `LoadConfig` mirrors
  `LoadSettings` verbatim (unmarshal jsonData, lift `MetricsURL` off root
  `config.URL`, copy decrypted secrets, apply the `clientSecret`-wins gating),
  then runs `ApplyDefaults` and `Validate`. `ApplyDefaults` is an intentional
  **no-op** — the editor persists no default values (there is no discriminator,
  and every field starts empty/false). Both helpers are exported so callers that
  assemble a `Config` directly can reuse the same contract.

### Exclusions

- **Secure Socks Proxy** (`jsonData.enableSecureSocksProxy`,
  `ConfigEditor.tsx:146-184`) — excluded per AGENTS.md. It is consumed
  transparently by the SDK's `HTTPClientOptions` proxy options
  (`settings.go:57-61`), never read by name.
- **Generic `DataSourceHttpSettings` TLS / header / access fields** — the widget
  also renders TLS client-cert settings (`jsonData.tlsAuth`,
  `tlsAuthWithCACert`, `serverName`, `secureJsonData.tlsCACert` /
  `tlsClientCert` / `tlsClientKey`), custom HTTP headers
  (`httpHeaderName<N>` / `httpHeaderValue<N>`), `withCredentials`, and the access
  mode. The plugin's HTTP client honors only `tlsSkipVerify` and the proxy
  options (`client/client.go:39-42`) — everything else is discarded (see
  *Upstream discrepancies*), so only `tlsSkipVerify` is modeled.

## Where the types are defined

**Frontend (plugin):**
- `AppDOptions` (extends `DataSourceJsonData`; jsonData) — `src/types.ts:77-83`.
- `AppDSecureJsonData` (secureJsonData) — `src/types.ts:85-89`.

**Backend (plugin):**
- `Settings` (loaded shape) — `pkg/models/settings.go:14-33`.
- Auth-type `iota` constants (`BasicAuth`, `APIClient`, `Unknown`) —
  `pkg/appd/auth/auth_provider.go:22-29`.

**Library / SDK config types:**
- `DataSourceJsonData` — `@grafana/data@^11.6.7` (base of `AppDOptions`).
- `DataSourceSettings` — `@grafana/data@^11.6.7` (the root settings model,
  including `url`, `basicAuth`, `basicAuthUser`, and `secureJsonData`, plus the
  generic HTTP settings written by `DataSourceHttpSettings`).
- `backend.DataSourceInstanceSettings` — `github.com/grafana/grafana-plugin-sdk-go`
  (source of `URL`, `BasicAuthUser`, `JSONData`, `DecryptedSecureJSONData` read by
  `LoadConfig` / `LoadSettings`).
- `tlsSkipVerify` — the jsonData key is part of the SDK HTTP-settings convention
  (written by `@grafana/ui`'s `DataSourceHttpSettings`); it is also declared on
  the plugin's own `Settings` (`pkg/models/settings.go:15`).

## Settings examples matrix

| Example key | Controller auth | `jsonData` | root | `secureJsonData` |
| --- | --- | --- | --- | --- |
| `""` (default) | none (fails validation) | `{}` | — | `clientSecret=""` |
| `apiClient` | API Client | `clientName`, `clientDomain` | `url` | `clientSecret` |
| `basicAuth` | Basic | `{}` | `url`, `basicAuth=true`, `basicAuthUser` | `basicAuthPassword` |
| `apiClientWithAnalytics` | API Client + Analytics | `clientName`, `clientDomain`, `analyticsURL`, `globalAccountName` | `url` | `clientSecret`, `analyticsAPIKey` |
| `basicAuthWithAnalytics` | Basic + Analytics | `analyticsURL`, `globalAccountName` | `url`, `basicAuth=true`, `basicAuthUser` | `basicAuthPassword`, `analyticsAPIKey` |

The `""` default example has an empty `clientSecret` placeholder and no `url`, so
it intentionally fails `LoadConfig`'s `Validate` step (covered by a test). All
secret values use obviously-fake `<…>` angle-bracket placeholders — never a
realistic token shape.

## Potential upstream bugs / discrepancies

- **Client Secret tooltip missing-space typo.** The tooltip is built by string
  concatenation without a separating space:
  `'…(basic) authentication' + 'Leave blank…'` →
  `"…(basic) authenticationLeave blank for username/password authentication."`
  (`ConfigEditor.tsx:130-131`). Preserved **verbatim** in the field
  `description`.
- **Two-layer auth precedence.** `NewMetricsProvider` checks basic auth **first**
  (`auth_provider.go:56`), but `LoadSettings` only populates the basic-auth
  fields when `clientSecret` is empty (`settings.go:46-51`). The net effect is
  "`clientSecret` wins", matching the Client Secret tooltip — but the runtime
  ordering is masked by the load-time gating. A datasource carrying **both** a
  `clientSecret` and a `basicAuthPassword` uses API Client auth; `LoadConfig`
  reproduces this (see `TestLoadConfigClientSecretOverridesBasicAuth`).
- **No stored discriminator.** Because the auth method is inferred, do not assume
  a missing field means "unconfigured": a datasource with only basic-auth fields
  is Basic auth; one with `clientName`/`clientDomain`/`clientSecret` is API
  Client auth.
- **Analytics silently disabled when incomplete.** If any of `analyticsURL` /
  `globalAccountName` / `analyticsAPIKey` is missing, `IsAnalyticsConfigured`
  returns false and Analytics is skipped with **no health error**
  (`health_diagnostics.go:70-84`); it does not fall back to Controller
  credentials.
- **`DEFAULT_ANALYTICS_URL` is dead code.** The constant exists
  (`selectors.ts:3`) but is never used; the editor seeds no default
  (`value={options.jsonData.analyticsURL ?? ''}`, `ConfigEditor.tsx:197`), so
  `analyticsURL` has no `defaultValue` in the schema.
- **On-prem Analytics URL heuristic.** When the Controller URL has an explicit
  port, the editor offers `<protocol>//<host>:9080/` as an Analytics API option
  (`ConfigEditor.tsx:44-51`) — the AppDynamics on-prem Events Service default
  port. Captured in the connection `instruction`.
- **TLS / headers rendered but ignored.** `DataSourceHttpSettings` renders TLS
  client-cert and custom-header fields, but the plugin's HTTP client
  (`client/client.go:29-44`) applies only `TLSSkipVerify` and `ProxyOptions`;
  the rest have no effect on requests. Only `tlsSkipVerify` is modeled.
- **Repo/module/id provenance divergence.** Plugin id `dlopes7-appdynamics-datasource`
  (`plugin.json:4`) vs Go module `github.com/grafana/appdynamics-grafana-datasource`
  (`go.mod`) vs the `info.links` "Repository" pointing at `grafana/grafana`
  (`plugin.json:26`). The entry uses the plugin id verbatim.

## Validation performed

- `dsconfig.ParseAndResolveSchemaJSON` + `Schema.Validate()` (via the conformance
  suite's `ConfigSchemaValid`).
- JSON Schema validation of `dsconfig.json` against `dsconfig/schema.json`
  (draft-07, `additionalProperties: false`, `ajv`) — **valid**.
- `go generate ./...` (regenerates the three `*.gen.json` artifacts).
- From `registry/`: `gofmt -l .` (clean), `go build ./...`, `go vet ./...`,
  `go test ./...` — all packages pass, including this entry's conformance suite
  (8 subtests) and `LoadConfig` / `ApplyDefaults` / `Validate` /
  `IsAnalyticsConfigured` / examples tests.
- The pre-existing `dsconfig` and `schema` workspace modules still build.
- `tsc --noEmit --strict settings.ts` (TypeScript 5) — passes.

## What `LoadConfig` guarantees

`LoadConfig(ctx, settings)` runs three phases and returns a fully-defaulted,
validated `Config` (contextual logging via `backend.Logger.FromContext(ctx)`):

1. **Parse** — unmarshal `settings.JSONData` into `Config`, lift the Controller
   URL off root `settings.URL` → `MetricsURL`, copy the decrypted secrets by
   known key into `DecryptedSecureJSONData`, and apply the `LoadSettings` auth
   gating: a non-empty `clientSecret` selects API Client auth and suppresses the
   basic-auth fields; otherwise root `BasicAuthUser` + `basicAuthPassword` select
   basic auth (`settings.go:46-51`).
2. **ApplyDefaults** — a no-op (the editor writes no defaults).
3. **Validate** — the `CheckSettings` contract: Controller URL required; at least
   one Controller auth method present; if any API Client field is set all three
   are required; if either basic-auth field is set both are required. Errors are
   joined. The optional Analytics API is **not** gated here.
