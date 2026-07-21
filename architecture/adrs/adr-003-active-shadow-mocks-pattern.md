# ADR 003: Active Shadow Mocks Pattern

## Status
Accepted

## Context
When templating files for different environments (e.g., local dev vs. production), we often need to include or exclude specific block configurations, files, dependencies, or names. 

If we comment out production-only blocks in the source file using standard block comments (e.g., wrapping them in `<!-- ... -->` in XML/HTML or `/* ... */` in Go/C) to prevent them from executing in local dev:
1. **Loss of IDE/LSP Validation**: The IDE treats the commented-out block as dead text. It stops validating its syntax, types, or attributes, leading to silent bugs that only appear when the block is uncommented in production.
2. **Visual Clutter**: The template code becomes cluttered with mixed commenting levels, making it hard for developers to read and edit.

We need a design pattern that keeps all environment-specific components active and visible to the IDE during development, while ensuring they are safely substituted or removed before deployment.

## Decision
We decide to adopt the **Active Shadow Mocks** pattern as the standard design practice for `ocalque` templates:
1. **No Comment-Outs for Active Elements**: Production-only elements must not be commented out in the source file. They must remain as active, syntactically valid code or configuration structures.
2. **Shadow Elements (Mocks)**: These elements are configured to point to local, harmless dummy targets (known as "shadows" or "mocks"). For example:
   * In a Go file, a production node is registered inside a dummy function `_ouvrage(func() { ... })` that does nothing in local dev.
   * In a Telegraf TOML file, production outputs are placed in a dummy namespace `[ouvrage.shadow.outputs.elasticsearch]`.
   * In an XML/WiX installer file, the production DLL component references a local empty `placeholder.dll` file.
3. **AOT Substitution**: During the build phase, `ocalque` uses geographic directives (e.g., `unwrap`, `replace`) to compile these shadow structures into their actual production configurations (e.g., replacing the dummy namespace with the real one, and the placeholder file with the actual secure production DLL).

## Consequences
* **Continuous Linting and Validation**: The IDE can lint, auto-format, and autocomplete all elements (dev and prod) because they exist as active structures in the source file.
* **Hermetic Local Execution**: The source template can be executed or compiled locally without errors because all production elements reference valid local placeholders.
* **Explicit AOT Mutations**: The AOT compiler performs surgical, line-restricted mutations to swap shadows with actual production values, preventing accidental or unintended replacements.
