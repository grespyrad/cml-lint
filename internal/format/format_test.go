package format_test

import (
	"bytes"
	"io/fs"
	"os"
	"testing"

	"github.com/grespyrad/go-cml"

	"github.com/grespyrad/cml-lint/internal/format"
)

// TestFormat проверяет контракт линтера и форматтера.
//
//nolint:gosmopolitan // Unicode проверяется сохранением токенов.
func TestFormat(t *testing.T) {
	t.Parallel()

	source := []byte(
		"// Сервис 文 🔧\nBoundedContext ServiceIntegration{domainVisionStatement=\"Сервис 文 🔧\" " +
			"Aggregate A{Entity ServiceIntegration{aggregateRoot String id - ServiceIntegration other}}} // trailing\n",
	)

	formatted, err := format.Source("test.cml", source)
	if err != nil {
		t.Fatal(err)
	}

	again, err := format.Source("test.cml", formatted)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(formatted, again) {
		t.Fatalf("not idempotent\n%s\n%s", formatted, again)
	}

	assertSameTokens(t, source, formatted)

	if bytes.Equal(source, formatted) || !bytes.Contains(formatted, []byte("    Entity ServiceIntegration {\n")) {
		t.Fatalf("unexpected layout\n%s", formatted)
	}
}

// TestRejectSyntax проверяет контракт линтера и форматтера.
func TestRejectSyntax(t *testing.T) {
	t.Parallel()

	if _, err := format.Source("test.cml", []byte("BoundedContext ServiceIntegration {")); err == nil {
		t.Fatal("invalid syntax accepted")
	}
}

// TestFormatCorpus проверяет контракт линтера и форматтера.
//
//nolint:gocognit // Обход corpus и точные проверки находятся в одном test scope.
func TestFormatCorpus(t *testing.T) {
	t.Parallel()

	err := fs.WalkDir(os.DirFS("../../testdata/upstream"), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		t.Run(path, func(t *testing.T) {
			t.Parallel()

			//nolint:gosec // Читается собственный fixture corpus.
			data, e := os.ReadFile("../../testdata/upstream/" + path)
			if e != nil {
				t.Fatal(e)
			}

			formatted, e := format.Source(path, data)
			if e != nil {
				t.Fatal(e)
			}

			again, e := format.Source(path, formatted)
			if e != nil {
				t.Fatal(e)
			}

			if !bytes.Equal(formatted, again) {
				t.Fatal("not idempotent")
			}

			assertSameTokens(t, data, formatted)
		})

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func assertSameTokens(t *testing.T, source, output []byte) {
	t.Helper()

	first, err := cml.Tokenize("test.cml", source)
	if err != nil {
		t.Fatal(err)
	}

	second, err := cml.Tokenize("test.cml", output)
	if err != nil {
		t.Fatal(err)
	}

	if len(first) != len(second) {
		t.Fatal("token count changed")
	}

	for i := range first {
		if first[i].Kind != second[i].Kind || first[i].Text != second[i].Text {
			t.Fatalf("token %d changed", i)
		}
	}
}
