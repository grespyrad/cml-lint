package main

import (
	"fmt"
	"io"
	"strings"

	cml "github.com/grespyrad/CMLGo"
)

func report(out io.Writer, code int, message string) int {
	if _, err := fmt.Fprintln(out, message); err != nil {
		return 1
	}

	return code
}

func printDiagnostics(result cml.Result, out, errOut io.Writer) error {
	for _, d := range result.Diagnostics {
		destination := out

		if d.Severity == "error" {
			destination = errOut
		}

		if _, err := fmt.Fprintf(
			destination,
			"%s in %s on line %d:%d [%s]: %s\n",
			strings.ToUpper(d.Severity),
			d.Position.File,
			d.Position.Line,
			d.Position.Column,
			d.Code,
			d.Message,
		); err != nil {
			return fmt.Errorf("write diagnostics: %w", err)
		}
	}

	return nil
}
