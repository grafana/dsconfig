# RFC-DSCONFIG-V2: `dsconfig` v2 — Multi-Connection Scopes, Semantic Field Roles, and Explicit Pair-Role Marking

| Field               | Value                                                                                                                                                                                            |
| ------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| RFC ID              | RFC-DSCONFIG-V2                                                                                                                                                                                  |
| Title               | `dsconfig` v2: Multi-Connection Scopes, Semantic Field Roles, and Explicit Pair-Role Marking                                                                                                     |
| Status              | Proposed                                                                                                                                                                                         |
| Amends              | RFC-DSCONFIG-V1 (`dsconfig` v1)                                                                                                                                                                  |
| Author              | Sriram                                                                                                                                                                                           |
| Component           | `grafana-plugin-sdk-go`, Grafana App Platform, Grafana Assistant                                                                                                                                 |
| Companion artifacts | `SCHEMA-V2.md`, `SCHEMA-V2.go`, `SCHEMA-V2.ts`, `SCHEMA-V2.json` — all new files, none modifying any RFC-DSCONFIG-V1 artifact (`SCHEMA-V1.md`, `SCHEMA-V1.go`, `SCHEMA-V1.ts`, `SCHEMA-V1.json`) |

---

## 1. Summary

This RFC proposes three additions to `dsconfig`, each closing a gap RFC-DSCONFIG-V1 named explicitly in its own "Known Limitations" section:

1. **Scopes** — a representation for datasource plugins with more than one independent connection inside a single instance (Limitation 7).
2. **Semantic field roles** (`role`, `roleConflicts`) — a fixed, closed vocabulary letting a field declare what it _means_, independent of what it's _named_ (Limitation 5).
3. **Explicit pair-role marking** (`pairRole`) — a way to declare which half of an `indexedPair` item schema is the name and which is the value, removing v1's reliance on declaration order (Limitations 13-14).

None of these changes anything about how Grafana stores datasource configuration. None requires an existing schema document to change to remain valid. The one genuinely structural consequence — and the reason this is a numbered RFC rather than a patch note — is that **scopes change what "the field set" means** for any consumer that builds something across a whole schema (a future HTTP client builder, most concretely), and that consequence needs to be named and sequenced deliberately rather than discovered later.

This RFC also discloses a specific, load-bearing constraint of its own reference implementation: because this round of work was produced under an explicit instruction not to modify any RFC-DSCONFIG-V1 artifact, the new capabilities are implemented as a **separate side-table document** (`SchemaV2Extensions`) rather than as new properties directly on `configField`. Section 5.5 states plainly why that's a workaround, not the proposed final shape, and what changes once a real version bump is allowed to touch `SCHEMA-V1.go`/`SCHEMA-V1.ts`/`SCHEMA-V1.json` directly.

---

## 2. Problem Statement

RFC-DSCONFIG-V1 shipped with a complete, honest accounting of what it deferred. Three of those deferrals have since hardened into specific, named blockers:

### 2.1 Multi-connection plugins are invisible to the schema

A plugin like AppDynamics has a Controller API and a separate EUM API, each independently configured — different URL, different auth, potentially different TLS settings — inside one datasource instance. v1's `ConfigField` has no way to express "this field belongs to connection A, that one belongs to connection B." Every field in a v1 schema is implicitly global. This isn't a corner case being theorized about — it's a real, named plugin shape that v1 cannot describe at all, and any future tooling built against v1's implicit "one schema, one connection" assumption will be silently wrong for this entire category of plugin.

### 2.2 A field's meaning is only ever guessable from its name

Two plugins' TLS client certificate fields might be named `tlsClientCert` and `clientCertificate` respectively. v1 has no mechanism for either field to declare "I am the TLS client certificate" independent of that name. This directly limits the two heaviest-weighted business drivers from RFC-DSCONFIG-V1's own Section 1: reliable automated HTTP client derivation, and Grafana Assistant generalizing across plugins it has never seen configured before. Both need to recognize a field's _meaning_; v1 only gives them a field's _name_.

### 2.3 Indexed-pair resolution can silently produce wrong results

