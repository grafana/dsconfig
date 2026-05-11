# Datasource Configuration Schema

## Overview

`dsconfig/schema` defines a declarative schema for Grafana datasource configuration. It describes every configurable field — its type, storage location, validation rules, UI hints, and relationships — in a single, language-neutral model.

The schema acts as a shared contract consumed by:

- **Config editors** — generate forms from schema instead of hand-writing React components
- **Validation** — enforce data contracts at provisioning time and in the UI
- **Documentation** — auto-generate field reference docs from schema definitions
- **LLM / automation tooling** — provide structured field metadata for AI-assisted configuration
- **Provisioning** — describe the exact shape of `jsonData`, `secureJsonData`, and root fields

The schema does **not** change Grafana's existing datasource config model. Fields still live in `root`, `jsonData`, and `secureJsonData` — the schema is a declarative layer on top of that structure.

Three implementations describe the same model:

| Language    | File                                    | Purpose                           |
| ----------- | --------------------------------------- | --------------------------------- |
| Go          | `schema.go`                             | Types, validation, SDK conversion |
| TypeScript  | `schema.ts`, `guards.ts`, `validate.ts` | Types, runtime guards, validation |
| JSON Schema | `schema.json`                           | External tooling, CI validation   |

See [`schema.md`](./schema.md) for the full design document including field reference tables, storage mapping details, and contract decisions.

---

## Quick Start

Every schema requires `schemaVersion`, `pluginType`, `pluginName`, and at least one field. Each field needs `id`, `key`, `valueType`, and `target` (for storage fields):

```json
{
  "schemaVersion": "v1",
  "pluginType": "my-datasource",
  "pluginName": "My Datasource",
  "fields": [
    {
      "id": "connection.url",
      "key": "url",
      "valueType": "string",
      "target": "root"
    }
  ]
}
```

**Validate in Go:**

```go
var s schema.DatasourceConfigSchema
json.Unmarshal(data, &s)
if err := s.Validate(); err != nil { /* ... */ }
```

**Validate in TypeScript:**

```ts
import { validateSchema } from "@grafana/dsconfig/schema";
const errors = validateSchema(JSON.parse(data));
```

**Validate against JSON Schema:**

```bash
npx ajv validate -s schema/schema.json -d my-datasource.schema.json
```

---

## How the Three Views Relate

Every example below shows three representations of the same datasource config:

1. **Schema** — the `dsconfig` schema definition (source of truth)
2. **Grafana Storage** — what gets persisted in Grafana's datasource config model
3. **SDK PluginSettings** — the OpenAPI spec produced by `ToPluginSettings()` for `grafana-plugin-sdk-go`

```
┌───────────-───┐     ToPluginSettings()     ┌───────────-----───────┐
│   dsconfig    │ ─────────────────────────► │  SDK PluginSettings   │
│   schema      │                            │  (spec + secureValues)│
└──────┬───-────┘                            └───────────────-----───┘
       │
       │  describes
       ▼
┌────────────-------------------──┐
│  Grafana                        │
│  Storage                        │
│  (root/jsonData/secureJsonData) │
└────────-------------------──────┘
```

---

## Concepts by Example

### 1. Root-level field

The simplest case: a single field stored at the top level of the datasource config.

**Schema:**

```json
{
  "schemaVersion": "v1",
  "pluginType": "example-simple",
  "pluginName": "Simple URL Datasource",
  "fields": [
    {
      "id": "connection.url",
      "key": "url",
      "description": "Base URL of the datasource",
      "valueType": "string",
      "semanticType": "url",
      "target": "root",
      "required": true,
      "validations": [
        {
          "type": "pattern",
          "pattern": "^https?://",
          "message": "Must be HTTP(S)"
        }
      ],
      "ui": { "component": "input", "placeholder": "https://example.com/api" }
    }
  ]
}
```

**Grafana Storage:**

```json
{
  "name": "My Simple DS",
  "type": "example-simple",
  "url": "https://example.com/api",
  "jsonData": {},
  "secureJsonData": {}
}
```

**SDK PluginSettings:**

