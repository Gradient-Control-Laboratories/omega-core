fn drill_hexXY(p: vec3<f32>, acrossFlats: f32) -> f32 {
  let h = acrossFlats / 2.0;
  return max(max(abs(p.x) - h, abs(0.5 * p.x + 0.8660254 * p.y) - h), abs(0.5 * p.x - 0.8660254 * p.y) - h);
}

fn drill_prismVoid(p: vec3<f32>, zTop: f32, section: f32, depth: f32) -> f32 {
  return max(section, zTop - depth - p.z);
}

fn gsl_remap(p: vec3<f32>) -> f32 {
  return max(min(max(max(max(length(p.xy) - 4.25, -p.z), p.z - 5.0), (length(p.xy) - 4.25 + p.z - 5.0 + 0.2) * 0.7071), max(max(max(length(p.xy) - 2.5, -16.0 - p.z), p.z - 2.5), (length(p.xy) - 2.5 + -p.z - 16.0 + 0.75) * 0.7071)), -drill_prismVoid(p, 5.0, drill_hexXY(p, 4.0), 2.5));
}

fn gsl_remap_2(p: vec3<f32>) -> f32 {
  return min(max(max(length(p.xy) - 4.25, -p.z), p.z - 5.0), max(max(length(p.xy) - 2.5, -16.0 - p.z), p.z - 2.5));
}

fn catalog_onRow(key: f32, rowKey: f32, tol: f32) -> f32 {
  return 1.0 - clamp(sign(abs(key - rowKey) - tol), 0.0, 1.0);
}

fn catalog_nutAcrossFlats(nominal: f32) -> f32 {
  let k = nominal;
  return 0.0 * catalog_onRow(k, 1.0, 0.0001) + 0.0 * catalog_onRow(k, 1.2, 0.0001) + 0.0 * catalog_onRow(k, 1.4, 0.0001) + 3.2 * catalog_onRow(k, 1.6, 0.0001) + 0.0 * catalog_onRow(k, 1.8, 0.0001) + 4.0 * catalog_onRow(k, 2.0, 0.0001) + 0.0 * catalog_onRow(k, 2.2, 0.0001) + 5.0 * catalog_onRow(k, 2.5, 0.0001) + 5.5 * catalog_onRow(k, 3.0, 0.0001) + 6.0 * catalog_onRow(k, 3.5, 0.0001) + 7.0 * catalog_onRow(k, 4.0, 0.0001) + 0.0 * catalog_onRow(k, 4.5, 0.0001) + 8.0 * catalog_onRow(k, 5.0, 0.0001) + 10.0 * catalog_onRow(k, 6.0, 0.0001) + 11.0 * catalog_onRow(k, 7.0, 0.0001) + 13.0 * catalog_onRow(k, 8.0, 0.0001) + 16.0 * catalog_onRow(k, 10.0, 0.0001) + 18.0 * catalog_onRow(k, 12.0, 0.0001) + 21.0 * catalog_onRow(k, 14.0, 0.0001) + 24.0 * catalog_onRow(k, 16.0, 0.0001) + 27.0 * catalog_onRow(k, 18.0, 0.0001) + 30.0 * catalog_onRow(k, 20.0, 0.0001) + 34.0 * catalog_onRow(k, 22.0, 0.0001) + 36.0 * catalog_onRow(k, 24.0, 0.0001) + 41.0 * catalog_onRow(k, 27.0, 0.0001) + 46.0 * catalog_onRow(k, 30.0, 0.0001) + 50.0 * catalog_onRow(k, 33.0, 0.0001) + 55.0 * catalog_onRow(k, 36.0, 0.0001) + 60.0 * catalog_onRow(k, 39.0, 0.0001) + 65.0 * catalog_onRow(k, 42.0, 0.0001) + 70.0 * catalog_onRow(k, 45.0, 0.0001) + 75.0 * catalog_onRow(k, 48.0, 0.0001) + 80.0 * catalog_onRow(k, 52.0, 0.0001) + 85.0 * catalog_onRow(k, 56.0, 0.0001) + 90.0 * catalog_onRow(k, 60.0, 0.0001) + 95.0 * catalog_onRow(k, 64.0, 0.0001);
}

fn catalog_nutHeight(nominal: f32) -> f32 {
  let k = nominal;
  return 0.0 * catalog_onRow(k, 1.0, 0.0001) + 0.0 * catalog_onRow(k, 1.2, 0.0001) + 0.0 * catalog_onRow(k, 1.4, 0.0001) + 1.3 * catalog_onRow(k, 1.6, 0.0001) + 0.0 * catalog_onRow(k, 1.8, 0.0001) + 1.6 * catalog_onRow(k, 2.0, 0.0001) + 0.0 * catalog_onRow(k, 2.2, 0.0001) + 2.0 * catalog_onRow(k, 2.5, 0.0001) + 2.4 * catalog_onRow(k, 3.0, 0.0001) + 2.8 * catalog_onRow(k, 3.5, 0.0001) + 3.2 * catalog_onRow(k, 4.0, 0.0001) + 0.0 * catalog_onRow(k, 4.5, 0.0001) + 4.7 * catalog_onRow(k, 5.0, 0.0001) + 5.2 * catalog_onRow(k, 6.0, 0.0001) + 6.5 * catalog_onRow(k, 7.0, 0.0001) + 6.8 * catalog_onRow(k, 8.0, 0.0001) + 8.4 * catalog_onRow(k, 10.0, 0.0001) + 10.8 * catalog_onRow(k, 12.0, 0.0001) + 12.8 * catalog_onRow(k, 14.0, 0.0001) + 14.8 * catalog_onRow(k, 16.0, 0.0001) + 15.8 * catalog_onRow(k, 18.0, 0.0001) + 18.0 * catalog_onRow(k, 20.0, 0.0001) + 19.4 * catalog_onRow(k, 22.0, 0.0001) + 21.5 * catalog_onRow(k, 24.0, 0.0001) + 23.8 * catalog_onRow(k, 27.0, 0.0001) + 25.6 * catalog_onRow(k, 30.0, 0.0001) + 28.7 * catalog_onRow(k, 33.0, 0.0001) + 31.0 * catalog_onRow(k, 36.0, 0.0001) + 33.4 * catalog_onRow(k, 39.0, 0.0001) + 34.0 * catalog_onRow(k, 42.0, 0.0001) + 36.0 * catalog_onRow(k, 45.0, 0.0001) + 38.0 * catalog_onRow(k, 48.0, 0.0001) + 42.0 * catalog_onRow(k, 52.0, 0.0001) + 45.0 * catalog_onRow(k, 56.0, 0.0001) + 48.0 * catalog_onRow(k, 60.0, 0.0001) + 51.0 * catalog_onRow(k, 64.0, 0.0001);
}

fn catalog_pitch(nominal: f32) -> f32 {
  let k = nominal;
  return 0.25 * catalog_onRow(k, 1.0, 0.0001) + 0.25 * catalog_onRow(k, 1.2, 0.0001) + 0.3 * catalog_onRow(k, 1.4, 0.0001) + 0.35 * catalog_onRow(k, 1.6, 0.0001) + 0.35 * catalog_onRow(k, 1.8, 0.0001) + 0.4 * catalog_onRow(k, 2.0, 0.0001) + 0.45 * catalog_onRow(k, 2.2, 0.0001) + 0.45 * catalog_onRow(k, 2.5, 0.0001) + 0.5 * catalog_onRow(k, 3.0, 0.0001) + 0.6 * catalog_onRow(k, 3.5, 0.0001) + 0.7 * catalog_onRow(k, 4.0, 0.0001) + 0.75 * catalog_onRow(k, 4.5, 0.0001) + 0.8 * catalog_onRow(k, 5.0, 0.0001) + 1.0 * catalog_onRow(k, 6.0, 0.0001) + 1.0 * catalog_onRow(k, 7.0, 0.0001) + 1.25 * catalog_onRow(k, 8.0, 0.0001) + 1.5 * catalog_onRow(k, 10.0, 0.0001) + 1.75 * catalog_onRow(k, 12.0, 0.0001) + 2.0 * catalog_onRow(k, 14.0, 0.0001) + 2.0 * catalog_onRow(k, 16.0, 0.0001) + 2.5 * catalog_onRow(k, 18.0, 0.0001) + 2.5 * catalog_onRow(k, 20.0, 0.0001) + 2.5 * catalog_onRow(k, 22.0, 0.0001) + 3.0 * catalog_onRow(k, 24.0, 0.0001) + 3.0 * catalog_onRow(k, 27.0, 0.0001) + 3.5 * catalog_onRow(k, 30.0, 0.0001) + 3.5 * catalog_onRow(k, 33.0, 0.0001) + 4.0 * catalog_onRow(k, 36.0, 0.0001) + 4.0 * catalog_onRow(k, 39.0, 0.0001) + 4.5 * catalog_onRow(k, 42.0, 0.0001) + 4.5 * catalog_onRow(k, 45.0, 0.0001) + 5.0 * catalog_onRow(k, 48.0, 0.0001) + 5.0 * catalog_onRow(k, 52.0, 0.0001) + 5.5 * catalog_onRow(k, 56.0, 0.0001) + 5.5 * catalog_onRow(k, 60.0, 0.0001) + 6.0 * catalog_onRow(k, 64.0, 0.0001);
}

fn thread_metricPitch(nominal: f32) -> f32 {
  return catalog_pitch(nominal);
}

fn thread_formDepth(pitchVal: f32) -> f32 {
  return 0.5412659 * pitchVal;
}

fn thread_minorDia(major: f32, pitchVal: f32) -> f32 {
  return major - 2.0 * thread_formDepth(pitchVal);
}

