# Splunk Infrastructure Monitoring (SignalFx) — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered from
[`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `grafana-splunk-monitoring-datasource` (Splunk Infrastructure Monitoring / SignalFx)
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/afrbqiyglbwu8e` (Grafana Enterprise 13.2.0). The plugin ships its **own** `ConfigEditor` (`@grafana/plugin-ui` `DataSourceDescription` + `ConfigSection`/`ConfigSubSection`) — **not** `DataSourceHttpSettings`. It renders a single **Authentication** section (Access Token, Realm Name) plus an optional **Custom URLs** subsection (Metrics MetaData URL, SignalFlow URL). No Custom HTTP Headers, no TLS/cert fields, no file pickers.
- **New UI:** `http://192.168.1.241:58899/?path=/story/configeditor-datasourceconfigwizard--tab&args=pluginType:grafana-splunk-monitoring-datasource` (Storybook, `--tab` and `--wizard` stories).
- **Result:** **One fix applied** — the mandatory `secureJsonData.accessToken` was modeled with `requiredWhen:"true"` (a degenerate always-true conditional) instead of a plain `required:true`. Converted to `required:true`. No HTTP-headers or `fileUpload` changes (legacy has neither). The genuinely-conditional `jsonData.realmName` requirement was **left as `requiredWhen`** (correct).

---

## Capture status (environment-limited)

Both live captures were **unavailable in this environment**; the assessment is grounded in `go test`,
the authoritative in-repo source-of-truth (`settings.go` + the upstream `ConfigEditor.tsx` references
catalogued in [`README.md`](README.md)), and the regenerated artifacts.

- **Legacy capture — not obtainable.** The provided UID `afrbqiyglbwu8e` does not resolve on the local
  Grafana (`GET /api/datasources/uid/afrbqiyglbwu8e` → `404`), the `grafana-splunk-monitoring-datasource`
  **plugin is not installed** (`GET /api/plugins/grafana-splunk-monitoring-datasource/settings` → `404`;
  absent from the 45 installed datasource plugins), and no SignalFx instance is provisioned. The
  unauthenticated Playwright run (`capture-legacy-expand.js afrbqiyglbwu8e`) was redirected to the Grafana
  login page (`legacy-expand-splunkmon-parity.png`), so its DOM probe (`hasCustomHeaders:false`,
  `addHeaderBtn:false`, `fileInputs:0`) reflects the **login page**, not the editor, and is **not** used as
  evidence. The field inventory below is taken from the plugin's own `ConfigEditor.tsx` (documented line-by-line
  in `README.md`) and the backend `settings.go`.
- **New-UI capture deferred (storybook offline).** `http://192.168.1.241:58899/` was unreachable
  (`curl` → status `000`, single attempt, not retried). The tab/wizard behavior described below is the
  **expected** rendering derived from the (fixed) schema, not a fresh screenshot.

The `required:true` correction itself is **fully verified** by `go generate` + `go test` (below) and the
generated-artifact diff, and it matches the identical fix already applied to sibling entries
(`grafana-azurecosmosdb-datasource` `accountKey`, `grafana-newrelic-datasource` secrets, etc.).

---

## The fix

The access token is the single, **unconditional** credential: the backend hard-fails when it is empty —
`LoadSettings` returns `"invalid/empty access token"` (`pkg/models/settings.go:27-30`) and
`NewSignalFxClient` returns `"required access token is missing"` (`pkg/client/client.go:62-63`) — and the
legacy editor is the only field marked `required` (`ConfigEditor.tsx:63`). It was declared with
`"requiredWhen": "true"` rather than `"required": true`. `requiredWhen` is for *conditional*
(auth-type-gated) requirements; an unconditional `"true"` is a hard requirement and must be `required:true`
so the generated schema emits a proper secure `required` flag and the wizard folds it into the synthetic
**General** step.

| Field | Before | After |
| ----- | ------ | ----- |
| `secureJsonData_accessToken` | `"requiredWhen": "true"` | `"required": true` |

Regenerated artifacts now emit `secureValues[].accessToken.required:true` (previously the secure value had
no `required` flag; `requiredWhen:"true"` on a secure field emitted nothing, since the converter copies only
`key`/`description`/`required` for secure values — `dsconfig/convert.go:67-74`).

The conditional `jsonData_realmName` requirement
(`requiredWhen: "jsonData_urlMetricsMetadata == '' || jsonData_urlSignalflow == ''"`) is a **genuine**
condition and was intentionally **left in place** — per the instruction to leave conditional `requiredWhen`.

---

## Findings

**No Custom HTTP Headers.** SignalFx authenticates solely with an API access token sent as the
`X-SF-TOKEN` header (`pkg/client/rest.go:225`); its config editor is a bespoke `plugin-ui` form
(`ConfigEditor.tsx:52-107`), **not** `DataSourceHttpSettings`, so there is **no** Custom HTTP Headers
section. Headers are correctly **not** modeled.

**No `fileUpload`.** All inputs are text: the access token (secret text), the realm, and the two optional
URL overrides — there is **no** TLS cert/key field and **no** file picker. `fileUpload` is correctly
**not** used.

