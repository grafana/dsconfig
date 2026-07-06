# grafana-wavefront-datasource

Declarative configuration schema for the **Wavefront (VMware Aria Operations for Applications)
datasource plugin** (`grafana-wavefront-datasource`).

## Upstream researched

- **Repo**: `github.com/grafana/plugins-private` (monorepo)
- **Ref**: `main`
- **Commit SHA**: `267f4937806ed6404b6628d13ae358a5d308e376`
- **Plugin path**: `plugins/grafana-wavefront-datasource`
- **Go module**: `github.com/grafana/wavefront-datasource` (`plugins/grafana-wavefront-datasource/go.mod`)

Every value in [`dsconfig.json`](dsconfig.json) — labels, placeholders, tooltips (bound to the
`tooltip` prop of the legacy form fields), section titles, defaults, `requiredWhen` expressions,
storage keys, storage targets, value types, group titles, and instructions — is traceable to a
specific `file:line` in the upstream plugin at this SHA. See [Field provenance](#field-provenance).

To reproduce this research (the monorepo is large; sparse-checkout the plugin path):

```bash
git clone https://github.com/grafana/plugins-private
cd plugins-private
git checkout 267f4937806ed6404b6628d13ae358a5d308e376
cd plugins/grafana-wavefront-datasource
```

If upstream `main` has advanced past this SHA, re-diff the sources listed under
[Sources researched](#sources-researched) before merging any changes to this entry.

## Files

| File | Purpose |
| --- | --- |
| [`dsconfig.json`](dsconfig.json) | dsconfig v1 schema — single source of truth for all config fields, groups, relationships, and instructions |
| [`settings.ts`](settings.ts) | TypeScript models: `RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig` |
| [`settings.go`](settings.go) | Go `Config` model (flat: jsonData fields + `DecryptedSecureJSONData`), `PluginID`, `SecureJsonDataKey` typed constant, and the `LoadConfig` utility (`parse → ApplyDefaults → Validate`) |
| [`schema.go`](schema.go) | k8s-style SDK plugin schema: embeds `dsconfig.json`, converts it via `dsconfig.NewSDKSchema`, and defines `SettingsExamples` for each connection variant |
| [`settings_test.go`](settings_test.go) | Table tests for `LoadConfig`, `ApplyDefaults`, and `Validate` |
| [`conformance_test.go`](conformance_test.go) | Runs `schema.RunPluginTests` — the shared dsconfig conformance suite — against `dsconfig.json` |
| `schema.gen.json`, `settings.gen.json`, `settings.examples.gen.json` | Committed schema artifacts (regenerate with `go generate ./...` in this directory; the `SchemaArtifactInSync` conformance subtest guards drift) |

There is no per-entry `go.mod` — every registry entry is a subpackage of the shared
[`registry/`](..) module (`github.com/grafana/dsconfig/registry`). This entry's package is
`wavefrontdatasource`.

## Sources researched

Every source below was read at the pinned upstream SHA
(`267f4937806ed6404b6628d13ae358a5d308e376`), plus external editor components at the exact versions
the plugin resolves through the monorepo catalog (`.yarnrc.yml`) and `yarn.lock`.

### Plugin (`plugins/grafana-wavefront-datasource`)

| File | What was read |
| --- | --- |
| `src/plugin.json:5` | `pluginType` (`id` = `grafana-wavefront-datasource`) |
| `src/plugin.json:3` | `pluginName` (`name` = `Wavefront`) |
| `src/plugin.json:28` | `docURL` = `https://grafana.com/docs/plugins/grafana-wavefront-datasource` (`info.links[0]`, the `Docs` link) |
| `src/components/ConfigEditor.tsx:16` | API URL input seeded from `jsonData.url \|\| DEFAULT_API_URL` (frontend default source) |
| `src/components/ConfigEditor.tsx:36-38` | `onURLUpdate` writes `jsonData.url` on blur |
| `src/components/ConfigEditor.tsx:39-51` | `onRequestTimeoutUpdate` writes `jsonData.requestTimeout`, or `null` when the value is falsy |
| `src/components/ConfigEditor.tsx:55` | `<h3>Wavefront settings</h3>` section title → group `wavefront-settings` |
| `src/components/ConfigEditor.tsx:57-67` | API URL `LegacyForms.FormField` — label / tooltip / placeholder / aria-label from selectors |
| `src/components/ConfigEditor.tsx:70-82` | Token `LegacyForms.SecretFormField` — writes `secureJsonData.token`, `isConfigured={secureJsonFields['token']}` |
| `src/components/ConfigEditor.tsx:86` | `<h3>Customization</h3>` section title → group `customization` |
| `src/components/ConfigEditor.tsx:88-98` | Request timeout `LegacyForms.FormField type="number"` — writes `jsonData.requestTimeout` |
| `src/components/ConfigEditor.tsx:100-132` | Secure Socks Proxy `InlineSwitch` — writes `jsonData.enableSecureSocksProxy`, feature-flag + version gated (excluded per AGENTS.md) |
| `src/selectors.ts:3` | `DEFAULT_API_URL = 'https://try.wavefront.com'` |
| `src/selectors.ts:7-25` | Every label, aria-label, tooltip, and placeholder the editor renders (ApiUrl, Token, Customization.RequestTimeout) |
| `src/types.ts:73-81` | Frontend types `WavefrontJsonData` (`url`, `requestTimeout?`, `enableSecureSocksProxy?`) and `SecureSettings` (`token`) |
| `pkg/models/settings.go:14-19` | Backend `Settings` struct: `URL json:"url"`, `RequestTimeout int64 json:"requestTimeout"`, `Token`, `ProxyOptions *proxy.Options` |
| `pkg/models/settings.go:22-45` | `LoadSettings`: seed `RequestTimeout = defaultRequestTimeout` (`:23-24`); `json.Unmarshal(config.JSONData)` (`:26-28`); copy `config.DecryptedSecureJSONData["token"]` (`:29-31`); require url (`:32-34`, `"invalid url"`); require token (`:35-37`, `"invalid credentials"`); derive `ProxyOptions` from `config.HTTPClientOptions(ctx)` (`:39-43`) |
| `pkg/models/constant.go:4` | `defaultRequestTimeout = 30` |
| `pkg/models/settings_test.go:13-85` | Confirms: invalid/empty JSONData is an unmarshal error; `{}` and `{"url":""}` → `"invalid url"`; valid url without creds → `"invalid credentials"`; default timeout 30; explicit override; `null` timeout → 30 |
| `pkg/datasource/datasource.go:36-65` | Instance factory: `LoadSettings`; `strings.TrimSuffix(settings.URL, "/")` (`:42`); the commented-out `// url = fmt.Sprintf("%s/api/v2", url)` (`:43`); `Authorization: Bearer <token>` header (`:45-47`); HTTP client + Wavefront client construction with `settings.RequestTimeout` / `settings.ProxyOptions` |
| `pkg/datasource/client.go:19-22` | `getHTTPClient` coerces any timeout `<= 0` to `30` seconds |
| `pkg/datasource/handler_healthcheck.go:100-111` | `CheckSettings` rejects empty URL (`"Enter an API URL."`) then empty token (`"Enter a token."`); health probe hits `api/v2/accesspolicy` (`:20`) |
| `pkg/wavefront/client.go:39-44` | Every outgoing request adds `Authorization: Bearer <token>` |
| `pkg/wavefront/wavefront.go:32-42` | `NewWaveFrontClient`: `apiBaseURL` fallback only when `baseURL == ""` (unreachable — url is required); trims trailing `/` |
| `pkg/wavefront/constants.go:4-5` | `apiBaseURL = "https://longboard.wavefront.com"`, `pluginID = "grafana-wavefront-datasource"` |
| `go.mod:1,9` | Module path `github.com/grafana/wavefront-datasource`; `grafana-plugin-sdk-go v0.279.0` |
| `package.json:33-42` | `@grafana/*` deps use the `catalog:` protocol |

### External editor components

Read at the versions the plugin resolves via the monorepo catalog (`.yarnrc.yml`) and `yarn.lock`.

| Component | Version | Source | What was read |
| --- | --- | --- | --- |
| `LegacyForms.FormField`, `LegacyForms.SecretFormField`, `InlineField`, `InlineSwitch` | `@grafana/ui@11.6.14` (catalog `^11.6.7`) | grafana/grafana `v11.6.14` `packages/grafana-ui/src/components/` | Legacy form field props (`label`, `tooltip`, `placeholder`, `inputWidth`, `labelWidth`, `value`, `onChange`, `onBlur`, `onReset`, `isConfigured`). These render text from `selectors.ts`; they do not write storage keys themselves — the editor's onChange handlers do |
| `config` (runtime config, feature toggles) | `@grafana/runtime@11.6.14` (catalog `^11.6.7`) | grafana/grafana `v11.6.14` `packages/grafana-runtime/src/config.ts` | `config.featureToggles['secureSocksDSProxyEnabled']` + `config.buildInfo.version` gate for the excluded Secure Socks Proxy switch (`ConfigEditor.tsx:100`) |
| `DataSourcePluginOptionsEditorProps`, `DataSourceJsonData` | `@grafana/data@11.6.14` (catalog `^11.6.7`) | grafana/grafana `v11.6.14` `packages/grafana-data/src/types/datasource.ts` | Base interface `WavefrontJsonData` extends; storage semantics of `onOptionsChange` |
| `E2ESelectors` | `@grafana/e2e-selectors@11.6.7` (pinned via workflow `resolutions`, not cataloged) | grafana/grafana `v11.6.7` `packages/grafana-e2e-selectors/src/` | Typing for `src/selectors.ts` |
| `@grafana/plugin-ui@0.13.1` (catalog `^0.13.1`) | 0.13.1 | `github.com/grafana/plugin-ui` tag `v0.13.1` | Listed as a plugin dependency but the ConfigEditor does not import it — no config storage keys contributed |

## Field provenance

Every dsconfig field, traced from its schema `id` to the upstream `file:line` where each of its
label, placeholder, tooltip, default, storage key, and value type is defined.

| Schema `id` | Storage key | Target | Label source | Placeholder / default source | Value type source | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `jsonData_url` | `url` | `jsonData` | `selectors.ts:8` (`"API URL"`) via `LegacyForms.FormField label` at `ConfigEditor.tsx:59` | Placeholder `selectors.ts:11` = `DEFAULT_API_URL` (`selectors.ts:3` = `"https://try.wavefront.com"`); default = same, matching the editor's `useState` init at `ConfigEditor.tsx:16` | `Settings.URL string json:"url"`, `pkg/models/settings.go:15`; TS `url: string`, `types.ts:74` | Description = tooltip `selectors.ts:10` (`"URL to Wavefront API"`); role `endpoint.baseUrl`; `requiredWhen: "true"` — backend returns `"invalid url"` when empty (`settings.go:32-34`). Backend does **not** default it |
| `secureJsonData_token` | `token` | `secureJsonData` | `selectors.ts:14` (`"Token"`) via `LegacyForms.SecretFormField label` at `ConfigEditor.tsx:72` | Placeholder `selectors.ts:17` (`"Wavefront token"`) | `SecureSettings.token: string`, `types.ts:79-81`; consumed via `config.DecryptedSecureJSONData["token"]` at `settings.go:29-31` | Description = tooltip `selectors.ts:16` (`"Wavefront token"`); role `auth.bearer.token` (`datasource.go:45-47`, `wavefront/client.go:43` send `Authorization: Bearer`); `requiredWhen: "true"` — backend returns `"invalid credentials"` when empty (`settings.go:35-37`) |
| `jsonData_requestTimeout` | `requestTimeout` | `jsonData` | `selectors.ts:21` (`"Request timeout in seconds"`) via `LegacyForms.FormField label` at `ConfigEditor.tsx:91` | Placeholder `selectors.ts:24` (`"30"`); default `30` (`constant.go:4`; seeded at `settings.go:23-24`) | `Settings.RequestTimeout int64 json:"requestTimeout"`, `pkg/models/settings.go:16`; TS `requestTimeout?: number`, `types.ts:75` | Description = tooltip `selectors.ts:23` (`"Request timeout in seconds. Defaults to 30"`); role `transport.timeoutSeconds`; optional (defaulted) |

## Field inventory summary

| Schema field | Storage key | Target | Editor label | Read by backend? |
| --- | --- | --- | --- | --- |
| `jsonData_url` | `url` | `jsonData` | API URL | Yes (required; `"invalid url"` when empty) |
| `secureJsonData_token` | `token` | `secureJsonData` | Token | Yes (required, sent as Bearer; `"invalid credentials"` when empty) |
| `jsonData_requestTimeout` | `requestTimeout` | `jsonData` | Request timeout in seconds | Yes (HTTP client timeout; default 30) |
| `jsonData_enableSecureSocksProxy` (excluded) | `enableSecureSocksProxy` | `jsonData` | Enable Secure Socks Proxy | Indirectly (via `config.HTTPClientOptions(ctx)`) — excluded per AGENTS.md |

### Frontend-only settings

None. Every editor-written field (except the excluded Secure Socks Proxy switch) is read by the
backend.

### Backend-only settings

None. Every backend-consumed setting has an editor UI. The upstream `Settings.ProxyOptions` field
is **derived** at load time from `config.HTTPClientOptions(ctx)` (`settings.go:39-43`), not stored,
so it is not a config field and is not modeled on the Go `Config`.

## Where the types are defined

Some fields and base types come from libraries/SDKs rather than the plugin itself.

### Frontend (TypeScript)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `WavefrontJsonData` (jsonData: `url`, `requestTimeout?`, `enableSecureSocksProxy?`), `SecureSettings` (`token`) | `src/types.ts:73-81` | plugin (`grafana-wavefront-datasource`) |
| `DataSourceJsonData` (base interface `WavefrontJsonData` extends) | `packages/grafana-data/src/types/datasource.ts` | `@grafana/data` `11.6.14` |
| `DataSourcePluginOptionsEditorProps` | `packages/grafana-data/src/types/` | `@grafana/data` `11.6.14` |

### Backend (Go)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `Settings` (`URL`, `RequestTimeout`, `Token`, `ProxyOptions`), `LoadSettings` | `pkg/models/settings.go:14-46` | plugin (`grafana-wavefront-datasource`) |
| `defaultRequestTimeout` | `pkg/models/constant.go:4` | plugin |
| `apiBaseURL`, `pluginID` | `pkg/wavefront/constants.go:4-5` | plugin |
| `backend.DataSourceInstanceSettings` (carries `JSONData`, `DecryptedSecureJSONData`, and root `URL`/`BasicAuth*` — root fields unused by this plugin) | `backend/common.go` | `github.com/grafana/grafana-plugin-sdk-go@v0.279.0` |
| `proxy.Options` (type of the derived `ProxyOptions`), `config.HTTPClientOptions(ctx)` | `backend/proxy`, `backend` | `github.com/grafana/grafana-plugin-sdk-go@v0.279.0` |

This entry flattens that spread into a single Go `Config` (jsonData fields + `DecryptedSecureJSONData`)
plus a `SecureJsonDataKey` typed constant. `settings.ts` keeps the three canonical TypeScript types;
`RootConfig` is a blank object because the Wavefront plugin stores nothing at the datasource root.

## Modeling decisions

- **`url` is jsonData, not root.url**. The editor writes `jsonData.url`
  (`ConfigEditor.tsx:36-38`) and the backend unmarshals it via `Settings.URL json:"url"` on the
  jsonData struct (`settings.go:15,26`). The root `settings.URL` is never touched.
- **`RootConfig` is a blank object**. Nothing lives at the datasource root.
- **`requiredWhen` encodes the backend contract**. Both `url` (`"invalid url"`) and `token`
  (`"invalid credentials"`) are `requiredWhen: "true"` because `LoadSettings` hard-fails on empty
  values (`settings.go:32-37`).
- **`url` carries both a `defaultValue` and `requiredWhen`**. The `defaultValue`
  (`https://try.wavefront.com`) is editor parity (the input pre-fill); `requiredWhen: "true"` is the
  backend contract. The Go `ApplyDefaults` deliberately does **not** default `url` — the backend
  errors on an empty url rather than supplying one, so `Validate` enforces it instead.
- **Description = tooltip**. The legacy form fields' `tooltip` prop is the only place descriptions
  surface; the schema copies the tooltip strings verbatim from `selectors.ts:10,16,23`.
- **Token role**. Marked `auth.bearer.token` — `datasource.go:45-47` and `wavefront/client.go:43`
  send `Authorization: Bearer <token>`.
- **Request timeout role**. Marked `transport.timeoutSeconds`; modeled as `int64` in Go to mirror
  the upstream `Settings.RequestTimeout int64` (`settings.go:16`) → schema `valueType: "number"`.
- **`RequestTimeout` default in Go**. `ApplyDefaults` sets `RequestTimeout` to 30 when `<= 0`,
  covering the missing/null/non-positive cases. This matches the tested upstream behavior (default
  30; `null` → 30) and the `getHTTPClient` coercion of `<= 0` to 30 (`client.go:20-22`). See
  [Upstream findings](#upstream-findings) for the one runtime-equivalent divergence on an explicit
  `requestTimeout: 0`.
- **Secure Socks Proxy excluded**. `jsonData.enableSecureSocksProxy` is omitted per AGENTS.md; the
  upstream `Settings` struct never carries it either (it is consumed transparently via
  `config.HTTPClientOptions(ctx)`), so it is also absent from the Go `Config` and json unmarshal
  silently ignores it.
- **`ProxyOptions` not modeled**. The upstream `Settings.ProxyOptions` is an SDK-derived runtime
  value (`config.HTTPClientOptions(ctx)`), not a stored setting, so it is not on `Config`.
- **Groups mirror the two editor `<h3>` sections**: `Wavefront settings` (order 1: url + token) and
  `Customization` (order 2, `optional: true`: requestTimeout).
- **Field ID naming convention**: `<target>_<camelCaseKey>` (`jsonData_url`, `secureJsonData_token`,
  `jsonData_requestTimeout`); `key` keeps the raw storage key.
- **`SecureJsonDataConfig` is a key list**. Secure values are write-only, so the secure type is just
  the array of secret key names (`token`); consumers read `secureJsonFields` to see what is set.

## SDK plugin schema and k8s-style examples (`schema.go`)

`NewSchema()` assembles the `grafana-plugin-sdk-go` `pluginschema.PluginSchema` bundle (the k8s-style
schema Grafana's datasource API server serves as `{apiVersion}.json`) from the embedded
`dsconfig.json`: the `jsonData` fields become the OpenAPI settings `spec`, and `token` becomes a
`secureValues` entry (never part of the spec).

`SettingsExamples()` provides the default configuration plus one example per connection variant. Each
is a full instance-settings object with the plugin configuration under `jsonData` and the write-only
token under `secureJsonData` (obviously-fake `<your-wavefront-api-token>` placeholders — replace with
a real token):

| Example | Connection | `jsonData` | `secureJsonData` |
| --- | --- | --- | --- |
| `""` (default) | Editor pre-fill (`https://try.wavefront.com`), timeout 30 | `url`, `requestTimeout` | `token` (empty — fails validation, as expected for a default) |
| `apiToken` | Hosted cluster (`https://try.wavefront.com`) | `url` | `token` |
| `selfManagedCluster` | Dedicated cluster (`https://mycluster.wavefront.com`) | `url` | `token` |
| `customTimeout` | Hosted cluster, `requestTimeout: 60` | `url`, `requestTimeout` | `token` |

There is no legacy example — the Wavefront plugin has always stored `jsonData.url` +
`secureJsonData.token`; there is no legacy storage shape or alternate auth method.

## `LoadConfig` utility (`settings.go`)

`LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error)` runs
the full three-phase load flow and returns a fully-defaulted, validated `Config`:

1. **Parse** — unconditionally `json.Unmarshal(settings.JSONData, &cfg)` (empty/malformed JSONData is
   a parse error, mirroring `pkg/models/settings.go:26-28`), then copy the decrypted `token` by known
   key into `DecryptedSecureJSONData`.
2. **`ApplyDefaults`** — fill the curated default: `RequestTimeout = 30` when `<= 0`
   (`pkg/models/constant.go:4`, `pkg/models/settings.go:23-24`, `pkg/datasource/client.go:20-22`).
3. **`Validate`** — enforce the runtime contract: `url` and the `token` secret must be non-empty
   (`pkg/models/settings.go:32-37`). Errors are joined so every problem surfaces at once.

Everything is logged via `backend.Logger.FromContext(ctx)` with `datasource_uid`, `datasource_name`,
and `plugin` labels. `(*Config).ApplyDefaults()` and `(Config).Validate() error` are exported
separately for callers that compose them themselves (provisioning preview, schema-example
round-trip, tests distinguishing parse-level from policy-level errors).

## Upstream findings

Potential bugs, misleading UX, and consistency issues found while researching upstream. All are
preserved verbatim in the schema — the schema records what the plugin **does**, not what it **should**
do.

1. **`url` is stored in `jsonData`, not at the datasource root**, despite being a connection URL.
   Most Grafana datasources use the root `settings.URL`; Wavefront reads `Settings.URL json:"url"`
   off the jsonData struct (`settings.go:15,26`). Relevant only to consumers writing provisioning
   payloads.
2. **The backend does NOT default an empty `url`** — it errors with `"invalid url"`
   (`settings.go:32-34`), even though the editor pre-fills `https://try.wavefront.com`
   (`ConfigEditor.tsx:16`). A datasource provisioned without `jsonData.url` fails to load even though
   the editor makes the URL look defaulted.
3. **Empty / absent JSONData is a fatal parse error.** `settings.go:26-28` calls `json.Unmarshal`
   unconditionally, so empty bytes fail with `"unexpected end of JSON input"`
   (`settings_test.go:22-29`). The editor never sends empty JSONData (url is pre-filled), so this is
   only observable via provisioning / API paths.
4. **Redundant request-timeout defaulting with a subtle explicit-`0` case.** The default 30 is
   applied twice: seeded before unmarshal (`settings.go:23-24`) and re-coerced for any value `<= 0`
   in `getHTTPClient` (`client.go:20-22`). Because of the pre-seed, an explicit `requestTimeout: 0`
   leaves `Settings.RequestTimeout == 0`, yet the effective HTTP timeout is still 30. This entry's
   `ApplyDefaults` normalizes `<= 0` to 30 for the `Config` value; the only difference from upstream
   is the stored field value for an explicit `0` (30 here vs. 0 upstream), which is runtime-equivalent
   because of the client coercion, and is untested upstream.
5. **Three different hard-coded "default" Wavefront URLs exist and none is applied on load.** The
   editor uses `DEFAULT_API_URL = "https://try.wavefront.com"` (`selectors.ts:3`); the Wavefront
   client has a fallback `apiBaseURL = "https://longboard.wavefront.com"`
   (`wavefront/constants.go:4`) that is only used when `baseURL == ""` (unreachable, since url is
   required — `wavefront.go:37-39`); and `LoadSettings` applies neither (it errors on empty url).
   The two hard-coded URLs disagree with each other.
6. **Dead code hints at abandoned URL manipulation.** `datasource.go:43` carries a commented-out
   `// url = fmt.Sprintf("%s/api/v2", url)`. The REST client does not append `/api/v2`; the health
   check joins the full path `api/v2/accesspolicy` directly (`handler_healthcheck.go:20`).
7. **Trailing-slash handling is single-level.** `datasource.go:42` and `wavefront.go:40` each trim
   exactly one trailing `/` via `strings.TrimSuffix` before joining API paths, so one trailing slash
   is tolerated but multiple trailing slashes or an embedded base path could still produce malformed
   URLs. The editor placeholder omits the trailing slash; provisioning payloads should too.
8. **The Token placeholder duplicates the tooltip.** Both are literally `"Wavefront token"`
   (`selectors.ts:16-17`). Preserved verbatim.
9. **`enableSecureSocksProxy` is written by the SDK-gated switch but never read by name** in the
   plugin's Go — it is consumed transparently through `config.HTTPClientOptions(ctx)`
   (`settings.go:39-43`). Excluded from this entry per AGENTS.md.

## Validation performed

- `dsconfig.ParseAndResolveSchemaJSON` + `Schema.Validate()` (Go validator in this repo) — passes
  (via the `ConfigSchemaValid` conformance subtest).
- JSON Schema validation against [`dsconfig/schema.json`](../../dsconfig/schema.json) — passes
  (`ajv validate --spec=draft7 --strict=false --all-errors -c ajv-formats`, the exact CI command).
- `go generate ./...` in this directory — regenerates the three `.gen.json` artifacts with no drift.
- `go build ./...`, `go vet ./...`, `gofmt -l .`, `go test ./...` in the `registry/` module — all
  clean (schema round-trip, artifact-in-sync, spec/secure separation, jsonData↔struct parity in both
  directions, secure-key parity, and `LoadConfig` / `ApplyDefaults` / `Validate` table tests).
- `tsc --noEmit --strict` on `settings.ts` — clean.
- The pre-existing `dsconfig` and `schema` workspace modules still build and test cleanly.
