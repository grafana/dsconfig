# grafana-redshift-datasource

Declarative configuration schema for the [Amazon Redshift datasource plugin](https://github.com/grafana/redshift-datasource) (`grafana-redshift-datasource`).

## Upstream researched

- **Repo**: `github.com/grafana/redshift-datasource`
- **Ref**: `main`
- **Commit SHA**: `5bb93760b4db87362c9aed3bf783f9c7c4344a60` (2026-06-16, `chore: make grafanaDependency prerelease-inclusive (#855)`)

Every value in [`dsconfig.json`](dsconfig.json) — labels, placeholders, tooltips, option
labels/values, section titles, defaults, validations, dependency and required-when
expressions, storage keys, storage targets, value types, group titles, and instructions —
is traceable to a specific `file:line` in the upstream plugin repo (or in the pinned
`@grafana/aws-sdk` version of the shared `ConnectionConfig` component) at this SHA. See
[Field provenance](#field-provenance) below.

To reproduce this research:

```bash
git clone --depth 20 https://github.com/grafana/redshift-datasource
cd redshift-datasource
git checkout 5bb93760b4db87362c9aed3bf783f9c7c4344a60

# The AWS auth surface is rendered by @grafana/aws-sdk's ConnectionConfig. Pin
# to the version redshift-datasource's package.json uses (0.10.2):
git clone https://github.com/grafana/grafana-aws-sdk-react
cd grafana-aws-sdk-react
git checkout v0.10.2
```

If upstream `main` has advanced past the pinned SHA, re-diff the sources listed under
[Sources researched](#sources-researched) and reconcile the schema before merging.

## Files

| File | Purpose |
| --- | --- |
| [`dsconfig.json`](dsconfig.json) | dsconfig v1 schema — single source of truth for all config fields, groups, and instructions |
| [`settings.ts`](settings.ts) | TypeScript models: `RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig` |
| [`settings.go`](settings.go) | Go `Config` model (flat: jsonData fields + `DecryptedSecureJSONData`), `PluginID`, `AWSAuthType` typed constants, `SecureJsonDataKey` typed constants, `ManagedSecret` nested struct, and the `LoadConfig` utility |
| [`schema.go`](schema.go) | k8s-style SDK plugin schema: embeds `dsconfig.json`, converts it via `dsconfig.NewSDKSchema`, and defines `SettingsExamples` for each (provisioning × credential mode) matrix quadrant plus each AWS auth provider |
| [`settings_test.go`](settings_test.go) | Table tests for `LoadConfig`, `ApplyDefaults`, and `Validate` |
| [`conformance_test.go`](conformance_test.go) | Runs `schema.RunPluginTests` — the shared dsconfig conformance suite — against `dsconfig.json` |
| [`schema.gen.json`](schema.gen.json), [`settings.gen.json`](settings.gen.json), [`settings.examples.gen.json`](settings.examples.gen.json) | Committed schema artifacts (regenerate with `go generate ./...` inside this directory; the `SchemaArtifactInSync` conformance subtest guards drift) |

There is no per-entry `go.mod` — every registry entry is a subpackage of the shared
[`registry/`](..) module (`github.com/grafana/dsconfig/registry`).

## Sources researched

Every source below was read at the pinned upstream SHA
(`5bb93760b4db87362c9aed3bf783f9c7c4344a60`) or, for `@grafana/aws-sdk`, at the exact
version pinned in the plugin's `package.json` (`0.10.2`).

### Plugin repo (`github.com/grafana/redshift-datasource@5bb9376`)

| File | What was read |
| --- | --- |
| `src/plugin.json:1-51` | `pluginType` (`id: "grafana-redshift-datasource"`), `pluginName` (`name: "Amazon Redshift"`), `docURL` |
| `src/ConfigEditor/ConfigEditor.tsx:1-394` | Editor layout: embedded `<ConnectionConfig>` (no `showHttpProxySettings` — the AWS proxy fields are NOT part of Redshift's surface), the `<ConfigSection title="Redshift Details">` block, the `AuthTypeSwitch` radio group writing `jsonData.useManagedSecret`, the `useServerless` Switch, the ClusterID / Workgroup / ManagedSecret ConfigSelects, the Database User / Database Inputs, and the `withEvent` Switch |
| `src/ConfigEditor/ConfigEditor.tsx:63-73` | `useManagedSecret` state hookup: writes `jsonData.useManagedSecret` on radio change |
| `src/ConfigEditor/ConfigEditor.tsx:85-118` | `useEffect` on `managedSecret.arn`: fetches the secret, rewrites `jsonData.clusterIdentifier` (or `jsonData.workgroupName` when Serverless) and `jsonData.dbUser` from the fetched secret |
| `src/ConfigEditor/ConfigEditor.tsx:152-172` | `getClusterUrl` / `getWorkgroupUrl`: derive root `url` = `${endpoint}/${database}` (display convenience only) |
| `src/ConfigEditor/ConfigEditor.tsx:189-199` | `onChangeManagedSecret` writes both `managedSecret.arn` (from Select value) and `managedSecret.name` (from Select label) |
| `src/ConfigEditor/ConfigEditor.tsx:201-227` | `onChangeClusterID` / `onChangeWorkgroupName` also rewrite root `url` from Select's `description` (`address:port`) |
| `src/ConfigEditor/ConfigEditor.tsx:243-388` | JSX: field composition, visibility gates (`hidden={props.options.jsonData.useServerless}`, `hidden={!useManagedSecret}`, etc.) |
| `src/ConfigEditor/AuthTypeSwitch.tsx:1-70` | `RadioButtonGroup` with `{label:"Temporary credentials",value:false}` and `{label:"AWS Secrets Manager",value:true}` — the two option labels are the entire "label" surface for the useManagedSecret field |
| `src/selectors.ts:1-81` | Field labels via the E2E selectors registry: `UseServerless.input="Serverless"`, `ManagedSecret.input="Managed Secret"`, `Workgroup.input="Workgroup"`, `ClusterID.input="Cluster Identifier"`, `Database.input="Database"`, `DatabaseUser.input="Database User"`, `WithEvent.input="Send events to Amazon EventBridge"` |
| `src/types.ts:51-64` | `RedshiftDataSourceOptions extends AwsAuthDataSourceJsonData` adds `withEvent`, `useManagedSecret`, `useServerless`, `workgroupName`, `clusterIdentifier`, `database`, `dbUser`, `managedSecret: {name, arn}`, `enableSecureSocksProxy` (excluded per AGENTS.md) |
| `src/types.ts:69` | `RedshiftDataSourceSecureJsonData extends AwsAuthDataSourceSecureJsonData` — no plugin-specific secure keys, only the AWS-shared ones |
| `pkg/redshift/models/settings.go:14-52` | Backend `RedshiftDataSourceSettings` embeds `awsds.AWSDatasourceSettings` and adds `ClusterIdentifier`, `WorkgroupName`, `Database`, `UseServerless`, `UseManagedSecret`, `WithEvent`, `DBUser`, and (untagged!) `ManagedSecret` |
| `pkg/redshift/models/settings.go:58-72` | `Load`: unmarshals jsonData when `len > 1`, copies `accessKey` / `secretKey` / `sessionToken` from decrypted secure data. Mirror this control-flow verbatim in `LoadConfig`. |
| `pkg/redshift/models/settings.go:74-87` | `Apply`: query-time overrides for `region` and `database` from `sqlds.Options` — not part of stored config |
| `package.json:23-59` | External component versions (see next table) |
| `go.mod:1-16` | `github.com/grafana/grafana-aws-sdk v1.4.3`, `github.com/grafana/grafana-plugin-sdk-go v0.291.1` |

### External editor components

Read at the exact versions pinned in the plugin's `package.json`.

| Component | Version | Source consulted | What was read |
| --- | --- | --- | --- |
| `ConnectionConfig`, `AwsAuthType`, `Divider`, `ConfigSelect` | `@grafana/aws-sdk@0.10.2` | grafana/grafana-aws-sdk-react tag `v0.10.2`, `src/components/ConnectionConfig.tsx`, `src/providers.ts`, `src/types.ts`, `src/sql/ConfigEditor/ConfigSelect.tsx`, `src/regions.ts` | Every AWS field's label, placeholder, description, and conditional render; the Select option labels for `authType`; the standard region list; ConfigSelect being a Select that disables itself until `defaultRegion` is set |
| `ConfigSection` | `@grafana/plugin-ui@^0.13.0` | Editor layout — no storage fields |
| `Field`, `Input`, `Switch`, `SecureSocksProxySettings` (excluded) | `@grafana/ui@12.4.2` | Prop names and rendering shape for the plugin-specific fields |
| `SelectableValue`, `DataSourcePluginOptionsEditorProps` | `@grafana/data@12.4.2` | Storage-key semantics of the editor's update helpers |

### Backend Go dependency (`grafana-aws-sdk` `v1.4.3`)

| File | What was read |
| --- | --- |
| `pkg/awsds/settings.go:13-91` | `AuthType` int enum + custom `MarshalJSON`/`UnmarshalJSON` — the storage⇆enum mapping (including legacy `arn`→`default` and `sharedCreds`→`credentials`) that we surface as `AWSAuthType` string constants |
| `pkg/awsds/settings.go:94-117` | `AWSDatasourceSettings` struct with the AWS-shared fields; note `AssumeRoleARN string \`json:"assumeRoleARN"\`` (uppercase RN) versus the frontend's camelCase `assumeRoleArn` — case-insensitive Unmarshal makes both work |
| `pkg/awsds/settings.go:120-141` | `Load` copies decrypted `accessKey`, `secretKey`, `sessionToken`, `proxyPassword` from secure JSON data; mirrors `defaultRegion`→`region` |

## Field provenance

Every dsconfig field, traced from its schema `id` to the upstream `file:line` where each
of its label, placeholder, tooltip, default, storage key, and value type is defined.

| Schema `id` | Storage key | Target | Label source | Placeholder / options / default source | Value type source | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `jsonData_authType` | `authType` | `jsonData` | `ConnectionConfig.tsx:105` (`<Field label="Authentication Provider">`) | Options `providers.ts:4-24`; description `ConnectionConfig.tsx:106`; default `awsds/settings.go:16` (`AuthTypeDefault` iota) → `"default"` | `AWSDatasourceSettings.AuthType` `awsds/settings.go:97` (int enum, string on wire) | Role `auth.discriminator`; `allowedValues` includes legacy `arn` |
| `jsonData_profile` | `profile` | `jsonData` | `ConnectionConfig.tsx:123` (`<Field label="Credentials Profile Name">`) | Placeholder `:129` (`"default"`); description `:124` | `AWSDatasourceSettings.Profile` `awsds/settings.go:95` | `dependsOn` from conditional render `ConnectionConfig.tsx:121` |
| `secureJsonData_accessKey` | `accessKey` | `secureJsonData` | `ConnectionConfig.tsx:137` (`<Field label="Access Key ID">`) | — | `AWSDatasourceSettings.AccessKey` `awsds/settings.go:113` | Role `auth.aws.accessKeyId`; conditional on `authType == 'keys'` |
| `secureJsonData_secretKey` | `secretKey` | `secureJsonData` | `ConnectionConfig.tsx:157` (`<Field label="Secret Access Key">`) | — | `AWSDatasourceSettings.SecretKey` `awsds/settings.go:114` | Role `auth.aws.secretAccessKey`; conditional on `authType == 'keys'` |
| `secureJsonData_sessionToken` | `sessionToken` | `secureJsonData` | — (no UI) | — | `AWSDatasourceSettings.SessionToken` `awsds/settings.go:115` | Role `auth.aws.sessionToken`; tagged `backend-only`; still read at `pkg/redshift/models/settings.go:67` |
| `jsonData_assumeRoleArn` | `assumeRoleArn` | `jsonData` | `ConnectionConfig.tsx:261` (`<Field ... label="Assume Role ARN">`) | Placeholder `:268` (`"arn:aws:iam:*"`); description `:262-264` | `AWSDatasourceSettings.AssumeRoleARN` `awsds/settings.go:98` (backend json tag `assumeRoleARN`) | `dependsOn` from conditional render (not `grafana_assume_role`); IAM role pattern validation |
| `jsonData_externalId` | `externalId` | `jsonData` | `ConnectionConfig.tsx:276` (`<Field ... label="External ID">`) | Placeholder `:281`; description `:277` | `AWSDatasourceSettings.ExternalID` `awsds/settings.go:99` | `dependsOn` from conditional render (not `grafana_assume_role`) |
| `jsonData_endpoint` | `endpoint` | `jsonData` | `ConnectionConfig.tsx:362` (`<Field label="Endpoint">`) | Placeholder `:368`; description `:363` | `AWSDatasourceSettings.Endpoint` `awsds/settings.go:102` | `dependsOn` from conditional render (not `grafana_assume_role`) |
| `jsonData_defaultRegion` | `defaultRegion` | `jsonData` | `ConnectionConfig.tsx:376` (`<Field label="Default Region">`) | Description `:377` (verbatim, including the padded backticks `` ` us-west-2 ` ``) | `AWSDatasourceSettings.DefaultRegion` `awsds/settings.go:105` | `<Select ... allowCustomValue={true}>` |
| `jsonData_useManagedSecret` | `useManagedSecret` | `jsonData` | (no field label — the RadioButtonGroup labels its two options directly) | Options `AuthTypeSwitch.tsx:60-63` (`"Temporary credentials"` value=false / `"AWS Secrets Manager"` value=true); default `false` (React state seeded from `!!jsonData.useManagedSecret`) | `RedshiftDataSourceSettings.UseManagedSecret bool` `pkg/redshift/models/settings.go:48` | Editor description is conditional (`AuthTypeSwitch.tsx:16-55`) so the field carries no static description |
| `jsonData_useServerless` | `useServerless` | `jsonData` | `selectors.ts:19` (`UseServerless.input="Serverless"`, referenced at `ConfigEditor.tsx:253`) | `<Switch>` (`ConfigEditor.tsx:257-271`); default `false` (backend zero) | `RedshiftDataSourceSettings.UseServerless bool` `pkg/redshift/models/settings.go:47` | Provisioning shape discriminator |
| `jsonData_clusterIdentifier` | `clusterIdentifier` | `jsonData` | `selectors.ts:34` (`ClusterID.input="Cluster Identifier"`, also `ClusterIDText.input`) | ConfigSelect populated at runtime from `/resources/clusters` (`ConfigEditor.tsx:120-134`); allows custom values when read-only from Secrets Manager | `RedshiftDataSourceSettings.ClusterIdentifier` `pkg/redshift/models/settings.go:44` | `dependsOn`/`requiredWhen` from `hidden={props.options.jsonData.useServerless}` (`:276`, `:294`) |
| `jsonData_workgroupName` | `workgroupName` | `jsonData` | `selectors.ts:30` (`WorkgroupText.input="Workgroup"`, referenced at `ConfigEditor.tsx:312`) | ConfigSelect populated at runtime from `/resources/workgroups` (`ConfigEditor.tsx:136-150`) | `RedshiftDataSourceSettings.WorkgroupName` `pkg/redshift/models/settings.go:45` | `dependsOn`/`requiredWhen` from `hidden={!props.options.jsonData.useServerless}` (`:311`) |
| `jsonData_managedSecret_arn` | `arn` (in `managedSecret`) | `jsonData` | `selectors.ts:22` (`ManagedSecret.input="Managed Secret"`, referenced at `ConfigEditor.tsx:327`) | ConfigSelect populated at runtime from `/resources/secrets` (`ConfigEditor.tsx:76-79`); Select value is the ARN (`:78`), label is the secret name | `RedshiftDataSourceOptions.managedSecret.arn` `src/types.ts:59-62`; backend `ManagedSecret.ARN` `pkg/redshift/models/settings.go:16` | `dependsOn`/`requiredWhen` from `hidden={!useManagedSecret}` (`:328`); modelled as nested object via `section: "managedSecret"` |
| `jsonData_managedSecret_name` | `name` (in `managedSecret`) | `jsonData` | — (no UI — populated from Select label) | Set from Select label in `onChangeManagedSecret` (`ConfigEditor.tsx:192`) | `RedshiftDataSourceOptions.managedSecret.name` `src/types.ts:60`; backend `ManagedSecret.Name` `pkg/redshift/models/settings.go:15` | Tagged `managed-by:jsonData_managedSecret_arn` — comes along for the ride when the ARN is chosen |
| `jsonData_dbUser` | `dbUser` | `jsonData` | `selectors.ts:46` (`DatabaseUser.input="Database User"`, referenced at `ConfigEditor.tsx:345`) | `<Input>` (`:350-356`); disabled when `useManagedSecret` (populated from the fetched secret) | `RedshiftDataSourceSettings.DBUser` `pkg/redshift/models/settings.go:50` | `dependsOn` from `hidden={useServerless && !useManagedSecret}` (`:346`); `requiredWhen` from backend contract: Provisioned temp-creds needs it, Serverless GetCredentials mints its own |
| `jsonData_database` | `database` | `jsonData` | `selectors.ts:42` (`Database.input="Database"`, referenced at `ConfigEditor.tsx:359`) | `<Input>` (`:360-366`) | `RedshiftDataSourceSettings.Database` `pkg/redshift/models/settings.go:46` | Always required — backend needs it to construct Redshift Data API requests |
| `jsonData_withEvent` | `withEvent` | `jsonData` | `selectors.ts:62` (`WithEvent.input="Send events to Amazon EventBridge"`, referenced at `ConfigEditor.tsx:368`) | `<Switch>` (`:369-383`); default `false` | `RedshiftDataSourceSettings.WithEvent bool` `pkg/redshift/models/settings.go:49` | Optional |

## Field inventory summary

| Schema field | Storage key | Target | Editor label | Read by backend? |
| --- | --- | --- | --- | --- |
| `jsonData_authType` | `authType` | `jsonData` | Authentication Provider | Yes (awsds) |
| `jsonData_profile` | `profile` | `jsonData` | Credentials Profile Name | Yes (awsds) |
| `secureJsonData_accessKey` | `accessKey` | `secureJsonData` | Access Key ID | Yes (awsds) |
| `secureJsonData_secretKey` | `secretKey` | `secureJsonData` | Secret Access Key | Yes (awsds) |
| `secureJsonData_sessionToken` | `sessionToken` | `secureJsonData` | — (no UI) | Yes (backend-only) |
| `jsonData_assumeRoleArn` | `assumeRoleArn` | `jsonData` | Assume Role ARN | Yes (awsds) |
| `jsonData_externalId` | `externalId` | `jsonData` | External ID | Yes (awsds) |
| `jsonData_endpoint` | `endpoint` | `jsonData` | Endpoint | Yes (awsds) |
| `jsonData_defaultRegion` | `defaultRegion` | `jsonData` | Default Region | Yes (awsds → Region) |
| `jsonData_useManagedSecret` | `useManagedSecret` | `jsonData` | — (radio group options) | Yes |
| `jsonData_useServerless` | `useServerless` | `jsonData` | Serverless | Yes |
| `jsonData_clusterIdentifier` | `clusterIdentifier` | `jsonData` | Cluster Identifier | Yes |
| `jsonData_workgroupName` | `workgroupName` | `jsonData` | Workgroup | Yes |
| `jsonData_managedSecret_arn` | `managedSecret.arn` | `jsonData` | Managed Secret | Yes |
| `jsonData_managedSecret_name` | `managedSecret.name` | `jsonData` | — (no UI) | Yes |
| `jsonData_dbUser` | `dbUser` | `jsonData` | Database User | Yes |
| `jsonData_database` | `database` | `jsonData` | Database | Yes |
| `jsonData_withEvent` | `withEvent` | `jsonData` | Send events to Amazon EventBridge | Yes |

### Frontend-only settings

None. Every editor-visible field is read by either the AWS SDK or the Redshift plugin
backend.

### Backend-only settings

- **`sessionToken`** — no editor UI; still read by `AWSDatasourceSettings.Load` at
  `awsds/settings.go:137` and copied through by `pkg/redshift/models/settings.go:67`.
  Used with `authType: "keys"` for temporary STS credentials. Provisioning is the
  practical way to set it.

### Fields excluded from this entry

- **AWS proxy fields** (`proxyType`, `proxyUrl`, `proxyUsername`, `proxyPassword`) —
  `ConnectionConfig.tsx:291` only renders the proxy subsection when the caller passes
  `showHttpProxySettings`. Redshift's `ConfigEditor.tsx:245` does not, so the proxy
  fields are neither editor-visible nor part of Redshift's declared surface.
- **Secure Socks Proxy** (`jsonData.enableSecureSocksProxy`) — excluded per AGENTS.md.
  Declared in `src/types.ts:63` and rendered in `ConfigEditor.tsx:246-248` when the
  Grafana runtime flag is on.
- **`region`** — the AWS SDK Go struct has a `Region` field
  (`awsds/settings.go:96`), but the frontend never writes it; the backend copies
  `defaultRegion` into it at load time.

## Where the types are defined

### Frontend (TypeScript)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `RedshiftDataSourceOptions` (jsonData), `RedshiftDataSourceSecureJsonData` | `src/types.ts:51-69` | plugin ([grafana/redshift-datasource](https://github.com/grafana/redshift-datasource)) |
| `RedshiftManagedSecret` (frontend shape mirrored under `jsonData.managedSecret`) | `src/types.ts:35-38` | plugin |
| `AwsAuthType`, `AwsAuthDataSourceJsonData` (base of `RedshiftDataSourceOptions`), `AwsAuthDataSourceSecureJsonData` (base of `RedshiftDataSourceSecureJsonData`) | `src/types.ts:3-32` | `@grafana/aws-sdk` `0.10.2` |
| `awsAuthProviderOptions` (Select options for `authType`) | `src/providers.ts:4-24` | `@grafana/aws-sdk` `0.10.2` |
| `ConnectionConfig` React component (renders every AWS-shared field) | `src/components/ConnectionConfig.tsx:36-404` | `@grafana/aws-sdk` `0.10.2` |
| `ConfigSelect` (used for cluster/workgroup/managedSecret) | `src/sql/ConfigEditor/ConfigSelect.tsx:45-97` | `@grafana/aws-sdk` `0.10.2` |
| `ConfigSection` | `src/components/ConfigEditor/` | `@grafana/plugin-ui` `^0.13.0` |
| `Field`, `Input`, `Switch`, `RadioButtonGroup`, `SecureSocksProxySettings` | `packages/grafana-ui/src/components/*` | `@grafana/ui` `12.4.2` |
| `DataSourcePluginOptionsEditorProps` | `packages/grafana-data/src/types/datasource.ts` | `@grafana/data` `12.4.2` |

### Backend (Go)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `RedshiftDataSourceSettings`, `Load`, `Apply`, `ManagedSecret`, `RedshiftSecret`, `RedshiftEndpoint`, `RedshiftCluster`, `RedshiftWorkgroup` | `pkg/redshift/models/settings.go:14-87` | plugin ([grafana/redshift-datasource](https://github.com/grafana/redshift-datasource)) |
| `AWSDatasourceSettings` (embedded base of `RedshiftDataSourceSettings`), `AuthType` int enum + custom Marshal/Unmarshal | `pkg/awsds/settings.go:13-141` | `github.com/grafana/grafana-aws-sdk` `v1.4.3` |
| `backend.DataSourceInstanceSettings` (carries `JSONData`, `DecryptedSecureJSONData`) | `backend/common.go` | `github.com/grafana/grafana-plugin-sdk-go` `v0.291.1` |
| Query layer (Redshift Data API driver, macros, resource routes) | `pkg/redshift/driver/*.go`, `pkg/redshift/routes/*.go` | plugin |
| AWS SDK v2 clients | `github.com/aws/aws-sdk-go-v2/service/{redshift,redshiftdata,redshiftserverless,secretsmanager}` | `github.com/aws/aws-sdk-go-v2` |

The models in this entry flatten that spread into a single Go `Config` type (jsonData
fields + `DecryptedSecureJSONData`) with a nested `ManagedSecret` sub-struct, plus a
`SecureJsonDataKey` typed constant list. `settings.ts` keeps the three canonical
TypeScript types (`RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig`).

## Modeling decisions

- **Single, flat Go `Config`.** The upstream `RedshiftDataSourceSettings` embeds
  `awsds.AWSDatasourceSettings`. We flatten those into one struct to match the pattern
  used by other dsconfig registry entries (see `registry/grafana-athena-datasource`,
  `registry/cloudwatch`) and to avoid pulling `grafana-aws-sdk` into the shared
  registry `go.mod`.
- **`managedSecret` modelled as a nested object via `section`.** The two subfields
  (`arn`, `name`) share a common editor gate (`useManagedSecret == true`) and are
  written together by a single Select. Using `section: "managedSecret"` lets the
  conformance suite recurse into the nested `ManagedSecret` Go sub-struct and match
  the schema's dotted paths against `managedSecret.arn` / `managedSecret.name`.
- **`ManagedSecret` gets an explicit `json:"managedSecret"` tag.** The upstream
  backend field is `ManagedSecret ManagedSecret` with **no json tag**, which would
  serialise as PascalCase `ManagedSecret`. We tag it camelCase because that is what
  the editor actually writes (`ConfigEditor.tsx:196`); Go's case-insensitive
  Unmarshal makes both spellings load correctly, and the
  `pascalcase ManagedSecret is accepted` LoadConfig test locks that in.
- **`useManagedSecret` has no field label.** `AuthTypeSwitch.tsx` renders it as a
  bare `RadioButtonGroup` inside a `<Label>` that carries only a conditional
  description; the RadioButtonGroup itself has no label text, so we set none.
  The two option labels (`"Temporary credentials"` / `"AWS Secrets Manager"`) are
  the entire label surface.
- **`useManagedSecret` and `useServerless` defaults preserved as `false`.** Both
  React states are seeded with `!!jsonData.useManagedSecret` / the raw
  `jsonData.useServerless` — the editor default is a plain `false`. We keep the
  schema `defaultValue: false` so the "" example carries both explicitly.
- **`requiredWhen` mirrors the backend contract, not the UI required marker.** The
  Redshift editor doesn't mark cluster / workgroup / dbUser / database as required,
  but `pkg/redshift/models/settings.go` and the query builder need them to actually
  construct a Redshift Data API request. We express those as `requiredWhen`
  expressions gated by the mode toggles.
- **`clusterIdentifier` allows custom values but `workgroupName` does not.** The
  editor's ClusterID `ConfigSelect` sets `allowCustomValue={true}`
  (`ConfigEditor.tsx:301`) so users can type a value when DescribeClusters is
  forbidden; the Workgroup `ConfigSelect` doesn't. We mirror that with
  `ui.allowCustom` on `clusterIdentifier` only.
- **Root `url` is not modelled.** The editor rewrites `props.options.url` from the
  cluster/workgroup endpoint plus `database` as a display convenience
  (`ConfigEditor.tsx:152-172`), but the backend never reads it. Provisioned configs
  can safely omit it.
- **`grafana_assume_role` is a schema-only value** — the editor only renders it
  when the `awsDatasourcesTempCredentials` feature toggle is on and the plugin is
  in the allow-list (`ConnectionConfig.tsx:18-28,53-64`). We list it as a schema
  `allowedValues` entry, mark it as visible-when in the depends-on expression of
  assumeRoleArn / externalId / endpoint, and note the gating in an instruction.
- **`arn` kept in `allowedValues`, not in UI options** — deprecated; provisioned
  datasources may still carry it. Listing it in `allowedValues` but not in the
  Select's `options` matches the backend's tolerance.
- **`sessionToken` included as a secure key with no UI** — the plugin doesn't
  offer a UI field, but `pkg/redshift/models/settings.go:67` reads it from
  decrypted secure data. Tagged `backend-only`.
- **AWS proxy fields excluded** — Redshift does not pass `showHttpProxySettings`
  to `ConnectionConfig`, so the proxy fields are not part of Redshift's editor
  surface. Consumers needing them should reach for the shared
  `aws_sdk_settings.json` pack directly.
- **`SecureJsonDataConfig` is a key list** — secure values are write-only, so the
  type is just the array of secret key names (`accessKey`, `secretKey`,
  `sessionToken`); consumers read `secureJsonFields` to see what is configured.

## SDK plugin schema and k8s-style examples (`schema.go`)

`NewSchema()` assembles the `grafana-plugin-sdk-go` `pluginschema.PluginSchema` bundle
(the k8s-style schema Grafana's datasource API server serves as `{apiVersion}.json`,
`v0alpha1` today) from the embedded `dsconfig.json`: jsonData fields (including the
`managedSecret` nested object) become the OpenAPI settings `spec`, secure fields become
`secureValues`.

`SettingsExamples()` covers the default configuration and every quadrant of the
(provisioning shape × credential mode) matrix, plus each AWS auth provider variant:

| Example | Shape | Credential mode | AWS auth | `secureJsonData` |
| --- | --- | --- | --- | --- |
| `""` (default) | — | — | AWS SDK Default | `accessKey` (empty) |
| `provisionedTempCredsKeys` | Provisioned | Temporary IAM creds | Access & secret key | `accessKey`, `secretKey` |
| `provisionedManagedSecretDefault` | Provisioned | Secrets Manager | AWS SDK Default | `accessKey` (empty) |
| `serverlessTempCredsIamRole` | Serverless | Temporary IAM creds | Workspace IAM Role | `accessKey` (empty) |
| `serverlessManagedSecretGrafanaAssume` | Serverless | Secrets Manager | Grafana Assume Role | `accessKey` (empty) |
| `credentialsFile` | Provisioned | Temporary IAM creds | Credentials file | `accessKey` (empty) |
| `assumeRoleFromKeys` | Provisioned | Temporary IAM creds | Keys + STS AssumeRole | `accessKey`, `secretKey` |
| `withEventBridge` | Provisioned | Temporary IAM creds | Access & secret key + `withEvent=true` | `accessKey`, `secretKey` |
| `legacyArnAuthType` | Provisioned | Temporary IAM creds | `arn` (legacy) | `accessKey` (empty) |

Every example carries at least one `secureJsonData` placeholder as required by the
conformance suite.

## `LoadConfig` utility (`settings.go`)

`LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error)`
runs the full three-phase load flow on a datasource instance's settings and returns a
fully-defaulted, validated `Config`:

1. **Parse** — unmarshal `settings.JSONData` into `Config` when `len > 1` (the exact
   check the upstream `Load` uses at `pkg/redshift/models/settings.go:59`), then copy
   the plugin's decrypted secrets (`accessKey`/`secretKey`/`sessionToken`) into
   `DecryptedSecureJSONData`. Case-insensitive `encoding/json` accepts either the
   frontend's camelCase `managedSecret` or the upstream backend's untagged PascalCase
   `ManagedSecret`.
2. **`ApplyDefaults`** — fills the single curated default: `AuthType` defaults to
   `AWSAuthTypeDefault`, matching the reference `aws_sdk_settings.json` pack and the
   backend `awsds.AuthTypeDefault` (iota zero). `useServerless` /
   `useManagedSecret` intentionally have no default because their zero value (false)
   IS the editor default (Provisioned + Temporary credentials) and would be
   indistinguishable from an unset field.
3. **`Validate`** — enforces the runtime contract: known `AuthType`; `accessKey`+
   `secretKey` present for `keys` auth; non-empty `defaultRegion`; the correct
   identifier for the provisioning shape (`clusterIdentifier` when
   `useServerless==false`, `workgroupName` when true); the correct credential input
   (`managedSecret.arn` for Secrets Manager, `dbUser` for Provisioned temp-creds;
   Serverless temp-creds does not require `dbUser` because `GetCredentials` mints
   the username itself); non-empty `database`. Errors are joined so callers see
   every problem at once.

Everything is logged via `backend.Logger.FromContext(ctx)` with `datasource_uid`,
`datasource_name`, and `plugin` labels so log lines carry request context.

### Direct access to individual phases

`(*Config).ApplyDefaults()` and `(Config).Validate() error` are exported for callers
that want to compose them themselves (provisioning preview, schema-example round-trip,
tests that need to distinguish parse-level from policy-level errors).

## Upstream findings

Potential bugs, misleading UX, and consistency issues discovered while researching
upstream. All preserved verbatim in the schema — the schema records what the plugin
**does**, not what it **should** do; these notes exist so reviewers can reproduce each
finding and decide separately whether to fix upstream.

1. **`ManagedSecret` field has no json tag on the backend struct.**
   `pkg/redshift/models/settings.go:51` declares `ManagedSecret ManagedSecret` with no
   json tag, so Go's `encoding/json` would emit `ManagedSecret` (PascalCase) but the
   frontend writes `managedSecret` (camelCase). Case-insensitive Unmarshal saves both
   at load, and no code path marshals settings back out, so this is a latent trap
   rather than a live bug — a stricter JSON decoder in another language would break
   every existing datasource.
2. **`AssumeRoleARN` json tag disagrees with the frontend's `assumeRoleArn`.**
   `awsds/settings.go:98` uses `` json:"assumeRoleARN" `` (uppercase RN), but the
   frontend writes `assumeRoleArn` (lowercase arn). Same case-insensitive-match
   rescue as ManagedSecret.
3. **Deprecated `arn` auth value silently maps to `default`.** `awsds/settings.go:87-88`
   maps any unknown auth type string to `AuthTypeDefault`. A provisioned config with
   `authType: "arn"` will load as if it were `default`, with no warning surfaced.
   Unlike CloudWatch (which shows an `ARN_DEPRECATION_WARNING_MESSAGE` banner),
   Redshift's editor does not warn the user.
4. **Legacy `sharedCreds` auth value silently mapped to `credentials`.**
   `awsds/settings.go:75-78` folds both storage values onto the same enum. Not
   preserved in this schema as an allowed value; a datasource stored with `authType:
   "sharedCreds"` would fail the schema's `allowedValues` check but still load fine
   on the backend.
5. **Frontend rewrites root `url` but backend never reads it.** `ConfigEditor.tsx:
   152-172` derives the datasource `url` from the cluster/workgroup endpoint plus
   `database` as a display value; `pkg/redshift/models/settings.go` never touches
   `settings.URL`. Provisioning payloads can safely omit `url`; if it's populated
   from a stored datasource, its value depends on frontend state at last save.
6. **`clusterIdentifier` and `workgroupName` can coexist in storage.** The editor
   never clears the *other* mode's identifier when the user toggles `useServerless`,
   so a provisioned Redshift datasource may carry both fields. The backend only
   reads the one the current `useServerless` flag selects; provisioning should still
   clear the unused one for tidiness.
7. **`managedSecret` isn't cleared when `useManagedSecret` becomes false.**
   Similar to (6): flipping the radio back to Temporary credentials leaves the
   `managedSecret` object in storage. Harmless because the backend gates on
   `useManagedSecret`, but a leftover ARN may confuse operators reading raw
   provisioned config.
8. **The "Grafana Assume Role" provider only appears when a feature toggle is on
   and the plugin is in an allow-list** (`ConnectionConfig.tsx:18-28,53-64`).
   Storage-side the value is valid regardless.
9. **`sessionToken` has no editor UI.** `grafana-aws-sdk-react` never exposes an
   input for `sessionToken`. Provisioning is the only way to set it. Same as every
   other AWS datasource.
10. **The "Default Region" description has padded backticks.** `ConnectionConfig.tsx:
    377` reads `` `Specify the region, such as for US West (Oregon) use ` us-west-2 `
    as the region.` `` — note the spaces around `us-west-2` inside the backticks.
    Preserved verbatim in the schema.
11. **`Load` gates on `len(settings.JSONData) > 1` instead of `> 0`.**
    `pkg/redshift/models/settings.go:59` skips json.Unmarshal when the body is 0 or
    1 bytes, so a 1-byte body (e.g. `{`) is silently accepted. Mirrored verbatim in
    LoadConfig; the malformed-jsonData test uses a longer body to actually trip the
    parser.

## Validation performed

- `dsconfig.ParseAndResolveSchemaJSON` + `Schema.Validate()` — passes.
- JSON Schema validation against [`dsconfig/schema.json`](../../dsconfig/schema.json)
  — passes (via `RunPluginTests` conformance suite).
- `go test ./...` on this entry — passes (schema bundle shape,
  `SchemaSpecHasNoSecureJSON`, `SecureValuesMatchLoadSettings`,
  `JSONDataMatchesStruct` including the `managedSecret.arn`/`managedSecret.name`
  nested paths, `JSONDataTypesMatchStruct`, `SchemaArtifactInSync`, `LoadConfig`
  including case-insensitive PascalCase `ManagedSecret` decode, `ApplyDefaults`,
  `Validate` per (shape × credential mode × auth type) combination).
- `settings.go` / `schema.go`: `go build`, `go vet`, `gofmt` — clean.
- All existing registry entries continue to build and test clean.
