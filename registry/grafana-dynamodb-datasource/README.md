# grafana-dynamodb-datasource

Declarative configuration schema for the [DynamoDB datasource plugin](https://github.com/grafana/dynamodb-datasource) (`grafana-dynamodb-datasource`).

## Upstream researched

- **Repo**: `github.com/grafana/dynamodb-datasource`
- **Ref**: `main`
- **Commit SHA**: `44f7fd60fd29de30e4c6d110ba528cb6f291ef63`

Every value in [`dsconfig.json`](dsconfig.json) — labels, placeholders, tooltips, option
labels/values, section titles, defaults, validations, dependency and required-when
expressions, storage keys, storage targets, value types, group titles, and instructions —
is traceable to a specific `file:line` in the upstream plugin repo (or in the pinned
`@grafana/aws-sdk` version of the shared `ConnectionConfig` component) at this SHA. See
[Field provenance](#field-provenance) below.

To reproduce this research:

```bash
git clone --depth 20 https://github.com/grafana/dynamodb-datasource
cd dynamodb-datasource
git checkout 44f7fd60fd29de30e4c6d110ba528cb6f291ef63

# The AWS auth surface is rendered by @grafana/aws-sdk's ConnectionConfig. Pin
# to the version dynamodb-datasource's package.json uses (currently 0.10.2):
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
| [`schema.go`](schema.go) | k8s-style SDK plugin schema: embeds `dsconfig.json`, converts it via `dsconfig.NewSDKSchema`, and defines `SettingsExamples` for each AWS auth provider plus the pre-V2 legacy shape |
| [`settings_test.go`](settings_test.go) | Table tests for `LoadConfig`, `ApplyDefaults`, and `Validate` (including parity with the upstream `pkg/models/settings_test.go` cases) |
| [`conformance_test.go`](conformance_test.go) | Runs `schema.RunPluginTests` — the shared dsconfig conformance suite — against `dsconfig.json` |
| [`schema.gen.json`](schema.gen.json), [`settings.gen.json`](settings.gen.json), [`settings.examples.gen.json`](settings.examples.gen.json) | Committed schema artifacts (regenerate with `go generate ./...` inside this directory; the `SchemaArtifactInSync` conformance subtest guards drift) |

There is no per-entry `go.mod` — every registry entry is a subpackage of the shared
[`registry/`](..) module (`github.com/grafana/dsconfig/registry`).

## Sources researched

Every source below was read at the pinned upstream SHA
(`44f7fd60fd29de30e4c6d110ba528cb6f291ef63`) or, for `@grafana/aws-sdk`, at the exact
version pinned in the plugin's `package.json` (`0.10.2` / SHA `fe0c4d8`).

### Plugin repo (`github.com/grafana/dynamodb-datasource@44f7fd6`)

| File | What was read |
| --- | --- |
| `src/plugin.json:1-37` | `pluginType` (`id`), `pluginName` (`name`), `docURL` (first `info.links[]` docs URL) |
| `src/components/ConfigEditor.tsx:1-69` | Entire editor: `<DataSourceDescription>` intro + `<ConnectionConfig options={options} onOptionsChange={onOptionsChange} hideAssumeRoleArn />`. No plugin-specific UI fields. The useEffect on lines 28-42 seeds `jsonData.isV2 = true` and `jsonData.authType = Keys` on empty jsonData; it also runs `migrateOptions(options)` for pre-V2 configs |
| `src/components/ConfigEditor.tsx:66` | `hideAssumeRoleArn` is set on `<ConnectionConfig>` — hides the entire Assume Role subsection (both `assumeRoleArn` and `externalId`) |
| `src/components/ConfigEditor.tsx:53-59` | Warning banner shown when the loaded config was V1 — user-visible reminder to re-enter secret credentials after migration |
| `src/types.ts:19-31` | `DynamoDBConfigOptions extends AwsAuthDataSourceJsonData` — adds `timeout`, `retries`, `pause`, `isV2`, `region` (legacy), `accessId` (legacy). `DynamoDBSecureConfigOptions extends AwsAuthDataSourceSecureJsonData` — adds nothing |
| `src/utils.ts:4-22` | `migrateOptions`: destructures `region`, `endpoint`, `accessId` out of jsonData, then rebuilds with `endpoint`, `defaultRegion: region`, `authType: Keys`, `isV2: true`; also flips `secureJsonFields.accessKey = true` and `secureJsonFields.secretKey = true` |
| `pkg/models/settings.go:26-35` | Backend `Settings` — embeds `awsds.AWSDatasourceSettings` and adds `LegacyAccessKey` (json tag `accessId`), `IsV2`, `Timeout`, `Retries`, `Pause` |
| `pkg/models/settings.go:38-85` | `LoadSettings`: parses jsonData, runs V1 migration when `IsV2 == false` (forces `AuthType = keys`; folds `LegacyAccessKey` and `secureJsonData.accessKey` into the modern AccessKey/SecretKey pair; copies `Region` → `DefaultRegion` when the latter is empty; picks up `sessionToken` from decrypted secure data), fails when `AccessKey`/`SecretKey`/`DefaultRegion` are missing, and finally applies `Timeout`/`Pause`/`Retries` defaults of `"60"`/`"5"`/`"5"` |
| `pkg/models/settings.go:87-121` | `DriverSettings`, `timeout`, `pause`, `Retries` — how the string driver settings feed sqlds at query time (`utils.ParseInt`) |
| `pkg/models/settings.go:123-128` | `migrateToSecureKey`: writes `settings.AccessKey = accessKey` (the LegacyAccessKey / jsonData.accessId) and, only when `settings.SecretKey == ""`, `settings.SecretKey = secretKey` (the V1 secureJsonData.accessKey value — hence the naming quirk) |
| `pkg/utils/utils.go` | `ParseInt` — lenient string→int for timeout/retries/pause |
| `pkg/models/settings_test.go` | Upstream test cases; our LoadConfig tests carry the same V1/V2 scenarios verbatim |
| `package.json` | External component versions (see next table) |

### External editor components

Read at the exact versions pinned in the plugin's `package.json`.

| Component | Version | Source consulted | What was read |
| --- | --- | --- | --- |
| `ConnectionConfig`, `AwsAuthType` | `@grafana/aws-sdk@0.10.2` | grafana/grafana-aws-sdk-react tag `v0.10.2` (SHA `fe0c4d8`), `src/components/ConnectionConfig.tsx`, `src/providers.ts`, `src/types.ts`, `src/regions.ts` | Every field's label, placeholder, description, and conditional render; the Select option labels for `authType`; the standard region list. Notably: `DS_TYPES_THAT_SUPPORT_TEMP_CREDS` (lines 18-28) does **not** include `grafana-dynamodb-datasource`, so the `grafana_assume_role` provider is filtered out of the Select at render time for DynamoDB |
| `DataSourceDescription` | `@grafana/plugin-ui@^0.13.0` | Editor intro block — no storage fields |
| `Alert`, `useTheme2` | `@grafana/ui@^12.0.1` | Warning banner for V1 configs; theming |
| `DataSourcePluginOptionsEditorProps`, `DataSourceSettings` | `@grafana/data@^12.0.2` | Storage-key semantics of the update helpers used by ConnectionConfig |

### Backend Go dependency (`grafana-aws-sdk`)

| File | What was read |
| --- | --- |
| `pkg/awsds/settings.go:13-91` (`v1.4.6`) | `AuthType` int enum + custom `MarshalJSON`/`UnmarshalJSON` — the storage⇆enum mapping (including legacy `arn`→`default` and `sharedCreds`→`credentials`) that we surface as `AWSAuthType` string constants |
| `pkg/awsds/settings.go:94-117` (`v1.4.6`) | `AWSDatasourceSettings` struct with the AWS-shared fields; note `AssumeRoleARN string \`json:"assumeRoleARN"\`` (uppercase `ARN`) while the frontend writes camelCase `assumeRoleArn` — not relevant for DynamoDB because those fields are hidden by `hideAssumeRoleArn` |
| `pkg/awsds/settings.go:120-141` (`v1.4.6`) | `Load` copies decrypted `accessKey`, `secretKey`, `sessionToken`, and `proxyPassword` from secure JSON data; also mirrors `defaultRegion`→`region` at load time |

## Field provenance

Every dsconfig field, traced from its schema `id` to the upstream `file:line` where each
of its label, placeholder, tooltip, default, storage key, and value type is defined.

| Schema `id` | Storage key | Target | Label source | Placeholder / options / default source | Value type source | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `jsonData_authType` | `authType` | `jsonData` | `ConnectionConfig.tsx:105` (`<Field label="Authentication Provider">`) | Options `providers.ts:4-24` filtered by ConnectionConfig.tsx:58-64 (removes `grafana_assume_role` for DynamoDB); description `ConnectionConfig.tsx:106`; default `awsds/settings.go:16` (`AuthTypeDefault` iota) → `"default"` | `AWSDatasourceSettings.AuthType` `awsds/settings.go:97` (int enum, string on wire) | Role `auth.discriminator`; validation `allowedValues` includes legacy `arn` but excludes `grafana_assume_role` (DynamoDB is not in `DS_TYPES_THAT_SUPPORT_TEMP_CREDS`) |
| `jsonData_profile` | `profile` | `jsonData` | `ConnectionConfig.tsx:123` (`<Field label="Credentials Profile Name">`) | Placeholder `ConnectionConfig.tsx:129` (`"default"`); description `:124` | `AWSDatasourceSettings.Profile` `awsds/settings.go:95` | `dependsOn` from conditional render `ConnectionConfig.tsx:121` |
| `secureJsonData_accessKey` | `accessKey` | `secureJsonData` | `ConnectionConfig.tsx:137` (`<Field label="Access Key ID">`) | — | `AWSDatasourceSettings.AccessKey` `awsds/settings.go:113` | Role `auth.aws.accessKeyId`; `dependsOn`/`requiredWhen` from conditional render `ConnectionConfig.tsx:135`. Under V1 storage this key actually held the SECRET, not the ID — see upstream findings |
| `secureJsonData_secretKey` | `secretKey` | `secureJsonData` | `ConnectionConfig.tsx:157` (`<Field label="Secret Access Key">`) | — | `AWSDatasourceSettings.SecretKey` `awsds/settings.go:114` | Role `auth.aws.secretAccessKey`; `dependsOn`/`requiredWhen` from conditional render `ConnectionConfig.tsx:135` |
| `secureJsonData_sessionToken` | `sessionToken` | `secureJsonData` | — (no UI) | — | `AWSDatasourceSettings.SessionToken` `awsds/settings.go:115` | Role `auth.aws.sessionToken`; tagged `backend-only`; still read at `dynamodb-datasource/pkg/models/settings.go:50,54` |
| `jsonData_endpoint` | `endpoint` | `jsonData` | `ConnectionConfig.tsx:362` (`<Field label="Endpoint">`) | Placeholder `:368` (DynamoDB passes no `defaultEndpoint`, so `'https://{service}.{region}.amazonaws.com'`); description `:363` | `AWSDatasourceSettings.Endpoint` `awsds/settings.go:102` | Editor-visible for every provider — DynamoDB does not pass `skipEndpoint`, and `grafana_assume_role` (which would hide it via `ConnectionConfig.tsx:360`) is unreachable for this plugin |
| `jsonData_defaultRegion` | `defaultRegion` | `jsonData` | `ConnectionConfig.tsx:376` (`<Field label="Default Region">`) | Description `:377` (verbatim, including the padded backticks `` ` us-west-2 ` ``); options at runtime from `standardRegions` (`regions.ts:1-47`) or the `standardRegions` prop — modelled as `select` with `allowCustom` | `AWSDatasourceSettings.DefaultRegion` `awsds/settings.go:105` | `<Select ... allowCustomValue={true}>` |
| `jsonData_isV2` | `isV2` | `jsonData` | — (no UI) | `src/components/ConfigEditor.tsx:35` writes `isV2: true` on empty jsonData | `models.Settings.IsV2` `pkg/models/settings.go:30` (`bool`) | Tagged `frontend-only,migration` — no user-visible input, but the editor's useEffect writes it and the backend reads it at `pkg/models/settings.go:44` to gate V1 migration |
| `jsonData_timeout` | `timeout` | `jsonData` | — (no UI) | Backend default `"60"` (`pkg/models/settings.go:75-77`) | `models.Settings.Timeout` `pkg/models/settings.go:31` (`string`) | Tagged `backend-only,driver`; parsed with `utils.ParseInt` at query time |
| `jsonData_retries` | `retries` | `jsonData` | — (no UI) | Backend default `"5"` (`pkg/models/settings.go:81-83`, and `defaultRetries = 5` at :20) | `models.Settings.Retries` `pkg/models/settings.go:32` (`string`) | Tagged `backend-only,driver` |
| `jsonData_pause` | `pause` | `jsonData` | — (no UI) | Backend default `"5"` (`pkg/models/settings.go:78-80`) | `models.Settings.Pause` `pkg/models/settings.go:33` (`string`) | Tagged `backend-only,driver` |
| `jsonData_accessId` | `accessId` | `jsonData` | — (no UI) | — | `models.Settings.LegacyAccessKey` `pkg/models/settings.go:29` (`string`, json tag `accessId`) | Tagged `backend-only,legacy`; V1 storage of the AWS Access Key ID. Under V1, `pkg/models/settings.go:46,123-128` folds this into `settings.AccessKey` |
| `jsonData_region` | `region` | `jsonData` | — (no UI) | — | `AWSDatasourceSettings.Region` `awsds/settings.go:96` (`string`) | Tagged `backend-only,legacy`; V1 storage of the region. `pkg/models/settings.go:47-49` copies it into `DefaultRegion` when the latter is empty |

## Field inventory summary

| Schema field | Storage key | Target | Editor label | Read by backend? |
| --- | --- | --- | --- | --- |
| `jsonData_authType` | `authType` | `jsonData` | Authentication Provider | Yes |
| `jsonData_profile` | `profile` | `jsonData` | Credentials Profile Name | Yes |
| `secureJsonData_accessKey` | `accessKey` | `secureJsonData` | Access Key ID | Yes |
| `secureJsonData_secretKey` | `secretKey` | `secureJsonData` | Secret Access Key | Yes |
| `secureJsonData_sessionToken` | `sessionToken` | `secureJsonData` | — (no UI) | Yes (backend-only) |
| `jsonData_endpoint` | `endpoint` | `jsonData` | Endpoint | Yes |
| `jsonData_defaultRegion` | `defaultRegion` | `jsonData` | Default Region | Yes |
| `jsonData_isV2` | `isV2` | `jsonData` | — (frontend-only marker) | Yes |
| `jsonData_timeout` | `timeout` | `jsonData` | — (no UI) | Yes (backend-only) |
| `jsonData_retries` | `retries` | `jsonData` | — (no UI) | Yes (backend-only) |
| `jsonData_pause` | `pause` | `jsonData` | — (no UI) | Yes (backend-only) |
| `jsonData_accessId` | `accessId` | `jsonData` | — (V1 legacy) | Yes (backend-only, legacy) |
| `jsonData_region` | `region` | `jsonData` | — (V1 legacy) | Yes (backend-only, legacy) |

### Frontend-only settings

- **`isV2`** — no editor input renders it, but the editor's own useEffect
  (`src/components/ConfigEditor.tsx:35`) writes it on empty jsonData and
  `src/utils.ts:15` sets it during V1→V2 migration. The backend still reads it at
  `pkg/models/settings.go:44` to decide whether to migrate.

### Backend-only settings

- **`sessionToken`** — no editor UI; still read by
  `pkg/models/settings.go:50,54` for both V1 and V2 loads. Used with `authType: "keys"`
  for temporary STS credentials. Provisioning is the practical way to set it.
- **`timeout` / `retries` / `pause`** — no editor UI; feed
  `sqlds.DriverSettings` at query time (`pkg/models/settings.go:87-121`). Strings on the
  wire, parsed with `utils.ParseInt`.
- **`accessId` / `region`** — legacy V1 fields. Kept in the schema for round-trip
  fidelity with pre-V2 datasources; the backend folds them into the modern shape at
  load time.

### Fields excluded from this entry

- **`assumeRoleArn`, `externalId`** — the plugin passes `hideAssumeRoleArn` to
  `ConnectionConfig` (`src/components/ConfigEditor.tsx:66`), which hides the entire
  Assume Role subsection. Neither field is editor-visible nor documented as
  DynamoDB-specific in `src/types.ts`. Provisioning them would work via `awsds.Load`
  under the hood, but doing so is not part of DynamoDB's declared surface — consumers
  needing STS AssumeRole should reach for the shared `aws_sdk_settings.json` pack.
- **`grafana_assume_role`** — filtered out of the Select at render time
  (`ConnectionConfig.tsx:58-64`) because `grafana-dynamodb-datasource` is not in
  `DS_TYPES_THAT_SUPPORT_TEMP_CREDS`. Removed from `allowedValues` and `ui.options`.
- **AWS proxy fields** (`proxyType`, `proxyUrl`, `proxyUsername`, `proxyPassword`) —
  `ConnectionConfig.tsx:291` only renders the proxy subsection when the caller passes
  `showHttpProxySettings`. DynamoDB's `ConfigEditor.tsx:66` does not.
- **Secure Socks Proxy** (`jsonData.enableSecureSocksProxy`) — excluded per AGENTS.md.
  Unlike Athena / X-Ray, the DynamoDB editor does not render `<SecureSocksProxySettings>`
  either, so the field would never even be written by this plugin's UI.

## Where the types are defined

The DynamoDB configuration types are spread across the plugin and its dependencies. Some
fields and base types come from libraries/SDKs rather than the plugin itself.

### Frontend (TypeScript)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `DynamoDBConfigOptions` (jsonData), `DynamoDBSecureConfigOptions` | `src/types.ts:19-31` | plugin ([grafana/dynamodb-datasource](https://github.com/grafana/dynamodb-datasource)) |
| `AwsAuthType`, `AwsAuthDataSourceJsonData` (base of `DynamoDBConfigOptions`), `AwsAuthDataSourceSecureJsonData` (base of `DynamoDBSecureConfigOptions`) | `src/types.ts:3-32` | `@grafana/aws-sdk` `0.10.2` (grafana/grafana-aws-sdk-react `v0.10.2`) |
| `awsAuthProviderOptions` (Select options for `authType`) | `src/providers.ts:4-25` | `@grafana/aws-sdk` `0.10.2` |
| `ConnectionConfig` React component (renders every AWS-shared field) | `src/components/ConnectionConfig.tsx:36-404` | `@grafana/aws-sdk` `0.10.2` |
| `standardRegions` list | `src/regions.ts:1-47` | `@grafana/aws-sdk` `0.10.2` |
| `DataSourceJsonData` (base type of `AwsAuthDataSourceJsonData`), `DataSourcePluginOptionsEditorProps` | `packages/grafana-data/src/types/datasource.ts` | `@grafana/data` `^12.0.2` |
| `DataSourceDescription` | `src/components/ConfigEditor/` | `@grafana/plugin-ui` `^0.13.0` |
| `Alert`, `useTheme2` | `packages/grafana-ui/src/*` | `@grafana/ui` `^12.0.1` |

### Backend (Go)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `Settings`, `LoadSettings`, `DriverSettings`, `migrateToSecureKey` | `pkg/models/settings.go:19-128` | plugin ([grafana/dynamodb-datasource](https://github.com/grafana/dynamodb-datasource)) |
| `AWSDatasourceSettings` (embedded base of `Settings`), `AuthType` int enum + custom Marshal/Unmarshal | `pkg/awsds/settings.go:13-141` | `github.com/grafana/grafana-aws-sdk` `v1.4.6` |
| `backend.DataSourceInstanceSettings` (carries `JSONData`, `DecryptedSecureJSONData`) | `backend/common.go` | `github.com/grafana/grafana-plugin-sdk-go` |
| `sqlds.DriverSettings` — DynamoDB SQL driver config | — | `github.com/grafana/sqlds/v5` |
| DynamoDB SQL driver / connection layer that consumes the flattened settings | `pkg/driver/*`, `pkg/database/*` | plugin ([grafana/dynamodb-datasource](https://github.com/grafana/dynamodb-datasource)) |
| AWS SDK v2 DynamoDB client build-up | — | `github.com/aws/aws-sdk-go-v2` |

The models in this entry flatten that spread into a single Go `Config` type (jsonData
fields + `DecryptedSecureJSONData`) plus a `SecureJsonDataKey` typed constant list.
`settings.ts` keeps the three canonical TypeScript types (`RootConfig`, `JsonDataConfig`,
`SecureJsonDataConfig`). `AWSAuthType` constants in `settings.go` mirror the string forms
`awsds.AuthType` (Un)Marshals, without carrying the int-enum + custom-JSON machinery.

## Modeling decisions

- **Single, flat Go `Config`.** The upstream `Settings` embeds
  `awsds.AWSDatasourceSettings`. We flatten those into one struct to match the pattern
  used by other AWS registry entries (see `grafana-athena-datasource`,
  `grafana-x-ray-datasource`) and to avoid pulling `grafana-aws-sdk` into the shared
  registry `go.mod`.
- **`AWSAuthType` allowedValues omit `grafana_assume_role`.** ConnectionConfig actively
  filters this option out of the Select for DynamoDB
  (`ConnectionConfig.tsx:53-64,58-64,113`) because
  `grafana-dynamodb-datasource` is not in `DS_TYPES_THAT_SUPPORT_TEMP_CREDS`. The
  backend `awsds.AuthType.UnmarshalJSON` would still accept it, but the plugin doesn't
  support the flow, so the schema treats it as unknown. LoadConfig / Validate reject it
  explicitly.
- **`arn` kept in `allowedValues`, not in UI options** — `AwsAuthType.ARN` is deprecated
  (`grafana-aws-sdk-react/src/types.ts:11`) and does not appear in
  `awsAuthProviderOptions`, but stored datasources may still carry it. Listing it in
  `allowedValues` (but not in the Select's `options`) matches the backend's tolerance
  (`awsds/settings.go:87-88`).
- **`requiredWhen` only on the two credential secrets.** The editor doesn't mark any
  jsonData field required; the backend enforces `defaultRegion` at load time
  (`pkg/models/settings.go:66-68`). We encode the two secret constraints via
  `requiredWhen: "jsonData_authType == 'keys'"` (matching the editor's conditional
  render) and defer the region check to `Config.Validate` since it isn't tied to
  authType.
- **`isV2` schema default is `true`** — matches what the editor writes on a fresh save
  (`src/components/ConfigEditor.tsx:35`). This does mean provisioning without setting
  `isV2` explicitly gets the modern shape rather than triggering a spurious V1
  migration.
- **V1 legacy fields (`accessId`, `region`) kept in the schema but not in any
  editor-visible group.** They live in the `legacy-migration` group (marked
  `optional: true`) so provisioning tools can round-trip pre-V2 configs verbatim
  without the schema erroring on unknown keys. LoadConfig folds them into the modern
  shape at read time.
- **Driver settings (`timeout`, `retries`, `pause`) modelled as `string`** — the
  backend's own struct types them as `string` (`pkg/models/settings.go:31-33`) and
  parses them with `utils.ParseInt`. Modelling them as `number` would drift from
  storage. Editor never writes them.
- **`SecureJsonDataConfig` is a key list** — secure values are write-only, so the type
  is just the array of secret key names (`accessKey`, `secretKey`, `sessionToken`);
  consumers read `secureJsonFields` to see what is configured. Under V1 the semantic
  meaning of `accessKey` was different (it was the secret, not the ID); the key name
  is preserved for storage compatibility.

## SDK plugin schema and k8s-style examples (`schema.go`)

`NewSchema()` assembles the `grafana-plugin-sdk-go` `pluginschema.PluginSchema` bundle
(the k8s-style schema Grafana's datasource API server serves as `{apiVersion}.json`,
`v0alpha1` today) from the embedded `dsconfig.json`: root fields plus a nested `jsonData`
object become the OpenAPI settings `spec`, secure fields become `secureValues`, and
virtual fields (none here) are skipped.

`SettingsExamples()` provides the default configuration plus one k8s-style example per
DynamoDB-supported AWS auth provider, plus a driver-settings-overrides example, a
session-token example, a legacy `arn` example, and a full pre-V2 legacy storage
example:

| Example | Auth | Extras | `secureJsonData` |
| --- | --- | --- | --- |
| `""` (default) | AWS SDK Default | — | `accessKey` (empty) |
| `awsSdkDefault` | AWS SDK Default | region + isV2 | `accessKey` (empty) |
| `accessAndSecretKey` | Access & secret key | region + isV2 + endpoint | `accessKey`, `secretKey` |
| `credentialsFile` | Credentials file | `profile`, region + isV2 | `accessKey` (empty) |
| `workspaceIamRole` | Workspace IAM Role | region + isV2 | `accessKey` (empty) |
| `keysWithSessionToken` | Access & secret key + STS session token | region + isV2 | `accessKey`, `secretKey`, `sessionToken` |
| `driverSettings` | AWS SDK Default | region + isV2 + `timeout`/`retries`/`pause` overrides | `accessKey` (empty) |
| `legacyArnAuthType` | `arn` (legacy — backend maps to `default`) | region + isV2 | `accessKey` (empty) |
| `legacyV1Shape` | (V1 — implicit `keys` on load) | `region`, `endpoint`, `accessId` (V1 storage) | `accessKey` = the V1 secret |

Every example carries at least one `secureJsonData` placeholder as required by the
conformance suite.

## `LoadConfig` utility (`settings.go`)

`LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error)`
runs the full three-phase load flow on a datasource instance's settings and returns a
fully-defaulted, validated `Config`:

1. **Parse** — unmarshal `settings.JSONData` into `Config`, then copy the plugin's
   decrypted secrets (`accessKey`/`secretKey`/`sessionToken`) into
   `DecryptedSecureJSONData`. Under V1 storage (`IsV2 == false`), reinterpret the
   secure keys: `secureJsonData.accessKey` becomes the `secretKey` value and
   `jsonData.accessId` (`LegacyAccessKey`) becomes the `accessKey` value; also copy
   `region` → `DefaultRegion` when the latter is empty, and force `AuthType = keys`.
   Mirrors `pkg/models/settings.go:38-55`.
2. **`ApplyDefaults`** — fills a curated list of zero-valued fields: `AuthType`
   defaults to `AWSAuthTypeDefault` (schema parity with the reference
   `aws_sdk_settings.json` pack and the backend `awsds.AuthTypeDefault` iota zero),
   and `Timeout`/`Retries`/`Pause` default to `"60"`/`"5"`/`"5"` matching
   `pkg/models/settings.go:75-83`.
3. **`Validate`** — enforces the runtime contract: known `AuthType` (rejects
   `grafana_assume_role` and other unknown values), `accessKey` + `secretKey` present
   for `keys` auth, and non-empty `DefaultRegion`. Errors are joined so callers see
   every problem at once. Error strings match the upstream (`"missing access key"`,
   `"missing secret key"`, `"missing region"`).

Everything is logged via `backend.Logger.FromContext(ctx)` with `datasource_uid`,
`datasource_name`, and `plugin` labels so log lines carry request context.

### Direct access to individual phases

`(*Config).ApplyDefaults()` and `(Config).Validate() error` are exported for callers
that want to compose them themselves (provisioning preview, schema-example round-trip,
tests that need to distinguish parse-level from policy-level errors). Skip them by never
calling `LoadConfig` in those flows — assemble a `Config` directly.

## Upstream findings

Potential bugs, misleading UX, and consistency issues discovered while researching
upstream. All preserved verbatim in the schema — the schema records what the plugin
**does**, not what it **should** do; these notes exist so reviewers can reproduce each
finding and decide separately whether to fix upstream.

1. **`secureJsonData.accessKey` under V1 was the SECRET key, not the ID.** V1 stored the
   Access Key **ID** in plain jsonData under the key `accessId`
   (`pkg/models/settings.go:29`) while the Secret Access Key went into
   `secureJsonData.accessKey` (see the V1 migration test at
   `pkg/models/settings_test.go:26-43` and the `migrateToSecureKey` helper at
   `pkg/models/settings.go:123-128`). The names are swapped from what a reader would
   expect. The V2 shape corrects this (`accessKey` is the ID, `secretKey` is the
   secret), so the confusion only affects pre-migration configs.
2. **The V1 warning banner never disappears until the user re-saves.** When
   `!isV2`, `src/components/ConfigEditor.tsx:53-59` shows an Alert warning the user
   that saving may drop `SecretKey` — but nothing rewrites the stored config on load;
   only `onOptionsChange` in the same file (via `migrateOptions`) updates the
   in-memory options, and only a subsequent save flushes them to storage. Users who
   read but never edit will keep the legacy shape indefinitely.
3. **`accessId` is a misleading storage key.** `pkg/models/settings.go:29` tags the
   field as `LegacyAccessKey` (Go name) with json tag `accessId`, but the value is the
   Access Key **ID**, not the AWS-terminology "access ID" (there is no such thing).
   Combined with finding #1, the naming is doubly confusing.
4. **`ConfigEditor.tsx:53-59` warning message contains ambiguous casing.** The banner
   refers to "SecretKey" and "AccessKey" as if they were literal storage keys, but
   secureJsonData uses `accessKey`/`secretKey` (lowercase initials) — the message
   would confuse anyone who tried to grep for the words in the codebase.
5. **`timeout`, `retries`, `pause` are strings for driver settings.** Everywhere else
   the plugin ecosystem uses `number` for timeouts; DynamoDB's backend
   (`pkg/models/settings.go:31-33`) declares them as `string` and hand-parses with
   `utils.ParseInt`. The choice appears historical and creates an unnecessary parsing
   surface — a `number` type would let the JSON layer do the validation.
6. **`grafana-dynamodb-datasource` cannot use Grafana Assume Role.** The
   `DS_TYPES_THAT_SUPPORT_TEMP_CREDS` allow-list
   (`grafana-aws-sdk-react/src/components/ConnectionConfig.tsx:18-28`) is hard-coded
   and lists only the OSS datasources — adding DynamoDB would require editing
   `@grafana/aws-sdk` upstream. Meanwhile the underlying `awsds.AuthType` would happily
   accept the value, so a provisioned config with `authType: "grafana_assume_role"`
   loads on the backend but the editor can't render it. This schema rejects that value
   in `allowedValues` to keep the surface honest.
7. **Deprecated `arn` auth value silently maps to `default`.** `awsds/settings.go:87-88`
   maps any unknown auth type string to `AuthTypeDefault`, comment: "For old 'arn'
   option". A provisioned config with `authType: "arn"` will load as if it were
   `default`, and there is no warning surfaced anywhere.
8. **Legacy `sharedCreds` auth value silently mapped to `credentials`.**
   `awsds/settings.go:75-78` folds `sharedCreds` and `credentials` onto the same enum.
   Not preserved in this schema's allowedValues (the reference pack also omits it); a
   datasource stored with `authType: "sharedCreds"` would fail the schema's
   `allowedValues` check but still load fine on the backend.
9. **Editor's fresh-datasource default is `keys`, not `default`.** The
   `useEffect` in `src/components/ConfigEditor.tsx:28-42` writes `authType:
   AwsAuthType.Keys` on an empty jsonData, even though the AWS pack default (and the
   backend iota-zero) is `Default`. So a datasource created via the UI starts life
   with `keys` selected, while one created via provisioning (with no `authType` set)
   falls through to `default`. The schema-side default here is `"default"` to match the
   reference AWS pack; provisioning tools that want to match the editor should
   explicitly write `"keys"`.
10. **The "Default Region" description has padded backticks.** `ConnectionConfig.tsx:377`
    reads `` `Specify the region, such as for US West (Oregon) use ` us-west-2 ` as
    the region.` `` — note the spaces around `us-west-2` inside the backticks.
    Preserved verbatim in the schema; would render inline code with padding in some
    Markdown renderers.
11. **`sessionToken` has no editor UI.** `grafana-aws-sdk-react` never exposes an input
    for `sessionToken` even though the backend reads it and it is required for
    temporary credentials. Users must provision it directly.
12. **No editor UI for `timeout`/`retries`/`pause`.** These driver settings have no
    editor input, so users can only configure them via provisioning; the defaults
    `"60"`/`"5"`/`"5"` are baked into the backend Load.

## Validation performed

- `dsconfig.ParseAndResolveSchemaJSON` + `Schema.Validate()` — passes.
- JSON Schema validation against [`dsconfig/schema.json`](../../dsconfig/schema.json)
  (draft-07, `additionalProperties: false`) — passes.
- `go test ./...` on this module — passes (schema bundle shape, `SchemaSpecHasNoSecureJSON`,
  `SecureValuesMatchLoadSettings`, `JSONDataMatchesStruct`, `JSONDataTypesMatchStruct`,
  `SchemaArtifactInSync`, `LoadConfig` including V1 migration parity with the upstream
  `pkg/models/settings_test.go` cases, `ApplyDefaults`, `Validate` per auth type).
- `settings.go` / `schema.go`: `go build`, `go vet`, `gofmt` — clean.
- `settings.ts`: `tsc --noEmit --strict` — clean.
