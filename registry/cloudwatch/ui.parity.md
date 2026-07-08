# CloudWatch — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered from
[`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `cloudwatch`
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/<uid>` (Grafana Enterprise 13.0.1)
- **New UI:** `http://localhost:64433/?path=/story/configeditor-datasourceconfigwizard--tab&args=pluginType:cloudwatch`
- **Method:** Playwright captured the legacy UI (`legacy-expand-core2-cloudwatch.png`) and drove the new UI (local schema served via `context.route(...)`).
- **Result:** **Parity already achieved — no `dsconfig.json` changes required.**

---

## Findings (no fixes needed)

**No Custom HTTP Headers.** CloudWatch is an AWS-SDK auth datasource (AWS SigV4 via access
keys / assume-role / credentials file / workspace IAM). Its legacy editor has **no** Custom HTTP
Headers section (`hasCustomHeaders:false`, `addHeaderBtn:false`). Headers are correctly **not**
modeled.

**No `fileUpload`.** AWS credentials are entered as text (Access Key ID / Secret Access Key) or
resolved from the environment / assumed role — there is **no** file-upload control
(`fileInputs:0`, `uploadButtons:[]`). `fileUpload` is correctly **not** used.

**No `required:true` fix.** CloudWatch has **no** `requiredWhen:"true"` fields. Its
conditionally-required credentials (access/secret key when auth = "Access & secret key", etc.)
use auth-type-gated `requiredWhen` expressions, which is correct. No change.

## New UI verification

Tab mode renders all sections: **Authentication**, **Assume Role**, **Proxy Configuration**
(opt), **CloudWatch Logs**, **Application Signals trace link** (opt). `hasHeadersEditor:false`
(correct). `urlPresent:true` — the optional AWS **Endpoint** override field.

## Field-by-field parity (highlights)

| Legacy field                                    | schema id                   | Target                | Status            |
| ----------------------------------------------- | --------------------------- | --------------------- | ----------------- |
| Authentication Provider                         | auth-provider discriminator | `jsonData`            | ✅ 🔀             |
| Access Key ID / Secret Access Key               | access/secret key fields    | root / secureJsonData | ✅ 🔀 (keys auth) |
| Assume Role ARN / External ID                   | assumeRoleArn / externalId  | `jsonData`            | ✅ 🔀             |
| Credentials Profile Name                        | profile                     | `jsonData`            | ✅ 🔀             |
| Default Region                                  | defaultRegion               | `jsonData`            | ✅                |
| Endpoint                                        | endpoint                    | `jsonData`            | ✅                |
| CloudWatch Logs (log groups) / X-Ray trace link | logs/trace-link fields      | `jsonData`            | ✅                |
| Proxy configuration (Secure Socks)              | proxy fields                | `jsonData`            | ✅ (opt)          |

## Conditional fields — tested

The Authentication Provider discriminator drives which credential fields are shown/required
(Access & secret key → key fields; Assume Role → ARN/external ID; Credentials file → profile;
Workspace IAM / default chain → none). All conditionals render per the schema `dependsOn`.

---

## Verification

```
go test ./registry/cloudwatch/...   # 8/8 conformance subtests PASS
```

No schema edit was made, so no regeneration was needed; committed artifacts remain in sync and
the full `./registry/... ./schema/...` suite passes.

## Files changed

**None.** CloudWatch was already at parity for HTTP headers (n/a), file upload (n/a), and the
required/General-step behaviour (conditional required is correct). This report documents the
validation only.
