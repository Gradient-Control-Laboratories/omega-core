fn gsl_floored(a: f32, n: f32) -> f32 {
  return a - n * floor(a / n);
}

fn builtin_cone(p: vec3<f32>, origin: vec3<f32>, direction: vec3<f32>, angle: f32) -> f32 {
  let offset = p - origin;
  let unitDirection = normalize(direction);
  let along = dot(offset, unitDirection);
  let perp = length(offset - unitDirection * along);
  return perp * cos(angle) - along * sin(angle);
}

fn builtin_cylinder(p: vec3<f32>, center: vec3<f32>, direction: vec3<f32>, radius: f32) -> f32 {
  return builtin_axis(p, center, direction) - radius;
}

fn builtin_cylinderZ(p: vec3<f32>, vector_dirZ: vec3<f32>, center: vec3<f32>, radius: f32) -> f32 {
  return builtin_cylinder(p, center, vector_dirZ, radius);
}

fn gsl_remap(p: vec3<f32>, gsl_param_18_11: f32, vector_dirZ: vec3<f32>, vector_zero: vec3<f32>) -> f32 {
  let holeDia = gsl_param_18_11;
  return max(builtin_gcl_closure_boxCentered(p, vec3<f32>(0.0, 0.0, -7.0), vec3<f32>(24.0, 24.0, 14.0)), -min(min(min(min(min(100000.0, max(builtin_cone(p, vec3<f32>(0.0, 0.0, -8.6), vector_dirZ, 0.7853982), builtin_axis(p, vector_zero, vector_dirZ) - 8.6)), max(builtin_axis(p, vector_zero, vector_dirZ) - 8.0, -5.0 - p.z)), max(max(builtin_axis(p, vector_zero, vector_dirZ) - 8.7, -5.0 - p.z), p.z - -4.0)), max(max(builtin_cone(p, vec3<f32>(0.0, 0.0, -5.0 - (holeDia + 1.0) / 2.0 / 1.0), vector_dirZ, 0.7853982), builtin_cylinderZ(p, vector_dirZ, vector_zero, (holeDia + 1.0) / 2.0)), p.z - -5.0)), builtin_cylinderZ(p, vector_dirZ, vector_zero, holeDia / 2.0)));
}

fn gsl_remap_2(p: vec3<f32>, gsl_param_18_11: f32, vector_dirZ: vec3<f32>, vector_zero: vec3<f32>) -> f32 {
  let holeDia = gsl_param_18_11;
  return max(builtin_gcl_closure_boxCentered(p, vec3<f32>(0.0, 0.0, -7.0), vec3<f32>(24.0, 24.0, 14.0)), -min(min(100000.0, max(builtin_cone(p, vec3<f32>(0.0, 0.0, -8.5), vector_dirZ, 0.7853982), builtin_axis(p, vector_zero, vector_dirZ) - 8.5)), builtin_cylinderZ(p, vector_dirZ, vector_zero, holeDia / 2.0)));
}

fn gsl_remap_3(p: vec3<f32>, gsl_param_18_11: f32, vector_dirZ: vec3<f32>, vector_zero: vec3<f32>) -> f32 {
  let tappedTile = max(builtin_gcl_closure_boxCentered(p, vec3<f32>(0.0, 0.0, -7.0), vec3<f32>(24.0, 24.0, 14.0)), -min(min(min(100000.0, max(builtin_cone(p, vec3<f32>(0.0, 0.0, -5.2), vector_dirZ, 0.7853982), builtin_axis(p, vector_zero, vector_dirZ) - 5.2)), max(builtin_axis(p, vector_zero, vector_dirZ) - 3.3234177, -10.0 - p.z)), max(builtin_axis(p, vector_zero, vector_dirZ) - 3.4, builtin_cone(p, vec3<f32>(0.0, 0.0, -14.042926), vector_dirZ, 1.0297443))));
  return tappedTile;
}