fn gsl_remap_3(p: vec3<f32>) -> f32 {
  return max(max(max(drill_hexXY(p, 8.0), -p.z), p.z - 4.7), -(length(p.xy) - 2.0669873));
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

fn thread_paint(base: vec3<f32>, height: f32, pitchVal: f32) -> vec3<f32> {
  var gsl_simpleValue = base;
  if (gsl_simpleDetail > 0.0) {
    let depth = clamp(-height / thread_formDepth(pitchVal), 0.0, 1.0);
    gsl_simpleValue = mix(gsl_simpleValue, color_mixOklab(base, base * 0.35, depth), gsl_simpleDetail);
  }
  let painted = gsl_simpleValue;
  return painted;
}

fn gsl_remap_4(p: vec3<f32>) -> f32 {
  let gslCarry4 = fastener_screw(max(max(max(length(p.xy) - 4.25, -p.z), p.z - 5.0), (length(p.xy) - 4.25 + p.z - 5.0 + 0.2) * 0.7071), max(max(max(length(p.xy) - 2.5, -16.0 - p.z), p.z - 2.5), (length(p.xy) - 2.5 + -p.z - 16.0 + 0.75) * 0.7071), drill_prismVoid(p, 5.0, max(max(abs(p.x) - 2.0, abs(0.5 * p.x + 0.8660254 * p.y) - 2.0), abs(0.5 * p.x - 0.8660254 * p.y) - 2.0), 2.5));
  return gslCarry4;
}

fn gsl_remap_5(p: vec3<f32>) -> f32 {
  let gslCarry1 = min(max(max(length(p.xy) - 4.25, -p.z), p.z - 5.0), max(max(length(p.xy) - 2.5, -16.0 - p.z), p.z - 2.5));
  return gslCarry1;
}

fn gsl_remap_6(p: vec3<f32>) -> f32 {
  let gslCarry10 = fastener_screw(max(max(max(length(p.xy) - 4.25, -p.z), p.z - 5.0), (length(p.xy) - 4.25 + p.z - 5.0 + 0.2) * 0.7071), max(max(max(length(p.xy) - 2.5, -16.0 - p.z), p.z - 2.5), (length(p.xy) - 2.5 + -p.z - 16.0 + 0.75) * 0.7071), drill_prismVoid(p, 5.0, max(max(abs(p.x) - 2.0, abs(0.5 * p.x + 0.8660254 * p.y) - 2.0), abs(0.5 * p.x - 0.8660254 * p.y) - 2.0), 2.5));
  return gslCarry10;
}

fn gsl_remap_7(p: vec3<f32>) -> f32 {
  let gslCarry7 = min(max(max(length(p.xy) - 4.25, -p.z), p.z - 5.0), max(max(length(p.xy) - 2.5, -16.0 - p.z), p.z - 2.5));
  return gslCarry7;
}

fn gsl_remap_8(p: vec3<f32>) -> f32 {
  let gslCarry16 = fastener_nut(p, 5.0);
  return gslCarry16;
}

fn gsl_remap_9(p: vec3<f32>) -> f32 {
  let gslCarry13 = max(max(max(max(max(abs(p.x) - 4.0, abs(0.5 * p.x + 0.8660254 * p.y) - 4.0), abs(0.5 * p.x - 0.8660254 * p.y) - 4.0), -p.z), p.z - 4.7), -(length(p.xy) - 2.0669873));
  return gslCarry13;
}

fn gsl_remap_10(p: vec3<f32>) -> f32 {
  let gslCarry22 = fastener_nut(p, 5.0);
  return gslCarry22;
}

fn gsl_remap_11(p: vec3<f32>) -> f32 {
  let gslCarry19 = max(max(max(max(max(abs(p.x) - 4.0, abs(0.5 * p.x + 0.8660254 * p.y) - 4.0), abs(0.5 * p.x - 0.8660254 * p.y) - 4.0), -p.z), p.z - 4.7), -(length(p.xy) - 2.0669873));
  return gslCarry19;
}

fn thread_reliefColor(base: vec3<f32>, height: f32, pitchVal: f32) -> vec3<f32> {
  var gsl_simpleValue = base;
  if (gsl_simpleDetail > 0.0) {
    let depth = clamp(-height / thread_formDepth(pitchVal), 0.0, 1.0);
    gsl_simpleValue = mix(gsl_simpleValue, color_mixOklab(base, base * 0.35, depth * (1.0 - gsl_dispDrawn)), gsl_simpleDetail);
  }
  let painted = gsl_simpleValue;
  return painted;
}

fn material_withDisplacement(m: mat4x4<f32>, height: f32) -> mat4x4<f32> {
  let volume = m[2];
  return mat4x4<f32>(m[0], m[1], vec4<f32>(volume.x, volume.y, height, volume.w), m[3]);
}

fn gsl_remap_12(p: vec3<f32>) -> f32 {
  let gslCarry28 = fastener_screw(max(max(max(length(p.xy) - 4.25, -p.z), p.z - 5.0), (length(p.xy) - 4.25 + p.z - 5.0 + 0.2) * 0.7071), max(max(max(length(p.xy) - 2.5, -16.0 - p.z), p.z - 2.5), (length(p.xy) - 2.5 + -p.z - 16.0 + 0.75) * 0.7071), drill_prismVoid(p, 5.0, max(max(abs(p.x) - 2.0, abs(0.5 * p.x + 0.8660254 * p.y) - 2.0), abs(0.5 * p.x - 0.8660254 * p.y) - 2.0), 2.5));
  return gslCarry28;
}

fn gsl_remap_13(p: vec3<f32>) -> f32 {
  let gslCarry25 = min(max(max(length(p.xy) - 4.25, -p.z), p.z - 5.0), max(max(length(p.xy) - 2.5, -16.0 - p.z), p.z - 2.5));
  return gslCarry25;
}

fn gsl_remap_14(p: vec3<f32>) -> f32 {
  let gslCarry34 = fastener_screw(max(max(max(length(p.xy) - 4.25, -p.z), p.z - 5.0), (length(p.xy) - 4.25 + p.z - 5.0 + 0.2) * 0.7071), max(max(max(length(p.xy) - 2.5, -16.0 - p.z), p.z - 2.5), (length(p.xy) - 2.5 + -p.z - 16.0 + 0.75) * 0.7071), drill_prismVoid(p, 5.0, max(max(abs(p.x) - 2.0, abs(0.5 * p.x + 0.8660254 * p.y) - 2.0), abs(0.5 * p.x - 0.8660254 * p.y) - 2.0), 2.5));
  return gslCarry34;
}

fn gsl_remap_15(p: vec3<f32>) -> f32 {
  let gslCarry31 = min(max(max(length(p.xy) - 4.25, -p.z), p.z - 5.0), max(max(length(p.xy) - 2.5, -16.0 - p.z), p.z - 2.5));
  return gslCarry31;
}

fn gsl_remap_16(p: vec3<f32>) -> f32 {
  let gslCarry40 = fastener_nut(p, 5.0);
  return gslCarry40;
}

fn gsl_remap_17(p: vec3<f32>) -> f32 {
  let gslCarry37 = max(max(max(max(max(abs(p.x) - 4.0, abs(0.5 * p.x + 0.8660254 * p.y) - 4.0), abs(0.5 * p.x - 0.8660254 * p.y) - 4.0), -p.z), p.z - 4.7), -(length(p.xy) - 2.0669873));
  return gslCarry37;
}

fn gsl_remap_18(p: vec3<f32>) -> f32 {
  let gslCarry46 = fastener_nut(p, 5.0);
  return gslCarry46;
}

fn gsl_remap_19(p: vec3<f32>) -> f32 {
  let gslCarry43 = max(max(max(max(max(abs(p.x) - 4.0, abs(0.5 * p.x + 0.8660254 * p.y) - 4.0), abs(0.5 * p.x - 0.8660254 * p.y) - 4.0), -p.z), p.z - 4.7), -(length(p.xy) - 2.0669873));
  return gslCarry43;
}

fn gsl_remap_20(p: vec3<f32>) -> f32 {
  let gslCarry52 = fastener_screw(max(max(max(length(p.xy) - 4.25, -p.z), p.z - 5.0), (length(p.xy) - 4.25 + p.z - 5.0 + 0.2) * 0.7071), max(max(max(length(p.xy) - 2.5, -16.0 - p.z), p.z - 2.5), (length(p.xy) - 2.5 + -p.z - 16.0 + 0.75) * 0.7071), drill_prismVoid(p, 5.0, max(max(abs(p.x) - 2.0, abs(0.5 * p.x + 0.8660254 * p.y) - 2.0), abs(0.5 * p.x - 0.8660254 * p.y) - 2.0), 2.5));
  return gslCarry52;
}

fn gsl_remap_21(p: vec3<f32>) -> f32 {
  let gslCarry49 = min(max(max(length(p.xy) - 4.25, -p.z), p.z - 5.0), max(max(length(p.xy) - 2.5, -16.0 - p.z), p.z - 2.5));
  return gslCarry49;
}

fn gsl_remap_22(p: vec3<f32>) -> f32 {
  let gslCarry58 = fastener_screw(max(max(max(length(p.xy) - 4.25, -p.z), p.z - 5.0), (length(p.xy) - 4.25 + p.z - 5.0 + 0.2) * 0.7071), max(max(max(length(p.xy) - 2.5, -16.0 - p.z), p.z - 2.5), (length(p.xy) - 2.5 + -p.z - 16.0 + 0.75) * 0.7071), drill_prismVoid(p, 5.0, max(max(abs(p.x) - 2.0, abs(0.5 * p.x + 0.8660254 * p.y) - 2.0), abs(0.5 * p.x - 0.8660254 * p.y) - 2.0), 2.5));
  return gslCarry58;
}

fn gsl_remap_23(p: vec3<f32>) -> f32 {
  let gslCarry55 = min(max(max(length(p.xy) - 4.25, -p.z), p.z - 5.0), max(max(length(p.xy) - 2.5, -16.0 - p.z), p.z - 2.5));
  return gslCarry55;
}

fn gsl_remap_24(p: vec3<f32>) -> f32 {
  let gslCarry64 = fastener_nut(p, 5.0);
  return gslCarry64;
}

fn gsl_remap_25(p: vec3<f32>) -> f32 {
  let gslCarry61 = max(max(max(max(max(abs(p.x) - 4.0, abs(0.5 * p.x + 0.8660254 * p.y) - 4.0), abs(0.5 * p.x - 0.8660254 * p.y) - 4.0), -p.z), p.z - 4.7), -(length(p.xy) - 2.0669873));
  return gslCarry61;
}

fn gsl_remap_26(p: vec3<f32>) -> f32 {
  let gslCarry70 = fastener_nut(p, 5.0);
  return gslCarry70;
}

fn gsl_remap_27(p: vec3<f32>) -> f32 {
  let gslCarry67 = max(max(max(max(max(abs(p.x) - 4.0, abs(0.5 * p.x + 0.8660254 * p.y) - 4.0), abs(0.5 * p.x - 0.8660254 * p.y) - 4.0), -p.z), p.z - 4.7), -(length(p.xy) - 2.0669873));
  return gslCarry67;
}

fn gsl_remap_28(p: vec3<f32>) -> f32 {
  let gsl_tmp_276 = 6.2831855 * p.z / 0.8;
  let gsl_tmp_275 = max(length(p.xy), 0.2);
  let gslCarry75 = fastener_screw(max(max(max(length(p.xy) - 4.25, -p.z), p.z - 5.0), (length(p.xy) - 4.25 + p.z - 5.0 + 0.2) * 0.7071), max(max(max((length(p.xy) - (2.5 - 0.21650636 * (1.0 - (cos(gsl_tmp_276) * p.x / gsl_tmp_275 + sin(gsl_tmp_276) * p.y / gsl_tmp_275)))) * 0.7, -16.0 - p.z), p.z - 2.5), (length(p.xy) - 2.5 + -p.z - 16.0 + 0.75) * 0.7071), drill_prismVoid(p, 5.0, max(max(abs(p.x) - 2.0, abs(0.5 * p.x + 0.8660254 * p.y) - 2.0), abs(0.5 * p.x - 0.8660254 * p.y) - 2.0), 2.5));
  return gslCarry75;
}

fn gsl_remap_29(p: vec3<f32>) -> f32 {
  let gslCarry73 = min(max(max(length(p.xy) - 4.25, -p.z), p.z - 5.0), max(max(length(p.xy) - 2.5, -16.0 - p.z), p.z - 2.5));
  return gslCarry73;
}

fn gsl_remap_30(p: vec3<f32>) -> f32 {
  let gsl_tmp_292 = 6.2831855 * p.z / 0.8;
  let gsl_tmp_291 = max(length(p.xy), 0.2);
  let gslCarry79 = fastener_screw(max(max(max(length(p.xy) - 4.25, -p.z), p.z - 5.0), (length(p.xy) - 4.25 + p.z - 5.0 + 0.2) * 0.7071), max(max(max((length(p.xy) - (2.5 - 0.21650636 * (1.0 - (cos(gsl_tmp_292) * p.x / gsl_tmp_291 + sin(gsl_tmp_292) * p.y / gsl_tmp_291)))) * 0.7, -16.0 - p.z), p.z - 2.5), (length(p.xy) - 2.5 + -p.z - 16.0 + 0.75) * 0.7071), drill_prismVoid(p, 5.0, max(max(abs(p.x) - 2.0, abs(0.5 * p.x + 0.8660254 * p.y) - 2.0), abs(0.5 * p.x - 0.8660254 * p.y) - 2.0), 2.5));
  return gslCarry79;
}

fn gsl_remap_31(p: vec3<f32>) -> f32 {
  let gslCarry77 = min(max(max(length(p.xy) - 4.25, -p.z), p.z - 5.0), max(max(length(p.xy) - 2.5, -16.0 - p.z), p.z - 2.5));
  return gslCarry77;
}

fn gsl_remap_32(p: vec3<f32>) -> f32 {
  let gsl_tmp_312 = 6.2831855 * p.z / 0.8;
  let gsl_tmp_311 = max(length(p.xy), 0.2);
  let gslCarry83 = max(fastener_nut(p, 5.0), -((length(p.xy) - (2.5 - 0.21650636 * (1.0 - (cos(gsl_tmp_312) * p.x / gsl_tmp_311 + sin(gsl_tmp_312) * p.y / gsl_tmp_311)))) * 0.7));
  return gslCarry83;
}

fn gsl_remap_33(p: vec3<f32>) -> f32 {
  let gslCarry81 = max(max(max(max(max(abs(p.x) - 4.0, abs(0.5 * p.x + 0.8660254 * p.y) - 4.0), abs(0.5 * p.x - 0.8660254 * p.y) - 4.0), -p.z), p.z - 4.7), -(length(p.xy) - 2.0669873));
  return gslCarry81;
}

fn gsl_remap_34(p: vec3<f32>) -> f32 {
  let gsl_tmp_328 = 6.2831855 * p.z / 0.8;
  let gsl_tmp_327 = max(length(p.xy), 0.2);
  let gslCarry87 = max(fastener_nut(p, 5.0), -((length(p.xy) - (2.5 - 0.21650636 * (1.0 - (cos(gsl_tmp_328) * p.x / gsl_tmp_327 + sin(gsl_tmp_328) * p.y / gsl_tmp_327)))) * 0.7));
  return gslCarry87;
}

fn gsl_remap_35(p: vec3<f32>) -> f32 {
  let gslCarry85 = max(max(max(max(max(abs(p.x) - 4.0, abs(0.5 * p.x + 0.8660254 * p.y) - 4.0), abs(0.5 * p.x - 0.8660254 * p.y) - 4.0), -p.z), p.z - 4.7), -(length(p.xy) - 2.0669873));
  return gslCarry85;
}

fn gsl_remap_36(p: vec3<f32>) -> f32 {
  let gsl_tmp_339 = length(p.xy);
  let gsl_tmp_341 = 6.2831855 * p.z / 0.8;
  let gsl_tmp_340 = max(length(p.xy), 0.2);
  let gslCarry91 = fastener_screw(max(max(max(length(p.xy) - 4.25, -p.z), p.z - 5.0), (length(p.xy) - 4.25 + p.z - 5.0 + 0.2) * 0.7071), max(max(max(max(length(p.xy) - 2.5, -max(2.0669873 - gsl_tmp_339, ((1.0 - acos(clamp(cos(gsl_tmp_341) * p.x / gsl_tmp_340 + sin(gsl_tmp_341) * p.y / gsl_tmp_340, -1.0, 1.0)) / 3.1415927) * 0.4 - 0.1 - (gsl_tmp_339 - 2.0669873) * 0.57735026) * 0.86383384)), -16.0 - p.z), p.z - 2.5), (length(p.xy) - 2.5 + -p.z - 16.0 + 0.75) * 0.7071), drill_prismVoid(p, 5.0, max(max(abs(p.x) - 2.0, abs(0.5 * p.x + 0.8660254 * p.y) - 2.0), abs(0.5 * p.x - 0.8660254 * p.y) - 2.0), 2.5));
  return gslCarry91;
}

fn gsl_remap_37(p: vec3<f32>) -> f32 {
  let gslCarry89 = min(max(max(length(p.xy) - 4.25, -p.z), p.z - 5.0), max(max(length(p.xy) - 2.5, -16.0 - p.z), p.z - 2.5));
  return gslCarry89;
}

fn gsl_remap_38(p: vec3<f32>) -> f32 {
  let gsl_tmp_356 = length(p.xy);
  let gsl_tmp_358 = 6.2831855 * p.z / 0.8;
  let gsl_tmp_357 = max(length(p.xy), 0.2);
  let gslCarry95 = fastener_screw(max(max(max(length(p.xy) - 4.25, -p.z), p.z - 5.0), (length(p.xy) - 4.25 + p.z - 5.0 + 0.2) * 0.7071), max(max(max(max(length(p.xy) - 2.5, -max(2.0669873 - gsl_tmp_356, ((1.0 - acos(clamp(cos(gsl_tmp_358) * p.x / gsl_tmp_357 + sin(gsl_tmp_358) * p.y / gsl_tmp_357, -1.0, 1.0)) / 3.1415927) * 0.4 - 0.1 - (gsl_tmp_356 - 2.0669873) * 0.57735026) * 0.86383384)), -16.0 - p.z), p.z - 2.5), (length(p.xy) - 2.5 + -p.z - 16.0 + 0.75) * 0.7071), drill_prismVoid(p, 5.0, max(max(abs(p.x) - 2.0, abs(0.5 * p.x + 0.8660254 * p.y) - 2.0), abs(0.5 * p.x - 0.8660254 * p.y) - 2.0), 2.5));
  return gslCarry95;
}

fn gsl_remap_39(p: vec3<f32>) -> f32 {
  let gslCarry93 = min(max(max(length(p.xy) - 4.25, -p.z), p.z - 5.0), max(max(length(p.xy) - 2.5, -16.0 - p.z), p.z - 2.5));
  return gslCarry93;
}

fn gsl_remap_40(p: vec3<f32>) -> f32 {
  let gsl_tmp_377 = length(p.xy);
  let gsl_tmp_379 = 6.2831855 * p.z / 0.8;
  let gsl_tmp_378 = max(length(p.xy), 0.2);
  let gslCarry99 = max(fastener_nut(p, 5.0), -max(length(p.xy) - 2.5, -max(2.0669873 - gsl_tmp_377, ((1.0 - acos(clamp(cos(gsl_tmp_379) * p.x / gsl_tmp_378 + sin(gsl_tmp_379) * p.y / gsl_tmp_378, -1.0, 1.0)) / 3.1415927) * 0.4 - 0.1 - (gsl_tmp_377 - 2.0669873) * 0.57735026) * 0.86383384)));
  return gslCarry99;
}

fn gsl_remap_41(p: vec3<f32>) -> f32 {
  let gslCarry97 = max(max(max(max(max(abs(p.x) - 4.0, abs(0.5 * p.x + 0.8660254 * p.y) - 4.0), abs(0.5 * p.x - 0.8660254 * p.y) - 4.0), -p.z), p.z - 4.7), -(length(p.xy) - 2.0669873));
  return gslCarry97;
}

fn gsl_remap_42(p: vec3<f32>) -> f32 {
  let gsl_tmp_394 = length(p.xy);
  let gsl_tmp_396 = 6.2831855 * p.z / 0.8;
  let gsl_tmp_395 = max(length(p.xy), 0.2);
  let gslCarry103 = max(fastener_nut(p, 5.0), -max(length(p.xy) - 2.5, -max(2.0669873 - gsl_tmp_394, ((1.0 - acos(clamp(cos(gsl_tmp_396) * p.x / gsl_tmp_395 + sin(gsl_tmp_396) * p.y / gsl_tmp_395, -1.0, 1.0)) / 3.1415927) * 0.4 - 0.1 - (gsl_tmp_394 - 2.0669873) * 0.57735026) * 0.86383384)));
  return gslCarry103;
}

fn gsl_remap_43(p: vec3<f32>) -> f32 {
  let gslCarry101 = max(max(max(max(max(abs(p.x) - 4.0, abs(0.5 * p.x + 0.8660254 * p.y) - 4.0), abs(0.5 * p.x - 0.8660254 * p.y) - 4.0), -p.z), p.z - 4.7), -(length(p.xy) - 2.0669873));
  return gslCarry101;
}

fn gsl_remap_44(p: vec3<f32>) -> vec3<f32> {
  let gsl_tmp_24 = 6.2831855 * p.z / 0.8;
  let gsl_tmp_23 = max(length(p.xy), 0.2);
  let gslCarry5 = thread_paint(vec3<f32>(0.77, 0.78, 0.78), -min(min(clamp((acos(clamp(cos(gsl_tmp_24) * p.x / gsl_tmp_23 + sin(gsl_tmp_24) * p.y / gsl_tmp_23, -1.0, 1.0)) / 3.1415927 * 0.4 - 0.05) * 1.7320508, 0.0, 0.43301272), max(0.0, 0.5 * min(min(0.0 - p.z, p.z - -16.0), 2.9330127 - length(p.xy)))), max(0.0, length(p.xy) - 1.6339746)), 0.8);
  return gslCarry5;
}

fn gsl_remap_45(p: vec3<f32>) -> vec3<f32> {
  return vec3<f32>(0.77, 0.78, 0.78);
}

fn gsl_remap_46(p: vec3<f32>) -> vec3<f32> {
  let gsl_tmp_40 = 6.2831855 * p.z / 0.8;
  let gsl_tmp_39 = max(length(p.xy), 0.2);
  let gslCarry11 = thread_paint(vec3<f32>(0.77, 0.78, 0.78), -min(min(clamp((acos(clamp(cos(gsl_tmp_40) * p.x / gsl_tmp_39 + sin(gsl_tmp_40) * p.y / gsl_tmp_39, -1.0, 1.0)) / 3.1415927 * 0.4 - 0.05) * 1.7320508, 0.0, 0.43301272), max(0.0, 0.5 * min(min(0.0 - p.z, p.z - -16.0), 2.9330127 - length(p.xy)))), max(0.0, length(p.xy) - 1.6339746)), 0.8);
  return gslCarry11;
}

fn gsl_remap_47(p: vec3<f32>) -> vec3<f32> {
  let gsl_tmp_64 = 6.2831855 * p.z / 0.8;
  let gsl_tmp_63 = max(length(p.xy), 0.2);
  let gslCarry17 = thread_paint(vec3<f32>(0.77, 0.78, 0.78), -min(max(0.43301272 - clamp((acos(clamp(cos(gsl_tmp_64) * p.x / gsl_tmp_63 + sin(gsl_tmp_64) * p.y / gsl_tmp_63, -1.0, 1.0)) / 3.1415927 * 0.4 - 0.05) * 1.7320508, 0.0, 0.43301272), 0.0), max(0.0, 0.5 * min(min(4.7 - p.z, p.z - 0.0), 3.3660254 - length(p.xy)))), 0.8);
  return gslCarry17;
}

fn gsl_remap_48(p: vec3<f32>) -> vec3<f32> {
  let gsl_tmp_84 = 6.2831855 * p.z / 0.8;
  let gsl_tmp_83 = max(length(p.xy), 0.2);
  let gslCarry23 = thread_paint(vec3<f32>(0.77, 0.78, 0.78), -min(max(0.43301272 - clamp((acos(clamp(cos(gsl_tmp_84) * p.x / gsl_tmp_83 + sin(gsl_tmp_84) * p.y / gsl_tmp_83, -1.0, 1.0)) / 3.1415927 * 0.4 - 0.05) * 1.7320508, 0.0, 0.43301272), 0.0), max(0.0, 0.5 * min(min(4.7 - p.z, p.z - 0.0), 3.3660254 - length(p.xy)))), 0.8);
  return gslCarry23;
}

fn gsl_remap_49(p: vec3<f32>) -> vec3<f32> {
  let gsl_tmp_98 = 6.2831855 * p.z / 0.8;
  let gsl_tmp_97 = max(length(p.xy), 0.2);
  let gslCarry29 = thread_reliefColor(vec3<f32>(0.77, 0.78, 0.78), -min(min(clamp((acos(clamp(cos(gsl_tmp_98) * p.x / gsl_tmp_97 + sin(gsl_tmp_98) * p.y / gsl_tmp_97, -1.0, 1.0)) / 3.1415927 * 0.4 - 0.05) * 1.7320508, 0.0, 0.43301272), max(0.0, 0.5 * min(min(0.0 - p.z, p.z - -16.0), 2.9330127 - length(p.xy)))), max(0.0, length(p.xy) - 1.6339746)), 0.8);
  return gslCarry29;
}

fn gsl_remap_50(p: vec3<f32>) -> vec3<f32> {
  let gsl_tmp_92 = 6.2831855 * p.z / 0.8;
  let gsl_tmp_91 = max(length(p.xy), 0.2);
  let gslCarry26 = thread_paint(vec3<f32>(0.77, 0.78, 0.78), -min(min(clamp((acos(clamp(cos(gsl_tmp_92) * p.x / gsl_tmp_91 + sin(gsl_tmp_92) * p.y / gsl_tmp_91, -1.0, 1.0)) / 3.1415927 * 0.4 - 0.05) * 1.7320508, 0.0, 0.43301272), max(0.0, 0.5 * min(min(0.0 - p.z, p.z - -16.0), 2.9330127 - length(p.xy)))), max(0.0, length(p.xy) - 1.6339746)), 0.8);
  return gslCarry26;
}

fn gsl_remap_51(p: vec3<f32>) -> vec3<f32> {
  let gsl_tmp_118 = 6.2831855 * p.z / 0.8;
  let gsl_tmp_117 = max(length(p.xy), 0.2);
  let gslCarry35 = thread_reliefColor(vec3<f32>(0.77, 0.78, 0.78), -min(min(clamp((acos(clamp(cos(gsl_tmp_118) * p.x / gsl_tmp_117 + sin(gsl_tmp_118) * p.y / gsl_tmp_117, -1.0, 1.0)) / 3.1415927 * 0.4 - 0.05) * 1.7320508, 0.0, 0.43301272), max(0.0, 0.5 * min(min(0.0 - p.z, p.z - -16.0), 2.9330127 - length(p.xy)))), max(0.0, length(p.xy) - 1.6339746)), 0.8);
  return gslCarry35;
}

fn gsl_remap_52(p: vec3<f32>) -> vec3<f32> {
  let gsl_tmp_112 = 6.2831855 * p.z / 0.8;
  let gsl_tmp_111 = max(length(p.xy), 0.2);
  let gslCarry32 = thread_paint(vec3<f32>(0.77, 0.78, 0.78), -min(min(clamp((acos(clamp(cos(gsl_tmp_112) * p.x / gsl_tmp_111 + sin(gsl_tmp_112) * p.y / gsl_tmp_111, -1.0, 1.0)) / 3.1415927 * 0.4 - 0.05) * 1.7320508, 0.0, 0.43301272), max(0.0, 0.5 * min(min(0.0 - p.z, p.z - -16.0), 2.9330127 - length(p.xy)))), max(0.0, length(p.xy) - 1.6339746)), 0.8);
  return gslCarry32;
}

fn gsl_remap_53(p: vec3<f32>) -> vec3<f32> {
  let gsl_tmp_146 = 6.2831855 * p.z / 0.8;
  let gsl_tmp_145 = max(length(p.xy), 0.2);
  let gslCarry41 = thread_reliefColor(vec3<f32>(0.77, 0.78, 0.78), -min(max(0.43301272 - clamp((acos(clamp(cos(gsl_tmp_146) * p.x / gsl_tmp_145 + sin(gsl_tmp_146) * p.y / gsl_tmp_145, -1.0, 1.0)) / 3.1415927 * 0.4 - 0.05) * 1.7320508, 0.0, 0.43301272), 0.0), max(0.0, 0.5 * min(min(4.7 - p.z, p.z - 0.0), 3.3660254 - length(p.xy)))), 0.8);
  return gslCarry41;
}

fn gsl_remap_54(p: vec3<f32>) -> vec3<f32> {
  let gsl_tmp_134 = 6.2831855 * p.z / 0.8;
  let gsl_tmp_133 = max(length(p.xy), 0.2);
  let gslCarry38 = thread_paint(vec3<f32>(0.77, 0.78, 0.78), -min(max(0.43301272 - clamp((acos(clamp(cos(gsl_tmp_134) * p.x / gsl_tmp_133 + sin(gsl_tmp_134) * p.y / gsl_tmp_133, -1.0, 1.0)) / 3.1415927 * 0.4 - 0.05) * 1.7320508, 0.0, 0.43301272), 0.0), max(0.0, 0.5 * min(min(4.7 - p.z, p.z - 0.0), 3.3660254 - length(p.xy)))), 0.8);
  return gslCarry38;
}

fn gsl_remap_55(p: vec3<f32>) -> vec3<f32> {
  let gsl_tmp_170 = 6.2831855 * p.z / 0.8;
  let gsl_tmp_169 = max(length(p.xy), 0.2);
  let gslCarry47 = thread_reliefColor(vec3<f32>(0.77, 0.78, 0.78), -min(max(0.43301272 - clamp((acos(clamp(cos(gsl_tmp_170) * p.x / gsl_tmp_169 + sin(gsl_tmp_170) * p.y / gsl_tmp_169, -1.0, 1.0)) / 3.1415927 * 0.4 - 0.05) * 1.7320508, 0.0, 0.43301272), 0.0), max(0.0, 0.5 * min(min(4.7 - p.z, p.z - 0.0), 3.3660254 - length(p.xy)))), 0.8);
  return gslCarry47;
}

fn gsl_remap_56(p: vec3<f32>) -> vec3<f32> {
  let gsl_tmp_158 = 6.2831855 * p.z / 0.8;
  let gsl_tmp_157 = max(length(p.xy), 0.2);
  let gslCarry44 = thread_paint(vec3<f32>(0.77, 0.78, 0.78), -min(max(0.43301272 - clamp((acos(clamp(cos(gsl_tmp_158) * p.x / gsl_tmp_157 + sin(gsl_tmp_158) * p.y / gsl_tmp_157, -1.0, 1.0)) / 3.1415927 * 0.4 - 0.05) * 1.7320508, 0.0, 0.43301272), 0.0), max(0.0, 0.5 * min(min(4.7 - p.z, p.z - 0.0), 3.3660254 - length(p.xy)))), 0.8);
  return gslCarry44;
}

fn gsl_remap_57(p: vec3<f32>) -> vec3<f32> {
  let gsl_tmp_186 = 6.2831855 * p.z / 0.8;
  let gsl_tmp_185 = max(length(p.xy), 0.2);
  let gslCarry53 = thread_reliefColor(vec3<f32>(0.77, 0.78, 0.78), -min(min(0.21650636 * (1.0 - clamp(cos(gsl_tmp_186) * p.x / gsl_tmp_185 + sin(gsl_tmp_186) * p.y / gsl_tmp_185, -1.0, 1.0)), max(0.0, 0.5 * min(min(0.0 - p.z, p.z - -16.0), 2.9330127 - length(p.xy)))), max(0.0, length(p.xy) - 1.6339746)), 0.8);
  return gslCarry53;
}

fn gsl_remap_58(p: vec3<f32>) -> vec3<f32> {
  let gsl_tmp_180 = 6.2831855 * p.z / 0.8;
  let gsl_tmp_179 = max(length(p.xy), 0.2);
  let gslCarry50 = thread_paint(vec3<f32>(0.77, 0.78, 0.78), -min(min(0.21650636 * (1.0 - clamp(cos(gsl_tmp_180) * p.x / gsl_tmp_179 + sin(gsl_tmp_180) * p.y / gsl_tmp_179, -1.0, 1.0)), max(0.0, 0.5 * min(min(0.0 - p.z, p.z - -16.0), 2.9330127 - length(p.xy)))), max(0.0, length(p.xy) - 1.6339746)), 0.8);
  return gslCarry50;
}

fn gsl_remap_59(p: vec3<f32>) -> vec3<f32> {
  let gsl_tmp_206 = 6.2831855 * p.z / 0.8;
  let gsl_tmp_205 = max(length(p.xy), 0.2);
  let gslCarry59 = thread_reliefColor(vec3<f32>(0.77, 0.78, 0.78), -min(min(0.21650636 * (1.0 - clamp(cos(gsl_tmp_206) * p.x / gsl_tmp_205 + sin(gsl_tmp_206) * p.y / gsl_tmp_205, -1.0, 1.0)), max(0.0, 0.5 * min(min(0.0 - p.z, p.z - -16.0), 2.9330127 - length(p.xy)))), max(0.0, length(p.xy) - 1.6339746)), 0.8);
  return gslCarry59;
}

fn gsl_remap_60(p: vec3<f32>) -> vec3<f32> {
  let gsl_tmp_200 = 6.2831855 * p.z / 0.8;
  let gsl_tmp_199 = max(length(p.xy), 0.2);
  let gslCarry56 = thread_paint(vec3<f32>(0.77, 0.78, 0.78), -min(min(0.21650636 * (1.0 - clamp(cos(gsl_tmp_200) * p.x / gsl_tmp_199 + sin(gsl_tmp_200) * p.y / gsl_tmp_199, -1.0, 1.0)), max(0.0, 0.5 * min(min(0.0 - p.z, p.z - -16.0), 2.9330127 - length(p.xy)))), max(0.0, length(p.xy) - 1.6339746)), 0.8);
  return gslCarry56;
}

fn gsl_remap_61(p: vec3<f32>) -> vec3<f32> {
  let gsl_tmp_234 = 6.2831855 * p.z / 0.8;
  let gsl_tmp_233 = max(length(p.xy), 0.2);
  let gslCarry65 = thread_reliefColor(vec3<f32>(0.77, 0.78, 0.78), -min(max(0.43301272 - 0.21650636 * (1.0 - clamp(cos(gsl_tmp_234) * p.x / gsl_tmp_233 + sin(gsl_tmp_234) * p.y / gsl_tmp_233, -1.0, 1.0)), 0.0), max(0.0, 0.5 * min(min(4.7 - p.z, p.z - 0.0), 3.3660254 - length(p.xy)))), 0.8);
  return gslCarry65;
}

fn gsl_remap_62(p: vec3<f32>) -> vec3<f32> {
  let gsl_tmp_222 = 6.2831855 * p.z / 0.8;
  let gsl_tmp_221 = max(length(p.xy), 0.2);
  let gslCarry62 = thread_paint(vec3<f32>(0.77, 0.78, 0.78), -min(max(0.43301272 - 0.21650636 * (1.0 - clamp(cos(gsl_tmp_222) * p.x / gsl_tmp_221 + sin(gsl_tmp_222) * p.y / gsl_tmp_221, -1.0, 1.0)), 0.0), max(0.0, 0.5 * min(min(4.7 - p.z, p.z - 0.0), 3.3660254 - length(p.xy)))), 0.8);
  return gslCarry62;
}

fn gsl_remap_63(p: vec3<f32>) -> vec3<f32> {
  let gsl_tmp_258 = 6.2831855 * p.z / 0.8;
  let gsl_tmp_257 = max(length(p.xy), 0.2);
  let gslCarry71 = thread_reliefColor(vec3<f32>(0.77, 0.78, 0.78), -min(max(0.43301272 - 0.21650636 * (1.0 - clamp(cos(gsl_tmp_258) * p.x / gsl_tmp_257 + sin(gsl_tmp_258) * p.y / gsl_tmp_257, -1.0, 1.0)), 0.0), max(0.0, 0.5 * min(min(4.7 - p.z, p.z - 0.0), 3.3660254 - length(p.xy)))), 0.8);
  return gslCarry71;
}

fn gsl_remap_64(p: vec3<f32>) -> vec3<f32> {
  let gsl_tmp_246 = 6.2831855 * p.z / 0.8;
  let gsl_tmp_245 = max(length(p.xy), 0.2);
  let gslCarry68 = thread_paint(vec3<f32>(0.77, 0.78, 0.78), -min(max(0.43301272 - 0.21650636 * (1.0 - clamp(cos(gsl_tmp_246) * p.x / gsl_tmp_245 + sin(gsl_tmp_246) * p.y / gsl_tmp_245, -1.0, 1.0)), 0.0), max(0.0, 0.5 * min(min(4.7 - p.z, p.z - 0.0), 3.3660254 - length(p.xy)))), 0.8);
  return gslCarry68;
}

fn gsl_remap_65(p: vec3<f32>) -> vec3<f32> {
  let gsl_tmp_268 = 6.2831855 * p.z / 0.8;
  let gsl_tmp_267 = max(length(p.xy), 0.2);
  let gslCarry74 = thread_paint(vec3<f32>(0.77, 0.78, 0.78), -min(min(0.21650636 * (1.0 - clamp(cos(gsl_tmp_268) * p.x / gsl_tmp_267 + sin(gsl_tmp_268) * p.y / gsl_tmp_267, -1.0, 1.0)), max(0.0, 0.5 * min(min(0.0 - p.z, p.z - -16.0), 2.9330127 - length(p.xy)))), max(0.0, length(p.xy) - 1.6339746)), 0.8);
  return gslCarry74;
}

fn gsl_remap_66(p: vec3<f32>) -> vec3<f32> {
  let gsl_tmp_284 = 6.2831855 * p.z / 0.8;
  let gsl_tmp_283 = max(length(p.xy), 0.2);
  let gslCarry78 = thread_paint(vec3<f32>(0.77, 0.78, 0.78), -min(min(0.21650636 * (1.0 - clamp(cos(gsl_tmp_284) * p.x / gsl_tmp_283 + sin(gsl_tmp_284) * p.y / gsl_tmp_283, -1.0, 1.0)), max(0.0, 0.5 * min(min(0.0 - p.z, p.z - -16.0), 2.9330127 - length(p.xy)))), max(0.0, length(p.xy) - 1.6339746)), 0.8);
  return gslCarry78;
}

fn gsl_remap_67(p: vec3<f32>) -> vec3<f32> {
  let gsl_tmp_302 = 6.2831855 * p.z / 0.8;
  let gsl_tmp_301 = max(length(p.xy), 0.2);
  let gslCarry82 = thread_paint(vec3<f32>(0.77, 0.78, 0.78), -min(max(0.43301272 - 0.21650636 * (1.0 - clamp(cos(gsl_tmp_302) * p.x / gsl_tmp_301 + sin(gsl_tmp_302) * p.y / gsl_tmp_301, -1.0, 1.0)), 0.0), max(0.0, 0.5 * min(min(4.7 - p.z, p.z - 0.0), 3.3660254 - length(p.xy)))), 0.8);
  return gslCarry82;
}

fn gsl_remap_68(p: vec3<f32>) -> vec3<f32> {
  let gsl_tmp_318 = 6.2831855 * p.z / 0.8;
  let gsl_tmp_317 = max(length(p.xy), 0.2);
  let gslCarry86 = thread_paint(vec3<f32>(0.77, 0.78, 0.78), -min(max(0.43301272 - 0.21650636 * (1.0 - clamp(cos(gsl_tmp_318) * p.x / gsl_tmp_317 + sin(gsl_tmp_318) * p.y / gsl_tmp_317, -1.0, 1.0)), 0.0), max(0.0, 0.5 * min(min(4.7 - p.z, p.z - 0.0), 3.3660254 - length(p.xy)))), 0.8);
  return gslCarry86;
}

fn gsl_remap_69(p: vec3<f32>) -> vec3<f32> {
  let gsl_tmp_332 = 6.2831855 * p.z / 0.8;
  let gsl_tmp_331 = max(length(p.xy), 0.2);
  let gslCarry90 = thread_paint(vec3<f32>(0.77, 0.78, 0.78), -min(min(clamp((acos(clamp(cos(gsl_tmp_332) * p.x / gsl_tmp_331 + sin(gsl_tmp_332) * p.y / gsl_tmp_331, -1.0, 1.0)) / 3.1415927 * 0.4 - 0.05) * 1.7320508, 0.0, 0.43301272), max(0.0, 0.5 * min(min(0.0 - p.z, p.z - -16.0), 2.9330127 - length(p.xy)))), max(0.0, length(p.xy) - 1.6339746)), 0.8);
  return gslCarry90;
}

fn gsl_remap_70(p: vec3<f32>) -> vec3<f32> {
  let gsl_tmp_349 = 6.2831855 * p.z / 0.8;
  let gsl_tmp_348 = max(length(p.xy), 0.2);
  let gslCarry94 = thread_paint(vec3<f32>(0.77, 0.78, 0.78), -min(min(clamp((acos(clamp(cos(gsl_tmp_349) * p.x / gsl_tmp_348 + sin(gsl_tmp_349) * p.y / gsl_tmp_348, -1.0, 1.0)) / 3.1415927 * 0.4 - 0.05) * 1.7320508, 0.0, 0.43301272), max(0.0, 0.5 * min(min(0.0 - p.z, p.z - -16.0), 2.9330127 - length(p.xy)))), max(0.0, length(p.xy) - 1.6339746)), 0.8);
  return gslCarry94;
}

fn gsl_remap_71(p: vec3<f32>) -> vec3<f32> {
  let gsl_tmp_368 = 6.2831855 * p.z / 0.8;
  let gsl_tmp_367 = max(length(p.xy), 0.2);
  let gslCarry98 = thread_paint(vec3<f32>(0.77, 0.78, 0.78), -min(max(0.43301272 - clamp((acos(clamp(cos(gsl_tmp_368) * p.x / gsl_tmp_367 + sin(gsl_tmp_368) * p.y / gsl_tmp_367, -1.0, 1.0)) / 3.1415927 * 0.4 - 0.05) * 1.7320508, 0.0, 0.43301272), 0.0), max(0.0, 0.5 * min(min(4.7 - p.z, p.z - 0.0), 3.3660254 - length(p.xy)))), 0.8);
  return gslCarry98;
}

fn gsl_remap_72(p: vec3<f32>) -> vec3<f32> {
  let gsl_tmp_385 = 6.2831855 * p.z / 0.8;
  let gsl_tmp_384 = max(length(p.xy), 0.2);
  let gslCarry102 = thread_paint(vec3<f32>(0.77, 0.78, 0.78), -min(max(0.43301272 - clamp((acos(clamp(cos(gsl_tmp_385) * p.x / gsl_tmp_384 + sin(gsl_tmp_385) * p.y / gsl_tmp_384, -1.0, 1.0)) / 3.1415927 * 0.4 - 0.05) * 1.7320508, 0.0, 0.43301272), 0.0), max(0.0, 0.5 * min(min(4.7 - p.z, p.z - 0.0), 3.3660254 - length(p.xy)))), 0.8);
  return gslCarry102;
}

fn gsl_remap_73(p: vec3<f32>) -> vec4<f32> {
  return vec4<f32>(0.45, 1.0, 1.5, 0.0);
}

fn gsl_remap_74(p: vec3<f32>) -> vec4<f32> {
  return vec4<f32>(0.0, 0.0, 0.0, 0.0);
}

fn gsl_remap_75(p: vec3<f32>) -> vec4<f32> {
  let gsl_tmp_102 = 6.2831855 * p.z / 0.8;
  let gsl_tmp_101 = max(length(p.xy), 0.2);
  let gslCarry30 = material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), -min(min(clamp((acos(clamp(cos(gsl_tmp_102) * p.x / gsl_tmp_101 + sin(gsl_tmp_102) * p.y / gsl_tmp_101, -1.0, 1.0)) / 3.1415927 * 0.4 - 0.05) * 1.7320508, 0.0, 0.43301272), max(0.0, 0.5 * min(min(0.0 - p.z, p.z - -16.0), 2.9330127 - length(p.xy)))), max(0.0, length(p.xy) - 1.6339746)));
  return gslCarry30[0];
}

