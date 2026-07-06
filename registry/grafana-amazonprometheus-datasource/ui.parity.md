# Amazon Managed Service for Prometheus — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered
from this entry's [`dsconfig.json`](dsconfig.json)).

This plugin is **Prometheus + AWS SigV4 auth**: the query/settings surface mirrors the core
Prometheus datasource, but authentication is locked to AWS SigV4 (via `@grafana/aws-sdk`'s
`SIGV4ConnectionConfig`) instead of the Basic/OAuth/No-auth model. The two parity fixes applied
here are **identical** to the ones done for `prometheus`.

- **Plugin id:** `grafana-amazonprometheus-datasource`
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/bfrbd04c66tc0a` (Grafana Enterprise, `@grafana/prometheus` PromSettings + `@grafana/aws-sdk` `SIGV4ConnectionConfig` inside `DataSourceHttpSettingsOverhaul`)
- **New UI:** `http://localhost:64433/?path=/story/configeditor-datasourceconfigwizard--tab&args=pluginType:grafana-amazonprometheus-datasource` (Storybook, `ConfigEditor/DatasourceConfigWizard`), also the `--wizard` story.
- **Method:** Playwright captured both UIs (full-page screenshots + DOM extraction + `Save & Test`-button console payloads). The Storybook story fetches the schema from `raw.githubusercontent.com/.../registry/grafana-amazonprometheus-datasource/dsconfig.json`; the local (edited) `dsconfig.json` was served to the new UI by intercepting that request with `context.route(...)`, so the captures reflect the local schema without pushing.
- **Result:** **Parity achieved.** One missing field was found (**Custom HTTP Headers**) and added; `fileUpload` was evaluated and correctly **not** used (0 file inputs in the legacy DOM); the header save routing was verified from the `Save & Test` payload (`name → jsonData.httpHeaderName1`, `value → secureJsonData.httpHeaderValue1`).

---

## TL;DR of changes

| #   | Change                                                                                                                  | File                             | Why                                                                                                                  |
| --- | ----------------------------------------------------------------------------------------------------------------------- | -------------------------------- | -------------------------------------------------------------------------------------------------------------------- |
| 1   | Changed `root_url` from `requiredWhen: "true"` → `required: true`                                                       | [`dsconfig.json`](dsconfig.json) | URL is unconditionally required; puts URL into the wizard's synthetic **General** step and emits `required: ["url"]` |
| 2   | Added **Custom HTTP Headers** field (`jsonData_httpHeaders`) with `indexedPair` storage                                 | [`dsconfig.json`](dsconfig.json) | Legacy UI renders an **HTTP headers** section with **Add header**; new UI had no headers editor                      |
| 3   | Added `jsonData_httpHeaders` to the `advanced-http` group's `fieldRefs` (after `jsonData_timeout`)                      | [`dsconfig.json`](dsconfig.json) | Surface the new field under **Advanced HTTP settings**, matching the legacy grouping                                 |
| 4   | Appended a note to the `connection`/`legacy` instruction: headers are now **modeled**                                   | [`dsconfig.json`](dsconfig.json) | Keep the embedded `llm`-tagged instructions truthful after change #2                                                 |
| 5   | Regenerated `schema.gen.json`, `settings.gen.json` via `go generate ./registry/grafana-amazonprometheus-datasource/...` | generated artifacts              | Keep committed artifacts in sync (guarded by `SchemaArtifactInSync`)                                                 |

No changes were made to `settings.go`, `settings.ts`, `README.md`, `conformance.go`, `schema.go`, or `plugin-ui`. All schema changes flow through `dsconfig.json`; the rest is produced by `go generate`.

**No conformance-test or plugin-ui change was required** — the shared `indexedPair` skip in
`schema/conformance.go` and the wizard's auth-group fold-in were already in place from earlier work.

---

## Section layout

The `groups` were left in their existing taxonomy; only the new field was slotted into
**Advanced HTTP settings**. Verified rendering top-to-bottom in the new UI (tab mode) —
`sectionButtons` from `newgen-amazonprom-fixed.json`:

