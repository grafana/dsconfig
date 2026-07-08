# grafana-athena-datasource

Declarative configuration schema for the [Amazon Athena datasource plugin](https://github.com/grafana/athena-datasource) (`grafana-athena-datasource`).

## Upstream researched

- **Repo**: `github.com/grafana/athena-datasource`
- **Ref**: `main`
- **Commit SHA**: `a708c50e54207b08a8f10045fb327288519660d0` (2026-07-02, `docs: add signed commits requirement to CONTRIBUTING.md`)

Every value in [`dsconfig.json`](dsconfig.json) — labels, placeholders, tooltips, option
labels/values, section titles, defaults, validations, dependency and required-when
expressions, storage keys, storage targets, value types, group titles, and instructions —
is traceable to a specific `file:line` in the upstream plugin repo (or in the pinned
`@grafana/aws-sdk` version of the shared `ConnectionConfig` component) at this SHA. See
[Field provenance](#field-provenance) below.

To reproduce this research:

```bash
git clone --depth 20 https://github.com/grafana/athena-datasource
cd athena-datasource
git checkout a708c50e54207b08a8f10045fb327288519660d0

# The AWS auth surface is rendered by @grafana/aws-sdk's ConnectionConfig. Pin
# to the version athena-datasource's package.json uses (currently 0.10.2):
git clone https://github.com/grafana/grafana-aws-sdk-react
cd grafana-aws-sdk-react
git checkout v0.10.2   # SHA fe0c4d8d657ee5ed053ae173293dc876619b5a2b
```

If upstream `main` has advanced past the pinned SHA, re-diff the sources listed under
[Sources researched](#sources-researched) and reconcile the schema before merging.

## Files

| File | Purpose |
| --- | --- |
| [`dsconfig.json`](dsconfig.json) | dsconfig v1 schema — single source of truth for all config fields, groups, and instructions |
| [`settings.ts`](settings.ts) | TypeScript models: `RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig` |
| [`settings.go`](settings.go) | Go `Config` model (flat: jsonData fields + `DecryptedSecureJSONData`), `PluginID`, `AWSAuthType` typed constants, `SecureJsonDataKey` typed constants, and the `LoadConfig` utility |
| [`schema.go`](schema.go) | k8s-style SDK plugin schema: embeds `dsconfig.json`, converts it via `dsconfig.NewSDKSchema`, and defines `SettingsExamples` for each AWS auth provider |
| [`settings_test.go`](settings_test.go) | Table tests for `LoadConfig`, `ApplyDefaults`, and `Validate` |
| [`conformance_test.go`](conformance_test.go) | Runs `schema.RunPluginTests` — the shared dsconfig conformance suite — against `dsconfig.json` |
| [`schema.gen.json`](schema.gen.json), [`settings.gen.json`](settings.gen.json), [`settings.examples.gen.json`](settings.examples.gen.json) | Committed schema artifacts (regenerate with `go generate ./...` inside this directory; the `SchemaArtifactInSync` conformance subtest guards drift) |

There is no per-entry `go.mod` — every registry entry is a subpackage of the shared
[`registry/`](..) module (`github.com/grafana/dsconfig/registry`).

## Sources researched

Every source below was read at the pinned upstream SHA
(`a708c50e54207b08a8f10045fb327288519660d0`) or, for `@grafana/aws-sdk`, at the exact
version pinned in the plugin's `package.json` (`0.10.2` / SHA `fe0c4d8`).

### Plugin repo (`github.com/grafana/athena-datasource@a708c50`)

| File | What was read |
| --- | --- |
| `src/plugin.json:2-52` | `pluginType` (`id`), `pluginName` (`name`), `docURL` (first `info.links[]` URL is the GitHub repo; the docs URL used matches Grafana's plugin catalog URL used by the plugin) |
| `src/ConfigEditor.tsx:1-11,105-181` | Editor layout, `ConnectionConfig` composition, `<ConfigSection title="Athena Details">`, `<Field>` labels via `selectors`, `<ConfigSelect>` for catalog/database/workgroup, `<Input>` for outputLocation with `placeholder="s3://"` |
| `src/ConfigEditor.tsx:112` | `<ConnectionConfig>` is called without `showHttpProxySettings`, so the AWS proxy fields are NOT editor-visible for Athena (excluded from this entry) |
| `src/ConfigEditor.tsx:113-115` | Conditional `SecureSocksProxySettings` render — writes `jsonData.enableSecureSocksProxy`; deliberately excluded per AGENTS.md |
| `src/ConfigEditor.tsx:70-74` | `useEffect` fetches externalId when `authType === GrafanaAssumeRole`; feature-gated behavior we surface as an instruction, not a field |
| `src/ConfigEditor.tsx:81-101` | `onChange('catalog'/'database'/'workgroup')` and `onChangeOutputLocation` — the frontend writes camelCase keys (`catalog`, `database`, `workgroup`, `outputLocation`) |
| `src/tests/selectors.ts:17-32` | `Field` labels: `catalog.input="Data source"`, `database.input="Database"`, `workgroup.input="Workgroup"`, `OutputLocation.input="Output Location"` |
| `src/types.ts:65-80` | `AthenaDataSourceOptions extends AwsAuthDataSourceJsonData` adds `catalog`, `database`, `workgroup`, `outputLocation` (all camelCase); `AthenaDataSourceSecureJsonData extends AwsAuthDataSourceSecureJsonData` adds nothing |
| `pkg/athena/models/settings.go:23-32` | Backend `AthenaDataSourceSettings` embeds `awsds.AWSDatasourceSettings` and adds `Database`/`Catalog`/`WorkGroup`/`OutputLocation`/`ResultReuseEnabled`/`ResultReuseMaxAgeInMinutes` — **note the PascalCase json tags** |
| `pkg/athena/models/settings.go:38-52` | `Load`: unmarshals jsonData verbatim, then copies `accessKey`/`secretKey`/`sessionToken` from decrypted secure data (case-insensitive JSON decoding makes the PascalCase↔camelCase mismatch harmless) |
| `pkg/athena/models/settings.go:54-81` | `Apply`: mutates settings from per-query `sqlds.Options` — why the `ResultReuseEnabled` / `ResultReuseMaxAgeInMinutes` json tags exist even though no editor UI writes them |
| `pkg/athena/api/api.go:82-113` | Consumes `Catalog`, `Database`, `WorkGroup`, `OutputLocation` when starting query executions (backend's actual use of the config) |
| `package.json` | External component versions (see next table) |

### External editor components

Read at the exact versions pinned in the plugin's `package.json`.

| Component | Version | Source consulted | What was read |
| --- | --- | --- | --- |
| `ConnectionConfig`, `AwsAuthType`, `Divider`, `ConfigSelect` | `@grafana/aws-sdk@0.10.2` | grafana/grafana-aws-sdk-react tag `v0.10.2` (SHA `fe0c4d8`), `src/components/ConnectionConfig.tsx`, `src/providers.ts`, `src/types.ts`, `src/sql/ConfigEditor/ConfigSelect.tsx`, `src/regions.ts` | Every field's label, placeholder, description, and conditional render; the Select option labels for `authType`; the standard region list (used by the editor when `loadRegions` is not passed); ConfigSelect being a Select that also disables itself until `defaultRegion` is set |
| `ConfigSection`, `DataSourceDescription` | `@grafana/plugin-ui@^0.13.0` | Editor layout / intro block — no storage fields |
| `SecureSocksProxySettings` (excluded) | `@grafana/ui@^12.2.0` | Writes `jsonData.enableSecureSocksProxy`; deliberately excluded from this entry |
| `Field`, `Input`, `SecureInput` (via ConnectionConfig), `Select` (via ConnectionConfig) | `@grafana/ui@^12.2.0` | Prop names (`label`, `placeholder`, `value`, `onChange`, `isConfigured`, `onReset`) so we knew which UI attributes to record |
| `DataSourcePluginOptionsEditorProps`, `onUpdateDatasourceJsonDataOption`, `onUpdateDatasourceSecureJsonDataOption` | `@grafana/data@^12.2.0` | Storage-key semantics of the update helpers used by ConnectionConfig and by Athena's editor |

### Backend Go dependency (`grafana-aws-sdk`)

| File | What was read |
| --- | --- |
| `pkg/awsds/settings.go:13-91` (`v1.4.3`) | `AuthType` int enum + custom `MarshalJSON`/`UnmarshalJSON` — the storage⇆enum mapping (including legacy `arn`→`default` and `sharedCreds`→`credentials`) that we surface as `AWSAuthType` string constants |
| `pkg/awsds/settings.go:94-117` (`v1.4.3`) | `AWSDatasourceSettings` struct with the AWS-shared fields; note `AssumeRoleARN string \`json:"assumeRoleARN"\`` (uppercase `ARN`) while the frontend writes camelCase `assumeRoleArn` — case-insensitive Unmarshal makes both work |
| `pkg/awsds/settings.go:120-141` (`v1.4.3`) | `Load` copies decrypted `accessKey`, `secretKey`, `sessionToken`, and `proxyPassword` from secure JSON data; also mirrors `defaultRegion`→`region` at load time |

## Field provenance

Every dsconfig field, traced from its schema `id` to the upstream `file:line` where each
of its label, placeholder, tooltip, default, storage key, and value type is defined.

| Schema `id` | Storage key | Target | Label source | Placeholder / options / default source | Value type source | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `jsonData_authType` | `authType` | `jsonData` | `ConnectionConfig.tsx:105` (`<Field label="Authentication Provider">`) | Options `providers.ts:4-24` (`awsAuthProviderOptions`); description `ConnectionConfig.tsx:106`; default `awsds/settings.go:16` (`AuthTypeDefault` iota) → `"default"` | `AWSDatasourceSettings.AuthType` `awsds/settings.go:97` (int enum, string on wire) | Role `auth.discriminator`; validation `allowedValues` includes legacy `arn` |
| `jsonData_profile` | `profile` | `jsonData` | `ConnectionConfig.tsx:123` (`<Field label="Credentials Profile Name">`) | Placeholder `ConnectionConfig.tsx:129` (`"default"`); description `:124` | `AWSDatasourceSettings.Profile` `awsds/settings.go:95` | `dependsOn` from conditional render `ConnectionConfig.tsx:121` |
| `secureJsonData_accessKey` | `accessKey` | `secureJsonData` | `ConnectionConfig.tsx:137` (`<Field label="Access Key ID">`) | — | `AWSDatasourceSettings.AccessKey` `awsds/settings.go:113` | Role `auth.aws.accessKeyId`; `dependsOn`/`requiredWhen` from conditional render `ConnectionConfig.tsx:135` |
| `secureJsonData_secretKey` | `secretKey` | `secureJsonData` | `ConnectionConfig.tsx:157` (`<Field label="Secret Access Key">`) | — | `AWSDatasourceSettings.SecretKey` `awsds/settings.go:114` | Role `auth.aws.secretAccessKey`; `dependsOn`/`requiredWhen` from conditional render `ConnectionConfig.tsx:135` |
| `secureJsonData_sessionToken` | `sessionToken` | `secureJsonData` | — (no UI) | — | `AWSDatasourceSettings.SessionToken` `awsds/settings.go:115` | Role `auth.aws.sessionToken`; tagged `backend-only`; still read at `athena-datasource/pkg/athena/models/settings.go:47` |
| `jsonData_assumeRoleArn` | `assumeRoleArn` | `jsonData` | `ConnectionConfig.tsx:261` (`<Field ... label="Assume Role ARN">`) | Placeholder `:268` (`"arn:aws:iam:*"`); description `:262-264` | `AWSDatasourceSettings.AssumeRoleARN` `awsds/settings.go:98` (backend json tag `assumeRoleARN`) | `dependsOn` from conditional render `ConnectionConfig.tsx:257,181` (only when `!hideAssumeRoleArn && awsAssumeRoleEnabled`); pattern validation is common across all AWS DS packs |
| `jsonData_externalId` | `externalId` | `jsonData` | `ConnectionConfig.tsx:276` (`<Field ... label="External ID">`) | Placeholder `:281`; description `:277` | `AWSDatasourceSettings.ExternalID` `awsds/settings.go:99` | `dependsOn` from conditional render `ConnectionConfig.tsx:273` (not `grafana_assume_role`) |
| `jsonData_endpoint` | `endpoint` | `jsonData` | `ConnectionConfig.tsx:362` (`<Field label="Endpoint">`) | Placeholder `:368` (Athena passes no `defaultEndpoint`, so `'https://{service}.{region}.amazonaws.com'`); description `:363` | `AWSDatasourceSettings.Endpoint` `awsds/settings.go:102` | `dependsOn` from conditional render `ConnectionConfig.tsx:360` (not `grafana_assume_role`) |
| `jsonData_defaultRegion` | `defaultRegion` | `jsonData` | `ConnectionConfig.tsx:376` (`<Field label="Default Region">`) | Description `:377` (verbatim, including the padded backticks `` ` us-west-2 ` ``); options at runtime from `standardRegions` (`regions.ts:1-47`) or via `loadRegions` prop — modelled as `select` with `allowCustom` | `AWSDatasourceSettings.DefaultRegion` `awsds/settings.go:105` | `<Select ... allowCustomValue={true}>` |
| `jsonData_catalog` | `catalog` | `jsonData` | `selectors.ts:18` (`"Data source"`, referenced by `ConfigEditor.tsx:119,129`) | ConfigSelect (Select populated at runtime by `/resources/catalogs`) | `AthenaDataSourceSettings.Catalog string \`json:"Catalog"\`` `pkg/athena/models/settings.go:27` (PascalCase — camelCase spelling still loads via case-insensitive Unmarshal) | `requiredWhen: "true"` mirrors editor gate: `ConfigSelect.tsx:74` disables the select until `defaultRegion` is set, and the backend reads all four selectors |
| `jsonData_database` | `database` | `jsonData` | `selectors.ts:22` (`"Database"`, referenced by `ConfigEditor.tsx:134,144`) | ConfigSelect (populated at runtime by `/resources/databases`, dependent on `catalog`) | `AthenaDataSourceSettings.Database` `pkg/athena/models/settings.go:26` | Same requiredWhen rationale as `catalog` |
| `jsonData_workgroup` | `workgroup` | `jsonData` | `selectors.ts:26` (`"Workgroup"`, referenced by `ConfigEditor.tsx:150,160`) | ConfigSelect (populated at runtime by `/resources/workgroups`) | `AthenaDataSourceSettings.WorkGroup` `pkg/athena/models/settings.go:28` | Same requiredWhen rationale |
| `jsonData_outputLocation` | `outputLocation` | `jsonData` | `selectors.ts:30` (`"Output Location"`, referenced by `ConfigEditor.tsx:165`) | Placeholder `ConfigEditor.tsx:172` (`"s3://"`); description `ConfigEditor.tsx:166` | `AthenaDataSourceSettings.OutputLocation` `pkg/athena/models/settings.go:29` | Optional — falls back to the workgroup's own default per description |

## Field inventory summary

| Schema field | Storage key | Target | Editor label | Read by backend? |
| --- | --- | --- | --- | --- |
| `jsonData_authType` | `authType` | `jsonData` | Authentication Provider | Yes |
| `jsonData_profile` | `profile` | `jsonData` | Credentials Profile Name | Yes |
| `secureJsonData_accessKey` | `accessKey` | `secureJsonData` | Access Key ID | Yes |
| `secureJsonData_secretKey` | `secretKey` | `secureJsonData` | Secret Access Key | Yes |
| `secureJsonData_sessionToken` | `sessionToken` | `secureJsonData` | — (no UI) | Yes (backend-only) |
| `jsonData_assumeRoleArn` | `assumeRoleArn` | `jsonData` | Assume Role ARN | Yes |
| `jsonData_externalId` | `externalId` | `jsonData` | External ID | Yes |
| `jsonData_endpoint` | `endpoint` | `jsonData` | Endpoint | Yes |
| `jsonData_defaultRegion` | `defaultRegion` | `jsonData` | Default Region | Yes |
| `jsonData_catalog` | `catalog` | `jsonData` | Data source | Yes |
| `jsonData_database` | `database` | `jsonData` | Database | Yes |
| `jsonData_workgroup` | `workgroup` | `jsonData` | Workgroup | Yes |
| `jsonData_outputLocation` | `outputLocation` | `jsonData` | Output Location | Yes |

### Frontend-only settings

None. The Athena config editor writes only fields the backend consumes.

### Backend-only settings

- **`sessionToken`** — no editor UI; still read by `AWSDatasourceSettings.Load` at
  `awsds/settings.go:137` and copied into `AthenaDataSourceSettings.SessionToken` at
  `pkg/athena/models/settings.go:47`. Used with `authType: "keys"` for temporary STS
  credentials. Provisioning is the practical way to set it.

### Fields excluded from this entry

- **AWS proxy fields** (`proxyType`, `proxyUrl`, `proxyUsername`, `proxyPassword`) —
  `ConnectionConfig.tsx:291` only renders the proxy subsection when the caller passes
  `showHttpProxySettings`. Athena's `ConfigEditor.tsx:112` does not, so proxy fields are
  neither editor-visible nor part of Athena's declared surface, even though `awsds.Load`
  would still consume them if provisioned. Consumers needing them should reach for the
  shared `aws_sdk_settings.json` pack directly.
- **Secure Socks Proxy** (`jsonData.enableSecureSocksProxy`) — excluded per AGENTS.md.
- **`region`** — the AWS SDK Go struct has a `Region` field
  (`awsds/settings.go:96`), but the frontend never writes it; the backend mirrors
  `defaultRegion` into it at load time (`awsds/settings.go:127-129`). Not a stored config.
- **`ResultReuseEnabled` / `ResultReuseMaxAgeInMinutes`** — declared with json tags on
  the backend `AthenaDataSourceSettings`, but only ever populated by
  `AthenaDataSourceSettings.Apply` (`pkg/athena/models/settings.go:54-80`) from per-query
  `sqlds.Options`. No editor UI writes them; treating them as datasource-level config
  would be misleading. See Upstream findings #3.

## Where the types are defined

The Athena configuration types are spread across the plugin and its dependencies. Some
fields and base types come from libraries/SDKs rather than the plugin itself.

### Frontend (TypeScript)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `AthenaDataSourceOptions` (jsonData), `AthenaDataSourceSecureJsonData` | `src/types.ts:68-78` | plugin ([grafana/athena-datasource](https://github.com/grafana/athena-datasource)) |
| `AwsAuthType`, `AwsAuthDataSourceJsonData` (base of `AthenaDataSourceOptions`), `AwsAuthDataSourceSecureJsonData` (base of `AthenaDataSourceSecureJsonData`) | `src/types.ts:3-32` | `@grafana/aws-sdk` `0.10.2` (grafana/grafana-aws-sdk-react `v0.10.2`) |
| `awsAuthProviderOptions` (Select options for `authType`) | `src/providers.ts:4-25` | `@grafana/aws-sdk` `0.10.2` |
| `ConnectionConfig` React component (renders every AWS-shared field) | `src/components/ConnectionConfig.tsx:36-404` | `@grafana/aws-sdk` `0.10.2` |
| `ConfigSelect` (used for catalog/database/workgroup) | `src/sql/ConfigEditor/ConfigSelect.tsx:45-97` | `@grafana/aws-sdk` `0.10.2` |
| `standardRegions` list | `src/regions.ts:1-47` | `@grafana/aws-sdk` `0.10.2` |
| `DataSourceJsonData` (base type of `AwsAuthDataSourceJsonData`) | `packages/grafana-data/src/types/datasource.ts` | `@grafana/data` `^12.2.0` |
| `ConfigSection`, `DataSourceDescription` | `src/components/ConfigEditor/` | `@grafana/plugin-ui` `^0.13.0` |
| `SecureSocksProxySettings` (excluded from this entry) | `packages/grafana-ui/src/components/DataSourceSettings/SecureSocksProxySettings.tsx` | `@grafana/ui` `^12.2.0` |

### Backend (Go)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `AthenaDataSourceSettings`, `Load`, `Apply` | `pkg/athena/models/settings.go:15-81` | plugin ([grafana/athena-datasource](https://github.com/grafana/athena-datasource)) |
| `AWSDatasourceSettings` (embedded base of `AthenaDataSourceSettings`), `AuthType` int enum + custom Marshal/Unmarshal | `pkg/awsds/settings.go:13-141` | `github.com/grafana/grafana-aws-sdk` `v1.4.3` |
| `backend.DataSourceInstanceSettings` (carries `JSONData`, `DecryptedSecureJSONData`) | `backend/common.go` | `github.com/grafana/grafana-plugin-sdk-go` `v0.292.1` |
| SQL driver, connection layer that consumes the flattened settings | `pkg/athena/driver/*.go`, `pkg/athena/api/api.go` | plugin ([grafana/athena-datasource](https://github.com/grafana/athena-datasource)) |
| AWS SDK v2 client build-up (`athena.NewFromConfig(...)`, `sts.NewFromConfig(...)`) | — | `github.com/aws/aws-sdk-go-v2` |

The models in this entry flatten that spread into a single Go `Config` type (jsonData
fields + `DecryptedSecureJSONData`) plus a `SecureJsonDataKey` typed constant list.
`settings.ts` keeps the three canonical TypeScript types (`RootConfig`, `JsonDataConfig`,
`SecureJsonDataConfig`). `AWSAuthType` constants in `settings.go` mirror the string forms
`awsds.AuthType` (Un)Marshals, without carrying the int-enum + custom-JSON machinery.

## Modeling decisions

- **Single, flat Go `Config`.** The upstream `AthenaDataSourceSettings` embeds
  `awsds.AWSDatasourceSettings`. We flatten those into one struct to match the pattern
  used by other dsconfig registry entries (see `registry/grafana-github-datasource`) and
  to avoid pulling `grafana-aws-sdk` into the shared registry `go.mod`.
- **camelCase json tags on the Go struct** — the schema and the Go struct use
  `catalog`/`database`/`workgroup`/`outputLocation` (what the editor writes), not the
  PascalCase spelling on the backend struct. Both spellings load correctly because Go's
  `encoding/json` performs case-insensitive matching; the PascalCase discrepancy is
  documented as [Upstream findings](#upstream-findings) #1. A dedicated LoadConfig test
  (`pascalcase keys are accepted (case-insensitive decode)`) locks this in.
- **`AssumeRoleARN` field vs `assumeRoleArn` tag** — the Go field name mirrors the
  upstream `awsds.AWSDatasourceSettings.AssumeRoleARN`, but the json tag mirrors what the
  frontend writes (`assumeRoleArn`). Same rationale as the PascalCase quirk above.
- **`requiredWhen` for Athena selectors** — the editor doesn't mark them required, but
  `pkg/athena/api/api.go:82-90` uses all four (`Catalog`, `Database`, `WorkGroup`,
  `OutputLocation`) to start a query and the editor's own Save-and-test flow refuses to
  proceed until they are chosen (`ConfigSelect.tsx:74` disables the select until
  `defaultRegion` is set). We encode that runtime contract as `requiredWhen: "true"` on
  catalog/database/workgroup (outputLocation stays optional because the workgroup's own
  default backfills it).
- **`grafana_assume_role` is a schema-only value** — the editor only renders it when
  the `awsDatasourcesTempCredentials` feature toggle is on and the plugin is in the
  allow-list (`ConnectionConfig.tsx:18-28,53-64`). We list it as a schema `allowedValues`
  entry, mark it as visible-when in the depends-on expression of assumeRoleArn / externalId /
  endpoint, and note the gating in an instruction.
- **`arn` kept in `allowedValues`, not in UI options** — `AwsAuthType.ARN` is deprecated
  (`grafana-aws-sdk-react/src/types.ts:11`) and does not appear in
  `awsAuthProviderOptions`, but stored datasources may still carry it. Listing it in
  `allowedValues` (but not in the Select's `options`) matches the backend's tolerance.
- **Athena selectors modelled as `select`, not `input`** — despite fetching options at
  runtime via `/resources/*`, they are stored as plain strings; the `select` UI hint
  captures the editor component. No `options` are inlined because the values are
  account-specific.
- **`sessionToken` included as a secure key with no UI** — the plugin doesn't offer a UI
  field, but `pkg/athena/models/settings.go:47` reads it from decrypted secure data. Tagged
  `backend-only`; matches the conformance suite's `SecureValuesMatchLoadSettings` check.
- **AWS proxy fields excluded** — Athena does not pass `showHttpProxySettings` to
  `ConnectionConfig`, so the proxy fields are not part of Athena's editor surface. They
  are not in the schema even though the shared `awsds.Load` would technically consume
  them. See [Field inventory summary → Fields excluded](#fields-excluded-from-this-entry).
- **`SecureJsonDataConfig` is a key list** — secure values are write-only, so the type
  is just the array of secret key names (`accessKey`, `secretKey`, `sessionToken`);
  consumers read `secureJsonFields` to see what is configured.

## SDK plugin schema and k8s-style examples (`schema.go`)

`NewSchema()` assembles the `grafana-plugin-sdk-go` `pluginschema.PluginSchema` bundle
(the k8s-style schema Grafana's datasource API server serves as `{apiVersion}.json`,
`v0alpha1` today) from the embedded `dsconfig.json`: root fields plus a nested `jsonData`
object become the OpenAPI settings `spec`, secure fields become `secureValues`, and
virtual fields (none here) are skipped.

`SettingsExamples()` provides the default configuration plus one k8s-style example per
AWS auth provider, plus one that combines keys auth with an STS AssumeRole, plus one
legacy `arn` example:

| Example | Auth | Extras | `secureJsonData` |
| --- | --- | --- | --- |
| `""` (default) | AWS SDK Default | — | `accessKey` (empty) |
| `awsSdkDefault` | AWS SDK Default | region + selectors | `accessKey` (empty) |
| `accessAndSecretKey` | Access & secret key | region + selectors + `outputLocation` | `accessKey`, `secretKey` |
| `credentialsFile` | Credentials file | `profile`, region + selectors | `accessKey` (empty) |
| `workspaceIamRole` | Workspace IAM Role | region + selectors | `accessKey` (empty) |
| `grafanaAssumeRole` | Grafana Assume Role | region + selectors | `accessKey` (empty) |
| `assumeRoleFromKeys` | Access & secret key + STS AssumeRole | `assumeRoleArn`, `externalId` | `accessKey`, `secretKey` |
| `legacyArnAuthType` | `arn` (legacy — backend maps to `default`) | region + selectors | `accessKey` (empty) |

Every example carries at least one `secureJsonData` placeholder as required by the
conformance suite.

## `LoadConfig` utility (`settings.go`)

`LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error)`
runs the full three-phase load flow on a datasource instance's settings and returns a
fully-defaulted, validated `Config`:

1. **Parse** — unmarshal `settings.JSONData` into `Config` (Go's case-insensitive
   `encoding/json` accepts either the frontend's camelCase Athena keys or the backend's
   PascalCase spelling), then copy the plugin's decrypted secrets
   (`accessKey`/`secretKey`/`sessionToken`) into `DecryptedSecureJSONData`. Mirrors
   `pkg/athena/models/settings.go:38-52`.
2. **`ApplyDefaults`** — fills the single curated default: `AuthType` defaults to
   `AWSAuthTypeDefault`, matching both the reference `aws_sdk_settings.json` pack and the
   backend `awsds.AuthTypeDefault` (iota zero). Athena selectors intentionally have no
   default because they must be picked from the connected AWS account.
3. **`Validate`** — enforces the runtime contract: known `AuthType`, `accessKey` +
   `secretKey` present for `keys` auth, and non-empty `defaultRegion`, `catalog`,
   `database`, `workgroup` for any query to actually run. Errors are joined so callers
   see every problem at once.

Everything is logged via `backend.Logger.FromContext(ctx)` with `datasource_uid`,
`datasource_name`, and `plugin` labels so log lines carry request context.

### Direct access to individual phases

`(*Config).ApplyDefaults()` and `(Config).Validate() error` are exported for callers that
want to compose them themselves (provisioning preview, schema-example round-trip, tests
that need to distinguish parse-level from policy-level errors). Skip them by never
calling `LoadConfig` in those flows — assemble a `Config` directly.

## Upstream findings

Potential bugs, misleading UX, and consistency issues discovered while researching
upstream. All preserved verbatim in the schema — the schema records what the plugin
**does**, not what it **should** do; these notes exist so reviewers can reproduce each
finding and decide separately whether to fix upstream.

1. **PascalCase vs camelCase storage keys for the Athena block.** The frontend
   (`src/ConfigEditor.tsx:81-101`) writes `jsonData.catalog`, `jsonData.database`,
   `jsonData.workgroup`, `jsonData.outputLocation` (camelCase), but the backend
   `AthenaDataSourceSettings` (`pkg/athena/models/settings.go:26-29`) declares those
   fields with `` json:"Catalog" ``, `` json:"Database" ``, `` json:"WorkGroup" ``,
   `` json:"OutputLocation" `` (PascalCase). It works because Go's `encoding/json`
   accepts case-insensitive matches. It's still a latent trap: a strict, standards-only
   JSON decoder in a different language will not match, and a linter that ever tightens
   Go's tag matching would silently break every existing Athena datasource. A single
   canonical spelling on both sides would be safer.
2. **`AssumeRoleARN` json tag disagrees with the frontend's `assumeRoleArn`.**
   `awsds/settings.go:98` uses `` json:"assumeRoleARN" `` (uppercase RN), but the frontend
   type (`grafana-aws-sdk-react/src/types.ts:17`) and every `onChange` in ConnectionConfig
   write `assumeRoleArn` (lowercase arn). Same case-insensitive-match rescue as above.
3. **`ResultReuseEnabled` / `ResultReuseMaxAgeInMinutes` are query-level settings with
   datasource-level json tags.** `pkg/athena/models/settings.go:30-31` declares json tags
   on these fields, but the frontend never writes them at datasource level; they only
   ever get populated by `Apply` from per-query `sqlds.Options`. Provisioning could
   accidentally set them at datasource level with confusing results.
4. **The "Grafana Assume Role" provider only appears when a feature toggle is on and the
   plugin is in an allow-list.** `ConnectionConfig.tsx:18-28,53-64` restricts the provider
   to plugins listed in `DS_TYPES_THAT_SUPPORT_TEMP_CREDS` and further gates it on
   `config.featureToggles.awsDatasourcesTempCredentials`. The list is hardcoded in
   `@grafana/aws-sdk`; adding a new AWS plugin requires editing the SDK. Storage-side the
   value is valid regardless, so a provisioned config with `authType:
   "grafana_assume_role"` still loads on a Grafana instance where the toggle is off.
5. **Assume-role rendering is triple-gated in unclear ways.** The whole Assume Role
   subsection hides when the *editor* prop `hideAssumeRoleArn=true` (`ConnectionConfig.tsx:181`);
   inside it, the two inputs render only when `awsAssumeRoleEnabled`
   (`ConnectionConfig.tsx:57,257`); and `externalId` further hides when `authType` is
   `grafana_assume_role` (`:273`). Athena doesn't pass `hideAssumeRoleArn`, so its default
   is `false`; `awsAssumeRoleEnabled` comes from a Grafana runtime config flag that
   defaults to `true`. In practice the assumeRoleArn and externalId inputs are visible
   for every non-`grafana_assume_role` provider — but the code makes that hard to see at
   a glance.
6. **Deprecated `arn` auth value silently maps to `default`.** `awsds/settings.go:87-88`
   maps any unknown auth type string to `AuthTypeDefault`, comment: "For old 'arn'
   option". A provisioned config with `authType: "arn"` will load as if it were
   `default`, and there is no warning surfaced anywhere.
7. **Legacy `sharedCreds` auth value silently mapped to `credentials`.**
   `awsds/settings.go:75-78` (`case "credentials"` falls through to `sharedCreds`) folds
   both storage values onto the same enum. Not preserved in this schema as an allowed
   value (the reference pack also omits it); a datasource stored with `authType:
   "sharedCreds"` would fail the schema's `allowedValues` check but still load fine on
   the backend.
8. **Auth type default depends on Grafana instance config.** The editor's `useEffect`
   (`ConnectionConfig.tsx:75-90`) picks `awsAllowedAuthProviders[0]`, and prefers
   `grafana_assume_role` if the feature is on. So the "default" auth type new users see
   is Grafana-instance-dependent, not always `default`. The stored schema default here is
   `default` (matching the reference pack and the backend iota zero), which may not be
   what a fresh Grafana Cloud editor writes.
9. **`sessionToken` has no editor UI.** `grafana-aws-sdk-react` never exposes an input
   for `sessionToken` (`grep -R sessionToken src/` finds only the type declaration), even
   though the backend reads it and it is required for temporary credentials. Users must
   provision it directly.
10. **The "Default Region" description has padded backticks.** `ConnectionConfig.tsx:377`
    reads `` `Specify the region, such as for US West (Oregon) use ` us-west-2 ` as the
    region.` `` — note the spaces around `us-west-2` inside the backticks. Preserved
    verbatim in the schema; would render inline code with padding in some Markdown
    renderers.

## Validation performed

- `dsconfig.ParseAndResolveSchemaJSON` + `Schema.Validate()` — passes.
- JSON Schema validation against [`dsconfig/schema.json`](../../dsconfig/schema.json)
  (draft-07, `additionalProperties: false`) — passes.
- `go test ./...` on this module — passes (schema bundle shape, `SchemaSpecHasNoSecureJSON`,
  `SecureValuesMatchLoadSettings`, `JSONDataMatchesStruct`, `JSONDataTypesMatchStruct`,
  `SchemaArtifactInSync`, `LoadConfig` including case-insensitive PascalCase decode,
  `ApplyDefaults`, `Validate` per auth type).
- `settings.go` / `schema.go`: `go build`, `go vet`, `gofmt` — clean.
- `settings.ts`: `tsc --noEmit --strict` — clean.
