fn builtin_sphere(p: vec3<f32>, radius: f32) -> f32 {
  return length(p) - radius;
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  gsl_partNear = vec3<f32>(1e10, 1e10, 1e10);
  let gsl_partLo = vec3<f32>(-21.0, -3.0, -3.0);
  let gsl_partHi = vec3<f32>(-15.0, 3.0, 3.0);
  let gsl_partBox = gsl_boxDistance(p, gsl_partLo, gsl_partHi);
  var gsl_part = gsl_partBox;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox <= gsl_vanishView.w) {
    let gsl_partVanish = gsl_vanish(gsl_partLo, gsl_partHi);
    gsl_part = 1e10;
    if (gsl_partVanish.x > 0.0) { gsl_part = builtin_sphere(p - vec3<f32>(-18.0, 0.0, 0.0), 3.0) + gsl_partVanish.y; }
    if (abs(gsl_part) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part), gsl_partVanish.zw); }
  }
  let matteBall = gsl_part;
  let gsl_partLo_2 = vec3<f32>(-12.0, -3.0, -3.0);
  let gsl_partHi_2 = vec3<f32>(-6.0, 3.0, 3.0);
  let gsl_partBox_2 = gsl_boxDistance(p, gsl_partLo_2, gsl_partHi_2);
  var gsl_part_2 = gsl_partBox_2;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_2 <= gsl_vanishView.w) {
    let gsl_partVanish_2 = gsl_vanish(gsl_partLo_2, gsl_partHi_2);
    gsl_part_2 = 1e10;
    if (gsl_partVanish_2.x > 0.0) { gsl_part_2 = builtin_sphere(p - vec3<f32>(-9.0, 0.0, 0.0), 3.0) + gsl_partVanish_2.y; }
    if (abs(gsl_part_2) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_2), gsl_partVanish_2.zw); }
  }
  let plasticBall = gsl_part_2;
  let gsl_partLo_3 = vec3<f32>(-3.0, -3.0, -3.0);
  let gsl_partHi_3 = vec3<f32>(3.0, 3.0, 3.0);
  let gsl_partBox_3 = gsl_boxDistance(p, gsl_partLo_3, gsl_partHi_3);
  var gsl_part_3 = gsl_partBox_3;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_3 <= gsl_vanishView.w) {
    let gsl_partVanish_3 = gsl_vanish(gsl_partLo_3, gsl_partHi_3);
    gsl_part_3 = 1e10;
    if (gsl_partVanish_3.x > 0.0) { gsl_part_3 = builtin_sphere(p - vec3<f32>(0.0, 0.0, 0.0), 3.0) + gsl_partVanish_3.y; }
    if (abs(gsl_part_3) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_3), gsl_partVanish_3.zw); }
  }
  let brassBall = gsl_part_3;
  let gsl_partLo_4 = vec3<f32>(6.0, -3.0, -3.0);
  let gsl_partHi_4 = vec3<f32>(12.0, 3.0, 3.0);
  let gsl_partBox_4 = gsl_boxDistance(p, gsl_partLo_4, gsl_partHi_4);
  var gsl_part_4 = gsl_partBox_4;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_4 <= gsl_vanishView.w) {
    let gsl_partVanish_4 = gsl_vanish(gsl_partLo_4, gsl_partHi_4);
    gsl_part_4 = 1e10;
    if (gsl_partVanish_4.x > 0.0) { gsl_part_4 = builtin_sphere(p - vec3<f32>(9.0, 0.0, 0.0), 3.0) + gsl_partVanish_4.y; }
    if (abs(gsl_part_4) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_4), gsl_partVanish_4.zw); }
  }
  let steelBall = gsl_part_4;
  let gsl_partLo_5 = vec3<f32>(15.0, -3.0, -3.0);
  let gsl_partHi_5 = vec3<f32>(21.0, 3.0, 3.0);
  let gsl_partBox_5 = gsl_boxDistance(p, gsl_partLo_5, gsl_partHi_5);
  var gsl_part_5 = gsl_partBox_5;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_5 <= gsl_vanishView.w) {
    let gsl_partVanish_5 = gsl_vanish(gsl_partLo_5, gsl_partHi_5);
    gsl_part_5 = 1e10;
    if (gsl_partVanish_5.x > 0.0) { gsl_part_5 = builtin_sphere(p - vec3<f32>(18.0, 0.0, 0.0), 3.0) + gsl_partVanish_5.y; }
    if (abs(gsl_part_5) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_5), gsl_partVanish_5.zw); }
  }
  let lampBall = gsl_part_5;
  let gsl_s1 = gsl_shape(matteBall, 1u);
  let gsl_s2 = gsl_shape(plasticBall, 2u);
  let gsl_s3 = gsl_shape(brassBall, 3u);
  let gsl_s4 = gsl_shape(steelBall, 4u);
  let gsl_s5 = gsl_shape(lampBall, 5u);
  return gsl_min(gsl_min(gsl_min(gsl_min(gsl_s1, gsl_s2), gsl_s3), gsl_s4), gsl_s5);
}

fn gsl_colorField(p: vec3<f32>, time: f32, gsl_colorIdx: u32) -> vec4<f32> {
  if (gsl_colorIdx == 1u) {
    let color_cherry = vec3<f32>(0.871, 0.192, 0.388);
    return vec4<f32>(color_cherry, 1.0);
  }
  if (gsl_colorIdx == 2u) {
    let color_teal = vec3<f32>(0.0, 0.5, 0.5);
    return vec4<f32>(color_teal, 1.0);
  }
  if (gsl_colorIdx == 3u) { return vec4<f32>(vec3<f32>(0.9593, 0.8952, 0.6822), 1.0); }
  if (gsl_colorIdx == 4u) { return vec4<f32>(vec3<f32>(0.77, 0.78, 0.78), 1.0); }
  if (gsl_colorIdx == 5u) {
    let color_orange = vec3<f32>(1.0, 0.5, 0.0);
    return vec4<f32>(color_orange, 1.0);
  }
  return vec4<f32>(0.0, 0.0, 0.0, 0.0);
}
fn gsl_materialField(p: vec3<f32>, time: f32, gsl_materialIdx: u32) -> gsl_material {
  if (gsl_materialIdx == 1u) {
    let material_matte = mat4x4<f32>(vec4<f32>(1.0, 0.0, 1.45, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    let gsl_m = material_matte; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 2u) {
    let material_shinyPlastic = mat4x4<f32>(vec4<f32>(0.1, 0.0, 1.46, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    let gsl_m = material_shinyPlastic; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 3u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.3, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 4u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.15, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 5u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.5, 0.0, 1.5, 0.0), vec4<f32>(1.0, 0.5, 0.0, 3.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
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
