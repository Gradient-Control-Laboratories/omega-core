
---
name: gsl-author
description: Expert in GSL (GCL Shape Language) creating Signed Distance Functions (SDFs) and geometric shapes.
---

# Skill: GSL Code Generator

## **Description**
The GSL (GCL Shape Language) is a C-like, domain-specific language
designed for defining Signed Distance Functions (SDFs) and geometric shapes.
It is a functional, Single Static Assignment (SSA) language that operates primarily on floating-point numbers and vectors.

---

## **Core Concepts**

* **Implicit Point (p):** The variable p is implicitly defined and represents the current point in space (a 3D vector vec3<f32>) being evaluated.
It is implied and should never appear in the list of a function's parameters.
Use remap (see below) to translate, rotate, and scale shapes by transforming p.

* **Time (time):** The variable time is implicitly defined and represents elapsed time in seconds (f32).
It is implied and should never appear in the list of a function's parameters.
Use it to animate shapes by varying distances, positions, or other values over time.

* **Units (mm):** Every length is in millimeters — coordinates, radii, offsets,
and the distance a field returns. `sphere(5)` is a 5 mm ball, and the viewer's
scene size, grid spacing, and export resolution are millimeters too.
When a drawing gives inches, write `2.0 * inch` rather than `50.8`, so the
drawing's number stays visible in the source (`inch`, `foot`, and `meter` convert
into millimeters; there is no `mm`, since `1.0 * mm` says nothing that `1.0` does not).

* **Single Static Assignment (SSA):**
The entire language is built on the concept of Single Static Assignment (SSA)
where variables are assigned exactly once.
Re-assigning a variable is not allowed.
For example,

```gsl
// This is allowed
x = max(1, 2);
y = max(x, 3);

// This is not allowed
x = max(1, 2);
x = max(x, 3);
```

* **Return Values:**
Functions can return multiple values.
The top-level function's return values are the parts of an assembly: each value is its own part, and they are not combined.
Merging them into one solid is only an option when exporting a mesh.
To get one shape, combine them yourself, for example "return unionSharp(s1, s2);".

* **Audio Output:**
A top-level program can optionally output audio using
`audio(signal)` for mono (the signal is sent to both channels),
or `audio(leftSignal, rightSignal)` for stereo.
This audio expression must appear as the last element in the return statement.
Audio signals should be in the range [-1, 1].

* **Data Types:**
  * **Float (f32):** Standard floating-point numbers used for all calculations,
but integers are allowed in the source code and are converted to floats.
  * **Vector:** Ordered collections of 2 or more scalars.
  * **List:** Ordered collections of 2 or more vectors of the same length.
Lists are indexed with constant integers and iterated with map and reduce;
swizzling and arithmetic do not apply to lists of vectors.

  * **No Booleans:** Logic uses 1.0 for true and 0.0 for false.
Numbers are converted to booleans using the following rules:
    * less than or equal to 0.0 is false
    * greater than 0.0 is true

* **First-Class Functions:** Functions can be named or anonymous, assigned to variables, and passed as arguments.

* **Formatting:** Indent nested blocks with **2 spaces** (never tabs, never 4).
The modeling library and the canonical examples all follow this convention;
mixing indent widths is the single biggest source of diff noise when humans
review AI-generated GSL.

---

## **Syntax & Grammar**

### **Declarations**

* **Scene Size and Grid (`@scene`):**
A directive at the very top of the first file in a package, before any imports
or declarations, that sets the viewport's viewing box and grid spacing — both
in millimeters, same as every other length in the language.
  ```gsl
  @scene(20)      // 20 mm viewport, grid defaults to size / 10
  @scene(20, 2)   // 20 mm viewport, 2 mm grid
  ```
No semicolon follows the directive.

* **Variable Assignment:** `name = expression;`
Supports multiple assignment: `x, y = someFunc();`
Top-level variables are evaluated in order: declare a name before any line that uses it. Functions may be called before their definition.

* **Variable Properties (sliders):**
A top-level variable assigned a single numeric literal can be annotated
with `name.min` and `name.max` properties (constant expressions).
The two properties together turn that variable into a live slider in
the editor's Parameters panel: dragging the slider rewrites the literal
in place and updates the running shader / audio without recompiling.
The literal's decimal places set the slider's rounding step
(`radius = 5;` snaps to integers, `radius = 0.50;` snaps to 0.01).
An optional `name.period` property (seconds, must be > 0) makes
the host drive the slider as a sinusoid through `[min, max]` while
the animation clock runs; each animated row gets a play/pause toggle so
the user can grab the slider manually at any time. For example:
  ```gsl
  radius = 5.0;
  radius.min = 0.5;
  radius.max = 10.0;
  radius.period = 4.0;  // animate radius over 4 seconds
  return remap(sphere(radius), p);
  ```

