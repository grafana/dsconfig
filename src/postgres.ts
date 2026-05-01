export type postgresConfig = {
  url: string;
  user: string;
  jsonData: {
    database: string;
    sslmode?: 'disable' | 'require' | 'verify-ca' | 'verify-full';
    tlsConfigurationMethod?: 'file-path' | 'file-content';
    sslRootCertFile?: string;
    sslCertFile?: string;
    sslKeyFile?: string;
    tlsSkipVerify?: boolean;
    postgresVersion?: number;
    timescaledb?: boolean;
    timeInterval?: string;
    maxOpenConns?: number;
    maxIdleConns?: number;
    maxIdleConnsAuto?: boolean;
    connMaxLifetime?: number;
    connectionTimeout?: number;
    timezone?: string;
    servername?: string;
    encrypt?: string;
    authenticationType?: string;
    allowCleartextPasswords?: boolean;
    enableSecureSocksProxy?: boolean;
    secureSocksProxyUsername?: boolean;
  };
  secureJsonData: {
    password?: string;
    tlsCACert?: string;
    tlsClientCert?: string;
    tlsClientKey?: string;
  };
};