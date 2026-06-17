import type { AWSAuthType } from "../cloudwatch/config";

export type grafanaTimestreamDatasourceConfig = {
  jsonData: {
    
    authType?: AWSAuthType;
    
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
