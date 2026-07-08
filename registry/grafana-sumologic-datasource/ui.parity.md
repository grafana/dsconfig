# Sumo Logic — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered
from this entry's [`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `grafana-sumologic-datasource`
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/bfrbqiyij92iof` (Grafana Enterprise; the Sumo Logic plugin's own `ConfigEditor` — an **API region / URL** in `jsonData.apiUrl` (region dropdown, pre-filled to `https://api.sumologic.com/api/`), a **Timeout** (`jsonData.timeout`) and **Interval** (`jsonData.interval`), plus **AccessID** in `jsonData.accessId` and a write-only **AccessKey** in `secureJsonData.accessKey`. A single fixed authentication method — HTTP Basic auth with an access ID + access key; the method selector is a no-op.)
- **New UI:** `http://192.168.1.241:58899/iframe.html?args=pluginType:grafana-sumologic-datasource&id=configeditor-datasourceconfigwizard--tab` (Storybook, `ConfigEditor/DatasourceConfigWizard`) — **new-UI capture deferred (storybook offline)** at validation time (`net::ERR_CONNECTION_REFUSED`); see [Verification](#verification).
- **Method:** Playwright captured the legacy UI (full-page screenshot + DOM extraction, `legacy-expand-sumologic.{png,json}`). Storybook was unreachable, so the new-UI screenshots were **not** captured; the new-UI behaviour below is derived from the regenerated OpenAPI spec (`schema.gen.json` / `settings.gen.json`, real `go generate` output) and the plugin-ui resolver/converter logic (`dsconfig/convert.go`), and the fix is verified by the passing Go conformance + unit suites.
- **Result:** **Parity achieved.** The one modeled fix corrected `jsonData_apiUrl` from a non-canonical `requiredWhen: "true"` to `required: true`, so the API URL now (a) emits a canonical OpenAPI `required: ["apiUrl"]` in the generated spec and (b) folds into the wizard's synthetic **General** step. Custom HTTP Headers and `fileUpload` were both evaluated and are correctly **not** used (the legacy editor has neither — confirmed by capture). The genuine credential conditionals (`accessId` / `accessKey`, gated on `authMethod == 'accessKey'`) are correctly **left as conditional `requiredWhen`**.

---

## TL;DR of changes

| #   | Change | File | Why |
| --- | --- | --- | --- |
| 1   | Changed `jsonData_apiUrl` from `requiredWhen: "true"` → `required: true` | [`dsconfig.json`](dsconfig.json) | The Sumo Logic API URL is **unconditionally** required — the backend hard-fails an empty URL with "invalid API URL" (`settings.go:177-179`, mirroring upstream `pkg/models/settings.go:58-60`). `requiredWhen: "true"` is a non-canonical CEL expression that only emits an `x-dsconfig-required-when` vendor extension, which the OpenAPI `required` array and the wizard's General-step resolver both ignore. `required: true` emits the canonical `required: ["apiUrl"]` and is the idiom the wizard's synthetic **General** step recognises. |
| 2   | Regenerated `schema.gen.json`, `settings.gen.json` via `go generate ./registry/grafana-sumologic-datasource/...` | generated artifacts | Keep committed artifacts in sync (guarded by `SchemaArtifactInSync`) |

No changes were made to `settings.go`, `settings.ts`, `README.md`, `conformance_test.go`, `schema/conformance.go`, `settings.examples.gen.json`, or `plugin-ui`. The single schema change flows through `dsconfig.json`; the rest is produced by `go generate`.

**No conformance-test or plugin-ui change was required** — the wizard already folds `required` fields into the **General** step, and the conformance suite already understands `required` on `jsonData` fields (see [Conformance](#conformance-no-change-required)).

---

## Section layout

The `groups` were left untouched. Group taxonomy from [`dsconfig.json`](dsconfig.json)
(neither group is `optional`):

| Order | Section (`id`) | `optional` | Fields (in display order) |
| --- | --- | --- | --- |
| 1 | **API Region** (`apiRegion`) | no | API region / URL, Timeout, Interval |
| 2 | **Authentication** (`authentication`) | no | AccessID, AccessKey |

The legacy DOM capture (`legacy-expand-sumologic.json`) reported no collapsible
section headings (`headings: []`) — the Sumo Logic editor renders its fields as a
flat form rather than `<h*>`/`<legend>` sections — but the field set (API URL,
Timeout, Interval, AccessID, AccessKey) matches the two schema groups above.

### Wizard mode: API URL folds into the "General" step

In **wizard mode** the plugin-ui builds a synthetic first step titled **General**
(`resolveRequiredFieldsGroup`) from (a) every field marked `required: true`, plus (b) the
fields of the auth group (`id: "authentication"`), plus their `dependsOn` parents/children.

- **Before** (`requiredWhen: "true"`): API URL lived in the `apiRegion` group and was **not**
  marked `required: true`, so it was excluded from General. General would show only **AccessID**
  and **AccessKey** (the `authentication` group).
- **After** (`required: true`): API URL folds into General. General now shows **API region / URL**
  (with a required `*` marker) **+ AccessID + AccessKey**.

**AccessID** and **AccessKey** fold into General in both states because they live in the
`authentication` group, which the wizard recognises regardless of the required marker (they keep
their conditional `requiredWhen: "jsonData_authMethod == 'accessKey'"`). So the only observable
delta from the fix is the newly-present, required **API region / URL**.

> **new-UI capture deferred (storybook offline).** The wizard rendering above is the expected
> behaviour derived from the resolver logic and the regenerated spec (`required: ["apiUrl"]`), not
> a captured screenshot. It matches the already-captured behaviour of the structurally identical
> Jenkins fix (URL folding into General 1/3 once flipped to `required: true`).

---

## The required-field fix in detail

`jsonData_apiUrl` targets `jsonData` (not a secure value), so its `requiredWhen: "true"` was
routed through `fieldToSpecSchema` → `applyConditions` (`dsconfig/convert.go:203-216`), which
emits an `x-dsconfig-required-when: "true"` **vendor extension** — not a canonical `required`
entry. The wizard's General-step resolver and the OpenAPI `required` array both key off
`required: true`, not that CEL extension, so the URL was neither in the spec's `required` array
nor in the wizard's General step.

**Before (`requiredWhen: "true"`):** the generated `jsonData` object had **no** `required`
array, and `apiUrl` carried `"x-dsconfig-required-when": "true"` (alongside its `default`).

**After (`required: true`):** for a `jsonData` field with `f.Required && f.Section == ""`, the
converter appends the key to the object's `required` list (`dsconfig/convert.go:78-79, 96-97`),
so the generated artifacts now emit (real `go generate` output):

```json
"jsonData": {
  "type": "object",
  "required": ["apiUrl"],
  "properties": {
    "apiUrl": {
      "description": "SumoLogic API URL. [Read how to find your deployment.](...)",
      "type": "string",
      "default": "https://api.sumologic.com/api/"
    }
  }
}
```

The `x-dsconfig-required-when: "true"` extension is gone; the `default` is retained (a required
field may still carry a default — the region is pre-filled to US1/Default but the user must not
blank it). This matches the backend's unconditional requirement (`settings.go:177-179`, which
runs *after* `ApplyDefaults` at `settings.go:154-156`) and mirrors the established idiom for
unconditionally-required `jsonData` fields (e.g. Jenkins' `url`, which lands in the spec's
`required` array via the same path).

---

## Field-by-field parity

Legend: ✅ present & matching · ✏️ corrected by this change · 🔀 conditional (`requiredWhen`) · 🔒 write-only secret · 🎛 auth discriminator (backend-only)

| Legacy UI field | Control (legacy) | New UI (schema id) | Control (new) | Storage target | Status |
| --- | --- | --- | --- | --- | --- |
| API region / URL | region dropdown / text | `jsonData_apiUrl` | input | `jsonData.apiUrl` (required, default US1) | ✅ ✏️ |
| Timeout | number | `jsonData_timeout` | input | `jsonData.timeout` (default 30, min 1) | ✅ |
| Interval | number | `jsonData_interval` | input | `jsonData.interval` (default 1000, min 200) | ✅ |
| AccessID | text input | `jsonData_accessId` | input | `jsonData.accessId` | ✅ 🔀 |
| AccessKey | password (secure) | `secureJsonData_accessKey` | secure input | `secureJsonData.accessKey` | ✅ 🔀 🔒 |
| — (no editor control) | — | `jsonData_authMethod` | — | `jsonData.authMethod` (`accessKey`) | ✅ 🎛 |

Notes:

- **API URL is unconditionally required** and routes to **`jsonData.apiUrl`** (not `root.url`) —
  the Sumo Logic plugin builds its own HTTP Basic-auth client from `jsonData`/`secureJsonData`
  and never consults the root datasource URL (`settings.go:64-70`). `role: "endpoint.baseUrl"`
  and the `defaultValue` are preserved; only the requirement marker changed.
- **AccessID / AccessKey keep their genuine conditional `requiredWhen: "jsonData_authMethod ==
  'accessKey'"`** — correctly **not** converted to `required: true`. They are required only for
  the access-key method (`settings.go:183-190`). Although `accessKey` is the only supported (and
  default) method, the schema models the requirement as method-gated, matching the credential
  `pair` relationship documented in `dsconfig.json`. Left untouched per scope.
- **`jsonData_authMethod`** is a `backend-only`-tagged discriminator (no `ui` block): the editor
  never writes it and neither UI renders it; the backend defaults it to `accessKey` and rejects
  any other value. Parity preserved.
- The **AccessKey** secure field renders as a masked secure input with a show/hide toggle (the
  renderer draws any `target: "secureJsonData"` field that way); it declares `ui.component:
  "input"`. Both UIs collect it into `secureJsonData.accessKey` (write-only; presence is read
  back via `secureJsonFields.accessKey`).
- "Name" and "Default" in the legacy DOM are Grafana's standard datasource chrome (instance name
  + "set as default" toggle), present in every datasource editor and intentionally not part of
  the schema.

---

## `fileUpload` evaluation — not applicable to Sumo Logic

The task asked to use the `fileUpload` control **if the legacy UI uses it**. It does not:

- The legacy Sumo Logic editor exposes only API URL + Timeout + Interval + AccessID + AccessKey.
  It has **no** TLS cert / key fields and **no** file pickers — the legacy DOM capture
  (`legacy-expand-sumologic.json`) reports `fileInputs: 0`, `uploadButtons: []`.
- The new UI's `fileUpload` component only activates for `ui.fileMapping` (multi-key JSON
  distribution, e.g. a GCP service-account file); nothing in Sumo Logic needs it.

