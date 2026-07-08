/**
 * Configuration models for the Azure Monitor datasource plugin
 * (`grafana-azure-monitor-datasource`).
 *
 * Sources of truth (https://github.com/grafana/grafana-azure-monitor-datasource
 * @ 87f88ede8122295a7b671420460535d75a4c02bf):
 * - `src/plugin.json:4,146-162` — plugin ID, name, docs URLs.
 * - `src/types/types.ts:30-53` — `AzureMonitorDataSourceJsonData` (extends
 *   `AzureDataSourceJsonData` from `@grafana/azure-sdk`).
 * - `src/types/types.ts:55-57` — `AzureMonitorDataSourceSecureJsonData`
 *   (extends `AzureDataSourceSecureJsonData` from `@grafana/azure-sdk`).
 * - `src/components/ConfigEditor/ConfigEditor.tsx` — top-level editor.
 * - `src/components/ConfigEditor/MonitorConfig.tsx` — composes
 *   `AzureCredentialsForm`, `DefaultSubscription`, `BasicLogsToggle`.
 * - `src/components/ConfigEditor/AzureCredentialsForm.tsx` — authType select
 *   + delegates to `AppRegistrationCredentials` / `CurrentUserFallbackCredentials`.
 * - `src/components/ConfigEditor/AppRegistrationCredentials.tsx` — Azure
 *   Cloud, Tenant ID, Client ID, Client Secret / Certificate.
 * - `src/components/ConfigEditor/CurrentUserFallbackCredentials.tsx` — the
 *   `serviceCredentialsEnabled` toggle plus fallback auth for `currentuser`.
 * - `src/components/ConfigEditor/DefaultSubscription.tsx` — Default
 *   Subscription select (writes `jsonData.subscriptionId`).
 * - `src/components/ConfigEditor/BasicLogsToggle.tsx` — Basic Logs switch.
 * - `src/credentials.ts` — legacy credential detection helper.
 * - `pkg/azuremonitor/types/types.go:28-32` — backend `AzureMonitorSettings`
 *   (SubscriptionId, LogAnalyticsDefaultWorkspace, AppInsightsAppId).
 * - `pkg/azuremonitor/types/types.go:34-37` — backend
 *   `AzureMonitorCustomizedCloudSettings` (CustomizedRoutes).
 * - `pkg/azuremonitor/azmoncredentials/builder.go` — credential parser (uses
 *   the shared `azcredentials.FromDatasourceData` first, falls back to a
 *   legacy top-level `{azureAuthType, cloudName, tenantId, clientId}` reader).
 * - `pkg/azuremonitor/loganalytics/azure-log-analytics-datasource.go:328,415-432`
 *   — backend reads `basicLogsEnabled` and the deprecated
 *   `azureLogAnalyticsSameAs` gate.
 * - `pkg/azuremonitor/routes.go:31-37,93-106` — backend reads
 *   `customizedRoutes` only when the resolved cloud is `AzureCustomizedCloud`.
 *
 * External:
 * - `@grafana/azure-sdk` `0.1.0` (`package.json:76`):
 *   - `src/credentials/AzureCredentials.ts:1-91` — the `AzureAuthType` /
 *     `AzureCredentials` union.
 *   - `src/credentials/AzureCredentialsConfig.ts:244-449` —
 *     `updateDatasourceCredentials` (secret writer), sets `oauthPassThru`
 *     for `clientsecret-obo`/`currentuser` and `disableGrafanaCache` for
 *     `currentuser`, cleans up legacy `{azureAuthType,cloudName,tenantId,clientId}`.
 *   - `src/settings.ts:5-28` — `AzureDataSourceJsonData` +
 *     `AzureDataSourceSecureJsonData`.
 * - `@grafana/plugin-ui` `0.13.1` (`package.json:79`):
 *   - `src/components/ConfigEditor/AdvancedSettings/AdvancedHttpSettings.tsx`
 *     — writes `jsonData.timeout` (number) and `jsonData.keepCookies` (string[]).
 * - `github.com/grafana/grafana-azure-sdk-go/v2` `v2.4.1` (`go.mod`):
 *   - `azcredentials/credentials.go:3-11` — the `AzureAuthType` constants
 *     the backend recognises (`currentuser`, `msi`, `workloadidentity`,
 *     `clientsecret`, `clientcertificate`, `clientsecret-obo`, `ad-password`).
 *   - `azcredentials/builder.go` — `FromDatasourceData` + per-authType parse
 *     (this is where the required secure keys per authType are enforced).
 */

