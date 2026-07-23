# Supabase configuration

How to configure the **Supabase** data source (`grafana-supabase-datasource`) in Grafana.

> This page is generated from [`dsconfig.json`](dsconfig.json). Do not edit it by hand — run `go generate ./...` to refresh.

## Configuration sections

- [Authentication](#authentication)

## Authentication

### Supabase personal token

_optional · select_

Supabase personal token.

| | |
|---|---|
| Default | `mgmt_bearer` |
| Allowed values | `mgmt_bearer` (Supabase personal token) |

### Token

_🔒 secret (write-only) · conditionally required · string_

Token for accessing the datasource API.

| | |
|---|---|
| Example | `Token value` |
| Shown when | **Supabase personal token** is **Supabase personal token** (`mgmt_bearer`) |