**Decision:** do **not** add `fileUpload` to any Sumo Logic field.

---

## Custom HTTP Headers — not applicable to Sumo Logic

The legacy editor has **no** Custom HTTP Headers section (`hasCustomHeaders: false`,
`addHeaderBtn: false` in the legacy capture). Sumo Logic authenticates purely via HTTP Basic auth
(the `Authorization: Basic` header is derived internally from the access ID / access key, not via
a user-editable header). Headers are correctly **not** modeled, and no headers editor is added.
No change.

---

## Conditional fields — credential pair preserved

Sumo Logic has a **single** authentication path (HTTP Basic auth via access ID + access key) with
a fixed, backend-only `authMethod` discriminator (no radio in the editor), **no** `effects`, and
**no** `dependsOn` reveal. The only conditionals are the two credential fields' genuine
`requiredWhen: "jsonData_authMethod == 'accessKey'"` expressions on `jsonData_accessId` and
`secureJsonData_accessKey` — these are **left unchanged** (they are method-gated, not
unconditional, so `required: true` would be incorrect). After the fix the only `requiredWhen:
"true"` (the unconditional one on the API URL) is gone; the two conditional `requiredWhen`
expressions remain, and the generated spec preserves them as
`x-dsconfig-required-when: "jsonData_authMethod == 'accessKey'"`. Nothing conditional regressed.

