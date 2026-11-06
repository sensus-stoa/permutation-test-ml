# Permutation testing for machine learning:
# which null hypothesis it tests, and what it does not prove

Evgeniy Dolgov, independent researcher, CV->0 program.

**Disclosure:** AI tools were used in preparing the code and text.

## Abstract

The permutation test is often presented as an answer to "is there
signal in the data". That is imprecise. The test checks a specific
null hypothesis: labels are exchangeable given the features, and the
object under test is the whole search procedure, not a single model.
On three reproducible examples we show: (1) the test is calibrated on
pure noise even when the procedure picks the best of many models;
(2) the test has power on a planted signal; (3) on data with a trend,
naive label shuffling inflates the type I error to about 0.7 at a
nominal 0.05. Every number reproduces with one command.

## 1. The practical question

An analyst tried dozens of features and configurations, found a model
with a good holdout score, and announced a discovery. The question:
what exactly guarantees this is not an artifact of multiple search?

The standard answer is to rerun the same procedure on shuffled labels
and ask how unusual the observed score is. That is the permutation
test. But its result is read correctly only if you know which null
hypothesis is being tested and when the permutation scheme is invalid.

## 2. Setup

Let `T(D)` be the score returned by the full search procedure
(candidate generation + fitting + selection on train + holdout
evaluation) on data `D = (X, y)`.

`H0`: the distribution of `T(X, y)` equals the distribution of
`T(X, pi(y))` for a uniformly random permutation `pi`. In words: the
labels `y` are exchangeable given `X`, as seen by the procedure.

p-value = P_H0(T >= T_obs), estimated by Monte Carlo with the finite
run correction: p = (k + 1) / (N + 1), where k is the number of
permutations scoring at least the observed value. The correction keeps
p strictly above zero on any finite run.

Two consequences that are often lost:

- The object under test is the whole procedure, including the
  selection of the best candidate. The test says nothing about a
  specific model in isolation from the procedure that produced it.
- The test is valid only if the permutation destroys the X-y link and
  nothing else. If the data carry structure (time, groups, trend),
  naive shuffling destroys that too, and the p-value loses meaning.

This matches the system-level null calibration of the CV->0
adjudication protocol (section 0.7, doi:10.5281/zenodo.23599882).

## 3. Experiment

Code: `src/permutation_test.py` (numpy, scikit-learn, fixed seed,
SEED=7). Search procedure: enumerate all single features and pairs,
fit OLS on each, keep the candidate with the best train R^2, report
its holdout R^2 on a held out quarter. This mimics common practice:
try many things, report the best.

**Part A, planted signal.** 120 observations, 6 features,
y = 3*x0 - 2*x1 + noise. Expect a small p.

**Part B, pure noise, wide search.** 60 observations, 12 features,
y is pure noise (78 candidates in the loop). On a single split the
"best" model can look decent. Expect a large p: the procedure sees
nothing unusual, because shuffled labels give the selection step the
same scores.

**Part C, spurious trend dependence.** Two independent random walks,
n=100. Naive label shuffling destroys the trend. Expect a falsely
small p: a demonstration that the scheme is invalid here.

Single run results (N_PERM=200):

| Part | observed R^2 | null mean / max | p-value |
|------|--------------|------------------|---------|
| A | 0.971 | -0.146 / 0.205 | 0.005 |
| B | -0.585 | -0.415 / 0.150 | 0.731 |
| C | 0.350 | -0.084 / 0.101 | 0.005 |

Part B is the central one: despite looping over 78 candidates and
keeping the best, the procedure behaves on noise exactly as it does
under H0. The "discovery" would be a selection artifact, and the test
sees it.

## 4. Repeated runs: false positive frequency

One run on one configuration says nothing about error rates.
`experiments/repeated_runs.py` repeats the whole experiment on
independently generated data (50 runs per scenario, alpha=0.05,
Wilson intervals):

| Scenario | Rejections of H0 | Rate | 95% CI |
|----------|------------------|------|--------|
| noise (H0 true) | 1/50 | 0.020 | [0.004, 0.105] |
| signal (H0 false) | 50/50 | 1.000 | [0.929, 1.000] |
| spurious_trend (H0 true) | 36/50 | 0.720 | [0.583, 0.825] |

Reading:

- **noise:** rejection rate 0.020 at alpha=0.05, interval covers the
  nominal level. The test is calibrated on independent data.
- **signal:** power 1.0 on a planted dependence of this size. A sanity
  check that the test can find signal at all.
- **spurious_trend:** type I error 0.720 instead of 0.05, a factor of
  14 above nominal. A nominal p<0.05 on trending data means nothing
  when the permutation scheme is naive.

## 5. Why Part C matters in practice

Random walks are an extreme case, but the same mechanics apply to any
data with autocorrelation, seasonality, group structure (patients,
stores, regions), or a shared external trend. Naive label shuffling
changes not only the X-y link but the structure of the data itself, so
the "null" distribution is built for the wrong problem.

Correct options exist: difference the series before testing, block
permutation, block bootstrap, permutation within groups. Each option
changes the null hypothesis, and that is fine: the hypothesis should
match the data structure, not the other way around. These options are
deliberately not implemented here (see "What I did not test").

## 6. Limitations

Full list in [limitations.md](limitations.md). In short:

- The test checks a procedure, not a specific deployed model.
- p-value measures neither effect size nor causality nor practical
  value.
- The result depends on the permutation scheme; there is no universal
  scheme.
- The Monte Carlo estimate of p is bounded below by 1/(N+1) and
  carries sampling error (shown as Wilson intervals in the code).

## 7. What I did not test

- Whether any specific deployed model is overfit.
- Causality or business value of found relationships.
- Dependence structures other than a shared trend (seasonality,
  groups, heteroskedasticity).
- A correct structure-preserving null for Part C (differencing, block
  bootstrap).
- Nonlinear models and non-R^2 scores (classification, ranking).

## Reproduce

```bash
pip install -r requirements.txt
python src/permutation_test.py
python experiments/repeated_runs.py --runs 50 --n-permutations 200
```

Seeds are fixed, dependencies are pinned in `requirements.txt`.
Repository: https://github.com/sensus-stoa/permutation-test-ml
