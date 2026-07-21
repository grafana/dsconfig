# Benchmarks — o11y-bench-2.0 + MCP

> **Scope:** This directory covers **o11y-bench-2.0 runs against the mcp-grafana MCP tools only**
> (the `datasource_config` task category). It measures how well an LLM agent configures Grafana
> data sources through the MCP `create_datasource` / `update_datasource` / etc. tools.
>
> The **assistant-flow** benchmark (Grafana Assistant configuring data sources via the
> `manage_datasources` tool) lives under [`../assistant/`](../assistant/). Do not mix the two here.

> **Note:** This suite was ported from o11y-bench 1.0 to **o11y-bench-2.0**. All three modes run
> against 2.0. The `asis` baseline lives on the `jck/dsconfig-spec` branch (temporary, until it
> merges to 2.0 `main`); `notools`/`noschema` use the `benchmarking/local-mcp-grafana` branch in
> o11y-bench-2.0 with the `benchmarking/no-tools` / `benchmarking/no-schema` branches in mcp-grafana.
> In 2.0, `mcp-grafana` is **built by o11y-bench-2.0's preflight** from its sibling checkout (see
> [Run modes](#run-modes)) — `run.sh` only checks out the branch, it no longer runs `go build`.

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
3. **`mcp-grafana` checked out as o11y-bench-2.0's sibling** (`../../../mcp-grafana`, override with
   `MCP_GRAFANA_DIR`) — **only required for the no-tools / no-schema modes.** Those modes check out a
   dedicated mcp-grafana branch (`benchmarking/no-tools` / `benchmarking/no-schema`); o11y-bench-2.0's
   `benchmarking/local-mcp-grafana` branch then **builds that sibling checkout in its preflight**
   (`go build`, needs a **Go toolchain**) and bakes the binary into the sidecar Docker image.
   `MCP_GRAFANA_DIR` must be the directory o11y-bench-2.0's preflight builds from (its `../mcp-grafana`
   sibling — the default). The default "mcp as is" mode does not need `mcp-grafana` — 2.0 downloads a
   pinned published release into its Docker image.

## Usage

Run everything from this directory:

```bash
cd benchmarks/mcp

# Run the benchmark + render results here (no git changes).
# Prompts for the run mode (mcp as is / no tools / no schema / All).
./run.sh

# Skip the prompt by setting MODE.
MODE=notools ./run.sh

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
| 2 | no tools | `notools` | ✅ | `report_notools.html`, `RESULTS.md` (no-tools column) |
| 3 | no schema | `noschema` | ✅ | `report_noschema.html`, `RESULTS.md` (no-schema column) |
| 4 | All | `all` | ✅ | runs `asis` → `notools` → `noschema` (all outputs above), publishes once |

Each mode checks out the branch(es) it needs in the sibling repos before running:

| Mode | mcp-grafana | o11y-bench-2.0 branch |
|---|---|---|
| `asis` | published (2.0's Docker image downloads a pinned release — no local build) | `jck/dsconfig-spec` _(temporary until it merges to 2.0 main)_ |
| `notools` | `benchmarking/no-tools` (built by o11y-bench-2.0's preflight) | `benchmarking/local-mcp-grafana` |
| `noschema` | `benchmarking/no-schema` (built by o11y-bench-2.0's preflight) | `benchmarking/local-mcp-grafana` |

- **Only the "mcp as is" mode writes `latest.json`.** The other modes update just their own column
  in `RESULTS.md` (and their own `report*.html`); the untouched columns are preserved via a hidden
  data block embedded in `RESULTS.md`.
- **no tools / no schema** check out an mcp-grafana branch in the sibling; the build happens on
  o11y-bench-2.0's side — its `benchmarking/local-mcp-grafana` branch has a preflight step (run
  automatically by `bench:job`) that `go build`s the `../mcp-grafana` sibling and bakes the binary
  into the sidecar Docker image. `run.sh` no longer runs `go build` itself. **asis** uses the
  published mcp-grafana, so it needs no mcp-grafana checkout — only the o11y-bench-2.0 baseline branch.
- **All** runs the three modes in sequence (one render each) and publishes once at the end. The
  repos are left on the last mode's branches afterward (no restore). With `JOB_NAME` set, each
  sub-run gets its own job dir (`<JOB_NAME>-asis`, `-notools`, `-noschema`) to avoid collisions.

### Environment variables

| Var | Default | Notes |
|---|---|---|
| `MODE` | _(prompt)_ | `asis` / `notools` / `noschema` / `all` — skips the interactive prompt. |
| `MODEL` | `anthropic/claude-sonnet-4-6` | Model to benchmark (`provider/model`). |
| `JOB_NAME` | _(unset)_ | Names the o11y-bench-2.0 job dir. **Use a fresh name whenever the task specs changed** — otherwise Harbor's lock rejects the changed task set. When set, results render from exactly that job. In `all` mode each sub-run uses `<JOB_NAME>-<mode>`. |
| `TASKS_PATH` | `tasks-spec/datasource_config` | Which o11y-bench-2.0 task specs to run (relative to the o11y-bench-2.0 repo). |
| `N_CONCURRENT` | `2` | Concurrent trials. |
| `SKIP_RUN` | `0` | `1` = skip the benchmark (and the jobs-dir wipe), just re-render the latest graded job still under `<o11y-bench-2.0>/jobs/`. |
| `PUBLISH` | `0` | `1` = commit + push the generated files after rendering. |
| `O11Y_BENCH_DIR` | `../../../o11y-bench-2.0` | Path to the o11y-bench-2.0 repo. |
| `MCP_GRAFANA_DIR` | `../../../mcp-grafana` | Path to the mcp-grafana repo, as o11y-bench-2.0's sibling (no-tools / no-schema modes; preflight builds it). |

### Typical flow after editing o11y-bench-2.0 task specs

Editing task specs changes their checksums, so you must use a **fresh job name**:

```bash
ANTHROPIC_API_KEY=... JOB_NAME=ds-v2 ./run.sh
```

## How results are produced

`run.sh` does the following:

1. **Select mode** — prompts (or reads `MODE`). `all` expands to `asis`, `notools`, `noschema`,
   each run end-to-end (steps 2–4) in sequence.
2. **Prepare repos** — checks out the branch(es) the mode needs (see [Run modes](#run-modes)):
   `asis` switches only o11y-bench-2.0 (published mcp-grafana); `notools`/`noschema` also switch
   mcp-grafana to the matching branch (o11y-bench-2.0's preflight then builds it — run.sh doesn't).
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
