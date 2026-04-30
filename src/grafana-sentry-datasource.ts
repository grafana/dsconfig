export type grafanaSentryDatasourceConfig = {
  jsonData: {
    /** REQUIRED. Sentry API URL. */
    url: string;
    /** REQUIRED. Sentry organization slug. */
    orgSlug: string;
    /** OPTIONAL. Enable Grafana secure SOCKS proxy. */
    enableSecureSocksProxy?: boolean;
    /** OPTIONAL. Skip TLS certificate verification. */
    tlsSkipVerify?: boolean;
  };
  secureJsonData: {
    /** REQUIRED. Sentry auth token. */
    authToken: string;
  };
};
