# o11y-bench Results — Datasource Config

_Generated from o11y-bench jobs — mcp as is: `anthropic-claude-sonnet-4-6-off-k3` (2026-07-19); no tools: `anthropic-claude-sonnet-4-6-off-k3` (2026-07-20)._

Benchmark of an LLM agent on the `datasource_config` task category (creating, editing, and explaining Grafana datasources via mcp-grafana tools).

📊 Full HTML reports with transcripts — [mcp as is](./report.html) · [no tools](./report_notools.html)

## Summary

| Metric | mcp as is | no tools | no schema |
|---|---|---|---|
| Model | Sonnet 4.6 | Sonnet 4.6 | — |
| Tasks | 26 | 26 | — |
| pass^3 (consistent) | 13/26 (50%) | 12/26 (46%) | — |
| pass@3 (any) | 17/26 (65%) | 16/26 (62%) | — |
| Mean score | 83% | 83% | — |
| Cost | $3.03 | $2.83 | — |
| Steps/trial | 2.6 | 2.3 | — |

- **pass^3** — task passes only if *all* 3 attempts pass (strict consistency).
- **pass@3** — task passes if *any* of 3 attempts pass.
- **Mean score** — average per-trial score (0–100%) across all trials.

## Per-task best score

| Task | mcp as is | no tools | no schema |
|---|---|---|---|
| `add-bigquery` | 100% | 100% | — |
| `add-clickhouse` | 100% | 50% | — |
| `add-infinity` | 60% | 60% | — |
| `add-infinity-auth` | 85% | 100% | — |
| `add-influxdb` | 100% | 100% | — |
| `add-loki` | 45% | 45% | — |
| `add-mysql` | 80% | 50% | — |
| `add-postgres` | 60% | 60% | — |
| `add-prometheus` | 55% | 95% | — |
| `add-tempo` | 45% | 60% | — |
| `check-datasource-health` | 100% | 100% | — |
| `diagnose-unhealthy-datasource` | 100% | 100% | — |
| `edit-clickhouse-protocol` | 100% | 100% | — |
| `edit-influxdb-database` | 100% | 100% | — |
| `edit-loki-derived-fields` | 100% | 100% | — |
| `edit-postgres-enable-tls` | 100% | 100% | — |
| `edit-prometheus-scrape-interval` | 100% | 100% | — |
| `edit-tempo-traces-to-logs` | 60% | 60% | — |
| `explain-bigquery-auth` | 100% | 70% | — |
| `explain-clickhouse-protocol-choice` | 85% | 85% | — |
| `explain-infinity-config` | 100% | 100% | — |
| `explain-influxdb-query-language` | 100% | 100% | — |
| `explain-postgres-tls-options` | 100% | 100% | — |
| `explain-prometheus-type-and-auth` | 100% | 100% | — |
| `provision-datasource-terraform` | 100% | 100% | — |
| `provision-datasources-yaml` | 100% | 100% | — |

> Per-task **best score** is the highest of the 3 attempts for that mode (matches the HTML report).

