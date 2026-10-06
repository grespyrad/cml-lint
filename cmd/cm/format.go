package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	cml "github.com/grespyrad/CMLGo"

	"github.com/grespyrad/cml-lint/internal/format"
)

//nolint:gosec,cyclop,funlen,gocognit,gocyclo,nestif // Локальные CLI paths доверены; поведение проверяется тестами.
func formatCommand(command string, args []string, out, errOut io.Writer, dialect cml.Dialect) int {
	flags := flag.NewFlagSet(command, flag.ContinueOnError)
	flags.SetOutput(errOut)

	var (
		input                         string
		write, check, fix, jsonOutput bool
	)

	flags.StringVar(&input, "i", "", "CML input")
	flags.StringVar(&input, "input", "", "CML input")
	flags.BoolVar(&write, "w", false, "update file atomically")
	flags.BoolVar(&check, "check", false, "check formatting")
	flags.BoolVar(&fix, "fix", false, "format before lint")
	flags.BoolVar(&jsonOutput, "json", false, "JSON diagnostics")

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}

		return exitArguments
	}

	if input == "" && flags.NArg() == 1 {
		input = flags.Arg(0)
	} else if flags.NArg() > 0 {
		return report(errOut, exitArguments, "Provide exactly one CML input")
	}

	if input == "" || write && check || command == "format" && (fix || jsonOutput) ||
		command == commandLint && (write || check) {
		return report(errOut, exitArguments, "Invalid arguments: provide input and a compatible mode")
	}

	if !strings.HasSuffix(input, ".cml") {
		return report(errOut, 1, "ERROR: input must be a .cml file")
	}

	source, err := os.ReadFile(input)
	if err != nil {
		return report(errOut, 1, fmt.Sprint("Read CML:", err))
	}

	formatted, err := format.SourceWithDialect(input, source, dialect)
	if err != nil {
		return report(errOut, 1, err.Error())
	}

	changed := !bytes.Equal(source, formatted)
	if (write || fix) && changed {
		if err := writeAtomic(input, source, formatted); err != nil {
			return report(errOut, 1, err.Error())
		}

		changed = false
	}

	if command == commandLint {
		result := cml.ValidateFileWithDialect(input, dialect)

		if changed {
			result.Valid = false
			result.Diagnostics = append(
				result.Diagnostics,
				cml.Diagnostic{
					Severity: "error",
					Code:     "formatting",
					Message:  "CML requires formatting; run format -w or lint --fix",
					Position: cml.Position{File: input, Line: 1, Column: 1, Offset: 0},
				},
			)
		}

		if jsonOutput {
			if err := writeJSON(out, result); err != nil {
				return report(errOut, 1, err.Error())
			}
		} else {
			if err := printDiagnostics(result, out, errOut); err != nil {
				return report(errOut, 1, err.Error())
			}
		}

		if !result.Valid {
			return 1
		}

		return 0
	}

	if check {
		if changed {
			return report(errOut, 1, fmt.Sprint("CML requires formatting:", input))
		}

		return 0
	}

	if !write {
		if _, err := out.Write(formatted); err != nil {
			return report(errOut, 1, fmt.Sprint("Write formatted CML:", err))
		}
	}

	return 0
}

//nolint:gosec,cyclop,gocognit,gocyclo // Локальные CLI paths доверены; поведение проверяется тестами.
func writeAtomic(path string, original, data []byte) (resultErr error) {
	info, err := os.Lstat(path)
	if err != nil {
		return fmt.Errorf("stat CML: %w", err)
	}

	if !info.Mode().IsRegular() {
		return errRegularFile
	}

	file, err := os.CreateTemp(filepath.Dir(path), ".cml-format-*")
	if err != nil {
		return fmt.Errorf("create formatted temporary file: %w", err)
	}

	defer func() {
		if cleanupErr := os.Remove(file.Name()); cleanupErr != nil && !errors.Is(cleanupErr, os.ErrNotExist) {
			resultErr = errors.Join(resultErr, fmt.Errorf("cleanup formatted temporary file: %w", cleanupErr))
		}
	}()

	if _, writeErr := file.Write(data); writeErr != nil {
		return fmt.Errorf("write formatted CML: %w", errors.Join(writeErr, file.Close()))
	}

	if chmodErr := file.Chmod(info.Mode().Perm()); chmodErr != nil {
		return fmt.Errorf("preserve CML permissions: %w", errors.Join(chmodErr, file.Close()))
	}

	if closeErr := file.Close(); closeErr != nil {
		return fmt.Errorf("close formatted CML: %w", closeErr)
	}

	current, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reread CML: %w", err)
	}

	if !bytes.Equal(current, original) {
		return errConcurrentChange
	}

	if renameErr := os.Rename(file.Name(), path); renameErr != nil {
		return fmt.Errorf("replace formatted CML: %w", renameErr)
	}

	return nil
}

var (
	errRegularFile      = errors.New("format -w requires a regular file")
	errConcurrentChange = errors.New("CML changed during formatting")
)
