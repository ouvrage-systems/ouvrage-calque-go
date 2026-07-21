package expr

import (
	"strings"
	"unicode"

	"github.com/ouvrage-systems/ouvrage-calque-go/pkg/geometry"
	"github.com/ouvrage-systems/ouvrage-calque-go/pkg/syntax/pratt"
)

// TokenizeExpression tokenizes a raw string expression into a Pratt TokenStream.
func TokenizeExpression(input string, lineNum int) (*pratt.TokenStream, error) {
	var tokens []pratt.Token
	runes := []rune(input)
	i := 0
	length := len(runes)

	for i < length {
		ch := runes[i]

		// Skip whitespace
		if unicode.IsSpace(ch) {
			i++
			continue
		}

		geo := geometry.NewGeometry(lineNum, lineNum, 0)

		// Parentheses and Brackets
		if ch == '(' || ch == '[' {
			tokens = append(tokens, pratt.Token{Type: pratt.TokenParenOpen, Value: string(ch), Geo: geo})
			i++
			continue
		}
		if ch == ')' || ch == ']' {
			tokens = append(tokens, pratt.Token{Type: pratt.TokenParenClose, Value: string(ch), Geo: geo})
			i++
			continue
		}

		// String literals "..." or '...'
		if ch == '"' || ch == '\'' {
			quote := ch
			i++
			start := i
			for i < length && runes[i] != quote {
				i++
			}
			strVal := string(runes[start:i])
			if i < length {
				i++ // consume closing quote
			}

			if strings.Contains(strVal, "${") {
				tokens = append(tokens, pratt.Token{Type: pratt.TokenSymbol, Value: "INTERPOLATED_STRING", Geo: geo})
				// Push interpolated parts
				parts := parseInterpolatedParts(strVal, geo)
				tokens = append(tokens, parts...)
			} else {
				tokens = append(tokens, pratt.Token{Type: pratt.TokenLiteral, Value: strVal, Geo: geo})
			}
			continue
		}

		// Number literals 0-9
		if unicode.IsDigit(ch) {
			start := i
			for i < length && (unicode.IsDigit(runes[i]) || runes[i] == '.') {
				i++
			}
			numStr := string(runes[start:i])
			tokens = append(tokens, pratt.Token{Type: pratt.TokenLiteral, Value: numStr, Geo: geo})
			continue
		}

		// Multi-char symbols: >=, <=, ==, !=
		if i+1 < length {
			twoChar := string(runes[i : i+2])
			if twoChar == ">=" || twoChar == "<=" || twoChar == "==" || twoChar == "!=" {
				tokens = append(tokens, pratt.Token{Type: pratt.TokenSymbol, Value: twoChar, Geo: geo})
				i += 2
				continue
			}
		}

		// Single-char symbols: +, -, *, /, %, >, <, =, ,
		if strings.ContainsRune("+-*/%><=,", ch) {
			tokens = append(tokens, pratt.Token{Type: pratt.TokenSymbol, Value: string(ch), Geo: geo})
			i++
			continue
		}

		// Identifiers or word operators (and, or, not, contains, in)
		if unicode.IsLetter(ch) || ch == '_' || ch == '.' {
			start := i
			for i < length && (unicode.IsLetter(runes[i]) || unicode.IsDigit(runes[i]) || runes[i] == '_' || runes[i] == '.') {
				i++
			}
			word := string(runes[start:i])

			if word == "and" || word == "or" || word == "not" || word == "contains" || word == "in" {
				tokens = append(tokens, pratt.Token{Type: pratt.TokenSymbol, Value: word, Geo: geo})
			} else {
				tokens = append(tokens, pratt.Token{Type: pratt.TokenIdentifier, Value: word, Geo: geo})
			}
			continue
		}

		i++
	}

	return pratt.NewTokenStream(tokens), nil
}

func parseInterpolatedParts(raw string, geo geometry.Geometry) []pratt.Token {
	var parts []pratt.Token
	for len(raw) > 0 {
		idx := strings.Index(raw, "${")
		if idx == -1 {
			parts = append(parts, pratt.Token{Type: pratt.TokenLiteral, Value: raw, Geo: geo})
			break
		}
		if idx > 0 {
			parts = append(parts, pratt.Token{Type: pratt.TokenLiteral, Value: raw[:idx], Geo: geo})
		}
		raw = raw[idx+2:]
		endIdx := strings.IndexByte(raw, '}')
		if endIdx == -1 {
			break
		}
		varName := raw[:endIdx]
		parts = append(parts, pratt.Token{Type: pratt.TokenIdentifier, Value: varName, Geo: geo})
		raw = raw[endIdx+1:]
	}
	parts = append(parts, pratt.Token{Type: pratt.TokenSymbol, Value: "END_INTERPOLATION", Geo: geo})
	return parts
}
