# ClickHouse — UI parity report

Parity validation between the **legacy plugin config editor** (the ClickHouse
plugin's own `CHConfigEditor` React component) and the **new schema-driven
config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered from this entry's
[`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `grafana-clickhouse-datasource` (an **HTTP / native-TCP** datasource — ClickHouse native protocol on 9000/9440 or HTTP on 8123/8443; it ships its own React config editor, not `@grafana/sql`)
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/clickhouse-test` (Grafana Enterprise 13.2.0). The instance renders the plugin's newer config-page design (banner *"You are viewing a new design for the ClickHouse configuration settings"*, behind the `newClickhouseConfigPageDesign` flag) — this is still the plugin-owned editor and uses the same v1 storage contract we model.
- **New UI:** `http://192.168.1.241:58899/iframe.html?id=configeditor-datasourceconfigwizard--tab&args=pluginType:grafana-clickhouse-datasource` (Storybook, `ConfigEditor/DatasourceConfigWizard`; also `--wizard`). **New-UI capture deferred (storybook offline)** — the Storybook host returned HTTP `000` on the first probe and was not retried per run policy.
- **Method:** Playwright captured the **legacy** UI (authenticated `admin/admin` → editor → expand collapsibles → toggle `protocol=HTTP` → expand *Optional HTTP settings*; full-page screenshots + DOM extraction of headings/labels/buttons/inputs). New-UI screenshots were skipped; the schema side is instead validated by the Go conformance suite (schema round-trip, artifact-in-sync, jsonData⇄struct parity in both directions, secure-key parity) — the same guard-rails that back every other entry.
- **Result:** **Parity achieved** for the modeled field set. **69 fields / 10 groups / 3 relationships** cover the entire legacy editor. The only change required was converting the two unconditionally-required fields (`host`, `port`) from `"requiredWhen": "true"` to `"required": true` so the wizard's synthetic **General** step pulls them in and the spec emits proper OpenAPI `required` arrays. **Custom HTTP Headers were already modeled** as a native `jsonData.httpHeaders` object array (see below) — validated, not added. There is **no file-upload** control in the legacy editor (`fileInputs: 0` in every state), so none was added.

---

## TL;DR of changes

| #   | Change | File | Why |
| --- | --- | --- | --- |
| 1   | Changed **`jsonData_host`** and **`jsonData_port`** from `"requiredWhen": "true"` → `"required": true` | [`dsconfig.json`](dsconfig.json) | Both are **unconditionally** required — the backend `Validate()` hard-fails on empty `Host` / zero `Port` (`settings.go:495-500`, mirroring upstream `isValid()` `pkg/plugin/settings.go:69-77`). The wizard's synthetic **General** step only pulls fields with `required: true`; `requiredWhen` is a CEL string the resolver does not inspect. Also emits proper OpenAPI `required: ["host","port"]` instead of the `x-dsconfig-required-when` extension. |
| 2   | Regenerated `schema.gen.json`, `settings.gen.json` via `go generate ./registry/grafana-clickhouse-datasource/...` | generated artifacts | Keep committed artifacts in sync (guarded by `SchemaArtifactInSync`). |

No changes were made to `settings.go`, `settings.ts`, `README.md`, `conformance_test.go`,
or `plugin-ui`. `settings.examples.gen.json` was unchanged by regeneration.

**Custom HTTP Headers were NOT added — they were already modeled** as a native
`jsonData.httpHeaders` object array (not an `indexedPair`). See
[Custom HTTP Headers — already modeled](#custom-http-headers--already-modeled).

### `requiredWhen` vs conditionals

`requiredWhen` appeared **exactly twice** in `dsconfig.json` — the literal
`"true"` on `host` and `port`. There were **no conditional `requiredWhen`**
expressions to preserve, and there was no pre-existing `required: true` (so no
redundant `requiredWhen` to delete). All of ClickHouse's real conditionals are
expressed with **`dependsOn`** (protocol / TLS / auth / config-mode) and were
**left untouched**:

| Field | `dependsOn` (unchanged) |
| --- | --- |
| `jsonData_path` (HTTP URL Path) | `jsonData_protocol == 'http'` |
| `jsonData_httpHeaders` (Custom HTTP Headers) | `jsonData_protocol == 'http'` |
| `jsonData_forwardGrafanaHeaders` | `jsonData_protocol == 'http'` |
| `secureJsonData_tlsCACert` (CA Cert) | `jsonData_tlsAuthWithCACert == true` |
| `secureJsonData_tlsClientCert` (Client Cert) | `jsonData_tlsAuth == true` |
| `secureJsonData_tlsClientKey` (Client Key) | `jsonData_tlsAuth == true` |
| `jsonData_signalType` (Signal type) | `jsonData_configMode == 'single-table'` |

---

## Section layout

Verified top-to-bottom in the **legacy** editor (section headings + collapsibles)
and matched to this entry's `groups`. The legacy editor groups everything under
five step-nav sections (Server and encryption / Database credentials / TLS/SSL
settings / Configuration mode / Additional settings); the schema splits the same
fields into 10 ordered groups (a finer-grained but equivalent decomposition).

| Order | Group (`id`) | `optional` | Fields |
| --- | --- | --- | --- |
| 1 | **Server** (`server`) | no | Server address\*, Server port\*, Protocol, Secure Connection, HTTP URL Path 🔀 |
| 2 | **TLS / SSL Settings** (`tls-ssl`) | no | Skip TLS Verify, TLS Client Auth, With CA Cert, CA Cert 🔀, Client Cert 🔀, Client Key 🔀 |
| 3 | **Credentials** (`credentials`) | no | Username, Password |
| 4 | **Configuration Mode** (`configuration-mode`) | no | Mode, Signal type 🔀 |
| 5 | **HTTP Headers** (`http-headers`) | yes | Custom HTTP Headers 🔀, Forward Grafana HTTP Headers 🔀 |
| 6 | **Default DB and table** (`default-db-table`) | yes | Default database, Default table |
| 7 | **Query settings** (`query-settings`) | yes | Dial Timeout, Query Timeout, Connection Max Lifetime, Max Idle Connections, Max Open Connections, Validate SQL, Suggest Map keys in filter editor |
| 8 | **Logs configuration** (`logs-config`) | yes | 11 fields (default DB/table, OTel toggle+version, column-role inputs, context columns, show-log-links) |
| 9 | **Traces configuration** (`traces-config`) | yes | 25 fields (default DB/table, OTel toggle+version, all column-role inputs, duration unit, prefixes, timestamp table suffix, show-trace-links) |
| 10 | **Additional settings** (`additional-settings`) | yes | Column Alias Tables, Custom Settings, Enable row limit, Hide table name in ad hoc filters |

Legacy step-nav headings observed (native default): `Server and encryption`,
`Database credentials`, `TLS/SSL settings` (optional), `Configuration mode`,
`Additional settings` (optional) with sub-headings `Default DB and table`,
`Query settings`, `Logs configuration`, `Traces configuration`,
`Column Alias Tables`, `Custom Settings`.

### Wizard mode: the "General" step (by construction)

Storybook was offline, so this was not re-captured for ClickHouse; the behaviour
is identical to the already-verified postgres/graphite runs. In wizard mode
`plugin-ui` builds a synthetic **General** step from (a) every field marked
`required: true`, plus (b) auth-group fields, plus `dependsOn` parents/children.

**Effect of this fix:** before, `host` and `port` used `requiredWhen: "true"`,
which the required-fields resolver does not inspect — so both were **absent**
from General (reachable only in the Server group). After `required: true` they
appear in General. Emitted OpenAPI proof (regenerated `settings.gen.json`):

```json
"jsonData": { "type": "object", "required": ["host", "port"], "properties": { … } }
```

Tab mode is unaffected (the synthetic required group is filtered out there; the
10 groups render in order).

---

## Field-by-field parity

Legend: ✅ present & matching · 🔀 conditional (revealed by `dependsOn`) · \* required

### Server + TLS + Credentials + Mode

| Legacy field | Control | New UI (schema id) | Storage target | Status |
| --- | --- | --- | --- | --- |
| Server address \* | text input | `jsonData_host` | `jsonData.host` | ✅ (**now `required: true`**) |
| Server port \* | number input | `jsonData_port` | `jsonData.port` | ✅ (**now `required: true`**) |
| Protocol | radio (Native/HTTP) | `jsonData_protocol` | `jsonData.protocol` | ✅ |
| Secure Connection | switch | `jsonData_secure` | `jsonData.secure` | ✅ |
| HTTP URL Path | text input | `jsonData_path` | `jsonData.path` | ✅ 🔀 (`protocol == 'http'`) |
| Skip TLS Verify | switch | `jsonData_tlsSkipVerify` | `jsonData.tlsSkipVerify` | ✅ |
| TLS Client Auth | switch | `jsonData_tlsAuth` | `jsonData.tlsAuth` | ✅ |
| With CA Cert | switch | `jsonData_tlsAuthWithCACert` | `jsonData.tlsAuthWithCACert` | ✅ |
| CA Cert | textarea | `secureJsonData_tlsCACert` | `secureJsonData.tlsCACert` | ✅ 🔀¹ (`tlsAuthWithCACert == true`) |
| Client Cert | textarea | `secureJsonData_tlsClientCert` | `secureJsonData.tlsClientCert` | ✅ 🔀¹ (`tlsAuth == true`) |
| Client Key | textarea | `secureJsonData_tlsClientKey` | `secureJsonData.tlsClientKey` | ✅ 🔀¹ (`tlsAuth == true`) |
| Username | text input | `jsonData_username` | `jsonData.username` | ✅² |
| Password | `SecretInput` | `secureJsonData_password` | `secureJsonData.password` | ✅ |
| Mode | radio (All databases/Single source) | `jsonData_configMode` | `jsonData.configMode` | ✅ |
| Signal type | radio (Logs/Traces) | `jsonData_signalType` | `jsonData.signalType` | ✅ 🔀 (`configMode == 'single-table'`) |

### HTTP Headers (conditional on `protocol == 'http'`)

| Legacy field | Control | New UI (schema id) | Storage target | Status |
| --- | --- | --- | --- | --- |
| Custom HTTP Headers | headers editor (Name / Value / Secure + "Add Header") | `jsonData_httpHeaders` (object array) | `jsonData.httpHeaders[]` (secure values → `secureJsonData["secureHttpHeaders.<Name>"]`) | ✅ 🔀 |
| Forward Grafana HTTP Headers to data source | switch | `jsonData_forwardGrafanaHeaders` | `jsonData.forwardGrafanaHeaders` | ✅ 🔀 |

### Additional settings (Default DB/table, Query settings, Logs, Traces, Alias, Custom, misc.)

All observed in the legacy editor and modeled inline in `dsconfig.json`:

- **Default DB/table:** Default database (`jsonData_defaultDatabase`), Default table (`jsonData_defaultTable`).
- **Query settings:** Dial Timeout (`jsonData_dialTimeout`), Query Timeout (`jsonData_queryTimeout`), Connection Max Lifetime (`jsonData_connMaxLifetime`), Max Idle Connections (`jsonData_maxIdleConns`), Max Open Connections (`jsonData_maxOpenConns`), Validate SQL (`jsonData_validateSql`), Suggest Map keys in filter editor (`jsonData_enableMapKeysDiscovery`).
- **Logs configuration** (`section: "logs"` → `jsonData.logs.*`): default DB/table, Use OTel + OTel version, Filter Time / Time / Log Level / Log Message columns, Auto-Select Columns, Context Columns (tag list), Show "View logs" links. (11 fields)
- **Traces configuration** (`section: "traces"` → `jsonData.traces.*`): default DB/table, Use OTel + version, Trace ID / Span ID / Operation Name / Parent Span ID / Service Name / Duration Time / Start Time / Tags / Service Tags / Kind / Status Code / Status Message / State / Library Name / Library Version columns, Duration Unit (select), Use Flatten Nested, Events prefix, Links prefix, Trace timestamp table suffix, Show "View trace" links. (25 fields)
- **Column Alias Tables** (`jsonData_aliasTables`, object array — Target/Alias Database+Table) with "Add Table".
- **Custom Settings** (`jsonData_customSettings`, object array — Setting/Value) with "Add custom setting".
- **Enable row limit** (`jsonData_enableRowLimit`), **Hide table name in ad hoc filters** (`jsonData_hideTableNameInAdhocFilters`).

Three fields are modeled but intentionally have **no editor UI** (not shown in
legacy, correctly carry no `ui`/group): `jsonData_version` (frontend-only stamp),
`jsonData_enableSchemaCache` and `jsonData_schemaCacheTTLSeconds` (backend-only
cache knobs). These are covered by the conformance `JSONDataMatchesStruct` test.

¹ **Not a discrepancy.** The schema declares `ui.component: "textarea"` for the
three TLS PEM fields, but because their `target` is `secureJsonData` the new
renderer draws them as masked secure inputs (same renderer policy documented for
postgres/graphite). Both UIs collect the same PEM text into the same
`secureJsonData` keys. Legacy renders them as textareas (`textareas` appear only
once the `tlsAuth` / `tlsAuthWithCACert` switches are on).

² **Minor legacy-only marker (out of scope, intentionally not changed).** The
legacy editor renders **Username** with a required `*`, but the backend
`Validate()` does **not** enforce username (only host, port, protocol —
`settings.go:492-508`). The schema follows the backend contract and leaves
`jsonData_username` optional. Marking it `required: true` would diverge from the
enforced contract the conformance model encodes, and it is outside this task's
explicit scope (host + port only), so it was left as-is.

---

## Gaps found

**None beyond the required-field fix.** The complete legacy field set is already
modeled. Unlike graphite (which was missing Custom HTTP Headers), ClickHouse
already models headers, so no field had to be added.

### Custom HTTP Headers — already modeled

ClickHouse **already** models Custom HTTP Headers as a **native
`jsonData.httpHeaders` object array** — **not** an `indexedPair`, and it did
**not** need to be added. Structure (`dsconfig.json:852-897`):

```
jsonData_httpHeaders  (valueType: array, target: jsonData, dependsOn: protocol == 'http')
  item (object) → name (string) · value (string) · secure (boolean)
```

Backed by the Go model `HttpHeader{ Name, Value, Secure }` and
`Config.HttpHeaders []HttpHeader` (`settings.go:108-112,220`), and emitted as an
`array`/`object` in `settings.gen.json`. Secure header values are written to
`secureJsonData["secureHttpHeaders.<Header Name>"]` while the plaintext
name/secure flag stays in `jsonData.httpHeaders` (the `value` is emptied when
`secure=true`) — see the schema `instructions` and `settings.go:378-392`.

**Legacy confirmation (captured):** with `protocol=HTTP` and the *Optional HTTP
settings* disclosure expanded, the legacy editor shows the **Custom HTTP
Headers** editor (Name / Value / Secure rows with an **"Add Header"** button) and
the **Forward Grafana HTTP Headers to data source** toggle:
`hasCustomHeaders: true`, `addHeaderBtn: true`, `forwardGrafanaHeaders: true`.
Under the default `protocol=native` these controls are correctly **hidden**
(`hasCustomHeaders: false`) — matching the `dependsOn: protocol == 'http'`
conditional. **Action: validated only; no schema change.**

---

## `fileUpload` evaluation — not applicable to ClickHouse

The task said to use the `fileUpload` control **only if the legacy UI uses it**.
It does not:

- Every legacy capture reported **`fileInputs: 0`** and **no upload buttons** —
  in the native default state, in the `protocol=HTTP` state, and with the headers
  editor open.
- The TLS certificate fields are inline **PEM textareas** (→ `secureJsonData`),
  gated behind the `tlsAuth` / `tlsAuthWithCACert` switches — there is no
  `<input type="file">` and no file path/upload control anywhere.
- The new UI's `fileUpload` component only activates for a field declaring
  `ui.fileMapping` (multi-key JSON distribution, e.g. a GCP service-account
  file); ClickHouse has no such field.

**Decision:** do **not** add `fileUpload` to any ClickHouse field.

---

## Conditional fields & effects — tested (legacy)

ClickHouse drives visibility with one **radio** (`protocol`) and several
**switches** (`tlsAuth`, `tlsAuthWithCACert`) plus one config-mode radio. Each
was exercised in the legacy editor:

| Scenario | Observed in legacy | Matches schema `dependsOn`? |
| --- | --- | --- |
| **`protocol = native`** (default) | HTTP URL Path **hidden**; *Optional HTTP settings* (Custom HTTP Headers + Forward Grafana HTTP Headers) **hidden** | ✅ all gated by `protocol == 'http'` |
| **`protocol = http`** | HTTP URL Path **shown**; *Optional HTTP settings* appears → **Custom HTTP Headers** editor + **Forward Grafana HTTP Headers** toggle **shown** | ✅ `path`, `httpHeaders`, `forwardGrafanaHeaders` all revealed |
| **`tlsAuth = true`** | Client Cert + Client Key textareas revealed | ✅ `secureJsonData_tlsClientCert` / `…tlsClientKey` (`tlsAuth == true`) |
| **`tlsAuthWithCACert = true`** | CA Cert textarea revealed | ✅ `secureJsonData_tlsCACert` (`tlsAuthWithCACert == true`) |
| **`configMode = single-table`** | Signal type (Logs/Traces) radio revealed | ✅ `jsonData_signalType` (`configMode == 'single-table'`) |

The datasource under test defaults to `protocol=native`, insecure, `All
databases` mode with TLS switches off — so by default the HTTP-headers, TLS-cert,
and signal-type fields are all hidden, exactly as the schema predicts.

**Effects:** ClickHouse's schema contains **no** `effects` blocks. All
conditional visibility is plain `dependsOn` CEL over the protocol radio, the two
TLS switches, and the config-mode radio; there is no virtual selector that fans
out to write multiple fields, so nothing for `effects` to model, and none were
added.

**Relationships (unchanged):** the schema records 3 — a `group` pairing
`tlsClientCert`+`tlsClientKey` (supplied together under `tlsAuth`), and two
`pair`s (`tlsAuthWithCACert`↔`tlsCACert`; `configMode`↔`signalType`).

---

## Verification

```
go generate ./registry/grafana-clickhouse-datasource/...   # regenerate schema.gen.json / settings.gen.json
go test ./registry/grafana-clickhouse-datasource/...        # ok — full package PASS
```

Conformance subtests (`TestSchemaConformance`) — **8/8 PASS**:
`BaseFieldsResolved`, `SchemaRoundTrip`, `SchemaArtifactInSync`,
`SchemaSpecHasNoSecureJSON`, `ConfigSchemaValid`, `JSONDataMatchesStruct`,
`JSONDataTypesMatchStruct`, `SecureValuesMatchLoadSettings`.

After regeneration, both `schema.gen.json` and `settings.gen.json` gained
`"required": ["host", "port"]` under `jsonData` and dropped the two
`x-dsconfig-required-when: "true"` extensions. `settings.examples.gen.json` was
unchanged.

---

## Files changed

- [`registry/grafana-clickhouse-datasource/dsconfig.json`](dsconfig.json) —
  changed `jsonData_host` and `jsonData_port` from `"requiredWhen": "true"` to
  `"required": true`. All `dependsOn` conditionals (protocol / TLS / auth /
  config-mode) were left untouched. Custom HTTP Headers were **not** modified
  (already modeled as a native array); no `fileUpload` was added.
- [`registry/grafana-clickhouse-datasource/schema.gen.json`](schema.gen.json),
  [`registry/grafana-clickhouse-datasource/settings.gen.json`](settings.gen.json)
  — regenerated by `go generate` (`host`/`port` now in the spec `required` array).

_Unchanged by design:_ `settings.go`, `settings.ts`, `README.md`,
`settings.examples.gen.json`, `conformance_test.go`, and `plugin-ui`.

---

## Notes / not dsconfig-fixable within scope

- **New-UI capture deferred (storybook offline).** `http://192.168.1.241:58899`
  returned HTTP `000` on the first probe and was not retried. Schema-side parity
  is instead assured by the conformance suite; the wizard "General" behaviour is
  by-construction identical to the verified postgres run.
- **Legacy UID differs from the task input.** The provided UID `dfrbqiwr4tukgc`
  does not exist on this instance (GCom/API `404`). The actual ClickHouse
  datasource is **`clickhouse-test`** (`grafana-clickhouse-datasource`, name
  "ClickHouse Test"), which was used for all legacy captures.
- **README staleness (not edited — out of scope).** `README.md:72,252,254` still
  describe host/port as encoded via `requiredWhen: "true"`. After this fix they
  are `required: true`. `README.md` is in the do-not-edit set, so it was left
  unchanged; a follow-up doc pass could refresh those three lines.
- **Username required marker (legacy-only, intentionally not changed).** See
  footnote ² above — legacy shows Username with `*`, but the backend does not
  enforce it, so the schema keeps it optional and outside this task's host/port
  scope.
