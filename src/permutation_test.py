#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
========================================================================
Permutation Test for Machine Learning:
What Null Hypothesis It Tests -- and What It Does Not Prove
========================================================================

Minimal reproducible example. Companion to the CV->0 adjudication
protocol (v2.2), Artifact 1 of the 90-day roadmap.

Run:        python src/permutation_test.py
Deps:       numpy>=1.24, scikit-learn>=1.3
Seed:       fixed (SEED=7). Runtime: a few seconds.

WHAT THIS SHOWS
  Part A. Cross-sectional data with a planted signal.
          The permutation test correctly returns a SMALL p-value.
  Part B. Pure noise, but a wide search procedure (many candidates,
          pick the best). The test correctly returns a LARGE p-value:
          the "discovery" is a selection artifact, not signal.
  Part C. Time series with a shared trend (spurious regression).
          NAIVE shuffling gives a falsely SMALL p-value. This shows
          why the null must preserve the data's dependence structure.

NULL HYPOTHESIS (formal):
  Let T(D) be the score returned by the FULL SEARCH PROCEDURE on data D.
  H0: the distribution of T(X, y) equals the distribution of T(X, pi(y))
      for a uniformly random permutation pi -- i.e. y is exchangeable
      given X, as seen by the procedure.
  p-value = P_H0( T >= T_observed ), estimated by Monte Carlo.

  The object under test is the PROCEDURE (search + selection +
  evaluation), not one fixed model. This matches the protocol's
  system-level null-calibration (protocol v2.2, section 0.7): run the
  whole pipeline on label-shuffled data and count false acceptances.

  The test is VALID only when the permutation scheme preserves everything
  about the data except the X-y association. If the data carry structure
  (time order, groups, trend), naive permutation breaks exchangeability
  and the p-value is meaningless. Part C demonstrates this.

WHAT THIS DOES NOT SHOW  (see "NOT TESTED" at the bottom)
  - It does not prove a specific trained model is overfit.
  - It does not test whether a found relationship is causal or useful.
  - It does not cover every dependence structure (only a shared trend).
========================================================================
"""

import numpy as np
from sklearn.linear_model import LinearRegression
from sklearn.model_selection import train_test_split

SEED = 7
N_PERM = 200          # permutations for the null distribution
TEST_SIZE = 0.25      # holdout fraction
RNG = np.random.default_rng(SEED)


# ---------------------------------------------------------------------
# The object under test: a SEARCH PROCEDURE, not a single model.
# It tries many candidate feature sets, fits OLS on each, keeps the one
# with the best TRAIN R^2, and reports that candidate's HOLDOUT R^2.
# This mimics real practice: try many things, report the best.
# ---------------------------------------------------------------------
def search_procedure(X, y, test_size=TEST_SIZE, random_state=SEED):
    """Holdout R^2 of the best-of-many linear candidates."""
    Xtr, Xho, ytr, yho = train_test_split(
        X, y, test_size=test_size, random_state=random_state
    )
    n_features = X.shape[1]
    # candidate feature sets: every single feature and every pair
    candidates = []
    for i in range(n_features):
        candidates.append([i])
        for j in range(i + 1, n_features):
            candidates.append([i, j])

    best_train_r2 = -np.inf
    best_holdout_r2 = np.nan
    for cols in candidates:
        model = LinearRegression().fit(Xtr[:, cols], ytr)
        tr_r2 = model.score(Xtr[:, cols], ytr)
        if tr_r2 > best_train_r2:
            best_train_r2 = tr_r2
            best_holdout_r2 = model.score(Xho[:, cols], yho)
    return best_holdout_r2


def binom_ci(k, n, z=1.96):
    """Wilson score interval for k successes out of n."""
    if n == 0:
        return 0.0, 1.0
    p = k / n
    denom = 1.0 + z * z / n
    centre = (p + z * z / (2.0 * n)) / denom
    half = (z * np.sqrt(p * (1.0 - p) / n + z * z / (4.0 * n * n))) / denom
    return max(0.0, centre - half), min(1.0, centre + half)


def permutation_test(X, y, n_perm=N_PERM, seed=SEED):
    """
    Monte-Carlo p-value for the observed search-procedure score under
    the label-permutation null. The train/holdout split is held FIXED
    across permutations (random_state fixed), so only label assignment
    varies. Uses the (k+1)/(n_perm+1) correction so p is never exactly
    zero on a finite run.
    """
    observed = search_procedure(X, y)
    null = np.empty(n_perm)
    rng = np.random.default_rng(seed)
    for b in range(n_perm):
        null[b] = search_procedure(X, rng.permutation(y))
    exceed = int(np.sum(null >= observed))
    p = (exceed + 1.0) / (n_perm + 1.0)
    lo, hi = binom_ci(exceed, n_perm)
    return observed, null, p, lo, hi


# ---------------------------------------------------------------------
# Part A: cross-sectional data with a planted linear signal.
# ---------------------------------------------------------------------
def part_a_signal():
    n, f = 120, 6
    X = RNG.normal(size=(n, f))
    # true law depends on features 0 and 1 only
    y = 3.0 * X[:, 0] - 2.0 * X[:, 1] + RNG.normal(scale=0.5, size=n)
    return X, y


# ---------------------------------------------------------------------
# Part B: pure noise; wide search space, so the best-of-many candidate
# can look deceptively good on a single split.
# ---------------------------------------------------------------------
def part_b_noise_many_candidates():
    n, f = 60, 12          # few samples, many features -> wide search
    X = RNG.normal(size=(n, f))
    y = RNG.normal(size=n)  # no signal at all
    return X, y


# ---------------------------------------------------------------------
# Part C: two INDEPENDENT random walks -> classic spurious regression.
# Naive label shuffling breaks the trend and yields a false positive.
# ---------------------------------------------------------------------
def part_c_spurious_trend():
    n = 100
    x = np.cumsum(RNG.normal(size=n))
    y = np.cumsum(RNG.normal(size=n))   # independent of x
    return x.reshape(-1, 1), y


def report(observed, null, p, lo, hi):
    print(f"  observed holdout R^2 = {observed:.3f}")
    print(f"  null mean / max      = {null.mean():.3f} / {null.max():.3f}")
    print(f"  p-value              = {p:.4f}")
    print(f"  95% CI on exceedance = [{lo:.4f}, {hi:.4f}]")


def main():
    print("=" * 70)
    print("PART A: planted signal (expect SMALL p-value)")
    X, y = part_a_signal()
    report(*permutation_test(X, y))

    print("-" * 70)
    print("PART B: pure noise, wide search (expect LARGE p-value)")
    X, y = part_b_noise_many_candidates()
    report(*permutation_test(X, y))
    print("  -> the 'best' model is indistinguishable from chance selection.")

    print("-" * 70)
    print("PART C: spurious trend, NAIVE shuffle (expect falsely SMALL p)")
    X, y = part_c_spurious_trend()
    report(*permutation_test(X, y))
    print("  -> x and y are INDEPENDENT random walks; a small p here is a")
    print("     false positive caused by shuffling destroying the trend.")

    print("=" * 70)
    print("NOT TESTED (limits of this example):")
    print("  - whether any specific deployed model is overfit")
    print("  - causality or business value of a found relationship")
    print("  - dependence structures other than a shared trend")
    print("  - the correct structure-preserving null for Part C")
    print("    (e.g. differencing the series, or a block/stationary bootstrap)")


if __name__ == "__main__":
    main()