/**
 * Azure authentication type union (`AzureAuthType`).
 * Source: `grafana-azure-sdk-react/src/credentials/AzureCredentials.ts:1-9`.
 * Note: the SDK's union declares `currentuser` twice — preserved here.
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
 * Discriminated-union credential shape stored at
 * `jsonData.azureCredentials`. Mirrors the shape returned by
 * `@grafana/azure-sdk`'s `AzureCredentialsConfig.getDatasourceCredentials` /
 * `updateDatasourceCredentials`. The secret component (clientSecret /
 * clientCertificate / privateKey / certificatePassword / password) never
 * lives on this object at rest — it is written write-only to `secureJsonData`
 * and read back as a `ConcealedSecret` symbol on the frontend.
 */
export type AzureCredentials =
  | { authType: 'msi' }
  | { authType: 'workloadidentity'; tenantId?: string; clientId?: string }
  | { authType: 'clientsecret'; azureCloud?: AzureCloud; tenantId?: string; clientId?: string }
  | { authType: 'clientsecret-obo'; azureCloud?: AzureCloud; tenantId?: string; clientId?: string }
  | {
      authType: 'clientcertificate';
      azureCloud?: AzureCloud;
      tenantId?: string;
      clientId?: string;
      certificateFormat?: 'pem' | 'pfx';
    }
  | { authType: 'ad-password'; userId?: string; clientId?: string }
  | {
      authType: 'currentuser';
      serviceCredentialsEnabled?: boolean;
      serviceCredentials?:
        | { authType: 'msi' }
        | { authType: 'workloadidentity' }
        | { authType: 'clientsecret'; azureCloud?: AzureCloud; tenantId?: string; clientId?: string }
        | {
            authType: 'clientcertificate';
            azureCloud?: AzureCloud;
            tenantId?: string;
            clientId?: string;
            certificateFormat?: 'pem' | 'pfx';
          };
    };

/**
 * Root (top-level datasource settings) fields the Azure Monitor plugin reads.
 * Azure Monitor's backend authenticates via `jsonData.azureCredentials` +
 * secure secrets and never inspects `settings.URL` / `settings.User` /
 * `settings.BasicAuthEnabled` — this is the empty-object case AGENTS.md
 * requires (never `null`).
 */
export type RootConfig = Record<string, never>;

/**
 * Backend `AzRoute` shape stored inside `jsonData.customizedRoutes` — used
 * only when `azureCloud == 'AzureCustomizedCloud'`. Mirrors
 * `pkg/azuremonitor/types/types.go:22-26` `AzRoute`.
 */
export interface AzRoute {
  URL: string;
  Scopes?: string[];
  Headers?: Record<string, string>;
}

/**
 * Fields stored in `jsonData`. Union of every field the editor writes plus
 * every field the backend reads (some are deprecated / frontend-only /
 * backend-only — flagged in the per-field comments).
 */
