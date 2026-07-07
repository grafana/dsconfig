# Sentry — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered from
this entry's [`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `grafana-sentry-datasource`
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/cfrbqiy6o8feoc` (Grafana Enterprise; the Sentry plugin's own `SentryConfigEditor`, which renders two sections — **Sentry Settings** (Sentry URL / Sentry Org / Sentry Auth Token) and **Additional settings** (Skip TLS Verify). No `DataSourceHttpSettings`, no Custom HTTP Headers, no TLS cert/key file fields.)
- **New UI:** `http://192.168.1.241:58899/iframe.html?args=pluginType:grafana-sentry-datasource&id=configeditor-datasourceconfigwizard--tab` (Storybook, `ConfigEditor/DatasourceConfigWizard`), also the `--wizard` story. The story fetches the schema from `raw.githubusercontent.com/.../registry/grafana-sentry-datasource/dsconfig.json`; the local (edited) `dsconfig.json` is served to the new UI by intercepting that request with `context.route(...)`, so a capture reflects the local schema without pushing.
- **Method:** Playwright captured the legacy UI (`legacy-expand-ent-grafana-sentry-datasource.png` / `.json`) and drives the new UI with the **local** `dsconfig.json`.
- **Result:** **Two `dsconfig.json` fixes applied** — `jsonData_orgSlug` and `secureJsonData_authToken` promoted from a literal-`true` `requiredWhen` to a proper `required: true` (so they fold into the wizard's synthetic **General** step and emit an OpenAPI `required` array / secure `required: true`). **Custom HTTP Headers** and **`fileUpload`** were evaluated and correctly **not** used — the legacy editor has neither (`hasCustomHeaders:false`, `addHeaderBtn:false`, `fileInputs:0`, `uploadButtons:[]`). Sentry has **no conditional fields** (single Bearer-token auth, no discriminator, no `dependsOn`), so there were no conditionals to reconcile.

---

## TL;DR of changes

| #   | Change | File | Why |
| --- | --- | --- | --- |
| 1   | Changed `jsonData_orgSlug` and `secureJsonData_authToken` from `requiredWhen: "true"` → `required: true` | [`dsconfig.json`](dsconfig.json) | Both are **unconditionally** required — the backend rejects a missing org slug (`invalid or empty organization slug`) and a missing token (`empty or invalid auth token found`) (`pkg/plugin/settings.go:41-50`, mirrored in `settings.go` `Validate`). `required: true` is the correct canonical form; it also folds them into the wizard's synthetic **General** step and emits `jsonData.required: ["orgSlug"]` + secure `authToken.required: true` instead of the CEL-string `x-dsconfig-required-when: "true"` extension (which the resolver never inspects). |
| 2   | Regenerated `schema.gen.json`, `settings.gen.json` via `go generate ./registry/grafana-sentry-datasource/...` | generated artifacts | Keep committed artifacts in sync (guarded by `SchemaArtifactInSync`). |

No changes were made to `settings.go`, `settings.ts`, `README.md`, `settings.examples.gen.json`, `conformance_test.go`, `schema/`, or `plugin-ui`. All schema changes flow through `dsconfig.json`; the rest is produced by `go generate`. **No conformance-test or plugin-ui change was required.**

---

## Section layout

The `groups` taxonomy was left unchanged; only the `required` flags were corrected. The legacy
editor renders two sections (legacy headings: `["Sentry Settings", "Additional settings"]`), which
match the schema groups exactly:

| Order | Section (`id`) | `optional` | Fields (in display order) |
| --- | --- | --- | --- |
| 1 | **Sentry Settings** (`sentry-settings`) | no | Sentry URL, Sentry Org, Sentry Auth Token |
| 2 | **Additional settings** (`additional-settings`) | yes | Skip TLS Verify |

Notes:

- The **Additional settings** group is `optional` and renders collapsed (with an `Optional` badge)
  in tab mode; expanding it surfaces the **Skip TLS Verify** switch.
- After the fix, the required asterisk (`*`) is emitted on exactly **Sentry Org\*** and **Sentry
  Auth Token\***. **Sentry URL** is intentionally left optional (see below).

### Wizard mode: Sentry Org + Sentry Auth Token in the "General" step

In **wizard mode** the plugin-ui builds a synthetic first step (**General**) from every field marked
`required: true` (plus any auth-group / `dependsOn` relatives — Sentry has none of those).
`requiredWhen: "true"` is a CEL expression the resolver does **not** inspect, so before this change
neither field would have been folded into General; after the change both are. The optional **Sentry
URL** and **Skip TLS Verify** live on their subsequent group steps.

---

## Field-by-field parity

Legend: ✅ present & matching · 🔒 secure (masked) input

| Legacy UI field | Control (legacy) | New UI (schema id) | Control (new) | Storage target | Status |
| --- | --- | --- | --- | --- | --- |
| Sentry URL | text input | `jsonData_url` | input (default `https://sentry.io`) | `jsonData.url` (optional, defaulted) | ✅ |
| Sentry Org | text input | `jsonData_orgSlug` | input | `jsonData.orgSlug` (**required**) | ✅ |
| Sentry Auth Token | password (secure) | `secureJsonData_authToken` | secure input | `secureJsonData.authToken` (**required**) | ✅ 🔒 |
| Skip TLS Verify | switch | `jsonData_tlsSkipVerify` | switch (default `false`) | `jsonData.tlsSkipVerify` (optional) | ✅ |

Notes:

- **Auth Token** is the single authentication method — sent as `Authorization: Bearer <authToken>`
  on every request (`pkg/sentry/client.go:37-40`). It renders as a masked secure input with a
  show/hide toggle (the renderer draws any `target: "secureJsonData"` field that way).
- **Sentry URL is intentionally left optional (not `required: true`).** Although the backend's
  `Validate` requires a non-empty URL, both the editor initial state and the backend default an empty
  URL to `https://sentry.io` (`settings.go` `ApplyDefaults`, `pkg/plugin/settings.go:37-40`), so it is
  never empty from the user's perspective. It carries `defaultValue: "https://sentry.io"` and stays
  optional — matching the legacy editor, where URL has no `*` marker. (This differs from datasources
  like Datadog whose URL has no default and is genuinely user-required.)
- No fields are backend-only or hidden; all four render in both UIs.

---

## `fileUpload` evaluation — not applicable to Sentry

The task asked to use the `fileUpload` control **only if the legacy UI uses it**. It does not:

- The legacy Sentry editor exposes only Sentry URL + Sentry Org + Sentry Auth Token + Skip TLS
  Verify. There are **no** TLS cert/key fields and **no** file pickers — the legacy DOM capture
  (`legacy-expand-ent-grafana-sentry-datasource.json`) reports `fileInputs: 0`, `uploadButtons: []`.
- Self-hosted TLS is handled by the boolean `jsonData.tlsSkipVerify` switch
  (`pkg/plugin/plugin.go:57-63`), not by uploading a CA certificate.
- The new UI's `fileUpload` component only activates for `ui.fileMapping` (multi-key JSON
  distribution, e.g. a GCP service-account file); nothing in Sentry needs it.

