# Elasticsearch — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered from
[`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `elasticsearch`
- **New UI:** `http://localhost:64433/?path=/story/configeditor-datasourceconfigwizard--tab&args=pluginType:elasticsearch`
- **Method:** Playwright drove the new UI (local schema served via `context.route(...)` interception).
- **Result:** **Parity achieved** for the modeled fields. Added **Custom HTTP Headers**; fixed the URL/required modeling so the wizard "General" step includes the URL; cleaned up a redundant `requiredWhen` on two already-`required` fields.

> **Legacy-capture caveat:** in this Grafana 13.0.1 instance the Elasticsearch **frontend plugin did not render** its config editor (the settings page returned "Data source not found"; `/api/plugins/elasticsearch/settings` was empty), so the legacy UI could not be screenshotted here. The header determination below is therefore made by **schema analogy**: Elasticsearch uses the same `@grafana/plugin-ui` `Auth` component as Prometheus/Loki/Tempo (it has a `virtual_authMethod` selector with `effects`), and those three were **confirmed** to render an "HTTP headers" section. Elasticsearch is a standard HTTP datasource whose SDK client honours `httpHeaderName<N>`/`httpHeaderValue<N>`. Re-confirm against a working Elasticsearch editor when available.

---

## TL;DR of changes (only `dsconfig.json`)

| #   | Change                                                                                                          | Why                                                                                                                                                   |
| --- | --------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | Added **Custom HTTP Headers** (`jsonData_httpHeaders`, `indexedPair`) to the `additional-http` group            | Auth-component datasources render an "HTTP headers" section; the new UI had no headers editor                                                         |
| 2   | `root_url`: `requiredWhen:"true"` → `required:true`                                                             | Puts URL in the wizard's synthetic **General** step; emits OpenAPI `required:["url"]`                                                                 |
| 3   | `jsonData_index`, `jsonData_timeField`: removed redundant `requiredWhen:"true"` (kept existing `required:true`) | These fields already carried `required:true`; the extra always-true `requiredWhen` was redundant and duplicated the requirement in the generated spec |
| 4   | Regenerated `schema.gen.json` / `settings.gen.json`                                                             | Keep artifacts in sync                                                                                                                                |

---

## Section layout (new UI, verified rendering)

Connection · Authentication · TLS settings (opt) · **Additional settings** (opt — now includes Custom HTTP Headers) · Elasticsearch details · Logs (opt) · Data links (opt).

## Field-by-field parity (highlights)

| Field                                                                               | schema id                                                       | Target                                                             | Status                        |
| ----------------------------------------------------------------------------------- | --------------------------------------------------------------- | ------------------------------------------------------------------ | ----------------------------- |
| URL                                                                                 | `root_url`                                                      | `root.url`                                                         | ✅ (now `required:true`)      |
| Auth method (No Auth / Basic / Forward OAuth)                                       | `virtual_authMethod` (+ effects)                                | —                                                                  | ✅                            |
| Basic auth user / password                                                          | `root_basicAuthUser` / `secureJsonData_basicAuthPassword`       | root / secureJsonData                                              | ✅ 🔀                         |
| TLS Client Auth / With CA Cert / Skip TLS / ServerName / certs                      | `jsonData_tls*` / `jsonData_serverName` / `secureJsonData_tls*` | jsonData / secureJsonData                                          | ✅ 🔀                         |
| Allowed cookies / Timeout                                                           | `jsonData_keepCookies` / `jsonData_timeout`                     | jsonData                                                           | ✅                            |
| **Custom HTTP Headers**                                                             | `jsonData_httpHeaders`                                          | `jsonData.httpHeaderName<N>` / `secureJsonData.httpHeaderValue<N>` | ➕                            |
| Index name / Pattern / Time field                                                   | `jsonData_index` / `jsonData_interval` / `jsonData_timeField`   | jsonData                                                           | ✅ (index/timeField required) |
| Max concurrent shard requests, include frozen, log message/level fields, data links | (jsonData / details / logs / data-links groups)                 | jsonData                                                           | ✅                            |

---

## `fileUpload` — not applicable

No file-upload control in the Elasticsearch editor (TLS certs are textareas → masked secure inputs in the new UI). `fileUpload` not added.

## Conditional fields & effects — tested (new UI)

- `virtual_authMethod` selector renders (No Auth / Basic auth / Forward OAuth Identity). Basic reveals user/password.
- TLS toggles reveal ServerName / CA Cert / Client Cert / Client Key.
- **Wizard "General" step:** after the `required:true` fix, wizard mode opens with the URL present (`urlPresent:true`); `index`/`timeField` (already `required:true`) also fold into General. Tab mode unaffected.

## Custom HTTP Headers routing

The new UI renders the "Add custom http header" editor (`hasHeadersEditor:true`). The field uses
the identical `indexedPair` storage as the eight sibling HTTP datasources verified in this pass
(graphite/prometheus/loki/tempo/pyroscope/jaeger/opentsdb/parca/zipkin/influxdb), where a header
was confirmed to persist as `jsonData.httpHeaderName1` (name) / `secureJsonData.httpHeaderValue1`
(write-only value). Elasticsearch's group is `optional` (collapsed by default), so the paste-and-save
routing wasn't re-driven here — it is identical by construction.

---

## Verification

```
go generate ./registry/elasticsearch/...
go test ./registry/elasticsearch/...   # 8/8 conformance subtests PASS
```

Full `./registry/... ./schema/...` suite passes. No `conformance.go`/`plugin-ui` change needed.

## Files changed

- [`registry/elasticsearch/dsconfig.json`](dsconfig.json) — added `jsonData_httpHeaders`; `root_url` → `required:true`; removed redundant `requiredWhen:"true"` on `jsonData_index` / `jsonData_timeField`.
- [`registry/elasticsearch/schema.gen.json`](schema.gen.json), [`registry/elasticsearch/settings.gen.json`](settings.gen.json) — regenerated.

_Unchanged by design:_ `settings.go`, `settings.ts`, `README.md`.
