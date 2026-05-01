export type grafanaAmazonprometheusDatasourceConfig = {
  jsonData: {
    
    authType?: string;
    
    defaultRegion?: string;
    
    assumeRoleArn?: string;
    
    externalId?: string;
    
    profile?: string;
    
    endpoint?: string;
    
    sigv4Service?: string;
    
    timeInterval?: string;
    
    httpMethod?: string;
    
    customQueryParameters?: string;
  };
  secureJsonData: {
    
    accessKey?: string;
    
    secretKey?: string;
    
    sessionToken?: string;
  };
};
