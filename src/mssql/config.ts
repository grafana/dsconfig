export type mssqlConfig = {
  jsonData: {
    
    authenticationType?: MSSQLAuthenticationType;
    
    encrypt?: MSSQLEncryptOption;
    
    sslRootCertFile?: string;
    
    serverName?: string;
    
    connectionTimeout?: number;
    
    database?: string;
    
    maxOpenConns?: number;
    
    maxIdleConns?: number;
    
    connMaxLifetime?: number;
    
    keytabFilePath?: string;
    
    credentialCache?: string;
    
    credentialCacheLookupFile?: string;
    
    configFilePath?: string;
    
    UDPConnectionLimit?: number;
    
    enableDNSLookupKDC?: string;
    
    timeInterval?: string;
  };
  secureJsonData: {
    
    password?: string;
  };
};

export type MSSQLAuthenticationType =
  | "SQL Server Authentication"
  | "Windows Authentication"
  | "Azure AD Authentication"
  | "Windows AD: Username + password"
  | "Windows AD: Keytab"
  | "Windows AD: Credential cache"
  | "Windows AD: Credential cache file";

export type MSSQLEncryptOption = "disable" | "false" | "true";
