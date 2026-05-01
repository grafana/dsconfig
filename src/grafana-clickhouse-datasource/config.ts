export type grafanaClickhouseDatasourceConfig = {
  jsonData: {
    
    host: string;
    
    port: number;
    
    protocol: ClickHouseProtocol;
    
    username: string;
    
    version?: string;
    
    secure?: boolean;
    
    path?: string;
    
    tlsSkipVerify?: boolean;
    
    tlsAuth?: boolean;
    
    tlsAuthWithCACert?: boolean;
    
    defaultDatabase?: string;
    
    defaultTable?: string;
    
    connMaxLifetime?: string;
    
    dialTimeout?: string;
    
    maxIdleConns?: string;
    
    maxOpenConns?: string;
    
    queryTimeout?: string;
    
    validateSql?: boolean;
    
    forwardGrafanaHeaders?: boolean;
    
    enableSecureSocksProxy?: boolean;
    
    enableRowLimit?: boolean;
  };
  secureJsonData: {
    
    password?: string;
    
    tlsCACert?: string;
    
    tlsClientCert?: string;
    
    tlsClientKey?: string;
  };
};

export type ClickHouseProtocol = "native" | "http";
