export type pagerDutyConfig = {
  jsonData: {
    
    servers?: {
      url: string;
      variables?: Record<string, string | number>;
    };
    
    auth?: {
      
      id: string;
      [key: string]: unknown;
    };
    
    enableSecureSocksProxy?: boolean;
  };
  secureJsonData: {
    
    [key: string]: string;
  };
};
