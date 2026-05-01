export type grafanaFalconlogscaleDatasourceConfig = {
  jsonData: {
    
    baseUrl?: string;
    
    oauthPassThru?: boolean;
    
    authenticateWithToken: boolean;
    
    defaultRepository?: string;
    
    enableSecureSocksProxy?: boolean;
    
    incrementalQuerying?: boolean;
    
    incrementalQueryOverlapWindow?: string;
  };
  secureJsonData: {
    
    accessToken?: string;
  };
};
