# grafana-datadog-datasource — dsconfig registry entry

Declarative configuration schema for the Grafana **Datadog** datasource plugin
(`grafana-datadog-datasource`). `dsconfig.json` is the single source of truth;
the Go and TypeScript models and the generated artifacts are derived from it.

## Files

| File | Purpose |
| --- | --- |
| `dsconfig.json` | dsconfig v1 schema — the single source of truth. |
| `settings.ts` | TypeScript config models: `RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig`. |
| `settings.go` | Flat Go `Config` (jsonData + root `basicAuth`/`basicAuthUser` + `DecryptedSecureJSONData`), typed enums, `LoadConfig`/`ApplyDefaults`/`Validate`. |
| `schema.go` | Embeds `dsconfig.json`; `ConfigSchema()`, `NewSchema()`, `SettingsExamples()`. |
| `conformance_test.go` | `schema.RunPluginTests` wrapper (guard-rails + artifact generation). |
| `settings_test.go` | `LoadConfig` / `ApplyDefaults` / `Validate` / examples tests. |
| `schema.gen.json`, `settings.gen.json`, `settings.examples.gen.json` | Generated artifacts (`go generate ./...`). |

Import path: `github.com/grafana/dsconfig/registry/grafana-datadog-datasource`
(package `datadogdatasource`).

## Source researched & how to reproduce

Researched against the **grafana/plugins-private** monorepo at commit
`267f4937806ed6404b6628d13ae358a5d308e376`, plugin path
`plugins/grafana-datadog-datasource/` (plugin version `3.17.4`,
`package.json:4`). All `file:line` references below are relative to that plugin
directory at that commit.

```sh
# In an existing plugins-private checkout:
git -C <plugins-private> fetch origin
git -C <plugins-private> checkout 267f4937806ed6404b6628d13ae358a5d308e376
# Plugin lives at plugins/grafana-datadog-datasource/
```

### Frontend sources

- `src/plugin.json:3-5,31` — plugin `name` (`"Datadog"`), `id`
  (`"grafana-datadog-datasource"`), docs URL
  (`info.links[0].url` = `https://grafana.com/docs/plugins/grafana-datadog-datasource`).
- `src/components/ConfigEditor.tsx`:
  - `41-68` — one-time migration of legacy `jsonData.api_key` / `jsonData.app_key`
    into `secureJsonData.apiKey` / `secureJsonData.appKey`.
  - `74-88`, `164-171`, `196-203` — `secureJsonData.apiKey` / `secureJsonData.appKey`
    writes (SecretInput).
  - `134-222` — `Auth` (`@grafana/plugin-ui`) with a single visible method driven
    by mode: BasicAuth in hosted-metrics mode, custom `API_AND_APP_KEY` method in
    default mode (`138-140`); `onAuthMethodSelect` is a **no-op** (`137`).
  - `237-331` — `ConnectionEditor`: `Mode` RadioButtonGroup (`285-293`),
    hosted-metrics URL `Input` (`295-310`), default-mode region `Select` with
    `allowCustomValue` (`311-327`); `onModeChange` writes `jsonData.pluginMode`,
    toggles root `basicAuth`, and swaps `jsonData.url` (`254-265`).
  - `333-502` — `AdditionalSettingsEditor`: `logApiRateLimits` (`385-391`),
    `rateLimitEnabled` (`396-402`), `rateLimitMetrics` (min 0 max 100, shown only
    when `rateLimitEnabled`; `>100` clamped on blur at `348-352`, `406-425`),
    `disableDataLinks` (`431-437`), `size` (`445-461`), Secure Socks Proxy switch
    (`463-498`).
  - `504-525` — `getEndpointUrl` (region-aware "get API/App key" links).
  - `527-535` — `getPluginMode(jsonData, basicAuth)`: use `jsonData.pluginMode`
    if set, else `hosted-metrics` when root `basicAuth`, else `default`.