```json
{
  "spec": {
    "type": "object",
    "required": ["url"],
    "properties": {
      "url": {
        "description": "Base URL of the datasource",
        "type": "string",
        "format": "uri",
        "pattern": "^https?://"
      }
    }
  }
}
```

Key points:

- `target: "root"` → the field sits at the top level in Grafana storage
- `semanticType: "url"` → becomes `"format": "uri"` in the SDK spec
- `validations[].pattern` → becomes `"pattern"` in JSON Schema
- No `secureValues` since nothing targets `secureJsonData`

---

### 2. Fields across all three targets

A typical datasource has fields in `root` (url, basicAuth), `jsonData` (settings), and `secureJsonData` (secrets).

**Schema** (abbreviated — full version in [`testdata/convert/root-jsondata-secure-mix/`](./testdata/convert/root-jsondata-secure-mix/)):

```json
{
  "schemaVersion": "v1",
  "pluginType": "example-mix",
  "pluginName": "Mixed Targets Datasource",
  "fields": [
    {
      "id": "connection.url",
      "key": "url",
      "valueType": "string",
      "semanticType": "url",
      "target": "root",
      "required": true
    },
    {
      "id": "connection.basicAuth",
      "key": "basicAuth",
      "valueType": "boolean",
      "target": "root"
    },
    {
      "id": "jsonData.timeout",
      "key": "timeout",
      "description": "Request timeout in seconds",
      "valueType": "number",
      "target": "jsonData",
      "defaultValue": 30,
      "validations": [{ "type": "range", "min": 1, "max": 300 }]
    },
    {
      "id": "jsonData.serverName",
      "key": "serverName",
      "valueType": "string",
      "semanticType": "hostname",
      "target": "jsonData"
    },
    {
      "id": "secure.basicAuthPassword",
      "key": "basicAuthPassword",
      "valueType": "string",
      "semanticType": "password",
      "target": "secureJsonData",
      "dependsOn": "connection.basicAuth == true"
    },
    {
      "id": "secure.tlsCACert",
      "key": "tlsCACert",
      "valueType": "string",
      "target": "secureJsonData"
    }
  ]
}
```

**Grafana Storage:**

```json
{
  "name": "My Mixed DS",
  "type": "example-mix",
  "url": "https://api.example.com",
  "basicAuth": true,
  "basicAuthUser": "your_username",
  "jsonData": {
    "tlsSkipVerify": false,
    "timeout": 60,
    "serverName": "api.example.com"
  },
  "secureJsonData": {
    "basicAuthPassword": "your_password",
    "tlsCACert": "-----BEGIN CERTIFICATE-----\n..."
  }
}
```

**SDK PluginSettings:**

```json
{
  "spec": {
    "type": "object",
    "required": ["url"],
    "properties": {
      "url": {
        "description": "Server URL",
        "type": "string",
        "format": "uri"
      },
      "basicAuth": {
        "type": "boolean"
      },
      "jsonData": {
        "type": "object",
        "properties": {
          "timeout": {
            "type": "number",
            "default": 30,
            "minimum": 1,
            "maximum": 300
          },
          "serverName": {
            "type": "string",
            "format": "hostname"
          }
        }
      }
    }
  },
  "secureValues": [
    {
      "key": "basicAuthPassword",
      "x-dsconfig-depends-on": "connection.basicAuth == true"
    },
    { "key": "tlsCACert" }
  ]
}
```

Key points:

- `target: "root"` fields become top-level properties in the SDK spec
- `target: "jsonData"` fields are nested under a `jsonData` object
- `target: "secureJsonData"` fields become `secureValues` entries — values are **write-only** and never returned in the spec
- `dependsOn` expressions become `x-dsconfig-depends-on` vendor extensions
- `validations[].range` maps to `minimum`/`maximum` in JSON Schema

---

### 3. Conditional auth with secrets

Auth methods often involve a selector that conditionally shows/requires a secret field.

**Schema** (full version in [`testdata/convert/bearer-token-auth/`](./testdata/convert/bearer-token-auth/)):

