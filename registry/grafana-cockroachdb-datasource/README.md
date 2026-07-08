# grafana-cockroachdb-datasource

Declarative configuration schema for the **CockroachDB datasource plugin** (`grafana-cockroachdb-datasource`).

CockroachDB is PostgreSQL wire-protocol compatible; the plugin connects with [`jackc/pgx`](https://github.com/jackc/pgx) v5 via `sqlds` on the CockroachDB default port `26257`. Despite the shared wire protocol, this plugin does **not** reuse `@grafana/sql` or the PostgreSQL config editor — it ships its own `ConfigEditor` and backend `Settings`, and it exposes auth methods (Kerberos, TLS/SSL) that PostgreSQL does not.

## Upstream researched

- **Monorepo**: `github.com/grafana/plugins-private`
- **Commit SHA**: `267f4937806ed6404b6628d13ae358a5d308e376` (Fri Jul 3 2026)
- **Plugin path**: `plugins/grafana-cockroachdb-datasource/`

Every value in [`dsconfig.json`](dsconfig.json) is traceable to a specific `file:line` in the upstream plugin at this SHA.

```bash
# Reproduce (monorepo already on disk; do NOT clone):
git -C <plugins-private> fetch origin
git -C <plugins-private> checkout 267f4937806ed6404b6628d13ae358a5d308e376
# then read plugins/grafana-cockroachdb-datasource/
```

## Files

| File | Purpose |
| --- | --- |
| [`dsconfig.json`](dsconfig.json) | dsconfig v1 schema — single source of truth |
| [`settings.ts`](settings.ts) | TypeScript models: `RootConfig` (blank), `JsonDataConfig`, `SecureJsonDataConfig` |
| [`settings.go`](settings.go) | Go `Config` (all-jsonData + `DecryptedSecureJSONData`), `LoadConfig` / `ApplyDefaults` / `Validate`, `AuthType` / `TLSMode` / `TLSMethod` / `SecureJsonDataKey` enums, `Password()` helper |
| [`schema.go`](schema.go) | k8s-style SDK plugin schema + 5 `SettingsExamples` |
| [`settings_test.go`](settings_test.go) | Table tests for `LoadConfig`, `ApplyDefaults`, `Validate`, secure keys |
| [`conformance_test.go`](conformance_test.go) | `schema.RunPluginTests` wrapper |
| `.gen.json` artifacts | Regenerate with `go generate ./...` |

## Sources researched

Read at the pinned monorepo SHA (paths relative to `plugins/grafana-cockroachdb-datasource/`):

| File | What was read |
| --- | --- |
| `src/plugin.json:3-5` | `type: datasource`, `pluginName` (`CockroachDB`), `pluginType` (`id: grafana-cockroachdb-datasource`) |
| `src/plugin.json:23-31,50` | Docs link; `grafanaDependency >=11.6.11-0` |
| `src/types.ts:7-12` | `CockroachTLSModes` (`disable`/`require`/`verify-ca`/`verify-full`) |
| `src/types.ts:14-17` | `CockroachTLSMethods` (`file-path`/`file-content`) |
| `src/types.ts:18-32` | `CockroachOptions extends SQLOptions` — incl. vestigial `postgresVersion`/`timescaledb` and excluded `enableSecureSocksProxy` |
| `src/types.ts:34-36` | `CockroachSecureJsonData { password }` |
| `src/types.ts:38-42` | `CockroachAuthenticationType` (`SQL Authentication`/`Kerberos Authentication`/`TLS/SSL Authentication`) |
| `src/components/ConfigEditor/ConfigEditor.tsx:71-75` | `DataSourceDescription` (`hasRequiredFields={true}`) |
| `src/components/ConfigEditor/ConfigEditor.tsx:77-99` | Connection section: `Host URL` (required, `localhost:26257`), `Database` (required, `defaultdb`) — both `jsonData` |
| `src/components/ConfigEditor/ConfigEditor.tsx:101-141` | Authentication: auth-type `Select` (placeholder `Choose your authentication method`), `User`, `Password` (SQL/TLS only) |
| `src/components/ConfigEditor/ConfigEditor.tsx:142-192` | `TLS/SSL Method` select (default `file-content`), shown for TLS auth when `sslmode != disable` |
| `src/components/ConfigEditor/ConfigEditor.tsx:193-284` | TLS cert fields — `TLSSecretsConfig` (file-content) vs file-path inputs |
| `src/components/ConfigEditor/ConfigEditor.tsx:286-336` | `Additional settings` (collapsible): `ConnectionLimits`, `krb5 config file path`, `TLS/SSL Mode` |
| `src/components/ConfigEditor/ConfigEditor.tsx:337-369` | Excluded: `Secure Socks Proxy` (gated on `config.secureSocksDSProxyEnabled`) |
| `src/components/ConfigEditor/Kerberos.tsx:22-49` | `Credential cache path` (required, `/tmp/krb5cc_1000`), `Kerberos server name` (optional, default `postgres`) |
| `src/components/ConfigEditor/ConnectionLimits.tsx:17-21,96-246` | `Max open` (5), `Auto max idle` (false), `Max idle` (2), `Max lifetime` (300), `Query timeout` (30) + tooltips |
| `src/components/ConfigEditor/TLSSecretsConfig.tsx:25-107` | Inline PEM secrets: `tlsCACert`, `tlsClientCert`, `tlsClientKey` |
| `pkg/plugin/settings.go:27-41` | Backend `Settings` struct — url/user/database/authType/pool all json-tagged (jsonData) |
| `pkg/plugin/settings.go:54-68` | `isValid` — url/user/database required; password required unless `authType == "Kerberos Authentication"` |
| `pkg/plugin/settings.go:115-144` | `generateKerberosConnectionString` — `sslmode=require authenticator=krb5 krb5-configfile=… krb5-credcachefile=…` + optional `krbsrvname` |
| `pkg/plugin/settings.go:146-222` | `generateTLSConfig` — file-path vs file-content, then DSN with `sslmode=<configDetails["sslmode"]>` |
| `pkg/plugin/settings.go:245-274` | `LoadSettings` — `json.Unmarshal(config.JSONData)`, `Password = DecryptedSecureJSONData["password"]`, pool defaults + query-timeout clamp |
| `pkg/plugin/driver.go:20-110` | `Connect` — TLS auth branch (`AuthTypeTLS`, `:22`) vs `generateConnectionString`; Secure Socks Proxy dialer; pool config |
| `pkg/plugin/tlsmanager.go:24-50` | `IsValidFilePathTLS` / `IsValidFileContentTLS` — all three paths/secrets required |
| `pkg/plugin/tlsmanager.go:106-139` | `GenerateTLSFileContentPaths` — writes inline PEMs to `<dataPath>/tls/<uid>generatedTLSCerts/` |
| `pkg/kerberos/kerberos.go:10-28` | `Auth { credentialCache, configFilePath }`, `GetKerberosSettings` (reads from jsonData) |
| `pkg/plugin/settings_test.go:10-146` | Backend `LoadSettings` expectations (defaults, query-timeout bounds, Kerberos-no-password) |

### External editor components

| Component | Version | Source | What was read |
| --- | --- | --- | --- |
| `SQLOptions`, `SQLQuery`, `SQLConnectionLimits` | `@grafana/plugin-ui` `^0.13.1` (catalog; resolved `0.13.1`) | [grafana/plugin-ui](https://github.com/grafana/plugin-ui) | Base `jsonData` shape (`CockroachOptions extends SQLOptions`) + connection-limit field keys (`maxOpenConns`/`maxIdleConns`/`connMaxLifetime`) |
| `ConfigSection`, `ConfigSubSection`, `DataSourceDescription` | `@grafana/plugin-ui` `^0.13.1` (resolved `0.13.1`) | grafana/plugin-ui | Section headings (`Connection`, `Authentication`, `Additional settings`, `Connection limits`) |
| `Field`, `Input`, `Select`, `SecretInput`, `SecretTextArea`, `Switch`, `Tooltip`, `Icon`, `Label`, `Stack`, `InlineLabel` | `@grafana/ui` `^11.6.7` (catalog; resolved `11.6.14`) | grafana/grafana `packages/grafana-ui` | UI primitives (labels, placeholders, tooltips) |
| `updateDatasourcePluginJsonDataOption`, `onUpdateDatasourceSecureJsonDataOption`, `updateDatasourcePluginResetOption`, `onUpdateDatasourceSecureJsonDataOption` | `@grafana/data` `^11.6.7` (resolved `11.6.14`) | grafana/grafana `packages/grafana-data` | Editor write helpers (which storage key each control writes) |
| `config` (`config.sqlConnectionLimits`, `config.secureSocksDSProxyEnabled`) | `@grafana/runtime` `^11.6.7` (resolved `11.6.14`) | grafana/grafana `packages/grafana-runtime` | Dynamic default placeholders for the pool fields; socks-proxy gating |

`@grafana/*` deps are declared as `catalog:` in the plugin's `package.json`; versions resolved from `plugins-private/.yarnrc.yml` catalog / `yarn.lock`.

## Field provenance

All storage fields are `jsonData` except the four secrets. There are **no root-level fields** (see Modeling decisions).

| Schema `id` | Storage key | Target | Editor label source | Default / placeholder | Read-by-backend |
| --- | --- | --- | --- | --- | --- |
| `jsonData_url` | `url` | jsonData | `ConfigEditor.tsx:78` (`<Field label="Host URL" required>`) | `:81` (`placeholder="localhost:26257"`) | Yes — `settings.go:31,247` |
| `jsonData_database` | `database` | jsonData | `:89` (`<Field label="Database" required>`) | `:92` (`placeholder="defaultdb"`) | Yes — `settings.go:29` |
| `jsonData_authType` | `authType` | jsonData | `:102-111` (`Select`, no label; placeholder `Choose your authentication method`) | options from `CockroachAuthenticationType` | Yes — `settings.go:39` |
| `jsonData_user` | `user` | jsonData | `:114` (`<Field label="User" required>`) | `:117` (`placeholder="User"`) | Yes — `settings.go:32` |
| `secureJsonData_password` | `password` | secureJsonData | `:128` (`<Field label="Password">`) — not marked required | `:131` (`placeholder="Password"`) | Yes — `settings.go:251` |
| `jsonData_credentialCache` | `credentialCache` | jsonData | `Kerberos.tsx:23` (`label="Credential cache path"` `required`) | `:29` (`placeholder="/tmp/krb5cc_1000"`) | Yes — `kerberos.go:20`, `settings.go:135` |
| `jsonData_kerberosServerName` | `kerberosServerName` | jsonData | `Kerberos.tsx:39` (`label="Kerberos server name"`) + description `:40` | `:43` (`placeholder="postgres"`) | Yes — `settings.go:38,139` |
| `jsonData_tlsConfigurationMethod` | `tlsConfigurationMethod` | jsonData | `ConfigEditor.tsx:149` (`<span>TLS/SSL Method</span>`) + tooltip `:152-163` | Default `file-content` (`:177-181`) | Yes — `settings.go:158,169` |
| `jsonData_sslRootCertFile` | `sslRootCertFile` | jsonData | `:202` (`<span>TLS/SSL Root Certificate</span>`) + tooltip `:205-208` | `:221` (`placeholder="TLS/SSL root cert file"`) | Yes — `settings.go:164`, `tlsmanager.go:26` |
| `jsonData_sslCertFile` | `sslCertFile` | jsonData | `:230` (`<span>TLS/SSL Client Certificate</span>`) + tooltip `:233-236` | `:248` (`placeholder="TLS/SSL client cert file"`) | Yes — `settings.go:165`, `tlsmanager.go:29` |
| `jsonData_sslKeyFile` | `sslKeyFile` | jsonData | `:257` (`<span>TLS/SSL Client Key</span>`) + tooltip `:260-264` | `:276` (`placeholder="TLS/SSL client key file"`) | Yes — `settings.go:166`, `tlsmanager.go:32` |
| `secureJsonData_tlsCACert` | `tlsCACert` | secureJsonData | `TLSSecretsConfig.tsx:29` (`<span>TLS/SSL Root Certificate</span>`) | `:40` (`placeholder="-----BEGIN CERTIFICATE-----"`) | Yes — `tlsmanager.go:40,108` |
| `secureJsonData_tlsClientCert` | `tlsClientCert` | secureJsonData | `TLSSecretsConfig.tsx:55` | `:70` (`placeholder="-----BEGIN CERTIFICATE-----"`) | Yes — `tlsmanager.go:43,109` |
| `secureJsonData_tlsClientKey` | `tlsClientKey` | secureJsonData | `TLSSecretsConfig.tsx:86` | `:97` (`placeholder="-----BEGIN RSA PRIVATE KEY-----"`) | Yes — `tlsmanager.go:46,110` |
| `jsonData_maxOpenConns` | `maxOpenConns` | jsonData | `ConnectionLimits.tsx:102` (`<span>Max open</span>`) + tooltip `:106-110` | Default `5` (`:17,254-256`) | Yes — `settings.go:33,254`, `driver.go:116` |
| `jsonData_maxIdleConnsAuto` | `maxIdleConnsAuto` | jsonData | `ConnectionLimits.tsx:133` (`<span>Auto max idle</span>`) + tooltip `:136-140` | Default `false` (`:30`) | **No — frontend-only** |
| `jsonData_maxIdleConns` | `maxIdleConns` | jsonData | `ConnectionLimits.tsx:156` (`<span>Max idle</span>`) + tooltip `:159-164` | Default `2` (`:18,257-259`) | Yes — `settings.go:34,257`, `driver.go:117` |
| `jsonData_connMaxLifetime` | `connMaxLifetime` | jsonData | `ConnectionLimits.tsx:191` (`<span>Max lifetime</span>`) + tooltip `:194-197` | Default `300` (`:19,260-262`) | Yes — `settings.go:35,260`, `driver.go:118` |
| `jsonData_queryTimeout` | `queryTimeout` | jsonData | `ConnectionLimits.tsx:220` (`<span>Query timeout</span>`) + tooltip `:222-226` | Default `30`, clamp 5–600 (`:21,263-272`) | Yes — `settings.go:40,263`, `driver.go:201` |
| `jsonData_configFilePath` | `configFilePath` | jsonData | `ConfigEditor.tsx:290` (`label="krb5 config file path"`) + description `:291-299` | Default `/etc/krb5.conf` (`:305`) | Yes — `kerberos.go:21`, `settings.go:135` |
| `jsonData_sslmode` | `sslmode` | jsonData | `ConfigEditor.tsx:314` (`<span>TLS/SSL Mode</span>`) + tooltip `:317-320` | Default `require` (`:331`) | Yes — `settings.go:218` |

## Frontend-only and backend-only settings

- **Frontend-only**: `jsonData.maxIdleConnsAuto` — written by `ConnectionLimits.tsx:70-92` to keep `maxIdleConns` synced to `maxOpenConns`, but the backend `Settings` struct has no such field (`settings.go:27-41`). Modeled (tagged `frontend-only`) because it is a real stored jsonData field.
- **Backend-only**: none — every backend `Settings` field is surfaced somewhere in the editor.
- **Excluded (present in editor, deliberately not in schema)**: `jsonData.enableSecureSocksProxy` (Secure Socks Proxy toggle, `ConfigEditor.tsx:351-366`, consumed at `driver.go:64-81`). Omitted from every registry entry per AGENTS.md.
- **Dead/vestigial type fields (not modeled)**: `jsonData.postgresVersion` and `jsonData.timescaledb` are declared on `CockroachOptions` (`types.ts:28-29`) but the editor never writes them and the backend never reads them — leftovers from the PostgreSQL-derived type. Excluded.

## Where the types are defined

| Type | Where |
| --- | --- |
| `Settings` (backend, upstream) | `pkg/plugin/settings.go:27-41` |
| `TLSConfigMethod`, `tlsSettings` (backend, upstream) | `pkg/plugin/settings.go:43-52` |
| `kerberos.Auth` / `kerberos.Lookup` (backend, upstream) | `pkg/kerberos/kerberos.go:10-22` |
| `CockroachOptions` (frontend, upstream) | `src/types.ts:18-32` |
| `CockroachTLSModes` / `CockroachTLSMethods` / `CockroachAuthenticationType` (frontend, upstream) | `src/types.ts:7-17,38-42` |
| `CockroachSecureJsonData` (frontend, upstream) | `src/types.ts:34-36` |
| `SQLOptions` / `SQLConnectionLimits` (frontend, external) | `@grafana/plugin-ui@0.13.1` — base of `CockroachOptions`; only wired fields modeled |
| `Config` (this entry) | `settings.go` — all-jsonData mirror of upstream `Settings` (root fields intentionally absent) |

## Modeling decisions

- **Everything is `jsonData`; there are no root fields.** Unlike `grafana-postgresql-datasource` / `grafana-yugabyte-datasource` (which read root `url`/`user`), this plugin's `LoadSettings` (`settings.go:247`) unmarshals `config.JSONData` for `url`/`user`/`database`/`authType`/… and only pulls `password` from `DecryptedSecureJSONData`. So `RootConfig` is a blank object and `Config` carries no `json:"-"` root fields.
- **`authType` is a stored discriminator (`role: auth.discriminator`), not virtual.** The editor writes `jsonData.authType` directly (`ConfigEditor.tsx:61-68`); no derived/virtual selector is needed (contrast the GitHub entry's virtual `selectedLicense`). Its option values are the full label strings (`"SQL Authentication"`, etc.), verbatim from `CockroachAuthenticationType`.
- **`dependsOn` mirrors editor visibility; `requiredWhen` mirrors the backend contract.** Password `requiredWhen` uses the two auth modes where it is both shown and required (`SQL`/`TLS`), aligning with visibility; the backend's underlying rule is "required unless Kerberos" (`settings.go:64`) and `Validate` encodes that.
- **TLS credential surfaces are split by method.** `file-path` → `jsonData.sslRootCertFile`/`sslCertFile`/`sslKeyFile`; `file-content` → `secureJsonData.tlsCACert`/`tlsClientCert`/`tlsClientKey`. Each side's `dependsOn`/`requiredWhen` encodes `authType == 'TLS/SSL Authentication' && sslmode != 'disable' && tlsConfigurationMethod == <method>`, matching both the editor (`ConfigEditor.tsx:193-284`) and the backend validators (`tlsmanager.go:25-50`).
- **`tlsConfigurationMethod` default is `file-content`** (`ConfigEditor.tsx:177-181`) — deliberately different from PostgreSQL's `file-path`. `ApplyDefaults` reflects this.
- **`ApplyDefaults` mirrors the backend pool defaults verbatim** (5/2/300/30 + query-timeout clamp to [5,600], `settings.go:254-272`) and additionally applies the editor-parity discriminator defaults (`sslmode='require'`, `tlsConfigurationMethod='file-content'`) that the backend does *not* set. Kept curated (only those zero-valued fields).
- **`"TLS/SSL Auth Details"` is a schema-only group.** In the editor these cert fields render inline between the Authentication and Additional settings sections with no wrapping `ConfigSection`; the group gives provisioning/UX consumers a coherent bucket without inventing an editor label.
- **`sslmode` and `configFilePath` are grouped under `Additional settings`** to match the editor, even though `sslmode` logically gates the TLS cert fields that appear earlier (see Upstream findings).
- **Secrets live only in `DecryptedSecureJSONData`.** `Config` deliberately omits the upstream `Settings.Password string json:"password"` field: that tag is dead (see finding 2) and carrying it would break the schema↔struct jsonData parity check. `Password()` exposes the decrypted value.

## Settings examples matrix

| Key | Summary | authType | sslmode / method | Secrets |
| --- | --- | --- | --- | --- |
| `""` | Default configuration | (unset) | `require` / `file-content` + pool defaults | `password:""` |
| `sqlAuth` | SQL Authentication | `SQL Authentication` | — | `password` |
| `kerberosAuth` | Kerberos Authentication | `Kerberos Authentication` | — | none (no password) |
| `tlsVerifyFullFilePath` | TLS, file-path certs | `TLS/SSL Authentication` | `verify-full` / `file-path` | `password` |
| `tlsVerifyCAFileContent` | TLS, inline PEM certs | `TLS/SSL Authentication` | `verify-ca` / `file-content` | `password`, `tlsCACert`, `tlsClientCert`, `tlsClientKey` |

## Upstream findings / discrepancies

1. **Connection fields live in `jsonData`, not root.** `url`/`user`/`database` are unmarshalled from `config.JSONData` (`settings.go:247`); root `settings.URL`/`settings.User` are never read. This diverges from PostgreSQL/MySQL/Yugabyte and means provisioning payloads must put these under `jsonData`.
2. **`Settings.Password` json tag is dead.** The struct declares `Password string json:"password,omitempty"` (`settings.go:30`) but `LoadSettings` immediately overwrites it with `DecryptedSecureJSONData["password"]` (`settings.go:251`). Any `jsonData.password` would be silently discarded; the password is effectively `secureJsonData`-only.
3. **Pool `0` sentinels are clobbered.** `LoadSettings` replaces a stored `0` with the defaults 5/2/300 (`settings.go:254-262`), contradicting the tooltips that say `0` = "no limit" / "reused forever". `queryTimeout` `0` → `30`, then clamped to `[5,600]` (`:263-272`). `ApplyDefaults` reproduces this exactly.
4. **`tlsConfigurationMethod` is written during render.** The default is applied by calling `updateDatasourcePluginJsonDataOption(...)` *inside* the `Select`'s `value` prop (`ConfigEditor.tsx:177-181`) — a render-time state mutation. The effect is a `file-content` default, but the pattern is fragile.
5. **`sslmode` control is placed after the fields it gates.** The `TLS/SSL Mode` select renders in `Additional settings` (`ConfigEditor.tsx:309-336`), *below* the TLS method + certificate fields it governs (rendered at `:142-284`). Users configure certs before seeing the mode that decides whether they apply.
6. **`generateTLSConfig` ignores `sslmode` when validating certs.** For TLS auth it branches only on `tlsConfigurationMethod` (`settings.go:158-181`); a TLS-auth datasource with `sslmode=disable` and no configured method/certs reaches `os.ReadFile("")` and fails at connect time. The editor hides cert fields when `sslmode=disable`, so this state is only reachable via provisioning. The schema gates cert `requiredWhen` on `sslmode != 'disable'` to match the editor.
7. **Kerberos config/cache paths are interpolated without escaping.** `generateKerberosConnectionString` (`settings.go:130-136`) `fmt.Sprintf`s `ConfigFilePath`/`CredentialCache` raw into the connection string (while user/host/db/port go through `escape()`), a minor injection surface for provisioning callers.
8. **`postgresVersion` / `timescaledb` are dead type fields.** Declared on `CockroachOptions` (`types.ts:28-29`) but neither written by the editor nor read by the backend — leftovers from the PostgreSQL-derived type. Excluded from the schema.

## Validation performed

- `go generate ./...` inside this entry — regenerates `schema.gen.json` / `settings.gen.json` / `settings.examples.gen.json`; conformance suite passes.
- `dsconfig.ParseAndResolveSchemaJSON` + `Schema.Validate()` (via `RunPluginTests`) — pass.
- JSON Schema validation of `dsconfig.json` against `dsconfig/schema.json` (draft-07, strict `additionalProperties:false`) — pass.
- Conformance guards: no `secureJsonData` in the settings spec; `secureValues` = `[password, tlsCACert, tlsClientCert, tlsClientKey]`; every `jsonData` key matches the `Config` json tags (both directions); `""` default example present; every example carries valid secure keys.
- `gofmt -l .`, `go vet ./...`, `go build ./...`, `go test ./...` inside `registry/` — clean across every entry.
- `tsc --noEmit --strict` on `settings.ts` — pass.
- Pre-existing `dsconfig` and `schema` workspace modules still `go build` — pass.
