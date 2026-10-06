# cml-extensions

## Purpose

Интеграция декларативных инвариантов и языковых возможностей CMLGo.

## Requirements

### Requirement: Расширенная модель
CLI SHALL проверять и форматировать CMLGo Invariant blocks с сохранением tokens,
Unicode names и optional Russian aliases. Строки MUST не переводиться.

#### Scenario: Декларативный инвариант
- **WHEN** lint --fix получает агрегат с именованным rule и trace links
- **THEN** файл SHALL форматироваться, семантика SHALL проверяться библиотекой.

### Requirement: Явный выбор диалекта
Глобальный --original SHALL включать исходный CML 6.12.0 перед командой.
--keywords SHALL читать JSON alias-to-canonical dictionary и применять его
ко всему import graph, format и validation. Неверный словарь MUST давать nonzero.

#### Scenario: Другой язык
- **WHEN** cm --keywords keywords.json validate -i model.cml использует Contexto
- **THEN** словарь SHALL работать без изменения последующих обычных вызовов.

#### Scenario: Исходная грамматика
- **WHEN** cm --original validate получает Invariant block
- **THEN** расширение MUST отклоняться.
