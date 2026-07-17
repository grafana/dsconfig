# o11y-bench Results — Datasource Config

_Last updated: 2026-07-17 · generated from o11y-bench job `anthropic-claude-sonnet-4-6-off-k3`_

Benchmark of an LLM agent on the `datasource_config` task category (creating, editing, and explaining Grafana datasources via mcp-grafana tools).

📊 [Full HTML report with transcripts](./report.html)

## Summary

| Model | Tasks | pass^3 (consistent) | pass@3 (any) | Mean score | Cost | Steps/trial |
|---|---|---|---|---|---|---|
| Sonnet 4.6 | 26 | 8/26 (31%) | 16/26 (62%) | 82% | $3.02 | 2.5 |

- **pass^3** — task passes only if *all* 3 attempts pass (strict consistency).
- **pass@3** — task passes if *any* of 3 attempts pass.
- **Mean score** — average per-trial score (0–100%) across all trials.

## Per-task results

| Task | Best score | pass@3 | pass^3 | Cost |
|---|---|---|---|---|
| `add-bigquery` | 100% | ✅ | ✅ | $0.17 |
| `add-clickhouse` | 90% | — | — | $0.16 |
| `add-infinity` | 100% | ✅ | — | $0.13 |
| `add-infinity-auth` | 85% | — | — | $0.10 |
| `add-influxdb` | 100% | ✅ | ✅ | $0.13 |
| `add-loki` | 45% | — | — | $0.17 |
| `add-mysql` | 80% | — | — | $0.15 |
| `add-postgres` | 50% | — | — | $0.08 |
| `add-prometheus` | 90% | — | — | $0.22 |
| `add-tempo` | 25% | — | — | $0.23 |
| `check-datasource-health` | 100% | ✅ | ✅ | $0.05 |
| `diagnose-unhealthy-datasource` | 100% | ✅ | ✅ | $0.08 |
| `edit-clickhouse-protocol` | 100% | ✅ | — | $0.11 |
| `edit-influxdb-database` | 100% | ✅ | — | $0.10 |
| `edit-loki-derived-fields` | 100% | ✅ | ✅ | $0.12 |
| `edit-postgres-enable-tls` | 100% | ✅ | ✅ | $0.11 |
| `edit-prometheus-scrape-interval` | 100% | ✅ | — | $0.10 |
| `edit-tempo-traces-to-logs` | 100% | ✅ | — | $0.16 |
| `explain-bigquery-auth` | 85% | — | — | $0.17 |
| `explain-clickhouse-protocol-choice` | 85% | — | — | $0.04 |
| `explain-infinity-config` | 100% | ✅ | — | $0.07 |
| `explain-influxdb-query-language` | 75% | — | — | $0.05 |
| `explain-postgres-tls-options` | 100% | ✅ | — | $0.09 |
| `explain-prometheus-type-and-auth` | 100% | ✅ | — | $0.06 |
| `provision-datasource-terraform` | 100% | ✅ | ✅ | $0.10 |
| `provision-datasources-yaml` | 100% | ✅ | ✅ | $0.06 |

> Per-task **best score** is the highest of the 3 attempts (matches the HTML report). The summary **mean score** averages every trial.
