
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
The top-level function's return values are automatically combined using a union operation (min) to produce the final shape.
For example, "return s1, s2;" renders the union of shape "s1" and shape "s2".

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
`material` package (roughness, metallic, ior, emission; see its
presets, `material.make`, and the `with…` modifiers).
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
by a number, a complement `-f`, `abs` and ternaries. A `remap` clears it; to move a
shape and keep its box, use the `transform` library (`transform.translate`,
`rotateX` / `rotateY` / `rotateZ`, `rotate`, `scale`, `affine`). An axis
it cannot bound reads as `-nothing` to `nothing`, and an empty box reads
with lo > hi. It follows sliders, and works in top-level expressions.
  ```gsl
  bracket = unionSharp(plate, boss) - 0.5;
  placed = transform.rotateZ(transform.translate(bracket, [40, 0, 0]), pi / 6);
  proxy = box(placed.box[0], placed.box[1]);   // its bounding box, as geometry
  ```

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

* **CSG Operations (Constructive Solid Geometry):**
  * **Union:** `min(shape1, shape2, shape3, ...)`
  * **Intersection:** `max(shape1, shape2, shape3, ...)`
  * **Difference:** `max(shape1, -shape2)` (subtracts shape2 from shape1)

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
// Distance function for a sphere at the origin with a given radius
sphere(radius) {
	return length(p) - radius;
}

// Distance function for a rectangular solid with given bounds
// Note: p is the implicit point in space being evaluated (a 3D vector vec3<f32>)
box(x1, y1, z1, x2, y2, z2) {
	return max(
		x1 - p.x, p.x - x2,
		y1 - p.y, p.y - y2,
		z1 - p.z, p.z - z2,
	);
}

// Distance function for a torus at the origin with given major and minor radii
torus(majorRadius, minorRadius) {
	// Use swizzling to access x and y components
	q = [length(p.xy) - majorRadius, p.z];
	return length(q) - minorRadius;
}

s = sphere(0.8);
r = box(0.3, 0.3, 0.3, 1.2, 1.2, 1.2);

// Use remap to translate a torus by 1.0 along the Z axis
// remap re-evaluates the given shape expression with a new implicit 'p'
remappedTorus = remap(() { return torus(1.3, 0.2); }(), p - [0.0, 0.0, 1.0]);

// Intersect the sphere and the box (max is intersection)
intersection = max(s, r);

// The result is the union of the translated torus and the intersection shape
return intersection, remappedTorus;
```
