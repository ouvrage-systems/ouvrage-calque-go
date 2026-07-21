package test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ouvrage-systems/ouvrage-calque-go/pkg/calque"
	"github.com/ouvrage-systems/ouvrage-calque-go/pkg/extractor"
	"github.com/ouvrage-systems/ouvrage-calque-go/pkg/geometry"
	"github.com/ouvrage-systems/ouvrage-calque-go/pkg/syntax/pratt"
	"gopkg.in/yaml.v3"
)

// YAML DTOs for Phase 1, Phase 2, Phase 3 autonomous serialization

type YAMLAnnotation struct {
	ID         string            `yaml:"id"`
	Geo        geometry.Geometry `yaml:"geo"`
	RawContent string            `yaml:"rawContent"`
}

type YAMLStroke struct {
	ID        string            `yaml:"id"`
	Name      string            `yaml:"name"`
	Role      string            `yaml:"role"`
	Geo       geometry.Geometry `yaml:"geo"`
	PrattTree any               `yaml:"prattTree,omitempty"`
}

type YAMLNode struct {
	Type       string         `yaml:"type"`
	ID         string         `yaml:"id"`
	Geo        *YAMLGeo       `yaml:"geo,omitempty"`
	Properties map[string]any `yaml:"properties,omitempty"`
	Children   []*YAMLNode    `yaml:"children,omitempty"`
}

type YAMLGeo struct {
	StartLine int `yaml:"startLine"`
	EndLine   int `yaml:"endLine"`
}

// TestExportRealYAMLFiles exports Phase 1, Phase 2, and Phase 3 directly as REAL autonomous .yaml files in test/output/
func TestExportRealYAMLFiles(t *testing.T) {
	labFilePath := filepath.Join("..", "architecture", "labs", "2026-07-19-decoupled-calque-dsl", "deployment-master.ocq")

	f, err := os.Open(labFilePath)
	if err != nil {
		t.Fatalf("Failed to open naked DSL lab file: %v", err)
	}
	defer f.Close()

	outputDir := filepath.Join(".", "output")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}

	// -------------------------------------------------------------------------
	// PHASE 1: EXTRACTION -> phase1_extracted_annotations.yaml
	// -------------------------------------------------------------------------
	f.Seek(0, 0)
	extracted, err := extractor.Extract(f, "deployment-master.ocq")
	if err != nil {
		t.Fatalf("Phase 1 Extractor failed: %v", err)
	}

	var yamlAnnList []YAMLAnnotation
	for _, ann := range extracted {
		yamlAnnList = append(yamlAnnList, YAMLAnnotation{
			ID:         ann.ID,
			Geo:        ann.Geo,
			RawContent: ann.Content,
		})
	}

	p1File := filepath.Join(outputDir, "phase1_extracted_annotations.yaml")
	p1Bytes, _ := marshalPhase1YAML(yamlAnnList)
	if err := os.WriteFile(p1File, p1Bytes, 0644); err != nil {
		t.Fatalf("Failed to write phase 1 yaml: %v", err)
	}
	t.Logf("✓ Wrote Phase 1 YAML: %s", p1File)

	// -------------------------------------------------------------------------
	// PHASE 2: DIRECTIVE PARSING -> phase2_parsed_strokes.yaml (Pure Pratt AST Tree)
	// -------------------------------------------------------------------------
	strokes, err := calque.ParseDirectives(extracted)
	if err != nil {
		t.Fatalf("Phase 2 Directive Parser failed: %v", err)
	}

	var yamlStrokeList []YAMLStroke
	for _, stroke := range strokes {
		var prattTree map[string]any
		if len(stroke.Args) > 0 {
			prattTree = make(map[string]any)
			for k, v := range stroke.Args {
				prattTree[k] = convertPrattNodeToMap(v)
			}
		}
		yamlStrokeList = append(yamlStrokeList, YAMLStroke{
			ID:        stroke.ID,
			Name:      stroke.Name,
			Role:      fmt.Sprintf("%s", stroke.Role),
			Geo:       stroke.Geo,
			PrattTree: prattTree,
		})
	}

	p2File := filepath.Join(outputDir, "phase2_parsed_strokes.yaml")
	p2Bytes, _ := yaml.Marshal(map[string]any{"strokes": yamlStrokeList})
	if err := os.WriteFile(p2File, p2Bytes, 0644); err != nil {
		t.Fatalf("Failed to write phase 2 yaml: %v", err)
	}
	t.Logf("✓ Wrote Phase 2 YAML: %s", p2File)

	// -------------------------------------------------------------------------
	// PHASE 3: COMPILATION -> phase3_compiled_ast.yaml (Fully Autonomous Domain AST)
	// -------------------------------------------------------------------------
	astNodes, err := calque.Compile(extracted)
	if err != nil {
		t.Fatalf("Phase 3 Compiler failed: %v", err)
	}

	var yamlASTList []*YAMLNode
	for _, node := range astNodes {
		yamlASTList = append(yamlASTList, convertNodeToYAMLDTO(node))
	}

	finalYAMLBytes, _ := yaml.Marshal(map[string]any{"astNodes": yamlASTList})
	p3File := filepath.Join(outputDir, "phase3_compiled_ast.yaml")
	if err := os.WriteFile(p3File, finalYAMLBytes, 0644); err != nil {
		t.Fatalf("Failed to write phase 3 yaml: %v", err)
	}
	t.Logf("✓ Wrote Phase 3 YAML with _expr metadata: %s", p3File)

	t.Logf("\n================================================================================")
	t.Logf("     SUCCESS! REAL AUTONOMOUS EXPLORABLE YAML FILES GENERATED IN test/output/")
	t.Logf("       1. test/output/phase1_extracted_annotations.yaml")
	t.Logf("       2. test/output/phase2_parsed_strokes.yaml")
	t.Logf("       3. test/output/phase3_compiled_ast.yaml")
	t.Logf("================================================================================\n")
}

