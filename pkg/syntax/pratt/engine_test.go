package pratt_test

import (
	"testing"

	"github.com/ouvrage-systems/ouvrage-calque-go/pkg/geometry"
	"github.com/ouvrage-systems/ouvrage-calque-go/pkg/syntax/pratt"
)

func buildExprRuleSet() *pratt.RuleSet {
	rs := pratt.NewRuleSet("Expression Grammar")
	rs.Register("or", 1, pratt.LeftAssoc, "LogicalOr")
	rs.Register("and", 2, pratt.LeftAssoc, "LogicalAnd")
	rs.Register("==", 4, pratt.LeftAssoc, "Equality")
	rs.Register("contains", 4, pratt.LeftAssoc, "Contains")
	rs.Register("+", 6, pratt.LeftAssoc, "Addition")
	rs.Register("*", 7, pratt.LeftAssoc, "Multiplication")
	rs.Register("not", 8, pratt.RightAssoc, "LogicalNot")
	return rs
}

func TestPrattEnginePrecedence(t *testing.T) {
	rules := buildExprRuleSet()
	engine := pratt.NewEngine(rules)

	// Expression: 8000 + index * 5
	// Tokens: 8000 (+), index (*), 5
	tokens := []pratt.Token{
		{Type: pratt.TokenLiteral, Value: "8000", Geo: geometry.NewGeometry(1, 1, 0)},
		{Type: pratt.TokenSymbol, Value: "+", Geo: geometry.NewGeometry(1, 1, 0)},
		{Type: pratt.TokenIdentifier, Value: "index", Geo: geometry.NewGeometry(1, 1, 0)},
		{Type: pratt.TokenSymbol, Value: "*", Geo: geometry.NewGeometry(1, 1, 0)},
		{Type: pratt.TokenLiteral, Value: "5", Geo: geometry.NewGeometry(1, 1, 0)},
	}

	stream := pratt.NewTokenStream(tokens)
	astNode, err := engine.Parse(stream, 0)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	binaryNode, ok := astNode.(*pratt.GenericBinaryNode)
	if !ok {
		t.Fatalf("Expected root node to be GenericBinaryNode")
	}

	// Top operator should be "+" (precedence 6)
	if binaryNode.Op != "+" {
		t.Errorf("Expected root operator '+', got '%s'", binaryNode.Op)
	}

	// Right child should be "*" (precedence 7)
	rightBinary, ok := binaryNode.Right.(*pratt.GenericBinaryNode)
	if !ok {
		t.Fatalf("Expected right child to be GenericBinaryNode (*)")
	}

	if rightBinary.Op != "*" {
		t.Errorf("Expected right operator '*', got '%s'", rightBinary.Op)
	}
}

func TestPrattEngineParentheses(t *testing.T) {
	rules := buildExprRuleSet()
	engine := pratt.NewEngine(rules)

	// Expression: (8000 + index) * 5
	tokens := []pratt.Token{
		{Type: pratt.TokenParenOpen, Value: "(", Geo: geometry.NewGeometry(1, 1, 0)},
		{Type: pratt.TokenLiteral, Value: "8000", Geo: geometry.NewGeometry(1, 1, 0)},
		{Type: pratt.TokenSymbol, Value: "+", Geo: geometry.NewGeometry(1, 1, 0)},
		{Type: pratt.TokenIdentifier, Value: "index", Geo: geometry.NewGeometry(1, 1, 0)},
		{Type: pratt.TokenParenClose, Value: ")", Geo: geometry.NewGeometry(1, 1, 0)},
		{Type: pratt.TokenSymbol, Value: "*", Geo: geometry.NewGeometry(1, 1, 0)},
		{Type: pratt.TokenLiteral, Value: "5", Geo: geometry.NewGeometry(1, 1, 0)},
	}

	stream := pratt.NewTokenStream(tokens)
	astNode, err := engine.Parse(stream, 0)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	binaryNode, ok := astNode.(*pratt.GenericBinaryNode)
	if !ok {
		t.Fatalf("Expected root node to be GenericBinaryNode")
	}

	// Top operator should now be "*" because parentheses override precedence
	if binaryNode.Op != "*" {
		t.Errorf("Expected root operator '*', got '%s'", binaryNode.Op)
	}
}

func TestEBNFGeneration(t *testing.T) {
	rules := buildExprRuleSet()
	ebnfDoc := rules.ToEBNF()

	if ebnfDoc == "" {
		t.Fatalf("Expected non-empty EBNF document")
	}

	t.Logf("Generated EBNF Document:\n%s", ebnfDoc)
}
