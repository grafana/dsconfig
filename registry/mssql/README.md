# mssql

Declarative configuration schema for the [Microsoft SQL Server datasource plugin](https://github.com/grafana/grafana-mssql-datasource) (`mssql`).

## Upstream researched

- **Repo**: `github.com/grafana/grafana-mssql-datasource`
- **Ref**: `main`
- **Commit SHA**: `c4133924fde76ab06fdf25688d9ccc076ffae4b7` (2026-07-02)

Every value in [`dsconfig.json`](dsconfig.json) is traceable to a specific `file:line` in the
upstream repo at this SHA.

```bash
git clone https://github.com/grafana/grafana-mssql-datasource
cd grafana-mssql-datasource
git checkout c4133924fde76ab06fdf25688d9ccc076ffae4b7
```

## Files

| File | Purpose |
| --- | --- |
| [`dsconfig.json`](dsconfig.json) | dsconfig v1 schema — single source of truth |
| [`settings.ts`](settings.ts) | TypeScript models: `RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig`, `AzureCredentials` union |
| [`settings.go`](settings.go) | Go `Config`, `AuthType` (7 values) / `EncryptOption` / `SecureJsonDataKey` typed constants, `LoadConfig` / `ApplyDefaults` / `Validate`, `EffectiveDatabase()` |
| [`schema.go`](schema.go) | k8s-style SDK plugin schema + 10 `SettingsExamples` covering every auth type |
| [`settings_test.go`](settings_test.go) | Table tests for `LoadConfig`, `ApplyDefaults`, `Validate`, `EffectiveDatabase` |
| [`conformance_test.go`](conformance_test.go) | `schema.RunPluginTests` wrapper |
| `.gen.json` artifacts | Regenerate with `go generate ./...` |

## Sources researched

Read at the pinned upstream SHA:

| File | What was read |
| --- | --- |
| `src/plugin.json:3-5` | `pluginType` (`id: mssql`), `pluginName` (`Microsoft SQL Server`) |
| `src/types.ts:16-24` | `MSSQLAuthenticationType` enum — 7 string values |
| `src/types.ts:26-30` | `MSSQLEncryptOptions` enum — `disable` / `false` / `true` |
| `src/types.ts:31-44` | `MssqlOptions extends SQLOptions` (adds `authenticationType`, `encrypt`, `sslRootCertFile`, `serverName`, `connectionTimeout`, `azureCredentials`, `keytabFilePath`, `credentialCache`, `credentialCacheLookupFile`, `configFilePath`, `UDPConnectionLimit`, `enableDNSLookupKDC`) |
| `src/configuration/ConfigurationEditor.tsx:45` | `useMigrateDatabaseFields(props)` |
| `src/configuration/ConfigurationEditor.tsx:100-118` | `buildAuthenticationOptions` — the 7 auth-type radio options |
| `src/configuration/ConfigurationEditor.tsx:120-127` | `encryptOptions` array |
| `src/configuration/ConfigurationEditor.tsx:131-135` | `DataSourceDescription` (`hasRequiredFields`) |
| `src/configuration/ConfigurationEditor.tsx:154-183` | `Host` (required, root URL) + `Database` (required, jsonData.database) |
| `src/configuration/ConfigurationEditor.tsx:187-236` | `Encrypt` field with default `false` and rich per-option description |
| `src/configuration/ConfigurationEditor.tsx:238-285` | Encrypt-conditional: `Skip TLS Verify`, `TLS/SSL Root Certificate`, `Hostname in server certificate` |
| `src/configuration/ConfigurationEditor.tsx:288-346` | `Authentication Type` combobox with the multi-bullet description |
| `src/configuration/ConfigurationEditor.tsx:353-392` | Username/Password fields (rendered for SQL Auth and Kerberos-raw) |
| `src/configuration/ConfigurationEditor.tsx:394-400` | Azure Authentication Settings (delegated to `AzureAuthSettings` component) |
| `src/configuration/ConfigurationEditor.tsx:413` | `ConnectionLimits` (from `@grafana/sql`) |
| `src/configuration/ConfigurationEditor.tsx:415-462` | `Connection details` sub-section: `Min time interval`, `Connection timeout` |
| `src/configuration/ConfigurationEditor.tsx:463-466` | Excluded: `SecureSocksProxySettings`, `KerberosAdvancedSettings` |
| `src/configuration/Kerberos.tsx:42-73` | Keytab sub-panel: `Username` + `Keytab file path` (required) |
| `src/configuration/Kerberos.tsx:76-97` | Credential cache sub-panel: `Credential cache path` (required) |
| `src/configuration/Kerberos.tsx:99-136` | Credential cache file sub-panel: `Username` + `Credential cache file path` (required) |
| `src/configuration/Kerberos.tsx:141-243` | `KerberosAdvancedSettings`: `UDP Preference Limit`, `DNS Lookup KDC`, `krb5 config file path` |
| `pkg/mssql/mssql.go:22-64` | Backend settings assembly (root URL/User/Database + jsonData unmarshal + Azure settings) |
| `pkg/mssql/mssql.go:29-36` | Backend defaults (`Encrypt: "false"`, `ConnectionTimeout: 0`, `SecureDSProxy: false`) |
| `pkg/mssql/mssql.go:49-52` | `jsonData.database` fallback to root `settings.Database` |
| `pkg/mssql/sqleng/sql_engine.go:48-69` | Shared `sqleng.JsonData` shape |
| `pkg/mssql/kerberos/kerberos.go:22-59` | `KerberosAuth` struct + `GetKerberosSettings` unmarshaling |
| `pkg/mssql/kerberos/kerberos.go:37` | Backend default `UDPConnectionLimit: 1` |
| `pkg/mssql/azure/connection.go:10-40` | `GetAzureCredentialDSNFragment` — how Azure credentials flow into the DSN |

### External editor components

| Component | Version | Source | What was read |
| --- | --- | --- | --- |
| `SQLOptions`, `ConnectionLimits`, `NumberInput`, `useMigrateDatabaseFields` | `@grafana/sql@13.0.2` | grafana/grafana `v13.0.2` `packages/grafana-sql/` | Shared jsonData shape + connection-pool sub-section |
| `AzureCredentials`, `AzureCredentialsConfig` | `@grafana/azure-sdk@0.1.0` | `github.com/grafana/grafana-azure-sdk-react` `src/credentials/AzureCredentials.ts` | The 7-variant `AzureCredentials` discriminated union (`msi`, `workloadidentity`, `clientsecret`, `clientsecret-obo`, `ad-password`, `clientcertificate`, `currentuser`) and the fact that `clientSecret` lives in `secureJsonData.azureClientSecret` (`AzureCredentialsConfig.ts:428`) |
| `DataSourceDescription`, `ConfigSection`, `ConfigSubSection` | `@grafana/plugin-ui@0.13.1+` | grafana/plugin-ui | Section layout |
| `Select`, `Input`, `Switch`, `SecretInput`, `SecureSocksProxySettings` | `@grafana/ui@13.0.2` | grafana/grafana | UI primitives |

## Field provenance (abbreviated)

| Schema `id` | Storage key | Target | Label source | Placeholder / default | Value type |
| --- | --- | --- | --- | --- | --- |
| `root_url` | `url` | root | `ConfigurationEditor.tsx:155` (`title-host` "Host") | `:166` (`placeholder="localhost:1433"`) | `settings.URL` |
| `jsonData_database` | `database` | jsonData | `:171` (`title-database` "Database") | `:180` (`placeholder-database` "database name") | `sqleng.JsonData.Database string`, `sql_engine.go:64` |
| `jsonData_encrypt` | `encrypt` | jsonData | `:227` ("Encrypt") | Default `false` (`:231`); options `:120-127` | `MSSQLEncryptOptions` union, `src/types.ts:26-30`; backend `sqleng.JsonData.Encrypt string`, `sql_engine.go:61` |
| `jsonData_tlsSkipVerify` | `tlsSkipVerify` | jsonData | `:242` ("Skip TLS Verify") | Default `false` | `SQLOptions.tlsSkipVerify boolean`; backend `sqleng.JsonData.TlsSkipVerify`, `sql_engine.go:56` |
| `jsonData_sslRootCertFile` | `sslRootCertFile` | jsonData | `:257` ("TLS/SSL Root Certificate") | `:263` (`placeholder-tls-cert`) | `MssqlOptions.sslRootCertFile string`, `src/types.ts:34`; backend `sqleng.JsonData.RootCertFile`, `sql_engine.go:57` |
| `jsonData_serverName` | `serverName` | jsonData | `:270` ("Hostname in server certificate") | `:274` (`placeholder-common-name`) | `MssqlOptions.serverName string`, `src/types.ts:35`; backend `sqleng.JsonData.Servername`, `sql_engine.go:62` |
| `jsonData_authenticationType` | `authenticationType` | jsonData | `:290` ("Authentication Type") | Default `SQL Server Authentication` (`:340`); options `:100-118` | `MSSQLAuthenticationType` union, `src/types.ts:16-24` |
| `root_user` | `user` | root | `:358` ("Username"; also Kerberos.tsx `:45,104`) | `:370` (`placeholder-user`) or `:369` (`name@EXAMPLE.COM` for kerberos-raw) | `settings.User string` |
| `secureJsonData_password` | `password` | secureJsonData | `:377` ("Password") | `:384` (`placeholder-password`) | `settings.DecryptedSecureJSONData["password"]` |
| `jsonData_keytabFilePath` | `keytabFilePath` | jsonData | `Kerberos.tsx:59` ("Keytab file path") | `:66` (`placeholder="/home/grot/grot.keytab"`) | `MssqlOptions.keytabFilePath string`, `src/types.ts:38`; backend `KerberosAuth.KeytabFilePath`, `kerberos.go:23` |
| `jsonData_credentialCache` | `credentialCache` | jsonData | `Kerberos.tsx:79` ("Credential cache path") | `:89` (`placeholder="/tmp/krb5cc_1000"`) | `MssqlOptions.credentialCache string`, `src/types.ts:39`; backend `KerberosAuth.CredentialCache`, `kerberos.go:24` |
| `jsonData_credentialCacheLookupFile` | `credentialCacheLookupFile` | jsonData | `Kerberos.tsx:118` ("Credential cache file path") | `:128` (`placeholder="/home/grot/cache.json"`) | `MssqlOptions.credentialCacheLookupFile string`, `src/types.ts:40`; backend `KerberosAuth.CredentialCacheLookupFile`, `kerberos.go:25` |
| `jsonData_UDPConnectionLimit` | `UDPConnectionLimit` | jsonData | `Kerberos.tsx:166` ("UDP Preference Limit") | Default `1` (`kerberos.go:37`) | `MssqlOptions.UDPConnectionLimit number`, `src/types.ts:42`; backend `KerberosAuth.UDPConnectionLimit int`, `kerberos.go:27` |
| `jsonData_enableDNSLookupKDC` | `enableDNSLookupKDC` | jsonData | `Kerberos.tsx:194` ("DNS Lookup KDC") | Default `'true'` (per description; string, not bool) | `MssqlOptions.enableDNSLookupKDC string`, `src/types.ts:43` |
| `jsonData_configFilePath` | `configFilePath` | jsonData | `Kerberos.tsx:217` ("krb5 config file path") | Default `/etc/krb5.conf` (`:237`) | `MssqlOptions.configFilePath string`, `src/types.ts:41` |
| `jsonData_azureCredentials` | `azureCredentials` | jsonData | — (delegated to `AzureAuthSettings`) | Object; see `@grafana/azure-sdk` `AzureCredentials` union | `MssqlOptions.azureCredentials AzureCredentials`, `src/types.ts:37` |
| `secureJsonData_azureClientSecret` | `azureClientSecret` | secureJsonData | — (@grafana/azure-sdk-managed) | Written by `AzureCredentialsConfig.ts:428` | `@grafana/azure-sdk` secret |
| `jsonData_connectionTimeout` | `connectionTimeout` | jsonData | `ConfigurationEditor.tsx:453` ("Connection timeout") | Default `0` (`:449`, `mssql.go:34`) | `MssqlOptions.connectionTimeout number`, `src/types.ts:36`; backend `sqleng.JsonData.ConnectionTimeout`, `sql_engine.go:52` |
| `jsonData_timeInterval` | `timeInterval` | jsonData | `:431` ("Min time interval") | `:436` (`placeholder="1m"`) | `SQLOptions.timeInterval string` |
| `jsonData_maxOpenConns` / `maxIdleConns` / `maxIdleConnsAuto` / `connMaxLifetime` | — | jsonData | `ConnectionLimits` component (`@grafana/sql`) | Defaults from `config.sqlConnectionLimits.*` | `SQLConnectionLimits`, `@grafana/sql` |

## Modeling decisions

- **Seven authentication types** — the most complex of the SQL family. Modeled as an
  `authenticationType` discriminator with fine-grained `dependsOn` and `requiredWhen` per branch
  (SQL Auth needs user+password; Kerberos-raw same; Kerberos keytab needs user+keytab path;
  Kerberos credential-cache needs the cache path; Kerberos cache-lookup-file needs user+lookup
  file; Windows and Azure need neither user nor password).
- **Auth-type label / value quirk**: the auth-type radio button labels differ slightly from the
  values themselves — e.g. "Windows AD: Keytab file" (label) vs `"Windows AD: Keytab"` (value).
  Schema records both faithfully.
- **`azureCredentials` as opaque `valueType: any`**: the Azure credential shape is a
  seven-variant discriminated union defined in `@grafana/azure-sdk`; modeling every branch in
  dsconfig would double the field count. Recorded as an opaque object with a description that
  lists the possible `authType` values and points at the SDK. Secret component
  (`azureClientSecret`) is modeled as a normal secure field.
- **Encrypt options include `'false'` as the default**, not `'disable'`. The MSSQL protocol
  distinguishes "no encryption at all" (`disable`) from "encrypt only the login packet"
  (`false`, the historic default). The schema respects that.
- **TLS-verify fields gated on `encrypt === 'true' && tlsSkipVerify !== true`**: three-way
  guard mirroring the editor's nested conditional (`ConfigurationEditor.tsx:238-283`).
- **Kerberos advanced settings apply to ALL four Kerberos variants**: `dependsOn` uses the OR
  chain of the four Kerberos auth-type values.
- **Root-level URL/User/Database on Config** because the backend reads them directly
  (`pkg/mssql/mssql.go:56-58`); same `EffectiveDatabase()` helper as MySQL/Postgres.
- **Secure Socks Proxy excluded** per AGENTS.md — `ConfigurationEditor.tsx:463-465`.

## Upstream findings

1. **Auth-type label vs value mismatch**: `Windows AD: Keytab file` renders in the radio but the
   stored value is `Windows AD: Keytab` — the plural label is a UI-only cosmetic difference. The
   schema uses the stored value (`Windows AD: Keytab`) as the canonical allowed value.
2. **`UDPConnectionLimit` can be legacy-stored as a string.** `pkg/mssql/kerberos/kerberos.go:50-56`
   handles the case where a legacy provisioning payload stores `UDPConnectionLimit` as a JSON
   string — falling back to `strconv.Atoi`. New configurations should always write it as a
   number.
3. **`enableDNSLookupKDC` is a string, not a bool.** Frontend type is `string` (`src/types.ts:43`),
   default `'true'`. This looks like a bug (should be a boolean), but it's the shipped surface.
4. **Kerberos-raw + Windows AD variants read `settings.user` even when the auth type doesn't
   render the field.** Provisioning payloads for `Windows AD: Credential cache` don't render a
   username field in the editor, but the backend may still consume it if set.
5. **Windows Authentication requires no credentials in the datasource.** SSO uses the Grafana
   process's Windows identity — running Grafana as a service account that has SQL access is a
   deployment concern, not a datasource-config concern.
6. **`configFilePath` default `/etc/krb5.conf` is only enforced by the editor.** The backend
   `KerberosAuth` struct defaults it to `""` (`kerberos.go:36`); an empty value falls back to
   the MIT krb5 library's own default, which is platform-dependent.

## Validation performed

- Go validator + JSON Schema (draft-07) — pass
- `go test -race ./...` — pass (14 LoadConfig + ApplyDefaults + 10 Validate + 2 EffectiveDatabase)
- `gofmt`, `go vet`, `go build` — clean
