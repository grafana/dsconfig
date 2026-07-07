# Azure DevOps — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered from
[`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `grafana-azuredevops-datasource`
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/bfrbqiwlb2dj4d` (Grafana Enterprise 13.0.1)
- **New UI:** `http://192.168.1.241:58899/iframe.html?id=configeditor-datasourceconfigwizard--tab&args=pluginType:grafana-azuredevops-datasource` (also `--wizard`)
- **Method:** Playwright captured the legacy UI (`legacy-expand-azuredevops.png`) and drove the new UI in both tab and wizard modes (local schema served via `context.route('**/registry/grafana-azuredevops-datasource/dsconfig.json', …)`).
- **Result:** **One `dsconfig.json` fix applied** — the PAT secret's unconditional `requiredWhen:"true"` was converted to `required:true`, so the General step now renders the required (`*`) marker that matches the backend contract. Headers (n/a) and file upload (n/a) needed no change.

---

## Fix applied

**Required-field / General-step fix.** `secureJsonData_patToken` used
`"requiredWhen": "true"` — an always-true conditional expression, which is the wrong
idiom for a field that is *unconditionally* required. It is now
`"required": true`. The PAT is a hard requirement of the backend: `Validate()` fails
with "invalid PAT" when `secureJsonData.patToken` is empty
(`pkg/plugin/settings.go:23-25,42-44`). After the fix, both the tab and wizard render
**PAT \*** with the red required asterisk, alongside the already-required **URL \***.

```diff
             "target": "secureJsonData",
             "role": "auth.basic.password",
-            "requiredWhen": "true",
+            "required": true,
```

No conditional `requiredWhen` expressions exist on this datasource, so none were touched.

## Findings (no other fixes needed)

**No Custom HTTP Headers.** Azure DevOps authenticates with a PAT sent as HTTP Basic auth
(empty username + PAT via `azuredevops.NewPatConnection`, or an explicit
`CreateBasicAuthHeaderValue(username, patToken)` when a username is set —
`pkg/plugin/plugin.go:74,76-84`). The legacy editor is a custom editor with **no** Custom
HTTP Headers section (`hasCustomHeaders:false`, `addHeaderBtn:false`). Headers are
correctly **not** modeled.

**No `fileUpload`.** All credentials are entered as text (URL, PAT, optional username) —
there is **no** file-upload control (`fileInputs:0`, `uploadButtons:[]`). `fileUpload` is
correctly **not** used.

## New UI verification

- **Tab mode** renders both sections: **Azure DevOps settings** (URL \*, PAT \*) and the
  collapsed, opt **Optional Configuration** (Projects limit, Username). "Fields marked
  with * are required" is shown. `hasHeadersEditor:false` (correct), `urlPresent:true`.
- **Wizard mode** — the **General** step (1/3) contains the two required fields **URL \***
  and **PAT \*** (the PAT rendered as a password input with a reveal/eye toggle). The
  token is on the General step, as required. `hasHeadersEditor:false`.

## Field-by-field parity

| Legacy field   | schema id                  | Target           | Status |
| -------------- | -------------------------- | ---------------- | ------ |
| URL            | `jsonData_url`             | `jsonData`       | ✅ (required) |
| PAT            | `secureJsonData_patToken`  | `secureJsonData` | ✅ (required — fixed) |
| Projects limit | `jsonData_projectsLimit`   | `jsonData`       | ✅ (opt, default 100) |
| Username       | `jsonData_username`        | `jsonData`       | ✅ (opt) |
| _(none)_       | `jsonData_authType`        | `jsonData`       | ✅ 🔀 (fixed `patToken`, backend-declared-unused; no UI control in legacy) |

Both editors group the fields identically: **Azure DevOps settings** (URL + PAT) and
**Optional Configuration** (Projects limit + Username). `jsonData.enableSecureSocksProxy`
(the legacy Secure Socks Proxy switch) is intentionally excluded per AGENTS.md.

## Conditional fields

None. Azure DevOps has a single authentication method (PAT); `jsonData.authType` is pinned
to `patToken` and the backend does not branch on it ("Not in use yet",
`pkg/plugin/settings.go:11`). There are no auth-gated or dependent fields.

---

## Verification

```
go generate ./registry/grafana-azuredevops-datasource/...   # clean
go test     ./registry/grafana-azuredevops-datasource/...   # PASS
```

`TestSchemaConformance` — 8/8 subtests PASS (BaseFieldsResolved, SchemaRoundTrip,
SchemaArtifactInSync, SchemaSpecHasNoSecureJSON, ConfigSchemaValid, JSONDataMatchesStruct,
JSONDataTypesMatchStruct, SecureValuesMatchLoadSettings). `TestLoadConfig` — all subtests
PASS. The regenerated `schema.gen.json` / `settings.gen.json` carry
`secureValues[patToken].required: true`, so committed artifacts are in sync.

## Files changed

- `dsconfig.json` — `secureJsonData_patToken`: `requiredWhen:"true"` → `required:true`.
- `schema.gen.json` — regenerated: `secureValues[patToken].required: true`.
- `settings.gen.json` — regenerated: `secureValues[patToken].required: true`.

No changes to `settings.go` / `settings.ts` / `README` / `conformance.go` / `plugin-ui`
were required.