fn gsl_remap_76(p: vec3<f32>) -> vec4<f32> {
  let gsl_tmp_102 = 6.2831855 * p.z / 0.8;
  let gsl_tmp_101 = max(length(p.xy), 0.2);
  let gslCarry30 = material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), -min(min(clamp((acos(clamp(cos(gsl_tmp_102) * p.x / gsl_tmp_101 + sin(gsl_tmp_102) * p.y / gsl_tmp_101, -1.0, 1.0)) / 3.1415927 * 0.4 - 0.05) * 1.7320508, 0.0, 0.43301272), max(0.0, 0.5 * min(min(0.0 - p.z, p.z - -16.0), 2.9330127 - length(p.xy)))), max(0.0, length(p.xy) - 1.6339746)));
  return gslCarry30[1];
}

fn gsl_remap_77(p: vec3<f32>) -> vec4<f32> {
  let gsl_tmp_102 = 6.2831855 * p.z / 0.8;
  let gsl_tmp_101 = max(length(p.xy), 0.2);
  let gslCarry30 = material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), -min(min(clamp((acos(clamp(cos(gsl_tmp_102) * p.x / gsl_tmp_101 + sin(gsl_tmp_102) * p.y / gsl_tmp_101, -1.0, 1.0)) / 3.1415927 * 0.4 - 0.05) * 1.7320508, 0.0, 0.43301272), max(0.0, 0.5 * min(min(0.0 - p.z, p.z - -16.0), 2.9330127 - length(p.xy)))), max(0.0, length(p.xy) - 1.6339746)));
  return gslCarry30[2];
}

fn gsl_remap_78(p: vec3<f32>) -> vec4<f32> {
  let gsl_tmp_122 = 6.2831855 * p.z / 0.8;
  let gsl_tmp_121 = max(length(p.xy), 0.2);
  let gslCarry36 = material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), -min(min(clamp((acos(clamp(cos(gsl_tmp_122) * p.x / gsl_tmp_121 + sin(gsl_tmp_122) * p.y / gsl_tmp_121, -1.0, 1.0)) / 3.1415927 * 0.4 - 0.05) * 1.7320508, 0.0, 0.43301272), max(0.0, 0.5 * min(min(0.0 - p.z, p.z - -16.0), 2.9330127 - length(p.xy)))), max(0.0, length(p.xy) - 1.6339746)));
  return gslCarry36[0];
}

fn gsl_remap_79(p: vec3<f32>) -> vec4<f32> {
  let gsl_tmp_122 = 6.2831855 * p.z / 0.8;
  let gsl_tmp_121 = max(length(p.xy), 0.2);
  let gslCarry36 = material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), -min(min(clamp((acos(clamp(cos(gsl_tmp_122) * p.x / gsl_tmp_121 + sin(gsl_tmp_122) * p.y / gsl_tmp_121, -1.0, 1.0)) / 3.1415927 * 0.4 - 0.05) * 1.7320508, 0.0, 0.43301272), max(0.0, 0.5 * min(min(0.0 - p.z, p.z - -16.0), 2.9330127 - length(p.xy)))), max(0.0, length(p.xy) - 1.6339746)));
  return gslCarry36[1];
}

fn gsl_remap_80(p: vec3<f32>) -> vec4<f32> {
  let gsl_tmp_122 = 6.2831855 * p.z / 0.8;
  let gsl_tmp_121 = max(length(p.xy), 0.2);
  let gslCarry36 = material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), -min(min(clamp((acos(clamp(cos(gsl_tmp_122) * p.x / gsl_tmp_121 + sin(gsl_tmp_122) * p.y / gsl_tmp_121, -1.0, 1.0)) / 3.1415927 * 0.4 - 0.05) * 1.7320508, 0.0, 0.43301272), max(0.0, 0.5 * min(min(0.0 - p.z, p.z - -16.0), 2.9330127 - length(p.xy)))), max(0.0, length(p.xy) - 1.6339746)));
  return gslCarry36[2];
}

fn gsl_remap_81(p: vec3<f32>) -> vec4<f32> {
  let gsl_tmp_152 = 6.2831855 * p.z / 0.8;
  let gsl_tmp_151 = max(length(p.xy), 0.2);
  let gslCarry42 = material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), -min(max(0.43301272 - clamp((acos(clamp(cos(gsl_tmp_152) * p.x / gsl_tmp_151 + sin(gsl_tmp_152) * p.y / gsl_tmp_151, -1.0, 1.0)) / 3.1415927 * 0.4 - 0.05) * 1.7320508, 0.0, 0.43301272), 0.0), max(0.0, 0.5 * min(min(4.7 - p.z, p.z - 0.0), 3.3660254 - length(p.xy)))));
  return gslCarry42[0];
}

fn gsl_remap_82(p: vec3<f32>) -> vec4<f32> {
  let gsl_tmp_152 = 6.2831855 * p.z / 0.8;
  let gsl_tmp_151 = max(length(p.xy), 0.2);
  let gslCarry42 = material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), -min(max(0.43301272 - clamp((acos(clamp(cos(gsl_tmp_152) * p.x / gsl_tmp_151 + sin(gsl_tmp_152) * p.y / gsl_tmp_151, -1.0, 1.0)) / 3.1415927 * 0.4 - 0.05) * 1.7320508, 0.0, 0.43301272), 0.0), max(0.0, 0.5 * min(min(4.7 - p.z, p.z - 0.0), 3.3660254 - length(p.xy)))));
  return gslCarry42[1];
}

fn gsl_remap_83(p: vec3<f32>) -> vec4<f32> {
  let gsl_tmp_152 = 6.2831855 * p.z / 0.8;
  let gsl_tmp_151 = max(length(p.xy), 0.2);
  let gslCarry42 = material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), -min(max(0.43301272 - clamp((acos(clamp(cos(gsl_tmp_152) * p.x / gsl_tmp_151 + sin(gsl_tmp_152) * p.y / gsl_tmp_151, -1.0, 1.0)) / 3.1415927 * 0.4 - 0.05) * 1.7320508, 0.0, 0.43301272), 0.0), max(0.0, 0.5 * min(min(4.7 - p.z, p.z - 0.0), 3.3660254 - length(p.xy)))));
  return gslCarry42[2];
}

fn gsl_remap_84(p: vec3<f32>) -> vec4<f32> {
  let gsl_tmp_176 = 6.2831855 * p.z / 0.8;
  let gsl_tmp_175 = max(length(p.xy), 0.2);
  let gslCarry48 = material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), -min(max(0.43301272 - clamp((acos(clamp(cos(gsl_tmp_176) * p.x / gsl_tmp_175 + sin(gsl_tmp_176) * p.y / gsl_tmp_175, -1.0, 1.0)) / 3.1415927 * 0.4 - 0.05) * 1.7320508, 0.0, 0.43301272), 0.0), max(0.0, 0.5 * min(min(4.7 - p.z, p.z - 0.0), 3.3660254 - length(p.xy)))));
  return gslCarry48[0];
}

