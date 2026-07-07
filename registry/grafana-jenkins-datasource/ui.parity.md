# Jenkins — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered
from this entry's [`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `grafana-jenkins-datasource`
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/cfrbqixexd14we` (Grafana Enterprise; the Jenkins plugin's own `ConfigEditor` — a Jenkins **URL** in `jsonData.url`, an optional **User** in `jsonData.username`, and an optional write-only **Password** / API token in `secureJsonData.password`. HTTP Basic auth or anonymous; no auth discriminator.)
- **New UI:** `http://192.168.1.241:58899/iframe.html?args=pluginType:grafana-jenkins-datasource&id=configeditor-datasourceconfigwizard--tab` (Storybook, `ConfigEditor/DatasourceConfigWizard`), also the `--wizard` story.
- **Method:** Playwright captured both UIs (full-page screenshots + DOM extraction). The Storybook story fetches the schema from `.../registry/grafana-jenkins-datasource/dsconfig.json`; the local (edited) `dsconfig.json` was served to the new UI by intercepting that request with `context.route(...)`, so the captures reflect the local schema without pushing. A **before** capture (temp copy with the original `requiredWhen: "true"`) was taken alongside the **after** to isolate the fix's effect.
- **Result:** **Parity achieved.** The one modeled fix corrected `jsonData_url` from a non-canonical `requiredWhen: "true"` to `required: true`, so the URL now (a) emits a canonical OpenAPI `required: ["url"]` in the generated spec and (b) folds into the wizard's synthetic **General** step. Custom HTTP Headers and `fileUpload` were both evaluated and are correctly **not** used (the legacy editor has neither). There are **no** conditional fields (single auth path, single secret).

---

## TL;DR of changes

