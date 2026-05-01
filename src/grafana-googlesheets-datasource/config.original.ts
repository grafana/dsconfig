import {
  type DataSourceOptions,
  type DataSourceSecureJsonData,
  GoogleAuthType,
  GOOGLE_AUTH_TYPE_OPTIONS,
} from "@grafana/google-sdk";

export const GoogleSheetsAuth = {
  ...GoogleAuthType,
  API: "key",
} as const;

export const googleSheetsAuthTypes = [
  { label: "API Key", value: GoogleSheetsAuth.API },
  ...GOOGLE_AUTH_TYPE_OPTIONS,
];

export interface GoogleSheetsSecureJSONData extends DataSourceSecureJsonData {
  apiKey?: string;
}

export interface GoogleSheetsDataSourceOptions extends DataSourceOptions {
  defaultSheetID?: string;
}
