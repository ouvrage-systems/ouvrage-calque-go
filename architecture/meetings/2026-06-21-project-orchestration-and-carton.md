# Meeting Notes: Project Orchestration & "The Carton" Concept

This document logs the design and naming discussions regarding project-level compilation, discovery, and template collection management.

---

> [!NOTE]
> The participants of this meeting (Gérard, Hans, Marc, Alex) are **Interactive Design Personas (Virtual Engineering Committee)** simulated during the pair-programming session. They are not autonomous AI agents or bots. For details on their profiles, see [personas.md](file:///home/gpineda/Documents/ouvrage/ouvrage-calque-go/architecture/meetings/personas.md).

## 1. General Info & Participants

* **Date**: June 21, 2026
* **Participants**:
  * **Guillaume** (Creator of the Ouvrage guild)
  * **Gérard** ("Bash-Guru": Traditional SRE/Ops SysAdmin)
  * **Hans** ("Windows-Ops": Siemens cRSP SRE)
  * **Marc** ("C-Bas-Niveau": C/Zig systems developer)
  * **Alex** ("CNCF-boy": Kubernetes/GitOps architect)

---

## 2. Evolution of the Project-Level Design

We debated how to handle the compilation of an entire repository (containing multiple template files) instead of processing single files one-by-one.

### A. Rejection of "Magic" Global Replacements
* **Initial Idea**: Define global replacements or defaults in a root configuration file (e.g., `globals.defaults.replace_patterns`).
* **Critique**: This was rejected as "too magic" and "obscure." Hiding transformation rules in a central external configuration violates the core philosophy of `ocalque` (which dictates that all transformations must be explicit, local, and visible inside the template files). 
* **Decision**: All template directives remain local to the source files. The compiler must not use global, implicit text replacement tables.

### B. Postponing Multi-File Orchestration (MVP Focus)
* **Decision**: We decide to postpone the implementation of repository-wide orchestration. To ensure a fast, robust, and minimalist initial version, we focus exclusively on the **MVP (file-by-file compilation)**. 
* **Next Steps**: We will build the core parser and transformer first. We will address repository-level orchestration only when confronted with actual scale problems during SRE deployments.

### C. The Concept of "The Carton" (Le Carton)
To stay aligned with the Ouvrage guild's design ontology (German for raw mechanics, French for artistic/UX/layout concepts), we defined the vocabulary for a collection of templates:
* **The Metaphor**: In architectural drafting workshops, architects store and carry their tracing papers and design sheets inside a **"Carton à dessins"** (a drawing portfolio case).
* **The Terminology**: A repository-level configuration or portfolio of templates will be called **"The Carton"** (Le Carton). 
* **Future Implementation**: When we eventually build project-level orchestration, the configuration file at the root of the repository will be named `carton.yaml`. It will govern:
  * **Discovery**: Scanning the repository for templates containing `@ocalque` annotations.
  * **Filters**: Filtering templates based on target environments or tags.
  * **Context**: Specifying which environment variables to load for the compilation.
