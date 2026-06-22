# Datasource Configuration Schema

Declarative schema for Grafana datasource configuration.

## Purpose and scope

dsconfig is a semantic description layer placed on top of Grafana's existing datasource configuration model (root-level datasource properties, `jsonData`, `secureJsonData`). It does not change how Grafana stores datasource configuration today. Every field a dsconfig schema describes still lives exactly where it lives now. Adopting dsconfig for an existing plugin requires no rename, no data migration, and no change to existing plugin behavior — see "Design posture: additive, not migratory" below.

dsconfig exists to give four consumers, who today have no shared, machine-readable contract to work from, exactly that:

- **Config editors** (frontend forms), today hand-written per plugin with no guarantee they match what the backend actually parses.
- **Backend settings parsing** (`grafana-plugin-sdk-go`), today reading untyped `jsonData`/`secureJsonData` maps with no schema-level contract.
- **Provisioning** (`datasources.yaml`) and the **Grafana App Platform / Kubernetes-style datasource API**, both of which need a description of a valid datasource config that exists independently of any running plugin instance, so a config can be validated before it is applied.
- **Automated and assisted configuration**, including the Grafana Assistant chat-driven datasource workflow, and any other tool that needs to read, generate, or validate a datasource configuration without parsing plugin source code.

### Why this schema exists: two primary drivers

**Driver 1 — App Platform / Kubernetes-style API compatibility.** Grafana's App Platform exposes resources through a Kubernetes-style API (CRD-shaped: `apiVersion`, `kind`, `metadata`, `spec`, `status`). A datasource's `spec` needs an OpenAPI-shaped schema describing what a valid instance of that resource looks like — the same way any Kubernetes Custom Resource Definition needs a structural schema for its `spec`. Datasource `jsonData`/`secureJsonData` today have no such schema; they are untyped maps. dsconfig is the semantic layer that produces that OpenAPI-shaped schema (see "SDK PluginSettings" in the topology diagram below) from one declarative source per plugin, so App Platform's CRUD operations on the datasource resource — create, read, update, delete, including admission-time validation — have a structural schema to validate against, exactly as App Platform's own resource model expects.

**Driver 2 — reliable, automatically derived HTTP clients.** Today, the logic that turns a plugin's stored configuration into a working HTTP client (TLS setup, auth header/round-tripper wiring, timeout configuration) is hand-written, per plugin, and frequently duplicated with small inconsistencies across otherwise similar plugins. Because dsconfig fields carry typed, structured metadata — storage location, validation rules, and, in future schema versions, semantic role information — rather than being opaque map entries, the same schema that drives config-editor generation and provisioning validation is also the input a future SDK helper can use to build a transport-correct, auth-correct HTTP client without per-plugin code. The current schema version does not implement that derivation; see "Known limitations" below. The schema is shaped, from this version forward, so that derivation is possible without a breaking change to any field already defined here.

### Design posture: additive, not migratory

Every design decision in this document follows one rule: **adopting dsconfig must never require a plugin to change what it stores, how it stores it, or where.** The schema describes `root`, `jsonData`, and `secureJsonData` exactly as Grafana persists them today. This is why storage target enumerates exactly the three locations Grafana already has (see "Storage target" below), why the `direct` storage mapping is a no-op over today's default behavior, and why adopting this schema for an existing plugin is, by construction, a documentation exercise first and an automation opportunity second.

## Root schema

| name          | type                | required  | description                                   |
| ------------- | ------------------- | --------- | --------------------------------------------- |
| schemaVersion | string              | Required. | Schema spec version (e.g. "v1").              |
| pluginType    | string              | Required. | Unique plugin identifier (e.g. "prometheus"). |
| pluginName    | string              | Required. | Human-readable name.                          |
| docURL        | string              | Optional  | documentation URL.                            |
| fields        | ConfigField[]       | Required. | Source of truth for all config fields.        |
| groups        | ConfigGroup[]       | Optional  | UI layout grouping.                           |
| instructions  | Instruction[]       | Optional  | Instructions for LLMs and other consumers.    |
| relationships | FieldRelationship[] | Optional  | semantic relationships between fields.        |

`schemaVersion` versions the dsconfig schema _format_ itself — the shape described by this document — not the plugin's own configuration version. `pluginType` is the join key between a schema document and a real, running plugin instance, and is also the key an App Platform resource's `apiVersion`/`kind` would use to locate the structural schema for a given datasource kind.

