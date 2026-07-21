package calque_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ouvrage-systems/ouvrage-calque-go/pkg/calque"
	"github.com/ouvrage-systems/ouvrage-calque-go/pkg/extractor"
	"github.com/ouvrage-systems/ouvrage-calque-go/pkg/syntax/pratt"
)

func TestDemoPipelinePhase1AndPhase2(t *testing.T) {
	realAssetInput := `# services/payment-service.yaml

# @ocq:If(cond=(env.ENV_NAME == "production" or env.REGION in ["eu-west-1", "us-east-1"]))
- name: PAYMENT_GW
  # @ocq:Replace(with="  value: https://api.stripe.com/v1", when=(env.USE_STRIPE == "true"))
  value: http://dev-payment-gateway:8080
# @ocq:Else
- name: PAYMENT_GW
  value: http://mock-payment:9090
# @ocq:EndIf
`

	t.Logf("\n================================================================================")
	t.Logf("            OUVRAGE PIPELINE DEMO : PHASE 1 & PHASE 2 DETAILED VIEW")
	t.Logf("================================================================================\n")

	// -------------------------------------------------------------------------
	// PHASE 1: EXTRACTION (Geometrical Comment Stripping & Deterministic ID)
	// -------------------------------------------------------------------------
	extracted, err := extractor.Extract(strings.NewReader(realAssetInput), "payment-service.yaml")
	if err != nil {
		t.Fatalf("Phase 1 Extractor failed: %v", err)
	}

	t.Logf("--- PHASE 1: EXTRACTED ANNOTATIONS (FLAT GEOMETRICAL INDEX) ---")
	for i, ann := range extracted {
		t.Logf("  [%d] ID: %s | %s | Indent: %d", i+1, ann.ID, ann.Geo, ann.Geo.Indent)
		t.Logf("      Unwrapped Raw Directive: %q\n", ann.Content)
	}

	// -------------------------------------------------------------------------
	// PHASE 2: DIRECTIVE PARSING (Pratt Engine Argument Decomposition)
	// -------------------------------------------------------------------------
	strokes, err := calque.ParseDirectives(extracted)
	if err != nil {
		t.Fatalf("Phase 2 Directive Parser failed: %v", err)
	}

	t.Logf("--- PHASE 2: PARSED FLAT DIRECTIVE STROKES (WITH PRATT ARG ASTs) ---")
	for i, stroke := range strokes {
		t.Logf("  [%d] ID: %s | Name: %-8s | Role: %-8s | Geo: %s",
			i+1, stroke.ID, stroke.Name, stroke.Role, stroke.Geo)

		if len(stroke.Args) > 0 {
			t.Logf("      Parsed Arguments (Pratt AST):")
			for argName, argAST := range stroke.Args {
				treeVisual := printASCIITree(argAST, "        ", true)
				t.Logf("        Arg %q:\n%s", argName, treeVisual)
			}
		} else {
			t.Logf("      (No Arguments)\n")
		}
	}

	t.Logf("================================================================================")
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
	case *pratt.GenericIdentifierNode:
		result += fmt.Sprintf("Identifier: %s\n", n.Name)
	case *pratt.GenericLiteralNode:
		result += fmt.Sprintf("Literal: %q\n", n.Value)
	default:
		result += fmt.Sprintf("Node (%T)\n", n)
	}

	return result
}