| Order | Section (`id`)                                | `optional` | Fields (in display order)                                                                                                                                                                                                                |
| ----- | --------------------------------------------- | ---------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1     | **Connection** (`connection`)                 | no         | Prometheus server URL                                                                                                                                                                                                                    |
| 2     | **SigV4 Auth Details** (`authentication`)     | no         | (managed `sigV4Auth`) → Authentication Provider → Credentials Profile Name → Access Key ID → Secret Access Key → Assume Role ARN → External ID → Default Region → Service → Forward Grafana User HTTP Header → (managed `oauthPassThru`) |
| 3     | **Alerting** (`alerting`)                     | yes        | Manage alerts via Alerting UI, Allow as recording rules target                                                                                                                                                                           |
| 4     | **Advanced HTTP settings** (`advanced-http`)  | yes        | Timeout, **Custom HTTP Headers** ➕, Allowed cookies                                                                                                                                                                                     |
| 5     | **Interval behaviour** (`interval-behaviour`) | yes        | Scrape interval, Query timeout                                                                                                                                                                                                           |
| 6     | **Query editor** (`query-editor`)             | yes        | Default editor, Disable metrics lookup                                                                                                                                                                                                   |
| 7     | **Performance** (`performance`)               | yes        | Prometheus type 🔒, Version 🔒, Cache level, Incremental querying → Query overlap window, Disable recording rules                                                                                                                        |
| 8     | **Other** (`other`)                           | yes        | Custom query parameters, HTTP method, Series limit, Query warning threshold, Query error threshold, Use series endpoint                                                                                                                  |
| 9     | **Exemplars** (`exemplars`)                   | yes        | Exemplar trace-ID destinations 🔒                                                                                                                                                                                                        |
| 10    | **Migration** (`migration`)                   | yes        | `prometheus-type-migration` sentinel 🔒                                                                                                                                                                                                  |

Notes:

- The **Custom HTTP Headers** field is placed in **Advanced HTTP settings** (alongside Timeout +
  Allowed cookies), matching where the legacy editor keeps the HTTP-transport knobs (legacy heading
  **HTTP headers**).
- `jsonData.prometheusType` and `jsonData.prometheusVersion` are `backend-only` (tagged): the plugin
  passes `hidePrometheusTypeVersion={true}` to `PromSettings` (`ConfigEditor.tsx:100`), so neither UI
  renders them. `jsonData.exemplarTraceIdDestinations` is likewise `backend-only` (`hideExemplars={true}`,
  `ConfigEditor.tsx:101`), and `jsonData['prometheus-type-migration']` is a frontend-only sentinel
  (never an input — it only drives the migration banner). Parity preserved in all four cases.
- `optional` groups render collapsed/collapsible in tab mode.

### Wizard mode: URL in the "General" step

In **wizard mode** the plugin-ui builds a synthetic first step titled **General**
(`resolveRequiredFieldsGroup`) from (a) every field marked `required: true`, plus (b) the fields of
the auth group, plus their `dependsOn` parents/children.

- **Before:** `root_url` used `requiredWhen: "true"` (a CEL expression the resolver does **not**
  inspect), so URL would not be folded into General; the generated spec carried an
  `x-dsconfig-required-when: "true"` extension instead of a proper `required` entry.
