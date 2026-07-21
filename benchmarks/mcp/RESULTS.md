# o11y-bench-2.0 Results — Datasource Config

_Generated from o11y-bench-2.0 jobs — mcp as is: `anthropic-claude-sonnet-4-6-off-k3` (2026-07-20); no tools: `anthropic-claude-sonnet-4-6-off-k3` (2026-07-21); no schema: `anthropic-claude-sonnet-4-6-off-k3` (2026-07-21)._

Benchmark of an LLM agent on the `datasource_config` task category (creating, editing, and explaining Grafana datasources via mcp-grafana tools).

📊 Full HTML reports with transcripts — [mcp as is](./report.html) · [no tools](./report_notools.html) · [no schema](./report_noschema.html)

## Summary

| Metric | mcp as is | no tools | no schema |
|---|---|---|---|
| Model | Sonnet 4.6 | Sonnet 4.6 | Sonnet 4.6 |
| Tasks | 27 | 27 | 27 |
| pass^3 (consistent) | 13/27 (48%) | 15/27 (56%) | 15/27 (56%) |
| pass@3 (any) | 20/27 (74%) | 21/27 (78%) | 21/27 (78%) |
| Mean score | 87% | 90% | 91% |
| Cost | $3.41 | $2.84 | $2.71 |
| Steps/trial | 2.3 | 2.1 | 2.3 |

- **pass^3** — task passes only if *all* 3 attempts pass (strict consistency).
- **pass@3** — task passes if *any* of 3 attempts pass.
- **Mean score** — average per-trial score (0–100%) across all trials.

## Per-task best score

