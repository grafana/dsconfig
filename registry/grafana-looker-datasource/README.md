# grafana-looker-datasource

dsconfig registry entry for the **Grafana Looker datasource plugin** (`grafana-looker-datasource`).

Looker is a Grafana Enterprise datasource that connects to a Looker instance's API3 endpoint using
API3 credentials (a client ID + client secret) exchanged for an access token by the Looker Go SDK.

## Files

| File | Purpose |
| --- | --- |
| `dsconfig.json` | dsconfig v1 schema — the single source of truth for the Looker configuration |
| `settings.ts` | TypeScript models: `RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig` |
| `settings.go` | Flat Go `Config` (jsonData + `DecryptedSecureJSONData`), typed enums, `SecureJsonDataKey`, `LoadConfig`, `ApplyDefaults`, `Validate` |
| `schema.go` | k8s-style SDK `PluginSchema`: embeds `dsconfig.json`, exposes `ConfigSchema()`/`NewSchema()`/`SettingsExamples()` |
| `conformance_test.go` | Wraps `schema.RunPluginTests` (artifact generation + guard-rail suite) |
| `settings_test.go` | Table tests for `LoadConfig`, `ApplyDefaults`, `Validate`, and the examples set |
| `schema.gen.json` / `settings.gen.json` / `settings.examples.gen.json` | Generated artifacts (via `go generate ./...`) |
| `README.md` | This file |

Import path: `github.com/grafana/dsconfig/registry/grafana-looker-datasource` (package `lookerdatasource`).

## Sources researched

Researched against the **`github.com/grafana/plugins-private` monorepo at commit
`267f4937806ed6404b6628d13ae358a5d308e376`**, plugin path `plugins/grafana-looker-datasource/`
(plugin `package.json` version `0.4.18`).

| Source | What it provided |
| --- | --- |
| `src/plugin.json:3-4,24` | Plugin id `grafana-looker-datasource`, name `Looker`, docs URL |
| `src/editors/configEditor.tsx:1-103` | The config editor: field order, components, visibility conditions, onBlur/onChange wiring |
| `src/selectors.ts:1-29` | Every field label, tooltip, and placeholder |
| `src/constants.ts:4-6` | `authOptions` — a single `{ value: 'client_secret', label: 'Client Secret' }` |
| `src/types.ts:12-24` | Frontend `AuthenticationType`, `Config`, `SecureKey`, `SecureConfig` |
| `src/editors/betaNotice.tsx:1-18` | Public-preview info alert (no config fields) |
| `src/datasource.ts:19-23` | `DataSourceWithBackend<Query, Config>` — confirms jsonData shape, no extra config fields |
| `pkg/models/config.go:14-74` | Backend `Config`, `AuthType`, `Validate`, `ApplyDefaults`, `LoadConfig` |
| `pkg/handler_healthcheck.go:12-47` | `CheckHealth` runs `Config.Validate()` and calls `AllLookmlModels` |
| `pkg/looker/client.go:10-39` | `NewClient(baseUrl, clientId, clientSecret)` → `rtl.ApiSettings{ApiVersion:"4.0"}` |
| `pkg/main.go:14-42` | Instance construction: `LoadConfig` → `NewClient(config.BaseURL, config.ClientId, config.ClientSecret)` |

### Library / SDK versions

Frontend `@grafana/*` are cataloged (`.yarnrc.yml:14-26`) and referenced via `catalog:` in the
plugin's `package.json:34-42`:

- `@grafana/data` `^11.6.7`, `@grafana/runtime` `^11.6.7`, `@grafana/schema` `^11.6.7`,
  `@grafana/ui` `^11.6.7` (all resolve to `11.6.14` in the workspace `yarn.lock`).
- `@looker/sdk` `^24.18.1` (plugin-local pin; frontend query/variable types only — not config).
- Package manager `yarn@4.15.0`, `node >=18`.

Backend (`plugins/grafana-looker-datasource/go.mod`):

- `github.com/grafana/grafana-plugin-sdk-go` `v0.279.0` (supplies `backend`, `backend/httpclient`,
  `experimental/errorsource`).
- `github.com/looker-open-source/sdk-codegen/go` `v0.0.2` (Looker Go SDK: `rtl`, `sdk/v4`).

`@grafana/ui@^11.6.7` components consulted for storage keys: `Input`, `RadioButtonGroup`,
`InlineField`, `SecretInput`, `Alert`. `SecretInput` renders the write-only secret and its reset
affordance; it writes only `secureJsonData.client_secret` (via
`onUpdateDatasourceSecureJsonDataOption`/`updateDatasourcePluginResetOption` from `@grafana/data`).

## Field inventory

| Schema ID | Storage key | Target | Editor label | Read by backend |
| --- | --- | --- | --- | --- |
| `jsonData_baseUrl` | `base_url` | `jsonData` | Looker URL | Yes — `config.go:15`; required `:30`; used by `client.go:24` |
| `jsonData_authType` | `auth_type` | `jsonData` | Authentication type | Yes — `config.go:16`; branched in `Validate :33`; defaulted in `ApplyDefaults :48` |
| `jsonData_clientId` | `client_id` | `jsonData` | Looker Client ID | Yes — `config.go:17`; required `:34`; passed to `client.go:26` |
| `secureJsonData_clientSecret` | `client_secret` | `secureJsonData` | Looker Client Secret | Yes — `config.go:69`; required `:37`; passed to `client.go:27` |

