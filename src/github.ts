export type githubConfig = {
    jsonData: {
        selectedAuthType?: GitHubAuthType;
        githubPlan?: GitHubLicenseType;
        githubUrl?: string;
        appId?: string;
        installationId?: string;
    };
    secureJsonData: {
        accessToken?: string;
        privateKey?: string;
    };
};

type GitHubLicenseType = 'github-basic' | 'github-enterprise-cloud' | 'github-enterprise-server';

type GitHubAuthType = 'personal-access-token' | 'github-app';
