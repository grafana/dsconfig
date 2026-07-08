/**
 * Configuration models for the DynamoDB datasource plugin
 * (`grafana-dynamodb-datasource`).
 *
 * Sources of truth (all read at pinned versions):
 * - Plugin (https://github.com/grafana/dynamodb-datasource@44f7fd6):
 *   - `src/plugin.json` — plugin id (`grafana-dynamodb-datasource`), name
 *     ("DynamoDB"), docs link
 *   - `src/components/ConfigEditor.tsx` — the entire config editor:
 *     `<DataSourceDescription>` + `<ConnectionConfig hideAssumeRoleArn>`
 *     (nothing else — no plugin-specific UI fields). Also runs a
 *     useEffect that writes `jsonData.isV2 = true` + `jsonData.authType =
 *     Keys` on fresh datasources and calls `migrateOptions` for pre-V2
 *     configs.
 *   - `src/types.ts:19-31` — `DynamoDBConfigOptions extends
 *     AwsAuthDataSourceJsonData` adds `timeout`, `retries`, `pause` (driver
 *     settings — string on the wire, parsed with `utils.ParseInt`), `isV2`
 *     (migration marker), plus two legacy fields carried for pre-V2
 *     compatibility: `region` and `accessId`. `DynamoDBSecureConfigOptions
 *     extends AwsAuthDataSourceSecureJsonData` adds nothing.
 *   - `src/utils.ts:4-22` — `migrateOptions`: drops `region`/`accessId`
 *     from jsonData, copies `region → defaultRegion`, keeps `endpoint`,
 *     sets `authType = Keys` and `isV2 = true`, and re-marks
 *     `accessKey`/`secretKey` as configured secureJsonFields.
 *   - `pkg/models/settings.go:26-85` — backend `Settings` (embeds
 *     `awsds.AWSDatasourceSettings`) + `LoadSettings`: on non-V2 loads it
 *     forces `authType = keys`, folds `LegacyAccessKey` (jsonData.accessId)
 *     into `AccessKey`, uses `secureJsonData.accessKey` as the SECRET key
 *     (V1 naming quirk), and copies `region → defaultRegion`. Timeout /
 *     Pause / Retries default to "60" / "5" / "5" as strings.
 * - `@grafana/aws-sdk` `0.10.2` (grafana/grafana-aws-sdk-react tag
 *   `v0.10.2`, SHA `fe0c4d8`):
 *   - `src/components/ConnectionConfig.tsx` — every AWS field's label,
 *     placeholder, description, and conditional render. DynamoDB passes
 *     `hideAssumeRoleArn`, so the entire Assume Role subsection is hidden
 *     and neither `assumeRoleArn` nor `externalId` is editor-visible.
 *     DynamoDB does NOT pass `showHttpProxySettings`, so the proxy fields
 *     are not visible either. DynamoDB does NOT pass `skipEndpoint` or
 *     `defaultEndpoint`, so the Endpoint field is visible with the generic
 *     `https://{service}.{region}.amazonaws.com` placeholder for every
 *     provider.
 *   - `src/types.ts` — `AwsAuthType`, `AwsAuthDataSourceJsonData`,
 *     `AwsAuthDataSourceSecureJsonData`, `ConnectionConfigProps`
 *   - `src/providers.ts` — the Select options for `authType`. Note the
 *     `grafana_assume_role` entry is filtered out for DynamoDB because
 *     `grafana-dynamodb-datasource` is NOT in
 *     `DS_TYPES_THAT_SUPPORT_TEMP_CREDS` (ConnectionConfig.tsx:18-28).
 * - Backend `grafana-aws-sdk` `v1.4.6` (`pkg/awsds/settings.go`):
 *   - `AWSDatasourceSettings` — Go struct that receives jsonData; note
 *     `AssumeRoleARN` uses `json:"assumeRoleARN"` (uppercase RN) versus
 *     the frontend's camelCase `assumeRoleArn`. Not relevant to DynamoDB
 *     because those fields are hidden, but included in the awsds surface
 *     for completeness.
 *   - `AuthType.MarshalJSON`/`UnmarshalJSON` — the storage⇆enum mapping
 *     that folds legacy `arn`→`default` and `sharedCreds`→`credentials`.
 */

/**
 * The AWS authentication provider values persisted to `jsonData.authType`.
 *
 * Four values are editor-selectable in DynamoDB's Select (`providers.ts`
 * minus `grafana_assume_role`, which ConnectionConfig filters out because
 * `grafana-dynamodb-datasource` is NOT in `DS_TYPES_THAT_SUPPORT_TEMP_CREDS`,
 * ConnectionConfig.tsx:18-28): `ec2_iam_role`, `default`, `keys`,
 * `credentials`. A fifth value, `arn`, is a legacy stored value the backend
 * maps to `default` (see `awsds.AuthType.UnmarshalJSON` /
 * `awsds/settings.go:87-88`).
 */
