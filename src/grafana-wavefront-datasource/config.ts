export type wavefrontConfig = {
  jsonData: {
    
    url: string;
    
    requestTimeout?: number;
    
    enableSecureSocksProxy?: boolean;
  };
  secureJsonData: {
    
    token: string;
  };
};