fn gsl_remap_85(p: vec3<f32>) -> vec4<f32> {
  let gsl_tmp_176 = 6.2831855 * p.z / 0.8;
  let gsl_tmp_175 = max(length(p.xy), 0.2);
  let gslCarry48 = material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), -min(max(0.43301272 - clamp((acos(clamp(cos(gsl_tmp_176) * p.x / gsl_tmp_175 + sin(gsl_tmp_176) * p.y / gsl_tmp_175, -1.0, 1.0)) / 3.1415927 * 0.4 - 0.05) * 1.7320508, 0.0, 0.43301272), 0.0), max(0.0, 0.5 * min(min(4.7 - p.z, p.z - 0.0), 3.3660254 - length(p.xy)))));
  return gslCarry48[1];
}

fn gsl_remap_86(p: vec3<f32>) -> vec4<f32> {
  let gsl_tmp_176 = 6.2831855 * p.z / 0.8;
  let gsl_tmp_175 = max(length(p.xy), 0.2);
  let gslCarry48 = material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), -min(max(0.43301272 - clamp((acos(clamp(cos(gsl_tmp_176) * p.x / gsl_tmp_175 + sin(gsl_tmp_176) * p.y / gsl_tmp_175, -1.0, 1.0)) / 3.1415927 * 0.4 - 0.05) * 1.7320508, 0.0, 0.43301272), 0.0), max(0.0, 0.5 * min(min(4.7 - p.z, p.z - 0.0), 3.3660254 - length(p.xy)))));
  return gslCarry48[2];
}

fn gsl_remap_87(p: vec3<f32>) -> vec4<f32> {
  let gsl_tmp_190 = 6.2831855 * p.z / 0.8;
  let gsl_tmp_189 = max(length(p.xy), 0.2);
  let gslCarry54 = material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), -min(min(0.21650636 * (1.0 - clamp(cos(gsl_tmp_190) * p.x / gsl_tmp_189 + sin(gsl_tmp_190) * p.y / gsl_tmp_189, -1.0, 1.0)), max(0.0, 0.5 * min(min(0.0 - p.z, p.z - -16.0), 2.9330127 - length(p.xy)))), max(0.0, length(p.xy) - 1.6339746)));
  return gslCarry54[0];
}

fn gsl_remap_88(p: vec3<f32>) -> vec4<f32> {
  let gsl_tmp_190 = 6.2831855 * p.z / 0.8;
  let gsl_tmp_189 = max(length(p.xy), 0.2);
  let gslCarry54 = material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), -min(min(0.21650636 * (1.0 - clamp(cos(gsl_tmp_190) * p.x / gsl_tmp_189 + sin(gsl_tmp_190) * p.y / gsl_tmp_189, -1.0, 1.0)), max(0.0, 0.5 * min(min(0.0 - p.z, p.z - -16.0), 2.9330127 - length(p.xy)))), max(0.0, length(p.xy) - 1.6339746)));
  return gslCarry54[1];
}

fn gsl_remap_89(p: vec3<f32>) -> vec4<f32> {
  let gsl_tmp_190 = 6.2831855 * p.z / 0.8;
  let gsl_tmp_189 = max(length(p.xy), 0.2);
  let gslCarry54 = material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), -min(min(0.21650636 * (1.0 - clamp(cos(gsl_tmp_190) * p.x / gsl_tmp_189 + sin(gsl_tmp_190) * p.y / gsl_tmp_189, -1.0, 1.0)), max(0.0, 0.5 * min(min(0.0 - p.z, p.z - -16.0), 2.9330127 - length(p.xy)))), max(0.0, length(p.xy) - 1.6339746)));
  return gslCarry54[2];
}

fn gsl_remap_90(p: vec3<f32>) -> vec4<f32> {
  let gsl_tmp_210 = 6.2831855 * p.z / 0.8;
  let gsl_tmp_209 = max(length(p.xy), 0.2);
  let gslCarry60 = material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), -min(min(0.21650636 * (1.0 - clamp(cos(gsl_tmp_210) * p.x / gsl_tmp_209 + sin(gsl_tmp_210) * p.y / gsl_tmp_209, -1.0, 1.0)), max(0.0, 0.5 * min(min(0.0 - p.z, p.z - -16.0), 2.9330127 - length(p.xy)))), max(0.0, length(p.xy) - 1.6339746)));
  return gslCarry60[0];
}

fn gsl_remap_91(p: vec3<f32>) -> vec4<f32> {
  let gsl_tmp_210 = 6.2831855 * p.z / 0.8;
  let gsl_tmp_209 = max(length(p.xy), 0.2);
  let gslCarry60 = material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), -min(min(0.21650636 * (1.0 - clamp(cos(gsl_tmp_210) * p.x / gsl_tmp_209 + sin(gsl_tmp_210) * p.y / gsl_tmp_209, -1.0, 1.0)), max(0.0, 0.5 * min(min(0.0 - p.z, p.z - -16.0), 2.9330127 - length(p.xy)))), max(0.0, length(p.xy) - 1.6339746)));
  return gslCarry60[1];
}

fn gsl_remap_92(p: vec3<f32>) -> vec4<f32> {
  let gsl_tmp_210 = 6.2831855 * p.z / 0.8;
  let gsl_tmp_209 = max(length(p.xy), 0.2);
  let gslCarry60 = material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), -min(min(0.21650636 * (1.0 - clamp(cos(gsl_tmp_210) * p.x / gsl_tmp_209 + sin(gsl_tmp_210) * p.y / gsl_tmp_209, -1.0, 1.0)), max(0.0, 0.5 * min(min(0.0 - p.z, p.z - -16.0), 2.9330127 - length(p.xy)))), max(0.0, length(p.xy) - 1.6339746)));
  return gslCarry60[2];
}

fn gsl_remap_93(p: vec3<f32>) -> vec4<f32> {
  let gsl_tmp_240 = 6.2831855 * p.z / 0.8;
  let gsl_tmp_239 = max(length(p.xy), 0.2);
  let gslCarry66 = material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), -min(max(0.43301272 - 0.21650636 * (1.0 - clamp(cos(gsl_tmp_240) * p.x / gsl_tmp_239 + sin(gsl_tmp_240) * p.y / gsl_tmp_239, -1.0, 1.0)), 0.0), max(0.0, 0.5 * min(min(4.7 - p.z, p.z - 0.0), 3.3660254 - length(p.xy)))));
  return gslCarry66[0];
}

fn gsl_remap_94(p: vec3<f32>) -> vec4<f32> {
  let gsl_tmp_240 = 6.2831855 * p.z / 0.8;
  let gsl_tmp_239 = max(length(p.xy), 0.2);
  let gslCarry66 = material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), -min(max(0.43301272 - 0.21650636 * (1.0 - clamp(cos(gsl_tmp_240) * p.x / gsl_tmp_239 + sin(gsl_tmp_240) * p.y / gsl_tmp_239, -1.0, 1.0)), 0.0), max(0.0, 0.5 * min(min(4.7 - p.z, p.z - 0.0), 3.3660254 - length(p.xy)))));
  return gslCarry66[1];
}

fn gsl_remap_95(p: vec3<f32>) -> vec4<f32> {
  let gsl_tmp_240 = 6.2831855 * p.z / 0.8;
  let gsl_tmp_239 = max(length(p.xy), 0.2);
  let gslCarry66 = material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), -min(max(0.43301272 - 0.21650636 * (1.0 - clamp(cos(gsl_tmp_240) * p.x / gsl_tmp_239 + sin(gsl_tmp_240) * p.y / gsl_tmp_239, -1.0, 1.0)), 0.0), max(0.0, 0.5 * min(min(4.7 - p.z, p.z - 0.0), 3.3660254 - length(p.xy)))));
  return gslCarry66[2];
}

fn gsl_remap_96(p: vec3<f32>) -> vec4<f32> {
  let gsl_tmp_264 = 6.2831855 * p.z / 0.8;
  let gsl_tmp_263 = max(length(p.xy), 0.2);
  let gslCarry72 = material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), -min(max(0.43301272 - 0.21650636 * (1.0 - clamp(cos(gsl_tmp_264) * p.x / gsl_tmp_263 + sin(gsl_tmp_264) * p.y / gsl_tmp_263, -1.0, 1.0)), 0.0), max(0.0, 0.5 * min(min(4.7 - p.z, p.z - 0.0), 3.3660254 - length(p.xy)))));
  return gslCarry72[0];
}

fn gsl_remap_97(p: vec3<f32>) -> vec4<f32> {
  let gsl_tmp_264 = 6.2831855 * p.z / 0.8;
  let gsl_tmp_263 = max(length(p.xy), 0.2);
  let gslCarry72 = material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), -min(max(0.43301272 - 0.21650636 * (1.0 - clamp(cos(gsl_tmp_264) * p.x / gsl_tmp_263 + sin(gsl_tmp_264) * p.y / gsl_tmp_263, -1.0, 1.0)), 0.0), max(0.0, 0.5 * min(min(4.7 - p.z, p.z - 0.0), 3.3660254 - length(p.xy)))));
  return gslCarry72[1];
}

fn gsl_remap_98(p: vec3<f32>) -> vec4<f32> {
  let gsl_tmp_264 = 6.2831855 * p.z / 0.8;
  let gsl_tmp_263 = max(length(p.xy), 0.2);
  let gslCarry72 = material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), -min(max(0.43301272 - 0.21650636 * (1.0 - clamp(cos(gsl_tmp_264) * p.x / gsl_tmp_263 + sin(gsl_tmp_264) * p.y / gsl_tmp_263, -1.0, 1.0)), 0.0), max(0.0, 0.5 * min(min(4.7 - p.z, p.z - 0.0), 3.3660254 - length(p.xy)))));
  return gslCarry72[2];
}

fn fastener_nut(p: vec3<f32>, nominalDia: f32) -> f32 {
  let af = catalog_nutAcrossFlats(nominalDia);
  let m = catalog_nutHeight(nominalDia);
  let body = max(max(max(max(drill_hexXY(p, af), -p.z), p.z - m), (length(p.xy) - 0.52 * af + p.z - m) * 0.7071), (length(p.xy) - 0.52 * af + -p.z) * 0.7071);
  let bore = length(p.xy) - thread_minorDia(nominalDia, thread_metricPitch(nominalDia)) / 2.0;
  let form = max(body, -bore);
  return form;
}

fn fastener_screw(head: f32, shank: f32, drive: f32) -> f32 {
  return max(min(head, shank), -drive);
}

