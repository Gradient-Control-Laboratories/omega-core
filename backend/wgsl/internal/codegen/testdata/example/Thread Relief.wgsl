fn drill_prismVoid(p: vec3<f32>, zTop: f32, section: f32, depth: f32) -> f32 {
  return max(section, zTop - depth - p.z);
}

fn gsl_remap(p: vec3<f32>) -> f32 {
  let screw = fastener_screw(max(max(max(length(p.xy) - 5.0, -p.z), p.z - 6.0), (length(p.xy) - 5.0 + p.z - 6.0 + 0.2) * 0.7071), max(max(max(length(p.xy) - 3.0, -20.0 - p.z), p.z - 3.0), (length(p.xy) - 3.0 + -p.z - 20.0 + 0.9) * 0.7071), drill_prismVoid(p, 6.0, max(max(abs(p.x) - 2.5, abs(0.5 * p.x + 0.8660254 * p.y) - 2.5), abs(0.5 * p.x - 0.8660254 * p.y) - 2.5), 3.0));
  return screw;
}

fn gsl_remap_2(p: vec3<f32>) -> f32 {
  let plate = max(builtin_gcl_closure_boxCentered(p, vec3<f32>(0.0, 0.0, 8.0) * 0.5, abs(vec3<f32>(12.0, 12.0, 8.0))), 2.458734 - length(p.xy));
  return plate;
}

fn gsl_remap_3(p: vec3<f32>) -> f32 {
  let stub = fastener_screw(max(max(max(length(p.xy) - 5.0, -p.z), p.z - 6.0), (length(p.xy) - 5.0 + p.z - 6.0 + 0.2) * 0.7071), max(max(max(length(p.xy) - 3.0, -16.0 - p.z), p.z - 3.0), (length(p.xy) - 3.0 + -p.z - 16.0 + 0.9) * 0.7071), drill_prismVoid(p, 6.0, max(max(abs(p.x) - 2.5, abs(0.5 * p.x + 0.8660254 * p.y) - 2.5), abs(0.5 * p.x - 0.8660254 * p.y) - 2.5), 3.0));
  return stub;
}

fn thread_formDepth(pitchVal: f32) -> f32 {
  return 0.5412659 * pitchVal;
}

fn thread_helixPhase(p: vec3<f32>, pitchVal: f32) -> f32 {
  let ang = 6.2831855 * p.z / max(pitchVal, 0.01);
  let radial = max(length(p.xy), 0.2);
  let along = cos(ang) * p.x / radial + sin(ang) * p.y / radial;
  return acos(clamp(along, -1.0, 1.0)) / 3.1415927;
}

fn thread_profileDepth(p: vec3<f32>, pitchVal: f32) -> f32 {
  let offset = thread_helixPhase(p, pitchVal) * pitchVal * 0.5;
  return clamp((offset - pitchVal / 16.0) * 1.7320508, 0.0, thread_formDepth(pitchVal));
}

fn thread_runout(p: vec3<f32>, zTop: f32, zBottom: f32, outerR: f32) -> f32 {
  return max(0.0, 0.5 * min(min(zTop - p.z, p.z - zBottom), outerR - length(p.xy)));
}

