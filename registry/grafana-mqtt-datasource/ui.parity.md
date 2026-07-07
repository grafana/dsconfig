# MQTT — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered from
this entry's [`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `grafana-mqtt-datasource`
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/efrbqixqsdo8wa` (Grafana Enterprise; the MQTT plugin's own `ConfigEditor` — a **Connection** section (URI, Client ID) and an **Authentication** section (Username / Password + three TLS toggles that reveal the TLS CA / client cert / client key inputs). No `DataSourceHttpSettings`, no Custom HTTP Headers, no file pickers.)
- **New UI:** `http://192.168.1.241:58899/iframe.html?args=pluginType:grafana-mqtt-datasource&id=configeditor-datasourceconfigwizard--tab` (Storybook, `ConfigEditor/DatasourceConfigWizard`), also the `--wizard` story.
- **Method:** Playwright captured the legacy UI (`legacy-expand-mqtt-verify.png/.json`) and drove the new UI. The Storybook story fetches the schema from `raw.githubusercontent.com/.../registry/grafana-mqtt-datasource/dsconfig.json`; the local (edited) `dsconfig.json` was served to the new UI by intercepting that request with `context.route(...)`, so the captures reflect the local schema without pushing.
- **Result:** **Parity achieved.** The single **unconditionally-required** field (`jsonData.uri`) was corrected from a non-canonical `requiredWhen: "true"` to `required: true`, so it now folds into the wizard's synthetic **General** step and emits a proper OpenAPI `required: ["uri"]`. **Custom HTTP Headers** and **`fileUpload`** were evaluated and correctly **not** used — the legacy editor has neither (`hasCustomHeaders:false`, `addHeaderBtn:false`, `fileInputs:0`, `uploadButtons:[]`). The TLS certificate inputs are gated by editor-only visibility toggles (`dependsOn`), which are correct and were left untouched — there is **no** conditional `requiredWhen` to reconcile.

---

## TL;DR of changes

| #   | Change | File | Why |
| --- | --- | --- | --- |
| 1   | Changed `jsonData_uri` from `requiredWhen: "true"` → `required: true` | [`dsconfig.json`](dsconfig.json) | The broker URI is **unconditionally** required — the backend `Validate` hard-fails an empty URI with "mqtt broker URI (jsonData.uri) is required" (`settings.go:173-174`), and `opts.AddBroker(o.URI)` is called unconditionally with no fallback (`pkg/mqtt/client.go:46`). `requiredWhen: "true"` is a non-canonical CEL expression the wizard's General-step resolver does **not** inspect, so the URI was **not** folded into General and the generated spec carried only the `x-dsconfig-required-when: "true"` extension. `required: true` is the canonical form: it folds `uri` into the **General** step and emits `jsonData.required: ["uri"]`. |
| 2   | Regenerated `schema.gen.json`, `settings.gen.json` via `go generate ./registry/grafana-mqtt-datasource/...` | generated artifacts | Keep committed artifacts in sync (guarded by `SchemaArtifactInSync`). |

No changes were made to `settings.go`, `settings.ts`, `README.md`, `conformance_test.go`, `schema/conformance.go`, `settings.examples.gen.json`, or `plugin-ui`. The schema change flows through `dsconfig.json`; the rest is produced by `go generate`. **No conformance-test or plugin-ui change was required.**

---

## Section layout

The `groups` taxonomy was left unchanged; only the `required` flag on `uri` was corrected. Verified
rendering top-to-bottom in the new UI (tab mode, `newgen-mqtt-tab.json` + screenshot), which matches
the legacy section order (legacy headings: `["Connection", "Authentication"]`):

| Order | Section (`id`) | `optional` | Fields (in display order) |
| --- | --- | --- | --- |
| 1 | **Connection** (`connection`) | no | URI \*, Client ID |
| 2 | **Authentication** (`authentication`) | no | Username, Password, Use TLS Client Auth, Skip TLS Verification, With CA Cert |
| 3 | **TLS Configuration** (`tls-configuration`) | yes | TLS CA Certificate, TLS Client Certificate, TLS Client Key |

Notes:

