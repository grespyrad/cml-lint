## MODIFIED Requirements

### Requirement: Версия поставки
CLI SHALL объявлять версию 0.3.0 и закреплять go-cml v0.3.0 в основном
и исследовательском Go-модулях без локального replace.

#### Scenario: Версия CLI
- **WHEN** пользователь вызывает cm --version
- **THEN** CLI SHALL вывести 0.3.0.
