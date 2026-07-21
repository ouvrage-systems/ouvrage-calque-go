package calque

import (
	"crypto/sha256"
	"fmt"
	"strconv"
	"strings"

	"github.com/ouvrage-systems/ouvrage-calque-go/pkg/extractor"
	"github.com/ouvrage-systems/ouvrage-calque-go/pkg/geometry"
	"github.com/ouvrage-systems/ouvrage-calque-go/pkg/syntax/pratt"
)

type frameBuilder interface {
	Node
	addNode(node Node)
	addBranch(cond *Condition)
	setEndLine(endLine int)
}

type buildingIfFrame struct {
	node *IfFrameNode
}

func (b *buildingIfFrame) GetID() string            { return b.node.ID }
func (b *buildingIfFrame) GetGeometry() geometry.Geometry { return b.node.Geo }
func (b *buildingIfFrame) GetProperties() map[string]any { return b.node.GetProperties() }
func (b *buildingIfFrame) String() string           { return b.node.String() }
func (b *buildingIfFrame) setEndLine(line int)      { b.node.Geo.EndLine = line }

func (b *buildingIfFrame) addNode(node Node) {
	if len(b.node.Branches) == 0 {
		b.node.Branches = append(b.node.Branches, IfBranch{})
	}
	lastIdx := len(b.node.Branches) - 1
	b.node.Branches[lastIdx].Nodes = append(b.node.Branches[lastIdx].Nodes, node)
}

func (b *buildingIfFrame) addBranch(cond *Condition) {
	b.node.Branches = append(b.node.Branches, IfBranch{Condition: cond})
}

type buildingGenericFrame struct {
	id    string
	geo   geometry.Geometry
	name  string
	nodes *[]Node
	node  Node
}

func (b *buildingGenericFrame) GetID() string            { return b.id }
func (b *buildingGenericFrame) GetGeometry() geometry.Geometry { return b.geo }
func (b *buildingGenericFrame) GetProperties() map[string]any { return b.node.GetProperties() }
func (b *buildingGenericFrame) String() string           { return b.node.String() }
func (b *buildingGenericFrame) setEndLine(line int)      {}
func (b *buildingGenericFrame) addNode(node Node) {
	*b.nodes = append(*b.nodes, node)
}
func (b *buildingGenericFrame) addBranch(cond *Condition) {}

// Compile performs Phase 3 compilation: converting ExtractedAnnotations into a strongly typed structural Calque AST tree.
func Compile(annotations []extractor.ExtractedAnnotation) ([]Node, error) {
	strokes, err := ParseDirectives(annotations)
	if err != nil {
		return nil, err
	}
	return CompileFromStrokes(strokes)
}

// CompileFromStrokes compiles pre-parsed ParsedStrokes (e.g. reloaded from a YAML checkpoint) directly into Phase 3 AST.
func CompileFromStrokes(strokes []*ParsedStroke) ([]Node, error) {
	var rootNodes []Node
	var stack []frameBuilder

	for _, stroke := range strokes {
		switch stroke.Role {
		case RoleOpen:
			builder := createFrameBuilder(stroke)
			stack = append(stack, builder)

		case RoleSplit:
			if len(stack) == 0 {
				return nil, fmt.Errorf("compile error at L%d: orphan '%s' directive outside of any frame block", stroke.Geo.StartLine, stroke.Name)
			}
			top := stack[len(stack)-1]
			ifIf, ok := top.(*buildingIfFrame)
			if !ok {
				return nil, fmt.Errorf("compile error at L%d: split directive '%s' is only allowed inside an 'If' frame", stroke.Geo.StartLine, stroke.Name)
			}
			condNode := getConditionArg(stroke)
			ifIf.addBranch(condNode)

		case RoleClose:
			if len(stack) == 0 {
				return nil, fmt.Errorf("compile error at L%d: unmatched closing directive '%s'", stroke.Geo.StartLine, stroke.Name)
			}

			// Pop frame
			top := stack[len(stack)-1]
			stack = stack[:len(stack)-1]

			// Update physical end line boundary
			top.setEndLine(stroke.Geo.EndLine)

			builtNode := unwrapBuilder(top, stroke.Geo.EndLine)

			if len(stack) > 0 {
				stack[len(stack)-1].addNode(builtNode)
			} else {
				rootNodes = append(rootNodes, builtNode)
			}

		case RoleAction, RoleContent:
			props := compileActionProperties(stroke)
			actionNode := &ActionNode{
				ID:         stroke.ID,
				Geo:        stroke.Geo,
				Name:       stroke.Name,
				Properties: props,
				TypedProps: props,
			}
			if len(stack) > 0 {
				stack[len(stack)-1].addNode(actionNode)
			} else {
				rootNodes = append(rootNodes, actionNode)
			}
		}
	}

	if len(stack) > 0 {
		unclosed := stack[len(stack)-1]
		return nil, fmt.Errorf("compile error at L%d: unclosed frame '%s' opened at L%d", unclosed.GetGeometry().StartLine, unclosed.String(), unclosed.GetGeometry().StartLine)
	}

	return rootNodes, nil
}