- The **TLS Configuration** group is `optional` and renders collapsed (with an `Optional` badge) in tab
  mode; its three certificate inputs are additionally gated by the `Use TLS Client Auth` /
  `With CA Cert` toggles via `dependsOn` (see [Conditional fields](#conditional-fields--dependson-visibility-gating)).
- The new UI shows the header **"Fields marked with \* are required"** and renders a required asterisk
  on exactly **URI\*** (`verify-mqtt-uri-tab.json`: `uriRequiredAsterisk:true`, `requiredHint:true`).
  Tab capture: `hasHeadersEditor:false`, sections `["Connection","Authentication"]` (+ the optional
  **TLS Configuration** accordion).

### Wizard mode: URI in the "General" step

In **wizard mode** the plugin-ui builds a synthetic first step titled **General**
(`resolveRequiredFieldsGroup`) from (a) every field marked `required: true`, plus (b) the fields of the
auth group (`id: "authentication"`), plus their `dependsOn` parents/children. `uri` lives in the
**`connection`** group — **not** the auth group — so its presence in General is driven purely by the
`required` marker. This makes the fix **visibly** change the wizard (unlike a field that already lives
in the auth group):

- **Before (`requiredWhen: "true"`, `verify-mqtt-uri-wiz-before.json`):** `requiredWhen` is a CEL
  expression the resolver does not inspect, so on the General step (**1/4**) `uriPresent:false` and
  `emptyBlocked:false` — the URI was **not** in General and **Save & Test was not gated** on it.
- **After (`required: true`, `verify-mqtt-uri-wiz.json`):** the wizard opens on **General (1/4)**
  containing the **URI\*** input (`uriPresent:true`, `uriRequiredAsterisk:true`), and **Save & Test is
  blocked while URI is empty** (`emptyBlocked:true`). Filling it and saving routes the value to
  `jsonData.uri` (see below).

---

## The required-field fix in detail

`uri` is a `target: "jsonData"` field, so the converter routes it through `fieldToSpecSchema`. With
the non-canonical `requiredWhen: "true"`, `applyConditions` emitted the `x-dsconfig-required-when: "true"`
extension but **did not** add the field to the object's `required` array — so neither the generated
contract nor the wizard's General-step resolver treated `uri` as required.

**After (`required: true`)** the generated artifacts (`schema.gen.json` / `settings.gen.json`) emit:

```json
"jsonData": {
  "type": "object",
  "required": [
    "uri"
  ],
  "properties": {
    "uri": {
      "type": "string"
    }
  }
}
```

The `x-dsconfig-required-when: "true"` extension is gone and `uri` is now in the canonical `required`
array. This matches the backend's unconditional requirement (`settings.go:173-174`) and mirrors the
established idiom for an unconditionally-required `jsonData` field (e.g. Honeycomb's `hostname` / `team`).

---

## Field-by-field parity

Legend: ✅ present & matching · ✏️ corrected by this change · 🔀 conditional (revealed by `dependsOn`) · 🔒 write-only secret

| Legacy UI field | Control (legacy) | New UI (schema id) | Control (new) | Storage target | Status |
| --- | --- | --- | --- | --- | --- |
| URI | text input | `jsonData_uri` | input (placeholder `tcp:// / tls:// / ws://`) | `jsonData.uri` (**required**) | ✅ ✏️ |
| Client ID | text input | `jsonData_clientID` | input | `jsonData.clientID` (optional) | ✅ |
| Username | text input | `jsonData_username` | input | `jsonData.username` (optional) | ✅ |
| Password | password (secure) | `secureJsonData_password` | secure input | `secureJsonData.password` (optional) | ✅ 🔒 |
| Use TLS Client Auth | switch | `jsonData_tlsAuth` | switch | `jsonData.tlsAuth` (editor toggle) | ✅ |
| Skip TLS Verification | switch | `jsonData_tlsSkipVerify` | switch | `jsonData.tlsSkipVerify` | ✅ |
| With CA Cert | switch | `jsonData_tlsAuthWithCACert` | switch | `jsonData.tlsAuthWithCACert` (editor toggle) | ✅ |
| TLS CA Certificate | textarea (secure) | `secureJsonData_tlsCACert` | secure textarea | `secureJsonData.tlsCACert` | ✅ 🔀 🔒 (`tlsAuthWithCACert == true`) |
| TLS Client Certificate | textarea (secure) | `secureJsonData_tlsClientCert` | secure textarea | `secureJsonData.tlsClientCert` | ✅ 🔀 🔒 (`tlsAuth == true`) |
| TLS Client Key | textarea (secure) | `secureJsonData_tlsClientKey` | secure textarea | `secureJsonData.tlsClientKey` | ✅ 🔀 🔒 (`tlsAuth == true`) |

Notes:

- **URI** is a paho broker URI (`tcp://` / `tls://` / `ws://` / `wss://`), **not** an HTTP URL — the
  generic `urlPresent` heuristic reports `false` because it only matches `http`/`localhost`/`9090`
  placeholders; the field is nonetheless present and required (confirmed by placeholder match
  `input[placeholder*="tcp://"]`).
- **Password / TLS cert material** render as masked secure inputs with a show/hide + reset toggle (the
  renderer draws any `target: "secureJsonData"` field that way). All values collect into the same
  `secureJsonData.*` keys the backend reads.
- `jsonData.tlsAuth` and `jsonData.tlsAuthWithCACert` are **editor-visibility toggles only** — the Go
  backend does not read them (`pkg/mqtt/client.go:26-35`); it decides TLS behavior purely by whether the
  corresponding secrets are non-empty. They are modeled because the editor writes them into `jsonData`.

### Save-payload storage-target validation

Filling the URI and clicking **Save & Test** in both tab and wizard modes routes the value to
`jsonData.uri` (`verify-mqtt-uri-tab.json` / `verify-mqtt-uri-wiz.json`,
`payloadURI:"tcp://broker.example.com:1883"`) — exactly the legacy storage shape the backend reads
(`pkg/plugin/datasource.go:60-83` → `jsonData.uri` → `opts.AddBroker(o.URI)`).

---

## `fileUpload` evaluation — not applicable to MQTT

The task asked to use the `fileUpload` control **only if the legacy UI uses it**. It does not:

- The legacy MQTT editor enters TLS certificate material as **PEM text in textareas** (CA cert / client
  cert / client key), not via file pickers. The legacy DOM capture (`legacy-expand-mqtt-verify.json`)
  reports `fileInputs: 0`, `uploadButtons: []`.
- The new UI's `fileUpload` component only activates for `ui.fileMapping` (multi-key JSON distribution,
  e.g. a GCP service-account file); nothing in MQTT needs it.

**Decision:** do **not** add `fileUpload` to any MQTT field.

---

## Custom HTTP Headers — not applicable to MQTT

The legacy editor has **no** Custom HTTP Headers section (`hasCustomHeaders:false`, `addHeaderBtn:false`
in `legacy-expand-mqtt-verify.json`). MQTT is a paho broker connection (TCP / TLS / WebSocket), not an
HTTP datasource — there is no user-managed header editor. Headers are correctly **not** modeled, and the
new UI shows no headers editor in either mode (`hasHeadersEditor:false` in `newgen-mqtt-tab.json`,
`verify-mqtt-uri-tab.json`, and `verify-mqtt-uri-wiz.json`). No change.

---

## Conditional fields — `dependsOn` visibility gating

MQTT has **no auth discriminator** and **no `effects`**. Its only conditionals are the two TLS
visibility gates, which are `dependsOn` expressions (not `requiredWhen`) and are **correct as-is**:

- `secureJsonData_tlsCACert` → `dependsOn: "jsonData_tlsAuthWithCACert == true"` (the **With CA Cert** switch).
- `secureJsonData_tlsClientCert` / `secureJsonData_tlsClientKey` → `dependsOn: "jsonData_tlsAuth == true"` (the **Use TLS Client Auth** switch).

These reproduce the legacy editor gating (`src/ConfigEditor.tsx:107-122`) and were left untouched. The
`uri` field's `requiredWhen: "true"` was a literal always-true flag (now correctly `required: true`),
**not** a genuine condition — so no real conditional `requiredWhen` expression was removed or left behind.
The `tlsClientCert`/`tlsClientKey` pair-requirement ("provide both together") is enforced by the backend
(`pkg/mqtt/client.go:66-73`, `settings.go:177-184`) and documented in the schema `relationships`, not as
a schema `requiredWhen` (it is a mutual pairing, not an unconditional requirement).

---

## Conformance (no change required)

The change only flips `uri` from a dropped `requiredWhen` to `required: true`. The shared conformance
suite already understands `required` on `jsonData` fields, so **no conformance change was needed**.
`SchemaSpecHasNoSecureJSON` still passes (no secret leaks into the spec), `JSONDataMatchesStruct` /
`JSONDataTypesMatchStruct` still match the `Config` struct (`uri string`), and
`SecureValuesMatchLoadSettings` still matches the four secure keys in `SecureJsonDataKeys`
(`settings.go:45-50`).

---

## Verification

```
go generate ./registry/grafana-mqtt-datasource/...              # regenerate schema.gen.json / settings.gen.json
go test ./registry/grafana-mqtt-datasource/...                  # PASS
go test ./registry/grafana-mqtt-datasource/... ./schema/...     # PASS (no regressions)
```

`TestSchemaConformance` subtests — all **8/8 PASS**:
`BaseFieldsResolved`, `SchemaRoundTrip`, `SchemaArtifactInSync`, `SchemaSpecHasNoSecureJSON`,
`ConfigSchemaValid`, `JSONDataMatchesStruct`, `JSONDataTypesMatchStruct`, `SecureValuesMatchLoadSettings`.

The hand-authored `TestLoadConfig` (15), `TestApplyDefaults` (2), and `TestValidate` (6) suites also pass
unchanged — including the negative cases (`missing_uri_errors`, `empty_URI_errors`) that assert an empty
URI is rejected, confirming the `required: true` conversion matches the backend contract.

Generated-spec delta: `jsonData.required: ["uri"]` added; `x-dsconfig-required-when: "true"` removed from
`uri` (both `settings.gen.json` and `schema.gen.json`); nothing else changed.

Playwright evidence (in the shared capture dir):

- `legacy-expand-mqtt-verify` (legacy inventory, UID `efrbqixqsdo8wa`) — headings `["Connection","Authentication"]`; `hasCustomHeaders:false`, `addHeaderBtn:false`, `fileInputs:0`, `uploadButtons:[]`.
- `newgen-mqtt-tab` / `verify-mqtt-uri-tab` (tab) — sections Connection / Authentication / TLS Configuration `Optional`; `hasHeadersEditor:false`; **URI\*** present + required asterisk; save routes to `jsonData.uri`.
- `verify-mqtt-uri-wiz` (wizard, after) — opens on **General 1/4** with required **URI\***; **Save & Test blocked on empty URI**; save routes to `jsonData.uri`; `hasHeadersEditor:false`, `fileInputs:0`.
- `verify-mqtt-uri-wiz-before` (wizard, before — original `requiredWhen:"true"` schema) — **URI not in General 1/4** (`uriPresent:false`) and **not gated** (`emptyBlocked:false`), confirming the fix has a real, visible effect.

---

## Files changed

- [`registry/grafana-mqtt-datasource/dsconfig.json`](dsconfig.json) — changed `jsonData_uri` from
  `requiredWhen: "true"` to `required: true`. No other field carried a `requiredWhen`, so none was left
  in place.
- [`registry/grafana-mqtt-datasource/schema.gen.json`](schema.gen.json),
  [`registry/grafana-mqtt-datasource/settings.gen.json`](settings.gen.json) — regenerated by
  `go generate` (`jsonData.required: ["uri"]` added; `x-dsconfig-required-when: "true"` removed from `uri`).

_Unchanged by design / constraint:_ `settings.go`, `settings.ts`, `README.md`,
`settings.examples.gen.json`, `conformance_test.go`, `schema/conformance.go`, and everything in `plugin-ui`.

_Nothing reported as out-of-scope / unfixable:_ the only requested fix (required-field / General-step)
was fully achievable through `dsconfig.json` alone. Custom HTTP Headers and `fileUpload` are correctly
n/a for this plugin (legacy has neither), so no `settings.go`/conformance/plugin-ui coordination was
needed.
