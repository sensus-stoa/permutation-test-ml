# Claims — permutation-test-ml (public Trust Record)

Protocol: CV→0 Trust Layer v3.5 (spec frozen for the cycle until 2027-02-04)
Object: https://github.com/sensus-stoa/permutation-test-ml
Object version: see trust/evidence-manifest.json (object.commit)
Criteria lock: the PASS/FAIL criteria below are fixed in THIS file.
The lock commit hash and timestamp are recorded in
trust/evidence-manifest.json (criteria_lock). Any change to these criteria
after the lock commit is a new version marked post-hoc and requires a
re-run for PASS.

## C-01 Reproducibility of the printed results
Claim: a clean run
  pip install -r requirements.txt
  python src/permutation_test.py
with SEED=7 prints exactly: Part A observed R2 0.971, p 0.005;
Part B observed R2 -0.585, p 0.731; Part C observed R2 0.350, p 0.005.
Scope: dependency versions pinned in requirements.txt; single machine.
PASS criterion: every printed observed R2 and p-value matches the values
above as printed (3 decimals).
Verification V-01: executor runs the two commands on a clean checkout and
compares printed numbers.

## C-02 Calibration under H0 (pure noise, exchangeable labels)
Claim: on pure noise the label-permutation test rejects H0 at a rate
consistent with alpha = 0.05.
Negative control NC-NOISE: N = 50 independent runs; false rejections = 1;
rate 0.020; Wilson 95% CI [0.004, 0.105].
PASS criterion: the Wilson 95% CI of the rejection rate contains 0.05.
Verification V-02: python experiments/repeated_runs.py --runs 50
--n-permutations 200; compare the noise row with the README table.

## C-03 Power under planted linear signal
Claim: on a planted linear signal the test rejects H0 in at least 90% of runs.
Control NC-SIGNAL: N = 50; rejections = 50; rate 1.000;
Wilson 95% CI [0.929, 1.000].
PASS criterion: lower bound of the Wilson 95% CI >= 0.90.
Verification: V-02, signal row.

## C-04 Naive label permutation is invalid under a shared trend
Claim: on two independent random walks, naive label permutation yields a
false-rejection rate far above alpha. This certifies the DEMONSTRATION of
a limitation, NOT capability on trended data.
Control NC-TREND: N = 50; false rejections = 36; rate 0.720;
Wilson 95% CI [0.583, 0.825].
PASS criterion: lower bound of the Wilson 95% CI > 0.20.
Verification: V-02, spurious_trend row.

## C-05 Applicability to real production ML pipelines
Status: INCONCLUSIVE, reason_code OUT_OF_SCOPE. No claim is made.
See trust/limitations.md.