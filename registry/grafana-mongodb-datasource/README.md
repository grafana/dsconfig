# grafana-mongodb-datasource

dsconfig registry entry for the **MongoDB** Grafana datasource plugin
(`grafana-mongodb-datasource`).

## Files

| File | Purpose |
| --- | --- |
| `dsconfig.json` | dsconfig v1 schema — the single source of truth |
| `settings.ts` | TypeScript models: `RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig` |
| `settings.go` | Flat Go `Config` (jsonData + root basic-auth fields + `DecryptedSecureJSONData`), `AuthType`/`SecureJsonDataKey` enums, `LoadConfig`/`ApplyDefaults`/`Validate` |
| `schema.go` | Embeds `dsconfig.json`; `ConfigSchema()`, `NewSchema()`, `SettingsExamples()` |
| `conformance_test.go` | `schema.RunPluginTests` wrapper (guard-rail suite / artifact generator) |
| `settings_test.go` | `LoadConfig` / `ApplyDefaults` / `Validate` / helper / example-shape tests |
| `schema.gen.json`, `settings.gen.json`, `settings.examples.gen.json` | Generated artifacts (`go generate ./...`) |

## Source & reproduce

Researched against the on-disk monorepo (not cloned):

- Repo: `github.com/grafana/plugins-private`
- Commit SHA: `267f4937806ed6404b6628d13ae358a5d308e376`
- Plugin path: `plugins/grafana-mongodb-datasource/`

```sh
git -C <plugins-private> rev-parse HEAD          # 267f4937806ed6404b6628d13ae358a5d308e376
# read plugins/grafana-mongodb-datasource/{src,pkg,docs}
```

Plugin identity from `src/plugin.json`: `id` = `grafana-mongodb-datasource` (`:5`),
`name` = `MongoDB` (`:3`), docs URL from `info.links[]` (`:20`) →
`https://grafana.com/docs/plugins/grafana-mongodb-datasource`. Plugin version `1.27.2`
(`package.json:3`).

## Sources researched (file:line)

Frontend (`plugins/grafana-mongodb-datasource/src`):

- `plugin.json:3,5,20` — name, id, docs URL.
- `types.ts:29-46` — `MongoDbJsonData` (frontend jsonData shape).
- `types.ts:52-57` — `SecureSettings` (frontend secureJsonData shape).
- `components/ConfigEditor.tsx` — editor: Connection string (`:145-179`), Additional Settings
  (Query Syntax Validation `:191-207`, Secure Socks Proxy `:209-246` (excluded), TLS CA Key File
  Password `:248-268`, Backend Response Rows Limit `:270-289`), and the legacy migration
  `parseAndConvertOptions` (`:296-313`).
- `components/AuthConfig.tsx` — auth method picker (`:13`), method labels/descriptions
  (`:15-20,46-55`), user/password tooltip overrides (`:43-44`), `onAuthMethodSelect` writing
  `jsonData.authType` (`:28-36`), selected method from `jsonData.authType` (`:57`).
- `components/KerberosAuthConfig.tsx` — Kerberos inputs & tooltips/placeholders (`:10-13,64-112`).

Backend (`plugins/grafana-mongodb-datasource/pkg`):

- `models/settings.go:15-45` — `Settings` struct + json tags.
- `models/settings.go:50-162` — `LoadSettings`: `responseRowsLimit` default (`:56-58`), authType
  coercion (`:66-71`), basic-auth warnings (`:76-83`), legacy `user`/`password` migration
  (`:100-107`), root `basicAuthUser`/`basicAuthPassword` override (`:109-136`), Kerberos-enabled
  detection (`:138-141`), `skipTLSValidation` → `tlsSkipVerify` (`:143-148`), `mapstructure.Decode`
  of secure data (`:150-158`).
- `models/settings_test.go` — legacy/new/kerberos/invalid load behaviors.
- `datasource/client.go` — consumption: connection required (`:134-137`), authType switch
  (`:139-161`), basic-auth credentials injected into the URI (`:62-72`), Kerberos user (`:74-80`),
  TLS CA (`:391-405`), TLS client cert/key + `TLSCertificateKeyFilePassword` + `ServerName`
  (`:407-431`).
- `docs/sources/configure.md` — connection (`:60-67`), auth methods (`:69-100`), TLS (`:102-116`),
  additional settings (`:118-132`), provisioning YAML/Terraform (`:150-212`).

Backend module: `github.com/grafana/mongodb-datasource`, `grafana-plugin-sdk-go v0.292.0`,
`github.com/mitchellh/mapstructure v1.5.0`, `go.mongodb.org/mongo-driver v1.17.9`,
`github.com/youmark/pkcs8` (encrypted-key decrypt).

