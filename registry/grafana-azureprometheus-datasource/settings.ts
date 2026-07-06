/**
 * Configuration models for the Azure Monitor Managed Service for Prometheus
 * datasource plugin (`grafana-azureprometheus-datasource`).
 *
 * Sources of truth (https://github.com/grafana/azure-prometheus-datasource
 * @ fe45d2eea9c7d923fbef1a98b8e0be468781525b):
 * - `src/plugin.json:2-6,94-108` — plugin ID (`"grafana-azureprometheus-datasource"`),
 *   name (`"Azure Monitor Managed Service for Prometheus"`), `info.links` is empty.
 * - `src/configuration/ConfigEditor.tsx:20-80` — top-level editor: renders a
 *   `DataSourceDescription` with `docsLink` pointing to the vanilla Prometheus
 *   docs (`// TODO Update this to Azure prom docs when available`), then
 *   composes `DataSourceHttpSettingsOverhaul` (URL + Azure-auth-only
 *   `Auth` block) and a collapsible `ConfigSection` "Advanced settings"
 *   containing `AdvancedHttpSettings`, `AlertingSettingsOverhaul<PromOptions>`,
 *   and `PromSettings`.
 * - `src/configuration/DataSourceHttpSettingsOverhaul.tsx:36-144` — the
 *   `Auth` wrapper with `visibleMethods=[azureAuthId]` (only Azure auth is
 *   selectable), `onAuthMethodSelect` clears `basicAuth`, `withCredentials`
 *   and `oauthPassThru` on every save.
 * - `src/configuration/AzureAuthSettings.tsx:17-43` — wraps
 *   `AzureCredentialsForm` with the `managedIdentityEnabled`,
 *   `workloadIdentityEnabled`, `userIdentityEnabled` flags read from
 *   `config.azure` — those toggle which auth types are visible.
 * - `src/configuration/AzureCredentialsForm.tsx:33-274` — authType select
 *   (`clientsecret`/`msi`/`workloadidentity`/`currentuser`) plus the
 *   App-Registration input triplet (azureCloud, tenantId, clientId, secret)
 *   rendered inline when `authType === 'clientsecret'`.
 * - `src/configuration/AppRegistrationCredentials.tsx:14-153` — the
 *   `AppRegistrationCredentials` sub-editor used inside
 *   `CurrentUserFallbackCredentials` for the `clientsecret` fallback.
 * - `src/configuration/CurrentUserFallbackCredentials.tsx:21-208` — the
 *   `serviceCredentialsEnabled` toggle plus fallback auth-type select for
 *   `authType === 'currentuser'`.
 * - `src/configuration/AzureCredentialsConfig.ts:15-72` — helpers
 *   (`getAzureCloudOptions`, `getDefaultCredentials`, `getCredentials`,
 *   `updateCredentials`, `setDefaultCredentials`, `resetCredentials`) and
 *   the `AzurePromDataSourceOptions` type: `extends PromOptions,
 *   AzureDataSourceJsonData` plus `azureEndpointResourceId?: string` and
 *   `'prometheus-type-migration'?: boolean`.
 * - `pkg/azureauth/azure.go:18-75` — backend
 *   `ConfigureAzureAuthentication`: reads jsonData via `promlib/utils`,
 *   parses credentials via `azcredentials.FromDatasourceData`, derives the
 *   Prometheus OAuth scope from the resolved cloud's `prometheusResourceId`
 *   property, appends `.default`, and installs `AddAzureAuthentication`
 *   with `AllowUserIdentity()`.
 * - `pkg/datasource.go:19-86` — datasource construction: `NewDatasource`
 *   builds a `promlib.Service` whose `extendClientOpts` invokes
 *   `azureauth.ConfigureAzureAuthentication` when `azureSettings.AzureAuthEnabled`.
 *
 * The Prometheus knobs (`httpMethod`, `timeInterval`, `queryTimeout`,
 * `prometheusType`, `prometheusVersion`, `cacheLevel`, `incrementalQuerying`,
 * `incrementalQueryOverlapWindow`, `disableRecordingRules`,
 * `customQueryParameters`, `seriesLimit`, `seriesEndpoint`, `defaultEditor`,
 * `disableMetricsLookup`, `exemplarTraceIdDestinations`, `manageAlerts`,
 * `allowAsRecordingRulesTarget`, `timeout`, `keepCookies`,
 * `maxSamplesProcessedWarningThreshold`, `maxSamplesProcessedErrorThreshold`)
 * come from the shared `@grafana/prometheus` package (`PromOptions` in
 * `packages/grafana-prometheus/src/types.ts`) at version `12.4.2`, and the
 * backend `PromOptions` shape lives in
 * `github.com/grafana/grafana-prometheus-datasource/pkg/promlib/models/settings.go`
 * at `v0.0.12` (the go.mod pin).
 *
 * External components consulted at their pinned versions (from `package.json`
 * at the pinned SHA):
 * - `@grafana/azure-sdk@0.1.0` — `AzureCredentials`, `AzureAuthType`,
 *   `AzureDataSourceJsonData`, `AzureDataSourceSecureJsonData`,
 *   `updateDatasourceCredentials`, `getDatasourceCredentials`,
 *   `getAzureClouds`, `getDefaultAzureCloud`.
 * - `@grafana/plugin-ui@0.13.1` — `Auth`, `AuthMethod`, `ConnectionSettings`,
 *   `convertLegacyAuthProps`, `AdvancedHttpSettings`, `DataSourceDescription`,
 *   `ConfigSection`.
 * - `@grafana/prometheus@12.4.2` — `PromOptions`, `PromSettings`,
 *   `AlertingSettingsOverhaul`, `overhaulStyles`.
 * - `@grafana/ui@12.4.2` — `Input`, `Select`, `Switch`, `TagsInput`, `Alert`,
 *   `SecureSocksProxySettings` (excluded), `TextLink`.
 * - `@grafana/data@12.4.2` — `DataSourcePluginOptionsEditorProps`,
 *   `DataSourceSettings`, `SelectableValue`.
 * - `@grafana/runtime@12.4.2` — `config` (reads `config.azure` and
 *   `config.secureSocksDSProxyEnabled`).
 *
 * The Secure Socks Proxy switch (`jsonData.enableSecureSocksProxy`) is
 * deliberately excluded from this registry entry per AGENTS.md.
 */

