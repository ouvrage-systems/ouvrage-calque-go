# ADR 010: Ostream Protocol, Token Ring Stream Architecture, and Immutable Transformations

## Status
Accepted

## Context
In ADR 009, we established the 5-pass compiler lifecycle operating on an in-memory Virtual File System (VFS). However, to enable transparent network piping, concurrent multi-file compilation, and zero-copy streaming across the Ouvrage ecosystem (and external consumers like CLI tools, MCP servers, and WASM runtimes), we need a formal, domain-agnostic stream specification.

Existing serialization formats (such as Helm charts or Jinja2 text templates) treat code generation as unstructured text replacement inside a black box. This results in cryptic errors, zero intermediate visibility, high memory consumption, and vulnerability to injection attacks.

We need a formal, reactive stream protocol that models code generation as an **immutable bus of typed resources ($S \xrightarrow{f} S'$)** while guaranteeing $O(1)$ RAM processing.

## Decision
We decide to adopt the **Ostream Protocol** (`application/x-yaml`) and formalize the **Token Ring Stream Architecture** across all Ouvrage compiler passes:

### 1. The "Envelope + Payloads" Stream Pattern ($2k / 2k+N$)
An Ostream stream is a sequence of multi-document resources separated by `---`. Every logical resource is serialized as a token group:
- **Resource Envelope ($2k$)**: A Kubernetes-style YAML control document (`apiVersion`, `kind`, `metadata`, `spec`). It is uniquely prefixed by a lightweight magic marker comment: `# ostream:envelope(index=N) // Optional comment`.
- **Resource Payloads ($2k+1 \dots 2k+N$)**: Sequential content payload documents following the envelope.

#### Payload Format Agnosticism
Ostream enforces **zero constraints on payload document formats**. A payload can be raw code (C, Go, Bash), XML, JSONL, Markdown, CSV, or binary base64. Because non-YAML/non-text payloads may not support `#` comments, Ostream **never** injects header comments inside payload documents. The magic marker `# ostream:envelope(...)` exists **exclusively above Envelope documents ($2k$)**, which declare the segmentation count $N$.

### 2. Implicit Payload Contracts by `kind` & Dynamic Overrides
Resource payload counts are governed by convention based on their `kind`:
- **`kind: Calque`**: Possesses an implicit contract of **exactly 3 mono-file payload slots** ($2k+1$=Extracted, $2k+2$=Parsed, $2k+3$=Compiled).
- **`kind: SymbolTable`**: Possesses an implicit contract of **1 multi-file payload slot** (DAG & Symbol Table).
- **Dynamic Override**: Generic or custom stream resources specify `spec.payloadCount` to dynamically declare $N$ payload documents.

### 3. Dynamic Delimiter Escaping (`spec.ostream.payload_separator`) & Bundle Stack
To handle payloads containing internal `---` string boundaries (such as Markdown files with YAML frontmatter or raw multi-document Kubernetes manifests), Ostream provides a dynamic delimiter escape mechanism and hierarchical scoping:

#### 3.1 Custom Delimiter Override
An envelope can declare a custom payload separator in its stream specification:
```yaml
spec:
  ostream:
    payload_separator: "===---==="  # Overrides default '---' for payloads!
```
The scanner uses `===---===` to delimit subsequent payloads, eliminating boundary collisions.

#### 3.2 Hierarchical Scoping Stack (`kind: Bundle`)
When an Ostream carries nested asset groups, a parent **`kind: Bundle`** resource pushes stream configuration rules (custom payload separators, compression, encoding) onto a **Scanner Execution Stack**:
- All subsequent child assets (`Asset1`, `Asset2`) inherit the parent Bundle's stream rules.
- When the scanner encounters a new `kind: Bundle` or closing marker, it pops the execution stack to restore the previous stream rules.

### 4. Immutable Functional Stream Transformations $S \xrightarrow{f} S'$
An Ostream $S$ is an **immutable snapshot** of the pipeline state. Compiler passes operate as pure mathematical functions $f(S_i) \rightarrow S_{i+1}$ without mutating streams in-place:

$$\text{Source Files} \xrightarrow{\text{Extract}} S_1 \xrightarrow{\text{Parse}} S_2 \xrightarrow{\text{Compile}} S_3 \xrightarrow{\text{Link}} S_4 \xrightarrow{\text{Render}} S_{\text{final}}$$

### 4. Pre-2000s Network Token Ring Parallelism & $O(1)$ RAM
The Ostream engine models its execution after classic **Token Ring (IEEE 802.5 / FDDI)** network frame processing:
- The Stream Scanner reads control envelopes ($2k$) and payload documents in $O(1)$ RAM memory.
- As soon as a full resource token is scanned, it is dispatched via Go channels to a worker pool of goroutines for concurrent compilation (Phases 1-3).
- The raw input buffer is immediately dropped from RAM before scanning the next resource.

### 5. Stream Payload Clearing (`--- # cleared`) vs. `--debug` Mode
- **Production / Standard Mode**: Obsolete intermediate phase payloads are cleared and emitted as `--- # cleared`. Only the active phase payload carries data, guaranteeing low bandwidth and $O(1)$ memory usage.
- **`--debug` Mode**: All phase payloads are preserved in the single Ostream file, providing a complete Intermediate Representation (IR) audit trail for developers and AI agents.

### 6. Phase 4 Linker: Standalone `kind: SymbolTable` Resource
Phase 3 AST (`kind: Calque`) remains 100% pure and uncorrupted—it does not inline or mutate imported AST fragments into itself. Phase 4 operates across multiple project files and emits a dedicated **`kind: SymbolTable`** Ostream resource containing the Directed Acyclic Graph (DAG) and namespaced symbol bindings.

### 7. Universal Protocol Decoupling & WASM Portability
The Ostream protocol is domain-agnostic and can transport HTTP requests, S3 bucket events, K8s manifests, or telemetry logs. Compiling the engine to WebAssembly (`WASM`) provides native bindings for Python (`wasmtime-py`), Node.js (`@ouvrage/ocalque-wasm`), Rust, C++, and browser IDEs.

## 8. Visual Architecture & Stream Schematics

### 8.1 The Ostream Envelope + Payloads Protocol ($2k / 2k+N$)

```text
================================================================================
                    OSTREAM STREAM MULTI-DOCUMENT FORMAT
================================================================================

---
# ostream:envelope(index=1) // Primary Calque Asset
apiVersion: ocalque.ouvrage.io/v1
kind: Calque
metadata: { name: "auth-service", path: "services/auth-service.yaml" }
spec: { activePhase: 3 } # (kind: Calque implicitly defines payloadCount=3)
---
# cleared
---
# cleared
---
astNodes:
  - type: ImportNode
    id: "9b67cdd6"
    properties: { source: "templates/pod-spec.yaml", as: "pod_blueprint" }
---
# ostream:envelope(index=2) // Linker Symbol Table Asset
apiVersion: ocalque.ouvrage.io/v1
kind: SymbolTable
metadata: { name: "auth-service-symbols", source: { id: "9b67cdd6" } }
spec: { activePhase: 4 }
---
symbols:
  imports: { pod_blueprint: { targetID: "dea497a0", path: "pod-spec.yaml" }}
  dag: ["templates/pod-spec.yaml", "services/auth-service.yaml"]
```

### 8.2 Token Ring Reactive Processing ($O(1)$ RAM Dispatch)

```text
================================================================================
            TOKEN RING REACTIVE STREAM DISPATCH (O(1) MEMORY)
================================================================================

 [ Incoming Ostream S1 ] (stdin / network pipe)
         │
         ▼
 ┌────────────────────────┐
 │  Ostream Scanner (Rx)  │ <--- Scans Envelope 2k + N Payloads in O(1) RAM
 └────────────────────────┘
         │
         │ (Dispatches Token Group via Go Channels)
    ┌────┼────┐
    ▼    ▼    ▼
 ┌────┐┌────┐┌────┐
 │ G1 ││ G2 ││ G3 │   <--- Goroutine Worker Pool (Compiles Phases 1..3 in Parallel)
 └────┘└────┘└────┘
    │    │    │ (Buffers dropped from RAM immediately upon processing)
    └────┼────┘
         ▼
 ┌────────────────────────┐
 │  Linker & Stream Sink  │ <--- Phase 4 Linker Barrier & SymbolTable Emission
 └────────────────────────┘
         │
         ▼
 [ Outgoing Ostream S4 ] (stdout / VFS memory)
```

## 9. Consequences
- **Predictable & Auditable**: Every generated line is cryptographically traceable via SHA-256 (`id`) and 2D line geometry (`geo`), satisfying DGA, NIS2, SecNumCloud, and ISO 26262 requirements.
- **High Performance**: Goroutine worker pools compile multi-file repositories concurrently in $O(1)$ RAM.
- **Zero Black-Box Failures**: Intermediate checkpoints can be inspected or reloaded without re-running earlier parsing passes.
- **Protocol Portability**: One single Go codebase runs everywhere (CLI, serverless, WASM, MCP servers).

