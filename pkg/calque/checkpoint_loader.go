package calque

import (
	"fmt"
	"io"
	"strings"

	"github.com/ouvrage-systems/ouvrage-calque-go/pkg/extractor"
	"github.com/ouvrage-systems/ouvrage-calque-go/pkg/geometry"
	"github.com/ouvrage-systems/ouvrage-calque-go/pkg/syntax/pratt"
	"gopkg.in/yaml.v3"
)

// -----------------------------------------------------------------------------
// PHASE 1 CHECKPOINT LOADER: ExtractedAnnotations
// -----------------------------------------------------------------------------

type YAMLExtractedAnnotationDTO struct {
	ID         string            `yaml:"id"`
	Geo        geometry.Geometry `yaml:"geo"`
	RawContent string            `yaml:"rawContent"`
}

// LoadExtractedAnnotationsFromYAML deserializes a phase1_extracted_annotations.yaml checkpoint file back into []extractor.ExtractedAnnotation.
func LoadExtractedAnnotationsFromYAML(r io.Reader) ([]extractor.ExtractedAnnotation, error) {
	var payload struct {
		Annotations []YAMLExtractedAnnotationDTO `yaml:"annotations"`
	}

	decoder := yaml.NewDecoder(r)
	if err := decoder.Decode(&payload); err != nil {
		return nil, fmt.Errorf("failed to decode phase 1 yaml checkpoint: %w", err)
	}

	var results []extractor.ExtractedAnnotation
	for _, dto := range payload.Annotations {
		results = append(results, extractor.ExtractedAnnotation{
			ID:      dto.ID,
			Geo:     dto.Geo,
			Content: dto.RawContent,
		})
	}

	return results, nil
}

// -----------------------------------------------------------------------------
// PHASE 2 CHECKPOINT LOADER: ParsedStrokes
// -----------------------------------------------------------------------------

type YAMLStrokeDTO struct {
	ID        string            `yaml:"id"`
	Name      string            `yaml:"name"`
	Role      string            `yaml:"role"`
	Geo       geometry.Geometry `yaml:"geo"`
	PrattTree map[string]any    `yaml:"prattTree"`
}

// LoadStrokesFromYAML deserializes a phase2_parsed_strokes.yaml checkpoint file back into []*ParsedStroke.
func LoadStrokesFromYAML(r io.Reader) ([]*ParsedStroke, error) {
	var payload struct {
		Strokes []YAMLStrokeDTO `yaml:"strokes"`
	}

	decoder := yaml.NewDecoder(r)
	if err := decoder.Decode(&payload); err != nil {
		return nil, fmt.Errorf("failed to decode phase 2 yaml checkpoint: %w", err)
	}

	var strokes []*ParsedStroke
	for _, dto := range payload.Strokes {
		role := parseRoleString(dto.Role)
		args := make(map[string]pratt.Node)

		for k, v := range dto.PrattTree {
			args[k] = deserializePrattNodeFromMap(v)
		}

		strokes = append(strokes, &ParsedStroke{
			ID:   dto.ID,
			Geo:  dto.Geo,
			Role: role,
			Name: dto.Name,
			Args: args,
		})
	}

	return strokes, nil
}

// -----------------------------------------------------------------------------
// PHASE 3 CHECKPOINT LOADER: Compiled AST Nodes
// -----------------------------------------------------------------------------

type YAMLNodeDTO struct {
	Type       string         `yaml:"type"`
	ID         string         `yaml:"id"`
	Geo        *YAMLGeoDTO    `yaml:"geo,omitempty"`
	Properties map[string]any `yaml:"properties,omitempty"`
	Children   []*YAMLNodeDTO `yaml:"children,omitempty"`
}

type YAMLGeoDTO struct {
	StartLine int `yaml:"startLine"`
	EndLine   int `yaml:"endLine"`
}

