/**
 * Configuration models for the Amazon Redshift datasource plugin
 * (`grafana-redshift-datasource`).
 *
 * Sources of truth (all read at pinned versions):
 * - Plugin (https://github.com/grafana/redshift-datasource@5bb9376):
 *   - `src/plugin.json` — plugin id, name, docs link
 *   - `src/ConfigEditor/ConfigEditor.tsx` — Redshift Details subsection
 *     (useManagedSecret radio via AuthTypeSwitch, useServerless switch,
 *     clusterIdentifier/workgroupName ConfigSelect, managedSecret ConfigSelect,
 *     dbUser Input, database Input, withEvent Switch) plus the embedded
 *     `@grafana/aws-sdk` ConnectionConfig — Redshift does NOT pass
 *     `showHttpProxySettings`, so the AWS proxy fields are not editor-visible
 *   - `src/ConfigEditor/AuthTypeSwitch.tsx` — the "Temporary credentials" /
 *     "AWS Secrets Manager" RadioButtonGroup that writes `jsonData.useManagedSecret`
 *   - `src/selectors.ts` — Field labels (`UseServerless.input="Serverless"`,
 *     `ManagedSecret.input="Managed Secret"`, `ClusterID.input="Cluster Identifier"`,
 *     `Workgroup.input="Workgroup"`, `Database.input="Database"`,
 *     `DatabaseUser.input="Database User"`,
 *     `WithEvent.input="Send events to Amazon EventBridge"`)
 *   - `src/types.ts` — `RedshiftDataSourceOptions extends AwsAuthDataSourceJsonData`
 *     and `RedshiftDataSourceSecureJsonData extends AwsAuthDataSourceSecureJsonData`
 * - `@grafana/aws-sdk` `0.10.2`
 *   (github.com/grafana/grafana-aws-sdk-react tag `v0.10.2`):
 *   - `src/types.ts` — `AwsAuthType`, `AwsAuthDataSourceJsonData`,
 *     `AwsAuthDataSourceSecureJsonData`
 *   - `src/components/ConnectionConfig.tsx` — every AWS field's label,
 *     placeholder, description, and conditional render
 *   - `src/providers.ts` — the Select options for `authType`
 * - Backend `grafana-aws-sdk` `v1.4.3` (`pkg/awsds/settings.go`):
 *   - `AWSDatasourceSettings` — the Go struct the backend embeds. Its custom
 *     `AuthType.UnmarshalJSON` maps legacy `sharedCreds` and `arn` to the
 *     modern values, which is why the schema's allowedValues includes `arn`.
 * - Backend plugin (`pkg/redshift/models/settings.go`):
 *   - `RedshiftDataSourceSettings` — embeds `awsds.AWSDatasourceSettings` and
 *     adds `ClusterIdentifier`, `WorkgroupName`, `Database`, `UseServerless`,
 *     `UseManagedSecret`, `WithEvent`, `DBUser`, and (untagged) `ManagedSecret`.
 */

/**
 * The AWS authentication provider values persisted to `jsonData.authType`.
 *
 * Five values are editor-selectable via the ConnectionConfig `Select`:
 * `ec2_iam_role`, `grafana_assume_role`, `default`, `keys`, `credentials`.
 * A sixth value, `arn`, is a legacy stored value the backend maps to
 * `default` (see `awsds.AuthType.UnmarshalJSON`).
 */
export type AwsAuthType =
  | 'default'
  | 'keys'
  | 'credentials'
  | 'ec2_iam_role'
  | 'grafana_assume_role'
  | 'arn';

/**
 * The AWS Secrets Manager selection stored in `jsonData.managedSecret`. The
 * editor writes both fields from a single Select whose value is the secret
 * ARN and whose label is the secret name (`src/ConfigEditor/ConfigEditor.tsx:189-199`).
 * Only `arn` is user-selectable; `name` comes along for the ride.
 */
export type RedshiftManagedSecret = {
  /** ARN of the AWS Secrets Manager secret. Required when useManagedSecret is true. */
  arn: string;
  /** Human-readable name of the secret (Select label). */
  name: string;
};

