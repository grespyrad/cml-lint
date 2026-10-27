package format

import (
	"strings"
	"unicode/utf8"

	"github.com/grespyrad/go-cml"
)

// writer применяет структурные границы и контекстные пробелы к исходным токенам.
type writer struct {
	data   []byte
	indent int
	column int
	angles int
}

func render(p layoutPlan) []byte {
	w := writer{}

	for i, token := range p.tokens {
		w.write(p, i, token)
	}

	if len(w.data) > 0 {
		w.data = append(w.data, '\n')
	}

	return w.data
}

func (w *writer) gap(lines int) {
	if len(w.data) == 0 || lines == 0 {
		return
	}

	for range lines {
		w.data = append(w.data, '\n')
	}

	w.column = 0
}

func (w *writer) text(text, separator string, continuation int) {
	if w.column == 0 {
		padding := strings.Repeat("  ", w.indent+continuation)

		w.data = append(w.data, padding...)
		w.column = len(padding)
	} else {
		w.data = append(w.data, separator...)

		w.column += len(separator)
	}

	w.data = append(w.data, text...)

	w.column += utf8.RuneCountInString(text)

	if _, after, ok := strings.CutLast(text, "\n"); ok {
		w.column = utf8.RuneCountInString(after)
	}
}

func (w *writer) write(p layoutPlan, index int, token cml.Lexeme) {
	previous := ""

	if index > 0 {
		previous = p.tokens[index-1].Text
	}

	if token.Text == "}" {
		w.indent--
	}

	lines := p.lineBreak(index)

	w.gap(lines)
	w.text(token.Text, w.space(previous, token.Text), p.continuation[index])

	switch token.Text {
	case "{":
		w.indent++
	case "<":
		w.angles++
	case ">":
		w.angles = max(0, w.angles-1)
	default:
	}
}

func (w *writer) space(previous, current string) string {
	if member(current, ", ; ) ] . :: ( <") || current == ">" && w.angles > 0 {
		return ""
	}

	if member(previous, "( [ . :: @ <") {
		return ""
	}

	if current == "*" && previous != "{" && previous != ";" {
		return ""
	}

	return " "
}

func (p *layoutPlan) lineBreak(index int) int {
	if index == 0 {
		return 0
	}

	token := p.tokens[index]
	previous := p.tokens[index-1]

	if token.Text == "}" || previous.Text == "{" {
		return 1
	}

	if token.Kind == commentKind {
		return p.commentBreak(index)
	}

	lines := p.breaks[index]

	if previous.Text == "}" || previous.Text == ";" || previous.Kind == commentKind {
		return max(lines, 1)
	}

	return lines
}

func (p *layoutPlan) commentBreak(index int) int {
	if p.trailing(index) {
		return 0
	}

	return max(p.breaks[index], 1)
}
