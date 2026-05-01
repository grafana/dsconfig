export type mongoDbConfig = {
  jsonData: {
    
    connection: string;
    
    authType?: MongoDbAuthType;
    
    user?: string;
    
    serverName?: string;
    
    tlsAuth?: boolean;
    
    tlsAuthWithCACert?: boolean;
    
    tlsSkipVerify?: boolean;
    
    skipTLSValidation?: boolean;
    
    enableSecureSocksProxy?: boolean;
    
    responseRowsLimit?: string;
    
    kerberosUser?: string;
    
    ccacheLookupFile?: string;
    
    keyTabFilePath?: string;
    
    globalCcacheFilePath?: string;
    
    validate?: boolean;
  };
  secureJsonData: {
    
    password?: string;
    
    kerberosPassword?: string;
    
    basicAuthPassword?: string;
    
    tlsCertificateKeyFilePassword?: string;
  };
};

export type MongoDbAuthType = "NoAuth" | "BasicAuth" | "custom-Kerberos";
