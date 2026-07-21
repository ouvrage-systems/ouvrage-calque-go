# ADR 009: Five-Pass Compiler Lifecycle, Memory VFS, and Static Linker

## Status
Accepted

## Context
In ADR 006, we defined a 3-phase compiler lifecycle (Extraction, Assembly, Evaluation) operating on flat line mutations. This lifecycle was insufficient for:
1. Verifying cross-file passive imports and registry signatures before runtime evaluation to prevent silent runtime template failures.
2. Decoupling the physical disk workspace from compiler evaluation to allow fast, memory-only execution.

We need a formal, pipeline-based compiler lifecycle in Go that handles static dependency graphs, memory-only workspace isolation, and delegates multi-file bundle orchestration to Carton.

## Decision
We decide to transition the Ocalque compiler to a formal **5-Pass Compiler Lifecycle**, operating as a strict **Single-Input, Single-Output Stream Processor**, while delegating multi-file aggregation to the Carton orchestrator:

1. **Pass 1: Extractor (Scanner)**:
   Extracts comment annotations or raw `.ocq` statements into physical 2D geometry ExtractedAnnotations.
2. **Pass 2: Directive Parser**:
   Parses raw statements into Pratt AST trees and flat directive ParsedStrokes (`RoleOpen`, `RoleSplit`, `RoleClose`, `RoleAction`).
3. **Pass 3: Compiler (Compile Pass)**:
   Transforms Pratt directive strokes into strongly typed polymorphic AST nodes (`calque.Compile`). Builds hierarchical Frame trees (`IfFrameNode`, `SectionFrameNode`, `MacroFrameNode`, `LoopFrameNode`) with strongly typed property structs and executable domain expression objects (`CompareExpr`, `LogicalExpr`).
4. **Pass 4: Linker**:
   Resolves cross-file dependencies and passive imports via the VFS. Emits a standalone `kind: SymbolTable` resource containing the Directed Acyclic Graph (DAG) and namespaced symbol bindings (`InsertOutput(source="alias:registry")`), raising static linker errors on missing symbols or import cycles.
5. **Pass 5: Evaluator & Renderer**:
   Traverses the AST bottom-up (inside-out), evaluates conditions and loop contexts, processes frame pipelines on CanvasLines, and renders the final static output stream.
6. **Carton Orchestration & Memory VFS**:
   Ocalque does **not** manage physical file writes or target file paths. It processes a single input stream to a single output stream. Multi-file composition, target routing (`OutputFile` mapping), and registry piping (capturing `RegisterOutput` and injecting it into another stream) are managed exclusively by Carton operating on a Memory Virtual File System (VFS).
7. **Namespaced Registry Aggregation**:
   We reject global magic merging of imported files. Imported registry buffers are strictly namespaced under their import alias (`alias:registry_name`), avoiding name collisions.

## Consequences
* **Compile-Time Safety**: Dependency circles, missing imports, and broken registry links are flagged AOT by the Linker (Pass 4) before Pass 5 runs.
* **Hermetic and Parallel Execution**: Since evaluation runs on an in-memory VFS, the compiler is side-effect free and can safely process independent compilation pipelines in parallel.
* **Predictable Linking**: Namespaced resolution eliminates global configuration collisions, making it easy to track where each block of config was imported from.
