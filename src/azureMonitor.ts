export type azureMonitorConfig = {
  jsonData: {
    /**
     * OPTIONAL.
     *
     * Azure cloud environment to connect to.
     *
     * Values:
     * - "azuremonitor": Azure Public Cloud (default).
     * - "chinaazuremonitor": Azure China (21Vianet).
     * - "govazuremonitor": Azure US Government.
     * - "customizedazuremonitor": Custom / private cloud with user-defined routes.
     *
     * Backend behavior:
     * - Legacy cloud names are resolved to standard Azure SDK cloud names (AzureCloud, AzureChinaCloud, etc.)
     *   via resolveLegacyCloudName().
     *   https://github.com/grafana/grafana/blob/main/pkg/tsdb/azuremonitor/azmoncredentials/builder.go#L33-L63
     *
     * UI hints:
     * - Shown as "Azure Cloud" dropdown in the ConfigEditor; options include Azure, Azure China, Azure US Government.
     *   https://github.com/grafana/grafana/blob/main/public/app/plugins/datasource/azuremonitor/e2e/selectors.ts#L0-L39
     *
     * Provisioning examples:
     * https://github.com/grafana/grafana/blob/main/docs/sources/datasources/azure-monitor/configure/index.md#L199-L209
     */
    cloudName?: AzureMonitorCloudName;

    /**
     * REQUIRED.
     *
     * Azure authentication type. Determines how the datasource authenticates with Azure APIs.
     *
     * Values:
     * - "clientsecret": App Registration with client secret (most common).
     * - "clientcertificate": App Registration with client certificate (PEM or PFX).
     * - "msi": Managed Identity (Grafana must run on an Azure VM/resource with an assigned identity).
     * - "workloadidentity": Workload Identity (for Kubernetes-based deployments with Azure AD federation).
     * - "currentuser": Current User identity (Azure AD passthrough; requires feature toggle).
     * - "ad-password": Azure AD username/password (used by MSSQL datasource; less common for Azure Monitor).
     *
     * Backend behavior:
     * - Parsed in azmoncredentials.FromDatasourceData(); if empty and tenantId+clientId are present,
     *   defaults to "clientsecret" for backward compatibility.
     *   https://github.com/grafana/grafana/blob/main/pkg/tsdb/azuremonitor/azmoncredentials/builder.go#L33-L63
     * - "currentuser" requires the "azureMonitorEnableUserAuth" feature toggle to be enabled;
     *   otherwise backend returns "current user authentication is not enabled for azure monitor".
     *   https://github.com/grafana/grafana/blob/main/pkg/tsdb/azuremonitor/azuremonitor.go#L160-L180
     * - "msi" requires managedIdentityEnabled in Grafana Azure settings.
     *   https://github.com/grafana/grafana/blob/main/public/app/plugins/datasource/azuremonitor/credentials.ts#L65-L74
     *
     * UI hints:
     * - Shown as "Authentication" dropdown in the MonitorConfig section of the ConfigEditor.
     *   https://github.com/grafana/grafana/blob/main/public/app/plugins/datasource/azuremonitor/e2e/selectors.ts#L0-L39
     *
     * Provisioning examples:
     * https://github.com/grafana/grafana/blob/main/pkg/tsdb/azuremonitor/azuremonitor_test.go#L82-L130
     */
    azureAuthType?: AzureMonitorAuthType;

    /**
     * OPTIONAL.
     *
     * Structured Azure credentials object (new format). When present, takes precedence over
     * the legacy flat fields (azureAuthType, cloudName, tenantId, clientId).
     *
     * Backend behavior:
     * - Parsed by azcredentials.FromDatasourceData() from the @grafana/azure-sdk-go package.
     *   https://github.com/grafana/grafana/blob/main/pkg/tsdb/azuremonitor/azmoncredentials/builder.go#L12-L31
     * - Falls back to legacy fields if this is not set.
     *   https://github.com/grafana/grafana/blob/main/public/app/plugins/datasource/azuremonitor/credentials.ts#L12-L31
     *
     * UI hints:
     * - The ConfigEditor stores credentials in this structured format when using the new credential UI.
     *   https://github.com/grafana/grafana/blob/main/public/app/plugins/datasource/azuremonitor/datasource.ts#L64-L82
     */
    azureCredentials?: AzureCredentials;

    /**
     * CONDITIONALLY REQUIRED: for azureAuthType "clientsecret" or "clientcertificate" (legacy format).
     *
     * Azure AD Directory (tenant) ID. GUID identifying the Microsoft Entra ID tenant.
     *
     * Backend behavior:
     * - Read from legacy credentials format; if azureAuthType is empty but tenantId and clientId are set,
     *   the backend infers "clientsecret" auth.
     *   https://github.com/grafana/grafana/blob/main/pkg/tsdb/azuremonitor/azmoncredentials/builder.go#L33-L63
     * - Validated on the frontend; missing tenantId errors with "The Tenant Id field is required."
     *   https://github.com/grafana/grafana/blob/main/public/app/plugins/datasource/azuremonitor/azure_monitor/azure_monitor_datasource.ts#L258-L277
     *
     * UI hints:
     * - Shown as "Directory (tenant) ID" input field.
     *   https://github.com/grafana/grafana/blob/main/public/app/plugins/datasource/azuremonitor/e2e/selectors.ts#L0-L39
     */
    tenantId?: string;

    /**
     * CONDITIONALLY REQUIRED: for azureAuthType "clientsecret" or "clientcertificate" (legacy format).
     *
     * Azure AD Application (client) ID. GUID for the app registration.
     *
     * Backend behavior:
     * - Required for client secret auth; missing clientId errors with "The Client Id field is required."
     *   https://github.com/grafana/grafana/blob/main/public/app/plugins/datasource/azuremonitor/azure_monitor/azure_monitor_datasource.ts#L258-L277
     *
     * UI hints:
     * - Shown as "Application (client) ID" input field.
     *   https://github.com/grafana/grafana/blob/main/public/app/plugins/datasource/azuremonitor/e2e/selectors.ts#L0-L39
     */
    clientId?: string;

    /**
     * OPTIONAL.
     *
     * Default Azure subscription ID. Pre-selected in the query editor for Azure Monitor and Logs queries.
     *
     * Backend behavior:
     * - Stored in AzureMonitorSettings.SubscriptionId (JSON tag: "subscriptionId").
     *   https://github.com/grafana/grafana/blob/main/pkg/tsdb/azuremonitor/types/types.go#L0-L36
     * - Read by the datasource constructor as defaultSubscriptionId.
     *   https://github.com/grafana/grafana/blob/main/public/app/plugins/datasource/azuremonitor/datasource.ts#L83-L84
     *
     * UI hints:
     * - Shown as "Default Subscription" dropdown; populated by clicking "Load Subscriptions".
     *   https://github.com/grafana/grafana/blob/main/public/app/plugins/datasource/azuremonitor/e2e/selectors.ts#L0-L39
     */
    subscriptionId?: string;

    /**
     * OPTIONAL.
     *
     * Enable Basic Logs queries for Azure Log Analytics. When enabled, allows querying
     * tables configured with the Basic Logs plan at reduced cost.
     *
     * Backend behavior:
     * - Read by the AzureMonitorDatasource constructor.
     *   https://github.com/grafana/grafana/blob/main/public/app/plugins/datasource/azuremonitor/azure_monitor/azure_monitor_datasource.ts#L57-L74
     */
    basicLogsEnabled?: boolean;

    /**
     * DEPRECATED: legacy Azure Logs credentials — use azureCredentials instead.
     *
     * When true, Azure Logs uses the same credentials as Azure Monitor.
     *
     * https://github.com/grafana/grafana/blob/main/public/app/plugins/datasource/azuremonitor/types/types.ts#L23-L57
     */
    azureLogAnalyticsSameAs?: boolean;

    /**
     * DEPRECATED: legacy Azure Logs credentials — use azureCredentials instead.
     *
     * Tenant ID for Azure Log Analytics (when using separate credentials).
     *
     * https://github.com/grafana/grafana/blob/main/public/app/plugins/datasource/azuremonitor/types/types.ts#L23-L57
     */
    logAnalyticsTenantId?: string;

    /**
     * DEPRECATED: legacy Azure Logs credentials — use azureCredentials instead.
     *
     * Client ID for Azure Log Analytics (when using separate credentials).
     *
     * https://github.com/grafana/grafana/blob/main/public/app/plugins/datasource/azuremonitor/types/types.ts#L23-L57
     */
    logAnalyticsClientId?: string;

    /**
     * DEPRECATED: legacy Azure Logs credentials — use azureCredentials instead.
     *
     * Subscription ID for Azure Log Analytics (when using separate credentials).
     *
     * https://github.com/grafana/grafana/blob/main/public/app/plugins/datasource/azuremonitor/types/types.ts#L23-L57
     */
    logAnalyticsSubscriptionId?: string;

    /**
     * DEPRECATED: legacy Azure Logs credentials — use azureCredentials instead.
     *
     * Default Log Analytics workspace ID.
     *
     * Backend behavior:
     * - Stored in AzureMonitorSettings.LogAnalyticsDefaultWorkspace.
     *   https://github.com/grafana/grafana/blob/main/pkg/tsdb/azuremonitor/types/types.go#L0-L36
     *
     * https://github.com/grafana/grafana/blob/main/public/app/plugins/datasource/azuremonitor/types/types.ts#L23-L57
     */
    logAnalyticsDefaultWorkspace?: string;

    /**
     * OPTIONAL.
     *
     * Application Insights App ID (classic; pre-workspace-based).
     *
     * Backend behavior:
     * - Stored in AzureMonitorSettings.AppInsightsAppId.
     *   https://github.com/grafana/grafana/blob/main/pkg/tsdb/azuremonitor/types/types.go#L0-L36
     *
     * https://github.com/grafana/grafana/blob/main/public/app/plugins/datasource/azuremonitor/types/types.ts#L23-L57
     */
    appInsightsAppId?: string;

    /**
     * OPTIONAL.
     *
     * Enable the Grafana secure SOCKS datasource proxy for this datasource.
     *
     * UI hints:
     * - Shown only when config.secureSocksDSProxyEnabled is true on the Grafana instance.
     *   https://github.com/grafana/grafana/blob/main/public/app/plugins/datasource/azuremonitor/components/ConfigEditor/ConfigEditor.tsx#L0-L34
     *
     * https://github.com/grafana/grafana/blob/main/public/app/plugins/datasource/azuremonitor/types/types.ts#L23-L57
     */
    enableSecureSocksProxy?: boolean;

    /**
     * OPTIONAL.
     *
     * HTTP request timeout in seconds for Azure API calls.
     *
     * https://github.com/grafana/grafana/blob/main/public/app/plugins/datasource/azuremonitor/types/types.ts#L23-L57
     */
    timeout?: number;

    /**
     * OPTIONAL.
     *
     * List of cookie names to forward to the Azure backend.
     *
     * https://github.com/grafana/grafana/blob/main/public/app/plugins/datasource/azuremonitor/types/types.ts#L23-L57
     */
    keepCookies?: string[];

    /**
     * OPTIONAL.
     *
     * Custom routes for customized cloud environments. Maps route names (e.g., "Azure Monitor",
     * "Azure Log Analytics") to custom endpoint configurations.
     *
     * Backend behavior:
     * - Parsed from AzureMonitorCustomizedCloudSettings when cloudName is "customizedazuremonitor".
     *   https://github.com/grafana/grafana/blob/main/pkg/tsdb/azuremonitor/types/types.go#L0-L36
     * - Each route has URL, Scopes, and Headers fields.
     *   https://github.com/grafana/grafana/blob/main/pkg/tsdb/azuremonitor/azuremonitor_test.go#L106-L130
     */
    customizedRoutes?: Record<string, AzureMonitorCustomRoute>;
  };

  secureJsonData: {
    /**
     * CONDITIONALLY REQUIRED: for azureAuthType "clientsecret" (legacy format).
     *
     * Azure AD client secret for App Registration authentication.
     *
     * Backend behavior:
     * - Read from DecryptedSecureJSONData["clientSecret"] in legacy credential loading.
     *   https://github.com/grafana/grafana/blob/main/pkg/tsdb/azuremonitor/azmoncredentials/builder.go#L92-L115
     * - If empty when authType is "clientsecret", errors with "clientSecret must be set".
     *
     * UI hints:
     * - Rendered as a SecretInput field.
     *   https://github.com/grafana/grafana/blob/main/public/app/plugins/datasource/azuremonitor/e2e/selectors.ts#L0-L39
     *
     * Provisioning examples:
     * https://github.com/grafana/grafana/blob/main/pkg/tests/api/azuremonitor/azuremonitor_test.go#L55-L80
     */
    clientSecret?: string;

    /**
     * CONDITIONALLY REQUIRED: for azureAuthType "clientsecret" (new azureCredentials format).
     *
     * Azure AD client secret stored under the new structured credentials key.
     *
     * Backend behavior:
     * - Read from DecryptedSecureJSONData["azureClientSecret"] by the @grafana/azure-sdk-go credential parser.
     *   https://github.com/grafana/grafana/blob/main/pkg/tsdb/azuremonitor/azmoncredentials/builder_test.go#L122-L144
     */
    azureClientSecret?: string;

    /**
     * OPTIONAL.
     *
     * Application Insights API key (classic; pre-workspace-based).
     *
     * https://github.com/grafana/grafana/blob/main/public/app/plugins/datasource/azuremonitor/types/types.ts#L23-L57
     */
    appInsightsApiKey?: string;

    /**
     * DEPRECATED: legacy Azure Logs credentials — use azureCredentials instead.
     *
     * Client secret for Azure Log Analytics (when using separate credentials).
     *
     * https://github.com/grafana/grafana/blob/main/public/app/plugins/datasource/azuremonitor/credentials.ts#L35-L58
     */
    logAnalyticsClientSecret?: string;
  };
};

