export type pagerDutyConfig = {
  jsonData: {
    /** OPTIONAL. PagerDuty API server configuration. */
    servers?: {
      url: string;
      variables?: Record<string, string | number>;
    };
    /** OPTIONAL. Authentication configuration. */
    auth?: {
      /** Authentication scheme identifier (e.g., "api_key_v2"). */
      id: string;
      [key: string]: unknown;
    };
    /** OPTIONAL. Enable Grafana secure SOCKS proxy. */
    enableSecureSocksProxy?: boolean;
  };
  secureJsonData: {
    /** Dynamic keys: secure credential values stored by auth scheme key name. */
    [key: string]: string;
  };
};
