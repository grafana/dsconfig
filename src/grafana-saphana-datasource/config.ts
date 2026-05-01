export type sapHanaConfig = {
  jsonData: {
    
    server: string;
    
    username: string;
    
    port?: number;
    
    instance?: number;
    
    databaseName?: string;
    
    defaultSchema?: string;
    
    tlsSkipVerify?: boolean;
    
    tlsAuth?: boolean;
    
    tlsAuthWithCACert?: boolean;
    
    timeout?: string;
    
    enableSecureSocksProxy?: boolean;
  };
  secureJsonData: {
    
    password: string;
    
    tlsCACert?: string;
    
    tlsClientCert?: string;
    
    tlsClientKey?: string;
  };
};
