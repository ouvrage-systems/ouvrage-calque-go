# ADR 004: Decoupling Lexer and Transformer via Calque IR

## Status
Accepted

## Context
We want `ocalque` to be highly versatile and support:
1. Various programming and configuration languages (Go, C, Bash, Python, SQL, HTML, YAML, TOML) using their native comment markers.
2. Structured formats without comments (CSV, JSON, XML, third-party vendor scripts) where inlines or comments are prohibited.

If the core transformation engine is coupled with the parsing of native comment syntaxes and regular expressions, the engine becomes bloated, fragile, and hard to extend. Every new language or formatting constraint would require modifying the core compiler code.

We need a clean separation of concerns to keep the engine language-agnostic, portable, and easy to test.

## Decision
We decide to decouple the compiler's frontend (Lexer/Parser) from its backend (Transformation Engine) using a structured intermediate representation named **`Calque IR`**.

### 1. The Linear Calque IR Schema (The Recipe)
The `Calque IR` is represented as a structured YAML or JSON document containing a flat list of directives. The schema is defined as follows:

```yaml
source_hash: "sha256_of_original_source_file"
directives:
  - description: "Clear comment describing the purpose of this mutation"
    id: "deterministic_hash_uuid"  # Calculated as a hash of the directive metadata for drift detection
    ref:
      line_nb: 12                  # The physical target line in the source file
      indent: 4                    # The local indentation offset at this line
      action: "inject_after"       # Position policy: "inject_before" | "inject_after" | "replace_line"
    type: "stdlib:import"          # Namespaced action type: stdlib:* or geometry:*
    scope: "db_setup"              # Named boundary scope for stack nested validation
    properties:                    # Dict of specific properties for this directive
      source: "prod_secrets.json"
      mode: "insert"
```

### 2. Decoupled Frontend Lexers
* **Comment Lexers**: Built for different comment styles (e.g. `#`, `//`, `<!-- -->`, `;`, `::`). They parse the source file, extract the directives from comments, map them to line numbers/indents, and compile them into this standard `Calque IR` schema.
* **Companion Files (`.ocalque`)**: For formats without comments (like CSV or raw JSON) or upstream vendor files that must not be modified, the developer writes this exact same `Calque IR` schema in a separate companion file (e.g. `vendor_config.json.ocalque`). 
* **Virtual Injections**: By using the `ref.action` properties `inject_before` and `inject_after` in the companion file, the compiler can perform virtual injections and modifications on raw, pristine source files without inserting comments.

### 3. Backend Transformer
The core transformation engine is entirely blind to comments, language syntax, or regexes. It takes the raw source file and the compiled `Calque IR` as inputs, parses the IR into the Semantic AST (using stack nesting validation), and performs line-oriented operations (stripping, inlining, offset-shifting) to output the final target file.

## Consequences
* **Zero-Modification of Upstream Files**: SREs can customize vendor configurations AOT without adding comments to the vendor files.
* **Extensible & Plug-and-Play**: Supporting a new file format only requires writing a simple comment lexer that outputs the standard `Calque IR` schema. The backend transformer code remains untouched.
* **Drift Detection**: By comparing the deterministic `id` hashes of the compiled IR against the running system, SRE pipelines can detect and block unauthorized configuration drift.
* **High Testability**: We can write unit tests for the core transformer by passing mock `Calque IR` YAML files directly, without needing to create mock source files with comments.
