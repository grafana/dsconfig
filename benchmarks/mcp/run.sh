#!/usr/bin/env bash
#
# Run the o11y-bench-2.0 datasource benchmark and publish results into this repo.
#
# On start it prompts for which benchmark to run:
#   1) mcp as is  — baseline mcp-grafana; updates RESULTS.md, latest.json, report.html
#   2) no tools   — no-tools mcp-grafana; updates the "no tools" column + report_notools.html
#   3) no schema  — no-schema mcp-grafana; updates the "no schema" column + report_noschema.html
#   4) All        — runs asis, then notools, then noschema (one render each), publishes once
# Set MODE=asis|notools|noschema|all to skip the prompt (e.g. for automation / SKIP_RUN reruns).
#
# NOTE: only "mcp as is" currently works against o11y-bench-2.0. The port from o11y-bench 1.0
# brought over just the datasource task specs (on the jck/dsconfig-spec branch); the no-tools /
# no-schema infrastructure was NOT ported. Those modes need three branches that don't yet exist:
#   o11y-bench-2.0: benchmarking/local-mcp-grafana   (bakes a local mcp-grafana build into Docker)
#   mcp-grafana:    benchmarking/no-tools, benchmarking/no-schema
# and 2.0 has no local mcp-grafana build hook (docker/Dockerfile downloads a published release).
# Until those are recreated, notools/noschema exit with an error and All runs only asis.
# TODO(port): restore notools/noschema for o11y-bench-2.0, then re-enable the guard below.
#
# The asis mode checks out the o11y-bench-2.0 baseline branch before running:
#   mode   o11y-bench-2.0 branch
#   asis   jck/dsconfig-spec   (temporary until the ds-config specs merge to 2.0 main — then "main")
# asis uses the published mcp-grafana (baked into 2.0's Docker image), so it needs no mcp-grafana
# checkout. The repo is left on the asis branch afterward (no restore).
#
# o11y-bench-2.0 refuses new runs when a job already exists under its jobs/ dir, so run.sh wipes
# that dir before each run. Prior results are already captured in RESULTS.md + report*.html. The
# last run's job therefore survives afterward (handy for SKIP_RUN=1 re-renders). Skipped under SKIP_RUN.
#
# Usage:
#   ./run.sh                       # prompt for mode, then run + render, no git
#   MODE=asis ./run.sh             # run the (only currently-working) as-is mode
#   PUBLISH=1 ./run.sh             # also commit + push to the current branch
#   MODEL=anthropic/claude-haiku-4-5-20251001 ./run.sh
#   JOB_NAME=ds-v2 ./run.sh        # name the job dir (use a fresh name after editing task specs,
#                                  # otherwise Harbor's lock rejects the changed task set).
#   SKIP_RUN=1 ./run.sh            # skip the (slow, paid) bench run; just re-render latest job
#
# Requirements: the o11y-bench-2.0 repo checked out as a sibling (or set O11Y_BENCH_DIR); Docker
# running; and model API keys exported (ANTHROPIC_API_KEY + provider key). (mcp-grafana as a
# sibling is only needed once the no-tools / no-schema modes are restored — see TODO above.)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
O11Y_BENCH_DIR="${O11Y_BENCH_DIR:-$(cd "$SCRIPT_DIR/../../.." && pwd)/o11y-bench-2.0}"
MCP_GRAFANA_DIR="${MCP_GRAFANA_DIR:-$(cd "$SCRIPT_DIR/../../.." && pwd)/mcp-grafana}"
MODEL="${MODEL:-anthropic/claude-sonnet-4-6}"
TASKS_PATH="${TASKS_PATH:-tasks-spec/datasource_config}"
N_CONCURRENT="${N_CONCURRENT:-2}"
JOB_NAME="${JOB_NAME:-}"
MODE="${MODE:-}"

# Baseline o11y-bench-2.0 branch for "mcp as is". TEMPORARY until the ds-config specs merge to 2.0
# main — set this to "main" then. (asis uses the published mcp-grafana, so only o11y-bench-2.0 matters.)
ASIS_O11Y_BRANCH="jck/dsconfig-spec"

if [[ ! -d "$O11Y_BENCH_DIR" ]]; then
  echo "o11y-bench-2.0 repo not found at: $O11Y_BENCH_DIR" >&2
  echo "Set O11Y_BENCH_DIR to its path and retry." >&2
  exit 1
fi

# Prompt for the benchmark mode unless MODE is already set.
if [[ -z "$MODE" ]]; then
  echo "Which mcp benchmark do you want to run?"
  PS3="Select an option: "
  select choice in "mcp as is" "no tools" "no schema" "All"; do
    case "$choice" in
      "mcp as is") MODE="asis"; break ;;
      "no tools") MODE="notools"; break ;;
      "no schema") MODE="noschema"; break ;;
      "All") MODE="all"; break ;;
      *) echo "Invalid selection — enter 1, 2, 3, or 4." ;;
    esac
  done
fi

case "$MODE" in
  asis | notools | noschema | all) ;;
  *)
    echo "Unknown MODE '$MODE' (expected: asis, notools, noschema, all)." >&2
    exit 1
    ;;
esac

