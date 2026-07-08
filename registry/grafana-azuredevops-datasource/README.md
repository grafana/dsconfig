# grafana-azuredevops-datasource

Declarative configuration schema for the Azure DevOps datasource plugin (`grafana-azuredevops-datasource`).

## Upstream researched

- **Monorepo**: `github.com/grafana/plugins-private`
- **Commit SHA**: `267f4937806ed6404b6628d13ae358a5d308e376`
- **Plugin path**: `plugins/grafana-azuredevops-datasource/`

Every value in [`dsconfig.json`](dsconfig.json) — labels, placeholders, tooltips (as field
`description`s), option values, section titles, defaults, validations, required markers, storage
keys, storage targets, value types, group titles, and instructions — is traceable to a specific
`file:line` in the upstream plugin at this SHA. See [Field provenance](#field-provenance).

To reproduce this research, read the sources under
`plugins/grafana-azuredevops-datasource/` at commit
`267f4937806ed6404b6628d13ae358a5d308e376`. `@grafana/*` dependencies use the `catalog:` protocol;
versions were resolved from the monorepo `.yarnrc.yml` catalog block.

## Files

| File | Purpose |
| --- | --- |
| [`dsconfig.json`](dsconfig.json) | dsconfig v1 schema — single source of truth for all config fields, groups, relationships, and instructions |
| [`settings.ts`](settings.ts) | TypeScript models: `RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig` |
| [`settings.go`](settings.go) | Go `Config` model (flat: jsonData fields + `DecryptedSecureJSONData`), `PluginID`, `AuthType` / `SecureJsonDataKey` typed constants, and the `LoadConfig` utility |
| [`schema.go`](schema.go) | k8s-style SDK plugin schema: embeds `dsconfig.json`, converts it via `dsconfig.NewSDKSchema`, and defines `SettingsExamples` for each connection variant |
| [`settings_test.go`](settings_test.go) | Table tests for `LoadConfig`, `ApplyDefaults`, and `Validate` |
| [`conformance_test.go`](conformance_test.go) | Runs `schema.RunPluginTests` — the shared dsconfig conformance suite — against `dsconfig.json` |
| [`schema.gen.json`](schema.gen.json), [`settings.gen.json`](settings.gen.json), [`settings.examples.gen.json`](settings.examples.gen.json) | Committed schema artifacts (regenerate with `go generate ./...`; the `SchemaArtifactInSync` conformance subtest guards drift) |

There is no per-entry `go.mod` — every registry entry is a subpackage of the shared
[`registry/`](..) module (`github.com/grafana/dsconfig/registry`). This entry's package is
`azuredevopsdatasource`.

## Sources researched

Every source below was read at the pinned monorepo SHA
(`267f4937806ed6404b6628d13ae358a5d308e376`).

### Plugin sources (`plugins/grafana-azuredevops-datasource/`)

| File | What was read |
| --- | --- |
| `src/plugin.json:3-4,33` | `pluginName` (`name` = "Azure DevOps"), `pluginType` (`id`), `docURL` (`info.links[]` Docs url) |
| `src/types.ts:4-11` | `AzDoConfig` (jsonData: `url`, `authType`, `projectsLimit`, `enableSecureSocksProxy`, `username`) and `AzDoSecureConfig` (secureJsonData: `patToken`) |
| `src/editors/AzDoConfigEditor.tsx:19-185` | The config editor: URL input (`:56-69`), PAT password/Reset input (`:70-101`), `authType:'patToken'` stamped on every jsonData write (`:35-40`), Projects limit input (`:106-123`), Username input (`:124-140`), Secure Socks Proxy switch (`:141-180`, excluded) |
| `src/selectors.ts:5-40` | `Components.ConfigEditor.AzDoSettings`: group titles + every field's label / placeholder / ariaLabel / tooltip |
| `pkg/plugin/settings.go:10-51` | `AzDoConfig` backend struct (json tags), `Validate` (url + patToken required), `GetSettings` (unmarshal → url check → copy `patToken` → patToken check → `projectsLimit < 1 → 100` → `Validate`) |
| `pkg/plugin/constants.go:6,11-12` | `PluginID`, `ErrorInvalidURL` ("invalid URL"), `ErrorInvalidPATToken` ("invalid PAT") |
| `pkg/plugin/plugin.go:63-138` | `GetInstance` consumption: `normalizeURL` (`:63-65`), `azuredevops.NewPatConnection(url, patToken)` (`:74`), `username != ""` → `CreateBasicAuthHeaderValue(username, patToken)` + normalized `BaseUrl` + `SuppressFedAuthRedirect` (`:76-84`), `s.HTTPClientOptions(ctx)` proxy wiring (`:86-98`), `ProjectsLimit` handed to the datasource (`:127`) |

### External components / libraries

Resolved from the plugin's `package.json` → monorepo `.yarnrc.yml` catalog. The config editor
composes only generic primitives; **all human-readable text is plugin-defined in
`src/selectors.ts`**, not library-driven.

| Component / type | Version | What was read |
| --- | --- | --- |
| `Button`, `Collapse`, `InlineFormLabel`, `Input`, `Switch` | `@grafana/ui@^11.6.7` | Generic form primitives (prop names `value`, `placeholder`, `tooltip`, `onChange`, `onBlur`, `type="password"`) — confirmed no library-supplied labels/storage keys |
| `DataSourceJsonData`, `DataSourcePluginOptionsEditorProps`, `DataSourceSettings`, `FeatureToggles` | `@grafana/data@^11.6.7` | Base `jsonData` interface `AzDoConfig` extends; editor prop shapes; `secureSocksDSProxyEnabled` toggle name |
| `config` | `@grafana/runtime@^11.6.7` | `config.featureToggles` / `config.buildInfo.version` — the Secure Socks Proxy render gate (`:31-34`) |
| `E2ESelectors` | `@grafana/e2e-selectors` (intentionally **not** cataloged — swapped per Grafana version) | Type wrapper for the `selectors.ts` map; the label/placeholder/tooltip strings themselves are plugin-defined |
| `css` | `@emotion/css@11.10.6` | Secure Socks Proxy toggle styling only |
| `backend.DataSourceInstanceSettings`, `httpclient.Options` / `ProxyOptions` | `github.com/grafana/grafana-plugin-sdk-go@v0.279.0` | `JSONData`, `DecryptedSecureJSONData`, `HTTPClientOptions(ctx)` |
| `azuredevops.Connection`, `NewPatConnection`, `CreateBasicAuthHeaderValue` | `github.com/microsoft/azure-devops-go-api/azuredevops/v7@v7.1.0` | How `url` + `patToken` (+ optional `username`) become the transport auth (Basic auth) |

## Field provenance

| Schema `id` | Storage key | Target | Label / description source | Placeholder / default source | Value type source |
| --- | --- | --- | --- | --- | --- |
| `jsonData_url` | `url` | `jsonData` | label `selectors.ts:9` ("URL"); description = tooltip `selectors.ts:12` ("Azure DevOps instance URL") | placeholder `selectors.ts:10` ("https://dev.azure.com/XXXX"); `required` from backend `settings.go:20-22,35-38` | `AzDoConfig.URL string` `settings.go:12`; TS `string` `types.ts:5` |
| `jsonData_authType` | `authType` | `jsonData` | — (no UI; stamped programmatically `AzDoConfigEditor.tsx:38`) | `defaultValue "patToken"` from the editor stamp + TS literal `types.ts:6`; `allowedValues:["patToken"]` | `AzDoConfig.AuthType string` `settings.go:11`; TS `'patToken'` `types.ts:6` |
| `secureJsonData_patToken` | `patToken` | `secureJsonData` | label `selectors.ts:15` ("PAT"); description = tooltip `selectors.ts:18` ("Azure DevOps personal access token") | placeholder `selectors.ts:16` ("Azure DevOps PAT"); `requiredWhen:"true"` from backend `settings.go:23-25,42-44` | `AzDoSecureConfig.patToken string` `types.ts:11`; backend `PATToken string` (from secure) `settings.go:14,39-41` |
| `jsonData_projectsLimit` | `projectsLimit` | `jsonData` | label `selectors.ts:26` ("Projects limit"); description = tooltip `selectors.ts:29` | placeholder `selectors.ts:27` ("100"); `defaultValue 100` from editor `AzDoConfigEditor.tsx:28` + backend `settings.go:46-48` | `AzDoConfig.ProjectsLimit int` `settings.go:13`; TS `number` `types.ts:7` |
| `jsonData_username` | `username` | `jsonData` | label `selectors.ts:32` ("Username"); description = tooltip `selectors.ts:35-36` | placeholder `selectors.ts:33` ("ado") | `AzDoConfig.Username string` `settings.go:16`; TS `string` `types.ts:9` |

## Field inventory summary

| Schema field | Storage key | Target | Editor label | Read by backend? |
| --- | --- | --- | --- | --- |
| `jsonData_url` | `url` | `jsonData` | URL | **Yes** (required; base URL for the connection) |
| `jsonData_authType` | `authType` | `jsonData` | — (stamped, no UI) | **Parsed but unused** — declared `// Not in use yet` (`settings.go:11`) |
| `secureJsonData_patToken` | `patToken` | `secureJsonData` | PAT | **Yes** (required; HTTP Basic password) |
| `jsonData_projectsLimit` | `projectsLimit` | `jsonData` | Projects limit | **Yes** (default 100; caps projects-list query) |
| `jsonData_username` | `username` | `jsonData` | Username | **Yes** (switches to explicit Basic auth) |
| _`enableSecureSocksProxy`_ | `enableSecureSocksProxy` | `jsonData` | Secure Socks Proxy | SDK-transparent — **excluded** from this entry (AGENTS.md) |

### Frontend-only settings

None. Every schema field is either read by the backend or (for `authType`) parsed into the backend
struct.

### Backend-only / no-UI settings

- **`authType`** has no rendered UI control. The editor writes it as a side effect of every
  `onOptionChange` (`AzDoConfigEditor.tsx:38`), always as `'patToken'`. The backend struct declares
  it (`AzDoConfig.AuthType`, `settings.go:11`) but marks it `// Not in use yet` and never branches
  on it — so it is modeled as a fixed single-value `auth.discriminator` and tagged
  `backend-declared-unused`. See [Upstream findings](#upstream-findings) #1.

### Excluded settings

- **`enableSecureSocksProxy`** (`AzDoConfigEditor.tsx:141-180`, gated on
  `config.featureToggles.secureSocksDSProxyEnabled` and Grafana ≥ 10.0.0) is written to `jsonData`
  and consumed transparently by the SDK via `s.HTTPClientOptions(ctx)` (`plugin.go:86`). The backend
  struct carries it as `ProxyEnabled` (`settings.go:15`) but never inspects it by name. Deliberately
  excluded per AGENTS.md; `json.Unmarshal` silently ignores it and the Go `Config` omits it.

## Where the types are defined

Only config type/field definitions are listed (UI components and functions/helpers such as
`GetSettings`, `normalizeURL`, and `AzDoConfigEditor` are omitted even where they are the reason a
field exists).

### Frontend (TypeScript)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `AzDoConfig` (jsonData: `url`, `authType`, `projectsLimit`, `enableSecureSocksProxy`, `username`), `AzDoSecureConfig` (`patToken`) | `src/types.ts:4-11` | plugin |
| `DataSourceJsonData` (base interface `AzDoConfig` extends) | `packages/grafana-data/src/types/datasource.ts` | `@grafana/data` `^11.6.7` |

There is **no frontend enum** for `authType`; it is the inline string literal `'patToken'`
(`types.ts:6`). The `AzDoAuthType` alias in `settings.ts` is derived from that literal.

### Backend (Go)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `AzDoConfig` (`authType`, `url`, `projectsLimit`, `enableSecureSocksProxy`/`ProxyEnabled`, `username`; `PATToken` from secure) | `pkg/plugin/settings.go:10-17` | plugin |
| `backend.DataSourceInstanceSettings` (carries `JSONData`, `DecryptedSecureJSONData`, and root fields like `URL`/`BasicAuth` — unused by this plugin) | `backend/common.go` | `grafana-plugin-sdk-go` `v0.279.0` |
| `httpclient.Options` / `ProxyOptions` | `backend/httpclient` | `grafana-plugin-sdk-go` `v0.279.0` |
| `azuredevops.Connection` (`AuthorizationString`, `BaseUrl`, `SuppressFedAuthRedirect`) — the transport the config is fed into | `azuredevops/connection.go` | `microsoft/azure-devops-go-api/v7` `v7.1.0` |

There is **no backend enum** for `authType` either — it is a plain `string` field. The `AuthType`
type + `AuthTypePAT` constant in `settings.go` are introduced by this entry for clarity.

## Modeling decisions

- **Single (degenerate) auth discriminator**: the plugin supports only PAT auth. `jsonData.authType`
  is modeled as an `auth.discriminator` with `allowedValues: ["patToken"]` and `defaultValue:
  "patToken"` because the field genuinely exists in storage and the backend struct, even though it
  currently has one value and the backend does not branch on it. The required PAT therefore uses a
  static `requiredWhen: "true"` (always required) rather than a per-method condition.
- **`url` lives in `jsonData`, not the datasource root**: the editor writes `jsonData.url`
  (`AzDoConfigEditor.tsx:66`) and the backend reads `AzDoConfig.URL` tagged `json:"url"`
  (`settings.go:12`). Unlike the GitLab datasource (root url), `RootConfig` here is a blank object
  and `url` is a normal `jsonData` field. The backend reads no root datasource fields at all.
- **`patToken` role = `auth.basic.password`**: Azure DevOps PAT auth is HTTP Basic. The backend
  builds the header via `azuredevops.CreateBasicAuthHeaderValue(username, patToken)` when a username
  is set (`plugin.go:77`) and via `azuredevops.NewPatConnection(url, patToken)` (empty-username Basic
  auth) otherwise (`plugin.go:74`). `username` is correspondingly `auth.basic.username`.
- **Groups mirror the editor's two `Collapse` sections**: "Azure DevOps settings" (URL + PAT,
  `selectors.ts:7`) then "Optional Configuration" (Projects limit + Username, `selectors.ts:24`,
  marked `optional: true`). `authType` is stamped programmatically and rendered nowhere, so — like
  GitHub's `githubPlan`/`cachingEnabled` — it is intentionally left out of every group.
- **ariaLabels are not modeled**: `selectors.ts` also defines per-field `ariaLabel`s, but the
  dsconfig `ui` vocabulary has no `ariaLabel` slot, so they are recorded here for provenance only and
  not carried into `dsconfig.json`.
- **Flat `Config` in Go**: `settings.go` mirrors the jsonData portion of the upstream `AzDoConfig`
  struct verbatim (`url`, `authType`, `projectsLimit`, `username`) plus `DecryptedSecureJSONData` for
  the `patToken` secret. `enableSecureSocksProxy`/`ProxyEnabled` is omitted (exclusion), and no
  root-level datasource fields are carried because the backend reads none.

## SDK plugin schema and k8s-style examples (`schema.go`)

`NewSchema()` assembles the `grafana-plugin-sdk-go` `pluginschema.PluginSchema` bundle from the
embedded `dsconfig.json`: the `jsonData` object (`url` required; `authType`, `projectsLimit`,
`username`) becomes the OpenAPI settings `spec`, `patToken` becomes the single `secureValues` entry,
and secure data never appears in the spec.

`SettingsExamples()` provides the default configuration plus one example per connection variant.
Each example is a full instance-settings object with the plugin configuration under `jsonData` and
the write-only PAT under `secureJsonData` (an obviously-fake `<your-azure-devops-pat>` placeholder;
the default example — keyed `""` — carries an empty `patToken` to show what must be filled in):

| Example | Auth | Connection | `secureJsonData` |
| --- | --- | --- | --- |
| `""` (default) | PAT (schema defaults) | — (url must be filled in) | `patToken` (empty) |
| `patToken` | PAT | Azure DevOps Services (`https://dev.azure.com/<org>`) | `patToken` |
| `azureDevOpsServer` | PAT + Basic username | Azure DevOps Server (`https://<server>/<collection>`) | `patToken` |

## `LoadConfig` utility (`settings.go`)

`LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error)` runs
the full three-phase load flow and returns a fully-defaulted, validated `Config`:

1. **Parse** — unmarshal `settings.JSONData` into `Config` (mirroring the upstream unconditional
   `json.Unmarshal`, `settings.go:31`, so empty bytes are a parse error) and copy the decrypted
   `patToken` into `DecryptedSecureJSONData`.
2. **`ApplyDefaults`** — fill the curated zero-valued fields the editor/backend default:
   `AuthType → "patToken"` (editor stamp) and `ProjectsLimit → 100` when `< 1`
   (`settings.go:46-48`).
3. **`Validate`** — enforce the runtime contract: non-empty `url` (`ErrorInvalidURL`) and non-empty
   `patToken` (`ErrorInvalidPATToken`). Errors are joined so every problem surfaces at once (upstream
   returns only the first).

Everything is logged via `backend.Logger.FromContext(ctx)` with `datasource_uid`, `datasource_name`,
and `plugin` labels. `ApplyDefaults` and `Validate` are exported separately so callers assembling a
`Config` directly (provisioning preview, round-trip tests) can invoke each phase individually.

## Upstream findings

Potential bugs, misleading UX, and consistency issues discovered while researching upstream. The
schema records what the plugin **does**, not what it **should** do.

1. **`authType` is dead weight.** `pkg/plugin/settings.go:11` declares `AuthType string
   json:"authType" // Not in use yet` and nothing in `pkg/` ever reads it. The editor still stamps
   `'patToken'` on every jsonData write (`AzDoConfigEditor.tsx:38`). Modeled as a fixed single-value
   discriminator.
2. **No required markers in the editor, hard-fail in the backend.** The editor renders plain
   `InlineFormLabel`s with no required indicator, yet `GetSettings` hard-fails on an empty `url`
   ("invalid URL", `settings.go:35-38`) or empty `patToken` ("invalid PAT", `settings.go:42-44`).
   The schema encodes the backend contract (`required: true` on url, `requiredWhen: "true"` on
   patToken).
3. **`projectsLimit` cannot be set below 1.** `settings.go:46-48` coerces any value `< 1` (including
   `0` and negatives) to `100`, so there is no way to request fewer than 1 — and the editor's number
   input allows entering `0` or negative values that are silently overridden.
4. **Setting `username` changes both auth *and* URL handling.** With a non-empty `username`, the
   backend abandons `NewPatConnection` for an explicit `CreateBasicAuthHeaderValue(username,
   patToken)` header and also lowercases + trims the trailing slash of the URL via `normalizeURL`
   (`plugin.go:76-84`); the no-username path does neither. The URL is thus normalized only when a
   username is present.
5. **`AzDoConfig.Validate()` is partly redundant.** `GetSettings` checks `url` and `patToken`
   inline (`settings.go:35-45`) and then calls `config.Validate()` (`:50`), which checks the same two
   fields again (`settings.go:19-27`). The second check can never fire in the `GetSettings` path.

## Validation performed

- `dsconfig.ParseAndResolveSchemaJSON` + `Schema.Validate()` (Go validator in this repo) — passes.
- Strict JSON Schema validation against [`dsconfig/schema.json`](../../dsconfig/schema.json)
  (draft-07, `additionalProperties: false`) — passes.
- `go generate ./...` regenerates `schema.gen.json` / `settings.gen.json` /
  `settings.examples.gen.json`; `go build ./...`, `go vet ./...`, `gofmt -l .`, and `go test ./...`
  across the `registry` module — all clean (65 packages ok, incl. the conformance suite:
  schema round-trip, artifact sync, spec/secure separation, jsonData↔struct parity, secure-key
  parity, and `LoadConfig`/`ApplyDefaults`/`Validate` table tests).
- `tsc --noEmit --strict` on `settings.ts` (`typescript@5.5.4`) — clean.
- The pre-existing `dsconfig` and `schema` workspace modules still build and test — passes.
