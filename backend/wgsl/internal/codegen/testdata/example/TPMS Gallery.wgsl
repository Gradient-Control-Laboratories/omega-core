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

fn builtin_linearMap(gsl_var_value: f32, inMin: f32, inMax: f32, outMin: f32, outMax: f32) -> f32 {
  let slope = (outMax - outMin) / (inMax - inMin);
  return (gsl_var_value - inMin) * slope + outMin;
}

fn builtin_ramp(gsl_var_value: f32, inMin: f32, inMax: f32, outMin: f32, outMax: f32) -> f32 {
  let mapped = builtin_linearMap(gsl_var_value, inMin, inMax, outMin, outMax);
  let low = min(outMin, outMax);
  let high = max(outMin, outMax);
  return clamp(mapped, low, high);
}

fn lattice_solidify(field: f32, bias: f32) -> f32 {
  return field - bias;
}

fn lattice_solidify_2(field: f32, bias: f32, thickness: f32) -> f32 {
  return abs(field - bias) - thickness * 0.5;
}

fn tpms_sum3(v: vec3<f32>) -> f32 {
  return v.x + v.y + v.z;
}

fn tpms_gyroid(p: vec3<f32>, period: vec3<f32>, drop: vec3<f32>) -> f32 {
  let freq = 6.2831855 / period;
  let xyz = p * freq;
  let yzx = vec3<f32>(xyz.y, xyz.z, xyz.x);
  let terms = drop * sin(xyz) * cos(yzx);
  return tpms_sum3(terms) * tpms_sum3(period) / 18.0;
}

fn tpms_diamond(p: vec3<f32>, period: vec3<f32>, drop: vec3<f32>, gyroidBlend: f32) -> f32 {
  let blendFactor = 1.0 - gyroidBlend;
  let freq = 6.2831855 / period;
  let xyz = p * freq;
  let sinXyz = sin(xyz);
  let term1 = blendFactor * sinXyz.x * sinXyz.y * sinXyz.z;
  let yzx = vec3<f32>(xyz.y, xyz.z, xyz.x);
  let cosYzx = cos(yzx);
  let cosZxy = vec3<f32>(cosYzx.y, cosYzx.z, cosYzx.x);
  let gsl_var_mix = cosZxy * blendFactor + gyroidBlend;
  let term2 = tpms_sum3(drop * sinXyz * cosYzx * gsl_var_mix);
  return (term1 + term2) * tpms_sum3(period) / 33.941124;
}

