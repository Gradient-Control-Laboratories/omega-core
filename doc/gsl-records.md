# Records in GSL — proposal

*omega · prompted by [liaison-graph.md](liaison-graph.md): an interface is a record (class, frame,
size, fit) and GSL has no way to hold one. Status: proposal. The stopgap, option 1, ships as
`joint.insert` and interface values; what building it found is below. Written 2026-10-16,
updated 2026-10-22.*

## Why

The joint library built for the interface sweep fused two ideas because it could not separate
them. `joint.bolted` and `joint.countersunk` take the same seven arguments and differ only in the
screw head; each decides both the **interface** (an axis, a diameter, a fit) and the **fastener**
inserted into it. The design we want separates them:

```gsl
bolt = joint.cylindrical([8.0, 26.75, -4.2], vector.dirZ, 3.0, fastener.fitMedium);
screw, plate, drum = joint.insert(bolt, joint.socketHead, plateBody, drumBody, 11.7);
```

That needs `bolt` to be **one value with named parts**. Today it would be four loose arguments
passed in step everywhere — the same keeping-in-step-by-hand the interface work exists to remove —
and the liaison graph could not be read from data flow, because no single value says "this is
the interface".

## What GSL has today

- **Types** (`common/program/program.proto` `TypeKind`): scalar, LIST (vectors of 2+ scalars),
  MATRIX (equal-length column vectors), FUNCTION (a named function's identity, no signature). No
  tuple, record or map kind, internal or user-facing.
- **Multiple return values** are a `[]*Type` in the checker (`types/infer.go` `inferOperandType`)
  and may only be destructured, never held as one value (`inferSingleType`: "multi-value
  functions … cannot be used directly in expressions; assign them first"). The backends already
  know how to carry several values together: WGSL synthesizes a struct `gsl_struct` with fields
  `v_1..v_n` (`agent/wgsl/internal/wgsl/type.go`); fidget splits them into separate trees
  (`convert/node.go` `MultiTreeNode`); the JS audio backend refuses them.
- **Declaration properties** (`.color`, `.material`, `.box`, `.min`, …) are a closed set
  (`scope/pkg_file.go` `validAssignablePropNames`) and cannot be read back in expressions
  (`link/link.go`: "properties are declaration metadata"), except `.box`. They are not a record.
- **Positional records by convention**: a material is a list of vec4 rows built by
  `material.make` and read through accessor functions (`material/material.gsl`); wood carries
  `finish`, `fungi` and `insects` records the same way. Unnamed rows, and a typeless constant
  fold can lose them (`build/visit_expression.go`).
- **Inference is per call site**: `inferUserCall` re-analyses a callee's body for each call's
  argument types — monomorphisation, with no type annotations anywhere in the language.
- **Backends index only by constant**, carry nothing wider than vec4 through a fold, and the
  fidget backend is fully scalarized. A record must vanish before it reaches them.

## Options

| | Records as … | For | Against |
|---|---|---|---|
| 1 | **positional lists** with accessor functions — the material convention | works today; precedent in the material and wood libraries | unnamed rows, mixed scalars and vectors must be padded into vec4 rows, errors are positional, a fold can flatten them |
| 2 | **open properties** — let `bolt.dia = 3.0;` be declared and read back | reuses existing syntax | reverses a deliberate rule (properties are metadata), a value's identity becomes its declaration, the carry pass is one level deep |
| 3 | **functions** — an interface is a function of the role it is asked for | one value, closes over its frame | a captured function is the case the closure pass bakes rather than lifts, and function types carry no signature |
| 4 | **a record type, scalarized before the backends** | named fields, typed, one value, nothing for the backends to learn | a language change: parser, linker, checker, one lowering pass |

## Recommendation: option 4, scalarized

A record is a value with named fields of any existing type. It exists only up to type analysis;
a lowering pass splits every record into its fields before the backends run, so WGSL, fidget and
JS never see one — the same way they never see a map or a closure today.

### The language

```gsl
// a record literal: named fields, any existing types
bolt = {seat: [8.0, 26.75, -4.2], dir: vector.dirZ, dia: 3.0, fit: fastener.fitMedium};

// field access
r = bolt.dia / 2.0;

// records pass in and come out of functions like any value
cylindrical(seat, dir, dia, fit) {
  return {seat: seat, dir: normalize(dir), dia: dia, fit: fit};
}
clearance(iface) {
  return catalog.clearance(iface.dia, iface.fit);
}
```

- **Structural, no declarations.** GSL has no type annotations; a record type is the set of its
  field names and their types, inferred like everything else. Two records with the same fields
  and types are the same type. Per-call-site inference already handles a function used with
  differently shaped records.
- **Literal syntax** `{name: expr, …}`. A function literal always starts `(params) {`, so a value
  starting with `{` is unambiguous — to be confirmed against the parser.
- **Field access** is `.name`, resolved by type: on a record it is a field; on a vector it stays
  a swizzle. Field names may not be `x`, `y`, `z`, `w` or the accessor names `dp`, `dtime`,
  `second`, `box`, so `bolt.box` and `bolt.dp` keep their meanings.
- **Nesting** is allowed (`{axis: {point: …, dir: …}, dia: …}`), but the grammar takes one
  accessor per factor, so `bolt.axis.point` needs the grammar to chain accessors (`Factor = Value
  ('.' Ident | '[' Expr ']')*`), which `m[0].x` has wanted anyway.
- **Not in the first version**: records in lists, records through `map`/`reduce`, records as the
  top-level rendered value, record update syntax (`{...bolt, dia: 4.0}`), and properties on a
  record (`.color` belongs to fields, not to an interface).

### The lowering

One pass after type analysis and before `lower.Transform` in the assembler pipeline
(`agent/assembler/lib/lib.go`), **scalar replacement of aggregates**:

- a record-typed declaration `bolt = {…}` becomes one declaration per field
  (`bolt__seat`, `bolt__dir`, …);
- a record parameter becomes one parameter per field, and every call passes the fields;
- a function returning a record becomes a multi-return function, and its callers destructure;
- `bolt.dia` becomes a reference to `bolt__dia`.

Everything after the pass is a program the backends already compile. Sliders keep working,
since a field derived from a slider is an ordinary declaration derived from a slider; and the
values are typed throughout, so the constant fold cannot flatten them the way it can a
positional row list.

### What it costs

| Layer | Change |
|---|---|
| parser (`agent/parser/internal/parser2`) | the literal; chained accessors |
| scope / link | `.name` on a record-typed name is a field, not a property read |
| types (`common/program/types`) | a RECORD kind in `program.proto`; inference for literals and field access; record-typed assignments and parameters |
| one new pass | scalar replacement, after `types.Analyze` |
| backends | none |
| docs | the preamble and `doc/gsl_lang_spec_2.md` |

## What it buys the joint library

```gsl
bolt = joint.cylindrical([8.0, 26.75, -4.2], vector.dirZ, 3.0, fastener.fitMedium);
face = joint.planarFace([0.0, 0.0, -7.5], vector.dirZ);
guide = joint.prismatic([0.0, 0.0, -4.4], vector.dirX, vector.dirZ, 44.0, 6.2, 0.5);

screw, plate, drum = joint.insert(bolt, joint.socketHead, plateBody, drumBody, 11.7);
cover = max(coverBlank, joint.below(face));
```

- `joint.bolted` and `joint.countersunk` become one `joint.insert` with the head as a catalog
  constant; a button head or a through-bolt with a nut is a new constant, not a new function.
- The interface is declared once and named; changing the screw type does not touch it.
- The **liaison graph is the data flow**: every part built from `bolt` is on that edge, found by
  following the value, not by matching call text as `gsl joints` does now.
- `verifyJoints` claims attach to `joint.insert`, one set for every head, with the head's own
  seat allowance chosen by the constant.

## Stopgap if the language change waits

Option 1 is buildable now: an interface as a positional list of vec4 rows with accessors
(`joint.diameter(bolt)`), following the material library. It removes the bolted/countersunk
duplication and gives the data flow one value to follow, at the cost of unnamed fields. It would
be replaced by records, not kept beside them.

## What building the stopgap found (2026-10-22)

Option 1 was built before any language change, to test the design: `joint.cylindrical`,
`joint.planarFace` and `joint.prismatic` return vec4 rows read through accessors, and
`joint.insert(iface, head, clamped, threaded, grip)` takes the head as a constant. It builds
geometry, finishes and proofs identical to `joint.bolted` and `joint.countersunk`, and an
Opus sweep with it tied the fused joints (9/10 pass, 7/10 clean first, against 8/8 and 6/8; see
[assembly-joints.md](assembly-joints.md)). So records are not justified by model scores on
this task; the case for them is the library.

Building it found three compiler limits, fixed on their own:

- **Names captured across packages**: a package function's name captured a parameter or
  local of the same name — `joint.nominal()` broke catalog's `nominal` parameters,
  `joint.channel()` joint.slide's `channel` local. Record fields would not collide this way;
  accessor functions did.
- **Finishes carried one call deep**: a library function returning another's part lost
  its `.material` and `.color`.
- **A constant choice was never pruned**: `insert` returns several values and takes
  fields, so the inliner never took it, and both heads shipped in every screw — 26% slower
  proofs. The new pass clones a call on the constants that settle a select.

Two bear on this proposal. Scalar replacement lowers a record-returning function to a
multi-return one, which is exactly the case that pass had to add; and the carry pass runs before
selects are pruned and carries nothing through one, so a record function that picks between
two constructions would still have to set its finishes itself, as `insert` does. Accessor
names also meet the dot-imported builtin namespace: `axis` is builtin's, so the axis accessor
is `joint.direction`.

## Open questions

- Does the parser accept `{` as the start of a value without ambiguity? (The anonymous-function
  form suggests yes.)
- Should field access on a vector-typed record field chain into a swizzle (`bolt.seat.z`)? With
  chained accessors it falls out; without, it needs a temporary.
- Can the scalar-replacement pass run before the closure and inline passes, so lifted and inlined
  functions see plain fields — or must it follow type analysis? The pipeline order in
  `agent/assembler/lib/lib.go` decides this, and it decides how much of the pass is new.
- Error messages: a misspelled field should name the record's fields.
