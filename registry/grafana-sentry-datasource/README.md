# grafana-sentry-datasource

Declarative configuration schema for the [Sentry datasource plugin](https://github.com/grafana/sentry-datasource) (`grafana-sentry-datasource`).

## Upstream researched

- **Repo**: `github.com/grafana/sentry-datasource`
- **Ref**: `main`
- **Commit SHA**: `cdb55dea50cf344f04cada3b3114b35af6cd032a` (2026-06-30, `Fix TruffleHog SentryToken false positive in CD publish workflow (#717)`)

Every value in [`dsconfig.json`](dsconfig.json) — labels, placeholders, tooltips (bound to
`<Field description>`), section titles, defaults, `requiredWhen` expressions, storage keys,
storage targets, value types, group titles, and instructions — is traceable to a specific
`file:line` in the upstream repo at this SHA. See [Field provenance](#field-provenance) below.

To reproduce this research:

```bash
git clone https://github.com/grafana/sentry-datasource
cd sentry-datasource
git checkout cdb55dea50cf344f04cada3b3114b35af6cd032a
```

If upstream `main` has advanced past this SHA, re-diff the sources listed under [Sources
researched](#sources-researched) before merging any changes to this entry.

## Files

| File | Purpose |
| --- | --- |
| [`dsconfig.json`](dsconfig.json) | dsconfig v1 schema — single source of truth for all config fields, groups, and instructions |
| [`settings.ts`](settings.ts) | TypeScript models: `RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig` |
| [`settings.go`](settings.go) | Go `Config` model (flat: jsonData fields + `DecryptedSecureJSONData`), `PluginID`, `SecureJsonDataKey` typed constants, and the `LoadConfig` utility |
| [`schema.go`](schema.go) | k8s-style SDK plugin schema: embeds `dsconfig.json`, converts it via `dsconfig.NewSDKSchema`, and defines `SettingsExamples` for each connection variant |
| [`settings_test.go`](settings_test.go) | Table tests for `LoadConfig`, `ApplyDefaults`, and `Validate` |
| [`conformance_test.go`](conformance_test.go) | Runs `schema.RunPluginTests` — the shared dsconfig conformance suite — against `dsconfig.json` |
| `schema.gen.json`, `settings.gen.json`, `settings.examples.gen.json` | Committed schema artifacts (regenerate with `go generate ./...` inside this directory; the `SchemaArtifactInSync` conformance subtest guards drift) |

There is no per-entry `go.mod` — every registry entry is a subpackage of the shared
[`registry/`](..) module (`github.com/grafana/dsconfig/registry`).

## Sources researched

Every source below was read at the pinned upstream SHA
(`cdb55dea50cf344f04cada3b3114b35af6cd032a`), plus external editor components at the exact
versions the plugin's `package.json` / `package-lock.json` pins.

### Plugin repo (`github.com/grafana/sentry-datasource@cdb55de`)

| File | What was read |
| --- | --- |
| `src/plugin.json:4-5` | `pluginType` (`id` = `grafana-sentry-datasource`), `pluginName` (`name` = `Sentry`). `info.links` is empty at `:22`, so `docURL` falls back to the editor's own `docsLink` |
| `src/editors/SentryConfigEditor.tsx:47-51` | `DataSourceDescription` with `dataSourceName="Sentry"`, `docsLink="https://grafana.com/grafana/plugins/grafana-sentry-datasource/"`, `hasRequiredFields` |
| `src/editors/SentryConfigEditor.tsx:19` | URL React state initial value = `jsonData?.url \|\| DEFAULT_SENTRY_URL` (frontend default source) |
| `src/editors/SentryConfigEditor.tsx:53` | `<ConfigSection title={ConfigEditorSelectors.SentrySettings.GroupTitle}>` → group title `Sentry Settings` |
| `src/editors/SentryConfigEditor.tsx:54-70` | URL `<Field>`: `required`, `invalid={!url}`, label / placeholder / description / aria-label read from selectors; `<Input>` writes `jsonData.url` on blur |
| `src/editors/SentryConfigEditor.tsx:71-87` | OrgSlug `<Field>`: `required`, `invalid={!jsonData.orgSlug}`, error `'Organization is required'`; `<Input>` writes `jsonData.orgSlug` on blur |
| `src/editors/SentryConfigEditor.tsx:88-133` | AuthToken `<Field>`: conditional Configured/Reset UI when `secureJsonFields.authToken`, otherwise `<Input type="password">` writing `secureJsonData.authToken` on blur when non-empty. `error='Auth token is required'` |
| `src/editors/SentryConfigEditor.tsx:24-43` | `onOptionChange` writes to `jsonData` (not root); `onSecureOptionChange` writes to `secureJsonData` and toggles `secureJsonFields[key]` |
| `src/editors/SentryConfigEditor.tsx:135` | `<AdditionalSettings jsonData={jsonData} onOptionChange={onOptionChange} />` |
| `src/components/config-editor/AdditionalSettings.tsx:17` | `shouldShowSection = config.secureSocksDSProxyEnabled && grafanaVersion >= 10.0.0` — the Secure Socks Proxy switch is Grafana-version gated |
| `src/components/config-editor/AdditionalSettings.tsx:23-27` | `<ConfigSection title="Additional settings" description="Additional settings are optional settings that can be configured for more control over your data source." isCollapsible>` |
| `src/components/config-editor/AdditionalSettings.tsx:29-40` | Secure Socks Proxy `<Switch>` writes `jsonData.enableSecureSocksProxy` — deliberately excluded from this entry |
| `src/components/config-editor/AdditionalSettings.tsx:41-50` | `tlsSkipVerify` `<Switch>` — label + tooltip from `Components.ConfigEditor.TLSSkipVerify` |
| `src/selectors.ts:8-38` | Every label, placeholder, tooltip, and aria-label the editor renders. Group title `Sentry Settings` at `:8` |
| `src/constants.ts:105` | `DEFAULT_SENTRY_URL = 'https://sentry.io'` |
| `src/types.ts:54-62` | Frontend types `SentryConfig` (jsonData: `url`, `orgSlug`, `enableSecureSocksProxy?`, `tlsSkipVerify?`) and `SentrySecureConfig` (`authToken: string`). Note `SentryConfig extends DataSourceJsonData`; `url` is `jsonData.url`, NOT root.url |
| `pkg/plugin/settings.go:11-16` | Backend `SentryConfig` struct: `URL / OrgSlug / TLSSkipVerify` with `json:"url" json:"orgSlug" json:"tlsSkipVerify"`, plus unexported `authToken` |
| `pkg/plugin/settings.go:18-29` | `Validate()`: errors when URL / OrgSlug / authToken are empty |
| `pkg/plugin/settings.go:31-52` | `GetSettings`: `json.Unmarshal(s.JSONData, cfg)` (fatal on empty JSONData); default `URL = "https://sentry.io"` when empty; require OrgSlug; copy `s.DecryptedSecureJSONData["authToken"]`; require authToken |
| `pkg/plugin/plugin.go:44-73` | Instance factory: `GetSettings`, then `s.HTTPClientOptions(ctx)`, then `if settings.TLSSkipVerify` set `opt.TLS.InsecureSkipVerify = true`, then `sentry.NewSentryClient(URL, OrgSlug, authToken, ...)` |
| `pkg/plugin/plugin.go:24-26` | `PluginID = "grafana-sentry-datasource"` — matches src/plugin.json:5 |
| `pkg/sentry/sentry.go:23-33` | `NewSentryClient` also defaults an empty baseURL to `DefaultSentryURL` |
| `pkg/sentry/sentry.go:54-56` | `sc.BaseURL + path` — string concatenation of the base URL and API path (source of the trailing-slash pitfall) |
| `pkg/sentry/client.go:37-40` | Every outgoing request adds `Authorization: Bearer <authToken>` |
| `pkg/errors/errors.go:15-19` | The three fatal errors: `ErrorUnmarshalingSettings`, `ErrorInvalidOrganizationSlug`, `ErrorInvalidAuthToken`, plus `ErrorInvalidSentryConfig` (returned by `SentryConfig.Validate` when URL is empty) |
| `pkg/util/util.go` | `DefaultSentryURL = "https://sentry.io"` (backend copy of the constant) |
| `pkg/plugin/settings_test.go:13-44` | Confirms empty JSONData is a parse error and default URL applies when empty |
| `package.json` / `package-lock.json` | External component versions (see next table) |

### External editor components

Read at the versions the plugin's `package.json` pins (and `package-lock.json` resolves).

| Component | Version | Source consulted | What was read |
| --- | --- | --- | --- |
| `DataSourceDescription`, `ConfigSection` | `@grafana/plugin-ui@0.13.0` | `github.com/grafana/plugin-ui` tag `v0.13.0` | `dataSourceName` header text + `docsLink` behavior; `title` / `isCollapsible` / `description` props (no storage fields written) |
| `SecureSocksProxySettings` behavior (excluded) | `@grafana/ui@12.4.3` (resolved via `^12.2.0`) | grafana/grafana `v12.4.3` `packages/grafana-ui/src/components/DataSourceSettings/SecureSocksProxySettings.tsx` | Storage key written: `jsonData.enableSecureSocksProxy` — confirmed and excluded per AGENTS.md |
| `Field`, `Input`, `Button`, `Switch`, `Divider` | `@grafana/ui@12.4.3` | grafana/grafana `v12.4.3` `packages/grafana-ui/src/components/` | Prop names (`label`, `description`, `placeholder`, `invalid`, `error`, `required`, `type`, `width`, `autoComplete`) — no storage fields written by the components themselves |
| `DataSourcePluginOptionsEditorProps`, `DataSourceJsonData` | `@grafana/data@12.4.3` (resolved via `^12.2.0`) | grafana/grafana `v12.4.3` `packages/grafana-data/src/types/datasource.ts` | Base interface `SentryConfig` extends; storage semantics of `onOptionsChange` |

## Field provenance

Every dsconfig field, traced from its schema `id` to the upstream `file:line` where each of its
label, placeholder, tooltip, default, storage key, and value type is defined.

| Schema `id` | Storage key | Target | Label source | Placeholder / options / default source | Value type source | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `jsonData_url` | `url` | `jsonData` | `selectors.ts:10` (label `"Sentry URL"`) via `Field label={...}` at `SentryConfigEditor.tsx:56` | Placeholder `selectors.ts:11` = `DEFAULT_SENTRY_URL` (`constants.ts:105` = `"https://sentry.io"`); default = same, matching `SentryConfigEditor.tsx:19` init and `settings.go:37-40` backend fallback | `Settings.URL string`, `pkg/plugin/settings.go:12`; TS `url: string`, `types.ts:55` | Description = tooltip `selectors.ts:13`; role `endpoint.baseUrl`. No `requiredWhen` — backend supplies default when empty |
| `jsonData_orgSlug` | `orgSlug` | `jsonData` | `selectors.ts:16` (`"Sentry Org"`) via `Field label={...}` at `SentryConfigEditor.tsx:73` | Placeholder `selectors.ts:17` (`"Sentry org slug"`) | `Settings.OrgSlug string`, `pkg/plugin/settings.go:13`; TS `orgSlug: string`, `types.ts:56` | Description = tooltip `selectors.ts:19`; `requiredWhen: "true"` because backend returns `ErrorInvalidOrganizationSlug` when empty (`settings.go:41-43`) |
| `secureJsonData_authToken` | `authToken` | `secureJsonData` | `selectors.ts:22` (`"Sentry Auth Token"`) via `Field label={...}` at `SentryConfigEditor.tsx:90,111` | Placeholder `selectors.ts:23` (`"Sentry Authentication Token"`) | `SentrySecureConfig.authToken: string`, `types.ts:60-62`; consumed via `s.DecryptedSecureJSONData["authToken"]` at `settings.go:44-45` | Description = tooltip `selectors.ts:25`; role `auth.bearer.token` (`pkg/sentry/client.go:37-40` `Authorization: Bearer`); `requiredWhen: "true"` because backend returns `ErrorInvalidAuthToken` when empty (`settings.go:47-50`) |
| `jsonData_tlsSkipVerify` | `tlsSkipVerify` | `jsonData` | `selectors.ts:36` (`"Skip TLS Verify"`) via `Field label={...}` at `AdditionalSettings.tsx:42` | Default `false` (`AdditionalSettings.tsx:47`: `value={jsonData.tlsSkipVerify \|\| false}`) | `Settings.TLSSkipVerify bool`, `pkg/plugin/settings.go:14`; TS `tlsSkipVerify?: boolean`, `types.ts:58` | Description = tooltip `selectors.ts:37`; role `transport.tlsSkipVerify`; only used to set `opt.TLS.InsecureSkipVerify = true` at `pkg/plugin/plugin.go:57-63` |

## Field inventory summary

| Schema field | Storage key | Target | Editor label | Read by backend? |
| --- | --- | --- | --- | --- |
| `jsonData_url` | `url` | `jsonData` | Sentry URL | Yes (defaulted when empty) |
| `jsonData_orgSlug` | `orgSlug` | `jsonData` | Sentry Org | Yes (required) |
| `secureJsonData_authToken` | `authToken` | `secureJsonData` | Sentry Auth Token | Yes (required, sent as Bearer) |
| `jsonData_tlsSkipVerify` | `tlsSkipVerify` | `jsonData` | Skip TLS Verify | Yes (feeds `opt.TLS.InsecureSkipVerify`) |
| `jsonData_enableSecureSocksProxy` (excluded) | `enableSecureSocksProxy` | `jsonData` | Enable Secure Socks Proxy | Indirectly (via `s.HTTPClientOptions(ctx)`) — excluded per AGENTS.md |

### Frontend-only settings

None. Every editor-written field is read by the backend (directly or via the SDK's
`s.HTTPClientOptions(ctx)`).

### Backend-only settings

None. Every backend-consumed setting has an editor UI, except the excluded Secure Socks Proxy
switch which is Grafana-version-gated and covered by the SDK's shared field pack.

## Where the types are defined

The configuration types are spread across the plugin and its dependencies — some fields and base
types come from libraries/SDKs rather than the plugin itself:

### Frontend (TypeScript)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `SentryConfig` (jsonData: url, orgSlug, enableSecureSocksProxy?, tlsSkipVerify?), `SentrySecureConfig` (authToken) | `src/types.ts:54-62` | plugin ([grafana/sentry-datasource](https://github.com/grafana/sentry-datasource)) |
| `DataSourceJsonData` (base interface `SentryConfig` extends: `authType?`, `defaultRegion?`, `profile?`, `manageAlerts?`, …) | `packages/grafana-data/src/types/datasource.ts` | `@grafana/data` `12.4.3` (grafana/grafana `v12.4.3`) |
| `DataSourcePluginOptionsEditorProps` | `packages/grafana-data/src/` | `@grafana/data` `12.4.3` |
| `DataSourceDescription`, `ConfigSection` (no storage fields written) | `src/components/ConfigEditor/`, `src/components/` | `@grafana/plugin-ui` `0.13.0` |
| `Field`, `Input`, `Button`, `Switch`, `Divider` (no storage fields written) | `packages/grafana-ui/src/components/` | `@grafana/ui` `12.4.3` |
| Secure Socks Proxy — `SecureSocksProxySettings` writes `jsonData.enableSecureSocksProxy` (excluded from this entry) | `packages/grafana-ui/src/components/DataSourceSettings/SecureSocksProxySettings.tsx` | `@grafana/ui` `12.4.3` |

### Backend (Go)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `SentryConfig` (URL, OrgSlug, TLSSkipVerify, authToken), `Validate`, `GetSettings` | `pkg/plugin/settings.go:11-52` | plugin ([grafana/sentry-datasource](https://github.com/grafana/sentry-datasource)) |
| `NewDatasource` (settings wiring, `opt.TLS.InsecureSkipVerify`, `sentry.NewSentryClient`) | `pkg/plugin/plugin.go:44-81` | plugin |
| `SentryClient`, `NewSentryClient`, `Do` (Bearer header) | `pkg/sentry/sentry.go`, `pkg/sentry/client.go` | plugin |
| `DefaultSentryURL` | `pkg/util/util.go` | plugin |
| `backend.DataSourceInstanceSettings` (carries `JSONData`, `DecryptedSecureJSONData`, `URL`, `BasicAuthEnabled` — all root fields unused by this plugin) | `backend/common.go` | `github.com/grafana/grafana-plugin-sdk-go` |
| `httpclient.Options.TLS.InsecureSkipVerify` (target of the `TLSSkipVerify` toggle) | `backend/httpclient/httpclient.go` | `github.com/grafana/grafana-plugin-sdk-go` |
| `s.HTTPClientOptions(ctx)` (consumes root `enableSecureSocksProxy` transparently) | `backend/httpclient` | `github.com/grafana/grafana-plugin-sdk-go` |

The models in this entry flatten that spread into a single Go `Config` type (jsonData fields +
`DecryptedSecureJSONData`) plus a `SecureJsonDataKey` typed constant list. `settings.ts` keeps
the three canonical TypeScript types (`RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig`);
`RootConfig` is a blank object because the Sentry plugin stores nothing at the root level.

## Modeling decisions

- **URL is jsonData, not root.url**. The editor writes to `jsonData.url` via `onOptionChange`
  (`SentryConfigEditor.tsx:24-32,67`), and the backend unmarshals it via
  `SentryConfig.URL \`json:"url"\`` on the jsonData struct (`settings.go:12`). The root
  `settings.URL` field is never touched by the Sentry plugin.
- **`RootConfig` is a blank object**. Nothing lives at the root level.
- **`requiredWhen` encodes the backend contract, not the editor markers**. `orgSlug` and
  `authToken` are `requiredWhen: "true"` because the backend hard-fails on empty values
  (`settings.go:41-50`). `url` is NOT marked required because the backend supplies a default
  when empty (`settings.go:37-40`), even though the editor renders it as required.
- **Description = tooltip**. `<Field description={...tooltip}>` is the only place descriptions
  appear in the editor; the schema copies tooltip strings verbatim from `selectors.ts:13,19,25,37`.
- **Auth token role**. Marked `auth.bearer.token` because
  `pkg/sentry/client.go:37-40` sends `Authorization: Bearer <authToken>` on every request.
- **TLS skip verify role**. Marked `transport.tlsSkipVerify` because it flows directly into
  the SDK's `httpclient.Options.TLS.InsecureSkipVerify` at `pkg/plugin/plugin.go:57-63`.
- **Secure Socks Proxy excluded**. `jsonData.enableSecureSocksProxy` is deliberately omitted
  per AGENTS.md, even though it is one of the two toggles in the "Additional settings"
  collapsible.
- **`Config.EnableSecureSocksProxy` kept in Go for round-tripping**. The struct still carries
  the field with `json:"enableSecureSocksProxy,omitempty"` so provisioning payloads survive
  parse → validate → back-to-JSON; the field is just not exposed in dsconfig.
- **Field ID naming convention**. IDs are prefixed with their storage target (`jsonData_` /
  `secureJsonData_`) followed by the camelCase storage key, e.g. `jsonData_orgSlug`,
  `secureJsonData_authToken`. The `key` property keeps the plugin's raw storage key.
- **Flat `Config` in Go**. `settings.go` mirrors the upstream `SentryConfig` verbatim (URL,
  OrgSlug, TLSSkipVerify with identical json tags) plus a `DecryptedSecureJSONData` map for
  the write-only secrets.
- **`SecureJsonDataConfig` is a key list**. Secure values are write-only, so the secure type is
  just the array of secret key names (`authToken`); consumers read `secureJsonFields` to see
  what is configured.

## SDK plugin schema and k8s-style examples (`schema.go`)

`NewSchema()` assembles the `grafana-plugin-sdk-go` `pluginschema.PluginSchema` bundle (the
k8s-style schema Grafana's datasource API server serves as `{apiVersion}.json`, `v0alpha1`
today) from the embedded `dsconfig.json`: root fields plus a nested `jsonData` object become
the OpenAPI settings `spec`, secure fields become `secureValues`.

`SettingsExamples()` provides the default configuration plus one k8s-style example per
connection variant. Each example is a full instance-settings object with the plugin
configuration nested under `jsonData` and the write-only auth token under `secureJsonData`
(placeholder values to be replaced with real secrets):

| Example | Connection | `secureJsonData` |
| --- | --- | --- |
| `""` (default) | `https://sentry.io`, empty org slug | `authToken` (empty) |
| `sentrySaaS` | `https://sentry.io`, org slug present | `authToken` |
| `selfHosted` | Self-hosted (`https://sentry.example.com`) | `authToken` |
| `selfHostedTLSSkipVerify` | Self-hosted with `tlsSkipVerify: true` | `authToken` |
| `legacyMissingURL` | No `jsonData.url` — defaults applied by `LoadConfig` | `authToken` |

## `LoadConfig` utility (`settings.go`)

`LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error)`
runs the full three-phase load flow on a datasource instance's settings and returns a
fully-defaulted, validated `Config`:

1. **Parse** — unmarshal jsonData into `Config` (empty JSONData is a parse error, mirroring
   `pkg/plugin/settings.go:33-36`), then copy decrypted secrets by known key into
   `DecryptedSecureJSONData`.
2. **`ApplyDefaults`** — fill the curated default: `URL = DefaultSentryURL` when empty
   (`pkg/plugin/settings.go:37-40`).
3. **`Validate`** — enforce the runtime contract: URL, OrgSlug, and the `authToken` secret
   must be non-empty. Errors are joined so every problem surfaces at once.

Everything is logged via `backend.Logger.FromContext(ctx)` with `datasource_uid`,
`datasource_name`, and `plugin` labels so log lines carry request context.

### Direct access to individual phases

`(*Config).ApplyDefaults()` and `(Config).Validate() error` are exported separately for callers
that want to compose them themselves (e.g. provisioning preview, schema-example round-trip,
tests that need to distinguish parse-level from policy-level errors).

## Upstream findings

Potential bugs, misleading UX, and consistency issues discovered while researching upstream.
All preserved verbatim in the schema — the schema records what the plugin **does**, not what
it **should** do.

1. **Empty JSONData is a fatal parse error, not a sensible default state.**
   `pkg/plugin/settings.go:33-36` calls `json.Unmarshal(s.JSONData, config)` unconditionally,
   so a brand-new datasource with no jsonData written yet fails with
   `ErrorUnmarshalingSettings` rather than being defaulted. The editor never sends empty
   JSONData because the initial URL state is set from `DEFAULT_SENTRY_URL`
   (`SentryConfigEditor.tsx:19`), so this is only observable in provisioning / API paths.
2. **URL storage lives in jsonData, not at the root**, despite the field being called `url`.
   This is inconsistent with most Grafana datasources (Prometheus, Loki, Postgres, MySQL,
   etc.) which use `settings.URL` (root). Only relevant to consumers writing provisioning
   payloads.
3. **`docURL` is not declared in `src/plugin.json`.** `info.links` is an empty array at
   `src/plugin.json:22`. The editor hard-codes its own docs link at
   `SentryConfigEditor.tsx:49`. This entry uses that hard-coded URL for `docURL`.
4. **Trailing slash on `jsonData.url` produces double-slash API URLs.** `pkg/sentry/sentry.go:54-56`
   builds request URLs as `sc.BaseURL + path` where `path` starts with `/api/0/...`. A trailing
   slash on the base URL yields `https://sentry.io//api/0/...`. The editor placeholder does not
   include a trailing slash; provisioning payloads should also omit it.
5. **`TLSSkipVerify` is not gated by URL scheme.** The setting is applied whenever the flag is
   true, even for `http://` targets (`pkg/plugin/plugin.go:57-63`). Harmless — TLS options are
   inert over plain HTTP — but potentially confusing.
6. **The secret input toggles `secureJsonFields.authToken=false` while typing.**
   `SentryConfigEditor.tsx:100-103` — clicking Reset calls `onSecureOptionChange('authToken',
   authToken, false)` with the CURRENT typed value and a `false` flag, which temporarily
   marks the secret as unconfigured until save. Intentional to keep the input editable, but
   means `secureJsonFields.authToken` can transiently be `false` even when a token exists.
7. **Editor's `hasRequiredFields` is `true` but the schema shows fields required in the UI
   only.** `DataSourceDescription hasRequiredFields` (`:50`) just controls the "Fields marked
   with * are required" note; each `<Field required>` prop drives per-field marks. `url`,
   `orgSlug`, and `authToken` are marked required in the editor, but the backend only rejects
   empty `orgSlug` / `authToken` (URL is defaulted). Preserved via `requiredWhen` on the two
   fields the backend actually enforces.
8. **`DEFAULT_SENTRY_URL` is declared twice**: once in `src/constants.ts:105` (used by the
   editor), and once in `pkg/util/util.go` (`DefaultSentryURL`, used by
   `pkg/plugin/settings.go` and `pkg/sentry/sentry.go`). The two are kept in sync manually.

## Validation performed

- `dsconfig.ParseAndResolveSchemaJSON` + `Schema.Validate()` (Go validator in this repo) — passes.
- JSON Schema validation against [`dsconfig/schema.json`](../../dsconfig/schema.json)
  (draft-07, `additionalProperties: false`) — passes.
- `go test ./...` on this module — passes (schema bundle shape, secure values, examples,
  `LoadConfig` for each variant + malformed input + missing-slug + missing-token cases,
  `SchemaArtifactInSync` guard).
- `settings.go` / `schema.go`: `go build`, `go vet`, `gofmt` — clean.
- `settings.ts`: `tsc --noEmit --strict` — clean.