fn gsl_remap_4(p: vec3<f32>, gsl_param_18_11: f32, vector_dirZ: vec3<f32>, vector_zero: vec3<f32>) -> f32 {
  let holeDia = gsl_param_18_11;
  return max(builtin_gcl_closure_boxCentered(p, vec3<f32>(0.0, 0.0, -7.0), vec3<f32>(24.0, 24.0, 14.0)), -min(min(min(100000.0, max(builtin_cone(p, vec3<f32>(0.0, 0.0, -7.5), vector_dirZ, 0.7853982), builtin_axis(p, vector_zero, vector_dirZ) - 7.5)), drill_prismVoid(p, 0.0, max(max(abs(p.x) - 6.0, abs(0.5 * p.x + 0.8660254 * p.y) - 6.0), abs(0.5 * p.x - 0.8660254 * p.y) - 6.0), 6.0)), builtin_cylinderZ(p, vector_dirZ, vector_zero, holeDia / 2.0)));
}

fn gsl_remap_5(p: vec3<f32>, gsl_param_18_11: f32, vector_dirZ: vec3<f32>, vector_zero: vec3<f32>) -> f32 {
  return max(builtin_gcl_closure_boxCentered(p, vec3<f32>(0.0, 0.0, -7.0), vec3<f32>(24.0, 24.0, 14.0)), -min(min(100000.0, max(builtin_cone(p, vec3<f32>(0.0, 0.0, -7.5), vector_dirZ, 0.7853982), builtin_axis(p, vector_zero, vector_dirZ) - 7.5)), max(builtin_axis(p, vector_zero, vector_dirZ) - 6.0, p.x - 3.6)));
}

fn gsl_remap_6(p: vec3<f32>, gsl_param_18_11: f32, vector_dirZ: vec3<f32>, vector_zero: vec3<f32>) -> f32 {
  let holeDia = gsl_param_18_11;
  return max(abs(p.y) - 1.3, length(p.xy) - (holeDia + 3.6) / 2.0);
}

fn gsl_remap_7(p: vec3<f32>, gsl_param_18_11: f32, vector_dirZ: vec3<f32>, vector_zero: vec3<f32>) -> f32 {
  let holeDia = gsl_param_18_11;
  let gsl_tmp_44 = length(p.xy);
  let gsl_tmp_45 = gsl_floored((atan2(p.y, p.x) + 0.5235988), 1.0471976) - 0.5235988;
  return max(builtin_gcl_closure_boxCentered(p, vec3<f32>(0.0, 0.0, -7.0), vec3<f32>(24.0, 24.0, 14.0)), -min(min(min(100000.0, max(builtin_cone(p, vec3<f32>(0.0, 0.0, 0.0 - (holeDia + 3.0 + 1.0) / 2.0 / 1.0), vector_dirZ, 0.7853982), builtin_cylinderZ(p, vector_dirZ, vector_zero, (holeDia + 3.0 + 1.0) / 2.0))), builtin_cylinderZ(p, vector_dirZ, vector_zero, holeDia / 2.0)), gsl_remap_6(vec3<f32>(gsl_tmp_44 * cos(gsl_tmp_45), gsl_tmp_44 * sin(gsl_tmp_45), p.z), gsl_param_18_11, vector_dirZ, vector_zero)));
}

fn gsl_remap_8(p: vec3<f32>, gsl_param_18_11: f32, vector_dirZ: vec3<f32>, vector_zero: vec3<f32>) -> f32 {
  let holeDia = gsl_param_18_11;
  return max(builtin_gcl_closure_boxCentered(p, vec3<f32>(0.0, 0.0, -7.0), vec3<f32>(24.0, 24.0, 14.0)), -min(min(min(100000.0, max(builtin_cone(p, vec3<f32>(0.0, 0.0, 0.0 - (holeDia + 3.0 + 1.0) / 2.0 / 1.0), vector_dirZ, 0.7853982), builtin_cylinderZ(p, vector_dirZ, vector_zero, (holeDia + 3.0 + 1.0) / 2.0))), builtin_cylinderZ(p, vector_dirZ, vector_zero, holeDia / 2.0)), max(max(abs(p.y) - 1.8, length(p.xy) - (holeDia / 2.0 + 1.8)), -p.x)));
}

