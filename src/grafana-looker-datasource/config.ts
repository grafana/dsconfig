export type lookerConfig = {
  jsonData: {
    
    base_url: string;
    
    auth_type: LookerAuthType;
    
    client_id: string;
  };
  secureJsonData: {
    
    client_secret: string;
  };
};

export type LookerAuthType = "client_secret";
