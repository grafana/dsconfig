# Azure Monitor Managed Service for Prometheus — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered
from this entry's [`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `grafana-azureprometheus-datasource`
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/dfrbqiwn8zj7kc` (Grafana Enterprise, `@grafana/prometheus` PromSettings + `@grafana/azure-sdk` `AzureCredentialsForm` inside `DataSourceHttpSettingsOverhaul`)
- **New UI:** `http://192.168.1.241:58899/?path=/story/configeditor-datasourceconfigwizard--tab&args=pluginType:grafana-azureprometheus-datasource` (Storybook, `ConfigEditor/DatasourceConfigWizard`), also the `--wizard` story.
- **Method:** Playwright captured both UIs (full-page screenshots + DOM extraction + `Save & Test`-button console payloads). The Storybook story fetches the schema from `raw.githubusercontent.com/.../registry/grafana-azureprometheus-datasource/dsconfig.json`; the local (edited) `dsconfig.json` was served to the new UI by intercepting that request with `context.route(...)`, so the captures reflect the local schema without pushing.
- **Result:** **Parity achieved** for the in-scope surface. One missing field was found (**Custom HTTP Headers**) and added; `fileUpload` was evaluated and correctly **not** used (0 file inputs in the legacy DOM); `root_url` was made unconditionally required so it lands in the wizard's synthetic **General** step. This entry mirrors `prometheus` for the shared Prometheus knobs; the auth block differs (Azure discriminated-union credentials, not the vanilla `virtual_authMethod` selector).

---

## TL;DR of changes

| #   | Change                                                                                           | File                             | Why                                                                                                            |
| --- | ------------------------------------------------------------------------------------------------ | -------------------------------- | -------------------------------------------------------------------------------------------------------------- |
| 1   | Changed `root_url` from `requiredWhen: "true"` → `required: true`                                | [`dsconfig.json`](dsconfig.json) | URL is unconditionally required; puts URL into the wizard's synthetic **General** step and emits `required: ["url"]` |
| 2   | Added **Custom HTTP Headers** field (`jsonData_httpHeaders`) with `indexedPair` storage           | [`dsconfig.json`](dsconfig.json) | Legacy UI renders an **HTTP headers** section with **Add header**; new UI had no headers editor                |
| 3   | Added `jsonData_httpHeaders` to the `advanced-http` group's `fieldRefs` (after `jsonData_timeout`) | [`dsconfig.json`](dsconfig.json) | Surface the new field under **Advanced HTTP settings**, matching the legacy grouping                          |
| 4   | Added a truthful `llm`-tagged instruction that headers **are** modeled                            | [`dsconfig.json`](dsconfig.json) | Keep the embedded instructions complete/parallel to `prometheus` after change #2 (see note below)             |
| 5   | Regenerated `schema.gen.json`, `settings.gen.json` via `go generate ./registry/grafana-azureprometheus-datasource/...` | generated artifacts | Keep committed artifacts in sync (guarded by `SchemaArtifactInSync`)                                          |

No changes were made to `settings.go`, `settings.ts`, `README.md`, `settings.examples.gen.json`, `conformance_test.go`, `schema/conformance.go`, or `plugin-ui`. All schema changes flow through `dsconfig.json`; the rest is produced by `go generate`.

