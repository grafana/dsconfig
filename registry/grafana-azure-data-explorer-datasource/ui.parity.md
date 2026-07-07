# Azure Data Explorer (grafana-azure-data-explorer-datasource) — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered from
[`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `grafana-azure-data-explorer-datasource`
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/<uid>` (`<uid>` == plugin id; provisioned/read-only, but every field renders).
- **New UI:** Storybook `configeditor-datasourceconfigwizard--tab`, `args=pluginType:grafana-azure-data-explorer-datasource` (the local `dsconfig.json` was served by intercepting the remote schema fetch with Playwright `context.route(...)`).
- **Method:** Both captured with Playwright. The new UI is a stepper — each group is clicked and its fields unioned.
- **Result:** **Parity already achieved — the new UI is a documented SUPERSET; no `dsconfig.json` changes required.**

---

## Findings (no fixes needed)

**No Custom HTTP Headers.** Azure Data Explorer is an Azure-AD/App-Registration datasource
(`@grafana/azure-sdk` credentials). Its legacy editor has no Custom HTTP Headers control
(`hasCustomHeaders:false`, `addHeaderBtn:false` in **both**). The **Add** button in the new-UI
capture is the **Allowed cookies** (`jsonData_keepCookies`) `TagsInput`/list — **not** headers —
exactly as with Azure Monitor. Headers are correctly **not** modeled.

**No `fileUpload`.** `fileInputs:0` in **both**. Unlike Azure Monitor, this plugin does **not**
expose the `clientcertificate` auth type — the only secret is the **Client Secret** (password
input). There are no certificate textareas or file pickers.

**No `required:true` fix.** The schema has **no** `requiredWhen`/`required:true` fields. The
credentials the legacy UI marks required (**Directory (tenant) ID \***, **Application (client)
ID \***, **Client Secret \***) are managed atomically inside the `@grafana/azure-sdk`
`AzureCredentialsForm` discriminated union (`jsonData_azureCredentials`), which enforces
requiredness per selected `authType` — not via dsconfig `requiredWhen`. Correct for the Azure SDK
pattern.

## New UI verification

Tab mode renders all ten groups (`groupTitles`): **Connection**, **Authentication**, **Query
Optimizations** (opt), **Database schema settings** (opt), **Application** (opt), **Tracking**
(opt), **Additional settings** (opt), **OpenAI (provisioning-only)** (opt), **Legacy top-level
credentials** (opt), **Deprecated / dead fields** (opt). `switches:4` (dynamic caching, managed
schema, user tracking, plus the `oauthPassThru` toggle). The legacy capture shows `switches:0`
only because its **Additional settings** section was collapsed (the "Expand section Additional
settings" button is present); those controls live behind it.

The legacy editor exposes: a **Configuration Help** docs block, **Authentication** (Azure Cloud,
Directory (tenant) ID, Application (client) ID, Client Secret), **Default cluster URL**, and a
collapsed **Additional settings** covering *query optimizations, schema settings, tracking, OpenAI,
request timeout, and forwarded cookies*. Every one of these is modeled below; the schema adds
clearly-labeled optional/provisioning-only/legacy/deprecated groups on top (the superset).

## Field-by-field parity (superset)

| Legacy field / area          | schema id                                     | Target           | Status                        |
| ---------------------------- | --------------------------------------------- | ---------------- | ----------------------------- |
| Authentication Method        | `jsonData_azureCredentials` (discriminator)   | `jsonData`       | ✅ 🔀 (azure-sdk union)       |
| Azure Cloud                  | in `jsonData_azureCredentials.azureCloud`     | `jsonData`       | ✅ (clientsecret)             |
| Directory (tenant) ID \*     | in `jsonData_azureCredentials.tenantId`       | `jsonData`       | ✅ 🔀 (clientsecret)          |
| Application (client) ID \*   | in `jsonData_azureCredentials.clientId`       | `jsonData`       | ✅ 🔀 (clientsecret)          |
| Client Secret \*             | `secureJsonData_azureClientSecret`            | `secureJsonData` | ✅ 🔀 (clientsecret)          |
| Default cluster URL (Opt)    | `jsonData_clusterUrl`                          | `jsonData`       | ✅ (Connection)               |
| Query timeout                | `jsonData_queryTimeout`                        | `jsonData`       | ✅ (Query Optimizations)      |
| Use dynamic caching          | `jsonData_dynamicCaching`                      | `jsonData`       | ✅ switch                     |
| Cache max age                | `jsonData_cacheMaxAge`                         | `jsonData`       | ✅                            |
| Data consistency             | `jsonData_dataConsistency`                     | `jsonData`       | ✅ select                     |
| Default editor mode          | `jsonData_defaultEditorMode`                   | `jsonData`       | ✅ select                     |
| Default database             | `jsonData_defaultDatabase`                     | `jsonData`       | ✅ (Database schema settings) |
| Use managed schema           | `jsonData_useSchemaMapping`                    | `jsonData`       | ✅ switch                     |
| Schema mappings              | `jsonData_schemaMappings`                      | `jsonData`       | ✅ 🔀 (gated by managed schema) |
| Application name (Opt)       | `jsonData_application`                         | `jsonData`       | ✅ (Application)              |
| Send username header to host | `jsonData_enableUserTracking`                  | `jsonData`       | ✅ switch (Tracking)          |
| Allowed cookies (forwarded)  | `jsonData_keepCookies`                         | `jsonData`       | ✅ list (Additional settings) |
| OpenAI API key               | `secureJsonData_OpenAIAPIKey`                  | `secureJsonData` | ✅ (provisioning-only group)  |

**Superset extras** (provisioning/legacy/deprecated, correctly labeled and optional): OAuth pass-through
(`jsonData_oauthPassThru`, required for `clientsecret-obo`), **Legacy top-level credentials**
(`jsonData_azureCloud` / `jsonData_onBehalfOf` / `jsonData_tenantId` / `jsonData_clientId` +
`secureJsonData_clientSecret` fallback), and **Deprecated / dead fields** (`jsonData_minimalCache`).

## Conditional fields & auth selector — tested

`jsonData_azureCredentials` is the discriminated-union **Authentication Method** written by
`@grafana/azure-sdk`. Auth types: `clientsecret` (App Registration + Client Secret — always
available), `msi` (Managed Identity), `workloadidentity`, `currentuser`, and `clientsecret-obo`
(On-Behalf-Of; requires `oauthPassThru == true` + the `adxOnBehalfOf` feature toggle). Azure Cloud
/ tenant / client / secret render for the client-secret types. **Schema mappings** are gated by
**Use managed schema**. All conditionals render per the schema.

---

## Verification

```
go test -count=1 ./registry/grafana-azure-data-explorer-datasource/...   # ok (TestSchemaConformance 8/8 subtests + load/validate suites PASS)
```

No schema edit was made, so no regeneration was needed; the committed artifacts remain in sync.

## Files changed

**None.** Azure Data Explorer was already at parity: HTTP headers n/a (the "Add" button is the
Allowed-cookies TagsInput), file upload n/a (no client-certificate auth; the secret is a password
input), and required credentials are handled by the azure-sdk credentials union. The new UI is a
documented superset — every legacy field is present.
