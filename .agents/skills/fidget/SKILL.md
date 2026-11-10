---
name: fidget-tree-builder
description: Documentation on properly building Fidget AST trees to avoid interval arithmetic explosions and mesher panics. Use when writing or debugging Fidget ternary logic, conditional statements, built-in function wrappers, or troubleshooting Rust mesher crashes.
---

# Skill: Building Fidget Trees with Interval Arithmetic in Mind

When constructing Fidget AST trees programmatically (such as compiling GSL down to Fidget indices),
it is critical to avoid creating unnecessary interval discontinuities or overlapping condition
branches that cause Fidget's Manifold Dual Contouring (MDC) mesher to panic
(`assertion failed: search[EDGE_SEARCH_SIZE - 1] >= 0.0`).

## 1. Always Ensure `[0, 1]` Bounds via `Max(0, Compare)` for Booleans

Fidget's `Compare(A, B)` evaluates spatial intervals continuously, outputting `[-1, 1]`.
Mapping boolean ternary branches directly over `[-1, 1]` constraints (using polynomial math like
`Abs(cond)`) introduces non-Lipschitz step boundaries and shears the 0-crossing interval bounding
boxes, crashing the MDC solver (`EDGE_SEARCH_SIZE` assertion failures).

Bound conditions cleanly to `[0, 1]` via `Max(0, Compare)`:

### Bad:
```go
// Mapping ternary interpolations directly over [-1, 1] ruins Fidget root bounds.
condTree := fidget.Compare(A, B)
```

### Good:
```go
// Restricts interval bounds to [0, 1], guaranteeing smooth linear transition boundaries.
condTree := fidget.Max(d.zero, fidget.Compare(A, B)) // use cached constant
```

## 2. Eliminate 2-Way Ternary Branch Overlap (Linear Interpolation)

When compiling a 2-way condition over a continuous variable (`cond ? trueResult : falseResult`),
use continuous linear interpolation tied to the `[0, 1]` bounded `condTree`:

### Optimal Formulation:
```go
// Cond restricted to [0, 1]
condTree := fidget.Max(d.zero, fidget.Compare(arg1, arg2)) // use cached constant

// result = falseVal + cond * (trueVal - falseVal)
diff := fidget.Sub(trueVal, falseVal)
result := fidget.Add(falseVal, fidget.Mul(condTree, diff))
```

## 3. Clamp Domains of Built-in Functions

During Manifold Dual Contouring (MDC), Fidget evaluates the bounds of **all branches** of a
conditional selector over a volumetric search space—even if the active logic path normally prevents
a domain violation. Generating `NaN` in the tree propagates upward and causes MDC panics.

### Bad:
```go
// Direct mapping of operations that yield NaN on invalid domains
fidget.Sqrt(arg)
fidget.Acos(arg)
```

### Good:
Clamp numeric domains on mathematical operations that can produce `NaN`:
```go
fidget.Sqrt(fidget.Max(arg, d.zero)) // use cached constant

// Acos/Asin require arguments strictly within [-1, 1]
clamped := fidget.Min(fidget.Max(arg, d.negOne), d.one) // use cached constants
fidget.Acos(clamped)
```

## 4. Safely Cache Delegate Constants

Route common numerical values through cached AST constants on the compiler's `Delegate` object
(`d.zero`, `d.one`, etc.), instantiated **after** `fidget.Clear()` inside `BuildModule`:

### Bad:
```go
// Instantiating before fidget.Clear() retains stale indices across resets!
func NewDelegate() *Delegate {
	return &Delegate{zero: fidget.NewConst(0)}
}
```

### Good:
```go
// Instantiate constants inside BuildModule AFTER fidget.Clear()!
func (d *Delegate) BuildModule(module *model.Module) error {
	fidget.Clear()
	...
	d.zero = fidget.NewConst(0)
}
```

## 5. Golden File Verification

When making changes to Fidget GLSL/CSG conversion pipelines that alter test goldens in
`internal/fidget/internal/convert/testdata/` (`.glsl`), follow the
[golden-verification skill](../golden-verification/SKILL.md) to verify mathematical equivalence.

