# Vercel (grafana-vercel-datasource) — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered from
[`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `grafana-vercel-datasource`
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/grafana-vercel-datasource` (provisioned → read-only "cannot be modified", but all fields still render).
- **New UI:** Storybook `configeditor-datasourceconfigwizard--tab`, `args=pluginType:grafana-vercel-datasource`; schema normally fetched from GitHub, the LOCAL `dsconfig.json` was served by intercepting that fetch with Playwright `context.route(...)`.
- **Method:** Both UIs captured with Playwright (labels, radios, switches, file inputs, "Add header" buttons, `bodyText`). The new UI is a stepper, so each group was clicked and fields unioned.
- **Framework:** openapi-framework — config under a service-keyed shape (`jsonData.services.vercel.auth.id` discriminator, `jsonData.variables.team_id` connection variable, flat dotted `secureJsonData` keys). Server URL is fixed at `https://api.vercel.com`.
- **Result:** **Parity achieved (functionally) — no `dsconfig.json` changes required.** One intentional, documented **cosmetic** divergence on the `Team ID` required marker (below); not a gap.

## Documented cosmetic divergence — `Team ID` asterisk

The legacy editor renders **`Team ID *`** because the openapi framework marks all connection variables as required. In the schema, `jsonData_variables_team_id` is modeled **optional** (no `required:true`) — the backend/docs treat it as optional (only needed for team-scoped tokens; account-scoped tokens need no team). See schema `instructions`: *"jsonData.variables.team_id is optional and only needed for team-scoped tokens."* Keeping it optional preserves provisioning correctness for account-scoped tokens, so the new UI renders **`Team ID`** without the asterisk. The legacy `*` is cosmetic; the field itself is present in both UIs. This is a deliberate divergence, not a missing field. (Verified by test `explicit_config_loads_without_team`.)

## Findings (no fixes needed)

- **No Custom HTTP Headers.** Both captures report `hasCustomHeaders:false` and `addHeaderBtn:false`; neither `bodyText` contains an "Add header" / custom-headers section. Correctly not modeled.
- **No `fileUpload`.** `fileInputs:0` in both legacy and new UI; the access token is a text secret. Not used.
- **Required-field handling.** The token is conditionally `requiredWhen: jsonData_services_vercel_auth_id == 'vercelApiKey'` (legacy `Token *`, new UI `Token*`). `team_id` is intentionally optional (see above).

## Auth selector

Single-option discriminator `jsonData_services_vercel_auth_id` ("Access Token" = `vercelApiKey`, the backend default). The token `dependsOn`/`requiredWhen` that value, so it is always revealed on the one auth path. The selector label + option and the token both appear in the new-UI `bodyText`; legacy shows the matching `Access Token` heading + `Token *`. New-UI groups observed: **Connection**, **Authentication**.

## Field-by-field

| Legacy field | schema id | target | status |
| --- | --- | --- | --- |
| Team ID * | `jsonData_variables_team_id` | `jsonData` (`variables.team_id`) | ✅ modeled optional (legacy `*` cosmetic) |
| Access Token (auth selector) | `jsonData_services_vercel_auth_id` | `jsonData` (`services.vercel.auth.id`) | ✅ |
| Token * | `secureJsonData_vercel_token` | `secureJsonData` (`vercel.token`) | ✅ |

## Verification

```
go test ./registry/grafana-vercel-datasource/...   # TestSchemaConformance 8/8 subtests PASS (+ TestLoadConfig, TestApplyDefaultsAndValidate)
```

No schema edit → no regeneration; committed artifacts remain in sync (`SchemaArtifactInSync` passes); `consoleErrors:[]` in the new UI.

## Files changed

**None.** Validation-only report; Vercel fields all render in both UIs. The only difference is the intentional, documented `Team ID` optional/asterisk cosmetic divergence — headers n/a, fileUpload n/a, single-option auth selector verified.
