# Infinity configuration

Configuration reference for the **Infinity** data source (`yesoreyeram-infinity-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/yesoreyeram-infinity-datasource).

> Generated from [`dsconfig.json`](dsconfig.json). Do not edit by hand — run `go generate ./...` to refresh.

## Fields

| Field | Type | Target | Required | Description |
|---|---|---|---|---|
| `url` | string | root |  | Base URL of the query. Leave blank if you want to handle it in the query editor. |
| `basicAuth` | boolean | root |  |  |
| `basicAuthUser` | string | root | conditional | User Name |
| `jsonData.auth_method` | enum (none, basicAuth, bearerToken, apiKey, digestAuth, oauthPassThru, oauth2, aws, azureBlob) | jsonData |  | Auth type |
| `secureJsonData.basicAuthPassword` 🔒 | string | secureJsonData | conditional | password |
| `secureJsonData.bearerToken` 🔒 | string | secureJsonData | conditional | bearer token |
| `jsonData.apiKeyKey` | string | jsonData | conditional | api key key |
| `secureJsonData.apiKeyValue` 🔒 | string | secureJsonData | conditional | api key value |
| `jsonData.apiKeyType` | enum (header, query) | jsonData |  | Add api key to header/query params. |
| `jsonData.oauthPassThru` | boolean | jsonData |  |  |
| `jsonData.aws.region` | enum (af-south-1, ap-east-1, ap-northeast-1, ap-northeast-2, ap-northeast-3, ap-south-1, ap-southeast-1, ap-southeast-2, ap-southeast-3, ca-central-1, cn-north-1, cn-northwest-1, eu-central-1, eu-north-1, eu-south-1, eu-west-1, eu-west-2, eu-west-3, me-south-1, sa-east-1, us-east-1, us-east-2, us-gov-east-1, us-gov-west-1, us-iso-east-1, us-west-1, us-west-2) | jsonData | conditional | Region |
| `jsonData.aws.service` | string | jsonData | conditional | Service |
| `jsonData.aws.authType` | enum (keys) | jsonData |  |  |
| `secureJsonData.awsAccessKey` 🔒 | string | secureJsonData | conditional | aws access key |
| `secureJsonData.awsSecretKey` 🔒 | string | secureJsonData | conditional | aws secret key |
| `jsonData.oauth2.oauth2_type` | enum (client_credentials, jwt, others) | jsonData |  | This refers to OAuth2 grant type |
| `jsonData.oauth2.authStyle` | enum (0, 1, 2) | jsonData |  | Auth Style |
| `jsonData.oauth2.client_id` | string | jsonData | conditional | Client ID |
| `secureJsonData.oauth2ClientSecret` 🔒 | string | secureJsonData | conditional | Client Secret |
| `jsonData.oauth2.token_url` | string | jsonData | conditional | Token URL |
| `jsonData.oauth2.scopes` | list | jsonData |  | Scopes optionally specifies a list of requested permission scopes. Enter comma separated values |
| `jsonData.oauth2EndPointParams` | list | jsonData |  | OAuth2 endpoint params |
| `jsonData.oauth2EndPointParams[].name` | string | jsonData | yes | Param |
| `jsonData.oauth2EndPointParams[].value` | string | jsonData |  | Value |
| `jsonData.oauth2.email` | string | jsonData | conditional | Email is the OAuth client identifier used when communicating with the configured OAuth provider. |
| `jsonData.oauth2.private_key_id` | string | jsonData |  | PrivateKeyID contains an optional hint indicating which key is being used |
| `secureJsonData.oauth2JWTPrivateKey` 🔒 | string | secureJsonData | conditional | PrivateKey contains the contents of an RSA private key or the contents of a PEM file that contains a private key. The provided private key is used to sign JWT payloads |
| `jsonData.oauth2.subject` | string | jsonData |  | Subject is the optional user to impersonate. |
| `jsonData.oauth2.authHeader` | string | jsonData |  | Once the token retrieved, the same will be sent to subsequent request's header with the key "Authorization". If the API require different key, provide the key here. Defaults to Authorization |
| `jsonData.oauth2.tokenTemplate` | string | jsonData |  | Token Template allows you to customize the token value using the template. This will be Authorization header value. String ${__oauth2.access_token} will be replaced with actual access token |
| `jsonData.oauth2TokenHeaders` | list | jsonData |  | OAuth2 token request headers |
| `jsonData.oauth2TokenHeaders[].name` | string | jsonData | yes | Param |
| `jsonData.oauth2TokenHeaders[].value` | string | jsonData |  | Value |
| `jsonData.azureBlobCloudType` | enum (AzureCloud, AzureUSGovernment, AzureChinaCloud) | jsonData |  | Azure cloud type |
| `jsonData.azureBlobAccountName` | string | jsonData | conditional | Azure blob storage account name |
| `secureJsonData.azureBlobAccountKey` 🔒 | string | secureJsonData | conditional | Azure blob storage account key |
| `jsonData.azureBlobAccountUrl` | string | jsonData |  | Azure Blob account URL template. Not exposed in the configuration editor; the backend fills it from azureBlobCloudType on load. |
| `jsonData.allowedHosts` | list | jsonData |  | List of allowed host names. Enter the base URL names. ex: https://example.com |
| `jsonData.tlsSkipVerify` | boolean | jsonData |  | Skip TLS Verify |
| `jsonData.tlsAuthWithCACert` | boolean | jsonData |  | Needed for verifying self-signed TLS Certs |
| `secureJsonData.tlsCACert` 🔒 | string (multiline) | secureJsonData | conditional | CA Cert |
| `jsonData.tlsAuth` | boolean | jsonData |  | TLS Client Auth |
| `jsonData.serverName` | string | jsonData | conditional | Server Name |
| `secureJsonData.tlsClientCert` 🔒 | string (multiline) | secureJsonData | conditional | Client Cert |
| `secureJsonData.tlsClientKey` 🔒 | string (multiline) | secureJsonData | conditional | Client Key |
| `jsonData.timeoutInSeconds` | number | jsonData |  | Timeout in seconds |
| `jsonData.proxy_type` | enum (env, none, url) | jsonData |  | Proxy Mode |
| `jsonData.proxy_url` | string | jsonData | conditional | Proxy URL. Don't set the username or password here |
| `jsonData.proxy_username` | string | jsonData |  | Optional: Proxy Username. This functionality should only be used with legacy web sites. RFC 2396 warns that interpreting Userinfo this way "is NOT RECOMMENDED, because the passing of authentication information in clear text (such as URI) has proven to be a security risk in almost every case where it has been used." |
| `secureJsonData.proxyUserPassword` 🔒 | string | secureJsonData |  | Optional: Proxy Password. This functionality should only be used with legacy web sites. RFC 2396 warns that interpreting Userinfo this way "is NOT RECOMMENDED, because the passing of authentication information in clear text (such as URI) has proven to be a security risk in almost every case where it has been used." |
| `jsonData.unsecuredQueryHandling` | enum (allow, warn, deny) | jsonData |  | Option to handle insecure query content such as sensitive headers in the dashboard query |
| `jsonData.ignoreStatusCodeCheck` | boolean | jsonData |  | When enabled, the datasource will process response body even for HTTP error status codes (4xx, 5xx). This is useful for APIs that return useful data in error responses, such as detailed error messages or partial data during service degradation. |
| `jsonData.allowDangerousHTTPMethods` | boolean | jsonData |  | By default Infinity only allows GET and POST HTTP methods to reduce the risk of accidental and potentially destructive payloads. If you need PUT, PATCH or DELETE methods, make use of this setting with caution. Note: Infinity does not evaluate any permissions against the underlying API |
| `jsonData.pathEncodedUrlsEnabled` | boolean | jsonData |  | Encode query parameters with %20 |
| `jsonData.keepCookies` | list | jsonData |  | List of cookies to forward. Enter the cookie keys. ex: access_token or grafana_session_expiry |
| `jsonData.customHealthCheckEnabled` | boolean | jsonData |  | Enable custom health check |
| `jsonData.customHealthCheckUrl` | string | jsonData | conditional | Health check URL |
| `jsonData.refData` | list | jsonData |  | Named inline datasets reusable in queries via source='reference'. |
| `jsonData.refData[].name` | string | jsonData |  | Name |
| `jsonData.refData[].data` | string (multiline) | jsonData |  | Data |
| `jsonData.global_queries` | list | jsonData |  | Named datasource-level saved queries that other queries can reference via type='global'. The individual InfinityQuery shape is defined by the query editor and is intentionally opaque at the datasource-config level. |
| `jsonData.is_mock` | boolean | jsonData |  | When true, the plugin swaps in the in-memory mock client (used only by the plugin's own tests). Not exposed in the configuration editor. |
| `jsonData.httpHeaders` | list | jsonData |  | Additional HTTP headers sent with every request. Header names are stored in jsonData (httpHeaderName<N>); header values are write-only in secureJsonData (httpHeaderValue<N>). |
| `jsonData.httpHeaders[].name` | string | jsonData | yes | Header |
| `jsonData.httpHeaders[].value` | string | jsonData |  | Value |
| `jsonData.secureQuery` | list | jsonData |  | URL query parameters appended to every request. Names are stored in jsonData (secureQueryName<N>); values are write-only in secureJsonData (secureQueryValue<N>). |
| `jsonData.secureQuery[].name` | string | jsonData | yes | Key |
| `jsonData.secureQuery[].value` | string | jsonData |  | Value |

## Provisioning examples

Each scenario below shows how to provision the data source in Grafana using a YAML file (loaded by Grafana's [file provisioner](https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources)) and using the [Grafana Terraform provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source).

Placeholders like `<YOUR_TOKEN>` must be replaced with real values before use.

### No Auth (`none`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: "Infinity"
    type: yesoreyeram-infinity-datasource
    access: proxy
    basicAuth: false
    jsonData:
      allowDangerousHTTPMethods: false
      auth_method: none
      customHealthCheckEnabled: false
      customHealthCheckUrl: "https://jsonplaceholder.typicode.com/users"
      ignoreStatusCodeCheck: false
      is_mock: false
      oauth2:
        authStyle: 0
        client_id: Client ID
        email: email
        token_url: Token URL
      oauthPassThru: false
      pathEncodedUrlsEnabled: false
      proxy_type: env
      proxy_url: "Example: https://localhost:3004"
      serverName: domain.example.com
      timeoutInSeconds: 60
      tlsAuth: false
      tlsAuthWithCACert: false
      tlsSkipVerify: false
      unsecuredQueryHandling: warn
    secureJsonData:
      oauth2ClientSecret: "<YOUR_CLIENT_SECRET>"
      oauth2JWTPrivateKey: "<YOUR_PRIVATE_KEY>"
      tlsCACert: "<YOUR_CA_CERT>"
      tlsClientCert: "<YOUR_CLIENT_CERT>"
      tlsClientKey: "<YOUR_CLIENT_KEY>"
```

**Terraform**

```hcl
resource "grafana_data_source" "yesoreyeram_infinity_datasource_none" {
  type = "yesoreyeram-infinity-datasource"
  name = "Infinity"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    allowDangerousHTTPMethods = false
    auth_method = "none"
    customHealthCheckEnabled = false
    customHealthCheckUrl = "https://jsonplaceholder.typicode.com/users"
    ignoreStatusCodeCheck = false
    is_mock = false
    oauth2 = {
      authStyle = 0
      client_id = "Client ID"
      email = "email"
      token_url = "Token URL"
    }
    oauthPassThru = false
    pathEncodedUrlsEnabled = false
    proxy_type = "env"
    proxy_url = "Example: https://localhost:3004"
    serverName = "domain.example.com"
    timeoutInSeconds = 60
    tlsAuth = false
    tlsAuthWithCACert = false
    tlsSkipVerify = false
    unsecuredQueryHandling = "warn"
  })

  secure_json_data_encoded = jsonencode({
    oauth2ClientSecret = "<YOUR_CLIENT_SECRET>"
    oauth2JWTPrivateKey = "<YOUR_PRIVATE_KEY>"
    tlsCACert = "<YOUR_CA_CERT>"
    tlsClientCert = "<YOUR_CLIENT_CERT>"
    tlsClientKey = "<YOUR_CLIENT_KEY>"
  })
}
```

### Basic Authentication (`basicAuth`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: "Infinity"
    type: yesoreyeram-infinity-datasource
    access: proxy
    basicAuth: false
    basicAuthUser: username
    jsonData:
      allowDangerousHTTPMethods: false
      auth_method: basicAuth
      customHealthCheckEnabled: false
      customHealthCheckUrl: "https://jsonplaceholder.typicode.com/users"
      ignoreStatusCodeCheck: false
      is_mock: false
      oauth2:
        authStyle: 0
        client_id: Client ID
        email: email
        token_url: Token URL
      oauthPassThru: false
      pathEncodedUrlsEnabled: false
      proxy_type: env
      proxy_url: "Example: https://localhost:3004"
      serverName: domain.example.com
      timeoutInSeconds: 60
      tlsAuth: false
      tlsAuthWithCACert: false
      tlsSkipVerify: false
      unsecuredQueryHandling: warn
    secureJsonData:
      basicAuthPassword: "<YOUR_PASSWORD>"
      oauth2ClientSecret: "<YOUR_CLIENT_SECRET>"
      oauth2JWTPrivateKey: "<YOUR_PRIVATE_KEY>"
      tlsCACert: "<YOUR_CA_CERT>"
      tlsClientCert: "<YOUR_CLIENT_CERT>"
      tlsClientKey: "<YOUR_CLIENT_KEY>"
```

**Terraform**

```hcl
resource "grafana_data_source" "yesoreyeram_infinity_datasource_basicAuth" {
  type = "yesoreyeram-infinity-datasource"
  name = "Infinity"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    allowDangerousHTTPMethods = false
    auth_method = "basicAuth"
    customHealthCheckEnabled = false
    customHealthCheckUrl = "https://jsonplaceholder.typicode.com/users"
    ignoreStatusCodeCheck = false
    is_mock = false
    oauth2 = {
      authStyle = 0
      client_id = "Client ID"
      email = "email"
      token_url = "Token URL"
    }
    oauthPassThru = false
    pathEncodedUrlsEnabled = false
    proxy_type = "env"
    proxy_url = "Example: https://localhost:3004"
    serverName = "domain.example.com"
    timeoutInSeconds = 60
    tlsAuth = false
    tlsAuthWithCACert = false
    tlsSkipVerify = false
    unsecuredQueryHandling = "warn"
  })

  secure_json_data_encoded = jsonencode({
    basicAuthPassword = "<YOUR_PASSWORD>"
    oauth2ClientSecret = "<YOUR_CLIENT_SECRET>"
    oauth2JWTPrivateKey = "<YOUR_PRIVATE_KEY>"
    tlsCACert = "<YOUR_CA_CERT>"
    tlsClientCert = "<YOUR_CLIENT_CERT>"
    tlsClientKey = "<YOUR_CLIENT_KEY>"
  })
}
```

### Bearer Token (`bearerToken`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: "Infinity"
    type: yesoreyeram-infinity-datasource
    access: proxy
    basicAuth: false
    jsonData:
      allowDangerousHTTPMethods: false
      auth_method: bearerToken
      customHealthCheckEnabled: false
      customHealthCheckUrl: "https://jsonplaceholder.typicode.com/users"
      ignoreStatusCodeCheck: false
      is_mock: false
      oauth2:
        authStyle: 0
        client_id: Client ID
        email: email
        token_url: Token URL
      oauthPassThru: false
      pathEncodedUrlsEnabled: false
      proxy_type: env
      proxy_url: "Example: https://localhost:3004"
      serverName: domain.example.com
      timeoutInSeconds: 60
      tlsAuth: false
      tlsAuthWithCACert: false
      tlsSkipVerify: false
      unsecuredQueryHandling: warn
    secureJsonData:
      bearerToken: "<YOUR_BEARER_TOKEN>"
      oauth2ClientSecret: "<YOUR_CLIENT_SECRET>"
      oauth2JWTPrivateKey: "<YOUR_PRIVATE_KEY>"
      tlsCACert: "<YOUR_CA_CERT>"
      tlsClientCert: "<YOUR_CLIENT_CERT>"
      tlsClientKey: "<YOUR_CLIENT_KEY>"
```

