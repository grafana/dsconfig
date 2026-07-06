# grafana-iot-sitewise-datasource

Declarative configuration schema for the [AWS IoT SiteWise datasource plugin](https://github.com/grafana/iot-sitewise-datasource) (`grafana-iot-sitewise-datasource`).

## Upstream researched

- **Repo**: `github.com/grafana/iot-sitewise-datasource`
- **Ref**: `main`
- **Commit SHA**: `5fed0c9f84c5e042d2a98f67d4bf6cb01b4ccd2e` (`docs: add signed commits requirement to CONTRIBUTING.md (#810)`)

Every value in [`dsconfig.json`](dsconfig.json) — labels, placeholders, tooltips, option
labels/values, section titles, defaults, validations, dependency and required-when
expressions, storage keys, storage targets, value types, group titles, and
instructions — is traceable to a specific `file:line` in the upstream plugin repo (or
in the pinned `@grafana/aws-sdk` version of the shared `ConnectionConfig` component)
at this SHA. See [Field provenance](#field-provenance).

To reproduce this research:

```bash
git clone --depth 20 https://github.com/grafana/iot-sitewise-datasource
cd iot-sitewise-datasource
git checkout 5fed0c9f84c5e042d2a98f67d4bf6cb01b4ccd2e

# The AWS auth surface is rendered by @grafana/aws-sdk's ConnectionConfig. Pin
# to the version iot-sitewise-datasource's package.json uses (currently 0.10.2):
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
| [`settings.go`](settings.go) | Go `Config` model (flat: jsonData fields + `DecryptedSecureJSONData`), `PluginID`, `AWSAuthType` + `EdgeAuthMode` typed constants, `SecureJsonDataKey` typed constants, and the `LoadConfig` utility |
| [`schema.go`](schema.go) | k8s-style SDK plugin schema: embeds `dsconfig.json`, converts it via `dsconfig.NewSDKSchema`, and defines `SettingsExamples` for each AWS auth provider and Edge Kernel mode |
| [`settings_test.go`](settings_test.go) | Table tests for `LoadConfig`, `ApplyDefaults`, and `Validate` |
| [`conformance_test.go`](conformance_test.go) | Runs `schema.RunPluginTests` — the shared dsconfig conformance suite — against `dsconfig.json` |
| [`schema.gen.json`](schema.gen.json), [`settings.gen.json`](settings.gen.json), [`settings.examples.gen.json`](settings.examples.gen.json) | Committed schema artifacts (regenerate with `go generate ./...` inside this directory; the `SchemaArtifactInSync` conformance subtest guards drift) |

There is no per-entry `go.mod` — every registry entry is a subpackage of the shared
[`registry/`](..) module (`github.com/grafana/dsconfig/registry`).

## Sources researched

Every source below was read at the pinned upstream SHA
(`5fed0c9f84c5e042d2a98f67d4bf6cb01b4ccd2e`) or, for `@grafana/aws-sdk`, at the exact
version pinned in the plugin's `package.json` (`0.10.2`).

### Plugin repo (`github.com/grafana/iot-sitewise-datasource@5fed0c9`)

| File | What was read |
| --- | --- |
| `src/plugin.json:1-43` | `pluginType` (`id: grafana-iot-sitewise-datasource`), `pluginName` (`name: AWS IoT SiteWise`), docs URL (`info.links[0].url: https://aws.amazon.com/iot-sitewise/` — plugin catalog URL used as `docURL`) |
| `src/components/ConfigEditor.tsx:1-42` | Top-level branch: if `jsonData.defaultRegion === 'Edge'` → render `EdgeConfig`, otherwise render `<ConnectionConfig>` with the sitewise-specific `standardRegions` list |
| `src/components/ConfigEditor.tsx:44-190` | `EdgeConfig`: renders a custom Endpoint + Default Region pair when `hasEdgeAuth` (edgeAuthMode !== 'default'), or falls through to `<ConnectionConfig>` otherwise; then renders the Edge settings section (Authentication Mode Select, Username/Password when `hasEdgeAuth`, SSL Certificate textarea) |
| `src/components/ConfigEditor.tsx:23-27` | `edgeAuthMethods` options: `{ value: 'default', label: 'Standard' }`, `{ value: 'linux', label: 'Linux' }`, `{ value: 'ldap', label: 'LDAP' }` with per-option descriptions |
| `src/components/ConfigEditor.tsx:37-39,186-188` | `SecureSocksProxySettings` render — writes `jsonData.enableSecureSocksProxy`; deliberately excluded per AGENTS.md |
| `src/components/ConfigEditor.tsx:37,110` | `<ConnectionConfig>` is called without `showHttpProxySettings`, so the AWS proxy fields are NOT editor-visible for IoT SiteWise (excluded from this entry) |
| `src/regions.ts:5-20` | `supportedRegions` (sitewise-specific): us-east-2, us-east-1, us-west-2, ap-south-1, ap-northeast-2, ap-southeast-1, ap-southeast-2, ap-northeast-1, ca-central-1, eu-central-1, eu-west-1, us-gov-west-1, cn-north-1, **`Edge`** (sentinel) |
| `src/regions.ts:22-28` | `DEFAULT_REGION = 'default'` sentinel used at query time; not stored in jsonData |
| `src/types.ts:264-274` | `SitewiseOptions extends AwsAuthDataSourceJsonData` adds `edgeAuthMode?`, `edgeAuthUser?`; `SitewiseSecureJsonData extends AwsAuthDataSourceSecureJsonData` adds `edgeAuthPass?`, `cert?` |
| `pkg/models/setting.go:11-14` | Backend constants: `EDGE_REGION = "Edge"`, `EDGE_AUTH_MODE_DEFAULT = "default"`, `EDGE_AUTH_MODE_LDAP = "ldap"`, `EDGE_AUTH_MODE_LINUX = "linux"` |
| `pkg/models/setting.go:16-22` | Backend `AWSSiteWiseDataSourceSetting` struct: embeds `awsds.AWSDatasourceSettings`, adds `Cert string`json:"-"``, `EdgeAuthMode string`json:"edgeAuthMode"``, `EdgeAuthUser string`json:"edgeAuthUser"``, `EdgeAuthPass string`json:"-"`` |
| `pkg/models/setting.go:24-50` | `Load`: unmarshals `JSONData` only when `len(JSONData) > 1`; substitutes `DefaultRegion` for `Region` when Region is empty/`"default"`; falls back to `config.Database` for `Profile` (legacy CloudWatch shim); defaults `EdgeAuthMode` to `"default"` when Region is Edge; copies `accessKey`, `secretKey`, `sessionToken`, `cert`, `edgeAuthPass` from decrypted secure data |
| `pkg/models/setting.go:52-74` | `Validate`: no-op when Region isn't Edge; when Edge, requires `Endpoint`, `Cert`, and (if `EdgeAuthMode != "default"`) `EdgeAuthUser` + `EdgeAuthPass`. Encoded as `requiredWhen` in the schema |
| `pkg/sitewise/client/client.go` (via imports) | Confirms Endpoint + Cert are consumed at HTTP client build time for Edge deployments |
| `package.json:41-43` | External component versions: `@grafana/aws-sdk@0.10.2`, `@grafana/plugin-ui@^0.13.0`, `@grafana/ui@^12.1.0`, `@grafana/data@^12.1.0` |

### External editor components

Read at the exact versions pinned in the plugin's `package.json`.

| Component | Version | Source consulted | What was read |
| --- | --- | --- | --- |
| `ConnectionConfig`, `ConnectionConfigProps`, `Divider` | `@grafana/aws-sdk@0.10.2` | grafana/grafana-aws-sdk-react tag `v0.10.2`, `src/components/ConnectionConfig.tsx`, `src/providers.ts`, `src/types.ts`, `src/regions.ts` | Every AWS field's label, placeholder, description, and conditional render; the Select option labels for `authType`; the fact that when the caller passes `standardRegions`, the region Select renders those with `allowCustomValue` |
| `ConfigSection` | `@grafana/plugin-ui@^0.13.0` | Editor layout — no storage fields |
| `SecureSocksProxySettings` (excluded) | `@grafana/ui@^12.1.0` | Writes `jsonData.enableSecureSocksProxy`; deliberately excluded from this entry |
| `Alert`, `Button`, `Field`, `Input`, `Select` | `@grafana/ui@^12.1.0` | Prop names (`label`, `placeholder`, `description`, `value`, `onChange`, `onReset`) so we knew which UI attributes to record |
| `updateDatasourcePluginJsonDataOption`, `updateDatasourcePluginSecureJsonDataOption`, `onUpdateDatasourceJsonDataOption`, `onUpdateDatasourceJsonDataOptionSelect`, `onUpdateDatasourceResetOption` | `@grafana/data@^12.1.0` | Storage-key semantics of the update helpers used by ConfigEditor and by ConnectionConfig |

### Backend Go dependency (`grafana-aws-sdk`)

| File | What was read |
| --- | --- |
| `pkg/awsds/settings.go` (`v1.4.3`) | `AuthType` int enum + custom `MarshalJSON`/`UnmarshalJSON` — the storage⇆enum mapping (including legacy `arn`→`default` and `sharedCreds`→`credentials`) that we surface as `AWSAuthType` string constants |
| `pkg/awsds/settings.go` (`v1.4.3`) | `AWSDatasourceSettings` struct with the AWS-shared fields; note `AssumeRoleARN string`json:"assumeRoleARN"`` (uppercase `ARN`) while the frontend writes camelCase `assumeRoleArn` — Go's case-insensitive Unmarshal makes both work |
| `pkg/awsds/settings.go` (`v1.4.3`) | `Load` copies decrypted `accessKey`, `secretKey`, `sessionToken`, and `proxyPassword` from secure JSON data; also mirrors `defaultRegion`→`region` at load time |

## Field provenance

Every dsconfig field, traced from its schema `id` to the upstream `file:line` where
each of its label, placeholder, tooltip, default, storage key, and value type is
defined.

| Schema `id` | Storage key | Target | Label source | Placeholder / options / default source | Value type source | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `jsonData_authType` | `authType` | `jsonData` | `ConnectionConfig.tsx` (`<Field label="Authentication Provider">`) | Options from `@grafana/aws-sdk` `src/providers.ts` (`awsAuthProviderOptions`); description from ConnectionConfig; default `awsds/settings.go` (`AuthTypeDefault` iota) → `"default"` | `AWSDatasourceSettings.AuthType` `awsds/settings.go` (int enum, string on wire) | Role `auth.discriminator`; validation `allowedValues` includes legacy `arn` |
| `jsonData_profile` | `profile` | `jsonData` | ConnectionConfig `<Field label="Credentials Profile Name">` | Placeholder `"default"`; description from ConnectionConfig | `AWSDatasourceSettings.Profile` `awsds/settings.go` | `dependsOn` from conditional render (authType == 'credentials'), extended with the Edge visibility gate |
| `secureJsonData_accessKey` | `accessKey` | `secureJsonData` | ConnectionConfig `<Field label="Access Key ID">` | — | `AWSDatasourceSettings.AccessKey` `awsds/settings.go` | Role `auth.aws.accessKeyId`; visibility gated by Edge state |
| `secureJsonData_secretKey` | `secretKey` | `secureJsonData` | ConnectionConfig `<Field label="Secret Access Key">` | — | `AWSDatasourceSettings.SecretKey` `awsds/settings.go` | Role `auth.aws.secretAccessKey`; visibility gated by Edge state |
| `secureJsonData_sessionToken` | `sessionToken` | `secureJsonData` | — (no UI) | — | `AWSDatasourceSettings.SessionToken` `awsds/settings.go` | Role `auth.aws.sessionToken`; tagged `backend-only`; still read at `pkg/models/setting.go:46` |
| `jsonData_assumeRoleArn` | `assumeRoleArn` | `jsonData` | ConnectionConfig `<Field label="Assume Role ARN">` | Placeholder `"arn:aws:iam:*"`; description from ConnectionConfig | `AWSDatasourceSettings.AssumeRoleARN` `awsds/settings.go` (backend json tag `assumeRoleARN`) | Frontend writes `assumeRoleArn`; case-insensitive Unmarshal rescues the mismatch |
| `jsonData_externalId` | `externalId` | `jsonData` | ConnectionConfig `<Field label="External ID">` | Placeholder + description from ConnectionConfig | `AWSDatasourceSettings.ExternalID` `awsds/settings.go` | `dependsOn` gates on `authType != 'grafana_assume_role'` plus Edge visibility |
| `jsonData_endpoint` | `endpoint` | `jsonData` | ConnectionConfig `<Field label="Endpoint">`, mirrored by `ConfigEditor.tsx:87` (Edge custom control) | Placeholder `https://{service}.{region}.amazonaws.com`; description `Optionally, specify a custom endpoint for the service` | `AWSDatasourceSettings.Endpoint` `awsds/settings.go` | `requiredWhen defaultRegion == 'Edge'` per `pkg/models/setting.go:57-59` |
| `jsonData_defaultRegion` | `defaultRegion` | `jsonData` | ConnectionConfig `<Field label="Default Region">`, mirrored by `ConfigEditor.tsx:98` (Edge custom control) | Options inlined from `src/regions.ts:5-20` (sitewise-specific list including sentinel `'Edge'`); description from AWS ConnectionConfig | `AWSDatasourceSettings.DefaultRegion` `awsds/settings.go` | `<Select ... allowCustomValue={true}>` |
| `jsonData_edgeAuthMode` | `edgeAuthMode` | `jsonData` | `ConfigEditor.tsx:122` (`<Field label="Authentication Mode">`) | Options from `edgeAuthMethods` (`ConfigEditor.tsx:23-27`); backend default `"default"` (`pkg/models/setting.go:40-42`) | `AWSSiteWiseDataSourceSetting.EdgeAuthMode` `pkg/models/setting.go:19` | Visible only when `defaultRegion == 'Edge'` |
| `jsonData_edgeAuthUser` | `edgeAuthUser` | `jsonData` | `ConfigEditor.tsx:135` (`<Field label="Username" description="The username set to local authentication proxy">`) | Description verbatim from source | `AWSSiteWiseDataSourceSetting.EdgeAuthUser` `pkg/models/setting.go:20` | `requiredWhen defaultRegion == 'Edge' && edgeAuthMode != 'default'` per `setting.go:64-67` |
| `secureJsonData_edgeAuthPass` | `edgeAuthPass` | `secureJsonData` | `ConfigEditor.tsx:146` (`<Field label="Password" description="The password sent to local authentication proxy">`) | Description verbatim from source | `AWSSiteWiseDataSourceSetting.EdgeAuthPass` `pkg/models/setting.go:21` (json:"-", loaded from decrypted secure data at `:48`) | `requiredWhen defaultRegion == 'Edge' && edgeAuthMode != 'default'` |
| `secureJsonData_cert` | `cert` | `secureJsonData` | `ConfigEditor.tsx:162` (`<Field label="SSL Certificate" description="Certificate for SSL enabled authentication.">`) | Placeholder `"Begins with -----BEGIN CERTIFICATE------"` (`ConfigEditor.tsx:180` — kept verbatim, including the six-dash typo) | `AWSSiteWiseDataSourceSetting.Cert` `pkg/models/setting.go:18` (json:"-", loaded from decrypted secure data at `:47`) | `requiredWhen defaultRegion == 'Edge'` per `setting.go:60-62` |

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
| `jsonData_edgeAuthMode` | `edgeAuthMode` | `jsonData` | Authentication Mode | Yes |
| `jsonData_edgeAuthUser` | `edgeAuthUser` | `jsonData` | Username | Yes |
| `secureJsonData_edgeAuthPass` | `edgeAuthPass` | `secureJsonData` | Password | Yes |
| `secureJsonData_cert` | `cert` | `secureJsonData` | SSL Certificate | Yes |

### Frontend-only settings

None. The IoT SiteWise config editor writes only fields the backend consumes.

### Backend-only settings

- **`sessionToken`** — no editor UI; still read by `pkg/models/setting.go:46` for
  temporary STS credentials paired with `authType: "keys"`. Provisioning is the
  practical way to set it.

### Fields excluded from this entry

- **AWS proxy fields** (`proxyType`, `proxyUrl`, `proxyUsername`, `proxyPassword`) —
  `ConfigEditor.tsx:36,110` calls `<ConnectionConfig>` without `showHttpProxySettings`,
  so the proxy subsection is not part of IoT SiteWise's declared editor surface.
- **Secure Socks Proxy** (`jsonData.enableSecureSocksProxy`) — excluded per AGENTS.md.
- **`region`** — the AWS SDK Go struct has a `Region` field
  (`awsds/settings.go`), but the frontend never writes it; the backend mirrors
  `defaultRegion` into it at load time (`pkg/models/setting.go:31-33`). Not a stored
  config.

## Where the types are defined

The IoT SiteWise configuration types are spread across the plugin and its
dependencies. Some fields and base types come from libraries/SDKs rather than the
plugin itself.

### Frontend (TypeScript)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `SitewiseOptions` (jsonData), `SitewiseSecureJsonData` | `src/types.ts:264-274` | plugin ([grafana/iot-sitewise-datasource](https://github.com/grafana/iot-sitewise-datasource)) |
| `AwsAuthType`, `AwsAuthDataSourceJsonData` (base of `SitewiseOptions`), `AwsAuthDataSourceSecureJsonData` (base of `SitewiseSecureJsonData`) | `src/types.ts` | `@grafana/aws-sdk` `0.10.2` (grafana/grafana-aws-sdk-react `v0.10.2`) |
| `awsAuthProviderOptions` (Select options for `authType`) | `src/providers.ts` | `@grafana/aws-sdk` `0.10.2` |
| `ConnectionConfig` React component (renders every AWS-shared field) | `src/components/ConnectionConfig.tsx` | `@grafana/aws-sdk` `0.10.2` |
| `edgeAuthMethods` (options for the Authentication Mode Select) | `src/components/ConfigEditor.tsx:23-27` | plugin |
| `supportedRegions` (includes `'Edge'`) | `src/regions.ts:5-20` | plugin |
| `DataSourceJsonData` (base type of `AwsAuthDataSourceJsonData`) | `packages/grafana-data/src/types/datasource.ts` | `@grafana/data` `^12.1.0` |
| `ConfigSection` | `src/components/ConfigEditor/` | `@grafana/plugin-ui` `^0.13.0` |
| `SecureSocksProxySettings` (excluded from this entry) | `packages/grafana-ui/src/components/DataSourceSettings/SecureSocksProxySettings.tsx` | `@grafana/ui` `^12.1.0` |

### Backend (Go)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `AWSSiteWiseDataSourceSetting`, `Load`, `Validate`, `ToAWSDatasourceSettings` | `pkg/models/setting.go:16-91` | plugin ([grafana/iot-sitewise-datasource](https://github.com/grafana/iot-sitewise-datasource)) |
| `AWSDatasourceSettings` (embedded base of `AWSSiteWiseDataSourceSetting`), `AuthType` int enum + custom Marshal/Unmarshal | `pkg/awsds/settings.go` | `github.com/grafana/grafana-aws-sdk` `v1.4.3` |
| `backend.DataSourceInstanceSettings` (carries `JSONData`, `DecryptedSecureJSONData`) | `backend/common.go` | `github.com/grafana/grafana-plugin-sdk-go` `v0.291.1` |
| SiteWise SDK client (consumes the loaded settings) | `pkg/sitewise/*`, `pkg/server/*` | plugin |
| AWS SDK v2 client build-up (`iotsitewise.NewFromConfig(...)`) | — | `github.com/aws/aws-sdk-go-v2/service/iotsitewise` |

The models in this entry flatten that spread into a single Go `Config` type (jsonData
fields + `DecryptedSecureJSONData`) plus a `SecureJsonDataKey` typed constant list.
`settings.ts` keeps the three canonical TypeScript types (`RootConfig`,
`JsonDataConfig`, `SecureJsonDataConfig`). `AWSAuthType` + `EdgeAuthMode` constants in
`settings.go` mirror the string forms `awsds.AuthType` (Un)Marshals and the plugin's
own `EDGE_AUTH_MODE_*` constants, without carrying the int-enum + custom-JSON
machinery.

## Modeling decisions

- **Single, flat Go `Config`.** The upstream `AWSSiteWiseDataSourceSetting` embeds
  `awsds.AWSDatasourceSettings`. We flatten those into one struct to match the pattern
  used by other dsconfig registry entries (see `registry/grafana-athena-datasource`)
  and to avoid pulling `grafana-aws-sdk` into the shared registry `go.mod`.
- **`AssumeRoleARN` field vs `assumeRoleArn` tag** — the Go field name mirrors the
  upstream `awsds.AWSDatasourceSettings.AssumeRoleARN`, but the json tag mirrors what
  the frontend writes (`assumeRoleArn`). Case-insensitive `encoding/json` rescue.
- **`dependsOn` encodes the Edge branch precisely.** The editor swaps `ConnectionConfig`
  for a custom Edge UI when `defaultRegion === 'Edge' && edgeAuthMode !== 'default'`.
  Every AWS auth field's `dependsOn` therefore combines the AWS-specific gate
  (e.g. `authType == 'keys'`) with the Edge visibility gate
  `(defaultRegion != 'Edge' || edgeAuthMode == 'default')`. `endpoint` is the
  exception: it's rendered by both branches, so its gate is
  `(defaultRegion == 'Edge' && edgeAuthMode != 'default') || authType != 'grafana_assume_role'`.
- **`requiredWhen` encodes the backend contract from `Validate`.** `endpoint` and
  `cert` are required when `defaultRegion == 'Edge'`; `edgeAuthUser` and `edgeAuthPass`
  are required when `defaultRegion == 'Edge' && edgeAuthMode != 'default'`.
- **Region options inlined** — unlike the other AWS entries where the region Select is
  populated at runtime, the sitewise list is short, plugin-specific, and includes the
  sentinel `'Edge'`, so inlining the options in the schema documents the possible
  values explicitly.
- **`Edge` region modelled as a plain string option** — not a virtual field. The
  editor branches on the storage value directly (`ConfigEditor.tsx:30`), and there is
  no side-effecting selector — `defaultRegion` is written once and every downstream
  visibility rule reads it as-is.
- **`grafana_assume_role` is a schema-only value** — the editor only renders it when
  the `awsDatasourcesTempCredentials` feature toggle is on and the plugin is in
  `@grafana/aws-sdk`'s allow-list. We list it as a schema `allowedValues` entry, note
  the gating in an instruction, and Edge mode users effectively cannot pick it because
  the Edge branch replaces the provider Select.
- **`arn` kept in `allowedValues`, not in UI options** — `AwsAuthType.ARN` is
  deprecated (`grafana-aws-sdk-react/src/types.ts`) and does not appear in
  `awsAuthProviderOptions`, but stored datasources may still carry it.
- **`sessionToken` included as a secure key with no UI** — the plugin doesn't offer a
  UI field, but `pkg/models/setting.go:46` reads it from decrypted secure data.
  Tagged `backend-only`.
- **AWS proxy fields excluded** — IoT SiteWise does not pass `showHttpProxySettings` to
  `ConnectionConfig`, so the proxy fields are not part of the editor surface.
- **`SecureJsonDataConfig` is a key list** — secure values are write-only, so the type
  is just the array of secret key names (`accessKey`, `secretKey`, `sessionToken`,
  `edgeAuthPass`, `cert`); consumers read `secureJsonFields` to see what is
  configured.
- **`textarea` UI component for the SSL cert** — matches the plain HTML `<textarea>`
  the plugin uses (`ConfigEditor.tsx:173`), preserving the seven-row height as a hint
  even though the current dsconfig `FieldUI` doesn't yet round-trip a `rows` value for
  textarea inputs. The placeholder `"Begins with -----BEGIN CERTIFICATE------"` is
  preserved verbatim including the upstream six-dash typo.

## SDK plugin schema and k8s-style examples (`schema.go`)

`NewSchema()` assembles the `grafana-plugin-sdk-go` `pluginschema.PluginSchema`
bundle (the k8s-style schema Grafana's datasource API server serves as
`{apiVersion}.json`, `v0alpha1` today) from the embedded `dsconfig.json`.

`SettingsExamples()` provides the default configuration plus one k8s-style example
per AWS auth provider, an STS AssumeRole variant, one example per Edge authentication
mode, and a legacy `arn` example:

| Example | Auth | Extras | `secureJsonData` |
| --- | --- | --- | --- |
| `""` (default) | AWS SDK Default | — | `accessKey` (empty) |
| `awsSdkDefault` | AWS SDK Default | `defaultRegion` | `accessKey` (empty) |
| `accessAndSecretKey` | Access & secret key | `defaultRegion` | `accessKey`, `secretKey` |
| `credentialsFile` | Credentials file | `profile`, `defaultRegion` | `accessKey` (empty) |
| `workspaceIamRole` | Workspace IAM Role | `defaultRegion` | `accessKey` (empty) |
| `grafanaAssumeRole` | Grafana Assume Role | `defaultRegion` | `accessKey` (empty) |
| `assumeRoleFromKeys` | Access & secret key + STS AssumeRole | `assumeRoleArn`, `externalId`, `defaultRegion` | `accessKey`, `secretKey` |
| `edgeStandard` | Edge Kernel, standard mode | `defaultRegion=Edge`, `endpoint`, `edgeAuthMode=default` | `accessKey`, `secretKey`, `cert` |
| `edgeLinux` | Edge Kernel, Linux auth | `defaultRegion=Edge`, `endpoint`, `edgeAuthMode=linux`, `edgeAuthUser` | `edgeAuthPass`, `cert` |
| `edgeLdap` | Edge Kernel, LDAP auth | `defaultRegion=Edge`, `endpoint`, `edgeAuthMode=ldap`, `edgeAuthUser` | `edgeAuthPass`, `cert` |
| `legacyArnAuthType` | `arn` (legacy — backend maps to `default`) | `defaultRegion` | `accessKey` (empty) |

Every example carries at least one `secureJsonData` placeholder as required by the
conformance suite.

## `LoadConfig` utility (`settings.go`)

`LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error)`
runs the full three-phase load flow on a datasource instance's settings and returns a
fully-defaulted, validated `Config`:

1. **Parse** — unmarshal `settings.JSONData` into `Config` (only when the payload has
   more than one byte, matching `pkg/models/setting.go:25`) and copy the plugin's
   decrypted secrets (`accessKey`/`secretKey`/`sessionToken`/`edgeAuthPass`/`cert`)
   into `DecryptedSecureJSONData`.
2. **`ApplyDefaults`** — fills two curated defaults: `AuthType` defaults to
   `AWSAuthTypeDefault` (matching the reference AWS pack + backend
   `awsds.AuthTypeDefault` iota zero); `EdgeAuthMode` defaults to `EdgeAuthModeDefault`
   when — and only when — `DefaultRegion == "Edge"`, mirroring `setting.go:40-42`.
   `DefaultRegion` intentionally has no default.
3. **`Validate`** — enforces the runtime contract: known `AuthType`, `accessKey` +
   `secretKey` present for `keys` auth, and — when `DefaultRegion == "Edge"` —
   `endpoint` and `cert` present plus (when `edgeAuthMode != "default"`) `edgeAuthUser`
   and `edgeAuthPass` present. Mirrors `pkg/models/setting.go:52-74` verbatim. Errors
   are joined so callers see every problem at once.

Everything is logged via `backend.Logger.FromContext(ctx)` with `datasource_uid`,
`datasource_name`, and `plugin` labels so log lines carry request context.

### Direct access to individual phases

`(*Config).ApplyDefaults()` and `(Config).Validate() error` are exported for callers
that want to compose them themselves (provisioning preview, schema-example round-trip,
tests that need to distinguish parse-level from policy-level errors). Skip them by
never calling `LoadConfig` in those flows — assemble a `Config` directly.

## Upstream findings

Potential bugs, misleading UX, and consistency issues discovered while researching
upstream. All preserved verbatim in the schema — the schema records what the plugin
**does**, not what it **should** do; these notes exist so reviewers can reproduce
each finding and decide separately whether to fix upstream.

1. **`Edge` is a plain-string sentinel wedged into an AWS-region list.** `src/regions.ts:5-20`
   mixes real AWS regions with the string `'Edge'`, which the editor pattern-matches
   on to switch into a completely different rendering path (`ConfigEditor.tsx:30`).
   A single misspelling (`edge`, `EDGE`) silently falls back to the AWS branch even
   though the backend's `EDGE_REGION` constant is exact-match ("Edge",
   `pkg/models/setting.go:11`). Modelling it as an explicit `deploymentMode` field
   would be safer.
2. **SSL Certificate placeholder has six trailing dashes.** `ConfigEditor.tsx:180`
   reads `placeholder="Begins with -----BEGIN CERTIFICATE------"` — a real PEM cert
   begins with exactly five dashes. Preserved verbatim in the schema (`secureJsonData_cert`).
3. **Edge password field's `secureJsonFields` key mismatch.** `ConfigEditor.tsx:73`
   resets the reveal marker under the key `password`
   (`secureJsonFields: { ..., password: false }`), but the actual secret key is
   `edgeAuthPass`. In practice the reset still clears the secret via
   `secureJsonData.edgeAuthPass = ''`, but the `secureJsonFields.password` flag never
   corresponds to any secret the backend cares about — it's dead state.
4. **`AssumeRoleARN` json tag disagrees with the frontend's `assumeRoleArn`.**
   `awsds/settings.go` uses `` json:"assumeRoleARN" `` (uppercase RN), but the
   frontend type (`grafana-aws-sdk-react/src/types.ts`) and every `onChange` in
   ConnectionConfig write `assumeRoleArn` (lowercase arn). Case-insensitive-match
   rescue.
5. **The "Grafana Assume Role" provider only appears when a feature toggle is on and
   the plugin is in an allow-list.** `@grafana/aws-sdk`'s `ConnectionConfig` restricts
   the provider to plugins listed in `DS_TYPES_THAT_SUPPORT_TEMP_CREDS` and further
   gates it on `config.featureToggles.awsDatasourcesTempCredentials`. Storage-side the
   value is valid regardless.
6. **Deprecated `arn` auth value silently maps to `default`.** `awsds/settings.go`
   maps any unknown auth type string to `AuthTypeDefault`, comment: "For old 'arn'
   option". A provisioned config with `authType: "arn"` will load as if it were
   `default`, with no warning surfaced.
7. **`sessionToken` has no editor UI.** `grafana-aws-sdk-react` never exposes an input
   for `sessionToken`, even though the backend reads it and it is required for
   temporary credentials. Users must provision it directly.
8. **`Profile` falls back to `config.Database` (legacy CloudWatch shim).**
   `pkg/models/setting.go:35-37` copies `settings.Database` into `Profile` when
   `Profile` is empty. The comment explicitly calls this "legacy support (only for
   cloudwatch?)" and it has no equivalent on the frontend, so datasources provisioned
   with a `database` root-level field get an unexpected profile mapping.