fn gsl_shapeParts(p: vec3<f32>, time: f32) -> array<f32, 24> {
  gsl_partNear = vec3<f32>(1e10, 1e10, 1e10);
  let gsl_partLo = vec3<f32>(-34.26, -20.26, -0.01);
  let gsl_partHi = vec3<f32>(-25.74, -11.74, 21.01);
  let gsl_partBox = gsl_boxDistance(p, gsl_partLo, gsl_partHi);
  var gsl_part = gsl_partBox;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox <= gsl_vanishView.w) {
    let gsl_partVanish = gsl_vanish(gsl_partLo, gsl_partHi);
    gsl_part = 1e10;
    let gsl_simpleWeight_45 = gsl_simple(vec3<f32>(-34.26, -20.26, -0.01), vec3<f32>(-25.74, -11.74, 21.01));
    var gsl_simplePart = 1e10;
    if (gsl_simpleWeight_45 > 0.0) {
      gsl_simplePart = gsl_remap(p - vec3<f32>(-30.0, -16.0, 16.0));
    }
    if (gsl_simpleWeight_45 < 1.0) {
      gsl_simplePart = mix(max(gsl_remap_2(p - vec3<f32>(-30.0, -16.0, 16.0)), gsl_boxSdf(p, vec3<f32>(-34.26, -20.26, -0.01), vec3<f32>(-25.74, -11.74, 21.01))), gsl_simplePart, gsl_simpleWeight_45);
    }
    if (gsl_partVanish.x > 0.0) { gsl_part = gsl_simplePart + gsl_partVanish.y; }
    if (abs(gsl_part) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part), gsl_partVanish.zw); }
  }
  let plainScrew = gsl_part;
  let gsl_partLo_2 = vec3<f32>(-34.26, -4.26, -0.01);
  let gsl_partHi_2 = vec3<f32>(-25.74, 4.26, 21.01);
  let gsl_partBox_2 = gsl_boxDistance(p, gsl_partLo_2, gsl_partHi_2);
  var gsl_part_2 = gsl_partBox_2;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_2 <= gsl_vanishView.w) {
    let gsl_partVanish_2 = gsl_vanish(gsl_partLo_2, gsl_partHi_2);
    gsl_part_2 = 1e10;
    let gsl_simpleWeight_46 = gsl_simple(vec3<f32>(-34.26, -4.26, -0.01), vec3<f32>(-25.74, 4.26, 21.01));
    var gsl_simplePart_2 = 1e10;
    if (gsl_simpleWeight_46 > 0.0) {
      gsl_simplePart_2 = gsl_remap(p - vec3<f32>(-30.0, 0.0, 16.0));
    }
    if (gsl_simpleWeight_46 < 1.0) {
      gsl_simplePart_2 = mix(max(gsl_remap_2(p - vec3<f32>(-30.0, 0.0, 16.0)), gsl_boxSdf(p, vec3<f32>(-34.26, -4.26, -0.01), vec3<f32>(-25.74, 4.26, 21.01))), gsl_simplePart_2, gsl_simpleWeight_46);
    }
    if (gsl_partVanish_2.x > 0.0) { gsl_part_2 = gsl_simplePart_2 + gsl_partVanish_2.y; }
    if (abs(gsl_part_2) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_2), gsl_partVanish_2.zw); }
  }
  let plainPairScrew = gsl_part_2;
  let gsl_partLo_3 = vec3<f32>(-34.01, -4.6288023, 3.19);
  let gsl_partHi_3 = vec3<f32>(-25.99, 4.6288023, 7.91);
  let gsl_partBox_3 = gsl_boxDistance(p, gsl_partLo_3, gsl_partHi_3);
  var gsl_part_3 = gsl_partBox_3;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_3 <= gsl_vanishView.w) {
    let gsl_partVanish_3 = gsl_vanish(gsl_partLo_3, gsl_partHi_3);
    gsl_part_3 = 1e10;
    let gsl_simpleWeight_47 = gsl_simple(vec3<f32>(-34.01, -4.6288023, 3.19), vec3<f32>(-25.99, 4.6288023, 7.91));
    var gsl_simplePart_3 = 1e10;
    if (gsl_simpleWeight_47 > 0.0) {
      gsl_simplePart_3 = fastener_nut(p - vec3<f32>(-30.0, 0.0, 3.2), 5.0);
    }
    if (gsl_simpleWeight_47 < 1.0) {
      gsl_simplePart_3 = mix(max(gsl_remap_3(p - vec3<f32>(-30.0, 0.0, 3.2)), gsl_boxSdf(p, vec3<f32>(-34.01, -4.6288023, 3.19), vec3<f32>(-25.99, 4.6288023, 7.91))), gsl_simplePart_3, gsl_simpleWeight_47);
    }
    if (gsl_partVanish_3.x > 0.0) { gsl_part_3 = gsl_simplePart_3 + gsl_partVanish_3.y; }
    if (abs(gsl_part_3) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_3), gsl_partVanish_3.zw); }
  }
  let plainPairNut = gsl_part_3;
  let gsl_partLo_4 = vec3<f32>(-34.01, 11.371198, -0.01);
  let gsl_partHi_4 = vec3<f32>(-25.99, 20.628801, 4.71);
  let gsl_partBox_4 = gsl_boxDistance(p, gsl_partLo_4, gsl_partHi_4);
  var gsl_part_4 = gsl_partBox_4;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_4 <= gsl_vanishView.w) {
    let gsl_partVanish_4 = gsl_vanish(gsl_partLo_4, gsl_partHi_4);
    gsl_part_4 = 1e10;
    let gsl_simpleWeight_48 = gsl_simple(vec3<f32>(-34.01, 11.371198, -0.01), vec3<f32>(-25.99, 20.628801, 4.71));
    var gsl_simplePart_4 = 1e10;
    if (gsl_simpleWeight_48 > 0.0) {
      gsl_simplePart_4 = fastener_nut(p - vec3<f32>(-30.0, 16.0, 0.0), 5.0);
    }
    if (gsl_simpleWeight_48 < 1.0) {
      gsl_simplePart_4 = mix(max(gsl_remap_3(p - vec3<f32>(-30.0, 16.0, 0.0)), gsl_boxSdf(p, vec3<f32>(-34.01, 11.371198, -0.01), vec3<f32>(-25.99, 20.628801, 4.71))), gsl_simplePart_4, gsl_simpleWeight_48);
    }
    if (gsl_partVanish_4.x > 0.0) { gsl_part_4 = gsl_simplePart_4 + gsl_partVanish_4.y; }
    if (abs(gsl_part_4) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_4), gsl_partVanish_4.zw); }
  }
  let plainNut = gsl_part_4;
  let gsl_partLo_5 = vec3<f32>(-22.26, -20.26, -0.01);
  let gsl_partHi_5 = vec3<f32>(-13.74, -11.74, 21.01);
  let gsl_partBox_5 = gsl_boxDistance(p, gsl_partLo_5, gsl_partHi_5);
  var gsl_part_5 = gsl_partBox_5;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_5 <= gsl_vanishView.w) {
    let gsl_partVanish_5 = gsl_vanish(gsl_partLo_5, gsl_partHi_5);
    gsl_part_5 = 1e10;
    let gsl_simpleWeight_49 = gsl_simple(vec3<f32>(-22.26, -20.26, -0.01), vec3<f32>(-13.74, -11.74, 21.01));
    var gsl_simplePart_5 = 1e10;
    if (gsl_simpleWeight_49 > 0.0) {
      gsl_simplePart_5 = gsl_remap_4(p - vec3<f32>(-18.0, -16.0, 16.0));
    }
    if (gsl_simpleWeight_49 < 1.0) {
      gsl_simplePart_5 = mix(max(gsl_remap_5(p - vec3<f32>(-18.0, -16.0, 16.0)), gsl_boxSdf(p, vec3<f32>(-22.26, -20.26, -0.01), vec3<f32>(-13.74, -11.74, 21.01))), gsl_simplePart_5, gsl_simpleWeight_49);
    }
    if (gsl_partVanish_5.x > 0.0) { gsl_part_5 = gsl_simplePart_5 + gsl_partVanish_5.y; }
    if (abs(gsl_part_5) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_5), gsl_partVanish_5.zw); }
  }
  let paintScrew = gsl_part_5;
  let gsl_partLo_6 = vec3<f32>(-22.26, -4.26, -0.01);
  let gsl_partHi_6 = vec3<f32>(-13.74, 4.26, 21.01);
  let gsl_partBox_6 = gsl_boxDistance(p, gsl_partLo_6, gsl_partHi_6);
  var gsl_part_6 = gsl_partBox_6;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_6 <= gsl_vanishView.w) {
    let gsl_partVanish_6 = gsl_vanish(gsl_partLo_6, gsl_partHi_6);
    gsl_part_6 = 1e10;
    let gsl_simpleWeight_50 = gsl_simple(vec3<f32>(-22.26, -4.26, -0.01), vec3<f32>(-13.74, 4.26, 21.01));
    var gsl_simplePart_6 = 1e10;
    if (gsl_simpleWeight_50 > 0.0) {
      gsl_simplePart_6 = gsl_remap_6(p - vec3<f32>(-18.0, 0.0, 16.0));
    }
    if (gsl_simpleWeight_50 < 1.0) {
      gsl_simplePart_6 = mix(max(gsl_remap_7(p - vec3<f32>(-18.0, 0.0, 16.0)), gsl_boxSdf(p, vec3<f32>(-22.26, -4.26, -0.01), vec3<f32>(-13.74, 4.26, 21.01))), gsl_simplePart_6, gsl_simpleWeight_50);
    }
    if (gsl_partVanish_6.x > 0.0) { gsl_part_6 = gsl_simplePart_6 + gsl_partVanish_6.y; }
    if (abs(gsl_part_6) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_6), gsl_partVanish_6.zw); }
  }
  let paintPairScrew = gsl_part_6;
  let gsl_partLo_7 = vec3<f32>(-22.01, -4.6288023, 3.19);
  let gsl_partHi_7 = vec3<f32>(-13.99, 4.6288023, 7.91);
  let gsl_partBox_7 = gsl_boxDistance(p, gsl_partLo_7, gsl_partHi_7);
  var gsl_part_7 = gsl_partBox_7;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_7 <= gsl_vanishView.w) {
    let gsl_partVanish_7 = gsl_vanish(gsl_partLo_7, gsl_partHi_7);
    gsl_part_7 = 1e10;
    let gsl_simpleWeight_51 = gsl_simple(vec3<f32>(-22.01, -4.6288023, 3.19), vec3<f32>(-13.99, 4.6288023, 7.91));
    var gsl_simplePart_7 = 1e10;
    if (gsl_simpleWeight_51 > 0.0) {
      gsl_simplePart_7 = gsl_remap_8(p - vec3<f32>(-18.0, 0.0, 3.2));
    }
    if (gsl_simpleWeight_51 < 1.0) {
      gsl_simplePart_7 = mix(max(gsl_remap_9(p - vec3<f32>(-18.0, 0.0, 3.2)), gsl_boxSdf(p, vec3<f32>(-22.01, -4.6288023, 3.19), vec3<f32>(-13.99, 4.6288023, 7.91))), gsl_simplePart_7, gsl_simpleWeight_51);
    }
    if (gsl_partVanish_7.x > 0.0) { gsl_part_7 = gsl_simplePart_7 + gsl_partVanish_7.y; }
    if (abs(gsl_part_7) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_7), gsl_partVanish_7.zw); }
  }
  let paintPairNut = gsl_part_7;
  let gsl_partLo_8 = vec3<f32>(-22.01, 11.371198, -0.01);
  let gsl_partHi_8 = vec3<f32>(-13.99, 20.628801, 4.71);
  let gsl_partBox_8 = gsl_boxDistance(p, gsl_partLo_8, gsl_partHi_8);
  var gsl_part_8 = gsl_partBox_8;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_8 <= gsl_vanishView.w) {
    let gsl_partVanish_8 = gsl_vanish(gsl_partLo_8, gsl_partHi_8);
    gsl_part_8 = 1e10;
    let gsl_simpleWeight_52 = gsl_simple(vec3<f32>(-22.01, 11.371198, -0.01), vec3<f32>(-13.99, 20.628801, 4.71));
    var gsl_simplePart_8 = 1e10;
    if (gsl_simpleWeight_52 > 0.0) {
      gsl_simplePart_8 = gsl_remap_10(p - vec3<f32>(-18.0, 16.0, 0.0));
    }
    if (gsl_simpleWeight_52 < 1.0) {
      gsl_simplePart_8 = mix(max(gsl_remap_11(p - vec3<f32>(-18.0, 16.0, 0.0)), gsl_boxSdf(p, vec3<f32>(-22.01, 11.371198, -0.01), vec3<f32>(-13.99, 20.628801, 4.71))), gsl_simplePart_8, gsl_simpleWeight_52);
    }
    if (gsl_partVanish_8.x > 0.0) { gsl_part_8 = gsl_simplePart_8 + gsl_partVanish_8.y; }
    if (abs(gsl_part_8) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_8), gsl_partVanish_8.zw); }
  }
  let paintNut = gsl_part_8;
  let gsl_partLo_9 = vec3<f32>(-10.26, -20.26, -0.01);
  let gsl_partHi_9 = vec3<f32>(-1.74, -11.74, 21.01);
  let gsl_partBox_9 = gsl_boxDistance(p, gsl_partLo_9, gsl_partHi_9);
  var gsl_part_9 = gsl_partBox_9;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_9 <= gsl_vanishView.w) {
    let gsl_partVanish_9 = gsl_vanish(gsl_partLo_9, gsl_partHi_9);
    gsl_part_9 = 1e10;
    let gsl_simpleWeight_53 = gsl_simple(vec3<f32>(-10.26, -20.26, -0.01), vec3<f32>(-1.74, -11.74, 21.01));
    var gsl_simplePart_9 = 1e10;
    if (gsl_simpleWeight_53 > 0.0) {
      gsl_simplePart_9 = gsl_remap_12(p - vec3<f32>(-6.0, -16.0, 16.0));
    }
    if (gsl_simpleWeight_53 < 1.0) {
      gsl_simplePart_9 = mix(max(gsl_remap_13(p - vec3<f32>(-6.0, -16.0, 16.0)), gsl_boxSdf(p, vec3<f32>(-10.26, -20.26, -0.01), vec3<f32>(-1.74, -11.74, 21.01))), gsl_simplePart_9, gsl_simpleWeight_53);
    }
    if (gsl_partVanish_9.x > 0.0) { gsl_part_9 = gsl_simplePart_9 + gsl_partVanish_9.y; }
    if (abs(gsl_part_9) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_9), gsl_partVanish_9.zw); }
  }
  let reliefScrew = gsl_part_9;
  let gsl_partLo_10 = vec3<f32>(-10.26, -4.26, -0.01);
  let gsl_partHi_10 = vec3<f32>(-1.74, 4.26, 21.01);
  let gsl_partBox_10 = gsl_boxDistance(p, gsl_partLo_10, gsl_partHi_10);
  var gsl_part_10 = gsl_partBox_10;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_10 <= gsl_vanishView.w) {
    let gsl_partVanish_10 = gsl_vanish(gsl_partLo_10, gsl_partHi_10);
    gsl_part_10 = 1e10;
    let gsl_simpleWeight_54 = gsl_simple(vec3<f32>(-10.26, -4.26, -0.01), vec3<f32>(-1.74, 4.26, 21.01));
    var gsl_simplePart_10 = 1e10;
    if (gsl_simpleWeight_54 > 0.0) {
      gsl_simplePart_10 = gsl_remap_14(p - vec3<f32>(-6.0, 0.0, 16.0));
    }
    if (gsl_simpleWeight_54 < 1.0) {
      gsl_simplePart_10 = mix(max(gsl_remap_15(p - vec3<f32>(-6.0, 0.0, 16.0)), gsl_boxSdf(p, vec3<f32>(-10.26, -4.26, -0.01), vec3<f32>(-1.74, 4.26, 21.01))), gsl_simplePart_10, gsl_simpleWeight_54);
    }
    if (gsl_partVanish_10.x > 0.0) { gsl_part_10 = gsl_simplePart_10 + gsl_partVanish_10.y; }
    if (abs(gsl_part_10) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_10), gsl_partVanish_10.zw); }
  }
  let reliefPairScrew = gsl_part_10;
  let gsl_partLo_11 = vec3<f32>(-10.01, -4.6288023, 3.19);
  let gsl_partHi_11 = vec3<f32>(-1.99, 4.6288023, 7.91);
  let gsl_partBox_11 = gsl_boxDistance(p, gsl_partLo_11, gsl_partHi_11);
  var gsl_part_11 = gsl_partBox_11;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_11 <= gsl_vanishView.w) {
    let gsl_partVanish_11 = gsl_vanish(gsl_partLo_11, gsl_partHi_11);
    gsl_part_11 = 1e10;
    let gsl_simpleWeight_55 = gsl_simple(vec3<f32>(-10.01, -4.6288023, 3.19), vec3<f32>(-1.99, 4.6288023, 7.91));
    var gsl_simplePart_11 = 1e10;
    if (gsl_simpleWeight_55 > 0.0) {
      gsl_simplePart_11 = gsl_remap_16(p - vec3<f32>(-6.0, 0.0, 3.2));
    }
    if (gsl_simpleWeight_55 < 1.0) {
      gsl_simplePart_11 = mix(max(gsl_remap_17(p - vec3<f32>(-6.0, 0.0, 3.2)), gsl_boxSdf(p, vec3<f32>(-10.01, -4.6288023, 3.19), vec3<f32>(-1.99, 4.6288023, 7.91))), gsl_simplePart_11, gsl_simpleWeight_55);
    }
    if (gsl_partVanish_11.x > 0.0) { gsl_part_11 = gsl_simplePart_11 + gsl_partVanish_11.y; }
    if (abs(gsl_part_11) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_11), gsl_partVanish_11.zw); }
  }
  let reliefPairNut = gsl_part_11;
  let gsl_partLo_12 = vec3<f32>(-10.01, 11.371198, -0.01);
  let gsl_partHi_12 = vec3<f32>(-1.99, 20.628801, 4.71);
  let gsl_partBox_12 = gsl_boxDistance(p, gsl_partLo_12, gsl_partHi_12);
  var gsl_part_12 = gsl_partBox_12;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_12 <= gsl_vanishView.w) {
    let gsl_partVanish_12 = gsl_vanish(gsl_partLo_12, gsl_partHi_12);
    gsl_part_12 = 1e10;
    let gsl_simpleWeight_56 = gsl_simple(vec3<f32>(-10.01, 11.371198, -0.01), vec3<f32>(-1.99, 20.628801, 4.71));
    var gsl_simplePart_12 = 1e10;
    if (gsl_simpleWeight_56 > 0.0) {
      gsl_simplePart_12 = gsl_remap_18(p - vec3<f32>(-6.0, 16.0, 0.0));
    }
    if (gsl_simpleWeight_56 < 1.0) {
      gsl_simplePart_12 = mix(max(gsl_remap_19(p - vec3<f32>(-6.0, 16.0, 0.0)), gsl_boxSdf(p, vec3<f32>(-10.01, 11.371198, -0.01), vec3<f32>(-1.99, 20.628801, 4.71))), gsl_simplePart_12, gsl_simpleWeight_56);
    }
    if (gsl_partVanish_12.x > 0.0) { gsl_part_12 = gsl_simplePart_12 + gsl_partVanish_12.y; }
    if (abs(gsl_part_12) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_12), gsl_partVanish_12.zw); }
  }
  let reliefNut = gsl_part_12;
  let gsl_partLo_13 = vec3<f32>(1.74, -20.26, -0.01);
  let gsl_partHi_13 = vec3<f32>(10.26, -11.74, 21.01);
  let gsl_partBox_13 = gsl_boxDistance(p, gsl_partLo_13, gsl_partHi_13);
  var gsl_part_13 = gsl_partBox_13;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_13 <= gsl_vanishView.w) {
    let gsl_partVanish_13 = gsl_vanish(gsl_partLo_13, gsl_partHi_13);
    gsl_part_13 = 1e10;
    let gsl_simpleWeight_57 = gsl_simple(vec3<f32>(1.74, -20.26, -0.01), vec3<f32>(10.26, -11.74, 21.01));
    var gsl_simplePart_13 = 1e10;
    if (gsl_simpleWeight_57 > 0.0) {
      gsl_simplePart_13 = gsl_remap_20(p - vec3<f32>(6.0, -16.0, 16.0));
    }
    if (gsl_simpleWeight_57 < 1.0) {
      gsl_simplePart_13 = mix(max(gsl_remap_21(p - vec3<f32>(6.0, -16.0, 16.0)), gsl_boxSdf(p, vec3<f32>(1.74, -20.26, -0.01), vec3<f32>(10.26, -11.74, 21.01))), gsl_simplePart_13, gsl_simpleWeight_57);
    }
    if (gsl_partVanish_13.x > 0.0) { gsl_part_13 = gsl_simplePart_13 + gsl_partVanish_13.y; }
    if (abs(gsl_part_13) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_13), gsl_partVanish_13.zw); }
  }
  let sinScrew = gsl_part_13;
  let gsl_partLo_14 = vec3<f32>(1.74, -4.26, -0.01);
  let gsl_partHi_14 = vec3<f32>(10.26, 4.26, 21.01);
  let gsl_partBox_14 = gsl_boxDistance(p, gsl_partLo_14, gsl_partHi_14);
  var gsl_part_14 = gsl_partBox_14;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_14 <= gsl_vanishView.w) {
    let gsl_partVanish_14 = gsl_vanish(gsl_partLo_14, gsl_partHi_14);
    gsl_part_14 = 1e10;
    let gsl_simpleWeight_58 = gsl_simple(vec3<f32>(1.74, -4.26, -0.01), vec3<f32>(10.26, 4.26, 21.01));
    var gsl_simplePart_14 = 1e10;
    if (gsl_simpleWeight_58 > 0.0) {
      gsl_simplePart_14 = gsl_remap_22(p - vec3<f32>(6.0, 0.0, 16.0));
    }
    if (gsl_simpleWeight_58 < 1.0) {
      gsl_simplePart_14 = mix(max(gsl_remap_23(p - vec3<f32>(6.0, 0.0, 16.0)), gsl_boxSdf(p, vec3<f32>(1.74, -4.26, -0.01), vec3<f32>(10.26, 4.26, 21.01))), gsl_simplePart_14, gsl_simpleWeight_58);
    }
    if (gsl_partVanish_14.x > 0.0) { gsl_part_14 = gsl_simplePart_14 + gsl_partVanish_14.y; }
    if (abs(gsl_part_14) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_14), gsl_partVanish_14.zw); }
  }
  let sinPairScrew = gsl_part_14;
  let gsl_partLo_15 = vec3<f32>(1.99, -4.6288023, 3.19);
  let gsl_partHi_15 = vec3<f32>(10.01, 4.6288023, 7.91);
  let gsl_partBox_15 = gsl_boxDistance(p, gsl_partLo_15, gsl_partHi_15);
  var gsl_part_15 = gsl_partBox_15;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_15 <= gsl_vanishView.w) {
    let gsl_partVanish_15 = gsl_vanish(gsl_partLo_15, gsl_partHi_15);
    gsl_part_15 = 1e10;
    let gsl_simpleWeight_59 = gsl_simple(vec3<f32>(1.99, -4.6288023, 3.19), vec3<f32>(10.01, 4.6288023, 7.91));
    var gsl_simplePart_15 = 1e10;
    if (gsl_simpleWeight_59 > 0.0) {
      gsl_simplePart_15 = gsl_remap_24(p - vec3<f32>(6.0, 0.0, 3.2));
    }
    if (gsl_simpleWeight_59 < 1.0) {
      gsl_simplePart_15 = mix(max(gsl_remap_25(p - vec3<f32>(6.0, 0.0, 3.2)), gsl_boxSdf(p, vec3<f32>(1.99, -4.6288023, 3.19), vec3<f32>(10.01, 4.6288023, 7.91))), gsl_simplePart_15, gsl_simpleWeight_59);
    }
    if (gsl_partVanish_15.x > 0.0) { gsl_part_15 = gsl_simplePart_15 + gsl_partVanish_15.y; }
    if (abs(gsl_part_15) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_15), gsl_partVanish_15.zw); }
  }
  let sinPairNut = gsl_part_15;
  let gsl_partLo_16 = vec3<f32>(1.99, 11.371198, -0.01);
  let gsl_partHi_16 = vec3<f32>(10.01, 20.628801, 4.71);
  let gsl_partBox_16 = gsl_boxDistance(p, gsl_partLo_16, gsl_partHi_16);
  var gsl_part_16 = gsl_partBox_16;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_16 <= gsl_vanishView.w) {
    let gsl_partVanish_16 = gsl_vanish(gsl_partLo_16, gsl_partHi_16);
    gsl_part_16 = 1e10;
    let gsl_simpleWeight_60 = gsl_simple(vec3<f32>(1.99, 11.371198, -0.01), vec3<f32>(10.01, 20.628801, 4.71));
    var gsl_simplePart_16 = 1e10;
    if (gsl_simpleWeight_60 > 0.0) {
      gsl_simplePart_16 = gsl_remap_26(p - vec3<f32>(6.0, 16.0, 0.0));
    }
    if (gsl_simpleWeight_60 < 1.0) {
      gsl_simplePart_16 = mix(max(gsl_remap_27(p - vec3<f32>(6.0, 16.0, 0.0)), gsl_boxSdf(p, vec3<f32>(1.99, 11.371198, -0.01), vec3<f32>(10.01, 20.628801, 4.71))), gsl_simplePart_16, gsl_simpleWeight_60);
    }
    if (gsl_partVanish_16.x > 0.0) { gsl_part_16 = gsl_simplePart_16 + gsl_partVanish_16.y; }
    if (abs(gsl_part_16) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_16), gsl_partVanish_16.zw); }
  }
  let sinNut = gsl_part_16;
  let gsl_partLo_17 = vec3<f32>(13.74, -20.26, -0.01);
  let gsl_partHi_17 = vec3<f32>(22.26, -11.74, 21.01);
  let gsl_partBox_17 = gsl_boxDistance(p, gsl_partLo_17, gsl_partHi_17);
  var gsl_part_17 = gsl_partBox_17;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_17 <= gsl_vanishView.w) {
    let gsl_partVanish_17 = gsl_vanish(gsl_partLo_17, gsl_partHi_17);
    gsl_part_17 = 1e10;
    let gsl_simpleWeight_61 = gsl_simple(vec3<f32>(13.74, -20.26, -0.01), vec3<f32>(22.26, -11.74, 21.01));
    var gsl_simplePart_17 = 1e10;
    if (gsl_simpleWeight_61 > 0.0) {
      gsl_simplePart_17 = gsl_remap_28(p - vec3<f32>(18.0, -16.0, 16.0));
    }
    if (gsl_simpleWeight_61 < 1.0) {
      gsl_simplePart_17 = mix(max(gsl_remap_29(p - vec3<f32>(18.0, -16.0, 16.0)), gsl_boxSdf(p, vec3<f32>(13.74, -20.26, -0.01), vec3<f32>(22.26, -11.74, 21.01))), gsl_simplePart_17, gsl_simpleWeight_61);
    }
    if (gsl_partVanish_17.x > 0.0) { gsl_part_17 = gsl_simplePart_17 + gsl_partVanish_17.y; }
    if (abs(gsl_part_17) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_17), gsl_partVanish_17.zw); }
  }
  let sinGeomScrew = gsl_part_17;
  let gsl_partLo_18 = vec3<f32>(13.74, -4.26, -0.01);
  let gsl_partHi_18 = vec3<f32>(22.26, 4.26, 21.01);
  let gsl_partBox_18 = gsl_boxDistance(p, gsl_partLo_18, gsl_partHi_18);
  var gsl_part_18 = gsl_partBox_18;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_18 <= gsl_vanishView.w) {
    let gsl_partVanish_18 = gsl_vanish(gsl_partLo_18, gsl_partHi_18);
    gsl_part_18 = 1e10;
    let gsl_simpleWeight_62 = gsl_simple(vec3<f32>(13.74, -4.26, -0.01), vec3<f32>(22.26, 4.26, 21.01));
    var gsl_simplePart_18 = 1e10;
    if (gsl_simpleWeight_62 > 0.0) {
      gsl_simplePart_18 = gsl_remap_30(p - vec3<f32>(18.0, 0.0, 16.0));
    }
    if (gsl_simpleWeight_62 < 1.0) {
      gsl_simplePart_18 = mix(max(gsl_remap_31(p - vec3<f32>(18.0, 0.0, 16.0)), gsl_boxSdf(p, vec3<f32>(13.74, -4.26, -0.01), vec3<f32>(22.26, 4.26, 21.01))), gsl_simplePart_18, gsl_simpleWeight_62);
    }
    if (gsl_partVanish_18.x > 0.0) { gsl_part_18 = gsl_simplePart_18 + gsl_partVanish_18.y; }
    if (abs(gsl_part_18) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_18), gsl_partVanish_18.zw); }
  }
  let sinGeomPairScrew = gsl_part_18;
  let gsl_partLo_19 = vec3<f32>(13.99, -4.6288023, 3.19);
  let gsl_partHi_19 = vec3<f32>(22.01, 4.6288023, 7.91);
  let gsl_partBox_19 = gsl_boxDistance(p, gsl_partLo_19, gsl_partHi_19);
  var gsl_part_19 = gsl_partBox_19;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_19 <= gsl_vanishView.w) {
    let gsl_partVanish_19 = gsl_vanish(gsl_partLo_19, gsl_partHi_19);
    gsl_part_19 = 1e10;
    let gsl_simpleWeight_63 = gsl_simple(vec3<f32>(13.99, -4.6288023, 3.19), vec3<f32>(22.01, 4.6288023, 7.91));
    var gsl_simplePart_19 = 1e10;
    if (gsl_simpleWeight_63 > 0.0) {
      gsl_simplePart_19 = gsl_remap_32(p - vec3<f32>(18.0, 0.0, 3.2));
    }
    if (gsl_simpleWeight_63 < 1.0) {
      gsl_simplePart_19 = mix(max(gsl_remap_33(p - vec3<f32>(18.0, 0.0, 3.2)), gsl_boxSdf(p, vec3<f32>(13.99, -4.6288023, 3.19), vec3<f32>(22.01, 4.6288023, 7.91))), gsl_simplePart_19, gsl_simpleWeight_63);
    }
    if (gsl_partVanish_19.x > 0.0) { gsl_part_19 = gsl_simplePart_19 + gsl_partVanish_19.y; }
    if (abs(gsl_part_19) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_19), gsl_partVanish_19.zw); }
  }
  let sinGeomPairNut = gsl_part_19;
  let gsl_partLo_20 = vec3<f32>(13.99, 11.371198, -0.01);
  let gsl_partHi_20 = vec3<f32>(22.01, 20.628801, 4.71);
  let gsl_partBox_20 = gsl_boxDistance(p, gsl_partLo_20, gsl_partHi_20);
  var gsl_part_20 = gsl_partBox_20;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_20 <= gsl_vanishView.w) {
    let gsl_partVanish_20 = gsl_vanish(gsl_partLo_20, gsl_partHi_20);
    gsl_part_20 = 1e10;
    let gsl_simpleWeight_64 = gsl_simple(vec3<f32>(13.99, 11.371198, -0.01), vec3<f32>(22.01, 20.628801, 4.71));
    var gsl_simplePart_20 = 1e10;
    if (gsl_simpleWeight_64 > 0.0) {
      gsl_simplePart_20 = gsl_remap_34(p - vec3<f32>(18.0, 16.0, 0.0));
    }
    if (gsl_simpleWeight_64 < 1.0) {
      gsl_simplePart_20 = mix(max(gsl_remap_35(p - vec3<f32>(18.0, 16.0, 0.0)), gsl_boxSdf(p, vec3<f32>(13.99, 11.371198, -0.01), vec3<f32>(22.01, 20.628801, 4.71))), gsl_simplePart_20, gsl_simpleWeight_64);
    }
    if (gsl_partVanish_20.x > 0.0) { gsl_part_20 = gsl_simplePart_20 + gsl_partVanish_20.y; }
    if (abs(gsl_part_20) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_20), gsl_partVanish_20.zw); }
  }
  let sinGeomNut = gsl_part_20;
  let gsl_partLo_21 = vec3<f32>(25.74, -20.26, -0.01);
  let gsl_partHi_21 = vec3<f32>(34.26, -11.74, 21.01);
  let gsl_partBox_21 = gsl_boxDistance(p, gsl_partLo_21, gsl_partHi_21);
  var gsl_part_21 = gsl_partBox_21;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_21 <= gsl_vanishView.w) {
    let gsl_partVanish_21 = gsl_vanish(gsl_partLo_21, gsl_partHi_21);
    gsl_part_21 = 1e10;
    let gsl_simpleWeight_65 = gsl_simple(vec3<f32>(25.74, -20.26, -0.01), vec3<f32>(34.26, -11.74, 21.01));
    var gsl_simplePart_21 = 1e10;
    if (gsl_simpleWeight_65 > 0.0) {
      gsl_simplePart_21 = gsl_remap_36(p - vec3<f32>(30.0, -16.0, 16.0));
    }
    if (gsl_simpleWeight_65 < 1.0) {
      gsl_simplePart_21 = mix(max(gsl_remap_37(p - vec3<f32>(30.0, -16.0, 16.0)), gsl_boxSdf(p, vec3<f32>(25.74, -20.26, -0.01), vec3<f32>(34.26, -11.74, 21.01))), gsl_simplePart_21, gsl_simpleWeight_65);
    }
    if (gsl_partVanish_21.x > 0.0) { gsl_part_21 = gsl_simplePart_21 + gsl_partVanish_21.y; }
    if (abs(gsl_part_21) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_21), gsl_partVanish_21.zw); }
  }
  let isoGeomScrew = gsl_part_21;
  let gsl_partLo_22 = vec3<f32>(25.74, -4.26, -0.01);
  let gsl_partHi_22 = vec3<f32>(34.26, 4.26, 21.01);
  let gsl_partBox_22 = gsl_boxDistance(p, gsl_partLo_22, gsl_partHi_22);
  var gsl_part_22 = gsl_partBox_22;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_22 <= gsl_vanishView.w) {
    let gsl_partVanish_22 = gsl_vanish(gsl_partLo_22, gsl_partHi_22);
    gsl_part_22 = 1e10;
    let gsl_simpleWeight_66 = gsl_simple(vec3<f32>(25.74, -4.26, -0.01), vec3<f32>(34.26, 4.26, 21.01));
    var gsl_simplePart_22 = 1e10;
    if (gsl_simpleWeight_66 > 0.0) {
      gsl_simplePart_22 = gsl_remap_38(p - vec3<f32>(30.0, 0.0, 16.0));
    }
    if (gsl_simpleWeight_66 < 1.0) {
      gsl_simplePart_22 = mix(max(gsl_remap_39(p - vec3<f32>(30.0, 0.0, 16.0)), gsl_boxSdf(p, vec3<f32>(25.74, -4.26, -0.01), vec3<f32>(34.26, 4.26, 21.01))), gsl_simplePart_22, gsl_simpleWeight_66);
    }
    if (gsl_partVanish_22.x > 0.0) { gsl_part_22 = gsl_simplePart_22 + gsl_partVanish_22.y; }
    if (abs(gsl_part_22) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_22), gsl_partVanish_22.zw); }
  }
  let isoGeomPairScrew = gsl_part_22;
  let gsl_partLo_23 = vec3<f32>(25.99, -4.6288023, 3.19);
  let gsl_partHi_23 = vec3<f32>(34.01, 4.6288023, 7.91);
  let gsl_partBox_23 = gsl_boxDistance(p, gsl_partLo_23, gsl_partHi_23);
  var gsl_part_23 = gsl_partBox_23;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_23 <= gsl_vanishView.w) {
    let gsl_partVanish_23 = gsl_vanish(gsl_partLo_23, gsl_partHi_23);
    gsl_part_23 = 1e10;
    let gsl_simpleWeight_67 = gsl_simple(vec3<f32>(25.99, -4.6288023, 3.19), vec3<f32>(34.01, 4.6288023, 7.91));
    var gsl_simplePart_23 = 1e10;
    if (gsl_simpleWeight_67 > 0.0) {
      gsl_simplePart_23 = gsl_remap_40(p - vec3<f32>(30.0, 0.0, 3.2));
    }
    if (gsl_simpleWeight_67 < 1.0) {
      gsl_simplePart_23 = mix(max(gsl_remap_41(p - vec3<f32>(30.0, 0.0, 3.2)), gsl_boxSdf(p, vec3<f32>(25.99, -4.6288023, 3.19), vec3<f32>(34.01, 4.6288023, 7.91))), gsl_simplePart_23, gsl_simpleWeight_67);
    }
    if (gsl_partVanish_23.x > 0.0) { gsl_part_23 = gsl_simplePart_23 + gsl_partVanish_23.y; }
    if (abs(gsl_part_23) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_23), gsl_partVanish_23.zw); }
  }
  let isoGeomPairNut = gsl_part_23;
  let gsl_partLo_24 = vec3<f32>(25.99, 11.371198, -0.01);
  let gsl_partHi_24 = vec3<f32>(34.01, 20.628801, 4.71);
  let gsl_partBox_24 = gsl_boxDistance(p, gsl_partLo_24, gsl_partHi_24);
  var gsl_part_24 = gsl_partBox_24;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_24 <= gsl_vanishView.w) {
    let gsl_partVanish_24 = gsl_vanish(gsl_partLo_24, gsl_partHi_24);
    gsl_part_24 = 1e10;
    let gsl_simpleWeight_68 = gsl_simple(vec3<f32>(25.99, 11.371198, -0.01), vec3<f32>(34.01, 20.628801, 4.71));
    var gsl_simplePart_24 = 1e10;
    if (gsl_simpleWeight_68 > 0.0) {
      gsl_simplePart_24 = gsl_remap_42(p - vec3<f32>(30.0, 16.0, 0.0));
    }
    if (gsl_simpleWeight_68 < 1.0) {
      gsl_simplePart_24 = mix(max(gsl_remap_43(p - vec3<f32>(30.0, 16.0, 0.0)), gsl_boxSdf(p, vec3<f32>(25.99, 11.371198, -0.01), vec3<f32>(34.01, 20.628801, 4.71))), gsl_simplePart_24, gsl_simpleWeight_68);
    }
    if (gsl_partVanish_24.x > 0.0) { gsl_part_24 = gsl_simplePart_24 + gsl_partVanish_24.y; }
    if (abs(gsl_part_24) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_24), gsl_partVanish_24.zw); }
  }
  let isoGeomNut = gsl_part_24;
  return array<f32, 24>(plainScrew, plainPairScrew, plainPairNut, plainNut, paintScrew, paintPairScrew, paintPairNut, paintNut, reliefScrew, reliefPairScrew, reliefPairNut, reliefNut, sinScrew, sinPairScrew, sinPairNut, sinNut, sinGeomScrew, sinGeomPairScrew, sinGeomPairNut, sinGeomNut, isoGeomScrew, isoGeomPairScrew, isoGeomPairNut, isoGeomNut);
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let gsl_d = gsl_shapeParts(p, time);
  let gsl_s1 = gsl_shape(gsl_d[0], 1u);
  let gsl_s2 = gsl_shape(gsl_d[1], 2u);
  let gsl_s3 = gsl_shape(gsl_d[2], 3u);
  let gsl_s4 = gsl_shape(gsl_d[3], 4u);
  let gsl_s5 = gsl_shape(gsl_d[4], 5u);
  let gsl_s6 = gsl_shape(gsl_d[5], 6u);
  let gsl_s7 = gsl_shape(gsl_d[6], 7u);
  let gsl_s8 = gsl_shape(gsl_d[7], 8u);
  let gsl_s9 = gsl_shape(gsl_d[8], 9u);
  let gsl_s10 = gsl_shape(gsl_d[9], 10u);
  let gsl_s11 = gsl_shape(gsl_d[10], 11u);
  let gsl_s12 = gsl_shape(gsl_d[11], 12u);
  let gsl_s13 = gsl_shape(gsl_d[12], 13u);
  let gsl_s14 = gsl_shape(gsl_d[13], 14u);
  let gsl_s15 = gsl_shape(gsl_d[14], 15u);
  let gsl_s16 = gsl_shape(gsl_d[15], 16u);
  let gsl_s17 = gsl_shape(gsl_d[16], 17u);
  let gsl_s18 = gsl_shape(gsl_d[17], 18u);
  let gsl_s19 = gsl_shape(gsl_d[18], 19u);
  let gsl_s20 = gsl_shape(gsl_d[19], 20u);
  let gsl_s21 = gsl_shape(gsl_d[20], 21u);
  let gsl_s22 = gsl_shape(gsl_d[21], 22u);
  let gsl_s23 = gsl_shape(gsl_d[22], 23u);
  let gsl_s24 = gsl_shape(gsl_d[23], 24u);
  return gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_s1, gsl_s2), gsl_s3), gsl_s4), gsl_s5), gsl_s6), gsl_s7), gsl_s8), gsl_s9), gsl_s10), gsl_s11), gsl_s12), gsl_s13), gsl_s14), gsl_s15), gsl_s16), gsl_s17), gsl_s18), gsl_s19), gsl_s20), gsl_s21), gsl_s22), gsl_s23), gsl_s24);
}
fn gsl_colorField(p: vec3<f32>, time: f32, gsl_colorIdx: u32) -> vec4<f32> {
  if (gsl_colorIdx == 1u) { return vec4<f32>(vec3<f32>(0.62, 0.64, 0.68), 1.0); }
  if (gsl_colorIdx == 2u) { return vec4<f32>(vec3<f32>(0.62, 0.64, 0.68), 1.0); }
  if (gsl_colorIdx == 3u) { return vec4<f32>(vec3<f32>(0.62, 0.64, 0.68), 1.0); }
  if (gsl_colorIdx == 4u) { return vec4<f32>(vec3<f32>(0.62, 0.64, 0.68), 1.0); }
  if (gsl_colorIdx == 5u) {
    let gsl_simpleWeight = gsl_simple(vec3<f32>(-22.26, -20.26, -0.01), vec3<f32>(-13.74, -11.74, 21.01));
    var gsl_simpleField = clamp(gsl_remap_45(p - vec3<f32>(-18.0, -16.0, 16.0)) * 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0));
    if (gsl_simpleWeight > 0.0) {
      gsl_simpleField = mix(gsl_simpleField, clamp(gsl_remap_44(p - vec3<f32>(-18.0, -16.0, 16.0)) * 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0)), gsl_simpleWeight);
    }
    return vec4<f32>(gsl_simpleField, 1.0);
  }
  if (gsl_colorIdx == 6u) {
    let gsl_simpleWeight_2 = gsl_simple(vec3<f32>(-22.26, -4.26, -0.01), vec3<f32>(-13.74, 4.26, 21.01));
    var gsl_simpleField_2 = clamp(gsl_remap_45(p - vec3<f32>(-18.0, 0.0, 16.0)) * 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0));
    if (gsl_simpleWeight_2 > 0.0) {
      gsl_simpleField_2 = mix(gsl_simpleField_2, clamp(gsl_remap_46(p - vec3<f32>(-18.0, 0.0, 16.0)) * 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0)), gsl_simpleWeight_2);
    }
    return vec4<f32>(gsl_simpleField_2, 1.0);
  }
  if (gsl_colorIdx == 7u) {
    let gsl_simpleWeight_3 = gsl_simple(vec3<f32>(-22.01, -4.6288023, 3.19), vec3<f32>(-13.99, 4.6288023, 7.91));
    var gsl_simpleField_3 = clamp(gsl_remap_45(p - vec3<f32>(-18.0, 0.0, 3.2)) * 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0));
    if (gsl_simpleWeight_3 > 0.0) {
      gsl_simpleField_3 = mix(gsl_simpleField_3, clamp(gsl_remap_47(p - vec3<f32>(-18.0, 0.0, 3.2)) * 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0)), gsl_simpleWeight_3);
    }
    return vec4<f32>(gsl_simpleField_3, 1.0);
  }
  if (gsl_colorIdx == 8u) {
    let gsl_simpleWeight_4 = gsl_simple(vec3<f32>(-22.01, 11.371198, -0.01), vec3<f32>(-13.99, 20.628801, 4.71));
    var gsl_simpleField_4 = clamp(gsl_remap_45(p - vec3<f32>(-18.0, 16.0, 0.0)) * 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0));
    if (gsl_simpleWeight_4 > 0.0) {
      gsl_simpleField_4 = mix(gsl_simpleField_4, clamp(gsl_remap_48(p - vec3<f32>(-18.0, 16.0, 0.0)) * 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0)), gsl_simpleWeight_4);
    }
    return vec4<f32>(gsl_simpleField_4, 1.0);
  }
  if (gsl_colorIdx == 9u) {
    let gsl_simpleWeight_5 = gsl_simple(vec3<f32>(-10.26, -20.26, -0.01), vec3<f32>(-1.74, -11.74, 21.01));
    var gsl_simpleField_5 = clamp(gsl_remap_50(p - vec3<f32>(-6.0, -16.0, 16.0)) * 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0));
    if (gsl_simpleWeight_5 > 0.0) {
      gsl_simpleField_5 = mix(gsl_simpleField_5, clamp(gsl_remap_49(p - vec3<f32>(-6.0, -16.0, 16.0)) * 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0)), gsl_simpleWeight_5);
    }
    return vec4<f32>(gsl_simpleField_5, 1.0);
  }
  if (gsl_colorIdx == 10u) {
    let gsl_simpleWeight_6 = gsl_simple(vec3<f32>(-10.26, -4.26, -0.01), vec3<f32>(-1.74, 4.26, 21.01));
    var gsl_simpleField_6 = clamp(gsl_remap_52(p - vec3<f32>(-6.0, 0.0, 16.0)) * 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0));
    if (gsl_simpleWeight_6 > 0.0) {
      gsl_simpleField_6 = mix(gsl_simpleField_6, clamp(gsl_remap_51(p - vec3<f32>(-6.0, 0.0, 16.0)) * 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0)), gsl_simpleWeight_6);
    }
    return vec4<f32>(gsl_simpleField_6, 1.0);
  }
  if (gsl_colorIdx == 11u) {
    let gsl_simpleWeight_7 = gsl_simple(vec3<f32>(-10.01, -4.6288023, 3.19), vec3<f32>(-1.99, 4.6288023, 7.91));
    var gsl_simpleField_7 = clamp(gsl_remap_54(p - vec3<f32>(-6.0, 0.0, 3.2)) * 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0));
    if (gsl_simpleWeight_7 > 0.0) {
      gsl_simpleField_7 = mix(gsl_simpleField_7, clamp(gsl_remap_53(p - vec3<f32>(-6.0, 0.0, 3.2)) * 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0)), gsl_simpleWeight_7);
    }
    return vec4<f32>(gsl_simpleField_7, 1.0);
  }
  if (gsl_colorIdx == 12u) {
    let gsl_simpleWeight_8 = gsl_simple(vec3<f32>(-10.01, 11.371198, -0.01), vec3<f32>(-1.99, 20.628801, 4.71));
    var gsl_simpleField_8 = clamp(gsl_remap_56(p - vec3<f32>(-6.0, 16.0, 0.0)) * 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0));
    if (gsl_simpleWeight_8 > 0.0) {
      gsl_simpleField_8 = mix(gsl_simpleField_8, clamp(gsl_remap_55(p - vec3<f32>(-6.0, 16.0, 0.0)) * 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0)), gsl_simpleWeight_8);
    }
    return vec4<f32>(gsl_simpleField_8, 1.0);
  }
  if (gsl_colorIdx == 13u) {
    let gsl_simpleWeight_9 = gsl_simple(vec3<f32>(1.74, -20.26, -0.01), vec3<f32>(10.26, -11.74, 21.01));
    var gsl_simpleField_9 = clamp(gsl_remap_58(p - vec3<f32>(6.0, -16.0, 16.0)) * 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0));
    if (gsl_simpleWeight_9 > 0.0) {
      gsl_simpleField_9 = mix(gsl_simpleField_9, clamp(gsl_remap_57(p - vec3<f32>(6.0, -16.0, 16.0)) * 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0)), gsl_simpleWeight_9);
    }
    return vec4<f32>(gsl_simpleField_9, 1.0);
  }
  if (gsl_colorIdx == 14u) {
    let gsl_simpleWeight_10 = gsl_simple(vec3<f32>(1.74, -4.26, -0.01), vec3<f32>(10.26, 4.26, 21.01));
    var gsl_simpleField_10 = clamp(gsl_remap_60(p - vec3<f32>(6.0, 0.0, 16.0)) * 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0));
    if (gsl_simpleWeight_10 > 0.0) {
      gsl_simpleField_10 = mix(gsl_simpleField_10, clamp(gsl_remap_59(p - vec3<f32>(6.0, 0.0, 16.0)) * 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0)), gsl_simpleWeight_10);
    }
    return vec4<f32>(gsl_simpleField_10, 1.0);
  }
  if (gsl_colorIdx == 15u) {
    let gsl_simpleWeight_11 = gsl_simple(vec3<f32>(1.99, -4.6288023, 3.19), vec3<f32>(10.01, 4.6288023, 7.91));
    var gsl_simpleField_11 = clamp(gsl_remap_62(p - vec3<f32>(6.0, 0.0, 3.2)) * 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0));
    if (gsl_simpleWeight_11 > 0.0) {
      gsl_simpleField_11 = mix(gsl_simpleField_11, clamp(gsl_remap_61(p - vec3<f32>(6.0, 0.0, 3.2)) * 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0)), gsl_simpleWeight_11);
    }
    return vec4<f32>(gsl_simpleField_11, 1.0);
  }
  if (gsl_colorIdx == 16u) {
    let gsl_simpleWeight_12 = gsl_simple(vec3<f32>(1.99, 11.371198, -0.01), vec3<f32>(10.01, 20.628801, 4.71));
    var gsl_simpleField_12 = clamp(gsl_remap_64(p - vec3<f32>(6.0, 16.0, 0.0)) * 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0));
    if (gsl_simpleWeight_12 > 0.0) {
      gsl_simpleField_12 = mix(gsl_simpleField_12, clamp(gsl_remap_63(p - vec3<f32>(6.0, 16.0, 0.0)) * 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0)), gsl_simpleWeight_12);
    }
    return vec4<f32>(gsl_simpleField_12, 1.0);
  }
  if (gsl_colorIdx == 17u) {
    let gsl_simpleWeight_13 = gsl_simple(vec3<f32>(13.74, -20.26, -0.01), vec3<f32>(22.26, -11.74, 21.01));
    var gsl_simpleField_13 = clamp(gsl_remap_65(p - vec3<f32>(18.0, -16.0, 16.0)) * 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0));
    if (gsl_simpleWeight_13 > 0.0) {
      gsl_simpleField_13 = mix(gsl_simpleField_13, clamp(gsl_remap_45(p - vec3<f32>(18.0, -16.0, 16.0)) * 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0)), gsl_simpleWeight_13);
    }
    return vec4<f32>(gsl_simpleField_13, 1.0);
  }
  if (gsl_colorIdx == 18u) {
    let gsl_simpleWeight_14 = gsl_simple(vec3<f32>(13.74, -4.26, -0.01), vec3<f32>(22.26, 4.26, 21.01));
    var gsl_simpleField_14 = clamp(gsl_remap_66(p - vec3<f32>(18.0, 0.0, 16.0)) * 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0));
    if (gsl_simpleWeight_14 > 0.0) {
      gsl_simpleField_14 = mix(gsl_simpleField_14, clamp(gsl_remap_45(p - vec3<f32>(18.0, 0.0, 16.0)) * 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0)), gsl_simpleWeight_14);
    }
    return vec4<f32>(gsl_simpleField_14, 1.0);
  }
  if (gsl_colorIdx == 19u) {
    let gsl_simpleWeight_15 = gsl_simple(vec3<f32>(13.99, -4.6288023, 3.19), vec3<f32>(22.01, 4.6288023, 7.91));
    var gsl_simpleField_15 = clamp(gsl_remap_67(p - vec3<f32>(18.0, 0.0, 3.2)) * 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0));
    if (gsl_simpleWeight_15 > 0.0) {
      gsl_simpleField_15 = mix(gsl_simpleField_15, clamp(gsl_remap_45(p - vec3<f32>(18.0, 0.0, 3.2)) * 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0)), gsl_simpleWeight_15);
    }
    return vec4<f32>(gsl_simpleField_15, 1.0);
  }
  if (gsl_colorIdx == 20u) {
    let gsl_simpleWeight_16 = gsl_simple(vec3<f32>(13.99, 11.371198, -0.01), vec3<f32>(22.01, 20.628801, 4.71));
    var gsl_simpleField_16 = clamp(gsl_remap_68(p - vec3<f32>(18.0, 16.0, 0.0)) * 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0));
    if (gsl_simpleWeight_16 > 0.0) {
      gsl_simpleField_16 = mix(gsl_simpleField_16, clamp(gsl_remap_45(p - vec3<f32>(18.0, 16.0, 0.0)) * 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0)), gsl_simpleWeight_16);
    }
    return vec4<f32>(gsl_simpleField_16, 1.0);
  }
  if (gsl_colorIdx == 21u) {
    let gsl_simpleWeight_17 = gsl_simple(vec3<f32>(25.74, -20.26, -0.01), vec3<f32>(34.26, -11.74, 21.01));
    var gsl_simpleField_17 = clamp(gsl_remap_69(p - vec3<f32>(30.0, -16.0, 16.0)) * 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0));
    if (gsl_simpleWeight_17 > 0.0) {
      gsl_simpleField_17 = mix(gsl_simpleField_17, clamp(gsl_remap_45(p - vec3<f32>(30.0, -16.0, 16.0)) * 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0)), gsl_simpleWeight_17);
    }
    return vec4<f32>(gsl_simpleField_17, 1.0);
  }
  if (gsl_colorIdx == 22u) {
    let gsl_simpleWeight_18 = gsl_simple(vec3<f32>(25.74, -4.26, -0.01), vec3<f32>(34.26, 4.26, 21.01));
    var gsl_simpleField_18 = clamp(gsl_remap_70(p - vec3<f32>(30.0, 0.0, 16.0)) * 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0));
    if (gsl_simpleWeight_18 > 0.0) {
      gsl_simpleField_18 = mix(gsl_simpleField_18, clamp(gsl_remap_45(p - vec3<f32>(30.0, 0.0, 16.0)) * 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0)), gsl_simpleWeight_18);
    }
    return vec4<f32>(gsl_simpleField_18, 1.0);
  }
  if (gsl_colorIdx == 23u) {
    let gsl_simpleWeight_19 = gsl_simple(vec3<f32>(25.99, -4.6288023, 3.19), vec3<f32>(34.01, 4.6288023, 7.91));
    var gsl_simpleField_19 = clamp(gsl_remap_71(p - vec3<f32>(30.0, 0.0, 3.2)) * 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0));
    if (gsl_simpleWeight_19 > 0.0) {
      gsl_simpleField_19 = mix(gsl_simpleField_19, clamp(gsl_remap_45(p - vec3<f32>(30.0, 0.0, 3.2)) * 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0)), gsl_simpleWeight_19);
    }
    return vec4<f32>(gsl_simpleField_19, 1.0);
  }
  if (gsl_colorIdx == 24u) {
    let gsl_simpleWeight_20 = gsl_simple(vec3<f32>(25.99, 11.371198, -0.01), vec3<f32>(34.01, 20.628801, 4.71));
    var gsl_simpleField_20 = clamp(gsl_remap_72(p - vec3<f32>(30.0, 16.0, 0.0)) * 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0));
    if (gsl_simpleWeight_20 > 0.0) {
      gsl_simpleField_20 = mix(gsl_simpleField_20, clamp(gsl_remap_45(p - vec3<f32>(30.0, 16.0, 0.0)) * 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0)), gsl_simpleWeight_20);
    }
    return vec4<f32>(gsl_simpleField_20, 1.0);
  }
  return vec4<f32>(0.0, 0.0, 0.0, 0.0);
}
fn gsl_materialField(p: vec3<f32>, time: f32, gsl_materialIdx: u32) -> gsl_material {
  if (gsl_materialIdx == 5u) {
    let gsl_simpleWeight_21 = gsl_simple(vec3<f32>(-22.26, -20.26, -0.01), vec3<f32>(-13.74, -11.74, 21.01));
    var gsl_simpleField_21 = mat4x4<f32>(clamp(gsl_remap_73(p - vec3<f32>(-18.0, -16.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_74(p - vec3<f32>(-18.0, -16.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_74(p - vec3<f32>(-18.0, -16.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    if (gsl_simpleWeight_21 > 0.0) {
      let gsl_simpleFull = mat4x4<f32>(clamp(gsl_remap_73(p - vec3<f32>(-18.0, -16.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_74(p - vec3<f32>(-18.0, -16.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_74(p - vec3<f32>(-18.0, -16.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
      gsl_simpleField_21 = mat4x4<f32>(mix(gsl_simpleField_21[0], gsl_simpleFull[0], gsl_simpleWeight_21), mix(gsl_simpleField_21[1], gsl_simpleFull[1], gsl_simpleWeight_21), mix(gsl_simpleField_21[2], gsl_simpleFull[2], gsl_simpleWeight_21), mix(gsl_simpleField_21[3], gsl_simpleFull[3], gsl_simpleWeight_21));
    }
    let gsl_m = gsl_simpleField_21; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 6u) {
    let gsl_simpleWeight_22 = gsl_simple(vec3<f32>(-22.26, -4.26, -0.01), vec3<f32>(-13.74, 4.26, 21.01));
    var gsl_simpleField_22 = mat4x4<f32>(clamp(gsl_remap_73(p - vec3<f32>(-18.0, 0.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_74(p - vec3<f32>(-18.0, 0.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_74(p - vec3<f32>(-18.0, 0.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    if (gsl_simpleWeight_22 > 0.0) {
      let gsl_simpleFull_2 = mat4x4<f32>(clamp(gsl_remap_73(p - vec3<f32>(-18.0, 0.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_74(p - vec3<f32>(-18.0, 0.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_74(p - vec3<f32>(-18.0, 0.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
      gsl_simpleField_22 = mat4x4<f32>(mix(gsl_simpleField_22[0], gsl_simpleFull_2[0], gsl_simpleWeight_22), mix(gsl_simpleField_22[1], gsl_simpleFull_2[1], gsl_simpleWeight_22), mix(gsl_simpleField_22[2], gsl_simpleFull_2[2], gsl_simpleWeight_22), mix(gsl_simpleField_22[3], gsl_simpleFull_2[3], gsl_simpleWeight_22));
    }
    let gsl_m = gsl_simpleField_22; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 7u) {
    let gsl_simpleWeight_23 = gsl_simple(vec3<f32>(-22.01, -4.6288023, 3.19), vec3<f32>(-13.99, 4.6288023, 7.91));
    var gsl_simpleField_23 = mat4x4<f32>(clamp(gsl_remap_73(p - vec3<f32>(-18.0, 0.0, 3.2)) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_74(p - vec3<f32>(-18.0, 0.0, 3.2)) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_74(p - vec3<f32>(-18.0, 0.0, 3.2)) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    if (gsl_simpleWeight_23 > 0.0) {
      let gsl_simpleFull_3 = mat4x4<f32>(clamp(gsl_remap_73(p - vec3<f32>(-18.0, 0.0, 3.2)) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_74(p - vec3<f32>(-18.0, 0.0, 3.2)) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_74(p - vec3<f32>(-18.0, 0.0, 3.2)) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
      gsl_simpleField_23 = mat4x4<f32>(mix(gsl_simpleField_23[0], gsl_simpleFull_3[0], gsl_simpleWeight_23), mix(gsl_simpleField_23[1], gsl_simpleFull_3[1], gsl_simpleWeight_23), mix(gsl_simpleField_23[2], gsl_simpleFull_3[2], gsl_simpleWeight_23), mix(gsl_simpleField_23[3], gsl_simpleFull_3[3], gsl_simpleWeight_23));
    }
    let gsl_m = gsl_simpleField_23; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 8u) {
    let gsl_simpleWeight_24 = gsl_simple(vec3<f32>(-22.01, 11.371198, -0.01), vec3<f32>(-13.99, 20.628801, 4.71));
    var gsl_simpleField_24 = mat4x4<f32>(clamp(gsl_remap_73(p - vec3<f32>(-18.0, 16.0, 0.0)) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_74(p - vec3<f32>(-18.0, 16.0, 0.0)) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_74(p - vec3<f32>(-18.0, 16.0, 0.0)) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    if (gsl_simpleWeight_24 > 0.0) {
      let gsl_simpleFull_4 = mat4x4<f32>(clamp(gsl_remap_73(p - vec3<f32>(-18.0, 16.0, 0.0)) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_74(p - vec3<f32>(-18.0, 16.0, 0.0)) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_74(p - vec3<f32>(-18.0, 16.0, 0.0)) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
      gsl_simpleField_24 = mat4x4<f32>(mix(gsl_simpleField_24[0], gsl_simpleFull_4[0], gsl_simpleWeight_24), mix(gsl_simpleField_24[1], gsl_simpleFull_4[1], gsl_simpleWeight_24), mix(gsl_simpleField_24[2], gsl_simpleFull_4[2], gsl_simpleWeight_24), mix(gsl_simpleField_24[3], gsl_simpleFull_4[3], gsl_simpleWeight_24));
    }
    let gsl_m = gsl_simpleField_24; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 9u) {
    let gsl_simpleWeight_25 = gsl_simple(vec3<f32>(-10.26, -20.26, -0.01), vec3<f32>(-1.74, -11.74, 21.01));
    var gsl_simpleField_25 = mat4x4<f32>(clamp(gsl_remap_73(p - vec3<f32>(-6.0, -16.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_74(p - vec3<f32>(-6.0, -16.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_74(p - vec3<f32>(-6.0, -16.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    if (gsl_simpleWeight_25 > 0.0) {
      let gsl_simpleFull_5 = mat4x4<f32>(clamp(gsl_remap_75(p - vec3<f32>(-6.0, -16.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_76(p - vec3<f32>(-6.0, -16.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_77(p - vec3<f32>(-6.0, -16.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
      gsl_simpleField_25 = mat4x4<f32>(mix(gsl_simpleField_25[0], gsl_simpleFull_5[0], gsl_simpleWeight_25), mix(gsl_simpleField_25[1], gsl_simpleFull_5[1], gsl_simpleWeight_25), mix(gsl_simpleField_25[2], gsl_simpleFull_5[2], gsl_simpleWeight_25), mix(gsl_simpleField_25[3], gsl_simpleFull_5[3], gsl_simpleWeight_25));
    }
    let gsl_m = gsl_simpleField_25; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 10u) {
    let gsl_simpleWeight_26 = gsl_simple(vec3<f32>(-10.26, -4.26, -0.01), vec3<f32>(-1.74, 4.26, 21.01));
    var gsl_simpleField_26 = mat4x4<f32>(clamp(gsl_remap_73(p - vec3<f32>(-6.0, 0.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_74(p - vec3<f32>(-6.0, 0.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_74(p - vec3<f32>(-6.0, 0.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    if (gsl_simpleWeight_26 > 0.0) {
      let gsl_simpleFull_6 = mat4x4<f32>(clamp(gsl_remap_78(p - vec3<f32>(-6.0, 0.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_79(p - vec3<f32>(-6.0, 0.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_80(p - vec3<f32>(-6.0, 0.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
      gsl_simpleField_26 = mat4x4<f32>(mix(gsl_simpleField_26[0], gsl_simpleFull_6[0], gsl_simpleWeight_26), mix(gsl_simpleField_26[1], gsl_simpleFull_6[1], gsl_simpleWeight_26), mix(gsl_simpleField_26[2], gsl_simpleFull_6[2], gsl_simpleWeight_26), mix(gsl_simpleField_26[3], gsl_simpleFull_6[3], gsl_simpleWeight_26));
    }
    let gsl_m = gsl_simpleField_26; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 11u) {
    let gsl_simpleWeight_27 = gsl_simple(vec3<f32>(-10.01, -4.6288023, 3.19), vec3<f32>(-1.99, 4.6288023, 7.91));
    var gsl_simpleField_27 = mat4x4<f32>(clamp(gsl_remap_73(p - vec3<f32>(-6.0, 0.0, 3.2)) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_74(p - vec3<f32>(-6.0, 0.0, 3.2)) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_74(p - vec3<f32>(-6.0, 0.0, 3.2)) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    if (gsl_simpleWeight_27 > 0.0) {
      let gsl_simpleFull_7 = mat4x4<f32>(clamp(gsl_remap_81(p - vec3<f32>(-6.0, 0.0, 3.2)) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_82(p - vec3<f32>(-6.0, 0.0, 3.2)) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_83(p - vec3<f32>(-6.0, 0.0, 3.2)) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
      gsl_simpleField_27 = mat4x4<f32>(mix(gsl_simpleField_27[0], gsl_simpleFull_7[0], gsl_simpleWeight_27), mix(gsl_simpleField_27[1], gsl_simpleFull_7[1], gsl_simpleWeight_27), mix(gsl_simpleField_27[2], gsl_simpleFull_7[2], gsl_simpleWeight_27), mix(gsl_simpleField_27[3], gsl_simpleFull_7[3], gsl_simpleWeight_27));
    }
    let gsl_m = gsl_simpleField_27; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 12u) {
    let gsl_simpleWeight_28 = gsl_simple(vec3<f32>(-10.01, 11.371198, -0.01), vec3<f32>(-1.99, 20.628801, 4.71));
    var gsl_simpleField_28 = mat4x4<f32>(clamp(gsl_remap_73(p - vec3<f32>(-6.0, 16.0, 0.0)) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_74(p - vec3<f32>(-6.0, 16.0, 0.0)) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_74(p - vec3<f32>(-6.0, 16.0, 0.0)) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    if (gsl_simpleWeight_28 > 0.0) {
      let gsl_simpleFull_8 = mat4x4<f32>(clamp(gsl_remap_84(p - vec3<f32>(-6.0, 16.0, 0.0)) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_85(p - vec3<f32>(-6.0, 16.0, 0.0)) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_86(p - vec3<f32>(-6.0, 16.0, 0.0)) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
      gsl_simpleField_28 = mat4x4<f32>(mix(gsl_simpleField_28[0], gsl_simpleFull_8[0], gsl_simpleWeight_28), mix(gsl_simpleField_28[1], gsl_simpleFull_8[1], gsl_simpleWeight_28), mix(gsl_simpleField_28[2], gsl_simpleFull_8[2], gsl_simpleWeight_28), mix(gsl_simpleField_28[3], gsl_simpleFull_8[3], gsl_simpleWeight_28));
    }
    let gsl_m = gsl_simpleField_28; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 13u) {
    let gsl_simpleWeight_29 = gsl_simple(vec3<f32>(1.74, -20.26, -0.01), vec3<f32>(10.26, -11.74, 21.01));
    var gsl_simpleField_29 = mat4x4<f32>(clamp(gsl_remap_73(p - vec3<f32>(6.0, -16.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_74(p - vec3<f32>(6.0, -16.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_74(p - vec3<f32>(6.0, -16.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    if (gsl_simpleWeight_29 > 0.0) {
      let gsl_simpleFull_9 = mat4x4<f32>(clamp(gsl_remap_87(p - vec3<f32>(6.0, -16.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_88(p - vec3<f32>(6.0, -16.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_89(p - vec3<f32>(6.0, -16.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
      gsl_simpleField_29 = mat4x4<f32>(mix(gsl_simpleField_29[0], gsl_simpleFull_9[0], gsl_simpleWeight_29), mix(gsl_simpleField_29[1], gsl_simpleFull_9[1], gsl_simpleWeight_29), mix(gsl_simpleField_29[2], gsl_simpleFull_9[2], gsl_simpleWeight_29), mix(gsl_simpleField_29[3], gsl_simpleFull_9[3], gsl_simpleWeight_29));
    }
    let gsl_m = gsl_simpleField_29; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 14u) {
    let gsl_simpleWeight_30 = gsl_simple(vec3<f32>(1.74, -4.26, -0.01), vec3<f32>(10.26, 4.26, 21.01));
    var gsl_simpleField_30 = mat4x4<f32>(clamp(gsl_remap_73(p - vec3<f32>(6.0, 0.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_74(p - vec3<f32>(6.0, 0.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_74(p - vec3<f32>(6.0, 0.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    if (gsl_simpleWeight_30 > 0.0) {
      let gsl_simpleFull_10 = mat4x4<f32>(clamp(gsl_remap_90(p - vec3<f32>(6.0, 0.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_91(p - vec3<f32>(6.0, 0.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_92(p - vec3<f32>(6.0, 0.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
      gsl_simpleField_30 = mat4x4<f32>(mix(gsl_simpleField_30[0], gsl_simpleFull_10[0], gsl_simpleWeight_30), mix(gsl_simpleField_30[1], gsl_simpleFull_10[1], gsl_simpleWeight_30), mix(gsl_simpleField_30[2], gsl_simpleFull_10[2], gsl_simpleWeight_30), mix(gsl_simpleField_30[3], gsl_simpleFull_10[3], gsl_simpleWeight_30));
    }
    let gsl_m = gsl_simpleField_30; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 15u) {
    let gsl_simpleWeight_31 = gsl_simple(vec3<f32>(1.99, -4.6288023, 3.19), vec3<f32>(10.01, 4.6288023, 7.91));
    var gsl_simpleField_31 = mat4x4<f32>(clamp(gsl_remap_73(p - vec3<f32>(6.0, 0.0, 3.2)) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_74(p - vec3<f32>(6.0, 0.0, 3.2)) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_74(p - vec3<f32>(6.0, 0.0, 3.2)) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    if (gsl_simpleWeight_31 > 0.0) {
      let gsl_simpleFull_11 = mat4x4<f32>(clamp(gsl_remap_93(p - vec3<f32>(6.0, 0.0, 3.2)) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_94(p - vec3<f32>(6.0, 0.0, 3.2)) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_95(p - vec3<f32>(6.0, 0.0, 3.2)) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
      gsl_simpleField_31 = mat4x4<f32>(mix(gsl_simpleField_31[0], gsl_simpleFull_11[0], gsl_simpleWeight_31), mix(gsl_simpleField_31[1], gsl_simpleFull_11[1], gsl_simpleWeight_31), mix(gsl_simpleField_31[2], gsl_simpleFull_11[2], gsl_simpleWeight_31), mix(gsl_simpleField_31[3], gsl_simpleFull_11[3], gsl_simpleWeight_31));
    }
    let gsl_m = gsl_simpleField_31; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 16u) {
    let gsl_simpleWeight_32 = gsl_simple(vec3<f32>(1.99, 11.371198, -0.01), vec3<f32>(10.01, 20.628801, 4.71));
    var gsl_simpleField_32 = mat4x4<f32>(clamp(gsl_remap_73(p - vec3<f32>(6.0, 16.0, 0.0)) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_74(p - vec3<f32>(6.0, 16.0, 0.0)) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_74(p - vec3<f32>(6.0, 16.0, 0.0)) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    if (gsl_simpleWeight_32 > 0.0) {
      let gsl_simpleFull_12 = mat4x4<f32>(clamp(gsl_remap_96(p - vec3<f32>(6.0, 16.0, 0.0)) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_97(p - vec3<f32>(6.0, 16.0, 0.0)) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_98(p - vec3<f32>(6.0, 16.0, 0.0)) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
      gsl_simpleField_32 = mat4x4<f32>(mix(gsl_simpleField_32[0], gsl_simpleFull_12[0], gsl_simpleWeight_32), mix(gsl_simpleField_32[1], gsl_simpleFull_12[1], gsl_simpleWeight_32), mix(gsl_simpleField_32[2], gsl_simpleFull_12[2], gsl_simpleWeight_32), mix(gsl_simpleField_32[3], gsl_simpleFull_12[3], gsl_simpleWeight_32));
    }
    let gsl_m = gsl_simpleField_32; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
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

var<private> gsl_simpleRung: f32;

fn gsl_simple(lo: vec3<f32>, hi: vec3<f32>) -> f32 {
  if (gsl_vanishEye.w <= 0.0 || gsl_simpleRung <= 0.0) { return 1.0; }
  let px = gsl_vanish(lo, hi).z;
  return smoothstep(0.5 * gsl_simpleRung, gsl_simpleRung, px);
}

fn gsl_boxSdf(q: vec3<f32>, lo: vec3<f32>, hi: vec3<f32>) -> f32 {
  let d = abs(q - 0.5 * (lo + hi)) - 0.5 * (hi - lo);
  return length(max(d, vec3<f32>(0.0))) + min(max(d.x, max(d.y, d.z)), 0.0);
}

var<private> gsl_simpleDetail: f32 = 1.0;

var<private> gsl_dispDrawn: f32 = 0.0;

fn gsl_heights(p: vec3<f32>, time: f32) -> array<f32, 24> {
  let gsl_simpleWeight_33 = gsl_simple(vec3<f32>(-22.26, -20.26, -0.01), vec3<f32>(-13.74, -11.74, 21.01));
  var gsl_simpleField_33 = mat4x4<f32>(clamp(gsl_remap_73(p - vec3<f32>(-18.0, -16.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_74(p - vec3<f32>(-18.0, -16.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_74(p - vec3<f32>(-18.0, -16.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
  if (gsl_simpleWeight_33 > 0.0) {
    let gsl_simpleFull_13 = mat4x4<f32>(clamp(gsl_remap_73(p - vec3<f32>(-18.0, -16.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_74(p - vec3<f32>(-18.0, -16.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_74(p - vec3<f32>(-18.0, -16.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    gsl_simpleField_33 = mat4x4<f32>(mix(gsl_simpleField_33[0], gsl_simpleFull_13[0], gsl_simpleWeight_33), mix(gsl_simpleField_33[1], gsl_simpleFull_13[1], gsl_simpleWeight_33), mix(gsl_simpleField_33[2], gsl_simpleFull_13[2], gsl_simpleWeight_33), mix(gsl_simpleField_33[3], gsl_simpleFull_13[3], gsl_simpleWeight_33));
  }
  let gsl_simpleWeight_34 = gsl_simple(vec3<f32>(-22.26, -4.26, -0.01), vec3<f32>(-13.74, 4.26, 21.01));
  var gsl_simpleField_34 = mat4x4<f32>(clamp(gsl_remap_73(p - vec3<f32>(-18.0, 0.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_74(p - vec3<f32>(-18.0, 0.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_74(p - vec3<f32>(-18.0, 0.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
  if (gsl_simpleWeight_34 > 0.0) {
    let gsl_simpleFull_14 = mat4x4<f32>(clamp(gsl_remap_73(p - vec3<f32>(-18.0, 0.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_74(p - vec3<f32>(-18.0, 0.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_74(p - vec3<f32>(-18.0, 0.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    gsl_simpleField_34 = mat4x4<f32>(mix(gsl_simpleField_34[0], gsl_simpleFull_14[0], gsl_simpleWeight_34), mix(gsl_simpleField_34[1], gsl_simpleFull_14[1], gsl_simpleWeight_34), mix(gsl_simpleField_34[2], gsl_simpleFull_14[2], gsl_simpleWeight_34), mix(gsl_simpleField_34[3], gsl_simpleFull_14[3], gsl_simpleWeight_34));
  }
  let gsl_simpleWeight_35 = gsl_simple(vec3<f32>(-22.01, -4.6288023, 3.19), vec3<f32>(-13.99, 4.6288023, 7.91));
  var gsl_simpleField_35 = mat4x4<f32>(clamp(gsl_remap_73(p - vec3<f32>(-18.0, 0.0, 3.2)) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_74(p - vec3<f32>(-18.0, 0.0, 3.2)) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_74(p - vec3<f32>(-18.0, 0.0, 3.2)) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
  if (gsl_simpleWeight_35 > 0.0) {
    let gsl_simpleFull_15 = mat4x4<f32>(clamp(gsl_remap_73(p - vec3<f32>(-18.0, 0.0, 3.2)) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_74(p - vec3<f32>(-18.0, 0.0, 3.2)) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_74(p - vec3<f32>(-18.0, 0.0, 3.2)) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    gsl_simpleField_35 = mat4x4<f32>(mix(gsl_simpleField_35[0], gsl_simpleFull_15[0], gsl_simpleWeight_35), mix(gsl_simpleField_35[1], gsl_simpleFull_15[1], gsl_simpleWeight_35), mix(gsl_simpleField_35[2], gsl_simpleFull_15[2], gsl_simpleWeight_35), mix(gsl_simpleField_35[3], gsl_simpleFull_15[3], gsl_simpleWeight_35));
  }
  let gsl_simpleWeight_36 = gsl_simple(vec3<f32>(-22.01, 11.371198, -0.01), vec3<f32>(-13.99, 20.628801, 4.71));
  var gsl_simpleField_36 = mat4x4<f32>(clamp(gsl_remap_73(p - vec3<f32>(-18.0, 16.0, 0.0)) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_74(p - vec3<f32>(-18.0, 16.0, 0.0)) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_74(p - vec3<f32>(-18.0, 16.0, 0.0)) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
  if (gsl_simpleWeight_36 > 0.0) {
    let gsl_simpleFull_16 = mat4x4<f32>(clamp(gsl_remap_73(p - vec3<f32>(-18.0, 16.0, 0.0)) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_74(p - vec3<f32>(-18.0, 16.0, 0.0)) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_74(p - vec3<f32>(-18.0, 16.0, 0.0)) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    gsl_simpleField_36 = mat4x4<f32>(mix(gsl_simpleField_36[0], gsl_simpleFull_16[0], gsl_simpleWeight_36), mix(gsl_simpleField_36[1], gsl_simpleFull_16[1], gsl_simpleWeight_36), mix(gsl_simpleField_36[2], gsl_simpleFull_16[2], gsl_simpleWeight_36), mix(gsl_simpleField_36[3], gsl_simpleFull_16[3], gsl_simpleWeight_36));
  }
  let gsl_simpleWeight_37 = gsl_simple(vec3<f32>(-10.26, -20.26, -0.01), vec3<f32>(-1.74, -11.74, 21.01));
  var gsl_simpleField_37 = mat4x4<f32>(clamp(gsl_remap_73(p - vec3<f32>(-6.0, -16.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_74(p - vec3<f32>(-6.0, -16.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_74(p - vec3<f32>(-6.0, -16.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
  if (gsl_simpleWeight_37 > 0.0) {
    let gsl_simpleFull_17 = mat4x4<f32>(clamp(gsl_remap_75(p - vec3<f32>(-6.0, -16.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_76(p - vec3<f32>(-6.0, -16.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_77(p - vec3<f32>(-6.0, -16.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    gsl_simpleField_37 = mat4x4<f32>(mix(gsl_simpleField_37[0], gsl_simpleFull_17[0], gsl_simpleWeight_37), mix(gsl_simpleField_37[1], gsl_simpleFull_17[1], gsl_simpleWeight_37), mix(gsl_simpleField_37[2], gsl_simpleFull_17[2], gsl_simpleWeight_37), mix(gsl_simpleField_37[3], gsl_simpleFull_17[3], gsl_simpleWeight_37));
  }
  let gsl_simpleWeight_38 = gsl_simple(vec3<f32>(-10.26, -4.26, -0.01), vec3<f32>(-1.74, 4.26, 21.01));
  var gsl_simpleField_38 = mat4x4<f32>(clamp(gsl_remap_73(p - vec3<f32>(-6.0, 0.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_74(p - vec3<f32>(-6.0, 0.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_74(p - vec3<f32>(-6.0, 0.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
  if (gsl_simpleWeight_38 > 0.0) {
    let gsl_simpleFull_18 = mat4x4<f32>(clamp(gsl_remap_78(p - vec3<f32>(-6.0, 0.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_79(p - vec3<f32>(-6.0, 0.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_80(p - vec3<f32>(-6.0, 0.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    gsl_simpleField_38 = mat4x4<f32>(mix(gsl_simpleField_38[0], gsl_simpleFull_18[0], gsl_simpleWeight_38), mix(gsl_simpleField_38[1], gsl_simpleFull_18[1], gsl_simpleWeight_38), mix(gsl_simpleField_38[2], gsl_simpleFull_18[2], gsl_simpleWeight_38), mix(gsl_simpleField_38[3], gsl_simpleFull_18[3], gsl_simpleWeight_38));
  }
  let gsl_simpleWeight_39 = gsl_simple(vec3<f32>(-10.01, -4.6288023, 3.19), vec3<f32>(-1.99, 4.6288023, 7.91));
  var gsl_simpleField_39 = mat4x4<f32>(clamp(gsl_remap_73(p - vec3<f32>(-6.0, 0.0, 3.2)) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_74(p - vec3<f32>(-6.0, 0.0, 3.2)) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_74(p - vec3<f32>(-6.0, 0.0, 3.2)) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
  if (gsl_simpleWeight_39 > 0.0) {
    let gsl_simpleFull_19 = mat4x4<f32>(clamp(gsl_remap_81(p - vec3<f32>(-6.0, 0.0, 3.2)) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_82(p - vec3<f32>(-6.0, 0.0, 3.2)) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_83(p - vec3<f32>(-6.0, 0.0, 3.2)) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    gsl_simpleField_39 = mat4x4<f32>(mix(gsl_simpleField_39[0], gsl_simpleFull_19[0], gsl_simpleWeight_39), mix(gsl_simpleField_39[1], gsl_simpleFull_19[1], gsl_simpleWeight_39), mix(gsl_simpleField_39[2], gsl_simpleFull_19[2], gsl_simpleWeight_39), mix(gsl_simpleField_39[3], gsl_simpleFull_19[3], gsl_simpleWeight_39));
  }
  let gsl_simpleWeight_40 = gsl_simple(vec3<f32>(-10.01, 11.371198, -0.01), vec3<f32>(-1.99, 20.628801, 4.71));
  var gsl_simpleField_40 = mat4x4<f32>(clamp(gsl_remap_73(p - vec3<f32>(-6.0, 16.0, 0.0)) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_74(p - vec3<f32>(-6.0, 16.0, 0.0)) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_74(p - vec3<f32>(-6.0, 16.0, 0.0)) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
  if (gsl_simpleWeight_40 > 0.0) {
    let gsl_simpleFull_20 = mat4x4<f32>(clamp(gsl_remap_84(p - vec3<f32>(-6.0, 16.0, 0.0)) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_85(p - vec3<f32>(-6.0, 16.0, 0.0)) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_86(p - vec3<f32>(-6.0, 16.0, 0.0)) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    gsl_simpleField_40 = mat4x4<f32>(mix(gsl_simpleField_40[0], gsl_simpleFull_20[0], gsl_simpleWeight_40), mix(gsl_simpleField_40[1], gsl_simpleFull_20[1], gsl_simpleWeight_40), mix(gsl_simpleField_40[2], gsl_simpleFull_20[2], gsl_simpleWeight_40), mix(gsl_simpleField_40[3], gsl_simpleFull_20[3], gsl_simpleWeight_40));
  }
  let gsl_simpleWeight_41 = gsl_simple(vec3<f32>(1.74, -20.26, -0.01), vec3<f32>(10.26, -11.74, 21.01));
  var gsl_simpleField_41 = mat4x4<f32>(clamp(gsl_remap_73(p - vec3<f32>(6.0, -16.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_74(p - vec3<f32>(6.0, -16.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_74(p - vec3<f32>(6.0, -16.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
  if (gsl_simpleWeight_41 > 0.0) {
    let gsl_simpleFull_21 = mat4x4<f32>(clamp(gsl_remap_87(p - vec3<f32>(6.0, -16.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_88(p - vec3<f32>(6.0, -16.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_89(p - vec3<f32>(6.0, -16.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    gsl_simpleField_41 = mat4x4<f32>(mix(gsl_simpleField_41[0], gsl_simpleFull_21[0], gsl_simpleWeight_41), mix(gsl_simpleField_41[1], gsl_simpleFull_21[1], gsl_simpleWeight_41), mix(gsl_simpleField_41[2], gsl_simpleFull_21[2], gsl_simpleWeight_41), mix(gsl_simpleField_41[3], gsl_simpleFull_21[3], gsl_simpleWeight_41));
  }
  let gsl_simpleWeight_42 = gsl_simple(vec3<f32>(1.74, -4.26, -0.01), vec3<f32>(10.26, 4.26, 21.01));
  var gsl_simpleField_42 = mat4x4<f32>(clamp(gsl_remap_73(p - vec3<f32>(6.0, 0.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_74(p - vec3<f32>(6.0, 0.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_74(p - vec3<f32>(6.0, 0.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
  if (gsl_simpleWeight_42 > 0.0) {
    let gsl_simpleFull_22 = mat4x4<f32>(clamp(gsl_remap_90(p - vec3<f32>(6.0, 0.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_91(p - vec3<f32>(6.0, 0.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_92(p - vec3<f32>(6.0, 0.0, 16.0)) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    gsl_simpleField_42 = mat4x4<f32>(mix(gsl_simpleField_42[0], gsl_simpleFull_22[0], gsl_simpleWeight_42), mix(gsl_simpleField_42[1], gsl_simpleFull_22[1], gsl_simpleWeight_42), mix(gsl_simpleField_42[2], gsl_simpleFull_22[2], gsl_simpleWeight_42), mix(gsl_simpleField_42[3], gsl_simpleFull_22[3], gsl_simpleWeight_42));
  }
  let gsl_simpleWeight_43 = gsl_simple(vec3<f32>(1.99, -4.6288023, 3.19), vec3<f32>(10.01, 4.6288023, 7.91));
  var gsl_simpleField_43 = mat4x4<f32>(clamp(gsl_remap_73(p - vec3<f32>(6.0, 0.0, 3.2)) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_74(p - vec3<f32>(6.0, 0.0, 3.2)) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_74(p - vec3<f32>(6.0, 0.0, 3.2)) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
  if (gsl_simpleWeight_43 > 0.0) {
    let gsl_simpleFull_23 = mat4x4<f32>(clamp(gsl_remap_93(p - vec3<f32>(6.0, 0.0, 3.2)) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_94(p - vec3<f32>(6.0, 0.0, 3.2)) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_95(p - vec3<f32>(6.0, 0.0, 3.2)) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    gsl_simpleField_43 = mat4x4<f32>(mix(gsl_simpleField_43[0], gsl_simpleFull_23[0], gsl_simpleWeight_43), mix(gsl_simpleField_43[1], gsl_simpleFull_23[1], gsl_simpleWeight_43), mix(gsl_simpleField_43[2], gsl_simpleFull_23[2], gsl_simpleWeight_43), mix(gsl_simpleField_43[3], gsl_simpleFull_23[3], gsl_simpleWeight_43));
  }
  let gsl_simpleWeight_44 = gsl_simple(vec3<f32>(1.99, 11.371198, -0.01), vec3<f32>(10.01, 20.628801, 4.71));
  var gsl_simpleField_44 = mat4x4<f32>(clamp(gsl_remap_73(p - vec3<f32>(6.0, 16.0, 0.0)) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_74(p - vec3<f32>(6.0, 16.0, 0.0)) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_74(p - vec3<f32>(6.0, 16.0, 0.0)) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
  if (gsl_simpleWeight_44 > 0.0) {
    let gsl_simpleFull_24 = mat4x4<f32>(clamp(gsl_remap_96(p - vec3<f32>(6.0, 16.0, 0.0)) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_97(p - vec3<f32>(6.0, 16.0, 0.0)) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_98(p - vec3<f32>(6.0, 16.0, 0.0)) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    gsl_simpleField_44 = mat4x4<f32>(mix(gsl_simpleField_44[0], gsl_simpleFull_24[0], gsl_simpleWeight_44), mix(gsl_simpleField_44[1], gsl_simpleFull_24[1], gsl_simpleWeight_44), mix(gsl_simpleField_44[2], gsl_simpleFull_24[2], gsl_simpleWeight_44), mix(gsl_simpleField_44[3], gsl_simpleFull_24[3], gsl_simpleWeight_44));
  }
  return array<f32, 24>(0.0, 0.0, 0.0, 0.0, gsl_simpleField_33[2].z, gsl_simpleField_34[2].z, gsl_simpleField_35[2].z, gsl_simpleField_36[2].z, gsl_simpleField_37[2].z, gsl_simpleField_38[2].z, gsl_simpleField_39[2].z, gsl_simpleField_40[2].z, gsl_simpleField_41[2].z, gsl_simpleField_42[2].z, gsl_simpleField_43[2].z, gsl_simpleField_44[2].z, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0);
}

fn gsl_sdShapeDisplaced(p: vec3<f32>, time: f32, gsl_dispK: f32) -> gsl_shape {
  let gsl_d = gsl_shapeParts(p, time);
  let gsl_h = gsl_heights(p, time);
  let gsl_s1 = gsl_shape(gsl_d[0], 1u);
  let gsl_s2 = gsl_shape(gsl_d[1], 2u);
  let gsl_s3 = gsl_shape(gsl_d[2], 3u);
  let gsl_s4 = gsl_shape(gsl_d[3], 4u);
  let gsl_s5 = gsl_shape(gsl_d[4] - gsl_h[4] * gsl_dispK, 5u);
  let gsl_s6 = gsl_shape(gsl_d[5] - gsl_h[5] * gsl_dispK, 6u);
  let gsl_s7 = gsl_shape(gsl_d[6] - gsl_h[6] * gsl_dispK, 7u);
  let gsl_s8 = gsl_shape(gsl_d[7] - gsl_h[7] * gsl_dispK, 8u);
  let gsl_s9 = gsl_shape(gsl_d[8] - gsl_h[8] * gsl_dispK, 9u);
  let gsl_s10 = gsl_shape(gsl_d[9] - gsl_h[9] * gsl_dispK, 10u);
  let gsl_s11 = gsl_shape(gsl_d[10] - gsl_h[10] * gsl_dispK, 11u);
  let gsl_s12 = gsl_shape(gsl_d[11] - gsl_h[11] * gsl_dispK, 12u);
  let gsl_s13 = gsl_shape(gsl_d[12] - gsl_h[12] * gsl_dispK, 13u);
  let gsl_s14 = gsl_shape(gsl_d[13] - gsl_h[13] * gsl_dispK, 14u);
  let gsl_s15 = gsl_shape(gsl_d[14] - gsl_h[14] * gsl_dispK, 15u);
  let gsl_s16 = gsl_shape(gsl_d[15] - gsl_h[15] * gsl_dispK, 16u);
  let gsl_s17 = gsl_shape(gsl_d[16], 17u);
  let gsl_s18 = gsl_shape(gsl_d[17], 18u);
  let gsl_s19 = gsl_shape(gsl_d[18], 19u);
  let gsl_s20 = gsl_shape(gsl_d[19], 20u);
  let gsl_s21 = gsl_shape(gsl_d[20], 21u);
  let gsl_s22 = gsl_shape(gsl_d[21], 22u);
  let gsl_s23 = gsl_shape(gsl_d[22], 23u);
  let gsl_s24 = gsl_shape(gsl_d[23], 24u);
  return gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_s1, gsl_s2), gsl_s3), gsl_s4), gsl_s5), gsl_s6), gsl_s7), gsl_s8), gsl_s9), gsl_s10), gsl_s11), gsl_s12), gsl_s13), gsl_s14), gsl_s15), gsl_s16), gsl_s17), gsl_s18), gsl_s19), gsl_s20), gsl_s21), gsl_s22), gsl_s23), gsl_s24);
}
