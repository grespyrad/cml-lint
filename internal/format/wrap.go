package format

import (
	"strings"
	"unicode/utf8"

	cml "github.com/grespyrad/CMLGo"
)

const preferredWidth = 100

// wrapLongLines переносит только разделители списков, не изменяя raw tokens.
func (p *layoutPlan) wrapLongLines(formatted []byte, tokens []cml.Lexeme) {
	lines := strings.Split(string(formatted), "\n")

	for start := 0; start < len(tokens); {
		end := start + 1
		line := tokens[start].Position.Line

		for end < len(tokens) && tokens[end].Position.Line == line {
			end++
		}

		if utf8.RuneCountInString(lines[line-1]) > preferredWidth {
			p.wrapLine(tokens, start, end)
		}

		start = end
	}
}

func (p *layoutPlan) wrapLine(tokens []cml.Lexeme, start, end int) {
	if p.wrapFlow(start, end) {
		return
	}

	scan := listScan{depth: listDepth{}, open: -1, separators: []int{}}

	for i := start; i < end; i++ {
		if scan.read(tokens[i].Text, i) {
			p.wrapParameters(scan.open, i, scan.separators)

			return
		}
	}

	if scan.open < 0 && len(scan.separators) > 0 {
		p.wrapList(tokens, end, scan.separators)
	}
}

type listScan struct {
	depth      listDepth
	open       int
	separators []int
}

func (s *listScan) read(text string, index int) bool {
	if s.depth.update(text) {
		return false
	}

	if s.depth.angles > 0 || s.depth.brackets > 0 {
		return false
	}

	switch text {
	case "(":
		s.open = index
		s.separators = s.separators[:0]
	case ")":
		if s.open >= 0 && index > s.open+1 {
			return true
		}

		s.open = -1
	case ",":
		s.separators = append(s.separators, index)
	default:
	}

	return false
}

func (p *layoutPlan) wrapParameters(open, closeIndex int, separators []int) {
	p.boundary(open+1, 1)
	p.boundary(closeIndex, 1)

	for _, index := range separators {
		p.boundary(index+1, 1)
	}

	for i := open + 1; i < closeIndex; i++ {
		p.continuation[i] = 1
	}
}

func (p *layoutPlan) wrapList(tokens []cml.Lexeme, end int, separators []int) {
	for _, index := range separators {
		next := index + 1
		if next < end && tokens[next].Kind == commentKind {
			next++
		}

		if next < end {
			p.boundary(next, 1)
		}
	}

	for i := separators[0] + 1; i < end; i++ {
		p.continuation[i] = 1
	}
}

func (p *layoutPlan) wrapFlow(start, end int) bool {
	first := -1

	for i := start; i < end; i++ {
		keyword := p.keyword(i)
		if keyword == "delegates to" || keyword == "emits event" {
			p.boundary(i, 1)

			if first < 0 {
				first = i
			}
		}
	}

	if first < 0 {
		return false
	}

	for i := first; i < end; i++ {
		p.continuation[i] = 1
	}

	return true
}

type listDepth struct {
	angles   int
	brackets int
}

func (d *listDepth) update(text string) bool {
	switch text {
	case "<":
		d.angles++
	case ">":
		d.angles = max(0, d.angles-1)
	case "[":
		d.brackets++
	case "]":
		d.brackets--
	default:
		return false
	}

	return true
}

func (p *layoutPlan) wrapCommentedParameters() {
	scan := listScan{depth: listDepth{}, open: -1, separators: []int{}}
	commented := false

	for index, token := range p.tokens {
		if scan.open >= 0 && token.Kind == commentKind {
			commented = true
		}

		if scan.read(token.Text, index) {
			if commented {
				p.wrapParameters(scan.open, index, scan.separators)
			}

			scan.open = -1
			scan.separators = scan.separators[:0]
			commented = false
		}
	}
}
