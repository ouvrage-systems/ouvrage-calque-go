# RFC 001: The `@ocq` Declarative Annotation Grammar & Composition Model

## Status
Proposed

## Context
Traditional templating systems (like Jinja, Helm, or Go templates) interleave control directives directly into active source code, breaking syntax validation, IDE helpers, auto-formatters, and static analysis tools.

`ocalque` (`@ocq`) resolves this by executing as an Ahead-of-Time (AOT) compiler. It operates on comments and preserves syntax integrity. 

However, parsing target languages to find brackets/scopes (like C `{}` or Python indents) to delineate blocks is too complex and fragile for a language-agnostic compiler. Conversely, linear stack-based push/pop commands are verbose and hard to read.

We need a design that:
1. Decouples **physical text layout management** (stripping wrappers, indentation) from **logical behavior** (macros, loops, conditionals).
2. Explicitly bounds scopes using **Named Scopes** to eliminate nesting ambiguity during compilation and code reviews.

---

## Proposal: The Golden Path & On-the-Fly Named Scopes

We propose a declarative annotation model that uses the **Golden Path** (simple annotations) by default, while supporting advanced nesting and formatting rules by allowing developers to instantiate and bind **Named Scopes on the fly**.

### 1. The Golden Path (Standard Directives)
By default, directives target the immediate next line or block. No scope naming is required.

```text
# @ocq:If(cond=(env.NAME == "production"))
# @ocq:Loop(in=env.PORTS, var="port")
# @ocq:Else
# @ocq:Fi
```

---

### 2. Implicit Scope Naming & Scope Propagation
Block directives automatically initialize their own named scopes under the hood:
- A macro definition `@ocq:Macro(name="foo")` implicitly instantiates the scope name `ocq-macro--foo`.
- Inner annotations (e.g. `@ocq:Replace`) inherit this scope context automatically without manual binding.
- The block can be closed explicitly using the implicit name: `@ocq:Fi(scope="ocq-macro--foo")`, or implicitly using the stack closure: `@ocq:Fi`.

---

### 3. Generic Block Containers (The `div` Paradigm)
In addition to semantic elements (like `Macro`, `Loop`, `If`), `ocalque` provides a generic block container:

```text
# @ocq:Section(name="section_name", use="profile", ...)
```

Acting like a `<div>` tag in HTML, `Section` does not carry logical execution rules by default. Instead, it is used to group code blocks, apply physical layout profiles (`use="profile"`), and attach custom scope rules on the fly.

---

## Core Examples

### 1. YAML Conditionals (Golden Path)
```yaml
# @ocq:If(cond=(env.NAME == "production"))
value: "postgresql://prod-db:5432/audit"
# @ocq:Else
value: "sqlite://dev_database.db"
# @ocq:Fi
```

---

### 2. YAML Loop (Golden Path)
```yaml
# @ocq:Loop(in=env.PORTS, var="port", if=(env.NAME == "production"), replace="- containerPort: ${port}")
- containerPort: 8080
```

---

### 3. C Macros: Definition & Invocation (Implicit Scopes)

#### Macro Definition:
The `@ocq:Macro` directive instantiates an implicit named scope (`ocq-macro--hw_port_setup`) under the hood. Inner `@ocq:Replace` directives propagate this scope. The block can be closed explicitly using the implicit scope name.

```c
// @ocq:Macro(name="hw_port_setup", use="c_function", args=["gateway_addr", "baud", "alias", "trailing"])
void _define_hw_port_setup() {
    // @ocq:Replace(with='printf("Initializing secure port for %s...\n", ${alias});')
    printf("Initializing corpA secure port for %s...\n", "mock_alias");
    
    // @ocq:Replace(with='INIT_HW_PORT(${gateway_addr}, ${baud});')
    INIT_HW_PORT("0x3F8", 115200);
    
    // @ocq:Replace(with='${trailing};')
    ENABLE_INTERRUPTS("0x3F8");
}
// @ocq:Fi(scope="ocq-macro--hw_port_setup")
```

#### Macro Invocation:
```c
int main() {
    // @ocq:Call(macro="hw_port_setup", gateway_addr="0x3F8", baud=115200, alias="corpA_COM1", trailing='ENABLE_INTERRUPTS("0x3F8")')
    printf("Local dev: fallback to simulated serial console.\n");
}
```

---

## Consequences & PR Reviewability

- **Explicit Scoping**: In Pull Requests, nested statements are clearly delimited by named scopes (`init_scope="db_routing"` -> `Fi(scope="db_routing")`). Reviewers can instantly match closures to their declarations.
- **Scope Propagation**: By implicitly propagating scopes to child directives, the code inside the block remains clean and uncluttered.
- **Language-Agnostic Simplicity**: Using `use="profile"` (e.g. `c_function`) on the block allows the compiler to strip local wrappers without parsing target language ASTs.


