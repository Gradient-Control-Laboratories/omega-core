fn boxAt(p: vec3<f32>, center: vec3<f32>, half: vec3<f32>) -> f32 {
  let q = p - center;
  return max(max(max(max(max(-half.x - q.x, q.x - half.x), -half.y - q.y), q.y - half.y), -half.z - q.z), q.z - half.z);
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  gsl_partNear = vec3<f32>(1e10, 1e10, 1e10);
  let gsl_partLo = vec3<f32>(-16.0, -16.0, -1.0);
  let gsl_partHi = vec3<f32>(16.0, 16.0, 20.0);
  let gsl_partBox = gsl_boxDistance(p, gsl_partLo, gsl_partHi);
  var gsl_part = gsl_partBox;
  let crateOuter = boxAt(p, vec3<f32>(0.0, 0.0, 9.5), vec3<f32>(16.0, 16.0, 10.5));
  let crateInner = boxAt(p, vec3<f32>(0.0, 0.0, 10.0), vec3<f32>(15.0, 15.0, 10.0));
  if (gsl_vanishView.w <= 0.0 || gsl_partBox <= gsl_vanishView.w) {
    let gsl_partVanish = gsl_vanish(gsl_partLo, gsl_partHi);
    gsl_part = 1e10;
    if (gsl_partVanish.x > 0.0) { gsl_part = max(crateOuter, -crateInner) + gsl_partVanish.y; }
    if (abs(gsl_part) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part), gsl_partVanish.zw); }
  }
  let gsl_var_crate = gsl_part;
  let gsl_partLo_2 = vec3<f32>(-11.0, -9.0, 1.0);
  let gsl_partHi_2 = vec3<f32>(-1.0, -1.0, 11.0);
  let gsl_partBox_2 = gsl_boxDistance(p, gsl_partLo_2, gsl_partHi_2);
  var gsl_part_2 = gsl_partBox_2;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_2 <= gsl_vanishView.w) {
    let gsl_partVanish_2 = gsl_vanish(gsl_partLo_2, gsl_partHi_2);
    gsl_part_2 = 1e10;
    if (gsl_partVanish_2.x > 0.0) { gsl_part_2 = boxAt(p, vec3<f32>(-6.0, -5.0, 6.0), vec3<f32>(5.0, 4.0, 5.0)) + gsl_partVanish_2.y; }
    if (abs(gsl_part_2) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_2), gsl_partVanish_2.zw); }
  }
  let block = gsl_part_2;
  let gsl_partLo_3 = vec3<f32>(-3.0, -7.0, 1.0);
  let gsl_partHi_3 = vec3<f32>(5.0, 1.0, 11.0);
  let gsl_partBox_3 = gsl_boxDistance(p, gsl_partLo_3, gsl_partHi_3);
  var gsl_part_3 = gsl_partBox_3;
  let gsl_tmp_1 = p - vec3<f32>(1.0, -3.0, 6.0);
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_3 <= gsl_vanishView.w) {
    let gsl_partVanish_3 = gsl_vanish(gsl_partLo_3, gsl_partHi_3);
    gsl_part_3 = 1e10;
    if (gsl_partVanish_3.x > 0.0) { gsl_part_3 = max(length(gsl_tmp_1.xy) - 4.0, abs(gsl_tmp_1.z) - 5.0) + gsl_partVanish_3.y; }
    if (abs(gsl_part_3) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_3), gsl_partVanish_3.zw); }
  }
  let drum = gsl_part_3;
  let gsl_partLo_4 = vec3<f32>(3.0, 7.0, 3.0);
  let gsl_partHi_4 = vec3<f32>(13.0, 17.0, 13.0);
  let gsl_partBox_4 = gsl_boxDistance(p, gsl_partLo_4, gsl_partHi_4);
  var gsl_part_4 = gsl_partBox_4;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_4 <= gsl_vanishView.w) {
    let gsl_partVanish_4 = gsl_vanish(gsl_partLo_4, gsl_partHi_4);
    gsl_part_4 = 1e10;
    if (gsl_partVanish_4.x > 0.0) { gsl_part_4 = length(p - vec3<f32>(8.0, 12.0, 8.0)) - 5.0 + gsl_partVanish_4.y; }
    if (abs(gsl_part_4) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_4), gsl_partVanish_4.zw); }
  }
  let ball = gsl_part_4;
  let gsl_s1 = gsl_shape(gsl_var_crate, 1u);
  let gsl_s2 = gsl_shape(block, 2u);
  let gsl_s3 = gsl_shape(drum, 3u);
  let gsl_s4 = gsl_shape(ball, 4u);
  return gsl_min(gsl_min(gsl_min(gsl_s1, gsl_s2), gsl_s3), gsl_s4);
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
