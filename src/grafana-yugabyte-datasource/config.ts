export type grafanaYugabyteDatasourceConfig = {
  jsonData: {
    
    database?: string;
    
    maxOpenConns?: number;
    
    maxIdleConns?: number;
    
    connMaxLifetime?: number;
    
    timeInterval?: string;
    
    enableSecureSocksProxy?: boolean;
  };
  secureJsonData: {
    
    password?: string;
  };
};
