# tc — trust_check

 механический валидатор Trust Record (Trust Layer v3.5). Один бинарник на Go,
 stdlib-only, без зависимостей.

Принцип: **машина проверяет механику, человек подтверждает смысл**.
Auto-PASS не повышает статус claim — он утверждает только «механика цела».

## Сборка и установка

```bash
./install.sh
# на нестандартном Go:
GO=~/sdk/go1.27.3/bin/go ./install.sh
```

Инсталлер собирает **статический stripped-бинарь** `bin/tc` и кладёт настоящий файл
`~/.local/bin/tc` (не симлинк — симлинк из системного PATH в рабочий каталог
репо хрупок). Апгрейд после правок — повторный `./install.sh`.
Бинарь без libc-зависимостей: переносится на ноут scp'ом как есть.

## Проверка

```bash
./bin/tc -root /path/to/repo          # отчёт: trust_check_report.json, exit 0/1
```

Проверки: FILES_PRESENT, JSON_VALID, HASHES_MATCH, CLAIMS_XREF, EVIDENCE_XREF,
CONTROLS, REASON_CODES, VALID_UNTIL, LOCK_BLOB, LOCK_ANCESTOR.

`LOCK_BLOB` — прямой урок инцидента I-001: criteria-lock коммит обязан
содержать НЕпустой `trust/claims.md`; пустой lock = невалидный lock.

## Подпись человека (verdict-слой)

```bash
./bin/tc -root . --signoff --signer "Evgeniy Dolgov"
```

Работает только поверх зелёной автопроверки; по каждому claim:
`Space/Enter = подтверждаю, s = пропустить, q = стоп`.
Результат — `trust/signoff.md`: имя, дата, подтверждённые claims, sha256 отчёта.
Зелёная автопроверка без `signoff.md` вердиктом не является; любое изменение
рекорда/объекта инвалидирует подпись и требует прогона заново.

## Самопроверка чекера (негатив-контроли)

```bash
go test -v ./
```

N=4 кейса: битый hash → FAIL; INCONCLUSIVE без reason_code → FAIL;
статистический PASS без контроля → FAIL; чистый рекорд → PASS.
Пятый тест — регрессия I-001 (пустой blob в lock-коммите → FAIL).
