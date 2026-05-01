export type honeycombConfig = {
  jsonData: {
    
    hostname: string;
    
    team: string;
    
    environment?: string;
    
    retentionLimit?: number;
  };
  secureJsonData: {
    
    apiKey: string;
  };
};
