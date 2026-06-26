# Authoring `dsconfig` Schemas for Grafana Datasources — A User Guide

This guide shows how to author a `dsconfig` schema for a Grafana datasource plugin, with one worked example per common configuration pattern. Each example shows three things, in order:

1. **The `dsconfig` schema** — what you author, once, per plugin.
2. **Grafana storage** — what gets persisted today, unchanged. This is what your plugin's backend already reads and what an existing `ConfigEditor` already writes.
3. **The App Platform schema artifact** — the Kubernetes-style API schema produced from your `dsconfig` schema, in the exact shape Grafana's plugin schema provider (`pluginschema.PluginSchema` / `pluginschema.Settings`) loads and serves. This is what App Platform validates `create`/`update` requests against, and what other tooling (Grafana Assistant, the plugin catalog, provisioning validators) can consume without reading your plugin's source code.

Read "Before you start" once, then jump to whichever use case matches what you're building.

---

## Before you start

### What you are and are not changing

Authoring a `dsconfig` schema never changes where or how your plugin's configuration is stored. Every field you describe continues to live exactly where it lives today: as a top-level property on the datasource record (`root`), inside `jsonData`, or inside `secureJsonData`. You are writing a _description_ of your plugin's existing configuration, not a new configuration format.

### The three storage targets

| `target`         | Where it lands                                                      | Encrypted | Readable after save |
| ---------------- | ------------------------------------------------------------------- | --------- | ------------------- |
| `root`           | Top-level datasource property (`url`, `basicAuth`, `database`, ...) | No        | Yes                 |
| `jsonData`       | `jsonData.*`                                                        | No        | Yes                 |
| `secureJsonData` | `secureJsonData.*`                                                  | Yes       | No — write-only     |

If a field holds a secret (a password, an API key, a token, a private key), it must target `secureJsonData`. Everything else targets `root` or `jsonData`.

### The two identifiers every field has

| Key   | Purpose                                                                                                                        | Example                    |
| ----- | ------------------------------------------------------------------------------------------------------------------------------ | -------------------------- |
| `id`  | Globally unique reference within the schema document. Used by `groups`, `relationships`, and `effects` to refer to this field. | `"auth.basicAuthPassword"` |
| `key` | The literal property name in storage.                                                                                          | `"basicAuthPassword"`      |

Use a dot-separated `id` (e.g. `"connection.url"`, `"auth.password"`). This is now partially enforced, not just a convention: each dot-separated segment must match `^[A-Za-z_][A-Za-z0-9_]*$` (letters, digits, underscore; no hyphens, brackets, or spaces), and no `id` may be a strict prefix of another `id` in the same document — `tls.clientAuth` and `tls.clientAuth.enabled` can't both exist. Both rules are checked at schema-validation time and will reject your document if violated; they exist because a future evaluator of `dependsOn`/`requiredWhen`/`effects[].when` will need to resolve these dotted paths unambiguously, and catching a problem now is better than discovering it once that evaluator exists.

### Minimum required schema document

Every schema document needs four things at the root: `schemaVersion`, `pluginType`, `pluginName`, and at least one entry in `fields`.

### How to read each use case below

Each use case is self-contained. The "App Platform schema artifact" JSON in every example is the literal output shape produced by converting the `dsconfig` schema (via `ToPluginSchemaSettings()`) and wrapping it in the `pluginschema.PluginSchema` envelope that Grafana's schema provider serves per API version. You can hand this JSON, as-is, to anything that needs to validate a datasource configuration without reading your plugin's code — including App Platform's resource admission validation and Grafana Assistant.

A field that is _not_ a secret never appears in the artifact's `secureValues` list — it appears as an ordinary property under `settings.spec` (directly, or nested under `settings.spec.properties.jsonData` if its `target` is `jsonData`). A field that _is_ a secret (`target: secureJsonData`) appears **only** in `settings.secureValues`, identified by `key`, `description`, and `required` — it never appears under `spec`, and its value is never present anywhere in this artifact, because the artifact describes shape, not stored values.

This has a direct, easy-to-miss consequence for `settingsExamples`: each entry's `value` is, by the OpenAPI convention this artifact follows, an example of a valid `spec` — and `spec` structurally excludes secrets. So **no `settingsExamples` entry in this guide ever includes a `secureJsonData` block**, for any use case with secrets — not because secrets are being hidden from you, but because `spec` is not where they would go even in principle. Every use case below that has secret fields shows this explicitly: the example's `value` covers only what `spec` actually validates, and a clearly labeled, separate note states which `secureValues` keys must _also_ be set for that example to represent a complete, working configuration. Treat the `value` block and its accompanying secure-values note as one pair, not the `value` block alone, when you're deciding what a full example configuration actually requires.

---

## Use case 1 — A single root field (`url`)

The simplest possible datasource: one field, the connection URL, stored at the root level.

### 1. `dsconfig` schema

```json
{
  "schemaVersion": "v1",
  "pluginType": "grafana-example-datasource",
  "pluginName": "Example Datasource",
  "fields": [
    {
      "id": "connection.url",
      "key": "url",
      "description": "Base URL of the datasource",
      "valueType": "string",
      "target": "root",
      "required": true,
      "validations": [
        {
          "type": "pattern",
          "pattern": "^https?://",
          "message": "Must be HTTP(S)"
        }
      ],
      "ui": {
        "component": "input",
        "placeholder": "https://example.com/api"
      }
    }
  ]
}
```

### 2. Grafana storage

```json
{
  "name": "example ds - dev",
  "type": "grafana-example-datasource",
  "url": "https://example.com/api",
  "jsonData": {},
  "secureJsonData": {}
}
```

### 3. App Platform schema artifact

```json
{
  "targetApiVersion": "v0alpha1",
  "settings": {
    "spec": {
      "required": ["url"],
      "properties": {
        "url": {
          "description": "Base URL of the datasource",
          "type": "string",
          "pattern": "^https?://"
        }
      },
      "additionalProperties": false,
      "example": {
        "url": "https://example.com/api"
      }
    }
  },
  "settingsExamples": {
    "examples": {
      "simple": {
        "description": "Minimal valid configuration",
        "value": {
          "url": "https://example.com/api"
        }
      }
    }
  }
}
```

**Notes:**

- `url` lands directly under `settings.spec.properties`, not nested under `jsonData`, because its `target` is `root`.
- `settings.secureValues` is absent entirely — this schema has no secret fields, so the key is omitted rather than present-but-empty.
- `validations[].pattern` becomes the JSON Schema `pattern` keyword directly.

---

## Use case 2 — Root field, `jsonData`, and `secureJsonData` together

A more typical datasource: a connection URL and basic auth flag at the root level, a couple of non-secret settings in `jsonData`, and two secrets in `secureJsonData`.

### 1. `dsconfig` schema

