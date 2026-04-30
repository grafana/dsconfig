export type grafanaClickhouseDatasourceConfig = {
  jsonData: {
    /** REQUIRED. ClickHouse server hostname. */
    host: string;
    /** REQUIRED. ClickHouse server port. */
    port: number;
    /** REQUIRED. Connection protocol. */
    protocol: ClickHouseProtocol;
    /** REQUIRED. ClickHouse username. */
    username: string;
    /** OPTIONAL. Plugin version. */
    version?: string;
    /** OPTIONAL. Enable TLS/SSL. */
    secure?: boolean;
    /** OPTIONAL. HTTP path prefix. */
    path?: string;
    /** OPTIONAL. Skip TLS certificate verification. */
    tlsSkipVerify?: boolean;
    /** OPTIONAL. Enable TLS client auth. */
    tlsAuth?: boolean;
    /** OPTIONAL. Enable TLS CA cert verification. */
    tlsAuthWithCACert?: boolean;
    /** OPTIONAL. Default database. */
    defaultDatabase?: string;
    /** OPTIONAL. Default table. */
    defaultTable?: string;
    /** OPTIONAL. Connection max lifetime. */
    connMaxLifetime?: string;
    /** OPTIONAL. Dial timeout. */
    dialTimeout?: string;
    /** OPTIONAL. Max idle connections. */
    maxIdleConns?: string;
    /** OPTIONAL. Max open connections. */
    maxOpenConns?: string;
    /** OPTIONAL. Query timeout. */
    queryTimeout?: string;
    /** OPTIONAL. Enable SQL validation. */
    validateSql?: boolean;
    /** OPTIONAL. Enable forwarding Grafana headers. */
    forwardGrafanaHeaders?: boolean;
    /** OPTIONAL. Enable Grafana secure SOCKS proxy. */
    enableSecureSocksProxy?: boolean;
    /** OPTIONAL. Enable row limit. */
    enableRowLimit?: boolean;
  };
  secureJsonData: {
    /** REQUIRED. ClickHouse password. */
    password?: string;
    /** OPTIONAL. TLS CA certificate. */
    tlsCACert?: string;
    /** OPTIONAL. TLS client certificate. */
    tlsClientCert?: string;
    /** OPTIONAL. TLS client key. */
    tlsClientKey?: string;
  };
};

export type ClickHouseProtocol = "native" | "http";