**Terraform**

```hcl
resource "grafana_data_source" "yesoreyeram_infinity_datasource_bearerToken" {
  type = "yesoreyeram-infinity-datasource"
  name = "Infinity"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    allowDangerousHTTPMethods = false
    auth_method = "bearerToken"
    customHealthCheckEnabled = false
    customHealthCheckUrl = "https://jsonplaceholder.typicode.com/users"
    ignoreStatusCodeCheck = false
    is_mock = false
    oauth2 = {
      authStyle = 0
      client_id = "Client ID"
      email = "email"
      token_url = "Token URL"
    }
    oauthPassThru = false
    pathEncodedUrlsEnabled = false
    proxy_type = "env"
    proxy_url = "Example: https://localhost:3004"
    serverName = "domain.example.com"
    timeoutInSeconds = 60
    tlsAuth = false
    tlsAuthWithCACert = false
    tlsSkipVerify = false
    unsecuredQueryHandling = "warn"
  })

  secure_json_data_encoded = jsonencode({
    bearerToken = "<YOUR_BEARER_TOKEN>"
    oauth2ClientSecret = "<YOUR_CLIENT_SECRET>"
    oauth2JWTPrivateKey = "<YOUR_PRIVATE_KEY>"
    tlsCACert = "<YOUR_CA_CERT>"
    tlsClientCert = "<YOUR_CLIENT_CERT>"
    tlsClientKey = "<YOUR_CLIENT_KEY>"
  })
}
```

