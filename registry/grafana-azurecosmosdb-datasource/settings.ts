/**
 * Configuration models for the Azure Cosmos DB datasource plugin
 * (plugin id: `grafana-azurecosmosdb-datasource`).
 *
 * Sources of truth (https://github.com/grafana/azure-cosmosdb-datasource @ 43dd04b):
 * - `src/plugin.json:5-6` — plugin id (`"grafana-azurecosmosdb-datasource"`),
 *   name (`"Azure Cosmos DB"`), and `info.links[0].url` at `:22`
 *   (`https://grafana.com/docs/plugins/grafana-azurecosmosdb-datasource`).
 * - `src/components/ConfigEditor.tsx:62-82` — the configuration editor:
 *   - `DataSourceDescription` at `:62-66` with
 *     `dataSourceName="Azure CosmosDB"`,
 *     `docsLink="https://grafana.com/grafana/plugins/grafana-azurecosmosdb-datasource/"`,
 *     `hasRequiredFields`. The editor renders no per-field `required` markers
 *     but the backend hard-fails on empty values (see below).
 *   - `<ConfigSection title="Account configuration">` at `:68` containing two
 *     required `<InlineField>`s (both `labelWidth={16}`) that write to
 *     `jsonData` or `secureJsonData`:
 *       - `accountEndpoint` → `jsonData.accountEndpoint`, label
 *         `"Account Endpoint"`, placeholder empty `""`, `<Input width={40}>`
 *         (`ConfigEditor.tsx:69-71`, `:25-31` `onPathChange`).
 *       - `accountKey` → `secureJsonData.accountKey`, label `"Account Key"`,
 *         placeholder `"Account Key"`, `<SecretInput width={40}>` with
 *         `isConfigured={secureJsonFields.accountKey}` and a Reset handler
 *         (`ConfigEditor.tsx:72-81`, `:33-55` `onAPIKeyChange` /
 *         `onResetAPIKey`).
 *   - `secureSocksDSProxyEnabled`-gated `<Switch>` at `:84-111` writes
 *     `jsonData.enableSecureSocksProxy`. Deliberately excluded from this
 *     registry entry per AGENTS.md.
 * - `src/types.ts:23-33` — the frontend config types `CosmosOptions`
 *   (jsonData: `accountEndpoint?`, `enableSecureSocksProxy?`) and
 *   `SecureJsonData` (`accountKey?: string`).
 * - `pkg/plugin/settings.go:10-13` — backend `Settings` struct:
 *   `AccountEndpoint string \`json:"accountEndpoint,omitempty"\`` and
 *   `AccountKey string \`json:"accountKey,omitempty"\``. The AccountKey json
 *   tag is misleading — the backend never unmarshals it from jsonData; it
 *   is populated from `s.DecryptedSecureJSONData["accountKey"]` at
 *   `settings.go:36-39` (see Upstream findings in the entry README).
 * - `pkg/plugin/settings.go:15-23` — `Settings.isValid()`: returns
 *   `DownstreamError(ErrorMessageEmptyAccountEndpoint)` /
 *   `DownstreamError(ErrorMessageEmptyAccountKey)` when either is empty.
 * - `pkg/plugin/settings.go:25-42` — `LoadSettings`: `json.Unmarshal` the
 *   jsonData into a `map[string]any`, pluck out `accountEndpoint`, copy
 *   `DecryptedSecureJSONData["accountKey"]` onto `settings.AccountKey`,
 *   then call `isValid()`.
 * - `pkg/plugin/datasource.go:27-46` — `NewDatasource`: `LoadSettings`,
 *   then `cosmos.NewClient(ctx, s.AccountKey, s.AccountEndpoint, settings)`.
 * - `pkg/cosmos/client.go:23-53` — `NewClient`: builds an
 *   `azcosmos.NewKeyCredential(accountKey)` + `azcosmos.NewClientWithKey(
 *   accountEndpoint, cred, ...)` from the Azure SDK for Go
 *   (`github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos@v1.4.2`). Any
 *   construction failure surfaces as a downstream error
 *   `"failed to create CosmosDB client, validate account key and endpoint"`.
 * - `pkg/plugin/errors.go:5-13` — the fatal error messages
 *   (`ErrorMessageEmptyAccountEndpoint`, `ErrorMessageEmptyAccountKey`,
 *   `ErrorMessageInvalidJSON`, plus query-time errors).
 *
 * External components consulted at their pinned versions:
 * - `@grafana/plugin-ui@^0.10.1` — `DataSourceDescription` (renders the
 *   header block with the docs link and the required-fields note when
 *   `hasRequiredFields` is set) and `ConfigSection` (renders section
 *   titles).
 * - `@grafana/ui@^11.6.7` — `InlineField`, `Input`, `SecretInput`, `Stack`,
 *   `Switch`, `Divider`, `TextLink`, `getTheme` (no storage fields written
 *   directly). `SecretInput` writes `secureJsonData.accountKey` /
 *   `secureJsonFields.accountKey` via the config editor's own handlers.
 * - `@grafana/data@^11.6.7` — `DataSourcePluginOptionsEditorProps`,
 *   `FeatureToggles`, `onUpdateDatasourceJsonDataOptionChecked` (the last
 *   only for the excluded Secure Socks Proxy switch).
 * - `@grafana/runtime@^11.6.7` — the `config` object read at
 *   `ConfigEditor.tsx:84` to decide whether to render the Secure Socks
 *   Proxy switch.
 *
 * The Secure Socks Proxy switch is deliberately excluded from this
 * registry entry (AGENTS.md exclusion for `jsonData.enableSecureSocksProxy`).
 */

