export type oracleConfig = {
  url: string;
  jsonData: {
    
    database?: string;
    
    timezone_name?: string;
    
    use_dst?: boolean;
    
    user?: string;
    
    useTNSNamesBasedConnection?: boolean;
    
    tnsNamesEntry?: string;
    
    useKerberosAuthentication?: boolean;
    
    enableSecureSocksProxy?: boolean;
    
    connectionPoolSize?: number;
    
    dataProxyTimeout?: number;
    
    prefetchRowsCount?: number;
    
    rowLimit?: number;
  };
  secureJsonData: {
    
    password: string;
  };
};
