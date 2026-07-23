# InfluxDB configuration

Configuration reference for the **InfluxDB** data source (`influxdb`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/grafana/latest/datasources/influxdb/).

> Generated from [`dsconfig.json`](dsconfig.json). Do not edit by hand — run `go generate ./...` to refresh.

## Fields

| Field | Type | Target | Required | Description |
|---|---|---|---|---|
| `url` | string | root | yes | URL |
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
| `jsonData.timeout` | number | jsonData |  | HTTP request timeout in seconds. |
| `jsonData.httpHeaders` | list | jsonData |  | Additional HTTP headers sent with every request. Header names are stored in jsonData (httpHeaderName<N>); header values are write-only in secureJsonData (httpHeaderValue<N>). |
| `jsonData.httpHeaders[].name` | string | jsonData | yes | Header |
| `jsonData.httpHeaders[].value` | string | jsonData |  | Value |
| `jsonData.version` | enum (InfluxQL, SQL, Flux) | jsonData |  | Query language |
| `jsonData.product` | enum (InfluxDB Cloud Dedicated, InfluxDB Cloud Serverless, InfluxDB Clustered, InfluxDB Enterprise 1.x, InfluxDB Enterprise 3.x, InfluxDB Cloud (TSM), InfluxDB Cloud 1, InfluxDB OSS 1.x, InfluxDB OSS 2.x, InfluxDB OSS 3.x) | jsonData |  | Use InfluxDB detection to identify the product |
| `jsonData.pdcInjected` | boolean | jsonData |  | Backend-controlled indicator that a Private Datasource Connect (PDC) proxy has been injected into this datasource's HTTP transport. Not editor-writable; read by the v2 LeftSideBar to render PDC-specific section headers (LeftSideBar.tsx:12). |
| `jsonData.dbName` | string | jsonData | conditional | Database |
| `user` | string | root |  | Legacy root-level user field written by the v1 InfluxQL editor. Distinct from root.basicAuthUser — the v1 editor writes options.user (SDK root User field) while the v2 editor writes options.basicAuthUser (SDK root BasicAuthUser field). Not consumed by the current backend or the SDK HTTPClientOptions auth handler; effectively a display-only echo unless the operator also enables root.basicAuth so the SDK reads basicAuthUser instead. |
| `secureJsonData.password` 🔒 | string | secureJsonData |  | Legacy secure password paired with root.user for the v1 InfluxQL editor's User + Password inputs. Distinct from secureJsonData.basicAuthPassword. Not consumed by the current backend or SDK HTTP auth path. |
| `jsonData.httpMode` | enum (GET, POST) | jsonData |  | You can use either GET or POST HTTP method to query your InfluxDB database. The POST method allows you to perform heavy requests (with a lots of WHERE clause) while the GET method will restrict you and return an error if the query is too large. |
| `jsonData.timeInterval` | string | jsonData |  | A lower limit for the auto group by time interval. Recommended to be set to write frequency, for example 1m if your data is written every minute. |
| `jsonData.showTagTime` | string | jsonData |  | This time range is used in the query editor's autocomplete to reduce the execution time of tag filter queries. |
| `jsonData.organization` | string | jsonData | conditional | Organization |
| `secureJsonData.token` 🔒 | string | secureJsonData | conditional | Token |
| `jsonData.defaultBucket` | string | jsonData | conditional | Default Bucket |
| `jsonData.insecureGrpc` | boolean | jsonData |  | Disable TLS for the FlightSQL gRPC connection used by the SQL query path. |
| `jsonData.maxSeries` | number | jsonData |  | Limit the number of series/tables that Grafana will process. Lower this number to prevent abuse, and increase it if you have lots of small time series and not all are shown. Defaults to 1000. |

## Provisioning examples

