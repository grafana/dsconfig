export type appDynamicsConfig = {
  url: string;
  basicAuth?: boolean;
  basicAuthUser?: string;
  jsonData: {
    /** OPTIONAL. AppDynamics controller client name for OAuth authentication. */
    clientName?: string;
    /** OPTIONAL. AppDynamics controller client domain. */
    clientDomain?: string;
    /** OPTIONAL. AppDynamics Analytics API URL. */
    analyticsURL?: string;
    /** OPTIONAL. AppDynamics global account name for Analytics queries. */
    globalAccountName?: string;
    /** OPTIONAL. Enable Grafana secure SOCKS proxy. */
    enableSecureSocksProxy?: boolean;
  };
  secureJsonData: {
    /** CONDITIONALLY REQUIRED: for basic auth. Controller password. */
    basicAuthPassword?: string;
    /** CONDITIONALLY REQUIRED: for OAuth. Client secret for OAuth authentication. */
    clientSecret?: string;
    /** OPTIONAL. Analytics API key for Analytics queries. */
    analyticsAPIKey?: string;
  };
};
