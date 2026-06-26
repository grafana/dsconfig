# dsconfig v2 Extensions — Reference Documentation

This document covers only what's new in v2. For everything else — the field model, storage targets, validation rules, indexed-pair storage, the lookup utilities, and the full list of v1 limitations — see `SCHEMA-V1.md`, which this document does not repeat or restate.

## Implementation note — read this first

v2's three new capabilities (scopes, roles, pair-role marking) are implemented in this deliverable as a **separate side-table document** — `SchemaV2Extensions` (Go: `SCHEMA-V2.go`; TypeScript: `SCHEMA-V2.ts`; wire format: `SCHEMA-V2.json`) — keyed by the same field `id` a v1 `Schema` document already uses, rather than as new properties added directly onto a v1 `configField` object.

This is a deliberate constraint of how this particular set of artifacts was produced — specifically, to avoid modifying `SCHEMA-V1.go`/`SCHEMA-V1.ts`/`SCHEMA-V1.json`/`SCHEMA-V1.md` at all. **It is not the proposed final shape.** The actual target shape, described in GRF-RFC-0043, is for `scopes`, `role`, `roleConflicts`, and `pairRole` to become four new, optional, `omitempty`-style properties declared directly on a v1 `configField` object — exactly the same kind of additive struct evolution every v1 property was already designed to support. Once that change is made for real, `SchemaV2Extensions` as a separate document disappears entirely, and a schema author writes one document instead of two.

Until then, treat `SchemaV2Extensions` as a faithful, behavior-preserving stand-in: every function and validation rule described below behaves exactly as it would if these were native `configField` properties — the side table just means a schema author authoring v2 capabilities today writes a second, parallel document and is responsible for keeping it in sync with the first by `id`, a real authoring cost that the eventual real implementation does not have.

## Why these three capabilities, and why now

Each of the three traces directly to a numbered item in `SCHEMA-V1.md`'s "Known limitations" list:

| v2 capability        | Closes                                                                                                                                  |
| -------------------- | --------------------------------------------------------------------------------------------------------------------------------------- |
| Scopes               | Limitation 7 — no representation for a plugin with more than one independent connection within a single datasource instance             |
| Role / RoleConflicts | Limitation 5 — no field carries a semantic role independent of its name                                                                 |
| PairRole             | Limitations 13 and 14 — indexed-pair resolution infers name/value by item-field declaration order, with no way to declare it explicitly |

## Scopes

### The problem

A plugin like AppDynamics has two genuinely independent connections — a Controller API and an End User Monitoring (EUM) API — each with its own URL, auth, and TLS settings, inside one datasource instance. v1 has no way to say "this field belongs to the Controller connection, that one belongs to EUM" — every field in a v1 schema is implicitly global.

### The shape

```json
{
  "scopeDefs": [
    { "id": "controller", "label": "Controller API" },
    { "id": "eum", "label": "End User Monitoring API" }
  ],
  "fields": {
    "controller.url": { "scopes": ["controller"] },
    "eum.url": { "scopes": ["eum"] },
    "shared.timeout": {}
  }
}
```

`controller.url` and `eum.url` are scoped to one connection each. `shared.timeout` has an empty extension entry — no `scopes` set — which means it applies to **both**, and would automatically apply to a third connection added in some future revision of the schema too.

### The omission rule, precisely

A field's `scopes` being absent or empty means "shared across every scope declared in `scopeDefs`, including any scope declared later." This is **not** the same as listing every currently-declared scope id explicitly — that's rejected by validation, specifically because it would not extend to a scope added later, and allowing both spellings of "all scopes" to coexist would let them silently diverge from each other over time. There is exactly one correct way to say "this field is universal": omit `scopes` entirely.

### What you get: `FieldsForScope`

```go
fields, err := dsconfig.FieldsForScope(schema, ext, "controller")
// fields == every field scoped to "controller", plus every field with no scopes at all
```

This is the function a future per-connection HTTP client builder, or a config editor rendering one connection's section of a form, would call to get exactly the fields relevant to one connection — without re-implementing the omission-means-shared rule itself.

