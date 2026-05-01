export type serviceNowConfig = {
  url: string;
  basicAuth?: boolean;
  basicAuthUser?: string;
  jsonData: {
    
    authMethod?: ServiceNowAuthMethod;
    
    enableSecureSocksProxy?: boolean;
    
    oauthClientID?: string;
    
    oauthEnabled?: boolean;
    
    useSysTables?: boolean;
    
    queryTimeoutSeconds?: number;
  };
  secureJsonData: {
    
    basicAuthPassword?: string;
    
    oauthClientSecret?: string;
  };
};

export type ServiceNowAuthMethod = "basicAuth" | "serviceNowOAuth";