> **Instruction note:** unlike `prometheus`, this entry's `instructions` array never contained a claim that headers were "not modeled", so there was nothing to *rewrite* there — a new truthful instruction was **added** instead. The stale "not modeled" language lives only in `README.md` (lines 156–160), which is **off-limits** per this task's constraints and was therefore left unchanged (see [Not fixable via dsconfig.json](#not-fixable-via-dsconfigjson)).

**No conformance-test or plugin-ui change was required** (see [Conformance](#conformance-no-change-required) and [Wizard mode](#wizard-mode-url-in-the-general-step)) — the `indexedPair` conformance skip and the auth-group→General generalisation were already in place from earlier work.

---

## Section layout

The `groups` were left in their existing taxonomy; only the new field was slotted into
**Advanced HTTP settings**. Verified rendering top-to-bottom in the new UI (tab mode, `newgen-azureprom-fixed`):

| Order | Section (`id`) | `optional` | Fields (in display order) |
| --- | --- | --- | --- |
| 1 | **Connection** (`connection`) | no | Prometheus server URL |
| 2 | **Authentication** (`authentication`) | no | Authentication (Azure discriminated-union) → Client Secret, legacy clientSecret, oauthPassThru, azureEndpointResourceId |
| 3 | **Alerting** (`alerting`) | yes | Manage alerts via Alerting UI, Allow as recording rules target |
| 4 | **Advanced HTTP settings** (`advanced-http`) | yes | Timeout, **Custom HTTP Headers** ➕, Allowed cookies |
| 5 | **Interval behaviour** (`interval-behaviour`) | yes | Scrape interval, Query timeout |
| 6 | **Query editor** (`query-editor`) | yes | Default editor, Disable metrics lookup |
| 7 | **Performance** (`performance`) | yes | Prometheus type, Version, Cache level, Incremental querying → Query overlap window, Disable recording rules |
| 8 | **Other** (`other`) | yes | Custom query parameters, HTTP method, Series limit, Use series endpoint |
| 9 | **Exemplars** (`exemplars`) | yes | Exemplar trace-ID destinations (array of objects) |
| 10 | **Migration** (`migration`) | yes | prometheus-type-migration sentinel (provisioning-only) |

Notes:

- The **Custom HTTP Headers** field is placed in **Advanced HTTP settings**, immediately after
  **Timeout** and before **Allowed cookies** — verified in the tab capture, which rendered
  `["Timeout", "Custom HTTP Headers", "Add custom http header", "Allowed cookies"]` in that order.
- Unlike the `prometheus` entry, there is **no `tls-settings` group** and **no `virtual_authMethod`
  selector**. Azure auth is locked (`visibleMethods=[azureAuthId]`), so basic-auth / OAuth-forward /
  TLS are intentionally **not modeled** here (documented in `README.md` → *Excluded settings* and
  `AGENTS.md`). This is a pre-existing decision and is **out of scope** for this parity pass — the
  legacy DOM still shows an (empty-for-Azure-auth) "TLS settings" heading rendered by the plugin-ui
  scaffold, but no TLS controls are exposed. Left unchanged.
- Two `jsonData` fields (`maxSamplesProcessedWarningThreshold`, `maxSamplesProcessedErrorThreshold`)
  are `backend-only` (tagged as such) and belong to no group — the PromSettings editor never renders
  them (feature-flagged off), so they are intentionally not surfaced in either UI. Parity preserved.

### Wizard mode: URL in the "General" step

In **wizard mode** the plugin-ui builds a synthetic first step titled **General**
(`resolveRequiredFieldsGroup`) from (a) every field marked `required: true`, plus (b) the
fields of the auth group, plus their `dependsOn` parents/children.

- **Before:** `root_url` used `requiredWhen: "true"` (a CEL expression the resolver does **not**
  inspect), so URL was **not** guaranteed into General.
- **After:** changing it to `required: true` (unconditionally required — the backend rejects an empty
  URL) puts URL into General and emits a proper OpenAPI `required: ["url"]` in the generated spec
  (instead of the `x-dsconfig-required-when: "true"` extension).

**Verified (`newgen-azureprom-fixed-wiz`):** the wizard opens on **General 1/11** with
`urlPresent: true` (the required `*` asterisk is present) and the Azure **Authentication** block
folded in from the auth group (Client Secret / clientSecret / oauthPassThru / azureEndpointResourceId).
The auth group folds into General because it uses the conventional `id: "authentication"`, which the
plugin-ui wizard already recognises. **No plugin-ui change was needed.**

---

## Field-by-field parity

Legend: ✅ present & matching · ➕ added by this change · 🔀 conditional (revealed by `dependsOn`) · ⚙ Azure discriminated-union auth (SDK-managed) · 🔒 backend/provisioning-only (no editor UI in either)

| Legacy UI field | Control (legacy) | New UI (schema id) | Control (new) | Storage target | Status |
| --- | --- | --- | --- | --- | --- |
| Prometheus server URL | text input | `root_url` | input | `root.url` (required) | ✅ |
| Authentication | Azure `AzureCredentialsForm` (App Registration / Managed Identity / Workload Identity / Current User) | `jsonData_azureCredentials` | opaque (`valueType: any`, `role: auth.discriminator`) | `jsonData.azureCredentials` | ✅ ⚙ |
| Client Secret | password (secure) | `secureJsonData_azureClientSecret` | secure input | `secureJsonData.azureClientSecret` | ✅ 🔀 ⚙ |
| — (legacy pre-migration key) | — | `secureJsonData_clientSecret` | (no UI) | `secureJsonData.clientSecret` | ✅ 🔒 |
| — (managed on save) | — | `jsonData_oauthPassThru` | (no UI) | `jsonData.oauthPassThru` | ✅ 🔒 |
| — (provisioning-only) | — | `jsonData_azureEndpointResourceId` | (no UI) | `jsonData.azureEndpointResourceId` | ✅ 🔒 |
| Timeout | number | `jsonData_timeout` | number | `jsonData.timeout` | ✅ |
| **Custom HTTP Headers** | HTTP headers → Add header → name input + value password | `jsonData_httpHeaders` | IndexedPair editor (key + secret value) | `jsonData.httpHeaderName<N>` / `secureJsonData.httpHeaderValue<N>` | ➕ |
| Allowed cookies | TagsInput | `jsonData_keepCookies` | list (string array) | `jsonData.keepCookies` | ✅ |
| Manage alerts via Alerting UI | switch | `jsonData_manageAlerts` | switch | `jsonData.manageAlerts` | ✅ |
| Allow as recording rules target | switch | `jsonData_allowAsRecordingRulesTarget` | switch | `jsonData.allowAsRecordingRulesTarget` | ✅ |
| Scrape interval | text input | `jsonData_timeInterval` | input | `jsonData.timeInterval` | ✅ |
| Query timeout | text input | `jsonData_queryTimeout` | input | `jsonData.queryTimeout` | ✅ |
| Default editor | select (Builder/Code) | `jsonData_defaultEditor` | select | `jsonData.defaultEditor` | ✅ |
| Disable metrics lookup | switch | `jsonData_disableMetricsLookup` | switch | `jsonData.disableMetricsLookup` | ✅ |
| Prometheus type | select (Prometheus/Cortex/Mimir/Thanos) | `jsonData_prometheusType` | select | `jsonData.prometheusType` | ✅ |
| Version | select (allowCustom) | `jsonData_prometheusVersion` | select | `jsonData.prometheusVersion` | ✅ 🔀 |
| Cache level | select (Low/Medium/High/None) | `jsonData_cacheLevel` | select | `jsonData.cacheLevel` | ✅ |
| Incremental querying (beta) | switch | `jsonData_incrementalQuerying` | switch | `jsonData.incrementalQuerying` | ✅ |
| Query overlap window | text input | `jsonData_incrementalQueryOverlapWindow` | input | `jsonData.incrementalQueryOverlapWindow` | ✅ 🔀 |
| Disable recording rules (beta) | switch | `jsonData_disableRecordingRules` | switch | `jsonData.disableRecordingRules` | ✅ |
| Custom query parameters | text input | `jsonData_customQueryParameters` | input | `jsonData.customQueryParameters` | ✅ |
| HTTP method | select (POST/GET) | `jsonData_httpMethod` | select | `jsonData.httpMethod` | ✅ |
| Series limit | number | `jsonData_seriesLimit` | input | `jsonData.seriesLimit` | ✅ |
| Use series endpoint | switch | `jsonData_seriesEndpoint` | switch | `jsonData.seriesEndpoint` | ✅ |
| Exemplars | repeated destination rows | `jsonData_exemplarTraceIdDestinations` | array-of-objects editor | `jsonData.exemplarTraceIdDestinations` | ✅ |
| Data source migrated banner | warning banner (read-only) | `jsonData_prometheusTypeMigration` | (no input; provisioning-only) | `jsonData['prometheus-type-migration']` | ✅ 🔒 |
| — (not in editor) | — | `jsonData_maxSamplesProcessedWarningThreshold` | — | `jsonData.maxSamplesProcessedWarningThreshold` | ✅ 🔒 |
| — (not in editor) | — | `jsonData_maxSamplesProcessedErrorThreshold` | — | `jsonData.maxSamplesProcessedErrorThreshold` | ✅ 🔒 |

The three TLS cert fields modeled by `prometheus` are **absent by design** here (Azure-auth lock).
Any `target: "secureJsonData"` field (e.g. Client Secret) is drawn by the new renderer as a masked
secure input with a show/hide toggle regardless of its declared `ui.component` — the same renderer
policy documented for `prometheus`; only the widget affordance differs, storage keys are identical.

---

## Gap found and fixed: Custom HTTP Headers

**Legacy behaviour (verified in `legacy-expand-azureprometheus-verify.json`, UID `dfrbqiwn8zj7kc`):**
the editor includes an **HTTP headers** section heading with an **Add header** button
(`hasHeaders: true`, `addHeaderBtn: true`). `@grafana/plugin-ui`'s CustomHeaders component persists
headers as indexed pairs — `jsonData.httpHeaderName<N>` for the name and
`secureJsonData.httpHeaderValue<N>` for the (secret) value, starting at `N = 1`.

**Before:** `dsconfig.json` did not model headers at all, so the new UI rendered no headers editor.

**Fix (in `dsconfig.json` only):** added the `jsonData_httpHeaders` array field — copied verbatim
from the `prometheus` entry — with an `indexedPair` storage mapping that reproduces the exact legacy
storage, plus item sub-fields for the header name (`http.header.name`, with a header-name pattern
validation) and value (`http.header.value`), and added it to the **Advanced HTTP settings** group's
`fieldRefs` after `jsonData_timeout`.

**After (verified in `newgen-azureprom-fixed`):** the new UI renders a **Custom HTTP Headers** row
under **Advanced HTTP settings** with an **Add custom http header** button and a key/secret-value
editor (`hasHeadersEditor: true`).

### Header save-payload storage-target validation

`capture-headers-generic.js` filled a header (name `X-Org-Id`, value `influx-tenant-9`) and clicked
**Save & Test**; the logged datasource payload was:

```
jsonData.httpHeaderName1: X-Org-Id | secureJsonData.httpHeaderValue1: influx-tenant-9 | url: http://influx.example.com:8086
```

The header **name lands in `jsonData.httpHeaderName1`** and the **value in
`secureJsonData.httpHeaderValue1`** — byte-for-byte the legacy CustomHeaders storage format. URL
routes to `root.url`.

---

## `fileUpload` evaluation — not applicable

The task asked to use the `fileUpload` control **if the legacy UI uses it**. It does not:

- The legacy Azure Prometheus editor exposes no file inputs and no upload buttons
  (`legacy-expand-azureprometheus-verify.json`: `fileInputs: 0`, `uploadButtons: []`).
- Azure credentials are collected as text/UUID inputs + a client-secret password by
  `@grafana/azure-sdk`'s form; there is no PEM/cert or service-account-file upload.

**Decision:** do **not** add `fileUpload` to any field. No packs apply.

---

## Conformance (no change required)

Adding an `indexedPair` field would normally break `JSONDataMatchesStruct` /
`JSONDataTypesMatchStruct`, because an `indexedPair` field (`jsonData_httpHeaders`) is a **logical
view** over dynamically-indexed legacy keys (`httpHeaderName1`, `httpHeaderValue1`, …) that are not
modeled as a single Go struct field. The shared conformance walker (`schema/conformance.go`) already
skips `indexedPair` fields via `isIndexedPairField()` — plugin-agnostic, added during earlier work —
so **no conformance change was needed here**. The per-header `httpHeaderValue<N>` secrets remain
dynamic and are correctly **not** listed among the static `SecureJsonDataKeys` (`settings.go`), and
the generated spec emits `httpHeaders` as a clean `{name, value}` array under `jsonData` with **no**
secure values leaked (`SchemaSpecHasNoSecureJSON` passes).

---

## Verification

```
go generate ./registry/grafana-azureprometheus-datasource/...     # regenerate schema.gen.json / settings.gen.json
go test     ./registry/grafana-azureprometheus-datasource/...     # PASS
go test     ./registry/grafana-azureprometheus-datasource/... ./schema/...   # PASS (shared walker, no regressions)
```

`TestSchemaConformance` subtests — all **PASS**:
`BaseFieldsResolved`, `SchemaRoundTrip`, `SchemaArtifactInSync`, `SchemaSpecHasNoSecureJSON`,
`ConfigSchemaValid`, `JSONDataMatchesStruct`, `JSONDataTypesMatchStruct`,
`SecureValuesMatchLoadSettings`.

The hand-authored `settings_test.go` suites also pass unchanged (they do not reference the new field).

New-UI captures: `newgen-azureprom-fixed` (tab, `hasHeadersEditor: true`, `urlPresent: true`),
`newgen-azureprom-fixed-wiz` (wizard opens on **General 1/11** with required URL + Azure auth block),
plus the `capture-headers-generic` header save-payload routing (above).
Legacy capture: `legacy-expand-azureprometheus-verify` (HTTP headers + Add header present; 0 file inputs).

---

## Not fixable via dsconfig.json

- **`README.md` is stale** re: headers. Its *Excluded settings* section (lines 156–160) still says
  Custom HTTP headers are "not rendered by this plugin's editor … not modeled". After this change
  they **are** modeled (and the legacy editor **does** render them). `README.md` edits are forbidden
  by this task's constraints, so it was left unchanged — this should be reconciled separately.
- No conformance / plugin-ui change was required, so none was made.

---

## Files changed

- [`registry/grafana-azureprometheus-datasource/dsconfig.json`](dsconfig.json) — changed `root_url`
  from `requiredWhen: "true"` to `required: true`; added the `jsonData_httpHeaders` `indexedPair`
  field (verbatim from `prometheus`) after `jsonData_timeout` and referenced it from the
  `advanced-http` group; added an `llm`-tagged instruction stating headers are now modeled.
- [`registry/grafana-azureprometheus-datasource/schema.gen.json`](schema.gen.json),
  [`registry/grafana-azureprometheus-datasource/settings.gen.json`](settings.gen.json) — regenerated
  by `go generate` (`url` now in the spec's `required` array; `httpHeaders` `{name, value}` array
  added under `jsonData`; `x-dsconfig-required-when: "true"` removed from `url`).

_Unchanged by design:_ `settings.go`, `settings.ts`, `README.md`, `settings.examples.gen.json`,
`conformance_test.go`, `schema/conformance.go`, and everything in `plugin-ui`.