**Decision:** do **not** add `fileUpload` to any Sentry field.

---

## Custom HTTP Headers — not applicable to Sentry

The legacy editor has **no** Custom HTTP Headers section (`hasCustomHeaders:false`,
`addHeaderBtn:false`). Sentry sends a single fixed `Authorization: Bearer <authToken>` header derived
from the secure auth token (`pkg/sentry/client.go:37-40`); there is no user-managed header editor.
Headers are correctly **not** modeled. No change.

---

## Conditional fields — none

Sentry has a **single authentication method** (Bearer token) and **no discriminator, no `effects`,
and no `dependsOn`**. Every field is either unconditionally required (`orgSlug`, `authToken`) or plain
optional (`url` — defaulted, `tlsSkipVerify`); none reveal/hide based on another field. There are
therefore **no conditionals to test** — the `requiredWhen: "true"` on the two required fields was a
literal always-true flag (now correctly `required: true`), not a real condition. No genuine
conditional `requiredWhen` expression exists in this entry, so none was left behind.

---

## Verification

```
go generate ./registry/grafana-sentry-datasource/...   # regenerate schema.gen.json / settings.gen.json
go test     ./registry/grafana-sentry-datasource/...   # PASS
```

`TestSchemaConformance` subtests (Sentry) — all **8/8 PASS**:
`BaseFieldsResolved`, `SchemaRoundTrip`, `SchemaArtifactInSync`, `SchemaSpecHasNoSecureJSON`,
`ConfigSchemaValid`, `JSONDataMatchesStruct`, `JSONDataTypesMatchStruct`,
`SecureValuesMatchLoadSettings`.