fn gsl_remap_9(p: vec3<f32>, gsl_param_18_11: f32, vector_dirZ: vec3<f32>, vector_zero: vec3<f32>) -> f32 {
  let holeDia = gsl_param_18_11;
  return max(builtin_gcl_closure_boxCentered(p, vec3<f32>(0.0, 0.0, -7.0), vec3<f32>(24.0, 24.0, 14.0)), -min(min(100000.0, drill_chamferSection(p, 0.0, length(vec2<f32>(max(abs(p.x * 1.0 + p.y * 0.0) - 5.0, 0.0), -p.x * 0.0 + p.y * 1.0)) - holeDia / 2.0, 0.4)), length(vec2<f32>(max(abs(p.x * 1.0 + p.y * 0.0) - 5.0, 0.0), -p.x * 0.0 + p.y * 1.0)) - holeDia / 2.0));
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

fn gsl_remap_10(p: vec3<f32>, gsl_param_18_11: f32, vector_dirZ: vec3<f32>, vector_zero: vec3<f32>) -> vec3<f32> {
  var gsl_simpleValue = vec3<f32>(0.48, 0.52, 0.58);
  if (gsl_simpleDetail > 0.0) {
    let gsl_tmp_37 = 6.2831855 * p.z / 1.25;
    let gsl_tmp_36 = max(length(p.xy), 0.2);
    gsl_simpleValue = mix(gsl_simpleValue, color_mixOklab(vec3<f32>(0.48, 0.52, 0.58), vec3<f32>(0.168, 0.182, 0.203), clamp(-(-min(max(0.6765824 - clamp((acos(clamp(cos(gsl_tmp_37) * p.x / gsl_tmp_36 + sin(gsl_tmp_37) * p.y / gsl_tmp_36, -1.0, 1.0)) / 3.1415927 * 0.625 - 0.078125) * 1.7320508, 0.0, 0.6765824), 0.0), max(0.0, 0.5 * min(min(0.0 - p.z, p.z - -10.0), 5.3531647 - length(p.xy))))) / 0.6765824, 0.0, 1.0) * (1.0 - gsl_dispDrawn)), gsl_simpleDetail);
  }
  return gsl_simpleValue;
}

fn gsl_remap_11(p: vec3<f32>, gsl_param_18_11: f32, vector_dirZ: vec3<f32>, vector_zero: vec3<f32>, material_standard: mat4x4<f32>) -> vec4<f32> {
  let gsl_tmp_39 = 6.2831855 * p.z / 1.25;
  let gsl_tmp_38 = max(length(p.xy), 0.2);
  return material_withDisplacement(material_standard, -min(max(0.6765824 - clamp((acos(clamp(cos(gsl_tmp_39) * p.x / gsl_tmp_38 + sin(gsl_tmp_39) * p.y / gsl_tmp_38, -1.0, 1.0)) / 3.1415927 * 0.625 - 0.078125) * 1.7320508, 0.0, 0.6765824), 0.0), max(0.0, 0.5 * min(min(0.0 - p.z, p.z - -10.0), 5.3531647 - length(p.xy)))))[0];
}

fn gsl_remap_12(p: vec3<f32>, gsl_param_18_11: f32, vector_dirZ: vec3<f32>, vector_zero: vec3<f32>, material_standard: mat4x4<f32>) -> vec4<f32> {
  let gsl_tmp_41 = 6.2831855 * p.z / 1.25;
  let gsl_tmp_40 = max(length(p.xy), 0.2);
  return material_withDisplacement(material_standard, -min(max(0.6765824 - clamp((acos(clamp(cos(gsl_tmp_41) * p.x / gsl_tmp_40 + sin(gsl_tmp_41) * p.y / gsl_tmp_40, -1.0, 1.0)) / 3.1415927 * 0.625 - 0.078125) * 1.7320508, 0.0, 0.6765824), 0.0), max(0.0, 0.5 * min(min(0.0 - p.z, p.z - -10.0), 5.3531647 - length(p.xy)))))[1];
}

fn gsl_remap_13(p: vec3<f32>, gsl_param_18_11: f32, vector_dirZ: vec3<f32>, vector_zero: vec3<f32>, material_standard: mat4x4<f32>) -> vec4<f32> {
  let gsl_tmp_43 = 6.2831855 * p.z / 1.25;
  let gsl_tmp_42 = max(length(p.xy), 0.2);
  return material_withDisplacement(material_standard, -min(max(0.6765824 - clamp((acos(clamp(cos(gsl_tmp_43) * p.x / gsl_tmp_42 + sin(gsl_tmp_43) * p.y / gsl_tmp_42, -1.0, 1.0)) / 3.1415927 * 0.625 - 0.078125) * 1.7320508, 0.0, 0.6765824), 0.0), max(0.0, 0.5 * min(min(0.0 - p.z, p.z - -10.0), 5.3531647 - length(p.xy)))))[2];
}

fn builtin_intersectSharp(fields: vec3<f32>) -> f32 {
  return max(max(fields.x, fields.y), fields.z);
}

fn builtin_intersectRound(fields: vec3<f32>, radius: f32) -> f32 {
  let sharp = builtin_intersectSharp(fields);
  let sumOfSquares = 0.0 + max(fields.x + radius, 0.0) * max(fields.x + radius, 0.0) + max(fields.y + radius, 0.0) * max(fields.y + radius, 0.0) + max(fields.z + radius, 0.0) * max(fields.z + radius, 0.0);
  return sharp - max(sharp + radius, 0.0) + sqrt(max(sumOfSquares, 0.000000000000000000000000000001));
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

fn builtin_axis(p: vec3<f32>, center: vec3<f32>, direction: vec3<f32>) -> f32 {
  let offset = p - center;
  let unitDirection = normalize(direction);
  let along = dot(offset, unitDirection);
  let perp = offset - unitDirection * along;
  return length(perp);
}

fn material_withDisplacement(m: mat4x4<f32>, height: f32) -> mat4x4<f32> {
  let volume = m[2];
  return mat4x4<f32>(m[0], m[1], vec4<f32>(volume.x, volume.y, height, volume.w), m[3]);
}

fn drill_prismVoid(p: vec3<f32>, zTop: f32, section: f32, depth: f32) -> f32 {
  return max(section, zTop - depth - p.z);
}

fn drill_chamferSection(p: vec3<f32>, plane: f32, section: f32, leg: f32) -> f32 {
  return max((section - leg - (p.z - plane)) * 0.7071, section - leg);
}

fn sdShape(p: vec3<f32>, time: f32, gsl_param_18_11: f32) -> gsl_shape {
  let vector_dirZ = vec3<f32>(0.0, 0.0, 1.0);
  let vector_zero = vec3<f32>(0.0, 0.0, 0.0);
  let counterbored = gsl_remap(vec3<f32>(p.x + 39.0, p.y - 13.0, p.z), gsl_param_18_11, vector_dirZ, vector_zero);
  let countersunk = gsl_remap_2(vec3<f32>(p.x + 13.0, p.y - 13.0, p.z), gsl_param_18_11, vector_dirZ, vector_zero);
  let tapped = gsl_remap_3(vec3<f32>(p.x - 13.0, p.y - 13.0, p.z), gsl_param_18_11, vector_dirZ, vector_zero);
  let hexed = gsl_remap_4(vec3<f32>(p.x - 39.0, p.y - 13.0, p.z), gsl_param_18_11, vector_dirZ, vector_zero);
  let dFlatted = gsl_remap_5(vec3<f32>(p.x + 39.0, p.y + 13.0, p.z), gsl_param_18_11, vector_dirZ, vector_zero);
  let splined = gsl_remap_7(vec3<f32>(p.x + 13.0, p.y + 13.0, p.z), gsl_param_18_11, vector_dirZ, vector_zero);
  let keyed = gsl_remap_8(vec3<f32>(p.x - 13.0, p.y + 13.0, p.z), gsl_param_18_11, vector_dirZ, vector_zero);
  let slotted = gsl_remap_9(vec3<f32>(p.x - 39.0, p.y + 13.0, p.z), gsl_param_18_11, vector_dirZ, vector_zero);
  let gsl_s1 = gsl_shape(counterbored, 1u);
  let gsl_s2 = gsl_shape(countersunk, 2u);
  let gsl_s3 = gsl_shape(tapped, 3u);
  let gsl_s4 = gsl_shape(hexed, 4u);
  let gsl_s5 = gsl_shape(dFlatted, 5u);
  let gsl_s6 = gsl_shape(splined, 6u);
  let gsl_s7 = gsl_shape(keyed, 7u);
  let gsl_s8 = gsl_shape(slotted, 8u);
  return gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_s1, gsl_s2), gsl_s3), gsl_s4), gsl_s5), gsl_s6), gsl_s7), gsl_s8);
}

