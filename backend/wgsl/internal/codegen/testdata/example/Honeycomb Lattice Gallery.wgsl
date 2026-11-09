fn gsl_floored(a: f32, n: f32) -> f32 {
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

fn map_triangleWaveOddPositive(x: f32, period: f32) -> f32 {
  return map_triangleDistance(x - period * 0.5, period);
}

fn builtin_unionSharp(fields: vec2<f32>) -> f32 {
  return min(fields.x, fields.y);
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

fn builtin_boxCentered(p: vec3<f32>, center: vec3<f32>, size: vec3<f32>) -> f32 {
  return builtin_gcl_closure_boxCentered(p, center, size);
}

fn map_repeatCentered(x: f32, period: f32) -> f32 {
  let half = period * 0.5;
  return gsl_floored((gsl_floored((x + half), period) + period), period) - half;
}

fn builtin_unionSharp_2(fields: vec3<f32>) -> f32 {
  return min(min(fields.x, fields.y), fields.z);
}

fn builtin_intersectSharp_2(fields: vec2<f32>) -> f32 {
  return max(fields.x, fields.y);
}

fn builtin_intersectRound_2(fields: vec2<f32>, radius: f32) -> f32 {
  let sharp = builtin_intersectSharp_2(fields);
  let sumOfSquares = 0.0 + max(fields.x + radius, 0.0) * max(fields.x + radius, 0.0) + max(fields.y + radius, 0.0) * max(fields.y + radius, 0.0);
  return sharp - max(sharp + radius, 0.0) + sqrt(max(sumOfSquares, 0.000000000000000000000000000001));
}

fn builtin_unionRound(fields: vec2<f32>, radius: f32) -> f32 {
  let negated = vec2<f32>(-fields.x, -fields.y);
  return -builtin_intersectRound_2(negated, radius);
}

fn builtin_gcl_closure_unionRound(fields: vec2<f32>) -> f32 {
  return builtin_unionRound(fields, 0.3);
}

fn map_triangleWaveEvenPositive(x: f32, period: f32) -> f32 {
  return map_triangleDistance(x, period);
}

fn builtin_unionRound_2(fields: vec3<f32>, radius: f32) -> f32 {
  let negated = vec3<f32>(-fields.x, -fields.y, -fields.z);
  return -builtin_intersectRound(negated, radius);
}

fn builtin_gcl_closure_unionRound_2(fields: vec3<f32>) -> f32 {
  return builtin_unionRound_2(fields, 0.3);
}

fn builtin_gcl_closure_unionRound_3(fields: vec3<f32>) -> f32 {
  return builtin_unionRound_2(fields, 0.3);
}

fn builtin_gcl_closure_unionRound_4(fields: vec3<f32>) -> f32 {
  return builtin_unionRound_2(fields, 0.2);
}

fn lattice_solidify(field: f32, bias: f32) -> f32 {
  return field - bias;
}

fn honeycomb_segmentDistance(point: vec2<f32>, a: vec2<f32>, b: vec2<f32>) -> f32 {
  let pa = point - a;
  let ba = b - a;
  let h = clamp(dot(pa, ba) / dot(ba, ba), 0.0, 1.0);
  return length(pa - ba * h);
}

fn honeycomb_gcl_closure_hexStar(point: vec2<f32>, starCenter: vec2<f32>, transition: vec2<f32>, topY: f32) -> f32 {
  let lineV = honeycomb_segmentDistance(point, starCenter, vec2<f32>(0.0, topY));
  let lineR = honeycomb_segmentDistance(point, starCenter, transition);
  let lineL = honeycomb_segmentDistance(point, starCenter, vec2<f32>(-transition.x, transition.y));
  return builtin_gcl_closure_unionRound_3(vec3<f32>(lineV, lineR, lineL));
}

fn honeycomb_gcl_closure_hexStar_2(point: vec2<f32>, starCenter: vec2<f32>, transition: vec2<f32>, topY: f32) -> f32 {
  let lineV = honeycomb_segmentDistance(point, starCenter, vec2<f32>(0.0, topY));
  let lineR = honeycomb_segmentDistance(point, starCenter, transition);
  let lineL = honeycomb_segmentDistance(point, starCenter, vec2<f32>(-transition.x, transition.y));
  return builtin_gcl_closure_unionRound_3(vec3<f32>(lineV, lineR, lineL));
}

fn honeycomb_gcl_closure_triStar(point: vec2<f32>, sx: f32, sy: f32) -> f32 {
  let distH = abs(point.y);
  let distP60 = abs(dot(point, normalize(vec2<f32>(sy, sx))));
  let distN60 = abs(dot(point, normalize(vec2<f32>(sy, -sx))));
  return builtin_gcl_closure_unionRound_4(vec3<f32>(distH, distP60, distN60));
}

fn honeycomb_gcl_closure_triStar_2(point: vec2<f32>, sx: f32, sy: f32) -> f32 {
  let distH = abs(point.y);
  let distP60 = abs(dot(point, normalize(vec2<f32>(sy, sx))));
  let distN60 = abs(dot(point, normalize(vec2<f32>(sy, -sx))));
  return builtin_gcl_closure_unionRound_4(vec3<f32>(distH, distP60, distN60));
}

fn specimen(p: vec3<f32>, field: f32, off: f32) -> f32 {
  return max(lattice_solidify(field, 0.5), builtin_boxCentered(p, vec3<f32>(off, 0.0, 0.0), vec3<f32>(28.0, 28.0, 28.0)));
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  gsl_partNear = vec3<f32>(1e10, 1e10, 1e10);
  let gsl_partLo = vec3<f32>(-116.0, -14.0, -14.0);
  let gsl_partHi = vec3<f32>(-88.0, 14.0, 14.0);
  let gsl_partBox = gsl_boxDistance(p, gsl_partLo, gsl_partHi);
  var gsl_part = gsl_partBox;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox <= gsl_vanishView.w) {
    let gsl_partVanish = gsl_vanish(gsl_partLo, gsl_partHi);
    gsl_part = 1e10;
    if (gsl_partVanish.x > 0.0) { gsl_part = specimen(p, builtin_unionSharp(vec2<f32>(map_triangleWaveOddPositive(p.x, 10.0), map_triangleWaveOddPositive(p.y, 10.0))), -102.0) + gsl_partVanish.y; }
    if (abs(gsl_part) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part), gsl_partVanish.zw); }
  }
  let sqr = gsl_part;
  let gsl_partLo_2 = vec3<f32>(-82.0, -14.0, -14.0);
  let gsl_partHi_2 = vec3<f32>(-54.0, 14.0, 14.0);
  let gsl_partBox_2 = gsl_boxDistance(p, gsl_partLo_2, gsl_partHi_2);
  var gsl_part_2 = gsl_partBox_2;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_2 <= gsl_vanishView.w) {
    let gsl_partVanish_2 = gsl_vanish(gsl_partLo_2, gsl_partHi_2);
    gsl_part_2 = 1e10;
    if (gsl_partVanish_2.x > 0.0) { gsl_part_2 = specimen(p, builtin_unionSharp(vec2<f32>(map_triangleDistance((p.x + p.y) / 1.4142135, 10.0), map_triangleDistance((p.x - p.y) / 1.4142135, 10.0))), -68.0) + gsl_partVanish_2.y; }
    if (abs(gsl_part_2) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_2), gsl_partVanish_2.zw); }
  }
  let dia = gsl_part_2;
  let gsl_partLo_3 = vec3<f32>(-48.0, -14.0, -14.0);
  let gsl_partHi_3 = vec3<f32>(-20.0, 14.0, 14.0);
  let gsl_partBox_3 = gsl_boxDistance(p, gsl_partLo_3, gsl_partHi_3);
  var gsl_part_3 = gsl_partBox_3;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_3 <= gsl_vanishView.w) {
    let gsl_partVanish_3 = gsl_vanish(gsl_partLo_3, gsl_partHi_3);
    gsl_part_3 = 1e10;
    if (gsl_partVanish_3.x > 0.0) { gsl_part_3 = specimen(p, builtin_unionSharp(vec2<f32>(builtin_unionSharp(vec2<f32>(map_triangleWaveOddPositive(p.x, 10.0), map_triangleWaveOddPositive(p.y, 10.0))), abs(dot(vec2<f32>(map_repeatCentered(p.x, 10.0), map_repeatCentered(p.y, 10.0)), vec2<f32>(0.70710677, 0.70710677))))), -34.0) + gsl_partVanish_3.y; }
    if (abs(gsl_part_3) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_3), gsl_partVanish_3.zw); }
  }
  let rf1 = gsl_part_3;
  let gsl_partLo_4 = vec3<f32>(-14.0, -14.0, -14.0);
  let gsl_partHi_4 = vec3<f32>(14.0, 14.0, 14.0);
  let gsl_partBox_4 = gsl_boxDistance(p, gsl_partLo_4, gsl_partHi_4);
  var gsl_part_4 = gsl_partBox_4;
  let gsl_tmp_1 = vec2<f32>(map_repeatCentered(p.x, 10.0), map_repeatCentered(p.y, 10.0));
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_4 <= gsl_vanishView.w) {
    let gsl_partVanish_4 = gsl_vanish(gsl_partLo_4, gsl_partHi_4);
    gsl_part_4 = 1e10;
    if (gsl_partVanish_4.x > 0.0) { gsl_part_4 = specimen(p, builtin_unionSharp_2(vec3<f32>(builtin_unionSharp(vec2<f32>(map_triangleWaveOddPositive(p.x, 10.0), map_triangleWaveOddPositive(p.y, 10.0))), abs(dot(gsl_tmp_1, vec2<f32>(0.70710677, 0.70710677))), abs(dot(gsl_tmp_1, vec2<f32>(-0.70710677, 0.70710677))))), 0.0) + gsl_partVanish_4.y; }
    if (abs(gsl_part_4) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_4), gsl_partVanish_4.zw); }
  }
  let rf2 = gsl_part_4;
  let gsl_partLo_5 = vec3<f32>(20.0, -14.0, -14.0);
  let gsl_partHi_5 = vec3<f32>(48.0, 14.0, 14.0);
  let gsl_partBox_5 = gsl_boxDistance(p, gsl_partLo_5, gsl_partHi_5);
  var gsl_part_5 = gsl_partBox_5;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_5 <= gsl_vanishView.w) {
    let gsl_partVanish_5 = gsl_vanish(gsl_partLo_5, gsl_partHi_5);
    gsl_part_5 = 1e10;
    if (gsl_partVanish_5.x > 0.0) { gsl_part_5 = specimen(p, builtin_gcl_closure_unionRound_2(vec3<f32>(builtin_gcl_closure_unionRound(vec2<f32>(map_triangleWaveOddPositive(p.x, 12.0), map_triangleWaveOddPositive(p.y, 12.0))), builtin_gcl_closure_unionRound(vec2<f32>(map_triangleWaveEvenPositive(p.x, 12.0), map_triangleWaveEvenPositive(p.y, 12.0))), builtin_gcl_closure_unionRound(vec2<f32>(map_triangleWaveEvenPositive((p.x + p.y) / 1.4142135, 8.485281), map_triangleWaveEvenPositive((p.x - p.y) / 1.4142135, 8.485281))))), 34.0) + gsl_partVanish_5.y; }
    if (abs(gsl_part_5) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_5), gsl_partVanish_5.zw); }
  }
  let oct = gsl_part_5;
  let gsl_partLo_6 = vec3<f32>(54.0, -14.0, -14.0);
  let gsl_partHi_6 = vec3<f32>(82.0, 14.0, 14.0);
  let gsl_partBox_6 = gsl_boxDistance(p, gsl_partLo_6, gsl_partHi_6);
  var gsl_part_6 = gsl_partBox_6;
  let gsl_tmp_2 = abs(map_repeatCentered(p.x, 8.0));
  let gsl_tmp_3 = abs(map_repeatCentered(p.y, 13.856406));
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_6 <= gsl_vanishView.w) {
    let gsl_partVanish_6 = gsl_vanish(gsl_partLo_6, gsl_partHi_6);
    gsl_part_6 = 1e10;
    if (gsl_partVanish_6.x > 0.0) { gsl_part_6 = specimen(p, select(honeycomb_gcl_closure_hexStar_2(vec2<f32>(gsl_tmp_2 - 4.0, 6.928203 - gsl_tmp_3), vec2<f32>(0.0, 4.618802), vec2<f32>(4.0, 2.309401), 13.856406), honeycomb_gcl_closure_hexStar(vec2<f32>(gsl_tmp_2, gsl_tmp_3), vec2<f32>(0.0, 4.618802), vec2<f32>(4.0, 2.309401), 13.856406), gsl_tmp_2 < 2.0), 68.0) + gsl_partVanish_6.y; }
    if (abs(gsl_part_6) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_6), gsl_partVanish_6.zw); }
  }
  let hex = gsl_part_6;
  let gsl_partLo_7 = vec3<f32>(88.0, -14.0, -14.0);
  let gsl_partHi_7 = vec3<f32>(116.0, 14.0, 14.0);
  let gsl_partBox_7 = gsl_boxDistance(p, gsl_partLo_7, gsl_partHi_7);
  var gsl_part_7 = gsl_partBox_7;
  let gsl_tmp_4 = abs(map_repeatCentered(p.x, 8.0));
  let gsl_tmp_5 = abs(map_repeatCentered(p.y, 13.856406));
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_7 <= gsl_vanishView.w) {
    let gsl_partVanish_7 = gsl_vanish(gsl_partLo_7, gsl_partHi_7);
    gsl_part_7 = 1e10;
    if (gsl_partVanish_7.x > 0.0) { gsl_part_7 = specimen(p, select(honeycomb_gcl_closure_triStar_2(vec2<f32>(gsl_tmp_4 - 4.0, gsl_tmp_5 - 6.928203), 8.0, 13.856406), honeycomb_gcl_closure_triStar(vec2<f32>(gsl_tmp_4, gsl_tmp_5), 8.0, 13.856406), gsl_tmp_5 < 3.4641016), 102.0) + gsl_partVanish_7.y; }
    if (abs(gsl_part_7) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_7), gsl_partVanish_7.zw); }
  }
  let tri = gsl_part_7;
  let gsl_s1 = gsl_shape(sqr, 1u);
  let gsl_s2 = gsl_shape(dia, 2u);
  let gsl_s3 = gsl_shape(rf1, 3u);
  let gsl_s4 = gsl_shape(rf2, 4u);
  let gsl_s5 = gsl_shape(oct, 5u);
  let gsl_s6 = gsl_shape(hex, 6u);
  let gsl_s7 = gsl_shape(tri, 7u);
  return gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_s1, gsl_s2), gsl_s3), gsl_s4), gsl_s5), gsl_s6), gsl_s7);
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
