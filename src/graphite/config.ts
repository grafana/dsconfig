export type graphiteConfig = {
  jsonData: {
    
    graphiteVersion: string;
    
    graphiteType: GraphiteType;
    
    rollupIndicatorEnabled?: boolean;
    
    importConfiguration?: Record<string, unknown>;
  };
  secureJsonData: {};
};

export type GraphiteType = "default" | "metrictank";
