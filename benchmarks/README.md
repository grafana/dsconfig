# Benchmarks

This directory holds two data-source-config benchmark suites. Each has its own `run.sh` and its
own README covering usage, environment variables, and outputs — **this top-level README only covers
what both suites need.**

| Suite | What it measures | Folder |
|---|---|---|
| **mcp** | An LLM agent configuring Grafana data sources through the raw **mcp-grafana** MCP tools (`create_datasource` / `update_datasource` / …), run via **o11y-bench** on the `datasource_config` task category. | [`mcp/`](./mcp/) |
| **assistant** | **Grafana Assistant** configuring data sources through its own `manage_datasources` tool, run via LLMSpec in **grafana-assistant-app**. | [`assistant/`](./assistant/) |

## Shared prerequisites

Both suites drive an external repo plus a local Docker stack, and both talk to the Anthropic API.

### Sibling repo checkouts

The harnesses expect these repos checked out **as siblings of this repo** (e.g. all under `~/src`).
Each suite's `run.sh` lets you override the path with the listed env var.

| Repo | Needed for | Override env var |
|---|---|---|
| `o11y-bench` | the **mcp** suite (always) | `O11Y_BENCH_DIR` |
| `grafana-assistant-app` | the **assistant** suite (always) | `GA_APP_DIR` |
| `mcp-grafana` | the **mcp** suite's **no-tools / no-schema** modes only | `MCP_GRAFANA_DIR` |

> **`mcp-grafana` is only required for the mcp suite's no-tools / no-schema modes.** Those modes
> check out a dedicated mcp-grafana branch locally and o11y-bench builds the MCP server from that
> sibling checkout. The default "mcp as is" mode does not need it. See [`mcp/README.md`](./mcp/README.md)
> for details.

### Docker

Docker must be running — both suites spin up a Grafana + data-source stack per trial.

### API keys

Export `ANTHROPIC_API_KEY` in your shell. It's used by the grader in both suites (and by the agent
when benchmarking an Anthropic model). The **mcp** suite additionally needs a provider key for
whatever non-Anthropic model you benchmark (e.g. `OPENAI_API_KEY`, `GOOGLE_API_KEY`).

## Running

See the per-suite READMEs for usage, environment variables, and how results are produced:

- [`mcp/README.md`](./mcp/README.md)
- [`assistant/README.md`](./assistant/README.md)
