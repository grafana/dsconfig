/**
 * Configuration models for the Azure Data Explorer datasource plugin
 * (`grafana-azure-data-explorer-datasource`).
 *
 * Sources of truth (https://github.com/grafana/azure-data-explorer-datasource
 * @ febca70fe0814596ffb8a7e399d9dc62c2196e0b):
 * - `src/plugin.json:1-61` — plugin ID, name, docs URLs.
 * - `src/types/index.ts:113-135` — `AdxDataSourceOptions` (extends
 *   `AzureDataSourceJsonData` from `@grafana/azure-sdk`) and
 *   `AdxDataSourceSecureOptions` (extends `AzureDataSourceSecureJsonData`).
 * - `src/components/ConfigEditor/index.tsx:37-215` — top-level editor.
 * - `src/components/ConfigEditor/AzureCredentialsForm.tsx:31-311` — authType
 *   select and per-authType field rendering.
 * - `src/components/ConfigEditor/AzureCredentialsConfig.ts:13-74` — the
 *   frontend credential loader (default + legacy fallback).
 * - `src/components/ConfigEditor/ConnectionConfig.tsx:16-36` — Default
 *   cluster URL input.
 * - `src/components/ConfigEditor/QueryConfig.tsx:16-141` — query timeout,
 *   caching, data consistency, default editor mode.
 * - `src/components/ConfigEditor/DatabaseConfig.tsx:135-234` — Default
 *   database + schema mapping list.
 * - `src/components/ConfigEditor/ApplicationConfig.tsx:13-37` — Application
 *   name input.
 * - `src/components/ConfigEditor/TrackingConfig.tsx:13-45` — Send username
 *   header switch.
 * - `pkg/azuredx/models/settings.go:18-95` — backend `DatasourceSettings.Load`.
 * - `pkg/azuredx/adxauth/adxcredentials/builder.go:12-121` —
 *   `FromDatasourceData` (modern-then-legacy credential parse),
 *   `getFromLegacy`, `ensureOnBehalfOfSupported`, `resolveLegacyCloudName`.
 * - `pkg/azuredx/datasource.go:34-76` — `NewDatasource` (reads
 *   `DecryptedSecureJSONData["OpenAIAPIKey"]`).
 * - `pkg/azuredx/resource_handler.go:25-89` — OpenAI API key consumption.
 *
 * External:
 * - `@grafana/azure-sdk` `0.1.0` (`package.json:103`):
 *   - `src/credentials/AzureCredentials.ts:1-91` — the `AzureAuthType` /
 *     `AzureCredentials` union.
 *   - `src/credentials/AzureCredentialsConfig.ts` —
 *     `updateDatasourceCredentials` (secret writer) sets
 *     `oauthPassThru = true` for `clientsecret-obo`.
 *   - `src/settings.ts:5-28` — `AzureDataSourceJsonData` +
 *     `AzureDataSourceSecureJsonData`.
 * - `@grafana/plugin-ui` `0.13.1` (`package.json:107`):
 *   - `DataSourceDescription`, `ConfigSection`, `ConfigSubSection` — layout.
 * - `github.com/grafana/grafana-azure-sdk-go/v2` `v2.4.1` (`go.mod:8`):
 *   - `azcredentials/credentials.go` — the `AzureAuthType` constants the
 *     backend recognises.
 *   - `azcredentials/builder.go` — `FromDatasourceData` + per-authType parse.
 */

/**
 * Azure authentication type union (subset of `AzureAuthType`) exposed by
 * this plugin's `AzureCredentialsForm`. The shared SDK also declares
 * `clientcertificate`, `ad-password` and `currentuser` but the ADX editor
 * only builds options for the five values below (`AzureCredentialsForm.tsx:31-71`).
 */
export type AzureAuthType = 'currentuser' | 'msi' | 'workloadidentity' | 'clientsecret' | 'clientsecret-obo';

/** Azure cloud identifier. */
export type AzureCloud = 'AzureCloud' | 'AzureChinaCloud' | 'AzureUSGovernment' | string;

/**
 * Discriminated-union credential shape stored at
 * `jsonData.azureCredentials`. Mirrors the subset of `AzureCredentials`
 * ADX's `AzureCredentialsForm` writes.
 */
export type AzureCredentials =
  | { authType: 'msi' }
  | { authType: 'workloadidentity' }
  | { authType: 'currentuser' }
  | { authType: 'clientsecret'; azureCloud?: AzureCloud; tenantId?: string; clientId?: string }
  | { authType: 'clientsecret-obo'; azureCloud?: AzureCloud; tenantId?: string; clientId?: string };

/**
 * Editor-local schema mapping row (`src/types/index.ts:90-96`). Written by
 * `DatabaseConfig` into `jsonData.schemaMappings[]`.
 */
export interface SchemaMapping {
  type: 'function' | 'table' | 'materializedView';
  value: string;
  name: string;
  database: string;
  displayName: string;
}

