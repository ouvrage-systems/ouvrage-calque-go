# ADR 002: Configuration Compilation vs. Text Generation

## Status
Proposed

## Context
For over a decade, the software industry has treated systems configuration (YAML, TOML, XML) and SRE deployment scripts (Bash, SQL) as plain, flat text. To manage variations across environments, developers adopted web-oriented template engines (such as Jinja or Go templates) designed for dynamic web-page rendering.

This introduced a fundamental architectural mismatch. Web templates are designed for high-throughput runtime interpolation based on ephemeral HTTP requests. Systems configurations, however, are static, structural blueprints that govern infrastructure topology, security policies, and compiler builds. 

> "The industry has confused 'generating text' with 'compiling a configuration'. By treating the template as an active syntactic entity and the Calque as a semantic AST, ocalque fills a decade-long vacuum. It is the first template engine designed by and for systems engineers, not web developers."

We need to establish a clear architectural distinction between runtime text generation and ahead-of-time configuration compilation.

## Decision
We decide to establish **`ocalque`** not as a text-renderer, but as a **Configuration Compiler**:
1. **Compilation Paradigm (AOT)**: `ocalque` treats the source template as a structured blueprint and compiles it into a target configuration. All transformations, parameter evaluations, and macro expansions are computed ahead-of-time (AOT) at build/deployment time (CI/CD).
2. **Strict Separation of Concerns**: We reject any runtime execution features. `ocalque` does not perform dynamic, request-bound rendering in production.
3. **Decoupled Validation**: We delegate syntax and type validation of the output to the target language's native toolchain (`gcc`, `go build`, `shellcheck`, `yamllint`) rather than attempting to parse all target syntaxes.

## Consequences
* **SRE-Centric Feature Set**: Design decisions prioritize safety, auditability (sourcemaps), and compliance (drift detection) over runtime rendering speed or web-specific features.
* **No Runtime Overhead**: Zero runtime dependency in production. The compiler only outputs static, validated, and optimized assets.
* **Compiler-Grade Tooling**: `ocalque` can build a structural intermediate representation (the `Calque` AST) representing the theoretical state of the system, enabling dependency analysis and graph generation (Mermaid/Kuzu).
