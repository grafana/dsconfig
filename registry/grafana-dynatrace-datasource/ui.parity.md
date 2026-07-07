# Dynatrace — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered from
[`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `grafana-dynatrace-datasource`
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/cfrbqix2zuhogd` (Grafana Enterprise, Dynatrace custom `ConfigEditor.tsx` — **not** the stock `DataSourceHttpSettings`)
- **New UI:** `http://192.168.1.241:58899/?path=/story/configeditor-datasourceconfigwizard--tab&args=pluginType:grafana-dynatrace-datasource` (Storybook, `ConfigEditor/DatasourceConfigWizard`), also the `--wizard` story.
- **Method:** Playwright captured the legacy UI (`legacy-expand-dynatrace-parity.png`) and drove the new UI (both `--tab` and `--wizard`). The Storybook story fetches the schema from `raw.githubusercontent.com/.../registry/grafana-dynatrace-datasource/dsconfig.json`; the local (edited) `dsconfig.json` was served to the new UI by intercepting that request with `context.route(...)`, so the captures reflect the local schema without pushing.
- **Result:** **Parity achieved with one fix.** `jsonData_environmentId` was promoted from `requiredWhen: "true"` → `required: true` so it renders in the wizard's synthetic **General** step and emits a proper `required: ["environmentId"]` in the generated spec. No HTTP headers and no `fileUpload` are modeled — the legacy editor has neither.

---

## TL;DR of changes

| #   | Change                                                                                        | File                             | Why                                                                                                                                       |
| --- | --------------------------------------------------------------------------------------------- | -------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | Changed `jsonData_environmentId` from `requiredWhen: "true"` → `required: true`               | [`dsconfig.json`](dsconfig.json) | Environment ID is **unconditionally** required in every apiType mode (`settings.go:197-199`, health check `handler_healthcheck.go:142-144`). Puts the field into the wizard's synthetic **General** step and emits `required: ["environmentId"]` instead of the opaque `x-dsconfig-required-when: "true"` extension. |
| 2   | Regenerated `schema.gen.json`, `settings.gen.json` via `go generate ./registry/grafana-dynatrace-datasource/...` | generated artifacts | Keep committed artifacts in sync (guarded by `SchemaArtifactInSync`).                                                                     |

No changes were made to `settings.go`, `settings.ts`, `README.md`, `conformance.go`, or `plugin-ui`.
All schema changes flow through `dsconfig.json`; the rest is produced by `go generate`.

**No conformance-test or plugin-ui change was required.** The plugin-ui wizard already recognises
the conventional `authentication` group id and already pulls `required: true` fields into the
General step; the shared conformance walker handled the change with no edits.

The four **real** conditional `requiredWhen` expressions were left untouched:

- `jsonData_domain` — `requiredWhen: "jsonData_apiType == 'managed'"`
- `secureJsonData_apiToken` — `requiredWhen: "secureJsonData_platformToken == ''"`
- `secureJsonData_platformToken` — `requiredWhen: "secureJsonData_apiToken == ''"`
- `secureJsonData_tlsCACert` — `requiredWhen: "jsonData_tlsAuthWithCACert == true"`

Only the always-true sentinel (`environmentId`) was converted; no redundant `required: true` pre-existed, so nothing was removed.

---

## Findings (fix + non-fixes)

**Required-field fix (`environmentId`).** The only change. `requiredWhen: "true"` is an opaque CEL
string the wizard's `resolveRequiredFieldsGroup` does **not** inspect, so the field was excluded
from the synthetic **General** step and the generated spec marked it with the
`x-dsconfig-required-when: "true"` extension instead of a real `required` entry. Since
`environmentId` is required in every mode, `required: true` is the correct, unconditional model.

**No Custom HTTP Headers.** Dynatrace uses a **custom** `ConfigEditor` (token fields + connection
radio), not Grafana's stock `DataSourceHttpSettings`. Its legacy editor has **no** Custom HTTP
Headers section (`hasCustomHeaders:false`, `addHeaderBtn:false`). Headers are correctly **not**
modeled.

**No `fileUpload`.** All secrets (API token, Platform token) and the optional CA certificate are
entered as text / textarea — there is **no** file-upload control (`fileInputs:0`,
`uploadButtons:[]`). `fileUpload` is correctly **not** used. (The CA Cert is a multiline `textarea`
that expects pasted PEM text, matching legacy `ConfigEditor.tsx:184`.)

**No auth-method selector.** Unlike AWS/Azure datasources, Dynatrace has **no** auth discriminator:
both token fields are always shown and at least one is required (enforced by the mutual
`requiredWhen` pair). There are therefore **no `effects` to exercise**.

---

## New UI verification

**Tab mode** renders all three sections: **Connection**, **Authentication**, **Additional
settings** (optional accordion). `hasHeadersEditor:false` (correct — parity with legacy).
`urlPresent:false` is expected: Dynatrace has no stock URL field; the base URL is built from
`apiType` + `environmentId` (+ `domain`). Environment ID renders with the required `*` marker
(`verify-dynatrace-result.json`: `tabSaaS.envIdRequired: true`).

**Wizard mode** opens on **General 1/4**. 

- **Before** the fix (`newgen-dynatrace-before-wiz.json`, served from a temp copy with
  `requiredWhen: "true"`): General step = **[Dynatrace API Token, Dynatrace Platform Token]** — no
  Environment ID, no required marker.
- **After** the fix (`newgen-dynatrace-wiz.json` / `verify-dynatrace-result.json`): General step =
  **[Environment ID \*, Dynatrace API Token, Dynatrace Platform Token]** (`wizardGeneral.hasEnvId:
  true`, `envIdRequired: true`, `hasApiToken: true`, `hasPlatformToken: true`).

The two tokens appear in General because they sit in the conventional `authentication` group, which
the wizard folds into the first step.

## Field-by-field parity

Legend: ✅ present & matching · 🔀 conditional (revealed by `dependsOn`) · 🏷️ relabeled by `overrides`

| Legacy UI field           | Control (legacy)             | New UI (schema id)              | Storage target                    | Status                    |
| ------------------------- | ---------------------------- | ------------------------------- | --------------------------------- | ------------------------- |
| Dynatrace API Type        | radio (SaaS/Managed/Raw URL) | `jsonData_apiType`              | `jsonData.apiType`                | ✅                        |
| Environment ID            | text input                   | `jsonData_environmentId`        | `jsonData.environmentId` (required) | ✅ (required) 🏷️ (→ URL) |
| Domain                    | text input                   | `jsonData_domain`               | `jsonData.domain`                 | ✅ 🔀 (managed)           |
| Dynatrace API Token       | password (secure)            | `secureJsonData_apiToken`       | `secureJsonData.apiToken`         | ✅ (req if no platform)   |
| Dynatrace Platform Token  | password (secure)            | `secureJsonData_platformToken`  | `secureJsonData.platformToken`    | ✅ (req if no api)        |
| Timeout                   | number input                 | `jsonData_httpClientTimeout`    | `jsonData.httpClientTimeout`      | ✅ (opt)                  |
| Skip TLS Verify           | checkbox                     | `jsonData_tlsSkipVerify`        | `jsonData.tlsSkipVerify`          | ✅ (opt)                  |
| With CA Cert              | checkbox                     | `jsonData_tlsAuthWithCACert`    | `jsonData.tlsAuthWithCACert`      | ✅ (opt)                  |
| CA Cert                   | textarea (PEM)               | `secureJsonData_tlsCACert`      | `secureJsonData.tlsCACert`        | ✅ 🔀 (With CA Cert)      |

**Note on secure inputs:** `dsconfig.json` declares `ui.component: "input"` for the token secrets,
but the new renderer draws any `target: "secureJsonData"` field as a masked secure input with a
show/hide toggle. Both UIs collect the same secret into the same `secureJsonData` key — only the
widget affordance differs.

## Conditional fields — tested

All conditionals render per the schema `dependsOn` / `overrides`
(`verify-dynatrace-result.json`, screenshots `verify-dynatrace-tab-{A,B,C,D}-*.png`):

| Trigger / relationship                              | Effect                                                  | Verified |
| --------------------------------------------------- | ------------------------------------------------------ | -------- |
| `jsonData_apiType == 'managed'` (`dependsOn`)       | **Domain** field revealed (hidden for saas/url)         | ✅ (`tabSaaS.hasDomain:false` → `tabManaged.hasDomain:true`) |
| `jsonData_apiType == 'url'` (`overrides`)           | Environment ID field relabeled / placeholder → **Dynatrace URL** | ✅ (`tabURL.placeholders` includes `"Dynatrace URL"`) |
| `jsonData_tlsAuthWithCACert == true` (`dependsOn`)  | **CA Cert** textarea revealed                           | ✅ (`tlsBefore.hasCACert:false` → `tlsAfter.hasCACert:true`) |
| token pair `requiredWhen` (apiToken / platformToken) | at least one token required (mutually)                 | ✅ (declared; both always shown; `relationships[type=group]`) |
| CA Cert pair `requiredWhen` (tlsAuthWithCACert / tlsCACert) | CA Cert required once "With CA Cert" is on       | ✅ (declared; `relationships[type=pair]`) |

---

## `fileUpload` evaluation — not applicable to Dynatrace

- Legacy editor exposes **no file inputs and no upload buttons**
  (`legacy-expand-dynatrace-parity.json`: `fileInputs: 0`, `uploadButtons: []`).
- Secrets are pasted into secure text inputs; the CA certificate is pasted as PEM text into a
  `textarea`. Nothing needs the multi-key `ui.fileMapping` that activates `fileUpload`.

**Decision:** do **not** add `fileUpload` to any Dynatrace field. **Packs:** none needed — no
multi-field bundle controls in the legacy editor.

---

## Verification

```
go generate ./registry/grafana-dynatrace-datasource/...          # regenerate schema.gen.json / settings.gen.json
go test ./registry/grafana-dynatrace-datasource/...              # PASS
go test ./registry/grafana-dynatrace-datasource/... ./schema/... # PASS (no regressions)
```

`TestSchemaConformance` subtests — all **PASS**:
`BaseFieldsResolved`, `SchemaRoundTrip`, `SchemaArtifactInSync`, `SchemaSpecHasNoSecureJSON`,
`ConfigSchemaValid`, `JSONDataMatchesStruct`, `JSONDataTypesMatchStruct`,
`SecureValuesMatchLoadSettings`.

The hand-authored `TestLoadConfig`, `TestLoadConfigValidCACert`, `TestApplyDefaults`,
`TestValidate`, and `TestSettingsExamples` suites also pass unchanged (they do not reference the
required flag; `Validate()` already rejected an empty `environmentId`).

New-UI captures: `newgen-dynatrace-before-wiz` (before: General step has no Environment ID),
`newgen-dynatrace-tab` / `newgen-dynatrace-wiz` (after), `verify-dynatrace-*` (required marker +
`apiType`/`tlsAuthWithCACert` conditionals). Legacy capture: `legacy-expand-dynatrace-parity`
(no Custom HTTP Headers, 0 file inputs).

---

## Files changed

- [`registry/grafana-dynatrace-datasource/dsconfig.json`](dsconfig.json) — changed
  `jsonData_environmentId` from `requiredWhen: "true"` to `required: true` (renders in the wizard's
  General step). Real conditional `requiredWhen` expressions (domain, apiToken, platformToken,
  tlsCACert) left unchanged.
- [`registry/grafana-dynatrace-datasource/schema.gen.json`](schema.gen.json),
  [`registry/grafana-dynatrace-datasource/settings.gen.json`](settings.gen.json) — regenerated by
  `go generate` (`jsonData.required: ["environmentId"]` added; `x-dsconfig-required-when: "true"`
  removed from `environmentId`).

_Unchanged by design:_ `settings.go`, `settings.ts`, `README.md`, `settings.examples.gen.json`,
`schema/conformance.go`, and everything in `plugin-ui`.
