---
name: clean-go-test-guards
description: >-
  Audits, identifies, and eliminates test-induced defensive checks, nil guards, and unused parameters
  from Go production code, restoring clean domain invariants while preserving defensive checks at
  system boundaries and updating unit test fixtures appropriately.
---

# Clean Go Test-Induced Guards & Invariant Enforcement

This skill defines the workflow for identifying and removing test-induced defensive checks,
unnecessary nil guards, and unused function parameters from Go production code.

Go production code should enforce and rely on well-defined domain invariants.
Defensive guards (such as `if node == nil`, `if ret != nil`, or empty-slice checks)
frequently enter internal production code not because production pipelines can ever produce
nil values, but because sparse unit tests directly instantiate structs without required fields.
Similarly, unused function parameters linger when tests or refactors change requirements.
These test-induced accommodations pollute core logic, mask invalid states, and weaken system
invariants.

## Core Principles

### 1. Preserve Defensive Checks at System Boundaries
System boundaries MUST perform strict defensive validation and explicit error handling:
- **External Inputs**: Network endpoints (HTTP, gRPC), CLI flag parsing, and user inputs.
- **Parsers & Deserializers**: Token parsers, JSON/Protobuf unmarshalers, and file loaders.
- **Public Library APIs**: Public exported functions intended for third-party consumers.

Defensive checks at these entry points are legitimate domain guards, protecting the system
against corrupt or malformed inputs.

### 2. Eliminate Test-Induced Guards in Internal Pipelines
Once data passes validation at the system boundary and enters internal core pipelines
(AST traversals, compiler lowering passes, internal business logic):
- Internal passes MUST trust the invariants established at the boundaries.
- Defensive nil skips (`if x == nil { continue }` or `if x != nil { ... }`) that only exist
  to appease incomplete or partial unit test fixtures must be eliminated.
- Test fixtures must be updated to construct valid domain instances rather than polluting
  production code with guards for impossible internal states.

### 3. Eliminate Unused Parameters with Caller Propagation
Function and method signatures should only accept parameters required for execution:
- Identify parameters that are never referenced in the function body.
- When an unused parameter is removed from a function signature:
  1. Update the function signature and implementation.
  2. Inspect and update ALL call sites across the codebase to remove the argument.
  3. Recursively check if arguments passed at call sites (local variables, expressions)
     are now unused in the caller; if so, remove or inline them.
- **Exception**: Parameters required to satisfy a defined interface or callback signature
  mandated by an external dependency or framework should be named `_` to signal intent.

### 4. Eliminate Unused or Test-Only Private Methods
Internal functions and methods must directly serve active production pathways:
- Audit unexported (private) functions and methods to verify they have live production callers.
- If an unexported function or method is never called in production code:
  1. If it is only called from `_test.go` files, evaluate whether it is dead code or a test helper.
  2. If it is dead or speculative production code, remove it permanently.
  3. If it is a legitimate test utility, move it into `_test.go` or a dedicated test fixture package
     rather than leaving test-only scaffolding in production files.

## Workflow

### 1. Identify Candidate Defensive Guards, Unused Parameters & Test-Only Scaffolding

Search the target Go codebase for defensive checks, signature clutter, and dead scaffolding:
- **Contrasting Invariants**: Compare sibling fields or adjacent methods. If `file.Returns`
  does not guard against nil, `file.ColorReturns` or `file.MaterialReturns` should not either.
- **Internal Nil Skips**: Look for defensive skips (`if x == nil { continue }`) in AST visitors,
  compiler lowering passes, serialization pipelines, or domain models.
- **Receiver Nil Checks**: Look for redundant `if t != nil && t.Method()` when pointer receiver
  methods already handle `t == nil` safely.
- **Unused Parameters**: Check function declarations where a parameter is unused in the body
  and not required by an interface contract.
- **Test-Only Private Methods**: Search for unexported functions or methods whose only call sites
  are in `_test.go` files or that have zero callers anywhere.

### 2. Determine Domain & Boundary Invariants

Trace the origin of the data structure through the pipeline:
- Does the parser, builder, deserializer, or any upstream code guarantee that elements are non-nil?
- If the contract guarantees that valid instances never contain nil elements,
  the internal check is suspect and likely test-induced.
- Is the function at a public system boundary? If yes, keep the check. If internal, remove it.

### 3. Hypothesis Testing via Guard Removal

Temporarily remove the defensive check or unused parameter from production code.
Run the package test suite:

```bash
go test -v ./path/to/pkg/...
```

Analyze the outcome:
1. **All tests pass**:
   The check was redundant defensive clutter.
   Remove it permanently and add an invariant comment in its place documenting what is expected.
2. **Tests fail**:
   Inspect the failure stack trace and the failing test code:
   - **Test-Induced Failure**: The failing test directly constructs a sparse or partial struct
     (e.g., passing nil operands or uninitialized fields). Proceed to Step 4.
   - **Legitimate Domain Edge Case**: The pipeline or public API boundary intentionally supports
     nil or optional elements as a valid state. Retain the check and document the invariant.

### 4. Remediate Unit Test Fixtures

When a test failure is caused by sparse or malformed test fixtures:
- Update the test to use standard domain constructors or helper builders.
- Populate required fields with valid, minimal domain instances rather than nil.
- Provide reusable domain test helpers (e.g., `NewValidNode()`) to make writing invariant-compliant
  tests easy and maintainable.
- If testing negative error paths, test through the proper boundary entry points that validate
  inputs.

### 5. Propagate Parameter Removals to Callers

When an unused parameter is removed:
1. Update all callers across production code and test files.
2. Check if removing the argument leaves caller variables unused; recursively clean them up.
3. Verify that interface implementations remain aligned.

### 6. Document Domain Invariants

Whenever a defensive check is removed,
add an explicit comment in its place documenting what is expected
(such as `// Invariant: <condition guaranteed by upstream code>`).
This preserves domain knowledge and clarifies assumed constraints without runtime overhead.

### 7. Final Codebase Verification

After removing defensive checks, documenting invariants, removing unused parameters, and updating test fixtures:
1. Run the full package and dependent test suites:
   ```bash
   go test -count=1 ./...
   ```
2. Review the diff to confirm that production code adheres strictly to domain invariants:
   ```bash
   git diff
   ```
