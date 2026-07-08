# Looker — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered from
this entry's [`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `grafana-looker-datasource`
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/afrbqixktmeiof` (Grafana Enterprise 13.0.1; the Looker plugin's own `ConfigEditor` — **Looker URL** + **Looker Client ID** + a write-only **Looker Client Secret**. The auth-type selector is hidden because `client_secret` is the only method. No `DataSourceHttpSettings`, no Custom HTTP Headers, no TLS/cert fields.)
- **New UI:** `http://192.168.1.241:58899/iframe.html?args=pluginType:grafana-looker-datasource&id=configeditor-datasourceconfigwizard--tab` (Storybook, `ConfigEditor/DatasourceConfigWizard`), also the `--wizard` story.
- **Method:** Playwright captured both UIs (full-page screenshots + DOM extraction). The Storybook story fetches the schema from `raw.githubusercontent.com/.../registry/grafana-looker-datasource/dsconfig.json`; the local (edited) `dsconfig.json` was served to the new UI by intercepting that request with `context.route(...)`, so the captures reflect the local schema without pushing.
- **Result:** **Parity achieved.** One modeled fix: the unconditionally-required **Looker URL** (`jsonData_baseUrl`) was corrected from the non-canonical `requiredWhen: "true"` → `required: true`. This folds it into the wizard's synthetic **General** step (empirically it was *absent* from General before, *present* after) and emits `jsonData.required: ["base_url"]` in the generated spec. **Custom HTTP Headers** and **`fileUpload`** were evaluated and correctly **not** used — the legacy editor has neither (`hasCustomHeaders:false`, `addHeaderBtn:false`, `fileInputs:0`, `uploadButtons:[]`). The two conditional credential fields (`client_id`, `client_secret`) keep their auth-gated `requiredWhen` and were **left unchanged**.

---

## TL;DR of changes

