# grafana-azurecosmosdb-datasource

Declarative configuration schema for the [Azure Cosmos DB datasource plugin](https://github.com/grafana/azure-cosmosdb-datasource) (`grafana-azurecosmosdb-datasource`).

## Upstream researched

- **Repo**: `github.com/grafana/azure-cosmosdb-datasource`
- **Ref**: `main`
- **Commit SHA**: `43dd04b20de03448dee065756bbbeb8b8c27cacf` (`Updating plugin-ci-workflows (#645)`)

Every value in [`dsconfig.json`](dsconfig.json) — labels, placeholders, section titles,
`requiredWhen` expressions, storage keys, storage targets, value types, group titles, and
instructions — is traceable to a specific `file:line` in the upstream repo at this SHA. See
[Field provenance](#field-provenance) below.

To reproduce this research:

```bash
git clone https://github.com/grafana/azure-cosmosdb-datasource
cd azure-cosmosdb-datasource
git checkout 43dd04b20de03448dee065756bbbeb8b8c27cacf
```

If upstream `main` has advanced past this SHA, re-diff the sources listed under [Sources
researched](#sources-researched) before merging any changes to this entry.

## Files

| File | Purpose |
| --- | --- |
| [`dsconfig.json`](dsconfig.json) | dsconfig v1 schema — single source of truth for all config fields, groups, and instructions |
| [`settings.ts`](settings.ts) | TypeScript models: `RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig` |
| [`settings.go`](settings.go) | Go `Config` model (flat: jsonData fields + `DecryptedSecureJSONData`), `PluginID`, `SecureJsonDataKey` typed constants, and the `LoadConfig` utility |
| [`schema.go`](schema.go) | k8s-style SDK plugin schema: embeds `dsconfig.json`, converts it via `dsconfig.NewSDKSchema`, and defines `SettingsExamples` |
| [`settings_test.go`](settings_test.go) | Table tests for `LoadConfig`, `ApplyDefaults`, and `Validate` |
| [`conformance_test.go`](conformance_test.go) | Runs `schema.RunPluginTests` — the shared dsconfig conformance suite — against `dsconfig.json` |
| `schema.gen.json`, `settings.gen.json`, `settings.examples.gen.json` | Committed schema artifacts (regenerate with `go generate ./...` inside this directory; the `SchemaArtifactInSync` conformance subtest guards drift) |

There is no per-entry `go.mod` — every registry entry is a subpackage of the shared
[`registry/`](..) module (`github.com/grafana/dsconfig/registry`).

## Sources researched

Every source below was read at the pinned upstream SHA
(`43dd04b20de03448dee065756bbbeb8b8c27cacf`), plus external editor components at the exact
versions the plugin's `package.json` pins.

### Plugin repo (`github.com/grafana/azure-cosmosdb-datasource@43dd04b`)

| File | What was read |
| --- | --- |
| `src/plugin.json:5-6` | `pluginType` (`id` = `grafana-azurecosmosdb-datasource`), `pluginName` (`name` = `Azure Cosmos DB`) |
| `src/plugin.json:22` | `docURL` from `info.links[0].url` (`https://grafana.com/docs/plugins/grafana-azurecosmosdb-datasource`) — note this differs slightly from the editor's hard-coded `docsLink` at `ConfigEditor.tsx:64` (`https://grafana.com/grafana/plugins/grafana-azurecosmosdb-datasource/`) |
| `src/components/ConfigEditor.tsx:62-66` | `DataSourceDescription` with `dataSourceName="Azure CosmosDB"`, `docsLink=...`, `hasRequiredFields` — surfaces the "* fields required" note but does not itself mark per-field required indicators |
| `src/components/ConfigEditor.tsx:68` | `<ConfigSection title={'Account configuration'}>` → group title `Account configuration` |
| `src/components/ConfigEditor.tsx:69-71` | `accountEndpoint` `<InlineField label="Account Endpoint" labelWidth={16}>` + `<Input onChange={onPathChange} value={jsonData.accountEndpoint \|\| ''} placeholder="" width={40}>` — placeholder is literally the empty string |
| `src/components/ConfigEditor.tsx:72-81` | `accountKey` `<InlineField label="Account Key" labelWidth={16}>` + `<SecretInput isConfigured={secureJsonFields.accountKey} value={secureJsonData.accountKey \|\| ''} placeholder="Account Key" width={40} onReset={onResetAPIKey} onChange={onAPIKeyChange}>` |
| `src/components/ConfigEditor.tsx:25-31` | `onPathChange` writes to `jsonData.accountEndpoint` |
| `src/components/ConfigEditor.tsx:33-41` | `onAPIKeyChange` writes to `secureJsonData.accountKey` (replaces the whole `secureJsonData`, but only `accountKey` exists) |
| `src/components/ConfigEditor.tsx:43-55` | `onResetAPIKey` sets `secureJsonFields.accountKey=false` and `secureJsonData.accountKey=''` — the standard Reset flow |
| `src/components/ConfigEditor.tsx:84-111` | Grafana-version-gated Secure Socks Proxy `<Switch>` writes `jsonData.enableSecureSocksProxy` — deliberately excluded from this entry |
| `src/types.ts:23-26` | `CosmosOptions` (jsonData): `accountEndpoint?: string`, `enableSecureSocksProxy?: boolean` |
| `src/types.ts:31-33` | `SecureJsonData`: `accountKey?: string` |
| `pkg/plugin/settings.go:10-13` | Backend `Settings` struct: `AccountEndpoint string \`json:"accountEndpoint,omitempty"\`` and `AccountKey string \`json:"accountKey,omitempty"\`` (see Upstream findings for the misleading json tag) |
| `pkg/plugin/settings.go:15-23` | `Settings.isValid()`: returns `DownstreamError(ErrorMessageEmptyAccountEndpoint)` when `AccountEndpoint == ""`, `DownstreamError(ErrorMessageEmptyAccountKey)` when `AccountKey == ""` |
| `pkg/plugin/settings.go:25-42` | `LoadSettings`: `json.Unmarshal` into `map[string]any` (wraps errors as `ErrorMessageInvalidJSON`), extract `accountEndpoint` if string, copy `DecryptedSecureJSONData["accountKey"]` onto `settings.AccountKey`, then `isValid()` |
| `pkg/plugin/errors.go:5-13` | Fatal error messages: `ErrorMessageEmptyAccountEndpoint`, `ErrorMessageEmptyAccountKey`, `ErrorMessageInvalidJSON` (plus query-time errors not covered by this entry) |
| `pkg/plugin/datasource.go:27-46` | `NewDatasource`: `LoadSettings`, then `cosmos.NewClient(ctx, s.AccountKey, s.AccountEndpoint, settings)` — no other settings consumed |
| `pkg/cosmos/client.go:23-53` | `NewClient`: `azcosmos.NewKeyCredential(accountKey)` + `settings.HTTPClientOptions(ctx)` (SDK — the transparent consumer of `enableSecureSocksProxy`) + `azcosmos.NewClientWithKey(accountEndpoint, cred, ...)`. Construction failures surface as downstream error `"failed to create CosmosDB client, validate account key and endpoint"` |
| `pkg/plugin/settings_test.go:11-118` | Confirms happy path, `ErrorMessageInvalidJSON` on `{invalid json`, endpoint-empty and key-empty failures |
| `go.mod:6-7` | Azure SDK for Go dependencies: `github.com/Azure/azure-sdk-for-go/sdk/azcore v1.21.1` and `github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos v1.4.2` |
| `package.json` | External editor component versions (see next table) |

### External editor components

Read at the versions the plugin's `package.json` pins.

| Component | Version | Source consulted | What was read |
| --- | --- | --- | --- |
| `DataSourceDescription`, `ConfigSection` | `@grafana/plugin-ui@^0.10.1` | `github.com/grafana/plugin-ui` around `v0.10.1` | `dataSourceName` header + `docsLink` behavior; `title` prop (no storage fields written) |
| `SecureSocksProxySettings` behavior (excluded) | `@grafana/ui@^11.6.7` | grafana/grafana around `v11.6.7` `packages/grafana-ui/src/components/DataSourceSettings/SecureSocksProxySettings.tsx` | Storage key written: `jsonData.enableSecureSocksProxy` — confirmed and excluded per AGENTS.md |
| `InlineField`, `Input`, `SecretInput`, `Stack`, `Switch`, `Divider`, `TextLink`, `getTheme` | `@grafana/ui@^11.6.7` | grafana/grafana `v11.6.x` `packages/grafana-ui/src/components/` | Prop names (`label`, `labelWidth`, `placeholder`, `value`, `width`, `isConfigured`, `onReset`, `onChange`) — no storage fields written by the components themselves; `SecretInput` writes via the editor's own `onAPIKeyChange` / `onResetAPIKey` handlers |
| `DataSourcePluginOptionsEditorProps`, `FeatureToggles`, `onUpdateDatasourceJsonDataOptionChecked` | `@grafana/data@^11.6.7` | grafana/grafana `v11.6.x` `packages/grafana-data/src/` | The last helper is only used to write the excluded Secure Socks Proxy toggle |
| `config` (feature toggle read) | `@grafana/runtime@^11.6.7` | grafana/grafana `v11.6.x` `packages/grafana-runtime/src/` | `config.featureToggles.secureSocksDSProxyEnabled` — gates the excluded Secure Socks Proxy switch |

## Field provenance

Every dsconfig field, traced from its schema `id` to the upstream `file:line` where each of its
label, placeholder, tooltip, default, storage key, and value type is defined.

| Schema `id` | Storage key | Target | Label source | Placeholder / options / default source | Value type source | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `jsonData_accountEndpoint` | `accountEndpoint` | `jsonData` | `ConfigEditor.tsx:69` (`<InlineField label="Account Endpoint">`) | `ConfigEditor.tsx:70` (`placeholder=""` — literally the empty string) | `Settings.AccountEndpoint string`, `pkg/plugin/settings.go:11`; TS `accountEndpoint?: string`, `src/types.ts:24` | Role `endpoint.baseUrl` because it is the base URI passed to `azcosmos.NewClientWithKey`; `requiredWhen: "true"` because `settings.go:16-18` returns `ErrorMessageEmptyAccountEndpoint` when empty |
| `secureJsonData_accountKey` | `accountKey` | `secureJsonData` | `ConfigEditor.tsx:72` (`<InlineField label="Account Key">`) | `ConfigEditor.tsx:76` (`placeholder="Account Key"`) | `SecureJsonData.accountKey?: string`, `src/types.ts:32`; consumed via `s.DecryptedSecureJSONData["accountKey"]` at `pkg/plugin/settings.go:36-39` and wrapped by `azcosmos.NewKeyCredential` at `pkg/cosmos/client.go:24` | No matching role in the closed vocabulary — the Cosmos DB master key is neither Bearer, Basic, API-key-in-header, nor Azure Blob storage key. Left unroled. `requiredWhen: "true"` because `settings.go:19-21` returns `ErrorMessageEmptyAccountKey` when empty |

## Field inventory summary

| Schema field | Storage key | Target | Editor label | Read by backend? |
| --- | --- | --- | --- | --- |
| `jsonData_accountEndpoint` | `accountEndpoint` | `jsonData` | Account Endpoint | Yes (required) |
| `secureJsonData_accountKey` | `accountKey` | `secureJsonData` | Account Key | Yes (required, wrapped by `azcosmos.NewKeyCredential`) |
| `jsonData_enableSecureSocksProxy` (excluded) | `enableSecureSocksProxy` | `jsonData` | Enable (under "Secure Socks Proxy") | Indirectly (via `settings.HTTPClientOptions(ctx)` at `pkg/cosmos/client.go:29`) — excluded per AGENTS.md |

### Frontend-only settings

None. Both editor-written fields are read by the backend.

### Backend-only settings

None. Every backend-consumed setting has an editor UI, except the excluded Secure Socks Proxy
switch which is Grafana-version-gated and covered by the SDK's shared field pack.

## Where the types are defined

The configuration types are spread across the plugin and its dependencies — some fields and base
types come from libraries/SDKs rather than the plugin itself:

### Frontend (TypeScript)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `CosmosOptions` (jsonData: `accountEndpoint?`, `enableSecureSocksProxy?`), `SecureJsonData` (`accountKey?`) | `src/types.ts:23-33` | plugin ([grafana/azure-cosmosdb-datasource](https://github.com/grafana/azure-cosmosdb-datasource)) |
| `DataSourceJsonData` (base interface `CosmosOptions` extends: `authType?`, `defaultRegion?`, `profile?`, `manageAlerts?`, …) | `packages/grafana-data/src/types/datasource.ts` | `@grafana/data` `11.6.x` (grafana/grafana `v11.6.x`) |
| `DataSourcePluginOptionsEditorProps`, `FeatureToggles`, `onUpdateDatasourceJsonDataOptionChecked` | `packages/grafana-data/src/` | `@grafana/data` `11.6.x` |
| `DataSourceDescription`, `ConfigSection` (no storage fields written) | `src/components/ConfigEditor/`, `src/components/` | `@grafana/plugin-ui` `^0.10.1` |
| `InlineField`, `Input`, `SecretInput`, `Stack`, `Switch`, `Divider`, `TextLink` (no storage fields written) | `packages/grafana-ui/src/components/` | `@grafana/ui` `^11.6.7` |
| Secure Socks Proxy — `SecureSocksProxySettings` writes `jsonData.enableSecureSocksProxy` (excluded from this entry) | `packages/grafana-ui/src/components/DataSourceSettings/SecureSocksProxySettings.tsx` | `@grafana/ui` `^11.6.7` |

### Backend (Go)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `Settings` (`AccountEndpoint`, `AccountKey`), `isValid`, `LoadSettings`, `PluginID`-style constants | `pkg/plugin/settings.go:10-42` | plugin ([grafana/azure-cosmosdb-datasource](https://github.com/grafana/azure-cosmosdb-datasource)) |
| `NewDatasource` (settings wiring, `cosmos.NewClient(ctx, key, endpoint, settings)`) | `pkg/plugin/datasource.go:27-46` | plugin |
| `cosmos.NewClient` (Cosmos DB client construction, `azcosmos.NewKeyCredential`, `azcosmos.NewClientWithKey`, `crossPartitionQueryPolicy`) | `pkg/cosmos/client.go:23-72` | plugin |
| `ErrorMessage*` constants (`EmptyAccountEndpoint`, `EmptyAccountKey`, `InvalidJSON`, plus query-time errors) | `pkg/plugin/errors.go:5-13` | plugin |
| `azcosmos.NewKeyCredential`, `azcosmos.NewClientWithKey` — the actual credential and client constructors the plugin delegates to | `sdk/data/azcosmos` | [`github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos`](https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos) `v1.4.2` |
| `azcore/policy`, `azcore/runtime` — request pipeline used by the `crossPartitionQueryPolicy` policy | `sdk/azcore/policy`, `sdk/azcore/runtime` | [`github.com/Azure/azure-sdk-for-go/sdk/azcore`](https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/azcore) `v1.21.1` |
| `backend.DataSourceInstanceSettings` (carries `JSONData`, `DecryptedSecureJSONData`, plus root fields — none of the root fields are read by this plugin) | `backend/common.go` | `github.com/grafana/grafana-plugin-sdk-go` `v0.281.0` |
| `httpclient.New`, `settings.HTTPClientOptions(ctx)` (SDK HTTP client that transparently consumes `enableSecureSocksProxy`) | `backend/httpclient` | `github.com/grafana/grafana-plugin-sdk-go` `v0.281.0` |
| `backend.DownstreamError`, `errorsource.DownstreamError` (error classification wrapping the plugin's own fatal errors) | `backend`, `experimental/errorsource` | `github.com/grafana/grafana-plugin-sdk-go` `v0.281.0` |

The models in this entry flatten that spread into a single Go `Config` type (jsonData field
`AccountEndpoint` + `DecryptedSecureJSONData` map for the write-only secret) plus a
`SecureJsonDataKey` typed constant list. `settings.ts` keeps the three canonical TypeScript
types (`RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig`); `RootConfig` is a blank object
because the Cosmos DB plugin stores nothing at the root level.

## Modeling decisions

- **No `@grafana/azure-sdk` credentials model.** Despite being an Azure datasource, this plugin
  does **not** use the shared Azure credentials form (`@grafana/azure-sdk`, `AzureCredentials`,
  `jsonData.azureCredentials.*`, `clientsecret` / `msi` / `workloadidentity` etc.). It
  authenticates with a single Azure Cosmos DB account master key (Cosmos-native key
  credential), so [`dsconfig/packs/azure_sdk_settings.json`](../../dsconfig/packs/azure_sdk_settings.json)
  is intentionally **not** referenced. See Azure cousin entries
  ([`grafana-azure-monitor-datasource`](../grafana-azure-monitor-datasource/),
  [`grafana-azure-data-explorer-datasource`](../grafana-azure-data-explorer-datasource/)) for the
  Entra credentials pattern.
- **`RootConfig` is a blank object.** The Cosmos DB plugin stores nothing at the root level.
- **Both fields marked `requiredWhen: "true"`.** The editor does not render per-field required
  markers (there are no `required` props on the `<InlineField>`s at `ConfigEditor.tsx:69,72`),
  but the backend hard-fails on empty values (`pkg/plugin/settings.go:15-23`). `requiredWhen`
  encodes the backend contract, per AGENTS.md.
- **`accountEndpoint` role `endpoint.baseUrl`.** It is the base URI the Azure SDK for Go's
  `azcosmos.NewClientWithKey` uses as the account host.
- **`accountKey` has no role.** The closed role vocabulary
  (`dsconfig/schema.go:540-605`) has no match for a Cosmos DB account master key: it is not a
  Bearer token, not an HTTP Basic password, not an HTTP header value, not an Azure Blob storage
  account key, not an AWS access key. Left unroled rather than misclassified.
- **AccountKey secret only in `DecryptedSecureJSONData`.** The upstream `Settings` struct
  declares both `AccountEndpoint` and `AccountKey` with json tags on the same struct
  (`pkg/plugin/settings.go:10-13`), but the backend never unmarshals `accountKey` from jsonData —
  it is populated from `s.DecryptedSecureJSONData["accountKey"]` at `settings.go:36-39`. Mirroring
  the upstream tag verbatim would put `accountKey` in the jsonData struct-json-tag set, breaking
  the `JSONDataMatchesStruct` conformance test (which expects secure keys to live only in
  `secureJsonData` / `DecryptedSecureJSONData`). This entry stores the secret exclusively in
  `DecryptedSecureJSONData` — see Upstream findings for the mismatch.
- **Placeholder preserved verbatim.** `jsonData_accountEndpoint`'s UI placeholder is the empty
  string `""` because that is what the editor sets (`ConfigEditor.tsx:70` — `placeholder=""`).
- **Secure Socks Proxy excluded.** `jsonData.enableSecureSocksProxy` is deliberately omitted per
  AGENTS.md, even though the editor gates it behind
  `config.featureToggles.secureSocksDSProxyEnabled` at `ConfigEditor.tsx:84`.
- **Field ID naming convention.** IDs are prefixed with their storage target (`jsonData_` /
  `secureJsonData_`) followed by the camelCase storage key, e.g. `jsonData_accountEndpoint`,
  `secureJsonData_accountKey`. The `key` property keeps the plugin's raw storage key.
- **Flat `Config` in Go.** `settings.go` carries the single jsonData field
  (`AccountEndpoint`) plus a `DecryptedSecureJSONData` map for the write-only secret.
- **`SecureJsonDataConfig` is a key list.** Secure values are write-only, so the secure type is
  just the array of secret key names (`accountKey`); consumers read `secureJsonFields` to see
  what is configured.

## SDK plugin schema and k8s-style examples (`schema.go`)

`NewSchema()` assembles the `grafana-plugin-sdk-go` `pluginschema.PluginSchema` bundle (the
k8s-style schema Grafana's datasource API server serves as `{apiVersion}.json`, `v0alpha1`
today) from the embedded `dsconfig.json`: `jsonData` becomes the OpenAPI settings `spec`,
secure fields become `secureValues`.

`SettingsExamples()` provides the default configuration plus one realistic account-key example.
Each example is a full instance-settings object with the plugin configuration nested under
`jsonData` and the write-only account key under `secureJsonData` (placeholder values to be
replaced with real secrets):

| Example | Connection | `secureJsonData` |
| --- | --- | --- |
| `""` (default) | Empty `accountEndpoint`, empty `accountKey` placeholder | `accountKey` (empty) |
| `accountKey` | `https://my-account.documents.azure.com:443/` | `accountKey` (placeholder) |

## `LoadConfig` utility (`settings.go`)

`LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error)`
runs the full three-phase load flow on a datasource instance's settings and returns a
fully-defaulted, validated `Config`:

1. **Parse** — unmarshal jsonData into `Config` (malformed JSONData is a parse error, mirroring
   the upstream `ErrorMessageInvalidJSON` return path at `pkg/plugin/settings.go:27-29`; empty
   JSONData is tolerated and simply leaves `AccountEndpoint` empty for `Validate` to reject),
   then copy decrypted secrets by known key into `DecryptedSecureJSONData`.
2. **`ApplyDefaults`** — no-op. The Cosmos DB plugin applies no editor-parity defaults; both
   `accountEndpoint` and `accountKey` are strictly required and have no sensible zero value.
   Kept exported so callers can compose the three phases uniformly across registry entries.
3. **`Validate`** — enforce the runtime contract: `AccountEndpoint` and the `accountKey` secret
   must be non-empty. Errors are joined so every problem surfaces at once. Mirrors
   `Settings.isValid()` at `pkg/plugin/settings.go:15-23`.

Everything is logged via `backend.Logger.FromContext(ctx)` with `datasource_uid`,
`datasource_name`, and `plugin` labels so log lines carry request context.

### Direct access to individual phases

`(*Config).ApplyDefaults()` and `(Config).Validate() error` are exported separately for callers
that want to compose them themselves (e.g. provisioning preview, schema-example round-trip,
tests that need to distinguish parse-level from policy-level errors).

## Upstream findings

Potential bugs, misleading UX, and consistency issues discovered while researching upstream.
All preserved verbatim in the schema — the schema records what the plugin **does**, not what
it **should** do.

1. **`Settings.AccountKey` has a misleading `json:"accountKey,omitempty"` tag.**
   `pkg/plugin/settings.go:12` — the field is populated from
   `s.DecryptedSecureJSONData["accountKey"]` (`settings.go:36-39`) and never from jsonData,
   but the json tag would leak the secret if someone ever marshalled a `Settings` value with
   the key present. The Config in this entry stores the secret exclusively in
   `DecryptedSecureJSONData` and drops the json-tagged secret field to keep the struct/schema
   parity clean.
2. **`docURL` differs between `plugin.json` and the editor's `docsLink`.**
   `src/plugin.json:22` uses `https://grafana.com/docs/plugins/grafana-azurecosmosdb-datasource`;
   `src/components/ConfigEditor.tsx:64` uses
   `https://grafana.com/grafana/plugins/grafana-azurecosmosdb-datasource/` (marketplace URL).
   Per AGENTS.md this entry uses `plugin.json`'s `info.links[0].url`.
3. **`Account Endpoint` placeholder is empty.** `src/components/ConfigEditor.tsx:70` has
   `placeholder=""` — no hint is shown to the user about the expected URI format
   (`https://<account>.documents.azure.com:443/`). The Azure portal's Keys blade shows this as
   "URI", which would be a reasonable placeholder value.
4. **`Settings.isValid()` returns errors classified as `DownstreamError` regardless of source.**
   `pkg/plugin/settings.go:16-22` wraps both empty-endpoint and empty-key errors in
   `backend.DownstreamError`, which Grafana's alerting/telemetry treats as data-source-side
   problems. An empty configuration is really a user-input problem, not a downstream failure.
5. **Editor shows no required markers even though both fields are required.** The
   `<InlineField>`s at `ConfigEditor.tsx:69,72` don't set `required` — the required-fields
   notice from `DataSourceDescription hasRequiredFields` at `:65` therefore points at asterisks
   that are never actually rendered next to any field. Preserved as `requiredWhen: "true"` in
   the schema to encode the backend contract.
6. **Empty JSONData is silently valid until `isValid`.** `pkg/plugin/settings.go:26-29` calls
   `json.Unmarshal(config.JSONData, &jsonData)` where `jsonData` is `map[string]any`. When
   `config.JSONData` is empty, `json.Unmarshal` returns "unexpected end of JSON input" —
   contradicting the surface behavior a reader might infer from the tolerant `map[string]any`
   assertion. In practice, the editor always writes some jsonData, so this only shows up on
   provisioning / API paths.
7. **The Secure Socks Proxy toggle is stored in jsonData but the plugin's own Go code never
   inspects it by name.** `pkg/cosmos/client.go:29` calls `settings.HTTPClientOptions(ctx)`
   which internally reads `enableSecureSocksProxy`; the plugin does not need any explicit
   handling. Correct behavior, but unusual — most other plugins mirror the field in their
   `Settings` struct for clarity.
8. **Only one auth method exists.** There is no support for Entra ID / OAuth / managed
   identity / workload identity, despite being an Azure datasource. Users needing Entra-based
   auth to Cosmos DB must use a different path (for example the Azure Monitor plugin's Log
   Analytics data reads of Cosmos DB metrics).

## Validation performed

- `dsconfig.ParseAndResolveSchemaJSON` + `Schema.Validate()` (Go validator in this repo) —
  passes.
- JSON Schema validation against [`dsconfig/schema.json`](../../dsconfig/schema.json) (draft
  2020-12, `additionalProperties: false`) — passes.
- `go test ./...` on the `registry/` module — passes (schema bundle shape, secure values,
  examples, `LoadConfig` for happy path, malformed input, missing endpoint, missing key,
  `SchemaArtifactInSync` guard).
- `settings.go` / `schema.go`: `go build`, `go vet`, `gofmt` — clean.
- `settings.ts`: strict TypeScript — clean.
