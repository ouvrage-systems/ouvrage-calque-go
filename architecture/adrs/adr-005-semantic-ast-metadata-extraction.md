# ADR 005: Semantic AST, Namespaced Directives, and Geometric Calibration

## Status
Superseded by [ADR 008](adr-008-banning-mutable-variables-and-ast-matrix.md)

## Context
Traditional template engines treat code and configuration files as flat, meaningless text streams. When generating large-scale infrastructure scripts or SRE runbooks:
1. **Mismatched Nested Blocks**: Nested loops and conditional branches (`if`/`for`) using generic, anonymous boundaries (like `end` or `endif`) are prone to nesting errors that are hard to debug at compile time.
2. **Lack of Language-Agnostic Geometric Control**: SRE configurations (especially YAML or Python) are extremely sensitive to indentation. Traditional template engines force developers to write verbose, error-prone manual spacing arithmetic on every single line (e.g. `{{ ' ' * i }}`).
3. **Lack of Semantic Traceability**: Auditors cannot easily verify the dependencies, owners, or target environments of generated scripts, leading to configuration drift.

We need a unified, language-agnostic way to parse, validate, and transpile templates AOT while preserving absolute syntax validity in local dev environments.

## Decision
We decide to make the **`Calque`** IR a structured **Semantic AST (Abstract Syntax Tree)** driven by explicit namespaced directives, an explicit indentation stack, and geometric calibration.

---

### 1. Namespaced Directives
To ensure clean modularity and allow future community extensions, all directives must be explicitly namespaced under two core namespaces:

* **`geometry:` (Physical document space)**:
  Handles 2D line transformations, spacing, and annotations without any awareness of target language syntax:
  * `geometry:indent:pushd value="[expr]"` / `geometry:indent:popd`: Explicitly drives the compiler's indentation stack.
  * `geometry:strip direction="next"|direction="following" lines=N`: Erases local dev wrappers or fallbacks.
  * `geometry:ref name="[name]"`: Defines a coordinates anchor in the document.
* **`stdlib:` (Logical compile-time flow)**:
  Handles variables, loops, conditional rendering, macros, and file inclusion:
  * `stdlib:if<scope>`, `stdlib:elif<scope>`, `stdlib:else<scope>`, `stdlib:fi<scope>`
  * `stdlib:for<scope>`, `stdlib:count<scope>`, `stdlib:end<scope>`
  * `stdlib:macro:define<name>`, `stdlib:macro:arg<name>`, `stdlib:macro:call<name>`
  * `stdlib:import`, `stdlib:meta`, `stdlib:var:define`, `stdlib:var:set`

---

### 2. Explicit Indentation Stack & Geometric Calibration
To eliminate magic compiler heuristics and keep the compiler simple, we reject automatic indentation guesses. Instead, we introduce an explicit indentation stack:
* **The Stack Equation**: The target indentation of any generated line is calculated as:
  $$I_{final} = I_{call} + I_{local} + \sum(\Delta_{stack})$$
* **Implicit Macro Reference**: Every macro definition `@ocalque:stdlib:macro:define<name>` acts as an implicit reference anchor. Inside the macro body, the compiler exposes `macro.self.ref.indent` (and `macro.<name>.ref.indent` externally).
* **Wrapper Calibration Pattern**: To strip dev-time wrappers (like functions) and align the output with the production call site indentation:
  ```bash
  # @ocalque:stdlib:macro:define<my_macro>
  # @ocalque:geometry:strip direction="next"
  _my_wrapper() {
      # @ocalque:geometry:indent:pushd value="{{ macro.self.ref.indent - self.indent }}"
      echo "Useful line"
      # @ocalque:geometry:indent:popd
      # @ocalque:geometry:strip direction="next"
  }
  # @ocalque:stdlib:macro:end<my_macro>
  ```

---

### 3. File Imports and Geometric Slicing
To support modular template assembly without introducing custom fragment annotations in libraries, we use coordinates-based geometric slicing:
* **Evaluation Mode (`mode="eval"`)**:
  Loads references and macros of an external file into a dedicated namespace:
  ```bash
  # @ocalque:stdlib:import source="lib.sh" mode="eval" as="my_lib"
  ```
  *Namespace Isolation*: All imported references are accessible under `my_lib.ref.<anchor_name>`. All imported macros are called via `my_lib.macro.<macro_name>`.
* **Insertion Mode (`mode="insert"`)**:
  Inlines a physical slice of the external file, delimited by the line numbers of two references resolved at compile time:
  ```bash
  # @ocalque:stdlib:import source="lib.sh" mode="insert" from="{{ my_lib.ref.start.line_nb }}" to="{{ my_lib.ref.end.line_nb }}"
  ```
  The compiler automatically aligns the indentation of the inserted lines to the import statement's indentation using the stack.

---

### 4. Include Guards & Type-Safe Variables
To manage variables and prevent double-import collisions, we introduce template variables with explicit compile-time typing:
* **Variable Definition**:
  `# @ocalque:stdlib:var:define name="port" type="int" value="8080"`
  Supported types are `int`, `float`, `string`, `bool`, `list`.
* **Variable Re-assignment**:
  `# @ocalque:stdlib:var:set name="port" value="{{ var.port + 1 }}"`
  The expression must resolve to the declared type; type mismatches halt the compilation AOT.
* **Include Guards**:
  Using the `defined()` compile-time function, we can prevent duplicate imports:
  ```bash
  # @ocalque:stdlib:if<guard> --- not defined(var.LIB_IMPORTED)
  # @ocalque:stdlib:var:define name="LIB_IMPORTED" type="bool" value="true"
  ...
  # @ocalque:stdlib:fi<guard>
  ```

---

### 5. Metadata and AST Graph Extraction
SRE automation can parse and export the compiled `Calque` AST block tree as a structured JSON/YAML graph.
* **Metadata annotations**: `# @ocalque:stdlib:meta key="val"` attaches metadata to the active block node.
* **Traceability (Sourcemaps)**:
  * The compiler generates an `audit-mutations.jsonl` file mapping every target line back to its source template line.
  * A compiler flag (`--sourcemap`) injects `# @ocalque:src template.sh:line` comments in the output for runtime shell debugging.

## Consequences
* **Absolute Universality**: Works with any language (Go, Python, C, Bash, HTML, KCL, XML, Windows Batch) by adapting comment markers.
* **Zero-Magic Compiler**: The Go compiler remains a simple, extremely fast, stack-based lines transformer.
* **Compile-Time Safety**: Type errors, nesting errors, and import guard conditions are resolved and validated AOT before deployment.
