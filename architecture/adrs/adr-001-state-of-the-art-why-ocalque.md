# ADR 001: State of the Art & Why ocalque

## Status
Proposed

## Context
Traditional configuration and code template engines (e.g., Jinja, EJS, Helm, Gomplate) interleave template DSL logic (such as `{{ if .Values.prod }}`) directly into the source files. While functional, this approach introduces several severe engineering drawbacks:
1. **IDE and LSP Disruption**: Inserting non-native template syntax breaks language parsers, syntax highlighting, autocompletion (LSP), linting, and automatic formatting (e.g., `go fmt`, `yamllint`).
2. **Broken Local Testability**: Templates cannot be compiled or run locally without a rendering step. Developers must often commit changes and wait for a CI/CD pipeline to verify template correctness, resulting in a slow feedback loop.
3. **Runtime Overhead & Security Risks**: Running a template engine in production invites Server-Side Template Injection (SSTI) vulnerabilities (e.g., Jinja sandbox escapes) and increases memory/CPU footprint.
4. **Fragility of Raw Text Replacements**: Ad-hoc preprocessors (like `sed` or regex-based shell scripts) are extremely fragile, error-prone when handling indentation (like in YAML), and difficult to maintain across multiple platforms (Windows/Linux).

We need a unified, language-agnostic way to handle environment-specific configurations and modular code generation while preserving the absolute syntactic integrity of the source files.

## Decision
We decide to build **`ocalque`**, an Ahead-of-Time (AOT) template engine and semantic compiler. It implements **Syntax-Preserving Templating** with the following constraints:
1. **Comment-based Directives**: All compilation and transformation instructions (e.g., loops, replacements, exclusions) must be written inside the native comment markers of the target file (`#` for shell/yaml, `//` for Go/C, `<!-- -->` for XML/HTML).
2. **Ahead-of-Time (AOT) Processing**: The template is evaluated and transformed at build/deployment time (CI/CD). The output is a pure, static file with zero template engine traces or runtime dependencies.
3. **Geographic Line-by-Line Transformations**: The engine operates as a stream transformer using spatial primitives (e.g., unwrapping blocks, stripping lines, auto-calculating indentation offsets) rather than global fuzzy regex searches.

## Consequences
* **Satisfied Toolchain**: Source files remain 100% syntactically valid in local dev environments. IDE linters and formatters function normally.
* **Local Runnability**: The source file can be run directly on a local laptop using local mocks or shadow targets without any pre-rendering step.
* **Zero Runtime Risk**: The production environment only executes static, audited, and optimized files, eliminating SSTI risks and rendering overhead.
* **Unix-like Simplicity**: `ocalque` delegates language semantics and validation to the target language's native compiler/linter (e.g., `go build`, `shellcheck`, `yamllint`) rather than embedding complex compilers itself.