<!-- BENCH_DATA
{
  "asis": {
    "generated": "2026-07-19",
    "job": "anthropic-claude-sonnet-4-6-off-k3",
    "model": "Sonnet 4.6",
    "shots_per_task": 3,
    "total_tasks": 26,
    "tasks_passed": 17,
    "tasks_consistent": 13,
    "pass_rate": 0.6538461538461539,
    "pass_hat_rate": 0.5,
    "mean_score": 0.8326923076923077,
    "total_cost": 3.0284562,
    "steps_per_trial": "2.6",
    "task_scores": {
      "add-bigquery": 1.0,
      "add-clickhouse": 1.0,
      "add-infinity-auth": 0.85,
      "add-infinity": 0.6,
      "add-influxdb": 1.0,
      "add-loki": 0.45,
      "add-mysql": 0.8,
      "add-postgres": 0.6000000000000001,
      "add-prometheus": 0.55,
      "add-tempo": 0.45,
      "check-datasource-health": 1.0,
      "diagnose-unhealthy-datasource": 1.0,
      "edit-clickhouse-protocol": 1.0,
      "edit-influxdb-database": 1.0,
      "edit-loki-derived-fields": 1.0,
      "edit-postgres-enable-tls": 1.0,
      "edit-prometheus-scrape-interval": 1.0,
      "edit-tempo-traces-to-logs": 0.6000000000000001,
      "explain-bigquery-auth": 1.0,
      "explain-clickhouse-protocol-choice": 0.85,
      "explain-infinity-config": 1.0,
      "explain-influxdb-query-language": 1.0,
      "explain-postgres-tls-options": 1.0,
      "explain-prometheus-type-and-auth": 1.0,
      "provision-datasource-terraform": 1.0,
      "provision-datasources-yaml": 1.0
    },
    "task_passed": {
      "add-bigquery": true,
      "add-clickhouse": true,
      "add-infinity-auth": false,
      "add-infinity": false,
      "add-influxdb": true,
      "add-loki": false,
      "add-mysql": false,
      "add-postgres": false,
      "add-prometheus": false,
      "add-tempo": false,
      "check-datasource-health": true,
      "diagnose-unhealthy-datasource": true,
      "edit-clickhouse-protocol": true,
      "edit-influxdb-database": true,
      "edit-loki-derived-fields": true,
      "edit-postgres-enable-tls": true,
      "edit-prometheus-scrape-interval": true,
      "edit-tempo-traces-to-logs": false,
      "explain-bigquery-auth": true,
      "explain-clickhouse-protocol-choice": false,
      "explain-infinity-config": true,
      "explain-influxdb-query-language": true,
      "explain-postgres-tls-options": true,
      "explain-prometheus-type-and-auth": true,
      "provision-datasource-terraform": true,
      "provision-datasources-yaml": true
    },
    "task_consistent": {
      "add-bigquery": true,
      "add-clickhouse": true,
      "add-infinity-auth": false,
      "add-infinity": false,
      "add-influxdb": true,
      "add-loki": false,
      "add-mysql": false,
      "add-postgres": false,
      "add-prometheus": false,
      "add-tempo": false,
      "check-datasource-health": true,
      "diagnose-unhealthy-datasource": true,
      "edit-clickhouse-protocol": false,
      "edit-influxdb-database": true,
      "edit-loki-derived-fields": true,
      "edit-postgres-enable-tls": true,
      "edit-prometheus-scrape-interval": true,
      "edit-tempo-traces-to-logs": false,
      "explain-bigquery-auth": false,
      "explain-clickhouse-protocol-choice": false,
      "explain-infinity-config": true,
      "explain-influxdb-query-language": false,
      "explain-postgres-tls-options": true,
      "explain-prometheus-type-and-auth": false,
      "provision-datasource-terraform": true,
      "provision-datasources-yaml": true
    },
    "task_cost": {
      "add-bigquery": 0.10893059999999999,
      "add-clickhouse": 0.1564968,
      "add-infinity-auth": 0.1009236,
      "add-infinity": 0.11505539999999999,
      "add-influxdb": 0.14080679999999998,
      "add-loki": 0.170547,
      "add-mysql": 0.10544759999999999,
      "add-postgres": 0.110007,
      "add-prometheus": 0.2480541,
      "add-tempo": 0.17568599999999998,
      "check-datasource-health": 0.0526344,
      "diagnose-unhealthy-datasource": 0.0703308,
      "edit-clickhouse-protocol": 0.11089379999999999,
      "edit-influxdb-database": 0.1035768,
      "edit-loki-derived-fields": 0.12285779999999999,
      "edit-postgres-enable-tls": 0.10641779999999998,
      "edit-prometheus-scrape-interval": 0.0953928,
      "edit-tempo-traces-to-logs": 0.1718016,
      "explain-bigquery-auth": 0.1791999,
      "explain-clickhouse-protocol-choice": 0.0388272,
      "explain-infinity-config": 0.07517940000000001,
      "explain-influxdb-query-language": 0.056176199999999996,
      "explain-postgres-tls-options": 0.1947528,
      "explain-prometheus-type-and-auth": 0.0581472,
      "provision-datasource-terraform": 0.1065696,
      "provision-datasources-yaml": 0.0537432
    }
  },
  "notools": {
    "generated": "2026-07-20",
    "job": "anthropic-claude-sonnet-4-6-off-k3",
    "model": "Sonnet 4.6",
    "shots_per_task": 3,
    "total_tasks": 26,
    "tasks_passed": 16,
    "tasks_consistent": 12,
    "pass_rate": 0.6153846153846154,
    "pass_hat_rate": 0.46153846153846156,
    "mean_score": 0.8256578947368421,
    "total_cost": 2.8316113499999997,
    "steps_per_trial": "2.3",
    "task_scores": {
      "add-bigquery": 1.0,
      "add-clickhouse": 0.5,
      "add-infinity-auth": 1.0,
      "add-infinity": 0.6,
      "add-influxdb": 1.0,
      "add-loki": 0.45,
      "add-mysql": 0.5,
      "add-postgres": 0.6000000000000001,
      "add-prometheus": 0.9500000000000001,
      "add-tempo": 0.6,
      "check-datasource-health": 1.0,
      "diagnose-unhealthy-datasource": 1.0,
      "edit-clickhouse-protocol": 1.0,
      "edit-influxdb-database": 1.0,
      "edit-loki-derived-fields": 1.0,
      "edit-postgres-enable-tls": 1.0,
      "edit-prometheus-scrape-interval": 1.0,
      "edit-tempo-traces-to-logs": 0.6000000000000001,
      "explain-bigquery-auth": 0.7,
      "explain-clickhouse-protocol-choice": 0.85,
      "explain-infinity-config": 1.0,
      "explain-influxdb-query-language": 1.0,
      "explain-postgres-tls-options": 1.0,
      "explain-prometheus-type-and-auth": 1.0,
      "provision-datasource-terraform": 1.0,
      "provision-datasources-yaml": 1.0
    },
    "task_passed": {
      "add-bigquery": true,
      "add-clickhouse": false,
      "add-infinity-auth": true,
      "add-infinity": false,
      "add-influxdb": true,
      "add-loki": false,
      "add-mysql": false,
      "add-postgres": false,
      "add-prometheus": false,
      "add-tempo": false,
      "check-datasource-health": true,
      "diagnose-unhealthy-datasource": true,
      "edit-clickhouse-protocol": true,
      "edit-influxdb-database": true,
      "edit-loki-derived-fields": true,
      "edit-postgres-enable-tls": true,
      "edit-prometheus-scrape-interval": true,
      "edit-tempo-traces-to-logs": false,
      "explain-bigquery-auth": false,
      "explain-clickhouse-protocol-choice": false,
      "explain-infinity-config": true,
      "explain-influxdb-query-language": true,
      "explain-postgres-tls-options": true,
      "explain-prometheus-type-and-auth": true,
      "provision-datasource-terraform": true,
      "provision-datasources-yaml": true
    },
    "task_consistent": {
      "add-bigquery": true,
      "add-clickhouse": false,
      "add-infinity-auth": true,
      "add-infinity": false,
      "add-influxdb": true,
      "add-loki": false,
      "add-mysql": false,
      "add-postgres": false,
      "add-prometheus": false,
      "add-tempo": false,
      "check-datasource-health": true,
      "diagnose-unhealthy-datasource": true,
      "edit-clickhouse-protocol": false,
      "edit-influxdb-database": true,
      "edit-loki-derived-fields": false,
      "edit-postgres-enable-tls": true,
      "edit-prometheus-scrape-interval": true,
      "edit-tempo-traces-to-logs": false,
      "explain-bigquery-auth": false,
      "explain-clickhouse-protocol-choice": false,
      "explain-infinity-config": true,
      "explain-influxdb-query-language": false,
      "explain-postgres-tls-options": true,
      "explain-prometheus-type-and-auth": false,
      "provision-datasource-terraform": true,
      "provision-datasources-yaml": true
    },
    "task_cost": {
      "add-bigquery": 0.09271109999999999,
      "add-clickhouse": 0.0818091,
      "add-infinity-auth": 0.0630024,
      "add-infinity": 0.06372149999999999,
      "add-influxdb": 0.19682955,
      "add-loki": 0.0739782,
      "add-mysql": 0.10520445,
      "add-postgres": 0.1331568,
      "add-prometheus": 0.18016754999999998,
      "add-tempo": 0.16526865,
      "check-datasource-health": 0.0849171,
      "diagnose-unhealthy-datasource": 0.1077519,
      "edit-clickhouse-protocol": 0.08027309999999999,
      "edit-influxdb-database": 0.1018368,
      "edit-loki-derived-fields": 0.1254048,
      "edit-postgres-enable-tls": 0.16128465,
      "edit-prometheus-scrape-interval": 0.09728579999999999,
      "edit-tempo-traces-to-logs": 0.117792,
      "explain-bigquery-auth": 0.12169725,
      "explain-clickhouse-protocol-choice": 0.09662355,
      "explain-infinity-config": 0.1144299,
      "explain-influxdb-query-language": 0.10614255,
      "explain-postgres-tls-options": 0.0916473,
      "explain-prometheus-type-and-auth": 0.0558777,
      "provision-datasource-terraform": 0.1014381,
      "provision-datasources-yaml": 0.11135955
    }
  }
}
-->
