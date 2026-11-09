fn gsl_floored_vec3(a: vec3<f32>, n: vec3<f32>) -> vec3<f32> {
  return a - n * floor(a / n);
}

fn gsl_floored_vec2(a: vec2<f32>, n: vec2<f32>) -> vec2<f32> {
  return a - n * floor(a / n);
}

fn builtin_lineSegment(p: vec3<f32>, pointA: vec3<f32>, pointB: vec3<f32>) -> f32 {
  let pa = p - pointA;
  let ba = pointB - pointA;
  let h = clamp(dot(pa, ba) / dot(ba, ba), 0.0, 1.0);
  return length(pa - ba * h);
}

fn builtin_unionSharp(fields: array<f32, 6>) -> f32 {
  return min(min(min(min(min(fields[0], fields[1]), fields[2]), fields[3]), fields[4]), fields[5]);
}

fn gsl_remap(p: vec3<f32>) -> f32 {
  return builtin_unionSharp(array<f32, 6>(builtin_lineSegment(p, vec3<f32>(0.0, 6.0, 0.0), vec3<f32>(0.0, 0.0, 6.0)), builtin_lineSegment(p, vec3<f32>(6.0, 0.0, 0.0), vec3<f32>(0.0, 0.0, 6.0)), builtin_lineSegment(p, vec3<f32>(6.0, 0.0, 0.0), vec3<f32>(0.0, 6.0, 0.0)), builtin_lineSegment(p, vec3<f32>(6.0, 6.0, 6.0), vec3<f32>(0.0, 0.0, 6.0)), builtin_lineSegment(p, vec3<f32>(6.0, 6.0, 6.0), vec3<f32>(0.0, 6.0, 0.0)), builtin_lineSegment(p, vec3<f32>(6.0, 6.0, 6.0), vec3<f32>(6.0, 0.0, 0.0))));
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

fn builtin_unionSharp_2(fields: array<f32, 8>) -> f32 {
  return min(min(min(min(min(min(min(fields[0], fields[1]), fields[2]), fields[3]), fields[4]), fields[5]), fields[6]), fields[7]);
}

fn gsl_remap_2(p: vec3<f32>) -> f32 {
  return builtin_unionSharp_2(array<f32, 8>(builtin_lineSegment(p, vec3<f32>(4.0, 4.0, 3.0), vec3<f32>(2.0, 2.0, 6.0)), builtin_lineSegment(p, vec3<f32>(4.0, 4.0, 3.0), vec3<f32>(2.0, 6.0, 0.0)), builtin_lineSegment(p, vec3<f32>(6.0, 2.0, 0.0), vec3<f32>(4.0, 4.0, 3.0)), builtin_lineSegment(p, vec3<f32>(6.0, 6.0, 6.0), vec3<f32>(4.0, 4.0, 3.0)), builtin_lineSegment(p, vec3<f32>(6.0, -2.0, 0.0), vec3<f32>(6.0, 2.0, 0.0)), builtin_lineSegment(p, vec3<f32>(-2.0, 6.0, 0.0), vec3<f32>(2.0, 6.0, 0.0)), builtin_lineSegment(p, vec3<f32>(-2.0, 2.0, 6.0), vec3<f32>(2.0, 2.0, 6.0)), builtin_lineSegment(p, vec3<f32>(2.0, -2.0, 6.0), vec3<f32>(2.0, 2.0, 6.0))));
}

fn builtin_unionSharp_3(fields: vec3<f32>) -> f32 {
  return min(min(fields.x, fields.y), fields.z);
}

fn gsl_remap_3(p: vec3<f32>) -> f32 {
  return builtin_unionSharp_3(vec3<f32>(builtin_lineSegment(p, vec3<f32>(6.0, 6.0, 6.0), vec3<f32>(0.0, 0.0, 6.0)), builtin_lineSegment(p, vec3<f32>(6.0, 6.0, 6.0), vec3<f32>(0.0, 6.0, 0.0)), builtin_lineSegment(p, vec3<f32>(6.0, 6.0, 6.0), vec3<f32>(6.0, 0.0, 0.0))));
}

fn gsl_remap_4(p: vec3<f32>) -> f32 {
  return builtin_unionSharp(array<f32, 6>(builtin_lineSegment(p, vec3<f32>(6.0, 3.0, 6.0), vec3<f32>(3.0, 6.0, 6.0)), builtin_lineSegment(p, vec3<f32>(6.0, 6.0, 3.0), vec3<f32>(3.0, 6.0, 6.0)), builtin_lineSegment(p, vec3<f32>(6.0, 6.0, 3.0), vec3<f32>(6.0, 3.0, 6.0)), builtin_lineSegment(p, vec3<f32>(3.0, 6.0, 6.0), vec3<f32>(0.0, 6.0, 6.0)), builtin_lineSegment(p, vec3<f32>(6.0, 3.0, 6.0), vec3<f32>(6.0, 0.0, 6.0)), builtin_lineSegment(p, vec3<f32>(6.0, 6.0, 3.0), vec3<f32>(6.0, 6.0, 0.0))));
}

fn gsl_remap_5(p: vec3<f32>) -> f32 {
  return builtin_unionSharp(array<f32, 6>(builtin_lineSegment(p, vec3<f32>(0.0, 6.0, 3.0), vec3<f32>(0.0, 3.0, 6.0)), builtin_lineSegment(p, vec3<f32>(3.0, 0.0, 6.0), vec3<f32>(0.0, 3.0, 6.0)), builtin_lineSegment(p, vec3<f32>(3.0, 6.0, 0.0), vec3<f32>(0.0, 6.0, 3.0)), builtin_lineSegment(p, vec3<f32>(6.0, 0.0, 3.0), vec3<f32>(3.0, 0.0, 6.0)), builtin_lineSegment(p, vec3<f32>(6.0, 3.0, 0.0), vec3<f32>(3.0, 6.0, 0.0)), builtin_lineSegment(p, vec3<f32>(6.0, 3.0, 0.0), vec3<f32>(6.0, 0.0, 3.0))));
}

fn builtin_unionSharp_4(fields: vec4<f32>) -> f32 {
  return min(min(min(fields.x, fields.y), fields.z), fields.w);
}

fn gsl_remap_6(p: vec3<f32>) -> f32 {
  return builtin_unionSharp_4(vec4<f32>(builtin_lineSegment(p, vec3<f32>(3.0, 3.0, 3.0), vec3<f32>(0.0, 0.0, 6.0)), builtin_lineSegment(p, vec3<f32>(3.0, 3.0, 3.0), vec3<f32>(0.0, 6.0, 0.0)), builtin_lineSegment(p, vec3<f32>(6.0, 0.0, 0.0), vec3<f32>(3.0, 3.0, 3.0)), builtin_lineSegment(p, vec3<f32>(6.0, 6.0, 6.0), vec3<f32>(3.0, 3.0, 3.0))));
}

fn gsl_remap_7(p: vec3<f32>) -> f32 {
  return builtin_unionSharp_3(vec3<f32>(builtin_lineSegment(p, vec3<f32>(-6.0, 0.0, 0.0), vec3<f32>(6.0, 0.0, 0.0)), builtin_lineSegment(p, vec3<f32>(0.0, -6.0, 0.0), vec3<f32>(0.0, 6.0, 0.0)), builtin_lineSegment(p, vec3<f32>(0.0, 0.0, -6.0), vec3<f32>(0.0, 0.0, 6.0))));
}

fn gsl_remap_8(p: vec3<f32>) -> f32 {
  return builtin_unionSharp_4(vec4<f32>(builtin_lineSegment(p, vec3<f32>(6.0, 6.0, 6.0), vec3<f32>(0.0, 6.0, 6.0)), builtin_lineSegment(p, vec3<f32>(6.0, 6.0, 6.0), vec3<f32>(6.0, 0.0, 6.0)), builtin_lineSegment(p, vec3<f32>(6.0, 6.0, 6.0), vec3<f32>(6.0, 6.0, 0.0)), builtin_lineSegment(p, vec3<f32>(6.0, 6.0, 6.0), vec3<f32>(0.0, 0.0, 0.0))));
}

fn builtin_unionSharp_5(fields: vec2<f32>) -> f32 {
  return min(fields.x, fields.y);
}

fn builtin_unionSharp_6(fields: array<f32, 7>) -> f32 {
  return min(min(min(min(min(min(fields[0], fields[1]), fields[2]), fields[3]), fields[4]), fields[5]), fields[6]);
}

fn gsl_remap_9(p: vec3<f32>) -> f32 {
  return builtin_unionSharp_6(array<f32, 7>(builtin_lineSegment(p, vec3<f32>(0.0, 6.0, 1.8), vec3<f32>(0.0, 0.0, 4.8)), builtin_lineSegment(p, vec3<f32>(0.0, 6.0, 6.0), vec3<f32>(0.0, 6.0, 1.8)), builtin_lineSegment(p, vec3<f32>(6.0, 0.0, 1.8), vec3<f32>(0.0, 0.0, 4.8)), builtin_lineSegment(p, vec3<f32>(6.0, 0.0, 6.0), vec3<f32>(6.0, 0.0, 1.8)), builtin_lineSegment(p, vec3<f32>(6.0, 6.0, 4.8), vec3<f32>(0.0, 6.0, 1.8)), builtin_lineSegment(p, vec3<f32>(6.0, 6.0, 4.8), vec3<f32>(6.0, 0.0, 1.8)), builtin_lineSegment(p, vec3<f32>(6.0, 6.0, 4.8), vec3<f32>(6.0, 6.0, 0.0))));
}

fn builtin_unionRound(fields: vec3<f32>, radius: f32) -> f32 {
  let negated = vec3<f32>(-fields.x, -fields.y, -fields.z);
  return -builtin_intersectRound(negated, radius);
}

fn builtin_gcl_closure_unionRound(fields: vec3<f32>) -> f32 {
  return builtin_unionRound(fields, 0.4);
}

fn gsl_remap_10(p: vec3<f32>) -> f32 {
  return builtin_gcl_closure_unionRound(vec3<f32>(builtin_lineSegment(p, vec3<f32>(-6.0, 0.0, 0.0), vec3<f32>(6.0, 0.0, 0.0)), builtin_lineSegment(p, vec3<f32>(0.0, -6.0, 0.0), vec3<f32>(0.0, 6.0, 0.0)), builtin_lineSegment(p, vec3<f32>(0.0, 0.0, -6.0), vec3<f32>(0.0, 0.0, 6.0))));
}

fn gsl_remap_11(p: vec3<f32>) -> f32 {
  return builtin_unionSharp_3(vec3<f32>(builtin_lineSegment(p, vec3<f32>(6.0, 6.0, 6.0), vec3<f32>(0.0, 6.0, 6.0)), builtin_lineSegment(p, vec3<f32>(6.0, 6.0, 6.0), vec3<f32>(6.0, 0.0, 6.0)), builtin_lineSegment(p, vec3<f32>(6.0, 6.0, 6.0), vec3<f32>(6.0, 6.0, 0.0))));
}

fn map_repeatCentered_2(x: vec2<f32>, period: vec2<f32>) -> vec2<f32> {
  let half = period * 0.5;
  return gsl_floored_vec2((gsl_floored_vec2((x + half), period) + period), period) - half;
}

fn lattice_solidify(field: f32, bias: f32) -> f32 {
  return field - bias;
}

fn map_repeatCentered(x: vec3<f32>, period: vec3<f32>) -> vec3<f32> {
  let half = period * 0.5;
  return gsl_floored_vec3((gsl_floored_vec3((x + half), period) + period), period) - half;
}

fn specimen(p: vec3<f32>, field: f32, off: f32) -> f32 {
  return max(lattice_solidify(field, 0.6), builtin_boxCentered(p, vec3<f32>(off, 0.0, 0.0), vec3<f32>(28.0, 28.0, 28.0)));
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  gsl_partNear = vec3<f32>(1e10, 1e10, 1e10);
  let gsl_partLo = vec3<f32>(-235.0, -14.0, -14.0);
  let gsl_partHi = vec3<f32>(-207.0, 14.0, 14.0);
  let gsl_partBox = gsl_boxDistance(p, gsl_partLo, gsl_partHi);
  var gsl_part = gsl_partBox;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox <= gsl_vanishView.w) {
    let gsl_partVanish = gsl_vanish(gsl_partLo, gsl_partHi);
    gsl_part = 1e10;
    if (gsl_partVanish.x > 0.0) { gsl_part = specimen(p, gsl_remap(abs(map_repeatCentered(p, vec3<f32>(12.0, 12.0, 12.0)))), -221.0) + gsl_partVanish.y; }
    if (abs(gsl_part) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part), gsl_partVanish.zw); }
  }
  let oct = gsl_part;
  let gsl_partLo_2 = vec3<f32>(-201.0, -14.0, -14.0);
  let gsl_partHi_2 = vec3<f32>(-173.0, 14.0, 14.0);
  let gsl_partBox_2 = gsl_boxDistance(p, gsl_partLo_2, gsl_partHi_2);
  var gsl_part_2 = gsl_partBox_2;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_2 <= gsl_vanishView.w) {
    let gsl_partVanish_2 = gsl_vanish(gsl_partLo_2, gsl_partHi_2);
    gsl_part_2 = 1e10;
    if (gsl_partVanish_2.x > 0.0) { gsl_part_2 = specimen(p, gsl_remap_2(abs(map_repeatCentered(p, vec3<f32>(12.0, 12.0, 12.0)))), -187.0) + gsl_partVanish_2.y; }
    if (abs(gsl_part_2) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_2), gsl_partVanish_2.zw); }
  }
  let toc = gsl_part_2;
  let gsl_partLo_3 = vec3<f32>(-167.0, -14.0, -14.0);
  let gsl_partHi_3 = vec3<f32>(-139.0, 14.0, 14.0);
  let gsl_partBox_3 = gsl_boxDistance(p, gsl_partLo_3, gsl_partHi_3);
  var gsl_part_3 = gsl_partBox_3;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_3 <= gsl_vanishView.w) {
    let gsl_partVanish_3 = gsl_vanish(gsl_partLo_3, gsl_partHi_3);
    gsl_part_3 = 1e10;
    if (gsl_partVanish_3.x > 0.0) { gsl_part_3 = specimen(p, builtin_lineSegment(abs(map_repeatCentered(p, vec3<f32>(12.0, 12.0, 12.0))), vec3<f32>(6.0, 6.0, 6.0), vec3<f32>(-6.0, -6.0, -6.0)), -153.0) + gsl_partVanish_3.y; }
    if (abs(gsl_part_3) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_3), gsl_partVanish_3.zw); }
  }
  let bcc = gsl_part_3;
  let gsl_partLo_4 = vec3<f32>(-133.0, -14.0, -14.0);
  let gsl_partHi_4 = vec3<f32>(-105.0, 14.0, 14.0);
  let gsl_partBox_4 = gsl_boxDistance(p, gsl_partLo_4, gsl_partHi_4);
  var gsl_part_4 = gsl_partBox_4;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_4 <= gsl_vanishView.w) {
    let gsl_partVanish_4 = gsl_vanish(gsl_partLo_4, gsl_partHi_4);
    gsl_part_4 = 1e10;
    if (gsl_partVanish_4.x > 0.0) { gsl_part_4 = specimen(p, gsl_remap_3(abs(map_repeatCentered(p, vec3<f32>(12.0, 12.0, 12.0)))), -119.0) + gsl_partVanish_4.y; }
    if (abs(gsl_part_4) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_4), gsl_partVanish_4.zw); }
  }
  let fcc = gsl_part_4;
  let gsl_partLo_5 = vec3<f32>(-99.0, -14.0, -14.0);
  let gsl_partHi_5 = vec3<f32>(-71.0, 14.0, 14.0);
  let gsl_partBox_5 = gsl_boxDistance(p, gsl_partLo_5, gsl_partHi_5);
  var gsl_part_5 = gsl_partBox_5;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_5 <= gsl_vanishView.w) {
    let gsl_partVanish_5 = gsl_vanish(gsl_partLo_5, gsl_partHi_5);
    gsl_part_5 = 1e10;
    if (gsl_partVanish_5.x > 0.0) { gsl_part_5 = specimen(p, gsl_remap_4(abs(map_repeatCentered(p, vec3<f32>(12.0, 12.0, 12.0)))), -85.0) + gsl_partVanish_5.y; }
    if (abs(gsl_part_5) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_5), gsl_partVanish_5.zw); }
  }
  let tcu = gsl_part_5;
  let gsl_partLo_6 = vec3<f32>(-65.0, -14.0, -14.0);
  let gsl_partHi_6 = vec3<f32>(-37.0, 14.0, 14.0);
  let gsl_partBox_6 = gsl_boxDistance(p, gsl_partLo_6, gsl_partHi_6);
  var gsl_part_6 = gsl_partBox_6;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_6 <= gsl_vanishView.w) {
    let gsl_partVanish_6 = gsl_vanish(gsl_partLo_6, gsl_partHi_6);
    gsl_part_6 = 1e10;
    if (gsl_partVanish_6.x > 0.0) { gsl_part_6 = specimen(p, gsl_remap_5(abs(map_repeatCentered(p, vec3<f32>(12.0, 12.0, 12.0)))), -51.0) + gsl_partVanish_6.y; }
    if (abs(gsl_part_6) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_6), gsl_partVanish_6.zw); }
  }
  let kel = gsl_part_6;
  let gsl_partLo_7 = vec3<f32>(-31.0, -14.0, -14.0);
  let gsl_partHi_7 = vec3<f32>(-3.0, 14.0, 14.0);
  let gsl_partBox_7 = gsl_boxDistance(p, gsl_partLo_7, gsl_partHi_7);
  var gsl_part_7 = gsl_partBox_7;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_7 <= gsl_vanishView.w) {
    let gsl_partVanish_7 = gsl_vanish(gsl_partLo_7, gsl_partHi_7);
    gsl_part_7 = 1e10;
    if (gsl_partVanish_7.x > 0.0) { gsl_part_7 = specimen(p, gsl_remap_6(abs(map_repeatCentered(p, vec3<f32>(12.0, 12.0, 12.0)))), -17.0) + gsl_partVanish_7.y; }
    if (abs(gsl_part_7) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_7), gsl_partVanish_7.zw); }
  }
  let flu = gsl_part_7;
  let gsl_partLo_8 = vec3<f32>(3.0, -14.0, -14.0);
  let gsl_partHi_8 = vec3<f32>(31.0, 14.0, 14.0);
  let gsl_partBox_8 = gsl_boxDistance(p, gsl_partLo_8, gsl_partHi_8);
  var gsl_part_8 = gsl_partBox_8;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_8 <= gsl_vanishView.w) {
    let gsl_partVanish_8 = gsl_vanish(gsl_partLo_8, gsl_partHi_8);
    gsl_part_8 = 1e10;
    if (gsl_partVanish_8.x > 0.0) { gsl_part_8 = specimen(p, builtin_unionSharp_5(vec2<f32>(gsl_remap_7(map_repeatCentered(p, vec3<f32>(12.0, 12.0, 12.0))), gsl_remap_8(abs(map_repeatCentered(p, vec3<f32>(12.0, 12.0, 12.0)))))), 17.0) + gsl_partVanish_8.y; }
    if (abs(gsl_part_8) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_8), gsl_partVanish_8.zw); }
  }
  let iso = gsl_part_8;
  let gsl_partLo_9 = vec3<f32>(37.0, -14.0, -14.0);
  let gsl_partHi_9 = vec3<f32>(65.0, 14.0, 14.0);
  let gsl_partBox_9 = gsl_boxDistance(p, gsl_partLo_9, gsl_partHi_9);
  var gsl_part_9 = gsl_partBox_9;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_9 <= gsl_vanishView.w) {
    let gsl_partVanish_9 = gsl_vanish(gsl_partLo_9, gsl_partHi_9);
    gsl_part_9 = 1e10;
    if (gsl_partVanish_9.x > 0.0) { gsl_part_9 = specimen(p, builtin_unionSharp_5(vec2<f32>(gsl_remap_9(abs(map_repeatCentered(p, vec3<f32>(12.0, 12.0, 12.0)))), builtin_lineSegment(map_repeatCentered(p, vec3<f32>(12.0, 12.0, 12.0)), vec3<f32>(0.0, 0.0, -4.8), vec3<f32>(0.0, 0.0, 4.8)))), 51.0) + gsl_partVanish_9.y; }
    if (abs(gsl_part_9) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_9), gsl_partVanish_9.zw); }
  }
  let ree = gsl_part_9;
  let gsl_partLo_10 = vec3<f32>(71.0, -14.0, -14.0);
  let gsl_partHi_10 = vec3<f32>(99.0, 14.0, 14.0);
  let gsl_partBox_10 = gsl_boxDistance(p, gsl_partLo_10, gsl_partHi_10);
  var gsl_part_10 = gsl_partBox_10;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_10 <= gsl_vanishView.w) {
    let gsl_partVanish_10 = gsl_vanish(gsl_partLo_10, gsl_partHi_10);
    gsl_part_10 = 1e10;
    if (gsl_partVanish_10.x > 0.0) { gsl_part_10 = specimen(p, gsl_remap_10(map_repeatCentered(p, vec3<f32>(12.0, 12.0, 12.0))), 85.0) + gsl_partVanish_10.y; }
    if (abs(gsl_part_10) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_10), gsl_partVanish_10.zw); }
  }
  let cub = gsl_part_10;
  let gsl_partLo_11 = vec3<f32>(105.0, -14.0, -14.0);
  let gsl_partHi_11 = vec3<f32>(133.0, 14.0, 14.0);
  let gsl_partBox_11 = gsl_boxDistance(p, gsl_partLo_11, gsl_partHi_11);
  var gsl_part_11 = gsl_partBox_11;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_11 <= gsl_vanishView.w) {
    let gsl_partVanish_11 = gsl_vanish(gsl_partLo_11, gsl_partHi_11);
    gsl_part_11 = 1e10;
    if (gsl_partVanish_11.x > 0.0) { gsl_part_11 = specimen(p, min(gsl_remap_7(map_repeatCentered(p, vec3<f32>(12.0, 12.0, 12.0))), gsl_remap_11(abs(map_repeatCentered(p, vec3<f32>(12.0, 12.0, 12.0))))), 119.0) + gsl_partVanish_11.y; }
    if (abs(gsl_part_11) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_11), gsl_partVanish_11.zw); }
  }
  let cud = gsl_part_11;
  let gsl_partLo_12 = vec3<f32>(139.0, -14.0, -14.0);
  let gsl_partHi_12 = vec3<f32>(167.0, 14.0, 14.0);
  let gsl_partBox_12 = gsl_boxDistance(p, gsl_partLo_12, gsl_partHi_12);
  var gsl_part_12 = gsl_partBox_12;
  let gsl_tmp_4 = map_repeatCentered_2(p.xy, vec2<f32>(12.0, 12.0));
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_12 <= gsl_vanishView.w) {
    let gsl_partVanish_12 = gsl_vanish(gsl_partLo_12, gsl_partHi_12);
    gsl_part_12 = 1e10;
    if (gsl_partVanish_12.x > 0.0) { gsl_part_12 = specimen(p, length(vec3<f32>(gsl_tmp_4.x, gsl_tmp_4.y, p.z).xy), 153.0) + gsl_partVanish_12.y; }
    if (abs(gsl_part_12) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_12), gsl_partVanish_12.zw); }
  }
  let colR = gsl_part_12;
  let gsl_partLo_13 = vec3<f32>(173.0, -14.0, -14.0);
  let gsl_partHi_13 = vec3<f32>(201.0, 14.0, 14.0);
  let gsl_partBox_13 = gsl_boxDistance(p, gsl_partLo_13, gsl_partHi_13);
  var gsl_part_13 = gsl_partBox_13;
  let gsl_tmp_6 = map_repeatCentered_2(p.xy - vec2<f32>(6.0, 10.392304), vec2<f32>(12.0, 20.784609));
  let gsl_tmp_5 = map_repeatCentered_2(p.xy, vec2<f32>(12.0, 20.784609));
  let gsl_tmp_7 = select(gsl_tmp_6, gsl_tmp_5, dot(gsl_tmp_5, gsl_tmp_5) < dot(gsl_tmp_6, gsl_tmp_6));
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_13 <= gsl_vanishView.w) {
    let gsl_partVanish_13 = gsl_vanish(gsl_partLo_13, gsl_partHi_13);
    gsl_part_13 = 1e10;
    if (gsl_partVanish_13.x > 0.0) { gsl_part_13 = specimen(p, length(vec3<f32>(gsl_tmp_7.x, gsl_tmp_7.y, p.z).xy), 187.0) + gsl_partVanish_13.y; }
    if (abs(gsl_part_13) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_13), gsl_partVanish_13.zw); }
  }
  let colT = gsl_part_13;
  let gsl_partLo_14 = vec3<f32>(207.0, -14.0, -14.0);
  let gsl_partHi_14 = vec3<f32>(235.0, 14.0, 14.0);
  let gsl_partBox_14 = gsl_boxDistance(p, gsl_partLo_14, gsl_partHi_14);
  var gsl_part_14 = gsl_partBox_14;
  let gsl_tmp_11 = map_repeatCentered_2(p.xy - vec2<f32>(6.928203, 20.0), vec2<f32>(13.856406, 24.0));
  let gsl_tmp_10 = map_repeatCentered_2(p.xy - vec2<f32>(6.928203, 12.0), vec2<f32>(13.856406, 24.0));
  let gsl_tmp_13 = select(gsl_tmp_11, gsl_tmp_10, dot(gsl_tmp_10, gsl_tmp_10) < dot(gsl_tmp_11, gsl_tmp_11));
  let gsl_tmp_9 = map_repeatCentered_2(p.xy - vec2<f32>(0.0, 8.0), vec2<f32>(13.856406, 24.0));
  let gsl_tmp_8 = map_repeatCentered_2(p.xy, vec2<f32>(13.856406, 24.0));
  let gsl_tmp_12 = select(gsl_tmp_9, gsl_tmp_8, dot(gsl_tmp_8, gsl_tmp_8) < dot(gsl_tmp_9, gsl_tmp_9));
  let gsl_tmp_14 = select(gsl_tmp_13, gsl_tmp_12, dot(gsl_tmp_12, gsl_tmp_12) < dot(gsl_tmp_13, gsl_tmp_13));
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_14 <= gsl_vanishView.w) {
    let gsl_partVanish_14 = gsl_vanish(gsl_partLo_14, gsl_partHi_14);
    gsl_part_14 = 1e10;
    if (gsl_partVanish_14.x > 0.0) { gsl_part_14 = specimen(p, length(vec3<f32>(gsl_tmp_14.x, gsl_tmp_14.y, p.z).xy), 221.0) + gsl_partVanish_14.y; }
    if (abs(gsl_part_14) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_14), gsl_partVanish_14.zw); }
  }
  let colH = gsl_part_14;
  let gsl_s1 = gsl_shape(oct, 1u);
  let gsl_s2 = gsl_shape(toc, 2u);
  let gsl_s3 = gsl_shape(bcc, 3u);
  let gsl_s4 = gsl_shape(fcc, 4u);
  let gsl_s5 = gsl_shape(tcu, 5u);
  let gsl_s6 = gsl_shape(kel, 6u);
  let gsl_s7 = gsl_shape(flu, 7u);
  let gsl_s8 = gsl_shape(iso, 8u);
  let gsl_s9 = gsl_shape(ree, 9u);
  let gsl_s10 = gsl_shape(cub, 10u);
  let gsl_s11 = gsl_shape(cud, 11u);
  let gsl_s12 = gsl_shape(colR, 12u);
  let gsl_s13 = gsl_shape(colT, 13u);
  let gsl_s14 = gsl_shape(colH, 14u);
  return gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_s1, gsl_s2), gsl_s3), gsl_s4), gsl_s5), gsl_s6), gsl_s7), gsl_s8), gsl_s9), gsl_s10), gsl_s11), gsl_s12), gsl_s13), gsl_s14);
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