Each scenario below shows how to provision the data source in Grafana using a YAML file (loaded by Grafana's [file provisioner](https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources)) and using the [Grafana Terraform provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source).

Placeholders like `<YOUR_TOKEN>` must be replaced with real values before use.

### `InfluxQL`

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: "InfluxDB"
    type: "influxdb"
    access: proxy
    basicAuth: false
    basicAuthUser: user
    url: "http://localhost:8086"
    withCredentials: false
    jsonData:
      dbName: mydb
      httpMode: GET
      maxSeries: 1000
      oauthPassThru: false
      pdcInjected: false
      serverName: domain.example.com
      tlsAuth: false
      tlsAuthWithCACert: false
      tlsSkipVerify: false
      version: "InfluxQL"
    secureJsonData:
      basicAuthPassword: "<YOUR_PASSWORD>"
      tlsCACert: "<YOUR_CA_CERT>"
      tlsClientCert: "<YOUR_CLIENT_CERT>"
      tlsClientKey: "<YOUR_CLIENT_KEY>"
```

**Terraform**

```hcl
resource "grafana_data_source" "influxdb_InfluxQL" {
  type = "influxdb"
  name = "InfluxDB"
  url = "http://localhost:8086"

  json_data_encoded = jsonencode({
    dbName = "mydb"
    httpMode = "GET"
    maxSeries = 1000
    oauthPassThru = false
    pdcInjected = false
    serverName = "domain.example.com"
    tlsAuth = false
    tlsAuthWithCACert = false
    tlsSkipVerify = false
    version = "InfluxQL"
  })

  secure_json_data_encoded = jsonencode({
    basicAuthPassword = "<YOUR_PASSWORD>"
    tlsCACert = "<YOUR_CA_CERT>"
    tlsClientCert = "<YOUR_CLIENT_CERT>"
    tlsClientKey = "<YOUR_CLIENT_KEY>"
  })
}
```

### `SQL`

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: "InfluxDB"
    type: "influxdb"
    access: proxy
    basicAuth: false
    basicAuthUser: user
    url: "http://localhost:8086"
    withCredentials: false
    jsonData:
      dbName: mydb
      insecureGrpc: false
      maxSeries: 1000
      oauthPassThru: false
      pdcInjected: false
      serverName: domain.example.com
      tlsAuth: false
      tlsAuthWithCACert: false
      tlsSkipVerify: false
      version: SQL
    secureJsonData:
      basicAuthPassword: "<YOUR_PASSWORD>"
      tlsCACert: "<YOUR_CA_CERT>"
      tlsClientCert: "<YOUR_CLIENT_CERT>"
      tlsClientKey: "<YOUR_CLIENT_KEY>"
      token: "<YOUR_TOKEN>"
```

**Terraform**

```hcl
resource "grafana_data_source" "influxdb_SQL" {
  type = "influxdb"
  name = "InfluxDB"
  url = "http://localhost:8086"

  json_data_encoded = jsonencode({
    dbName = "mydb"
    insecureGrpc = false
    maxSeries = 1000
    oauthPassThru = false
    pdcInjected = false
    serverName = "domain.example.com"
    tlsAuth = false
    tlsAuthWithCACert = false
    tlsSkipVerify = false
    version = "SQL"
  })

  secure_json_data_encoded = jsonencode({
    basicAuthPassword = "<YOUR_PASSWORD>"
    tlsCACert = "<YOUR_CA_CERT>"
    tlsClientCert = "<YOUR_CLIENT_CERT>"
    tlsClientKey = "<YOUR_CLIENT_KEY>"
    token = "<YOUR_TOKEN>"
  })
}
```

### `Flux`

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: "InfluxDB"
    type: "influxdb"
    access: proxy
    basicAuth: false
    basicAuthUser: user
    url: "http://localhost:8086"
    withCredentials: false
    jsonData:
      defaultBucket: default bucket
      maxSeries: 1000
      oauthPassThru: false
      organization: myorg
      pdcInjected: false
      serverName: domain.example.com
      tlsAuth: false
      tlsAuthWithCACert: false
      tlsSkipVerify: false
      version: Flux
    secureJsonData:
      basicAuthPassword: "<YOUR_PASSWORD>"
      tlsCACert: "<YOUR_CA_CERT>"
      tlsClientCert: "<YOUR_CLIENT_CERT>"
      tlsClientKey: "<YOUR_CLIENT_KEY>"
      token: "<YOUR_TOKEN>"
```

**Terraform**

```hcl
resource "grafana_data_source" "influxdb_Flux" {
  type = "influxdb"
  name = "InfluxDB"
  url = "http://localhost:8086"

  json_data_encoded = jsonencode({
    defaultBucket = "default bucket"
    maxSeries = 1000
    oauthPassThru = false
    organization = "myorg"
    pdcInjected = false
    serverName = "domain.example.com"
    tlsAuth = false
    tlsAuthWithCACert = false
    tlsSkipVerify = false
    version = "Flux"
  })

  secure_json_data_encoded = jsonencode({
    basicAuthPassword = "<YOUR_PASSWORD>"
    tlsCACert = "<YOUR_CA_CERT>"
    tlsClientCert = "<YOUR_CLIENT_CERT>"
    tlsClientKey = "<YOUR_CLIENT_KEY>"
    token = "<YOUR_TOKEN>"
  })
}
```

