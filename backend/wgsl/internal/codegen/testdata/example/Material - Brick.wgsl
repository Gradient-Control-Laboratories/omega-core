fn ease_quadratic(t: f32, noiseTex: texture_3d<f32>) -> f32 {
  return t * t * (3.0 - 2.0 * t);
}

fn builtin_intersectSharp_2(d1: f32, d2: f32, noiseTex: texture_3d<f32>) -> f32 {
  return max(d1, d2);
}

fn ease_cubic(t: f32, noiseTex: texture_3d<f32>) -> f32 {
  return t * t * t * (t * (t * 6.0 - 15.0) + 10.0);
}

fn gsl_remap(p: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  return noise_gradientLattice(p / 37.0, noiseTex);
}

fn material_blend(a: mat4x4<f32>, b: mat4x4<f32>, t: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  return mat4x4<f32>(a[0] + (b[0] - a[0]) * t, a[1] + (b[1] - a[1]) * t, a[2] + (b[2] - a[2]) * t, a[3] + (b[3] - a[3]) * t);
}

fn builtin_intersectSharp(fields: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  return max(max(fields.x, fields.y), fields.z);
}

fn builtin_intersectRound(fields: vec3<f32>, radius: f32, noiseTex: texture_3d<f32>) -> f32 {
  let sharp = builtin_intersectSharp(fields, noiseTex);
  let sumOfSquares = 0.0 + max(fields.x + radius, 0.0) * max(fields.x + radius, 0.0) + max(fields.y + radius, 0.0) * max(fields.y + radius, 0.0) + max(fields.z + radius, 0.0) * max(fields.z + radius, 0.0);
  return sharp - max(sharp + radius, 0.0) + sqrt(max(sumOfSquares, 0.000000000000000000000000000001));
}

fn builtin_gcl_closure_intersectRound(fields: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  return builtin_intersectRound(fields, 0.0, noiseTex);
}

fn builtin_gcl_closure_boxCentered(p: vec3<f32>, center: vec3<f32>, size: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let halfSize = size * 0.5;
  let d = abs(p - center) - halfSize;
  let solid = builtin_gcl_closure_intersectRound(vec3<f32>(d.x, d.y, d.z), noiseTex);
  return solid;
}

fn color_srgbToLinearChannel(s: f32, noiseTex: texture_3d<f32>) -> f32 {
  return select(pow((s + 0.055) / 1.055, 2.4), s / 12.92, s <= 0.04045);
}

fn color_linearize(c: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  return vec3<f32>(color_srgbToLinearChannel(c.x, noiseTex), color_srgbToLinearChannel(c.y, noiseTex), color_srgbToLinearChannel(c.z, noiseTex));
}

fn color_cbrt(x: f32, noiseTex: texture_3d<f32>) -> f32 {
  return sign(x) * pow(abs(x), 0.33333334);
}

fn color_rgb2oklab(c: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let lin = color_linearize(c, noiseTex);
  let l = dot(vec3<f32>(0.41222146, 0.53633255, 0.051445995), lin);
  let l_ = color_cbrt(l, noiseTex);
  let m = dot(vec3<f32>(0.2119035, 0.6806995, 0.10739696), lin);
  let m_ = color_cbrt(m, noiseTex);
  let s = dot(vec3<f32>(0.08830246, 0.28171885, 0.6299787), lin);
  let s_ = color_cbrt(s, noiseTex);
  let lms = vec3<f32>(l_, m_, s_);
  return vec3<f32>(dot(vec3<f32>(0.21045426, 0.7936178, -0.004072047), lms), dot(vec3<f32>(1.9779985, -2.4285922, 0.4505937), lms), dot(vec3<f32>(0.025904037, 0.78277177, -0.80867577), lms));
}

fn color_linearToSrgbChannel(l: f32, noiseTex: texture_3d<f32>) -> f32 {
  return select(1.055 * pow(l, 0.41666666) - 0.055, l * 12.92, l <= 0.0031308);
}

fn color_gammaEncode(c: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  return vec3<f32>(color_linearToSrgbChannel(c.x, noiseTex), color_linearToSrgbChannel(c.y, noiseTex), color_linearToSrgbChannel(c.z, noiseTex));
}

fn color_oklab2rgb(c: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let l_ = c.x + 0.39633778 * c.y + 0.21580376 * c.z;
  let l = l_ * l_ * l_;
  let m_ = c.x - 0.105561346 * c.y - 0.06385417 * c.z;
  let m = m_ * m_ * m_;
  let s_ = c.x - 0.08948418 * c.y - 1.2914855 * c.z;
  let s = s_ * s_ * s_;
  let lms = vec3<f32>(l, m, s);
  let lin = vec3<f32>(dot(vec3<f32>(4.0767417, -3.3077116, 0.23096994), lms), dot(vec3<f32>(-1.268438, 2.6097574, -0.34131938), lms), dot(vec3<f32>(-0.0041960864, -0.7034186, 1.7076147), lms));
  return color_gammaEncode(lin, noiseTex);
}

