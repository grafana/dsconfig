import type { AWSAuthType } from "../cloudwatch/config";

export type grafanaAthenaDatasourceConfig = {
  jsonData: {
    
    authType?: AWSAuthType;
    
    defaultRegion?: string;
    
    assumeRoleArn?: string;
    
    externalId?: string;
    
    profile?: string;
    
    endpoint?: string;
    
    catalog?: string;
    
    database?: string;
    
    workgroup?: string;
    
    outputLocation?: string;
  };
  secureJsonData: {
    
    accessKey?: string;
    
    secretKey?: string;
    
    sessionToken?: string;
  };
};