**Labels/tooltips/placeholders** (verbatim from `src/selectors.ts`):

- `base_url` — label `Looker URL` (`:4`), tooltip → `description` `Looker base URL. Example: https://00001234-1234-1ab2-1234-a1b2c3d4.looker.app` (`:5`), placeholder `https://xxxxx.looker.app` (`:6`).
- `auth_type` — label `Authentication type` (`:11`), tooltip → `description` `Looker authentication type` (`:12`); option label `Client Secret` / value `client_secret` (`constants.ts:5`).
- `client_id` — label `Looker Client ID` (`:17`), tooltip → `description` `API credentials Looker client id` (`:18`), placeholder `Client ID` (`:19`).
- `client_secret` — label `Looker Client Secret` (`:24`), tooltip → `description` `API credentials Looker client secret` (`:25`), placeholder `Looker Client secret` (`:26`).

### Frontend-only settings

None. Every stored field is read by the backend.

### Backend-only settings

None. The backend `Config` also has `ClientSecret` and `HttpClientOptions`, but both are `json:"-"`
(not jsonData): `ClientSecret` is the decrypted secret (modeled here via `DecryptedSecureJSONData`)
and `HttpClientOptions` is a computed value, never persisted (and never actually applied — see
discrepancies).

## Modeling decisions

- **`RootConfig` is a blank object.** The backend reads `jsonData.base_url` (`config.go:15`), never
  the root `url`, and reads no other named root fields (`config.go:58-74`). Per AGENTS.md the root
  config is `Record<string, never>`, not null.
- **`auth_type` is modeled as a stored discriminator with a single allowed value.** It carries
  `role: auth.discriminator`, `defaultValue: "client_secret"`, and an `allowedValues: ["client_secret"]`
  validation. It is placed in the `authentication` group to reflect its intended editor position, even
  though the editor **never renders the selector** (see discrepancies) — the value is stored and read
  by the backend, so omitting it would break jsonData/struct parity.
- **Roles.** `base_url` → `endpoint.baseUrl`; `auth_type` → `auth.discriminator`; `client_id` →
  `auth.oauth2.clientId`; `client_secret` → `auth.oauth2.clientSecret`. Looker "API3 credentials" are
  a client-id/client-secret pair that the Looker SDK's `rtl.AuthSession` exchanges for an access token
  (an OAuth2 client-credentials flow), so the `auth.oauth2.*` roles are the closest matches in the
  closed role vocabulary.
