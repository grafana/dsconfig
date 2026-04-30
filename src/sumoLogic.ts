export type sumoLogicConfig = {
  jsonData: {
    /** OPTIONAL. Authentication method. Defaults to "accessKey". */
    authMethod?: SumoLogicAuthMethod;
    /** REQUIRED. Sumo Logic API URL (e.g., "https://api.sumologic.com/api"). */
    apiUrl: string;
    /** REQUIRED. Sumo Logic Access ID. */
    accessId: string;
    /** OPTIONAL. Request timeout in seconds. */
    timeout?: number;
    /** OPTIONAL. Polling interval in milliseconds. Defaults to 1000. */
    interval?: number;
  };
  secureJsonData: {
    /** REQUIRED. Sumo Logic Access Key. */
    accessKey: string;
  };
};

export type SumoLogicAuthMethod = "accessKey";