| #   | Change | File | Why |
| --- | --- | --- | --- |
| 1   | Changed `jsonData_url` from `requiredWhen: "true"` → `required: true` | [`dsconfig.json`](dsconfig.json) | The Jenkins URL is **unconditionally** required — the backend hard-fails an empty URL with "jenkins URL (jsonData.url) is required" (`settings.go:157-159`, mirroring upstream `pkg/plugin/settings.go:23-25`'s `DownstreamError("URL is missing")`). `requiredWhen: "true"` is a non-canonical CEL expression that only emits an `x-dsconfig-required-when` vendor extension, which the OpenAPI `required` array and the wizard's General-step resolver both ignore. `required: true` emits the canonical `required: ["url"]` and is the idiom the wizard's synthetic **General** step recognises. |
| 2   | Regenerated `schema.gen.json`, `settings.gen.json` via `go generate ./registry/grafana-jenkins-datasource/...` | generated artifacts | Keep committed artifacts in sync (guarded by `SchemaArtifactInSync`) |

No changes were made to `settings.go`, `settings.ts`, `README.md`, `conformance_test.go`, `schema/conformance.go`, `settings.examples.gen.json`, or `plugin-ui`. The single schema change flows through `dsconfig.json`; the rest is produced by `go generate`.

**No conformance-test or plugin-ui change was required** — the wizard already folds `required` fields into the **General** step, and the conformance suite already understands `required` on `jsonData` fields (see [Conformance](#conformance-no-change-required)).

---

## Section layout

The `groups` were left untouched. Verified rendering top-to-bottom in the new UI (tab mode,
capture `newgen-jenkins-tab.json`):

| Order | Section (`id`) | `optional` | Fields (in display order) |
| --- | --- | --- | --- |
| 1 | **Connection** (`connection`) | no | URL |
| 2 | **Authentication** (`authentication`) | no | User, Password |

This mirrors the legacy editor's headings (`legacy-expand-jenkins-verify.json`):
`["Connection", "Authentication"]`. New-UI tab capture: `hasHeadersEditor:false`,
`urlPresent:true`, sections `["Connection","Authentication"]` — **identical before and after**
the fix (`newgen-jenkins-before-tab.json` vs `newgen-jenkins-tab.json`), so tab mode did not
regress.

### Wizard mode: URL folds into the "General" step

In **wizard mode** the plugin-ui builds a synthetic first step titled **General**
(`resolveRequiredFieldsGroup`) from (a) every field marked `required: true`, plus (b) the
fields of the auth group (`id: "authentication"`), plus their `dependsOn` parents/children.

- **Before** (`requiredWhen: "true"`): URL lives in the `connection` group and was **not**
  marked `required: true`, so it was excluded from General. The General step (**1/3**) showed
  only **User** and **Password** (the `authentication` group) — `urlPresent:false`
  (`newgen-jenkins-before-wiz.json`: `fieldLabels: ["1/3","User","Password","Save & Test"]`).
- **After** (`required: true`): URL folds into General. The General step (**1/3**) now shows
  **URL** (with a required `*` asterisk) **+ User + Password** — `urlPresent:true`
  (`newgen-jenkins-wiz.json`: `fieldLabels: ["1/3","*","User","Password","Save & Test"]`).

The 3-step wizard count (`1/3`) is unchanged (General + the two groups); only the General
step's contents changed. **User** and **Password** fold into General in both states because
they live in the `authentication` group, which the wizard recognises regardless of the
required marker — so the only observable delta is the newly-present, required **URL**.

**What the fix changes:** both the **generated contract** (`required: ["url"]` — see next
section) **and** the wizard's General-step rendering (URL now surfaces there). The change makes
`dsconfig.json` express the URL's unconditional requirement in the canonical way.

---

## The required-field fix in detail

`jsonData_url` targets `jsonData` (not a secure value), so its `requiredWhen: "true"` was
routed through `fieldToSpecSchema` → `applyConditions` (`dsconfig/convert.go:202-216`), which
emits an `x-dsconfig-required-when: "true"` **vendor extension** — not a canonical `required`
entry. The wizard's General-step resolver and the OpenAPI `required` array both key off
`required: true`, not that CEL extension, so the URL was neither in the spec's `required` array
nor in the wizard's General step.

**Before (`requiredWhen: "true"`):** the generated `jsonData` object had **no** `required`
array, and `url` carried `"x-dsconfig-required-when": "true"`.

**After (`required: true`):** for a `jsonData` field with `f.Required && f.Section == ""`, the
converter appends the key to the object's `required` list (`dsconfig/convert.go:76-80, 96-98`),
so the generated artifacts now emit:

```json
"jsonData": {
  "type": "object",
  "required": ["url"],
  "properties": {
    "url":      { "description": "Jenkins URL, e.g. https://jenkins.example.com", "type": "string" },
    "username": { "description": "The username to use for authentication", "type": "string" }
  }
}
```

The `x-dsconfig-required-when: "true"` extension is gone. This matches the backend's
unconditional requirement (`settings.go:157-159`) and mirrors the established idiom for
unconditionally-required `jsonData` fields (e.g. ServiceNow's `url` / `basicAuthUser`, which
land in the spec's `required` array via the same path).

---

## Field-by-field parity

Legend: ✅ present & matching · ✏️ corrected by this change · 🔒 write-only secret

| Legacy UI field | Control (legacy) | New UI (schema id) | Control (new) | Storage target | Status |
| --- | --- | --- | --- | --- | --- |
| URL | text input | `jsonData_url` | input | `jsonData.url` (required) | ✅ ✏️ |
| User | text input | `jsonData_username` | input | `jsonData.username` (optional) | ✅ |
| Password | password (secure) | `secureJsonData_password` | secure input | `secureJsonData.password` (optional, write-only) | ✅ 🔒 |

Notes:

- **URL is unconditionally required** and routes to **`jsonData.url`** (not `root.url`) — the
  Jenkins plugin reads its base URL from `jsonData.url` and never consults the root datasource
  URL (`settings.go:52-59`). `role: "endpoint.baseUrl"` is preserved.
- **User / Password are intentionally optional.** The backend wires HTTP Basic auth only when
  `jsonData.username` is non-empty (`pkg/plugin/datasource.go:66-71`); an empty username is a
  supported **anonymous access** configuration, and password is only consulted when a username
  is present. Neither UI marks them required, matching the backend (`Validate()` checks only the
  URL — `settings.go:154-161`).
- The **Password** secure field renders as a masked secure input with a show/hide toggle (the
  renderer draws any `target: "secureJsonData"` field that way); it declares
  `ui.component: "input"`. Both UIs collect it into `secureJsonData.password`.
- The `username`/`password` **pair** relationship documented in `dsconfig.json` (Basic auth
  applied only when username is non-empty) is a backend runtime behaviour, not a UI conditional,
  and is preserved unchanged.
- "Name" and "Default" in the legacy DOM are Grafana's standard datasource chrome (instance name
  + "set as default" toggle), present in every datasource editor and intentionally not part of
  the schema.

---

## `fileUpload` evaluation — not applicable to Jenkins

The task asked to use the `fileUpload` control **if the legacy UI uses it**. It does not:

- The legacy Jenkins editor exposes only URL + User + Password. It has **no** TLS cert / key
  fields and **no** file pickers — the legacy DOM capture (`legacy-expand-jenkins-verify.json`)
  reports `fileInputs: 0`, `uploadButtons: []`.
- The new UI's `fileUpload` component only activates for `ui.fileMapping` (multi-key JSON
  distribution, e.g. a GCP service-account file); nothing in Jenkins needs it.

**Decision:** do **not** add `fileUpload` to any Jenkins field.

---

## Custom HTTP Headers — not applicable to Jenkins

The legacy editor has **no** Custom HTTP Headers section (`hasCustomHeaders:false`,
`addHeaderBtn:false` in the legacy capture). Jenkins authenticates purely via HTTP Basic auth
(the `Authorization: Basic` header is set internally from the username/password, not via a
user-editable header). Headers are correctly **not** modeled, and the new UI shows no headers
editor (`hasHeadersEditor:false` in both tab and wizard captures). No change.

---

## Conditional fields — none

Jenkins has a **single** authentication path (HTTP Basic auth or anonymous) with **no** auth
discriminator, `effects`, or `dependsOn`/`requiredWhen` conditionals. After the fix there is no
`requiredWhen` anywhere in the schema (the only occurrence was the unconditional
`"true"` on the URL, now `required: true`). There is nothing conditional to exercise, and
nothing regressed by the fix.

---

## Conformance (no change required)

The change only flips `url` from an `x-dsconfig-required-when` extension to a canonical
`required: ["url"]` entry on the `jsonData` object. The shared conformance suite already
supports `required` on `jsonData` fields, so **no conformance change was needed**:

- `SchemaSpecHasNoSecureJSON` still passes — `url` is a `jsonData` property (not secure), and
  the password remains the sole `secureValues` entry; nothing leaked.
- `JSONDataMatchesStruct` / `JSONDataTypesMatchStruct` still pass — `url` and `username` still
  map 1:1 to the `Config` struct's json-tagged fields (`settings.go:57-64`); adding `url` to the
  `required` array does not change the property set or types.
- `SecureValuesMatchLoadSettings` still matches the single `password` key declared in
  `SecureJsonDataKeys` (`settings.go:38-40`).
- `SchemaArtifactInSync` passes after regeneration.

---

## Verification

```
go generate ./registry/grafana-jenkins-datasource/...   # regenerate schema.gen.json / settings.gen.json
go test ./registry/grafana-jenkins-datasource/...        # ok — PASS
```

`TestSchemaConformance` subtests — all **8 PASS**:
`BaseFieldsResolved`, `SchemaRoundTrip`, `SchemaArtifactInSync`, `SchemaSpecHasNoSecureJSON`,
`ConfigSchemaValid`, `JSONDataMatchesStruct`, `JSONDataTypesMatchStruct`,
`SecureValuesMatchLoadSettings`.

The hand-authored `TestLoadConfig`, `TestApplyDefaults`, and `TestValidate` suites also pass
unchanged (they do not reference the URL's required flag; `Validate()` already hard-fails an
empty URL).

Generated-spec delta: `"required": ["url"]` added to the `jsonData` object and
`"x-dsconfig-required-when": "true"` removed from `url` (both `settings.gen.json` and
`schema.gen.json`); nothing else changed.

New-UI captures: `newgen-jenkins-tab` / `newgen-jenkins-before-tab` (tab,
`hasHeadersEditor:false`, `urlPresent:true`, sections Connection/Authentication — identical
before/after); `newgen-jenkins-wiz` (after — wizard General step **1/3** with required **URL** +
User + Password, `urlPresent:true`); `newgen-jenkins-before-wiz` (before — General step **1/3**
with only User + Password, `urlPresent:false`), confirming the fix folds URL into General with
no other UI change. Legacy capture: `legacy-expand-jenkins-verify` (Connection / Authentication;
`hasCustomHeaders:false`, `addHeaderBtn:false`, `fileInputs:0`, `uploadButtons:[]`).

---

## Files changed

- [`registry/grafana-jenkins-datasource/dsconfig.json`](dsconfig.json) — changed `jsonData_url`
  from `requiredWhen: "true"` to `required: true`.
- [`registry/grafana-jenkins-datasource/schema.gen.json`](schema.gen.json),
  [`registry/grafana-jenkins-datasource/settings.gen.json`](settings.gen.json) — regenerated by
  `go generate` (`jsonData.required: ["url"]` added; `x-dsconfig-required-when: "true"` removed
  from `url`).

_Unchanged by design:_ `settings.go`, `settings.ts`, `README.md`, `settings.examples.gen.json`,
`conformance_test.go`, `schema/conformance.go`, and everything in `plugin-ui`.

_Nothing reported as out-of-scope / unfixable:_ the only requested fix (required-field /
General-step) was fully achievable through `dsconfig.json` alone. Custom HTTP Headers and
`fileUpload` are correctly n/a for this plugin (legacy has neither), so no
`settings.go`/conformance/plugin-ui coordination was needed.
