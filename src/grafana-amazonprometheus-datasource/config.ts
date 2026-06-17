import type { AWSAuthType } from "../cloudwatch/config";

export type grafanaAmazonprometheusDatasourceConfig = {
  jsonData: {
    
    authType?: AWSAuthType;
    
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
