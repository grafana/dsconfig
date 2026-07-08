# grafana-timestream-datasource

Declarative configuration schema for the [Amazon Timestream datasource plugin](https://github.com/grafana/timestream-datasource) (`grafana-timestream-datasource`).

## Upstream researched

- **Repo**: `github.com/grafana/timestream-datasource`
- **Ref**: `main`
- **Commit SHA**: `9e34c64a3bb9208a11b15617d9d523432527bc4f` (`security hardening (#684)`)

Every value in [`dsconfig.json`](dsconfig.json) — labels, placeholders, tooltips, option
labels/values, section titles, defaults, validations, dependency and required-when
expressions, storage keys, storage targets, value types, group titles, and instructions —
is traceable to a specific `file:line` in the upstream plugin repo (or in the pinned
`@grafana/aws-sdk` version of the shared `ConnectionConfig` component) at this SHA. See
[Field provenance](#field-provenance) below.

To reproduce this research:

```bash
git clone --depth 20 https://github.com/grafana/timestream-datasource
cd timestream-datasource
git checkout 9e34c64a3bb9208a11b15617d9d523432527bc4f

# The AWS auth surface is rendered by @grafana/aws-sdk's ConnectionConfig. Pin
# to the version timestream-datasource's package.json uses (currently 0.10.2):
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
| [`settings.go`](settings.go) | Go `Config` model (flat: jsonData fields + root `Database` legacy field + `DecryptedSecureJSONData`), `PluginID`, `AWSAuthType` typed constants, `SecureJsonDataKey` typed constants, and the `LoadConfig` utility |
| [`schema.go`](schema.go) | k8s-style SDK plugin schema: embeds `dsconfig.json`, converts it via `dsconfig.NewSDKSchema`, and defines `SettingsExamples` for each AWS auth provider plus a Timestream-macro-defaults variant |
| [`settings_test.go`](settings_test.go) | Table tests for `LoadConfig`, `ApplyDefaults`, `Validate`, and `EffectiveProfile` |
| [`conformance_test.go`](conformance_test.go) | Runs `schema.RunPluginTests` — the shared dsconfig conformance suite — against `dsconfig.json` |
| [`schema.gen.json`](schema.gen.json), [`settings.gen.json`](settings.gen.json), [`settings.examples.gen.json`](settings.examples.gen.json) | Committed schema artifacts (regenerate with `go generate ./...` inside this directory; the `SchemaArtifactInSync` conformance subtest guards drift) |

There is no per-entry `go.mod` — every registry entry is a subpackage of the shared
[`registry/`](..) module (`github.com/grafana/dsconfig/registry`).

## Sources researched

Every source below was read at the pinned upstream SHA
(`9e34c64a3bb9208a11b15617d9d523432527bc4f`) or, for `@grafana/aws-sdk`, at the exact
version pinned in the plugin's `package.json` (`0.10.2` / SHA `fe0c4d8`).

### Plugin repo (`github.com/grafana/timestream-datasource@9e34c64`)

| File | What was read |
| --- | --- |
| `src/plugin.json:2-38` | `pluginType` (`id`), `pluginName` (`name`), `docURL` (Grafana plugin catalog URL) |
| `src/components/ConfigEditor.tsx:1-151` | Editor layout, `ConnectionConfig` composition with `standardRegions` and `defaultEndpoint="https://query-{cell}.timestream.{region}.amazonaws.com"`, the `<ConfigSection title="Timestream Details" description="Default values to be used as macros">`, three `<ConfigSelect>` fields for defaultDatabase/defaultTable/defaultMeasure with cascading dependencies |
| `src/components/ConfigEditor.tsx:65-74` | The `onChange` handler writes camelCase keys to jsonData: `defaultDatabase`, `defaultTable`, `defaultMeasure` |
| `src/components/ConfigEditor.tsx:84-86` | Conditional `SecureSocksProxySettings` render — writes `jsonData.enableSecureSocksProxy`; deliberately excluded per AGENTS.md |
| `src/components/selectors.ts:14-27` | Field labels: `defaultDatabase.input="Database"`, `defaultTable.input="Table"`, `defaultMeasure.input="Measure"` |
| `src/regions.ts:1-11` | Timestream-specific standardRegions list (9 regions — Timestream is not available in every AWS region) passed to ConnectionConfig |
| `src/types.ts:94-102` | `TimestreamOptions extends AwsAuthDataSourceJsonData` adds `defaultDatabase?`, `defaultTable?`, `defaultMeasure?` (all camelCase); `TimestreamSecureJsonData extends AwsAuthDataSourceSecureJsonData` adds nothing (comment: "nothing for now") |
| `pkg/models/settings.go:12-21` | Backend `DatasourceSettings` embeds `awsds.AWSDatasourceSettings` and adds `DefaultDatabase`, `DefaultTable`, `DefaultMeasure` with camelCase json tags matching the frontend |
| `pkg/models/settings.go:23-44` | `Load`: unmarshals jsonData verbatim, then adds two legacy fallbacks: `Region → DefaultRegion` when Region is empty/"default" (`:32-34`) and `Profile → config.Database` when Profile is empty (`:36-38`, comment: `"legacy support (only for cloudwatch?)"`) |
| `pkg/models/settings.go:40-42` | Copies `accessKey`, `secretKey`, `sessionToken` from decrypted secure data |
| `pkg/models/settings_test.go:9-31` | Confirms the backend Load unmarshals into `DefaultDatabase`, `DefaultRegion`, `DefaultTable`, `DefaultMeasure` |
| `package.json:23-91` | External component versions (see next table) |
| `go.mod:5-12` | `grafana-aws-sdk v1.4.4`, `grafana-plugin-sdk-go v0.292.0`, `aws-sdk-go-v2/service/timestreamquery v1.36.16` |

### External editor components

Read at the exact versions pinned in the plugin's `package.json`.

| Component | Version | Source consulted | What was read |
| --- | --- | --- | --- |
| `ConnectionConfig`, `AwsAuthType`, `ConfigSelect` | `@grafana/aws-sdk@0.10.2` | grafana/grafana-aws-sdk-react tag `v0.10.2` (SHA `fe0c4d8`), `src/components/ConnectionConfig.tsx`, `src/providers.ts`, `src/types.ts`, `src/sql/ConfigEditor/ConfigSelect.tsx` | Every AWS field's label, placeholder, description, and conditional render; the Select option labels for `authType`; the ConfigSelect widget that disables itself until `defaultRegion` is set; the `DS_TYPES_THAT_SUPPORT_TEMP_CREDS` allow-list that includes `grafana-timestream-datasource` (line 26) |
| `ConfigSection` | `@grafana/plugin-ui@0.13.1` | Editor section header — `title="Timestream Details"` and `description="Default values to be used as macros"` |
| `SecureSocksProxySettings` (excluded) | `@grafana/ui@12.4.1` | Writes `jsonData.enableSecureSocksProxy`; deliberately excluded per AGENTS.md |
| `Field`, `Divider` (via `@grafana/ui`), `Select` / `Input` (via ConnectionConfig) | `@grafana/ui@12.4.1` | Prop names (`label`, `placeholder`, `value`, `onChange`, `isConfigured`, `onReset`) so we knew which UI attributes to record |
| `DataSourcePluginOptionsEditorProps`, `onUpdateDatasourceJsonDataOption`, `onUpdateDatasourceSecureJsonDataOption` | `@grafana/data@12.4.1` | Storage-key semantics of the update helpers used by ConnectionConfig |

### Backend Go dependency (`grafana-aws-sdk`)

| File | What was read |
| --- | --- |
| `pkg/awsds/settings.go` (`v1.4.4`) | `AuthType` int enum + custom `MarshalJSON`/`UnmarshalJSON` — the storage⇆enum mapping (including legacy `arn`→`default` and `sharedCreds`→`credentials`) that we surface as `AWSAuthType` string constants |
| `pkg/awsds/settings.go` (`v1.4.4`) | `AWSDatasourceSettings` struct with the AWS-shared fields; note `AssumeRoleARN string \`json:"assumeRoleARN"\`` (uppercase `ARN`) while the frontend writes camelCase `assumeRoleArn` — case-insensitive Unmarshal makes both work |

## Field provenance

Every dsconfig field, traced from its schema `id` to the upstream `file:line` where each
of its label, placeholder, tooltip, default, storage key, and value type is defined.

| Schema `id` | Storage key | Target | Label source | Placeholder / options / default source | Value type source | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `jsonData_authType` | `authType` | `jsonData` | `ConnectionConfig.tsx:105` (`<Field label="Authentication Provider">`) | Options `providers.ts:4-25`; description `ConnectionConfig.tsx:106`; default `awsds.AuthTypeDefault` iota-zero → `"default"` | `AWSDatasourceSettings.AuthType` (int enum, string on wire) | Role `auth.discriminator`; validation `allowedValues` includes legacy `arn` |
| `jsonData_profile` | `profile` | `jsonData` | `ConnectionConfig.tsx:123` (`<Field label="Credentials Profile Name">`) | Placeholder `:129` (`"default"`); description `:124` | `AWSDatasourceSettings.Profile` | `dependsOn` from conditional render `ConnectionConfig.tsx:121` |
| `secureJsonData_accessKey` | `accessKey` | `secureJsonData` | `ConnectionConfig.tsx:137` (`<Field label="Access Key ID">`) | — | `AWSDatasourceSettings.AccessKey` | Role `auth.aws.accessKeyId`; `dependsOn`/`requiredWhen` from conditional render `ConnectionConfig.tsx:135` |
| `secureJsonData_secretKey` | `secretKey` | `secureJsonData` | `ConnectionConfig.tsx:157` (`<Field label="Secret Access Key">`) | — | `AWSDatasourceSettings.SecretKey` | Role `auth.aws.secretAccessKey`; `dependsOn`/`requiredWhen` from conditional render `ConnectionConfig.tsx:135` |
| `secureJsonData_sessionToken` | `sessionToken` | `secureJsonData` | — (no UI) | — | `AWSDatasourceSettings.SessionToken` | Role `auth.aws.sessionToken`; tagged `backend-only`; still read at `pkg/models/settings.go:42` |
| `jsonData_assumeRoleArn` | `assumeRoleArn` | `jsonData` | `ConnectionConfig.tsx:261` (`<Field ... label="Assume Role ARN">`) | Placeholder `:268` (`"arn:aws:iam:*"`); description `:262-264` (verbatim, including the 21-space indentation) | `AWSDatasourceSettings.AssumeRoleARN` (backend json tag `assumeRoleARN`) | Visible when `!hideAssumeRoleArn && awsAssumeRoleEnabled` (Timestream doesn't pass `hideAssumeRoleArn`); pattern validation common across AWS DS packs |
| `jsonData_externalId` | `externalId` | `jsonData` | `ConnectionConfig.tsx:276` (`<Field ... label="External ID">`) | Placeholder `:281`; description `:277` | `AWSDatasourceSettings.ExternalID` | `dependsOn` from conditional render `ConnectionConfig.tsx:273` (not `grafana_assume_role`) |
| `jsonData_endpoint` | `endpoint` | `jsonData` | `ConnectionConfig.tsx:362` (`<Field label="Endpoint">`) | Placeholder from `props.defaultEndpoint` (`ConfigEditor.tsx:81` passes `"https://query-{cell}.timestream.{region}.amazonaws.com"`); description `:363` | `AWSDatasourceSettings.Endpoint` | `dependsOn` from conditional render `ConnectionConfig.tsx:360` (not `grafana_assume_role`, and Timestream doesn't pass `skipEndpoint`) |
| `jsonData_defaultRegion` | `defaultRegion` | `jsonData` | `ConnectionConfig.tsx:376` (`<Field label="Default Region">`) | Description `:377` (verbatim, including the padded backticks `` ` us-west-2 ` ``); Timestream-specific options from `regions.ts:1-11` (9-region list) | `AWSDatasourceSettings.DefaultRegion` | `<Select ... allowCustomValue={true}>` |
| `jsonData_defaultDatabase` | `defaultDatabase` | `jsonData` | `selectors.ts:16` (`"Database"`, referenced by `ConfigEditor.tsx:90,101`) | ConfigSelect (populated at runtime by `/resources/databases`) | `DatasourceSettings.DefaultDatabase string \`json:"defaultDatabase,omitempty"\`` `pkg/models/settings.go:18` | Optional — feeds the {{database}} query macro |
| `jsonData_defaultTable` | `defaultTable` | `jsonData` | `selectors.ts:20` (`"Table"`, referenced by `ConfigEditor.tsx:107,118`) | ConfigSelect (populated at runtime by `/resources/tables`, dependent on defaultDatabase) | `DatasourceSettings.DefaultTable` `pkg/models/settings.go:19` | Optional — feeds the {{table}} query macro; `dependsOn` mirrors the editor's `dependencies={[defaultDatabase]}` (`ConfigEditor.tsx:120`) |
| `jsonData_defaultMeasure` | `defaultMeasure` | `jsonData` | `selectors.ts:24` (`"Measure"`, referenced by `ConfigEditor.tsx:125,136`) | ConfigSelect (populated at runtime by `/resources/measures`, dependent on defaultDatabase + defaultTable) | `DatasourceSettings.DefaultMeasure` `pkg/models/settings.go:20` | Optional — feeds the {{measure}} query macro; `dependsOn` mirrors the editor's `dependencies={[defaultDatabase, defaultTable]}` (`ConfigEditor.tsx:138`) |

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
| `jsonData_defaultDatabase` | `defaultDatabase` | `jsonData` | Database | Yes |
| `jsonData_defaultTable` | `defaultTable` | `jsonData` | Table | Yes |
| `jsonData_defaultMeasure` | `defaultMeasure` | `jsonData` | Measure | Yes |

### Frontend-only settings

None. Every editor-written field is either consumed by the backend `DatasourceSettings.Load`
or by the shared `awsds.AWSDatasourceSettings` machinery.

### Backend-only settings

- **`sessionToken`** — no editor UI; still read by `DatasourceSettings.Load` at
  `pkg/models/settings.go:42` (`s.SessionToken = config.DecryptedSecureJSONData["sessionToken"]`).
  Used with `authType: "keys"` for temporary STS credentials. Provisioning is the
  practical way to set it.
- **Root-level `database`** (top-level datasource field, not `jsonData.database`) — no
  editor UI writes it; the backend uses it as a legacy fallback for `Profile` when
  Profile is empty (`pkg/models/settings.go:36-38`). Modelled on `Config.Database`
  (json:"-") so consumers can round-trip it.

### Fields excluded from this entry

- **AWS proxy fields** (`proxyType`, `proxyUrl`, `proxyUsername`, `proxyPassword`) —
  `ConnectionConfig.tsx:291` only renders the proxy subsection when the caller passes
  `showHttpProxySettings`. Timestream's `ConfigEditor.tsx:78-83` does not, so proxy
  fields are neither editor-visible nor part of Timestream's declared surface.
- **Secure Socks Proxy** (`jsonData.enableSecureSocksProxy`) — excluded per AGENTS.md.
- **`region`** — the AWS SDK Go struct has a `Region` field, but the frontend never
  writes it; the backend mirrors `defaultRegion` into it at load time
  (`pkg/models/settings.go:32-34`). Not a stored config.

## Where the types are defined

The Timestream configuration types are spread across the plugin and its dependencies.
Some fields and base types come from libraries/SDKs rather than the plugin itself.

### Frontend (TypeScript)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `TimestreamOptions` (jsonData), `TimestreamSecureJsonData` | `src/types.ts:94-102` | plugin ([grafana/timestream-datasource](https://github.com/grafana/timestream-datasource)) |
| Editor `ResourceType` (`'defaultDatabase' \| 'defaultTable' \| 'defaultMeasure'`) | `src/components/ConfigEditor.tsx:16` | plugin (editor-local; not stored) |
| `AwsAuthType`, `AwsAuthDataSourceJsonData` (base of `TimestreamOptions`), `AwsAuthDataSourceSecureJsonData` (base of `TimestreamSecureJsonData`) | `src/types.ts:3-32` | `@grafana/aws-sdk` `0.10.2` (grafana/grafana-aws-sdk-react `v0.10.2`) |
| `awsAuthProviderOptions` (Select options for `authType`) | `src/providers.ts:4-25` | `@grafana/aws-sdk` `0.10.2` |
| `ConnectionConfig` React component (renders every AWS-shared field) | `src/components/ConnectionConfig.tsx:36-404` | `@grafana/aws-sdk` `0.10.2` |
| `ConfigSelect` (used for defaultDatabase/defaultTable/defaultMeasure) | `src/sql/ConfigEditor/ConfigSelect.tsx:45-97` | `@grafana/aws-sdk` `0.10.2` |
| `standardRegions` list (Timestream-specific 9-region subset) | `src/regions.ts:1-11` | plugin |
| `DataSourceJsonData` (base type of `AwsAuthDataSourceJsonData`) | `packages/grafana-data/src/types/datasource.ts` | `@grafana/data` `12.4.1` |
| `ConfigSection` | `src/components/ConfigEditor/` | `@grafana/plugin-ui` `0.13.1` |
| `Divider`, `Field`, `SecureSocksProxySettings` (excluded from this entry) | `packages/grafana-ui/src/components/` | `@grafana/ui` `12.4.1` |

### Backend (Go)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `DatasourceSettings`, `Load` | `pkg/models/settings.go:12-45` | plugin ([grafana/timestream-datasource](https://github.com/grafana/timestream-datasource)) |
| `AWSDatasourceSettings` (embedded base of `DatasourceSettings`), `AuthType` int enum + custom Marshal/Unmarshal | `pkg/awsds/settings.go` | `github.com/grafana/grafana-aws-sdk` `v1.4.4` |
| `backend.DataSourceInstanceSettings` (carries `JSONData`, `DecryptedSecureJSONData`, `Database`) | `backend/common.go` | `github.com/grafana/grafana-plugin-sdk-go` `v0.292.0` |
| AWS SDK v2 client build-up (`timestreamquery.NewFromConfig(...)`) | — | `github.com/aws/aws-sdk-go-v2/service/timestreamquery` `v1.36.16` |

The models in this entry flatten that spread into a single Go `Config` type (jsonData
fields + `DecryptedSecureJSONData` + root `Database` legacy fallback) plus a
`SecureJsonDataKey` typed constant list. `settings.ts` keeps the three canonical
TypeScript types (`RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig`). `AWSAuthType`
constants in `settings.go` mirror the string forms `awsds.AuthType` (Un)Marshals, without
carrying the int-enum + custom-JSON machinery.

## Modeling decisions

- **Single, flat Go `Config`.** The upstream `DatasourceSettings` embeds
  `awsds.AWSDatasourceSettings`. We flatten those into one struct to match the pattern
  used by other AWS registry entries (`grafana-athena-datasource`, `grafana-x-ray-datasource`,
  `cloudwatch`) and to avoid pulling `grafana-aws-sdk` into the shared registry `go.mod`.
- **camelCase json tags on the Go struct** — the schema and the Go struct use
  `defaultDatabase`/`defaultTable`/`defaultMeasure` matching both the editor and the
  upstream backend struct's json tags (Timestream's backend is already camelCase-clean,
  unlike Athena's mixed PascalCase/camelCase spellings).
- **`AssumeRoleARN` field vs `assumeRoleArn` tag** — the Go field name mirrors the
  upstream `awsds.AWSDatasourceSettings.AssumeRoleARN`, but the json tag mirrors what
  the frontend writes (`assumeRoleArn`). Go's case-insensitive Unmarshal decodes either.
- **Root-level `database` legacy fallback modelled on `Config`.** Same rationale as
  `registry/grafana-x-ray-datasource`: not a schema field (no editor UI), but consumers
  need it to round-trip legacy provisioned configs. `LoadConfig` populates it from
  `settings.Database`. Tagged `json:"-"` so it does not collide with a `database` key
  inside jsonData.
- **Timestream Details macro fields are NOT `requiredWhen: "true"`** — unlike Athena's
  catalog/database/workgroup selectors, the Timestream editor does not gate save on
  them and the backend `Load` does not require them; they are pure query-macro defaults.
- **`dependsOn` on defaultTable / defaultMeasure** mirrors the editor's cascading
  ConfigSelect fetch (`ConfigEditor.tsx:120,138`): defaultTable is only meaningful once
  defaultDatabase is chosen, and defaultMeasure requires both. Editor-visibility is
  encoded, not backend requiredness.
- **`grafana_assume_role` is a schema value with runtime gating** — the editor only
  renders it when the `awsDatasourcesTempCredentials` feature toggle is on AND the
  plugin is in the allow-list `DS_TYPES_THAT_SUPPORT_TEMP_CREDS`
  (`ConnectionConfig.tsx:18-28,53-64`). `grafana-timestream-datasource` is on the list
  (line 26).
- **`arn` kept in `allowedValues`, not in UI options** — `AwsAuthType.ARN` is deprecated
  (`grafana-aws-sdk-react/src/types.ts:11`) and does not appear in
  `awsAuthProviderOptions`, but stored datasources may still carry it. Listing it in
  `allowedValues` (but not in the Select's `options`) matches the backend's tolerance.
- **Timestream selectors modelled as `select`, not `input`** — despite fetching options
  at runtime via `/resources/*`, they are stored as plain strings; the `select` UI hint
  captures the editor component. No `options` are inlined because the values are
  account-specific.
- **`sessionToken` included as a secure key with no UI** — the plugin doesn't offer a
  UI field, but `pkg/models/settings.go:42` reads it from decrypted secure data. Tagged
  `backend-only`; matches the conformance suite's `SecureValuesMatchLoadSettings` check.
- **AWS proxy fields excluded** — Timestream does not pass `showHttpProxySettings` to
  `ConnectionConfig`, so the proxy fields are not part of Timestream's editor surface.
  They are not in the schema even though the shared `awsds.Load` would technically
  consume them.
- **Timestream-specific Endpoint placeholder** — the plugin passes
  `defaultEndpoint="https://query-{cell}.timestream.{region}.amazonaws.com"` to
  `ConnectionConfig` (`ConfigEditor.tsx:81`), so the Endpoint placeholder is
  Timestream-cell-aware, not the generic `{service}.{region}` used by other AWS
  plugins. Preserved verbatim.
- **`SecureJsonDataConfig` is a key list** — secure values are write-only, so the type
  is just the array of secret key names (`accessKey`, `secretKey`, `sessionToken`);
  consumers read `secureJsonFields` to see what is configured.

## SDK plugin schema and k8s-style examples (`schema.go`)

`NewSchema()` assembles the `grafana-plugin-sdk-go` `pluginschema.PluginSchema` bundle
(the k8s-style schema Grafana's datasource API server serves as `{apiVersion}.json`)
from the embedded `dsconfig.json`: root fields plus a nested `jsonData` object become
the OpenAPI settings `spec`, secure fields become `secureValues`, and virtual fields
(none here) are skipped.

`SettingsExamples()` provides the default configuration plus one example per AWS auth
provider, plus one that combines keys auth with an STS AssumeRole, plus one legacy
`arn` example, plus one that seeds the Timestream macro defaults:

| Example | Auth | Extras | `secureJsonData` |
| --- | --- | --- | --- |
| `""` (default) | AWS SDK Default | — | `accessKey` (empty) |
| `awsSdkDefault` | AWS SDK Default | region | `accessKey` (empty) |
| `accessAndSecretKey` | Access & secret key | region | `accessKey`, `secretKey` |
| `credentialsFile` | Credentials file | `profile`, region | `accessKey` (empty) |
| `workspaceIamRole` | Workspace IAM Role | region | `accessKey` (empty) |
| `grafanaAssumeRole` | Grafana Assume Role | region | `accessKey` (empty) |
| `assumeRoleFromKeys` | Access & secret key + STS AssumeRole | `assumeRoleArn`, `externalId`, region | `accessKey`, `secretKey` |
| `legacyArnAuthType` | `arn` (legacy — backend maps to `default`) | region | `accessKey` (empty) |
| `withMacroDefaults` | AWS SDK Default | region + `defaultDatabase`, `defaultTable`, `defaultMeasure` | `accessKey` (empty) |

Every example carries at least one `secureJsonData` placeholder as required by the
conformance suite.

## `LoadConfig` utility (`settings.go`)

`LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error)`
runs the full three-phase load flow on a datasource instance's settings and returns a
fully-defaulted, validated `Config`:

1. **Parse** — unmarshal `settings.JSONData` into `Config`, capture the top-level
   `settings.Database` as the legacy profile fallback, then copy the plugin's decrypted
   secrets (`accessKey`/`secretKey`/`sessionToken`) into `DecryptedSecureJSONData`.
   Mirrors `pkg/models/settings.go:24-44`.
2. **`ApplyDefaults`** — fills the single curated default: `AuthType` defaults to
   `AWSAuthTypeDefault`, matching both the reference `aws_sdk_settings.json` pack and
   the backend `awsds.AuthTypeDefault` (iota zero). `DefaultRegion` and the Timestream
   macro fields intentionally have no default because they must be picked from the
   connected AWS account / Timestream store.
3. **`Validate`** — enforces the runtime contract: known `AuthType`, `accessKey` +
   `secretKey` present for `keys` auth, and non-empty `defaultRegion` for any query to
   actually run. The macro defaults (`defaultDatabase`/`defaultTable`/`defaultMeasure`)
   are not validated because they are optional and can be overridden per query. Errors
   are joined so callers see every problem at once.

Everything is logged via `backend.Logger.FromContext(ctx)` with `datasource_uid`,
`datasource_name`, and `plugin` labels so log lines carry request context.

### Direct access to individual phases

`(*Config).ApplyDefaults()` and `(Config).Validate() error` are exported for callers
that want to compose them themselves (provisioning preview, schema-example round-trip,
tests that need to distinguish parse-level from policy-level errors). Skip them by
never calling `LoadConfig` in those flows — assemble a `Config` directly.

`(Config).EffectiveProfile() string` returns the profile the plugin actually uses at
runtime, mirroring the legacy fallback in `pkg/models/settings.go:36-38`: explicit
`jsonData.profile` wins; otherwise the top-level `database` value is used.

## Upstream findings

Potential bugs, misleading UX, and consistency issues discovered while researching
upstream. All preserved verbatim in the schema — the schema records what the plugin
**does**, not what it **should** do; these notes exist so reviewers can reproduce each
finding and decide separately whether to fix upstream.

1. **`Profile` falls back to top-level `database` with an uncertain comment.**
   `pkg/models/settings.go:36-38` copies the datasource's top-level `Database` field
   into `Profile` when Profile is empty, with the comment `// legacy support (only for
   cloudwatch?)`. The comment is copy-pasted from other AWS plugins and the "(only for
   cloudwatch?)" hedging suggests the author was unsure whether Timestream needs the
   fallback. It still fires for provisioned configs that carry a top-level `database`
   value from years-old exports.
2. **`AssumeRoleARN` json tag disagrees with the frontend's `assumeRoleArn`.**
   `awsds/settings.go` uses `` json:"assumeRoleARN" `` (uppercase RN), but the frontend
   type (`grafana-aws-sdk-react/src/types.ts:17`) and every `onChange` in
   ConnectionConfig write `assumeRoleArn` (lowercase arn). Case-insensitive-match
   rescue in Go; would break strict JSON decoders in other languages.
3. **The "Grafana Assume Role" provider only appears when a feature toggle is on and
   the plugin is in an allow-list.** `ConnectionConfig.tsx:18-28,53-64` restricts the
   provider to plugins listed in `DS_TYPES_THAT_SUPPORT_TEMP_CREDS`
   (`grafana-timestream-datasource` is on the list) and further gates it on
   `config.featureToggles.awsDatasourcesTempCredentials`. The list is hardcoded in
   `@grafana/aws-sdk`; adding a new AWS plugin requires editing the SDK. Storage-side
   the value is valid regardless, so a provisioned config with `authType:
   "grafana_assume_role"` still loads on a Grafana instance where the toggle is off.
4. **Deprecated `arn` auth value silently maps to `default`.**
   `awsds/settings.go` maps any unknown auth type string to `AuthTypeDefault`, with a
   comment about the "old 'arn' option". A provisioned config with `authType: "arn"`
   loads as if it were `default`, and there is no warning surfaced anywhere.
5. **Legacy `sharedCreds` auth value silently mapped to `credentials`.**
   `awsds/settings.go` folds both storage values onto the same enum. Not preserved in
   this schema as an allowed value; a datasource stored with `authType: "sharedCreds"`
   would fail the schema's `allowedValues` check but still load fine on the backend.
6. **Timestream's `standardRegions` list is manually curated, so it can drift.**
   `src/regions.ts` hardcodes 9 regions. AWS has added Timestream regions over time
   (e.g., ap-southeast-1, eu-north-1) and the list has not been kept in sync. A
   provisioned config with a real Timestream region outside the list still works
   because `<Select allowCustomValue={true}>` accepts anything, but the dropdown will
   not show it. The schema keeps the same 9-region option list verbatim.
7. **Auth type default depends on Grafana instance config.** The editor's `useEffect`
   (`ConnectionConfig.tsx:75-90`) picks `awsAllowedAuthProviders[0]`, and prefers
   `grafana_assume_role` if the feature is on. So the "default" auth type new users
   see is Grafana-instance-dependent, not always `default`. The stored schema default
   here is `default` (matching the reference pack and the backend iota zero), which
   may not be what a fresh Grafana Cloud editor writes.
8. **`sessionToken` has no editor UI.** `grafana-aws-sdk-react` never exposes an input
   for `sessionToken` even though the backend reads it and it is required for
   temporary credentials. Users must provision it directly.
9. **The "Default Region" description has padded backticks.** `ConnectionConfig.tsx:377`
   reads `` `Specify the region, such as for US West (Oregon) use ` us-west-2 ` as the
   region.` `` — note the spaces around `us-west-2` inside the backticks. Preserved
   verbatim in the schema; would render inline code with padding in some Markdown
   renderers.
10. **Editor Save-and-test hack that toggles state.** `ConfigEditor.tsx:22-37` tracks a
    local `saved` boolean that is set true after each `PUT /api/datasources/:id` so
    the ConfigSelect widgets can fetch resources against the saved credentials. There
    is no user-facing indication that saving is happening; it is triggered as a
    side-effect of interacting with each Timestream Details select.
11. **`Divider` renders unconditionally between ConnectionConfig and Timestream
    Details** (`ConfigEditor.tsx:87`) — a nice-to-have visual break but slightly at
    odds with other AWS plugins that use `<ConfigSubSection>` inside the same
    `<ConfigSection>` for hierarchy.

## Validation performed

- `dsconfig.ParseAndResolveSchemaJSON` + `Schema.Validate()` on `dsconfig.json` —
  passes.
- JSON Schema validation against [`dsconfig/schema.json`](../../dsconfig/schema.json)
  (draft-07, `additionalProperties: false`) — passes.
- `go test ./...` on this module — passes (schema bundle shape,
  `SchemaSpecHasNoSecureJSON`, `SecureValuesMatchLoadSettings`,
  `JSONDataMatchesStruct`, `JSONDataTypesMatchStruct`, `SchemaArtifactInSync`,
  `LoadConfig` including PascalCase decode, legacy `database → profile` fallback,
  `ApplyDefaults`, `Validate` per auth type, `EffectiveProfile`).
- `settings.go` / `schema.go`: `go build`, `go vet`, `gofmt` — clean.
- `settings.ts`: `tsc --noEmit --strict` — clean.
