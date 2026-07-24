# Grafana Datasource Configuration Schema SDK

`dsconfig` is the reference implementation of Grafana's declarative datasource
configuration schema. It provides a single, language-neutral source of truth
that describes every configurable field of a Grafana datasource — its type,
storage location, validation rules, UI hints, and inter-field relationships —
together with the Go SDK, tooling, and reusable field packs required to author,
validate, and consume that schema across the Grafana ecosystem.

The project is designed to eliminate the drift between the many independent
representations of a datasource's configuration surface (backend structs,
frontend forms, provisioning files, documentation, and API contracts) by
promoting a single schema document to the role of contract between them.

---

## Background

Grafana ships with, and supports, an ever-growing catalog of first- and
third-party datasource plugins — relational databases, time-series stores,
observability backends, cloud vendor APIs, SaaS products, and more. Each of
those plugins exposes a configuration surface that users fill in when they
add an instance of the datasource: connection URLs, authentication methods,
TLS material, timeouts, cloud credentials, feature toggles, and so on.

Historically, the shape of that configuration surface has been described
implicitly, in several places at once:

- A Go `struct` in the plugin backend that unmarshals the settings Grafana
  sends at query time.
- A React configuration editor in the plugin frontend that renders inputs,
  applies validation, and posts values back to Grafana.
- Hand-written provisioning documentation showing example YAML for the
  `jsonData` and `secureJsonData` blobs.
