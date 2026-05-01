export type googleSheetsConfig = {
  jsonData: {
    authenticationType?: GoogleAuthType;
    authType?: GoogleAuthType;
    defaultProject?: string;
    clientEmail?: string;
    tokenUri?: string;
    privateKeyPath?: string;
    serviceAccountToImpersonate?: string;
    usingImpersonation?: boolean;
    workloadIdentityPoolProvider?: string;
    wifServiceAccountEmail?: string;
    defaultSheetID?: string;
  };
  secureJsonData: {
    privateKey?: string;
    apiKey?: string;
    jwt?: string;
  };
};

export type GoogleAuthType = "jwt" | "key" | "gce" | "workloadIdentityFederation";