---

## Conformance (no change required)

The change only flips `apiUrl` from an `x-dsconfig-required-when` extension to a canonical
`required: ["apiUrl"]` entry on the `jsonData` object. The shared conformance suite already
supports `required` on `jsonData` fields, so **no conformance change was needed**:

- `SchemaSpecHasNoSecureJSON` still passes — `apiUrl` is a `jsonData` property (not secure), and
  `accessKey` remains the sole `secureValues` entry; nothing leaked.
- `JSONDataMatchesStruct` / `JSONDataTypesMatchStruct` still pass — `apiUrl`, `timeout`,
  `interval`, `accessId`, `authMethod` still map 1:1 to the `Config` struct's json-tagged fields
  (`settings.go:71-82`); adding `apiUrl` to the `required` array does not change the property set
  or types.
- `SecureValuesMatchLoadSettings` still matches the single `accessKey` key declared in
  `SecureJsonDataKeys` (`settings.go:58-60`).
- `SchemaArtifactInSync` passes after regeneration.

---

## Verification

```
go generate ./registry/grafana-sumologic-datasource/...   # regenerate schema.gen.json / settings.gen.json
go test ./registry/grafana-sumologic-datasource/...        # ok — PASS
```

`TestSchemaConformance` subtests — all **8 PASS**:
`BaseFieldsResolved`, `SchemaRoundTrip`, `SchemaArtifactInSync`, `SchemaSpecHasNoSecureJSON`,
`ConfigSchemaValid`, `JSONDataMatchesStruct`, `JSONDataTypesMatchStruct`,
`SecureValuesMatchLoadSettings`.