## External components (catalog versions)

Resolved from `.yarnrc.yml` `catalog:` (workspace root) unless plugin-pinned in `package.json`:

- `@grafana/plugin-ui` `^0.13.1` — `Auth` + `convertLegacyAuthProps`
  (`dist/esm/components/ConfigEditor/Auth/utils.js`, `.../auth-method/BasicAuth.js`, `.../types.js`).
  Verified against the published package: `AuthMethod.NoAuth = "NoAuth"`, `AuthMethod.BasicAuth =
  "BasicAuth"`; default method labels — NoAuth **"No Authentication"** / description "Data source is
  available without authentication"; the basic-auth `BasicAuth` component renders **"User"** /
  **"Password"** labels and placeholders. `convertLegacyAuthProps` stores the basic-auth username at
  the **root** `basicAuthUser` field and the password at `secureJsonData.basicAuthPassword`.
- `@grafana/ui` `^11.6.7` — `Input`, `Switch`, `SecretInput`, `InlineField`.
- `@grafana/data` `^11.6.7` — `DataSourceJsonData`, `DataSourceSettings` (root `basicAuth` /
  `basicAuthUser`), editor prop types.
- `@grafana/runtime` `^11.6.7`, `@grafana/schema` `^11.6.7`, `@grafana/e2e-selectors` `^11.6.7`
  (plugin-pinned, `package.json:52`).

## Field provenance / inventory

| Schema field ID | storage key | target | editor label | read by backend |
| --- | --- | --- | --- | --- |
| `jsonData_connection` | `connection` | jsonData | Connection string | yes (`client.go:134`, `settings.go:17`) |
| `jsonData_authType` | `authType` | jsonData | Authentication method (radio) | yes (`settings.go:66`, `client.go:139`) |
| `root_basicAuth` | `basicAuth` | root | — (managed) | yes → `BasicAuthEnabled` (`settings.go:98`) |
| `root_basicAuthUser` | `basicAuthUser` | root | User | yes (`settings.go:111-119`) |
| `secureJsonData_basicAuthPassword` | `basicAuthPassword` | secureJsonData | Password | yes (`settings.go:122`) |
| `jsonData_kerberosUser` | `kerberosUser` | jsonData | User | yes (`settings.go:23`) |
| `secureJsonData_kerberosPassword` | `kerberosPassword` | secureJsonData | Password | yes (`settings.go:140`) |
| `jsonData_keyTabFilePath` | `keyTabFilePath` | jsonData | KeyTab file path | yes (`settings.go:25`) |
| `jsonData_globalCcacheFilePath` | `globalCcacheFilePath` | jsonData | Global ccache file path | yes (`settings.go:26`) |
| `jsonData_ccacheLookupFile` | `ccacheLookupFile` | jsonData | Ccache lookup file | yes (`settings.go:27`) |
| `jsonData_validate` | `validate` | jsonData | Enable syntax validation | **no** (frontend-only) |
| `secureJsonData_tlsCertificateKeyFilePassword` | `tlsCertificateKeyFilePassword` | secureJsonData | Password | yes (`client.go:419`) |
| `jsonData_responseRowsLimit` | `responseRowsLimit` | jsonData | Rows to Return | yes (`settings.go:37`) |
| `jsonData_serverName` | `serverName` | jsonData | — (provisioning) | yes (`client.go:428`) |
| `jsonData_tlsAuth` | `tlsAuth` | jsonData | — (provisioning) | yes (`client.go:407`) |
| `jsonData_tlsAuthWithCACert` | `tlsAuthWithCACert` | jsonData | — (provisioning) | yes (`client.go:391`) |
| `jsonData_tlsSkipVerify` | `tlsSkipVerify` | jsonData | — (provisioning/migration) | yes (`client.go:401`) |
| `secureJsonData_tlsCACert` | `tlsCACert` | secureJsonData | — (provisioning) | yes (`client.go:393`) |
| `secureJsonData_tlsClientCert` | `tlsClientCert` | secureJsonData | — (provisioning) | yes (`client.go:412`) |
| `secureJsonData_tlsClientKey` | `tlsClientKey` | secureJsonData | — (provisioning) | yes (`client.go:413`) |
| `jsonData_user` | `user` | jsonData | — (legacy) | yes → migrated (`settings.go:42,102`) |
| `jsonData_skipTLSValidation` | `skipTLSValidation` | jsonData | — (legacy) | yes → migrated (`settings.go:44,144`) |
| `jsonData_credentials` | `credentials` | jsonData | — (legacy) | **no** (frontend-only) |
| `secureJsonData_password` | `password` | secureJsonData | — (legacy) | yes → migrated (`settings.go:105`) |

### Frontend-only settings

