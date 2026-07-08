# Oracle Database (grafana-oracle-datasource) — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered from
[`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `grafana-oracle-datasource`
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/<uid>` (`<uid>` == plugin id; provisioned/read-only, but every field renders).
- **New UI:** Storybook `configeditor-datasourceconfigwizard--tab`, `args=pluginType:grafana-oracle-datasource` (the local `dsconfig.json` was served by intercepting the remote schema fetch with Playwright `context.route(...)`).
- **Method:** Both captured with Playwright. The new UI is a stepper — each group is clicked and its fields unioned.
- **Result:** **Parity already achieved — no `dsconfig.json` changes required.**

## Findings (no fixes needed)

- **No Custom HTTP Headers.** Oracle is a native SQL driver datasource; its legacy editor has no generic HTTP-headers section (`hasCustomHeaders:false`, `addHeaderBtn:false` in **both** UIs). Correctly not modeled.
- **No `fileUpload`.** `fileInputs:0` in **both**. The password is a secret text input; there are no certificate uploads.
- **Required-field handling (minor cosmetic difference, new UI is more correct).** The legacy editor marks **User \*** required but renders **Password** *without* an asterisk. The new UI marks **both** required — both `jsonData_user` and `secureJsonData_password` carry `requiredWhen: "jsonData_useKerberosAuthentication != true"` (i.e. required on the Basic auth path). This makes the new UI the more consistent/correct of the two. No unconditional `required:true` on credentials.

## Conditional fields & selectors — tested

Oracle has **two independent virtual selectors**, both defaulting to their first option in each UI:

- **Connection methods** (`virtual_connectionType`, computed from `jsonData.useTNSNamesBasedConnection`): **Host with TCP Port** (`tcp`, default) reveals **Host** + **Database**; **TNSNames Entry** (`tns`) reveals **TNSName** (`jsonData_tnsNamesEntry`). TNSName is conditional and not shown at the default `tcp` — it is modeled.
- **Oracle authentication** (`virtual_authType`, computed from `jsonData.useKerberosAuthentication`): **Basic Authentication** (default) reveals **User** + **Password**; **Kerberos** (TNSNames-only, no user/password). Kerberos is conditional and modeled. Both UIs open on `tcp` + `basic`, so both capture the same visible set.

## Field-by-field parity

| Legacy field           | schema id                  | Target                | Status                          |
| ---------------------- | -------------------------- | --------------------- | ------------------------------- |
| Connection methods     | `virtual_connectionType`   | `jsonData` (computed) | ✅ selector                     |
| Host \*                | `root_url`                 | `root`                | ✅ (tcp)                        |
| Database \*            | `jsonData_database`        | `jsonData`            | ✅ (tcp)                        |
| TNSName                | `jsonData_tnsNamesEntry`   | `jsonData`            | ✅ conditional (tns only)       |
| Oracle authentication  | `virtual_authType`         | `jsonData` (computed) | ✅ selector                     |
| User \*                | `jsonData_user`            | `jsonData`            | ✅ (basic)                      |
| Password               | `secureJsonData_password`  | `secureJsonData`      | ✅ (new UI marks required)      |
| Time zone              | `jsonData_timezone_name`   | `jsonData`            | ✅ (Additional Settings, opt)   |
| Connection Pool size   | `jsonData_connectionPoolSize` | `jsonData`         | ✅ (Additional Settings, opt)   |
| Dataproxy Timeout      | `jsonData_dataProxyTimeout` | `jsonData`           | ✅ (Additional Settings, opt)   |
| Prefetch Row Size      | `jsonData_prefetchRowsCount` | `jsonData`          | ✅ (Additional Settings, opt)   |
| Row Limit              | `jsonData_rowLimit`        | `jsonData`            | ✅ (Additional Settings, opt)   |

Groups observed in the new UI: **Connection**, **Authentication**, **Additional Settings** (optional) — matching the legacy sections (Connection / Authentication / Additional Settings). `jsonData_use_dst` is a backend-only/per-query field, correctly rendered by neither UI.

## Verification

```
go test -count=1 ./registry/grafana-oracle-datasource/...   # ok (TestSchemaConformance 8/8 subtests + LoadConfig/Validate/Defaults/Examples all PASS)
```

No schema edit was made, so no regeneration was needed; the committed artifacts remain in sync.

## Files changed

**None.** Validation-only report; Oracle was already at parity (headers n/a, fileUpload n/a, conditional required correct — the new UI's dual-required marking is the more accurate rendering).
