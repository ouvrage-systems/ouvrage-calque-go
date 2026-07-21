package calque

import (
	"fmt"

	"github.com/ouvrage-systems/ouvrage-calque-go/pkg/geometry"
	"github.com/ouvrage-systems/ouvrage-calque-go/pkg/syntax/pratt"
)

// Node is the universal interface for all nodes in the structural Calque AST.
type Node interface {
	GetID() string
	GetGeometry() geometry.Geometry
	GetProperties() map[string]any
	String() string
}

// Pipeline represents an ordered sequence of transformation action nodes.
type Pipeline []Node

// Condition wraps a compiled, strongly typed domain expression AST (CompareExpr, LogicalExpr, etc.).
type Condition struct {
	Expr Expr
}

func (c *Condition) String() string {
	if c == nil || c.Expr == nil {
		return "null"
	}
	return c.Expr.String()
}

// -----------------------------------------------------------------------------
// 1. CONDITIONAL FRAME: IfFrameNode & IfBranch
// -----------------------------------------------------------------------------

// IfBranch represents a conditional branch within an IfFrameNode (If, Elif, or Else).
type IfBranch struct {
	Condition *Condition // Pratt AST condition (nil for the Else branch!)
	Nodes     []Node     // Child actions and nested frames inside this branch
}

// IfFrameNode represents an If / Elif / Else conditional boundary block.
type IfFrameNode struct {
	ID       string            // Deterministic SHA-256 ID of the opening If directive
	Geo      geometry.Geometry // Physical 2D range (from If to EndIf). Zero for synthetic nodes.
	Branches []IfBranch        // Branch 0 = If, 1..N-1 = Elif, Last = Else (Condition == nil)
}

func (f *IfFrameNode) GetID() string            { return f.ID }
func (f *IfFrameNode) GetGeometry() geometry.Geometry { return f.Geo }
func (f *IfFrameNode) GetProperties() map[string]any {
	props := make(map[string]any)
	var branchList []map[string]any
	for bIdx, b := range f.Branches {
		branchMap := map[string]any{
			"branchIndex": bIdx,
			"condition":   b.Condition,
			"nodeCount":   len(b.Nodes),
			"children":    b.Nodes,
		}
		branchList = append(branchList, branchMap)
	}
	props["branches"] = branchList
	return props
}
func (f *IfFrameNode) String() string {
	return fmt.Sprintf("IfFrameNode[%s] Geo:%s Branches:%d", f.ID, f.Geo, len(f.Branches))
}

// -----------------------------------------------------------------------------
// 2. LOOP FRAME: LoopFrameNode
// -----------------------------------------------------------------------------

// LoopFrameNode represents a loop boundary block (e.g. Loop(item=srv, in=fleet.servers)).
type LoopFrameNode struct {
	ID         string            // Deterministic SHA-256 ID of opening Loop directive
	Geo        geometry.Geometry // Physical 2D range (from Loop to EndLoop)
	ItemVar    string            // e.g. "srv"
	IndexVar   string            // e.g. "index" (defaults to "loop.index")
	Collection pratt.Node        // Pratt AST evaluating to iterable collection
	Nodes      []Node            // Body nodes inside loop
	TypedProps LoopProps         // Strongly typed properties struct
}

func (f *LoopFrameNode) GetID() string            { return f.ID }
func (f *LoopFrameNode) GetGeometry() geometry.Geometry { return f.Geo }
func (f *LoopFrameNode) GetProperties() map[string]any {
	return map[string]any{
		"itemVar":    f.ItemVar,
		"indexVar":   f.IndexVar,
		"collection": f.Collection,
	}
}
func (f *LoopFrameNode) String() string {
	return fmt.Sprintf("LoopFrameNode[%s] ItemVar:%s Geo:%s Nodes:%d", f.ID, f.ItemVar, f.Geo, len(f.Nodes))
}

// -----------------------------------------------------------------------------
// 3. MACRO FRAME: MacroFrameNode
// -----------------------------------------------------------------------------

// MacroFrameNode represents a macro declaration boundary block.
type MacroFrameNode struct {
	ID        string            // Deterministic SHA-256 ID of opening Macro directive
	Geo       geometry.Geometry // Physical 2D range (from Macro to EndMacro)
	Name      string            // Macro name (e.g. "auth_env")
	Params    []string          // Declared parameters
	Pipelines Pipeline          // Ordered sequence of transformation nodes
	Nodes     []Node            // Macro template body nodes
	TypedProps MacroProps       // Strongly typed properties struct
}

func (f *MacroFrameNode) GetID() string            { return f.ID }
func (f *MacroFrameNode) GetGeometry() geometry.Geometry { return f.Geo }
func (f *MacroFrameNode) GetProperties() map[string]any {
	return map[string]any{
		"name":     f.Name,
		"params":   f.Params,
		"pipeline": f.Pipelines,
	}
}
func (f *MacroFrameNode) String() string {
	return fmt.Sprintf("MacroFrameNode[%s] Name:%s Geo:%s Pipelines:%d Nodes:%d", f.ID, f.Name, f.Geo, len(f.Pipelines), len(f.Nodes))
}

// -----------------------------------------------------------------------------
// 4. SECTION / REGISTER OUTPUT FRAME: SectionFrameNode
// -----------------------------------------------------------------------------

// SectionFrameNode represents a simple boundary section or register output block.
type SectionFrameNode struct {
	ID         string            // Deterministic SHA-256 ID of opening directive
	Geo        geometry.Geometry // Physical 2D range (from Section to EndSection)
	Name       string            // Section or RegisterOutput target name
	Pipelines  Pipeline          // Pipeline transform actions (e.g. ForEachLine)
	Nodes      []Node            // Child nodes
	TypedProps SectionProps      // Strongly typed properties struct
}

func (f *SectionFrameNode) GetID() string            { return f.ID }
func (f *SectionFrameNode) GetGeometry() geometry.Geometry { return f.Geo }
func (f *SectionFrameNode) GetProperties() map[string]any {
	return map[string]any{
		"name":     f.Name,
		"pipeline": f.Pipelines,
	}
}
func (f *SectionFrameNode) String() string {
	return fmt.Sprintf("SectionFrameNode[%s] Name:%s Geo:%s Nodes:%d", f.ID, f.Name, f.Geo, len(f.Nodes))
}

// -----------------------------------------------------------------------------
// 5. ACTION STROKE: ActionNode
// -----------------------------------------------------------------------------

// ActionNode represents a single directive action stroke (e.g. Replace, Strip, Remove, MacroCall, Import, InsertFragment).
type ActionNode struct {
	ID         string            // Deterministic SHA-256 ID
	Geo        geometry.Geometry // Physical 2D location (zero for synthetic pipeline nodes)
	Name       string            // e.g. "Replace", "Strip", "Remove", "MacroCall", "Import", "InsertFragment"
	Properties map[string]any    // Evaluated directive properties map
	TypedProps any               // Strongly typed properties struct (ImportProps, ReplaceProps, etc.)
}

func (a *ActionNode) GetID() string            { return a.ID }
func (a *ActionNode) GetGeometry() geometry.Geometry { return a.Geo }
func (a *ActionNode) GetProperties() map[string]any { return a.Properties }
func (a *ActionNode) String() string {
	return fmt.Sprintf("ActionNode[%s] Name:%s Geo:%s Props:%d", a.ID, a.Name, a.Geo, len(a.Properties))
}
