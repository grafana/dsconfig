# o11y-bench Results — Datasource Config

> **Variant: mcp-grafana _without_ the schema-review gate** — `create_datasource` writes directly
> instead of first returning a schema and asking the user for required fields. Compare against
> [RESULTS.md](./RESULTS.md) (gate present).

_Last updated: 2026-07-07 · generated from o11y-bench job `anthropic-claude-sonnet-4-6-off-k3`_

Benchmark of an LLM agent on the `datasource_config` task category (creating, editing, and explaining Grafana datasources via mcp-grafana tools).

📊 [Full HTML report with transcripts](./report_noschema.html)

## Summary

| Model | Tasks | pass^3 (consistent) | pass@3 (any) | Mean score | Cost | Steps/trial |
|---|---|---|---|---|---|---|
| Sonnet 4.6 | 21 | 7/21 (33%) | 12/21 (57%) | 87% | $2.36 | 2.8 |

- **pass^3** — task passes only if *all* 3 attempts pass (strict consistency).
- **pass@3** — task passes if *any* of 3 attempts pass.
- **Mean score** — average per-trial score (0–100%) across all trials.

## Per-task results

| Task | Best score | pass@3 | pass^3 | Cost |
|---|---|---|---|---|
| `add-bigquery` | 100% | ✅ | — | $0.14 |
| `add-clickhouse` | 90% | — | — | $0.12 |
| `add-infinity` | 100% | ✅ | ✅ | $0.17 |
| `add-influxdb` | 100% | ✅ | — | $0.26 |
| `add-loki` | 53% | — | — | $0.11 |
| `add-mysql` | 79% | — | — | $0.08 |
| `add-postgres` | 100% | ✅ | ✅ | $0.10 |
| `add-prometheus` | 88% | — | — | $0.11 |
| `add-tempo` | 82% | — | — | $0.11 |
| `edit-clickhouse-protocol` | 100% | ✅ | ✅ | $0.11 |
| `edit-influxdb-database` | 100% | ✅ | ✅ | $0.10 |
| `edit-loki-derived-fields` | 100% | ✅ | ✅ | $0.12 |
| `edit-postgres-enable-tls` | 100% | ✅ | ✅ | $0.11 |
| `edit-prometheus-scrape-interval` | 100% | ✅ | — | $0.10 |
| `edit-tempo-traces-to-logs` | 100% | ✅ | — | $0.16 |
| `explain-bigquery-auth` | 80% | — | — | $0.14 |
| `explain-clickhouse-protocol-choice` | 85% | — | — | $0.04 |
| `explain-infinity-config` | 100% | ✅ | ✅ | $0.08 |
| `explain-influxdb-query-language` | 65% | — | — | $0.05 |
| `explain-postgres-tls-options` | 85% | — | — | $0.08 |
| `explain-prometheus-type-and-auth` | 100% | ✅ | — | $0.07 |

> Per-task **best score** is the highest of the 3 attempts (matches the HTML report). The summary **mean score** averages every trial.