fn thread_maleRelief(p: vec3<f32>, major: f32, pitchVal: f32, zTop: f32, zBottom: f32) -> f32 {
  let outerR = major / 2.0 + thread_formDepth(pitchVal);
  let innerR = major / 2.0 - 2.0 * thread_formDepth(pitchVal);
  return -min(min(thread_profileDepth(p, pitchVal), thread_runout(p, zTop, zBottom, outerR)), max(0.0, length(p.xy) - innerR));
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

fn gsl_remap_4(p: vec3<f32>) -> vec3<f32> {
  var gsl_simpleValue = vec3<f32>(0.77, 0.78, 0.78);
  if (gsl_simpleDetail > 0.0) {
    gsl_simpleValue = mix(gsl_simpleValue, color_mixOklab(vec3<f32>(0.77, 0.78, 0.78), vec3<f32>(0.2695, 0.273, 0.273), clamp(-thread_maleRelief(p, 6.0, 1.0, 0.0, -20.0) / 0.5412659, 0.0, 1.0) * (1.0 - gsl_dispDrawn)), gsl_simpleDetail);
  }
  return gsl_simpleValue;
}

fn gsl_remap_5(p: vec3<f32>) -> vec3<f32> {
  var gsl_simpleValue_2 = vec3<f32>(0.96, 0.96, 0.97);
  if (gsl_simpleDetail > 0.0) {
    let gsl_tmp_18 = 6.2831855 * p.z / 1.0;
    let gsl_tmp_17 = max(length(p.xy), 0.2);
    gsl_simpleValue_2 = mix(gsl_simpleValue_2, color_mixOklab(vec3<f32>(0.96, 0.96, 0.97), vec3<f32>(0.336, 0.336, 0.3395), clamp(-(-min(max(0.5412659 - clamp((acos(clamp(cos(gsl_tmp_18) * p.x / gsl_tmp_17 + sin(gsl_tmp_18) * p.y / gsl_tmp_17, -1.0, 1.0)) / 3.1415927 * 0.5 - 0.0625) * 1.7320508, 0.0, 0.5412659), 0.0), max(0.0, 0.5 * min(min(8.0 - p.z, p.z - 0.0), 4.082532 - length(p.xy))))) / 0.5412659, 0.0, 1.0) * (1.0 - gsl_dispDrawn)), gsl_simpleDetail);
  }
  return gsl_simpleValue_2;
}

fn gsl_remap_6(p: vec3<f32>) -> vec3<f32> {
  var gsl_simpleValue_3 = vec3<f32>(0.96, 0.96, 0.97);
  if (gsl_simpleDetail > 0.0) {
    let gsl_tmp_20 = 6.2831855 * p.z / 1.0;
    let gsl_tmp_19 = max(length(p.xy), 0.2);
    gsl_simpleValue_3 = mix(gsl_simpleValue_3, color_mixOklab(vec3<f32>(0.96, 0.96, 0.97), vec3<f32>(0.336, 0.336, 0.3395), clamp(-(-min(max(0.5412659 - clamp((acos(clamp(cos(gsl_tmp_20) * p.x / gsl_tmp_19 + sin(gsl_tmp_20) * p.y / gsl_tmp_19, -1.0, 1.0)) / 3.1415927 * 0.5 - 0.0625) * 1.7320508, 0.0, 0.5412659), 0.0), max(0.0, 0.5 * min(min(8.0 - p.z, p.z - 0.0), 4.082532 - length(p.xy))))) / 0.5412659, 0.0, 1.0) * (1.0 - gsl_dispDrawn)), gsl_simpleDetail);
  }
  return gsl_simpleValue_3;
}

fn gsl_remap_7(p: vec3<f32>) -> vec3<f32> {
  var gsl_simpleValue_4 = vec3<f32>(0.77, 0.78, 0.78);
  if (gsl_simpleDetail > 0.0) {
    gsl_simpleValue_4 = mix(gsl_simpleValue_4, color_mixOklab(vec3<f32>(0.77, 0.78, 0.78), vec3<f32>(0.2695, 0.273, 0.273), clamp(-thread_maleRelief(p, 6.0, 1.0, 0.0, -16.0) / 0.5412659, 0.0, 1.0) * (1.0 - gsl_dispDrawn)), gsl_simpleDetail);
  }
  return gsl_simpleValue_4;
}

fn gsl_remap_8(p: vec3<f32>) -> vec4<f32> {
  return material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), thread_maleRelief(p, 6.0, 1.0, 0.0, -20.0))[0];
}

fn gsl_remap_9(p: vec3<f32>) -> vec4<f32> {
  return material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), thread_maleRelief(p, 6.0, 1.0, 0.0, -20.0))[1];
}

fn gsl_remap_10(p: vec3<f32>) -> vec4<f32> {
  return material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), thread_maleRelief(p, 6.0, 1.0, 0.0, -20.0))[2];
}

