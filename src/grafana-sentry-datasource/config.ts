export type grafanaSentryDatasourceConfig = {
  jsonData: {
    
    url: string;
    
    orgSlug: string;
    
    enableSecureSocksProxy?: boolean;
    
    tlsSkipVerify?: boolean;
  };
  secureJsonData: {
    
    authToken: string;
  };
};
