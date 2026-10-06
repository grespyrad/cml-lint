//nolint:cyclop // Средняя сложность отражает dispatch команд и grammar layout.
package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

// TestCLI проверяет контракт линтера и форматтера.
func TestCLI(t *testing.T) {
	t.Parallel()

	cases := []struct {
		args     []string
		code     int
		contains string
	}{
		{[]string{"validate", "-i", "../../testdata/extensions/invariants.cml"}, 0, "validated without errors"},
		{[]string{"validate", "-i", "../../testdata/reference/service.cml"}, 0, "validated without errors"},
		{[]string{"validate", "--input", "../../testdata/reference/service.cml", "--json"}, 0, "\"valid\":true"},
		{[]string{"validate", "-i", "missing.cml"}, 1, "ERROR"},
		{[]string{"validate"}, 2, "input"},
		{[]string{"validate", "-h"}, 0, "Usage"},
		{[]string{"--version"}, 0, "0.2.3"},
		{[]string{"unknown"}, 2, "Unknown command"},
	}
	for _, tc := range cases {
		t.Run(strings.Join(tc.args, " "), func(t *testing.T) {
			t.Parallel()

			var out, err bytes.Buffer

			code := run(tc.args, &out, &err)

			if code != tc.code || !strings.Contains(out.String()+err.String(), tc.contains) {
				t.Fatalf("code=%d output=%s %s", code, &out, &err)
			}
		})
	}
}

// TestJSONFailure проверяет контракт линтера и форматтера.
func TestJSONFailure(t *testing.T) {
	t.Parallel()

	var out, err bytes.Buffer

	if code := run([]string{"validate", "--json", "-i", "missing.cml"}, &out, &err); code != 1 {
		t.Fatal(code)
	}

	var result struct {
		Valid       bool  `json:"valid"`
		Diagnostics []any `json:"diagnostics"`
	}

	if e := json.Unmarshal(out.Bytes(), &result); e != nil || result.Valid || len(result.Diagnostics) == 0 {
		t.Fatalf("%s", &out)
	}
}

// TestGenerate проверяет контракт линтера и форматтера.
func TestGenerate(t *testing.T) {
	t.Parallel()

	var out, err bytes.Buffer

	if code := run(
		[]string{"generate", "-g", "plantuml", "-i", "../../testdata/reference/service.cml", "-o", t.TempDir()},
		&out,
		&err,
	); code != 0 {
		t.Fatalf("%d %s", code, &err)
	}
}
