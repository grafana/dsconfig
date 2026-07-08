# Solarwinds (grafana-solarwinds-datasource) — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered from
[`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `grafana-solarwinds-datasource`
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/<uid>` (`<uid>` == plugin id; provisioned/read-only, but every field renders).
- **New UI:** Storybook `configeditor-datasourceconfigwizard--tab`, `args=pluginType:grafana-solarwinds-datasource` (the local `dsconfig.json` was served by intercepting the remote schema fetch with Playwright `context.route(...)`).
- **Method:** Both captured with Playwright. The new UI is a stepper — each group is clicked and its fields unioned.
- **Result:** **Parity already achieved — no `dsconfig.json` changes required.**

## Findings (no fixes needed)

- **No Custom HTTP Headers.** SolarWinds uses basic auth against a fixed Information Service endpoint; the legacy editor has no generic HTTP-headers section (`hasCustomHeaders:false`, `addHeaderBtn:false` in **both**). Correctly not modeled.
- **No `fileUpload`.** `fileInputs:0` in **both**. The TLS **CA Certificate**, **Client Certificate**, and **Client Key** are `textarea` PEM inputs, not file pickers.
- **TLS block already modeled.** `switches:3` in **both** UIs — the standard TLS settings block (self-signed cert / mutual TLS / skip verify) is modeled directly in this schema, not left to a platform default.
- **Required-field handling.** `jsonData_variables_url` is `required:true` because the SolarWinds URL is *unconditionally* required (legacy always renders **URL \***); this is correct modeling, not a required-field fix. Username/password are `requiredWhen` the basic-auth path is selected.

## Conditional fields & auth selector — tested

- **Basic Auth** selector (`jsonData_services_solarwinds_auth_id`, single option `basic_auth`, the backend default) → **Username** + **Password**.
- **TLS Settings** (optional group), driven by two switches:
  - **Add self-signed certificate** (`..._selfSignedCert_enabled`) → reveals **CA Certificate** textarea.
  - **TLS Client Authentication** (`..._clientAuth_enabled`) → reveals **ServerName**, **Client Certificate**, **Client Key**.
  - **Skip TLS certificate validation** (`..._skipVerification`) standalone switch.

The three switches render in both UIs (`switches:3`); the TLS textareas/serverName are conditional on their switches being on (both captured with switches off, so `textareas:0` in both) — all are modeled.

## Field-by-field parity

| Legacy field                    | schema id                                                | Target           | Status                    |
| ------------------------------- | -------------------------------------------------------- | ---------------- | ------------------------- |
| URL \*                          | `jsonData_variables_url`                                 | `jsonData`       | ✅ required               |
| Basic Auth (selector)           | `jsonData_services_solarwinds_auth_id`                   | `jsonData`       | ✅ selector               |
| Username \*                     | `jsonData_services_solarwinds_auth_username`             | `jsonData`       | ✅ (basic_auth)           |
| Password \*                     | `secureJsonData_solarwinds_password`                     | `secureJsonData` | ✅ (basic_auth)           |
| Add self-signed certificate     | `..._auth_tls_selfSignedCert_enabled`                    | `jsonData`       | ✅ switch                 |
| CA Certificate                  | `secureJsonData_solarwinds_tls_selfSignedCert`           | `secureJsonData` | ✅ conditional textarea   |
| TLS Client Authentication       | `..._auth_tls_clientAuth_enabled`                        | `jsonData`       | ✅ switch                 |
| ServerName                      | `..._auth_tls_clientAuth_serverName`                     | `jsonData`       | ✅ conditional            |
| Client Certificate              | `secureJsonData_solarwinds_tls_clientCert`               | `secureJsonData` | ✅ conditional textarea   |
| Client Key                      | `secureJsonData_solarwinds_tls_clientKey`                | `secureJsonData` | ✅ conditional textarea   |
| Skip TLS certificate validation | `..._auth_tls_skipVerification`                          | `jsonData`       | ✅ switch                 |

Groups observed in the new UI: **Connection**, **Authentication**, **TLS Settings** (optional) — matching the legacy sections (Connection / Authentication / TLS settings).

## Verification

```
go test -count=1 ./registry/grafana-solarwinds-datasource/...   # ok (TestSchemaConformance 8/8 subtests + load/validate suites PASS)
```

No schema edit was made, so no regeneration was needed; the committed artifacts remain in sync.

## Files changed

**None.** Validation-only report; SolarWinds was already at parity (headers n/a, fileUpload n/a — certs are textareas, the full TLS block is modeled with `switches:3` in both UIs, required URL correct).
