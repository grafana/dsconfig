# Data Source Configuration — LLMSpec Scenarios

LLM eval scenarios testing whether the Grafana Assistant correctly **configures data sources**: adding new ones, editing existing settings, and explaining configuration options.

Format follows `grafana/grafana-assistant-app` LLMSpec v2.0 (`tools/llmspec/scenarios/assistant/alerting` as the reference). One scenario per `.convo.yaml`, text + semantic-criteria assertions, fact-backed where the controlled environment supports it.

## Why these 9 sources

Every scenario targets a data source that exists in the LLMSpec controlled environment (`environment/provisioning/datasources/datasources.yaml`). That makes every scenario **fact-groundable** — criteria can be graded against the real provisioned config fetched live at eval time, not just the judge's prior. Sources not in the environment (Elasticsearch, LogicMonitor, GitHub, etc.) were intentionally excluded for this reason.

| Category framing | Env sources |
|---|---|
| Grafana stack | Prometheus, Loki, Tempo |
| SQL-based | Postgres, MySQL, ClickHouse |
| Time-series / OSS | InfluxDB |
| Cloud / Enterprise | BigQuery |
| API / Developer tools | Infinity |

## Scenario matrix (21)

**Add — connect a new source** (`category: behavior`, `facts: /api/datasources`)

| File | Source |
|---|---|
| `add-prometheus.convo.yaml` | Prometheus |
| `add-loki.convo.yaml` | Loki |
| `add-tempo.convo.yaml` | Tempo |
| `add-mysql.convo.yaml` | MySQL |
| `add-postgres.convo.yaml` | Postgres |
| `add-clickhouse.convo.yaml` | ClickHouse |
| `add-bigquery.convo.yaml` | BigQuery |
| `add-influxdb.convo.yaml` | InfluxDB |
| `add-infinity.convo.yaml` | Infinity |

**Edit — change existing settings** (`category: behavior`, `facts: /api/datasources/uid/<uid>` grounds the real before-state)

| File | What it tests |
|---|---|
| `edit-prometheus-scrape-interval.convo.yaml` | Change scrape interval 30s → 15s |
| `edit-loki-derived-fields.convo.yaml` | Add trace_id derived field linking to Tempo |
| `edit-postgres-enable-tls.convo.yaml` | Move sslmode off `disable` to a TLS-enforcing mode |
| `edit-clickhouse-protocol.convo.yaml` | Switch native → HTTP protocol (and port) |
| `edit-influxdb-database.convo.yaml` | Repoint target database |
| `edit-tempo-traces-to-logs.convo.yaml` | Configure trace-to-logs against Loki |

**Explain — recommend / explain options** (`category: learn`, fact on the configured example where useful)

| File | What it tests |
|---|---|
| `explain-prometheus-type-and-auth.convo.yaml` | Type/flavor + auth + access mode |
| `explain-postgres-tls-options.convo.yaml` | sslmode options and when to use each |
| `explain-clickhouse-protocol-choice.convo.yaml` | Native vs HTTP, ports, TLS |
| `explain-influxdb-query-language.convo.yaml` | InfluxQL / Flux / SQL by version |
| `explain-bigquery-auth.convo.yaml` | Service-account JWT vs GCE, required roles |
| `explain-infinity-config.convo.yaml` | Supported formats, allowed hosts, auth |

## Conventions

- **`category`** — limited to LLMSpec's enum (`observe | discover | dashboard | learn | safety | behavior`). Config actions (add/edit) → `behavior`; explanatory → `learn`. There is no "configure" category.
- **`area`** — set to the source name for clean per-source filtering and reporting.
- **`facts`** — `source: grafana` + `path:` resolves a real Grafana API resource at grading time. `/api/datasources` for the inventory; `/api/datasources/uid/<uid>` for a specific source's config.
- **Assertions** — one `semantic` text assertion per scenario, all criteria inside it (per the docs' best practice).

## Running

From `tools/llmspec/` in a checkout of `grafana/grafana-assistant-app`:

```bash
# All datasource-configuration scenarios (once placed under scenarios/assistant/datasources/)
mise run llmspec -- --agent=grafana_assistant_web --scenarios=assistant/datasources

# A single scenario
mise run llmspec -- --agent=grafana_assistant_web --scenarios=assistant/datasources/add-postgres

# Filter by intent category
mise run llmspec -- --agent=grafana_assistant_web --categories=behavior
```

To use upstream, copy these files into `tools/llmspec/scenarios/assistant/datasources/` in the repo.

## Notes & open items

- Scenarios assert the Assistant's **guidance** (text), matching the linked alerting examples — none assert `tool_use`. The agent does not appear to expose a create/update-datasource tool; if one lands, the add/edit scenarios could be upgraded to `tool_use` assertions for tighter grading.
- All input prompts use fictional hosts/projects so the "add" scenarios don't collide with the real provisioned sources (which keep the canonical UIDs the facts resolve against).
- `edit-influxdb-database` assumes the provisioned InfluxDB database is `testdb` (per `datasources.yaml`); the criterion names it explicitly. Update if the fixture changes.
