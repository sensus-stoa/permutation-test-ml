# Permutation testing for machine learning:
# which null hypothesis it tests, and what it does not prove

> **Disclosure:** AI tools were used in preparing the code and text.
> Author: Evgeny Dolgov, independent researcher, CV->0 program.
> Contact: [to be filled in at publication].
>
> Russian version: [README.ru.md](README.ru.md).

## One line

We test not "is there signal" but "how unusual is the result of the
whole search procedure under a specific null hypothesis".

## What is here

- **Part A.** Data with a planted linear dependence. The test returns
  a small p-value. As expected.
- **Part B.** Pure noise, but a wide search procedure (many candidates,
  best picked on train). The test returns a large p-value: the
  "discovery" is a selection artifact, not signal.
- **Part C.** Two independent random walks with a shared trend. Naive
  label shuffling destroys the structure and returns a falsely small
  p-value. This shows why the permutation scheme must preserve the
  dependence structure.

## Null hypothesis (formal)

Let `T(D)` be the score returned by the full search procedure on data `D`.

`H0`: the distribution of `T(X, y)` equals the distribution of
`T(X, pi(y))` for a uniformly random permutation `pi`. In words: `y`
is exchangeable given `X`, as seen by the procedure.

The object under test is the procedure (search + selection +
evaluation), not one fixed model. This matches the system-level null
calibration of the CV->0 adjudication protocol (section 0.7,
doi:10.5281/zenodo.23599882): run the whole pipeline on label-shuffled
data and count false acceptances.

## Run

```bash
pip install -r requirements.txt
python src/permutation_test.py
```

Fixed seed: `SEED=7`. Runtime: a few seconds.
Dependencies are pinned in `requirements.txt` (`numpy`, `scikit-learn`).

## What reproduces

A clean run with one command. The seed is fixed. Every printed number
is deterministic given the same library versions.

Single run (SEED=7, N_PERM=200):

| Part | Data | observed R^2 | p-value | Reading |
|------|------|--------------|---------|---------|
| A | planted linear signal | 0.971 | 0.005 | H0 rejected, signal visible |
| B | pure noise, best of 78 candidates | -0.585 | 0.731 | the "best" model is indistinguishable from chance selection |
| C | two independent random walks | 0.350 | 0.005 | false rejection: naive shuffling destroyed the trend |

## Repeated runs: false positive frequency

The outer loop in `experiments/repeated_runs.py` repeats the whole
experiment on independently generated data (50 runs per scenario,
alpha=0.05):

| Scenario | Rejections of H0 | Rate | 95% CI (Wilson) |
|----------|------------------|------|------------------|
| noise (H0 true) | 1/50 | 0.020 | [0.004, 0.105] |
| signal (H0 false) | 50/50 | 1.000 | [0.929, 1.000] |
| spurious_trend (H0 true, naive shuffle) | 36/50 | 0.720 | [0.583, 0.825] |

Reading: on pure noise the test is calibrated (the rejection rate is
consistent with alpha=0.05). On the planted signal the test has power.
On dependent data the naive scheme inflates the type I error to about
0.7 instead of 0.05, a factor of 14 above nominal.

Reproduce the summary:

```bash
python experiments/repeated_runs.py --runs 50 --n-permutations 200
```

Machine-readable output is written to `results/repeated_runs.csv`
(the `results/` directory is git-ignored; regenerate with the command
above).

## What is not here

- Proof that any specific deployed model is overfit
- A test of causality or business value of a found relationship
- Coverage of all dependence structures (only a shared trend here)
- A correct structure-preserving null for Part C
  (for example differencing the series, or a block bootstrap)

## Extended limitations

See `docs/limitations.md`. Article text: `docs/article.md`.

## License

MIT.