- **After:** changing it to `required: true` (unconditionally required — `promlib` admission and this
  entry's `Validate()`, `settings.go:313`, reject an empty URL) puts URL into General and emits a
  proper OpenAPI `required: ["url"]` in `schema.gen.json`.

**Verified (`newgen-amazonprom-fixed-wiz.json`):** the wizard opens on **General 1/11** with
`urlPresent: true` (the Prometheus-server-URL input with its required `*` marker) plus the SigV4 auth
group folded in — `Authentication Provider`, `sigV4Auth`, `Assume Role ARN`, `Default Region`,
`Service`, `Forward Grafana User HTTP Header`, `oauthPassThru`. The auth group folds into General
because it uses the conventional `id: "authentication"`, which the wizard already recognises. **No
plugin-ui change was needed.**

---

## Field-by-field parity

Legend: ✅ present & matching · ➕ added by this change · 🔀 conditional (revealed by `dependsOn`) · ⚙ SDK-managed (written implicitly, no direct input) · 🔒 backend-only / hidden (no editor UI in either)

| Legacy UI field                        | Control (legacy)                                                                         | New UI (schema id)                             | Control (new)                           | Storage target                                                     | Status |
| -------------------------------------- | ---------------------------------------------------------------------------------------- | ---------------------------------------------- | --------------------------------------- | ------------------------------------------------------------------ | ------ |
| Prometheus server URL                  | text input                                                                               | `root_url`                                     | input                                   | `root.url` (required)                                              | ✅     |
| — (forced true on mount)               | —                                                                                        | `jsonData_sigV4Auth`                           | (hidden, managed)                       | `jsonData.sigV4Auth`                                               | ✅ ⚙   |
| Authentication Provider                | select (Workspace IAM / Grafana Assume Role / AWS SDK Default / Keys / Credentials file) | `jsonData_sigV4AuthType`                       | select                                  | `jsonData.sigV4AuthType`                                           | ✅     |
| Credentials Profile Name               | text input                                                                               | `jsonData_sigV4Profile`                        | input                                   | `jsonData.sigV4Profile`                                            | ✅ 🔀  |
| Access Key ID                          | text/secure input                                                                        | `secureJsonData_sigV4AccessKey`                | secure input                            | `secureJsonData.sigV4AccessKey`                                    | ✅ 🔀  |
| Secret Access Key                      | password (secure)                                                                        | `secureJsonData_sigV4SecretKey`                | secure input                            | `secureJsonData.sigV4SecretKey`                                    | ✅ 🔀  |
| Assume Role ARN                        | text input                                                                               | `jsonData_sigV4AssumeRoleArn`                  | input                                   | `jsonData.sigV4AssumeRoleArn`                                      | ✅     |
| External ID                            | text input                                                                               | `jsonData_sigV4ExternalId`                     | input                                   | `jsonData.sigV4ExternalId`                                         | ✅ 🔀  |
| Default Region                         | select (allowCustom)                                                                     | `jsonData_sigV4Region`                         | select                                  | `jsonData.sigV4Region`                                             | ✅     |
| Service                                | text input                                                                               | `jsonData_sigv4Service`                        | input                                   | `jsonData.sigv4Service` (default `aps`)                            | ✅     |
| Forward Grafana User HTTP Header       | switch                                                                                   | `jsonData_forwardGrafanaUserHeader`            | switch                                  | `jsonData.forwardGrafanaUserHeader`                                | ✅     |
| — (cleared false on save)              | —                                                                                        | `jsonData_oauthPassThru`                       | (hidden, managed)                       | `jsonData.oauthPassThru`                                           | ✅ ⚙   |
| Manage alerts via Alerting UI          | switch                                                                                   | `jsonData_manageAlerts`                        | switch                                  | `jsonData.manageAlerts`                                            | ✅     |
| Allow as recording rules target        | switch                                                                                   | `jsonData_allowAsRecordingRulesTarget`         | switch                                  | `jsonData.allowAsRecordingRulesTarget`                             | ✅     |
| Timeout                                | number                                                                                   | `jsonData_timeout`                             | number                                  | `jsonData.timeout`                                                 | ✅     |
| **Custom HTTP Headers**                | Add header → name input + value password                                                 | `jsonData_httpHeaders`                         | IndexedPair editor (key + secret value) | `jsonData.httpHeaderName<N>` / `secureJsonData.httpHeaderValue<N>` | ➕     |
| Allowed cookies                        | TagsInput                                                                                | `jsonData_keepCookies`                         | list (string array)                     | `jsonData.keepCookies`                                             | ✅     |
| Scrape interval                        | text input                                                                               | `jsonData_timeInterval`                        | input                                   | `jsonData.timeInterval`                                            | ✅     |
| Query timeout                          | text input                                                                               | `jsonData_queryTimeout`                        | input                                   | `jsonData.queryTimeout`                                            | ✅     |
| Default editor                         | select (Builder/Code)                                                                    | `jsonData_defaultEditor`                       | select                                  | `jsonData.defaultEditor`                                           | ✅     |
| Disable metrics lookup                 | switch                                                                                   | `jsonData_disableMetricsLookup`                | switch                                  | `jsonData.disableMetricsLookup`                                    | ✅     |
| — (hidden `hidePrometheusTypeVersion`) | —                                                                                        | `jsonData_prometheusType`                      | —                                       | `jsonData.prometheusType`                                          | ✅ 🔒  |
| — (hidden `hidePrometheusTypeVersion`) | —                                                                                        | `jsonData_prometheusVersion`                   | —                                       | `jsonData.prometheusVersion`                                       | ✅ 🔒  |
| Cache level                            | select (Low/Medium/High/None)                                                            | `jsonData_cacheLevel`                          | select                                  | `jsonData.cacheLevel`                                              | ✅     |
| Incremental querying (beta)            | switch                                                                                   | `jsonData_incrementalQuerying`                 | switch                                  | `jsonData.incrementalQuerying`                                     | ✅     |
| Query overlap window                   | text input                                                                               | `jsonData_incrementalQueryOverlapWindow`       | input                                   | `jsonData.incrementalQueryOverlapWindow`                           | ✅ 🔀  |
| Disable recording rules (beta)         | switch                                                                                   | `jsonData_disableRecordingRules`               | switch                                  | `jsonData.disableRecordingRules`                                   | ✅     |
| Custom query parameters                | text input                                                                               | `jsonData_customQueryParameters`               | input                                   | `jsonData.customQueryParameters`                                   | ✅     |
| HTTP method                            | select (POST/GET)                                                                        | `jsonData_httpMethod`                          | select                                  | `jsonData.httpMethod`                                              | ✅     |
| Series limit                           | number                                                                                   | `jsonData_seriesLimit`                         | input                                   | `jsonData.seriesLimit`                                             | ✅     |
| Query warning threshold                | number                                                                                   | `jsonData_maxSamplesProcessedWarningThreshold` | input                                   | `jsonData.maxSamplesProcessedWarningThreshold`                     | ✅     |
| Query error threshold                  | number                                                                                   | `jsonData_maxSamplesProcessedErrorThreshold`   | input                                   | `jsonData.maxSamplesProcessedErrorThreshold`                       | ✅     |
| Use series endpoint                    | switch                                                                                   | `jsonData_seriesEndpoint`                      | switch                                  | `jsonData.seriesEndpoint`                                          | ✅     |
| — (hidden `hideExemplars`)             | —                                                                                        | `jsonData_exemplarTraceIdDestinations`         | —                                       | `jsonData.exemplarTraceIdDestinations`                             | ✅ 🔒  |
| — (migration sentinel)                 | —                                                                                        | `jsonData_prometheusTypeMigration`             | —                                       | `jsonData['prometheus-type-migration']`                            | ✅ 🔒  |

Secure fields render as masked secure inputs with a show/hide toggle regardless of the declared
`ui.component`, per the new renderer's `target: "secureJsonData"` policy. Both UIs collect the same
values into the same `secureJsonData` keys.

---

## Gap found and fixed: Custom HTTP Headers

**Legacy behaviour (verified in `legacy-expand-p3-amazonprom.json` / `legacy-expand-p3-amazonprom-verify.json`):**
the editor includes an **HTTP headers** section heading with an **Add header** button
(`hasCustomHeaders: true`, `addHeaderBtn: true`). This comes from `@grafana/plugin-ui`'s `Auth`
wrapper, which renders the `CustomHeaders` component **regardless** of the SigV4-only auth lock
(`visibleMethods=[sigV4Id]` only hides the auth _methods_, not the headers editor). `CustomHeaders`
persists headers as indexed pairs — `jsonData.httpHeaderName<N>` for the name and
`secureJsonData.httpHeaderValue<N>` for the (secret) value, starting at `N = 1`.

**Before:** `dsconfig.json` did not model headers, so the new UI rendered no headers editor.

**Fix (in `dsconfig.json` only):** added the `jsonData_httpHeaders` array field with an `indexedPair`
storage mapping that reproduces the exact legacy storage, plus item sub-fields for the header name
(`http.header.name`, with a header-name pattern validation) and value (`http.header.value`), and
referenced it from the **Advanced HTTP settings** group's `fieldRefs` (after `jsonData_timeout`).

**After (verified in `newgen-amazonprom-fixed.json`):** the new UI renders a **Custom HTTP Headers**
row under **Advanced HTTP settings** with an add-header button and a key/secret-value editor —
`hasHeadersEditor: true`, `urlPresent: true`.

---

## `fileUpload` evaluation — not applicable

The task asked to use the `fileUpload` control **if the legacy UI uses it**. It does not:

- The legacy Amazon Prometheus editor exposes no cert/file inputs — SigV4 credentials are plain
  text/secure inputs. `legacy-expand-p3-amazonprom.json` reports `fileInputs: 0`, `uploadButtons: []`.
- The new UI's `fileUpload` component only activates for `ui.fileMapping` (multi-key JSON
  distribution); it does not model single-value upload.

