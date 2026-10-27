package main

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/grespyrad/go-cml"
)

func formatBatch(inputs []string, options formatOptions, out, errOut io.Writer, dialect cml.Dialect) int {
	paths, issues := formatInputs(inputs)
	code := 0

	for _, issue := range issues {
		code = report(errOut, 1, issue.Error())
	}

	for _, path := range paths {
		code = max(code, formatFile("format", path, options, out, errOut, dialect))
	}

	return code
}

func formatInputs(inputs []string) ([]string, []error) {
	seen := map[string]bool{}

	var (
		paths  []string
		issues []error
	)

	add := func(path string) {
		absolute, err := filepath.Abs(path)
		if err != nil {
			issues = append(issues, fmt.Errorf("resolve CML path %q: %w", path, err))

			return
		}

		if !seen[absolute] {
			paths = append(paths, path)
			seen[absolute] = true
		}
	}

	for _, input := range inputs {
		//nolint:gosec // CLI намеренно проверяет явно заданные пользователем пути.
		info, err := os.Lstat(input)
		switch {
		case err != nil:
			issues = append(issues, fmt.Errorf("stat CML input %q: %w", input, err))
		case info.IsDir():
			issues = append(issues, walkCML(input, add)...)
		case !info.Mode().IsRegular():
			issues = append(issues, fmt.Errorf("CML input %q: %w", input, errRegularFile))
		case !strings.HasSuffix(input, ".cml"):
			issues = append(issues, fmt.Errorf("CML input %q: %w", input, errCMLExtension))
		default:
			add(input)
		}
	}

	return paths, issues
}

func walkCML(root string, add func(string)) []error {
	var issues []error

	//nolint:gosec // root явно задаётся пользователем CLI; symlinks не обходятся.
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			issues = append(issues, fmt.Errorf("walk CML path %q: %w", path, err))

			return nil
		}

		if entry.IsDir() {
			if excludedDirectory(root, path, entry.Name()) {
				return filepath.SkipDir
			}

			return nil
		}

		if regularCML(entry) {
			add(path)
		}

		return nil
	})
	if err != nil {
		issues = append(issues, fmt.Errorf("walk CML directory %q: %w", root, err))
	}

	return issues
}

var errCMLExtension = errors.New("input must be a .cml file")

func excludedDirectory(root, path, name string) bool {
	return path != root && slices.Contains([]string{".git", "vendor", "node_modules"}, name)
}

func regularCML(entry fs.DirEntry) bool {
	return entry.Type().IsRegular() && strings.HasSuffix(entry.Name(), ".cml")
}