export type JsonDataConfig = {
  /**
   * Discriminated-union credentials written by `@grafana/azure-sdk`'s
   * `AzureCredentialsForm`. The `authType` discriminates which sibling
   * fields (azureCloud / tenantId / clientId / userId / certificateFormat /
   * serviceCredentialsEnabled / serviceCredentials) are present.
   */
  azureCredentials?: AzureCredentials;

  /** Default Azure subscription ID chosen in the config editor. */
  subscriptionId?: string;

  /**
   * Enable Basic Logs pricing tier for Log Analytics queries. Read by
   * `pkg/azuremonitor/loganalytics/azure-log-analytics-datasource.go:328`.
   */
  basicLogsEnabled?: boolean;

  /**
   * HTTP request timeout in seconds. Written by `AdvancedHttpSettings`
   * (`@grafana/plugin-ui`). Consumed by Grafana's shared HTTP client.
   */
  timeout?: number;

  /**
   * Cookies to forward through Grafana's proxy. Written by
   * `AdvancedHttpSettings` (`@grafana/plugin-ui`).
   */
  keepCookies?: string[];

  /**
   * Set to `true` by `@grafana/azure-sdk` for auth types
   * `clientsecret-obo` and `currentuser` — makes Grafana's shared HTTP
   * client forward the caller's OAuth token.
   */
  oauthPassThru?: boolean;

  /**
   * Set to `true` by `@grafana/azure-sdk` when `currentuser` auth is
   * selected — prevents shared-cache leakage across users.
   * (`grafana-azure-sdk-react/src/credentials/AzureCredentialsConfig.ts:424`.)
   */
  disableGrafanaCache?: boolean;

  /**
   * Backend-only. Map of route name → route override. Only consulted when
   * `azureCloud == 'AzureCustomizedCloud'`. Not rendered by any editor UI.
   * `pkg/azuremonitor/routes.go:31-37,93-106`.
   */
  customizedRoutes?: Record<string, AzRoute>;

  // --- Deprecated Application Insights / Log Analytics fields ---
  /**
   * Backend-only. Still parsed by `pkg/azuremonitor/types/types.go:31`.
   * Marked `@deprecated` in `src/types/types.ts:48`.
   */
  appInsightsAppId?: string;
  /**
   * Backend-only. Still parsed by `pkg/azuremonitor/types/types.go:30`.
   * Marked `@deprecated` in `src/types/types.ts:45`.
   */
  logAnalyticsDefaultWorkspace?: string;
  /**
   * Backend gate: if defined AND not truthy, Log Analytics queries fail
   * hard (`azure-log-analytics-datasource.go:415-432`). Historically stored
   * as a bool; the backend also accepts a string that `strconv.ParseBool`
   * can parse.
   */
  azureLogAnalyticsSameAs?: boolean | string;
  /** Frontend-only. Marked `@deprecated` in `src/types/types.ts:39`. */
  logAnalyticsTenantId?: string;
  /** Frontend-only. Marked `@deprecated` in `src/types/types.ts:41`. */
  logAnalyticsClientId?: string;
  /** Frontend-only. Marked `@deprecated` in `src/types/types.ts:43`. */
  logAnalyticsSubscriptionId?: string;

  // --- Legacy top-level credentials (pre-`azureCredentials`) ---
  /**
   * Legacy top-level auth-type discriminator. Read by
   * `azmoncredentials/builder.go` `getFromLegacy`. Cleared by
   * `updateDatasourceCredentials` on save.
   */
  azureAuthType?: AzureAuthType;
  /**
   * Legacy Azure cloud discriminator (`azuremonitor` / `chinaazuremonitor`
   * / `govazuremonitor` / `customizedazuremonitor`). Also gates
   * `customizedRoutes`.
   */
  cloudName?: string;
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
 *   (`authType == 'clientsecret'`, `'clientsecret-obo'`, or `'currentuser'`
 *   with a `clientsecret` fallback).
 * - `clientCertificate` — PEM text (or base64 PFX) for
 *   `authType == 'clientcertificate'`.
 * - `privateKey` — PEM private key paired with `clientCertificate` when
 *   `certificateFormat == 'pem'`.
 * - `certificatePassword` — PFX bundle password when
 *   `certificateFormat == 'pfx'`.
 * - `password` — Entra password for `authType == 'ad-password'` (no
 *   editor UI today; reachable via provisioning).
 * - `clientSecret` — legacy client secret preserved so pre-migration
 *   datasources still authenticate; the backend reads it as a fallback
 *   when `azureClientSecret` is missing.
 * - `appInsightsApiKey` — deprecated Application Insights API key
 *   preserved on migrated datasources.
 */
export type SecureJsonDataConfig = Array<
  | 'azureClientSecret'
  | 'clientCertificate'
  | 'privateKey'
  | 'certificatePassword'
  | 'password'
  | 'clientSecret'
  | 'appInsightsApiKey'
>;