func injectYAMLComments(node *yaml.Node) {
	if node == nil {
		return
	}
	if node.Kind == yaml.MappingNode {
		for i := 0; i < len(node.Value)-1; i += 2 {
			keyNode := node.Content[i]
			valNode := node.Content[i+1]
			if keyNode.Value == "condition" {
				keyNode.HeadComment = " Condition: line.content contains \"ENV_PLACEHOLDER\""
			}
			injectYAMLComments(valNode)
		}
	}
	for _, child := range node.Content {
		injectYAMLComments(child)
	}
}

// TestReloadPhase1YAMLCheckpoint proves Phase 1 YAML can be reloaded as a checkpoint.
func TestReloadPhase1YAMLCheckpoint(t *testing.T) {
	p1File := filepath.Join(".", "output", "phase1_extracted_annotations.yaml")
	f, err := os.Open(p1File)
	if err != nil {
		t.Fatalf("Failed to open phase1_extracted_annotations.yaml checkpoint: %v", err)
	}
	defer f.Close()

	annotations, err := calque.LoadExtractedAnnotationsFromYAML(f)
	if err != nil {
		t.Fatalf("Failed to load Phase 1 annotations from YAML checkpoint: %v", err)
	}

	if len(annotations) != 6 {
		t.Fatalf("Expected 6 extracted annotations reloaded from Phase 1 checkpoint, got %d", len(annotations))
	}
	t.Logf("✓ PHASE 1 CHECKPOINT RELOADED SUCCESSFULLY (%d annotations)!", len(annotations))
}

// TestReloadPhase2YAMLCheckpointAndCompile proves Phase 2 YAML can be reloaded as a checkpoint and compiled into Phase 3 AST.
func TestReloadPhase2YAMLCheckpointAndCompile(t *testing.T) {
	p2File := filepath.Join(".", "output", "phase2_parsed_strokes.yaml")
	f, err := os.Open(p2File)
	if err != nil {
		t.Fatalf("Failed to open phase2_parsed_strokes.yaml checkpoint: %v", err)
	}
	defer f.Close()

	reloadedStrokes, err := calque.LoadStrokesFromYAML(f)
	if err != nil {
		t.Fatalf("Failed to load strokes from Phase 2 YAML checkpoint: %v", err)
	}

	if len(reloadedStrokes) == 0 {
		t.Fatalf("Expected non-empty strokes reloaded from checkpoint")
	}

	astNodes, err := calque.CompileFromStrokes(reloadedStrokes)
	if err != nil {
		t.Fatalf("Phase 3 Compilation from reloaded Phase 2 checkpoint failed: %v", err)
	}

	if len(astNodes) == 0 {
		t.Fatalf("Expected non-empty AST compiled from reloaded Phase 2 checkpoint")
	}

	t.Logf("✓ PHASE 2 CHECKPOINT RELOADED & COMPILED SUCCESSFULLY INTO %d AST NODES!", len(astNodes))
}