### API Key Value pair (`apiKey`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: "Infinity"
    type: yesoreyeram-infinity-datasource
    access: proxy
    basicAuth: false
    jsonData:
      allowDangerousHTTPMethods: false
      apiKeyKey: api key key
      apiKeyType: header
      auth_method: apiKey
      customHealthCheckEnabled: false
      customHealthCheckUrl: "https://jsonplaceholder.typicode.com/users"
      ignoreStatusCodeCheck: false
      is_mock: false
      oauth2:
        authStyle: 0
        client_id: Client ID
        email: email
        token_url: Token URL
      oauthPassThru: false
      pathEncodedUrlsEnabled: false
      proxy_type: env
      proxy_url: "Example: https://localhost:3004"
      serverName: domain.example.com
      timeoutInSeconds: 60
      tlsAuth: false
      tlsAuthWithCACert: false
      tlsSkipVerify: false
      unsecuredQueryHandling: warn
    secureJsonData:
      apiKeyValue: "<YOUR_VALUE>"
      oauth2ClientSecret: "<YOUR_CLIENT_SECRET>"
      oauth2JWTPrivateKey: "<YOUR_PRIVATE_KEY>"
      tlsCACert: "<YOUR_CA_CERT>"
      tlsClientCert: "<YOUR_CLIENT_CERT>"
      tlsClientKey: "<YOUR_CLIENT_KEY>"