- `src/components/tooltips.tsx:3-137` — `ConfigComponentProps`: every label,
  placeholder, and tooltip the editor renders.
- `src/constants.ts:6-12` — `regions` (US1/Default, US3, US5, EU, US1-FED).
- `src/types.ts:193-220` — `pluginMode`, `DataDogJsonData`, `SecureSettings`.

### Backend sources

- `pkg/models/settings.go`:
  - `12-17` — `PluginMode` type + constants (`default`, `hosted-metrics`).
  - `19-36` — `Settings` struct (loaded shape).
  - `38-82` — `LoadSettings`: default seed → unmarshal → legacy migration →
    field mapping → `getPluginMode`.
  - `52-54`, `114-118` — `migrateToSecureKey` (legacy `api_key`/`app_key`).
  - `56-72` — field mapping (`URL` from `jsonBody.URL`; secrets from
    `secureSettings`; root `BasicAuthEnabled`/`BasicAuthUser`).
  - `63-65` — `RateLimitMetrics` coerced to 100 when enabled and 0.
  - `84-92` — `getPluginMode` (legacy `basicAuth` → hosted-metrics fallback).
  - `94-105` — internal `jsonSettings` parse struct.
  - `107-112` — `defaultJSONSettings` (url → `DefaultDatadogAPIURL`, size → 100).
  - `120-133` — `boolMaybeQuoted` lenient bool parser.
- `pkg/models/constants.go:4,7` — `DefaultDatadogAPIURL = "https://api.datadoghq.com"`,
  `DefaultDatadogAPIResponseSize = 100`.
- `pkg/datadog/health_diagnostics.go:22-27,81-105` — `CheckSettings`: default mode
  requires apiKey (`83-85`) + appKey (`86-88`); hosted-metrics requires url ≠
  default (`91-93`) + basic-auth username (`94-96`) + password (`97-99`); url
  required in all modes (`101-103`).
- `pkg/datadog/client/client_v1.go:52-74,85-99,118-120,222-248` — custom server
  URL, `DD-API-KEY` / `DD-APPLICATION-KEY` headers (`97-99`), URL userinfo basic
  auth in hosted-metrics mode (`85-87`), `/api/v1` path join (`118-120`),
  `dd.ContextAPIKeys` + `dd.ContextBasicAuth` wiring (`222-248`).
- `pkg/datadog/client/http_client.go:36-51` — the plugin's own HTTP client:
  fixed `hcOptions`; only `ProxyOptions` is taken from the SDK client options.
- `pkg/grafdog/client.go:29-41` — GrafDog client: same DD headers + hosted-metrics
  URL userinfo.
- `pkg/datadog/datasource.go:48-80` — `NewInstance` → `LoadSettings`.
- `pkg/models/settings_test.go:16-184` — upstream `LoadSettings` tests (default,
  size, direct mode, quoted booleans, hosted metrics, legacy migration).

### External components (versions from the workspace catalog)

Resolved via the `catalog:` protocol in `package.json:34-51` against
`.yarnrc.yml:14-26`:

- **`@grafana/plugin-ui@^0.13.1`** —
  - `Auth`, `convertLegacyAuthProps`
    (`dist/esm/components/ConfigEditor/Auth/utils.js`): `getSelectedMethod` maps
    root `basicAuth` → BasicAuth; `getBasicAuthProps` writes root `basicAuthUser`
    and `secureJsonData.basicAuthPassword`; `getTLSProps` / `getCustomHeaders`
    write generic TLS + `httpHeaderName<N>`/`httpHeaderValue<N>` fields.
  - `BasicAuth`
    (`dist/esm/components/ConfigEditor/Auth/auth-method/BasicAuth.js`): default
    `userLabel`/`passwordLabel` = **`"User"`** / **`"Password"`**, placeholders
    `"User"` / `"Password"`. The Datadog editor overrides only the **tooltips**
    with the hosted-metrics help text (`ConfigEditor.tsx:141-145`).
  - `ConfigSection`, `ConfigSubSection`, `DataSourceDescription`.
