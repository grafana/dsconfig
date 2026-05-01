export type grafanaMqttDatasourceConfig = {
  jsonData: {
    
    uri: string;
    
    username?: string;
    
    clientID?: string;
    
    tlsAuth: boolean;
    
    tlsAuthWithCACert: boolean;
    
    tlsSkipVerify: boolean;
  };
  secureJsonData: {
    
    password?: string;
    
    tlsCACert?: string;
    
    tlsClientKey?: string;
    
    tlsClientCert?: string;
  };
};
