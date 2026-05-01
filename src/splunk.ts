export type splunkConfig = {
  url: string;
  jsonData: {
    /** OPTIONAL. Splunk REST API URL (if different from url). */
    apiURL?: string;
    /** OPTIONAL. Authentication type ("BasicAuth" or token-based). */
    authType?: string;
    /** OPTIONAL. Splunk username for basic auth. */
    username?: string;
    /** OPTIONAL. Enable OAuth passthrough. */
    oauthPassThru?: boolean;
    /** OPTIONAL. Enable search result polling (async mode). */
    pollSearchResult?: boolean;
    /** OPTIONAL. Enable stream mode for real-time searches. */
    streamMode?: boolean;
    /** OPTIONAL. Enable preview mode (partial results). */
    previewMode?: boolean;
    /** OPTIONAL. Auto-cancel search after duration. */
    autoCancel?: string;
    /** OPTIONAL. Status buckets for search progress. */
    statusBuckets?: string;
    /** OPTIONAL. Default earliest time for searches. */
    defaultEarliestTime?: string;
    /** OPTIONAL. Enable filtering of internal fields. */
    internalFieldsFiltration?: boolean;
    /** OPTIONAL. Regex pattern for internal fields. Defaults to "^_.+". */
    internalFieldPattern?: string;
    /** OPTIONAL. Minimum poll interval in ms. */
    minPollInterval?: number;
    /** OPTIONAL. Maximum poll interval in ms. */
    maxPollInterval?: number;
    /** OPTIONAL. Timestamp field name. Defaults to "_time". */
    tsField?: string;
    /** OPTIONAL. Field search type. */
    fieldSearchType?: string;
    /** OPTIONAL. Variable search level. */
    variableSearchLevel?: string;
    /** OPTIONAL. Maximum result count per search. */
    maxResultCount?: number;
    /** OPTIONAL. Request timeout in seconds. Defaults to 30. */
    timeoutInSeconds?: number;
    /** OPTIONAL. Data link configurations for field→URL mappings. */
    dataLinks?: SplunkDataLinkConfig[];
    /** OPTIONAL. Cookie names to forward. */
    keepCookies?: string[];
    /** OPTIONAL. Enable Grafana secure SOCKS proxy. */
    enableSecureSocksProxy?: boolean;
  };
  secureJsonData: {
    /** CONDITIONALLY REQUIRED: for basic auth. Splunk password. */
    basicAuthPassword?: string;
    /** CONDITIONALLY REQUIRED: for token auth. Splunk auth token. */
    authToken?: string;
  };
};

export type SplunkDataLinkConfig = {
  field: string;
  label: string;
  matcherRegex: string;
  url: string;
  datasourceUid?: string;
};
