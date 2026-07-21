# ADR 008: Banning Mutable Variables, Pure AST Primitives Matrix, and Metadata Render Passes

## Status
Accepted

## Context
In ADR 005, we proposed namespaced directives, include guards, and mutable template variables (e.g., `stdlib:var:set`). However, allowing mutable state variables introduces temporal execution dependencies (order of execution dictates variable values), breaking spatial invariance (reordering blocks silently changes logic) and causing "Helm Hell" (unlintable, complex, unmaintainable configuration code).

Additionally, using ActionStrokes (like `Append` or `Replace`) inside block Frame bodies to generate new lines of text is semantically confusing, as ActionStrokes are mutations, not content generators. Finally, static metadata harvesting (Passes 1-3) fails to capture dynamic metadata injected during rendering passes (Pass 5).

## Decision
We decide to enforce strict statelessness, define a clean 3-category AST primitives matrix, and treat metadata harvesting as a rendering pass artifact:

1. **Banning Mutable Variables**:
   We reject all mutable variables and temporal states. No variable reassignment is allowed.
2. **Static Constant Binding (`Const` directive)**:
   We allow declaring immutable constants (`Const(name="...", value="...")`) bound to the Frame's lexical scope. Redeclaration of a constant symbol in the same scope raises a static `Compile-Time Redeclaration Error`.
3. **Loop Projections and Env Context**:
   Sequential increments (like network ports) are computed statelessly using loop metadata (e.g., `loop.index`) and mathematical projections (`${8000 + (loop.index * 5)}`). External configurations are read-only properties under the `env` namespace.
4. **AST Primitives Matrix**:
   We partition AST nodes into three strict semantic categories:
   * **Frames (Structure)**: `Section`, `Loop`, `If`, `OutputFile`, `RegisterOutput`.
   * **Strokes (Content Generators)**: `Line(text)`, `InsertFragment`, `InsertOutput`, `Call`.
   * **ActionStrokes (Mutations)**: `Replace`, `Strip`, `Append`, `Prepend` (strictly restricted to pipelines or inline target annotations).
5. **Metadata Render Pass**:
   To capture dynamic pipeline-generated metadata, the metadata extraction is a rendering pass artifact (Pass 5) that loads and parses the Canvas to evaluate all transformations, but intercepts `@ocq:Meta` nodes and discards the generated text payload, outputting only the collected JSON index.
   * *Fast-Path Optimization*: The compiler provides the `--ignore-runtime-metadata` flag to bypass Canvas loading and run only Passes 1-3 when the user knows their metadata is static.

## Consequences
* **State-Free Predictability**: Config templates remain 100% declarative, predictable, and easy to audit.
* **Separation of Concerns**: Complex data calculation logic is pushed upstream to the inventory generator (Python/Go/SQL), keeping Ocalque strictly focused on configuration layout compilation.
* **Mathematical Precision**: The compiler AST is clean, typed, and mathematically sound, with zero semantic overlap between content generation (`Line`) and line mutation (`Replace`/`Append`).
