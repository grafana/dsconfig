# grafana-splunk-monitoring-datasource

Declarative configuration schema for the **Splunk Infrastructure Monitoring** (formerly SignalFx,
also known as Splunk Observability) datasource plugin (`grafana-splunk-monitoring-datasource`).

## Upstream researched

- **Repo (monorepo)**: `github.com/grafana/plugins-private`
- **Plugin path**: `plugins/grafana-splunk-monitoring-datasource/`
- **Commit SHA**: `267f4937806ed6404b6628d13ae358a5d308e376`
- **Plugin's own canonical repo / Go module**: `github.com/grafana/signalfx-datasource`
  (`package.json:32` `"repository": "github:grafana/signalfx-datasource"`; the backend imports use
  `github.com/grafana/signalfx-datasource/pkg/...`). The plugin is developed in that repo and
  vendored into the `plugins-private` monorepo — all `file:line` references below are relative to
  the plugin path inside the monorepo at the pinned SHA.

Every value in [`dsconfig.json`](dsconfig.json) — labels, placeholders, tooltips (bound to
`InlineField tooltip` / editors surface `description` as the tooltip), section titles/descriptions,
`requiredWhen` expressions, storage keys, storage targets, value types, group titles, and
instructions — is traceable to a specific `file:line` in the upstream plugin at this SHA. See
[Field provenance](#field-provenance) below.

To reproduce this research:

```bash
git clone https://github.com/grafana/plugins-private
cd plugins-private
git checkout 267f4937806ed6404b6628d13ae358a5d308e376
cd plugins/grafana-splunk-monitoring-datasource
```

If upstream `main` has advanced past this SHA, re-diff the sources listed under [Sources
researched](#sources-researched) before merging any changes to this entry.

## Files

| File | Purpose |
| --- | --- |
| [`dsconfig.json`](dsconfig.json) | dsconfig v1 schema — single source of truth for all config fields, groups, relationships, and instructions |
| [`settings.ts`](settings.ts) | TypeScript models: `RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig` |
| [`settings.go`](settings.go) | Go `Config` model (flat: jsonData fields + `DecryptedSecureJSONData`), `PluginID`, `SecureJsonDataKey` typed constants, and the `LoadConfig` utility |
| [`schema.go`](schema.go) | k8s-style SDK plugin schema: embeds `dsconfig.json`, converts it via `dsconfig.NewSDKSchema`, and defines `SettingsExamples` for each connection variant |
| [`settings_test.go`](settings_test.go) | Table tests for `LoadConfig`, `ApplyDefaults`, and `Validate` |
| [`conformance_test.go`](conformance_test.go) | Runs `schema.RunPluginTests` — the shared dsconfig conformance suite — against `dsconfig.json` |
| `schema.gen.json`, `settings.gen.json`, `settings.examples.gen.json` | Committed schema artifacts (regenerate with `go generate ./...` inside this directory; the `SchemaArtifactInSync` conformance subtest guards drift) |

There is no per-entry `go.mod` — every registry entry is a subpackage of the shared
[`registry/`](..) module (`github.com/grafana/dsconfig/registry`). The Go package name is
`splunkmonitoringdatasource`.

## Sources researched

Every source below was read at the pinned monorepo SHA
(`267f4937806ed6404b6628d13ae358a5d308e376`), plus external editor components at the catalog
versions the plugin resolves.

### Plugin (`plugins/grafana-splunk-monitoring-datasource` @ `267f4937`)

| File | What was read |
| --- | --- |
| `src/plugin.json:3-5` | `type: "datasource"`, `pluginName` (`name` = `Splunk Infrastructure Monitoring`), `pluginType` (`id` = `grafana-splunk-monitoring-datasource`) |
| `src/plugin.json:31` | `info.links[0].url` → `docURL` = `https://grafana.com/docs/plugins/grafana-splunk-monitoring-datasource` |
| `src/plugin.json:41-42` | `grafanaDependency: ">=11.6.7-0"`, `grafanaVersion: "11.x.x"` |
| `src/components/ConfigEditor.tsx:24` | `accessTokenConfigured = secureJsonFields.accessToken` — the write-only-secret read path |
| `src/components/ConfigEditor.tsx:29-33` | `renderSSP` — Secure Socks Proxy subsection is gated on `secureSocksDSProxyEnabled` + Grafana >= 10.0.0 |
| `src/components/ConfigEditor.tsx:35-48` | `onResetAccessToken` — clears `secureJsonFields.accessToken` + `secureJsonData.accessToken` |
| `src/components/ConfigEditor.tsx:52-56` | `DataSourceDescription` (`dataSourceName="Splunk Infrastructure Monitoring"`, `docsLink="https://grafana.com/grafana/plugins/grafana-splunk-monitoring-datasource/"`, `hasRequiredFields`) |
| `src/components/ConfigEditor.tsx:60-62` | `ConfigSection title="Authentication"` — the only top-level section title |
| `src/components/ConfigEditor.tsx:63-71` | Access Token `InlineField` (`label="Access Token"`, `labelWidth={24}`, `required`) + `SecretInput` writing `secureJsonData.accessToken` (`onUpdateDatasourceSecureJsonDataOption(props, 'accessToken')` at `:67`) |
| `src/components/ConfigEditor.tsx:72-81` | Realm Name `InlineField` (`label="Realm Name"`, `labelWidth={24}`, no `required`, no tooltip) + `Input` writing `jsonData.realmName`, `placeholder="us1"` (`:78`) |
| `src/components/ConfigEditor.tsx:83-87` | `ConfigSubSection title="Custom URLs"` + verbatim `description` |
| `src/components/ConfigEditor.tsx:88-97` | Metrics MetaData URL `InlineField` (`label="Metrics MetaData URL"`, `tooltip={'Optional Metrics MetaData URL.'}`) + `Input` writing `jsonData.url_metrics_metadata`, `placeholder="https://api.us1.signalfx.com"` (`:94`) |
| `src/components/ConfigEditor.tsx:98-107` | SignalFlow URL `InlineField` (`label="SignalFlow URL"`, `tooltip={'Optional SignalFlow URL'}`) + `Input` writing `jsonData.url_signalflow`, `placeholder="https://stream.us1.signalfx.com"` (`:104`) |
| `src/components/ConfigEditor.tsx:110-140` | Conditional `ConfigSubSection "Secure Socks Proxy"` writing `jsonData.enableSecureSocksProxy` — deliberately excluded from this entry |
| `src/types.ts:3-10` | `SignalFxJsonData` (jsonData): `realmName: string` (`:4`), `url_metrics_metadata?: string` (`:6`), `url_signalflow?: string` (`:8`), `enableSecureSocksProxy?: boolean` (`:9`) |
| `src/types.ts:18-20` | `SignalFxSecureJsonData`: `accessToken: string` (`:19`) |
| `src/data/SfxDatasource.ts:1-126` | Frontend datasource — confirms no config migration; realm/URLs are consumed only by the backend |
| `src/data/migrate.ts:1-32` | `migrateVariableQuery` — query (not config) migration; no legacy config storage shape |
| `pkg/models/settings.go:13-19` | Backend `Settings` struct: `AccessToken` (no json tag; from decrypted secret), `Realm` `json:"realmName,omitempty"`, `URLMetricsMetaData` `json:"url_metrics_metadata,omitempty"`, `URLSignalFlow` `json:"url_signalflow,omitempty"`, `HttpClientOptions httpclient.Options` `json:"-"` |
| `pkg/models/settings.go:21-42` | `LoadSettings`: unconditional `json.Unmarshal(s.JSONData, &settings)` (fatal on empty/malformed, `:22-26`); require decrypted `accessToken` → `backend.DownstreamError("invalid/empty access token")` (`:27-31`); load `s.HTTPClientOptions(ctx)` for proxy (`:34-39`) |
| `pkg/models/settings_test.go:17-115` | `TestLoadSettings`: empty/empty-token → error; `realmName`-only and `realmName` + both custom URLs happy paths |
| `pkg/client/client.go:61-76` | `NewSignalFxClient`: hard-fail `"required access token is missing"` when the token is empty (`:62-63`); passes `Realm` to the REST client (`:68`) |
| `pkg/client/rest.go:225` | Every request adds the `X-SF-TOKEN: <accessToken>` header |
| `pkg/client/rest.go:82-88` | `CheckHealth` → `GetMetrics` — the health check that hits the metrics-metadata base URL |
| `pkg/client/rest.go:90-120` | `Query` — SignalFlow execute against `GetBaseURL(RestAPITypeSignalFlow, c.Realm)` (`:91`) |
| `pkg/client/rest.go:205-208,248-271` | `request` / `getURL` — realm trailing-dot handling and metrics-metadata URL assembly |
| `pkg/client/rest.go:307-314` | Secure Socks Proxy wiring via `settings.HttpClientOptions.ProxyOptions` (why `enableSecureSocksProxy` is backend-consumed only through the SDK) |
| `pkg/client/rest.go:332-353` | `RestAPIType` constants + `GetBaseURL`: `https://api.{realm}.signalfx.com` (metrics-metadata, `:342-346`), `https://stream.{realm}.signalfx.com` (SignalFlow, `:347-352`), fallthrough default `:353` |
| `package.json:38-47` | Config-editor dependencies, all via the `catalog:` protocol (resolved below) |

### External editor components

Read at the catalog versions the plugin resolves. `package.json` references every `@grafana/*`
dependency via the `catalog:` protocol; the concrete versions come from the monorepo catalog
(`plugins-private/.yarnrc.yml:14-60`). No plugin-local overrides exist for these packages, so the
catalog pins apply.

| Component / helper | Catalog version | Package | What was read |
| --- | --- | --- | --- |
| `DataSourceDescription`, `ConfigSection`, `ConfigSubSection` | `^0.13.1` | `@grafana/plugin-ui` | Header block + docs link; section `title` / `description` props (no storage fields written) |
| `Input`, `InlineField`, `InlineSwitch`, `InlineFormLabel`, `SecretInput`, `useStyles2` | `^11.6.7` | `@grafana/ui` | Prop names (`label`, `labelWidth`, `required`, `placeholder`, `value`, `onChange`, `isConfigured`, `onReset`, `aria-label`) so the correct UI attributes were recorded; `SecretInput` writes no placeholder |
| `DataSourceJsonData`, `DataSourcePluginOptionsEditorProps`, `FeatureToggles`, `onUpdateDatasourceJsonDataOption`, `onUpdateDatasourceJsonDataOptionChecked`, `onUpdateDatasourceSecureJsonDataOption` | `^11.6.7` | `@grafana/data` | Base jsonData interface + storage-key semantics of the update helpers used by the editor |
| `config` (buildInfo / featureToggles) | `^11.6.7` | `@grafana/runtime` | Read at `ConfigEditor.tsx:29-33` to gate the Secure Socks Proxy switch |
| `cx` | `11.10.6` | `@emotion/css` | Editor styling only; writes no storage fields |

## Field provenance

Every dsconfig field, traced from its schema `id` to the upstream `file:line` where each of its
label, placeholder, tooltip, storage key, and value type is defined.

| Schema `id` | Storage key | Target | Label source | Placeholder / tooltip source | Value type source | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `secureJsonData_accessToken` | `accessToken` | `secureJsonData` | `ConfigEditor.tsx:63` (`InlineField label="Access Token"`) | No placeholder (SecretInput, `ConfigEditor.tsx:64-70`); no tooltip | `SignalFxSecureJsonData.accessToken: string`, `types.ts:19`; consumed via `s.DecryptedSecureJSONData["accessToken"]`, `settings.go:27` | Role `auth.apiKey.value` (sent as `X-SF-TOKEN` header, `rest.go:225`); `requiredWhen: "true"` — backend hard-fails when empty (`settings.go:27-30`, `client.go:62-63`); only field marked `required` in the editor |
| `jsonData_realmName` | `realmName` | `jsonData` | `ConfigEditor.tsx:72` (`InlineField label="Realm Name"`) | `ConfigEditor.tsx:78` (`placeholder="us1"`); no tooltip | `Settings.Realm string` `json:"realmName"`, `settings.go:15`; TS `realmName: string`, `types.ts:4` | No role (the closed vocabulary has no realm/region concept); `requiredWhen: "jsonData_urlMetricsMetadata == '' \|\| jsonData_urlSignalflow == ''"` — needed for URL derivation unless both overrides are set (`rest.go:339-353`) |
| `jsonData_urlMetricsMetadata` | `url_metrics_metadata` | `jsonData` | `ConfigEditor.tsx:88` (`InlineField label="Metrics MetaData URL"`) | `ConfigEditor.tsx:94` (`placeholder="https://api.us1.signalfx.com"`); tooltip `"Optional Metrics MetaData URL."` (`:88`) → `description` | `Settings.URLMetricsMetaData string` `json:"url_metrics_metadata"`, `settings.go:16`; TS `url_metrics_metadata?: string`, `types.ts:6` | No role (see [Modeling decisions](#modeling-decisions)); overrides the metrics-metadata base (`rest.go:342-346`) |
| `jsonData_urlSignalflow` | `url_signalflow` | `jsonData` | `ConfigEditor.tsx:98` (`InlineField label="SignalFlow URL"`) | `ConfigEditor.tsx:104` (`placeholder="https://stream.us1.signalfx.com"`); tooltip `"Optional SignalFlow URL"` (`:98`) → `description` | `Settings.URLSignalFlow string` `json:"url_signalflow"`, `settings.go:17`; TS `url_signalflow?: string`, `types.ts:8` | No role; overrides the SignalFlow base (`rest.go:347-352`) |

## Field inventory summary

| Schema field | Storage key | Target | Editor label | Read by backend? |
| --- | --- | --- | --- | --- |
| `secureJsonData_accessToken` | `accessToken` | `secureJsonData` | Access Token | Yes (required; sent as `X-SF-TOKEN`) |
| `jsonData_realmName` | `realmName` | `jsonData` | Realm Name | Yes (derives both base URLs) |
| `jsonData_urlMetricsMetadata` | `url_metrics_metadata` | `jsonData` | Metrics MetaData URL | Yes (overrides metrics-metadata base) |
| `jsonData_urlSignalflow` | `url_signalflow` | `jsonData` | SignalFlow URL | Yes (overrides SignalFlow base) |
| `jsonData_enableSecureSocksProxy` (excluded) | `enableSecureSocksProxy` | `jsonData` | Enable Secure Socks Proxy | Indirectly, via `settings.HttpClientOptions` (`rest.go:307-314`) — excluded per AGENTS.md |

### Frontend-only settings

None. Every editor-written field is read by the backend (directly, or — for the excluded Secure
Socks Proxy toggle — through the SDK's proxy plumbing).

### Backend-only settings

None. Every backend-consumed setting has an editor UI, except the excluded, Grafana-version-gated
Secure Socks Proxy switch.

## Where the types are defined

The configuration types are spread across the plugin and its dependencies — some base types come
from libraries/SDKs rather than the plugin itself. Only config type/field definitions are listed
(UI components and helper functions are omitted even where they are the reason a field exists).

### Frontend (TypeScript)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `SignalFxJsonData` (jsonData: `realmName`, `url_metrics_metadata?`, `url_signalflow?`, `enableSecureSocksProxy?`), `SignalFxSecureJsonData` (`accessToken`) | `src/types.ts:3-20` | plugin ([grafana/signalfx-datasource](https://github.com/grafana/signalfx-datasource)) |
| `DataSourceJsonData` (base interface `SignalFxJsonData` extends: `authType?`, `defaultRegion?`, `manageAlerts?`, …) | `packages/grafana-data/src/types/datasource.ts` | `@grafana/data` (catalog `^11.6.7`) |
| `FeatureToggles` (`secureSocksDSProxyEnabled`, read to gate the excluded Secure Socks Proxy switch) | `packages/grafana-data/src/types/featureToggles.gen.ts` | `@grafana/data` (catalog `^11.6.7`) |
| Secure Socks Proxy — the editor writes `jsonData.enableSecureSocksProxy` directly (`ConfigEditor.tsx:132-137`); excluded from this entry | `src/components/ConfigEditor.tsx:110-140` | plugin |

### Backend (Go)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `Settings` (jsonData: `Realm`, `URLMetricsMetaData`, `URLSignalFlow`; plus `AccessToken` from the decrypted secret; plus `HttpClientOptions`), `LoadSettings` | `pkg/models/settings.go:13-42` | plugin ([grafana/signalfx-datasource](https://github.com/grafana/signalfx-datasource)) |
| `RestAPIType` (`RestAPITypeMetricMetaData`, `RestAPITypeSignalFlow`), `GetBaseURL` (realm → base URLs) | `pkg/client/rest.go:332-353` | plugin |
| `backend.DataSourceInstanceSettings` (carries `JSONData`, `DecryptedSecureJSONData`, and root fields like `URL`, `BasicAuthEnabled` — all root fields unused by this plugin) | `backend/common.go` | `github.com/grafana/grafana-plugin-sdk-go` `v0.292.1` |
| `httpclient.Options` (proxy plumbing target of `HttpClientOptions`; `ProxyOptions` read at `rest.go:307-314`) | `backend/httpclient` | `github.com/grafana/grafana-plugin-sdk-go` `v0.292.1` |

This entry flattens that spread into a single Go `Config` type (jsonData fields +
`DecryptedSecureJSONData`) plus a `SecureJsonDataKey` typed constant list. `settings.ts` keeps the
three canonical TypeScript types (`RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig`);
`RootConfig` is a blank object because the plugin stores nothing at the root level.

## Modeling decisions

- **Single API-token auth, no discriminator.** The plugin has exactly one auth method (an API
  access token in `secureJsonData.accessToken`), so there is no auth-type discriminator and no
  virtual selector fields — this entry is closest to the Sentry API-token cousin.
- **`accessToken` role = `auth.apiKey.value`.** The token is sent as the custom `X-SF-TOKEN`
  header (`rest.go:225`), not an `Authorization: Bearer` header, so `auth.apiKey.value` is the
  accurate role. The header name is hard-coded, so there is no configurable `auth.apiKey.key`
  field.
- **No role on `realmName` or the custom URL fields.** The closed role vocabulary
  ([`dsconfig/schema.go:540-605`](../../dsconfig/schema.go)) has no realm/region concept, so
  `realmName` gets none. The two custom URL fields *are* base URLs, but they are optional
  per-API overrides that are usually blank — the datasource's actual endpoint is realm-derived —
  so tagging one (or both) of two optional overrides with `endpoint.baseUrl` would be arbitrary
  and misleading. Both are left roleless.
- **Groups mirror the editor verbatim.** The editor puts the access token and realm under a
  section literally titled **"Authentication"** (`ConfigEditor.tsx:60`), with the two custom URLs
  in a nested **"Custom URLs"** subsection (`:83`). This entry keeps that structure: an
  `authentication` group (`accessToken`, `realmName`) and an `optional` `custom-urls` group with
  the subsection's verbatim description. The realm is a connection field but the editor groups it
  under Authentication; fidelity to the editor wins.
- **`requiredWhen` encodes the backend contract, not the editor markers.** `accessToken` is
  `requiredWhen: "true"` (backend hard-fails without it). `realmName` is
  `requiredWhen: "jsonData_urlMetricsMetadata == '' || jsonData_urlSignalflow == ''"`: the backend
  derives both base URLs from the realm and only skips it when *both* custom URLs are provided
  (`rest.go:339-353`). The editor marks only the access token required.
- **Secure Socks Proxy excluded.** `jsonData.enableSecureSocksProxy` is deliberately omitted per
  AGENTS.md, even though the editor renders the toggle (Grafana-version-gated) and the backend
  honors it through `settings.HttpClientOptions` (`rest.go:307-314`).
- **Flat `Config` in Go.** `settings.go` mirrors the upstream `Settings` (`pkg/models/settings.go:13-19`)
  verbatim for the jsonData fields (`realmName`, `url_metrics_metadata`, `url_signalflow` with
  identical json tags) plus a `DecryptedSecureJSONData` map for the write-only secret. The upstream
  `AccessToken` field (populated from the decrypted secret) and `HttpClientOptions` field
  (`json:"-"` SDK proxy plumbing) are not carried, matching the API-token cousin entries. Root
  datasource fields are not carried because the plugin never reads them.
- **`ApplyDefaults` is a no-op.** The plugin defines no config defaults (the `us1` realm
  placeholder is only a placeholder), so there is nothing to default. It is kept exported for API
  parity and documented as intentionally empty.
- **`Validate` mirrors the effective runtime contract.** It enforces the access token
  (`settings.go:27-30`, `client.go:62-63`) and the realm-or-both-URLs rule from `GetBaseURL` /
  `CheckHealth` (`rest.go:82-88,339-353`). Upstream `LoadSettings` validates only the access token;
  like the Datadog entry's `Validate` (which mirrors the health check rather than `LoadSettings`),
  this encodes the contract a *working* datasource requires.
- **`SecureJsonDataConfig` is a key list.** Secure values are write-only, so the secure type is
  just the array of secret key names (`accessToken`); consumers read `secureJsonFields` to see
  what is configured.

## SDK plugin schema and k8s-style examples (`schema.go`)

`NewSchema()` assembles the `grafana-plugin-sdk-go` `pluginschema.PluginSchema` bundle (the
k8s-style schema Grafana's datasource API server serves as `{apiVersion}.json`, `v0alpha1` today)
from the embedded `dsconfig.json`: the nested `jsonData` object becomes the OpenAPI settings
`spec`, and the secure field becomes `secureValues`.

`SettingsExamples()` provides the default configuration plus one k8s-style example per connection
variant. Each example is a full instance-settings object with the plugin configuration nested
under `jsonData` and the write-only access token under `secureJsonData` (obviously-fake
angle-bracket placeholders — replace with a real token; the default example carries an empty
token):

| Example | Connection | `secureJsonData` |
| --- | --- | --- |
| `""` (default) | Empty realm, empty token — fails validation until filled | `accessToken` (empty) |
| `realm` | Realm `us1`, default derived URLs | `accessToken` |
| `customUrls` | Realm `us1` plus both custom URL overrides | `accessToken` |
| `customUrlsWithoutRealm` | No realm; both custom URLs set (valid without a realm) | `accessToken` |

## `LoadConfig` utility (`settings.go`)

`LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error)`
runs the full three-phase load flow and returns a fully-defaulted, validated `Config`:

1. **Parse** — `json.Unmarshal(settings.JSONData, &cfg)`, mirroring `LoadSettings`
   (`pkg/models/settings.go:22-26`) verbatim: empty or malformed JSONData is a parse error. Then
   copy the decrypted `accessToken` into `DecryptedSecureJSONData`.
2. **`ApplyDefaults`** — a no-op (the plugin defines no config defaults).
3. **`Validate`** — enforce the access token and the realm-or-both-custom-URLs contract. Errors are
   joined so every problem surfaces at once.

Everything is logged via `backend.Logger.FromContext(ctx)` with `datasource_uid`,
`datasource_name`, and `plugin` labels so log lines carry request context.

`(*Config).ApplyDefaults()` and `(Config).Validate() error` are exported separately for callers
that assemble a `Config` themselves (provisioning preview, schema-example round-trip, tests that
distinguish parse-level from policy-level errors).

## Upstream findings

Potential bugs, misleading UX, and consistency issues discovered while researching upstream. All
are preserved verbatim in the schema — the schema records what the plugin **does**, not what it
**should** do.

1. **An empty realm with no custom URLs produces a broken host.** `GetBaseURL`
   (`pkg/client/rest.go:340-353`) formats `https://api.%s.signalfx.com` / `https://stream.%s.signalfx.com`
   with the realm; an empty realm yields `https://api..signalfx.com` (double dot). There is no
   default realm and no validation of it in `LoadSettings`, so a datasource saved without a realm
   (and without both custom URLs) loads successfully but every request fails at connection time.
   Encoded here as `requiredWhen` on `realmName` and enforced in `Config.Validate`.
2. **The realm is effectively required but not marked required in the editor.**
   `DataSourceDescription` sets `hasRequiredFields` (`ConfigEditor.tsx:55`) and only the Access
   Token `InlineField` carries `required` (`:63`); Realm Name does not (`:72`), despite being
   necessary for URL derivation.
3. **Empty JSONData is a fatal parse error, not a defaulted state.** `LoadSettings`
   (`pkg/models/settings.go:22-26`) calls `json.Unmarshal(s.JSONData, &settings)` unconditionally,
   so a datasource created with no jsonData written yet fails with a JSON parse error rather than
   defaulting. The editor always writes at least `realmName`, so this is observable mainly via
   provisioning / API paths. Mirrored by `LoadConfig` (empty JSONData → `parse jsonData` error).
4. **Redundant realm dot handling between the two request paths.** For metrics-metadata requests,
   `request` appends a trailing `.` to a non-empty realm (`rest.go:205-208`) and passes it to
   `getURL` → `GetBaseURL`, which immediately strips it with `strings.TrimSuffix(realm, ".")`
   (`rest.go:340`) — a no-op round-trip. The SignalFlow path passes `c.Realm` directly without the
   dot (`rest.go:91`). Harmless (both end up trimmed) but inconsistent.
5. **Custom URLs are used without trailing-slash normalization.** `GetBaseURL` returns
   `Settings.URLMetricsMetaData` / `URLSignalFlow` verbatim (`rest.go:343-345,348-350`) and callers
   string-concatenate paths onto the base (e.g. `getURL` builds `%s/%s?%s` at `rest.go:270`), so a
   custom URL with a trailing slash yields double-slash request URLs. The placeholders omit a
   trailing slash; provisioning payloads should too.
6. **The docs URL differs between `plugin.json` and the editor.** `src/plugin.json:31` links to
   `https://grafana.com/docs/plugins/grafana-splunk-monitoring-datasource`, while the editor's
   `DataSourceDescription docsLink` (`ConfigEditor.tsx:54`) uses
   `https://grafana.com/grafana/plugins/grafana-splunk-monitoring-datasource/`. This entry uses the
   `plugin.json` value for `docURL` per AGENTS.md.
7. **The `SecretInput` reset toggles `secureJsonFields.accessToken=false`.** `onResetAccessToken`
   (`ConfigEditor.tsx:35-48`) clears both the flag and the value, which is expected; but note that
   `secureJsonFields.accessToken` is the only reliable read-side signal of whether a token is
   configured (the value itself is write-only).

## Validation performed

- `dsconfig.ParseAndResolveSchemaJSON` + `Schema.Validate()` (Go validator in this repo) — passes
  (via the `ConfigSchemaValid` conformance subtest).
- JSON Schema validation of `dsconfig.json` against [`dsconfig/schema.json`](../../dsconfig/schema.json)
  (draft-07, `additionalProperties: false`) — passes.
- `go generate ./...` in this directory regenerates the three `*.gen.json` artifacts; the
  `SchemaArtifactInSync` conformance subtest confirms they are in sync.
- `go build ./...`, `go vet ./...`, `gofmt -l .`, and `go test ./...` in the `registry/` module —
  clean / passing (schema round-trip, spec/secure separation, jsonData↔struct key + type parity,
  secure-key parity, and `LoadConfig` / `ApplyDefaults` / `Validate` table tests).
- `tsc --noEmit --strict` on `settings.ts` (TypeScript `5.5.4`) — clean.
- The pre-existing `dsconfig` and `schema` workspace modules still build.
