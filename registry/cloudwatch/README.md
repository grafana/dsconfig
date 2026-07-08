# cloudwatch

Declarative configuration schema for the [Amazon CloudWatch datasource plugin](https://github.com/grafana/grafana-cloudwatch-datasource) (plugin id `cloudwatch`).

## Upstream researched

- **Repo**: `github.com/grafana/grafana-cloudwatch-datasource`
- **Ref**: `main`
- **Commit SHA**: `6e21d10b2d3a65ac140d06b822c64af4617190eb` (2026-07-02, `docs: add signed commits requirement to CONTRIBUTING.md (#574)`)

Every value in [`dsconfig.json`](dsconfig.json) — labels, placeholders, tooltips, option
labels/values, section titles, defaults, validations, dependency and required-when
expressions, storage keys, storage targets, value types, group titles, and instructions —
is traceable to a specific `file:line` in the upstream plugin repo (or in the pinned
`@grafana/aws-sdk` version of the shared `ConnectionConfig` component) at this SHA. See
[Field provenance](#field-provenance) below.

To reproduce this research:

```bash
git clone --depth 20 https://github.com/grafana/grafana-cloudwatch-datasource
cd grafana-cloudwatch-datasource
git checkout 6e21d10b2d3a65ac140d06b822c64af4617190eb

# The AWS auth surface is rendered by @grafana/aws-sdk's ConnectionConfig. Pin
# to the version grafana-cloudwatch-datasource's package.json uses (0.10.2):
git clone https://github.com/grafana/grafana-aws-sdk-react
cd grafana-aws-sdk-react
git checkout v0.10.2   # SHA fe0c4d8d657ee5ed053ae173293dc876619b5a2b

# And the backend Go SDK version pinned in go.mod (v1.4.4):
git clone https://github.com/grafana/grafana-aws-sdk
cd grafana-aws-sdk
git checkout v1.4.4   # SHA c9f152cdd1a2e686d8f76f003a4595efd418e94c
```

If upstream `main` has advanced past the pinned SHA, re-diff the sources listed under
[Sources researched](#sources-researched) and reconcile the schema before merging.

## Files

| File | Purpose |
| --- | --- |
| [`dsconfig.json`](dsconfig.json) | dsconfig v1 schema — single source of truth for all config fields, groups, and instructions |
| [`settings.ts`](settings.ts) | TypeScript models: `RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig` |
| [`settings.go`](settings.go) | Go `Config` model (flat: jsonData fields + `DecryptedSecureJSONData`), `PluginID`, `AWSAuthType` / `AWSProxyType` typed constants, `SecureJsonDataKey` typed constants, the plugin-mirror `Duration` type, and the `LoadConfig` utility |
| [`schema.go`](schema.go) | k8s-style SDK plugin schema: embeds `dsconfig.json`, converts it via `dsconfig.NewSDKSchema`, and defines `SettingsExamples` for each AWS auth provider plus CloudWatch Logs / URL-proxy / legacy variants |
| [`settings_test.go`](settings_test.go) | Table tests for `LoadConfig`, `ApplyDefaults`, `Validate`, and the custom `Duration` UnmarshalJSON |
| [`conformance_test.go`](conformance_test.go) | Runs `schema.RunPluginTests` — the shared dsconfig conformance suite — against `dsconfig.json` |
| [`schema.gen.json`](schema.gen.json), [`settings.gen.json`](settings.gen.json), [`settings.examples.gen.json`](settings.examples.gen.json) | Committed schema artifacts (regenerate with `go generate ./...` inside this directory; the `SchemaArtifactInSync` conformance subtest guards drift) |

There is no per-entry `go.mod` — every registry entry is a subpackage of the shared
[`registry/`](..) module (`github.com/grafana/dsconfig/registry`).

## Sources researched

Every source below was read at the pinned upstream SHA
(`6e21d10b2d3a65ac140d06b822c64af4617190eb`) or, for `@grafana/aws-sdk`, at the exact
version pinned in the plugin's `package.json` (`0.10.2` / SHA `fe0c4d8`), and, for
`grafana-aws-sdk` (Go), the version pinned in `go.mod` (`v1.4.4` / SHA `c9f152c`).

### Plugin repo (`github.com/grafana/grafana-cloudwatch-datasource@6e21d10`)

| File | What was read |
| --- | --- |
| `src/plugin.json:1-70` | `pluginType` (`id: "cloudwatch"`), `pluginName` (`name: "CloudWatch"`), `docURL` (`info.links[1].url`) |
| `src/components/ConfigEditor/ConfigEditor.tsx:36-187` | Editor layout: DataSourceDescription intro, embedded `<ConnectionConfig showHttpProxySettings ...>`, the "Namespaces of Custom Metrics" `<Field>` as its child, the conditional `<SecureSocksProxySettingsNewStyling>` (excluded), the "Cloudwatch Logs" `<ConfigSection>` with `logsTimeout` and `Default Log Groups`, and the `<XrayLinkConfig>` block |
| `src/components/ConfigEditor/ConfigEditor.tsx:29-34` | Editor warning message strings (`ARN_DEPRECATION_WARNING_MESSAGE`, `CREDENTIALS_AUTHENTICATION_WARNING_MESSAGE`) surfaced when `authType === "arn"` or credentials-without-profile |
| `src/components/ConfigEditor/ConfigEditor.tsx:102-108` | "Namespaces of Custom Metrics" `<Field>` and `<Input>` with placeholder `Namespace1,Namespace2` (no description) |
| `src/components/ConfigEditor/ConfigEditor.tsx:114-129` | "Query Result Timeout" `<Field label>` and description text for `jsonData.logsTimeout`, `<Input>` with placeholder `30m` and `width={60}` |
| `src/components/ConfigEditor/ConfigEditor.tsx:130-177` | "Default Log Groups" `<Field label>` and description; delegates to `<LogGroupsFieldWrapper>` which writes `jsonData.logGroups` (new object shape) via `onChange` and `jsonData.defaultLogGroups` (legacy string shape) via `legacyOnChange`, migrating the legacy value when it fetches ARNs |
| `src/components/ConfigEditor/XrayLinkConfig.tsx:23-90` | "Application Signals trace link" section title / description; writes `jsonData.tracingDatasourceUid` via `DataSourcePicker` with `pluginId="grafana-x-ray-datasource"`; the "Data source" `<Field>` label is what the schema records |
| `src/components/ConfigEditor/SecureSocksProxySettingsNewStyling.tsx:13-32` | Writes `jsonData.enableSecureSocksProxy` — excluded per AGENTS.md |
| `src/types.ts:36-56` | `CloudWatchJsonData extends AwsAuthDataSourceJsonData` (adds customMetricsNamespaces, logsTimeout, logGroups, defaultLogGroups, tracingDatasourceUid; also declares `timeField` and `database` which are frontend-only leftovers); `CloudWatchSecureJsonData extends AwsAuthDataSourceSecureJsonData` (adds accessKey, secretKey — same as base) |
| `src/dataquery.ts:326-343` | `LogGroup` interface: required `arn` and `name`, optional `accountId` and `accountLabel` |
| `pkg/cloudwatch/models/settings.go:13-24` | Backend `CloudWatchSettings` struct: embeds `awsds.AWSDatasourceSettings` and adds `Namespace string \`json:"customMetricsNamespaces"\``, `SecureSocksProxyEnabled bool \`json:"enableSecureSocksProxy"\``, `LogsTimeout Duration \`json:"logsTimeout"\`` |
| `pkg/cloudwatch/models/settings.go:26-50` | `LoadCloudWatchSettings`: parses jsonData, calls `awsds.Load` (which copies decrypted secrets), then applies the 30-minute LogsTimeout default and reads AWS auth settings from the Grafana context |
| `pkg/cloudwatch/models/settings.go:52-77` | Custom `Duration.UnmarshalJSON`: accepts float64 (nanoseconds) or string (`time.ParseDuration` — the empty string leaves the value zero, letting the LoadCloudWatchSettings default fire); returns a downstream error on parse failure |
| `pkg/cloudwatch/models/settings_test.go:17-279` | Verbatim expectations for LoadCloudWatchSettings: parse of `keys`, `arn`-legacy interpretation, sessionToken loading, and the four Duration cases (default when unset, default when empty string, string duration, raw nanosecond number, invalid duration returns downstream error) |
| `package.json:85-97` | External component versions (see next table) |
| `go.mod` | `github.com/grafana/grafana-aws-sdk v1.4.4`, `github.com/grafana/grafana-plugin-sdk-go v0.291.1` |

### External editor components

Read at the exact versions pinned in the plugin's `package.json`.

| Component | Version | Source consulted | What was read |
| --- | --- | --- | --- |
| `ConnectionConfig`, `AwsAuthType`, `Divider` | `@grafana/aws-sdk@0.10.2` | grafana/grafana-aws-sdk-react tag `v0.10.2` (SHA `fe0c4d8`), `src/components/ConnectionConfig.tsx`, `src/providers.ts`, `src/types.ts`, `src/regions.ts` | Every field's label, placeholder, description, and conditional render; the Select option labels for `authType`; the standard region list used when `loadRegions` is not passed |
| `ConfigSection`, `DataSourceDescription`, `EditorField`, `EditorRow` | `@grafana/plugin-ui@^0.13.0` | Editor layout, intro block, log-group scope selector — no storage fields |
| `Field`, `Input`, `Select`, `ButtonGroup`, `Divider`, `Alert` | `@grafana/ui@^13.0.0` | Prop names (`label`, `placeholder`, `value`, `onChange`, `isConfigured`, `onReset`) so we knew which UI attributes to record |
| `DataSourcePicker` | `@grafana/runtime@^13.0.0` | Prop shape for `XrayLinkConfig` (`pluginId`, `onChange({ uid })`, `current`, `noDefault`) |
| `DataSourcePluginOptionsEditorProps`, `onUpdateDatasourceJsonDataOption`, `updateDatasourcePluginJsonDataOption`, `rangeUtil.describeInterval` | `@grafana/data@^13.0.0` | Storage-key semantics of the update helpers used by ConnectionConfig and by CloudWatch's editor; the timeout validator |

### Backend Go dependency (`grafana-aws-sdk` `v1.4.4`)

| File | What was read |
| --- | --- |
| `pkg/awsds/settings.go:13-91` | `AuthType` int enum + custom `MarshalJSON`/`UnmarshalJSON` — the storage⇆enum mapping (including legacy `arn`→`default` and `sharedCreds`→`credentials`) that we surface as `AWSAuthType` string constants |
| `pkg/awsds/settings.go:94-117` | `AWSDatasourceSettings` struct with the AWS-shared fields; note `AssumeRoleARN string \`json:"assumeRoleARN"\`` (uppercase RN) versus the frontend's camelCase `assumeRoleArn` — case-insensitive Unmarshal makes both work |
| `pkg/awsds/settings.go:120-141` | `Load` copies decrypted `accessKey`, `secretKey`, `sessionToken`, and `proxyPassword` from secure JSON data; mirrors `defaultRegion`→`region` and falls back to the root-level `Database` field when `profile` is empty (legacy CloudWatch support) |

## Field provenance

Every dsconfig field, traced from its schema `id` to the upstream `file:line` where each
of its label, placeholder, tooltip, default, storage key, and value type is defined.

| Schema `id` | Storage key | Target | Label source | Placeholder / options / default source | Value type source | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `jsonData_authType` | `authType` | `jsonData` | `ConnectionConfig.tsx:105` (`<Field label="Authentication Provider">`) | Options `providers.ts:4-25`; description `ConnectionConfig.tsx:106`; default `awsds/settings.go:16` (iota-zero) → `"default"` | `AWSDatasourceSettings.AuthType` `awsds/settings.go:97` (int enum, string on wire via custom Marshal) | Role `auth.discriminator`; validation `allowedValues` includes legacy `arn` |
| `jsonData_profile` | `profile` | `jsonData` | `ConnectionConfig.tsx:123` (`<Field label="Credentials Profile Name">`) | Placeholder `:129` (`"default"`); description `:124` | `AWSDatasourceSettings.Profile` `awsds/settings.go:95` | `dependsOn` from conditional render `ConnectionConfig.tsx:121` |
| `secureJsonData_accessKey` | `accessKey` | `secureJsonData` | `ConnectionConfig.tsx:137` (`<Field label="Access Key ID">`) | — | Decrypted secret; copied at `awsds/settings.go:135` into `AccessKey` (`awsds/settings.go:113`, `json:"-"`) | Role `auth.aws.accessKeyId`; `dependsOn`/`requiredWhen` from `ConnectionConfig.tsx:135` |
| `secureJsonData_secretKey` | `secretKey` | `secureJsonData` | `ConnectionConfig.tsx:157` (`<Field label="Secret Access Key">`) | — | Decrypted secret; copied at `awsds/settings.go:136` into `SecretKey` (`awsds/settings.go:114`) | Role `auth.aws.secretAccessKey`; same conditional as accessKey |
| `secureJsonData_sessionToken` | `sessionToken` | `secureJsonData` | — (no UI) | — | Decrypted secret; copied at `awsds/settings.go:137` into `SessionToken` (`awsds/settings.go:115`) | Role `auth.aws.sessionToken`; tagged `backend-only`; also verified in `pkg/cloudwatch/models/settings_test.go:230-244` |
| `jsonData_assumeRoleArn` | `assumeRoleArn` | `jsonData` | `ConnectionConfig.tsx:261` (`<Field ... label="Assume Role ARN">`) | Placeholder `:268` (`"arn:aws:iam:*"`); description `:262-264` (verbatim including the newline break inside the description string) | `AWSDatasourceSettings.AssumeRoleARN` `awsds/settings.go:98` (backend json tag `assumeRoleARN`) | Editor gate: `hideAssumeRoleArn=false` (CloudWatch passes nothing) AND `awsAssumeRoleEnabled` (defaults to `true`); no dependsOn — visible for every auth type. Pattern validation is common across all AWS DS packs |
| `jsonData_externalId` | `externalId` | `jsonData` | `ConnectionConfig.tsx:276` (`<Field ... label="External ID">`) | Placeholder `:281` (`"External ID"`); description `:277` | `AWSDatasourceSettings.ExternalID` `awsds/settings.go:99` | `dependsOn` from conditional render `ConnectionConfig.tsx:273` (not `grafana_assume_role`) |
| `jsonData_proxyType` | `proxyType` | `jsonData` | `ConnectionConfig.tsx:294` (`<Field label="Proxy Type">`) | Options `:301-305` (`Environment (default)` / `None` / `URL`); description `:295`; default `:300` (`"env"`) | `AWSDatasourceSettings.ProxyType` `awsds/settings.go:108` | Editor-visible when `showHttpProxySettings` (CloudWatch passes it, `ConfigEditor.tsx:86`) AND `config.awsPerDatasourceHTTPProxyEnabled` (runtime toggle) |
| `jsonData_proxyUrl` | `proxyUrl` | `jsonData` | `ConnectionConfig.tsx:322` (`<Field label="Proxy URL">`) | Placeholder `:328` (`"Example: https://localhost:3004"`); description `:323` | `AWSDatasourceSettings.ProxyUrl` `awsds/settings.go:109` | `dependsOn`/`requiredWhen` from conditional render `:319` (`proxyType === 'url'`) |
| `jsonData_proxyUsername` | `proxyUsername` | `jsonData` | `ConnectionConfig.tsx:334` (`<Field label="Proxy Username">`) | Description `:335` (RFC-2396 warning) | `AWSDatasourceSettings.ProxyUsername` `awsds/settings.go:110` | `dependsOn` from conditional render `:319` |
| `secureJsonData_proxyPassword` | `proxyPassword` | `secureJsonData` | `ConnectionConfig.tsx:345` (`<Field label="Proxy Password">`) | Description `:346` (RFC-2396 warning) | Decrypted secret; copied at `awsds/settings.go:138` into `ProxyPassword` (`awsds/settings.go:116`) | `dependsOn` from conditional render `:319` |
| `jsonData_endpoint` | `endpoint` | `jsonData` | `ConnectionConfig.tsx:362` (`<Field label="Endpoint">`) | Placeholder `:368` (CloudWatch passes no `defaultEndpoint`, so `'https://{service}.{region}.amazonaws.com'`); description `:363` | `AWSDatasourceSettings.Endpoint` `awsds/settings.go:102` | `dependsOn` from conditional render `:360` (not `grafana_assume_role` AND not `skipEndpoint`) |
| `jsonData_defaultRegion` | `defaultRegion` | `jsonData` | `ConnectionConfig.tsx:376` (`<Field label="Default Region">`) | Description `:377` (verbatim, including padded backticks `` ` us-west-2 ` ``); options at runtime from `standardRegions` (`regions.ts`) or CloudWatch's `loadRegions` prop (`ConfigEditor.tsx:87-99`) — modelled as `select` with `allowCustom` | `AWSDatasourceSettings.DefaultRegion` `awsds/settings.go:105` | `<Select ... allowCustomValue={true}>` |
| `jsonData_customMetricsNamespaces` | `customMetricsNamespaces` | `jsonData` | `ConfigEditor.tsx:102` (`<Field label="Namespaces of Custom Metrics">`) | Placeholder `:104` (`"Namespace1,Namespace2"`); no description | Backend field `Namespace string \`json:"customMetricsNamespaces"\`` `pkg/cloudwatch/models/settings.go:18` | Field label has no description — the plugin author left it as pure input |
| `jsonData_logsTimeout` | `logsTimeout` | `jsonData` | `ConfigEditor.tsx:117` (`<Field label="Query Result Timeout">`) | Placeholder `:124` (`"30m"`); description `:118` (verbatim including the two spaces before `"30s"` in the source string) | Backend field `LogsTimeout Duration \`json:"logsTimeout"\`` `pkg/cloudwatch/models/settings.go:20` (custom UnmarshalJSON at `:52-77`) | Frontend validates via `rangeUtil.describeInterval` (`ConfigEditor.tsx:207-227`); default 30m applied in `LoadCloudWatchSettings` (`settings.go:42-44`) |
| `jsonData_logGroups` | `logGroups` | `jsonData` | `ConfigEditor.tsx:131` (`<Field label="Default Log Groups">`) | Description `:132` | `CloudWatchJsonData.logGroups?: LogGroup[]` `src/types.ts:46`; `LogGroup` shape at `src/dataquery.ts:326-343` | Not read by `CloudWatchSettings` — consumed at query time; tagged `frontend-only`. Modelled as an object-item array (`arn`, `name`, `accountId?`, `accountLabel?`) |
| `jsonData_defaultLogGroups` | `defaultLogGroups` | `jsonData` | (no dedicated UI — legacy) | Written by `LegacyLogGroupSelection` when the `cloudWatchCrossAccountQuerying` feature toggle is off | `CloudWatchJsonData.defaultLogGroups?: string[]` `src/types.ts:50` (`@deprecated use logGroups`) | Tagged `frontend-only, legacy`; kept for round-trip fidelity |
| `jsonData_tracingDatasourceUid` | `tracingDatasourceUid` | `jsonData` | `XrayLinkConfig.tsx:43` (`<Field ... label="Data source">`) | Description `:44` (`"Application Signals data source containing traces"`) | `CloudWatchJsonData.tracingDatasourceUid?: string` `src/types.ts:44` | Not read by CloudWatchSettings — powers a frontend link only; `datasourceReference` relationship targets `grafana-x-ray-datasource` |

## Field inventory summary

| Schema field | Storage key | Target | Editor label | Read by CloudWatch backend? |
| --- | --- | --- | --- | --- |
| `jsonData_authType` | `authType` | `jsonData` | Authentication Provider | Yes (awsds) |
| `jsonData_profile` | `profile` | `jsonData` | Credentials Profile Name | Yes (awsds) |
| `secureJsonData_accessKey` | `accessKey` | `secureJsonData` | Access Key ID | Yes (awsds) |
| `secureJsonData_secretKey` | `secretKey` | `secureJsonData` | Secret Access Key | Yes (awsds) |
| `secureJsonData_sessionToken` | `sessionToken` | `secureJsonData` | — (no UI) | Yes (awsds, backend-only) |
| `jsonData_assumeRoleArn` | `assumeRoleArn` | `jsonData` | Assume Role ARN | Yes (awsds) |
| `jsonData_externalId` | `externalId` | `jsonData` | External ID | Yes (awsds) |
| `jsonData_proxyType` | `proxyType` | `jsonData` | Proxy Type | Yes (awsds) |
| `jsonData_proxyUrl` | `proxyUrl` | `jsonData` | Proxy URL | Yes (awsds) |
| `jsonData_proxyUsername` | `proxyUsername` | `jsonData` | Proxy Username | Yes (awsds) |
| `secureJsonData_proxyPassword` | `proxyPassword` | `secureJsonData` | Proxy Password | Yes (awsds) |
| `jsonData_endpoint` | `endpoint` | `jsonData` | Endpoint | Yes (awsds) |
| `jsonData_defaultRegion` | `defaultRegion` | `jsonData` | Default Region | Yes (awsds → Region) |
| `jsonData_customMetricsNamespaces` | `customMetricsNamespaces` | `jsonData` | Namespaces of Custom Metrics | Yes (CloudWatchSettings.Namespace) |
| `jsonData_logsTimeout` | `logsTimeout` | `jsonData` | Query Result Timeout | Yes (CloudWatchSettings.LogsTimeout) |
| `jsonData_logGroups` | `logGroups` | `jsonData` | Default Log Groups | No (query-time consumer) |
| `jsonData_defaultLogGroups` | `defaultLogGroups` | `jsonData` | Default Log Groups (legacy) | No (query-time consumer) |
| `jsonData_tracingDatasourceUid` | `tracingDatasourceUid` | `jsonData` | Data source | No (frontend-only link) |

### Frontend-only settings

- **`logGroups`** and **`defaultLogGroups`** — written by the config editor's log-group
  selector, but never read by `CloudWatchSettings`. They surface later via the logs
  query builder at query time, so they belong on the datasource-config schema even though
  they are not part of the runtime `CloudWatchSettings`.
- **`tracingDatasourceUid`** — written by `XrayLinkConfig` to link log entries containing
  an `@xrayTraceId` field to an Application Signals (X-Ray) datasource; not read by the
  CloudWatch backend.

### Backend-only settings

- **`sessionToken`** — no editor UI; still read by `awsds/settings.go:137` into
  `AWSDatasourceSettings.SessionToken`. Used with `authType: "keys"` for temporary STS
  credentials. Provisioning is the practical way to set it. Verified by
  `pkg/cloudwatch/models/settings_test.go:230-244`.

### Fields excluded from this entry

- **Secure Socks Proxy** (`jsonData.enableSecureSocksProxy`) — excluded per AGENTS.md.
  Written by `SecureSocksProxySettingsNewStyling` (`src/components/ConfigEditor/`) and
  consumed by the backend `CloudWatchSettings.SecureSocksProxyEnabled` field
  (`pkg/cloudwatch/models/settings.go:19`), but registry entries deliberately omit it.
- **`timeField`** — declared on the frontend `CloudWatchJsonData` type
  (`src/types.ts:37`) but never written by the config editor and never read by
  `CloudWatchSettings`. Vestigial from an older datasource shape.
- **`database` (jsonData variant)** — declared on `CloudWatchJsonData`
  (`src/types.ts:38`) but not written by the config editor. `awsds.Load` reads the
  *root-level* `settings.Database` field (`awsds/settings.go:132`) as a legacy profile
  fallback for CloudWatch — that is the plugin instance settings' top-level `database`
  field, not a jsonData property, and it is not a stored config decision this schema
  needs to model.
- **`region`** — the AWS SDK Go struct has a `Region` field (`awsds/settings.go:96`),
  but the frontend never writes it; the backend mirrors `defaultRegion` into it at load
  time (`awsds/settings.go:127-129`).

## Where the types are defined

The CloudWatch configuration types are spread across the plugin and its dependencies.
Some fields and base types come from libraries/SDKs rather than the plugin itself.

### Frontend (TypeScript)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `CloudWatchJsonData` (jsonData), `CloudWatchSecureJsonData` | `src/types.ts:36-56` | plugin ([grafana/grafana-cloudwatch-datasource](https://github.com/grafana/grafana-cloudwatch-datasource)) |
| `LogGroup` (item shape of `jsonData.logGroups`) | `src/dataquery.ts:326-343` | plugin |
| `AwsAuthType`, `AwsAuthDataSourceJsonData` (base of `CloudWatchJsonData`), `AwsAuthDataSourceSecureJsonData` (base of `CloudWatchSecureJsonData`) | `src/types.ts:3-32` | `@grafana/aws-sdk` `0.10.2` (grafana/grafana-aws-sdk-react `v0.10.2`) |
| `awsAuthProviderOptions` (Select options for `authType`) | `src/providers.ts:4-25` | `@grafana/aws-sdk` `0.10.2` |
| `ConnectionConfig` React component (renders every AWS-shared field) | `src/components/ConnectionConfig.tsx:36-404` | `@grafana/aws-sdk` `0.10.2` |
| `standardRegions` list (fallback region options) | `src/regions.ts` | `@grafana/aws-sdk` `0.10.2` |
| `LogGroupsField`, `LogGroupsFieldWrapper`, `LegacyLogGroupSelection` | `src/components/shared/LogGroups/` | plugin |
| `XrayLinkConfig` | `src/components/ConfigEditor/XrayLinkConfig.tsx` | plugin |
| `DataSourceJsonData` (base type of `AwsAuthDataSourceJsonData`) | `packages/grafana-data/src/types/datasource.ts` | `@grafana/data` `^13.0.0` |
| `DataSourcePicker` (writes `tracingDatasourceUid`) | `@grafana/runtime` | `@grafana/runtime` `^13.0.0` |
| `ConfigSection`, `DataSourceDescription`, `EditorField`, `EditorRow` | `@grafana/plugin-ui` | `@grafana/plugin-ui` `^0.13.0` |
| `SecureSocksProxySettings` (excluded from this entry — the plugin uses its own `SecureSocksProxySettingsNewStyling`) | `packages/grafana-ui/src/components/DataSourceSettings/SecureSocksProxySettings.tsx` | `@grafana/ui` `^13.0.0` |

### Backend (Go)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `CloudWatchSettings`, `LoadCloudWatchSettings`, `Duration` (custom UnmarshalJSON) | `pkg/cloudwatch/models/settings.go:13-77` | plugin ([grafana/grafana-cloudwatch-datasource](https://github.com/grafana/grafana-cloudwatch-datasource)) |
| `AWSDatasourceSettings` (embedded base of `CloudWatchSettings`), `AuthType` int enum + custom Marshal/Unmarshal | `pkg/awsds/settings.go:13-141` | `github.com/grafana/grafana-aws-sdk` `v1.4.4` |
| `backend.DataSourceInstanceSettings` (carries `JSONData`, `DecryptedSecureJSONData`) | `backend/common.go` | `github.com/grafana/grafana-plugin-sdk-go` `v0.291.1` |
| CloudWatch service clients (Metrics, Logs, OAM, EC2) | `pkg/cloudwatch/models/api.go`, `pkg/cloudwatch/**/*.go` | plugin |
| AWS SDK v2 clients (`cloudwatch.NewFromConfig(...)`, `cloudwatchlogs.NewFromConfig(...)`, `oam.NewFromConfig(...)`, `sts.NewFromConfig(...)`) | — | `github.com/aws/aws-sdk-go-v2` |

The models in this entry flatten that spread into a single Go `Config` type (jsonData
fields + `DecryptedSecureJSONData`) plus a `SecureJsonDataKey` typed constant list.
`settings.ts` keeps the three canonical TypeScript types (`RootConfig`, `JsonDataConfig`,
`SecureJsonDataConfig`). The `Duration` type is mirrored verbatim from the upstream
plugin (same UnmarshalJSON logic) so LoadConfig accepts the same duration inputs
`LoadCloudWatchSettings` accepts. `AWSAuthType` and `AWSProxyType` constants mirror the
string forms `awsds.AuthType` (Un)Marshals to and the `proxyType` Select stores, without
carrying the int-enum + custom-JSON machinery.

## Modeling decisions

- **Single, flat Go `Config`.** The upstream `CloudWatchSettings` embeds
  `awsds.AWSDatasourceSettings`. We flatten those into one struct to match the pattern
  used by other dsconfig registry entries (see `registry/grafana-athena-datasource`,
  `registry/grafana-github-datasource`) and to avoid pulling `grafana-aws-sdk` into the
  shared registry `go.mod`.
- **camelCase json tag for `AssumeRoleARN`.** The Go field name mirrors upstream
  `awsds.AWSDatasourceSettings.AssumeRoleARN`, but the json tag is `assumeRoleArn` (what
  the frontend writes). Go's `encoding/json` is case-insensitive, so both spellings load
  correctly. A dedicated LoadConfig test locks the PascalCase spelling in.
- **`Duration` mirrored verbatim.** `settings.go` includes its own `Duration` type
  duplicating the plugin's `models.Duration.UnmarshalJSON` (string parsing via
  `time.ParseDuration`, raw nanoseconds via `float64`, empty string as zero, downstream
  error on failure). This lets `LoadConfig` accept the same on-wire shapes the plugin's
  own `LoadCloudWatchSettings` accepts, and matches the settings_test.go cases upstream.
- **Proxy fields included.** CloudWatch is one of the few AWS datasources that opts into
  `showHttpProxySettings` (`src/components/ConfigEditor/ConfigEditor.tsx:86`), so
  `proxyType` / `proxyUrl` / `proxyUsername` / `proxyPassword` are part of its editor
  surface even though the runtime `awsPerDatasourceHTTPProxyEnabled` toggle further
  gates them. Athena's entry excludes them because Athena doesn't opt in.
- **`grafana_assume_role` is a schema-only value.** The editor only renders it when
  the `awsDatasourcesTempCredentials` feature toggle is on and the plugin id is in
  `DS_TYPES_THAT_SUPPORT_TEMP_CREDS` (which does include `"cloudwatch"`,
  `ConnectionConfig.tsx:18-28`). We list it as a schema `allowedValues` entry, mark it
  as visible-when in the depends-on expressions of externalId / endpoint, and note the
  gating in an instruction.
- **`arn` kept in `allowedValues`, not in UI options.** `AwsAuthType.ARN` is deprecated
  (`grafana-aws-sdk-react/src/types.ts:11`) and does not appear in
  `awsAuthProviderOptions`, but stored datasources may still carry it. Listing it in
  `allowedValues` (but not in the Select's `options`) matches the backend's tolerance.
  The editor surfaces `ARN_DEPRECATION_WARNING_MESSAGE` (`ConfigEditor.tsx:29-30`) when
  this value is loaded.
- **`logGroups` modelled as an object-item array; `defaultLogGroups` as a string array.**
  Both are stored simultaneously by the editor (the modern selector writes `logGroups`
  and clears `defaultLogGroups`; the legacy selector writes `defaultLogGroups` and
  leaves `logGroups` empty). Represented with the dsconfig `item.valueType: "object"`
  schema (required `arn` / `name`; optional `accountId` / `accountLabel`) and
  `item.valueType: "string"` respectively.
- **`tracingDatasourceUid` modelled with a `datasourceReference` relationship.** The
  UID must resolve to a `grafana-x-ray-datasource` instance; the relationship carries
  `targetPluginType` so downstream tooling can enforce that.
- **`customMetricsNamespaces` label carries no description.** The upstream editor
  simply renders `<Field label="Namespaces of Custom Metrics">` without a description
  string (`ConfigEditor.tsx:102`). The schema keeps it that way — no invented tooltip.
- **`sessionToken` and `proxyPassword` included as secure keys with special handling.**
  `sessionToken` has no editor UI (backend-only, same as Athena's entry).
  `proxyPassword` does have a UI, gated on `proxyType == 'url'`. Both are `write-only`
  from the schema's perspective; consumers read `secureJsonFields` to check
  configuration state.
- **Runtime-required fields not marked `required` in the schema.** `defaultRegion` is
  effectively required at runtime — the backend uses it to build every AWS client — but
  the editor doesn't mark it required. `requiredWhen: "true"` would be misleading given
  the editor accepts saving without it, and the athena precedent (which does mark its
  selectors `requiredWhen: "true"`) is a poor match here because the CloudWatch editor
  doesn't gate its save button on region. LoadConfig's `Validate` step enforces it
  instead.
- **Secure Socks Proxy field excluded.** Excluded per AGENTS.md. The plugin uses its
  own `SecureSocksProxySettingsNewStyling` (not the `@grafana/ui` component) — same
  storage key `jsonData.enableSecureSocksProxy`, same exclusion.
- **`SecureJsonDataConfig` is a key list.** Secure values are write-only, so the type
  is just the array of secret key names (`accessKey`, `secretKey`, `sessionToken`,
  `proxyPassword`); consumers read `secureJsonFields` to see what is configured.

## SDK plugin schema and k8s-style examples (`schema.go`)

`NewSchema()` assembles the `grafana-plugin-sdk-go` `pluginschema.PluginSchema` bundle
(the k8s-style schema Grafana's datasource API server serves as `{apiVersion}.json`,
`v0alpha1` today) from the embedded `dsconfig.json`: top-level jsonData fields (and
nested array items) become the OpenAPI settings `spec`; secure fields become
`secureValues`.

`SettingsExamples()` provides the default configuration plus one k8s-style example per
AWS auth provider, plus a URL-HTTP-proxy variant, a CloudWatch-Logs defaults variant,
the legacy `arn` value, and the legacy `defaultLogGroups` shape:

| Example | Auth | Extras | `secureJsonData` |
| --- | --- | --- | --- |
| `""` (default) | AWS SDK Default | proxyType=env, logsTimeout=30m0s | `accessKey` (empty) |
| `awsSdkDefault` | AWS SDK Default | region + customMetricsNamespaces | `accessKey` (empty) |
| `accessAndSecretKey` | Access & secret key | region | `accessKey`, `secretKey` |
| `credentialsFile` | Credentials file | `profile`, region | `accessKey` (empty) |
| `workspaceIamRole` | Workspace IAM Role | region | `accessKey` (empty) |
| `grafanaAssumeRole` | Grafana Assume Role | region | `accessKey` (empty) |
| `assumeRoleFromKeys` | Access & secret key + STS AssumeRole | `assumeRoleArn`, `externalId` | `accessKey`, `secretKey` |
| `urlProxy` | AWS SDK Default | proxyType=url + proxyUrl + proxyUsername | `proxyPassword` |
| `cloudwatchLogsDefaults` | AWS SDK Default | logsTimeout=10m + logGroups[] + tracingDatasourceUid + customMetricsNamespaces | `accessKey` (empty) |
| `legacyArnAuthType` | `arn` (legacy — backend maps to `default`) | region | `accessKey` (empty) |
| `legacyDefaultLogGroups` | AWS SDK Default | `defaultLogGroups` (deprecated string array) | `accessKey` (empty) |

Every example carries at least one `secureJsonData` placeholder as required by the
conformance suite.

## `LoadConfig` utility (`settings.go`)

`LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error)`
runs the full three-phase load flow on a datasource instance's settings and returns a
fully-defaulted, validated `Config`:

1. **Parse** — unmarshal `settings.JSONData` into `Config` (`len(JSONData) > 1` gate
   mirrors `LoadCloudWatchSettings`, so `"{}"` is tolerated the same way); the
   `Duration` custom UnmarshalJSON accepts both string durations and raw nanosecond
   numbers exactly as the upstream `models.Duration` does. Then copy the plugin's
   decrypted secrets (`accessKey`/`secretKey`/`sessionToken`/`proxyPassword`) into
   `DecryptedSecureJSONData`.
2. **`ApplyDefaults`** — fills three curated defaults: `AuthType` →
   `AWSAuthTypeDefault` (matches reference pack and backend iota-zero), `ProxyType` →
   `AWSProxyTypeEnv` (matches `ConnectionConfig.tsx:300`), and `LogsTimeout` →
   `DefaultLogsTimeout` (30m, matches `pkg/cloudwatch/models/settings.go:42-44`).
3. **`Validate`** — enforces the runtime contract: known `AuthType`, `accessKey` +
   `secretKey` present for `keys` auth, known `ProxyType`, `proxyUrl` present when
   `proxyType == "url"`, and non-empty `defaultRegion`. Errors are joined so callers
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

1. **`AssumeRoleARN` json tag disagrees with the frontend's `assumeRoleArn`.**
   `awsds/settings.go:98` uses `` json:"assumeRoleARN" `` (uppercase RN), but the
   frontend type (`grafana-aws-sdk-react/src/types.ts:17`) and every `onChange` in
   `ConnectionConfig` write `assumeRoleArn` (lowercase arn). Case-insensitive `encoding/json`
   Unmarshal makes both work, but a stricter decoder would fail. Same latent trap
   affects every AWS datasource.
2. **`logsTimeout` accepts two totally different on-wire shapes.**
   `pkg/cloudwatch/models/settings.go:52-77` decodes either a string (`"30m"`,
   `"2000ms"`, `"1.5s"`) or a raw float64 nanosecond number (`1500000000`). Provisioning
   authors picking the wrong one silently get a different value: `1500000000` is 1.5s,
   not 1.5 hours as a human might guess. The editor only writes strings; the number
   path is a JSON quirk exposed only via provisioning. The schema's `valueType: string`
   only captures the editor path; consumers should treat the numeric path as an
   undocumented fallback.
3. **`logsTimeout` invalid duration returns a `downstream` error, not a validation
   error.** `pkg/cloudwatch/models/settings.go:69` wraps `time.ParseDuration`'s error in
   `backend.DownstreamError`, which the plugin SDK treats as caller-fault. That's
   correct if a query-time value is bad, but at *config-load time* it presents a bad
   provisioning payload as if the AWS API were misbehaving. Users chasing the error
   have to dig to realize it's their config.
4. **Deprecated `arn` auth value silently maps to `default`.** `awsds/settings.go:87-88`
   maps any unknown auth type string to `AuthTypeDefault` with the comment "For old
   'arn' option". `ConfigEditor.tsx:63-64` sets `ARN_DEPRECATION_WARNING_MESSAGE` as a
   dismissable banner but does not rewrite the stored value; the datasource continues to
   store `authType: "arn"` on the next save unless the user explicitly picks a new
   provider.
5. **Legacy `sharedCreds` auth value silently mapped to `credentials`.**
   `awsds/settings.go:75-78` (`case "credentials"` falls through to `sharedCreds`) folds
   both storage values onto the same enum. Not preserved in this schema as an allowed
   value (the reference pack also omits it); a datasource stored with
   `authType: "sharedCreds"` fails the schema's `allowedValues` check but still loads
   fine on the backend.
6. **Auth type default depends on Grafana instance config.** The editor's `useEffect`
   (`ConnectionConfig.tsx:75-90`) picks `awsAllowedAuthProviders[0]`, preferring
   `grafana_assume_role` if the feature is on. So the "default" auth type a new user
   sees is Grafana-instance-dependent, not always `default`. The stored schema default
   here is `default` (matching the reference pack and the backend iota zero), which may
   not be what a fresh Grafana Cloud editor writes.
7. **`sessionToken` has no editor UI.** `grafana-aws-sdk-react` never exposes an input
   for `sessionToken`, even though the backend reads it and it's required for temporary
   credentials. Users must provision it directly. `pkg/cloudwatch/models/settings_test.go:230-244`
   explicitly tests that provisioning path.
8. **`proxyPassword` gated on TWO conditions the user can't see.** Proxy fields render
   only when `showHttpProxySettings` (compile-time, only some plugins pass it — CloudWatch
   does) AND `config.awsPerDatasourceHTTPProxyEnabled` (runtime, feature-toggle-like).
   The second is a `@ts-ignore` in `ConnectionConfig.tsx:56`. Users of an instance
   without the toggle can never enter a proxy password through the editor; provisioning
   is the only route.
9. **The "Default Region" description has padded backticks.** `ConnectionConfig.tsx:377`
   reads `` `Specify the region, such as for US West (Oregon) use ` us-west-2 ` as the
   region.` `` — note the spaces around `us-west-2` inside the backticks. Preserved
   verbatim in the schema; would render inline code with padding in some Markdown
   renderers.
10. **`logGroups` migrates `defaultLogGroups` but never clears it deterministically.**
    `LogGroupsField` (`src/components/shared/LogGroups/LogGroupsField.tsx:68-98`) fetches
    ARNs for each name in `defaultLogGroups`, then hands the resulting `logGroups` array
    to `onChange`. The editor's own `onChange` handler at `ConfigEditor.tsx:157-166`
    does clear `defaultLogGroups: undefined` on every write, but only when the user
    interacts with the log group control — legacy-only configs that never touch the
    control retain both fields side by side. Consumers must read both.
11. **`CloudWatchJsonData` declares `timeField` and `database` fields that no code
    writes or reads.** `src/types.ts:37-38` still declares these two properties. They
    have no config editor UI and no backend consumer (`CloudWatchSettings` doesn't
    mention them; the root-level `settings.Database` — which IS read as a legacy
    profile fallback — is a different field entirely). Vestigial from earlier iterations.
12. **`tracingDatasourceUid` has no schema-side type constraint on the target plugin.**
    The frontend `DataSourcePicker` restricts the choice at edit time (`pluginId:
    'grafana-x-ray-datasource'`, `XrayLinkConfig.tsx:47`), but stored data has no such
    constraint. A provisioned datasource could point `tracingDatasourceUid` at any UID
    and the CloudWatch datasource would happily store it. The `datasourceReference`
    relationship in this schema carries `targetPluginType: "grafana-x-ray-datasource"`
    so at least the schema layer records the intent.
13. **"Cloudwatch" title case is inconsistent upstream.** The plugin's own display name
    is "CloudWatch" (`src/plugin.json:4`), but the editor section title is "Cloudwatch
    Logs" (`ConfigEditor.tsx:114`) and the warning strings use both "CloudWatch" and
    "Cloudwatch" almost at random. The schema preserves the section title verbatim.
14. **`cloudWatchCrossAccountQuerying` is a Grafana feature toggle, not a stored
    datasource field.** The task brief hinted at a "cross-account observability toggle
    (`crossAccount`)" storage field; there isn't one. Cross-account querying is gated
    entirely on `config.featureToggles.cloudWatchCrossAccountQuerying` (a runtime
    Grafana toggle) and affects which log-group selector renders and how log groups
    (`arn`/`accountId`) are populated. No jsonData or secureJsonData key stores the
    toggle state.

## Validation performed

- `dsconfig.ParseAndResolveSchemaJSON` + `Schema.Validate()` — passes.
- JSON Schema validation against [`dsconfig/schema.json`](../../dsconfig/schema.json)
  (draft 2020-12, `additionalProperties: false`) — passes.
- `go test ./...` on this module — passes (schema bundle shape, `SchemaSpecHasNoSecureJSON`,
  `SecureValuesMatchLoadSettings`, `JSONDataMatchesStruct`, `JSONDataTypesMatchStruct`,
  `SchemaArtifactInSync`, `LoadConfig` including PascalCase-`assumeRoleARN` decode
  and Duration string / nanosecond / empty-string / invalid cases, `ApplyDefaults`,
  `Validate` per auth type and per proxy type).
- `settings.go` / `schema.go`: `go build`, `go vet`, `gofmt` — clean.
- `settings.ts`: `tsc --noEmit --strict` — clean.
