// Command cm validates CML and generates PlantUML without Java.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	cml "github.com/grespyrad/CMLGo"
)

const version = "0.2.3"

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

//nolint:cyclop,gocognit,gocyclo,nestif // Layout грамматики проверен token conservation и idempotency corpus.
func runWithDialect(args []string, out, errOut io.Writer, dialect cml.Dialect) int {
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		return report(out, 0, "Usage: cm validate -i model.cml [--json]\n"+
			"       cm lint [--fix] -i model.cml\n"+
			"       cm format [-w|--check] model.cml\n"+
			"       cm generate -g plantuml -i model.cml -o directory\n"+
			"       cm --version")
	}

	if args[0] == "version" || args[0] == "--version" || args[0] == "-V" {
		return report(out, 0, "Context Mapper Go "+version)
	}

	command := args[0]
	if command == "format" || command == commandLint {
		return formatCommand(command, args[1:], out, errOut, dialect)
	}

	if command != "validate" && command != commandGenerate {
		return report(errOut, exitArguments, "Unknown command: "+command)
	}

	options, err := parseOptions(command, args[1:], errOut)
	if errors.Is(err, flag.ErrHelp) {
		return 0
	}

	if err != nil {
		return report(errOut, exitArguments, err.Error())
	}

	input, output := options.input, options.output
	asJSON := options.asJSON

	result := cml.ValidateFileWithDialect(input, dialect)

	if asJSON {
		if err := writeJSON(out, result); err != nil {
			return report(errOut, 1, err.Error())
		}
	} else {
		if err := printDiagnostics(result, out, errOut); err != nil {
			return report(errOut, 1, err.Error())
		}

		if result.Valid {
			if code := report(out, 0, "The CML file '"+input+"' has been validated without errors."); code != 0 {
				return code
			}
		}
	}

	if !result.Valid {
		return 1
	}

	if command == commandGenerate {
		return generateDiagrams(result, output, errOut)
	}

	return 0
}

func writeJSON(out io.Writer, result cml.Result) error {
	if err := json.NewEncoder(out).Encode(result); err != nil {
		return fmt.Errorf("write JSON diagnostics: %w", err)
	}

	return nil
}

const (
	exitArguments  = 2
	directoryMode  = 0o750
	outputFileMode = 0o600
)

type cliOptions struct {
	input, output, generator string
	asJSON                   bool
}

func parseOptions(command string, args []string, errOut io.Writer) (cliOptions, error) {
	var options cliOptions

	flags := flag.NewFlagSet(command, flag.ContinueOnError)
	flags.SetOutput(errOut)
	flags.StringVar(&options.input, "i", "", "CML input file")
	flags.StringVar(&options.input, "input", "", "CML input file")
	flags.BoolVar(&options.asJSON, "json", false, "JSON diagnostics")

	if command == commandGenerate {
		flags.StringVar(&options.output, "o", "", "output directory")
		flags.StringVar(&options.output, "outputDir", "", "output directory")
		flags.StringVar(&options.generator, "g", "", "generator (plantuml)")
		flags.StringVar(&options.generator, "generator", "", "generator (plantuml)")
	}

	if err := flags.Parse(args); err != nil {
		return options, fmt.Errorf("parse CLI arguments: %w", err)
	}

	if options.input == "" || flags.NArg() > 0 {
		return options, fmt.Errorf("required: -i/--input model.cml: %w", errArguments)
	}

	if !strings.HasSuffix(options.input, ".cml") {
		return options, fmt.Errorf("input must be a .cml file: %w", errArguments)
	}

	if command == commandGenerate && (options.generator != "plantuml" || options.output == "") {
		return options, fmt.Errorf("required: -g plantuml -o directory: %w", errArguments)
	}

	return options, nil
}

func generateDiagrams(result cml.Result, output string, errOut io.Writer) int {
	files, err := cml.PlantUML(result)
	if err != nil {
		return report(errOut, 1, err.Error())
	}

	if dirErr := os.MkdirAll(output, directoryMode); dirErr != nil {
		return report(errOut, 1, dirErr.Error())
	}

	for _, name := range cml.DiagramNames(files) {
		if writeErr := os.WriteFile(filepath.Join(output, name), []byte(files[name]), outputFileMode); writeErr != nil {
			return report(errOut, 1, writeErr.Error())
		}
	}

	return 0
}

var errArguments = errors.New("invalid arguments")

const (
	commandLint     = "lint"
	commandGenerate = "generate"
)
