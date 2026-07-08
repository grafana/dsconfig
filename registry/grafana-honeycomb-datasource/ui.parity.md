# Honeycomb — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered
from this entry's [`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `grafana-honeycomb-datasource`
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/ffrbqixcwxz40c` (Grafana Enterprise; the Honeycomb plugin's own `ConfigEditor`, which renders three sections — **Access** (API Key), **Environment** (URL / Team Name / Environment Name), and **Advanced Settings** (Time Window). No `DataSourceHttpSettings`, no Custom HTTP Headers, no TLS/cert fields.)
- **New UI:** `http://192.168.1.241:58899/iframe.html?args=pluginType:grafana-honeycomb-datasource&id=configeditor-datasourceconfigwizard--tab` (Storybook, `ConfigEditor/DatasourceConfigWizard`), also the `--wizard` story.
- **Method:** Playwright captured both UIs (full-page screenshots + DOM extraction). The Storybook story fetches the schema from `raw.githubusercontent.com/.../registry/grafana-honeycomb-datasource/dsconfig.json`; the local (edited) `dsconfig.json` was served to the new UI by intercepting that request with `context.route(...)`, so the screenshots reflect the local schema without pushing.
- **Result:** **Parity achieved.** The three **unconditionally-required** fields were corrected from `requiredWhen: "true"` → `required: true` (so they fold into the wizard's synthetic **General** step and emit a proper OpenAPI `required` array / secure `required: true`). **Custom HTTP Headers** and **`fileUpload`** were evaluated and correctly **not** used — the legacy editor has neither (`hasCustomHeaders:false`, `addHeaderBtn:false`, `fileInputs:0`, `uploadButtons:[]`). Honeycomb has **no conditional fields** (single auth method, no discriminator, no `dependsOn`), so there were no conditionals to reconcile.

---

## TL;DR of changes

| #   | Change | File | Why |
| --- | --- | --- | --- |
| 1   | Changed `secureJsonData_apiKey`, `jsonData_hostname`, and `jsonData_team` from `requiredWhen: "true"` → `required: true` | [`dsconfig.json`](dsconfig.json) | All three are **unconditionally** required — the backend `Validate` rejects a missing API key (`enter an API key`), empty/non-https hostname (`enter a URL` / `scheme must be https`), and missing team (`enter a Honeycomb team name`) (`pkg/models/settings.go:45-71`). `required: true` is the correct canonical form; it also folds them into the wizard's synthetic **General** step and emits `jsonData.required: ["hostname","team"]` + secure `apiKey.required: true` instead of the CEL-string `x-dsconfig-required-when: "true"` extension. |
| 2   | Regenerated `schema.gen.json`, `settings.gen.json` via `go generate ./registry/grafana-honeycomb-datasource/...` | generated artifacts | Keep committed artifacts in sync (guarded by `SchemaArtifactInSync`). |

No changes were made to `settings.go`, `settings.ts`, `README.md`, `conformance.go`, `schema/`, or `plugin-ui`. All schema changes flow through `dsconfig.json`; the rest is produced by `go generate`. **No conformance-test or plugin-ui change was required.**

---

## Section layout

The `groups` taxonomy was left unchanged; only the `required` flags were corrected. Verified
rendering top-to-bottom in the new UI (tab mode, `newgen-honeycomb-tab.json` + screenshot), which
matches the legacy section order exactly (legacy headings: `["Access", "Environment", "Advanced Settings"]`):

| Order | Section (`id`) | `optional` | Fields (in display order) |
| --- | --- | --- | --- |
| 1 | **Access** (`access`) | no | Honeycomb API Key |
| 2 | **Environment** (`environment`) | no | URL, Team Name, Environment Name |
| 3 | **Advanced Settings** (`advanced-settings`) | yes | Time Window (days) |

Notes:

- The **Advanced Settings** group is `optional` and renders collapsed (with an `Optional` badge) in
  tab mode; expanding it surfaces the **Time Window (days)** number input.
- The new UI shows the header **"Fields marked with \* are required"** and renders a required
  asterisk on exactly **Honeycomb API Key\***, **URL\***, and **Team Name\*** — verified per-field in
  `dump-honeycomb-labels.json`: `{apiKey:*, URL:*, Team Name:*, Environment Name:(none), Time Window (days):(none)}`.

### Wizard mode: API Key + URL + Team in the "General" step

In **wizard mode** the plugin-ui builds a synthetic first step (**General**, step **1/4**) from every
field marked `required: true` (plus any auth-group / `dependsOn` relatives — Honeycomb has none of
those). `requiredWhen: "true"` is a CEL expression the resolver does **not** inspect, so before this
change none of the three fields would have been folded into General; after the change they are.

**Verified (`verify-honeycomb-wizard-walk.json`):** the wizard opens on **General (1/4)** containing
exactly **Honeycomb API Key\***, **URL\***, and **Team Name\*** — all three marked required — and with
`headersEditor:false` / `fileInputs:0`. Progression past General is gated until those required
fields are filled (the optional **Environment Name** and **Time Window (days)** live on the
subsequent group steps: Access / Environment / Advanced Settings → the 4 steps of "1/4").

---

## Field-by-field parity

Legend: ✅ present & matching · 🔒 secure (masked) input

| Legacy UI field | Control (legacy) | New UI (schema id) | Control (new) | Storage target | Status |
| --- | --- | --- | --- | --- | --- |
| Honeycomb API Key | password (secure) | `secureJsonData_apiKey` | secure input | `secureJsonData.apiKey` (**required**) | ✅ 🔒 |
| URL | text input | `jsonData_hostname` | input (default `https://api.honeycomb.io`, `^https://` pattern) | `jsonData.hostname` (**required**) | ✅ |
| Team Name | text input | `jsonData_team` | input | `jsonData.team` (**required**) | ✅ |
| Environment Name | text input | `jsonData_environment` | input | `jsonData.environment` (optional) | ✅ |
| Time Window (days) | number input | `jsonData_retentionLimit` | number (default `7`) | `jsonData.retentionLimit` (optional) | ✅ |

Notes:

- **API Key** is the single authentication method (sent as the `X-Honeycomb-Team` header). It
  renders as a masked secure input with a show/hide toggle (the renderer draws any
  `target: "secureJsonData"` field that way).
- **URL / Team Name are unconditionally required**, not `dependsOn`-gated. The health check fails
  without a team even though it only affects data links (`pkg/plugin/healthcheck.go`).
- No fields are backend-only or hidden; all five render in both UIs.

---

## `fileUpload` evaluation — not applicable to Honeycomb

The task asked to use the `fileUpload` control **only if the legacy UI uses it**. It does not:

- The legacy Honeycomb editor exposes only the API Key + URL + Team/Environment + Time Window. It has
  **no** TLS cert/key fields and **no** file pickers — the legacy DOM capture
  (`legacy-expand-honeycomb-verify.json`) reports `fileInputs: 0`, `uploadButtons: []`.
- The new UI's `fileUpload` component only activates for `ui.fileMapping` (multi-key JSON
  distribution, e.g. a GCP service-account file); nothing in Honeycomb needs it.

