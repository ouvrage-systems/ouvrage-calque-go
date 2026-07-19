# Meeting Notes: ocalque Design & Architecture Session (AOT Semantic Templating)

This document summarizes the architectural debate, the evolution of design choices, and the technical decisions made during the pair-programming session for the **`ocalque`** AOT (Ahead-of-Time) template engine and transpiler.

---

> [!NOTE]
> The participants of this meeting (Gérard, Hans, Marc, Alex) are **Interactive Design Personas (Virtual Engineering Committee)** simulated during the pair-programming session. They are not autonomous AI agents or bots. For details on their profiles, see [personas.md](./personas.md).

## 1. General Info & Participants

* **Date**: June 18, 2026
* **Participants**:
  * **Guillaume** (Creator of the Ouvrage guild)
  * **Gérard** ("Bash-Guru": Traditional SRE/Ops SysAdmin, proponent of sed/awk/heredocs)
  * **Hans** ("Windows-Ops": Siemens cRSP SRE, specialist in XML/WiX/MSI/PowerShell)
  * **Marc** ("C-Bas-Niveau": C/Zig systems developer, purist of performance and debugging)
  * **Alex** ("CNCF-boy": Kubernetes/GitOps architect, specialist in Helm/Kustomize/YAML)

---

## 2. Evolution of Needs and Design Choices

The debate began with a comparison of traditional templating approaches (Jinja, Kustomize, sed) and our vision for a non-intrusive preprocessor.

### A. Rejection of Traditional Text-Based Rendering Engines (DSL)
* **Jinja/Helm Flaws**: They break development tools (LSP, auto-completion, formatters like `go fmt`). They introduce Server-Side Template Injection (SSTI) / RCE risks at runtime.
* **sed/regex Flaws**: Fragile under whitespace changes, unreadable for multi-line blocks, impossible to test locally without executing the production script.
* **Decision**: `ocalque` will be a **line-oriented AOT (Ahead-of-Time) engine**, completely agnostic of the target languages, whose build instructions are hidden in the native comments of each file.

### B. The Comment Trap and the "Active Shadow Mocks" Correction
* **Initial Error**: Putting production-specific code in XML block comments or standard code comments.
* **The Problem**: The IDE does not analyze or validate the content of comments, causing a loss of WiX auto-completion or C/Go compiler checks.
* **Chosen Solution (Active Shadow Mocks)**: Production code is written as an "active shadow" (a mock or valid placeholder, such as `placeholder.dll` or a fake namespace `[ouvrage.shadow]`). The code is 100% active and validated in local dev. The AOT engine `ocalque` replaces this shadow with real production values during the build phase.

### C. Strict Scope and Boilerplate Elimination (Bash/C)
* **The Problem**: Inlining functions containing `local` variables in Bash causes fatal syntax errors in the global scope of production.
* **Chosen Solution**:
  1. Priority given to direct injection (placing `{{ ... }}` directly inside commands/calls without intermediate variables).
  2. For necessary local dev boilerplate, the macro creator uses `ocalque`'s spatial primitives (e.g., `# @ouvrage:strip_matching pattern="local .*"`) to clean up their own code during compilation. `ocalque` remains simple and delegates validation to the host compiler/linter (`go build`, `shellcheck`).

### D. Auto-Calculation of Indentation Offsets
* **Decision**: The engine automatically calculates the difference in spaces between the annotation line's indent and the block's inner content indent. During unwrapping or importing, it shifts the lines geometrically without requiring manual parameters.

---

## 3. The Semantic Dimension: The "Calque" (SSoT)

Throughout the debate, the internal IR (Intermediate Representation) was renamed **`Calque`** (the only structure named in French within the codebase, representing the Ubiquitous Language of the guild).

```go
type BlockNode struct {
    Name       string            `json:"name"`
    Type       string            `json:"type"`       // "block", "loop", "if"
    Metadata   map[string]string `json:"metadata"`   // "owner", "description"
    Children   []BlockNode       `json:"children"`
    SourceLine int               `json:"source_line"`
}

type Calque struct {
    SourceHash string      `json:"source_hash"`
    RootBlocks []BlockNode `json:"blocks"`
}
```

### A. Role of the Calque as the Infrastructure SSoT
* The `Calque` extracts named block annotations (`block:start name="X"`) and metadata (`// @ocalque:meta key="owner"`).
* It becomes the theoretical **Single Source of Truth** (SSoT) of what the code is supposed to configure.
* This structured file (JSON/YAML) can be exported to graph databases (like **Kuzu DB**) or visualization tools (like **Graphviz/Mermaid**) to generate live architecture diagrams without manual documentation drift.
* Enables semantic drift detection by comparing the theoretical calque graph with actual active operating system ports or processes.

### B. Handling Files Without Comments (CSV, TXT)
* **Decision**: Since these files cannot contain inline comments without breaking, they use a **companion file** (`.ouvrage` or `.calque`) containing the same IR directives.
* **Anti-Drift Security**: To prevent modifications of the CSV from shifting rows out of sync with the companion file, the companion file includes a **SHA-256 hash lockfile** of the source file. If the hash does not match, the CI fails.

---

## 4. Target Use Cases & Limitations

### Priority AOT Use Cases:
* **SRE / Ops**: Multi-environment Bash scripts testable locally.
* **Windows / Siemens cRSP**: XML WiX (MSI) files with secure Siemens components inlined AOT.
* **Systems / C / Zig**: Hardware register maps or USB HID configurations with native `#line` sourcemaps for gdb.
* **GitOps / K8s**: Complex YAML configuration files validated locally before commit.

### Explicit Exclusions (Out of Scope):
* **Dynamic Runtime Rendering**: Rendering HTML pages for web applications (React SSR, Flask/Express template engines). `ocalque` is not designed for runtime network latency or caching performance.
* **Complex Semantic Code Generation**: Generating full ASTs or parsers from grammars (use Lex/Yacc or ANTLR).
