# Databricks — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered
from this entry's [`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `grafana-databricks-datasource` (an **enterprise** datasource using the Databricks SQL connector; the legacy editor is the plugin's own `src/ConfigEditor.tsx`, not the `@grafana/sql` shared editor)
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/bfrbqiwx630n4b` (Grafana Enterprise 13.0.1)
- **New UI:** `http://192.168.1.241:58899/iframe.html?args=pluginType:grafana-databricks-datasource&id=configeditor-datasourceconfigwizard--tab` (Storybook, `ConfigEditor/DatasourceConfigWizard`; also `--wizard`)
- **Method:** Playwright captured both UIs (full-page screenshots + DOM extraction + step/section/conditional probing). The Storybook story fetches the schema from `raw.githubusercontent.com/.../registry/grafana-databricks-datasource/dsconfig.json`; the local (edited) `dsconfig.json` was served to the new UI by intercepting that request with `context.route(...)`, so the screenshots reflect the local schema without pushing.
- **Result:** **Parity achieved.** All modeled fields render in the new UI and route to the same storage targets. The one change required was making the two unconditionally-required connection fields (`host`, `httpPath`) use `required: true` so the wizard's synthetic **General** step pulls them in. Databricks talks to its SQL warehouse over the Databricks SQL connector, so it correctly has **no HTTP-headers editor** and **no file-upload** control (the legacy editor has neither). All five `authType` conditionals were exercised and reveal the correct auth fields. One renderer-level observation (the sdk-managed `azureCredentials` composite for Azure On-Behalf-Of) is documented below; it is out of scope for the required-field fix and not fixable via `dsconfig.json`.

---

## TL;DR of changes

| #   | Change | File | Why |
| --- | --- | --- | --- |
| 1   | Changed **`jsonData_host`** and **`jsonData_httpPath`** from `"requiredWhen": "true"` → `"required": true` | [`dsconfig.json`](dsconfig.json) | These are unconditionally required for every auth method (the backend `Validate()` returns `ErrMissingHost` / `ErrMissingHTTPPath` when either is empty — `settings.go:212-217`). The wizard's synthetic **General** step only pulls fields with `required: true`; `requiredWhen` is a CEL expression the resolver does not inspect for the General step. Also emits a proper OpenAPI `required` array instead of the `x-dsconfig-required-when` extension. |
| 2   | Regenerated `schema.gen.json`, `settings.gen.json` via `go generate ./registry/grafana-databricks-datasource/...` | generated artifacts | Keep committed artifacts in sync (guarded by `SchemaArtifactInSync`). |

No changes were made to `settings.go`, `settings.ts`, `README.md`, `conformance_test.go`, `schema.go`, or `plugin-ui`.
The `requiredWhen` conditions that are **real** conditions (the auth-type-gated
`token` / `azureCredentials` / `azureClientSecret` / `clientId` / `clientSecret` / `tenantId`
expressions) were **left untouched** — only the two literal `"requiredWhen": "true"` values were
converted.

No `conformance_test.go` change was needed here. No `plugin-ui` change was needed either: the auth
group already uses the conventional id `authentication`, which the wizard's required-fields resolver
already recognises, so the auth fields fold into General correctly.

---

## Section layout

Verified rendering top-to-bottom in the new UI (tab mode).

| Order | Section (`id`) | `optional` | Fields (in display order) |
| --- | --- | --- | --- |
| 1 | **Connection** (`connection`) | no | Host, Http Path |
| 2 | **Authentication** (`authentication`) | no | Authentication Type, then the auth-type-gated secret/credential fields (see conditionals) |
| 3 | **Additional settings** (`additional-settings`) | no | Retries, Pause, Timeout, Max Rows, Retry Timeout, Debug, Unity Catalog Support, Default Query Format |

