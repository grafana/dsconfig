# GitHub — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered from
[`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `grafana-github-datasource`
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/<uid>` (Grafana Enterprise 13.0.1)
- **New UI:** Storybook `configeditor-datasourceconfigwizard` story, `pluginType:grafana-github-datasource` (local schema served via `context.route(...)` interception).
- **Result:** **Parity already achieved — no `dsconfig.json` changes required.**

## Findings (no fixes needed)

- **No Custom HTTP Headers.** The GitHub datasource authenticates to the GitHub API with a token / GitHub App; its legacy editor has no generic "HTTP headers" section (`hasCustomHeaders:false`, `addHeaderBtn:false`). New UI `hasHeadersEditor:false`. Correctly not modeled.
- **No `fileUpload`.** GitHub App private key is entered as text (secret); legacy `fileInputs:0`. Not used.
- **No `required:true` fix.** No unconditional `requiredWhen:"true"` fields — required credentials are conditionally required on the selected auth type.

## Conditional fields & effects — tested

GitHub has an **authentication-type selector** (`effects`): **Personal Access Token** vs **GitHub App**.

- **Personal Access Token** → reveals the Access Token secret.
- **GitHub App** → reveals App ID, Installation ID, and the App private key.
  The optional **Connection** group carries the GitHub Enterprise URL (`root.url`) for GHE. All conditionals reveal per the schema and match the legacy editor. New-UI sections observed: **Connection** (optional), **Authentication**.

## Verification

```
go test ./registry/grafana-github-datasource/...   # 8/8 conformance subtests PASS
```

No schema edit → no regeneration; committed artifacts remain in sync; full suite passes.

## Files changed

**None.** Validation-only report; GitHub was already at parity (headers n/a, fileUpload n/a, conditional required correct, auth-selector effects verified).