/**
 * Root (top-level datasource settings) fields.
 *
 * The Azure Cosmos DB plugin stores every configuration value in
 * `jsonData` / `secureJsonData`; nothing lives at the root level. The
 * backend never reads `settings.URL`, `settings.BasicAuth`, etc. — see
 * `pkg/plugin/settings.go:25-42` (`LoadSettings` only touches `JSONData`
 * and `DecryptedSecureJSONData`) and `pkg/cosmos/client.go:23-52`
 * (the client is built solely from `accountEndpoint` + `accountKey`).
 * So `RootConfig` is a blank object.
 */
export type RootConfig = Record<string, never>;

/**
 * Fields stored in `jsonData`. Matches the plugin's frontend
 * `CosmosOptions` (`src/types.ts:23-26`).
 */
export type JsonDataConfig = {
  /**
   * Azure Cosmos DB account endpoint URI. The Azure portal presents this
   * as the "URI" on the account's Keys blade, typically
   * `https://<account>.documents.azure.com:443/`. Passed verbatim to
   * `azcosmos.NewClientWithKey` at `pkg/cosmos/client.go:47`. Required at
   * both the editor (backend hard-fails via
   * `pkg/plugin/settings.go:16-18` returning
   * `ErrorMessageEmptyAccountEndpoint`).
   */
  accountEndpoint?: string;
  /**
   * Written by `@grafana/ui`'s `Switch` inside the conditional Secure
   * Socks Proxy block (`ConfigEditor.tsx:84-111`) when Grafana's
   * `secureSocksDSProxyEnabled` feature toggle is set. Consumed
   * transparently by the SDK's `settings.HTTPClientOptions(ctx)` at
   * `pkg/cosmos/client.go:29`; the plugin's own Go code never inspects
   * this field by name and the upstream `Settings` struct
   * (`pkg/plugin/settings.go:10-13`) does not carry it. Deliberately
   * excluded from the dsconfig registry entry per AGENTS.md.
   */
  enableSecureSocksProxy?: boolean;
};

/**
 * Secret key names stored in `secureJsonData` (write-only; read existing
 * config via `secureJsonFields`).
 *
 * - `accountKey` — Azure Cosmos DB account master key (primary or
 *   secondary). Wrapped by `azcosmos.NewKeyCredential` at
 *   `pkg/cosmos/client.go:24` and used to sign every Cosmos DB REST
 *   request. Required (`pkg/plugin/settings.go:19-21` returns
 *   `ErrorMessageEmptyAccountKey`).
 */
export type SecureJsonDataConfig = Array<'accountKey'>;