fn gsl_remap_11(p: vec3<f32>) -> vec4<f32> {
  let gsl_tmp_22 = 6.2831855 * p.z / 1.0;
  let gsl_tmp_21 = max(length(p.xy), 0.2);
  return material_withDisplacement(mat4x4<f32>(vec4<f32>(0.4, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), -min(max(0.5412659 - clamp((acos(clamp(cos(gsl_tmp_22) * p.x / gsl_tmp_21 + sin(gsl_tmp_22) * p.y / gsl_tmp_21, -1.0, 1.0)) / 3.1415927 * 0.5 - 0.0625) * 1.7320508, 0.0, 0.5412659), 0.0), max(0.0, 0.5 * min(min(8.0 - p.z, p.z - 0.0), 4.082532 - length(p.xy)))))[0];
}

fn gsl_remap_12(p: vec3<f32>) -> vec4<f32> {
  let gsl_tmp_24 = 6.2831855 * p.z / 1.0;
  let gsl_tmp_23 = max(length(p.xy), 0.2);
  return material_withDisplacement(mat4x4<f32>(vec4<f32>(0.4, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), -min(max(0.5412659 - clamp((acos(clamp(cos(gsl_tmp_24) * p.x / gsl_tmp_23 + sin(gsl_tmp_24) * p.y / gsl_tmp_23, -1.0, 1.0)) / 3.1415927 * 0.5 - 0.0625) * 1.7320508, 0.0, 0.5412659), 0.0), max(0.0, 0.5 * min(min(8.0 - p.z, p.z - 0.0), 4.082532 - length(p.xy)))))[1];
}

fn gsl_remap_13(p: vec3<f32>) -> vec4<f32> {
  let gsl_tmp_26 = 6.2831855 * p.z / 1.0;
  let gsl_tmp_25 = max(length(p.xy), 0.2);
  return material_withDisplacement(mat4x4<f32>(vec4<f32>(0.4, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), -min(max(0.5412659 - clamp((acos(clamp(cos(gsl_tmp_26) * p.x / gsl_tmp_25 + sin(gsl_tmp_26) * p.y / gsl_tmp_25, -1.0, 1.0)) / 3.1415927 * 0.5 - 0.0625) * 1.7320508, 0.0, 0.5412659), 0.0), max(0.0, 0.5 * min(min(8.0 - p.z, p.z - 0.0), 4.082532 - length(p.xy)))))[2];
}

fn gsl_remap_14(p: vec3<f32>) -> vec4<f32> {
  let gsl_tmp_28 = 6.2831855 * p.z / 1.0;
  let gsl_tmp_27 = max(length(p.xy), 0.2);
  return material_withDisplacement(mat4x4<f32>(vec4<f32>(0.4, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), -min(max(0.5412659 - clamp((acos(clamp(cos(gsl_tmp_28) * p.x / gsl_tmp_27 + sin(gsl_tmp_28) * p.y / gsl_tmp_27, -1.0, 1.0)) / 3.1415927 * 0.5 - 0.0625) * 1.7320508, 0.0, 0.5412659), 0.0), max(0.0, 0.5 * min(min(8.0 - p.z, p.z - 0.0), 4.082532 - length(p.xy)))))[0];
}

fn gsl_remap_15(p: vec3<f32>) -> vec4<f32> {
  let gsl_tmp_30 = 6.2831855 * p.z / 1.0;
  let gsl_tmp_29 = max(length(p.xy), 0.2);
  return material_withDisplacement(mat4x4<f32>(vec4<f32>(0.4, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), -min(max(0.5412659 - clamp((acos(clamp(cos(gsl_tmp_30) * p.x / gsl_tmp_29 + sin(gsl_tmp_30) * p.y / gsl_tmp_29, -1.0, 1.0)) / 3.1415927 * 0.5 - 0.0625) * 1.7320508, 0.0, 0.5412659), 0.0), max(0.0, 0.5 * min(min(8.0 - p.z, p.z - 0.0), 4.082532 - length(p.xy)))))[1];
}

fn gsl_remap_16(p: vec3<f32>) -> vec4<f32> {
  let gsl_tmp_32 = 6.2831855 * p.z / 1.0;
  let gsl_tmp_31 = max(length(p.xy), 0.2);
  return material_withDisplacement(mat4x4<f32>(vec4<f32>(0.4, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), -min(max(0.5412659 - clamp((acos(clamp(cos(gsl_tmp_32) * p.x / gsl_tmp_31 + sin(gsl_tmp_32) * p.y / gsl_tmp_31, -1.0, 1.0)) / 3.1415927 * 0.5 - 0.0625) * 1.7320508, 0.0, 0.5412659), 0.0), max(0.0, 0.5 * min(min(8.0 - p.z, p.z - 0.0), 4.082532 - length(p.xy)))))[2];
}

fn gsl_remap_17(p: vec3<f32>) -> vec4<f32> {
  return material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), thread_maleRelief(p, 6.0, 1.0, 0.0, -16.0))[0];
}

fn gsl_remap_18(p: vec3<f32>) -> vec4<f32> {
  return material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), thread_maleRelief(p, 6.0, 1.0, 0.0, -16.0))[1];
}

fn gsl_remap_19(p: vec3<f32>) -> vec4<f32> {
  return material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), thread_maleRelief(p, 6.0, 1.0, 0.0, -16.0))[2];
}