/**
 * Root (top-level datasource settings) fields.
 *
 * The Redshift datasource stores no plugin-specific fields at the root level
 * (`url`, `basicAuth`, etc. are unused by the plugin's backend; the editor
 * DOES rewrite the root `url` to `${clusterEndpoint}/${database}` or
 * `${workgroupEndpoint}/${database}` as a display convenience — see
 * `ConfigEditor.tsx:152-172` — but `pkg/redshift/models/settings.go` never
 * reads it back). This is a blank object rather than null.
 */
export type RootConfig = Record<string, never>;

/**
 * Fields stored in `jsonData`. Union of the AWS-shared fields from
 * `AwsAuthDataSourceJsonData` (types.ts:15-25 of `@grafana/aws-sdk@0.10.2`)
 * and the Redshift-specific fields from `RedshiftDataSourceOptions`
 * (`src/types.ts:51-64`).
 *
 * The Redshift config editor does not render the AWS proxy fields
 * (`proxyType`, `proxyUrl`, `proxyUsername`) because `ConfigEditor.tsx:245`
 * calls `<ConnectionConfig>` without `showHttpProxySettings`, so those
 * fields are omitted from this shape.
 */
export type JsonDataConfig = {
  // ---- AWS SDK ConnectionConfig fields (`@grafana/aws-sdk@0.10.2`) ----

  /** AWS credentials chain to use. Discriminator for accessKey/secretKey/profile/assumeRoleArn. */
  authType?: AwsAuthType;
  /** Credentials profile name from `~/.aws/credentials`. Editor-visible when `authType === 'credentials'`. */
  profile?: string;
  /**
   * ARN of an IAM role to assume via STS. Editor-visible for every auth type
   * except `grafana_assume_role`. Note: the backend's json tag is
   * `assumeRoleARN` (uppercase RN); Go's case-insensitive Unmarshal makes the
   * frontend spelling work too.
   */
  assumeRoleArn?: string;
  /** External ID passed to STS AssumeRole. Editor-visible when `authType !== 'grafana_assume_role'`. */
  externalId?: string;
  /** Optional custom AWS service endpoint. Editor-visible when `authType !== 'grafana_assume_role'`. */
  endpoint?: string;
  /** Default AWS region, e.g. `us-east-1`. Also used to seed the runtime `region` (backend). */
  defaultRegion?: string;

  // ---- Redshift-specific (`RedshiftDataSourceOptions`) ----

  /**
   * Provisioning mode. `false` = Provisioned (requires `clusterIdentifier`);
   * `true` = Serverless (requires `workgroupName`).
   */
  useServerless?: boolean;
  /**
   * Credential mode. `false` = temporary IAM creds via GetClusterCredentials
   * (Provisioned) or GetCredentials (Serverless) — requires `dbUser` on
   * Provisioned; `true` = read a secret from AWS Secrets Manager — requires
   * `managedSecret.arn` and populates `dbUser` from the secret.
   */
  useManagedSecret?: boolean;
  /** Redshift Provisioned cluster identifier. Required when `useServerless === false`. */
  clusterIdentifier?: string;
  /** Redshift Serverless workgroup name. Required when `useServerless === true`. */
  workgroupName?: string;
  /**
   * AWS Secrets Manager selection (ARN + display name). Required when
   * `useManagedSecret === true`.
   */
  managedSecret?: RedshiftManagedSecret;
  /**
   * Database user for temporary IAM credential minting. Required on
   * Provisioned when `useManagedSecret === false`; auto-populated from the
   * secret when `useManagedSecret === true`.
   */
  dbUser?: string;
  /** Redshift database name. Always required by the backend to run queries. */
  database?: string;
  /**
   * Send Redshift query execution events to Amazon EventBridge. Optional
   * feature toggle rendered as a switch (`ConfigEditor.tsx:368-384`).
   */
  withEvent?: boolean;
};

/**
 * Secret key names stored in `secureJsonData` (write-only; read existing
 * config via `secureJsonFields`).
 *
 * - `accessKey` / `secretKey` — set when `authType === 'keys'`
 * - `sessionToken` — backend-only (`pkg/redshift/models/settings.go:67`);
 *   no editor UI writes it, provisioning-only
 */
export type SecureJsonDataConfig = Array<'accessKey' | 'secretKey' | 'sessionToken'>;
