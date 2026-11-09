fn sdShape(p: vec3<f32>, time: f32, gsl_param_18_10: f32) -> gsl_shape {
  gsl_partNear = vec3<f32>(1e10, 1e10, 1e10);
  let crankR = gsl_param_18_10;
  let gsl_partLo = vec3<f32>(0.0 - max(crankR + 0.4, 0.0), 0.0 - max(crankR + 0.4, 0.0), -0.45);
  let gsl_partHi = vec3<f32>(0.0 + max(crankR + 0.4, 0.0), 0.0 + max(crankR + 0.4, 0.0), 0.45);
  let gsl_partBox = gsl_boxDistance(p, gsl_partLo, gsl_partHi);
  var gsl_part = gsl_partBox;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox <= gsl_vanishView.w) {
    let gsl_partVanish = gsl_vanish(gsl_partLo, gsl_partHi);
    gsl_part = 1e10;
    if (gsl_partVanish.x > 0.0) { gsl_part = max(length(p.xy) - (crankR + 0.4), abs(p.z) - 0.45) + gsl_partVanish.y; }
    if (abs(gsl_part) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part), gsl_partVanish.zw); }
  }
  let crank = gsl_part;
  let theta = 6.2831855 * time;
  let pinX = crankR * cos(theta);
  let pinY = crankR * sin(theta);
  let gsl_partLo_2 = vec3<f32>(pinX - 0.55, pinY - 0.55, -0.8);
  let gsl_partHi_2 = vec3<f32>(pinX + 0.55, pinY + 0.55, 0.8);
  let gsl_partBox_2 = gsl_boxDistance(p, gsl_partLo_2, gsl_partHi_2);
  var gsl_part_2 = gsl_partBox_2;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_2 <= gsl_vanishView.w) {
    let gsl_partVanish_2 = gsl_vanish(gsl_partLo_2, gsl_partHi_2);
    gsl_part_2 = 1e10;
    if (gsl_partVanish_2.x > 0.0) { gsl_part_2 = max(length(p.xy - vec2<f32>(pinX, pinY)) - 0.55, abs(p.z) - 0.8) + gsl_partVanish_2.y; }
    if (abs(gsl_part_2) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_2), gsl_partVanish_2.zw); }
  }
  let pin = gsl_part_2;
  let qX = p.x - pinX;
  let slideX = pinX + sqrt(64.0 - pinY * pinY);
  let dX = slideX - pinX;
  let qY = p.y - pinY;
  let dY = 0.0 - pinY;
  let h = clamp((qX * dX + qY * dY) / (dX * dX + dY * dY), 0.0, 1.0);
  let rod = length(vec3<f32>(qX - dX * h, qY - dY * h, p.z)) - 0.45;
  let piston = max(length(p.yz) - 1.35, abs(p.x - slideX - 1.2) - 1.5);
  let gsl_partLo_3 = vec3<f32>(3.82, -6.0, -6.0);
  let gsl_partHi_3 = vec3<f32>(4.18, 6.0, 6.0);
  let gsl_partBox_3 = gsl_boxDistance(p, gsl_partLo_3, gsl_partHi_3);
  var gsl_part_3 = gsl_partBox_3;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_3 <= gsl_vanishView.w) {
    let gsl_partVanish_3 = gsl_vanish(gsl_partLo_3, gsl_partHi_3);
    gsl_part_3 = 1e10;
    if (gsl_partVanish_3.x > 0.0) { gsl_part_3 = max(max(abs(p.x - 4.0) - 0.18, 2.9 - length(p.yz)), length(p.yz) - 6.0) + gsl_partVanish_3.y; }
    if (abs(gsl_part_3) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_3), gsl_partVanish_3.zw); }
  }
  let wall = gsl_part_3;
  let gsl_s1 = gsl_shape(crank, 1u);
  let gsl_s2 = gsl_shape(pin, 2u);
  let gsl_s3 = gsl_shape(rod, 3u);
  let gsl_s4 = gsl_shape(piston, 4u);
  let gsl_s5 = gsl_shape(wall, 5u);
  return gsl_min(gsl_min(gsl_min(gsl_min(gsl_s1, gsl_s2), gsl_s3), gsl_s4), gsl_s5);
}

fn gsl_colorField(p: vec3<f32>, time: f32, gsl_colorIdx: u32, gsl_param_18_10: f32) -> vec4<f32> {
  if (gsl_colorIdx == 1u) { return vec4<f32>(vec3<f32>(0.78, 0.62, 0.35), 1.0); }
  if (gsl_colorIdx == 2u) { return vec4<f32>(vec3<f32>(0.85, 0.72, 0.42), 1.0); }
  if (gsl_colorIdx == 3u) { return vec4<f32>(vec3<f32>(0.72, 0.72, 0.76), 1.0); }
  if (gsl_colorIdx == 4u) { return vec4<f32>(vec3<f32>(0.55, 0.3, 0.28), 1.0); }
  if (gsl_colorIdx == 5u) { return vec4<f32>(vec3<f32>(0.42, 0.48, 0.58), 1.0); }
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
