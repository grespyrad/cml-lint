package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestExtendedCLI проверяет интеграцию диалектов и инвариантов через command boundary.
//
//nolint:cyclop,gocognit,gocyclo,gosec // Command integration читает только созданные в t.TempDir fixtures.
func TestExtendedCLI(t *testing.T) {
	t.Parallel()

	directory := t.TempDir()
	path := filepath.Join(directory, "model.cml")
	source := []byte(
		`ОграниченныйКонтекст Сервис {Агрегат Запросы {Invariant SingleOwner {
 rule "Один writer" testedBy "example.contracts.TestOwner" implementedBy "example.requests.Handle"
}}}`,
	)

	if err := os.WriteFile(path, source, 0o600); err != nil {
		t.Fatal(err)
	}

	for _, command := range [][]string{{"validate", "-i", path}, {"lint", "--fix", "-i", path}} {
		var out, errOut bytes.Buffer

		if code := run(command, &out, &errOut); code != 0 {
			t.Fatalf("%v: %d %s %s", command, code, &out, &errOut)
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(data), "ОграниченныйКонтекст") || !strings.Contains(string(data), "\n      rule") {
		t.Fatal("keys lost or rule not formatted")
	}

	var out, errOut bytes.Buffer

	if code := run([]string{"--original", "validate", "-i", path}, &out, &errOut); code != 1 {
		t.Fatalf("original code %d: %s", code, &errOut)
	}

	dictionary := filepath.Join(directory, "keywords.json")
	if writeErr := os.WriteFile(dictionary, []byte(`{"Contexto":"BoundedContext"}`), 0o600); writeErr != nil {
		t.Fatal(writeErr)
	}

	if writeErr := os.WriteFile(path, []byte("Contexto ServiceIntegration"), 0o600); writeErr != nil {
		t.Fatal(writeErr)
	}

	out.Reset()
	errOut.Reset()

	if code := run([]string{"--keywords", dictionary, "validate", "-i", path}, &out, &errOut); code != 0 {
		t.Fatalf("dictionary code %d: %s", code, &errOut)
	}
}
