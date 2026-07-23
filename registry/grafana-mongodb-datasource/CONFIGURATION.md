# MongoDB configuration

Configuration reference for the **MongoDB** data source (`grafana-mongodb-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-mongodb-datasource).

> Generated from [`dsconfig.json`](dsconfig.json). Do not edit by hand — run `go generate ./...` to refresh.

## Fields

| Field | Type | Target | Required | Description |
|---|---|---|---|---|
| `jsonData.connection` | string | jsonData | yes | A connection string contains the parameters required to connect to MongoDB. |
| `jsonData.authType` | enum (NoAuth, BasicAuth, custom-Kerberos) | jsonData |  | Choose an authentication method to access the data source |
| `basicAuth` | boolean | root |  | Standard Grafana basic-auth enabled flag. The backend initializes BasicAuthEnabled from this value and also forces it on when a username or password is present (pkg/models/settings.go). Set by datasource provisioning and by the config editor's on-load migration of legacy datasources, not by selecting the Credentials method in the current editor. |
| `basicAuthUser` | string | root |  | The username assigned to the MongoDB account. |
| `secureJsonData.basicAuthPassword` 🔒 | string | secureJsonData |  | The password assigned to the MongoDB account. |
| `jsonData.kerberosUser` | string | jsonData |  | The client principal's username. Enabled when connection string includes query string authMethod=GSSAPI |
| `secureJsonData.kerberosPassword` 🔒 | string | secureJsonData |  | The client principal password that will be used to authenticate. Optional if a keytab or cache file is present. |
| `jsonData.keyTabFilePath` | string | jsonData |  | Absolute file path KeyTab for keytab file. If present will ignore password. Enabled when connection string includes query string authMethod=GSSAPI |
| `jsonData.globalCcacheFilePath` | string | jsonData |  | Absolute file path to global compiler cache (ccache) file. If present will ignore password. Enabled when connection string includes query string authMethod=GSSAPI |
| `jsonData.ccacheLookupFile` | string | jsonData |  | Absolute file path to  the JSON file that provides the Kerberos compiler cache (ccache) based on username principal and connection string. If present will ignore password. Enabled when connection string includes query string authMethod=GSSAPI |
| `jsonData.validate` | boolean | jsonData |  | Enable real time query syntax validation. MongoDB BSON syntax will be validated as you type and show contextual errors. |
| `secureJsonData.tlsCertificateKeyFilePassword` 🔒 | string | secureJsonData |  | Password |
| `jsonData.responseRowsLimit` | string | jsonData |  | Increasing this too much may lead to performance issues for larger queries |
| `jsonData.httpHeaders` | list | jsonData |  | Additional HTTP headers sent with every request. Header names are stored in jsonData (httpHeaderName<N>); header values are write-only in secureJsonData (httpHeaderValue<N>). |
| `jsonData.httpHeaders[].name` | string | jsonData | yes | Header |
| `jsonData.httpHeaders[].value` | string | jsonData |  | Value |
| `jsonData.serverName` | string | jsonData |  | TLS server name used to verify the hostname on the server's certificate when tlsAuth is enabled. Not exposed in the configuration editor; set via datasource provisioning. |
| `jsonData.tlsAuth` | boolean | jsonData |  | Enables TLS client-certificate authentication; the backend supplies tlsClientCert and tlsClientKey to the server. Not exposed in the configuration editor; set via datasource provisioning. |
| `jsonData.tlsAuthWithCACert` | boolean | jsonData |  | Enables verification of the server's TLS certificate against a custom CA (tlsCACert). Not exposed in the configuration editor; set via datasource provisioning. |
| `jsonData.tlsSkipVerify` | boolean | jsonData |  | Skips verification of the server's TLS certificate chain and host name (applied by the backend when tlsAuthWithCACert is enabled). Not exposed in the configuration editor; set via datasource provisioning. |
| `secureJsonData.tlsCACert` 🔒 | string | secureJsonData |  | CA certificate PEM used to verify the server's TLS certificate when tlsAuthWithCACert is enabled. Not exposed in the configuration editor; set via datasource provisioning. |
| `secureJsonData.tlsClientCert` 🔒 | string | secureJsonData |  | Client certificate PEM used when tlsAuth is enabled. Not exposed in the configuration editor; set via datasource provisioning. |
| `secureJsonData.tlsClientKey` 🔒 | string | secureJsonData |  | Client private key PEM used when tlsAuth is enabled. Not exposed in the configuration editor; set via datasource provisioning. |
| `jsonData.user` | string | jsonData |  | Legacy username field. Datasources created before v1.9.0 stored the MongoDB username here; the backend migrates it to the root basicAuthUser and enables basic auth (pkg/models/settings.go). New configurations use the root basicAuthUser field. |
| `jsonData.skipTLSValidation` | boolean | jsonData |  | Legacy flag; the backend copies it to tlsSkipVerify (InsecureSkipVerify) at load time (pkg/models/settings.go). New configurations use tlsSkipVerify. |
| `jsonData.credentials` | boolean | jsonData |  | Legacy frontend-only flag used by the config editor's on-load migration to detect pre-v1.9.0 basic-auth datasources. Never read by the backend. |
| `secureJsonData.password` 🔒 | string | secureJsonData |  | Legacy secure password. Datasources created before v1.9.0 stored the MongoDB password here; the backend migrates it to the basic-auth password (pkg/models/settings.go). New configurations use basicAuthPassword. |

## Provisioning examples

Each scenario below shows how to provision the data source in Grafana using a YAML file (loaded by Grafana's [file provisioner](https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources)) and using the [Grafana Terraform provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source).

Placeholders like `<YOUR_TOKEN>` must be replaced with real values before use.

### No Authentication (`NoAuth`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: MongoDB
    type: grafana-mongodb-datasource
    access: proxy
    jsonData:
      authType: NoAuth
      connection: "mongodb+srv://cluster.host.net/dbname?retryWrites=true&w=majority"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_mongodb_datasource_NoAuth" {
  type = "grafana-mongodb-datasource"
  name = "MongoDB"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    authType = "NoAuth"
    connection = "mongodb+srv://cluster.host.net/dbname?retryWrites=true&w=majority"
  })
}
```

### Credentials (`BasicAuth`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: MongoDB
    type: grafana-mongodb-datasource
    access: proxy
    jsonData:
      authType: BasicAuth
      connection: "mongodb+srv://cluster.host.net/dbname?retryWrites=true&w=majority"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_mongodb_datasource_BasicAuth" {
  type = "grafana-mongodb-datasource"
  name = "MongoDB"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    authType = "BasicAuth"
    connection = "mongodb+srv://cluster.host.net/dbname?retryWrites=true&w=majority"
  })
}
```

### Kerberos (`custom-Kerberos`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: MongoDB
    type: grafana-mongodb-datasource
    access: proxy
    jsonData:
      authType: custom-Kerberos
      connection: "mongodb+srv://cluster.host.net/dbname?retryWrites=true&w=majority"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_mongodb_datasource_custom_Kerberos" {
  type = "grafana-mongodb-datasource"
  name = "MongoDB"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    authType = "custom-Kerberos"
    connection = "mongodb+srv://cluster.host.net/dbname?retryWrites=true&w=majority"
  })
}
```

