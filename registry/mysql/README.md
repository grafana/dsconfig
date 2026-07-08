# mysql

Declarative configuration schema for the [MySQL datasource plugin](https://github.com/grafana/grafana-mysql-datasource) (`mysql`).

## Upstream researched

- **Repo**: `github.com/grafana/grafana-mysql-datasource`
- **Ref**: `main`
- **Commit SHA**: `98f55a8ee6881d02ef4d2df5c73ac9860aae69fd` (2026-07-02)

Every value in [`dsconfig.json`](dsconfig.json) is traceable to a specific `file:line` in the
upstream repo at this SHA. To reproduce:

```bash
git clone https://github.com/grafana/grafana-mysql-datasource
cd grafana-mysql-datasource
git checkout 98f55a8ee6881d02ef4d2df5c73ac9860aae69fd
```

## Files

| File | Purpose |
| --- | --- |
| [`dsconfig.json`](dsconfig.json) | dsconfig v1 schema — single source of truth |
| [`settings.ts`](settings.ts) | TypeScript models: `RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig` |
| [`settings.go`](settings.go) | Go `Config` (root URL/User/Database + jsonData fields + `DecryptedSecureJSONData`), `LoadConfig` / `ApplyDefaults` / `Validate`, `EffectiveDatabase()` |
| [`schema.go`](schema.go) | k8s-style SDK plugin schema + 7 `SettingsExamples` |
| [`settings_test.go`](settings_test.go) | Table tests for `LoadConfig`, `ApplyDefaults`, `Validate`, `EffectiveDatabase` |
| [`conformance_test.go`](conformance_test.go) | `schema.RunPluginTests` wrapper |
| `.gen.json` artifacts | Regenerate with `go generate ./...` |

## Sources researched

Read at the pinned upstream SHA:

| File | What was read |
| --- | --- |
| `src/plugin.json:3-4` | `pluginType` (`id: mysql`), `pluginName` (`MySQL`) |
| `src/plugin.json:26-29` | `info.links[]` — no `Docs` link; docURL derived from Grafana docs |
| `src/types.ts:1-7` | `MySQLOptions extends SQLOptions` (adds only `allowCleartextPasswords`) |
| `src/configuration/ConfigurationEditor.tsx:34` | `useMigrateDatabaseFields(props)` — migrates root `database` → `jsonData.database` |
| `src/configuration/ConfigurationEditor.tsx:56-60` | `DataSourceDescription` (`hasRequiredFields={true}`) |
| `src/configuration/ConfigurationEditor.tsx:64-71` | "User Permission" `Collapse` (informational, not modeled as a field) |
| `src/configuration/ConfigurationEditor.tsx:77-86` | `Host URL` field: `required`, `placeholder="localhost:3306"`, writes root-level `url` |
| `src/configuration/ConfigurationEditor.tsx:88-96` | `Database name` field: `placeholder="Database"`, writes `jsonData.database` |
| `src/configuration/ConfigurationEditor.tsx:104-121` | `Username` (required, root `user`) + `Password` (`SecretInput`, `secureJsonData.password`) |
| `src/configuration/ConfigurationEditor.tsx:123-152` | TLS switches: `tlsAuth`, `tlsAuthWithCACert`, `tlsSkipVerify`, `allowCleartextPasswords`, all with their descriptions verbatim |
| `src/configuration/ConfigurationEditor.tsx:156-173` | `TLS/SSL Auth Details` section (conditional on `tlsAuth \|\| tlsAuthWithCACert`) |
| `src/configuration/ConfigurationEditor.tsx:177-251` | `Additional settings` collapsible: `timezone`, `timeInterval`, `ConnectionLimits` composite |
| `pkg/mysql/mysql.go:45-72` | Settings assembly: root URL/User/Database + jsonData unmarshal + decrypted secrets |
| `pkg/mysql/mysql.go:58-61` | `jsonData.database` fallback to root `database` |
| `pkg/mysql/mysql.go:98-108` | DSN construction; `allowCleartextPasswords=true` appended when `jsonData.allowCleartextPasswords` |
| `pkg/mysql/mysql.go:130-132` | `SET time_zone='...'` when `jsonData.timezone` is set |
| `pkg/mysql/mysql.go:155-157` | Connection pool: `SetMaxOpenConns`, `SetMaxIdleConns`, `SetConnMaxLifetime` |
| `pkg/mysql/sqleng/sql_engine.go:40-61` | `sqleng.JsonData` struct (the shared jsonData shape across SQL datasources) |

### External editor components

| Component | Version | Source | What was read |
| --- | --- | --- | --- |
| `SQLOptions`, `SQLConnectionLimits` | `@grafana/sql@13.0.1` | `grafana/grafana` `v13.0.1` `packages/grafana-sql/src/types.ts:30-46` | Base jsonData shape: `maxOpenConns`, `maxIdleConns`, `maxIdleConnsAuto`, `connMaxLifetime`, `tlsAuth`, `tlsAuthWithCACert`, `timezone`, `tlsSkipVerify`, `user`, `database`, `url`, `timeInterval` |
| `useMigrateDatabaseFields` | `@grafana/sql@13.0.1` | `packages/grafana-sql/src/components/configuration/useMigrateDatabaseFields.ts:13-77` | Migrates root `database` → `jsonData.database`; sets `maxIdleConnsAuto=true`, `maxOpenConns`/`maxIdleConns`/`connMaxLifetime` defaults from `config.sqlConnectionLimits` on first render |
| `ConnectionLimits` | `@grafana/sql@13.0.1` | `packages/grafana-sql/src/components/configuration/ConnectionLimits.tsx:18-178` | Renders `Max open`, `Auto max idle`, `Max idle`, `Max lifetime` fields with their tooltips |
| `TLSSecretsConfig` | `@grafana/sql@13.0.1` | `packages/grafana-sql/src/components/configuration/TLSSecretsConfig.tsx:19-140` | `TLS/SSL Client Certificate` / `TLS/SSL Root Certificate` / `TLS/SSL Client Key` `SecretTextArea` fields with placeholders `-----BEGIN CERTIFICATE-----` and `-----BEGIN RSA PRIVATE KEY-----`, `rows={7}` |
| `DataSourceDescription`, `ConfigSection`, `ConfigSubSection`, `EditorStack` | `@grafana/plugin-ui@0.13.1` | grafana/plugin-ui | Section headings and layout |
| `SecureSocksProxySettings` | `@grafana/ui@12.3.2+` | grafana/grafana `packages/grafana-ui/…/SecureSocksProxySettings.tsx` | Excluded per AGENTS.md |

## Field provenance

| Schema `id` | Storage key | Target | Label source | Placeholder / default source | Value type |
| --- | --- | --- | --- | --- | --- |
| `root_url` | `url` | root | `ConfigurationEditor.tsx:77` (`<Field ... label="Host URL" required>`) | `:83` (`placeholder="localhost:3306"`) | `settings.URL` (SDK-native `string`) |
| `jsonData_database` | `database` | jsonData | `:88` (`<Field ... label="Database name">`) | `:93` (`placeholder="Database"`) | `sqleng.JsonData.Database string`, `sql_engine.go:56` |
| `root_user` | `user` | root | `:104` (`<Field ... label="Username" required>`) | `:108` (`placeholder="Username"`) | `settings.User string` |
| `secureJsonData_password` | `password` | secureJsonData | `:113` (`<Field ... label="Password">`) | `:116` (`placeholder="Password"`) | `settings.DecryptedSecureJSONData["password"]`, consumed at `mysql.go:100` |
| `jsonData_tlsAuth` | `tlsAuth` | jsonData | `:125` (`<Field label="Use TLS Client Auth" description="Enables TLS authentication using client cert configured in secure json data.">`) | Default `false`; switch `:128` | `SQLOptions.tlsAuth boolean`, grafana-sql `types.ts:38` |
| `jsonData_tlsAuthWithCACert` | `tlsAuthWithCACert` | jsonData | `:131` (`label="With CA Cert" description="Needed for verifying self-signed TLS Certs."`) | Default `false`; switch `:132` | `SQLOptions.tlsAuthWithCACert boolean`, `types.ts:39` |
| `jsonData_tlsSkipVerify` | `tlsSkipVerify` | jsonData | `:137` (`label="Skip TLS Verification" description="When enabled, skips verification of the MySQL server's TLS certificate chain and host name."`) | Default `false`; switch `:140` | `SQLOptions.tlsSkipVerify boolean`, `types.ts:41` |
| `jsonData_allowCleartextPasswords` | `allowCleartextPasswords` | jsonData | `:145` (`label="Allow Cleartext Passwords" description="Allows using the cleartext client side plugin if required by an account."`) | Default `false`; switch `:148-151` | `MySQLOptions.allowCleartextPasswords boolean`, `src/types.ts:4`; backend `sqleng.JsonData.AllowCleartextPasswords`, `sql_engine.go:59` |
| `secureJsonData_tlsClientCert` | `tlsClientCert` | secureJsonData | `TLSSecretsConfig.tsx:32` (`"TLS/SSL Client Certificate"`) | `:52` (`placeholder="-----BEGIN CERTIFICATE-----"`, `rows={7}`) | grafana-sql secret |
| `secureJsonData_tlsCACert` | `tlsCACert` | secureJsonData | `TLSSecretsConfig.tsx:71` (`"TLS/SSL Root Certificate"`) | `:91` (`placeholder="-----BEGIN CERTIFICATE-----"`) | grafana-sql secret |
| `secureJsonData_tlsClientKey` | `tlsClientKey` | secureJsonData | `TLSSecretsConfig.tsx:109` (`"TLS/SSL Client Key"`) | `:128` (`placeholder="-----BEGIN RSA PRIVATE KEY-----"`) | grafana-sql secret |
| `jsonData_timezone` | `timezone` | jsonData | `ConfigurationEditor.tsx:186` (`<span>Session timezone</span>` with tooltip at `:187-200`) | `:209` (`placeholder="Europe/Berlin or +02:00"`) | `SQLOptions.timezone string`, `types.ts:40`; consumed at `mysql.go:130-132` |
| `jsonData_timeInterval` | `timeInterval` | jsonData | `:218` (`<span>Min time interval</span>` + tooltip) | `:237` (`placeholder="1m"`) | `SQLOptions.timeInterval string`, `types.ts:45` |
| `jsonData_maxOpenConns` | `maxOpenConns` | jsonData | `MaxOpenConnectionsField.tsx:23` ("Max open") | Default from `config.sqlConnectionLimits.maxOpenConns` (`:45`) — a Grafana-instance-wide setting; do not populate unless overriding | `SQLConnectionLimits.maxOpenConns number`, `types.ts:31` |
| `jsonData_maxIdleConnsAuto` | `maxIdleConnsAuto` | jsonData | `ConnectionLimits.tsx:105` ("Auto max idle") | Default `true` (set by `useMigrateDatabaseFields.ts:42-45` on first render) | `SQLConnectionLimits.maxIdleConnsAuto boolean`, `types.ts:33` |
| `jsonData_maxIdleConns` | `maxIdleConns` | jsonData | `ConnectionLimits.tsx:136` ("Max idle") | Auto-tracked to `maxOpenConns` when `maxIdleConnsAuto=true` (`:79-83`); otherwise from `config.sqlConnectionLimits.maxIdleConns` (`:161`) | `SQLConnectionLimits.maxIdleConns number`, `types.ts:32` |
| `jsonData_connMaxLifetime` | `connMaxLifetime` | jsonData | `MaxLifetimeField.tsx:22` ("Max lifetime") | Default from `config.sqlConnectionLimits.connMaxLifetime` (`:42`) | `SQLConnectionLimits.connMaxLifetime number`, `types.ts:34` |

## Modeling decisions

- **Root-level `url`, `user`, `database`**: MySQL is one of the few plugins that reads root
  fields — `pkg/mysql/mysql.go:65-72` populates `dsInfo.URL`, `dsInfo.User`, `dsInfo.Database`
  from `backend.DataSourceInstanceSettings` directly, and `mysql.go:58-61` implements the
  `jsonData.database` → root `database` fallback. `Config` carries all three with
  `json:"-"` so they don't collide with jsonData during unmarshal.
- **`EffectiveDatabase()` helper**: encapsulates the `jsonData.database` → root `database`
  fallback so callers don't have to remember it.
- **Editor-marked required fields**: unlike most datasources, the MySQL editor renders
  `DataSourceDescription` with `hasRequiredFields={true}` (`ConfigurationEditor.tsx:59`) and
  marks Host URL and Username as `required`. `dsconfig.json` uses `requiredWhen: "true"`
  on both to make this always-required rather than conditional.
- **TLS switches independent**: `tlsAuth` and `tlsAuthWithCACert` are independent booleans;
  the schema does not force one implies the other. The TLS *secrets* however are
  `dependsOn` the corresponding switch.
- **Secure Socks Proxy excluded** per AGENTS.md — `ConfigurationEditor.tsx:247-249`.
- **Connection-pool defaults are runtime-only**: `useMigrateDatabaseFields` sets them from
  `config.sqlConnectionLimits.*` on first render, and the backend pulls the same instance-wide
  values at `pkg/mysql/mysql.go:45-51`. The schema does not encode a specific numeric default
  because it varies by deployment; `ApplyDefaults` only sets `maxIdleConnsAuto=true` when the
  entire pool is unconfigured.

## Upstream findings

1. **Editor allows `tlsSkipVerify=true` with `tlsAuth=false`** — the switches are independent, so a
   user can toggle "Skip TLS Verification" without any TLS actually being enabled. Harmless but
   confusing UX.
2. **User Permission Collapse is informational, not enforced** — the top-level warning at
   `ConfigurationEditor.tsx:64-71` telling users to grant SELECT only. Nothing checks it.
3. **Backend does not fail when URL/User are empty** — it builds a DSN string with empty values
   and lets MySQL reject the login. Our `Validate` fails fast on empty URL/User instead.

## Validation performed

- Go validator + JSON Schema (draft-07) — pass
- `go test -race ./...` in shared `registry/` module — pass
- `gofmt`, `go vet`, `go build` — clean
