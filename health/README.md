# health

> Status: experimental — reference implementation for the RFC in
> [grafana/dsconfig#110](https://github.com/grafana/dsconfig/issues/110).

`health` normalizes data source `CheckHealth` failures (the errors behind
**Save & test**) into a consistent, safe, machine-classified and *actionable*
result:

- a stable, machine-readable **error code** (closed taxonomy),
- a clean, **secret-free message**,
- structured, context-aware **remediation**, and
- a redacted **verbose** detail for support.

The core depends **only** on `github.com/grafana/grafana-plugin-sdk-go`. Provider
knowledge (AWS, Azure, SQL drivers, …) is injected by family libraries via
`RegisterClassifier` / `RegisterRules`; the core never imports those SDKs.

## Usage

Transport / driver errors:

```go
func (d *Datasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
    if err := d.ping(ctx); err != nil {
        return health.Result(ctx, err,
            health.WithContext(health.Context{
                DatasourceType: "postgres",
                DatasourceName: req.PluginContext.DataSourceInstanceSettings.Name,
                AuthType:       "password",
            }),
            health.WithLogger(log.DefaultLogger),
        ), nil
    }
    return health.OK("Data source is working"), nil
}
```

HTTP data sources (response-aware, so HTML/odd bodies classify correctly):

```go
resp, err := client.Do(req)
if err != nil {
    return health.Result(ctx, err, health.WithContext(dims)), nil
}
defer resp.Body.Close()
body, _ := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
return health.ResultForResponse(ctx, resp, body, nil, health.WithContext(dims)), nil
```

Family libraries register provider classifiers + remediation once via `init()`
— see `Example_familyRegistration` in the tests.

## What's covered

| Area | Files |
|---|---|
| Taxonomy + canonical copy | `code.go` |
| Explicit error tagging | `errors.go` |
| Diagnosis dimensions | `diagnosis.go` |
| Classification pipeline (net/TLS/timeout, cancellation) | `classify.go` |
| HTML / inconsistent-JSON response handling | `httpresponse.go` |
| Secret redaction | `redact.go` |
| Specificity-ranked remediation rules | `remediation.go` |
| Shaping + UI/logs/metrics/trace fan-out | `result.go` |

Tests include table-driven unit cases and `httptest`-backed end-to-end scenarios
(HTML gateway, SSO redirect, WAF block, JSON error envelope, connection refused,
timeout, cancellation). Run them with:

```sh
go test ./...
```
