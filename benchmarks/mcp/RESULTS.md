# o11y-bench-2.0 Results — Datasource Config

_Generated from o11y-bench-2.0 jobs — mcp as is: `anthropic-claude-sonnet-4-6-off-k3` (2026-07-21); no tools: `anthropic-claude-sonnet-4-6-off-k3` (2026-07-21); no schema: `anthropic-claude-sonnet-4-6-off-k3` (2026-07-21)._

Benchmark of an LLM agent on the `datasource_config` task category (creating, editing, and explaining Grafana datasources via mcp-grafana tools).

📊 Full HTML reports with transcripts — [mcp as is](./report.html) · [no tools](./report_notools.html) · [no schema](./report_noschema.html)

## Summary

| Metric | mcp as is | no tools | no schema |
|---|---|---|---|
| Model | Sonnet 4.6 | Sonnet 4.6 | Sonnet 4.6 |
| Tasks | 27 | 27 | 27 |
| pass^3 (consistent) | 15/27 (56%) | 15/27 (56%) | 13/27 (48%) |
| pass@3 (any) | 23/27 (85%) | 20/27 (74%) | 20/27 (74%) |
| Mean score | 92% | 89% | 88% |
| Cost | $4.24 | $3.34 | $3.35 |
| Steps/trial | 2.7 | 2.3 | 2.3 |

- **pass^3** — task passes only if *all* 3 attempts pass (strict consistency).
- **pass@3** — task passes if *any* of 3 attempts pass.
- **Mean score** — average per-trial score (0–100%) across all trials.

## Per-task best score

| Task | mcp as is | no tools | no schema |
|---|---|---|---|
| `add-bigquery` | 100% | 100% | 85% |
| `add-clickhouse` | 100% | 60% | 100% |
| `add-infinity` | 100% | 95% | 100% |
| `add-infinity-auth` | 85% | 100% | 100% |
| `add-influxdb` | 100% | 100% | 100% |
| `add-loki` | 100% | 85% | 100% |
| `add-mysql` | 100% | 100% | 60% |
| `add-postgres` | 100% | 60% | 60% |
| `add-prometheus` | 100% | 100% | 100% |
| `add-tempo` | 65% | 65% | 65% |
| `check-datasource-health` | 100% | 100% | 100% |
| `diagnose-unhealthy-datasource` | 100% | 100% | 100% |
| `edit-clickhouse-protocol` | 100% | 100% | 100% |
| `edit-infinity-auth` | 100% | 100% | 100% |
| `edit-influxdb-database` | 100% | 100% | 100% |
| `edit-loki-derived-fields` | 100% | 100% | 100% |
| `edit-postgres-enable-tls` | 100% | 100% | 100% |
| `edit-prometheus-scrape-interval` | 85% | 85% | 85% |
| `edit-tempo-traces-to-logs` | 60% | 100% | 40% |
| `explain-bigquery-auth` | 100% | 70% | 85% |
| `explain-clickhouse-protocol-choice` | 100% | 100% | 100% |
| `explain-infinity-config` | 100% | 100% | 100% |
| `explain-influxdb-query-language` | 100% | 100% | 100% |
| `explain-postgres-tls-options` | 100% | 100% | 100% |
| `explain-prometheus-type-and-auth` | 100% | 100% | 100% |
| `provision-datasource-terraform` | 100% | 100% | 100% |
| `provision-datasources-yaml` | 100% | 100% | 100% |

> Per-task **best score** is the highest of the 3 attempts for that mode (matches the HTML report).

