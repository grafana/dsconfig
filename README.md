# Grafana Datasource Configuration Schema SDK

`dsconfig` is a declarative schema for Grafana datasource configuration. It
describes every configurable field of a datasource — its type, storage
location, validation rules, UI hints, and inter-field relationships — in a
single, language-neutral document that the backend, frontend, provisioning
system, and documentation pipeline can all consume as a shared contract.

## Background

Grafana supports a large and growing catalog of first- and third-party
datasource plugins — relational databases, time-series stores, observability
backends, cloud vendor APIs, and SaaS products. Each plugin exposes a
configuration surface that users fill in when adding an instance of the
datasource: connection URLs, authentication methods, TLS material, timeouts,
cloud credentials, feature toggles, and more.

That configuration surface has historically been described implicitly, in
several disconnected places at once:

- A Go struct in the plugin backend that unmarshals the settings Grafana
  sends at query time.
- A React configuration editor in the plugin frontend that renders inputs
  and applies validation.
- Hand-written provisioning documentation showing example YAML for the
  `jsonData` and `secureJsonData` blobs.
- Reusable helpers in shared SDKs (`grafana-plugin-sdk-go`,
  `grafana-aws-sdk`, `grafana-azure-sdk-go`, `grafana-google-sdk-go`) that
  define common vocabularies which many plugins reuse verbatim.
- Reference documentation on grafana.com, generally maintained by hand.

Nothing in the platform ties these representations together. In practice
each plugin ends up owning several parallel definitions of "what fields does
my datasource have?", and those definitions drift over time. Fields are
renamed in one place but not another, validation diverges between the UI and
the backend, and provisioning docs fall behind the code.

## Reasoning

`dsconfig` was created to collapse those parallel definitions into a single,
declarative, versioned artifact — a JSON document that describes the full
configuration surface of a datasource in one place. From that document every
other representation can be derived or validated: the SDK plugin contract,
the configuration form, provisioning examples, generated reference
documentation, and inputs to automation and AI-assisted tooling.

Two design choices shape the project:

- **Additive, not disruptive.** The schema does not attempt to redesign
  Grafana's datasource storage model. Fields still live in `root`,
  `jsonData`, and `secureJsonData`, and existing plugins, provisioning
  files, and APIs continue to work unchanged. `dsconfig` is a semantic
  layer on top of the existing storage.
- **Single source of truth.** A datasource declares its configuration
  once. Downstream consumers read from that declaration instead of
  maintaining their own copy, which eliminates the drift and duplication
  that motivated the project.

## How it fits together

The schema sits between the datasource plugin author and every system that
needs to understand the plugin's configuration:

```
                    ┌─────────────────────────┐
                    │     dsconfig schema     │
                    │  (single JSON document) │
                    └────────────┬────────────┘
                                 │
        ┌────────────────┬───────┴────────┬────────────────┐
        ▼                ▼                ▼                ▼
   Grafana storage   SDK plugin      Config editor    Docs and
   (root/jsonData/   contract        (form render)    provisioning
    secureJsonData)  (grafana-                        examples
                      plugin-sdk-go)
```

Each field in the schema declares:

- Where it is stored (`root`, `jsonData`, or `secureJsonData`).
- Its type, default, and validation rules.
- How it should be rendered in the UI.
- How it relates to other fields (dependencies, conditional visibility,
  conditional requirement).

## A minimal example

Every schema requires `schemaVersion`, `pluginType`, `pluginName`, and at
least one field:

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
      "target": "root",
      "required": true
    }
  ]
}
```

See the [module documentation](./dsconfig/README.md) for worked examples
covering all three storage targets, conditional visibility, validations, and
the full field vocabulary.

## Who consumes it

- **Grafana plugin backends** — the schema is converted into the
  `PluginSchema` value consumed by `grafana-plugin-sdk-go`, giving the
  backend a typed spec and an explicit list of secure values.
- **Configuration editors** — forms can be rendered from the schema rather
  than hand-authored per plugin.
- **Provisioning and documentation pipelines** — the same schema drives
  generated reference docs and provisioning examples, and can validate
  provisioned files before deployment.
- **Automation and AI tooling** — a structured description of every
  field, including semantic relationships and constraints, is a
  first-class input for LLM-assisted configuration, migration, and
  troubleshooting workflows.

## Learn more

- [Module documentation](./dsconfig/README.md) — full field reference and
  worked examples.
- [Design document](./dsconfig/schema.md) — schema semantics, expressions,
  and versioning.

## License

Licensed under the terms of the [LICENSE](./LICENSE) file in this
repository.
