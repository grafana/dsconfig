# SAP HANA® configuration

Configuration reference for the **SAP HANA®** data source (`grafana-saphana-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-saphana-datasource).

> Generated from [`dsconfig.json`](dsconfig.json). Do not edit by hand — run `go generate ./...` to refresh.

## Fields

| Field | Type | Target | Required | Description |
|---|---|---|---|---|
| `jsonData.server` | string | jsonData | yes | SAP HANA server address |
| `jsonData.port` | number | jsonData | conditional | SAP HANA server port (optional if database name filled). Typically this will be 443 for SAP HANA Cloud. For on-prem/multi-tenanted instances, use the corresponding port number |
| `jsonData.username` | string | jsonData | conditional | SAP HANA username |
| `secureJsonData.password` 🔒 | string | secureJsonData | conditional | SAP HANA password |
| `jsonData.tlsDisabled` | boolean | jsonData |  | Enable TLS/SSL encryption for the connection to SAP HANA. Enabled by default. Disable only when your SAP HANA instance does not have TLS configured (plaintext connections). |
| `jsonData.tlsSkipVerify` | boolean | jsonData |  | Skip TLS Verify |
| `jsonData.tlsAuth` | boolean | jsonData |  | TLS Client Auth |
| `secureJsonData.tlsClientCert` 🔒 | string (multiline) | secureJsonData | conditional | Client Cert |
| `secureJsonData.tlsClientKey` 🔒 | string (multiline) | secureJsonData | conditional | Client Key |
| `jsonData.tlsAuthWithCACert` | boolean | jsonData |  | Needed for verifying self-signed TLS Certs |
| `secureJsonData.tlsCACert` 🔒 | string (multiline) | secureJsonData |  | CA Cert |
| `jsonData.databaseName` | string | jsonData |  | Tenant database name (optional). If database name is set as well as the instance number, port is not required. |
| `jsonData.instance` | string | jsonData |  | SAP HANA tenant instance number (optional). If instance number is set, port is not required. |
| `jsonData.defaultSchema` | string | jsonData |  | Default schema to be used. Can be empty. |
| `jsonData.timeout` | string | jsonData |  | Timeout |

## Provisioning examples

Each scenario below shows how to provision the data source in Grafana using a YAML file (loaded by Grafana's [file provisioner](https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources)) and using the [Grafana Terraform provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source).

Placeholders like `<YOUR_TOKEN>` must be replaced with real values before use.

### Default configuration

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: SAP HANA®
    type: grafana-saphana-datasource
    access: proxy
    jsonData:
      port: 0
      server: Server address
      timeout: "30"
      tlsAuth: false
      tlsAuthWithCACert: false
      tlsDisabled: false
      tlsSkipVerify: false
      username: Username
    secureJsonData:
      password: "<YOUR_PASSWORD>"
      tlsClientCert: "<YOUR_CLIENT_CERT>"
      tlsClientKey: "<YOUR_CLIENT_KEY>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_saphana_datasource" {
  type = "grafana-saphana-datasource"
  name = "SAP HANA®"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    port = 0
    server = "Server address"
    timeout = "30"
    tlsAuth = false
    tlsAuthWithCACert = false
    tlsDisabled = false
    tlsSkipVerify = false
    username = "Username"
  })

  secure_json_data_encoded = jsonencode({
    password = "<YOUR_PASSWORD>"
    tlsClientCert = "<YOUR_CLIENT_CERT>"
    tlsClientKey = "<YOUR_CLIENT_KEY>"
  })
}
```

