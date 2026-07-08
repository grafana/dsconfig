# grafana-jira-datasource

Declarative configuration schema for the Jira datasource plugin (`grafana-jira-datasource`).

## Upstream researched

- **Repo**: `github.com/grafana/plugins-private` (monorepo)
- **Ref**: `main`
- **Commit SHA**: `267f4937806ed6404b6628d13ae358a5d308e376`
- **Plugin path**: `plugins/grafana-jira-datasource/`
- **Backend Go module**: `github.com/grafana/jira-datasource` (`pkg/`)

Every value in [`dsconfig.json`](dsconfig.json) — labels, placeholders, tooltips, option
labels/values, section titles, defaults, validations, dependency / required-when / disabled-when
expressions, storage keys, storage targets, value types, group titles, and instructions — is
traceable to a specific `file:line` in the upstream plugin at this SHA. See
[Field provenance](#field-provenance) below.

To reproduce this research (the plugin lives inside the `plugins-private` monorepo; `@grafana/*`
frontend deps use `catalog:` and resolve from the monorepo `.yarnrc.yml` catalog):

```bash
git clone https://github.com/grafana/plugins-private
cd plugins-private
git checkout 267f4937806ed6404b6628d13ae358a5d308e376
cd plugins/grafana-jira-datasource
```

If upstream `main` has advanced past this SHA, re-diff the sources listed under
[Sources researched](#sources-researched) before merging any changes to this entry.

## Files

| File | Purpose |
| --- | --- |
| [`dsconfig.json`](dsconfig.json) | dsconfig v1 schema — single source of truth for all config fields, groups, relationships, and instructions |
| [`settings.ts`](settings.ts) | TypeScript models: `RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig` |
| [`settings.go`](settings.go) | Go `Config` model (flat: jsonData fields + `DecryptedSecureJSONData`), `PluginID`, `AuthMethod` / `Hosting` / `SecureJsonDataKey` typed constants, and the `LoadConfig` utility |
| [`schema.go`](schema.go) | k8s-style SDK plugin schema: embeds `dsconfig.json`, converts it via `dsconfig.NewSDKSchema`, and defines `SettingsExamples` for each auth/connection variant |
| [`settings_test.go`](settings_test.go) | Table tests for `LoadConfig`, `ApplyDefaults`, and `Validate` |
| [`conformance_test.go`](conformance_test.go) | Runs `schema.RunPluginTests` — the shared dsconfig conformance suite — against `dsconfig.json` |
| [`schema.gen.json`](schema.gen.json), [`settings.gen.json`](settings.gen.json), [`settings.examples.gen.json`](settings.examples.gen.json) | Committed schema artifacts (regenerate with `go generate ./...` inside this directory; the `SchemaArtifactInSync` conformance subtest guards drift) |

There is no per-entry `go.mod` — every registry entry is a subpackage of the shared
[`registry/`](..) module (`github.com/grafana/dsconfig/registry`). Package name: `jiradatasource`.

## Sources researched

Every source below was read at the pinned upstream SHA
(`267f4937806ed6404b6628d13ae358a5d308e376`), plus external editor components at the exact
versions the plugin's `package.json` / monorepo `.yarnrc.yml` catalog pins.

### Plugin (`plugins/grafana-jira-datasource` @ `267f4937`)

| File | What was read |
| --- | --- |
| `src/plugin.json:3-5,28` | `type` (`datasource`), `pluginType` (`id` = `grafana-jira-datasource`), `pluginName` (`name` = `Jira`), `docURL` (`info.links[0].url`) |
| `src/components/ConfigEditor.tsx:23-26` | `types` array — Provider radio option labels (`Jira Cloud`, `Jira Data Center / Jira Server`) mapped to `Provider.CLOUD`/`Provider.SERVER` |
| `src/components/ConfigEditor.tsx:158-162` | Load-time defaults: `hosting` → `cloud`, `scopedToken` → `false`, `authMethod` → `basicAuth` |
| `src/components/ConfigEditor.tsx:167-171` | `DataSourceDescription` (`hasRequiredFields`) intro |
| `src/components/ConfigEditor.tsx:175-231` | `<ConfigSection title="Connection">`: Provider radio → `jsonData.hosting` (label, tooltip, `disabled` when `oauth2`), URL input → `jsonData.url` (label, tooltip, placeholder, `required`) |
| `src/components/ConfigEditor.tsx:88-98` | `onAuthMethodChange` — selecting `oauth2` forces `jsonData.hosting = 'cloud'` |
| `src/components/ConfigEditor.tsx:235-245` | `<Auth visibleMethods={['custom-jira']} ...>` custom method `id: 'custom-jira'`, `label: 'Jira authentication'`, `description: 'Provide information to grant access to the data source.'` |
| `src/components/ConfigEditor.tsx:248-262` | Authentication method radio → `jsonData.authMethod` (label, tooltip) |
| `src/components/ConfigEditor.tsx:264-349` | basicAuth block: User email → `jsonData.user`, API Token → `secureJsonData.token` (tooltip + `here` link), Scoped Token `Switch` → `jsonData.scopedToken`, Jira App Cloud Id → `jsonData.cloudId` (rendered only when `scopedToken`) |
| `src/components/ConfigEditor.tsx:351-408` | oauth2 block: Client ID → `jsonData.oauthClientID`, Client Secret → `secureJsonData.oauthClientSecret`, Jira App Cloud Id → `jsonData.cloudId` (always) |
| `src/components/ConfigEditor.tsx:415-461` | Conditional `<ConfigSection title="Additional settings">` → Secure Socks Proxy checkbox `jsonData.enableSecureSocksProxy` — deliberately excluded from this entry |
| `src/components/selectors.ts:3-33` | E2E `inputId`/`aria-label` map (labels/placeholders themselves are hard-coded in the editor) |
| `src/types.ts:3-8` | `JiraAuthMethod` union + `authMethodOptions` (radio labels/values) |
| `src/types.ts:31-48` | `JiraOptions` (jsonData) and `JiraSecureOptions` (secureJsonData: `token`, `oauthClientSecret`) |
| `src/types.ts:57-60` | `Provider` enum (`cloud`/`server`) written to `jsonData.hosting` |
| `pkg/models/settings.go:14-28` | `Settings` struct + json tags (`url`, `user`, `hosting`, `scopedToken`, `cloudId`, `authMethod`, `oauthClientID`) plus non-tagged `Token`, `OAuthClientSecret`, `HttpClientOptions` |
| `pkg/models/settings.go:31-82` | `LoadSettings`: json parse, `GetAuthMethod` resolution, URL-empty check, `https://` scheme prepend, per-auth-method secret copy + required-field checks, `HTTPClientOptions(ctx)` |
| `pkg/models/auth_method.go:3-16` | `AuthMethod` alias, `AuthMethodBasicAuth`/`AuthMethodOAuth2`, `GetAuthMethod` (defaults to basicAuth) |
| `pkg/models/settings_test.go:12-159` | `LoadSettings` behavior table (legacy no-authMethod, explicit basicAuth, oauth2, each missing-field error, missing-URL error) — mirrored by `settings_test.go` here |
| `pkg/plugin.go:171-228` | `getInstance` / `getEndpoint`: `hosting=='server'` → REST API v2 else v3; base URL is `jsonData.url` for non-scoped basic auth, `https://api.atlassian.com/ex/jira/<cloudId>` for scoped tokens and oauth2 |
| `pkg/plugin.go:230-267` | `getHttpClient`: oauth2 → client-credentials `authclient`; else `BasicAuthTransport` (user+token) or, when `user==""`, `BearerAuthTransport` (token only) |
| `go.mod:1,9` | Backend module `github.com/grafana/jira-datasource`; `grafana-plugin-sdk-go v0.292.0` |

### External editor components

Read at the exact versions pinned in the plugin's `package.json` (all `@grafana/*` via
`catalog:`, resolved from the monorepo `.yarnrc.yml` catalog).

| Component(s) | Version | Source consulted | What was read |
| --- | --- | --- | --- |
| `Auth`, `ConfigSection`, `ConfigSubSection`, `DataSourceDescription`, `convertLegacyAuthProps` | `@grafana/plugin-ui@^0.13.1` (`.yarnrc.yml:22`) | `github.com/grafana/plugin-ui`, `src/components/ConfigEditor/Auth/Auth.tsx`, `.../auth-method/AuthMethodSettings.tsx` | `Auth` renders `<ConfigSection title="Authentication">`; with a single `visibleMethods` entry it renders a `<ConfigSubSection>` titled with the custom method label (`Jira authentication`) and its description; `TLS` prop (from `convertLegacyAuthProps`) renders the standard TLS settings subsection |
| `RadioButtonGroup`, `Input`, `SecretInput`, `Switch`, `InlineField`, `InlineLabel`, `Checkbox`, `Tooltip`, `Icon`, `FieldValidationMessage` | `@grafana/ui@^11.6.7` (`.yarnrc.yml:26`) | `github.com/grafana/grafana`, `packages/grafana-ui/src/components/` | Prop names (`label`, `placeholder`, `value`, `onChange`, `isConfigured`, `onReset`) so we knew which UI attributes to record; `Switch` → boolean, `SecretInput` → write-only secret |
| `DataSourceJsonData`, `DataSourcePluginOptionsEditorProps`, `SelectableValue`, `FeatureToggles` | `@grafana/data@^11.6.7` (`.yarnrc.yml:19`) | `github.com/grafana/grafana`, `packages/grafana-data/src/` | Base `DataSourceJsonData` interface that `JiraOptions` extends; editor prop shape |
| `config`, `reportInteraction` | `@grafana/runtime@^11.6.7` (`.yarnrc.yml:23`) | `github.com/grafana/grafana`, `packages/grafana-runtime/src/` | `config.featureToggles`/`buildInfo.version` gate the Secure Socks Proxy section; `reportInteraction('grafana_jira_provider_type_clicked')` fires on provider change |

## Field provenance

Every dsconfig field, traced from its schema `id` to the upstream `file:line` where each of its
label, placeholder, tooltip, default, storage key, and value type is defined.

| Schema `id` | Storage key | Target | Label source | Placeholder / options / default source | Value type source | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `jsonData_hosting` | `hosting` | `jsonData` | `ConfigEditor.tsx:190` (`InlineLabel` `Provider`) | Options `ConfigEditor.tsx:24-25` + `types.ts:57-60`; default `cloud` `ConfigEditor.tsx:160` | `Settings.Hosting string` `settings.go:17`; TS `hosting: string` `types.ts:34` | Tooltip `ConfigEditor.tsx:182-186`; `disabledWhen` `ConfigEditor.tsx:200`; forced to `cloud` on oauth2 `ConfigEditor.tsx:94-96` |
| `jsonData_url` | `url` | `jsonData` | `ConfigEditor.tsx:204` (`InlineField label="URL"`) | Placeholder `ConfigEditor.tsx:222` (`"URL"`) | `Settings.URL string` `settings.go:15`; TS `url?: string` `types.ts:32` | Tooltip `ConfigEditor.tsx:208-211`; `required` `ConfigEditor.tsx:215`; role `endpoint.baseUrl`; stored in jsonData, **not** the datasource root url |
| `jsonData_authMethod` | `authMethod` | `jsonData` | `ConfigEditor.tsx:249` (`InlineField label="Authentication method"`) | Options `types.ts:5-8`; default `basicAuth` `ConfigEditor.tsx:162` + `auth_method.go:11-15` | `Settings.AuthMethod` (`AuthMethod` = string) `settings.go:20`, `auth_method.go:3`; TS `JiraAuthMethod` `types.ts:3` | Tooltip `ConfigEditor.tsx:251`; role `auth.discriminator` |
| `jsonData_user` | `user` | `jsonData` | `ConfigEditor.tsx:267` (`InlineField label="User email"`) | Placeholder `ConfigEditor.tsx:278` (`"User email"`) | `Settings.User string` `settings.go:16`; TS `user?: string` `types.ts:33` | Tooltip `ConfigEditor.tsx:269`; `dependsOn` basicAuth `ConfigEditor.tsx:264`; role `auth.basic.username`; empty `user` → Bearer transport (`plugin.go:253-264`) |
| `secureJsonData_token` | `token` | `secureJsonData` | `ConfigEditor.tsx:286` (`InlineField label="API Token"`) | Placeholder `ConfigEditor.tsx:311` (`"API Token"`) | `Settings.Token string` `settings.go:24`; TS `JiraSecureOptions.token` `types.ts:46` | Tooltip `ConfigEditor.tsx:288-300`; `docURL` `ConfigEditor.tsx:293`; `dependsOn`/`requiredWhen` basicAuth (`settings.go:69-71`); role `auth.basic.password` |
| `jsonData_scopedToken` | `scopedToken` | `jsonData` | `ConfigEditor.tsx:319` (`InlineField label="Scoped Token"`) | Default `false` `ConfigEditor.tsx:161` | `Settings.ScopedToken bool` `settings.go:18`; TS `scopedToken?: boolean` `types.ts:35` | Tooltip `ConfigEditor.tsx:321`; `Switch` component `ConfigEditor.tsx:326`; `dependsOn` basicAuth `ConfigEditor.tsx:264` |
| `jsonData_cloudId` | `cloudId` | `jsonData` | `ConfigEditor.tsx:331`,`391` (`InlineField label="Jira App Cloud Id"`) | Placeholder `ConfigEditor.tsx:342`,`402` (`"Jira App Cloud Id"`) | `Settings.CloudId string` `settings.go:19`; TS `cloudId?: string` `types.ts:36` | basicAuth tooltip `ConfigEditor.tsx:333`; oauth2 tooltip (`override`) `ConfigEditor.tsx:393`; shown for scoped basic auth `:329` and oauth2 `:390`; required oauth2 (`settings.go:62-64`) / scoped (`plugin.go:221-224`) |
| `jsonData_oauthClientID` | `oauthClientID` | `jsonData` | `ConfigEditor.tsx:354` (`InlineField label="Client ID"`) | Placeholder `ConfigEditor.tsx:365` (`"Client ID"`) | `Settings.OAuthClientID string` `settings.go:21`; TS `oauthClientID?: string` `types.ts:39` | Tooltip `ConfigEditor.tsx:356`; `dependsOn`/`requiredWhen` oauth2 (`settings.go:56-58`); role `auth.oauth2.clientId` |
| `secureJsonData_oauthClientSecret` | `oauthClientSecret` | `secureJsonData` | `ConfigEditor.tsx:373` (`InlineField label="Client Secret"`) | Placeholder `ConfigEditor.tsx:383` (`"Client Secret"`) | `Settings.OAuthClientSecret string` `settings.go:25`; TS `oauthClientSecret?: string` `types.ts:47` | Tooltip `ConfigEditor.tsx:374`; `dependsOn`/`requiredWhen` oauth2 (`settings.go:59-61`); role `auth.oauth2.clientSecret` |

## Field inventory summary

| Schema field | Storage key | Target | Editor label | Read by backend? |
| --- | --- | --- | --- | --- |
| `jsonData_hosting` | `hosting` | `jsonData` | Provider | Yes (REST API version + endpoint fallback, `plugin.go:178,198,308`) |
| `jsonData_url` | `url` | `jsonData` | URL | Yes (base URL for non-scoped basic auth, `plugin.go:227`) |
| `jsonData_authMethod` | `authMethod` | `jsonData` | Authentication method | Yes (`settings.go:37,52`) |
| `jsonData_user` | `user` | `jsonData` | User email | Yes (Basic username vs Bearer switch, `plugin.go:253-264`) |
| `secureJsonData_token` | `token` | `secureJsonData` | API Token | Yes (`settings.go:67`; Basic password / Bearer token) |
| `jsonData_scopedToken` | `scopedToken` | `jsonData` | Scoped Token | Yes (endpoint routing, `plugin.go:221`) |
| `jsonData_cloudId` | `cloudId` | `jsonData` | Jira App Cloud Id | Yes (`api.atlassian.com` gateway, `plugin.go:219,225`) |
| `jsonData_oauthClientID` | `oauthClientID` | `jsonData` | Client ID | Yes (`plugin.go:240`) |
| `secureJsonData_oauthClientSecret` | `oauthClientSecret` | `secureJsonData` | Client Secret | Yes (`plugin.go:241`) |

### Frontend-only settings

None. Every editor-visible field is read by the backend. (The only field written by the editor
that the plugin's own Go code never reads by name is `jsonData.enableSecureSocksProxy`, which is
the excluded Secure Socks Proxy field — see [Modeling decisions](#modeling-decisions).)

### Backend-only settings

None. Every field in the backend `Settings` struct (`pkg/models/settings.go:14-28`) has a
corresponding editor control. `Settings.HttpClientOptions` (`settings.go:27`) is not configuration
storage — it is populated by `config.HTTPClientOptions(ctx)` (`settings.go:75`) from the standard
datasource transport settings — so it is not a schema field.

## Where the types are defined

Only config type/field definitions are listed. UI components (`Auth`, `ConfigSection`,
`DataSourceDescription`, …) and functions/helpers (`LoadSettings`, `GetAuthMethod`,
`convertLegacyAuthProps`, …) are omitted even where they are the reason a field exists.

### Frontend (TypeScript)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `JiraOptions` (jsonData: `url`, `user`, `hosting`, `scopedToken`, `cloudId`, `authMethod`, `oauthClientID`, `enableSecureSocksProxy`), `JiraSecureOptions` (`token`, `oauthClientSecret`), `JiraAuthMethod`, `Provider` | `src/types.ts:3-60` | plugin (`grafana-jira-datasource`) |
| `DataSourceJsonData` (base interface `JiraOptions` extends) | `packages/grafana-data/src/types/datasource.ts` | `@grafana/data` `^11.6.7` |

### Backend (Go)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `Settings` (jsonData fields + decrypted `Token`/`OAuthClientSecret` + `HttpClientOptions`) | `pkg/models/settings.go:14-28` | plugin (`github.com/grafana/jira-datasource`) |
| `AuthMethod` (`AuthMethodBasicAuth`, `AuthMethodOAuth2`) | `pkg/models/auth_method.go:3-8` | plugin (`github.com/grafana/jira-datasource`) |
| `backend.DataSourceInstanceSettings` (carries `JSONData`, `DecryptedSecureJSONData`, and root fields like `URL`, `BasicAuth`, `User` — all unused by this plugin) | `backend/common.go` | `github.com/grafana/grafana-plugin-sdk-go` `v0.292.0` (plugin) / `v0.292.1` (registry) |
| `httpclient.Options` (transport/TLS/proxy options carried on `Settings.HttpClientOptions`) | `backend/httpclient/` | `github.com/grafana/grafana-plugin-sdk-go` `v0.292.0` (plugin) / `v0.292.1` (registry) |
| Provider/hosting has **no backend enum** — the backend types `hosting` as a plain `string` (`settings.go:17`) and compares it to the literal `"server"` (`plugin.go:178`) | — | — |

The models in this entry flatten that spread into a single Go `Config` type (jsonData fields +
`DecryptedSecureJSONData`) plus `SecureJsonDataKey`, `AuthMethod`, and `Hosting` typed constant
lists. `AuthMethod` mirrors the plugin's own `auth_method.go` constants verbatim; `Hosting`
constants are derived from the frontend `Provider` enum since the backend defines none.
`settings.ts` keeps the three canonical TypeScript types (`RootConfig`, `JsonDataConfig`,
`SecureJsonDataConfig`).

## Modeling decisions

- **`RootConfig` is a blank object.** The plugin stores `url` and `user` in `jsonData` (not the
  datasource root) and never reads any root-level datasource field: `pkg/plugin.go:171-267` builds
  the client from `jsonData` + decrypted secrets only. So `RootConfig = Record<string, never>` and
  the Go `Config` carries no `json:"-"` root fields.
- **Single `authMethod` discriminator.** `jsonData.authMethod` (`auth.discriminator`) selects
  `basicAuth` vs `oauth2`; `dependsOn` on each credential field mirrors the editor's conditional
  render (`ConfigEditor.tsx:264,351`) and `requiredWhen` mirrors the backend's per-method checks
  (`settings.go:52-72`).
- **Shared `cloudId` field with an override.** The editor renders two separate "Jira App Cloud Id"
  inputs that write the same `jsonData.cloudId` key — one in the basic-auth block gated by
  `scopedToken` (`ConfigEditor.tsx:329-347`) and one always shown in the oauth2 block
  (`:390-406`). They are modeled as one field whose `dependsOn` covers both cases and whose
  `overrides[0]` swaps the tooltip to the oauth2 wording (`:393`) when `authMethod == 'oauth2'`.
- **`hosting` disabled and forced under oauth2.** OAuth 2.0 is Jira Cloud only: `onAuthMethodChange`
  forces `jsonData.hosting = 'cloud'` (`ConfigEditor.tsx:94-96`) and the Provider radio is disabled
  (`:200`). Captured as `disabledWhen: "jsonData_authMethod == 'oauth2'"` plus a `pair` relationship
  and an instruction. (The forced-write side effect is documented rather than encoded as an
  `effect`, since `authMethod` is a real storage radio, not a virtual selector.)
- **`API Token` role.** `secureJsonData.token` is labeled "API Token" and used as the HTTP Basic
  password when `user` is set (`plugin.go:259-263`), so it carries role `auth.basic.password`. When
  `user` is empty the same token is sent as a Bearer token (`plugin.go:253-258`) — a dual use noted
  in the instructions rather than split into two fields (there is one storage key).
- **Authentication group title.** The group is titled **Authentication**, matching the
  `<ConfigSection title="Authentication">` the `@grafana/plugin-ui` `Auth` component renders. The
  editor additionally nests a `ConfigSubSection` titled **Jira authentication** (the custom method
  label, `ConfigEditor.tsx:244`) with description "Provide information to grant access to the data
  source." (`:245`); dsconfig groups model top-level sections, so the subsection label is recorded
  here rather than as the group title.
- **Standard TLS settings out of scope.** `convertLegacyAuthProps` (`ConfigEditor.tsx:37-40`) feeds
  a `TLS` prop into `Auth` (`:240`), so the editor also renders the standard TLS-settings
  subsection (self-signed cert / TLS client auth / skip TLS verify), backed by the standard
  datasource `jsonData` (`tlsAuth`, `tlsSkipVerify`, `serverName`, …) and `secureJsonData`
  (`tlsCACert`, `tlsClientCert`, `tlsClientKey`) keys and consumed transparently by
  `config.HTTPClientOptions(ctx)` (`settings.go:75`). These are generic SDK transport fields (the
  `plugin_sdk_settings` pack territory), not Jira-specific config, so — consistent with the
  gold-standard `grafana-github-datasource` entry — they are not enumerated here.
- **Secure Socks Proxy excluded.** The conditional `Additional settings` section
  (`ConfigEditor.tsx:415-461`) writes `jsonData.enableSecureSocksProxy`; this field is deliberately
  omitted per the repo AGENTS.md exclusion.
- **Field ID naming convention.** IDs are prefixed with their storage target
  (`jsonData_`, `secureJsonData_`) followed by the camelCase storage key; the `key` property keeps
  the plugin's raw storage key (`id` is the schema reference, `key` is the storage contract).
- **Flat `Config` in Go.** `settings.go` mirrors the jsonData portion of the upstream `Settings`
  (`pkg/models/settings.go:14-21`) verbatim (fields + json tags) and models the two write-only
  secrets in `DecryptedSecureJSONData map[SecureJsonDataKey]string`. It does not carry
  `HttpClientOptions` (not config storage) or root datasource fields (unused by the backend).
- **`SecureJsonDataConfig` is a key list.** Secure values are write-only, so the secure type is the
  array of secret key names (`token`, `oauthClientSecret`); consumers read `secureJsonFields` to see
  what is configured.

## SDK plugin schema and k8s-style examples (`schema.go`)

`NewSchema()` assembles the `grafana-plugin-sdk-go` `pluginschema.PluginSchema` bundle from the
embedded `dsconfig.json`: the `jsonData` fields become the OpenAPI settings `spec`, secure fields
become `secureValues`, and Grafana's datasource API server serves the bundle as `{apiVersion}.json`.

`SettingsExamples()` provides the default configuration plus one k8s-style example per
authentication method and connection variant. Each example is a full instance-settings object with
the plugin configuration under `jsonData` and the relevant write-only secrets under
`secureJsonData`. **All secret values are obviously-fake angle-bracket placeholders**
(e.g. `<your-jira-api-token>`) — no realistic token shapes — and the default example (`""`) carries
an empty token to show what must be filled in:

| Example | `authMethod` | `hosting` | Notes | `secureJsonData` |
| --- | --- | --- | --- | --- |
| `""` (default) | `basicAuth` | `cloud` | schema defaults; needs `url` + `token` filled in | `token` (empty) |
| `basicAuthCloud` | `basicAuth` | `cloud` | Atlassian Cloud, email + API token (HTTP Basic) | `token` |
| `basicAuthServer` | `basicAuth` | `server` | Jira Data Center / Server (REST API v2), email + token | `token` |
| `bearerTokenServer` | `basicAuth` | `server` | Jira Data Center PAT: empty `user` → Bearer token | `token` |
| `basicAuthScopedToken` | `basicAuth` | `cloud` | scoped token → `api.atlassian.com/ex/jira/<cloudId>`; `cloudId` required | `token` |
| `oauth2` | `oauth2` | `cloud` | client-credentials grant; `oauthClientID` + `cloudId` required; Cloud only | `oauthClientSecret` |
| `legacyBasicAuthNoAuthMethod` | *(absent)* | *(absent)* | pre-`authMethod` instance; backend resolves to `basicAuth` | `token` |

## `LoadConfig` utility (`settings.go`)

`LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error)`
runs the full three-phase load flow on a datasource instance's settings and returns a
fully-defaulted, validated `Config`:

1. **Parse** — `json.Unmarshal(settings.JSONData, &cfg)` (empty `JSONData` is a parse error, matching
   upstream `pkg/models/settings.go:33`), then copy decrypted secrets into
   `DecryptedSecureJSONData` by known key.
2. **`ApplyDefaults`** — normalize the two fields the plugin's own `LoadSettings` normalizes on
   every load: `AuthMethod` via `ResolveAuthMethod` (empty/unknown → `basicAuth`, mirroring
   `GetAuthMethod` at `auth_method.go:11-15`) and prepend `https://` to a non-empty scheme-less
   `URL` (`settings.go:43-45`). The URL prefix is guarded on non-empty so an unset URL stays empty
   for `Validate` to reject (upstream checks URL-empty before normalizing).
3. **`Validate`** — enforce the runtime contract (`settings.go:39-72` + the `getEndpoint` contract
   in `plugin.go:214-228`): `url` non-empty; oauth2 requires `oauthClientID` +
   `oauthClientSecret` + `cloudId`; basicAuth (and any non-oauth2 value, matching the upstream
   switch default) requires `token`, and a scoped token additionally requires `cloudId`. Errors are
   joined so every problem surfaces at once (upstream returns only the first).

Everything is logged via `backend.Logger.FromContext(ctx)` with `datasource_uid`,
`datasource_name`, and `plugin` labels. This is the intended shape for the plugin's own upstream
`LoadSettings` to sync to.

### Direct access to individual phases

`(*Config).ApplyDefaults()` and `(Config).Validate() error` are exported for callers that compose
them themselves (provisioning preview, schema-example round-trip, tests distinguishing parse-level
from policy-level errors); assemble a `Config` directly and call them without `LoadConfig`.

## Upstream findings

Potential bugs, misleading UX, and consistency issues discovered while researching upstream. All
preserved verbatim in the schema — the schema records what the plugin **does**, not what it
**should** do.

1. **`url` is required even when it is ignored.** `LoadSettings` rejects an empty `url`
   (`settings.go:39-41`), but for scoped tokens and OAuth 2.0 the API base URL is
   `https://api.atlassian.com/ex/jira/<cloudId>` and `settings.URL` is never used as the endpoint
   (`plugin.go:214-226`). A scoped-token or oauth2 datasource must still supply some non-empty
   `url` that has no effect on where requests go.
2. **Bearer / PAT auth is undocumented in the editor.** `getHttpClient` sends the token as a Bearer
   token whenever `user` is empty (`plugin.go:253-258`) — the intended path for a Jira Data Center
   personal access token — but the editor only labels the field "User email" with no hint that
   leaving it blank changes the auth scheme. Recorded in the instructions and the `bearerTokenServer`
   example.
3. **`hosting` value can be stale after switching to OAuth 2.0 and back.** Selecting oauth2 forces
   `hosting='cloud'` (`ConfigEditor.tsx:94-96`), but switching back to basicAuth leaves it at
   `cloud`; a Data Center user must re-select the provider. The Provider radio is also disabled
   while oauth2 is selected (`:200`), so the forced value cannot be seen changing.
4. **Leftover debug `console.log` in `onSettingChange`.** `ConfigEditor.tsx:55` logs every
   non-secret jsonData change (setting name, value, old/new options) to the browser console — a
   development artifact shipped in the editor. No effect on stored config.
5. **`GetAuthMethod` silently swallows unknown auth methods.** `auth_method.go:11-15` maps any value
   other than the exact string `"oauth2"` (including typos and `"basicAuth"` casing variants) to
   `basicAuth`. A misconfigured `authMethod` therefore never errors — it just authenticates as basic
   auth and fails later if no token is present. Mirrored by `ResolveAuthMethod` and the
   "unknown authMethod resolves to basic auth" test.
6. **Editor requires only URL; backend requires credentials.** The editor marks only URL as
   `required` (`ConfigEditor.tsx:215`) and shows an inline "Please enter a URL." message
   (`:226-230`); the per-auth-method credential requirements are enforced only at backend load
   (`settings.go:52-72`). `requiredWhen` encodes that backend contract even though the editor shows
   no required markers on those fields.
7. **`OAuth 2.0 — Service Account` label uses an em dash.** `types.ts:7` — the option label contains
   a Unicode em dash (`—`), preserved verbatim in the `authMethod` option label and the `oauth2`
   example summary.

## Validation performed

- `dsconfig.ParseAndResolveSchemaJSON` + `Schema.Validate()` (Go validator in this repo) — passes
  (run by the conformance suite).
- JSON Schema validation against [`dsconfig/schema.json`](../../dsconfig/schema.json) (draft-07,
  strict `additionalProperties: false`) — passes.
- `go generate ./...` in this directory, then `gofmt -l .`, `go vet ./...`, `go test ./...` in
  [`registry/`](..) — all clean/passing (schema round-trip, spec/secure separation, jsonData/struct
  parity both directions, secure-key parity, `SchemaArtifactInSync` drift guard, and `LoadConfig` /
  `ApplyDefaults` / `Validate` table tests including each auth method, the legacy fallback, URL
  normalization, and malformed input).
- `settings.ts`: `tsc --noEmit --strict` — clean.
- The pre-existing `dsconfig` and `schema` workspace modules — still build.