fn builtin_intersectSharp(fields: vec3<f32>) -> f32 {
  return max(max(fields.x, fields.y), fields.z);
}

fn builtin_intersectRound(fields: vec3<f32>, radius: f32) -> f32 {
  let sumOfSquares = 0.0 + max(fields.x + radius, 0.0) * max(fields.x + radius, 0.0) + max(fields.y + radius, 0.0) * max(fields.y + radius, 0.0) + max(fields.z + radius, 0.0) * max(fields.z + radius, 0.0);
  return min(-radius, builtin_intersectSharp(fields)) + sqrt(sumOfSquares);
}

fn builtin_gcl_closure_intersectRound(fields: vec3<f32>) -> f32 {
  return builtin_intersectRound(fields, 0.0);
}

fn builtin_gcl_closure_boxCentered(p: vec3<f32>, center: vec3<f32>, size: vec3<f32>) -> f32 {
  let halfSize = size * 0.5;
  let d = abs(p - center) - halfSize;
  let solid = builtin_gcl_closure_intersectRound(vec3<f32>(d.x, d.y, d.z));
  return solid;
}

fn material_withDisplacement(m: mat4x4<f32>, height: f32) -> mat4x4<f32> {
  let volume = m[2];
  return mat4x4<f32>(m[0], m[1], vec4<f32>(volume.x, volume.y, height, volume.w), m[3]);
}

fn fastener_screw(head: f32, shank: f32, drive: f32) -> f32 {
  return max(min(head, shank), -drive);
}

