package pratt

import (
	"fmt"

	"github.com/ouvrage-systems/ouvrage-calque-go/pkg/geometry"
)

// TokenType represents lexical token categories for Pratt parsing.
type TokenType int

const (
	TokenSymbol TokenType = iota
	TokenIdentifier
	TokenLiteral
	TokenParenOpen
	TokenParenClose
	TokenEOF
)

// Token represents an atomic token passed to the Pratt Engine.
type Token struct {
	Type  TokenType
	Value string
	Geo   geometry.Geometry
}

// TokenStream provides a peekable stream of tokens for Pratt parsing.
type TokenStream struct {
	tokens []Token
	cursor int
}

// NewTokenStream creates a new TokenStream.
func NewTokenStream(tokens []Token) *TokenStream {
	return &TokenStream{tokens: tokens, cursor: 0}
}

func (ts *TokenStream) Peek() Token {
	if ts.cursor >= len(ts.tokens) {
		return Token{Type: TokenEOF}
	}
	return ts.tokens[ts.cursor]
}

func (ts *TokenStream) Next() Token {
	tok := ts.Peek()
	if tok.Type != TokenEOF {
		ts.cursor++
	}
	return tok
}

// AST Nodes produced by the generic Pratt engine
type GenericLiteralNode struct {
	Value string
	Geo   geometry.Geometry
}

func (n *GenericLiteralNode) GetGeometry() geometry.Geometry { return n.Geo }

type GenericIdentifierNode struct {
	Name string
	Geo  geometry.Geometry
}

func (n *GenericIdentifierNode) GetGeometry() geometry.Geometry { return n.Geo }

type GenericBinaryNode struct {
	Left  Node
	Op    string
	Right Node
	Geo   geometry.Geometry
}

func (n *GenericBinaryNode) GetGeometry() geometry.Geometry { return n.Geo }

type GenericUnaryNode struct {
	Op   string
	Expr Node
	Geo  geometry.Geometry
}

func (n *GenericUnaryNode) GetGeometry() geometry.Geometry { return n.Geo }

type GenericListNode struct {
	Elements []Node
	Geo      geometry.Geometry
}

func (n *GenericListNode) GetGeometry() geometry.Geometry { return n.Geo }

type GenericCallNode struct {
	Function string
	Args     []Node
	Geo      geometry.Geometry
}

func (n *GenericCallNode) GetGeometry() geometry.Geometry { return n.Geo }

type InterpolatedStringNode struct {
	Parts []Node
	Geo   geometry.Geometry
}

func (n *InterpolatedStringNode) GetGeometry() geometry.Geometry { return n.Geo }

// Engine implements the generic Pratt Parser algorithm.
type Engine struct {
	Rules *RuleSet
}

// NewEngine creates a new Pratt Engine configured with a RuleSet.
func NewEngine(rules *RuleSet) *Engine {
	return &Engine{Rules: rules}
}

// Parse parses a stream of tokens into a generic AST Node using top-down operator precedence.
func (e *Engine) Parse(stream *TokenStream, precedence int) (Node, error) {
	tok := stream.Next()
	if tok.Type == TokenEOF {
		return nil, fmt.Errorf("unexpected end of expression")
	}

	left, err := e.parsePrefix(tok, stream)
	if err != nil {
		return nil, err
	}

	for {
		peekTok := stream.Peek()
		if peekTok.Type != TokenSymbol {
			break
		}

		rule, exists := e.Rules.GetRule(peekTok.Value)
		if !exists || rule.Precedence <= precedence {
			break
		}

		stream.Next() // Consume operator symbol

		nextPrecedence := rule.Precedence
		if rule.Associativity == RightAssoc {
			nextPrecedence--
		}

		right, err := e.Parse(stream, nextPrecedence)
		if err != nil {
			return nil, err
		}

		geo := geometry.NewGeometry(left.GetGeometry().StartLine, right.GetGeometry().EndLine, left.GetGeometry().Indent)
		left = &GenericBinaryNode{
			Left:  left,
			Op:    rule.Symbol,
			Right: right,
			Geo:   geo,
		}
	}

	return left, nil
}

