export type databricksConfig = {
  jsonData: {
    
    host: string;
    
    httpPath: string;
    
    authType: DatabricksAuthType;
    
    timeout?: string;
    
    retries?: string;
    
    pause?: string;
    
    rows?: string;
    
    retryTimeout?: string;
    
    debug?: boolean;
    
    defaultQueryFormat?: number;
    
    enableUnitySupport?: boolean;
    
    database?: string;
    
    oauthPassThru?: boolean;
    
    tenantId?: string;
    
    clientId?: string;
    
    azureCloud?: string;
    
    enableSecureSocksProxy?: boolean;
  };
  secureJsonData: {
    
    token?: string;
    
    clientSecret?: string;
    
    azureClientSecret?: string;
  };
};

export type DatabricksAuthType = "" | "Pat" | "OauthM2M" | "OauthPT" | "OauthOBO" | "AzureM2M";
