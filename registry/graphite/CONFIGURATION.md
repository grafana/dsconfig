# Graphite configuration

Configuration reference for the **Graphite** data source (`graphite`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/grafana/latest/datasources/graphite/).

> Generated from [`dsconfig.json`](dsconfig.json). Do not edit by hand — run `go generate ./...` to refresh.

## Fields

| Field | Type | Target | Required | Description |
|---|---|---|---|---|
| `url` | string | root | yes | Specify a complete HTTP URL (for example http://your_server:8080) |
| `basicAuth` | boolean | root |  | Basic auth |
| `withCredentials` | boolean | root |  | Whether credentials such as cookies or auth headers should be sent with cross-site requests. |
| `basicAuthUser` | string | root | conditional | User |
| `secureJsonData.basicAuthPassword` 🔒 | string | secureJsonData | conditional | Password |
| `jsonData.tlsAuth` | boolean | jsonData |  | TLS Client Auth |
| `jsonData.tlsAuthWithCACert` | boolean | jsonData |  | Needed for verifying self-signed TLS Certs |
| `jsonData.tlsSkipVerify` | boolean | jsonData |  | Skip TLS Verify |
| `jsonData.oauthPassThru` | boolean | jsonData |  | Forward the user's upstream OAuth identity to the data source (Their access token gets passed along). |
| `jsonData.serverName` | string | jsonData | conditional | ServerName |
| `secureJsonData.tlsCACert` 🔒 | string (multiline) | secureJsonData | conditional | CA Cert |
| `secureJsonData.tlsClientCert` 🔒 | string (multiline) | secureJsonData | conditional | Client Cert |
| `secureJsonData.tlsClientKey` 🔒 | string (multiline) | secureJsonData | conditional | Client Key |
| `jsonData.keepCookies` | list | jsonData |  | Grafana proxy deletes forwarded cookies by default. Specify cookies by name that should be forwarded to the data source. |
| `jsonData.timeout` | number | jsonData |  | HTTP request timeout in seconds |
| `jsonData.graphiteVersion` | enum (0.9, 1.0, 1.1) | jsonData |  | This option controls what functions are available in the Graphite query editor. |
| `jsonData.graphiteType` | enum (default, metrictank) | jsonData |  | There are different types of Graphite compatible backends. Here you can specify the type you are using. For Metrictank, this will enable specific features, like query processing meta data. Metrictank         is a multi-tenant timeseries engine for Graphite and friends. |
| `jsonData.rollupIndicatorEnabled` | boolean | jsonData |  | Shows up as an info icon in panel headers when data is aggregated. |
| `jsonData.importConfiguration` | object | jsonData |  | Label mappings |
| `jsonData.httpHeaders` | list | jsonData |  | Additional HTTP headers sent with every request. Header names are stored in jsonData (httpHeaderName<N>); header values are write-only in secureJsonData (httpHeaderValue<N>). Rendered by the @grafana/ui CustomHeadersSettings component (DataSourceHttpSettings.tsx:352) and forwarded to Graphite by the SDK's HTTPClientOptions. |
| `jsonData.httpHeaders[].name` | string | jsonData | yes | Header |
| `jsonData.httpHeaders[].value` | string | jsonData |  | Value |

## Provisioning examples

Each scenario below shows how to provision the data source in Grafana using a YAML file (loaded by Grafana's [file provisioner](https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources)) and using the [Grafana Terraform provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source).

Placeholders like `<YOUR_TOKEN>` must be replaced with real values before use.

### Default configuration

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Graphite
    type: graphite
    access: proxy
    basicAuth: false
    basicAuthUser: user
    url: "http://localhost:8080"
    withCredentials: false
    jsonData:
      graphiteVersion: "1.1"
      oauthPassThru: false
      rollupIndicatorEnabled: false
      serverName: domain.example.com
      tlsAuth: false
      tlsAuthWithCACert: false
      tlsSkipVerify: false
    secureJsonData:
      basicAuthPassword: "<YOUR_PASSWORD>"
      tlsCACert: "<YOUR_CA_CERT>"
      tlsClientCert: "<YOUR_CLIENT_CERT>"
      tlsClientKey: "<YOUR_CLIENT_KEY>"
```

**Terraform**

```hcl
resource "grafana_data_source" "graphite" {
  type = "graphite"
  name = "Graphite"
  url = "http://localhost:8080"

  json_data_encoded = jsonencode({
    graphiteVersion = "1.1"
    oauthPassThru = false
    rollupIndicatorEnabled = false
    serverName = "domain.example.com"
    tlsAuth = false
    tlsAuthWithCACert = false
    tlsSkipVerify = false
  })

  secure_json_data_encoded = jsonencode({
    basicAuthPassword = "<YOUR_PASSWORD>"
    tlsCACert = "<YOUR_CA_CERT>"
    tlsClientCert = "<YOUR_CLIENT_CERT>"
    tlsClientKey = "<YOUR_CLIENT_KEY>"
  })
}
```

