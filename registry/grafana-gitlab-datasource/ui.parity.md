# GitLab — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered
from this entry's [`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `grafana-gitlab-datasource`
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/afrbqix6y6pdsa` (Grafana Enterprise; the GitLab plugin's own `ConfigEditor` — root `url` + a single write-only personal access token in `secureJsonData.accessToken`, plus an optional `jsonData.pageLimit`)
- **New UI:** `http://192.168.1.241:58899/iframe.html?args=pluginType:grafana-gitlab-datasource&id=configeditor-datasourceconfigwizard--tab` (Storybook, `ConfigEditor/DatasourceConfigWizard`), also the `--wizard` story.
- **Method:** Playwright captured both UIs (full-page screenshots + DOM extraction + `Save & Test`-button console payloads). The Storybook story fetches the schema from `raw.githubusercontent.com/.../schema-discovery/registry/grafana-gitlab-datasource/dsconfig.json`; the local (edited) `dsconfig.json` was served to the new UI by intercepting that request with `context.route(...)`, so the captures reflect the local schema without pushing.
- **Result:** **Parity achieved.** The one modeled fix was correcting `secureJsonData_accessToken` from a non-canonical `requiredWhen: "true"` to `required: true` so the **generated OpenAPI spec** now marks the token required (it previously did not — see below). Custom HTTP Headers and `fileUpload` were both evaluated and are correctly **not** used (the legacy editor has neither). There are **no** conditional fields (single auth method, single secret).

---

## TL;DR of changes

| #   | Change | File | Why |
| --- | --- | --- | --- |
| 1   | Changed `secureJsonData_accessToken` from `requiredWhen: "true"` → `required: true` | [`dsconfig.json`](dsconfig.json) | The access token is **unconditionally** required — the backend hard-fails an empty token with "access token … can not be blank" (`settings.go:176-181`, mirroring `pkg/models/settings.go:48-50`). `requiredWhen: "true"` is a non-canonical CEL expression that the converter **silently drops for secure fields**, so the generated spec did not mark the token required. `required: true` emits the canonical `"required": true` on the secure value and is the idiom the wizard's synthetic **General** step recognises. |
| 2   | Regenerated `schema.gen.json`, `settings.gen.json` via `go generate ./registry/grafana-gitlab-datasource/...` | generated artifacts | Keep committed artifacts in sync (guarded by `SchemaArtifactInSync`) |

No changes were made to `settings.go`, `settings.ts`, `README.md`, `conformance_test.go`, `schema/conformance.go`, `settings.examples.gen.json`, or `plugin-ui`. The single schema change flows through `dsconfig.json`; the rest is produced by `go generate`.

