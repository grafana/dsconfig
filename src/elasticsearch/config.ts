export type elasticsearchConfig = {
  jsonData: {
    
    timeField: string;
    
    interval?: ElasticsearchInterval;
    
    timeInterval?: string;
    
    maxConcurrentShardRequests?: number;
    
    logMessageField?: string;
    
    logLevelField?: string;
    
    dataLinks?: ElasticsearchDataLinkConfig[];
    
    includeFrozen?: boolean;
    
    index?: string;
    
    sigV4Auth?: boolean;
    
    oauthPassThru?: boolean;
    
    defaultQueryMode?: string;
    
    apiKeyAuth?: boolean;
  };
  secureJsonData: {
    
    apiKey?: string;
  };
};

export type ElasticsearchInterval = "Hourly" | "Daily" | "Weekly" | "Monthly" | "Yearly";

export type ElasticsearchDataLinkConfig = {
  field: string;
  url: string;
  urlDisplayLabel?: string;
  datasourceUid?: string;
};
