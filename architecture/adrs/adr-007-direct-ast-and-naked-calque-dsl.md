# ADR 007: Direct AST Compilation and Naked Calque DSL

## Status
Accepted

## Context
In ADR 004, we proposed compiling code comments into an intermediate JSON/YAML document named `Calque IR` before rendering. While decoupled, this approach creates unnecessary disk write overhead, introduces serialization complexity, and limits interactive developer tooling.

Additionally, when working on purely structural configurations (e.g., master orchestrations or compilation blueprints), developers should not be forced to write comments inside empty files. They need a dedicated file format that contains raw compiler instructions without host language syntax or comment markers.

## Decision
We decide to bypass any flat serialization to `Calque IR` and compile files directly into an in-memory Abstract Syntax Tree (AST). We also establish two distinct compilation modes and official file formats:

1. **Direct AST Parsing**:
   The compiler parser processes raw source files directly into a semantic AST in memory, which is then passed to the execution engine. This eliminates the intermediate `Calque IR` file format.
2. **The Naked Calque DSL (`.ocq` files)**:
   For files that contain *only* Ocalque directives (such as master scripts or standalone modules), we establish the `.ocq` file extension. In `.ocq` files:
   * Directives are written natively **without** comment prefixes (e.g., `Import(source="...")`, `If(cond=...)`).
   * Lines of text to output are declared using the explicit `Line("...")` Stroke.
3. **Inline Comment Mode**:
   For Canvas files (like `db.yaml` or `build.bat`), directives are embedded inside host comments (`# @ocq:` or `:: @ocq:`) to keep the source files 100% syntactically valid in local dev environments.

## Consequences
* **Enhanced Performance**: Bypassing intermediate disk serialization speeds up compilation, making it suitable for live pre-rendering.
* **Cleaner Master Templates**: Standalone scripts are written in clean, naked Ocalque syntax, avoiding the visual noise of prefixing every line with comment markers.
* **Unified AST Representation**: Both inline comments and `.ocq` naked files compile to the exact same in-memory AST nodes, keeping the backend evaluator completely unified.
