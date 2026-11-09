fn builtin_intersectSharp(fields: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  return max(max(fields.x, fields.y), fields.z);
}

fn builtin_intersectLp(fields: vec3<f32>, width: f32, exponent: f32, noiseTex: texture_3d<f32>) -> f32 {
  let peak = builtin_intersectSharp(fields, noiseTex);
  let peakInset = max(peak + width, 0.0);
  let normalizer = max(peakInset, 0.000000001);
  let sumOfPowers = 0.0 + pow(max(fields.x + width, 0.0) / normalizer, exponent) + pow(max(fields.y + width, 0.0) / normalizer, exponent) + pow(max(fields.z + width, 0.0) / normalizer, exponent);
  return min(-width, peak) + peakInset * pow(sumOfPowers, 1.0 / exponent);
}

fn builtin_gcl_closure_intersectLp(width: f32, exponent: f32, fields: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  return builtin_intersectLp(fields, width, exponent, noiseTex);
}

fn builtin_cropToBox(p: vec3<f32>, shape: f32, low: vec3<f32>, high: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let lowCorner = min(low, high) - 0.01;
  let highCorner = max(low, high) + 0.01;
  let d = abs(p - (lowCorner + highCorner) * 0.5) - (highCorner - lowCorner) * 0.5;
  let outside = length(vec3<f32>(max(d.x, 0.0), max(d.y, 0.0), max(d.z, 0.0)));
  let inside = min(max(d.x, max(d.y, d.z)), 0.0);
  let gsl_capped = max(shape, outside + inside);
  if (gsl_boxCap) { return gsl_capped; }
  if (gsl_capped <= max(shape, 0.0)) { return shape; }
  return select(-1.0, 1.0, shape >= 0.0) * max(abs(shape), gsl_capped);
}

var<private> gsl_boxCap: bool = false;

fn gsl_remap(p: vec3<f32>, gsl_param_23_5: f32, noiseTex: texture_3d<f32>) -> f32 {
  let P = gsl_param_23_5;
  let gsl_tmp_5 = abs(p - vec3<f32>(0.0, 0.0, 0.0)) - vec3<f32>(3.5, 3.5, 3.5);
  return builtin_cropToBox(p, builtin_gcl_closure_intersectLp(0.25 * P, P, vec3<f32>(gsl_tmp_5.x, gsl_tmp_5.y, gsl_tmp_5.z), noiseTex), vec3<f32>(-3.5, -3.5, -3.5), vec3<f32>(3.5, 3.5, 3.5), noiseTex);
}

fn ease_cubic(t: f32, noiseTex: texture_3d<f32>) -> f32 {
  return t * t * t * (t * (t * 6.0 - 15.0) + 10.0);
}

fn gsl_remap_2(p: vec3<f32>, gsl_param_23_5: f32, material_plastic: mat4x4<f32>, noiseTex: texture_3d<f32>) -> f32 {
  return noise_gradientLattice(p / 4.0, noiseTex) + 0.5 * noise_gradientLattice(p / 1.3333334, noiseTex);
}

fn gsl_remap_3(p: vec3<f32>, gsl_param_23_5: f32, material_plastic: mat4x4<f32>, noiseTex: texture_3d<f32>) -> f32 {
  return noise_gradientLattice(p / 4.0, noiseTex);
}

fn builtin_intersectExponential(fields: vec3<f32>, width: f32, noiseTex: texture_3d<f32>) -> f32 {
  let peak = builtin_intersectSharp(fields, noiseTex);
  let scale = width / 2.7182817;
  let total = 0.0 + exp((fields.x - peak) / scale) + exp((fields.y - peak) / scale) + exp((fields.z - peak) / scale);
  return peak + scale * log(total);
}

fn builtin_gcl_closure_intersectExponential(fields: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  return builtin_intersectExponential(fields, 0.5, noiseTex);
}

fn builtin_gcl_closure_boxCentered_2(p: vec3<f32>, center: vec3<f32>, size: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let halfSize = size * 0.5;
  let d = abs(p - center) - halfSize;
  let solid = builtin_gcl_closure_intersectExponential(vec3<f32>(d.x, d.y, d.z), noiseTex);
  return solid;
}

