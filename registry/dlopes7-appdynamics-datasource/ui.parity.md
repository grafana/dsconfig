# AppDynamics — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered
from this entry's [`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `dlopes7-appdynamics-datasource`
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/bfrbqiw9v14hsc` (Grafana Enterprise, `@grafana/plugin-ui` `DataSourceHttpSettings` + AppDynamics `ConfigEditor`)
- **New UI:** `http://192.168.1.241:58899/?path=/story/configeditor-datasourceconfigwizard--tab&args=pluginType:dlopes7-appdynamics-datasource` (Storybook, `ConfigEditor/DatasourceConfigWizard`), also the `--wizard` story.
- **Method:** Playwright captured both UIs (full-page screenshots + DOM extraction + `Save & Test`-button console payloads). The Storybook story fetches the schema from `raw.githubusercontent.com/.../registry/dlopes7-appdynamics-datasource/dsconfig.json`; the local (edited) `dsconfig.json` was served to the new UI by intercepting that request with `context.route(...)`, so the captures reflect the local schema without pushing.
- **Result:** **Parity achieved.** One missing field was found (**Custom HTTP Headers**) and added; `fileUpload` was evaluated and correctly **not** used (legacy has 0 file inputs); `root_url` was promoted to `required: true` so it renders in the wizard's synthetic **General** step; the `dependsOn` basic-auth reveal and the header storage routing were exercised and verified from the save payload.

---

## TL;DR of changes

| #   | Change                                                                                           | File                             | Why                                                                                                            |
| --- | ------------------------------------------------------------------------------------------------ | -------------------------------- | -------------------------------------------------------------------------------------------------------------- |
| 1   | Changed `root_url` from `requiredWhen: "true"` → `required: true`                                | [`dsconfig.json`](dsconfig.json) | URL is unconditionally required (`Validate()` rejects an empty Controller URL); puts URL into the wizard's synthetic **General** step and emits `required: ["url"]` |
| 2   | Added **Custom HTTP Headers** field (`jsonData_httpHeaders`) with `indexedPair` storage           | [`dsconfig.json`](dsconfig.json) | Legacy UI renders a **Custom HTTP Headers** section with **Add header**; new UI had no headers editor          |
| 3   | Added `jsonData_httpHeaders` to the **`authentication`** group's `fieldRefs` (at the end)         | [`dsconfig.json`](dsconfig.json) | No `advanced-http` group exists; the auth group holds the HTTP settings, and legacy renders headers right after Auth |
| 4   | Reworded the auth/connection instruction: headers are now **modeled** (removed from the "ignored" DataSourceHttpSettings list) | [`dsconfig.json`](dsconfig.json) | Keep the embedded `llm`-tagged instructions truthful after change #2                                           |
| 5   | Regenerated `schema.gen.json`, `settings.gen.json` via `go generate ./registry/dlopes7-appdynamics-datasource/...` | generated artifacts              | Keep committed artifacts in sync (guarded by `SchemaArtifactInSync`)                                          |

No changes were made to `settings.go`, `settings.ts`, `README.md`, `conformance.go`, or `plugin-ui`. All schema changes flow through `dsconfig.json`; the rest is produced by `go generate`.

**No conformance-test or plugin-ui change was required.** The shared conformance walker already skips `indexedPair` fields, and the plugin-ui wizard already recognises the conventional `authentication` group id — both were in place from the graphite work.

---

## Section layout

The `groups` were left in their existing taxonomy; only the new field was slotted into
**Authentication**. Verified rendering top-to-bottom in the new UI (tab mode):

| Order | Section (`id`) | `optional` | Fields (in display order) |
| --- | --- | --- | --- |
| 1 | **Connection** (`connection`) | no | URL |
| 2 | **Authentication** (`authentication`) | no | Basic auth → User → Password, Client Name → Client Domain → Client Secret, **Custom HTTP Headers** ➕ |
| 3 | **Analytics** (`analytics`) | yes | Analytics API URL, Global Account Name, Analytics API Key |
| 4 | **TLS settings** (`tls-settings`) | yes | Skip TLS Verify |

### Header placement decision

The task suggested the **`authentication`** group, and that is where I placed
`jsonData_httpHeaders` (at the end of its `fieldRefs`). This is the correct home:

- There is **no `advanced-http` group** in this schema. The AppDynamics `authentication`
  group already carries the HTTP-transport / auth knobs (basic auth, API Client), i.e. the
  fields that the legacy stock `DataSourceHttpSettings` block renders.
- The **legacy section order is `Metrics → HTTP → Auth → Custom HTTP Headers → Analytics`**
  (verified: `legacy-expand-appd-parity.json` `headings`), so Custom HTTP Headers sits
  immediately after Auth. Appending it to the end of the `authentication` group reproduces
  that adjacency and keeps it out of the optional Analytics / TLS accordions.

### Wizard mode: URL in the "General" step

In **wizard mode** the plugin-ui builds a synthetic first step titled **General**
(`resolveRequiredFieldsGroup`) from (a) every field marked `required: true`, plus (b) the
fields of the auth group, plus their `dependsOn` parents/children.

- **Before:** `root_url` used `requiredWhen: "true"` (a CEL expression the resolver does **not**
  inspect), so URL was **not** in General — the new-UI wizard capture showed `urlPresent: false`
  (`newgen-appd-before-wiz.json`).
- **After:** changing it to `required: true` puts URL into General and emits a proper OpenAPI
  `required: ["url"]` in the generated spec (instead of the `x-dsconfig-required-when: "true"`
  extension). This is unconditionally correct — `Validate()` (settings.go) and the upstream
  health check reject a datasource with no Controller URL.

**Verified (`newgen-appd-fixed-wiz.json`):** the wizard opens on **General 1/5** containing
**URL \*** (with the required asterisk), the API Client fields (**Client Name / Client Domain /
Client Secret**), the **Basic auth** switch, and the **Custom HTTP Headers** editor with an
**Add custom http header** button. `urlPresent: true`, `hasHeadersEditor: true`. (User / Password
are correctly hidden until the **Basic auth** switch is on — `dependsOn: root_basicAuth == true`.)

Because the auth group uses the conventional `id: "authentication"` (which the plugin-ui wizard
already recognises), **no plugin-ui change was needed**.

---

## Field-by-field parity

Legend: ✅ present & matching · ➕ added by this change · 🔀 conditional (revealed by `dependsOn`)

| Legacy UI field | Control (legacy) | New UI (schema id) | Control (new) | Storage target | Status |
| --- | --- | --- | --- | --- | --- |
| URL | text input | `root_url` | input | `root.url` (required) | ✅ |
| Basic auth | switch | `root_basicAuth` | switch | `root.basicAuth` | ✅ |
| User | text input | `root_basicAuthUser` | input | `root.basicAuthUser` | ✅ 🔀 |
| Password | password (secure) | `secureJsonData_basicAuthPassword` | secure input | `secureJsonData.basicAuthPassword` | ✅ 🔀 |
| Client Name | text input | `jsonData_clientName` | input | `jsonData.clientName` | ✅ |
| Client Domain | text input | `jsonData_clientDomain` | input | `jsonData.clientDomain` | ✅ |
| Client Secret | password (secure) | `secureJsonData_clientSecret` | secure input | `secureJsonData.clientSecret` | ✅ |
| **Custom HTTP Headers** | Add header → name input + value password | `jsonData_httpHeaders` | IndexedPair editor (key + secret value) | `jsonData.httpHeaderName<N>` / `secureJsonData.httpHeaderValue<N>` | ➕ |
| Analytics API URL | select (allowCustom) | `jsonData_analyticsURL` | select (allowCustom) | `jsonData.analyticsURL` | ✅ |
| Global Account Name | text input | `jsonData_globalAccountName` | input | `jsonData.globalAccountName` | ✅ |
| Analytics API Key | password (secure) | `secureJsonData_analyticsAPIKey` | secure input | `secureJsonData.analyticsAPIKey` | ✅ |
| Skip TLS Verify | switch | `jsonData_tlsSkipVerify` | switch | `jsonData.tlsSkipVerify` | ✅ |

**Note on secure inputs:** `dsconfig.json` declares `ui.component: "input"` for the three secret
fields (Password, Client Secret, Analytics API Key), but the new renderer draws any
`target: "secureJsonData"` field as a masked secure input with a show/hide toggle. Both UIs
collect the same secret into the same `secureJsonData` key — only the widget affordance differs.