fn gsl_colorField(p: vec3<f32>, time: f32, gsl_colorIdx: u32, gsl_param_18_11: f32) -> vec4<f32> {
  if (gsl_colorIdx == 1u) { return vec4<f32>(vec3<f32>(0.62, 0.64, 0.68), 1.0); }
  if (gsl_colorIdx == 2u) { return vec4<f32>(vec3<f32>(0.55, 0.58, 0.64), 1.0); }
  if (gsl_colorIdx == 3u) {
    let vector_dirZ = vec3<f32>(0.0, 0.0, 1.0);
    let vector_zero = vec3<f32>(0.0, 0.0, 0.0);
    return vec4<f32>(clamp(gsl_remap_10(vec3<f32>(p.x - 13.0, p.y - 13.0, p.z), gsl_param_18_11, vector_dirZ, vector_zero) * 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0)), 1.0);
  }
  if (gsl_colorIdx == 4u) { return vec4<f32>(vec3<f32>(0.42, 0.46, 0.54), 1.0); }
  if (gsl_colorIdx == 5u) { return vec4<f32>(vec3<f32>(0.66, 0.62, 0.56), 1.0); }
  if (gsl_colorIdx == 6u) { return vec4<f32>(vec3<f32>(0.6, 0.56, 0.5), 1.0); }
  if (gsl_colorIdx == 7u) { return vec4<f32>(vec3<f32>(0.52, 0.5, 0.46), 1.0); }
  if (gsl_colorIdx == 8u) { return vec4<f32>(vec3<f32>(0.7, 0.68, 0.64), 1.0); }
  return vec4<f32>(0.0, 0.0, 0.0, 0.0);
}
fn gsl_materialField(p: vec3<f32>, time: f32, gsl_materialIdx: u32, gsl_param_18_11: f32) -> gsl_material {
  if (gsl_materialIdx == 3u) {
    let vector_dirZ = vec3<f32>(0.0, 0.0, 1.0);
    let vector_zero = vec3<f32>(0.0, 0.0, 0.0);
    let material_standard = mat4x4<f32>(vec4<f32>(0.5, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    let gsl_m = mat4x4<f32>(clamp(gsl_remap_11(vec3<f32>(p.x - 13.0, p.y - 13.0, p.z), gsl_param_18_11, vector_dirZ, vector_zero, material_standard) * 1.0, vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 10000.0, 10000.0)), clamp(gsl_remap_12(vec3<f32>(p.x - 13.0, p.y - 13.0, p.z), gsl_param_18_11, vector_dirZ, vector_zero, material_standard) * 1.0, vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 10000.0)), clamp(gsl_remap_13(vec3<f32>(p.x - 13.0, p.y - 13.0, p.z), gsl_param_18_11, vector_dirZ, vector_zero, material_standard) * 1.0, vec4<f32>(0.0, 0.0, -10000.0, 0.0), vec4<f32>(1.0, 10000.0, 10000.0, 10000.0)), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  return gsl_material(vec4<f32>(0.5, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), 0.0);
}
var<private> gsl_simpleDetail: f32 = 1.0;

var<private> gsl_dispDrawn: f32 = 0.0;
