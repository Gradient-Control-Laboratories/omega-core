fn gsl_floored(a: f32, n: f32) -> f32 {
  return a - n * floor(a / n);
}

fn builtin_sphere(p: vec3<f32>, radius: f32, noiseTex: texture_3d<f32>) -> f32 {
  return length(p) - radius;
}

fn builtin_intersectSharp(fields: vec2<f32>, noiseTex: texture_3d<f32>) -> f32 {
  return max(fields.x, fields.y);
}

fn builtin_intersectRound(fields: vec2<f32>, radius: f32, noiseTex: texture_3d<f32>) -> f32 {
  let sharp = builtin_intersectSharp(fields, noiseTex);
  let sumOfSquares = 0.0 + max(fields.x + radius, 0.0) * max(fields.x + radius, 0.0) + max(fields.y + radius, 0.0) * max(fields.y + radius, 0.0);
  return sharp - max(sharp + radius, 0.0) + sqrt(max(sumOfSquares, 0.000000000000000000000000000001));
}

fn builtin_gcl_closure_intersectRound(fields: vec2<f32>, noiseTex: texture_3d<f32>) -> f32 {
  return builtin_intersectRound(fields, 0.3, noiseTex);
}

fn builtin_gcl_closure_cylinderCapped(p: vec3<f32>, pointA: vec3<f32>, pointB: vec3<f32>, radius: f32, noiseTex: texture_3d<f32>) -> f32 {
  let pa = p - pointA;
  let ba = pointB - pointA;
  let paba = dot(pa, ba);
  let baba = dot(ba, ba);
  let t = clamp(paba / baba, 0.0, 1.0);
  let radialDistance = length(pa - ba * t) - radius;
  let axialDistance = (abs(paba - baba * 0.5) - baba * 0.5) / sqrt(baba);
  let form = builtin_gcl_closure_intersectRound(vec2<f32>(radialDistance, axialDistance), noiseTex);
  return form;
}

fn map_triangleWave(x: f32, period: f32, noiseTex: texture_3d<f32>) -> f32 {
  let half = period * 0.5;
  let twoPeriod = period * 2.0;
  let u = gsl_floored((gsl_floored((x + half), twoPeriod) + twoPeriod), twoPeriod);
  return half - abs(u - period);
}

fn material_withDisplacement(m: mat4x4<f32>, height: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  let volume = m[2];
  return mat4x4<f32>(m[0], m[1], vec4<f32>(volume.x, volume.y, height, volume.w), m[3]);
}

fn gsl_shapeParts(p: vec3<f32>, time: f32, noiseTex: texture_3d<f32>) -> array<f32, 4> {
  gsl_partNear = vec3<f32>(1e10, 1e10, 1e10);
  let gsl_partLo = vec3<f32>(-16.5, -3.0, -3.0);
  let gsl_partHi = vec3<f32>(-10.5, 3.0, 3.0);
  let gsl_partBox = gsl_boxDistance(p, gsl_partLo, gsl_partHi);
  var gsl_part = gsl_partBox;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox <= gsl_vanishView.w) {
    let gsl_partVanish = gsl_vanish(gsl_partLo, gsl_partHi);
    gsl_part = 1e10;
    if (gsl_partVanish.x > 0.0) { gsl_part = builtin_sphere(p - vec3<f32>(-13.5, 0.0, 0.0), 3.0, noiseTex) + gsl_partVanish.y; }
    if (abs(gsl_part) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part), gsl_partVanish.zw); }
  }
  let smoothBall = gsl_part;
  let gsl_partLo_2 = vec3<f32>(-7.5, -3.0, -3.0);
  let gsl_partHi_2 = vec3<f32>(-1.5, 3.0, 3.0);
  let gsl_partBox_2 = gsl_boxDistance(p, gsl_partLo_2, gsl_partHi_2);
  var gsl_part_2 = gsl_partBox_2;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_2 <= gsl_vanishView.w) {
    let gsl_partVanish_2 = gsl_vanish(gsl_partLo_2, gsl_partHi_2);
    gsl_part_2 = 1e10;
    if (gsl_partVanish_2.x > 0.0) { gsl_part_2 = builtin_sphere(p - vec3<f32>(-4.5, 0.0, 0.0), 3.0, noiseTex) + gsl_partVanish_2.y; }
    if (abs(gsl_part_2) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_2), gsl_partVanish_2.zw); }
  }
  let peelBall = gsl_part_2;
  let gsl_partLo_3 = vec3<f32>(1.5, -3.0, -3.0);
  let gsl_partHi_3 = vec3<f32>(7.5, 3.0, 3.0);
  let gsl_partBox_3 = gsl_boxDistance(p, gsl_partLo_3, gsl_partHi_3);
  var gsl_part_3 = gsl_partBox_3;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_3 <= gsl_vanishView.w) {
    let gsl_partVanish_3 = gsl_vanish(gsl_partLo_3, gsl_partHi_3);
    gsl_part_3 = 1e10;
    if (gsl_partVanish_3.x > 0.0) { gsl_part_3 = builtin_sphere(p - vec3<f32>(4.5, 0.0, 0.0), 3.0, noiseTex) + gsl_partVanish_3.y; }
    if (abs(gsl_part_3) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_3), gsl_partVanish_3.zw); }
  }
  let roughBall = gsl_part_3;
  let gsl_partLo_4 = vec3<f32>(10.99, -2.51, -5.51);
  let gsl_partHi_4 = vec3<f32>(16.01, 2.51, 5.51);
  let gsl_partBox_4 = gsl_boxDistance(p, gsl_partLo_4, gsl_partHi_4);
  var gsl_part_4 = gsl_partBox_4;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_4 <= gsl_vanishView.w) {
    let gsl_partVanish_4 = gsl_vanish(gsl_partLo_4, gsl_partHi_4);
    gsl_part_4 = 1e10;
    if (gsl_partVanish_4.x > 0.0) { gsl_part_4 = builtin_gcl_closure_cylinderCapped(p - vec3<f32>(13.5, 0.0, 0.0), vec3<f32>(0.0, 0.0, -3.0), vec3<f32>(0.0, 0.0, 3.0), 2.5, noiseTex) + gsl_partVanish_4.y; }
    if (abs(gsl_part_4) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_4), gsl_partVanish_4.zw); }
  }
  let spool = gsl_part_4;
  return array<f32, 4>(smoothBall, peelBall, roughBall, spool);
}

