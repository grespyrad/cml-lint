## ADDED Requirements

### Requirement: Каноническая зависимость go-cml
CLI и исследовательский bench/lex SHALL импортировать
github.com/grespyrad/go-cml v0.3.0 с именем пакета cml без import alias.
Документация, ссылки и архитектурные проверки MUST использовать новый путь.

#### Scenario: Сборка из опубликованного модуля
- **WHEN** пользователь собирает CLI или запускает bench/lex tests
- **THEN** Go SHALL загружать go-cml v0.3.0 без локального replace.
