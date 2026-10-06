# cml-performance

## Purpose

Воспроизводимые тесты корректности и производительности на одинаковой работе.

## Requirements

### Requirement: Корректность перед значениею
Исследование SHALL проверять parity Go/Rust/C++ на corpus перед замерами.
Альтернативная Go-библиотека lexer SHALL сравниваться по типам, исходному тексту
и Unicode позициям токенов, включая неверные UTF-8, escapes и комментарии.

#### Scenario: Пограничный ввод
- **WHEN** тесты получают незакрытые строки, surrogate без пары и keyword prefix
- **THEN** оба lexer SHALL совпадать по отказу или полной последовательности токенов.

### Requirement: Разделение стадий
Go benchmark tests SHALL измерять Tokenize, CheckSyntax, Parse, Validate и Unmarshal
на одинаковых эталонах; SHALL выводить ns/op, B/op и allocs/op. CPU и allocation
профили SHALL сохраняться с командами воспроизведения. Профиль SHALL отделять
runtime от прикладного кода, а сравнение lexer MUST не выдаваться за semantic parity.

#### Scenario: Сравнение библиотеки
- **WHEN** запускается benchmark lexer CMLGo и Participle
- **THEN** отчёт SHALL содержать результаты обоих и ограничения адаптера;
  pipeline выигрыш MUST не вычисляться простым переносом lexer valueup.

### Requirement: Без нестабильных порогов
Тесты MUST не падать из-за случайного отклонения времени на машине. Отчёт SHALL
сохранять повторные сырые измерения и версии, показывать выигрыш и проигрыш.

#### Scenario: Повторный запуск
- **WHEN** benchmark выполняется на другой машине
- **THEN** correctness gates SHALL сохраняться, timing SHALL пересчитываться без заданного рейтинга.
