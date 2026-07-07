# Sqlyze (ODBC) — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered
from this entry's [`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `grafana-odbc-datasource` (`pluginName` **Sqlyze Datasource**; an **ODBC / SQL-family** datasource built on `github.com/grafana/sqlds/v5`, but with its **own custom React config editor** — not `@grafana/sql` — using `@grafana/plugin-ui` `ConfigSection`/`DataSourceDescription` + `@grafana/ui` `LegacyForms.FormField`/`IconButton`)
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/efrbqixwr4xz4e`. **The legacy plugin failed to start in this environment** — `GET /api/plugins/grafana-odbc-datasource/settings` returns `{"message":"Plugin failed to start"}` and the edit page renders **"Data source not found"** with a *"Plugin failed to load"* banner. The datasource entry itself resolves (`GET /api/datasources/uid/efrbqixwr4xz4e` → `type: grafana-odbc-datasource`, id 143), so the UID is correct; only the backend binary won't run here (expected for a native-ODBC plugin outside a driver-provisioned host). **The live legacy editor could therefore not be captured.** The legacy field inventory below is taken from this entry's committed [`README.md`](README.md) source-code analysis of `ConfigEditor.tsx` / `selectors.ts` / `types.ts`, cross-checked against the datasource-API entry.
- **New UI:** `http://192.168.1.241:58899/iframe.html?id=configeditor-datasourceconfigwizard--tab&args=pluginType:grafana-odbc-datasource&globals=theme:light&viewMode=story` (Storybook, `ConfigEditor/DatasourceConfigWizard`; also `--wizard`).
- **Method:** Playwright captured the new UI (full-page screenshots + DOM extraction of section buttons / field labels / headers-editor presence) in both **tab** and **wizard** modes. The Storybook story fetches the schema from `…/registry/grafana-odbc-datasource/dsconfig.json`; the local (edited) `dsconfig.json` was served to the new UI by intercepting that request with `context.route(...)`, so the screenshots reflect the local schema without pushing. The legacy edit page was also visited (`capture-legacy-expand.js efrbqixwr4xz4e`) but returned the plugin-failed-to-start error page (see above).
- **Result:** **Parity achieved for the modeled field set.** All editor-visible fields (Driver, Timeout, Driver Settings with its name/value/secure sub-fields) are present in the new UI and route to identical storage targets. Sqlyze speaks ODBC over a native driver, **not HTTP**, so it correctly has **no HTTP-headers editor** and **no file-upload** control. The schema has **no conditional fields** (`dependsOn`) and **no `effects`**. The one change required was making the unconditionally-required **Driver** field use `required: true` so the wizard's synthetic **General** step pulls it in.

---

## TL;DR of changes

| #   | Change | File | Why |
| --- | --- | --- | --- |
| 1   | Changed **`jsonData_driver`** from `"requiredWhen": "true"` → `"required": true` | [`dsconfig.json`](dsconfig.json) | The driver is unconditionally required — `(Config).Validate` rejects an empty driver (`settings.go:164-168`, `"driver is required"`), and the upstream backend rejects it via `CheckDriverFileExists` before building the connection string (`pkg/models/settings.go:68-71`). The wizard's synthetic **General** step only pulls fields with `required: true`; `requiredWhen` is a CEL expression the required-fields resolver does not inspect. Also emits a proper OpenAPI `required: ["driver"]` array instead of the `x-dsconfig-required-when` extension. |
| 2   | Regenerated `schema.gen.json`, `settings.gen.json` via `go generate ./registry/grafana-odbc-datasource/...` | generated artifacts | Keep committed artifacts in sync (guarded by `SchemaArtifactInSync`). |

`jsonData_driver` had **no** existing `required` key and its `requiredWhen` value was the literal
string `"true"` (an unconditional marker), so it was converted rather than deleted. It is the
**only** `requiredWhen` in the schema — there are **no conditional `requiredWhen` values to leave
untouched**.

No changes were made to `settings.go`, `settings.ts`, `README.md`, `settings.examples.gen.json`,
`schema/conformance.go`, or `plugin-ui`. No `conformance.go` change was needed (this plugin models
no `indexedPair` field). No `plugin-ui` change was needed.

---

## Section layout

The schema declares a **single group**:

| Order | Section (`id`) | `optional` | Fields (in display order) |
| --- | --- | --- | --- |
| 1 | **Connection Settings** (`connection`) | no | Driver, Timeout (seconds), Driver Settings |