`instructions` deserves a specific callout given dsconfig's role in assisted configuration: this is the field a chat-driven workflow such as the Grafana Assistant reads for plugin-specific guidance that individual field definitions don't otherwise convey — for example, "ask for the externally reachable URL, not an internal one" or "TLS settings only matter if the server enforces mutual TLS." `instructions` has no effect on validation or storage; its only audience is a consumer trying to reach a working configuration in as few exchanges as possible.

## Field identity: `id` vs `key`

| Property | Purpose                    | Scope                                        | Example                   |
| -------- | -------------------------- | -------------------------------------------- | ------------------------- |
| `id`     | Canonical schema reference | Globally unique across the entire schema     | `"httpHeaders.item.name"` |
| `key`    | Storage/object key         | Local to its storage target or parent object | `"name"`                  |

Groups and relationships reference fields by `id`.

This separation is what keeps the schema additive. `key` always matches what is already being stored, today, by the plugin, with zero changes — it is the identifier Grafana's existing storage model and any existing plugin backend code already expect. `id`, by contrast, is a schema-authoring concern: it never appears in stored configuration, and changing a field's `id` changes nothing about how or where its value is persisted. `id` exists purely to give every other mechanism in this schema — groups, relationships, effects, and any future role- or scope-based mechanism — a stable, storage-independent name to reference. A recommended, but not currently enforced, convention is a dot-separated path describing the field's logical position (as in the examples above); see "Known limitations."

### `id` format constraints

Every `id` is checked against two rules at validation time:

1. **Character set.** Each dot-separated segment must match `^[A-Za-z_][A-Za-z0-9_]*$` — ASCII letters, digits, and underscore, not starting with a digit. No hyphens, brackets, or spaces.
2. **No prefix collisions.** No `id` may be a strict dotted-path prefix of another `id` in the same document. `tls.clientAuth` and `tls.clientAuth.enabled` cannot both exist.

Both rules exist for the same reason: `dependsOn`, `requiredWhen`, `disabledWhen`, and `effects[].when` are expression strings that, once a future evaluator exists, will need to resolve `id`-shaped references inside them by parsing dotted paths — the same syntax CEL itself uses for member access. An `id` containing a character CEL treats specially, or an `id` that's ambiguous as a path prefix of another `id`, would silently break that resolution the moment an evaluator tried to use it, with no warning until then. Enforcing both rules now, before any evaluator exists, converts that invisible future failure into an immediate schema-validation error today. It does not, by itself, make any expression string correct or evaluated — see "Expression language," below, and "Known limitations."

## Storage target

`target` specifies where the field is stored in Grafana's datasource config:

| Value            | Maps to                                     |
| ---------------- | ------------------------------------------- |
| `root`           | Top-level fields (`url`, `basicAuth`, etc.) |
| `jsonData`       | `jsonData.*`                                |
| `secureJsonData` | `secureJsonData.*` (write-only)             |

**Required** for storage fields. **Omitted** for virtual fields and item fields.

These three values are exactly the three storage locations Grafana's datasource model already provides; this schema introduces no fourth location, consistent with "Design posture: additive, not migratory" above. Each storage field, together with its `valueType`, `validations`, and required/`requiredWhen` state, supplies exactly the information an OpenAPI-style structural schema needs for one property — type, constraints, and required-ness. This is the basis for the conversion into the SDK `PluginSettings` shape shown in the topology section below, and in turn the basis for the structural schema a Kubernetes-style Custom Resource Definition needs for this plugin's datasource `spec` under Grafana's App Platform.

### Secure fields

Fields targeting `secureJsonData` are **write-only**. When reading existing config, consumers should use `secureJsonFields` (a `Record<string, boolean>`) to determine whether a secret is already configured. The schema describes the field; it does not imply the secret value is retrievable.

This is the one place the schema is describing an asymmetry that already exists in Grafana's storage model, not introducing one: `secureJsonFields` is Grafana's existing mechanism for letting a config editor show "a value is set" without ever exposing the value itself, and this schema's write-only framing of `secureJsonData` fields is intentionally consistent with that existing behavior rather than a new constraint. See "Known limitations" for how a field described here relates to `secureJsonFields` at the level of detail a generated config editor would need.