- **`@grafana/ui@^11.6.7`** — `Input`, `Select`, `RadioButtonGroup`, `Checkbox`,
  `SecretInput`, `InlineField`, `Switch`, `Tooltip`, `Icon`, `LinkButton`.
- **`@grafana/data@^11.6.7`** — `DataSourceJsonData` (base of `DataDogJsonData`),
  `DataSourcePluginOptionsEditorProps`, `DataSourceSettings`, `SelectableValue`,
  `FeatureToggles`.
- **`@grafana/runtime@^11.6.7`** — `config` (feature-toggle read gating the
  Secure Socks Proxy switch).

## Field inventory

| schema id | storage key | target | editor label | read by backend |
| --- | --- | --- | --- | --- |
| `virtual_mode` | `mode` (virtual) | — | Mode | derived (`getPluginMode`) |
| `jsonData_pluginMode` | `pluginMode` | jsonData | (managed by Mode) | yes — `getPluginMode` (`settings.go:84-92`) |
| `root_basicAuth` | `basicAuth` | root | (managed by Mode) | yes — `BasicAuthEnabled` (`settings.go:68`) |
| `jsonData_url` | `url` | jsonData | API URL / Region¹ | yes — `settings.URL` (`settings.go:56`) |
| `secureJsonData_apiKey` | `apiKey` | secureJsonData | API key | yes — `DD-API-KEY` (`client_v1.go:98`) |
| `secureJsonData_appKey` | `appKey` | secureJsonData | App key | yes — `DD-APPLICATION-KEY` (`client_v1.go:99`) |
| `root_basicAuthUser` | `basicAuthUser` | root | User² | yes — `BasicAuthUser` (`settings.go:69`) |
| `secureJsonData_basicAuthPassword` | `basicAuthPassword` | secureJsonData | Password² | yes — `basicAuthPassword` (`settings.go:70`) |
| `jsonData_logApiRateLimits` | `logApiRateLimits` | jsonData | Show API rate limits | yes — `ShowAPIRateLimits` (`settings.go:59`) |
| `jsonData_rateLimitEnabled` | `rateLimitEnabled` | jsonData | Enable API rate limit threshold | yes — `RateLimitEnabled` (`settings.go:60`) |
| `jsonData_rateLimitMetrics` | `rateLimitMetrics` | jsonData | API rate limit threshold % | yes — `RateLimitMetrics` (`settings.go:61-65`) |
| `jsonData_disableDataLinks` | `disableDataLinks` | jsonData | Disable data links | yes — `DisableDataLinks` (`settings.go:62`) |
| `jsonData_size` | `size` | jsonData | Response Size | yes — `Size` (`settings.go:66`) |

¹ `jsonData_url` label is **"API URL / Region"** in default mode (a region
`Select` with `allowCustomValue`) and **"Hosted metrics URL"** in hosted-metrics
mode (an `Input`). Both write the same `jsonData.url` key. See *Modeling
decisions*.

² `root_basicAuthUser` and `secureJsonData_basicAuthPassword` are rendered by
`@grafana/plugin-ui`'s `BasicAuth`; labels/placeholders are its defaults
(`"User"` / `"Password"`). The Datadog editor overrides only the tooltips with
the hosted-metrics help text (`ConfigEditor.tsx:141-145`).

### Frontend-only settings

None. Every modeled `jsonData` field is read by the backend `LoadSettings`.

### Backend-only settings

None. Every backend-read field has a corresponding editor control.

### Legacy / migrated keys (not modeled as schema fields)

- `jsonData.api_key`, `jsonData.app_key` — 1.x plaintext credentials. On load,
  both the editor (`ConfigEditor.tsx:41-68`) and the backend
  (`settings.go:52-54,114-118`) migrate them into `secureJsonData.apiKey` /
  `secureJsonData.appKey` and stop reading the jsonData copies. `LoadConfig`
  mirrors this migration; the keys are not persisted on `Config`.
