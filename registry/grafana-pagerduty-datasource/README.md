# grafana-pagerduty-datasource

Declarative configuration schema for the **PagerDuty datasource plugin**
(`grafana-pagerduty-datasource`).

PagerDuty is built on the shared **OpenAPI datasource** framework (`src/openapids/` on the
frontend, `pkg/openapids/` on the backend). Its configuration UI and storage shape are driven
by a bundled OpenAPI spec (`pkg/spec.json`, the PagerDuty REST API) and a customization file
(`pkg/customization.json`) rather than a hand-written `ConfigEditor.tsx`. There is therefore no
plugin-specific config editor to read — the editor is generic and the plugin-specific behavior
lives entirely in the spec + customization.

## Upstream researched

- **Repo**: `github.com/grafana/plugins-private` (monorepo)
- **Ref**: `main`
- **Commit SHA**: `267f4937806ed6404b6628d13ae358a5d308e376` (`Wavefront: add lowpass to cspell word list (#4275)`)
- **Plugin path**: `plugins/grafana-pagerduty-datasource` (plugin version `1.2.15`, `package.json:3`)

Every value in [`dsconfig.json`](dsconfig.json) — labels, the API-key tooltip, the auth scheme
discriminator, defaults, `requiredWhen`, storage keys, storage targets, value types, the group
title, and instructions — is traceable to a specific `file:line` in the monorepo at this SHA.
See [Field provenance](#field-provenance).

To reproduce this research (paths relative to the monorepo root):

```bash
git -C <plugins-private> fetch origin && git -C <plugins-private> checkout 267f4937806ed6404b6628d13ae358a5d308e376
cd <plugins-private>/plugins/grafana-pagerduty-datasource
```

If upstream `main` has advanced past this SHA, re-diff the sources under
[Sources researched](#sources-researched) before merging changes to this entry.

## Files

| File | Purpose |
| --- | --- |
| [`dsconfig.json`](dsconfig.json) | dsconfig v1 schema — single source of truth for all config fields, groups, relationships, and instructions |
| [`settings.ts`](settings.ts) | TypeScript models: `RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig` |
| [`settings.go`](settings.go) | Go `Config` model (nested `auth` object + `DecryptedSecureJSONData`), `PluginID`, `AuthSchemeID` / `SecureJsonDataKey` typed constants, and the `LoadConfig` utility (`parse → ApplyDefaults → Validate`) |
| [`schema.go`](schema.go) | k8s-style SDK plugin schema: embeds `dsconfig.json`, converts it via `dsconfig.NewSDKSchema`, and defines `SettingsExamples` |
| [`settings_test.go`](settings_test.go) | Table tests for `LoadConfig`, `ApplyDefaults`, and `Validate` |
| [`conformance_test.go`](conformance_test.go) | Runs `schema.RunPluginTests` — the shared dsconfig conformance suite — against `dsconfig.json` |
| `schema.gen.json`, `settings.gen.json`, `settings.examples.gen.json` | Committed schema artifacts (regenerate with `go generate ./...`; the `SchemaArtifactInSync` conformance subtest guards drift) |

There is no per-entry `go.mod` — every registry entry is a subpackage of the shared
[`registry/`](..) module (`github.com/grafana/dsconfig/registry`). Import path:
`github.com/grafana/dsconfig/registry/grafana-pagerduty-datasource` (package
`pagerdutydatasource`).

## Sources researched

Read at the pinned monorepo SHA, plus external editor components at the versions the plugin's
`package.json` catalog-resolves.

### Plugin (monorepo `plugins/grafana-pagerduty-datasource` @ `267f493`)

| File | What was read |
| --- | --- |
| `src/plugin.json:5,4,3,24` | `pluginType` (`id` = `grafana-pagerduty-datasource`), `pluginName` (`name` = `PagerDuty`), `type` = `datasource`, docs link → `docURL` |
| `src/module.tsx:1-3` | Entry point: `new OpenApiDatasourcePlugin()` — the plugin is the generic OpenAPI datasource, no bespoke ConfigEditor |
| `pkg/customization.json:4-12` | Security block: `supportsNoAuth:false` (`:5`); single `api_key` scheme with `description` "PagerDuty REST API Key (prefer generating read-only key)" (`:8`) and `apiKeyPrefix:"Token token="` (`:9`) |
| `pkg/customization.json:14-18` | `/incidents` GET is marked `healthcheck:true` (`:18`) — the Save & test target |
| `pkg/spec.json:2-11` | OpenAPI 3.0.2; `info.title` = "PagerDuty API" (health-check success message uses it) |
| `pkg/spec.json:13-17` | Global `security: [{ api_key: [] }]` |
| `pkg/spec.json:164-169` | `servers`: exactly one, `https://api.pagerduty.com`, no variables |
| `pkg/spec.json:3699-3706` | `securitySchemes.api_key`: `type:"apiKey"`, `name:"Authorization"`, `in:"header"` |
| `src/openapids/types.ts:4-22` | Generic frontend `Config` (jsonData: `servers?`, `auth?{id}`, `enableSecureSocksProxy?`) and `SecureConfig` (`{[key]: string}`) |
| `src/openapids/components/config-editor/ConfigEditor.tsx:1-57` | Loads the spec + customization over resource endpoints, then renders `EditorForm` |
| `src/openapids/components/config-editor/EditorForm.tsx:53-115` | Renders `DataSourceDescription`, `Connection`, `Auth`, then an "Optional Settings" `ConfigSubSection` containing only the (feature-gated) Secure Socks Proxy switch |
| `src/openapids/components/config-editor/Connection.tsx:38-45` | Returns `null` when there is a single server with no variables — PagerDuty's case, so no connection section is rendered |
| `src/openapids/components/config-editor/Auth/Auth.tsx:35-64` | `getAuthMethods` + `onAuthMethodChange` writes `jsonData.auth.id`; `supportsNoAuth` = false so No-Auth is not offered |
| `src/openapids/components/config-editor/Auth/Auth.tsx:108-129` | `onApiKeyChange` / `onApiKeyReset` write/clear `secureJsonData["auth.<scheme>.apiKey"]` and toggle `secureJsonFields` |
| `src/openapids/components/config-editor/Auth/Auth.tsx:171-199` | `visibleMethods`; the mount `useEffect` auto-sets `auth.id` to the only method when `!supportsNoAuth && !auth.id` |
| `src/openapids/components/config-editor/Auth/Auth.tsx:221-240` | Custom method for the apiKey scheme: `label:'API key'`, `description` = customization description (fallback `'Key for accessing the API'`), renders `ApiKey` with that tooltip |
| `src/openapids/components/config-editor/Auth/ApiKey.tsx:16-31` | `InlineField label="API key"`, `tooltip={tooltip \|\| 'Key for accessing the API'}`, `SecretInput` (write-only) |
| `pkg/openapids/options.go:11-70` | Backend `Options` (jsonData `servers`, `auth.id`, per-scheme `Credentials`) and `loadOptionsFromPluginSettings`: nil jsonData → `{}` (`:38-43`), copy `DecryptedSecureJSONData` (`:69`) |
| `pkg/openapids/httpclient.go:19-49` | `authId := options.Auth.Id`; look up the scheme; for `apiKey`, read `DecryptedSecureJSONData["auth.<authId>.apiKey"]` (`:43`), prepend `apiKeyPrefix` if missing (`:44-47`), set header `securityScheme.Name` (`:48`) |
| `pkg/openapids/httpclient.go:80-124` | `ResolveBaseUrl`: empty `servers.url` → first spec server (`https://api.pagerduty.com`) |
| `pkg/openapids/httpclient.go:126-155` | `getHealthCheckRequest`: builds `GET <baseUrl>/incidents` for the `healthcheck:true` path |
| `pkg/openapids/plugin.go:27-51,55-95` | Instance factory + `CheckHealth`: 401/non-200 → error; success → "`<Info.Title>` datasource connected successfully" |
| `pkg/main.go:23-38` | `dsID = "grafana-pagerduty-datasource"`; wires the openapids `Driver` with the embedded `spec.json` + `customization.json` |
| `docs/sources/_index.md:47,64-99` | Auth guidance (read-only key recommended), and the canonical provisioning example that fixes the storage shape (`jsonData.auth.id: api_key` + `secureJsonData: auth.api_key.apiKey`) |

### External editor components

The plugin's `package.json` uses the `catalog:` protocol; versions resolve from
`plugins-private/.yarnrc.yml`.

| Component | Version (catalog) | Source consulted | What was read |
| --- | --- | --- | --- |
| `Auth`, `AuthMethodSettings`, `ConfigSection`, `ConfigSubSection`, `DataSourceDescription` | `@grafana/plugin-ui` `^0.13.1` | `github.com/grafana/plugin-ui` (`src/components/ConfigEditor/Auth/Auth.tsx`, `.../auth-method/AuthMethodSettings.tsx`) | `Auth` renders `<ConfigSection title="Authentication">`; when only one method is visible, `AuthMethodSettings` sets `hasSelect=false` and renders **no** method-selector dropdown — the subsection title becomes the method label ("API key") and only the method's own fields (the API key input) render. No storage keys are written by these wrappers themselves |
| `SecretInput`, `InlineField`, `Switch`, `Divider`, `Alert`, `Select`, `Input` | `@grafana/ui` `^11.6.7` | `github.com/grafana/grafana` `packages/grafana-ui` | `SecretInput` (write-only secret), `InlineField` (label/tooltip) — no storage keys of their own |
| `DataSourceJsonData`, `DataSourcePluginOptionsEditorProps` | `@grafana/data` `^11.6.7` | `github.com/grafana/grafana` `packages/grafana-data` | Base interface the generic `Config` extends; `onOptionsChange` storage semantics |
| `config` (feature toggles) | `@grafana/runtime` `^11.6.7` | `github.com/grafana/grafana` `packages/grafana-runtime` | Gates the (excluded) Secure Socks Proxy switch at `EditorForm.tsx:78-79` |

> Version caveat: the exact `0.13.1` git tag of `grafana/plugin-ui` was not fetchable at research
> time (tags API returned empty; likely unauthenticated rate-limiting), so the `Auth` /
> `AuthMethodSettings` sources were read from the repository's `main` branch. The two facts relied
> on — the `"Authentication"` `ConfigSection` title and the single-method "no selector, title =
> method label" behavior — are long-standing and stable across the `0.13.x` line.

## Field provenance

| Schema `id` | Storage key (full path) | Target | Label / tooltip source | Value type / default source | Notes |
| --- | --- | --- | --- | --- | --- |
| `jsonData_authId` | `jsonData.auth.id` (`key:"id"`, `section:"auth"`) | `jsonData` | No visible control — for a single scheme `AuthMethodSettings` renders no selector (`AuthMethodSettings.tsx`, `hasSelect=false`) | `string`; default `"api_key"`, `allowedValues:["api_key"]` — the only scheme (`customization.json:7`), auto-set on mount (`Auth.tsx:190-199`) | Role `auth.discriminator`; written by the editor, read by the backend (`httpclient.go:34-35`). Tagged `frontend-managed` |
| `secureJsonData_authApiKeyApiKey` | `secureJsonData["auth.api_key.apiKey"]` | `secureJsonData` | `label:"API key"` (`ApiKey.tsx:17`); `description` = tooltip "PagerDuty REST API Key (prefer generating read-only key)" (`customization.json:8` via `Auth.tsx:226,237`) | `string` (`SecretInput`, `ApiKey.tsx:23`) | Role `auth.apiKey.value`; `requiredWhen: "jsonData_authId == 'api_key'"` — the key is required for the (only) auth scheme; read at `httpclient.go:43` |

## Field inventory summary

| Schema field | Storage key | Target | Editor label | Read by backend? |
| --- | --- | --- | --- | --- |
| `jsonData_authId` | `auth.id` | `jsonData` | (none — auto-selected scheme discriminator) | Yes — selects the security scheme (`httpclient.go:34-35`) |
| `secureJsonData_authApiKeyApiKey` | `auth.api_key.apiKey` | `secureJsonData` | API key | Yes — sent as `Authorization: Token token=<key>` (`httpclient.go:42-48`) |
| `jsonData.servers` (**not modeled**) | `servers` | `jsonData` | (none — single fixed server) | Read generically, but inert: base URL always resolves to `https://api.pagerduty.com` (`spec.json:164-169`, `httpclient.go:80-124`) |
| `jsonData.enableSecureSocksProxy` (**excluded**) | `enableSecureSocksProxy` | `jsonData` | Secure Socks Proxy | Indirectly (SDK proxy options, `options.go:60-67`) — excluded per AGENTS.md |

### Frontend-only settings

`jsonData.auth.id` is written by the editor automatically (not via a user control) but **is** read
by the backend, so it is not "frontend-only". There are no purely frontend-only fields.

### Backend-only settings

None with an editor control. `jsonData.servers` is a generic-framework field the backend can read
but the PagerDuty editor never renders/writes (single server, no variables), so it is neither
editor-visible nor meaningfully backend-consumed — see [Modeling decisions](#modeling-decisions).

## Where the types are defined

The configuration types are spread across the plugin and its dependencies; several come from the
shared OpenAPI-datasource framework and Grafana libraries rather than PagerDuty-specific code.

### Frontend (TypeScript)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `Config` (jsonData: `servers?`, `auth?{id}`, `enableSecureSocksProxy?`), `SecureConfig` (`{[key]: string}`) | `src/openapids/types.ts:4-22` | plugin — shared OpenAPI-datasource framework (`plugins-private`) |
| `Customization` (`spec.components.security.schemes[].description`, …) | `src/openapids/types.ts:74-97` | plugin — framework |
| `DataSourceJsonData` (base interface `Config` extends) | `packages/grafana-data/src/types/datasource.ts` | `@grafana/data` `^11.6.7` (grafana/grafana) |
| `DataSourcePluginOptionsEditorProps` | `packages/grafana-data/src/types/datasource.ts` | `@grafana/data` `^11.6.7` |

The PagerDuty-specific shape of these generic types (auth scheme = `api_key`, secret key =
`auth.api_key.apiKey`) is fixed by `pkg/spec.json` + `pkg/customization.json`, not by a TypeScript
type. `settings.ts` narrows the generic types to that shape.

### Backend (Go)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `Options` (jsonData `servers`, `auth.id`, per-scheme `Credentials`), `loadOptionsFromPluginSettings` | `pkg/openapids/options.go:11-70` | plugin — framework (`plugins-private`) |
| `Customization` (`components.security.schemes[].{description,apiKeyPrefix}`), `loadCustomization` | `pkg/openapids/customization.go:46-77` | plugin — framework |
| `httpClient` (reads `auth.<scheme>.apiKey`, sets the `Authorization` header) | `pkg/openapids/httpclient.go:19-78` | plugin — framework |
| `backend.DataSourceInstanceSettings` (carries `JSONData`, `DecryptedSecureJSONData`, and unused root fields) | `backend/common.go` | `github.com/grafana/grafana-plugin-sdk-go` |

This entry flattens that spread into a single Go `Config` (a nested `auth.id` + a
`DecryptedSecureJSONData` map) plus an `AuthSchemeID` and a `SecureJsonDataKey` typed constant
list. `settings.ts` keeps the three canonical TypeScript types; `RootConfig` is a blank object
because PagerDuty stores nothing at the datasource root level.

## Modeling decisions

- **Nested `jsonData.auth.id` via `section`.** The discriminator is stored at `jsonData.auth.id`,
  so the field uses `key:"id"` + `section:"auth"` (`dsconfig` nests sections into objects,
  `dsconfig/convert.go:122-158`). The Go `Config` mirrors this with a nested `AuthConfig{ID}`
  under `json:"auth"`, which the conformance walker flattens to the `auth.id` leaf
  (`schema/conformance.go:285-354`).
- **Dotted secure key kept verbatim.** The framework namespaces secrets by scheme id, so the
  secret's `key` is the literal `auth.api_key.apiKey` (`httpclient.go:43`), and
  `SecureJsonDataKeyAPIKey` uses the same string. Confirmed by the official provisioning example
  (`docs/sources/_index.md:97-98`).
- **Discriminator has no `ui` and is not grouped.** For a single scheme the editor renders no
  method selector (`AuthMethodSettings`, `hasSelect=false`), so `jsonData_authId` carries no `ui`
  block and sits outside the group (mirroring how the gold-standard GitHub entry leaves its
  non-visible `cachingEnabled` field out of any group). It is still modeled because the backend
  reads it and provisioning must set it. Tagged `frontend-managed`.
- **`requiredWhen` encodes the backend contract.** The API key is `requiredWhen:
  "jsonData_authId == 'api_key'"` — the working-datasource contract, not an editor marker (the
  editor shows no required asterisk).
- **`servers` not modeled.** PagerDuty's spec has one server and no variables, so the editor never
  writes `jsonData.servers` (`Connection.tsx:43-45`) and the base URL always resolves to
  `https://api.pagerduty.com` (`httpclient.go:80-124`). Modeling it would add an inert nested
  object; instead it is omitted from both `dsconfig.json` and Go `Config`, and `json.Unmarshal`
  silently ignores it if present (covered by a `settings_test.go` case).
- **Secure Socks Proxy excluded.** `jsonData.enableSecureSocksProxy` (`EditorForm.tsx:98-111`) is
  omitted per AGENTS.md.
- **`RootConfig` is a blank object.** Nothing lives at the datasource root level; the backend
  reads only `jsonData` + `secureJsonData`.
- **`LoadConfig` is stricter than the upstream loader.** The generic
  `loadOptionsFromPluginSettings` never validates. `LoadConfig` mirrors its **parse** verbatim
  (nil jsonData → `{}`, `options.go:38-43`) but then runs `ApplyDefaults` (default `auth.id` to
  `api_key`, mirroring the editor mount effect) and `Validate` (require the `api_key` scheme + a
  non-empty key). This is the intended shape for the plugin's own loader to sync to.

## SDK plugin schema and k8s-style examples (`schema.go`)

`NewSchema()` assembles the `grafana-plugin-sdk-go` `pluginschema.PluginSchema` bundle (served by
Grafana's datasource API server as `{apiVersion}.json`, `v0alpha1` today) from the embedded
`dsconfig.json`: `jsonData.auth.id` becomes the OpenAPI settings `spec` (a nested `auth` object),
and the API key becomes a `secureValues` entry (never part of the spec).

`SettingsExamples()` provides the default configuration plus the single auth/connection variant.
Each example is a full instance-settings object with the secret under `secureJsonData` using an
**obviously-fake angle-bracket placeholder**:

| Example | jsonData | `secureJsonData["auth.api_key.apiKey"]` |
| --- | --- | --- |
| `""` (default) | `auth.id = api_key` | `""` (empty — fails `Validate`, by design) |
| `apiKey` | `auth.id = api_key` | `<your-pagerduty-api-token>` |

## `LoadConfig` utility (`settings.go`)

`LoadConfig(ctx, settings) (Config, error)` runs the full three-phase flow and returns a
fully-defaulted, validated `Config`:

1. **Parse** — nil/empty `JSONData` → `{}` (mirrors `pkg/openapids/options.go:38-43`), unmarshal
   `jsonData` into `Config` (recovers `auth.id`), copy decrypted secrets by known key into
   `DecryptedSecureJSONData`.
2. **`ApplyDefaults`** — default an empty `auth.id` to `api_key` (editor mount parity,
   `Auth.tsx:190-199`).
3. **`Validate`** — require the `api_key` scheme and a non-empty API key secret; joined errors.

Everything is logged via `backend.Logger.FromContext(ctx)` with `datasource_uid`,
`datasource_name`, and `plugin` labels. `(*Config).ApplyDefaults()` and `(Config).Validate()` are
exported separately for callers that assemble a `Config` themselves (provisioning preview,
schema-example round-trip, tests distinguishing parse-level from policy-level errors).

## Upstream findings

Potential bugs, misleading UX, and consistency issues found while researching. All preserved in
the schema — it records what the plugin **does**, not what it **should** do.

1. **The secret key name leaks the internal scheme id.** Provisioning must use the literal,
   dotted, framework-internal key `secureJsonData["auth.api_key.apiKey"]`
   (`httpclient.go:43`, `docs/sources/_index.md:98`). It is unusual and easy to get wrong (the
   `api_key` segment is the OpenAPI scheme name from `spec.json:3700`, not a user-facing label).
2. **No auth is silently unauthenticated, not rejected.** With an empty/unknown `auth.id` the
   backend builds a plain HTTP client (`httpclient.go:38-40`) and only fails at health-check time
   (HTTP 401). The generic loader performs no validation, so a misconfigured datasource "loads"
   fine. `Validate` in this entry closes that gap.
3. **The API-key input shows no placeholder and no "required" marker**, even though the datasource
   cannot work without it (`supportsNoAuth:false`). The requirement is encoded via `requiredWhen`.
4. **`apiKeyPrefix` is applied with a plain `strings.HasPrefix` guard** (`httpclient.go:45-47`): if
   a user pastes a key that already begins with `Token token=`, it is not double-prefixed — but a
   key that legitimately begins with that literal substring would be mishandled. Practically
   harmless for PagerDuty keys.
5. **US region only.** The bundled spec defines a single server, `https://api.pagerduty.com`
   (`spec.json:164-169`); the EU region (`api.eu.pagerduty.com`) is not selectable, and there is
   no URL/host setting.
6. **`docURL` is taken from `src/plugin.json:24`** (`info.links[0]` "Docs"); other `info.links`
   point at `grafana/grafana` and the org support page (`plugin.json:26-27`), which are not the
   plugin docs.

## Validation performed

- `dsconfig.ParseAndResolveSchemaJSON` + `Schema.Validate()` (Go validator in this repo) — passes
  (via the `ConfigSchemaValid` conformance subtest).
- JSON Schema validation of `dsconfig.json` against [`dsconfig/schema.json`](../../dsconfig/schema.json)
  (strict, `additionalProperties:false`) — passes (0 errors).
- `go generate ./...` (regenerates the three artifacts), then `gofmt -l .`, `go vet ./...`,
  `go build ./...`, `go test ./...` inside [`registry/`](..) — all clean. The conformance suite
  guards: no `secureJsonData` in the settings spec; `secureValues` == `["auth.api_key.apiKey"]`;
  `jsonData` fields match the `Config` json tags in both directions (including the nested
  `auth.id`); the `""` default example exists; every example has `jsonData` + a `secureJsonData`
  using only known secret keys; artifact drift.
- `settings_test.go`: `LoadConfig` for the default + apiKey examples, empty settings, malformed
  jsonData, unknown scheme, and ignored `servers`/`enableSecureSocksProxy`; plus `ApplyDefaults`
  and `Validate` tables.
- `tsc --noEmit --strict` on `settings.ts` — clean.
- The pre-existing `dsconfig` and `schema` workspace modules still build and test — clean.
- Secret scan: the only credential placeholder committed is the obviously-fake
  `<your-pagerduty-api-token>`; no realistic token shapes or hex/base64 blobs.
