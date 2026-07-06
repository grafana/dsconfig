# grafana-github-datasource

Declarative configuration schema for the [GitHub datasource plugin](https://github.com/grafana/github-datasource) (`grafana-github-datasource`).

## Upstream researched

- **Repo**: `github.com/grafana/github-datasource`
- **Ref**: `main`
- **Commit SHA**: `3bb6e75f9f1057fa6efaa3d0f5c7e489b1d5a3d0` (2026-07-02, `docs: add signed commits requirement to CONTRIBUTING.md`)

Every value in [`dsconfig.json`](dsconfig.json) — labels, placeholders, tooltips, option labels/values,
section titles, help markdown, defaults, validations, dependency and required-when expressions,
storage keys, storage targets, value types, group titles, and instructions — is traceable to a
specific `file:line` in the upstream repo at this SHA. See [Field provenance](#field-provenance)
below.

To reproduce this research:

```bash
git clone https://github.com/grafana/github-datasource
cd github-datasource
git checkout 3bb6e75f9f1057fa6efaa3d0f5c7e489b1d5a3d0
```

If upstream `main` has advanced past this SHA, re-diff the sources listed under [Sources
researched](#sources-researched) before merging any changes to this entry.

## Files

| File | Purpose |
| --- | --- |
| [`dsconfig.json`](dsconfig.json) | dsconfig v1 schema — single source of truth for all config fields, groups, effects, and instructions |
| [`settings.ts`](settings.ts) | TypeScript models: `RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig` |
| [`settings.go`](settings.go) | Go `Config` model (flat: jsonData fields + `DecryptedSecureJSONData`), `PluginID`, `SecureJsonDataKey` typed constants, and the `LoadConfig` utility |
| [`schema.go`](schema.go) | k8s-style SDK plugin schema: embeds `dsconfig.json`, converts it via `dsconfig.NewSDKSchema`, and defines `SettingsExamples` for each auth/connection variant |
| [`settings_test.go`](settings_test.go) | Table tests for `LoadConfig`, `ApplyDefaults`, and `Validate` |
| [`conformance_test.go`](conformance_test.go) | Runs `schema.RunPluginTests` — the shared dsconfig conformance suite — against `dsconfig.json` |
| [`schema.gen.json`](schema.gen.json), [`settings.gen.json`](settings.gen.json), [`settings.examples.gen.json`](settings.examples.gen.json) | Committed schema artifacts (regenerate with `go generate ./...` inside this directory; the `SchemaArtifactInSync` conformance subtest guards drift) |

There is no per-entry `go.mod` — every registry entry is a subpackage of the shared
[`registry/`](..) module (`github.com/grafana/dsconfig/registry`).

## Sources researched

Every source below was read at the pinned upstream SHA
(`3bb6e75f9f1057fa6efaa3d0f5c7e489b1d5a3d0`), plus external editor components at the exact
versions the plugin's `package.json` pins.

### Plugin repo (`github.com/grafana/github-datasource@3bb6e75`)

| File | What was read |
| --- | --- |
| `src/plugin.json:4-5,39` | `pluginType` (`id`), `pluginName` (`name`), `docURL` (`info.links[0].url`) |
| `src/views/ConfigEditor.tsx:36-45` | Auth-type and license-type radio option labels/values |
| `src/views/ConfigEditor.tsx:47,54-58` | Virtual `selectedLicense` initial value (React state derivation) |
| `src/views/ConfigEditor.tsx:60-79` | `onSettingUpdate` / `onSettingReset` — how secrets are written |
| `src/views/ConfigEditor.tsx:81-98` | `onAuthChange`, `onLicenseChange` — the multi-field effects captured as virtual effects |
| `src/views/ConfigEditor.tsx:100-105` | `useEffect` that defaults `selectedAuthType` → `personal-access-token` for new datasources |
| `src/views/ConfigEditor.tsx:109-113` | `DataSourceDescription` (`hasRequiredFields={false}`) — why no `required` marks in editor |
| `src/views/ConfigEditor.tsx:117-152` | "Access Token & Permissions" `Collapse` — the `help` drawer markdown, with typos preserved |
| `src/views/ConfigEditor.tsx:155,209,215,234` | Section titles `Authentication` and `Connection` (collapsible → `optional: true`) |
| `src/views/ConfigEditor.tsx:156-208` | Every field's label, placeholder, storage key, `dependsOn`, `SecretInput`/`SecretTextArea`/`Input` component type and geometry (`rows: 7`) |
| `src/views/ConfigEditor.tsx:210-212` | Conditional `SecureSocksProxySettings` — deliberately excluded from this entry |
| `src/views/ConfigEditor.tsx:215-233` | `Connection` collapsible section, `selectedLicense` radio, conditional `githubUrl` field |
| `src/types/config.ts:3-19` | `GitHubLicenseType`, `GitHubAuthType`, `GitHubDataSourceOptions`, `GitHubSecureJsonDataKeys` |
| `pkg/models/settings.go:12-17` | `AuthType` alias + `AuthTypePAT` / `AuthTypeGithubApp` constants |
| `pkg/models/settings.go:19-33` | `Settings` struct fields and their json tags |
| `pkg/models/settings.go:35-59` | `LoadSettings`: parse order, conditional int64 conversion for `AppId`/`InstallationId` under `github-app` auth, decrypted-secret copies, legacy `accessToken`-without-`selectedAuthType` → PAT fallback |
| `pkg/models/settings.go:61-67` | `rawMessageToInt64` (string-or-number normalization mirrored by our `Config.UnmarshalJSON`) |
| `pkg/github/client/client.go:58-72` | `New`: hard-fail with `"access token or app token are required"` when neither credential set is populated (encoded as `requiredWhen` in our schema) |
| `pkg/github/client/client.go:74-97` | `createAppClient`: how `AppIdInt64`, `InstallationIdInt64`, `PrivateKey`, and `GitHubURL` (as `%s/api/v3`) feed the GitHub App transport |
| `pkg/github/client/client.go:99-128` | `createAccessTokenClient`: how `AccessToken` and `GitHubURL` feed the OAuth2/REST/GraphQL clients; asymmetric secure-socks proxy handling |
| `pkg/github/client/client.go:130-146` | `useGitHubEnterprise`: `WithEnterpriseURLs` (normalized) vs raw `%s/api/graphql` (not normalized) |
| `pkg/plugin/instance.go:34-48` | `NewDataSourceInstance` unconditionally sets `datasourceSettings.CachingEnabled = true` — the reason `cachingEnabled` is effectively fixed |
| `package.json` | External component versions (see next table) |

### External editor components

Read at the exact versions pinned in the plugin's `package.json`. Sources checked out at the
corresponding upstream tags.

| Component | Version | Source consulted | What was read |
| --- | --- | --- | --- |
| `DataSourceDescription`, `ConfigSection` | `@grafana/plugin-ui@0.13.1` | `github.com/grafana/plugin-ui` tag `v0.13.1`, `src/components/ConfigEditor/` | Intro text prop shape, `title`/`isCollapsible` props, `hasRequiredFields` behavior |
| `SecureSocksProxySettings` (excluded) | `@grafana/ui@12.4.2` | `github.com/grafana/grafana` tag `v12.4.2`, `packages/grafana-ui/src/components/DataSourceSettings/SecureSocksProxySettings.tsx` | Storage key it writes (`jsonData.enableSecureSocksProxy`) — confirmed and excluded |
| `RadioButtonGroup`, `Field`, `Input`, `Label`, `Collapse`, `SecretInput`, `SecretTextArea`, `Divider` | `@grafana/ui@12.4.2` | grafana/grafana `v12.4.2` `packages/grafana-ui/src/components/` | Prop names (`label`, `placeholder`, `value`, `onChange`, `isConfigured`, `onReset`, `cols`, `rows`, `width`) so we knew which UI attributes to record |
| `onUpdateDatasourceJsonDataOption`, `onUpdateDatasourceSecureJsonDataOption`, `DataSourcePluginOptionsEditorProps` | `@grafana/data@12.4.2` | grafana/grafana `v12.4.2` `packages/grafana-data/src/` | Storage-key semantics of the update helpers used by the config editor |

## Field provenance

Every dsconfig field, traced from its schema `id` to the upstream `file:line` where each of its
label, placeholder, tooltip, default, storage key, and value type is defined. Where a field draws
from multiple lines (e.g. label from one place, default from another), all lines are listed.

| Schema `id` | Storage key | Target | Label source | Placeholder / options / default source | Value type source | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `virtual_selectedLicense` | — | virtual | `ConfigEditor.tsx:216` (`<Label>GitHub License Type</Label>`) | Options `ConfigEditor.tsx:41-45`; `defaultValue` derives from init state `ConfigEditor.tsx:54-58` | Union of 3 strings, `types/config.ts:3` | Storage-computed read expression mirrors `ConfigEditor.tsx:54-58`; write effects mirror `onLicenseChange` `ConfigEditor.tsx:88-98` |
| `jsonData_githubPlan` | `githubPlan` | `jsonData` | — (no UI; managed by `virtual_selectedLicense`) | Values `types/config.ts:3` | `GitHubLicenseType`, `types/config.ts:3` (`string` union) | Tagged `frontend-only` because backend never reads it (see [Upstream findings](#upstream-findings) #1) |
| `jsonData_githubUrl` | `githubUrl` | `jsonData` | `ConfigEditor.tsx:225` (`<Field label="GitHub Enterprise Server URL">`) | `ConfigEditor.tsx:227` (`placeholder="http(s)://HOSTNAME/"`) | `Settings.GitHubURL string` `pkg/models/settings.go:21`; TS `string`, `types/config.ts:9` | `dependsOn`/`requiredWhen` from conditional render at `ConfigEditor.tsx:224` |
| `jsonData_selectedAuthType` | `selectedAuthType` | `jsonData` | `ConfigEditor.tsx:156` (`<Field label="Authentication Type">`) | Options `ConfigEditor.tsx:36-39`; default `personal-access-token` from `useEffect` `ConfigEditor.tsx:100-105` | `AuthType string`, `pkg/models/settings.go:12`; TS `GitHubAuthType`, `types/config.ts:5` | Role `auth.discriminator` |
| `secureJsonData_accessToken` | `accessToken` | `secureJsonData` | `ConfigEditor.tsx:166` (`<Field label="Personal Access Token">`) | `ConfigEditor.tsx:168` (`placeholder="Personal Access Token"`) | `Settings.AccessToken string`, `pkg/models/settings.go:26` | Help drawer verbatim from `ConfigEditor.tsx:117-152`; `dependsOn`/`requiredWhen` from conditional render `ConfigEditor.tsx:165` and backend hard-fail `client.go:71` |
| `jsonData_appId` | `appId` | `jsonData` | `ConfigEditor.tsx:181` (`<Field label="App ID">`) | `ConfigEditor.tsx:183` (`placeholder="App ID"`) | `Settings.AppId json.RawMessage` (string-or-number), `pkg/models/settings.go:28`; TS `string`, `types/config.ts:11` | `dependsOn`/`requiredWhen` from conditional render `ConfigEditor.tsx:179` + backend `ParseInt` `pkg/models/settings.go:39-42` |
| `jsonData_installationId` | `installationId` | `jsonData` | `ConfigEditor.tsx:189` (`<Field label="Installation ID">`) | `ConfigEditor.tsx:191` (`placeholder="Installation ID"`) | `Settings.InstallationId json.RawMessage`, `pkg/models/settings.go:30`; TS `string`, `types/config.ts:12` | Same conditional/required story as `appId` |
| `secureJsonData_privateKey` | `privateKey` | `secureJsonData` | `ConfigEditor.tsx:197` (`<Field label="Private Key">`) | `ConfigEditor.tsx:199` (`placeholder="-----BEGIN CERTIFICATE-----"` — upstream typo preserved); `rows: 7` `ConfigEditor.tsx:204` | `Settings.PrivateKey string`, `pkg/models/settings.go:32` | Role `auth.jwt.signingKey`; textarea with `multiline: true, rows: 7` |
| `jsonData_cachingEnabled` | `cachingEnabled` | `jsonData` | — (no UI) | Default `true` mirrors backend override `pkg/plugin/instance.go:48` | `Settings.CachingEnabled bool`, `pkg/models/settings.go:22` | Tagged `backend-only`; described in schema `description`, no editor `label`/`placeholder` |

## Field inventory summary

| Schema field | Storage key | Target | Editor label | Read by backend? |
| --- | --- | --- | --- | --- |
| `virtual_selectedLicense` | — (virtual) | — | GitHub License Type | — (editor-local state) |
| `jsonData_githubPlan` | `githubPlan` | `jsonData` | — (managed by `virtual_selectedLicense`) | **No — frontend-only** |
| `jsonData_githubUrl` | `githubUrl` | `jsonData` | GitHub Enterprise Server URL | Yes |
| `jsonData_selectedAuthType` | `selectedAuthType` | `jsonData` | Authentication Type | Yes |
| `secureJsonData_accessToken` | `accessToken` | `secureJsonData` | Personal Access Token | Yes |
| `jsonData_appId` | `appId` | `jsonData` | App ID | Yes |
| `jsonData_installationId` | `installationId` | `jsonData` | Installation ID | Yes |
| `secureJsonData_privateKey` | `privateKey` | `secureJsonData` | Private Key | Yes |
| `jsonData_cachingEnabled` | `cachingEnabled` | `jsonData` | — (no UI) | Yes (backend-only) |

### Frontend-only settings

- **`githubPlan`** is written and read only by the config editor to drive the "GitHub License Type"
  radio (`ConfigEditor.tsx:216-222`). The backend never reads it — it infers Enterprise Server
  solely from a non-empty `githubUrl` (`pkg/github/client/client.go:86-96` for github-app,
  `:119-127` for PAT). Selecting "Free, Pro & Team" vs "Enterprise Cloud" changes nothing in
  backend behavior.
- **`virtual_selectedLicense`** does not exist in storage at all: the editor's radio is backed by
  local React state (`selectedLicense` in `ConfigEditor.tsx:54-58`), derived from
  `githubPlan`/`githubUrl` on load. It is modeled here as a `kind: "virtual"` field with a
  `storage.computed.read` expression and `effects` describing the writes it performs.

### Backend-only settings

- **`cachingEnabled`** has no editor UI. It exists in `Settings` (`pkg/models/settings.go:22`),
  but `pkg/plugin/instance.go:48` unconditionally overrides it to `true` after `LoadSettings`
  runs. See [Upstream findings](#upstream-findings) #3.

## Where the types are defined

The configuration types are spread across the plugin and its dependencies — some fields and base
types come from libraries/SDKs rather than the plugin itself. Only config type/field definitions
are listed below — UI components (e.g. `ConfigSection`, `DataSourceDescription`,
`SecureSocksProxySettings`) and functions/helpers (e.g. `LoadSettings`, `rawMessageToInt64`,
`onUpdateDatasourceJsonDataOption`) are omitted even where they are the reason a field exists.

### Frontend (TypeScript)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `GitHubDataSourceOptions` (jsonData), `GitHubAuthType`, `GitHubLicenseType`, `GitHubSecureJsonDataKeys`, `GitHubSecureJsonData` | `src/types/config.ts:1-19` | plugin ([grafana/github-datasource](https://github.com/grafana/github-datasource)) |
| `DataSourceJsonData` (base interface `GitHubDataSourceOptions` extends: `authType`, `defaultRegion`, `profile`, `manageAlerts`, …) | `packages/grafana-data/src/types/datasource.ts` | `@grafana/data` `12.4.2` (grafana/grafana `v12.4.2`) |
| `SecureSocksProxyConfig` (interface adding the `enableSecureSocksProxy` jsonData field; excluded from this entry) | `packages/grafana-ui/src/components/DataSourceSettings/SecureSocksProxySettings.tsx` | `@grafana/ui` `12.4.2` (grafana/grafana `v12.4.2`) |

### Backend (Go)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `Settings` (jsonData + decrypted secrets), `AuthType` (`AuthTypePAT`, `AuthTypeGithubApp`) | `pkg/models/settings.go:12-67` | plugin ([grafana/github-datasource](https://github.com/grafana/github-datasource)) |
| `backend.DataSourceInstanceSettings` (carries `JSONData`, `DecryptedSecureJSONData`, and root fields like `URL`, `BasicAuthEnabled` — unused by this plugin) | `backend/common.go` | `github.com/grafana/grafana-plugin-sdk-go` `v0.292.1` |
| `httpclient.Options` (timeouts, TLS, `ProxyOptions` fields) | `backend/httpclient` | `github.com/grafana/grafana-plugin-sdk-go` `v0.292.1` |
| `LicenseType` has **no backend equivalent** — `githubPlan` exists only in the frontend types | — | — |

The models in this entry flatten that spread into a single Go `Config` type (jsonData fields +
`DecryptedSecureJSONData`) plus a `SecureJsonDataKey` typed constant list. `settings.ts` keeps the
three canonical TypeScript types (`RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig`).
`LicenseType` constants in `settings.go` are derived from the frontend union type since the
backend defines none.

## Modeling decisions

- **Virtual license selector**: `onLicenseChange` (`ConfigEditor.tsx:88-98`) writes `githubPlan`
  and clears `githubUrl` unless "Enterprise Server" is selected. This multi-field write is
  captured as `effects` on the virtual `virtual_selectedLicense` field; `jsonData_githubPlan` is
  tagged `managed-by:virtual_selectedLicense` and has no UI of its own.
- **`requiredWhen` vs the editor**: the editor renders `DataSourceDescription` with
  `hasRequiredFields={false}` (`ConfigEditor.tsx:112`) and marks nothing required, but the backend
  hard-fails without credentials (`client.go:71` returns `"access token or app token are required"`).
  The `requiredWhen` rules encode that backend contract; an instruction records the editor
  discrepancy.
- **Help drawer**: the editor's top-level "Access Token & Permissions" `Collapse`
  (`ConfigEditor.tsx:117-152`) is attached as the `help` drawer of `secureJsonData_accessToken`,
  with the markdown preserved verbatim (including upstream typos — see below).
- **Secure Socks Proxy excluded**: the editor conditionally renders `SecureSocksProxySettings`
  (`ConfigEditor.tsx:210-212`) writing `jsonData.enableSecureSocksProxy` when the Grafana instance
  has `secureSocksDSProxyEnabled`, and both backend auth paths honor it. The field is deliberately
  omitted from this registry entry.
- **Field ID naming convention**: IDs are prefixed with their storage target for easy
  discoverability — `root_`, `jsonData_`, or `secureJsonData_` (and `virtual_` for virtual fields,
  which have no storage target) — followed by the camelCase storage key, e.g. `jsonData_appId`,
  `secureJsonData_accessToken`. The `key` property keeps the plugin's raw storage key (`appId`) —
  `id` is the schema reference, `key` is the storage contract.
- **Flat `Config` in Go**: `settings.go` collapses jsonData fields and decrypted secrets onto a
  single `Config` struct (mirroring the upstream `Settings` in `pkg/models/settings.go:19-33`
  verbatim, json tags included) with `AppId`/`InstallationId` normalized from string-or-number to
  `string` via a custom `UnmarshalJSON`, plus parsed `AppIdInt64`/`InstallationIdInt64`. Root-level
  datasource fields (`url`, `basicAuth`, etc.) are not carried because the plugin does not use
  them. `settings.ts` keeps the three canonical TS types.
- **`SecureJsonDataConfig` is a key list**: secure values are write-only, so the secure type is
  just the array of secret key names (`accessToken`, `privateKey`); consumers read
  `secureJsonFields` to see what is configured.

## SDK plugin schema and k8s-style examples (`schema.go`)

`NewSchema()` assembles the `grafana-plugin-sdk-go` `pluginschema.PluginSchema` bundle (the
k8s-style schema Grafana's datasource API server serves as `{apiVersion}.json`, `v0alpha1` today)
from the embedded `dsconfig.json`: root fields plus a nested `jsonData` object become the OpenAPI
settings `spec`, secure fields become `secureValues`, and virtual fields are skipped.

`SettingsExamples()` provides the default configuration plus one k8s-style example per
authentication type and connection variant. Each example is a full instance-settings object with
the plugin configuration nested under `jsonData` and the relevant write-only secrets under
`secureJsonData` (placeholder values to be replaced with real secrets; the default example — keyed
by the empty string `""` — carries an empty `accessToken` to show what must be filled in):

| Example | Auth | Connection | `secureJsonData` |
| --- | --- | --- | --- |
| `""` (default) | Personal Access Token (schema defaults) | GitHub.com (Free, Pro & Team) | `accessToken` (empty) |
| `personalAccessToken` | Personal Access Token | GitHub.com (Free, Pro & Team) | `accessToken` |
| `githubApp` | GitHub App | GitHub.com (Free, Pro & Team) | `privateKey` |
| `enterpriseCloud` | Personal Access Token | Enterprise Cloud (same endpoints as GitHub.com) | `accessToken` |
| `enterpriseServer` | Personal Access Token | Enterprise Server (`githubUrl`) | `accessToken` |
| `githubAppEnterpriseServer` | GitHub App | Enterprise Server (`githubUrl`) | `privateKey` |
| `legacyAccessTokenOnly` | Legacy: token with no auth type | GitHub.com | `accessToken` |

## `LoadConfig` utility (`settings.go`)

`LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error)`
runs the full three-phase load flow on a datasource instance's settings and returns a
fully-defaulted, validated `Config`:

1. **Parse** — unmarshal jsonData into `Config` (the `Config.UnmarshalJSON` normalizes the legacy
   string-or-number `appId` / `installationId` — mirrors `rawMessageToInt64` at
   `pkg/models/settings.go:61-67`), copy decrypted secrets into `DecryptedSecureJSONData`, run the
   upstream legacy fallback that promotes a lone `accessToken` to `personal-access-token`
   (`pkg/models/settings.go:53-57`), and — under `github-app` auth only — parse `AppIdInt64` /
   `InstallationIdInt64` (`pkg/models/settings.go:39-45`).
2. **`ApplyDefaults`** — fill a curated set of zero-valued discriminators with the same defaults
   the editor writes for a fresh datasource (`SelectedAuthType=AuthTypePAT` per
   `ConfigEditor.tsx:100-105`; `GithubPlan=LicenseTypeBasic` per `ConfigEditor.tsx:54-58` initial
   state).
3. **`Validate`** — enforce the runtime contract (auth method + its required inputs, and
   `githubUrl` when the plan is Enterprise Server). Errors are joined so every problem surfaces at
   once.

Everything is logged via `backend.Logger.FromContext(ctx)` with `datasource_uid`,
`datasource_name`, and `plugin` labels so log lines carry request context.

This is the intended shape for the plugin's own upstream `LoadSettings` to sync to: a load returns
a config that is safe to use, or an error explaining why it isn't.

### Direct access to individual phases

`(*Config).ApplyDefaults()` and `(Config).Validate() error` are still exported for callers that
want to compose them themselves (e.g. provisioning preview, schema-example round-trip, tests that
need to distinguish parse-level from policy-level errors). Skip them by never calling `LoadConfig`
in those flows — assemble a `Config` directly.

## Upstream findings

Potential bugs, misleading UX, and consistency issues discovered while researching upstream. All
preserved verbatim in the schema — the schema records what the plugin **does**, not what it
**should** do; these notes exist so reviewers can reproduce each finding and decide separately
whether to fix upstream.

1. **`githubPlan` is dead weight for the backend.** `pkg/github/client/client.go:86-96,119-127` —
   the backend decides Enterprise Server purely from `settings.GitHubURL != ""`, so a provisioned
   datasource with `githubPlan: "github-enterprise-server"` but no `githubUrl` silently behaves
   like github.com.
2. **Stale `githubUrl` overrides the stored plan in the editor.** The load derivation at
   `src/views/ConfigEditor.tsx:54-58` is `githubPlan === 'github-enterprise-server' || githubUrl
   ? 'github-enterprise-server' : githubPlan || 'github-basic'` — a datasource with `githubPlan:
   "github-basic"` but a leftover `githubUrl` (possible via provisioning or the API) displays and
   behaves as Enterprise Server.
3. **`cachingEnabled` cannot actually be disabled.** `pkg/plugin/instance.go:48` unconditionally
   sets `datasourceSettings.CachingEnabled = true` after `LoadSettings`, so the stored value is
   ignored and there is no way to turn the caching wrapper off.
4. **Misleading Private Key placeholder.** `src/views/ConfigEditor.tsx:199` — the `SecretTextArea`
   placeholder is `-----BEGIN CERTIFICATE-----`, but a GitHub App private key is an RSA private
   key PEM (`-----BEGIN RSA PRIVATE KEY-----`), not a certificate. Preserved verbatim in the
   schema.
5. **Trailing slash in `githubUrl` produces double-slash API URLs.** `src/views/ConfigEditor.tsx:227`
   suggests `http(s)://HOSTNAME/`, yet `pkg/github/client/client.go:94` builds `fmt.Sprintf("%s/api/v3",
   settings.GitHubURL)` for the GitHub App transport and `client.go:143` builds
   `fmt.Sprintf("%s/api/graphql", settings.GitHubURL)` for GraphQL, yielding `HOSTNAME//api/…` when the
   placeholder is followed literally. Only the REST client (`WithEnterpriseURLs`, `client.go:136`)
   normalizes the URL.
6. **Incomplete GitHub App config hard-fails settings load.** `pkg/models/settings.go:39-45` — with
   `selectedAuthType: "github-app"`, `LoadSettings` runs `strconv.ParseInt` (via `rawMessageToInt64`,
   `settings.go:61-67`) on `appId`/`installationId`; missing or non-numeric values return `"error
   parsing app id"` and instance creation fails. The editor performs no validation on these inputs
   before save (`src/views/ConfigEditor.tsx:181-196`).
7. **Typos in the editor help text** (preserved verbatim in the schema's `help` markdown to match
   the UI): `src/views/ConfigEditor.tsx:118` "a access token" (should be "an access token"),
   `:136` "the Github documentation." (GitHub), `:147-148` "Meta data" (GitHub calls the
   permission "Metadata"), and inconsistent mid-sentence capitalization ("Select" `:141`,
   "Ensure" `:145`).
8. **`SecretInput` update quirk for the access token.** `src/views/ConfigEditor.tsx:172` — the
   token's `onChange` is `onSettingUpdate('accessToken', false)`, which sets
   `secureJsonFields.accessToken = false` while typing — intentional to keep the input editable,
   but it means `secureJsonFields` temporarily reports the secret as unconfigured until save.
9. **Asymmetric proxy wiring between auth paths** (works, but inconsistent): the GitHub App path
   gets secure-socks support implicitly via SDK `httpclient.New(opts)` at
   `pkg/github/client/client.go:75`, while the PAT path manually rebuilds the `oauth2` transport
   only when `proxy.New(opts.ProxyOptions).SecureSocksProxyEnabled()` at `client.go:106-118`.
10. **Legacy auth-type fallback is one-way.** `pkg/models/settings.go:53-57` defaults
    `SelectedAuthType` to `personal-access-token` when an `accessToken` exists with no auth type,
    but there is no equivalent fallback for old GitHub App configs; the schema records the PAT
    fallback as an instruction.

## Validation performed

- `dsconfig.ParseAndResolveSchemaJSON` + `Schema.Validate()` (Go validator in this repo) — passes.
- JSON Schema validation against [`dsconfig/schema.json`](../../dsconfig/schema.json) (draft-07,
  `additionalProperties: false`) — passes.
- `go test ./...` on this module — passes (schema bundle shape, secure values, examples,
  `LoadConfig` incl. legacy fallback and id parsing, `SchemaArtifactInSync` guard).
- `settings.go`/`schema.go`: `go build`, `go vet`, `gofmt` — clean.
- `settings.ts`: `tsc --noEmit --strict` — clean.
