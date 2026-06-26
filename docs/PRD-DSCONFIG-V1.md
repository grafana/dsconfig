# PRD: `dsconfig` v1 — A Schema Contract for Grafana Datasource Plugin Configuration

| Field    | Value                                                                                                        |
| -------- | ------------------------------------------------------------------------------------------------------------ |
| Document | Product Requirements Document                                                                                |
| Product  | `dsconfig` (v1)                                                                                              |
| Status   | Draft                                                                                                        |
| Owner    | @yesoreyeram                                                                                                 |
| Related  | RFC-DSCONFIG-V1 (technical design and reference implementation), `SCHEMA-AUTHORING-V1.md` (authoring how-to) |

This PRD defines **what** `dsconfig` v1 needs to deliver and **why**, in product terms. It does not restate the schema's field-by-field design or its Go/TypeScript implementation — that lives in RFC-DSCONFIG-V1 and is treated here as already-specified. Where this document and the RFC appear to overlap, the RFC is the source of truth for technical shape; this document is the source of truth for scope, priority, and success criteria.

---

## 1. Problem

Grafana ships with, and the community maintains, configuration for 150+ datasource plugins. Every one of them independently decides:

- What fields its connection settings have, and what they're named.
- Where each field is stored (`root`, `jsonData`, or the encrypted `secureJsonData`).
- How its authentication method is selected — sometimes one explicit field, sometimes several independent toggles.
- How its `ConfigEditor` (the settings-page UI) and its backend Go code agree on what that configuration means — today, only by the original author's memory.

There is no machine-readable description of any of this, for any plugin, anywhere. That absence is the root cause of four concrete, recurring problems:

1. **Configuration errors surface too late.** A malformed setting is discovered inside `CheckHealth` or a live query, not when it's saved — with no indication of which field was wrong.
2. **The config editor and backend can silently drift.** Nothing checks that what the UI collects matches what the backend expects, especially once a plugin outlives its original author.
3. **Nothing outside a plugin's own source code can reason about its configuration.** Not provisioning validation, not the plugin catalog's review tooling, not a future conversational assistant.
4. **Grafana's App Platform (its Kubernetes-style resource API) has no structural schema to register datasource configuration against**, which blocks treating a datasource as a properly validated, typed resource the way every other App Platform resource already is.

## 2. Why now

Two efforts inside Grafana currently have a hard dependency on solving this, which is what makes this the right time to invest rather than defer further:

- **Grafana Assistant** wants to create and edit datasources conversationally. Doing that reliably — getting to a working, validated connection in the fewest turns, without guessing at undocumented form behavior — requires a structured artifact describing what a plugin needs. Today there isn't one.
- **Grafana App Platform** requires every resource type it manages to have an OpenAPI-shaped structural schema for admission validation and typed CRUD. Datasource configuration, stored as two untyped JSON blobs, cannot be registered as a proper App Platform resource without first having that schema produced from somewhere.

Both efforts are currently blocked on the same missing artifact. `dsconfig` is that artifact.

## 3. Goals

1. Give every datasource plugin a single, authored-once, machine-readable description of its configuration — fields, types, storage location, validation rules, and presentation hints.
2. Make that description the literal input to the schema App Platform needs for its resource API, via one deterministic conversion — not two separately maintained schemas.
3. Make that description sufficient for Grafana Assistant to create or repair a datasource configuration without reading plugin source code.
4. Require zero changes to any plugin's existing stored configuration, storage format, or `ConfigEditor` as a precondition of adoption.
5. Make adoption incremental and field-by-field — a plugin author should get value from annotating one field, without committing to a full rewrite.

## 4. Non-Goals (for v1)

- **Not** a new storage format for datasource configuration. `root`/`jsonData`/`secureJsonData` are unchanged.
- **Not** a guarantee that every existing plugin will have a schema authored for it by any particular date. v1 ships the _capability_; ecosystem-wide authoring is a separate, ongoing effort.
- **Not** an HTTP client builder, a config-editor renderer, or a conversational agent. v1 is the schema and its validation/conversion layer; these are named, explicitly deferred consumers (see Section 8).
- **Not** an expression evaluator for conditional logic (`dependsOn`, `requiredWhen`, etc.). v1 can _store_ these as structured intent; it does not execute them. See RFC-DSCONFIG-V1 Section 10.1 for the full reasoning.
- **Not** a fix for plugins with more than one independent connection (AppDynamics-style). That's explicitly scoped to a v2 follow-on (RFC-DSCONFIG-V2).

## 5. Target users and personas