- `jsonData.cacheInterval`, `jsonData.cacheSize`, `jsonData.naming_strategy` —
  dead 1.x keys, kept only as a comment in `src/types.ts:205-213`.

## Modeling decisions

- **Mode / auth as a virtual discriminator.** The editor's `Mode` radio value is
  derived (`getPluginMode`, `ConfigEditor.tsx:527-535`) and selecting it performs
  a multi-field write. It is modeled as `virtual_mode` (kind `virtual`) with a
  `storage.computed.read` mirroring `getPluginMode`
  (`jsonData.pluginMode` → else root `basicAuth` → else `default`) and `effects`
  that set `jsonData_pluginMode` + `root_basicAuth`. The real storage
  discriminator `jsonData_pluginMode` (role `auth.discriminator`) and the legacy
  signal `root_basicAuth` (role `auth.basic.enabled`) are tagged
  `managed-by:virtual_mode`. This matches the gold-standard
  `grafana-github-datasource` (`virtual_selectedLicense`) and
  `marcusolsson-csv-datasource` (`virtual_authMethod`) patterns.
- **`dependsOn` vs `requiredWhen`.** Credential visibility uses `dependsOn` on
  the virtual selector (`virtual_mode == 'default'` / `== 'hosted-metrics'`),
  mirroring editor visibility; requiredness uses `requiredWhen` on the storage
  discriminator (`jsonData_pluginMode != 'hosted-metrics'` /
  `== 'hosted-metrics'`), mirroring the backend contract in `CheckSettings`.
- **`jsonData_url` modeled as a free-text `input`.** The editor renders this key
  two different ways — a region `Select` (`allowCustomValue`) with options
  US1/Default (`https://api.datadoghq.com`), US3, US5, EU
  (`https://api.datadoghq.eu`), US1-FED in default mode, and a plain `Input` in
  hosted-metrics mode. Because both write the same `jsonData.url` and the
  hosted-metrics value is an arbitrary Grafana Cloud proxy URL, the schema models
  `url` as a **free-text `input`** (no `enum`). Encoding the default-mode region
  options as `ui.options` would make the SDK converter emit a hard OpenAPI `enum`
  (see `dsconfig/convert.go:applyUIEnum`) that wrongly rejects hosted-metrics and
  custom API URLs in the served settings spec. The region choices are preserved
  in the connection `instruction` and this README. A mode-dependent `override`
  supplies the hosted-metrics placeholder/description; `FieldOverride` cannot
  change the label, so the base label stays "API URL / Region" (default mode).
- **Root fields carried by `Config`.** Unlike most datasources, the Datadog
  backend reads two root-level fields — `config.BasicAuthEnabled` and
  `config.BasicAuthUser` (`settings.go:68-69`). They are modeled with
  `target: "root"` and carried on the Go `Config` with `json:"-"`. The root
  `url` is **not** read (the backend uses `jsonData.url`), so it is not carried.
- **Lenient boolean parsing.** `logApiRateLimits`, `rateLimitEnabled`, and
  `disableDataLinks` use the upstream `boolMaybeQuoted` type (accepts `true` or
  `"true"`) so `LoadConfig` parses the same encodings as `LoadSettings`. The Go
  kind is still `bool`, so the conformance type-parity check maps them to the
  `boolean` value type.
- **`LoadConfig` = parse → `ApplyDefaults` → `Validate`.** `LoadConfig` seeds the
  url/size defaults (mirroring `defaultJSONSettings`), unmarshals jsonData,
  migrates legacy `api_key`/`app_key`, copies decrypted secrets, and lifts the
  root basic-auth fields, then applies curated defaults (pluginMode via the
  `basicAuth` fallback, url, size, and the `rateLimitMetrics` 0→100 coercion) and
  validates the health-check contract. `ApplyDefaults` and `Validate` are exported
  for callers that assemble a `Config` directly.

