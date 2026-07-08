# o11y-bench Results — Datasource Config

> **Variant: mcp-grafana _with_ the schema-review gate** — `create_datasource` first returns a
> schema and instructs the agent to ask the user for required fields before writing. Compare against
> [RESULTS_NOSCHEMA.md](./RESULTS_NOSCHEMA.md) (gate removed).

_Last updated: 2026-07-07 · generated from o11y-bench job `anthropic-claude-sonnet-4-6-off-k3`_

Benchmark of an LLM agent on the `datasource_config` task category (creating, editing, and explaining Grafana datasources via mcp-grafana tools).

📊 [Full HTML report with transcripts](./report.html)

## Summary

| Model | Tasks | pass^3 (consistent) | pass@3 (any) | Mean score | Cost | Steps/trial |
|---|---|---|---|---|---|---|
| Sonnet 4.6 | 21 | 5/21 (24%) | 9/21 (43%) | 80% | $2.62 | 2.7 |

- **pass^3** — task passes only if *all* 3 attempts pass (strict consistency).
- **pass@3** — task passes if *any* of 3 attempts pass.
- **Mean score** — average per-trial score (0–100%) across all trials.

## Per-task results

| Task | Best score | pass@3 | pass^3 | Cost |
|---|---|---|---|---|
| `add-bigquery` | 89% | — | — | $0.14 |
| `add-clickhouse` | 90% | — | — | $0.17 |
| `add-infinity` | 86% | — | — | $0.14 |
| `add-influxdb` | 100% | ✅ | ✅ | $0.14 |
| `add-loki` | 53% | — | — | $0.17 |
| `add-mysql` | 79% | — | — | $0.09 |
| `add-postgres` | 89% | — | — | $0.10 |
| `add-prometheus` | 88% | — | — | $0.30 |
| `add-tempo` | 82% | — | — | $0.18 |
| `edit-clickhouse-protocol` | 100% | ✅ | ✅ | $0.11 |
| `edit-influxdb-database` | 100% | ✅ | ✅ | $0.10 |
| `edit-loki-derived-fields` | 100% | ✅ | ✅ | $0.12 |
| `edit-postgres-enable-tls` | 100% | ✅ | ✅ | $0.11 |
| `edit-prometheus-scrape-interval` | 100% | ✅ | — | $0.06 |
| `edit-tempo-traces-to-logs` | 65% | — | — | $0.15 |
| `explain-bigquery-auth` | 100% | ✅ | — | $0.22 |
| `explain-clickhouse-protocol-choice` | 85% | — | — | $0.04 |
| `explain-infinity-config` | 100% | ✅ | — | $0.09 |
| `explain-influxdb-query-language` | 65% | — | — | $0.05 |
| `explain-postgres-tls-options` | 85% | — | — | $0.08 |
| `explain-prometheus-type-and-auth` | 100% | ✅ | — | $0.06 |

> Per-task **best score** is the highest of the 3 attempts (matches the HTML report). The summary **mean score** averages every trial.
