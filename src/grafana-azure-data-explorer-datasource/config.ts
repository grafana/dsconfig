export type grafanaAzureDataExplorerDatasourceConfig = {
  jsonData: {
    
    clusterUrl: string;
    
    defaultDatabase: string;
    
    azureAuthType?: string;
    
    cloudName?: string;
    
    tenantId?: string;
    
    clientId?: string;
    
    minimalCache?: number;
    
    defaultEditorMode?: string;
    
    queryTimeout?: string;
    
    dataConsistency?: string;
    
    cacheMaxAge?: string;
    
    dynamicCaching?: boolean;
    
    useSchemaMapping?: boolean;
    
    enableUserTracking?: boolean;
    
    application?: string;
    
    enableSecureSocksProxy?: boolean;
  };
  secureJsonData: {
    
    clientSecret?: string;
    
    OpenAIAPIKey?: string;
  };
};
