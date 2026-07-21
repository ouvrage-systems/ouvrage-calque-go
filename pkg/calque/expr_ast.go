package calque

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ouvrage-systems/ouvrage-calque-go/pkg/syntax/pratt"
	"github.com/zclconf/go-cty/cty"
)

// OpType represents structured comparison & logical operators in Phase 3 Compiled Expression AST.
type OpType string

const (
	OpEq       OpType = "eq"
	OpNeq      OpType = "neq"
	OpGt       OpType = "gt"
	OpGte      OpType = "gte"
	OpLt       OpType = "lt"
	OpLte      OpType = "lte"
	OpContains OpType = "contains"
	OpIn       OpType = "in"
	OpAnd      OpType = "and"
	OpOr       OpType = "or"
	OpNot      OpType = "not"
)

// Expr is the universal interface for strongly typed, executable domain expression nodes in Phase 3 AST.
type Expr interface {
	ToMap() map[string]any
	String() string
}

// -----------------------------------------------------------------------------
// 1. COMPARISON EXPRESSION: CompareExpr
// -----------------------------------------------------------------------------

type CompareExpr struct {
	Op    OpType `json:"op" yaml:"op"`
	Left  Expr   `json:"left" yaml:"left"`
	Right Expr   `json:"right" yaml:"right"`
}

func (e *CompareExpr) ToMap() map[string]any {
	return map[string]any{
		"_expr": e.String(),
		"op":    string(e.Op),
		"left":  e.Left.ToMap(),
		"right": e.Right.ToMap(),
	}
}

func (e *CompareExpr) String() string {
	return fmt.Sprintf("%s %s %s", e.Left, e.Op, e.Right)
}

// -----------------------------------------------------------------------------
// 2. LOGICAL EXPRESSION: LogicalExpr (AND / OR)
// -----------------------------------------------------------------------------

type LogicalExpr struct {
	Op       OpType `json:"op" yaml:"op"`
	Operands []Expr `json:"operands" yaml:"operands"`
}

func (e *LogicalExpr) ToMap() map[string]any {
	var ops []any
	for _, op := range e.Operands {
		ops = append(ops, op.ToMap())
	}
	return map[string]any{
		"_expr":    e.String(),
		"op":       string(e.Op),
		"operands": ops,
	}
}

func (e *LogicalExpr) String() string {
	var parts []string
	for _, op := range e.Operands {
		parts = append(parts, op.String())
	}
	return fmt.Sprintf("(%s)", strings.Join(parts, fmt.Sprintf(" %s ", e.Op)))
}

// -----------------------------------------------------------------------------
// 3. UNARY NOT EXPRESSION: NotExpr
// -----------------------------------------------------------------------------

type NotExpr struct {
	Operand Expr `json:"operand" yaml:"operand"`
}

func (e *NotExpr) ToMap() map[string]any {
	return map[string]any{
		"_expr":   e.String(),
		"op":      "not",
		"operand": e.Operand.ToMap(),
	}
}
func (e *NotExpr) String() string {
	return fmt.Sprintf("not (%s)", e.Operand)
}

// -----------------------------------------------------------------------------
// 4. VARIABLE REFERENCE EXPRESSION: VarRefExpr
// -----------------------------------------------------------------------------

type VarRefExpr struct {
	Path []string `json:"path" yaml:"path"`
}

func (e *VarRefExpr) ToMap() map[string]any {
	return map[string]any{
		"var": strings.Join(e.Path, "."),
	}
}
func (e *VarRefExpr) String() string {
	return strings.Join(e.Path, ".")
}

// -----------------------------------------------------------------------------
// 5. LITERAL EXPRESSION: LiteralExpr
// -----------------------------------------------------------------------------

type LiteralExpr struct {
	Value cty.Value `json:"value" yaml:"value"`
}

func (e *LiteralExpr) ToMap() map[string]any {
	if e.Value.Type() == cty.String {
		return map[string]any{"literal": e.Value.AsString()}
	}
	if e.Value.Type() == cty.Number {
		bf := e.Value.AsBigFloat()
		f64, _ := bf.Float64()
		return map[string]any{"literal": f64}
	}
	if e.Value.Type() == cty.Bool {
		return map[string]any{"literal": e.Value.True()}
	}
	return map[string]any{"literal": fmt.Sprintf("%v", e.Value)}
}
func (e *LiteralExpr) String() string {
	if e.Value.Type() == cty.String {
		return fmt.Sprintf("%q", e.Value.AsString())
	}
	return fmt.Sprintf("%v", e.Value)
}

// -----------------------------------------------------------------------------
// 6. FUNCTION CALL EXPRESSION: CallExpr
// -----------------------------------------------------------------------------

type CallExpr struct {
	Function string `json:"function" yaml:"function"`
	Args     []Expr `json:"args" yaml:"args"`
}

func (e *CallExpr) ToMap() map[string]any {
	var argsList []any
	for _, a := range e.Args {
		argsList = append(argsList, a.ToMap())
	}
	return map[string]any{
		"call": e.Function,
		"args": argsList,
	}
}
func (e *CallExpr) String() string {
	var args []string
	for _, a := range e.Args {
		args = append(args, a.String())
	}
	return fmt.Sprintf("%s(%s)", e.Function, strings.Join(args, ", "))
}

// -----------------------------------------------------------------------------
// COMPILER CONVERTER: Pratt AST -> Domain Expr AST
// -----------------------------------------------------------------------------

// CompilePrattToDomainExpr converts a raw Pratt AST node into a strongly typed Domain Expr AST.
func CompilePrattToDomainExpr(node pratt.Node) Expr {
	if node == nil {
		return nil
	}

	switch n := node.(type) {
	case *pratt.GenericIdentifierNode:
		return &VarRefExpr{Path: strings.Split(n.Name, ".")}

	case *pratt.GenericLiteralNode:
		rawVal := strings.Trim(n.Value, "\"")
		if rawVal == "true" {
			return &LiteralExpr{Value: cty.True}
		}
		if rawVal == "false" {
			return &LiteralExpr{Value: cty.False}
		}
		if num, err := strconv.ParseFloat(rawVal, 64); err == nil {
			return &LiteralExpr{Value: cty.NumberFloatVal(num)}
		}
		return &LiteralExpr{Value: cty.StringVal(rawVal)}

	case *pratt.GenericBinaryNode:
		op := mapPrattOperator(n.Op)
		left := CompilePrattToDomainExpr(n.Left)
		right := CompilePrattToDomainExpr(n.Right)

		if op == OpAnd || op == OpOr {
			return &LogicalExpr{Op: op, Operands: []Expr{left, right}}
		}

		return &CompareExpr{Op: op, Left: left, Right: right}

	case *pratt.GenericCallNode:
		var args []Expr
		for _, a := range n.Args {
			args = append(args, CompilePrattToDomainExpr(a))
		}
		return &CallExpr{Function: n.Function, Args: args}

	default:
		return &LiteralExpr{Value: cty.StringVal(fmt.Sprintf("%v", node))}
	}
}

func mapPrattOperator(op string) OpType {
	switch op {
	case "==":
		return OpEq
	case "!=":
		return OpNeq
	case ">":
		return OpGt
	case ">=":
		return OpGte
	case "<":
		return OpLt
	case "<=":
		return OpLte
	case "contains":
		return OpContains
	case "in":
		return OpIn
	case "and":
		return OpAnd
	case "or":
		return OpOr
	case "not":
		return OpNot
	default:
		return OpEq
	}
}
