# Verifier Track Record — tc (trust_check)

Принцип: чекер, который всегда говорит PASS, бесполезен. Каждая версия чекера
обязана проходить собственные негатив-контроли: посаженные дефекты обязаны
ловиться, чистый рекорд — проходить. Это Verifier Track Track-practice,
применённый к самому инструменту: чекер декларирует надёжность, значит сам
подчиняется тем же правилам.

## Прогон N=5 — 2026-11-09

- Tool: `tc` = trust_check v0.1.0, Go 1.27.3, stdlib-only
- Объект: sensus-stoa/permutation-test-ml, Trust Record V-PTM-001
- Инфраструктура дефектов: `checks_test.go` собирает фикстуры в `t.TempDir()`
  (копия живого `trust/`, мутация JSON; для кейса 5 — `git init` с пустым
  `trust/claims.md` в lock-коммите)

| # | Кейс | Ожидание | Наблюдение |
|---|------|----------|------------|
| 1 | Чистый рекорд (репозиторий как есть) | overall PASS, exit 0 | PASS, exit 0 — 10/10 проверок |
| 2 | Посажен битый sha256 (E-01, один nibble) | FAIL в HASHES_MATCH | FAIL: declared vs actual, exit 1 |
| 3 | INCONCLUSIVE без reason_code | FAIL в REASON_CODES | FAIL, exit 1 |
| 4 | Статистический PASS (verification V-02) без отрицательного контроля | FAIL в CONTROLS | FAIL, exit 1 |
| 5 | Регрессия I-001: lock-коммит с пустым trust/claims.md | FAIL в LOCK_BLOB | FAIL, exit 1 |

Прогон: `go test -v .` → 5/5 PASS; бинарные прогоны: exit-коды 1 (дефект) / 0
(чистый) — контракт для CI.

## Правило

Любое изменение `checks.go` требует повторного прогона всех пяти кейсов до
коммита. Чекер не выпускается с неизвестным поведением на собственных
дефектах.

Связь с инцидентом: кейс 5 — постоянная регрессия I-001 (trust/claims.md
был закоммичен пустым в критерии-лок; см. `trust/trust-incidents.md`).
