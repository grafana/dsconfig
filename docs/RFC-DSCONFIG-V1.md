# RFC-DSCONFIG-V1: `dsconfig` — A Declarative Schema Contract for Grafana Datasource Plugin Configuration

| Field               | Value                                                                                                                                                                                                  |
| ------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| RFC ID              | RFC-DSCONFIG-V1                                                                                                                                                                                        |
| Title               | `dsconfig`: A Declarative Schema Contract for Grafana Datasource Plugin Configuration                                                                                                                  |
| Status              | Proposed                                                                                                                                                                                               |
| Author              | Sriram                                                                                                                                                                                                 |
| Component           | `grafana-plugin-sdk-go`, `grafana/grafana` (datasource provisioning, plugin settings), Grafana App Platform, Plugin Catalog, Grafana Assistant                                                         |
| Companion artifacts | `SCHEMA-V1.md` (reference documentation), `SCHEMA-V1.go` (Go reference implementation), `SCHEMA-V1.ts` (TypeScript reference implementation), `SCHEMA-V1.json` (formal JSON Schema of the wire format) |

---

## 1. Executive Summary

This RFC proposes `dsconfig`: a declarative, language-neutral schema for describing the configuration surface of a Grafana datasource plugin. A `dsconfig` schema document, authored once per plugin, formally describes every configurable field — its type, its storage location, its validation rules, its presentation hints, and its relationships to other fields — as a single artifact consumable by multiple independent systems that today have no shared, machine-readable contract to work from.

`dsconfig` is a semantic layer placed on top of Grafana's existing datasource configuration storage. It does not change that storage. Every field a `dsconfig` schema describes continues to live exactly where it lives today — as a root-level datasource property, within `jsonData`, or within `secureJsonData`. No plugin is required to rename a field, move a value, or alter its persisted configuration shape in order to adopt this schema. This RFC is additive by design, not migratory.

Four business drivers motivate this proposal, in order of direct organizational impact:

1. **Grafana Assistant and AI-assisted configuration.** A chat-driven assistant capable of creating or modifying a datasource on a person's behalf needs a single, structured artifact describing what fields exist, which are required, what values are valid, and which fields are secret — so that it can drive toward a working, validated connection in the fewest possible turns, without parsing plugin source code or guessing at undocumented form behavior.
2. **Grafana App Platform / Kubernetes-style resource API compatibility.** App Platform manages Grafana resources through a Kubernetes-style API requiring an OpenAPI-shaped structural schema for every resource's `spec`. Datasource configuration has no such schema today; `dsconfig` is the mechanism that produces one, per plugin, from a single declarative source, enabling App Platform to perform create/read/update/delete and admission-time validation against datasource configuration the same way it does for any other Kubernetes-style resource.
3. **Reliable, automatically derived HTTP clients.** The logic that turns a plugin's configuration into a working network client is hand-written and duplicated, with inconsistencies, across the plugin catalog. `dsconfig` structures field metadata so that this derivation becomes possible as a shared, tested capability rather than per-plugin code, directly serving the goal of a successful connection being established as quickly and reliably as possible — the same goal that motivates the Assistant use case above.
4. **A verifiable contract between configuration editors, backend parsing, and provisioning.** Today these three surfaces are kept in agreement only by shared developer memory. `dsconfig` gives them one source of truth.

This RFC defines the schema's complete structure (Section 5), the conversion of that structure into the OpenAPI-shaped specification consumed by `grafana-plugin-sdk-go` and, transitively, by the App Platform resource API (Section 6), the Grafana Assistant consumption model (Section 7), an explicit accounting of this design's advantages and disadvantages together with mitigations for each disadvantage (Section 8), and a full, itemized statement of what this initial release deliberately does not yet do (Section 10).

## 2. Problem Statement

### 2.1 Absence of a verifiable contract between configuration surfaces

A datasource plugin's configuration is read and written by at least four independent surfaces: the plugin's `ConfigEditor` React component, the plugin's Go backend settings-parsing code, Grafana's provisioning YAML loader, and the persisted datasource record itself (`root` fields, `jsonData`, `secureJsonData`). Grafana's `DataSourceJsonData` and `DataSourceSettings` types declare the `jsonData` and `secureJsonData` payloads as untyped maps. No mechanism in the current system verifies that the shape written by the config editor matches the shape expected by the backend parser, or that either matches what a provisioning file supplies.

This absence is tolerated today because plugin authorship and plugin maintenance are typically performed by the same small group of engineers across both the frontend and backend halves of a given plugin, within a short time window. The agreement between the two halves is maintained by shared developer context, not by any verifiable artifact. This condition does not hold as plugin maintenance is transferred across teams or across time; the agreement degrades silently, and no system component is positioned to detect the degradation.

### 2.2 Deferred error surfacing

Because no validation step exists between configuration entry and plugin execution, malformed configuration is discovered only when the plugin attempts to use it — inside `CheckHealth` or `QueryData`. Errors at this point lack field-level attribution and are frequently reported as generic transport or authentication failures, with no path back to the specific configuration field responsible.

### 2.3 Absence of a shared vocabulary for recurring configuration concerns

TLS configuration, basic authentication, HTTP timeout settings, and custom HTTP headers recur across a substantial fraction of the datasource plugin catalog. No canonical definition of these field sets exists; each plugin's field names, value types, and storage placement for these concerns are determined independently by each plugin's authors, with consistency across plugins arising only where one plugin's implementation was copied as a starting point for another.

### 2.4 Absence of a machine-readable configuration description

No artifact exists today that describes a datasource plugin's configuration surface independently of that plugin's source code. Documentation, where it exists, is maintained separately from the parsing code it describes and is not verified against it. This is tolerable for a human reader, who can resolve minor documentation staleness through judgment. It is not tolerable for any non-human consumer — a provisioning validator, a plugin catalog auditor, a code-generation tool, or, most materially for this RFC, **a conversational assistant acting on a person's behalf**, which has no structured input to act on in the absence of such an artifact and must otherwise either guess or refuse to help.

### 2.5 Absence of a structural schema for the App Platform resource model

Grafana's App Platform manages resources as Kubernetes-style custom resources, each requiring a structural, OpenAPI-shaped schema describing the valid shape of its `spec`, in the same way every Kubernetes Custom Resource Definition requires a structural schema for admission validation, `kubectl` client-side validation, and API server storage validation. Datasource configuration cannot be exposed through this resource model today because `jsonData` and `secureJsonData` are untyped maps with no per-plugin structural schema for App Platform to validate against. This is a distinct, concrete blocker — not a generalization of Section 2.4 — because App Platform's CRUD machinery requires a schema in a specific OpenAPI-compatible shape as a hard precondition of resource registration, not merely as a documentation convenience.

### 2.6 Scope of this RFC relative to the problem statement

This RFC addresses Sections 2.1 through 2.5 by introducing a schema artifact, a structural validation function, and a conversion path to an SDK- and App-Platform-consumable specification. It does not, in its initial scope, address runtime evaluation of conditional logic between fields, automatic derivation of network clients from configuration, or standardization of authentication discriminator patterns across the existing plugin population. These are recorded in Section 10 (Known Limitations) and Section 11 (Future Work).

## 3. Goals

1. Define a single, versioned schema format capable of describing every field in a datasource plugin's configuration, regardless of its storage location (`root`, `jsonData`, `secureJsonData`).
2. Provide a structural validation function that verifies a schema document is internally well-formed and self-consistent prior to any use.
3. Provide a conversion function that transforms a `dsconfig` schema into the OpenAPI-shaped settings specification already consumed by `grafana-plugin-sdk-go`, and which is structurally suitable for direct use as the `spec` schema of a Grafana App Platform Kubernetes-style datasource resource.
4. Preserve Grafana's existing datasource storage model without modification. The schema describes the existing `root`/`jsonData`/`secureJsonData` structure; it does not propose a new storage format, and no migration of any existing stored configuration is required at any point.
5. Support the configuration patterns already present in the plugin catalog, including direct field storage, legacy indexed key/value pair storage (e.g., custom HTTP headers), array-of-object storage, and computed/derived (virtual) fields.
6. Provide a declarative mechanism for expressing that a single user-facing control (e.g., an authentication method selector) writes to multiple underlying storage fields, without requiring a general-purpose expression evaluator.
7. Provide UI presentation metadata (component type, options, layout width) separable from the data validation contract, such that validation tooling is never required to parse UI metadata to determine a field's data constraints.
8. Provide grouping and relationship metadata sufficient to drive a structured, schema-driven configuration page layout, separate from the validation contract.
9. Provide a single artifact sufficiently structured and complete that a conversational assistant can consume it directly to construct, explain, and validate a datasource configuration without access to plugin source code.
10. Shape the schema, from its first version forward, such that automatic derivation of a working HTTP client from a schema and a configuration payload is a natural future extension rather than a redesign.