/**
 * Azure authentication type union (`AzureAuthType`).
 * Source: `@grafana/azure-sdk` `AzureCredentials.ts:1-9`. Mirrored in
 * `github.com/grafana/grafana-azure-sdk-go/v2/azcredentials/credentials.go`.
 */
export type AzureAuthType =
  | 'currentuser'
  | 'msi'
  | 'workloadidentity'
  | 'clientsecret'
  | 'clientsecret-obo'
  | 'ad-password'
  | 'clientcertificate';

/** Azure cloud identifier. */
export type AzureCloud =
  | 'AzureCloud'
  | 'AzureChinaCloud'
  | 'AzureUSGovernment'
  | 'AzureCustomizedCloud'
  | string;

/**
 * Discriminated-union credential object stored at `jsonData.azureCredentials`.
 * Only the four `authType` values selectable in this plugin's editor
 * (`clientsecret`, `msi`, `workloadidentity`, `currentuser`) are enumerated
 * here; the backend also accepts `clientsecret-obo`, `ad-password`, and
 * `clientcertificate` because it delegates parsing to
 * `grafana-azure-sdk-go/v2/azcredentials.FromDatasourceData`.
 */
export type AzureCredentials =
  | { authType: 'msi' }
  | { authType: 'workloadidentity'; tenantId?: string; clientId?: string }
  | { authType: 'clientsecret'; azureCloud?: AzureCloud; tenantId?: string; clientId?: string }
  | {
      authType: 'currentuser';
      serviceCredentialsEnabled?: boolean;
      serviceCredentials?:
        | { authType: 'msi' }
        | { authType: 'workloadidentity' }
        | { authType: 'clientsecret'; azureCloud?: AzureCloud; tenantId?: string; clientId?: string };
    };

/** Prometheus flavor type ("Prometheus" | "Cortex" | "Mimir" | "Thanos"). */
export type PromApplication = 'Prometheus' | 'Cortex' | 'Mimir' | 'Thanos';

/** Query editor mode ("builder" | "code"). */
export type QueryEditorMode = 'builder' | 'code';

/** Browser query cache level ("Low" | "Medium" | "High" | "None"). */
export type PrometheusCacheLevel = 'Low' | 'Medium' | 'High' | 'None';

/** HTTP method used for Prometheus range/instant queries. Defaults to "POST". */
export type PromHTTPMethod = 'POST' | 'GET';

/**
 * A single exemplar trace ID destination entry. When `datasourceUid` is set
 * the editor treats the exemplar as an internal link and takes precedence
 * over `url`.
 */
export type ExemplarTraceIdDestination = {
  name: string;
  url?: string;
  urlDisplayLabel?: string;
  datasourceUid?: string;
};

/**
 * Root (top-level datasource settings) fields the Azure Prometheus plugin reads.
 *
 * `url` is required and read by the backend (via `promlib`). Basic-auth and
 * cross-site-credentials root fields are not exposed by this plugin's editor
 * — `visibleMethods=[azureAuthId]` locks the auth picker to Azure auth, and
 * `onAuthMethodSelect` clears `basicAuth` / `withCredentials` on every save
 * (`src/configuration/DataSourceHttpSettingsOverhaul.tsx:122-131`).
 * `options.access === 'direct'` (Browser mode) is rejected with an inline
 * banner (`src/configuration/ConfigEditor.tsx:41-48`).
 */
export type RootConfig = {
  /** Complete HTTP URL of the Azure-hosted Prometheus workspace query endpoint. */
  url?: string;
};

