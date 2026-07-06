# grafana-servicenow-datasource

Declarative configuration schema for the [ServiceNow datasource plugin](https://grafana.com/docs/plugins/grafana-servicenow-datasource) (`grafana-servicenow-datasource`).

## Upstream researched

- **Repo**: `github.com/grafana/plugins-private` (monorepo)
- **Plugin path**: `plugins/grafana-servicenow-datasource`
- **Ref**: `main`
- **Commit SHA**: `267f4937806ed6404b6628d13ae358a5d308e376`

Every value in [`dsconfig.json`](dsconfig.json) — labels, placeholders, tooltips (surfaced as
field `description`s), option labels/values, section titles, help markdown, defaults, validations,
dependency and required-when expressions, storage keys, storage targets, value types, group titles,
and instructions — is traceable to a specific `file:line` in the upstream plugin at this SHA. See
[Field provenance](#field-provenance) below.

To reproduce this research (paths relative to the monorepo root):

```bash
git -C <plugins-private> fetch origin && git -C <plugins-private> checkout 267f4937806ed6404b6628d13ae358a5d308e376
# then read plugins/grafana-servicenow-datasource/...
```

If upstream `main` has advanced past this SHA, re-diff the sources listed under [Sources
researched](#sources-researched) before merging any changes to this entry.

## Files

| File | Purpose |
| --- | --- |
| [`dsconfig.json`](dsconfig.json) | dsconfig v1 schema — single source of truth for all config fields, groups, relationships, and instructions |
| [`settings.ts`](settings.ts) | TypeScript models: `RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig` |
| [`settings.go`](settings.go) | Go `Config` model (root fields + jsonData fields + `DecryptedSecureJSONData`), `PluginID`, `AuthMethod`/`SecureJsonDataKey` typed constants, `GetAuthMethod`, and the `LoadConfig`/`ApplyDefaults`/`Validate` utilities |
| [`schema.go`](schema.go) | k8s-style SDK plugin schema: embeds `dsconfig.json`, converts it via `dsconfig.NewSDKSchema`, and defines `SettingsExamples` for each auth/legacy variant |
| [`settings_test.go`](settings_test.go) | Table tests for `LoadConfig`, `ApplyDefaults`, `Validate`, and the settings examples |
| [`conformance_test.go`](conformance_test.go) | Runs `schema.RunPluginTests` — the shared dsconfig conformance suite — against `dsconfig.json` |
| [`schema.gen.json`](schema.gen.json), [`settings.gen.json`](settings.gen.json), [`settings.examples.gen.json`](settings.examples.gen.json) | Committed schema artifacts (regenerate with `go generate ./...` inside this directory; the `SchemaArtifactInSync` conformance subtest guards drift) |

There is no per-entry `go.mod` — every registry entry is a subpackage of the shared
[`registry/`](..) module (`github.com/grafana/dsconfig/registry`). This entry's package is
`servicenowdatasource`.

## Sources researched

Every source below was read at the pinned upstream SHA
(`267f4937806ed6404b6628d13ae358a5d308e376`), plus external editor components at the exact versions
the plugin's `package.json` pins via the monorepo `.yarnrc.yml` `catalog:` block.

### Plugin (`plugins/grafana-servicenow-datasource`)

| File | What was read |
| --- | --- |
| `src/plugin.json:3,4,38` | `pluginName` (`name` = "ServiceNow"), `pluginType` (`id`), `docURL` (`info.links[0].url`) |
| `src/types/index.ts:5-15` | `ServiceNowOptions` (jsonData): `authMethod`, `enableSecureSocksProxy` (excluded), `oauthClientID`, `oauthEnabled` (legacy), `useSysTables`, `queryTimeoutSeconds` |
| `src/types/index.ts:17-22` | `ServiceNowAuthMethod` union and `authOptions` radio option labels/values ("Basic auth"/`basicAuth`, "ServiceNow OAuth"/`serviceNowOAuth`) |
| `src/types/index.ts:24-27` | `ServiceNowSecureOptions`: `basicAuthPassword`, `oauthClientSecret` |
| `src/selectors.ts:28-64` | `Components.ConfigEditor` label/placeholder/tooltip/id map for every editor field (labels and tooltips live here, not inline in the editor) |
| `src/components/ConfigEditor.tsx:25-38` | `initAuthMethod` — the legacy `oauthEnabled` → `serviceNowOAuth` display derivation on load |
| `src/components/ConfigEditor.tsx:47-57` | React state seeds: `url` (`options.url`), `basicAuthUser` (`options.basicAuthUser`), `basicAuthPassword` (`secureJsonData.basicAuthPassword`), `useSysTables`, `queryTimeoutSeconds` (default 30), `oauthClientID`, `oauthClientSecret`, `authMethod` |
| `src/components/ConfigEditor.tsx:66-121` | Change handlers: URL → root `url`, Username → root `basicAuthUser`, Client ID → `jsonData.oauthClientID`, Password/Client Secret → secure keys, `onAuthMethodChange` → `jsonData.authMethod` |
| `src/components/ConfigEditor.tsx:123-129` | `toggleSysTablesSwitch` → `jsonData.useSysTables` |
| `src/components/ConfigEditor.tsx:135` | `ConfigSection` title "ServiceNow Instance Settings" |
| `src/components/ConfigEditor.tsx:136-212` | Field render order and conditional OAuth block (`authMethod === 'serviceNowOAuth'`) for Client ID / Client Secret |
| `src/components/ConfigEditor.tsx:214-216` | "Permissions" row rendering `PermissionsHelp` (modeled as the `help` drawer on `root_basicAuthUser`) |
| `src/components/ConfigEditor.tsx:218-254` | `useSysTables` switch (label "Use Sys Tables?" `:220`, tooltip `:221`) and `queryTimeoutSeconds` input (label "Query Timeout" `:236`, tooltip `:237`, placeholder "30" `:243`, `min={1}` `:242`, editor default 30 `:246`) |
| `src/components/ConfigEditor.tsx:256` | `CustomHeadersSettings` — dynamic `httpHeaderName<N>`/`httpHeaderValue<N>` pairs (not modeled as first-class fields) |
| `src/components/ConfigEditor.tsx:259-299` | Conditional `Secure Socks Proxy` switch writing `jsonData.enableSecureSocksProxy` — deliberately excluded from this entry |
| `src/components/PermissionsHelp.tsx:22-155` | The permissions modal content (ACL table + "Limited permissions") captured verbatim as the `help` markdown of `root_basicAuthUser` |
| `pkg/models/settings.go:18-37` | Backend `Settings` struct (resolved shape: `AuthMethod`, `BasicAuthEnabled`, `BasicAuthPassword`, `BasicAuthUser`, `Headers`, `OAuthClientID`, `OAuthClientSecret`, `URL`, `UseSysTables`, `QueryTimeoutSeconds`) |
| `pkg/models/settings.go:40-105` | `LoadSettings`: anonymous jsonData parse struct (`:41-51`, the authoritative json tags), custom-header extraction, query-timeout default (`:73-77`), root `URL`/`BasicAuthUser` copies (`:82,91,98`), secret copies (`:92,99,101`), `BasicAuthEnabled` derived and provisioned value overwritten (`:87-96`) |
| `pkg/models/settings.go:108-134` | `IsValid`: URL required, and the header-conditioned Basic/OAuth requirement branching |
| `pkg/models/auth_method.go:4-23` | `AuthMethod` alias, `AuthMethodBasicAuth`/`AuthMethodServicenowOAuth`, and `GetAuthMethod` (legacy `oauthEnabled` fallback) |
| `pkg/models/settings_test.go:21-175` | Confirms defaults (empty → basicAuth), OAuth copies user/password too, and that a provisioned `BasicAuthEnabled` is ignored |
| `pkg/httputil/client.go:19,37-82` | Fixed 5-minute HTTP `Timeout`; `AuthCredentials` (username/password required for OAuth too) and `GetHTTPClient` mapping settings → credentials |
| `pkg/httputil/auth.go:23-36` | `oauthRequestBody`: the OAuth2 password grant (`client_id`, `client_secret`, `grant_type=password`, `username`, `password`) — proves OAuth needs all five inputs |
| `pkg/httputil/auth.go:97-116` | `standardizeRequest`: OAuth path fetches a token then sets `Authorization`; Basic path calls `SetBasicAuth(username, password)` |
| `pkg/datasource.go:39-81` | `NewDatasource`: `LoadSettings` → `url.Parse(settings.URL)` → HTTP client; `useSysTables` gates the Schema API (`:53-55`) |
| `pkg/newyork/table_api_v2.go:519,586-588` | `QueryTimeoutSeconds` applied as a per-query context timeout (not the HTTP transport timeout) |
| `pkg/snerrors/errors.go:28-38` | Settings error identities (`invalid server name`, `invalid username`, `invalid password`, `invalid oauth configuration: client ID can't be blank`) |
| `docs/sources/configure.md:45-263` | Authentication tables, additional-settings tables, and the provisioning/Terraform examples that confirm the storage shape and the deprecated `oauthEnabled` note |
| `package.json:38-49` | External component versions (resolved through the catalog — see next table) |

### External editor components

Read at the exact versions the plugin's `package.json` pins via `catalog:` in
`.yarnrc.yml:14-63`.

| Component / type | Version | Package | What was read |
| --- | --- | --- | --- |
| `RadioButtonGroup`, `Input`, `Switch`, `SecretInput`, `InlineField`, `InlineFormLabel`, `useTheme2` | `@grafana/ui@^11.6.7` | grafana/grafana | Prop names (`label`, `placeholder`, `value`, `onChange`, `isConfigured`, `onReset`) — which UI attributes to record, and that `SecretInput` marks a secureJsonData value |
| `DataSourceJsonData`, `DataSourcePluginOptionsEditorProps`, `FeatureToggles` | `@grafana/data@^11.6.7` | grafana/grafana | The base interface `ServiceNowOptions` extends, and the root `url`/`basicAuthUser`/`basicAuth` fields on `DataSourceSettings` the editor mutates |
| `config` (buildInfo / featureToggles) | `@grafana/runtime@^11.6.7` | grafana/grafana | Gates the Secure Socks Proxy switch (excluded) |
| `CustomHeadersSettings` | `@grafana/plugin-ui@^0.13.1` | grafana/plugin-ui | Storage keys it writes: `jsonData.httpHeaderName<N>` + `secureJsonData.httpHeaderValue<N>` — confirmed dynamic/indexed, so not modeled as first-class fields |
| `ConfigSection` | `@grafana/plugin-ui@^0.13.1` | grafana/plugin-ui | The single collapsible section title "ServiceNow Instance Settings" |

## Field provenance

Every dsconfig field, traced from its schema `id` to the upstream `file:line` where each of its
label, placeholder, tooltip (→ `description`), default, storage key, and value type is defined.

| Schema `id` | Storage key | Target | Label source | Placeholder / options / default source | Value type source | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `root_url` | `url` | `root` | `selectors.ts:37` (`label: 'URL'`) | Placeholder `selectors.ts:38`; tooltip `selectors.ts:35-36` | `Settings.URL string` `pkg/models/settings.go:34` (from `config.URL` `:82`) | `requiredWhen: "true"` from `IsValid:109`; role `endpoint.baseUrl`; written via `options.url` `ConfigEditor.tsx:67` |
| `jsonData_authMethod` | `authMethod` | `jsonData` | `selectors.ts:31` (`label: 'Authentication Type'`) | Options `types/index.ts:19-22`; tooltip `selectors.ts:32`; default `basicAuth` `auth_method.go:22`, `ConfigEditor.tsx:37` | `AuthMethod string` `auth_method.go:4`; TS `ServiceNowAuthMethod` `types/index.ts:17` | Role `auth.discriminator`; written by `onAuthMethodChange` `ConfigEditor.tsx:112-121` |
| `root_basicAuthUser` | `basicAuthUser` | `root` | `selectors.ts:44` (`label: 'Username'`) | Placeholder `selectors.ts:42`; tooltip `selectors.ts:43` | `Settings.BasicAuthUser string` `pkg/models/settings.go:29` (from `config.BasicAuthUser` `:91,98`) | `requiredWhen: "true"` (both methods need it: `IsValid:114`, password grant `auth.go:32`); role `auth.basic.username`; `help` drawer from `PermissionsHelp.tsx` |
| `secureJsonData_basicAuthPassword` | `basicAuthPassword` | `secureJsonData` | `selectors.ts:50` (`label: 'Password'`) | Placeholder `selectors.ts:48`; tooltip `selectors.ts:49` | `Settings.BasicAuthPassword string` `pkg/models/settings.go:28` (from `DecryptedSecureJSONData["basicAuthPassword"]` `:92,99`) | `requiredWhen: "true"` (`IsValid:117`, password grant `auth.go:33`); role `auth.basic.password` |
| `jsonData_oauthClientID` | `oauthClientID` | `jsonData` | `selectors.ts:56` (`label: 'Client ID'`) | Placeholder `selectors.ts:54`; tooltip `selectors.ts:55` | `Settings.OAuthClientID string` `pkg/models/settings.go:31` (from `jsonData.oauthClientID` `:43,100`) | `dependsOn`/`requiredWhen` `serviceNowOAuth` from conditional render `ConfigEditor.tsx:183` + `IsValid:121`; role `auth.oauth2.clientId` |
| `secureJsonData_oauthClientSecret` | `oauthClientSecret` | `secureJsonData` | `selectors.ts:62` (`label: 'Client Secret'`) | Placeholder `selectors.ts:60`; tooltip `selectors.ts:61` | `Settings.OAuthClientSecret string` `pkg/models/settings.go:32` (from `DecryptedSecureJSONData["oauthClientSecret"]` `:101`) | `dependsOn`/`requiredWhen` `serviceNowOAuth` from `ConfigEditor.tsx:183`; role `auth.oauth2.clientSecret`; needed by password grant `auth.go:26` |
| `jsonData_useSysTables` | `useSysTables` | `jsonData` | `ConfigEditor.tsx:220` (`label="Use Sys Tables?"`) | Tooltip `ConfigEditor.tsx:221`; switch (unchecked by default) | `Settings.UseSysTables bool` `pkg/models/settings.go:35` (from `jsonData.useSysTables` `:49,83`) | Switch; gates the Schema API `datasource.go:53-55`; `defaultValue: false` |
| `jsonData_queryTimeoutSeconds` | `queryTimeoutSeconds` | `jsonData` | `ConfigEditor.tsx:236` (`label="Query Timeout"`) | Tooltip `ConfigEditor.tsx:237`; placeholder "30" `:243`; editor default 30 `:51,246`; backend default 30 `pkg/models/settings.go:73-77` | `Settings.QueryTimeoutSeconds int` `pkg/models/settings.go:36` (from `jsonData.queryTimeoutSeconds` `:50`) | `valueType: number` (dsconfig has no `integer`); `defaultValue: 30`; per-query timeout `table_api_v2.go:519` — deliberately no `transport.timeoutSeconds` role |
| `jsonData_oauthEnabled` | `oauthEnabled` | `jsonData` | — (no UI) | — | `bool json:"oauthEnabled"` `pkg/models/settings.go:47`; TS `types/index.ts:11` | Legacy; read by `initAuthMethod` `ConfigEditor.tsx:33` and `GetAuthMethod` `auth_method.go:17-20`; tagged `legacy` |

## Field inventory summary

| Schema field | Storage key | Target | Editor label | Read by backend? |
| --- | --- | --- | --- | --- |
| `root_url` | `url` | `root` | URL | Yes (`config.URL`) |
| `jsonData_authMethod` | `authMethod` | `jsonData` | Authentication Type | Yes |
| `root_basicAuthUser` | `basicAuthUser` | `root` | Username | Yes (`config.BasicAuthUser`) |
| `secureJsonData_basicAuthPassword` | `basicAuthPassword` | `secureJsonData` | Password | Yes |
| `jsonData_oauthClientID` | `oauthClientID` | `jsonData` | Client ID | Yes (OAuth) |
| `secureJsonData_oauthClientSecret` | `oauthClientSecret` | `secureJsonData` | Client Secret | Yes (OAuth) |
| `jsonData_useSysTables` | `useSysTables` | `jsonData` | Use Sys Tables? | Yes |
| `jsonData_queryTimeoutSeconds` | `queryTimeoutSeconds` | `jsonData` | Query Timeout | Yes |
| `jsonData_oauthEnabled` | `oauthEnabled` | `jsonData` | — (no UI) | Yes (legacy) |

### Frontend-only settings

None. Every modeled field is read by the backend. (`jsonData.enableSecureSocksProxy` and the
dynamic custom-header pairs are written by the editor and consumed by the SDK transport, but are
excluded from this entry — see [Modeling decisions](#modeling-decisions).)

### Backend-only settings

None modeled as first-class fields. The backend also reads dynamic custom HTTP headers
(`httpHeaderName<N>` in jsonData + `httpHeaderValue<N>` in secureJsonData,
`pkg/models/settings.go:59-66`), which the `CustomHeadersSettings` editor writes; these indexed
pairs are intentionally not modeled (see below).

## Where the types are defined

Only config type/field definitions are listed — UI components (`ConfigSection`,
`CustomHeadersSettings`, `PermissionsHelp`, `SecretInput`) and functions/helpers (`LoadSettings`,
`GetAuthMethod`, `initAuthMethod`) are omitted even where they are the reason a field exists.

### Frontend (TypeScript)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `ServiceNowOptions` (jsonData), `ServiceNowAuthMethod`, `ServiceNowSecureOptions`, `authOptions` | `src/types/index.ts:5-27` | plugin (`grafana-servicenow-datasource`) |
| `DataSourceJsonData` (base interface `ServiceNowOptions extends`) | `@grafana/data` `^11.6.7` | `@grafana/data` |
| Root `url`, `basicAuthUser`, `basicAuth` (on `DataSourceSettings`, mutated via `options.*`) | `packages/grafana-data/src/types/datasource.ts` | `@grafana/data` `^11.6.7` |

### Backend (Go)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `Settings` (resolved config shape) and the anonymous jsonData parse struct | `pkg/models/settings.go:18-37,41-51` | plugin (`grafana-servicenow-datasource`) |
| `AuthMethod` (`AuthMethodBasicAuth`, `AuthMethodServicenowOAuth`) | `pkg/models/auth_method.go:4-9` | plugin (`grafana-servicenow-datasource`) |
| `backend.DataSourceInstanceSettings` (carries `JSONData`, `DecryptedSecureJSONData`, root `URL`, `BasicAuthUser`, `BasicAuthEnabled`) | `backend/common.go` | `github.com/grafana/grafana-plugin-sdk-go` `v0.292.1` |

This entry flattens that spread into a single Go `Config` (root `URL`/`BasicAuthUser` as `json:"-"`
+ the five jsonData fields + `DecryptedSecureJSONData`) plus a `SecureJsonDataKey` typed constant
list. `settings.ts` keeps the three canonical TypeScript types (`RootConfig`, `JsonDataConfig`,
`SecureJsonDataConfig`).

## Modeling decisions

- **`authMethod` as a plain jsonData discriminator (not virtual).** Unlike Prometheus's derived
  `authMethod`, ServiceNow stores a real `jsonData.authMethod` that the radio writes directly
  (`onAuthMethodChange`, `ConfigEditor.tsx:112-121`); there are no multi-field write effects. It is
  therefore modeled as a `jsonData` field with role `auth.discriminator` and `defaultValue: "basicAuth"`,
  mirroring the GitHub entry's `selectedAuthType`. The editor's load-time display derivation
  (`initAuthMethod`, `ConfigEditor.tsx:25-38`) and the backend's legacy fallback are captured in an
  instruction, in `LoadConfig`/`ApplyDefaults`, and via the separate `jsonData_oauthEnabled` field.
- **Root fields `url` and `basicAuthUser`.** The backend reads `config.URL`
  (`pkg/models/settings.go:82`) and `config.BasicAuthUser` (`:91,98`), so both are modeled with
  `target: "root"` and carried on the Go `Config` as `json:"-"`. The account **password** reuses the
  standard Grafana Basic-auth secret key `basicAuthPassword`.
- **Root `basicAuth` (enabled) is NOT modeled.** The editor never writes it, and the backend
  ignores its provisioned value, deriving `BasicAuthEnabled` from `authMethod`
  (`pkg/models/settings.go:87-96`, confirmed by `settings_test.go:151-175`). See
  [Upstream findings](#upstream-findings) #1.
- **Permissions modal → `help` drawer.** The always-visible "Permissions" row
  (`ConfigEditor.tsx:214-216`) opens the `PermissionsHelp` modal describing the ACL access the
  ServiceNow account needs. Its content is attached verbatim (translated from JSX to Markdown) as the
  `help` drawer of `root_basicAuthUser` — the most relevant field (the service account).
- **`requiredWhen` encodes the working contract.** `root.url`, `root.basicAuthUser`, and
  `secureJsonData.basicAuthPassword` are required in all cases (`requiredWhen: "true"`) because both
  auth methods need them — ServiceNow OAuth is an OAuth2 **password grant**
  (`pkg/httputil/auth.go:23-36`). `oauthClientID`/`oauthClientSecret` are required only for
  `serviceNowOAuth`. This is deliberately stricter than the upstream `IsValid`, whose header-conditioned
  branching is inconsistent (see [Upstream findings](#upstream-findings) #3–#5).
- **`queryTimeoutSeconds` is `valueType: number`** (dsconfig has no `integer`), with `defaultValue: 30`.
  No role is assigned: it is a per-query context timeout (`table_api_v2.go:519`), not the HTTP transport
  timeout (a fixed 5 minutes, `client.go:19`), so `transport.timeoutSeconds` would be misleading.
- **Field ID naming convention.** IDs are prefixed with their storage target (`root_`, `jsonData_`,
  `secureJsonData_`) followed by the camelCase storage key; the `key` property keeps the plugin's raw
  storage key.
- **Exclusions.** The Secure Socks Proxy switch (`jsonData.enableSecureSocksProxy`,
  `ConfigEditor.tsx:259-299`, AGENTS.md exclusion) and the dynamic custom-header pairs
  (`httpHeaderName<N>`/`httpHeaderValue<N>` written by `CustomHeadersSettings`, `ConfigEditor.tsx:256`
  — indexed keys, not first-class fields) are omitted.

## SDK plugin schema and k8s-style examples (`schema.go`)

`NewSchema()` assembles the `grafana-plugin-sdk-go` `pluginschema.PluginSchema` bundle from the
embedded `dsconfig.json`: the two root fields (`url`, `basicAuthUser`) plus a nested `jsonData`
object become the OpenAPI settings `spec`, the secure fields become `secureValues`, and virtual
fields (none here) are skipped.

`SettingsExamples()` provides the default configuration plus one example per authentication method
and the legacy shape. Each example is a full instance-settings object with root fields, the plugin
configuration under `jsonData`, and the relevant write-only secrets under `secureJsonData`
(obviously-fake angle-bracket placeholders; the default example carries empty required fields to
show what must be filled in):

| Example | Auth | `secureJsonData` |
| --- | --- | --- |
| `""` (default) | Basic auth (schema defaults, empty credentials) | `basicAuthPassword` (empty) |
| `basicAuth` | Basic auth | `basicAuthPassword` |
| `serviceNowOAuth` | ServiceNow OAuth (password grant) | `basicAuthPassword`, `oauthClientSecret` |
| `legacyOAuthEnabled` | Legacy `oauthEnabled: true` (resolves to OAuth) | `basicAuthPassword`, `oauthClientSecret` |

## `LoadConfig` utility (`settings.go`)

`LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error)`
runs the full three-phase load flow and returns a fully-defaulted, validated `Config`:

1. **Parse** — copy the root `URL` and `BasicAuthUser` from `settings`, unmarshal `settings.JSONData`
   into `Config` (mirroring the plugin's anonymous jsonData parse struct,
   `pkg/models/settings.go:41-51`), and copy the decrypted secrets (`basicAuthPassword`,
   `oauthClientSecret`) into `DecryptedSecureJSONData`.
2. **`ApplyDefaults`** — resolve the effective auth method via `GetAuthMethod` (empty → `basicAuth`;
   legacy `oauthEnabled=true` → `serviceNowOAuth`, even over an explicit `basicAuth`; unrecognized →
   `basicAuth`) and clamp `QueryTimeoutSeconds < 1` to `30`
   (`pkg/models/settings.go:73-77`, `pkg/models/auth_method.go:12-22`).
3. **`Validate`** — enforce the runtime contract: `url` always required; `basicAuth` needs
   `basicAuthUser` + `basicAuthPassword`; `serviceNowOAuth` needs `basicAuthUser` +
   `basicAuthPassword` + `oauthClientID` + `oauthClientSecret` (the full password-grant contract).
   Errors are joined so every problem surfaces at once.

Everything is logged via `backend.Logger.FromContext(ctx)` with `datasource_uid`, `datasource_name`,
and `plugin` labels. `ApplyDefaults` and `Validate` are exported so callers that assemble a `Config`
directly (provisioning preview, round-trip tools, tests distinguishing parse-level from policy-level
errors) can invoke each phase individually.

`LoadConfig` guards truly-empty `settings.JSONData` (`len > 0`) so an empty payload defaults cleanly
and fails at `Validate` (missing URL) rather than at `json.Unmarshal`; upstream `LoadSettings`
unmarshals unconditionally, but Grafana always sends at least `{}`.

## Upstream findings

Potential bugs, misleading UX, and consistency issues discovered while researching upstream. The
schema records what the plugin **does**, not what it **should** do; these notes exist so reviewers
can reproduce each finding.

1. **Provisioned root `basicAuth` is ignored.** `pkg/models/settings.go:87-96` overwrites
   `config.BasicAuthEnabled` based on `authMethod`, so a datasource provisioned with
   `basic_auth_enabled: true` (as the Terraform example in `docs/sources/configure.md:221` does) has
   no effect — enablement is derived from `authMethod`. Confirmed dead by
   `pkg/models/settings_test.go:151-175` ("field `BasicAuthEnabled` has no real effect").
2. **Legacy `oauthEnabled` overrides an explicit `authMethod: "basicAuth"`.**
   `GetAuthMethod` (`pkg/models/auth_method.go:12-22`) only short-circuits on an explicit
   `serviceNowOAuth`; it then returns OAuth whenever `oauthEnabled` is true, **before** honoring an
   explicit `basicAuth`. So `{authMethod:"basicAuth", oauthEnabled:true}` resolves to
   `serviceNowOAuth`. Faithfully mirrored in `GetAuthMethod`/`ApplyDefaults` and covered by a test.
3. **`IsValid` never validates `oauthClientID` in the common (no-headers) case.**
   `pkg/models/settings.go:113` enters the Basic branch when `authMethod == basicAuth || len(Headers) == 0`.
   For `serviceNowOAuth` without custom headers this is `true`, so it requires username+password and the
   `else if serviceNowOAuth` `oauthClientID` check (`:120-123`) is **unreachable**. `oauthClientID` is
   only validated when custom headers are present.
4. **`IsValid` never validates `oauthClientSecret` at all**, yet the OAuth2 password grant requires it
   (`pkg/httputil/auth.go:26`). An OAuth datasource missing the client secret loads successfully but
   fails at connect time. (This entry's `Validate` requires it — a deliberate divergence.)
5. **Dead `IsValid` branches + typo.** Because `GetAuthMethod` always returns `basicAuth` or
   `serviceNowOAuth`, the `authMethod == ""` and `else` branches (`pkg/models/settings.go:124-128`) are
   unreachable via `LoadSettings`. The empty-branch error message also has a typo: "authentication
   authentication not set" (`:125`).
6. **OAuth requires Basic credentials.** ServiceNow OAuth is an OAuth2 resource-owner **password
   grant** (`pkg/httputil/auth.go:31-33`), so username + password are mandatory in OAuth mode too —
   the editor reflects this by always showing the Username/Password fields and only conditionally
   showing Client ID/Secret (`ConfigEditor.tsx:183`). This entry encodes `root.basicAuthUser` /
   `secureJsonData.basicAuthPassword` as `requiredWhen: "true"`.
7. **Custom headers change validation requirements.** The "if custom headers used, do not require
   basic auth" branch (`pkg/models/settings.go:112-113`) means adding a custom HTTP header flips which
   fields `IsValid` enforces. Custom headers are not modeled in this schema.
8. **Password placeholder duplicates the tooltip.** Both `selectors.ts:48` (placeholder) and `:49`
   (tooltip) are "Password for the ServiceNow account" — harmless, preserved verbatim.
9. **`dependsOn` vs legacy configs.** `jsonData_oauthClientID`/`oauthClientSecret` use
   `dependsOn: "jsonData_authMethod == 'serviceNowOAuth'"`, which references the raw stored field. A
   legacy datasource with `oauthEnabled: true` and no `authMethod` would not satisfy that expression,
   yet the editor's `initAuthMethod` (`ConfigEditor.tsx:25-38`) still displays the OAuth fields. The
   legacy interplay is documented in an instruction rather than special-cased in `dependsOn`.

## Validation performed

- `dsconfig.ParseAndResolveSchemaJSON` + `Schema.Validate()` (Go validator in this repo) — passes
  (via the `ConfigSchemaValid` conformance subtest).
- JSON Schema validation against [`dsconfig/schema.json`](../../dsconfig/schema.json) under both
  draft-07 (the schema's declared draft) and draft 2020-12, strict (`additionalProperties: false`) —
  passes.
- `go generate ./...` (regenerates the three artifacts), then `gofmt -l .`, `go vet ./...`,
  `go build ./...`, and `go test ./...` inside `registry/` — all clean (schema round-trip, artifact
  drift, spec/secure separation, jsonData/struct parity in both directions, secure-key parity,
  `LoadConfig`/`ApplyDefaults`/`Validate`, and the settings-examples guard).
- `tsc --noEmit --strict` (TypeScript `5.5.4`, matching the catalog) on `settings.ts` — clean.
- The pre-existing `dsconfig` and `schema` workspace modules still build and test cleanly.
