# Drone (grafana-drone-datasource) — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered from
[`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `grafana-drone-datasource`
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/grafana-drone-datasource` (provisioned → read-only "cannot be modified", but all fields still render).
- **New UI:** Storybook `configeditor-datasourceconfigwizard--tab`, `args=pluginType:grafana-drone-datasource`; schema normally fetched from GitHub, the LOCAL `dsconfig.json` was served by intercepting that fetch with Playwright `context.route(...)`.
- **Method:** Both UIs captured with Playwright (labels, radios, switches, file inputs, "Add header" buttons, `bodyText`). The new UI is a stepper, so each group was clicked and fields unioned.
- **Framework:** openapi-framework — config under a service-keyed shape (`jsonData.services.drone.auth.id` discriminator, `jsonData.variables.url` connection variable, flat dotted `secureJsonData` keys).
- **Result:** **Parity already achieved — no `dsconfig.json` changes required.**

## Findings (no fixes needed)

- **No Custom HTTP Headers.** Both captures report `hasCustomHeaders:false` and `addHeaderBtn:false`; neither `bodyText` contains an "Add header" / custom-headers section. Correctly not modeled.
- **No `fileUpload`.** `fileInputs:0` in both legacy and new UI; the API token is a text secret. Not used.
- **Required-field handling.** `url` is unconditionally `required:true` (legacy `URL *`); the token is conditionally `requiredWhen: jsonData_services_drone_auth_id == 'auth_bearer'` (legacy `Token *`). New UI renders both with the asterisk (`URL*`, `Token*`).

## Auth selector

Single-option discriminator `jsonData_services_drone_auth_id` ("Drone API token" = `auth_bearer`, the backend default). The token `dependsOn`/`requiredWhen` that value, so it is always revealed on the one auth path. The selector label + option and the token both appear in the new-UI `bodyText`; legacy shows the matching `Drone API token` heading + `Token *`. New-UI groups observed: **Connection**, **Authentication**.

## Field-by-field

| Legacy field | schema id | target | status |
| --- | --- | --- | --- |
| URL * | `jsonData_variables_url` | `jsonData` (`variables.url`) | ✅ |
| Drone API token (auth selector) | `jsonData_services_drone_auth_id` | `jsonData` (`services.drone.auth.id`) | ✅ |
| Token * | `secureJsonData_drone_token` | `secureJsonData` (`drone.token`) | ✅ |

## Verification

```
go test ./registry/grafana-drone-datasource/...   # TestSchemaConformance 8/8 subtests PASS (+ TestLoadConfig, TestValidate)
```

No schema edit → no regeneration; committed artifacts remain in sync (`SchemaArtifactInSync` passes); `consoleErrors:[]` in the new UI.

## Files changed

**None.** Validation-only report; Drone was already at parity (headers n/a, fileUpload n/a, required handling correct, single-option auth selector verified).