**Decision:** do **not** add `fileUpload` to any field.

---

## Conditional fields & effects

Auth here is a **discriminator** (`jsonData_sigV4AuthType`), not a virtual selector with `effects`
(that pattern is unique to core Prometheus). Visibility is driven by `dependsOn` expressions on the
discriminator, plus two SDK-managed flags:

| Field                                         | Rule                                                                                 | Behaviour |
| --------------------------------------------- | ------------------------------------------------------------------------------------ | --------- |
| `jsonData_sigV4Auth`                          | forced `true` on mount (`DataSourceHttpSettingsOverhaul.tsx:27-38`)                  | managed ⚙ |
| `jsonData_oauthPassThru`                      | cleared `false` on save (auth locked to SigV4)                                       | managed ⚙ |
| `jsonData_sigV4Profile`                       | `dependsOn: sigV4AuthType == 'credentials'`                                          | 🔀        |
| `secureJsonData_sigV4AccessKey` / `SecretKey` | `dependsOn: sigV4AuthType == 'keys'` (both `requiredWhen`)                           | 🔀        |
| `jsonData_sigV4ExternalId`                    | `dependsOn: sigV4AuthType != 'grafana_assume_role'` (hidden for Grafana Assume Role) | 🔀        |

### Save-payload storage-target validation (the two-fix contract)

