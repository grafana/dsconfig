# grafana-iot-twinmaker-datasource

Declarative configuration schema for the AWS IoT TwinMaker datasource, which
ships as a nested datasource inside the [AWS IoT TwinMaker App plugin](https://github.com/grafana/grafana-iot-twinmaker-app)
(app id `grafana-iot-twinmaker-app`). The datasource itself has plugin id
`grafana-iot-twinmaker-datasource`.

**This entry is scoped to the datasource, not the app.** The app plugin
`grafana-iot-twinmaker-app` bundles four panels (Alarm Configuration, Query
Editor, Scene Viewer, Video Player) plus this datasource in
`src/datasource/`. All persistent configuration lives on the datasource;
the app plugin.json only lists the includes.

## Upstream researched

- **App repo**: `github.com/grafana/grafana-iot-twinmaker-app`
- **Ref**: `main`
- **Commit SHA**: `a24885e092fb398b0ba34e324cccd2eaced8e6c2` (`Simplify CODEOWNERS to use a global owner (#787)`)

Every value in [`dsconfig.json`](dsconfig.json) — labels, placeholders,
tooltips, option labels/values, section titles, defaults, validations,
dependency and required-when expressions, storage keys, storage targets,
value types, group titles, and instructions — is traceable to a specific
`file:line` in the upstream plugin repo (or in the pinned `@grafana/aws-sdk`
version of the shared `ConnectionConfig` component) at this SHA.

To reproduce this research:

```bash
git clone --depth 20 https://github.com/grafana/grafana-iot-twinmaker-app
cd grafana-iot-twinmaker-app
git checkout a24885e092fb398b0ba34e324cccd2eaced8e6c2

# The AWS auth surface is rendered by @grafana/aws-sdk's ConnectionConfig.
# The plugin's package.json pins this to 0.8.3 (note: this is older than
# the version used by sibling AWS entries like sitewise or x-ray).
git clone https://github.com/grafana/grafana-aws-sdk-react
cd grafana-aws-sdk-react
git checkout v0.8.3
```

If upstream `main` has advanced past the pinned SHA, re-diff the sources
listed under [Sources researched](#sources-researched) and reconcile the
schema before merging.

## Files

| File | Purpose |
| --- | --- |
| [`dsconfig.json`](dsconfig.json) | dsconfig v1 schema — single source of truth for all config fields, groups, and instructions |
| [`settings.ts`](settings.ts) | TypeScript models: `RootConfig` (blank), `JsonDataConfig`, `SecureJsonDataConfig` |
| [`settings.go`](settings.go) | Go `Config` model (flat: jsonData fields + `DecryptedSecureJSONData`), `PluginID`, `AWSAuthType` typed constants, `SecureJsonDataKey` typed constants, and the `LoadConfig` utility |
| [`schema.go`](schema.go) | k8s-style SDK plugin schema: embeds `dsconfig.json`, converts it via `dsconfig.NewSDKSchema`, and defines `SettingsExamples` for each AWS auth provider plus AssumeRole/write-permissions/legacy variants |
| [`settings_test.go`](settings_test.go) | Table tests for `LoadConfig`, `ApplyDefaults`, and `Validate` |
| [`conformance_test.go`](conformance_test.go) | Runs `schema.RunPluginTests` — the shared dsconfig conformance suite — against `dsconfig.json` |
| [`schema.gen.json`](schema.gen.json), [`settings.gen.json`](settings.gen.json), [`settings.examples.gen.json`](settings.examples.gen.json) | Committed schema artifacts (regenerate with `go generate ./...` inside this directory; the `SchemaArtifactInSync` conformance subtest guards drift) |

There is no per-entry `go.mod` — every registry entry is a subpackage of the
shared [`registry/`](..) module (`github.com/grafana/dsconfig/registry`).

## App plugin vs datasource plugin

The app plugin.json (`src/plugin.json`) has `type: app`,
`id: grafana-iot-twinmaker-app`, and `name: AWS IoT TwinMaker App`. It has
no persistent configuration fields of its own; it exists to bundle the
datasource and four panels and to set `autoEnabled: true`.

The datasource plugin.json (`src/datasource/plugin.json`) has
`type: datasource`, `id: grafana-iot-twinmaker-datasource`, and
`name: AWS IoT TwinMaker`. All persistent configuration for querying IoT
TwinMaker lives here — this schema models that datasource.

## Sources researched

Every source below was read at the pinned upstream SHA
(`a24885e092fb398b0ba34e324cccd2eaced8e6c2`) or, for `@grafana/aws-sdk`, at
the exact version pinned in the plugin's `package.json` (`0.8.3`).

### App / plugin repo (`github.com/grafana/grafana-iot-twinmaker-app@a24885e`)

| File | What was read |
| --- | --- |
| `src/plugin.json:1-56` | App plugin.json: `type: app`, `id: grafana-iot-twinmaker-app`, `name: AWS IoT TwinMaker App`, `info.links[0].url: https://github.com/grafana/grafana-iot-twinmaker-app`, and the `includes[]` list of the nested datasource + four panels |
| `src/datasource/plugin.json:1-24` | Datasource plugin.json: `type: datasource`, `id: grafana-iot-twinmaker-datasource`, `name: AWS IoT TwinMaker`, `backend: true`, `alerting: true`. This is the pluginType/pluginName in this dsconfig entry |
| `src/datasource/components/ConfigEditor.tsx:23-166` | Top-level editor: renders `<ConnectionConfig ... standardRegions={standardRegions}>`, an assumeRoleArn Alert, `<SecureSocksProxySettings>` (excluded per AGENTS.md), a `Twinmaker Settings` `ConfigSection` with a Workspace Select (workspaceId), an alarm-config Switch, and a conditional Assume Role ARN Write Input |
| `src/datasource/components/ConfigEditor.tsx:31-36` | On-mount write: if `defaultRegion` is empty, set it to `us-east-1` |
| `src/datasource/components/ConfigEditor.tsx:38-49` | Save-state tracking used only to gate workspace lazy-load — not persisted |
| `src/datasource/components/ConfigEditor.tsx:102-114` | Red `<Alert title="Assume Role ARN" severity="error">` shown when `!assumeRoleArn`, directing users to the AWS IoT TwinMaker dashboard IAM role docs |
| `src/datasource/components/ConfigEditor.tsx:120-140` | Workspace `<Select label="Workspace">` — `allowCustomValue`, lazy-loads options via `datasource.info.listWorkspaces()` once saved; placeholder `Select a workspace`; writes `jsonData.workspaceId` |
| `src/datasource/components/ConfigEditor.tsx:141-147` | Switch `<Field label="Define write permissions for Alarm Configuration Panel">`. Not stored — React state only. Toggling off clears `assumeRoleArnWriter` (line 92) |
| `src/datasource/components/ConfigEditor.tsx:149-162` | Conditional `<Field label="Assume Role ARN Write" description="Specify the ARN of a role to assume when writing property values in IoT TwinMaker">`, placeholder `arn:aws:iam:*`, writes `jsonData.assumeRoleArnWriter` |
| `src/datasource/regions.ts:3-15` | `standardRegions`: `ap-south-1`, `ap-northeast-1`, `ap-northeast-2`, `ap-southeast-1`, `ap-southeast-2`, `eu-central-1`, `eu-west-1`, `us-east-1`, `us-west-2`, `us-gov-west-1`, `cn-north-1`. (No `us-east-2`, no `ca-central-1`, no `Edge` sentinel.) |
| `src/datasource/types.ts:24-27` | `TwinMakerDataSourceOptions extends AwsAuthDataSourceJsonData` adds `workspaceId?`, `assumeRoleArnWriter?` |
| `src/datasource/types.ts:28-31` | `TwinMakerSecureJsonData extends AwsAuthDataSourceSecureJsonData` adds an unused `anythingSecure?` placeholder — kept out of this schema because no editor writes it and no backend reads it |
| `pkg/models/settings.go:12-18` | Backend `TwinMakerDataSourceSetting`: embeds `awsds.AWSDatasourceSettings`, adds `ProxyOptions`, `AssumeRoleARNWriter string`json:"assumeRoleArnWriter"``, `WorkspaceID string`json:"workspaceId"``, `UID string`json:"uid"`` (UID is set from `config.UID` at runtime, not from jsonData) |
| `pkg/models/settings.go:20-40` | `Load`: unmarshals JSONData only when `len(JSONData) > 1`; substitutes `DefaultRegion` for `Region` when Region is empty/`"default"`; falls back to `us-east-1` when both are empty; copies `accessKey` and `secretKey` (but NOT `sessionToken`) from decrypted secure data |
| `pkg/models/settings.go:42-45` | `Validate`: is a no-op — the runtime contract is enforced by CheckHealth instead |
| `pkg/models/settings.go:47-66` | `ToAWSDatasourceSettings` / `ToAWSDatasourceSettingsWriter`: the writer variant swaps `AssumeRoleARN` for `AssumeRoleARNWriter` before building the AWS client |
| `pkg/plugin/datasource.go:171-184` | `CheckHealth`: fails with `Missing WorkspaceID configuration` when `workspaceId` is empty and `Assume Role ARN is required` when `assumeRoleArn` is empty. This is the backend contract we encode as `requiredWhen: "true"` |
| `pkg/plugin/datasource.go:225-240` | `CheckHealth` also probes the writer session when `assumeRoleArnWriter` is set — confirms the field is optional (only the presence check is triggered) |
| `package.json` | External component versions: `@grafana/aws-sdk@0.8.3`, `@grafana/plugin-ui@^0.4.11`, `@grafana/ui@^10.4.0`, `@grafana/data@^10.4.0`. Note the older aws-sdk-react pin vs sitewise (0.10.2) and x-ray (0.10.2) |
| `go.mod` | Backend deps: `github.com/grafana/grafana-aws-sdk v1.4.3`, `github.com/grafana/grafana-plugin-sdk-go v0.290.1` |

### External editor components

Read at the exact versions pinned in the plugin's `package.json`.

| Component | Version | Source consulted | What was read |
| --- | --- | --- | --- |
| `ConnectionConfig`, `ConnectionConfigProps`, `Divider` | `@grafana/aws-sdk@0.8.3` | grafana/grafana-aws-sdk-react tag `v0.8.3`, `src/components/ConnectionConfig.tsx`, `src/providers.ts`, `src/types.ts`, `src/regions.ts` | Every AWS field's label, placeholder, description, and conditional render; the Select option labels for `authType`; the `DS_TYPES_THAT_SUPPORT_TEMP_CREDS` allow-list that filters `grafana_assume_role` out for TwinMaker |
| `ConfigSection` | `@grafana/plugin-ui@^0.4.11` | Editor layout — no storage fields |
| `SecureSocksProxySettings` (excluded) | `@grafana/ui@^10.4.0` | Writes `jsonData.enableSecureSocksProxy`; deliberately excluded from this entry |
| `Alert`, `Field`, `Input`, `Select`, `Switch` | `@grafana/ui@^10.4.0` | Prop names (`label`, `placeholder`, `description`, `value`, `onChange`, `id`, `htmlFor`) so we knew which UI attributes to record |
| `updateDatasourcePluginJsonDataOption`, `onUpdateDatasourceJsonDataOption` | `@grafana/data@^10.4.0` | Storage-key semantics of the update helpers used by ConfigEditor and by ConnectionConfig |

### Backend Go dependency (`grafana-aws-sdk`)

| File | What was read |
| --- | --- |
| `pkg/awsds/settings.go` (`v1.4.3`) | `AuthType` int enum + custom `MarshalJSON`/`UnmarshalJSON` — the storage⇆enum mapping (including legacy `arn`→`default` and `sharedCreds`→`credentials`) that we surface as `AWSAuthType` string constants |
| `pkg/awsds/settings.go` (`v1.4.3`) | `AWSDatasourceSettings` struct with the AWS-shared fields; note `AssumeRoleARN string`json:"assumeRoleARN"`` (uppercase `ARN`) while the frontend writes camelCase `assumeRoleArn` — Go's case-insensitive Unmarshal makes both work |

## Field provenance

Every dsconfig field, traced from its schema `id` to the upstream `file:line`
where each of its label, placeholder, tooltip, default, storage key, and
value type is defined.

| Schema `id` | Storage key | Target | Label source | Placeholder / options / default source | Value type source | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `jsonData_authType` | `authType` | `jsonData` | `ConnectionConfig.tsx` (`<Field label="Authentication Provider">`) | Options from `@grafana/aws-sdk@0.8.3` `src/providers.ts` (`awsAuthProviderOptions`); description from ConnectionConfig; default `awsds/settings.go` (`AuthTypeDefault` iota) → `"default"` | `AWSDatasourceSettings.AuthType` `awsds/settings.go` | Role `auth.discriminator`; `allowedValues` includes legacy `arn` and (schema-only) `grafana_assume_role` — the latter is filtered out of the editor by `DS_TYPES_THAT_SUPPORT_TEMP_CREDS` |
| `jsonData_profile` | `profile` | `jsonData` | ConnectionConfig `<Field label="Credentials Profile Name">` | Placeholder `"default"`; description from ConnectionConfig | `AWSDatasourceSettings.Profile` `awsds/settings.go` | `dependsOn: jsonData_authType == 'credentials'` |
| `secureJsonData_accessKey` | `accessKey` | `secureJsonData` | ConnectionConfig `<Field label="Access Key ID">` | — | `AWSDatasourceSettings.AccessKey` `awsds/settings.go` | Role `auth.aws.accessKeyId` |
| `secureJsonData_secretKey` | `secretKey` | `secureJsonData` | ConnectionConfig `<Field label="Secret Access Key">` | — | `AWSDatasourceSettings.SecretKey` `awsds/settings.go` | Role `auth.aws.secretAccessKey` |
| `jsonData_assumeRoleArn` | `assumeRoleArn` | `jsonData` | ConnectionConfig `<Field label="Assume Role ARN">` | Placeholder `arn:aws:iam:*`; description from ConnectionConfig (multi-line as in source) | `AWSDatasourceSettings.AssumeRoleARN` `awsds/settings.go` (backend json tag `assumeRoleARN`) | `requiredWhen: "true"` per CheckHealth; frontend writes lowercase `assumeRoleArn` — case-insensitive Unmarshal rescue |
| `jsonData_externalId` | `externalId` | `jsonData` | ConnectionConfig `<Field label="External ID">` | Placeholder `External ID`; description from ConnectionConfig | `AWSDatasourceSettings.ExternalID` `awsds/settings.go` | `dependsOn: jsonData_authType != 'grafana_assume_role'` |
| `jsonData_endpoint` | `endpoint` | `jsonData` | ConnectionConfig `<Field label="Endpoint">` | Placeholder `https://{service}.{region}.amazonaws.com`; description from ConnectionConfig | `AWSDatasourceSettings.Endpoint` `awsds/settings.go` | `dependsOn: jsonData_authType != 'grafana_assume_role'` |
| `jsonData_defaultRegion` | `defaultRegion` | `jsonData` | ConnectionConfig `<Field label="Default Region">` | Options inlined from `src/datasource/regions.ts:3-15` (twinmaker-specific 11-region list); description from ConnectionConfig (verbatim including the space-padded backticks around `us-west-2`); default `us-east-1` (`ConfigEditor.tsx:31-36`, `settings.go:33-35`) | `AWSDatasourceSettings.DefaultRegion` `awsds/settings.go` | `<Select ... allowCustomValue={true}>` |
| `jsonData_workspaceId` | `workspaceId` | `jsonData` | `ConfigEditor.tsx:120` (`<Field label="Workspace">`) | Placeholder `Select a workspace` | `TwinMakerDataSourceSetting.WorkspaceID` `pkg/models/settings.go:16` | `requiredWhen: "true"` per CheckHealth |
| `virtual_alarmConfigEnabled` | (virtual) | — | `ConfigEditor.tsx:141` (`<Field htmlFor="alarmConfigChecked" label="Define write permissions for Alarm Configuration Panel">`) | — | React state (`useState`, `ConfigEditor.tsx:24`) | `kind: virtual`; `storage.computed.read` derives from `assumeRoleArnWriter != ''`; `effects` clear the write ARN when toggled off |
| `jsonData_assumeRoleArnWriter` | `assumeRoleArnWriter` | `jsonData` | `ConfigEditor.tsx:152` (`<Field label="Assume Role ARN Write" description="Specify the ARN of a role to assume when writing property values in IoT TwinMaker">`) | Placeholder `arn:aws:iam:*` | `TwinMakerDataSourceSetting.AssumeRoleARNWriter` `pkg/models/settings.go:15` | `dependsOn: virtual_alarmConfigEnabled == true`; `tags: [managed-by:virtual_alarmConfigEnabled]` |

## Field inventory summary

| Schema field | Storage key | Target | Editor label | Read by backend? |
| --- | --- | --- | --- | --- |
| `jsonData_authType` | `authType` | `jsonData` | Authentication Provider | Yes |
| `jsonData_profile` | `profile` | `jsonData` | Credentials Profile Name | Yes |
| `secureJsonData_accessKey` | `accessKey` | `secureJsonData` | Access Key ID | Yes |
| `secureJsonData_secretKey` | `secretKey` | `secureJsonData` | Secret Access Key | Yes |
| `jsonData_assumeRoleArn` | `assumeRoleArn` | `jsonData` | Assume Role ARN | Yes (required by CheckHealth) |
| `jsonData_externalId` | `externalId` | `jsonData` | External ID | Yes |
| `jsonData_endpoint` | `endpoint` | `jsonData` | Endpoint | Yes |
| `jsonData_defaultRegion` | `defaultRegion` | `jsonData` | Default Region | Yes |
| `jsonData_workspaceId` | `workspaceId` | `jsonData` | Workspace | Yes (required by CheckHealth) |
| `virtual_alarmConfigEnabled` | (virtual) | — | Define write permissions for Alarm Configuration Panel | No — editor-local state |
| `jsonData_assumeRoleArnWriter` | `assumeRoleArnWriter` | `jsonData` | Assume Role ARN Write | Yes (optional) |

### Frontend-only settings

None. Every editor-writable field is consumed by the backend.

### Backend-only settings

None. Unlike sibling AWS entries, TwinMaker's backend Load does not copy
`sessionToken` from decrypted secure data (see the discrepancies section
below), so we do not include it as a schema field.

### Fields excluded from this entry

- **`sessionToken`** — declared on the AWS shared secure shape
  (`AwsAuthDataSourceSecureJsonData`) but not read by TwinMaker's backend
  (`pkg/models/settings.go:37-38`). Included as a discrepancy note (see
  below) instead of a schema field.
- **AWS proxy fields** (`proxyType`, `proxyUrl`, `proxyUsername`,
  `proxyPassword`) — TwinMaker's ConfigEditor calls `<ConnectionConfig>`
  without `showHttpProxySettings`, so the proxy subsection is not part of
  TwinMaker's declared editor surface.
- **Secure Socks Proxy** (`jsonData.enableSecureSocksProxy`) — excluded
  per AGENTS.md.
- **`region`** — declared on `AwsAuthDataSourceJsonData` but the frontend
  never writes it; the backend mirrors `defaultRegion` into it at load
  time (`pkg/models/settings.go:30-32`) and finally falls back to
  `us-east-1` (`:33-35`). Not a stored config.
- **`anythingSecure`** — a placeholder secure field declared on
  `TwinMakerSecureJsonData` but never written by any editor and never read
  by any backend code path.
- **`uid`** — appears on the Go `TwinMakerDataSourceSetting` struct with a
  json tag, but is populated from `backend.DataSourceInstanceSettings.UID`
  at runtime (`pkg/models/settings.go:28`), not from stored jsonData.

## Where the types are defined

The TwinMaker configuration types are spread across the plugin and its
dependencies. Some fields and base types come from libraries/SDKs rather
than the plugin itself.

### Frontend (TypeScript)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `TwinMakerDataSourceOptions` (jsonData), `TwinMakerSecureJsonData` | `src/datasource/types.ts:24-31` | plugin ([grafana/grafana-iot-twinmaker-app](https://github.com/grafana/grafana-iot-twinmaker-app)) |
| `AwsAuthType`, `AwsAuthDataSourceJsonData` (base of `TwinMakerDataSourceOptions`), `AwsAuthDataSourceSecureJsonData` (base of `TwinMakerSecureJsonData`) | `src/types.ts:3-28` | `@grafana/aws-sdk` `0.8.3` (grafana/grafana-aws-sdk-react `v0.8.3`) |
| `awsAuthProviderOptions` (Select options for `authType`) | `src/providers.ts:4-25` | `@grafana/aws-sdk` `0.8.3` |
| `standardRegions` (11-region twinmaker-specific list) | `src/datasource/regions.ts:3-15` | plugin |
| `DataSourceJsonData` (base type of `AwsAuthDataSourceJsonData`) | `packages/grafana-data/src/types/datasource.ts` | `@grafana/data` `^10.4.0` |

### Backend (Go)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `TwinMakerDataSourceSetting`, `Load`, `Validate`, `ToAWSDatasourceSettings`, `ToAWSDatasourceSettingsWriter` | `pkg/models/settings.go:12-67` | plugin ([grafana/grafana-iot-twinmaker-app](https://github.com/grafana/grafana-iot-twinmaker-app)) |
| `AWSDatasourceSettings` (embedded base of `TwinMakerDataSourceSetting`), `AuthType` int enum + custom Marshal/Unmarshal | `pkg/awsds/settings.go` | `github.com/grafana/grafana-aws-sdk` `v1.4.3` |
| `backend.DataSourceInstanceSettings` (carries `JSONData`, `DecryptedSecureJSONData`, `UID`) | `backend/common.go` | `github.com/grafana/grafana-plugin-sdk-go` `v0.290.1` |
| TwinMaker AWS SDK client build-up (`iottwinmaker.NewFromConfig(...)`) | `pkg/plugin/twinmaker/client.go` | plugin (uses `github.com/aws/aws-sdk-go-v2/service/iottwinmaker`) |

The models in this entry flatten that spread into a single Go `Config`
type (jsonData fields + `DecryptedSecureJSONData`) plus a
`SecureJsonDataKey` typed constant list. `settings.ts` keeps the three
canonical TypeScript types (`RootConfig`, `JsonDataConfig`,
`SecureJsonDataConfig`). `AWSAuthType` constants in `settings.go` mirror
the string forms `awsds.AuthType` (Un)Marshals, without carrying the
int-enum + custom-JSON machinery.

## Modeling decisions

- **Datasource-scoped, not app-scoped.** The `grafana-iot-twinmaker-app`
  package is a Grafana app plugin that bundles panels and a datasource;
  its own plugin.json has no persistent configuration. All configuration
  relevant to querying IoT TwinMaker lives on the nested datasource at
  `src/datasource/plugin.json` (id `grafana-iot-twinmaker-datasource`),
  which is what this schema targets. The registry directory name and
  `pluginType` therefore match the datasource id.
- **Single, flat Go `Config`.** The upstream `TwinMakerDataSourceSetting`
  embeds `awsds.AWSDatasourceSettings`. We flatten those into one struct
  to match the pattern used by other AWS registry entries (see
  `registry/grafana-iot-sitewise-datasource`,
  `requiredWhen: "true"` on `jsonData_assumeRoleArn` and `jsonData_workspaceId`)
  and to avoid pulling `grafana-aws-sdk` into the shared registry `go.mod`.
- **`AssumeRoleARN` field vs `assumeRoleArn` tag** — the Go field name
  mirrors the upstream `awsds.AWSDatasourceSettings.AssumeRoleARN`, but the
  json tag mirrors what the frontend writes (`assumeRoleArn`).
  Case-insensitive `encoding/json` rescue.
- **`assumeRoleArn` is `requiredWhen: "true"`** even though the editor
  labels the field "Optional". This reflects the backend contract enforced
  by `CheckHealth` (`pkg/plugin/datasource.go:179-184`).
- **`workspaceId` is `requiredWhen: "true"`** for the same reason
  (`pkg/plugin/datasource.go:172-177`).
- **`alarmConfigEnabled` modelled as a virtual field.** The switch
  ("Define write permissions for Alarm Configuration Panel") is React state,
  not stored config. Modelled as `kind: virtual` with:
  - `storage.computed.read: "jsonData.assumeRoleArnWriter != null && jsonData.assumeRoleArnWriter != ''"`
    to derive the switch state on load.
  - `effects` writing `jsonData_assumeRoleArnWriter: ''` when the switch is
    toggled off, matching `ConfigEditor.tsx:89-94`.
- **`jsonData_assumeRoleArnWriter` `dependsOn: virtual_alarmConfigEnabled == true`**
  and tagged `managed-by:virtual_alarmConfigEnabled` per AGENTS.md.
- **`grafana_assume_role` in `allowedValues` but not in UI options.**
  `@grafana/aws-sdk@0.8.3`'s ConnectionConfig restricts the provider to
  `DS_TYPES_THAT_SUPPORT_TEMP_CREDS = ['cloudwatch',
  'grafana-athena-datasource', 'grafana-amazonprometheus-datasource']`,
  which does not include TwinMaker. The editor will never show this
  option, so we leave it out of the UI options list. Storage-side the
  value remains valid, so `allowedValues` still lists it.
- **`arn` legacy value kept in `allowedValues`, not in UI options.**
  `AwsAuthType.ARN` is deprecated (`grafana-aws-sdk-react/src/types.ts:11`)
  and does not appear in `awsAuthProviderOptions`. Stored datasources may
  still carry it.
- **Region options inlined** — TwinMaker's supportedRegions list
  (`src/datasource/regions.ts`) is short and plugin-specific, so inlining
  the options documents the possible values explicitly. `allowCustom: true`
  because the plugin passes `allowCustomValue={true}` to ConnectionConfig.
- **`defaultRegion` default `us-east-1`.** Unlike sitewise/x-ray, TwinMaker
  writes `us-east-1` on editor mount if empty (`ConfigEditor.tsx:31-36`)
  and repeats the fallback in the backend Load (`settings.go:33-35`).
- **`SecureJsonDataConfig` is a key list with only `accessKey` and
  `secretKey`.** Secure values are write-only. `sessionToken` — the third
  standard AWS secret — is not present because TwinMaker's backend Load
  does not read it (see the first discrepancy below).
- **AWS proxy fields excluded** — TwinMaker does not pass
  `showHttpProxySettings` to `ConnectionConfig`, so the proxy fields are
  not part of the editor surface.

## SDK plugin schema and k8s-style examples (`schema.go`)

`NewSchema()` assembles the `grafana-plugin-sdk-go`
`pluginschema.PluginSchema` bundle (the k8s-style schema Grafana's
datasource API server serves as `{apiVersion}.json`, `v0alpha1` today) from
the embedded `dsconfig.json`.

`SettingsExamples()` provides the default configuration plus one k8s-style
example per editor-selectable AWS auth provider, an External-ID variant, a
write-permissions variant, and a legacy `arn` example:

| Example | Auth | Extras | `secureJsonData` |
| --- | --- | --- | --- |
| `""` (default) | AWS SDK Default | `defaultRegion=us-east-1` (no workspaceId/assumeRoleArn) | `accessKey` (empty) |
| `awsSdkDefault` | AWS SDK Default | `defaultRegion`, `workspaceId`, `assumeRoleArn` | `accessKey` (empty) |
| `accessAndSecretKey` | Access & secret key | `defaultRegion`, `workspaceId`, `assumeRoleArn` | `accessKey`, `secretKey` |
| `credentialsFile` | Credentials file | `profile`, `defaultRegion`, `workspaceId`, `assumeRoleArn` | `accessKey` (empty) |
| `workspaceIamRole` | Workspace IAM Role | `defaultRegion`, `workspaceId`, `assumeRoleArn` | `accessKey` (empty) |
| `withExternalId` | AWS SDK Default + STS External ID | `defaultRegion`, `workspaceId`, `assumeRoleArn`, `externalId` | `accessKey` (empty) |
| `withAlarmWriteRole` | AWS SDK Default | `defaultRegion`, `workspaceId`, `assumeRoleArn`, `assumeRoleArnWriter` | `accessKey` (empty) |
| `legacyArnAuthType` | `arn` (legacy — backend maps to `default`) | `defaultRegion`, `workspaceId`, `assumeRoleArn` | `accessKey` (empty) |

Every example carries at least one `secureJsonData` placeholder as
required by the conformance suite.

## `LoadConfig` utility (`settings.go`)

`LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error)`
runs the full three-phase load flow on a datasource instance's settings and
returns a fully-defaulted, validated `Config`:

1. **Parse** — unmarshal `settings.JSONData` into `Config` (only when the
   payload has more than one byte, matching
   `pkg/models/settings.go:21`) and copy the plugin's decrypted secrets
   (`accessKey`, `secretKey`) into `DecryptedSecureJSONData`.
   `sessionToken` is deliberately not copied — see the discrepancy note.
2. **`ApplyDefaults`** — fills two curated defaults: `AuthType` defaults to
   `AWSAuthTypeDefault` (matching the reference AWS pack + backend
   `awsds.AuthTypeDefault` iota zero); `DefaultRegion` defaults to
   `us-east-1` (matching the editor's on-mount write + the backend
   fallback). `WorkspaceID`, `AssumeRoleARN`, and `AssumeRoleARNWriter`
   have no defaults.
3. **`Validate`** — enforces the runtime contract: known `AuthType`,
   `accessKey`+`secretKey` present for `keys` auth, `workspaceId`
   non-empty, and `assumeRoleArn` non-empty. Mirrors the two rejection
   paths in `pkg/plugin/datasource.go:172-184`. Errors are joined so
   callers see every problem at once.

Everything is logged via `backend.Logger.FromContext(ctx)` with
`datasource_uid`, `datasource_name`, and `plugin` labels so log lines
carry request context.

### Direct access to individual phases

`(*Config).ApplyDefaults()` and `(Config).Validate() error` are exported
for callers that want to compose them themselves (provisioning preview,
schema-example round-trip, tests that need to distinguish parse-level from
policy-level errors). Skip them by never calling `LoadConfig` in those
flows — assemble a `Config` directly.

## Upstream findings

Potential bugs, misleading UX, and consistency issues discovered while
researching upstream. All preserved verbatim in the schema — the schema
records what the plugin **does**, not what it **should** do; these notes
exist so reviewers can reproduce each finding and decide separately
whether to fix upstream.

1. **`sessionToken` is silently dropped by the backend Load.**
   `pkg/models/settings.go:37-38` reads only `accessKey` and `secretKey`
   from decrypted secure data. `awsds.AWSDatasourceSettings.SessionToken`
   remains empty, and `ToAWSDatasourceSettings` (`:47-61`) passes that
   empty value to the AWS SDK. Any datasource provisioned with
   `secureJsonData.sessionToken` set will silently drop it — temporary
   STS credentials scoped only to `accessKey`/`secretKey`/`sessionToken`
   cannot be used to authenticate against TwinMaker. This is not the case
   for sibling AWS datasources (sitewise, x-ray) that route through
   `awsds.Load`.
2. **`assumeRoleArn` label says "Optional" but is required at runtime.**
   `ConnectionConfig.tsx:242-245` prefixes the description with
   `"Optional."`, but TwinMaker's `CheckHealth` fails outright when the
   field is empty (`pkg/plugin/datasource.go:179-184`) and the editor
   renders a red Alert with docs link when the field is empty
   (`ConfigEditor.tsx:102-114`). The Alert is the "real" UX; the
   ConnectionConfig description is misleading.
3. **`AssumeRoleARN` json tag disagrees with the frontend's
   `assumeRoleArn`.** `awsds/settings.go` uses `` json:"assumeRoleARN" ``
   (uppercase RN), but the frontend type
   (`grafana-aws-sdk-react/src/types.ts`) and every `onChange` in
   `ConnectionConfig` write `assumeRoleArn` (lowercase arn). Both work
   thanks to Go's case-insensitive `encoding/json`.
4. **Deprecated `arn` auth value silently maps to `default`.**
   `awsds/settings.go` maps any unknown auth type string to
   `AuthTypeDefault`, comment: "For old 'arn' option". A provisioned
   config with `authType: "arn"` will load as if it were `default`, with
   no warning surfaced.
5. **`Grafana Assume Role` is unreachable in the editor for TwinMaker.**
   `@grafana/aws-sdk@0.8.3` restricts this provider to
   `DS_TYPES_THAT_SUPPORT_TEMP_CREDS = ['cloudwatch',
   'grafana-athena-datasource', 'grafana-amazonprometheus-datasource']`
   (ConnectionConfig.tsx:17-21). TwinMaker is not on that list; users can
   never select this provider via the editor even with
   `awsDatasourcesTempCredentials` toggled on. Storage-wise the value is
   still valid (and would be accepted by the AWS backend), so we keep it
   in `allowedValues`.
6. **`useEffectOnce` writes `defaultRegion = 'us-east-1'` on mount.**
   `ConfigEditor.tsx:31-36` mutates state via
   `updateDatasourcePluginJsonDataOption` before the user has interacted
   with the editor. That marks the datasource as dirty and shows the Save
   button — the user hasn't changed anything but the form appears
   modified. Harmless but surprising.
7. **`Load` accepts an empty `JSONData` slice but rejects a one-byte
   slice.** `pkg/models/settings.go:21` guards the unmarshal with
   `if len(config.JSONData) > 1`, so a literal payload of one character
   (e.g. `{`) is treated as if there were no settings at all. `LoadConfig`
   mirrors this exactly for parity with upstream.
8. **`Region` fallback trio.** `pkg/models/settings.go:30-35` first
   substitutes `DefaultRegion` for `Region` (when `Region == "default" ||
   Region == ""`) and then falls back to `us-east-1` when both are still
   empty. The frontend never writes `Region`, so in practice this always
   runs and always uses `DefaultRegion` (or the fallback).
9. **`anythingSecure` in the secure shape is dead code.**
   `TwinMakerSecureJsonData` (`src/datasource/types.ts:28-31`) declares
   an `anythingSecure?` field. No editor writes it and no backend reads
   it — it's a placeholder that never got fleshed out. Excluded from this
   schema.
10. **`uid` is a Go json field but never a stored one.**
    `TwinMakerDataSourceSetting.UID` is tagged
    `` json:"uid" ``, but `Load` populates it from
    `backend.DataSourceInstanceSettings.UID`
    (`pkg/models/settings.go:28`), not from stored jsonData. If a
    provisioning payload set `jsonData.uid`, the unmarshal would set it
    briefly before being overwritten.
11. **Backend `Validate` is a no-op.** `pkg/models/settings.go:42-45`
    returns `nil` unconditionally. The runtime contract is enforced only
    by CheckHealth, so a datasource can be saved (via provisioning or the
    HTTP API) with no `workspaceId` and no `assumeRoleArn` and only fail
    when the health probe runs. `LoadConfig` in this entry surfaces those
    requirements at parse-time instead.

## Validation performed

- `dsconfig.ParseAndResolveSchemaJSON` + `Schema.Validate()` — passes.
- JSON Schema validation against
  [`dsconfig/schema.json`](../../dsconfig/schema.json) (draft-07,
  `additionalProperties: false`) — passes via the shared conformance
  suite.
- `go test ./...` on this module — passes (schema bundle shape,
  `SchemaSpecHasNoSecureJSON`, `SecureValuesMatchLoadSettings`,
  `JSONDataMatchesStruct`, `JSONDataTypesMatchStruct`,
  `SchemaArtifactInSync`, `LoadConfig`, `ApplyDefaults`, `Validate` per
  auth type and per required-field path).
- `settings.go` / `schema.go`: `go build`, `go vet`, `gofmt` — clean.
- `settings.ts`: intended for `tsc --noEmit --strict` (TypeScript 5).
- All 28 pre-existing registry entries still `go test` clean.
