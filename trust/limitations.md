# Limitations — permutation-test-ml

Public Trust Record. Companion to `trust/claims.md`, `trust/evidence-manifest.json`, `trust/verdict.json`.

## From the article (What I Did Not Test)

- Block permutation or other structure-preserving nulls for Part C (e.g. differencing the series, phase randomization, stationary bootstrap) are not implemented or tested. Part C demonstrates the *limitation* of naive label shuffling; the *remedy* is out of scope for this artifact.
- Only synthetic data from this script. No real production pipelines, no deployed models, no client data were exercised.
- Only linear dependencies captured by `search_procedure` (single-feature and two-feature OLS candidates). Nonlinear or non-additive signals, if present, are invisible to this test statistic.
- `N_PERM = 200`. Minimum achievable p-value ≈ 1/201 ≈ 0.005. Distinguishing p = 0.001 from p = 0.0001 is not possible with this run size.
- Type I error measured at N = 50 per scenario. Wilson 95% CIs are wide by design; the numbers are honest but imprecise.

## Trust Layer service limitations

- **Independent reproduction: NONE yet.** All PASS verdicts in `verdict.json` are author-executed. Self-verification is useful but is *not* labeled independent. An independent third-party run would be recorded as a new `verdict.json` entry with `verifier` = external party.
- **Criteria-lock applies from commit `6169f91…` forward.** The 50-run calibration in the article (published 2026-11-06) was not pre-registered by a commit; formal criteria-lock per v3.5 starts with the criteria-lock commit. Any post-hoc criteria change requires a new run for PASS.
- **`results/` is gitignored.** Evidence `E-05` (`results/repeated_runs.csv`) is reproducible by the documented command but not stored in the repo. Consumers must regenerate.
- **AI assistance** was used in code and text; disclosed in README, Medium article, and `evidence-manifest.json` (provenance_note). This affects *provenance*, not *integrity* (file hashes are still valid) nor *reproducibility* (runs are deterministic given the same seeds and library versions).
- **Test statistic is holdout R² of best-of-many linear candidates.** The permutation null is built around this specific T(·). Claims do not transfer automatically to other test statistics (accuracy, log-likelihood, AUC, etc.) without a new run.

## Scope of applicability

- Exchangeable labels (i.i.d. rows, no time order, no grouping, no shared trend).
- Linear search space as implemented in `search_procedure`.
- Single-machine runs with pinned dependency versions in `requirements.txt`.

Outside this scope the verdict is **INCONCLUSIVE, reason_code: OUT_OF_SCOPE** — not a silent PASS.

## Conditions for revision

This verdict (`V-PTM-001`) is marked `superseded` (not silently overwritten) when any of:

- Code in `src/permutation_test.py` or `experiments/repeated_runs.py` changes (object_version mismatch).
- A claim's PASS/FAIL criteria are changed post-lock without a new run.
- A negative control is re-run with a materially different N or outcome.
- An independent reproduction produces a result that conflicts with `verdict.json` on a PASS claim.

The old verdict is kept in git history; the new one gets a new `verdict_id`. No back-dating.

## Incidents log

See `trust/trust-incidents.md` (created on first incident; absent = no incidents recorded yet).