# Meeting: 2026-07-14 - Syntax Evolution: Stack-Based vs. Declarative Annotations

## Metadata
* **Date**: July 14, 2026 (session extended until July 15, around 01:15)
* **Topic**: Exploration of the evolution of `ocalque` directive syntax and comparison of design paradigms.
* **Status**: Active exploration - several design decisions emerged but still need to be formalized.

---

## 1. Starting Point: The Stack / Assembler Model

The first `ocalque` design used an imperative, stream-assembler style model:
linear geometric directives (push/pop indentation, strip lines, replace next line)
organized into implicit stacks.

```yaml
# @ocalque:stdlib:if<prod_db> --- "{{ env.NAME }}" == "production"
#   @ocalque:geometry:strip direction="next"
value: "postgresql://..."
#   @ocalque:geometry:strip direction="next"
#   @ocalque:replace_line --- value: {{ env.DB_URL }}
# @ocalque:stdlib:else<prod_db>
value: "sqlite://dev.db"
# @ocalque:stdlib:fi<prod_db>
```

**Identified issues**:
- Excessive verbosity (5-10 directive lines for 1 line of useful code).
- High cognitive load in PRs: the reviewer has to "compile it in their head".
- Manual arithmetic for indentation offsets (pushd/popd).
- Fragile implicit scopes -> silent nesting errors.

---

## 2. Pivot: Toward a Declarative Annotation Model

**Inspiration**: Rust (`#[cfg(...)]`), Kubernetes Annotations, Python Decorators, Java `@Annotation`.

### 2.1 General Principle
Directives become structured annotations, with a typed grammar, attached
directly to the target code block (the line or block immediately after the annotation).
Each language lexer strips the comment prefix (`#`, `//`, `<!-- -->`, `::`)
to expose a single, identical grammar to the evaluation engine.

### 2.2 The Golden Path (common cases)
```yaml
# @ocq:If(cond=(env.NAME == "production"))
value: "postgresql://prod-db:5432/audit"
# @ocq:Else
value: "sqlite://dev_database.db"
# @ocq:EndIf

# @ocq:Loop(in=env.PORTS, var="port")
- containerPort: 8080
# @ocq:EndLoop
```

### 2.3 Typed End Tags (`EndIf`, `EndMacro`, `EndLoop`)
Inspired by the HTML DOM and Bash (`fi`, `done`, `esac`):
each construct has its own typed semantic closing tag.
Benefits:
- The compiler validates **type + name** at close time.
- Vertical readability in large files or Git diffs.
- Trivial AST construction for the future LSP.

### 2.4 `@ocq` Prefix
Unified naming across the Ouvrage ecosystem (`ostream`, `olutrin`, `ogate`, ...).
The `o` prefix asserts the Ouvrage identity across all tools.

---

## 3. Scope Management

### 3.1 Implicit Scopes (Golden Path)
Each block directive initializes its own named scope under the hood.
Internal naming convention: `ocq-<type>--<name>`.
Example: `@ocq:Macro(name="hw_port_setup")` -> internal scope `ocq-macro--hw_port_setup`.

Child annotations inherit and propagate this scope automatically.
Explicit closing if needed: `@ocq:EndMacro(name="hw_port_setup")`.

### 3.2 The Generic Container `@ocq:Section`
Analogous to an HTML `<div>`. A neutral container with no execution logic of its own.
Lets you group blocks and attach geometric profiles or custom rules to them.

---

## 4. Macros & Profiles

### 4.1 Macro Definition
```c
// @ocq:Macro(name="hw_port_setup", use="c_function", args=["gateway_addr", "baud", "alias"])
void _define_hw_port_setup() {
    // @ocq:Replace(with='INIT_HW_PORT(${gateway_addr}, ${baud});')
    INIT_HW_PORT("0x3F8", 115200);
}
// @ocq:EndMacro(name="hw_port_setup")
```

### 4.2 Geometric Profiles (`use="profile"`)
Profiles (for example `c_function`, `bash_block`) encapsulate reusable geometric rules
(stripping the wrapper signature, indentation calculation, etc.).
The compiler remains blind to the target language.
The community can contribute profiles without touching the compiler core.

---

## 5. Rules with a `when` Clause (Emerging Decision)

Concept prototyped at the end of the session. Distinct from the sequential pipeline:

- **Pipeline**: executes each step unconditionally on every entity (like chained `sed`).
- **Rules**: independent rules with a conditional `when` clause (like CSS / Ansible `when:`).

```yaml
# @ocq:Section(name="port_gen", rules=[
#   Rule(Replace(pattern="MOCK_PORT", with="${port}"), when=(port.protocol == "tcp")),
#   Rule(Strip(), when=(port.internal == true)),
#   Rule(Append("  # auto-generated"), when=(env.DEBUG == true))
# ])
# @ocq:Loop(in=env.PORTS, var="port")
- containerPort: MOCK_PORT
# @ocq:EndLoop
# @ocq:EndSection
```

Each `Rule` is independent and commutative. Evaluated on each entity in the scope.
**Status**: prototype - to be formalized in a future RFC.

---

## 6. `ocalque` as a Constrained Mini-Language (DSL)

### 6.1 Acceptance of the DSL Paradigm
`ocalque` is evolving into a full declarative DSL. This is intentional and consistent
with Ouvrage's "Anything Declarative" philosophy.

### 6.2 Security Barriers by Design (to be formalized)
Inspired by Dhall, CUE, and Jsonnet:

```
Allowed:
  ✅ Constants (no mutable variables)
  ✅ Execution context (env.*, args.*, loop vars, scope constants)
  ✅ Pure boolean expressions (conditions, when)
  ✅ String interpolation (${var})
  ✅ Limited pure functions (defined(), len(), contains()...)

Forbidden by design:
  ❌ Network / database / filesystem access
  ❌ Mutable variables (no state between two rules)
  ❌ Recursion (no recursive macro)
  ❌ Importing arbitrary code
  ❌ Infinite loops (iterations bounded by finite collections)
```

Selling point in critical environments (Siemens cRSP, banking, defense):
**deterministic compilation, no possible side effects, auditable.**

---

## 7. Positioning vs. Nix

Nix generates configuration from a pure functional DSL.
`ocalque` **augments** existing files in their native language (C, Bash, YAML, WiX...).
The two are not in competition: they are fundamentally different paradigms.
Key differentiator for `ocalque`: the **Shadow Pattern** (a source file that is locally
executable with mocks, AOT-compiled for production). Nix has no equivalent.

---

## 8. LSP Roadmap

A `ocalque` LSP is identified as a **phase 2 feature**, after the reference Go
implementation. It would address the main remaining onboarding barrier (understanding
the Shadow Pattern) through:
- Inline AOT preview (code lens hover).
- Real-time scope validation.
- Directive autocompletion based on file type.
- Highlighting of Shadow lines (local mock vs. production code).

---

## 9. Open Questions & Next Steps

1. Formalize the grammar for `Rules` with `when` in a dedicated RFC (RFC 002?).
2. Define the canonical list of Rule primitives (`Replace`, `Strip`, `Indent`, `Append`, `Filter`).
3. Specify the formal barriers of the DSL (RFC 003?).
4. Write the reference Go compiler implementation.
5. Update the README with an introduction to the Shadow Pattern.
