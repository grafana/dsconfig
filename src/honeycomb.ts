export type honeycombConfig = {
  jsonData: {
    /** REQUIRED. Honeycomb API hostname. Defaults to "https://api.honeycomb.io". */
    hostname: string;
    /** REQUIRED. Honeycomb team slug. */
    team: string;
    /** OPTIONAL. Honeycomb environment name. */
    environment?: string;
    /** OPTIONAL. Query retention limit in days. Defaults to 7. */
    retentionLimit?: number;
  };
  secureJsonData: {
    /** REQUIRED. Honeycomb API key. */
    apiKey: string;
  };
};
