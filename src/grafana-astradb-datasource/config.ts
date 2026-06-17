export type AstraDbAuthKind = 0 | 1; // 0: token, 1: credentials

export type grafanaAstradbDatasourceConfig = {
  jsonData: {
    
    uri: string;
    
    authKind: AstraDbAuthKind;
    
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
