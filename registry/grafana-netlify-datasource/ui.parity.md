# Netlify (grafana-netlify-datasource) — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered from
[`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `grafana-netlify-datasource`
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/grafana-netlify-datasource` (provisioned → read-only "cannot be modified", but all fields still render).
- **New UI:** Storybook `configeditor-datasourceconfigwizard--tab`, `args=pluginType:grafana-netlify-datasource`; schema normally fetched from GitHub, the LOCAL `dsconfig.json` was served by intercepting that fetch with Playwright `context.route(...)`.
- **Method:** Both UIs captured with Playwright (labels, radios, switches, file inputs, "Add header" buttons, `bodyText`). The new UI is a stepper, so each group was clicked and fields unioned.
- **Framework:** openapi-framework — config under a service-keyed shape (`jsonData.services.Netlify.auth.id` discriminator — note the capitalized service id — flat dotted `secureJsonData` keys). Server URL is fixed at `https://api.netlify.com/api/v1` (no connection variables).
- **Result:** **Parity already achieved — no `dsconfig.json` changes required.**

## Findings (no fixes needed)

- **No Custom HTTP Headers.** Both captures report `hasCustomHeaders:false` and `addHeaderBtn:false`; neither `bodyText` contains an "Add header" / custom-headers section. Correctly not modeled.
- **No `fileUpload`.** `fileInputs:0` in both legacy and new UI; the personal access token is a text secret. Not used.
- **Required-field handling.** The token is conditionally `requiredWhen: jsonData_services_Netlify_auth_id == 'bearer_token'` (legacy `Token *`). New UI renders `Token*`. No connection/URL field exists (fixed base URL), matching the legacy editor, which shows no `URL` label.

## Auth selector

Single-option discriminator `jsonData_services_Netlify_auth_id` ("Personal access tokens" = `bearer_token`, the backend default). The token `dependsOn`/`requiredWhen` that value, so it is always revealed on the one auth path. The selector label + option and the token both appear in the new-UI `bodyText`; legacy shows the matching `Personal access tokens` heading + `Token *`. New-UI group observed: **Authentication**.

## Field-by-field

| Legacy field | schema id | target | status |
| --- | --- | --- | --- |
| Personal access tokens (auth selector) | `jsonData_services_Netlify_auth_id` | `jsonData` (`services.Netlify.auth.id`) | ✅ |
| Token * | `secureJsonData_Netlify_token` | `secureJsonData` (`Netlify.token`) | ✅ |

## Verification

```
go test ./registry/grafana-netlify-datasource/...   # TestSchemaConformance 8/8 subtests PASS (+ TestLoadConfig, TestValidate)
```

No schema edit → no regeneration; committed artifacts remain in sync (`SchemaArtifactInSync` passes); `consoleErrors:[]` in the new UI.

## Files changed

**None.** Validation-only report; Netlify was already at parity (headers n/a, fileUpload n/a, required handling correct, single-option auth selector verified).
