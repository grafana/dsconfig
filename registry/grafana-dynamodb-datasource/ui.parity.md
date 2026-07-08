# DynamoDB — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered from
[`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `grafana-dynamodb-datasource`
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/afrbd04w5dkw0e` (Grafana Enterprise 13.0.1)
- **New UI:** `http://localhost:64433/?path=/story/configeditor-datasourceconfigwizard--tab&args=pluginType:grafana-dynamodb-datasource`
- **Method:** Playwright captured the legacy UI (`legacy-expand-dynamodb-parity.png`) and drove the new UI in both tab (`newgen-dynamodb-tab.png`) and wizard (`newgen-dynamodb-wizard.png`) modes, with the local schema served via `context.route(...)`. Auth-provider conditionals were driven per value (`verify-dynamodb-cond-*.png`).
- **Result:** **Parity already achieved — no `dsconfig.json` changes required.**

---

## Findings (no fixes needed)

**No Custom HTTP Headers.** DynamoDB is an AWS-SDK auth datasource (its config editor is
`<ConnectionConfig hideAssumeRoleArn>` from `@grafana/aws-sdk` — no `url`, no header UI). Its
legacy editor has **no** Custom HTTP Headers section (`hasCustomHeaders:false`,
`addHeaderBtn:false`). The new UI renders **no** headers editor (`hasHeadersEditor:false`).
Headers are correctly **not** modeled.

**No `fileUpload`.** AWS credentials are entered as text (Access Key ID / Secret Access Key),
read from a named profile in `~/.aws/credentials`, or resolved from the environment / instance
role — there is **no** file-upload control (`fileInputs:0`, `uploadButtons:[]`). `fileUpload` is
correctly **not** used.

**No `required:true` fix.** DynamoDB has **no** `requiredWhen:"true"` fields. The only required
fields — `secureJsonData.accessKey` / `secureJsonData.secretKey` — are auth-type-gated
(`requiredWhen: "jsonData_authType == 'keys'"`), which is correct. No change.

**`sessionToken` raw label — accepted cross-plugin renderer behavior, not a gap.** The tab render
surfaces a raw `sessionToken` field label in the Authentication group. This field is
`backend-only` (provisioning-only; the legacy `ConnectionConfig` never renders an input for it)
and carries no `label`/`ui`, so the renderer falls back to the key name. This is **identical**
to CloudWatch's `secureJsonData_sessionToken` definition and render (`newgen-cw-cur.json`), and
CloudWatch was accepted as "parity already achieved." Suppressing a `backend-only` field's label
would be a `plugin-ui` renderer change (out of scope) — **not** a dynamodb `dsconfig.json`
fixable gap. No change.

## New UI verification

Tab mode renders all four groups: **Authentication**, **Additional Settings**, **Driver
Settings** (optional), **Legacy Migration** (optional). `hasHeadersEditor:false` (correct).
`urlPresent:true` — the optional AWS **Endpoint** override field (not a base URL). Wizard mode
renders the same schema step-by-step (step 1/5 = Authentication) with `hasHeadersEditor:false`.

The optional **Driver Settings** (`timeout`/`retries`/`pause`) and **Legacy Migration**
(`isV2`/`accessId`/`region`) groups intentionally surface backend-only/provisioning fields the
legacy editor never rendered; they are gated behind `optional` collapsible sections and do not
affect the primary credential flow.

## Field-by-field parity (highlights)

| Legacy field                      | schema id                                               | Target           | Status                   |
| --------------------------------- | ------------------------------------------------------- | ---------------- | ------------------------ |
| Authentication Provider           | `jsonData_authType` (auth.discriminator)                | `jsonData`       | ✅ 🔀                    |
| Access Key ID / Secret Access Key | `secureJsonData_accessKey` / `secureJsonData_secretKey` | `secureJsonData` | ✅ 🔀 (keys auth)        |
| Credentials Profile Name          | `jsonData_profile`                                      | `jsonData`       | ✅ 🔀 (credentials auth) |
| Endpoint                          | `jsonData_endpoint`                                     | `jsonData`       | ✅                       |
| Default Region                    | `jsonData_defaultRegion`                                | `jsonData`       | ✅                       |
| Assume Role ARN / External ID     | — (hidden via `hideAssumeRoleArn`)                      | n/a              | ✅ (hidden in both)      |

## Conditional fields — tested

The Authentication Provider discriminator drives which credential fields are shown/required.
Driving the select in the new UI (`verify-dynamodb-conditionals-result.json`) reproduced the
legacy `ConnectionConfig` behavior exactly:

| authType                            | Access/Secret Key    | Profile   | Assume Role ARN / External ID | Endpoint / Region |
| ----------------------------------- | -------------------- | --------- | ----------------------------- | ----------------- |
| `default` (AWS SDK Default)         | hidden               | hidden    | never                         | shown             |
| `keys` (Access & secret key)        | **shown (required)** | hidden    | never                         | shown             |
| `credentials` (Credentials file)    | hidden               | **shown** | never                         | shown             |
| `ec2_iam_role` (Workspace IAM Role) | hidden               | hidden    | never                         | shown             |

`assumeRoleArn` / `externalId` are **never** present in any state (confirms `hideAssumeRoleArn`
parity), and `grafana_assume_role` is correctly absent from the provider Select
(`grafana-dynamodb-datasource` is not in `DS_TYPES_THAT_SUPPORT_TEMP_CREDS`).

---

## Verification

```
go test -count=1 ./registry/grafana-dynamodb-datasource/...   # ok (0.475s) — conformance suite PASS
```

No schema edit was made, so no regeneration was needed; committed `.gen.json` artifacts remain
in sync.

## Files changed

**None.** DynamoDB was already at parity for HTTP headers (n/a), file upload (n/a), and the
required/conditional credential behavior (auth-type-gated `requiredWhen` is correct). The
`sessionToken` raw-label render is an accepted cross-plugin renderer behavior matching the
CloudWatch reference and is not fixable via this plugin's `dsconfig.json`. This report documents
the validation only.
