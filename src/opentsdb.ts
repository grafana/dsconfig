export type opentsdbConfig = {
  jsonData: {
    /** REQUIRED. OpenTSDB version (1, 2, or 3). */
    tsdbVersion: number;
    /** REQUIRED. Time resolution (1=second, 2=millisecond). */
    tsdbResolution: number;
    /** REQUIRED. Maximum number of results for metric lookups. */
    lookupLimit: number;
  };
  secureJsonData: {};
};