* **Control annotations:**
A top-level variable assigned a literal can instead carry one annotation
naming the control it wants: `@range(min, max)` is the slider
above in one line; `@point` on an `[x, y, z]` marks a
point in the scene, `@direction` on an `[x, y, z]` a
direction, drawn from the origin or from `@direction(anchor)`'s
point (a `@point` variable or a constant),
`@color` on an sRGB `[r, g, b]` a colour, and
`@transform` on a 4x4 matrix a frame. The annotation changes nothing
about the value. The same annotations on a function's parameters declare what
kind of argument each takes, so a tool writing a call to the function can give
each argument its control:
  ```gsl
  powderCoat(@color coatColor, @range(0, 100) glossUnits, @range(0, 1) peel) { ... }
  turned(@point axisOrigin, @direction(axisOrigin) axisDir, @range(0.05, 0.5) feed) { ... }
  ```
On a parameter, `@direction`'s argument names the sibling parameter
the arrow is drawn from. `@frame` is a parameter kind alone: the
body's own frame, `[body.u, body.v]`, which a tool writing the call
reads off the body rather than asks for. `@surface` marks a
function that takes a surface first and returns the modified one, the
overload of a surface operator that takes only those slots and bakes the
operator's other arguments: what the Bodies panel wraps a surface with.

* **Shape Color (`name.color`):**
A declaration property that gives a shape an RGB color:
`name.color = [r, g, b];` — a length-3 vector with components in
`[0, 1]`. It is a field, not a constant, so it may vary with
`p`; the `color` package supplies named constants and ramps.
Color is carried by the distance rather than painted on afterward: where shapes
combine, each contributes in proportion to how much it drives the combined
field, so a sharp union hard-selects, a smooth union blends, and a chamfer
facets — with no per-operator rules. A colored intermediate acts as ONE object,
so an uncolored shape unioned into a colored group inherits the group's color.
Colors transport through `remap` with the shape they belong to.
  ```gsl
  ball = sphere(5);
  ball.color = color.viridis(p.z * 0.1);
  return ball;
  ```

* **Shape Material (`name.material`):**
A declaration property that gives a shape a surface finish, independent of its
color: `name.material = material.chrome;` — a record from the
`material` package. It holds roughness, metallic and index of refraction
(`material.make(rough, metal, ior)`); an emitted colour and strength
(`material.glowing(rgb, strength)`); transmission, with the absorption distance
in mm over which the shape's color comes through
(`material.translucent(m, transmission, absorptionDistance)`); a displacement
height in mm, drawn by the renderer only (`material.withDisplacement(m, height)`);
and a lay, grooves too fine to see that stretch the highlight along a direction,
as on brushed or turned metal (`material.withLay(m, direction, anisotropy)`).
Each `with…` modifier returns a copy of `m` with one part replaced, and
`material.blend(a, b, t)` mixes two records, `t` a field, so one field can drive
every part at once. Build and read records with these functions rather than
indexing their rows.
A shape may set color, material, both, or neither; each is a field of
`p`, blended and inherited across combinations exactly like color,
and a shape that sets neither renders with `material.standard`.
Presets carry no color, so a metal comes from the `metal` package
as a zero-argument function returning a PAIR: its finish and its color, in
that order. Transmissive materials come from the `glass` package the
same way (`glass.clear()`, `glass.frosted()`,
`glass.resin()`, …), as do the measured materials in
`physical`; for glass the shape's color is the tint that comes
through, and parts inside the glass are seen through it.
  ```gsl
  boltMat, boltColor = metal.brass();
  bolt = cylinderZ(4, 20);
  bolt.material = boltMat;
  bolt.color = boltColor;
  return bolt;
  ```
  ```gsl
  bar = max(cylinderZ(vector.zero, 4), abs(p.z) - 20);
  barMat, barColor = metal.steel();
  bar.material = material.withLay(barMat, bar.u.dp, 0.8);   // grooves run round the bar
  bar.color = barColor;
  return bar;
  ```

