# Atlassian Statuspage (grafana-atlassianstatuspage-datasource) — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered from
[`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `grafana-atlassianstatuspage-datasource`
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/<uid>` (`<uid>` == plugin id; provisioned/read-only, but every field renders).
- **New UI:** Storybook `configeditor-datasourceconfigwizard--tab`, `args=pluginType:grafana-atlassianstatuspage-datasource` (local schema served by intercepting the remote fetch with Playwright `context.route(...)`).
- **Method:** Both captured with Playwright. The new UI is a stepper — each group is clicked and its fields unioned.
- **Result:** **Parity already achieved — no `dsconfig.json` changes required.**

This is an **openapi-framework** datasource: connection under `jsonData.variables.<name>`. It queries the public Statuspage API, so there is **no authentication** and no `secureJsonData`.

## Findings (no fixes needed)

- **No Custom HTTP Headers.** `hasCustomHeaders:false` and `addHeaderBtn:false` in **both** UIs. Correctly not modeled.
- **No `fileUpload`.** `fileInputs:0` in **both**. Not used.
- **Required-field handling.** Legacy shows `URL *`; new UI shows `URL*`. Modeled as `required:true`. Matches.

## Auth note

There are **no auth fields** in either UI — the datasource is unauthenticated. `switches:0`, `radios:0`, `selects:0` in both, and no `Reveal` button in the new UI (no secret). The legacy `Server configuration` subheading carries no extra fields (base URL is derived from the `url` variable as `{url}/api/v2`), matching the single-field new Connection group.

## Field-by-field

| Legacy field | Schema id | Target | Parity |
| --- | --- | --- | --- |
| URL * | `jsonData_variables_url` (`url`, role `endpoint.baseUrl`) | `jsonData` | ✅ |

## Verification

```
go test ./registry/grafana-atlassianstatuspage-datasource/...   # ok — TestSchemaConformance 7/7 subtests PASS
```

(7 subtests: no `SchemaRoundTrip` step because there are no secure keys.) Also PASS: settings suite (`TestLoadConfig` 3). No schema edit → no regeneration; committed artifacts remain in sync.

## Files changed

**None.** Validation-only report; Atlassian Statuspage was already at parity (headers n/a, fileUpload n/a, no-auth single-URL config correct).
