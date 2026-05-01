export type grafanaTimestreamDatasourceConfig = {
  jsonData: {
    
    authType?: string;
    
    defaultRegion?: string;
    
    assumeRoleArn?: string;
    
    externalId?: string;
    
    profile?: string;
    
    endpoint?: string;
    
    defaultDatabase?: string;
    
    defaultTable?: string;
    
    defaultMeasure?: string;
  };
  secureJsonData: {
    
    accessKey?: string;
    
    secretKey?: string;
    
    sessionToken?: string;
  };
};
