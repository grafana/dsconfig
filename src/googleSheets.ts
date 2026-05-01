export type googleSheetsConfig = {
  jsonData: {
    /**
     * REQUIRED.
     *
     * Which authentication mechanism the datasource will use to talk to Google APIs.
     *
     * Values:
     * - "jwt": Service account JWT / JSON key (supports private spreadsheets).
     * - "key": API key (spreadsheets must be public).
     * - "gce": Default GCE service account (Grafana must run on Google Compute Engine).
     *
     * Provisioning examples:
     * https://github.com/grafana/google-sheets-datasource/blob/7a4e822d46faba2404778c756a6847f6fecef27e/docs/sources/configure.md#L176-L256
     *
     * Backend behavior:
     * - If empty, backend errors with "missing AuthenticationType setting".
     *   https://github.com/grafana/google-sheets-datasource/blob/7a4e822d46faba2404778c756a6847f6fecef27e/pkg/googlesheets/googleclient.go#L123-L131
     *
     * UI hints:
     * - Selected in the ConfigEditor as the "authentication type" selector; UI shows guidance text for each option.
     *   https://github.com/grafana/google-sheets-datasource/blob/7a4e822d46faba2404778c756a6847f6fecef27e/src/components/ConfigEditor.tsx#L1-L100
     *
     * Defaults:
     * - Docs state default is "Google JWT File" (service account) in UI; however, provisioning requires explicitly setting it.
     *   https://github.com/grafana/google-sheets-datasource/blob/7a4e822d46faba2404778c756a6847f6fecef27e/docs/sources/configure.md#L55-L62
     */
    authenticationType?: GoogleSheetsAuthenticationType;

    /**
     * OPTIONAL legacy alias for authenticationType.
     *
     * Backend migration:
     * - If AuthType is set, AuthenticationType = AuthType
     *   https://github.com/grafana/google-sheets-datasource/blob/7a4e822d46faba2404778c756a6847f6fecef27e/pkg/models/settings.go#L45-L52
     *
     * UI normalization:
     * - authenticationType = authenticationType || authType
     *   https://github.com/grafana/google-sheets-datasource/blob/7a4e822d46faba2404778c756a6847f6fecef27e/src/utils.ts#L4-L15
     */
    authType?: GoogleSheetsAuthenticationType;

    /**
     * OPTIONAL.
     *
     * Default Spreadsheet ID to pre-fill in new queries.
     *
     * UI hints:
     * - In JWT mode, UI can load spreadsheet IDs from datasource (if options.uid exists) and show a selectable list.
     *   https://github.com/grafana/google-sheets-datasource/blob/7a4e822d46faba2404778c756a6847f6fecef27e/src/components/ConfigEditor.tsx#L24-L90
     *
     * Docs:
     * - Optional; can select / paste URL / manually enter; pre-fills query editor.
     *   https://github.com/grafana/google-sheets-datasource/blob/7a4e822d46faba2404778c756a6847f6fecef27e/docs/sources/configure.md#L127-L143
     */
    defaultSheetID?: string;

    /**
     * CONDITIONALLY REQUIRED: for `authenticationType: "gce"`.
     *
     * Docs:
     * - Appears as "Default project" only for GCE authentication.
     *   https://github.com/grafana/google-sheets-datasource/blob/7a4e822d46faba2404778c756a6847f6fecef27e/docs/sources/configure.md#L118-L126
     */
    defaultProject?: string;

    /**
     * CONDITIONALLY REQUIRED: for `authenticationType: "jwt"` when provisioning with explicit fields.
     *
     * Docs provisioning example:
     * https://github.com/grafana/google-sheets-datasource/blob/7a4e822d46faba2404778c756a6847f6fecef27e/docs/sources/configure.md#L202-L215
     */
    clientEmail?: string;

    /**
     * CONDITIONALLY REQUIRED: for `authenticationType: "jwt"` when provisioning with explicit fields.
     *
     * Default commonly "https://oauth2.googleapis.com/token" (per docs example).
     * https://github.com/grafana/google-sheets-datasource/blob/7a4e822d46faba2404778c756a6847f6fecef27e/docs/sources/configure.md#L208-L215
     */
    tokenUri?: string;

    /**
     * OPTIONAL: JWT auth alternative for self-hosted envs.
     *
     * Path to a local private key file (docs: not supported in hosted envs like Grafana Cloud).
     * https://github.com/grafana/google-sheets-datasource/blob/7a4e822d46faba2404778c756a6847f6fecef27e/docs/sources/configure.md#L219-L239
     */
    privateKeyPath?: string;
  };

  secureJsonData: {
    /**
     * REQUIRED when `authenticationType: "key"`.
     *
     * UI: SecretInput with placeholder "Enter API key"; reset clears secureJsonFields.apiKey and sets apiKey to ''.
     * https://github.com/grafana/google-sheets-datasource/blob/7a4e822d46faba2404778c756a6847f6fecef27e/src/components/ConfigEditor.tsx#L32-L52
     *
     * Backend: required; otherwise error "missing API Key".
     * https://github.com/grafana/google-sheets-datasource/blob/7a4e822d46faba2404778c756a6847f6fecef27e/pkg/googlesheets/googleclient.go#L133-L144
     */
    apiKey?: string;

    /**
     * REQUIRED when `authenticationType: "jwt"` (unless using privateKeyPath or legacy `jwt`).
     *
     * Docs provisioning example uses secureJsonData.privateKey:
     * https://github.com/grafana/google-sheets-datasource/blob/7a4e822d46faba2404778c756a6847f6fecef27e/docs/sources/configure.md#L202-L215
     */
    privateKey?: string;

    /**
     * OPTIONAL legacy secure field for JWT content.
     *
     * Backend still reads it:
     * https://github.com/grafana/google-sheets-datasource/blob/7a4e822d46faba2404778c756a6847f6fecef27e/pkg/models/settings.go#L41-L44
     */
    jwt?: string;
  };
};

export type GoogleSheetsAuthenticationType = "jwt" | "key" | "gce";