package expr_test

import (
	"fmt"
	"testing"

	"github.com/ouvrage-systems/ouvrage-calque-go/pkg/syntax/expr"
	"github.com/ouvrage-systems/ouvrage-calque-go/pkg/syntax/pratt"
)

// TestDemoComplexExpressionParsing runs a captivating demo-grade test on a production SRE condition.
func TestDemoComplexExpressionParsing(t *testing.T) {
	// A complex production-grade SRE condition string combining logic, math, lists, and string matching
	productionCondition := `(env.ENV_NAME == "production" or env.REGION in ["eu-west-1", "us-east-1"]) and (8000 + (loop.index * 5) >= 8050) and not (line.content contains "DEPRECATED")`

	t.Logf("\n================================================================================")
	t.Logf("                OCALQUE PRATT ENGINE : DEMO AST VISUALIZER")
	t.Logf("================================================================================")
	t.Logf("INPUT CONDITION STRING:\n  %s\n", productionCondition)

	// Step 1: Tokenize
	stream, err := expr.TokenizeExpression(productionCondition, 1)
	if err != nil {
		t.Fatalf("Tokenization error: %v", err)
	}

	// Step 2: Parse into Pratt AST using Ocalque Expression Ruleset
	rules := expr.NewExpressionRuleSet()
	engine := pratt.NewEngine(rules)

	astNode, err := engine.Parse(stream, 0)
	if err != nil {
		t.Fatalf("Pratt AST Parsing error: %v", err)
	}

	// Step 3: Render ASCII Tree Visualization
	treeStr := printASCIITree(astNode, "", true)
	t.Logf("PARSED PRATT AST TREE:\n%s", treeStr)

	// Step 4: Export Reflective Zero-Drift EBNF
	ebnfDoc := rules.ToEBNF()
	t.Logf("EXPORTED REFLECTIVE EBNF SPECIFICATION:\n%s", ebnfDoc)
	t.Logf("================================================================================")
}

func TestFunctionCallsAndInterpolation(t *testing.T) {
	// Testing function call len(fleet.rows) and string interpolation "DEV_${index}.internal"
	inputExpr := `len(fleet.rows) > 0 and node.name == "DEV_${index}.internal"`

	stream, err := expr.TokenizeExpression(inputExpr, 1)
	if err != nil {
		t.Fatalf("Tokenize failed: %v", err)
	}

	rules := expr.NewExpressionRuleSet()
	engine := pratt.NewEngine(rules)

	astNode, err := engine.Parse(stream, 0)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	treeStr := printASCIITree(astNode, "", true)
	t.Logf("\nPARSED INTERPOLATION & FUNCTION CALL AST TREE:\n%s", treeStr)
}

func printASCIITree(node pratt.Node, indent string, last bool) string {
	var result string
	prefix := indent
	if last {
		result += prefix + "└── "
		indent += "    "
	} else {
		result += prefix + "├── "
		indent += "│   "
	}

	switch n := node.(type) {
	case *pratt.GenericBinaryNode:
		result += fmt.Sprintf("BinaryExpr [ %s ]\n", n.Op)
		result += printASCIITree(n.Left, indent, false)
		result += printASCIITree(n.Right, indent, true)
	case *pratt.GenericUnaryNode:
		result += fmt.Sprintf("UnaryExpr [ %s ]\n", n.Op)
		result += printASCIITree(n.Expr, indent, true)
	case *pratt.GenericListNode:
		result += fmt.Sprintf("ListLiteral [ %d elements ]\n", len(n.Elements))
		for i, elem := range n.Elements {
			result += printASCIITree(elem, indent, i == len(n.Elements)-1)
		}
	case *pratt.GenericCallNode:
		result += fmt.Sprintf("CallExpr [ %s() ]\n", n.Function)
		for i, arg := range n.Args {
			result += printASCIITree(arg, indent, i == len(n.Args)-1)
		}
	case *pratt.InterpolatedStringNode:
		result += fmt.Sprintf("InterpolatedString [ %d parts ]\n", len(n.Parts))
		for i, part := range n.Parts {
			result += printASCIITree(part, indent, i == len(n.Parts)-1)
		}
	case *pratt.GenericIdentifierNode:
		result += fmt.Sprintf("Identifier: %s\n", n.Name)
	case *pratt.GenericLiteralNode:
		result += fmt.Sprintf("Literal: %q\n", n.Value)
	default:
		result += fmt.Sprintf("UnknownNode (%T)\n", n)
	}

	return result
}