```json
{
  "schemaVersion": "v1",
  "pluginType": "example-bearer",
  "pluginName": "Bearer Token Datasource",
  "fields": [
    {
      "id": "connection.url",
      "key": "url",
      "valueType": "string",
      "semanticType": "url",
      "target": "root",
      "required": true
    },
    {
      "id": "auth.method",
      "key": "authMethod",
      "valueType": "string",
      "target": "jsonData",
      "defaultValue": "no-auth",
      "validations": [
        { "type": "allowedValues", "values": ["no-auth", "bearer-token"] }
      ],
      "ui": {
        "component": "select",
        "options": [
          { "label": "No Auth", "value": "no-auth" },
          { "label": "Bearer Token", "value": "bearer-token" }
        ]
      }
    },
    {
      "id": "auth.bearerToken",
      "key": "bearerToken",
      "valueType": "string",
      "semanticType": "token",
      "target": "secureJsonData",
      "dependsOn": "auth.method == 'bearer-token'",
      "requiredWhen": "auth.method == 'bearer-token'"
    }
  ]
}
```

**Grafana Storage:**

```json
{
  "name": "My Bearer DS",
  "type": "example-bearer",
  "url": "https://api.example.com",
  "jsonData": {
    "authMethod": "bearer-token"
  },
  "secureJsonData": {
    "bearerToken": "sk-secret-token-value"
  }
}
```

**SDK PluginSettings:**

```json
{
  "spec": {
    "type": "object",
    "required": ["url"],
    "properties": {
      "url": { "type": "string", "format": "uri" },
      "jsonData": {
        "type": "object",
        "properties": {
          "authMethod": {
            "type": "string",
            "default": "no-auth",
            "enum": ["no-auth", "bearer-token"]
          }
        }
      }
    }
  },
  "secureValues": [
    {
      "key": "bearerToken",
      "x-dsconfig-depends-on": "auth.method == 'bearer-token'",
      "x-dsconfig-required-when": "auth.method == 'bearer-token'"
    }
  ]
}
```

Key points:

- `allowedValues` validation → `enum` in JSON Schema
- `ui.options` is for **presentation** (labels); `validations[].values` is the **data contract**
- `dependsOn` / `requiredWhen` are CEL expressions — stored as opaque strings, not evaluated by the schema validator

---

### 4. Repeatable indexed pairs (HTTP headers)

Grafana's legacy storage for HTTP headers uses indexed key/value pairs (`httpHeaderName1`, `httpHeaderValue1`, etc.). The schema models this as an array with an `indexedPair` storage mapping.

**Schema** (full version in [`testdata/convert/indexed-headers-storage/`](./testdata/convert/indexed-headers-storage/)):

```json
{
  "schemaVersion": "v1",
  "pluginType": "example-headers",
  "pluginName": "HTTP Headers Datasource",
  "fields": [
    {
      "id": "connection.url",
      "key": "url",
      "valueType": "string",
      "semanticType": "url",
      "target": "root",
      "required": true
    },
    {
      "id": "httpHeaders",
      "key": "httpHeaders",
      "description": "Additional headers sent with every request",
      "valueType": "array",
      "target": "jsonData",
      "item": {
        "valueType": "object",
        "fields": [
          {
            "id": "httpHeaders.item.name",
            "key": "name",
            "valueType": "string",
            "isItemField": true,
            "required": true,
            "validations": [
              { "type": "pattern", "pattern": "^[A-Za-z][A-Za-z0-9-]*$" }
            ]
          },
          {
            "id": "httpHeaders.item.value",
            "key": "value",
            "valueType": "string",
            "isItemField": true
          }
        ]
      },
      "storage": {
        "type": "indexedPair",
        "key": { "target": "jsonData", "pattern": "httpHeaderName{index}" },
        "value": {
          "target": "secureJsonData",
          "pattern": "httpHeaderValue{index}"
        },
        "startIndex": 1
      },
      "validations": [
        {
          "type": "itemCount",
          "max": 10,
          "message": "Maximum 10 custom headers"
        }
      ]
    }
  ]
}
```

**Grafana Storage** (the legacy indexed format):

