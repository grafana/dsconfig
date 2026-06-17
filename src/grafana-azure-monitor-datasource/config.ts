export type azureMonitorConfig = {
  jsonData: {
    
    cloudName?: AzureMonitorCloudName;

    
    azureAuthType?: AzureMonitorAuthType;

    
    azureCredentials?: AzureCredentials;

    
    tenantId?: string;

    
    clientId?: string;

    
    subscriptionId?: string;

    
    basicLogsEnabled?: boolean;

    
    azureLogAnalyticsSameAs?: boolean;

    
    logAnalyticsTenantId?: string;

    
    logAnalyticsClientId?: string;

    
    logAnalyticsSubscriptionId?: string;

    
    logAnalyticsDefaultWorkspace?: string;

    
    appInsightsAppId?: string;

    
    enableSecureSocksProxy?: boolean;

    
    timeout?: number;

    
    keepCookies?: string[];

    
    customizedRoutes?: Record<string, AzureMonitorCustomRoute>;
  };

  secureJsonData: {
    
    clientSecret?: string;

    
    azureClientSecret?: string;

    
    appInsightsApiKey?: string;

    
    logAnalyticsClientSecret?: string;
  };
};

export type AzureMonitorCloudName =
  | "azuremonitor"
  | "chinaazuremonitor"
  | "govazuremonitor"
  | "customizedazuremonitor";

export type AzureMonitorAuthType =
  | "currentuser"
  | "msi"
  | "workloadidentity"
  | "clientsecret"
  | "clientsecret-obo"
  | "currentuser"
  | "ad-password"
  | "clientcertificate";

export type AzureCredentials = {
  authType: AzureMonitorAuthType;
  azureCloud?: string;
  tenantId?: string;
  clientId?: string;
  clientSecret?: string;
  
  serviceCredentials?: AzureCredentials;
};

export type AzureMonitorCustomRoute = {
  
  URL: string;
  
  Scopes?: string[];
  
  Headers?: Record<string, string>;
};
