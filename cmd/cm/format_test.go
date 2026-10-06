package main

import (
	"bytes"
	"os"
	"testing"
)

// TestFormatCLI проверяет контракт линтера и форматтера.
func TestFormatCLI(t *testing.T) {
	t.Parallel()

	path := t.TempDir() + "/model.cml"
	source := []byte(
		"BoundedContext ServiceIntegration{Aggregate A{Entity ServiceIntegration{aggregateRoot String id}}}",
	)

	if err := os.WriteFile(path, source, 0o600); err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		args []string
		code int
	}{
		{[]string{"format", "--check", path}, 1},
		{[]string{"lint", "--json", "-i", path}, 1},
		{[]string{"lint", "--fix", "-i", path}, 0},
		{[]string{"format", "--check", path}, 0},
		{[]string{"lint", "-i", path}, 0},
	} {
		var out, err bytes.Buffer

		if code := run(tc.args, &out, &err); code != tc.code {
			t.Fatalf("%v: %d %s %s", tc.args, code, &out, &err)
		}
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}

	if info.Mode().Perm() != 0o600 {
		t.Fatal("permissions changed")
	}
}

// TestFormatWriteRejectsInvalid проверяет контракт линтера и форматтера.
func TestFormatWriteRejectsInvalid(t *testing.T) {
	t.Parallel()

	path := t.TempDir() + "/model.cml"
	source := []byte("BoundedContext ServiceIntegration {")

	if err := os.WriteFile(path, source, 0o600); err != nil {
		t.Fatal(err)
	}

	var out, err bytes.Buffer

	if code := run([]string{"format", "-w", path}, &out, &err); code != 1 {
		t.Fatal(code)
	}

	//nolint:gosec // Файл создан этим тестом в t.TempDir.
	//nolint:gosec // Файл создан этим тестом в t.TempDir.
	actual, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}

	if !bytes.Equal(actual, source) {
		t.Fatal("invalid source changed")
	}
}
