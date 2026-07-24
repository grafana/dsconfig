# Grafana Datasource Configuration Schema SDK

`dsconfig` is a declarative schema for Grafana datasource configuration. It
describes every configurable field of a datasource — its type, storage
location, validation rules, UI hints, and relationships — in a single,
language-neutral document that the backend, frontend, provisioning, and
documentation pipelines can all consume.

## Background

Grafana supports a large and growing catalog of datasource plugins. Each
plugin exposes a configuration surface — connection URLs, authentication
methods, TLS material, credentials, feature toggles — that has historically
been described in several disconnected places at once: the Go backend struct,
the React configuration editor, provisioning documentation, and shared SDK
helpers.

Nothing in the platform tied those representations together, so they drifted:
fields were renamed in one place but not another, validation diverged between
the UI and the backend, and provisioning docs fell behind the code.

## Reasoning

`dsconfig` collapses those parallel definitions into a single, versioned
artifact. From one schema document, every other representation can be
derived or validated — the SDK plugin contract, the configuration form, the
provisioning example, the reference docs, and inputs to automation and
AI-assisted tooling.

Two design choices shape the project:

- **Additive, not disruptive.** The schema does not change Grafana's
  existing datasource storage model. Fields still live in `root`,
  `jsonData`, and `secureJsonData`; `dsconfig` is a semantic layer on top.
- **Single source of truth.** A datasource declares its configuration
  once. Downstream consumers read from that declaration instead of
  maintaining their own copy.

## What it enables

- A typed plugin contract for `grafana-plugin-sdk-go` without hand-written
  boilerplate.
- Configuration editors generated from the schema instead of hand-authored
  per plugin.
- Auto-generated field reference documentation and provisioning examples.
- Structured metadata for validation, migration, and LLM-assisted tooling.

## Learn more

- [Module documentation](./dsconfig/README.md) — full field reference and
  worked examples covering `root`, `jsonData`, and `secureJsonData`
  targets, conditional visibility, and validations.
- [Design document](./dsconfig/schema.md) — schema semantics, expressions,
  and versioning.

## License

Licensed under the terms of the [LICENSE](./LICENSE) file in this
repository.
