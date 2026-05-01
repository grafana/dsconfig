export type cockroachDbConfig = {
  url: string;
  user?: string;
  jsonData: {
    
    database: string;
    
    sslmode?: CockroachDbTlsMode;
    
    tlsConfigurationMethod?: CockroachDbTlsMethod;
    
    sslRootCertFile?: string;
    
    sslCertFile?: string;
    
    sslKeyFile?: string;
    
    authType?: string;
    
    maxOpenConns?: number;
    
    maxIdleConns?: number;
    
    maxIdleConnsAuto?: boolean;
    
    connMaxLifetime?: number;
    
    queryTimeout?: number;
    
    configFilePath?: string;
    
    credentialCache?: string;
    
    kerberosServerName?: string;
    
    postgresVersion?: number;
    
    timescaledb?: boolean;
    
    enableSecureSocksProxy?: boolean;
  };
  secureJsonData: {
    
    password: string;
  };
};

export type CockroachDbTlsMode = "disable" | "require" | "verify-ca" | "verify-full";
export type CockroachDbTlsMethod = "file-path" | "file-content";