**Decision:** do **not** add `fileUpload` to any Honeycomb field.

---

## Custom HTTP Headers — not applicable to Honeycomb

The legacy editor has **no** Custom HTTP Headers section (`hasCustomHeaders:false`,
`addHeaderBtn:false`). Honeycomb sends a single fixed `X-Honeycomb-Team` header derived from the API
key (`pkg/httpclient/client.go:39-42`); there is no user-managed header editor. Headers are correctly
**not** modeled, and the new UI shows no headers editor in either mode (`hasHeadersEditor:false` in
`newgen-honeycomb-tab.json` and `verify-honeycomb-wizard-walk.json`). No change.

---

## Conditional fields — none

Honeycomb has a **single authentication method** and **no discriminator, no `effects`, and no
`dependsOn`**. Every field is either unconditionally required (`apiKey`, `hostname`, `team`) or plain
optional (`environment`, `retentionLimit`); none reveal/hide based on another field. There are
therefore **no conditionals to test** — the `requiredWhen: "true"` on the three required fields was a
literal always-true flag (now correctly `required: true`), not a real condition. No genuine
conditional `requiredWhen` expression exists in this entry, so none was left behind.

---

## Verification

```
go generate ./registry/grafana-honeycomb-datasource/...   # regenerate schema.gen.json / settings.gen.json
go test ./registry/grafana-honeycomb-datasource/...       # PASS
```

`TestSchemaConformance` subtests (Honeycomb) — all **8/8 PASS**:
`BaseFieldsResolved`, `SchemaRoundTrip`, `SchemaArtifactInSync`, `SchemaSpecHasNoSecureJSON`,
`ConfigSchemaValid`, `JSONDataMatchesStruct`, `JSONDataTypesMatchStruct`,
`SecureValuesMatchLoadSettings`.

The hand-authored `TestLoadConfig` (13), `TestApplyDefaults` (3), and `TestValidate` (7) suites also
pass unchanged — including the negative cases that assert an empty API key / missing team / empty or
non-https hostname are rejected, confirming the `required: true` conversion matches the backend
contract.

New-UI captures: `newgen-honeycomb-tab` (tab — sections Access / Environment / Advanced Settings
`Optional`, `hasHeadersEditor:false`, `urlPresent:true`), `dump-honeycomb-labels` (per-field required
asterisks: API Key\* / URL\* / Team Name\* required; Environment Name + Time Window optional),
`verify-honeycomb-wizard-walk` (wizard opens on **General 1/4** with required API Key + URL + Team
Name; no headers editor; 0 file inputs). Legacy capture:
`legacy-expand-honeycomb-verify` (headings Access / Environment / Advanced Settings; `hasCustomHeaders:false`;
`fileInputs:0`; `uploadButtons:[]`).

---

## Files changed

- [`registry/grafana-honeycomb-datasource/dsconfig.json`](dsconfig.json) — changed
  `secureJsonData_apiKey`, `jsonData_hostname`, and `jsonData_team` from `requiredWhen: "true"` to
  `required: true`. No other fields carried a `requiredWhen`, so none was left in place.
- [`registry/grafana-honeycomb-datasource/schema.gen.json`](schema.gen.json),
  [`registry/grafana-honeycomb-datasource/settings.gen.json`](settings.gen.json) — regenerated by
  `go generate` (`jsonData.required: ["hostname","team"]` added; secure value `apiKey` now
  `required: true`; `x-dsconfig-required-when: "true"` removed from `hostname` / `team`).

_Unchanged by design:_ `settings.go`, `settings.ts`, `README.md`, `settings.examples.gen.json`,
`schema/conformance.go`, and everything in `plugin-ui`.

_Nothing required a change outside `dsconfig.json` (+ generated `.gen.json`)._ No conformance or
plugin-ui edits were necessary; the fix is a pure schema correction.