```json
{
  "name": "My Headers DS",
  "type": "example-headers",
  "url": "https://api.example.com",
  "jsonData": {
    "httpHeaderName1": "X-Custom-Header",
    "httpHeaderName2": "X-API-Token"
  },
  "secureJsonData": {
    "httpHeaderValue1": "custom-value",
    "httpHeaderValue2": "your-api-token"
  }
}
```

**SDK PluginSettings:**

```json
{
  "spec": {
    "type": "object",
    "required": ["url"],
    "properties": {
      "url": { "type": "string", "format": "uri" },
      "jsonData": {
        "type": "object",
        "properties": {
          "httpHeaders": {
            "type": "array",
            "maxItems": 10,
            "items": {
              "type": "object",
              "required": ["name"],
              "properties": {
                "name": {
                  "type": "string",
                  "pattern": "^[A-Za-z][A-Za-z0-9-]*$"
                },
                "value": { "type": "string" }
              }
            }
          }
        }
      }
    }
  }
}
```

Key points:

- The schema models headers as a clean `array` of objects
- `storage.type: "indexedPair"` describes how the array maps to Grafana's `httpHeaderName{index}` / `httpHeaderValue{index}` legacy pattern
- Header names go to `jsonData`, values go to `secureJsonData` — split across targets
- `isItemField: true` is required on all fields inside `item.fields` — they don't need their own `target`
- `itemCount` validation limits the array size
- The SDK spec sees a clean array; the storage mapping is a separate concern

---

### 5. Array of objects (no legacy mapping)

When the storage format is already a JSON array (e.g. Loki derived fields), no `storage` mapping is needed.

**Schema** (full version in [`testdata/convert/nested-object-jsondata/`](./testdata/convert/nested-object-jsondata/)):

```json
{
  "schemaVersion": "v1",
  "pluginType": "example-nested",
  "pluginName": "Nested Object Datasource",
  "fields": [
    {
      "id": "connection.url",
      "key": "url",
      "valueType": "string",
      "target": "root",
      "required": true
    },
    {
      "id": "jsonData.derivedFields",
      "key": "derivedFields",
      "valueType": "array",
      "target": "jsonData",
      "item": {
        "valueType": "object",
        "fields": [
          {
            "id": "derivedFields.item.name",
            "key": "name",
            "valueType": "string",
            "isItemField": true,
            "required": true
          },
          {
            "id": "derivedFields.item.matcherRegex",
            "key": "matcherRegex",
            "valueType": "string",
            "isItemField": true,
            "required": true
          },
          {
            "id": "derivedFields.item.url",
            "key": "url",
            "valueType": "string",
            "semanticType": "url",
            "isItemField": true
          },
          {
            "id": "derivedFields.item.datasourceUid",
            "key": "datasourceUid",
            "valueType": "string",
            "isItemField": true
          }
        ]
      }
    }
  ]
}
```

**Grafana Storage** (stored as-is in jsonData):

```json
{
  "name": "My Loki DS",
  "type": "example-nested",
  "url": "https://loki.example.com",
  "jsonData": {
    "derivedFields": [
      {
        "name": "TraceID",
        "matcherRegex": "traceID=(\\w+)",
        "url": "https://tempo.example.com/trace/${__value.raw}",
        "datasourceUid": "tempo-uid-123"
      }
    ]
  }
}
```

**SDK PluginSettings:**

```json
{
  "spec": {
    "type": "object",
    "required": ["url"],
    "properties": {
      "url": { "type": "string", "format": "uri" },
      "jsonData": {
        "type": "object",
        "properties": {
          "derivedFields": {
            "type": "array",
            "items": {
              "type": "object",
              "required": ["name", "matcherRegex"],
              "properties": {
                "name": { "type": "string" },
                "matcherRegex": {
                  "type": "string",
                  "minLength": 1,
                  "maxLength": 500
                },
                "url": { "type": "string", "format": "uri" },
                "datasourceUid": { "type": "string" }
              }
            }
          }
        }
      }
    }
  }
}
```

Key points:

