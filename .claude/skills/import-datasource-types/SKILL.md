---
name: import-datasource-types
description: Generate a detailed TypeScript type file for a Grafana datasource plugin by exploring its GitHub repository. Use when asked to create or update a datasource config type with rich source-linked JSDoc comments.
---

# Import Datasource Types

Generate a fully documented TypeScript type file for a Grafana datasource plugin's provisioning/configuration, with detailed JSDoc comments linking to the actual source code.

## Arguments

The user should provide:
- A **datasource plugin name or GitHub repo** (required), e.g. `google-sheets-datasource`, `grafana/tempo`, `grafana-github-datasource`
- Optionally, a **specific branch or commit SHA** to pin links to (recommended: use a commit SHA for permalink stability)

If only a short name is given (e.g. "cloudwatch", "mysql"), resolve to the canonical Grafana GitHub repo:
- Core datasources live in `grafana/grafana` under `pkg/tsdb/<name>/` and `public/app/plugins/datasource/<name>/`
- Plugin datasources live in `grafana/<name>-datasource` (e.g. `grafana/google-sheets-datasource`)

## Output

A single TypeScript file in `src/<datasourceName>.ts` following the project conventions.

## Steps

### 1. Identify the Source Repository

Resolve the datasource to its GitHub repository:
- If the user gives a full repo like `grafana/tempo`, use that.
- If they give a plugin name like `google-sheets`, resolve to `grafana/google-sheets-datasource`.
- For core Grafana datasources (prometheus, loki, tempo, postgres, mysql, elasticsearch, cloudwatch, etc.), the source is `grafana/grafana` with code split between:
  - Backend: `pkg/tsdb/<name>/` or `pkg/services/ngalert/` etc.
  - Frontend: `public/app/plugins/datasource/<name>/`
  - Plugin JSON: `public/app/plugins/datasource/<name>/plugin.json`

### 2. Find the Settings / Configuration Model

Search the repo for the datasource settings struct or type. Look in these locations (in priority order):

**Backend (Go):**
- `pkg/models/settings.go` or `pkg/models/` — settings struct with JSON tags
- `pkg/**/settings.go` — look for structs like `Settings`, `DatasourceSettings`, `Config`
- `pkg/**/datasource.go` — constructor that reads settings from `backend.DataSourceInstanceSettings`
- Look for `json:"fieldName"` tags — these map to `jsonData` fields
- Look for `SecureJSONData` or `DecryptedSecureJSONData` usage — these map to `secureJsonData` fields

**Frontend (TypeScript):**
- `src/types.ts` or `src/types/` — TypeScript interfaces
- `src/components/ConfigEditor.tsx` — the provisioning UI; field names, labels, defaults, validation
- `src/datasource.ts` — how settings are read at runtime

**Documentation:**
- `docs/sources/configure.md` or `docs/sources/setup.md` — provisioning YAML examples
- `README.md` — sometimes has provisioning examples
- Look for provisioning example blocks showing `jsonData` and `secureJsonData`

### 3. Pin a Commit SHA

For link stability, find a recent commit SHA on the default branch. All source links in comments MUST use this format:
```
https://github.com/<owner>/<repo>/blob/<SHA>/<path>#L<start>-L<end>
```

Never use branch names (like `main`) in permalinks — they break when files change.

### 4. Catalog Every Configuration Field

For each field found in settings structs, ConfigEditor, and documentation, collect:

| Attribute | Source |
|-----------|--------|
| **Field name** | Go JSON tag or TypeScript property name |
| **Type** | Go type / TS type; create companion enum/type aliases for constrained values |
| **Source definition** | Permalink to the exact struct/interface/type where this field is declared (see below) |
| **Package origin** | If the field is inherited from an external package, permalink to that package's type definition |
| **Required / Optional / Conditionally required** | Backend validation, docs, UI `required` props |
| **Default value** | Backend defaults, UI defaults, docs |
| **Backend behavior** | What happens if missing/empty; validation errors; migration logic |
| **UI hints** | ConfigEditor label, placeholder, conditional visibility, help text |
| **Provisioning examples** | From docs or README |
| **Legacy / migration notes** | Field renames, fallback logic, deprecated alternatives |

#### 4a. Tracing Field Origins — Source Definition & Package Links

Many datasource config types extend or embed types from external packages. Every field MUST include:

