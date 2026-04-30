export type yesoreyeramInfinityDatasourceConfig = {
  jsonData: {
    /** OPTIONAL. Authentication method. */
    auth_method?: string;
    /** OPTIONAL. API key header name. */
    apiKeyKey?: string;
    /** OPTIONAL. API key location (header or query). */
    apiKeyType?: string;
    /** OPTIONAL. Skip TLS certificate verification. */
    tlsSkipVerify?: boolean;
    /** OPTIONAL. Enable TLS client auth. */
    tlsAuth?: boolean;
    /** OPTIONAL. TLS server name. */
    serverName?: string;
    /** OPTIONAL. Enable TLS CA cert verification. */
    tlsAuthWithCACert?: boolean;
    /** OPTIONAL. Request timeout in seconds. */
    timeoutInSeconds?: number;
    /** OPTIONAL. Enable OAuth passthrough. */
    oauthPassThru?: boolean;
    /** OPTIONAL. Allowed host patterns. */
    allowedHosts?: string[];
    /** OPTIONAL. Enable custom health check. */
    customHealthCheckEnabled?: boolean;
    /** OPTIONAL. Custom health check URL. */
    customHealthCheckUrl?: string;
    /** OPTIONAL. Enable Grafana secure SOCKS proxy. */
    enableSecureSocksProxy?: boolean;
    /** OPTIONAL. Enable path-encoded URLs. */
    pathEncodedUrlsEnabled?: boolean;
    /** OPTIONAL. Cookie names to forward. */
    keepCookies?: string[];
  };
  secureJsonData: {
    /** OPTIONAL. Basic auth password. */
    basicAuthPassword?: string;
    /** OPTIONAL. TLS CA certificate. */
    tlsCACert?: string;
    /** OPTIONAL. TLS client certificate. */
    tlsClientCert?: string;
    /** OPTIONAL. TLS client key. */
    tlsClientKey?: string;
    /** OPTIONAL. API key value. */
    apiKeyValue?: string;
    /** OPTIONAL. Bearer token. */
    bearerToken?: string;
    /** OPTIONAL. OAuth2 client secret. */
    oauth2ClientSecret?: string;
  };
};