- **`dependsOn` mirrors editor visibility; `requiredWhen` mirrors the backend contract.** `client_id`
  and `client_secret` use `dependsOn: "jsonData_authType == 'client_secret' || jsonData_authType == ''"`
  (the editor's `auth_type === 'client_secret' || !auth_type`, `configEditor.tsx:65`) and
  `requiredWhen: "jsonData_authType == 'client_secret'"` (`config.go:33-40`). `base_url` uses
  `requiredWhen: "true"` (`config.go:30`, always required).
- **`Config` drops the upstream `ClientSecret`/`HttpClientOptions` fields.** Following the
  github/datadog gold-standard entries, the decrypted secret lives in
  `DecryptedSecureJSONData map[SecureJsonDataKey]string` and the unused (and discarded) HTTP options
  are omitted. The three json-tagged fields (`base_url`, `auth_type`, `client_id`) exactly match the
  schema's jsonData fields, satisfying the conformance walker.
- **`LoadConfig` composes `parse → ApplyDefaults → Validate`** (AGENTS.md contract). The upstream
  splits this: `pkg/models/config.go` `LoadConfig` runs only `ApplyDefaults`, and `Validate` runs
  later in `CheckHealth` (`pkg/handler_healthcheck.go:13`). `ApplyDefaults` mirrors upstream verbatim
  — defaults `auth_type` to `client_secret` and trims `base_url` (incl. a trailing slash), `client_id`,
  and the decrypted `client_secret`. `Validate` reuses the upstream error strings verbatim.
- **Secure Socks Proxy exclusion is not applicable** — the plugin has no `enableSecureSocksProxy`
  field (its HTTP options are never wired into the Looker SDK at all).
- **Two examples only.** There is a single auth type and a single connection style, so the example set
  is the `""` default plus one realistic `clientSecret` example. There is no legacy storage shape, so
  no legacy example.

## Where the types are defined

**Frontend (config types only):**

- `src/types.ts:13` — `AuthenticationType` (`'client_secret'`).
- `src/types.ts:14-18` — `Config` (`base_url`, `auth_type`, `client_id`), extending
  `DataSourceJsonData`.
- `src/types.ts:22-23` — `SecureKey` (`'client_secret'`), `SecureConfig`.
- `DataSourceJsonData` — from `@grafana/data@^11.6.7` (base type of the frontend `Config`).

**Backend (config types only):**

- `pkg/models/config.go:14-20` — `Config` (`base_url`, `auth_type`, `client_id`; plus `json:"-"`
  `ClientSecret` and `HttpClientOptions`).
- `pkg/models/config.go:22-26` — `AuthType` + `AuthTypeClientSecret`.
- `httpclient.Options` (the type of the unused `HttpClientOptions` field) — from
  `github.com/grafana/grafana-plugin-sdk-go@v0.279.0` `backend/httpclient`.
- `rtl.ApiSettings` (the credentials' consumer, `pkg/looker/client.go:23`) — from
  `github.com/looker-open-source/sdk-codegen/go@v0.0.2`.

## Settings examples matrix

| Example key | Summary | `auth_type` | `base_url` | `secureJsonData` |
| --- | --- | --- | --- | --- |
| `""` | Default configuration | `client_secret` | `""` | `client_secret: ""` |
| `clientSecret` | Client Secret (Looker API3 credentials) | `client_secret` | `https://your-instance.looker.app` | `client_secret: <your-client-secret>` (`client_id: <your-client-id>`) |

All secret values use obviously-fake angle-bracket placeholders (`<your-client-id>`,
`<your-client-secret>`) — no realistic token shapes. The `""` example intentionally has empty values,
so `LoadConfig`'s `Validate` step rejects it with `invalid/empty Looker base url` (asserted in
`settings_test.go`).

## Potential upstream bugs / discrepancies

1. **The "Authentication type" selector is never rendered.** `authOptions` has exactly one entry
   (`src/constants.ts:4-6`), and the radio only renders when `authOptions.length > 1`
   (`src/editors/configEditor.tsx:50`). So the `auth_type` label/tooltip in `src/selectors.ts:9-14`
   are dead UI today. The field is still stored and read by the backend, so it is modeled here.
2. **The computed HTTP client options are discarded.** `pkg/models/config.go:70-71` computes
   `httpClientOptions` (and appends an error-source middleware) but never assigns it to
   `config.HttpClientOptions`, and `looker.NewClient` (`pkg/looker/client.go:22-38`) builds its own
   `rtl.AuthSession` without them. Net effect: Grafana's standard datasource HTTP settings (TLS
   skip-verify, timeout, proxy) are **not** applied to Looker API calls. (This is also why this plugin
   has no "verify SSL" / "timeout" / "port" config fields — the base URL carries scheme+host and the
   Looker SDK manages its own transport.)
3. **`config.go:70` joins the `HTTPClientOptions` error into the return value** even though the result
   is unused, so an `HTTPClientOptions(ctx)` failure would fail datasource load for a value that has no
   effect.
4. **`LoadConfig` does not validate.** Upstream `LoadConfig` (`pkg/models/config.go:58-74`) calls only
   `ApplyDefaults`; `Validate` runs separately in `CheckHealth`. Consequently `New` (`pkg/main.go:28-41`)
   can construct a `DataSource` from an invalid config; the error only surfaces on a health check. This
   entry's `LoadConfig` adds `Validate` per AGENTS.md.
5. **Whitespace-only secret is treated as empty.** `ApplyDefaults` trims the decrypted `client_secret`
   (`config.go:54`), so a whitespace-only value fails `Validate` with `invalid/empty Looker client secret`.
6. **`base_url`/`client_id` persist only on blur.** The editor keeps them in local React state and
   writes to jsonData in `onBlur` (`configEditor.tsx:22-23,44,77`). Programmatic/provisioned configs are
   unaffected, but a user who edits and navigates away without blurring may not persist the change.

## Validation performed

- `go generate ./...` in the entry dir — regenerated `schema.gen.json` / `settings.gen.json` /
  `settings.examples.gen.json`.
- `gofmt -l .` (clean), `go vet ./...` (clean), `go test ./...` — all pass across the `registry` module,
  including this entry's `TestSchemaConformance` (all 8 guard-rail subtests: `BaseFieldsResolved`,
  `SchemaRoundTrip`, `SchemaArtifactInSync`, `SchemaSpecHasNoSecureJSON`, `ConfigSchemaValid`,
  `JSONDataMatchesStruct`, `JSONDataTypesMatchStruct`, `SecureValuesMatchLoadSettings`) plus
  `TestLoadConfig`, `TestApplyDefaults`, `TestApplyDefaultsTrimsSecret`, `TestValidate`,
  `TestSettingsExamples`.
- `dsconfig.ParseAndResolveSchemaJSON` + `Schema.Validate()` on `dsconfig.json` (via the
  `ConfigSchemaValid` conformance subtest).
- Strict JSON Schema validation of `dsconfig.json` against `dsconfig/schema.json`
  (draft-07 meta-schema, `additionalProperties: false`, `$schema` const enforced) via Ajv — VALID.
- `tsc --noEmit --strict` on `settings.ts` (TypeScript 5) — no errors.
- `go build ./...` in the `dsconfig` and `schema` workspace modules — both still build.
