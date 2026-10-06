//nolint:cyclop // Средняя сложность отражает dispatch команд и grammar layout.
package format_test

import (
	"bytes"
	"io/fs"
	"os"
	"testing"

	cml "github.com/grespyrad/CMLGo"

	"github.com/grespyrad/cml-lint/internal/format"
)

// TestFormat проверяет контракт линтера и форматтера.
//
//nolint:gosmopolitan,cyclop,gocognit,gocyclo // Unicode и layout проверяются сохранением токенов и corpus.
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

	first, err := cml.Tokenize("test.cml", source)
	if err != nil {
		t.Fatal(err)
	}

	second, err := cml.Tokenize("test.cml", formatted)
	if err != nil {
		t.Fatal(err)
	}

	if len(first) != len(second) {
		t.Fatal("token count changed")
	}

	for i, token := range first {
		if token.Text != second[i].Text {
			t.Fatalf("token %d changed", i)
		}
	}

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
//nolint:cyclop,gocognit,gocyclo // Layout грамматики проверен token conservation и idempotency corpus.
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

			a, e := cml.Tokenize(path, data)
			if e != nil {
				t.Fatal(e)
			}

			b, e := cml.Tokenize(path, formatted)
			if e != nil {
				t.Fatal(e)
			}

			if len(a) != len(b) {
				t.Fatal("token count changed")
			}

			for i := range a {
				if a[i].Text != b[i].Text {
					t.Fatalf("token %d changed", i)
				}
			}
		})

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
