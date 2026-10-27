# Changelog

All notable changes to this project are documented here.
Format based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
versioning follows [SemVer](https://semver.org/).

## [Unreleased]

## [0.1.0] - 2026-10-27

### Added
- `src/permutation_test.py`: permutation calibration of a search
  procedure (candidate enumeration + best-of selection on train +
  holdout R^2) under the null of labels being exchangeable given the
  features. Three parts: planted signal, pure noise with a wide
  search, spurious dependence of two random walks under naive
  shuffling.
- `experiments/repeated_runs.py`: outer loop estimating rejection
  frequencies over repeated runs with Wilson intervals. Results
  (50 runs per scenario, alpha=0.05): noise 1/50, signal 50/50,
  spurious_trend 36/50.
- Documentation: `docs/article.md`, `docs/limitations.md`.
- `README.md` (English, primary), `README.ru.md` (Russian),
  `LICENSE` (MIT), `requirements.txt` (pinned versions).
