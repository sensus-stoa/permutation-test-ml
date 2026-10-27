# permutation-test-ml

Оценка статистической значимости метрики качества ML-модели
с помощью перестановочного теста (permutation test).

## Установка

    python -m venv .venv
    source .venv/bin/activate
    pip install -r requirements.txt

## Быстрый старт

    from src.permutation_test import permutation_test

    p_value = permutation_test(
        y_true=y_true,
        y_pred_a=pred_a,
        y_pred_b=pred_b,
        metric=my_metric,
        n_permutations=10_000,
        random_state=42,
    )
    print(f"p-value = {p_value:.4f}")

## Воспроизведение экспериментов

    python experiments/repeated_runs.py --runs 1000 --n-permutations 1000

Результаты сохраняются в `results/` (машиночитаемый формат — CSV/JSON).

## Структура репозитория

| Путь | Назначение |
|---|---|
| `src/permutation_test.py` | реализация перестановочного теста |
| `experiments/repeated_runs.py` | внешний цикл: частота ложных срабатываний |
| `results/` | машиночитаемые результаты экспериментов |
| `docs/article.md` | текст статьи |
| `docs/limitations.md` | ограничения метода |

## Лицензия

MIT — см. [LICENSE](LICENSE).
