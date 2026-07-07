# PagerDuty (grafana-pagerduty-datasource) — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered from
[`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `grafana-pagerduty-datasource`
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/<uid>` (`<uid>` == plugin id; provisioned/read-only, but every field renders).
- **New UI:** Storybook `configeditor-datasourceconfigwizard--tab`, `args=pluginType:grafana-pagerduty-datasource` (local schema served by intercepting the remote fetch with Playwright `context.route(...)`).
- **Method:** Both captured with Playwright. The new UI is a stepper — each group is clicked and its fields unioned.
- **Result:** **Parity already achieved — no `dsconfig.json` changes required.**

## Findings (no fixes needed)

- **No Custom HTTP Headers.** `hasCustomHeaders:false` and `addHeaderBtn:false` in **both** UIs. Correctly not modeled.
- **No `fileUpload`.** The API key is a text secret; `fileInputs:0` in **both**. Not used.
- **Required-field handling.** Legacy label is `API key` (no asterisk); new UI shows `API key*`. The secret is effectively always required — the health check (`GET /incidents`) returns 401 without it — enforced via `requiredWhen: jsonData_authId == 'api_key'`. Minor cosmetic difference only.

## Empty "Optional Settings" — parity, not a gap

The legacy editor shows an **empty** `Optional Settings` collapsible: PagerDuty has a fixed base URL (`https://api.pagerduty.com`, single OpenAPI server, no server variables), so there is no server/connection config and no fields render under it. The new UI correctly renders no such empty section. This is parity.

## Auth note

Single fixed OpenAPI `apiKey` scheme. The discriminator `jsonData_authId` (default `api_key`, allowedValues `['api_key']`) is tagged `frontend-managed` and is surfaced in **neither** UI — the editor auto-selects it. No no-auth option exists.

## Field-by-field

| Legacy field | Schema id | Target | Parity |
| --- | --- | --- | --- |
| API key (secret) | `secureJsonData_authApiKeyApiKey` (`auth.api_key.apiKey`) | `secureJsonData` | ✅ |
| Auth scheme (not shown) | `jsonData_authId` (discriminator, frontend-managed) | `jsonData` | ✅ n/a |
| Optional Settings (empty) | — (fixed base URL, no fields) | — | ✅ |

## Verification

```
go test ./registry/grafana-pagerduty-datasource/...   # ok — TestSchemaConformance 8/8 subtests PASS
```

Also PASS: settings suite (`TestLoadConfig` 7, `TestApplyDefaults` 3, `TestValidate` 4). No schema edit → no regeneration; committed artifacts remain in sync.

## Files changed

**None.** Validation-only report; PagerDuty was already at parity (headers n/a, fileUpload n/a, empty Optional Settings correctly omitted, conditional required correct).
