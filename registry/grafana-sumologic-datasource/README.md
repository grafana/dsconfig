# grafana-sumologic-datasource — dsconfig registry entry

Declarative configuration schema for the Grafana **Sumo Logic** datasource plugin
(`grafana-sumologic-datasource`). `dsconfig.json` is the single source of truth;
the Go and TypeScript models and the generated artifacts are derived from it.

## Files

| File | Purpose |
| --- | --- |
| `dsconfig.json` | dsconfig v1 schema — the single source of truth. |
| `settings.ts` | TypeScript config models: `RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig`. |
| `settings.go` | Flat Go `Config` (jsonData fields + `DecryptedSecureJSONData`), `PluginID`, typed `AuthenticationMethod` / `SecureJsonDataKey`, `LoadConfig` / `ApplyDefaults` / `Validate`. |
| `schema.go` | Embeds `dsconfig.json`; `ConfigSchema()`, `NewSchema()`, `SettingsExamples()`. |
| `conformance_test.go` | `schema.RunPluginTests` wrapper (guard-rails + artifact generation). |
| `settings_test.go` | `LoadConfig` / `ApplyDefaults` / `Validate` / examples tests. |
| `schema.gen.json`, `settings.gen.json`, `settings.examples.gen.json` | Generated artifacts (`go generate ./...`). |

Import path: `github.com/grafana/dsconfig/registry/grafana-sumologic-datasource`
(package `sumologicdatasource`). There is no per-entry `go.mod` — every registry
entry is a subpackage of the shared `registry/` module.

## Source researched & how to reproduce

Researched against the **grafana/plugins-private** monorepo at commit
`267f4937806ed6404b6628d13ae358a5d308e376`, plugin path
`plugins/grafana-sumologic-datasource/` (plugin version `1.6.17`,
`package.json:3`; Go module `github.com/grafana/sumologic-datasource`,
`go.mod:1`). All `file:line` references below are relative to that plugin
directory at that commit.

```sh
# In an existing plugins-private checkout:
git -C <plugins-private> fetch origin
git -C <plugins-private> checkout 267f4937806ed6404b6628d13ae358a5d308e376
# Plugin lives at plugins/grafana-sumologic-datasource/
```

### Frontend sources

- `src/plugin.json:3-4,25` — plugin `id` (`"grafana-sumologic-datasource"`),
  `name` (`"Sumo Logic"`), docs URL
  (`info.links[0].url` = `https://grafana.com/docs/plugins/grafana-sumologic-datasource`).
- `src/editor/ConfigEditor.tsx`:
  - `71-75` — `DataSourceDescription` (`dataSourceName="SumoLogic"`,
    `hasRequiredFields`).
  - `79-133` — `ConfigSection title="API Region"` containing:
    `apiUrl` `Select` (`allowCustomValue`, `isClearable`, options from
    `ApiEndpointOptions`; label `82`, control `100-109`), `timeout` number
    `Input` (`min="1"`; `111-119`), `interval` number `Input` (`min="200"`;
    `120-132`).
  - `137-185` — `Auth` (`@grafana/plugin-ui`) with a single visible method
    `custom-sumo`; `onAuthMethodSelect` is a **no-op** (`139`), `selectedMethod`
    is fixed (`141`). Custom method label `"Authentication method"` /
    description `"Provide information to grant access to the data source."`
    (`145-146`). Renders `accessId` `Input` (label `"AccessID"`, placeholder
    `"Sumo Logic Access Id"`; `149-162`) and `accessKey` `SecretInput` (label
    `"AccessKey"`, placeholder `"Access key"`; `163-180`).
  - `44-67` — `onAccessKeyChange` / `onResetAccessKey`: `secureJsonData.accessKey`
    writes.
- `src/types.ts:6` — `AuthenticationMethod = 'accessKey'`.
- `src/types.ts:8-14` — frontend `Config` (jsonData): `authMethod`, `apiUrl`,
  `accessId`, `timeout`, `interval`.
- `src/types.ts:20-22` — `SecureConfig` (`accessKey`).
- `src/constants.ts:3-41` — `ApiEndpointOptions` (nine regional API URLs).
- `src/constants.ts:53,55` — `DefaultTimeout = 30`, `DefaultInterval = 1000`.

