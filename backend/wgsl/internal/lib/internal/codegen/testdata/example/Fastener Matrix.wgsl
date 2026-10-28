fn gsl_floored(a: f32, n: f32) -> f32 {
  return a - n * floor(a / n);
}

fn gsl_remap(p: vec3<f32>) -> f32 {
  return max(min(max(max(max(length(p.xy) - 5.0, -p.z), p.z - 6.0), (length(p.xy) - 5.0 + p.z - 6.0 + 0.2) * 0.7071), fastener_plainShank(p, 6.0, 18.0, 3.0)), -drill_prismVoid(p, 6.0, drill_hexXY(p, 5.0), 3.0));
}

fn gsl_remap_2(p: vec3<f32>) -> f32 {
  return min(max(max(length(p.xy) - 5.0, -p.z), p.z - 6.0), max(max(length(p.xy) - 3.0, -18.0 - p.z), p.z - 3.0));
}

fn gsl_remap_3(p: vec3<f32>) -> f32 {
  return max(min(max(length(vec3<f32>(p.x, p.y, p.z - -2.5261364)) - 5.8261366, -p.z), fastener_plainShank(p, 6.0, 18.0, 1.65)), -drill_prismVoid(p, 3.3, drill_hexXY(p, 4.0), 2.08));
}

fn gsl_remap_4(p: vec3<f32>) -> f32 {
  return min(max(length(vec3<f32>(p.x, p.y, p.z - -2.5261364)) - 5.8261366, -p.z), max(max(length(p.xy) - 3.0, -18.0 - p.z), p.z - 1.65));
}

fn gsl_remap_5(p: vec3<f32>) -> f32 {
  let gsl_tmp_67 = length(p.xy);
  let gsl_tmp_68 = abs(p.x);
  let gsl_tmp_69 = abs(p.y);
  let gsl_tmp_70 = abs(gsl_tmp_68 - gsl_tmp_69) * 0.70710677;
  let gsl_tmp_71 = ((gsl_tmp_68 + gsl_tmp_69) * 0.70710677 - 1.205) * 0.99496853 - (p.z - 0.5150396) * 0.10018806;
  return max(min(fastener_panHead(p, 12.0, 4.6, 10.0), fastener_plainShank(p, 6.0, 18.0, 2.3)), -max(max(gsl_tmp_67 * 0.89493436 + (3.9860268 - p.z) * 0.4461978 - 3.0875235, (-0.12567022 - p.z) * 0.88294756 + gsl_tmp_67 * 0.46947157), -max(gsl_tmp_70 * 0.2923717 - gsl_tmp_71 * 0.9563047, (gsl_tmp_70 - 1.0025) * 0.6946584 - (gsl_tmp_71 - 0.306495) * 0.7193398)));
}

fn gsl_remap_6(p: vec3<f32>) -> f32 {
  return min(fastener_panHead(p, 12.0, 4.6, 10.0), max(max(length(p.xy) - 3.0, -18.0 - p.z), p.z - 2.3));
}

fn gsl_remap_7(p: vec3<f32>) -> f32 {
  return max(min(max(max((length(p.xy) - (6.72 + p.z * 1.0)) * 0.70710677, -3.72 - p.z), p.z), fastener_plainShank(p, 6.0, 18.0, 0.0)), -drill_prismVoid(p, 0.0, drill_hexXY(p, 4.0), 2.2));
}

fn gsl_remap_8(p: vec3<f32>) -> f32 {
  return min(max(max((length(p.xy) - (6.72 + p.z * 1.0)) * 0.70710677, -3.72 - p.z), p.z), max(max(length(p.xy) - 3.0, -18.0 - p.z), p.z - 0.0));
}

fn gsl_remap_9(p: vec3<f32>) -> f32 {
  return min(max(max(max(drill_hexXY(p, 10.0), -p.z), p.z - 4.0), (length(p.xy) - 5.2 + p.z - 4.0) * 0.7071), fastener_plainShank(p, 6.0, 18.0, 2.0));
}

fn gsl_remap_10(p: vec3<f32>) -> f32 {
  return min(max(max(drill_hexXY(p, 10.0), -p.z), p.z - 4.0), max(max(length(p.xy) - 3.0, -18.0 - p.z), p.z - 2.0));
}

fn gsl_remap_11(p: vec3<f32>) -> f32 {
  return min(max(max(max(max(abs(p.x) - 5.0, abs(p.y) - 5.0), -p.z), p.z - 4.5), (length(p.xy) - 5.5 + p.z - 4.5) * 0.7071), fastener_plainShank(p, 6.0, 18.0, 0.0));
}

fn gsl_remap_12(p: vec3<f32>) -> f32 {
  return min(max(max(max(max(max(max(max(p.x - 5.0, 0.309017 * p.x + 0.951057 * p.y - 5.0), -0.809017 * p.x + 0.587785 * p.y - 5.0), -0.809017 * p.x - 0.587785 * p.y - 5.0), 0.309017 * p.x - 0.951057 * p.y - 5.0), -p.z), p.z - 4.5), (length(p.xy) - 5.3 + p.z - 4.5) * 0.7071), fastener_plainShank(p, 6.0, 18.0, 0.0));
}

fn gsl_remap_13(p: vec3<f32>) -> f32 {
  return max(min(min(max(max(length(p.xy) - 5.0, 0.75 - p.z), p.z - 6.0), max(max(max(length(p.xy) - 7.0, -p.z), p.z - 1.5), (length(p.xy) - 7.0 + p.z - 1.5 + 1.4) * 0.7071)), fastener_plainShank(p, 6.0, 18.0, 0.0)), -drill_prismVoid(p, 6.0, drill_hexXY(p, 5.0), 3.0));
}

fn gsl_remap_14(p: vec3<f32>) -> f32 {
  return max(min(100000.0, fastener_ended(fastener_plainShank(p, 6.0, 10.0, 0.0), min(fastener_flatTop(p, 6.0, 0.0, 0.4), fastener_flatEnd(p, 6.0, -10.0, 0.4)), 1000.0)), -drill_prismVoid(p, 0.0, drill_hexXY(p, 3.0), 5.0));
}

fn gsl_remap_15(p: vec3<f32>) -> f32 {
  return max(min(fastener_panHead(p, 12.0, 4.6, 10.0), fastener_plainShank(p, 6.0, 18.0, 0.0)), -drill_prismVoid(p, 4.6, drill_hexXY(p, 5.0), 3.0));
}

fn gsl_remap_16(p: vec3<f32>) -> f32 {
  return max(min(fastener_panHead(p, 12.0, 4.6, 10.0), fastener_plainShank(p, 6.0, 18.0, 0.0)), -max(max(max(max(abs(p.x) - 1.6795, abs(p.y) - 1.6795), 0.6838334 - p.z), (abs(p.x) - 1.6795 + 0.6838334 - p.z + 1.6795) * 0.70710677), (abs(p.y) - 1.6795 + 0.6838334 - p.z + 1.6795) * 0.70710677));
}

fn gsl_remap_17(p: vec3<f32>) -> f32 {
  let gsl_tmp_142 = length(p.xy);
  let gsl_tmp_143 = abs(p.x);
  let gsl_tmp_144 = abs(p.y);
  let gsl_tmp_145 = abs(gsl_tmp_143 - gsl_tmp_144) * 0.70710677;
  let gsl_tmp_146 = ((gsl_tmp_143 + gsl_tmp_144) * 0.70710677 - 1.205) * 0.99496853 - (p.z - 0.5150396) * 0.10018806;
  return max(min(fastener_panHead(p, 12.0, 4.6, 10.0), fastener_plainShank(p, 6.0, 18.0, 2.3)), -max(max(gsl_tmp_142 * 0.89493436 + (3.9860268 - p.z) * 0.4461978 - 3.0875235, (-0.12567022 - p.z) * 0.88294756 + gsl_tmp_142 * 0.46947157), -max(gsl_tmp_145 * 0.2923717 - gsl_tmp_146 * 0.9563047, (gsl_tmp_145 - 1.0025) * 0.6946584 - (gsl_tmp_146 - 0.306495) * 0.7193398)));
}

fn drill_engrave(p: vec3<f32>, surface: f32, section: f32, depth: f32, zFrom: f32) -> f32 {
  return max(max(section, -(surface + depth)), zFrom - p.z);
}

fn gsl_remap_18(p: vec3<f32>) -> f32 {
  let gsl_tmp_158 = (p.x - p.y) * 0.70710677;
  let gsl_tmp_157 = (p.x + p.y) * 0.70710677;
  let gsl_tmp_191 = length(p.xy);
  let gsl_tmp_192 = abs(p.x);
  let gsl_tmp_194 = (1.36 + (p.z - 1.1741018) * 0.10069471) * 0.70710677;
  let gsl_tmp_193 = abs(p.y);
  return max(min(fastener_panHead_2(p, 12.0, 4.6, 10.0, min(max(max(abs(gsl_tmp_158) - 0.5, abs(gsl_tmp_157) - 5.4), 3.9 - abs(gsl_tmp_157)), max(max(abs(gsl_tmp_157) - 0.5, abs(gsl_tmp_158) - 5.4), 3.9 - abs(gsl_tmp_158))), 0.45), fastener_plainShank(p, 6.0, 18.0, 0.0)), -min(max(max(gsl_tmp_191 * 0.89493436 + (4.0221815 - p.z) * 0.4461978 - 2.9980302, (0.14790268 - p.z) * 0.88294756 + gsl_tmp_191 * 0.46947157), min(min(abs(p.y) - 0.71, abs(p.x) - 0.71), max((gsl_tmp_192 - gsl_tmp_194) * 0.9772311 + (gsl_tmp_193 - gsl_tmp_194) * 0.21217768, (gsl_tmp_193 - gsl_tmp_194) * 0.9772311 + (gsl_tmp_192 - gsl_tmp_194) * 0.21217768) / 1.0035796)), (abs(abs(p.x) - abs(p.y)) * 0.70710677 * 0.8660254 - (p.z - (3.8221812 + ((abs(p.x) + abs(p.y)) * 0.70710677 - 1.6467866) * 0.11742509)) * 0.5) / 1.0017221));
}

