package test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/ouvrage-systems/ouvrage-calque-go/pkg/calque"
	"github.com/ouvrage-systems/ouvrage-calque-go/pkg/extractor"
	"github.com/ouvrage-systems/ouvrage-calque-go/pkg/syntax/pratt"
)

// TestDemoNakedDSLDeploymentMaster demonstrates Phase 1, 2, 3 on the naked DSL lab file:
// architecture/labs/2026-07-19-decoupled-calque-dsl/deployment-master.ocq
func TestDemoNakedDSLDeploymentMaster(t *testing.T) {
	labFilePath := filepath.Join("..", "architecture", "labs", "2026-07-19-decoupled-calque-dsl", "deployment-master.ocq")

	f, err := os.Open(labFilePath)
	if err != nil {
		t.Fatalf("Failed to open naked DSL lab file: %v", err)
	}
	defer f.Close()

	t.Logf("\n================================================================================")
	t.Logf("     DEMO STEP-BY-STEP: NAKED DSL FILE (deployment-master.ocq)")
	t.Logf("================================================================================\n")

	// -------------------------------------------------------------------------
	// STEP 0: INPUT ASSET
	// -------------------------------------------------------------------------
	assetBytes, _ := os.ReadFile(labFilePath)
	t.Logf("--- INPUT ASSET (deployment-master.ocq) ---\n%s\n", string(assetBytes))

	// -------------------------------------------------------------------------
	// PHASE 1: EXTRACTION (NAKED DSL)
	// -------------------------------------------------------------------------
	f.Seek(0, 0)
	extracted, err := extractor.Extract(f, "deployment-master.ocq")
	if err != nil {
		t.Fatalf("Phase 1 Extractor failed: %v", err)
	}

	t.Logf("--- PHASE 1: EXTRACTED ANNOTATIONS (YAML LIST) ---")
	t.Logf("annotations:")
	for _, ann := range extracted {
		t.Logf("  - id: %q", ann.ID)
		t.Logf("    geo: { startLine: %d, endLine: %d, indent: %d }", ann.Geo.StartLine, ann.Geo.EndLine, ann.Geo.Indent)
		t.Logf("    rawContent: %q", ann.Content)
	}
	t.Logf("")

	// -------------------------------------------------------------------------
	// PHASE 2: DIRECTIVE PARSING
	// -------------------------------------------------------------------------
	strokes, err := calque.ParseDirectives(extracted)
	if err != nil {
		t.Fatalf("Phase 2 Directive Parser failed: %v", err)
	}

	t.Logf("--- PHASE 2: PARSED FLAT STROKES (YAML LIST) ---")
	t.Logf("strokes:")
	for _, stroke := range strokes {
		t.Logf("  - id: %q", stroke.ID)
		t.Logf("    name: %q", stroke.Name)
		t.Logf("    role: %q", stroke.Role)
		t.Logf("    geo: { startLine: %d, endLine: %d, indent: %d }", stroke.Geo.StartLine, stroke.Geo.EndLine, stroke.Geo.Indent)
		t.Logf("    argCount: %d", len(stroke.Args))
	}
	t.Logf("")

	// -------------------------------------------------------------------------
	// PHASE 3: STRUCTURAL ASSEMBLY
	// -------------------------------------------------------------------------
	astNodes, err := calque.Assemble(extracted)
	if err != nil {
		t.Fatalf("Phase 3 Assembler failed: %v", err)
	}

	t.Logf("--- PHASE 3: ASSEMBLED STRUCTURAL AST (YAML LIST) ---")
	t.Logf("astNodes:")
	for _, node := range astNodes {
		t.Logf("  - %s", renderNodeToYAMLElement(node, "    "))
	}

	t.Logf("================================================================================")
}

