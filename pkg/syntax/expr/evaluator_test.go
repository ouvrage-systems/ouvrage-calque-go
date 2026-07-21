package expr_test

import (
	"testing"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"

	"github.com/ouvrage-systems/ouvrage-calque-go/pkg/syntax/expr"
	"github.com/ouvrage-systems/ouvrage-calque-go/pkg/syntax/pratt"
)

func TestEvalMathAndLogic(t *testing.T) {
	inputStr := "8000 + (loop.index * 5) >= 8050"

	stream, _ := expr.TokenizeExpression(inputStr, 1)
	rules := expr.NewExpressionRuleSet()
	engine := pratt.NewEngine(rules)
	astNode, err := engine.Parse(stream, 0)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	ctx := expr.NewEvalContext()
	ctx.SetVariable("loop.index", cty.NumberIntVal(10))

	resVal, err := expr.Eval(astNode, ctx)
	if err != nil {
		t.Fatalf("Eval error: %v", err)
	}

	t.Logf("\n--------------------------------------------------------------------------------")
	t.Logf("EVALUATION TEST 1:")
	t.Logf("  Expression : %s", inputStr)
	t.Logf("  Context    : loop.index = 10")
	t.Logf("  Result     : %s (Type: %s)", resVal.GoString(), resVal.Type().FriendlyName())
	t.Logf("--------------------------------------------------------------------------------")

	if !resVal.True() {
		t.Errorf("Expected True evaluation for 8000 + (10 * 5) >= 8050, got False")
	}
}

func TestEvalStringInterpolation(t *testing.T) {
	inputStr := `"DEV_${index}.internal"`

	stream, _ := expr.TokenizeExpression(inputStr, 1)
	rules := expr.NewExpressionRuleSet()
	engine := pratt.NewEngine(rules)
	astNode, _ := engine.Parse(stream, 0)

	ctx := expr.NewEvalContext()
	ctx.SetVariable("index", cty.NumberIntVal(4))

	resVal, err := expr.Eval(astNode, ctx)
	if err != nil {
		t.Fatalf("Eval error: %v", err)
	}

	t.Logf("\n--------------------------------------------------------------------------------")
	t.Logf("EVALUATION TEST 2:")
	t.Logf("  Expression : %s", inputStr)
	t.Logf("  Context    : index = 4")
	t.Logf("  Result     : %q (Type: %s)", resVal.AsString(), resVal.Type().FriendlyName())
	t.Logf("--------------------------------------------------------------------------------")

	if resVal.AsString() != "DEV_4.internal" {
		t.Errorf("Expected 'DEV_4.internal', got '%s'", resVal.AsString())
	}
}

func TestEvalFunctionLen(t *testing.T) {
	inputStr := "len(fleet.rows) > 0"

	stream, _ := expr.TokenizeExpression(inputStr, 1)
	rules := expr.NewExpressionRuleSet()
	engine := pratt.NewEngine(rules)
	astNode, _ := engine.Parse(stream, 0)

	ctx := expr.NewEvalContext()
	ctx.SetVariable("fleet.rows", cty.TupleVal([]cty.Value{cty.StringVal("node1"), cty.StringVal("node2")}))

	resVal, err := expr.Eval(astNode, ctx)
	if err != nil {
		t.Fatalf("Eval error: %v", err)
	}

	t.Logf("\n--------------------------------------------------------------------------------")
	t.Logf("EVALUATION TEST 3:")
	t.Logf("  Expression : %s", inputStr)
	t.Logf("  Context    : fleet.rows = [\"node1\", \"node2\"]")
	t.Logf("  Result     : %s (Type: %s)", resVal.GoString(), resVal.Type().FriendlyName())
	t.Logf("--------------------------------------------------------------------------------")

	if !resVal.True() {
		t.Errorf("Expected len(fleet.rows) > 0 to evaluate to True")
	}
}

func TestEvalCalculator(t *testing.T) {
	// Pure arithmetic calculator expression without comparison operators
	inputStr := "1000 + (base_port * 2) - (50 / 2)"

	stream, _ := expr.TokenizeExpression(inputStr, 1)
	rules := expr.NewExpressionRuleSet()
	engine := pratt.NewEngine(rules)
	astNode, err := engine.Parse(stream, 0)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	ctx := expr.NewEvalContext()
	ctx.SetVariable("base_port", cty.NumberIntVal(4000)) // 1000 + (4000 * 2) - (50 / 2) = 1000 + 8000 - 25 = 8975

	resVal, err := expr.Eval(astNode, ctx)
	if err != nil {
		t.Fatalf("Eval error: %v", err)
	}

	t.Logf("\n--------------------------------------------------------------------------------")
	t.Logf("EVALUATION TEST 4 (PURE CALCULATOR):")
	t.Logf("  Expression : %s", inputStr)
	t.Logf("  Context    : base_port = 4000")
	t.Logf("  Result     : %s (Type: %s)", resVal.GoString(), resVal.Type().FriendlyName())
	t.Logf("--------------------------------------------------------------------------------")

	var resultNum int64
	gocty.FromCtyValue(resVal, &resultNum)
	if resultNum != 8975 {
		t.Errorf("Expected 8975, got %d", resultNum)
	}
}