**No conformance-test or plugin-ui change was required** — the wizard already folds the `authentication` group + `required` fields into the **General** step, and the conformance suite already understands `required` on secure values (see [Conformance](#conformance-no-change-required)).

---

## Section layout

The `groups` were left untouched. Verified rendering top-to-bottom in the new UI (tab mode,
capture `newgen-gitlab-after-tab.json`):

| Order | Section (`id`) | `optional` | Fields (in display order) |
| --- | --- | --- | --- |
| 1 | **Connection** (`connection`) | no | URL |
| 2 | **Authentication** (`authentication`) | no | Access token |
| 3 | **Additional Settings** (`additional-settings`) | yes | Page limit |

This mirrors the legacy editor's headings (`legacy-expand-gitlab-verify.json`):
`["Connection", "Authentication", "Gitlab authentication", "Additional Settings", "Page limit"]`
("Gitlab authentication" is the plugin's sub-heading over the access-token field). New-UI
tab capture: `hasHeadersEditor:false`, `urlPresent:true`, sections `["Connection","Authentication"]`
(+ the `optional` **Additional Settings** accordion).

### Wizard mode: access token in the "General" step

In **wizard mode** the plugin-ui builds a synthetic first step titled **General**
(`resolveRequiredFieldsGroup`) from (a) every field marked `required: true`, plus (b) the
fields of the auth group (`id: "authentication"`), plus their `dependsOn` parents/children.

- The **Access token** renders on the **General** step (step **1/4**) in both the before and
  after captures — because it lives in the `authentication` group, which the wizard folds into
  General **regardless** of the required marker. Verified `hasToken:true`, placeholder
  `"Access token [password]"` (`verify-gitlab-wizard.json`).
- The token shows a **required asterisk** and **Save & Test is blocked while it is empty** in
  both before and after (`verify-gitlab-enforce-before.json` / `verify-gitlab-enforce-after.json`:
  `tokenReqAsterisk:true`, `blockedOnEmptyToken:true`, `payloads:0`). The Storybook wizard reads
  the **raw `dsconfig.json`** directly, so it honoured `requiredWhen: "true"` for the UI too —
  the **visible** wizard behaviour was already at parity.
- **URL is intentionally NOT in General.** It is optional: the backend defaults an empty URL to
  `https://gitlab.com/api/v4` (`settings.go:157-164`, `ApplyDefaults`), so it is neither
  `required` nor in the auth group and correctly does not fold into General.

**So what does the fix change?** Only the **generated contract** (see next section) — not the
already-correct wizard rendering. The change makes `dsconfig.json` express the token's
unconditional requirement in the canonical way, so downstream consumers of the generated spec
(the datasource API server / wizard validation) see `required: true` instead of nothing.

---

## The required-field fix in detail

**Before (`requiredWhen: "true"`):** for a `target: "secureJsonData"` field the converter emits a
`SecureValueInfo` from **only** `Key`, `Description`, and `Required` (`dsconfig/convert.go:67-73`).
The `requiredWhen` → `x-dsconfig-required-when` mapping lives in `applyConditions`
(`dsconfig/convert.go:203-215`), which is only reached via `fieldToSpecSchema` — a path **never
taken for secure values**. Net effect: `requiredWhen: "true"` on the token was **silently dropped**;
the generated `settings.gen.json` / `schema.gen.json` listed `accessToken` with **no** `required`
marker at all.

**After (`required: true`):** the converter now sets `Required: true` on the secure value, so the
generated artifacts emit:

```json
"secureValues": [
  {
    "key": "accessToken",
    "description": "Provide information to grant access to this data source. …",
    "required": true
  }
]
```

This matches the backend's unconditional requirement (`settings.go:176-181` returns "access token
(secureJsonData.accessToken) can not be blank" when it is empty) and mirrors the established idiom
for an unconditionally-required secret (e.g. ServiceNow's `basicAuthPassword`, which uses
`"required": true` in its `secureValues`).

---

## Field-by-field parity

Legend: ✅ present & matching · ✏️ corrected by this change · 🔒 write-only secret

| Legacy UI field | Control (legacy) | New UI (schema id) | Control (new) | Storage target | Status |
| --- | --- | --- | --- | --- | --- |
| URL | text input | `root_url` | input | `root.url` (optional, default `https://gitlab.com/api/v4`) | ✅ |
| Access token | password (secure) | `secureJsonData_accessToken` | secure input | `secureJsonData.accessToken` (required) | ✅ ✏️ 🔒 |
| Page limit | number | `jsonData_pageLimit` | number | `jsonData.pageLimit` (optional, default 5) | ✅ |

Notes:

- The **Access token** renders as a masked secure input with a show/hide toggle (the renderer draws
  any `target: "secureJsonData"` field that way); it declares `ui.component: "input"`. Both UIs
  collect the same value into `secureJsonData.accessToken`.
- "Name" and "Default" in the legacy DOM are Grafana's standard datasource chrome (instance name +
  "set as default" toggle), present in every datasource editor and intentionally not part of the schema.

### Save-payload storage-target validation

Filling URL + token and clicking **Save & Test** in the wizard logs the datasource payload
(`verify-gitlab-wizard.json`):

```json
{
  "url": "https://gitlab.com/api/v4",
  "jsonData": { "pageLimit": 5 },
  "secureJsonData": { "accessToken": "glpat-secret-token-123" },
  "secureJsonFields": { "accessToken": false }
}
```

The token routes to **`secureJsonData.accessToken`** and the instance URL to **`root.url`** —
exactly the legacy storage shape the backend reads (`settings.go:114-127`).

---

## `fileUpload` evaluation — not applicable to GitLab

The task asked to use the `fileUpload` control **if the legacy UI uses it**. It does not:

- The legacy GitLab editor exposes only URL + a text/secret access token + page limit. It has **no**
  cert/key fields and **no** file pickers — the legacy DOM capture
  (`legacy-expand-ent-grafana-gitlab-datasource.json` / `legacy-expand-gitlab-verify.json`) reports
  `fileInputs:0`, `uploadButtons:[]`.
- The new UI's `fileUpload` component only activates for `ui.fileMapping` (multi-key JSON
  distribution, e.g. a GCP service-account file); nothing in GitLab needs it.

**Decision:** do **not** add `fileUpload` to any GitLab field.

---

## Custom HTTP Headers — not applicable to GitLab

The legacy editor has **no** Custom HTTP Headers section (`hasCustomHeaders:false`,
`addHeaderBtn:false` in both legacy captures). GitLab authenticates purely via the personal access
token, which go-gitlab sends as the `PRIVATE-TOKEN` request header internally (not a user-editable
header). Headers are correctly **not** modeled, and the new UI shows no headers editor
(`hasHeadersEditor:false`). No change.

---

## Conditional fields — none

GitLab has a **single** authentication method (one personal access token) and **no** auth
discriminator, `effects`, or `dependsOn`/`requiredWhen` conditionals. The only remaining field,
`jsonData.pageLimit`, is a plain optional number with a default of 5. There is nothing conditional
to exercise, and nothing regressed by the fix.

---

## Conformance (no change required)

The change only flips `accessToken` from a dropped `requiredWhen` to `required: true`. The shared
conformance suite already supports `required` on secure values (`SecureValueInfo.Required`), so
**no conformance change was needed**. `SchemaSpecHasNoSecureJSON` still passes (the token stays out
of the settings `spec` and remains a `secureValues` entry — `required: true` does not leak the
value), and `SecureValuesMatchLoadSettings` still matches the single `accessToken` key declared in
`SecureJsonDataKeys` (`settings.go:49-53`).

---

## Verification

```
go generate ./registry/grafana-gitlab-datasource/...   # regenerate schema.gen.json / settings.gen.json
go test ./registry/grafana-gitlab-datasource/...       # ok — PASS
```

`TestSchemaConformance` subtests — all **8 PASS**:
`BaseFieldsResolved`, `SchemaRoundTrip`, `SchemaArtifactInSync`, `SchemaSpecHasNoSecureJSON`,
`ConfigSchemaValid`, `JSONDataMatchesStruct`, `JSONDataTypesMatchStruct`,
`SecureValuesMatchLoadSettings`.

The hand-authored `TestLoadConfig`, `TestApplyDefaults`, and `TestValidate` suites also pass
unchanged (they do not reference the token's required flag).

Generated-spec delta: `"required": true` added to the `accessToken` entry in `secureValues`
(both `settings.gen.json` and `schema.gen.json`); nothing else changed.

New-UI captures: `newgen-gitlab-before-tab` / `newgen-gitlab-after-tab` (tab,
`hasHeadersEditor:false`, `urlPresent:true`, sections Connection/Authentication/Additional
Settings); `verify-gitlab-wizard` (wizard General step 1/4 with the required **Access token**;
save-payload routes token → `secureJsonData.accessToken`, URL → `root.url`);
`verify-gitlab-enforce-before` / `verify-gitlab-enforce-after` (token required + Save & Test blocked
on empty token — identical before/after, confirming no UI regression). Legacy captures:
`legacy-expand-ent-grafana-gitlab-datasource` and `legacy-expand-gitlab-verify` (Connection /
Authentication / Additional Settings; `hasCustomHeaders:false`, `addHeaderBtn:false`,
`fileInputs:0`, `uploadButtons:[]`).

---

## Files changed

- [`registry/grafana-gitlab-datasource/dsconfig.json`](dsconfig.json) — changed
  `secureJsonData_accessToken` from `requiredWhen: "true"` to `required: true`.
- [`registry/grafana-gitlab-datasource/schema.gen.json`](schema.gen.json),
  [`registry/grafana-gitlab-datasource/settings.gen.json`](settings.gen.json) — regenerated by
  `go generate` (`accessToken` secure value now `required: true`).

_Unchanged by design:_ `settings.go`, `settings.ts`, `README.md`, `settings.examples.gen.json`,
`conformance_test.go`, `schema/conformance.go`, and everything in `plugin-ui`.

_Nothing reported as out-of-scope / unfixable:_ the only requested fix (required-field / General-step)
was fully achievable through `dsconfig.json` alone. Custom HTTP Headers and `fileUpload` are correctly
n/a for this plugin (legacy has neither), so no `settings.go`/conformance/plugin-ui coordination was needed.