```

**Terraform**

```hcl
resource "grafana_data_source" "yesoreyeram_infinity_datasource_apiKey" {
  type = "yesoreyeram-infinity-datasource"
  name = "Infinity"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    allowDangerousHTTPMethods = false
    apiKeyKey = "api key key"
    apiKeyType = "header"
    auth_method = "apiKey"
    customHealthCheckEnabled = false
    customHealthCheckUrl = "https://jsonplaceholder.typicode.com/users"
    ignoreStatusCodeCheck = false
    is_mock = false
    oauth2 = {
      authStyle = 0
      client_id = "Client ID"
      email = "email"
      token_url = "Token URL"
    }
    oauthPassThru = false
    pathEncodedUrlsEnabled = false
    proxy_type = "env"
    proxy_url = "Example: https://localhost:3004"
    serverName = "domain.example.com"
    timeoutInSeconds = 60
    tlsAuth = false
    tlsAuthWithCACert = false
    tlsSkipVerify = false
    unsecuredQueryHandling = "warn"
  })

  secure_json_data_encoded = jsonencode({
    apiKeyValue = "<YOUR_VALUE>"
    oauth2ClientSecret = "<YOUR_CLIENT_SECRET>"
    oauth2JWTPrivateKey = "<YOUR_PRIVATE_KEY>"
    tlsCACert = "<YOUR_CA_CERT>"
    tlsClientCert = "<YOUR_CLIENT_CERT>"
    tlsClientKey = "<YOUR_CLIENT_KEY>"
  })
}
```

### Digest Auth (`digestAuth`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: "Infinity"
    type: yesoreyeram-infinity-datasource
    access: proxy
    basicAuth: false
    basicAuthUser: username
    jsonData:
      allowDangerousHTTPMethods: false
      auth_method: digestAuth
      customHealthCheckEnabled: false
      customHealthCheckUrl: "https://jsonplaceholder.typicode.com/users"
      ignoreStatusCodeCheck: false
      is_mock: false
      oauth2:
        authStyle: 0
        client_id: Client ID
        email: email
        token_url: Token URL
      oauthPassThru: false
      pathEncodedUrlsEnabled: false
      proxy_type: env
      proxy_url: "Example: https://localhost:3004"
      serverName: domain.example.com
      timeoutInSeconds: 60
      tlsAuth: false
      tlsAuthWithCACert: false
      tlsSkipVerify: false
      unsecuredQueryHandling: warn
    secureJsonData:
      basicAuthPassword: "<YOUR_PASSWORD>"
      oauth2ClientSecret: "<YOUR_CLIENT_SECRET>"
      oauth2JWTPrivateKey: "<YOUR_PRIVATE_KEY>"
      tlsCACert: "<YOUR_CA_CERT>"
      tlsClientCert: "<YOUR_CLIENT_CERT>"
      tlsClientKey: "<YOUR_CLIENT_KEY>"
```

