fn gsl_floored(a: f32, n: f32) -> f32 {
  return a - n * floor(a / n);
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

fn gsl_ball_remap(p: vec3<f32>) -> f32 {
  return length(p) - 25.0;
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  gsl_partNear = vec3<f32>(1e10, 1e10, 1e10);
  let gsl_partLo = vec3<f32>(-320.0, -85.0, -29.0);
  let gsl_partHi = vec3<f32>(320.0, 85.0, -25.0);
  let gsl_partBox = gsl_boxDistance(p, gsl_partLo, gsl_partHi);
  var gsl_part = gsl_partBox;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox <= gsl_vanishView.w) {
    let gsl_partVanish = gsl_vanish(gsl_partLo, gsl_partHi);
    gsl_part = 1e10;
    if (gsl_partVanish.x > 0.0) { gsl_part = builtin_gcl_closure_boxCentered(p, vec3<f32>(0.0, 0.0, -27.0), vec3<f32>(640.0, 170.0, 4.0)) + gsl_partVanish.y; }
    if (abs(gsl_part) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part), gsl_partVanish.zw); }
  }
  let board = gsl_part;
  let plasticPLA = gsl_ball_remap(p - vec3<f32>(-262.5, 37.5, 0.0));
  let plasticPET = gsl_ball_remap(p - vec3<f32>(-187.5, 37.5, 0.0));
  let plasticPC = gsl_ball_remap(p - vec3<f32>(-112.5, 37.5, 0.0));
  let plasticPP = gsl_ball_remap(p - vec3<f32>(-37.5, 37.5, 0.0));
  let plasticPVC = gsl_ball_remap(p - vec3<f32>(37.5, 37.5, 0.0));
  let plasticPUR = gsl_ball_remap(p - vec3<f32>(112.5, 37.5, 0.0));
  let plasticAcrylic = gsl_ball_remap(p - vec3<f32>(187.5, 37.5, 0.0));
  let polystyreneFoam = gsl_ball_remap(p - vec3<f32>(262.5, 37.5, 0.0));
  let diamond = gsl_ball_remap(p - vec3<f32>(-262.5, -37.5, 0.0));
  let sapphire = gsl_ball_remap(p - vec3<f32>(-187.5, -37.5, 0.0));
  let quartz = gsl_ball_remap(p - vec3<f32>(-112.5, -37.5, 0.0));
  let glassBorosilicate = gsl_ball_remap(p - vec3<f32>(-37.5, -37.5, 0.0));
  let ice = gsl_ball_remap(p - vec3<f32>(37.5, -37.5, 0.0));
  let salt = gsl_ball_remap(p - vec3<f32>(112.5, -37.5, 0.0));
  let amber = gsl_ball_remap(p - vec3<f32>(187.5, -37.5, 0.0));
  let pearl = gsl_ball_remap(p - vec3<f32>(262.5, -37.5, 0.0));
  let gsl_s1 = gsl_shape(board, 1u);
  let gsl_s2 = gsl_shape(plasticPLA, 2u);
  let gsl_s3 = gsl_shape(plasticPET, 3u);
  let gsl_s4 = gsl_shape(plasticPC, 4u);
  let gsl_s5 = gsl_shape(plasticPP, 5u);
  let gsl_s6 = gsl_shape(plasticPVC, 6u);
  let gsl_s7 = gsl_shape(plasticPUR, 7u);
  let gsl_s8 = gsl_shape(plasticAcrylic, 8u);
  let gsl_s9 = gsl_shape(polystyreneFoam, 9u);
  let gsl_s10 = gsl_shape(diamond, 10u);
  let gsl_s11 = gsl_shape(sapphire, 11u);
  let gsl_s12 = gsl_shape(quartz, 12u);
  let gsl_s13 = gsl_shape(glassBorosilicate, 13u);
  let gsl_s14 = gsl_shape(ice, 14u);
  let gsl_s15 = gsl_shape(salt, 15u);
  let gsl_s16 = gsl_shape(amber, 16u);
  let gsl_s17 = gsl_shape(pearl, 17u);
  return gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_s1, gsl_s2), gsl_s3), gsl_s4), gsl_s5), gsl_s6), gsl_s7), gsl_s8), gsl_s9), gsl_s10), gsl_s11), gsl_s12), gsl_s13), gsl_s14), gsl_s15), gsl_s16), gsl_s17);
}