* **Surface frame (`name.u`, `name.v`):** two coordinates of a point on a shape's own surface. The primitives write them, so you rarely do: a cylinder's `u` is the angle round its axis in radians and its `v` the distance along it in mm, a sphere's are azimuth and polar angle, a box's are the in-plane distances across each face. Read one as `bar.u` to hand a material the frame of the shape it is on, rather than a direction you keep in step with the geometry by hand; `bar.u.dp` is the direction of increasing angle, round the axis. A shape with no frame reads 0. A sharp union hard-selects the nearer shape's frame, so it flips across an edge; only a blended boolean averages them. An angle jumps by 2 pi where it wraps but its gradient does not, so take `.dp` of it rather than indexing by its value. Declare your own with `name.u = expr;`.

* **Shape Box (`name.box`):**
A declaration property that says where a shape is: `name.box = [lo, hi];`,
two vec3 corners. The declaration is cut to that box, grown by `boxPad`, so
a correct box changes nothing and a tight one crops. The viewer cuts the part
open, so its inside shows through where the box crosses it; export caps it,
so the exported part stays closed. Declare one where the box cannot be
inferred — on a remapped
shape, or inside a function in terms of its parameters, as a library part
would — and read it back with `name.box` (see Bounding box below).
  ```gsl
  wave = remap(sphere(5), [p.x, p.y + sin(p.x), p.z]);
  wave.box = [[-5, -6, -5], [5, 6, 5]];
  return wave;
  ```

* **Stand-in (`name.simple`):**
What a value draws as when its part is small on screen; the viewer blends
to it across the octave below the Detail slider's rung for it. Every value's
is itself until one is declared, as with `.box`. On a part,
`name.simple = proxy;` is a simpler shape drawn in its place below the
Simple rung, cropped to the part's box, with the proxy's own colour and
material, or the part's where the proxy has none. Only a part can have a
shape stand-in: a named value of the document that only a min or a return
reads, and whose box is bounded. On a colour or material,
`name.simple = flat;` is the flat look the value fades to below the
Textures rung, and it is not evaluated past it; library textures declare
theirs, so a checkerboard reads as its even mix. Export, meshing and every
measurement use the value itself. Keep a stand-in cheap; it is there to cost
less than what it stands in for.
  ```gsl
  knob = boxCentered([0, 0, 0], [20, 20, 40]) - 2;
  knob.color = [0.8, 0.5 + 0.3 * sin(p.z), 0.2];
  knobProxy = boxCentered([0, 0, 0], [24, 24, 44]);
  knobProxy.color = [0.8, 0.5, 0.2];
  knob.simple = knobProxy;
  return knob;
  ```

* **Properties carried out of a function:** a function that builds a shape, sets `.color`, `.material` or `.simple` on it and returns it hands those properties to the shape the caller gets, so the caller need not set them again. A property the caller sets on its own name wins.
  ```gsl
  finished(r) {
    s = sphere(vector.zero, r);
    s.material = material.chrome;
    s.color = [0.9, 0.2, 0.2];
    return s;
  }
  ball = finished(5);             // chrome and red
  dull = finished(5);
  dull.color = [0.2, 0.2, 0.9];   // the caller's color wins
  ```
  A function that returns several shapes keeps each one's own (`a, b = pair();`). The function must return a local variable, or a call to another function that carries; returning an expression built from one, such as `remap(s, q)`, carries nothing out.

* **Function Definition:**
  ```gsl
  // Named Function
  myFunc(a, b) {
    return a + b;
  }

  // Anonymous Function
  myVar = (x) { return x * x; };
  -- or --
  result = oneFunc((x) { return x * x; }, 2, 43);
  ```

* **Return Statement:** `return expr1, expr2;` (Functions can return multiple values).
* The return statement can include an optional audio expression as the last element.

### **Expressions**

* **Vector Construction:** `[val1, val2, val3]`

* **List Construction:** A list of vectors nests the brackets, e.g. a
matrix written column-by-column as `[[a, b], [c, d]]`. Elements are
separated by commas.

* **Swizzling:** One way to access vector components is using `.x, .y, .z, .w`.
Supports any combination of components (e.g., `.xy, .zyx, .xxxx`) up to 4 components.
Example: `v.xy` returns a vec2; `v.z` returns a scalar.

* **Vector Indexing:** Another way to access vector components is using `[index]`.
Example: `v[0]` returns the first component; `v[1]` returns the second component.

