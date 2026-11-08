package codegen

import (
	"fmt"

	noise "gradientcontrol.com/omega/core/program/noise"
	"gradientcontrol.com/omega/core/program"
)

// Cellular (Worley) noise, emitted as ONE helper with a real WGSL `for` loop.
//
// Written as a GSL library fold, the 27-neighbour scan is unrolled by the build
// layer at compile time, so every call site inlines 27 copies of the hash and the
// distance. That is why a heavily textured frame's shader is enormous and its
// Metal compile takes seconds — gpuparity's imageSkipExamples documents frames
// whose pixels move when unrelated dead code is added, which is a symptom of a
// shader that large. As a builtin the scan becomes a loop: the dynamic work is
// the same 27 iterations, but the emitted code is a few lines per call site
// instead of hundreds.
//
// One scan yields everything the cellular family needs, so the helper returns a
// struct and each builtin reads one field. The arithmetic must match
// common/noise.WorleyF1 (the Go reference) and tape-core/cellular.rs (the
// Fidget oracles) term for term, including the `pf - (o + h)` grouping and the
// offset order, or the backends disagree and gpuparity catches it.

// cellularField names the struct field each cellular builtin reads, and whether
// that field is a vec3. Kept beside the WGSL source so the two cannot drift.
type cellularField struct {
	field string
	vec3  bool
}

// The scan's own constants, named so the emitted shader holds no bare numbers.
//
//   - neighboursPerAxis / candidates: a feature point can lie in any immediate
//     neighbour when the jitter fills the cell, and no further, so 3 per axis is
//     both necessary and sufficient. Must match tape-core's NEIGHBOURS_PER_AXIS.
//   - unreachableDistance: larger than any possible F1 or F2 in cell units,
//     which is at most the half-diagonal of the 3-cell neighbourhood. The scan
//     seeds its running minima with its square, since it compares squared
//     distances.
//   - tileWrapMask / tileSpan: the baked tile is a power of two on each axis, so
//     the floored modulus is a bitmask (see the noiseHash3 emitter). Both derive
//     from noise.TileSize rather than restating 64 and 63.
const (
	neighboursPerAxis   = 3
	candidates          = neighboursPerAxis * neighboursPerAxis * neighboursPerAxis
	neighbourPlane      = neighboursPerAxis * neighboursPerAxis
	neighbourCentre     = neighboursPerAxis / 2
	unreachableDistance = 8.0
	tileWrapMask        = noise.TileSize - 1
	tileSpan            = noise.TileSize
)

// cellularWGSL is built rather than written out so the tile size and the
// neighbourhood size each appear once, from their defining constant.
var cellularWGSL = fmt.Sprintf(`struct gsl_CellularScan {
  f1: f32,
  f2: f32,
  feature: vec3<f32>,
  farFeature: vec3<f32>,
  nearCell: vec3<f32>,
  farCell: vec3<f32>,
}

fn gsl_cellularHash(cell: vec3<f32>, oct: i32, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let fine = fract(textureLoad(noiseTex, vec3<i32>(floor(cell)) & vec3<i32>(%[1]d), 0).rgb);
  if (oct <= 1) { return fine; }
  // fract(fract(a) + b) == fract(a + b): the inner fract only removes an
  // integer, so folding the supercell on top of the already-fracted fine
  // lookup is the same value noiseHash3's two-lookup form produces.
  return fract(fine + textureLoad(noiseTex, vec3<i32>(floor(cell / %[2]d.0)) & vec3<i32>(%[1]d), 0).rgb);
}

fn gsl_cellularScan(q: vec3<f32>, oct: i32, noiseTex: texture_3d<f32>) -> gsl_CellularScan {
  let pi = floor(q);
  let pf = q - pi;
  // Seeded with an unreachable distance, squared, because the scan compares
  // squared distances and roots the two winners at the end.
  var best = %[3]f;
  var second = %[3]f;
  var bestCell = pi;
  var bestH = vec3<f32>(0.0);
  var secondCell = pi;
  var secondH = vec3<f32>(0.0);
  for (var i = 0; i < %[4]d; i = i + 1) {
    // The same offset order cellularOffsets() lists: x fastest, then y, then z.
    let o = vec3<f32>(f32(i %% %[5]d) - %[7]d.0, f32((i / %[5]d) %% %[5]d) - %[7]d.0, f32(i / %[6]d) - %[7]d.0);
    let h = gsl_cellularHash(pi + o, oct, noiseTex);
    let v = pf - (o + h);
    let d2 = dot(v, v);
    if (d2 < best) {
      second = best;
      secondCell = bestCell;
      secondH = bestH;
      best = d2;
      bestCell = pi + o;
      bestH = h;
    } else if (d2 < second) {
      second = d2;
      secondCell = pi + o;
      secondH = h;
    }
  }
  var out: gsl_CellularScan;
  out.f1 = sqrt(best);
  out.f2 = sqrt(second);
  // Both closest points, so both gradients are available without a second hash.
  out.feature = bestCell + bestH;
  out.farFeature = secondCell + secondH;
  out.nearCell = bestCell;
  out.farCell = secondCell;
  return out;
}`,
	tileWrapMask, tileSpan, unreachableDistance*unreachableDistance,
	candidates, neighboursPerAxis, neighbourPlane, neighbourCentre)

// cellularFields is the builtin-to-field map the emitter dispatches on.
var cellularFields = map[string]cellularField{
	program.WorleyF1:         {field: "f1"},
	program.WorleyF2:         {field: "f2"},
	program.WorleyFeature:    {field: "feature", vec3: true},
	program.WorleyFarFeature: {field: "farFeature", vec3: true},
	program.WorleyNearCell:   {field: "nearCell", vec3: true},
	program.WorleyFarCell:    {field: "farCell", vec3: true},
}
