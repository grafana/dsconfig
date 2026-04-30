# Datasource Configuration Schema

Declarative schema for Grafana datasource configuration.

## Goals

- Single source of truth for datasource config: UI, validation, storage mapping, docs, and LLM tooling
- Language-neutral contract: Go, TypeScript, and JSON Schema all describe the same model
- Support the existing Grafana datasource config shape without changing it

## Non-goals (this PR)

- Runtime config value validation (follow-up)
- CEL expression evaluation (follow-up)
- React UI rendering (follow-up)
- Storage mapping engine (follow-up)

## Root schema

| name          | type                | required  | description                                   |
| ------------- | ------------------- | --------- | --------------------------------------------- |
| schemaVersion | string              | Required. | Schema spec version (e.g. "v1").              |
| pluginType    | string              | Required. | Unique plugin identifier (e.g. "prometheus"). |
| pluginName    | string              | Required. | Human-readable name.                          |
| docURL        | string              | Optional  | documentation URL.                            |
| fields        | ConfigField[]       | Required. | Source of truth for all config fields.        |
| groups        | ConfigGroup[]       | Optional  | UI layout grouping.                           |
| relationships | FieldRelationship[] | Optional  | semantic relationships between fields.        |

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

## Groups and relationships

**Groups** define UI layout sections. They reference fields by `id`:

```json
{
  "id": "auth",
  "title": "Authentication",
  "fieldRefs": ["auth.user", "auth.password"]
}
```

**Relationships** define semantic connections between fields:

```json
{
  "type": "pair",
  "fields": ["auth.user", "auth.password"],
  "description": "Credentials"
}
```

Groups and relationships are metadata — they do not affect storage or validation.

## Lifecycle

Fields can be marked with a lifecycle stage:

| Value          | Meaning                             |
| -------------- | ----------------------------------- |
| `stable`       | Production-ready                    |
| `deprecated`   | Will be removed in a future version |
| `experimental` | Subject to change                   |

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

## Examples

See [`examples/`](./examples/) for copy-pasteable schema examples:

- `simple-url.schema.json` — Minimal URL field
- `bearer-token.schema.json` — Auth method select + secure token
- `indexed-headers.schema.json` — HTTP headers with indexedPair mapping
- `virtual-auth.schema.json` — Basic auth with virtual computed field
- `array-of-objects.schema.json` — Array of trace-to-metrics queries
