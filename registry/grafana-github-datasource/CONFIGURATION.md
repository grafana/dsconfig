# GitHub configuration

How to configure the **GitHub** data source (`grafana-github-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-github-datasource).

> This page is generated from [`dsconfig.json`](dsconfig.json). Do not edit it by hand — run `go generate ./...` to refresh.

## Configuration sections

- [Connection](#connection) — _optional_
- [Authentication](#authentication)

## Connection

_This section is optional._

### GitHub License Type

_optional · radio_

| | |
|---|---|
| Default | `github-basic` |
| Allowed values | `github-basic` (Free, Pro & Team), `github-enterprise-cloud` (Enterprise Cloud), `github-enterprise-server` (Enterprise Server) |

### GitHub Enterprise Server URL

_conditionally required · string_

| | |
|---|---|
| Example | `http(s)://HOSTNAME/` |
| Shown when | **GitHub License Type** is **Enterprise Server** (`github-enterprise-server`) |
| Required when | **githubPlan** is `github-enterprise-server` |

## Authentication

### Authentication Type

_optional · radio_

| | |
|---|---|
| Default | `personal-access-token` |
| Allowed values | `personal-access-token` (Personal Access Token), `github-app` (GitHub App) |

### Personal Access Token

_🔒 secret (write-only) · conditionally required · string_

| | |
|---|---|
| Example | `Personal Access Token` |
| Shown when | **Authentication Type** is **Personal Access Token** (`personal-access-token`) |

**Access Token & Permissions**

#### How to create a access token

To create a new fine grained access token, navigate to [Personal Access Tokens](https://github.com/settings/personal-access-tokens/new) or refer the guidelines from [the Github documentation.](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens#creating-a-fine-grained-personal-access-token)

#### Repository access

In the **Repository access** section, Select the required repositories you want to use with the plugin.

#### Permissions

In the repository permissions, Ensure to provide **read-only access** to the necessary section which you want to use with the plugin. **The plugin does not require any write access.**
Along with other permissions such as `Issues`, `Pull Requests`, ensure to provide read-only access to `Meta data` section as well.
This plugin does not require any org level permissions

### App ID

_conditionally required · string_

| | |
|---|---|
| Example | `App ID` |
| Shown when | **Authentication Type** is **GitHub App** (`github-app`) |

### Installation ID

_conditionally required · string_

| | |
|---|---|
| Example | `Installation ID` |
| Shown when | **Authentication Type** is **GitHub App** (`github-app`) |

### Private Key

_🔒 secret (write-only) · conditionally required · multiline text_

| | |
|---|---|
| Example | `-----BEGIN CERTIFICATE-----` |
| Shown when | **Authentication Type** is **GitHub App** (`github-app`) |

## Other settings

### cachingEnabled

_optional · boolean_

Enables the query caching wrapper in the plugin backend. Not exposed in the configuration editor; the backend currently enables caching for every datasource instance.

| | |
|---|---|
| Default | `true` |