1. **Source definition link**: A permalink to the file and line(s) where the field is declared in the datasource's own type definition.
   - For Go: the struct field with its `json:"..."` tag (e.g., `settings.go` or `types.go`).
   - For TypeScript: the interface property in `types.ts` or similar.

2. **Package origin link** (when applicable): If the field is inherited from a parent type in an external package, include a permalink to that package's type definition too.

Common package inheritance patterns:

| Datasource | Frontend type extends | Package repo |
|------------|----------------------|-------------|
| CloudWatch | `AwsAuthDataSourceJsonData` | [`@grafana/aws-sdk`](https://github.com/grafana/grafana-aws-sdk-react) |
| CloudWatch (backend) | `awsds.AWSDatasourceSettings` | [`grafana/grafana-aws-sdk`](https://github.com/grafana/grafana-aws-sdk) — `pkg/awsds/settings.go` |
| Azure Monitor | `AzureDataSourceJsonData` | [`@grafana/azure-sdk`](https://github.com/grafana/grafana-azure-sdk-react) |
| Azure Monitor (backend) | `azcredentials.AzureCredentials` | [`grafana/grafana-azure-sdk-go`](https://github.com/grafana/grafana-azure-sdk-go) |
| Google Cloud (Sheets, etc.) | Custom per-plugin | Plugin repo itself |

**How to trace a field's origin:**

1. Find the datasource's main settings type (e.g., `CloudWatchJsonData`, `AzureMonitorDataSourceJsonData`).
2. Check if it extends/embeds another type (e.g., `extends AwsAuthDataSourceJsonData`, `embeds awsds.AWSDatasourceSettings`).
3. For each field, determine whether it's declared directly on the datasource type or inherited from the parent.
4. Include both links when inherited.

**Example JSDoc pattern for inherited fields:**

```typescript
/**
 * REQUIRED.
 *
 * AWS authentication method.
 *
 * Source definition:
 * - Datasource type: CloudWatchJsonData extends AwsAuthDataSourceJsonData
 *   https://github.com/grafana/grafana/blob/<SHA>/public/app/plugins/datasource/cloudwatch/types.ts#L20-L40
 * - Inherited from AwsAuthDataSourceJsonData.authType in @grafana/aws-sdk:
 *   https://github.com/grafana/grafana-aws-sdk-react/blob/<SHA>/src/types.ts#L5-L15
 * - Backend struct: awsds.AWSDatasourceSettings.AuthType
 *   https://github.com/grafana/grafana-aws-sdk/blob/<SHA>/pkg/awsds/settings.go#L91-L100
 *
 * Backend behavior:
 * ...
 */
authType?: CloudWatchAuthType;
```

**Example JSDoc pattern for fields declared directly on the datasource type:**

```typescript
/**
 * OPTIONAL.
 *
 * Timeout for CloudWatch Logs queries.
 *
 * Source definition:
 * - Frontend: CloudWatchJsonData.logsTimeout
 *   https://github.com/grafana/grafana/blob/<SHA>/public/app/plugins/datasource/cloudwatch/types.ts#L28
 * - Backend: CloudWatchSettings.LogsTimeout
 *   https://github.com/grafana/grafana/blob/<SHA>/pkg/tsdb/cloudwatch/models/settings.go#L18-L20
 *
 * Backend behavior:
 * ...
 */
logsTimeout?: string;
```

### 5. Generate the TypeScript Type File

Follow this exact structure and conventions:

```typescript
export type <datasourceName>Config = {
  // Top-level fields (url, basicAuth, user, etc.) only if the datasource uses them

  jsonData: {
    /**
     * REQUIRED | OPTIONAL | CONDITIONALLY REQUIRED: <condition>.
     *
     * <One-line description of what this field does.>
     *
     * Values:
     * - "<value1>": <description>
     * - "<value2>": <description>
     *
     * Backend behavior:
     * - <What happens if missing/empty; error messages; validation logic>
     *   <permalink to relevant backend source>
     *
     * UI hints:
     * - <How ConfigEditor displays this field; conditional visibility; labels>
     *   <permalink to relevant frontend source>
     *
     * Defaults:
     * - <Default value and where it's set>
     *   <permalink if applicable>
     *
     * Provisioning examples:
     * <permalink to docs provisioning section>
     */
    fieldName?: FieldType;
  };

  secureJsonData: {
    /**
     * REQUIRED when <condition> | OPTIONAL.
     *
     * <Description.>
     *
     * <Source links as above.>
     */
    secretField?: string;
  };
};

// Companion types for constrained values
export type SomeEnumType = "value1" | "value2" | "value3";
```

### 6. Comment Quality Checklist

Every field comment MUST include:

- [ ] **Requirement level**: One of `REQUIRED`, `OPTIONAL`, or `CONDITIONALLY REQUIRED: <when>` as the first word
- [ ] **Description**: One-line summary of what the field controls
- [ ] **Source definition link(s)**: Permalink to the type/struct where this field is declared — both the datasource's own type AND the parent package type if inherited (see Step 4a)
- [ ] **Backend behavior**: What error or behavior occurs if the field is missing or invalid (with link)
- [ ] **UI hints** (if a ConfigEditor exists): How the field appears in the Grafana UI (with link)

Optional but encouraged:
- **Values**: Enumerate allowed values for string unions/enums
- **Defaults**: Note any default value
- **Provisioning examples**: Link to docs showing YAML provisioning
- **Legacy/migration**: Note if the field replaces or aliases another field
- **Package origin**: If the field comes from an external SDK/package, note the package name and link

### 7. Naming Conventions

- **File name**: `src/<camelCaseName>.ts` — e.g., `googleSheets.ts`, `cloudWatch.ts`, `azureMonitor.ts`
- **Main type**: `export type <camelCaseName>Config` — e.g., `googleSheetsConfig`, `cloudWatchConfig`
- **Companion types**: PascalCase, prefixed with datasource name — e.g., `GoogleSheetsAuthenticationType`, `CloudWatchRegion`
- Use `type` aliases for string unions (not `enum`) to match provisioning YAML values literally

### 8. Register the Export

Add the new type export to `src/types.ts`:
```typescript
export { <name>Config } from "./<name>";
```

### 9. Validate

After generating the file:
1. Run `npx tsc --noEmit` to check for type errors
2. Verify every permalink resolves (spot-check 2-3 links)
3. Ensure the field set covers what's in the ConfigEditor, backend settings struct, AND docs provisioning examples — the union of all three sources

## Reference: googleSheets.ts Pattern

The gold standard is `src/googleSheets.ts`. Key qualities to match:

1. **Permalink format**: `https://github.com/grafana/<repo>/blob/<SHA>/<path>#L<start>-L<end>` — always with commit SHA, never branch name
2. **Requirement prefix**: Every field starts with `REQUIRED`, `OPTIONAL`, or `CONDITIONALLY REQUIRED: for \`<condition>\``
3. **Multi-source evidence**: Comments cite backend Go code, frontend ConfigEditor, AND docs when all three mention the field
4. **Source definition links**: Every field links to the exact line in the type/struct/interface where it is declared
5. **Package origin links**: When a field is inherited from an external package (e.g., `@grafana/aws-sdk`, `grafana-azure-sdk-go`), the comment includes a link to both the datasource's type AND the external package's type definition
6. **Legacy handling**: Legacy/alias fields are documented with migration logic references
7. **Companion types**: String unions extracted as named type aliases at the bottom of the file

### Source Linking Examples from Existing Types

For a CloudWatch field inherited from `@grafana/aws-sdk`:
```
 * Source definition:
 * - Frontend: CloudWatchJsonData extends AwsAuthDataSourceJsonData
 *   https://github.com/grafana/grafana/blob/<SHA>/public/app/plugins/datasource/cloudwatch/types.ts#L20-L40
 * - Package: AwsAuthDataSourceJsonData in @grafana/aws-sdk
 *   https://github.com/grafana/grafana-aws-sdk-react/blob/<SHA>/src/types.ts#L5-L15
 * - Backend: awsds.AWSDatasourceSettings in grafana-aws-sdk
 *   https://github.com/grafana/grafana-aws-sdk/blob/<SHA>/pkg/awsds/settings.go#L91-L100
```

For a field declared directly on the datasource type:
```
 * Source definition:
 * - Frontend: CloudWatchJsonData.logsTimeout
 *   https://github.com/grafana/grafana/blob/<SHA>/public/app/plugins/datasource/cloudwatch/types.ts#L28
 * - Backend: CloudWatchSettings.LogsTimeout
 *   https://github.com/grafana/grafana/blob/<SHA>/pkg/tsdb/cloudwatch/models/settings.go#L18-L20
```

## Common Datasource Repos

| Datasource | Repository | Settings location |
|------------|-----------|-------------------|
| Prometheus | `grafana/grafana` | `pkg/tsdb/prometheus/models/`, `public/app/plugins/datasource/prometheus/` |
| Loki | `grafana/grafana` | `pkg/tsdb/loki/`, `public/app/plugins/datasource/loki/` |
| Tempo | `grafana/grafana` | `pkg/tsdb/tempo/`, `public/app/plugins/datasource/tempo/` |
| PostgreSQL | `grafana/grafana` | `pkg/tsdb/grafana-postgresql-datasource/`, `public/app/plugins/datasource/postgres/` |
| MySQL | `grafana/grafana` | `pkg/tsdb/mysql/`, `public/app/plugins/datasource/mysql/` |
| CloudWatch | `grafana/grafana` | `pkg/tsdb/cloudwatch/`, `public/app/plugins/datasource/cloudwatch/` |
| Elasticsearch | `grafana/grafana` | `pkg/tsdb/elasticsearch/`, `public/app/plugins/datasource/elasticsearch/` |
| Google Sheets | `grafana/google-sheets-datasource` | `pkg/models/settings.go`, `src/components/ConfigEditor.tsx` |
| GitHub | `grafana/github-datasource` | `pkg/github/datasource.go`, `src/components/ConfigEditor.tsx` |
| Infinity | `grafana/grafana-infinity-datasource` | `pkg/infinity/`, `src/` |
| PagerDuty | `grafana/pagerduty-datasource` | `pkg/pagerduty/`, `src/` |
| Splunk | `grafana/splunk-datasource` | `pkg/`, `src/` |
| Datadog | `grafana/datadog-datasource` | `pkg/`, `src/` |
| New Relic | `grafana/newrelic-datasource` | `pkg/`, `src/` |

## Tips

- **Private repos**: If the repo is private, use the `github_repo` tool to search for code. It can search any Grafana org repo.
- **Multiple settings structs**: Some datasources split settings across multiple structs (e.g., one for jsonData, one for secureJsonData). Merge them into the single output type.
- **Feature flags**: Some fields are gated behind feature flags. Note this in the comment if found.
- **Cross-datasource fields**: Fields like `timeout`, `tlsAuth`, `tlsSkipVerify`, `keepCookies`, `pdcInjected` are common across many datasources. Still document them with source links specific to this datasource's usage.
- **Tracing inherited fields**: Always follow the `extends`/`embeds` chain. For example, `CloudWatchJsonData extends AwsAuthDataSourceJsonData` — every field from `AwsAuthDataSourceJsonData` needs a link to both the CloudWatch types file and the aws-sdk package where it originates.
- **Go module dependencies**: Use `go.mod` to find the exact version of embedded packages (e.g., `grafana-aws-sdk`, `grafana-azure-sdk-go`). Search the package repo for the struct definition.
- **npm package dependencies**: Use `package.json` to find the exact version of frontend SDK packages (e.g., `@grafana/aws-sdk`, `@grafana/azure-sdk`). Search the package repo for the TypeScript interface.

## Common External Package Repos

| Package | Repository | Key types |
|---------|-----------|----------|
| `@grafana/aws-sdk` (frontend) | `grafana/grafana-aws-sdk-react` | `AwsAuthDataSourceJsonData`, `AwsAuthDataSourceSecureJsonData` |
| `grafana-aws-sdk` (Go) | `grafana/grafana-aws-sdk` | `awsds.AWSDatasourceSettings`, `awsds.AuthSettings`, `awsds.AuthType` |
| `@grafana/azure-sdk` (frontend) | `grafana/grafana-azure-sdk-react` | `AzureDataSourceJsonData`, `AzureDataSourceSecureJsonData`, `AzureCredentials` |
| `grafana-azure-sdk-go` (Go) | `grafana/grafana-azure-sdk-go` | `azcredentials.AzureCredentials`, `azsettings.AzureSettings` |
| `@grafana/google-sdk` (frontend) | `grafana/grafana-google-sdk-react` | `GoogleAuthType` |
| `grafana-google-sdk-go` (Go) | `grafana/grafana-google-sdk-go` | Google auth settings |
