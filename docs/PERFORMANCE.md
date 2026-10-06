# Исследование производительности

Production parser и semantic validator находятся в CMLGo. Rust/C++ probes
реализуют только распознавание исходной CML 6.12.0 grammar. Их время нельзя
сравнивать со временем полной семантической проверки как эквивалентные операции.

## Воспроизведение

```sh
python3 scripts/build_bench.py
python3 scripts/bench_parity.py
python3 scripts/compare.py --rounds 7 --iterations 100
python3 scripts/compare_stages.py --rounds 5 --benchtime 200ms
GOWORK=off GOMAXPROCS=1 go -C bench/lex test -run='^$' -bench=BenchmarkStages -benchmem
```

Reference workloads `service-before.cml` и `service.cml` описывают условную интеграцию
сервиса. После замены моделей
предыдущие времена неприменимы; готовые сравнительные результаты не публикуются.
Замеры и профили сохраняются локально и игнорируются Git: они зависят от машины
и могут включать пути окружения. Время не является assertion.

## Что проверяется

`bench/lex` закрепляет CMLGo и Participle v2.1.4. TestLexerParity сравнивает
raw tokens/kinds/positions на upstream, conformance, escape и reference corpus;
TestLexerEdges проверяет Unicode, keywords, quoting и невалидный UTF-8.
`task check` включает эти проверки. Parity на corpus не доказывает эквивалентность
на любом возможном входе. Новые примеры с Invariant относятся к расширенному
CMLGo и не являются входом Original syntax probes.

BenchmarkStages измеряет Tokenize, CheckSyntax, Parse, Validate и Unmarshal.
Каждый API выполняется отдельно; их времена нельзя складывать. Unmarshal
строит typed model и не выполняет semantic validation. Participle измеряется
с CML адаптером и keyword lookup, без setup внутри timed loop.

```sh
python3 scripts/oracle_compare.py --reference-lib /path/to/context-mapper-cli-6.12.0/lib --java-bin /path/to/jdk/bin
```

В `bench/results/differential-conformance.json` сохранены 273/273 совпадений
validity с полным Xtext IResourceValidator CheckMode.ALL.
`cli-load-only-conformance.json` — отдельный результат старого CLI load/getErrors,
который не заменяет semantic oracle. Пути в сохранённых отчётах относительные.
Production не требует JVM.