// Backward compatibility alias for Assemble
func Assemble(annotations []extractor.ExtractedAnnotation) ([]Node, error) {
	return Compile(annotations)
}

func createFrameBuilder(stroke *ParsedStroke) frameBuilder {
	switch stroke.Name {
	case "If":
		cond := getConditionArg(stroke)
		ifNode := &IfFrameNode{
			ID:  stroke.ID,
			Geo: stroke.Geo,
			Branches: []IfBranch{
				{Condition: cond},
			},
		}
		return &buildingIfFrame{node: ifNode}

	case "Loop":
		itemVar := getArgString(stroke, "item", "srv")
		coll := getPrattNodeArg(stroke, "in")
		loopNode := &LoopFrameNode{
			ID:         stroke.ID,
			Geo:        stroke.Geo,
			ItemVar:    itemVar,
			Collection: coll,
			TypedProps: LoopProps{
				ItemVar:    itemVar,
				Collection: coll,
			},
		}
		return &buildingGenericFrame{
			id:    stroke.ID,
			geo:   stroke.Geo,
			name:  stroke.Name,
			nodes: &loopNode.Nodes,
			node:  loopNode,
		}

	case "Macro":
		macroName := getArgString(stroke, "name", "macro")
		params := getArgStringList(stroke, "args")
		pipelines := getArgPipeline(stroke, "pipelines")
		macroNode := &MacroFrameNode{
			ID:        stroke.ID,
			Geo:       stroke.Geo,
			Name:      macroName,
			Params:    params,
			Pipelines: pipelines,
			TypedProps: MacroProps{
				Name:     macroName,
				Params:   params,
				Pipeline: pipelines,
			},
		}
		return &buildingGenericFrame{
			id:    stroke.ID,
			geo:   stroke.Geo,
			name:  stroke.Name,
			nodes: &macroNode.Nodes,
			node:  macroNode,
		}

	default:
		secName := getArgString(stroke, "name", stroke.Name)
		pipelines := getArgPipeline(stroke, "pipeline")
		secNode := &SectionFrameNode{
			ID:        stroke.ID,
			Geo:       stroke.Geo,
			Name:      secName,
			Pipelines: pipelines,
			TypedProps: SectionProps{
				Name:     secName,
				Pipeline: pipelines,
			},
		}
		return &buildingGenericFrame{
			id:    stroke.ID,
			geo:   stroke.Geo,
			name:  stroke.Name,
			nodes: &secNode.Nodes,
			node:  secNode,
		}
	}
}

func unwrapBuilder(b frameBuilder, endLine int) Node {
	switch v := b.(type) {
	case *buildingIfFrame:
		v.node.Geo.EndLine = endLine
		return v.node
	case *buildingGenericFrame:
		switch fn := v.node.(type) {
		case *LoopFrameNode:
			fn.Geo.EndLine = endLine
			return fn
		case *MacroFrameNode:
			fn.Geo.EndLine = endLine
			return fn
		case *SectionFrameNode:
			fn.Geo.EndLine = endLine
			return fn
		}
	}
	return b
}

func compileActionProperties(stroke *ParsedStroke) map[string]any {
	props := make(map[string]any)

	switch stroke.Name {
	case "Import":
		props["source"] = getArgString(stroke, "source", "")
		props["as"] = getArgString(stroke, "as", "")

	case "Replace":
		if pat := getArgString(stroke, "pattern", ""); pat != "" {
			props["pattern"] = pat
		}
		props["with"] = getArgString(stroke, "with", "")
		if when := getConditionArg(stroke); when != nil {
			props["when"] = when
		}

	case "InsertFragment":
		props["source"] = getArgString(stroke, "source", "")

	case "InsertOutput":
		props["target"] = getArgString(stroke, "target", "")

	case "Strip":
		props["lines"] = getArgInt(stroke, "lines", 1)

	case "Remove":
		props["lines"] = getArgInt(stroke, "lines", 1)

	default:
		// Convert all pratt.Node arguments recursively into AST properties
		for k, v := range stroke.Args {
			if strings.HasPrefix(k, "arg_") {
				continue
			}
			props[k] = compilePrattValueToAST(v, stroke.ID+"_"+k)
		}
	}

	return props
}

func compilePrattValueToAST(node pratt.Node, parentID string) any {
	switch n := node.(type) {
	case *pratt.GenericListNode:
		var list []any
		for i, elem := range n.Elements {
			elemID := generateDeterministicID(fmt.Sprintf("%s_elem_%d", parentID, i))
			list = append(list, compilePrattValueToAST(elem, elemID))
		}
		return list

	case *pratt.GenericCallNode:
		return assembleCallExprToNode(n, parentID)

	case *pratt.GenericIdentifierNode:
		return n.Name

	case *pratt.GenericLiteralNode:
		return strings.Trim(n.Value, "\"")

	case *pratt.InterpolatedStringNode:
		return fmt.Sprintf("%v", n.Parts)

	default:
		return node
	}
}

