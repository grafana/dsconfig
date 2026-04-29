export type githubConfig = {
    url: string;
    basicAuth: boolean;
    basicAuthUser?: string;
    jsonData?: {
        selectedAuthType?: 'personal-access-token' | 'github-app';
        githubPlan?: 'github-basic' | 'github-enterprise-cloud' | 'github-enterprise-server';
        githubUrl?: string;
        appId?: string;
        installationId?: string;
        timeout?: number;
        serverName?: string;
        tlsAuth?: boolean;
        tlsAuthWithCACert?: boolean;
        tlsSkipVerify?: boolean;
        keepCookies?: string[];
        pdcInjected?: boolean;
    };
    secureJsonData?: {
        basicAuthPassword?: string;
        accessToken?: string;
        privateKey?: string;
        tlsCACert?: string;
        tlsClientCert?: string;
        tlsClientKey?: string;
    };
};
