package format_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	cml "github.com/grespyrad/CMLGo"

	"github.com/grespyrad/cml-lint/internal/format"
)

// TestPrettyGolden проверяет точный layout всех конструкций и повторный format.
func TestPrettyGolden(t *testing.T) {
	t.Parallel()

	for _, path := range fixturePaths(t, "../../testdata/format/*.input.cml") {
		t.Run(filepath.Base(path), func(t *testing.T) {
			t.Parallel()

			source := readFixture(t, path)
			want := readFixture(t, strings.TrimSuffix(path, ".input.cml")+".golden.cml")
			assertPretty(t, source, want, cml.Dialect{})
		})
	}
}

func fixturePaths(t *testing.T, pattern string) []string {
	t.Helper()

	paths, err := filepath.Glob(pattern)
	if err != nil {
		t.Fatal(err)
	}

	if len(paths) == 0 {
		t.Fatal("missing fixtures")
	}

	return paths
}

func readFixture(t *testing.T, path string) []byte {
	t.Helper()

	//nolint:gosec // Читаются fixtures репозитория.
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	return data
}

func assertPretty(t *testing.T, source, want []byte, dialect cml.Dialect) {
	t.Helper()

	got, err := format.SourceWithDialect("test.cml", source, dialect)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(got, want) {
		t.Fatalf("layout mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}

	again, err := format.SourceWithDialect("test.cml", got, dialect)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(again, want) {
		t.Fatalf("not idempotent:\n%s", again)
	}
}

// TestPrettyDialects проверяет aliases и строгий диалект без подмены raw tokens.
func TestPrettyDialects(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name         string
		fixture      string
		dialect      cml.Dialect
		replacements []string
	}{
		{"original", "blocks", cml.Dialect{Aliases: nil, Original: true}, nil},
		{"russian", "blocks", cml.Dialect{}, []string{
			"BoundedContext", "ОграниченныйКонтекст", "Aggregate", "Агрегат", "Entity", "Сущность",
		}},
		{"custom", "invariants", cml.Dialect{Original: false, Aliases: map[string]string{
			"Contexto": "BoundedContext", "Group": "Aggregate", "regla": "rule",
		}}, []string{"BoundedContext", "Contexto", "Aggregate", "Group", "rule", "regla"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			base := "../../testdata/format/" + tc.fixture
			replace := strings.NewReplacer(tc.replacements...)
			source := replace.Replace(string(readFixture(t, base+".input.cml")))
			want := replace.Replace(string(readFixture(t, base+".golden.cml")))
			assertPretty(t, []byte(source), []byte(want), tc.dialect)
		})
	}
}

// TestConformanceFormatting проверяет весь синтаксически валидный upstream test corpus.
func TestConformanceFormatting(t *testing.T) {
	t.Parallel()

	for _, path := range fixturePaths(t, "../../testdata/conformance/*.cml") {
		t.Run(filepath.Base(path), func(t *testing.T) {
			t.Parallel()

			source := readFixture(t, path)
			if !cml.Parse(path, source).Valid {
				t.Skip("negative syntax fixture")
			}

			output, err := format.Source(path, source)
			if err != nil {
				t.Fatal(err)
			}

			assertPretty(t, source, output, cml.Dialect{})
			assertSameTokens(t, source, output)
		})
	}
}

// TestWhitespaceIndependence проверяет стиль при произвольных входных отступах.
func TestWhitespaceIndependence(t *testing.T) {
	t.Parallel()

	for _, path := range fixturePaths(t, "../../testdata/format/*.golden.cml") {
		t.Run(filepath.Base(path), func(t *testing.T) {
			t.Parallel()

			want := readFixture(t, path)

			tokens, err := cml.Tokenize(path, want)
			if err != nil {
				t.Fatal(err)
			}

			source := disturbWhitespace(t, tokens)
			assertPretty(t, []byte(source), want, cml.Dialect{})
		})
	}
}

func disturbWhitespace(t *testing.T, tokens []cml.Lexeme) string {
	t.Helper()

	parts := make([]string, 0, len(tokens))
	for _, token := range tokens {
		if token.Kind == "comment" {
			t.Skip("comment attachment depends on the original line")
		}

		parts = append(parts, token.Text)
	}

	return "\r\n\t " + strings.Join(parts, "\r\n\t ")
}

// TestReferenceSnapshot сравнивает скачанный образец с выбранным стилем и Java tokens.
func TestReferenceSnapshot(t *testing.T) {
	t.Parallel()

	source := readFixture(t, "../../testdata/upstream/insurance-example/Insurance-Example-Stage-3.cml")
	want := readFixture(t, "../../testdata/format-reference/insurance-stage-3.pretty.cml")
	java := readFixture(t, "../../testdata/format-reference/insurance-stage-3.java.cml")
	assertPretty(t, source, want, cml.Dialect{})
	assertSameTokens(t, source, java)
}

// TestRejectUnbalancedDelimiters закрепляет найденный fuzz случай без panic и записи.
func TestRejectUnbalancedDelimiters(t *testing.T) {
	t.Parallel()

	for _, source := range []string{
		"BoundedContext Request}{", "ContextMap }{", "BoundedContext Request{", "BoundedContext Request([)]",
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()

			output, err := format.Source("test.cml", []byte(source))
			if !errors.Is(err, cml.ErrInvalidModel) || output != nil {
				t.Fatalf("got %q, %v", output, err)
			}
		})
	}
}

// TestPrettyCRLF нормализует разделители, сохраняя исходные raw comments.
func TestPrettyCRLF(t *testing.T) {
	t.Parallel()

	source := "/* Model\r\n * Comment\r\n */\r\nBoundedContext ServiceIntegration{}\r\n"
	want := "/* Model\r\n * Comment\r\n */\nBoundedContext ServiceIntegration {\n}\n"
	assertPretty(t, []byte(source), []byte(want), cml.Dialect{})
}

// TestPreferredWidth проверяет границу по Unicode code points, а не UTF-8 bytes.
func TestPreferredWidth(t *testing.T) {
	t.Parallel()

	const width = 100

	prefix := "  responsibilities = "
	suffix := ", \"Policy\""

	for _, extra := range []int{0, 1} {
		t.Run(strconv.Itoa(extra), func(t *testing.T) {
			t.Parallel()

			value := "\"" + strings.Repeat("Ж", width-len(prefix)-len(suffix)-2+extra) + "\""
			source := "BoundedContext ServiceIntegration{" + prefix + value + suffix + "}"
			line := prefix + value + suffix

			if extra > 0 {
				line = prefix + value + ",\n    \"Policy\""
			}

			want := "BoundedContext ServiceIntegration {\n" + line + "\n}\n"
			assertPretty(t, []byte(source), []byte(want), cml.Dialect{})
		})
	}
}