## Storage mapping

The `storage` property defines how logical fields map to Grafana's legacy storage format.

| Type          | Description                                                                    |
| ------------- | ------------------------------------------------------------------------------ |
| `direct`      | Default. `target` + `key` maps directly.                                       |
| `indexedPair` | Legacy indexed key/value pattern (e.g. `httpHeaderName1`, `httpHeaderValue1`). |
| `computed`    | Declarative read/write expressions. Execution is runtime-specific.             |

`computed` mappings store CEL-like expressions but are **not evaluated** by the schema validator.

`direct` is, by construction, a no-op description of what Grafana already does for an ordinary field — declaring it is optional precisely because it changes nothing about today's behavior. `indexedPair` is the one mapping type that describes an existing legacy convention rather than today's default behavior: HTTP headers, and a handful of other Infinity-style plugin conventions (query string parameters, secure query string parameters), are stored today as a numbered series of properties split across `jsonData` and `secureJsonData`. This mapping type gives that existing, already-shipping convention a structured description; it does not change the convention itself. See "Known limitations" for the current scope of what reads an `indexedPair` or `computed` mapping against real data.

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

This distinction matters specifically for the assisted-configuration use case: a chat-driven workflow generating or repairing a datasource configuration should read `validations[]` to know what is actually acceptable, since `ui.options` may legitimately be a subset chosen for a particular rendering and is not guaranteed to be authoritative.

### Rule types

| Type            | Required fields    | Purpose                               |
| --------------- | ------------------ | ------------------------------------- |
| `pattern`       | `pattern`          | Regex validation for strings          |
| `range`         | `min` and/or `max` | Numeric bounds                        |
| `length`        | `min` and/or `max` | String length bounds                  |
| `itemCount`     | `min` and/or `max` | Array size bounds                     |
| `allowedValues` | `values`           | Enumerated allowed values             |
| `custom`        | `expression`       | CEL expression (evaluated at runtime) |

`pattern`, `range`, `length`, `itemCount`, and `allowedValues` are translated directly into real OpenAPI/JSON Schema constraints when a schema is converted into SDK `PluginSettings` form (`pattern`, `minimum`/`maximum`, `minLength`/`maxLength`, `minItems`/`maxItems`, and `enum`, respectively) — these five rule types are fully operative today. `custom`'s `expression` is the one rule type whose "evaluated at runtime" description is aspirational rather than current: see "Known limitations."

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

Because `any` fields skip type-level validation, they are also the fields most likely to be ambiguous to an automated consumer — including an assisted-configuration workflow trying to populate a value with no type signal to constrain it. The `description` field is doing real work here, not just documentation; a thorough, example-bearing `description` is the only signal such a consumer has for an `any`-typed field, which is the practical reason this section's guidance to "document the expected shapes" is a requirement in practice, not a suggestion.

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

The array-of-objects shape shown here — rather than the legacy numbered-property convention described under "Storage mapping" — is the recommended shape for any _new_ field that needs a user-extensible list of structured values. `indexedPair` exists to describe an already-shipped legacy convention faithfully; it is not the preferred shape for new plugin development. A new plugin needing, for example, a list of custom HTTP headers should model it as an array field exactly as shown here, with a `direct` (or simply omitted) storage mapping, not as a from-scratch `indexedPair`.

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

`effects` is deliberately structured rather than expressed as a CEL write expression. The set of "selector value picked → these fields get these values" relationships that occur in practice across datasource plugins is small and enumerable, and representing it as a validated `when`/`set` pair — rather than an opaque string naming a side-effecting function — lets schema validation confirm every `set` key resolves to a real field `id` without needing to parse or evaluate the `when` condition itself to do so. `when` remains a CEL-like string today and is, consistent with the rest of this schema's expression fields, not evaluated; only its presence is checked. See "Known limitations" and "Expression language" below.

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

The `tags: ["managed-by:..."]` convention named above is documentation, not a mechanism: it is free text intended for a human (or an LLM) reading the schema, and it is not parsed or checked against the actual `effects` declarations that establish the same relationship. The authoritative record of "what drives this field" is the `effects[].set` entries that name it; `managed-by` tags are a convenience cross-reference layered on top, not a second source of truth. See "Known limitations" regarding `tags` more generally.

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

