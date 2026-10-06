# cml-lint

Линтер и форматтер Context Mapper Language 6.12.0 на Go. Импортирует отдельную
[CMLGo](https://github.com/grespyrad/CMLGo) `v0.2.3`; parser и typed model
находятся в библиотеке. Runtime не требует Java. Поддерживает исходный CML и расширения CMLGo v0.2.0: инварианты, Unicode names
и необязательные DDD aliases. Репозиторий основан на etalon-go-library.

## Использование

Для сборки нужен Go 1.27.1.

```sh
GOWORK=off go build -o bin/cm ./cmd/cm
bin/cm validate -i model.cml --json
bin/cm lint -i model.cml --json
bin/cm lint --fix -i model.cml
bin/cm format model.cml
bin/cm format --check model.cml
bin/cm format -w model.cml
bin/cm generate -g plantuml -i model.cml -o diagrams
```

Exit codes: 0 — проверка успешна; 1 — ошибка модели/IO или необходимое
форматирование; 2 — неверные аргументы. `format` проверяет синтаксис;
`lint` также проверяет семантику. `--fix` форматирует синтаксически корректный
файл и затем снова проверяет семантику. Импорты проверяет CMLGo.

Форматтер сохраняет токены, исходный текст строк и комментариев. Запись `-w`
атомарна, сохраняет права и отказывает при symlink или изменении исходника во
время работы. Синтаксическая ошибка оставляет файл без изменения.
Unicode поддержан в строках и комментариях, включая escapes/surrogate pairs.
С v0.2.0 доступны Unicode names, декларативные Invariant blocks и необязательные
Russian aliases. English остаётся основным вариантом. Глобальный `--original`
включает исходный ASCII CML 6.12.0; `--keywords dict.json` задаёт другой словарь.

PlantUML генерация включает context map и class diagrams, сущности, ссылки,
команды, события и операции. Layout и имена файлов отличаются от Java generator;
остальные генераторы Context Mapper не реализованы. Для SVG/PNG используйте
внешний PlantUML renderer. Это покрывает существующий model-render сценарий.

## Проверки

```sh
go tool -modfile=tools/task.mod task tools
go tool -modfile=tools/task.mod task check
go tool -modfile=tools/task.mod task hooks
```

Включены OpenSpec, полный lint profile эталона, race/shuffle, module graph,
architecture positive/negative probes, vulnerability scan и portable instruction
checker. Formatter corpus проверяет сохранение токенов и идемпотентность на 55
официальных примерах. Reference models описывают синтетическую интеграцию сервиса.
CMLGo semantic parity проверена на 273 upstream test snippets полным Xtext
validator, а не только load/getErrors старого CLI. Это проверка сохранённого corpus,
не доказательство любого возможного CML-файла или идентичности всех диагностик.

[Исследование производительности](docs/PERFORMANCE.md) содержит воспроизводимые
Go/Rust/C++ syntax probes, correctness parity и сохранённые отчёты совместимости в `bench/results`.
Rust/C++ probes не реализуют semantic validation. Production остаётся на Go.
Также сохранены benchmark tests стадий и сравнение lexer с Participle.
Методика явно разделяет полную валидацию и syntax-only probes.

Атрибуция: [NOTICE](NOTICE), лицензия Apache-2.0. Независимый проект, не официальный
продукт Context Mapper. Tooling и ограничения: [QUALITY.md](docs/QUALITY.md).

```sh
bin/cm --original validate -i model.cml
bin/cm --keywords keywords.json lint --fix -i model.cml
```

DDD покрытие, синтаксис инвариантов и книжные источники: [CMLGo DDD analysis](https://github.com/grespyrad/CMLGo/blob/main/docs/DDD-COVERAGE.md).

[Парные примеры интеграции ExternalService](examples/README.md): русские и
английские имена, четыре области инвариантов, описание ограничений и команды CLI.

[Модель эталонного Go-сервиса](examples/README.md#модель-эталонного-go-сервиса):
контексты, агрегаты, repositories, состояния, операции и application flows;
варианты с исходными и обобщёнными именами. Выбранные правила проверяются
нейтральной Go-реализацией в CMLGo examples/reference.