```json
{
  "schemaVersion": "v1",
  "pluginType": "grafana-example-datasource",
  "pluginName": "Example Datasource",
  "fields": [
    {
      "id": "connection.url",
      "key": "url",
      "valueType": "string",
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
      "id": "connection.basicAuthUser",
      "key": "basicAuthUser",
      "valueType": "string",
      "target": "root"
    },
    {
      "id": "jsonData.tlsSkipVerify",
      "key": "tlsSkipVerify",
      "valueType": "boolean",
      "target": "jsonData"
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
      "target": "jsonData"
    },
    {
      "id": "secure.basicAuthPassword",
      "key": "basicAuthPassword",
      "valueType": "string",
      "target": "secureJsonData",
      "requiredWhen": "connection.basicAuth == true"
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

### 2. Grafana storage

```json
{
  "name": "example ds - dev",
  "type": "grafana-example-datasource",
  "url": "https://example.com/api",
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

### 3. App Platform schema artifact

```json
{
  "targetApiVersion": "v0alpha1",
  "settings": {
    "spec": {
      "required": ["url"],
      "properties": {
        "url": {
          "description": "Server URL",
          "type": "string"
        },
        "basicAuth": {
          "type": "boolean"
        },
        "basicAuthUser": {
          "type": "string"
        },
        "jsonData": {
          "type": "object",
          "properties": {
            "tlsSkipVerify": {
              "type": "boolean"
            },
            "timeout": {
              "type": "number",
              "default": 30,
              "minimum": 1,
              "maximum": 300
            },
            "serverName": {
              "type": "string"
            }
          }
        }
      },
      "additionalProperties": false
    },
    "secureValues": [
      {
        "key": "basicAuthPassword",
        "x-dsconfig-required-when": "connection.basicAuth == true"
      },
      { "key": "tlsCACert" }
    ]
  },
  "settingsExamples": {
    "examples": {
      "simple": {
        "description": "Basic auth over HTTPS, with a custom CA certificate. Requires secureValues to also be set — see note below; this value alone is not a complete configuration.",
        "value": {
          "url": "https://example.com/api",
          "basicAuth": true,
          "basicAuthUser": "your_username",
          "jsonData": {
            "timeout": 60,
            "serverName": "api.example.com"
          }
        }
      }
    }
  }
}
```

**This example also requires (not expressible inside `settingsExamples`, since `secureValues` sits beside `spec`, not inside it):**

| `secureValues` key  | Example value for this scenario    |
| ------------------- | ---------------------------------- |
| `basicAuthPassword` | `your_password`                    |
| `tlsCACert`         | `-----BEGIN CERTIFICATE-----\n...` |

**Notes:**

- `secureValues[].x-dsconfig-required-when` is a vendor extension carrying the field's `requiredWhen` expression through to the artifact. It is present for visibility and future tooling; nothing evaluates it yet (see your schema's "Known limitations").
- `settingsExamples.examples.simple.value` never includes a `secureJsonData` block — not because secrets are omitted from the example, but because `value` is an example of `spec`, and `spec` structurally excludes secrets (see "Before you start"). The table above is this guide's convention for showing the secret half of an example that `settingsExamples` itself has no field for; it is documentation, not part of the artifact.
- `basicAuthPassword` and `tlsCACert` appear in `secureValues` and **nowhere else** in this document. They are not duplicated under `spec` in any form.

---

## Use case 3 — HTTP headers (legacy indexed key/value pairs)

Grafana's existing storage for custom HTTP headers splits each header across two numbered fields: the name in `jsonData`, the value in `secureJsonData`. Model this with `valueType: "array"` plus a `storage.indexedPair` mapping — you do **not** need to invent numbered fields by hand in your schema; the array shape is what you author, and the mapping describes how that array corresponds to the legacy storage convention.

### 1. `dsconfig` schema

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

### 2. Grafana storage

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

### 3. App Platform schema artifact

```json
{
  "targetApiVersion": "v0alpha1",
  "settings": {
    "spec": {
      "required": ["url"],
      "properties": {
        "url": { "type": "string" },
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
      },
      "additionalProperties": false
    }
  },
  "settingsExamples": {
    "examples": {
      "simple": {
        "description": "Two custom headers",
        "value": {
          "url": "https://api.example.com",
          "jsonData": {
            "httpHeaders": [
              { "name": "X-Custom-Header" },
              { "name": "X-API-Token" }
            ]
          }
        }
      }
    }
  }
}
```

**Notes — read this one carefully, it is the sharpest edge in the whole guide:**

- The artifact's `spec.properties.jsonData.properties.httpHeaders` shows `name` **and** `value` as ordinary array-item properties, with no indication that `value` is a secret. This is a known, documented gap: today's SDK conversion routes fields to `secureValues` by each field's _own_ `target` only, and an `indexedPair`-mapped array's per-item secrecy (declared inside `storage`, not on the item field itself) is not yet propagated into the artifact. **Do not treat the absence of `value` from `secureValues` as evidence the header value isn't a secret — it is, per Grafana storage above; the artifact just doesn't say so yet.**
- Because of the same gap, the example payload above only fills in `name`, deliberately, to avoid implying header values belong in a plain example. If you need to show a worked example with header values, say so in `description`, not in `settingsExamples`.
- `validations[].itemCount.max` becomes `maxItems` on the array schema.
- To read this field's actual configured value back out of real Grafana storage (not the schema, the stored config), use `ResolveIndexedPairs` or `ResolveIndexedPairsAsMap` — see "Reading values back out by `id`," below. Don't try to read it with `ValueByID`; that function explicitly refuses `indexedPair`-mapped fields and points you at these two instead, since there's no single storage key to read for a field like this one.

---

## Use case 4 — TLS settings

TLS configuration is one of the most-repeated field sets across the plugin catalog: a client-auth toggle, a CA-cert toggle, a skip-verify flag, and the certificate/key material itself (always secret).

### 1. `dsconfig` schema

```json
{
  "schemaVersion": "v1",
  "pluginType": "example-tls",
  "pluginName": "TLS Example Datasource",
  "fields": [
    {
      "id": "connection.url",
      "key": "url",
      "valueType": "string",
      "target": "root",
      "required": true
    },
    {
      "id": "tls.skipVerify",
      "key": "tlsSkipVerify",
      "description": "Skip TLS certificate verification",
      "valueType": "boolean",
      "target": "jsonData",
      "defaultValue": false
    },
    {
      "id": "tls.authWithCACert",
      "key": "tlsAuthWithCACert",
      "description": "Authenticate the server using a custom CA certificate",
      "valueType": "boolean",
      "target": "jsonData",
      "defaultValue": false
    },
    {
      "id": "tls.clientAuth",
      "key": "tlsAuth",
      "description": "Authenticate Grafana to the server using a client certificate",
      "valueType": "boolean",
      "target": "jsonData",
      "defaultValue": false
    },
    {
      "id": "tls.serverName",
      "key": "serverName",
      "description": "Used to verify the hostname on the server's certificate",
      "valueType": "string",
      "target": "jsonData",
      "dependsOn": "tls.clientAuth == true"
    },
    {
      "id": "secure.tlsCACert",
      "key": "tlsCACert",
      "description": "PEM-encoded CA certificate",
      "valueType": "string",
      "target": "secureJsonData",
      "requiredWhen": "tls.authWithCACert == true",
      "ui": { "component": "textarea", "multiline": true, "rows": 6 }
    },
    {
      "id": "secure.tlsClientCert",
      "key": "tlsClientCert",
      "description": "PEM-encoded client certificate",
      "valueType": "string",
      "target": "secureJsonData",
      "requiredWhen": "tls.clientAuth == true",
      "ui": { "component": "textarea", "multiline": true, "rows": 6 }
    },
    {
      "id": "secure.tlsClientKey",
      "key": "tlsClientKey",
      "description": "PEM-encoded client private key",
      "valueType": "string",
      "target": "secureJsonData",
      "requiredWhen": "tls.clientAuth == true",
      "ui": { "component": "textarea", "multiline": true, "rows": 6 }
    }
  ]
}
```

