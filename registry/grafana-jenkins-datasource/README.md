# grafana-jenkins-datasource

Declarative configuration schema for the
[Jenkins datasource plugin](https://github.com/grafana/jenkins-datasource)
(`grafana-jenkins-datasource`).

## Upstream researched

- **Repo**: `github.com/grafana/jenkins-datasource`
- **Ref**: `main`
- **Commit SHA**: `7f8efb4571d2a5cf6fb6350cbb1d3efde6658319`
  (2026-07-01, `Updating plugin-ci-workflows (#74)`)

Every value in [`dsconfig.json`](dsconfig.json) — labels, placeholders,
tooltips (bound to `<InlineField tooltip=…>`), section titles,
defaults, `requiredWhen` expressions, storage keys, storage targets,
value types, group titles, and instructions — is traceable to a
specific `file:line` in the upstream repo at this SHA. See
[Field provenance](#field-provenance) below.

To reproduce this research:

```bash
git clone https://github.com/grafana/jenkins-datasource
cd jenkins-datasource
git checkout 7f8efb4571d2a5cf6fb6350cbb1d3efde6658319
```

If upstream `main` has advanced past this SHA, re-diff the sources
listed under [Sources researched](#sources-researched) before merging
any changes to this entry.

## Files

| File | Purpose |
| --- | --- |
| [`dsconfig.json`](dsconfig.json) | dsconfig v1 schema — single source of truth for all config fields, groups, and instructions |
| [`settings.ts`](settings.ts) | TypeScript models: `RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig` |
| [`settings.go`](settings.go) | Go `Config` model (flat: jsonData fields + `DecryptedSecureJSONData`), `PluginID`, `SecureJsonDataKey` typed constants, and the `LoadConfig` utility |
| [`schema.go`](schema.go) | k8s-style SDK plugin schema: embeds `dsconfig.json`, converts it via `dsconfig.NewSDKSchema`, and defines `SettingsExamples` for each auth variant |
| [`settings_test.go`](settings_test.go) | Table tests for `LoadConfig`, `ApplyDefaults`, and `Validate` |
| [`conformance_test.go`](conformance_test.go) | Runs `schema.RunPluginTests` — the shared dsconfig conformance suite — against `dsconfig.json` |
| `schema.gen.json`, `settings.gen.json`, `settings.examples.gen.json` | Committed schema artifacts (regenerate with `go generate ./...` inside this directory; the `SchemaArtifactInSync` conformance subtest guards drift) |

There is no per-entry `go.mod` — every registry entry is a subpackage
of the shared [`registry/`](..) module
(`github.com/grafana/dsconfig/registry`).

## Sources researched

Every source below was read at the pinned upstream SHA
(`7f8efb4571d2a5cf6fb6350cbb1d3efde6658319`), plus external editor
components at the exact versions the plugin's `package.json` /
`package-lock.json` resolve.

### Plugin repo (`github.com/grafana/jenkins-datasource@7f8efb4`)

| File | What was read |
| --- | --- |
| `src/plugin.json:3-5` | `pluginType` (`id` = `grafana-jenkins-datasource`), `pluginName` (`name` = `Jenkins`) |
| `src/plugin.json:34-39` | `info.links[]` — `Docs` URL = `https://grafana.com/docs/plugins/grafana-jenkins-datasource` (used as `docURL`) |
| `src/components/ConfigEditor.tsx:67-71` | `DataSourceDescription` with `dataSourceName="Jenkins"`, `docsLink="https://grafana.com/docs/plugins/grafana-jenkins-datasource/"`, `hasRequiredFields={true}` |
| `src/components/ConfigEditor.tsx:73` | `<Legend>Connection</Legend>` → group title `Connection` |
| `src/components/ConfigEditor.tsx:74-83` | URL `<InlineField>`: `required`, `invalid={!jsonData.url}`, `error={'URL is required'}`, tooltip `"Jenkins URL, e.g. https://jenkins.example.com"`; `<Input>` placeholder `"Jenkins URL, e.g. https://jenkins.example.com"` writing `jsonData.url` via `onUrlChange` (`:16-24`) |
| `src/components/ConfigEditor.tsx:85` | `<Legend>Authentication</Legend>` → group title `Authentication` |
| `src/components/ConfigEditor.tsx:86-95` | Username `<InlineField>`: label `"User"`, tooltip `"The username to use for authentication"`, `autoComplete="off"`, `<Input>` placeholder `"Username"` writing `jsonData.username` via `onUsernameChange` (`:26-34`) |
| `src/components/ConfigEditor.tsx:96-106` | Password `<InlineField>`: label `"Password"`, tooltip `"The password to use for authentication"`; `<SecretInput>` placeholder `"Password"` bound to `secureJsonFields.password` / `secureJsonData?.password`, writing `secureJsonData.password` via `onPasswordChange` (`:36-43`) and clearing it via `onResetPassword` (`:45-57`) |
| `src/components/ConfigEditor.tsx:108-121` | Secure Socks Proxy `<ConfigSubSection>` — gated on `config.featureToggles.secureSocksDSProxyEnabled && >=10.0.0` (`:59-63`); writes `jsonData.enableSecureSocksProxy`. Excluded per AGENTS.md |
| `src/types.ts:35-46` | Frontend types `JenkinsConfig` (jsonData: `url?`, `username?`, `enableSecureSocksProxy?`) and `JenkinsSecureConfig` (`password?: string`). Note `JenkinsConfig extends DataSourceJsonData`; `url` is `jsonData.url`, NOT root `url` |
| `pkg/plugin/settings.go:10-14` | Backend `Settings` struct: `URL / Username` with `json:"url"` / `json:"username"`, plus an unexported `Password` field populated from `source.DecryptedSecureJSONData["password"]` |
| `pkg/plugin/settings.go:16-28` | `LoadSettings`: `json.Unmarshal(source.JSONData, &settings)` (fatal `"could not unmarshal plugin settings json"` when empty / malformed); `DownstreamError("URL is missing")` when `settings.URL == ""`; then `settings.Password = source.DecryptedSecureJSONData["password"]` |
| `pkg/plugin/datasource.go:50-87` | Instance factory: `LoadSettings`, `dss.HTTPClientOptions(ctx)`, set a 5-minute `Timeouts.Timeout`, then `if settings.Username != ""` wire `httpclient.BasicAuthOptions{User, Password}` (`:66-71`); otherwise no auth is configured. Then `jenkins.NewClient(settings.URL, WithHTTPClient(...))` |
| `pkg/jenkins/client.go:222-257` | `NewClient` normalizes the base URL with `strings.TrimRight(baseURL, "/")` (`:226`); default HTTP timeout 5 minutes; default concurrency 20; default page size 500 |
| `pkg/jenkins/client.go:498-500` | Every outgoing request calls `req.SetBasicAuth(username, password)` when credentials are present |
| `pkg/jenkins/client.go:525-527` | `urlJoin` — joins base URL and API path with exactly one `/` |
| `package.json` / `package-lock.json` | External component versions (see next table) |

### External editor components

Read at the versions the plugin's `package.json` pins (and
`package-lock.json` resolves).

| Component | Version | Source consulted | What was read |
| --- | --- | --- | --- |
| `DataSourceDescription`, `ConfigSubSection` | `@grafana/plugin-ui@0.10.7` (resolved via `^0.10.4`) | `github.com/grafana/plugin-ui` tag `v0.10.7` | `dataSourceName` header text + `docsLink` behavior; `ConfigSubSection` `title` / `description` (no storage fields written) |
| `SecureSocksProxySettings` behavior (excluded) | `@grafana/ui@10.4.19` (resolved via `^10.4.8`) | grafana/grafana `v10.4.19` `packages/grafana-ui/src/components/DataSourceSettings/SecureSocksProxySettings.tsx` and the `onUpdateDatasourceJsonDataOptionChecked(props, 'enableSecureSocksProxy')` helper in `@grafana/data` | Storage key written: `jsonData.enableSecureSocksProxy` — confirmed and excluded per AGENTS.md |
| `Divider`, `InlineField`, `InlineFormLabel`, `InlineSwitch`, `Input`, `Legend`, `SecretInput` | `@grafana/ui@10.4.19` | grafana/grafana `v10.4.19` `packages/grafana-ui/src/components/` | Prop names (`label`, `tooltip`, `placeholder`, `invalid`, `error`, `required`, `interactive`, `labelWidth`, `width`, `autoComplete`, `isConfigured`, `onReset`) — no storage fields written by the components themselves |
| `DataSourcePluginOptionsEditorProps`, `DataSourceJsonData`, `FeatureToggles`, `onUpdateDatasourceJsonDataOptionChecked` | `@grafana/data@10.4.19` (resolved via `^10.4.8`) | grafana/grafana `v10.4.19` `packages/grafana-data/src/types/` and `packages/grafana-data/src/utils/datasource.ts` | Base interface `JenkinsConfig` extends; storage semantics of `onOptionsChange`; the helper writes `enableSecureSocksProxy` on `jsonData` |
| `config` | `@grafana/runtime@10.4.19` (resolved via `^10.4.8`) | grafana/grafana `v10.4.19` `packages/grafana-runtime/src/` | Read at `ConfigEditor.tsx:61-62` to decide whether to render the Secure Socks Proxy sub-section |

## Field provenance

Every dsconfig field, traced from its schema `id` to the upstream
`file:line` where each of its label, placeholder, tooltip, default,
storage key, and value type is defined.

| Schema `id` | Storage key | Target | Label source | Placeholder / options / default source | Value type source | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `jsonData_url` | `url` | `jsonData` | `ConfigEditor.tsx:74` (label `"URL"`) | Placeholder `ConfigEditor.tsx:80` (`"Jenkins URL, e.g. https://jenkins.example.com"`); no default | `Settings.URL string`, `pkg/plugin/settings.go:11`; TS `url?: string`, `types.ts:36` | Description = tooltip `ConfigEditor.tsx:75`; role `endpoint.baseUrl`; `requiredWhen: "true"` because backend returns `DownstreamError("URL is missing")` when empty (`settings.go:23-25`) |
| `jsonData_username` | `username` | `jsonData` | `ConfigEditor.tsx:86` (label `"User"`) | Placeholder `ConfigEditor.tsx:90` (`"Username"`); no default | `Settings.Username string`, `pkg/plugin/settings.go:12`; TS `username?: string`, `types.ts:37` | Description = tooltip `ConfigEditor.tsx:86`; role `auth.basic.username`; not required — empty username means anonymous access (`datasource.go:66-71`) |
| `secureJsonData_password` | `password` | `secureJsonData` | `ConfigEditor.tsx:96` (label `"Password"`) | Placeholder `ConfigEditor.tsx:101` (`"Password"`); no default | Unexported `Settings.Password`, `pkg/plugin/settings.go:13`; consumed via `source.DecryptedSecureJSONData["password"]` at `settings.go:27`; TS `password?: string`, `types.ts:45` | Description = tooltip `ConfigEditor.tsx:96`; role `auth.basic.password`; not required — only wired into `BasicAuthOptions.Password` when username is non-empty (`datasource.go:66-71`) |

## Field inventory summary

| Schema field | Storage key | Target | Editor label | Read by backend? |
| --- | --- | --- | --- | --- |
| `jsonData_url` | `url` | `jsonData` | URL | Yes (required — feeds `jenkins.NewClient` base URL) |
| `jsonData_username` | `username` | `jsonData` | User | Yes (gates BasicAuth wiring) |
| `secureJsonData_password` | `password` | `secureJsonData` | Password | Yes (feeds `httpclient.BasicAuthOptions.Password` when username is set) |
| `jsonData_enableSecureSocksProxy` (excluded) | `enableSecureSocksProxy` | `jsonData` | Enable Secure Socks Proxy | Indirectly (via `dss.HTTPClientOptions(ctx)`) — excluded per AGENTS.md |

### Frontend-only settings

None. Every editor-written field is read by the backend, directly or
via the SDK's `dss.HTTPClientOptions(ctx)`.

### Backend-only settings

None. Every backend-consumed setting has an editor UI, except the
excluded Secure Socks Proxy switch which is Grafana-version-gated and
covered by the SDK's shared field pack.

## Where the types are defined

The configuration types are spread across the plugin and its
dependencies — some fields and base types come from libraries/SDKs
rather than the plugin itself:

### Frontend (TypeScript)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `JenkinsConfig` (jsonData: `url?`, `username?`, `enableSecureSocksProxy?`), `JenkinsSecureConfig` (`password?`) | `src/types.ts:35-46` | plugin ([grafana/jenkins-datasource](https://github.com/grafana/jenkins-datasource)) |
| `DataSourceJsonData` (base interface `JenkinsConfig` extends: `authType?`, `defaultRegion?`, `profile?`, `manageAlerts?`, …) | `packages/grafana-data/src/types/datasource.ts` | `@grafana/data` `10.4.19` (grafana/grafana `v10.4.19`) |
| `DataSourcePluginOptionsEditorProps`, `FeatureToggles`, `onUpdateDatasourceJsonDataOptionChecked` | `packages/grafana-data/src/` | `@grafana/data` `10.4.19` |
| `DataSourceDescription`, `ConfigSubSection` (no storage fields written) | `src/components/` | `@grafana/plugin-ui` `0.10.7` |
| `Divider`, `InlineField`, `InlineFormLabel`, `InlineSwitch`, `Input`, `Legend`, `SecretInput` (no storage fields written) | `packages/grafana-ui/src/components/` | `@grafana/ui` `10.4.19` |
| Secure Socks Proxy — `SecureSocksProxySettings` writes `jsonData.enableSecureSocksProxy` (excluded from this entry) | `packages/grafana-ui/src/components/DataSourceSettings/SecureSocksProxySettings.tsx` | `@grafana/ui` `10.4.19` |

### Backend (Go)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `Settings` (URL, Username, Password), `LoadSettings` | `pkg/plugin/settings.go:10-29` | plugin ([grafana/jenkins-datasource](https://github.com/grafana/jenkins-datasource)) |
| `NewDatasource` (settings wiring, `httpclient.BasicAuthOptions`, `jenkins.NewClient`) | `pkg/plugin/datasource.go:50-87` | plugin |
| `Client`, `NewClient`, `httpGet` (Basic-auth header via `req.SetBasicAuth`) | `pkg/jenkins/client.go` | plugin |
| `backend.DataSourceInstanceSettings` (carries `JSONData`, `DecryptedSecureJSONData`, `URL`, `BasicAuthEnabled` — all root fields unused by this plugin) | `backend/common.go` | `github.com/grafana/grafana-plugin-sdk-go` |
| `httpclient.BasicAuthOptions` (target of the username/password wiring) | `backend/httpclient/httpclient.go` | `github.com/grafana/grafana-plugin-sdk-go` |
| `s.HTTPClientOptions(ctx)` (consumes `jsonData.enableSecureSocksProxy` transparently) | `backend/httpclient` | `github.com/grafana/grafana-plugin-sdk-go` |

The models in this entry flatten that spread into a single Go `Config`
type (jsonData fields + `DecryptedSecureJSONData`) plus a
`SecureJsonDataKey` typed constant list. `settings.ts` keeps the three
canonical TypeScript types (`RootConfig`, `JsonDataConfig`,
`SecureJsonDataConfig`); `RootConfig` is a blank object because the
Jenkins plugin stores nothing at the root level.

## Modeling decisions

- **URL is jsonData, not `root.url`**. The editor writes to
  `jsonData.url` via `onUrlChange` (`ConfigEditor.tsx:16-24,78-79`),
  and the backend unmarshals it via
  `Settings.URL \`json:"url"\`` on the jsonData struct
  (`settings.go:11`). The root `settings.URL` field is never touched
  by the Jenkins plugin.
- **`RootConfig` is a blank object**. Nothing lives at the root level.
- **No auth discriminator**. Jenkins does not have an
  `selectedAuthType`-style field. The backend chooses HTTP Basic auth
  vs. anonymous based purely on whether `jsonData.username` is empty
  (`datasource.go:66-71`), so there is no discriminator field to model
  in the schema.
- **`requiredWhen` encodes the backend contract**. `jsonData.url` is
  `requiredWhen: "true"` because the backend hard-fails on empty URL
  (`settings.go:23-25`); this matches the editor's `required` +
  `invalid={!jsonData.url}` markers at `ConfigEditor.tsx:74`. Neither
  `jsonData.username` nor `secureJsonData.password` is required — the
  editor renders no `required` marker on either, and the backend
  performs no non-empty check.
- **Description = tooltip**. `<InlineField tooltip={…}>` is the only
  place descriptions appear in the editor; the schema copies tooltip
  strings verbatim from `ConfigEditor.tsx:75,86,96`.
- **Username / password roles**. Marked `auth.basic.username` and
  `auth.basic.password` because
  `pkg/plugin/datasource.go:66-71` wires them into
  `httpclient.BasicAuthOptions`, and every outgoing request sets the
  `Authorization: Basic` header (`pkg/jenkins/client.go:498-500`).
- **URL role**. Marked `endpoint.baseUrl` because it is the base URL
  the Jenkins client string-concatenates API paths onto
  (`pkg/jenkins/client.go:226,525-527`).
- **Password may be a Jenkins API token**. Jenkins accepts an API
  token in the Basic-auth password position, so no discriminator is
  needed; captured in a dedicated schema example (`apiToken`) but
  represented in dsconfig as the same `secureJsonData.password` field.
- **Secure Socks Proxy excluded**. `jsonData.enableSecureSocksProxy`
  is deliberately omitted per AGENTS.md, even though it is one of the
  toggles rendered by the editor when Grafana feature-flags it on.
- **Field ID naming convention**. IDs are prefixed with their storage
  target (`jsonData_` / `secureJsonData_`) followed by the camelCase
  storage key, e.g. `jsonData_username`, `secureJsonData_password`.
  The `key` property keeps the plugin's raw storage key.
- **Flat `Config` in Go**. `settings.go` mirrors the upstream
  `Settings` verbatim (URL, Username with identical json tags) plus a
  `DecryptedSecureJSONData` map for the write-only secrets.
- **`SecureJsonDataConfig` is a key list**. Secure values are
  write-only, so the secure type is just the array of secret key names
  (`password`); consumers read `secureJsonFields` to see what is
  configured.

## SDK plugin schema and k8s-style examples (`schema.go`)

`NewSchema()` assembles the `grafana-plugin-sdk-go`
`pluginschema.PluginSchema` bundle (the k8s-style schema Grafana's
datasource API server serves as `{apiVersion}.json`, `v0alpha1`
today) from the embedded `dsconfig.json`: jsonData fields become the
OpenAPI settings `spec`, secure fields become `secureValues`.

`SettingsExamples()` provides the default configuration plus one
k8s-style example per authentication variant the editor supports.
Each example is a full instance-settings object with the plugin
configuration nested under `jsonData` and the write-only password
under `secureJsonData` (placeholder values to be replaced with real
secrets):

| Example | Connection | `secureJsonData.password` |
| --- | --- | --- |
| `""` (default) | Empty URL (must be filled in) | empty |
| `anonymous` | `https://jenkins.example.com`, no username | empty |
| `basicAuth` | `https://jenkins.example.com`, username `grafana` | password |
| `apiToken` | `https://jenkins.example.com`, username `grafana` | Jenkins API token |
| `legacyUsernameOnly` | `https://jenkins.example.com`, username `grafana`, no password | empty |

## `LoadConfig` utility (`settings.go`)

`LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error)`
runs the full three-phase load flow on a datasource instance's
settings and returns a fully-defaulted, validated `Config`:

1. **Parse** — unmarshal jsonData into `Config` (empty JSONData is a
   parse error, mirroring `pkg/plugin/settings.go:18-21`), then copy
   decrypted secrets by known key into `DecryptedSecureJSONData`.
2. **`ApplyDefaults`** — no-op (the Jenkins plugin has no
   editor-parity defaults to apply). Kept as a stable extension point.
3. **`Validate`** — enforce the runtime contract: `URL` must be
   non-empty (`pkg/plugin/settings.go:23-25`). Username / password are
   deliberately not required; the backend treats an empty username as
   "anonymous access".

Everything is logged via `backend.Logger.FromContext(ctx)` with
`datasource_uid`, `datasource_name`, and `plugin` labels so log lines
carry request context.

### Direct access to individual phases

`(*Config).ApplyDefaults()` and `(Config).Validate() error` are
exported separately for callers that want to compose them themselves
(e.g. provisioning preview, schema-example round-trip, tests that need
to distinguish parse-level from policy-level errors).

## Upstream findings

Potential bugs, misleading UX, and consistency issues discovered while
researching upstream. All preserved verbatim in the schema — the
schema records what the plugin **does**, not what it **should** do.

1. **Empty JSONData is a fatal parse error, not a sensible default
   state.** `pkg/plugin/settings.go:18-21` calls
   `json.Unmarshal(source.JSONData, &settings)` unconditionally, so a
   brand-new datasource with no jsonData written yet fails with
   `"could not unmarshal plugin settings json"` rather than being
   defaulted. Only observable in provisioning / API paths since the
   editor always writes at least an empty `jsonData` object.
2. **URL storage lives in `jsonData`, not at the root**, despite the
   field being called `url`. This is inconsistent with most Grafana
   datasources (Prometheus, Loki, Postgres, MySQL, etc.) which use
   `settings.URL` (root). Only relevant to consumers writing
   provisioning payloads.
3. **The Jenkins docs link is stored twice.** `src/plugin.json:35`
   uses `https://grafana.com/docs/plugins/grafana-jenkins-datasource`
   (no trailing slash), while `ConfigEditor.tsx:69` renders
   `docsLink="https://grafana.com/docs/plugins/grafana-jenkins-datasource/"`
   (with trailing slash). Both resolve to the same page but the two
   are kept in sync manually. This entry uses the `plugin.json` value
   for `docURL`.
4. **Password without username is silently ignored.** The backend
   only wires `httpclient.BasicAuthOptions` when
   `settings.Username != ""` (`pkg/plugin/datasource.go:66-71`), so a
   datasource with a password but no username makes anonymous
   requests. The editor renders no warning; consumers writing
   provisioning payloads for anonymous access can safely omit
   `secureJsonData.password`.
5. **Username-with-empty-password sends an empty Basic-auth
   password.** Symmetrically, when a username is set but no password
   has been provided, the backend still wires
   `BasicAuthOptions{User: <username>, Password: ""}` — Jenkins then
   decides whether the resulting empty-password request is
   authorised. Preserved as the `legacyUsernameOnly` example.
6. **Trailing slash on `jsonData.url` is tolerated.**
   `pkg/jenkins/client.go:226` runs
   `strings.TrimRight(baseURL, "/")` at client construction, and
   `urlJoin` at `:525-527` further normalises join points. Unlike
   several other datasources (Sentry, GitHub), a trailing slash does
   NOT produce double-slash URLs here — but the editor placeholder
   still shows it without a trailing slash.
7. **`error='URL is required'` in the editor vs. `"URL is missing"`
   in the backend.** The editor's `<InlineField error='URL is required'>`
   at `ConfigEditor.tsx:74` and the backend's
   `DownstreamError("URL is missing")` at
   `pkg/plugin/settings.go:23-25` disagree on the exact string.
   Preserved via the description on the URL field (tooltip verbatim)
   and the `LoadConfig` validation error (`jenkins URL (jsonData.url)
   is required`), which is a canonical form.
8. **`DataSourceDescription hasRequiredFields={true}` but only URL
   is marked required.** Only the URL `<InlineField>` has the
   `required` prop; `User` and `Password` do not. `hasRequiredFields`
   just controls whether the "Fields marked with * are required"
   header note is rendered.

## Validation performed

- `dsconfig.ParseAndResolveSchemaJSON` + `Schema.Validate()` (Go
  validator in this repo) — passes.
- JSON Schema validation against
  [`dsconfig/schema.json`](../../dsconfig/schema.json)
  (draft-07, `additionalProperties: false`) — passes.
- `go test ./...` on this module — passes (schema bundle shape,
  secure values, examples, `LoadConfig` for each variant + malformed
  input + missing-URL cases, `SchemaArtifactInSync` guard).
- `settings.go` / `schema.go`: `go build`, `go vet`, `gofmt` — clean.
- `settings.ts`: `tsc --noEmit --strict` — clean.
