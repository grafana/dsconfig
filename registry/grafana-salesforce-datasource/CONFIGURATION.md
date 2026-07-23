# Salesforce configuration

How to configure the **Salesforce** data source (`grafana-salesforce-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-salesforce-datasource).

> This page is generated from [`dsconfig.json`](dsconfig.json). Do not edit it by hand — run `go generate ./...` to refresh.

## Configuration sections

- [Connection Settings](#connection-settings)
- [Optional Settings](#optional-settings) — _optional_

## Connection Settings

### Authentication

_optional · radio_

| | |
|---|---|
| Default | `user` |
| Allowed values | `user` (Credentials), `jwt` (JWT) |

### User Name

_conditionally required · string_

| | |
|---|---|
| Example | `Salesforce User` |
| Required when | **Authentication** is **Credentials** (`user`) |

### Password

_🔒 secret (write-only) · conditionally required · string_

| | |
|---|---|
| Example | `Salesforce Password` |
| Shown when | **Authentication** is **Credentials** (`user`) |

### Security Token

_🔒 secret (write-only) · optional · string_

| | |
|---|---|
| Example | `Salesforce Security Token` |
| Shown when | **Authentication** is **Credentials** (`user`) |

### Consumer Key

_🔒 secret (write-only) · conditionally required · string_

| | |
|---|---|
| Example | `Connected App Consumer Key` |
| Required when | **Authentication** is **Credentials** (`user`) |

### Consumer Secret

_🔒 secret (write-only) · conditionally required · string_

| | |
|---|---|
| Example | `Connected App Consumer Secret` |
| Shown when | **Authentication** is **Credentials** (`user`) |

### Certificate

_🔒 secret (write-only) · conditionally required · multiline text_

| | |
|---|---|
| Example | `-----BEGIN CERTIFICATE-----` |
| Shown when | **Authentication** is **JWT** (`jwt`) |

### Private Key

_🔒 secret (write-only) · conditionally required · multiline text_

| | |
|---|---|
| Example | `-----BEGIN PRIVATE KEY-----` |
| Shown when | **Authentication** is **JWT** (`jwt`) |

## Optional Settings

_This section is optional._

### Environment

_optional · select_

| | |
|---|---|
| Default | `https://login.salesforce.com` |
| Allowed values | `https://login.salesforce.com` (Production), `https://test.salesforce.com` (SandBox) |

## Other settings

### sandbox

_optional · boolean_

Legacy boolean that selects the Salesforce login/token host when `tokenUrl` is empty (`true` → https://test.salesforce.com, `false` → https://login.salesforce.com). Deprecated in favor of `tokenUrl`; the config editor no longer writes it but reads it to derive the initial Environment selection, and the backend still honors it for backwards compatibility and provisioning.

| | |
|---|---|
| Default | `false` |

