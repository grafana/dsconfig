export type stackdriverConfig = {
  jsonData: {
    
    authenticationType?: string;
    
    defaultProject?: string;
    
    gceDefaultProject?: string;
    
    clientEmail?: string;
    
    tokenUri?: string;
    
    enableSecureSocksProxy?: boolean;
    
    universeDomain?: string;
  };
  secureJsonData: {
    
    privateKey?: string;
  };
};
