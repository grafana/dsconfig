export type newRelicConfig = {
  jsonData: {
    
    region?: NewRelicSupportedRegion;

    
    timeoutInSeconds?: number;

    
    restBaseURL?: string;

    
    infrastructureBaseURL?: string;

    
    nerdGraphBaseURL?: string;
  };

  secureJsonData: {
    
    accountId?: string;

    
    personalApiKey?: string;
  };
};

export type NewRelicSupportedRegion = "US" | "EU";
