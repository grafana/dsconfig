export type graphiteConfig = {
  jsonData: {
    /** REQUIRED. Graphite version string. */
    graphiteVersion: string;
    /** REQUIRED. Graphite backend type. */
    graphiteType: GraphiteType;
    /** OPTIONAL. Enable rollup indicator. */
    rollupIndicatorEnabled?: boolean;
    /** OPTIONAL. Import configuration for query migration. */
    importConfiguration?: Record<string, unknown>;
  };
  secureJsonData: {};
};

export type GraphiteType = "default" | "metrictank";