func renderNodeToYAMLElement(node calque.Node, indent string) string {
	var out string
	geo := node.GetGeometry()

	switch n := node.(type) {
	case *calque.IfFrameNode:
		out += fmt.Sprintf("type: IfFrameNode\n")
		out += fmt.Sprintf("%sid: %q\n", indent, n.ID)
		if geo.StartLine > 0 {
			out += fmt.Sprintf("%sgeo: { startLine: %d, endLine: %d }\n", indent, geo.StartLine, geo.EndLine)
		}
		out += fmt.Sprintf("%sproperties:\n", indent)
		out += renderPropertiesMap(n.GetProperties(), indent+"  ")

	case *calque.MacroFrameNode:
		out += fmt.Sprintf("type: MacroFrameNode\n")
		out += fmt.Sprintf("%sid: %q\n", indent, n.ID)
		if geo.StartLine > 0 {
			out += fmt.Sprintf("%sgeo: { startLine: %d, endLine: %d }\n", indent, geo.StartLine, geo.EndLine)
		}
		out += fmt.Sprintf("%sproperties:\n", indent)
		out += renderPropertiesMap(n.GetProperties(), indent+"  ")
		if len(n.Nodes) > 0 {
			out += fmt.Sprintf("%schildren:\n", indent)
			for _, child := range n.Nodes {
				out += fmt.Sprintf("%s  - %s", indent, renderNodeToYAMLElement(child, indent+"    "))
			}
		}

	case *calque.SectionFrameNode:
		out += fmt.Sprintf("type: SectionFrameNode\n")
		out += fmt.Sprintf("%sid: %q\n", indent, n.ID)
		if geo.StartLine > 0 {
			out += fmt.Sprintf("%sgeo: { startLine: %d, endLine: %d }\n", indent, geo.StartLine, geo.EndLine)
		}
		out += fmt.Sprintf("%sproperties:\n", indent)
		out += renderPropertiesMap(n.GetProperties(), indent+"  ")
		if len(n.Nodes) > 0 {
			out += fmt.Sprintf("%schildren:\n", indent)
			for _, child := range n.Nodes {
				out += fmt.Sprintf("%s  - %s", indent, renderNodeToYAMLElement(child, indent+"    "))
			}
		}

	case *calque.ActionNode:
		out += fmt.Sprintf("type: %sNode\n", n.Name)
		out += fmt.Sprintf("%sid: %q\n", indent, n.ID)
		if geo.StartLine > 0 {
			out += fmt.Sprintf("%sgeo: { startLine: %d, endLine: %d }\n", indent, geo.StartLine, geo.EndLine)
		}
		if len(n.Properties) > 0 {
			out += fmt.Sprintf("%sproperties:\n", indent)
			out += renderPropertiesMap(n.Properties, indent+"  ")
		}

	default:
		out += fmt.Sprintf("type: Unknown (%T)\n", node)
	}
	return out
}

func renderPropertiesMap(props map[string]any, indent string) string {
	var out string
	for k, v := range props {
		switch val := v.(type) {
		case calque.Pipeline:
			out += fmt.Sprintf("%s%s:\n", indent, k)
			for _, pNode := range val {
				out += fmt.Sprintf("%s  - %s", indent, renderNodeToYAMLElement(pNode, indent+"    "))
			}
		case pratt.Node:
			out += fmt.Sprintf("%s%s: %s\n", indent, k, renderPrattNodeBrief(val))
		case *calque.Condition:
			out += fmt.Sprintf("%s%s: %s\n", indent, k, renderConditionBrief(val))
		default:
			out += fmt.Sprintf("%s%s: %v\n", indent, k, val)
		}
	}
	return out
}

func renderConditionBrief(c *calque.Condition) string {
	if c == nil || c.Expr == nil {
		return "null"
	}
	return c.Expr.String()
}

func renderPrattNodeBrief(node pratt.Node) string {
	switch n := node.(type) {
	case *pratt.GenericIdentifierNode:
		return n.Name
	case *pratt.GenericLiteralNode:
		return fmt.Sprintf("%q", n.Value)
	case *pratt.InterpolatedStringNode:
		return fmt.Sprintf("%q", n.Parts)
	case *pratt.GenericCallNode:
		return fmt.Sprintf("CallExpr(%s)", n.Function)
	case *pratt.GenericListNode:
		return fmt.Sprintf("ListLiteral[%d elements]", len(n.Elements))
	case *pratt.GenericBinaryNode:
		return fmt.Sprintf("BinaryExpr(%s %s %s)", renderPrattNodeBrief(n.Left), n.Op, renderPrattNodeBrief(n.Right))
	default:
		if node != nil {
			return fmt.Sprintf("%v", node)
		}
		return "null"
	}
}
