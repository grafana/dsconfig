# grafana-astradb-datasource

Declarative configuration schema for the [AstraDB datasource plugin](https://github.com/grafana/astradb-datasource) (`grafana-astradb-datasource`).

## Upstream researched

- **Repo**: `github.com/grafana/astradb-datasource`
- **Ref**: `main`
- **Commit SHA**: `5c4a0400ea91e20f6bb5070d539e71e259cb91fd` (2026-07-02, `docs: add signed commits requirement to CONTRIBUTING.md (#758)`)

Every value in [`dsconfig.json`](dsconfig.json) — labels, placeholders, option labels/values,
section titles, defaults, validations, dependency and required-when expressions, storage keys,
storage targets, value types, group titles, and instructions — is traceable to a specific
`file:line` in the upstream repo at this SHA. See [Field provenance](#field-provenance) below.

To reproduce this research:

```bash
git clone https://github.com/grafana/astradb-datasource
cd astradb-datasource
git checkout 5c4a0400ea91e20f6bb5070d539e71e259cb91fd
```

If upstream `main` has advanced past this SHA, re-diff the sources listed under [Sources
researched](#sources-researched) before merging any changes to this entry.

## Files

| File | Purpose |
| --- | --- |
| [`dsconfig.json`](dsconfig.json) | dsconfig v1 schema — single source of truth for all config fields, groups, relationships, and instructions |
| [`settings.ts`](settings.ts) | TypeScript models: `RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig` |
| [`settings.go`](settings.go) | Go `Config` model (flat: jsonData fields + `DecryptedSecureJSONData`), `PluginID`, `AuthKind` and `SecureJsonDataKey` typed constants, and the `LoadConfig` utility |
| [`schema.go`](schema.go) | k8s-style SDK plugin schema: embeds `dsconfig.json`, converts it via `dsconfig.NewSDKSchema`, and defines `SettingsExamples` for each auth/connection variant |
| [`settings_test.go`](settings_test.go) | Table tests for `LoadConfig`, `ApplyDefaults`, and `Validate` |
| [`conformance_test.go`](conformance_test.go) | Runs `schema.RunPluginTests` — the shared dsconfig conformance suite — against `dsconfig.json` |
| [`schema.gen.json`](schema.gen.json), [`settings.gen.json`](settings.gen.json), [`settings.examples.gen.json`](settings.examples.gen.json) | Committed schema artifacts (regenerate with `go generate ./...` inside this directory; the `SchemaArtifactInSync` conformance subtest guards drift) |

There is no per-entry `go.mod` — every registry entry is a subpackage of the shared
[`registry/`](..) module (`github.com/grafana/dsconfig/registry`).

## Sources researched

Every source below was read at the pinned upstream SHA
(`5c4a0400ea91e20f6bb5070d539e71e259cb91fd`), plus external editor components at the exact
versions the plugin's `package.json` pins.

### Plugin repo (`github.com/grafana/astradb-datasource@5c4a0400`)

| File | What was read |
| --- | --- |
| `src/plugin.json:4-5,22-26` | `pluginType` (`id`), `pluginName` (`name`), `docURL` (`info.links[0].url`) |
| `src/components/ConfigEditor.tsx:11-14` | `Connection` numeric enum (`TOKEN = 0`, `CREDENTIALS = 1`) — the values `jsonData.authKind` stores |
| `src/components/ConfigEditor.tsx:16-19` | Radio option `label`/`value` pairs (`"Token"` → 0, `"Credentials"` → 1) |
| `src/components/ConfigEditor.tsx:27-47` | `onSecureSettingChange`, `onSettingChange`, `onReset` — how each key is written to jsonData / secureJsonData |
| `src/components/ConfigEditor.tsx:64-71` | `setConnectionType` + `const kind = jsonData.authKind || Connection.TOKEN` — the load-time default fallback |
| `src/components/ConfigEditor.tsx:74-78` | Top-level `<InlineLabel>Authentication</InlineLabel>` + radio (`jsonData.authKind`) |
| `src/components/ConfigEditor.tsx:80-108` | Token-mode fields: `URI` (`jsonData.uri`, placeholder `$ASTRA_CLUSTER_ID-$ASTRA_REGION.apps.astra.datastax.com:443`) and `Token` (`secureJsonData.token`, placeholder `AstraCS:xxxxx`) |
| `src/components/ConfigEditor.tsx:110-167` | Credentials-mode fields: `GRPC Endpoint`, `Auth Endpoint`, `User Name`, `Password`, and the `Secure` checkbox (with the copy-paste `user` placeholder `localhost:8090` — see [Upstream findings](#upstream-findings)) |
| `src/types.ts:5-14` | `AstraSettings` (jsonData shape); includes dead `database` field and misleading non-secret `password`/`user` — see modeling decisions below |
| `src/types.ts:16-19` | `SecureSettings` — the secret keys (`token`, `password`) |
| `pkg/models/settings.go:11-16` | `AuthType uint8` + `AuthTypeToken = iota` / `AuthTypeCredentials` — numeric values used in Go |
| `pkg/models/settings.go:18-27` | `Settings` struct: the storage shape and json tags for jsonData fields |
| `pkg/models/settings.go:29-56` | `LoadSettings`: parse order, mapstructure-based decoding of secure values by field name (matching secureJsonData keys `token`/`password`), the auth-mode-gated secret copy |
| `pkg/plugin/plugin.go:20-31` | `NewDatasource` and the `AstraDatasource{settings}` runtime shape |
| `pkg/plugin/handlers_checkhealth.go:12-58` | The mandatory-field checks — encoded as `requiredWhen` in the schema and mirrored in `Config.Validate` |
| `pkg/plugin/handlers_querydata.go:92-138` | `connect()`: how URI, token, gRPC endpoint, auth endpoint, user, password, and secure are consumed when building the gRPC client and the two token providers |
| `package.json` | External component versions (see next table) |

### External editor components

Read at the exact versions pinned in the plugin's `package.json`.

| Component | Version | What was read |
| --- | --- | --- |
| `LegacyForms.FormField`, `LegacyForms.SecretFormField` | `@grafana/ui@^13.0.1` | Prop shape (`label`, `placeholder`, `isConfigured`, `onReset`, `onChange`, `labelWidth`, `inputWidth`, `value`) so the recorded editor labels/placeholders map correctly |
| `RadioButtonGroup`, `InlineLabel`, `InlineFormLabel`, `Checkbox` | `@grafana/ui@^13.0.1` | Component behavior for the auth-mode radio and the `secure` checkbox |
| `DataSourcePluginOptionsEditorProps`, `onUpdateDatasourceJsonDataOptionChecked` | `@grafana/data@^13.0.1` | Storage-key semantics of the checkbox update helper (writes `jsonData.secure`) |

The plugin's `package.json` also declares `@grafana/plugin-ui@^0.13.1`, but the config editor
does not use any component from it (it uses `@grafana/ui` LegacyForms and `@emotion/css`
directly). No jsonData or secureJsonData storage keys come from an external component in this
plugin.

## Field provenance

Every dsconfig field, traced from its schema `id` to the upstream `file:line` where each of its
label, placeholder, default, storage key, and value type is defined.

| Schema `id` | Storage key | Target | Label source | Placeholder / options / default source | Value type source | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `jsonData_authKind` | `authKind` | `jsonData` | `ConfigEditor.tsx:76` (`<InlineLabel>Authentication</InlineLabel>`) | Options `ConfigEditor.tsx:16-19`; default resolves from `jsonData.authKind || Connection.TOKEN` (`:71`) — numerically `0` | `AuthType uint8` `pkg/models/settings.go:11`, `Connection` enum `ConfigEditor.tsx:11-14` | Role `auth.discriminator`; validated to the exact numeric set `{0, 1}` |
| `jsonData_uri` | `uri` | `jsonData` | `ConfigEditor.tsx:84` (`label="URI"`) | `ConfigEditor.tsx:89` (`placeholder="$ASTRA_CLUSTER_ID-$ASTRA_REGION.apps.astra.datastax.com:443"`) | `Settings.URI string`, `pkg/models/settings.go:19` | Role `endpoint.baseUrl`; `dependsOn`/`requiredWhen` from conditional render `ConfigEditor.tsx:80` + CheckHealth guard `handlers_checkhealth.go:15-17` |
| `secureJsonData_token` | `token` | `secureJsonData` | `ConfigEditor.tsx:98` (`label="Token"`) | `ConfigEditor.tsx:99` (`placeholder="AstraCS:xxxxx"`) | `SecureSettings.token` `src/types.ts:17`; `Settings.Token string` (secret alias) `pkg/models/settings.go:20` | Role `auth.bearer.token`; `dependsOn`/`requiredWhen` from conditional render `ConfigEditor.tsx:80` + CheckHealth guard `handlers_checkhealth.go:19-21` |
| `jsonData_grpcEndpoint` | `grpcEndpoint` | `jsonData` | `ConfigEditor.tsx:114` (`label="GRPC Endpoint"`) | `ConfigEditor.tsx:119` (`placeholder="localhost:8090"`) | `Settings.GRPCEndpoint string`, `pkg/models/settings.go:21` | Role `endpoint.baseUrl`; `dependsOn`/`requiredWhen` from conditional render `ConfigEditor.tsx:110` + CheckHealth guard `handlers_checkhealth.go:25-27` |
| `jsonData_authEndpoint` | `authEndpoint` | `jsonData` | `ConfigEditor.tsx:124` (`label="Auth Endpoint"`) | `ConfigEditor.tsx:129` (`placeholder="localhost:8081"`) | `Settings.AuthEndpoint string`, `pkg/models/settings.go:22` | `dependsOn`/`requiredWhen` from conditional render `ConfigEditor.tsx:110` + CheckHealth guard `handlers_checkhealth.go:29-31` |
| `jsonData_user` | `user` | `jsonData` | `ConfigEditor.tsx:134` (`label="User Name"`) | `ConfigEditor.tsx:139` (`placeholder="localhost:8090"` — copy-paste error preserved verbatim; see [Upstream findings](#upstream-findings) #1) | `Settings.UserName string` (`json:"user"`), `pkg/models/settings.go:23` | Role `auth.basic.username`; `dependsOn`/`requiredWhen` from conditional render `ConfigEditor.tsx:110` + CheckHealth guard `handlers_checkhealth.go:33-35` |
| `secureJsonData_password` | `password` | `secureJsonData` | `ConfigEditor.tsx:147` (`label="Password"`) | `ConfigEditor.tsx:148` (`placeholder="xxxxx"`) | `SecureSettings.password` `src/types.ts:18`; `Settings.Password string` (secret alias) `pkg/models/settings.go:24` | Role `auth.basic.password`; `dependsOn`/`requiredWhen` from conditional render `ConfigEditor.tsx:110` + CheckHealth guard `handlers_checkhealth.go:37-39` |
| `jsonData_secure` | `secure` | `jsonData` | `ConfigEditor.tsx:157` (`<InlineFormLabel>Secure</InlineFormLabel>`) | Default `false` from Go zero value + `Settings.Secure bool` `pkg/models/settings.go:25` | `Settings.Secure bool`, `pkg/models/settings.go:25` | `dependsOn` from conditional render `ConfigEditor.tsx:110`; **not** marked `requiredWhen` — checkbox has no CheckHealth guard, `false` is a legitimate value |

## Field inventory summary

| Schema field | Storage key | Target | Editor label | Read by backend? |
| --- | --- | --- | --- | --- |
| `jsonData_authKind` | `authKind` | `jsonData` | Authentication | Yes |
| `jsonData_uri` | `uri` | `jsonData` | URI | Yes (Token mode) |
| `secureJsonData_token` | `token` | `secureJsonData` | Token | Yes (Token mode) |
| `jsonData_grpcEndpoint` | `grpcEndpoint` | `jsonData` | GRPC Endpoint | Yes (Credentials mode) |
| `jsonData_authEndpoint` | `authEndpoint` | `jsonData` | Auth Endpoint | Yes (Credentials mode) |
| `jsonData_user` | `user` | `jsonData` | User Name | Yes (Credentials mode) |
| `secureJsonData_password` | `password` | `secureJsonData` | Password | Yes (Credentials mode) |
| `jsonData_secure` | `secure` | `jsonData` | Secure | Yes (Credentials mode) |

### Frontend-only settings

- **None.** Every editor-written field is also read by the backend (though only for the auth
  mode selected via `authKind`).

### Backend-only settings

- **None.** There are no jsonData/secureJsonData fields that the backend reads but the editor
  doesn't render.

### Excluded / dead fields

- **`jsonData.database`** — declared in `src/types.ts:7` as an optional `string`, but the
  editor never writes it and the backend `Settings` struct (`pkg/models/settings.go:18-27`) has
  no corresponding field. Pure dead weight in the TypeScript type; not modeled in the schema.
- **`AstraSettings.password: string`** and **`AstraSettings.user: string`** (`src/types.ts:9-10`)
  — the `AstraSettings` interface declares a non-optional `password` inside `jsonData`, but the
  editor only ever writes the password to `secureJsonData`. `user` is a real jsonData field
  (editor writes it via `onSettingChange('user')`), but `password` at the jsonData level is
  never written. Only the real storage locations are modeled.
- **`Settings.Token`** (`json:"token"`) and **`Settings.Password`** (`json:"password"`) in
  `pkg/models/settings.go:20,24` — these aliases collide with the secret keys of the same
  name. Upstream's `LoadSettings` overwrites them from `DecryptedSecureJSONData` after the
  jsonData unmarshal, so if `jsonData.token`/`jsonData.password` were ever populated, the
  values are discarded. Treated here as a backend implementation artifact, not a jsonData
  storage contract; not modeled as jsonData fields. See [Upstream findings](#upstream-findings)
  #4.
- **Secure Socks Proxy** (`jsonData.enableSecureSocksProxy`) — not implemented by this plugin
  (no `SecureSocksProxySettings` in the editor, no `SecureSocksProxyEnabled` check in the
  backend). Not applicable, so not present.

## Where the types are defined

The configuration types are spread across the plugin and its dependencies — some types come
from libraries/SDKs rather than the plugin itself:

### Frontend (TypeScript)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `AstraSettings` (jsonData), `SecureSettings` (secret keys), `AstraQuery`, `Connection` enum | `src/types.ts:5-19`, `src/components/ConfigEditor.tsx:11-14` | plugin ([grafana/astradb-datasource](https://github.com/grafana/astradb-datasource)) |
| `DataSourceJsonData` (base interface `AstraSettings` extends) | `packages/grafana-data/src/types/datasource.ts` | `@grafana/data` `^13.0.1` |
| `DataSourcePluginOptionsEditorProps`, `onUpdateDatasourceJsonDataOptionChecked` | `packages/grafana-data/src/` | `@grafana/data` `^13.0.1` |
| `LegacyForms.FormField`, `LegacyForms.SecretFormField`, `RadioButtonGroup`, `InlineLabel`, `InlineFormLabel`, `Checkbox` | `packages/grafana-ui/src/components/` | `@grafana/ui` `^13.0.1` |
| `SQLQuery` (base interface `AstraQuery` extends — not part of the config schema) | `@grafana/plugin-ui` | `@grafana/plugin-ui` `^0.13.1` |

### Backend (Go)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `Settings` (jsonData + secret-alias fields), `AuthType` (`AuthTypeToken`, `AuthTypeCredentials`), `LoadSettings` | `pkg/models/settings.go:11-56` | plugin ([grafana/astradb-datasource](https://github.com/grafana/astradb-datasource)) |
| `backend.DataSourceInstanceSettings` (carries `JSONData`, `DecryptedSecureJSONData`, and root fields like `URL`, `User` — **unused by this plugin**) | `backend/common.go` | `github.com/grafana/grafana-plugin-sdk-go` |
| `mapstructure.Decode` (used to copy `DecryptedSecureJSONData` into `Settings.Token` / `Settings.Password` by field name) | — | `github.com/mitchellh/mapstructure` |
| gRPC/TLS transports the settings feed into: `grpc.NewClient`, `credentials.NewTLS`, `insecure.NewCredentials`, `auth.NewStaticTokenProvider`, `auth.NewTableBasedTokenProvider` | — | `google.golang.org/grpc`, `github.com/stargate/stargate-grpc-go-client` |

The models in this entry flatten that spread into a single Go `Config` type (jsonData fields +
`DecryptedSecureJSONData`) plus a `SecureJsonDataKey` typed constant list. `settings.ts` keeps
the three canonical TypeScript types (`RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig`).

## Modeling decisions

- **`authKind` as a numeric discriminator, not virtual.** The upstream editor stores the auth
  mode as a JSON number (`0` / `1`) in `jsonData.authKind` — no derivation, no
  `secureJsonFields`-based inference. So `jsonData_authKind` is modeled as an ordinary storage
  field with `valueType: "number"`, `role: "auth.discriminator"`, and `allowedValues` `[0, 1]`.
  No virtual selector layer is needed.
- **`requiredWhen` mirrors the CheckHealth guards, not the editor.** The editor renders no
  `required` markers at all (LegacyForms `FormField`/`SecretFormField` do not surface a
  required affordance). But `pkg/plugin/handlers_checkhealth.go:12-40` hard-fails on any empty
  Token-mode URI / token or Credentials-mode grpcEndpoint / authEndpoint / user / password.
  `requiredWhen` encodes that backend contract.
- **`jsonData_secure` has no `requiredWhen`.** The checkbox has no CheckHealth guard — `false`
  is a legitimate stored value (plaintext dev connections). Only `dependsOn` is set.
- **Two-way "group" relationships plus one basic-auth "pair".** The Token and Credentials
  mode field sets are declared as separate `group` relationships to make the mutually
  exclusive shape explicit for downstream consumers, and `jsonData_user` +
  `secureJsonData_password` form a `pair` because they're the basic-auth username/password
  posted to the Stargate table-based auth endpoint.
- **No virtual field for the "either token XOR credentials" invariant.** Since `authKind`
  itself is a real storage field and its two values already drive `dependsOn`/`requiredWhen`
  on every other field, a virtual layer would just duplicate the same logic. Kept flat.
- **Flat `Config` in Go with a deliberate deviation from upstream.** `settings.go` mirrors
  `pkg/models/settings.go` for the six real jsonData fields (URI, GRPCEndpoint, AuthEndpoint,
  UserName, Secure, AuthKind) but **omits** the `Token`/`Password` string fields with
  `json:"token"`/`json:"password"` tags that upstream carries. Those tags are ambiguous — they
  imply `token`/`password` live in jsonData, but they don't; they're mapstructure-decoded from
  `DecryptedSecureJSONData` at the same field name. Modeling them as jsonData would put them
  in the OpenAPI settings spec next to the real fields and violate the conformance suite's
  spec/secure separation guarantee. Instead, both secrets live only in
  `DecryptedSecureJSONData` on our `Config`. See [Upstream findings](#upstream-findings) #4
  for the recommended upstream fix.
- **`SecureJsonDataConfig` is a key list**: secure values are write-only, so the secure type
  is just the array of secret key names (`token`, `password`); consumers read
  `secureJsonFields` to see what is configured.
- **Field ID naming convention**: IDs are prefixed with their storage target for
  discoverability — `jsonData_` or `secureJsonData_` — followed by the camelCase storage key.
  The `key` property keeps the plugin's raw storage key (`authKind`, `token`) — `id` is the
  schema reference, `key` is the storage contract.

## SDK plugin schema and k8s-style examples (`schema.go`)

`NewSchema()` assembles the `grafana-plugin-sdk-go` `pluginschema.PluginSchema` bundle (the
k8s-style schema Grafana's datasource API server serves as `{apiVersion}.json`, `v0alpha1`
today) from the embedded `dsconfig.json`: root fields plus a nested `jsonData` object become
the OpenAPI settings `spec`, secure fields become `secureValues`.

`SettingsExamples()` provides the default configuration plus one k8s-style example per
authentication mode and connection variant. Each example is a full instance-settings object
with the plugin configuration nested under `jsonData` and the relevant write-only secrets
under `secureJsonData`:

| Example | Auth | Connection | `secureJsonData` |
| --- | --- | --- | --- |
| `""` (default) | Token (schema defaults) | authKind=0, empty uri | `token` (empty placeholder) |
| `tokenAstraCloud` | Token | Astra Cloud gRPC (443) | `token` |
| `credentialsSelfHostedTLS` | Credentials | Self-hosted Stargate, `secure=true` (TLS) | `password` |
| `credentialsSelfHostedPlaintext` | Credentials | Self-hosted Stargate, `secure=false` (plaintext) | `password` |
| `legacyMissingAuthKind` | Legacy: no `authKind` stored | URI present; interpreted as Token | `token` |

## `LoadConfig` utility (`settings.go`)

`LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error)`
runs the full three-phase load flow on a datasource instance's settings and returns a
fully-defaulted, validated `Config`:

1. **Parse** — unmarshal jsonData into `Config` and copy every decrypted secret from
   `settings.DecryptedSecureJSONData` into `Config.DecryptedSecureJSONData` (both keys,
   unconditionally — see "Two deliberate deviations" in the doc comment on `LoadConfig`).
2. **`ApplyDefaults`** — fill the curated set of zero-valued discriminators with the same
   defaults the editor writes for a fresh datasource. Today that's just `AuthKind` →
   `AuthKindToken`, which is numerically a no-op (both are `0`) but is preserved as
   documentation of intent.
3. **`Validate`** — enforce the runtime contract from
   `pkg/plugin/handlers_checkhealth.go:12-40`: `AuthKindToken` requires `URI` and the token
   secret; `AuthKindCredentials` requires `GRPCEndpoint`, `AuthEndpoint`, `UserName`, and the
   password secret. Unknown `AuthKind` values are rejected. Errors are joined so every problem
   surfaces at once.

Everything is logged via `backend.Logger.FromContext(ctx)` with `datasource_uid`,
`datasource_name`, and `plugin` labels so log lines carry request context.

This is the intended shape for the plugin's own upstream `LoadSettings` to sync to: a load
returns a config that is safe to use, or an error explaining why it isn't.

### Direct access to individual phases

`(*Config).ApplyDefaults()` and `(Config).Validate() error` are exported for callers that want
to compose them themselves (e.g. provisioning preview, schema-example round-trip, tests that
need to distinguish parse-level from policy-level errors).

## Upstream findings

Potential bugs, misleading UX, and consistency issues discovered while researching upstream.
All preserved verbatim in the schema — the schema records what the plugin **does**, not what
it **should** do; these notes exist so reviewers can reproduce each finding and decide
separately whether to fix upstream.

1. **Copy-paste placeholder on the "User Name" field.** `src/components/ConfigEditor.tsx:139`
   — the placeholder is `"localhost:8090"` (identical to the GRPC Endpoint placeholder on
   `:119`), suggesting a host:port when the field is actually a basic-auth username. Preserved
   verbatim in the schema.
2. **`authKind` uses a numeric enum, not a string.** `src/components/ConfigEditor.tsx:11-14`
   declares `enum Connection { TOKEN = 0, CREDENTIALS = 1 }`, and
   `pkg/models/settings.go:11-16` mirrors that as `AuthType uint8`. Serialization to
   `jsonData.authKind` is therefore a JSON number, not a string — a common source of
   provisioning bugs (`{"authKind": "0"}` won't parse). Documented in the LLM instructions.
3. **`jsonData.authKind || Connection.TOKEN` treats `0` as unset.** The load-time default at
   `src/components/ConfigEditor.tsx:71` uses JavaScript's `||` which falls back on any falsy
   left-hand side — including `0`. `Connection.TOKEN` happens to also be `0`, so today this
   accidentally works. If either value ever changes (e.g. if `Connection.CREDENTIALS` becomes
   `0`), this fallback becomes silently wrong.
4. **`Settings.Token` and `Settings.Password` shadow the secret keys.**
   `pkg/models/settings.go:20,24` tags `Token string`/`Password string` as `json:"token"` /
   `json:"password"` — the same names as the secure keys. `LoadSettings` then overrides those
   struct fields via `mapstructure.Decode(config.DecryptedSecureJSONData, &secureSettings)`
   (`:42-44,50-52`), keyed by field name. Anyone who accidentally sets `jsonData.token` or
   `jsonData.password` on a datasource will have those values silently discarded. Suggested
   fix upstream: rename the struct fields (or tag them `json:"-"`) and load secrets into a
   separate map, matching this entry's `DecryptedSecureJSONData` shape.
5. **Auth-mode-gated secret loading loses the other secret.**
   `pkg/models/settings.go:40-54` only copies the token when `AuthKind == AuthTypeToken` and
   only copies the password when `AuthKind == AuthTypeCredentials`. Switching auth modes at
   runtime therefore drops any previously-loaded "other" secret from `Settings` even when it
   still exists in `DecryptedSecureJSONData`. This entry's `LoadConfig` loads both secrets
   unconditionally; downstream callers authenticate using only the secret that matches
   `AuthKind`.
6. **`authEndpoint` scheme is prepended by the backend, not the user.**
   `pkg/plugin/handlers_querydata.go:117,125` builds
   `fmt.Sprintf("https://%s/v1/auth", d.settings.AuthEndpoint)` or `"http://%s/v1/auth"`
   depending on `Secure`. A user who copies a full URL (`https://host:8081`) into the field
   ends up with an invalid `https://https://host:8081/v1/auth`. The editor placeholder
   (`localhost:8081`) hints at host:port only but does not enforce it.
7. **`jsonData.database` is dead weight.** `src/types.ts:7` declares an optional `database`
   field that has no editor UI (`src/components/ConfigEditor.tsx`), no backend read
   (`pkg/models/settings.go:18-27` and the rest of `pkg/`), and no runtime effect. Left in the
   TypeScript type by past refactoring; not modeled in the schema.
8. **`AstraSettings.password` and `AstraSettings.user` inside `jsonData` are misleading.**
   `src/types.ts:9-10` declares `user: string` and `password: string` as non-optional
   jsonData fields; only `user` is actually written by the editor. `password` inside jsonData
   is never written — the field name collides with the secure key and would be discarded even
   if written. See finding #4.
9. **CheckHealth executes a real query on the health check.**
   `pkg/plugin/handlers_checkhealth.go:51-56` runs `select keyspace_name from
   system_schema.keyspaces` against the target. Failed authentication surfaces as a query
   error instead of a friendlier "invalid credentials" message; the error text is returned
   verbatim to the UI. Not modeled in the schema, but worth knowing for provisioning
   consumers.

## Validation performed

- `dsconfig.ParseAndResolveSchemaJSON` + `Schema.Validate()` (Go validator in this repo) —
  passes.
- JSON Schema validation against [`dsconfig/schema.json`](../../dsconfig/schema.json) (draft
  2020-12, `additionalProperties: false`) — passes via the shared conformance suite.
- `go test ./...` on this module — passes (schema bundle shape, secure values, examples,
  `LoadConfig` including legacy-missing-authKind fallback, `SchemaArtifactInSync` guard).
- `go generate ./...` on this entry — regenerates the `.gen.json` artifacts cleanly.
- `settings.go`/`schema.go`: `go build`, `go vet`, `gofmt` — clean.
