# Loki — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered
from this entry's [`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `loki`
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/cfraniavq60w0d` (Grafana Enterprise 13.0.1, `@grafana/plugin-ui` `ConfigEditor`)
- **New UI:** `http://localhost:64433/?path=/story/configeditor-datasourceconfigwizard--tab&args=pluginType:loki` (Storybook, `ConfigEditor/DatasourceConfigWizard`)
- **Method:** Playwright captured both UIs (full-page screenshots + DOM extraction + `Save`-button console payload). The Storybook story fetches the schema from `raw.githubusercontent.com/.../registry/loki/dsconfig.json`; the local (edited) `dsconfig.json` was served to the new UI by intercepting that request with `context.route(...)`, so the screenshots reflect the local schema without pushing.
- **Result:** **Parity achieved.** One missing field was found (**Custom HTTP Headers**) and added; `fileUpload` was evaluated and correctly **not** used; the `virtual_authMethod` selector and its `effects` were exercised in both branches; all conditional fields' storage targets were verified from the save payload.

---

## TL;DR of changes

| #   | Change                                                                                          | File                             | Why                                                                                                          |
| --- | ----------------------------------------------------------------------------------------------- | -------------------------------- | ------------------------------------------------------------------------------------------------------------ |
| 1   | Added **Custom HTTP Headers** field (`jsonData_httpHeaders`) with `indexedPair` storage; added it to the `advanced-http` group | [`dsconfig.json`](dsconfig.json) | Legacy UI renders an **HTTP headers** section with an **Add header** button; new UI had no headers editor    |
| 2   | Changed `root_url` from `"requiredWhen": "true"` to `"required": true`                           | [`dsconfig.json`](dsconfig.json) | Puts **URL** into the wizard's **General** step and emits OpenAPI `required: ["url"]` (URL is unconditionally required — the Loki backend reads `settings.URL` directly) |
| 3   | Updated two embedded `instructions` that said headers were "not modeled as first-class fields"  | [`dsconfig.json`](dsconfig.json) | Keep the embedded LLM instructions truthful after change #1                                                  |
| 4   | Regenerated `schema.gen.json`, `settings.gen.json` via `go generate ./registry/loki/...`        | generated artifacts              | Keep committed artifacts in sync (guarded by `SchemaArtifactInSync`); `url` is now in the spec's `required` array |

**No Go or plugin-ui change was required.** The shared conformance walker already
skips `indexedPair` fields (`schema/conformance.go` `isIndexedPairField()`), and the
plugin-ui wizard resolver already recognises the `authentication` group id — both were
generalised during the earlier graphite parity work. `settings.go`, `settings.ts`,
`README.md`, `settings.examples.gen.json`, `schema/conformance.go`, and `plugin-ui` are
**unchanged**.

---

## Section layout

Loki's existing group taxonomy was **kept as-is** (only the new headers field was slotted
into `advanced-http`). Verified rendering top-to-bottom in the new UI tab sidebar:

| Order | Section (`id`) | `optional` | Fields (in display order) |
| --- | --- | --- | --- |
| 1 | **Connection** (`connection`) | no | URL |
| 2 | **Authentication** (`authentication`) | no | Authentication method (`virtual_authMethod`), User, Password |
| 3 | **TLS settings** (`tls-settings`) | yes | Add self-signed certificate → CA Certificate, TLS Client Authentication → ServerName / Client Certificate / Client Key, Skip TLS certificate validation |
| 4 | **Advanced HTTP settings** (`advanced-http`) | yes | Allowed cookies, Timeout, **Custom HTTP Headers** ➕ |
| 5 | **Alerting** (`alerting`) | yes | Manage alert rules in Alerting UI |
| 6 | **Queries** (`queries`) | yes | Maximum lines |
| 7 | **Derived fields** (`derived-fields`) | yes | Derived fields (array-of-objects editor) |

Notes:

- The legacy UI renders **HTTP headers** as its own top-level section (headings observed:
  `Connection`, `Authentication`, `Authentication methods`, `TLS settings`, `HTTP headers`,
  `Additional settings`). This schema places **Custom HTTP Headers** under the existing
  **Advanced HTTP settings** group, matching Grafana's "Advanced HTTP settings" grouping and
  keeping the schema's established section list unchanged. Both UIs collect the same storage
  keys.
