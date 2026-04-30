export type elasticsearchConfig = {
  jsonData: {
    /** REQUIRED. Timestamp field name. */
    timeField: string;
    /** OPTIONAL. Index pattern interval (Hourly, Daily, etc.). */
    interval?: ElasticsearchInterval;
    /** OPTIONAL. Minimum time interval for auto-grouping. */
    timeInterval?: string;
    /** OPTIONAL. Max concurrent shard requests. */
    maxConcurrentShardRequests?: number;
    /** OPTIONAL. Field name for log messages. */
    logMessageField?: string;
    /** OPTIONAL. Field name for log levels. */
    logLevelField?: string;
    /** OPTIONAL. Data link configurations. */
    dataLinks?: ElasticsearchDataLinkConfig[];
    /** OPTIONAL. Include frozen indices in searches. */
    includeFrozen?: boolean;
    /** OPTIONAL. Index name or pattern. */
    index?: string;
    /** OPTIONAL. Enable AWS SigV4 authentication. */
    sigV4Auth?: boolean;
    /** OPTIONAL. Enable OAuth passthrough. */
    oauthPassThru?: boolean;
    /** OPTIONAL. Default query mode. */
    defaultQueryMode?: string;
    /** OPTIONAL. Enable API key authentication. */
    apiKeyAuth?: boolean;
  };
  secureJsonData: {
    /** CONDITIONALLY REQUIRED: for apiKeyAuth. Elasticsearch API key. */
    apiKey?: string;
  };
};

export type ElasticsearchInterval = "Hourly" | "Daily" | "Weekly" | "Monthly" | "Yearly";

export type ElasticsearchDataLinkConfig = {
  field: string;
  url: string;
  urlDisplayLabel?: string;
  datasourceUid?: string;
};
