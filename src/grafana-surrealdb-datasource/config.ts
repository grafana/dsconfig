export type grafanaSurrealdbDatasourceConfig = {
  jsonData: {
    
    endpoint?: string;
    
    database?: string;
    
    namespace?: string;
    
    scope?: string;
    
    username?: string;
  };
  secureJsonData: {
    
    password?: string;
  };
};
