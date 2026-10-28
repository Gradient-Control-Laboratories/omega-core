fn ease_cubic(t: f32, noiseTex: texture_3d<f32>) -> f32 {
  return t * t * t * (t * (t * 6.0 - 15.0) + 10.0);
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

fn ease_quadratic(t: f32, noiseTex: texture_3d<f32>) -> f32 {
  return t * t * (3.0 - 2.0 * t);
}

fn texture_mask(field: f32, edge: f32, width: f32, noiseTex: texture_3d<f32>) -> f32 {
  let t = clamp((field - edge) / width * 0.5 + 0.5, 0.0, 1.0);
  return ease_quadratic(t, noiseTex);
}

fn texture_fadeDetail(size: f32, noiseTex: texture_3d<f32>) -> f32 {
  return 1.0 - texture_mask(gsl_pw, 0.375 * size, 0.125 * size, noiseTex);
}

fn material_make(rough: f32, metal: f32, index: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  return mat4x4<f32>(vec4<f32>(rough, metal, index, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
}

fn material_withDisplacement(m: mat4x4<f32>, height: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  let volume = m[2];
  return mat4x4<f32>(m[0], m[1], vec4<f32>(volume.x, volume.y, height, volume.w), m[3]);
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

fn builtin_intersectSharp(fields: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  return max(max(fields.x, fields.y), fields.z);
}

fn builtin_intersectRound(fields: vec3<f32>, radius: f32, noiseTex: texture_3d<f32>) -> f32 {
  let sumOfSquares = 0.0 + max(fields.x + radius, 0.0) * max(fields.x + radius, 0.0) + max(fields.y + radius, 0.0) * max(fields.y + radius, 0.0) + max(fields.z + radius, 0.0) * max(fields.z + radius, 0.0);
  return min(-radius, builtin_intersectSharp(fields, noiseTex)) + sqrt(sumOfSquares);
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

fn sdShape(p: vec3<f32>, time: f32, gsl_param_10_13: f32, gsl_param_13_13: f32, noiseTex: texture_3d<f32>) -> gsl_shape {
  gsl_partNear = vec3<f32>(1e10, 1e10, 1e10);
  let gsl_partLo = vec3<f32>(-12.0, -12.0, -4.0);
  let gsl_partHi = vec3<f32>(12.0, 12.0, 4.0);
  let gsl_partBox = gsl_boxDistance(p, gsl_partLo, gsl_partHi);
  var gsl_part = gsl_partBox;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox <= gsl_vanishView.w) {
    let gsl_partVanish = gsl_vanish(gsl_partLo, gsl_partHi);
    gsl_part = 1e10;
    if (gsl_partVanish.x > 0.0) { gsl_part = builtin_gcl_closure_boxCentered(p - vec3<f32>(0.0, -0.0, 0.0), vec3<f32>(0.0, 0.0, 0.0) * 0.5, abs(vec3<f32>(24.0, 24.0, 8.0)), noiseTex) - 0.0 + gsl_partVanish.y; }
    if (abs(gsl_part) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part), gsl_partVanish.zw); }
  }
  let sand = gsl_part;
  let gsl_s1 = gsl_shape(sand, 1u);
  return gsl_s1;
}

fn gsl_colorField(p: vec3<f32>, time: f32, gsl_colorIdx: u32, gsl_param_10_13: f32, gsl_param_13_13: f32, noiseTex: texture_3d<f32>) -> vec4<f32> {
  if (gsl_colorIdx == 1u) {
    let windSpeed = gsl_param_10_13;
    let gsl_tmp_4 = (vec3<f32>(0.7029593, 0.013074556, 0.07086855) + vec3<f32>(0.12094366, -0.0037923725, -0.004386356) * 0.0375 * clamp((windSpeed - 5.0) / 1.0 * 0.5 + 0.5, 0.0, 1.0) * clamp((windSpeed - 5.0) / 1.0 * 0.5 + 0.5, 0.0, 1.0) * (3.0 - 2.0 * clamp((windSpeed - 5.0) / 1.0 * 0.5 + 0.5, 0.0, 1.0))).x + 0.39633778 * (vec3<f32>(0.7029593, 0.013074556, 0.07086855) + vec3<f32>(0.12094366, -0.0037923725, -0.004386356) * 0.0375 * clamp((windSpeed - 5.0) / 1.0 * 0.5 + 0.5, 0.0, 1.0) * clamp((windSpeed - 5.0) / 1.0 * 0.5 + 0.5, 0.0, 1.0) * (3.0 - 2.0 * clamp((windSpeed - 5.0) / 1.0 * 0.5 + 0.5, 0.0, 1.0))).y + 0.21580376 * (vec3<f32>(0.7029593, 0.013074556, 0.07086855) + vec3<f32>(0.12094366, -0.0037923725, -0.004386356) * 0.0375 * clamp((windSpeed - 5.0) / 1.0 * 0.5 + 0.5, 0.0, 1.0) * clamp((windSpeed - 5.0) / 1.0 * 0.5 + 0.5, 0.0, 1.0) * (3.0 - 2.0 * clamp((windSpeed - 5.0) / 1.0 * 0.5 + 0.5, 0.0, 1.0))).z;
    let gsl_tmp_5 = (vec3<f32>(0.7029593, 0.013074556, 0.07086855) + vec3<f32>(0.12094366, -0.0037923725, -0.004386356) * 0.0375 * clamp((windSpeed - 5.0) / 1.0 * 0.5 + 0.5, 0.0, 1.0) * clamp((windSpeed - 5.0) / 1.0 * 0.5 + 0.5, 0.0, 1.0) * (3.0 - 2.0 * clamp((windSpeed - 5.0) / 1.0 * 0.5 + 0.5, 0.0, 1.0))).x - 0.105561346 * (vec3<f32>(0.7029593, 0.013074556, 0.07086855) + vec3<f32>(0.12094366, -0.0037923725, -0.004386356) * 0.0375 * clamp((windSpeed - 5.0) / 1.0 * 0.5 + 0.5, 0.0, 1.0) * clamp((windSpeed - 5.0) / 1.0 * 0.5 + 0.5, 0.0, 1.0) * (3.0 - 2.0 * clamp((windSpeed - 5.0) / 1.0 * 0.5 + 0.5, 0.0, 1.0))).y - 0.06385417 * (vec3<f32>(0.7029593, 0.013074556, 0.07086855) + vec3<f32>(0.12094366, -0.0037923725, -0.004386356) * 0.0375 * clamp((windSpeed - 5.0) / 1.0 * 0.5 + 0.5, 0.0, 1.0) * clamp((windSpeed - 5.0) / 1.0 * 0.5 + 0.5, 0.0, 1.0) * (3.0 - 2.0 * clamp((windSpeed - 5.0) / 1.0 * 0.5 + 0.5, 0.0, 1.0))).z;
    let gsl_tmp_6 = (vec3<f32>(0.7029593, 0.013074556, 0.07086855) + vec3<f32>(0.12094366, -0.0037923725, -0.004386356) * 0.0375 * clamp((windSpeed - 5.0) / 1.0 * 0.5 + 0.5, 0.0, 1.0) * clamp((windSpeed - 5.0) / 1.0 * 0.5 + 0.5, 0.0, 1.0) * (3.0 - 2.0 * clamp((windSpeed - 5.0) / 1.0 * 0.5 + 0.5, 0.0, 1.0))).x - 0.08948418 * (vec3<f32>(0.7029593, 0.013074556, 0.07086855) + vec3<f32>(0.12094366, -0.0037923725, -0.004386356) * 0.0375 * clamp((windSpeed - 5.0) / 1.0 * 0.5 + 0.5, 0.0, 1.0) * clamp((windSpeed - 5.0) / 1.0 * 0.5 + 0.5, 0.0, 1.0) * (3.0 - 2.0 * clamp((windSpeed - 5.0) / 1.0 * 0.5 + 0.5, 0.0, 1.0))).y - 1.2914855 * (vec3<f32>(0.7029593, 0.013074556, 0.07086855) + vec3<f32>(0.12094366, -0.0037923725, -0.004386356) * 0.0375 * clamp((windSpeed - 5.0) / 1.0 * 0.5 + 0.5, 0.0, 1.0) * clamp((windSpeed - 5.0) / 1.0 * 0.5 + 0.5, 0.0, 1.0) * (3.0 - 2.0 * clamp((windSpeed - 5.0) / 1.0 * 0.5 + 0.5, 0.0, 1.0))).z;
    let gsl_tmp_7 = vec3<f32>(gsl_tmp_4 * gsl_tmp_4 * gsl_tmp_4, gsl_tmp_5 * gsl_tmp_5 * gsl_tmp_5, gsl_tmp_6 * gsl_tmp_6 * gsl_tmp_6);
    var gsl_simpleValue = color_gammaEncode(vec3<f32>(dot(vec3<f32>(4.0767417, -3.3077116, 0.23096994), gsl_tmp_7), dot(vec3<f32>(-1.268438, 2.6097574, -0.34131938), gsl_tmp_7), dot(vec3<f32>(-0.0041960864, -0.7034186, 1.7076147), gsl_tmp_7)), noiseTex);
    if (gsl_simpleDetail > 0.0) {
      let windAngle = gsl_param_13_13;
      let wind = vec3<f32>(cos(windAngle * 3.1415927 / 180.0), sin(windAngle * 3.1415927 / 180.0), 0.0);
      let gsl_tmp_1 = p - normalize(wind) * 0.021 * pow(windSpeed / 8.0, 3.0) * clamp((windSpeed - 5.0) / 1.0 * 0.5 + 0.5, 0.0, 1.0) * clamp((windSpeed - 5.0) / 1.0 * 0.5 + 0.5, 0.0, 1.0) * (3.0 - 2.0 * clamp((windSpeed - 5.0) / 1.0 * 0.5 + 0.5, 0.0, 1.0)) * time;
      let gsl_tmp_2 = fract(dot(gsl_tmp_1, normalize(wind)) / ((0.06 + 0.0075 * windSpeed) * 60.0) + 0.35 * noise_gradientLattice(gsl_tmp_1 / (3.0 * (0.06 + 0.0075 * windSpeed) * 60.0) + vec3<f32>(5.0, 23.0, 41.0), noiseTex));
      let gsl_tmp_3 = select(cos(1.5707963 * (gsl_tmp_2 - 0.7) / 0.3), sin(1.5707963 * gsl_tmp_2 / 0.7), gsl_tmp_2 < 0.7);
      gsl_simpleValue = mix(gsl_simpleValue, color_mixOklab(color_mixOklab(vec3<f32>(0.76, 0.64, 0.44), vec3<f32>(0.3, 0.26, 0.22), 0.5 * (1.0 - gsl_tmp_3) * (1.0 - gsl_tmp_3) * (1.0 - gsl_tmp_3), noiseTex), vec3<f32>(0.86, 0.76, 0.58), 0.25 * clamp((windSpeed - 5.0) / 1.0 * 0.5 + 0.5, 0.0, 1.0) * clamp((windSpeed - 5.0) / 1.0 * 0.5 + 0.5, 0.0, 1.0) * (3.0 - 2.0 * clamp((windSpeed - 5.0) / 1.0 * 0.5 + 0.5, 0.0, 1.0)) * texture_mask(noise_gradientLattice((p - normalize(wind) * windSpeed * 60.0 * time) / vec3<f32>(0.25 * (0.06 + 0.0075 * windSpeed) * 60.0, 0.25 * (0.06 + 0.0075 * windSpeed) * 60.0, 0.25 * (0.06 + 0.0075 * windSpeed) * 60.0) + vec3<f32>(3.0, 7.0, 11.0), noiseTex), 0.3, 0.2, noiseTex) * texture_fadeDetail(0.25 * (0.06 + 0.0075 * windSpeed) * 60.0, noiseTex), noiseTex), gsl_simpleDetail);
    }
    let sandColor = gsl_simpleValue;
    return vec4<f32>(sandColor, 1.0);
  }
  return vec4<f32>(0.0, 0.0, 0.0, 0.0);
}
fn gsl_materialField(p: vec3<f32>, time: f32, gsl_materialIdx: u32, gsl_param_10_13: f32, gsl_param_13_13: f32, noiseTex: texture_3d<f32>) -> gsl_material {
  if (gsl_materialIdx == 1u) {
    let windSpeed = gsl_param_10_13;
    var gsl_simpleValue_2 = mat4x4<f32>(vec4<f32>(0.91341454, 0.0, 1.54, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, (0.06 + 0.0075 * windSpeed) * 60.0 / 17.0 * 0.637, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    if (gsl_simpleDetail > 0.0) {
      let windAngle = gsl_param_13_13;
      let wind = vec3<f32>(cos(windAngle * 3.1415927 / 180.0), sin(windAngle * 3.1415927 / 180.0), 0.0);
      let gsl_tmp_1 = p - normalize(wind) * 0.021 * pow(windSpeed / 8.0, 3.0) * clamp((windSpeed - 5.0) / 1.0 * 0.5 + 0.5, 0.0, 1.0) * clamp((windSpeed - 5.0) / 1.0 * 0.5 + 0.5, 0.0, 1.0) * (3.0 - 2.0 * clamp((windSpeed - 5.0) / 1.0 * 0.5 + 0.5, 0.0, 1.0)) * time;
      let gsl_tmp_2 = fract(dot(gsl_tmp_1, normalize(wind)) / ((0.06 + 0.0075 * windSpeed) * 60.0) + 0.35 * noise_gradientLattice(gsl_tmp_1 / (3.0 * (0.06 + 0.0075 * windSpeed) * 60.0) + vec3<f32>(5.0, 23.0, 41.0), noiseTex));
      let gsl_tmp_3 = select(cos(1.5707963 * (gsl_tmp_2 - 0.7) / 0.3), sin(1.5707963 * gsl_tmp_2 / 0.7), gsl_tmp_2 < 0.7);
      let gsl_simpleFull = material_withDisplacement(material_make(sqrt(sqrt(0.6561 + (1.0 - texture_fadeDetail((0.06 + 0.0075 * windSpeed) * 60.0, noiseTex) * gsl_dispDrawn) * 0.04)), 0.0, 1.54, noiseTex), (0.06 + 0.0075 * windSpeed) * 60.0 / 17.0 * gsl_tmp_3 * texture_fadeDetail((0.06 + 0.0075 * windSpeed) * 60.0, noiseTex), noiseTex);
      gsl_simpleValue_2 = mat4x4<f32>(mix(gsl_simpleValue_2[0], gsl_simpleFull[0], gsl_simpleDetail), mix(gsl_simpleValue_2[1], gsl_simpleFull[1], gsl_simpleDetail), mix(gsl_simpleValue_2[2], gsl_simpleFull[2], gsl_simpleDetail), mix(gsl_simpleValue_2[3], gsl_simpleFull[3], gsl_simpleDetail));
    }
    let sandMat = gsl_simpleValue_2;
    let gsl_m = sandMat; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
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

var<private> gsl_simpleDetail: f32 = 1.0;

var<private> gsl_dispDrawn: f32 = 0.0;

const gsl_foldK = array<vec3<f32>, 8>(vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 0.0, 0.0), vec3<f32>(0.0, 1.0, 0.0), vec3<f32>(1.0, 1.0, 0.0), vec3<f32>(0.0, 0.0, 1.0), vec3<f32>(1.0, 0.0, 1.0), vec3<f32>(0.0, 1.0, 1.0), vec3<f32>(1.0, 1.0, 1.0));
