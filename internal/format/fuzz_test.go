//nolint:cyclop // Corpus и fuzz проверяют все ветви отказа форматтера.
package format_test

import (
	"bytes"
	"testing"

	"github.com/grespyrad/cml-lint/internal/format"
)

// FuzzSource проверяет идемпотентность любого принятого formatter input.
//
//nolint:gosmopolitan // Seed проверяет сохранение Unicode описаний.
func FuzzSource(f *testing.F) {
	f.Add([]byte(`BoundedContext ServiceIntegration { domainVisionStatement "Сервис 🔧 文" }`))
	f.Add([]byte("// comment\nContextMap {}"))
	f.Add([]byte(`"bad`))
	f.Fuzz(func(t *testing.T, source []byte) {
		output, err := format.Source("fuzz.cml", source)
		if err != nil {
			return
		}

		again, err := format.Source("fuzz.cml", output)
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(output, again) {
			t.Fatal("format is not idempotent")
		}
	})
}
