package pratt

import (
	"fmt"

	"github.com/ouvrage-systems/ouvrage-calque-go/pkg/geometry"
)

// Associativity defines operator associativity.
type Associativity int

const (
	LeftAssoc Associativity = iota
	RightAssoc
)

// OperatorRule defines the binding power and behavior of a grammar symbol.
type OperatorRule struct {
	Symbol        string        // e.g. "==", "and", "+", "contains"
	Precedence    int           // Binding power (higher value = tighter binding)
	Associativity Associativity // Left or Right associativity
	Description   string        // Human description for EBNF generation & errors
}

// RuleSet represents a configurable grammar rule table for the Pratt Engine.
type RuleSet struct {
	Name      string
	Operators map[string]OperatorRule
}

// NewRuleSet creates a new empty RuleSet.
func NewRuleSet(name string) *RuleSet {
	return &RuleSet{
		Name:      name,
		Operators: make(map[string]OperatorRule),
	}
}

// Register adds an operator rule to the RuleSet.
func (rs *RuleSet) Register(symbol string, precedence int, assoc Associativity, desc string) {
	rs.Operators[symbol] = OperatorRule{
		Symbol:        symbol,
		Precedence:    precedence,
		Associativity: assoc,
		Description:   desc,
	}
}

// GetRule retrieves a rule by its symbol.
func (rs *RuleSet) GetRule(symbol string) (OperatorRule, bool) {
	rule, exists := rs.Operators[symbol]
	return rule, exists
}

// ToEBNF exports the RuleSet as a formal EBNF grammar document.
func (rs *RuleSet) ToEBNF() string {
	var ebnf string
	ebnf += fmt.Sprintf("(* Formal EBNF Specification for %s *)\n\n", rs.Name)

	for symbol, rule := range rs.Operators {
		ebnf += fmt.Sprintf("%-15s ::= Operand '%s' Operand (* prec:%d, %s *) ;\n",
			rule.Description, symbol, rule.Precedence, assocToString(rule.Associativity))
	}
	return ebnf
}

func assocToString(a Associativity) string {
	if a == LeftAssoc {
		return "left-associative"
	}
	return "right-associative"
}

// Node is the base interface for all nodes produced by the Pratt Engine.
type Node interface {
	GetGeometry() geometry.Geometry
}
