fn ease_quadratic(t: f32, noiseTex: texture_3d<f32>) -> f32 {
  return t * t * (3.0 - 2.0 * t);
}

fn texture_mask(field: f32, edge: f32, width: f32, noiseTex: texture_3d<f32>) -> f32 {
  let t = clamp((field - edge) / width * 0.5 + 0.5, 0.0, 1.0);
  return ease_quadratic(t, noiseTex);
}

fn texture_coverageDisc(gsl_var_distance: f32, radius: f32, noiseTex: texture_3d<f32>) -> f32 {
  let drawn = max(radius, 0.5 * gsl_pw);
  let soft = max(0.5 * gsl_pw, 0.01 * radius);
  return (1.0 - texture_mask(gsl_var_distance, drawn, soft, noiseTex)) * radius / drawn * radius / drawn;
}

fn texture_fadeDetail(size: f32, noiseTex: texture_3d<f32>) -> f32 {
  return 1.0 - texture_mask(gsl_pw, 0.375 * size, 0.125 * size, noiseTex);
}

fn texture_fadeTo(gsl_var_value: f32, mean: f32, size: f32, noiseTex: texture_3d<f32>) -> f32 {
  return mean + (gsl_var_value - mean) * texture_fadeDetail(size, noiseTex);
}

fn builtin_sphere(p: vec3<f32>, radius: f32, noiseTex: texture_3d<f32>) -> f32 {
  return length(p) - radius;
}

fn texture_squareWaveIntegral(x: f32, period: f32, noiseTex: texture_3d<f32>) -> f32 {
  let whole = floor(x / period);
  return whole * 0.5 * period + max(x - whole * period - 0.5 * period, 0.0);
}

fn texture_squareWave(x: f32, period: f32, noiseTex: texture_3d<f32>) -> f32 {
  let exact = select(1.0, 0.0, fract(x / period) < 0.5);
  let width = gsl_pw;
  let averaged = (texture_squareWaveIntegral(x + 0.5 * width, period, noiseTex) - texture_squareWaveIntegral(x - 0.5 * width, period, noiseTex)) / max(width, 0.000001);
  return select(exact, averaged, width > 0.0001 * period);
}

fn noise_cellularHash(cell: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  return fract(textureLoad(noiseTex, vec3<i32>(floor(cell)) & vec3<i32>(63), 0).rgb);
}

fn ease_cubic(t: f32, noiseTex: texture_3d<f32>) -> f32 {
  return t * t * t * (t * (t * 6.0 - 15.0) + 10.0);
}

fn noise_cellularEdgesZBlendLattice_gsl_anon(pi: vec2<f32>, zBand: f32, zBlendT: f32, pf: vec2<f32>, a: vec2<f32>, o2: vec2<f32>, noiseTex: texture_3d<f32>) -> vec2<f32> {
  let cellLo = vec3<f32>(pi.x + o2.x, pi.y + o2.y, zBand);
  let hLo = noise_cellularHash(cellLo, noiseTex);
  let cellHi = vec3<f32>(pi.x + o2.x, pi.y + o2.y, zBand + 1.0);
  let hHi = noise_cellularHash(cellHi, noiseTex);
  let h = hLo + (hHi - hLo) * ease_cubic(zBlendT, noiseTex);
  let v = pf - (o2 + vec2<f32>(h.x, h.y));
  let d = dot(v, v);
  return vec2<f32>(min(a.x, d), min(max(a.x, d), a.y));
}

fn noise_cellularEdgesZBlendLattice(qxy: vec2<f32>, zBand: f32, zBlendT: f32, noiseTex: texture_3d<f32>) -> gsl_struct {
  let pi = floor(qxy);
  let pf = qxy - pi;
  var gsl_fold = vec2<f32>(64.0, 64.0);
  for (var gsl_i = 0; gsl_i < 9; gsl_i = gsl_i + 1) {
    gsl_fold = noise_cellularEdgesZBlendLattice_gsl_anon(pi, zBand, zBlendT, pf, gsl_fold, gsl_foldK[gsl_i], noiseTex);
  }
  let acc = gsl_fold;
  var gsl_result : gsl_struct;
  gsl_result.v_1 = sqrt(acc.x);
  gsl_result.v_2 = sqrt(acc.y);
  return gsl_result;
}

fn noise_cellularEdgesZBlend(p: vec3<f32>, size: f32, zBandSize: f32, noiseTex: texture_3d<f32>) -> f32 {
  let qxy = vec2<f32>(p.x / size, p.y / size);
  let zBand = floor(p.z / zBandSize);
  let zBlendT = fract(p.z / zBandSize);
  let gsl_result = noise_cellularEdgesZBlendLattice(qxy, zBand, zBlendT, noiseTex);
  let f2 = gsl_result.v_2;
  let f1 = gsl_result.v_1;
  return (f2 - f1) * size;
}