// TestReloadPhase3YAMLCheckpoint proves Phase 3 YAML can be reloaded directly into strongly typed AST Nodes.
func TestReloadPhase3YAMLCheckpoint(t *testing.T) {
	p3File := filepath.Join(".", "output", "phase3_compiled_ast.yaml")
	f, err := os.Open(p3File)
	if err != nil {
		t.Fatalf("Failed to open phase3_compiled_ast.yaml checkpoint: %v", err)
	}
	defer f.Close()

	astNodes, err := calque.LoadASTFromYAML(f)
	if err != nil {
		t.Fatalf("Failed to load AST from Phase 3 YAML checkpoint: %v", err)
	}

	if len(astNodes) != 4 {
		t.Fatalf("Expected 4 AST nodes reloaded from Phase 3 checkpoint, got %d", len(astNodes))
	}

	t.Logf("✓ PHASE 3 CHECKPOINT RELOADED SUCCESSFULLY (%d AST nodes ready for Linker/Renderer)!", len(astNodes))
}

func convertPrattNodeToMap(node pratt.Node) any {
	if node == nil {
		return nil
	}

	switch n := node.(type) {
	case *pratt.GenericCallNode:
		var argsList []any
		for _, arg := range n.Args {
			argsList = append(argsList, convertPrattNodeToMap(arg))
		}
		return map[string]any{
			"type":     "GenericCallNode",
			"function": n.Function,
			"args":     argsList,
		}

	case *pratt.GenericBinaryNode:
		return map[string]any{
			"type":  "GenericBinaryNode",
			"op":    n.Op,
			"left":  convertPrattNodeToMap(n.Left),
			"right": convertPrattNodeToMap(n.Right),
		}

	case *pratt.GenericIdentifierNode:
		return map[string]any{
			"type": "GenericIdentifierNode",
			"name": n.Name,
		}

	case *pratt.GenericLiteralNode:
		return map[string]any{
			"type":  "GenericLiteralNode",
			"value": strings.Trim(n.Value, "\""),
		}

	case *pratt.GenericListNode:
		var elementsList []any
		for _, elem := range n.Elements {
			elementsList = append(elementsList, convertPrattNodeToMap(elem))
		}
		return map[string]any{
			"type":     "GenericListNode",
			"elements": elementsList,
		}

	case *pratt.InterpolatedStringNode:
		var partsList []any
		for _, part := range n.Parts {
			partsList = append(partsList, convertPrattNodeToMap(part))
		}
		return map[string]any{
			"type":  "InterpolatedStringNode",
			"parts": partsList,
		}

	default:
		return map[string]any{
			"type":  fmt.Sprintf("%T", node),
			"raw":   fmt.Sprintf("%v", node),
		}
	}
}

func convertNodeToYAMLDTO(node calque.Node) *YAMLNode {
	if node == nil {
		return nil
	}

	dto := &YAMLNode{
		ID: node.GetID(),
	}

	geo := node.GetGeometry()
	if geo.StartLine > 0 {
		dto.Geo = &YAMLGeo{StartLine: geo.StartLine, EndLine: geo.EndLine}
	}

	switch n := node.(type) {
	case *calque.IfFrameNode:
		dto.Type = "IfFrameNode"
		props := make(map[string]any)
		if len(n.Branches) > 0 {
			var branchList []map[string]any
			for idx, b := range n.Branches {
				bMap := map[string]any{
					"branchIndex": idx,
				}
				if b.Condition != nil && b.Condition.Expr != nil {
					bMap["condition"] = b.Condition.Expr.ToMap()
				}
				if len(b.Nodes) > 0 {
					var childList []*YAMLNode
					for _, cn := range b.Nodes {
						childList = append(childList, convertNodeToYAMLDTO(cn))
					}
					bMap["nodes"] = childList
				}
				branchList = append(branchList, bMap)
			}
			props["branches"] = branchList
		}
		dto.Properties = props

	case *calque.MacroFrameNode:
		dto.Type = "MacroFrameNode"
		dto.Properties = convertPropertiesMap(n.GetProperties())
		for _, child := range n.Nodes {
			dto.Children = append(dto.Children, convertNodeToYAMLDTO(child))
		}

	case *calque.SectionFrameNode:
		dto.Type = "SectionFrameNode"
		dto.Properties = convertPropertiesMap(n.GetProperties())
		for _, child := range n.Nodes {
			dto.Children = append(dto.Children, convertNodeToYAMLDTO(child))
		}

	case *calque.ActionNode:
		dto.Type = fmt.Sprintf("%sNode", n.Name)
		dto.Properties = convertPropertiesMap(n.Properties)
	}

	return dto
}

