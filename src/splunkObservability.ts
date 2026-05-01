export type splunkObservabilityConfig = {
  jsonData: {
    /** REQUIRED. Splunk Observability Cloud realm name (e.g., "us0", "eu0"). */
    realmName: string;
    /** OPTIONAL. Override URL for Metrics Metadata API. Defaults to "https://api.{REALM}.signalfx.com". */
    url_metrics_metadata?: string;
    /** OPTIONAL. Override URL for SignalFlow API. Defaults to "https://stream.{REALM}.signalfx.com". */
    url_signalflow?: string;
    /** OPTIONAL. Enable Grafana secure SOCKS proxy. */
    enableSecureSocksProxy?: boolean;
  };
  secureJsonData: {
    /** REQUIRED. Splunk Observability Cloud access token. */
    accessToken: string;
  };
};