The new UI renders these three groups as titled panels with a left-hand "Connect data source"
stepper (Connection → Authentication → Additional settings) and a "Fields marked with * are
required" hint. The **legacy** enterprise editor lays the same fields out in a **single flat form**
and emits **no `h1`–`h6`/`legend` section headings** (DOM capture `legacy-expand-databricks-parity`:
`headings: []`), so there is no 1:1 heading match to assert — the three groups are the schema's own
grouping of the identical field set.

### Wizard mode: the "General" step

In **wizard mode** the plugin-ui builds a synthetic first step titled **General**
(`resolveRequiredFieldsGroup`) from (a) every field marked `required: true`, plus (b) every field in
the auth group (`authentication`), plus their `dependsOn` parents/children. The stepper reports
**General 1/4** (General + the three groups).

**Effect of the `required: true` fix (before/after, both captured with Playwright):**

| Field | Before (`requiredWhen: "true"`) | After (`required: true`) |
| --- | --- | --- |
| Host (`jsonData_host`, Connection group) | **absent** from General | **present** in General ✅ (`Host *`) |
| Http Path (`jsonData_httpPath`, Connection group) | **absent** from General | **present** in General ✅ (`Http Path *`) |
| Authentication Type (`jsonData_authType`) | present (auth-group member) | present ✅ |
| Token (`secureJsonData_token`) | present (auth-group member, `Token *`) | present ✅ |

Before the fix the wizard's General step showed only **Authentication Type** and **Token** — the two
connection fields were only reachable in the `Connection` step
(`verify-databricks-wizard-before.json`: `hostVisible:false httpPathVisible:false`). After the fix
all required fields appear in General (`verify-databricks-wizard-after.json`:
`hostVisible:true httpPathVisible:true`, both carrying `*` markers). Tab mode is unaffected — the
synthetic `_required` group is filtered out there, so it still shows the three sections in order.

---

## Field-by-field parity

Legend: ✅ present & matching · 🔀 conditional (revealed by `dependsOn`) · ⚙️ no editor UI (backend/side-effect)

| Legacy UI field | Control (legacy) | New UI (schema id) | Control (new) | Storage target | Status |
| --- | --- | --- | --- | --- | --- |
| Host \* | text input | `jsonData_host` | input | `jsonData.host` | ✅ |
| Http Path \* | text input | `jsonData_httpPath` | input | `jsonData.httpPath` | ✅ |
| Authentication Type | select (5 options) | `jsonData_authType` | select | `jsonData.authType` | ✅ |
| Token | password (`SecretInput`) | `secureJsonData_token` | secure input | `secureJsonData.token` | ✅ 🔀 |
| Client ID | text input | `jsonData_clientId` | input | `jsonData.clientId` | ✅ 🔀 |
| Client Secret (M2M) | password | `secureJsonData_clientSecret` | secure input | `secureJsonData.clientSecret` | ✅ 🔀 |
| Directory (tenant) ID | text input | `jsonData_tenantId` | input | `jsonData.tenantId` | ✅ 🔀 |
| Azure Cloud | select (Azure/China/US Gov) | `jsonData_azureCloud` | select | `jsonData.azureCloud` | ✅ 🔀 |
| Client Secret (OBO) | password | `secureJsonData_azureClientSecret` | secure input | `secureJsonData.azureClientSecret` | ✅ 🔀 |
| Azure credentials (OBO) | `@grafana/azure-sdk` `AzureCredentialsForm` | `jsonData_azureCredentials` | *(not rendered — see note)* | `jsonData.azureCredentials` | ⚙️ 🔀 ¹ |
| *(none — side-effect)* | — | `jsonData_oauthPassThru` | — | `jsonData.oauthPassThru` | ⚙️ |
| Retries | text input | `jsonData_retries` | input | `jsonData.retries` | ✅ |
| Pause | text input | `jsonData_pause` | input | `jsonData.pause` | ✅ |
| Timeout | text input | `jsonData_timeout` | input | `jsonData.timeout` | ✅ |
| Max Rows | text input | `jsonData_rows` | input | `jsonData.rows` | ✅ |
| Retry Timeout | text input | `jsonData_retryTimeout` | input | `jsonData.retryTimeout` | ✅ |
| Debug | switch/checkbox | `jsonData_debug` | checkbox | `jsonData.debug` | ✅ |
| Unity Catalog Support | switch/checkbox | `jsonData_enableUnitySupport` | checkbox | `jsonData.enableUnitySupport` | ✅ |
| Default Query Format | select (Timeseries/Table) | `jsonData_defaultQueryFormat` | select | `jsonData.defaultQueryFormat` | ✅ |
| *(none — backend-only)* | — | `jsonData_cloudFetch` | — | `jsonData.cloudFetch` | ⚙️ |

