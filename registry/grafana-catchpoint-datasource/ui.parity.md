# Catchpoint (grafana-catchpoint-datasource) — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered from
[`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `grafana-catchpoint-datasource`
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/grafana-catchpoint-datasource` (provisioned → read-only "cannot be modified", but all fields still render).
- **New UI:** Storybook `configeditor-datasourceconfigwizard--tab`, `args=pluginType:grafana-catchpoint-datasource`; schema normally fetched from GitHub, the LOCAL `dsconfig.json` was served by intercepting that fetch with Playwright `context.route(...)`.
- **Method:** Both UIs captured with Playwright (labels, radios, switches, file inputs, "Add header" buttons, `bodyText`). The new UI is a stepper, so each group was clicked and fields unioned.
- **Framework:** openapi-framework — config under a service-keyed shape (`jsonData.services.catchpoint.auth.id` discriminator, flat dotted `secureJsonData` keys). Server URL is fixed at `https://io.catchpoint.com/api/v2` (no connection variables).
- **Result:** **Parity already achieved — no `dsconfig.json` changes required.**

## Findings (no fixes needed)

- **No Custom HTTP Headers.** Both captures report `hasCustomHeaders:false` and `addHeaderBtn:false`; neither `bodyText` contains an "Add header" / custom-headers section. Correctly not modeled.
- **No `fileUpload`.** `fileInputs:0` in both legacy and new UI; the API v2 key is a text secret. Not used.
- **Required-field handling.** The key is conditionally `requiredWhen: jsonData_services_catchpoint_auth_id == 'bearer_token'` (legacy `Token *`). New UI renders `Token*`. No connection/URL field exists (fixed base URL), matching the legacy editor, which shows no `URL` label.

## Auth selector

Single-option discriminator `jsonData_services_catchpoint_auth_id` ("API v2 Key" = `bearer_token`, the backend default). The token `dependsOn`/`requiredWhen` that value, so it is always revealed on the one auth path. The selector label + option and the token both appear in the new-UI `bodyText`; legacy shows the matching `API v2 Key` heading + `Token *`. New-UI group observed: **Authentication**.

## Field-by-field

| Legacy field | schema id | target | status |
| --- | --- | --- | --- |
| API v2 Key (auth selector) | `jsonData_services_catchpoint_auth_id` | `jsonData` (`services.catchpoint.auth.id`) | ✅ |
| Token * | `secureJsonData_catchpoint_token` | `secureJsonData` (`catchpoint.token`) | ✅ |

## Verification

```
go test ./registry/grafana-catchpoint-datasource/...   # TestSchemaConformance 8/8 subtests PASS (+ TestLoadConfig, TestValidate)
```

No schema edit → no regeneration; committed artifacts remain in sync (`SchemaArtifactInSync` passes); `consoleErrors:[]` in the new UI.

## Files changed

**None.** Validation-only report; Catchpoint was already at parity (headers n/a, fileUpload n/a, required handling correct, single-option auth selector verified).
