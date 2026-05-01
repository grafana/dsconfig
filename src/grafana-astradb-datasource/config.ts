export type grafanaAstradbDatasourceConfig = {
  jsonData: {
    
    uri: string;
    
    authKind: number;
    
    user: string;
    
    grpcEndpoint: string;
    
    authEndpoint: string;
    
    secure?: boolean;
    
    database?: string;
  };
  secureJsonData: {
    
    token?: string;
    
    password?: string;
  };
};
