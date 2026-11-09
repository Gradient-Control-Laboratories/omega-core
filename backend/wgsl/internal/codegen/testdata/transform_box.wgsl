fn builtin_intersectRound_2(d1: f32, d2: f32, radius: f32) -> f32 {
  let inset1 = max(d1 + radius, 0.0);
  let inset2 = max(d2 + radius, 0.0);
  return min(-radius, max(d1, d2)) + length(vec2<f32>(inset1, inset2));
}

fn builtin_unionRound(d1: f32, d2: f32, radius: f32) -> f32 {
  return -builtin_intersectRound_2(-d1, -d2, radius);
}

fn gsl_remap(p: vec3<f32>, gsl_param_3_9: f32) -> f32 {
  let body = builtin_unionRound(builtin_gcl_closure_boxCentered(p, vec3<f32>(2.0, 1.0, 1.0) * 0.5, abs(vec3<f32>(2.0, 1.0, 1.0))), length(p - vec3<f32>(2.0, 0.5, 0.5)) - 0.6, 0.2);
  return body;
}

fn gsl_remap_2(p: vec3<f32>, gsl_param_3_9: f32) -> f32 {
  return gsl_remap(p - vec3<f32>(1.0, 2.0, 0.0), gsl_param_3_9);
}

fn builtin_gcl_closure_box(p: vec3<f32>, pointA: vec3<f32>, pointB: vec3<f32>) -> f32 {
  let center = (pointA + pointB) * 0.5;
  let size = abs(pointB - pointA);
  return builtin_gcl_closure_boxCentered(p, center, size);
}

fn builtin_box(p: vec3<f32>, pointA: vec3<f32>, pointB: vec3<f32>) -> f32 {
  return builtin_gcl_closure_box(p, pointA, pointB);
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

fn sdShape(p: vec3<f32>, time: f32, gsl_param_3_9: f32) -> gsl_shape {
  gsl_partNear = vec3<f32>(1e10, 1e10, 1e10);
  let angle = gsl_param_3_9;
  let gsl_partLo = vec3<f32>(cos(angle) * 2.3 - sin(angle) * 2.5 - (abs(cos(angle)) * 1.5 + abs(sin(angle)) * 0.8), sin(angle) * 2.3 + cos(angle) * 2.5 - (abs(sin(angle)) * 1.5 + abs(cos(angle)) * 0.8), -0.3);
  let gsl_partHi = vec3<f32>(cos(angle) * 2.3 - sin(angle) * 2.5 + abs(cos(angle)) * 1.5 + abs(sin(angle)) * 0.8, sin(angle) * 2.3 + cos(angle) * 2.5 + abs(sin(angle)) * 1.5 + abs(cos(angle)) * 0.8, 1.3);
  let gsl_partBox = gsl_boxDistance(p, gsl_partLo, gsl_partHi);
  var gsl_part = gsl_partBox;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox <= gsl_vanishView.w) {
    let gsl_partVanish = gsl_vanish(gsl_partLo, gsl_partHi);
    gsl_part = 1e10;
    if (gsl_partVanish.x > 0.0) { gsl_part = gsl_remap_2(vec3<f32>(p.x * cos(angle) + p.y * sin(angle), -p.x * sin(angle) + p.y * cos(angle), p.z), gsl_param_3_9) + gsl_partVanish.y; }
    if (abs(gsl_part) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part), gsl_partVanish.zw); }
  }
  let placed = gsl_part;
  let gsl_partLo_2 = vec3<f32>(-0.1, -0.15, -0.15);
  let gsl_partHi_2 = vec3<f32>(1.4, 0.65, 0.65);
  let gsl_partBox_2 = gsl_boxDistance(p, gsl_partLo_2, gsl_partHi_2);
  var gsl_part_2 = gsl_partBox_2;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_2 <= gsl_vanishView.w) {
    let gsl_partVanish_2 = gsl_vanish(gsl_partLo_2, gsl_partHi_2);
    gsl_part_2 = 1e10;
    if (gsl_partVanish_2.x > 0.0) { gsl_part_2 = gsl_remap(p / 0.5, gsl_param_3_9) * 0.5 + gsl_partVanish_2.y; }
    if (abs(gsl_part_2) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_2), gsl_partVanish_2.zw); }
  }
  let scaled = gsl_part_2;
  let gsl_partLo_3 = vec3<f32>(min(cos(angle) * 2.3 - sin(angle) * 2.5 - (abs(cos(angle)) * 1.5 + abs(sin(angle)) * 0.8), cos(angle) * 2.3 - sin(angle) * 2.5 + abs(cos(angle)) * 1.5 + abs(sin(angle)) * 0.8), min(sin(angle) * 2.3 + cos(angle) * 2.5 - (abs(sin(angle)) * 1.5 + abs(cos(angle)) * 0.8), sin(angle) * 2.3 + cos(angle) * 2.5 + abs(sin(angle)) * 1.5 + abs(cos(angle)) * 0.8), -0.3);
  let gsl_partHi_3 = vec3<f32>(max(cos(angle) * 2.3 - sin(angle) * 2.5 - (abs(cos(angle)) * 1.5 + abs(sin(angle)) * 0.8), cos(angle) * 2.3 - sin(angle) * 2.5 + abs(cos(angle)) * 1.5 + abs(sin(angle)) * 0.8), max(sin(angle) * 2.3 + cos(angle) * 2.5 - (abs(sin(angle)) * 1.5 + abs(cos(angle)) * 0.8), sin(angle) * 2.3 + cos(angle) * 2.5 + abs(sin(angle)) * 1.5 + abs(cos(angle)) * 0.8), 1.3);
  let gsl_partBox_3 = gsl_boxDistance(p, gsl_partLo_3, gsl_partHi_3);
  var gsl_part_3 = gsl_partBox_3;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_3 <= gsl_vanishView.w) {
    let gsl_partVanish_3 = gsl_vanish(gsl_partLo_3, gsl_partHi_3);
    gsl_part_3 = 1e10;
    if (gsl_partVanish_3.x > 0.0) { gsl_part_3 = builtin_box(p, vec3<f32>(cos(angle) * 2.3 - sin(angle) * 2.5 - (abs(cos(angle)) * 1.5 + abs(sin(angle)) * 0.8), sin(angle) * 2.3 + cos(angle) * 2.5 - (abs(sin(angle)) * 1.5 + abs(cos(angle)) * 0.8), -0.3), vec3<f32>(cos(angle) * 2.3 - sin(angle) * 2.5 + abs(cos(angle)) * 1.5 + abs(sin(angle)) * 0.8, sin(angle) * 2.3 + cos(angle) * 2.5 + abs(sin(angle)) * 1.5 + abs(cos(angle)) * 0.8, 1.3)) + gsl_partVanish_3.y; }
    if (abs(gsl_part_3) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_3), gsl_partVanish_3.zw); }
  }
  let hull = gsl_part_3;
  let gsl_s1 = gsl_shape(placed, 1u);
  let gsl_s2 = gsl_shape(scaled, 2u);
  let gsl_s3 = gsl_shape(hull, 3u);
  return gsl_min(gsl_min(gsl_s1, gsl_s2), gsl_s3);
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
