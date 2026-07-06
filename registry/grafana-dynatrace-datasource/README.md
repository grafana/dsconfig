# grafana-dynatrace-datasource

Declarative configuration schema for the Grafana **Dynatrace** datasource plugin
(`grafana-dynatrace-datasource`).

## Upstream researched

- **Monorepo**: `github.com/grafana/plugins-private` (private)
- **Commit SHA**: `267f4937806ed6404b6628d13ae358a5d308e376`
- **Plugin path**: `plugins/grafana-dynatrace-datasource/`
- **Plugin ID**: `grafana-dynatrace-datasource` (from `src/plugin.json:4` `id`)
- **Plugin Go module**: `github.com/grafana/dynatrace-datasource` (`pkg/**`)

Every value in [`dsconfig.json`](dsconfig.json) — labels, placeholders, tooltips (as field
`description`s), option labels/values, section titles, defaults, validations, dependency and
required-when expressions, storage keys, storage targets, value types, group titles, and
instructions — is traceable to a specific `file:line` in the upstream plugin at this SHA. See
[Field provenance](#field-provenance).

To reproduce this research (the source is a monorepo already on disk — do **not** clone):

```bash
git -C <plugins-private-checkout> fetch origin
git -C <plugins-private-checkout> checkout 267f4937806ed6404b6628d13ae358a5d308e376
# then read plugins/grafana-dynatrace-datasource/
```

If upstream `main` has advanced past this SHA, re-diff the sources under
[Sources researched](#sources-researched) before merging changes to this entry.

## Files

| File | Purpose |
| --- | --- |
| [`dsconfig.json`](dsconfig.json) | dsconfig v1 schema — single source of truth for all config fields, groups, relationships, and instructions |
| [`settings.ts`](settings.ts) | TypeScript models: `RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig` |
| [`settings.go`](settings.go) | Go `Config` model (flat: jsonData fields + `DecryptedSecureJSONData`), `PluginID`, `APIType` + `SecureJsonDataKey` typed constants, and the `LoadConfig` utility (parse → `ApplyDefaults` → `Validate`) |
| [`schema.go`](schema.go) | k8s-style SDK plugin schema: embeds `dsconfig.json`, converts it via `dsconfig.NewSDKSchema`, and defines `SettingsExamples` for each connection type / token variant |
| [`settings_test.go`](settings_test.go) | Table tests for `LoadConfig`, `ApplyDefaults`, and `Validate` (incl. a runtime-generated CA cert for the valid-PEM path) |
| [`conformance_test.go`](conformance_test.go) | Runs `schema.RunPluginTests` — the shared dsconfig conformance suite — against `dsconfig.json` |
| [`schema.gen.json`](schema.gen.json), [`settings.gen.json`](settings.gen.json), [`settings.examples.gen.json`](settings.examples.gen.json) | Committed schema artifacts (regenerate with `go generate ./...`; the `SchemaArtifactInSync` conformance subtest guards drift) |

There is no per-entry `go.mod` — every registry entry is a subpackage of the shared
[`registry/`](..) module (`github.com/grafana/dsconfig/registry`). Package name:
`dynatracedatasource`.

## Sources researched

All paths relative to `plugins/grafana-dynatrace-datasource/` at SHA `267f4937`.

### Plugin sources

| File | What was read |
| --- | --- |
| `src/plugin.json:3-4,29` | `pluginName` (`name` = "Dynatrace"), `pluginType` (`id`), `docURL` (`info.links[0].url`) |
| `src/components/config/ConfigEditor.tsx:21-25` | `DynatraceConfigAPITypes` radio options (SaaS Environment/`saas`, Managed Cluster/`managed`, Raw URL/`url`) |
| `src/components/config/ConfigEditor.tsx:65-71` | `apiType` RadioButtonGroup; default `jsonData.apiType || 'saas'` |
| `src/components/config/ConfigEditor.tsx:73-90` | `environmentId` Input: dynamic label (`'URL'` when `apiType==='url'`, else `'Environment ID'`), dynamic tooltip and placeholder |
| `src/components/config/ConfigEditor.tsx:92-105` | `domain` Input, rendered only when `apiType==='managed'`; placeholder `"Your Domain"` |
| `src/components/config/ConfigEditor.tsx:107-124` | `apiToken` SecretInput (label/tooltip/placeholder from selectors) |
| `src/components/config/ConfigEditor.tsx:127-144` | `platformToken` SecretInput (label/tooltip/placeholder from selectors) |
| `src/components/config/ConfigEditor.tsx:147-162` | `httpClientTimeout` number Input, default 30, `min={0}`, placeholder `"30"` |
| `src/components/config/ConfigEditor.tsx:164-169` | `tlsSkipVerify` Checkbox, hard-coded label `"Skip TLS Verify"` |
| `src/components/config/ConfigEditor.tsx:171-182` | `tlsAuthWithCACert` Checkbox (label/tooltip from selectors) |
| `src/components/config/ConfigEditor.tsx:184-200` | `tlsCACert` SecretTextArea, rendered only when `tlsAuthWithCACert`; label `"CA Cert"`, tooltip, placeholder `"Begins with -----BEGIN CERTIFICATE-----"`, `rows={5}` |
| `src/components/config/ConfigEditor.tsx:202-231` | Secure Socks Proxy Checkbox (`jsonData.enableSecureSocksProxy`) — deliberately excluded |
| `src/selectors.ts:8-51` | Static labels/tooltips/placeholders: APIType label, Managed Domain label, APIToken/PlatformToken label+tooltip+placeholder, Timeout label+tooltip, CACert `With CA Cert` label+tooltip |
| `src/types.ts:6-20` | `DynatraceConfigAPIType`, `DynatraceDataSourceOptions` (jsonData), `DynatraceDataSourceSecureOptions` (secrets) |
| `pkg/models/settings.go:16-37` | API-type constants and the `Settings` struct fields + json tags |
| `pkg/models/settings.go:40-62` | `LoadSettings`: unmarshal jsonData, copy `apiToken`/`platformToken`/`tlsCACert` secrets, default `httpClientTimeout` to 30 when `<= 0`, load SDK proxy options |
| `pkg/models/settings.go:64-77` | `Settings.Validate`: token presence (`invalid API Token or Platform Token`) and CA-cert PEM checks (`invalid TLS certificate` / `failed to parse TLS CA PEM certificate`) |
| `pkg/dynatrace/client/rest.go:24-53` | `saasURL`/`managedURL`/`rawURL` templates and `GetHostURL` (how `apiType`/`environmentId`/`domain` build the base URL) |
| `pkg/dynatrace/client/rest.go:62-68` | SaaS Grail host switch `.live.dynatrace.com` → `.apps.dynatrace.com` and `/api/` prefix drop for `platform/` paths |
| `pkg/dynatrace/client/rest.go:158-207` | `getHTTPClient`: `Api-Token <apiToken>` for the `api` client vs `Bearer <platformToken>` for the `platform` client, TLS options, timeout, proxy |
| `pkg/dynatrace/handler_healthcheck.go:16-25,141-159` | Health-check error strings and `CheckSettings`: environmentId required, domain required when managed, at least one token, then `Settings.Validate()` |
| `pkg/dynatrace/handler_healthcheck.go:37-60` | `CheckHealth`: runs Metrics health when `apiToken` set and Grail health when `platformToken` set |
| `pkg/dynatrace/datasource.go:54-103` | `GetInstance`: `LoadSettings` then builds classic (`NewClient`/`NewClientAPI`) and platform (`NewPlatformClient`) clients |
| `pkg/models/settings_test.go:14-188` | Confirms timeout defaulting (null/negative → 30, 260 preserved) and `Validate` cases |
| `package.json:30-43` | `@grafana/*` dependency versions via `catalog:` |

### External editor components

Read at the versions pinned by the workspace catalog (`.yarnrc.yml:14-26`, referenced via
`catalog:` in the plugin's `package.json:30-43`). The config editor composes **only**
`@grafana/ui` primitives (no `@grafana/plugin-ui` sections).

| Component / type | Version | Source | What was read |
| --- | --- | --- | --- |
| `InlineField`, `RadioButtonGroup`, `Checkbox`, `Input`, `SecretInput`, `SecretTextArea`, `InlineFormLabel` | `@grafana/ui@^11.6.7` | grafana/grafana `packages/grafana-ui/src/components/` | Prop names (`label`, `tooltip`, `placeholder`, `value`, `onChange`, `isConfigured`, `onReset`, `rows`, `cols`) so we knew which UI attributes to record |
| `DataSourcePluginOptionsEditorProps`, `SelectableValue`, `DataSourceJsonData` | `@grafana/data@^11.6.7` | grafana/grafana `packages/grafana-data/src/` | Editor props shape; `DataSourceJsonData` is the base interface `DynatraceDataSourceOptions` extends |
| `config` | `@grafana/runtime@^11.6.7` | grafana/grafana `packages/grafana-runtime/src/` | `featureToggles.secureSocksDSProxyEnabled` + `buildInfo.version` gate for the (excluded) Secure Socks Proxy switch (`ConfigEditor.tsx:202`) |
| `E2ESelectors` | `@grafana/e2e-selectors` (intentionally **not** cataloged — swapped per Grafana version) | grafana/grafana `packages/grafana-e2e-selectors/src/` | Type backing `src/selectors.ts` (the label/tooltip/placeholder source map) |

## Field provenance

| Schema `id` | Storage key | Target | Label source | Placeholder / options / default source | Value type source | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `jsonData_apiType` | `apiType` | `jsonData` | `selectors.ts:11` (`'Dynatrace API Type'`) | Options `ConfigEditor.tsx:21-25`; default `saas` from `ConfigEditor.tsx:68` (`jsonData.apiType \|\| 'saas'`) | `DynatraceConfigAPIType` union `types.ts:6`; `Settings.APIType string` `settings.go:24` | No editor tooltip → no `description` |
| `jsonData_environmentId` | `environmentId` | `jsonData` | `ConfigEditor.tsx:75` (`'Environment ID'`; relabeled `'URL'` when `apiType==='url'`) | Placeholder `ConfigEditor.tsx:88`; tooltip `ConfigEditor.tsx:77-81` (used as `description`) | `Settings.EnvironmentID string` `settings.go:25`; TS `string` `types.ts:9` | `requiredWhen: "true"` (health check `handler_healthcheck.go:142-144`); `overrides` for url-mode description/placeholder; role `endpoint.baseUrl` |
| `jsonData_domain` | `domain` | `jsonData` | `selectors.ts:13` (`'Domain'`) | Placeholder `ConfigEditor.tsx:100` (`'Your Domain'`) | `Settings.Domain string` `settings.go:26`; TS `string` `types.ts:10` | `dependsOn`/`requiredWhen` from conditional render `ConfigEditor.tsx:92` + `handler_healthcheck.go:147-149`; role `endpoint.domain` |
| `secureJsonData_apiToken` | `apiToken` | `secureJsonData` | `selectors.ts:20` (`'Dynatrace API Token'`) | Placeholder `selectors.ts:19`; tooltip `selectors.ts:22` (used as `description`) | `DynatraceDataSourceSecureOptions.apiToken` `types.ts:17`; copied `settings.go:46` | Sent as `Api-Token` header `rest.go:171-172`; role `auth.apiKey.value`; `requiredWhen` = platformToken empty |
| `secureJsonData_platformToken` | `platformToken` | `secureJsonData` | `selectors.ts:28` (`'Dynatrace Platform Token'`) | Placeholder `selectors.ts:27`; tooltip `selectors.ts:30` (used as `description`) | `DynatraceDataSourceSecureOptions.platformToken` `types.ts:18`; copied `settings.go:47` | Sent as `Bearer` header `rest.go:173-174`; role `auth.bearer.token`; `requiredWhen` = apiToken empty |
| `jsonData_httpClientTimeout` | `httpClientTimeout` | `jsonData` | `selectors.ts:39` (`'Timeout'`) | Placeholder `ConfigEditor.tsx:158` (`'30'`); default 30 `ConfigEditor.tsx:156` / `settings.go:50-53`; `min={0}` `ConfigEditor.tsx:159`; tooltip `selectors.ts:40` | `Settings.HttpClientTimeout int` `settings.go:32`; TS `number` `types.ts:14` | Role `transport.timeoutSeconds`; `range` min 0 |
| `jsonData_tlsSkipVerify` | `tlsSkipVerify` | `jsonData` | `ConfigEditor.tsx:164` (`'Skip TLS Verify'`) | Default `false` (`Settings` zero value) | `Settings.SkipTLSVerify bool` `settings.go:27`; TS `boolean` `types.ts:11` | No editor tooltip → no `description`; role `transport.tlsSkipVerify` |
| `jsonData_tlsAuthWithCACert` | `tlsAuthWithCACert` | `jsonData` | `selectors.ts:45` (`'With CA Cert'`) | Tooltip `selectors.ts:47` (used as `description`); default `false` | `Settings.TLSAuthWithCACert bool` `settings.go:28`; TS `boolean` `types.ts:12` | No matching role |
| `secureJsonData_tlsCACert` | `tlsCACert` | `secureJsonData` | `ConfigEditor.tsx:187` (`'CA Cert'`) | Placeholder `ConfigEditor.tsx:193`; tooltip `ConfigEditor.tsx:186` (used as `description`); `rows={5}` `ConfigEditor.tsx:194` | `DynatraceDataSourceSecureOptions.tlsCACert` `types.ts:19`; copied `settings.go:48` | `dependsOn`/`requiredWhen` from conditional render `ConfigEditor.tsx:184` + `settings.go:68-76`; role `tls.caCert`; textarea |

## Field inventory summary

| Schema field | Storage key | Target | Editor label | Read by backend? |
| --- | --- | --- | --- | --- |
| `jsonData_apiType` | `apiType` | `jsonData` | Dynatrace API Type | Yes (URL construction) |
| `jsonData_environmentId` | `environmentId` | `jsonData` | Environment ID / URL | Yes |
| `jsonData_domain` | `domain` | `jsonData` | Domain | Yes (managed only) |
| `secureJsonData_apiToken` | `apiToken` | `secureJsonData` | Dynatrace API Token | Yes |
| `secureJsonData_platformToken` | `platformToken` | `secureJsonData` | Dynatrace Platform Token | Yes |
| `jsonData_httpClientTimeout` | `httpClientTimeout` | `jsonData` | Timeout | Yes |
| `jsonData_tlsSkipVerify` | `tlsSkipVerify` | `jsonData` | Skip TLS Verify | Yes |
| `jsonData_tlsAuthWithCACert` | `tlsAuthWithCACert` | `jsonData` | With CA Cert | Yes |
| `secureJsonData_tlsCACert` | `tlsCACert` | `secureJsonData` | CA Cert | Yes (when `tlsAuthWithCACert`) |

### Frontend-only settings

None. Every editor-visible field is read by the backend.

### Backend-only settings

None modeled. The backend `Settings` struct also declares `SdkProxyOptions` (`settings.go:35`,
loaded from the SDK's `HTTPClientOptions`, `json:"-"`) and `Inputs []models.Input`
(`settings.go:36`) — both are runtime/dead rather than editor configuration and are not modeled
(see [Upstream findings](#upstream-findings) #1).

### Excluded

- **`enableSecureSocksProxy`** (`jsonData`, `ConfigEditor.tsx:202-231`) — the Secure Socks Proxy
  switch, excluded from every registry entry per AGENTS.md. It is written by the editor and
  consumed transparently by the SDK's `config.HTTPClientOptions(ctx)` (`settings.go:55-59`); the
  plugin's own `Settings` struct does not carry it.

## Where the types are defined

Only config type/field definitions are listed (UI components and functions/helpers such as
`LoadSettings`, `GetHostURL`, `getHTTPClient` are omitted even where they are the reason a field
exists).

### Frontend (TypeScript)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `DynatraceDataSourceOptions` (jsonData: `apiType`, `environmentId`, `domain`, `tlsSkipVerify`, `tlsAuthWithCACert`, `enableSecureSocksProxy`, `httpClientTimeout`), `DynatraceDataSourceSecureOptions` (`apiToken`, `platformToken`, `tlsCACert`), `DynatraceConfigAPIType` | `src/types.ts:6-20` | plugin (`grafana-dynatrace-datasource`) |
| `DataSourceJsonData` (base interface `DynatraceDataSourceOptions` extends: `authType`, `defaultRegion`, `manageAlerts`, …) | `packages/grafana-data/src/types/datasource.ts` | `@grafana/data` `^11.6.7` |

### Backend (Go)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `Settings` (jsonData fields + `APIToken`/`PlatformToken`/`TlsCACert` secrets), `SettingsAPIType{Saas,Managed,URL}` | `pkg/models/settings.go:16-37` | plugin (`github.com/grafana/dynatrace-datasource`) |
| `backend.DataSourceInstanceSettings` (carries `JSONData`, `DecryptedSecureJSONData`, and root `URL`/`BasicAuth*` — unused by this plugin) | `backend/common.go` | `github.com/grafana/grafana-plugin-sdk-go` `v0.292.1` |
| `proxy.Options` (`Settings.SdkProxyOptions`) | `backend/proxy` | `github.com/grafana/grafana-plugin-sdk-go` |
| `httpclient.TLSOptions` / `httpclient.TimeoutOptions` (built from `SkipTLSVerify`/`TlsCACert`/`HttpClientTimeout`) | `backend/httpclient` | `github.com/grafana/grafana-plugin-sdk-go` |
| `models.Input` (`Settings.Inputs`, dead on this struct) | `api-query/pkg/models` | `github.com/grafana/plugins-private` (`@grafana/plugins-private-api-query`) |

This entry flattens the backend spread into a single Go `Config` (jsonData fields +
`DecryptedSecureJSONData`) plus an `APIType` typed alias and a `SecureJsonDataKey` typed constant
list. `settings.ts` keeps the three canonical TypeScript types.

## Modeling decisions

- **Two independent tokens, no discriminator**: the editor always shows both `apiToken` and
  `platformToken` (there is no auth-method radio). `apiToken` authenticates the classic API
  (`Api-Token` header, `rest.go:171-172`); `platformToken` authenticates Grail (`Bearer` header,
  `rest.go:173-174`). At least one is required (`handler_healthcheck.go:151-153`). This is modeled
  as two `secureJsonData` fields, each with `requiredWhen` referencing the sibling secret
  (`secureJsonData_apiToken == ''` / `secureJsonData_platformToken == ''`) to encode the
  at-least-one contract, plus a `group` relationship describing it. There is no `alternative`
  relationship type in the vocabulary, so `group` is used.
- **`apiType` is a connection discriminator, not an auth one**: no `auth.discriminator` role is
  applied (that role is auth-specific); the three-way URL construction is captured in a `group`
  relationship and an instruction instead.
- **Overloaded `environmentId`**: its meaning changes by `apiType` (SaaS tenant ID / Managed
  environment ID / full base URL) and the editor relabels it to `'URL'` in url mode
  (`ConfigEditor.tsx:75`). `FieldOverride` cannot change a label, so the base label stays
  `Environment ID`; the url-mode `description` and `placeholder` are captured via `overrides`, and
  the label change is called out in an instruction. Role `endpoint.baseUrl` (the dominant
  semantic across all modes).
- **`requiredWhen` vs the editor**: the editor marks nothing required, but the backend health
  check hard-fails without `environmentId`, without a token, and (managed) without `domain`. The
  `requiredWhen` rules encode that runtime contract.
- **CA cert pairing**: `secureJsonData_tlsCACert` is `dependsOn`/`requiredWhen`
  `jsonData_tlsAuthWithCACert == true` (editor conditional render `ConfigEditor.tsx:184` +
  backend `settings.go:68-70`), captured as a `pair` relationship.
- **Flat `Config` in Go**: mirrors the json-tagged fields of the upstream `Settings`
  (`settings.go:23-37`) verbatim; secrets become `DecryptedSecureJSONData`. Root fields are not
  carried (the backend never reads them). The runtime-only `SdkProxyOptions` and the dead `Inputs`
  field are omitted (see below).
- **`LoadConfig` = parse → `ApplyDefaults` → `Validate`**: upstream splits this across
  `LoadSettings` (parse + timeout default) and the health check's `CheckSettings` +
  `Settings.Validate` (validation). This entry folds the full contract into one `Validate` and
  moves the `apiType`→`saas` and `httpClientTimeout`→`30` defaults into `ApplyDefaults`. Parsing
  mirrors upstream exactly (unconditional `json.Unmarshal`, so empty `JSONData` is a parse error).
- **Field ID naming**: `<target>_<camelCaseKey>` (`jsonData_`, `secureJsonData_`); `key` keeps the
  raw storage key.

## Settings examples matrix (`schema.go`)

Each example is a full instance-settings object (`jsonData` + write-only `secureJsonData`
placeholders). Secrets use obviously-fake angle-bracket placeholders.

| Example | apiType | Tokens | Notes |
| --- | --- | --- | --- |
| `""` (default) | `saas` | `apiToken` (empty) | Schema defaults; shows what must be filled in |
| `saasApiToken` | `saas` | `apiToken` | Classic API endpoints |
| `saasPlatformToken` | `saas` | `platformToken` | Grail platform API |
| `saasBothTokens` | `saas` | `apiToken` + `platformToken` | Classic + Grail together |
| `managed` | `managed` | `apiToken` | Requires `domain` |
| `rawUrl` | `url` | `apiToken` | `environmentId` is the full base URL |
| `tlsCACert` | `saas` | `apiToken` + `tlsCACert` | Custom CA (placeholder PEM) |

The `""` and `tlsCACert` examples intentionally fail `LoadConfig` validation (empty token
placeholder / non-parseable PEM placeholder respectively) — `settings_test.go` asserts this. A
runtime-generated self-signed certificate exercises the valid-PEM path.

## `LoadConfig` utility (`settings.go`)

`LoadConfig(ctx, settings) (Config, error)` runs the full three-phase flow and returns a
fully-defaulted, validated `Config`:

1. **Parse** — unconditional `json.Unmarshal` of `settings.JSONData` (mirrors
   `LoadSettings`, `settings.go:42-44`; empty bytes are a parse error), then copy the decrypted
   `apiToken`/`platformToken`/`tlsCACert` secrets by known key.
2. **`ApplyDefaults`** — `apiType`→`saas` (editor parity, `ConfigEditor.tsx:68`) and
   `httpClientTimeout`→`30` when `<= 0` (`settings.go:50-53`).
3. **`Validate`** — folds the health check `CheckSettings` (`handler_healthcheck.go:141-159`) with
   `Settings.Validate` (`settings.go:64-77`): environmentId required, domain required when managed,
   at least one token, and (when `tlsAuthWithCACert`) a present, parseable PEM CA certificate.
   Errors are joined.

Everything is logged via `backend.Logger.FromContext(ctx)` with `datasource_uid`,
`datasource_name`, and `plugin` labels. `ApplyDefaults` and `Validate` stay exported for callers
that assemble a `Config` directly.

## Upstream findings

Potential bugs / inconsistencies discovered while researching upstream. The schema records what
the plugin **does**, not what it should do; these notes let reviewers reproduce each finding.

1. **`Settings.Inputs` is a dead field.** `pkg/models/settings.go:36` declares
   `Inputs []models.Input` with no json tag (so it would bind to a JSON key `Inputs`), but
   `LoadSettings` never populates it and no code reads `settings.Inputs` — `Inputs` is only used on
   the separate query model (`pkg/models/query.go`). It is omitted from this entry's `Config`
   (carrying it would force a spurious `Inputs` jsonData field).
2. **Dead / misleading health-check error strings.** `NoDomainError = "Enter a domain."`
   (`handler_healthcheck.go:23`) and `DatasourceError` (`:17`) are unused in that file; the
   managed-mode missing-domain branch (`:147-149`) instead returns `NoUrlError =
   "Enter a domain URL."` — a domain field reported with a "URL" message.
3. **`environmentId` is overloaded and relabeled.** The same stored key is a tenant ID (`saas`),
   an environment ID under a cluster (`managed`), or a full base URL (`url`), and the editor
   relabels the field `'URL'` only in url mode (`ConfigEditor.tsx:75`). Consumers must interpret it
   by `apiType`.
4. **Validation is deferred past instance creation.** `GetInstance`/`LoadSettings`
   (`datasource.go:54-58`) never call `Validate`; a datasource with no token can be instantiated
   and only fails at health check, and `getHTTPClient` returns
   `"no auth token configured for platformType=%s"` (`rest.go:177-179`) lazily when a query needs a
   token family that is not configured (e.g. a Grail query with only `apiToken` set).
5. **Trailing-slash / scheme pitfalls in `url` mode.** `rawURL = "%s/api/%s%s"` (`rest.go:27`)
   concatenates onto `environmentId` verbatim, so a trailing slash yields `…//api/…`, and a
   missing scheme yields an unparseable/relative URL.
6. **SaaS-only Grail host rewrite.** For `platform/` paths the client rewrites
   `.live.dynatrace.com` → `.apps.dynatrace.com` and strips `/api/` (`rest.go:62-68`); this rewrite
   only matches SaaS hosts, so Grail on `managed`/`url` relies on the same host serving platform
   APIs.
7. **`plugin.json` "Repository" link points to `grafana/grafana`** (`src/plugin.json:31`), not the
   plugin's own repository — cosmetic.

## Validation performed

- `dsconfig.ParseAndResolveSchemaJSON` + `Schema.Validate()` (Go validator in this repo) — passes
  (via the conformance suite).
- JSON Schema validation against [`dsconfig/schema.json`](../../dsconfig/schema.json) (draft-07,
  strict `additionalProperties: false`) — passes.
- `go generate ./...` (regenerates artifacts), then from `registry/`:
  `gofmt -l .` (clean), `go vet ./...` (clean), `go build ./...` (clean), `go test ./...` (passes,
  including `SchemaArtifactInSync`, spec/secure separation, jsonData/struct parity, secure-key
  parity, and `LoadConfig`/`ApplyDefaults`/`Validate`).
- `settings.ts`: `tsc --noEmit --strict` (typescript 5.5.4) — clean.
- Sibling `dsconfig` and `schema` workspace modules still build and test.
