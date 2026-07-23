# Falcon LogScale configuration

Configuration reference for the **Falcon LogScale** data source (`grafana-falconlogscale-datasource`) in Grafana.

For more information, see the [official documentation](https://github.com/grafana/falconlogscale-datasource).

> Generated from [`dsconfig.json`](dsconfig.json). Do not edit by hand — run `go generate ./...` to refresh.

## Fields

| Field | Type | Target | Required | Description |
|---|---|---|---|---|
| `url` | string | root | yes | URL |
| `jsonData.mode` | enum (LogScale, NGSIEM) | jsonData |  | Select the data source mode. NGSIEM mode only supports OAuth2 client secret authentication. |
| `jsonData.authenticateWithToken` | boolean | jsonData |  |  |
| `jsonData.oauth2` | boolean | jsonData |  |  |
| `jsonData.oauthPassThru` | boolean | jsonData |  |  |
| `basicAuth` | boolean | root |  |  |
| `secureJsonData.accessToken` 🔒 | string | secureJsonData | conditional | Token |
| `jsonData.oauth2ClientId` | string | jsonData | conditional | The OAuth2 client ID |
| `secureJsonData.oauth2ClientSecret` 🔒 | string | secureJsonData | conditional | The OAuth2 client secret |
| `basicAuthUser` | string | root | conditional | User |
| `secureJsonData.basicAuthPassword` 🔒 | string | secureJsonData | conditional | Password |
| `jsonData.baseUrl` | string | jsonData |  | Snapshot of the datasource URL written by the LogScale token authentication component (src/components/ConfigEditor/ConfigEditor.tsx:155). Never read by the backend, which always reads settings.URL (pkg/plugin/settings.go:38). Preserved for round-trip fidelity. |
| `jsonData.defaultRepository` | enum | jsonData |  | Default Repository |
| `jsonData.dataLinks` | list | jsonData |  | Add links to existing fields. Links will be shown in log row details next to the field value. |
| `jsonData.dataLinks[].field` | string | jsonData | yes | Can be exact field name or a regex pattern that will match on the field name. |
| `jsonData.dataLinks[].label` | string | jsonData |  | Use to provide a meaningful label to the data matched in the regex |
| `jsonData.dataLinks[].matcherRegex` | string | jsonData | yes | Use to parse and capture some part of the log message. You can use the captured groups in the template. |
| `jsonData.dataLinks[].url` | string | jsonData |  | URL |
| `jsonData.dataLinks[].datasourceUid` | string | jsonData |  | UID of a Grafana data source. When set, the derived data link is treated as an internal link to that data source and the URL field is interpreted as a Query. |
| `jsonData.incrementalQuerying` | boolean | jsonData |  | Results may be incomplete or incorrect in some cases. On auto-refresh, query new data and merge it with the cached result. This applies only to relative time ranges without aggregation functions. |
| `jsonData.incrementalQueryOverlapWindow` | string | jsonData |  | Time window to re-fetch on each incremental query to catch late-arriving data (e.g. "10m", "30s", "1h"). Changes take effect after saving and reloading. |
| `jsonData.keepCookies` | list | jsonData |  | Grafana proxy deletes forwarded cookies by default. Specify cookies by name that should be forwarded to the data source. |
| `jsonData.timeout` | number | jsonData |  | HTTP request timeout in seconds |
| `jsonData.httpHeaders` | list | jsonData |  | Additional HTTP headers sent with every request. Header names are stored in jsonData (httpHeaderName<N>); header values are write-only in secureJsonData (httpHeaderValue<N>). |
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
  - name: Falcon LogScale
    type: grafana-falconlogscale-datasource
    access: proxy
    basicAuth: false
    basicAuthUser: User
    url: "<YOUR_URL>"
    jsonData:
      authenticateWithToken: false
      incrementalQueryOverlapWindow: "10m"
      incrementalQuerying: false
      mode: LogScale
      oauth2: false
      oauth2ClientId: Client ID
      oauthPassThru: false
    secureJsonData:
      accessToken: "<YOUR_TOKEN>"
      basicAuthPassword: "<YOUR_PASSWORD>"
      oauth2ClientSecret: "<YOUR_CLIENT_SECRET>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_falconlogscale_datasource" {
  type = "grafana-falconlogscale-datasource"
  name = "Falcon LogScale"
  url = "<YOUR_URL>"

  json_data_encoded = jsonencode({
    authenticateWithToken = false
    incrementalQueryOverlapWindow = "10m"
    incrementalQuerying = false
    mode = "LogScale"
    oauth2 = false
    oauth2ClientId = "Client ID"
    oauthPassThru = false
  })

  secure_json_data_encoded = jsonencode({
    accessToken = "<YOUR_TOKEN>"
    basicAuthPassword = "<YOUR_PASSWORD>"
    oauth2ClientSecret = "<YOUR_CLIENT_SECRET>"
  })
}
```