- No `storage` mapping needed — the array is stored directly in `jsonData.derivedFields`
- Contrast with the indexed headers example where `storage: { type: "indexedPair" }` is required for the legacy format
- Item fields with `required: true` become `required` in the SDK spec's `items` schema

---

### 6. Virtual (computed) fields

Virtual fields are not stored. They derive a value from other fields for UI logic or tooling.

**Schema** (full version in [`testdata/convert/virtual-fields-skipped/`](./testdata/convert/virtual-fields-skipped/)):

```json
{
  "id": "derived.authConfigured",
  "key": "authConfigured",
  "description": "True when basic auth user and password are both set",
  "valueType": "boolean",
  "kind": "virtual",
  "lifecycle": "experimental",
  "storage": {
    "type": "computed",
    "read": "auth.basicAuth == true && auth.basicAuthUser != ''"
  }
}
```

**Grafana Storage:** _Not present._ Virtual fields have no storage representation.

**SDK PluginSettings:** _Not present._ `ToPluginSettings()` skips virtual fields entirely.

Key points:

- `kind: "virtual"` means no `target` is needed
- The `computed` storage mapping's `read` expression defines how to derive the value
- Virtual fields are useful for UI toggles ("show advanced"), computed summaries, or conditional logic anchors

---

## Field Requirements Summary

| Property      | Required?                   | Notes                                                                                |
| ------------- | --------------------------- | ------------------------------------------------------------------------------------ |
| `id`          | Always                      | Globally unique (e.g. `"auth.password"`)                                             |
| `key`         | Always                      | Local storage key (e.g. `"password"`)                                                |
| `valueType`   | Always                      | `string`, `number`, `boolean`, `array`, `object`                                     |
| `target`      | For storage fields          | `root`, `jsonData`, or `secureJsonData`. Omit for `virtual` and `isItemField` fields |
| `item`        | When `valueType` is `array` | Defines the array element schema                                                     |
| `isItemField` | Inside `item.fields`        | Must be `true` for all nested item fields                                            |

---

## Examples

See [`examples/`](./examples/) for complete, validated schema files:

| File                           | Pattern                                     |
| ------------------------------ | ------------------------------------------- |
| `simple-url.schema.json`       | Minimal — single URL field                  |
| `bearer-token.schema.json`     | Conditional auth select + secret            |
| `indexed-headers.schema.json`  | Array with `indexedPair` storage mapping    |
| `virtual-auth.schema.json`     | Basic auth with virtual computed field      |
| `array-of-objects.schema.json` | Array of objects (trace-to-metrics queries) |

See [`testdata/convert/`](./testdata/convert/) for input→output pairs showing the full schema-to-SDK-to-storage relationship for 18 scenarios including real datasource schemas (Prometheus, Loki, Tempo).

---

## Testing

```bash
# Go — schema validation, converter, round-trip
go test ./schema/ -count=1

# TypeScript — validation, guards, JSON Schema, examples
yarn test

# Both test suites validate the same example files,
# ensuring Go and TypeScript agree on what's valid.
```

---

## Known Gaps and Limitations

### Gaps (defined in schema, not yet implemented)

| #   | Gap                                                | Detail                                                                                                                                                                                                                               |
| --- | -------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| 1   | **CEL expressions are opaque strings**             | `dependsOn`, `requiredWhen`, `disabledWhen`, `overrides[].when`, `storage.computed.read/write`, and `custom` validation `expression` are stored but never parsed or evaluated. A typo won't be caught until a runtime engine exists. |
| 2   | **Storage mapping is metadata-only**               | `storage` (`direct`, `indexedPair`, `computed`) is validated structurally but has no runtime engine. No code reads an `indexedPair` mapping and expands `httpHeaderName{index}` into real keys.                                      |
| 3   | **No runtime value validation**                    | The schema defines validation rules (pattern, range, allowedValues, etc.) but there's no function that takes a schema + a config payload and returns validation errors against real data.                                            |
| 4   | **No React UI renderer**                           | `ui` hints (component, options, placeholder, width) are defined but nothing consumes them to generate a form.                                                                                                                        |
| 5   | **`FormState` / `SecureFieldState` are type-only** | These TypeScript types exist in `schema.ts` but nothing constructs, manages, or serializes them.                                                                                                                                     |