fn color_mixOklab(a: vec3<f32>, b: vec3<f32>, t: f32, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let la = color_rgb2oklab(a, noiseTex);
  let lb = color_rgb2oklab(b, noiseTex);
  return color_oklab2rgb(la + (lb - la) * t, noiseTex);
}

fn material_withDisplacement(m: mat4x4<f32>, height: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  let volume = m[2];
  return mat4x4<f32>(m[0], m[1], vec4<f32>(volume.x, volume.y, height, volume.w), m[3]);
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

fn texture_mask(field: f32, edge: f32, width: f32, noiseTex: texture_3d<f32>) -> f32 {
  let t = clamp((field - edge) / width * 0.5 + 0.5, 0.0, 1.0);
  return ease_quadratic(t, noiseTex);
}

fn texture_paint(field: f32, edge: f32, width: f32, colorA: vec3<f32>, colorB: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  return color_mixOklab(colorA, colorB, texture_mask(field, edge, width, noiseTex), noiseTex);
}

fn texture_paintMaterial(field: f32, edge: f32, width: f32, matA: mat4x4<f32>, matB: mat4x4<f32>, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  return material_blend(matA, matB, texture_mask(field, edge, width, noiseTex), noiseTex);
}

fn headJoint(course: f32, k: f32, noiseTex: texture_3d<f32>) -> f32 {
  return (k + (fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(k, course, k + course + 5.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(k, course, k + course + 5.0)) / 64.0)) & vec3<i32>(63), 0).rgb).x - 0.5) * 0.036) * 225.0;
}

fn sdShape(p: vec3<f32>, time: f32, noiseTex: texture_3d<f32>) -> gsl_shape {
  gsl_partNear = vec3<f32>(1e10, 1e10, 1e10);
  let gsl_partLo = vec3<f32>(-300.0, -50.0, 0.0);
  let gsl_partHi = vec3<f32>(300.0, 50.0, 400.0);
  let gsl_partBox = gsl_boxDistance(p, gsl_partLo, gsl_partHi);
  var gsl_part = gsl_partBox;
  let course = floor(p.z / 75.0);
  let courseHash = fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(course, course + 3.0, 2.0 * course + 11.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(course, course + 3.0, 2.0 * course + 11.0)) / 64.0)) & vec3<i32>(63), 0).rgb);
  let shifted = p.x + 225.0 * fract(course * 0.5) + (courseHash.x - 0.5) * 30.0;
  let slot = floor(shifted / 225.0);
  let after = shifted >= headJoint(course, slot + 1.0, noiseTex);
  let before = shifted < headJoint(course, slot, noiseTex);
  let brickEnd = select(select(headJoint(course, slot + 1.0, noiseTex), headJoint(course, slot + 2.0, noiseTex), after), headJoint(course, slot, noiseTex), before);
  let brickStart = select(select(headJoint(course, slot, noiseTex), headJoint(course, slot + 1.0, noiseTex), after), headJoint(course, slot - 1.0, noiseTex), before);
  let brickMid = (brickStart + brickEnd) * 0.5;
  let alongBrick = shifted - brickMid;
  let brickHash = fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(brickMid, course, brickMid + course))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(brickMid, course, brickMid + course)) / 64.0)) & vec3<i32>(63), 0).rgb);
  let upBrick = p.z - (course + 0.5) * 75.0 - (brickHash.z - 0.5) * 3.0;
  let toJoint = min((brickEnd - brickStart) * 0.5 - abs(alongBrick), 37.5 - abs(upBrick));
  let inJoint = 1.0 - texture_mask(toJoint, 5.0, 0.8, noiseTex);
  let proud = (brickHash.x - 0.5) * 2.0 + (brickHash.y - 0.5) * 2.4 * alongBrick / 225.0;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox <= gsl_vanishView.w) {
    let gsl_partVanish = gsl_vanish(gsl_partLo, gsl_partHi);
    gsl_part = 1e10;
    if (gsl_partVanish.x > 0.0) { gsl_part = builtin_intersectSharp_2(builtin_gcl_closure_boxCentered(p, vec3<f32>(0.0, 0.0, 200.0), vec3<f32>(600.0, 100.0, 400.0), noiseTex), -p.y - 50.0 + 3.0 * inJoint - proud * (1.0 - inJoint), noiseTex) + gsl_partVanish.y; }
    if (abs(gsl_part) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part), gsl_partVanish.zw); }
  }
  let wall = gsl_part;
  let gsl_s1 = gsl_shape(wall, 1u);
  return gsl_s1;
}