fn texture_coverageLine(gsl_var_distance: f32, width: f32, noiseTex: texture_3d<f32>) -> f32 {
  let half = max(0.5 * width, 0.5 * gsl_pw);
  return (1.0 - texture_mask(gsl_var_distance, half, 0.25 * half, noiseTex)) * 0.5 * width / half;
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

fn material_blend(a: mat4x4<f32>, b: mat4x4<f32>, t: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  return mat4x4<f32>(a[0] + (b[0] - a[0]) * t, a[1] + (b[1] - a[1]) * t, a[2] + (b[2] - a[2]) * t, a[3] + (b[3] - a[3]) * t);
}

fn playful_glitterShare(density: f32, noiseTex: texture_3d<f32>) -> f32 {
  return 1.0 - exp(-4.18879 * density * density * density);
}

fn playful_glitterFlakes(f1: f32, flakeCellSize: f32, density: f32, noiseTex: texture_3d<f32>) -> f32 {
  return texture_fadeTo(texture_coverageDisc(f1, density * flakeCellSize, noiseTex), playful_glitterShare(density, noiseTex), flakeCellSize, noiseTex);
}

fn sdShape(p: vec3<f32>, time: f32, noiseTex: texture_3d<f32>) -> gsl_shape {
  gsl_partNear = vec3<f32>(1e10, 1e10, 1e10);
  let gsl_partLo = vec3<f32>(-75.0, -25.0, -25.0);
  let gsl_partHi = vec3<f32>(-25.0, 25.0, 25.0);
  let gsl_partBox = gsl_boxDistance(p, gsl_partLo, gsl_partHi);
  var gsl_part = gsl_partBox;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox <= gsl_vanishView.w) {
    let gsl_partVanish = gsl_vanish(gsl_partLo, gsl_partHi);
    gsl_part = 1e10;
    if (gsl_partVanish.x > 0.0) { gsl_part = builtin_sphere(p - vec3<f32>(-50.0, 0.0, 0.0), 25.0, noiseTex) + gsl_partVanish.y; }
    if (abs(gsl_part) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part), gsl_partVanish.zw); }
  }
  let glitterBall = gsl_part;
  let gsl_partLo_2 = vec3<f32>(-25.0, -25.0, -25.0);
  let gsl_partHi_2 = vec3<f32>(25.0, 25.0, 25.0);
  let gsl_partBox_2 = gsl_boxDistance(p, gsl_partLo_2, gsl_partHi_2);
  var gsl_part_2 = gsl_partBox_2;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_2 <= gsl_vanishView.w) {
    let gsl_partVanish_2 = gsl_vanish(gsl_partLo_2, gsl_partHi_2);
    gsl_part_2 = 1e10;
    if (gsl_partVanish_2.x > 0.0) { gsl_part_2 = builtin_sphere(p - vec3<f32>(0.0, 0.0, 0.0), 25.0, noiseTex) + gsl_partVanish_2.y; }
    if (abs(gsl_part_2) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_2), gsl_partVanish_2.zw); }
  }
  let checkerBall = gsl_part_2;
  let gsl_partLo_3 = vec3<f32>(25.0, -25.0, -25.0);
  let gsl_partHi_3 = vec3<f32>(75.0, 25.0, 25.0);
  let gsl_partBox_3 = gsl_boxDistance(p, gsl_partLo_3, gsl_partHi_3);
  var gsl_part_3 = gsl_partBox_3;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_3 <= gsl_vanishView.w) {
    let gsl_partVanish_3 = gsl_vanish(gsl_partLo_3, gsl_partHi_3);
    gsl_part_3 = 1e10;
    if (gsl_partVanish_3.x > 0.0) { gsl_part_3 = builtin_sphere(p - vec3<f32>(50.0, 0.0, 0.0), 25.0, noiseTex) + gsl_partVanish_3.y; }
    if (abs(gsl_part_3) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_3), gsl_partVanish_3.zw); }
  }
  let skinBall = gsl_part_3;
  let gsl_s1 = gsl_shape(glitterBall, 1u);
  let gsl_s2 = gsl_shape(checkerBall, 2u);
  let gsl_s3 = gsl_shape(skinBall, 3u);
  return gsl_min(gsl_min(gsl_s1, gsl_s2), gsl_s3);
}

