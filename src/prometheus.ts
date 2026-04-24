export type prometheusConfig = {
    url: string;
    basicAuth: boolean;
    basicAuthUser?: string;
    jsonData: {
        timeInterval?: string;
        queryTimeout?: string;
        httpMethod?: string; // type HTTP_METHOD = "GET" | "POST"
        customQueryParameters?: string;
        disableMetricsLookup?: boolean;
        exemplarTraceIdDestinations?: ExemplarTraceIdDestination[];
        prometheusType?: PromApplication; // type PromApplication = 'Prometheus' | 'Mimir' | 'Cortex' | 'Thanos'
        prometheusVersion?: string;
        cacheLevel?: PrometheusCacheLevel; // type PrometheusCacheLevel = 'Low' | 'Medium' | 'High' | 'None'
        defaultEditor?: QueryEditorMode; // type QueryEditorMode = 'code' | 'builder'
        incrementalQuerying?: boolean;
        incrementalQueryOverlapWindow?: string;
        disableRecordingRules?: boolean;
        manageAlerts?: boolean;
        allowAsRecordingRulesTarget?: boolean;
        oauthPassThru?: boolean;
        seriesEndpoint?: boolean;
        seriesLimit?: number;
        timeout?: number;
        serverName?: string;
        tlsAuth?: boolean;
        tlsAuthWithCACert?: boolean;
        tlsSkipVerify?: boolean;
        keepCookies?: string[];
        pdcInjected?: boolean;
    };
    secureJsonData: {
        basicAuthPassword?: string;
        tlsCACert?: string;
        tlsClientCert?: string;
        tlsClientKey?: string;
    };
};

enum PromApplication {
    Cortex = "Cortex",
    Mimir = "Mimir",
    Prometheus = "Prometheus",
    Thanos = "Thanos",
}

enum PrometheusCacheLevel {
    Low = "Low",
    Medium = "Medium",
    High = "High",
    None = "None",
}

enum QueryEditorMode {
    Code = "code",
    Builder = "builder",
}

type ExemplarTraceIdDestination = {
    name: string;
    url?: string;
    urlDisplayLabel?: string;
    datasourceUid?: string;
};
