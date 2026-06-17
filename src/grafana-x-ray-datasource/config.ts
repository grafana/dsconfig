import type { AWSAuthType } from "../cloudwatch/config";

export type grafanaXRayDatasourceConfig = {
  jsonData: {
    
    authType?: AWSAuthType;
    
    defaultRegion?: string;
    
    assumeRoleArn?: string;
    
    externalId?: string;
    
    profile?: string;
    
    endpoint?: string;
  };
  secureJsonData: {
    
    accessKey?: string;
    
    secretKey?: string;
    
    sessionToken?: string;
  };
};
