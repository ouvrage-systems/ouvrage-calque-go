package calque

import (
	"fmt"

	"github.com/ouvrage-systems/ouvrage-calque-go/pkg/geometry"
	"github.com/ouvrage-systems/ouvrage-calque-go/pkg/syntax/pratt"
)

// StrokeRole defines the syntactic role of a directive in the stream flow.
type StrokeRole int

const (
	RoleAction StrokeRole = iota // Mutates lines (e.g. Replace, Strip, DeactivateAction)
	RoleOpen                     // Opens a boundary block (e.g. If, Loop, Section, RegisterOutput, Macro)
	RoleClose                    // Closes a boundary block (e.g. EndIf, EndLoop, EndSection, EndRegisterOutput, EndMacro)
	RoleSplit                    // Splits a boundary block (e.g. Else, Elif)
	RoleContent                  // Generates content (e.g. Line, InsertFragment, InsertOutput, Const)
)

func (r StrokeRole) String() string {
	switch r {
	case RoleAction:
		return "Action"
	case RoleOpen:
		return "Open"
	case RoleClose:
		return "Close"
	case RoleSplit:
		return "Split"
	case RoleContent:
		return "Content"
	default:
		return "Unknown"
	}
}

// ParsedStroke represents a flat, typed directive stroke parsed via the Pratt Engine.
type ParsedStroke struct {
	ID   string            // Inherited deterministic SHA-256 hash ID
	Geo  geometry.Geometry // Physical 2D geometry (StartLine, EndLine, Indent)
	Role StrokeRole        // Open, Close, Split, Action, Content
	Name string            // e.g. "If", "Else", "EndIf", "Replace", "Line", "Const"
	Args map[string]pratt.Node // Parameter AST nodes parsed by the Pratt Engine (e.g. "cond", "with", "when")
}

// String returns a clean representation of the ParsedStroke for debugging & logs.
func (ps *ParsedStroke) String() string {
	return fmt.Sprintf("ParsedStroke[%s] Name:%s Role:%s Geo:%s Args:%d",
		ps.ID, ps.Name, ps.Role, ps.Geo, len(ps.Args))
}
