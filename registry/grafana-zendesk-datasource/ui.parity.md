# Zendesk (grafana-zendesk-datasource) — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered from
[`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `grafana-zendesk-datasource`
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/<uid>` (`<uid>` == plugin id; provisioned/read-only, but every field renders).
- **New UI:** Storybook `configeditor-datasourceconfigwizard--tab`, `args=pluginType:grafana-zendesk-datasource` (local schema served by intercepting the remote fetch with Playwright `context.route(...)`).
- **Method:** Both captured with Playwright. The new UI is a stepper — each group is clicked and its fields unioned.
- **Result:** **Parity already achieved — no `dsconfig.json` changes required.**

This is an **openapi-framework** datasource: `jsonData.services.zendesk.auth.*`, `jsonData.variables.<name>`, flat dotted `secureJsonData` keys.

## Findings (no fixes needed)

- **No Custom HTTP Headers.** `hasCustomHeaders:false` and `addHeaderBtn:false` in **both** UIs. Correctly not modeled.
- **No `fileUpload`.** The API token is a text secret; `fileInputs:0` in **both**. Not used.
- **Required-field handling.** Legacy shows `Subdomain *`, `Email *`, `API Token *`; new UI shows the same three with `*`. Subdomain is `required:true`; Email and API Token are `requiredWhen: ...auth_id == 'basic_auth'`. Matches.

## Auth note

Single fixed auth method — basic auth (email + API token). The discriminator `jsonData_services_zendesk_auth_id` (default `basic_auth`) renders as the `Basic Auth` selector in the new UI; the legacy editor surfaces the same method as its `Basic authentication` heading. Email + API Token reveal on `basic_auth` in both. The legacy `Server configuration` subheading carries no extra fields (base URL derived from `subdomain`), matching the new Connection group.

## Field-by-field

| Legacy field | Schema id | Target | Parity |
| --- | --- | --- | --- |
| Subdomain * | `jsonData_variables_subdomain` (`subdomain`) | `jsonData` | ✅ |
| Basic Auth (auth method) | `jsonData_services_zendesk_auth_id` (discriminator) | `jsonData` | ✅ |
| Email * | `jsonData_services_zendesk_auth_username` (`username`) | `jsonData` | ✅ |
| API Token * (secret) | `secureJsonData_zendesk_password` (`zendesk.password`) | `secureJsonData` | ✅ |

## Verification

```
go test ./registry/grafana-zendesk-datasource/...   # ok — TestSchemaConformance 8/8 subtests PASS
```

Also PASS: settings suite (`TestLoadConfig` 5, `TestApplyDefaults` 2, `TestValidate` 6). No schema edit → no regeneration; committed artifacts remain in sync.

## Files changed

**None.** Validation-only report; Zendesk was already at parity (headers n/a, fileUpload n/a, basic-auth fields + required handling correct).