**Terraform**

```hcl
resource "grafana_data_source" "yesoreyeram_infinity_datasource_digestAuth" {
  type = "yesoreyeram-infinity-datasource"
  name = "Infinity"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    allowDangerousHTTPMethods = false
    auth_method = "digestAuth"
    customHealthCheckEnabled = false
    customHealthCheckUrl = "https://jsonplaceholder.typicode.com/users"
    ignoreStatusCodeCheck = false
    is_mock = false
    oauth2 = {
      authStyle = 0
      client_id = "Client ID"
      email = "email"
      token_url = "Token URL"
    }
    oauthPassThru = false
    pathEncodedUrlsEnabled = false
    proxy_type = "env"
    proxy_url = "Example: https://localhost:3004"
    serverName = "domain.example.com"
    timeoutInSeconds = 60
    tlsAuth = false
    tlsAuthWithCACert = false
    tlsSkipVerify = false
    unsecuredQueryHandling = "warn"
  })

  secure_json_data_encoded = jsonencode({
    basicAuthPassword = "<YOUR_PASSWORD>"
    oauth2ClientSecret = "<YOUR_CLIENT_SECRET>"
    oauth2JWTPrivateKey = "<YOUR_PRIVATE_KEY>"
    tlsCACert = "<YOUR_CA_CERT>"
    tlsClientCert = "<YOUR_CLIENT_CERT>"
    tlsClientKey = "<YOUR_CLIENT_KEY>"
  })
}
```