// LoadASTFromYAML deserializes a phase3_compiled_ast.yaml checkpoint file back into []Node AST structures.
func LoadASTFromYAML(r io.Reader) ([]Node, error) {
	var payload struct {
		ASTNodes []*YAMLNodeDTO `yaml:"astNodes"`
	}

	decoder := yaml.NewDecoder(r)
	if err := decoder.Decode(&payload); err != nil {
		return nil, fmt.Errorf("failed to decode phase 3 yaml checkpoint: %w", err)
	}

	var results []Node
	for _, dto := range payload.ASTNodes {
		results = append(results, deserializeNodeFromDTO(dto))
	}

	return results, nil
}

func deserializeNodeFromDTO(dto *YAMLNodeDTO) Node {
	if dto == nil {
		return nil
	}

	var geo geometry.Geometry
	if dto.Geo != nil {
		geo = geometry.Geometry{StartLine: dto.Geo.StartLine, EndLine: dto.Geo.EndLine}
	}

	switch dto.Type {
	case "IfFrameNode":
		return &IfFrameNode{
			ID:  dto.ID,
			Geo: geo,
		}

	case "MacroFrameNode":
		var children []Node
		for _, c := range dto.Children {
			children = append(children, deserializeNodeFromDTO(c))
		}
		return &MacroFrameNode{
			ID:    dto.ID,
			Geo:   geo,
			Nodes: children,
		}

	case "SectionFrameNode":
		var children []Node
		for _, c := range dto.Children {
			children = append(children, deserializeNodeFromDTO(c))
		}
		name, _ := dto.Properties["name"].(string)
		return &SectionFrameNode{
			ID:    dto.ID,
			Geo:   geo,
			Name:  name,
			Nodes: children,
		}

	default:
		// ActionNode (ReplaceNode, ImportNode, InsertFragmentNode, etc.)
		actionName := strings.TrimSuffix(dto.Type, "Node")
		return &ActionNode{
			ID:         dto.ID,
			Geo:        geo,
			Name:       actionName,
			Properties: dto.Properties,
		}
	}
}

func parseRoleString(roleStr string) StrokeRole {
	switch roleStr {
	case "Action":
		return RoleAction
	case "Open":
		return RoleOpen
	case "Close":
		return RoleClose
	case "Split":
		return RoleSplit
	case "Content":
		return RoleContent
	default:
		return RoleAction
	}
}

func deserializePrattNodeFromMap(val any) pratt.Node {
	if val == nil {
		return nil
	}

	m, ok := val.(map[string]any)
	if !ok {
		return nil
	}

	nodeType, _ := m["type"].(string)

	switch nodeType {
	case "GenericLiteralNode":
		vStr, _ := m["value"].(string)
		return &pratt.GenericLiteralNode{Value: fmt.Sprintf("%q", vStr)}

	case "GenericIdentifierNode":
		nameStr, _ := m["name"].(string)
		return &pratt.GenericIdentifierNode{Name: nameStr}

	case "GenericBinaryNode":
		opStr, _ := m["op"].(string)
		left := deserializePrattNodeFromMap(m["left"])
		right := deserializePrattNodeFromMap(m["right"])
		return &pratt.GenericBinaryNode{Op: opStr, Left: left, Right: right}

	case "GenericCallNode":
		fnStr, _ := m["function"].(string)
		var args []pratt.Node
		if rawArgs, ok := m["args"].([]any); ok {
			for _, a := range rawArgs {
				args = append(args, deserializePrattNodeFromMap(a))
			}
		}
		return &pratt.GenericCallNode{Function: fnStr, Args: args}

	case "GenericListNode":
		var elems []pratt.Node
		if rawElems, ok := m["elements"].([]any); ok {
			for _, e := range rawElems {
				elems = append(elems, deserializePrattNodeFromMap(e))
			}
		}
		return &pratt.GenericListNode{Elements: elems}

	default:
		return nil
	}
}
