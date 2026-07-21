package extractor_test

import (
	"strings"
	"testing"

	"github.com/ouvrage-systems/ouvrage-calque-go/pkg/extractor"
)

func TestExtractYAMLAnnotations(t *testing.T) {
	yamlInput := `# services/auth-service.yaml

# @ocq:If(cond=(env.ENV_NAME == "production"))
- name: AUTH_URL
  value: https://auth.company.com
# @ocq:Else
- name: AUTH_URL
  value: http://auth-service:8080
# @ocq:EndIf
`

	annotations, err := extractor.Extract(strings.NewReader(yamlInput), "auth-service.yaml")
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if len(annotations) != 3 {
		t.Fatalf("Expected 3 extracted annotations, got %d", len(annotations))
	}

	// Annotation 1: If
	ann1 := annotations[0]
	if ann1.Geo.StartLine != 3 {
		t.Errorf("Expected StartLine 3, got %d", ann1.Geo.StartLine)
	}
	if ann1.Content != `If(cond=(env.ENV_NAME == "production"))` {
		t.Errorf("Unexpected content: %s", ann1.Content)
	}
	if ann1.ID == "" {
		t.Errorf("Expected non-empty hash ID")
	}

	// Annotation 2: Else
	ann2 := annotations[1]
	if ann2.Geo.StartLine != 6 {
		t.Errorf("Expected StartLine 6, got %d", ann2.Geo.StartLine)
	}
	if ann2.Content != "Else" {
		t.Errorf("Unexpected content: %s", ann2.Content)
	}

	// Annotation 3: EndIf
	ann3 := annotations[2]
	if ann3.Geo.StartLine != 9 {
		t.Errorf("Expected StartLine 9, got %d", ann3.Geo.StartLine)
	}
	if ann3.Content != "EndIf" {
		t.Errorf("Unexpected content: %s", ann3.Content)
	}
}

func TestExtractBatchAnnotations(t *testing.T) {
	batchInput := `@echo off
:: @ocq:Config(comment_style="::")

:: @ocq:Replace(with="SET COMPILER_PATH=C:\MSBuild\bin", when=(env.ENV_NAME == "production"))
SET COMPILER_PATH=C:\DevCompiler\bin
`

	annotations, err := extractor.Extract(strings.NewReader(batchInput), "build.bat")
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if len(annotations) != 2 {
		t.Fatalf("Expected 2 extracted annotations, got %d", len(annotations))
	}

	if annotations[0].Content != `Config(comment_style="::")` {
		t.Errorf("Unexpected content: %s", annotations[0].Content)
	}

	if annotations[1].Geo.StartLine != 4 {
		t.Errorf("Expected line 4, got %d", annotations[1].Geo.StartLine)
	}
}
