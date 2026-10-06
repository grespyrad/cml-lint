// Command cml-bench измеряет parser, recognizer и validator в памяти.
//
//nolint:cyclop // Средняя сложность отражает dispatch команд и grammar layout.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"runtime/pprof"
	"time"

	cml "github.com/grespyrad/CMLGo"
)

var errIterations = errors.New("iterations must be positive")

type measurement struct {
	Language   string `json:"language"`
	Bytes      int    `json:"bytes"`
	Iterations int    `json:"iterations"`
	NSPerOp    int64  `json:"ns_per_op"`
	SyntaxOnly bool   `json:"syntax_only"`
	Mode       string `json:"mode"`
}

func main() {
	if err := bench(); err != nil {
		if _, writeErr := fmt.Fprintln(os.Stderr, err); writeErr != nil {
			os.Exit(1)
		}

		os.Exit(1)
	}
}

//nolint:cyclop,funlen,gocognit,gocyclo // Layout грамматики проверен token conservation и idempotency corpus.
func bench() (resultErr error) {
	flags := flag.NewFlagSet("cml-bench", flag.ContinueOnError)
	input := flags.String("input", "testdata/reference/service.cml", "CML input")
	iterations := flags.Int("n", defaultIterations, "iterations")
	profile := flags.String("cpu", "", "CPU profile")
	syntax := flags.Bool("syntax", false, "recognition without AST")
	mode := flags.String("mode", "parse", "parse or validate")

	if err := flags.Parse(os.Args[1:]); err != nil {
		return fmt.Errorf("parse benchmark options: %w", err)
	}

	if *iterations < 1 {
		return errIterations
	}

	source, err := os.ReadFile(*input)
	if err != nil {
		return fmt.Errorf("read benchmark input: %w", err)
	}

	run := func() cml.Result {
		if *syntax {
			return cml.CheckSyntaxWithDialect(*input, source, cml.Dialect{Original: true, Aliases: nil})
		}

		if *mode == "validate" {
			return cml.Validate(*input, source)
		}

		return cml.Parse(*input, source)
	}
	if result := run(); !result.Valid {
		return fmt.Errorf("benchmark input: %v: %w", result.Diagnostics, cml.ErrInvalidModel)
	}

	if *profile != "" {
		file, err := os.Create(*profile)
		if err != nil {
			return fmt.Errorf("create CPU profile: %w", err)
		}

		defer func() {
			pprof.StopCPUProfile()

			if closeErr := file.Close(); closeErr != nil {
				resultErr = errors.Join(resultErr, fmt.Errorf("close CPU profile: %w", closeErr))
			}
		}()

		if err := pprof.StartCPUProfile(file); err != nil {
			return fmt.Errorf("start CPU profile: %w", err)
		}
	}

	started := time.Now()

	for range *iterations {
		if result := run(); !result.Valid {
			return fmt.Errorf("benchmark input: %v: %w", result.Diagnostics, cml.ErrInvalidModel)
		}
	}

	elapsed := time.Since(started)
	result := measurement{
		Language:   "go",
		Bytes:      len(source),
		Iterations: *iterations,
		NSPerOp:    elapsed.Nanoseconds() / int64(*iterations),
		SyntaxOnly: *syntax,
		Mode:       *mode,
	}

	if err := json.NewEncoder(os.Stdout).Encode(result); err != nil {
		return fmt.Errorf("write measurement: %w", err)
	}

	return nil
}

const defaultIterations = 200
