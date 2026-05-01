export type yesoreyeramInfinityDatasourceConfig = {
  jsonData: {
    
    auth_method?: string;
    
    apiKeyKey?: string;
    
    apiKeyType?: string;
    
    tlsSkipVerify?: boolean;
    
    tlsAuth?: boolean;
    
    serverName?: string;
    
    tlsAuthWithCACert?: boolean;
    
    timeoutInSeconds?: number;
    
    oauthPassThru?: boolean;
    
    allowedHosts?: string[];
    
    customHealthCheckEnabled?: boolean;
    
    customHealthCheckUrl?: string;
    
    enableSecureSocksProxy?: boolean;
    
    pathEncodedUrlsEnabled?: boolean;
    
    keepCookies?: string[];
  };
  secureJsonData: {
    
    basicAuthPassword?: string;
    
    tlsCACert?: string;
    
    tlsClientCert?: string;
    
    tlsClientKey?: string;
    
    apiKeyValue?: string;
    
    bearerToken?: string;
    
    oauth2ClientSecret?: string;
  };
};
