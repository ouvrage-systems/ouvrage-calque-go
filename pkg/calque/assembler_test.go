package calque_test

import (
	"strings"
	"testing"

	"github.com/ouvrage-systems/ouvrage-calque-go/pkg/calque"
	"github.com/ouvrage-systems/ouvrage-calque-go/pkg/extractor"
)

func TestAssemblePolymorphicIfElse(t *testing.T) {
	yamlInput := `# services/auth-service.yaml

# @ocq:If(cond=(env.ENV_NAME == "production"))
- name: AUTH_URL
  # @ocq:Replace(with="  value: https://auth.company.com", when=(env.USE_HTTPS == "true"))
  value: http://auth-service:8080
# @ocq:Else
- name: AUTH_URL
  value: http://dev-auth:8080
# @ocq:EndIf
`

	extracted, err := extractor.Extract(strings.NewReader(yamlInput), "auth-service.yaml")
	if err != nil {
		t.Fatalf("Phase 1 Extractor failed: %v", err)
	}

	astNodes, err := calque.Assemble(extracted)
	if err != nil {
		t.Fatalf("Phase 3 Assembler failed: %v", err)
	}

	if len(astNodes) != 1 {
		t.Fatalf("Expected 1 root IfFrameNode, got %d", len(astNodes))
	}

	ifFrame, ok := astNodes[0].(*calque.IfFrameNode)
	if !ok {
		t.Fatalf("Expected root node to be *calque.IfFrameNode, got %T", astNodes[0])
	}

	if ifFrame.Geo.StartLine != 3 || ifFrame.Geo.EndLine != 10 {
		t.Errorf("Expected IfFrame boundary L3-L10, got %s", ifFrame.Geo)
	}

	// Verify branches: Branch 0 (If), Branch 1 (Else)
	if len(ifFrame.Branches) != 2 {
		t.Fatalf("Expected 2 branches inside IfFrameNode, got %d", len(ifFrame.Branches))
	}

	// Branch 0 (If): contains Replace action
	if ifFrame.Branches[0].Condition == nil {
		t.Errorf("Expected Branch 0 (If) to have condition AST")
	}
	if len(ifFrame.Branches[0].Nodes) != 1 {
		t.Fatalf("Expected 1 node in Branch 0, got %d", len(ifFrame.Branches[0].Nodes))
	}
	action1, ok := ifFrame.Branches[0].Nodes[0].(*calque.ActionNode)
	if !ok || action1.Name != "Replace" {
		t.Errorf("Expected node in Branch 0 to be ActionNode 'Replace', got %T", ifFrame.Branches[0].Nodes[0])
	}
	if _, hasWith := action1.Properties["with"]; !hasWith {
		t.Errorf("Expected ActionNode 'Replace' to have 'with' property in Properties map")
	}

	// Branch 1 (Else): condition is nil
	if ifFrame.Branches[1].Condition != nil {
		t.Errorf("Expected Branch 1 (Else) to have nil condition AST")
	}

	t.Logf("\nPhase 3 Assembly Success (Polymorphic IfFrameNode):\n  Root: %s\n  Branch 0 (If): Nodes=%d\n  Branch 1 (Else): Nodes=%d", ifFrame, len(ifFrame.Branches[0].Nodes), len(ifFrame.Branches[1].Nodes))
}

func TestAssemblePolymorphicMacro(t *testing.T) {
	yamlInput := `# services/auth-service.yaml

# @ocq:Macro(name="auth_env", args=[auth_url, auth_key], pipelines=[Replace(...)])
.template-01: &anchor01
  - name: AUTH_URL
# @ocq:EndMacro
`

	extracted, err := extractor.Extract(strings.NewReader(yamlInput), "auth-service.yaml")
	if err != nil {
		t.Fatalf("Phase 1 Extractor failed: %v", err)
	}

	astNodes, err := calque.Assemble(extracted)
	if err != nil {
		t.Fatalf("Phase 3 Assembler failed: %v", err)
	}

	if len(astNodes) != 1 {
		t.Fatalf("Expected 1 root MacroFrameNode, got %d", len(astNodes))
	}

	macroFrame, ok := astNodes[0].(*calque.MacroFrameNode)
	if !ok {
		t.Fatalf("Expected root node to be *calque.MacroFrameNode, got %T", astNodes[0])
	}

	if macroFrame.Name != "auth_env" {
		t.Errorf("Expected Macro name='auth_env', got '%s'", macroFrame.Name)
	}

	t.Logf("\nPhase 3 Assembly Success (Polymorphic MacroFrameNode):\n  Root: %s", macroFrame)
}