| Task | mcp as is | no tools | no schema |
|---|---|---|---|
| `add-bigquery` | 100% | 100% | 100% |
| `add-clickhouse` | 90% | 90% | 90% |
| `add-infinity` | 40% | 40% | 100% |
| `add-infinity-auth` | 85% | 100% | 85% |
| `add-influxdb` | 100% | 100% | 100% |
| `add-loki` | 75% | 85% | 100% |
| `add-mysql` | 100% | 60% | 60% |
| `add-postgres` | 60% | 100% | 60% |
| `add-prometheus` | 90% | 95% | 90% |
| `add-tempo` | 65% | 100% | 100% |
| `check-datasource-health` | 100% | 100% | 100% |
| `diagnose-unhealthy-datasource` | 100% | 100% | 100% |
| `edit-clickhouse-protocol` | 100% | 100% | 100% |
| `edit-infinity-auth` | 100% | 100% | 100% |
| `edit-influxdb-database` | 100% | 100% | 100% |
| `edit-loki-derived-fields` | 100% | 100% | 100% |
| `edit-postgres-enable-tls` | 100% | 100% | 100% |
| `edit-prometheus-scrape-interval` | 100% | 100% | 100% |
| `edit-tempo-traces-to-logs` | 100% | 100% | 100% |
| `explain-bigquery-auth` | 100% | 85% | 70% |
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
    "generated": "2026-07-20",
    "job": "anthropic-claude-sonnet-4-6-off-k3",
    "model": "Sonnet 4.6",
    "shots_per_task": 3,
    "total_tasks": 27,
    "tasks_passed": 20,
    "tasks_consistent": 13,
    "pass_rate": 0.7407407407407407,
    "pass_hat_rate": 0.48148148148148145,
    "mean_score": 0.8716049382716049,
    "total_cost": 3.4062945,
    "steps_per_trial": "2.3",
    "task_scores": {
      "add-bigquery": 1.0,
      "add-clickhouse": 0.9,
      "add-infinity-auth": 0.85,
      "add-infinity": 0.4,
      "add-influxdb": 1.0,
      "add-loki": 0.75,
      "add-mysql": 1.0,
      "add-postgres": 0.6,
      "add-prometheus": 0.9,
      "add-tempo": 0.65,
      "check-datasource-health": 1.0,
      "diagnose-unhealthy-datasource": 1.0,
      "edit-clickhouse-protocol": 1.0,
      "edit-infinity-auth": 1.0,
      "edit-influxdb-database": 1.0,
      "edit-loki-derived-fields": 1.0,
      "edit-postgres-enable-tls": 1.0,
      "edit-prometheus-scrape-interval": 1.0,
      "edit-tempo-traces-to-logs": 1.0,
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
      "add-clickhouse": false,
      "add-infinity-auth": false,
      "add-infinity": false,
      "add-influxdb": true,
      "add-loki": false,
      "add-mysql": true,
      "add-postgres": false,
      "add-prometheus": false,
      "add-tempo": false,
      "check-datasource-health": true,
      "diagnose-unhealthy-datasource": true,
      "edit-clickhouse-protocol": true,
      "edit-infinity-auth": true,
      "edit-influxdb-database": true,
      "edit-loki-derived-fields": true,
      "edit-postgres-enable-tls": true,
      "edit-prometheus-scrape-interval": true,
      "edit-tempo-traces-to-logs": true,
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
      "add-clickhouse": false,
      "add-infinity-auth": false,
      "add-infinity": false,
      "add-influxdb": false,
      "add-loki": false,
      "add-mysql": false,
      "add-postgres": false,
      "add-prometheus": false,
      "add-tempo": false,
      "check-datasource-health": true,
      "diagnose-unhealthy-datasource": true,
      "edit-clickhouse-protocol": false,
      "edit-infinity-auth": true,
      "edit-influxdb-database": true,
      "edit-loki-derived-fields": true,
      "edit-postgres-enable-tls": true,
      "edit-prometheus-scrape-interval": true,
      "edit-tempo-traces-to-logs": false,
      "explain-bigquery-auth": false,
      "explain-clickhouse-protocol-choice": true,
      "explain-infinity-config": false,
      "explain-influxdb-query-language": true,
      "explain-postgres-tls-options": true,
      "explain-prometheus-type-and-auth": false,
      "provision-datasource-terraform": true,
      "provision-datasources-yaml": true
    },
    "task_cost": {
      "add-bigquery": 0.2586444,
      "add-clickhouse": 0.1653072,
      "add-infinity-auth": 0.10026059999999999,
      "add-infinity": 0.1669941,
      "add-influxdb": 0.1218474,
      "add-loki": 0.10146060000000001,
      "add-mysql": 0.1008726,
      "add-postgres": 0.1453683,
      "add-prometheus": 0.1688787,
      "add-tempo": 0.1047456,
      "check-datasource-health": 0.05284439999999999,
      "diagnose-unhealthy-datasource": 0.0798732,
      "edit-clickhouse-protocol": 0.16689990000000002,
      "edit-infinity-auth": 0.1439637,
      "edit-influxdb-database": 0.21901500000000002,
      "edit-loki-derived-fields": 0.1807839,
      "edit-postgres-enable-tls": 0.225138,
      "edit-prometheus-scrape-interval": 0.09527279999999999,
      "edit-tempo-traces-to-logs": 0.1383138,
      "explain-bigquery-auth": 0.1870023,
      "explain-clickhouse-protocol-choice": 0.0424422,
      "explain-infinity-config": 0.0756294,
      "explain-influxdb-query-language": 0.0513762,
      "explain-postgres-tls-options": 0.0991392,
      "explain-prometheus-type-and-auth": 0.0609522,
      "provision-datasource-terraform": 0.10306560000000001,
      "provision-datasources-yaml": 0.0502032
    }
  },
  "notools": {
    "generated": "2026-07-21",
    "job": "anthropic-claude-sonnet-4-6-off-k3",
    "model": "Sonnet 4.6",
    "shots_per_task": 3,
    "total_tasks": 27,
    "tasks_passed": 21,
    "tasks_consistent": 15,
    "pass_rate": 0.7777777777777778,
    "pass_hat_rate": 0.5555555555555556,
    "mean_score": 0.9,
    "total_cost": 2.8363845,
    "steps_per_trial": "2.1",
    "task_scores": {
      "add-bigquery": 1.0,
      "add-clickhouse": 0.9,
      "add-infinity-auth": 1.0,
      "add-infinity": 0.4,
      "add-influxdb": 1.0,
      "add-loki": 0.8500000000000001,
      "add-mysql": 0.6,
      "add-postgres": 1.0,
      "add-prometheus": 0.9500000000000001,
      "add-tempo": 1.0,
      "check-datasource-health": 1.0,
      "diagnose-unhealthy-datasource": 1.0,
      "edit-clickhouse-protocol": 1.0,
      "edit-infinity-auth": 1.0,
      "edit-influxdb-database": 1.0,
      "edit-loki-derived-fields": 1.0,
      "edit-postgres-enable-tls": 1.0,
      "edit-prometheus-scrape-interval": 1.0,
      "edit-tempo-traces-to-logs": 1.0,
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
      "add-bigquery": true,
      "add-clickhouse": false,
      "add-infinity-auth": true,
      "add-infinity": false,
      "add-influxdb": true,
      "add-loki": false,
      "add-mysql": false,
      "add-postgres": true,
      "add-prometheus": false,
      "add-tempo": true,
      "check-datasource-health": true,
      "diagnose-unhealthy-datasource": true,
      "edit-clickhouse-protocol": true,
      "edit-infinity-auth": true,
      "edit-influxdb-database": true,
      "edit-loki-derived-fields": true,
      "edit-postgres-enable-tls": true,
      "edit-prometheus-scrape-interval": true,
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
      "add-bigquery": true,
      "add-clickhouse": false,
      "add-infinity-auth": true,
      "add-infinity": false,
      "add-influxdb": false,
      "add-loki": false,
      "add-mysql": false,
      "add-postgres": true,
      "add-prometheus": false,
      "add-tempo": true,
      "check-datasource-health": true,
      "diagnose-unhealthy-datasource": true,
      "edit-clickhouse-protocol": false,
      "edit-infinity-auth": false,
      "edit-influxdb-database": true,
      "edit-loki-derived-fields": false,
      "edit-postgres-enable-tls": true,
      "edit-prometheus-scrape-interval": true,
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
      "add-bigquery": 0.0909981,
      "add-clickhouse": 0.0667743,
      "add-infinity-auth": 0.0790332,
      "add-infinity": 0.0783024,
      "add-influxdb": 0.17710365,
      "add-loki": 0.10565325,
      "add-mysql": 0.0882291,
      "add-postgres": 0.1346088,
      "add-prometheus": 0.0569424,
      "add-tempo": 0.053048399999999996,
      "check-datasource-health": 0.0858921,
      "diagnose-unhealthy-datasource": 0.09853499999999998,
      "edit-clickhouse-protocol": 0.14223585,
      "edit-infinity-auth": 0.12816705,
      "edit-influxdb-database": 0.1029168,
      "edit-loki-derived-fields": 0.17749065,
      "edit-postgres-enable-tls": 0.1077648,
      "edit-prometheus-scrape-interval": 0.09746579999999999,
      "edit-tempo-traces-to-logs": 0.21856845,
      "explain-bigquery-auth": 0.11286135,
      "explain-clickhouse-protocol-choice": 0.09596355000000001,
      "explain-infinity-config": 0.0643944,
      "explain-influxdb-query-language": 0.055496699999999996,
      "explain-postgres-tls-options": 0.14805405,
      "explain-prometheus-type-and-auth": 0.0634377,
      "provision-datasource-terraform": 0.14960294999999998,
      "provision-datasources-yaml": 0.0568437
    }
  },
  "noschema": {
    "generated": "2026-07-21",
    "job": "anthropic-claude-sonnet-4-6-off-k3",
    "model": "Sonnet 4.6",
    "shots_per_task": 3,
    "total_tasks": 27,
    "tasks_passed": 21,
    "tasks_consistent": 15,
    "pass_rate": 0.7777777777777778,
    "pass_hat_rate": 0.5555555555555556,
    "mean_score": 0.9111111111111111,
    "total_cost": 2.7068946,
    "steps_per_trial": "2.3",
    "task_scores": {
      "add-bigquery": 1.0,
      "add-clickhouse": 0.9,
      "add-infinity-auth": 0.85,
      "add-infinity": 1.0,
      "add-influxdb": 1.0,
      "add-loki": 1.0,
      "add-mysql": 0.6,
      "add-postgres": 0.6,
      "add-prometheus": 0.9,
      "add-tempo": 1.0,
      "check-datasource-health": 1.0,
      "diagnose-unhealthy-datasource": 1.0,
      "edit-clickhouse-protocol": 1.0,
      "edit-infinity-auth": 1.0,
      "edit-influxdb-database": 1.0,
      "edit-loki-derived-fields": 1.0,
      "edit-postgres-enable-tls": 1.0,
      "edit-prometheus-scrape-interval": 1.0,
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
      "add-infinity-auth": false,
      "add-infinity": true,
      "add-influxdb": true,
      "add-loki": true,
      "add-mysql": false,
      "add-postgres": false,
      "add-prometheus": false,
      "add-tempo": true,
      "check-datasource-health": true,
      "diagnose-unhealthy-datasource": true,
      "edit-clickhouse-protocol": true,
      "edit-infinity-auth": true,
      "edit-influxdb-database": true,
      "edit-loki-derived-fields": true,
      "edit-postgres-enable-tls": true,
      "edit-prometheus-scrape-interval": true,
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
      "add-bigquery": true,
      "add-clickhouse": false,
      "add-infinity-auth": false,
      "add-infinity": false,
      "add-influxdb": false,
      "add-loki": true,
      "add-mysql": false,
      "add-postgres": false,
      "add-prometheus": false,
      "add-tempo": true,
      "check-datasource-health": true,
      "diagnose-unhealthy-datasource": true,
      "edit-clickhouse-protocol": false,
      "edit-infinity-auth": true,
      "edit-influxdb-database": false,
      "edit-loki-derived-fields": true,
      "edit-postgres-enable-tls": true,
      "edit-prometheus-scrape-interval": true,
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
      "add-bigquery": 0.1407942,
      "add-clickhouse": 0.12813239999999998,
      "add-infinity-auth": 0.0890952,
      "add-infinity": 0.1370844,
      "add-influxdb": 0.23790360000000002,
      "add-loki": 0.060406800000000004,
      "add-mysql": 0.08812619999999999,
      "add-postgres": 0.0988092,
      "add-prometheus": 0.11549999999999999,
      "add-tempo": 0.0595218,
      "check-datasource-health": 0.0524088,
      "diagnose-unhealthy-datasource": 0.0798024,
      "edit-clickhouse-protocol": 0.10931160000000001,
      "edit-infinity-auth": 0.0839262,
      "edit-influxdb-database": 0.1015296,
      "edit-loki-derived-fields": 0.12235860000000001,
      "edit-postgres-enable-tls": 0.1073106,
      "edit-prometheus-scrape-interval": 0.09513060000000001,
      "edit-tempo-traces-to-logs": 0.1513074,
      "explain-bigquery-auth": 0.1574532,
      "explain-clickhouse-protocol-choice": 0.040949400000000004,
      "explain-infinity-config": 0.0727488,
      "explain-influxdb-query-language": 0.0539034,
      "explain-postgres-tls-options": 0.1068642,
      "explain-prometheus-type-and-auth": 0.061244400000000004,
      "provision-datasource-terraform": 0.1007862,
      "provision-datasources-yaml": 0.0544854
    }
  }
}
-->
