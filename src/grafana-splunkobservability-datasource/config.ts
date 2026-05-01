export type splunkObservabilityConfig = {
  jsonData: {
    
    realmName: string;
    
    url_metrics_metadata?: string;
    
    url_signalflow?: string;
    
    enableSecureSocksProxy?: boolean;
  };
  secureJsonData: {
    
    accessToken: string;
  };
};