The new UI (tab mode) renders exactly this: a **Connection Settings** accordion containing
**Driver\***, **Timeout (seconds)** (default `10`), and **Driver Settings** (an `Add driver settings`
key-value list), with the header *"Fields marked with \* are required"*. This matches the legacy
editor's structure per `ConfigEditor.tsx` (a single `ConfigSection` with the Driver and Timeout
`FormField`s and an `<h5>Driver Settings</h5>` repeatable list).

Two modeled fields are intentionally **not** in any group and therefore render in **neither** UI:

- **`jsonData_dsn`** (`DSN`) — `backend-only`, has no editor UI (provisioning-only; `README.md` §Backend-only settings).
- **`secureJsonData_pwd`** (`pwd`) — a **representative** dynamic secret. Real secrets are collected inline via a Driver-Setting's *secure* toggle and stored in `secureJsonData` under a key equal to the setting's `name`; there is no fixed `pwd` field in the editor.

### Wizard mode: the "General" step

In **wizard mode** the plugin-ui builds a synthetic first step titled **General**
(`resolveRequiredFieldsGroup`) from every field marked `required: true` (plus any auth-group members
and their `dependsOn` parents/children — of which this schema has none).

**Effect of the `required: true` fix (both states captured):**

| State | Wizard steps | General step? | Driver location |
| --- | --- | --- | --- |
| **Before** (`requiredWhen: "true"`) | **1** — `Connection Settings 1/1` | **absent** (no `required:true` field, no auth group ⇒ no synthetic step) | only reachable inside the `Connection Settings` step |
| **After** (`required: true`) | **2** — `General 1/2` → `Connection Settings 2/2` | **present** ✅ | **Driver\*** now leads the **General** step |

Captured directly:

- `newgen-odbc-wiz-before.*`: step header **`Connection Settings 1/1`**, step dropdown shows only `Connection…` — the whole group is the only step; there is **no** General step.
- `newgen-odbc-wiz-after.*`: step header **`General 1/2`**, step dropdown shows `General`; the step contains a single field **`Driver *`** (placeholder `DSN or path to ODBC Driver`).

Note the Driver field shows an inline `*` required-marker in **both** states — the renderer evaluates
`requiredWhen`/`required` for the asterisk — but only `required: true` promotes it into the synthetic
**General** wizard step. Tab mode is unaffected (the synthetic `_required` group is filtered out
there); it shows the single `Connection Settings` section in both states.

---

## Field-by-field parity

Legend: ✅ present & matching · ➖ intentionally not rendered (backend-only / representative)

| Legacy UI field | Control (legacy) | New UI (schema id) | Control (new) | Storage target | Status |
| --- | --- | --- | --- | --- | --- |
| Driver \* | text input (`FormField`, `ConfigEditor.tsx:176`) | `jsonData_driver` | input | `jsonData.driver` | ✅ |
| Timeout (seconds) | text input (`ConfigEditor.tsx:191`, default `10`) | `jsonData_timeout` | input | `jsonData.timeout` | ✅ |
| Driver Settings | repeatable `{name,value,secure}` list (`<h5>` + `FormField`s + lock `IconButton`) | `jsonData_settings` | keyvalue (`Add driver settings`) | `jsonData.settings[]` | ✅ |
| &nbsp;&nbsp;• Name | text input (`ConfigEditor.tsx:218`) | `jsonData_settings.item.name` | input | `settings[].name` | ✅ |
| &nbsp;&nbsp;• Value | text input (`ConfigEditor.tsx:227`) | `jsonData_settings.item.value` | input | `settings[].value` | ✅ |
| &nbsp;&nbsp;• Secure | lock toggle `IconButton` (`ConfigEditor.tsx:237-246`) | `jsonData_settings.item.secure` | switch | `settings[].secure` | ✅ |
| (secure setting value → dynamic secret) | inline password value (routed by `name`) | `secureJsonData_pwd` (representative) | — | `secureJsonData[<name>]` | ➖ |
| DSN | — (no editor UI; provisioning only) | `jsonData_dsn` | — | `jsonData.DSN` | ➖ |

All editor-visible fields render in the new UI's single **Connection Settings** section
(`newgen-odbc-tab.*`: `Driver *`, `Timeout (seconds)` = `10`, `Driver Settings` + `Add driver
settings`). The Grafana editor chrome at the top (datasource **Name**, **Default** toggle) is not
part of the datasource config and is correctly **not** modeled. `DSN` and the representative `pwd`
secret are deliberately ungrouped and render in neither UI (matching legacy: DSN has no UI, secrets
are inline in the settings list).