* **Derivatives:** `.dtime` and `.dp` compute derivatives of a field by automatic differentiation.
  * `f.dtime` — the rate of change of `f` with respect to `time` (same shape as `f`). Composes for higher orders: `f.dtime.dtime` is the second time derivative (e.g. velocity, then acceleration).
  * `f.dp` — the spatial gradient of a scalar field `f` with respect to the point `p`, returning a vec3. For a distance field this points along the surface normal; use `normalize(f.dp)` for a unit normal.
  * Works in top-level expressions and inside functions (a function using `.dp` / `.dtime` on a parameter is differentiated at each call site). Use parentheses to combine with other accessors, e.g. `(f.dp).x`.
  Example: `vel = shape.dtime;` `n = normalize(shape.dp);`

* **Second-closest distance:** `f.second` is the distance to the
SECOND-nearest primitive surface of `f`, over the surfaces reachable
through `min` / `max` (and affine offsets). Where only one
surface is in play there is no second, and `.second` returns a large
finite sentinel that loses every comparison. It composes with the other
accessors, so `f.second.dp` is its gradient. Because `f` and
`f.second` vanish together exactly where two surfaces meet on the
boundary, `length([f, f.second])` is a distance-to-crease field — the
usual way to find and shade the edges of a composition.
  ```gsl
  combined = unionSharp(s, r, t);
  edges = length([combined, combined.second]);   // 0 along the creases
  combined.color = ramp(edges, 0.005, 0.008, color.darkGray, color.white);
  return combined;
  ```