## 4. Non-Goals

1. This RFC does not propose changes to Grafana's persisted datasource data model.
2. This RFC does not implement runtime evaluation of any conditional expression. Expression fields are accepted, stored, and structurally validated as well-formed strings; their evaluation is explicitly out of scope (Section 10.1).
3. This RFC does not implement automatic derivation of HTTP clients, SQL connection strings, or any other runtime client object from schema and configuration data. The schema describes configuration; consumption of that description to construct runtime objects is left to future work (Section 11.6).
4. This RFC does not propose retroactive migration of the existing plugin catalog's field names, storage formats, or authentication conventions. Existing plugins are not required to change their stored configuration shape to adopt this schema.
5. This RFC does not define a standardized authentication discriminator pattern across plugins. Plugins remain free to express authentication configuration using whatever field structure they currently use; the schema describes that structure as-is.
6. This RFC does not address multi-connection datasource plugins (plugins requiring more than one independent network connection within a single datasource instance) as a distinct first-class concept.
7. This RFC does not implement the App Platform resource type registration itself, nor the Grafana Assistant tool integration itself. It defines the schema artifact both depend on; the integration work for each is separate, subsequent engineering scoped outside this document.

## 5. Design Overview

### 5.1 Position relative to existing storage

`dsconfig` is a descriptive layer. It does not change where Grafana stores datasource configuration. Every field described by a `dsconfig` schema resolves to one of three existing storage locations:

| Target           | Resolves to                                                                              | Encrypted at rest | Readable after save |
| ---------------- | ---------------------------------------------------------------------------------------- | ----------------- | ------------------- |
| `root`           | Top-level fields on the datasource record (e.g., `url`, `basicAuth`, `database`, `user`) | No                | Yes                 |
| `jsonData`       | `jsonData.*`                                                                             | No                | Yes                 |
| `secureJsonData` | `secureJsonData.*`                                                                       | Yes               | No (write-only)     |

A `dsconfig` schema is valid and meaningful independent of any change to Grafana core, to existing plugins' stored configuration, or to the provisioning file format. This is the single rule every other design decision in this RFC follows: **adopting `dsconfig` must never require a plugin to change what it stores, how it stores it, or where.**

### 5.2 Schema document structure

The full root-level structure, with every key explained, is as follows. The authoritative, machine-readable version of this table is the companion file **`SCHEMA-V1.json`**; the authoritative reference implementations are **`SCHEMA-V1.go`** (Go) and **`SCHEMA-V1.ts`** (TypeScript).

| Key             | Type                         | Required | Explanation                                                                                                                                                                                                                                                                                                                                                 |
| --------------- | ---------------------------- | -------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `schemaVersion` | string                       | Yes      | Identifies which version of the `dsconfig` specification this document conforms to (e.g., `"v1"`). Every consumer must check this before interpreting the rest of the document, since the meaning of every other key is defined relative to this version. A version increment is required for any change to this specification that is not purely additive. |
| `pluginType`    | string                       | Yes      | The plugin's unique, stable type identifier, matching the value already declared in the plugin's `plugin.json`. This is the join key used by every system that needs to locate the correct schema for a given datasource instance — provisioning, App Platform, the plugin catalog, and Grafana Assistant alike.                                            |
| `pluginName`    | string                       | Yes      | A human-readable display name for the plugin (e.g., `"Prometheus"`). Used in generated documentation, configuration page rendering, and Assistant responses; has no effect on validation or storage.                                                                                                                                                        |
| `docURL`        | string                       | No       | An optional link to the plugin's external documentation, surfaced in generated documentation, configuration page UI, and Assistant responses.                                                                                                                                                                                                               |
| `fields`        | array of `ConfigField`       | Yes      | The complete, authoritative list of every configuration field the plugin exposes. This is the schema's sole source of truth: no configuration field may exist outside this list, and every other top-level key (`groups`, `relationships`, `instructions`) only ever references entries here by `id`.                                                       |
| `groups`        | array of `ConfigGroup`       | No       | Optional layout metadata grouping fields into named, orderable sections (e.g., "Connection", "Authentication", "Advanced") for a rendered configuration page. Affects presentation only.                                                                                                                                                                    |
| `instructions`  | array of `Instruction`       | No       | Optional, free-form guidance entries intended primarily for non-human consumers — most directly, Grafana Assistant (see Section 7). Carries no validation or storage semantics.                                                                                                                                                                             |
| `relationships` | array of `FieldRelationship` | No       | Optional declarations of semantic connections between two or more fields not otherwise captured by individual field definitions (e.g., a username and password forming one logical credential). Metadata only.                                                                                                                                              |

A minimal valid schema document requires only `schemaVersion`, `pluginType`, `pluginName`, and at least one entry in `fields`.

### 5.3 Field identity: the `id`/`key` distinction

Every field carries two distinct identifiers, each serving a different audience:

| Key   | Purpose                                                                                                                                                                                                                                                                                                                                                                   | Scope                                        | Example                    |
| ----- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------- | -------------------------- |
| `id`  | The field's canonical, globally unique reference within the document. This is the only identifier used by `groups[].fieldRefs`, `relationships[].fields`, and `effects[].set` to refer to a field — it exists specifically so those references remain stable and unambiguous even when a field's storage key is short, generic, or reused in meaning across many plugins. | Globally unique across the entire document   | `"auth.basicAuthPassword"` |
| `key` | The field's local storage name — the literal JSON property name occupied within its storage `target`, or within its parent object/array element if nested. This is what actually appears in stored configuration; `id` never does.                                                                                                                                        | Local to its storage target or parent object | `"basicAuthPassword"`      |

A dot-separated, hierarchical naming convention for `id` (e.g., `"connection.url"`, `"auth.basicAuthPassword"`) is recommended for readability and consistency, but is not structurally enforced by schema validation in this release (see Section 10.6).

### 5.4 Full field definition: every key explained

Each entry in `fields` is a `ConfigField`. The complete set of keys, and the purpose of each, is as follows.