This pattern works for exactly the depth shown — `section` supports one level of nesting per field, so a depth-2 recursive structure like the one above is expressed as two distinct dotted `section` values (`"azureCredentials"` and `"azureCredentials.serviceCredentials"`), not as a single field referencing itself. See "Known limitations" for the boundary of what this approach can express if a real plugin's recursion goes deeper than the bounded depth assumed here.

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

This is currently the only documented use of `computed`'s `write` expression in this schema, and it is narrower than the general "any CEL write expression" framing might suggest: it is, in substance, one named operation (split an array's elements between two storage locations by a boolean field) wearing CEL call syntax. See "Known limitations" and the accompanying RFC for the distinction between genuinely general expression evaluation and this narrower, nameable class of operation.

### Shared field sets

Many datasources (~30+) share TLS, basic auth, timeout, and cookie-forwarding fields. Rather than schema-level `$ref` or includes, use **code-level helpers** that inject common field sets during schema construction:

- **Go:** `schema.BasicAuthFields()`, `schema.TLSFields()`, `schema.CommonNetworkFields()`, `schema.HTTPHeaderFields()`
- **TypeScript:** `basicAuthFields()`, `tlsFields()`, `commonNetworkFields()`, `httpHeaderFields()` from `schema/common.ts`

Generated JSON files remain self-contained — no resolution step needed for consumers.

This convergence is also why HTTP client derivation (Driver 2, above) is realistic as a near-term goal specifically for TLS and basic-auth fields: plugins built using these shared helpers already use near-identical field names and shapes, which is the single largest practical obstacle to deriving a working HTTP client automatically from a schema. A future role- or name-based resolution mechanism (see the accompanying RFC) has the most leverage exactly where this convergence already exists.

## Groups and relationships

**Groups** define UI layout sections. They reference fields by `id`.
Set `"optional": true` on groups that can be collapsed or hidden by default (e.g. advanced sections):

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

## Expression language

Expression fields (`dependsOn`, `requiredWhen`, `disabledWhen`, `overrides[].when`, `storage.computed.read/write`, `custom` validation `expression`) are **opaque strings** in v1. The intended language is CEL. This PR stores expressions but **does not evaluate them**. Runtime evaluation is follow-up work.

This is the single most consequential limitation in this schema version, stated here without euphemism: every example elsewhere in this document that uses `dependsOn` or `requiredWhen` (for instance, "show username only when `auth.method == 'basic-auth'`") documents _intent_. None of it is _enforced_ by anything in the current reference implementations. A schema document can declare a condition referencing a field that does not exist, or contain a typo in the comparison operator, and no validation step in this schema version will detect it. Authors and reviewers should treat every expression-string example in this document as illustrative of the target behavior, not as a description of behavior that exists today. See "Known limitations" and the accompanying RFC's evaluation of whether, and how narrowly, runtime evaluation should be built.

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

## Reading values back out by `id`

Everything above describes how to _author_ a schema. This section covers the read-side counterpart: given an `id` and a real configuration payload, how do you get the actual configured value back out? Every reference implementation (`SCHEMA-V1.go`, `SCHEMA-V1.ts`) ships the same four functions for this, under the same names (`FieldByID`/`fieldById`, `ValueByID`/`valueById`, `ResolveIndexedPairs`/`resolveIndexedPairs`, `ResolveIndexedPairsAsMap`/`resolveIndexedPairsAsMap`).

This matters for any consumer that's handed an `id` from somewhere else — a person or an assistant referring to a field by name, a `groups[].fieldRefs` entry being rendered, a future runtime validator checking a real payload — and needs to resolve it without re-implementing the `target`/`section`/`key`/`storage` walk by hand.

**`FieldByID(schema, id)`** resolves an `id` to its field definition. Pure schema lookup; no configuration data involved.

