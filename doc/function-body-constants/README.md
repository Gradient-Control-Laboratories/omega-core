# Constants inside function bodies are not emitted as constants

*Found building `joint.bolted` (doc/assembly-joints.md), 2026-10-01. Not fixed; parked here.*

## The gap

At top level, a declaration whose value fold can work out is evaluated at compile time:
`pitchVal = thread.metricPitch(d)` costs nothing in the shader. Inside a function body the same
call is emitted as the computation itself. [repro.gsl](repro.gsl):

```gsl
f(x) {
  k = catalog.pitch(6.0);
  return length(p) - k - x;
}
a = f(sin(p.x));
b = remap(f(cos(p.y)), p - [3.0, 0.0, 0.0]);
return a, b;
```

emits, in `fn f`, all 36 rows of the catalog table, each `metric[i][1] * catalog_onRow(6.0,
metric[i][0], 0.0001)`, with the whole table passed in as a parameter: a per-pixel lookup of a
number that is 1.0. This holds for a literal argument, so it is not about parameters.

`eval.Fold` does stamp `ConstValues` inside function bodies (`localVals`, fold.go), so the
information exists. The WGSL builder does not use it there. It is unknown whether the fidget
converter has the same gap.

## Why it matters

Any library function that takes a size alongside a field, and is not inlined, pays for every
catalog lookup per pixel. The usual inlining rules keep a function from being inlined when it
reads `p`, its arguments are not all constant, or a multi-return function is called more than
once. `joint.bolted` is the first such function in a common path. The Bolted Joint example
went from 36 KB of WGSL to 78 KB. Five of the largest functions are catalog lookups, about
7.5 KB each.

## What was tried

1. **Relax the inliner's call-count guard** for multi-return calls. No effect, because the joint's
   body arguments are not constant.
2. **Inline mixed constant/field library calls at top level** (each call written out). Bolted
   Joint went to 172 KB and Terrain Bracket and Terrain Planet grew 2.7×. The colour, material and
   height fields repeat each part's expression, so an inlined body is multiplied again.
   Reverted.
3. **Specialize on constant arguments** ([specialize.go.txt](specialize.go.txt), parked; it was
   `common/program/inline/specialize.go`). It clones the callee into the caller's file once per
   distinct tuple of closed constant arguments, written in as expressions so sliders stay live,
   and calls the clone with the rest. It works: the joint's size, grip, fit and seat reach the
   clone's body. But the clone is a function body, so its catalog calls are emitted per pixel
   anyway.

## The fix, in order

1. Emit fold-stamped constants inside function bodies as literals: the WGSL builder first, then
   check the fidget converter. Every function body with a constant call gets cheaper, and many
   goldens will move, smaller.
2. Then restore the specialization pass, so a size passed alongside a field reaches the body as
   a constant.
