export type dynatraceConfig = {
  jsonData: {
    /** REQUIRED. Dynatrace deployment type. */
    apiType: DynatraceApiType;
    /** CONDITIONALLY REQUIRED: for "saas" or "managed". Dynatrace environment ID. */
    environmentId?: string;
    /** CONDITIONALLY REQUIRED: for "managed". Dynatrace Managed domain. */
    domain?: string;
    /** OPTIONAL. Skip TLS certificate verification. */
    tlsSkipVerify?: boolean;
    /** OPTIONAL. Enable TLS authentication with CA certificate. */
    tlsAuthWithCACert?: boolean;
    /** OPTIONAL. Enable Grafana secure SOCKS proxy. */
    enableSecureSocksProxy?: boolean;
    /** OPTIONAL. HTTP client timeout in seconds. Defaults to 30. */
    httpClientTimeout?: number;
  };
  secureJsonData: {
    /** REQUIRED. Dynatrace API token. */
    apiToken: string;
    /** OPTIONAL. Dynatrace platform token (for Grail/DQL). */
    platformToken?: string;
    /** OPTIONAL. TLS CA certificate content. */
    tlsCACert?: string;
  };
};

export type DynatraceApiType = "saas" | "managed" | "url";
