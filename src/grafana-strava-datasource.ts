export type grafanaStravaDatasourceConfig = {
  jsonData: {
    /** REQUIRED. Strava OAuth2 client ID. */
    clientID: string;
    /** OPTIONAL. Cache TTL duration string (e.g., "1h"). */
    cacheTTL?: string;
    /** OPTIONAL. Enable OAuth passthrough. */
    oauthPassThru?: boolean;
  };
  secureJsonData: {
    /** REQUIRED. Strava OAuth2 client secret. */
    clientSecret: string;
  };
};