All **20 modeled fields** are accounted for: 18 render in the new UI (2 connection + 8 authentication
+ 8 additional settings), and 2 (`oauthPassThru`, `cloudFetch`) are intentionally UI-less
side-effect / backend-only fields with no legacy control either. The legacy `Name` and `Default`
controls at the top are Grafana editor chrome (datasource name + default toggle), not part of the
datasource config, and are correctly **not** modeled in `dsconfig.json`.

¹ See the Azure On-Behalf-Of note under **Conditional fields & effects**.

---

## Gaps found

**None beyond the required-field fix.** The complete legacy field set is already modeled in
`dsconfig.json`, so no field had to be added. Databricks keeps its fields **inline** (no `packs`).

### Custom HTTP Headers — not applicable (verified)

Databricks connects to its SQL warehouse through the Databricks SQL connector (built from
`jsonData.host` + `jsonData.httpPath`), **not** through Grafana's generic HTTP-headers datasource
plumbing, so it has no Custom-HTTP-Headers concept. Legacy DOM capture
(`legacy-expand-databricks-parity`): `hasCustomHeaders: false`, `addHeaderBtn: false`. New UI (tab +
wizard): `hasHeadersEditor: false`. Correctly **not** added.

---

## `fileUpload` evaluation — not applicable to databricks

The task asked to use the `fileUpload` control **if the legacy UI uses it**. It does not:

- The legacy Databricks editor has **no `<input type="file">` and no upload button**
  (`fileInputs: 0`, `uploadButtons: []` in every capture). All credentials are plain text / secret
  inputs.
- The new UI's `fileUpload` component only activates when a field declares `ui.fileMapping`
  (multi-key JSON distribution); no databricks field needs that.

**Decision:** do **not** add `fileUpload` to any databricks field.

---

## Conditional fields & effects — tested

Databricks drives its conditionals with a single **`Authentication Type` select** (`jsonData.authType`,
5 values). Each scenario was run on a fresh page and the visible auth inputs probed by placeholder
(`verify-databricks-conditionals.js`; secret fields detected as inputs, the shared
`clientSecret`/`tenantId` placeholder counted). `host` and `httpPath` stayed visible in **all** five
scenarios, confirming they are unconditional connection fields.

| Scenario (`authType`) | Token | Client ID | Client Secret / tenant inputs | azureClientSecret | Azure Cloud | Matches schema `dependsOn`? |
| --- | --- | --- | --- | --- | --- | --- |
| **A** `Pat` (default) | **shown** | hidden | 0 | hidden | hidden | ✅ token gated by `authType == 'Pat' \|\| authType == ''` |
| **B** `OauthPT` | hidden | hidden | 0 | hidden | hidden | ✅ passthrough needs no stored secret |
| **C** `OauthM2M` | hidden | **shown** | 1 (clientSecret) | hidden | hidden | ✅ `clientId` + `clientSecret` gated by `OauthM2M \|\| AzureM2M` |
| **D** `AzureM2M` | hidden | **shown** | 2 (clientSecret + tenantId) | hidden | **shown** | ✅ adds `tenantId` + `azureCloud` gated by `AzureM2M` |
| **E** `OauthOBO` | hidden | hidden | 0 | **shown** | hidden | ✅ `azureClientSecret` gated by `OauthOBO` |