### Exclusions

- **Secure Socks Proxy** (`jsonData.enableSecureSocksProxy`,
  `ConfigEditor.tsx:463-498`) — excluded per AGENTS.md.
- **Generic SDK TLS / custom-header fields** — the `@grafana/plugin-ui` `Auth`
  widget also renders a TLS settings section and custom HTTP headers (via
  `convertLegacyAuthProps` → `getTLSProps` / `getCustomHeaders`), writing
  `jsonData.tlsAuth`, `tlsAuthWithCACert`, `tlsSkipVerify`, `serverName`,
  `secureJsonData.tlsCACert` / `tlsClientCert` / `tlsClientKey`, and indexed
  `httpHeaderName<N>` / `httpHeaderValue<N>`. These are generic SDK fields, not
  Datadog config, and the Datadog backend's custom HTTP client **ignores them**
  (see *Upstream discrepancies*). They are excluded to keep the entry focused on
  the plugin's real config surface.

## Where the types are defined

**Frontend (plugin):**
- `pluginMode` — `src/types.ts:193`.
- `DataDogJsonData` (extends `DataSourceJsonData`) — `src/types.ts:195-214`.
- `SecureSettings` (`apiKey`, `appKey`, `basicAuthPassword`) — `src/types.ts:216-220`.

**Backend (plugin):**
- `PluginMode` + constants — `pkg/models/settings.go:12-17`.
- `Settings` (loaded shape) — `pkg/models/settings.go:19-36`.
- `jsonSettings` (jsonData parse shape, incl. legacy `api_key`/`app_key`) —
  `pkg/models/settings.go:94-105`.
- `boolMaybeQuoted` — `pkg/models/settings.go:120`.
- `DefaultDatadogAPIURL`, `DefaultDatadogAPIResponseSize` —
  `pkg/models/constants.go:4,7`.

**Library / SDK config types:**
- `DataSourceJsonData`, `DataSourceSettings` — `@grafana/data@^11.6.7` (base of
  the frontend jsonData + root settings model, including root `basicAuth` /
  `basicAuthUser` and `secureJsonData`).
- `backend.DataSourceInstanceSettings` — `github.com/grafana/grafana-plugin-sdk-go`
  (source of `BasicAuthEnabled`, `BasicAuthUser`, `JSONData`,
  `DecryptedSecureJSONData` read by `LoadConfig` / `LoadSettings`).

## Settings examples matrix

| Example key | Mode | `jsonData` | root | `secureJsonData` |
| --- | --- | --- | --- | --- |
| `""` (default) | default | `pluginMode=default`, `url=https://api.datadoghq.com` | — | `apiKey=""`, `appKey=""` |
| `directApiAppKey` | default (US1) | `pluginMode=default`, `url=https://api.datadoghq.com` | — | `apiKey`, `appKey` |
| `directApiAppKeyEU` | default (EU) | `pluginMode=default`, `url=https://api.datadoghq.eu` | — | `apiKey`, `appKey` |
| `hostedMetrics` | hosted-metrics | `pluginMode=hosted-metrics`, `url=…grafana.net/datadog` | `basicAuth=true`, `basicAuthUser` | `basicAuthPassword` |
| `legacyHostedMetricsNoPluginMode` | hosted-metrics (legacy) | `url=…grafana.net/datadog` (no `pluginMode`) | `basicAuth=true`, `basicAuthUser` | `basicAuthPassword` |

The `""` default example has empty secret placeholders, so it intentionally
fails `LoadConfig`'s `Validate` step (covered by a test). All secret values use
obviously-fake `<…>` angle-bracket placeholders. The legacy `api_key`/`app_key`
jsonData migration is covered by inline `settings_test.go` cases rather than an
example (the migrated values live in jsonData, not secureJsonData).

## Potential upstream bugs / discrepancies

