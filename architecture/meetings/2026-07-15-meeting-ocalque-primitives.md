# Meeting: 2026-07-15 - Brainstorming: Primitives, Blocks, and Instructions

## Metadata
* **Date**: July 15, 2026
* **Topic**: Continuation of the July 14 reflection. Free-form brainstorming on compiler primitives, the Block/Instruction model, and the `@ocq` grammar.
* **Status**: Active exploration - emerging decisions to formalize in an RFC.

---

## 1. Foundation: The HTML `div` / `span` Analogy

The mental model chosen to organize `@ocq` directives is the HTML DOM:
- **`div`** -> block container (multiple lines, explicit open/close).
- **`span`** -> single-target instruction (one line, no closing tag).

This interface is generic enough to support directive evolution
without changing the base model. Specific directives (`Loop`, `If`,
`Macro`, `Replace`, etc.) are just semantic variants of this interface.

---

## 2. Fundamental Distinction: Blocks vs. Instructions

### 2.1 Blocks (type `div`)
Directives that **wrap multiple lines**. They open and close
with a typed End tag. They can carry `rules`, a `use` (geometric profile),
and initialize an implicit scope.

| Directive | Closing tag | Role |
|:---|:---|:---|
| `@ocq:Section(name=...)` | `@ocq:EndSection` | Generic container (the "div") |
| `@ocq:If(cond=...)` | `@ocq:EndIf` | Condition on a block |
| `@ocq:Loop(in=..., var=...)` | `@ocq:EndLoop` | Iteration over a collection |
| `@ocq:Macro(name=..., args=[...])` | `@ocq:EndMacro` | Reusable macro definition |

### 2.2 Instructions (type `span` / single-shot)
Directives **without closing tags**. They target only the **next line of code**
that follows immediately in the file. Self-contained, single-shot.

| Directive | Role |
|:---|:---|
| `@ocq:Replace(with=..., if=...)` | Replaces the next line with the evaluated value |
| `@ocq:Strip(lines=N, if=...)` | Removes the next N lines |
| `@ocq:Inject(content=..., if=...)` | Inserts content without an existing target |
| `@ocq:Indent(delta=N)` | Shifts the indentation of the next line |
| `@ocq:Call(macro=..., args...)` | Invokes a macro at the current position |

---

## 3. Absolute Rule: Dedicated-Line Directives Only (Never Trailing)

Every `@ocq` directive must be written on its **own dedicated comment line**,
always **before** its target. Never as a trailing end-of-line comment.

```yaml
# Correct
# @ocq:Replace(with="value: ${env.DB_URL}", if=(env.NAME == "production"))
value: "sqlite://dev.db"

# Forbidden
value: "sqlite://dev.db"  # @ocq:Replace(with="value: ${env.DB_URL}")
```

**Reason**: Trailing comments are ambiguous across target languages (strings
may contain `#` or `//`). The rule "directive = clean line before the target"
guarantees a simple, fast, unambiguous lexer.

---

## 4. Open Questions for the Next Session

1. Canonical and exhaustive list of Instructions (are `Replace`,
   `Strip`, `Inject`, `Indent`, `Call` enough?).
2. Can `rules` be attached directly to a `Loop` (without an enclosing Section)?
   -> Likely answer: yes, via inline `rules=[...]` on the Loop.
3. Can `Section` have a `repeat=` property (implicit Loop alias)?
4. RFC 002: Formalization of `Rules` with a `when` clause.
5. RFC 003: Formal barriers of the DSL.

---

## 5. Calque Architecture (IR): Stroke Production

### 5.1 The Conceptual Model

The **Calque** is a transparent layer that overlays the source file.
It contains **only the strokes** (`@ocq` directives). Zero source-code lines.

To compile, we combine:
```
Source (code lines) ⊕ Calque (positioned strokes) -> Output
```

The calque is **ephemeral and derived**: always regenerated from the source.
It is not a persistent source of truth (except for companion `.ocalque` files).

### 5.2 Two Processing Passes

**Pass 1: Extraction (Raw)**
Line-by-line scan. We identify multiline `@ocq` comment blocks,
strip the comment prefixes (`//`, `#`, `<!-- -->`) and collapse
the multiline block into a single raw string. No semantic interpretation.

An extracted stroke = `{ line_start, line_end, target_line, raw_string }`.

```
// @ocq:Macro(        ┐
//   name="hw_port",  │ strip "//" + join -> "@ocq:Macro(name='hw_port', use='c_function')"
//   use="c_function" │
// )                  ┘
```

**Pass 2: Parsing (Structured)**
The `raw_string` values extracted in pass 1 are parsed into typed objects:
```
raw: "@ocq:Replace(if=(env.NAME == 'production'), with='value: ${env.DB_URL}')"
  -> Replace { cond: (env.NAME == "production"), with: "value: ${env.DB_URL}", target: L23 }
```

Parsed calque + source = inputs to the compilation engine.

### 5.3 Two Stroke Positioning Modes

**Static Mode** (generated automatically from inline `@ocq` directives in the source)
```json
{ "line_start": 42, "line_end": 44, "target_line": 45 }
```
Fast and trivial. Fragile if the source is edited (lines added/removed).
Acceptable because the calque is always regenerated at compile time.

**Anchor Mode** (used in companion `.ocalque` files for vendor/upstream sources)
```json
{
  "anchor": {
    "pattern": "value: \"sqlite://dev.db\"",
    "occurrence": 1
  }
}
```
The compiler searches for the pattern in the source and attaches to the specified
occurrence. Resilient to vendor edits (added/shifting lines).
`occurrence` handles ambiguity when the same pattern appears multiple times.

| Mode | Positioning | Main usage |
|:---|:---|:---|
| `StaticStroke` | `line_start`, `line_end`, `target_line` | Inline `@ocq` in the source |
| `AnchorStroke` | `pattern` + `occurrence` | Companion `.ocalque` (vendor/upstream) |

Both are resolved into positional strokes before the compilation pass.
The engine only sees one resolved stroke, regardless of its origin mode.

### 5.4 Note on Companion `.ocalque` Files
For vendor files (read-only, no comments allowed), the companion
`.ocalque` is the only source of strokes. In this case:
- The calque is **written by hand** (not generated).
- **Anchor mode** is recommended for resilience to vendor updates.
- A **hash of the target line** can be added for drift detection.