Observed transitions, exactly matching the schema:

- **`Pat`** (default) reveals only **Token**; every OAuth/Azure field is gated off.
- **`OauthPT`** reveals no secret fields at all (the backend uses the forwarded user OAuth token).
- **`OauthM2M`** reveals **Client ID** + **Client Secret** (`clientId` / `clientSecret`).
- **`AzureM2M`** reveals **Client ID**, **Client Secret**, **Directory (tenant) ID**, and
  **Azure Cloud** (`clientId` / `clientSecret` / `tenantId` / `azureCloud`, default `AzureCloud`).
  Both `*`-required markers appear on Client ID / Client Secret / Directory (tenant) ID because
  their `requiredWhen` conditions are active.
- **`OauthOBO`** reveals the **Client Secret** (`azureClientSecret`) input.

**Azure On-Behalf-Of note (¹).** The `jsonData_azureCredentials` field has **no `ui` component** — it
is tagged `sdk-managed` and, in the real enterprise editor, is rendered by the `@grafana/azure-sdk`
`AzureCredentialsForm` (a composite tenant/client/cloud sub-form). The schema-driven `plugin-ui`
renderer does **not** embed `@grafana/azure-sdk` components, so under `OauthOBO` the new UI renders
only the `azureClientSecret` secret input and not the composite credentials sub-form
(`verify-databricks-cond-E-oauthOBO`: the only extra visible input is `Client Secret`). This is a
**renderer limitation in `plugin-ui`, not a schema gap** — the field is modeled, tagged, and
documented in `dsconfig.json`, and its storage target (`jsonData.azureCredentials`) is correct. It is
**out of scope** for the required-field fix and **not fixable via `dsconfig.json`** (it would require
`plugin-ui` to support the azure-sdk composite component). Reported, not changed.

**Effects:** databricks's schema contains **no** `effects` blocks. Its auth visibility is a set of
plain `dependsOn` CEL expressions over the single `authType` select; there is no virtual selector that
fans out to write multiple fields, so nothing for `effects` to model, and none were added.

---

## Verification

```
go generate ./registry/grafana-databricks-datasource/...   # regenerate schema.gen.json / settings.gen.json
go test ./registry/grafana-databricks-datasource/...        # 8/8 conformance subtests + unit tests PASS
go test ./registry/grafana-databricks-datasource/... ./schema/...   # PASS (no regressions)
```

Conformance subtests (databricks): `BaseFieldsResolved`, `SchemaRoundTrip`, `SchemaArtifactInSync`,
`SchemaSpecHasNoSecureJSON`, `ConfigSchemaValid`, `JSONDataMatchesStruct`, `JSONDataTypesMatchStruct`,
`SecureValuesMatchLoadSettings` — all **PASS**. Unit tests (`TestLoadConfig`, `TestApplyDefaults`,
`TestValidate`) — all **PASS**.

After regeneration, `settings.gen.json` / `schema.gen.json` moved `host` + `httpPath` into the
`jsonData` `required` array and dropped the two `x-dsconfig-required-when: "true"` extensions.

---

## Files changed

- [`registry/grafana-databricks-datasource/dsconfig.json`](dsconfig.json) — changed
  `jsonData_host` and `jsonData_httpPath` from `"requiredWhen": "true"` to `"required": true`
  (so they render in the wizard's General step and emit OpenAPI `required`). The six real
  auth-type-gated `requiredWhen` conditions were left untouched.
- [`registry/grafana-databricks-datasource/schema.gen.json`](schema.gen.json),
  [`registry/grafana-databricks-datasource/settings.gen.json`](settings.gen.json) — regenerated by
  `go generate` (`host`/`httpPath` now in the spec `required` arrays).

_Unchanged by design:_ `settings.go`, `settings.ts`, `README.md`, `settings.examples.gen.json`,
`schema.go`, `conformance_test.go`, and `plugin-ui`.
