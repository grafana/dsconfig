export type grafanaIotSitewiseDatasourceConfig = {
  jsonData: {
    
    authType?: string;
    
    defaultRegion?: string;
    
    assumeRoleArn?: string;
    
    externalId?: string;
    
    profile?: string;
    
    endpoint?: string;
    
    edgeAuthMode?: string;
    
    edgeAuthUser?: string;
  };
  secureJsonData: {
    
    accessKey?: string;
    
    secretKey?: string;
    
    sessionToken?: string;
    
    edgeAuthPass?: string;
    
    cert?: string;
  };
};
