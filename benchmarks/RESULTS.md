# o11y-bench Results — Datasource Config

_Last updated: 2026-07-06 · generated from o11y-bench job `anthropic-claude-sonnet-4-6-off-k3`_

Benchmark of an LLM agent on the `datasource_config` task category (creating, editing, and explaining Grafana datasources via mcp-grafana tools).

📊 [Full HTML report with transcripts](./report.html)

## Summary

| Model | Tasks | pass^3 (consistent) | pass@3 (any) | Mean score | Cost | Steps/trial |
|---|---|---|---|---|---|---|
| Sonnet 4.6 | 21 | 4/21 (19%) | 11/21 (52%) | 78% | $2.59 | 2.8 |

- **pass^3** — task passes only if *all* 3 attempts pass (strict consistency).
- **pass@3** — task passes if *any* of 3 attempts pass.
- **Mean score** — average per-trial score (0–100%) across all trials.

## Per-task results

| Task | Best score | pass@3 | pass^3 | Cost |
|---|---|---|---|---|
| `add-bigquery` | 68% | — | — | $0.11 |
| `add-clickhouse` | 82% | — | — | $0.17 |
| `add-infinity` | 100% | ✅ | — | $0.15 |
| `add-influxdb` | 91% | — | — | $0.14 |
| `add-loki` | 53% | — | — | $0.17 |
| `add-mysql` | 68% | — | — | $0.11 |
| `add-postgres` | 100% | ✅ | — | $0.13 |
| `add-prometheus` | 88% | — | — | $0.22 |
| `add-tempo` | 53% | — | — | $0.18 |
| `edit-clickhouse-protocol` | 100% | ✅ | ✅ | $0.11 |
| `edit-influxdb-database` | 100% | ✅ | ✅ | $0.10 |
| `edit-loki-derived-fields` | 100% | ✅ | ✅ | $0.12 |
| `edit-postgres-enable-tls` | 100% | ✅ | ✅ | $0.11 |
| `edit-prometheus-scrape-interval` | 100% | ✅ | — | $0.10 |
| `edit-tempo-traces-to-logs` | 100% | ✅ | — | $0.14 |
| `explain-bigquery-auth` | 100% | ✅ | — | $0.18 |
| `explain-clickhouse-protocol-choice` | 85% | — | — | $0.04 |
| `explain-infinity-config` | 100% | ✅ | — | $0.07 |
| `explain-influxdb-query-language` | 65% | — | — | $0.06 |
| `explain-postgres-tls-options` | 85% | — | — | $0.12 |
| `explain-prometheus-type-and-auth` | 100% | ✅ | — | $0.07 |

> Per-task **best score** is the highest of the 3 attempts (matches the HTML report). The summary **mean score** averages every trial.
