# Specification: `gsl eval`: Two Faults a Side-by-Side Run Found

## 1. Why

Four ordinary parts (an enclosure, a pipe tee, a ribbed bracket, a hand knob) were each modeled by an agent in
GSL and in a B-rep kernel, on release 0.1.2. Two of `gsl eval`'s answers were wrong on those parts, a
field's gradient did not compile on a third, and `volume` took a minute and a half on a blended part. This track
fixes all four. It follows `eval_answers_20261115`.

## 2. Faults and Decisions

| # | Fault                                                                     | Evidence                                                                                                                                | Decision                                                                                                           |
| : | :------------------------------------------------------------------------ | :-------------------------------------------------------------------------------------------------------------------------------------- | :----------------------------------------------------------------------------------------------------------------- |
| 1 | A slider read through a closure's `remap` is a name nothing defines       | The bracket and the tee, which join bodies with `unionRoundExact`: `shader` in `glsl-es300` refuses `gsl_param_9_23`                    | The helper that re-derives the slider's variable takes the slider as a parameter, and its callers pass it          |
| 2 | `bounds` says `proven` with a box far off the part                        | The knob, whose outline folds its angle with `atan`: `"proven": true` with a `cellSize` of 9.0 mm in x and 5.3 mm in y, on a 50 mm part | A box is proven only within 1/64 of the part's span of points known inside, however its search ended               |
| 3 | A field's gradient does not compile where a `reduce` step reads its index | The bracket agent's `.dp` of an extruded `profileXY`: `undefined identifier: index` at `2d.gsl` 212:27                                  | The gradient's `reduce` passes the step its index, and an index reading a fold-local name is evaluated, not copied |
| 4 | `volume` takes about 100 seconds on a blended part                        | The tee: 100 s at the default 5%, `converged: false` by budget; agents called it two or three times a part                              | A default time limit of 10 s and a `timeLimit` request field; the answer says `time` where it ended the search     |

## 3. The Slider in a Closure

`unionRoundExact(a, b)` calls each closure at a moved point, with `remap`, inside its own function. The remap's
helper re-derives every primary-scope field variable the closure reads, a slider's among them: `let r =
gsl_param_2_18;`. At the top level the helper takes the slider as a parameter, because the primary scope's
parameters are handed to it. Inside a function's body they are not, and the reference the assignment recorded
when the slider was declared is on the primary scope, which the helper's scope never sees.

The fix records the slider's parameter as a reference of the function being generated at the moment that
function emits a variable defined by it. The helper then takes it, the function calling the helper takes it,
and `sdShape` passes it. A helper that already took it, the top-level case, is unchanged: no golden moves.

`gsl compile -t wgsl` wrote the same undefined name: nothing validates that target's module. The `glsl-es300`
target is a transpile, which does.

Not examined: a UI control's literal (`uiParamVars`) read the same way.

## 4. The Loose Box

`Enclose` gave up a face that the budget ended more than 1/64 of the span from a point known inside. A face
whose search ended on a leaf it could not drop, with budget to spare, was not held to that. An `atan` of a box
across the angle's seam spans the whole turn, so no leaf touching the seam can be dropped, and the search ends
on them at once: the box stood 9 mm off the knob and was answered as proven.

The rule now applies to every face. Where it fails with budget to spare the reason is the new `loose`; where
the budget ran out it is `budget`, as before. The box answered is then the grid's.

## 5. The Gradient of a Reduce

A gradient (`.dp`) is built by evaluating the field again in a dual-number domain. That evaluation stepped a
`reduce` with `(accumulator, value)` only, so a step that names its third parameter, `index`, read a name nothing
bound; lowering passes it. `profileXY`'s distance pools read `index` both as a value and inside a list index,
`points[mod(index - 1 + count, count)]`, and an index is copied into the output, where `index` and the inlined
local `count` do not exist. The step now gets its index, and an index that reads a name local to the evaluation is
evaluated in the domain and its plain value used. An index that reads nothing of the kind is resolved as before,
so no golden moves.

## 6. A Time Limit on Volume

`volume` refines its cells until the bracket meets the tolerance or 64 million interval evaluations are spent. A
round blend's intervals are loose, so on such a part it spends them all: about 100 seconds. It now also stops at a
time limit, 10 seconds unless the request names another, checked between batches of cells. Proven-inside cells
only add to the lower bound and unresolved cells stay in the upper, so the bracket it answers with holds wherever it
stops. The evaluation budget stays as a backstop, and its reason, `budget`, wins where both are spent.

## 7. Out of Scope

- Tighter intervals for `atan` and the remainder across the seam, which would let such a part be proven.
- `unionRoundExact`'s surface on a seam that is not square, where the same run measured it 1.66 mm full.
