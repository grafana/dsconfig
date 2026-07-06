# grafana-aurora-datasource

Declarative configuration schema for the [Amazon Aurora datasource plugin](https://github.com/grafana/grafana-aurora-datasource) (`grafana-aurora-datasource`).

## Upstream researched

- **Repo**: `github.com/grafana/grafana-aurora-datasource`
- **Ref**: `main`
- **Commit SHA**: `c7452e8b63724389a973d35f1d28dea779b0c72a`

Every value in [`dsconfig.json`](dsconfig.json) is traceable to a specific
`file:line` in the upstream repo at this SHA.

```bash
git clone https://github.com/grafana/grafana-aurora-datasource
cd grafana-aurora-datasource
git checkout c7452e8b63724389a973d35f1d28dea779b0c72a
```

## Files

| File | Purpose |
| --- | --- |
| [`dsconfig.json`](dsconfig.json) | dsconfig v1 schema — single source of truth |
| [`settings.ts`](settings.ts) | TypeScript models: `RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig` |
| [`settings.go`](settings.go) | Go `Config` (flat: AWS SDK jsonData + Aurora-specific jsonData + `DecryptedSecureJSONData`), `LoadConfig` / `ApplyDefaults` / `Validate`, `AWSAuthType` / `AuroraEngine` enums, `SecureJsonDataKey` constants |
| [`schema.go`](schema.go) | k8s-style SDK plugin schema + 10 `SettingsExamples` |
| [`settings_test.go`](settings_test.go) | Table tests for `LoadConfig`, `ApplyDefaults`, `Validate` |
| [`conformance_test.go`](conformance_test.go) | `schema.RunPluginTests` wrapper |
| `.gen.json` artifacts | Regenerate with `go generate ./...` |

## Sources researched

Read at the pinned upstream SHA:

| File | What was read |
| --- | --- |
| `src/plugin.json:2-9` | `pluginType` (`id: grafana-aurora-datasource`), `pluginName` (`Amazon Aurora`), backend, alerting, executable, category |
| `src/plugin.json:21-26` | `docURL` → `https://grafana.com/docs/plugins/grafana-aurora-datasource` |
| `src/types.ts:32-40` | `AuroraConfigOptions extends AwsAuthDataSourceJsonData` — Aurora-specific jsonData shape (`engine`, `dbName`, `dbUser`, `dbHost`, `dbPort`, `dbHostAuth`, `dbPortAuth`) |
| `src/types.ts:42` | `AuroraSecureConfigOptions extends AwsAuthDataSourceSecureJsonData` — no plugin-specific secrets |
| `src/types.ts:44-47` | `SupportedEngines` enum (`aurora-mysql`, `aurora-postgres`) |
| `src/components/ConfigEditor.tsx:31-34` | `<ConnectionConfig>` — full AWS auth section; `showHttpProxySettings` not passed, so proxy fields are excluded |
| `src/components/ConfigEditor.tsx:39-52` | `Database Settings` subsection start + Engine `Select` (`Aurora (PostgreSQL Compatible)`/`aurora-postgres`, `Aurora (MySQL Compatible)`/`aurora-mysql`) with `aurora-postgres` default |
| `src/components/ConfigEditor.tsx:53-60` | `Database Name` field — `placeholder="Database"`, no description, optional |
| `src/components/ConfigEditor.tsx:61-75` | `Database User` field — required, `placeholder="postgres"` |
| `src/components/ConfigEditor.tsx:76-100` | `Database Host` field — required, description with link to AWS Aurora reader-endpoints docs, `placeholder="Host"` |
| `src/components/ConfigEditor.tsx:101-116` | `Database Port` field — required, `type="number"`, dynamic placeholder (`3306` for mysql, `5432` otherwise) |
| `src/components/ConfigEditor.tsx:118-129` | Advanced subsection title `Advanced: Separate Host and Port for Auth` + description with link to `rds/generate-db-auth-token` docs |
| `src/components/ConfigEditor.tsx:130-143` | `Advanced: DB Host For Auth` field — optional, description `Optional, if not provided, the dbHost above will be used.`, `placeholder="separate host for generating auth token"` |
| `src/components/ConfigEditor.tsx:144-156` | `Advanced: DB Port For Auth` field — optional, description `Optional, if not provided, the dbPort above will be used.`, `placeholder="separate port for generating auth token"` |
| `src/components/ConfigEditor.tsx:35-37` | Excluded: `SecureSocksProxySettings` |
| `pkg/plugin/consts.go:5-17` | `SupportedEngine` type + `AuroraMysql`/`AuroraPostgres` constants + `SupportedEngines` string map |
| `pkg/plugin/driver.go:90-101` | `AuroraConfigSettings` — embeds `awsds.AWSDatasourceSettings`, adds `Engine`, `DBUser`, `DBName`, `DBHost`, `DBPort` (int), `DefaultRegion`, `DBHostAuth`, `DBPortAuth` (int) |
| `pkg/plugin/driver.go:103-115` | `parseSettings` — unmarshal jsonData then copy `accessKey`/`secretKey`/`sessionToken` from `settings.DecryptedSecureJSONData` |
| `pkg/plugin/connect.go:23-72` | `ConnectToAurora` — resolves AWS credentials via `driverDependencies.GetAWSConfig`, then builds an RDS IAM auth token with `BuildAuthToken(ctx, "${host}:${port}", region, dbUser, credentials)`. Host/port fall back to `DBHost`/`DBPort` unless `DBHostAuth`/`DBPortAuth` is set (`:56-67`). |
| `pkg/plugin/connect.go:81-138` | Engine dispatch — postgres/mysql DSN construction; empty/unknown engine falls back to postgres (`:83-85, 135-138`) |
| `pkg/plugin/datasource.go:11-19` | Plugin bootstrap — no per-instance settings caching beyond what sqlds provides |

### External editor components

| Component | Version | Source | What was read |
| --- | --- | --- | --- |
| `ConnectionConfig`, `Divider` | `@grafana/aws-sdk@0.10.2` | `github.com/grafana/grafana-aws-sdk-react` `v0.10.2` `src/components/ConnectionConfig.tsx` | AWS auth Select + all conditional fields (label, placeholder, description, `dependsOn` shapes) — Aurora does NOT pass `showHttpProxySettings`, so proxy fields are excluded |
| `AwsAuthType`, `AwsAuthDataSourceJsonData`, `AwsAuthDataSourceSecureJsonData` | `@grafana/aws-sdk@0.10.2` | `src/types.ts` | Base jsonData / secureJsonData types Aurora extends |
| `ConfigSubSection` | `@grafana/plugin-ui@0.13.0` | `github.com/grafana/plugin-ui` | Subsection heading component (Database Settings / Advanced) |
| `SecureSocksProxySettings` | `@grafana/ui@12.0.1+` | grafana/grafana `packages/grafana-ui` | Secure Socks proxy panel — excluded from this schema per AGENTS.md |

### Backend AWS SDK

| Component | Version | Source | What was read |
| --- | --- | --- | --- |
| `awsds.AWSDatasourceSettings` | `grafana-aws-sdk v1.4.6` | `pkg/awsds/settings.go:94-117` | Embedded backend AWS settings struct (all fields tagged camelCase to match the frontend writes) |
| `awsds.AuthType` (UnmarshalJSON) | `grafana-aws-sdk v1.4.6` | `pkg/awsds/settings.go:13-91` | Custom JSON unmarshal maps legacy `sharedCreds` → `credentials` and `arn` → `default` |

## Field provenance

| Schema `id` | Storage key | Target | Editor label source | Value type |
| --- | --- | --- | --- | --- |
| `jsonData_authType` | `authType` | jsonData | `@grafana/aws-sdk` `ConnectionConfig.tsx` "Authentication Provider" | `AwsAuthType` (`awsds.AuthType` on backend) |
| `jsonData_profile` | `profile` | jsonData | "Credentials Profile Name" | `string` |
| `secureJsonData_accessKey` | `accessKey` | secureJsonData | "Access Key ID" | `string` (secret) |
| `secureJsonData_secretKey` | `secretKey` | secureJsonData | "Secret Access Key" | `string` (secret) |
| `secureJsonData_sessionToken` | `sessionToken` | secureJsonData | (no editor label — backend-only) | `string` (secret), backend-only |
| `jsonData_assumeRoleArn` | `assumeRoleArn` | jsonData | "Assume Role ARN" | `string` |
| `jsonData_externalId` | `externalId` | jsonData | "External ID" | `string` |
| `jsonData_endpoint` | `endpoint` | jsonData | "Endpoint" | `string` |
| `jsonData_defaultRegion` | `defaultRegion` | jsonData | "Default Region" | `string` |
| `jsonData_engine` | `engine` | jsonData | `ConfigEditor.tsx:40` "Engine" | `AuroraEngine` (`aurora-postgres` / `aurora-mysql`) |
| `jsonData_dbName` | `dbName` | jsonData | `ConfigEditor.tsx:53` "Database Name" (`placeholder="Database"`) | `string` |
| `jsonData_dbUser` | `dbUser` | jsonData | `ConfigEditor.tsx:62` "Database User" (required, `placeholder="postgres"`) | `string` |
| `jsonData_dbHost` | `dbHost` | jsonData | `ConfigEditor.tsx:77` "Database Host" (required, reader-endpoint link) | `string` |
| `jsonData_dbPort` | `dbPort` | jsonData | `ConfigEditor.tsx:102` "Database Port" (required, engine-dependent placeholder) | `number` (Go `int`) |
| `jsonData_dbHostAuth` | `dbHostAuth` | jsonData | `ConfigEditor.tsx:132` "Advanced: DB Host For Auth" | `string` |
| `jsonData_dbPortAuth` | `dbPortAuth` | jsonData | `ConfigEditor.tsx:145` "Advanced: DB Port For Auth" | `number` (Go `int`) |

## Where the types are defined

**Frontend (plugin-owned)**

- `src/types.ts:32-40` — `AuroraConfigOptions` (`extends AwsAuthDataSourceJsonData`)
- `src/types.ts:42` — `AuroraSecureConfigOptions` (`extends AwsAuthDataSourceSecureJsonData`)
- `src/types.ts:44-47` — `SupportedEngines` enum

**Frontend (external, `@grafana/aws-sdk@0.10.2`)**

- `src/types.ts` — `AwsAuthType`, `AwsAuthDataSourceJsonData`, `AwsAuthDataSourceSecureJsonData`

**Backend (plugin-owned)**

- `pkg/plugin/driver.go:90-101` — `AuroraConfigSettings` (`embeds awsds.AWSDatasourceSettings + backend.DataSourceInstanceSettings`)
- `pkg/plugin/consts.go:5-17` — `SupportedEngine` type + `AuroraMysql`/`AuroraPostgres` constants + `SupportedEngines` string map

**Backend (external, `grafana-aws-sdk v1.4.6`)**

- `pkg/awsds/settings.go:13-91` — `AuthType` int-enum + custom UnmarshalJSON
- `pkg/awsds/settings.go:94-117` — `AWSDatasourceSettings` struct

## Frontend-only vs backend-only settings

| Setting | Read by editor | Written by editor | Read by backend | Notes |
| --- | --- | --- | --- | --- |
| `sessionToken` (secureJsonData) | ✗ | ✗ | ✓ (`pkg/plugin/driver.go:112`) | **Backend-only**. No editor UI; provisioning-only for temporary STS creds |
| `dbHostAuth`, `dbPortAuth` (jsonData) | ✓ | ✓ | ✓ (`pkg/plugin/connect.go:59-66`) | Not backend-only — editor exposes them under "Advanced" |
| Every other jsonData/secureJsonData field | ✓ | ✓ | ✓ | Editor-visible and backend-consumed |

## Modeling decisions

- **`engine` is a required-in-schema string enum** even though the backend
  silently coerces empty/unknown values to `aurora-postgres` at connect time
  (`pkg/plugin/connect.go:83-85, 135-138`). `Validate` rejects unknown
  engines so a mistyped provisioning value produces a real error instead of
  silently landing on the wrong dialect. Legacy configs with no engine at
  all round-trip through `ApplyDefaults` which sets `aurora-postgres`.
- **Port placeholder is fixed at `5432` in the schema** because dsconfig
  fields carry a single placeholder string. The upstream editor dynamically
  swaps the placeholder to `3306` when `engine === 'aurora-mysql'`; that
  behaviour is documented in `instructions` and the `jsonData_engine` /
  `jsonData_dbPort` `relationships.pair` entry rather than modelled as a
  virtual field.
- **`dbPort` and `dbPortAuth` are `valueType: "number"`** in the schema and
  Go `int` on the backend. The frontend type declares them as `number |
  null` because a cleared input writes `null`; the backend `omitempty`
  tag makes zero-valued ports round-trip as absent, matching the editor's
  behaviour.
- **No password / no basic-auth**: Aurora authenticates to the DB with an
  RDS IAM auth token generated at connect time from the resolved AWS
  credentials (`pkg/plugin/connect.go:34-72`). The only secrets are the
  AWS credentials themselves.
- **`endpoint.domain` + `endpoint.port` roles** are applied to `dbHost` /
  `dbPort` because they collectively define the Aurora cluster endpoint.
  `endpoint.baseUrl` is intentionally NOT applied to `dbHost` alone since
  it is a host-only string, not a full URL.
- **Secure Socks Proxy excluded** per AGENTS.md
  (`ConfigEditor.tsx:35-37`).
- **AWS proxy fields excluded**: Aurora's `<ConnectionConfig>` invocation
  omits `showHttpProxySettings`, so the schema does not model the
  `proxyType` / `proxyUrl` / `proxyUsername` / `proxyPassword` fields the
  AWS pack would otherwise include. The pack itself is NOT included via
  `baseFields` — the fields are declared inline to keep the excluded-set
  visible in one place, matching how Athena and Redshift model their
  entries.

## Settings examples matrix

| Example key | AWS auth | Engine | Notes |
| --- | --- | --- | --- |
| `""` (default) | `default` | `aurora-postgres` | Post-defaults empty config; fails `Validate` (missing selectors) |
| `awsSdkDefaultPostgres` | `default` | `aurora-postgres` | Full config, port 5432 |
| `accessAndSecretKeyMysql` | `keys` | `aurora-mysql` | Full config, port 3306 + accessKey/secretKey secrets |
| `credentialsFile` | `credentials` | `aurora-postgres` | `profile: "my-aurora-profile"` |
| `workspaceIamRole` | `ec2_iam_role` | `aurora-postgres` | Instance/task/pod role |
| `grafanaAssumeRole` | `grafana_assume_role` | `aurora-postgres` | Grafana Cloud broker |
| `assumeRoleFromKeys` | `keys` + `assumeRoleArn` | `aurora-postgres` | Cross-account STS |
| `splitAuthEndpoint` | `ec2_iam_role` | `aurora-postgres` | dbHost=LB, dbHostAuth=primary endpoint |
| `legacyArnAuthType` | `arn` (legacy) | `aurora-postgres` | Backend maps to `default` via `awsds.AuthType.UnmarshalJSON` |
| `legacyMissingEngine` | `default` | *(unset)* | Backend falls back to `aurora-postgres` |

## Upstream findings

1. **Engine fallback is silent.** `pkg/plugin/connect.go:135-138` treats any
   unknown `settings.Engine` value as `aurora-postgres` after logging a
   debug message ("Unknown or unsupported engine, falling back to aurora
   postgres engine. Go back to datasource config page, reselect an engine,
   and save"). A datasource provisioned with `engine: "aurora-mysql"` typoed
   to `"aurora-msyql"` silently connects to Postgres. Our `Validate`
   rejects unknown engines so provisioning callers see the error instead.
2. **`dbUser` and `dbHost` are marked `required` in the editor but not
   enforced client-side beyond an `invalid` marker.** The `Field`s render
   an error banner if the input is empty AND dirty, but the "Save & test"
   button is not disabled, so an operator can still save an incomplete
   config; the failure surfaces on the backend's first connection attempt.
   Our `Validate` returns the missing-field errors up front.
3. **Aurora MySQL requires an external cert fetch at connect time.**
   `pkg/plugin/connect.go:151-172` unconditionally downloads
   `https://s3.amazonaws.com/rds-downloads/rds-combined-ca-bundle.pem` and
   registers a TLS config named `rds` with `InsecureSkipVerify: true` on
   the MySQL driver. That fetch happens on every connect call — a network
   failure to that S3 endpoint makes new MySQL connections fail even when
   the RDS cluster itself is reachable. The Postgres path skips it.
4. **RDS certs are registered with `InsecureSkipVerify: true`.** In the
   MySQL path (`connect.go:167`), the TLS config trusts the fetched CA
   bundle but sets `InsecureSkipVerify: true`, which bypasses hostname
   verification. This is a documented AWS-SDK-Go-v2 workaround
   ([`aws/aws-sdk-go-v2#2698`](https://github.com/aws/aws-sdk-go-v2/issues/2698))
   but worth calling out to operators.
5. **`plugin.json` declares `grafanaDependency: ">=9.4.0"`** but the
   Postgres proxy path (`connect.go:114-134`) uses features from
   `@grafana/plugin-sdk-go` that only became stable in later Grafana
   releases. In practice the plugin ships against Grafana 10+.
6. **No `Load`/`Validate` on the backend Settings struct.** Unlike the
   PostgreSQL and Redshift datasources, `AuroraConfigSettings` has no
   dedicated Load method — jsonData is unmarshalled directly and secrets
   are copied inline in `parseSettings` (`driver.go:103-115`). There is
   no upstream defaulting or validation to sync from.

## Validation performed

- Go validator + JSON Schema (draft-07) via `TestSchemaConformance`
  (`conformance_test.go`)
- Full `LoadConfig` / `ApplyDefaults` / `Validate` unit tests
  (`settings_test.go`)
- `gofmt`, `go vet`, `go test ./...` inside `registry/` — all clean