/**
 * Root (top-level datasource settings) fields the ADX plugin reads.
 * The ADX backend authenticates via `jsonData.azureCredentials` + secure
 * secrets and never inspects `settings.URL` / `settings.User` /
 * `settings.BasicAuthEnabled` — this is the empty-object case AGENTS.md
 * requires (never `null`).
 */
export type RootConfig = Record<string, never>;

/**
 * Fields stored in `jsonData`. Union of every field the editor writes plus
 * every field the backend reads.
 */
export type JsonDataConfig = {
  /**
   * Default Kusto cluster URL. Backend sanitizes it via
   * `pkg/azuredx/helpers.SanitizeClusterUri` and stores the result on
   * `DatasourceSettings.ClusterURL`.
   */
  clusterUrl?: string;

  /**
   * Discriminated-union credentials written by `@grafana/azure-sdk`'s
   * `AzureCredentialsForm`. Shape depends on `authType`.
   */
  azureCredentials?: AzureCredentials;

  /**
   * Application name to be displayed in ADX. Passed as the client
   * application identifier in Kusto requests (`pkg/azuredx/datasource.go:181`).
   */
  application?: string;

  /** Default database used when a query does not specify one. */
  defaultDatabase?: string;

  /** Client query timeout as a Go duration string (e.g. `30s`); default 30s. */
  queryTimeout?: string;

  /** Enable per-query dynamic caching. */
  dynamicCaching?: boolean;

  /** Cache max age as a Kusto timespan (e.g. `0m`); cache disabled when unset/zero. */
  cacheMaxAge?: string;

  /** Kusto query consistency; defaults to `strongconsistency`. */
  dataConsistency?: 'strongconsistency' | 'weakconsistency';

  /**
   * Default query editor mode. `visual` and `raw` are the editor-selectable
   * values; `openai` is also declared on `EditorMode`
   * (`src/types/index.ts:50-54`) but the ADX config editor does not offer
   * it as a default.
   */
  defaultEditorMode?: 'visual' | 'raw' | 'openai';

  /** Enable user-friendly schema mapping. */
  useSchemaMapping?: boolean;

  /** Schema mapping rows managed by `DatabaseConfig`. */
  schemaMappings?: Array<Partial<SchemaMapping>>;

  /**
   * When enabled, Grafana forwards the signed-in user's login in the
   * `x-ms-user-id` and `x-ms-client-request-id` headers
   * (`pkg/azuredx/datasource.go:155-161`).
   */
  enableUserTracking?: boolean;

  /**
   * Cookies to forward through Grafana's proxy. Written by the inline
   * `TagsInput` in `ConfigEditor/index.tsx:158-164`.
   */
  keepCookies?: string[];

  /**
   * Set to `true` by `@grafana/azure-sdk` when `authType == 'clientsecret-obo'`.
   * Backend requires this to be true for OBO auth
   * (`pkg/azuredx/adxauth/adxcredentials/builder.go:89-98`).
   */
  oauthPassThru?: boolean;

  /**
   * Frontend-only, dead. Declared on `AdxDataSourceOptions`
   * (`src/types/index.ts:115`) but never written by the editor and never
   * read by the backend `DatasourceSettings`.
   */
  minimalCache?: number;

  // --- Legacy top-level credentials (pre-`azureCredentials`) ---
  /**
   * Legacy Azure cloud discriminator (`azuremonitor` / `chinaazuremonitor`
   * / `govazuremonitor`). Read by
   * `pkg/azuredx/adxauth/adxcredentials/builder.go:107-121`.
   */
  azureCloud?: string;
  /**
   * Legacy On-Behalf-Of flag. When true and paired with the legacy
   * top-level tenant/client/secret triple, the backend routes to the OBO
   * code path (`pkg/azuredx/adxauth/adxcredentials/builder.go:71-84`).
   */
  onBehalfOf?: boolean;
  /** Legacy top-level tenant ID (paired with legacy `clientId`/`clientSecret`). */
  tenantId?: string;
  /** Legacy top-level client ID (paired with legacy `tenantId`/`clientSecret`). */
  clientId?: string;
};

/**
 * Secure key names stored in `secureJsonData` (write-only). Consumers read
 * `secureJsonFields.<key>` to check whether a secret is configured.
 *
 * - `azureClientSecret` — modern app-registration client secret
 *   (`authType == 'clientsecret'` or `'clientsecret-obo'`).
 * - `clientSecret` — legacy client secret preserved so pre-migration
 *   datasources still authenticate; the backend reads it as a fallback
 *   when `azureClientSecret` is missing (`pkg/azuredx/adxauth/adxcredentials/builder.go:57`).
 * - `OpenAIAPIKey` — OpenAI API key powering the `askOpenAI` resource
 *   endpoint (`pkg/azuredx/resource_handler.go`). No editor UI; set via
 *   provisioning YAML or the datasource HTTP API.
 */
export type SecureJsonDataConfig = Array<'azureClientSecret' | 'clientSecret' | 'OpenAIAPIKey'>;
