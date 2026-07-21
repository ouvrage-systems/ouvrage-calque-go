package expr

import "github.com/ouvrage-systems/ouvrage-calque-go/pkg/syntax/pratt"

// NewExpressionRuleSet returns the official Pratt RuleSet for Ocalque when/cond expressions.
func NewExpressionRuleSet() *pratt.RuleSet {
	rs := pratt.NewRuleSet("Ocalque Expression Sub-Grammar")
	rs.Register("or", 1, pratt.LeftAssoc, "LogicalOr")
	rs.Register("and", 2, pratt.LeftAssoc, "LogicalAnd")
	rs.Register("=", 3, pratt.RightAssoc, "Assignment")
	rs.Register("==", 4, pratt.LeftAssoc, "Equality")
	rs.Register("!=", 4, pratt.LeftAssoc, "Inequality")
	rs.Register(">=", 4, pratt.LeftAssoc, "GreaterOrEqual")
	rs.Register("<=", 4, pratt.LeftAssoc, "LessOrEqual")
	rs.Register(">", 4, pratt.LeftAssoc, "GreaterThan")
	rs.Register("<", 4, pratt.LeftAssoc, "LessThan")
	rs.Register("contains", 4, pratt.LeftAssoc, "Contains")
	rs.Register("in", 4, pratt.LeftAssoc, "InCollection")
	rs.Register("+", 6, pratt.LeftAssoc, "Addition")
	rs.Register("-", 6, pratt.LeftAssoc, "Subtraction")
	rs.Register("*", 7, pratt.LeftAssoc, "Multiplication")
	rs.Register("/", 7, pratt.LeftAssoc, "Division")
	rs.Register("%", 7, pratt.LeftAssoc, "Modulo")
	rs.Register("not", 8, pratt.RightAssoc, "LogicalNot")
	return rs
}
