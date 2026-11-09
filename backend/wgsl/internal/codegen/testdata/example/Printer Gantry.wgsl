fn builtin_gcl_closure_box(p: vec3<f32>, pointA: vec3<f32>, pointB: vec3<f32>) -> f32 {
  let center = (pointA + pointB) * 0.5;
  let size = abs(pointB - pointA);
  return builtin_gcl_closure_boxCentered(p, center, size);
}

fn builtin_box(p: vec3<f32>, pointA: vec3<f32>, pointB: vec3<f32>) -> f32 {
  return builtin_gcl_closure_box(p, pointA, pointB);
}

fn drill_prismVoid(p: vec3<f32>, zTop: f32, section: f32, depth: f32) -> f32 {
  return max(section, zTop - depth - p.z);
}

fn gsl_remap(p: vec3<f32>, gsl_param_33_8: f32, gsl_param_36_10: f32) -> f32 {
  let m5 = fastener_screw(max(length(vec3<f32>(p.x, p.y, p.z - -2.7272727)) - 5.4772725, -p.z), max(max(max(length(p.xy) - 2.5, -10.0 - p.z), p.z - 1.375), (length(p.xy) - 2.5 + -p.z - 10.0 + 0.75) * 0.7071), drill_prismVoid(p, 2.75, max(max(abs(p.x) - 1.5, abs(0.5 * p.x + 0.8660254 * p.y) - 1.5), abs(0.5 * p.x - 0.8660254 * p.y) - 1.5), 1.56));
  return m5;
}

fn gsl_remap_2(p: vec3<f32>, gsl_param_33_8: f32, gsl_param_36_10: f32) -> f32 {
  return gsl_remap(vec3<f32>(p.x * 0.00000000000000006123234 - p.z * -1.0, p.y, p.z * 0.00000000000000006123234 + p.x * -1.0), gsl_param_33_8, gsl_param_36_10);
}

fn gsl_remap_3(p: vec3<f32>, gsl_param_33_8: f32, gsl_param_36_10: f32) -> f32 {
  return gsl_remap(vec3<f32>(p.x * 0.00000000000000006123234 - p.z * 1.0, p.y, p.z * 0.00000000000000006123234 + p.x * 1.0), gsl_param_33_8, gsl_param_36_10);
}

fn gsl_remap_4(p: vec3<f32>, gsl_param_33_8: f32, gsl_param_36_10: f32) -> f32 {
  return gsl_remap(vec3<f32>(p.x, p.y * 0.00000000000000006123234 + p.z * 1.0, -p.y * 1.0 + p.z * 0.00000000000000006123234), gsl_param_33_8, gsl_param_36_10);
}

fn gsl_remap_5(p: vec3<f32>, gsl_param_33_8: f32, gsl_param_36_10: f32) -> f32 {
  let m5long = fastener_screw(max(max(max(length(p.xy) - 4.25, -p.z), p.z - 5.0), (length(p.xy) - 4.25 + p.z - 5.0 + 0.2) * 0.7071), max(max(max(length(p.xy) - 2.5, -30.0 - p.z), p.z - 2.5), (length(p.xy) - 2.5 + -p.z - 30.0 + 0.75) * 0.7071), drill_prismVoid(p, 5.0, max(max(abs(p.x) - 2.0, abs(0.5 * p.x + 0.8660254 * p.y) - 2.0), abs(0.5 * p.x - 0.8660254 * p.y) - 2.0), 2.5));
  return m5long;
}

fn gsl_remap_6(p: vec3<f32>, gsl_param_33_8: f32, gsl_param_36_10: f32) -> f32 {
  return gsl_remap_5(vec3<f32>(p.x * 0.00000000000000006123234 - p.z * -1.0, p.y, p.z * 0.00000000000000006123234 + p.x * -1.0), gsl_param_33_8, gsl_param_36_10);
}

fn gsl_remap_7(p: vec3<f32>, gsl_param_33_8: f32, gsl_param_36_10: f32) -> f32 {
  return gsl_remap_5(vec3<f32>(p.x * 0.00000000000000006123234 - p.z * 1.0, p.y, p.z * 0.00000000000000006123234 + p.x * 1.0), gsl_param_33_8, gsl_param_36_10);
}

fn builtin_intersectSharp_2(fields: vec2<f32>) -> f32 {
  return max(fields.x, fields.y);
}

fn builtin_intersectRound_2(fields: vec2<f32>, radius: f32) -> f32 {
  let sharp = builtin_intersectSharp_2(fields);
  let sumOfSquares = 0.0 + max(fields.x + radius, 0.0) * max(fields.x + radius, 0.0) + max(fields.y + radius, 0.0) * max(fields.y + radius, 0.0);
  return sharp - max(sharp + radius, 0.0) + sqrt(max(sumOfSquares, 0.000000000000000000000000000001));
}

fn builtin_gcl_closure_intersectRound_2(fields: vec2<f32>) -> f32 {
  return builtin_intersectRound_2(fields, 0.0);
}

fn builtin_gcl_closure_cylinderCapped(p: vec3<f32>, pointA: vec3<f32>, pointB: vec3<f32>, radius: f32) -> f32 {
  let pa = p - pointA;
  let ba = pointB - pointA;
  let paba = dot(pa, ba);
  let baba = dot(ba, ba);
  let t = clamp(paba / baba, 0.0, 1.0);
  let radialDistance = length(pa - ba * t) - radius;
  let axialDistance = (abs(paba - baba * 0.5) - baba * 0.5) / sqrt(baba);
  let form = builtin_gcl_closure_intersectRound_2(vec2<f32>(radialDistance, axialDistance));
  return form;
}

fn builtin_cylinderCapped(p: vec3<f32>, pointA: vec3<f32>, pointB: vec3<f32>, radius: f32) -> f32 {
  return builtin_gcl_closure_cylinderCapped(p, pointA, pointB, radius);
}

