package lexbench

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	cml "github.com/grespyrad/CMLGo"
)

// Все corpus проверяются до сравнения скорости; timing не является assertion.
func TestLexerParity(t *testing.T) {
	files := []string{}
	for _, folder := range []string{"upstream", "conformance", "escapes", "reference"} {
		err := filepath.WalkDir(
			filepath.Join("../../testdata", folder),
			func(path string, entry os.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if !entry.IsDir() && filepath.Ext(path) == ".cml" {
					files = append(files, path)
				}
				return nil
			},
		)
		if err != nil {
			t.Fatal(err)
		}
	}
	for _, path := range files {
		t.Run(path, func(t *testing.T) {
			source, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			assertParity(t, path, source)
		})
	}
}

func TestLexerEdges(t *testing.T) {
	for name, source := range map[string][]byte{
		"unicode":       []byte("// Сервис 🔧 文\nBoundedContext ServiceIntegration { domainVisionStatement 'Привет 🌍' }"),
		"prefix":        []byte("ContextMapSuffix ^ContextMap 123 // end"),
		"strings":       []byte(`"\u0414\uD83D\uDD27" '\n\t\'\"\\'`),
		"unicode-names": []byte("Сервис ContextMapСервис Сущность ^Сущность"),
		"empty":         {}, "unclosed": []byte(`"bad`), "comment": []byte("/* bad"),
		"escape": []byte(`"\x"`), "surrogate": []byte(`"\uD800"`),
		"low": []byte(`"\uDC00"`), "invalid-utf8": {0xff},
	} {
		t.Run(name, func(t *testing.T) { assertParity(t, name, source) })
	}
}

func assertParity(t *testing.T, name string, source []byte) {
	t.Helper()
	want, werr := cml.Tokenize(name, source)
	got, gerr := alternative(name, source)
	if (werr == nil) != (gerr == nil) {
		t.Fatalf("errors CMLGo=%v Participle=%v", werr, gerr)
	}
	if werr == nil && !reflect.DeepEqual(want, got) {
		if len(want) != len(got) {
			t.Fatalf("token counts %d / %d", len(want), len(got))
		}
		for i := range want {
			if !reflect.DeepEqual(want[i], got[i]) {
				t.Fatalf("token %d Go=%#v Participle=%#v", i, want[i], got[i])
			}
		}
	}
}

func BenchmarkStages(b *testing.B) {
	for _, name := range []string{"service-before", "service"} {
		path := "../../testdata/reference/" + name + ".cml"
		source, err := os.ReadFile(path)
		if err != nil {
			b.Fatal(err)
		}
		stages := map[string]func() error{
			"Tokenize/CMLGo":      func() error { _, err := cml.Tokenize(path, source); return err },
			"Tokenize/Participle": func() error { _, err := alternative(path, source); return err },
			"CheckSyntax": func() error {
				r := cml.CheckSyntax(path, source)
				if !r.Valid {
					return cml.ErrInvalidModel
				}
				return nil
			},
			"Parse": func() error {
				r := cml.Parse(path, source)
				if !r.Valid {
					return cml.ErrInvalidModel
				}
				return nil
			},
			"Validate": func() error {
				r := cml.Validate(path, source)
				if !r.Valid {
					return cml.ErrInvalidModel
				}
				return nil
			},
			"Unmarshal": func() error { var model cml.ContextMappingModel; return cml.Unmarshal(source, &model) },
		}
		for stage, run := range stages {
			b.Run(name+"/"+stage, func(b *testing.B) {
				if err := run(); err != nil {
					b.Fatal(err)
				}
				b.ReportAllocs()
				b.SetBytes(int64(len(source)))
				b.ResetTimer()
				for b.Loop() {
					if err := run(); err != nil {
						b.Fatal(err)
					}
				}
			})
		}
	}
}
