export type grafanaSurrealdbDatasourceConfig = {
  jsonData: {
    /** OPTIONAL. SurrealDB endpoint URL. */
    endpoint?: string;
    /** OPTIONAL. Database name. */
    database?: string;
    /** OPTIONAL. Namespace. */
    namespace?: string;
    /** OPTIONAL. Authentication scope. */
    scope?: string;
    /** OPTIONAL. Username. */
    username?: string;
  };
  secureJsonData: {
    /** OPTIONAL. Password. */
    password?: string;
  };
};
