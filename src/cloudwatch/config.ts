export type cloudWatchConfig = {
  jsonData: {
    
    authType?: CloudWatchAuthType;

    
    defaultRegion?: string;

    
    profile?: string;

    
    assumeRoleArn?: string;

    
    externalId?: string;

    
    endpoint?: string;

    
    customMetricsNamespaces?: string;

    
    logsTimeout?: string;

    
    logGroups?: CloudWatchLogGroup[];

    
    defaultLogGroups?: string[];

    
    tracingDatasourceUid?: string;

    
    enableSecureSocksProxy?: boolean;

    
    timeField?: string;

    
    database?: string;

    
    proxyType?: CloudWatchProxyType;

    
    proxyUrl?: string;

    
    proxyUsername?: string;
  };

  secureJsonData: {
    
    accessKey?: string;

    
    secretKey?: string;

    
    sessionToken?: string;

    
    proxyPassword?: string;
  };
};

export type CloudWatchAuthType =
  | "default"
  | "credentials"
  | "keys"
  | "ec2_iam_role"
  | "grafana_assume_role"
  | "arn";

export type CloudWatchProxyType = "none" | "env" | "url";

export type CloudWatchLogGroup = {
  
  arn: string;
  
  name: string;
};
