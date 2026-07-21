package calque

import (
	"fmt"
	"strings"

	"github.com/ouvrage-systems/ouvrage-calque-go/pkg/extractor"
	"github.com/ouvrage-systems/ouvrage-calque-go/pkg/syntax/expr"
	"github.com/ouvrage-systems/ouvrage-calque-go/pkg/syntax/pratt"
)

// ParseDirectives converts a slice of ExtractedAnnotations into a flat slice of ParsedStrokes.
func ParseDirectives(annotations []extractor.ExtractedAnnotation) ([]*ParsedStroke, error) {
	rules := expr.NewExpressionRuleSet()
	engine := pratt.NewEngine(rules)

	var strokes []*ParsedStroke

	for _, ann := range annotations {
		stroke, err := parseSingleAnnotation(ann, engine)
		if err != nil {
			return nil, err
		}
		strokes = append(strokes, stroke)
	}

	return strokes, nil
}

func parseSingleAnnotation(ann extractor.ExtractedAnnotation, engine *pratt.Engine) (*ParsedStroke, error) {
	content := strings.TrimSpace(ann.Content)
	if content == "" {
		return nil, fmt.Errorf("directive error at L%d: empty annotation", ann.Geo.StartLine)
	}

	stream, err := expr.TokenizeExpression(content, ann.Geo.StartLine)
	if err != nil {
		return nil, fmt.Errorf("tokenization error at L%d: %w", ann.Geo.StartLine, err)
	}

	astNode, err := engine.Parse(stream, 0)
	if err != nil {
		return nil, fmt.Errorf("parsing error at L%d for '%s': %w", ann.Geo.StartLine, content, err)
	}

	stroke := &ParsedStroke{
		ID:   ann.ID,
		Geo:  ann.Geo,
		Args: make(map[string]pratt.Node),
	}

	switch n := astNode.(type) {
	case *pratt.GenericCallNode:
		stroke.Name = n.Function
		stroke.Role = determineRole(stroke.Name)
		stroke.Args = extractArgsFromCall(n)

	case *pratt.GenericIdentifierNode:
		stroke.Name = n.Name
		stroke.Role = determineRole(stroke.Name)

	default:
		return nil, fmt.Errorf("directive error at L%d: invalid directive format '%s'", ann.Geo.StartLine, content)
	}

	return stroke, nil
}

func determineRole(name string) StrokeRole {
	switch {
	case strings.HasPrefix(name, "End") || name == "EndIf" || name == "EndLoop" || name == "EndSection" || name == "EndRegisterOutput" || name == "EndMacro":
		return RoleClose
	case name == "Else" || name == "Elif":
		return RoleSplit
	case name == "If" || name == "Loop" || name == "Section" || name == "RegisterOutput" || name == "Macro":
		return RoleOpen
	case name == "Line" || name == "InsertFragment" || name == "InsertOutput" || name == "Call" || name == "Import" || name == "Const":
		return RoleContent
	default:
		return RoleAction
	}
}

func extractArgsFromCall(call *pratt.GenericCallNode) map[string]pratt.Node {
	argsMap := make(map[string]pratt.Node)

	for i, arg := range call.Args {
		if bin, ok := arg.(*pratt.GenericBinaryNode); ok && (bin.Op == "=" || bin.Op == "==") {
			if ident, ok := bin.Left.(*pratt.GenericIdentifierNode); ok {
				argsMap[ident.Name] = bin.Right
				continue
			}
		}
		if callArg, ok := arg.(*pratt.GenericCallNode); ok {
			if len(callArg.Args) == 1 {
				argsMap[callArg.Function] = callArg.Args[0]
			} else {
				argsMap[callArg.Function] = callArg
			}
			continue
		}
		// Positional argument
		argsMap[fmt.Sprintf("arg_%d", i)] = arg
	}

	return argsMap
}
