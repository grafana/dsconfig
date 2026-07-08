# grafana-zendesk-datasource

Configuration schema for the [Zendesk datasource plugin](https://grafana.com/docs/plugins/grafana-zendesk-datasource)
(`grafana-zendesk-datasource`), which lives in the [grafana/plugins](https://github.com/grafana/plugins)
monorepo.

Unlike a typical datasource it has **no hand-written `ConfigEditor.tsx` and no per-plugin backend
`Settings` model**. Both the config editor and the backend are provided by shared packages and
specialized by the plugin's `src/spec.ts`:

- **Backend SDK**: `sdk/pluginspec` (`github.com/grafana/plugins/sdk/pluginspec`) — parses the spec and
  the stored config, builds HTTP service clients, applies auth.
- **Frontend package**: `packages/declarative-plugin` (`@grafana/declarative-plugin`, `v0.0.2`) — renders
  the config editor and query editor from the spec.

So the "config editor" this schema captures is the **generic shared editor specialized by the Zendesk
spec**, and the storage shape is the shared service-keyed model, not something bespoke to Zendesk.

## Upstream researched

- **Repo**: `github.com/grafana/plugins` (monorepo; plugin lives under `plugins/grafana-zendesk-datasource/`)
- **Ref**: `main`
- **Commit SHA**: `4b176ec1f74d80c231be2aeb1ce4713c833a76af` (2026-07-02, `Fix publishing docs only (#745)`)

To reproduce this research:

```bash
git clone https://github.com/grafana/plugins
cd plugins
git checkout 4b176ec1f74d80c231be2aeb1ce4713c833a76af
```

If upstream `main` has advanced past this SHA, re-diff the sources under [Sources researched](#sources-researched)
— in particular the shared code in `packages/declarative-plugin/` and `sdk/pluginspec/`, which is
common to every plugin in this monorepo — before merging changes to this entry.

## Files

| File | Purpose |
| --- | --- |
| [`dsconfig.json`](dsconfig.json) | dsconfig v1 schema — single source of truth for all config fields, groups, relationships, and instructions |
| [`settings.ts`](settings.ts) | TypeScript models: `RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig` |
| [`settings.go`](settings.go) | Go `Config` model (spec-specific projection of the framework's service-keyed jsonData), `PluginID`, `AuthMethodID`/`SecureJsonDataKey` typed constants, and the `LoadConfig` utility |
| [`schema.go`](schema.go) | k8s-style SDK plugin schema: embeds `dsconfig.json`, converts it via `dsconfig.NewSDKSchema`, and defines `SettingsExamples` |
| [`settings_test.go`](settings_test.go) | Table tests for `LoadConfig`, `ApplyDefaults`, and `Validate` |
| [`conformance_test.go`](conformance_test.go) | Runs `schema.RunPluginTests` — the shared dsconfig conformance suite — against `dsconfig.json` |
| [`schema.gen.json`](schema.gen.json), [`settings.gen.json`](settings.gen.json), [`settings.examples.gen.json`](settings.examples.gen.json) | Committed schema artifacts (regenerate with `go generate ./...`; the `SchemaArtifactInSync` conformance subtest guards drift) |

There is no per-entry `go.mod` — every registry entry is a subpackage of the shared
[`registry/`](..) module (`github.com/grafana/dsconfig/registry`).

## Sources researched

All paths are relative to the grafana/plugins repo root at the pinned SHA.

### Plugin-specific

| File | What was read |
| --- | --- |
| `plugins/grafana-zendesk-datasource/src/plugin.json:3-5,25-26` | `type`/`name`/`id` → `pluginType`/`pluginName`; `info.links[0]` ("Docs") → `docURL` |
| `plugins/grafana-zendesk-datasource/src/spec.ts:8` | `services: ['zendesk']` — the single service id |
| `plugins/grafana-zendesk-datasource/src/spec.ts:10-15` | `$defs.variables.subdomain` — `name` ("Subdomain") and `description` |
| `plugins/grafana-zendesk-datasource/src/spec.ts:16-32` | `$defs.authMethods.basic_auth` — type `basic`, user label "Email"/placeholder, password label "API Token"/placeholder/description |
| `plugins/grafana-zendesk-datasource/src/spec.ts:33-41` | `$defs.servers.zendesk_api` — URL `https://{subdomain}.zendesk.com/api/v2/`, variable ref `subdomain` (not `required`), auth method `basic_auth` |
| `plugins/grafana-zendesk-datasource/src/spec.ts:42-58` | `$defs.services.zendesk` — name "Tickets", type `rest`, server `zendesk_api` |

### Shared framework (common to every plugin in the monorepo)

| File | What was read |
| --- | --- |
| `sdk/pluginspec/pluginspec.go:9-114` | `Spec`/`Service`/`Server`/`Variable`/`AuthMethod`/`AuthType` definitions and the auth-type vocabulary |
| `sdk/pluginspec/pluginclient/config.go:1-37` | `JsonData` (`services`, `variables`, `enableSecureSocksProxy`) and `ServiceConfig` (`disabled`, `server.id`, `auth.*`) — the storage model |
| `sdk/pluginspec/pluginclient/pluginclient.go:21-127` | `New`: jsonData parse; **server.id defaults to the first server; auth.id defaults to the first auth method**; secret keys read as `<serviceId>.password` / `.token` / `.clientSecret` / `.apiKey.<k>` / `.tls.*`; lenient (no hard-fail on missing creds) |
| `sdk/pluginspec/pluginclient/serviceclient.go:241-277` | `applyAuth`: basic auth requires non-empty username and password **at request time** |
| `sdk/pluginspec/pluginclient/serviceclient.go:301-327` | `validateVariables`: only variables whose ref is `required: true` are enforced (subdomain is **not**) |
| `sdk/pluginspec/pluginclient/serviceclient.go:279-299` | `getBaseurl`: builds the server URL, substituting variables like `{subdomain}` |
| `packages/declarative-plugin/src/components/config-editor/EditorForm.tsx:16-113` | Top-level editor: writes `jsonData.services.<id>` and `jsonData.variables.<name>`; secure keys via `secureJsonData[key]` |
| `packages/declarative-plugin/src/components/config-editor/rest/ServiceConfig.tsx:40-116` | Per-service layout: Connection section (if >1 server or variables), Auth section (if auth methods) |
| `packages/declarative-plugin/src/components/config-editor/rest/Connection.tsx:42-79` | Single-server + variables ⇒ **server selector hidden**, only the variable(s) rendered; ConfigSection title "Connection" |
| `packages/declarative-plugin/src/components/common/VariablesForm.tsx:22-61` | Variable inputs use `variable.name` as label; **every variable is marked `required` in the UI** regardless of the ref's `required` flag |
| `packages/declarative-plugin/src/components/config-editor/Auth.tsx:47-87` | Basic auth: `username` → `jsonData.services.<id>.auth.username`; password → secure key `<id>.password` |
| `packages/declarative-plugin/package.json` | Framework version (`0.0.2`) |

## Field inventory

| Schema `id` | Storage key (relative to target) | Target | Editor label | Read by backend? |
| --- | --- | --- | --- | --- |
| `jsonData_variables_subdomain` | `variables.subdomain` | `jsonData` | Subdomain | Yes — substituted into the server URL |
| `jsonData_services_zendesk_auth_id` | `services.zendesk.auth.id` | `jsonData` | (auth method selector) | Yes — discriminator; backend defaults to `basic_auth` |
| `jsonData_services_zendesk_auth_username` | `services.zendesk.auth.username` | `jsonData` | Email | Yes — basic-auth username |
| `secureJsonData_zendesk_password` | `zendesk.password` | `secureJsonData` | API Token | Yes — basic-auth password |

### Field provenance

| Schema `id` | Label source | Placeholder / default / options source | Value type source |
| --- | --- | --- | --- |
| `jsonData_variables_subdomain` | `spec.ts:12` (`variables.subdomain.name`) | description `spec.ts:13`; required from editor (`VariablesForm.tsx:37`) | `map[string]string` value, `config.go:5` |
| `jsonData_services_zendesk_auth_id` | — (framework auth selector) | value `basic_auth` = `$defs.authMethods` key `spec.ts:17`; default from backend `pluginclient.go:55-57` | `ServiceConfig.Auth.Id string`, `config.go:15` |
| `jsonData_services_zendesk_auth_username` | `spec.ts:21` (`user.label` "Email") | placeholder `spec.ts:22`; description `spec.ts:23` | `ServiceConfig.Auth.UserName string`, `config.go:16` |
| `secureJsonData_zendesk_password` | `spec.ts:26` (`password.label` "API Token") | placeholder `spec.ts:27`; description `spec.ts:28-30` | secret read at `pluginclient.go:98` (`<serviceId>.password`) |

### Frontend-only settings

None. Every field this entry models is read by the backend.

### Backend-only / not-modeled settings

- **`jsonData.services.zendesk.server.id`** — part of the `ServiceConfig` storage model (`config.go:11-13`)
  and read by the backend, but **not rendered** for a single-server service (`Connection.tsx:42-50`); the
  backend defaults it to the sole server (`pluginclient.go:52-54`). Omitted from the schema because it is
  never user-configurable for this plugin. (Multi-server plugins will model it as a selector.)
- **`jsonData.services.zendesk.disabled`** — exists in `ServiceConfig` (`config.go:10`) and gates client
  creation (`pluginclient.go:112`), but the config editor never renders a disable toggle. Not meaningful for a
  single-service plugin; omitted.
- **`jsonData.enableSecureSocksProxy`** — the Secure Socks Proxy (PDC) field (`ServiceConfig.tsx:107-113`),
  deliberately excluded from all registry entries per repo policy.

## Where the types are defined

Config types come almost entirely from the **shared framework**, not the plugin. Only config
type/field definitions are listed (UI components and functions/helpers are omitted).

### Frontend (TypeScript)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `Spec`, `ServiceDefRest`, `ServerDef`, `AuthMethodDef*`, `VariableDef` | `packages/declarative-plugin/src/types/spec.ts` | `@grafana/declarative-plugin` `0.0.2` |
| `Config` (`{ services, variables, enableSecureSocksProxy }`), `SecureConfig`, `ServiceConfig`, `AuthConfig`, `VariablesConfig` | `packages/declarative-plugin/src/types/datasource.ts` | `@grafana/declarative-plugin` `0.0.2` |
| The concrete service/server/auth/variable **data** (ids, labels, URL, method) | `plugins/grafana-zendesk-datasource/src/spec.ts` | plugin |

### Backend (Go)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `Spec`, `Service`, `Server`, `Variable`, `AuthMethod`, `AuthType` | `sdk/pluginspec/pluginspec.go` | `github.com/grafana/plugins/sdk/pluginspec` |
| `JsonData` (`services`, `variables`, `enableSecureSocksProxy`), `ServiceConfig` (`disabled`, `server.id`, `auth.*`) | `sdk/pluginspec/pluginclient/config.go` | `github.com/grafana/plugins/sdk/pluginspec/pluginclient` |
| `backend.DataSourceInstanceSettings` (carries `JSONData`, `DecryptedSecureJSONData`) | `backend/common.go` | `github.com/grafana/grafana-plugin-sdk-go` |

The `Config` in [`settings.go`](settings.go) is a **spec-specific projection** of the framework's generic
map-based `JsonData` (see [Modeling decisions](#modeling-decisions)); `settings.ts` keeps the three
canonical TypeScript types.

## Modeling decisions

- **Service-keyed storage → nested `section` fields.** These plugins store config under
  `jsonData.services.<serviceId>.…` and `jsonData.variables.<name>`, with **flat dotted secureJsonData keys**
  (`<serviceId>.password`). This is modeled with dotted `section` values (`services.zendesk.auth`,
  `variables`) and full dotted secure `key`s (`zendesk.password`). The dsconfig converter resolves dotted
  sections recursively into nested OpenAPI objects (`dsconfig/convert.go` `placeInSectionPath`).
- **Concrete Go structs instead of the framework's maps.** The framework's runtime model is map-based
  (`Services map[string]ServiceConfig`, `Variables map[string]string`). The dsconfig conformance guard
  (`JSONDataMatchesStruct`) walks the settings struct and only recurses `section` paths through **struct**
  fields, so a map would fail parity. `Config` therefore uses concrete, spec-specific nested structs
  (`Services.Zendesk.Auth.{Id,UserName}`, `Variables.Subdomain`). The wire JSON is identical to what the
  framework parses — this is a faithful, plugin-specific view of the same bytes.
- **Auth discriminator with a single value.** `services.zendesk.auth.id` is the auth-method discriminator
  (`role: auth.discriminator`) with the single allowed value `basic_auth` and a default matching the backend
  (which defaults `auth.id` to the server's first method). This establishes the multi-auth pattern other
  plugins in this monorepo will reuse.
- **`subdomain` marked required.** The variable is essential (it builds the API URL) and the editor renders it
  required, so the schema marks it `required: true`. The backend `validateVariables` does **not** enforce it
  (the ref lacks `required: true`) — recorded under [Upstream findings](#upstream-findings).
- **`requiredWhen`/`dependsOn` vs the editor.** `dependsOn`/`requiredWhen` on the username and password point
  at `jsonData_services_zendesk_auth_id == 'basic_auth'`, mirroring both editor visibility and the runtime
  contract. Because Zendesk exposes only basic auth, these are always satisfied today; they exist so the
  pattern generalizes.
- **Field ID naming.** IDs are prefixed with the storage target and mirror the dotted path with underscores:
  `jsonData_services_zendesk_auth_username`, `secureJsonData_zendesk_password`. `key` keeps the raw storage
  leaf/secret key.

## `LoadConfig` utility (`settings.go`)

`LoadConfig(ctx, settings)` runs the full three-phase flow and returns a fully-defaulted, validated `Config`:

1. **Parse** — unmarshal `settings.JSONData` into `Config` and copy the decrypted secret the plugin reads
   (`zendesk.password`). Mirrors the parse the shared backend performs in `pluginclient.New`.
2. **`ApplyDefaults`** — fill the zero-valued discriminator: `services.zendesk.auth.id` → `basic_auth`
   (matching the backend's first-auth-method default).
3. **`Validate`** — enforce the **health-check contract**: a known auth method with its inputs
   (username + API token) and a subdomain for the URL.

`pluginclient.New` is itself lenient (it builds clients without hard-failing on missing credentials — auth is
enforced per request in `applyAuth`). `Validate` encodes the stricter contract a working datasource needs, so
the default example (empty placeholders) is expected to fail validation. `ApplyDefaults` and `Validate` are
exported for callers that assemble a `Config` directly.

## Settings examples matrix (`schema.go`)

| Example | Auth | `jsonData` | `secureJsonData` |
| --- | --- | --- | --- |
| `""` (default) | `basic_auth` (defaulted) | `services.zendesk.auth.id`, empty `username`/`subdomain` | `zendesk.password` (empty) |
| `basicAuth` | `basic_auth` | populated `username` + `subdomain` | `zendesk.password` (placeholder) |

## Upstream findings

Discovered while researching upstream; recorded here (not "fixed" in the schema — the schema records what the
plugin does).

1. **`subdomain` is required in the editor but not enforced by the backend.** `VariablesForm.tsx:37` marks
   every variable `required` in the UI, but `serviceclient.go:301-327` only enforces variables whose ref has
   `required: true`. The `subdomain` ref (`spec.ts:38`) omits it, so a missing subdomain is not caught at
   config/health time — it surfaces later as a broken request URL (`https://.zendesk.com/api/v2/…`).
2. **Credentials are not validated at instance creation.** `pluginclient.New` builds service clients without
   checking auth; missing username/password only fail at request time via `applyAuth`
   (`serviceclient.go:244-260`). A datasource can be saved in a non-working state and only reports the problem
   on a health check or query.
3. **Single-server services hide server selection.** `Connection.tsx:42-50` renders no server URL control when
   there is one server, and the backend defaults `services.<id>.server.id` to it (`pluginclient.go:52-54`); the
   field can still be set via provisioning/API.
4. **`plugin.json` metadata is generic.** `info.description` is empty and the "Repository" link points at
   `github.com/grafana/grafana` rather than the actual home (`github.com/grafana/plugins`); the docs link is
   used here for `docURL`.

## Validation performed

- `dsconfig.ParseAndResolveSchemaJSON` + `Schema.Validate()` (via the `ConfigSchemaValid` conformance subtest) — passes.
- JSON Schema validation against [`dsconfig/schema.json`](../../dsconfig/schema.json) with `ajv --spec=draft7 --strict=false -c ajv-formats` (as CI runs) — `valid`.
- `go build ./...`, `go vet ./...`, `gofmt -l .`, `go test -race ./...` inside `registry/` — clean; the full conformance suite (schema round-trip, artifact-in-sync, spec/secure separation, jsonData⇔struct key + type parity, secure-key parity) passes.
- `settings.ts`: `tsc --noEmit --strict` — clean.
- Pre-existing `dsconfig` and `schema` workspace modules still build; all other registry entries still pass `go test`.
