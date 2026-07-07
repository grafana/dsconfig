# Salesforce (grafana-salesforce-datasource) — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered from
[`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `grafana-salesforce-datasource`
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/<uid>` (`<uid>` == plugin id; provisioned/read-only, but every field renders).
- **New UI:** Storybook `configeditor-datasourceconfigwizard--tab`, `args=pluginType:grafana-salesforce-datasource` (the local `dsconfig.json` was served by intercepting the remote schema fetch with Playwright `context.route(...)`).
- **Method:** Both captured with Playwright. The new UI is a stepper — each group is clicked and its fields unioned.
- **Result:** **Parity already achieved — no `dsconfig.json` changes required.**

## Findings (no fixes needed)

- **No Custom HTTP Headers.** Salesforce authenticates via OAuth2 (password grant or JWT bearer); the legacy editor has no generic HTTP-headers section (`hasCustomHeaders:false`, `addHeaderBtn:false` in **both**). Correctly not modeled.
- **No `fileUpload`.** `fileInputs:0` in **both**. The JWT **Certificate** and **Private Key** are `textarea` PEM inputs (not file pickers), and the credentials secrets are password inputs.
- **Required-field handling.** No unconditional `required:true`. Credentials-path fields (`jsonData_user`, `secureJsonData_password`, `secureJsonData_clientID`, `secureJsonData_clientSecret`) are `requiredWhen: "jsonData_authType == 'user'"`; the JWT-path `secureJsonData_cert`/`secureJsonData_privateKey` are `requiredWhen: "... == 'jwt'"`. The new UI marks required per the selected path; **Security Token** is optional in both.

## Conditional fields & auth selector — tested

Salesforce has an **Authentication radio** (`jsonData_authType`, `radios:2` in **both** UIs), defaulting to **Credentials** (`user`):

- **Credentials** (`user`, default) → **User Name**, **Password**, **Security Token** (optional), **Consumer Key**, **Consumer Secret**.
- **JWT** (`jwt`) → **Certificate** + **Private Key** (PEM textareas), plus the shared consumer key / username used to sign the assertion.

Both UIs open on **Credentials**, so the JWT textareas are conditional and not shown by default (`textareas:0` in both) — they are modeled and revealed on `jwt`. The **Environment** select (`jsonData_tokenUrl`: Production / SandBox) lives under an optional **Optional Settings** group. Legacy `jsonData_sandbox` is a deprecated boolean read-only fallback, correctly rendered by neither UI.

## Field-by-field parity

| Legacy field    | schema id                       | Target           | Status                     |
| --------------- | ------------------------------- | ---------------- | -------------------------- |
| Authentication  | `jsonData_authType`             | `jsonData`       | ✅ radio (discriminator)   |
| User Name       | `jsonData_user`                 | `jsonData`       | ✅ (user)                  |
| Password        | `secureJsonData_password`       | `secureJsonData` | ✅ (user)                  |
| Security Token  | `secureJsonData_securityToken`  | `secureJsonData` | ✅ optional (user)         |
| Consumer Key    | `secureJsonData_clientID`       | `secureJsonData` | ✅ (user)                  |
| Consumer Secret | `secureJsonData_clientSecret`   | `secureJsonData` | ✅ (user)                  |
| Certificate     | `secureJsonData_cert`           | `secureJsonData` | ✅ conditional textarea (jwt) |
| Private Key     | `secureJsonData_privateKey`     | `secureJsonData` | ✅ conditional textarea (jwt) |
| Environment     | `jsonData_tokenUrl`             | `jsonData`       | ✅ (Optional Settings)     |

Groups observed in the new UI: **Connection Settings**, **Optional Settings** (optional) — matching the legacy sections.

## Verification

```
go test -count=1 ./registry/grafana-salesforce-datasource/...   # ok (TestSchemaConformance 8/8 subtests + load/validate suites PASS)
```

No schema edit was made, so no regeneration was needed; the committed artifacts remain in sync.

## Files changed

**None.** Validation-only report; Salesforce was already at parity (headers n/a, fileUpload n/a — cert/key are textareas, conditional required correct, `radios:2` auth selector verified in both UIs).
