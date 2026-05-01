export type gitLabConfig = {
  
  url?: string;
  jsonData: {
    
    pageLimit?: number;
    
    enableSecureSocksProxy?: boolean;
  };
  secureJsonData: {
    
    accessToken: string;
  };
};