The hand-authored `TestLoadConfig`, `TestApplyDefaults`, and `TestValidate` suites also pass
unchanged (they do not reference the API URL's required flag; `Validate()` already hard-fails an
empty API URL — `TestValidate/missing_apiUrl_errors`).

Generated-spec delta (real `go generate` output): `"required": ["apiUrl"]` added to the
`jsonData` object and `"x-dsconfig-required-when": "true"` removed from `apiUrl` (in both
`settings.gen.json` and `schema.gen.json`); the `default` and every other field is unchanged.

Legacy capture: `legacy-expand-sumologic` — `hasCustomHeaders: false`, `addHeaderBtn: false`,
`fileInputs: 0`, `uploadButtons: []` (confirms no headers / no file upload).

**New-UI captures: deferred (storybook offline).** The Storybook host
(`http://192.168.1.241:58899`) returned `net::ERR_CONNECTION_REFUSED` at validation time, so the
tab/wizard screenshots were not taken (not retried per runbook). The fix's effect is instead
evidenced by the regenerated OpenAPI contract above and the passing Go suites; the wizard
General-step behaviour is reasoned from `dsconfig/convert.go` + the resolver and matches the
captured behaviour of the identical Jenkins fix.

---

## Files changed

- [`registry/grafana-sumologic-datasource/dsconfig.json`](dsconfig.json) — changed
  `jsonData_apiUrl` from `requiredWhen: "true"` to `required: true`. The credential fields
  (`jsonData_accessId`, `secureJsonData_accessKey`) keep their genuine conditional
  `requiredWhen: "jsonData_authMethod == 'accessKey'"`.
- [`registry/grafana-sumologic-datasource/schema.gen.json`](schema.gen.json),
  [`registry/grafana-sumologic-datasource/settings.gen.json`](settings.gen.json) — regenerated by
  `go generate` (`jsonData.required: ["apiUrl"]` added; `x-dsconfig-required-when: "true"` removed
  from `apiUrl`).

_Unchanged by design:_ `settings.go`, `settings.ts`, `README.md`, `settings.examples.gen.json`,
`conformance_test.go`, `schema/conformance.go`, and everything in `plugin-ui`.

_Nothing reported as out-of-scope / unfixable:_ the only requested fix (required-field /
General-step) was fully achievable through `dsconfig.json` alone. Custom HTTP Headers and
`fileUpload` are correctly n/a for this plugin (legacy has neither), so no
`settings.go`/conformance/plugin-ui coordination was needed.