### Backend sources

- `pkg/models/settings.go`:
  - `11-15` — `AuthenticationMethod` type + `AuthenticationMethodAccessKey`.
  - `17-19` — `DefaultApiURL`, `DefaultTimeout`, `DefaultInterval`.
  - `21-28` — `Settings` struct (loaded shape + json tags). `AccessKey` is
    `json:"-"`, populated from decrypted secrets.
  - `30-54` — `LoadSettings`: nil jsonData → `{}`, unmarshal, copy
    `DecryptedSecureJSONData["accessKey"]`, default `authMethod` → accessKey,
    `apiUrl` → `DefaultApiURL`, `timeout` → 30, `interval` → 1000.
  - `56-72` — `Validate`: `apiUrl`, `authMethod` required; for accessKey,
    `accessId` + `accessKey` required.
- `pkg/sumo/client.go`:
  - `24-26` — `Client.Validate` → `Settings.Validate` (run by the health check).
  - `45-58` — `httpclient.Options`: `timeout(settings)` for the HTTP timeout;
    for accessKey auth, `BasicAuthOptions{User: AccessID, Password: AccessKey}`;
    any other method returns `ErrorIncorrectAuthMethod`.
  - `145` — request URL join: `strings.TrimSuffix(fr.ApiUrl, "/") + "/" + endpoint`.
  - `158-163` — `timeout()` converts `settings.Timeout` seconds to a duration.
- `pkg/sumo/metrics_query.go:39,137`, `pkg/sumo/metrics_results_query.go:11`,
  `pkg/sumo/logs_query.go:103,129,168` — `settings.ApiURL` used as the request
  base URL.
- `pkg/sumo/logs_query.go:80` — `settings.Interval` used as the log-polling
  interval (`time.Duration(Interval) * time.Millisecond`).
- `pkg/errors/errors.go:5-8` — `ErrorIncorrectAuthMethod`
  (`"incorrect authentication method selected"`).
- `pkg/plugin/plugin.go:47-62`, `pkg/plugin/handlers_checkhealth.go:13-38` —
  `GetInstance` → `LoadSettings` → `NewClient`; `CheckHealth` → `Client.Validate`.
- `pkg/models/settings_test.go:13-93` — upstream `LoadSettings` / `Validate`
  tests (default seeding, interval defaulting, validation errors).

### External components (versions from the workspace catalog)

Resolved via the `catalog:` protocol in `package.json:73-84` against
`.yarnrc.yml:14-26`:

- **`@grafana/plugin-ui@^0.13.1`** —
  - `Auth` (`dist/esm/components/ConfigEditor/Auth/Auth.js`): wraps the
    authentication UI in a `ConfigSection` titled **"Authentication"**.
  - `AuthMethodSettings`
    (`dist/esm/components/ConfigEditor/Auth/auth-method/AuthMethodSettings.js`):
    with a single visible method (`hasSelect === false`) it renders **no**
    method `Select`; the section title/description come from the custom method's
    `label` / `description`, and the custom method's `component` is rendered.
    This is why the editor never persists an `authMethod` value.
  - `ConfigSection`, `DataSourceDescription`.
- **`@grafana/ui@^11.6.7`** — `Input`, `Select`, `SecretInput`, `InlineField`,
  `useTheme2`.
- **`@grafana/data@^11.6.7`** — `DataSourceJsonData` (base of the frontend
  `Config`), `DataSourcePluginOptionsEditorProps`, `SelectableValue`.
- **`@grafana/schema@^11.6.7`** — `DataQuery` (query type only; not config).

Backend: `github.com/grafana/grafana-plugin-sdk-go v0.279.0` (`go.mod:9`) —
`backend.DataSourceInstanceSettings` (`JSONData`, `DecryptedSecureJSONData`),
`backend/httpclient` (`Options`, `BasicAuthOptions`, `TimeoutOptions`).

## Field inventory