- `validate` — drives real-time BSON syntax validation in the query editor (`ConfigEditor.tsx:192-206`);
  not present in the backend `Settings`.
- `credentials` — legacy flag read only by the editor migration (`ConfigEditor.tsx:299`, `types.ts:32`).

### Backend-only settings (no editor UI; provisioning or migration)

- `serverName`, `tlsAuth`, `tlsAuthWithCACert`, `tlsSkipVerify`, `tlsCACert`, `tlsClientCert`,
  `tlsClientKey` — read by `pkg/datasource/client.go` but the current `ConfigEditor.tsx` renders **no**
  TLS certificate UI (only the "TLS CA Key File Password" secret). Configured via provisioning
  (`docs/sources/configure.md:188-212`).
- `basicAuth` (root) — the standard Grafana enabled flag; the backend seeds `BasicAuthEnabled` from it
  and also forces it on when a username/password is present.
- `user`, `skipTLSValidation`, `password` — legacy pre-v1.9.0 fields migrated at load time.

## Where the types are defined

Frontend:

- `MongoDbJsonData` — `src/types.ts:29-46` (extends `DataSourceJsonData` from `@grafana/data`).
- `SecureSettings` — `src/types.ts:52-57`.
- Auth method value type: `AuthMethod` — `@grafana/plugin-ui` (`.../Auth/types.js`) supplies `NoAuth` /
  `BasicAuth`; `custom-Kerberos` is a plugin string literal (`AuthConfig.tsx:11`).
- Root basic-auth fields (`basicAuth`, `basicAuthUser`): `DataSourceSettings` / `DataSourceSettingsMeta`
  from `@grafana/data`, wired by `convertLegacyAuthProps` in `@grafana/plugin-ui`.

Backend:

- `Settings` — `pkg/models/settings.go:15-45`.
- Root fields (`BasicAuthEnabled`, `BasicAuthUser`, `DecryptedSecureJSONData`):
  `backend.DataSourceInstanceSettings` in `grafana-plugin-sdk-go`.

## Settings examples matrix

| Example key | authType | connection | secrets | notes |
| --- | --- | --- | --- | --- |
| `""` (default) | BasicAuth | — (empty) | `basicAuthPassword:""` | schema defaults; fails `Validate` (empty connection) |
| `credentials` | BasicAuth | host/db | `basicAuthPassword` | username at root `basicAuthUser` |
| `noAuth` | NoAuth | host/db | none (empty `secureJsonData`) | credential-free variant |
| `kerberos` | custom-Kerberos | `?authMechanism=GSSAPI` | `kerberosPassword` | custom build; not on Grafana Cloud |
| `tlsClientAuth` | BasicAuth | `?tls=true` | `basicAuthPassword`, `tlsCACert`, `tlsClientCert`, `tlsClientKey` | TLS via provisioning |
| `legacyCredentials` | (defaulted) | host/db | `password` | legacy `jsonData.user` + secure `password` |

All secret values are obviously-fake angle-bracket placeholders (`<your-password>`,
`<kerberos-password>`); PEM material uses `-----BEGIN …-----\n<redacted>\n-----END …-----`.

## Modeling decisions

- **Auth discriminator.** `jsonData.authType` (`role: auth.discriminator`, radio) with values
  `NoAuth` / `BasicAuth` / `custom-Kerberos` (proven from the backend at `settings.go:66` and
  `client.go:139-161`), default `BasicAuth`. Labels: `NoAuth` → "No Authentication" (library default),
  `BasicAuth` → "Credentials" (`AuthConfig.tsx:17`), `custom-Kerberos` → "Kerberos"
  (`AuthConfig.tsx:49`).
- **Basic auth storage.** Username → **root** `basicAuthUser` (`auth.basic.username`), password →
  `secureJsonData.basicAuthPassword` (`auth.basic.password`), proven by the backend
  (`settings.go:111-136`) and `@grafana/plugin-ui` `convertLegacyAuthProps`. The root `basicAuth`
  enabled flag is modeled as a backend/legacy field (`auth.basic.enabled`).
- **TLS secrets vs jsonData.** The backend `Settings` declares json tags for `tlsCACert`,
  `tlsClientCert`, `tlsClientKey` and `tlsCertificateKeyFilePassword`, but the config editor and
  provisioning store these in `secureJsonData`, and `mapstructure.Decode(DecryptedSecureJSONData,
  &settings)` (`settings.go:154`) overwrites them from the secure data — so they are modeled as
  **secureJsonData** secrets. The switches/hints `tlsAuth`, `tlsAuthWithCACert`, `tlsSkipVerify`,
  `serverName` remain jsonData (backend-only).