| #   | Change | File | Why |
| --- | --- | --- | --- |
| 1   | Changed `jsonData_baseUrl` from `requiredWhen: "true"` → `required: true` | [`dsconfig.json`](dsconfig.json) | The base URL is **unconditionally** required — the backend `Validate()` rejects an empty value with `"invalid/empty Looker base url"` (per the schema's own `instructions`, citing `pkg/models/config.go:28-45`). `requiredWhen: "true"` is a non-canonical always-true CEL string that the wizard's `resolveRequiredFieldsGroup` does **not** inspect, so the field never folded into the synthetic **General** step. `required: true` is the canonical idiom: it emits `jsonData.required: ["base_url"]` in the generated OpenAPI spec and puts the URL on the General step. |
| 2   | Regenerated `schema.gen.json`, `settings.gen.json` via `go generate ./registry/grafana-looker-datasource/...` | generated artifacts | Keep committed artifacts in sync (guarded by `SchemaArtifactInSync`). |

No changes were made to `settings.go`, `settings.ts`, `README.md`, `settings.examples.gen.json`, `conformance_test.go`, `schema/conformance.go`, or `plugin-ui`. All schema changes flow through `dsconfig.json`; the rest is produced by `go generate`. **No conformance-test or plugin-ui change was required.**

---

## Section layout

The `groups` taxonomy was left unchanged; only the `base_url` required flag was corrected. Verified
rendering top-to-bottom in the new UI (tab mode, `newgen-looker-tab.json` + screenshot):

| Order | Section (`id`) | `optional` | Fields (in display order) |
| --- | --- | --- | --- |
| 1 | **Connection** (`connection`) | no | Looker URL |
| 2 | **Authentication** (`authentication`) | no | Authentication type, Looker Client ID, Looker Client Secret |

The legacy editor renders the same fields as a flat list (**Looker URL**, **Looker Client ID**,
**Looker Client Secret**) with no section headers (`legacy-expand-looker-verify.json`:
`headings:[]`) and with the single-option auth-type selector hidden. New-UI tab capture:
`hasHeadersEditor:false`, `urlPresent:true`, sections `["Connection","Authentication"]`, and the
header **"Fields marked with \* are required"** with a required asterisk on **Looker URL\***,
**Looker Client ID\***, and **Looker Client Secret\***.

### Wizard mode: the General-step fix (before / after)

In **wizard mode** the plugin-ui builds a synthetic first step titled **General**
(`resolveRequiredFieldsGroup`) from (a) every field marked `required: true`, plus (b) the fields of
the auth group (`id: "authentication"`) and their `dependsOn` relatives. `requiredWhen: "true"` is a
CEL string the resolver does **not** evaluate, so a URL flagged that way was *not* pulled into General.

This is the one behaviour the fix changes, and it was captured empirically (General step is **1/3**
in both cases):

| Wizard General (step 1/3) | Before (`requiredWhen: "true"`) | After (`required: true`) |
| --- | --- | --- |
| Looker URL field on General | **absent** — `urlPresent:false` (`newgen-looker-wizard-before.json`) | **present, Looker URL\*** — `urlPresent:true` (`newgen-looker-wizard.json`) |
| Authentication type | present (folded via auth group) | present (folded via auth group) |
| Looker Client ID\* | present | present |
| Looker Client Secret\* | present | present |
| `hasHeadersEditor` | false | false |

The `authType` / `client_id` / `client_secret` fields fold into General regardless (they live in the
`authentication` group); only **Looker URL** was missing before and is now correctly included.

---

## The required-field fix in detail

**Before (`requiredWhen: "true"`):** for a `target: "jsonData"` field this maps to the
`x-dsconfig-required-when: "true"` extension — a literal always-true condition rather than a first-class
required marker. The generated `jsonData` object had **no** `required` array, and the wizard's
General-step resolver ignored the field.

**After (`required: true`):** the converter now lists the key in the settings `spec`'s `required`
array. Generated delta (`schema.gen.json` / `settings.gen.json`):

```json
"jsonData": {
  "type": "object",
  "required": [
    "base_url"
  ],
  "properties": {
    "base_url": {
      "description": "Looker base URL. Example: https://...looker.app",
      "type": "string"
    }
  }
}
```

The `x-dsconfig-required-when: "true"` line on `base_url` is removed. This matches the backend's
unconditional requirement (empty base URL → `"invalid/empty Looker base url"`) and mirrors the
established idiom for an always-required endpoint field (e.g. Databricks' `host` / `httpPath`, which
use `required: true` and appear in `jsonData.required`).

---

## Field-by-field parity

Legend: ✅ present & matching · ✏️ corrected by this change · 🔀 discriminator-gated · 🔒 write-only secret

| Legacy UI field | Control (legacy) | New UI (schema id) | Control (new) | Storage target | Status |
| --- | --- | --- | --- | --- | --- |
| Looker URL | text input | `jsonData_baseUrl` | input (placeholder `https://xxxxx.looker.app`) | `jsonData.base_url` (**required**) | ✅ ✏️ |
| _(hidden — single option)_ | n/a | `jsonData_authType` | radio (one option: Client Secret, default `client_secret`) | `jsonData.auth_type` | ✅ 🔀 |
| Looker Client ID | text input | `jsonData_clientId` | input | `jsonData.client_id` (required when `auth_type == 'client_secret'`) | ✅ 🔀 |
| Looker Client Secret | password (secure) | `secureJsonData_clientSecret` | secure input | `secureJsonData.client_secret` (required when `auth_type == 'client_secret'`) | ✅ 🔀 🔒 |

Notes:

- **Looker Client Secret** renders as a masked secure input with a show/hide toggle (the renderer
  draws any `target: "secureJsonData"` field that way); it is write-only (read back via
  `secureJsonFields.client_secret`).
- **Authentication type** is a single-option discriminator (`auth.discriminator`). The legacy editor
  **hides** it because `client_secret` is the only method; the new UI shows it as a one-option radio.
  The backend defaults a missing/empty `auth_type` to `client_secret` (`ApplyDefaults`,
  `pkg/models/config.go:47-50`), so both UIs converge on the same stored value. This visible-vs-hidden
  difference is a pre-existing modeling choice, **out of scope** for the required-field fix (see below).
- "Name" and "Default" in the legacy DOM are Grafana's standard datasource chrome (instance name +
  "set as default" toggle), present in every editor and intentionally not part of the schema.

---

## Conditional fields — left unchanged

`client_id` and `secure client_secret` are gated on the auth discriminator and keep their conditional
markers (unchanged by this task):

- `jsonData_clientId`: `dependsOn: "jsonData_authType == 'client_secret' || jsonData_authType == ''"`,
  `requiredWhen: "jsonData_authType == 'client_secret'"` → emitted as `x-dsconfig-required-when` in
  `schema.gen.json`.
- `secureJsonData_clientSecret`: same `dependsOn` / `requiredWhen`. (The converter emits secure values
  from only key/description/`required`, so a *conditional* `requiredWhen` on a secure field is not
  represented as an extension in the generated spec — a **pre-existing converter behaviour**, not a
  regression from this change, and out of scope here.)

