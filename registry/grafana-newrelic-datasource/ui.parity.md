# New Relic — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered from
[`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `grafana-newrelic-datasource`
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/cfrbqixuqpvy8c` (Grafana Enterprise; the New Relic plugin's own `ConfigEditor`, which renders a single **New Relic API Credentials** section — Personal API Key, Account ID, Region, Timeout. No `DataSourceHttpSettings`, no Custom HTTP Headers, no TLS/cert fields, no file pickers.)
- **New UI:** `http://192.168.1.241:58899/iframe.html?args=pluginType:grafana-newrelic-datasource&id=configeditor-datasourceconfigwizard--tab` (Storybook, `ConfigEditor/DatasourceConfigWizard`), also the `--wizard` story.
- **Method:** Playwright captured the legacy UI (`legacy-expand-newrelic-verify.png/.json`) and drove the new UI. The Storybook story fetches the schema from `raw.githubusercontent.com/.../registry/grafana-newrelic-datasource/dsconfig.json`; the local (edited) `dsconfig.json` was served to the new UI by intercepting that request with `context.route(...)`, so the screenshots reflect the local schema without pushing.
- **Result:** **Parity achieved.** The two **unconditionally-required** secrets were corrected from `requiredWhen: "true"` → `required: true`, so they fold into the wizard's synthetic **General** step and emit a proper secure `required: true` in the generated artifacts. **Custom HTTP Headers** and **`fileUpload`** were evaluated and correctly **not** used — the legacy editor has neither (`hasCustomHeaders:false`, `addHeaderBtn:false`, `fileInputs:0`, `uploadButtons:[]`). New Relic has **no conditional fields** (single auth method, no discriminator, no `dependsOn`), so there were no conditionals to reconcile.

---

## TL;DR of changes

| #   | Change | File | Why |
| --- | --- | --- | --- |
| 1   | Changed `secureJsonData_personalApiKey` and `secureJsonData_accountId` from `requiredWhen: "true"` → `required: true` | [`dsconfig.json`](dsconfig.json) | Both secrets are **unconditionally** required — there is no auth-type discriminator; the backend refuses to create the instance without either. `Validate` rejects a missing/whitespace personal API key (`"Enter a personal API key."`) and a missing/non-numeric/zero account ID (`"Enter an account ID. This must be a valid, positive number."`) (`settings.go:198-209`; upstream `pkg/datasource/handler_checkhealth.go:139-145`; `pkg/models/settings.go:42-46`). `required: true` is the correct canonical form; it also folds them into the wizard's synthetic **General** step and emits secure `required: true` instead of the CEL-string `x-dsconfig-required-when: "true"` extension (which the resolver stores but never evaluates). |
| 2   | Regenerated `schema.gen.json`, `settings.gen.json` via `go generate ./registry/grafana-newrelic-datasource/...` | generated artifacts | Keep committed artifacts in sync (guarded by `SchemaArtifactInSync`). |

No changes were made to `settings.go`, `settings.ts`, `README.md`, `settings.examples.gen.json`, `schema/conformance.go`, `schema/`, or `plugin-ui`. All schema changes flow through `dsconfig.json`; the rest is produced by `go generate`. **No conformance-test or plugin-ui change was required.**

---

## New UI verification

Tab mode (`newgen-newrelic-tab.json` + `.png`, per-field `probe-newrelic-tab-dom.json`) renders the
single **New Relic API Credentials** section with the header **"Fields marked with \* are required"**.
`hasHeadersEditor:false` (correct — no Custom HTTP Headers). `urlPresent:false` (correct — the New
Relic backend never reads the datasource root `url`; base URLs are backend-only overrides, see below).
Per-field required asterisks resolve exactly as expected:

| Field | Required (asterisk)? |
| --- | --- |
| Personal API Key / User API key | **yes \*** |
| Account ID | **yes \*** |
| Region | no |
| Timeout in Seconds | no |

### Wizard mode: Personal API Key + Account ID in the "General" step

In **wizard mode** the plugin-ui builds a synthetic first step (**General**, step **1/2**) from every
field marked `required: true`. `requiredWhen: "true"` is a CEL expression the resolver does **not**
inspect, so before this change neither secret would have been folded into General and the wizard would
have had a single step; after the change they are folded in.

**Verified (`newgen-newrelic-wiz.json`, `probe-newrelic-wiz-dom.json`, `newgen-newrelic-wiz.png`):**
the wizard opens on **General (1/2)** containing exactly **Personal API Key / User API key\*** and
**Account ID\*** — both marked required, both secure inputs with show/hide toggles — with
`headersEditor:false` / `fileInputs:0`. Region and Timeout are **not** on General (`found:false` in the
probe); they live on the subsequent group step (2/2 — **New Relic API Credentials**). Progression past
General is gated until the two required secrets are filled (Save & Test is disabled on step 1).

---

## Field-by-field parity

Legend: ✅ present & matching · 🔒 secure (masked) input · 🚫 backend-only (not rendered)

| Legacy UI field | Control (legacy) | New UI (schema id) | Control (new) | Storage target | Status |
| --- | --- | --- | --- | --- | --- |
| Personal API Key / User API key | password (secure) | `secureJsonData_personalApiKey` | secure input | `secureJsonData.personalApiKey` (**required**) | ✅ 🔒 |
| Account ID | number (secure) | `secureJsonData_accountId` | secure input | `secureJsonData.accountId` (**required**) | ✅ 🔒 |
| Region | Select (US / EU) | `jsonData_region` | select (placeholder `default`) | `jsonData.region` (optional) | ✅ |
| Timeout in Seconds | number input | `jsonData_timeoutInSeconds` | number (default `300`) | `jsonData.timeoutInSeconds` (optional) | ✅ |
| _(none — internal)_ | _(not in editor)_ | `jsonData_restBaseURL` | not rendered | `jsonData.restBaseURL` | ✅ 🚫 |
| _(none — internal)_ | _(not in editor)_ | `jsonData_infrastructureBaseURL` | not rendered | `jsonData.infrastructureBaseURL` | ✅ 🚫 |
| _(none — internal)_ | _(not in editor)_ | `jsonData_nerdGraphBaseURL` | not rendered | `jsonData.nerdGraphBaseURL` | ✅ 🚫 |

Notes:

- **Personal API Key + Account ID** are the single authentication method — both are always required,
  there is no auth-type discriminator. They render as masked secure inputs with show/hide toggles (the
  renderer draws any `target: "secureJsonData"` field that way). Account ID is a number in the legacy
  editor but is stored as a `secureJsonData` string (parsed to `Settings.AccountID int`).
- **Region** is optional (placeholder `default`); an empty region falls back to the New Relic client
  default (US). **Timeout in Seconds** is optional and defaults to 300.
- **`restBaseURL` / `infrastructureBaseURL` / `nerdGraphBaseURL`** are `backend-only` base-URL
  overrides for internal testing/mocking (`tags: ["backend-only"]`); they are intentionally **not**
  rendered in either UI and carry no `ui` block — parity with the legacy editor, which also never
  exposes them.

---

## `fileUpload` evaluation — not applicable to New Relic

The task asked to use the `fileUpload` control **only if the legacy UI uses it**. It does not:

- The legacy New Relic editor exposes only the API Key + Account ID + Region + Timeout. It has **no**
  TLS cert/key fields and **no** file pickers — the legacy DOM capture
  (`legacy-expand-newrelic-verify.json`) reports `fileInputs: 0`, `uploadButtons: []`.
- The new UI's `fileUpload` component only activates for `ui.fileMapping` (multi-key JSON
  distribution, e.g. a GCP service-account file); nothing in New Relic needs it.

**Decision:** do **not** add `fileUpload` to any New Relic field.

---

## Custom HTTP Headers — not applicable to New Relic

The legacy editor has **no** Custom HTTP Headers section (`hasCustomHeaders:false`, `addHeaderBtn:false`
in `legacy-expand-newrelic-verify.json`). The New Relic backend authenticates with the API key applied
via the client's `ConfigPersonalAPIKey` (`pkg/datasource/newrelic_client.go:43`); there is no
user-managed header editor. Headers are correctly **not** modeled, and the new UI shows no headers
editor in either mode (`hasHeadersEditor:false` in `newgen-newrelic-tab.json` and
`newgen-newrelic-wiz.json`). No change.

---

## Conditional fields — none

New Relic has a **single authentication method** and **no discriminator, no `effects`, and no
`dependsOn`**. Every field is either unconditionally required (`personalApiKey`, `accountId`) or plain
optional (`region`, `timeoutInSeconds`; the three base-URL overrides are backend-only). None
reveal/hide based on another field. There are therefore **no conditionals to test** — the
`requiredWhen: "true"` on the two secrets was a literal always-true flag (now correctly
`required: true`), not a real condition. No genuine conditional `requiredWhen` expression exists in
this entry, so none was left behind.

---

## Verification

```
go generate ./registry/grafana-newrelic-datasource/...   # regenerate schema.gen.json / settings.gen.json
go test ./registry/grafana-newrelic-datasource/...       # PASS
```

`TestSchemaConformance` subtests (New Relic) — all **8/8 PASS**:
`BaseFieldsResolved`, `SchemaRoundTrip`, `SchemaArtifactInSync`, `SchemaSpecHasNoSecureJSON`,
`ConfigSchemaValid`, `JSONDataMatchesStruct`, `JSONDataTypesMatchStruct`,
`SecureValuesMatchLoadSettings`.

The hand-authored `TestLoadConfig`, `TestApplyDefaults`, and `TestValidate` suites also pass unchanged
— including the negative cases that assert an empty/whitespace personal API key and a
missing/non-numeric/zero account ID are rejected, confirming the `required: true` conversion matches
the backend contract.

Generated-artifact effect (`git diff`): both secure values now carry `required: true` —

```
  "secureValues": [
-   { "key": "personalApiKey", "description": "Used for NRQL queries" },
-   { "key": "accountId",      "description": "Your New Relic Account ID" }
+   { "key": "personalApiKey", "description": "Used for NRQL queries", "required": true },
+   { "key": "accountId",      "description": "Your New Relic Account ID", "required": true }
  ]
```

New-UI captures: `newgen-newrelic-tab` (tab — single **New Relic API Credentials** section,
`hasHeadersEditor:false`, `urlPresent:false`), `probe-newrelic-tab-dom` (per-field required asterisks:
Personal API Key\* / Account ID\* required; Region + Timeout optional), `newgen-newrelic-wiz` +
`probe-newrelic-wiz-dom` (wizard opens on **General 1/2** with required Personal API Key + Account ID;
no headers editor; 0 file inputs). Legacy capture: `legacy-expand-newrelic-verify` (heading
**New Relic API Credentials**; `hasCustomHeaders:false`; `fileInputs:0`; `uploadButtons:[]`).

---

## Files changed

- [`registry/grafana-newrelic-datasource/dsconfig.json`](dsconfig.json) — changed
  `secureJsonData_personalApiKey` and `secureJsonData_accountId` from `requiredWhen: "true"` to
  `required: true`. No other fields carried a `requiredWhen`, so none was left in place.
- [`registry/grafana-newrelic-datasource/schema.gen.json`](schema.gen.json),
  [`registry/grafana-newrelic-datasource/settings.gen.json`](settings.gen.json) — regenerated by
  `go generate` (secure values `personalApiKey` and `accountId` now `required: true`).

_Unchanged by design:_ `settings.go`, `settings.ts`, `README.md`, `settings.examples.gen.json`,
`schema/conformance.go`, and everything in `plugin-ui`.

_Nothing required a change outside `dsconfig.json` (+ generated `.gen.json`)._ No conformance or
plugin-ui edits were necessary; the fix is a pure schema correction.