### Auth model — no virtual selector (differs from prometheus)

Unlike prometheus, AppDynamics has **no `virtual_authMethod` discriminator**. The backend picks
the Controller auth method at runtime from which credentials are present (`clientSecret` wins;
otherwise basic auth). The schema models this directly, so there are **no `effects` to exercise**;
instead the cross-field contract is expressed with `dependsOn` + `requiredWhen`:

| Trigger / relationship | Effect | Verified |
| --- | --- | --- |
| `root_basicAuth == true` (`dependsOn`) | User + Password revealed | ✅ (hidden by default in the wizard General step; declared reveal) |
| API Client group `requiredWhen` (`clientName` / `clientDomain` / `clientSecret`) | setting any one makes the other two required (all-or-nothing) | ✅ (declared; `relationships[type=group]`) |
| Basic-auth pair `requiredWhen` (`basicAuthUser` / `basicAuthPassword`) | both required together | ✅ (declared; `relationships[type=pair]`) |

### Save-payload storage-target validation (Custom HTTP Headers)

Filling the form and clicking **Save & Test** logs the exact datasource payload the wizard would
PUT. Custom header (name `X-Org-Id`, value `influx-tenant-9`) — `capture-headers-generic` console:

```
jsonData.httpHeaderName1: X-Org-Id | secureJsonData.httpHeaderValue1: influx-tenant-9 | url: http://influx.example.com:8086
```

The header **name lands in `jsonData.httpHeaderName1`** and the **value in
`secureJsonData.httpHeaderValue1`** — byte-for-byte the legacy `@grafana/plugin-ui`
CustomHeaders storage format. URL routes to `root.url`.

---

## Gap found and fixed: Custom HTTP Headers

**Legacy behaviour (verified in `legacy-expand-appd-parity.json`):** the editor includes a
**Custom HTTP Headers** section heading with an **Add header** button (`hasCustomHeaders: true`,
`addHeaderBtn: true`; section order `Metrics, HTTP, Auth, Custom HTTP Headers, Analytics`).
`@grafana/plugin-ui`'s CustomHeaders component persists headers as indexed pairs —
`jsonData.httpHeaderName<N>` for the name and `secureJsonData.httpHeaderValue<N>` for the
(secret) value, starting at `N = 1`.

**Before:** `dsconfig.json` did not model headers at all (the old instruction listed "custom
headers" among the DataSourceHttpSettings fields the plugin ignores), so the new UI rendered no
headers editor (`hasHeadersEditor: false` in `newgen-appd-before-tab.json`).

**Fix (in `dsconfig.json` only):** added the `jsonData_httpHeaders` array field with an
`indexedPair` storage mapping that reproduces the exact legacy storage, plus item sub-fields for
the header name (`http.header.name`, with a header-name pattern validation) and value
(`http.header.value`), and added it to the **`authentication`** group's `fieldRefs`.

**After (verified in `newgen-appd-fixed-tab.json` / `newgen-appd-fixed-wiz.json`):** the new UI
renders a **Custom HTTP Headers** row under **Authentication** with an **Add custom http header**
button and a key/secret-value editor (`hasHeadersEditor: true` in both tab and wizard).

### Backend caveat (out of scope for UI parity)

Per the existing (source-cited) instruction, this plugin's HTTP client only honors
`tlsSkipVerify` + proxy options (`pkg/appd/client/client.go:29-44`) and does **not** consume the
`httpHeaderName<N>`/`httpHeaderValue<N>` keys at request time — the same is true of the other
`DataSourceHttpSettings` knobs the legacy editor renders. The header field is therefore modeled
for **config-editor parity and provisioning fidelity** (matching what the legacy UI shows and how
it stores), not to change backend behavior. This is a backend property of the upstream plugin, not
a UI-parity gap, and is **not fixable (nor in scope) via `dsconfig.json`**. The instruction block
was updated to state this truthfully rather than claiming headers are "ignored/not modeled".

---

## `fileUpload` evaluation — not applicable to AppDynamics

The task asked to use the `fileUpload` control **if the legacy UI uses it**. It does not:

- The legacy AppDynamics editor exposes **no file inputs and no upload buttons**
  (`legacy-expand-appd-parity.json`: `fileInputs: 0`, `uploadButtons: []`). Secrets
  (Password, Client Secret, Analytics API Key) are pasted into text/secure inputs, and the
  plugin does not surface the TLS client-certificate PEM fields at all.
- The new UI's `fileUpload` component only activates for `ui.fileMapping` (multi-key JSON
  distribution, e.g. a GCP service-account file); nothing here needs it.

