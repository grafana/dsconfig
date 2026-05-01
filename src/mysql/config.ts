export type mysqlConfig = {
  jsonData: {
    
    database?: string;
    
    maxOpenConns?: number;
    
    maxIdleConns?: number;
    
    connMaxLifetime?: number;
    
    allowCleartextPasswords?: boolean;
    
    timeInterval?: string;
  };
  secureJsonData: {
    
    password?: string;
  };
};
