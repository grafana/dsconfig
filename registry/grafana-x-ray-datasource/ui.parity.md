# AWS X-Ray / Application Signals — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered from
[`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `grafana-x-ray-datasource` (product name "AWS Application Signals")
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/efrbd04ucg7wge` (Grafana Enterprise 13.0.1)
- **New UI:** `http://localhost:64433/?path=/story/configeditor-datasourceconfigwizard--tab&args=pluginType:grafana-x-ray-datasource`
- **Method:** Playwright captured the legacy UI (`legacy-expand-xray-parity.png`) and drove the new UI (local schema served via `context.route(...)`), including a conditional sweep across all five auth providers (`verify-xray-cond-*.png`).
- **Result:** **Parity already achieved — no `dsconfig.json` changes required.**

---

## Findings (no fixes needed)

**No Custom HTTP Headers.** X-Ray is an AWS-SDK auth datasource (AWS SigV4 via access
keys / assume-role / credentials file / workspace IAM / Grafana Assume Role). Its legacy
editor has **no** Custom HTTP Headers section (`hasCustomHeaders:false`, `addHeaderBtn:false`).
The new UI renders no headers editor (`hasHeadersEditor:false`). Headers are correctly **not**
modeled.

**No `fileUpload`.** AWS credentials are entered as text (Access Key ID / Secret Access Key)
or resolved from the environment / assumed role — there is **no** file-upload control
(`fileInputs:0`, `uploadButtons:[]`). `fileUpload` is correctly **not** used.

**No `required:true` fix.** X-Ray has **no** `requiredWhen:"true"` fields. Its
conditionally-required credentials (access/secret key when auth = "Access & secret key")
use auth-type-gated `requiredWhen` expressions (`jsonData_authType == 'keys'`), which is
correct. No change.

## New UI verification

Tab mode renders all sections: **Authentication**, **Assume Role**, **Additional Settings**.
`hasHeadersEditor:false` (correct). `urlPresent:true` — the optional AWS **Endpoint** override
field. Wizard mode renders the same schema as a 4-step flow with no headers editor.

## Field-by-field parity

| Legacy field                      | schema id                                               | Target         | Status                             |
| --------------------------------- | ------------------------------------------------------- | -------------- | ---------------------------------- |
| Authentication Provider           | `jsonData_authType` (auth.discriminator)                | `jsonData`     | ✅ 🔀                              |
| Access Key ID / Secret Access Key | `secureJsonData_accessKey` / `secureJsonData_secretKey` | secureJsonData | ✅ 🔀 (keys auth)                  |
| Credentials Profile Name          | `jsonData_profile`                                      | `jsonData`     | ✅ 🔀 (credentials auth)           |
| Assume Role ARN / External ID     | `jsonData_assumeRoleArn` / `jsonData_externalId`        | `jsonData`     | ✅ 🔀                              |
| Endpoint                          | `jsonData_endpoint`                                     | `jsonData`     | ✅                                 |
| Default Region                    | `jsonData_defaultRegion`                                | `jsonData`     | ✅                                 |
| _(none — provisioning only)_      | `secureJsonData_sessionToken` (`backend-only`)          | secureJsonData | ⚠️ renderer surfaces it — see note |

## Conditional fields — tested

The Authentication Provider discriminator drives which credential fields are shown/required.
All five editor-selectable providers were exercised in the new UI and render per the schema
`dependsOn` / `requiredWhen` expressions:

| Auth provider (`authType`)                  | Profile | Access/Secret Key   | Assume Role ARN | External ID | Endpoint |
| ------------------------------------------- | ------- | ------------------- | --------------- | ----------- | -------- |
| AWS SDK Default (`default`)                 | ✗       | ✗                   | ✓               | ✓           | ✓        |
| Access & secret key (`keys`)                | ✗       | ✓ **(required \*)** | ✓               | ✓           | ✓        |
| Credentials file (`credentials`)            | ✓       | ✗                   | ✓               | ✓           | ✓        |
| Workspace IAM Role (`ec2_iam_role`)         | ✗       | ✗                   | ✓               | ✓           | ✓        |
| Grafana Assume Role (`grafana_assume_role`) | ✗       | ✗                   | ✓               | **✗**       | **✗**    |

External ID and Endpoint correctly disappear for `grafana_assume_role`
(`dependsOn "jsonData_authType != 'grafana_assume_role'"`); Access Key ID / Secret Access Key
appear and are marked required only for `keys`
(evidence: `verify-xray-cond-B-keys.png`, `verify-xray-cond-C-credentials.png`,
`verify-xray-cond-E-grafana-assume-role.png`).

## Note: backend-only `sessionToken` (plugin-ui behaviour, not a dsconfig gap)

The new schema-driven renderer surfaces `secureJsonData.sessionToken` as an empty secure
input in the Authentication group under every auth provider, whereas the legacy AWS
`ConnectionConfig` editor **never** renders an input for it (it is provisioning-only, tagged
`backend-only`). This is **not** an X-Ray-specific defect and is **not** fixable via
`dsconfig.json`:

- The field is intentionally modeled (with `role: auth.aws.sessionToken`, `tags: ["backend-only"]`)
  so provisioning and LLM consumers know about it; the backend still reads it.
- `backend-only` is free-form metadata; the schema's only visibility knob (`hidden`) is
  **reserved/ignored** (`dsconfig/schema.json` `FieldPatch.hidden`, `schema.md:77`), and there
  is no field-level property that suppresses rendering.
- The **CloudWatch** reference has the byte-identical field (same `key`, `role`, `tags`, and
  Authentication-group membership) and its new UI renders `sessionToken` the same way, yet it
  was signed off as _"parity already achieved — no changes"_. Removing it from `fieldRefs` or
  bolting on a never-true `dependsOn` would diverge from that reference and drop metadata.

Honouring the `backend-only` tag (so the renderer skips such fields) is a **plugin-ui**
change that would apply to all AWS-SDK datasources — out of scope for this schema and
explicitly excluded from these edits. Reported for awareness only.

---

## Verification

```
go test ./registry/grafana-x-ray-datasource/...   # PASS (0.420s)
```

`TestSchemaConformance` 8/8 subtests PASS (`BaseFieldsResolved`, `SchemaRoundTrip`,
`SchemaArtifactInSync`, `SchemaSpecHasNoSecureJSON`, `ConfigSchemaValid`,
`JSONDataMatchesStruct`, `JSONDataTypesMatchStruct`, `SecureValuesMatchLoadSettings`), plus
all `TestLoadConfig` / `TestApplyDefaults` / `TestValidate` / `TestEffectiveProfile` subtests.
No schema edit was made, so no regeneration was needed; committed artifacts remain in sync.

## Files changed

**None.** X-Ray was already at parity for HTTP headers (n/a), file upload (n/a), and the
required/conditional auth behaviour (auth-type-gated `requiredWhen` is correct). This report
documents the validation only. The `sessionToken` rendering delta is a plugin-ui renderer
behaviour, identical to CloudWatch, and is not addressable via `dsconfig.json`.