fn builtin_intersectRound(fields: vec3<f32>, radius: f32, noiseTex: texture_3d<f32>) -> f32 {
  let sharp = builtin_intersectSharp(fields, noiseTex);
  let sumOfSquares = 0.0 + max(fields.x + radius, 0.0) * max(fields.x + radius, 0.0) + max(fields.y + radius, 0.0) * max(fields.y + radius, 0.0) + max(fields.z + radius, 0.0) * max(fields.z + radius, 0.0);
  return sharp - max(sharp + radius, 0.0) + sqrt(max(sumOfSquares, 0.000000000000000000000000000001));
}

fn builtin_gcl_closure_intersectRound(fields: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  return builtin_intersectRound(fields, 0.0, noiseTex);
}

fn builtin_gcl_closure_boxCentered_3(p: vec3<f32>, center: vec3<f32>, size: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let halfSize = size * 0.5;
  let d = abs(p - center) - halfSize;
  let solid = builtin_gcl_closure_intersectRound(vec3<f32>(d.x, d.y, d.z), noiseTex);
  return solid;
}

fn material_withRoughness(m: mat4x4<f32>, rough: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  let surface = m[0];
  return mat4x4<f32>(vec4<f32>(rough, surface.y, surface.z, surface.w), m[1], m[2], m[3]);
}

fn material_withIor(m: mat4x4<f32>, index: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  let surface = m[0];
  return mat4x4<f32>(vec4<f32>(surface.x, surface.y, index, surface.w), m[1], m[2], m[3]);
}

fn noise_cellHash3(cell: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  return fract(textureLoad(noiseTex, vec3<i32>(floor(cell)) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((cell) / 64.0)) & vec3<i32>(63), 0).rgb);
}

fn noise_cornerGradient(cell: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  return normalize(noise_cellHash3(cell, noiseTex) * 2.0 - 1.0);
}

fn noise_gradientLattice_gsl_anon(pi: vec3<f32>, pf: vec3<f32>, u: f32, v: f32, w: f32, acc: f32, o: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let wx = o.x * u + (1.0 - o.x) * (1.0 - u);
  let wy = o.y * v + (1.0 - o.y) * (1.0 - v);
  let wz = o.z * w + (1.0 - o.z) * (1.0 - w);
  let g = noise_cornerGradient(pi + o, noiseTex);
  let d = pf - o;
  return acc + wx * wy * wz * dot(g, d);
}

fn noise_gradientLattice(q: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let pi = floor(q);
  let pf = q - pi;
  let u = ease_cubic(pf.x, noiseTex);
  let v = ease_cubic(pf.y, noiseTex);
  let w = ease_cubic(pf.z, noiseTex);
  var gsl_fold = 0.0;
  for (var gsl_i = 0; gsl_i < 8; gsl_i = gsl_i + 1) {
    gsl_fold = noise_gradientLattice_gsl_anon(pi, pf, u, v, w, gsl_fold, gsl_foldK[gsl_i], noiseTex);
  }
  return gsl_fold;
}

fn gsl_pin_remap(p: vec3<f32>, x: f32, noiseTex: texture_3d<f32>) -> f32 {
  let gsl_tmp_1 = length(p.xy);
  let gsl_tmp_3 = 6.2831855 * p.z / 0.25;
  let gsl_tmp_2 = max(length(p.xy), 0.2);
  return max(max(max(length(p.xy) - 0.6, -max(0.46468353 - gsl_tmp_1, ((1.0 - acos(clamp(cos(gsl_tmp_3) * p.x / gsl_tmp_2 + sin(gsl_tmp_3) * p.y / gsl_tmp_2, -1.0, 1.0)) / 3.1415927) * 0.125 - 0.03125 - (gsl_tmp_1 - 0.46468353) * 0.57735026) * 0.8604493)), abs(p.z) - 2.5), (length(p.xy) - 0.6 + abs(p.z) - 2.5 + 0.15) * 0.7071);
}

fn pin(p: vec3<f32>, x: f32, noiseTex: texture_3d<f32>) -> f32 {
  return gsl_pin_remap(p - vec3<f32>(x, 0.0, 0.0), x, noiseTex);
}

