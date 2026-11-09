fn builtin_cropToBox(p: vec3<f32>, shape: f32, low: vec3<f32>, high: vec3<f32>) -> f32 {
  let lowCorner = min(low, high) - 0.01;
  let highCorner = max(low, high) + 0.01;
  let d = abs(p - (lowCorner + highCorner) * 0.5) - (highCorner - lowCorner) * 0.5;
  let outside = length(vec3<f32>(max(d.x, 0.0), max(d.y, 0.0), max(d.z, 0.0)));
  let inside = min(max(d.x, max(d.y, d.z)), 0.0);
  let gsl_capped = max(shape, outside + inside);
  if (gsl_boxCap) { return gsl_capped; }
  if (gsl_capped <= max(shape, 0.0)) { return shape; }
  return select(-1.0, 1.0, shape >= 0.0) * max(abs(shape), gsl_capped);
}

var<private> gsl_boxCap: bool = false;

fn builtin_torus(p: vec3<f32>, majorRadius: f32, minorRadius: f32) -> f32 {
  let radial = length(p.xy) - majorRadius;
  return length(vec2<f32>(radial, p.z)) - minorRadius;
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

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  gsl_partNear = vec3<f32>(1e10, 1e10, 1e10);
  let gsl_partVanish = gsl_vanish(vec3<f32>(-0.51, -1.01, -1.01), vec3<f32>(1.01, 1.01, 1.01));
  var gsl_part = 1e10;
  if (gsl_partVanish.x > 0.0) { gsl_part = builtin_cropToBox(p, length(p) - 1.0, vec3<f32>(-0.5, -1.0, -1.0), vec3<f32>(1.0, 1.0, 1.0)) + gsl_partVanish.y; }
  if (abs(gsl_part) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part), gsl_partVanish.zw); }
  let ball = gsl_part;
  let gsl_partVanish_2 = gsl_vanish(vec3<f32>(-3.26, -3.26, -0.26), vec3<f32>(3.26, 3.26, 0.26));
  var gsl_part_2 = 1e10;
  if (gsl_partVanish_2.x > 0.0) { gsl_part_2 = builtin_cropToBox(p, builtin_torus(vec3<f32>(p.y, p.x, p.z), 3.0, 0.25), vec3<f32>(-3.25, -3.25, -0.25), vec3<f32>(3.25, 3.25, 0.25)) + gsl_partVanish_2.y; }
  if (abs(gsl_part_2) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_2), gsl_partVanish_2.zw); }
  let big = gsl_part_2;
  let gsl_partLo = vec3<f32>(-3.26, -3.26, -0.26);
  let gsl_partHi = vec3<f32>(3.26, 3.26, 0.26);
  let gsl_partBox = gsl_boxDistance(p, gsl_partLo, gsl_partHi);
  var gsl_part_3 = gsl_partBox;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox <= gsl_vanishView.w) {
    let gsl_partVanish_3 = gsl_vanish(gsl_partLo, gsl_partHi);
    gsl_part_3 = 1e10;
    if (gsl_partVanish_3.x > 0.0) { gsl_part_3 = builtin_gcl_closure_boxCentered(p, vec3<f32>(0.0, 0.0, 0.0) * 0.5, abs(vec3<f32>(6.52, 6.52, 0.52))) + gsl_partVanish_3.y; }
    if (abs(gsl_part_3) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_3), gsl_partVanish_3.zw); }
  }
  let hull = gsl_part_3;
  let gsl_s1 = gsl_shape(ball, 1u);
  let gsl_s2 = gsl_shape(big, 2u);
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