### Forward OAuth (`oauthPassThru`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: "Infinity"
    type: yesoreyeram-infinity-datasource
    access: proxy
    basicAuth: false
    jsonData:
      allowDangerousHTTPMethods: false
      auth_method: oauthPassThru
      customHealthCheckEnabled: false
      customHealthCheckUrl: "https://jsonplaceholder.typicode.com/users"
      ignoreStatusCodeCheck: false
      is_mock: false
      oauth2:
        authStyle: 0
        client_id: Client ID
        email: email
        token_url: Token URL
      oauthPassThru: false
      pathEncodedUrlsEnabled: false
      proxy_type: env
      proxy_url: "Example: https://localhost:3004"
      serverName: domain.example.com
      timeoutInSeconds: 60
      tlsAuth: false
      tlsAuthWithCACert: false
      tlsSkipVerify: false
      unsecuredQueryHandling: warn
    secureJsonData:
      oauth2ClientSecret: "<YOUR_CLIENT_SECRET>"
      oauth2JWTPrivateKey: "<YOUR_PRIVATE_KEY>"
      tlsCACert: "<YOUR_CA_CERT>"
      tlsClientCert: "<YOUR_CLIENT_CERT>"
      tlsClientKey: "<YOUR_CLIENT_KEY>"
```

**Terraform**

```hcl
resource "grafana_data_source" "yesoreyeram_infinity_datasource_oauthPassThru" {
  type = "yesoreyeram-infinity-datasource"
  name = "Infinity"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    allowDangerousHTTPMethods = false
    auth_method = "oauthPassThru"
    customHealthCheckEnabled = false
    customHealthCheckUrl = "https://jsonplaceholder.typicode.com/users"
    ignoreStatusCodeCheck = false
    is_mock = false
    oauth2 = {
      authStyle = 0
      client_id = "Client ID"
      email = "email"
      token_url = "Token URL"
    }
    oauthPassThru = false
    pathEncodedUrlsEnabled = false
    proxy_type = "env"
    proxy_url = "Example: https://localhost:3004"
    serverName = "domain.example.com"
    timeoutInSeconds = 60
    tlsAuth = false
    tlsAuthWithCACert = false
    tlsSkipVerify = false
    unsecuredQueryHandling = "warn"
  })

  secure_json_data_encoded = jsonencode({
    oauth2ClientSecret = "<YOUR_CLIENT_SECRET>"
    oauth2JWTPrivateKey = "<YOUR_PRIVATE_KEY>"
    tlsCACert = "<YOUR_CA_CERT>"
    tlsClientCert = "<YOUR_CLIENT_CERT>"
    tlsClientKey = "<YOUR_CLIENT_KEY>"
  })
}
```

### OAuth2 (`oauth2`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: "Infinity"
    type: yesoreyeram-infinity-datasource
    access: proxy
    basicAuth: false
    jsonData:
      allowDangerousHTTPMethods: false
      auth_method: oauth2
      customHealthCheckEnabled: false
      customHealthCheckUrl: "https://jsonplaceholder.typicode.com/users"
      ignoreStatusCodeCheck: false
      is_mock: false
      oauth2:
        authStyle: 0
        client_id: Client ID
        email: email
        oauth2_type: client_credentials
        token_url: Token URL
      oauthPassThru: false
      pathEncodedUrlsEnabled: false
      proxy_type: env
      proxy_url: "Example: https://localhost:3004"
      serverName: domain.example.com
      timeoutInSeconds: 60
      tlsAuth: false
      tlsAuthWithCACert: false
      tlsSkipVerify: false
      unsecuredQueryHandling: warn
    secureJsonData:
      oauth2ClientSecret: "<YOUR_CLIENT_SECRET>"
      oauth2JWTPrivateKey: "<YOUR_PRIVATE_KEY>"
      tlsCACert: "<YOUR_CA_CERT>"
      tlsClientCert: "<YOUR_CLIENT_CERT>"
      tlsClientKey: "<YOUR_CLIENT_KEY>"
```

