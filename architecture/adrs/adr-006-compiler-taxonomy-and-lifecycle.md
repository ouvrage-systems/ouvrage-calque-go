# ADR 006: Compiler Taxonomy and Lifecycle (Canvas, Calque, Stroke, Frame)

## Status
Accepted

## Context
In previous design iterations, terms like "Directive", "Annotation", "BlockNode", and "LeafNode" were used interchangeably. This created conceptual confusion between:
1. The physical representation of instructions in the source file (comments).
2. The logical AST representation assembled by the parser.
3. The original code file being mutated.

To build a clean, language-agnostic compiler in Go, we need a unified, domain-specific vocabulary (Ubiquitous Language) inspired by the Ouvrage guild's design ontology (craftsmanship drafting-table metaphor). This vocabulary must map cleanly to each phase of the compiler's lifecycle.

## Decision
We decide to adopt a unified taxonomy and partition the compiler execution lifecycle into three distinct, decoupled phases.

### 1. Unified Taxonomy

*   **Canvas**: The physical file at rest, and the base sheet of paper during compilation. It represents the underlying document (containing code and comments).
*   **Calque**: The transparent overlay layer containing our compiler instructions.
*   **Stroke**: The atomic physical unit of drawing. Every single comment line starting with `@ocq:` is a **Stroke** (a pencil mark drawn on the calque).
*   **Frame**: A logical, bounded 2D block. A Frame is not written directly; it is *induced* by a pair of boundary Strokes and encloses a range of Canvas lines, sub-frames, and action strokes.

### 2. Phase-by-Phase Compiler Lifecycle

We establish the following pipeline and naming transitions:

```
[ Canvas Source ] (At rest, containing raw @ocq comments)
       |
       v (Phase 1: Extraction - Lexer)
[ CanvasLines ] + [ Calque (flat Strokes categorized by StrokeRole) ]
       |
       v (Phase 2: Assembly - AST Parser)
[ Calque AST (composed of nested Frames & ActionStrokes) ]
       |
       v (Phase 3: Evaluation & Rendering - Engine)
[ Rendered Canvas ] (Output file: static, clean, production-ready)
```

#### Phase 1: Extraction (Lexing)
*   **Input**: Raw `Canvas` file.
*   **Operations**: The lexer scans the file, strips comment markers, and separates the code from the instructions.
*   **Entities**:
    *   **CanvasLines**: The raw lines of code (clean Canvas).
    *   **Calque**: The flat container of instructions.
    *   **Strokes**: The individual instructions, each assigned a **StrokeRole**:
        *   `OpenStroke`: Opens a Frame boundary (e.g. `@ocq:If`).
        *   `CloseStroke`: Closes a Frame boundary (e.g. `@ocq:EndIf`).
        *   `SplitStroke`: Divides Frame branches (e.g. `@ocq:Else`).
        *   `ActionStroke`: Executes a one-shot line mutation (e.g. `@ocq:Replace`).

#### Phase 2: Assembly (AST Parsing)
*   **Input**: Flat `CanvasLines` and the `Calque` of `Strokes`.
*   **Operations**: A stack-based parser balances boundary Strokes to assemble the tree structure.
*   **Entities**:
    *   **Calque AST**: The final semantic tree.
    *   **Frames**: Logical branch nodes representing blocks and scopes.
    *   **ActionStrokes**: Leaf nodes representing mutations.
    *   **CanvasLines**: Code lines grouped under their respective Frame or root context.

#### Phase 3: Evaluation & Rendering (Engine)
*   **Input**: `Calque AST` + Environment Variables/Context.
*   **Operations**: The engine evaluates conditions on `Frames`, resolves variables, applies `ActionStrokes` to `CanvasLines`, and aligns indentation.
*   **Output**: **Rendered Canvas** (the compiled output configuration file).

## Consequences
*   **Conceptual Clarity**: SREs and developers use the exact same terminology when writing templates, reviewing PRs, and reading compiler errors.
*   **Go Code Cleanliness**: Structs in the Go codebase map 1:1 to these terms (`Canvas`, `Frame`, `Stroke`, `CanvasLine`), making the code self-documenting.
*   **Decoupled Parser Implementation**: We can write Pass 1 (Scanner) and Pass 2 (AST Assembler) to handle generic `Strokes` and `Frames` without knowing the logic of specific directives (e.g., how a Loop resolves). New directives can be added later by simply defining how their `ActionStroke` or `Frame` evaluates in Pass 3.