- Reusable helpers in SDKs such as
  [`grafana-plugin-sdk-go`](https://github.com/grafana/grafana-plugin-sdk-go),
  [`grafana-aws-sdk`](https://github.com/grafana/grafana-aws-sdk),
  [`grafana-azure-sdk-go`](https://github.com/grafana/grafana-azure-sdk-go),
  and
  [`grafana-google-sdk-go`](https://github.com/grafana/grafana-google-sdk-go)
  that define well-known field vocabularies (basic auth, TLS, SigV4, Azure
  AD, GCE service accounts) that many plugins reuse verbatim.
- Marketing and reference documentation on grafana.com, generally maintained
  by hand.

Nothing in the platform ties these representations together. In practice
each plugin ends up owning several parallel definitions of "what fields does
my datasource have?", and those definitions drift over time. Fields are
renamed in one place but not another, validation rules diverge between the
UI and the backend, provisioning docs fall behind the code, and shared SDK
fields are copy-pasted into each new plugin.

## Reasoning

`dsconfig` was created to collapse those parallel definitions into a single,
declarative, machine-readable artifact — a JSON document that describes the
full configuration surface of a datasource in one place. From that document,
every other representation can be derived or validated:

- **Storage shape.** Each field declares its `target` (`root`, `jsonData`,
  or `secureJsonData`), so the exact on-disk shape of a Grafana datasource
  configuration is fully described without changing Grafana's existing
  storage model.
- **Backend contracts.** The schema is converted into the `PluginSchema`
  value consumed by `grafana-plugin-sdk-go`, giving the backend a typed spec
  and an explicit list of secure values without hand-written boilerplate.
- **Frontend forms.** With `valueType`, `ui`, `validations`, `dependsOn`,
  and `requiredWhen` on every field, a configuration editor can be
  generated from the schema rather than hand-authored per plugin.
- **Provisioning and documentation.** The same schema drives generated
  reference documentation, provisioning examples, and validation of
  provisioned files before deployment.
- **Automation and AI tooling.** A structured description of every field —
  including semantic relationships, defaults, and constraints — is a
  first-class input for LLM-assisted configuration, migration, and
  troubleshooting workflows.

Two design choices are worth calling out because they follow directly from
this reasoning:

1. **Additive, not disruptive.** The schema does not attempt to redesign
   Grafana's datasource storage model. Existing plugins, provisioning
   files, and APIs continue to work unchanged. `dsconfig` is a semantic
   layer over the current `root` / `jsonData` / `secureJsonData` split.
2. **Reusable field packs.** Rather than force each plugin to redeclare
   the fields defined by shared SDKs, `dsconfig` provides curated
   [field packs](#field-packs) that a plugin composes via `baseFields`.
   A plugin only declares what is genuinely its own; common SDK fields
   are inherited, and can be surgically excluded or lightly patched for
   presentation.

The result is a single, versioned contract per datasource that the backend,
frontend, provisioning system, documentation pipeline, and automation
tooling all consume — dramatically reducing the drift, duplication, and
maintenance cost that motivated the project.

---

## Why dsconfig

Datasource plugins historically maintain the shape of their configuration in
several disconnected places: the Go backend, the React configuration editor,
provisioning documentation, and hand-maintained OpenAPI or JSON Schema files.
Keeping these representations aligned is expensive, error-prone, and a common
source of bugs and regressions.

`dsconfig` addresses this by promoting the schema itself to a first-class,
versioned artifact that downstream systems consume directly:

- **LLM and automation tooling** — a structured description of every field
  enables AI-assisted configuration, generation, and migration workflows.
- **Configuration editors** — forms can be rendered from the schema rather
  than hand-written per plugin.
- **Documentation** — the field reference for each datasource can be
  auto-generated from the same schema that powers the UI.
- **Provisioning** — the exact shape of `root`, `jsonData`, and
  `secureJsonData` is described declaratively and can be validated ahead of
  deployment.
- **Validation** — a single set of rules is enforced consistently in the UI,
  at provisioning time, and in backend code paths.

The schema is additive: it does not change Grafana's existing datasource
configuration model. Fields continue to live in `root`, `jsonData`, and
`secureJsonData`. `dsconfig` is a semantic layer on top of that established
storage model.

---

## Repository Layout

This repository is a Go workspace composed of two modules and their supporting
assets.

| Path | Description |
| --- | --- |
| [`dsconfig/`](./dsconfig) | Primary Go module. Defines the schema types, JSON Schema, resolver, validators, converters, and the field-pack registry. See the module [README](./dsconfig/README.md) for the full field-by-field reference and worked examples. |
| [`dsconfig/packs/`](./dsconfig/packs) | Built-in reusable field packs (`plugin_sdk_settings`, `aws_sdk_settings`, `azure_sdk_settings`, `google_sdk_settings`, `sqleng_settings`) that datasources compose via `baseFields`. Each pack ships both a Go registration and a companion JSON definition. |
| [`dsconfig/cmd/gen-schema-json/`](./dsconfig/cmd/gen-schema-json) | Code generator that keeps `dsconfig/schema.json` in sync with the on-disk pack JSON files, so editor autocomplete and JSON Schema validation always reflect the current pack contents. |
| [`dsconfig/schema.json`](./dsconfig/schema.json) | JSON Schema for `dsconfig` documents. Point your editor at this file to get autocomplete and validation while authoring plugin schemas. |
| [`dsconfig/schema.md`](./dsconfig/schema.md) | Full design document describing schema semantics, field targets, validations, expressions, and versioning. |
| [`schema/`](./schema) | Compatibility module that adapts a `dsconfig` schema into the `grafana-plugin-sdk-go` `PluginSchema` (spec + secure values) consumed by Grafana's plugin runtime. |
| [`docker-compose.yaml`](./docker-compose.yaml) | Local Grafana Enterprise environment preloaded with the datasource plugins used to develop and exercise the schemas end-to-end. |
| [`go.work`](./go.work) | Go workspace pinning both modules together for local development. |

---

## Architecture at a Glance

```
        ┌─────────────────────────────┐
        │       dsconfig schema       │   single source of truth
        │  (JSON, versioned artifact) │
        └───────────────┬─────────────┘
                        │
        ┌───────────────┼──────────────────────────┬─────────────────┐
        ▼               ▼                          ▼                 ▼
  ┌───────────┐   ┌────────────┐          ┌────────────────┐   ┌──────────┐
  │ Grafana   │   │ SDK        │          │ Config editors │   │ Docs &   │
  │ storage   │   │ PluginSchema│         │ (form render)  │   │ provisio-│
  │ (root/    │   │ (spec +    │          │                │   │ ning     │
  │  jsonData/│   │  secure    │          │                │   │          │
  │  secure)  │   │  values)   │          │                │   │          │
  └───────────┘   └────────────┘          └────────────────┘   └──────────┘
```

- **`dsconfig` module** parses, validates, resolves base-field packs, and
  converts a schema into the shape expected by `grafana-plugin-sdk-go`.
- **`schema` module** provides a thin backward-compatible façade over the
  primary API for existing consumers.
- **Packs** are curated bundles of common fields (auth, TLS, AWS/Azure/GCP
  credentials, SQL engine tuning) that datasources include by reference to
  avoid duplication and to guarantee a consistent field vocabulary across the
  ecosystem.

---

## Getting Started

### Prerequisites

- Go 1.26 or newer (as declared in [`go.work`](./go.work))
- Optional: Docker and Docker Compose, to run the bundled Grafana Enterprise
  environment

### Install the SDK

Add the module to your Go project:

```sh
go get github.com/grafana/dsconfig/dsconfig
```

### Author a schema

A minimal schema requires `schemaVersion`, `pluginType`, `pluginName`, and at
least one field. Each field declares an `id`, `key`, `valueType`, and a storage
`target`:

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

For the full field reference, worked examples covering root, `jsonData`, and
`secureJsonData` targets, conditional visibility, validations, and base-field
packs, see the [module documentation](./dsconfig/README.md) and
[design document](./dsconfig/schema.md).

### Consume a schema in Go

The `dsconfig` package parses and validates schema documents, resolves any
referenced field packs, and produces the `PluginSchema` value consumed by
`grafana-plugin-sdk-go`:

```go
import (
    "github.com/grafana/dsconfig/dsconfig"
    _ "github.com/grafana/dsconfig/dsconfig/packs" // register built-in packs
)
```

Refer to the module README for the complete resolver, validator, and converter
API surface.

---

## Field Packs

The following packs ship in this repository and are registered by importing
`github.com/grafana/dsconfig/dsconfig/packs`:

| Pack ID | Purpose |
| --- | --- |
| `plugin_sdk_settings` | Fields common to every Grafana datasource (URL, basic auth, TLS, proxy, headers, OAuth pass-through). |
| `aws_sdk_settings` | AWS authentication and region selection shared across AWS-backed datasources. |
| `azure_sdk_settings` | Azure AD / Managed Identity authentication shared across Azure-backed datasources. |
| `google_sdk_settings` | Google Cloud authentication (JWT, GCE) shared across Google-backed datasources. |
| `sqleng_settings` | Connection pool and query behavior fields shared across SQL datasources. |

Datasources compose packs through the `baseFields` array in their schema and
may override or omit individual pack fields on a per-plugin basis.

---

## Local Development

Run the standard Go workflows from the repository root:

```sh
# Fetch dependencies for both modules
go work sync

# Build every package in the workspace
go build ./...

# Run the full test suite (includes conformance tests for the SDK adapter)
go test ./...

# Regenerate the enum lists inside dsconfig/schema.json after editing a pack
go generate ./...
```

### Running Grafana locally

A Docker Compose file is provided to start Grafana Enterprise with the
datasource plugins that this schema targets:

```sh
# Optionally export a license: GF_ENTERPRISE_LICENSE_TEXT=...
docker compose up
```

Grafana will be available on <http://localhost:3000> with anonymous admin
access enabled for development convenience.

---

## Versioning and Stability

The schema is versioned via the `schemaVersion` field on every document. The
current version is `v1`. Future backward-incompatible changes to the schema
model will increment this version; the SDK will continue to support previous
versions in accordance with Grafana's compatibility policy.

The Go API of the `dsconfig` module follows standard Go module semantic
versioning. The `schema` module is retained as a thin compatibility layer over
`dsconfig` for existing callers and will be maintained until those callers
migrate.

---

## Contributing

Contributions are welcome. When proposing changes, please:

1. Open an issue describing the motivation and expected impact before making
   substantial API or schema changes.
2. Ensure `go build ./...`, `go test ./...`, and `go generate ./...` all
   succeed locally.
3. When adding or modifying a field pack, update the companion JSON file and
   regenerate `dsconfig/schema.json` via `go generate ./...` so editor
   autocomplete stays consistent with the pack contents.
4. Follow the existing documentation style and update the relevant reference
   material in [`dsconfig/README.md`](./dsconfig/README.md) or
   [`dsconfig/schema.md`](./dsconfig/schema.md).

---

## License

This project is licensed under the terms of the [LICENSE](./LICENSE) file
included in this repository.
