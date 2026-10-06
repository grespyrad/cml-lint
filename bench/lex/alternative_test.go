package lexbench

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/alecthomas/participle/v2/lexer"
	cml "github.com/grespyrad/CMLGo"
)

var (
	definitionOnce sync.Once
	definition     *lexer.StatefulDefinition
	definitionErr  error
	symbols        map[lexer.TokenType]string
	keywords       map[string]bool
)

// Адаптер существует только в research module. Правила получены из одного
// snapshot; проверка escapes обязательна даже при возврате исходного текста.
func alternative(file string, source []byte) ([]cml.Lexeme, error) {
	if !utf8.Valid(source) || len(source) > 8<<20 {
		return nil, fmt.Errorf("invalid source")
	}
	definitionOnce.Do(buildDefinition)
	if definitionErr != nil {
		return nil, definitionErr
	}
	scanner, err := definition.LexString(file, string(source))
	if err != nil {
		return nil, err
	}
	out := make([]cml.Lexeme, 0)
	for {
		token, err := scanner.Next()
		if err != nil {
			return nil, err
		}
		if token.Type == lexer.EOF {
			break
		}
		kind := symbols[token.Type]
		if kind == "id" && keywords[token.Value] {
			kind = "lit"
		}
		if (kind == "other" && (token.Value == "'" || token.Value == "\"")) ||
			(kind != "comment" && strings.HasPrefix(string(source[token.Pos.Offset:]), "/*")) {
			return nil, fmt.Errorf("unterminated literal")
		}
		if kind == "string" {
			if err := checkString(token.Value); err != nil {
				return nil, err
			}
		}
		out = append(
			out,
			cml.Lexeme{
				Kind: kind,
				Text: token.Value,
				Position: cml.Position{
					File:   file,
					Line:   token.Pos.Line,
					Column: token.Pos.Column,
					Offset: token.Pos.Offset,
				},
			},
		)
	}
	return out, nil
}

func buildDefinition() {
	data, err := os.ReadFile("../grammar/grammar.json")
	if err != nil {
		definitionErr = err
		return
	}
	var tree any
	if err = json.Unmarshal(data, &tree); err != nil {
		definitionErr = err
		return
	}
	unique := map[string]bool{}
	var walk func(any)
	walk = func(value any) {
		switch value := value.(type) {
		case map[string]any:
			if value["op"] == "lit" {
				unique[value["text"].(string)] = true
			}
			for _, child := range value {
				walk(child)
			}
		case []any:
			for _, child := range value {
				walk(child)
			}
		}
	}
	walk(tree)
	literals := make([]string, 0, len(unique))
	for literal := range unique {
		literals = append(literals, literal)
	}
	slices.SortFunc(literals, func(a, b string) int {
		if len(a) != len(b) {
			return len(b) - len(a)
		}
		return strings.Compare(a, b)
	})
	patterns := make([]string, 0, len(literals))
	keywords = map[string]bool{}
	word := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z_0-9]*$`)
	for alias := range cml.RussianAliases() {
		keywords[alias] = true
	}
	for _, literal := range literals {
		if word.MatchString(literal) {
			keywords[literal] = true
			continue
		}
		pattern := regexp.QuoteMeta(literal)
		last := literal[len(literal)-1]
		if last == '_' || last >= 'A' && last <= 'Z' || last >= 'a' && last <= 'z' || last >= '0' && last <= '9' {
			pattern += `\b`
		}
		patterns = append(patterns, pattern)
	}
	definition, definitionErr = lexer.NewSimple([]lexer.SimpleRule{
		{Name: "whitespace", Pattern: `[ \t\r\n\f]+`},
		{Name: "Comment", Pattern: `//[^\n]*|/\*(?s:.*?)\*/`},
		{Name: "String", Pattern: `"(?:[^"\\]|\\[\s\S])*"|'(?:[^'\\]|\\[\s\S])*'`},
		{Name: "Escaped", Pattern: `\^[\p{L}_][\p{L}\p{Nd}\p{M}_]*`},
		{Name: "Literal", Pattern: strings.Join(patterns, "|")},
		{Name: "Ident", Pattern: `[\p{L}_][\p{L}\p{Nd}\p{M}_]*`},
		{Name: "Integer", Pattern: `[0-9]+`},
		{Name: "Other", Pattern: `[\s\S]`},
	})
	if definitionErr != nil {
		return
	}
	symbols = map[lexer.TokenType]string{}
	for name, kind := range map[string]string{"Comment": "comment", "String": "string", "Escaped": "id", "Literal": "lit", "Ident": "id", "Integer": "int", "Other": "other"} {
		symbols[definition.Symbols()[name]] = kind
	}
}

func checkString(text string) error {
	// Декодирование включено в timed work; контракт Tokenize возвращает raw.
	decoded := make([]byte, 0)
	for i := 1; i < len(text)-1; i++ {
		if text[i] != '\\' {
			decoded = append(decoded, text[i])
			continue
		}
		i++
		switch text[i] {
		case 'u':
			unit, err := strconv.ParseUint(text[i+1:min(i+5, len(text)-1)], 16, 16)
			if err != nil || i+5 > len(text)-1 {
				return fmt.Errorf("invalid unicode escape")
			}
			i += 4
			if unit >= 0xd800 && unit <= 0xdbff {
				if i+6 >= len(text)-1 || text[i+1:i+3] != `\u` {
					return fmt.Errorf("missing surrogate")
				}
				low, err := strconv.ParseUint(text[i+3:i+7], 16, 16)
				if err != nil || low < 0xdc00 || low > 0xdfff {
					return fmt.Errorf("invalid surrogate")
				}
				unit = 0x10000 + (unit-0xd800)*0x400 + (low - 0xdc00)
				i += 6
			} else if unit >= 0xdc00 && unit <= 0xdfff {
				return fmt.Errorf("unpaired surrogate")
			}
			decoded = utf8.AppendRune(decoded, rune(unit))
		case 'n', 'r', 't', 'b', 'f', '\\', '\'', '"':
			decoded = append(decoded, text[i])
		default:
			return fmt.Errorf("invalid escape")
		}
	}
	_ = string(decoded)
	return nil
}