| schema id | storage key | target | editor label | read by backend |
| --- | --- | --- | --- | --- |
| `jsonData_apiUrl` | `apiUrl` | jsonData | API region / URL | yes — `settings.ApiURL` (request base URL, `client.go:145`) |
| `jsonData_timeout` | `timeout` | jsonData | Timeout | yes — HTTP client timeout (`client.go:45-48,158-163`) |
| `jsonData_interval` | `interval` | jsonData | Interval | yes — log-poll interval (`logs_query.go:80`) |
| `jsonData_authMethod` | `authMethod` | jsonData | — (no control writes it) | yes — `settings.AuthenticationMethod` (`settings.go:40-42`, `client.go:51`) |
| `jsonData_accessId` | `accessId` | jsonData | AccessID | yes — basic-auth username (`client.go:53`) |
| `secureJsonData_accessKey` | `accessKey` | secureJsonData | AccessKey | yes — basic-auth password (`client.go:54`) |

- **Root fields**: none. `RootConfig` is a blank object — the backend builds its
  own HTTP basic auth from `jsonData.accessId` + `secureJsonData.accessKey`
  (`client.go:51-55`) and never reads the datasource's root `url`/`basicAuth`
  fields.
- **Virtual fields**: none. There is only one authentication method and no
  editor-local derived selector, so no `kind: "virtual"` field is needed.

### Frontend-only settings

None. Every modeled `jsonData` field is read by the backend.

### Backend-only settings

- **`authMethod`** is read and defaulted by the backend (`settings.go:40-42`,
  `client.go:51`) but the config editor never writes it: the `Auth` widget is
  configured with a single fixed method and a no-op `onAuthMethodSelect`
  (`ConfigEditor.tsx:138-139`). It is declared in the frontend `Config` type
  (`src/types.ts:9`) but no control mutates it. Tagged `backend-only`; modeled
  with role `auth.discriminator`, `defaultValue: "accessKey"`, and an
  `allowedValues` of `["accessKey"]`. It carries no editor `label`/`placeholder`
  because no control renders it (its `description` explains the field, mirroring
  how the gold-standard GitHub entry documents its backend-only `cachingEnabled`).

## Modeling decisions

- **`apiUrl` modeled as a free-text `input`, not a `select`.** The editor renders
  `apiUrl` as a region `Select` with `allowCustomValue={true}` and
  `isClearable={true}` (`ConfigEditor.tsx:100-104`), options from
  `ApiEndpointOptions` (`src/constants.ts:3-41`). Encoding those options as
  `ui.options` on a `select` field would make the SDK converter emit a hard
  OpenAPI `enum` (`dsconfig/convert.go:applyUIEnum`, which does not honor
  `allowCustom`) that would wrongly reject custom/on-prem API URLs in the served
  settings spec. Because custom values are explicitly allowed, `apiUrl` is
  modeled as a free-text `input` (with `defaultValue` and the URL tooltip as its
  `description`); the nine regions are preserved in the connection `instruction`,
  the `accessKeyEU` example, and below. This mirrors the sibling
  `grafana-datadog-datasource` entry's handling of its region `Select`.
- **`requiredWhen` vs the editor.** The editor renders `DataSourceDescription`
  with `hasRequiredFields` and marks only the API URL `required`
  (`ConfigEditor.tsx:74,97`); `accessId` / `accessKey` carry **no** required
  marker, yet the backend `Validate` rejects them when empty under accessKey auth
  (`settings.go:64-70`). `requiredWhen: "jsonData_authMethod == 'accessKey'"`
  encodes that backend contract; `apiUrl` uses `requiredWhen: "true"` (always
  required per `settings.go:58-60`).
- **`range` validations are editor-intent, not backend-enforced.** `timeout`
  (`min: 1`) and `interval` (`min: 200`) mirror the number inputs'
  `min` attributes (`ConfigEditor.tsx:115,128`) and the interval tooltip
  ("Min value is 200"). The backend only substitutes defaults for **zero**
  values (`settings.go:46-51`) — it does not clamp small positive values — so
  the ranges capture the editor's intent (see *Upstream discrepancies*).
- **Groups mirror the editor sections.** `apiRegion` (title "API Region",
  `ConfigEditor.tsx:79`) comes first with `apiUrl` / `timeout` / `interval`, then
  `authentication` (title "Authentication", the section the `Auth` widget renders)
  with `accessId` / `accessKey`. `authMethod` (no control) is in no group, like
  the GitHub entry's backend-only `cachingEnabled`.