**Decision:** do **not** add `fileUpload` to any AppDynamics field.

**Packs:** none evaluated/needed — there are no multi-field bundle controls in the legacy editor.

---

## Conformance (no change required)

Adding an `indexedPair` field would normally break `JSONDataMatchesStruct` /
`JSONDataTypesMatchStruct`, because an `indexedPair` field (`jsonData_httpHeaders`) is a **logical
view** over dynamically-indexed legacy keys (`httpHeaderName1`, `httpHeaderValue1`, …) that are
not modeled as a single Go struct field. The shared conformance walker (`schema/conformance.go`)
already skips `indexedPair` fields via `isIndexedPairField()` (lines 157, 180, 245-247) — this is
plugin-agnostic — so **no conformance change was needed**. The per-header `httpHeaderValue<N>`
secrets remain dynamic and are correctly **not** listed among the static `SecureJsonDataKeys`
(`settings.go:59-63`), and the generated spec emits `httpHeaders` as a clean array under
`jsonData` with **no** secure values leaked (`SchemaSpecHasNoSecureJSON` passes).

---

## Verification

```
go generate ./registry/dlopes7-appdynamics-datasource/...   # regenerate schema.gen.json / settings.gen.json
go test ./registry/dlopes7-appdynamics-datasource/...       # PASS
go test ./registry/dlopes7-appdynamics-datasource/... ./schema/...   # PASS (no regressions)
```

`TestSchemaConformance` subtests — all **PASS**:
`BaseFieldsResolved`, `SchemaRoundTrip`, `SchemaArtifactInSync`, `SchemaSpecHasNoSecureJSON`,
`ConfigSchemaValid`, `JSONDataMatchesStruct`, `JSONDataTypesMatchStruct`,
`SecureValuesMatchLoadSettings`.

The hand-authored `TestLoadConfig`, `TestApplyDefaults`, `TestValidate`,
`TestIsAnalyticsConfigured`, and `TestSettingsExamples` suites also pass unchanged (they do not
reference the new field).

New-UI captures: `newgen-appd-before-tab` / `newgen-appd-before-wiz` (before: `hasHeadersEditor:
false`; wizard `urlPresent: false`), `newgen-appd-fixed-tab` (`hasHeadersEditor: true`,
`urlPresent: true`), `newgen-appd-fixed-wiz` (wizard opens on **General 1/5** with required URL +
Custom HTTP Headers), `capture-headers-generic` console (header save-payload routing).
Legacy capture: `legacy-expand-appd-parity` (Custom HTTP Headers + Add header present; 0 file
inputs; headings `Metrics, HTTP, Auth, Custom HTTP Headers, Analytics`).

---

## Files changed

- [`registry/dlopes7-appdynamics-datasource/dsconfig.json`](dsconfig.json) — changed `root_url`
  from `requiredWhen: "true"` to `required: true` (renders in the wizard's General step); added
  the `jsonData_httpHeaders` `indexedPair` field and referenced it from the `authentication`
  group; reworded the auth/connection instruction so it states headers are now modeled (removed
  "custom headers" from the ignored-DataSourceHttpSettings list and added the storage-format +
  backend caveat).
- [`registry/dlopes7-appdynamics-datasource/schema.gen.json`](schema.gen.json),
  [`registry/dlopes7-appdynamics-datasource/settings.gen.json`](settings.gen.json) — regenerated
  by `go generate` (`url` now in the spec's `required` array; `httpHeaders` array added under
  `jsonData`; `x-dsconfig-required-when: "true"` removed from `url`).

_Unchanged by design:_ `settings.go`, `settings.ts`, `README.md`, `settings.examples.gen.json`,
`schema/conformance.go`, and everything in `plugin-ui`.