v1's `ResolveIndexedPairs`/`ResolveIndexedPairsAsMap` infer which item field is a pair's name and which is its value by declaration order — first item field is the name, second is the value. A schema author who declares them in the opposite order gets silently swapped results, with no validation error at any point. This is a real, live correctness risk in already-shipped functionality, not a hypothetical.

## 3. Goals

1. Let a schema describe a plugin with more than one independent connection, with zero cost to the overwhelming majority of plugins that have exactly one.
2. Let a field declare a semantic role from a closed, versioned vocabulary, independent of its `key`.
3. Let a field declare which other roles must not be simultaneously active alongside its own.
4. Let an `indexedPair` item schema declare explicitly, rather than rely on declaration order, which item field is the pair's name and which is its value.
5. Make every one of the above fully optional and additive: a v1 schema document, used with no v2 capability, must remain valid and must behave identically to how it behaved under RFC-DSCONFIG-V1 alone.
6. State plainly, rather than gloss over, the one place this RFC's actual reference implementation diverges from its own proposed final shape, and why.

## 4. Non-Goals

1. This RFC does not make `ToPluginSchemaSettings` multi-connection-aware. A v2 schema can fully _describe_ a multi-connection plugin; producing a separate App Platform resource per connection from that description is out of scope here (see Section 10.2; Section 11).
2. This RFC does not implement runtime enforcement of `roleConflicts` against a real configuration payload — only structural, schema-shape validation (Section 10.3).
3. This RFC does not retroactively re-annotate any existing plugin's schema with roles or scopes. Adoption is opt-in, field by field, exactly as RFC-DSCONFIG-V1's own migration strategy established for v1.
4. This RFC does not implement expression evaluation. `dependsOn`/`requiredWhen`/`disabledWhen`/`effects[].when`/`storage.computed` remain exactly as unevaluated under v2 as they were under v1.
5. This RFC does not change anything about `root`/`jsonData`/`secureJsonData` storage, or about any of v1's existing types, functions, or validation behavior. Every RFC-DSCONFIG-V1 artifact is unmodified by this proposal's reference implementation.

## 5. Design

### 5.1 Scopes

A schema may declare a list of named scopes (`ScopeDef{ID, Label}`) at the document root. A field may declare which scope(s) it belongs to. **A field with no scope declaration is implicitly shared across every declared scope — including any scope declared in a future revision of the same schema.**

This last point is a deliberate semantic, not an implementation detail, and it's the reason this RFC rejects an alternative that looks superficially equivalent: a field listing every _currently_ declared scope id is **not** treated as equivalent to omitting the scope declaration entirely, and is rejected by validation. The reason is staleness — an explicit full list does not automatically extend to a scope added later, while omission does, and allowing both spellings of "applies everywhere" to coexist invites them to silently diverge from each other the moment a new scope is introduced. There is exactly one correct way to declare a field universal: omit the scope declaration.

```go
fields, err := dsconfig.FieldsForScope(schema, ext, "controller")
```

`FieldsForScope` returns every field scoped to the given scope, plus every field with no scope declaration at all — the function any future multi-connection-aware consumer (most concretely, a per-connection HTTP client builder) calls to get the field set relevant to one connection, without re-implementing the omission-means-shared rule itself.

**Why "scopes," not "services":** the design history behind this proposal considered, and rejected, a narrower `serviceId`/`serviceIds` pair of fields modeled directly on the AppDynamics case. The broader name and single-array-field shape were chosen once a second, structurally identical need — partitioning fields by deployment mode, not just by connection — surfaced during design: both are instances of "which configuration context does this field apply to," and a single generalized mechanism, rather than one bespoke field per axis, was judged the more honest design once a second real axis existed. "Scope" is deliberately a domain-neutral name for that mechanism.

### 5.2 Semantic roles

