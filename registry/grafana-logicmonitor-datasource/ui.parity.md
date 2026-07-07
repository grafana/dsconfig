# LogicMonitor Devices (grafana-logicmonitor-datasource) — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered from
[`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `grafana-logicmonitor-datasource`
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/<uid>` (`<uid>` == plugin id; provisioned/read-only, but every field renders).
- **New UI:** Storybook `configeditor-datasourceconfigwizard--tab`, `args=pluginType:grafana-logicmonitor-datasource` (local schema served by intercepting the remote fetch with Playwright `context.route(...)`).
- **Method:** Both captured with Playwright. The new UI is a stepper — each group is clicked and its fields unioned.
- **Result:** **Parity already achieved — no `dsconfig.json` changes required.**

This is an **openapi-framework** datasource: `jsonData.services.logicmonitor.auth.*`, `jsonData.variables.<name>`, flat dotted `secureJsonData` keys.

## Findings (no fixes needed)

- **No Custom HTTP Headers.** `hasCustomHeaders:false` and `addHeaderBtn:false` in **both** UIs. Correctly not modeled.
- **No `fileUpload`.** The token is a text secret; `fileInputs:0` in **both**. Not used.
- **Required-field handling.** Legacy shows `Account Name *` and `Token *`; new UI shows `Account Name*` and `Token*`. Account Name is `required:true`; the token is `requiredWhen: ...auth_id == 'auth_bearer'`. Matches.

## Auth note

Single fixed auth method — LogicMonitor REST API v3 bearer token. The discriminator `jsonData_services_logicmonitor_auth_id` (default `auth_bearer`) renders as the `API v3 Key` selector in the new UI; the legacy editor surfaces the same method as its `API v3 Key` heading. The token is revealed on `auth_bearer` in both. The legacy `Server configuration` subheading carries no extra fields (server URL is derived from `account_name`), matching the new Connection group.

## Field-by-field

| Legacy field | Schema id | Target | Parity |
| --- | --- | --- | --- |
| Account Name * | `jsonData_variables_account_name` (`account_name`) | `jsonData` | ✅ |
| API v3 Key (auth method) | `jsonData_services_logicmonitor_auth_id` (discriminator) | `jsonData` | ✅ |
| Token * (secret) | `secureJsonData_logicmonitor_token` (`logicmonitor.token`) | `secureJsonData` | ✅ |

## Verification

```
go test ./registry/grafana-logicmonitor-datasource/...   # ok — TestSchemaConformance 8/8 subtests PASS
```

Also PASS: settings suite (`TestLoadConfig` 3, `TestValidate`). No schema edit → no regeneration; committed artifacts remain in sync.

## Files changed

**None.** Validation-only report; LogicMonitor was already at parity (headers n/a, fileUpload n/a, single-method auth + required fields correct).
