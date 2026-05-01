export type influxdbConfig = {
  jsonData: {
    
    version?: InfluxDBVersion;
    
    timeInterval?: string;
    
    httpMode?: string;
    
    dbName?: string;
    
    oauthPassThru?: boolean;
    
    organization?: string;
    
    defaultBucket?: string;
    
    maxSeries?: number;
    
    metadata?: Array<Record<string, string>>;
    
    insecureGrpc?: boolean;
  };
  secureJsonData: {
    
    token?: string;
    
    password?: string;
  };
};

export type InfluxDBVersion = "InfluxQL" | "Flux" | "SQL";
