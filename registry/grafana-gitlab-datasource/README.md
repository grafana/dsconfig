# grafana-gitlab-datasource

Declarative configuration schema for the GitLab datasource plugin (`grafana-gitlab-datasource`).

## Upstream researched

- **Repo**: `github.com/grafana/plugins-private` (monorepo)
- **Ref**: `main`
- **Commit SHA**: `267f4937806ed6404b6628d13ae358a5d308e376`
- **Plugin path**: `plugins/grafana-gitlab-datasource`
- **Plugin version**: `2.4.4` (`package.json:3`)

Every value in [`dsconfig.json`](dsconfig.json) — labels, placeholders, tooltips, section titles,
defaults, dependency/required-when expressions, storage keys, storage targets, value types, group
titles, and instructions — is traceable to a specific `file:line` in the upstream monorepo at this
SHA. See [Field provenance](#field-provenance) below.

To reproduce this research (the monorepo was read on disk, not cloned):

```bash
git -C <plugins-private> rev-parse HEAD   # 267f4937806ed6404b6628d13ae358a5d308e376
# plugin sources under plugins/grafana-gitlab-datasource/{src,pkg}
```

If upstream `main` has advanced past this SHA, re-diff the sources listed under [Sources
researched](#sources-researched) before merging any changes to this entry.

## Files

| File | Purpose |
| --- | --- |
| [`dsconfig.json`](dsconfig.json) | dsconfig v1 schema — single source of truth for all config fields, groups, relationships, and instructions |
| [`settings.ts`](settings.ts) | TypeScript models: `RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig` |
| [`settings.go`](settings.go) | Go `Config` model (flat: root `URL` + jsonData `PageLimit` + `DecryptedSecureJSONData`), `PluginID`, `SecureJsonDataKey` typed constants, and the `LoadConfig` utility |
| [`schema.go`](schema.go) | k8s-style SDK plugin schema: embeds `dsconfig.json`, converts it via `dsconfig.NewSDKSchema`, and defines `SettingsExamples` for the default + each connection variant |
| [`settings_test.go`](settings_test.go) | Table tests for `LoadConfig`, `ApplyDefaults`, `Validate`, and the `SettingsExamples` shape |
| [`conformance_test.go`](conformance_test.go) | Runs `schema.RunPluginTests` — the shared dsconfig conformance suite — against `dsconfig.json` |
| [`schema.gen.json`](schema.gen.json), [`settings.gen.json`](settings.gen.json), [`settings.examples.gen.json`](settings.examples.gen.json) | Committed schema artifacts (regenerate with `go generate ./...` inside this directory; the `SchemaArtifactInSync` conformance subtest guards drift) |

There is no per-entry `go.mod` — every registry entry is a subpackage of the shared
[`registry/`](..) module (`github.com/grafana/dsconfig/registry`).

## Sources researched

Every source below was read at the pinned monorepo SHA
(`267f4937806ed6404b6628d13ae358a5d308e376`), plus external editor components at the exact versions
the plugin's `package.json` pins (resolved through the monorepo `.yarnrc.yml` catalog) and the
backend `go.mod` dependencies.

### Plugin (`plugins/grafana-gitlab-datasource`)

| File | What was read |
| --- | --- |
| `src/plugin.json:4,5,24` | `pluginName` (`name` = `"GitLab"`), `pluginType` (`id` = `"grafana-gitlab-datasource"`), `docURL` (`info.links[0].url`) |
| `src/views/ConfigEditor.tsx:14-18` | `onURLChange` — writes the root `options.url` (`value \|\| DefaultURL`) |
| `src/views/ConfigEditor.tsx:21-29,41-54` | `onAccessTokenChange` / `onResetAccessToken` — how the secret is written/reset |
| `src/views/ConfigEditor.tsx:31-34` | `onPageLimitChange` — `jsonData.pageLimit = parseInt(value, 10)` |
| `src/views/ConfigEditor.tsx:36-39,184-211` | `onSecureSocksProxyChange` and the conditional Secure Socks Proxy block — deliberately excluded |
| `src/views/ConfigEditor.tsx:73-77` | `DataSourceDescription` (`dataSourceName="Gitlab"`, `docsLink`, `hasRequiredFields`) |
| `src/views/ConfigEditor.tsx:79-98` | `ConfigSection title="Connection"` + the URL `Input` (label `:81`, tooltip `:82`, `required` `:84`, placeholder `:93`) |
| `src/views/ConfigEditor.tsx:100-148` | `Auth visibleMethods={['custom-gitlab']}` with the single custom method "Gitlab authentication" (`:108`, description `:109`) and its Access token `SecretInput` (label `:114`, `required` `:115`, tooltip `:116-128` incl. link `:121`, placeholder `:137`) |
| `src/views/ConfigEditor.tsx:149-153,178-183` | The two inline "Please enter an access token." `FieldValidationMessage` blocks |
| `src/views/ConfigEditor.tsx:155-176` | `ConfigSection title="Additional Settings"` (description `:157`) → `ConfigSubSection title="Page limit"` → Page limit `Input` (label `:161`, tooltip `:163`, placeholder `:172`) |
| `src/types.ts:35-38` | `GitLabDataSourceOptions` (jsonData: `pageLimit`, `enableSecureSocksProxy`) |
| `src/types.ts:40-42` | `GitLabSecureJsonData` (`accessToken`) |
| `src/types.ts:63` | `DefaultURL = 'https://gitlab.com/api/v4'` |
| `src/components/selectors.ts:3-14` | E2E selector map (`Config Editor URL` / `Config Editor Access Token` / `Config Editor page limit`) — aria-labels only |
| `pkg/models/settings.go:16-21` | `Settings` struct (`URL` `json:"url"`, `AccessToken` `json:"accessToken"`, `PageLimit` `json:"pageLimit,omitempty"`, `SdkClientOptions` `json:"-"`) |
| `pkg/models/settings.go:23,25` | `baseGitLabURL = "https://gitlab.com/api/v4"`, `basePageLimit = 5` |
| `pkg/models/settings.go:28-63` | `LoadSettings`: unmarshal jsonData (`:30`), `settings.URL = config.URL` (`:34`) overwriting jsonData.url, default URL (`:35-37`), default pageLimit (`:39-41`), copy `accessToken` (`:47`), hard-fail on empty token (`:48-50`), build HTTP client options (`:52-61`) |
| `pkg/gitlab/datasource.go:166-186` | `NewDatasource`: `gitlab.WithBaseURL(settings.URL)` (`:167`), `gitlab.NewClient(settings.AccessToken, ...)` (`:176`) |
| `pkg/gitlab/datasource.go:104-130` | `CheckHealth` error mapping: `unsupported protocol scheme`/`invalid character` → `ErrorBadScheme` (`:110-111`), `404` → `ErrorBadURL` (`:112-113`), `401`/cert → `ErrorBadAccessToken` (`:114-115`) |
| `pkg/gitlab/datasource.go:30-97` | `PageLimit` consumed by every list handler (`d.settings.PageLimit`, e.g. `:31`) |
| `pkg/errors/errors.go:19,22,25,28` | `ErrorBadURL`, `ErrorBadAccessToken`, `ErrorBadScheme`, `ErrorEmptyAccessToken` ("access token can not be blank") |
| `pkg/plugin/instance.go:25-51` | `NewDataSourceInstance` → `LoadSettings` → `gitlab.NewDatasource` wiring |
| `package.json` | External component + dependency versions (see next tables) |

### External editor components

Read at the exact versions pinned in the plugin's `package.json` (`catalog:` → monorepo
`.yarnrc.yml`). The `@grafana/plugin-ui` sources were read from the published package (`0.13.1`).

| Component | Version | What was read |
| --- | --- | --- |
| `Auth` | `@grafana/plugin-ui@^0.13.1` | Renders `ConfigSection title="Authentication"` wrapping `AuthMethodSettings` (`Auth.js`) |
| `AuthMethodSettings` | `@grafana/plugin-ui@^0.13.1` | With a single visible method it drops the method `Select` (`hasSelect=false`) and renders the custom component under a `ConfigSubSection` whose title/description are the method's `label`/`description` (`AuthMethodSettings.js`) — i.e. "Gitlab authentication" / "Provide information to grant access to the data source." |
| `ConfigSection`, `ConfigSubSection` | `@grafana/plugin-ui@^0.13.1` | `GenericConfigSection` defaults `isCollapsible=false`; renders `<h3>`/`<h6>` titles + optional description (`GenericConfigSection.js`) |
| `DataSourceDescription` | `@grafana/plugin-ui@^0.13.1` | Header block with the docs link and the "Required fields" note driven by `hasRequiredFields` |
| `Input`, `SecretInput`, `InlineField`, `InlineSwitch`, `FieldValidationMessage`, `useTheme2` | `@grafana/ui@^11.6.7` | Prop names (`label`, `tooltip`, `placeholder`, `value`, `onChange`, `isConfigured`, `onReset`, `required`) so we knew which attributes to record |
| `DataSourcePluginOptionsEditorProps`, `DataSourceJsonData`, `FeatureToggles` | `@grafana/data@^11.6.7` | Base interface `GitLabDataSourceOptions` extends; editor props; feature-toggle gate for the SSP block |
| `config` | `@grafana/runtime@^11.6.7` | `config.featureToggles.secureSocksDSProxyEnabled` + `config.buildInfo.version` gate for the SSP block |

### Backend dependencies (`go.mod`)

| Module | Version | Why it matters |
| --- | --- | --- |
| `github.com/grafana/grafana-plugin-sdk-go` | `v0.290.0` | `backend.DataSourceInstanceSettings` (root `URL`, `JSONData`, `DecryptedSecureJSONData`), `httpclient.Options` |
| `github.com/xanzy/go-gitlab` | `v0.105.0` | `NewClient` sets `authType = PrivateToken` (`gitlab.go:254-261`) and sends `PRIVATE-TOKEN: <token>` (`gitlab.go:855-858`); `WithBaseURL`/`setBaseURL` append `api/v4/` when missing (`gitlab.go:564-578`) |

## Field provenance

| Schema `id` | Storage key | Target | Label source | Placeholder / default source | Value type source | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `root_url` | `url` | `root` | `ConfigEditor.tsx:81` (`label="URL"`) | Placeholder `ConfigEditor.tsx:93` (`` `Default: ${DefaultURL}` ``); default `types.ts:63` / `settings.go:23` | `Settings.URL string` `pkg/models/settings.go:17`; TS root `DataSourceSettings.url` (`@grafana/data`) | Tooltip `ConfigEditor.tsx:82`. Root field: editor writes `options.url` (`:14-18`), backend reads `config.URL` (`:34`). Role `endpoint.baseUrl`. No `requiredWhen` — backend defaults empty → `baseGitLabURL` |
| `secureJsonData_accessToken` | `accessToken` | `secureJsonData` | `ConfigEditor.tsx:114` (`label="Access token"`) | Placeholder `ConfigEditor.tsx:137` (`"Access token"`) | `Settings.AccessToken string` `pkg/models/settings.go:18`; TS `GitLabSecureJsonData.accessToken` `types.ts:41` | Description = tooltip `ConfigEditor.tsx:116-128` (inline link `:121`). Role `auth.bearer.token`. `requiredWhen: "true"` from backend hard-fail `settings.go:48-50` |
| `jsonData_pageLimit` | `pageLimit` | `jsonData` | `ConfigEditor.tsx:161` (`label="Page limit"`) | Placeholder `ConfigEditor.tsx:172` (`"Page limit"`); default `5` `settings.go:25,39-41` | `Settings.PageLimit int` `pkg/models/settings.go:19`; TS `number` `types.ts:36` | Description = tooltip `ConfigEditor.tsx:163`. In collapsible-style "Additional Settings" section (`optional: true`) |

## Field inventory summary

| Schema field | Storage key | Target | Editor label | Read by backend? |
| --- | --- | --- | --- | --- |
| `root_url` | `url` | `root` | URL | Yes (`config.URL`, `settings.go:34`) |
| `secureJsonData_accessToken` | `accessToken` | `secureJsonData` | Access token | Yes (`settings.go:47`; PRIVATE-TOKEN header) |
| `jsonData_pageLimit` | `pageLimit` | `jsonData` | Page limit | Yes (`settings.PageLimit`, `datasource.go:31` etc.) |

### Frontend-only settings

None. All three modeled fields are read by the backend.

### Backend-only settings

None. Every backend `Settings` field is either editor-visible (`url`, `accessToken`, `pageLimit`)
or runtime-only transport state (`SdkClientOptions`, not configuration).

### Excluded settings

- **`jsonData.enableSecureSocksProxy`** — written by the Secure Socks Proxy switch
  (`ConfigEditor.tsx:36-39,184-211`, gated by `config.featureToggles.secureSocksDSProxyEnabled` +
  Grafana ≥ 10.0.0) and consumed transparently by the SDK's `config.HTTPClientOptions(ctx)` call
  (`settings.go:53`). Deliberately omitted per AGENTS.md; `Config`'s json unmarshal silently
  ignores it.

## Where the types are defined

Only config type/field definitions are listed — UI components (`Auth`, `ConfigSection`,
`DataSourceDescription`, `SecretInput`, …) and functions/helpers (`LoadSettings`, `NewDatasource`,
go-gitlab's `NewClient`) are omitted even where they are the reason a field exists.

### Frontend (TypeScript)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `GitLabDataSourceOptions` (jsonData: `pageLimit`, `enableSecureSocksProxy`), `GitLabSecureJsonData` (`accessToken`) | `src/types.ts:35-42` | plugin |
| `DataSourceJsonData` (base interface `GitLabDataSourceOptions` extends) | `packages/grafana-data/src/types/datasource.ts` | `@grafana/data` `^11.6.7` |
| `DataSourceSettings.url` (the root `url` field the editor writes via `options.url`) | `packages/grafana-data/src/types/datasource.ts` | `@grafana/data` `^11.6.7` |

### Backend (Go)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `Settings` (`URL`, `AccessToken`, `PageLimit`, `SdkClientOptions`) | `pkg/models/settings.go:16-21` | plugin |
| `backend.DataSourceInstanceSettings` (carries root `URL`, `JSONData`, `DecryptedSecureJSONData`) | `backend/common.go` | `grafana-plugin-sdk-go` `v0.290.0` |
| `httpclient.Options` (type of `Settings.SdkClientOptions`) | `backend/httpclient` | `grafana-plugin-sdk-go` `v0.290.0` |

`Settings.URL` is tagged `json:"url"` but is never read from jsonData — `LoadSettings` overwrites
it with `config.URL` (`settings.go:34`), so it is a root field. This entry models it as `root_url`
in the schema and as `Config.URL` tagged `json:"-"` (populated from `settings.URL`) in Go.

## Modeling decisions

- **URL is a root field, not jsonData.** The config editor writes `options.url`
  (`ConfigEditor.tsx:14-18`) and the backend reads `config.URL` (`settings.go:34`). Although the
  upstream `Settings.URL` field is tagged `json:"url"`, `LoadSettings` unmarshals jsonData and then
  immediately overwrites `URL` with `config.URL`, so any `jsonData.url` is dead. Modeled as
  `root_url` (target `root`) and `Config.URL` tagged `json:"-"`.
- **No `requiredWhen` on the URL.** The editor marks the URL required (`hasRequiredFields`
  `ConfigEditor.tsx:76` + `required` `:84`), but the backend defaults an empty URL to
  `https://gitlab.com/api/v4` (`settings.go:35-37`), so it is not part of the backend data
  contract. The URL carries a `defaultValue` instead, and this editor-vs-backend discrepancy is
  recorded here (mirrors how the Sentry entry treats its `url`).
- **`requiredWhen: "true"` on the access token.** The editor shows an inline "Please enter an
  access token." message (`ConfigEditor.tsx:149-153`) and the backend hard-fails without it
  (`ErrorEmptyAccessToken`, `settings.go:48-50`), so the token is a genuine backend requirement.
- **Role `auth.bearer.token` for the access token** is the closest match in the closed role
  vocabulary. Note go-gitlab actually transmits it as the `PRIVATE-TOKEN` request header
  (`gitlab.go:855-858`), not an `Authorization: Bearer` header — the role is a semantic
  approximation, documented here.
- **Groups mirror the editor sections**: `Connection` (the `ConfigSection`, `:79`) →
  `Authentication` (the section the `Auth` component renders, `:100`) → `Additional Settings`
  (`:155`). The "Additional Settings" group is marked `optional: true`: although `ConfigSection`
  defaults `isCollapsible=false`, that section holds only the optional page limit and its own
  description calls it "optional settings".
- **Secure Socks Proxy excluded** (AGENTS.md) — see [Excluded settings](#excluded-settings).
- **Field ID naming convention**: IDs are prefixed with their storage target (`root_`, `jsonData_`,
  `secureJsonData_`) followed by the camelCase storage key; the `key` property keeps the raw
  storage key.
- **Flat `Config` in Go**: `settings.go` collapses the root `URL`, the jsonData `PageLimit`, and
  the decrypted secret onto a single `Config`. `SecureJsonDataConfig` is the key list
  (`accessToken`) since secure values are write-only.

## SDK plugin schema and k8s-style examples (`schema.go`)

`NewSchema()` assembles the `grafana-plugin-sdk-go` `pluginschema.PluginSchema` bundle (served by
Grafana's datasource API server as `{apiVersion}.json`, `v0alpha1` today) from the embedded
`dsconfig.json`: the root `url` becomes a top-level settings `spec` property, `jsonData.pageLimit`
becomes a nested `jsonData` property, and `accessToken` becomes a `secureValues` entry (never part
of the spec).

`SettingsExamples()` provides the default configuration plus one example per connection variant.
All secret placeholders are obviously-fake angle-bracket tokens (`<gitlab-personal-access-token>`),
never a realistic `glpat-…` shape, so commit push-protection scanners do not flag them:

| Example | Connection | `url` | `jsonData` | `secureJsonData` |
| --- | --- | --- | --- | --- |
| `""` (default) | GitLab SaaS (schema defaults) | `https://gitlab.com/api/v4` | `pageLimit: 5` | `accessToken` (empty) |
| `gitlabSaaS` | GitLab SaaS (gitlab.com) | `https://gitlab.com/api/v4` | `pageLimit: 5` | `accessToken` |
| `selfHosted` | Self-hosted (scheme only) | `https://gitlab.example.com` | `pageLimit: 5` | `accessToken` |
| `selfHostedApiV4CustomPageLimit` | Self-hosted (explicit `/api/v4`) | `https://gitlab.example.com/api/v4` | `pageLimit: 10` | `accessToken` |

## `LoadConfig` utility (`settings.go`)

`LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error)`
runs the full three-phase load flow and returns a fully-defaulted, validated `Config`:

1. **Parse** — unmarshal `settings.JSONData` into `Config` (only `pageLimit`), take the URL from
   the datasource root (`cfg.URL = settings.URL`, mirroring `settings.go:34`), and copy the
   decrypted `accessToken` into `DecryptedSecureJSONData`. Unmarshal is unconditional, mirroring
   the upstream `LoadSettings` (`settings.go:30`): a truly-empty `JSONData` payload is a parse
   error (Grafana always sends at least `"{}"`).
2. **`ApplyDefaults`** — default an empty `URL` to `https://gitlab.com/api/v4` and a `0` page limit
   to `5` (mirrors `settings.go:35-41`).
3. **`Validate`** — enforce the one backend hard requirement: a non-empty access token
   (`settings.go:48-50`). Errors are joined so every problem surfaces at once.

Everything is logged via `backend.Logger.FromContext(ctx)` with `datasource_uid`, `datasource_name`,
and `plugin` labels. `(*Config).ApplyDefaults()` and `(Config).Validate()` stay exported for callers
that assemble a `Config` directly (provisioning preview, schema-example round-trip, tests that need
to distinguish parse-level from policy-level errors).

## Upstream findings

Potential bugs, misleading UX, and consistency issues discovered while researching upstream. All
are preserved verbatim in the schema — the schema records what the plugin **does**, not what it
**should** do.

1. **The URL lives at the datasource root, and `Settings.URL`'s `json:"url"` tag is dead.**
   `LoadSettings` unmarshals `config.JSONData` into `settings` (`settings.go:30`) and then
   unconditionally overwrites `settings.URL = config.URL` (`:34`), so any `jsonData.url` is
   discarded. Consumers must set the datasource root `url`, never `jsonData.url`.
2. **The URL is editor-required but backend-optional.** The editor renders `hasRequiredFields`
   (`ConfigEditor.tsx:76`) and marks the URL `required` (`:84`), yet the backend defaults an empty
   URL to `https://gitlab.com/api/v4` (`settings.go:35-37`) and never rejects it.
3. **The URL tooltip is misleading.** `ConfigEditor.tsx:82` suggests `gitlab.domain.com` — a bare
   host with no scheme and no `/api/v4`. But go-gitlab requires an `http`/`https` scheme (a
   scheme-less host makes the health check fail with `ErrorBadScheme`, "unsupported scheme. Only
   HTTP and HTTPS are supported", via `datasource.go:110-111`), and the stored default is the full
   `https://gitlab.com/api/v4`. go-gitlab does append `api/v4/` when missing (`gitlab.go:564-578`),
   so `https://gitlab.example.com` works, but a bare host does not.
4. **The access token is sent as `PRIVATE-TOKEN`, not `Authorization: Bearer`.**
   `gitlab.NewClient(settings.AccessToken)` (`datasource.go:176`) sets go-gitlab's
   `authType = PrivateToken` (`gitlab.go:254-261`), which emits `PRIVATE-TOKEN: <token>`
   (`gitlab.go:855-858`). The schema's `auth.bearer.token` role is the closest vocabulary match.
5. **`pageLimit` can be written as `NaN`.** `onPageLimitChange` stores
   `parseInt(event.target.value, 10)` (`ConfigEditor.tsx:33`); clearing the field yields
   `parseInt('') === NaN`, which `JSON.stringify` serializes to `null`. The backend then sees a
   missing/zero `pageLimit` and defaults it to `5` (`settings.go:39-41`), so the effect is benign.
6. **Duplicated / malformed access-token validation inside the Page limit subsection.**
   `ConfigEditor.tsx:178-183` re-checks the access token with
   `{!secureJsonData.accessToken || (secureJsonData.accessToken === '' && !secureJsonFields.accessToken && (...))}`
   nested inside the "Page limit" `ConfigSubSection` — a redundant copy of the top-level check
   (`:149-153`) whose `||` short-circuit renders awkwardly. Presentation-only; no schema impact.
7. **`dataSourceName` casing.** `DataSourceDescription` is given `dataSourceName="Gitlab"`
   (`ConfigEditor.tsx:74`) while the plugin name is `"GitLab"` (`plugin.json:4`). Cosmetic.

## Validation performed

- `dsconfig.ParseAndResolveSchemaJSON` + `Schema.Validate()` (Go validator in this repo) — passes
  (via the `TestSchemaConformance` round-trip subtest).
- JSON Schema validation against [`dsconfig/schema.json`](../../dsconfig/schema.json) (draft-07,
  strict `additionalProperties: false`) — passes.
- `go build ./...`, `go vet ./...`, `gofmt -l` (clean), `go test ./...` — pass inside `registry/`
  (schema bundle shape, spec/secure separation, jsonData/struct parity, secure-key parity,
  `SchemaArtifactInSync` drift guard, and the `LoadConfig`/`ApplyDefaults`/`Validate` table tests).
- `tsc --noEmit --strict` on `settings.ts` — clean.
- The pre-existing `dsconfig` and `schema` workspace modules still build and test green.