- **`accessId` / `accessKey` are a basic-auth pair.** Modeled with roles
  `auth.basic.username` / `auth.basic.password` and a `pair` relationship, because
  `client.go:52-55` wires them into `httpclient.BasicAuthOptions{User, Password}`.
- **Flat `Config` in Go.** `settings.go` mirrors the upstream `Settings`
  (`pkg/models/settings.go:21-28`) json tags for the jsonData fields and holds the
  decrypted access key in `DecryptedSecureJSONData` (rather than the upstream's
  `AccessKey string json:"-"` field). No root fields are carried. `LoadConfig`
  runs parse → `ApplyDefaults` → `Validate`; `ApplyDefaults` and `Validate` are
  exported for callers assembling a `Config` directly.
- **`SecureJsonDataConfig` is a key list.** Secure values are write-only, so the
  secure type is just the array of secret key names (`accessKey`); consumers read
  `secureJsonFields` to see what is configured.

### Exclusions

- **Secure Socks Proxy** — the Sumo Logic editor does **not** render the
  `SecureSocksProxySettings` control and the backend does not read
  `jsonData.enableSecureSocksProxy`, so there is nothing to exclude here (the
  field simply does not exist in this plugin). Noted for completeness per
  AGENTS.md.

## Where the types are defined

Only config type/field definitions are listed — UI components (`ConfigEditor`,
`Auth`, `DataSourceDescription`, `SecretInput`, …) and functions/helpers
(`LoadSettings`, `NewClient`, `timeout`, …) are omitted even where they are the
reason a field exists.

**Frontend (plugin):**
- `AuthenticationMethod` (`'accessKey'`) — `src/types.ts:6`.
- `Config` (jsonData: `authMethod`, `apiUrl`, `accessId`, `timeout`, `interval`)
  — `src/types.ts:8-14`.
- `SecureConfig` (`accessKey`) — `src/types.ts:20-22`.

**Backend (plugin):**
- `AuthenticationMethod` + `AuthenticationMethodAccessKey` —
  `pkg/models/settings.go:11-15`.
- `Settings` (loaded shape) — `pkg/models/settings.go:21-28`.
- `DefaultApiURL`, `DefaultTimeout`, `DefaultInterval` —
  `pkg/models/settings.go:17-19`.

**Library / SDK config types:**
- `DataSourceJsonData` — `@grafana/data@^11.6.7` (base interface the frontend
  `Config` intersects with; source of unused root jsonData fields).
- `backend.DataSourceInstanceSettings` (`JSONData`, `DecryptedSecureJSONData`),
  `backend/httpclient.{Options,BasicAuthOptions,TimeoutOptions}` —
  `github.com/grafana/grafana-plugin-sdk-go@v0.279.0` (the shape `LoadConfig` /
  `LoadSettings` read and the HTTP-client options the credentials feed).

## Settings examples matrix

| Example key | Auth | Region (`apiUrl`) | `secureJsonData` |
| --- | --- | --- | --- |
| `""` (default) | accessKey (schema defaults) | US1 / Default (`https://api.sumologic.com/api/`) | `accessKey` (empty) |
| `accessKey` | accessKey | US1 / Default | `accessKey` |
| `accessKeyEU` | accessKey | EU (`https://api.eu.sumologic.com/api/`) | `accessKey` |
| `legacyNoAuthMethod` | accessKey (defaulted; no `authMethod`) | US1 / Default | `accessKey` |

The `""` default example has an empty `accessKey` placeholder and no `accessId`,
so it intentionally fails `LoadConfig`'s `Validate` step (covered by a test). All
secret values use obviously-fake `<…>` angle-bracket placeholders
(`<your-access-id>`, `<your-access-key>`).

## Potential upstream bugs / discrepancies

All preserved verbatim in the schema — the schema records what the plugin
**does**, not what it **should** do.

1. **The editor never persists `authMethod`.** The `Auth` widget uses a single
   custom method with a no-op `onAuthMethodSelect` (`ConfigEditor.tsx:138-139`),
   so a datasource created through the UI stores no `authMethod`; the backend
   fills in `accessKey` on load (`settings.go:40-42`). The one real value
   (`accessKey`) is also unrelated to the widget's method id (`custom-sumo`).
