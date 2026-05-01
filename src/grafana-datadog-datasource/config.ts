export type datadogConfig = {
  jsonData: {
    
    pluginMode?: DatadogPluginMode;
    
    url: string;
    
    logApiRateLimits?: boolean;
    
    disableDataLinks?: boolean;
    
    rateLimitEnabled?: boolean;
    
    rateLimitMetrics?: number;
    
    enableSecureSocksProxy?: boolean;
    
    size?: number;
  };
  secureJsonData: {
    
    apiKey: string;
    
    appKey: string;
  };
};

export type DatadogPluginMode = "default" | "hosted-metrics";
