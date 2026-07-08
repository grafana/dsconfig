# o11y-bench Results — Datasource Config

_Last updated: 2026-07-07 · generated from o11y-bench job `anthropic-claude-sonnet-4-6-off-k3`_

Benchmark of an LLM agent on the `datasource_config` task category (creating, editing, and explaining Grafana datasources via mcp-grafana tools).

📊 [Full HTML report with transcripts](./report_notools.html)

## Summary

| Model | Tasks | pass^3 (consistent) | pass@3 (any) | Mean score | Cost | Steps/trial |
|---|---|---|---|---|---|---|
| Sonnet 4.6 | 21 | 7/21 (33%) | 13/21 (62%) | 85% | $1.88 | 2.3 |

- **pass^3** — task passes only if *all* 3 attempts pass (strict consistency).
- **pass@3** — task passes if *any* of 3 attempts pass.
- **Mean score** — average per-trial score (0–100%) across all trials.

## Per-task results

| Task | Best score | pass@3 | pass^3 | Cost |
|---|---|---|---|---|
| `add-bigquery` | 89% | — | — | $0.08 |
| `add-clickhouse` | 90% | — | — | $0.07 |
| `add-infinity` | 48% | — | — | $0.07 |
| `add-influxdb` | 100% | ✅ | — | $0.13 |
| `add-loki` | 53% | — | — | $0.10 |
| `add-mysql` | 89% | — | — | $0.06 |
| `add-postgres` | 100% | ✅ | — | $0.13 |
| `add-prometheus` | 100% | ✅ | — | $0.10 |
| `add-tempo` | 82% | — | — | $0.10 |
| `edit-clickhouse-protocol` | 100% | ✅ | ✅ | $0.09 |
| `edit-influxdb-database` | 100% | ✅ | ✅ | $0.09 |
| `edit-loki-derived-fields` | 100% | ✅ | ✅ | $0.11 |
| `edit-postgres-enable-tls` | 100% | ✅ | ✅ | $0.10 |
| `edit-prometheus-scrape-interval` | 100% | ✅ | ✅ | $0.09 |
| `edit-tempo-traces-to-logs` | 100% | ✅ | — | $0.13 |
| `explain-bigquery-auth` | 100% | ✅ | ✅ | $0.14 |
| `explain-clickhouse-protocol-choice` | 85% | — | — | $0.04 |
| `explain-infinity-config` | 100% | ✅ | ✅ | $0.07 |
| `explain-influxdb-query-language` | 65% | — | — | $0.05 |
| `explain-postgres-tls-options` | 100% | ✅ | — | $0.06 |
| `explain-prometheus-type-and-auth` | 100% | ✅ | — | $0.06 |

> Per-task **best score** is the highest of the 3 attempts (matches the HTML report). The summary **mean score** averages every trial.
