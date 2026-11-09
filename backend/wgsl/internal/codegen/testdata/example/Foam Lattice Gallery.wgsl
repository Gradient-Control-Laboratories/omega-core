fn gsl_floored(a: f32, n: f32) -> f32 {
  return a - n * floor(a / n);
}

fn gsl_floored_vec3(a: vec3<f32>, n: vec3<f32>) -> vec3<f32> {
  return a - n * floor(a / n);
}

fn map_triangleWave(x: f32, period: f32) -> f32 {
  let half = period * 0.5;
  let twoPeriod = period * 2.0;
  let u = gsl_floored((gsl_floored((x + half), twoPeriod) + twoPeriod), twoPeriod);
  return half - abs(u - period);
}

fn map_triangleDistance(x: f32, period: f32) -> f32 {
  return abs(map_triangleWave(x, period));
}

fn map_triangleWaveEvenPositive(x: f32, period: f32) -> f32 {
  return map_triangleDistance(x, period);
}

fn builtin_intersectSharp(fields: vec3<f32>) -> f32 {
  return max(max(fields.x, fields.y), fields.z);
}

fn builtin_intersectRound(fields: vec3<f32>, radius: f32) -> f32 {
  let sharp = builtin_intersectSharp(fields);
  let sumOfSquares = 0.0 + max(fields.x + radius, 0.0) * max(fields.x + radius, 0.0) + max(fields.y + radius, 0.0) * max(fields.y + radius, 0.0) + max(fields.z + radius, 0.0) * max(fields.z + radius, 0.0);
  return sharp - max(sharp + radius, 0.0) + sqrt(max(sumOfSquares, 0.000000000000000000000000000001));
}

fn builtin_unionRound(fields: vec3<f32>, radius: f32) -> f32 {
  let negated = vec3<f32>(-fields.x, -fields.y, -fields.z);
  return -builtin_intersectRound(negated, radius);
}

fn builtin_gcl_closure_unionRound(fields: vec3<f32>) -> f32 {
  return builtin_unionRound(fields, 0.5);
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

fn builtin_boxCentered(p: vec3<f32>, center: vec3<f32>, size: vec3<f32>) -> f32 {
  return builtin_gcl_closure_boxCentered(p, center, size);
}

fn map_triangleWaveOddPositive(x: f32, period: f32) -> f32 {
  return map_triangleDistance(x - period * 0.5, period);
}

fn builtin_gcl_closure_unionRound_2(fields: vec3<f32>) -> f32 {
  return builtin_unionRound(fields, 0.5);
}

fn map_repeatCentered(x: vec3<f32>, period: vec3<f32>) -> vec3<f32> {
  let half = period * 0.5;
  return gsl_floored_vec3((gsl_floored_vec3((x + half), period) + period), period) - half;
}

fn builtin_plane(p: vec3<f32>, point: vec3<f32>, normal: vec3<f32>) -> f32 {
  return dot(p - point, normalize(normal));
}

fn builtin_intersectSharp_2(fields: vec2<f32>) -> f32 {
  return max(fields.x, fields.y);
}

fn builtin_intersectRound_2(fields: vec2<f32>, radius: f32) -> f32 {
  let sharp = builtin_intersectSharp_2(fields);
  let sumOfSquares = 0.0 + max(fields.x + radius, 0.0) * max(fields.x + radius, 0.0) + max(fields.y + radius, 0.0) * max(fields.y + radius, 0.0);
  return sharp - max(sharp + radius, 0.0) + sqrt(max(sumOfSquares, 0.000000000000000000000000000001));
}

fn builtin_unionRound_2(fields: vec2<f32>, radius: f32) -> f32 {
  let negated = vec2<f32>(-fields.x, -fields.y);
  return -builtin_intersectRound_2(negated, radius);
}

fn builtin_gcl_closure_unionRound_3(fields: vec2<f32>) -> f32 {
  return builtin_unionRound_2(fields, 0.5);
}

fn gsl_remap(p: vec3<f32>) -> f32 {
  return builtin_gcl_closure_unionRound_3(vec2<f32>(abs(builtin_plane(p, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(12.0, 12.0, 0.0))), abs(builtin_plane(p, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(12.0, -12.0, 0.0)))));
}

fn builtin_gcl_closure_unionRound_4(fields: vec3<f32>) -> f32 {
  return builtin_unionRound(fields, 0.5);
}

fn builtin_gcl_closure_unionRound_5(fields: vec3<f32>) -> f32 {
  return builtin_unionRound(fields, 0.5);
}

fn builtin_gcl_closure_unionRound_6(fields: vec2<f32>) -> f32 {
  return builtin_unionRound_2(fields, 0.5);
}

fn gsl_remap_2(p: vec3<f32>) -> f32 {
  return builtin_gcl_closure_unionRound_6(vec2<f32>(builtin_gcl_closure_unionRound_5(vec3<f32>(abs(builtin_plane(p, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, -1.0, -1.0))), abs(builtin_plane(p, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(-1.0, 1.0, -1.0))), abs(builtin_plane(p, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(-1.0, -1.0, 1.0))))), abs(builtin_plane(p, vec3<f32>(0.0, 6.0, 6.0), vec3<f32>(1.0, 1.0, 1.0)))));
}

