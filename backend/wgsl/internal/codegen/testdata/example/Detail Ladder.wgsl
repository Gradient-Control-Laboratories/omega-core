fn gsl_floored_vec2(a: vec2<f32>, n: vec2<f32>) -> vec2<f32> {
  return a - n * floor(a / n);
}

fn thread_formDepth(pitchVal: f32, noiseTex: texture_3d<f32>) -> f32 {
  return 0.5412659 * pitchVal;
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

fn thread_paint(base: vec3<f32>, height: f32, pitchVal: f32, noiseTex: texture_3d<f32>) -> vec3<f32> {
  var gsl_simpleValue = base;
  if (gsl_simpleDetail > 0.0) {
    let depth = clamp(-height / thread_formDepth(pitchVal, noiseTex), 0.0, 1.0);
    gsl_simpleValue = mix(gsl_simpleValue, color_mixOklab(base, base * 0.35, depth, noiseTex), gsl_simpleDetail);
  }
  let painted = gsl_simpleValue;
  return painted;
}

fn thread_reliefColor(base: vec3<f32>, height: f32, pitchVal: f32, noiseTex: texture_3d<f32>) -> vec3<f32> {
  var gsl_simpleValue = base;
  if (gsl_simpleDetail > 0.0) {
    let depth = clamp(-height / thread_formDepth(pitchVal, noiseTex), 0.0, 1.0);
    gsl_simpleValue = mix(gsl_simpleValue, color_mixOklab(base, base * 0.35, depth * (1.0 - gsl_dispDrawn), noiseTex), gsl_simpleDetail);
  }
  let painted = gsl_simpleValue;
  return painted;
}

fn drill_prismVoid(p: vec3<f32>, zTop: f32, section: f32, depth: f32, noiseTex: texture_3d<f32>) -> f32 {
  return max(section, zTop - depth - p.z);
}

fn gsl_remap(p: vec3<f32>, material_brushedMetal: mat4x4<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let gslCarry4 = fastener_screw(max(max(max(length(p.xy) - 1.9, -p.z), p.z - 2.0), (length(p.xy) - 1.9 + p.z - 2.0 + 0.2) * 0.7071), max(max(max(length(p.xy) - 1.0, -4.0 - p.z), p.z - 1.0), (length(p.xy) - 1.0 + -p.z - 4.0 + 0.3) * 0.7071), drill_prismVoid(p, 2.0, max(max(abs(p.x) - 0.75, abs(0.5 * p.x + 0.8660254 * p.y) - 0.75), abs(0.5 * p.x - 0.8660254 * p.y) - 0.75), 1.0, noiseTex), noiseTex);
  return gslCarry4;
}

fn gsl_remap_2(p: vec3<f32>, material_brushedMetal: mat4x4<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let gslCarry1 = min(max(max(length(p.xy) - 1.9, -p.z), p.z - 2.0), max(max(length(p.xy) - 1.0, -4.0 - p.z), p.z - 1.0));
  return gslCarry1;
}

fn builtin_torus(p: vec3<f32>, majorRadius: f32, minorRadius: f32, noiseTex: texture_3d<f32>) -> f32 {
  let radial = length(p.xy) - majorRadius;
  return length(vec2<f32>(radial, p.z)) - minorRadius;
}

fn builtin_sphere(p: vec3<f32>, center: vec3<f32>, radius: f32, noiseTex: texture_3d<f32>) -> f32 {
  return length(p - center) - radius;
}

fn builtin_sphere_2(p: vec3<f32>, radius: f32, noiseTex: texture_3d<f32>) -> f32 {
  return length(p) - radius;
}

fn map_repeatCentered(x: vec2<f32>, period: vec2<f32>, noiseTex: texture_3d<f32>) -> vec2<f32> {
  let half = period * 0.5;
  return gsl_floored_vec2((gsl_floored_vec2((x + half), period) + period), period) - half;
}

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

fn gsl_remap_3(p: vec3<f32>, material_brushedMetal: mat4x4<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let gsl_tmp_24 = map_repeatCentered(p.xy - vec2<f32>(0.8660254, 2.5), vec2<f32>(1.7320508, 3.0), noiseTex);
  let gsl_tmp_23 = map_repeatCentered(p.xy - vec2<f32>(0.8660254, 1.5), vec2<f32>(1.7320508, 3.0), noiseTex);
  let gsl_tmp_26 = select(gsl_tmp_24, gsl_tmp_23, dot(gsl_tmp_23, gsl_tmp_23) < dot(gsl_tmp_24, gsl_tmp_24));
  let gsl_tmp_22 = map_repeatCentered(p.xy - vec2<f32>(0.0, 1.0), vec2<f32>(1.7320508, 3.0), noiseTex);
  let gsl_tmp_21 = map_repeatCentered(p.xy, vec2<f32>(1.7320508, 3.0), noiseTex);
  let gsl_tmp_25 = select(gsl_tmp_22, gsl_tmp_21, dot(gsl_tmp_21, gsl_tmp_21) < dot(gsl_tmp_22, gsl_tmp_22));
  let gsl_tmp_27 = select(gsl_tmp_26, gsl_tmp_25, dot(gsl_tmp_25, gsl_tmp_25) < dot(gsl_tmp_26, gsl_tmp_26));
  let gsl_tmp_28 = vec3<f32>(0.6151075, -0.05253712, -0.094294235) + vec3<f32>(0.38489252, 0.05253712, 0.09429427) * texture_fadeTo(texture_coverageDisc(length(vec3<f32>(gsl_tmp_27.x, gsl_tmp_27.y, p.z).xy), 0.25, noiseTex), 0.15114996, 1.0, noiseTex);
  let gsl_tmp_29 = gsl_tmp_28.x + 0.39633778 * gsl_tmp_28.y + 0.21580376 * gsl_tmp_28.z;
  let gsl_tmp_30 = gsl_tmp_28.x - 0.105561346 * gsl_tmp_28.y - 0.06385417 * gsl_tmp_28.z;
  let gsl_tmp_31 = gsl_tmp_28.x - 0.08948418 * gsl_tmp_28.y - 1.2914855 * gsl_tmp_28.z;
  let gsl_tmp_32 = vec3<f32>(gsl_tmp_29 * gsl_tmp_29 * gsl_tmp_29, gsl_tmp_30 * gsl_tmp_30 * gsl_tmp_30, gsl_tmp_31 * gsl_tmp_31 * gsl_tmp_31);
  let gsl_tmp_166 = vec3<f32>(dot(vec3<f32>(4.0767417, -3.3077116, 0.23096994), gsl_tmp_32), dot(vec3<f32>(-1.268438, 2.6097574, -0.34131938), gsl_tmp_32), dot(vec3<f32>(-0.0041960864, -0.7034186, 1.7076147), gsl_tmp_32));
  return vec3<f32>(color_linearToSrgbChannel(gsl_tmp_166.x, noiseTex), color_linearToSrgbChannel(gsl_tmp_166.y, noiseTex), color_linearToSrgbChannel(gsl_tmp_166.z, noiseTex));
}

fn ease_cubic(t: f32, noiseTex: texture_3d<f32>) -> f32 {
  return t * t * t * (t * (t * 6.0 - 15.0) + 10.0);
}

fn gsl_remap_4(p: vec3<f32>, material_brushedMetal: mat4x4<f32>, color_gray: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let gslCarry10 = fastener_screw(max(max(max(length(p.xy) - 3.5, -p.z), p.z - 4.0), (length(p.xy) - 3.5 + p.z - 4.0 + 0.2) * 0.7071), max(max(max(length(p.xy) - 2.0, -8.0 - p.z), p.z - 2.0), (length(p.xy) - 2.0 + -p.z - 8.0 + 0.6) * 0.7071), drill_prismVoid(p, 4.0, max(max(abs(p.x) - 1.5, abs(0.5 * p.x + 0.8660254 * p.y) - 1.5), abs(0.5 * p.x - 0.8660254 * p.y) - 1.5), 2.0, noiseTex), noiseTex);
  return gslCarry10;
}

fn gsl_remap_5(p: vec3<f32>, material_brushedMetal: mat4x4<f32>, color_gray: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let gslCarry7 = min(max(max(length(p.xy) - 3.5, -p.z), p.z - 4.0), max(max(length(p.xy) - 2.0, -8.0 - p.z), p.z - 2.0));
  return gslCarry7;
}

fn gsl_remap_6(p: vec3<f32>, material_brushedMetal: mat4x4<f32>, color_gray: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let gsl_tmp_57 = map_repeatCentered(p.xy - vec2<f32>(1.7320508, 5.0), vec2<f32>(3.4641016, 6.0), noiseTex);
  let gsl_tmp_56 = map_repeatCentered(p.xy - vec2<f32>(1.7320508, 3.0), vec2<f32>(3.4641016, 6.0), noiseTex);
  let gsl_tmp_59 = select(gsl_tmp_57, gsl_tmp_56, dot(gsl_tmp_56, gsl_tmp_56) < dot(gsl_tmp_57, gsl_tmp_57));
  let gsl_tmp_55 = map_repeatCentered(p.xy - vec2<f32>(0.0, 2.0), vec2<f32>(3.4641016, 6.0), noiseTex);
  let gsl_tmp_54 = map_repeatCentered(p.xy, vec2<f32>(3.4641016, 6.0), noiseTex);
  let gsl_tmp_58 = select(gsl_tmp_55, gsl_tmp_54, dot(gsl_tmp_54, gsl_tmp_54) < dot(gsl_tmp_55, gsl_tmp_55));
  let gsl_tmp_60 = select(gsl_tmp_59, gsl_tmp_58, dot(gsl_tmp_58, gsl_tmp_58) < dot(gsl_tmp_59, gsl_tmp_59));
  let gsl_tmp_61 = vec3<f32>(0.6151075, -0.05253712, -0.094294235) + vec3<f32>(0.38489252, 0.05253712, 0.09429427) * texture_fadeTo(texture_coverageDisc(length(vec3<f32>(gsl_tmp_60.x, gsl_tmp_60.y, p.z).xy), 0.5, noiseTex), 0.15114996, 2.0, noiseTex);
  let gsl_tmp_62 = gsl_tmp_61.x + 0.39633778 * gsl_tmp_61.y + 0.21580376 * gsl_tmp_61.z;
  let gsl_tmp_63 = gsl_tmp_61.x - 0.105561346 * gsl_tmp_61.y - 0.06385417 * gsl_tmp_61.z;
  let gsl_tmp_64 = gsl_tmp_61.x - 0.08948418 * gsl_tmp_61.y - 1.2914855 * gsl_tmp_61.z;
  let gsl_tmp_65 = vec3<f32>(gsl_tmp_62 * gsl_tmp_62 * gsl_tmp_62, gsl_tmp_63 * gsl_tmp_63 * gsl_tmp_63, gsl_tmp_64 * gsl_tmp_64 * gsl_tmp_64);
  let gsl_tmp_171 = vec3<f32>(dot(vec3<f32>(4.0767417, -3.3077116, 0.23096994), gsl_tmp_65), dot(vec3<f32>(-1.268438, 2.6097574, -0.34131938), gsl_tmp_65), dot(vec3<f32>(-0.0041960864, -0.7034186, 1.7076147), gsl_tmp_65));
  return vec3<f32>(color_linearToSrgbChannel(gsl_tmp_171.x, noiseTex), color_linearToSrgbChannel(gsl_tmp_171.y, noiseTex), color_linearToSrgbChannel(gsl_tmp_171.z, noiseTex));
}

fn gsl_remap_7(p: vec3<f32>, material_brushedMetal: mat4x4<f32>, color_gray: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let gslCarry16 = fastener_screw(max(max(max(length(p.xy) - 6.5, -p.z), p.z - 8.0), (length(p.xy) - 6.5 + p.z - 8.0 + 0.2) * 0.7071), max(max(max(length(p.xy) - 4.0, -16.0 - p.z), p.z - 4.0), (length(p.xy) - 4.0 + -p.z - 16.0 + 1.2) * 0.7071), drill_prismVoid(p, 8.0, max(max(abs(p.x) - 3.0, abs(0.5 * p.x + 0.8660254 * p.y) - 3.0), abs(0.5 * p.x - 0.8660254 * p.y) - 3.0), 4.0, noiseTex), noiseTex);
  return gslCarry16;
}

fn gsl_remap_8(p: vec3<f32>, material_brushedMetal: mat4x4<f32>, color_gray: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let gslCarry13 = min(max(max(length(p.xy) - 6.5, -p.z), p.z - 8.0), max(max(length(p.xy) - 4.0, -16.0 - p.z), p.z - 4.0));
  return gslCarry13;
}

fn gsl_remap_9(p: vec3<f32>, material_brushedMetal: mat4x4<f32>, color_gray: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let gsl_tmp_90 = map_repeatCentered(p.xy - vec2<f32>(3.4641016, 10.0), vec2<f32>(6.928203, 12.0), noiseTex);
  let gsl_tmp_89 = map_repeatCentered(p.xy - vec2<f32>(3.4641016, 6.0), vec2<f32>(6.928203, 12.0), noiseTex);
  let gsl_tmp_92 = select(gsl_tmp_90, gsl_tmp_89, dot(gsl_tmp_89, gsl_tmp_89) < dot(gsl_tmp_90, gsl_tmp_90));
  let gsl_tmp_88 = map_repeatCentered(p.xy - vec2<f32>(0.0, 4.0), vec2<f32>(6.928203, 12.0), noiseTex);
  let gsl_tmp_87 = map_repeatCentered(p.xy, vec2<f32>(6.928203, 12.0), noiseTex);
  let gsl_tmp_91 = select(gsl_tmp_88, gsl_tmp_87, dot(gsl_tmp_87, gsl_tmp_87) < dot(gsl_tmp_88, gsl_tmp_88));
  let gsl_tmp_93 = select(gsl_tmp_92, gsl_tmp_91, dot(gsl_tmp_91, gsl_tmp_91) < dot(gsl_tmp_92, gsl_tmp_92));
  let gsl_tmp_94 = vec3<f32>(0.6151075, -0.05253712, -0.094294235) + vec3<f32>(0.38489252, 0.05253712, 0.09429427) * texture_fadeTo(texture_coverageDisc(length(vec3<f32>(gsl_tmp_93.x, gsl_tmp_93.y, p.z).xy), 1.0, noiseTex), 0.15114996, 4.0, noiseTex);
  let gsl_tmp_95 = gsl_tmp_94.x + 0.39633778 * gsl_tmp_94.y + 0.21580376 * gsl_tmp_94.z;
  let gsl_tmp_96 = gsl_tmp_94.x - 0.105561346 * gsl_tmp_94.y - 0.06385417 * gsl_tmp_94.z;
  let gsl_tmp_97 = gsl_tmp_94.x - 0.08948418 * gsl_tmp_94.y - 1.2914855 * gsl_tmp_94.z;
  let gsl_tmp_98 = vec3<f32>(gsl_tmp_95 * gsl_tmp_95 * gsl_tmp_95, gsl_tmp_96 * gsl_tmp_96 * gsl_tmp_96, gsl_tmp_97 * gsl_tmp_97 * gsl_tmp_97);
  let gsl_tmp_176 = vec3<f32>(dot(vec3<f32>(4.0767417, -3.3077116, 0.23096994), gsl_tmp_98), dot(vec3<f32>(-1.268438, 2.6097574, -0.34131938), gsl_tmp_98), dot(vec3<f32>(-0.0041960864, -0.7034186, 1.7076147), gsl_tmp_98));
  return vec3<f32>(color_linearToSrgbChannel(gsl_tmp_176.x, noiseTex), color_linearToSrgbChannel(gsl_tmp_176.y, noiseTex), color_linearToSrgbChannel(gsl_tmp_176.z, noiseTex));
}

fn gsl_remap_10(p: vec3<f32>, material_brushedMetal: mat4x4<f32>, color_gray: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let gslCarry22 = fastener_screw(max(max(max(length(p.xy) - 12.0, -p.z), p.z - 16.0), (length(p.xy) - 12.0 + p.z - 16.0 + 0.2) * 0.7071), max(max(max(length(p.xy) - 8.0, -32.0 - p.z), p.z - 8.0), (length(p.xy) - 8.0 + -p.z - 32.0 + 2.4) * 0.7071), drill_prismVoid(p, 16.0, max(max(abs(p.x) - 7.0, abs(0.5 * p.x + 0.8660254 * p.y) - 7.0), abs(0.5 * p.x - 0.8660254 * p.y) - 7.0), 8.0, noiseTex), noiseTex);
  return gslCarry22;
}

fn gsl_remap_11(p: vec3<f32>, material_brushedMetal: mat4x4<f32>, color_gray: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let gslCarry19 = min(max(max(length(p.xy) - 12.0, -p.z), p.z - 16.0), max(max(length(p.xy) - 8.0, -32.0 - p.z), p.z - 8.0));
  return gslCarry19;
}

fn gsl_remap_12(p: vec3<f32>, material_brushedMetal: mat4x4<f32>, color_gray: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let gsl_tmp_123 = map_repeatCentered(p.xy - vec2<f32>(6.928203, 20.0), vec2<f32>(13.856406, 24.0), noiseTex);
  let gsl_tmp_122 = map_repeatCentered(p.xy - vec2<f32>(6.928203, 12.0), vec2<f32>(13.856406, 24.0), noiseTex);
  let gsl_tmp_125 = select(gsl_tmp_123, gsl_tmp_122, dot(gsl_tmp_122, gsl_tmp_122) < dot(gsl_tmp_123, gsl_tmp_123));
  let gsl_tmp_121 = map_repeatCentered(p.xy - vec2<f32>(0.0, 8.0), vec2<f32>(13.856406, 24.0), noiseTex);
  let gsl_tmp_120 = map_repeatCentered(p.xy, vec2<f32>(13.856406, 24.0), noiseTex);
  let gsl_tmp_124 = select(gsl_tmp_121, gsl_tmp_120, dot(gsl_tmp_120, gsl_tmp_120) < dot(gsl_tmp_121, gsl_tmp_121));
  let gsl_tmp_126 = select(gsl_tmp_125, gsl_tmp_124, dot(gsl_tmp_124, gsl_tmp_124) < dot(gsl_tmp_125, gsl_tmp_125));
  let gsl_tmp_127 = vec3<f32>(0.6151075, -0.05253712, -0.094294235) + vec3<f32>(0.38489252, 0.05253712, 0.09429427) * texture_fadeTo(texture_coverageDisc(length(vec3<f32>(gsl_tmp_126.x, gsl_tmp_126.y, p.z).xy), 2.0, noiseTex), 0.15114996, 8.0, noiseTex);
  let gsl_tmp_128 = gsl_tmp_127.x + 0.39633778 * gsl_tmp_127.y + 0.21580376 * gsl_tmp_127.z;
  let gsl_tmp_129 = gsl_tmp_127.x - 0.105561346 * gsl_tmp_127.y - 0.06385417 * gsl_tmp_127.z;
  let gsl_tmp_130 = gsl_tmp_127.x - 0.08948418 * gsl_tmp_127.y - 1.2914855 * gsl_tmp_127.z;
  let gsl_tmp_131 = vec3<f32>(gsl_tmp_128 * gsl_tmp_128 * gsl_tmp_128, gsl_tmp_129 * gsl_tmp_129 * gsl_tmp_129, gsl_tmp_130 * gsl_tmp_130 * gsl_tmp_130);
  let gsl_tmp_181 = vec3<f32>(dot(vec3<f32>(4.0767417, -3.3077116, 0.23096994), gsl_tmp_131), dot(vec3<f32>(-1.268438, 2.6097574, -0.34131938), gsl_tmp_131), dot(vec3<f32>(-0.0041960864, -0.7034186, 1.7076147), gsl_tmp_131));
  return vec3<f32>(color_linearToSrgbChannel(gsl_tmp_181.x, noiseTex), color_linearToSrgbChannel(gsl_tmp_181.y, noiseTex), color_linearToSrgbChannel(gsl_tmp_181.z, noiseTex));
}

fn gsl_remap_13(p: vec3<f32>, material_brushedMetal: mat4x4<f32>, color_gray: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let gslCarry28 = fastener_screw(max(max(max(length(p.xy) - 22.5, -p.z), p.z - 30.0), (length(p.xy) - 22.5 + p.z - 30.0 + 0.2) * 0.7071), max(max(max(length(p.xy) - 15.0, -60.0 - p.z), p.z - 15.0), (length(p.xy) - 15.0 + -p.z - 60.0 + 4.5) * 0.7071), drill_prismVoid(p, 30.0, max(max(abs(p.x) - 11.0, abs(0.5 * p.x + 0.8660254 * p.y) - 11.0), abs(0.5 * p.x - 0.8660254 * p.y) - 11.0), 15.5, noiseTex), noiseTex);
  return gslCarry28;
}

fn gsl_remap_14(p: vec3<f32>, material_brushedMetal: mat4x4<f32>, color_gray: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let gslCarry25 = min(max(max(length(p.xy) - 22.5, -p.z), p.z - 30.0), max(max(length(p.xy) - 15.0, -60.0 - p.z), p.z - 15.0));
  return gslCarry25;
}

fn gsl_remap_15(p: vec3<f32>, material_brushedMetal: mat4x4<f32>, color_gray: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let gsl_tmp_156 = map_repeatCentered(p.xy - vec2<f32>(13.856406, 40.0), vec2<f32>(27.712812, 48.0), noiseTex);
  let gsl_tmp_155 = map_repeatCentered(p.xy - vec2<f32>(13.856406, 24.0), vec2<f32>(27.712812, 48.0), noiseTex);
  let gsl_tmp_158 = select(gsl_tmp_156, gsl_tmp_155, dot(gsl_tmp_155, gsl_tmp_155) < dot(gsl_tmp_156, gsl_tmp_156));
  let gsl_tmp_154 = map_repeatCentered(p.xy - vec2<f32>(0.0, 16.0), vec2<f32>(27.712812, 48.0), noiseTex);
  let gsl_tmp_153 = map_repeatCentered(p.xy, vec2<f32>(27.712812, 48.0), noiseTex);
  let gsl_tmp_157 = select(gsl_tmp_154, gsl_tmp_153, dot(gsl_tmp_153, gsl_tmp_153) < dot(gsl_tmp_154, gsl_tmp_154));
  let gsl_tmp_159 = select(gsl_tmp_158, gsl_tmp_157, dot(gsl_tmp_157, gsl_tmp_157) < dot(gsl_tmp_158, gsl_tmp_158));
  let gsl_tmp_160 = vec3<f32>(0.6151075, -0.05253712, -0.094294235) + vec3<f32>(0.38489252, 0.05253712, 0.09429427) * texture_fadeTo(texture_coverageDisc(length(vec3<f32>(gsl_tmp_159.x, gsl_tmp_159.y, p.z).xy), 4.0, noiseTex), 0.15114996, 16.0, noiseTex);
  let gsl_tmp_161 = gsl_tmp_160.x + 0.39633778 * gsl_tmp_160.y + 0.21580376 * gsl_tmp_160.z;
  let gsl_tmp_162 = gsl_tmp_160.x - 0.105561346 * gsl_tmp_160.y - 0.06385417 * gsl_tmp_160.z;
  let gsl_tmp_163 = gsl_tmp_160.x - 0.08948418 * gsl_tmp_160.y - 1.2914855 * gsl_tmp_160.z;
  let gsl_tmp_164 = vec3<f32>(gsl_tmp_161 * gsl_tmp_161 * gsl_tmp_161, gsl_tmp_162 * gsl_tmp_162 * gsl_tmp_162, gsl_tmp_163 * gsl_tmp_163 * gsl_tmp_163);
  let gsl_tmp_186 = vec3<f32>(dot(vec3<f32>(4.0767417, -3.3077116, 0.23096994), gsl_tmp_164), dot(vec3<f32>(-1.268438, 2.6097574, -0.34131938), gsl_tmp_164), dot(vec3<f32>(-0.0041960864, -0.7034186, 1.7076147), gsl_tmp_164));
  return vec3<f32>(color_linearToSrgbChannel(gsl_tmp_186.x, noiseTex), color_linearToSrgbChannel(gsl_tmp_186.y, noiseTex), color_linearToSrgbChannel(gsl_tmp_186.z, noiseTex));
}

fn gsl_remap_16(p: vec3<f32>, material_brushedMetal: mat4x4<f32>, color_gray: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let gsl_tmp_10 = 6.2831855 * p.z / 0.4;
  let gsl_tmp_9 = max(length(p.xy), 0.2);
  let gslCarry5 = thread_reliefColor(vec3<f32>(0.77, 0.78, 0.78), -min(min(clamp((acos(clamp(cos(gsl_tmp_10) * p.x / gsl_tmp_9 + sin(gsl_tmp_10) * p.y / gsl_tmp_9, -1.0, 1.0)) / 3.1415927 * 0.2 - 0.025) * 1.7320508, 0.0, 0.21650636), max(0.0, 0.5 * min(min(0.0 - p.z, p.z - -4.0), 1.2165064 - length(p.xy)))), max(0.0, length(p.xy) - 0.5669873)), 0.4, noiseTex);
  return gslCarry5;
}

fn gsl_remap_17(p: vec3<f32>, material_brushedMetal: mat4x4<f32>, color_gray: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let gsl_tmp_4 = 6.2831855 * p.z / 0.4;
  let gsl_tmp_3 = max(length(p.xy), 0.2);
  let gslCarry2 = thread_paint(vec3<f32>(0.77, 0.78, 0.78), -min(min(clamp((acos(clamp(cos(gsl_tmp_4) * p.x / gsl_tmp_3 + sin(gsl_tmp_4) * p.y / gsl_tmp_3, -1.0, 1.0)) / 3.1415927 * 0.2 - 0.025) * 1.7320508, 0.0, 0.21650636), max(0.0, 0.5 * min(min(0.0 - p.z, p.z - -4.0), 1.2165064 - length(p.xy)))), max(0.0, length(p.xy) - 0.5669873)), 0.4, noiseTex);
  return gslCarry2;
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

fn gsl_remap_18(p: vec3<f32>, material_brushedMetal: mat4x4<f32>, color_gray: vec3<f32>, color_white: vec3<f32>, color_black: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let gsl_tmp_43 = 6.2831855 * p.z / 0.7;
  let gsl_tmp_42 = max(length(p.xy), 0.2);
  let gslCarry11 = thread_reliefColor(vec3<f32>(0.77, 0.78, 0.78), -min(min(clamp((acos(clamp(cos(gsl_tmp_43) * p.x / gsl_tmp_42 + sin(gsl_tmp_43) * p.y / gsl_tmp_42, -1.0, 1.0)) / 3.1415927 * 0.35 - 0.04375) * 1.7320508, 0.0, 0.37888613), max(0.0, 0.5 * min(min(0.0 - p.z, p.z - -8.0), 2.3788862 - length(p.xy)))), max(0.0, length(p.xy) - 1.2422278)), 0.7, noiseTex);
  return gslCarry11;
}

fn gsl_remap_19(p: vec3<f32>, material_brushedMetal: mat4x4<f32>, color_gray: vec3<f32>, color_white: vec3<f32>, color_black: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let gsl_tmp_37 = 6.2831855 * p.z / 0.7;
  let gsl_tmp_36 = max(length(p.xy), 0.2);
  let gslCarry8 = thread_paint(vec3<f32>(0.77, 0.78, 0.78), -min(min(clamp((acos(clamp(cos(gsl_tmp_37) * p.x / gsl_tmp_36 + sin(gsl_tmp_37) * p.y / gsl_tmp_36, -1.0, 1.0)) / 3.1415927 * 0.35 - 0.04375) * 1.7320508, 0.0, 0.37888613), max(0.0, 0.5 * min(min(0.0 - p.z, p.z - -8.0), 2.3788862 - length(p.xy)))), max(0.0, length(p.xy) - 1.2422278)), 0.7, noiseTex);
  return gslCarry8;
}

fn gsl_remap_20(p: vec3<f32>, material_brushedMetal: mat4x4<f32>, color_gray: vec3<f32>, color_white: vec3<f32>, color_black: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let gsl_tmp_76 = 6.2831855 * p.z / 1.25;
  let gsl_tmp_75 = max(length(p.xy), 0.2);
  let gslCarry17 = thread_reliefColor(vec3<f32>(0.77, 0.78, 0.78), -min(min(clamp((acos(clamp(cos(gsl_tmp_76) * p.x / gsl_tmp_75 + sin(gsl_tmp_76) * p.y / gsl_tmp_75, -1.0, 1.0)) / 3.1415927 * 0.625 - 0.078125) * 1.7320508, 0.0, 0.6765824), max(0.0, 0.5 * min(min(0.0 - p.z, p.z - -16.0), 4.6765823 - length(p.xy)))), max(0.0, length(p.xy) - 2.6468353)), 1.25, noiseTex);
  return gslCarry17;
}

fn gsl_remap_21(p: vec3<f32>, material_brushedMetal: mat4x4<f32>, color_gray: vec3<f32>, color_white: vec3<f32>, color_black: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let gsl_tmp_70 = 6.2831855 * p.z / 1.25;
  let gsl_tmp_69 = max(length(p.xy), 0.2);
  let gslCarry14 = thread_paint(vec3<f32>(0.77, 0.78, 0.78), -min(min(clamp((acos(clamp(cos(gsl_tmp_70) * p.x / gsl_tmp_69 + sin(gsl_tmp_70) * p.y / gsl_tmp_69, -1.0, 1.0)) / 3.1415927 * 0.625 - 0.078125) * 1.7320508, 0.0, 0.6765824), max(0.0, 0.5 * min(min(0.0 - p.z, p.z - -16.0), 4.6765823 - length(p.xy)))), max(0.0, length(p.xy) - 2.6468353)), 1.25, noiseTex);
  return gslCarry14;
}

fn gsl_remap_22(p: vec3<f32>, material_brushedMetal: mat4x4<f32>, color_gray: vec3<f32>, color_white: vec3<f32>, color_black: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let gsl_tmp_109 = 6.2831855 * p.z / 2.0;
  let gsl_tmp_108 = max(length(p.xy), 0.2);
  let gslCarry23 = thread_reliefColor(vec3<f32>(0.77, 0.78, 0.78), -min(min(clamp((acos(clamp(cos(gsl_tmp_109) * p.x / gsl_tmp_108 + sin(gsl_tmp_109) * p.y / gsl_tmp_108, -1.0, 1.0)) / 3.1415927 * 1.0 - 0.125) * 1.7320508, 0.0, 1.0825318), max(0.0, 0.5 * min(min(0.0 - p.z, p.z - -32.0), 9.082532 - length(p.xy)))), max(0.0, length(p.xy) - 5.8349366)), 2.0, noiseTex);
  return gslCarry23;
}

fn gsl_remap_23(p: vec3<f32>, material_brushedMetal: mat4x4<f32>, color_gray: vec3<f32>, color_white: vec3<f32>, color_black: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let gsl_tmp_103 = 6.2831855 * p.z / 2.0;
  let gsl_tmp_102 = max(length(p.xy), 0.2);
  let gslCarry20 = thread_paint(vec3<f32>(0.77, 0.78, 0.78), -min(min(clamp((acos(clamp(cos(gsl_tmp_103) * p.x / gsl_tmp_102 + sin(gsl_tmp_103) * p.y / gsl_tmp_102, -1.0, 1.0)) / 3.1415927 * 1.0 - 0.125) * 1.7320508, 0.0, 1.0825318), max(0.0, 0.5 * min(min(0.0 - p.z, p.z - -32.0), 9.082532 - length(p.xy)))), max(0.0, length(p.xy) - 5.8349366)), 2.0, noiseTex);
  return gslCarry20;
}

fn gsl_remap_24(p: vec3<f32>, material_brushedMetal: mat4x4<f32>, color_gray: vec3<f32>, color_white: vec3<f32>, color_black: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let gsl_tmp_142 = 6.2831855 * p.z / 3.5;
  let gsl_tmp_141 = max(length(p.xy), 0.2);
  let gslCarry29 = thread_reliefColor(vec3<f32>(0.77, 0.78, 0.78), -min(min(clamp((acos(clamp(cos(gsl_tmp_142) * p.x / gsl_tmp_141 + sin(gsl_tmp_142) * p.y / gsl_tmp_141, -1.0, 1.0)) / 3.1415927 * 1.75 - 0.21875) * 1.7320508, 0.0, 1.8944306), max(0.0, 0.5 * min(min(0.0 - p.z, p.z - -60.0), 16.89443 - length(p.xy)))), max(0.0, length(p.xy) - 11.211139)), 3.5, noiseTex);
  return gslCarry29;
}

fn gsl_remap_25(p: vec3<f32>, material_brushedMetal: mat4x4<f32>, color_gray: vec3<f32>, color_white: vec3<f32>, color_black: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let gsl_tmp_136 = 6.2831855 * p.z / 3.5;
  let gsl_tmp_135 = max(length(p.xy), 0.2);
  let gslCarry26 = thread_paint(vec3<f32>(0.77, 0.78, 0.78), -min(min(clamp((acos(clamp(cos(gsl_tmp_136) * p.x / gsl_tmp_135 + sin(gsl_tmp_136) * p.y / gsl_tmp_135, -1.0, 1.0)) / 3.1415927 * 1.75 - 0.21875) * 1.7320508, 0.0, 1.8944306), max(0.0, 0.5 * min(min(0.0 - p.z, p.z - -60.0), 16.89443 - length(p.xy)))), max(0.0, length(p.xy) - 11.211139)), 3.5, noiseTex);
  return gslCarry26;
}

fn gsl_remap_26(p: vec3<f32>, material_brushedMetal: mat4x4<f32>, color_gray: vec3<f32>, color_white: vec3<f32>, color_black: vec3<f32>, noiseTex: texture_3d<f32>) -> vec4<f32> {
  let gsl_tmp_14 = 6.2831855 * p.z / 0.4;
  let gsl_tmp_13 = max(length(p.xy), 0.2);
  let gslCarry6 = material_withDisplacement(material_brushedMetal, -min(min(clamp((acos(clamp(cos(gsl_tmp_14) * p.x / gsl_tmp_13 + sin(gsl_tmp_14) * p.y / gsl_tmp_13, -1.0, 1.0)) / 3.1415927 * 0.2 - 0.025) * 1.7320508, 0.0, 0.21650636), max(0.0, 0.5 * min(min(0.0 - p.z, p.z - -4.0), 1.2165064 - length(p.xy)))), max(0.0, length(p.xy) - 0.5669873)), noiseTex);
  return gslCarry6[0];
}

fn gsl_remap_27(p: vec3<f32>, material_brushedMetal: mat4x4<f32>, color_gray: vec3<f32>, color_white: vec3<f32>, color_black: vec3<f32>, noiseTex: texture_3d<f32>) -> vec4<f32> {
  let gsl_tmp_14 = 6.2831855 * p.z / 0.4;
  let gsl_tmp_13 = max(length(p.xy), 0.2);
  let gslCarry6 = material_withDisplacement(material_brushedMetal, -min(min(clamp((acos(clamp(cos(gsl_tmp_14) * p.x / gsl_tmp_13 + sin(gsl_tmp_14) * p.y / gsl_tmp_13, -1.0, 1.0)) / 3.1415927 * 0.2 - 0.025) * 1.7320508, 0.0, 0.21650636), max(0.0, 0.5 * min(min(0.0 - p.z, p.z - -4.0), 1.2165064 - length(p.xy)))), max(0.0, length(p.xy) - 0.5669873)), noiseTex);
  return gslCarry6[1];
}

fn gsl_remap_28(p: vec3<f32>, material_brushedMetal: mat4x4<f32>, color_gray: vec3<f32>, color_white: vec3<f32>, color_black: vec3<f32>, noiseTex: texture_3d<f32>) -> vec4<f32> {
  let gsl_tmp_14 = 6.2831855 * p.z / 0.4;
  let gsl_tmp_13 = max(length(p.xy), 0.2);
  let gslCarry6 = material_withDisplacement(material_brushedMetal, -min(min(clamp((acos(clamp(cos(gsl_tmp_14) * p.x / gsl_tmp_13 + sin(gsl_tmp_14) * p.y / gsl_tmp_13, -1.0, 1.0)) / 3.1415927 * 0.2 - 0.025) * 1.7320508, 0.0, 0.21650636), max(0.0, 0.5 * min(min(0.0 - p.z, p.z - -4.0), 1.2165064 - length(p.xy)))), max(0.0, length(p.xy) - 0.5669873)), noiseTex);
  return gslCarry6[2];
}

fn gsl_remap_29(p: vec3<f32>, material_brushedMetal: mat4x4<f32>, color_gray: vec3<f32>, color_white: vec3<f32>, color_black: vec3<f32>, noiseTex: texture_3d<f32>) -> vec4<f32> {
  return vec4<f32>(0.45, 1.0, 1.5, 0.0);
}

fn gsl_remap_30(p: vec3<f32>, material_brushedMetal: mat4x4<f32>, color_gray: vec3<f32>, color_white: vec3<f32>, color_black: vec3<f32>, noiseTex: texture_3d<f32>) -> vec4<f32> {
  return vec4<f32>(0.0, 0.0, 0.0, 0.0);
}

fn gsl_remap_31(p: vec3<f32>, material_brushedMetal: mat4x4<f32>, color_gray: vec3<f32>, color_white: vec3<f32>, color_black: vec3<f32>, material_matte: mat4x4<f32>, material_plastic: mat4x4<f32>, noiseTex: texture_3d<f32>) -> vec4<f32> {
  let gsl_tmp_47 = 6.2831855 * p.z / 0.7;
  let gsl_tmp_46 = max(length(p.xy), 0.2);
  let gslCarry12 = material_withDisplacement(material_brushedMetal, -min(min(clamp((acos(clamp(cos(gsl_tmp_47) * p.x / gsl_tmp_46 + sin(gsl_tmp_47) * p.y / gsl_tmp_46, -1.0, 1.0)) / 3.1415927 * 0.35 - 0.04375) * 1.7320508, 0.0, 0.37888613), max(0.0, 0.5 * min(min(0.0 - p.z, p.z - -8.0), 2.3788862 - length(p.xy)))), max(0.0, length(p.xy) - 1.2422278)), noiseTex);
  return gslCarry12[0];
}

fn gsl_remap_32(p: vec3<f32>, material_brushedMetal: mat4x4<f32>, color_gray: vec3<f32>, color_white: vec3<f32>, color_black: vec3<f32>, material_matte: mat4x4<f32>, material_plastic: mat4x4<f32>, noiseTex: texture_3d<f32>) -> vec4<f32> {
  let gsl_tmp_47 = 6.2831855 * p.z / 0.7;
  let gsl_tmp_46 = max(length(p.xy), 0.2);
  let gslCarry12 = material_withDisplacement(material_brushedMetal, -min(min(clamp((acos(clamp(cos(gsl_tmp_47) * p.x / gsl_tmp_46 + sin(gsl_tmp_47) * p.y / gsl_tmp_46, -1.0, 1.0)) / 3.1415927 * 0.35 - 0.04375) * 1.7320508, 0.0, 0.37888613), max(0.0, 0.5 * min(min(0.0 - p.z, p.z - -8.0), 2.3788862 - length(p.xy)))), max(0.0, length(p.xy) - 1.2422278)), noiseTex);
  return gslCarry12[1];
}

fn gsl_remap_33(p: vec3<f32>, material_brushedMetal: mat4x4<f32>, color_gray: vec3<f32>, color_white: vec3<f32>, color_black: vec3<f32>, material_matte: mat4x4<f32>, material_plastic: mat4x4<f32>, noiseTex: texture_3d<f32>) -> vec4<f32> {
  let gsl_tmp_47 = 6.2831855 * p.z / 0.7;
  let gsl_tmp_46 = max(length(p.xy), 0.2);
  let gslCarry12 = material_withDisplacement(material_brushedMetal, -min(min(clamp((acos(clamp(cos(gsl_tmp_47) * p.x / gsl_tmp_46 + sin(gsl_tmp_47) * p.y / gsl_tmp_46, -1.0, 1.0)) / 3.1415927 * 0.35 - 0.04375) * 1.7320508, 0.0, 0.37888613), max(0.0, 0.5 * min(min(0.0 - p.z, p.z - -8.0), 2.3788862 - length(p.xy)))), max(0.0, length(p.xy) - 1.2422278)), noiseTex);
  return gslCarry12[2];
}

fn gsl_remap_34(p: vec3<f32>, material_brushedMetal: mat4x4<f32>, color_gray: vec3<f32>, color_white: vec3<f32>, color_black: vec3<f32>, material_matte: mat4x4<f32>, material_plastic: mat4x4<f32>, noiseTex: texture_3d<f32>) -> vec4<f32> {
  return vec4<f32>(0.45, 1.0, 1.5, 0.0);
}

fn gsl_remap_35(p: vec3<f32>, material_brushedMetal: mat4x4<f32>, color_gray: vec3<f32>, color_white: vec3<f32>, color_black: vec3<f32>, material_matte: mat4x4<f32>, material_plastic: mat4x4<f32>, noiseTex: texture_3d<f32>) -> vec4<f32> {
  return vec4<f32>(0.0, 0.0, 0.0, 0.0);
}

fn gsl_remap_36(p: vec3<f32>, material_brushedMetal: mat4x4<f32>, color_gray: vec3<f32>, color_white: vec3<f32>, color_black: vec3<f32>, material_matte: mat4x4<f32>, material_plastic: mat4x4<f32>, noiseTex: texture_3d<f32>) -> vec4<f32> {
  let gsl_tmp_80 = 6.2831855 * p.z / 1.25;
  let gsl_tmp_79 = max(length(p.xy), 0.2);
  let gslCarry18 = material_withDisplacement(material_brushedMetal, -min(min(clamp((acos(clamp(cos(gsl_tmp_80) * p.x / gsl_tmp_79 + sin(gsl_tmp_80) * p.y / gsl_tmp_79, -1.0, 1.0)) / 3.1415927 * 0.625 - 0.078125) * 1.7320508, 0.0, 0.6765824), max(0.0, 0.5 * min(min(0.0 - p.z, p.z - -16.0), 4.6765823 - length(p.xy)))), max(0.0, length(p.xy) - 2.6468353)), noiseTex);
  return gslCarry18[0];
}

fn gsl_remap_37(p: vec3<f32>, material_brushedMetal: mat4x4<f32>, color_gray: vec3<f32>, color_white: vec3<f32>, color_black: vec3<f32>, material_matte: mat4x4<f32>, material_plastic: mat4x4<f32>, noiseTex: texture_3d<f32>) -> vec4<f32> {
  let gsl_tmp_80 = 6.2831855 * p.z / 1.25;
  let gsl_tmp_79 = max(length(p.xy), 0.2);
  let gslCarry18 = material_withDisplacement(material_brushedMetal, -min(min(clamp((acos(clamp(cos(gsl_tmp_80) * p.x / gsl_tmp_79 + sin(gsl_tmp_80) * p.y / gsl_tmp_79, -1.0, 1.0)) / 3.1415927 * 0.625 - 0.078125) * 1.7320508, 0.0, 0.6765824), max(0.0, 0.5 * min(min(0.0 - p.z, p.z - -16.0), 4.6765823 - length(p.xy)))), max(0.0, length(p.xy) - 2.6468353)), noiseTex);
  return gslCarry18[1];
}

fn gsl_remap_38(p: vec3<f32>, material_brushedMetal: mat4x4<f32>, color_gray: vec3<f32>, color_white: vec3<f32>, color_black: vec3<f32>, material_matte: mat4x4<f32>, material_plastic: mat4x4<f32>, noiseTex: texture_3d<f32>) -> vec4<f32> {
  let gsl_tmp_80 = 6.2831855 * p.z / 1.25;
  let gsl_tmp_79 = max(length(p.xy), 0.2);
  let gslCarry18 = material_withDisplacement(material_brushedMetal, -min(min(clamp((acos(clamp(cos(gsl_tmp_80) * p.x / gsl_tmp_79 + sin(gsl_tmp_80) * p.y / gsl_tmp_79, -1.0, 1.0)) / 3.1415927 * 0.625 - 0.078125) * 1.7320508, 0.0, 0.6765824), max(0.0, 0.5 * min(min(0.0 - p.z, p.z - -16.0), 4.6765823 - length(p.xy)))), max(0.0, length(p.xy) - 2.6468353)), noiseTex);
  return gslCarry18[2];
}

fn gsl_remap_39(p: vec3<f32>, material_brushedMetal: mat4x4<f32>, color_gray: vec3<f32>, color_white: vec3<f32>, color_black: vec3<f32>, material_matte: mat4x4<f32>, material_plastic: mat4x4<f32>, noiseTex: texture_3d<f32>) -> vec4<f32> {
  let gsl_tmp_113 = 6.2831855 * p.z / 2.0;
  let gsl_tmp_112 = max(length(p.xy), 0.2);
  let gslCarry24 = material_withDisplacement(material_brushedMetal, -min(min(clamp((acos(clamp(cos(gsl_tmp_113) * p.x / gsl_tmp_112 + sin(gsl_tmp_113) * p.y / gsl_tmp_112, -1.0, 1.0)) / 3.1415927 * 1.0 - 0.125) * 1.7320508, 0.0, 1.0825318), max(0.0, 0.5 * min(min(0.0 - p.z, p.z - -32.0), 9.082532 - length(p.xy)))), max(0.0, length(p.xy) - 5.8349366)), noiseTex);
  return gslCarry24[0];
}

fn gsl_remap_40(p: vec3<f32>, material_brushedMetal: mat4x4<f32>, color_gray: vec3<f32>, color_white: vec3<f32>, color_black: vec3<f32>, material_matte: mat4x4<f32>, material_plastic: mat4x4<f32>, noiseTex: texture_3d<f32>) -> vec4<f32> {
  let gsl_tmp_113 = 6.2831855 * p.z / 2.0;
  let gsl_tmp_112 = max(length(p.xy), 0.2);
  let gslCarry24 = material_withDisplacement(material_brushedMetal, -min(min(clamp((acos(clamp(cos(gsl_tmp_113) * p.x / gsl_tmp_112 + sin(gsl_tmp_113) * p.y / gsl_tmp_112, -1.0, 1.0)) / 3.1415927 * 1.0 - 0.125) * 1.7320508, 0.0, 1.0825318), max(0.0, 0.5 * min(min(0.0 - p.z, p.z - -32.0), 9.082532 - length(p.xy)))), max(0.0, length(p.xy) - 5.8349366)), noiseTex);
  return gslCarry24[1];
}

fn gsl_remap_41(p: vec3<f32>, material_brushedMetal: mat4x4<f32>, color_gray: vec3<f32>, color_white: vec3<f32>, color_black: vec3<f32>, material_matte: mat4x4<f32>, material_plastic: mat4x4<f32>, noiseTex: texture_3d<f32>) -> vec4<f32> {
  let gsl_tmp_113 = 6.2831855 * p.z / 2.0;
  let gsl_tmp_112 = max(length(p.xy), 0.2);
  let gslCarry24 = material_withDisplacement(material_brushedMetal, -min(min(clamp((acos(clamp(cos(gsl_tmp_113) * p.x / gsl_tmp_112 + sin(gsl_tmp_113) * p.y / gsl_tmp_112, -1.0, 1.0)) / 3.1415927 * 1.0 - 0.125) * 1.7320508, 0.0, 1.0825318), max(0.0, 0.5 * min(min(0.0 - p.z, p.z - -32.0), 9.082532 - length(p.xy)))), max(0.0, length(p.xy) - 5.8349366)), noiseTex);
  return gslCarry24[2];
}

fn gsl_remap_42(p: vec3<f32>, material_brushedMetal: mat4x4<f32>, color_gray: vec3<f32>, color_white: vec3<f32>, color_black: vec3<f32>, material_matte: mat4x4<f32>, material_plastic: mat4x4<f32>, noiseTex: texture_3d<f32>) -> vec4<f32> {
  let gsl_tmp_146 = 6.2831855 * p.z / 3.5;
  let gsl_tmp_145 = max(length(p.xy), 0.2);
  let gslCarry30 = material_withDisplacement(material_brushedMetal, -min(min(clamp((acos(clamp(cos(gsl_tmp_146) * p.x / gsl_tmp_145 + sin(gsl_tmp_146) * p.y / gsl_tmp_145, -1.0, 1.0)) / 3.1415927 * 1.75 - 0.21875) * 1.7320508, 0.0, 1.8944306), max(0.0, 0.5 * min(min(0.0 - p.z, p.z - -60.0), 16.89443 - length(p.xy)))), max(0.0, length(p.xy) - 11.211139)), noiseTex);
  return gslCarry30[0];
}

fn gsl_remap_43(p: vec3<f32>, material_brushedMetal: mat4x4<f32>, color_gray: vec3<f32>, color_white: vec3<f32>, color_black: vec3<f32>, material_matte: mat4x4<f32>, material_plastic: mat4x4<f32>, noiseTex: texture_3d<f32>) -> vec4<f32> {
  let gsl_tmp_146 = 6.2831855 * p.z / 3.5;
  let gsl_tmp_145 = max(length(p.xy), 0.2);
  let gslCarry30 = material_withDisplacement(material_brushedMetal, -min(min(clamp((acos(clamp(cos(gsl_tmp_146) * p.x / gsl_tmp_145 + sin(gsl_tmp_146) * p.y / gsl_tmp_145, -1.0, 1.0)) / 3.1415927 * 1.75 - 0.21875) * 1.7320508, 0.0, 1.8944306), max(0.0, 0.5 * min(min(0.0 - p.z, p.z - -60.0), 16.89443 - length(p.xy)))), max(0.0, length(p.xy) - 11.211139)), noiseTex);
  return gslCarry30[1];
}

fn gsl_remap_44(p: vec3<f32>, material_brushedMetal: mat4x4<f32>, color_gray: vec3<f32>, color_white: vec3<f32>, color_black: vec3<f32>, material_matte: mat4x4<f32>, material_plastic: mat4x4<f32>, noiseTex: texture_3d<f32>) -> vec4<f32> {
  let gsl_tmp_146 = 6.2831855 * p.z / 3.5;
  let gsl_tmp_145 = max(length(p.xy), 0.2);
  let gslCarry30 = material_withDisplacement(material_brushedMetal, -min(min(clamp((acos(clamp(cos(gsl_tmp_146) * p.x / gsl_tmp_145 + sin(gsl_tmp_146) * p.y / gsl_tmp_145, -1.0, 1.0)) / 3.1415927 * 1.75 - 0.21875) * 1.7320508, 0.0, 1.8944306), max(0.0, 0.5 * min(min(0.0 - p.z, p.z - -60.0), 16.89443 - length(p.xy)))), max(0.0, length(p.xy) - 11.211139)), noiseTex);
  return gslCarry30[2];
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

fn playful_checkerboard(p: vec3<f32>, cellSize: f32, colorA: vec3<f32>, colorB: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  var gsl_simpleValue = color_mixOklab(colorA, colorB, 0.5, noiseTex);
  if (gsl_simpleDetail > 0.0) {
    let x = texture_squareWave(p.x, 2.0 * cellSize, noiseTex);
    let y = texture_squareWave(p.y, 2.0 * cellSize, noiseTex);
    let xy = x + y - 2.0 * x * y;
    let z = texture_squareWave(p.z, 2.0 * cellSize, noiseTex);
    let odd = xy + z - 2.0 * xy * z;
    gsl_simpleValue = mix(gsl_simpleValue, color_mixOklab(colorA, colorB, odd, noiseTex), gsl_simpleDetail);
  }
  let board = gsl_simpleValue;
  return board;
}

fn fastener_screw(head: f32, shank: f32, drive: f32, noiseTex: texture_3d<f32>) -> f32 {
  return max(min(head, shank), -drive);
}

fn knob(p: vec3<f32>, r: f32, noiseTex: texture_3d<f32>) -> f32 {
  let groove = builtin_torus(p, 0.8 * r, 0.125 * r, noiseTex);
  let dimples = min(min(builtin_sphere(p, vec3<f32>(0.6 * r, 0.0, 0.8 * r), 0.2 * r, noiseTex), builtin_sphere(p, vec3<f32>(-0.6 * r, 0.0, 0.8 * r), 0.2 * r, noiseTex)), min(builtin_sphere(p, vec3<f32>(0.0, 0.6 * r, 0.8 * r), 0.2 * r, noiseTex), builtin_sphere(p, vec3<f32>(0.0, -0.6 * r, 0.8 * r), 0.2 * r, noiseTex)));
  return max(builtin_sphere_2(p, r, noiseTex), -min(groove, dimples));
}

fn gsl_shapeParts(p: vec3<f32>, time: f32, noiseTex: texture_3d<f32>) -> array<f32, 15> {
  gsl_partNear = vec3<f32>(1e10, 1e10, 1e10);
  let gsl_partLo = vec3<f32>(-81.91, -62.91, -5.01);
  let gsl_partHi = vec3<f32>(-78.09, -59.09, 1.01);
  let gsl_partBox = gsl_boxDistance(p, gsl_partLo, gsl_partHi);
  var gsl_part = gsl_partBox;
  let material_brushedMetal = mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
  if (gsl_vanishView.w <= 0.0 || gsl_partBox <= gsl_vanishView.w) {
    let gsl_partVanish = gsl_vanish(gsl_partLo, gsl_partHi);
    gsl_part = 1e10;
    let gsl_simpleWeight_16 = gsl_simple(vec3<f32>(-81.91, -62.91, -5.01), vec3<f32>(-78.09, -59.09, 1.01));
    var gsl_simplePart = 1e10;
    if (gsl_simpleWeight_16 > 0.0) {
      gsl_simplePart = gsl_remap(p - vec3<f32>(-80.0, -61.0, -1.0), material_brushedMetal, noiseTex);
    }
    if (gsl_simpleWeight_16 < 1.0) {
      gsl_simplePart = mix(max(gsl_remap_2(p - vec3<f32>(-80.0, -61.0, -1.0), material_brushedMetal, noiseTex), gsl_boxSdf(p, vec3<f32>(-81.91, -62.91, -5.01), vec3<f32>(-78.09, -59.09, 1.01))), gsl_simplePart, gsl_simpleWeight_16);
    }
    if (gsl_partVanish.x > 0.0) { gsl_part = gsl_simplePart + gsl_partVanish.y; }
    if (abs(gsl_part) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part), gsl_partVanish.zw); }
  }
  let screw4 = gsl_part;
  let gsl_partLo_2 = vec3<f32>(-2.0, -63.0, -2.0);
  let gsl_partHi_2 = vec3<f32>(2.0, -59.0, 2.0);
  let gsl_partBox_2 = gsl_boxDistance(p, gsl_partLo_2, gsl_partHi_2);
  var gsl_part_2 = gsl_partBox_2;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_2 <= gsl_vanishView.w) {
    let gsl_partVanish_2 = gsl_vanish(gsl_partLo_2, gsl_partHi_2);
    gsl_part_2 = 1e10;
    let gsl_simpleWeight_17 = gsl_simple(vec3<f32>(-2.0, -63.0, -2.0), vec3<f32>(2.0, -59.0, 2.0));
    var gsl_simplePart_2 = 1e10;
    if (gsl_simpleWeight_17 > 0.0) {
      gsl_simplePart_2 = knob(p - vec3<f32>(0.0, -61.0, 0.0), 2.0, noiseTex);
    }
    if (gsl_simpleWeight_17 < 1.0) {
      gsl_simplePart_2 = mix(max(builtin_sphere_2(p - vec3<f32>(0.0, -61.0, 0.0), 2.0, noiseTex), gsl_boxSdf(p, vec3<f32>(-2.0, -63.0, -2.0), vec3<f32>(2.0, -59.0, 2.0))), gsl_simplePart_2, gsl_simpleWeight_17);
    }
    if (gsl_partVanish_2.x > 0.0) { gsl_part_2 = gsl_simplePart_2 + gsl_partVanish_2.y; }
    if (abs(gsl_part_2) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_2), gsl_partVanish_2.zw); }
  }
  let checker4 = gsl_part_2;
  let gsl_partLo_3 = vec3<f32>(78.0, -63.0, -2.0);
  let gsl_partHi_3 = vec3<f32>(82.0, -59.0, 2.0);
  let gsl_partBox_3 = gsl_boxDistance(p, gsl_partLo_3, gsl_partHi_3);
  var gsl_part_3 = gsl_partBox_3;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_3 <= gsl_vanishView.w) {
    let gsl_partVanish_3 = gsl_vanish(gsl_partLo_3, gsl_partHi_3);
    gsl_part_3 = 1e10;
    let gsl_simpleWeight_18 = gsl_simple(vec3<f32>(78.0, -63.0, -2.0), vec3<f32>(82.0, -59.0, 2.0));
    var gsl_simplePart_3 = 1e10;
    if (gsl_simpleWeight_18 > 0.0) {
      gsl_simplePart_3 = knob(p - vec3<f32>(80.0, -61.0, 0.0), 2.0, noiseTex);
    }
    if (gsl_simpleWeight_18 < 1.0) {
      gsl_simplePart_3 = mix(max(builtin_sphere_2(p - vec3<f32>(80.0, -61.0, 0.0), 2.0, noiseTex), gsl_boxSdf(p, vec3<f32>(78.0, -63.0, -2.0), vec3<f32>(82.0, -59.0, 2.0))), gsl_simplePart_3, gsl_simpleWeight_18);
    }
    if (gsl_partVanish_3.x > 0.0) { gsl_part_3 = gsl_simplePart_3 + gsl_partVanish_3.y; }
    if (abs(gsl_part_3) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_3), gsl_partVanish_3.zw); }
  }
  let relief4 = gsl_part_3;
  let gsl_partLo_4 = vec3<f32>(-83.51, -50.51, -10.01);
  let gsl_partHi_4 = vec3<f32>(-76.49, -43.49, 2.01);
  let gsl_partBox_4 = gsl_boxDistance(p, gsl_partLo_4, gsl_partHi_4);
  var gsl_part_4 = gsl_partBox_4;
  let color_gray = vec3<f32>(0.5, 0.5, 0.5);
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_4 <= gsl_vanishView.w) {
    let gsl_partVanish_4 = gsl_vanish(gsl_partLo_4, gsl_partHi_4);
    gsl_part_4 = 1e10;
    let gsl_simpleWeight_19 = gsl_simple(vec3<f32>(-83.51, -50.51, -10.01), vec3<f32>(-76.49, -43.49, 2.01));
    var gsl_simplePart_4 = 1e10;
    if (gsl_simpleWeight_19 > 0.0) {
      gsl_simplePart_4 = gsl_remap_4(p - vec3<f32>(-80.0, -47.0, -2.0), material_brushedMetal, color_gray, noiseTex);
    }
    if (gsl_simpleWeight_19 < 1.0) {
      gsl_simplePart_4 = mix(max(gsl_remap_5(p - vec3<f32>(-80.0, -47.0, -2.0), material_brushedMetal, color_gray, noiseTex), gsl_boxSdf(p, vec3<f32>(-83.51, -50.51, -10.01), vec3<f32>(-76.49, -43.49, 2.01))), gsl_simplePart_4, gsl_simpleWeight_19);
    }
    if (gsl_partVanish_4.x > 0.0) { gsl_part_4 = gsl_simplePart_4 + gsl_partVanish_4.y; }
    if (abs(gsl_part_4) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_4), gsl_partVanish_4.zw); }
  }
  let screw8 = gsl_part_4;
  let gsl_partLo_5 = vec3<f32>(-4.0, -51.0, -4.0);
  let gsl_partHi_5 = vec3<f32>(4.0, -43.0, 4.0);
  let gsl_partBox_5 = gsl_boxDistance(p, gsl_partLo_5, gsl_partHi_5);
  var gsl_part_5 = gsl_partBox_5;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_5 <= gsl_vanishView.w) {
    let gsl_partVanish_5 = gsl_vanish(gsl_partLo_5, gsl_partHi_5);
    gsl_part_5 = 1e10;
    let gsl_simpleWeight_20 = gsl_simple(vec3<f32>(-4.0, -51.0, -4.0), vec3<f32>(4.0, -43.0, 4.0));
    var gsl_simplePart_5 = 1e10;
    if (gsl_simpleWeight_20 > 0.0) {
      gsl_simplePart_5 = knob(p - vec3<f32>(0.0, -47.0, 0.0), 4.0, noiseTex);
    }
    if (gsl_simpleWeight_20 < 1.0) {
      gsl_simplePart_5 = mix(max(builtin_sphere_2(p - vec3<f32>(0.0, -47.0, 0.0), 4.0, noiseTex), gsl_boxSdf(p, vec3<f32>(-4.0, -51.0, -4.0), vec3<f32>(4.0, -43.0, 4.0))), gsl_simplePart_5, gsl_simpleWeight_20);
    }
    if (gsl_partVanish_5.x > 0.0) { gsl_part_5 = gsl_simplePart_5 + gsl_partVanish_5.y; }
    if (abs(gsl_part_5) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_5), gsl_partVanish_5.zw); }
  }
  let checker8 = gsl_part_5;
  let gsl_partLo_6 = vec3<f32>(76.0, -51.0, -4.0);
  let gsl_partHi_6 = vec3<f32>(84.0, -43.0, 4.0);
  let gsl_partBox_6 = gsl_boxDistance(p, gsl_partLo_6, gsl_partHi_6);
  var gsl_part_6 = gsl_partBox_6;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_6 <= gsl_vanishView.w) {
    let gsl_partVanish_6 = gsl_vanish(gsl_partLo_6, gsl_partHi_6);
    gsl_part_6 = 1e10;
    let gsl_simpleWeight_21 = gsl_simple(vec3<f32>(76.0, -51.0, -4.0), vec3<f32>(84.0, -43.0, 4.0));
    var gsl_simplePart_6 = 1e10;
    if (gsl_simpleWeight_21 > 0.0) {
      gsl_simplePart_6 = knob(p - vec3<f32>(80.0, -47.0, 0.0), 4.0, noiseTex);
    }
    if (gsl_simpleWeight_21 < 1.0) {
      gsl_simplePart_6 = mix(max(builtin_sphere_2(p - vec3<f32>(80.0, -47.0, 0.0), 4.0, noiseTex), gsl_boxSdf(p, vec3<f32>(76.0, -51.0, -4.0), vec3<f32>(84.0, -43.0, 4.0))), gsl_simplePart_6, gsl_simpleWeight_21);
    }
    if (gsl_partVanish_6.x > 0.0) { gsl_part_6 = gsl_simplePart_6 + gsl_partVanish_6.y; }
    if (abs(gsl_part_6) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_6), gsl_partVanish_6.zw); }
  }
  let relief8 = gsl_part_6;
  let gsl_partLo_7 = vec3<f32>(-86.51, -33.51, -20.01);
  let gsl_partHi_7 = vec3<f32>(-73.49, -20.49, 4.01);
  let gsl_partBox_7 = gsl_boxDistance(p, gsl_partLo_7, gsl_partHi_7);
  var gsl_part_7 = gsl_partBox_7;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_7 <= gsl_vanishView.w) {
    let gsl_partVanish_7 = gsl_vanish(gsl_partLo_7, gsl_partHi_7);
    gsl_part_7 = 1e10;
    let gsl_simpleWeight_22 = gsl_simple(vec3<f32>(-86.51, -33.51, -20.01), vec3<f32>(-73.49, -20.49, 4.01));
    var gsl_simplePart_7 = 1e10;
    if (gsl_simpleWeight_22 > 0.0) {
      gsl_simplePart_7 = gsl_remap_7(p - vec3<f32>(-80.0, -27.0, -4.0), material_brushedMetal, color_gray, noiseTex);
    }
    if (gsl_simpleWeight_22 < 1.0) {
      gsl_simplePart_7 = mix(max(gsl_remap_8(p - vec3<f32>(-80.0, -27.0, -4.0), material_brushedMetal, color_gray, noiseTex), gsl_boxSdf(p, vec3<f32>(-86.51, -33.51, -20.01), vec3<f32>(-73.49, -20.49, 4.01))), gsl_simplePart_7, gsl_simpleWeight_22);
    }
    if (gsl_partVanish_7.x > 0.0) { gsl_part_7 = gsl_simplePart_7 + gsl_partVanish_7.y; }
    if (abs(gsl_part_7) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_7), gsl_partVanish_7.zw); }
  }
  let screw16 = gsl_part_7;
  let gsl_partLo_8 = vec3<f32>(-8.0, -35.0, -8.0);
  let gsl_partHi_8 = vec3<f32>(8.0, -19.0, 8.0);
  let gsl_partBox_8 = gsl_boxDistance(p, gsl_partLo_8, gsl_partHi_8);
  var gsl_part_8 = gsl_partBox_8;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_8 <= gsl_vanishView.w) {
    let gsl_partVanish_8 = gsl_vanish(gsl_partLo_8, gsl_partHi_8);
    gsl_part_8 = 1e10;
    let gsl_simpleWeight_23 = gsl_simple(vec3<f32>(-8.0, -35.0, -8.0), vec3<f32>(8.0, -19.0, 8.0));
    var gsl_simplePart_8 = 1e10;
    if (gsl_simpleWeight_23 > 0.0) {
      gsl_simplePart_8 = knob(p - vec3<f32>(0.0, -27.0, 0.0), 8.0, noiseTex);
    }
    if (gsl_simpleWeight_23 < 1.0) {
      gsl_simplePart_8 = mix(max(builtin_sphere_2(p - vec3<f32>(0.0, -27.0, 0.0), 8.0, noiseTex), gsl_boxSdf(p, vec3<f32>(-8.0, -35.0, -8.0), vec3<f32>(8.0, -19.0, 8.0))), gsl_simplePart_8, gsl_simpleWeight_23);
    }
    if (gsl_partVanish_8.x > 0.0) { gsl_part_8 = gsl_simplePart_8 + gsl_partVanish_8.y; }
    if (abs(gsl_part_8) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_8), gsl_partVanish_8.zw); }
  }
  let checker16 = gsl_part_8;
  let gsl_partLo_9 = vec3<f32>(72.0, -35.0, -8.0);
  let gsl_partHi_9 = vec3<f32>(88.0, -19.0, 8.0);
  let gsl_partBox_9 = gsl_boxDistance(p, gsl_partLo_9, gsl_partHi_9);
  var gsl_part_9 = gsl_partBox_9;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_9 <= gsl_vanishView.w) {
    let gsl_partVanish_9 = gsl_vanish(gsl_partLo_9, gsl_partHi_9);
    gsl_part_9 = 1e10;
    let gsl_simpleWeight_24 = gsl_simple(vec3<f32>(72.0, -35.0, -8.0), vec3<f32>(88.0, -19.0, 8.0));
    var gsl_simplePart_9 = 1e10;
    if (gsl_simpleWeight_24 > 0.0) {
      gsl_simplePart_9 = knob(p - vec3<f32>(80.0, -27.0, 0.0), 8.0, noiseTex);
    }
    if (gsl_simpleWeight_24 < 1.0) {
      gsl_simplePart_9 = mix(max(builtin_sphere_2(p - vec3<f32>(80.0, -27.0, 0.0), 8.0, noiseTex), gsl_boxSdf(p, vec3<f32>(72.0, -35.0, -8.0), vec3<f32>(88.0, -19.0, 8.0))), gsl_simplePart_9, gsl_simpleWeight_24);
    }
    if (gsl_partVanish_9.x > 0.0) { gsl_part_9 = gsl_simplePart_9 + gsl_partVanish_9.y; }
    if (abs(gsl_part_9) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_9), gsl_partVanish_9.zw); }
  }
  let relief16 = gsl_part_9;
  let gsl_partLo_10 = vec3<f32>(-92.01, -7.01, -40.01);
  let gsl_partHi_10 = vec3<f32>(-67.99, 17.01, 8.01);
  let gsl_partBox_10 = gsl_boxDistance(p, gsl_partLo_10, gsl_partHi_10);
  var gsl_part_10 = gsl_partBox_10;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_10 <= gsl_vanishView.w) {
    let gsl_partVanish_10 = gsl_vanish(gsl_partLo_10, gsl_partHi_10);
    gsl_part_10 = 1e10;
    let gsl_simpleWeight_25 = gsl_simple(vec3<f32>(-92.01, -7.01, -40.01), vec3<f32>(-67.99, 17.01, 8.01));
    var gsl_simplePart_10 = 1e10;
    if (gsl_simpleWeight_25 > 0.0) {
      gsl_simplePart_10 = gsl_remap_10(p - vec3<f32>(-80.0, 5.0, -8.0), material_brushedMetal, color_gray, noiseTex);
    }
    if (gsl_simpleWeight_25 < 1.0) {
      gsl_simplePart_10 = mix(max(gsl_remap_11(p - vec3<f32>(-80.0, 5.0, -8.0), material_brushedMetal, color_gray, noiseTex), gsl_boxSdf(p, vec3<f32>(-92.01, -7.01, -40.01), vec3<f32>(-67.99, 17.01, 8.01))), gsl_simplePart_10, gsl_simpleWeight_25);
    }
    if (gsl_partVanish_10.x > 0.0) { gsl_part_10 = gsl_simplePart_10 + gsl_partVanish_10.y; }
    if (abs(gsl_part_10) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_10), gsl_partVanish_10.zw); }
  }
  let screw32 = gsl_part_10;
  let gsl_partLo_11 = vec3<f32>(-16.0, -11.0, -16.0);
  let gsl_partHi_11 = vec3<f32>(16.0, 21.0, 16.0);
  let gsl_partBox_11 = gsl_boxDistance(p, gsl_partLo_11, gsl_partHi_11);
  var gsl_part_11 = gsl_partBox_11;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_11 <= gsl_vanishView.w) {
    let gsl_partVanish_11 = gsl_vanish(gsl_partLo_11, gsl_partHi_11);
    gsl_part_11 = 1e10;
    let gsl_simpleWeight_26 = gsl_simple(vec3<f32>(-16.0, -11.0, -16.0), vec3<f32>(16.0, 21.0, 16.0));
    var gsl_simplePart_11 = 1e10;
    if (gsl_simpleWeight_26 > 0.0) {
      gsl_simplePart_11 = knob(p - vec3<f32>(0.0, 5.0, 0.0), 16.0, noiseTex);
    }
    if (gsl_simpleWeight_26 < 1.0) {
      gsl_simplePart_11 = mix(max(builtin_sphere_2(p - vec3<f32>(0.0, 5.0, 0.0), 16.0, noiseTex), gsl_boxSdf(p, vec3<f32>(-16.0, -11.0, -16.0), vec3<f32>(16.0, 21.0, 16.0))), gsl_simplePart_11, gsl_simpleWeight_26);
    }
    if (gsl_partVanish_11.x > 0.0) { gsl_part_11 = gsl_simplePart_11 + gsl_partVanish_11.y; }
    if (abs(gsl_part_11) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_11), gsl_partVanish_11.zw); }
  }
  let checker32 = gsl_part_11;
  let gsl_partLo_12 = vec3<f32>(64.0, -11.0, -16.0);
  let gsl_partHi_12 = vec3<f32>(96.0, 21.0, 16.0);
  let gsl_partBox_12 = gsl_boxDistance(p, gsl_partLo_12, gsl_partHi_12);
  var gsl_part_12 = gsl_partBox_12;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_12 <= gsl_vanishView.w) {
    let gsl_partVanish_12 = gsl_vanish(gsl_partLo_12, gsl_partHi_12);
    gsl_part_12 = 1e10;
    let gsl_simpleWeight_27 = gsl_simple(vec3<f32>(64.0, -11.0, -16.0), vec3<f32>(96.0, 21.0, 16.0));
    var gsl_simplePart_12 = 1e10;
    if (gsl_simpleWeight_27 > 0.0) {
      gsl_simplePart_12 = knob(p - vec3<f32>(80.0, 5.0, 0.0), 16.0, noiseTex);
    }
    if (gsl_simpleWeight_27 < 1.0) {
      gsl_simplePart_12 = mix(max(builtin_sphere_2(p - vec3<f32>(80.0, 5.0, 0.0), 16.0, noiseTex), gsl_boxSdf(p, vec3<f32>(64.0, -11.0, -16.0), vec3<f32>(96.0, 21.0, 16.0))), gsl_simplePart_12, gsl_simpleWeight_27);
    }
    if (gsl_partVanish_12.x > 0.0) { gsl_part_12 = gsl_simplePart_12 + gsl_partVanish_12.y; }
    if (abs(gsl_part_12) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_12), gsl_partVanish_12.zw); }
  }
  let relief32 = gsl_part_12;
  let gsl_partLo_13 = vec3<f32>(-102.51, 38.49, -75.01);
  let gsl_partHi_13 = vec3<f32>(-57.49, 83.51, 15.01);
  let gsl_partBox_13 = gsl_boxDistance(p, gsl_partLo_13, gsl_partHi_13);
  var gsl_part_13 = gsl_partBox_13;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_13 <= gsl_vanishView.w) {
    let gsl_partVanish_13 = gsl_vanish(gsl_partLo_13, gsl_partHi_13);
    gsl_part_13 = 1e10;
    let gsl_simpleWeight_28 = gsl_simple(vec3<f32>(-102.51, 38.49, -75.01), vec3<f32>(-57.49, 83.51, 15.01));
    var gsl_simplePart_13 = 1e10;
    if (gsl_simpleWeight_28 > 0.0) {
      gsl_simplePart_13 = gsl_remap_13(p - vec3<f32>(-80.0, 61.0, -15.0), material_brushedMetal, color_gray, noiseTex);
    }
    if (gsl_simpleWeight_28 < 1.0) {
      gsl_simplePart_13 = mix(max(gsl_remap_14(p - vec3<f32>(-80.0, 61.0, -15.0), material_brushedMetal, color_gray, noiseTex), gsl_boxSdf(p, vec3<f32>(-102.51, 38.49, -75.01), vec3<f32>(-57.49, 83.51, 15.01))), gsl_simplePart_13, gsl_simpleWeight_28);
    }
    if (gsl_partVanish_13.x > 0.0) { gsl_part_13 = gsl_simplePart_13 + gsl_partVanish_13.y; }
    if (abs(gsl_part_13) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_13), gsl_partVanish_13.zw); }
  }
  let screw64 = gsl_part_13;
  let gsl_partLo_14 = vec3<f32>(-32.0, 29.0, -32.0);
  let gsl_partHi_14 = vec3<f32>(32.0, 93.0, 32.0);
  let gsl_partBox_14 = gsl_boxDistance(p, gsl_partLo_14, gsl_partHi_14);
  var gsl_part_14 = gsl_partBox_14;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_14 <= gsl_vanishView.w) {
    let gsl_partVanish_14 = gsl_vanish(gsl_partLo_14, gsl_partHi_14);
    gsl_part_14 = 1e10;
    let gsl_simpleWeight_29 = gsl_simple(vec3<f32>(-32.0, 29.0, -32.0), vec3<f32>(32.0, 93.0, 32.0));
    var gsl_simplePart_14 = 1e10;
    if (gsl_simpleWeight_29 > 0.0) {
      gsl_simplePart_14 = knob(p - vec3<f32>(0.0, 61.0, 0.0), 32.0, noiseTex);
    }
    if (gsl_simpleWeight_29 < 1.0) {
      gsl_simplePart_14 = mix(max(builtin_sphere_2(p - vec3<f32>(0.0, 61.0, 0.0), 32.0, noiseTex), gsl_boxSdf(p, vec3<f32>(-32.0, 29.0, -32.0), vec3<f32>(32.0, 93.0, 32.0))), gsl_simplePart_14, gsl_simpleWeight_29);
    }
    if (gsl_partVanish_14.x > 0.0) { gsl_part_14 = gsl_simplePart_14 + gsl_partVanish_14.y; }
    if (abs(gsl_part_14) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_14), gsl_partVanish_14.zw); }
  }
  let checker64 = gsl_part_14;
  let gsl_partLo_15 = vec3<f32>(48.0, 29.0, -32.0);
  let gsl_partHi_15 = vec3<f32>(112.0, 93.0, 32.0);
  let gsl_partBox_15 = gsl_boxDistance(p, gsl_partLo_15, gsl_partHi_15);
  var gsl_part_15 = gsl_partBox_15;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_15 <= gsl_vanishView.w) {
    let gsl_partVanish_15 = gsl_vanish(gsl_partLo_15, gsl_partHi_15);
    gsl_part_15 = 1e10;
    let gsl_simpleWeight_30 = gsl_simple(vec3<f32>(48.0, 29.0, -32.0), vec3<f32>(112.0, 93.0, 32.0));
    var gsl_simplePart_15 = 1e10;
    if (gsl_simpleWeight_30 > 0.0) {
      gsl_simplePart_15 = knob(p - vec3<f32>(80.0, 61.0, 0.0), 32.0, noiseTex);
    }
    if (gsl_simpleWeight_30 < 1.0) {
      gsl_simplePart_15 = mix(max(builtin_sphere_2(p - vec3<f32>(80.0, 61.0, 0.0), 32.0, noiseTex), gsl_boxSdf(p, vec3<f32>(48.0, 29.0, -32.0), vec3<f32>(112.0, 93.0, 32.0))), gsl_simplePart_15, gsl_simpleWeight_30);
    }
    if (gsl_partVanish_15.x > 0.0) { gsl_part_15 = gsl_simplePart_15 + gsl_partVanish_15.y; }
    if (abs(gsl_part_15) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_15), gsl_partVanish_15.zw); }
  }
  let relief64 = gsl_part_15;
  return array<f32, 15>(screw4, checker4, relief4, screw8, checker8, relief8, screw16, checker16, relief16, screw32, checker32, relief32, screw64, checker64, relief64);
}