fn specimen(p: vec3<f32>, field: f32, ox: f32, oy: f32) -> f32 {
  return max(field, builtin_boxCentered(p, vec3<f32>(ox, oy, 0.0), vec3<f32>(28.0, 28.0, 28.0)));
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  gsl_partNear = vec3<f32>(1e10, 1e10, 1e10);
  let gsl_partLo = vec3<f32>(-116.0, 22.0, -14.0);
  let gsl_partHi = vec3<f32>(-88.0, 50.0, 14.0);
  let gsl_partBox = gsl_boxDistance(p, gsl_partLo, gsl_partHi);
  var gsl_part = gsl_partBox;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox <= gsl_vanishView.w) {
    let gsl_partVanish = gsl_vanish(gsl_partLo, gsl_partHi);
    gsl_part = 1e10;
    if (gsl_partVanish.x > 0.0) { gsl_part = specimen(p, tpms_gyroid(p, vec3<f32>(22.0, 22.0, 22.0), vec3<f32>(1.0, 1.0, 1.0)), -102.0, 36.0) + gsl_partVanish.y; }
    if (abs(gsl_part) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part), gsl_partVanish.zw); }
  }
  let gyr = gsl_part;
  let gsl_partLo_2 = vec3<f32>(-82.0, 22.0, -14.0);
  let gsl_partHi_2 = vec3<f32>(-54.0, 50.0, 14.0);
  let gsl_partBox_2 = gsl_boxDistance(p, gsl_partLo_2, gsl_partHi_2);
  var gsl_part_2 = gsl_partBox_2;
  let gsl_tmp_1 = p * vec3<f32>(0.28559932, 0.28559932, 0.28559932);
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_2 <= gsl_vanishView.w) {
    let gsl_partVanish_2 = gsl_vanish(gsl_partLo_2, gsl_partHi_2);
    gsl_part_2 = 1e10;
    if (gsl_partVanish_2.x > 0.0) { gsl_part_2 = specimen(p, tpms_sum3(vec3<f32>(1.0, 1.0, 1.0) * cos(vec3<f32>(gsl_tmp_1.y, gsl_tmp_1.z, gsl_tmp_1.x)) * (sin(gsl_tmp_1) * 0.0 - 1.0)) * 66.0 / 36.0, -68.0, 36.0) + gsl_partVanish_2.y; }
    if (abs(gsl_part_2) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_2), gsl_partVanish_2.zw); }
  }
  let sch = gsl_part_2;
  let gsl_partLo_3 = vec3<f32>(-48.0, 22.0, -14.0);
  let gsl_partHi_3 = vec3<f32>(-20.0, 50.0, 14.0);
  let gsl_partBox_3 = gsl_boxDistance(p, gsl_partLo_3, gsl_partHi_3);
  var gsl_part_3 = gsl_partBox_3;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_3 <= gsl_vanishView.w) {
    let gsl_partVanish_3 = gsl_vanish(gsl_partLo_3, gsl_partHi_3);
    gsl_part_3 = 1e10;
    if (gsl_partVanish_3.x > 0.0) { gsl_part_3 = specimen(p, tpms_diamond(p, vec3<f32>(22.0, 22.0, 22.0), vec3<f32>(1.0, 1.0, 1.0), 0.0), -34.0, 36.0) + gsl_partVanish_3.y; }
    if (abs(gsl_part_3) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_3), gsl_partVanish_3.zw); }
  }
  let dia = gsl_part_3;
  let gsl_partLo_4 = vec3<f32>(-14.0, 22.0, -14.0);
  let gsl_partHi_4 = vec3<f32>(14.0, 50.0, 14.0);
  let gsl_partBox_4 = gsl_boxDistance(p, gsl_partLo_4, gsl_partHi_4);
  var gsl_part_4 = gsl_partBox_4;
  let gsl_tmp_2 = p * vec3<f32>(0.28559932, 0.28559932, 0.28559932);
  let gsl_tmp_3 = cos(gsl_tmp_2 * 2.0);
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_4 <= gsl_vanishView.w) {
    let gsl_partVanish_4 = gsl_vanish(gsl_partLo_4, gsl_partHi_4);
    gsl_part_4 = 1e10;
    if (gsl_partVanish_4.x > 0.0) { gsl_part_4 = specimen(p, (tpms_sum3(vec3<f32>(1.0, 1.0, 1.0) * sin(vec3<f32>(gsl_tmp_2.z, gsl_tmp_2.x, gsl_tmp_2.y)) * cos(vec3<f32>(gsl_tmp_2.y, gsl_tmp_2.z, gsl_tmp_2.x)) * (sin(gsl_tmp_2 * 2.0) * 1.0 + 0.0)) - 1.0 * tpms_sum3(gsl_tmp_3 * vec3<f32>(gsl_tmp_3.y, gsl_tmp_3.z, gsl_tmp_3.x))) * 66.0 / 72.0, 0.0, 36.0) + gsl_partVanish_4.y; }
    if (abs(gsl_part_4) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_4), gsl_partVanish_4.zw); }
  }
  let lid = gsl_part_4;
  let gsl_partLo_5 = vec3<f32>(20.0, 22.0, -14.0);
  let gsl_partHi_5 = vec3<f32>(48.0, 50.0, 14.0);
  let gsl_partBox_5 = gsl_boxDistance(p, gsl_partLo_5, gsl_partHi_5);
  var gsl_part_5 = gsl_partBox_5;
  let gsl_tmp_4 = cos(p * vec3<f32>(0.28559932, 0.28559932, 0.28559932)) * vec3<f32>(1.0, 1.0, 1.0);
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_5 <= gsl_vanishView.w) {
    let gsl_partVanish_5 = gsl_vanish(gsl_partLo_5, gsl_partHi_5);
    gsl_part_5 = 1e10;
    if (gsl_partVanish_5.x > 0.0) { gsl_part_5 = specimen(p, (-tpms_sum3(gsl_tmp_4) - 1.3333334 * gsl_tmp_4.x * gsl_tmp_4.y * gsl_tmp_4.z) * 66.0 / 52.0, 34.0, 36.0) + gsl_partVanish_5.y; }
    if (abs(gsl_part_5) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_5), gsl_partVanish_5.zw); }
  }
  let neo = gsl_part_5;
  let gsl_partLo_6 = vec3<f32>(54.0, 22.0, -14.0);
  let gsl_partHi_6 = vec3<f32>(82.0, 50.0, 14.0);
  let gsl_partBox_6 = gsl_boxDistance(p, gsl_partLo_6, gsl_partHi_6);
  var gsl_part_6 = gsl_partBox_6;
  let gsl_tmp_5 = p * vec3<f32>(0.28559932, 0.28559932, 0.28559932);
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_6 <= gsl_vanishView.w) {
    let gsl_partVanish_6 = gsl_vanish(gsl_partLo_6, gsl_partHi_6);
    gsl_part_6 = 1e10;
    if (gsl_partVanish_6.x > 0.0) { gsl_part_6 = specimen(p, (2.0 * tpms_sum3(vec3<f32>(1.0, 1.0, 1.0) * cos(gsl_tmp_5) * cos(vec3<f32>(gsl_tmp_5.y, gsl_tmp_5.z, gsl_tmp_5.x))) - tpms_sum3(cos(gsl_tmp_5 * 2.0))) * 66.0 / 48.0, 68.0, 36.0) + gsl_partVanish_6.y; }
    if (abs(gsl_part_6) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_6), gsl_partVanish_6.zw); }
  }
  let wp = gsl_part_6;
  let gsl_partLo_7 = vec3<f32>(88.0, 22.0, -14.0);
  let gsl_partHi_7 = vec3<f32>(116.0, 50.0, 14.0);
  let gsl_partBox_7 = gsl_boxDistance(p, gsl_partLo_7, gsl_partHi_7);
  var gsl_part_7 = gsl_partBox_7;
  let gsl_tmp_6 = p * vec3<f32>(0.28559932, 0.28559932, 0.28559932);
  let gsl_tmp_7 = sin(gsl_tmp_6 * 2.0);
  let gsl_tmp_8 = cos(gsl_tmp_6 * 2.0);
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_7 <= gsl_vanishView.w) {
    let gsl_partVanish_7 = gsl_vanish(gsl_partLo_7, gsl_partHi_7);
    gsl_part_7 = 1e10;
    if (gsl_partVanish_7.x > 0.0) { gsl_part_7 = specimen(p, (-1.0 * tpms_sum3(gsl_tmp_7 * cos(vec3<f32>(gsl_tmp_6.y, gsl_tmp_6.z, gsl_tmp_6.x)) * sin(vec3<f32>(gsl_tmp_6.z, gsl_tmp_6.x, gsl_tmp_6.y))) + 0.5 * tpms_sum3(gsl_tmp_7 * gsl_tmp_8) + 0.25 * tpms_sum3(gsl_tmp_8)) * 66.0 / 36.0, 102.0, 36.0) + gsl_partVanish_7.y; }
    if (abs(gsl_part_7) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_7), gsl_partVanish_7.zw); }
  }
  let spl = gsl_part_7;
  let gsl_partLo_8 = vec3<f32>(-65.0, -14.0, -14.0);
  let gsl_partHi_8 = vec3<f32>(-37.0, 14.0, 14.0);
  let gsl_partBox_8 = gsl_boxDistance(p, gsl_partLo_8, gsl_partHi_8);
  var gsl_part_8 = gsl_partBox_8;
  let gf = tpms_gyroid(p, vec3<f32>(22.0, 22.0, 22.0), vec3<f32>(1.0, 1.0, 1.0));
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_8 <= gsl_vanishView.w) {
    let gsl_partVanish_8 = gsl_vanish(gsl_partLo_8, gsl_partHi_8);
    gsl_part_8 = 1e10;
    if (gsl_partVanish_8.x > 0.0) { gsl_part_8 = specimen(p, lattice_solidify(gf, 0.0), -51.0, 0.0) + gsl_partVanish_8.y; }
    if (abs(gsl_part_8) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_8), gsl_partVanish_8.zw); }
  }
  let vThick = gsl_part_8;
  let gsl_partLo_9 = vec3<f32>(-31.0, -14.0, -14.0);
  let gsl_partHi_9 = vec3<f32>(-3.0, 14.0, 14.0);
  let gsl_partBox_9 = gsl_boxDistance(p, gsl_partLo_9, gsl_partHi_9);
  var gsl_part_9 = gsl_partBox_9;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_9 <= gsl_vanishView.w) {
    let gsl_partVanish_9 = gsl_vanish(gsl_partLo_9, gsl_partHi_9);
    gsl_part_9 = 1e10;
    if (gsl_partVanish_9.x > 0.0) { gsl_part_9 = specimen(p, lattice_solidify_2(gf, 0.0, 1.4), -17.0, 0.0) + gsl_partVanish_9.y; }
    if (abs(gsl_part_9) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_9), gsl_partVanish_9.zw); }
  }
  let vThin = gsl_part_9;
  let gsl_partLo_10 = vec3<f32>(3.0, -14.0, -14.0);
  let gsl_partHi_10 = vec3<f32>(31.0, 14.0, 14.0);
  let gsl_partBox_10 = gsl_boxDistance(p, gsl_partLo_10, gsl_partHi_10);
  var gsl_part_10 = gsl_partBox_10;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_10 <= gsl_vanishView.w) {
    let gsl_partVanish_10 = gsl_vanish(gsl_partLo_10, gsl_partHi_10);
    gsl_part_10 = 1e10;
    if (gsl_partVanish_10.x > 0.0) { gsl_part_10 = specimen(p, lattice_solidify(-gf, 0.0), 17.0, 0.0) + gsl_partVanish_10.y; }
    if (abs(gsl_part_10) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_10), gsl_partVanish_10.zw); }
  }
  let vInverse = gsl_part_10;
  let gsl_partLo_11 = vec3<f32>(37.0, -14.0, -14.0);
  let gsl_partHi_11 = vec3<f32>(65.0, 14.0, 14.0);
  let gsl_partBox_11 = gsl_boxDistance(p, gsl_partLo_11, gsl_partHi_11);
  var gsl_part_11 = gsl_partBox_11;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_11 <= gsl_vanishView.w) {
    let gsl_partVanish_11 = gsl_vanish(gsl_partLo_11, gsl_partHi_11);
    gsl_part_11 = 1e10;
    if (gsl_partVanish_11.x > 0.0) { gsl_part_11 = specimen(p, -lattice_solidify_2(gf, 0.0, 1.4), 51.0, 0.0) + gsl_partVanish_11.y; }
    if (abs(gsl_part_11) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_11), gsl_partVanish_11.zw); }
  }
  let vThinInv = gsl_part_11;
  let gsl_partLo_12 = vec3<f32>(-119.0, -50.0, -14.0);
  let gsl_partHi_12 = vec3<f32>(119.0, -22.0, 14.0);
  let gsl_partBox_12 = gsl_boxDistance(p, gsl_partLo_12, gsl_partHi_12);
  var gsl_part_12 = gsl_partBox_12;
  let gyrToDia = builtin_ramp(p.x, -119.0, 119.0, 1.0, 0.0);
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_12 <= gsl_vanishView.w) {
    let gsl_partVanish_12 = gsl_vanish(gsl_partLo_12, gsl_partHi_12);
    gsl_part_12 = 1e10;
    if (gsl_partVanish_12.x > 0.0) { gsl_part_12 = max(tpms_diamond(p, vec3<f32>(22.0, 22.0, 22.0), vec3<f32>(1.0, 1.0, 1.0), gyrToDia), builtin_gcl_closure_boxCentered(p, vec3<f32>(0.0, -36.0, 0.0), vec3<f32>(238.0, 28.0, 28.0))) + gsl_partVanish_12.y; }
    if (abs(gsl_part_12) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_12), gsl_partVanish_12.zw); }
  }
  let morph = gsl_part_12;
  let gsl_s1 = gsl_shape(gyr, 1u);
  let gsl_s2 = gsl_shape(sch, 2u);
  let gsl_s3 = gsl_shape(dia, 3u);
  let gsl_s4 = gsl_shape(lid, 4u);
  let gsl_s5 = gsl_shape(neo, 5u);
  let gsl_s6 = gsl_shape(wp, 6u);
  let gsl_s7 = gsl_shape(spl, 7u);
  let gsl_s8 = gsl_shape(vThick, 8u);
  let gsl_s9 = gsl_shape(vThin, 9u);
  let gsl_s10 = gsl_shape(vInverse, 10u);
  let gsl_s11 = gsl_shape(vThinInv, 11u);
  let gsl_s12 = gsl_shape(morph, 12u);
  return gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_s1, gsl_s2), gsl_s3), gsl_s4), gsl_s5), gsl_s6), gsl_s7), gsl_s8), gsl_s9), gsl_s10), gsl_s11), gsl_s12);
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
