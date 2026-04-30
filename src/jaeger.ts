export type jaegerConfig = {
  jsonData: {
    /** OPTIONAL. Node graph visualization options. */
    nodeGraph?: { enabled?: boolean };
    /** OPTIONAL. Whether to include time params when querying by trace ID. */
    traceIdTimeParams?: { enabled?: boolean };
  };
  secureJsonData: {};
};
