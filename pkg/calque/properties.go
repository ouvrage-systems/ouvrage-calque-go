package calque

import "github.com/ouvrage-systems/ouvrage-calque-go/pkg/syntax/pratt"

// -----------------------------------------------------------------------------
// STRONGLY TYPED ACTION & FRAME PROPERTY STRUCTS
// -----------------------------------------------------------------------------

// ImportProps defines typed parameters for Import directives.
type ImportProps struct {
	Source string `json:"source" yaml:"source" doc:"Path of the imported canvas fragment file"`
	As     string `json:"as" yaml:"as" doc:"Alias name for the imported fragment"`
}

// ReplaceProps defines typed parameters for Replace action strokes.
type ReplaceProps struct {
	Pattern string     `json:"pattern,omitempty" yaml:"pattern,omitempty" doc:"Sub-string pattern to target"`
	With    string     `json:"with" yaml:"with" doc:"Replacement string text"`
	When    *Condition `json:"when,omitempty" yaml:"when,omitempty" doc:"Guard condition for replacement"`
}

// InsertFragmentProps defines typed parameters for InsertFragment content strokes.
type InsertFragmentProps struct {
	Source string `json:"source" yaml:"source" doc:"Alias name of imported fragment to evaluate and insert"`
}

// InsertOutputProps defines typed parameters for InsertOutput content strokes.
type InsertOutputProps struct {
	Target string `json:"target" yaml:"target" doc:"Name of registered section output target"`
}

// StripProps defines typed parameters for Strip action strokes.
type StripProps struct {
	Lines int `json:"lines" yaml:"lines" doc:"Number of lines to strip"`
}

// RemoveProps defines typed parameters for Remove action strokes.
type RemoveProps struct {
	Lines int `json:"lines" yaml:"lines" doc:"Number of lines to remove"`
}

// MacroCallProps defines typed parameters for MacroCall action strokes.
type MacroCallProps struct {
	Name string       `json:"name" yaml:"name" doc:"Name of macro being invoked"`
	Args []pratt.Node `json:"args" yaml:"args" doc:"Arguments passed into macro"`
}

// SectionProps defines typed parameters for Section boundary frames.
type SectionProps struct {
	Name     string   `json:"name" yaml:"name" doc:"Section target name"`
	Pipeline Pipeline `json:"pipeline,omitempty" yaml:"pipeline,omitempty" doc:"Pipeline transform sequence"`
}

// MacroProps defines typed parameters for Macro boundary frames.
type MacroProps struct {
	Name     string   `json:"name" yaml:"name" doc:"Declared macro name"`
	Params   []string `json:"params" yaml:"params" doc:"Declared parameter names"`
	Pipeline Pipeline `json:"pipeline,omitempty" yaml:"pipeline,omitempty" doc:"Pipeline transform sequence"`
}

// LoopProps defines typed parameters for Loop boundary frames.
type LoopProps struct {
	ItemVar    string     `json:"itemVar" yaml:"itemVar" doc:"Item variable name"`
	IndexVar   string     `json:"indexVar,omitempty" yaml:"indexVar,omitempty" doc:"Index variable name"`
	Collection pratt.Node `json:"collection" yaml:"collection" doc:"Pratt AST evaluating to iterable collection"`
}
