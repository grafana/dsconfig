import type { AWSAuthType } from "../cloudwatch/config";

export type grafanaRedshiftDatasourceConfig = {
  jsonData: {
    
    authType?: AWSAuthType;
    
    defaultRegion?: string;
    
    assumeRoleArn?: string;
    
    externalId?: string;
    
    profile?: string;
    
    endpoint?: string;
    
    withEvent?: boolean;
    
    useManagedSecret?: boolean;
    
    useServerless?: boolean;
    
    workgroupName?: string;
    
    clusterIdentifier?: string;
    
    database?: string;
    
    dbUser?: string;
    
    managedSecret?: { name: string; arn: string };
    
    enableSecureSocksProxy?: boolean;
  };
  secureJsonData: {
    
    accessKey?: string;
    
    secretKey?: string;
    
    sessionToken?: string;
  };
};
