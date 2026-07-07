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
# Usage:
#   ./run.sh                       # run + render, no git
#   PUBLISH=1 ./run.sh             # also commit + push to the current branch
#   MODEL=anthropic/claude-haiku-4-5-20251001 ./run.sh
#   SKIP_RUN=1 ./run.sh            # skip the (slow, paid) bench run; just re-render latest job
#
# Requirements: the o11y-bench repo checked out as a sibling (or set O11Y_BENCH_DIR),
# Docker running, and model API keys exported (ANTHROPIC_API_KEY + provider key).

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
O11Y_BENCH_DIR="${O11Y_BENCH_DIR:-$(cd "$SCRIPT_DIR/../.." && pwd)/o11y-bench}"
MODEL="${MODEL:-anthropic/claude-sonnet-4-6}"
TASKS_PATH="${TASKS_PATH:-tasks-spec/datasource_config}"
N_CONCURRENT="${N_CONCURRENT:-2}"

if [[ ! -d "$O11Y_BENCH_DIR" ]]; then
  echo "o11y-bench repo not found at: $O11Y_BENCH_DIR" >&2
  echo "Set O11Y_BENCH_DIR to its path and retry." >&2
  exit 1
fi

if [[ "${SKIP_RUN:-0}" != "1" ]]; then
  echo "==> Running o11y-bench ($MODEL) on $TASKS_PATH"
  ( cd "$O11Y_BENCH_DIR" && mise run bench:job -- \
      --model "$MODEL" --path "$TASKS_PATH" --n-concurrent "$N_CONCURRENT" )
else
  echo "==> SKIP_RUN=1: skipping bench run, re-rendering latest job"
fi

echo "==> Rendering results into $SCRIPT_DIR"
uv run --project "$O11Y_BENCH_DIR" python "$SCRIPT_DIR/render.py" \
  --o11y-root "$O11Y_BENCH_DIR" --out-dir "$SCRIPT_DIR"

if [[ "${PUBLISH:-0}" == "1" ]]; then
  echo "==> Publishing (commit + push)"
  cd "$SCRIPT_DIR/.."
  git add benchmarks
  git commit -m "Update o11y-bench datasource results ($(date -u +%Y-%m-%d))"
  git push
else
  echo "==> Done. Review benchmarks/ then commit, or re-run with PUBLISH=1 to push."
fi
