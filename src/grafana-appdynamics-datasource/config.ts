export type appDynamicsConfig = {
  url: string;
  basicAuth?: boolean;
  basicAuthUser?: string;
  jsonData: {
    
    clientName?: string;
    
    clientDomain?: string;
    
    analyticsURL?: string;
    
    globalAccountName?: string;
    
    enableSecureSocksProxy?: boolean;
  };
  secureJsonData: {
    
    basicAuthPassword?: string;
    
    clientSecret?: string;
    
    analyticsAPIKey?: string;
  };
};
