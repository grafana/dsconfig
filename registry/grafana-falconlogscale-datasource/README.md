# grafana-falconlogscale-datasource

Declarative configuration schema for the [Falcon LogScale datasource
plugin](https://github.com/grafana/falconlogscale-datasource)
(`grafana-falconlogscale-datasource`).

## Upstream researched

- **Repo**: `github.com/grafana/falconlogscale-datasource`
- **Ref**: `main`
- **Commit SHA**: `1e83b294390e6d93865156cec0c72ed252e791c0`

Every value in [`dsconfig.json`](dsconfig.json) — labels, placeholders, tooltips, option
labels/values, section titles, defaults, validations, dependency and required-when expressions,
storage keys, storage targets, value types, group titles, and instructions — is traceable to a
specific `file:line` in the upstream repo at this SHA.

To reproduce this research:

```bash
git clone https://github.com/grafana/falconlogscale-datasource
cd falconlogscale-datasource
git checkout 1e83b294390e6d93865156cec0c72ed252e791c0
```

If upstream `main` has advanced past this SHA, re-diff the sources listed under [Sources
researched](#sources-researched) before merging any changes to this entry.

## Files

| File | Purpose |
| --- | --- |
| [`dsconfig.json`](dsconfig.json) | dsconfig v1 schema — single source of truth for all config fields, groups, effects, and instructions |
| [`settings.ts`](settings.ts) | TypeScript models: `RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig` |
| [`settings.go`](settings.go) | Go `Config` model (flat: root + jsonData fields + `DecryptedSecureJSONData`), `PluginID`, `SecureJsonDataKey` typed constants, discriminator enums (`DataSourceMode`, `AuthMethod`), and the `LoadConfig` utility |
| [`schema.go`](schema.go) | k8s-style SDK plugin schema: embeds `dsconfig.json`, converts it via `dsconfig.NewSDKSchema`, and defines `SettingsExamples` for each auth/mode variant |
| [`settings_test.go`](settings_test.go) | Table tests for `LoadConfig`, `ApplyDefaults`, and `Validate` |
| [`conformance_test.go`](conformance_test.go) | Runs `schema.RunPluginTests` — the shared dsconfig conformance suite — against `dsconfig.json` |
| [`schema.gen.json`](schema.gen.json), [`settings.gen.json`](settings.gen.json), [`settings.examples.gen.json`](settings.examples.gen.json) | Committed schema artifacts (regenerate with `go generate ./...` inside this directory; the `SchemaArtifactInSync` conformance subtest guards drift) |

There is no per-entry `go.mod` — every registry entry is a subpackage of the shared
[`registry/`](..) module (`github.com/grafana/dsconfig/registry`).

## Sources researched

Every source below was read at the pinned upstream SHA
(`1e83b294390e6d93865156cec0c72ed252e791c0`), plus external editor components at the exact
versions the plugin's `package.json` pins.

### Plugin repo (`github.com/grafana/falconlogscale-datasource@1e83b29`)

| File | What was read |
| --- | --- |
| `src/plugin.json:4-5,26-28` | `pluginType` (`id`), `pluginName` (`name`), `docURL` (`info.links[0].url`) |
| `src/types.ts:5-8` | `DataSourceMode` enum (`LogScale`, `NGSIEM`) |
| `src/types.ts:10-23` | `LogScaleOptions` interface (frontend jsonData shape) |
| `src/types.ts:25-29` | `SecretLogScaleOptions` interface (accessToken, basicAuthPassword, oauth2ClientSecret) |
| `src/types.ts:52` | `NGSIEMRepos = ['search-all', 'investigate_view', 'third-party']` |
| `src/components/ConfigEditor/ConfigEditor.tsx:59-60` | Mode default = `LogScale`; `isNGSIEMMode` derived flag |
| `src/components/ConfigEditor/ConfigEditor.tsx:61-68` | `clearAuthSettings()` — resets `authenticateWithToken`, `oauth2`, `oauth2ClientId`, `oauthPassThru` on every auth-selector change |
| `src/components/ConfigEditor/ConfigEditor.tsx:127-137,193-200` | NGSIEM mode auto-sets `defaultRepository = 'search-all'` |
| `src/components/ConfigEditor/ConfigEditor.tsx:140-169` | `logscaleTokenComponent` — Token `SecretInput` label, placeholder, and the `authenticateWithToken=true` write on save |
| `src/components/ConfigEditor/ConfigEditor.tsx:180-183` | `modeOptions` — Select label/value for LogScale / NGSIEM |
| `src/components/ConfigEditor/ConfigEditor.tsx:219-223` | `DataSourceDescription` (`hasRequiredFields`) |
| `src/components/ConfigEditor/ConfigEditor.tsx:225` | `ConnectionSettings` (URL input) |
| `src/components/ConfigEditor/ConfigEditor.tsx:227-232` | Mode field label ("Mode") and description ("Select the data source mode. NGSIEM mode only supports OAuth2 client secret authentication.") |
| `src/components/ConfigEditor/ConfigEditor.tsx:234-279` | `Auth` (from @grafana/plugin-ui, via `convertLegacyAuthProps`) — custom methods `custom-token` and `custom-oauth-client-secret`; `visibleMethods` differs per mode |
| `src/components/ConfigEditor/ConfigEditor.tsx:281-287` | "Advanced settings" collapsible → `AdvancedHttpSettings` (Allowed cookies, Timeout) |
| `src/components/ConfigEditor/ConfigEditor.tsx:289-322` | "Additional settings" collapsible (default open) → `DefaultRepository`, `DataLinks`, `incrementalQuerying` switch, `incrementalQueryOverlapWindow` input |
| `src/components/ConfigEditor/ConfigEditor.tsx:348-372` | Conditional `SecureSocksProxySettings` — deliberately excluded from this entry |
| `src/components/ConfigEditor/OAuth2Component.tsx:58-66` | Client ID label ("Client ID"), description ("The OAuth2 client ID"), placeholder ("Client ID") |
| `src/components/ConfigEditor/OAuth2Component.tsx:67-81` | Client Secret label ("Client Secret"), description ("The OAuth2 client secret"), placeholder ("Client Secret") |
| `src/components/ConfigEditor/DefaultRepository.tsx:58-83` | "Default Repository" select + "Load Repositories" button (button hidden in NGSIEM mode) |
| `src/components/DataLinks/DataLinks.tsx:29-33` | "Data links" section title and description ("Add links to existing fields. Links will be shown in log row details next to the field value.") |
| `src/components/DataLinks/DataLink.tsx:46-88` | Per-item labels (Field, Label, Regex, URL) and tooltips |
| `src/components/DataLinks/types.ts:1-7` | `DataLinkConfig` type |
| `pkg/plugin/settings.go:11-26` | Backend `Settings` struct (fields, json tags) — narrower than the frontend jsonData shape |
| `pkg/plugin/settings.go:32-62` | `LoadSettings`: parse jsonData, overwrite `BaseURL` from `config.URL`, hard-fail on empty URL, derive `GraphqlEndpoint` / `RestEndpoint` based on `AuthenticateWithToken`, copy decrypted secrets |
| `pkg/plugin/plugin.go:15-50` | `NewDataSourceInstance`: builds HTTP + streaming client, forwards headers via `OAuthPassThru` |
| `pkg/plugin/plugin.go:52-70` | `newClient`: appends `/humio` to `BaseURL` when `Mode == "NGSIEM"`; passes token or OAuth2 config to `humio.NewClient` |
| `pkg/plugin/healthcheck.go:26-48` | Mode-conditional health check: NGSIEM uses `OauthClientSecretHealthCheck`; LogScale lists repositories via GraphQL |
| `package.json` | External component versions (see next table) |

### External editor components

Read at the versions pinned in the plugin's `package.json`.

| Component | Version | Source consulted | What was read |
| --- | --- | --- | --- |
| `DataSourceDescription`, `ConnectionSettings`, `Auth` (via `convertLegacyAuthProps`), `AuthMethod` enum, `AdvancedHttpSettings`, `ConfigSection` | `@grafana/plugin-ui@^0.13.0` | `github.com/grafana/plugin-ui` tag `v0.13.x`, `src/components/ConfigEditor/` | `AuthMethod.BasicAuth`, `AuthMethod.OAuthForward`, `AuthMethod.NoAuth`, custom-method registration, URL input label ("URL"), Basic Auth field labels ("User", "Password"), Allowed cookies / Timeout labels |
| `SecureSocksProxySettings` (excluded) | `@grafana/ui@^11.3.2` | `github.com/grafana/grafana` tag `v11.x`, `packages/grafana-ui/src/components/DataSourceSettings/SecureSocksProxySettings.tsx` | Storage key it writes (`jsonData.enableSecureSocksProxy`) — confirmed and excluded |
| `Field`, `Input`, `SecretInput`, `Select`, `Switch`, `Button`, `DataLinkInput` | `@grafana/ui@^11.3.2` | grafana/grafana `v11.x` `packages/grafana-ui/src/components/` | Prop shapes for the LogScale token component, OAuth2 component, DefaultRepository component, and DataLink item editor |

## Field provenance

Every dsconfig field, traced from its schema `id` to the upstream `file:line`.

| Schema `id` | Storage key | Target | Label source | Options / placeholder / default source | Value type source | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `root_url` | `url` | `root` | `@grafana/plugin-ui` `ConnectionSettings.tsx` — "URL" | `settings.URL` at runtime | `backend.DataSourceInstanceSettings.URL` (string) | `requiredWhen: "true"` — backend returns `errEmptyURL` at `pkg/plugin/settings.go:41` |
| `jsonData_mode` | `mode` | `jsonData` | `ConfigEditor.tsx:228` (`<Field label="Mode">`) | Options `ConfigEditor.tsx:180-183`; default `LogScale` from `ConfigEditor.tsx:59` | `DataSourceMode` enum `src/types.ts:5-8`; backend `Settings.Mode string` `pkg/plugin/settings.go:20` | Description verbatim from `ConfigEditor.tsx:229` |
| `virtual_authMethod` | `authMethod` | virtual | Derived from Auth component labels (`ConfigEditor.tsx:236-249`) | Options mirror `visibleMethods` (`ConfigEditor.tsx:274-278`) plus the two `customMethods`; default `custom-token` from `ConfigEditor.tsx:176-178` | Union of 4 strings | Storage-computed `read` derives the method from the four flags; `effects` mirror `clearAuthSettings` + the selected flag write |
| `jsonData_authenticateWithToken` | `authenticateWithToken` | `jsonData` | — (no UI; managed by `virtual_authMethod`) | Written by `logscaleTokenComponent` at `ConfigEditor.tsx:156` | `Settings.AuthenticateWithToken bool` `pkg/plugin/settings.go:14` | Tagged `managed-by:virtual_authMethod` |
| `jsonData_oauth2` | `oauth2` | `jsonData` | — (managed by `virtual_authMethod`) | Written by `OAuth2Component.tsx:19,34` | `Settings.OAuth2 bool` `pkg/plugin/settings.go:17` | Tagged `managed-by:virtual_authMethod` |
| `jsonData_oauthPassThru` | `oauthPassThru` | `jsonData` | — (managed by `virtual_authMethod`) | Written by the AuthMethod.OAuthForward branch at `ConfigEditor.tsx:263` | `Settings.OAuthPassThru bool` `pkg/plugin/settings.go:16` | Role `auth.forwardOAuthToken.enabled` |
| `root_basicAuth` | `basicAuth` | `root` | — (managed by `virtual_authMethod`) | Written by @grafana/plugin-ui's `Auth` component | `backend.DataSourceInstanceSettings.BasicAuthEnabled bool` | Role `auth.basic.enabled` |
| `secureJsonData_accessToken` | `accessToken` | `secureJsonData` | `ConfigEditor.tsx:140,144` (`<Field label="Token">`, `SecretInput label='Token'`) | Placeholder `ConfigEditor.tsx:146` (`"Token"`) | `Settings.AccessToken string` `pkg/plugin/settings.go:13` + `secureSettings["accessToken"]` `settings.go:56` | Role `auth.bearer.token` |
| `jsonData_oauth2ClientId` | `oauth2ClientId` | `jsonData` | `OAuth2Component.tsx:58` (`<Field label="Client ID">`) | Description `OAuth2Component.tsx:58`; placeholder `OAuth2Component.tsx:62` (`"Client ID"`) | `Settings.OAuth2ClientID string` `pkg/plugin/settings.go:18` | Role `auth.oauth2.clientId` |
| `secureJsonData_oauth2ClientSecret` | `oauth2ClientSecret` | `secureJsonData` | `OAuth2Component.tsx:67` (`<Field label="Client Secret">`) | Description `OAuth2Component.tsx:67`; placeholder `OAuth2Component.tsx:73` (`"Client Secret"`) | `Settings.OAuth2ClientSecret string` `pkg/plugin/settings.go:19` + `secureSettings["oauth2ClientSecret"]` `settings.go:57` | Role `auth.oauth2.clientSecret` |
| `root_basicAuthUser` | `basicAuthUser` | `root` | @grafana/plugin-ui `Auth`/`BasicAuth` — "User" | Placeholder — `"User"` from @grafana/plugin-ui | `backend.DataSourceInstanceSettings.BasicAuthUser string` used by backend at `pkg/plugin/settings.go:59` | Role `auth.basic.username` |
| `secureJsonData_basicAuthPassword` | `basicAuthPassword` | `secureJsonData` | @grafana/plugin-ui `Auth`/`BasicAuth` — "Password" | Placeholder — `"Password"` from @grafana/plugin-ui | `secureSettings["basicAuthPassword"]` `pkg/plugin/settings.go:60` | Role `auth.basic.password` |
| `jsonData_baseUrl` | `baseUrl` | `jsonData` | — (no UI) | Written by `logscaleTokenComponent` at `ConfigEditor.tsx:155` | `Settings.BaseURL string` `pkg/plugin/settings.go:12` (upstream tags this `json:"baseURL"` — case-insensitive match — see [Upstream findings](#upstream-findings) #1) | Tagged `frontend-only`; backend overwrites from `config.URL` immediately after unmarshal |
| `jsonData_defaultRepository` | `defaultRepository` | `jsonData` | `DefaultRepository.tsx:60` (`<Field label="Default Repository">`) | Options loaded dynamically via `/api/datasources/.../resources/repositories`; in NGSIEM mode restricted to `NGSIEMRepos` (`src/types.ts:52`) and auto-set to `"search-all"` (`ConfigEditor.tsx:133,199`) | `LogScaleOptions.defaultRepository?: string` `src/types.ts:17` | Not read by backend; consumed by frontend `DataSource.ts:62,168` |
| `jsonData_dataLinks` | `dataLinks` | `jsonData` | `DataLinks.tsx:29` (`"Data links"` heading) | Description `DataLinks.tsx:32-33`; item fields from `DataLink.tsx` | `DataLinkConfig[]` `src/components/DataLinks/types.ts:1-7` | Not read by backend; consumed by frontend result transformer (`src/logs.ts`) |
| `jsonData_incrementalQuerying` | `incrementalQuerying` | `jsonData` | `ConfigEditor.tsx:313` (`<Field label="Incremental querying (experimental)">`) | Description `ConfigEditor.tsx:314`; default `false` from `?? false` at `ConfigEditor.tsx:318` | `LogScaleOptions.incrementalQuerying?: boolean` `src/types.ts:21` | Not read by backend; consumed by `src/DataSource.ts:78,100` |
| `jsonData_incrementalQueryOverlapWindow` | `incrementalQueryOverlapWindow` | `jsonData` | `ConfigEditor.tsx:326` (`<Field label="Query overlap window">`) | Description `ConfigEditor.tsx:327`; default `"10m"` from `?? '10m'` at `ConfigEditor.tsx:56` | `LogScaleOptions.incrementalQueryOverlapWindow?: string` `src/types.ts:22` | `dependsOn: jsonData_incrementalQuerying == true` mirrors conditional render at `ConfigEditor.tsx:324` |
| `jsonData_keepCookies` | `keepCookies` | `jsonData` | @grafana/plugin-ui `AdvancedHttpSettings` — "Allowed cookies" | Placeholder — "New cookie (hit enter to add)" from @grafana/plugin-ui | `Settings.KeepCookies []string` `pkg/plugin/settings.go:15` | Written by editor via @grafana/plugin-ui's `AdvancedHttpSettings` |
| `jsonData_timeout` | `timeout` | `jsonData` | @grafana/plugin-ui `AdvancedHttpSettings` — "Timeout" | Placeholder — "Timeout in seconds" from @grafana/plugin-ui | `httpclient.Options.Timeouts.Timeout` (SDK) | Role `transport.timeoutSeconds`; consumed by the SDK via `settings.HTTPClientOptions(ctx)` |

## Field inventory summary

| Schema field | Storage key | Target | Editor label | Read by backend? |
| --- | --- | --- | --- | --- |
| `root_url` | `url` | `root` | URL | Yes (`pkg/plugin/settings.go:38`) |
| `jsonData_mode` | `mode` | `jsonData` | Mode | Yes (`pkg/plugin/settings.go:20`, `pkg/plugin/plugin.go:54`, `pkg/plugin/healthcheck.go:26`) |
| `virtual_authMethod` | — | virtual | Authentication method | — (editor-local state) |
| `jsonData_authenticateWithToken` | `authenticateWithToken` | `jsonData` | — (managed) | Yes (`pkg/plugin/settings.go:14,44-50`) |
| `jsonData_oauth2` | `oauth2` | `jsonData` | — (managed) | Yes (`pkg/plugin/settings.go:17`, `pkg/plugin/plugin.go:65`) |
| `jsonData_oauthPassThru` | `oauthPassThru` | `jsonData` | — (managed) | Yes (`pkg/plugin/settings.go:16`, `pkg/plugin/plugin.go:24,31`) |
| `root_basicAuth` | `basicAuth` | `root` | — (managed) | Indirectly (via `config.BasicAuthEnabled` and `config.BasicAuthUser` in `pkg/plugin/settings.go:59`) |
| `secureJsonData_accessToken` | `accessToken` | `secureJsonData` | Token | Yes (`pkg/plugin/settings.go:56`) |
| `jsonData_oauth2ClientId` | `oauth2ClientId` | `jsonData` | Client ID | Yes (`pkg/plugin/settings.go:18`, `pkg/plugin/plugin.go:66`) |
| `secureJsonData_oauth2ClientSecret` | `oauth2ClientSecret` | `secureJsonData` | Client Secret | Yes (`pkg/plugin/settings.go:57`) |
| `root_basicAuthUser` | `basicAuthUser` | `root` | User | Yes (`pkg/plugin/settings.go:59`) |
| `secureJsonData_basicAuthPassword` | `basicAuthPassword` | `secureJsonData` | Password | Yes (`pkg/plugin/settings.go:60`) |
| `jsonData_baseUrl` | `baseUrl` | `jsonData` | — (no UI) | **No — frontend-only** (backend overwrites from `config.URL`) |
| `jsonData_defaultRepository` | `defaultRepository` | `jsonData` | Default Repository | **No — frontend-only** (used by `src/DataSource.ts:62,168`) |
| `jsonData_dataLinks` | `dataLinks` | `jsonData` | Data links | **No — frontend-only** (used by `src/logs.ts`) |
| `jsonData_incrementalQuerying` | `incrementalQuerying` | `jsonData` | Incremental querying (experimental) | **No — frontend-only** |
| `jsonData_incrementalQueryOverlapWindow` | `incrementalQueryOverlapWindow` | `jsonData` | Query overlap window | **No — frontend-only** |
| `jsonData_keepCookies` | `keepCookies` | `jsonData` | Allowed cookies | Yes (`pkg/plugin/settings.go:15`; also read by SDK's `HTTPClientOptions`) |
| `jsonData_timeout` | `timeout` | `jsonData` | Timeout | Yes (via SDK `httpclient.Options.Timeouts.Timeout`) |

### Frontend-only settings

- **`baseUrl`** is a snapshot of `root.url` written by the LogScale token authentication
  component (`ConfigEditor.tsx:155`). The backend overwrites `settings.BaseURL` from `config.URL`
  before use (`pkg/plugin/settings.go:39`) — the stored `baseUrl` value is never consulted.
- **`defaultRepository`** is consumed by `src/DataSource.ts:62,168` to fill an empty
  `LogScaleQuery.repository`. Backend queries carry the repository per-request.
- **`dataLinks`** is consumed by the frontend result transformer at `src/logs.ts` to build data
  links on log rows; the backend never inspects it.
- **`incrementalQuerying`** and **`incrementalQueryOverlapWindow`** control the frontend
  `QueryCache` at `src/DataSource.ts:60,100-105`.

### Backend-only settings

- None. The backend's `Settings` struct is a strict subset of what the editor writes; there are no
  fields the backend consumes without a matching editor path.

## Modeling decisions

- **Virtual auth-method selector**: the editor's `Auth` component (via `convertLegacyAuthProps`)
  exposes four visible methods across the two modes — `custom-token`, `custom-oauth-client-secret`,
  `BasicAuth`, `OAuthForward` — but there is no single discriminator field. The
  `virtual_authMethod` field encodes the derivation (`storage.computed.read`) and the multi-field
  writes (`effects`) that `clearAuthSettings()` (`ConfigEditor.tsx:61-68`) plus each auth branch
  perform. The four flag fields (`authenticateWithToken`, `oauth2`, `oauthPassThru`, `basicAuth`)
  are tagged `managed-by:virtual_authMethod`.
- **Mode as a first-class jsonData field**: `mode` is a plain enum field with `defaultValue:
  "LogScale"` — no virtual layer needed, because the mode maps 1:1 to a stored jsonData key.
  NGSIEM's auto-set of `defaultRepository = "search-all"` is enforced in Go's `ApplyDefaults` (so
  provisioned NGSIEM datasources land in the same state as UI-saved ones) but not encoded as a
  virtual effect on `mode` — the schema keeps the two fields independent.
- **Frontend-only `baseUrl` retained**: rather than dropping the field, it is kept as a
  `jsonData_baseUrl` field with `tags: ["frontend-only"]` and a description explaining what
  actually happens on the backend. This preserves round-trip fidelity for datasources exported via
  provisioning.
- **`requiredWhen` vs the editor**: the editor renders `DataSourceDescription` with
  `hasRequiredFields` (truthy in `ConfigEditor.tsx:222`) but no field-level required markers. The
  backend hard-fails without a URL (`pkg/plugin/settings.go:41`) and — implicitly, via the health
  check — needs auth-specific credentials to pass. `requiredWhen` rules on the secret and auth
  fields encode that contract.
- **Secure Socks Proxy excluded**: the editor conditionally renders the
  `SecureSocksProxySettings` block (`ConfigEditor.tsx:348-372`) writing
  `jsonData.enableSecureSocksProxy`. The field is deliberately omitted from this registry entry
  per project convention.
- **Data links vs derived fields**: the plugin uses a local `DataLinks`/`DataLink` component
  family (not `@grafana/ui`'s DerivedFields). The schema models `dataLinks` as an array of objects
  with `field`, `label`, `matcherRegex`, `url`, and optional `datasourceUid` — matching
  `src/components/DataLinks/types.ts:1-7`.
- **Field ID naming convention**: `<target>_<camelCaseKey>` (`root_url`, `jsonData_oauth2`,
  `secureJsonData_accessToken`, `virtual_authMethod`). The `key` property preserves the raw
  storage key.
- **Flat `Config` in Go**: `settings.go` collapses root fields, jsonData fields, and decrypted
  secrets onto a single `Config` struct. Root fields are tagged `json:"-"` and populated in
  `LoadConfig` from `backend.DataSourceInstanceSettings` (URL, BasicAuthEnabled, BasicAuthUser).
- **`SecureJsonDataConfig` is a key list**: secure values are write-only, so the secure type is
  the array of secret key names (`accessToken`, `oauth2ClientSecret`, `basicAuthPassword`).

## Where the types are defined

Only config type/field definitions are listed below — UI components and helpers are omitted even
where they are the reason a field exists.

### Frontend (TypeScript)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `DataSourceMode`, `LogScaleOptions` (jsonData), `SecretLogScaleOptions`, `NGSIEMRepos` | `src/types.ts:5-52` | plugin ([grafana/falconlogscale-datasource](https://github.com/grafana/falconlogscale-datasource)) |
| `DataLinkConfig` (per-item shape for `jsonData.dataLinks`) | `src/components/DataLinks/types.ts:1-7` | plugin |
| `DataSourceJsonData` (base interface `LogScaleOptions` extends) | `packages/grafana-data/src/types/datasource.ts` | `@grafana/data` `^11.6.0` |
| `AuthMethod` enum (`BasicAuth`, `OAuthForward`, `NoAuth`) | `packages/plugin-ui/src/components/ConfigEditor/Auth/` | `@grafana/plugin-ui@^0.13.0` |
| `SecureSocksProxyConfig` (adds `jsonData.enableSecureSocksProxy`; excluded from this entry) | `packages/grafana-ui/src/components/DataSourceSettings/SecureSocksProxySettings.tsx` | `@grafana/ui@^11.3.2` |

### Backend (Go)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `Settings` (jsonData + decrypted secrets, plus derived `GraphqlEndpoint` / `RestEndpoint`) | `pkg/plugin/settings.go:11-26` | plugin ([grafana/falconlogscale-datasource](https://github.com/grafana/falconlogscale-datasource)) |
| `backend.DataSourceInstanceSettings` (carries `JSONData`, `DecryptedSecureJSONData`, root fields `URL`, `BasicAuthEnabled`, `BasicAuthUser`) | `backend/common.go` | `github.com/grafana/grafana-plugin-sdk-go` `v0.292.1` |
| `httpclient.Options` (Timeouts, TLS, ProxyOptions, ForwardHTTPHeaders) | `backend/httpclient` | `github.com/grafana/grafana-plugin-sdk-go` `v0.292.1` |
| `humio.Config`, `humio.OAuth2Config` (Address, Token, OAuth2ClientID / OAuth2ClientSecret) | `pkg/humio/client.go` | plugin |
| `DataLinkConfig` (frontend-only; no backend equivalent) | — | — |

## Settings examples matrix

| Example key | Auth method | Mode | secureJsonData |
| --- | --- | --- | --- |
| `""` (default) | LogScale token (empty placeholder) | LogScale | `accessToken` (empty) |
| `logscaleToken` | LogScale token | LogScale | `accessToken` |
| `logscaleOAuth2Client` | OAuth2 client credentials | LogScale | `oauth2ClientSecret` |
| `logscaleBasicAuth` | HTTP Basic auth | LogScale | `basicAuthPassword` |
| `logscaleOAuthForward` | Forward OAuth Identity | LogScale | `accessToken` (empty) |
| `ngsiemOAuth2` | OAuth2 client credentials | NGSIEM | `oauth2ClientSecret` |
| `logscaleWithDataLinks` | LogScale token + data links + incremental querying | LogScale | `accessToken` |

## `LoadConfig` utility (`settings.go`)

`LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error)`
runs the full three-phase load flow on a datasource instance's settings and returns a
fully-defaulted, validated `Config`:

1. **Parse** — copy root fields (`URL`, `BasicAuthEnabled`, `BasicAuthUser`) from
   `backend.DataSourceInstanceSettings`, unmarshal `settings.JSONData` into `Config`, and copy
   decrypted secrets into `DecryptedSecureJSONData` by known key name.
2. **`ApplyDefaults`** — set `Mode` to `LogScale` when zero-valued (mirrors
   `ConfigEditor.tsx:59`), and auto-set `DefaultRepository = "search-all"` when `Mode == NGSIEM`
   and no repository is stored (mirrors `ConfigEditor.tsx:127-137,193-200`).
3. **`Validate`** — enforce the runtime contract: URL required; mode is `LogScale` /
   `NGSIEM` / empty; at most one auth flag enabled at a time (mirrors `clearAuthSettings()`);
   each selected auth method has its required inputs (token / OAuth2 client credentials / Basic
   auth user + password); NGSIEM mode requires OAuth2; timeout non-negative; per-item data-link
   `field` and `matcherRegex` non-empty. Errors are joined so every problem surfaces at once.

Everything is logged via `backend.Logger.FromContext(ctx)` with `datasource_uid`,
`datasource_name`, and `plugin` labels so log lines carry request context.

### Direct access to individual phases

`(*Config).ApplyDefaults()` and `(Config).Validate() error` are exported so callers that assemble
a `Config` outside `LoadConfig` (provisioning preview, tests that need to distinguish parse-level
from policy-level errors, schema-example round-trip tools) can invoke them individually.

## Upstream findings

Potential bugs, misleading UX, and consistency issues discovered while researching upstream. All
preserved verbatim in the schema — the schema records what the plugin **does**, not what it
**should** do.

1. **`baseURL` vs `baseUrl` case mismatch.** `pkg/plugin/settings.go:12` declares
   `BaseURL string \`json:"baseURL"\`` (uppercase URL), while the frontend writes
   `jsonData.baseUrl` (lowercase u) at `src/components/ConfigEditor/ConfigEditor.tsx:155`. Go's
   `encoding/json` is case-insensitive on unmarshal, so the value round-trips, but any consumer
   that reads the raw JSON verbatim (e.g. a k8s client) will see two different key spellings
   depending on the code path. This entry follows the frontend's `baseUrl` spelling since that is
   the key persisted to storage; the Go `Config.BaseURL` field also uses `json:"baseUrl"`.
2. **`baseUrl` is dead weight for the backend.** Regardless of what the editor writes, the
   backend overwrites `settings.BaseURL = config.URL` at `pkg/plugin/settings.go:39` before use.
   Provisioning a datasource with `jsonData.baseUrl` set to something other than `url` has no
   effect.
3. **`basicAuthUser` declared on `LogScaleOptions` is misleading.** The frontend type
   `LogScaleOptions` at `src/types.ts:18` declares `basicAuthUser?: string` as a jsonData field,
   but the editor never writes it to jsonData — @grafana/plugin-ui's `Auth` component writes it as
   a root field, and the backend reads `config.BasicAuthUser` (root). Treat any
   `jsonData.basicAuthUser` value as noise: not written by the editor, not read by the backend.
   The schema does not include it.
4. **NGSIEM mode requires exactly one auth method but the editor's mode switch clears everything
   without setting one.** `ConfigEditor.tsx:186-215` switches to NGSIEM mode by calling
   `clearAuthSettings()` (which resets all four auth flags) and then setting `setAuthSelected`
   to `custom-oauth-client-secret` — but does not set `jsonData.oauth2 = true`. The user must
   still fill in the Client ID + Secret fields (which each write the flag via
   `OAuth2Component.tsx:19,34`) before the datasource is usable. A minimally-configured NGSIEM
   datasource can be saved with no auth flags true; the backend health check will then fail with
   an OAuth2-specific error message.
5. **`clearAuthSettings` is called on every mode change AND on every auth-method change.**
   `ConfigEditor.tsx:61-68` returns a fresh reset object; mode switches (`onSelectedMode`,
   `ConfigEditor.tsx:186-215`) and auth-method switches (`onAuthMethodSelect`,
   `ConfigEditor.tsx:251-272`) both spread it. Any concurrent draft state in the auth-specific
   subcomponents (e.g. a typed-but-unsaved Client ID) is silently dropped.
6. **Health check derives repositories via GraphQL in LogScale mode but hardcodes them in NGSIEM
   mode.** `pkg/plugin/healthcheck.go:29,40` — the LogScale branch counts repositories via
   `GetAllRepoNames()`; the NGSIEM branch just verifies OAuth2 exchange. NGSIEM repositories are
   the fixed set at `src/types.ts:52` (`['search-all', 'investigate_view', 'third-party']`).
7. **`Timeout uint` in the upstream Settings is commented out.** `pkg/plugin/settings.go:21` has
   `//Timeout uint \`json:"timeout,omitempty"\`` (disabled). The `jsonData.timeout` field is
   still written by @grafana/plugin-ui's `AdvancedHttpSettings` and consumed by the SDK's
   `HTTPClientOptions(ctx)` — it just doesn't land on the plugin's own Settings struct.
8. **NGSIEM mode's `/humio` URL suffix is applied unconditionally.** `pkg/plugin/plugin.go:54-56`
   appends `/humio` to `settings.BaseURL` regardless of what the URL already ends with. A URL
   that already includes `/humio` results in `.../humio/humio`, which the upstream server will
   reject. The editor does not validate against this.
9. **Editor `disabled` computation misses the OAuth2 client-id-without-secret case.**
   `ConfigEditor.tsx:119-124` — the Default Repository is enabled only when the token / OAuth2 /
   basic-auth requirements are met, but the OAuth2 condition uses
   `secureJsonFields?.oauth2ClientSecret`. Between typing a Client Secret and pressing Save, the
   secret is set but `secureJsonFields` still reports it as unconfigured, so Default Repository
   stays disabled until the datasource is saved.
10. **`docURL` fallback**: `src/plugin.json` only declares a `Website` link
    (`https://github.com/grafana/falconlogscale-datasource`), not a dedicated docs URL. The schema
    uses that same GitHub URL as the plugin's `docURL`.

## Validation performed

- `dsconfig.ParseAndResolveSchemaJSON` + `Schema.Validate()` (Go validator in this repo) — passes.
- JSON Schema validation against [`dsconfig/schema.json`](../../dsconfig/schema.json) — passes
  (via `ajv-cli --strict=false`).
- `go build ./... && go vet ./... && gofmt -l . && go test ./...` inside `registry/` — clean.
- `tsc --noEmit --strict` on `settings.ts` (TypeScript 5.9.3) — clean.
- Conformance suite: schema round-trip, artifact drift, spec/secure separation,
  jsonData/struct-tag parity in both directions, secure-key parity — all passing.