fn sdShape(p: vec3<f32>, time: f32, noiseTex: texture_3d<f32>) -> gsl_shape {
  let gsl_d = gsl_shapeParts(p, time, noiseTex);
  let gsl_s1 = gsl_shape(gsl_d[0], 1u);
  let gsl_s2 = gsl_shape(gsl_d[1], 2u);
  let gsl_s3 = gsl_shape(gsl_d[2], 3u);
  let gsl_s4 = gsl_shape(gsl_d[3], 4u);
  return gsl_min(gsl_min(gsl_min(gsl_s1, gsl_s2), gsl_s3), gsl_s4);
}
fn gsl_colorField(p: vec3<f32>, time: f32, gsl_colorIdx: u32, noiseTex: texture_3d<f32>) -> vec4<f32> {
  if (gsl_colorIdx == 1u) {
    let color_teal = vec3<f32>(0.0, 0.5, 0.5);
    return vec4<f32>(color_teal, 1.0);
  }
  if (gsl_colorIdx == 2u) {
    let color_teal = vec3<f32>(0.0, 0.5, 0.5);
    return vec4<f32>(color_teal, 1.0);
  }
  if (gsl_colorIdx == 3u) {
    let color_teal = vec3<f32>(0.0, 0.5, 0.5);
    return vec4<f32>(color_teal, 1.0);
  }
  if (gsl_colorIdx == 4u) { return vec4<f32>(vec3<f32>(0.8373, 0.8204, 0.7966), 1.0); }
  return vec4<f32>(0.0, 0.0, 0.0, 0.0);
}
fn gsl_materialField(p: vec3<f32>, time: f32, gsl_materialIdx: u32, noiseTex: texture_3d<f32>) -> gsl_material {
  if (gsl_materialIdx == 1u) {
    let material_plastic = mat4x4<f32>(vec4<f32>(0.35, 0.0, 1.46, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    let gsl_m = material_plastic; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 2u) {
    let material_plastic = mat4x4<f32>(vec4<f32>(0.35, 0.0, 1.46, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    let gsl_tmp_1 = p / 1.5;
    let pebble = (gsl_cellularScan(gsl_tmp_1, 1, noiseTex).f2 - gsl_cellularScan(gsl_tmp_1, 1, noiseTex).f1) * 1.5;
    let gsl_m = material_withDisplacement(material_plastic, pebble * 0.2, noiseTex); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 3u) {
    let material_plastic = mat4x4<f32>(vec4<f32>(0.35, 0.0, 1.46, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    let gsl_tmp_1 = p / 1.5;
    let pebble = (gsl_cellularScan(gsl_tmp_1, 1, noiseTex).f2 - gsl_cellularScan(gsl_tmp_1, 1, noiseTex).f1) * 1.5;
    let gsl_m = material_withDisplacement(material_plastic, pebble * 0.6, noiseTex); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 4u) {
    let rings = map_triangleWave(p.z, 0.6, noiseTex);
    let gsl_m = material_withDisplacement(mat4x4<f32>(vec4<f32>(0.35, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), rings, noiseTex); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  return gsl_material(vec4<f32>(0.5, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), 0.0);
}
struct gsl_CellularScan {
  f1: f32,
  f2: f32,
  feature: vec3<f32>,
  farFeature: vec3<f32>,
  nearCell: vec3<f32>,
  farCell: vec3<f32>,
}

fn gsl_cellularHash(cell: vec3<f32>, oct: i32, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let fine = fract(textureLoad(noiseTex, vec3<i32>(floor(cell)) & vec3<i32>(63), 0).rgb);
  if (oct <= 1) { return fine; }
  // fract(fract(a) + b) == fract(a + b): the inner fract only removes an
  // integer, so folding the supercell on top of the already-fracted fine
  // lookup is the same value noiseHash3's two-lookup form produces.
  return fract(fine + textureLoad(noiseTex, vec3<i32>(floor(cell / 64.0)) & vec3<i32>(63), 0).rgb);
}

fn gsl_cellularScan(q: vec3<f32>, oct: i32, noiseTex: texture_3d<f32>) -> gsl_CellularScan {
  let pi = floor(q);
  let pf = q - pi;
  // Seeded with an unreachable distance, squared, because the scan compares
  // squared distances and roots the two winners at the end.
  var best = 64.000000;
  var second = 64.000000;
  var bestCell = pi;
  var bestH = vec3<f32>(0.0);
  var secondCell = pi;
  var secondH = vec3<f32>(0.0);
  for (var i = 0; i < 27; i = i + 1) {
    // The same offset order cellularOffsets() lists: x fastest, then y, then z.
    let o = vec3<f32>(f32(i % 3) - 1.0, f32((i / 3) % 3) - 1.0, f32(i / 9) - 1.0);
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
}

var<private> gsl_vanishEye: vec4<f32>;
var<private> gsl_vanishView: vec4<f32>;
var<private> gsl_partNear: vec3<f32>;

fn gsl_vanish(lo: vec3<f32>, hi: vec3<f32>) -> vec4<f32> {
  if (gsl_vanishEye.w <= 0.0) { return vec4<f32>(1.0, 0.0, 1e10, 1e10); }
  let size = hi - lo;
  let dist = length(gsl_vanishEye.xyz - clamp(gsl_vanishEye.xyz, lo, hi));
  let pixel = gsl_vanishEye.w * mix(max(dist, 1.0), 1.0, gsl_vanishView.x);
  let edge = min(size.x, min(size.y, size.z));
  let px = edge / pixel;
  let w = smoothstep(gsl_vanishView.y, max(gsl_vanishView.z, gsl_vanishView.y + 1e-3), px);
  return vec4<f32>(w, (1.0 - w) * 0.5 * edge, px, edge);
}

fn gsl_boxDistance(q: vec3<f32>, lo: vec3<f32>, hi: vec3<f32>) -> f32 {
  return length(max(max(lo - q, q - hi), vec3<f32>(0.0)));
}

fn gsl_heights(p: vec3<f32>, time: f32, noiseTex: texture_3d<f32>) -> array<f32, 4> {
  let material_plastic = mat4x4<f32>(vec4<f32>(0.35, 0.0, 1.46, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
  let gsl_tmp_1 = p / 1.5;
  let pebble = (gsl_cellularScan(gsl_tmp_1, 1, noiseTex).f2 - gsl_cellularScan(gsl_tmp_1, 1, noiseTex).f1) * 1.5;
  let rings = map_triangleWave(p.z, 0.6, noiseTex);
  return array<f32, 4>(0.0, material_withDisplacement(material_plastic, pebble * 0.2, noiseTex)[2].z, material_withDisplacement(material_plastic, pebble * 0.6, noiseTex)[2].z, material_withDisplacement(mat4x4<f32>(vec4<f32>(0.35, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), rings, noiseTex)[2].z);
}

fn gsl_sdShapeDisplaced(p: vec3<f32>, time: f32, gsl_dispK: f32, noiseTex: texture_3d<f32>) -> gsl_shape {
  let gsl_d = gsl_shapeParts(p, time, noiseTex);
  let gsl_h = gsl_heights(p, time, noiseTex);
  let gsl_s1 = gsl_shape(gsl_d[0], 1u);
  let gsl_s2 = gsl_shape(gsl_d[1] - gsl_h[1] * gsl_dispK, 2u);
  let gsl_s3 = gsl_shape(gsl_d[2] - gsl_h[2] * gsl_dispK, 3u);
  let gsl_s4 = gsl_shape(gsl_d[3] - gsl_h[3] * gsl_dispK, 4u);
  return gsl_min(gsl_min(gsl_min(gsl_s1, gsl_s2), gsl_s3), gsl_s4);
}