fn sdShape(p: vec3<f32>, time: f32, noiseTex: texture_3d<f32>) -> gsl_shape {
  let gsl_d = gsl_shapeParts(p, time, noiseTex);
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
  return gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_s1, gsl_s2), gsl_s3), gsl_s4), gsl_s5), gsl_s6), gsl_s7), gsl_s8), gsl_s9), gsl_s10), gsl_s11), gsl_s12), gsl_s13), gsl_s14), gsl_s15);
}
fn gsl_colorField(p: vec3<f32>, time: f32, gsl_colorIdx: u32, noiseTex: texture_3d<f32>) -> vec4<f32> {
  if (gsl_colorIdx == 1u) {
    let material_brushedMetal = mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    let color_gray = vec3<f32>(0.5, 0.5, 0.5);
    let gsl_simpleWeight = gsl_simple(vec3<f32>(-81.91, -62.91, -5.01), vec3<f32>(-78.09, -59.09, 1.01));
    var gsl_simpleField = clamp(gsl_remap_17(p - vec3<f32>(-80.0, -61.0, -1.0), material_brushedMetal, color_gray, noiseTex) * 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0));
    if (gsl_simpleWeight > 0.0) {
      gsl_simpleField = mix(gsl_simpleField, clamp(gsl_remap_16(p - vec3<f32>(-80.0, -61.0, -1.0), material_brushedMetal, color_gray, noiseTex) * 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0)), gsl_simpleWeight);
    }
    return vec4<f32>(gsl_simpleField, 1.0);
  }
  if (gsl_colorIdx == 2u) {
    let color_white = vec3<f32>(1.0, 1.0, 1.0);
    let color_black = vec3<f32>(0.0, 0.0, 0.0);
    return vec4<f32>(playful_checkerboard(p, 0.5, color_white, color_black, noiseTex), 1.0);
  }
  if (gsl_colorIdx == 3u) {
    var gsl_simpleValue = vec3<f32>(0.37482947, 0.61767787, 0.7900267);
    if (gsl_simpleDetail > 0.0) {
      let material_brushedMetal = mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
      gsl_simpleValue = mix(gsl_simpleValue, gsl_remap_3(vec3<f32>(dot(p, vec3<f32>(1.0, 0.0, 0.0)), dot(p, vec3<f32>(0.0, 1.0, 0.0)), dot(p, vec3<f32>(0.0, 0.0, 1.0))), material_brushedMetal, noiseTex), gsl_simpleDetail);
    }
    let dots4 = gsl_simpleValue;
    var gsl_simpleValue_2 = dots4;
    if (gsl_simpleDetail > 0.0) {
      let color_gray = vec3<f32>(0.5, 0.5, 0.5);
      gsl_simpleValue_2 = mix(gsl_simpleValue_2, color_mixOklab(dots4, color_gray, 0.3 + 0.15 * noise_gradientLattice(p / 0.25, noiseTex), noiseTex), gsl_simpleDetail);
    }
    let reliefColor4 = gsl_simpleValue_2;
    return vec4<f32>(reliefColor4, 1.0);
  }
  if (gsl_colorIdx == 4u) {
    let material_brushedMetal = mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    let color_gray = vec3<f32>(0.5, 0.5, 0.5);
    let color_white = vec3<f32>(1.0, 1.0, 1.0);
    let color_black = vec3<f32>(0.0, 0.0, 0.0);
    let gsl_simpleWeight_2 = gsl_simple(vec3<f32>(-83.51, -50.51, -10.01), vec3<f32>(-76.49, -43.49, 2.01));
    var gsl_simpleField_2 = clamp(gsl_remap_19(p - vec3<f32>(-80.0, -47.0, -2.0), material_brushedMetal, color_gray, color_white, color_black, noiseTex) * 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0));
    if (gsl_simpleWeight_2 > 0.0) {
      gsl_simpleField_2 = mix(gsl_simpleField_2, clamp(gsl_remap_18(p - vec3<f32>(-80.0, -47.0, -2.0), material_brushedMetal, color_gray, color_white, color_black, noiseTex) * 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0)), gsl_simpleWeight_2);
    }
    return vec4<f32>(gsl_simpleField_2, 1.0);
  }
  if (gsl_colorIdx == 5u) {
    let color_white = vec3<f32>(1.0, 1.0, 1.0);
    let color_black = vec3<f32>(0.0, 0.0, 0.0);
    return vec4<f32>(playful_checkerboard(p, 1.0, color_white, color_black, noiseTex), 1.0);
  }
  if (gsl_colorIdx == 6u) {
    let color_gray = vec3<f32>(0.5, 0.5, 0.5);
    var gsl_simpleValue_3 = vec3<f32>(0.37482947, 0.61767787, 0.7900267);
    if (gsl_simpleDetail > 0.0) {
      let material_brushedMetal = mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
      gsl_simpleValue_3 = mix(gsl_simpleValue_3, gsl_remap_6(vec3<f32>(dot(p, vec3<f32>(1.0, 0.0, 0.0)), dot(p, vec3<f32>(0.0, 1.0, 0.0)), dot(p, vec3<f32>(0.0, 0.0, 1.0))), material_brushedMetal, color_gray, noiseTex), gsl_simpleDetail);
    }
    let dots8 = gsl_simpleValue_3;
    var gsl_simpleValue_4 = dots8;
    if (gsl_simpleDetail > 0.0) {
      gsl_simpleValue_4 = mix(gsl_simpleValue_4, color_mixOklab(dots8, color_gray, 0.3 + 0.15 * noise_gradientLattice(p / 0.5, noiseTex), noiseTex), gsl_simpleDetail);
    }
    let reliefColor8 = gsl_simpleValue_4;
    return vec4<f32>(reliefColor8, 1.0);
  }
  if (gsl_colorIdx == 7u) {
    let material_brushedMetal = mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    let color_gray = vec3<f32>(0.5, 0.5, 0.5);
    let color_white = vec3<f32>(1.0, 1.0, 1.0);
    let color_black = vec3<f32>(0.0, 0.0, 0.0);
    let gsl_simpleWeight_3 = gsl_simple(vec3<f32>(-86.51, -33.51, -20.01), vec3<f32>(-73.49, -20.49, 4.01));
    var gsl_simpleField_3 = clamp(gsl_remap_21(p - vec3<f32>(-80.0, -27.0, -4.0), material_brushedMetal, color_gray, color_white, color_black, noiseTex) * 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0));
    if (gsl_simpleWeight_3 > 0.0) {
      gsl_simpleField_3 = mix(gsl_simpleField_3, clamp(gsl_remap_20(p - vec3<f32>(-80.0, -27.0, -4.0), material_brushedMetal, color_gray, color_white, color_black, noiseTex) * 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0)), gsl_simpleWeight_3);
    }
    return vec4<f32>(gsl_simpleField_3, 1.0);
  }
  if (gsl_colorIdx == 8u) {
    let color_white = vec3<f32>(1.0, 1.0, 1.0);
    let color_black = vec3<f32>(0.0, 0.0, 0.0);
    return vec4<f32>(playful_checkerboard(p, 2.0, color_white, color_black, noiseTex), 1.0);
  }
  if (gsl_colorIdx == 9u) {
    let color_gray = vec3<f32>(0.5, 0.5, 0.5);
    var gsl_simpleValue_5 = vec3<f32>(0.37482947, 0.61767787, 0.7900267);
    if (gsl_simpleDetail > 0.0) {
      let material_brushedMetal = mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
      gsl_simpleValue_5 = mix(gsl_simpleValue_5, gsl_remap_9(vec3<f32>(dot(p, vec3<f32>(1.0, 0.0, 0.0)), dot(p, vec3<f32>(0.0, 1.0, 0.0)), dot(p, vec3<f32>(0.0, 0.0, 1.0))), material_brushedMetal, color_gray, noiseTex), gsl_simpleDetail);
    }
    let dots16 = gsl_simpleValue_5;
    var gsl_simpleValue_6 = dots16;
    if (gsl_simpleDetail > 0.0) {
      gsl_simpleValue_6 = mix(gsl_simpleValue_6, color_mixOklab(dots16, color_gray, 0.3 + 0.15 * noise_gradientLattice(p / 1.0, noiseTex), noiseTex), gsl_simpleDetail);
    }
    let reliefColor16 = gsl_simpleValue_6;
    return vec4<f32>(reliefColor16, 1.0);
  }
  if (gsl_colorIdx == 10u) {
    let material_brushedMetal = mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    let color_gray = vec3<f32>(0.5, 0.5, 0.5);
    let color_white = vec3<f32>(1.0, 1.0, 1.0);
    let color_black = vec3<f32>(0.0, 0.0, 0.0);
    let gsl_simpleWeight_4 = gsl_simple(vec3<f32>(-92.01, -7.01, -40.01), vec3<f32>(-67.99, 17.01, 8.01));
    var gsl_simpleField_4 = clamp(gsl_remap_23(p - vec3<f32>(-80.0, 5.0, -8.0), material_brushedMetal, color_gray, color_white, color_black, noiseTex) * 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0));
    if (gsl_simpleWeight_4 > 0.0) {
      gsl_simpleField_4 = mix(gsl_simpleField_4, clamp(gsl_remap_22(p - vec3<f32>(-80.0, 5.0, -8.0), material_brushedMetal, color_gray, color_white, color_black, noiseTex) * 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0)), gsl_simpleWeight_4);
    }
    return vec4<f32>(gsl_simpleField_4, 1.0);
  }
  if (gsl_colorIdx == 11u) {
    let color_white = vec3<f32>(1.0, 1.0, 1.0);
    let color_black = vec3<f32>(0.0, 0.0, 0.0);
    return vec4<f32>(playful_checkerboard(p, 4.0, color_white, color_black, noiseTex), 1.0);
  }
  if (gsl_colorIdx == 12u) {
    let color_gray = vec3<f32>(0.5, 0.5, 0.5);
    var gsl_simpleValue_7 = vec3<f32>(0.37482947, 0.61767787, 0.7900267);
    if (gsl_simpleDetail > 0.0) {
      let material_brushedMetal = mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
      gsl_simpleValue_7 = mix(gsl_simpleValue_7, gsl_remap_12(vec3<f32>(dot(p, vec3<f32>(1.0, 0.0, 0.0)), dot(p, vec3<f32>(0.0, 1.0, 0.0)), dot(p, vec3<f32>(0.0, 0.0, 1.0))), material_brushedMetal, color_gray, noiseTex), gsl_simpleDetail);
    }
    let dots32 = gsl_simpleValue_7;
    var gsl_simpleValue_8 = dots32;
    if (gsl_simpleDetail > 0.0) {
      gsl_simpleValue_8 = mix(gsl_simpleValue_8, color_mixOklab(dots32, color_gray, 0.3 + 0.15 * noise_gradientLattice(p / 2.0, noiseTex), noiseTex), gsl_simpleDetail);
    }
    let reliefColor32 = gsl_simpleValue_8;
    return vec4<f32>(reliefColor32, 1.0);
  }
  if (gsl_colorIdx == 13u) {
    let material_brushedMetal = mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    let color_gray = vec3<f32>(0.5, 0.5, 0.5);
    let color_white = vec3<f32>(1.0, 1.0, 1.0);
    let color_black = vec3<f32>(0.0, 0.0, 0.0);
    let gsl_simpleWeight_5 = gsl_simple(vec3<f32>(-102.51, 38.49, -75.01), vec3<f32>(-57.49, 83.51, 15.01));
    var gsl_simpleField_5 = clamp(gsl_remap_25(p - vec3<f32>(-80.0, 61.0, -15.0), material_brushedMetal, color_gray, color_white, color_black, noiseTex) * 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0));
    if (gsl_simpleWeight_5 > 0.0) {
      gsl_simpleField_5 = mix(gsl_simpleField_5, clamp(gsl_remap_24(p - vec3<f32>(-80.0, 61.0, -15.0), material_brushedMetal, color_gray, color_white, color_black, noiseTex) * 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0)), gsl_simpleWeight_5);
    }
    return vec4<f32>(gsl_simpleField_5, 1.0);
  }
  if (gsl_colorIdx == 14u) {
    let color_white = vec3<f32>(1.0, 1.0, 1.0);
    let color_black = vec3<f32>(0.0, 0.0, 0.0);
    return vec4<f32>(playful_checkerboard(p, 8.0, color_white, color_black, noiseTex), 1.0);
  }
  if (gsl_colorIdx == 15u) {
    let color_gray = vec3<f32>(0.5, 0.5, 0.5);
    var gsl_simpleValue_9 = vec3<f32>(0.37482947, 0.61767787, 0.7900267);
    if (gsl_simpleDetail > 0.0) {
      let material_brushedMetal = mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
      gsl_simpleValue_9 = mix(gsl_simpleValue_9, gsl_remap_15(vec3<f32>(dot(p, vec3<f32>(1.0, 0.0, 0.0)), dot(p, vec3<f32>(0.0, 1.0, 0.0)), dot(p, vec3<f32>(0.0, 0.0, 1.0))), material_brushedMetal, color_gray, noiseTex), gsl_simpleDetail);
    }
    let dots64 = gsl_simpleValue_9;
    var gsl_simpleValue_10 = dots64;
    if (gsl_simpleDetail > 0.0) {
      gsl_simpleValue_10 = mix(gsl_simpleValue_10, color_mixOklab(dots64, color_gray, 0.3 + 0.15 * noise_gradientLattice(p / 4.0, noiseTex), noiseTex), gsl_simpleDetail);
    }
    let reliefColor64 = gsl_simpleValue_10;
    return vec4<f32>(reliefColor64, 1.0);
  }
  return vec4<f32>(0.0, 0.0, 0.0, 0.0);
}
fn gsl_materialField(p: vec3<f32>, time: f32, gsl_materialIdx: u32, noiseTex: texture_3d<f32>) -> gsl_material {
  if (gsl_materialIdx == 1u) {
    let material_brushedMetal = mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    let color_gray = vec3<f32>(0.5, 0.5, 0.5);
    let color_white = vec3<f32>(1.0, 1.0, 1.0);
    let color_black = vec3<f32>(0.0, 0.0, 0.0);
    let gsl_simpleWeight_6 = gsl_simple(vec3<f32>(-81.91, -62.91, -5.01), vec3<f32>(-78.09, -59.09, 1.01));
    var gsl_simpleField_6 = mat4x4<f32>(clamp(gsl_remap_29(p - vec3<f32>(-80.0, -61.0, -1.0), material_brushedMetal, color_gray, color_white, color_black, noiseTex) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_30(p - vec3<f32>(-80.0, -61.0, -1.0), material_brushedMetal, color_gray, color_white, color_black, noiseTex) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_30(p - vec3<f32>(-80.0, -61.0, -1.0), material_brushedMetal, color_gray, color_white, color_black, noiseTex) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    if (gsl_simpleWeight_6 > 0.0) {
      let gsl_simpleFull = mat4x4<f32>(clamp(gsl_remap_26(p - vec3<f32>(-80.0, -61.0, -1.0), material_brushedMetal, color_gray, color_white, color_black, noiseTex) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_27(p - vec3<f32>(-80.0, -61.0, -1.0), material_brushedMetal, color_gray, color_white, color_black, noiseTex) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_28(p - vec3<f32>(-80.0, -61.0, -1.0), material_brushedMetal, color_gray, color_white, color_black, noiseTex) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
      gsl_simpleField_6 = mat4x4<f32>(mix(gsl_simpleField_6[0], gsl_simpleFull[0], gsl_simpleWeight_6), mix(gsl_simpleField_6[1], gsl_simpleFull[1], gsl_simpleWeight_6), mix(gsl_simpleField_6[2], gsl_simpleFull[2], gsl_simpleWeight_6), mix(gsl_simpleField_6[3], gsl_simpleFull[3], gsl_simpleWeight_6));
    }
    let gsl_m = gsl_simpleField_6; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 2u) {
    let material_matte = mat4x4<f32>(vec4<f32>(1.0, 0.0, 1.45, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    let gsl_m = material_withDisplacement(material_matte, noise_gradientLattice(p / 0.5, noiseTex) * 0.5 / 4.0 * 0.03125, noiseTex); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 3u) {
    let material_plastic = mat4x4<f32>(vec4<f32>(0.35, 0.0, 1.46, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    let gsl_m = material_withDisplacement(material_plastic, noise_gradientLattice(p / 1.0, noiseTex) * 1.0 / 4.0 * 0.0625, noiseTex); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 4u) {
    let material_brushedMetal = mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    let color_gray = vec3<f32>(0.5, 0.5, 0.5);
    let color_white = vec3<f32>(1.0, 1.0, 1.0);
    let color_black = vec3<f32>(0.0, 0.0, 0.0);
    let material_matte = mat4x4<f32>(vec4<f32>(1.0, 0.0, 1.45, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    let material_plastic = mat4x4<f32>(vec4<f32>(0.35, 0.0, 1.46, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    let gsl_simpleWeight_7 = gsl_simple(vec3<f32>(-83.51, -50.51, -10.01), vec3<f32>(-76.49, -43.49, 2.01));
    var gsl_simpleField_7 = mat4x4<f32>(clamp(gsl_remap_34(p - vec3<f32>(-80.0, -47.0, -2.0), material_brushedMetal, color_gray, color_white, color_black, material_matte, material_plastic, noiseTex) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_35(p - vec3<f32>(-80.0, -47.0, -2.0), material_brushedMetal, color_gray, color_white, color_black, material_matte, material_plastic, noiseTex) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_35(p - vec3<f32>(-80.0, -47.0, -2.0), material_brushedMetal, color_gray, color_white, color_black, material_matte, material_plastic, noiseTex) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    if (gsl_simpleWeight_7 > 0.0) {
      let gsl_simpleFull_2 = mat4x4<f32>(clamp(gsl_remap_31(p - vec3<f32>(-80.0, -47.0, -2.0), material_brushedMetal, color_gray, color_white, color_black, material_matte, material_plastic, noiseTex) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_32(p - vec3<f32>(-80.0, -47.0, -2.0), material_brushedMetal, color_gray, color_white, color_black, material_matte, material_plastic, noiseTex) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_33(p - vec3<f32>(-80.0, -47.0, -2.0), material_brushedMetal, color_gray, color_white, color_black, material_matte, material_plastic, noiseTex) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
      gsl_simpleField_7 = mat4x4<f32>(mix(gsl_simpleField_7[0], gsl_simpleFull_2[0], gsl_simpleWeight_7), mix(gsl_simpleField_7[1], gsl_simpleFull_2[1], gsl_simpleWeight_7), mix(gsl_simpleField_7[2], gsl_simpleFull_2[2], gsl_simpleWeight_7), mix(gsl_simpleField_7[3], gsl_simpleFull_2[3], gsl_simpleWeight_7));
    }
    let gsl_m = gsl_simpleField_7; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 5u) {
    let material_matte = mat4x4<f32>(vec4<f32>(1.0, 0.0, 1.45, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    let gsl_m = material_withDisplacement(material_matte, noise_gradientLattice(p / 1.0, noiseTex) * 1.0 / 4.0 * 0.0625, noiseTex); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 6u) {
    let material_plastic = mat4x4<f32>(vec4<f32>(0.35, 0.0, 1.46, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    let gsl_m = material_withDisplacement(material_plastic, noise_gradientLattice(p / 2.0, noiseTex) * 2.0 / 4.0 * 0.125, noiseTex); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 7u) {
    let material_brushedMetal = mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    let color_gray = vec3<f32>(0.5, 0.5, 0.5);
    let color_white = vec3<f32>(1.0, 1.0, 1.0);
    let color_black = vec3<f32>(0.0, 0.0, 0.0);
    let material_matte = mat4x4<f32>(vec4<f32>(1.0, 0.0, 1.45, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    let material_plastic = mat4x4<f32>(vec4<f32>(0.35, 0.0, 1.46, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    let gsl_simpleWeight_8 = gsl_simple(vec3<f32>(-86.51, -33.51, -20.01), vec3<f32>(-73.49, -20.49, 4.01));
    var gsl_simpleField_8 = mat4x4<f32>(clamp(gsl_remap_34(p - vec3<f32>(-80.0, -27.0, -4.0), material_brushedMetal, color_gray, color_white, color_black, material_matte, material_plastic, noiseTex) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_35(p - vec3<f32>(-80.0, -27.0, -4.0), material_brushedMetal, color_gray, color_white, color_black, material_matte, material_plastic, noiseTex) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_35(p - vec3<f32>(-80.0, -27.0, -4.0), material_brushedMetal, color_gray, color_white, color_black, material_matte, material_plastic, noiseTex) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    if (gsl_simpleWeight_8 > 0.0) {
      let gsl_simpleFull_3 = mat4x4<f32>(clamp(gsl_remap_36(p - vec3<f32>(-80.0, -27.0, -4.0), material_brushedMetal, color_gray, color_white, color_black, material_matte, material_plastic, noiseTex) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_37(p - vec3<f32>(-80.0, -27.0, -4.0), material_brushedMetal, color_gray, color_white, color_black, material_matte, material_plastic, noiseTex) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_38(p - vec3<f32>(-80.0, -27.0, -4.0), material_brushedMetal, color_gray, color_white, color_black, material_matte, material_plastic, noiseTex) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
      gsl_simpleField_8 = mat4x4<f32>(mix(gsl_simpleField_8[0], gsl_simpleFull_3[0], gsl_simpleWeight_8), mix(gsl_simpleField_8[1], gsl_simpleFull_3[1], gsl_simpleWeight_8), mix(gsl_simpleField_8[2], gsl_simpleFull_3[2], gsl_simpleWeight_8), mix(gsl_simpleField_8[3], gsl_simpleFull_3[3], gsl_simpleWeight_8));
    }
    let gsl_m = gsl_simpleField_8; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 8u) {
    let material_matte = mat4x4<f32>(vec4<f32>(1.0, 0.0, 1.45, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    let gsl_m = material_withDisplacement(material_matte, noise_gradientLattice(p / 2.0, noiseTex) * 2.0 / 4.0 * 0.125, noiseTex); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 9u) {
    let material_plastic = mat4x4<f32>(vec4<f32>(0.35, 0.0, 1.46, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    let gsl_m = material_withDisplacement(material_plastic, noise_gradientLattice(p / 4.0, noiseTex) * 4.0 / 4.0 * 0.25, noiseTex); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 10u) {
    let material_brushedMetal = mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    let color_gray = vec3<f32>(0.5, 0.5, 0.5);
    let color_white = vec3<f32>(1.0, 1.0, 1.0);
    let color_black = vec3<f32>(0.0, 0.0, 0.0);
    let material_matte = mat4x4<f32>(vec4<f32>(1.0, 0.0, 1.45, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    let material_plastic = mat4x4<f32>(vec4<f32>(0.35, 0.0, 1.46, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    let gsl_simpleWeight_9 = gsl_simple(vec3<f32>(-92.01, -7.01, -40.01), vec3<f32>(-67.99, 17.01, 8.01));
    var gsl_simpleField_9 = mat4x4<f32>(clamp(gsl_remap_34(p - vec3<f32>(-80.0, 5.0, -8.0), material_brushedMetal, color_gray, color_white, color_black, material_matte, material_plastic, noiseTex) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_35(p - vec3<f32>(-80.0, 5.0, -8.0), material_brushedMetal, color_gray, color_white, color_black, material_matte, material_plastic, noiseTex) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_35(p - vec3<f32>(-80.0, 5.0, -8.0), material_brushedMetal, color_gray, color_white, color_black, material_matte, material_plastic, noiseTex) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    if (gsl_simpleWeight_9 > 0.0) {
      let gsl_simpleFull_4 = mat4x4<f32>(clamp(gsl_remap_39(p - vec3<f32>(-80.0, 5.0, -8.0), material_brushedMetal, color_gray, color_white, color_black, material_matte, material_plastic, noiseTex) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_40(p - vec3<f32>(-80.0, 5.0, -8.0), material_brushedMetal, color_gray, color_white, color_black, material_matte, material_plastic, noiseTex) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_41(p - vec3<f32>(-80.0, 5.0, -8.0), material_brushedMetal, color_gray, color_white, color_black, material_matte, material_plastic, noiseTex) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
      gsl_simpleField_9 = mat4x4<f32>(mix(gsl_simpleField_9[0], gsl_simpleFull_4[0], gsl_simpleWeight_9), mix(gsl_simpleField_9[1], gsl_simpleFull_4[1], gsl_simpleWeight_9), mix(gsl_simpleField_9[2], gsl_simpleFull_4[2], gsl_simpleWeight_9), mix(gsl_simpleField_9[3], gsl_simpleFull_4[3], gsl_simpleWeight_9));
    }
    let gsl_m = gsl_simpleField_9; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 11u) {
    let material_matte = mat4x4<f32>(vec4<f32>(1.0, 0.0, 1.45, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    let gsl_m = material_withDisplacement(material_matte, noise_gradientLattice(p / 4.0, noiseTex) * 4.0 / 4.0 * 0.25, noiseTex); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 12u) {
    let material_plastic = mat4x4<f32>(vec4<f32>(0.35, 0.0, 1.46, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    let gsl_m = material_withDisplacement(material_plastic, noise_gradientLattice(p / 8.0, noiseTex) * 8.0 / 4.0 * 0.5, noiseTex); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 13u) {
    let material_brushedMetal = mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    let color_gray = vec3<f32>(0.5, 0.5, 0.5);
    let color_white = vec3<f32>(1.0, 1.0, 1.0);
    let color_black = vec3<f32>(0.0, 0.0, 0.0);
    let material_matte = mat4x4<f32>(vec4<f32>(1.0, 0.0, 1.45, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    let material_plastic = mat4x4<f32>(vec4<f32>(0.35, 0.0, 1.46, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    let gsl_simpleWeight_10 = gsl_simple(vec3<f32>(-102.51, 38.49, -75.01), vec3<f32>(-57.49, 83.51, 15.01));
    var gsl_simpleField_10 = mat4x4<f32>(clamp(gsl_remap_34(p - vec3<f32>(-80.0, 61.0, -15.0), material_brushedMetal, color_gray, color_white, color_black, material_matte, material_plastic, noiseTex) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_35(p - vec3<f32>(-80.0, 61.0, -15.0), material_brushedMetal, color_gray, color_white, color_black, material_matte, material_plastic, noiseTex) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_35(p - vec3<f32>(-80.0, 61.0, -15.0), material_brushedMetal, color_gray, color_white, color_black, material_matte, material_plastic, noiseTex) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    if (gsl_simpleWeight_10 > 0.0) {
      let gsl_simpleFull_5 = mat4x4<f32>(clamp(gsl_remap_42(p - vec3<f32>(-80.0, 61.0, -15.0), material_brushedMetal, color_gray, color_white, color_black, material_matte, material_plastic, noiseTex) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_43(p - vec3<f32>(-80.0, 61.0, -15.0), material_brushedMetal, color_gray, color_white, color_black, material_matte, material_plastic, noiseTex) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_44(p - vec3<f32>(-80.0, 61.0, -15.0), material_brushedMetal, color_gray, color_white, color_black, material_matte, material_plastic, noiseTex) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
      gsl_simpleField_10 = mat4x4<f32>(mix(gsl_simpleField_10[0], gsl_simpleFull_5[0], gsl_simpleWeight_10), mix(gsl_simpleField_10[1], gsl_simpleFull_5[1], gsl_simpleWeight_10), mix(gsl_simpleField_10[2], gsl_simpleFull_5[2], gsl_simpleWeight_10), mix(gsl_simpleField_10[3], gsl_simpleFull_5[3], gsl_simpleWeight_10));
    }
    let gsl_m = gsl_simpleField_10; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 14u) {
    let material_matte = mat4x4<f32>(vec4<f32>(1.0, 0.0, 1.45, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    let gsl_m = material_withDisplacement(material_matte, noise_gradientLattice(p / 8.0, noiseTex) * 8.0 / 4.0 * 0.5, noiseTex); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 15u) {
    let material_plastic = mat4x4<f32>(vec4<f32>(0.35, 0.0, 1.46, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    let gsl_m = material_withDisplacement(material_plastic, noise_gradientLattice(p / 16.0, noiseTex) * 16.0 / 4.0 * 1.0, noiseTex); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
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

const gsl_foldK = array<vec3<f32>, 8>(vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 0.0, 0.0), vec3<f32>(0.0, 1.0, 0.0), vec3<f32>(1.0, 1.0, 0.0), vec3<f32>(0.0, 0.0, 1.0), vec3<f32>(1.0, 0.0, 1.0), vec3<f32>(0.0, 1.0, 1.0), vec3<f32>(1.0, 1.0, 1.0));

fn gsl_heights(p: vec3<f32>, time: f32, noiseTex: texture_3d<f32>) -> array<f32, 15> {
  let material_brushedMetal = mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
  let color_gray = vec3<f32>(0.5, 0.5, 0.5);
  let color_white = vec3<f32>(1.0, 1.0, 1.0);
  let color_black = vec3<f32>(0.0, 0.0, 0.0);
  let gsl_simpleWeight_11 = gsl_simple(vec3<f32>(-81.91, -62.91, -5.01), vec3<f32>(-78.09, -59.09, 1.01));
  var gsl_simpleField_11 = mat4x4<f32>(clamp(gsl_remap_29(p - vec3<f32>(-80.0, -61.0, -1.0), material_brushedMetal, color_gray, color_white, color_black, noiseTex) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_30(p - vec3<f32>(-80.0, -61.0, -1.0), material_brushedMetal, color_gray, color_white, color_black, noiseTex) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_30(p - vec3<f32>(-80.0, -61.0, -1.0), material_brushedMetal, color_gray, color_white, color_black, noiseTex) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
  if (gsl_simpleWeight_11 > 0.0) {
    let gsl_simpleFull_6 = mat4x4<f32>(clamp(gsl_remap_26(p - vec3<f32>(-80.0, -61.0, -1.0), material_brushedMetal, color_gray, color_white, color_black, noiseTex) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_27(p - vec3<f32>(-80.0, -61.0, -1.0), material_brushedMetal, color_gray, color_white, color_black, noiseTex) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_28(p - vec3<f32>(-80.0, -61.0, -1.0), material_brushedMetal, color_gray, color_white, color_black, noiseTex) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    gsl_simpleField_11 = mat4x4<f32>(mix(gsl_simpleField_11[0], gsl_simpleFull_6[0], gsl_simpleWeight_11), mix(gsl_simpleField_11[1], gsl_simpleFull_6[1], gsl_simpleWeight_11), mix(gsl_simpleField_11[2], gsl_simpleFull_6[2], gsl_simpleWeight_11), mix(gsl_simpleField_11[3], gsl_simpleFull_6[3], gsl_simpleWeight_11));
  }
  let material_matte = mat4x4<f32>(vec4<f32>(1.0, 0.0, 1.45, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
  let material_plastic = mat4x4<f32>(vec4<f32>(0.35, 0.0, 1.46, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
  let gsl_simpleWeight_12 = gsl_simple(vec3<f32>(-83.51, -50.51, -10.01), vec3<f32>(-76.49, -43.49, 2.01));
  var gsl_simpleField_12 = mat4x4<f32>(clamp(gsl_remap_34(p - vec3<f32>(-80.0, -47.0, -2.0), material_brushedMetal, color_gray, color_white, color_black, material_matte, material_plastic, noiseTex) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_35(p - vec3<f32>(-80.0, -47.0, -2.0), material_brushedMetal, color_gray, color_white, color_black, material_matte, material_plastic, noiseTex) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_35(p - vec3<f32>(-80.0, -47.0, -2.0), material_brushedMetal, color_gray, color_white, color_black, material_matte, material_plastic, noiseTex) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
  if (gsl_simpleWeight_12 > 0.0) {
    let gsl_simpleFull_7 = mat4x4<f32>(clamp(gsl_remap_31(p - vec3<f32>(-80.0, -47.0, -2.0), material_brushedMetal, color_gray, color_white, color_black, material_matte, material_plastic, noiseTex) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_32(p - vec3<f32>(-80.0, -47.0, -2.0), material_brushedMetal, color_gray, color_white, color_black, material_matte, material_plastic, noiseTex) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_33(p - vec3<f32>(-80.0, -47.0, -2.0), material_brushedMetal, color_gray, color_white, color_black, material_matte, material_plastic, noiseTex) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    gsl_simpleField_12 = mat4x4<f32>(mix(gsl_simpleField_12[0], gsl_simpleFull_7[0], gsl_simpleWeight_12), mix(gsl_simpleField_12[1], gsl_simpleFull_7[1], gsl_simpleWeight_12), mix(gsl_simpleField_12[2], gsl_simpleFull_7[2], gsl_simpleWeight_12), mix(gsl_simpleField_12[3], gsl_simpleFull_7[3], gsl_simpleWeight_12));
  }
  let gsl_simpleWeight_13 = gsl_simple(vec3<f32>(-86.51, -33.51, -20.01), vec3<f32>(-73.49, -20.49, 4.01));
  var gsl_simpleField_13 = mat4x4<f32>(clamp(gsl_remap_34(p - vec3<f32>(-80.0, -27.0, -4.0), material_brushedMetal, color_gray, color_white, color_black, material_matte, material_plastic, noiseTex) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_35(p - vec3<f32>(-80.0, -27.0, -4.0), material_brushedMetal, color_gray, color_white, color_black, material_matte, material_plastic, noiseTex) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_35(p - vec3<f32>(-80.0, -27.0, -4.0), material_brushedMetal, color_gray, color_white, color_black, material_matte, material_plastic, noiseTex) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
  if (gsl_simpleWeight_13 > 0.0) {
    let gsl_simpleFull_8 = mat4x4<f32>(clamp(gsl_remap_36(p - vec3<f32>(-80.0, -27.0, -4.0), material_brushedMetal, color_gray, color_white, color_black, material_matte, material_plastic, noiseTex) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_37(p - vec3<f32>(-80.0, -27.0, -4.0), material_brushedMetal, color_gray, color_white, color_black, material_matte, material_plastic, noiseTex) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_38(p - vec3<f32>(-80.0, -27.0, -4.0), material_brushedMetal, color_gray, color_white, color_black, material_matte, material_plastic, noiseTex) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    gsl_simpleField_13 = mat4x4<f32>(mix(gsl_simpleField_13[0], gsl_simpleFull_8[0], gsl_simpleWeight_13), mix(gsl_simpleField_13[1], gsl_simpleFull_8[1], gsl_simpleWeight_13), mix(gsl_simpleField_13[2], gsl_simpleFull_8[2], gsl_simpleWeight_13), mix(gsl_simpleField_13[3], gsl_simpleFull_8[3], gsl_simpleWeight_13));
  }
  let gsl_simpleWeight_14 = gsl_simple(vec3<f32>(-92.01, -7.01, -40.01), vec3<f32>(-67.99, 17.01, 8.01));
  var gsl_simpleField_14 = mat4x4<f32>(clamp(gsl_remap_34(p - vec3<f32>(-80.0, 5.0, -8.0), material_brushedMetal, color_gray, color_white, color_black, material_matte, material_plastic, noiseTex) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_35(p - vec3<f32>(-80.0, 5.0, -8.0), material_brushedMetal, color_gray, color_white, color_black, material_matte, material_plastic, noiseTex) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_35(p - vec3<f32>(-80.0, 5.0, -8.0), material_brushedMetal, color_gray, color_white, color_black, material_matte, material_plastic, noiseTex) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
  if (gsl_simpleWeight_14 > 0.0) {
    let gsl_simpleFull_9 = mat4x4<f32>(clamp(gsl_remap_39(p - vec3<f32>(-80.0, 5.0, -8.0), material_brushedMetal, color_gray, color_white, color_black, material_matte, material_plastic, noiseTex) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_40(p - vec3<f32>(-80.0, 5.0, -8.0), material_brushedMetal, color_gray, color_white, color_black, material_matte, material_plastic, noiseTex) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_41(p - vec3<f32>(-80.0, 5.0, -8.0), material_brushedMetal, color_gray, color_white, color_black, material_matte, material_plastic, noiseTex) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    gsl_simpleField_14 = mat4x4<f32>(mix(gsl_simpleField_14[0], gsl_simpleFull_9[0], gsl_simpleWeight_14), mix(gsl_simpleField_14[1], gsl_simpleFull_9[1], gsl_simpleWeight_14), mix(gsl_simpleField_14[2], gsl_simpleFull_9[2], gsl_simpleWeight_14), mix(gsl_simpleField_14[3], gsl_simpleFull_9[3], gsl_simpleWeight_14));
  }
  let gsl_simpleWeight_15 = gsl_simple(vec3<f32>(-102.51, 38.49, -75.01), vec3<f32>(-57.49, 83.51, 15.01));
  var gsl_simpleField_15 = mat4x4<f32>(clamp(gsl_remap_34(p - vec3<f32>(-80.0, 61.0, -15.0), material_brushedMetal, color_gray, color_white, color_black, material_matte, material_plastic, noiseTex) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_35(p - vec3<f32>(-80.0, 61.0, -15.0), material_brushedMetal, color_gray, color_white, color_black, material_matte, material_plastic, noiseTex) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_35(p - vec3<f32>(-80.0, 61.0, -15.0), material_brushedMetal, color_gray, color_white, color_black, material_matte, material_plastic, noiseTex) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
  if (gsl_simpleWeight_15 > 0.0) {
    let gsl_simpleFull_10 = mat4x4<f32>(clamp(gsl_remap_42(p - vec3<f32>(-80.0, 61.0, -15.0), material_brushedMetal, color_gray, color_white, color_black, material_matte, material_plastic, noiseTex) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_43(p - vec3<f32>(-80.0, 61.0, -15.0), material_brushedMetal, color_gray, color_white, color_black, material_matte, material_plastic, noiseTex) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_44(p - vec3<f32>(-80.0, 61.0, -15.0), material_brushedMetal, color_gray, color_white, color_black, material_matte, material_plastic, noiseTex) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    gsl_simpleField_15 = mat4x4<f32>(mix(gsl_simpleField_15[0], gsl_simpleFull_10[0], gsl_simpleWeight_15), mix(gsl_simpleField_15[1], gsl_simpleFull_10[1], gsl_simpleWeight_15), mix(gsl_simpleField_15[2], gsl_simpleFull_10[2], gsl_simpleWeight_15), mix(gsl_simpleField_15[3], gsl_simpleFull_10[3], gsl_simpleWeight_15));
  }
  return array<f32, 15>(gsl_simpleField_11[2].z, material_withDisplacement(material_matte, noise_gradientLattice(p / 0.5, noiseTex) * 0.5 / 4.0 * 0.03125, noiseTex)[2].z, material_withDisplacement(material_plastic, noise_gradientLattice(p / 1.0, noiseTex) * 1.0 / 4.0 * 0.0625, noiseTex)[2].z, gsl_simpleField_12[2].z, material_withDisplacement(material_matte, noise_gradientLattice(p / 1.0, noiseTex) * 1.0 / 4.0 * 0.0625, noiseTex)[2].z, material_withDisplacement(material_plastic, noise_gradientLattice(p / 2.0, noiseTex) * 2.0 / 4.0 * 0.125, noiseTex)[2].z, gsl_simpleField_13[2].z, material_withDisplacement(material_matte, noise_gradientLattice(p / 2.0, noiseTex) * 2.0 / 4.0 * 0.125, noiseTex)[2].z, material_withDisplacement(material_plastic, noise_gradientLattice(p / 4.0, noiseTex) * 4.0 / 4.0 * 0.25, noiseTex)[2].z, gsl_simpleField_14[2].z, material_withDisplacement(material_matte, noise_gradientLattice(p / 4.0, noiseTex) * 4.0 / 4.0 * 0.25, noiseTex)[2].z, material_withDisplacement(material_plastic, noise_gradientLattice(p / 8.0, noiseTex) * 8.0 / 4.0 * 0.5, noiseTex)[2].z, gsl_simpleField_15[2].z, material_withDisplacement(material_matte, noise_gradientLattice(p / 8.0, noiseTex) * 8.0 / 4.0 * 0.5, noiseTex)[2].z, material_withDisplacement(material_plastic, noise_gradientLattice(p / 16.0, noiseTex) * 16.0 / 4.0 * 1.0, noiseTex)[2].z);
}

fn gsl_sdShapeDisplaced(p: vec3<f32>, time: f32, gsl_dispK: f32, noiseTex: texture_3d<f32>) -> gsl_shape {
  let gsl_d = gsl_shapeParts(p, time, noiseTex);
  let gsl_h = gsl_heights(p, time, noiseTex);
  let gsl_s1 = gsl_shape(gsl_d[0] - gsl_h[0] * gsl_dispK, 1u);
  let gsl_s2 = gsl_shape(gsl_d[1] - gsl_h[1] * gsl_dispK, 2u);
  let gsl_s3 = gsl_shape(gsl_d[2] - gsl_h[2] * gsl_dispK, 3u);
  let gsl_s4 = gsl_shape(gsl_d[3] - gsl_h[3] * gsl_dispK, 4u);
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
  return gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_s1, gsl_s2), gsl_s3), gsl_s4), gsl_s5), gsl_s6), gsl_s7), gsl_s8), gsl_s9), gsl_s10), gsl_s11), gsl_s12), gsl_s13), gsl_s14), gsl_s15);
}
