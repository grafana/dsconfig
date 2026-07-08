# Datadog — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered from
[`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `grafana-datadog-datasource`
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/bfrbqiwz1i9z4d` (Grafana Enterprise 13.0.1)
- **New UI:** `http://192.168.1.241:58899/iframe.html?args=pluginType:grafana-datadog-datasource&id=configeditor-datasourceconfigwizard--tab` (and `--wizard`)
- **Method:** Playwright captured the legacy UI (`legacy-expand-datadog-verify.png`) and drove the new UI with the **local** `dsconfig.json` served via `context.route(...)`.
- **Result:** **One `dsconfig.json` fix applied** — `jsonData_url` promoted from a literal-`true` `requiredWhen` to a proper `required: true`. HTTP headers and file upload are correctly not modeled (legacy has neither).

---

## Fix applied

**`required:true` on `jsonData_url`.** The field carried `"requiredWhen": "true"` — a
literal-`true` CEL expression that generated `x-dsconfig-required-when: "true"` in the
committed artifacts instead of a real JSON-schema requirement. Since the URL is required
in **every** mode (the backend health check fails with `Enter a URL.`), this is an
unconditional requirement and is now modeled as `"required": true`. After regeneration,
`schema.gen.json` / `settings.gen.json` emit `"jsonData": { "required": ["url"], ... }`.

The four **real conditional** `requiredWhen` expressions were left untouched:

| Field                            | requiredWhen (kept)                        |
| -------------------------------- | ------------------------------------------ |
| `secureJsonData_apiKey`          | `jsonData_pluginMode != 'hosted-metrics'`  |
| `secureJsonData_appKey`          | `jsonData_pluginMode != 'hosted-metrics'`  |
| `root_basicAuthUser`             | `jsonData_pluginMode == 'hosted-metrics'`  |
| `secureJsonData_basicAuthPassword` | `jsonData_pluginMode == 'hosted-metrics'` |

## Findings

**No Custom HTTP Headers.** Datadog authenticates with API/App keys (Default mode) or
Grafana Cloud basic auth (Hosted Metrics mode). Its legacy editor has **no** Custom HTTP
Headers section (`hasCustomHeaders:false`, `addHeaderBtn:false`). Headers are correctly
**not** modeled.

**No `fileUpload`.** All credentials are text inputs; there is **no** file-upload control
in the legacy editor (`fileInputs:0`, `uploadButtons:[]`). `fileUpload` is correctly
**not** used.

## New UI verification

Tab mode renders all sections: **Connection**, **Authentication**, **Additional settings**
(opt). `hasHeadersEditor:false` (correct). `urlPresent:true` and **API URL / Region** now
displays the required `*` marker in both tab mode and wizard **General (1/4)** step.

## Field-by-field parity (highlights)

| Legacy field                         | schema id                          | Target                  | Status          |
| ------------------------------------ | ---------------------------------- | ----------------------- | --------------- |
| Mode (Default / Hosted Datadog Metrics) | `virtual_mode` (discriminator)  | jsonData/root (effects) | ✅ 🔀           |
| API URL / Region                     | `jsonData_url`                     | `jsonData`              | ✅ (now `required`) |
| API key                              | `secureJsonData_apiKey`           | `secureJsonData`        | ✅ 🔀 (default)  |
| App key                              | `secureJsonData_appKey`           | `secureJsonData`        | ✅ 🔀 (default)  |
| User                                 | `root_basicAuthUser`              | `root`                  | ✅ 🔀 (hosted)   |
| Password                             | `secureJsonData_basicAuthPassword` | `secureJsonData`      | ✅ 🔀 (hosted)   |
| Show API rate limits                 | `jsonData_logApiRateLimits`       | `jsonData`              | ✅ (opt)         |
| Enable API rate limit threshold      | `jsonData_rateLimitEnabled`       | `jsonData`              | ✅ (opt)         |
| API rate limit threshold %           | `jsonData_rateLimitMetrics`       | `jsonData`              | ✅ 🔀 (opt)      |
| Disable data links                   | `jsonData_disableDataLinks`       | `jsonData`              | ✅ (opt)         |
| Response Size                        | `jsonData_size`                   | `jsonData`              | ✅ (opt)         |

## Conditional fields & effects — tested

The **Mode** selector (`virtual_mode`) is a discriminator with an `effects` block. Playwright
drove the radio and captured the Save & Test payload in both modes:

- **Default** → payload `jsonData.pluginMode="default"`, `basicAuth=false`, credentials in
  `secureJsonData.apiKey`/`appKey`. UI shows API key / App key (required); User / Password
  hidden.
- **Hosted Datadog Metrics** → payload `jsonData.pluginMode="hosted-metrics"`, `basicAuth=true`,
  `basicAuthUser` set, `secureJsonData.basicAuthPassword` set. UI shows User / Password;
  API key / App key hidden; the URL **override** swaps the placeholder to the hosted-metrics
  proxy (`https://dd-prod-10-prod-us-central-0.grafana.net/datadog`).

Both `effects` branches write the correct storage fields, `dependsOn` correctly reveals/hides
each mode's credentials, and the `overrides` URL placeholder fires — the `effects` field
renders and works end-to-end.

---

## Verification

```
go generate ./registry/grafana-datadog-datasource/...   # regenerated schema.gen.json + settings.gen.json
go test     ./registry/grafana-datadog-datasource/...    # PASS
```

`TestSchemaConformance` 8/8 subtests PASS (incl. `SchemaArtifactInSync`), plus
`TestLoadConfig`, `TestValidate`, `TestApplyDefaults`, `TestSettingsExamples` (subtests such
as `default mode missing url` and `hosted_metrics_with_default_url_errors` confirm the URL is
genuinely required).

## Out-of-scope observations (not addressed — outside the required-field mandate)

- **TLS settings.** Legacy shows a *TLS settings* section (Add self-signed certificate / TLS
  Client Authentication / Skip TLS certificate validation). These are not modeled in the
  schema; adding them would require new fields beyond this required-field fix.
- **URL component.** Legacy renders *API URL / Region* as a region **dropdown** (US1/US3/US5/
  EU/US1-FED) in Default mode; the schema uses a text `input` with the API URL as default.
  Pre-existing component choice, unchanged.
- **Response Size** shows a `*` in legacy but sits in the optional *Additional settings* group
  in the schema (defaultValue `100`, never empty). Left as-is per scope.

## Files changed

- `dsconfig.json` — `jsonData_url`: `"requiredWhen": "true"` → `"required": true` (1 line).
- `schema.gen.json`, `settings.gen.json` — regenerated (`jsonData.required: ["url"]`; removed
  `x-dsconfig-required-when` on `url`).

No `settings.go` / `settings.ts` / `conformance.go` / `plugin-ui` changes were required.
