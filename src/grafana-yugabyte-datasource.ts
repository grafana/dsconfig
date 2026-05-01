export type grafanaYugabyteDatasourceConfig = {
  jsonData: {
    /** OPTIONAL. Database name. */
    database?: string;
    /** OPTIONAL. Maximum open connections. */
    maxOpenConns?: number;
    /** OPTIONAL. Maximum idle connections. */
    maxIdleConns?: number;
    /** OPTIONAL. Connection max lifetime in seconds. */
    connMaxLifetime?: number;
    /** OPTIONAL. Minimum time interval for auto-grouping. */
    timeInterval?: string;
    /** OPTIONAL. Enable Grafana secure SOCKS proxy. */
    enableSecureSocksProxy?: boolean;
  };
  secureJsonData: {
    /** REQUIRED. Database password. */
    password?: string;
  };
};
