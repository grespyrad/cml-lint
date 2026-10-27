# Оригинальный Java formatter и выбранный стиль

Образец скачан из [Context Mapper examples, commit 3c28b419](https://github.com/ContextMapper/context-mapper-examples/blob/3c28b419e965e74aa3bfe616467cf5ebfff261a8/src/main/cml/insurance-example/Insurance-Example-Stage-3.cml).
Он уже входит в `testdata/upstream/insurance-example/Insurance-Example-Stage-3.cml`;
повторная загрузка сверена побайтово. Copyright comments сохранены, Apache-2.0;
атрибуция в корневом NOTICE.

`insurance-stage-3.java.cml` — результат настоящего Xtext `IFormatter2` из
Maven `org.contextmapper:context-mapper-cli:6.12.0`. Сохранены его tabs и внутренние пустые
строки; окончание файла нормализовано до одного LF. Изучены [ContextMappingDSLFormatter.xtend](https://github.com/ContextMapper/context-mapper-dsl/blob/eb63d120265600cc9d5f531cdbfc9ef44837d78c/org.contextmapper.dsl/src/org/contextmapper/dsl/formatting2/ContextMappingDSLFormatter.xtend)
и [TacticDDDLanguageFormatter.xtend](https://github.com/ContextMapper/context-mapper-dsl/blob/eb63d120265600cc9d5f531cdbfc9ef44837d78c/org.contextmapper.dsl/src/org/contextmapper/tactic/dsl/formatting2/TacticDDDLanguageFormatter.xtend).

`insurance-stage-3.pretty.cml` — канонический результат Go formatter. Он использует
два пробела, регулярные пустые строки между блоками и пробелы вокруг стрелок.
Равенство Java whitespace не является целью; обе версии сохраняют токены образца.
Другие случаи с generic types, operations, comments, long lists и расширениями
проверяются точными fixtures в `testdata/format`.

Для воспроизведения скачайте [CLI 6.12.0 TAR](https://repo.maven.apache.org/maven2/org/contextmapper/context-mapper-cli/6.12.0/context-mapper-cli-6.12.0.tar),
распакуйте его и укажите пути к `lib`, JDK 17 `bin` и собранному `cm`:

```sh
go build -o bin/cm ./cmd/cm
python3 scripts/formatter_reference.py \
  --reference-lib /absolute/path/context-mapper-cli-6.12.0/lib \
  --java-bin /absolute/path/jdk-17/bin \
  --cm bin/cm --output /tmp/cml-format-reference
diff -u testdata/format-reference/insurance-stage-3.java.cml /tmp/cml-format-reference/insurance-stage-3.java.cml
diff -u testdata/format-reference/insurance-stage-3.pretty.cml /tmp/cml-format-reference/insurance-stage-3.pretty.cml
```

Обычные Go tests используют сохранённые snapshots и не требуют сети или Java.