Because `auth_type` has exactly one value today (`client_secret`, also the default), these conditions
are effectively always-satisfied and both fields render/require in practice — confirmed by the
required asterisks on **Looker Client ID\*** and **Looker Client Secret\*** in both tab and wizard
captures. The task instruction to **leave conditional `requiredWhen`** was followed.

---

## `fileUpload` evaluation — not applicable to Looker

The task asked to use the `fileUpload` control **only if the legacy UI uses it**. It does not:

- The legacy Looker editor exposes only URL + Client ID + Client Secret. It has **no** cert/key fields
  and **no** file pickers — the legacy DOM capture (`legacy-expand-looker-verify.json`) reports
  `fileInputs:0`, `uploadButtons:[]`.
- The new UI's `fileUpload` component only activates for `ui.fileMapping` (multi-key JSON distribution,
  e.g. a GCP service-account file); nothing in Looker needs it.

**Decision:** do **not** add `fileUpload` to any Looker field.

---

## Custom HTTP Headers — not applicable to Looker

The legacy editor has **no** Custom HTTP Headers section (`hasCustomHeaders:false`,
`addHeaderBtn:false`). Looker authenticates via API3 client-credentials through the Looker Go SDK's own
`rtl.AuthSession`; Grafana's standard datasource HTTP options are not even applied to the SDK client
(per the schema's `instructions`, citing `pkg/models/config.go:70-71` and `pkg/looker/client.go:22-38`),
so there is no user-managed header editor. Headers are correctly **not** modeled, and the new UI shows
no headers editor in either mode (`hasHeadersEditor:false`). No change.

---

## Verification

```
go generate ./registry/grafana-looker-datasource/...   # regenerate schema.gen.json / settings.gen.json
go test ./registry/grafana-looker-datasource/...        # ok — PASS
```

`TestSchemaConformance` — **PASS** (`ok github.com/grafana/dsconfig/registry/grafana-looker-datasource`).
The shared suite covers schema round-trip, artifact drift (`SchemaArtifactInSync`), spec/secure
separation (`SchemaSpecHasNoSecureJSON`), `jsonData`↔struct parity both directions, and secure-key
parity — all green. The hand-authored `settings_test.go` suites also pass unchanged (they assert the
backend rejects empty `base_url` / `client_id` / `client_secret`, matching the `required: true`
conversion).

Generated-spec delta: `jsonData.required: ["base_url"]` added and `x-dsconfig-required-when: "true"`
removed from `base_url` (in both `settings.gen.json` and `schema.gen.json`); nothing else changed.

New-UI captures: `newgen-looker-tab` (tab — sections Connection / Authentication,
`hasHeadersEditor:false`, `urlPresent:true`, Looker URL\* / Client ID\* / Client Secret\* required),
`newgen-looker-wizard` (wizard General **1/3** with the required **Looker URL\*** + auth fields;
`urlPresent:true`), and `newgen-looker-wizard-before` (the pre-fix General **1/3** with **no** Looker
URL; `urlPresent:false`) — the before/after that proves the General-step fix. Legacy capture:
`legacy-expand-looker-verify` (`hasCustomHeaders:false`, `addHeaderBtn:false`, `fileInputs:0`,
`uploadButtons:[]`).

---

## Files changed

- [`registry/grafana-looker-datasource/dsconfig.json`](dsconfig.json) — changed `jsonData_baseUrl`
  from `requiredWhen: "true"` to `required: true`. The two conditional `requiredWhen` fields
  (`client_id`, `client_secret`) were left in place.
- [`registry/grafana-looker-datasource/schema.gen.json`](schema.gen.json),
  [`registry/grafana-looker-datasource/settings.gen.json`](settings.gen.json) — regenerated by
  `go generate` (`jsonData.required: ["base_url"]` added; `x-dsconfig-required-when: "true"` removed
  from `base_url`).

_Unchanged by design:_ `settings.go`, `settings.ts`, `README.md`, `settings.examples.gen.json`,
`conformance_test.go`, `schema/conformance.go`, and everything in `plugin-ui`.

_Nothing required a change outside `dsconfig.json` (+ generated `.gen.json`)._ The only requested fix
(required-field / General-step) was fully achievable through `dsconfig.json` alone. Custom HTTP Headers
and `fileUpload` are correctly n/a for this plugin (legacy has neither). One **out-of-scope observation**
(not fixed): the new UI shows the single-option **Authentication type** radio that the legacy editor
hides, and the conditional `requiredWhen` on the secure `client_secret` is not emitted into the
generated spec (a pre-existing converter behaviour) — neither is a required-field/General-step issue and
neither needs a conformance or plugin-ui change.