fn gsl_colorField(p: vec3<f32>, time: f32, gsl_colorIdx: u32, noiseTex: texture_3d<f32>) -> vec4<f32> {
  if (gsl_colorIdx == 1u) {
    var gsl_simpleValue = vec3<f32>(0.0032853691, 0.003285369, 0.003285368);
    if (gsl_simpleDetail > 0.0) {
      let color_black = vec3<f32>(0.0, 0.0, 0.0);
      let color_white = vec3<f32>(1.0, 1.0, 1.0);
      let gsl_tmp_1 = p / 1.0;
      let gsl_tmp_2 = playful_glitterFlakes(gsl_cellularScan(gsl_tmp_1, 1, noiseTex).f1 * 1.0, 1.0, 0.25, noiseTex);
      gsl_simpleValue = mix(gsl_simpleValue, color_mixOklab(color_black, color_white, gsl_tmp_2, noiseTex), gsl_simpleDetail);
    }
    let glitterColor = gsl_simpleValue;
    return vec4<f32>(glitterColor, 1.0);
  }
  if (gsl_colorIdx == 2u) {
    var gsl_simpleValue_2 = vec3<f32>(0.38857287, 0.38857284, 0.3885728);
    if (gsl_simpleDetail > 0.0) {
      let color_white = vec3<f32>(1.0, 1.0, 1.0);
      let color_black = vec3<f32>(0.0, 0.0, 0.0);
      let gsl_tmp_10 = texture_squareWave(p.x, 20.0, noiseTex);
      let gsl_tmp_11 = texture_squareWave(p.y, 20.0, noiseTex);
      let gsl_tmp_13 = gsl_tmp_10 + gsl_tmp_11 - 2.0 * gsl_tmp_10 * gsl_tmp_11;
      let gsl_tmp_12 = texture_squareWave(p.z, 20.0, noiseTex);
      gsl_simpleValue_2 = mix(gsl_simpleValue_2, color_mixOklab(color_white, color_black, gsl_tmp_13 + gsl_tmp_12 - 2.0 * gsl_tmp_13 * gsl_tmp_12, noiseTex), gsl_simpleDetail);
    }
    return vec4<f32>(gsl_simpleValue_2, 1.0);
  }
  if (gsl_colorIdx == 3u) {
    var gsl_simpleValue_3 = vec3<f32>(0.8056361, 0.8831174, 0.8809763);
    if (gsl_simpleDetail > 0.0) {
      let color_white = vec3<f32>(1.0, 1.0, 1.0);
      let color_teal = vec3<f32>(0.0, 0.5, 0.5);
      gsl_simpleValue_3 = mix(gsl_simpleValue_3, color_mixOklab(color_white, color_teal, texture_fadeTo(texture_coverageLine(noise_cellularEdgesZBlend(p, 5.0, 2.5, noiseTex), 1.0, noiseTex), 0.228, 5.0, noiseTex), noiseTex), gsl_simpleDetail);
    }
    let skinColor = gsl_simpleValue_3;
    return vec4<f32>(skinColor, 1.0);
  }
  return vec4<f32>(0.0, 0.0, 0.0, 0.0);
}
fn gsl_materialField(p: vec3<f32>, time: f32, gsl_materialIdx: u32, noiseTex: texture_3d<f32>) -> gsl_material {
  if (gsl_materialIdx == 1u) {
    let material_polishedMetal = mat4x4<f32>(vec4<f32>(0.15, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    var gsl_simpleValue_4 = mat4x4<f32>(material_polishedMetal[0] + (vec4<f32>(0.08, 1.0, 1.5, 0.0) - material_polishedMetal[0]) * 0.06335398, material_polishedMetal[1] + (vec4<f32>(0.0, 0.0, 0.0, 0.0) - material_polishedMetal[1]) * 0.06335398, material_polishedMetal[2] + (vec4<f32>(0.0, 0.0, 0.0, 0.0) - material_polishedMetal[2]) * 0.06335398, material_polishedMetal[3] + (vec4<f32>(0.0, 0.0, 0.0, 0.0) - material_polishedMetal[3]) * 0.06335398);
    if (gsl_simpleDetail > 0.0) {
      let gsl_tmp_1 = p / 1.0;
      let gsl_tmp_2 = playful_glitterFlakes(gsl_cellularScan(gsl_tmp_1, 1, noiseTex).f1 * 1.0, 1.0, 0.25, noiseTex);
      let gsl_simpleFull = material_blend(material_polishedMetal, mat4x4<f32>(vec4<f32>(0.08, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_tmp_2, noiseTex);
      gsl_simpleValue_4 = mat4x4<f32>(mix(gsl_simpleValue_4[0], gsl_simpleFull[0], gsl_simpleDetail), mix(gsl_simpleValue_4[1], gsl_simpleFull[1], gsl_simpleDetail), mix(gsl_simpleValue_4[2], gsl_simpleFull[2], gsl_simpleDetail), mix(gsl_simpleValue_4[3], gsl_simpleFull[3], gsl_simpleDetail));
    }
    let glitterMat = gsl_simpleValue_4;
    let gsl_m = glitterMat; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 2u) {
    let material_matte = mat4x4<f32>(vec4<f32>(1.0, 0.0, 1.45, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    let gsl_m = material_matte; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 3u) {
    let material_matte = mat4x4<f32>(vec4<f32>(1.0, 0.0, 1.45, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    let gsl_m = material_matte; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
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

var<private> gsl_simpleDetail: f32 = 1.0;

const gsl_foldK = array<vec2<f32>, 9>(vec2<f32>(-1.0, -1.0), vec2<f32>(0.0, -1.0), vec2<f32>(1.0, -1.0), vec2<f32>(-1.0, 0.0), vec2<f32>(0.0, 0.0), vec2<f32>(1.0, 0.0), vec2<f32>(-1.0, 1.0), vec2<f32>(0.0, 1.0), vec2<f32>(1.0, 1.0));

struct gsl_struct {
  v_1 : f32,
  v_2 : f32,
}
