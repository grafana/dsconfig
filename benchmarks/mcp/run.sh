#!/usr/bin/env bash
#
# Run the o11y-bench datasource benchmark and publish results into this repo.
#
# On start it prompts for which benchmark to run:
#   1) mcp as is  — baseline mcp-grafana; updates RESULTS.md, latest.json, report.html
#   2) no tools   — no-tools mcp-grafana; updates the "no tools" column + report_notools.html
#   3) no schema  — no-schema mcp-grafana; updates the "no schema" column + report_noschema.html
#   4) All        — runs asis, then notools, then noschema (one render each), publishes once
# Set MODE=asis|notools|noschema|all to skip the prompt (e.g. for automation / SKIP_RUN reruns).
#
# Each mode checks out the branch(es) it needs in the sibling repos before running:
#   mode      mcp-grafana                     o11y-bench branch
#   asis      (published, via o11y-bench)     jck/ds-config-spec   (temporary; see below)
#   notools   benchmarking/no-tools (local)   benchmarking/local-mcp-grafana
#   noschema  benchmarking/no-schema (local)  benchmarking/local-mcp-grafana
# asis uses the published mcp-grafana (o11y-bench pulls it), so it needs no mcp-grafana checkout.
# notools/noschema build mcp-grafana from the local sibling, which o11y-bench's local-mcp-grafana
# branch bakes into its Docker image. Repos are left on the last mode's branches (no restore).
#
# Usage:
#   ./run.sh                       # prompt for mode, then run + render, no git
#   MODE=all ./run.sh              # run every mode
#   PUBLISH=1 ./run.sh             # also commit + push to the current branch
#   MODEL=anthropic/claude-haiku-4-5-20251001 ./run.sh
#   JOB_NAME=ds-v2 ./run.sh        # name the job dir (use a fresh name after editing task specs,
#                                  # otherwise Harbor's lock rejects the changed task set).
#                                  # In All mode each sub-run uses "<JOB_NAME>-<mode>".
#   SKIP_RUN=1 ./run.sh            # skip the (slow, paid) bench run; just re-render latest job
#
# Requirements: the o11y-bench repo checked out as a sibling (or set O11Y_BENCH_DIR); the
# mcp-grafana repo too (or set MCP_GRAFANA_DIR) for the no-tools / no-schema modes; Docker
# running; and model API keys exported (ANTHROPIC_API_KEY + provider key).

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
O11Y_BENCH_DIR="${O11Y_BENCH_DIR:-$(cd "$SCRIPT_DIR/../../.." && pwd)/o11y-bench}"
MCP_GRAFANA_DIR="${MCP_GRAFANA_DIR:-$(cd "$SCRIPT_DIR/../../.." && pwd)/mcp-grafana}"
MODEL="${MODEL:-anthropic/claude-sonnet-4-6}"
TASKS_PATH="${TASKS_PATH:-tasks-spec/datasource_config}"
N_CONCURRENT="${N_CONCURRENT:-2}"
JOB_NAME="${JOB_NAME:-}"
MODE="${MODE:-}"

# Baseline o11y-bench branch for "mcp as is". TEMPORARY until the ds-config-spec changes merge to
# main — set this to "main" then. (asis uses the published mcp-grafana, so only o11y-bench matters.)
ASIS_O11Y_BRANCH="jck/ds-config-spec"

if [[ ! -d "$O11Y_BENCH_DIR" ]]; then
  echo "o11y-bench repo not found at: $O11Y_BENCH_DIR" >&2
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

# mcp-grafana is only needed for the local-build modes (notools / noschema, and all, which runs
# them). asis uses the published mcp-grafana via o11y-bench. Check upfront so an All run doesn't
# fail partway through.
if [[ "$MODE" != "asis" && ! -d "$MCP_GRAFANA_DIR" ]]; then
  echo "mcp-grafana repo not found at: $MCP_GRAFANA_DIR" >&2
  echo "Set MCP_GRAFANA_DIR to its path and retry (needed for the no-tools / no-schema modes)." >&2
  exit 1
fi

# Check out the sibling-repo branches a mode needs. asis just switches o11y-bench (published
# mcp-grafana); notools/noschema also build mcp-grafana from the local sibling, which o11y-bench's
# local-mcp-grafana branch bakes into its Docker image.
prepare_repos() {
  local mode="$1" mcp_branch o11y_branch
  case "$mode" in
    asis)     mcp_branch="";                       o11y_branch="$ASIS_O11Y_BRANCH" ;;
    notools)  mcp_branch="benchmarking/no-tools";  o11y_branch="benchmarking/local-mcp-grafana" ;;
    noschema) mcp_branch="benchmarking/no-schema"; o11y_branch="benchmarking/local-mcp-grafana" ;;
  esac
  if [[ -n "$mcp_branch" ]]; then
    echo "==> $mode: mcp-grafana@$mcp_branch (+ go build ./cmd/mcp-grafana), o11y-bench@$o11y_branch"
    ( cd "$MCP_GRAFANA_DIR" && git checkout "$mcp_branch" && go build ./cmd/mcp-grafana )
  else
    echo "==> $mode: o11y-bench@$o11y_branch (published mcp-grafana)"
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
    echo "==> Running o11y-bench ($MODEL) on $TASKS_PATH${job_name:+ as job '$job_name'} [$mode]"
    ( cd "$O11Y_BENCH_DIR" && mise run bench:job -- "${job_args[@]}" )
  else
    echo "==> SKIP_RUN=1: skipping bench run for $mode, re-rendering latest job"
  fi

  echo "==> Rendering $mode results into $SCRIPT_DIR"
  local render_args=(--o11y-root "$O11Y_BENCH_DIR" --out-dir "$SCRIPT_DIR" --mode "$mode")
  [[ -n "$job_name" ]] && render_args+=(--job-dir "$O11Y_BENCH_DIR/jobs/$job_name")
  uv run --project "$O11Y_BENCH_DIR" python "$SCRIPT_DIR/render.py" "${render_args[@]}"
}

if [[ "$MODE" == "all" ]]; then
  RUN_ALL=1
  modes_to_run=(asis notools noschema)
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
  git commit -m "Update o11y-bench datasource results ($(date -u +%Y-%m-%d))"
  git push
else
  echo "==> Done. Review benchmarks/ then commit, or re-run with PUBLISH=1 to push."
fi