fn sdShape(p: vec3<f32>, time: f32, gsl_param_23_5: f32, noiseTex: texture_3d<f32>) -> gsl_shape {
  gsl_partNear = vec3<f32>(1e10, 1e10, 1e10);
  let lp = gsl_remap(p - vec3<f32>(-5.0, 0.0, 0.0), gsl_param_23_5, noiseTex);
  let exponential = builtin_gcl_closure_boxCentered_2(p - vec3<f32>(5.0, 0.0, 0.0), vec3<f32>(0.0, 0.0, 0.0) * 0.5, abs(vec3<f32>(7.0, 7.0, 7.0)), noiseTex);
  let pinLp = pin(p, -5.0, noiseTex);
  let pinExp = pin(p, 5.0, noiseTex);
  let gsl_partLo = vec3<f32>(-12.0, -7.0, -4.49);
  let gsl_partHi = vec3<f32>(12.0, 7.0, -3.49);
  let gsl_partBox = gsl_boxDistance(p, gsl_partLo, gsl_partHi);
  var gsl_part = gsl_partBox;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox <= gsl_vanishView.w) {
    let gsl_partVanish = gsl_vanish(gsl_partLo, gsl_partHi);
    gsl_part = 1e10;
    if (gsl_partVanish.x > 0.0) { gsl_part = builtin_gcl_closure_boxCentered_3(p, vec3<f32>(0.0, 0.0, -3.99), vec3<f32>(24.0, 14.0, 1.0), noiseTex) + gsl_partVanish.y; }
    if (abs(gsl_part) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part), gsl_partVanish.zw); }
  }
  let board = gsl_part;
  let gsl_s1 = gsl_shape(lp, 1u);
  let gsl_s2 = gsl_shape(exponential, 2u);
  let gsl_s3 = gsl_shape(pinLp, 3u);
  let gsl_s4 = gsl_shape(pinExp, 4u);
  let gsl_s5 = gsl_shape(board, 5u);
  return gsl_min(gsl_min(gsl_min(gsl_min(gsl_s1, gsl_s2), gsl_s3), gsl_s4), gsl_s5);
}

fn gsl_colorField(p: vec3<f32>, time: f32, gsl_colorIdx: u32, gsl_param_23_5: f32, noiseTex: texture_3d<f32>) -> vec4<f32> {
  if (gsl_colorIdx == 1u) { return vec4<f32>(vec3<f32>(1.0, 1.0, 1.0), 1.0); }
  if (gsl_colorIdx == 2u) { return vec4<f32>(vec3<f32>(1.0, 1.0, 1.0), 1.0); }
  if (gsl_colorIdx == 3u) { return vec4<f32>(vec3<f32>(0.8373, 0.8204, 0.7966), 1.0); }
  if (gsl_colorIdx == 4u) { return vec4<f32>(vec3<f32>(0.8373, 0.8204, 0.7966), 1.0); }
  if (gsl_colorIdx == 5u) {
    let lodge = p.x + 3.0 * abs(2.0 * fract(p.y / 6.0) - 1.0);
    return vec4<f32>(select(vec3<f32>(0.1, 0.035, 0.035), vec3<f32>(0.84, 0.86, 0.9), fract(lodge / 2.0) < 0.5), 1.0);
  }
  return vec4<f32>(0.0, 0.0, 0.0, 0.0);
}
fn gsl_materialField(p: vec3<f32>, time: f32, gsl_materialIdx: u32, gsl_param_23_5: f32, noiseTex: texture_3d<f32>) -> gsl_material {
  if (gsl_materialIdx == 1u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.02, 0.0, 1.52, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 2u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.02, 0.0, 1.52, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 3u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.35, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 4u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.35, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 5u) {
    let material_plastic = mat4x4<f32>(vec4<f32>(0.35, 0.0, 1.46, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    let floorPoint = p - vec3<f32>(select(5.0, -5.0, p.x < 0.0), 0.0, 0.0);
    let gsl_m = material_withIor(material_withRoughness(material_plastic, 0.22 + 0.25 * gsl_remap_2(floorPoint, gsl_param_23_5, material_plastic, noiseTex), noiseTex), 1.5 + 0.5 * gsl_remap_3(floorPoint, gsl_param_23_5, material_plastic, noiseTex), noiseTex); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  return gsl_material(vec4<f32>(0.5, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), 0.0);
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

const gsl_foldK = array<vec3<f32>, 8>(vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 0.0, 0.0), vec3<f32>(0.0, 1.0, 0.0), vec3<f32>(1.0, 1.0, 0.0), vec3<f32>(0.0, 0.0, 1.0), vec3<f32>(1.0, 0.0, 1.0), vec3<f32>(0.0, 1.0, 1.0), vec3<f32>(1.0, 1.0, 1.0));
