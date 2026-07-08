# AWS IoT SiteWise — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered from
[`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `grafana-iot-sitewise-datasource`
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/<uid>` (Grafana Enterprise 13.0.1)
- **New UI:** Storybook `configeditor-datasourceconfigwizard` story, `pluginType:grafana-iot-sitewise-datasource` (local schema served via `context.route(...)` interception).
- **Result:** **Parity already achieved — no `dsconfig.json` changes required.**

## Findings (no fixes needed)

- **No Custom HTTP Headers.** AWS-SDK auth datasource; legacy has no "HTTP headers" section (`hasCustomHeaders:false`, `addHeaderBtn:false`). New UI `hasHeadersEditor:false`. Correctly not modeled.
- **No `fileUpload`.** AWS credentials are text / profile / IAM role; legacy `fileInputs:0`, `uploadButtons:[]`. Not used.
- **No `required:true` fix.** No unconditional `requiredWhen:"true"` fields — the access/secret keys are conditionally required on the AWS auth provider (`authType == 'keys'`), which is correct.

## Conditional fields — tested

The AWS authentication-provider discriminator drives the credential fields: `keys` → Access Key ID + Secret Access Key; `credentials` → Credentials Profile Name; `default` / `ec2_iam_role` → none. Assume-role ARN / External ID / Endpoint / Default Region and the Edge settings render per the schema. All conditionals match the legacy editor.

## Verification

```
go test ./registry/grafana-iot-sitewise-datasource/...   # 8/8 conformance subtests PASS
```

No schema edit → no regeneration; committed artifacts remain in sync; full suite passes.

## Known cross-cutting note (shared AWS-registry, plugin-ui)

Like cloudwatch/athena/redshift/x-ray/timestream/dynamodb, the new UI renders the
`backend-only`-tagged `secureJsonData.sessionToken` secret that the legacy editor omits. Honouring
the `backend-only` tag is a `plugin-ui` renderer concern (out of scope here) — flagged, not changed.

## Files changed

**None.** Validation-only report; iot-sitewise was already at parity (headers n/a, fileUpload n/a, conditional required correct).
