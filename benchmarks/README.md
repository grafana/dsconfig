# Benchmarks — o11y-bench + MCP

> **Scope:** This directory covers **o11y-bench runs against the mcp-grafana MCP tools only**
> (the `datasource_config` task category). It measures how well an LLM agent configures Grafana
> data sources through the MCP `create_datasource` / `update_datasource` / etc. tools.
>
> The **assistant-flow** benchmark (Grafana Assistant configuring data sources via the
> `manage_datasources` tool) lives under [`assistant/`](./assistant/). Do not mix the two here.

## What's in here

| File | Purpose |
|---|---|
| `run.sh` | Orchestrator: runs the o11y-bench benchmark, renders results here, optionally commits + pushes. |
| `render.py` | Parses a completed o11y-bench job and writes `RESULTS.md`, `latest.json`, and `report.html`. Reuses o11y-bench's own reporting code so the numbers match its report exactly. |
| `RESULTS.md` | Human-facing summary — renders natively on GitHub. Pass rates + per-task table. **Generated; do not edit by hand.** |
| `latest.json` | Slim structured metrics for the latest run — diffable across runs. **Generated.** |
| `report.html` | The full o11y-bench HTML report (with untruncated tool-call arguments via `--full-args`). **Generated.** |

## Prerequisites

1. **The o11y-bench repo checked out as a sibling** of this repo (i.e. `../../o11y-bench` relative
   to this folder). Override with `O11Y_BENCH_DIR=/path/to/o11y-bench` if it lives elsewhere.
2. **Docker running** — o11y-bench spins up a Grafana + Prometheus + Loki + Tempo + mcp-grafana
   stack per trial.
3. **API keys exported** in your shell:
   - `ANTHROPIC_API_KEY` — used by o11y-bench's grader (and by the agent when benchmarking an
     Anthropic model).
   - A provider key for whichever model you benchmark (e.g. `OPENAI_API_KEY`, `GOOGLE_API_KEY`).

## Usage

Run everything from this directory:

```bash
cd benchmarks

# Run the benchmark + render results here (no git changes)
./run.sh

# Same, but also commit + push the generated files to the current branch
PUBLISH=1 ./run.sh

# Re-render from the most recent existing job WITHOUT running a new (slow, paid) benchmark
SKIP_RUN=1 ./run.sh
```

### Environment variables

| Var | Default | Notes |
|---|---|---|
| `MODEL` | `anthropic/claude-sonnet-4-6` | Model to benchmark (`provider/model`). |
| `JOB_NAME` | _(unset)_ | Names the o11y-bench job dir. **Use a fresh name whenever the task specs changed** — otherwise Harbor's lock rejects the changed task set. When set, results render from exactly that job. |
| `TASKS_PATH` | `tasks-spec/datasource_config` | Which o11y-bench task specs to run (relative to the o11y-bench repo). |
| `N_CONCURRENT` | `2` | Concurrent trials. |
| `SKIP_RUN` | `0` | `1` = skip the benchmark, just re-render the latest graded job. |
| `PUBLISH` | `0` | `1` = commit + push the generated files after rendering. |
| `O11Y_BENCH_DIR` | `../../o11y-bench` | Path to the o11y-bench repo. |

### Typical flow after editing o11y-bench task specs

Editing task specs changes their checksums, so you must use a **fresh job name**:

```bash
ANTHROPIC_API_KEY=... JOB_NAME=ds-v2 ./run.sh
```

## How results are produced

`run.sh` does three things:

1. **Run** — invokes `mise run bench:job` in the o11y-bench repo for `TASKS_PATH`
   (skipped if `SKIP_RUN=1`).
2. **Render** — runs `render.py` under o11y-bench's Python environment (`uv run --project`), which
   loads the job via o11y-bench's `reporting.compare_report.load_job` and writes `RESULTS.md`,
   `latest.json`, and `report.html` here.
3. **Publish** (optional) — `git add benchmarks && git commit && git push` when `PUBLISH=1`.

By default `render.py` picks the most-recently-modified graded job under `<o11y-bench>/jobs/`;
when `JOB_NAME` is set, `run.sh` points it at that specific job dir.

## Viewing `report.html`

`report.html` is self-contained but GitHub serves raw `.html` as source, not rendered. To view it:

- **GitHub Pages** (if enabled for this repo): `https://grafana.github.io/dsconfig/benchmarks/report.html`
- **htmlpreview** (zero setup, public repos): prefix the file's GitHub URL with
  `https://htmlpreview.github.io/?`
- Or download and open locally.

`RESULTS.md` always renders in the GitHub repo browser and is the primary at-a-glance surface.

## Metrics glossary

- **pass^k** — a task passes only if **all** k attempts pass (strict consistency). Headline metric.
- **pass@k** — a task passes if **any** of the k attempts pass.
- **Mean score** — average per-trial score (0–100%) across all trials.
- **Best score** (per-task column) — the highest score across a task's k attempts.