fn gsl_remap_19(p: vec3<f32>) -> f32 {
  return length(vec2<f32>(p.x - 2.645, p.y)) - 0.62;
}

fn gsl_remap_20(p: vec3<f32>) -> f32 {
  let gsl_tmp_262 = length(p.xy);
  let gsl_tmp_263 = gsl_floored((atan2(p.y, p.x) + 0.5235988), 1.0471976) - 0.5235988;
  return max(min(fastener_panHead(p, 12.0, 4.6, 10.0), fastener_plainShank(p, 6.0, 18.0, 2.3)), -drill_prismVoid(p, 4.2, max(length(p.xy) - 2.8, -gsl_remap_19(vec3<f32>(gsl_tmp_262 * cos(gsl_tmp_263), gsl_tmp_262 * sin(gsl_tmp_263), p.z))), 2.42));
}

fn gsl_remap_21(p: vec3<f32>) -> f32 {
  return max(min(fastener_panHead(p, 12.0, 4.6, 10.0), fastener_plainShank(p, 6.0, 18.0, 0.0)), -drill_prismVoid(p, 4.6, abs(p.y) - 0.6, 2.0));
}

fn gsl_remap_22(p: vec3<f32>) -> f32 {
  return max(min(max(max(max(length(p.xy) - 6.5, -p.z), p.z - 6.0), (length(p.xy) - 6.5 + p.z - 6.0 + 0.2) * 0.7071), min(fastener_plainShank(p, 8.0, 12.0, 3.0), fastener_plainShank(vec3<f32>(p.x, p.y, p.z + 12.0), 5.0, 8.0, 1.25))), -drill_prismVoid(p, 6.0, drill_hexXY(p, 6.0), 3.5));
}

fn gsl_remap_23(p: vec3<f32>) -> f32 {
  let setBody = fastener_plainShank(p, 6.0, 10.0, 0.0);
  let setTop = -((length(p.xy) - 3.0 + p.z - 0.0 + 0.4) * 0.7071);
  let setSocket = drill_prismVoid(p, 0.0, max(max(abs(p.x) - 1.5, abs(0.5 * p.x + 0.8660254 * p.y) - 1.5), abs(0.5 * p.x - 0.8660254 * p.y) - 1.5), 5.0);
  return max(min(100000.0, fastener_ended(setBody, min(setTop, fastener_flatEnd(p, 6.0, -10.0, 0.4)), 1000.0)), -setSocket);
}

fn builtin_cone(p: vec3<f32>, origin: vec3<f32>, direction: vec3<f32>, angle: f32) -> f32 {
  let offset = p - origin;
  let unitDirection = normalize(direction);
  let along = dot(offset, unitDirection);
  let perp = length(offset - unitDirection * along);
  return perp * cos(angle) - along * sin(angle);
}

fn gsl_remap_24(p: vec3<f32>, vector_dirZ: vec3<f32>) -> f32 {
  let setBody = fastener_plainShank(p, 6.0, 10.0, 0.0);
  let setTop = -((length(p.xy) - 3.0 + p.z - 0.0 + 0.4) * 0.7071);
  let setSocket = drill_prismVoid(p, 0.0, max(max(abs(p.x) - 1.5, abs(0.5 * p.x + 0.8660254 * p.y) - 1.5), abs(0.5 * p.x - 0.8660254 * p.y) - 1.5), 5.0);
  return max(min(100000.0, fastener_ended(setBody, min(min(setTop, fastener_flatEnd(p, 6.0, -10.0, 0.4)), builtin_cone(p, vec3<f32>(0.0, 0.0, -8.8), vec3<f32>(-0.0, -0.0, -1.0), 1.0297443)), 1000.0)), -setSocket);
}

fn gsl_remap_25(p: vec3<f32>, vector_dirZ: vec3<f32>) -> f32 {
  let setBody = fastener_plainShank(p, 6.0, 10.0, 0.0);
  let setTop = -((length(p.xy) - 3.0 + p.z - 0.0 + 0.4) * 0.7071);
  let setSocket = drill_prismVoid(p, 0.0, max(max(abs(p.x) - 1.5, abs(0.5 * p.x + 0.8660254 * p.y) - 1.5), abs(0.5 * p.x - 0.8660254 * p.y) - 1.5), 5.0);
  return max(min(100000.0, fastener_ended(setBody, min(setTop, -builtin_cone(p, vec3<f32>(0.0, 0.0, -10.0), vector_dirZ, 0.7853982)), 1000.0)), -setSocket);
}

fn gsl_remap_26(p: vec3<f32>, vector_dirZ: vec3<f32>) -> f32 {
  let setBody = fastener_plainShank(p, 6.0, 10.0, 0.0);
  let setTop = -((length(p.xy) - 3.0 + p.z - 0.0 + 0.4) * 0.7071);
  let setSocket = drill_prismVoid(p, 0.0, max(max(abs(p.x) - 1.5, abs(0.5 * p.x + 0.8660254 * p.y) - 1.5), abs(0.5 * p.x - 0.8660254 * p.y) - 1.5), 5.0);
  return max(min(100000.0, fastener_ended(setBody, min(setTop, max(-(length(vec3<f32>(p.x, p.y, p.z - -4.0)) - 6.0), p.z - -9.196153)), 1000.0)), -setSocket);
}

fn gsl_remap_27(p: vec3<f32>, vector_dirZ: vec3<f32>) -> f32 {
  let setBody = fastener_plainShank(p, 6.0, 10.0, 0.0);
  let setTop = -((length(p.xy) - 3.0 + p.z - 0.0 + 0.4) * 0.7071);
  let setSocket = drill_prismVoid(p, 0.0, max(max(abs(p.x) - 1.5, abs(0.5 * p.x + 0.8660254 * p.y) - 1.5), abs(0.5 * p.x - 0.8660254 * p.y) - 1.5), 5.0);
  return max(min(100000.0, fastener_ended(setBody, min(setTop, max(-(length(vec3<f32>(p.x, p.y, p.z - -7.0)) - 3.0), p.z - -7.0)), 1000.0)), -setSocket);
}

fn gsl_remap_28(p: vec3<f32>, vector_dirZ: vec3<f32>) -> f32 {
  let setBody = fastener_plainShank(p, 6.0, 10.0, 0.0);
  let setTop = -((length(p.xy) - 3.0 + p.z - 0.0 + 0.4) * 0.7071);
  let setSocket = drill_prismVoid(p, 0.0, max(max(abs(p.x) - 1.5, abs(0.5 * p.x + 0.8660254 * p.y) - 1.5), abs(0.5 * p.x - 0.8660254 * p.y) - 1.5), 5.0);
  return max(min(100000.0, fastener_ended(setBody, setTop, max(max(max(length(p.xy) - 1.75, -12.5 - p.z), p.z - -9.125), (length(p.xy) - 1.75 + -12.5 - p.z + 0.525) * 0.7071))), -setSocket);
}

fn gsl_remap_29(p: vec3<f32>, vector_dirZ: vec3<f32>) -> f32 {
  let setBody = fastener_plainShank(p, 6.0, 10.0, 0.0);
  let setTop = -((length(p.xy) - 3.0 + p.z - 0.0 + 0.4) * 0.7071);
  let setSocket = drill_prismVoid(p, 0.0, max(max(abs(p.x) - 1.5, abs(0.5 * p.x + 0.8660254 * p.y) - 1.5), abs(0.5 * p.x - 0.8660254 * p.y) - 1.5), 5.0);
  return max(min(100000.0, fastener_ended(setBody, min(setTop, fastener_flatEnd(p, 6.0, -10.0, 0.4)), 1000.0)), -setSocket);
}

fn gsl_remap_30(p: vec3<f32>, vector_dirZ: vec3<f32>) -> f32 {
  return max(max(max(length(p.xy) - 2.458734, -12.0 - p.z), p.z - -8.770633), (length(p.xy) - 2.458734 + -12.0 - p.z + 0.9834936) * 0.7071);
}

fn gsl_remap_31(p: vec3<f32>, vector_dirZ: vec3<f32>) -> f32 {
  let setBody = fastener_plainShank(p, 6.0, 10.0, 0.0);
  let setTop = -((length(p.xy) - 3.0 + p.z - 0.0 + 0.4) * 0.7071);
  let setSocket = drill_prismVoid(p, 0.0, max(max(abs(p.x) - 1.5, abs(0.5 * p.x + 0.8660254 * p.y) - 1.5), abs(0.5 * p.x - 0.8660254 * p.y) - 1.5), 5.0);
  return max(min(100000.0, fastener_ended(setBody, setTop, 1000.0)), -setSocket);
}

fn gsl_remap_32(p: vec3<f32>, vector_dirZ: vec3<f32>) -> f32 {
  let gsl_tmp_261 = max(max(abs(p.y - 1.1) - 1.1, -p.x), p.z - -10.0);
  return gsl_tmp_261;
}