fn gsl_shapeParts(p: vec3<f32>, time: f32) -> array<f32, 4> {
  let looseScrew = gsl_remap(vec3<f32>(p.x + 18.0, p.y, p.z - 20.0));
  let tappedPlate = gsl_remap_2(vec3<f32>(p.x, p.y, p.z));
  let jointPlate = gsl_remap_2(vec3<f32>(p.x - 18.0, p.y, p.z));
  let jointScrew = gsl_remap_3(vec3<f32>(p.x - 18.0, p.y, p.z - 14.0));
  return array<f32, 4>(looseScrew, tappedPlate, jointPlate, jointScrew);
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let gsl_d = gsl_shapeParts(p, time);
  let gsl_s1 = gsl_shape(gsl_d[0], 1u);
  let gsl_s2 = gsl_shape(gsl_d[1], 2u);
  let gsl_s3 = gsl_shape(gsl_d[2], 3u);
  let gsl_s4 = gsl_shape(gsl_d[3], 4u);
  return gsl_min(gsl_min(gsl_min(gsl_s1, gsl_s2), gsl_s3), gsl_s4);
}
fn gsl_colorField(p: vec3<f32>, time: f32, gsl_colorIdx: u32) -> vec4<f32> {
  if (gsl_colorIdx == 1u) { return vec4<f32>(clamp(gsl_remap_4(vec3<f32>(p.x + 18.0, p.y, p.z - 20.0)) * 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0)), 1.0); }
  if (gsl_colorIdx == 2u) { return vec4<f32>(clamp(gsl_remap_5(vec3<f32>(p.x, p.y, p.z)) * 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0)), 1.0); }
  if (gsl_colorIdx == 3u) { return vec4<f32>(clamp(gsl_remap_6(vec3<f32>(p.x - 18.0, p.y, p.z)) * 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0)), 1.0); }
  if (gsl_colorIdx == 4u) { return vec4<f32>(clamp(gsl_remap_7(vec3<f32>(p.x - 18.0, p.y, p.z - 14.0)) * 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0)), 1.0); }
  return vec4<f32>(0.0, 0.0, 0.0, 0.0);
}
fn gsl_materialField(p: vec3<f32>, time: f32, gsl_materialIdx: u32) -> gsl_material {
  if (gsl_materialIdx == 1u) { let gsl_m = mat4x4<f32>(clamp(gsl_remap_8(vec3<f32>(p.x + 18.0, p.y, p.z - 20.0)) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_9(vec3<f32>(p.x + 18.0, p.y, p.z - 20.0)) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_10(vec3<f32>(p.x + 18.0, p.y, p.z - 20.0)) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 2u) { let gsl_m = mat4x4<f32>(clamp(gsl_remap_11(vec3<f32>(p.x, p.y, p.z)) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_12(vec3<f32>(p.x, p.y, p.z)) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_13(vec3<f32>(p.x, p.y, p.z)) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 3u) { let gsl_m = mat4x4<f32>(clamp(gsl_remap_14(vec3<f32>(p.x - 18.0, p.y, p.z)) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_15(vec3<f32>(p.x - 18.0, p.y, p.z)) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_16(vec3<f32>(p.x - 18.0, p.y, p.z)) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 4u) { let gsl_m = mat4x4<f32>(clamp(gsl_remap_17(vec3<f32>(p.x - 18.0, p.y, p.z - 14.0)) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_18(vec3<f32>(p.x - 18.0, p.y, p.z - 14.0)) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_19(vec3<f32>(p.x - 18.0, p.y, p.z - 14.0)) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  return gsl_material(vec4<f32>(0.5, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), 0.0);
}
var<private> gsl_simpleDetail: f32 = 1.0;

var<private> gsl_dispDrawn: f32 = 0.0;

fn gsl_heights(p: vec3<f32>, time: f32) -> array<f32, 4> {
  return array<f32, 4>(clamp(gsl_remap_10(vec3<f32>(p.x + 18.0, p.y, p.z - 20.0)) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)).z, clamp(gsl_remap_13(vec3<f32>(p.x, p.y, p.z)) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)).z, clamp(gsl_remap_16(vec3<f32>(p.x - 18.0, p.y, p.z)) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)).z, clamp(gsl_remap_19(vec3<f32>(p.x - 18.0, p.y, p.z - 14.0)) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)).z);
}

fn gsl_sdShapeDisplaced(p: vec3<f32>, time: f32, gsl_dispK: f32) -> gsl_shape {
  let gsl_d = gsl_shapeParts(p, time);
  let gsl_h = gsl_heights(p, time);
  let gsl_s1 = gsl_shape(gsl_d[0] - gsl_h[0] * gsl_dispK, 1u);
  let gsl_s2 = gsl_shape(gsl_d[1] - gsl_h[1] * gsl_dispK, 2u);
  let gsl_s3 = gsl_shape(gsl_d[2] - gsl_h[2] * gsl_dispK, 3u);
  let gsl_s4 = gsl_shape(gsl_d[3] - gsl_h[3] * gsl_dispK, 4u);
  return gsl_min(gsl_min(gsl_min(gsl_s1, gsl_s2), gsl_s3), gsl_s4);
}
