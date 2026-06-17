import type { AzureMonitorAuthType } from "../grafana-azure-monitor-datasource/config";

export type grafanaAzurepromethousDatasourceConfig = {
  jsonData: {
    
    azureAuthType?: AzureMonitorAuthType;
    
    cloudName?: string;
    
    tenantId?: string;
    
    clientId?: string;
    
    azureEndpointResourceId?: string;
    
    timeInterval?: string;
    
    httpMethod?: string;
    
    customQueryParameters?: string;
  };
  secureJsonData: {
    
    clientSecret?: string;
  };
};
