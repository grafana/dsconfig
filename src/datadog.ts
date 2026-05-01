export type datadogConfig = {
  jsonData: {
    /** OPTIONAL. Plugin operating mode. */
    pluginMode?: DatadogPluginMode;
    /** REQUIRED. Datadog API site URL (e.g., "https://api.datadoghq.com"). */
    url: string;
    /** OPTIONAL. Log API rate limit info in backend logs. */
    logApiRateLimits?: boolean;
    /** OPTIONAL. Disable data links to Datadog UI. */
    disableDataLinks?: boolean;
    /** OPTIONAL. Enable client-side rate limiting. */
    rateLimitEnabled?: boolean;
    /** OPTIONAL. Rate limit for metrics API (requests/sec). */
    rateLimitMetrics?: number;
    /** OPTIONAL. Enable Grafana secure SOCKS proxy. */
    enableSecureSocksProxy?: boolean;
    /** OPTIONAL. API response page size. */
    size?: number;
  };
  secureJsonData: {
    /** REQUIRED. Datadog API key. */
    apiKey: string;
    /** REQUIRED. Datadog Application key. */
    appKey: string;
  };
};

export type DatadogPluginMode = "default" | "hosted-metrics";
