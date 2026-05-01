export type grafanaOpensearchDatasourceConfig = {
  jsonData: {
    
    database: string;
    
    timeField: string;
    
    version: string;
    
    flavor: OpenSearchFlavor;
    
    interval?: string;
    
    timeInterval: string;
    
    maxConcurrentShardRequests?: number;
    
    logMessageField?: string;
    
    logLevelField?: string;
    
    pplEnabled?: boolean;
    
    sigV4Auth?: boolean;
    
    serverless?: boolean;
    
    enableSecureSocksProxy?: boolean;
  };
  secureJsonData: {};
};

export type OpenSearchFlavor = "opensearch" | "elasticsearch";
