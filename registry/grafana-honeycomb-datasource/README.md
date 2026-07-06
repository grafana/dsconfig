# grafana-honeycomb-datasource

Declarative configuration schema for the Honeycomb datasource plugin (`grafana-honeycomb-datasource`).

## Upstream researched

- **Monorepo**: `github.com/grafana/plugins-private`
- **Commit SHA**: `267f4937806ed6404b6628d13ae358a5d308e376`
- **Plugin path**: `plugins/grafana-honeycomb-datasource/`
- **Plugin version**: `2.15.3` (`package.json:3`)
- **Upstream Go module**: `github.com/grafana/honeycomb-datasource` (`go.mod:1`)

Every value in [`dsconfig.json`](dsconfig.json) — labels, placeholders, tooltips, section titles,
help markdown, defaults, validations, required-when expressions, storage keys, storage targets,
value types, group titles, and instructions — is traceable to a specific `file:line` in the
plugin at this SHA. See [Field provenance](#field-provenance) below.

To reproduce this research (the plugin lives inside a monorepo — do **not** clone it standalone):

```bash
git -C <path-to>/plugins-private fetch origin
git -C <path-to>/plugins-private checkout 267f4937806ed6404b6628d13ae358a5d308e376
cd <path-to>/plugins-private/plugins/grafana-honeycomb-datasource
```

If the monorepo `main` has advanced past this SHA, re-diff the sources listed under
[Sources researched](#sources-researched) before merging any changes to this entry.

## Files

| File | Purpose |
| --- | --- |
| [`dsconfig.json`](dsconfig.json) | dsconfig v1 schema — single source of truth for all config fields, groups, relationships, and instructions |
| [`settings.ts`](settings.ts) | TypeScript models: `RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig` |
| [`settings.go`](settings.go) | Go `Config` model (flat: jsonData fields + `DecryptedSecureJSONData`), `PluginID`, `SecureJsonDataKey` typed constants, defaults constants, and the `LoadConfig`/`ApplyDefaults`/`Validate` utilities |
| [`schema.go`](schema.go) | k8s-style SDK plugin schema: embeds `dsconfig.json`, converts it via `dsconfig.NewSDKSchema`, and defines `SettingsExamples` for each connection variant |
| [`settings_test.go`](settings_test.go) | Table tests for `LoadConfig`, `ApplyDefaults`, and `Validate` |
| [`conformance_test.go`](conformance_test.go) | Runs `schema.RunPluginTests` — the shared dsconfig conformance suite — against `dsconfig.json` |
| [`schema.gen.json`](schema.gen.json), [`settings.gen.json`](settings.gen.json), [`settings.examples.gen.json`](settings.examples.gen.json) | Committed schema artifacts (regenerate with `go generate ./...` inside this directory; the `SchemaArtifactInSync` conformance subtest guards drift) |

There is no per-entry `go.mod` — every registry entry is a subpackage of the shared
[`registry/`](..) module (`github.com/grafana/dsconfig/registry`).

## Sources researched

Every source below was read at the pinned monorepo SHA
(`267f4937806ed6404b6628d13ae358a5d308e376`), plus external editor components at the exact
versions the monorepo `.yarnrc.yml` catalog pins.

### Plugin (`plugins/grafana-honeycomb-datasource` @ `267f493`)

| File | What was read |
| --- | --- |
| `src/plugin.json:3-5,25` | `pluginType` (`id`), `pluginName` (`name`), `docURL` (`info.links[0].url`, the "Docs" link) |
| `src/types.ts:121-134` | `HoneycombOptions` (jsonData: `hostname`, `team`, `environment`, `retentionLimit`), `defaultConfigOptions` (`hostname` default), `HoneycombSecureOptions` (`apiKey`) |
| `src/Views/ConfigEditor.tsx:52-157` | Every field's label, tooltip, placeholder, storage key, section heading (`Access` `:70`, `Environment` `:89`, `Advanced Settings` `:138`), and the `InfoBox` help content (`:54-68`) |
| `src/Views/ConfigEditor.tsx:41-50` | `handler` — retentionLimit is `parseInt(value \|\| '7')`; all other jsonData fields stored as strings |
| `src/Views/ConfigEditor.tsx:21-39` | `onSecureJsonUpdate`/`onSettingReset` — how `apiKey` is written to secureJsonData |
| `src/components/selectors.ts:3-20` | E2E selector labels for each config input |
| `src/Datasource.ts:109-136` | `honeycombUiUrl` — builds the data-link URL from `hostname` (`api`→`ui`), `team`, `environment` |
| `pkg/models/settings.go:13-19` | `Settings` struct: `APIKey` (`json:"-"`), `Env`/environment, `Hostname`, `RetentionLimit`, `Team` |
| `pkg/models/settings.go:23-39` | `LoadSettings`: seeds `apiKey` from decrypted secrets, `hostname=https://api.honeycomb.io`, `retentionLimit=7`, then unconditionally unmarshals jsonData |
| `pkg/models/settings.go:45-71` | `Validate`: non-empty hostname (`:48-51`), parseable request URI (`:52-56`), https scheme (`:57-60`), non-empty apiKey (`:61-64`), non-empty team (`:65-68`) |
| `pkg/models/settings_test.go:11-53` | Confirms the seeded defaults (hostname, retentionLimit) and jsonData parse shape |
| `pkg/httpclient/client.go:16-42` | The api key is sent as the `X-Honeycomb-Team` header on every request (plus a `User-Agent`) |
| `pkg/main.go:40-66` | Instance factory: `LoadSettings` → `httpClient(apiKey)` → `requestor.New(hostname, ...)` → `WithRetentionLimit(retentionLimit*24h)`; root url/basicAuth never read |
| `pkg/main.go:84-93` | HTTP client built from an empty `sdkhttpclient.Options{}` — `settings.HTTPClientOptions(ctx)` is never called |
| `pkg/requestor/requestor.go:27-81` | `New(addr)` parses the hostname; each request **replaces** `url.Path` rather than concatenating |
| `pkg/plugin/healthcheck.go:12-49` | `CheckHealth` calls `Settings.Validate()`, then requires `api_key_access["queries"]` and `api_key_access["columns"]` |
| `pkg/plugin/querydata.go:200-225` | `linkTo` — builds the "Open in Honeycomb" data link from hostname/team/environment |
| `pkg/plugin/querydata.go:281-318` | `retentionLimit` clamps query start times and drives the "Partial results" warning |
| `pkg/models/auth.go:5-19` | `AuthResponse` shape used by the health check's api-key permission probe |
| `package.json`, `.yarnrc.yml` | External component versions (see next table) |

### External editor components

The config editor imports only `@grafana/ui` and `@grafana/data` components; it does **not** use
`@grafana/plugin-ui`, `ConfigSection`, `DataSourceDescription`, or `SecureSocksProxySettings`.
Versions resolved from the monorepo `.yarnrc.yml` catalog (`catalog:` protocol).

| Component / type | Version | Source | What was read |
| --- | --- | --- | --- |
| `LegacyForms.SecretFormField`, `Input`, `InfoBox`, `Icon`, `InlineFormLabel` | `@grafana/ui` `^11.6.7` | catalog `.yarnrc.yml:26` | Prop names (`label`, `placeholder`, `value`, `onChange`, `onBlur`, `onReset`, `isConfigured`, `tooltip`) — confirmed every label/placeholder is passed inline from `ConfigEditor.tsx`, and none of these components write a hidden storage key |
| `DataSourcePluginOptionsEditorProps`, `DataSourceJsonData` | `@grafana/data` `^11.6.7` | catalog `.yarnrc.yml:19` | The editor props type and the base interface `HoneycombOptions` extends |
| `E2ESelectors` | `@grafana/e2e-selectors` (uncataloged; swapped per Grafana version via `resolutions`, aligned with `grafanaDependency >=11.6.7-0`) | `src/components/selectors.ts:1` | Selector-map type only |

## Field provenance

| Schema `id` | Storage key | Target | Label source | Placeholder / default / tooltip source | Value type source |
| --- | --- | --- | --- | --- | --- |
| `secureJsonData_apiKey` | `apiKey` | `secureJsonData` | `ConfigEditor.tsx:74` (`label="Honeycomb API Key"`) | Placeholder `ConfigEditor.tsx:80`; help drawer from `InfoBox` `ConfigEditor.tsx:54-68`; no tooltip in editor | `HoneycombSecureOptions.apiKey` `types.ts:133`; backend `Settings.APIKey` `settings.go:14` |
| `jsonData_hostname` | `hostname` | `jsonData` | `ConfigEditor.tsx:95` (`<InlineFormLabel>URL</InlineFormLabel>`) | Tooltip `ConfigEditor.tsx:93`; placeholder `ConfigEditor.tsx:102`; default `types.ts:129` / `settings.go:28` | `Settings.Hostname string` `settings.go:16`; TS `string` `types.ts:122` |
| `jsonData_team` | `team` | `jsonData` | `ConfigEditor.tsx:111` (`Team Name`) | Tooltip `ConfigEditor.tsx:109`; no placeholder | `Settings.Team string` `settings.go:18`; TS `string` `types.ts:123` |
| `jsonData_environment` | `environment` | `jsonData` | `ConfigEditor.tsx:126` (`Environment Name`) | Tooltip `ConfigEditor.tsx:124`; no placeholder | `Settings.Env string` (`json:"environment"`) `settings.go:15`; TS `string?` `types.ts:124` |
| `jsonData_retentionLimit` | `retentionLimit` | `jsonData` | `ConfigEditor.tsx:144` (`Time Window (days)`) | Tooltip `ConfigEditor.tsx:142`; placeholder `ConfigEditor.tsx:151`; default `7` `settings.go:29` | `Settings.RetentionLimit int` `settings.go:17`; TS `number?` `types.ts:125` |

The `^https://` pattern validation and the `requiredWhen: "true"` markers on hostname, team, and
apiKey encode the backend `Settings.Validate` contract (`settings.go:45-71`), not editor markers —
the editor renders no required indicators.

## Field inventory summary

| Schema field | Storage key | Target | Editor label | Read by backend? |
| --- | --- | --- | --- | --- |
| `secureJsonData_apiKey` | `apiKey` | `secureJsonData` | Honeycomb API Key | Yes — `X-Honeycomb-Team` header |
| `jsonData_hostname` | `hostname` | `jsonData` | URL | Yes — request base URL + data-link base |
| `jsonData_team` | `team` | `jsonData` | Team Name | Yes — required by Validate; used in data links |
| `jsonData_environment` | `environment` | `jsonData` | Environment Name | Yes — used in data links (optional) |
| `jsonData_retentionLimit` | `retentionLimit` | `jsonData` | Time Window (days) | Yes — query start-time clamp (optional) |

### Frontend-only settings

None. Every stored field is read by the backend.

### Backend-only settings

None. Every stored field has an editor control.

### Virtual fields

None. There is no editor-local derived selector (single auth method, no mode/plan discriminator).

## Where the types are defined

Only config type/field definitions are listed. UI components (`SecretFormField`, `InfoBox`,
`InlineFormLabel`, …) and functions/helpers (`LoadSettings`, `honeycombUiUrl`, `handler`, …) are
omitted even where they are the reason a field exists.

### Frontend (TypeScript)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `HoneycombOptions` (jsonData: `hostname`, `team`, `environment`, `retentionLimit`), `HoneycombSecureOptions` (`apiKey`), `defaultConfigOptions` | `src/types.ts:121-134` | plugin (`grafana-honeycomb-datasource`) |
| `DataSourceJsonData` (base interface `HoneycombOptions` extends) | `packages/grafana-data/.../types/datasource.ts` | `@grafana/data` `^11.6.7` |

### Backend (Go)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `Settings` (`APIKey` `json:"-"`, `Env`, `Hostname`, `RetentionLimit`, `Team`) | `pkg/models/settings.go:13-19` | plugin (`grafana-honeycomb-datasource`) |
| `backend.DataSourceInstanceSettings` (carries `JSONData`, `DecryptedSecureJSONData`; root `URL`/`BasicAuth*` unused by this plugin) | `backend/common.go` | `github.com/grafana/grafana-plugin-sdk-go` `v0.292.0` |

This entry flattens that spread into a single Go `Config` (jsonData fields +
`DecryptedSecureJSONData`) plus a `SecureJsonDataKey` typed constant. `settings.ts` keeps the three
canonical TypeScript types (`RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig`).

## Modeling decisions

- **Single API-key auth, no discriminator**: the plugin has exactly one auth method (the
  `X-Honeycomb-Team` API-key header), so there is no `auth.discriminator`, no virtual selector, and
  no mode-conditional fields — closest in shape to the Sentry entry.
- **Groups mirror the editor's sections verbatim**: `Access` (order 1, the API key), `Environment`
  (order 2, URL/team/environment), and `Advanced Settings` (order 3, the retention limit). The
  editor lists Access before the connection section, so this entry preserves that order rather than
  forcing a connection-first layout. `Advanced Settings` is marked `optional: true` (its only field
  is optional).
- **`requiredWhen` vs the editor**: the editor renders no required markers, but the backend
  `Settings.Validate` (`settings.go:45-71`, invoked by the health check) hard-fails without
  `hostname`, `apiKey`, and `team`. Those three carry `requiredWhen: "true"`; an instruction records
  that the editor shows no markers.
- **https pattern validation**: `Validate` rejects a non-https hostname (`settings.go:57-60`), so
  `jsonData_hostname` carries a `pattern` validation `^https://`. The authoritative check remains
  `Config.Validate` in Go.
- **Help drawer**: the editor's top `InfoBox` (`ConfigEditor.tsx:54-68`) is attached as the `help`
  drawer of `secureJsonData_apiKey`, markdown preserved verbatim (including the article-less "new
  Team API Key" phrasing), with a `docURL` to the Honeycomb account page. The API-key field has no
  editor tooltip, so it gets no `description`.
- **No invented tooltips**: `description` is set only on the four fields whose `InlineFormLabel`
  carries a `tooltip` (hostname, team, environment, retentionLimit), verbatim.
- **Flat `Config` in Go**: mirrors the upstream `Settings` (`settings.go:13-19`) verbatim — same
  json tags (`environment`, `hostname`, `retentionLimit`, `team`), same types — with the upstream
  `APIKey` (`json:"-"`) replaced by `DecryptedSecureJSONData`. No root fields are carried because
  the plugin reads none.
- **`LoadConfig` = parse → ApplyDefaults → Validate**: the parse phase mirrors `LoadSettings`
  verbatim (seed hostname + retentionLimit defaults, then unconditionally unmarshal jsonData, then
  copy the decrypted `apiKey`). `ApplyDefaults` re-applies the curated defaults (hostname when empty,
  retentionLimit when zero) for direct-construction callers; `Validate` enforces the health-check
  contract with joined errors. See [Upstream findings](#upstream-findings) #4 for the one behavioral
  divergence this introduces.
- **`SecureJsonDataConfig` is a key list**: the secret is write-only, so the secure type is just the
  array `["apiKey"]`; consumers read `secureJsonFields` to see what is configured.
- **Field ID naming**: `<target>_<camelCaseKey>` (`secureJsonData_apiKey`, `jsonData_hostname`, …).
  `key` keeps the plugin's raw storage key.

## Settings examples (`schema.go`)

`NewSchema()` assembles the `grafana-plugin-sdk-go` `pluginschema.PluginSchema` bundle from the
embedded `dsconfig.json`: the `jsonData` fields become the OpenAPI settings `spec`, `apiKey` becomes
`secureValues`, and no secure value leaks into the spec. `SettingsExamples()` provides the default
configuration plus one example per connection variant; every example is a full instance-settings
object with `jsonData` and the write-only `secureJsonData.apiKey` (obviously-fake angle-bracket
placeholders — the default `""` example carries an empty key to show what must be filled in):

| Example | hostname | Extra jsonData | `secureJsonData.apiKey` |
| --- | --- | --- | --- |
| `""` (default) | `https://api.honeycomb.io` | `team:""`, `retentionLimit:7` | `""` (empty) |
| `usApi` | `https://api.honeycomb.io` | `team` | `<your-honeycomb-api-key>` |
| `euApi` | `https://api.eu1.honeycomb.io` | `team` | `<your-honeycomb-api-key>` |
| `withEnvironment` | `https://api.honeycomb.io` | `team`, `environment` | `<your-honeycomb-api-key>` |
| `extendedRetention` | `https://api.honeycomb.io` | `team`, `retentionLimit:30` | `<your-honeycomb-api-key>` |

The EU host (`https://api.eu1.honeycomb.io`) is Honeycomb's public EU API region, not a value
hard-coded by the plugin (the plugin only hard-codes the US default `https://api.honeycomb.io`); it
is used as a realistic non-default `hostname`.

## `LoadConfig` utility (`settings.go`)

`LoadConfig(ctx, settings)` runs the full three-phase flow and returns a fully-defaulted, validated
`Config`:

1. **Parse** — mirrors `LoadSettings` (`settings.go:23-39`) verbatim: seed `hostname` and
   `retentionLimit` defaults, unconditionally `json.Unmarshal` jsonData (empty/malformed bytes are a
   parse error, as upstream), and copy the decrypted `apiKey` into `DecryptedSecureJSONData`.
2. **`ApplyDefaults`** — fill the curated zero-valued fields (`Hostname` → default when empty,
   `RetentionLimit` → 7 when zero).
3. **`Validate`** — enforce the health-check contract (`settings.go:45-71`): non-empty https
   hostname, non-empty apiKey, non-empty team. Errors are joined.

Everything is logged via `backend.Logger.FromContext(ctx)` with `datasource_uid`,
`datasource_name`, and `plugin` labels. `ApplyDefaults` and `Validate` stay exported for callers
that assemble a `Config` directly.

## Upstream findings

Potential bugs, misleading UX, and consistency issues found while researching upstream. All are
preserved verbatim in the schema — the schema records what the plugin **does**, not what it
**should** do.

1. **`X-Honeycomb-Team` header carries the API key, not the team.** `pkg/httpclient/client.go:41`
   sets `X-Honeycomb-Team: <apiKey>`. The confusingly named header takes the API key (a Honeycomb
   API convention); `jsonData.team` is **not** sent in it.
2. **`team` is required but only used for data links.** `Settings.Validate` hard-fails without
   `team` (`settings.go:65-68`), yet the backend uses `team` only to build "Open in Honeycomb"
   data-link URLs (`querydata.go:209`; `src/Datasource.ts:127`) — queries authenticate with the API
   key alone. A datasource can fetch data but still fail the health check purely for a missing team.
3. **`api`→`ui` is a naive substring replace.** `querydata.go:201` and `src/Datasource.ts:126` do
   `strings.ReplaceAll(hostname, "api", "ui")` / `hostname.replace('api','ui')`. For
   `https://api.honeycomb.io` this yields the correct `https://ui.honeycomb.io`, but a hostname
   without an `api` segment (a proxy) produces a wrong or unchanged UI host, and a hostname with
   `api` elsewhere is mangled (`replace` only swaps the first occurrence in JS but `ReplaceAll` swaps
   every occurrence in Go — the two also disagree on multi-match hosts).
4. **`retentionLimit: 0` degenerates queries.** `LoadSettings` seeds `7` before unmarshal
   (`settings.go:29`), but an explicit `"retentionLimit":0` overrides to `0`, giving a zero-duration
   window so `adjustedStartTime` clamps every query to ~now (`querydata.go:282-284`). Nothing
   validates it. This entry's `ApplyDefaults` deliberately coerces `0`→`7` for editor-parity, so
   `LoadConfig` diverges from upstream `LoadSettings` for this single edge case (documented here).
5. **Empty JSONData fails to load.** `LoadSettings` unmarshals `s.JSONData` unconditionally
   (`settings.go:32`), so a datasource with empty/nil JSONData errors with "unexpected end of JSON
   input" rather than a friendly message. (Normal datasources always store at least `{}`.)
6. **No SDK HTTP options / no Secure Socks Proxy.** `pkg/main.go:85` builds the client from an empty
   `sdkhttpclient.Options{}` and never calls `settings.HTTPClientOptions(ctx)`, so standard
   datasource TLS/proxy/timeout options — including the Secure Socks Proxy — are ignored. There is
   consequently no `jsonData.enableSecureSocksProxy` field to exclude from this entry.
7. **https-only, stricter than the editor.** `Validate` rejects any non-https hostname
   (`settings.go:57-60`), but the editor shows no such constraint and no required marker, so an
   `http://` URL saves cleanly and only fails at the health check.
8. **`retentionLimit` frontend default is handler-only.** `defaultConfigOptions` (`types.ts:128-130`)
   defaults only `hostname`; the `7`-day default is applied by the change handler's
   `parseInt(value \|\| '7')` (`ConfigEditor.tsx:46`) and the backend seed (`settings.go:29`), not by
   `defaultConfigOptions`. A freshly added datasource that never touches the field stores no
   `retentionLimit`, and the backend default applies.

## Validation performed

- `dsconfig.ParseAndResolveSchemaJSON` + `Schema.Validate()` (Go validator in this repo) — passes
  (via the `ConfigSchemaValid` conformance subtest).
- JSON Schema validation against [`dsconfig/schema.json`](../../dsconfig/schema.json) (draft-07,
  strict, `additionalProperties: false`) — passes.
- `go generate ./...` inside this directory (regenerates the three `*.gen.json` artifacts) — clean.
- `go build ./...`, `go vet ./...`, `gofmt -l .`, `go test ./...` inside `registry/` — all pass
  (schema round-trip, artifact-in-sync, spec/secure separation, jsonData/struct parity both
  directions, secure-key parity, and `LoadConfig`/`ApplyDefaults`/`Validate` table tests).
- The pre-existing `dsconfig` and `schema` workspace modules still build and test — pass.
- `tsc --noEmit --strict` on `settings.ts` — clean.
