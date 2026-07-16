#!/usr/bin/env bash
#
# Run the Grafana Assistant datasource-config benchmark and publish results here.
#
# Unlike the mcp/ harness (which post-processes an o11y-bench job), LLMSpec emits
# the self-contained HTML report itself, so this is essentially one `mise run`:
#
#   mise run llmspec -- --agent=grafana_assistant \
#     --scenarios=assistant/datasource-config --shots=3 --report=<here>/report.html
#
# Steps:
#   1. Bring up the local LLMSpec datasource stack (skip with SKIP_ENV=1).
#   2. Run the datasource-config scenarios and write report.html + latest.json here.
#   3. Optionally commit + push (PUBLISH=1).
#
# Usage:
#   ./run.sh                       # env up + run, no git
#   PUBLISH=1 ./run.sh             # also commit + push to the current branch
#   SKIP_ENV=1 ./run.sh            # reuse an already-running stack
#   SHOTS=1 ./run.sh               # single-shot instead of pass^3
#   MODEL=claude-haiku-4-5-20251001 ./run.sh
#
# Requirements: the grafana-assistant-app repo checked out as a sibling (or set
# GA_APP_DIR), Docker running, and ANTHROPIC_API_KEY exported.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GA_APP_DIR="${GA_APP_DIR:-$(cd "$SCRIPT_DIR/../../.." && pwd)/grafana-assistant-app}"
AGENT="${AGENT:-grafana_assistant}"
SCENARIOS="${SCENARIOS:-assistant/datasource-config}"
SHOTS="${SHOTS:-3}"

if [[ ! -d "$GA_APP_DIR" ]]; then
  echo "grafana-assistant-app repo not found at: $GA_APP_DIR" >&2
  echo "Set GA_APP_DIR to its path and retry." >&2
  exit 1
fi

if [[ "${SKIP_ENV:-0}" != "1" ]]; then
  echo "==> Bringing up the local LLMSpec datasource stack"
  echo "==> Setting directory as $GA_APP_DIR"
  mise run -C "$GA_APP_DIR" llmspec:env:up
fi

echo "==> Running $AGENT on $SCENARIOS (shots=$SHOTS)"
LLMSPEC_ARGS=(--agent="$AGENT" --scenarios="$SCENARIOS" --shots="$SHOTS" --report="$SCRIPT_DIR/report.html")
[[ -n "${MODEL:-}" ]] && LLMSPEC_ARGS+=(--agent-model="$MODEL")
mise run -C "$GA_APP_DIR" llmspec -- "${LLMSPEC_ARGS[@]}"

# Copy the just-written job's structured result alongside the HTML report.
LATEST_JOB="$(ls -dt "$GA_APP_DIR"/tools/llmspec/jobs/*/ 2>/dev/null | head -1 || true)"
if [[ -n "$LATEST_JOB" && -f "${LATEST_JOB}result.json" ]]; then
  cp "${LATEST_JOB}result.json" "$SCRIPT_DIR/latest.json"
  echo "==> Wrote $SCRIPT_DIR/latest.json (job $(basename "$LATEST_JOB"))"
fi

if [[ "${PUBLISH:-0}" == "1" ]]; then
  echo "==> Publishing (commit + push)"
  cd "$SCRIPT_DIR/.."
  git add assistant
  git commit -m "Update Grafana Assistant datasource results ($(date -u +%Y-%m-%d))"
  git push
else
  echo "==> Done. Review benchmarks/assistant/ then commit, or re-run with PUBLISH=1 to push."
fi