fn gsl_colorField(p: vec3<f32>, time: f32, gsl_colorIdx: u32) -> vec4<f32> {
  if (gsl_colorIdx == 1u) { return vec4<f32>(select(vec3<f32>(0.12, 0.12, 0.12), vec3<f32>(0.85, 0.85, 0.85), 1.0 - 2.0 * gsl_floored((gsl_floored((floor(p.x / 12.0) + floor(p.y / 12.0)), 2.0) + 2.0), 2.0) > 0.0), 1.0); }
  if (gsl_colorIdx == 2u) { return vec4<f32>(vec3<f32>(1.0, 1.0, 1.0), 1.0); }
  if (gsl_colorIdx == 3u) { return vec4<f32>(vec3<f32>(1.0, 1.0, 1.0), 1.0); }
  if (gsl_colorIdx == 4u) { return vec4<f32>(vec3<f32>(1.0, 1.0, 1.0), 1.0); }
  if (gsl_colorIdx == 5u) { return vec4<f32>(vec3<f32>(1.0, 1.0, 1.0), 1.0); }
  if (gsl_colorIdx == 6u) { return vec4<f32>(vec3<f32>(1.0, 1.0, 1.0), 1.0); }
  if (gsl_colorIdx == 7u) { return vec4<f32>(vec3<f32>(1.0, 1.0, 1.0), 1.0); }
  if (gsl_colorIdx == 8u) { return vec4<f32>(vec3<f32>(1.0, 1.0, 1.0), 1.0); }
  if (gsl_colorIdx == 9u) { return vec4<f32>(vec3<f32>(0.9257, 0.9253, 0.9266), 1.0); }
  if (gsl_colorIdx == 10u) { return vec4<f32>(vec3<f32>(1.0, 1.0, 1.0), 1.0); }
  if (gsl_colorIdx == 11u) { return vec4<f32>(vec3<f32>(0.8379, 0.8881, 0.9333), 1.0); }
  if (gsl_colorIdx == 12u) { return vec4<f32>(vec3<f32>(1.0, 1.0, 1.0), 1.0); }
  if (gsl_colorIdx == 13u) { return vec4<f32>(vec3<f32>(0.9947, 0.9965, 0.9934), 1.0); }
  if (gsl_colorIdx == 14u) { return vec4<f32>(vec3<f32>(0.988, 0.9978, 1.0), 1.0); }
  if (gsl_colorIdx == 15u) { return vec4<f32>(vec3<f32>(1.0, 1.0, 1.0), 1.0); }
  if (gsl_colorIdx == 16u) { return vec4<f32>(vec3<f32>(0.9212, 0.5731, 0.2091), 1.0); }
  if (gsl_colorIdx == 17u) { return vec4<f32>(vec3<f32>(0.9063, 0.8808, 0.8543), 1.0); }
  return vec4<f32>(0.0, 0.0, 0.0, 0.0);
}
fn gsl_materialField(p: vec3<f32>, time: f32, gsl_materialIdx: u32) -> gsl_material {
  if (gsl_materialIdx == 1u) {
    let material_matte = mat4x4<f32>(vec4<f32>(1.0, 0.0, 1.45, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    let gsl_m = material_matte; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 2u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.0, 0.0, 1.4586, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 10.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 3u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.0, 0.0, 1.575, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 10.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 4u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.0, 0.0, 1.5848, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 10.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 5u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.0, 0.0, 1.492, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 10.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 6u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.0, 0.0, 1.542, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 10.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 7u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.0, 0.0, 1.6, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 10.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 8u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.0, 0.0, 1.4905, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 10.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 9u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.5, 0.0, 1.6, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 10u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.0, 0.0, 2.4168, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 10.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 11u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.0, 0.0, 1.7682, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 10.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 12u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.0, 0.0, 1.54, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 10.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 13u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.0, 0.0, 1.52, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 500.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 14u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.0, 0.0, 1.3098, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1000.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 15u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.2, 0.0, 1.5441, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 10.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 16u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.0, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 10.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 17u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.35, 0.0, 1.68, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
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