**Terraform**

```hcl
resource "grafana_data_source" "yesoreyeram_infinity_datasource_oauth2" {
  type = "yesoreyeram-infinity-datasource"
  name = "Infinity"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    allowDangerousHTTPMethods = false
    auth_method = "oauth2"
    customHealthCheckEnabled = false
    customHealthCheckUrl = "https://jsonplaceholder.typicode.com/users"
    ignoreStatusCodeCheck = false
    is_mock = false
    oauth2 = {
      authStyle = 0
      client_id = "Client ID"
      email = "email"
      oauth2_type = "client_credentials"
      token_url = "Token URL"
    }
    oauthPassThru = false
    pathEncodedUrlsEnabled = false
    proxy_type = "env"
    proxy_url = "Example: https://localhost:3004"
    serverName = "domain.example.com"
    timeoutInSeconds = 60
    tlsAuth = false
    tlsAuthWithCACert = false
    tlsSkipVerify = false
    unsecuredQueryHandling = "warn"
  })

  secure_json_data_encoded = jsonencode({
    oauth2ClientSecret = "<YOUR_CLIENT_SECRET>"
    oauth2JWTPrivateKey = "<YOUR_PRIVATE_KEY>"
    tlsCACert = "<YOUR_CA_CERT>"
    tlsClientCert = "<YOUR_CLIENT_CERT>"
    tlsClientKey = "<YOUR_CLIENT_KEY>"
  })
}
```

### AWS (`aws`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: "Infinity"
    type: yesoreyeram-infinity-datasource
    access: proxy
    basicAuth: false
    jsonData:
      allowDangerousHTTPMethods: false
      auth_method: aws
      aws:
        authType: keys
        region: af-south-1
        service: monitoring
      customHealthCheckEnabled: false
      customHealthCheckUrl: "https://jsonplaceholder.typicode.com/users"
      ignoreStatusCodeCheck: false
      is_mock: false
      oauth2:
        authStyle: 0
        client_id: Client ID
        email: email
        token_url: Token URL
      oauthPassThru: false
      pathEncodedUrlsEnabled: false
      proxy_type: env
      proxy_url: "Example: https://localhost:3004"
      serverName: domain.example.com
      timeoutInSeconds: 60
      tlsAuth: false
      tlsAuthWithCACert: false
      tlsSkipVerify: false
      unsecuredQueryHandling: warn
    secureJsonData:
      awsAccessKey: "<YOUR_ACCESS_KEY>"
      awsSecretKey: "<YOUR_SECRET_KEY>"
      oauth2ClientSecret: "<YOUR_CLIENT_SECRET>"
      oauth2JWTPrivateKey: "<YOUR_PRIVATE_KEY>"
      tlsCACert: "<YOUR_CA_CERT>"
      tlsClientCert: "<YOUR_CLIENT_CERT>"
      tlsClientKey: "<YOUR_CLIENT_KEY>"
