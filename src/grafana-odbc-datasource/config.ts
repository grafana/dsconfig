export type odbcConfig = {
  jsonData: {
    
    driver: string;
    
    timeout?: string;
    
    settings?: OdbcSetting[];
  };
  secureJsonData: {
    
    [key: string]: string;
  };
};

export type OdbcSetting = {
  
  name: string;
  
  value?: string;
  
  secure: boolean;
};