- **`responseRowsLimit`.** Stored as a string; no `defaultValue` in the schema because the editor
  merely displays a `100000` fallback (never persisted) while the backend defaults an empty value to
  `"10000"`. `Config.ApplyDefaults` uses the backend value `"10000"` for parity.
- **Legacy migration mirrored.** `LoadConfig` reproduces `LoadSettings`: `jsonData.user` →
  `BasicAuthUser`, root `basicAuthUser` overrides it, a modern/legacy password enables basic auth,
  and `skipTLSValidation` → `tlsSkipVerify`. `BasicAuthPassword()` prefers `basicAuthPassword` and
  falls back to legacy `password`; `KerberosEnabled()` mirrors `settings.go:138`.
- **`LoadConfig` guarantees.** parse (`json.Unmarshal` + legacy fallbacks + secret copy) →
  `ApplyDefaults` (authType coercion, `responseRowsLimit`) → `Validate` (connection required, auth
  type valid, TLS material present when enabled). `ApplyDefaults`/`Validate` stay exported for callers
  that build a `Config` directly.
- **Exclusions.** `jsonData.enableSecureSocksProxy` (Secure Socks Proxy) and PDC are omitted per
  registry policy.

## Potential upstream bugs / discrepancies

1. **Docs describe a TLS UI that does not exist.** `configure.md:102-116` documents a "TLS settings"
   section (ServerName, CA certificate, client certificate/key, skip validation) as editor fields, but
   `ConfigEditor.tsx` renders none of them — they are provisioning-only. Only "TLS CA Key File
   Password" is an actual editor secret.
2. **`responseRowsLimit` default mismatch.** Editor display fallback `100000` (`ConfigEditor.tsx:277`)
   and docs "The default is `100000`" (`configure.md:132`) vs backend `LoadSettings` default `"10000"`
   (`settings.go:57`).
3. **`responseRowsLimit` default is applied before unmarshal.** `settings.go:56-58` sets the default
   on a freshly zero-valued struct *before* `json.Unmarshal` (`:60`) runs, so the guard is effectively
   dead code that happens to work (unmarshal overrides when the key is present).
4. **Dual storage path for TLS material.** `tlsCACert`/`tlsClientCert`/`tlsClientKey`/
   `tlsCertificateKeyFilePassword` are declared as jsonData json tags yet read from `secureJsonData`
   via `mapstructure` (which wins). The jsonData path is effectively unused by the editor/provisioning.
5. **Zero-width spaces in a tooltip.** The Kerberos password tooltip (`KerberosAuthConfig.tsx:13`)
   contains two `U+200B` zero-width spaces before "Optional". Omitted from the schema description as an
   accidental artifact.
6. **Double space preserved.** The Ccache lookup file tooltip has "to  the JSON" (double space,
   `KerberosAuthConfig.tsx:108`), kept verbatim.
7. **Editor-required vs backend-required.** The editor marks Connection and the basic-auth
   User/Password required, but the backend only hard-fails on an empty connection (`client.go:134`);
   empty basic-auth credentials merely log a warning (`settings.go:76-83`). Modeled: `connection`
   `requiredWhen:"true"`; basic-auth credentials have `dependsOn` only.
8. **`basicAuth` flag not set by method selection.** The overridden `onAuthMethodSelect`
   (`AuthConfig.tsx:28-36`) writes only `jsonData.authType`; root `basicAuth` is set by the on-load
   migration (`ConfigEditor.tsx:299-307`) or provisioning, while the backend also derives
   `BasicAuthEnabled` from username/password presence.
9. **AuthType coercion.** Empty or unrecognized `authType` is coerced to `BasicAuth`
   (`settings.go:66-71`); mirrored in `ApplyDefaults`.

## Validation performed

- `dsconfig.ParseAndResolveSchemaJSON` + `Schema.Validate()` via the conformance suite — pass.
- Strict JSON Schema validation of `dsconfig.json` against `dsconfig/schema.json` (Draft-07,
  `additionalProperties:false`) — **VALID**.
- `go generate ./...` (regenerated `schema.gen.json`, `settings.gen.json`,
  `settings.examples.gen.json`) — pass.
- `gofmt -l .` (registry) — clean; `go vet ./...` (registry) — pass; `go build ./...` — pass.
- `go test ./...` (registry, all entries) — pass. This entry's `TestSchemaConformance` guards
  jsonData↔struct parity (both directions), jsonData type parity, secure-key parity, no
  `secureJsonData` in the settings spec, and artifact drift.
- `tsc --noEmit --strict --skipLibCheck settings.ts` (typescript 5.4.5) — pass.
- `dsconfig` and `schema` workspace modules — `go build ./...` pass.
- Secret scan of all committed files — only angle-bracket placeholders and `<redacted>` PEM bodies;
  no realistic tokens/credentials.