### Limitations (known constraints in the current design)

| #   | Limitation                                        | Detail                                                                                                                                                                                                                              |
| --- | ------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | **SDK converter ignores `storage` mappings**      | `ToPluginSettings()` routes fields by `target` only. For `indexedPair` arrays where item values are stored in `secureJsonData`, the SDK output shows a clean array under `jsonData` with **no indication that values are secrets**. |
| 2   | **SDK converter only processes top-level fields** | Item fields inside `item.fields` are nested via `itemSchemaToSpec()`. An item field can't independently become a `secureValue`.                                                                                                     |
| 3   | **`section` is undocumented**                     | The `Section` field (for nested jsonData paths like `tracesToLogs.datasourceUid`) is implemented in Go and the converter but not mentioned in `schema.md`, `schema.json`, or the TypeScript validator.                              |
| 4   | **`section` only supports one nesting level**     | `placeInSection()` creates a single sub-object. Deeply nested paths like `jsonData.a.b.c` aren't supported.                                                                                                                         |
| 5   | **`repeatable` / `pattern` fields are vestigial** | `ConfigField` has `repeatable` and `pattern` but they're never validated, never used by the converter, and overlap with `storage.indexedPair`.                                                                                      |
| 6   | **No cross-field validation**                     | All validation rules operate on a single field. "Field A must be less than field B" requires a `custom` CEL expression (which isn't evaluated).                                                                                     |
| 7   | **`overrides` don't affect SDK conversion**       | Field overrides (conditional defaults, descriptions, validations) are stored but `ToPluginSettings()` ignores them.                                                                                                                 |
| 8   | **`tags` and `examples` are unused**              | These metadata fields exist on `ConfigField` but nothing reads them.                                                                                                                                                                |
| 9   | **Go validation returns first error only**        | Go `Validate()` stops at the first error. TypeScript `validateSchema()` collects all errors. Behavior differs across languages.                                                                                                     |
| 10  | **No schema version migration**                   | `schemaVersion` is required but there's no code to handle version differences or backwards compatibility.                                                                                                                           |

### Trade-offs (deliberate decisions with downsides)

| #   | Trade-off                                            | Downside                                                                                                                                                                          |
| --- | ---------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | **`id` + `key` duality**                             | Every field needs two identifiers. No convention is enforced for `id` format — dot-separated paths are recommended but not validated.                                             |
| 2   | **`validations[]` vs `ui.options` separation**       | Select field allowed values may need to be maintained in two places. The converter's `applyUIEnum()` fallback (auto-derive `enum` from UI options) blurs the intended separation. |
| 3   | **Per-field `target`**                               | Flexible but verbose. You can't tell "what's in jsonData" at a glance without scanning all fields.                                                                                |
| 4   | **No first-class object field for jsonData nesting** | jsonData sub-objects (like `tracesToLogs`) use the implicit `section` mechanism rather than a dedicated `valueType: "object"` field with its own schema.                          |
| 5   | **JSON Schema `oneOf` for validation rules**         | Correct but produces verbose AJV error messages — every non-matching variant reports an error.                                                                                    |

### Open Questions

| #   | Question                                                                                                                   |
| --- | -------------------------------------------------------------------------------------------------------------------------- |
| 1   | Should `storage` mapping affect SDK conversion? (e.g. should `indexedPair` value patterns produce `secureValues` entries?) |
| 2   | How do `section` and `storage` interact when both are set on a field?                                                      |
| 3   | Should `repeatable` / `pattern` be removed in favor of `storage.indexedPair`?                                              |
| 4   | Should item fields support per-field `target` overrides for split-target patterns?                                         |
| 5   | How should `secureJsonFields` (read-side boolean map) be modeled in the schema or SDK output?                              |
| 6   | Should validation rule evaluation order, short-circuit behavior, or severity levels be defined?                            |
| 7   | Should groups enforce completeness? (every field must belong to at least one group)                                        |
| 8   | Should the TypeScript package be published as a standalone npm package?                                                    |