func convertPropertiesMap(props map[string]any) map[string]any {
	result := make(map[string]any)
	for k, v := range props {
		result[k] = cleanAnyValue(v)
	}
	return result
}

func cleanAnyValue(v any) any {
	if v == nil {
		return nil
	}

	switch val := v.(type) {
	case calque.Node:
		return convertNodeToYAMLDTO(val)

	case calque.Pipeline:
		var pipelineList []*YAMLNode
		for _, pNode := range val {
			pipelineList = append(pipelineList, convertNodeToYAMLDTO(pNode))
		}
		return pipelineList

	case []calque.Node:
		var nodeList []*YAMLNode
		for _, pNode := range val {
			nodeList = append(nodeList, convertNodeToYAMLDTO(pNode))
		}
		return nodeList

	case []any:
		var list []any
		for _, item := range val {
			list = append(list, cleanAnyValue(item))
		}
		return list

	case pratt.Node:
		return cleanPrattNodeValue(val)

	case *calque.Condition:
		if val != nil && val.Expr != nil {
			return val.Expr.ToMap()
		}
		return nil

	default:
		return val
	}
}

func cleanPrattNodeValue(node pratt.Node) any {
	if node == nil {
		return nil
	}

	switch n := node.(type) {
	case *pratt.GenericIdentifierNode:
		return n.Name

	case *pratt.GenericLiteralNode:
		return strings.Trim(n.Value, "\"")

	case *pratt.GenericCallNode:
		callMap := map[string]any{"call": n.Function}
		if len(n.Args) > 0 {
			var argsList []any
			for _, a := range n.Args {
				argsList = append(argsList, cleanPrattNodeValue(a))
			}
			callMap["args"] = argsList
		}
		return callMap

	case *pratt.GenericBinaryNode:
		return fmt.Sprintf("%v %s %v", cleanPrattNodeValue(n.Left), n.Op, cleanPrattNodeValue(n.Right))

	case *pratt.GenericListNode:
		var list []any
		for _, elem := range n.Elements {
			list = append(list, cleanPrattNodeValue(elem))
		}
		return list

	case *pratt.InterpolatedStringNode:
		return fmt.Sprintf("%v", n.Parts)

	default:
		return fmt.Sprintf("%v", node)
	}
}

func marshalPhase1YAML(annotations []YAMLAnnotation) ([]byte, error) {
	doc := &yaml.Node{Kind: yaml.DocumentNode}
	rootMap := &yaml.Node{Kind: yaml.MappingNode}

	keyNode := &yaml.Node{Kind: yaml.ScalarNode, Value: "annotations"}
	listNode := &yaml.Node{Kind: yaml.SequenceNode}

	for _, ann := range annotations {
		annMap := &yaml.Node{Kind: yaml.MappingNode}

		// id
		annMap.Content = append(annMap.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: "id"},
			&yaml.Node{Kind: yaml.ScalarNode, Value: ann.ID},
		)

		// geo
		geoMap := &yaml.Node{Kind: yaml.MappingNode, Style: yaml.FlowStyle}
		geoMap.Content = append(geoMap.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: "startline"},
			&yaml.Node{Kind: yaml.ScalarNode, Value: fmt.Sprintf("%d", ann.Geo.StartLine)},
			&yaml.Node{Kind: yaml.ScalarNode, Value: "endline"},
			&yaml.Node{Kind: yaml.ScalarNode, Value: fmt.Sprintf("%d", ann.Geo.EndLine)},
			&yaml.Node{Kind: yaml.ScalarNode, Value: "indent"},
			&yaml.Node{Kind: yaml.ScalarNode, Value: fmt.Sprintf("%d", ann.Geo.Indent)},
		)
		annMap.Content = append(annMap.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: "geo"},
			geoMap,
		)

		// rawContent
		contentNode := &yaml.Node{Kind: yaml.ScalarNode, Value: ann.RawContent}
		if strings.Contains(ann.RawContent, "\n") {
			contentNode.Value = strings.TrimRight(ann.RawContent, "\r\n")
			contentNode.Style = yaml.LiteralStyle
		}
		annMap.Content = append(annMap.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: "rawContent"},
			contentNode,
		)

		listNode.Content = append(listNode.Content, annMap)
	}

	rootMap.Content = append(rootMap.Content, keyNode, listNode)
	doc.Content = append(doc.Content, rootMap)
	return yaml.Marshal(doc)
}
