# cml-format

## Purpose

Линтер и форматтер исходного CML на основе отдельной библиотеки CMLGo.

## Requirements

### Requirement: Внешняя библиотека
CLI MUST импортировать CMLGo через закреплённый Go module version. Копии
parser, grammar или generated model types в production cml-lint MUST отсутствовать.
Research-only grammar snapshots под bench SHALL быть явно отделены от production.

#### Scenario: Изолированный build
- **WHEN** go build выполняется с GOWORK=off с опубликованной зависимостью
- **THEN** бинарник SHALL собраться без локального replace.

### Requirement: Безопасное форматирование
format SHALL сохранять последовательность токенов, значения строк, комментарии
и Unicode. Format MUST быть идемпотентным и отклонять синтаксически неверный CML.
По умолчанию результат SHALL идти в stdout; -w SHALL атомарно обновлять файл,
--check SHALL возвращать 1, если файл требует форматирования.

#### Scenario: Комментарии и описания
- **WHEN** format получает minified CML с русскими комментариями и emoji
- **THEN** все комментарии и строки SHALL сохраниться без изменения содержимого.

#### Scenario: Неверный ввод
- **WHEN** format -w получает CML с отсутствующей скобкой
- **THEN** исходный файл MUST остаться неизменённым.

### Requirement: Совместная проверка
lint SHALL выполнять semantic validation и проверку canonical formatting;
--fix SHALL форматировать только валидный синтаксис и затем повторно проверить файл.
validate -i и generate -g plantuml SHALL сохраняться для существующего сценария.
JSON diagnostics MUST поддерживать машинное использование.

#### Scenario: Проверка стиля
- **WHEN** lint получает валидный, но неформатированный CML
- **THEN** exit code SHALL быть 1 и диагностика SHALL указывать formatting.
