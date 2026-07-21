# Meeting: 2026-07-18 - Taxonomy of AST Nodes and Grammar Primitives

## Metadata
* **Date**: July 18, 2026
* **Participants**: gpineda (Ouvrage Lead / Architect), Gemini (AI Exoskeleton / Design Partner)
* **Topic**: Standardizing the vocabulary for `@ocq` compiler primitives (Canvas, Strokes, Frames) to stabilize the core parser and AST design in Go.
* **Status**: Decided and locked.

---

## 1. The Design Ontology: Drafting Craftsmanship

To maintain alignment with the Ouvrage guild's open-core philosophy (blending German engineering precision with French design and layout elegance), we establish a physical drafting-table metaphor:

* **Canvas**: The underlying, read-only source document (raw native code or configuration).
* **Calque**: The transparent overlay containing our compiler instructions.
* **Stroke**: The atomic physical unit of drawing. Every single comment line starting with `@ocq:` is a **Stroke** (a pencil line drawn on the calque). We reject generic terms like "Annotation" or "Directive" to avoid conceptual bloat.
* **Frame**: The logical, bounded 2D block. A Frame is not written directly; it is *induced* by boundary Strokes and encloses a range of Canvas lines, sub-frames, and action strokes.

This terminology is professional, descriptive, and matches the 2D geometric operations (line heights, indentation offsets) computed by the compiler.

---

## 2. Stroke Roles (Syntax Classification)

During the line-by-line extraction pass, every **Stroke** is parsed and assigned a specific **StrokeRole** that defines how it affects the AST assembly:

1. **`OpenStroke`**: Opens a new Frame boundary and instantiates a nested scope (e.g. `@ocq:If`, `@ocq:Loop`, `@ocq:Macro`).
2. **`CloseStroke`**: Closes and seals the active Frame boundary (e.g. `@ocq:EndIf`, `@ocq:EndLoop`, `@ocq:EndMacro`).
3. **`SplitStroke`** (or *Divider Stroke*): Partitions the internal space of a Frame into multiple branches/areas without closing the active scope (e.g. `@ocq:Else`, `@ocq:Elif`).
4. **`ActionStroke`** (or *Inline Stroke*): A single-shot instruction targeting the immediate next line of the Canvas (e.g. `@ocq:Replace`, `@ocq:Strip`, `@ocq:Inject`).

---

## 3. AST Structs Implementation in Go

This classification translates into a clean, decoupled implementation in Go. In the final AST, boundary Strokes are folded into the `Frame` object, while only `ActionStrokes` remain as active leaf `Stroke` nodes:

```go
package ast

// Node is the common interface for any structural component in the compiled Calque.
type Node interface {
	StartLine() int
	EndLine() int
}

// StrokeRole defines the parser classification for a physical stroke.
type StrokeRole int

const (
	OpenStroke StrokeRole = iota
	CloseStroke
	SplitStroke
	ActionStroke
)

// Stroke represents a physical annotation line parsed from the comments.
type Stroke struct {
	Role       StrokeRole
	Type       string            // e.g., "Replace", "Strip", "If", "Else", "EndIf"
	Properties map[string]string // Key-value arguments (e.g., cond, with)
	LineNumber int               // The physical line in the file
}

func (s *Stroke) StartLine() int { return s.LineNumber }
func (s *Stroke) EndLine() int   { return s.LineNumber }

// Frame represents a logical bounded block of code/config (e.g., If/Else, Loops).
type Frame struct {
	Type        string            // e.g., "If", "Loop", "Macro"
	ScopeName   string            // Named boundary scope for validation
	Properties  map[string]string // Inherited properties from the opening Stroke
	OpenStroke  *Stroke           // The Stroke that initiated this block
	CloseStroke *Stroke           // The Stroke that sealed this block
	SplitStrokes []*Stroke        // Intermediate branch dividers (Else/Elif)
	Children    []Node            // Nested Frames, ActionStrokes, or CanvasLines
}

func (f *Frame) StartLine() int { return f.OpenStroke.LineNumber }
func (f *Frame) EndLine() int {
	if f.CloseStroke != nil {
		return f.CloseStroke.LineNumber
	}
	return f.OpenStroke.LineNumber
}

// CanvasLine represents a raw line of code/config from the original source file.
type CanvasLine struct {
	Content    string
	LineNumber int
	Indent     int
}

func (c *CanvasLine) StartLine() int { return c.LineNumber }
func (c *CanvasLine) EndLine() int   { return c.LineNumber }

// Canvas represents the root of the parsed document, containing the target lines and Calque AST.
type Canvas struct {
	SourcePath string
	SourceHash string
	RootNodes  []Node
}
```

---

## 4. Compiler Lifecycle Benefits

By adopting this nomenclature:
* **Parser Simplicity**: Pass 1 scans the `Canvas` and produces a list of `Strokes` with their `StrokeRole`.
* **Structural Safety**: Pass 2 builds the AST by nesting `Node` elements into `Frames` using a simple state machine driven by `OpenStroke`, `SplitStroke`, and `CloseStroke`. 
* **Geometric Calibration**: Indentation offset calculations inside `Frames` are driven by comparing the `OpenStroke` indentation against nested `CanvasLine` indentations.

---

## 5. Scope Alignment: Core (ocalque) vs. Carton (Orchestration)

To prevent scope creep and guarantee a swift delivery of the V1 reference implementation, the committee aligned on the division of responsibilities for multi-file generation:

### 5.1 The Docker / Docker Compose Analogy
We establish a clear operational distinction between unit compilation and project-wide orchestration:
*   **`ocalque` (equivalent to `docker`)**: The unit compiler. It is entirely blind to other files. It takes a single `Canvas`, applies its `Calque` (Strokes and Frames), and renders the output.
*   **`ocalque carton` (equivalent to `docker-compose`)**: The project orchestrator. It reads the `carton.yaml` manifest, manages global context variables, loops over targets, and invokes the `ocalque` unit compiler for each file.

### 5.2 Release Roadmap
*   **V1 Release**: Purely file-by-file compilation (`ocalque compile <source> <destination>`). The `carton` layer is excluded. SREs can orchestrate multi-file setups using standard scripting (Makefiles, Bash loops).
*   **V2 Release**: Introduction of `carton` orchestration and the `carton.yaml` manifest for declarative multi-file environments.
