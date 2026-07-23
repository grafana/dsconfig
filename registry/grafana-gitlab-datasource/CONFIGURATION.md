# GitLab configuration

Configuration reference for the **GitLab** data source (`grafana-gitlab-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-gitlab-datasource).

> Generated from [`dsconfig.json`](dsconfig.json). Do not edit by hand — run `go generate ./...` to refresh.

## Fields

| Field | Type | Target | Required | Description |
|---|---|---|---|---|
| `url` | string | root |  | The URL for your GitLab instance (ex: gitlab.domain.com). Leave blank if you use gitlab.com |
| `secureJsonData.accessToken` 🔒 | string | secureJsonData | yes | Provide information to grant access to this data source. To learn more about access tokens, [click here.](https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html) |
| `jsonData.pageLimit` | number | jsonData |  | The page limit is the maximum number of pages returned when creating a query. The default is 5. |

## Provisioning examples

Each scenario below shows how to provision the data source in Grafana using a YAML file (loaded by Grafana's [file provisioner](https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources)) and using the [Grafana Terraform provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source).

Placeholders like `<YOUR_TOKEN>` must be replaced with real values before use.

### Default configuration

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: GitLab
    type: grafana-gitlab-datasource
    access: proxy
    url: "https://gitlab.com/api/v4"
    jsonData:
      pageLimit: 5
    secureJsonData:
      accessToken: "<YOUR_ACCESS_TOKEN>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_gitlab_datasource" {
  type = "grafana-gitlab-datasource"
  name = "GitLab"
  url = "https://gitlab.com/api/v4"

  json_data_encoded = jsonencode({
    pageLimit = 5
  })

  secure_json_data_encoded = jsonencode({
    accessToken = "<YOUR_ACCESS_TOKEN>"
  })
}
```

