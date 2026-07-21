# Benchmarks — o11y-bench-2.0 + MCP

> **Scope:** This directory covers **o11y-bench-2.0 runs against the mcp-grafana MCP tools only**
> (the `datasource_config` task category). It measures how well an LLM agent configures Grafana
> data sources through the MCP `create_datasource` / `update_datasource` / etc. tools.
>
> The **assistant-flow** benchmark (Grafana Assistant configuring data sources via the
> `manage_datasources` tool) lives under [`../assistant/`](../assistant/). Do not mix the two here.

> **⚠️ Port status:** The suite was ported from o11y-bench 1.0 to **o11y-bench-2.0**. Only the
> **mcp as is** mode currently works — the port brought over just the datasource task specs (on the
> `jck/dsconfig-spec` branch). The **no tools** / **no schema** modes still depend on branches that
> were not ported (`benchmarking/local-mcp-grafana` in o11y-bench-2.0; `benchmarking/no-tools` /
> `benchmarking/no-schema` in mcp-grafana), and 2.0 has no local mcp-grafana build hook. Until those
> are recreated, `run.sh` errors out for those modes and `All` runs only `asis`. Their existing
> `report_notools.html` / `report_noschema.html` and `RESULTS.md` columns are **stale** (from 1.0).

## What's in here

| File | Purpose |
|---|---|
| `run.sh` | Orchestrator: prompts for the run mode, runs the o11y-bench-2.0 benchmark, renders results here, optionally commits + pushes. |
| `render.py` | Parses a completed o11y-bench-2.0 job and writes `RESULTS.md`, the mode's `report*.html`, and (as-is only) `latest.json`. Reuses o11y-bench-2.0's own reporting code so the numbers match its report exactly. |
| `RESULTS.md` | Human-facing summary — renders natively on GitHub. Summary + per-task tables with **one column per mode** (mcp as is / no tools / no schema). **Generated; do not edit by hand.** |
| `latest.json` | Slim structured metrics — diffable across runs. **Written only by the "mcp as is" mode.** |
| `report.html` | Full o11y-bench-2.0 HTML report for the **mcp as is** run (untruncated tool-call args via `--full-args`). **Generated.** |
| `report_notools.html` | Same, for the **no tools** run. **Generated.** |
| `report_noschema.html` | Same, for the **no schema** run. **Generated.** |

## Prerequisites

Shared prerequisites (sibling repo checkouts, Docker, `ANTHROPIC_API_KEY`) live in the
[top-level benchmarks README](../README.md). Specific to this suite:

1. **`o11y-bench-2.0` checked out as a sibling** of this repo (`../../../o11y-bench-2.0` relative to
   this folder). Override with `O11Y_BENCH_DIR=/path/to/o11y-bench-2.0`. **`run.sh` checks out
   branches in this repo** per mode (see [Run modes](#run-modes)), so commit or stash local work
   there first — `git checkout` aborts on conflicting uncommitted changes.
2. **A provider key** for whatever non-Anthropic model you benchmark (e.g. `OPENAI_API_KEY`,
   `GOOGLE_API_KEY`), in addition to `ANTHROPIC_API_KEY` (used by the grader).
3. **`mcp-grafana` checked out as a sibling** (`../../../mcp-grafana`, override with `MCP_GRAFANA_DIR`)
   — **only for the no-tools / no-schema modes, which are not yet ported to o11y-bench-2.0** (see the
   port-status note above). Once those modes are restored they will check out a dedicated mcp-grafana
   branch and build the MCP server locally; o11y-bench-2.0's `benchmarking/local-mcp-grafana` branch
   would then bake that build into its Docker image. The default "mcp as is" mode does not need
   `mcp-grafana` — 2.0 bakes a published mcp-grafana release into its Docker image.

## Usage

Run everything from this directory:

```bash
cd benchmarks/mcp

# Run the benchmark + render results here (no git changes).
# Prompts for the run mode (mcp as is / no tools / no schema / All).
./run.sh

# Skip the prompt by setting MODE. (Only asis works today; notools/noschema error out — see above.)
MODE=asis ./run.sh

# Same, but also commit + push the generated files to the current branch
PUBLISH=1 ./run.sh

# Re-render from the most recent existing job WITHOUT running a new (slow, paid) benchmark
SKIP_RUN=1 ./run.sh
```

## Run modes

On start `run.sh` prompts for which benchmark to run (or set `MODE` to skip the prompt):

| # | Mode | `MODE` | Status | Outputs |
|---|---|---|---|---|
| 1 | mcp as is | `asis` | ✅ | `report.html`, `RESULTS.md` (as-is column), `latest.json` |
| 2 | no tools | `notools` | 🚫 not ported to 2.0 | (was `report_notools.html`, no-tools column) |
| 3 | no schema | `noschema` | 🚫 not ported to 2.0 | (was `report_noschema.html`, no-schema column) |
| 4 | All | `all` | ⚠️ asis only | runs just `asis` for now (notools/noschema skipped), publishes once |

> `notools` / `noschema` exit with an error against o11y-bench-2.0 until their branches are recreated
> (see the port-status note at the top). `all` currently reduces to `asis`.

Each mode checks out the branch(es) it needs in the sibling repos before running:

| Mode | mcp-grafana | o11y-bench-2.0 branch |
|---|---|---|
| `asis` | published (baked into 2.0's Docker image — no local build) | `jck/dsconfig-spec` _(temporary until it merges to 2.0 main)_ |
| `notools` _(not ported)_ | `benchmarking/no-tools` (would build locally via `go build ./cmd/mcp-grafana`) | `benchmarking/local-mcp-grafana` |
| `noschema` _(not ported)_ | `benchmarking/no-schema` (would build locally) | `benchmarking/local-mcp-grafana` |

- **Only the "mcp as is" mode writes `latest.json`.** The other modes (once restored) update just
  their own column in `RESULTS.md` (and their own `report*.html`); untouched columns are preserved
  via a hidden data block embedded in `RESULTS.md`.
- **no tools / no schema** _(not yet ported to 2.0)_ would build mcp-grafana from the local sibling;
  o11y-bench-2.0's `benchmarking/local-mcp-grafana` branch would bake that binary into its Docker
  image. **asis** uses the published mcp-grafana, so it needs no mcp-grafana checkout — only the
  o11y-bench-2.0 baseline branch.
- **All** currently runs only `asis` (notools/noschema aren't ported) and publishes once at the end.
  The repo is left on the asis branch afterward (no restore). Once the other modes are restored, All
  will run all three in sequence and, with `JOB_NAME` set, give each sub-run its own job dir
  (`<JOB_NAME>-asis`, `-notools`, `-noschema`) to avoid collisions.

### Environment variables

| Var | Default | Notes |
|---|---|---|
| `MODE` | _(prompt)_ | `asis` / `notools` / `noschema` / `all` — skips the interactive prompt. **Only `asis` works today**; `notools`/`noschema` error out and `all` reduces to `asis` (see port-status note). |
| `MODEL` | `anthropic/claude-sonnet-4-6` | Model to benchmark (`provider/model`). |
| `JOB_NAME` | _(unset)_ | Names the o11y-bench-2.0 job dir. **Use a fresh name whenever the task specs changed** — otherwise Harbor's lock rejects the changed task set. When set, results render from exactly that job. (Once `all` runs multiple modes again, each sub-run uses `<JOB_NAME>-<mode>`.) |
| `TASKS_PATH` | `tasks-spec/datasource_config` | Which o11y-bench-2.0 task specs to run (relative to the o11y-bench-2.0 repo). |
| `N_CONCURRENT` | `2` | Concurrent trials. |
| `SKIP_RUN` | `0` | `1` = skip the benchmark (and the jobs-dir wipe), just re-render the latest graded job still under `<o11y-bench-2.0>/jobs/`. |
| `PUBLISH` | `0` | `1` = commit + push the generated files after rendering. |
| `O11Y_BENCH_DIR` | `../../../o11y-bench-2.0` | Path to the o11y-bench-2.0 repo. |
| `MCP_GRAFANA_DIR` | `../../../mcp-grafana` | Path to the mcp-grafana repo (no-tools / no-schema modes — not yet ported to 2.0). |

### Typical flow after editing o11y-bench-2.0 task specs

Editing task specs changes their checksums, so you must use a **fresh job name**:

```bash
ANTHROPIC_API_KEY=... JOB_NAME=ds-v2 ./run.sh
```

## How results are produced

`run.sh` does the following:

1. **Select mode** — prompts (or reads `MODE`). `all` currently reduces to just `asis` (notools/
   noschema aren't ported to 2.0); once restored it will expand to `asis`, `notools`, `noschema`,
   each run end-to-end (steps 2–4) in sequence.
2. **Prepare repos** — checks out the branch(es) the mode needs (see [Run modes](#run-modes)):
   `asis` switches only o11y-bench-2.0 (published mcp-grafana). (Restored `notools`/`noschema` would
   also `go build ./cmd/mcp-grafana` from the local sibling.)
3. **Clear jobs** — wipes `<o11y-bench-2.0>/jobs/` (prior results are already captured in `RESULTS.md`
   + `report*.html`). o11y-bench-2.0 refuses to run when a job already exists, so this must happen
   **before** the run. Skipped under `SKIP_RUN=1`.
4. **Run** — invokes `mise run bench:job` in the o11y-bench-2.0 repo for `TASKS_PATH`
   (skipped if `SKIP_RUN=1`).
5. **Render** — runs `render.py` under o11y-bench-2.0's Python environment (`uv run --project`), which
   loads the job via o11y-bench-2.0's `reporting.compare_report.load_job` and writes the mode's column
   into `RESULTS.md`, the mode's `report*.html`, and — for `asis` only — `latest.json`.
6. **Publish** (optional, once) — `git add benchmarks && git commit && git push` when `PUBLISH=1`.

By default `render.py` picks the most-recently-modified graded job under `<o11y-bench-2.0>/jobs/`;
when `JOB_NAME` is set, `run.sh` points it at that specific job dir. The jobs dir is cleared
**before** each run (not after), so the most recent run's job survives afterward — `SKIP_RUN=1`
re-renders it as long as you don't start another run first.

## Viewing the HTML reports

The `report*.html` files are self-contained but GitHub serves raw `.html` as source, not rendered.
To view one:

- **GitHub Pages** (if enabled for this repo): `https://grafana.github.io/dsconfig/benchmarks/mcp/report.html`
- **htmlpreview** (zero setup, public repos): prefix the file's GitHub URL with
  `https://htmlpreview.github.io/?`
- Or download and open locally.

`RESULTS.md` always renders in the GitHub repo browser and is the primary at-a-glance surface.

## Metrics glossary

- **pass^k** — a task passes only if **all** k attempts pass (strict consistency). Headline metric.
- **pass@k** — a task passes if **any** of the k attempts pass.
- **Mean score** — average per-trial score (0–100%) across all trials.
- **Best score** (per-task column) — the highest score across a task's k attempts.
