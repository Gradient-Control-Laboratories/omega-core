fn cylZ(p: vec3<f32>, r: f32, z0: f32, z1: f32) -> f32 {
  return max(max(length(p.xy) - r, z0 - p.z), p.z - z1);
}

fn sdShape(p: vec3<f32>, time: f32, gsl_param_13_11: f32, gsl_param_16_10: f32) -> gsl_shape {
  gsl_partNear = vec3<f32>(1e10, 1e10, 1e10);
  let gsl_partLo = vec3<f32>(-12.0, -12.0, -18.0);
  let gsl_partHi = vec3<f32>(12.0, 12.0, 6.0);
  let gsl_partBox = gsl_boxDistance(p, gsl_partLo, gsl_partHi);
  var gsl_part = gsl_partBox;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox <= gsl_vanishView.w) {
    let gsl_partVanish = gsl_vanish(gsl_partLo, gsl_partHi);
    gsl_part = 1e10;
    if (gsl_partVanish.x > 0.0) { gsl_part = max(max(max(length(p.xy) - 12.0, -18.0 - p.z), p.z - 6.0), -max(max(length(p.xy) - 8.0, -16.0 - p.z), p.z - 7.0)) + gsl_partVanish.y; }
    if (abs(gsl_part) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part), gsl_partVanish.zw); }
  }
  let block = gsl_part;
  let pistonR = gsl_param_13_11;
  let stroke = gsl_param_16_10;
  let gsl_partLo_2 = vec3<f32>(min(0.0 - max(pistonR, 0.0), -2.5), min(0.0 - max(pistonR, 0.0), -2.5), min(-(-4.0 - stroke) / -1.0, -(3.0 - stroke) / -1.0));
  let gsl_partHi_2 = vec3<f32>(max(0.0 + max(pistonR, 0.0), 2.5), max(0.0 + max(pistonR, 0.0), 2.5), max(-(-0.0 - (4.0 - stroke)), 22.0));
  let gsl_partBox_2 = gsl_boxDistance(p, gsl_partLo_2, gsl_partHi_2);
  var gsl_part_2 = gsl_partBox_2;
  let gsl_partLo_3 = vec3<f32>(0.0 - max(pistonR, 0.0), 0.0 - max(pistonR, 0.0), -(-4.0 - stroke) / -1.0);
  let gsl_partHi_3 = vec3<f32>(0.0 + max(pistonR, 0.0), 0.0 + max(pistonR, 0.0), -(-0.0 - (4.0 - stroke)));
  let gsl_partBox_3 = gsl_boxDistance(p, gsl_partLo_3, gsl_partHi_3);
  var gsl_part_3 = gsl_partBox_3;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_3 <= gsl_vanishView.w) {
    let gsl_partVanish_3 = gsl_vanish(gsl_partLo_3, gsl_partHi_3);
    gsl_part_3 = 1e10;
    if (gsl_partVanish_3.x > 0.0) { gsl_part_3 = cylZ(p, pistonR, -4.0 - stroke, 4.0 - stroke) + gsl_partVanish_3.y; }
    if (abs(gsl_part_3) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_3), gsl_partVanish_3.zw); }
  }
  let head = gsl_part_3;
  let gsl_partLo_4 = vec3<f32>(-2.5, -2.5, -(3.0 - stroke) / -1.0);
  let gsl_partHi_4 = vec3<f32>(2.5, 2.5, 22.0);
  let gsl_partBox_4 = gsl_boxDistance(p, gsl_partLo_4, gsl_partHi_4);
  var gsl_part_4 = gsl_partBox_4;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_4 <= gsl_vanishView.w) {
    let gsl_partVanish_4 = gsl_vanish(gsl_partLo_4, gsl_partHi_4);
    gsl_part_4 = 1e10;
    if (gsl_partVanish_4.x > 0.0) { gsl_part_4 = cylZ(p, 2.5, 3.0 - stroke, 22.0) + gsl_partVanish_4.y; }
    if (abs(gsl_part_4) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_4), gsl_partVanish_4.zw); }
  }
  let rod = gsl_part_4;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_2 <= gsl_vanishView.w) {
    let gsl_partVanish_2 = gsl_vanish(gsl_partLo_2, gsl_partHi_2);
    gsl_part_2 = 1e10;
    if (gsl_partVanish_2.x > 0.0) { gsl_part_2 = min(head, rod) + gsl_partVanish_2.y; }
    if (abs(gsl_part_2) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_2), gsl_partVanish_2.zw); }
  }
  let piston = gsl_part_2;
  let gsl_s1 = gsl_shape(block, 1u);
  let gsl_s2 = gsl_shape(piston, 2u);
  return gsl_min(gsl_s1, gsl_s2);
}

fn gsl_colorField(p: vec3<f32>, time: f32, gsl_colorIdx: u32, gsl_param_13_11: f32, gsl_param_16_10: f32) -> vec4<f32> {
  if (gsl_colorIdx == 1u) { return vec4<f32>(vec3<f32>(0.42, 0.48, 0.58), 1.0); }
  if (gsl_colorIdx == 2u) { return vec4<f32>(vec3<f32>(0.78, 0.62, 0.35), 1.0); }
  return vec4<f32>(0.0, 0.0, 0.0, 0.0);
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