fn builtin_coneCapped(p: vec3<f32>, pointA: vec3<f32>, pointB: vec3<f32>, radiusA: f32, radiusB: f32) -> f32 {
  let papa = dot(p - pointA, p - pointA);
  let baba = dot(pointB - pointA, pointB - pointA);
  let paba = dot(p - pointA, pointB - pointA) / baba;
  let x = sqrt(papa - paba * paba * baba);
  let rba = radiusB - radiusA;
  let k = rba * rba + baba;
  let f = clamp((rba * (x - radiusA) + paba * baba) / k, 0.0, 1.0);
  let cbx = x - radiusA - f * rba;
  let cay = abs(paba - 0.5) - 0.5;
  let s = select(1.0, -1.0, cbx < 0.0 && cay < 0.0);
  let cax = max(0.0, x - select(radiusB, radiusA, paba < 0.5));
  let cby = paba - f;
  return s * sqrt(min(cax * cax + cay * cay * baba, cbx * cbx + cby * cby * baba));
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

fn fastener_screw(head: f32, shank: f32, drive: f32) -> f32 {
  return max(min(head, shank), -drive);
}

fn slot(u: f32, v: f32) -> f32 {
  let opening = max(abs(u - 9.35) - 1.15, abs(v) - 3.1);
  let cavity = max(abs(u - 6.05) - 2.15, abs(v) - 5.5);
  return min(opening, cavity);
}

fn section2020(a: f32, b: f32) -> f32 {
  return max(max(max(max(abs(a) - 10.0, abs(b) - 10.0), -slot(abs(a), b)), -slot(abs(b), a)), 2.1 - length(vec2<f32>(a, b)));
}

fn beam2020(a: f32, b: f32, t: f32, span: f32) -> f32 {
  return max(section2020(a, b), abs(t - span / 2.0) - span / 2.0);
}

fn section2040(a: f32, b: f32) -> f32 {
  let c = abs(b) - 10.0;
  return max(max(max(max(abs(a) - 10.0, abs(b) - 20.0), -slot(abs(a), c)), -slot(abs(b) - 10.0, a)), 2.1 - length(vec2<f32>(a, c)));
}

fn beam2040(a: f32, b: f32, t: f32, span: f32) -> f32 {
  return max(section2040(a, b), abs(t - span / 2.0) - span / 2.0);
}

fn sdShape(p: vec3<f32>, time: f32, gsl_param_33_8: f32, gsl_param_36_10: f32) -> gsl_shape {
  gsl_partNear = vec3<f32>(1e10, 1e10, 1e10);
  let lift = gsl_param_33_8;
  let gz = -200.0 + lift;
  let gsl_partLo = vec3<f32>(-210.0, -210.0, min(-200.0, -(-0.0 - gz) - 10.0));
  let gsl_partHi = vec3<f32>(210.0, 210.0, max(200.0, -(-0.0 - gz) + 10.0));
  let gsl_partBox = gsl_boxDistance(p, gsl_partLo, gsl_partHi);
  var gsl_part = gsl_partBox;
  let gsl_partLo_2 = vec3<f32>(-210.0, -210.0, -200.0);
  let gsl_partHi_2 = vec3<f32>(-190.0, 210.0, -180.0);
  let gsl_partBox_2 = gsl_boxDistance(p, gsl_partLo_2, gsl_partHi_2);
  var gsl_part_2 = gsl_partBox_2;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_2 <= gsl_vanishView.w) {
    let gsl_partVanish_2 = gsl_vanish(gsl_partLo_2, gsl_partHi_2);
    gsl_part_2 = 1e10;
    if (gsl_partVanish_2.x > 0.0) { gsl_part_2 = beam2020(p.x + 200.0, p.z - -190.0, p.y + 210.0, 420.0) + gsl_partVanish_2.y; }
    if (abs(gsl_part_2) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_2), gsl_partVanish_2.zw); }
  }
  let leftRail = gsl_part_2;
  let gsl_partLo_3 = vec3<f32>(190.0, -210.0, -200.0);
  let gsl_partHi_3 = vec3<f32>(210.0, 210.0, -180.0);
  let gsl_partBox_3 = gsl_boxDistance(p, gsl_partLo_3, gsl_partHi_3);
  var gsl_part_3 = gsl_partBox_3;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_3 <= gsl_vanishView.w) {
    let gsl_partVanish_3 = gsl_vanish(gsl_partLo_3, gsl_partHi_3);
    gsl_part_3 = 1e10;
    if (gsl_partVanish_3.x > 0.0) { gsl_part_3 = beam2020(p.x - 200.0, p.z - -190.0, p.y + 210.0, 420.0) + gsl_partVanish_3.y; }
    if (abs(gsl_part_3) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_3), gsl_partVanish_3.zw); }
  }
  let rightRail = gsl_part_3;
  let gsl_partLo_4 = vec3<f32>(-190.0, -210.0, -200.0);
  let gsl_partHi_4 = vec3<f32>(190.0, -190.0, -180.0);
  let gsl_partBox_4 = gsl_boxDistance(p, gsl_partLo_4, gsl_partHi_4);
  var gsl_part_4 = gsl_partBox_4;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_4 <= gsl_vanishView.w) {
    let gsl_partVanish_4 = gsl_vanish(gsl_partLo_4, gsl_partHi_4);
    gsl_part_4 = 1e10;
    if (gsl_partVanish_4.x > 0.0) { gsl_part_4 = beam2020(p.y + 200.0, p.z - -190.0, p.x + 190.0, 380.0) + gsl_partVanish_4.y; }
    if (abs(gsl_part_4) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_4), gsl_partVanish_4.zw); }
  }
  let frontBeam = gsl_part_4;
  let gsl_partLo_5 = vec3<f32>(-190.0, 190.0, -200.0);
  let gsl_partHi_5 = vec3<f32>(190.0, 210.0, -180.0);
  let gsl_partBox_5 = gsl_boxDistance(p, gsl_partLo_5, gsl_partHi_5);
  var gsl_part_5 = gsl_partBox_5;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_5 <= gsl_vanishView.w) {
    let gsl_partVanish_5 = gsl_vanish(gsl_partLo_5, gsl_partHi_5);
    gsl_part_5 = 1e10;
    if (gsl_partVanish_5.x > 0.0) { gsl_part_5 = beam2020(p.y - 200.0, p.z - -190.0, p.x + 190.0, 380.0) + gsl_partVanish_5.y; }
    if (abs(gsl_part_5) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_5), gsl_partVanish_5.zw); }
  }
  let backBeam = gsl_part_5;
  let gsl_partLo_6 = vec3<f32>(-210.0, -20.0, -180.0);
  let gsl_partHi_6 = vec3<f32>(-190.0, 20.0, 180.0);
  let gsl_partBox_6 = gsl_boxDistance(p, gsl_partLo_6, gsl_partHi_6);
  var gsl_part_6 = gsl_partBox_6;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_6 <= gsl_vanishView.w) {
    let gsl_partVanish_6 = gsl_vanish(gsl_partLo_6, gsl_partHi_6);
    gsl_part_6 = 1e10;
    if (gsl_partVanish_6.x > 0.0) { gsl_part_6 = beam2040(p.x + 200.0, p.y, p.z - -180.0, 360.0) + gsl_partVanish_6.y; }
    if (abs(gsl_part_6) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_6), gsl_partVanish_6.zw); }
  }
  let leftUpright = gsl_part_6;
  let gsl_partLo_7 = vec3<f32>(190.0, -20.0, -180.0);
  let gsl_partHi_7 = vec3<f32>(210.0, 20.0, 180.0);
  let gsl_partBox_7 = gsl_boxDistance(p, gsl_partLo_7, gsl_partHi_7);
  var gsl_part_7 = gsl_partBox_7;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_7 <= gsl_vanishView.w) {
    let gsl_partVanish_7 = gsl_vanish(gsl_partLo_7, gsl_partHi_7);
    gsl_part_7 = 1e10;
    if (gsl_partVanish_7.x > 0.0) { gsl_part_7 = beam2040(p.x - 200.0, p.y, p.z - -180.0, 360.0) + gsl_partVanish_7.y; }
    if (abs(gsl_part_7) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_7), gsl_partVanish_7.zw); }
  }
  let rightUpright = gsl_part_7;
  let gsl_partLo_8 = vec3<f32>(-210.0, -10.0, 180.0);
  let gsl_partHi_8 = vec3<f32>(210.0, 10.0, 200.0);
  let gsl_partBox_8 = gsl_boxDistance(p, gsl_partLo_8, gsl_partHi_8);
  var gsl_part_8 = gsl_partBox_8;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_8 <= gsl_vanishView.w) {
    let gsl_partVanish_8 = gsl_vanish(gsl_partLo_8, gsl_partHi_8);
    gsl_part_8 = 1e10;
    if (gsl_partVanish_8.x > 0.0) { gsl_part_8 = beam2020(p.y, p.z - 190.0, p.x + 210.0, 420.0) + gsl_partVanish_8.y; }
    if (abs(gsl_part_8) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_8), gsl_partVanish_8.zw); }
  }
  let topBeam = gsl_part_8;
  let gsl_partLo_9 = vec3<f32>(-190.0, -20.0, -(-0.0 - gz) - 10.0);
  let gsl_partHi_9 = vec3<f32>(190.0, 0.0, -(-0.0 - gz) + 10.0);
  let gsl_partBox_9 = gsl_boxDistance(p, gsl_partLo_9, gsl_partHi_9);
  var gsl_part_9 = gsl_partBox_9;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_9 <= gsl_vanishView.w) {
    let gsl_partVanish_9 = gsl_vanish(gsl_partLo_9, gsl_partHi_9);
    gsl_part_9 = 1e10;
    if (gsl_partVanish_9.x > 0.0) { gsl_part_9 = beam2020(p.y + 10.0, p.z - gz, p.x + 190.0, 380.0) + gsl_partVanish_9.y; }
    if (abs(gsl_part_9) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_9), gsl_partVanish_9.zw); }
  }
  let gantryBeam = gsl_part_9;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox <= gsl_vanishView.w) {
    let gsl_partVanish = gsl_vanish(gsl_partLo, gsl_partHi);
    gsl_part = 1e10;
    if (gsl_partVanish.x > 0.0) { gsl_part = min(min(min(min(min(min(min(leftRail, rightRail), frontBeam), backBeam), leftUpright), rightUpright), topBeam), gantryBeam) + gsl_partVanish.y; }
    if (abs(gsl_part) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part), gsl_partVanish.zw); }
  }
  let frame = gsl_part;
  let travel = gsl_param_36_10;
  let gsl_partLo_10 = vec3<f32>(min(-215.0, min(travel - 30.0, travel + 30.0)), -40.0, min(min(min(-200.0, min(gz - 30.0, gz + 30.0)), min(gz - 30.0, gz + 30.0)), min(gz - 30.0, gz + 30.0)));
  let gsl_partHi_10 = vec3<f32>(max(215.0, max(travel - 30.0, travel + 30.0)), 40.0, max(max(max(200.0, max(gz - 30.0, gz + 30.0)), max(gz - 30.0, gz + 30.0)), max(gz - 30.0, gz + 30.0)));
  let gsl_partBox_10 = gsl_boxDistance(p, gsl_partLo_10, gsl_partHi_10);
  var gsl_part_10 = gsl_partBox_10;
  let gsl_partLo_11 = vec3<f32>(-215.0, -40.0, -200.0);
  let gsl_partHi_11 = vec3<f32>(-210.0, 40.0, -115.0);
  let gsl_partBox_11 = gsl_boxDistance(p, gsl_partLo_11, gsl_partHi_11);
  var gsl_part_11 = gsl_partBox_11;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_11 <= gsl_vanishView.w) {
    let gsl_partVanish_11 = gsl_vanish(gsl_partLo_11, gsl_partHi_11);
    gsl_part_11 = 1e10;
    if (gsl_partVanish_11.x > 0.0) { gsl_part_11 = builtin_gcl_closure_boxCentered(p, vec3<f32>(-425.0, 0.0, -315.0) * 0.5, abs(vec3<f32>(5.0, 80.0, 85.0))) + gsl_partVanish_11.y; }
    if (abs(gsl_part_11) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_11), gsl_partVanish_11.zw); }
  }
  let lowPlateL = gsl_part_11;
  let gsl_partLo_12 = vec3<f32>(210.0, -40.0, -200.0);
  let gsl_partHi_12 = vec3<f32>(215.0, 40.0, -115.0);
  let gsl_partBox_12 = gsl_boxDistance(p, gsl_partLo_12, gsl_partHi_12);
  var gsl_part_12 = gsl_partBox_12;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_12 <= gsl_vanishView.w) {
    let gsl_partVanish_12 = gsl_vanish(gsl_partLo_12, gsl_partHi_12);
    gsl_part_12 = 1e10;
    if (gsl_partVanish_12.x > 0.0) { gsl_part_12 = builtin_gcl_closure_boxCentered(p, vec3<f32>(425.0, 0.0, -315.0) * 0.5, abs(vec3<f32>(5.0, 80.0, 85.0))) + gsl_partVanish_12.y; }
    if (abs(gsl_part_12) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_12), gsl_partVanish_12.zw); }
  }
  let lowPlateR = gsl_part_12;
  let gsl_partLo_13 = vec3<f32>(-215.0, -20.0, 130.0);
  let gsl_partHi_13 = vec3<f32>(-210.0, 20.0, 200.0);
  let gsl_partBox_13 = gsl_boxDistance(p, gsl_partLo_13, gsl_partHi_13);
  var gsl_part_13 = gsl_partBox_13;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_13 <= gsl_vanishView.w) {
    let gsl_partVanish_13 = gsl_vanish(gsl_partLo_13, gsl_partHi_13);
    gsl_part_13 = 1e10;
    if (gsl_partVanish_13.x > 0.0) { gsl_part_13 = builtin_gcl_closure_boxCentered(p, vec3<f32>(-425.0, 0.0, 330.0) * 0.5, abs(vec3<f32>(5.0, 40.0, 70.0))) + gsl_partVanish_13.y; }
    if (abs(gsl_part_13) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_13), gsl_partVanish_13.zw); }
  }
  let topPlateL = gsl_part_13;
  let gsl_partLo_14 = vec3<f32>(210.0, -20.0, 130.0);
  let gsl_partHi_14 = vec3<f32>(215.0, 20.0, 200.0);
  let gsl_partBox_14 = gsl_boxDistance(p, gsl_partLo_14, gsl_partHi_14);
  var gsl_part_14 = gsl_partBox_14;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_14 <= gsl_vanishView.w) {
    let gsl_partVanish_14 = gsl_vanish(gsl_partLo_14, gsl_partHi_14);
    gsl_part_14 = 1e10;
    if (gsl_partVanish_14.x > 0.0) { gsl_part_14 = builtin_gcl_closure_boxCentered(p, vec3<f32>(425.0, 0.0, 330.0) * 0.5, abs(vec3<f32>(5.0, 40.0, 70.0))) + gsl_partVanish_14.y; }
    if (abs(gsl_part_14) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_14), gsl_partVanish_14.zw); }
  }
  let topPlateR = gsl_part_14;
  let gsl_partLo_15 = vec3<f32>(-210.0, -25.0, min(gz - 30.0, gz + 30.0));
  let gsl_partHi_15 = vec3<f32>(-150.0, -20.0, max(gz - 30.0, gz + 30.0));
  let gsl_partBox_15 = gsl_boxDistance(p, gsl_partLo_15, gsl_partHi_15);
  var gsl_part_15 = gsl_partBox_15;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_15 <= gsl_vanishView.w) {
    let gsl_partVanish_15 = gsl_vanish(gsl_partLo_15, gsl_partHi_15);
    gsl_part_15 = 1e10;
    if (gsl_partVanish_15.x > 0.0) { gsl_part_15 = builtin_box(p, vec3<f32>(-210.0, -25.0, gz - 30.0), vec3<f32>(-150.0, -20.0, gz + 30.0)) + gsl_partVanish_15.y; }
    if (abs(gsl_part_15) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_15), gsl_partVanish_15.zw); }
  }
  let gantryPlateL = gsl_part_15;
  let gsl_partLo_16 = vec3<f32>(150.0, -25.0, min(gz - 30.0, gz + 30.0));
  let gsl_partHi_16 = vec3<f32>(210.0, -20.0, max(gz - 30.0, gz + 30.0));
  let gsl_partBox_16 = gsl_boxDistance(p, gsl_partLo_16, gsl_partHi_16);
  var gsl_part_16 = gsl_partBox_16;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_16 <= gsl_vanishView.w) {
    let gsl_partVanish_16 = gsl_vanish(gsl_partLo_16, gsl_partHi_16);
    gsl_part_16 = 1e10;
    if (gsl_partVanish_16.x > 0.0) { gsl_part_16 = builtin_box(p, vec3<f32>(150.0, -25.0, gz - 30.0), vec3<f32>(210.0, -20.0, gz + 30.0)) + gsl_partVanish_16.y; }
    if (abs(gsl_part_16) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_16), gsl_partVanish_16.zw); }
  }
  let gantryPlateR = gsl_part_16;
  let gsl_partLo_17 = vec3<f32>(min(travel - 30.0, travel + 30.0), -26.0, min(gz - 30.0, gz + 30.0));
  let gsl_partHi_17 = vec3<f32>(max(travel - 30.0, travel + 30.0), -20.0, max(gz - 30.0, gz + 30.0));
  let gsl_partBox_17 = gsl_boxDistance(p, gsl_partLo_17, gsl_partHi_17);
  var gsl_part_17 = gsl_partBox_17;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_17 <= gsl_vanishView.w) {
    let gsl_partVanish_17 = gsl_vanish(gsl_partLo_17, gsl_partHi_17);
    gsl_part_17 = 1e10;
    if (gsl_partVanish_17.x > 0.0) { gsl_part_17 = builtin_box(p, vec3<f32>(travel - 30.0, -26.0, gz - 30.0), vec3<f32>(travel + 30.0, -20.0, gz + 30.0)) + gsl_partVanish_17.y; }
    if (abs(gsl_part_17) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_17), gsl_partVanish_17.zw); }
  }
  let carriagePlate = gsl_part_17;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_10 <= gsl_vanishView.w) {
    let gsl_partVanish_10 = gsl_vanish(gsl_partLo_10, gsl_partHi_10);
    gsl_part_10 = 1e10;
    if (gsl_partVanish_10.x > 0.0) { gsl_part_10 = min(min(min(min(min(min(lowPlateL, lowPlateR), topPlateL), topPlateR), gantryPlateL), gantryPlateR), carriagePlate) + gsl_partVanish_10.y; }
    if (abs(gsl_part_10) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_10), gsl_partVanish_10.zw); }
  }
  let plates = gsl_part_10;
  let gsl_partLo_18 = vec3<f32>(min(min(-217.76, min(min(min(-4.76 + travel - 22.0, -4.76 + travel + 22.0), -4.76 + travel - 22.0), -4.76 + travel + 22.0)), -215.01), -204.26, min(min(min(-194.76, min(min(min(min(min(min(min(-4.76 + gz - 20.0, -4.76 + gz + 20.0), -4.76 + gz), -4.76 + gz), -4.76 + gz - 20.0), -4.76 + gz + 20.0), -4.76 + gz), -4.76 + gz)), min(min(min(-4.76 + gz - 22.0, -4.76 + gz - 22.0), -4.76 + gz + 22.0), -4.76 + gz + 22.0)), -194.26));
  let gsl_partHi_18 = vec3<f32>(max(max(217.76, max(max(max(4.76 + travel - 22.0, 4.76 + travel + 22.0), 4.76 + travel - 22.0), 4.76 + travel + 22.0)), 215.01), 204.26, max(max(max(194.76, max(max(max(max(max(max(max(4.76 + gz - 20.0, 4.76 + gz + 20.0), 4.76 + gz), 4.76 + gz), 4.76 + gz - 20.0), 4.76 + gz + 20.0), 4.76 + gz), 4.76 + gz)), max(max(max(4.76 + gz - 22.0, 4.76 + gz - 22.0), 4.76 + gz + 22.0), 4.76 + gz + 22.0)), -185.74));
  let gsl_partBox_18 = gsl_boxDistance(p, gsl_partLo_18, gsl_partHi_18);
  var gsl_part_18 = gsl_partBox_18;
  let gsl_partLo_19 = vec3<f32>(-217.76, -29.76, -194.76);
  let gsl_partHi_19 = vec3<f32>(-204.99, 29.76, -125.24);
  let gsl_partBox_19 = gsl_boxDistance(p, gsl_partLo_19, gsl_partHi_19);
  var gsl_part_19 = gsl_partBox_19;
  let gsl_partLo_20 = vec3<f32>(-217.76, -29.76, -194.76);
  let gsl_partHi_20 = vec3<f32>(-204.99, -20.24, -185.24);
  let gsl_partBox_20 = gsl_boxDistance(p, gsl_partLo_20, gsl_partHi_20);
  var gsl_part_20 = gsl_partBox_20;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_20 <= gsl_vanishView.w) {
    let gsl_partVanish_20 = gsl_vanish(gsl_partLo_20, gsl_partHi_20);
    gsl_part_20 = 1e10;
    if (gsl_partVanish_20.x > 0.0) { gsl_part_20 = gsl_remap_2(p - vec3<f32>(-215.0, -25.0, -190.0), gsl_param_33_8, gsl_param_36_10) + gsl_partVanish_20.y; }
    if (abs(gsl_part_20) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_20), gsl_partVanish_20.zw); }
  }
  let lowL1 = gsl_part_20;
  let gsl_partLo_21 = vec3<f32>(-217.76, 20.24, -194.76);
  let gsl_partHi_21 = vec3<f32>(-204.99, 29.76, -185.24);
  let gsl_partBox_21 = gsl_boxDistance(p, gsl_partLo_21, gsl_partHi_21);
  var gsl_part_21 = gsl_partBox_21;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_21 <= gsl_vanishView.w) {
    let gsl_partVanish_21 = gsl_vanish(gsl_partLo_21, gsl_partHi_21);
    gsl_part_21 = 1e10;
    if (gsl_partVanish_21.x > 0.0) { gsl_part_21 = gsl_remap_2(p - vec3<f32>(-215.0, 25.0, -190.0), gsl_param_33_8, gsl_param_36_10) + gsl_partVanish_21.y; }
    if (abs(gsl_part_21) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_21), gsl_partVanish_21.zw); }
  }
  let lowL2 = gsl_part_21;
  let gsl_partLo_22 = vec3<f32>(-217.76, -14.76, -159.76);
  let gsl_partHi_22 = vec3<f32>(-204.99, -5.24, -150.24);
  let gsl_partBox_22 = gsl_boxDistance(p, gsl_partLo_22, gsl_partHi_22);
  var gsl_part_22 = gsl_partBox_22;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_22 <= gsl_vanishView.w) {
    let gsl_partVanish_22 = gsl_vanish(gsl_partLo_22, gsl_partHi_22);
    gsl_part_22 = 1e10;
    if (gsl_partVanish_22.x > 0.0) { gsl_part_22 = gsl_remap_2(p - vec3<f32>(-215.0, -10.0, -155.0), gsl_param_33_8, gsl_param_36_10) + gsl_partVanish_22.y; }
    if (abs(gsl_part_22) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_22), gsl_partVanish_22.zw); }
  }
  let lowL3 = gsl_part_22;
  let gsl_partLo_23 = vec3<f32>(-217.76, 5.24, -159.76);
  let gsl_partHi_23 = vec3<f32>(-204.99, 14.76, -150.24);
  let gsl_partBox_23 = gsl_boxDistance(p, gsl_partLo_23, gsl_partHi_23);
  var gsl_part_23 = gsl_partBox_23;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_23 <= gsl_vanishView.w) {
    let gsl_partVanish_23 = gsl_vanish(gsl_partLo_23, gsl_partHi_23);
    gsl_part_23 = 1e10;
    if (gsl_partVanish_23.x > 0.0) { gsl_part_23 = gsl_remap_2(p - vec3<f32>(-215.0, 10.0, -155.0), gsl_param_33_8, gsl_param_36_10) + gsl_partVanish_23.y; }
    if (abs(gsl_part_23) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_23), gsl_partVanish_23.zw); }
  }
  let lowL4 = gsl_part_23;
  let gsl_partLo_24 = vec3<f32>(-217.76, -14.76, -134.76);
  let gsl_partHi_24 = vec3<f32>(-204.99, -5.24, -125.24);
  let gsl_partBox_24 = gsl_boxDistance(p, gsl_partLo_24, gsl_partHi_24);
  var gsl_part_24 = gsl_partBox_24;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_24 <= gsl_vanishView.w) {
    let gsl_partVanish_24 = gsl_vanish(gsl_partLo_24, gsl_partHi_24);
    gsl_part_24 = 1e10;
    if (gsl_partVanish_24.x > 0.0) { gsl_part_24 = gsl_remap_2(p - vec3<f32>(-215.0, -10.0, -130.0), gsl_param_33_8, gsl_param_36_10) + gsl_partVanish_24.y; }
    if (abs(gsl_part_24) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_24), gsl_partVanish_24.zw); }
  }
  let lowL5 = gsl_part_24;
  let gsl_partLo_25 = vec3<f32>(-217.76, 5.24, -134.76);
  let gsl_partHi_25 = vec3<f32>(-204.99, 14.76, -125.24);
  let gsl_partBox_25 = gsl_boxDistance(p, gsl_partLo_25, gsl_partHi_25);
  var gsl_part_25 = gsl_partBox_25;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_25 <= gsl_vanishView.w) {
    let gsl_partVanish_25 = gsl_vanish(gsl_partLo_25, gsl_partHi_25);
    gsl_part_25 = 1e10;
    if (gsl_partVanish_25.x > 0.0) { gsl_part_25 = gsl_remap_2(p - vec3<f32>(-215.0, 10.0, -130.0), gsl_param_33_8, gsl_param_36_10) + gsl_partVanish_25.y; }
    if (abs(gsl_part_25) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_25), gsl_partVanish_25.zw); }
  }
  let lowL6 = gsl_part_25;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_19 <= gsl_vanishView.w) {
    let gsl_partVanish_19 = gsl_vanish(gsl_partLo_19, gsl_partHi_19);
    gsl_part_19 = 1e10;
    if (gsl_partVanish_19.x > 0.0) { gsl_part_19 = min(min(min(min(min(lowL1, lowL2), lowL3), lowL4), lowL5), lowL6) + gsl_partVanish_19.y; }
    if (abs(gsl_part_19) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_19), gsl_partVanish_19.zw); }
  }
  let lowBoltsL = gsl_part_19;
  let gsl_partLo_26 = vec3<f32>(204.99, -29.76, -194.76);
  let gsl_partHi_26 = vec3<f32>(217.76, 29.76, -125.24);
  let gsl_partBox_26 = gsl_boxDistance(p, gsl_partLo_26, gsl_partHi_26);
  var gsl_part_26 = gsl_partBox_26;
  let gsl_partLo_27 = vec3<f32>(204.99, -29.76, -194.76);
  let gsl_partHi_27 = vec3<f32>(217.76, -20.24, -185.24);
  let gsl_partBox_27 = gsl_boxDistance(p, gsl_partLo_27, gsl_partHi_27);
  var gsl_part_27 = gsl_partBox_27;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_27 <= gsl_vanishView.w) {
    let gsl_partVanish_27 = gsl_vanish(gsl_partLo_27, gsl_partHi_27);
    gsl_part_27 = 1e10;
    if (gsl_partVanish_27.x > 0.0) { gsl_part_27 = gsl_remap_3(p - vec3<f32>(215.0, -25.0, -190.0), gsl_param_33_8, gsl_param_36_10) + gsl_partVanish_27.y; }
    if (abs(gsl_part_27) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_27), gsl_partVanish_27.zw); }
  }
  let lowR1 = gsl_part_27;
  let gsl_partLo_28 = vec3<f32>(204.99, 20.24, -194.76);
  let gsl_partHi_28 = vec3<f32>(217.76, 29.76, -185.24);
  let gsl_partBox_28 = gsl_boxDistance(p, gsl_partLo_28, gsl_partHi_28);
  var gsl_part_28 = gsl_partBox_28;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_28 <= gsl_vanishView.w) {
    let gsl_partVanish_28 = gsl_vanish(gsl_partLo_28, gsl_partHi_28);
    gsl_part_28 = 1e10;
    if (gsl_partVanish_28.x > 0.0) { gsl_part_28 = gsl_remap_3(p - vec3<f32>(215.0, 25.0, -190.0), gsl_param_33_8, gsl_param_36_10) + gsl_partVanish_28.y; }
    if (abs(gsl_part_28) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_28), gsl_partVanish_28.zw); }
  }
  let lowR2 = gsl_part_28;
  let gsl_partLo_29 = vec3<f32>(204.99, -14.76, -159.76);
  let gsl_partHi_29 = vec3<f32>(217.76, -5.24, -150.24);
  let gsl_partBox_29 = gsl_boxDistance(p, gsl_partLo_29, gsl_partHi_29);
  var gsl_part_29 = gsl_partBox_29;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_29 <= gsl_vanishView.w) {
    let gsl_partVanish_29 = gsl_vanish(gsl_partLo_29, gsl_partHi_29);
    gsl_part_29 = 1e10;
    if (gsl_partVanish_29.x > 0.0) { gsl_part_29 = gsl_remap_3(p - vec3<f32>(215.0, -10.0, -155.0), gsl_param_33_8, gsl_param_36_10) + gsl_partVanish_29.y; }
    if (abs(gsl_part_29) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_29), gsl_partVanish_29.zw); }
  }
  let lowR3 = gsl_part_29;
  let gsl_partLo_30 = vec3<f32>(204.99, 5.24, -159.76);
  let gsl_partHi_30 = vec3<f32>(217.76, 14.76, -150.24);
  let gsl_partBox_30 = gsl_boxDistance(p, gsl_partLo_30, gsl_partHi_30);
  var gsl_part_30 = gsl_partBox_30;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_30 <= gsl_vanishView.w) {
    let gsl_partVanish_30 = gsl_vanish(gsl_partLo_30, gsl_partHi_30);
    gsl_part_30 = 1e10;
    if (gsl_partVanish_30.x > 0.0) { gsl_part_30 = gsl_remap_3(p - vec3<f32>(215.0, 10.0, -155.0), gsl_param_33_8, gsl_param_36_10) + gsl_partVanish_30.y; }
    if (abs(gsl_part_30) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_30), gsl_partVanish_30.zw); }
  }
  let lowR4 = gsl_part_30;
  let gsl_partLo_31 = vec3<f32>(204.99, -14.76, -134.76);
  let gsl_partHi_31 = vec3<f32>(217.76, -5.24, -125.24);
  let gsl_partBox_31 = gsl_boxDistance(p, gsl_partLo_31, gsl_partHi_31);
  var gsl_part_31 = gsl_partBox_31;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_31 <= gsl_vanishView.w) {
    let gsl_partVanish_31 = gsl_vanish(gsl_partLo_31, gsl_partHi_31);
    gsl_part_31 = 1e10;
    if (gsl_partVanish_31.x > 0.0) { gsl_part_31 = gsl_remap_3(p - vec3<f32>(215.0, -10.0, -130.0), gsl_param_33_8, gsl_param_36_10) + gsl_partVanish_31.y; }
    if (abs(gsl_part_31) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_31), gsl_partVanish_31.zw); }
  }
  let lowR5 = gsl_part_31;
  let gsl_partLo_32 = vec3<f32>(204.99, 5.24, -134.76);
  let gsl_partHi_32 = vec3<f32>(217.76, 14.76, -125.24);
  let gsl_partBox_32 = gsl_boxDistance(p, gsl_partLo_32, gsl_partHi_32);
  var gsl_part_32 = gsl_partBox_32;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_32 <= gsl_vanishView.w) {
    let gsl_partVanish_32 = gsl_vanish(gsl_partLo_32, gsl_partHi_32);
    gsl_part_32 = 1e10;
    if (gsl_partVanish_32.x > 0.0) { gsl_part_32 = gsl_remap_3(p - vec3<f32>(215.0, 10.0, -130.0), gsl_param_33_8, gsl_param_36_10) + gsl_partVanish_32.y; }
    if (abs(gsl_part_32) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_32), gsl_partVanish_32.zw); }
  }
  let lowR6 = gsl_part_32;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_26 <= gsl_vanishView.w) {
    let gsl_partVanish_26 = gsl_vanish(gsl_partLo_26, gsl_partHi_26);
    gsl_part_26 = 1e10;
    if (gsl_partVanish_26.x > 0.0) { gsl_part_26 = min(min(min(min(min(lowR1, lowR2), lowR3), lowR4), lowR5), lowR6) + gsl_partVanish_26.y; }
    if (abs(gsl_part_26) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_26), gsl_partVanish_26.zw); }
  }
  let lowBoltsR = gsl_part_26;
  let gsl_partLo_33 = vec3<f32>(-217.76, -14.76, 140.24);
  let gsl_partHi_33 = vec3<f32>(-204.99, 14.76, 194.76);
  let gsl_partBox_33 = gsl_boxDistance(p, gsl_partLo_33, gsl_partHi_33);
  var gsl_part_33 = gsl_partBox_33;
  let gsl_partLo_34 = vec3<f32>(-217.76, -14.76, 140.24);
  let gsl_partHi_34 = vec3<f32>(-204.99, -5.24, 149.76);
  let gsl_partBox_34 = gsl_boxDistance(p, gsl_partLo_34, gsl_partHi_34);
  var gsl_part_34 = gsl_partBox_34;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_34 <= gsl_vanishView.w) {
    let gsl_partVanish_34 = gsl_vanish(gsl_partLo_34, gsl_partHi_34);
    gsl_part_34 = 1e10;
    if (gsl_partVanish_34.x > 0.0) { gsl_part_34 = gsl_remap_2(p - vec3<f32>(-215.0, -10.0, 145.0), gsl_param_33_8, gsl_param_36_10) + gsl_partVanish_34.y; }
    if (abs(gsl_part_34) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_34), gsl_partVanish_34.zw); }
  }
  let topL1 = gsl_part_34;
  let gsl_partLo_35 = vec3<f32>(-217.76, 5.24, 140.24);
  let gsl_partHi_35 = vec3<f32>(-204.99, 14.76, 149.76);
  let gsl_partBox_35 = gsl_boxDistance(p, gsl_partLo_35, gsl_partHi_35);
  var gsl_part_35 = gsl_partBox_35;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_35 <= gsl_vanishView.w) {
    let gsl_partVanish_35 = gsl_vanish(gsl_partLo_35, gsl_partHi_35);
    gsl_part_35 = 1e10;
    if (gsl_partVanish_35.x > 0.0) { gsl_part_35 = gsl_remap_2(p - vec3<f32>(-215.0, 10.0, 145.0), gsl_param_33_8, gsl_param_36_10) + gsl_partVanish_35.y; }
    if (abs(gsl_part_35) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_35), gsl_partVanish_35.zw); }
  }
  let topL2 = gsl_part_35;
  let gsl_partLo_36 = vec3<f32>(-217.76, -14.76, 160.24);
  let gsl_partHi_36 = vec3<f32>(-204.99, -5.24, 169.76);
  let gsl_partBox_36 = gsl_boxDistance(p, gsl_partLo_36, gsl_partHi_36);
  var gsl_part_36 = gsl_partBox_36;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_36 <= gsl_vanishView.w) {
    let gsl_partVanish_36 = gsl_vanish(gsl_partLo_36, gsl_partHi_36);
    gsl_part_36 = 1e10;
    if (gsl_partVanish_36.x > 0.0) { gsl_part_36 = gsl_remap_2(p - vec3<f32>(-215.0, -10.0, 165.0), gsl_param_33_8, gsl_param_36_10) + gsl_partVanish_36.y; }
    if (abs(gsl_part_36) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_36), gsl_partVanish_36.zw); }
  }
  let topL3 = gsl_part_36;
  let gsl_partLo_37 = vec3<f32>(-217.76, 5.24, 160.24);
  let gsl_partHi_37 = vec3<f32>(-204.99, 14.76, 169.76);
  let gsl_partBox_37 = gsl_boxDistance(p, gsl_partLo_37, gsl_partHi_37);
  var gsl_part_37 = gsl_partBox_37;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_37 <= gsl_vanishView.w) {
    let gsl_partVanish_37 = gsl_vanish(gsl_partLo_37, gsl_partHi_37);
    gsl_part_37 = 1e10;
    if (gsl_partVanish_37.x > 0.0) { gsl_part_37 = gsl_remap_2(p - vec3<f32>(-215.0, 10.0, 165.0), gsl_param_33_8, gsl_param_36_10) + gsl_partVanish_37.y; }
    if (abs(gsl_part_37) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_37), gsl_partVanish_37.zw); }
  }
  let topL4 = gsl_part_37;
  let gsl_partLo_38 = vec3<f32>(-217.76, -4.76, 185.24);
  let gsl_partHi_38 = vec3<f32>(-204.99, 4.76, 194.76);
  let gsl_partBox_38 = gsl_boxDistance(p, gsl_partLo_38, gsl_partHi_38);
  var gsl_part_38 = gsl_partBox_38;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_38 <= gsl_vanishView.w) {
    let gsl_partVanish_38 = gsl_vanish(gsl_partLo_38, gsl_partHi_38);
    gsl_part_38 = 1e10;
    if (gsl_partVanish_38.x > 0.0) { gsl_part_38 = gsl_remap_2(p - vec3<f32>(-215.0, 0.0, 190.0), gsl_param_33_8, gsl_param_36_10) + gsl_partVanish_38.y; }
    if (abs(gsl_part_38) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_38), gsl_partVanish_38.zw); }
  }
  let topL5 = gsl_part_38;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_33 <= gsl_vanishView.w) {
    let gsl_partVanish_33 = gsl_vanish(gsl_partLo_33, gsl_partHi_33);
    gsl_part_33 = 1e10;
    if (gsl_partVanish_33.x > 0.0) { gsl_part_33 = min(min(min(min(topL1, topL2), topL3), topL4), topL5) + gsl_partVanish_33.y; }
    if (abs(gsl_part_33) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_33), gsl_partVanish_33.zw); }
  }
  let topBoltsL = gsl_part_33;
  let gsl_partLo_39 = vec3<f32>(204.99, -14.76, 140.24);
  let gsl_partHi_39 = vec3<f32>(217.76, 14.76, 194.76);
  let gsl_partBox_39 = gsl_boxDistance(p, gsl_partLo_39, gsl_partHi_39);
  var gsl_part_39 = gsl_partBox_39;
  let gsl_partLo_40 = vec3<f32>(204.99, -14.76, 140.24);
  let gsl_partHi_40 = vec3<f32>(217.76, -5.24, 149.76);
  let gsl_partBox_40 = gsl_boxDistance(p, gsl_partLo_40, gsl_partHi_40);
  var gsl_part_40 = gsl_partBox_40;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_40 <= gsl_vanishView.w) {
    let gsl_partVanish_40 = gsl_vanish(gsl_partLo_40, gsl_partHi_40);
    gsl_part_40 = 1e10;
    if (gsl_partVanish_40.x > 0.0) { gsl_part_40 = gsl_remap_3(p - vec3<f32>(215.0, -10.0, 145.0), gsl_param_33_8, gsl_param_36_10) + gsl_partVanish_40.y; }
    if (abs(gsl_part_40) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_40), gsl_partVanish_40.zw); }
  }
  let topR1 = gsl_part_40;
  let gsl_partLo_41 = vec3<f32>(204.99, 5.24, 140.24);
  let gsl_partHi_41 = vec3<f32>(217.76, 14.76, 149.76);
  let gsl_partBox_41 = gsl_boxDistance(p, gsl_partLo_41, gsl_partHi_41);
  var gsl_part_41 = gsl_partBox_41;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_41 <= gsl_vanishView.w) {
    let gsl_partVanish_41 = gsl_vanish(gsl_partLo_41, gsl_partHi_41);
    gsl_part_41 = 1e10;
    if (gsl_partVanish_41.x > 0.0) { gsl_part_41 = gsl_remap_3(p - vec3<f32>(215.0, 10.0, 145.0), gsl_param_33_8, gsl_param_36_10) + gsl_partVanish_41.y; }
    if (abs(gsl_part_41) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_41), gsl_partVanish_41.zw); }
  }
  let topR2 = gsl_part_41;
  let gsl_partLo_42 = vec3<f32>(204.99, -14.76, 160.24);
  let gsl_partHi_42 = vec3<f32>(217.76, -5.24, 169.76);
  let gsl_partBox_42 = gsl_boxDistance(p, gsl_partLo_42, gsl_partHi_42);
  var gsl_part_42 = gsl_partBox_42;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_42 <= gsl_vanishView.w) {
    let gsl_partVanish_42 = gsl_vanish(gsl_partLo_42, gsl_partHi_42);
    gsl_part_42 = 1e10;
    if (gsl_partVanish_42.x > 0.0) { gsl_part_42 = gsl_remap_3(p - vec3<f32>(215.0, -10.0, 165.0), gsl_param_33_8, gsl_param_36_10) + gsl_partVanish_42.y; }
    if (abs(gsl_part_42) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_42), gsl_partVanish_42.zw); }
  }
  let topR3 = gsl_part_42;
  let gsl_partLo_43 = vec3<f32>(204.99, 5.24, 160.24);
  let gsl_partHi_43 = vec3<f32>(217.76, 14.76, 169.76);
  let gsl_partBox_43 = gsl_boxDistance(p, gsl_partLo_43, gsl_partHi_43);
  var gsl_part_43 = gsl_partBox_43;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_43 <= gsl_vanishView.w) {
    let gsl_partVanish_43 = gsl_vanish(gsl_partLo_43, gsl_partHi_43);
    gsl_part_43 = 1e10;
    if (gsl_partVanish_43.x > 0.0) { gsl_part_43 = gsl_remap_3(p - vec3<f32>(215.0, 10.0, 165.0), gsl_param_33_8, gsl_param_36_10) + gsl_partVanish_43.y; }
    if (abs(gsl_part_43) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_43), gsl_partVanish_43.zw); }
  }
  let topR4 = gsl_part_43;
  let gsl_partLo_44 = vec3<f32>(204.99, -4.76, 185.24);
  let gsl_partHi_44 = vec3<f32>(217.76, 4.76, 194.76);
  let gsl_partBox_44 = gsl_boxDistance(p, gsl_partLo_44, gsl_partHi_44);
  var gsl_part_44 = gsl_partBox_44;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_44 <= gsl_vanishView.w) {
    let gsl_partVanish_44 = gsl_vanish(gsl_partLo_44, gsl_partHi_44);
    gsl_part_44 = 1e10;
    if (gsl_partVanish_44.x > 0.0) { gsl_part_44 = gsl_remap_3(p - vec3<f32>(215.0, 0.0, 190.0), gsl_param_33_8, gsl_param_36_10) + gsl_partVanish_44.y; }
    if (abs(gsl_part_44) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_44), gsl_partVanish_44.zw); }
  }
  let topR5 = gsl_part_44;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_39 <= gsl_vanishView.w) {
    let gsl_partVanish_39 = gsl_vanish(gsl_partLo_39, gsl_partHi_39);
    gsl_part_39 = 1e10;
    if (gsl_partVanish_39.x > 0.0) { gsl_part_39 = min(min(min(min(topR1, topR2), topR3), topR4), topR5) + gsl_partVanish_39.y; }
    if (abs(gsl_part_39) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_39), gsl_partVanish_39.zw); }
  }
  let topBoltsR = gsl_part_39;
  let gsl_partLo_45 = vec3<f32>(-204.76, -27.76, min(min(min(min(min(min(min(-4.76 + gz - 20.0, -4.76 + gz + 20.0), -4.76 + gz), -4.76 + gz), -4.76 + gz - 20.0), -4.76 + gz + 20.0), -4.76 + gz), -4.76 + gz));
  let gsl_partHi_45 = vec3<f32>(204.76, -14.99, max(max(max(max(max(max(max(4.76 + gz - 20.0, 4.76 + gz + 20.0), 4.76 + gz), 4.76 + gz), 4.76 + gz - 20.0), 4.76 + gz + 20.0), 4.76 + gz), 4.76 + gz));
  let gsl_partBox_45 = gsl_boxDistance(p, gsl_partLo_45, gsl_partHi_45);
  var gsl_part_45 = gsl_partBox_45;
  let gsl_partLo_46 = vec3<f32>(-204.76, -27.76, -4.76 + gz - 20.0);
  let gsl_partHi_46 = vec3<f32>(-195.24, -14.99, 4.76 + gz - 20.0);
  let gsl_partBox_46 = gsl_boxDistance(p, gsl_partLo_46, gsl_partHi_46);
  var gsl_part_46 = gsl_partBox_46;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_46 <= gsl_vanishView.w) {
    let gsl_partVanish_46 = gsl_vanish(gsl_partLo_46, gsl_partHi_46);
    gsl_part_46 = 1e10;
    if (gsl_partVanish_46.x > 0.0) { gsl_part_46 = gsl_remap_4(p - vec3<f32>(-200.0, -25.0, gz - 20.0), gsl_param_33_8, gsl_param_36_10) + gsl_partVanish_46.y; }
    if (abs(gsl_part_46) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_46), gsl_partVanish_46.zw); }
  }
  let gantryL1 = gsl_part_46;
  let gsl_partLo_47 = vec3<f32>(-204.76, -27.76, -4.76 + gz + 20.0);
  let gsl_partHi_47 = vec3<f32>(-195.24, -14.99, 4.76 + gz + 20.0);
  let gsl_partBox_47 = gsl_boxDistance(p, gsl_partLo_47, gsl_partHi_47);
  var gsl_part_47 = gsl_partBox_47;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_47 <= gsl_vanishView.w) {
    let gsl_partVanish_47 = gsl_vanish(gsl_partLo_47, gsl_partHi_47);
    gsl_part_47 = 1e10;
    if (gsl_partVanish_47.x > 0.0) { gsl_part_47 = gsl_remap_4(p - vec3<f32>(-200.0, -25.0, gz + 20.0), gsl_param_33_8, gsl_param_36_10) + gsl_partVanish_47.y; }
    if (abs(gsl_part_47) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_47), gsl_partVanish_47.zw); }
  }
  let gantryL2 = gsl_part_47;
  let gsl_partLo_48 = vec3<f32>(-184.76, -27.76, -4.76 + gz);
  let gsl_partHi_48 = vec3<f32>(-175.24, -14.99, 4.76 + gz);
  let gsl_partBox_48 = gsl_boxDistance(p, gsl_partLo_48, gsl_partHi_48);
  var gsl_part_48 = gsl_partBox_48;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_48 <= gsl_vanishView.w) {
    let gsl_partVanish_48 = gsl_vanish(gsl_partLo_48, gsl_partHi_48);
    gsl_part_48 = 1e10;
    if (gsl_partVanish_48.x > 0.0) { gsl_part_48 = gsl_remap_4(p - vec3<f32>(-180.0, -25.0, gz), gsl_param_33_8, gsl_param_36_10) + gsl_partVanish_48.y; }
    if (abs(gsl_part_48) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_48), gsl_partVanish_48.zw); }
  }
  let gantryL3 = gsl_part_48;
  let gsl_partLo_49 = vec3<f32>(-164.76, -27.76, -4.76 + gz);
  let gsl_partHi_49 = vec3<f32>(-155.24, -14.99, 4.76 + gz);
  let gsl_partBox_49 = gsl_boxDistance(p, gsl_partLo_49, gsl_partHi_49);
  var gsl_part_49 = gsl_partBox_49;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_49 <= gsl_vanishView.w) {
    let gsl_partVanish_49 = gsl_vanish(gsl_partLo_49, gsl_partHi_49);
    gsl_part_49 = 1e10;
    if (gsl_partVanish_49.x > 0.0) { gsl_part_49 = gsl_remap_4(p - vec3<f32>(-160.0, -25.0, gz), gsl_param_33_8, gsl_param_36_10) + gsl_partVanish_49.y; }
    if (abs(gsl_part_49) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_49), gsl_partVanish_49.zw); }
  }
  let gantryL4 = gsl_part_49;
  let gsl_partLo_50 = vec3<f32>(195.24, -27.76, -4.76 + gz - 20.0);
  let gsl_partHi_50 = vec3<f32>(204.76, -14.99, 4.76 + gz - 20.0);
  let gsl_partBox_50 = gsl_boxDistance(p, gsl_partLo_50, gsl_partHi_50);
  var gsl_part_50 = gsl_partBox_50;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_50 <= gsl_vanishView.w) {
    let gsl_partVanish_50 = gsl_vanish(gsl_partLo_50, gsl_partHi_50);
    gsl_part_50 = 1e10;
    if (gsl_partVanish_50.x > 0.0) { gsl_part_50 = gsl_remap_4(p - vec3<f32>(200.0, -25.0, gz - 20.0), gsl_param_33_8, gsl_param_36_10) + gsl_partVanish_50.y; }
    if (abs(gsl_part_50) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_50), gsl_partVanish_50.zw); }
  }
  let gantryR1 = gsl_part_50;
  let gsl_partLo_51 = vec3<f32>(195.24, -27.76, -4.76 + gz + 20.0);
  let gsl_partHi_51 = vec3<f32>(204.76, -14.99, 4.76 + gz + 20.0);
  let gsl_partBox_51 = gsl_boxDistance(p, gsl_partLo_51, gsl_partHi_51);
  var gsl_part_51 = gsl_partBox_51;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_51 <= gsl_vanishView.w) {
    let gsl_partVanish_51 = gsl_vanish(gsl_partLo_51, gsl_partHi_51);
    gsl_part_51 = 1e10;
    if (gsl_partVanish_51.x > 0.0) { gsl_part_51 = gsl_remap_4(p - vec3<f32>(200.0, -25.0, gz + 20.0), gsl_param_33_8, gsl_param_36_10) + gsl_partVanish_51.y; }
    if (abs(gsl_part_51) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_51), gsl_partVanish_51.zw); }
  }
  let gantryR2 = gsl_part_51;
  let gsl_partLo_52 = vec3<f32>(175.24, -27.76, -4.76 + gz);
  let gsl_partHi_52 = vec3<f32>(184.76, -14.99, 4.76 + gz);
  let gsl_partBox_52 = gsl_boxDistance(p, gsl_partLo_52, gsl_partHi_52);
  var gsl_part_52 = gsl_partBox_52;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_52 <= gsl_vanishView.w) {
    let gsl_partVanish_52 = gsl_vanish(gsl_partLo_52, gsl_partHi_52);
    gsl_part_52 = 1e10;
    if (gsl_partVanish_52.x > 0.0) { gsl_part_52 = gsl_remap_4(p - vec3<f32>(180.0, -25.0, gz), gsl_param_33_8, gsl_param_36_10) + gsl_partVanish_52.y; }
    if (abs(gsl_part_52) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_52), gsl_partVanish_52.zw); }
  }
  let gantryR3 = gsl_part_52;
  let gsl_partLo_53 = vec3<f32>(155.24, -27.76, -4.76 + gz);
  let gsl_partHi_53 = vec3<f32>(164.76, -14.99, 4.76 + gz);
  let gsl_partBox_53 = gsl_boxDistance(p, gsl_partLo_53, gsl_partHi_53);
  var gsl_part_53 = gsl_partBox_53;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_53 <= gsl_vanishView.w) {
    let gsl_partVanish_53 = gsl_vanish(gsl_partLo_53, gsl_partHi_53);
    gsl_part_53 = 1e10;
    if (gsl_partVanish_53.x > 0.0) { gsl_part_53 = gsl_remap_4(p - vec3<f32>(160.0, -25.0, gz), gsl_param_33_8, gsl_param_36_10) + gsl_partVanish_53.y; }
    if (abs(gsl_part_53) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_53), gsl_partVanish_53.zw); }
  }
  let gantryR4 = gsl_part_53;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_45 <= gsl_vanishView.w) {
    let gsl_partVanish_45 = gsl_vanish(gsl_partLo_45, gsl_partHi_45);
    gsl_part_45 = 1e10;
    if (gsl_partVanish_45.x > 0.0) { gsl_part_45 = min(min(min(min(min(min(min(gantryL1, gantryL2), gantryL3), gantryL4), gantryR1), gantryR2), gantryR3), gantryR4) + gsl_partVanish_45.y; }
    if (abs(gsl_part_45) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_45), gsl_partVanish_45.zw); }
  }
  let gantryBolts = gsl_part_45;
  let gsl_partLo_54 = vec3<f32>(min(min(min(-4.76 + travel - 22.0, -4.76 + travel + 22.0), -4.76 + travel - 22.0), -4.76 + travel + 22.0), -28.76, min(min(min(-4.76 + gz - 22.0, -4.76 + gz - 22.0), -4.76 + gz + 22.0), -4.76 + gz + 22.0));
  let gsl_partHi_54 = vec3<f32>(max(max(max(4.76 + travel - 22.0, 4.76 + travel + 22.0), 4.76 + travel - 22.0), 4.76 + travel + 22.0), -15.99, max(max(max(4.76 + gz - 22.0, 4.76 + gz - 22.0), 4.76 + gz + 22.0), 4.76 + gz + 22.0));
  let gsl_partBox_54 = gsl_boxDistance(p, gsl_partLo_54, gsl_partHi_54);
  var gsl_part_54 = gsl_partBox_54;
  let gsl_partLo_55 = vec3<f32>(-4.76 + travel - 22.0, -28.76, -4.76 + gz - 22.0);
  let gsl_partHi_55 = vec3<f32>(4.76 + travel - 22.0, -15.99, 4.76 + gz - 22.0);
  let gsl_partBox_55 = gsl_boxDistance(p, gsl_partLo_55, gsl_partHi_55);
  var gsl_part_55 = gsl_partBox_55;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_55 <= gsl_vanishView.w) {
    let gsl_partVanish_55 = gsl_vanish(gsl_partLo_55, gsl_partHi_55);
    gsl_part_55 = 1e10;
    if (gsl_partVanish_55.x > 0.0) { gsl_part_55 = gsl_remap_4(p - vec3<f32>(travel - 22.0, -26.0, gz - 22.0), gsl_param_33_8, gsl_param_36_10) + gsl_partVanish_55.y; }
    if (abs(gsl_part_55) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_55), gsl_partVanish_55.zw); }
  }
  let carriage1 = gsl_part_55;
  let gsl_partLo_56 = vec3<f32>(-4.76 + travel + 22.0, -28.76, -4.76 + gz - 22.0);
  let gsl_partHi_56 = vec3<f32>(4.76 + travel + 22.0, -15.99, 4.76 + gz - 22.0);
  let gsl_partBox_56 = gsl_boxDistance(p, gsl_partLo_56, gsl_partHi_56);
  var gsl_part_56 = gsl_partBox_56;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_56 <= gsl_vanishView.w) {
    let gsl_partVanish_56 = gsl_vanish(gsl_partLo_56, gsl_partHi_56);
    gsl_part_56 = 1e10;
    if (gsl_partVanish_56.x > 0.0) { gsl_part_56 = gsl_remap_4(p - vec3<f32>(travel + 22.0, -26.0, gz - 22.0), gsl_param_33_8, gsl_param_36_10) + gsl_partVanish_56.y; }
    if (abs(gsl_part_56) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_56), gsl_partVanish_56.zw); }
  }
  let carriage2 = gsl_part_56;
  let gsl_partLo_57 = vec3<f32>(-4.76 + travel - 22.0, -28.76, -4.76 + gz + 22.0);
  let gsl_partHi_57 = vec3<f32>(4.76 + travel - 22.0, -15.99, 4.76 + gz + 22.0);
  let gsl_partBox_57 = gsl_boxDistance(p, gsl_partLo_57, gsl_partHi_57);
  var gsl_part_57 = gsl_partBox_57;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_57 <= gsl_vanishView.w) {
    let gsl_partVanish_57 = gsl_vanish(gsl_partLo_57, gsl_partHi_57);
    gsl_part_57 = 1e10;
    if (gsl_partVanish_57.x > 0.0) { gsl_part_57 = gsl_remap_4(p - vec3<f32>(travel - 22.0, -26.0, gz + 22.0), gsl_param_33_8, gsl_param_36_10) + gsl_partVanish_57.y; }
    if (abs(gsl_part_57) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_57), gsl_partVanish_57.zw); }
  }
  let carriage3 = gsl_part_57;
  let gsl_partLo_58 = vec3<f32>(-4.76 + travel + 22.0, -28.76, -4.76 + gz + 22.0);
  let gsl_partHi_58 = vec3<f32>(4.76 + travel + 22.0, -15.99, 4.76 + gz + 22.0);
  let gsl_partBox_58 = gsl_boxDistance(p, gsl_partLo_58, gsl_partHi_58);
  var gsl_part_58 = gsl_partBox_58;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_58 <= gsl_vanishView.w) {
    let gsl_partVanish_58 = gsl_vanish(gsl_partLo_58, gsl_partHi_58);
    gsl_part_58 = 1e10;
    if (gsl_partVanish_58.x > 0.0) { gsl_part_58 = gsl_remap_4(p - vec3<f32>(travel + 22.0, -26.0, gz + 22.0), gsl_param_33_8, gsl_param_36_10) + gsl_partVanish_58.y; }
    if (abs(gsl_part_58) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_58), gsl_partVanish_58.zw); }
  }
  let carriage4 = gsl_part_58;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_54 <= gsl_vanishView.w) {
    let gsl_partVanish_54 = gsl_vanish(gsl_partLo_54, gsl_partHi_54);
    gsl_part_54 = 1e10;
    if (gsl_partVanish_54.x > 0.0) { gsl_part_54 = min(min(min(carriage1, carriage2), carriage3), carriage4) + gsl_partVanish_54.y; }
    if (abs(gsl_part_54) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_54), gsl_partVanish_54.zw); }
  }
  let carriageBolts = gsl_part_54;
  let gsl_partLo_59 = vec3<f32>(-215.01, -204.26, -194.26);
  let gsl_partHi_59 = vec3<f32>(215.01, 204.26, -185.74);
  let gsl_partBox_59 = gsl_boxDistance(p, gsl_partLo_59, gsl_partHi_59);
  var gsl_part_59 = gsl_partBox_59;
  let gsl_partLo_60 = vec3<f32>(-215.01, -204.26, -194.26);
  let gsl_partHi_60 = vec3<f32>(-179.99, -195.74, -185.74);
  let gsl_partBox_60 = gsl_boxDistance(p, gsl_partLo_60, gsl_partHi_60);
  var gsl_part_60 = gsl_partBox_60;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_60 <= gsl_vanishView.w) {
    let gsl_partVanish_60 = gsl_vanish(gsl_partLo_60, gsl_partHi_60);
    gsl_part_60 = 1e10;
    if (gsl_partVanish_60.x > 0.0) { gsl_part_60 = gsl_remap_6(p - vec3<f32>(-210.0, -200.0, -190.0), gsl_param_33_8, gsl_param_36_10) + gsl_partVanish_60.y; }
    if (abs(gsl_part_60) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_60), gsl_partVanish_60.zw); }
  }
  let blindL1 = gsl_part_60;
  let gsl_partLo_61 = vec3<f32>(-215.01, 195.74, -194.26);
  let gsl_partHi_61 = vec3<f32>(-179.99, 204.26, -185.74);
  let gsl_partBox_61 = gsl_boxDistance(p, gsl_partLo_61, gsl_partHi_61);
  var gsl_part_61 = gsl_partBox_61;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_61 <= gsl_vanishView.w) {
    let gsl_partVanish_61 = gsl_vanish(gsl_partLo_61, gsl_partHi_61);
    gsl_part_61 = 1e10;
    if (gsl_partVanish_61.x > 0.0) { gsl_part_61 = gsl_remap_6(p - vec3<f32>(-210.0, 200.0, -190.0), gsl_param_33_8, gsl_param_36_10) + gsl_partVanish_61.y; }
    if (abs(gsl_part_61) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_61), gsl_partVanish_61.zw); }
  }
  let blindL2 = gsl_part_61;
  let gsl_partLo_62 = vec3<f32>(179.99, -204.26, -194.26);
  let gsl_partHi_62 = vec3<f32>(215.01, -195.74, -185.74);
  let gsl_partBox_62 = gsl_boxDistance(p, gsl_partLo_62, gsl_partHi_62);
  var gsl_part_62 = gsl_partBox_62;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_62 <= gsl_vanishView.w) {
    let gsl_partVanish_62 = gsl_vanish(gsl_partLo_62, gsl_partHi_62);
    gsl_part_62 = 1e10;
    if (gsl_partVanish_62.x > 0.0) { gsl_part_62 = gsl_remap_7(p - vec3<f32>(210.0, -200.0, -190.0), gsl_param_33_8, gsl_param_36_10) + gsl_partVanish_62.y; }
    if (abs(gsl_part_62) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_62), gsl_partVanish_62.zw); }
  }
  let blindR1 = gsl_part_62;
  let gsl_partLo_63 = vec3<f32>(179.99, 195.74, -194.26);
  let gsl_partHi_63 = vec3<f32>(215.01, 204.26, -185.74);
  let gsl_partBox_63 = gsl_boxDistance(p, gsl_partLo_63, gsl_partHi_63);
  var gsl_part_63 = gsl_partBox_63;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_63 <= gsl_vanishView.w) {
    let gsl_partVanish_63 = gsl_vanish(gsl_partLo_63, gsl_partHi_63);
    gsl_part_63 = 1e10;
    if (gsl_partVanish_63.x > 0.0) { gsl_part_63 = gsl_remap_7(p - vec3<f32>(210.0, 200.0, -190.0), gsl_param_33_8, gsl_param_36_10) + gsl_partVanish_63.y; }
    if (abs(gsl_part_63) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_63), gsl_partVanish_63.zw); }
  }
  let blindR2 = gsl_part_63;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_59 <= gsl_vanishView.w) {
    let gsl_partVanish_59 = gsl_vanish(gsl_partLo_59, gsl_partHi_59);
    gsl_part_59 = 1e10;
    if (gsl_partVanish_59.x > 0.0) { gsl_part_59 = min(min(min(blindL1, blindL2), blindR1), blindR2) + gsl_partVanish_59.y; }
    if (abs(gsl_part_59) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_59), gsl_partVanish_59.zw); }
  }
  let blindBolts = gsl_part_59;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_18 <= gsl_vanishView.w) {
    let gsl_partVanish_18 = gsl_vanish(gsl_partLo_18, gsl_partHi_18);
    gsl_part_18 = 1e10;
    if (gsl_partVanish_18.x > 0.0) { gsl_part_18 = min(min(min(min(min(min(lowBoltsL, lowBoltsR), topBoltsL), topBoltsR), gantryBolts), carriageBolts), blindBolts) + gsl_partVanish_18.y; }
    if (abs(gsl_part_18) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_18), gsl_partVanish_18.zw); }
  }
  let bolts = gsl_part_18;
  let gsl_partLo_64 = vec3<f32>(min(min(min(travel - 14.0, travel + 14.0), min(travel, travel) - 11.0), min(travel - 8.0, travel + 8.0)), -51.0, min(min(min(gz - 12.0, gz + 12.0), min(gz - 38.0, gz - 8.0) - 11.0), min(gz - 50.0, gz - 38.0)));
  let gsl_partHi_64 = vec3<f32>(max(max(max(travel - 14.0, travel + 14.0), max(travel, travel) + 11.0), max(travel - 8.0, travel + 8.0)), -26.0, max(max(max(gz - 12.0, gz + 12.0), max(gz - 38.0, gz - 8.0) + 11.0), max(gz - 50.0, gz - 38.0)));
  let gsl_partBox_64 = gsl_boxDistance(p, gsl_partLo_64, gsl_partHi_64);
  var gsl_part_64 = gsl_partBox_64;
  let gsl_partLo_65 = vec3<f32>(min(travel - 14.0, travel + 14.0), -30.0, min(gz - 12.0, gz + 12.0));
  let gsl_partHi_65 = vec3<f32>(max(travel - 14.0, travel + 14.0), -26.0, max(gz - 12.0, gz + 12.0));
  let gsl_partBox_65 = gsl_boxDistance(p, gsl_partLo_65, gsl_partHi_65);
  var gsl_part_65 = gsl_partBox_65;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_65 <= gsl_vanishView.w) {
    let gsl_partVanish_65 = gsl_vanish(gsl_partLo_65, gsl_partHi_65);
    gsl_part_65 = 1e10;
    if (gsl_partVanish_65.x > 0.0) { gsl_part_65 = builtin_box(p, vec3<f32>(travel - 14.0, -30.0, gz - 12.0), vec3<f32>(travel + 14.0, -26.0, gz + 12.0)) + gsl_partVanish_65.y; }
    if (abs(gsl_part_65) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_65), gsl_partVanish_65.zw); }
  }
  let mount = gsl_part_65;
  let gsl_partLo_66 = vec3<f32>(min(travel, travel) - 11.0, -51.0, min(gz - 38.0, gz - 8.0) - 11.0);
  let gsl_partHi_66 = vec3<f32>(max(travel, travel) + 11.0, -29.0, max(gz - 38.0, gz - 8.0) + 11.0);
  let gsl_partBox_66 = gsl_boxDistance(p, gsl_partLo_66, gsl_partHi_66);
  var gsl_part_66 = gsl_partBox_66;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_66 <= gsl_vanishView.w) {
    let gsl_partVanish_66 = gsl_vanish(gsl_partLo_66, gsl_partHi_66);
    gsl_part_66 = 1e10;
    if (gsl_partVanish_66.x > 0.0) { gsl_part_66 = builtin_cylinderCapped(p, vec3<f32>(travel, -40.0, gz - 38.0), vec3<f32>(travel, -40.0, gz - 8.0), 11.0) + gsl_partVanish_66.y; }
    if (abs(gsl_part_66) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_66), gsl_partVanish_66.zw); }
  }
  let heatsink = gsl_part_66;
  let gsl_partLo_67 = vec3<f32>(min(travel - 8.0, travel + 8.0), -48.0, min(gz - 50.0, gz - 38.0));
  let gsl_partHi_67 = vec3<f32>(max(travel - 8.0, travel + 8.0), -32.0, max(gz - 50.0, gz - 38.0));
  let gsl_partBox_67 = gsl_boxDistance(p, gsl_partLo_67, gsl_partHi_67);
  var gsl_part_67 = gsl_partBox_67;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_67 <= gsl_vanishView.w) {
    let gsl_partVanish_67 = gsl_vanish(gsl_partLo_67, gsl_partHi_67);
    gsl_part_67 = 1e10;
    if (gsl_partVanish_67.x > 0.0) { gsl_part_67 = builtin_box(p, vec3<f32>(travel - 8.0, -48.0, gz - 50.0), vec3<f32>(travel + 8.0, -32.0, gz - 38.0)) + gsl_partVanish_67.y; }
    if (abs(gsl_part_67) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_67), gsl_partVanish_67.zw); }
  }
  let heaterBlock = gsl_part_67;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_64 <= gsl_vanishView.w) {
    let gsl_partVanish_64 = gsl_vanish(gsl_partLo_64, gsl_partHi_64);
    gsl_part_64 = 1e10;
    if (gsl_partVanish_64.x > 0.0) { gsl_part_64 = min(min(mount, heatsink), heaterBlock) + gsl_partVanish_64.y; }
    if (abs(gsl_part_64) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_64), gsl_partVanish_64.zw); }
  }
  let hotend = gsl_part_64;
  let gsl_partLo_68 = vec3<f32>(min(travel, travel) - 3.5, -43.5, min(gz - 50.0, gz - 56.0) - 3.5);
  let gsl_partHi_68 = vec3<f32>(max(travel, travel) + 3.5, -36.5, max(gz - 50.0, gz - 56.0) + 3.5);
  let gsl_partBox_68 = gsl_boxDistance(p, gsl_partLo_68, gsl_partHi_68);
  var gsl_part_68 = gsl_partBox_68;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_68 <= gsl_vanishView.w) {
    let gsl_partVanish_68 = gsl_vanish(gsl_partLo_68, gsl_partHi_68);
    gsl_part_68 = 1e10;
    if (gsl_partVanish_68.x > 0.0) { gsl_part_68 = builtin_coneCapped(p, vec3<f32>(travel, -40.0, gz - 50.0), vec3<f32>(travel, -40.0, gz - 56.0), 3.5, 0.6) + gsl_partVanish_68.y; }
    if (abs(gsl_part_68) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_68), gsl_partVanish_68.zw); }
  }
  let nozzle = gsl_part_68;
  let gsl_s1 = gsl_shape(frame, 1u);
  let gsl_s2 = gsl_shape(plates, 2u);
  let gsl_s3 = gsl_shape(bolts, 3u);
  let gsl_s4 = gsl_shape(hotend, 4u);
  let gsl_s5 = gsl_shape(nozzle, 5u);
  return gsl_min(gsl_min(gsl_min(gsl_min(gsl_s1, gsl_s2), gsl_s3), gsl_s4), gsl_s5);
}

