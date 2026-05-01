export type dynatraceConfig = {
  jsonData: {
    
    apiType: DynatraceApiType;
    
    environmentId?: string;
    
    domain?: string;
    
    tlsSkipVerify?: boolean;
    
    tlsAuthWithCACert?: boolean;
    
    enableSecureSocksProxy?: boolean;
    
    httpClientTimeout?: number;
  };
  secureJsonData: {
    
    apiToken: string;
    
    platformToken?: string;
    
    tlsCACert?: string;
  };
};

export type DynatraceApiType = "saas" | "managed" | "url";