export type AwsAuthType =
  | 'default'
  | 'keys'
  | 'credentials'
  | 'ec2_iam_role'
  | 'arn';

/**
 * Root (top-level datasource settings) fields.
 *
 * The DynamoDB datasource stores no plugin-specific fields at the root
 * level (`url`, `basicAuth`, etc. are unused), so this is a blank object
 * rather than null. The plugin's backend `LoadSettings`
 * (`pkg/models/settings.go:38-85`) never reads any root-level field.
 */
export type RootConfig = Record<string, never>;

/**
 * Fields stored in `jsonData`. Union of the AWS-shared fields from
 * `AwsAuthDataSourceJsonData` (`@grafana/aws-sdk@0.10.2`,
 * `src/types.ts:15-25`) that DynamoDB's ConfigEditor exposes, plus the
 * DynamoDB-specific fields from `DynamoDBConfigOptions`
 * (`src/types.ts:19-26`).
 *
 * DynamoDB passes `hideAssumeRoleArn` to `ConnectionConfig`, so
 * `assumeRoleArn`/`externalId` are NOT editor-visible and are omitted
 * here. DynamoDB does not pass `showHttpProxySettings`, so the AWS proxy
 * fields are omitted too. `region` is a legacy V1-only field the backend
 * folds into `defaultRegion` when the latter is empty
 * (`pkg/models/settings.go:47-49`); it's kept here for round-trip
 * fidelity but marked legacy.
 */
export type JsonDataConfig = {
  // ---- AWS SDK ConnectionConfig fields (@grafana/aws-sdk@0.10.2) ----

  /** AWS credentials chain to use. Discriminator for accessKey/secretKey/profile. */
  authType?: AwsAuthType;
  /** Credentials profile name from `~/.aws/credentials`. Editor-visible when `authType === 'credentials'`. */
  profile?: string;
  /** Optional custom AWS service endpoint. Editor-visible for every provider (DynamoDB does not pass `skipEndpoint`). */
  endpoint?: string;
  /** Default AWS region, e.g. `us-east-1`. Also used to seed the runtime `region` (backend). */
  defaultRegion?: string;

  // ---- DynamoDB-specific (`DynamoDBConfigOptions`, src/types.ts:19-26) ----

  /**
   * V2 marker written by the editor's useEffect on fresh datasources
   * (`src/components/ConfigEditor.tsx:28-42`). Frontend-only in the sense
   * that no UI control writes it; the backend still reads it in
   * `pkg/models/settings.go:44` to decide whether to trigger V1
   * migration.
   */
  isV2?: boolean;
  /**
   * Query timeout in seconds. String on the wire (parsed with
   * `utils.ParseInt` server-side). Defaults to `"60"` when empty
   * (`pkg/models/settings.go:75-77`). Backend-only: no editor UI.
   */
  timeout?: string;
  /**
   * Retry count for the DynamoDB SQL driver. String on the wire.
   * Defaults to `"5"` when empty (`pkg/models/settings.go:81-83`).
   * Backend-only.
   */
  retries?: string;
  /**
   * Pause (in seconds) between retries. String on the wire. Defaults to
   * `"5"` when empty (`pkg/models/settings.go:78-80`). Backend-only.
   */
  pause?: string;
  /**
   * Legacy V1 field — pre-V2 datasources stored the region here rather
   * than in `defaultRegion`. The backend copies it into `defaultRegion`
   * when the latter is empty (`pkg/models/settings.go:47-49`). Read by
   * the backend; the frontend's `migrateOptions` strips it on edit.
   */
  region?: string;
  /**
   * Legacy V1 field — pre-V2 datasources stored the AWS Access Key ID
   * here as plain jsonData (matched by the `LegacyAccessKey` backend
   * tag, `pkg/models/settings.go:29`). Under V1, the secret was stored
   * as `secureJsonData.accessKey`; V1 migration
   * (`pkg/models/settings.go:44-50`) folds this into
   * `awsds.AWSDatasourceSettings.AccessKey`. Backend-only; the
   * frontend's `migrateOptions` strips it on edit.
   */
  accessId?: string;
};

/**
 * Secret key names stored in `secureJsonData` (write-only; read existing
 * config via `secureJsonFields`).
 *
 * - `accessKey` / `secretKey` — set when `authType === 'keys'` (V2). Note
 *   that under V1 storage, `secureJsonData.accessKey` actually held the
 *   AWS **Secret** Access Key (the naming was fixed at V2); see the
 *   README for the migration table.
 * - `sessionToken` — backend-only (`pkg/models/settings.go:50,54`); no
 *   editor UI writes it. Used for temporary STS credentials.
 */
export type SecureJsonDataConfig = Array<'accessKey' | 'secretKey' | 'sessionToken'>;
