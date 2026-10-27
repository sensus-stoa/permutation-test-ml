"""Оценка частоты ложных срабатываний (Type I error) перестановочного теста.

Под H0 (обе модели эквивалентны) p-value должен быть равномерен на [0, 1],
а доля p < alpha — близка к alpha. Скрипт многократно генерирует данные
под H0, считает p-value и сохраняет результаты в results/.
"""
from __future__ import annotations

import argparse
import csv
import os
import sys

import numpy as np
from sklearn.metrics import accuracy_score

sys.path.insert(0, os.path.join(os.path.dirname(__file__), "..", "src"))
from permutation_test import permutation_test  # noqa: E402


def run_once(rng: np.random.Generator, n: int, n_permutations: int) -> float:
    y = rng.integers(0, 2, size=n)
    err_a = rng.random(n) < 0.2
    err_b = rng.random(n) < 0.2
    a = np.where(err_a, 1 - y, y)
    b = np.where(err_b, 1 - y, y)
    return permutation_test(
        y, a, b,
        metric=accuracy_score,
        n_permutations=n_permutations,
        random_state=int(rng.integers(0, 2**31 - 1)),
    )


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--runs", type=int, default=1000)
    parser.add_argument("--n-samples", type=int, default=200)
    parser.add_argument("--n-permutations", type=int, default=1000)
    parser.add_argument("--alpha", type=float, default=0.05)
    parser.add_argument("--seed", type=int, default=42)
    parser.add_argument("--out", default="results/type1_error.csv")
    args = parser.parse_args()

    os.makedirs(os.path.dirname(args.out), exist_ok=True)
    rng = np.random.default_rng(args.seed)

    rows = []
    for i in range(args.runs):
        p = run_once(rng, args.n_samples, args.n_permutations)
        rows.append({"run": i, "p_value": p, "reject_h0": int(p < args.alpha)})

    with open(args.out, "w", newline="") as f:
        writer = csv.DictWriter(f, fieldnames=["run", "p_value", "reject_h0"])
        writer.writeheader()
        writer.writerows(rows)

    type1 = np.mean([r["reject_h0"] for r in rows])
    print(f"Записано {len(rows)} строк в {args.out}")
    print(f"Type I error rate (alpha={args.alpha}): {type1:.4f}")


if __name__ == "__main__":
    main()
