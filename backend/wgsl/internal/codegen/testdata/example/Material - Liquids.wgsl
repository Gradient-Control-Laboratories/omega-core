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
  let water = gsl_ball_remap(p - vec3<f32>(-150.0, 112.5, 0.0));
  let vodka40 = gsl_ball_remap(p - vec3<f32>(-75.0, 112.5, 0.0));
  let ethanol = gsl_ball_remap(p - vec3<f32>(0.0, 112.5, 0.0));
  let gasoline = gsl_ball_remap(p - vec3<f32>(75.0, 112.5, 0.0));
  let cookingOil = gsl_ball_remap(p - vec3<f32>(150.0, 112.5, 0.0));
  let beerPaleLager = gsl_ball_remap(p - vec3<f32>(-150.0, 37.5, 0.0));
  let wineWhite = gsl_ball_remap(p - vec3<f32>(-75.0, 37.5, 0.0));
  let wineRed = gsl_ball_remap(p - vec3<f32>(0.0, 37.5, 0.0));
  let cola = gsl_ball_remap(p - vec3<f32>(75.0, 37.5, 0.0));
  let coffee = gsl_ball_remap(p - vec3<f32>(150.0, 37.5, 0.0));
  let juiceApple = gsl_ball_remap(p - vec3<f32>(-150.0, -37.5, 0.0));
  let juiceWhiteGrapefruit = gsl_ball_remap(p - vec3<f32>(-75.0, -37.5, 0.0));
  let juiceRubyGrapefruit = gsl_ball_remap(p - vec3<f32>(0.0, -37.5, 0.0));
  let juiceCranberry = gsl_ball_remap(p - vec3<f32>(75.0, -37.5, 0.0));
  let juiceGrape = gsl_ball_remap(p - vec3<f32>(150.0, -37.5, 0.0));
  let honeyLiquid = gsl_ball_remap(p - vec3<f32>(-150.0, -112.5, 0.0));
  let milk = gsl_ball_remap(p - vec3<f32>(-75.0, -112.5, 0.0));
  let cream = gsl_ball_remap(p - vec3<f32>(0.0, -112.5, 0.0));
  let petroleum = gsl_ball_remap(p - vec3<f32>(75.0, -112.5, 0.0));
  let bloodOxygenated = gsl_ball_remap(p - vec3<f32>(150.0, -112.5, 0.0));
  let gsl_partLo = vec3<f32>(-190.0, -150.0, -29.0);
  let gsl_partHi = vec3<f32>(190.0, 150.0, -25.0);
  let gsl_partBox = gsl_boxDistance(p, gsl_partLo, gsl_partHi);
  var gsl_part = gsl_partBox;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox <= gsl_vanishView.w) {
    let gsl_partVanish = gsl_vanish(gsl_partLo, gsl_partHi);
    gsl_part = 1e10;
    if (gsl_partVanish.x > 0.0) { gsl_part = builtin_gcl_closure_boxCentered(p, vec3<f32>(0.0, 0.0, -27.0), vec3<f32>(380.0, 300.0, 4.0)) + gsl_partVanish.y; }
    if (abs(gsl_part) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part), gsl_partVanish.zw); }
  }
  let board = gsl_part;
  let gsl_s1 = gsl_shape(water, 1u);
  let gsl_s2 = gsl_shape(vodka40, 2u);
  let gsl_s3 = gsl_shape(ethanol, 3u);
  let gsl_s4 = gsl_shape(gasoline, 4u);
  let gsl_s5 = gsl_shape(cookingOil, 5u);
  let gsl_s6 = gsl_shape(beerPaleLager, 6u);
  let gsl_s7 = gsl_shape(wineWhite, 7u);
  let gsl_s8 = gsl_shape(wineRed, 8u);
  let gsl_s9 = gsl_shape(cola, 9u);
  let gsl_s10 = gsl_shape(coffee, 10u);
  let gsl_s11 = gsl_shape(juiceApple, 11u);
  let gsl_s12 = gsl_shape(juiceWhiteGrapefruit, 12u);
  let gsl_s13 = gsl_shape(juiceRubyGrapefruit, 13u);
  let gsl_s14 = gsl_shape(juiceCranberry, 14u);
  let gsl_s15 = gsl_shape(juiceGrape, 15u);
  let gsl_s16 = gsl_shape(honeyLiquid, 16u);
  let gsl_s17 = gsl_shape(milk, 17u);
  let gsl_s18 = gsl_shape(cream, 18u);
  let gsl_s19 = gsl_shape(petroleum, 19u);
  let gsl_s20 = gsl_shape(bloodOxygenated, 20u);
  let gsl_s21 = gsl_shape(board, 21u);
  return gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_s1, gsl_s2), gsl_s3), gsl_s4), gsl_s5), gsl_s6), gsl_s7), gsl_s8), gsl_s9), gsl_s10), gsl_s11), gsl_s12), gsl_s13), gsl_s14), gsl_s15), gsl_s16), gsl_s17), gsl_s18), gsl_s19), gsl_s20), gsl_s21);
}

