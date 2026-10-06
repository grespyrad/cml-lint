package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	cml "github.com/grespyrad/CMLGo"
)

//nolint:gocognit,gosec,nestif // Глобальные CLI options и доверенный локальный путь словаря проверены command tests.
func run(args []string, out, errOut io.Writer) int {
	dialect := cml.Dialect{Aliases: nil, Original: false}

	if len(args) > 0 && (args[0] == "--original" || args[0] == "--keywords") {
		flags := flag.NewFlagSet("cm", flag.ContinueOnError)
		flags.SetOutput(errOut)

		var dictionary string

		flags.BoolVar(&dialect.Original, "original", false, "strict original CML 6.12.0")
		flags.StringVar(&dictionary, "keywords", "", "JSON alias-to-canonical dictionary")

		if err := flags.Parse(args); err != nil {
			if errors.Is(err, flag.ErrHelp) {
				return 0
			}

			return report(errOut, exitArguments, err.Error())
		}

		if dictionary != "" {
			data, err := os.ReadFile(dictionary)
			if err != nil {
				return report(errOut, exitArguments, fmt.Sprint("Read keyword dictionary: ", err))
			}

			if decodeErr := json.Unmarshal(data, &dialect.Aliases); decodeErr != nil {
				return report(errOut, exitArguments, decodeErr.Error())
			}
		}

		if result := cml.CheckSyntaxWithDialect("dictionary.cml", nil, dialect); !result.Valid {
			return report(errOut, exitArguments, result.Diagnostics[0].Message)
		}

		args = flags.Args()
	}

	return runWithDialect(args, out, errOut, dialect)
}