func getConditionArg(stroke *ParsedStroke) *Condition {
	if cond, ok := stroke.Args["cond"]; ok {
		return &Condition{Expr: CompilePrattToDomainExpr(cond)}
	}
	if when, ok := stroke.Args["when"]; ok {
		return &Condition{Expr: CompilePrattToDomainExpr(when)}
	}
	if def, ok := stroke.Args["default"]; ok {
		return &Condition{Expr: CompilePrattToDomainExpr(def)}
	}
	return nil
}

func getPrattNodeArg(stroke *ParsedStroke, key string) pratt.Node {
	if node, ok := stroke.Args[key]; ok {
		return node
	}
	return nil
}

func getArgInt(stroke *ParsedStroke, key string, fallback int) int {
	if s := getArgString(stroke, key, ""); s != "" {
		if val, err := strconv.Atoi(s); err == nil {
			return val
		}
	}
	return fallback
}

func getArgString(stroke *ParsedStroke, key string, fallback string) string {
	if node, ok := stroke.Args[key]; ok {
		if ident, ok := node.(*pratt.GenericIdentifierNode); ok {
			return ident.Name
		}
		if lit, ok := node.(*pratt.GenericLiteralNode); ok {
			return strings.Trim(lit.Value, "\"")
		}
	}
	// Also check arg_1 or positional
	if node, ok := stroke.Args["arg_1"]; ok {
		if lit, ok := node.(*pratt.GenericLiteralNode); ok {
			return strings.Trim(lit.Value, "\"")
		}
	}
	return fallback
}

func getArgStringList(stroke *ParsedStroke, key string) []string {
	var results []string
	for k, v := range stroke.Args {
		if k == key || k == "arg_3" {
			if list, ok := v.(*pratt.GenericListNode); ok {
				for _, elem := range list.Elements {
					if ident, ok := elem.(*pratt.GenericIdentifierNode); ok {
						results = append(results, ident.Name)
					} else if lit, ok := elem.(*pratt.GenericLiteralNode); ok {
						results = append(results, strings.Trim(lit.Value, "\""))
					}
				}
			}
		}
	}
	return results
}

func getArgPipeline(stroke *ParsedStroke, key string) Pipeline {
	var pipeline Pipeline
	for _, v := range stroke.Args {
		if list, ok := v.(*pratt.GenericListNode); ok {
			for idx, elem := range list.Elements {
				if call, ok := elem.(*pratt.GenericCallNode); ok {
					nodeID := generateDeterministicID(fmt.Sprintf("%s_pipe_%d_%s", stroke.ID, idx, call.Function))
					node := assembleCallExprToNode(call, nodeID)
					pipeline = append(pipeline, node)
				}
			}
		}
	}
	return pipeline
}

func assembleCallExprToNode(call *pratt.GenericCallNode, nodeID string) Node {
	argsMap := extractArgsFromCall(call)
	stroke := &ParsedStroke{
		ID:   nodeID,
		Name: call.Function,
		Args: argsMap,
	}

	// If call expression is a nested frame directive (like If, Loop, ForEachLine)
	if call.Function == "If" {
		cond := getConditionArg(stroke)
		nodesArg := stroke.Args["nodes"]
		var innerNodes []Node
		if listNode, ok := nodesArg.(*pratt.GenericListNode); ok {
			for idx, elem := range listNode.Elements {
				elemID := generateDeterministicID(fmt.Sprintf("%s_if_%d", nodeID, idx))
				if innerCall, ok := elem.(*pratt.GenericCallNode); ok {
					innerNodes = append(innerNodes, assembleCallExprToNode(innerCall, elemID))
				}
			}
		}
		return &IfFrameNode{
			ID:  nodeID,
			Geo: geometry.Geometry{},
			Branches: []IfBranch{
				{Condition: cond, Nodes: innerNodes},
			},
		}
	}

	props := compileActionProperties(stroke)
	return &ActionNode{
		ID:         nodeID,
		Geo:        geometry.Geometry{}, // Zero geometry for synthetic pipeline inner nodes
		Name:       call.Function,
		Properties: props,
		TypedProps: props,
	}
}

func convertPrattArgsToProperties(argsMap map[string]pratt.Node) map[string]any {
	props := make(map[string]any)
	for k, v := range argsMap {
		if strings.HasPrefix(k, "arg_") {
			continue // Omit raw positional index keys in clean domain Properties map
		}
		props[k] = compilePrattValueToAST(v, k)
	}
	return props
}

func generateDeterministicID(raw string) string {
	h := sha256.New()
	h.Write([]byte(raw))
	return fmt.Sprintf("%x", h.Sum(nil))[:12]
}