2. **The auth discriminator is effectively vestigial.** `accessKey` is the only
   supported method; `client.go:56-57` returns `ErrorIncorrectAuthMethod` for
   anything else. There is no alternative method to select, so `authMethod`
   exists only as a required, single-valued discriminator.
3. **`accessId` / `accessKey` requiredness is backend-only.** The editor marks
   neither field `required` (`ConfigEditor.tsx:149-180`) even though
   `DataSourceDescription` sets `hasRequiredFields` (`:74`), but the backend
   `Validate` rejects empty values (`settings.go:64-70`). `requiredWhen` encodes
   the backend contract.
4. **Editor `min` bounds are not backend-enforced.** `timeout` (`min="1"`) and
   `interval` (`min="200"`, plus the "Min value is 200" tooltip) are only
   validated in the browser. `LoadSettings` substitutes defaults only for **zero**
   values (`settings.go:46-51`); a provisioned `interval: 50` is accepted and used
   directly (`logs_query.go:80`). The schema keeps the `range` validations to
   record the editor's intent.
5. **`isClearable` resets rather than clears.** The `apiUrl` `Select` is
   `isClearable`, but its `onChange` coalesces an empty selection back to the
   default URL (`ConfigEditor.tsx:106-108`), so clearing it stores the default
   rather than an empty value.
6. **One-word "SumoLogic" vs "Sumo Logic".** The editor's
   `DataSourceDescription` uses `dataSourceName="SumoLogic"` (`:72`) and the API
   URL tooltip says "SumoLogic API URL" (`:86`), while the plugin `name` is
   "Sumo Logic" (`plugin.json:4`). Cosmetic.
7. **Inconsistent docs URLs.** The editor's `docsLink` is
   `https://grafana.com/grafana/plugins/grafana-sumologic-datasource/`
   (`ConfigEditor.tsx:73`) while `plugin.json`'s docs link is
   `https://grafana.com/docs/plugins/grafana-sumologic-datasource`
   (`plugin.json:25`). The schema uses the `plugin.json` value for `docURL`.
8. **Redundant fallback constant.** `pkg/sumo/logs_query.go:18` redefines a local
   `ApiURL` constant equal to `models.DefaultApiURL` and uses it as a fallback in
   `utils.Value(sc.Settings.ApiURL, ApiURL)` (`:103,129,168`), even though
   `LoadSettings` already defaults `settings.ApiURL`. Harmless.

## Validation performed

- `dsconfig.ParseAndResolveSchemaJSON` + `Schema.Validate()` (via the conformance
  suite's schema round-trip) — passes.
- JSON Schema validation of `dsconfig.json` against `dsconfig/schema.json`
  (draft 2020-12, strict `additionalProperties: false`) — passes.
- `go generate ./...` in this directory (regenerates the three `*.gen.json`
  artifacts).
- From `registry/`: `gofmt -l .` (clean), `go build ./...`, `go vet ./...`,
  `go test ./...` (all entries pass, including this entry's conformance +
  `LoadConfig` / `ApplyDefaults` / `Validate` / examples tests).
- The pre-existing `dsconfig` and `schema` workspace modules still build and test.
- `tsc --noEmit --strict` on `settings.ts` (TypeScript 5) — passes.

## What `LoadConfig` guarantees

`LoadConfig(ctx, settings)` runs three phases and returns a fully-defaulted,
validated `Config`:

1. **Parse** — treat nil/empty `settings.JSONData` as `{}`, unmarshal it into
   `Config`, and copy the decrypted `accessKey` into `DecryptedSecureJSONData`
   (mirrors `LoadSettings`, `settings.go:30-39`).
2. **ApplyDefaults** — `authMethod` → `accessKey`, `apiUrl` → `DefaultApiURL`,
   `timeout` → 30, `interval` → 1000 (`settings.go:40-51`).
3. **Validate** — `apiUrl` + `authMethod` required; for accessKey auth,
   `accessId` + `accessKey` required (`settings.go:56-72`). Errors are joined.

Everything is logged via `backend.Logger.FromContext(ctx)` with `datasource_uid`,
`datasource_name`, and `plugin` labels.
