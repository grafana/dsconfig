#!/usr/bin/env python3
"""Render o11y-bench datasource results into this repo.

Reads a completed o11y-bench job directory and writes, into --out-dir:

  RESULTS.md    human-facing summary (renders natively on GitHub). The Summary
                and Per-task tables have one column per run mode (mcp as is /
                no tools / no schema). It carries a hidden BENCH_DATA JSON block
                so each mode's numbers survive when another mode is re-rendered.
  report*.html  the full o11y-bench HTML report, regenerated with --full-args
                (untruncated tool-call arguments). Each mode writes its own file:
                report.html (asis), report_notools.html, report_noschema.html.
  latest.json   slim structured metrics (diffable across runs) — mcp-as-is only

Every mode updates its own RESULTS.md column and its own report*.html. Only the
"asis" mode additionally writes latest.json.

It reuses o11y-bench's own parsing (reporting.compare_report.load_job) so the
numbers match run_report.html exactly. It is executed with o11y-bench's Python
environment (see run.sh), and adds the o11y-bench repo root to sys.path so the
`reporting` package is importable
"""

import argparse
import datetime
import json
import re
import sys
from pathlib import Path

# (data key, human-facing column label), in display order.
MODES = [("asis", "mcp as is"), ("notools", "no tools"), ("noschema", "no schema")]

# HTML report filename written per mode.
REPORT_NAMES = {
    "asis": "report.html",
    "notools": "report_notools.html",
    "noschema": "report_noschema.html",
}

# Marker for the machine-readable per-mode data embedded in RESULTS.md.
_DATA_RE = re.compile(r"<!-- BENCH_DATA\n(.*?)\n-->", re.S)


def _pct(x: float) -> str:
    return f"{x * 100:.0f}%"


def _read_embedded_data(results_path: Path) -> dict:
    """Recover the per-mode data block previously embedded in RESULTS.md."""
    if not results_path.exists():
        return {}
    match = _DATA_RE.search(results_path.read_text())
    if not match:
        return {}
    try:
        return json.loads(match.group(1))
    except json.JSONDecodeError:
        return {}