fn gsl_colorField(p: vec3<f32>, time: f32, gsl_colorIdx: u32) -> vec4<f32> {
  if (gsl_colorIdx == 1u) { return vec4<f32>(vec3<f32>(0.9862, 0.9982, 0.9987), 1.0); }
  if (gsl_colorIdx == 2u) { return vec4<f32>(vec3<f32>(1.0, 1.0, 1.0), 1.0); }
  if (gsl_colorIdx == 3u) { return vec4<f32>(vec3<f32>(1.0, 1.0, 1.0), 1.0); }
  if (gsl_colorIdx == 4u) { return vec4<f32>(vec3<f32>(1.0, 0.9867, 0.8077), 1.0); }
  if (gsl_colorIdx == 5u) { return vec4<f32>(vec3<f32>(0.8745, 0.8472, 0.3336), 1.0); }
  if (gsl_colorIdx == 6u) { return vec4<f32>(vec3<f32>(0.9495, 0.8937, 0.7723), 1.0); }
  if (gsl_colorIdx == 7u) { return vec4<f32>(vec3<f32>(0.9528, 0.9476, 0.8978), 1.0); }
  if (gsl_colorIdx == 8u) { return vec4<f32>(vec3<f32>(0.5926, 0.3152, 0.2552), 1.0); }
  if (gsl_colorIdx == 9u) { return vec4<f32>(vec3<f32>(0.6327, 0.4685, 0.3227), 1.0); }
  if (gsl_colorIdx == 10u) { return vec4<f32>(vec3<f32>(0.6993, 0.4002, 0.2029), 1.0); }
  if (gsl_colorIdx == 11u) { return vec4<f32>(vec3<f32>(0.9338, 0.8886, 0.7791), 1.0); }
  if (gsl_colorIdx == 12u) { return vec4<f32>(vec3<f32>(0.988, 0.9849, 0.9686), 1.0); }
  if (gsl_colorIdx == 13u) { return vec4<f32>(vec3<f32>(0.3318, 0.2029, 0.1334), 1.0); }
  if (gsl_colorIdx == 14u) { return vec4<f32>(vec3<f32>(0.8328, 0.6507, 0.5657), 1.0); }
  if (gsl_colorIdx == 15u) { return vec4<f32>(vec3<f32>(0.6246, 0.3336, 0.2552), 1.0); }
  if (gsl_colorIdx == 16u) { return vec4<f32>(vec3<f32>(0.9217, 0.7803, 0.2121), 1.0); }
  if (gsl_colorIdx == 17u) { return vec4<f32>(vec3<f32>(0.9138, 0.9128, 0.8445), 1.0); }
  if (gsl_colorIdx == 18u) { return vec4<f32>(vec3<f32>(0.9894, 0.9547, 0.8677), 1.0); }
  if (gsl_colorIdx == 19u) { return vec4<f32>(vec3<f32>(0.1897, 0.1792, 0.168), 1.0); }
  if (gsl_colorIdx == 20u) { return vec4<f32>(vec3<f32>(0.8233, 0.0388, 0.061), 1.0); }
  if (gsl_colorIdx == 21u) { return vec4<f32>(select(vec3<f32>(0.12, 0.12, 0.12), vec3<f32>(0.85, 0.85, 0.85), 1.0 - 2.0 * gsl_floored((gsl_floored((floor(p.x / 12.0) + floor(p.y / 12.0)), 2.0) + 2.0), 2.0) > 0.0), 1.0); }
  return vec4<f32>(0.0, 0.0, 0.0, 0.0);
}
fn gsl_materialField(p: vec3<f32>, time: f32, gsl_materialIdx: u32) -> gsl_material {
  if (gsl_materialIdx == 1u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.0, 0.0, 1.3325, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1000.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 2u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.0, 0.0, 1.3583, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 10.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 3u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.0, 0.0, 1.362, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 10.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 4u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.0, 0.0, 1.427, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 10.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 5u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.0, 0.0, 1.47, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 10.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 6u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.0, 0.0, 1.33, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 10.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 7u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.0, 0.0, 1.3325, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 10.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 8u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.0, 0.0, 1.3325, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 10.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 9u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.0, 0.0, 1.3325, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 10.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 10u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.0, 0.0, 1.34, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 11u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.0, 0.0, 1.3325, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 10.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 12u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.0, 0.0, 1.3325, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 10.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 13u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.0, 0.0, 1.3325, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 10.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 14u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.0, 0.0, 1.3325, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 10.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 15u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.0, 0.0, 1.3325, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 10.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 16u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.0, 0.0, 1.504, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 10.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 17u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.0, 0.0, 1.348, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 18u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.0, 0.0, 1.348, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 19u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.0, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 20u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.0, 0.0, 1.35, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 0.8, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 21u) {
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