### 2. Grafana storage

```json
{
  "name": "TLS Example DS",
  "type": "example-tls",
  "url": "https://secure.example.com",
  "jsonData": {
    "tlsSkipVerify": false,
    "tlsAuthWithCACert": true,
    "tlsAuth": true,
    "serverName": "secure.example.com"
  },
  "secureJsonData": {
    "tlsCACert": "-----BEGIN CERTIFICATE-----\n...",
    "tlsClientCert": "-----BEGIN CERTIFICATE-----\n...",
    "tlsClientKey": "-----BEGIN PRIVATE KEY-----\n..."
  }
}
```

### 3. App Platform schema artifact

```json
{
  "targetApiVersion": "v0alpha1",
  "settings": {
    "spec": {
      "required": ["url"],
      "properties": {
        "url": { "type": "string" },
        "jsonData": {
          "type": "object",
          "properties": {
            "tlsSkipVerify": { "type": "boolean", "default": false },
            "tlsAuthWithCACert": { "type": "boolean", "default": false },
            "tlsAuth": { "type": "boolean", "default": false },
            "serverName": {
              "type": "string",
              "x-dsconfig-depends-on": "tls.clientAuth == true"
            }
          }
        }
      },
      "additionalProperties": false
    },
    "secureValues": [
      {
        "key": "tlsCACert",
        "description": "PEM-encoded CA certificate",
        "x-dsconfig-required-when": "tls.authWithCACert == true"
      },
      {
        "key": "tlsClientCert",
        "description": "PEM-encoded client certificate",
        "x-dsconfig-required-when": "tls.clientAuth == true"
      },
      {
        "key": "tlsClientKey",
        "description": "PEM-encoded client private key",
        "x-dsconfig-required-when": "tls.clientAuth == true"
      }
    ]
  },
  "settingsExamples": {
    "examples": {
      "mutualTLS": {
        "description": "Mutual TLS with a custom CA. Requires secureValues to also be set — see note below.",
        "value": {
          "url": "https://secure.example.com",
          "jsonData": {
            "tlsAuthWithCACert": true,
            "tlsAuth": true,
            "serverName": "secure.example.com"
          }
        }
      }
    }
  }
}
```

**This example also requires:**

| `secureValues` key | Example value for this scenario    |
| ------------------ | ---------------------------------- |
| `tlsCACert`        | `-----BEGIN CERTIFICATE-----\n...` |
| `tlsClientCert`    | `-----BEGIN CERTIFICATE-----\n...` |
| `tlsClientKey`     | `-----BEGIN PRIVATE KEY-----\n...` |

**Notes:**

- All three certificate/key fields are correctly excluded from `spec` and present only in `secureValues` — unlike Use Case 3's header values, these are plain `secureJsonData` fields with a direct `target`, not values nested inside an `indexedPair` array, so the known propagation gap from Use Case 3 does not apply here. Per-field `target` is read correctly by the conversion; only the `indexedPair` case has the gap.
- `dependsOn` and `requiredWhen` both carry through as `x-dsconfig-*` vendor extensions. Neither is evaluated by anything today — they are present for forward compatibility and for tooling (including an assistant) that wants to reason about the condition itself, even without an evaluator.
- This field set (`tlsSkipVerify`, `tlsAuthWithCACert`, `tlsAuth`, `serverName`, `tlsCACert`, `tlsClientCert`, `tlsClientKey`) is consistent enough across the plugin catalog that you should use these exact `key` values when your plugin's existing storage already uses them, rather than inventing new names — this is what lets future tooling recognize "this is TLS configuration" across plugins.

---

## Use case 5 — Array of objects (no legacy mapping)

When your plugin already stores a field as a native JSON array — no legacy indexed-pair convention involved — describe it directly with `valueType: "array"` and an `item` schema. No `storage` mapping is needed; this is the array's actual stored shape already, exactly as Loki's derived-fields configuration stores it.

