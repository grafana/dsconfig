# AppDynamics configuration

How to configure the **AppDynamics** data source (`dlopes7-appdynamics-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/dlopes7-appdynamics-datasource).

> This page is generated from [`dsconfig.json`](dsconfig.json). Do not edit it by hand — run `go generate ./...` to refresh.

## Configuration sections

- [Connection](#connection)
- [Authentication](#authentication)
- [Analytics](#analytics) — _optional_
- [TLS settings](#tls-settings) — _optional_

## Connection

### URL

_**required** · string_

| | |
|---|---|
| Example | `http://localhost:8086` |

## Authentication

### Basic auth

_optional · toggle_

| | |
|---|---|
| Default | `false` |

### User

_conditionally required · string_

| | |
|---|---|
| Example | `user` |
| Shown when | **Basic auth** is `true` |

### Password

_🔒 secret (write-only) · conditionally required · string_

| | |
|---|---|
| Example | `Password` |
| Shown when | **Basic auth** is `true` |

### Client Name

_conditionally required · string_

| | |
|---|---|
| Example | `Client Name` |
| Required when | `jsonData_clientDomain != '' || secureJsonData_clientSecret != ''` |

### Client Domain

_conditionally required · string_

| | |
|---|---|
| Example | `Client Domain` |
| Required when | `jsonData_clientName != '' || secureJsonData_clientSecret != ''` |

### Client Secret

_🔒 secret (write-only) · conditionally required · string_

Authenticate to AppDynamics using an API key. This will override username/password (basic) authenticationLeave blank for username/password authentication.

| | |
|---|---|
| Example | `Paste the client secret here...` |
| Required when | `jsonData_clientName != '' || jsonData_clientDomain != ''` |

### Custom HTTP Headers

_optional · list_

Additional HTTP headers sent with every request. Header names are stored in jsonData (httpHeaderName<N>); header values are write-only in secureJsonData (httpHeaderValue<N>).

Each item has the following fields:

#### Header

_**required** · string_

| | |
|---|---|
| Example | `X-Custom-Header` |
| Must match | `^[A-Za-z][A-Za-z0-9-]*$` |

#### Value

_optional · string_

| | |
|---|---|
| Example | `Header Value` |

## Analytics

_This section is optional._

### Analytics API URL

_optional · select_

The Analytics API URL.

### Global Account Name

_optional · string_

The global account name, as shown in the Controller UI License page.

### Analytics API Key

_🔒 secret (write-only) · optional · string_

The Analytics API Key.

| | |
|---|---|
| Example | `Paste in Analytics API Key here...` |

**API Key**

You can generate and use an API key for each API access call into your Controller by generating an access token through the Administration UI. These API keys usually have a longer expiration. The account administrator can generate and distribute to parties/teams who need Controller access, but do not want to refresh such tokens frequently.

Create an API key in the Controller Administration UI under `Account > API Clients` (`<controller-url>/controller/#/location=ACCOUNT_ADMIN_API_CLIENTS`).

## TLS settings

_This section is optional._

### Skip TLS Verify

_optional · toggle_

| | |
|---|---|
| Default | `false` |

