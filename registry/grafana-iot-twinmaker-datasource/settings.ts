/**
 * Configuration models for the AWS IoT TwinMaker datasource, which ships as
 * a nested datasource inside the AWS IoT TwinMaker App plugin
 * (`grafana-iot-twinmaker-app`). The datasource itself has plugin id
 * `grafana-iot-twinmaker-datasource`.
 *
 * Sources of truth (all read at pinned versions):
 * - App plugin (https://github.com/grafana/grafana-iot-twinmaker-app@a24885e):
 *   - `src/plugin.json` — the APP plugin.json (type: app,
 *     id: grafana-iot-twinmaker-app). Not used as a schema anchor except to
 *     confirm the includes[] list.
 *   - `src/datasource/plugin.json:1-24` — the nested DATASOURCE plugin.json
 *     (type: datasource, id: grafana-iot-twinmaker-datasource,
 *     name: "AWS IoT TwinMaker"). This is the pluginType / pluginName in
 *     `dsconfig.json`.
 *   - `src/datasource/components/ConfigEditor.tsx:1-172` — the datasource
 *     config editor: `<ConnectionConfig standardRegions=...>`, an Alert when
 *     assumeRoleArn is unset, a `<SecureSocksProxySettings>` (excluded per
 *     AGENTS.md), and a custom "Twinmaker Settings" section with a Workspace
 *     Select, an alarm-config Switch, and a conditional "Assume Role ARN
 *     Write" Input.
 *   - `src/datasource/types.ts:24-31` — `TwinMakerDataSourceOptions extends
 *     AwsAuthDataSourceJsonData` adds `workspaceId?` + `assumeRoleArnWriter?`;
 *     `TwinMakerSecureJsonData extends AwsAuthDataSourceSecureJsonData` adds
 *     an unused `anythingSecure?` placeholder.
 *   - `src/datasource/regions.ts:3-15` — the TwinMaker-specific
 *     supportedRegions list (11 regions, no `Edge` sentinel, no `us-east-2`).
 *   - `pkg/models/settings.go:12-45` — backend `TwinMakerDataSourceSetting`
 *     embeds `awsds.AWSDatasourceSettings` and adds `AssumeRoleARNWriter`,
 *     `WorkspaceID`, and an in-memory `UID`. `Load` reads jsonData verbatim
 *     but only copies `accessKey`/`secretKey` from decrypted secure data
 *     (not `sessionToken` — see the discrepancy note in the README).
 *   - `pkg/plugin/datasource.go:171-184` — `CheckHealth` fails when
 *     `workspaceId` or `assumeRoleArn` is empty; this is the backend
 *     contract encoded as `requiredWhen: "true"` in the schema.
 * - `@grafana/aws-sdk` `0.8.3` (grafana/grafana-aws-sdk-react tag `v0.8.3`):
 *   - `src/types.ts` — `AwsAuthType`, `AwsAuthDataSourceJsonData`,
 *     `AwsAuthDataSourceSecureJsonData`, `ConnectionConfigProps`
 *   - `src/components/ConnectionConfig.tsx` — every AWS field's label,
 *     placeholder, description, and conditional render. Note twinmaker does
 *     NOT pass `showHttpProxySettings`, `hideAssumeRoleArn`, or
 *     `skipEndpoint`, so Assume Role and Endpoint are visible for every
 *     non-`grafana_assume_role` provider. `DS_TYPES_THAT_SUPPORT_TEMP_CREDS`
 *     lists only `cloudwatch`/`grafana-athena-datasource`/`grafana-
 *     amazonprometheus-datasource`, so Grafana Assume Role is filtered out
 *     of the editor for TwinMaker regardless of feature toggles.
 *   - `src/providers.ts:4-25` — `awsAuthProviderOptions` (Select options for
 *     `authType`).
 * - Backend `grafana-aws-sdk` `v1.4.3` (`pkg/awsds/settings.go`):
 *   - `AWSDatasourceSettings` — Go struct that receives jsonData; note the
 *     `assumeRoleARN` (uppercase RN) json tag versus the frontend's
 *     `assumeRoleArn` (lowercase arn). Case-insensitive Unmarshal rescue.
 */

/**
 * The AWS authentication provider values persisted to `jsonData.authType`.
 *
 * Four values are editor-selectable via the ConnectionConfig `Select` for
 * TwinMaker: `ec2_iam_role`, `default`, `keys`, `credentials`. Two more
 * are storage-valid but not editor-selectable: `grafana_assume_role`
 * (filtered by ConnectionConfig's `DS_TYPES_THAT_SUPPORT_TEMP_CREDS`
 * allow-list) and `arn` (legacy value the backend maps to `default`).
 */
