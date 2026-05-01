export type azureDevOpsConfig = {
  jsonData: {
    
    url: string;
    
    authType: "patToken";
    
    projectsLimit?: number;
    
    enableSecureSocksProxy?: boolean;
    
    username?: string;
  };
  secureJsonData: {
    
    patToken: string;
  };
};
