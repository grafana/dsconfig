export type marcusolssonCsvDatasourceConfig = {
  jsonData: {
    /** OPTIONAL. Storage type (http or local). */
    storage?: string;
    /** OPTIONAL. Default query string appended to all requests. */
    queryParams?: string;
  };
  secureJsonData: {};
};
