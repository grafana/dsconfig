export type wavefrontConfig = {
  jsonData: {
    /** REQUIRED. Wavefront cluster URL (e.g., "https://your-cluster.wavefront.com"). */
    url: string;
    /** OPTIONAL. Request timeout in seconds. Defaults to 30. */
    requestTimeout?: number;
    /** OPTIONAL. Enable Grafana secure SOCKS proxy. */
    enableSecureSocksProxy?: boolean;
  };
  secureJsonData: {
    /** REQUIRED. Wavefront API token. */
    token: string;
  };
};
