#!/usr/bin/env python3
"""Render o11y-bench datasource results into this repo.

Reads a completed o11y-bench job directory and writes, into --out-dir:

  RESULTS.md    human-facing summary (renders natively on GitHub)
  latest.json   slim structured metrics (diffable across runs)
  report.html   the full o11y-bench HTML report, regenerated with --full-args
                (untruncated tool-call arguments)

It reuses o11y-bench's own parsing (reporting.compare_report.load_job) so the
numbers match run_report.html exactly. It is executed with o11y-bench's Python
environment (see run.sh), and adds the o11y-bench repo root to sys.path so the
`reporting` package is importable.
"""

import argparse
import datetime
import json
import sys
from pathlib import Path


def _pct(x: float) -> str:
    return f"{x * 100:.0f}%"


def _yn(flag: bool) -> str:
    return "✅" if flag else "—"


def _latest_job_dir(jobs_dir: Path) -> Path:
    """Most-recently-modified job dir that actually contains graded trials."""
    candidates = [
        d
        for d in jobs_dir.iterdir()
        if d.is_dir() and any(d.glob("*/verifier/grading_details.json"))
    ]
    if not candidates:
        raise SystemExit(f"No graded job directories found under {jobs_dir}")
    return max(candidates, key=lambda d: d.stat().st_mtime)


def _build_markdown(summary: dict, report_link: str | None) -> str:
    k = summary["shots_per_task"]
    tasks = sorted(summary["task_scores"])

    lines = [
        "# o11y-bench Results — Datasource Config",
        "",
        f"_Last updated: {summary['generated']} · generated from o11y-bench job "
        f"`{summary['job']}`_",
        "",
        "Benchmark of an LLM agent on the `datasource_config` task category "
        "(creating, editing, and explaining Grafana datasources via mcp-grafana tools).",
        "",
    ]
    if report_link:
        lines += [f"📊 [Full HTML report with transcripts]({report_link})", ""]

    lines += [
        "## Summary",
        "",
        f"| Model | Tasks | pass^{k} (consistent) | pass@{k} (any) | Mean score | Cost | Steps/trial |",
        "|---|---|---|---|---|---|---|",
        "| {model} | {total} | {cons}/{total} ({consp}) | {pass}/{total} ({passp}) | "
        "{mean} | ${cost:.2f} | {steps} |".format(
            model=summary["model"],
            total=summary["total_tasks"],
            cons=summary["tasks_consistent"],
            consp=_pct(summary["pass_hat_rate"]),
            **{"pass": summary["tasks_passed"]},
            passp=_pct(summary["pass_rate"]),
            mean=_pct(summary["mean_score"]),
            cost=summary["total_cost"],
            steps=summary["steps_per_trial"],
        ),
        "",
        f"- **pass^{k}** — task passes only if *all* {k} attempts pass (strict consistency).",
        f"- **pass@{k}** — task passes if *any* of {k} attempts pass.",
        "- **Mean score** — average per-trial score (0–100%) across all trials.",
        "",
        "## Per-task results",
        "",
        f"| Task | Best score | pass@{k} | pass^{k} | Cost |",
        "|---|---|---|---|---|",
    ]
    for task in tasks:
        lines.append(
            "| `{t}` | {best} | {atk} | {hatk} | ${cost:.2f} |".format(
                t=task,
                best=_pct(summary["task_scores"][task]),
                atk=_yn(summary["task_passed"][task]),
                hatk=_yn(summary["task_consistent"][task]),
                cost=summary["task_cost"].get(task, 0.0),
            )
        )
    lines += [
        "",
        "> Per-task **best score** is the highest of the "
        f"{k} attempts (matches the HTML report). The summary **mean score** "
        "averages every trial.",
        "",
    ]
    return "\n".join(lines)


def main() -> None:
    parser = argparse.ArgumentParser(description="Render o11y-bench results into this repo")
    parser.add_argument("--o11y-root", type=Path, required=True, help="Path to the o11y-bench repo")
    parser.add_argument("--out-dir", type=Path, required=True, help="Where to write results files")
    parser.add_argument(
        "--job-dir",
        type=Path,
        default=None,
        help="Specific job dir (default: latest graded job under <o11y-root>/jobs)",
    )
    args = parser.parse_args()

    o11y_root = args.o11y_root.resolve()
    sys.path.insert(0, str(o11y_root))
    try:
        from reporting.compare_report import load_job
        from reporting.run_report import write_report
    except ImportError as exc:  # pragma: no cover
        raise SystemExit(
            f"Could not import o11y-bench reporting modules from {o11y_root}: {exc}"
        ) from exc

    job_dir = args.job_dir.resolve() if args.job_dir else _latest_job_dir(o11y_root / "jobs")
    out_dir = args.out_dir.resolve()
    out_dir.mkdir(parents=True, exist_ok=True)

    job = load_job(job_dir, tasks_dir=None)
    summary = {
        "generated": datetime.date.today().isoformat(),
        "job": job_dir.name,
        "model": job["model_display"],
        "shots_per_task": job["shots_per_task"],
        "total_tasks": job["total_tasks"],
        "tasks_passed": job["tasks_passed"],
        "tasks_consistent": job["tasks_consistent"],
        "pass_rate": job["pass_rate"],
        "pass_hat_rate": job["pass_hat_rate"],
        "mean_score": job["mean_score"],
        "total_cost": job["total_cost"],
        "steps_per_trial": job["steps_per_trial"],
        "task_scores": job["task_scores"],
        "task_passed": job["task_passed"],
        "task_consistent": job["task_consistent"],
        "task_cost": job["task_cost"],
    }

    # Regenerate the HTML report with untruncated tool-call arguments (--full-args).
    report_path = out_dir / "report.html"
    write_report(job_dir, output=report_path, full_args=True)

    (out_dir / "latest.json").write_text(json.dumps(summary, indent=2))
    (out_dir / "RESULTS.md").write_text(_build_markdown(summary, report_link="./report.html"))

    print(f"Job:        {job_dir}")
    print(f"Wrote:      {out_dir / 'RESULTS.md'}")
    print(f"Wrote:      {out_dir / 'latest.json'}")
    print(f"Wrote:      {report_path}")
    print(
        f"Summary:    pass^{summary['shots_per_task']} "
        f"{summary['tasks_consistent']}/{summary['total_tasks']} · "
        f"mean {_pct(summary['mean_score'])} · ${summary['total_cost']:.2f}"
    )


if __name__ == "__main__":
    main()
