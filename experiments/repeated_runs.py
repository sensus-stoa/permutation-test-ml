"""
Calibration check of the permutation test over repeated runs.

Three scenarios, same null hypothesis, different data and different
exchangeability conditions:

  noise          H0 is true (X and y are independent).
                 Expected: the fraction of p < alpha is close to alpha.

  signal         H0 is false (y carries a linear dependence on X).
                 Expected: the fraction of p < alpha is far above alpha
                 (power). A sanity check that the test works at all.

  spurious_trend H0 is true (x and y are independent random walks),
                 but naive label shuffling destroys the trend.
                 Expected: the fraction of p < alpha is far above alpha.
                 This demonstrates why the permutation scheme must
                 preserve the dependence structure of the data.

Usage:
    python experiments/repeated_runs.py
    python experiments/repeated_runs.py --runs 500 --n-permutations 200
    python experiments/repeated_runs.py --scenario noise --runs 1000

Output is written to results/repeated_runs.csv, summary to stdout.
"""
from __future__ import annotations

import argparse
import csv
import os
import sys
import time

import numpy as np

# Import from src/permutation_test.py
sys.path.insert(0, os.path.join(os.path.dirname(__file__), "..", "src"))
from permutation_test import (                          # noqa: E402
    part_a_signal,
    part_b_noise_many_candidates,
    part_c_spurious_trend,
    permutation_test,
)


SCENARIOS = {
    "noise":          part_b_noise_many_candidates,
    "signal":         part_a_signal,
    "spurious_trend": part_c_spurious_trend,
}


def run_one(scenario: str, seed: int, n_permutations: int) -> float:
    """One run: generate data under the chosen scenario,
    return the p-value."""
    X, y = _generate(scenario, seed)
    _, _, p, _, _ = permutation_test(X, y, n_perm=n_permutations, seed=seed)
    return float(p)


def _generate(scenario: str, seed: int):
    """Local data generators so that every run sees independent data.
    Same logic as in src/permutation_test.py, but the seed is a
    parameter, not a module-level constant."""
    rng = np.random.default_rng(seed)
    if scenario == "noise":
        n, f = 60, 12
        X = rng.normal(size=(n, f))
        y = rng.normal(size=n)
    elif scenario == "signal":
        n, f = 120, 6
        X = rng.normal(size=(n, f))
        y = 3.0 * X[:, 0] - 2.0 * X[:, 1] + rng.normal(scale=0.5, size=n)
    elif scenario == "spurious_trend":
        n = 100
        x = np.cumsum(rng.normal(size=n))
        y = np.cumsum(rng.normal(size=n))
        X = x.reshape(-1, 1)
    else:
        raise ValueError(f"Unknown scenario: {scenario}")
    return X, y


def binom_ci(k: int, n: int, z: float = 1.96) -> tuple[float, float]:
    """Wilson score interval for a proportion. Matches the one in
    src/permutation_test.py."""
    if n == 0:
        return 0.0, 1.0
    p = k / n
    denom = 1.0 + z * z / n
    centre = (p + z * z / (2.0 * n)) / denom
    half = (z * np.sqrt(p * (1.0 - p) / n + z * z / (4.0 * n * n))) / denom
    return max(0.0, centre - half), min(1.0, centre + half)


def summarize(label: str, pvals: list[float], alpha: float) -> dict:
    n = len(pvals)
    rejects = sum(1 for p in pvals if p < alpha)
    rate = rejects / n if n else 0.0
    lo, hi = binom_ci(rejects, n)
    return {
        "scenario": label,
        "n": n,
        "rejects": rejects,
        "rate": rate,
        "ci_low": lo,
        "ci_high": hi,
    }


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--runs", type=int, default=200,
                        help="runs per scenario (default 200)")
    parser.add_argument("--n-permutations", type=int, default=200,
                        help="permutations inside one test (default 200)")
    parser.add_argument("--alpha", type=float, default=0.05)
    parser.add_argument("--seed0", type=int, default=0,
                        help="base seed; run i uses seed0 + i")
    parser.add_argument("--scenario", choices=list(SCENARIOS) + ["all"],
                        default="all",
                        help="run one scenario or all of them (default all)")
    parser.add_argument("--out", default="results/repeated_runs.csv")
    args = parser.parse_args()

    os.makedirs(os.path.dirname(args.out), exist_ok=True)

    scenarios = (
        list(SCENARIOS.keys()) if args.scenario == "all" else [args.scenario]
    )

    rows = []
    t0 = time.time()
    for sc in scenarios:
        print(f"\n=== {sc} ({args.runs} runs, {args.n_permutations} perms) ===")
        for i in range(args.runs):
            seed = args.seed0 + i
            p = run_one(sc, seed, args.n_permutations)
            rows.append({
                "scenario": sc,
                "run": i,
                "seed": seed,
                "p_value": p,
                "reject_h0": int(p < args.alpha),
            })
            if (i + 1) % max(1, args.runs // 10) == 0:
                rej = sum(r["reject_h0"] for r in rows if r["scenario"] == sc)
                print(f"  {i + 1}/{args.runs}  "
                      f"reject so far: {rej}/{i + 1} = "
                      f"{rej / (i + 1):.3f}")
        elapsed = time.time() - t0
        print(f"  done in {elapsed:.1f}s")

    with open(args.out, "w", newline="") as f:
        w = csv.DictWriter(f, fieldnames=[
            "scenario", "run", "seed", "p_value", "reject_h0"
        ])
        w.writeheader()
        w.writerows(rows)

    print(f"\nWrote {len(rows)} rows to {args.out}")
    print(f"\nSummary (alpha={args.alpha}):")
    print(f"{'scenario':<16} {'rejects':>8} {'rate':>7} {'95% CI':>16}")
    print("-" * 52)
    for sc in scenarios:
        pvals = [r["p_value"] for r in rows if r["scenario"] == sc]
        s = summarize(sc, pvals, args.alpha)
        print(f"{sc:<16} {s['rejects']:>8} "
              f"{s['rate']:>7.3f} "
              f"[{s['ci_low']:.3f}, {s['ci_high']:.3f}]")

    print()
    print("Interpretation:")
    print("  noise           -> rate should be close to alpha (calibration).")
    print("  signal          -> rate >> alpha (the test has power).")
    print("  spurious_trend  -> rate >> alpha (naive shuffling is invalid")
    print("                     on dependent data).")


if __name__ == "__main__":
    main()
