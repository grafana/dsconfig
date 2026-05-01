export type influxdbConfig = {
  jsonData: {
    /** OPTIONAL. InfluxDB query language version. Defaults to InfluxQL. */
    version?: InfluxDBVersion;
    /** OPTIONAL. Minimum time interval for auto-grouping. */
    timeInterval?: string;
    /** OPTIONAL. HTTP method (GET or POST). */
    httpMode?: string;
    /** OPTIONAL. Database name (InfluxQL mode). */
    dbName?: string;
    /** OPTIONAL. Enable OAuth passthrough. */
    oauthPassThru?: boolean;
    /** OPTIONAL. Organization name (Flux mode). */
    organization?: string;
    /** OPTIONAL. Default bucket (Flux mode). */
    defaultBucket?: string;
    /** OPTIONAL. Maximum series returned (Flux mode). */
    maxSeries?: number;
    /** OPTIONAL. Metadata key-value pairs (SQL mode). */
    metadata?: Array<Record<string, string>>;
    /** OPTIONAL. Use insecure gRPC connection (SQL mode). */
    insecureGrpc?: boolean;
  };
  secureJsonData: {
    /** CONDITIONALLY REQUIRED: for Flux mode. API token. */
    token?: string;
    /** OPTIONAL. Database password (InfluxQL 1.x). */
    password?: string;
  };
};

export type InfluxDBVersion = "InfluxQL" | "Flux" | "SQL";
