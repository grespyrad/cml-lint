# cml-validation

## Purpose

Независимый Go CLI с внешней CMLGo библиотекой для проверки исходного CML. Контракт основан
на эталоне etalon-go-library и грамматиках Context Mapper 6.12.0.

## Requirements

### Requirement: Полная грамматика CML
Parser SHALL принимать все правила стратегической и тактической грамматики
6.12.0, включая домены, требования, stakeholders, values, application flows,
coordination, services, repositories, resources, traits, DTO, enum и constraints.
Неизвестные конструкции и оставшийся после модели текст MUST отклоняться.

#### Scenario: Учебные модели
- **WHEN** проверяются синтетические модели интеграции сервиса service.cml и service-before.cml
- **THEN** обе модели SHALL быть валидны.

#### Scenario: Синтаксическая ошибка
- **WHEN** пропущена скобка, строка не закрыта или указана неизвестная конструкция
- **THEN** диагностика SHALL содержать файл, строку, колонку и код ошибки.

### Requirement: Семантика и ссылки
Validator SHALL проверять typed references, imports, uniqueness, relationship
roles, map membership, aggregate roots/lifecycle, team ownership, flows,
coordination и ограничения тактических свойств. Warnings MUST не менять valid.
Поведение SHALL сверяться с официальным валидатором, расхождения MUST быть явны.

#### Scenario: Невалидные ссылки и агрегат
- **WHEN** модель содержит отсутствующий context, две aggregateRoot или duplicate context
- **THEN** valid SHALL быть false и ошибки SHALL объяснять нарушенное правило.

#### Scenario: Импорты
- **WHEN** модель импортирует соседний файл, включая повторный или циклический импорт
- **THEN** ссылки SHALL разрешаться без бесконечной рекурсии; отсутствующий файл MUST давать ошибку.

### Requirement: Unicode без изменения формата
Строки и комментарии SHALL принимать UTF-8, включая кириллицу, CJK и emoji.
Идентификаторы SHALL следовать оригинальному CML, включая escaped identifiers.
Невалидный UTF-8 MUST отклоняться; YAML и расширения инвариантов MUST не вводиться.

#### Scenario: Unicode описание
- **WHEN** domainVisionStatement содержит "Сервис 🔧 文" и комментарий содержит Unicode
- **THEN** модель SHALL приниматься без изменения содержимого строки.

### Requirement: Публичный API и CLI
CLI SHALL импортировать versioned CMLGo для parser и semantic validation.
CLI SHALL поддерживать validate -i/--input, --json, help и version.
Exit codes SHALL быть 0 для валидной модели, 1 для ошибок модели/IO и 2 для неверных аргументов.
Production CLI MUST не требовать JVM для parsing, validation или PlantUML text generation.

#### Scenario: Машинная проверка
- **WHEN** вызывается cm validate --json -i model.cml
- **THEN** stdout SHALL содержать JSON valid/diagnostics, а exit code SHALL соответствовать valid.

### Requirement: Диаграммы для существующего сценария
CLI SHALL предоставлять generate -g plantuml -i model.cml -o dir для context map
и тактических class diagrams. Генерация MUST сначала валидировать модель.
Точное совпадение layout и всех оригинальных генераторов SHALL не утверждаться.

#### Scenario: Генерация синтетической модели
- **WHEN** генерируется PlantUML для service.cml
- **THEN** файлы SHALL включать context map, сущности, ссылки, команды и события.

### Requirement: Проверяемая производительность
Benchmark SHALL сравнивать одинаковую грамматику и corpus Go/Rust,
разделять startup, syntax и full validation и сохранять методику и сырые результаты.
Rust prototype MUST не выдаваться за полный эквивалент семантического валидатора.

#### Scenario: Воспроизводимый замер
- **WHEN** запускаются benchmark команды из документации
- **THEN** результаты SHALL содержать версии, workload, единицы и ограничения сравнения.
