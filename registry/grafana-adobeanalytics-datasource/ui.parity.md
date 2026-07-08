# Adobe Analytics (grafana-adobeanalytics-datasource) — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered from
[`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `grafana-adobeanalytics-datasource`
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/<uid>` (`<uid>` == plugin id; provisioned/read-only, but every field renders).
- **New UI:** Storybook `configeditor-datasourceconfigwizard--tab`, `args=pluginType:grafana-adobeanalytics-datasource` (local schema served by intercepting the remote fetch with Playwright `context.route(...)`).
- **Method:** Both captured with Playwright. The new UI is a stepper — each group is clicked and its fields unioned.
- **Result:** **Parity already achieved — no `dsconfig.json` changes required.**

This is an **openapi-framework** datasource: `jsonData.services.adobe_analytics.auth.*`, `jsonData.variables.<name>`, flat dotted `secureJsonData` keys.

## Findings (no fixes needed)

- **No Custom HTTP Headers.** `hasCustomHeaders:false` and `addHeaderBtn:false` in **both** UIs. Correctly not modeled.
- **No `fileUpload`.** The client secret is a text secret; `fileInputs:0` in **both**. Not used.
- **Required-field handling.** Legacy shows `Global Company ID *`, `Client ID *`, `Client Secret *`; new UI shows the same three with `*`. Global Company ID is `required:true`; Client ID and Client Secret are `requiredWhen: ...auth_id == 'oauth2_m2m'`. Matches.

## Auth note

Single fixed auth method — OAuth2 server-to-server (client credentials). The discriminator `jsonData_services_adobe_analytics_auth_id` (default `oauth2_m2m`) renders as the `OAuth server to server authentication` selector in the new UI; the legacy editor surfaces the same method as its heading. Client ID + Client Secret reveal on `oauth2_m2m` in both. The legacy `Server configuration` subheading carries no extra fields (base URL derived from `global_company_id`), matching the new Connection group.

## Field-by-field

| Legacy field | Schema id | Target | Parity |
| --- | --- | --- | --- |
| Global Company ID * | `jsonData_variables_global_company_id` (`global_company_id`) | `jsonData` | ✅ |
| OAuth server to server (auth method) | `jsonData_services_adobe_analytics_auth_id` (discriminator) | `jsonData` | ✅ |
| Client ID * | `jsonData_services_adobe_analytics_auth_clientId` (`clientId`) | `jsonData` | ✅ |
| Client Secret * (secret) | `secureJsonData_adobe_analytics_clientSecret` (`adobe_analytics.clientSecret`) | `secureJsonData` | ✅ |

## Verification

```
go test ./registry/grafana-adobeanalytics-datasource/...   # ok — TestSchemaConformance 8/8 subtests PASS
```

Also PASS: settings suite (`TestLoadConfig` 3, `TestValidate`). No schema edit → no regeneration; committed artifacts remain in sync.

## Files changed

**None.** Validation-only report; Adobe Analytics was already at parity (headers n/a, fileUpload n/a, OAuth2 fields + required handling correct).
