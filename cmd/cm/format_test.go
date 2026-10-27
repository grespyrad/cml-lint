package main

import (
	"bytes"
	"os"
	"slices"
	"strings"
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

// TestPrettyCLI проверяет одинаковый точный результат stdout, -w и lint --fix.
func TestPrettyCLI(t *testing.T) {
	t.Parallel()

	source := "BoundedContext ServiceIntegration{Aggregate Requests{Entity Request{aggregateRoot String id}}}"
	want := []byte("BoundedContext ServiceIntegration {\n  Aggregate Requests {\n" +
		"    Entity Request {\n      aggregateRoot\n      String id\n    }\n  }\n}\n")

	for _, command := range [][]string{{"format"}, {"format", "-w", "--"}, {"lint", "--fix"}} {
		t.Run(command[0]+strings.Join(command[1:], ""), func(t *testing.T) {
			t.Parallel()

			assertPrettyCommand(t, command, source, want)
		})
	}
}

func assertPrettyCommand(t *testing.T, command []string, source string, want []byte) {
	t.Helper()

	path := writeBatchFixture(t, t.TempDir(), "model.cml", source)

	var out, errOut bytes.Buffer

	args := append(slices.Clone(command), path)

	if code := run(args, &out, &errOut); code != 0 {
		t.Fatalf("code %d: %s", code, &errOut)
	}

	if len(command) == 1 {
		if !bytes.Equal(out.Bytes(), want) {
			t.Fatalf("stdout: %s", &out)
		}

		assertFileBytes(t, path, []byte(source))
	} else {
		assertFileBytes(t, path, want)
	}
}
