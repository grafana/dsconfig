# Sentry configuration

How to configure the **Sentry** data source (`grafana-sentry-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/grafana/plugins/grafana-sentry-datasource/).

> This page is generated from [`dsconfig.json`](dsconfig.json). Do not edit it by hand — run `go generate ./...` to refresh.

## Configuration sections

- [Sentry Settings](#sentry-settings)
- [Additional settings](#additional-settings) — _optional_

## Sentry Settings

### Sentry URL

_optional · string_

Sentry URL to be used. If left blank, https://sentry.io will be used.

| | |
|---|---|
| Default | `https://sentry.io` |
| Example | `https://sentry.io` |

### Sentry Org

_**required** · string_

Sentry Org slug. Typically this will be the last segment of the URL: https://sentry.io/organizations/{organization_slug}/ - only the slug should be entered here.

| | |
|---|---|
| Example | `Sentry org slug` |

### Sentry Auth Token

_🔒 secret (write-only) · **required** · string_

Sentry authentication token. Auth tokens can be created from https://sentry.io/settings/{organization_slug}/developer-settings.

| | |
|---|---|
| Example | `Sentry Authentication Token` |

## Additional settings

_This section is optional._

### Skip TLS Verify

_optional · toggle_

Skip TLS certificate verification. Use this option for self-hosted Sentry instances with self-signed certificates.

| | |
|---|---|
| Default | `false` |