9. **`endpoint` in Edge mode uses itself as its placeholder.** `ConfigEditor.tsx:93`
   reads `placeholder={endpoint ?? 'https://{service}.{region}.amazonaws.com'}` — the
   `??` operator only falls back to the AWS URL template when `endpoint` is
   `null`/`undefined`, so once the user types anything the placeholder becomes the
   value itself. Harmless (the placeholder is hidden when a value is present) but
   redundant.
10. **`Load` accepts an empty `JSONData` slice but rejects a one-byte slice.**
    `setting.go:25` guards the unmarshal with `if len(config.JSONData) > 1`, so a
    literal payload of one character (e.g. `{`) is treated as if there were no
    settings at all. `LoadConfig` mirrors this exactly for parity with upstream.

## Validation performed

- `dsconfig.ParseAndResolveSchemaJSON` + `Schema.Validate()` — passes.
- JSON Schema validation against [`dsconfig/schema.json`](../../dsconfig/schema.json)
  (draft-07, `additionalProperties: false`) — passes.
- `go test ./...` on this module — passes (schema bundle shape,
  `SchemaSpecHasNoSecureJSON`, `SecureValuesMatchLoadSettings`, `JSONDataMatchesStruct`,
  `JSONDataTypesMatchStruct`, `SchemaArtifactInSync`, `LoadConfig`, `ApplyDefaults`,
  `Validate` per auth type / Edge mode).
- `settings.go` / `schema.go`: `go build`, `go vet`, `gofmt` — clean.
- `settings.ts`: `tsc --noEmit --strict` (TypeScript 5) — clean.
- All 26 pre-existing registry entries still `go test` clean.