A field may declare a `role` from a fixed, closed vocabulary (`SCHEMA-V2.go`'s `Role` type; 20 values in this release, spanning `endpoint.*`, `transport.*`, `tls.*`, `auth.*`, `identity.*`, `http.header.*`). A `role` value outside this vocabulary is rejected at validation time — the vocabulary is closed by design, not extensible by a schema author, exactly as `valueType` and `target` already are in v1.

A field may also declare `roleConflicts`: roles that must not be simultaneously active alongside its own. This directly addresses the auth-pattern ambiguity surfaced repeatedly in this package's design history — Prometheus/Loki-style independent boolean auth flags (`basicAuth`, `sigV4Auth`, `oauthPassThru`, ...) are not mutually exclusive in storage, but might be mutually exclusive in meaning for a given plugin's actual API. `roleConflicts` lets a schema author declare that exclusivity explicitly.

`ValidateRoleConflicts` checks this structurally and unconditionally: every named role must be a real vocabulary value, and no two fields within the same effective scope (per `FieldsForScope`) may carry roles that conflict with each other — caught as a static, schema-shape fact, before any real configuration data is involved. What it does **not** do — stated here precisely so it isn't assumed — is check whether two conflicting-by-declaration fields are _both actually set to a truthy value_ in any real, live configuration payload. That's runtime enforcement, and it's out of scope for this RFC (Section 10.3), for the same reason v1's `StorageMapping` was accepted structurally for a full version before `ResolveIndexedPairs` gave it a runtime: describing a capability and executing it are treated as separable, sequenced pieces of work in this package's established discipline, not one inseparable deliverable.

### 5.3 Explicit pair-role marking

An item field inside an `indexedPair` mapping's item schema may declare `pairRole: "key"` or `pairRole: "value"`. When present, `ResolveIndexedPairsV2`/`ResolveIndexedPairsAsMapV2` (new functions; see Section 5.4) use this declaration directly instead of inferring it from item-field order.

The rule is strict by design: if **either** item field in a pair declares `pairRole`, the **other must too** — declaring it on only one is rejected as ambiguous, rather than silently falling back to positional inference for the undeclared half. Mixing explicit and positional inference for the two halves of the same pair was judged more likely to mask an authoring mistake than to accommodate a legitimate partial-adoption case. If **neither** item field declares `pairRole`, v1's positional inference applies exactly as it always has.

### 5.4 New resolver functions, not changes to existing ones

`ResolveIndexedPairsV2` and `ResolveIndexedPairsAsMapV2` are new functions. RFC-DSCONFIG-V1's `ResolveIndexedPairs` and `ResolveIndexedPairsAsMap` are **completely unmodified** and remain purely positional, with zero awareness of `pairRole`, permanently — a caller who wants `pairRole`-aware resolution calls the new functions; a caller who doesn't, or has no v2 extensions available, keeps calling the original ones exactly as before.

This pattern — add a new function rather than change an existing one's behavior — is the same discipline this RFC applies to validation (`ValidateV2` as a separate entry point from `Validate`, never folded into it) and is the load-bearing mechanism by which this entire proposal stays additive at the _behavior_ level, on top of being additive at the _schema shape_ level.

### 5.5 The side-table constraint, and what it is standing in for

This is the one place this RFC needs to be read carefully against what it actually ships versus what it proposes.

**What ships:** `SchemaV2Extensions`, a document separate from a v1 `Schema`, keyed by the same field `ID` v1 already treats as the stable cross-reference. A schema author wanting v2 capabilities today authors two documents and keeps them in sync by hand or by tooling.

**Why:** this round of work was produced under an explicit instruction not to modify any RFC-DSCONFIG-V1 artifact. Go does not allow a struct's field list to be split across two files — `ConfigField` is declared once, in `SCHEMA-V1.go`, and a separate file cannot add fields to it without editing that declaration. TypeScript's `ConfigField` interface could technically be extended via declaration merging without touching `SCHEMA-V1.ts`, but every existing v1 function (`fieldById`, `valueById`, `validateSchema`, and the rest) would still only ever see the plain, unextended shape unless duplicated against an extended type — the same practical cascade the Go side avoids by using a side table. For parity between the two language implementations — a property this package has maintained since v1 specifically so a schema author never gets different answers depending on which language validated a document — both reference implementations use the identical side-table approach.

**What should actually ship, in a real version bump:** `scopes`, `role`, `roleConflicts`, and `pairRole` become four new, optional, `omitempty`-style properties declared directly on `configField` in `SCHEMA-V1.go`/`SCHEMA-V1.ts`/`SCHEMA-V1.json` — exactly the same kind of additive struct evolution every v1 property was already designed to support, and exactly consistent with RFC-DSCONFIG-V1's own founding principle that "adopting `dsconfig` must never require a plugin to change what it stores." Adding four optional fields to an existing struct is, by Go's own semantics, a purely additive change — no existing field moves, no JSON tag changes, no existing caller breaks. Once that edit is made, `SchemaV2Extensions` as a separate document disappears entirely, and a schema author writes one document instead of two.

This RFC's reference implementation is offered as a **faithful, behavior-preserving stand-in** for that target shape — every function and validation rule documented in Section 5.1-5.4 behaves exactly as it would under the four-new-properties shape — produced this way only because of the constraint this particular round of work was done under, not because the side table is considered a good permanent design.

## 6. Breaking Changes

This is the part worth being precise about, table-first.

| Change                            | Breaks an existing v1 schema document?                                                  | Breaks existing code consuming `dsconfig`?                                                                                                                                                                                                              |
| --------------------------------- | --------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Scopes                            | No — a schema author who never declares `scopeDefs` or any field `scopes` is unaffected | **Yes, for any future client-derivation function that assumes one connection per schema and returns a single client rather than a scope-keyed result.** No such function ships in v1; this breaks a _planned_ future consumer, not anything live today. |
| Role / RoleConflicts              | No — both are optional, per field                                                       | No                                                                                                                                                                                                                                                      |
| PairRole                          | No — optional, falls back to v1 positional inference when absent                        | No — `ResolveIndexedPairs`/`ResolveIndexedPairsAsMap` are unmodified; only the new `V2` functions read `pairRole`                                                                                                                                       |
| `ValidateV2` as a new entry point | No                                                                                      | No — existing callers of `Validate`/`validateSchema` are completely unaffected; `ValidateV2` is opt-in and additional                                                                                                                                   |

**The honest summary: no change in this RFC breaks an existing, valid v1 schema document, and no change breaks any code that calls v1's existing functions.** What can break is code written against an assumption v1 never actually promised — specifically, any future code that builds something "for the whole schema" without first checking whether the schema declares more than one scope. Section 7's migration guidance addresses this directly.

## 7. Migration Path

**For schema authors with a single-connection plugin (the overwhelming majority):** no action. A v1 schema with no corresponding `SchemaV2Extensions` document remains fully valid, fully equivalent, and requires no `schemaVersion` change. There is no migration step here because there is nothing to migrate.

**For schema authors who want v2 capabilities:** author a `SchemaV2Extensions` document alongside the existing v1 `Schema` document, keyed by the same field `id`s. Adopt incrementally — `role` annotations on TLS fields can ship independently of `scopes`, which can ship independently of `pairRole` markers, exactly mirroring how RFC-DSCONFIG-V1's own v1 migration strategy treated role-like annotations as independently valuable, not an all-or-nothing adoption.

**For schema authors with a genuine multi-connection plugin:** declare `scopeDefs` and the relevant fields' `scopes`. Fields that are genuinely shared across every connection (a global timeout, for instance) should have no `scopes` entry — not a `scopes` entry listing every connection's id, per Section 5.1's staleness rule.

**For anyone building new client-derivation tooling on top of `dsconfig` going forward:** write it against a scope-keyed signature from the start (a function returning a result keyed by scope id, with a single implicit entry for the common single-connection case) rather than a signature assuming one connection per schema. This is the direct, actionable consequence of Section 6's breaking-change analysis — getting this right at the first implementation avoids a second migration once multi-connection plugins are actually exercised by real tooling.

**No automatic `SchemaV2Extensions` generator and no `schemaVersion` migration tooling are proposed.** Nothing in v2 requires an existing v1 document to change, so there is nothing to automatically rewrite — the migration path is "author new, optional metadata when you want new capabilities," not "convert existing documents."

**The one piece of migration that genuinely matters operationally:** once this RFC's side-table constraint (Section 5.5) is lifted and `scopes`/`role`/`roleConflicts`/`pairRole` move onto `configField` directly, any `SchemaV2Extensions` documents already authored need to be merged into their companion `Schema` documents. This is a one-time, mechanical transformation (for each field `id` with a non-empty extension entry, copy its properties onto the matching `configField`) and should be scoped as its own small follow-on task at the point that real edit actually lands — not attempted speculatively now, since the target shape of `configField` at that point is the actual source of truth for what the merge produces.

## 8. Stakeholders

Unchanged from RFC-DSCONFIG-V1 Section 7, with two additions specific to this proposal:

- **Multi-connection plugin authors** (AppDynamics-style) — the first stakeholder group v1 had no answer for at all.
- **Future HTTP-client-derivation tooling** — the consumer Section 6's breaking-change analysis is written for; this RFC's main actionable guidance is aimed at whoever builds that tooling next.

## 9. Benefits

- Multi-connection plugins gain a real, validated representation instead of being structurally invisible.
- Semantic roles are the concrete mechanism that lets automated HTTP client derivation and Grafana Assistant generalize across plugins by _meaning_ rather than by _name_ — directly serving both of RFC-DSCONFIG-V1's primary business drivers.
- `roleConflicts` gives the Pattern-B auth ambiguity (independent, non-mutually-exclusive boolean flags) a real, structural place to be declared, rather than living only in undocumented plugin-specific backend logic.
- Pair-role marking closes a real, silent-failure-mode bug in already-shipped resolver functions.
- Every one of the above is adoptable incrementally, field by field, schema by schema — consistent with the adoption philosophy that made RFC-DSCONFIG-V1 itself tractable across an existing plugin population.

## 10. Known Limitations

### 10.1 The side-table representation is not the proposed final shape

See Section 5.5 in full. `SchemaV2Extensions` as a document separate from `Schema` is a constraint of this specific reference implementation, not a recommended permanent design, and carries a real synchronization burden (keeping two documents' field `id`s aligned) that the proposed direct-properties-on-`configField` shape does not have.

### 10.2 No multi-connection-aware App Platform artifact

`ToPluginSchemaSettings` is unmodified by this RFC and continues to emit exactly one `Settings{Spec, SecureValues}` pair per schema document, regardless of how many scopes a schema declares. A schema can fully describe a multi-connection plugin; nothing yet derives a separate App Platform resource, or a scope-partitioned `Spec`, from that description.

### 10.3 `roleConflicts` is structural, not runtime-enforced

`ValidateRoleConflicts` confirms a schema's `roleConflicts` declarations are internally consistent — no two fields in the same effective scope claim the same role, and no field's conflict list names a role another field in its scope actually carries. It does not, and cannot, check a real configuration payload for whether two conflicting fields are simultaneously active.

### 10.4 Role validity doesn't imply role appropriateness

Nothing cross-checks a field's `role` against its `valueType` or `target`. A boolean field could declare `role: "tls.clientCert"` and nothing in this release would object.

### 10.5 The role vocabulary is closed and version-fixed

A schema author cannot extend the vocabulary. A field whose true meaning isn't represented in this release's 20 roles simply has no `role` to declare.

### 10.6 Pair-role declaration is all-or-nothing per pair

Declaring `pairRole` on one item field of a pair but not its partner is rejected outright, rather than completed by inference for the missing half.

### 10.7 v1's resolver functions are permanently unaffected

`ResolveIndexedPairs`/`ResolveIndexedPairsAsMap` will never read `pairRole`, by design — this is not a temporary gap, it's the intended permanent boundary between the v1 and v2 resolver functions (Section 5.4).

## 11. Future Work

### 11.1 Direct `configField` properties (closing Section 10.1)

The actual, proposed-final version of this RFC's three capabilities: `scopes`, `role`, `roleConflicts`, and `pairRole` as native, optional properties on `configField` in `SCHEMA-V1.go`/`SCHEMA-V1.ts`/`SCHEMA-V1.json`, with `SchemaV2Extensions` retired and a one-time mechanical merge of any already-authored extension documents into their companion schemas.

### 11.2 Multi-connection-aware App Platform conversion

Extending `ToPluginSchemaSettings` (or a new, scope-aware sibling function) to produce a separate `Settings{Spec, SecureValues}` per declared scope, or a single `Spec` partitioned by scope — closing Section 10.2. This is the natural unlock once Section 11.1 lands, since a scope-aware conversion function is simpler to design against `configField.Scopes` directly than against a side-table lookup.

### 11.3 Runtime `roleConflicts` enforcement

A function checking a real configuration payload against a schema's `roleConflicts` declarations — the runtime counterpart to `ValidateRoleConflicts`' structural check, closing Section 10.3, in the same spirit that `ResolveIndexedPairs` gave v1's `StorageMapping` a runtime after a full version of being structural-only.

### 11.4 Role/Target and Role/ValueType cross-validation

Closing Section 10.4 — a small, additive validation rule checking that a field's declared `role` is compatible with its `valueType` and `target` (for example, rejecting `role: "tls.clientCert"` on anything other than a string field targeting `secureJsonData`).

### 11.5 Vocabulary growth process

A documented process for proposing and landing new `Role` values across future minor revisions, distinct from a `v3` schema version bump — since growing the vocabulary is additive and does not, by itself, require anything else in this RFC to change.

### 11.6 Per-scope `CheckHealth` semantics

Once Section 11.2 lands, a multi-connection plugin plausibly wants to report per-connection health ("Controller: OK, EUM: unreachable") rather than one pass/fail for the whole datasource — a backend-contract question this RFC does not attempt to resolve, named here so it isn't lost.

## 12. Alternatives Considered

**A `serviceId`/`serviceIds` pair of fields, modeled narrowly on the AppDynamics case.** Considered and rejected in favor of the single, generalized `scopes` array, once a second, structurally identical partitioning need (deployment mode) surfaced during design — see Section 5.1.

**Folding role/scope/pair-role metadata into the existing, unstructured `tags` field instead of new, validated mechanisms.** Rejected: `tags` is deliberately inert and unvalidated (RFC-DSCONFIG-V1 Design Decision D5); putting load-bearing semantics into a field whose entire design contract is "decorative, nothing reads this" is a category error, not a simplification.

**Bridging this entire proposal through App Platform `x-*` vendor extensions instead of schema-level fields.** Considered and rejected, for the same reasons RFC-DSCONFIG-V1 Section 6.3 rejects authoring the App Platform artifact directly: it would still require an authoring layer to generate the extensions correctly, would scatter metadata across per-property extensions at arbitrary nesting depth instead of a flat, validated field list, and would burden every generic OpenAPI consumer of the `Spec` schema with metadata it has no interest in.

**Shipping a full expression evaluator alongside `roleConflicts` so it could be runtime-enforced immediately.** Rejected on the same evidence-bar grounds RFC-DSCONFIG-V1 already established for CEL generally: `roleConflicts`' structural check is cheap, safe, and addresses the schema-shape half of the problem; runtime enforcement is real future work (Section 11.3) but doesn't need a general evaluator to eventually exist — a narrow, purpose-built check against a real payload is sufficient and is the more honest scope for this RFC to commit to now.

## 13. Rollout Plan

This RFC's reference implementation (`SCHEMA-V2.go`, `SCHEMA-V2.ts`, `SCHEMA-V2.json`, `SCHEMA-V2.md`) ships as new, standalone artifacts, consumable by any plugin author on an opt-in basis, with zero required changes to any RFC-DSCONFIG-V1 artifact or any existing schema document. Section 11.1 (folding these properties directly onto `configField`) is the natural next step once a real version-bump merge — rather than a side-by-side artifact set — is in scope, and should be sequenced before Section 11.2 (multi-connection-aware App Platform conversion), per Section 5.5's reasoning.

## 14. References

- `SCHEMA-V2.md` — full reference documentation for v2's additions, including worked examples for scopes, roles, and pair-role marking.
- `SCHEMA-V2.go` — Go reference implementation.
- `SCHEMA-V2.ts` — TypeScript reference implementation, structurally equivalent to `SCHEMA-V2.go`, verified to type-check cleanly under `tsc --strict` against the unmodified v1 `SCHEMA-V1.ts`.
- `SCHEMA-V2.json` — formal JSON Schema describing the `SchemaV2Extensions` wire format, with its `role` enum verified to match `SCHEMA-V2.go` and `SCHEMA-V2.ts` exactly, value for value.
- RFC-DSCONFIG-V1 and its companion artifacts (`SCHEMA-V1.md`, `SCHEMA-V1.go`, `SCHEMA-V1.ts`, `SCHEMA-V1.json`) — the v1 foundation this RFC amends and does not modify.