def _build_markdown(data: dict) -> str:
    """Render RESULTS.md from the per-mode data map ({mode: summary})."""
    labels = [label for _, label in MODES]
    # Representative shot count for the pass^k / pass@k labels: prefer as-is.
    k = next((data[m]["shots_per_task"] for m, _ in MODES if m in data), 0)

    def cell(mode: str, fn) -> str:
        summary = data.get(mode)
        return fn(summary) if summary else "—"

    lines = ["# o11y-bench Results — Datasource Config", ""]

    provenance = [
        f"{label}: `{data[m]['job']}` ({data[m]['generated']})"
        for m, label in MODES
        if m in data
    ]
    if provenance:
        lines += ["_Generated from o11y-bench jobs — " + "; ".join(provenance) + "._", ""]

    lines += [
        "Benchmark of an LLM agent on the `datasource_config` task category "
        "(creating, editing, and explaining Grafana datasources via mcp-grafana tools).",
        "",
    ]
    report_links = [
        f"[{label}](./{REPORT_NAMES[m]})" for m, label in MODES if m in data
    ]
    if report_links:
        lines += ["📊 Full HTML reports with transcripts — " + " · ".join(report_links), ""]

    header = "| Metric | " + " | ".join(labels) + " |"
    separator = "|" + "---|" * (len(labels) + 1)

    summary_rows = [
        ("Model", lambda s: s["model"]),
        ("Tasks", lambda s: str(s["total_tasks"])),
        (
            f"pass^{k} (consistent)",
            lambda s: f"{s['tasks_consistent']}/{s['total_tasks']} ({_pct(s['pass_hat_rate'])})",
        ),
        (
            f"pass@{k} (any)",
            lambda s: f"{s['tasks_passed']}/{s['total_tasks']} ({_pct(s['pass_rate'])})",
        ),
        ("Mean score", lambda s: _pct(s["mean_score"])),
        ("Cost", lambda s: f"${s['total_cost']:.2f}"),
        ("Steps/trial", lambda s: str(s["steps_per_trial"])),
    ]

    lines += ["## Summary", "", header, separator]
    for name, fn in summary_rows:
        cells = " | ".join(cell(m, fn) for m, _ in MODES)
        lines.append(f"| {name} | {cells} |")
    lines += [
        "",
        f"- **pass^{k}** — task passes only if *all* {k} attempts pass (strict consistency).",
        f"- **pass@{k}** — task passes if *any* of {k} attempts pass.",
        "- **Mean score** — average per-trial score (0–100%) across all trials.",
        "",
    ]

    tasks = sorted(
        {task for m, _ in MODES if m in data for task in data[m]["task_scores"]}
    )
    lines += ["## Per-task best score", "", "| Task | " + " | ".join(labels) + " |", separator]
    for task in tasks:
        def score_cell(mode: str, task: str = task) -> str:
            summary = data.get(mode)
            if not summary or task not in summary["task_scores"]:
                return "—"
            return _pct(summary["task_scores"][task])

        cells = " | ".join(score_cell(m) for m, _ in MODES)
        lines.append(f"| `{task}` | {cells} |")
    lines += [
        "",
        f"> Per-task **best score** is the highest of the {k} attempts for that mode "
        "(matches the HTML report).",
        "",
    ]
    return "\n".join(lines)


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
    parser.add_argument(
        "--mode",
        choices=[m for m, _ in MODES],
        default="asis",
        help="Which run this is. Only 'asis' writes latest.json + report.html; "
        "every mode updates its own column in RESULTS.md (default: asis)",
    )
    parser.add_argument(
        "--fresh-run",
        action="store_true",
        help="This render follows a fresh benchmark run, so the chosen job is authoritative for "
        "--mode. Suppresses the job-reuse warning (o11y-bench auto-names jobs by model/config, so "
        "different modes' fresh jobs legitimately share a name).",
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

    # Merge this mode's numbers into the per-mode data carried by RESULTS.md, then
    # re-render the whole file (the other modes' columns are preserved).
    results_path = out_dir / "RESULTS.md"
    data = _read_embedded_data(results_path)

    # The job is chosen by mtime, so re-rendering a mode without a fresh run for it
    # (e.g. SKIP_RUN=1 after a different mode's run) can pull another mode's job into
    # this column. Warn if the chosen job is already recorded under a different mode —
    # but not after a fresh run, where the job is authoritative for this mode (and modes
    # legitimately share an auto-generated job name).
    clashes = sorted(
        m for m, prev in data.items() if m != args.mode and prev.get("job") == job_dir.name
    )
    if clashes and not args.fresh_run:
        print(
            f"WARNING: job '{job_dir.name}' is already recorded under mode(s) "
            f"[{', '.join(clashes)}]; rendering it as '{args.mode}' attributes the same "
            f"run to multiple modes. If you didn't just run a fresh '{args.mode}' "
            f"benchmark, pass --job-dir (JOB_NAME in run.sh) to select the right job.",
            file=sys.stderr,
        )

    data[args.mode] = summary
    markdown = _build_markdown(data)
    markdown += "\n<!-- BENCH_DATA\n" + json.dumps(data, indent=2) + "\n-->\n"
    results_path.write_text(markdown)

    # Every mode writes its own HTML report (regenerated with untruncated args).
    report_path = out_dir / REPORT_NAMES[args.mode]
    write_report(job_dir, output=report_path, full_args=True)
    wrote = [str(results_path), str(report_path)]

    # Only the as-is run owns latest.json.
    if args.mode == "asis":
        (out_dir / "latest.json").write_text(json.dumps(summary, indent=2))
        wrote.append(str(out_dir / "latest.json"))

    print(f"Job:        {job_dir}")
    print(f"Mode:       {args.mode}")
    for path in wrote:
        print(f"Wrote:      {path}")
    print(
        f"Summary:    pass^{summary['shots_per_task']} "
        f"{summary['tasks_consistent']}/{summary['total_tasks']} · "
        f"mean {_pct(summary['mean_score'])} · ${summary['total_cost']:.2f}"
    )


if __name__ == "__main__":
    main()