fn gsl_colorField(p: vec3<f32>, time: f32, gsl_colorIdx: u32, gsl_param_33_8: f32, gsl_param_36_10: f32) -> vec4<f32> {
  if (gsl_colorIdx == 1u) { return vec4<f32>(vec3<f32>(0.9621, 0.9654, 0.9658), 1.0); }
  if (gsl_colorIdx == 2u) { return vec4<f32>(vec3<f32>(0.08, 0.08, 0.09), 1.0); }
  if (gsl_colorIdx == 3u) { return vec4<f32>(vec3<f32>(0.22, 0.22, 0.24), 1.0); }
  if (gsl_colorIdx == 4u) { return vec4<f32>(vec3<f32>(0.9621, 0.9654, 0.9658), 1.0); }
  if (gsl_colorIdx == 5u) { return vec4<f32>(vec3<f32>(0.9593, 0.8952, 0.6822), 1.0); }
  return vec4<f32>(0.0, 0.0, 0.0, 0.0);
}
fn gsl_materialField(p: vec3<f32>, time: f32, gsl_materialIdx: u32, gsl_param_33_8: f32, gsl_param_36_10: f32) -> gsl_material {
  if (gsl_materialIdx == 1u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.4, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 2u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.5, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 3u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.55, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 4u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.4, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 5u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.3, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
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