| Persona                                                     | What they need from `dsconfig` v1                                                                                                                     | What they don't need to do differently                                                    |
| ----------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------- |
| **Plugin author (core team)**                               | A documented format to describe their plugin's existing configuration; clear guidance on storage targets, secrets, and legacy patterns (headers, TLS) | Rename any field, move any value between storage buckets, or rewrite their `ConfigEditor` |
| **Plugin author (community)**                               | The same, at lower stakes and with no obligation                                                                                                      | Coordinate with anyone else's release schedule — adoption is fully opt-in                 |
| **Datasource operator (provisioning via YAML / Terraform)** | A way to validate a config file before deploying it, with field-level errors                                                                          | Nothing — this is a pure upside once a plugin has a schema                                |
| **Plugin catalog / review team**                            | A way to programmatically check a submitted plugin's configuration for conformance and completeness                                                   | Manually inspect every config editor by hand                                              |
| **Grafana Assistant team**                                  | A structured artifact sufficient to drive a datasource-creation conversation to a working connection                                                  | Hand-write per-plugin prompts or field lists                                              |
| **App Platform team**                                       | An OpenAPI-shaped `spec`/`secureValues` pair per plugin, produced deterministically                                                                   | Design or maintain a second, parallel schema for datasource resources                     |

## 6. User stories

**As a plugin author**, I want to describe my plugin's existing `jsonData`/`secureJsonData` fields in one document, so that I get save-time validation and a path to App Platform support without changing anything my plugin already stores.

**As a plugin author with a TLS or basic-auth field set**, I want to reuse a well-known shape rather than re-inventing field names, so that future tooling (and other engineers reading my schema) recognize what my fields mean without guessing.

**As an operator using provisioning YAML**, I want a malformed datasource config to fail in CI, with a specific field and a specific reason, rather than failing only after `grafana-server` starts and the datasource is already broken in production.

**As someone building the App Platform datasource resource type**, I want one conversion function that turns a `dsconfig` schema into the exact `Settings{Spec, SecureValues}` shape my resource registration needs, so that I am never reconciling two independently authored descriptions of the same plugin.

**As someone building Grafana Assistant's datasource-configuration flow**, I want to know, for any plugin with a schema, which fields are required, which hold secrets, and what a valid value looks like — without parsing that plugin's Go source — so that I can get a person to a working connection in as few turns as possible.

**As a plugin reviewer**, I want to run an automated check against a submitted plugin's schema (does it validate structurally? does every secret target `secureJsonData`?) as part of review, rather than reading the `ConfigEditor` component by hand every time.

## 7. Scope: what v1 actually ships

This section is the product-level summary of RFC-DSCONFIG-V1's design; see that document for the full specification.

**In scope:**

- A versioned schema format (`Schema`, `ConfigField`) covering root/jsonData/secureJsonData fields, including nested objects (`section`), arrays and maps (`item`), and the legacy indexed key/value pair convention used for HTTP headers (`storage.indexedPair`).
- Structural validation of a schema document (`Validate`/`validateSchema`), including `id` format enforcement so that future conditional-expression evaluation isn't silently broken by a malformed reference.
- A deterministic conversion (`ToPluginSchemaSettings`) from a `dsconfig` schema into the `Settings{Spec, SecureValues}` shape App Platform and `grafana-plugin-sdk-go` both need.
- Read-side lookup utilities (`FieldByID`, `ValueByID`, `ResolveIndexedPairs`, `ResolveIndexedPairsAsMap`) so that any consumer — an editor, a future validator, an assistant — can resolve a field reference against real configuration data without re-implementing storage-path resolution.
- Reference implementations in Go and TypeScript, kept in agreement by design and verified to match (see RFC-DSCONFIG-V1 Section 10.11 for the one known, documented divergence in error-reporting behavior between them).
- An authoring guide with worked examples, including three real, shipped plugins (Sentry, GitHub, BigQuery), covering every storage pattern a plugin author is likely to encounter.

**Explicitly out of scope, named and tracked, not silently dropped:**

- Expression evaluation, HTTP client derivation, multi-connection plugins, semantic field roles, runtime validation against live configuration. Each of these is named in RFC-DSCONFIG-V1 Section 10/11 with a stated reason for deferral and a clear path to follow-on work — several (semantic roles, multi-connection scopes) are already specified in RFC-DSCONFIG-V2 (v2).

## 8. Success metrics

v1's success is about **adoption and reliability of the contract**, not about any single plugin's UX — the UX wins (better config editors, assistant-driven setup) are downstream consumers this PRD explicitly does not promise to deliver in v1.

| Metric                                                                                                                                                     | What it tells us                                                                                                                                                                                                              |
| ---------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| % of core plugins with an authored, validating `dsconfig` schema                                                                                           | Direct measure of foundation-layer adoption                                                                                                                                                                                   |
| % of those schemas that pass structural validation on first authoring attempt vs. require iteration                                                        | Signal on whether the authoring guide and error messages are actually clear                                                                                                                                                   |
| Number of plugins whose schema required _any_ change to existing stored configuration to adopt                                                             | Should be zero, by design — any non-zero count is a regression against the additive-only goal                                                                                                                                 |
| App Platform: number of plugins with a generated `Settings` artifact consumed by resource registration                                                     | Direct measure of the App Platform dependency being unblocked                                                                                                                                                                 |
| Time-to-first-successful-connection in any Assistant-driven datasource creation flow built against a `dsconfig`-described plugin, vs. a plugin without one | Direct measure of the Assistant dependency being unblocked (requires the Assistant integration itself, which is downstream of this PRD — tracked as a leading indicator once that integration exists, not a v1 launch metric) |
| Number of plugin-catalog review findings related to config/secret-handling mistakes, before vs. after schema-based review tooling                          | Measures the review-automation benefit                                                                                                                                                                                        |

