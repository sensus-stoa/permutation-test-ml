"""Permutation test for comparing two models on the same data.

Реализация парного перестановочного теста: проверяем гипотезу,
что две модели имеют одинаковое качество по выбранной метрике.

Идея: под нулевой гипотезой предсказания моделей для каждого объекта
обмениваемы. Многократно случайно переставляем предсказания между
моделями по каждому объекту, считаем метрику и строим
эмпирическое распределение разности. p-value — доля перестановок
с |разностью| не меньше наблюдаемой.
"""
from __future__ import annotations

from typing import Callable, Optional

import numpy as np
from sklearn.metrics import accuracy_score


def permutation_test(
    y_true: np.ndarray,
    y_pred_a: np.ndarray,
    y_pred_b: np.ndarray,
    metric: Callable[[np.ndarray, np.ndarray], float] = accuracy_score,
    n_permutations: int = 10_000,
    random_state: Optional[int] = None,
) -> float:
    """Парный перестановочный тест на равенство качества двух моделей."""
    rng = np.random.default_rng(random_state)

    y_true = np.asarray(y_true)
    y_pred_a = np.asarray(y_pred_a)
    y_pred_b = np.asarray(y_pred_b)

    if not (y_true.shape == y_pred_a.shape == y_pred_b.shape):
        raise ValueError("Все массивы должны иметь одинаковую длину.")

    observed_diff = abs(metric(y_true, y_pred_a) - metric(y_true, y_pred_b))

    n = len(y_true)
    count = 0
    for _ in range(n_permutations):
        swap = rng.random(n) < 0.5
        perm_a = np.where(swap, y_pred_b, y_pred_a)
        perm_b = np.where(swap, y_pred_a, y_pred_b)
        diff = abs(metric(y_true, perm_a) - metric(y_true, perm_b))
        if diff >= observed_diff:
            count += 1

    return (count + 1) / (n_permutations + 1)


if __name__ == "__main__":
    rng = np.random.default_rng(0)
    y = rng.integers(0, 2, size=200)
    a = y.copy()
    b = y.copy()
    flip = rng.random(200) < 0.1
    b[flip] = 1 - b[flip]
    p = permutation_test(y, a, b, n_permutations=2_000, random_state=0)
    print(f"p-value = {p:.4f}")
