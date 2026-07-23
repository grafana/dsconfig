# GitLab configuration

How to configure the **GitLab** data source (`grafana-gitlab-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-gitlab-datasource).

> This page is generated from [`dsconfig.json`](dsconfig.json). Do not edit it by hand — run `go generate ./...` to refresh.

## Configuration sections

- [Connection](#connection)
- [Authentication](#authentication)
- [Additional Settings](#additional-settings) — _optional_

## Connection

### URL

_optional · string_

The URL for your GitLab instance (ex: gitlab.domain.com). Leave blank if you use gitlab.com.

| | |
|---|---|
| Default | `https://gitlab.com/api/v4` |
| Example | `Default: https://gitlab.com/api/v4` |

## Authentication

### Access token

_🔒 secret (write-only) · **required** · string_

Provide information to grant access to this data source. To learn more about access tokens, [click here.](https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html).

| | |
|---|---|
| Example | `Access token` |

## Additional Settings

Additional settings are optional settings that can be configured for more control over your data source.

_This section is optional._

### Page limit

_optional · number_

The page limit is the maximum number of pages returned when creating a query. The default is 5.

| | |
|---|---|
| Default | `5` |
| Example | `Page limit` |

