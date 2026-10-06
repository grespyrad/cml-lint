//nolint:cyclop // Layout grammar и AST boundaries проверяются corpus, dialect и fuzz tests.
package format

import cml "github.com/grespyrad/CMLGo"

//nolint:cyclop,gocognit,gocyclo // AST показывает property boundaries; обход не меняет lexer tokens.
func collectInvariantBoundaries(node *cml.Node, boundaries map[int]bool, tokens []cml.Lexeme) {
	indexes := map[int]int{}

	for index, token := range tokens {
		indexes[token.Position.Offset] = index
	}

	var visit func(*cml.Node)

	visit = func(node *cml.Node) {
		if node.Kind == "Invariant" {
			for _, field := range []string{"rule", "rationale", "testedBy", "implementedBy"} {
				values := node.Fields[field]
				if len(values) == 0 {
					continue
				}

				index := indexes[values[0].Position.Offset] - 1
				if index >= 0 && tokens[index].Text == "=" {
					index--
				}

				if index >= 0 {
					boundaries[tokens[index].Position.Offset] = true
				}
			}
		}

		for _, values := range node.Fields {
			for _, value := range values {
				if value.Node != nil {
					visit(value.Node)
				}
			}
		}
	}
	visit(node)
}
