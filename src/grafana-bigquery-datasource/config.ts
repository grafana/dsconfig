export type grafanaBigqueryDatasourceConfig = {
  jsonData: {
    
    authenticationType?: string;
    
    defaultProject?: string;
    
    clientEmail?: string;
    
    tokenUri?: string;
    
    flatRateProject?: string;
    
    processingLocation?: string;
    
    queryPriority?: string;
    
    enableSecureSocksProxy?: boolean;
    
    MaxBytesBilled?: number;
    
    serviceEndpoint?: string;
    
    oauthPassThru?: boolean;
  };
  secureJsonData: {
    
    privateKey?: string;
  };
};
