# AstraDB (grafana-astradb-datasource) — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered from
[`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `grafana-astradb-datasource`
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/<uid>` (`<uid>` == plugin id; provisioned/read-only, but every field renders).
- **New UI:** Storybook `configeditor-datasourceconfigwizard--tab`, `args=pluginType:grafana-astradb-datasource` (local schema served by intercepting the remote fetch with Playwright `context.route(...)`).
- **Method:** Both captured with Playwright. The new UI is a stepper — each group is clicked and its fields unioned.
- **Result:** **Parity already achieved — no `dsconfig.json` changes required.**

AstraDB uses a **bespoke plugin editor** (not the openapi-framework generic one): a numeric `authKind` discriminator (`0`=Token / `1`=Credentials) drives two mutually-exclusive field sets.

## Findings (no fixes needed)

- **No Custom HTTP Headers.** `hasCustomHeaders:false` and `addHeaderBtn:false` in **both** UIs. Correctly not modeled.
- **No `fileUpload`.** Token and password are text secrets; `fileInputs:0` in **both**. Not used.
- **Required-field handling.** The bespoke legacy editor renders no `*` markers (labels `URI`, `Token`); the new UI shows `URI*` / `Token*`. Both are genuinely required — the health check returns `Invalid AstraDB URL` / `Invalid AstraDB Token` if empty — via `requiredWhen: jsonData_authKind == 0`. Minor cosmetic difference only.

## Auth selector (radio) & conditional fields

Both UIs render a 2-option **radio** (`radios:2` in both): **Token** / **Credentials**, from `jsonData_authKind`. The capture is in the default **Token** mode (`authKind == 0`), so both UIs show only the Token-path fields (URI, Token); the Credentials-path fields (GRPC Endpoint, Auth Endpoint, User Name, Password, Secure) are modeled but conditionally hidden identically in **both** (gated by `authKind == 1`). This matches.

## Field-by-field

| Legacy field | Schema id | Target | Parity |
| --- | --- | --- | --- |
| Authentication (Token/Credentials radio) | `jsonData_authKind` (discriminator) | `jsonData` | ✅ |
| URI (Token path) | `jsonData_uri` (`uri`, `authKind==0`) | `jsonData` | ✅ |
| Token (secret, Token path) | `secureJsonData_token` (`token`, `authKind==0`) | `secureJsonData` | ✅ |
| GRPC Endpoint (Credentials path) | `jsonData_grpcEndpoint` (`authKind==1`) | `jsonData` | ✅ |
| Auth Endpoint (Credentials path) | `jsonData_authEndpoint` (`authKind==1`) | `jsonData` | ✅ |
| User Name (Credentials path) | `jsonData_user` (`authKind==1`) | `jsonData` | ✅ |
| Password (secret, Credentials path) | `secureJsonData_password` (`authKind==1`) | `secureJsonData` | ✅ |
| Secure (TLS toggle, Credentials path) | `jsonData_secure` (checkbox, `authKind==1`) | `jsonData` | ✅ |

## Verification

```
go test ./registry/grafana-astradb-datasource/...   # ok — TestSchemaConformance 8/8 subtests PASS
```

Also PASS: settings suite (`TestLoadConfig` 13, `TestApplyDefaults` 2, `TestValidate` 7 — covering both Token and Credentials modes). No schema edit → no regeneration; committed artifacts remain in sync.

## Files changed

**None.** Validation-only report; AstraDB was already at parity (headers n/a, fileUpload n/a, 2-mode radio + conditional fields modeled and revealed identically).
