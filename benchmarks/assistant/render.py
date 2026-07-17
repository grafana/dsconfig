#!/usr/bin/env python3
"""Render an LLMSpec job into this repo's slim `latest.json`.

The mcp/ harness gets this schema for free from o11y-bench's own `load_job`.
LLMSpec has no equivalent, so we recompute the same diffable summary from a
job's per-execution result.json + verifier/assertions.json files.

Reads a completed LLMSpec job directory (default: the latest completed job
under <ga-app>/tools/llmspec/jobs) and writes, into --out-dir:

  latest.json   slim structured metrics, matching benchmarks/mcp/latest.json

Per-task aggregation mirrors the mcp report:
  - task score  = best (max) mean-assertion-score across the k shots
  - pass@k      = any shot passed        (task_passed)
  - pass^k      = all shots passed       (task_consistent)
  - mean score  = average per-trial score across every shot
"""

import argparse
import datetime
import glob
import json
import os
from collections import defaultdict
from pathlib import Path

# Pretty display names for the model ids LLMSpec records; falls back to a
# generic "claude-" stripper for anything not listed here.
_MODEL_NAMES = {
    "claude-sonnet-4-6": "Sonnet 4.6",
    "claude-opus-4-8": "Opus 4.8",
    "claude-haiku-4-5-20251001": "Haiku 4.5",
}


def _model_display(model_id: str) -> str:
    if model_id in _MODEL_NAMES:
        return _MODEL_NAMES[model_id]
    # e.g. "claude-sonnet-4-6" -> "Sonnet 4.6"
    parts = model_id.replace("claude-", "").split("-")
    if len(parts) >= 2:
        family = parts[0].capitalize()
        version = ".".join(p for p in parts[1:] if p.isdigit())
        return f"{family} {version}".strip()
    return model_id


def _exec_score(exec_dir: str, passed: bool) -> float:
    """Mean assertion score for an execution (0-1).

    If the run was never graded (no assertions.json), fall back to the binary
    pass/fail so the number is still meaningful.
    """
    path = os.path.join(exec_dir, "verifier", "assertions.json")
    if not os.path.exists(path):
        return 1.0 if passed else 0.0
    assertions = json.load(open(path))
    if not assertions:
        return 1.0 if passed else 0.0
    return sum(a.get("score", 0) for a in assertions) / len(assertions)


def _latest_job_dir(jobs_dir: Path) -> Path:
    """Most-recently-modified job dir with a completed top-level result.json."""
    candidates = []
    for d in jobs_dir.iterdir():
        if not d.is_dir():
            continue
        result = d / "result.json"
        if not result.exists():
            continue
        try:
            status = json.load(open(result)).get("status")
        except (json.JSONDecodeError, OSError):
            continue
        if status == "completed":
            candidates.append(d)
    if not candidates:
        raise SystemExit(f"No completed job directories found under {jobs_dir}")
    return max(candidates, key=lambda d: d.stat().st_mtime)


def build_summary(job_dir: Path) -> dict:
    job = json.load(open(job_dir / "result.json"))
    shots = job.get("repetitions") or 1

    scores = defaultdict(list)   # task -> [per-shot score]
    passed = defaultdict(list)   # task -> [per-shot passed?]
    costs = defaultdict(float)   # task -> summed cost
    steps = []                   # per-shot tool-call counts (for steps/trial)

    exec_dirs = [
        d
        for d in glob.glob(str(job_dir / "*__*"))
        if os.path.isdir(d) and os.path.exists(os.path.join(d, "result.json"))
    ]
    for d in exec_dirs:
        r = json.load(open(os.path.join(d, "result.json")))
        task = r["scenarioId"].split("/")[-1]
        is_pass = r.get("status") == "passed"
        scores[task].append(_exec_score(d, is_pass))
        passed[task].append(is_pass)
        costs[task] += r.get("totalCost", 0.0) or 0.0
        steps.append(r.get("totalToolCalls", 0) or 0)

    tasks = sorted(scores)
    task_scores = {t: max(scores[t]) for t in tasks}
    task_passed = {t: any(passed[t]) for t in tasks}
    task_consistent = {t: all(passed[t]) for t in tasks}
    task_cost = {t: costs[t] for t in tasks}

    total_tasks = len(tasks)
    n_passed = sum(task_passed.values())
    n_consistent = sum(task_consistent.values())
    all_scores = [s for t in tasks for s in scores[t]]
    mean_score = sum(all_scores) / len(all_scores) if all_scores else 0.0
    steps_per_trial = f"{(sum(steps) / len(steps)):.1f}" if steps else "0.0"

    return {
        "generated": datetime.date.today().isoformat(),
        "job": job_dir.name,
        "model": _model_display(job.get("model", "")),
        "shots_per_task": shots,
        "total_tasks": total_tasks,
        "tasks_passed": n_passed,
        "tasks_consistent": n_consistent,
        "pass_rate": (n_passed / total_tasks) if total_tasks else 0.0,
        "pass_hat_rate": (n_consistent / total_tasks) if total_tasks else 0.0,
        "mean_score": mean_score,
        "total_cost": job.get("totalCost", sum(task_cost.values())),
        "steps_per_trial": steps_per_trial,
        "task_scores": task_scores,
        "task_passed": task_passed,
        "task_consistent": task_consistent,
        "task_cost": task_cost,
    }


def main() -> None:
    parser = argparse.ArgumentParser(description="Render an LLMSpec job into latest.json")
    parser.add_argument(
        "--jobs-dir",
        type=Path,
        required=True,
        help="Path to <ga-app>/tools/llmspec/jobs",
    )
    parser.add_argument("--out-dir", type=Path, required=True, help="Where to write latest.json")
    parser.add_argument(
        "--job-dir",
        type=Path,
        default=None,
        help="Specific job dir (default: latest completed job under --jobs-dir)",
    )
    args = parser.parse_args()

    job_dir = args.job_dir.resolve() if args.job_dir else _latest_job_dir(args.jobs_dir.resolve())
    out_dir = args.out_dir.resolve()
    out_dir.mkdir(parents=True, exist_ok=True)

    summary = build_summary(job_dir)
    (out_dir / "latest.json").write_text(json.dumps(summary, indent=2) + "\n")

    print(f"Job:      {job_dir}")
    print(f"Wrote:    {out_dir / 'latest.json'}")
    print(
        f"Summary:  pass^{summary['shots_per_task']} "
        f"{summary['tasks_consistent']}/{summary['total_tasks']} · "
        f"pass@{summary['shots_per_task']} {summary['tasks_passed']}/{summary['total_tasks']} · "
        f"mean {summary['mean_score'] * 100:.0f}% · ${summary['total_cost']:.2f}"
    )


if __name__ == "__main__":
    main()