```

**Terraform**

```hcl
resource "grafana_data_source" "yesoreyeram_infinity_datasource_aws" {
  type = "yesoreyeram-infinity-datasource"
  name = "Infinity"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    allowDangerousHTTPMethods = false
    auth_method = "aws"
    aws = {
      authType = "keys"
      region = "af-south-1"
      service = "monitoring"
    }
    customHealthCheckEnabled = false
    customHealthCheckUrl = "https://jsonplaceholder.typicode.com/users"
    ignoreStatusCodeCheck = false
    is_mock = false
    oauth2 = {
      authStyle = 0
      client_id = "Client ID"
      email = "email"
      token_url = "Token URL"
    }
    oauthPassThru = false
    pathEncodedUrlsEnabled = false
    proxy_type = "env"
    proxy_url = "Example: https://localhost:3004"
    serverName = "domain.example.com"
    timeoutInSeconds = 60
    tlsAuth = false
    tlsAuthWithCACert = false
    tlsSkipVerify = false
    unsecuredQueryHandling = "warn"
  })

  secure_json_data_encoded = jsonencode({
    awsAccessKey = "<YOUR_ACCESS_KEY>"
    awsSecretKey = "<YOUR_SECRET_KEY>"
    oauth2ClientSecret = "<YOUR_CLIENT_SECRET>"
    oauth2JWTPrivateKey = "<YOUR_PRIVATE_KEY>"
    tlsCACert = "<YOUR_CA_CERT>"
    tlsClientCert = "<YOUR_CLIENT_CERT>"
    tlsClientKey = "<YOUR_CLIENT_KEY>"
  })
}
```

### Azure Blob (`azureBlob`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: "Infinity"
    type: yesoreyeram-infinity-datasource
    access: proxy
    basicAuth: false
    jsonData:
      allowDangerousHTTPMethods: false
      auth_method: azureBlob
      azureBlobAccountName: Azure blob storage account name
      azureBlobCloudType: AzureCloud
      customHealthCheckEnabled: false
      customHealthCheckUrl: "https://jsonplaceholder.typicode.com/users"
      ignoreStatusCodeCheck: false
      is_mock: false
      oauth2:
        authStyle: 0
        client_id: Client ID
        email: email
        token_url: Token URL
      oauthPassThru: false
      pathEncodedUrlsEnabled: false
      proxy_type: env
      proxy_url: "Example: https://localhost:3004"
      serverName: domain.example.com
      timeoutInSeconds: 60
      tlsAuth: false
      tlsAuthWithCACert: false
      tlsSkipVerify: false
      unsecuredQueryHandling: warn
    secureJsonData:
      azureBlobAccountKey: "<YOUR_STORAGE_ACCOUNT_KEY>"
      oauth2ClientSecret: "<YOUR_CLIENT_SECRET>"
      oauth2JWTPrivateKey: "<YOUR_PRIVATE_KEY>"
      tlsCACert: "<YOUR_CA_CERT>"
      tlsClientCert: "<YOUR_CLIENT_CERT>"
      tlsClientKey: "<YOUR_CLIENT_KEY>"
```

**Terraform**

```hcl
resource "grafana_data_source" "yesoreyeram_infinity_datasource_azureBlob" {
  type = "yesoreyeram-infinity-datasource"
  name = "Infinity"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    allowDangerousHTTPMethods = false
    auth_method = "azureBlob"
    azureBlobAccountName = "Azure blob storage account name"
    azureBlobCloudType = "AzureCloud"
    customHealthCheckEnabled = false
    customHealthCheckUrl = "https://jsonplaceholder.typicode.com/users"
    ignoreStatusCodeCheck = false
    is_mock = false
    oauth2 = {
      authStyle = 0
      client_id = "Client ID"
      email = "email"
      token_url = "Token URL"
    }
    oauthPassThru = false
    pathEncodedUrlsEnabled = false
    proxy_type = "env"
    proxy_url = "Example: https://localhost:3004"
    serverName = "domain.example.com"
    timeoutInSeconds = 60
    tlsAuth = false
    tlsAuthWithCACert = false
    tlsSkipVerify = false
    unsecuredQueryHandling = "warn"
  })

  secure_json_data_encoded = jsonencode({
    azureBlobAccountKey = "<YOUR_STORAGE_ACCOUNT_KEY>"
    oauth2ClientSecret = "<YOUR_CLIENT_SECRET>"
    oauth2JWTPrivateKey = "<YOUR_PRIVATE_KEY>"
    tlsCACert = "<YOUR_CA_CERT>"
    tlsClientCert = "<YOUR_CLIENT_CERT>"
    tlsClientKey = "<YOUR_CLIENT_KEY>"
  })
}
```