/**
 * Fields stored in `jsonData`. Union of the plugin's own `AzurePromDataSourceOptions`
 * (`extends PromOptions, AzureDataSourceJsonData`) plus every field the
 * `@grafana/prometheus` `PromSettings` component writes and every field the
 * `promlib` backend parses.
 */
export type JsonDataConfig = {
  /**
   * Discriminated-union Azure credentials written by `@grafana/azure-sdk`'s
   * `AzureCredentialsForm` (`src/configuration/AzureCredentialsForm.tsx`).
   * The `authType` discriminates which sibling fields are present.
   */
  azureCredentials?: AzureCredentials;

  /**
   * Optional Azure resource ID (audience) override. Defined on
   * `AzurePromDataSourceOptions` (`AzureCredentialsConfig.ts:68`); no editor
   * UI. Backend derives the scope from `azureCloud.prometheusResourceId` when
   * this is unset (`pkg/azureauth/azure.go:58-63`). Backend-only.
   */
  azureEndpointResourceId?: string;

  /**
   * Sentinel flag set to `true` when a vanilla Prometheus data source is
   * migrated to Azure Prometheus. Triggers the migration banner at
   * `DataSourceHttpSettingsOverhaul.tsx:101-117`. Storage key uses a hyphen,
   * hence the string key here.
   */
  'prometheus-type-migration'?: boolean;

  /**
   * Set to `true` by `@grafana/azure-sdk` for `currentuser` auth and by the
   * SDK's own convertLegacyAuthProps path. `DataSourceHttpSettingsOverhaul`
   * always clears this to `false` on save because the plugin's
   * `visibleMethods=[azureAuthId]` never selects OAuthForward.
   */
  oauthPassThru?: boolean;

  /** Manage alert rules for this data source (`AlertingSettingsOverhaul`). */
  manageAlerts?: boolean;
  /** Allow this datasource as a target for writing recording rules. */
  allowAsRecordingRulesTarget?: boolean;

  /** HTTP request timeout in seconds (`AdvancedHttpSettings`). */
  timeout?: number;
  /** Cookies to forward to the datasource, by name (`AdvancedHttpSettings`). */
  keepCookies?: string[];

  /** Scrape interval like `'15s'` (`PromSettings.tsx`). */
  timeInterval?: string;
  /** Prometheus query timeout like `'60s'`. */
  queryTimeout?: string;
  /** Default editor when opening a query — Builder or Code. */
  defaultEditor?: QueryEditorMode;
  /** Disable metrics chooser + metric/label autocomplete. */
  disableMetricsLookup?: boolean;
  /** Prometheus flavor. */
  prometheusType?: PromApplication;
  /** Free-form version string; option list from `PromFlavorVersions[prometheusType]`. */
  prometheusVersion?: string;
  /** Browser cache level for editor queries. */
  cacheLevel?: PrometheusCacheLevel;
  /** Turn on incremental query caching (beta). */
  incrementalQuerying?: boolean;
  /** Duration string; defaults to `'10m'` when `incrementalQuerying` is true. */
  incrementalQueryOverlapWindow?: string;
  /** Disable recording rules (beta). */
  disableRecordingRules?: boolean;
  /** URL query-parameter string appended to Prometheus requests. */
  customQueryParameters?: string;
  /** POST (default) or GET. Backend validates in `pkg/promlib/models/settings.go:92-95`. */
  httpMethod?: PromHTTPMethod;
  /** Series/label endpoint limit; empty = 40000, 0 = no limit. */
  seriesLimit?: number;
  /** Prefer /api/v1/series over /api/v1/label/*&#47;values. */
  seriesEndpoint?: boolean;
  /** Exemplar trace-ID destinations. */
  exemplarTraceIdDestinations?: ExemplarTraceIdDestination[];

  /**
   * Backend-only: parsed by `pkg/promlib/models/settings.go:41`. Not
   * rendered by this plugin's editor (feature-flagged off — `PromSettings`
   * only shows the input when `showQuerySamplesProcessedThresholdFields` is
   * true, and `ConfigEditor.tsx` never passes that prop).
   */
  maxSamplesProcessedWarningThreshold?: number;
  /** Backend-only: see `maxSamplesProcessedWarningThreshold`. */
  maxSamplesProcessedErrorThreshold?: number;
};

/**
 * Secure key names stored in `secureJsonData` (write-only; read existing
 * config via `secureJsonFields`):
 *
 * - `azureClientSecret` — modern app-registration client secret
 *   (`authType === 'clientsecret'` or `currentuser` with a `clientsecret`
 *   fallback).
 * - `clientSecret` — legacy client secret preserved so pre-migration
 *   datasources still authenticate; the backend reads it as a fallback when
 *   `azureClientSecret` is missing.
 */
export type SecureJsonDataConfig = Array<'azureClientSecret' | 'clientSecret'>;