/**
 * Azure cloud environment name (legacy Azure Monitor format).
 *
 * https://github.com/grafana/grafana/blob/main/pkg/tsdb/azuremonitor/azmoncredentials/builder.go#L33-L63
 */
export type AzureMonitorCloudName =
  | "azuremonitor"
  | "chinaazuremonitor"
  | "govazuremonitor"
  | "customizedazuremonitor";

/**
 * Azure authentication type.
 *
 * https://github.com/grafana/grafana/blob/main/pkg/tsdb/azuremonitor/azmoncredentials/builder.go#L33-L63
 * https://github.com/grafana/grafana/blob/main/public/app/plugins/datasource/azuremonitor/credentials.ts#L65-L74
 */
export type AzureMonitorAuthType =
  | "clientsecret"
  | "clientcertificate"
  | "msi"
  | "workloadidentity"
  | "currentuser"
  | "ad-password";

/**
 * Structured Azure credentials object (new format used by @grafana/azure-sdk).
 *
 * https://github.com/grafana/grafana/blob/main/public/app/plugins/datasource/azuremonitor/datasource.ts#L64-L82
 */
export type AzureCredentials = {
  authType: AzureMonitorAuthType;
  azureCloud?: string;
  tenantId?: string;
  clientId?: string;
  clientSecret?: string;
  /** Fallback service credentials for current-user auth mode. */
  serviceCredentials?: AzureCredentials;
};

/**
 * Custom route configuration for customized Azure cloud environments.
 *
 * https://github.com/grafana/grafana/blob/main/pkg/tsdb/azuremonitor/types/types.go#L0-L36
 */
export type AzureMonitorCustomRoute = {
  /** Custom endpoint URL for this Azure service route. */
  URL: string;
  /** OAuth2 scopes for token acquisition. */
  Scopes?: string[];
  /** Custom HTTP headers to attach to requests. */
  Headers?: Record<string, string>;
};