The header-routing capture (`capture-headers-generic.js`) filled a custom header and clicked
**Save & Test**, logging the exact datasource payload the wizard would PUT:

```
jsonData.httpHeaderName1: X-Org-Id | secureJsonData.httpHeaderValue1: influx-tenant-9 | url: http://...:8086
```

The header **name lands in `jsonData.httpHeaderName1`** and the **value in
`secureJsonData.httpHeaderValue1`** — byte-for-byte the legacy `CustomHeaders` storage format. The
**URL routes to `root.url`**. (The name/value strings are the generic script's fixtures; the routing
is what matters.)

---

## Conformance (no change required)

Adding an `indexedPair` field would normally break `JSONDataMatchesStruct` /
`JSONDataTypesMatchStruct`, because an `indexedPair` field (`jsonData_httpHeaders`) is a **logical
view** over dynamically-indexed legacy keys (`httpHeaderName1`, `httpHeaderValue1`, …) that are not
modeled as a single Go struct field. The shared conformance walker (`schema/conformance.go`) already
skips `indexedPair` fields via `isIndexedPairField()` (lines 157, 180, 245) — plugin-agnostic — so
**no conformance change was needed**. The per-header `httpHeaderValue<N>` secrets remain dynamic and
are correctly **not** listed among the static `SecureJsonDataKeys` (`settings.go:111-114`, which lists
only `sigV4AccessKey` / `sigV4SecretKey`), and the generated spec emits `httpHeaders` as a clean array
under `jsonData` with **no** secure values leaked (`SchemaSpecHasNoSecureJSON` passes).

---

## Verification

```
go generate ./registry/grafana-amazonprometheus-datasource/...   # regenerate schema.gen.json / settings.gen.json
go test ./registry/grafana-amazonprometheus-datasource/...       # PASS
go test ./registry/grafana-amazonprometheus-datasource/... ./schema/...   # PASS (no regressions)
```

`TestSchemaConformance` subtests — all **PASS**:
`BaseFieldsResolved`, `SchemaRoundTrip`, `SchemaArtifactInSync`, `SchemaSpecHasNoSecureJSON`,
`ConfigSchemaValid`, `JSONDataMatchesStruct`, `JSONDataTypesMatchStruct`,
`SecureValuesMatchLoadSettings`.

The hand-authored `TestLoadConfig`, `TestDefaultExampleShape`, `TestApplyDefaults`, and `TestValidate`
suites also pass unchanged (they do not reference the new field).

New-UI captures: `newgen-amazonprom-fixed` (tab, `hasHeadersEditor: true`, `urlPresent: true`),
`newgen-amazonprom-fixed-wiz` (wizard opens on **General 1/11** with required URL + SigV4 auth group).
Header routing: `capture-headers-generic` (`name → jsonData.httpHeaderName1`,
`value → secureJsonData.httpHeaderValue1`, `url → root.url`).
Legacy capture: `legacy-expand-p3-amazonprom` / `-verify` (HTTP headers + Add header present; 0 file inputs).

---

## Files changed

- [`registry/grafana-amazonprometheus-datasource/dsconfig.json`](dsconfig.json) — changed `root_url`
  from `requiredWhen: "true"` to `required: true` (renders in the wizard's General step); added the
  `jsonData_httpHeaders` `indexedPair` field and referenced it from the `advanced-http` group;
  appended a note to the `connection`/`legacy` instruction stating headers are now modeled.
- [`registry/grafana-amazonprometheus-datasource/schema.gen.json`](schema.gen.json),
  [`registry/grafana-amazonprometheus-datasource/settings.gen.json`](settings.gen.json) — regenerated
  by `go generate` (`url` now in the spec's `required` array; `httpHeaders` array added under
  `jsonData`; `x-dsconfig-required-when: "true"` removed from `url`).

_Unchanged by design:_ `settings.go`, `settings.ts`, `README.md`, `settings.examples.gen.json`,
`schema.go`, `schema/conformance.go`, and everything in `plugin-ui`.

> **Out-of-scope note (not fixable via `dsconfig.json`):** `README.md:168-170` still lists **Custom
> HTTP headers** under **Excluded settings** ("not rendered by this plugin's editor"). That claim is
> now stale — the legacy `Auth` wrapper _does_ render the headers editor (confirmed by the legacy
> capture), and the schema now models it. Correcting the README is outside the allowed edit surface
> for this task (README edits are prohibited); flagging it here for a follow-up.
