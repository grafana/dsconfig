# grafana-salesforce-datasource

Declarative configuration schema for the Salesforce datasource plugin (`grafana-salesforce-datasource`).

## Upstream researched

- **Repo**: `github.com/grafana/plugins-private` (monorepo)
- **Plugin path**: `plugins/grafana-salesforce-datasource`
- **Ref**: `main`
- **Commit SHA**: `267f4937806ed6404b6628d13ae358a5d308e376`
- **Plugin version**: `1.7.22` (`package.json:3`)

Every value in [`dsconfig.json`](dsconfig.json) — labels, placeholders, option labels/values,
section titles, defaults, dependency and required-when expressions, storage keys, storage targets,
value types, group titles, and instructions — is traceable to a specific `file:line` in the upstream
monorepo at this SHA. See [Field provenance](#field-provenance) below.

To reproduce this research (the monorepo was read on disk, not cloned):

```bash
git -C <plugins-private> rev-parse HEAD   # 267f4937806ed6404b6628d13ae358a5d308e376
sed -n '1,366p' plugins/grafana-salesforce-datasource/src/views/SFConfigEditor.tsx
sed -n '1,125p'  plugins/grafana-salesforce-datasource/pkg/models/settings.go
```

If upstream `main` has advanced past this SHA, re-diff the sources listed under [Sources
researched](#sources-researched) before merging any changes to this entry.

## Files

| File | Purpose |
| --- | --- |
| [`dsconfig.json`](dsconfig.json) | dsconfig v1 schema — single source of truth for all config fields, groups, relationships, and instructions |
| [`settings.ts`](settings.ts) | TypeScript models: `RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig` |
| [`settings.go`](settings.go) | Go `Config` model (flat: jsonData fields + `DecryptedSecureJSONData`), `PluginID`, `AuthType`/`SecureJsonDataKey` typed constants, and the `LoadConfig`/`ApplyDefaults`/`Validate` utilities |
| [`schema.go`](schema.go) | k8s-style SDK plugin schema: embeds `dsconfig.json`, converts it via `dsconfig.NewSDKSchema`, and defines `SettingsExamples` for each auth/environment variant |
| [`settings_test.go`](settings_test.go) | Table tests for `LoadConfig`, `ApplyDefaults`, and `Validate` |
| [`conformance_test.go`](conformance_test.go) | Runs `schema.RunPluginTests` — the shared dsconfig conformance suite — against `dsconfig.json` |
| [`schema.gen.json`](schema.gen.json), [`settings.gen.json`](settings.gen.json), [`settings.examples.gen.json`](settings.examples.gen.json) | Committed schema artifacts (regenerate with `go generate ./...` in this directory; the `SchemaArtifactInSync` conformance subtest guards drift) |

There is no per-entry `go.mod` — every registry entry is a subpackage of the shared
[`registry/`](..) module (`github.com/grafana/dsconfig/registry`). Package name:
`salesforcedatasource`.

## Sources researched

Every source below was read at the pinned monorepo SHA
(`267f4937806ed6404b6628d13ae358a5d308e376`), plus external editor components at the exact versions
the plugin's `package.json` pins via the monorepo `.yarnrc.yml` catalog.

### Plugin (`plugins/grafana-salesforce-datasource` @ 267f4937)

| File | What was read |
| --- | --- |
| `src/plugin.json:3-5,28` | `type` (datasource), `name` (`Salesforce` → `pluginName`), `id` (`grafana-salesforce-datasource` → `pluginType`), docs link (`info.links[0].url` → `docURL`) |
| `src/views/SFConfigEditor.tsx:163-166` | `FieldSet label="Connection Settings"`, the `Authentication` label and the auth-type `RadioButtonGroup` |
| `src/views/SFConfigEditor.tsx:29-37` | `normalizeAuthType` — the editor's secureJsonFields-based auth detection (init state) |
| `src/views/SFConfigEditor.tsx:39-47` | `getTokenUrl` — the editor's tokenUrl/sandbox init derivation |
| `src/views/SFConfigEditor.tsx:148-154` | `onChangeAuthType` — writes `jsonData.authType` directly |
| `src/views/SFConfigEditor.tsx:168-241` | `authType==='user'` block: User Name, Password, Security Token, Consumer Key, Consumer Secret — labels, placeholders, storage keys |
| `src/views/SFConfigEditor.tsx:242-305` | `authType==='jwt'` block: Certificate, Private Key, User Name, Consumer Key + the `Generate`/`Download` certificate buttons and the `Connected App Digital Signature`/`Connected App Credentials` sub-headers |
| `src/views/SFConfigEditor.tsx:307-325` | `FieldSet label="Optional Settings"`, the `Environment` label and the tokenUrl `Select` |
| `src/views/SFConfigEditor.tsx:327-362` | Conditional `Enable Secure Socks Proxy` switch — deliberately excluded from this entry |
| `src/types.ts:4-33` | `AuthType`, `TokenURL`, `DatasourceOptions` (jsonData), `SecureJsonData` |
| `src/constants.ts:4-12` | `AuthTypes` (Credentials/JWT) and `TokenURLs` (Production/SandBox) option labels/values |
| `src/selectors.ts:3-20` | Config-editor E2E aria-label selectors |
| `src/views/__fixtures__/ConfigEditor.fixtures.ts:21` | Confirms fresh jsonData shape `{user, sandbox:false, authType:'user'}` |
| `pkg/models/settings.go:13-39` | `TokenURLProd`/`TokenURLSandbox`, `AuthType` constants, and the `Settings` struct with json tags |
| `pkg/models/settings.go:41-80` | `GetSettings`: parse order, `normalizeAuthType`/`normalizeTokenUrl`, decrypted-secret copies by auth method, `HTTPClientOptions` |
| `pkg/models/settings.go:82-100` | `normalizeAuthType` (empty → user; jwt auto-detect from jsonData struct fields) and `normalizeTokenUrl` (empty → sandbox?test:prod) |
| `pkg/models/settings.go:102-124` | `Validate`: jwt requires cert+privateKey; user requires user+password+clientID+clientSecret |
| `pkg/plugin/client.go:76-147` | `fetchToken`: `Validate` gate, `<tokenUrl>/services/oauth2/token`, JWT bearer vs password grant form fields |
| `pkg/plugin/client.go:368-389` | `get`: uses the token response `instance_url` (not tokenUrl) as the API base + `Authorization: Bearer` |
| `pkg/jwt/jwt.go:23-73` | `CreateJWT`/`Create`: privateKey signs, clientID=issuer, user=subject, tokenUrl=audience |
| `pkg/plugin/datasource.go:14-28` | `NewInstance` → `GetSettings` → `NewClient`; root `url` never read |
| `pkg/plugin/handlers_checkhealth.go:11-23` | Health check calls `fetchToken` (so `Validate` is the load-time contract) |
| `package.json:34-46` | External component versions (via catalog) |

### External editor components

Read at the versions the plugin pins through the monorepo `.yarnrc.yml` catalog
(`plugins-private/.yarnrc.yml:14-50`). `@grafana/e2e-selectors` is intentionally not cataloged.

| Component / type | Version | Package | What was read |
| --- | --- | --- | --- |
| `RadioButtonGroup`, `Input`, `Select`, `LegacyForms.SecretFormField`, `CertificationKey`, `FieldSet`, `InlineField`, `InlineFormLabel`, `InlineSwitch` | `@grafana/ui@^11.6.7` | grafana/grafana | Prop names (`label`, `placeholder`, `value`, `options`, `isConfigured`, `onReset`) to know which attributes to record; `Select` has no `allowCustomValue`, so the editor offers only the two options; `CertificationKey` renders a TextArea for cert/privateKey |
| `DataSourceJsonData`, `DataSourcePluginOptionsEditorProps`, `onUpdateDatasourceJsonDataOption`, `SelectableValue` | `@grafana/data@^11.6.7` | grafana/grafana | Base jsonData interface and the update helper's storage-key semantics |
| `config`, `getDataSourceSrv` | `@grafana/runtime@^11.6.7` | grafana/grafana | `config.featureToggles`/`buildInfo` gate the excluded Secure Socks Proxy switch; `getDataSourceSrv` backs the JWT Generate/Download buttons |
| `Settings` (backend), `httpclient.Options` | `grafana-plugin-sdk-go@v0.279.0` | grafana/grafana-plugin-sdk-go | `DataSourceInstanceSettings` (JSONData, DecryptedSecureJSONData, unused root fields); `HTTPClientOptions` |
| `jwt` (RS256 signing) | `golang-jwt/jwt/v4@v4.5.2` | golang-jwt | Claim mapping (issuer/subject/audience) used by `pkg/jwt` |

## Field provenance

| Schema `id` | Storage key | Target | Label source | Placeholder / options / default source | Value type source | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `jsonData_authType` | `authType` | `jsonData` | `SFConfigEditor.tsx:165` (`Authentication`) | Options `constants.ts:4-7` (Credentials=`user`, JWT=`jwt`); default `user` from `normalizeAuthType` `settings.go:82-90` + editor init `SFConfigEditor.tsx:60` | `AuthType`, `types.ts:4` / `settings.go:18-23` | Role `auth.discriminator`; `allowedValues` [user, jwt] |
| `jsonData_user` | `user` | `jsonData` | `SFConfigEditor.tsx:170,278` (`User Name`) | `SFConfigEditor.tsx:172,280` (`Salesforce User`) | `Settings.User string` `settings.go:27`; TS `string` `types.ts:8` | Shown in both auth modes; `requiredWhen authType=='user'` per `Validate` `settings.go:112-113` |
| `secureJsonData_password` | `password` | `secureJsonData` | `SFConfigEditor.tsx:182` (`Password`) | `SFConfigEditor.tsx:183` (`Salesforce Password`) | `Settings.Password string` `settings.go:28`; TS `types.ts:27` | `dependsOn`/`requiredWhen authType=='user'` (`settings.go:115-117`) |
| `secureJsonData_securityToken` | `securityToken` | `secureJsonData` | `SFConfigEditor.tsx:197` (`Security Token`) | `SFConfigEditor.tsx:198` (`Salesforce Security Token`) | `Settings.SecurityToken string` `settings.go:29`; TS `types.ts:28` | `dependsOn authType=='user'`; **no** `requiredWhen` — used (`client.go:104`) but not validated |
| `secureJsonData_clientID` | `clientID` | `secureJsonData` | `SFConfigEditor.tsx:212,290` (`Consumer Key`) | `SFConfigEditor.tsx:213,291` (`Connected App Consumer Key`) | `Settings.ClientID string` `settings.go:30`; TS `types.ts:29` | Role `auth.oauth2.clientId`; shown in both modes; `requiredWhen authType=='user'` (validated for user `settings.go:118-119`; used but not validated for jwt) |
| `secureJsonData_clientSecret` | `clientSecret` | `secureJsonData` | `SFConfigEditor.tsx:227` (`Consumer Secret`) | `SFConfigEditor.tsx:228` (`Connected App Consumer Secret`) | `Settings.ClientSecret string` `settings.go:31`; TS `types.ts:30` | Role `auth.oauth2.clientSecret`; `dependsOn`/`requiredWhen authType=='user'` (`settings.go:121-123`) |
| `secureJsonData_cert` | `cert` | `secureJsonData` | `SFConfigEditor.tsx:255` (`Certificate`) | `SFConfigEditor.tsx:256` (`-----BEGIN CERTIFICATE-----`) | `Settings.Cert string` `settings.go:32`; TS `types.ts:31` | `textarea`; `dependsOn`/`requiredWhen authType=='jwt'` (`settings.go:103-106`) |
| `secureJsonData_privateKey` | `privateKey` | `secureJsonData` | `SFConfigEditor.tsx:269` (`Private Key`) | `SFConfigEditor.tsx:270` (`-----BEGIN PRIVATE KEY-----`) | `Settings.PrivateKey string` `settings.go:33`; TS `types.ts:32` | Role `auth.oauth2.jwtPrivateKey`; `textarea`; `dependsOn`/`requiredWhen authType=='jwt'` (`settings.go:107-109`) |
| `jsonData_tokenUrl` | `tokenUrl` | `jsonData` | `SFConfigEditor.tsx:309` (`Environment`) | Options `constants.ts:9-12` (Production=`https://login.salesforce.com`, SandBox=`https://test.salesforce.com`); default prod `settings.go:92-99` | `Settings.TokenURL string` `settings.go:35`; TS `TokenURL` `types.ts:5,20` | Role `auth.oauth2.tokenUrl`; `select` (2 options); backend accepts any string (see discrepancies) |
| `jsonData_sandbox` | `sandbox` | `jsonData` | — (no UI) | Default `false` (fixture `ConfigEditor.fixtures.ts:21`) | `Settings.Sandbox bool` `settings.go:34`; TS `types.ts:13` | Legacy; read by `getTokenUrl` `SFConfigEditor.tsx:43` and `normalizeTokenUrl` `settings.go:92-100`; `tags:[legacy]`, `description` only |

## Field inventory summary

| Schema field | Storage key | Target | Editor label | Read by backend? |
| --- | --- | --- | --- | --- |
| `jsonData_authType` | `authType` | `jsonData` | Authentication | Yes |
| `jsonData_user` | `user` | `jsonData` | User Name | Yes |
| `secureJsonData_password` | `password` | `secureJsonData` | Password | Yes (user auth) |
| `secureJsonData_securityToken` | `securityToken` | `secureJsonData` | Security Token | Yes (user auth; not validated) |
| `secureJsonData_clientID` | `clientID` | `secureJsonData` | Consumer Key | Yes (both auth methods) |
| `secureJsonData_clientSecret` | `clientSecret` | `secureJsonData` | Consumer Secret | Yes (user auth) |
| `secureJsonData_cert` | `cert` | `secureJsonData` | Certificate | Yes (jwt auth) |
| `secureJsonData_privateKey` | `privateKey` | `secureJsonData` | Private Key | Yes (jwt auth) |
| `jsonData_tokenUrl` | `tokenUrl` | `jsonData` | Environment | Yes |
| `jsonData_sandbox` | `sandbox` | `jsonData` | — (no UI) | Yes (legacy fallback) |

### Frontend-only settings

None. Every stored field is read by the backend. (`enableSecureSocksProxy` is written by the editor
and consumed by the SDK's `HTTPClientOptions`, but is excluded from this entry per AGENTS.md and is
not read by the plugin's own Go code.)

### Backend-only settings

- **`sandbox`** has no config-editor UI (the editor replaced it with the `Environment` select that
  writes `tokenUrl`). It remains part of the jsonData contract: the editor *reads* it to derive the
  initial Environment selection (`SFConfigEditor.tsx:39-47`) and the backend *reads* it in
  `normalizeTokenUrl` (`settings.go:92-100`). Retained for backwards compatibility / provisioning.

## Where the types are defined

Only config type/field definitions are listed. UI components (`RadioButtonGroup`,
`CertificationKey`, `SecretFormField`, `SecureSocksProxySettings`, …) and functions/helpers
(`GetSettings`, `normalizeAuthType`, `CreateJWT`, `onUpdateDatasourceJsonDataOption`, …) are omitted
even where they are the reason a field exists.

### Frontend (TypeScript)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `AuthType`, `TokenURL`, `DatasourceOptions` (jsonData), `SecureJsonData` | `src/types.ts:4-33` | plugin (`grafana-salesforce-datasource`) |
| `AuthTypes`, `TokenURLs` (option label/value sets) | `src/constants.ts:4-12` | plugin |
| `DataSourceJsonData` (base interface `DatasourceOptions` extends) | `packages/grafana-data/src/types/datasource.ts` | `@grafana/data@^11.6.7` (grafana/grafana) |

### Backend (Go)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `Settings` (jsonData + secret fields), `AuthType` (`AuthTypeUser`, `AuthTypeJWT`), `TokenURLProd`/`TokenURLSandbox` | `pkg/models/settings.go:13-39` | plugin (`grafana-salesforce-datasource`) |
| `backend.DataSourceInstanceSettings` (carries `JSONData`, `DecryptedSecureJSONData`, and unused root fields like `URL`, `BasicAuthEnabled`) | `backend/common.go` | `grafana-plugin-sdk-go@v0.279.0` |
| `httpclient.Options` | `backend/httpclient` | `grafana-plugin-sdk-go@v0.279.0` |

This entry flattens the spread into a single Go `Config` (jsonData fields + `DecryptedSecureJSONData`)
plus a `SecureJsonDataKey` typed constant list. `settings.ts` keeps the three canonical TypeScript
types (`RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig`).

## Modeling decisions

- **`authType` is a real stored discriminator, not a virtual field.** Unlike the GitHub entry's
  virtual license selector, the Salesforce editor writes `jsonData.authType` directly
  (`SFConfigEditor.tsx:148-154`), so it is a plain `jsonData` field with role `auth.discriminator`.
- **Secrets modeled in `DecryptedSecureJSONData`, not as struct fields.** The upstream `Settings`
  struct declares the six secrets with json tags (`json:"password"`, `json:"cert"`, …), but the
  editor stores them in `secureJsonData` and `GetSettings` copies them from `DecryptedSecureJSONData`
  (`settings.go:48-72`). Modeling them as jsonData struct fields would put them in the settings spec
  and break the conformance `JSONDataMatchesStruct` / `SchemaSpecHasNoSecureJSON` guards, so `Config`
  carries only the four jsonData fields (`authType`, `user`, `sandbox`, `tokenUrl`) and the decrypted
  secrets map. See discrepancy #1.
- **`requiredWhen` mirrors `Validate`, not editor visibility.** The editor marks nothing required, but
  `Settings.Validate` (`settings.go:102-124`) hard-fails settings load: jwt needs cert+privateKey;
  user needs user+password+clientID+clientSecret. Those are the `requiredWhen` rules. `dependsOn`
  mirrors editor visibility instead (which secrets render under each auth type).
- **`user` and `clientID` have no `dependsOn`.** Both render under *both* the `user` and `jwt` blocks
  (`SFConfigEditor.tsx:170,278` and `:211,289`), and `authType` has only those two values, so they
  are always visible. `clientID`/`user` are `requiredWhen authType=='user'` because `Validate` only
  enforces them for user auth (see discrepancy #3).
- **`securityToken` has no `requiredWhen`.** It is used (concatenated onto the password,
  `client.go:104`) but never validated, so it is optional in the backend contract.
- **`tokenUrl` modeled as a two-option select.** The editor renders a plain `Select` (no
  `allowCustomValue`) with Production/SandBox, and the TS `TokenURL` type is exactly those two values,
  so the generated settings spec enum of two values is faithful. The backend accepts any string —
  captured as discrepancy #4 rather than relaxing the schema.
- **`sandbox` retained with no UI.** Legacy jsonData field still read by both frontend init and
  backend `normalizeTokenUrl`; tagged `legacy` with a `description` (no `label`, since it has no UI).
- **Secure Socks Proxy excluded** (`SFConfigEditor.tsx:327-362`, writes `jsonData.enableSecureSocksProxy`).

## SDK plugin schema and k8s-style examples (`schema.go`)

`NewSchema()` assembles the `grafana-plugin-sdk-go` `pluginschema.PluginSchema` bundle from the
embedded `dsconfig.json`: the `jsonData` object becomes the OpenAPI settings `spec`, the six secrets
become `secureValues`, and no `secureJsonData` leaks into the spec.

`SettingsExamples()` provides the default plus one example per auth method and environment variant.
All secret placeholders are obviously-fake angle-bracket values (or `<redacted>` PEM bodies) so
GitHub push protection never trips:

| Example | Auth | Environment | `secureJsonData` |
| --- | --- | --- | --- |
| `""` (default) | Credentials (schema defaults) | Production | password, securityToken, clientID, clientSecret (empty) |
| `userCredentials` | Credentials | Production | password, securityToken, clientID, clientSecret |
| `userCredentialsSandbox` | Credentials | SandBox | password, securityToken, clientID, clientSecret |
| `jwt` | JWT bearer | Production | clientID, cert, privateKey |
| `jwtSandbox` | JWT bearer | SandBox | clientID, cert, privateKey |
| `legacySandboxFlag` | Legacy (no authType) | derived from `sandbox:true` | password, clientID, clientSecret |

## `LoadConfig` utility (`settings.go`)

`LoadConfig(ctx, settings) (Config, error)` runs the full three-phase flow and returns a
fully-defaulted, validated `Config`:

1. **Parse** — `json.Unmarshal(settings.JSONData, &cfg)` (empty JSONData is a parse error, mirroring
   `GetSettings` `settings.go:43-45`), then copy the decrypted secrets by known key into
   `DecryptedSecureJSONData`.
2. **`ApplyDefaults`** — mirror `normalizeAuthType`/`normalizeTokenUrl`: empty `authType` → `user`;
   empty `tokenUrl` → `TokenURLSandbox` when `sandbox` is true, else `TokenURLProd`.
3. **`Validate`** — mirror `Settings.Validate` (jwt: cert+privateKey; otherwise user+password+
   clientID+clientSecret), with errors joined so every problem surfaces at once.

Everything is logged via `backend.Logger.FromContext(ctx)` with `datasource_uid`, `datasource_name`,
and `plugin` labels. `(*Config).ApplyDefaults()` and `(Config).Validate()` stay exported for callers
that assemble a `Config` directly (provisioning preview, example round-trips, tests distinguishing
parse-level from policy-level errors).

## Upstream findings

Potential bugs, misleading UX, and consistency issues found while researching upstream. All are
preserved verbatim in the schema — the schema records what the plugin **does**, not what it should
do; these notes let reviewers reproduce each finding.

1. **Backend `Settings` declares secrets with jsonData json tags.** `settings.go:28-33` tags
   `Password`/`SecurityToken`/`ClientID`/`ClientSecret`/`Cert`/`PrivateKey` as `json:"password"` etc.,
   but these values live in `secureJsonData` and are overwritten from `DecryptedSecureJSONData`
   (`settings.go:48-72`). The json tags are effectively dead for the real storage model (they would
   only bind if a secret were mistakenly placed in `jsonData`). This entry models them as
   `secureJsonData` secrets to match the editor and intended storage.
2. **The backend's JWT auto-detection is effectively dead.** `normalizeAuthType` (`settings.go:82-90`)
   infers `jwt` from `Cert != "" && PrivateKey != "" && Password == ""`, but it runs *before* secrets
   are copied from `secureJsonData` (`settings.go:46` vs `:48-72`) and inspects the jsonData-level
   struct fields, which are empty for real (secure-stored) configs. So a missing `authType` always
   resolves to `user` on the backend. The frontend does the real detection from `secureJsonFields`
   (`SFConfigEditor.tsx:29-37`) — a frontend/backend inconsistency for legacy configs.
   `ApplyDefaults` mirrors the backend (defaults to `user`) and does not consult the decrypted
   secrets.
3. **JWT requires `clientID`/`user` at runtime but `Validate` does not enforce them.** `Validate`
   (`settings.go:103-110`) checks only cert+privateKey for jwt, yet `CreateJWT` (`client.go:96`,
   `jwt.go:57-64`) needs `clientID` (issuer) and `user` (subject). A jwt datasource missing either
   loads successfully but fails at token fetch. `requiredWhen` follows `Validate` (the hard load-time
   contract); the runtime need is recorded in the JWT instruction.
4. **`tokenUrl` is typed/edited as two values but the backend accepts any string.** The TS type is a
   two-value union and the editor `Select` offers only Production/SandBox, but `normalizeTokenUrl`
   (`settings.go:92-100`) and `fetchToken` (`client.go:84-92`) accept any string — the type comment
   (`types.ts:15-20`) documents an internal mocking escape hatch. The schema models the editor's two
   options (enum); provisioning could set an out-of-enum URL.
5. **`securityToken` is shown and stored but never validated.** It is concatenated onto the password
   (`client.go:104`) but absent from `Validate`, so a user-auth datasource without a security token
   loads fine (correct when the Salesforce org does not require one).
6. **Private Key placeholder says `-----BEGIN PRIVATE KEY-----`.** `SFConfigEditor.tsx:270` uses a
   PKCS#8 header, but a connected-app key may be PKCS#1 (`-----BEGIN RSA PRIVATE KEY-----`);
   `jwt.ParseRSAPrivateKeyFromPEM` accepts both. The schema preserves the editor placeholder verbatim;
   the JWT example uses the RSA header per secret-placeholder guidance.
7. **The datasource root `url` field is unused.** The API base is the `instance_url` returned by the
   token endpoint (`client.go:143-144,368-379`), not any configured URL — so `RootConfig` is blank
   and no root fields are carried on `Config`.
8. **The JWT `Generate` button mutates config as a side effect.** `SFConfigEditor.tsx:76-100`
   generates a self-signed cert/key pair, writes them to `secureJsonData.cert`/`privateKey`, sets
   `authType='jwt'`, downloads the cert, and submits the form. This is editor-only behavior with no
   storage field of its own; documented here rather than modeled.

## Validation performed

- `dsconfig.ParseAndResolveSchemaJSON` + `Schema.Validate()` (Go validator, via the `ConfigSchemaValid`
  conformance subtest) — passes.
- JSON Schema validation against [`dsconfig/schema.json`](../../dsconfig/schema.json) (draft-07,
  `additionalProperties: false`) — passes.
- `go test ./...` in the `registry` module — passes (schema round-trip, `SchemaArtifactInSync`,
  `SchemaSpecHasNoSecureJSON`, `JSONDataMatchesStruct`, `JSONDataTypesMatchStruct`,
  `SecureValuesMatchLoadSettings`, and `LoadConfig`/`ApplyDefaults`/`Validate` table tests).
- `go build ./...`, `go vet ./...`, `gofmt -l .` in `registry` — clean.
- `settings.ts`: `tsc --noEmit --strict` — clean.
- The pre-existing `dsconfig` and `schema` workspace modules still build.
