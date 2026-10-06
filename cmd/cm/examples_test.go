package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// TestPublishedExamples проверяет validate, format, lint и diagrams для обоих языков имён.
func TestPublishedExamples(t *testing.T) {
	t.Parallel()

	for _, name := range []string{"service-ru", "service-en", "etalon-service", "etalon-service-neutral"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			path := "../../examples/" + name + ".cml"
			for _, args := range [][]string{
				{"validate", "-i", path},
				{"lint", "-i", path},
				{"format", "--check", path},
				{"generate", "-g", "plantuml", "-i", path, "-o", t.TempDir()},
			} {
				assertExampleCommand(t, args, 0)
			}

			assertExampleCommand(t, []string{"--original", "validate", "-i", path}, 1)
			assertExampleAtomicFormat(t, path)
		})
	}
}

func assertExampleCommand(t *testing.T, args []string, want int) {
	t.Helper()

	var out, errOut bytes.Buffer

	if code := run(args, &out, &errOut); code != want {
		t.Fatalf("%v: %d, want %d; %s %s", args, code, want, &out, &errOut)
	}
}

func assertExampleAtomicFormat(t *testing.T, sourcePath string) {
	t.Helper()

	//nolint:gosec // Путь выбран из фиксированного списка собственных examples.
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(t.TempDir(), "model.cml")
	//nolint:gosec // Путь создан через t.TempDir и фиксированное имя model.cml.
	if writeErr := os.WriteFile(path, source, 0o600); writeErr != nil {
		t.Fatal(writeErr)
	}

	assertExampleCommand(t, []string{"lint", "--fix", "-i", path}, 0)

	//nolint:gosec // Читается файл, созданный в t.TempDir.
	formatted, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(source, formatted) {
		t.Fatal("published example was not formatted canonically")
	}
}