<!-- BENCH_DATA
{
  "asis": {
    "generated": "2026-07-21",
    "job": "anthropic-claude-sonnet-4-6-off-k3",
    "model": "Sonnet 4.6",
    "shots_per_task": 3,
    "total_tasks": 27,
    "tasks_passed": 23,
    "tasks_consistent": 15,
    "pass_rate": 0.8518518518518519,
    "pass_hat_rate": 0.5555555555555556,
    "mean_score": 0.9216049382716051,
    "total_cost": 4.2391182,
    "steps_per_trial": "2.7",
    "task_scores": {
      "add-bigquery": 1.0,
      "add-clickhouse": 1.0,
      "add-infinity-auth": 0.85,
      "add-infinity": 1.0,
      "add-influxdb": 1.0,
      "add-loki": 1.0,
      "add-mysql": 1.0,
      "add-postgres": 1.0,
      "add-prometheus": 1.0,
      "add-tempo": 0.65,
      "check-datasource-health": 1.0,
      "diagnose-unhealthy-datasource": 1.0,
      "edit-clickhouse-protocol": 1.0,
      "edit-infinity-auth": 1.0,
      "edit-influxdb-database": 1.0,
      "edit-loki-derived-fields": 1.0,
      "edit-postgres-enable-tls": 1.0,
      "edit-prometheus-scrape-interval": 0.8500000000000001,
      "edit-tempo-traces-to-logs": 0.6000000000000001,
      "explain-bigquery-auth": 1.0,
      "explain-clickhouse-protocol-choice": 1.0,
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
      "add-infinity": true,
      "add-influxdb": true,
      "add-loki": true,
      "add-mysql": true,
      "add-postgres": true,
      "add-prometheus": true,
      "add-tempo": false,
      "check-datasource-health": true,
      "diagnose-unhealthy-datasource": true,
      "edit-clickhouse-protocol": true,
      "edit-infinity-auth": true,
      "edit-influxdb-database": true,
      "edit-loki-derived-fields": true,
      "edit-postgres-enable-tls": true,
      "edit-prometheus-scrape-interval": false,
      "edit-tempo-traces-to-logs": false,
      "explain-bigquery-auth": true,
      "explain-clickhouse-protocol-choice": true,
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
      "add-loki": true,
      "add-mysql": false,
      "add-postgres": true,
      "add-prometheus": true,
      "add-tempo": false,
      "check-datasource-health": true,
      "diagnose-unhealthy-datasource": true,
      "edit-clickhouse-protocol": true,
      "edit-infinity-auth": true,
      "edit-influxdb-database": false,
      "edit-loki-derived-fields": false,
      "edit-postgres-enable-tls": false,
      "edit-prometheus-scrape-interval": false,
      "edit-tempo-traces-to-logs": false,
      "explain-bigquery-auth": false,
      "explain-clickhouse-protocol-choice": true,
      "explain-infinity-config": true,
      "explain-influxdb-query-language": true,
      "explain-postgres-tls-options": true,
      "explain-prometheus-type-and-auth": false,
      "provision-datasource-terraform": false,
      "provision-datasources-yaml": true
    },
    "task_cost": {
      "add-bigquery": 0.1972044,
      "add-clickhouse": 0.22233779999999997,
      "add-infinity-auth": 0.1921404,
      "add-infinity": 0.2715288,
      "add-influxdb": 0.23829419999999998,
      "add-loki": 0.15429299999999999,
      "add-mysql": 0.2130972,
      "add-postgres": 0.23509619999999998,
      "add-prometheus": 0.1952118,
      "add-tempo": 0.2156592,
      "check-datasource-health": 0.16965,
      "diagnose-unhealthy-datasource": 0.08207999999999999,
      "edit-clickhouse-protocol": 0.128991,
      "edit-infinity-auth": 0.1132266,
      "edit-influxdb-database": 0.12004319999999999,
      "edit-loki-derived-fields": 0.19662539999999995,
      "edit-postgres-enable-tls": 0.10012380000000001,
      "edit-prometheus-scrape-interval": 0.12307199999999999,
      "edit-tempo-traces-to-logs": 0.22064099999999998,
      "explain-bigquery-auth": 0.1770135,
      "explain-clickhouse-protocol-choice": 0.037773,
      "explain-infinity-config": 0.074118,
      "explain-influxdb-query-language": 0.052107,
      "explain-postgres-tls-options": 0.080706,
      "explain-prometheus-type-and-auth": 0.1411275,
      "provision-datasource-terraform": 0.0895974,
      "provision-datasources-yaml": 0.19735979999999997
    }
  },
  "notools": {
    "generated": "2026-07-21",
    "job": "anthropic-claude-sonnet-4-6-off-k3",
    "model": "Sonnet 4.6",
    "shots_per_task": 3,
    "total_tasks": 27,
    "tasks_passed": 20,
    "tasks_consistent": 15,
    "pass_rate": 0.7407407407407407,
    "pass_hat_rate": 0.5555555555555556,
    "mean_score": 0.8932098765432098,
    "total_cost": 3.3362526,
    "steps_per_trial": "2.3",
    "task_scores": {
      "add-bigquery": 1.0,
      "add-clickhouse": 0.6,
      "add-infinity-auth": 1.0,
      "add-infinity": 0.9500000000000001,
      "add-influxdb": 1.0,
      "add-loki": 0.8500000000000001,
      "add-mysql": 1.0,
      "add-postgres": 0.6,
      "add-prometheus": 1.0,
      "add-tempo": 0.65,
      "check-datasource-health": 1.0,
      "diagnose-unhealthy-datasource": 1.0,
      "edit-clickhouse-protocol": 1.0,
      "edit-infinity-auth": 1.0,
      "edit-influxdb-database": 1.0,
      "edit-loki-derived-fields": 1.0,
      "edit-postgres-enable-tls": 1.0,
      "edit-prometheus-scrape-interval": 0.8500000000000001,
      "edit-tempo-traces-to-logs": 1.0,
      "explain-bigquery-auth": 0.7,
      "explain-clickhouse-protocol-choice": 1.0,
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
      "add-mysql": true,
      "add-postgres": false,
      "add-prometheus": true,
      "add-tempo": false,
      "check-datasource-health": true,
      "diagnose-unhealthy-datasource": true,
      "edit-clickhouse-protocol": true,
      "edit-infinity-auth": true,
      "edit-influxdb-database": true,
      "edit-loki-derived-fields": true,
      "edit-postgres-enable-tls": true,
      "edit-prometheus-scrape-interval": false,
      "edit-tempo-traces-to-logs": true,
      "explain-bigquery-auth": false,
      "explain-clickhouse-protocol-choice": true,
      "explain-infinity-config": true,
      "explain-influxdb-query-language": true,
      "explain-postgres-tls-options": true,
      "explain-prometheus-type-and-auth": true,
      "provision-datasource-terraform": true,
      "provision-datasources-yaml": true
    },
    "task_consistent": {
      "add-bigquery": false,
      "add-clickhouse": false,
      "add-infinity-auth": true,
      "add-infinity": false,
      "add-influxdb": false,
      "add-loki": false,
      "add-mysql": false,
      "add-postgres": false,
      "add-prometheus": true,
      "add-tempo": false,
      "check-datasource-health": true,
      "diagnose-unhealthy-datasource": true,
      "edit-clickhouse-protocol": true,
      "edit-infinity-auth": true,
      "edit-influxdb-database": true,
      "edit-loki-derived-fields": true,
      "edit-postgres-enable-tls": true,
      "edit-prometheus-scrape-interval": false,
      "edit-tempo-traces-to-logs": false,
      "explain-bigquery-auth": false,
      "explain-clickhouse-protocol-choice": true,
      "explain-infinity-config": true,
      "explain-influxdb-query-language": true,
      "explain-postgres-tls-options": true,
      "explain-prometheus-type-and-auth": false,
      "provision-datasource-terraform": true,
      "provision-datasources-yaml": true
    },
    "task_cost": {
      "add-bigquery": 0.1161636,
      "add-clickhouse": 0.1140216,
      "add-infinity-auth": 0.1275906,
      "add-infinity": 0.23418299999999997,
      "add-influxdb": 0.1958097,
      "add-loki": 0.10284059999999999,
      "add-mysql": 0.1136616,
      "add-postgres": 0.13783289999999998,
      "add-prometheus": 0.16904655,
      "add-tempo": 0.1099356,
      "check-datasource-health": 0.14070375,
      "diagnose-unhealthy-datasource": 0.096843,
      "edit-clickhouse-protocol": 0.1316145,
      "edit-infinity-auth": 0.11950259999999999,
      "edit-influxdb-database": 0.1273845,
      "edit-loki-derived-fields": 0.15490379999999998,
      "edit-postgres-enable-tls": 0.14153549999999998,
      "edit-prometheus-scrape-interval": 0.12572850000000002,
      "edit-tempo-traces-to-logs": 0.1705278,
      "explain-bigquery-auth": 0.13490475000000002,
      "explain-clickhouse-protocol-choice": 0.0382335,
      "explain-infinity-config": 0.084783,
      "explain-influxdb-query-language": 0.05483249999999999,
      "explain-postgres-tls-options": 0.07467599999999999,
      "explain-prometheus-type-and-auth": 0.0654585,
      "provision-datasource-terraform": 0.0873687,
      "provision-datasources-yaml": 0.16616595
    }
  },
  "noschema": {
    "generated": "2026-07-21",
    "job": "anthropic-claude-sonnet-4-6-off-k3",
    "model": "Sonnet 4.6",
    "shots_per_task": 3,
    "total_tasks": 27,
    "tasks_passed": 20,
    "tasks_consistent": 13,
    "pass_rate": 0.7407407407407407,
    "pass_hat_rate": 0.48148148148148145,
    "mean_score": 0.8825,
    "total_cost": 3.3514944,
    "steps_per_trial": "2.3",
    "task_scores": {
      "add-bigquery": 0.8500000000000001,
      "add-clickhouse": 1.0,
      "add-infinity-auth": 1.0,
      "add-infinity": 1.0,
      "add-influxdb": 1.0,
      "add-loki": 1.0,
      "add-mysql": 0.6,
      "add-postgres": 0.6,
      "add-prometheus": 1.0,
      "add-tempo": 0.65,
      "check-datasource-health": 1.0,
      "diagnose-unhealthy-datasource": 1.0,
      "edit-clickhouse-protocol": 1.0,
      "edit-infinity-auth": 1.0,
      "edit-influxdb-database": 1.0,
      "edit-loki-derived-fields": 1.0,
      "edit-postgres-enable-tls": 1.0,
      "edit-prometheus-scrape-interval": 0.8500000000000001,
      "edit-tempo-traces-to-logs": 0.4,
      "explain-bigquery-auth": 0.85,
      "explain-clickhouse-protocol-choice": 1.0,
      "explain-infinity-config": 1.0,
      "explain-influxdb-query-language": 1.0,
      "explain-postgres-tls-options": 1.0,
      "explain-prometheus-type-and-auth": 1.0,
      "provision-datasource-terraform": 1.0,
      "provision-datasources-yaml": 1.0
    },
    "task_passed": {
      "add-bigquery": false,
      "add-clickhouse": true,
      "add-infinity-auth": true,
      "add-infinity": true,
      "add-influxdb": true,
      "add-loki": true,
      "add-mysql": false,
      "add-postgres": false,
      "add-prometheus": true,
      "add-tempo": false,
      "check-datasource-health": true,
      "diagnose-unhealthy-datasource": true,
      "edit-clickhouse-protocol": true,
      "edit-infinity-auth": true,
      "edit-influxdb-database": true,
      "edit-loki-derived-fields": true,
      "edit-postgres-enable-tls": true,
      "edit-prometheus-scrape-interval": false,
      "edit-tempo-traces-to-logs": false,
      "explain-bigquery-auth": false,
      "explain-clickhouse-protocol-choice": true,
      "explain-infinity-config": true,
      "explain-influxdb-query-language": true,
      "explain-postgres-tls-options": true,
      "explain-prometheus-type-and-auth": true,
      "provision-datasource-terraform": true,
      "provision-datasources-yaml": true
    },
    "task_consistent": {
      "add-bigquery": false,
      "add-clickhouse": false,
      "add-infinity-auth": false,
      "add-infinity": false,
      "add-influxdb": false,
      "add-loki": true,
      "add-mysql": false,
      "add-postgres": false,
      "add-prometheus": false,
      "add-tempo": false,
      "check-datasource-health": true,
      "diagnose-unhealthy-datasource": true,
      "edit-clickhouse-protocol": true,
      "edit-infinity-auth": true,
      "edit-influxdb-database": false,
      "edit-loki-derived-fields": true,
      "edit-postgres-enable-tls": true,
      "edit-prometheus-scrape-interval": false,
      "edit-tempo-traces-to-logs": false,
      "explain-bigquery-auth": false,
      "explain-clickhouse-protocol-choice": true,
      "explain-infinity-config": true,
      "explain-influxdb-query-language": true,
      "explain-postgres-tls-options": true,
      "explain-prometheus-type-and-auth": false,
      "provision-datasource-terraform": true,
      "provision-datasources-yaml": true
    },
    "task_cost": {
      "add-bigquery": 0.1563912,
      "add-clickhouse": 0.147723,
      "add-infinity-auth": 0.20672729999999997,
      "add-infinity": 0.26666399999999996,
      "add-influxdb": 0.16977119999999998,
      "add-loki": 0.10922339999999998,
      "add-mysql": 0.11070539999999998,
      "add-postgres": 0.1475058,
      "add-prometheus": 0.1593783,
      "add-tempo": 0.11195639999999998,
      "check-datasource-health": 0.0524904,
      "diagnose-unhealthy-datasource": 0.08753160000000001,
      "edit-clickhouse-protocol": 0.131649,
      "edit-infinity-auth": 0.1086114,
      "edit-influxdb-database": 0.089673,
      "edit-loki-derived-fields": 0.145281,
      "edit-postgres-enable-tls": 0.134877,
      "edit-prometheus-scrape-interval": 0.11938499999999999,
      "edit-tempo-traces-to-logs": 0.1881366,
      "explain-bigquery-auth": 0.1348704,
      "explain-clickhouse-protocol-choice": 0.0395952,
      "explain-infinity-config": 0.07691039999999999,
      "explain-influxdb-query-language": 0.0568992,
      "explain-postgres-tls-options": 0.0976206,
      "explain-prometheus-type-and-auth": 0.0647352,
      "provision-datasource-terraform": 0.11314619999999999,
      "provision-datasources-yaml": 0.1240362
    }
  }
}
-->