fn gsl_colorField(p: vec3<f32>, time: f32, gsl_colorIdx: u32, noiseTex: texture_3d<f32>) -> vec4<f32> {
  if (gsl_colorIdx == 1u) {
    let course = floor(p.z / 75.0);
    let courseHash = fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(course, course + 3.0, 2.0 * course + 11.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(course, course + 3.0, 2.0 * course + 11.0)) / 64.0)) & vec3<i32>(63), 0).rgb);
    let shifted = p.x + 225.0 * fract(course * 0.5) + (courseHash.x - 0.5) * 30.0;
    let slot = floor(shifted / 225.0);
    let after = shifted >= headJoint(course, slot + 1.0, noiseTex);
    let before = shifted < headJoint(course, slot, noiseTex);
    let brickEnd = select(select(headJoint(course, slot + 1.0, noiseTex), headJoint(course, slot + 2.0, noiseTex), after), headJoint(course, slot, noiseTex), before);
    let brickStart = select(select(headJoint(course, slot, noiseTex), headJoint(course, slot + 1.0, noiseTex), after), headJoint(course, slot - 1.0, noiseTex), before);
    let brickMid = (brickStart + brickEnd) * 0.5;
    let alongBrick = shifted - brickMid;
    let brickHash = fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(brickMid, course, brickMid + course))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(brickMid, course, brickMid + course)) / 64.0)) & vec3<i32>(63), 0).rgb);
    let upBrick = p.z - (course + 0.5) * 75.0 - (brickHash.z - 0.5) * 3.0;
    let toJoint = min((brickEnd - brickStart) * 0.5 - abs(alongBrick), 37.5 - abs(upBrick));
    let brickCentre = vec3<f32>(brickMid, p.y, (course + 0.5) * 75.0);
    let tone = gsl_remap(brickCentre, noiseTex);
    let brickTone = color_mixOklab(vec3<f32>(0.5488, 0.3406, 0.274), vec3<f32>(0.38, 0.2, 0.15), texture_mask(tone, 0.05, 0.2, noiseTex), noiseTex);
    let mottle = noise_gradientLattice(p / 6.0, noiseTex);
    let brickFace = color_mixOklab(brickTone, vec3<f32>(0.66, 0.45, 0.36), 0.35 * texture_mask(mottle, 0.1, 0.3, noiseTex), noiseTex);
    return vec4<f32>(texture_paint(toJoint, 5.0, 0.8, vec3<f32>(0.7275456, 0.71601796, 0.6752879), brickFace, noiseTex), 1.0);
  }
  return vec4<f32>(0.0, 0.0, 0.0, 0.0);
}
fn gsl_materialField(p: vec3<f32>, time: f32, gsl_materialIdx: u32, noiseTex: texture_3d<f32>) -> gsl_material {
  if (gsl_materialIdx == 1u) {
    let course = floor(p.z / 75.0);
    let courseHash = fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(course, course + 3.0, 2.0 * course + 11.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(course, course + 3.0, 2.0 * course + 11.0)) / 64.0)) & vec3<i32>(63), 0).rgb);
    let shifted = p.x + 225.0 * fract(course * 0.5) + (courseHash.x - 0.5) * 30.0;
    let slot = floor(shifted / 225.0);
    let after = shifted >= headJoint(course, slot + 1.0, noiseTex);
    let before = shifted < headJoint(course, slot, noiseTex);
    let brickEnd = select(select(headJoint(course, slot + 1.0, noiseTex), headJoint(course, slot + 2.0, noiseTex), after), headJoint(course, slot, noiseTex), before);
    let brickStart = select(select(headJoint(course, slot, noiseTex), headJoint(course, slot + 1.0, noiseTex), after), headJoint(course, slot - 1.0, noiseTex), before);
    let brickMid = (brickStart + brickEnd) * 0.5;
    let alongBrick = shifted - brickMid;
    let brickHash = fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(brickMid, course, brickMid + course))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(brickMid, course, brickMid + course)) / 64.0)) & vec3<i32>(63), 0).rgb);
    let upBrick = p.z - (course + 0.5) * 75.0 - (brickHash.z - 0.5) * 3.0;
    let toJoint = min((brickEnd - brickStart) * 0.5 - abs(alongBrick), 37.5 - abs(upBrick));
    let mottle = noise_gradientLattice(p / 6.0, noiseTex);
    let chipped = (1.0 - texture_mask(toJoint, 8.5, 2.0, noiseTex)) * (1.0 - texture_mask(mottle, -0.2, 0.2, noiseTex));
    let brickRelief = 0.35 * mottle - 1.0 * chipped;
    let mortarRelief = 0.25 * (fract(textureLoad(noiseTex, vec3<i32>(floor(floor(p / 0.8) + vec3<f32>(7.0, 0.0, 0.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((floor(p / 0.8) + vec3<f32>(7.0, 0.0, 0.0)) / 64.0)) & vec3<i32>(63), 0).rgb).x - 0.5);
    let inJoint = 1.0 - texture_mask(toJoint, 5.0, 0.8, noiseTex);
    let relief = brickRelief + (mortarRelief - brickRelief) * inJoint;
    let gsl_m = material_withDisplacement(texture_paintMaterial(toJoint, 5.0, 0.8, mat4x4<f32>(vec4<f32>(0.5, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), mat4x4<f32>(vec4<f32>(1.0, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), noiseTex), relief, noiseTex); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
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