| Key            | Type            | Required      | Explanation                                                                                                                                                                                                                                                                                    |
| -------------- | --------------- | ------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `id`           | string          | Yes           | See Section 5.3.                                                                                                                                                                                                                                                                               |
| `key`          | string          | Yes           | See Section 5.3.                                                                                                                                                                                                                                                                               |
| `label`        | string          | No            | A human-readable display label for the field, used by configuration page UIs in place of `key` or `id`.                                                                                                                                                                                        |
| `description`  | string          | No            | A human-readable explanation of the field's purpose. Surfaced in generated documentation, configuration page helper text, and to Grafana Assistant as context for explaining or populating the field on a person's behalf.                                                                     |
| `docURL`       | string          | No            | A field-specific documentation link, distinct from the plugin-level `docURL`.                                                                                                                                                                                                                  |
| `valueType`    | enum            | Yes           | The field's JSON data type: `string`, `number`, `boolean`, `array`, `object`, `map`, or `any`. This is the single source of truth for the field's type across every consumer, including the OpenAPI-shaped spec produced for `grafana-plugin-sdk-go` and the App Platform resource API.        |
| `target`       | enum            | Conditionally | One of `root`, `jsonData`, `secureJsonData` (Section 5.1). Required for every field that is actually persisted; omitted for virtual fields (Section 5.7) and for fields declared inside an array/map item schema (Section 5.5), which inherit their storage context from their parent.         |
| `section`      | string          | No            | A dotted path prefix locating the field within a nested object inside its `target` (e.g., `target: jsonData`, `section: "tracesToLogs"` places the field at `jsonData.tracesToLogs.<key>`). Describes one level of existing object nesting; see Section 10.5 for the current depth limitation. |
| `kind`         | enum            | No            | `storage` (default) — a field that is actually persisted — or `virtual` (Section 5.7) — a field with no direct storage location of its own, used to model configuration-page state that does not map one-to-one onto a stored value.                                                           |
| `isItemField`  | boolean         | Conditionally | `true` for every field declared within an `item.fields` list (Section 5.5); required in that context, disallowed otherwise.                                                                                                                                                                    |
| `ui`           | object          | No            | Presentation hints only (Section 5.8). Never affects validation or storage.                                                                                                                                                                                                                    |
| `validations`  | array           | No            | The field's data contract — the authoritative description of what values are acceptable (Section 5.6).                                                                                                                                                                                         |
| `dependsOn`    | string          | No            | A conditional-visibility expression (intended grammar: CEL). Stored and structurally referenced; not evaluated in this release (Section 10.1).                                                                                                                                                 |
| `required`     | boolean         | No            | Unconditional requiredness, independent of any conditional expression.                                                                                                                                                                                                                         |
| `requiredWhen` | string          | No            | A conditional-requiredness expression. Same evaluation status as `dependsOn` (Section 10.1).                                                                                                                                                                                                   |
| `disabledWhen` | string          | No            | A conditional disabled/read-only-state expression. Same evaluation status as `dependsOn` (Section 10.1).                                                                                                                                                                                       |
| `overrides`    | array           | No            | Conditional modifications to this field's default value, description, placeholder, tooltip, validations, or UI options, each gated by a `when` expression. Same evaluation status as `dependsOn` for the `when` condition (Section 10.1); see also Section 10.9.                               |
| `effects`      | array           | No            | Structured, fully-validated multi-field write declarations (Section 5.7.1). Distinct from the expression-string mechanisms above: the _target_ of an effect (`set`) is validated today; only the _condition_ (`when`) is an unevaluated expression.                                            |
| `item`         | object          | Conditionally | Required when `valueType` is `array` or `map`; describes the element (array) or value (map) schema (Section 5.5).                                                                                                                                                                              |
| `repeatable`   | boolean         | No            | Reserved for describing legacy indexed-field patterns. Not validated or consumed in this release (Section 10.8); its relationship to `storage.indexedPair` is an open question.                                                                                                                |
| `pattern`      | string          | No            | Reserved, same status as `repeatable`. Distinct from `validations[].pattern` and from `storage`'s indexed-pair `pattern` sub-key.                                                                                                                                                              |
| `storage`      | object          | No            | An explicit storage mapping strategy, used when a field's persisted representation does not directly mirror its logical shape (Section 5.5.1).                                                                                                                                                 |
| `tags`         | array of string | No            | Free-form, unvalidated labels for loose conventions (Section 8, Design Decision D5). Not consumed by any resolver in this release.                                                                                                                                                             |
| `examples`     | array           | No            | Example values for documentation and placeholder purposes. Not consumed by validation or SDK conversion.                                                                                                                                                                                       |
| `defaultValue` | any             | No            | The value this field should take when not otherwise specified. Surfaced in configuration page UIs, documentation, and the produced OpenAPI-shaped spec.                                                                                                                                        |

### 5.5 Array and map fields

A field with `valueType: "array"` or `valueType: "map"` requires an `item` property describing the element (array) or value (map; map keys are always JSON strings) schema. An item schema is either a primitive `valueType`, or, when `valueType: "object"`, a list of nested `fields`, each of which must declare `isItemField: true` and may omit `target` (inherited from the parent).

This accommodates two storage patterns already present in the plugin catalog:

- **Direct array storage** — the persisted `jsonData` already contains a native JSON array of objects (e.g., Loki's `derivedFields`). No additional mapping is required.
- **Legacy indexed-pair storage** — the persisted configuration represents a logically array-like structure (e.g., custom HTTP headers) as a series of separately numbered scalar fields split across two storage targets (e.g., `jsonData.httpHeaderName1` paired with `secureJsonData.httpHeaderValue1`). Described via the `storage` property (Section 5.5.1) rather than requiring this legacy convention to change.

#### 5.5.1 Storage mapping

The optional `storage` property declares how a field's logical shape maps onto its actual persisted representation, when that representation diverges from a direct correspondence.

| `storage.type` | Explanation                                                                                                                                                                                                                                    |
| -------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `direct`       | The default. `target` and `key` map directly; no further declaration required.                                                                                                                                                                 |
| `indexedPair`  | The legacy numbered key/value pattern. Declares a `key` sub-mapping (`target` and `pattern`, e.g., `jsonData`, `"httpHeaderName{index}"`) and a `value` sub-mapping (e.g., `secureJsonData`, `"httpHeaderValue{index}"`), plus a `startIndex`. |
| `computed`     | The field's persisted value is derived via a declarative `read` and/or `write` expression rather than stored directly. Accepted and structurally validated; not executed in this release (Section 10.1, 10.2).                                 |

### 5.6 Validation rules

The `validations` array is the field's data contract — the authoritative description of acceptable values, independent of presentation.

| `validations[].type` | Required sub-keys  | Applies to     | Explanation                                                                                                                                                |
| -------------------- | ------------------ | -------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `pattern`            | `pattern`          | string fields  | The value must match the given regular expression.                                                                                                         |
| `range`              | `min` and/or `max` | numeric fields | The value must fall within the inclusive bound(s) given.                                                                                                   |
| `length`             | `min` and/or `max` | string fields  | The value's length must fall within the inclusive bound(s) given.                                                                                          |
| `itemCount`          | `min` and/or `max` | array fields   | The array's element count must fall within the inclusive bound(s) given.                                                                                   |
| `allowedValues`      | `values`           | any field      | The value must be one of the enumerated set given. This is the data-contract source of truth for allowed values.                                           |
| `custom`             | `expression`       | any field      | An arbitrary expression (intended grammar: CEL) the value must satisfy. Accepted and structurally validated; not evaluated in this release (Section 10.1). |

A field's `ui.options` (Section 5.8) may separately declare a displayed list of choices for a select-style control. `ui.options` is never authoritative for what values are acceptable; `validations[].type: allowedValues` is. Where both are present and a tool must determine acceptable values programmatically, `validations` governs. See Section 8, Design Decision D2.

### 5.7 Virtual fields and multi-field write declarations

A field with `kind: "virtual"` represents configuration-page state with no directly persisted value of its own. Virtual fields are excluded entirely from SDK conversion (Section 6); they exist only to support interaction patterns that do not map one-to-one onto a single stored value.

The canonical case is a selector control — for example, an "Authentication Method" dropdown — whose selection determines the values of several independently persisted fields at once (selecting "Basic Authentication" should set one persisted boolean `true` and another, unrelated boolean `false`). Two complementary mechanisms express this:

- A `storage.computed` mapping with a `read` expression, deriving the virtual field's displayed value from the present state of the underlying booleans when loading existing configuration. Not evaluated in this release (Section 10.1).
- An `effects` array (Section 5.7.1), declaring what happens when the virtual field's value changes.

#### 5.7.1 Effects

The `effects` array provides a structured, non-opaque mechanism for declaring multi-field write behavior.

| Key    | Type                       | Required | Explanation                                                                                                                                                                                                                                       |
| ------ | -------------------------- | -------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `when` | string                     | Yes      | A condition evaluated against the field's own current value (convention: the literal token `value` refers to it). Not evaluated in this release (Section 10.1) — this is the one part of `effects` that remains an unevaluated expression string. |
| `set`  | map of field `id` to value | Yes      | The fields, referenced by `id`, that should be set, and the values they take, when `when` holds. Must be non-empty.                                                                                                                               |

Every field `id` referenced in `effects[].set` is validated, at schema-validation time, against the complete set of field IDs declared in the document. This validation is structural, unconditional, and fully enforced today — it is the one part of the conditional-behavior surface in this schema that does not depend on a future expression evaluator, because the _target_ of the effect is a structured map rather than a second expression string.

### 5.8 UI metadata

The `ui` property declares presentation intent, independent of the field's data contract.

| Key                                                     | Explanation                                                                                                                      |
| ------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------- |
| `component`                                             | The UI control type: `input`, `textarea`, `select`, `multiselect`, `radio`, `checkbox`, `switch`, `code`, `keyvalue`, or `list`. |
| `options`                                               | For selection-type components, the `{label, value, description}` entries to display. Presentation only — see Section 5.6.        |
| `placeholder`, `multiline`, `rows`, `width`, `language` | Additional rendering hints (`language` is a syntax-highlighting hint for `code` components, e.g. `"promql"`, `"sql"`, `"json"`). |

`ui` metadata is validated for internal consistency (e.g., a declared `component` must be a recognized value; option values must match the field's `valueType`) but is not consumed by any rendering engine in this release. Construction of a renderer that builds a complete configuration page directly from `ui` and `groups` metadata is recorded as future work (Section 11.1).

### 5.9 Groups, relationships, and instructions

- **`groups`** define layout sections for a configuration page, each referencing field IDs and optionally marked `optional` (intended for collapsible/advanced sections). Validated against the document's field ID set.
- **`relationships`** declare semantic connections between fields: `pair` (two fields forming one logical unit, e.g., username/password), `group` (a looser related set), or `datasourceReference` (a field holding the UID of another datasource, optionally constrained to a specific plugin type via `targetPluginType`). Validated against the document's field ID set.
- **`instructions`** are free-form, tagged guidance strings intended for automated or AI-assisted tooling — most directly, Grafana Assistant (Section 7) — reasoning about the plugin's configuration. An instruction carries a `msg` and an optional `tags` list; it has no further validated structure.

None of these three affects how a field is stored or validated. Removing any of them from a schema document does not change the document's data contract.

## 6. SDK Conversion and the App Platform Resource API

### 6.1 Conversion mechanism and output envelope

A `dsconfig` schema, once structurally validated, is converted into the artifact already consumed by Grafana's plugin schema provider (`pluginschema.PluginSchema`) and, transitively, by `grafana-plugin-sdk-go` and the App Platform resource API. This artifact has two layers, and it matters which layer `dsconfig` actually populates.

**The outer envelope** — `PluginSchema{TargetAPIVersion, SettingsSchema, SettingsExamples, Routes, QueryTypes, QueryExamples}` — is what Grafana's schema provider (`NewSchemaProvider`/`NewCompositeFileSchemaProvider`) actually loads and serves per API version. `dsconfig`'s conversion populates `TargetAPIVersion` and `SettingsSchema`; `SettingsExamples` is populated where a schema author has supplied worked example payloads; `Routes`, `QueryTypes`, and `QueryExamples` are outside `dsconfig`'s scope entirely — they describe a plugin's resource routes and query-type schemas, not its connection configuration, and are populated independently by whatever else builds a plugin's full `PluginSchema` document.

**The inner `SettingsSchema` (`Settings{Spec, SecureValues}`)** is what `dsconfig`'s conversion function actually produces, and is the part this RFC's design decisions (Sections 5–8) are about:

- A **`Spec`** object: an OpenAPI-shaped schema (`k8s.io/kube-openapi`'s `spec.Schema`) describing the complete instance-settings object — `root`-target fields and a nested `jsonData` object — including types, descriptions, defaults, and validation constraints translated from `validations` (`pattern` → regex pattern, `range` → minimum/maximum, `length` → minLength/maxLength, `itemCount` → minItems/maxItems, `allowedValues` → enum).
- A **`SecureValues`** list: one entry per field targeting `secureJsonData`, carrying the field's key, description, and requiredness — structurally separate from `Spec`, never nested inside it, because a secret's existence and requiredness is information `Spec` is not the right place to carry (see Section 6.2's discussion of why `Spec` and `SecureValues` are siblings, not one merged object).

Virtual fields (`kind: "virtual"`) are excluded entirely from this conversion. Fields nested within a `section` are placed into the corresponding nested object automatically. Conditional-behavior expressions (`dependsOn`, `requiredWhen`, `disabledWhen`), where present, are carried into the produced specification as vendor extension properties (`x-dsconfig-depends-on`, `x-dsconfig-required-when`, `x-dsconfig-disabled-when`), preserving declared intent even though nothing in this release evaluates them.

### 6.2 Why this conversion target was chosen

The produced `Spec`/`SecureValues` pair is deliberately OpenAPI-shaped rather than a Grafana-proprietary format, for one direct reason: **this is the exact shape a Kubernetes-style Custom Resource Definition requires for its `spec` schema.** Grafana's App Platform represents resources — and is intended to represent datasources — as Kubernetes-style custom resources (`apiVersion`, `kind`, `metadata`, `spec`, `status`), and the App Platform's resource registration, admission-time validation, and API server storage validation machinery all require a structural, OpenAPI-compatible schema for `spec` as a precondition.

By producing this exact shape as a deterministic function of a `dsconfig` schema document, this RFC gives App Platform a path to expose datasource configuration as a properly typed, validated Kubernetes-style resource — supporting create, read, update, and delete operations against datasource configuration through the same resource model used elsewhere in App Platform — without requiring any change to Grafana's underlying `root`/`jsonData`/`secureJsonData` storage. The `spec` schema App Platform validates against and the storage Grafana actually persists to remain two different representations of the same configuration, connected by the conversion function defined in this section, exactly as a Kubernetes Custom Resource Definition's `spec` schema and a controller's actual reconciled state are two different representations connected by controller logic.

`SecureValues` sits beside `Spec`, not inside it, for the same reason `secureJsonData` is a separate storage bucket from `jsonData` in Grafana's existing model (Section 5.1): a `spec` schema that's meant to double as a standard, externally-validated OpenAPI document should not need every generic consumer of that document (the API server's admission validation, `kubectl`-equivalent client tooling) to special-case which of its properties are actually write-only secrets nested somewhere inside it. Keeping `SecureValues` a sibling list, identified only by key/description/required, is what lets `Spec` stay a clean, standard OpenAPI schema that any generic Kubernetes-style tooling can validate without knowing anything about Grafana's secret-handling conventions.

This is the same artifact, the same conversion function, and the same schema document — App Platform's resource API and `grafana-plugin-sdk-go`'s settings validation are two consumers of one conversion output, not two separately maintained schemas.

### 6.3 Why `dsconfig` rather than authoring the App Platform settings artifact directly

A reasonable question, given Section 6.2's framing, is why this RFC proposes an authoring layer (`dsconfig`'s `Schema`/`ConfigField` types) at all, rather than having each plugin author write `Settings{Spec, SecureValues}` — or the full `PluginSchema` envelope — directly, by hand, per plugin, in the shape Section 6.1 already shows. This section states why directly, rather than leaving it implicit in the rest of the design.

**Authoring the App Platform artifact directly reproduces, rather than avoids, the core problem this RFC exists to solve.** Plain OpenAPI/JSON Schema has no native vocabulary for the concerns specific to Grafana datasource configuration:

- **No stable, storage-independent field reference.** OpenAPI has no equivalent of `id` (Section 5.3) — the only way to refer to a property is its literal path in the `spec` tree (`properties.jsonData.properties.authMethod`). A hand-authored document expressing "when this field changes, set these other two fields" (`dsconfig`'s `effects`, Section 5.7.1) would have to reference fields by that path, which breaks the moment a property is renamed or moved between `root` and `jsonData` — exactly the kind of silent drift Section 2.1 describes, recurring at the authoring layer instead of the configuration layer.
- **No declared relationship between `Spec` and `SecureValues`.** Nothing in plain OpenAPI says "if `authMethod` is `oauth2`, then the `clientSecret` entry in `SecureValues` becomes required." A plugin author hand-authoring this relationship has to invent a vendor-extension convention for it from scratch, per plugin, with no shared convention — reproducing Section 2.3's "no shared vocabulary for recurring configuration concerns" one layer up, in the conditional-expression syntax instead of in field names.
- **No representation for legacy storage patterns.** The `indexedPair` convention (Section 5.5.1) — numbered key/value pairs split across `jsonData` and `secureJsonData` — has no OpenAPI representation at all. A plugin author hand-authoring the `Spec` for a plugin using this convention must either describe the array shape while losing the fact that its values are secret (reproducing, by hand, the exact correctness gap recorded in Section 10.3), or invent a bespoke vendor extension for it with no guarantee any other plugin author solves it the same way.
- **No shared field-set authoring.** TLS, basic auth, and timeout fields recur across a large fraction of the datasource plugin catalog. `dsconfig` schemas can share these via code-level helpers; a hand-authored OpenAPI document has no such mechanism, so every plugin's `Spec` re-declares the identical fields from scratch, and a correction to that shared shape must be propagated by hand to every plugin's copy.
- **No structural self-check before the artifact is trusted.** `Schema.Validate()` (Section 5, throughout) catches a missing required field, a dangling reference, or — as of this RFC's current revision — a malformed `id`, before conversion ever runs. A hand-written `Spec` document has no equivalent; a malformed or internally inconsistent document is only discovered when something downstream tries to use it, reproducing Section 2.2's deferred-error-surfacing problem at the authoring layer.
- **No home for metadata the App Platform artifact was never meant to carry.** `groups`, `instructions`, and `ui` hints (Sections 5.8–5.9) have no place in a `Spec` schema and structurally should not — broadening `Spec` to carry them would compromise the property (a clean, minimal, standard OpenAPI document) that makes it useful to App Platform's generic Kubernetes-style tooling in the first place (Section 6.2). A plugin author who only authors the App Platform artifact directly has, by construction, nowhere to put this information and no reason to produce it — which directly forecloses the schema-driven configuration page (Section 11.1) and Grafana Assistant (Section 7) use cases this RFC is also motivated by, not merely makes them harder.

**Extending the App Platform artifact itself with vendor extensions to carry this metadata was considered and is deliberately not the primary mechanism this RFC proposes.** It is technically possible — `dsconfig`'s own `x-dsconfig-depends-on`/`x-dsconfig-required-when`/`x-dsconfig-disabled-when` extensions (Section 6.1) already demonstrate the mechanism working for a narrow, specific purpose. But generalizing it to carry `id`, `groups`, `ui`, and `instructions` wholesale would (a) still require an authoring layer to generate those extensions correctly, so it does not eliminate the need for `dsconfig` or something like it; (b) scatter a document's metadata across per-property extensions at arbitrary nesting depth inside someone else's standard document, which is harder to validate and harder for a stable cross-field reference scheme to work against than a flat `id`-keyed field list; and (c) burden every generic consumer of the `Spec` schema — including consumers with no interest in Grafana-specific UI metadata — with parsing past it. This RFC's position is that `x-dsconfig-*` extensions remain appropriate for the few conditional-behavior expressions already carried this way, and that UI/layout/assistant metadata belongs in the `dsconfig` source document, with the App Platform artifact remaining a deliberately narrow projection of it rather than becoming the document of record for everything.

**Net:** the App Platform artifact (`Spec`/`SecureValues`, within the `PluginSchema` envelope) is the correct _output_. `dsconfig` is what makes authoring that output tractable, keeps it in sync with what `grafana-plugin-sdk-go` and a config editor also need from the same source, and is the only one of the three options considered (hand-authoring the App Platform artifact directly; smuggling all metadata into it via vendor extensions; a separate `dsconfig` source with the App Platform artifact as a generated projection) that doesn't either reproduce this RFC's founding problems one layer up or compromise the artifact's primary, external-facing purpose.

### 6.4 Resolving field identity and values back out of a real payload

Sections 5.3 and 6.1 establish `id` as the stable reference every cross-field mechanism in this schema uses, and establish that `Spec`/`SecureValues` describe shape, not stored values. The companion reference implementations (`SCHEMA-V1.go`, `SCHEMA-V1.ts`) provide the read-side counterpart needed by any consumer — a config editor, a runtime validator, Grafana Assistant — that is handed an `id` and needs to resolve it against a real configuration payload rather than against the schema alone:

- **`FieldByID`/`fieldById`** resolves an `id` to its `ConfigField` definition.
- **`ValueByID`/`valueById`** resolves an `id` to its actual configured value in a given payload, for fields with a direct storage mapping.
- **`ResolveIndexedPairs`/`resolveIndexedPairs`** and **`ResolveIndexedPairsAsMap`/`resolveIndexedPairsAsMap`** resolve an `indexedPair`-mapped field (Section 5.5.1) into, respectively, an order- and duplicate-preserving array or a gap-tolerant flat map — two different, deliberately non-unifiable trade-offs documented in full in Section 10.

These functions are read-side utilities operating on real configuration data; they are not part of schema authoring or structural validation, and none of them evaluates `dependsOn`/`requiredWhen`/`disabledWhen`/`effects[].when` or a `computed` storage mapping, both of which remain explicitly out of scope (Section 10.1). They are, however, the concrete mechanism by which the `id`-format enforcement introduced in this revision (Section 10.6) earns its keep today, ahead of any future evaluator: every `id` these functions are asked to resolve is, by construction, already guaranteed safe for dotted-path resolution, because `Validate`/`validateSchema` would have rejected the schema otherwise.

## 7. Grafana Assistant Consumption Model

### 7.1 The problem this section addresses

A chat-driven assistant operating inside Grafana, asked to create or modify a datasource on a person's behalf, must answer several questions before it can act: what fields does this plugin require; which are optional; what does a valid value look like for each; which fields hold secrets and therefore cannot be read back once set; and what, if anything, must be true of one field given the value of another. Today, none of these questions has a structured answer — only plugin source code, which an assistant is not expected to parse, and prose documentation, which is not guaranteed to be accurate or complete.

The direct business consequence of this gap is turn count and failure rate: an assistant without a structured contract must either ask the person a long sequence of clarifying questions, guess and risk a failed connection, or decline to help with datasource configuration at all. Each of these outcomes works against the explicit goal of getting a working, validated datasource connection established as quickly as possible.

### 7.2 What `dsconfig` provides to this workflow

A `dsconfig` schema document is, by construction, sufficient for an assistant to:

- Enumerate every field required for a working connection, in the order a config editor would present them, using `fields`, `required`, and `groups`.
- Distinguish fields that must never be echoed back to the person once set (`target: secureJsonData`) from fields that may be displayed or confirmed.
- Validate a proposed value against `validations` before attempting to save it, surfacing a specific, field-attributed problem to the person rather than discovering a failure only after a connection attempt.
- Read `description`, `label`, and `docURL` to explain, in natural language, what a field is for and where its value can be obtained, without needing a separate, hand-maintained prompt per plugin.
- Read `instructions` for guidance that is awkward to express as field-level metadata alone — for example, clarifying which of two superficially similar URL fields a particular plugin actually expects, or noting a non-obvious prerequisite for a given field's value.
- Resolve an `id` a person refers to by name, or an `id` taken from a `groups[].fieldRefs` entry, to its actual definition and — where the person is editing an existing datasource rather than creating a new one — its current configured value, via `FieldByID`/`ValueByID` (Section 6.4), without needing to re-implement that resolution independently of the reference implementations.

### 7.3 Why this directly serves the "fast, successful connection" goal

Every field-level mechanism in Section 5 — typed values, an explicit secure/non-secure distinction, a single data contract independent of presentation, structured multi-field write declarations — exists in service of one outcome for this workflow specifically: an assistant that can construct a syntactically and semantically valid configuration payload **on its first attempt**, for the fields the schema can fully describe, rather than iterating through failed `CheckHealth` calls. This is the same precondition that Section 6's HTTP-client-derivation motivation depends on (Section 10.10): a schema expressive enough to support reliable automated client construction is, by the same property, expressive enough to support reliable automated configuration construction by an assistant.

### 7.4 Current limitations specific to this workflow

The `instructions` mechanism (Section 5.9) is, at this release, an unstructured, untyped free-text list — sufficient for a human-authored hint but not yet validated, versioned, or scoped to a specific consumer or interaction stage. Conditional requirements expressed via `requiredWhen`/`dependsOn` are not evaluated by any component (Section 10.1), meaning an assistant today must independently re-implement the logic those expressions describe, in its own reasoning, rather than rely on a shared evaluator — this is a direct, near-term cost of Section 10.1's scope boundary as it applies to this specific workflow, and is called out here in addition to its general treatment in Section 10.

## 8. Design Decisions, Advantages, and Disadvantages, With Mitigations

This section states, without qualification, what this design does well, what it does poorly, and what is proposed to address each weakness. Items are not ordered by severity within either subsection; Section 8.3 maps every disadvantage to its mitigation and the section of this RFC where that mitigation is detailed.

### 8.1 Design decisions and their rationale

**D1 — `id` and `key` are two separate properties, not one.** A field's globally unique reference and its locally significant storage key serve different audiences and have different constraints: the storage key is frequently constrained by backward compatibility with an already-persisted configuration shape, while the global reference is unconstrained and can be chosen for clarity. Collapsing these into one property would force every field's storage key to also serve as a globally unique, human-legible reference, which is not achievable for plugins whose existing storage keys are short or generic (e.g., many plugins use the storage key `"password"`; only a dedicated `id` can disambiguate which plugin's password field a cross-document reference means).

**D2 — `validations` is the data contract; `ui.options` is presentation only.** This separation exists so that any tool needing to know what values are acceptable — a validator, a documentation generator, a provisioning checker, Grafana Assistant — never needs to parse UI metadata to find that answer, and so that a config editor's presentation choices (which options to show, in what order, with what label) can change without altering the field's actual data contract.

**D3 — `effects` is structured; `dependsOn`/`requiredWhen`/`disabledWhen` are expression strings.** Multi-field write declarations (`effects[].set`) have a small, fully enumerable shape — a map of field ID to value — and are therefore expressed and validated structurally. Single-field conditional behavior (visibility, requiredness, disabled state) was judged, at the time of this schema's design, to potentially require more general boolean composition across fields, and was therefore reserved as an expression string pending evidence of what conditions actually recur in practice (Section 10.1). This is a deliberate asymmetry, not an oversight: the one part of the conditional-behavior surface that could be cleanly structured, was.

**D4 — No general expression evaluator is included in this release.** A sandboxed, cross-language expression evaluator (for the CEL-flavored strings this schema stores) is a substantial, ongoing engineering and security-review commitment. Building it before concrete evidence exists of what conditional logic actually recurs across the plugin catalog risks building a capability broader than what is needed, while deferring it risks the specific gap described in Section 10.1. This release defers; Section 11.4 records the criteria under which building it would become justified.

**D5 — `tags` is deliberately unstructured.** `tags` exists for loose, human-readable conventions that do not warrant a validated mechanism (for example, a free-text note that a field is driven by a particular selector). It is not promoted to a validated mechanism in this release because doing so prematurely — before a concrete need is identified — risks the same problem `id`'s formerly entirely unenforced convention illustrated before this revision's partial enforcement (Section 10.6): a mechanism that looks load-bearing but silently is not. The fix applied to `id` (structural format and prefix-collision enforcement, not full convention enforcement) is itself an instance of the right level of intervention — enforce what's actually load-bearing, leave stylistic convention alone — and is the model this RFC would apply to `tags` too, if and when `tags` is ever made load-bearing for something.

**D6 — Storage mapping (`storage`) is described independently of whether it is executed.** `indexedPair` and `computed` are validated structurally in this release even though no runtime engine yet executes either against real data (Section 10.2). This ordering — describe first, execute later — was chosen so that schema authors can record an accurate, complete description of a plugin's existing legacy storage convention today, without that description's value being contingent on a runtime engine that does not yet exist.

### 8.2 Advantages

- **Additive by construction.** No existing plugin, no existing stored configuration, and no existing provisioning file requires any change as a precondition of this RFC's adoption (Section 7.1–7.2 compatibility, Sections 5.1, 4.4).
- **One schema serves both the plugin SDK and the App Platform resource API.** A single authored document produces, via one deterministic conversion, the specification both `grafana-plugin-sdk-go` and App Platform's Kubernetes-style resource machinery require (Section 6).
- **Separates data contract from presentation cleanly**, so that validation tooling, documentation generation, and Grafana Assistant never need to interpret UI-specific metadata to determine what is actually required (Section 5.6, D2).
- **Structurally validated today**, not merely a documentation convention: missing required properties, unknown enum values, and dangling field-ID references are caught at schema-authoring time, before any of this schema's other consumers ever see the document (Sections 5.2–5.9, throughout).
- **Directly serves the Grafana Assistant and HTTP-client-derivation goals from its first version**, without requiring a breaking change to reach either: every field is already typed, located, and validated in a way that a future client-derivation or assistant-integration layer can consume incrementally as it is built (Sections 6, 7).

### 8.3 Disadvantages, mapped to mitigations

| #   | Disadvantage                                                                                                                                                                                                                                                                                                  | Mitigation                                                                                                                                                                                                                                                                                                                                                                                                                                                                               | Where addressed                    |
| --- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------- |
| 1   | Conditional-behavior expression strings (`dependsOn`, `requiredWhen`, `disabledWhen`, `overrides[].when`, `storage.computed`, `custom` validation) are stored but never evaluated; a malformed expression is undetectable until a future evaluator exists.                                                    | Continue using the structured, already-validated `effects` mechanism wherever a need is expressible as a multi-field write; scope and build a deliberately narrow evaluator only if concrete cross-plugin evidence justifies it.                                                                                                                                                                                                                                                         | Section 10.1; Section 11.4; D3, D4 |
| 2   | `indexedPair` and `computed` storage mappings are validated structurally but not executed against real configuration data.                                                                                                                                                                                    | Build a dedicated runtime resolution function as a follow-on proposal; this is a pure additive extension requiring no change to any already-published schema document.                                                                                                                                                                                                                                                                                                                   | Section 10.2; Section 11.3         |
| 3   | The SDK conversion does not propagate secure-value status through `indexedPair` mappings, so a generated specification can fail to indicate that certain array values are secret.                                                                                                                             | Treat as the highest-priority fix among all limitations; correct the conversion function (an internal implementation fix, not a schema shape change) as part of the same follow-on work as Mitigation 2.                                                                                                                                                                                                                                                                                 | Section 10.3                       |
| 4   | No field carries a semantic role independent of its name, limiting any automated consumer's ability to recognize that two differently-named fields across plugins mean the same thing — directly limiting how reliably an HTTP client or an assistant's field-population logic can generalize across plugins. | Reserve, in a future minor schema version, an additive, optional field-level property carrying a semantic role from a fixed vocabulary, resolved with a zero-effort default-name lookup for already-convergent field names (e.g., TLS fields) and an explicit override for the rest; no existing schema document requires any change to remain valid once this lands.                                                                                                                    | Section 10.4; Section 11.7         |
| 5   | Authentication representation has no canonical shape; some plugins use an explicit discriminator field, others use independent boolean flags that are not mutually exclusive, with no schema-level distinction between "how Grafana authenticates" and "whether to forward the requesting user's identity."   | Address in the same future role-vocabulary work as Mitigation 4, adding an explicit conflict-declaration mechanism for incompatible boolean combinations; do not attempt a retroactive standardization of existing plugins' field shapes.                                                                                                                                                                                                                                                | Section 10.4; Section 11.7, 11.8   |
| 6   | `section` nesting supports only one level, forcing recursive configuration shapes to be manually flattened to a bounded depth.                                                                                                                                                                                | Document the current constraint precisely (this RFC); scope a deeper-nesting or first-class nested-object field type only if a real plugin's recursion depth is shown to exceed what manual flattening can reasonably support.                                                                                                                                                                                                                                                           | Section 10.5                       |
| 7   | `id` format was, at the time this mitigation was first scoped, recommended but not enforced, risking the same convention-drift problem this schema exists to prevent, one layer up.                                                                                                                           | **Closed in this revision.** `Validate`/`validateSchema` now reject an `id` segment outside `[A-Za-z_][A-Za-z0-9_]*` and reject one `id` being a strict dotted-path prefix of another, closing the specific risk to future dotted-path expression resolution. The broader, purely stylistic hierarchical-by-meaning convention remains unenforced by design (Section 10.6) — that is a documentation convention, not a correctness risk, and is not in scope for structural enforcement. | Section 10.6                       |
| 8   | No representation exists for a plugin requiring more than one independent connection within a single datasource instance.                                                                                                                                                                                     | Scope as a distinct, explicitly additive future proposal (an optional, schema-level grouping construct with full backward compatibility for the overwhelming majority of single-connection plugins, which would see no change at all).                                                                                                                                                                                                                                                   | Section 10.7; Section 11.9         |
| 9   | No runtime function constructs an HTTP client, SQL connection, or any other object from a schema and a real configuration payload, despite this being a stated driver for the schema's design.                                                                                                                | Sequence as explicit, separately-proposed follow-on work, scoped first to the narrowest, highest-confidence case (standard HTTP transport and authentication patterns) rather than attempting a universal client factory.                                                                                                                                                                                                                                                                | Section 10.10; Section 11.6        |
| 10  | Validation error-reporting behavior differs between the Go and TypeScript reference implementations (first-error-only vs. all-errors).                                                                                                                                                                        | Align both implementations' error-collection behavior as part of the same follow-on work that introduces runtime validation against real configuration data, since that is the natural point to unify this behavior across languages.                                                                                                                                                                                                                                                    | Section 10.11                      |
| 11  | `tags`, `examples`, `repeatable`, and field-level `pattern` are defined but not consumed by any component, suggesting speculative design ahead of an established need.                                                                                                                                        | Retain as accepted, forward-compatible, non-breaking metadata; do not promote any of them to a validated mechanism without a concrete, demonstrated consumer first (see D5).                                                                                                                                                                                                                                                                                                             | Section 10.8, 10.9                 |
| 12  | `schemaVersion` is required but no component branches on its value; there is no defined migration path for a future breaking schema change.                                                                                                                                                                   | Define a version-migration policy as part of any future change to this specification that is not purely additive; not required for this release, since this release introduces no breaking change to migrate from.                                                                                                                                                                                                                                                                       | Section 10.12                      |
| 13  | The indexed-pair lookup utilities introduced in this revision (`ResolveIndexedPairs`/`ResolveIndexedPairsAsMap`) infer pair-role by item-field position rather than declaration, and the map-shaped variant collapses duplicate names and conflates a missing value with a genuinely empty one.               | Document both trade-offs precisely (done, this revision); add an explicit pair-role marker to item fields used inside an `indexedPair` mapping as a follow-on, additive schema enhancement if a real schema's authored item-field order is ever found to be wrong in practice.                                                                                                                                                                                                           | Section 10.13, 10.14               |

## 9. Compatibility

### 9.1 Backward compatibility

This RFC introduces no change to Grafana's persisted datasource configuration format. A plugin may adopt `dsconfig` by authoring a schema document describing its existing configuration shape exactly as it is currently stored; no change to stored configuration, to the plugin's existing `ConfigEditor`, or to its existing backend parsing code is required as a condition of adoption.

### 9.2 Adoption is opt-in and additive

Authoring a `dsconfig` schema for a plugin is independent of, and does not interfere with, that plugin's continued operation in the absence of the schema. A plugin without a `dsconfig` schema continues to function exactly as it does today. This RFC does not propose a deadline or enforcement mechanism by which existing plugins must adopt the schema.

### 9.3 Forward compatibility

Every disadvantage and mitigation recorded in Section 8.3 is, by explicit design, addressable without breaking any schema document that is valid under this RFC's initial version. No mitigation in that table requires a previously published schema document to be rewritten as a condition of a future enhancement landing.

## 10. Known Limitations

The following are explicit, acknowledged limitations of the design described in this RFC. They are not omissions discovered after the fact; they represent scope boundaries deliberately drawn for the initial release. This section is reproduced, in substance, in the "Known limitations" section of the companion `schema-v1.md` and in the corresponding comments in `SCHEMA-V1.go` and `SCHEMA-V1.ts`, so that a reader of any one artifact sees the identical, current list.

### 10.1 No expression evaluation

`dependsOn`, `requiredWhen`, `disabledWhen`, `overrides[].when`, `effects[].when`, `storage.computed.read`/`write`, and `validations[].type: custom`'s `expression` are all stored as strings in a CEL-like syntax. None is parsed beyond basic presence validation. No component in this release evaluates these expressions against real configuration data. A schema document containing a syntactically malformed expression string passes schema validation; the malformation is not detected until a future evaluation engine exists.

### 10.2 No runtime engine for storage mappings

The `indexedPair` and `computed` storage mapping types are validated structurally but have no associated runtime function operating against actual configuration data. No code in this release expands a declared `indexedPair` mapping into resolved key/value pairs from a real payload, and no code evaluates a `computed` mapping's `read` or `write` expression.

### 10.3 Incomplete secure-value propagation through storage mappings

The SDK conversion routes fields to the `secureValues` list based solely on each field's own `target`. For an `indexedPair`-mapped array field whose value pattern targets `secureJsonData`, the produced specification represents the field as an ordinary array property with no indication that its values are secret.

### 10.4 No semantic role independent of field name

Two fields with identical meaning but different names across plugins (e.g., `apiUrl`, `baseURL`, `endpoint`, all meaning "base URL") cannot be recognized as equivalent by any automated consumer without that consumer hard-coding plugin-specific name lists. This limits the reliability of any future automated HTTP client derivation or assistant-driven field population that must generalize across plugins rather than being written per plugin.

### 10.5 `section` nesting is limited to one level

Recursive configuration shapes must be flattened to a bounded, known depth using sequential `section` declarations — a workaround that depends on actual recursion depth in practice remaining small. Additionally, `section`, while implemented, is not described in the published reference documentation at the time of this RFC's predecessor draft and is corrected as part of this version's companion `schema-v1.md`.

### 10.6 `id` format is now partially enforced

As of this revision, every `id` is checked against two rules at validation time: each dot-separated segment must match `^[A-Za-z_][A-Za-z0-9_]*$`, and no `id` may be a strict dotted-path prefix of another `id` in the same document. This closes the specific failure mode of an `id` that would silently break a future expression evaluator's dotted-path resolution of `dependsOn`/`requiredWhen`/`disabledWhen`/`effects[].when` (Section 10.1) — an `id` containing a character outside this set, or one that is ambiguous as a prefix of another, is now rejected immediately at schema-validation time rather than remaining undetected until some future evaluator is run against it. This does **not** enforce the recommended hierarchical-by-meaning naming convention beyond the two structural rules above, and it does not, by itself, evaluate any expression. A schema document that validated successfully under a prior revision of this specification and happens to violate either new rule will now fail validation; this is treated as a deliberate, newly-enforced behavior change within the `v1` schema version rather than a `v1`-to-`v2` migration, consistent with this being a tightening of previously unenforced behavior rather than a change to any field's shape.

### 10.7 No representation for multi-connection plugins

A plugin requiring more than one logically independent connection within a single datasource instance has no first-class representation; nothing in the schema distinguishes which fields belong to which logical connection.

### 10.8 Unused metadata fields

`tags`, `examples`, and `pattern`/`repeatable` are defined on `ConfigField` but are not read or acted upon by any component described in this RFC.

### 10.9 Override mechanism does not affect SDK conversion

The `overrides` property is accepted and structurally validated but is not consulted by the SDK conversion described in Section 6. A field's overrides have no effect on the produced specification in this release.

### 10.10 No automatic derivation of runtime objects

The schema, including a fully authored and validated one, does not produce an HTTP client, a database connection string, or any other runtime object in this release, notwithstanding this being one of the two primary drivers motivating the schema's design (Section 1, Driver 2; Section 6.2). Consuming a schema to construct such objects requires hand-written, plugin-specific code today, exactly as it does in the absence of this RFC.

### 10.11 Validation error reporting is inconsistent across language implementations

The Go reference implementation's structural validation function returns upon the first validation error encountered. The TypeScript reference implementation collects and returns all validation errors found in a single pass. A schema author validating the same malformed document with each implementation observes different error-reporting behavior.

### 10.12 No schema-version migration mechanism

`schemaVersion` is a required field on every schema document, but no code in this release inspects its value to apply version-specific parsing, validation, or compatibility behavior.

### 10.13 Indexed-pair resolution infers pair-role by position, not by declaration

`ResolveIndexedPairs` and `ResolveIndexedPairsAsMap` (Section 6.4) assume a field's first declared item field is the pair's "name" and its second is its "value," because neither `ConfigField` nor its item schema carries an explicit pair-role tag today. A schema declaring these two item fields in the opposite order produces silently swapped results, with no validation error at any point — neither at schema-validation time nor at resolution time.

### 10.14 Indexed-pair-as-map resolution collapses duplicate names and value absence

`ResolveIndexedPairsAsMap` resolves two distinct stored pairs sharing the same name into a single map entry, with the underlying language's map/object iteration order — and therefore which pair's value survives — unspecified. It also represents a name whose corresponding value is absent as an empty string, indistinguishable from a value that was genuinely configured as an empty string. `ResolveIndexedPairs` has neither limitation, at the cost of not being gap-tolerant the way `ResolveIndexedPairsAsMap` is (Section 10.2); the two functions exist as a deliberate, non-unifiable trade-off rather than one superseding the other.

### 10.15 The lookup and resolution utilities cannot read a live secret value

`ValueByID`, `ResolveIndexedPairs`, and `ResolveIndexedPairsAsMap` (Section 6.4) all produce the same "value absent" result for a `secureJsonData`-targeted field whether that field was never configured or whether it is simply write-only and therefore unreadable post-save — Grafana's own API never returns a previously saved secret value, consistent with Section 5.1's storage-target table. These functions behave as documented against a schema's own example payloads, or any payload under direct, local construction that genuinely embeds secret values; they do not, and cannot, behave equivalently against a live, already-saved datasource's settings for any field targeting `secureJsonData`.

## 11. Future Work

The following are identified as natural extensions of this RFC's scope, explicitly deferred from the initial release pending separate proposals. Each item is cross-referenced to the limitation it addresses, per Section 8.3.

### 11.1 Schema-driven configuration pages

A rendering engine constructing a complete datasource configuration page directly from a schema's `fields`, `ui`, and `groups` metadata, eliminating the need for plugin authors to hand-write a `ConfigEditor` component for plugins whose configuration is fully describable by the schema. This is the natural consumer of metadata defined in Sections 5.8–5.9 but used by no renderer in this release.

### 11.2 Runtime validation against real configuration data

A function accepting a schema and an actual configuration payload, returning a structured list of validation failures — the runtime counterpart to the structural, schema-only validation defined in this RFC. A prerequisite for provisioning-time validation, pre-deploy CI checks, contract testing, and unifying the error-reporting inconsistency in Section 10.11.

### 11.3 Runtime resolution of storage mappings

**Partially delivered in this revision** for `indexedPair`: `ResolveIndexedPairs` and `ResolveIndexedPairsAsMap` (Section 6.4) now expand a real `indexedPair` mapping against actual stored data, with the trade-offs recorded in Sections 10.13–10.15. Still open: an execution engine for the `computed` mapping type (Section 10.2's `read`/`write` expressions remain unevaluated), and correcting the SDK conversion itself so that an `indexedPair` field's secure-value status is reflected in the generated `Spec`/`SecureValues` output (Section 10.3) — the new resolver functions operate on real configuration data at read time and do not change what the static conversion produces.

### 11.4 Expression evaluation

A narrowly scoped evaluation engine for the conditional-behavior expressions identified in Section 10.1, to be pursued only if structured alternatives (such as the existing `effects` mechanism, or a future structured-condition object) prove insufficient for concrete, recurring cases — particularly genuine cross-field validation, which has no structured alternative today.

### 11.5 Schema-derived test fixture generation

Automatic generation of valid and invalid configuration fixtures from a schema document, for automated contract testing between a plugin's configuration editor and its backend validation logic.

### 11.6 Automatic derivation of network clients

Construction of HTTP clients, and where feasible, database connection parameters, directly from a schema and a validated configuration payload, for plugins whose configuration follows sufficiently standardized patterns. Directly addresses Section 10.10 and the second primary driver stated in Section 1.

### 11.7 Semantic field roles

An additive, optional field-level property carrying a field's meaning from a fixed vocabulary, independent of its name, resolved via a default-name lookup table for already-convergent conventions (e.g., TLS field names) with an explicit override for the remainder. Addresses Section 10.4.

### 11.8 Standardized authentication vocabulary

A common, cross-plugin vocabulary and conflict-declaration mechanism for authentication configuration, addressing Section 10.4's auth-specific manifestation, without retroactively standardizing existing plugins' field shapes.

### 11.9 Multi-connection plugin support

A first-class, additive mechanism for declaring that a plugin's configuration spans more than one logically independent connection, addressing Section 10.7, designed so that the overwhelming majority of single-connection plugins are entirely unaffected.

### 11.10 Plugin catalog adoption tooling

Tooling within the plugin catalog and review process to surface schema presence and conformance as part of plugin submission and review, including an adoption metric derived directly from the schema corpus itself.

### 11.11 Grafana Assistant tool integration

The concrete integration connecting Grafana Assistant's tool-use capability to a per-plugin `dsconfig` schema, building on the consumption model defined in Section 7 but not itself specified by this RFC.

### 11.12 App Platform resource type registration

The concrete registration of the datasource resource type within Grafana App Platform, consuming the conversion output defined in Section 6 but not itself specified by this RFC.

### 11.13 Explicit pair-role marker on indexed-pair item fields

An additive tag on an item field used inside an `indexedPair` mapping's `item.fields`, declaring explicitly whether that field is the pair's name or its value, removing `ResolveIndexedPairs`/`ResolveIndexedPairsAsMap`'s current reliance on item-field declaration order (Section 10.13). No existing schema document would need to change to remain valid; the marker would be optional, with position-based inference remaining the fallback for schemas that don't supply it.

## 12. Alternatives Considered

**Continue without a schema.** Rejected. This preserves every problem identified in Section 2 indefinitely, with the cost of resolution increasing as the plugin catalog and its contributor population grow, and leaves both the App Platform resource model and any Grafana Assistant datasource workflow with no structural artifact to build on.

**Adopt a generic, off-the-shelf JSON Schema dialect directly, without a Grafana-specific intermediate format.** Considered, but rejected for the initial release. Plain JSON Schema lacks native vocabulary for Grafana-specific concerns — the `root`/`jsonData`/`secureJsonData` storage split, the legacy indexed-pair pattern, and multi-field write declarations — all of which would require the same vendor-extension mechanisms this RFC already defines, while losing the more ergonomic `id`/`key`/`target` authoring surface purpose-built for this domain. The SDK conversion (Section 6) already produces an OpenAPI/JSON-Schema-compatible artifact as its output, preserving interoperability with generic tooling — including App Platform's resource machinery — without requiring schema authors to work directly in that more verbose format.

**Require full expression evaluation in the initial release.** Considered, but rejected on the basis of the cross-language implementation and security-review burden of a sandboxed expression evaluator, weighed against the limited set of concrete cases observed in the current plugin catalog. Recorded as future work (Section 11.4).

**Design the App Platform resource schema and the plugin SDK settings schema as two separately authored artifacts.** Considered, but rejected: maintaining two independently authored descriptions of the same configuration surface reintroduces exactly the drift problem this RFC exists to eliminate (Section 2.1), one layer higher. A single schema document with one deterministic conversion function, as specified in Section 6, was chosen instead.

## 13. Rollout Plan

This RFC, upon acceptance, is implemented as a standalone Go and TypeScript package (the companion `SCHEMA-V1.go` and `SCHEMA-V1.ts` artifacts), consumable by any plugin on an opt-in basis. No change to Grafana core, to the plugin SDK's existing public interfaces, or to any existing plugin is required as a condition of this RFC's acceptance.

Subsequent, separately proposed work — runtime validation, storage-mapping execution, semantic field roles, network client derivation, App Platform resource type registration, and Grafana Assistant tool integration — is sequenced as follow-on proposals per Section 11, each scoped so that no prior step's adopters are required to make any change as a precondition of a later step landing.

## 14. References

- `schema-v1.md` — full reference documentation for the schema, including worked examples for every storage and modeling pattern described in Section 5, and the lookup/resolution utilities described in Section 6.4.
- `SCHEMA-V1.go` — Go reference implementation, including structural validation (now including `id` format enforcement, Section 10.6), SDK conversion, and the `FieldByID`/`ValueByID`/`ResolveIndexedPairs`/`ResolveIndexedPairsAsMap` lookup and resolution functions described in Section 6.4.
- `SCHEMA-V1.ts` — TypeScript reference implementation, structurally equivalent to `SCHEMA-V1.go`, including its own `fieldById`/`valueById`/`resolveIndexedPairs`/`resolveIndexedPairsAsMap` mirror of the same functions.
- `SCHEMA-V1.json` — formal JSON Schema (draft 2020-12) describing the wire format defined in Section 5, including the `fieldId`/`fieldIdSegment` pattern constraints introduced in this revision, suitable for direct use by non-Go, non-TypeScript tooling.
