# Datasource Configuration Schema

Declarative schema for Grafana datasource configuration.

## Root schema

| name          | type                | required  | description                                                                  |
| ------------- | ------------------- | --------- | ---------------------------------------------------------------------------- |
| schemaVersion | string              | Required. | Schema spec version (e.g. "v1").                                             |
| pluginType    | string              | Required. | Unique plugin identifier (e.g. "prometheus").                                |
| pluginName    | string              | Required. | Human-readable name.                                                         |
| docURL        | string              | Optional  | documentation URL.                                                           |
| baseFields    | BaseFieldRef[]      | Optional  | SDK field packs to merge before validation. See [Base fields](#base-fields). |
| fields        | ConfigField[]       | Required. | Source of truth for all config fields.                                       |
| groups        | ConfigGroup[]       | Optional  | UI layout grouping.                                                          |
| instructions  | Instruction[]       | Optional  | Instructions for LLMs and other consumers.                                   |
| relationships | FieldRelationship[] | Optional  | semantic relationships between fields.                                       |

## Base fields

SDK libraries such as [`grafana-plugin-sdk-go`](https://github.com/grafana/grafana-plugin-sdk-go), [`grafana-aws-sdk`](https://github.com/grafana/grafana-aws-sdk), [`grafana-azure-sdk-go`](https://github.com/grafana/grafana-azure-sdk-go), and [`grafana-google-sdk-go`](https://github.com/grafana/grafana-google-sdk-go) define a fixed set of well-known fields (URL, basicAuth, TLS, SigV4 auth, etc.). Without `baseFields`, every plugin that uses these SDKs must redeclare those fields verbatim — creating copy-paste drift and maintenance overhead.

`baseFields` solves this by letting a plugin _declare_ which SDK field packs it uses. Pack fields are merged into `fields` before validation; a plugin only declares what is genuinely its own.

> **Note:** All built-in packs (`plugin_sdk_settings`, `aws_sdk_settings`, `azure_sdk_settings`, `google_sdk_settings`, `sqleng_settings`) are now populated; `exclude` and `patch` may reference any top-level field ID defined under `fields` in the respective pack JSON under [`dsconfig/packs/`](packs/). Nested item-field IDs (for example, `plugin_sdk_settings.httpHeaders.item.name`) are not eligible.

```json
{
  "schemaVersion": "v1",
  "pluginType": "prometheus",
  "pluginName": "Prometheus",
  "baseFields": [
    {
      "from": "plugin_sdk_settings",
      "patch": {
        "plugin_sdk_settings.url": {
          "label": "Prometheus server URL",
          "placeholder": "http://localhost:9090"
        }
      }
    },
    {
      "from": "aws_sdk_settings",
      "exclude": ["aws_sdk_settings.assumeRoleArn"]
    }
  ],
  "fields": [
    {
      "id": "jsonData.httpMethod",
      "key": "httpMethod",
      "valueType": "string",
      "target": "jsonData"
    }
  ]
}
```

### `BaseFieldRef` properties

| property  | type                         | required | description                                                            |
| --------- | ---------------------------- | -------- | ---------------------------------------------------------------------- |
| `from`    | `FieldPackID` (string)       | Required | Identifier of the built-in field pack to include.                      |
| `exclude` | `string[]`                   | Optional | Field IDs from the pack to omit entirely. Each must exist in the pack. |
| `patch`   | `Record<string, FieldPatch>` | Optional | Presentation-only overrides keyed by field ID.                         |

### `FieldPatch` properties

`FieldPatch` allows customising _presentation_ properties of a pack field without redefining it. Structural properties (`id`, `key`, `valueType`, `target`, `role`) are intentionally absent — those are the pack's immutable contract.

| property       | type      | description                                                             |
| -------------- | --------- | ----------------------------------------------------------------------- |
| `label`        | `string`  | Override the field label.                                               |
| `description`  | `string`  | Override the field description / tooltip.                               |
| `placeholder`  | `string`  | Override the input placeholder text.                                    |
| `defaultValue` | `any`     | Override the field default value.                                       |
| `required`     | `boolean` | Override whether the field is required.                                 |
| `hidden`       | `boolean` | Reserved for future use (currently ignored by `baseFields` resolution). |

### Built-in packs

| `from` value          | Source SDK              | Content                                             |
| --------------------- | ----------------------- | --------------------------------------------------- |
| `plugin_sdk_settings` | `grafana-plugin-sdk-go` | Standard HTTP datasource fields shared by `backend/httpclient/options.go` (`Options`, `BasicAuthOptions`, `TLSOptions`, `TimeoutOptions`, `SigV4Config`, `Header`) and the `@grafana/ui` `DataSourceHttpSettings` editor (URL, access, basic auth, TLS, timeout, OAuth pass-through, SigV4 toggle, custom HTTP headers) — see [`plugin_sdk_settings.json`](packs/plugin_sdk_settings.json) |
| `aws_sdk_settings`    | `grafana-aws-sdk-go`    | AWS credentials and auth (auth provider, profile, access/secret keys, session token, assume-role ARN, external ID, endpoint, default region, proxy settings) — see [`aws_sdk_settings.json`](packs/aws_sdk_settings.json) |
| `azure_sdk_settings`  | `grafana-azure-sdk-go`  | Azure credentials and auth (App Registration client secret / client certificate, Managed Identity, Workload Identity, Current User, On-Behalf-Of, Entra Password) — see [`azure_sdk_settings.json`](packs/azure_sdk_settings.json) |
| `google_sdk_settings` | `grafana-google-sdk-go` | Google credentials and auth (JWT, GCE, WIF, Forward OAuth) — see [`google_sdk_settings.json`](packs/google_sdk_settings.json) |
| `sqleng_settings`     | `grafana sqleng`        | Common SQL datasource fields shared by grafana-postgresql-datasource, grafana-mysql-datasource, and grafana-mssql-datasource (host URL, database, username/password, TLS certs, TLS server name, connection pool tuning, Secure Socks Proxy). Vendor-specific options (postgres `sslmode`/`timescaledb`, mysql `tlsAuth`/`allowCleartextPasswords`, mssql `encrypt`/`authenticationType`) stay in the individual plugin schemas. mssql plugins should set `exclude: ["sqleng_settings.servername"]` and declare their own `serverName` field (camelCase), since mssql uses a different storage key than postgres/mysql for TLS server name — see [`sqleng_settings.json`](packs/sqleng_settings.json) |

Pack field definitions live in `dsconfig/packs/` as JSON files (e.g. `plugin_sdk_settings.json`), each validated against `dsconfig/packs/pack-schema.json`. After editing a pack JSON file, run `go generate ./...` from the `dsconfig` module to refresh the per-pack `exclude`/`patch` enums in `schema.json`.

### Resolution rules

1. `baseFields` is resolved by calling `ResolveBaseFields()` (or the convenience wrapper `ParseAndResolveSchemaJSON()`). **`Validate()` will return an error if called on an unresolved schema.**
2. Pack fields are prepended to `fields` in declaration order.
3. If a pack field ID collides with a plugin-declared field, the **plugin field wins** (explicit beats inherited).
4. `exclude` is validated: every excluded ID must exist in the pack.
5. `patch` is validated: every patched ID must exist in the pack and must not also appear in `exclude`.
6. Duplicate `from` values in the same `baseFields` array are an error.

## Field identity: `id` vs `key`

| Property | Purpose                    | Scope                                        | Example                   |
| -------- | -------------------------- | -------------------------------------------- | ------------------------- |
| `id`     | Canonical schema reference | Globally unique across the entire schema     | `"httpHeaders.item.name"` |
| `key`    | Storage/object key         | Local to its storage target or parent object | `"name"`                  |

Groups and relationships reference fields by `id`.

## Storage target

`target` specifies where the field is stored in Grafana's datasource config:

| Value            | Maps to                                     |
| ---------------- | ------------------------------------------- |
| `root`           | Top-level fields (`url`, `basicAuth`, etc.) |
| `jsonData`       | `jsonData.*`                                |
| `secureJsonData` | `secureJsonData.*` (write-only)             |

**Required** for storage fields. **Omitted** for virtual fields and item fields.

### Secure fields

Fields targeting `secureJsonData` are **write-only**. When reading existing config, consumers should use `secureJsonFields` (a `Record<string, boolean>`) to determine whether a secret is already configured. The schema describes the field; it does not imply the secret value is retrievable.

## Storage mapping

The `storage` property defines how logical fields map to Grafana's legacy storage format.

| Type          | Description                                                                    |
| ------------- | ------------------------------------------------------------------------------ |
| `direct`      | Default. `target` + `key` maps directly.                                       |
| `indexedPair` | Legacy indexed key/value pattern (e.g. `httpHeaderName1`, `httpHeaderValue1`). |
| `computed`    | Declarative read/write expressions. Execution is runtime-specific.             |

`computed` mappings store CEL-like expressions but are **not evaluated** by the schema validator.

## Validation rules

`validations[]` defines the **data contract**. `ui.options` defines **presentation**.

```json
{
  "validations": [{ "type": "allowedValues", "values": ["GET", "POST"] }],
  "ui": {
    "component": "select",
    "options": [
      { "label": "GET", "value": "GET" },
      { "label": "POST", "value": "POST" }
    ]
  }
}
```

Tools, docs generators, provisioning, and LLM integrations should use `validations[]` — not `ui.options` — as the source of truth for allowed values.

### Rule types

| Type            | Required fields    | Purpose                               |
| --------------- | ------------------ | ------------------------------------- |
| `pattern`       | `pattern`          | Regex validation for strings          |
| `range`         | `min` and/or `max` | Numeric bounds                        |
| `length`        | `min` and/or `max` | String length bounds                  |
| `itemCount`     | `min` and/or `max` | Array size bounds                     |
| `allowedValues` | `values`           | Enumerated allowed values             |
| `custom`        | `expression`       | CEL expression (evaluated at runtime) |

## Field help

Help uses the same language — Markdown — at two intensities, so there is no duplication:

| Source          | Markdown    | Scope                                | Editor surface |
| --------------- | ----------- | ------------------------------------ | -------------- |
| `description`   | inline only | docs, provisioning, LLM, **tooltip** | tooltip        |
| `help.markdown` | block       | editor-only rich help                | drawer         |

For **short help**, use the field's `description` — a one-liner in inline Markdown (emphasis, links,
code spans). Editors surface it as an accessible tooltip on the label. No extra schema is needed.
Plain text is valid Markdown, so existing descriptions keep working.

For **rich, custom help** that is too involved for a tooltip (multi-step instructions, links, or
code), add an optional `help` object whose `markdown` is block Markdown. Editors render it as a
drawer/side panel opened from the label. The presence of `help` is the signal to render a drawer;
don't restate the `description` here.

```json
{
  "id": "secure.bearerToken",
  "key": "bearerToken",
  "label": "Bearer token",
  "description": "Bearer token sent in the Authorization header.",
  "valueType": "string",
  "target": "secureJsonData",
  "help": {
    "title": "How to get a bearer token",
    "subtitle": "Generate a token from the Example developer console",
    "markdown": "1. Sign in to the developer console.\n2. Open **Settings → API access**.\n3. Select **Create token** and copy it.",
    "docURL": "https://example.com/docs/authentication"
  }
}
```

| Property   | Type   | Required | Description                                                   |
| ---------- | ------ | -------- | ------------------------------------------------------------- |
| `title`    | string | Optional | Drawer heading; also the trigger label that opens the drawer. |
| `subtitle` | string | Optional | Secondary drawer heading.                                     |
| `markdown` | string | Required | Help body in Markdown (multi-step instructions, links, code). |
| `docURL`   | string | Optional | Documentation link surfaced in the drawer.                    |

## Field roles

A field may carry an optional `role`: a tag from a **closed, versioned vocabulary** that says _what
the field means_, independent of what it is _named_. Two plugins may store a TLS client certificate
as `tlsClientCert` and `clientCertificate`; both can declare `"role": "tls.clientCert"` so a generic
consumer can find "the TLS client cert" without knowing either plugin's field names.

`role` serves two concrete goals:

- **Automated HTTP client construction** — a generic builder can locate the timeout, TLS cert, or
  basic-auth password across any plugin by role instead of hard-coding field names.
- **Assistant reasoning** — tools populating a never-before-seen config can recognize, for example,
  "the secret the user must paste," regardless of the author's field name.

`role` is **optional** and adoption is **incremental**: a field with no `role` behaves exactly as
before. The vocabulary is fixed in the package — an unknown value is rejected at validation, the same
discipline `valueType` and `target` use. If a field's meaning isn't represented yet, it simply gets
no `role`; growing the vocabulary is an additive change to the package, not something a schema
controls.

### Vocabulary

| Namespace       | Roles                                                                                                   |
| --------------- | ------------------------------------------------------------------------------------------------------- |
| `endpoint.*`    | `endpoint.baseUrl`, `endpoint.scheme`, `endpoint.domain`, `endpoint.port`                               |
| `transport.*`   | `transport.timeoutSeconds`, `transport.tlsSkipVerify`                                                   |
| `tls.*`         | `tls.clientCert`, `tls.clientKey`, `tls.caCert`, `tls.serverName`                                       |
| `auth.*`        | `auth.discriminator`                                                                                    |
|                 | `auth.basic.enabled`, `auth.basic.username`, `auth.basic.password`,`auth.basic.token`                   |
|                 | `auth.bearer.token`                                                                                     |
|                 | `auth.oauth2.clientId`, `auth.oauth2.clientSecret`, `auth.oauth2.tokenUrl`, `auth.oauth2.jwtPrivateKey` |
|                 | `auth.jwt.signingKey`,                                                                                  |
|                 | `auth.awsSigV4.enabled`, `auth.aws.accessKeyId`, `auth.aws.secretAccessKey`                             |
|                 | `auth.azureBlob.storageAccountKey`                                                                      |
|                 | `auth.forwardOAuthToken.enabled`                                                                        |
|                 | `auth.apiKey.key`, `auth.apiKey.value`                                                                  |
| `http.header.*` | `http.header` (the array field), `http.header.name`, `http.header.value` (item fields)                  |
| `http.query.*`  | `http.query` (the array field), `http.query.name`, `http.query.value` (item fields)                     |

An endpoint may be modeled as a single `endpoint.baseUrl`, or split into `endpoint.scheme`,
`endpoint.domain`, and `endpoint.port` for plugins that store the parts separately.

```json
{
  "id": "secure.tlsClientCert",
  "key": "clientCertificate",
  "valueType": "string",
  "target": "secureJsonData",
  "role": "tls.clientCert"
}
```

Compound fields use a **parent + item** pattern: the array field carries the collection role
(`http.header`) and its item sub-fields carry the element roles (`http.header.name`,
`http.header.value`):

```json
{
  "id": "jsonData.httpHeaders",
  "key": "httpHeaders",
  "valueType": "array",
  "target": "jsonData",
  "role": "http.header",
  "item": {
    "valueType": "object",
    "fields": [
      {
        "id": "httpHeaders.item.name",
        "key": "name",
        "valueType": "string",
        "isItemField": true,
        "role": "http.header.name"
      },
      {
        "id": "httpHeaders.item.value",
        "key": "value",
        "valueType": "string",
        "isItemField": true,
        "role": "http.header.value"
      }
    ]
  }
}
```

HTTP query-string parameters follow the same parent + item pattern with `http.query`,
`http.query.name`, and `http.query.value`.

A valid `role` does not yet guarantee an _appropriate_ one — nothing currently cross-checks, for
example, `tls.clientCert` against the field being a `string` in `secureJsonData`, or `http.header.name`
against the field being an item field under an `http.header` array. That cross-validation, and the
companion `roleConflicts` mechanism, are tracked as follow-up work.

## Map fields

When `valueType` is `"map"` it represents an object with dynamic string keys. Like arrays, maps require an `item` property that describes the value type:

```json
{
  "id": "jsonData.labels",
  "key": "labels",
  "valueType": "map",
  "target": "jsonData",
  "item": { "valueType": "string" }
}
```

For maps with structured values :

```json
{
  "id": "jsonData.customizedRoutes",
  "key": "customizedRoutes",
  "valueType": "map",
  "target": "jsonData",
  "item": {
    "valueType": "object",
    "fields": [
      {
        "id": "customizedRoutes.item.URL",
        "key": "URL",
        "valueType": "string",
        "isItemField": true
      },
      {
        "id": "customizedRoutes.item.Scopes",
        "key": "Scopes",
        "valueType": "array",
        "isItemField": true,
        "item": { "valueType": "string" }
      }
    ]
  }
}
```

Map keys are always strings (JSON constraint). The `item` schema describes the **values**.

## Any fields

When `valueType` is `"any"`, the field accepts multiple runtime types (e.g. `string | string[]`). Use sparingly — only for genuinely polymorphic fields where a single type cannot describe the data:

```json
{
  "id": "search.filters.item.value",
  "key": "value",
  "valueType": "any",
  "isItemField": true,
  "description": "Filter value. May be a single string or array of strings."
}
```

Fields with `valueType: "any"` do not require an `item` property and skip type-level validation. Consumers should document the expected shapes in the `description`.

## Array item fields

When `valueType` is `"array"`, the field must have an `item` property:

```json
{
  "valueType": "array",
  "item": {
    "valueType": "object",
    "fields": [
      {
        "id": "headers.item.name",
        "key": "name",
        "valueType": "string",
        "isItemField": true
      }
    ]
  }
}
```

- `item.fields` is only allowed when `item.valueType` is `"object"`
- Every field inside `item.fields` **must** have `isItemField: true`
- Item fields do not require `target` (they inherit storage from the parent)

## Virtual fields

Fields with `kind: "virtual"` are derived/computed and not stored directly:

```json
{
  "id": "derived.hasAuth",
  "key": "hasAuth",
  "valueType": "boolean",
  "kind": "virtual"
}
```

Virtual fields:

- Do not require `target`
- May have a `computed` storage mapping with `read`/`write` expressions
- Are useful for UI state that doesn't map 1:1 to storage

## Effects: virtual selector → multi-field writes

Many datasources have a **selector dropdown** (e.g. "Authentication method") that controls **multiple storage fields** simultaneously. The `effects` array provides a structured, machine-readable way to declare these side-effects without opaque CEL expressions.

```json
{
  "id": "auth.method",
  "kind": "virtual",
  "defaultValue": "no-auth",
  "validations": [
    {
      "type": "allowedValues",
      "values": ["no-auth", "basic-auth", "forward-oauth"]
    }
  ],
  "ui": {
    "component": "select",
    "options": [
      { "label": "No Authentication", "value": "no-auth" },
      { "label": "Basic authentication", "value": "basic-auth" },
      { "label": "Forward OAuth Identity", "value": "forward-oauth" }
    ]
  },
  "storage": {
    "type": "computed",
    "read": "root.basicAuth == true ? 'basic-auth' : (jsonData.oauthPassThru == true ? 'forward-oauth' : 'no-auth')"
  },
  "effects": [
    {
      "when": "value == 'no-auth'",
      "set": { "auth.basicAuth": false, "auth.oauthPassThru": false }
    },
    {
      "when": "value == 'basic-auth'",
      "set": { "auth.basicAuth": true, "auth.oauthPassThru": false }
    },
    {
      "when": "value == 'forward-oauth'",
      "set": { "auth.basicAuth": false, "auth.oauthPassThru": true }
    }
  ]
}
```

### Effect rules

| Property | Type                     | Required | Description                                                            |
| -------- | ------------------------ | -------- | ---------------------------------------------------------------------- |
| `when`   | string (CEL)             | Yes      | Condition. Use `value` to refer to the field's current value.          |
| `set`    | `Record<fieldId, value>` | Yes      | Maps field IDs to the values they should be set to. Must not be empty. |

### How effects work with other primitives

- **`storage.computed.read`**: Derives the virtual field's value when loading existing config.
- **`effects[].set`**: Declares what to write when the virtual field changes.
- **`dependsOn`**: On target storage fields, controls visibility (e.g. show username only when `auth.method == 'basic-auth'`).
- **`requiredWhen`**: On target storage fields, conditional validation.
- **`tags: ["managed-by:auth.method"]`**: Convention for tagging fields that are driven by a selector.

Effects keys (`set`) reference field **IDs**, consistent with groups and relationships. They are validated against the schema's field ID set.

## Modeling patterns

### Recursive types

TypeScript types that reference themselves (e.g. `AzureCredentials.serviceCredentials?: AzureCredentials`) should be **flattened** using `section` with dotted paths. In practice, recursion is always bounded to a known depth:

```json
[
  {
    "id": "auth.credentials.authType",
    "key": "authType",
    "target": "jsonData",
    "section": "azureCredentials",
    "valueType": "string"
  },
  {
    "id": "auth.credentials.tenantId",
    "key": "tenantId",
    "target": "jsonData",
    "section": "azureCredentials",
    "valueType": "string"
  },
  {
    "id": "auth.svcCreds.authType",
    "key": "authType",
    "target": "jsonData",
    "section": "azureCredentials.serviceCredentials",
    "valueType": "string"
  },
  {
    "id": "auth.svcCreds.tenantId",
    "key": "tenantId",
    "target": "jsonData",
    "section": "azureCredentials.serviceCredentials",
    "valueType": "string"
  }
]
```

### Per-item secure fields

Some datasources have arrays where individual items may be secrets (e.g. Snowflake settings with a `secure: boolean` flag). Model the `secure` flag as a regular boolean item field and use a `computed` storage mapping to express the split:

```json
{
  "id": "jsonData.settings",
  "key": "settings",
  "valueType": "array",
  "target": "jsonData",
  "item": {
    "valueType": "object",
    "fields": [
      {
        "id": "settings.item.name",
        "key": "name",
        "valueType": "string",
        "isItemField": true
      },
      {
        "id": "settings.item.value",
        "key": "value",
        "valueType": "string",
        "isItemField": true
      },
      {
        "id": "settings.item.secure",
        "key": "secure",
        "valueType": "boolean",
        "isItemField": true
      }
    ]
  },
  "storage": {
    "type": "computed",
    "write": "splitByField(settings, 'secure', jsonData.settings, secureJsonData.settings)"
  }
}
```

## Groups and relationships

**Groups** define UI layout sections. They reference fields by `id`.
Set `"optional": true` on groups that can be collapsed or hidden by default (e.g. advanced sections).
The optional `ui` object holds presentation hints. Its `icon` is a **Grafana** icon name (a Grafana
`IconName` such as `plug`, `lock`, or `shield`) that a Grafana editor may render next to the group in
a settings sidebar. It is a Grafana-specific hint only — non-Grafana consumers should ignore it, so
it is always safe to omit or leave unhandled:

```json
{
  "id": "auth",
  "title": "Authentication",
  "description": "How to prove who you are to the API.",
  "ui": { "icon": "lock" },
  "fieldRefs": ["auth.user", "auth.password"]
}
```

| Property      | Type     | Required | Description                               |
| ------------- | -------- | -------- | ----------------------------------------- |
| `id`          | string   | Required | Unique group identifier.                  |
| `title`       | string   | Required | Human-readable section title.             |
| `fieldRefs`   | string[] | Required | Field `id`s shown in this group.          |
| `description` | string   | Optional | One-line section description.             |
| `ui`          | GroupUI  | Optional | Presentation hints (see below).           |
| `order`       | number   | Optional | Explicit ordering hint.                   |
| `optional`    | boolean  | Optional | Group can be collapsed/hidden by default. |

**GroupUI**

| Property | Type   | Required | Description                                                               |
| -------- | ------ | -------- | ------------------------------------------------------------------------- |
| `icon`   | string | Optional | Grafana `IconName` for a Grafana editor; non-Grafana consumers ignore it. |

**Relationships** define semantic connections between fields:

```json
{
  "type": "pair",
  "fields": ["auth.user", "auth.password"],
  "description": "Credentials"
}
```

Groups and relationships are metadata — they do not affect storage or validation.

## Expression language

Expression fields (`dependsOn`, `requiredWhen`, `disabledWhen`, `overrides[].when`, `storage.computed.read/write`, `custom` validation `expression`) are **opaque strings** in v1. The intended language is CEL. This PR stores expressions but **does not evaluate them**. Runtime evaluation is follow-up work.

## Contract decisions

| Topic                         | Decision                             |
| ----------------------------- | ------------------------------------ |
| Existing Grafana config shape | Not changed                          |
| `id`                          | Canonical globally unique reference  |
| `key`                         | Local storage/object key             |
| `target`                      | root / jsonData / secureJsonData     |
| `storage`                     | Optional mapping strategy            |
| `validations[]`               | Data contract                        |
| `ui.options`                  | Presentation only                    |
| Secure fields                 | Values are write-only                |
| Expressions                   | Stored as strings, evaluated later   |
| Groups                        | Layout metadata, not source of truth |
| Relationships                 | Semantic metadata, not storage       |
