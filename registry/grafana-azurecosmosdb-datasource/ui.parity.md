# Azure Cosmos DB — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered from
[`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `grafana-azurecosmosdb-datasource`
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/bfrbqiwjanbi8e` (Grafana Enterprise 13.0.1, `a100054f21`)
- **New UI:** `http://192.168.1.241:58899/?path=/story/configeditor-datasourceconfigwizard--tab&args=pluginType:grafana-azurecosmosdb-datasource`
- **Method:** Playwright captured the legacy UI (`legacy-expand-azurecosmosdb-verify.png`) and drove the new UI in both **tab** and **wizard** modes (local `dsconfig.json` served via `context.route(...)` intercepting the `raw.githubusercontent.com` fetch).
- **Result:** **One fix applied** — the two mandatory credential fields were modeled with `requiredWhen:"true"` (a degenerate always-true conditional) instead of a plain `required:true`. Converted both to `required:true`. No HTTP-headers or `fileUpload` changes (legacy confirmed to have neither).

---

## The fix

Both fields the backend strictly requires — `pkg/plugin/settings.go:15-23` (`LoadSettings` →
`"account endpoint is empty"` / `"account key is empty"`) — were declared with
`"requiredWhen": "true"` rather than `"required": true`. `requiredWhen` is for *conditional*
(auth-type-gated) requirements; an unconditional `"true"` is simply a hard requirement and must
be `required:true` so the generated schema emits a proper JSON-Schema `required` entry and the
new UI renders the mandatory `*`.

| Field | Before | After |
| ----- | ------ | ----- |
| `jsonData_accountEndpoint` | `"requiredWhen": "true"` | `"required": true` |
| `secureJsonData_accountKey` | `"requiredWhen": "true"` | `"required": true` |

Regenerated artifacts now emit `jsonData.required:["accountEndpoint"]` and
`secureValues[].accountKey.required:true` (previously `x-dsconfig-required-when:"true"` on the
jsonData field and no `required` flag on the secure value).

---

## Findings

**No Custom HTTP Headers.** Azure Cosmos DB authenticates solely with an account master key
(`azcosmos.NewKeyCredential` + `NewClientWithKey`, `pkg/cosmos/client.go:24-52`). Its legacy
editor has **no** Custom HTTP Headers section (`hasCustomHeaders:false`, `addHeaderBtn:false`).
Headers are correctly **not** modeled.

**No `fileUpload`.** Credentials are the account endpoint URI (text) and the account key (secret
text) — there is **no** file-upload control (`fileInputs:0`, `uploadButtons:[]`). `fileUpload`
is correctly **not** used.

**Required fix applied** (see above). Unlike a conditional credential, both fields are
*unconditionally* required, so `required:true` (not `requiredWhen`) is the correct model.

## New UI verification

**Tab mode** renders the single **Account configuration** section with **Account Endpoint** `*`
and **Account Key** `*` (secret input with reveal toggle), both marked required.
`hasHeadersEditor:false` (correct). `urlPresent:false` — Cosmos DB has no proxied HTTP URL; the
account endpoint is a plain `jsonData` field (`role: endpoint.baseUrl`), not the datasource URL.

**Wizard mode** — the **General** step (`1/2`) contains both required fields: **Account
Endpoint** `*` and **Account Key** `*`, with the "Fields marked with * are required" note.

## Field-by-field parity

| Legacy field     | schema id                    | target           | Status                |
| ---------------- | ---------------------------- | ---------------- | --------------------- |
| Account Endpoint | `jsonData_accountEndpoint`   | `jsonData`       | ✅ required           |
| Account Key      | `secureJsonData_accountKey`  | `secureJsonData` | ✅ required (secret)  |

## Conditional fields

**None.** Azure Cosmos DB has a single auth method (account key); there are no discriminators or
conditional (`dependsOn` / `requiredWhen`) fields. Both fields are unconditionally required.

---

## Verification

```
go generate ./registry/grafana-azurecosmosdb-datasource/...   # regenerates .gen.json (idempotent)
go test     ./registry/grafana-azurecosmosdb-datasource/...   # PASS
```

`TestSchemaConformance` (8/8 subtests: BaseFieldsResolved, SchemaRoundTrip,
SchemaArtifactInSync, SchemaSpecHasNoSecureJSON, ConfigSchemaValid, JSONDataMatchesStruct,
JSONDataTypesMatchStruct, SecureValuesMatchLoadSettings) plus `TestLoadConfig`,
`TestApplyDefaults`, and `TestValidate` all pass. Re-running `go generate` produces no further
drift; committed `schema.gen.json` / `settings.gen.json` / `settings.examples.gen.json` remain
in sync.

## Files changed

- `registry/grafana-azurecosmosdb-datasource/dsconfig.json` — `requiredWhen:"true"` → `required:true` on `jsonData_accountEndpoint` and `secureJsonData_accountKey`.
- `registry/grafana-azurecosmosdb-datasource/schema.gen.json` — regenerated.
- `registry/grafana-azurecosmosdb-datasource/settings.gen.json` — regenerated.
- `settings.examples.gen.json` — regenerated identically (no content change).

No `settings.go` / `settings.ts` / `README` / `schema` / `conformance.go` / `plugin-ui` change
was required.
