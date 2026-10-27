# Limitations

Extended list of limitations of the permutation calibration of a
search procedure, as implemented in `src/permutation_test.py`.

## About the object under test

- **The procedure is tested, not a model.** The p-value refers to the
  bundle "search + selection + evaluation". Changing one step (a
  different selection metric, a different candidate set) makes it a
  different object and, potentially, gives a different answer.
- **It does not prove a specific model is overfit.** A large p on
  noise means "the procedure extracted nothing beyond chance
  selection", not a verdict on a single trained model.

## About the null hypothesis

- **The result depends on the permutation scheme.** H0 is stated
  relative to a scheme. Naive label shuffling is valid only for
  independent observations. For time series, groups, or spatial data
  the scheme must preserve the structure (block permutation,
  within-group permutation, differencing), otherwise the type I error
  inflates: in Part C of this repository it reaches 0.72 at a nominal
  0.05.
- **Exchangeability is an assumption, not a fact.** The naive scheme
  assumes that under H0 the labels are exchangeable given the
  features. This is a substantive claim about the data-generating
  mechanism, and the test itself does not check it.

## About computation

- **Monte Carlo error.** p is estimated from a finite number of
  permutations. The smallest reachable p is 1/(N+1); resolving p<0.001
  needs at least about 1000 permutations, and the estimate still
  carries sampling error (the code reports it as a Wilson interval).
- **Cost.** Each permutation is a full run of the search procedure.
  The cost of the test is N_PERM times the cost of the procedure. For
  heavy pipelines this is the main practical barrier.

## About the examples

- **One dependence structure.** Part C demonstrates only a shared
  trend (two random walks). Seasonality, group structure,
  heteroskedasticity, and spatial autocorrelation are not covered.
- **One score, one model family.** Linear regression and R^2 only.
  The behavior of the test with other scores (classification, ranking)
  and nonlinear models is not examined here.
- **Fixed split.** The train/holdout split is the same across all
  permutations (this is correct: only the label assignment changes),
  but the result is not averaged over splits.

## What the test does not do at all

- It does not test causality.
- It does not measure the size or practical value of an effect (use
  confidence intervals and domain analysis).
- It does not replace an honest validation protocol (preregistered
  hypotheses, external data, independent replication).
