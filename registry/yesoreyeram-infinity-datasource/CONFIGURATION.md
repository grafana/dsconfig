# Infinity configuration

How to configure the **Infinity** data source (`yesoreyeram-infinity-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/yesoreyeram-infinity-datasource).

> This page is generated from [`dsconfig.json`](dsconfig.json). Do not edit it by hand — run `go generate ./...` to refresh.

## Configuration sections

- [Main](#main)
- [Authentication](#authentication)
- [URL, Headers & Params](#url-headers--params) — _optional_
- [Network](#network) — _optional_
- [Security](#security) — _optional_
- [Health check](#health-check) — _optional_
- [Advanced configuration](#advanced-configuration) — _optional_

## Main

### Base URL

_optional · string_

Base URL of the query. Leave blank if you want to handle it in the query editor.

| | |
|---|---|
| Example | `Leave blank and you can specify full URL in the query.` |

### Allowed hosts

_optional · list_

List of allowed host names. Enter the base URL names. ex: https://example.com.

| | |
|---|---|
| Example | `https://example.com` |
| Shown when | `jsonData_authMethod != 'none' && jsonData_authMethod != 'azureBlob'` |

## Authentication

### Auth type

_optional · select_

| | |
|---|---|
| Default | `none` |
| Allowed values | `none` (No Auth), `basicAuth` (Basic Authentication), `bearerToken` (Bearer Token), `apiKey` (API Key Value pair), `digestAuth` (Digest Auth), `oauthPassThru` (Forward OAuth), `oauth2` (OAuth2), `aws` (AWS), `azureBlob` (Azure Blob) |

### User Name

_conditionally required · string_

| | |
|---|---|
| Example | `username` |
| Shown when | `jsonData_authMethod == 'basicAuth' || jsonData_authMethod == 'digestAuth'` |

### Password

_🔒 secret (write-only) · conditionally required · string_

password.

| | |
|---|---|
| Example | `password` |
| Shown when | `jsonData_authMethod == 'basicAuth' || jsonData_authMethod == 'digestAuth'` |

### Bearer token

_🔒 secret (write-only) · conditionally required · string_

bearer token.

| | |
|---|---|
| Example | `bearer token` |
| Shown when | **Auth type** is **Bearer Token** (`bearerToken`) |

### Key

_conditionally required · string_

api key key.

| | |
|---|---|
| Example | `api key key` |
| Shown when | **Auth type** is **API Key Value pair** (`apiKey`) |

### Value

_🔒 secret (write-only) · conditionally required · string_

api key value.

| | |
|---|---|
| Example | `api key value` |
| Shown when | **Auth type** is **API Key Value pair** (`apiKey`) |

### Add to

_optional · radio_

Add api key to header/query params.

| | |
|---|---|
| Default | `header` |
| Allowed values | `header` (Header), `query` (Query Param) |
| Shown when | **Auth type** is **API Key Value pair** (`apiKey`) |

### Region

_conditionally required · select_

| | |
|---|---|
| Example | `us-east-2` |
| Allowed values | `af-south-1`, `ap-east-1`, `ap-northeast-1`, `ap-northeast-2`, `ap-northeast-3`, `ap-south-1`, `ap-southeast-1`, `ap-southeast-2`, `ap-southeast-3`, `ca-central-1`, `cn-north-1`, `cn-northwest-1`, `eu-central-1`, `eu-north-1`, `eu-south-1`, `eu-west-1`, `eu-west-2`, `eu-west-3`, `me-south-1`, `sa-east-1`, `us-east-1`, `us-east-2`, `us-gov-east-1`, `us-gov-west-1`, `us-iso-east-1`, `us-west-1`, `us-west-2` |
| Shown when | **Auth type** is **AWS** (`aws`) |

### Service

_conditionally required · string_

| | |
|---|---|
| Example | `monitoring` |
| Shown when | **Auth type** is **AWS** (`aws`) |

### Access Key

_🔒 secret (write-only) · conditionally required · string_

aws access key.

| | |
|---|---|
| Example | `aws access key` |
| Shown when | **Auth type** is **AWS** (`aws`) |

### Secret Key

_🔒 secret (write-only) · conditionally required · string_

aws secret key.

| | |
|---|---|
| Example | `aws secret key` |
| Shown when | **Auth type** is **AWS** (`aws`) |

### Grant Type

_optional · radio_

This refers to OAuth2 grant type.

| | |
|---|---|
| Default | `client_credentials` |
| Allowed values | `client_credentials` (Client Credentials), `jwt` (JWT), `others` (Others) |
| Shown when | **Auth type** is **OAuth2** (`oauth2`) |

### Auth Style

_optional · radio_

| | |
|---|---|
| Default | `0` |
| Allowed values | `0` (Auto), `1` (In Params), `2` (In Header) |
| Shown when | `jsonData_authMethod == 'oauth2' && jsonData_oauth2Type == 'client_credentials'` |

### Client ID

_conditionally required · string_

| | |
|---|---|
| Example | `Client ID` |
| Shown when | `jsonData_authMethod == 'oauth2' && jsonData_oauth2Type == 'client_credentials'` |

### Client Secret

_🔒 secret (write-only) · conditionally required · string_

| | |
|---|---|
| Example | `Client secret` |
| Shown when | `jsonData_authMethod == 'oauth2' && jsonData_oauth2Type == 'client_credentials'` |

### Token URL

_conditionally required · string_

| | |
|---|---|
| Example | `Token URL` |
| Shown when | `jsonData_authMethod == 'oauth2' && (jsonData_oauth2Type == 'client_credentials' || jsonData_oauth2Type == 'jwt')` |

### Scopes

_optional · list_

Scopes optionally specifies a list of requested permission scopes. Enter comma separated values.

| | |
|---|---|
| Example | `Comma separated values of scopes` |
| Shown when | `jsonData_authMethod == 'oauth2' && (jsonData_oauth2Type == 'client_credentials' || jsonData_oauth2Type == 'jwt')` |

### Custom Token Header

_optional · string_

Once the token retrieved, the same will be sent to subsequent request's header with the key "Authorization". If the API require different key, provide the key here. Defaults to Authorization.

| | |
|---|---|
| Example | `Authorization` |
| Shown when | `jsonData_authMethod == 'oauth2' && (jsonData_oauth2Type == 'client_credentials' || jsonData_oauth2Type == 'jwt')` |

### Custom Token Template

_optional · string_

Token Template allows you to customize the token value using the template. This will be Authorization header value. String ${__oauth2.access_token} will be replaced with actual access token.

| | |
|---|---|
| Example | `Bearer ${__oauth2.access_token}` |
| Shown when | `jsonData_authMethod == 'oauth2' && (jsonData_oauth2Type == 'client_credentials' || jsonData_oauth2Type == 'jwt')` |

### Email

_conditionally required · string_

Email is the OAuth client identifier used when communicating with the configured OAuth provider.

| | |
|---|---|
| Example | `email` |
| Shown when | `jsonData_authMethod == 'oauth2' && jsonData_oauth2Type == 'jwt'` |

### Private Key Identifier

_optional · string_

PrivateKeyID contains an optional hint indicating which key is being used.

| | |
|---|---|
| Example | `(optional) private key identifier` |
| Shown when | `jsonData_authMethod == 'oauth2' && jsonData_oauth2Type == 'jwt'` |

### Private Key

_🔒 secret (write-only) · conditionally required · string_

PrivateKey contains the contents of an RSA private key or the contents of a PEM file that contains a private key. The provided private key is used to sign JWT payloads.

| | |
|---|---|
| Example | `Private Key` |
| Shown when | `jsonData_authMethod == 'oauth2' && jsonData_oauth2Type == 'jwt'` |

### OAuth2 endpoint params

_optional · list_

OAuth2 endpoint params.

| | |
|---|---|
| Shown when | `jsonData_authMethod == 'oauth2' && (jsonData_oauth2Type == 'client_credentials')` |

Each item has the following fields:

#### Param

_**required** · string_

| | |
|---|---|
| Example | `key` |

#### Value

_optional · string_

| | |
|---|---|
| Example | `value` |

### Subject

_optional · string_

Subject is the optional user to impersonate.

| | |
|---|---|
| Example | `(optional) Subject` |
| Shown when | `jsonData_authMethod == 'oauth2' && jsonData_oauth2Type == 'jwt'` |

### Token request headers

_optional · list_

OAuth2 token request headers.

| | |
|---|---|
| Shown when | **Auth type** is **OAuth2** (`oauth2`) |

Each item has the following fields:

#### Param

_**required** · string_

| | |
|---|---|
| Example | `key` |

#### Value

_optional · string_

| | |
|---|---|
| Example | `value` |

### Azure cloud

_optional · select_

Azure cloud type.

| | |
|---|---|
| Default | `AzureCloud` |
| Allowed values | `AzureCloud` (Azure), `AzureUSGovernment` (Azure US Government), `AzureChinaCloud` (Azure China) |
| Shown when | **Auth type** is **Azure Blob** (`azureBlob`) |

### Storage account name

_conditionally required · string_

Azure blob storage account name.

| | |
|---|---|
| Example | `Azure blob storage account name` |
| Shown when | **Auth type** is **Azure Blob** (`azureBlob`) |

### Storage account key

_🔒 secret (write-only) · conditionally required · string_

Azure blob storage account key.

| | |
|---|---|
| Example | `Azure blob storage account key` |
| Shown when | **Auth type** is **Azure Blob** (`azureBlob`) |

## URL, Headers & Params

_This section is optional._

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

### URL Query Params

_optional · list_

URL query parameters appended to every request. Names are stored in jsonData (secureQueryName<N>); values are write-only in secureJsonData (secureQueryValue<N>).

Each item has the following fields:

#### Key

_**required** · string_

| | |
|---|---|
| Example | `key` |

#### Value

_optional · string_

| | |
|---|---|
| Example | `value` |

### Ignore status code check

_optional · toggle_

When enabled, the datasource will process response body even for HTTP error status codes (4xx, 5xx). This is useful for APIs that return useful data in error responses, such as detailed error messages or partial data during service degradation.

| | |
|---|---|
| Default | `false` |

### Allow dangerous HTTP methods

_optional · toggle_

By default Infinity only allows GET and POST HTTP methods to reduce the risk of accidental and potentially destructive payloads. If you need PUT, PATCH or DELETE methods, make use of this setting with caution. Note: Infinity does not evaluate any permissions against the underlying API.

| | |
|---|---|
| Default | `false` |

### Encode query parameters with %20

_optional · toggle_

| | |
|---|---|
| Default | `false` |

### Include cookies

_optional · list_

List of cookies to forward. Enter the cookie keys. ex: access_token or grafana_session_expiry.

| | |
|---|---|
| Example | `Enter the cookie names (enter key to add)` |

## Network

_This section is optional._

### Timeout in seconds

_optional · number_

| | |
|---|---|
| Default | `60` |
| Example | `timeout in seconds` |
| Range | 0 – 300 |

### Skip TLS Verify

_optional · toggle_

Skip TLS Verify.

| | |
|---|---|
| Default | `false` |

### With CA Cert

_optional · toggle_

Needed for verifying self-signed TLS Certs.

| | |
|---|---|
| Default | `false` |

### CA Cert

_🔒 secret (write-only) · conditionally required · multiline text_

| | |
|---|---|
| Example | `Begins with -----BEGIN CERTIFICATE-----` |
| Shown when | **With CA Cert** is `true` |

### TLS Client Auth

_optional · toggle_

TLS Client Auth.

| | |
|---|---|
| Default | `false` |

### Server Name

_conditionally required · string_

Server Name.

| | |
|---|---|
| Example | `domain.example.com` |
| Shown when | **TLS Client Auth** is `true` |

### Client Cert

_🔒 secret (write-only) · conditionally required · multiline text_

| | |
|---|---|
| Example | `Begins with -----BEGIN CERTIFICATE-----` |
| Shown when | **TLS Client Auth** is `true` |

### Client Key

_🔒 secret (write-only) · conditionally required · multiline text_

| | |
|---|---|
| Example | `Begins with -----BEGIN RSA PRIVATE KEY-----` |
| Shown when | **TLS Client Auth** is `true` |

### Proxy Mode

_optional · radio_

| | |
|---|---|
| Default | `env` |
| Allowed values | `env` (From environment variable / Default), `none` (None), `url` (URL) |

### Proxy URL

_conditionally required · string_

Proxy URL. Don't set the username or password here.

| | |
|---|---|
| Example | `Example: https://localhost:3004` |
| Shown when | **Proxy Mode** is **URL** (`url`) |

### Proxy User Name

_optional · string_

Optional: Proxy Username. This functionality should only be used with legacy web sites. RFC 2396 warns that interpreting Userinfo this way "is NOT RECOMMENDED, because the passing of authentication information in clear text (such as URI) has proven to be a security risk in almost every case where it has been used.".

| | |
|---|---|
| Example | `Example: foo` |
| Shown when | **Proxy Mode** is **URL** (`url`) |

### Proxy Password

_🔒 secret (write-only) · optional · string_

Optional: Proxy Password. This functionality should only be used with legacy web sites. RFC 2396 warns that interpreting Userinfo this way "is NOT RECOMMENDED, because the passing of authentication information in clear text (such as URI) has proven to be a security risk in almost every case where it has been used.".

| | |
|---|---|
| Example | `Proxy Password` |
| Shown when | **Proxy Mode** is **URL** (`url`) |

## Security

_This section is optional._

### Query security

_optional · radio_

Option to handle insecure query content such as sensitive headers in the dashboard query.

| | |
|---|---|
| Default | `warn` |
| Allowed values | `allow` (Allow), `warn` (Warn), `deny` (Deny) |

## Health check

_This section is optional._

### Enable custom health check

_optional · toggle_

| | |
|---|---|
| Default | `false` |

### Health check URL

_conditionally required · string_

| | |
|---|---|
| Example | `https://jsonplaceholder.typicode.com/users` |
| Shown when | **Enable custom health check** is `true` |

## Advanced configuration

_This section is optional._

### Reference data

_optional · list_

Named inline datasets reusable in queries via source='reference'.

Each item has the following fields:

#### Name

_optional · string_

| | |
|---|---|
| Example | `Give an unique name to your reference data` |

#### Data

_optional · multiline text_

| | |
|---|---|
| Example | `Enter data here. either json / csv / tsv / xml / html` |

### global_queries

_optional · list_

Named datasource-level saved queries that other queries can reference via type='global'. The individual InfinityQuery shape is defined by the query editor and is intentionally opaque at the datasource-config level.

## Other settings

### azureBlobAccountUrl

_optional · string_

Azure Blob account URL template. Not exposed in the configuration editor; the backend fills it from azureBlobCloudType on load.

### is_mock

_optional · boolean_

When true, the plugin swaps in the in-memory mock client (used only by the plugin's own tests). Not exposed in the configuration editor.

| | |
|---|---|
| Default | `false` |

