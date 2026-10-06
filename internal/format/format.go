// Package format задаёт канонический layout CML с сохранением исходных токенов.
//
//nolint:cyclop // Средняя сложность отражает dispatch команд и grammar layout.
package format

import (
	"fmt"
	"slices"
	"strings"

	cml "github.com/grespyrad/CMLGo"
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

	boundaries := map[int]bool{}
	collectBoundaries(result.Model, boundaries)
	collectInvariantBoundaries(result.Model, boundaries, tokens)

	layout := writer{data: nil, indent: 0, lineStart: false, boundaries: boundaries, compact: false}

	for i, token := range tokens {
		layout.write(token, i, tokens)
	}

	layout.newline()

	formatted := layout.data

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

func collectBoundaries(n *cml.Node, boundaries map[int]bool) {
	if n == nil {
		return
	}

	if !member(
		n.Kind,
		"ComplexType Parameter StateTransition StateTransitionTarget TargetState EitherCommandOrOperation",
	) &&
		n.Kind != "ContextMappingModel" &&
		n.Kind != "NormalFeature" &&
		n.Kind != "StoryFeature" {
		boundaries[n.Position.Offset] = true
	}

	for _, values := range n.Fields {
		for _, value := range values {
			if value.Node != nil {
				collectBoundaries(value.Node, boundaries)
			}
		}
	}
}

type writer struct {
	data       []byte
	indent     int
	lineStart  bool
	boundaries map[int]bool
	compact    bool
}

func (w *writer) newline() {
	data := w.data
	if len(data) > 0 && data[len(data)-1] != '\n' {
		w.data = append(w.data, '\n')
	}

	w.lineStart = true
	w.compact = false
}

func (w *writer) text(text string) {
	if len(w.data) == 0 || w.lineStart {
		w.data = append(w.data, strings.Repeat("  ", w.indent)...)

		w.lineStart = false
	} else if !w.compact {
		w.data = append(w.data, ' ')
	}

	w.data = append(w.data, text...)

	w.compact = false
}

//nolint:cyclop,funlen,gocognit,gocyclo // Layout грамматики проверен token conservation и idempotency corpus.
func (w *writer) write(token cml.Lexeme, index int, tokens []cml.Lexeme) {
	if token.Kind == "comment" {
		if index == 0 || tokens[index-1].Position.Line != token.Position.Line {
			w.newline()
		}

		w.text(token.Text)
		w.newline()

		return
	}

	if w.boundaries[token.Position.Offset] {
		w.newline()
	}

	switch token.Text {
	case "{":
		w.text("{")
		w.newline()

		w.indent++
	case "}":
		w.newline()

		if w.indent > 0 {
			w.indent--
		}

		w.text("}")
		w.newline()
	case ";":
		w.compact = true
		w.text(";")
		w.newline()
	case ",":
		w.compact = true
		w.text(",")
	case "(":
		w.compact = true
		w.text("(")

		w.compact = true
	case ")":
		w.compact = true
		w.text(")")
	case ".", "::":
		w.compact = true
		w.text(token.Text)

		w.compact = true
	default:
		if token.Kind == "lit" &&
			member(token.Text, newlineProperties) {
			w.newline()
		}

		w.text(token.Text)
	}
}

func member(value, list string) bool { return slices.Contains(strings.Fields(list), value) }

const (
	noNewlineKinds = "ComplexType Parameter StateTransition StateTransitionTarget " +
		"TargetState EitherCommandOrOperation"
	newlineProperties = "contains type state domainVisionStatement responsibilities " +
		"implementationTechnology knowledgeLevel owner aggregateRoot"
)