# no-tools / no-schema aren't ported to o11y-bench-2.0 yet (see header). Fail fast on a direct
# selection; All is reduced to asis-only further down. Once the sibling branches are recreated and
# a local mcp-grafana build hook exists, drop this guard and restore the mcp-grafana upfront check.
# TODO(port): re-enable notools/noschema for o11y-bench-2.0.
if [[ "$MODE" == "notools" || "$MODE" == "noschema" ]]; then
  echo "MODE '$MODE' is not yet supported against o11y-bench-2.0." >&2
  echo "It needs branches that weren't ported from o11y-bench 1.0:" >&2
  echo "  o11y-bench-2.0: benchmarking/local-mcp-grafana" >&2
  echo "  mcp-grafana:    benchmarking/$([[ "$MODE" == notools ]] && echo no-tools || echo no-schema)" >&2
  echo "Recreate those branches (and a local mcp-grafana build hook) first — see run.sh header." >&2
  exit 1
fi

# Check out the sibling-repo branches a mode needs. asis just switches o11y-bench-2.0 (published
# mcp-grafana); notools/noschema also build mcp-grafana from the local sibling, which o11y-bench-2.0's
# local-mcp-grafana branch would bake into its Docker image (once that branch is restored — see header).
prepare_repos() {
  local mode="$1" mcp_branch o11y_branch
  case "$mode" in
    asis)     mcp_branch="";                       o11y_branch="$ASIS_O11Y_BRANCH" ;;
    notools)  mcp_branch="benchmarking/no-tools";  o11y_branch="benchmarking/local-mcp-grafana" ;;
    noschema) mcp_branch="benchmarking/no-schema"; o11y_branch="benchmarking/local-mcp-grafana" ;;
  esac
  if [[ -n "$mcp_branch" ]]; then
    echo "==> $mode: mcp-grafana@$mcp_branch (+ go build ./cmd/mcp-grafana), o11y-bench-2.0@$o11y_branch"
    ( cd "$MCP_GRAFANA_DIR" && git checkout "$mcp_branch" && go build ./cmd/mcp-grafana )
  else
    echo "==> $mode: o11y-bench-2.0@$o11y_branch (published mcp-grafana)"
  fi
  ( cd "$O11Y_BENCH_DIR" && git checkout "$o11y_branch" )
}

# Prepare repos, run the benchmark (unless SKIP_RUN=1), and render this mode's results.
run_one() {
  local mode="$1"
  prepare_repos "$mode"

  # In All mode, give each sub-run its own job dir so the runs don't collide / overwrite.
  local job_name="$JOB_NAME"
  [[ "$RUN_ALL" == "1" && -n "$job_name" ]] && job_name="$JOB_NAME-$mode"

  local job_args=(--model "$MODEL" --path "$TASKS_PATH" --n-concurrent "$N_CONCURRENT")
  [[ -n "$job_name" ]] && job_args+=(--job-name "$job_name")

  if [[ "${SKIP_RUN:-0}" != "1" ]]; then
    # o11y-bench-2.0 refuses to run when any job already exists under jobs/, so wipe it first. Prior
    # results are already captured in RESULTS.md + report*.html by earlier renders. Remove the
    # whole dir (catches hidden files/locks) and recreate it empty. The :? guard makes an empty
    # path impossible (never "rm -rf /jobs").
    echo "==> Clearing o11y-bench-2.0 jobs dir ($O11Y_BENCH_DIR/jobs)"
    rm -rf "${O11Y_BENCH_DIR:?}/jobs"
    mkdir -p "$O11Y_BENCH_DIR/jobs"

    echo "==> Running o11y-bench-2.0 ($MODEL) on $TASKS_PATH${job_name:+ as job '$job_name'} [$mode]"
    ( cd "$O11Y_BENCH_DIR" && mise run bench:job -- "${job_args[@]}" )
  else
    echo "==> SKIP_RUN=1: skipping bench run for $mode, re-rendering latest job"
  fi

  echo "==> Rendering $mode results into $SCRIPT_DIR"
  local render_args=(--o11y-root "$O11Y_BENCH_DIR" --out-dir "$SCRIPT_DIR" --mode "$mode")
  [[ "${SKIP_RUN:-0}" != "1" ]] && render_args+=(--fresh-run)
  [[ -n "$job_name" ]] && render_args+=(--job-dir "$O11Y_BENCH_DIR/jobs/$job_name")
  uv run --project "$O11Y_BENCH_DIR" python "$SCRIPT_DIR/render.py" "${render_args[@]}"
}

if [[ "$MODE" == "all" ]]; then
  # notools/noschema aren't ported to o11y-bench-2.0 yet (see header), so All runs only asis for
  # now. TODO(port): restore `modes_to_run=(asis notools noschema)` and `RUN_ALL=1` once they work.
  echo "==> All: no-tools / no-schema not yet supported on o11y-bench-2.0 — running only 'mcp as is'." >&2
  RUN_ALL=0
  modes_to_run=(asis)
else
  RUN_ALL=0
  modes_to_run=("$MODE")
fi

for m in "${modes_to_run[@]}"; do
  run_one "$m"
done

if [[ "${PUBLISH:-0}" == "1" ]]; then
  echo "==> Publishing (commit + push)"
  cd "$SCRIPT_DIR/.."
  git add benchmarks
  git commit -m "Update o11y-bench-2.0 datasource results ($(date -u +%Y-%m-%d))"
  git push
else
  echo "==> Done. Review benchmarks/ then commit, or re-run with PUBLISH=1 to push."
fi
