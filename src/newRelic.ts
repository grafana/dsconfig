export type newRelicConfig = {
  jsonData: {
    /**
     * OPTIONAL.
     *
     * New Relic data center region. Determines which API endpoints the plugin connects to.
     *
     * Values:
     * - "US": United States data center (default).
     * - "EU": European Union data center.
     *
     * Source definition:
     * - Frontend: NewRelicJsonData.region
     *   https://github.com/grafana/plugins-private/blob/main/plugins/grafana-newrelic-datasource/src/types.ts#L5-L7
     * - Backend: Settings.Region
     *   https://github.com/grafana/plugins-private/blob/main/plugins/grafana-newrelic-datasource/pkg/models/settings.go#L12-L13
     *
     * UI hints:
     * - Shown as a dropdown selector with "US" and "EU" options.
     *   https://github.com/grafana/plugins-private/blob/main/plugins/grafana-newrelic-datasource/src/types.ts#L183-L186
     *   https://github.com/grafana/plugins-private/blob/main/plugins/grafana-newrelic-datasource/src/components/ConfigEditor.tsx#L0-L191
     */
    region?: NewRelicSupportedRegion;

    /**
     * OPTIONAL.
     *
     * HTTP request timeout in seconds for New Relic API calls.
     *
     * Source definition:
     * - Frontend: NewRelicJsonData.timeoutInSeconds
     *   https://github.com/grafana/plugins-private/blob/main/plugins/grafana-newrelic-datasource/src/types.ts#L5-L7
     * - Backend: Settings.TimeoutInSeconds
     *   https://github.com/grafana/plugins-private/blob/main/plugins/grafana-newrelic-datasource/pkg/models/settings.go#L13
     *
     * Backend behavior:
     * - Defaults to 300 seconds (5 minutes) if not set or less than 1.
     *   https://github.com/grafana/plugins-private/blob/main/plugins/grafana-newrelic-datasource/pkg/models/settings.go#L37-L39
     *
     * Defaults:
     * - 300 (seconds).
     */
    timeoutInSeconds?: number;

    /**
     * OPTIONAL. Internal / testing only — not exposed in the UI.
     *
     * Override for the New Relic REST API base URL. Used for internal testing and mocking.
     *
     * Source definition:
     * - Backend: Settings.RestBaseUrl
     *   https://github.com/grafana/plugins-private/blob/main/plugins/grafana-newrelic-datasource/pkg/models/settings.go#L21
     *
     * Backend behavior:
     * - Parsed from JSON but not exposed in the ConfigEditor. Used to override the default
     *   REST API endpoint for testing.
     *   https://github.com/grafana/plugins-private/blob/main/plugins/grafana-newrelic-datasource/pkg/models/settings_test.go#L11-L34
     */
    restBaseURL?: string;

    /**
     * OPTIONAL. Internal / testing only — not exposed in the UI.
     *
     * Override for the New Relic Infrastructure API base URL.
     *
     * Source definition:
     * - Backend: Settings.InfraBaseUrl
     *   https://github.com/grafana/plugins-private/blob/main/plugins/grafana-newrelic-datasource/pkg/models/settings.go#L22
     */
    infrastructureBaseURL?: string;

    /**
     * OPTIONAL. Internal / testing only — not exposed in the UI.
     *
     * Override for the New Relic NerdGraph (GraphQL) API base URL.
     *
     * Source definition:
     * - Backend: Settings.NerdGraphBaseURL
     *   https://github.com/grafana/plugins-private/blob/main/plugins/grafana-newrelic-datasource/pkg/models/settings.go#L23
     */
    nerdGraphBaseURL?: string;
  };

  secureJsonData: {
    /**
     * REQUIRED.
     *
     * New Relic Account ID. Identifies which New Relic account to query.
     * Stored as a string in secureJsonData but parsed to an integer on the backend.
     *
     * Source definition:
     * - Frontend: NewRelicSecureJsonData.accountId
     *   https://github.com/grafana/plugins-private/blob/main/plugins/grafana-newrelic-datasource/src/types.ts#L8-L11
     * - Backend: Settings.AccountID (parsed from DecryptedSecureJSONData["accountId"])
     *   https://github.com/grafana/plugins-private/blob/main/plugins/grafana-newrelic-datasource/pkg/models/settings.go#L17-L18
     *
     * Backend behavior:
     * - Parsed from string to int via strconv.Atoi; non-numeric values result in AccountID = 0.
     *   https://github.com/grafana/plugins-private/blob/main/plugins/grafana-newrelic-datasource/pkg/models/settings.go#L41-L45
     * - If AccountID is 0, health check fails with "Missing Configuration" error.
     *   https://github.com/grafana/plugins-private/blob/main/plugins/grafana-newrelic-datasource/pkg/datasource/handler_checkhealth.go#L136-L153
     *
     * UI hints:
     * - Rendered as a SecretInput field in the ConfigEditor.
     *   https://github.com/grafana/plugins-private/blob/main/plugins/grafana-newrelic-datasource/src/components/ConfigEditor.tsx#L0-L191
     * - Legacy: previously stored in jsonData; the ConfigEditor migrates it to secureJsonData on mount.
     *   https://github.com/grafana/plugins-private/blob/main/plugins/grafana-newrelic-datasource/src/components/ConfigEditor.tsx#L15-L40
     */
    accountId?: string;

    /**
     * REQUIRED.
     *
     * New Relic Personal API Key (User key). Used for authenticating all API requests
     * (REST API, NerdGraph, Insights, Infrastructure).
     *
     * Source definition:
     * - Frontend: NewRelicSecureJsonData.personalApiKey
     *   https://github.com/grafana/plugins-private/blob/main/plugins/grafana-newrelic-datasource/src/types.ts#L8-L11
     * - Backend: Settings.PersonalAPIKey (from DecryptedSecureJSONData["personalApiKey"])
     *   https://github.com/grafana/plugins-private/blob/main/plugins/grafana-newrelic-datasource/pkg/models/settings.go#L16-L17
     *
     * Backend behavior:
     * - Read from DecryptedSecureJSONData["personalApiKey"].
     *   https://github.com/grafana/plugins-private/blob/main/plugins/grafana-newrelic-datasource/pkg/models/settings.go#L33
     * - If empty, health check fails with "Missing Configuration" error: "No Personal API Key".
     *   https://github.com/grafana/plugins-private/blob/main/plugins/grafana-newrelic-datasource/pkg/datasource/handler_checkhealth.go#L136-L153
     * - Validated during health check by making an info query call to the New Relic API.
     *   https://github.com/grafana/plugins-private/blob/main/plugins/grafana-newrelic-datasource/pkg/datasource/handler_checkhealth.go#L103-L174
     *
     * UI hints:
     * - Rendered as a SecretInput field in the ConfigEditor.
     *   https://github.com/grafana/plugins-private/blob/main/plugins/grafana-newrelic-datasource/src/components/ConfigEditor.tsx#L0-L191
     */
    personalApiKey?: string;
  };
};

/**
 * New Relic data center region.
 *
 * Source definition:
 * https://github.com/grafana/plugins-private/blob/main/plugins/grafana-newrelic-datasource/src/types.ts#L4
 */
export type NewRelicSupportedRegion = "US" | "EU";
