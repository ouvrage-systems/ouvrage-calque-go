package expr

import (
	"fmt"
	"strings"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"

	"github.com/ouvrage-systems/ouvrage-calque-go/pkg/syntax/pratt"
)

// EvalContext holds context variables represented as HashiCorp cty.Value types.
type EvalContext struct {
	Variables map[string]cty.Value
}

// NewEvalContext creates a new empty EvalContext.
func NewEvalContext() *EvalContext {
	return &EvalContext{
		Variables: make(map[string]cty.Value),
	}
}

// SetVariable binds a variable name to a cty.Value.
func (ctx *EvalContext) SetVariable(name string, val cty.Value) {
	ctx.Variables[name] = val
}

// GetVariable retrieves a variable by name.
func (ctx *EvalContext) GetVariable(name string) (cty.Value, bool) {
	val, exists := ctx.Variables[name]
	return val, exists
}

// Eval recursively evaluates a Pratt AST Node into a cty.Value.
func Eval(node pratt.Node, ctx *EvalContext) (cty.Value, error) {
	if node == nil {
		return cty.NilVal, fmt.Errorf("eval error: nil node")
	}

	switch n := node.(type) {
	case *pratt.GenericLiteralNode:
		return parseLiteralToCty(n.Value), nil

	case *pratt.GenericIdentifierNode:
		val, exists := ctx.GetVariable(n.Name)
		if !exists {
			return cty.NilVal, fmt.Errorf("eval error at L%d: unknown variable '%s'", n.Geo.StartLine, n.Name)
		}
		return val, nil

	case *pratt.InterpolatedStringNode:
		var str string
		for _, part := range n.Parts {
			v, err := Eval(part, ctx)
			if err != nil {
				return cty.NilVal, err
			}
			str += ctyValToString(v)
		}
		return cty.StringVal(str), nil

	case *pratt.GenericListNode:
		var vals []cty.Value
		for _, elem := range n.Elements {
			v, err := Eval(elem, ctx)
			if err != nil {
				return cty.NilVal, err
			}
			vals = append(vals, v)
		}
		if len(vals) == 0 {
			return cty.TupleVal([]cty.Value{}), nil
		}
		return cty.TupleVal(vals), nil

	case *pratt.GenericCallNode:
		return evalFunction(n.Function, n.Args, ctx)

	case *pratt.GenericUnaryNode:
		val, err := Eval(n.Expr, ctx)
		if err != nil {
			return cty.NilVal, err
		}
		if n.Op == "not" || n.Op == "!" {
			return cty.BoolVal(!val.True()), nil
		}
		return cty.NilVal, fmt.Errorf("eval error: unknown unary operator '%s'", n.Op)

	case *pratt.GenericBinaryNode:
		return evalBinaryNode(n, ctx)

	default:
		return cty.NilVal, fmt.Errorf("eval error: unhandled node type %T", node)
	}
}

func evalBinaryNode(n *pratt.GenericBinaryNode, ctx *EvalContext) (cty.Value, error) {
	// Handle short-circuiting for logical AND / OR
	if n.Op == "and" {
		left, err := Eval(n.Left, ctx)
		if err != nil {
			return cty.NilVal, err
		}
		if left.False() {
			return cty.False, nil // Short-circuit
		}
		right, err := Eval(n.Right, ctx)
		if err != nil {
			return cty.NilVal, err
		}
		return cty.BoolVal(right.True()), nil
	}

	if n.Op == "or" {
		left, err := Eval(n.Left, ctx)
		if err != nil {
			return cty.NilVal, err
		}
		if left.True() {
			return cty.True, nil // Short-circuit
		}
		right, err := Eval(n.Right, ctx)
		if err != nil {
			return cty.NilVal, err
		}
		return cty.BoolVal(right.True()), nil
	}

	left, err := Eval(n.Left, ctx)
	if err != nil {
		return cty.NilVal, err
	}
	right, err := Eval(n.Right, ctx)
	if err != nil {
		return cty.NilVal, err
	}

	switch n.Op {
	case "==":
		return left.Equals(right), nil
	case "!=":
		return left.Equals(right).Not(), nil
	case "+":
		if left.Type() == cty.Number && right.Type() == cty.Number {
			return left.Add(right), nil
		}
		return cty.StringVal(ctyValToString(left) + ctyValToString(right)), nil
	case "-":
		return left.Subtract(right), nil
	case "*":
		return left.Multiply(right), nil
	case "/":
		return left.Divide(right), nil
	case ">=":
		return left.GreaterThanOrEqualTo(right), nil
	case "<=":
		return left.LessThanOrEqualTo(right), nil
	case ">":
		return left.GreaterThan(right), nil
	case "<":
		return left.LessThan(right), nil
	case "contains":
		return cty.BoolVal(evalContains(left, right)), nil
	case "in":
		return cty.BoolVal(evalIn(left, right)), nil
	default:
		return cty.NilVal, fmt.Errorf("eval error: unsupported binary operator '%s'", n.Op)
	}
}

func evalFunction(fn string, args []pratt.Node, ctx *EvalContext) (cty.Value, error) {
	if fn == "len" && len(args) == 1 {
		val, err := Eval(args[0], ctx)
		if err != nil {
			return cty.NilVal, err
		}
		if val.Type() == cty.String {
			return cty.NumberIntVal(int64(len(val.AsString()))), nil
		}
		if val.Type().IsTupleType() {
			return cty.NumberIntVal(int64(val.LengthInt())), nil
		}
	}
	return cty.NilVal, fmt.Errorf("eval error: unknown function '%s'", fn)
}

func parseLiteralToCty(raw string) cty.Value {
	if raw == "true" {
		return cty.True
	}
	if raw == "false" {
		return cty.False
	}

	var num int64
	if _, err := fmt.Sscanf(raw, "%d", &num); err == nil {
		return cty.NumberIntVal(num)
	}

	return cty.StringVal(raw)
}

func ctyValToString(v cty.Value) string {
	if v.Type() == cty.String {
		return v.AsString()
	}
	if v.Type() == cty.Number {
		var num int64
		gocty.FromCtyValue(v, &num)
		return fmt.Sprintf("%d", num)
	}
	if v.Type() == cty.Bool {
		if v.True() {
			return "true"
		}
		return "false"
	}
	return ""
}

func evalContains(left, right cty.Value) bool {
	lStr := ctyValToString(left)
	rStr := ctyValToString(right)
	return strings.Contains(lStr, rStr)
}

func evalIn(item, collection cty.Value) bool {
	itemStr := ctyValToString(item)
	if collection.Type().IsTupleType() {
		it := collection.ElementIterator()
		for it.Next() {
			_, val := it.Element()
			if ctyValToString(val) == itemStr {
				return true
			}
		}
	}
	return false
}
