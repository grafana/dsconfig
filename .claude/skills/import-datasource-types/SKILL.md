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
| **Required / Optional / Conditionally required** | Backend validation, docs, UI `required` props |
| **Default value** | Backend defaults, UI defaults, docs |
| **Backend behavior** | What happens if missing/empty; validation errors; migration logic |
| **UI hints** | ConfigEditor label, placeholder, conditional visibility, help text |
| **Provisioning examples** | From docs or README |
| **Legacy / migration notes** | Field renames, fallback logic, deprecated alternatives |

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
- [ ] **Source permalink**: At least one link to where the field is defined or validated in the backend
- [ ] **Backend behavior**: What error or behavior occurs if the field is missing or invalid (with link)
- [ ] **UI hints** (if a ConfigEditor exists): How the field appears in the Grafana UI (with link)

Optional but encouraged:
- **Values**: Enumerate allowed values for string unions/enums
- **Defaults**: Note any default value
- **Provisioning examples**: Link to docs showing YAML provisioning
- **Legacy/migration**: Note if the field replaces or aliases another field

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
4. **Legacy handling**: Legacy/alias fields are documented with migration logic references
5. **Companion types**: String unions extracted as named type aliases at the bottom of the file

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
