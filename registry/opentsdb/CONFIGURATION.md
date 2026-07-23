# OpenTSDB configuration

Configuration reference for the **OpenTSDB** data source (`opentsdb`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/grafana/latest/datasources/opentsdb/).

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
| `jsonData.httpHeaders` | list | jsonData |  | Additional HTTP headers sent with every request. Header names are stored in jsonData (httpHeaderName<N>); header values are write-only in secureJsonData (httpHeaderValue<N>). |
| `jsonData.httpHeaders[].name` | string | jsonData | yes | Header |
| `jsonData.httpHeaders[].value` | string | jsonData |  | Value |
| `jsonData.tsdbVersion` | enum (1, 2, 3, 4) | jsonData |  | Version |
| `jsonData.tsdbResolution` | enum (1, 2) | jsonData |  | Resolution |
| `jsonData.lookupLimit` | number | jsonData |  | Lookup limit |

## Provisioning examples

Each scenario below shows how to provision the data source in Grafana using a YAML file (loaded by Grafana's [file provisioner](https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources)) and using the [Grafana Terraform provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source).

Placeholders like `<YOUR_TOKEN>` must be replaced with real values before use.

### Default configuration

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: OpenTSDB
    type: opentsdb
    access: proxy
    basicAuth: false
    basicAuthUser: user
    url: "http://localhost:4242"
    withCredentials: false
    jsonData:
      lookupLimit: 1000
      oauthPassThru: false
      serverName: domain.example.com
      tlsAuth: false
      tlsAuthWithCACert: false
      tlsSkipVerify: false
      tsdbResolution: 1
      tsdbVersion: 1
    secureJsonData:
      basicAuthPassword: "<YOUR_PASSWORD>"
      tlsCACert: "<YOUR_CA_CERT>"
      tlsClientCert: "<YOUR_CLIENT_CERT>"
      tlsClientKey: "<YOUR_CLIENT_KEY>"
```

**Terraform**

```hcl
resource "grafana_data_source" "opentsdb" {
  type = "opentsdb"
  name = "OpenTSDB"
  url = "http://localhost:4242"

  json_data_encoded = jsonencode({
    lookupLimit = 1000
    oauthPassThru = false
    serverName = "domain.example.com"
    tlsAuth = false
    tlsAuthWithCACert = false
    tlsSkipVerify = false
    tsdbResolution = 1
    tsdbVersion = 1
  })

  secure_json_data_encoded = jsonencode({
    basicAuthPassword = "<YOUR_PASSWORD>"
    tlsCACert = "<YOUR_CA_CERT>"
    tlsClientCert = "<YOUR_CLIENT_CERT>"
    tlsClientKey = "<YOUR_CLIENT_KEY>"
  })
}
```

