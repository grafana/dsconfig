export type grafanaStravaDatasourceConfig = {
  jsonData: {
    
    clientID: string;
    
    cacheTTL?: string;
    
    oauthPassThru?: boolean;
  };
  secureJsonData: {
    
    clientSecret: string;
  };
};