### What scopes do not yet do

`ToPluginSchemaSettings` (the v1 conversion to an App Platform-consumable artifact) is **unaware of scopes**. A multi-connection plugin's schema, even with `scopeDefs` and field-level `scopes` fully declared, still produces exactly one `Settings{Spec, SecureValues}` pair — the same shape v1 always produced. Scopes can fully _describe_ a multi-connection plugin today; nothing yet _derives a separate App Platform resource per connection_ from that description. This is the single largest gap between what v2 lets you say and what the existing conversion pipeline does with it — see "Known limitations," below.

## Semantic roles

### The problem

Two plugins might both have "the field that holds the TLS client certificate," but one calls it `tlsClientCert` and the other calls it `clientCertificate`. Nothing in v1 lets a consumer — a future HTTP client builder, or Grafana Assistant reasoning about a plugin it's never seen before — recognize that these two differently-named fields mean the same thing.

### The shape

```json
{
  "fields": {
    "secure.tlsClientCert": { "role": "tls.clientCert" },
    "secure.tlsClientKey": { "role": "tls.clientKey" },
    "jsonData.tlsAuth": {
      "role": "auth.basic.enabled",
      "roleConflicts": ["auth.awsSigV4.enabled"]
    }
  }
}
```

### The vocabulary

Role is a closed, fixed set — 20 values in this release, spanning `endpoint.*`, `transport.*`, `tls.*`, `auth.*`, `identity.*`, and `http.header.*`. The complete list lives in `SCHEMA-V2.json`'s `role` enum, and in `SCHEMA-V2.go`/`SCHEMA-V2.ts`'s `Role` type — all three are kept in exact agreement (verified directly, not just asserted, as part of producing this release). A `role` value outside this set is rejected at validation time.

The vocabulary is closed deliberately: a plugin with a genuine semantic concept this version hasn't anticipated has no way to express it except leaving `role` unset for that field. Growing the vocabulary is additive (new constants, new enum values) and doesn't require a `v3` — but it does require an actual revision of this package, not something a schema author can extend unilaterally.

### `roleConflicts`, and what it does and doesn't guarantee

`roleConflicts` declares which other roles must never be simultaneously active alongside this field's role. This exists specifically for the Pattern-B auth case discussed at length in this package's design history: independent boolean auth flags (`basicAuth`, `sigV4Auth`, `oauthPassThru`, ...) that aren't mutually exclusive in storage but might be mutually exclusive in meaning for a given plugin's API.

What's actually checked, today: that every role named in `roleConflicts` is a real vocabulary value, and that — as a static, structural fact about the schema — no two fields within the same effective scope (per `FieldsForScope`) carry roles that conflict with each other. This is `ValidateRoleConflicts`/`validateRoleConflicts`.

What's **not** checked: whether two conflicting-by-declaration fields are actually _both set to a truthy value_ in any real configuration payload. `roleConflicts` is validated against the schema's shape, not against live data — there's no runtime enforcement here yet, the same posture v1's `StorageMapping` had before `ResolveIndexedPairs` gave it a runtime, and the same posture `roleConflicts` is in until something equivalent is built for it.

## Pair-role marking

### The problem

`ResolveIndexedPairs`/`ResolveIndexedPairsAsMap` (v1) assume the _first_ declared item field in an `indexedPair` mapping's `item.fields` is the pair's name, and the _second_ is its value. Nothing stops a schema author from declaring them in the opposite order, and nothing catches it — the result is silently swapped name/value pairs.

### The shape

```json
{
  "fields": {
    "httpHeaders.item.name": { "pairRole": "key" },
    "httpHeaders.item.value": { "pairRole": "value" }
  }
}
```

### The rule

If either item field in a pair declares `pairRole`, the **other must too** — declaring it on only one is rejected as ambiguous, rather than silently falling back to positional inference for just the undeclared half. If **neither** item field declares `pairRole`, v1's positional inference applies exactly as before — a schema with no `pairRole` extensions anywhere behaves identically under v2 to how it behaved under v1.

