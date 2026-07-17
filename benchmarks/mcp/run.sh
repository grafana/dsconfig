#!/usr/bin/env bash
#
# Run the o11y-bench datasource benchmark and publish results into this repo.
#
# Steps:
#   1. Run o11y-bench on the datasource_config task category.
#   2. Render RESULTS.md + latest.json + report.html into this directory
#      (reusing o11y-bench's own parsing, so numbers match its report).
#   3. Optionally commit + push (PUBLISH=1).
#
# On start it prompts for which benchmark to run:
#   1) mcp as is  — the current setup; updates RESULTS.md, latest.json, report.html
#   2) no tools   — build the no-tools mcp-grafana; updates only the "no tools" column of RESULTS.md
#   3) no schema  — (not implemented yet)
#   4) All        — (not implemented yet)
# Set MODE=asis|notools to skip the prompt (e.g. for automation / SKIP_RUN reruns).
#
# Usage:
#   ./run.sh                       # prompt for mode, then run + render, no git
#   PUBLISH=1 ./run.sh             # also commit + push to the current branch
#   MODEL=anthropic/claude-haiku-4-5-20251001 ./run.sh
#   JOB_NAME=ds-v2 ./run.sh        # name the job dir (use a fresh name after editing task specs,
#                                  # otherwise Harbor's lock rejects the changed task set)
#   SKIP_RUN=1 ./run.sh            # skip the (slow, paid) bench run; just re-render latest job
#   MODE=notools ./run.sh          # pick the no-tools mode non-interactively
#
# Requirements: the o11y-bench repo checked out as a sibling (or set O11Y_BENCH_DIR),
# the mcp-grafana repo checked out as a sibling (or set MCP_GRAFANA_DIR) for the no-tools mode,
# Docker running, and model API keys exported (ANTHROPIC_API_KEY + provider key).

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
O11Y_BENCH_DIR="${O11Y_BENCH_DIR:-$(cd "$SCRIPT_DIR/../../.." && pwd)/o11y-bench}"
MCP_GRAFANA_DIR="${MCP_GRAFANA_DIR:-$(cd "$SCRIPT_DIR/../../.." && pwd)/mcp-grafana}"
MODEL="${MODEL:-anthropic/claude-sonnet-4-6}"
TASKS_PATH="${TASKS_PATH:-tasks-spec/datasource_config}"
N_CONCURRENT="${N_CONCURRENT:-2}"
JOB_NAME="${JOB_NAME:-}"
MODE="${MODE:-}"

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
  asis | notools) ;;
  noschema | all)
    echo "Mode '$MODE' is not implemented yet." >&2
    exit 1
    ;;
  *)
    echo "Unknown MODE '$MODE' (expected: asis, notools, noschema, all)." >&2
    exit 1
    ;;
esac

# no-tools: point the local checkouts at the no-tools mcp-grafana build. o11y-bench's
# benchmarking/local-mcp-grafana branch builds mcp-grafana from this sibling checkout into its
# Docker build context, so checking out the branches here is enough — the usual run below picks
# up the no-tools build.
if [[ "$MODE" == "notools" ]]; then
  echo "==> no-tools: preparing local repos for a no-tools run"
  if [[ ! -d "$MCP_GRAFANA_DIR" ]]; then
    echo "mcp-grafana repo not found at: $MCP_GRAFANA_DIR" >&2
    echo "Set MCP_GRAFANA_DIR to its path and retry." >&2
    exit 1
  fi
  echo "==> mcp-grafana: checkout benchmarking/no-tools && go build ./cmd/mcp-grafana"
  ( cd "$MCP_GRAFANA_DIR" && git checkout benchmarking/no-tools && go build ./cmd/mcp-grafana )
  echo "==> o11y-bench: checkout benchmarking/local-mcp-grafana"
  ( cd "$O11Y_BENCH_DIR" && git checkout benchmarking/local-mcp-grafana )
fi

JOB_ARGS=(--model "$MODEL" --path "$TASKS_PATH" --n-concurrent "$N_CONCURRENT")
[[ -n "$JOB_NAME" ]] && JOB_ARGS+=(--job-name "$JOB_NAME")

if [[ "${SKIP_RUN:-0}" != "1" ]]; then
  echo "==> Running o11y-bench ($MODEL) on $TASKS_PATH${JOB_NAME:+ as job '$JOB_NAME'}"
  ( cd "$O11Y_BENCH_DIR" && mise run bench:job -- "${JOB_ARGS[@]}" )
else
  echo "==> SKIP_RUN=1: skipping bench run, re-rendering latest job"
fi

echo "==> Rendering results into $SCRIPT_DIR"
RENDER_ARGS=(--o11y-root "$O11Y_BENCH_DIR" --out-dir "$SCRIPT_DIR" --mode "$MODE")
[[ -n "$JOB_NAME" ]] && RENDER_ARGS+=(--job-dir "$O11Y_BENCH_DIR/jobs/$JOB_NAME")
uv run --project "$O11Y_BENCH_DIR" python "$SCRIPT_DIR/render.py" "${RENDER_ARGS[@]}"

if [[ "${PUBLISH:-0}" == "1" ]]; then
  echo "==> Publishing (commit + push)"
  cd "$SCRIPT_DIR/.."
  git add benchmarks
  git commit -m "Update o11y-bench datasource results ($(date -u +%Y-%m-%d))"
  git push
else
  echo "==> Done. Review benchmarks/ then commit, or re-run with PUBLISH=1 to push."
fi