**What v1 does not need to move, to be considered successful:** end-user-visible config editor UI (no schema-driven renderer ships in v1 — see Non-Goals), and full ecosystem coverage (partial, voluntary adoption is the expected and accepted v1 state).

## 9. Rollout plan

v1 is additive and opt-in at every layer, so rollout is "make it available," not "migrate everyone by a date":

1. **Ship the schema format, reference implementations, and authoring guide.** No plugin is required to do anything.
2. **Author schemas for a small number of core plugins first** (good candidates: plugins already exercising the patterns the authoring guide documents — direct fields, TLS, indexed-pair headers, discriminator-based auth) to validate the format against real, already-shipped configuration before asking the broader plugin population to adopt it.
3. **Wire the conversion output into App Platform's resource registration** for at least one plugin, end to end, as the first real consumer proving the artifact is sufficient for its stated purpose.
4. **Open authoring to the wider plugin catalog**, community included, with the authoring guide as the primary onboarding artifact and plugin-catalog review tooling as the feedback loop.
5. **Track adoption via the metrics in Section 8**, without a forced deadline — partial coverage is an accepted, expected state for v1, not a failure condition.

## 10. Risks

| Risk                                                                                                                                                                          | Mitigation                                                                                                                                                                                                                                                                                               |
| ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Plugin authors see schema-authoring as pure overhead with no immediate payoff, since v1 ships no client-derivation or UI-rendering benefit yet                                | Lead adoption with the plugins that benefit soonest from App Platform/Assistant integration; be explicit in the authoring guide about what v1 unlocks today (validation, App Platform eligibility) vs. what's coming (Section 11 of the RFC)                                                             |
| A plugin author misclassifies a field's storage target (e.g., a secret accidentally described as `jsonData`)                                                                  | Structural validation catches type/shape errors but cannot catch a schema author asserting the wrong `target` for a field that's actually a secret — this is a documentation and review-checklist responsibility, called out explicitly in the authoring guide's "Common mistakes" section               |
| The known `indexedPair` secure-value propagation gap (RFC Section 10.3) ships unfixed and a plugin's header values are silently exposed as non-secret in a generated artifact | Treated as the single highest-priority fix in the RFC's own limitations list; should not be allowed to ship to a plugin actually using `indexedPair` for secrets without the fix landing first, or with extremely clear documentation of the gap in the interim (as already done in the authoring guide) |
| Two reference implementations (Go/TS) drift in behavior over time                                                                                                             | Documented, known divergence (error-collection style) is already tracked; any _new_ divergence should block release of whichever implementation introduces it                                                                                                                                            |
| Adoption stalls because nothing forces it                                                                                                                                     | Accepted as the correct trade-off for v1 — forcing adoption before the format is proven against real plugins would risk a worse outcome (a flawed mandatory format) than slower, voluntary, validated adoption                                                                                           |

## 11. Dependencies

- **Grafana App Platform** must have a registration path ready to consume `Settings{Spec, SecureValues}` for this to deliver its stated business value end-to-end; v1 ships the producer side of that contract regardless of the consumer side's readiness.
- **Grafana Assistant** integration is a separate, downstream workstream (RFC-DSCONFIG-V1 Section 11.11) — v1 makes it _possible_, it does not itself build it.
- **Plugin catalog review tooling** updates to actually run structural validation as part of submission review is a separate, small follow-on task, not part of this PRD's delivery.

## 12. Open questions for stakeholder review

1. Which core plugins should be the first to get an authored schema, and who owns writing them?
2. What's the actual integration timeline on the App Platform side — does anything in this PRD's rollout plan need to be resequenced around it?
3. Should plugin-catalog submission _require_ a `dsconfig` schema at some future point, or remain permanently optional? (This PRD takes no position; RFC-DSCONFIG-V1 explicitly treats this as undecided.)
4. Is there an appetite to prioritize the `indexedPair` secure-value propagation fix (Section 10, Risks) ahead of broader rollout, given it's a correctness bug rather than a missing feature?

## 13. References

- RFC-DSCONFIG-V1 — full technical design, schema shape, and known limitations.
- RFC-DSCONFIG-V2 — v2 proposal (multi-connection scopes, semantic roles, pair-role marking), relevant context for what's intentionally deferred from v1.
- `dsconfig-authoring-guide.md` — the practical, example-driven authoring reference this PRD's rollout plan leans on for onboarding.
