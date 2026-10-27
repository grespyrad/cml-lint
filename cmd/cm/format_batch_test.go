package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// TestFormatBatch проверяет несколько путей, рекурсивный обход и продолжение после ошибки.
func TestFormatBatch(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	first := writeBatchFixture(t, root, "first model.cml", "BoundedContext ServiceIntegration{}")
	second := writeBatchFixture(t, root, "nested/second.cml", "BoundedContext ExternalService{}")
	invalid := writeBatchFixture(t, root, "invalid.cml", "BoundedContext Missing{")
	ignored := writeBatchFixture(t, root, "node_modules/ignored.cml", "BoundedContext Ignored{}")
	writeBatchFixture(t, root, "notes.txt", "Plain text")

	for _, tc := range []struct {
		args []string
		code int
	}{
		{[]string{"format", "--check", first, second}, 1},
		{[]string{"format", "-w", first, second, first}, 0},
		{[]string{"format", "--check", first, second}, 0},
		{[]string{"format", "-w", root}, 1},
		{[]string{"format", "--check", first, second}, 0},
		{[]string{"format", "--check", invalid}, 1},
		{[]string{"format", first, second}, 2},
		{[]string{"lint", "--fix", first, second}, 2},
		{[]string{"format", "-w"}, 2},
	} {
		var out, errOut bytes.Buffer

		if got := run(tc.args, &out, &errOut); got != tc.code {
			t.Fatalf("%v: got %d, want %d: %s", tc.args, got, tc.code, &errOut)
		}
	}

	assertFileBytes(t, invalid, []byte("BoundedContext Missing{"))
	assertFileBytes(t, ignored, []byte("BoundedContext Ignored{}"))
}

// TestFormatDirectory проверяет пропуски и продолжение обработки при отсутствующем пути.
func TestFormatDirectory(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	first := writeBatchFixture(t, root, "model.cml", "BoundedContext ServiceIntegration{}")

	for _, directory := range []string{".git", "vendor", "node_modules"} {
		writeBatchFixture(t, root, directory+"/invalid.cml", "invalid")
	}

	link := filepath.Join(root, "linked.cml")
	if err := os.Symlink(first, link); err != nil {
		t.Fatal(err)
	}

	var out, errOut bytes.Buffer

	if code := run([]string{"format", "-w", root, filepath.Join(root, "missing.cml")}, &out, &errOut); code != 1 {
		t.Fatalf("code %d: %s", code, &errOut)
	}

	assertFileBytes(t, first, []byte("BoundedContext ServiceIntegration {\n}\n"))
	assertExampleCommand(t, []string{"format", "--check", root}, 0)
	assertExampleCommand(t, []string{"format", "-w", link}, 1)
	assertExampleCommand(t, []string{"format", "-w", filepath.Join(root, "vendor", "invalid.cml")}, 1)
}

func writeBatchFixture(t *testing.T, root, name, source string) string {
	t.Helper()

	path := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}

	return path
}

func assertFileBytes(t *testing.T, path string, want []byte) {
	t.Helper()

	//nolint:gosec // Читается fixture, созданный в t.TempDir.
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(got, want) {
		t.Fatalf("%s: want %q, got %q", path, want, got)
	}
}