---

## Gaps found

**None beyond the required-field fix.** The complete legacy editor field set (`driver`, `timeout`,
`settings` with `name`/`value`/`secure`) is already modeled in `dsconfig.json`, so — unlike graphite
(which was missing Custom HTTP Headers) — no field had to be added. Sqlyze keeps all fields **inline**
in one group (no `packs`), matching the single-`ConfigSection` legacy editor.

### Custom HTTP Headers — not applicable (verified)

Sqlyze connects through a native ODBC driver over the ODBC protocol, **not HTTP**, so it has no
HTTP-headers concept. New UI (`newgen-odbc-tab.*`, `newgen-odbc-wiz-after.*`):
`hasHeadersEditor: false` in both tab and wizard modes. The legacy `ConfigEditor.tsx` has no headers
control (`README.md` confirms only `driver`/`timeout`/`settings` on top of base jsonData). Correctly
**not** added.

---

## `fileUpload` evaluation — not applicable to Sqlyze

The task asks to use the `fileUpload` control **if the legacy UI uses it**. It does not:

- The legacy editor's **Driver** field is a plain text input taking either a driver alias in braces
  (`{MyDSN}`) or an **absolute path** to the ODBC driver shared library typed by hand — not an
  uploaded file. There is no `<input type="file">` and no upload button in `ConfigEditor.tsx`.
- There are no certificate/PEM fields at all (no TLS section).
- The new UI's `fileUpload` component only activates when a field declares `ui.fileMapping`; no field here does.

**Decision:** do **not** add `fileUpload` to any Sqlyze field.

---

## Conditional fields & effects — none

The Sqlyze schema contains **no** `dependsOn`, `disabledWhen`, or `effects` blocks, and (after this
change) **no** `requiredWhen`. There is no auth-method selector — the ODBC driver *is* the connection
method (`instructions[0]`) — so there are no conditional reveals to exercise and none were added.
Every field is unconditionally visible within its section.

---

## Verification

```
go generate ./registry/grafana-odbc-datasource/...   # regenerate schema.gen.json / settings.gen.json
go test ./registry/grafana-odbc-datasource/...        # PASS
```

`go test ./registry/grafana-odbc-datasource/...` → `ok` (all tests PASS). Conformance subtests
(`TestSchemaConformance`): `BaseFieldsResolved`, `SchemaRoundTrip`, `SchemaArtifactInSync`,
`SchemaSpecHasNoSecureJSON`, `ConfigSchemaValid`, `JSONDataMatchesStruct`, `JSONDataTypesMatchStruct`,
`SecureValuesMatchLoadSettings` — all **PASS**. The plugin's own `TestLoadConfig`, `TestValidate`,
`TestApplyDefaults`, and `TestLoadConfigResolvesSecureSettingValue` suites also **PASS**.

After regeneration, `schema.gen.json` and `settings.gen.json` both moved `driver` into the `jsonData`
`required: ["driver"]` array and dropped the `x-dsconfig-required-when: "true"` extension.

---

## Files changed

- [`registry/grafana-odbc-datasource/dsconfig.json`](dsconfig.json) — changed `jsonData_driver` from
  `"requiredWhen": "true"` to `"required": true` (so it renders in the wizard's General step and emits
  OpenAPI `required`). It was the only `requiredWhen` in the schema.
- [`registry/grafana-odbc-datasource/schema.gen.json`](schema.gen.json),
  [`registry/grafana-odbc-datasource/settings.gen.json`](settings.gen.json) — regenerated by
  `go generate` (`driver` now in the `jsonData` `required` array; `x-dsconfig-required-when` removed).

_Unchanged by design:_ `settings.go`, `settings.ts`, `README.md`, `settings.examples.gen.json`,
`schema/conformance.go`, and `plugin-ui`.

> **Not dsconfig-fixable (noted, not changed):** `README.md` still describes the driver field as
> `requiredWhen: "true"` (lines 104, 186-189, 246) and is now slightly stale relative to the schema's
> `required: true`. Editing `README.md` is out of scope for this task, so it was left as-is.
> **Environment limitation:** the legacy Sqlyze plugin fails to start on this host, so the live legacy
> editor could not be screenshotted; legacy parity was established from the committed `README.md`
> source analysis of `ConfigEditor.tsx` plus the datasource-API entry.
