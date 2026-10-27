// Package format задаёт канонический layout CML с сохранением исходных токенов.
package format

import (
	"fmt"

	"github.com/grespyrad/go-cml"
)

// Source форматирует синтаксически валидный CML; строки и комментарии сохраняются.
func Source(file string, source []byte) ([]byte, error) {
	return SourceWithDialect(file, source, cml.Dialect{Aliases: nil, Original: false})
}

// SourceWithDialect форматирует выбранный диалект с сохранением raw keywords.
func SourceWithDialect(file string, source []byte, dialect cml.Dialect) ([]byte, error) {
	result := cml.ParseWithDialect(file, source, dialect)
	if !result.Valid {
		return nil, fmt.Errorf(
			"format CML at %d:%d: %s: %w",
			result.Diagnostics[0].Position.Line,
			result.Diagnostics[0].Position.Column,
			result.Diagnostics[0].Message,
			cml.ErrInvalidModel,
		)
	}

	tokens, err := cml.TokenizeWithDialect(file, source, dialect)
	if err != nil {
		return nil, fmt.Errorf("format CML: %w", err)
	}

	if balanceErr := validateDelimiters(tokens); balanceErr != nil {
		return nil, balanceErr
	}

	plan := planLayout(result.Model, tokens, dialect)
	formatted := render(plan)

	flatTokens, err := cml.TokenizeWithDialect(file, formatted, dialect)
	if err != nil {
		return nil, fmt.Errorf("tokenize pretty layout: %w", err)
	}

	plan.wrapLongLines(formatted, flatTokens)

	formatted = render(plan)

	check, err := cml.TokenizeWithDialect(file, formatted, dialect)
	if err != nil {
		return nil, fmt.Errorf("verify formatted CML: %w", err)
	}

	if len(check) != len(tokens) {
		return nil, fmt.Errorf("formatter changed token count: %w", cml.ErrInvalidModel)
	}

	for i := range tokens {
		if tokens[i].Text != check[i].Text {
			return nil, fmt.Errorf(
				"formatter changed token %d (%q -> %q): %w",
				i,
				tokens[i].Text,
				check[i].Text,
				cml.ErrInvalidModel,
			)
		}
	}

	if after := cml.ParseWithDialect(file, formatted, dialect); !after.Valid {
		return nil, fmt.Errorf("formatter produced invalid syntax: %w", cml.ErrInvalidModel)
	}

	return formatted, nil
}

// validateDelimiters защищает layout даже при слишком мягком принятии parser.
func validateDelimiters(tokens []cml.Lexeme) error {
	var stack []string

	closing := map[string]string{"}": "{", ")": "(", "]": "["}

	for _, token := range tokens {
		switch token.Text {
		case "{", "(", "[":
			stack = append(stack, token.Text)
		case "}", ")", "]":
			if len(stack) == 0 || stack[len(stack)-1] != closing[token.Text] {
				return fmt.Errorf("unbalanced delimiter %q at %d:%d: %w", token.Text,
					token.Position.Line, token.Position.Column, cml.ErrInvalidModel)
			}

			stack = stack[:len(stack)-1]
		default:
		}
	}

	if len(stack) != 0 {
		return fmt.Errorf("unclosed delimiter: %w", cml.ErrInvalidModel)
	}

	return nil
}
