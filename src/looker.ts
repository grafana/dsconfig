export type lookerConfig = {
  jsonData: {
    /** REQUIRED. Looker API base URL. */
    base_url: string;
    /** REQUIRED. Authentication type. Currently only "client_secret". */
    auth_type: LookerAuthType;
    /** REQUIRED. Looker API client ID. */
    client_id: string;
  };
  secureJsonData: {
    /** REQUIRED. Looker API client secret. */
    client_secret: string;
  };
};

export type LookerAuthType = "client_secret";