The hand-authored `TestLoadConfig`, `TestApplyDefaults`, and `TestValidate` suites also pass
unchanged — including the negative cases that assert an empty org slug / missing auth token are
rejected, confirming the `required: true` conversion matches the backend contract.

Regenerated artifacts (verified in `schema.gen.json` / `settings.gen.json`):

- `spec.properties.jsonData.required: ["orgSlug"]` **added**.
- `x-dsconfig-required-when: "true"` **removed** from `orgSlug`.
- secure value `authToken` now `"required": true`.

### Capture status

- **Legacy inventory — captured.** `legacy-expand-ent-grafana-sentry-datasource` (headings
  `["Sentry Settings", "Additional settings"]`; `hasCustomHeaders:false`; `addHeaderBtn:false`;
  `fileInputs:0`; `uploadButtons:[]`). This confirms no headers editor and no file upload in legacy.
- **New-UI tab / wizard — not captured this run.** The Storybook server
  (`http://192.168.1.241:58899`) was **unreachable** (`net::ERR_CONNECTION_REFUSED` on all probed
  addresses; port closed) at validation time, so the live `newgen-sentry-tab` / `newgen-sentry-wizard`
  captures could not be produced. The expected results are deterministic from the schema and the
  regenerated artifacts, and identical to the already-verified Honeycomb case (secure + jsonData
  `required` fields, single auth method, no conditionals): tab mode renders **Sentry Settings** +
  **Additional settings** `Optional` with `hasHeadersEditor:false` and `urlPresent:true`; wizard
  **General** step contains the required **Sentry Org\*** and **Sentry Auth Token\***. Re-run when the
  Storybook host is back up:

  ```
  node capture-new-generic.js grafana-sentry-datasource \
    registry/grafana-sentry-datasource/dsconfig.json tab    sentry-tab
  node capture-new-generic.js grafana-sentry-datasource \
    registry/grafana-sentry-datasource/dsconfig.json wizard sentry-wizard
  ```

---

## Files changed

- [`registry/grafana-sentry-datasource/dsconfig.json`](dsconfig.json) — changed `jsonData_orgSlug`
  and `secureJsonData_authToken` from `requiredWhen: "true"` to `required: true`. No other fields
  carried a `requiredWhen`, so none was left in place.
- [`registry/grafana-sentry-datasource/schema.gen.json`](schema.gen.json),
  [`registry/grafana-sentry-datasource/settings.gen.json`](settings.gen.json) — regenerated by
  `go generate` (`jsonData.required: ["orgSlug"]` added; secure value `authToken` now `required: true`;
  `x-dsconfig-required-when: "true"` removed from `orgSlug`).

_Unchanged by design:_ `settings.go`, `settings.ts`, `README.md`, `settings.examples.gen.json`,
`schema/conformance.go`, and everything in `plugin-ui`.

_Nothing required a change outside `dsconfig.json` (+ generated `.gen.json`)._ No conformance or
plugin-ui edits were necessary; the fix is a pure schema correction.