**Required fix applied** (see above). The token is *unconditionally* required, so `required:true` (not
`requiredWhen`) is the correct model.

**Secure Socks Proxy intentionally excluded.** The legacy editor renders a conditional
`ConfigSubSection "Secure Socks Proxy"` writing `jsonData.enableSecureSocksProxy`
(`ConfigEditor.tsx:110-140`); consistent with the other registry entries and the backend (which never reads
it by name), it is deliberately not modeled in this schema. No change.

## New UI — expected behavior (from schema; not re-captured, storybook offline)

**Tab mode** renders the **Authentication** section (**Access Token** `*` as a secret input with reveal
toggle; **Realm Name**, placeholder `us1`) and the optional **Custom URLs** subsection (**Metrics MetaData
URL**, **SignalFlow URL**). `hasHeadersEditor:false`. `urlPresent` is driven only by the two optional
URL-override placeholders — SignalFx has no proxied datasource root `url`; endpoints are derived from the
realm (`pkg/client/rest.go:339-353`).

**Wizard mode** — the synthetic **General** step is built from every field marked `required:true`. After
the fix, **Access Token** `*` folds into General (before the fix it would not have, because
`requiredWhen:"true"` is a CEL string the resolver stores but does not evaluate for General-step folding).
Realm Name uses a *conditional* `requiredWhen`, so it correctly stays on the group step rather than General.

## Field-by-field parity

Legend: ✅ present & matching · 🔒 secure (masked) input

| Legacy field (ConfigEditor.tsx) | schema id | target | Status |
| ------------------------------- | --------- | ------ | ------ |
| Access Token (`:63`, `required`) | `secureJsonData_accessToken` | `secureJsonData` | ✅ 🔒 **required** (FIX) — sent as `X-SF-TOKEN` |
| Realm Name (`:72`, placeholder `us1`) | `jsonData_realmName` | `jsonData` | ✅ conditional `requiredWhen` (derives base URLs) |
| Metrics MetaData URL (`:88`) | `jsonData_urlMetricsMetadata` | `jsonData` | ✅ optional (Custom URLs) |
| SignalFlow URL (`:98`) | `jsonData_urlSignalflow` | `jsonData` | ✅ optional (Custom URLs) |

Group structure matches the legacy editor: **Authentication** (Access Token + Realm Name) and the optional
**Custom URLs** subsection (both URL overrides).

## Conditional fields

- **`accessToken`** — *unconditionally* required (now `required:true`); not a real condition.
- **`realmName`** — *genuinely* conditional: required unless **both** custom URLs are set
  (`requiredWhen: "jsonData_urlMetricsMetadata == '' || jsonData_urlSignalflow == ''"`), mirroring the
  backend contract (`settings.go:170-173`; `pkg/client/rest.go:339-353` — an empty realm with no overrides
  yields the broken host `https://api..signalfx.com`). Correctly **left as `requiredWhen`**.

---

## Verification

```
go generate ./registry/grafana-splunk-monitoring-datasource/...   # regenerates .gen.json
go test     ./registry/grafana-splunk-monitoring-datasource/...   # PASS
```

`TestSchemaConformance` (BaseFieldsResolved, SchemaRoundTrip, SchemaArtifactInSync,
SchemaSpecHasNoSecureJSON, ConfigSchemaValid, JSONDataMatchesStruct, JSONDataTypesMatchStruct,
SecureValuesMatchLoadSettings) plus the hand-authored `TestLoadConfig` / `TestApplyDefaults` /
`TestValidate` all **PASS** — including the negative case asserting an empty access token is rejected,
confirming the `required:true` conversion matches the backend contract. Re-running `go generate` produces
no further drift; committed `schema.gen.json` / `settings.gen.json` / `settings.examples.gen.json` remain
in sync.

Generated-artifact effect (`git diff`):

```
  "secureValues": [
-   { "key": "accessToken" }
+   { "key": "accessToken", "required": true }
  ]
```

## Files changed

- [`registry/grafana-splunk-monitoring-datasource/dsconfig.json`](dsconfig.json) — `requiredWhen:"true"` →
  `required:true` on `secureJsonData_accessToken`. The conditional `requiredWhen` on `jsonData_realmName`
  was left unchanged.
- [`registry/grafana-splunk-monitoring-datasource/schema.gen.json`](schema.gen.json),
  [`registry/grafana-splunk-monitoring-datasource/settings.gen.json`](settings.gen.json) — regenerated by
  `go generate` (`accessToken` secure value now `required:true`). `settings.examples.gen.json` was
  regenerated with no content change.

_Unchanged by design:_ `settings.go`, `settings.ts`, `README.md`, `schema/conformance.go`, `schema/`, and
everything in `plugin-ui`. No conformance-test or plugin-ui change was required; the fix is a pure schema
correction.

> Note: `README.md` line 113 still describes the token as `requiredWhen: "true"` (its pre-fix state). It is
> now slightly stale, but README edits are out of scope for this task (and forbidden), so it was left as-is.
