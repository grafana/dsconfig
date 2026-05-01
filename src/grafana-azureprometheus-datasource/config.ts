export type grafanaAzurepromethousDatasourceConfig = {
  jsonData: {
    
    azureAuthType?: string;
    
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