- **Authentication** uses the `virtual_authMethod` selector (a computed discriminator) rather
  than three independent switches — see [Conditional fields & effects](#conditional-fields--effects--tested).
- The two flag fields the selector drives — `root_basicAuth` and `jsonData_oauthPassThru` —
  are tagged `managed-by:virtual_authMethod` and have **no direct control**; they are not
  referenced by any group and are written only by the selector's `effects`.

### Wizard mode: URL in the "General" step

In **wizard mode** the plugin-ui builds a synthetic first step titled **General**
(`resolveRequiredFieldsGroup`) from (a) every field marked `required: true`, plus (b) every
field in the auth group, plus their `dependsOn` parents/children.

Because the plugin-ui resolver already recognises the `authentication` group id (generalised
during the graphite work), the **General step already exists for loki regardless of the URL
change** — it is synthesised from the auth group. The `required: true` change specifically
adds **URL** into that step. Verified with a before/after capture serving each schema variant
to the wizard:

| root_url modeling | General step present | URL in General | Auth method in General |
| --- | --- | --- | --- |
| `requiredWhen: "true"` (before) | yes (`1/8`) | **no** | yes |
| `required: true` (after) | yes (`1/8`) | **yes** | yes |

The resolver does not inspect the `requiredWhen: "true"` CEL expression, so URL was absent
from General before; `required: true` makes URL unconditionally required (the backend reads
`settings.URL` directly, `pkg/loki/loki.go:66`) and emits OpenAPI `required: ["url"]` in
`settings.gen.json` instead of the `x-dsconfig-required-when` extension. Tab mode is
unaffected (the synthetic `_required` group is filtered out there — it still shows the seven
sections in order, headers editor present).

---

## Field-by-field parity

Legend: ✅ present & matching · ➕ added by this change · 🔀 conditional (revealed by `dependsOn`) · ⚙️ hidden field written by selector `effects`

| Legacy UI field | Control (legacy) | New UI (schema id) | Control (new) | Storage target | Status |
| --- | --- | --- | --- | --- | --- |
| URL | text input | `root_url` | input | `root.url` | ✅ (now `required: true`) |
| Authentication method | select (Basic auth / Forward OAuth / No auth) | `virtual_authMethod` | select | computed from `root.basicAuth` / `jsonData.oauthPassThru` | ✅ ⚙️ |
| — (basic-auth flag) | set by the auth dropdown | `root_basicAuth` | none (`managed-by`) | `root.basicAuth` | ✅ ⚙️ |
| — (OAuth-forward flag) | set by the auth dropdown | `jsonData_oauthPassThru` | none (`managed-by`) | `jsonData.oauthPassThru` | ✅ ⚙️ |
| User | text input | `root_basicAuthUser` | input | `root.basicAuthUser` | ✅ 🔀 |
| Password | password (secure) | `secureJsonData_basicAuthPassword` | secure input | `secureJsonData.basicAuthPassword` | ✅ 🔀 |
| Add self-signed certificate | switch | `jsonData_tlsAuthWithCACert` | switch | `jsonData.tlsAuthWithCACert` | ✅ |
| CA Certificate | **textarea** | `secureJsonData_tlsCACert` | secure input¹ | `secureJsonData.tlsCACert` | ✅ 🔀 |
| TLS Client Authentication | switch | `jsonData_tlsAuth` | switch | `jsonData.tlsAuth` | ✅ |
| ServerName | text input | `jsonData_serverName` | input | `jsonData.serverName` | ✅ 🔀 |
| Client Certificate | **textarea** | `secureJsonData_tlsClientCert` | secure input¹ | `secureJsonData.tlsClientCert` | ✅ 🔀 |
| Client Key | **textarea** | `secureJsonData_tlsClientKey` | secure input¹ | `secureJsonData.tlsClientKey` | ✅ 🔀 |
| Skip TLS certificate validation | switch | `jsonData_tlsSkipVerify` | switch | `jsonData.tlsSkipVerify` | ✅ |
| Allowed cookies | TagsInput | `jsonData_keepCookies` | list (string array) | `jsonData.keepCookies` | ✅ |
| Timeout | number | `jsonData_timeout` | number | `jsonData.timeout` | ✅ |
| **Custom HTTP Headers** | Add header → name input + value password | `jsonData_httpHeaders` | IndexedPair editor (key + secret value) | `jsonData.httpHeaderName<N>` / `secureJsonData.httpHeaderValue<N>` | ➕ 🔀 |
| Manage alert rules in Alerting UI | switch | `jsonData_manageAlerts` | switch | `jsonData.manageAlerts` | ✅ |
| Maximum lines | text input | `jsonData_maxLines` | input | `jsonData.maxLines` | ✅ |
| Derived fields | array editor | `jsonData_derivedFields` | array-of-objects editor (Add button) | `jsonData.derivedFields` | ✅ |

¹ **Not a discrepancy.** The `dsconfig.json` declares `ui.component: "textarea"` for the three
TLS cert fields, but the new renderer draws any `target: secureJsonData` field as a masked
secure input with a show/hide toggle (a plugin-ui renderer policy, not a schema gap). Both UIs
collect the same PEM text into the same `secureJsonData` keys.

---

## Gap found and fixed: Custom HTTP Headers

**Legacy behaviour (verified — `legacy-expand-loki`):** the editor renders an **HTTP headers**
section with an **Add header** button (`hasCustomHeaders: true`, `addHeaderBtn: true`). Adding
a header shows a header-name text input and a header-value **password** input. The
`@grafana/plugin-ui` `CustomHeaders` component persists these as indexed pairs —
`jsonData.httpHeaderName<N>` for the name and `secureJsonData.httpHeaderValue<N>` for the
(secret) value, starting at `N = 1`.

**Before:** `dsconfig.json` did not model headers at all (the old instructions explicitly said
they were "not modeled as first-class fields in this schema"), so the new UI rendered no
headers section.

**Fix (in `dsconfig.json` only):** added the `jsonData_httpHeaders` array field with an
`indexedPair` storage mapping that reproduces the exact legacy storage, plus item sub-fields for
the header name (`http.header.name`, with a header-name pattern validation) and value
(`http.header.value`). It is placed in the **Advanced HTTP settings** section:

```jsonc
{
  "id": "jsonData_httpHeaders",
  "key": "httpHeaders",
  "label": "Custom HTTP Headers",
  "valueType": "array",
  "target": "jsonData",
  "role": "http.header",
  "item": { "valueType": "object", "fields": [ /* name, value item fields */ ] },
  "storage": {
    "type": "indexedPair",
    "key":   { "target": "jsonData",       "pattern": "httpHeaderName{index}" },
    "value": { "target": "secureJsonData", "pattern": "httpHeaderValue{index}" },
    "startIndex": 1
  }
}
```

**After (verified — `newgen-loki-fixed-tab`, `loki-interact-02-header`):** the new UI renders a
**Custom HTTP Headers** editor with an **Add custom http header** button and a
key/secret-value row (`hasHeadersEditor: true`). Filling one header (name `X-Scope-OrgID`,
value `tenant-42`) and clicking **Save & Test** produced:

```json
{
  "jsonData":        { "httpHeaderName1": "X-Scope-OrgID" },
  "secureJsonData":  { "httpHeaderValue1": "tenant-42" },
  "secureJsonFields":{ "httpHeaderValue1": false }
}
```

The header **name lands in `jsonData.httpHeaderName1`** and the **value in
`secureJsonData.httpHeaderValue1`** — byte-for-byte the legacy `CustomHeaders` storage format.
The per-header `httpHeaderValue<N>` secret is dynamic, so it is intentionally **not** listed in
`SecureJsonDataKeys` (matching `settings.go`), and the conformance walker skips the `indexedPair`
field in its jsonData↔struct parity checks.

---

## `fileUpload` evaluation — not applicable to loki

The task asked to use the `fileUpload` control **if the legacy UI uses it**. It does not, for
loki:

- In Grafana 13.0.1, the legacy TLS cert fields (CA Certificate / Client Certificate / Client
  Key) render as **plain textareas** (placeholders `Begins with --- BEGIN CERTIFICATE ---` /
  `--- RSA PRIVATE KEY CERTIFICATE ---`). No file-upload button and no `<input type="file">`
  were found in the legacy DOM (`legacy-expand-loki`: `fileInputs: 0`, `uploadButtons: []`).
- The new UI's `fileUpload` component only activates when a field declares `ui.fileMapping`
  (multi-key JSON distribution, e.g. a GCP service-account file); it does not model single-PEM
  upload.

**Decision:** do **not** add `fileUpload` to any loki field. The cert fields keep their current
modeling; both UIs collect PEM text into the same `secureJsonData` keys.

---

## Conditional fields & effects — tested

### The `virtual_authMethod` selector and its `effects`

`virtual_authMethod` is a `kind: virtual` discriminator. Its value is **read** by a `computed`
storage expression (`root.basicAuth == true ? 'BasicAuth' : (jsonData.oauthPassThru == true ? 'OAuthForward' : 'NoAuth')`),
and **writes** are fanned out by `effects` to the two underlying flag fields. Both branches were
exercised in the new UI and confirmed against the `Save` console payload:

| Selection | `effects` set | User/Password revealed | Save payload (verified) |
| --- | --- | --- | --- |
| **Basic authentication** | `root_basicAuth: true`, `jsonData_oauthPassThru: false` | ✅ yes | `basicAuth: true`, `basicAuthUser: "loki-user"` (root), `secureJsonData.basicAuthPassword: "s3cret-pass"`, `jsonData.oauthPassThru: false` |
| **Forward OAuth Identity** | `root_basicAuth: false`, `jsonData_oauthPassThru: true` | n/a | `basicAuth: false`, `jsonData.oauthPassThru: true` |
| **No Authentication** (default) | both `false` | n/a | both flags absent/false |

So the selector both **renders** (default value "No Authentication" shown) and **works**:
picking a method rewrites both managed flags atomically and toggles the dependent
User/Password inputs, exactly matching the legacy dropdown's `onAuthMethodSelect` behaviour
(which always writes both `basicAuth` and `oauthPassThru`).

### `dependsOn` conditionals

| Trigger | Revealed field(s) | Verified |
| --- | --- | --- |
| `virtual_authMethod == 'BasicAuth'` | User, Password | ✅ appear on selection; `basicAuthUser` → root, `basicAuthPassword` → secureJsonData |
| `jsonData_tlsAuth == true` | ServerName, Client Certificate, Client Key | ✅ |
| `jsonData_tlsAuthWithCACert == true` | CA Certificate | ✅ |

### Representative save payload (Basic auth + one custom header)

```json
{
  "url": "http://loki.example.com:3100",
  "basicAuth": true,
  "basicAuthUser": "loki-user",
  "jsonData": {
    "oauthPassThru": false,
    "tlsAuth": false,
    "tlsAuthWithCACert": false,
    "tlsSkipVerify": false,
    "httpHeaderName1": "X-Scope-OrgID"
  },
  "secureJsonData":  { "basicAuthPassword": "s3cret-pass", "httpHeaderValue1": "tenant-42" },
  "secureJsonFields":{ "basicAuthPassword": false, "httpHeaderValue1": false }
}
```

All fields route to `root` / `jsonData` / `secureJsonData` exactly as declared.

---

## `required: true` / General-step fix

- **Before:** `root_url` used `requiredWhen: "true"` (a CEL expression the wizard resolver does
  **not** inspect), so URL was absent from the wizard's General step (verified: `urlPresent: false`).
- **After:** `required: true` (unconditionally required) puts URL into General
  (`urlPresent: true`) and emits a proper OpenAPI `required: ["url"]` in `settings.gen.json`
  (instead of the `x-dsconfig-required-when` extension). This matches the runtime contract —
  the Loki backend reads `settings.URL` directly (`pkg/loki/loki.go:66`) and `Config.Validate`
  fails on an empty URL.

Real `requiredWhen` conditions on other fields (`root_basicAuthUser`, `serverName`, the TLS
secrets) were **left unchanged** — only the literal `"true"` sentinel on `root_url` was
converted.

---

## Conformance — no change needed

Adding an `indexedPair` field did **not** require a conformance change for loki: the shared
walker already has `isIndexedPairField()` and skips such fields in `JSONDataMatchesStruct` /
`JSONDataTypesMatchStruct` (added during the graphite work). An `indexedPair` field is a logical
view over dynamically-indexed legacy keys (`httpHeaderName1`, `httpHeaderValue1`, …) that are
intentionally not modeled as Go struct fields — the SDK's `HTTPClientOptions` reads them by
prefix, and `settings.go` already documents that these dynamic secrets are omitted from
`SecureJsonDataKeys`.

---

## Verification

```
go generate ./registry/loki/...      # regenerate schema.gen.json / settings.gen.json
go test ./registry/loki/...          # PASS (8/8 conformance subtests + settings_test.go)
gofmt -l registry/loki/              # clean
go vet ./registry/loki/...           # clean
```

Conformance subtests (loki): `BaseFieldsResolved`, `SchemaRoundTrip`, `SchemaArtifactInSync`,
`SchemaSpecHasNoSecureJSON`, `ConfigSchemaValid`, `JSONDataMatchesStruct`,
`JSONDataTypesMatchStruct`, `SecureValuesMatchLoadSettings` — all **PASS**.

Playwright evidence (temp workspace `graphite-parity/`): `legacy-expand-loki` (legacy
inventory), `newgen-loki-fixed-tab` (tab mode: 7 sections + headers editor), `newgen-loki-fixed-wiz`
+ `inspect-loki-wizard-before/after` (wizard General step, URL before/after), `loki-interact`
(BasicAuth effect + header routing save payload), `loki-oauth` (OAuthForward effect).

---

## Files changed

- [`registry/loki/dsconfig.json`](dsconfig.json) — added `jsonData_httpHeaders` field and added
  it to the `advanced-http` group's `fieldRefs`; changed `root_url` from `requiredWhen: "true"`
  to `required: true`; updated two embedded `instructions` to state that custom HTTP headers are
  now modeled as a first-class field.
- [`registry/loki/schema.gen.json`](schema.gen.json), [`registry/loki/settings.gen.json`](settings.gen.json)
  — regenerated by `go generate` (`url` now in the spec's `required` array; `httpHeaders` array
  emitted under `jsonData`; `secureValues` unchanged — still the four static secrets).

_Unchanged by design:_ `settings.go`, `settings.ts`, `README.md`, `settings.examples.gen.json`,
`schema/conformance.go`, and `plugin-ui`.