func (e *Engine) parsePrefix(tok Token, stream *TokenStream) (Node, error) {
	switch tok.Type {
	case TokenLiteral:
		return &GenericLiteralNode{Value: tok.Value, Geo: tok.Geo}, nil
	case TokenIdentifier:
		peek := stream.Peek()
		if peek.Type == TokenParenOpen && peek.Value == "(" {
			stream.Next() // consume (
			var args []Node
			for {
				peekTok := stream.Peek()
				if peekTok.Type == TokenParenClose && peekTok.Value == ")" {
					stream.Next() // consume )
					break
				}
				if peekTok.Type == TokenEOF {
					return nil, fmt.Errorf("syntax error at L%d: unclosed function call '%s('", tok.Geo.StartLine, tok.Value)
				}
				arg, err := e.Parse(stream, 0)
				if err != nil {
					return nil, err
				}
				args = append(args, arg)
				commaOrClose := stream.Peek()
				if commaOrClose.Type == TokenSymbol && commaOrClose.Value == "," {
					stream.Next() // consume ,
				}
			}
			return &GenericCallNode{Function: tok.Value, Args: args, Geo: tok.Geo}, nil
		}
		return &GenericIdentifierNode{Name: tok.Value, Geo: tok.Geo}, nil
	case TokenParenOpen:
		openSym := tok.Value
		if openSym == "[" {
			// Parse list literal [...]
			var elements []Node
			for {
				peekTok := stream.Peek()
				if peekTok.Type == TokenParenClose && peekTok.Value == "]" {
					stream.Next() // consume ]
					break
				}
				if peekTok.Type == TokenEOF {
					return nil, fmt.Errorf("syntax error at L%d: unclosed list '['", tok.Geo.StartLine)
				}
				elem, err := e.Parse(stream, 0)
				if err != nil {
					return nil, err
				}
				elements = append(elements, elem)
				commaOrClose := stream.Peek()
				if commaOrClose.Type == TokenSymbol && commaOrClose.Value == "," {
					stream.Next() // consume ,
				}
			}
			return &GenericListNode{Elements: elements, Geo: tok.Geo}, nil
		}

		exprNode, err := e.Parse(stream, 0)
		if err != nil {
			return nil, err
		}
		closeTok := stream.Next()
		if closeTok.Type != TokenParenClose || closeTok.Value != ")" {
			return nil, fmt.Errorf("syntax error at L%d: expected ')' after sub-expression", tok.Geo.StartLine)
		}
		return exprNode, nil
	case TokenSymbol:
		if tok.Value == "INTERPOLATED_STRING" {
			var parts []Node
			for {
				peekTok := stream.Peek()
				if peekTok.Type == TokenSymbol && peekTok.Value == "END_INTERPOLATION" {
					stream.Next() // consume END_INTERPOLATION
					break
				}
				if peekTok.Type == TokenEOF {
					break
				}
				partNode, err := e.Parse(stream, 0)
				if err != nil {
					return nil, err
				}
				parts = append(parts, partNode)
			}
			return &InterpolatedStringNode{Parts: parts, Geo: tok.Geo}, nil
		}

		// Check unary operator (e.g. not, -)
		rule, exists := e.Rules.GetRule(tok.Value)
		if !exists {
			return nil, fmt.Errorf("syntax error at L%d: unexpected token '%s'", tok.Geo.StartLine, tok.Value)
		}
		subExpr, err := e.Parse(stream, rule.Precedence)
		if err != nil {
			return nil, err
		}
		return &GenericUnaryNode{
			Op:   tok.Value,
			Expr: subExpr,
			Geo:  tok.Geo,
		}, nil
	default:
		return nil, fmt.Errorf("syntax error at L%d: unexpected token '%s'", tok.Geo.StartLine, tok.Value)
	}
}
