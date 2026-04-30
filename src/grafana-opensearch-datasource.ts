export type grafanaOpensearchDatasourceConfig = {
  jsonData: {
    /** REQUIRED. Index name or pattern. */
    database: string;
    /** REQUIRED. Timestamp field name. */
    timeField: string;
    /** REQUIRED. OpenSearch version string. */
    version: string;
    /** REQUIRED. OpenSearch flavor. */
    flavor: OpenSearchFlavor;
    /** OPTIONAL. Index pattern interval. */
    interval?: string;
    /** OPTIONAL. Minimum time interval for auto-grouping. */
    timeInterval: string;
    /** OPTIONAL. Max concurrent shard requests. */
    maxConcurrentShardRequests?: number;
    /** OPTIONAL. Field name for log messages. */
    logMessageField?: string;
    /** OPTIONAL. Field name for log levels. */
    logLevelField?: string;
    /** OPTIONAL. Enable PPL query support. */
    pplEnabled?: boolean;
    /** OPTIONAL. Enable AWS SigV4 authentication. */
    sigV4Auth?: boolean;
    /** OPTIONAL. Enable OpenSearch Serverless mode. */
    serverless?: boolean;
    /** OPTIONAL. Enable Grafana secure SOCKS proxy. */
    enableSecureSocksProxy?: boolean;
  };
  secureJsonData: {};
};

export type OpenSearchFlavor = "opensearch" | "elasticsearch";
