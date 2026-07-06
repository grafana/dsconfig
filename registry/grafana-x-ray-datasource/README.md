# grafana-x-ray-datasource

Declarative configuration schema for the [AWS Application Signals (X-Ray) datasource plugin](https://github.com/grafana/x-ray-datasource) (`grafana-x-ray-datasource`).

The plugin was renamed from "X-Ray" to "AWS Application Signals" in v2.16.0
(upstream PR #384). The plugin id and registry directory keep the historical
`grafana-x-ray-datasource` form for backward compatibility.

## Upstream researched

- **Repo**: `github.com/grafana/x-ray-datasource`
- **Ref**: `main`
- **Commit SHA**: `3d8a237e953f53b4ba11cc5decb1e47f2245ad8b`

Every value in [`dsconfig.json`](dsconfig.json) — labels, placeholders, tooltips, option
labels/values, section titles, defaults, validations, dependency and required-when
expressions, storage keys, storage targets, value types, group titles, and instructions —
is traceable to a specific `file:line` in the upstream plugin repo (or in the pinned
`@grafana/aws-sdk` version of the shared `ConnectionConfig` component) at this SHA. See
[Field provenance](#field-provenance) below.

To reproduce this research:

```bash
git clone --depth 20 https://github.com/grafana/x-ray-datasource
cd x-ray-datasource
git checkout 3d8a237e953f53b4ba11cc5decb1e47f2245ad8b

# The AWS auth surface is rendered by @grafana/aws-sdk's ConnectionConfig. Pin
# to the version x-ray-datasource's package.json uses (currently 0.10.2):
git clone --branch v0.10.2 https://github.com/grafana/grafana-aws-sdk-react
# SHA fe0c4d8d657ee5ed053ae173293dc876619b5a2b
```

If upstream `main` has advanced past the pinned SHA, re-diff the sources listed under
[Sources researched](#sources-researched) and reconcile the schema before merging.

## Files

| File | Purpose |
| --- | --- |
| [`dsconfig.json`](dsconfig.json) | dsconfig v1 schema — single source of truth for all config fields, groups, and instructions |
| [`settings.ts`](settings.ts) | TypeScript models: `RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig` |
| [`settings.go`](settings.go) | Go `Config` model (jsonData fields + `Database` legacy fallback + `DecryptedSecureJSONData`), `PluginID`, `AWSAuthType` typed constants, `SecureJsonDataKey` typed constants, and the `LoadConfig` utility |
| [`schema.go`](schema.go) | k8s-style SDK plugin schema: embeds `dsconfig.json`, converts it via `dsconfig.NewSDKSchema`, and defines `SettingsExamples` for each AWS auth provider |
| [`settings_test.go`](settings_test.go) | Table tests for `LoadConfig`, `ApplyDefaults`, `Validate`, and `EffectiveProfile` |
| [`conformance_test.go`](conformance_test.go) | Runs `schema.RunPluginTests` — the shared dsconfig conformance suite — against `dsconfig.json` |
| [`schema.gen.json`](schema.gen.json), [`settings.gen.json`](settings.gen.json), [`settings.examples.gen.json`](settings.examples.gen.json) | Committed schema artifacts (regenerate with `go generate ./...` inside this directory; the `SchemaArtifactInSync` conformance subtest guards drift) |

There is no per-entry `go.mod` — every registry entry is a subpackage of the shared
[`registry/`](..) module (`github.com/grafana/dsconfig/registry`).

## Sources researched

Every source below was read at the pinned upstream SHA
(`3d8a237e953f53b4ba11cc5decb1e47f2245ad8b`) or, for `@grafana/aws-sdk`, at the exact
version pinned in the plugin's `package.json` (`0.10.2` / SHA `fe0c4d8`).

### Plugin repo (`github.com/grafana/x-ray-datasource@3d8a237`)

| File | What was read |
| --- | --- |
| `src/plugin.json:2-47` | `pluginType` (`id` = `grafana-x-ray-datasource`), `pluginName` (`name` = "AWS Application Signals"), `info.links[]` (only points at the GitHub repo — the docs URL used matches Grafana's plugin catalog URL pattern for consistency with the AWS cousins) |
| `src/components/ConfigEditor/ConfigEditor.tsx:1-20` | Entire editor: renders `<ConnectionConfig {...props} standardRegions={standardRegions} />` and, when `secureSocksDSProxyEnabled` and Grafana ≥ 10.0.0, `<SecureSocksProxySettings>`. No plugin-specific fields; no `showHttpProxySettings`, `hideAssumeRoleArn`, or `skipEndpoint` are passed |
| `src/components/ConfigEditor/regions.ts:1-28` | The 26-region `standardRegions` list fed to ConnectionConfig |
| `src/types.ts:106-108` | `XrayJsonData extends AwsAuthDataSourceJsonData` with a "Can add X-Ray specific values here" placeholder — the plugin has no jsonData fields of its own |
| `pkg/datasource/configuration.go:8-20` | Backend `getDsSettings`: delegates to `awsds.AWSDatasourceSettings.Load` verbatim, then applies the legacy `settings.Database → Profile` fallback when `Profile` is empty |
| `pkg/datasource/datasource.go:32-55` | Backend consumes `settings.Region` (mirrored from `defaultRegion` by awsds.Load) when building the X-Ray / Application Signals clients — confirms `defaultRegion` is required for runtime queries |
| `pkg/datasource/configuration_test.go:12-47` | The plugin's own parse test — mirrors what LoadConfig must accept, including `assumeRoleARN` (uppercase RN) in jsonData and `sessionToken` in decrypted secure data |
| `CHANGELOG.md:1-11` (v2.17.0) | "Add sessionToken handling to support Grafana Assume Role" — why sessionToken matters for `grafana_assume_role` even though no editor UI writes it |
| `CHANGELOG.md` (v2.16.0) | "Rename plugin from X-Ray to App Signals" — why the pluginName differs from the pluginType |
| `package.json` | External component versions (see next table) |
| `go.mod` | `github.com/grafana/grafana-aws-sdk v1.4.3` — the backend AWS auth surface |

### External editor components

Read at the exact versions pinned in the plugin's `package.json`.

| Component | Version | Source consulted | What was read |
| --- | --- | --- | --- |
| `ConnectionConfig`, `AwsAuthType` | `@grafana/aws-sdk@0.10.2` | grafana/grafana-aws-sdk-react tag `v0.10.2` (SHA `fe0c4d8`), `src/components/ConnectionConfig.tsx`, `src/providers.ts`, `src/types.ts`, `src/regions.ts` | Every AWS-shared field's label, placeholder, description, and conditional render; the Select option labels for `authType`; the standard region list; the `DS_TYPES_THAT_SUPPORT_TEMP_CREDS` allow-list containing `grafana-x-ray-datasource` |
| `SecureSocksProxySettings` (excluded) | `@grafana/ui@12.4.3` | Writes `jsonData.enableSecureSocksProxy`; deliberately excluded per AGENTS.md |
| `Field`, `Input`, `Select`, `ButtonGroup`, `ToolbarButton`, `Collapse` (via ConnectionConfig) | `@grafana/ui@12.4.3` | Prop names (`label`, `placeholder`, `value`, `onChange`, `isConfigured`, `onReset`) so we knew which UI attributes to record |
| `ConfigSection`, `ConfigSubSection` | `@grafana/plugin-ui@0.13.1` | Editor layout — no storage fields |
| `DataSourcePluginOptionsEditorProps`, `onUpdateDatasourceJsonDataOption`, `onUpdateDatasourceSecureJsonDataOption`, `onUpdateDatasourceJsonDataOptionSelect`, `onUpdateDatasourceResetOption` | `@grafana/data@12.4.3` | Storage-key semantics of the update helpers used by ConnectionConfig |

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
| `jsonData_authType` | `authType` | `jsonData` | `ConnectionConfig.tsx:105` (`<Field label="Authentication Provider">`) | Options `providers.ts:4-25` (`awsAuthProviderOptions`); description `ConnectionConfig.tsx:106`; default `awsds/settings.go:16` (`AuthTypeDefault` iota) → `"default"` | `AWSDatasourceSettings.AuthType` `awsds/settings.go:97` (int enum, string on wire) | Role `auth.discriminator`; validation `allowedValues` includes legacy `arn` |
| `jsonData_profile` | `profile` | `jsonData` | `ConnectionConfig.tsx:123` (`<Field label="Credentials Profile Name">`) | Placeholder `ConnectionConfig.tsx:129` (`"default"`); description `:124` | `AWSDatasourceSettings.Profile` `awsds/settings.go:95` | `dependsOn` from conditional render `ConnectionConfig.tsx:121` |
| `secureJsonData_accessKey` | `accessKey` | `secureJsonData` | `ConnectionConfig.tsx:137` (`<Field label="Access Key ID">`) | — | `AWSDatasourceSettings.AccessKey` `awsds/settings.go:113` | Role `auth.aws.accessKeyId`; `dependsOn`/`requiredWhen` from conditional render `ConnectionConfig.tsx:135` |
| `secureJsonData_secretKey` | `secretKey` | `secureJsonData` | `ConnectionConfig.tsx:157` (`<Field label="Secret Access Key">`) | — | `AWSDatasourceSettings.SecretKey` `awsds/settings.go:114` | Role `auth.aws.secretAccessKey`; `dependsOn`/`requiredWhen` from conditional render `ConnectionConfig.tsx:135` |
| `secureJsonData_sessionToken` | `sessionToken` | `secureJsonData` | — (no UI) | — | `AWSDatasourceSettings.SessionToken` `awsds/settings.go:115` | Role `auth.aws.sessionToken`; tagged `backend-only`; consumed by `awsds/settings.go:137` and, since plugin v2.17.0, by the Grafana Assume Role flow |
| `jsonData_assumeRoleArn` | `assumeRoleArn` | `jsonData` | `ConnectionConfig.tsx:261` (`<Field ... label="Assume Role ARN">`) | Placeholder `:268` (`"arn:aws:iam:*"`); description `:262-264` (verbatim, including the multi-line whitespace) | `AWSDatasourceSettings.AssumeRoleARN` `awsds/settings.go:98` (backend json tag `assumeRoleARN`) | Rendered when `!hideAssumeRoleArn && awsAssumeRoleEnabled` — X-Ray passes neither, so it is visible for every auth provider (including `grafana_assume_role`, where Grafana still uses it as a display hint). No jsonData-based `dependsOn` — the two gating flags are compile-time / Grafana-instance-config |
| `jsonData_externalId` | `externalId` | `jsonData` | `ConnectionConfig.tsx:276` (`<Field ... label="External ID">`) | Placeholder `:281`; description `:277` | `AWSDatasourceSettings.ExternalID` `awsds/settings.go:99` | `dependsOn` from conditional render `ConnectionConfig.tsx:273` (not `grafana_assume_role`) |
| `jsonData_endpoint` | `endpoint` | `jsonData` | `ConnectionConfig.tsx:362` (`<Field label="Endpoint">`) | Placeholder `:368` (X-Ray passes no `defaultEndpoint`, so `'https://{service}.{region}.amazonaws.com'`); description `:363` | `AWSDatasourceSettings.Endpoint` `awsds/settings.go:102` | `dependsOn` from conditional render `ConnectionConfig.tsx:360` (not `grafana_assume_role`) — X-Ray does not pass `skipEndpoint` |
| `jsonData_defaultRegion` | `defaultRegion` | `jsonData` | `ConnectionConfig.tsx:376` (`<Field label="Default Region">`) | Description `:377` (verbatim, including the padded backticks `` ` us-west-2 ` ``); options at runtime from `standardRegions` (`src/components/ConfigEditor/regions.ts:1-28`, mirroring `@grafana/aws-sdk/src/regions.ts`) — modelled as `select` with `allowCustom` | `AWSDatasourceSettings.DefaultRegion` `awsds/settings.go:105` | `<Select ... allowCustomValue={true}>` |

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

### Frontend-only settings

None. The X-Ray config editor writes only fields the backend consumes.

### Backend-only settings

- **`sessionToken`** — no editor UI; still read by `AWSDatasourceSettings.Load` at
  `awsds/settings.go:137`. Used with `authType: "keys"` for temporary STS credentials
  and, since plugin v2.17.0 (upstream PR #628), to support the `grafana_assume_role`
  flow. Provisioning is the only practical way to set it.

### Root-level settings the backend reads

- **`database`** — the top-level datasource `database` field. The X-Ray backend's
  `getDsSettings` (`pkg/datasource/configuration.go:16-18`) uses it as a legacy fallback
  for `jsonData.profile` when the latter is empty. No editor UI writes it. Modelled on
  the Go `Config` as a `Database string \`json:"-"\`` (root-level, not jsonData) and
  populated by `LoadConfig` from `settings.Database`; `RootConfig` in TypeScript keeps
  it as an optional `database` string. Never appears in `dsconfig.json` because it is
  a Grafana-native root field, not a plugin-declared one.

### Fields excluded from this entry

- **AWS proxy fields** (`proxyType`, `proxyUrl`, `proxyUsername`, `proxyPassword`) —
  `ConnectionConfig.tsx:291` only renders the proxy subsection when the caller passes
  `showHttpProxySettings`. X-Ray's `ConfigEditor.tsx` does not, so proxy fields are
  neither editor-visible nor part of X-Ray's declared surface, even though `awsds.Load`
  would still consume them if provisioned. Consumers needing them should reach for the
  shared `aws_sdk_settings.json` pack directly.
- **Secure Socks Proxy** (`jsonData.enableSecureSocksProxy`) — excluded per AGENTS.md.
- **`region`** — the AWS SDK Go struct has a `Region` field (`awsds/settings.go:96`),
  but the frontend never writes it; the backend mirrors `defaultRegion` into it at load
  time (`awsds/settings.go:127-129`). Not a stored config.

## Where the types are defined

The X-Ray configuration types are almost entirely borrowed from `@grafana/aws-sdk` and
`grafana-aws-sdk` — the plugin adds no jsonData fields of its own.

### Frontend (TypeScript)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `XrayJsonData` (empty extension of `AwsAuthDataSourceJsonData`) | `src/types.ts:106-108` | plugin ([grafana/x-ray-datasource](https://github.com/grafana/x-ray-datasource)) |
| `standardRegions` (identical copy of the SDK list) | `src/components/ConfigEditor/regions.ts:1-28` | plugin (mirrors `@grafana/aws-sdk`) |
| `AwsAuthType`, `AwsAuthDataSourceJsonData`, `AwsAuthDataSourceSecureJsonData` | `src/types.ts:3-32` | `@grafana/aws-sdk` `0.10.2` (grafana/grafana-aws-sdk-react `v0.10.2`) |
| `awsAuthProviderOptions` (Select options for `authType`) | `src/providers.ts:4-25` | `@grafana/aws-sdk` `0.10.2` |
| `ConnectionConfig` React component | `src/components/ConnectionConfig.tsx:36-404` | `@grafana/aws-sdk` `0.10.2` |
| `DataSourceJsonData` (base type of `AwsAuthDataSourceJsonData`) | `packages/grafana-data/src/types/datasource.ts` | `@grafana/data` `12.4.3` |
| `ConfigSection`, `ConfigSubSection` | `src/components/ConfigEditor/` | `@grafana/plugin-ui` `0.13.1` |
| `SecureSocksProxySettings` (excluded from this entry) | `packages/grafana-ui/src/components/DataSourceSettings/SecureSocksProxySettings.tsx` | `@grafana/ui` `12.4.3` |

### Backend (Go)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `getDsSettings` (the entire plugin-specific settings-loading surface, including the legacy `Database → Profile` fallback) | `pkg/datasource/configuration.go:8-20` | plugin ([grafana/x-ray-datasource](https://github.com/grafana/x-ray-datasource)) |
| `AWSDatasourceSettings` (the settings model this plugin uses directly, without a wrapping type), `AuthType` int enum + custom Marshal/Unmarshal, `Load` (copies decrypted secrets, mirrors `defaultRegion`→`region`) | `pkg/awsds/settings.go:13-141` | `github.com/grafana/grafana-aws-sdk` `v1.4.3` |
| `backend.DataSourceInstanceSettings` (carries `JSONData`, `DecryptedSecureJSONData`, and the root-level `Database` field the plugin reads) | `backend/common.go` | `github.com/grafana/grafana-plugin-sdk-go` |
| AWS SDK v2 client build-up (`xray.NewFromConfig(...)`, `applicationsignals.NewFromConfig(...)`) | `pkg/client/client.go` | `github.com/aws/aws-sdk-go-v2` |

The models in this entry flatten that spread into a single Go `Config` type (jsonData
fields + a root-level `Database` field + `DecryptedSecureJSONData`) plus a
`SecureJsonDataKey` typed constant list. `settings.ts` keeps the three canonical
TypeScript types (`RootConfig` — carrying the legacy `database` fallback,
`JsonDataConfig`, `SecureJsonDataConfig`). `AWSAuthType` constants in `settings.go`
mirror the string forms `awsds.AuthType` (Un)Marshals, without carrying the int-enum +
custom-JSON machinery.

## Modeling decisions

- **Single, flat Go `Config`.** The upstream plugin does not define a wrapping settings
  type — it uses `awsds.AWSDatasourceSettings` directly via `getDsSettings`. We flatten
  that struct into one Go value to match the pattern used by the other AWS registry
  entries (see `registry/grafana-athena-datasource`, `registry/cloudwatch`) and to
  avoid pulling `grafana-aws-sdk` into the shared registry `go.mod`.
- **Root-level `Database` on the Go `Config`.** X-Ray is one of the few AWS datasources
  where the plugin's own backend reads a top-level Grafana field: the legacy
  `settings.Database → Profile` fallback in `configuration.go:16-18`. We carry
  `Database string \`json:"-"\`` on `Config` and populate it from
  `settings.Database` in `LoadConfig` so consumers can call `EffectiveProfile()` and
  match the plugin's runtime behavior exactly.
- **`assumeRoleArn` has no jsonData-based `dependsOn`.** In ConnectionConfig, the field
  is gated on `!hideAssumeRoleArn && awsAssumeRoleEnabled` (`ConnectionConfig.tsx:181,257`);
  neither flag depends on jsonData, and X-Ray does not pass `hideAssumeRoleArn`. The
  field is visible even for `grafana_assume_role` (where Grafana derives its own ARN),
  so encoding a `dependsOn` on `authType` would be inaccurate.
- **`AssumeRoleARN` field vs `assumeRoleArn` tag** — the Go field name mirrors the
  upstream `awsds.AWSDatasourceSettings.AssumeRoleARN`, but the json tag mirrors what
  the frontend writes (`assumeRoleArn`). Go's case-insensitive `encoding/json` accepts
  the backend PascalCase spelling too, and the plugin's own `configuration_test.go:19`
  exercises `assumeRoleARN` from provisioning — a dedicated LoadConfig test locks the
  case-insensitive decode in.
- **`grafana_assume_role` is a schema-listed value** — the editor only renders it when
  the `awsDatasourcesTempCredentials` feature toggle is on AND the plugin is in
  `DS_TYPES_THAT_SUPPORT_TEMP_CREDS` (`ConnectionConfig.tsx:18-28,53-64`); X-Ray IS in
  that allow-list. We list it in `allowedValues` and in the Select options.
- **`arn` kept in `allowedValues`, not in UI options** — `AwsAuthType.ARN` is deprecated
  (`grafana-aws-sdk-react/src/types.ts`) and does not appear in
  `awsAuthProviderOptions`, but stored datasources may still carry it. Listing it in
  `allowedValues` (but not in the Select's `options`) matches the backend's tolerance.
- **`sessionToken` included as a secure key with no UI** — the plugin doesn't offer a
  UI field, but `awsds/settings.go:137` reads it, and CHANGELOG v2.17.0 confirms the
  plugin now needs it for Grafana Assume Role. Tagged `backend-only`.
- **AWS proxy fields excluded** — X-Ray does not pass `showHttpProxySettings` to
  `ConnectionConfig`, so the proxy fields are not part of X-Ray's editor surface.
  They are not in the schema even though the shared `awsds.Load` would technically
  consume them.
- **`SecureJsonDataConfig` is a key list** — secure values are write-only, so the type
  is just the array of secret key names (`accessKey`, `secretKey`, `sessionToken`);
  consumers read `secureJsonFields` to see what is configured.

## SDK plugin schema and k8s-style examples (`schema.go`)

`NewSchema()` assembles the `grafana-plugin-sdk-go` `pluginschema.PluginSchema` bundle
(the k8s-style schema Grafana's datasource API server serves as `{apiVersion}.json`,
`v0alpha1` today) from the embedded `dsconfig.json`: root fields plus a nested
`jsonData` object become the OpenAPI settings `spec`, secure fields become
`secureValues`, and virtual fields (none here) are skipped.

`SettingsExamples()` provides the default configuration plus one k8s-style example per
AWS auth provider, plus one that combines keys auth with an STS AssumeRole, plus one
legacy `arn` example:

| Example | Auth | Extras | `secureJsonData` |
| --- | --- | --- | --- |
| `""` (default) | AWS SDK Default | — | `accessKey` (empty) |
| `awsSdkDefault` | AWS SDK Default | `defaultRegion` | `accessKey` (empty) |
| `accessAndSecretKey` | Access & secret key | `defaultRegion` | `accessKey`, `secretKey` |
| `credentialsFile` | Credentials file | `profile`, `defaultRegion` | `accessKey` (empty) |
| `workspaceIamRole` | Workspace IAM Role | `defaultRegion` | `accessKey` (empty) |
| `grafanaAssumeRole` | Grafana Assume Role | `defaultRegion` | `accessKey` (empty) |
| `assumeRoleFromKeys` | Access & secret key + STS AssumeRole | `assumeRoleArn`, `externalId`, `defaultRegion` | `accessKey`, `secretKey` |
| `legacyArnAuthType` | `arn` (legacy — backend maps to `default`) | `defaultRegion` | `accessKey` (empty) |

Every example carries at least one `secureJsonData` placeholder as required by the
conformance suite.

## `LoadConfig` utility (`settings.go`)

`LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error)`
runs the full three-phase load flow on a datasource instance's settings and returns a
fully-defaulted, validated `Config`:

1. **Parse** — unmarshal `settings.JSONData` into `Config` (Go's case-insensitive
   `encoding/json` accepts either the frontend's camelCase `assumeRoleArn` or the
   backend's PascalCase `assumeRoleARN`), copy the plugin's decrypted secrets
   (`accessKey`/`secretKey`/`sessionToken`) into `DecryptedSecureJSONData`, and capture
   `settings.Database` on the Config's `Database` field for the legacy profile
   fallback. Mirrors `pkg/datasource/configuration.go:8-20`.
2. **`ApplyDefaults`** — fills the single curated default: `AuthType` defaults to
   `AWSAuthTypeDefault`, matching both the reference `aws_sdk_settings.json` pack and
   the backend `awsds.AuthTypeDefault` (iota zero). `DefaultRegion` intentionally has
   no default because it must be picked from the connected AWS account.
3. **`Validate`** — enforces the runtime contract: known `AuthType`, `accessKey` +
   `secretKey` present for `keys` auth, and non-empty `defaultRegion` for any query
   to actually run. Errors are joined so callers see every problem at once.

`EffectiveProfile()` returns `Profile` if set, otherwise `Database` — mirroring the
`getDsSettings` fallback in a callable helper.

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
**does**, not what it **should** do; these notes exist so reviewers can reproduce each
finding and decide separately whether to fix upstream.

1. **Product name diverges from plugin id.** `src/plugin.json:3-4` sets
   `name: "AWS Application Signals"` but keeps `id: "grafana-x-ray-datasource"`. The
   rename happened in v2.16.0 (upstream PR #384). Callers keying off `id` still see
   `grafana-x-ray-datasource`; UI callers keying off `name` see the new brand.
2. **Root-level `database` is silently repurposed as an AWS profile fallback.**
   `pkg/datasource/configuration.go:16-18` reads Grafana's top-level `database` field
   and treats it as the AWS credentials profile name when `jsonData.profile` is empty.
   This is a very old back-compat behavior (X-Ray predates the frontend `profile`
   field) and is not documented anywhere in the config editor. Provisioned datasources
   that happen to set `database` for any other reason will silently pick up an
   unexpected AWS profile.
3. **`XrayJsonData` is an empty extension.** `src/types.ts:106-108` declares
   `interface XrayJsonData extends AwsAuthDataSourceJsonData { // Can add X-Ray specific values here }`.
   The type exists only as a placeholder; it adds no fields. Removing it would be
   safe from a schema standpoint but might change import chains in dependents.
4. **`standardRegions` is a hand-maintained copy of `@grafana/aws-sdk`'s own list.**
   `src/components/ConfigEditor/regions.ts:1-28` re-exports the same 26 regions that
   `@grafana/aws-sdk/src/regions.ts` already exports. Drift between the two lists
   would be silent — the plugin would show a stale region set while other AWS plugins
   would show the SDK's fresh one.
5. **`assumeRoleARN` json tag disagrees with the frontend's `assumeRoleArn`.**
   `awsds/settings.go:98` uses `` json:"assumeRoleARN" `` (uppercase RN), but the
   frontend type (`grafana-aws-sdk-react/src/types.ts`) and every `onChange` in
   ConnectionConfig write `assumeRoleArn` (lowercase arn). Go's case-insensitive
   Unmarshal rescues both. The plugin's own `configuration_test.go:19` exercises the
   uppercase form from provisioning.
6. **Deprecated `arn` auth value silently maps to `default`.** `awsds/settings.go:87-88`
   maps any unknown auth type string to `AuthTypeDefault`, comment: "For old 'arn'
   option". A provisioned config with `authType: "arn"` will load as if it were
   `default`, and no warning is surfaced by X-Ray's editor (unlike CloudWatch, which
   shows an `ARN_DEPRECATION_WARNING_MESSAGE` banner).
7. **Legacy `sharedCreds` auth value silently mapped to `credentials`.**
   `awsds/settings.go:75-78` (`case "credentials"` falls through to `sharedCreds`)
   folds both storage values onto the same enum. Not preserved in this schema as an
   allowed value; a datasource stored with `authType: "sharedCreds"` would fail the
   schema's `allowedValues` check but still load fine on the backend.
8. **Auth type default depends on Grafana instance config.** The editor's `useEffect`
   (`ConnectionConfig.tsx:75-90`) picks `awsAllowedAuthProviders[0]`, and prefers
   `grafana_assume_role` if the feature is on. So the "default" auth type new users
   see is Grafana-instance-dependent, not always `default`. The stored schema default
   here is `default` (matching the reference pack and the backend iota zero), which
   may not be what a fresh Grafana Cloud editor writes.
9. **`sessionToken` has no editor UI even though it is required for Grafana Assume
   Role.** Plugin v2.17.0 added sessionToken handling to make `grafana_assume_role`
   work, but `grafana-aws-sdk-react` never exposes an input for `sessionToken`. Users
   who don't provision the secret directly will need to rely on Grafana Cloud's
   internal broker to populate it — otherwise, temporary credentials silently fall
   back to a non-STS flow.
10. **The "Default Region" description has padded backticks.**
    `ConnectionConfig.tsx:377` reads `` `Specify the region, such as for US West
    (Oregon) use ` us-west-2 ` as the region.` `` — note the spaces around
    `us-west-2` inside the backticks. Preserved verbatim in the schema; would render
    inline code with padding in some Markdown renderers.
11. **The "Assume Role ARN" description is multi-line-indented.**
    `ConnectionConfig.tsx:262-264` uses a raw JSX string with template literal
    line breaks and indentation whitespace inside `description`. Preserved verbatim
    in the schema even though the extra spaces would collapse to a single space in
    most Markdown renderers.

## Validation performed

- `dsconfig.ParseAndResolveSchemaJSON` + `Schema.Validate()` — passes.
- JSON Schema validation against [`dsconfig/schema.json`](../../dsconfig/schema.json)
  (draft 2020-12, `additionalProperties: false`) — passes.
- `go test ./...` on this module — passes (schema bundle shape,
  `SchemaSpecHasNoSecureJSON`, `SecureValuesMatchLoadSettings`, `JSONDataMatchesStruct`,
  `JSONDataTypesMatchStruct`, `SchemaArtifactInSync`, `LoadConfig` including
  case-insensitive `assumeRoleARN` decode and the `Database → Profile` legacy
  fallback, `ApplyDefaults`, `Validate` per auth type, `EffectiveProfile`).
- `settings.go` / `schema.go`: `go build`, `go vet`, `gofmt` — clean.
- `settings.ts`: `tsc --noEmit --strict` — clean.
