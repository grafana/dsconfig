export type splunkConfig = {
  url: string;
  jsonData: {
    
    apiURL?: string;
    
    authType?: string;
    
    username?: string;
    
    oauthPassThru?: boolean;
    
    pollSearchResult?: boolean;
    
    streamMode?: boolean;
    
    previewMode?: boolean;
    
    autoCancel?: string;
    
    statusBuckets?: string;
    
    defaultEarliestTime?: string;
    
    internalFieldsFiltration?: boolean;
    
    internalFieldPattern?: string;
    
    minPollInterval?: number;
    
    maxPollInterval?: number;
    
    tsField?: string;
    
    fieldSearchType?: string;
    
    variableSearchLevel?: string;
    
    maxResultCount?: number;
    
    timeoutInSeconds?: number;
    
    dataLinks?: SplunkDataLinkConfig[];
    
    keepCookies?: string[];
    
    enableSecureSocksProxy?: boolean;
  };
  secureJsonData: {
    
    basicAuthPassword?: string;
    
    authToken?: string;
  };
};

export type SplunkDataLinkConfig = {
  field: string;
  label: string;
  matcherRegex: string;
  url: string;
  datasourceUid?: string;
};
