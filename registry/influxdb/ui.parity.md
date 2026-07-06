# InfluxDB — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered from
[`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `influxdb`
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/<uid>` (Grafana Enterprise 13.0.1)
- **New UI:** `http://localhost:64433/?path=/story/configeditor-datasourceconfigwizard--tab&args=pluginType:influxdb`
- **Method:** Playwright captured both UIs (screenshots + DOM + `Save` console payload). The Storybook story fetches the schema from the remote `schema-discovery` branch; the local edited `dsconfig.json` was served to the new UI by intercepting that request with `context.route(...)`.
- **Result:** **Parity achieved** for the connection/auth/TLS/query-language fields. One missing field (**Custom HTTP Headers**) was added; `fileUpload` is correctly not used; the wizard "General" step now includes the URL. One pre-existing new-UI limitation is flagged (`||` in `dependsOn`, see below).

---

## TL;DR of changes

| #   | Change                                                                                                 | Why                                                                                                                                        |
| --- | ------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------ |
| 1   | Added **Custom HTTP Headers** (`jsonData_httpHeaders`, `indexedPair`) to the `connection` (HTTP) group | Legacy has an "HTTP headers" section (Add header); the new UI had no headers editor                                                        |
| 2   | `root_url`: `requiredWhen:"true"` → `required:true`                                                    | Puts the URL in the wizard's synthetic **General** step (the resolver only pulls `required:true` fields); emits OpenAPI `required:["url"]` |
| 3   | Updated the secure-values instruction to describe headers as modeled                                   | Keep the embedded LLM instructions truthful                                                                                                |
| 4   | Regenerated `schema.gen.json` / `settings.gen.json` via `go generate ./...`                            | Keep committed artifacts in sync (`SchemaArtifactInSync`)                                                                                  |

Only `dsconfig.json` was edited (plus the `go generate`d artifacts). `settings.go`, `settings.ts`, `README.md` untouched.

---

## Section layout (new UI, in order)

| Order | Section (`id`)                                   | `optional` | Fields                                                                                                               |
| ----- | ------------------------------------------------ | ---------- | -------------------------------------------------------------------------------------------------------------------- |
| 1     | HTTP (`connection`)                              | no         | URL, Allowed cookies, Timeout, **Custom HTTP Headers** (added)                                                       |
| 2     | Auth (`authentication`)                          | no         | Basic auth, With Credentials, TLS Client Auth, With CA Cert, Skip TLS Verify, Forward OAuth Identity, User, Password |
| 3     | TLS/SSL Auth Details (`tls-details`)             | yes        | ServerName, CA Cert, Client Cert, Client Key                                                                         |
| 4     | Query language (`query-language`)                | no         | Version (InfluxQL/Flux/SQL), Product                                                                                 |
| 5     | InfluxDB Details (InfluxQL) (`influxql-details`) | yes        | Database, User, Password, HTTP Method, Min time interval, Show tag time                                              |
| 6     | InfluxDB Details (Flux) (`flux-details`)         | yes        | Organization, Token, Default bucket                                                                                  |
| 7     | InfluxDB Details (SQL) (`sql-details`)           | yes        | Insecure gRPC                                                                                                        |
| 8     | Other settings (`other-settings`)                | no         | Max series, PDC injected                                                                                             |

All eight sections render in the new UI (verified: tab-mode capture lists every section).

---

## Field-by-field parity

Legend: ✅ present & matching · ➕ added · 🔀 conditional (`dependsOn`)

| Field                                                                        | Legacy                           | new (schema id)                                                                                                                      | Target                                                             | Status                              |
| ---------------------------------------------------------------------------- | -------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------ | ----------------------------------- |
| URL                                                                          | text input                       | `root_url`                                                                                                                           | `root.url`                                                         | ✅ (now `required:true`)            |
| Allowed cookies                                                              | tags input                       | `jsonData_keepCookies`                                                                                                               | `jsonData.keepCookies`                                             | ✅                                  |
| Timeout                                                                      | number                           | `jsonData_timeout`                                                                                                                   | `jsonData.timeout`                                                 | ✅                                  |
| **Custom HTTP Headers**                                                      | Add header → name + secret value | `jsonData_httpHeaders`                                                                                                               | `jsonData.httpHeaderName<N>` / `secureJsonData.httpHeaderValue<N>` | ➕                                  |
| Basic auth / With Credentials                                                | switches                         | `root_basicAuth` / `root_withCredentials`                                                                                            | `root.*`                                                           | ✅                                  |
| TLS Client Auth / With CA Cert / Skip TLS Verify / Forward OAuth             | switches                         | `jsonData_tlsAuth` / `tlsAuthWithCACert` / `tlsSkipVerify` / `oauthPassThru`                                                         | `jsonData.*`                                                       | ✅                                  |
| User / Password                                                              | text / secret                    | `root_basicAuthUser` / `secureJsonData_basicAuthPassword`                                                                            | root / secureJsonData                                              | ✅ 🔀 (basicAuth)                   |
| ServerName / CA Cert / Client Cert / Client Key                              | text / textareas                 | `jsonData_serverName` / `secureJsonData_tls*`                                                                                        | jsonData / secureJsonData                                          | ✅ 🔀 (tlsAuth / tlsAuthWithCACert) |
| Version / Product                                                            | selects                          | `jsonData_version` / `jsonData_product`                                                                                              | `jsonData.*`                                                       | ✅                                  |
| Database / User / Password / HTTP Method / Min time interval / Show tag time | InfluxQL fields                  | `jsonData_dbName` / `root_user` / `secureJsonData_password` / `jsonData_httpMode` / `jsonData_timeInterval` / `jsonData_showTagTime` | mixed                                                              | ✅ 🔀 (version)                     |
| Organization / Token / Default bucket                                        | Flux fields                      | `jsonData_organization` / `secureJsonData_token` / `jsonData_defaultBucket`                                                          | mixed                                                              | ✅ 🔀 (version)                     |
| Insecure gRPC                                                                | switch                           | `jsonData_insecureGrpc`                                                                                                              | `jsonData.insecureGrpc`                                            | ✅ 🔀 (version)                     |
| Max series / PDC injected                                                    | number / (internal)              | `jsonData_maxSeries` / `jsonData_pdcInjected`                                                                                        | `jsonData.*`                                                       | ✅                                  |

---

## Gap fixed: Custom HTTP Headers

Legacy influxdb renders an **HTTP headers** section with an **Add header** button (verified:
`hasCustomHeaders:true, addHeaderBtn:true`, `fileInputs:0`). The schema didn't model it, so
the new UI showed no headers editor. Added the same `indexedPair` field used across the other
HTTP core datasources (graphite/prometheus/loki/tempo/…), placed in the HTTP group.

**Save-payload routing verified** (new UI `Save & Test` console payload): a header named
`X-Org-Id` = `influx-tenant-9` persisted as:

```json
"jsonData":       { "httpHeaderName1": "X-Org-Id" },
"secureJsonData": { "httpHeaderValue1": "influx-tenant-9" }
```

Byte-for-byte the legacy `CustomHeadersSettings` indexed-pair format (name→jsonData, value→secureJsonData write-only).

---

## `fileUpload` evaluation — not applicable

Legacy influxdb has **0 file inputs / 0 upload buttons** (TLS certs are textareas; secrets are
masked inputs). `fileUpload` is not used → not added.

---

## Conditional fields — tested

- Basic auth → reveals User/Password ✅
- TLS Client Auth → ServerName, Client Cert, Client Key; With CA Cert → CA Cert ✅
- The **Query language** selector (`jsonData_version`: InfluxQL / Flux / SQL) drives the three
  version-specific detail groups. Each detail group renders as its own (optional) section.
- **`root_url` → wizard General:** after the `required:true` fix, wizard mode opens on a
  **General** step containing the URL (`urlPresent:true`); tab mode unaffected.

### Known limitation (pre-existing, flagged — not introduced by this change)

Several InfluxDB version fields use a **compound `dependsOn`** with `||`, e.g.
`jsonData_dbName.dependsOn = "jsonData_version == 'InfluxQL' || jsonData_version == 'SQL'"`
(also `timeInterval`, `token`). The new UI's CEL evaluator (`plugin-ui` `config.ts`
`parseDependsOn`) currently supports only a **single** comparison (`fieldId == 'value'` / `!=`),
not `||`/`&&`. Fields with a `||` condition are therefore not gated correctly by version in the
new UI (they do not hide/show per the compound rule). This is a **schema-authoring vs. renderer**
gap that predates this parity pass and is **out of scope** for the headers/required fix (fixing it
means either splitting into single-comparison `dependsOn` per version — which changes modeling —
or extending the plugin-ui CEL parser to support `||`/`&&`). Recommended as a follow-up. All
single-comparison conditionals evaluate correctly.

---

## Verification

```
go generate ./registry/influxdb/...
go test ./registry/influxdb/...   # 8/8 conformance subtests PASS
go test ./registry/... ./schema/...  # full suite PASS (no regressions)
```

Conformance subtests: `BaseFieldsResolved`, `SchemaRoundTrip`, `SchemaArtifactInSync`,
`SchemaSpecHasNoSecureJSON`, `ConfigSchemaValid`, `JSONDataMatchesStruct`,
`JSONDataTypesMatchStruct`, `SecureValuesMatchLoadSettings` — all PASS. No `conformance.go` or
`plugin-ui` change was required (the `indexedPair` skip and `authentication`/`auth` group
recognition were already landed in earlier work).

## Files changed

- [`registry/influxdb/dsconfig.json`](dsconfig.json) — added `jsonData_httpHeaders` (in the `connection`/HTTP group); `root_url` → `required:true`; refreshed the secure-values instruction.
- [`registry/influxdb/schema.gen.json`](schema.gen.json), [`registry/influxdb/settings.gen.json`](settings.gen.json) — regenerated by `go generate`.

_Unchanged by design:_ `settings.go`, `settings.ts`, `README.md`, `settings.examples.gen.json`.
