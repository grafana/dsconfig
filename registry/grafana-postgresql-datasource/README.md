# grafana-postgresql-datasource

Declarative configuration schema for the [PostgreSQL datasource plugin](https://github.com/grafana/grafana-postgresql-datasource) (`grafana-postgresql-datasource`, alias `postgres`).

## Upstream researched

- **Repo**: `github.com/grafana/grafana-postgresql-datasource`
- **Ref**: `main`
- **Commit SHA**: `c5d28c45780938cad1b24cb151194634c0a934d9` (2026-07-03)

Every value in [`dsconfig.json`](dsconfig.json) is traceable to a specific `file:line` in the
upstream repo at this SHA.

```bash
git clone https://github.com/grafana/grafana-postgresql-datasource
cd grafana-postgresql-datasource
git checkout c5d28c45780938cad1b24cb151194634c0a934d9
```

## Files

| File | Purpose |
| --- | --- |
| [`dsconfig.json`](dsconfig.json) | dsconfig v1 schema — single source of truth |
| [`settings.ts`](settings.ts) | TypeScript models: `RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig` |
| [`settings.go`](settings.go) | Go `Config` (root URL/User/Database + jsonData + `DecryptedSecureJSONData`), `LoadConfig` / `ApplyDefaults` / `Validate`, `TLSMode` / `TLSMethod` enums, `EffectiveDatabase()` |
| [`schema.go`](schema.go) | k8s-style SDK plugin schema + 7 `SettingsExamples` |
| [`settings_test.go`](settings_test.go) | Table tests for `LoadConfig`, `ApplyDefaults`, `Validate` |
| [`conformance_test.go`](conformance_test.go) | `schema.RunPluginTests` wrapper |
| `.gen.json` artifacts | Regenerate with `go generate ./...` |

## Sources researched

Read at the pinned upstream SHA:

| File | What was read |
| --- | --- |
| `src/plugin.json:3-7` | `pluginType` (`id: grafana-postgresql-datasource`), `pluginName` (`PostgreSQL`), `aliasIDs: ["postgres"]` |
| `src/types.ts:1-27` | `PostgresOptions extends SQLOptions`, `PostgresTLSModes` (`disable`/`require`/`verify-ca`/`verify-full`), `PostgresTLSMethods` (`file-path`/`file-content`), `SecureJsonData` |
| `src/configuration/ConfigurationEditor.tsx:38-52` | `postgresVersions` — 13 numeric-encoded PG versions (900 = 9.0 … 1500 = 15) |
| `src/configuration/ConfigurationEditor.tsx:54-59` | `useAutoDetectFeatures`, `useMigrateDatabaseFields` |
| `src/configuration/ConfigurationEditor.tsx:68-78` | `tlsModes` / `tlsMethods` option arrays (labels/values) |
| `src/configuration/ConfigurationEditor.tsx:108-112` | `DataSourceDescription` (`hasRequiredFields={true}`) |
| `src/configuration/ConfigurationEditor.tsx:116-123` | User Permissions Collapse (informational) |
| `src/configuration/ConfigurationEditor.tsx:129-148` | `Host URL` and `Database name` fields (both `required`) |
| `src/configuration/ConfigurationEditor.tsx:156-173` | `Username` (required, root `user`), `Password` (`SecretInput`, `secureJsonData.password`) |
| `src/configuration/ConfigurationEditor.tsx:175-201` | `TLS/SSL Mode` combobox with default `require` |
| `src/configuration/ConfigurationEditor.tsx:203-239` | `TLS/SSL Method` conditional combobox with default `file-path` |
| `src/configuration/ConfigurationEditor.tsx:243-345` | `TLS/SSL Auth Details` section — file-path inputs vs `TLSSecretsConfig` inline content |
| `src/configuration/ConfigurationEditor.tsx:349-451` | `Additional settings`: `Version`, `Min time interval`, `TimescaleDB`, `MaxOpenConnectionsField`, `MaxLifetimeField` |
| `src/configuration/ConfigurationEditor.tsx:447-449` | Excluded: `SecureSocksProxySettings` |
| `pkg/postgresql/postgres.go:80-115` | Backend settings assembly: root URL/User/Database + jsonData unmarshal + decrypted secrets |
| `pkg/postgresql/postgres.go:87-94` | Default `sqleng.JsonData` (from Grafana `SQL()` config): `MaxOpenConns`, `MaxIdleConns`, `ConnMaxLifetime`, `Timescaledb: false`, `ConfigurationMethod: "file-path"`, `SecureDSProxy: false` |
| `pkg/postgresql/postgres.go:101-104` | `jsonData.database` fallback to root `settings.Database` |
| `pkg/postgresql/postgres.go:347-354` | `applyPoolConfig`: `MaxConnLifetime`, `MaxConns` |

### External editor components

| Component | Version | Source | What was read |
| --- | --- | --- | --- |
| `SQLOptions` | `@grafana/sql@13.0.1` | grafana/grafana `v13.0.1` `packages/grafana-sql/src/types.ts:30-46` | Base jsonData shape — inherited but PostgreSQL overrides several fields via its own union |
| `useMigrateDatabaseFields` | `@grafana/sql@13.0.1` | `packages/grafana-sql/src/components/configuration/useMigrateDatabaseFields.ts` | Migrates root `database` → `jsonData.database` on first render |
| `MaxOpenConnectionsField`, `MaxLifetimeField` | `@grafana/sql@13.0.1` | `packages/grafana-sql/src/components/configuration/` | Labels `Max open` / `Max lifetime`, defaults from `config.sqlConnectionLimits` |
| `TLSSecretsConfig` | `@grafana/sql@13.0.1` | `packages/grafana-sql/src/components/configuration/TLSSecretsConfig.tsx` | Inline-PEM TLS secret fields (used only when `tlsConfigurationMethod === 'file-content'`) |
| `DataSourceDescription`, `ConfigSection`, `ConfigSubSection`, `EditorStack` | `@grafana/plugin-ui@0.13.1` | grafana/plugin-ui | Section headings |
| `Combobox`, `Field`, `Input`, `Switch`, `SecretInput` | `@grafana/ui@12.3.2+` | grafana/grafana `packages/grafana-ui` | UI primitives |

## Field provenance

| Schema `id` | Storage key | Target | Label source | Default / placeholder | Value type |
| --- | --- | --- | --- | --- | --- |
| `root_url` | `url` | root | `ConfigurationEditor.tsx:129` (`<Field label="Host URL" required>`) | `:135` (`placeholder="localhost:5432"`) | `settings.URL` |
| `jsonData_database` | `database` | jsonData | `:140` (`<Field label="Database name" required>`) | `:145` (`placeholder="Database"`) | `sqleng.JsonData.Database string`, `sql_engine.go:56` |
| `root_user` | `user` | root | `:156` (`<Field label="Username" required>`) | `:160` (`placeholder="Username"`) | `settings.User string` |
| `secureJsonData_password` | `password` | secureJsonData | `:165` (`<Field label="Password">`) | `:168` (`placeholder="Password"`) | `settings.DecryptedSecureJSONData["password"]` |
| `jsonData_sslmode` | `sslmode` | jsonData | `:180` (`<span>TLS/SSL Mode</span>`) with tooltip `:181-193` | Default `require` (`:197`); options from `tlsModes` at `:68-73` | `PostgresTLSModes`, `src/types.ts:3-8`; backend `sqleng.JsonData.Mode string`, `sql_engine.go:46` |
| `jsonData_tlsConfigurationMethod` | `tlsConfigurationMethod` | jsonData | `:209` (`<span>TLS/SSL Method</span>`) with tooltip `:210-227` | Default `file-path` (`:234`, `postgres.go:92`); options from `tlsMethods` at `:75-78` | `PostgresTLSMethods`, `src/types.ts:10-13`; backend `sqleng.JsonData.ConfigurationMethod string`, `sql_engine.go:47` |
| `jsonData_sslRootCertFile` | `sslRootCertFile` | jsonData | `:263` (`<span>TLS/SSL Root Certificate</span>`) | `:281` (`placeholder="TLS/SSL root cert file"`) | `PostgresOptions.sslRootCertFile string`, `src/types.ts:16`; backend `sqleng.JsonData.RootCertFile`, `sql_engine.go:49` |
| `jsonData_sslCertFile` | `sslCertFile` | jsonData | `:290` (`<span>TLS/SSL Client Certificate</span>`) | `:308` (`placeholder="TLS/SSL client cert file"`) | `PostgresOptions.sslCertFile string`, `src/types.ts:17`; backend `sqleng.JsonData.CertFile`, `sql_engine.go:50` |
| `jsonData_sslKeyFile` | `sslKeyFile` | jsonData | `:317` (`<span>TLS/SSL Client Key</span>`) | `:336` (`placeholder="TLS/SSL client key file"`) | `PostgresOptions.sslKeyFile string`, `src/types.ts:18`; backend `sqleng.JsonData.CertKeyFile`, `sql_engine.go:51` |
| `secureJsonData_tlsCACert` | `tlsCACert` | secureJsonData | `TLSSecretsConfig.tsx:71` | `:91` (`placeholder="-----BEGIN CERTIFICATE-----"`) | grafana-sql secret |
| `secureJsonData_tlsClientCert` | `tlsClientCert` | secureJsonData | `TLSSecretsConfig.tsx:32` | `:52` (`placeholder="-----BEGIN CERTIFICATE-----"`) | grafana-sql secret |
| `secureJsonData_tlsClientKey` | `tlsClientKey` | secureJsonData | `TLSSecretsConfig.tsx:109` | `:128` (`placeholder="-----BEGIN RSA PRIVATE KEY-----"`) | grafana-sql secret |
| `jsonData_postgresVersion` | `postgresVersion` | jsonData | `ConfigurationEditor.tsx:358` (`<span>Version</span>`) | Default `903` (`:371`); 13 options from `postgresVersions` `:38-52` | `PostgresOptions.postgresVersion number`, `src/types.ts:19` |
| `jsonData_timeInterval` | `timeInterval` | jsonData | `:382` (`<span>Min time interval</span>`) | `:399` (`placeholder="1m"`) | `SQLOptions.timeInterval string` |
| `jsonData_timescaledb` | `timescaledb` | jsonData | `:410` (`<span>TimescaleDB</span>`) | Default `false` (`:427`) | `PostgresOptions.timescaledb boolean`, `src/types.ts:20`; backend `sqleng.JsonData.Timescaledb`, `sql_engine.go:45` |
| `jsonData_maxOpenConns` | `maxOpenConns` | jsonData | `MaxOpenConnectionsField.tsx:23` | Default from `config.sqlConnectionLimits.maxOpenConns` | grafana-sql pool config |
| `jsonData_connMaxLifetime` | `connMaxLifetime` | jsonData | `MaxLifetimeField.tsx:22` | Default from `config.sqlConnectionLimits.connMaxLifetime` | grafana-sql pool config |

## Modeling decisions

- **Two parallel TLS credential surfaces**: `file-path` mode stores paths in **jsonData**
  (`sslRootCertFile` / `sslCertFile` / `sslKeyFile`); `file-content` mode stores inline PEMs
  in **secureJsonData** (`tlsCACert` / `tlsClientCert` / `tlsClientKey`). Only one set is
  active at a time — `dependsOn` on each field encodes which.
- **Root certificate only for verify-ca / verify-full**: the editor at
  `ConfigurationEditor.tsx:250-252` renders the root-cert secret only when sslmode is one
  of those two; the file-path root-cert field also only really applies to those modes even
  though the editor renders it unconditionally. Schema encodes the verify-only rule in
  `dependsOn`.
- **PostgreSQL alias ID**: `plugin.json` declares `aliasIDs: ["postgres"]` for backward
  compatibility with the legacy short id. The schema uses the canonical
  `grafana-postgresql-datasource` for `pluginType`; an instruction records the alias for
  provisioning consumers.
- **`postgresVersion` is UI-only**: `pkg/postgresql/*` never reads it — the actual protocol
  is autodetected. Value kept as a discriminator for the query builder only.
- **Secure Socks Proxy excluded** per AGENTS.md — `ConfigurationEditor.tsx:447-449`.
- **Only maxOpenConns and connMaxLifetime exposed**: unlike MySQL and MSSQL, this editor does
  not expose `maxIdleConns` / `maxIdleConnsAuto` (see the shortened `ConfigSubSection` at
  `:432-445` — only `MaxOpenConnectionsField` and `MaxLifetimeField`). Schema mirrors that.
- **Root-level URL/User/Database on Config** because the backend reads them directly from
  `backend.DataSourceInstanceSettings` (`pkg/postgresql/postgres.go:108-114`); same
  `EffectiveDatabase()` helper as MySQL.

## Upstream findings

1. **`useAutoDetectFeatures` may auto-write `postgresVersion`.** The `useAutoDetectFeatures`
   hook (`src/configuration/useAutoDetectFeatures.ts`) queries the connected database on
   config-page load and can overwrite `jsonData.postgresVersion` and `jsonData.timescaledb`
   without user interaction — potentially surprising if the user set them explicitly.
2. **`ssl*File` paths are trusted verbatim.** The backend reads `sslRootCertFile` /
   `sslCertFile` / `sslKeyFile` via the underlying pgx driver with no path sanitization; a
   provisioning caller with write access to the datasource config can point the Grafana
   process at any file it can read.
3. **`sslmode` default drift.** The editor defaults `sslmode` to `'require'`
   (`ConfigurationEditor.tsx:197`), but the backend does not set a default — an omitted
   `sslmode` in provisioning payloads gets an empty string, which pgx treats as `disable`.
   Our `ApplyDefaults` sets `require` to match the editor.
4. **`file-content` TLS writes secrets to disk at connection time.** `pkg/postgresql/tlsmanager.go`
   writes `secureJsonData.tls*` values to Grafana's data path as files before opening the
   connection. Operators should ensure the data-path directory has appropriate permissions.

## Validation performed

- Go validator + JSON Schema (draft-07) — pass
- `go test -race ./...` — pass
- `gofmt`, `go vet`, `go build` — clean
