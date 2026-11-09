fn texture_squareWaveIntegral(x: f32, period: f32) -> f32 {
  let whole = floor(x / period);
  return whole * 0.5 * period + max(x - whole * period - 0.5 * period, 0.0);
}

fn texture_squareWave(x: f32, period: f32) -> f32 {
  let exact = select(1.0, 0.0, fract(x / period) < 0.5);
  let width = gsl_pw;
  let averaged = (texture_squareWaveIntegral(x + 0.5 * width, period) - texture_squareWaveIntegral(x - 0.5 * width, period)) / max(width, 0.000001);
  return select(exact, averaged, width > 0.0001 * period);
}

fn color_srgbToLinearChannel(s: f32) -> f32 {
  return select(pow((s + 0.055) / 1.055, 2.4), s / 12.92, s <= 0.04045);
}

fn color_linearize(c: vec3<f32>) -> vec3<f32> {
  return vec3<f32>(color_srgbToLinearChannel(c.x), color_srgbToLinearChannel(c.y), color_srgbToLinearChannel(c.z));
}

fn color_cbrt(x: f32) -> f32 {
  return sign(x) * pow(abs(x), 0.33333334);
}

fn color_rgb2oklab(c: vec3<f32>) -> vec3<f32> {
  let lin = color_linearize(c);
  let l = dot(vec3<f32>(0.41222146, 0.53633255, 0.051445995), lin);
  let l_ = color_cbrt(l);
  let m = dot(vec3<f32>(0.2119035, 0.6806995, 0.10739696), lin);
  let m_ = color_cbrt(m);
  let s = dot(vec3<f32>(0.08830246, 0.28171885, 0.6299787), lin);
  let s_ = color_cbrt(s);
  let lms = vec3<f32>(l_, m_, s_);
  return vec3<f32>(dot(vec3<f32>(0.21045426, 0.7936178, -0.004072047), lms), dot(vec3<f32>(1.9779985, -2.4285922, 0.4505937), lms), dot(vec3<f32>(0.025904037, 0.78277177, -0.80867577), lms));
}

fn color_linearToSrgbChannel(l: f32) -> f32 {
  return select(1.055 * pow(l, 0.41666666) - 0.055, l * 12.92, l <= 0.0031308);
}

fn color_gammaEncode(c: vec3<f32>) -> vec3<f32> {
  return vec3<f32>(color_linearToSrgbChannel(c.x), color_linearToSrgbChannel(c.y), color_linearToSrgbChannel(c.z));
}

fn color_oklab2rgb(c: vec3<f32>) -> vec3<f32> {
  let l_ = c.x + 0.39633778 * c.y + 0.21580376 * c.z;
  let l = l_ * l_ * l_;
  let m_ = c.x - 0.105561346 * c.y - 0.06385417 * c.z;
  let m = m_ * m_ * m_;
  let s_ = c.x - 0.08948418 * c.y - 1.2914855 * c.z;
  let s = s_ * s_ * s_;
  let lms = vec3<f32>(l, m, s);
  let lin = vec3<f32>(dot(vec3<f32>(4.0767417, -3.3077116, 0.23096994), lms), dot(vec3<f32>(-1.268438, 2.6097574, -0.34131938), lms), dot(vec3<f32>(-0.0041960864, -0.7034186, 1.7076147), lms));
  return color_gammaEncode(lin);
}

fn color_mixOklab(a: vec3<f32>, b: vec3<f32>, t: f32) -> vec3<f32> {
  let la = color_rgb2oklab(a);
  let lb = color_rgb2oklab(b);
  return color_oklab2rgb(la + (lb - la) * t);
}

fn zzf_gsl_anon(time: f32, acc: f32, i: f32) -> f32 {
  return acc + sin(time + i);
}

fn zzf(time: f32, q: vec3<f32>) -> vec3<f32> {
  var gsl_fold = 0.0;
  for (var gsl_i = 0; gsl_i < 8; gsl_i = gsl_i + 1) {
    gsl_fold = zzf_gsl_anon(time, gsl_fold, f32(gsl_i));
  }
  let r = gsl_fold;
  return color_linearize(q - vec3<f32>(r, 0.0, 0.0));
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  gsl_partNear = vec3<f32>(1e10, 1e10, 1e10);
  let gsl_partLo = vec3<f32>(-10.0, -10.0, -10.0);
  let gsl_partHi = vec3<f32>(10.0, 10.0, 10.0);
  let gsl_partBox = gsl_boxDistance(p, gsl_partLo, gsl_partHi);
  var gsl_part = gsl_partBox;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox <= gsl_vanishView.w) {
    let gsl_partVanish = gsl_vanish(gsl_partLo, gsl_partHi);
    gsl_part = 1e10;
    if (gsl_partVanish.x > 0.0) { gsl_part = length(p) - 10.0 + gsl_partVanish.y; }
    if (abs(gsl_part) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part), gsl_partVanish.zw); }
  }
  let b = gsl_part;
  let gsl_partLo_2 = vec3<f32>(-25.0, -25.0, -25.0);
  let gsl_partHi_2 = vec3<f32>(25.0, 25.0, 25.0);
  let gsl_partBox_2 = gsl_boxDistance(p, gsl_partLo_2, gsl_partHi_2);
  var gsl_part_2 = gsl_partBox_2;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_2 <= gsl_vanishView.w) {
    let gsl_partVanish_2 = gsl_vanish(gsl_partLo_2, gsl_partHi_2);
    gsl_part_2 = 1e10;
    if (gsl_partVanish_2.x > 0.0) { gsl_part_2 = length(p) - 25.0 + gsl_partVanish_2.y; }
    if (abs(gsl_part_2) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_2), gsl_partVanish_2.zw); }
  }
  let ball = gsl_part_2;
  let gsl_s1 = gsl_shape(b, 1u);
  let gsl_s2 = gsl_shape(ball, 2u);
  return gsl_min(gsl_s1, gsl_s2);
}

fn gsl_colorField(p: vec3<f32>, time: f32, gsl_colorIdx: u32) -> vec4<f32> {
  if (gsl_colorIdx == 1u) { return vec4<f32>(zzf(time, p), 1.0); }
  if (gsl_colorIdx == 2u) {
    var gsl_simpleValue = vec3<f32>(0.38857287, 0.38857284, 0.3885728);
    if (gsl_simpleDetail > 0.0) {
      let color_white = vec3<f32>(1.0, 1.0, 1.0);
      let color_black = vec3<f32>(0.0, 0.0, 0.0);
      let gsl_tmp_5 = texture_squareWave(p.x, 20.0);
      let gsl_tmp_6 = texture_squareWave(p.y, 20.0);
      let gsl_tmp_8 = gsl_tmp_5 + gsl_tmp_6 - 2.0 * gsl_tmp_5 * gsl_tmp_6;
      let gsl_tmp_7 = texture_squareWave(p.z, 20.0);
      gsl_simpleValue = mix(gsl_simpleValue, color_mixOklab(color_white, color_black, gsl_tmp_8 + gsl_tmp_7 - 2.0 * gsl_tmp_8 * gsl_tmp_7), gsl_simpleDetail);
    }
    return vec4<f32>(gsl_simpleValue, 1.0);
  }
  return vec4<f32>(0.0, 0.0, 0.0, 0.0);
}
fn gsl_materialField(p: vec3<f32>, time: f32, gsl_materialIdx: u32) -> gsl_material {
  if (gsl_materialIdx == 2u) {
    let material_matte = mat4x4<f32>(vec4<f32>(1.0, 0.0, 1.45, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    let gsl_m = material_matte; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
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