fn lattice_solidify(field: f32, bias: f32) -> f32 {
  return field - bias;
}

fn specimen(p: vec3<f32>, field: f32, off: f32) -> f32 {
  return max(lattice_solidify(field, 0.6), builtin_boxCentered(p, vec3<f32>(off, 0.0, 0.0), vec3<f32>(28.0, 28.0, 28.0)));
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  gsl_partNear = vec3<f32>(1e10, 1e10, 1e10);
  let gsl_partLo = vec3<f32>(-65.0, -14.0, -14.0);
  let gsl_partHi = vec3<f32>(-37.0, 14.0, 14.0);
  let gsl_partBox = gsl_boxDistance(p, gsl_partLo, gsl_partHi);
  var gsl_part = gsl_partBox;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox <= gsl_vanishView.w) {
    let gsl_partVanish = gsl_vanish(gsl_partLo, gsl_partHi);
    gsl_part = 1e10;
    if (gsl_partVanish.x > 0.0) { gsl_part = specimen(p, builtin_gcl_closure_unionRound(vec3<f32>(map_triangleWaveEvenPositive(p.x, 12.0), map_triangleWaveEvenPositive(p.y, 12.0), map_triangleWaveEvenPositive(p.z, 12.0))), -51.0) + gsl_partVanish.y; }
    if (abs(gsl_part) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part), gsl_partVanish.zw); }
  }
  let cub = gsl_part;
  let gsl_partLo_2 = vec3<f32>(-31.0, -14.0, -14.0);
  let gsl_partHi_2 = vec3<f32>(-3.0, 14.0, 14.0);
  let gsl_partBox_2 = gsl_boxDistance(p, gsl_partLo_2, gsl_partHi_2);
  var gsl_part_2 = gsl_partBox_2;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_2 <= gsl_vanishView.w) {
    let gsl_partVanish_2 = gsl_vanish(gsl_partLo_2, gsl_partHi_2);
    gsl_part_2 = 1e10;
    if (gsl_partVanish_2.x > 0.0) { gsl_part_2 = specimen(p, builtin_gcl_closure_unionRound_2(vec3<f32>(map_triangleWaveOddPositive(p.x, 12.0), map_triangleWaveOddPositive(p.y, 12.0), map_triangleWaveOddPositive(p.z, 12.0))), -17.0) + gsl_partVanish_2.y; }
    if (abs(gsl_part_2) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_2), gsl_partVanish_2.zw); }
  }
  let cgr = gsl_part_2;
  let gsl_partLo_3 = vec3<f32>(3.0, -14.0, -14.0);
  let gsl_partHi_3 = vec3<f32>(31.0, 14.0, 14.0);
  let gsl_partBox_3 = gsl_boxDistance(p, gsl_partLo_3, gsl_partHi_3);
  var gsl_part_3 = gsl_partBox_3;
  let gsl_tmp_1 = map_repeatCentered(p, vec3<f32>(12.0, 12.0, 12.0));
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_3 <= gsl_vanishView.w) {
    let gsl_partVanish_3 = gsl_vanish(gsl_partLo_3, gsl_partHi_3);
    gsl_part_3 = 1e10;
    if (gsl_partVanish_3.x > 0.0) { gsl_part_3 = specimen(p, builtin_gcl_closure_unionRound_4(vec3<f32>(gsl_remap(gsl_tmp_1), gsl_remap(vec3<f32>(gsl_tmp_1.y, gsl_tmp_1.z, gsl_tmp_1.x)), gsl_remap(vec3<f32>(gsl_tmp_1.z, gsl_tmp_1.x, gsl_tmp_1.y)))), 17.0) + gsl_partVanish_3.y; }
    if (abs(gsl_part_3) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_3), gsl_partVanish_3.zw); }
  }
  let bcc = gsl_part_3;
  let gsl_partLo_4 = vec3<f32>(37.0, -14.0, -14.0);
  let gsl_partHi_4 = vec3<f32>(65.0, 14.0, 14.0);
  let gsl_partBox_4 = gsl_boxDistance(p, gsl_partLo_4, gsl_partHi_4);
  var gsl_part_4 = gsl_partBox_4;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_4 <= gsl_vanishView.w) {
    let gsl_partVanish_4 = gsl_vanish(gsl_partLo_4, gsl_partHi_4);
    gsl_part_4 = 1e10;
    if (gsl_partVanish_4.x > 0.0) { gsl_part_4 = specimen(p, gsl_remap_2(abs(map_repeatCentered(p, vec3<f32>(12.0, 12.0, 12.0)))), 51.0) + gsl_partVanish_4.y; }
    if (abs(gsl_part_4) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_4), gsl_partVanish_4.zw); }
  }
  let fcc = gsl_part_4;
  let gsl_s1 = gsl_shape(cub, 1u);
  let gsl_s2 = gsl_shape(cgr, 2u);
  let gsl_s3 = gsl_shape(bcc, 3u);
  let gsl_s4 = gsl_shape(fcc, 4u);
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
