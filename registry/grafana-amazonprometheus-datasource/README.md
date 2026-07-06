# grafana-amazonprometheus-datasource

Declarative configuration schema for the
[Amazon Managed Service for Prometheus datasource plugin](https://github.com/grafana/grafana-amazonprometheus-datasource)
(`grafana-amazonprometheus-datasource`).

## Upstream researched

- **Repo**: `github.com/grafana/grafana-amazonprometheus-datasource`
- **Ref**: `main`
- **Commit SHA**: `34eb30afef47d8550382dd23b99deb81c32471a9` (HEAD at time of
  authoring — `Docs: Updated Amazon Managed Prometheus docs (#752)`)

Every value in [`dsconfig.json`](dsconfig.json) — labels, placeholders,
tooltips, option labels/values, section titles, defaults, validations,
dependency and required-when expressions, storage keys, storage targets,
value types, group titles, and instructions — is traceable to a specific
`file:line` in the upstream repo at this SHA or in a pinned external
component. See [Field provenance](#field-inventory) below.

To reproduce this research:

```bash
git clone --depth 20 https://github.com/grafana/grafana-amazonprometheus-datasource
cd grafana-amazonprometheus-datasource
git checkout 34eb30afef47d8550382dd23b99deb81c32471a9
```

## Files

| File | Purpose |
| --- | --- |
| [`dsconfig.json`](dsconfig.json) | dsconfig v1 schema — single source of truth for all config fields, groups, relationships, and instructions |
| [`settings.ts`](settings.ts) | TypeScript models: `RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig`, plus enum unions |
| [`settings.go`](settings.go) | Go `Config` model (flat: root `URL` tagged `json:"-"`, SigV4 + Prometheus + Amazon jsonData fields, `DecryptedSecureJSONData`), `PluginID`, `SigV4AuthType` / `HTTPMethod` / `PromApplication` / `PrometheusCacheLevel` / `QueryEditorMode` / `SecureJsonDataKey` typed constants, and the `LoadConfig` utility |
| [`schema.go`](schema.go) | k8s-style SDK plugin schema: embeds `dsconfig.json`, produces the `pluginschema.PluginSchema` bundle via `dsconfig.NewSDKSchema`, and defines `SettingsExamples` for each SigV4 auth variant |
| [`settings_test.go`](settings_test.go) | Table tests for `LoadConfig`, `ApplyDefaults`, `Validate`, and default-example shape |
| [`conformance_test.go`](conformance_test.go) | Runs `schema.RunPluginTests` — the shared dsconfig conformance suite |
| [`schema.gen.json`](schema.gen.json), [`settings.gen.json`](settings.gen.json), [`settings.examples.gen.json`](settings.examples.gen.json) | Committed schema artifacts (regenerate with `go generate ./...` inside this directory; the `SchemaArtifactInSync` conformance subtest guards drift) |

There is no per-entry `go.mod` — every registry entry is a subpackage of the
shared [`registry/`](..) module (`github.com/grafana/dsconfig/registry`).

## Sources researched

### Plugin repo (`github.com/grafana/grafana-amazonprometheus-datasource@34eb30a`)

| File | What was read |
| --- | --- |
| `src/plugin.json:2-8,88-108` | `pluginType` (`id` = `"grafana-amazonprometheus-datasource"`), `pluginName` (`name` = `"Amazon Managed Service for Prometheus"`), `info.links[0].url` = `"https://aws.amazon.com/prometheus/"` (used as `docURL`) |
| `src/configuration/ConfigEditor.tsx:17-107` | Top-level editor composition: `Alert` rejecting browser-mode access (`:26-31`), `DataSourceDescription` (`:32-35`), warning `Alert` gated by `jsonData['prometheus-type-migration']` (`:37-48`), `DataSourceHttpSettingsOverhaul` with a `renderSigV4Editor` slot embedding `SIGV4ConnectionConfig` + a manually rendered `Service` `Field`/`Input` for `jsonData.sigv4Service` (`:52-83`), then a collapsible `ConfigSection` "Advanced settings" (`:86-104`) wrapping `AdvancedHttpSettings`, `AlertingSettingsOverhaul<DataSourceOptions>`, and `PromSettings` (`:97-103`) with `hidePrometheusTypeVersion={true}`, `hideExemplars={true}`, `showQuerySamplesProcessedThresholdFields={true}` |
| `src/configuration/DataSourceHttpSettingsOverhaul.tsx:17-153` | The `Auth` wrapper: registers a single `customMethods` entry keyed `sigV4Id = 'custom-sigV4Id'` (`:46-53`), sets `visibleMethods=[sigV4Id]` (`:120`), and `onAuthMethodSelect` (`:103-115`) writes `basicAuth`, `withCredentials`, `jsonData.sigV4Auth`, `jsonData.oauthPassThru` — the only visible method is `sigV4Id`, so all four ultimately settle at `basicAuth=false`, `withCredentials=false`, `sigV4Auth=true`, `oauthPassThru=false`. `useEffectOnce` at `:27-38` forces `jsonData.sigV4Auth = true` on every mount. The `forwardGrafanaUserHeader` `InlineSwitch` is rendered inline (`:122-143`) |
| `src/configuration/DataSourceOptions.ts:1-8` | `DataSourceOptions extends PromOptions` with `'prometheus-type-migration'?: boolean`, `sigV4Auth?: boolean`, `sigv4Service?: string`, `forwardGrafanaUserHeader?: boolean` — note the lowercase `v` in `sigv4Service` |
| `src/configuration/ConfigEditor.tsx:56-79` | The `sigv4Service` `Field`/`Input`: label `"Service"`, description `"Specify the AWS service to sign requests against (e.g., 'aps' for Prometheus)."`, `placeholder="aps"`, `defaultValue="aps"` |
| `pkg/datasource.go:24-46` | `NewDatasource`: parses `jsonData.forwardGrafanaUserHeader` via `promlib/utils.GetJsonData` + `maputil.GetBoolOptional`, builds a `promlib.Service` with `extendClientOpts` |
| `pkg/datasource.go:84-102` | `contextualMiddlewares`: always installs `awsauth.NewSigV4Middleware()` (`:89`); when `forwardGrafanaUser` is set and the incoming request carries `X-Grafana-User`, adds a `forwardHeaderMiddleware` that copies the header onto the upstream request |
| `pkg/datasource.go:117-131` | `extendClientOpts`: reads `jsonData.sigv4Service` via `maputil.GetStringOptional`; when missing or empty, sets `clientOpts.SigV4.Service = "aps"`; otherwise writes the value verbatim |
| `src/module.ts` | Registers `PrometheusDatasource` from `@grafana/prometheus` with a custom `ConfigEditor` — all query/resource/health-check execution is delegated to the shared Prometheus components |

### External components (pinned to `package.json` / `go.mod`)

- **`@grafana/aws-sdk@0.11.0`** (`package.json:76`) — `SIGV4ConnectionConfig`
  (`src/components/SIGV4ConnectionConfig.tsx:11-70`) wraps `ConnectionConfig`
  with `skipHeader` + `skipEndpoint` and maps `authType`/`profile`/
  `assumeRoleArn`/`externalId`/`defaultRegion`/`endpoint` onto their
  `sigV4`-prefixed counterparts. `awsAuthProviderOptions`
  (`src/providers.ts:4-25`) is the fixed option list — labels: `"Workspace IAM Role"`,
  `"Grafana Assume Role"`, `"AWS SDK Default"`, `"Access & secret key"`,
  `"Credentials file"`. `AwsAuthType` values (`src/types.ts:3-13`):
  `default`, `keys`, `credentials`, `ec2_iam_role`, `grafana_assume_role`,
  `arn` (deprecated).
- **`@grafana/plugin-ui@0.16.0`** (`package.json:79`) — `Auth`, `AuthMethod`,
  `ConnectionSettings`, `convertLegacyAuthProps`, `AdvancedHttpSettings`,
  `DataSourceDescription`, `ConfigSection`, `ConfigSubSection`.
- **`@grafana/prometheus@13.1.6`** (`package.json:80`) — `PromOptions`
  (`packages/grafana-prometheus/src/types.ts:33-53`), `PromSettings`,
  `AlertingSettingsOverhaul`, `overhaulStyles`, `docsTip`. Every Prometheus
  jsonData field this schema carries is defined in that `types.ts` and
  rendered by `PromSettings.tsx`.
- **`@grafana/ui@13.0.2`**, **`@grafana/data@13.0.2`**, **`@grafana/runtime@13.0.2`** —
  editor plumbing.
- **`github.com/grafana/grafana-aws-sdk@v1.4.6`** (`go.mod:6`) —
  `awsauth.NewSigV4Middleware`, `awsds.ReadAuthSettings`, `awsds.AuthSettings`.
- **`github.com/grafana/grafana-prometheus-datasource/pkg/promlib@v0.0.12`**
  (`go.mod:8`) — `promlib.Service`, `promlib/utils.GetJsonData`.

## Field inventory

| Schema `id` | Storage key | Target | Editor label | Read by backend? |
| --- | --- | --- | --- | --- |
| `root_url` | `url` | `root` | Prometheus server URL | Yes (via `promlib`) |
| `jsonData_sigV4Auth` | `sigV4Auth` | `jsonData` | — (forced true) | Yes (SigV4 middleware) |
| `jsonData_sigV4AuthType` | `sigV4AuthType` | `jsonData` | Authentication Provider | Yes (SigV4 middleware) |
| `jsonData_sigV4Profile` | `sigV4Profile` | `jsonData` | Credentials Profile Name | Yes |
| `secureJsonData_sigV4AccessKey` | `sigV4AccessKey` | `secureJsonData` | Access Key ID | Yes |
| `secureJsonData_sigV4SecretKey` | `sigV4SecretKey` | `secureJsonData` | Secret Access Key | Yes |
| `jsonData_sigV4AssumeRoleArn` | `sigV4AssumeRoleArn` | `jsonData` | Assume Role ARN | Yes |
| `jsonData_sigV4ExternalId` | `sigV4ExternalId` | `jsonData` | External ID | Yes |
| `jsonData_sigV4Region` | `sigV4Region` | `jsonData` | Default Region | Yes |
| `jsonData_sigv4Service` | `sigv4Service` | `jsonData` | Service | Yes (`pkg/datasource.go:120-129`) |
| `jsonData_forwardGrafanaUserHeader` | `forwardGrafanaUserHeader` | `jsonData` | Forward Grafana User HTTP Header | Yes (`pkg/datasource.go:95-99`) |
| `jsonData_oauthPassThru` | `oauthPassThru` | `jsonData` | — (managed by editor) | Yes (SDK forward) |
| `jsonData_manageAlerts` | `manageAlerts` | `jsonData` | Manage alerts via Alerting UI | Yes (`promlib`) |
| `jsonData_allowAsRecordingRulesTarget` | `allowAsRecordingRulesTarget` | `jsonData` | Allow as recording rules target | Yes |
| `jsonData_timeout` | `timeout` | `jsonData` | Timeout | Yes (SDK) |
| `jsonData_keepCookies` | `keepCookies` | `jsonData` | Allowed cookies | Yes (SDK) |
| `jsonData_timeInterval` | `timeInterval` | `jsonData` | Scrape interval | Yes (`promlib`) |
| `jsonData_queryTimeout` | `queryTimeout` | `jsonData` | Query timeout | Yes (`promlib`) |
| `jsonData_defaultEditor` | `defaultEditor` | `jsonData` | Default editor | Parsed; UI-only |
| `jsonData_disableMetricsLookup` | `disableMetricsLookup` | `jsonData` | Disable metrics lookup | Parsed; UI-only |
| `jsonData_prometheusType` | `prometheusType` | `jsonData` | — (no UI) | Parsed; heuristics |
| `jsonData_prometheusVersion` | `prometheusVersion` | `jsonData` | — (no UI) | Parsed; heuristics |
| `jsonData_cacheLevel` | `cacheLevel` | `jsonData` | Cache level | Parsed; editor caching |
| `jsonData_incrementalQuerying` | `incrementalQuerying` | `jsonData` | Incremental querying (beta) | Parsed |
| `jsonData_incrementalQueryOverlapWindow` | `incrementalQueryOverlapWindow` | `jsonData` | Query overlap window | Parsed |
| `jsonData_disableRecordingRules` | `disableRecordingRules` | `jsonData` | Disable recording rules (beta) | Parsed |
| `jsonData_customQueryParameters` | `customQueryParameters` | `jsonData` | Custom query parameters | Yes (middleware) |
| `jsonData_httpMethod` | `httpMethod` | `jsonData` | HTTP method | Yes |
| `jsonData_seriesLimit` | `seriesLimit` | `jsonData` | Series limit | Parsed |
| `jsonData_maxSamplesProcessedWarningThreshold` | `maxSamplesProcessedWarningThreshold` | `jsonData` | Query warning threshold | Yes |
| `jsonData_maxSamplesProcessedErrorThreshold` | `maxSamplesProcessedErrorThreshold` | `jsonData` | Query error threshold | Yes |
| `jsonData_seriesEndpoint` | `seriesEndpoint` | `jsonData` | Use series endpoint | Parsed |
| `jsonData_exemplarTraceIdDestinations` | `exemplarTraceIdDestinations` | `jsonData` | — (no UI) | Parsed; result transformer |
| `jsonData_prometheusTypeMigration` | `prometheus-type-migration` | `jsonData` | — (banner sentinel) | No (frontend-only banner) |

### Frontend-only settings

- **`prometheus-type-migration`** — sentinel flag toggling the migration
  banner at `ConfigEditor.tsx:37-48`. Nothing in the runtime depends on it;
  the editor reads it and renders a warning.

### Backend-only settings

- **`prometheusType` / `prometheusVersion`** — parsed by `promlib` for
  flavor-specific query heuristics, but this plugin's editor never renders
  the type/version dropdowns (`hidePrometheusTypeVersion={true}` at
  `ConfigEditor.tsx:100`). Provisioning may still set them.
- **`exemplarTraceIdDestinations`** — parsed by `promlib` to emit exemplar
  links from query results, but this plugin's editor never renders the
  exemplar editor (`hideExemplars={true}` at `ConfigEditor.tsx:101`).
  Provisioning-only.
- **`oauthPassThru`** — cleared to `false` on every save by
  `DataSourceHttpSettingsOverhaul.onAuthMethodSelect` because
  `visibleMethods=[sigV4Id]` never selects OAuthForward. Consumed by the
  SDK's shared HTTP client.

### Editor-visible fields unique to this plugin

- **`maxSamplesProcessedWarningThreshold` / `maxSamplesProcessedErrorThreshold`** —
  vanilla Prometheus and Azure Prometheus hide these fields
  (feature-flagged off in `PromSettings`). Amazon Prometheus is the only
  plugin in the registry that passes
  `showQuerySamplesProcessedThresholdFields={true}`
  (`ConfigEditor.tsx:102`), so both threshold inputs are user-editable.

### Excluded settings

- **Secure Socks Proxy** (`jsonData.enableSecureSocksProxy`) — rendered by
  `DataSourceHttpSettingsOverhaul.tsx:145-150` when
  `config.secureSocksDSProxyEnabled` is true. Excluded per AGENTS.md.
- **Basic auth, OAuth Forward, Cross-Site Credentials, and TLS** —
  `visibleMethods=[sigV4Id]` locks the editor to SigV4 auth only, and
  `onAuthMethodSelect` clears `basicAuth` / `withCredentials` /
  `oauthPassThru` on every save. TLS settings are not rendered because
  they live inside the `Auth` component's method sub-panels, and no
  auth method other than SigV4 is visible. These fields can still be set
  via provisioning (SDK consumes them) but are not modeled in this schema.
- **Custom HTTP headers** (`jsonData.httpHeaderName<N>` /
  `secureJsonData.httpHeaderValue<N>`) — not rendered by this plugin's
  editor. Same rationale as basic auth / TLS.
- **SigV4 endpoint override** (`jsonData.sigV4Endpoint`) — `SIGV4ConnectionConfig`
  wraps `ConnectionConfig` with `skipEndpoint`, so the endpoint input is
  never rendered. The field still exists on the AWS SDK type and would be
  read at signing time if a provisioning template set it; not modeled here
  because the plugin's editor never writes it and the AWS SDK-shared
  behavior is documented in `aws_sdk_settings.json`.
- **SigV4 session token** (`secureJsonData.sigV4SessionToken`) — not written
  by `SIGV4ConnectionConfig` (`ConnectionConfig` only writes `accessKey` and
  `secretKey`). Kept out of the schema to avoid modelling a field neither
  the editor nor the plugin's own Go code touches.

## Where the types are defined

### Frontend (TypeScript)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `DataSourceOptions` | `src/configuration/DataSourceOptions.ts:3-8` | plugin |
| `AwsAuthType`, `AwsAuthDataSourceJsonData`, `AwsAuthDataSourceSecureJsonData`, `ConnectionConfigProps` | `src/types.ts:3-51` | `@grafana/aws-sdk` `0.11.0` |
| `PromOptions`, `PromApplication`, `PrometheusCacheLevel`, `ExemplarTraceIdDestination` | `packages/grafana-prometheus/src/types.ts:20-64` | `@grafana/prometheus` `13.1.6` |
| `QueryEditorMode` | `packages/grafana-prometheus/src/querybuilder/shared/types.ts` | `@grafana/prometheus` `13.1.6` |
| `DataSourceJsonData`, `DataSourcePluginOptionsEditorProps`, `DataSourceSettings`, `SelectableValue` | `packages/grafana-data/src/` | `@grafana/data` `13.0.2` |
| `Auth`, `AuthMethod`, `ConnectionSettings`, `convertLegacyAuthProps`, `AdvancedHttpSettings`, `DataSourceDescription`, `ConfigSection`, `ConfigSubSection` | `src/components/ConfigEditor/…` | `@grafana/plugin-ui` `0.16.0` |

### Backend (Go)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `PromOptions` (jsonData mirror), `ExemplarTraceIDDestination` | `packages/grafana-prometheus/src/types.ts:33-64` (frontend contract — no typed Go equivalent in `promlib@v0.0.12` yet; `promlib/utils.GetJsonData` returns a `map[string]any`) | shared |
| `awsds.AuthSettings`, `awsds.ReadAuthSettings`, `awsds.AWSDatasourceSettings` | `pkg/awsds/settings.go` | `github.com/grafana/grafana-aws-sdk` `v1.4.6` |
| `awsauth.NewSigV4Middleware` | `pkg/awsauth/sigv4.go` | `github.com/grafana/grafana-aws-sdk` `v1.4.6` |
| `backend.DataSourceInstanceSettings`, `sdkhttpclient.Options`, `sdkhttpclient.Middleware` | `backend/common.go`, `backend/httpclient/` | `github.com/grafana/grafana-plugin-sdk-go` `v0.292.1` |
| `Datasource`, `NewDatasource`, `contextualMiddlewares`, `extendClientOpts`, `forwardHeaderMiddleware` | `pkg/datasource.go:19-133` | plugin |

The models in this entry flatten that spread into a single Go `Config` type
(root `URL` tagged `json:"-"`, SigV4 + Prometheus + Amazon jsonData fields,
`DecryptedSecureJSONData`) plus a `SecureJsonDataKey` typed constant list.
`settings.ts` keeps the three canonical TypeScript types (`RootConfig`,
`JsonDataConfig`, `SecureJsonDataConfig`).

## Modeling decisions

- **First-class SigV4 fields**: sibling AWS-flavored datasources
  (`elasticsearch`, `grafana-opensearch-datasource`) only model `sigV4Auth`
  as a top-level field and leave the SigV4 sub-fields for provisioning docs.
  Amazon Prometheus is different: SigV4 is the **only** authentication
  mechanism (forced on every mount, all other methods hidden), so every
  `sigV4`-prefixed field the editor writes is modeled explicitly in this
  schema. This gives the API server enough information to render the same
  form the plugin's own editor renders.
- **`sigv4Service` vs `sigV4*` capitalization**: the plugin uses lowercase
  `v` in `sigv4Service` (`DataSourceOptions.ts:6`) but PascalCase-in-camel
  `V` in every other sigV4-prefixed key. This is preserved verbatim in the
  schema key and in the Go struct's json tag — flagged in
  [Upstream findings](#upstream-findings).
- **No basic-auth / TLS / OAuth Forward fields**: locked out by
  `visibleMethods=[sigV4Id]` and cleared on every save. See
  [Excluded settings](#excluded-settings).
- **Prometheus knobs mirrored from vanilla Prometheus**: everything under
  Advanced settings comes from `@grafana/prometheus`'s `PromSettings`
  component. Kept the same labels/descriptions/defaults so tools that
  already understand the Prometheus schema can consume this one without
  re-authoring. The one deviation: Amazon Prometheus surfaces
  `maxSamplesProcessedWarningThreshold` and
  `maxSamplesProcessedErrorThreshold` in the editor
  (`showQuerySamplesProcessedThresholdFields={true}`), so they carry labels
  + placeholders here, unlike the vanilla / Azure entries.
- **`prometheus-type-migration` field ID**: storage key has a hyphen; the
  schema field ID is camelCased (`jsonData_prometheusTypeMigration`) per
  AGENTS.md, and the raw storage key stays on the `key` property.
- **Root fields**: only `url` is carried. Basic-auth root fields are
  omitted because the plugin actively clears them on save; this differs
  from the vanilla Prometheus entry, which carries them because the
  vanilla Prometheus editor uses them.
- **Field ID naming convention**: IDs are prefixed with their storage
  target for discoverability — `root_`, `jsonData_`, or `secureJsonData_` —
  followed by the camelCase form of the storage key.
- **Flat `Config` in Go**: mirrors the frontend `DataSourceOptions` (which
  extends `PromOptions`) plus the SigV4 credential fields; keeps `URL` at
  root with `json:"-"`. The plugin's own backend has no typed `Settings`
  struct — it uses `promlib/utils.GetJsonData` (returning
  `map[string]any`) and reads individual keys via `maputil`. This entry
  consolidates that spread into a single typed struct.
- **`SecureJsonDataConfig` is a key list**: secure values are write-only,
  so the type is just the array of secret key names (`sigV4AccessKey`,
  `sigV4SecretKey`).
- **`LoadConfig` phases**: `parse → ApplyDefaults → Validate`, matching
  the registry-wide convention. `ApplyDefaults` forces `sigV4Auth=true`
  (mirroring `useEffectOnce` in `DataSourceHttpSettingsOverhaul.tsx:27-38`),
  defaults `sigv4Service` to `"aps"` when empty (mirroring
  `pkg/datasource.go:124-129`), and uppercases + defaults `httpMethod` to
  POST. `Validate` enforces URL, HTTP method, `sigV4AuthType` (known
  values only), `sigV4Region` (required — signer needs it), and per-authType
  secret requirements.

## Settings examples matrix

`SettingsExamples()` (`schema.go`) provides:

| Example key | authType | Secret keys | Notes |
| --- | --- | --- | --- |
| `""` (default) | none | `sigV4AccessKey`, `sigV4SecretKey` (both empty) | Bare defaults; the operator picks an authType + region |
| `ec2IamRole` | `ec2_iam_role` | — | Workspace IAM role — no secret required |
| `accessKeys` | `keys` | `sigV4AccessKey`, `sigV4SecretKey` | Static IAM user credentials |
| `credentialsFile` | `credentials` | — | Named profile in `~/.aws/credentials` |
| `awsSdkDefault` | `default` | — | AWS SDK default credential chain |
| `grafanaAssumeRole` | `grafana_assume_role` | — | Grafana Cloud STS broker |
| `assumeRoleCrossAccount` | `keys` + AssumeRole ARN + External ID | `sigV4AccessKey`, `sigV4SecretKey` | Cross-account role chaining |
| `forwardGrafanaUserHeader` | `ec2_iam_role` | — | `forwardGrafanaUserHeader=true` |
| `migratedFromPrometheus` | `ec2_iam_role` | — | `prometheus-type-migration=true` + `prometheusType`/`prometheusVersion` |

## Upstream findings

Potential bugs, misleading UX, and consistency issues discovered while
researching upstream. Preserved verbatim in the schema — the schema records
what the plugin **does**, not what it **should** do; these notes exist so
reviewers can reproduce each finding and decide separately whether to fix
upstream.

1. **`sigv4Service` casing inconsistency** — `DataSourceOptions.ts:6`
   spells the key `sigv4Service` (lowercase `v`), but every other
   sigV4-prefixed key in the same file and in
   `@grafana/aws-sdk` uses PascalCase-in-camel (`sigV4Auth`, `sigV4AuthType`,
   `sigV4AccessKey`, ...). Reads in `pkg/datasource.go:124` (`maputil.GetStringOptional(jsonData, "sigv4Service")`)
   are case-sensitive, so provisioning that used `sigV4Service` (natural
   guess) would silently fall back to the `"aps"` default. Preserved
   verbatim.
2. **`useEffectOnce` re-writes `jsonData` on first mount** —
   `DataSourceHttpSettingsOverhaul.tsx:27-38` unconditionally calls
   `onOptionsChange` with `sigV4Auth: true` on every editor mount, even
   when the stored value is already `true`. This dirties the datasource
   on every editor open and can trigger auto-save side effects in some
   Grafana UIs.
3. **Auth picker locked to SigV4 but state is still cleared on save** —
   `DataSourceHttpSettingsOverhaul.onAuthMethodSelect` (`:103-115`) is a
   full state reset that always writes `basicAuth: false`,
   `withCredentials: false`, and `oauthPassThru: false`, even though the
   only selectable method is `sigV4Id`. Provisioning that sets any of
   those fields will see them cleared the first time an operator saves
   in the UI.
4. **`prometheus-type-migration` is a magic string** — the sentinel key
   contains a hyphen and lives in `jsonData` alongside camelCased fields.
   Storage-format-wise it works, but it's an outlier that scripts scanning
   `jsonData` for known keys may miss.
5. **`DataSourceDescription.docsLink` points at the marketplace page, not
   AWS docs** — `ConfigEditor.tsx:34` sets
   `docsLink="https://grafana.com/grafana/plugins/grafana-amazonprometheus-datasource/"`
   while `plugin.json.info.links[0].url` points at
   `https://aws.amazon.com/prometheus/`. Two different links coexist. This
   entry picks the `plugin.json` URL for `docURL`, matching the
   AGENTS.md rule.
6. **`SigV4ConnectionConfig` writes `sigV4Endpoint` but the field is
   `skipEndpoint` in the editor** — `SIGV4ConnectionConfig.tsx:26` maps
   `endpoint` → `sigV4Endpoint` unconditionally when passing through the
   `ConnectionConfig` `onOptionsChange`, but the editor is instantiated
   with `skipEndpoint={true}` (`SIGV4ConnectionConfig.tsx:67`) so the
   endpoint input is never rendered. Provisioning-set values still round-
   trip through the editor, but the plugin's own Go code
   (`pkg/datasource.go:117-131`) never reads `sigV4Endpoint` — only
   `sigv4Service` is consumed from jsonData at client-construction time.
   The signer's endpoint therefore comes from the resolved `root.url`
   only.
7. **`awsauth.NewSigV4Middleware()` is installed unconditionally** —
   `pkg/datasource.go:89` always appends the SigV4 middleware to the
   context, regardless of `jsonData.sigV4Auth`. This is fine in practice
   (the editor forces `sigV4Auth = true`) but means the middleware would
   sign requests even for a hypothetical provisioning payload with
   `sigV4Auth = false`. Coupled with the `useEffectOnce` force-true above,
   there is no realistic path to an unsigned request.
8. **No health check for AWS credential validity at save time** — the
   plugin relies on the shared Prometheus `CheckHealth` (`promlib`). A
   misconfigured region or an unrecognised `sigV4AuthType` produces a
   generic Prometheus error at query time rather than a specific SigV4
   configuration error at save.

## Validation performed

- `dsconfig.ParseAndResolveSchemaJSON` + `Schema.Validate()` — passes.
- JSON Schema validation against [`dsconfig/schema.json`](../../dsconfig/schema.json)
  (draft 2020-12, strict) — passes via conformance tests.
- `go generate ./...` inside this directory regenerates the schema
  artifacts cleanly.
- `go test ./...` in the shared `registry/` module — passes
  (`SchemaConformance`, `LoadConfig` incl. every SigV4 auth variant,
  malformed input, and default-example shape, `ApplyDefaults`,
  `Validate`).
- `go build`, `go vet`, `gofmt` on this package — clean.
