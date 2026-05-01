export type serviceNowConfig = {
  url: string;
  basicAuth?: boolean;
  basicAuthUser?: string;
  jsonData: {
    /** OPTIONAL. Authentication method. Defaults to "basicAuth". */
    authMethod?: ServiceNowAuthMethod;
    /** OPTIONAL. Enable Grafana secure SOCKS proxy. */
    enableSecureSocksProxy?: boolean;
    /** CONDITIONALLY REQUIRED: for OAuth. OAuth client ID. */
    oauthClientID?: string;
    /** DEPRECATED: use authMethod "serviceNowOAuth". Legacy OAuth toggle. */
    oauthEnabled?: boolean;
    /** OPTIONAL. Use sys_ prefixed tables for metadata queries. */
    useSysTables?: boolean;
    /** OPTIONAL. Query timeout in seconds. Defaults to 30. */
    queryTimeoutSeconds?: number;
  };
  secureJsonData: {
    /** CONDITIONALLY REQUIRED: for basicAuth. Basic auth password. */
    basicAuthPassword?: string;
    /** CONDITIONALLY REQUIRED: for OAuth. OAuth client secret. */
    oauthClientSecret?: string;
  };
};

export type ServiceNowAuthMethod = "basicAuth" | "serviceNowOAuth";