export type AwsAuthType =
  | 'default'
  | 'keys'
  | 'credentials'
  | 'ec2_iam_role'
  | 'grafana_assume_role'
  | 'arn';

/**
 * Root (top-level datasource settings) fields.
 *
 * The TwinMaker datasource stores no plugin-specific fields at the root
 * level (`url`, `basicAuth`, etc. are unused; the plugin's backend Load
 * only consumes jsonData + decrypted secure data), so this is a blank
 * object rather than null.
 */
export type RootConfig = Record<string, never>;

/**
 * Fields stored in `jsonData`. Union of the AWS-shared fields from
 * `AwsAuthDataSourceJsonData` (`@grafana/aws-sdk@0.8.3`, `src/types.ts:15-22`)
 * and the TwinMaker-specific fields from `TwinMakerDataSourceOptions`
 * (`src/datasource/types.ts:24-27`).
 *
 * TwinMaker does NOT pass `showHttpProxySettings` to `ConnectionConfig`, so
 * the AWS proxy fields (`proxyType`, `proxyUrl`, `proxyUsername`) are
 * neither editor-visible nor part of the plugin's declared surface. `region`
 * is declared on `AwsAuthDataSourceJsonData` but the frontend never writes
 * it; the backend mirrors `defaultRegion` into `Region` at load time
 * (`pkg/models/settings.go:30-32`).
 */
export type JsonDataConfig = {
  // ---- AWS SDK ConnectionConfig fields (`@grafana/aws-sdk@0.8.3`) ----

  /** AWS credentials chain to use. Discriminator for accessKey/secretKey/profile/assumeRoleArn. */
  authType?: AwsAuthType;
  /** Credentials profile name from `~/.aws/credentials`. Editor-visible when `authType === 'credentials'`. */
  profile?: string;
  /**
   * ARN of an IAM role to assume via STS. Editor-visible for every auth type
   * except `grafana_assume_role`. Required at runtime by TwinMaker's
   * CheckHealth (`pkg/plugin/datasource.go:179-184`), even though the
   * ConnectionConfig field label calls it "Optional". Note: the backend's
   * json tag is `assumeRoleARN` (uppercase RN); Go's case-insensitive
   * Unmarshal makes the frontend spelling work too.
   */
  assumeRoleArn?: string;
  /** External ID passed to STS AssumeRole. Editor-visible when `authType !== 'grafana_assume_role'`. */
  externalId?: string;
  /** Optional custom AWS service endpoint. Editor-visible when `authType !== 'grafana_assume_role'`. */
  endpoint?: string;
  /**
   * Default AWS region. Editor writes `us-east-1` on mount if empty
   * (`ConfigEditor.tsx:31-36`); backend Load falls back to `us-east-1` as
   * well (`pkg/models/settings.go:30-35`).
   */
  defaultRegion?: string;

  // ---- TwinMaker-specific (`TwinMakerDataSourceOptions`, `src/datasource/types.ts:24-27`) ----

  /**
   * TwinMaker workspace id. Required at runtime (`CheckHealth` at
   * `pkg/plugin/datasource.go:172-177`). The editor renders a Select that
   * lazy-loads options from `datasource.info.listWorkspaces()` once the
   * datasource has been saved (`ConfigEditor.tsx:120-140`), and also
   * supports free-form entry via `allowCustomValue`.
   */
  workspaceId?: string;
  /**
   * Optional ARN of a second STS role used by the AWS IoT TwinMaker Alarm
   * Configuration Panel when writing property values
   * (`pkg/models/settings.go:63-66 → ToAWSDatasourceSettingsWriter`).
   * Editor-visible only when the "Define write permissions for Alarm
   * Configuration Panel" switch is on (`ConfigEditor.tsx:141-162`); toggling
   * the switch off clears the field.
   */
  assumeRoleArnWriter?: string;
};

/**
 * Secret key names stored in `secureJsonData` (write-only; read existing
 * config via `secureJsonFields`).
 *
 * - `accessKey` / `secretKey` — set when `authType === 'keys'`
 *
 * The AWS shared secure shape (`AwsAuthDataSourceSecureJsonData`) also
 * declares `sessionToken`, but TwinMaker's backend `Load`
 * (`pkg/models/settings.go:37-38`) only copies `accessKey` and `secretKey`
 * from decrypted secure data. `sessionToken` is not read; setting it has no
 * effect. See the README's discrepancies section.
 */
export type SecureJsonDataConfig = Array<'accessKey' | 'secretKey'>;