### What you get: `ResolveIndexedPairsV2` / `ResolveIndexedPairsAsMapV2`

These are new functions, not changes to v1's `ResolveIndexedPairs`/`ResolveIndexedPairsAsMap` — which remain completely untouched and continue to use pure positional inference with no awareness of `pairRole`, ever. Call the `V2` versions when you have a `SchemaV2Extensions` document available and want `pairRole`-aware resolution; call the v1 versions when you don't.

```go
pairs, err := dsconfig.ResolveIndexedPairsV2(headersField, ext, jsonData, secureJsonData)
```

## Validation: `ValidateV2`

`ValidateV2(schema, ext)` performs every v2-specific structural check — scope reference validation, the omission-vs-full-list staleness rule, role vocabulary validation, role-conflict structural consistency, and pair-role contradiction detection. It is a **separate entry point from v1's `Validate`/`validateSchema`**, called in addition to it, never instead of it:

```go
if err := schema.Validate(); err != nil {
    return err // v1 structural validation, unchanged
}
if err := dsconfig.ValidateV2(schema, ext); err != nil {
    return err // v2-specific checks
}
```

This separation exists so a v1 document — one with no corresponding `SchemaV2Extensions` at all — is never subjected to v2-only checks it was never written against. Only call `ValidateV2` for a document that actually has v2 extensions to validate, even if that extensions document is empty.

## Known limitations (v2)

These are additional to, not a replacement for, `SCHEMA-V1.md`'s own "Known limitations" list — every v1 limitation v2 doesn't close remains exactly as documented there.

1. **The side-table representation itself.** `SchemaV2Extensions` is a separate document from its companion `Schema`, kept in sync by `id`, by hand or by tooling. Nothing in this release enforces that every `id` in a `Schema` has a corresponding entry (even an empty one) in the extensions document, or vice versa, beyond what `ValidateV2` happens to check while walking the field tree it's given. The proposed upstream shape (properties directly on `configField`) does not have this two-document synchronization burden at all.
2. **No multi-connection-aware App Platform artifact.** `ToPluginSchemaSettings` still emits exactly one `Settings{Spec, SecureValues}` pair per schema document, regardless of how many scopes are declared.
3. **`roleConflicts` is structural, not runtime-enforced.** It catches a schema that could never be satisfied; it does not catch a real configuration where two conflicting-by-declaration fields are both actually active at once.
4. **A valid role doesn't mean an appropriate one.** Nothing cross-checks `role: "tls.clientCert"` against the field's `valueType` or `target` actually being suitable for that role — a boolean field could carry that role and nothing would object.
5. **The role vocabulary is closed and fixed by this package version.** A schema author cannot extend it; a field whose meaning isn't in the vocabulary simply has no `role`.
6. **`pairRole` is all-or-nothing per pair.** Declaring it on one item field but not its partner is rejected, not silently completed by inference.
7. **v1's resolver functions are completely unaffected.** `ResolveIndexedPairs`/`ResolveIndexedPairsAsMap` never look at `pairRole`, ever — use the `V2` functions for that.

## Cross-reference

| Concept                                                         | v1 source                                                        | v2 source                                                        |
| --------------------------------------------------------------- | ---------------------------------------------------------------- | ---------------------------------------------------------------- |
| Field identity, storage targets, validation rules               | `SCHEMA-V1.md`, `SCHEMA-V1.go`, `SCHEMA-V1.ts`, `SCHEMA-V1.json` | —                                                                |
| Lookup by id, indexed-pair resolution                           | `SCHEMA-V1.md` ("Reading values back out by id")                 | `ResolveIndexedPairsV2`/`ResolveIndexedPairsAsMapV2` extend this |
| Multi-connection plugins                                        | Named as Known Limitation 7                                      | This document, "Scopes"                                          |
| Semantic field roles                                            | Named as Known Limitation 5                                      | This document, "Semantic roles"                                  |
| Pair-role inference                                             | Named as Known Limitations 13-14                                 | This document, "Pair-role marking"                               |
| Full design rationale, breaking-change analysis, migration path | —                                                                | GRF-RFC-0043                                                     |