fn gsl_remap_33(p: vec3<f32>, vector_dirZ: vec3<f32>) -> f32 {
  let gsl_tmp_261 = max(max(abs(p.y - 1.1) - 1.1, -p.x), p.z - -10.0);
  return max(max(max(length(p.xy) - 2.458734, builtin_cone(p, vec3<f32>(0.0, 0.0, -16.0), vector_dirZ, 1.0297443)), p.z - -8.770633), -min(gsl_tmp_261, gsl_remap_32(vec3<f32>(-p.x, -p.y, p.z), vector_dirZ)));
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

fn thread_helixPhase(p: vec3<f32>, pitchVal: f32) -> f32 {
  let ang = 6.2831855 * p.z / max(pitchVal, 0.01);
  let radial = max(length(p.xy), 0.2);
  let along = cos(ang) * p.x / radial + sin(ang) * p.y / radial;
  return acos(clamp(along, -1.0, 1.0)) / 3.1415927;
}

fn thread_formDepth(pitchVal: f32) -> f32 {
  return 0.5412659 * pitchVal;
}

fn thread_profileDepth(p: vec3<f32>, pitchVal: f32) -> f32 {
  let offset = thread_helixPhase(p, pitchVal) * pitchVal * 0.5;
  return clamp((offset - pitchVal / 16.0) * 1.7320508, 0.0, thread_formDepth(pitchVal));
}

fn drill_hexXY(p: vec3<f32>, acrossFlats: f32) -> f32 {
  let h = acrossFlats / 2.0;
  return max(max(abs(p.x) - h, abs(0.5 * p.x + 0.8660254 * p.y) - h), abs(0.5 * p.x - 0.8660254 * p.y) - h);
}

fn drill_prismVoid(p: vec3<f32>, zTop: f32, section: f32, depth: f32) -> f32 {
  return max(section, zTop - depth - p.z);
}

fn fastener_plainShank(p: vec3<f32>, dia: f32, shankLength: f32, zTop: f32) -> f32 {
  let form = max(max(max(length(p.xy) - dia / 2.0, -shankLength - p.z), p.z - zTop), (length(p.xy) - dia / 2.0 + -p.z - shankLength + 0.15 * dia) * 0.7071);
  return form;
}

fn fastener_panHead(p: vec3<f32>, dk: f32, k: f32, crownR: f32) -> f32 {
  let head = max(max(length(p.xy) - dk / 2.0, -p.z), length(vec3<f32>(p.x, p.y, p.z - (k - crownR))) - crownR);
  return head;
}

fn fastener_flatTop(p: vec3<f32>, dia: f32, zTop: f32, leg: f32) -> f32 {
  return -((length(p.xy) - dia / 2.0 + p.z - zTop + leg) * 0.7071);
}

fn fastener_flatEnd(p: vec3<f32>, dia: f32, zEnd: f32, leg: f32) -> f32 {
  return -((length(p.xy) - dia / 2.0 + zEnd - p.z + leg) * 0.7071);
}

fn fastener_ended(shank: f32, cut: f32, add: f32) -> f32 {
  return min(max(shank, -cut), add);
}

fn fastener_markHead(p: vec3<f32>, head: f32, k: f32, markSection: f32, markDepth: f32) -> f32 {
  return max(head, -drill_engrave(p, head, markSection, markDepth, k / 2.0));
}

fn fastener_panHead_2(p: vec3<f32>, dk: f32, k: f32, crownR: f32, markSection: f32, markDepth: f32) -> f32 {
  return fastener_markHead(p, fastener_panHead(p, dk, k, crownR), k, markSection, markDepth);
}

fn gsl_threadColorOn_remap(p: vec3<f32>) -> f32 {
  return -min(min(thread_profileDepth(p, 1.0), max(0.0, 0.5 * min(min(0.0 - p.z, p.z - -18.0), 3.541266 - length(p.xy)))), max(0.0, length(p.xy) - 1.9174682));
}

fn gsl_threadColorOn_remap_2(p: vec3<f32>) -> f32 {
  return -min(min(thread_profileDepth(p, 1.0), max(0.0, 0.5 * min(min(-3.72 - p.z, p.z - -18.0), 3.541266 - length(p.xy)))), max(0.0, length(p.xy) - 1.9174682));
}

fn gsl_threadColorOn_remap_3(p: vec3<f32>) -> f32 {
  return -min(min(thread_profileDepth(p, 1.0), max(0.0, 0.5 * min(min(0.0 - p.z, p.z - -10.0), 3.541266 - length(p.xy)))), max(0.0, length(p.xy) - 1.9174682));
}

fn gsl_threadColorOn_remap_4(p: vec3<f32>) -> f32 {
  return -min(min(thread_profileDepth(p, 1.0), max(0.0, 0.5 * min(min(-9.0 - p.z, p.z - -18.0), 3.541266 - length(p.xy)))), max(0.0, length(p.xy) - 1.9174682));
}

fn gsl_threadColorOn_remap_5(p: vec3<f32>) -> f32 {
  return -min(min(thread_profileDepth(p, 0.8), max(0.0, 0.5 * min(min(-12.0 - p.z, p.z - -20.0), 2.9330127 - length(p.xy)))), max(0.0, length(p.xy) - 1.6339746));
}

fn gsl_threadColorOn_remap_6(p: vec3<f32>) -> f32 {
  return -min(min(thread_profileDepth(p, 0.75), max(0.0, 0.5 * min(min(0.0 - p.z, p.z - -18.0), 3.4059494 - length(p.xy)))), max(0.0, length(p.xy) - 2.188101));
}

fn gsl_shapeParts(p: vec3<f32>, time: f32) -> array<f32, 30> {
  gsl_partNear = vec3<f32>(1e10, 1e10, 1e10);
  let gsl_partLo = vec3<f32>(-69.01, 16.99, -18.01);
  let gsl_partHi = vec3<f32>(-58.99, 27.01, 6.01);
  let gsl_partBox = gsl_boxDistance(p, gsl_partLo, gsl_partHi);
  var gsl_part = gsl_partBox;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox <= gsl_vanishView.w) {
    let gsl_partVanish = gsl_vanish(gsl_partLo, gsl_partHi);
    gsl_part = 1e10;
    let gsl_simpleWeight = gsl_simple(vec3<f32>(-69.01, 16.99, -18.01), vec3<f32>(-58.99, 27.01, 6.01));
    var gsl_simplePart = 1e10;
    if (gsl_simpleWeight > 0.0) {
      gsl_simplePart = gsl_remap(p - vec3<f32>(-64.0, 22.0, 0.0));
    }
    if (gsl_simpleWeight < 1.0) {
      gsl_simplePart = mix(max(gsl_remap_2(p - vec3<f32>(-64.0, 22.0, 0.0)), gsl_boxSdf(p, vec3<f32>(-69.01, 16.99, -18.01), vec3<f32>(-58.99, 27.01, 6.01))), gsl_simplePart, gsl_simpleWeight);
    }
    if (gsl_partVanish.x > 0.0) { gsl_part = gsl_simplePart + gsl_partVanish.y; }
    if (abs(gsl_part) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part), gsl_partVanish.zw); }
  }
  let capHead = gsl_part;
  let gsl_partLo_2 = vec3<f32>(-53.26, 16.74, -18.01);
  let gsl_partHi_2 = vec3<f32>(-42.74, 27.26, 3.31);
  let gsl_partBox_2 = gsl_boxDistance(p, gsl_partLo_2, gsl_partHi_2);
  var gsl_part_2 = gsl_partBox_2;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_2 <= gsl_vanishView.w) {
    let gsl_partVanish_2 = gsl_vanish(gsl_partLo_2, gsl_partHi_2);
    gsl_part_2 = 1e10;
    let gsl_simpleWeight_2 = gsl_simple(vec3<f32>(-53.26, 16.74, -18.01), vec3<f32>(-42.74, 27.26, 3.31));
    var gsl_simplePart_2 = 1e10;
    if (gsl_simpleWeight_2 > 0.0) {
      gsl_simplePart_2 = gsl_remap_3(p - vec3<f32>(-48.0, 22.0, 0.0));
    }
    if (gsl_simpleWeight_2 < 1.0) {
      gsl_simplePart_2 = mix(max(gsl_remap_4(p - vec3<f32>(-48.0, 22.0, 0.0)), gsl_boxSdf(p, vec3<f32>(-53.26, 16.74, -18.01), vec3<f32>(-42.74, 27.26, 3.31))), gsl_simplePart_2, gsl_simpleWeight_2);
    }
    if (gsl_partVanish_2.x > 0.0) { gsl_part_2 = gsl_simplePart_2 + gsl_partVanish_2.y; }
    if (abs(gsl_part_2) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_2), gsl_partVanish_2.zw); }
  }
  let buttonH = gsl_part_2;
  let gsl_partLo_3 = vec3<f32>(-38.01, 15.99, -18.01);
  let gsl_partHi_3 = vec3<f32>(-25.99, 28.01, 4.61);
  let gsl_partBox_3 = gsl_boxDistance(p, gsl_partLo_3, gsl_partHi_3);
  var gsl_part_3 = gsl_partBox_3;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_3 <= gsl_vanishView.w) {
    let gsl_partVanish_3 = gsl_vanish(gsl_partLo_3, gsl_partHi_3);
    gsl_part_3 = 1e10;
    let gsl_simpleWeight_3 = gsl_simple(vec3<f32>(-38.01, 15.99, -18.01), vec3<f32>(-25.99, 28.01, 4.61));
    var gsl_simplePart_3 = 1e10;
    if (gsl_simpleWeight_3 > 0.0) {
      gsl_simplePart_3 = gsl_remap_5(p - vec3<f32>(-32.0, 22.0, 0.0));
    }
    if (gsl_simpleWeight_3 < 1.0) {
      gsl_simplePart_3 = mix(max(gsl_remap_6(p - vec3<f32>(-32.0, 22.0, 0.0)), gsl_boxSdf(p, vec3<f32>(-38.01, 15.99, -18.01), vec3<f32>(-25.99, 28.01, 4.61))), gsl_simplePart_3, gsl_simpleWeight_3);
    }
    if (gsl_partVanish_3.x > 0.0) { gsl_part_3 = gsl_simplePart_3 + gsl_partVanish_3.y; }
    if (abs(gsl_part_3) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_3), gsl_partVanish_3.zw); }
  }
  let panH = gsl_part_3;
  let gsl_partLo_4 = vec3<f32>(-22.73, 15.27, -18.01);
  let gsl_partHi_4 = vec3<f32>(-9.27, 28.73, 0.01);
  let gsl_partBox_4 = gsl_boxDistance(p, gsl_partLo_4, gsl_partHi_4);
  var gsl_part_4 = gsl_partBox_4;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_4 <= gsl_vanishView.w) {
    let gsl_partVanish_4 = gsl_vanish(gsl_partLo_4, gsl_partHi_4);
    gsl_part_4 = 1e10;
    let gsl_simpleWeight_4 = gsl_simple(vec3<f32>(-22.73, 15.27, -18.01), vec3<f32>(-9.27, 28.73, 0.01));
    var gsl_simplePart_4 = 1e10;
    if (gsl_simpleWeight_4 > 0.0) {
      gsl_simplePart_4 = gsl_remap_7(p - vec3<f32>(-16.0, 22.0, 0.0));
    }
    if (gsl_simpleWeight_4 < 1.0) {
      gsl_simplePart_4 = mix(max(gsl_remap_8(p - vec3<f32>(-16.0, 22.0, 0.0)), gsl_boxSdf(p, vec3<f32>(-22.73, 15.27, -18.01), vec3<f32>(-9.27, 28.73, 0.01))), gsl_simplePart_4, gsl_simpleWeight_4);
    }
    if (gsl_partVanish_4.x > 0.0) { gsl_part_4 = gsl_simplePart_4 + gsl_partVanish_4.y; }
    if (abs(gsl_part_4) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_4), gsl_partVanish_4.zw); }
  }
  let flatH = gsl_part_4;
  let gsl_partLo_5 = vec3<f32>(-5.01, 16.216497, -18.01);
  let gsl_partHi_5 = vec3<f32>(5.01, 27.783503, 4.01);
  let gsl_partBox_5 = gsl_boxDistance(p, gsl_partLo_5, gsl_partHi_5);
  var gsl_part_5 = gsl_partBox_5;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_5 <= gsl_vanishView.w) {
    let gsl_partVanish_5 = gsl_vanish(gsl_partLo_5, gsl_partHi_5);
    gsl_part_5 = 1e10;
    let gsl_simpleWeight_5 = gsl_simple(vec3<f32>(-5.01, 16.216497, -18.01), vec3<f32>(5.01, 27.783503, 4.01));
    var gsl_simplePart_5 = 1e10;
    if (gsl_simpleWeight_5 > 0.0) {
      gsl_simplePart_5 = gsl_remap_9(p - vec3<f32>(0.0, 22.0, 0.0));
    }
    if (gsl_simpleWeight_5 < 1.0) {
      gsl_simplePart_5 = mix(max(gsl_remap_10(p - vec3<f32>(0.0, 22.0, 0.0)), gsl_boxSdf(p, vec3<f32>(-5.01, 16.216497, -18.01), vec3<f32>(5.01, 27.783503, 4.01))), gsl_simplePart_5, gsl_simpleWeight_5);
    }
    if (gsl_partVanish_5.x > 0.0) { gsl_part_5 = gsl_simplePart_5 + gsl_partVanish_5.y; }
    if (abs(gsl_part_5) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_5), gsl_partVanish_5.zw); }
  }
  let hexH = gsl_part_5;
  let gsl_partLo_6 = vec3<f32>(10.99, 16.99, -18.01);
  let gsl_partHi_6 = vec3<f32>(21.01, 27.01, 4.51);
  let gsl_partBox_6 = gsl_boxDistance(p, gsl_partLo_6, gsl_partHi_6);
  var gsl_part_6 = gsl_partBox_6;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_6 <= gsl_vanishView.w) {
    let gsl_partVanish_6 = gsl_vanish(gsl_partLo_6, gsl_partHi_6);
    gsl_part_6 = 1e10;
    if (gsl_partVanish_6.x > 0.0) { gsl_part_6 = gsl_remap_11(p - vec3<f32>(16.0, 22.0, 0.0)) + gsl_partVanish_6.y; }
    if (abs(gsl_part_6) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_6), gsl_partVanish_6.zw); }
  }
  let squareH = gsl_part_6;
  let gsl_partLo_7 = vec3<f32>(25.80966, 16.11215, -18.01);
  let gsl_partHi_7 = vec3<f32>(37.01, 27.88785, 4.51);
  let gsl_partBox_7 = gsl_boxDistance(p, gsl_partLo_7, gsl_partHi_7);
  var gsl_part_7 = gsl_partBox_7;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_7 <= gsl_vanishView.w) {
    let gsl_partVanish_7 = gsl_vanish(gsl_partLo_7, gsl_partHi_7);
    gsl_part_7 = 1e10;
    if (gsl_partVanish_7.x > 0.0) { gsl_part_7 = gsl_remap_12(p - vec3<f32>(32.0, 22.0, 0.0)) + gsl_partVanish_7.y; }
    if (abs(gsl_part_7) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_7), gsl_partVanish_7.zw); }
  }
  let pentH = gsl_part_7;
  let gsl_partLo_8 = vec3<f32>(40.99, 14.99, -18.01);
  let gsl_partHi_8 = vec3<f32>(55.01, 29.01, 6.01);
  let gsl_partBox_8 = gsl_boxDistance(p, gsl_partLo_8, gsl_partHi_8);
  var gsl_part_8 = gsl_partBox_8;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_8 <= gsl_vanishView.w) {
    let gsl_partVanish_8 = gsl_vanish(gsl_partLo_8, gsl_partHi_8);
    gsl_part_8 = 1e10;
    if (gsl_partVanish_8.x > 0.0) { gsl_part_8 = gsl_remap_13(p - vec3<f32>(48.0, 22.0, 0.0)) + gsl_partVanish_8.y; }
    if (abs(gsl_part_8) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_8), gsl_partVanish_8.zw); }
  }
  let flangeH = gsl_part_8;
  let gsl_partLo_9 = vec3<f32>(60.99, 18.99, -10.01);
  let gsl_partHi_9 = vec3<f32>(67.01, 25.01, 0.01);
  let gsl_partBox_9 = gsl_boxDistance(p, gsl_partLo_9, gsl_partHi_9);
  var gsl_part_9 = gsl_partBox_9;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_9 <= gsl_vanishView.w) {
    let gsl_partVanish_9 = gsl_vanish(gsl_partLo_9, gsl_partHi_9);
    gsl_part_9 = 1e10;
    if (gsl_partVanish_9.x > 0.0) { gsl_part_9 = gsl_remap_14(p - vec3<f32>(64.0, 22.0, 0.0)) + gsl_partVanish_9.y; }
    if (abs(gsl_part_9) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_9), gsl_partVanish_9.zw); }
  }
  let setScrew = gsl_part_9;
  let gsl_partLo_10 = vec3<f32>(-46.01, -6.01, -18.01);
  let gsl_partHi_10 = vec3<f32>(-33.99, 6.01, 4.61);
  let gsl_partBox_10 = gsl_boxDistance(p, gsl_partLo_10, gsl_partHi_10);
  var gsl_part_10 = gsl_partBox_10;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_10 <= gsl_vanishView.w) {
    let gsl_partVanish_10 = gsl_vanish(gsl_partLo_10, gsl_partHi_10);
    gsl_part_10 = 1e10;
    if (gsl_partVanish_10.x > 0.0) { gsl_part_10 = gsl_remap_15(p - vec3<f32>(-40.0, 0.0, 0.0)) + gsl_partVanish_10.y; }
    if (abs(gsl_part_10) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_10), gsl_partVanish_10.zw); }
  }
  let hexDrive = gsl_part_10;
  let gsl_partLo_11 = vec3<f32>(-30.01, -6.01, -18.01);
  let gsl_partHi_11 = vec3<f32>(-17.99, 6.01, 4.61);
  let gsl_partBox_11 = gsl_boxDistance(p, gsl_partLo_11, gsl_partHi_11);
  var gsl_part_11 = gsl_partBox_11;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_11 <= gsl_vanishView.w) {
    let gsl_partVanish_11 = gsl_vanish(gsl_partLo_11, gsl_partHi_11);
    gsl_part_11 = 1e10;
    if (gsl_partVanish_11.x > 0.0) { gsl_part_11 = gsl_remap_16(p - vec3<f32>(-24.0, 0.0, 0.0)) + gsl_partVanish_11.y; }
    if (abs(gsl_part_11) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_11), gsl_partVanish_11.zw); }
  }
  let sqDrive = gsl_part_11;
  let gsl_partLo_12 = vec3<f32>(-14.01, -6.01, -18.01);
  let gsl_partHi_12 = vec3<f32>(-1.99, 6.01, 4.61);
  let gsl_partBox_12 = gsl_boxDistance(p, gsl_partLo_12, gsl_partHi_12);
  var gsl_part_12 = gsl_partBox_12;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_12 <= gsl_vanishView.w) {
    let gsl_partVanish_12 = gsl_vanish(gsl_partLo_12, gsl_partHi_12);
    gsl_part_12 = 1e10;
    let gsl_simpleWeight_6 = gsl_simple(vec3<f32>(-14.01, -6.01, -18.01), vec3<f32>(-1.99, 6.01, 4.61));
    var gsl_simplePart_6 = 1e10;
    if (gsl_simpleWeight_6 > 0.0) {
      gsl_simplePart_6 = gsl_remap_17(p - vec3<f32>(-8.0, 0.0, 0.0));
    }
    if (gsl_simpleWeight_6 < 1.0) {
      gsl_simplePart_6 = mix(max(gsl_remap_6(p - vec3<f32>(-8.0, 0.0, 0.0)), gsl_boxSdf(p, vec3<f32>(-14.01, -6.01, -18.01), vec3<f32>(-1.99, 6.01, 4.61))), gsl_simplePart_6, gsl_simpleWeight_6);
    }
    if (gsl_partVanish_12.x > 0.0) { gsl_part_12 = gsl_simplePart_6 + gsl_partVanish_12.y; }
    if (abs(gsl_part_12) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_12), gsl_partVanish_12.zw); }
  }
  let phDrive = gsl_part_12;
  let gsl_partLo_13 = vec3<f32>(1.99, -6.01, -18.01);
  let gsl_partHi_13 = vec3<f32>(14.01, 6.01, 4.61);
  let gsl_partBox_13 = gsl_boxDistance(p, gsl_partLo_13, gsl_partHi_13);
  var gsl_part_13 = gsl_partBox_13;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_13 <= gsl_vanishView.w) {
    let gsl_partVanish_13 = gsl_vanish(gsl_partLo_13, gsl_partHi_13);
    gsl_part_13 = 1e10;
    if (gsl_partVanish_13.x > 0.0) { gsl_part_13 = gsl_remap_18(p - vec3<f32>(8.0, 0.0, 0.0)) + gsl_partVanish_13.y; }
    if (abs(gsl_part_13) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_13), gsl_partVanish_13.zw); }
  }
  let pozDrive = gsl_part_13;
  let gsl_partLo_14 = vec3<f32>(17.99, -6.01, -18.01);
  let gsl_partHi_14 = vec3<f32>(30.01, 6.01, 4.61);
  let gsl_partBox_14 = gsl_boxDistance(p, gsl_partLo_14, gsl_partHi_14);
  var gsl_part_14 = gsl_partBox_14;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_14 <= gsl_vanishView.w) {
    let gsl_partVanish_14 = gsl_vanish(gsl_partLo_14, gsl_partHi_14);
    gsl_part_14 = 1e10;
    let gsl_simpleWeight_7 = gsl_simple(vec3<f32>(17.99, -6.01, -18.01), vec3<f32>(30.01, 6.01, 4.61));
    var gsl_simplePart_7 = 1e10;
    if (gsl_simpleWeight_7 > 0.0) {
      gsl_simplePart_7 = gsl_remap_20(p - vec3<f32>(24.0, 0.0, 0.0));
    }
    if (gsl_simpleWeight_7 < 1.0) {
      gsl_simplePart_7 = mix(max(gsl_remap_6(p - vec3<f32>(24.0, 0.0, 0.0)), gsl_boxSdf(p, vec3<f32>(17.99, -6.01, -18.01), vec3<f32>(30.01, 6.01, 4.61))), gsl_simplePart_7, gsl_simpleWeight_7);
    }
    if (gsl_partVanish_14.x > 0.0) { gsl_part_14 = gsl_simplePart_7 + gsl_partVanish_14.y; }
    if (abs(gsl_part_14) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_14), gsl_partVanish_14.zw); }
  }
  let torxDrive = gsl_part_14;
  let gsl_partLo_15 = vec3<f32>(33.99, -6.01, -18.01);
  let gsl_partHi_15 = vec3<f32>(46.01, 6.01, 4.61);
  let gsl_partBox_15 = gsl_boxDistance(p, gsl_partLo_15, gsl_partHi_15);
  var gsl_part_15 = gsl_partBox_15;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_15 <= gsl_vanishView.w) {
    let gsl_partVanish_15 = gsl_vanish(gsl_partLo_15, gsl_partHi_15);
    gsl_part_15 = 1e10;
    if (gsl_partVanish_15.x > 0.0) { gsl_part_15 = gsl_remap_21(p - vec3<f32>(40.0, 0.0, 0.0)) + gsl_partVanish_15.y; }
    if (abs(gsl_part_15) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_15), gsl_partVanish_15.zw); }
  }
  let slotDrive = gsl_part_15;
  let gsl_partLo_16 = vec3<f32>(-37.01, -27.01, -18.01);
  let gsl_partHi_16 = vec3<f32>(-26.99, -16.99, 6.01);
  let gsl_partBox_16 = gsl_boxDistance(p, gsl_partLo_16, gsl_partHi_16);
  var gsl_part_16 = gsl_partBox_16;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_16 <= gsl_vanishView.w) {
    let gsl_partVanish_16 = gsl_vanish(gsl_partLo_16, gsl_partHi_16);
    gsl_part_16 = 1e10;
    let gsl_simpleWeight_8 = gsl_simple(vec3<f32>(-37.01, -27.01, -18.01), vec3<f32>(-26.99, -16.99, 6.01));
    var gsl_simplePart_8 = 1e10;
    if (gsl_simpleWeight_8 > 0.0) {
      gsl_simplePart_8 = gsl_remap(p - vec3<f32>(-32.0, -22.0, 0.0));
    }
    if (gsl_simpleWeight_8 < 1.0) {
      gsl_simplePart_8 = mix(max(gsl_remap_2(p - vec3<f32>(-32.0, -22.0, 0.0)), gsl_boxSdf(p, vec3<f32>(-37.01, -27.01, -18.01), vec3<f32>(-26.99, -16.99, 6.01))), gsl_simplePart_8, gsl_simpleWeight_8);
    }
    if (gsl_partVanish_16.x > 0.0) { gsl_part_16 = gsl_simplePart_8 + gsl_partVanish_16.y; }
    if (abs(gsl_part_16) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_16), gsl_partVanish_16.zw); }
  }
  let plainSh = gsl_part_16;
  let gsl_partLo_17 = vec3<f32>(-21.01, -27.01, -18.01);
  let gsl_partHi_17 = vec3<f32>(-10.99, -16.99, 6.01);
  let gsl_partBox_17 = gsl_boxDistance(p, gsl_partLo_17, gsl_partHi_17);
  var gsl_part_17 = gsl_partBox_17;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_17 <= gsl_vanishView.w) {
    let gsl_partVanish_17 = gsl_vanish(gsl_partLo_17, gsl_partHi_17);
    gsl_part_17 = 1e10;
    let gsl_simpleWeight_9 = gsl_simple(vec3<f32>(-21.01, -27.01, -18.01), vec3<f32>(-10.99, -16.99, 6.01));
    var gsl_simplePart_9 = 1e10;
    if (gsl_simpleWeight_9 > 0.0) {
      gsl_simplePart_9 = gsl_remap(p - vec3<f32>(-16.0, -22.0, 0.0));
    }
    if (gsl_simpleWeight_9 < 1.0) {
      gsl_simplePart_9 = mix(max(gsl_remap_2(p - vec3<f32>(-16.0, -22.0, 0.0)), gsl_boxSdf(p, vec3<f32>(-21.01, -27.01, -18.01), vec3<f32>(-10.99, -16.99, 6.01))), gsl_simplePart_9, gsl_simpleWeight_9);
    }
    if (gsl_partVanish_17.x > 0.0) { gsl_part_17 = gsl_simplePart_9 + gsl_partVanish_17.y; }
    if (abs(gsl_part_17) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_17), gsl_partVanish_17.zw); }
  }
  let fullTh = gsl_part_17;
  let gsl_partLo_18 = vec3<f32>(-5.01, -27.01, -18.01);
  let gsl_partHi_18 = vec3<f32>(5.01, -16.99, 6.01);
  let gsl_partBox_18 = gsl_boxDistance(p, gsl_partLo_18, gsl_partHi_18);
  var gsl_part_18 = gsl_partBox_18;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_18 <= gsl_vanishView.w) {
    let gsl_partVanish_18 = gsl_vanish(gsl_partLo_18, gsl_partHi_18);
    gsl_part_18 = 1e10;
    let gsl_simpleWeight_10 = gsl_simple(vec3<f32>(-5.01, -27.01, -18.01), vec3<f32>(5.01, -16.99, 6.01));
    var gsl_simplePart_10 = 1e10;
    if (gsl_simpleWeight_10 > 0.0) {
      gsl_simplePart_10 = gsl_remap(p - vec3<f32>(0.0, -22.0, 0.0));
    }
    if (gsl_simpleWeight_10 < 1.0) {
      gsl_simplePart_10 = mix(max(gsl_remap_2(p - vec3<f32>(0.0, -22.0, 0.0)), gsl_boxSdf(p, vec3<f32>(-5.01, -27.01, -18.01), vec3<f32>(5.01, -16.99, 6.01))), gsl_simplePart_10, gsl_simpleWeight_10);
    }
    if (gsl_partVanish_18.x > 0.0) { gsl_part_18 = gsl_simplePart_10 + gsl_partVanish_18.y; }
    if (abs(gsl_part_18) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_18), gsl_partVanish_18.zw); }
  }
  let partTh = gsl_part_18;
  let shoulder = gsl_remap_22(p - vec3<f32>(16.0, -22.0, 0.0));
  let gsl_partLo_19 = vec3<f32>(26.99, -27.01, -18.01);
  let gsl_partHi_19 = vec3<f32>(37.01, -16.99, 6.01);
  let gsl_partBox_19 = gsl_boxDistance(p, gsl_partLo_19, gsl_partHi_19);
  var gsl_part_19 = gsl_partBox_19;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_19 <= gsl_vanishView.w) {
    let gsl_partVanish_19 = gsl_vanish(gsl_partLo_19, gsl_partHi_19);
    gsl_part_19 = 1e10;
    let gsl_simpleWeight_11 = gsl_simple(vec3<f32>(26.99, -27.01, -18.01), vec3<f32>(37.01, -16.99, 6.01));
    var gsl_simplePart_11 = 1e10;
    if (gsl_simpleWeight_11 > 0.0) {
      gsl_simplePart_11 = gsl_remap(p - vec3<f32>(32.0, -22.0, 0.0));
    }
    if (gsl_simpleWeight_11 < 1.0) {
      gsl_simplePart_11 = mix(max(gsl_remap_2(p - vec3<f32>(32.0, -22.0, 0.0)), gsl_boxSdf(p, vec3<f32>(26.99, -27.01, -18.01), vec3<f32>(37.01, -16.99, 6.01))), gsl_simplePart_11, gsl_simpleWeight_11);
    }
    if (gsl_partVanish_19.x > 0.0) { gsl_part_19 = gsl_simplePart_11 + gsl_partVanish_19.y; }
    if (abs(gsl_part_19) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_19), gsl_partVanish_19.zw); }
  }
  let fineTh = gsl_part_19;
  let gsl_partLo_20 = vec3<f32>(-59.01, -47.01, -10.01);
  let gsl_partHi_20 = vec3<f32>(-52.99, -40.99, 0.01);
  let gsl_partBox_20 = gsl_boxDistance(p, gsl_partLo_20, gsl_partHi_20);
  var gsl_part_20 = gsl_partBox_20;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_20 <= gsl_vanishView.w) {
    let gsl_partVanish_20 = gsl_vanish(gsl_partLo_20, gsl_partHi_20);
    gsl_part_20 = 1e10;
    if (gsl_partVanish_20.x > 0.0) { gsl_part_20 = gsl_remap_23(p - vec3<f32>(-56.0, -44.0, 0.0)) + gsl_partVanish_20.y; }
    if (abs(gsl_part_20) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_20), gsl_partVanish_20.zw); }
  }
  let flatPt = gsl_part_20;
  let gsl_partLo_21 = vec3<f32>(-43.01, -47.01, -10.01);
  let gsl_partHi_21 = vec3<f32>(-36.99, -40.99, 0.01);
  let gsl_partBox_21 = gsl_boxDistance(p, gsl_partLo_21, gsl_partHi_21);
  var gsl_part_21 = gsl_partBox_21;
  let vector_dirZ = vec3<f32>(0.0, 0.0, 1.0);
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_21 <= gsl_vanishView.w) {
    let gsl_partVanish_21 = gsl_vanish(gsl_partLo_21, gsl_partHi_21);
    gsl_part_21 = 1e10;
    if (gsl_partVanish_21.x > 0.0) { gsl_part_21 = gsl_remap_24(p - vec3<f32>(-40.0, -44.0, 0.0), vector_dirZ) + gsl_partVanish_21.y; }
    if (abs(gsl_part_21) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_21), gsl_partVanish_21.zw); }
  }
  let cupPt = gsl_part_21;
  let gsl_partLo_22 = vec3<f32>(-27.01, -47.01, -10.01);
  let gsl_partHi_22 = vec3<f32>(-20.99, -40.99, 0.01);
  let gsl_partBox_22 = gsl_boxDistance(p, gsl_partLo_22, gsl_partHi_22);
  var gsl_part_22 = gsl_partBox_22;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_22 <= gsl_vanishView.w) {
    let gsl_partVanish_22 = gsl_vanish(gsl_partLo_22, gsl_partHi_22);
    gsl_part_22 = 1e10;
    if (gsl_partVanish_22.x > 0.0) { gsl_part_22 = gsl_remap_25(p - vec3<f32>(-24.0, -44.0, 0.0), vector_dirZ) + gsl_partVanish_22.y; }
    if (abs(gsl_part_22) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_22), gsl_partVanish_22.zw); }
  }
  let conePt = gsl_part_22;
  let gsl_partLo_23 = vec3<f32>(-11.01, -47.01, -10.01);
  let gsl_partHi_23 = vec3<f32>(-4.99, -40.99, 0.01);
  let gsl_partBox_23 = gsl_boxDistance(p, gsl_partLo_23, gsl_partHi_23);
  var gsl_part_23 = gsl_partBox_23;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_23 <= gsl_vanishView.w) {
    let gsl_partVanish_23 = gsl_vanish(gsl_partLo_23, gsl_partHi_23);
    gsl_part_23 = 1e10;
    if (gsl_partVanish_23.x > 0.0) { gsl_part_23 = gsl_remap_26(p - vec3<f32>(-8.0, -44.0, 0.0), vector_dirZ) + gsl_partVanish_23.y; }
    if (abs(gsl_part_23) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_23), gsl_partVanish_23.zw); }
  }
  let ovalPt = gsl_part_23;
  let gsl_partLo_24 = vec3<f32>(4.99, -47.01, -10.01);
  let gsl_partHi_24 = vec3<f32>(11.01, -40.99, 0.01);
  let gsl_partBox_24 = gsl_boxDistance(p, gsl_partLo_24, gsl_partHi_24);
  var gsl_part_24 = gsl_partBox_24;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_24 <= gsl_vanishView.w) {
    let gsl_partVanish_24 = gsl_vanish(gsl_partLo_24, gsl_partHi_24);
    gsl_part_24 = 1e10;
    if (gsl_partVanish_24.x > 0.0) { gsl_part_24 = gsl_remap_27(p - vec3<f32>(8.0, -44.0, 0.0), vector_dirZ) + gsl_partVanish_24.y; }
    if (abs(gsl_part_24) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_24), gsl_partVanish_24.zw); }
  }
  let ballPt = gsl_part_24;
  let gsl_partLo_25 = vec3<f32>(20.99, -47.01, -12.51);
  let gsl_partHi_25 = vec3<f32>(27.01, -40.99, 0.01);
  let gsl_partBox_25 = gsl_boxDistance(p, gsl_partLo_25, gsl_partHi_25);
  var gsl_part_25 = gsl_partBox_25;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_25 <= gsl_vanishView.w) {
    let gsl_partVanish_25 = gsl_vanish(gsl_partLo_25, gsl_partHi_25);
    gsl_part_25 = 1e10;
    if (gsl_partVanish_25.x > 0.0) { gsl_part_25 = gsl_remap_28(p - vec3<f32>(24.0, -44.0, 0.0), vector_dirZ) + gsl_partVanish_25.y; }
    if (abs(gsl_part_25) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_25), gsl_partVanish_25.zw); }
  }
  let dogPt = gsl_part_25;
  let gsl_partLo_26 = vec3<f32>(36.99, -47.01, -10.01);
  let gsl_partHi_26 = vec3<f32>(43.01, -40.99, 0.01);
  let gsl_partBox_26 = gsl_boxDistance(p, gsl_partLo_26, gsl_partHi_26);
  var gsl_part_26 = gsl_partBox_26;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_26 <= gsl_vanishView.w) {
    let gsl_partVanish_26 = gsl_vanish(gsl_partLo_26, gsl_partHi_26);
    gsl_part_26 = 1e10;
    if (gsl_partVanish_26.x > 0.0) { gsl_part_26 = gsl_remap_29(p - vec3<f32>(40.0, -44.0, 0.0), vector_dirZ) + gsl_partVanish_26.y; }
    if (abs(gsl_part_26) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_26), gsl_partVanish_26.zw); }
  }
  let softBody = gsl_part_26;
  let gsl_partLo_27 = vec3<f32>(37.531265, -46.468735, -12.01);
  let gsl_partHi_27 = vec3<f32>(42.468735, -41.531265, -8.7606325);
  let gsl_partBox_27 = gsl_boxDistance(p, gsl_partLo_27, gsl_partHi_27);
  var gsl_part_27 = gsl_partBox_27;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_27 <= gsl_vanishView.w) {
    let gsl_partVanish_27 = gsl_vanish(gsl_partLo_27, gsl_partHi_27);
    gsl_part_27 = 1e10;
    if (gsl_partVanish_27.x > 0.0) { gsl_part_27 = gsl_remap_30(p - vec3<f32>(40.0, -44.0, 0.0), vector_dirZ) + gsl_partVanish_27.y; }
    if (abs(gsl_part_27) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_27), gsl_partVanish_27.zw); }
  }
  let softPad = gsl_part_27;
  let gsl_partLo_28 = vec3<f32>(52.99, -47.01, -10.01);
  let gsl_partHi_28 = vec3<f32>(59.01, -40.99, 0.01);
  let gsl_partBox_28 = gsl_boxDistance(p, gsl_partLo_28, gsl_partHi_28);
  var gsl_part_28 = gsl_partBox_28;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_28 <= gsl_vanishView.w) {
    let gsl_partVanish_28 = gsl_vanish(gsl_partLo_28, gsl_partHi_28);
    gsl_part_28 = 1e10;
    if (gsl_partVanish_28.x > 0.0) { gsl_part_28 = gsl_remap_31(p - vec3<f32>(56.0, -44.0, 0.0), vector_dirZ) + gsl_partVanish_28.y; }
    if (abs(gsl_part_28) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_28), gsl_partVanish_28.zw); }
  }
  let drillPt = gsl_part_28;
  let gsl_partLo_29 = vec3<f32>(53.531265, -46.468735, -16.01);
  let gsl_partHi_29 = vec3<f32>(58.468735, -41.531265, -8.7606325);
  let gsl_partBox_29 = gsl_boxDistance(p, gsl_partLo_29, gsl_partHi_29);
  var gsl_part_29 = gsl_partBox_29;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_29 <= gsl_vanishView.w) {
    let gsl_partVanish_29 = gsl_vanish(gsl_partLo_29, gsl_partHi_29);
    gsl_part_29 = 1e10;
    if (gsl_partVanish_29.x > 0.0) { gsl_part_29 = gsl_remap_33(p - vec3<f32>(56.0, -44.0, 0.0), vector_dirZ) + gsl_partVanish_29.y; }
    if (abs(gsl_part_29) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_29), gsl_partVanish_29.zw); }
  }
  let drillTip = gsl_part_29;
  return array<f32, 30>(capHead, buttonH, panH, flatH, hexH, squareH, pentH, flangeH, setScrew, hexDrive, sqDrive, phDrive, pozDrive, torxDrive, slotDrive, plainSh, fullTh, partTh, shoulder, fineTh, flatPt, cupPt, conePt, ovalPt, ballPt, dogPt, softBody, softPad, drillPt, drillTip);
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
  let gsl_s25 = gsl_shape(gsl_d[24], 25u);
  let gsl_s26 = gsl_shape(gsl_d[25], 26u);
  let gsl_s27 = gsl_shape(gsl_d[26], 27u);
  let gsl_s28 = gsl_shape(gsl_d[27], 28u);
  let gsl_s29 = gsl_shape(gsl_d[28], 29u);
  let gsl_s30 = gsl_shape(gsl_d[29], 30u);
  return gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_s1, gsl_s2), gsl_s3), gsl_s4), gsl_s5), gsl_s6), gsl_s7), gsl_s8), gsl_s9), gsl_s10), gsl_s11), gsl_s12), gsl_s13), gsl_s14), gsl_s15), gsl_s16), gsl_s17), gsl_s18), gsl_s19), gsl_s20), gsl_s21), gsl_s22), gsl_s23), gsl_s24), gsl_s25), gsl_s26), gsl_s27), gsl_s28), gsl_s29), gsl_s30);
}
fn gsl_colorField(p: vec3<f32>, time: f32, gsl_colorIdx: u32) -> vec4<f32> {
  if (gsl_colorIdx == 1u) { return vec4<f32>(thread_reliefColor(vec3<f32>(0.77, 0.78, 0.78), gsl_threadColorOn_remap(vec3<f32>(p.x - -64.0, p.y - 22.0, p.z)), 1.0), 1.0); }
  if (gsl_colorIdx == 2u) { return vec4<f32>(thread_reliefColor(vec3<f32>(0.77, 0.78, 0.78), gsl_threadColorOn_remap(vec3<f32>(p.x - -48.0, p.y - 22.0, p.z)), 1.0), 1.0); }
  if (gsl_colorIdx == 3u) { return vec4<f32>(thread_reliefColor(vec3<f32>(0.77, 0.78, 0.78), gsl_threadColorOn_remap(vec3<f32>(p.x - -32.0, p.y - 22.0, p.z)), 1.0), 1.0); }
  if (gsl_colorIdx == 4u) { return vec4<f32>(thread_reliefColor(vec3<f32>(0.77, 0.78, 0.78), gsl_threadColorOn_remap_2(vec3<f32>(p.x - -16.0, p.y - 22.0, p.z)), 1.0), 1.0); }
  if (gsl_colorIdx == 5u) { return vec4<f32>(thread_reliefColor(vec3<f32>(0.77, 0.78, 0.78), gsl_threadColorOn_remap(vec3<f32>(p.x - 0.0, p.y - 22.0, p.z)), 1.0), 1.0); }
  if (gsl_colorIdx == 6u) { return vec4<f32>(thread_reliefColor(vec3<f32>(0.77, 0.78, 0.78), gsl_threadColorOn_remap(vec3<f32>(p.x - 16.0, p.y - 22.0, p.z)), 1.0), 1.0); }
  if (gsl_colorIdx == 7u) { return vec4<f32>(thread_reliefColor(vec3<f32>(0.77, 0.78, 0.78), gsl_threadColorOn_remap(vec3<f32>(p.x - 32.0, p.y - 22.0, p.z)), 1.0), 1.0); }
  if (gsl_colorIdx == 8u) { return vec4<f32>(thread_reliefColor(vec3<f32>(0.77, 0.78, 0.78), gsl_threadColorOn_remap(vec3<f32>(p.x - 48.0, p.y - 22.0, p.z)), 1.0), 1.0); }
  if (gsl_colorIdx == 9u) { return vec4<f32>(thread_reliefColor(vec3<f32>(0.77, 0.78, 0.78), gsl_threadColorOn_remap_3(vec3<f32>(p.x - 64.0, p.y - 22.0, p.z)), 1.0), 1.0); }
  if (gsl_colorIdx == 10u) { return vec4<f32>(thread_reliefColor(vec3<f32>(0.77, 0.78, 0.78), gsl_threadColorOn_remap(vec3<f32>(p.x - -40.0, p.y - 0.0, p.z)), 1.0), 1.0); }
  if (gsl_colorIdx == 11u) { return vec4<f32>(thread_reliefColor(vec3<f32>(0.77, 0.78, 0.78), gsl_threadColorOn_remap(vec3<f32>(p.x - -24.0, p.y - 0.0, p.z)), 1.0), 1.0); }
  if (gsl_colorIdx == 12u) { return vec4<f32>(thread_reliefColor(vec3<f32>(0.77, 0.78, 0.78), gsl_threadColorOn_remap(vec3<f32>(p.x - -8.0, p.y - 0.0, p.z)), 1.0), 1.0); }
  if (gsl_colorIdx == 13u) { return vec4<f32>(thread_reliefColor(vec3<f32>(0.77, 0.78, 0.78), gsl_threadColorOn_remap(vec3<f32>(p.x - 8.0, p.y - 0.0, p.z)), 1.0), 1.0); }
  if (gsl_colorIdx == 14u) { return vec4<f32>(thread_reliefColor(vec3<f32>(0.77, 0.78, 0.78), gsl_threadColorOn_remap(vec3<f32>(p.x - 24.0, p.y - 0.0, p.z)), 1.0), 1.0); }
  if (gsl_colorIdx == 15u) { return vec4<f32>(thread_reliefColor(vec3<f32>(0.77, 0.78, 0.78), gsl_threadColorOn_remap(vec3<f32>(p.x - 40.0, p.y - 0.0, p.z)), 1.0), 1.0); }
  if (gsl_colorIdx == 16u) { return vec4<f32>(vec3<f32>(0.77, 0.78, 0.78), 1.0); }
  if (gsl_colorIdx == 17u) { return vec4<f32>(thread_reliefColor(vec3<f32>(0.77, 0.78, 0.78), gsl_threadColorOn_remap(vec3<f32>(p.x - -16.0, p.y - -22.0, p.z)), 1.0), 1.0); }
  if (gsl_colorIdx == 18u) { return vec4<f32>(thread_reliefColor(vec3<f32>(0.77, 0.78, 0.78), gsl_threadColorOn_remap_4(vec3<f32>(p.x - 0.0, p.y - -22.0, p.z)), 1.0), 1.0); }
  if (gsl_colorIdx == 19u) { return vec4<f32>(thread_reliefColor(vec3<f32>(0.77, 0.78, 0.78), gsl_threadColorOn_remap_5(vec3<f32>(p.x - 16.0, p.y - -22.0, p.z)), 0.8), 1.0); }
  if (gsl_colorIdx == 20u) { return vec4<f32>(thread_reliefColor(vec3<f32>(0.77, 0.78, 0.78), gsl_threadColorOn_remap_6(vec3<f32>(p.x - 32.0, p.y - -22.0, p.z)), 0.75), 1.0); }
  if (gsl_colorIdx == 21u) { return vec4<f32>(thread_reliefColor(vec3<f32>(0.77, 0.78, 0.78), gsl_threadColorOn_remap_3(vec3<f32>(p.x - -56.0, p.y - -44.0, p.z)), 1.0), 1.0); }
  if (gsl_colorIdx == 22u) { return vec4<f32>(thread_reliefColor(vec3<f32>(0.77, 0.78, 0.78), gsl_threadColorOn_remap_3(vec3<f32>(p.x - -40.0, p.y - -44.0, p.z)), 1.0), 1.0); }
  if (gsl_colorIdx == 23u) { return vec4<f32>(thread_reliefColor(vec3<f32>(0.77, 0.78, 0.78), gsl_threadColorOn_remap_3(vec3<f32>(p.x - -24.0, p.y - -44.0, p.z)), 1.0), 1.0); }
  if (gsl_colorIdx == 24u) { return vec4<f32>(thread_reliefColor(vec3<f32>(0.77, 0.78, 0.78), gsl_threadColorOn_remap_3(vec3<f32>(p.x - -8.0, p.y - -44.0, p.z)), 1.0), 1.0); }
  if (gsl_colorIdx == 25u) { return vec4<f32>(thread_reliefColor(vec3<f32>(0.77, 0.78, 0.78), gsl_threadColorOn_remap_3(vec3<f32>(p.x - 8.0, p.y - -44.0, p.z)), 1.0), 1.0); }
  if (gsl_colorIdx == 26u) { return vec4<f32>(thread_reliefColor(vec3<f32>(0.77, 0.78, 0.78), gsl_threadColorOn_remap_3(vec3<f32>(p.x - 24.0, p.y - -44.0, p.z)), 1.0), 1.0); }
  if (gsl_colorIdx == 27u) { return vec4<f32>(thread_reliefColor(vec3<f32>(0.22, 0.22, 0.24), gsl_threadColorOn_remap_3(vec3<f32>(p.x - 40.0, p.y - -44.0, p.z)), 1.0), 1.0); }
  if (gsl_colorIdx == 28u) { return vec4<f32>(vec3<f32>(0.93, 0.91, 0.85), 1.0); }
  if (gsl_colorIdx == 29u) { return vec4<f32>(thread_reliefColor(vec3<f32>(0.8, 0.82, 0.84), gsl_threadColorOn_remap_3(vec3<f32>(p.x - 56.0, p.y - -44.0, p.z)), 1.0), 1.0); }
  if (gsl_colorIdx == 30u) { return vec4<f32>(vec3<f32>(0.77, 0.78, 0.78), 1.0); }
  return vec4<f32>(0.0, 0.0, 0.0, 0.0);
}
fn gsl_materialField(p: vec3<f32>, time: f32, gsl_materialIdx: u32) -> gsl_material {
  if (gsl_materialIdx == 1u) { let gsl_m = material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_threadColorOn_remap(vec3<f32>(p.x - -64.0, p.y - 22.0, p.z))); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 2u) { let gsl_m = material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_threadColorOn_remap(vec3<f32>(p.x - -48.0, p.y - 22.0, p.z))); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 3u) { let gsl_m = material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_threadColorOn_remap(vec3<f32>(p.x - -32.0, p.y - 22.0, p.z))); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 4u) { let gsl_m = material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_threadColorOn_remap_2(vec3<f32>(p.x - -16.0, p.y - 22.0, p.z))); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 5u) { let gsl_m = material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_threadColorOn_remap(vec3<f32>(p.x - 0.0, p.y - 22.0, p.z))); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 6u) { let gsl_m = material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_threadColorOn_remap(vec3<f32>(p.x - 16.0, p.y - 22.0, p.z))); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 7u) { let gsl_m = material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_threadColorOn_remap(vec3<f32>(p.x - 32.0, p.y - 22.0, p.z))); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 8u) { let gsl_m = material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_threadColorOn_remap(vec3<f32>(p.x - 48.0, p.y - 22.0, p.z))); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 9u) { let gsl_m = material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_threadColorOn_remap_3(vec3<f32>(p.x - 64.0, p.y - 22.0, p.z))); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 10u) { let gsl_m = material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_threadColorOn_remap(vec3<f32>(p.x - -40.0, p.y - 0.0, p.z))); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 11u) { let gsl_m = material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_threadColorOn_remap(vec3<f32>(p.x - -24.0, p.y - 0.0, p.z))); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 12u) { let gsl_m = material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_threadColorOn_remap(vec3<f32>(p.x - -8.0, p.y - 0.0, p.z))); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 13u) { let gsl_m = material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_threadColorOn_remap(vec3<f32>(p.x - 8.0, p.y - 0.0, p.z))); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 14u) { let gsl_m = material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_threadColorOn_remap(vec3<f32>(p.x - 24.0, p.y - 0.0, p.z))); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 15u) { let gsl_m = material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_threadColorOn_remap(vec3<f32>(p.x - 40.0, p.y - 0.0, p.z))); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 16u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 17u) { let gsl_m = material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_threadColorOn_remap(vec3<f32>(p.x - -16.0, p.y - -22.0, p.z))); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 18u) { let gsl_m = material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_threadColorOn_remap_4(vec3<f32>(p.x - 0.0, p.y - -22.0, p.z))); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 19u) { let gsl_m = material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_threadColorOn_remap_5(vec3<f32>(p.x - 16.0, p.y - -22.0, p.z))); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 20u) { let gsl_m = material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_threadColorOn_remap_6(vec3<f32>(p.x - 32.0, p.y - -22.0, p.z))); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 21u) { let gsl_m = material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_threadColorOn_remap_3(vec3<f32>(p.x - -56.0, p.y - -44.0, p.z))); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 22u) { let gsl_m = material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_threadColorOn_remap_3(vec3<f32>(p.x - -40.0, p.y - -44.0, p.z))); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 23u) { let gsl_m = material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_threadColorOn_remap_3(vec3<f32>(p.x - -24.0, p.y - -44.0, p.z))); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 24u) { let gsl_m = material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_threadColorOn_remap_3(vec3<f32>(p.x - -8.0, p.y - -44.0, p.z))); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 25u) { let gsl_m = material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_threadColorOn_remap_3(vec3<f32>(p.x - 8.0, p.y - -44.0, p.z))); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 26u) { let gsl_m = material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_threadColorOn_remap_3(vec3<f32>(p.x - 24.0, p.y - -44.0, p.z))); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 27u) { let gsl_m = material_withDisplacement(mat4x4<f32>(vec4<f32>(0.55, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_threadColorOn_remap_3(vec3<f32>(p.x - 40.0, p.y - -44.0, p.z))); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 28u) {
    let material_plastic = mat4x4<f32>(vec4<f32>(0.35, 0.0, 1.46, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    let gsl_m = material_plastic; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 29u) { let gsl_m = material_withDisplacement(mat4x4<f32>(vec4<f32>(0.55, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_threadColorOn_remap_3(vec3<f32>(p.x - 56.0, p.y - -44.0, p.z))); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 30u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.3, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
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

fn gsl_heights(p: vec3<f32>, time: f32) -> array<f32, 30> {
  return array<f32, 30>(material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_threadColorOn_remap(vec3<f32>(p.x - -64.0, p.y - 22.0, p.z)))[2].z, material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_threadColorOn_remap(vec3<f32>(p.x - -48.0, p.y - 22.0, p.z)))[2].z, material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_threadColorOn_remap(vec3<f32>(p.x - -32.0, p.y - 22.0, p.z)))[2].z, material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_threadColorOn_remap_2(vec3<f32>(p.x - -16.0, p.y - 22.0, p.z)))[2].z, material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_threadColorOn_remap(vec3<f32>(p.x - 0.0, p.y - 22.0, p.z)))[2].z, material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_threadColorOn_remap(vec3<f32>(p.x - 16.0, p.y - 22.0, p.z)))[2].z, material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_threadColorOn_remap(vec3<f32>(p.x - 32.0, p.y - 22.0, p.z)))[2].z, material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_threadColorOn_remap(vec3<f32>(p.x - 48.0, p.y - 22.0, p.z)))[2].z, material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_threadColorOn_remap_3(vec3<f32>(p.x - 64.0, p.y - 22.0, p.z)))[2].z, material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_threadColorOn_remap(vec3<f32>(p.x - -40.0, p.y - 0.0, p.z)))[2].z, material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_threadColorOn_remap(vec3<f32>(p.x - -24.0, p.y - 0.0, p.z)))[2].z, material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_threadColorOn_remap(vec3<f32>(p.x - -8.0, p.y - 0.0, p.z)))[2].z, material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_threadColorOn_remap(vec3<f32>(p.x - 8.0, p.y - 0.0, p.z)))[2].z, material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_threadColorOn_remap(vec3<f32>(p.x - 24.0, p.y - 0.0, p.z)))[2].z, material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_threadColorOn_remap(vec3<f32>(p.x - 40.0, p.y - 0.0, p.z)))[2].z, 0.0, material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_threadColorOn_remap(vec3<f32>(p.x - -16.0, p.y - -22.0, p.z)))[2].z, material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_threadColorOn_remap_4(vec3<f32>(p.x - 0.0, p.y - -22.0, p.z)))[2].z, material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_threadColorOn_remap_5(vec3<f32>(p.x - 16.0, p.y - -22.0, p.z)))[2].z, material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_threadColorOn_remap_6(vec3<f32>(p.x - 32.0, p.y - -22.0, p.z)))[2].z, material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_threadColorOn_remap_3(vec3<f32>(p.x - -56.0, p.y - -44.0, p.z)))[2].z, material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_threadColorOn_remap_3(vec3<f32>(p.x - -40.0, p.y - -44.0, p.z)))[2].z, material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_threadColorOn_remap_3(vec3<f32>(p.x - -24.0, p.y - -44.0, p.z)))[2].z, material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_threadColorOn_remap_3(vec3<f32>(p.x - -8.0, p.y - -44.0, p.z)))[2].z, material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_threadColorOn_remap_3(vec3<f32>(p.x - 8.0, p.y - -44.0, p.z)))[2].z, material_withDisplacement(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_threadColorOn_remap_3(vec3<f32>(p.x - 24.0, p.y - -44.0, p.z)))[2].z, material_withDisplacement(mat4x4<f32>(vec4<f32>(0.55, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_threadColorOn_remap_3(vec3<f32>(p.x - 40.0, p.y - -44.0, p.z)))[2].z, 0.0, material_withDisplacement(mat4x4<f32>(vec4<f32>(0.55, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_threadColorOn_remap_3(vec3<f32>(p.x - 56.0, p.y - -44.0, p.z)))[2].z, 0.0);
}

fn gsl_sdShapeDisplaced(p: vec3<f32>, time: f32, gsl_dispK: f32) -> gsl_shape {
  let gsl_d = gsl_shapeParts(p, time);
  let gsl_h = gsl_heights(p, time);
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
  let gsl_s16 = gsl_shape(gsl_d[15], 16u);
  let gsl_s17 = gsl_shape(gsl_d[16] - gsl_h[16] * gsl_dispK, 17u);
  let gsl_s18 = gsl_shape(gsl_d[17] - gsl_h[17] * gsl_dispK, 18u);
  let gsl_s19 = gsl_shape(gsl_d[18] - gsl_h[18] * gsl_dispK, 19u);
  let gsl_s20 = gsl_shape(gsl_d[19] - gsl_h[19] * gsl_dispK, 20u);
  let gsl_s21 = gsl_shape(gsl_d[20] - gsl_h[20] * gsl_dispK, 21u);
  let gsl_s22 = gsl_shape(gsl_d[21] - gsl_h[21] * gsl_dispK, 22u);
  let gsl_s23 = gsl_shape(gsl_d[22] - gsl_h[22] * gsl_dispK, 23u);
  let gsl_s24 = gsl_shape(gsl_d[23] - gsl_h[23] * gsl_dispK, 24u);
  let gsl_s25 = gsl_shape(gsl_d[24] - gsl_h[24] * gsl_dispK, 25u);
  let gsl_s26 = gsl_shape(gsl_d[25] - gsl_h[25] * gsl_dispK, 26u);
  let gsl_s27 = gsl_shape(gsl_d[26] - gsl_h[26] * gsl_dispK, 27u);
  let gsl_s28 = gsl_shape(gsl_d[27], 28u);
  let gsl_s29 = gsl_shape(gsl_d[28] - gsl_h[28] * gsl_dispK, 29u);
  let gsl_s30 = gsl_shape(gsl_d[29], 30u);
  return gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_s1, gsl_s2), gsl_s3), gsl_s4), gsl_s5), gsl_s6), gsl_s7), gsl_s8), gsl_s9), gsl_s10), gsl_s11), gsl_s12), gsl_s13), gsl_s14), gsl_s15), gsl_s16), gsl_s17), gsl_s18), gsl_s19), gsl_s20), gsl_s21), gsl_s22), gsl_s23), gsl_s24), gsl_s25), gsl_s26), gsl_s27), gsl_s28), gsl_s29), gsl_s30);
}
