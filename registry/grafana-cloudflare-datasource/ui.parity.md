# Cloudflare (grafana-cloudflare-datasource) — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered from
[`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `grafana-cloudflare-datasource`
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/grafana-cloudflare-datasource` (provisioned → read-only "cannot be modified", but all fields still render).
- **New UI:** Storybook `configeditor-datasourceconfigwizard--tab`, `args=pluginType:grafana-cloudflare-datasource`; schema normally fetched from GitHub, the LOCAL `dsconfig.json` was served by intercepting that fetch with Playwright `context.route(...)`.
- **Method:** Both UIs captured with Playwright (labels, radios, switches, file inputs, "Add header" buttons, `bodyText`). The new UI is a stepper, so each group was clicked and fields unioned.
- **Framework:** openapi-framework — config under a service-keyed shape (`jsonData.services.cloudflare.auth.id` discriminator, flat dotted `secureJsonData` keys). Server URL is fixed at `https://api.cloudflare.com/client/v4` (no connection variables).
- **Result:** **Parity already achieved — no `dsconfig.json` changes required.**

## Findings (no fixes needed)

- **No Custom HTTP Headers.** Both captures report `hasCustomHeaders:false` and `addHeaderBtn:false`; neither `bodyText` contains an "Add header" / custom-headers section. Correctly not modeled.
- **No `fileUpload`.** `fileInputs:0` in both legacy and new UI; the API token is a text secret. Not used.
- **Required-field handling.** The token is conditionally `requiredWhen: jsonData_services_cloudflare_auth_id == 'bearer_token'` (legacy `Token *`). New UI renders `Token*`. No connection/URL field exists (fixed base URL), matching the legacy editor, which shows no `URL` label.

## Auth selector

Single-option discriminator `jsonData_services_cloudflare_auth_id` ("API Key" = `bearer_token`, the backend default). The token `dependsOn`/`requiredWhen` that value, so it is always revealed on the one auth path. The selector label + option and the token both appear in the new-UI `bodyText`; legacy shows the matching `API Key` heading + `Token *`. New-UI group observed: **Authentication**.

## Field-by-field

| Legacy field | schema id | target | status |
| --- | --- | --- | --- |
| API Key (auth selector) | `jsonData_services_cloudflare_auth_id` | `jsonData` (`services.cloudflare.auth.id`) | ✅ |
| Token * | `secureJsonData_cloudflare_token` | `secureJsonData` (`cloudflare.token`) | ✅ |

## Verification

```
go test ./registry/grafana-cloudflare-datasource/...   # TestSchemaConformance 8/8 subtests PASS (+ TestLoadConfig, TestValidate)
```

No schema edit → no regeneration; committed artifacts remain in sync (`SchemaArtifactInSync` passes); `consoleErrors:[]` in the new UI.

## Files changed

**None.** Validation-only report; Cloudflare was already at parity (headers n/a, fileUpload n/a, required handling correct, single-option auth selector verified).