* **Bounding box:** `f.box` is an axis-aligned box `[lo, hi]` (two
vec3) containing the surface of `f`, where `f` is 0. It is
read from the builtin primitives (`sphere`, `box`, `boxCentered`,
`torus`, the capped and rounded cylinders and cones, `lineSegment`, the
axis cylinders, `octahedron`) and carried through `min` and the unions (the
union of the boxes), `max`, the intersections and the differences (their
intersection, or the target's box), offsets `f - r` (grown by `r`), scaling
by a number, a complement `-f`, `abs` and ternaries. A hand-written `remap` clears it; to move a
shape and keep its box, use the `transform` library (`transform.translate`,
`rotateX` / `rotateY` / `rotateZ`, `rotate`, `scale`, `affine`). An axis
it cannot bound reads as `-nothing` to `nothing`, and an empty box reads
with lo > hi. It follows sliders, and works in top-level expressions.
  ```gsl
  bracket = unionSharp(plate, boss) - 0.5;
  placed = transform.rotateZ(transform.translate(bracket, [40, 0, 0]), pi / 6);
  proxy = box(placed.box[0], placed.box[1]);   // its bounding box, as geometry
  ```

* **Replacement bounding box:** a field the box analysis cannot follow has no box, so the viewer cannot place the part or draw its stand-in when it is small on screen. Declare one with `name.box = [lo, hi]` for:
  - a `remap` that is not a translation, rotation or scale: a warp, bend, twist, shear or taper, such as `[p.x, p.y + sin(p.x), p.z]`;
  - a field written out in `p` that is not a plane, a distance to a point or a line, or one of those scaled by a number: `sin(p.x) * 0.1`, `f + g`, `f * g`;
  - an offset `f - r` where `r` depends on `p`.

  The corners must enclose the surface, because the viewer crops the surface to the box; one that is too large costs only speed. A declared box then flows through `transform`, the unions, the intersections and the differences like a primitive's, with one catch: it bounds the part's surface but not its solid, so cut holes in a shape first and add the declared part afterwards, as the example does.

* **map / reduce:** `map(fn, collection)` applies `fn(value, index)`
to each element and collects the results. A function that returns a vector
produces a list of vectors, which can be indexed, mapped, or reduced.
`reduce(initial, fn, collection)` folds `fn(accumulator, value)`
over the elements starting from the initial value; the accumulator may be a vector.

### **Operators**

All of the math related to vectors in GSL is Hadamard or component-wise vector algebra.

* **Arithmetic:** `+, -, *, /`
* **Modulo:** `%`
* **Comparison:** `==, !=, >, >=, <, <=`
* **Spaceship Operator (<=>):** Returns 1 if left > right, 0 if equal, -1 if left < right.
* **Logic:** 0.0 for false, otherwise true.
  * `&&` - logical and
  * `||` - logical or
  * `!` - logical not
* **Conditional (Ternary):** `condition ? trueVal : falseVal` (the condition is true when it is > 0)
* **Extended Conditional:** `value ? positiveVal : zeroVal : negativeVal`
Selects result based on whether value is positive, zero, or negative.  
Use surrounding parentheses when nesting!

* **Booleans (CSG):** combine shapes with `unionSharp`, `intersectSharp` and `differenceSharp`, each taking the shapes as arguments or as one list: `differenceSharp(plate, holes)`. To blend the seam instead, change the name and add the blend's parameters as the last arguments: `unionRound(a, b, 2)` fillets it with radius 2, and `Lp`, `Exponential`, `Chamfer`, `ChamferRound` and `Rvachev` do likewise (a radius, a width or a setback), each with a matching `difference`.
  * **A blend is a value.** Leave the shapes out and a blend returns the operator itself, which takes a list of shapes: `intersectRound(2)`. Hand it to anything that wants a blend. `boxCentered(vector.zero, [20, 20, 20], intersectRound(2))` rounds every edge of the box, and `intersectChamferBlend(a, b, setback, blend)` takes another blend to combine its chamfer planes. Take one as a parameter of your own function and it works with any blend; call it through the parameter's name:
  ```gsl
  joined(blend) { return blend([sphere(vector.zero, 10), sphere([8, 0, 0], 10)]); }
  fused = joined(unionRound(2));
  ```
  * The sharp ones are `min`, `max` and `max(a, -b)` underneath.

### **Operator Precedence**

| Precedence   | Operators                                   | Description                                                        |
| :---         | :---                                        | :---                                                               |
| 1  (Highest) | `(), f(), x, p.x`                 | Parentheses, function calls, literals, references, property access |
| 2            | `!, +, -`                         | Unary logical NOT, unary plus/minus                                |
| 3            | `*, /, %`                         | Multiplication, division, modulo                                   |
| 4            | `+, -`                            | Addition, subtraction                                              |
| 5            | `<, <=, >, >=`                    | Relational comparison                                              |
| 6            | `==, !=`                          | Equality                                                           |
| 7            | `&&`                              | Logical AND                                                        |
| 8            | `\|\|`                          | Logical OR                                                         |
| 9            | `?:, ?::`                         | Conditional (ternary / extended ternary)                           |
| 10 (Lowest)  | `=`                               | Assignment                                                         |

---

## **Example Program**

```gsl
// A mounting plate with four bolt holes, a ball and a wavy rib, in
// millimetres. Positions, sizes and offsets are vectors: write them as
// [x, y, z] and combine them with +, -, * and the vector functions.

plateSize = [60, 40, 6];
holeSpacing = [44, 24, 0];
boltRadius = 3;

// Library shapes take a centre and a size; there is nothing to define.
plate = boxCentered(vector.zero, plateSize);

// One hole per corner: scale the spacing by each corner's signs.
corners = [[1, 1, 0], [1, -1, 0], [-1, 1, 0], [-1, -1, 0]];
holes = unionSharp(map((signs) {
  return cylinderZ(signs * holeSpacing / 2, boltRadius);
}, corners));

// Move shapes with transform.translate, not a hand-written remap. The ball
// sits half sunk in the top face of the plate.
topFace = vector.dirZ * plateSize.z / 2;
ball = transform.translate(sphere(vector.zero, 10), topFace + [0, 0, 6]);

// A rib bent into a wave. The warp is a hand-written remap, which loses the
// shape's bounding box, so declare one whose corners enclose the wave.
rib = remap(boxCentered(vector.zero, [40, 4, 10]), [p.x, p.y - 2 * sin(p.x / 4), p.z]);
rib.box = [[-20, -4, -5], [20, 4, 5]];
ribOnPlate = transform.translate(rib, topFace + [0, 0, 5]);

// unionSharp and differenceSharp combine shapes. Cut the holes first, then add
// the rib: a declared box bounds the rib's surface but not its solid.
drilled = differenceSharp(unionSharp(plate, ball), holes);
body = unionSharp(drilled, ribOnPlate);

return body;
```

**Prefer the library.** Combine shapes with `unionSharp`, `intersectSharp` and `differenceSharp` rather than `min`, `max` and `max(a, neg(b))`, and move them with `transform.translate`, `rotateX` / `rotateY` / `rotateZ`, `rotate`, `scale` and `affine` rather than `remap`. Each gives the same field as the hand-written form, says what you mean, and keeps the shape's bounding box. Write `min`, `max` or `remap` yourself only for something the library has no function for.
