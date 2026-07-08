# Supabase (grafana-supabase-datasource) — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered from
[`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `grafana-supabase-datasource`
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/grafana-supabase-datasource` (provisioned → read-only "cannot be modified", but all fields still render).
- **New UI:** Storybook `configeditor-datasourceconfigwizard--tab`, `args=pluginType:grafana-supabase-datasource`; schema normally fetched from GitHub, the LOCAL `dsconfig.json` was served by intercepting that fetch with Playwright `context.route(...)`.
- **Method:** Both UIs captured with Playwright (labels, radios, switches, file inputs, "Add header" buttons, `bodyText`). The new UI is a stepper, so each group was clicked and fields unioned.
- **Framework:** openapi-framework — config under a service-keyed shape (`jsonData.services.mgmt.auth.id` discriminator, flat dotted `secureJsonData` keys). Server URL is fixed at `https://api.supabase.com` (no connection variables).
- **Result:** **Parity already achieved — no `dsconfig.json` changes required.**

## Findings (no fixes needed)

- **No Custom HTTP Headers.** Both captures report `hasCustomHeaders:false` and `addHeaderBtn:false`; neither `bodyText` contains an "Add header" / custom-headers section. Correctly not modeled.
- **No `fileUpload`.** `fileInputs:0` in both legacy and new UI; the personal token is a text secret. Not used.
- **Required-field handling.** The token is conditionally `requiredWhen: jsonData_services_mgmt_auth_id == 'mgmt_bearer'` (legacy `Token *`). New UI renders `Token*`. No connection/URL field exists (fixed base URL), matching the legacy editor, which shows no `URL` label.

## Auth selector

Single-option discriminator `jsonData_services_mgmt_auth_id` ("Supabase personal token" = `mgmt_bearer`, the backend default). The token `dependsOn`/`requiredWhen` that value, so it is always revealed on the one auth path. The selector label + option and the token both appear in the new-UI `bodyText`; legacy shows the matching `Supabase personal token` heading + `Token *`. New-UI group observed: **Authentication**.

## Field-by-field

| Legacy field | schema id | target | status |
| --- | --- | --- | --- |
| Supabase personal token (auth selector) | `jsonData_services_mgmt_auth_id` | `jsonData` (`services.mgmt.auth.id`) | ✅ |
| Token * | `secureJsonData_mgmt_token` | `secureJsonData` (`mgmt.token`) | ✅ |

## Verification

```
go test ./registry/grafana-supabase-datasource/...   # TestSchemaConformance 8/8 subtests PASS (+ TestLoadConfig, TestValidate)
```

No schema edit → no regeneration; committed artifacts remain in sync (`SchemaArtifactInSync` passes); `consoleErrors:[]` in the new UI.

## Files changed

**None.** Validation-only report; Supabase was already at parity (headers n/a, fileUpload n/a, required handling correct, single-option auth selector verified).