- **TLS settings are rendered but ignored by the backend.** The `Auth` widget
  renders a TLS settings section (`ConfigEditor.tsx:221`, `TLS={TLS}`), but the
  plugin's own HTTP client (`pkg/datadog/client/http_client.go:36-51`) builds
  from a fixed `hcOptions` and only copies `ProxyOptions` from the SDK client
  options — the TLS config derived from `jsonData.tlsAuth` / `tlsSkipVerify` /
  etc. is discarded. Configuring TLS in the editor has no effect on Datadog API
  requests.
- **DD-API-KEY / DD-APPLICATION-KEY headers are always sent.** Even in
  hosted-metrics mode, `client_v1.go:97-99` sets both header values (possibly
  empty) in addition to the URL userinfo basic auth. Harmless, but the API-key
  headers are not validated in hosted-metrics mode.
- **Inconsistent docs anchors in tooltips.** The API-key tooltip links to
  `…/latest/#get-an-api-key-and-application-key-from-datadog`
  (`tooltips.tsx:57`) while the App-key tooltip links to
  `…/latest/#get-api-key-and-application-key-from-datadog` (`tooltips.tsx:34`) —
  note "get-**an**-api-key" vs "get-api-key". Both are preserved verbatim in the
  field descriptions.
- **`plugin.json` description typo.** `info.description` reads "Datadog datasource
  plugin for Grafana. Grafana Data dog datasource plugin" (`plugin.json:20`).
- **`rateLimitMetrics` clamping is one-sided in the editor.** The editor clamps
  values `>100` to 100 on blur (`ConfigEditor.tsx:348-352`) but does not clamp
  negatives; the backend separately coerces `0` → `100` when
  `rateLimitEnabled` (`settings.go:63-65`). The schema encodes a `range` 0–100
  validation to capture the intended bounds.
- **Health-check requiredness vs editor markers.** In default mode both apiKey
  and appKey are backend-required (`health_diagnostics.go:82-89`) and the editor
  marks the fields `required`; in hosted-metrics mode a non-default url + basic
  auth username + password are backend-required (`health_diagnostics.go:90-100`).
  `requiredWhen` encodes the backend contract for all four.

## Validation performed

- `dsconfig.ParseAndResolveSchemaJSON` + `Schema.Validate()` (via the conformance
  suite's `ConfigSchemaValid`).
- JSON Schema validation of `dsconfig.json` against `dsconfig/schema.json`
  (draft-07, `additionalProperties: false`) — passes.
- `go generate ./...` (regenerates the three `*.gen.json` artifacts).
- From `registry/`: `gofmt -l .` (clean), `go build ./...`, `go vet ./...`,
  `go test ./...` (all 52 packages pass, including this entry's conformance +
  `LoadConfig`/`ApplyDefaults`/`Validate`/examples tests).
- The pre-existing `dsconfig` and `schema` workspace modules still build and test.
- `tsc --noEmit --strict settings.ts` (TypeScript 5) — passes.

## What `LoadConfig` guarantees

`LoadConfig(ctx, settings)` runs three phases and returns a fully-defaulted,
validated `Config`:

1. **Parse** — seed `url`/`size` defaults (mirroring `defaultJSONSettings`),
   unmarshal `settings.JSONData` into `Config` (lenient `boolMaybeQuoted`
   booleans), migrate legacy `api_key`/`app_key` into `secureJsonData` when the
   modern secret is unset, copy decrypted secrets by known key, and lift root
   `BasicAuthEnabled` / `BasicAuthUser`.
2. **ApplyDefaults** — `PluginMode` (hosted-metrics when root `basicAuth` and
   unset, else default), `url` → `DefaultDatadogAPIURL`, `size` → 100,
   `rateLimitMetrics` → 100 when enabled and 0.
3. **Validate** — the `CheckSettings` contract: default mode needs apiKey +
   appKey; hosted-metrics needs a non-default url + basic-auth username +
   password; url required in all modes. Errors are joined.