**`ValueByID(schema, id, settings)`** resolves an `id` to its actual configured value in a given `settings` payload (shaped like Grafana's real storage: root-level keys at the top, `jsonData` and `secureJsonData` as nested objects). This only works for fields with no `storage` mapping, or `storage.type: "direct"` — for an `indexedPair` field it returns an error pointing you at the next two functions instead of guessing at a single storage location that doesn't exist for that mapping type.

**`ResolveIndexedPairs(field, jsonData, secureJsonData)`** resolves an `indexedPair`-mapped field (HTTP headers, and anything using the same convention) into an array of objects matching the field's `item` schema — `[{name, value}, ...]` — by scanning sequentially from `storage.startIndex` and stopping at the first missing index.

**`ResolveIndexedPairsAsMap(field, jsonData, secureJsonData)`** resolves the same kind of field into a flat `name -> value` map instead, by scanning every key actually present rather than stopping at a gap. This is gap-tolerant where the array version isn't, at the cost of collapsing duplicate names into one entry and representing a missing value as an empty string rather than an absent key. Use whichever cost your case can tolerate — see "Known limitations" for the full trade-off.

```go
val, err := schema.ValueByID("connection.orgSlug", settings)
// val == "my-organization"

pairs, err := schema.ResolveIndexedPairs(headersField, settings["jsonData"], settings["secureJsonData"])
// pairs == []map[string]any{{"name": "X-Custom-Header", "value": "custom-value"}, ...}

flat, err := schema.ResolveIndexedPairsAsMap(headersField, settings["jsonData"], settings["secureJsonData"])
// flat == map[string]string{"X-Custom-Header": "custom-value", ...}
```

None of these four functions evaluate `dependsOn`/`requiredWhen`/`disabledWhen`/`effects[].when`, and none of them read a `computed` storage mapping — both remain unevaluated, per "Expression language" and "Known limitations." They also cannot read a secret value out of a real, already-saved datasource's settings, because `secureJsonData` is write-only once saved and Grafana's own API never returns it; called against such a payload, a secret-targeted field looks identical to one that was simply never configured. They work as expected against a schema's own example payloads, or any payload under direct, local construction that genuinely embeds the secret values.

## DSConfig schema topology

Every example in this document implies three representations of the datasource config / schema, shown here explicitly:

1. **Schema** — the dsconfig schema definition (source of truth)
2. **Grafana Storage** — what gets persisted in Grafana's datasource config model (`root` / `jsonData` / `secureJsonData`, unchanged from today)
3. **SDK PluginSettings** — the OpenAPI-shaped spec produced from the schema for `grafana-plugin-sdk-go`, and, in turn, the structural schema App Platform's Kubernetes-style datasource resource API requires for this plugin's `spec`

```text
┌───────────────┐    convert to OpenAPI shape   ┌────────────────────────┐
│   dsconfig    │ ─────────────────────────────►│   SDK PluginSettings   │
│   schema      │                               │  (spec + secureValues) │
└──────┬────────┘                               └────────────────────────┘
       │
       │  describes (unchanged)
       ▼
┌─────────────────────────────────┐
│  Grafana Storage                │
│  (root / jsonData /             │
│   secureJsonData)                │
└─────────────────────────────────┘
```

The middle box — SDK PluginSettings — is also the artifact the Grafana App Platform layer consumes when describing this plugin's datasource as a Kubernetes-style resource: the same OpenAPI-shaped `spec` that `grafana-plugin-sdk-go` validates against is the structural schema App Platform's CRUD and admission-validation machinery needs for that resource's `spec`. One schema document, authored once per plugin, is the source for both consumers.

## Known limitations

This section records, deliberately and without euphemism, what the current schema version does not yet do. Each item below is scoped as a future, additive enhancement — none requires changing the shape of any field already documented above, and none requires migrating any already-stored datasource configuration or any already-published schema document. See the accompanying RFC for the proposed sequencing of this work and the corresponding "Known Limitations" sections in `SCHEMA-V1.go` and `SCHEMA-V1.ts`, which enumerate the identical list for their respective reference implementations.

1. **Expression strings are not parsed or evaluated.** `dependsOn`, `requiredWhen`, `disabledWhen`, `overrides[].when`, `effects[].when`, `storage.computed.read`/`write`, and `custom` validation's `expression` are all opaque, CEL-flavored strings. None is parsed against a grammar at schema-validation time, and none is evaluated against real configuration data by any reference implementation. A typo, or a reference to a nonexistent field, inside one of these strings is not detected until — at the earliest — a future runtime evaluator exists.
2. **`storage` is descriptive metadata, not yet an executable mapping.** `indexedPair` and `computed` are validated structurally (the right sub-properties are present for the declared type), but no reference implementation reads a `storage` mapping and applies it against real stored configuration — for example, expanding an `indexedPair` mapping's pattern against an actual `jsonData`/`secureJsonData` payload to produce a real list of name/value pairs.
3. **The OpenAPI conversion does not yet read `storage` at all.** It places fields by `target` and `section` only. A consequence specific to `indexedPair`: when a pair's value target is `secureJsonData`, the generated OpenAPI settings give no indication that the corresponding array's values are secret. This is a correctness gap in the generated output, not merely a missing feature, and is the highest-priority item in this list.
4. **`section` supports exactly one level of nesting.** The modeling pattern this document recommends for self-referential structures (see "Recursive types," above) relies on manually flattening each level of recursion into its own `section` value, which remains practical only while actual recursion depth in real plugin configurations stays small.
5. **No field carries a semantic role independent of its name.** Two fields that mean the same thing (for example, a base URL) but are named differently across plugins (`apiUrl`, `baseURL`, `endpoint`) cannot currently be recognized as equivalent by any automated consumer without that consumer hard-coding plugin-specific name lists. This directly limits the reliability of automated HTTP client derivation (Driver 2, above): without a name-independent way to identify "this field is the TLS client certificate" or "this field is the basic-auth password," a generic client-builder cannot be written against this schema alone today.
6. **Auth representation is whatever the plugin author chose**, with no schema-level distinction between an explicit discriminator field, a set of independently-toggleable boolean flags, or a hybrid of the two, and no detection of mutually incompatible combinations of the latter.
7. **No mechanism exists for a plugin with more than one independent connection** within a single datasource instance (for example, a plugin that calls two unrelated backend APIs, each with its own URL, auth, and TLS settings).
8. **`id` format is now partially enforced.** Every `id` segment is checked against `^[A-Za-z_][A-Za-z0-9_]*$`, and no `id` may be a strict dotted-path prefix of another `id` in the same document — see "`id` format constraints," above. This closes the specific failure mode of an `id` that would break a future expression evaluator's dotted-path resolution. It does not enforce the recommended hierarchical-by-meaning convention beyond that, and it does not make `dependsOn`/`requiredWhen`/`disabledWhen`/`effects[].when` evaluated — see item 1. A schema document written before this check existed that happens to violate either rule will now fail validation; this is a deliberate, newly-enforced behavior change within the `v1` schema version, not a `v1`-to-`v2` migration.
9. **`tags` and `examples` are accepted and stored but not read by any reference implementation.** They are retained as forward-compatible, purely descriptive metadata.
10. **`repeatable` and `pattern`** (on a field, distinct from a storage mapping field's `pattern`) **are accepted and stored but not read** by validation or by the OpenAPI conversion. Their relationship to `storage`'s `indexedPair` — which models the same legacy indexed-field convention more explicitly — is not yet resolved.
11. **There is no schema-version migration mechanism.** `schemaVersion` is a required string, but no reference implementation interprets it beyond requiring its presence.
12. **Validation error-collection behavior differs across reference implementations.** The Go validator returns on the first error encountered; the TypeScript validator collects every error found before returning. A single invalid schema document can currently surface a different number of reported problems depending on which language validated it.
13. **`ResolveIndexedPairs` infers pair-role by item-field position, not by declaration.** It assumes the first declared `item.fields` entry is the pair's "name" and the second is its "value," because neither `ConfigField` nor the item schema carries an explicit pair-role tag today. A schema declaring these two item fields in the opposite order produces silently swapped results, with no validation error at any point.
14. **`ResolveIndexedPairsAsMap` collapses duplicate names and can't distinguish a missing value from a genuinely empty one.** Two distinct stored pairs sharing the same name resolve to one map entry (iteration order, and therefore which one wins, is unspecified), and a name with no corresponding value present maps to an empty string — indistinguishable from a value that was actually configured as an empty string. `ResolveIndexedPairs` has neither limitation but is not gap-tolerant the way this function is; pick based on which cost your case can tolerate.
15. **None of the lookup/resolver functions can read a live secret value.** `ValueByID`, `ResolveIndexedPairs`, and `ResolveIndexedPairsAsMap` all produce the same "absent" result for a `secureJsonData`-targeted field whether it was never configured or whether it's simply write-only and unreadable post-save, because Grafana's own API never returns a saved secret. These functions work as expected against a schema's own example payloads or any payload under direct, local construction.