### 1. `dsconfig` schema

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
      "description": "Fields to extract from log lines and use to link to other datasources",
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
            "required": true,
            "validations": [{ "type": "length", "min": 1, "max": 500 }]
          },
          {
            "id": "derivedFields.item.url",
            "key": "url",
            "valueType": "string",
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

### 2. Grafana storage

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

### 3. App Platform schema artifact

```json
{
  "targetApiVersion": "v0alpha1",
  "settings": {
    "spec": {
      "required": ["url"],
      "properties": {
        "url": { "type": "string" },
        "jsonData": {
          "type": "object",
          "properties": {
            "derivedFields": {
              "type": "array",
              "description": "Fields to extract from log lines and use to link to other datasources",
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
                  "url": { "type": "string" },
                  "datasourceUid": { "type": "string" }
                }
              }
            }
          }
        }
      },
      "additionalProperties": false
    }
  },
  "settingsExamples": {
    "examples": {
      "traceLink": {
        "description": "Link a TraceID field to a Tempo datasource",
        "value": {
          "url": "https://loki.example.com",
          "jsonData": {
            "derivedFields": [
              {
                "name": "TraceID",
                "matcherRegex": "traceID=(\\w+)",
                "datasourceUid": "tempo-uid-123"
              }
            ]
          }
        }
      }
    }
  }
}
```

**Notes:**

- `datasourceUid` here is a plain string in this minimal example. If you want App Platform, an assistant, or a config editor to know this field specifically references another Grafana datasource (so it can offer a picker, or validate the UID actually resolves to a datasource of an expected plugin type), declare a `relationships` entry with `type: "datasourceReference"` and, optionally, `targetPluginType`, at the schema root. `relationships` is metadata only — it does not change anything shown in this artifact, but it's there for any consumer that looks for it.
- Compare this case to Use Case 3 directly: the difference is entirely whether a `storage` mapping is present. Same `valueType: "array"` and `item` shape; Use Case 3 needed `storage.indexedPair` because the real storage splits the array into numbered scalars, this case doesn't because the real storage is already a native array.

---

## Use case 6 — Nested JSON object (`section`)

Some plugins group several related `jsonData` fields under one nested object — for example, trace-to-logs linking configuration nested at `jsonData.tracesToLogs.*`. Use `section` to place fields under a one-level-deep object path within their `target`, without needing a separate `valueType: "object"` field wrapping them.

### 1. `dsconfig` schema

```json
{
  "schemaVersion": "v1",
  "pluginType": "example-tracing",
  "pluginName": "Tracing Example Datasource",
  "fields": [
    {
      "id": "connection.url",
      "key": "url",
      "valueType": "string",
      "target": "root",
      "required": true
    },
    {
      "id": "tracesToLogs.datasourceUid",
      "key": "datasourceUid",
      "description": "Datasource to link traces to logs with",
      "valueType": "string",
      "target": "jsonData",
      "section": "tracesToLogs"
    },
    {
      "id": "tracesToLogs.tags",
      "key": "tags",
      "description": "Tags to use for log/trace correlation",
      "valueType": "array",
      "target": "jsonData",
      "section": "tracesToLogs",
      "item": { "valueType": "string" }
    },
    {
      "id": "tracesToLogs.filterByTraceID",
      "key": "filterByTraceID",
      "valueType": "boolean",
      "target": "jsonData",
      "section": "tracesToLogs",
      "defaultValue": true
    }
  ],
  "relationships": [
    {
      "type": "datasourceReference",
      "fields": ["tracesToLogs.datasourceUid"],
      "description": "References the Loki/Tempo-family datasource to link traces to logs with"
    }
  ]
}
```

### 2. Grafana storage

```json
{
  "name": "Tracing Example DS",
  "type": "example-tracing",
  "url": "https://tracing.example.com",
  "jsonData": {
    "tracesToLogs": {
      "datasourceUid": "loki-uid-456",
      "tags": ["job", "instance"],
      "filterByTraceID": true
    }
  }
}
```

### 3. App Platform schema artifact

```json
{
  "targetApiVersion": "v0alpha1",
  "settings": {
    "spec": {
      "required": ["url"],
      "properties": {
        "url": { "type": "string" },
        "jsonData": {
          "type": "object",
          "properties": {
            "tracesToLogs": {
              "type": "object",
              "properties": {
                "datasourceUid": {
                  "type": "string",
                  "description": "Datasource to link traces to logs with"
                },
                "tags": {
                  "type": "array",
                  "description": "Tags to use for log/trace correlation",
                  "items": { "type": "string" }
                },
                "filterByTraceID": {
                  "type": "boolean",
                  "default": true
                }
              }
            }
          }
        }
      },
      "additionalProperties": false
    }
  },
  "settingsExamples": {
    "examples": {
      "simple": {
        "description": "Link traces to a Loki datasource by tag",
        "value": {
          "url": "https://tracing.example.com",
          "jsonData": {
            "tracesToLogs": {
              "datasourceUid": "loki-uid-456",
              "tags": ["job", "instance"]
            }
          }
        }
      }
    }
  }
}
```

**Notes:**

- Three fields sharing one `section` value (`"tracesToLogs"`) collapse into one nested object — `jsonData.tracesToLogs` — in both Grafana storage and the artifact. You declare each field flat, at the top of `fields`; the nesting is purely a result of `section` + `target` agreeing.
- `section` supports exactly one level of nesting today. If your plugin's real configuration needs `jsonData.a.b.c`, you cannot express that directly — flatten the deepest level you can, or wait for deeper-nesting support, rather than trying to chain `section` values.
- The `relationships` entry doesn't appear in this artifact at all. It's schema-level metadata for any consumer that explicitly looks for relationship declarations (a config editor wanting to render a datasource picker, an assistant wanting to know a field is a cross-reference) — it has no effect on the produced spec.

---

## Use case 7 — Real plugin: Sentry

The Grafana Sentry datasource plugin's actual configuration is small: a Sentry URL and org slug in `jsonData`, and an auth token in `secureJsonData`. This is a direct application of Use Case 2's pattern to a real, shipped plugin — shown here so you can see what a complete, real-world minimal schema looks like end to end.

### 1. `dsconfig` schema

```json
{
  "schemaVersion": "v1",
  "pluginType": "grafana-sentry-datasource",
  "pluginName": "Sentry",
  "docURL": "https://grafana.com/grafana/plugins/grafana-sentry-datasource/",
  "fields": [
    {
      "id": "connection.url",
      "key": "url",
      "label": "Sentry URL",
      "description": "Sentry URL to be used. If left blank, the default is https://sentry.io.",
      "valueType": "string",
      "target": "jsonData",
      "defaultValue": "https://sentry.io",
      "validations": [{ "type": "pattern", "pattern": "^https?://" }]
    },
    {
      "id": "connection.orgSlug",
      "key": "orgSlug",
      "label": "Sentry Org",
      "description": "Sentry organization slug, as it appears in https://sentry.io/organizations/{organization_slug}/",
      "valueType": "string",
      "target": "jsonData",
      "required": true
    },
    {
      "id": "auth.authToken",
      "key": "authToken",
      "label": "Sentry Auth Token",
      "description": "Internal integration token generated from Sentry's Developer Settings, with read access to Project, Issue and Event, and Organization.",
      "valueType": "string",
      "target": "secureJsonData",
      "required": true
    }
  ],
  "groups": [
    {
      "id": "connection",
      "title": "Connection",
      "fieldRefs": ["connection.url", "connection.orgSlug"]
    },
    {
      "id": "auth",
      "title": "Authentication",
      "fieldRefs": ["auth.authToken"]
    }
  ],
  "instructions": [
    {
      "msg": "Sentry Org is the organization slug found in the Sentry dashboard URL, not the organization's display name.",
      "tags": ["assistant"]
    },
    {
      "msg": "The auth token must be an internal integration token (Developer Settings > Custom Integrations > Internal Integration), not a personal API key, and needs Read permission on Project, Issue and Event, and Organization.",
      "tags": ["assistant"]
    }
  ]
}
```

### 2. Grafana storage

```json
{
  "name": "Sentry",
  "type": "grafana-sentry-datasource",
  "access": "proxy",
  "jsonData": {
    "url": "https://sentry.io",
    "orgSlug": "my-organization"
  },
  "secureJsonData": {
    "authToken": "your-internal-integration-token"
  }
}
```

### 3. App Platform schema artifact

```json
{
  "targetApiVersion": "v0alpha1",
  "settings": {
    "spec": {
      "required": ["orgSlug"],
      "properties": {
        "jsonData": {
          "type": "object",
          "required": ["orgSlug"],
          "properties": {
            "url": {
              "description": "Sentry URL to be used. If left blank, the default is https://sentry.io.",
              "type": "string",
              "default": "https://sentry.io",
              "pattern": "^https?://"
            },
            "orgSlug": {
              "description": "Sentry organization slug, as it appears in https://sentry.io/organizations/{organization_slug}/",
              "type": "string"
            }
          }
        }
      },
      "additionalProperties": false
    },
    "secureValues": [
      {
        "key": "authToken",
        "description": "Internal integration token generated from Sentry's Developer Settings, with read access to Project, Issue and Event, and Organization.",
        "required": true
      }
    ]
  },
  "settingsExamples": {
    "examples": {
      "default": {
        "description": "Sentry Cloud, default region. Requires secureValues to also be set — see note below.",
        "value": {
          "jsonData": {
            "orgSlug": "my-organization"
          }
        }
      },
      "selfHosted": {
        "description": "Self-hosted Sentry instance. Requires secureValues to also be set — see note below.",
        "value": {
          "jsonData": {
            "url": "https://sentry.mycompany.internal",
            "orgSlug": "my-organization"
          }
        }
      }
    }
  }
}
```

**Both examples above also require:**

| `secureValues` key | Example value                     |
| ------------------ | --------------------------------- |
| `authToken`        | `your-internal-integration-token` |

**Notes:**

- Note that `url` here targets `jsonData`, not `root` — unlike Use Case 1 and 2's `url` field. Sentry's real, existing storage puts its URL in `jsonData.url`, not as a root-level property. Always describe what your plugin _actually_ stores; do not assume every "URL" field belongs at `root` just because Use Case 1 put it there.
- `instructions` entries tagged `"assistant"` are exactly the kind of plugin-specific knowledge ("Org is a slug, not a display name"; "token must be this specific kind, with these specific scopes") that has no other home in the schema and that materially reduces the chance an assistant — or a person — fills in the wrong kind of value on the first attempt.
- Required-but-not-secret (`orgSlug`) and required-and-secret (`authToken`) are both expressed, correctly, in two different places: `orgSlug` appears in `spec.properties.jsonData.required`; `authToken`'s requiredness appears as `secureValues[].required: true`, not in any `required` array under `spec`.

---

## Use case 8 — Real plugin: GitHub (conditional, multi-mode authentication)

The Grafana GitHub datasource plugin supports two authentication modes, selected by an explicit discriminator field, `jsonData.selectedAuthType`: a personal access token, or a GitHub App (App ID + Installation ID + private key). It also has an Enterprise Server URL field that only matters for self-hosted GitHub. This is a real-world, Pattern A discriminator case — exactly the structure described generally in Use Case 2's `requiredWhen`, applied to a plugin where the discriminator genuinely has more than two branches and each branch requires a different secret.

### 1. `dsconfig` schema

```json
{
  "schemaVersion": "v1",
  "pluginType": "grafana-github-datasource",
  "pluginName": "GitHub",
  "docURL": "https://grafana.com/docs/plugins/grafana-github-datasource/latest/configure/",
  "fields": [
    {
      "id": "auth.selectedAuthType",
      "key": "selectedAuthType",
      "label": "Authentication Type",
      "valueType": "string",
      "target": "jsonData",
      "required": true,
      "defaultValue": "personal-access-token",
      "validations": [
        {
          "type": "allowedValues",
          "values": ["personal-access-token", "github-app"]
        }
      ],
      "ui": {
        "component": "radio",
        "options": [
          {
            "label": "Personal Access Token",
            "value": "personal-access-token"
          },
          { "label": "GitHub App", "value": "github-app" }
        ]
      }
    },
    {
      "id": "auth.accessToken",
      "key": "accessToken",
      "label": "Personal Access Token",
      "description": "Classic or fine-grained personal access token with the required repository scopes.",
      "valueType": "string",
      "target": "secureJsonData",
      "dependsOn": "auth.selectedAuthType == 'personal-access-token'",
      "requiredWhen": "auth.selectedAuthType == 'personal-access-token'"
    },
    {
      "id": "auth.appId",
      "key": "appId",
      "label": "App ID",
      "valueType": "string",
      "target": "jsonData",
      "dependsOn": "auth.selectedAuthType == 'github-app'",
      "requiredWhen": "auth.selectedAuthType == 'github-app'"
    },
    {
      "id": "auth.installationId",
      "key": "installationId",
      "label": "Installation ID",
      "valueType": "string",
      "target": "jsonData",
      "dependsOn": "auth.selectedAuthType == 'github-app'",
      "requiredWhen": "auth.selectedAuthType == 'github-app'"
    },
    {
      "id": "auth.privateKey",
      "key": "privateKey",
      "label": "Private Key",
      "description": "Private key generated for the GitHub App, in PEM format.",
      "valueType": "string",
      "target": "secureJsonData",
      "dependsOn": "auth.selectedAuthType == 'github-app'",
      "requiredWhen": "auth.selectedAuthType == 'github-app'",
      "ui": { "component": "textarea", "multiline": true, "rows": 6 }
    },
    {
      "id": "connection.isEnterprise",
      "key": "isEnterprise",
      "label": "GitHub Enterprise Server",
      "description": "Enable if connecting to a self-hosted GitHub Enterprise Server instance rather than github.com.",
      "valueType": "boolean",
      "target": "jsonData",
      "defaultValue": false
    },
    {
      "id": "connection.githubUrl",
      "key": "githubUrl",
      "label": "GitHub Enterprise Server URL",
      "description": "Base URL of the GitHub Enterprise Server instance, e.g. https://github.mycompany.com",
      "valueType": "string",
      "target": "jsonData",
      "dependsOn": "connection.isEnterprise == true",
      "requiredWhen": "connection.isEnterprise == true",
      "validations": [{ "type": "pattern", "pattern": "^https?://" }]
    }
  ],
  "groups": [
    {
      "id": "auth",
      "title": "Authentication",
      "fieldRefs": [
        "auth.selectedAuthType",
        "auth.accessToken",
        "auth.appId",
        "auth.installationId",
        "auth.privateKey"
      ]
    },
    {
      "id": "connection",
      "title": "Connection",
      "fieldRefs": ["connection.isEnterprise", "connection.githubUrl"],
      "optional": true
    }
  ],
  "instructions": [
    {
      "msg": "GitHub App authentication is recommended over personal access tokens for organizational use: it provides fine-grained, repository-scoped access and short-lived tokens that rotate automatically.",
      "tags": ["assistant"]
    },
    {
      "msg": "Only ask for the Enterprise Server URL if the person says they use GitHub Enterprise Server (self-hosted). Otherwise leave isEnterprise false and skip githubUrl entirely.",
      "tags": ["assistant"]
    }
  ]
}
```

### 2. Grafana storage

Personal access token mode:

```json
{
  "name": "GitHub (Personal Access Token)",
  "type": "grafana-github-datasource",
  "jsonData": {
    "selectedAuthType": "personal-access-token"
  },
  "secureJsonData": {
    "accessToken": "github_pat_xxxxxxxxxxxxxxxxxxxxxx"
  }
}
```

GitHub App mode:

```json
{
  "name": "GitHub (App)",
  "type": "grafana-github-datasource",
  "jsonData": {
    "selectedAuthType": "github-app",
    "appId": "123456",
    "installationId": "78901234"
  },
  "secureJsonData": {
    "privateKey": "-----BEGIN RSA PRIVATE KEY-----\n..."
  }
}
```

### 3. App Platform schema artifact

```json
{
  "targetApiVersion": "v0alpha1",
  "settings": {
    "spec": {
      "properties": {
        "jsonData": {
          "type": "object",
          "required": ["selectedAuthType"],
          "properties": {
            "selectedAuthType": {
              "type": "string",
              "default": "personal-access-token",
              "enum": ["personal-access-token", "github-app"]
            },
            "appId": {
              "type": "string",
              "x-dsconfig-depends-on": "auth.selectedAuthType == 'github-app'",
              "x-dsconfig-required-when": "auth.selectedAuthType == 'github-app'"
            },
            "installationId": {
              "type": "string",
              "x-dsconfig-depends-on": "auth.selectedAuthType == 'github-app'",
              "x-dsconfig-required-when": "auth.selectedAuthType == 'github-app'"
            },
            "isEnterprise": {
              "type": "boolean",
              "default": false
            },
            "githubUrl": {
              "type": "string",
              "pattern": "^https?://",
              "x-dsconfig-depends-on": "connection.isEnterprise == true",
              "x-dsconfig-required-when": "connection.isEnterprise == true"
            }
          }
        }
      },
      "additionalProperties": false
    },
    "secureValues": [
      {
        "key": "accessToken",
        "description": "Classic or fine-grained personal access token with the required repository scopes.",
        "x-dsconfig-depends-on": "auth.selectedAuthType == 'personal-access-token'",
        "x-dsconfig-required-when": "auth.selectedAuthType == 'personal-access-token'"
      },
      {
        "key": "privateKey",
        "description": "Private key generated for the GitHub App, in PEM format.",
        "x-dsconfig-depends-on": "auth.selectedAuthType == 'github-app'",
        "x-dsconfig-required-when": "auth.selectedAuthType == 'github-app'"
      }
    ]
  },
  "settingsExamples": {
    "examples": {
      "personalAccessToken": {
        "description": "github.com with a personal access token. Requires secureValues to also be set — see note below.",
        "value": {
          "jsonData": {
            "selectedAuthType": "personal-access-token"
          }
        }
      },
      "githubApp": {
        "description": "github.com with a GitHub App. Requires secureValues to also be set — see note below.",
        "value": {
          "jsonData": {
            "selectedAuthType": "github-app",
            "appId": "123456",
            "installationId": "78901234"
          }
        }
      },
      "enterpriseServer": {
        "description": "Self-hosted GitHub Enterprise Server with a personal access token. Requires secureValues to also be set — see note below.",
        "value": {
          "jsonData": {
            "selectedAuthType": "personal-access-token",
            "isEnterprise": true,
            "githubUrl": "https://github.mycompany.com"
          }
        }
      }
    }
  }
}
```

**Each example above also requires its own, mode-specific `secureValues` entry — the two auth modes never need the same secret:**

| Example               | `secureValues` key | Example value                          |
| --------------------- | ------------------ | -------------------------------------- |
| `personalAccessToken` | `accessToken`      | `github_pat_xxxxxxxxxxxxxxxxxxxxxx`    |
| `githubApp`           | `privateKey`       | `-----BEGIN RSA PRIVATE KEY-----\n...` |
| `enterpriseServer`    | `accessToken`      | `github_pat_xxxxxxxxxxxxxxxxxxxxxx`    |

**Notes:**

- This is the cleanest real-world example of Pattern A (explicit discriminator field) from the earlier design discussion: `selectedAuthType` is one field, with an `allowedValues`/`enum` of exactly two modes, and every other auth field's `dependsOn`/`requiredWhen` is keyed off its value.
- Because nothing in this release evaluates `dependsOn`/`requiredWhen`, a config editor or assistant consuming this artifact must still implement "only show/require `appId`/`installationId`/`privateKey` when `selectedAuthType == 'github-app'`" as its own logic, reading the `x-dsconfig-*` extensions as a hint rather than relying on the schema to enforce it. This is the direct, practical cost of the "no expression evaluation" limitation, applied to a real plugin rather than discussed abstractly.
- `accessToken` and `privateKey` are mutually exclusive in practice (only one is ever set, depending on `selectedAuthType`), but nothing in this schema version _enforces_ that mutual exclusivity — both simply carry the matching `dependsOn`/`requiredWhen` conditions, and it is the consuming tool's responsibility to apply them. There is no field-level mechanism today that says "exactly one of these two secrets must be set." The table above shows this concretely: never both secrets at once, for any one example.
- Three `settingsExamples` entries are shown, one per real configuration mode, deliberately — for a plugin with more than one valid "shape" of configuration, a single example is misleading; show one example per mode a person might actually choose.

---

## Use case 9 — Real plugin: BigQuery (three-way authentication, including identity forwarding)

The Grafana BigQuery datasource plugin's authentication is selected by `jsonData.authenticationType`, with three real values: `jwt` (a GCP service account JSON key, with the private key in `secureJsonData`), `gce` (use the GCE metadata server's default credentials — no explicit credential fields at all), and `forwardOAuthIdentity` (forward the logged-in Grafana user's own OAuth identity). The third mode is the real-world instance of the "is this actually an auth-type variant, or a different axis (identity forwarding)" distinction raised earlier — BigQuery's own field structure treats it as a fourth `authenticationType` value, which this schema describes faithfully as that plugin's actual existing convention, not as a redesign.

### 1. `dsconfig` schema

```json
{
  "schemaVersion": "v1",
  "pluginType": "grafana-bigquery-datasource",
  "pluginName": "Google BigQuery",
  "docURL": "https://grafana.com/docs/plugins/grafana-bigquery-datasource/latest/configure/",
  "fields": [
    {
      "id": "auth.authenticationType",
      "key": "authenticationType",
      "label": "Authentication Type",
      "valueType": "string",
      "target": "jsonData",
      "required": true,
      "defaultValue": "jwt",
      "validations": [
        {
          "type": "allowedValues",
          "values": ["jwt", "gce", "forwardOAuthIdentity"]
        }
      ],
      "ui": {
        "component": "select",
        "options": [
          { "label": "GCP Service Account (JWT)", "value": "jwt" },
          { "label": "GCE Default Credentials", "value": "gce" },
          { "label": "Forward OAuth Identity", "value": "forwardOAuthIdentity" }
        ]
      }
    },
    {
      "id": "auth.clientEmail",
      "key": "clientEmail",
      "label": "Service Account Email",
      "valueType": "string",
      "target": "jsonData",
      "dependsOn": "auth.authenticationType == 'jwt'",
      "requiredWhen": "auth.authenticationType == 'jwt'"
    },
    {
      "id": "auth.tokenUri",
      "key": "tokenUri",
      "label": "Token URI",
      "valueType": "string",
      "target": "jsonData",
      "defaultValue": "https://oauth2.googleapis.com/token",
      "dependsOn": "auth.authenticationType == 'jwt'",
      "requiredWhen": "auth.authenticationType == 'jwt'"
    },
    {
      "id": "auth.privateKey",
      "key": "privateKey",
      "label": "Private Key",
      "description": "Private key from the GCP service account JSON key file.",
      "valueType": "string",
      "target": "secureJsonData",
      "dependsOn": "auth.authenticationType == 'jwt'",
      "requiredWhen": "auth.authenticationType == 'jwt'",
      "ui": { "component": "textarea", "multiline": true, "rows": 6 }
    },
    {
      "id": "auth.usingImpersonation",
      "key": "usingImpersonation",
      "label": "Impersonate Service Account",
      "valueType": "boolean",
      "target": "jsonData",
      "defaultValue": false,
      "dependsOn": "auth.authenticationType == 'gce'"
    },
    {
      "id": "auth.serviceAccountToImpersonate",
      "key": "serviceAccountToImpersonate",
      "label": "Service Account To Impersonate",
      "valueType": "string",
      "target": "jsonData",
      "dependsOn": "auth.usingImpersonation == true",
      "requiredWhen": "auth.usingImpersonation == true"
    },
    {
      "id": "connection.defaultProject",
      "key": "defaultProject",
      "label": "Default Project",
      "description": "The GCP project queries run against by default.",
      "valueType": "string",
      "target": "jsonData",
      "required": true
    },
    {
      "id": "connection.processingLocation",
      "key": "processingLocation",
      "label": "Processing Location",
      "valueType": "string",
      "target": "jsonData",
      "defaultValue": "US"
    },
    {
      "id": "connection.maxBytesBilled",
      "key": "MaxBytesBilled",
      "label": "Max Bytes Billed",
      "valueType": "number",
      "target": "jsonData",
      "validations": [{ "type": "range", "min": 0 }]
    },
    {
      "id": "connection.serviceEndpoint",
      "key": "serviceEndpoint",
      "label": "Service Endpoint",
      "valueType": "string",
      "target": "jsonData",
      "defaultValue": "https://bigquery.googleapis.com/bigquery/v2/"
    }
  ],
  "groups": [
    {
      "id": "auth",
      "title": "Authentication",
      "fieldRefs": [
        "auth.authenticationType",
        "auth.clientEmail",
        "auth.tokenUri",
        "auth.privateKey",
        "auth.usingImpersonation",
        "auth.serviceAccountToImpersonate"
      ]
    },
    {
      "id": "connection",
      "title": "Connection",
      "fieldRefs": [
        "connection.defaultProject",
        "connection.processingLocation",
        "connection.maxBytesBilled",
        "connection.serviceEndpoint"
      ],
      "optional": true
    }
  ],
  "instructions": [
    {
      "msg": "Forward OAuth Identity only works if Grafana itself is configured with a generic OAuth provider that also grants BigQuery access to the logged-in user — it is not a credential the person configuring the datasource provides directly.",
      "tags": ["assistant"]
    },
    {
      "msg": "GCE Default Credentials only works when Grafana itself is running on a Google Compute Engine virtual machine with an attached service account. Do not offer this mode unless the person has confirmed that is their deployment.",
      "tags": ["assistant"]
    }
  ]
}
```

### 2. Grafana storage

JWT service account mode:

```json
{
  "name": "BigQuery",
  "type": "grafana-bigquery-datasource",
  "jsonData": {
    "authenticationType": "jwt",
    "clientEmail": "grafana-reader@my-project.iam.gserviceaccount.com",
    "tokenUri": "https://oauth2.googleapis.com/token",
    "defaultProject": "my-project",
    "processingLocation": "US"
  },
  "secureJsonData": {
    "privateKey": "-----BEGIN PRIVATE KEY-----\n..."
  }
}
```

GCE default credentials mode:

```json
{
  "name": "BigQuery (GCE)",
  "type": "grafana-bigquery-datasource",
  "jsonData": {
    "authenticationType": "gce",
    "defaultProject": "my-project"
  }
}
```

Forward OAuth identity mode:

```json
{
  "name": "BigQuery (User Identity)",
  "type": "grafana-bigquery-datasource",
  "jsonData": {
    "authenticationType": "forwardOAuthIdentity",
    "defaultProject": "my-project",
    "oauthPassThru": true
  }
}
```

### 3. App Platform schema artifact

```json
{
  "targetApiVersion": "v0alpha1",
  "settings": {
    "spec": {
      "properties": {
        "jsonData": {
          "type": "object",
          "required": ["authenticationType", "defaultProject"],
          "properties": {
            "authenticationType": {
              "type": "string",
              "default": "jwt",
              "enum": ["jwt", "gce", "forwardOAuthIdentity"]
            },
            "clientEmail": {
              "type": "string",
              "x-dsconfig-depends-on": "auth.authenticationType == 'jwt'",
              "x-dsconfig-required-when": "auth.authenticationType == 'jwt'"
            },
            "tokenUri": {
              "type": "string",
              "default": "https://oauth2.googleapis.com/token",
              "x-dsconfig-depends-on": "auth.authenticationType == 'jwt'",
              "x-dsconfig-required-when": "auth.authenticationType == 'jwt'"
            },
            "usingImpersonation": {
              "type": "boolean",
              "default": false,
              "x-dsconfig-depends-on": "auth.authenticationType == 'gce'"
            },
            "serviceAccountToImpersonate": {
              "type": "string",
              "x-dsconfig-depends-on": "auth.usingImpersonation == true",
              "x-dsconfig-required-when": "auth.usingImpersonation == true"
            },
            "defaultProject": {
              "type": "string",
              "description": "The GCP project queries run against by default."
            },
            "processingLocation": {
              "type": "string",
              "default": "US"
            },
            "MaxBytesBilled": {
              "type": "number",
              "minimum": 0
            },
            "serviceEndpoint": {
              "type": "string",
              "default": "https://bigquery.googleapis.com/bigquery/v2/"
            }
          }
        }
      },
      "additionalProperties": false
    },
    "secureValues": [
      {
        "key": "privateKey",
        "description": "Private key from the GCP service account JSON key file.",
        "x-dsconfig-depends-on": "auth.authenticationType == 'jwt'",
        "x-dsconfig-required-when": "auth.authenticationType == 'jwt'"
      }
    ]
  },
  "settingsExamples": {
    "examples": {
      "serviceAccount": {
        "description": "GCP service account JWT credentials. Requires secureValues to also be set — see note below.",
        "value": {
          "jsonData": {
            "authenticationType": "jwt",
            "clientEmail": "grafana-reader@my-project.iam.gserviceaccount.com",
            "defaultProject": "my-project"
          }
        }
      },
      "gceDefault": {
        "description": "Default credentials from the GCE metadata server. No secureValues entry applies to this mode — see note below.",
        "value": {
          "jsonData": {
            "authenticationType": "gce",
            "defaultProject": "my-project"
          }
        }
      },
      "forwardIdentity": {
        "description": "Forward the logged-in Grafana user's own OAuth identity. No secureValues entry applies to this mode — see note below.",
        "value": {
          "jsonData": {
            "authenticationType": "forwardOAuthIdentity",
            "defaultProject": "my-project"
          }
        }
      }
    }
  }
}
```

**Only the `serviceAccount` example requires a `secureValues` entry — this is the one use case in this guide where most example modes genuinely need no secret at all:**

| Example           | `secureValues` key | Example value                                                                                                 |
| ----------------- | ------------------ | ------------------------------------------------------------------------------------------------------------- |
| `serviceAccount`  | `privateKey`       | `-----BEGIN PRIVATE KEY-----\n...`                                                                            |
| `gceDefault`      | _none_             | Credentials come from the GCE metadata server, not from anything Grafana stores.                              |
| `forwardIdentity` | _none_             | Credentials come from the logged-in user's own session, not from anything Grafana stores for this datasource. |

**Notes:**

- This schema deliberately models `forwardOAuthIdentity` as a third value of the same `authenticationType` enum, matching the plugin's real, existing field structure exactly — even though the general design discussion earlier distinguished "how Grafana authenticates to the upstream API" from "whether to forward the user's identity" as two different conceptual axes. **Authoring a schema is about describing what a plugin actually does, not about correcting its design.** If you are designing a _new_ plugin from scratch, treating identity-forwarding as a separate, orthogonal field (as `oauthPassThru` is treated as its own boolean field in several other plugins) is the better target shape — see Use Case 4's TLS pattern for the general principle of keeping independent concerns in independent fields. But for an existing plugin like BigQuery, describe its real three-way enum as-is.
- `oauthPassThru` appears in the real Grafana storage payload for the forward-identity mode but is **not** declared as a field in this schema. This is intentional: `oauthPassThru` is a generic, cross-plugin mechanism Grafana's HTTP client settings already provide (see `DataSourceHttpSettings`'s "Forward OAuth Identity" toggle), not a BigQuery-specific configuration field, so it does not need its own `dsconfig` field entry — your schema only needs to describe fields specific to your plugin's own configuration.
- `usingImpersonation`'s `dependsOn` references `auth.authenticationType == 'gce'`, not `'jwt'` — service account impersonation is a `gce`-mode-only feature for this plugin, per its real documented behavior. Always check the actual conditional relationships in your plugin's existing behavior rather than assuming a pattern from a different plugin's schema carries over.
- `MaxBytesBilled`'s `key` is capitalized (`"MaxBytesBilled"`, not `"maxBytesBilled"`) because that is the real, existing storage key for this plugin. `key` must always match what is actually persisted today, even when it breaks a camelCase convention used elsewhere in the same schema — `key` is not a place to "clean up" a plugin's existing field names.
- This is a useful reminder that not every secret-capable schema needs a secret in every one of its modes — `gce` and `forwardOAuthIdentity` are both real, valid, fully-working configurations with an empty `secureJsonData`, which is exactly why `secureValues` entries (Section 6.1's `requiredWhen`/`dependsOn` extensions) are per-field conditional rather than globally required.

---

## Reading values back out by `id`

Everything above is about authoring a schema and seeing what it produces. This section is the reverse direction: given an `id` and a real, already-saved configuration payload, how do you get the actual configured value back out — for a config editor rendering a field, an assistant answering "what's the current timeout set to," or any tooling resolving a `groups[].fieldRefs` entry to something concrete? Both reference implementations ship the same four functions for this, under matching names (Go: `FieldByID`, `ValueByID`, `ResolveIndexedPairs`, `ResolveIndexedPairsAsMap`; TypeScript: `fieldById`, `valueById`, `resolveIndexedPairs`, `resolveIndexedPairsAsMap`).

### Direct fields — `ValueByID`

For any field with no `storage` mapping (or `storage.type: "direct"`), `ValueByID` resolves an `id` straight to its value, using the field's `target`/`section`/`key` to know where to look. Using Use Case 7's Sentry schema:

```go
settings := map[string]any{
    "jsonData": map[string]any{
        "url":     "https://sentry.io",
        "orgSlug": "my-organization",
    },
}

val, err := schema.ValueByID("connection.orgSlug", settings)
// val == "my-organization"
```

This is the common case and covers the large majority of fields in any schema in this guide. It refuses, with a clear error rather than a guess, for a virtual field, an item field, or a field with a non-direct `storage` mapping — including, notably, `httpHeaders` from Use Case 3, which is exactly the case the next two functions exist for.

### Indexed-pair fields — `ResolveIndexedPairs` and `ResolveIndexedPairsAsMap`

For an `indexedPair`-mapped field like Use Case 3's `httpHeaders`, there's no single storage key to read — the real value is scattered across `jsonData.httpHeaderName1`, `httpHeaderName2`, ... and the matching `secureJsonData.httpHeaderValue{N}` entries. These two functions assemble that into the field's actual logical shape:

```go
jsonData := map[string]any{
    "httpHeaderName1": "X-Custom-Header",
    "httpHeaderName2": "X-API-Token",
}
secureJsonData := map[string]any{
    "httpHeaderValue1": "custom-value",
    "httpHeaderValue2": "your-api-token",
}

pairs, err := dsconfig.ResolveIndexedPairs(headersField, jsonData, secureJsonData)
// pairs == []map[string]any{
//   {"name": "X-Custom-Header", "value": "custom-value"},
//   {"name": "X-API-Token",     "value": "your-api-token"},
// }

flat, err := dsconfig.ResolveIndexedPairsAsMap(headersField, jsonData, secureJsonData)
// flat == map[string]string{
//   "X-Custom-Header": "custom-value",
//   "X-API-Token":      "your-api-token",
// }
```

**Pick based on which trade-off you can tolerate, not by which one looks simpler:**

|                                     | `ResolveIndexedPairs` (array)                                      | `ResolveIndexedPairsAsMap` (flat map)                                                   |
| ----------------------------------- | ------------------------------------------------------------------ | --------------------------------------------------------------------------------------- |
| Stops at the first gap in numbering | Yes — a deleted-but-not-renumbered entry hides everything after it | No — scans every key present, regardless of gaps                                        |
| Two pairs with the same name        | Both kept, as separate entries                                     | Collapsed into one; which one wins is unspecified                                       |
| Missing value                       | Entry has no `value` key at all                                    | Entry's value is `""` — indistinguishable from a value that's genuinely an empty string |

For HTTP headers specifically, the map shape is usually what you want — duplicate header names are rare and gap-tolerance is a real, documented bug class worth not reproducing. For a hypothetical future indexed-pair use case where duplicate names are expected and meaningful (Infinity-style repeated query-string parameters, for example, where `?tag=a&tag=b` is a real and different thing from `?tag=a`), reach for the array version instead.

### What none of these four functions do

They don't evaluate `dependsOn`, `requiredWhen`, `disabledWhen`, or `effects[].when` — those remain unevaluated expression strings regardless of which lookup function you use (see "Common mistakes to avoid," below). They also can't read a secret value out of a _live, already-saved_ datasource's settings: `secureJsonData` is write-only once saved, so calling any of these against a real running datasource's settings for a `secureJsonData`-targeted field returns the same "not present" result whether the field was never configured or whether it's simply unreadable post-save. These functions work exactly as shown against a schema's own example payloads, or any payload you construct directly yourself with the secret values included — they just can't get a secret value back from Grafana once it's actually been saved there for real.

## Quick reference: choosing the right pattern

| Your plugin's situation                                                                          | Use this pattern                                                                              | See               |
| ------------------------------------------------------------------------------------------------ | --------------------------------------------------------------------------------------------- | ----------------- |
| One simple field at the top level                                                                | Plain field, `target: "root"`                                                                 | Use Case 1        |
| A typical mix of root + jsonData + a couple of secrets                                           | Plain fields, three different `target` values                                                 | Use Case 2        |
| Custom headers stored as `httpHeaderName1`/`httpHeaderValue1`-style pairs                        | `valueType: "array"` + `storage.indexedPair`                                                  | Use Case 3        |
| TLS client/CA certs and related toggles                                                          | The standard TLS field set (Use Case 4's exact `key` names, if your plugin already uses them) | Use Case 4        |
| A field that's already a native JSON array in storage                                            | `valueType: "array"` + `item`, no `storage` mapping                                           | Use Case 5        |
| Several fields that belong under one nested `jsonData` sub-object                                | `section`                                                                                     | Use Case 6        |
| A field references another datasource's UID                                                      | `relationships[].type: "datasourceReference"`                                                 | Use Case 6        |
| An auth method selected by one explicit field, with each mode requiring different fields/secrets | `allowedValues` + `dependsOn`/`requiredWhen` keyed to the discriminator's value               | Use Cases 8, 9    |
| Independent, non-exclusive boolean auth/TLS toggles (no single discriminator field)              | Plain boolean fields, each with its own `dependsOn` as needed                                 | Use Case 4        |
| Guidance a config editor or assistant needs but that doesn't fit any field                       | `instructions`, tagged `"assistant"` by convention                                            | Use Cases 7, 8, 9 |

## Common mistakes to avoid

- **Inventing a field name instead of using your plugin's real, existing one.** `key` must equal the literal property name already in storage. If you're not sure, check your plugin's provisioning documentation or its Go settings struct before writing the schema.
- **Putting a secret under `jsonData` instead of `secureJsonData`.** If the value is a password, token, key, or anything else that shouldn't be readable after save, it must target `secureJsonData`, with no exception.
- **Assuming `dependsOn`/`requiredWhen` are enforced.** They are not evaluated by any component today. Writing them is still worthwhile — they document real conditional relationships and travel into the App Platform artifact as `x-dsconfig-*` extensions — but do not rely on them to actually block an invalid save in this schema version.
- **Treating `ui.options` as the data contract.** It isn't. If a field has an enumerated set of valid values, declare `validations[].type: "allowedValues"`. `ui.options` is presentation only and should be kept consistent with `validations`, not used in its place.
- **Putting per-item secrecy only inside a `storage.indexedPair` mapping and assuming it's reflected in the artifact.** It currently is not (Use Case 3). State the secrecy in your field's `description` as well, until this gap closes.
- **Chaining `section` values to express more than one level of nesting.** It supports exactly one level. Flatten deeper structures, or wait for deeper-nesting support, rather than trying to nest `section` paths.
