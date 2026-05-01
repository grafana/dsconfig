export type sumoLogicConfig = {
  jsonData: {
    
    authMethod?: SumoLogicAuthMethod;
    
    apiUrl: string;
    
    accessId: string;
    
    timeout?: number;
    
    interval?: number;
  };
  secureJsonData: {
    
    accessKey: string;
  };
};

export type SumoLogicAuthMethod = "accessKey";
