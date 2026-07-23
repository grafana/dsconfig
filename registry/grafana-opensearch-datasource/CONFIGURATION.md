# OpenSearch configuration

Configuration reference for the **OpenSearch** data source (`grafana-opensearch-datasource`) in Grafana.

For more information, see the [official documentation](https://github.com/grafana/opensearch-datasource).

> Generated from [`dsconfig.json`](dsconfig.json). Do not edit by hand — run `go generate ./...` to refresh.

## Fields

| Field | Type | Target | Required | Description |
|---|---|---|---|---|
| `url` | string | root | yes | Specify a complete HTTP URL (for example http://your_server:8080) |
| `access` | enum (proxy, direct) | root |  | Access |
| `basicAuth` | boolean | root |  | Basic auth |
| `basicAuthUser` | string | root | conditional | User |
| `secureJsonData.basicAuthPassword` 🔒 | string | secureJsonData | conditional | Password |
| `withCredentials` | boolean | root |  | Whether credentials such as cookies or auth headers should be sent with cross-site requests. |
| `jsonData.oauthPassThru` | boolean | jsonData |  | Forward the user's upstream OAuth identity to the data source (Their access token gets passed along). |
| `jsonData.sigV4Auth` | boolean | jsonData |  | SigV4 auth |
| `jsonData.tlsAuth` | boolean | jsonData |  | TLS Client Auth |
| `jsonData.serverName` | string | jsonData | conditional | ServerName |
| `secureJsonData.tlsClientCert` 🔒 | string (multiline) | secureJsonData | conditional | Client Cert |
| `secureJsonData.tlsClientKey` 🔒 | string (multiline) | secureJsonData | conditional | Client Key |
| `jsonData.tlsAuthWithCACert` | boolean | jsonData |  | Needed for verifying self-signed TLS Certs |
| `secureJsonData.tlsCACert` 🔒 | string (multiline) | secureJsonData | conditional | CA Cert |
| `jsonData.tlsSkipVerify` | boolean | jsonData |  | Skip TLS Verify |
| `jsonData.keepCookies` | list | jsonData |  | Grafana proxy deletes forwarded cookies by default. Specify cookies by name that should be forwarded to the data source. |
| `jsonData.timeout` | number | jsonData |  | HTTP request timeout in seconds |
| `jsonData.httpHeaders` | list | jsonData |  | Additional HTTP headers sent with every request. Header names are stored in jsonData (httpHeaderName<N>); header values are write-only in secureJsonData (httpHeaderValue<N>). |
| `jsonData.httpHeaders[].name` | string | jsonData | yes | Header |
| `jsonData.httpHeaders[].value` | string | jsonData |  | Value |
| `jsonData.database` | string | jsonData | yes | Index name |
| `jsonData.interval` | enum (, Hourly, Daily, Weekly, Monthly, Yearly) | jsonData |  | Pattern |
| `jsonData.timeField` | string | jsonData | yes | Time field name |
| `jsonData.serverless` | boolean | jsonData |  | If this is a DataSource to query a serverless OpenSearch service. |
| `jsonData.flavor` | enum (opensearch, elasticsearch) | jsonData |  |  |
| `jsonData.version` | string | jsonData | yes | Version |
| `jsonData.versionLabel` | string | jsonData |  |  |
| `jsonData.maxConcurrentShardRequests` | number | jsonData |  | Max concurrent Shard Requests |
| `jsonData.timeInterval` | string | jsonData |  | A lower limit for the auto group by time interval. Recommended to be set to write frequency, for example 1m if your data is written every minute. |
| `jsonData.pplEnabled` | boolean | jsonData |  | Allow Piped Processing Language as an alternative query syntax in the OpenSearch query editor. |
| `jsonData.logMessageField` | string | jsonData |  | Message field name |
| `jsonData.logLevelField` | string | jsonData |  | Level field name |
| `jsonData.dataLinks` | list | jsonData |  | Add links to existing fields. Links will be shown in log row details next to the field value. |
| `jsonData.dataLinks[].field` | string | jsonData | yes | Can be exact field name or a regex pattern that will match on the field name. |
| `jsonData.dataLinks[].title` | string | jsonData |  | Title |
| `jsonData.dataLinks[].url` | string | jsonData |  | URL |
| `jsonData.dataLinks[].datasourceUid` | string | jsonData |  | Internal link |

## Provisioning examples

Each scenario below shows how to provision the data source in Grafana using a YAML file (loaded by Grafana's [file provisioner](https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources)) and using the [Grafana Terraform provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source).

Placeholders like `<YOUR_TOKEN>` must be replaced with real values before use.

### `opensearch`

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: OpenSearch
    type: grafana-opensearch-datasource
    access: proxy
    basicAuth: false
    basicAuthUser: user
    url: "http://localhost:9200"
    withCredentials: false
    jsonData:
      database: es-index-name
      flavor: opensearch
      oauthPassThru: false
      pplEnabled: true
      serverName: domain.example.com
      serverless: false
      sigV4Auth: false
      timeField: "@timestamp"
      tlsAuth: false
      tlsAuthWithCACert: false
      tlsSkipVerify: false
      version: version required
    secureJsonData:
      basicAuthPassword: "<YOUR_PASSWORD>"
      tlsCACert: "<YOUR_CA_CERT>"
      tlsClientCert: "<YOUR_CLIENT_CERT>"
      tlsClientKey: "<YOUR_CLIENT_KEY>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_opensearch_datasource_opensearch" {
  type = "grafana-opensearch-datasource"
  name = "OpenSearch"
  access = "proxy"
  url = "http://localhost:9200"

  json_data_encoded = jsonencode({
    database = "es-index-name"
    flavor = "opensearch"
    oauthPassThru = false
    pplEnabled = true
    serverName = "domain.example.com"
    serverless = false
    sigV4Auth = false
    timeField = "@timestamp"
    tlsAuth = false
    tlsAuthWithCACert = false
    tlsSkipVerify = false
    version = "version required"
  })

  secure_json_data_encoded = jsonencode({
    basicAuthPassword = "<YOUR_PASSWORD>"
    tlsCACert = "<YOUR_CA_CERT>"
    tlsClientCert = "<YOUR_CLIENT_CERT>"
    tlsClientKey = "<YOUR_CLIENT_KEY>"
  })
}
```

### `elasticsearch`

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: OpenSearch
    type: grafana-opensearch-datasource
    access: proxy
    basicAuth: false
    basicAuthUser: user
    url: "http://localhost:9200"
    withCredentials: false
    jsonData:
      database: es-index-name
      flavor: elasticsearch
      oauthPassThru: false
      pplEnabled: true
      serverName: domain.example.com
      serverless: false
      sigV4Auth: false
      timeField: "@timestamp"
      tlsAuth: false
      tlsAuthWithCACert: false
      tlsSkipVerify: false
      version: version required
    secureJsonData:
      basicAuthPassword: "<YOUR_PASSWORD>"
      tlsCACert: "<YOUR_CA_CERT>"
      tlsClientCert: "<YOUR_CLIENT_CERT>"
      tlsClientKey: "<YOUR_CLIENT_KEY>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_opensearch_datasource_elasticsearch" {
  type = "grafana-opensearch-datasource"
  name = "OpenSearch"
  access = "proxy"
  url = "http://localhost:9200"

  json_data_encoded = jsonencode({
    database = "es-index-name"
    flavor = "elasticsearch"
    oauthPassThru = false
    pplEnabled = true
    serverName = "domain.example.com"
    serverless = false
    sigV4Auth = false
    timeField = "@timestamp"
    tlsAuth = false
    tlsAuthWithCACert = false
    tlsSkipVerify = false
    version = "version required"
  })

  secure_json_data_encoded = jsonencode({
    basicAuthPassword = "<YOUR_PASSWORD>"
    tlsCACert = "<YOUR_CA_CERT>"
    tlsClientCert = "<YOUR_CLIENT_CERT>"
    tlsClientKey = "<YOUR_CLIENT_KEY>"
  })
}
```

