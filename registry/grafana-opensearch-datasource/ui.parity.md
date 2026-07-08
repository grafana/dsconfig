# OpenSearch — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered from
[`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `grafana-opensearch-datasource`
- **New UI:** `http://localhost:64433/?path=/story/configeditor-datasourceconfigwizard--tab&args=pluginType:grafana-opensearch-datasource`
- **Method:** Playwright drove both UIs — the new UI with the local schema served via `context.route(...)` interception; the legacy editor at `/connections/datasources/edit/bfrbd04e93rpce`.
- **Result:** **Parity achieved** for the modeled fields. Added **Custom HTTP Headers**; fixed the URL/required modeling so the wizard "General" step includes the URL; removed a redundant always-true `requiredWhen` on three already-`required` fields.

> OpenSearch is an Elasticsearch fork (adds AWS SigV4 auth, flavor/version, PPL, serverless); the same three fixes applied to `elasticsearch` apply here. Unlike Elasticsearch, the legacy OpenSearch editor **did render** in this instance, so the header/file-input findings below are **directly captured**, not inferred by analogy.

---

## TL;DR of changes (only `dsconfig.json`)

| #   | Change                                                                                                                                 | Why                                                                                                                                                   |
| --- | -------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | Added **Custom HTTP Headers** (`jsonData_httpHeaders`, `indexedPair`) to the `additional-http` group                                   | Legacy editor has a "Custom HTTP Headers" section with **Add header**; the new UI had no headers editor                                               |
| 2   | `root_url`: `requiredWhen:"true"` → `required:true`                                                                                    | Folds URL into the wizard's synthetic **General** step; emits OpenAPI `required:["url"]`                                                              |
| 3   | `jsonData_database`, `jsonData_timeField`, `jsonData_version`: removed redundant `requiredWhen:"true"` (kept existing `required:true`) | These fields already carried `required:true`; the extra always-true `requiredWhen` was redundant and duplicated the requirement in the generated spec |
| 4   | Updated the secure-values instruction: custom headers are now **modeled** (was "not modeled here")                                     | Keep LLM guidance accurate                                                                                                                            |
| 5   | Regenerated `schema.gen.json` / `settings.gen.json`                                                                                    | Keep artifacts in sync                                                                                                                                |

No Go change was required: `schema/conformance.go` excludes `indexedPair` fields from the jsonData↔struct parity and secure-key parity checks (`isIndexedPairField`), exactly as it does for the eight sibling HTTP datasources.

---

## Section layout (new UI, verified rendering)

HTTP · **Additional HTTP settings** (opt — now includes Custom HTTP Headers) · Auth · TLS/SSL Auth Details (opt) · OpenSearch details · Logs (opt) · Data links (opt).

Legacy headings captured (UID `bfrbd04e93rpce`): `HTTP`, `Auth`, `Custom HTTP Headers`, `OpenSearch details`, `Logs`, `Data links`. The new UI carries the same content; the "Custom HTTP Headers" control lives inside the optional **Additional HTTP settings** group (same placement as elasticsearch).

## Field-by-field parity (highlights)

| Field                                                                                               | schema id                                                                    | Target                                                             | Status                                  |
| --------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------- | ------------------------------------------------------------------ | --------------------------------------- |
| URL                                                                                                 | `root_url`                                                                   | `root.url`                                                         | ✅ (now `required:true`)                |
| Access (Server / Browser)                                                                           | `root_access`                                                                | `root.access`                                                      | ✅                                      |
| Basic auth toggle / User / Password                                                                 | `root_basicAuth` / `root_basicAuthUser` / `secureJsonData_basicAuthPassword` | root / secureJsonData                                              | ✅ 🔀                                   |
| With Credentials                                                                                    | `root_withCredentials`                                                       | `root.withCredentials`                                             | ✅                                      |
| SigV4 auth                                                                                          | `jsonData_sigV4Auth`                                                         | jsonData                                                           | ✅                                      |
| Forward OAuth Identity                                                                              | `jsonData_oauthPassThru`                                                     | jsonData                                                           | ✅                                      |
| TLS Client Auth / With CA Cert / Skip TLS Verify / ServerName / certs                               | `jsonData_tls*` / `jsonData_serverName` / `secureJsonData_tls*`              | jsonData / secureJsonData                                          | ✅ 🔀                                   |
| Allowed cookies / Timeout                                                                           | `jsonData_keepCookies` / `jsonData_timeout`                                  | jsonData                                                           | ✅                                      |
| **Custom HTTP Headers**                                                                             | `jsonData_httpHeaders`                                                       | `jsonData.httpHeaderName<N>` / `secureJsonData.httpHeaderValue<N>` | ➕                                      |
| Index name / Pattern / Time field                                                                   | `jsonData_database` / `jsonData_interval` / `jsonData_timeField`             | jsonData                                                           | ✅ (database/timeField required)        |
| Flavor / Version / Serverless                                                                       | `jsonData_flavor` / `jsonData_version` / `jsonData_serverless`               | jsonData                                                           | ✅ (version required unless serverless) |
| Max concurrent shard requests, min time interval, PPL enabled, log message/level fields, data links | (details / logs / data-links groups)                                         | jsonData                                                           | ✅                                      |

`🔀` = target crosses jsonData/root/secureJsonData boundaries. `➕` = newly added.

---

## `fileUpload` — not applicable

Legacy capture reported `fileInputs:0` and no upload buttons. TLS certs are textareas → masked secure inputs in the new UI. `fileUpload` not added.

## Conditional fields & effects — tested (new UI)

- Auth is a set of **independent boolean toggles** (no discriminator / no `virtual_authMethod`), matching @grafana/ui `DataSourceHttpSettings`: Basic auth, With Credentials, SigV4 auth, TLS Client Auth, With CA Cert, Skip TLS Verify, Forward OAuth Identity all render.
- TLS toggles reveal ServerName / CA Cert / Client Cert / Client Key (TLS/SSL Auth Details group).
- `jsonData_serverless` gates `jsonData_version` / `jsonData_maxConcurrentShardRequests` via `dependsOn`.
- **Wizard "General" step:** after the `required:true` fix, wizard mode opens at step `1/8` with the URL present (`urlPresent:true`); `database`/`timeField`/`version` (already `required:true`) also fold into General. Tab mode unaffected (`urlPresent:true`, `hasHeadersEditor:true`).

## Custom HTTP Headers routing (driven, not inferred)

The new UI mounts plugin-ui's dedicated CustomHeaders editor for the `role:"http.header"` / `indexedPair` field — confirmed by its native placeholders (`key` / `custom http header value`). Driving the tab story (fill required fields → **Add custom http header** → name `X-Scope-OrgID`, value `opensearch-tenant-42` → **Save & Test**) produced:

```
jsonData.httpHeaderName1   = "X-Scope-OrgID"
secureJsonData.httpHeaderValue1 = "opensearch-tenant-42"   (write-only; secureJsonFields.httpHeaderValue1 = false)
url = "https://opensearch.example.com:9200"
```

This is the identical `indexedPair` mapping (`httpHeaderName{index}` / `httpHeaderValue{index}`, `startIndex:1`) verified across the sibling HTTP datasources (graphite/prometheus/loki/tempo/opentsdb/influxdb/jaeger/zipkin), so the value persists write-only in `secureJsonData`.

> Bonus verification of the required-field fix: before all four required fields (url, index, version, timeField) were filled, **Save & Test** emitted no payload — confirming the strengthened required gating is live.

---

## Verification

```
go generate ./registry/grafana-opensearch-datasource/...
go test ./registry/grafana-opensearch-datasource/...   # all subtests PASS
```

`TestSchemaConformance` (8 subtests: BaseFieldsResolved, SchemaRoundTrip, SchemaArtifactInSync, SchemaSpecHasNoSecureJSON, ConfigSchemaValid, JSONDataMatchesStruct, JSONDataTypesMatchStruct, SecureValuesMatchLoadSettings) plus `TestLoadConfig` / `TestApplyDefaults` / `TestValidate` all pass. The broader `./registry/grafana-opensearch-datasource/... ./schema/...` suite also passes. No `conformance.go` / `plugin-ui` change needed.

## Files changed

- [`registry/grafana-opensearch-datasource/dsconfig.json`](dsconfig.json) — added `jsonData_httpHeaders`; `root_url` → `required:true`; removed redundant `requiredWhen:"true"` on `jsonData_database` / `jsonData_timeField` / `jsonData_version`; updated the secure-values instruction to note headers are now modeled.
- [`registry/grafana-opensearch-datasource/schema.gen.json`](schema.gen.json), [`registry/grafana-opensearch-datasource/settings.gen.json`](settings.gen.json) — regenerated.

_Unchanged by design:_ `settings.go`, `settings.ts`, `README.md`, `conformance.go`, `plugin-ui`.

> Note: `README.md` still documents `root_url` as `requiredWhen: "true"` (rows in the field table and the "Field notes" section). That is now stale relative to `dsconfig.json`, but `README.md` is out of scope for this pass and was intentionally left untouched.
