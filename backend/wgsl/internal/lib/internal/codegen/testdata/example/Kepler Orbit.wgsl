fn keplerStep(prevE: f32, meanAnom: f32, eccen: f32) -> f32 {
  let fVal = prevE - eccen * sin(prevE) - meanAnom;
  let fPrime = 1.0 - eccen * cos(prevE);
  return prevE - fVal / fPrime;
}

fn orbitPos(semiMajor: f32, semiMinor: f32, focusOffset: f32, E: f32) -> vec3<f32> {
  let xc = semiMajor * cos(E);
  let yc = semiMinor * sin(E);
  return vec3<f32>(xc - focusOffset, yc, 0.0);
}

fn sphere(p: vec3<f32>, center: vec3<f32>, radius: f32) -> f32 {
  return length(p - center) - radius;
}

fn cylinderRounded(p: vec3<f32>, a: vec3<f32>, b: vec3<f32>, radius: f32) -> f32 {
  let pa = p - a;
  let ba = b - a;
  let h = clamp(dot(pa, ba) / dot(ba, ba), 0.0, 1.0);
  return length(pa - ba * h) - radius;
}

fn sdShape(p: vec3<f32>, time: f32, gsl_param_23_28: f32, gsl_param_24_28: f32) -> gsl_shape {
  gsl_partNear = vec3<f32>(1e10, 1e10, 1e10);
  let semiMajor = gsl_param_23_28;
  let ecc = gsl_param_24_28;
  let semiMinor = semiMajor * sqrt(1.0 - ecc * ecc);
  let focusOffset = semiMajor * ecc;
  let meanAnomaly = time;
  let eccAnom0 = meanAnomaly;
  let eccAnom1 = keplerStep(eccAnom0, meanAnomaly, ecc);
  let eccAnom2 = keplerStep(eccAnom1, meanAnomaly, ecc);
  let eccAnom3 = keplerStep(eccAnom2, meanAnomaly, ecc);
  let eccAnom4 = keplerStep(eccAnom3, meanAnomaly, ecc);
  let eccAnom = eccAnom4;
  let center = orbitPos(semiMajor, semiMinor, focusOffset, eccAnom);
  let gsl_partLo = vec3<f32>(center.x - 0.25, center.y - 0.25, center.z - 0.25);
  let gsl_partHi = vec3<f32>(center.x + 0.25, center.y + 0.25, center.z + 0.25);
  let gsl_partBox = gsl_boxDistance(p, gsl_partLo, gsl_partHi);
  var gsl_part = gsl_partBox;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox <= gsl_vanishView.w) {
    let gsl_partVanish = gsl_vanish(gsl_partLo, gsl_partHi);
    gsl_part = 1e10;
    if (gsl_partVanish.x > 0.0) { gsl_part = sphere(p, center, 0.25) + gsl_partVanish.y; }
    if (abs(gsl_part) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part), gsl_partVanish.zw); }
  }
  let body = gsl_part;
  let gsl_tmp_1 = 1.0 - ((1.0 - ecc * cos(eccAnom0) * 1.0 - 1.0) * (1.0 - ecc * cos(eccAnom0)) - (eccAnom0 - ecc * sin(eccAnom0) - meanAnomaly) * -(ecc * -sin(eccAnom0) * 1.0)) / ((1.0 - ecc * cos(eccAnom0)) * (1.0 - ecc * cos(eccAnom0)));
  let gsl_tmp_2 = gsl_tmp_1 - ((gsl_tmp_1 - ecc * cos(eccAnom1) * gsl_tmp_1 - 1.0) * (1.0 - ecc * cos(eccAnom1)) - (eccAnom1 - ecc * sin(eccAnom1) - meanAnomaly) * -(ecc * -sin(eccAnom1) * gsl_tmp_1)) / ((1.0 - ecc * cos(eccAnom1)) * (1.0 - ecc * cos(eccAnom1)));
  let gsl_tmp_3 = gsl_tmp_2 - ((gsl_tmp_2 - ecc * cos(eccAnom2) * gsl_tmp_2 - 1.0) * (1.0 - ecc * cos(eccAnom2)) - (eccAnom2 - ecc * sin(eccAnom2) - meanAnomaly) * -(ecc * -sin(eccAnom2) * gsl_tmp_2)) / ((1.0 - ecc * cos(eccAnom2)) * (1.0 - ecc * cos(eccAnom2)));
  let gsl_tmp_4 = gsl_tmp_3 - ((gsl_tmp_3 - ecc * cos(eccAnom3) * gsl_tmp_3 - 1.0) * (1.0 - ecc * cos(eccAnom3)) - (eccAnom3 - ecc * sin(eccAnom3) - meanAnomaly) * -(ecc * -sin(eccAnom3) * gsl_tmp_3)) / ((1.0 - ecc * cos(eccAnom3)) * (1.0 - ecc * cos(eccAnom3)));
  let tanPoint = center + vec3<f32>(semiMajor * -sin(eccAnom) * gsl_tmp_4, semiMajor * 0.0 / (2.0 * (sqrt(1.0 - ecc * ecc) + 1.0 - select(0.0, 1.0, (sqrt(1.0 - ecc * ecc) > 0.0)))) * sin(eccAnom) + semiMinor * cos(eccAnom) * gsl_tmp_4, 0.0) * 0.5;
  let tangent = cylinderRounded(p, center, tanPoint, 0.1);
  let gsl_tmp_5 = 1.0 - ((1.0 - ecc * cos(eccAnom0) * 1.0 - 1.0) * (1.0 - ecc * cos(eccAnom0)) - (eccAnom0 - ecc * sin(eccAnom0) - meanAnomaly) * -(ecc * -sin(eccAnom0) * 1.0)) / ((1.0 - ecc * cos(eccAnom0)) * (1.0 - ecc * cos(eccAnom0)));
  let gsl_tmp_6 = gsl_tmp_5 - ((gsl_tmp_5 - ecc * cos(eccAnom1) * gsl_tmp_5 - 1.0) * (1.0 - ecc * cos(eccAnom1)) - (eccAnom1 - ecc * sin(eccAnom1) - meanAnomaly) * -(ecc * -sin(eccAnom1) * gsl_tmp_5)) / ((1.0 - ecc * cos(eccAnom1)) * (1.0 - ecc * cos(eccAnom1)));
  let gsl_tmp_7 = gsl_tmp_6 - ((gsl_tmp_6 - ecc * cos(eccAnom2) * gsl_tmp_6 - 1.0) * (1.0 - ecc * cos(eccAnom2)) - (eccAnom2 - ecc * sin(eccAnom2) - meanAnomaly) * -(ecc * -sin(eccAnom2) * gsl_tmp_6)) / ((1.0 - ecc * cos(eccAnom2)) * (1.0 - ecc * cos(eccAnom2)));
  let gsl_tmp_8 = gsl_tmp_7 - ((gsl_tmp_7 - ecc * cos(eccAnom3) * gsl_tmp_7 - 1.0) * (1.0 - ecc * cos(eccAnom3)) - (eccAnom3 - ecc * sin(eccAnom3) - meanAnomaly) * -(ecc * -sin(eccAnom3) * gsl_tmp_7)) / ((1.0 - ecc * cos(eccAnom3)) * (1.0 - ecc * cos(eccAnom3)));
  let gsl_tmp_9 = 1.0 - ecc * cos(eccAnom0) * 1.0 - 1.0;
  let gsl_tmp_10 = -(ecc * -sin(eccAnom0) * 1.0);
  let gsl_tmp_11 = gsl_tmp_9 * (1.0 - ecc * cos(eccAnom0)) - (eccAnom0 - ecc * sin(eccAnom0) - meanAnomaly) * gsl_tmp_10;
  let gsl_tmp_12 = (1.0 - ecc * cos(eccAnom0)) * (1.0 - ecc * cos(eccAnom0));
  let gsl_tmp_13 = 1.0 - gsl_tmp_11 / gsl_tmp_12 - ecc * cos(eccAnom1) * (1.0 - gsl_tmp_11 / gsl_tmp_12) - 1.0;
  let gsl_tmp_14 = -(ecc * -sin(eccAnom1) * (1.0 - gsl_tmp_11 / gsl_tmp_12));
  let gsl_tmp_15 = gsl_tmp_13 * (1.0 - ecc * cos(eccAnom1)) - (eccAnom1 - ecc * sin(eccAnom1) - meanAnomaly) * gsl_tmp_14;
  let gsl_tmp_16 = (1.0 - ecc * cos(eccAnom1)) * (1.0 - ecc * cos(eccAnom1));
  let gsl_tmp_17 = 1.0 - gsl_tmp_11 / gsl_tmp_12 - gsl_tmp_15 / gsl_tmp_16;
  let gsl_tmp_18 = gsl_tmp_17 - ecc * cos(eccAnom2) * gsl_tmp_17 - 1.0;
  let gsl_tmp_19 = -(ecc * -sin(eccAnom2) * gsl_tmp_17);
  let gsl_tmp_20 = gsl_tmp_18 * (1.0 - ecc * cos(eccAnom2)) - (eccAnom2 - ecc * sin(eccAnom2) - meanAnomaly) * gsl_tmp_19;
  let gsl_tmp_21 = (1.0 - ecc * cos(eccAnom2)) * (1.0 - ecc * cos(eccAnom2));
  let gsl_tmp_22 = gsl_tmp_17 - gsl_tmp_20 / gsl_tmp_21 - ecc * cos(eccAnom3) * (gsl_tmp_17 - gsl_tmp_20 / gsl_tmp_21) - 1.0;
  let gsl_tmp_23 = -(ecc * -sin(eccAnom3) * (gsl_tmp_17 - gsl_tmp_20 / gsl_tmp_21));
  let gsl_tmp_24 = gsl_tmp_22 * (1.0 - ecc * cos(eccAnom3)) - (eccAnom3 - ecc * sin(eccAnom3) - meanAnomaly) * gsl_tmp_23;
  let gsl_tmp_25 = (1.0 - ecc * cos(eccAnom3)) * (1.0 - ecc * cos(eccAnom3));
  let gsl_tmp_26 = gsl_tmp_17 - gsl_tmp_20 / gsl_tmp_21 - gsl_tmp_24 / gsl_tmp_25;
  let gsl_tmp_27 = -(ecc * -sin(eccAnom0) * 1.0);
  let gsl_tmp_28 = -(((-(ecc * -sin(eccAnom0) * 1.0 * 1.0) * (1.0 - ecc * cos(eccAnom0)) + gsl_tmp_9 * gsl_tmp_27 - ((1.0 - ecc * cos(eccAnom0) * 1.0 - 1.0) * gsl_tmp_10 + (eccAnom0 - ecc * sin(eccAnom0) - meanAnomaly) * -(ecc * -(cos(eccAnom0) * 1.0) * 1.0))) * gsl_tmp_12 - gsl_tmp_11 * (gsl_tmp_27 * (1.0 - ecc * cos(eccAnom0)) + (1.0 - ecc * cos(eccAnom0)) * gsl_tmp_27)) / (gsl_tmp_12 * gsl_tmp_12));
  let gsl_tmp_29 = -(ecc * -sin(eccAnom1) * gsl_tmp_5);
  let gsl_tmp_30 = gsl_tmp_28 - (((gsl_tmp_28 - ecc * (-sin(eccAnom1) * gsl_tmp_5 * (1.0 - gsl_tmp_11 / gsl_tmp_12) + cos(eccAnom1) * gsl_tmp_28)) * (1.0 - ecc * cos(eccAnom1)) + gsl_tmp_13 * gsl_tmp_29 - ((gsl_tmp_5 - ecc * cos(eccAnom1) * gsl_tmp_5 - 1.0) * gsl_tmp_14 + (eccAnom1 - ecc * sin(eccAnom1) - meanAnomaly) * -(ecc * (-(cos(eccAnom1) * gsl_tmp_5) * (1.0 - gsl_tmp_11 / gsl_tmp_12) + -sin(eccAnom1) * gsl_tmp_28)))) * gsl_tmp_16 - gsl_tmp_15 * (gsl_tmp_29 * (1.0 - ecc * cos(eccAnom1)) + (1.0 - ecc * cos(eccAnom1)) * gsl_tmp_29)) / (gsl_tmp_16 * gsl_tmp_16);
  let gsl_tmp_31 = -(ecc * -sin(eccAnom2) * gsl_tmp_6);
  let gsl_tmp_32 = gsl_tmp_30 - (((gsl_tmp_30 - ecc * (-sin(eccAnom2) * gsl_tmp_6 * gsl_tmp_17 + cos(eccAnom2) * gsl_tmp_30)) * (1.0 - ecc * cos(eccAnom2)) + gsl_tmp_18 * gsl_tmp_31 - ((gsl_tmp_6 - ecc * cos(eccAnom2) * gsl_tmp_6 - 1.0) * gsl_tmp_19 + (eccAnom2 - ecc * sin(eccAnom2) - meanAnomaly) * -(ecc * (-(cos(eccAnom2) * gsl_tmp_6) * gsl_tmp_17 + -sin(eccAnom2) * gsl_tmp_30)))) * gsl_tmp_21 - gsl_tmp_20 * (gsl_tmp_31 * (1.0 - ecc * cos(eccAnom2)) + (1.0 - ecc * cos(eccAnom2)) * gsl_tmp_31)) / (gsl_tmp_21 * gsl_tmp_21);
  let gsl_tmp_33 = -(ecc * -sin(eccAnom3) * gsl_tmp_7);
  let gsl_tmp_34 = gsl_tmp_32 - (((gsl_tmp_32 - ecc * (-sin(eccAnom3) * gsl_tmp_7 * (gsl_tmp_17 - gsl_tmp_20 / gsl_tmp_21) + cos(eccAnom3) * gsl_tmp_32)) * (1.0 - ecc * cos(eccAnom3)) + gsl_tmp_22 * gsl_tmp_33 - ((gsl_tmp_7 - ecc * cos(eccAnom3) * gsl_tmp_7 - 1.0) * gsl_tmp_23 + (eccAnom3 - ecc * sin(eccAnom3) - meanAnomaly) * -(ecc * (-(cos(eccAnom3) * gsl_tmp_7) * (gsl_tmp_17 - gsl_tmp_20 / gsl_tmp_21) + -sin(eccAnom3) * gsl_tmp_32)))) * gsl_tmp_25 - gsl_tmp_24 * (gsl_tmp_33 * (1.0 - ecc * cos(eccAnom3)) + (1.0 - ecc * cos(eccAnom3)) * gsl_tmp_33)) / (gsl_tmp_25 * gsl_tmp_25);
  let accPoint = center + vec3<f32>(semiMajor * (-(cos(eccAnom) * gsl_tmp_8) * gsl_tmp_26 + -sin(eccAnom) * gsl_tmp_34), semiMajor * 0.0 / (2.0 * (sqrt(1.0 - ecc * ecc) + 1.0 - select(0.0, 1.0, (sqrt(1.0 - ecc * ecc) > 0.0)))) * cos(eccAnom) * gsl_tmp_8 + semiMajor * 0.0 / (2.0 * (sqrt(1.0 - ecc * ecc) + 1.0 - select(0.0, 1.0, (sqrt(1.0 - ecc * ecc) > 0.0)))) * cos(eccAnom) * gsl_tmp_26 + semiMinor * (-sin(eccAnom) * gsl_tmp_8 * gsl_tmp_26 + cos(eccAnom) * gsl_tmp_34), 0.0) * 0.1;
  let acc = cylinderRounded(p, center, accPoint, 0.1);
  let gsl_partLo_2 = vec3<f32>(-0.5, -0.5, -0.5);
  let gsl_partHi_2 = vec3<f32>(0.5, 0.5, 0.5);
  let gsl_partBox_2 = gsl_boxDistance(p, gsl_partLo_2, gsl_partHi_2);
  var gsl_part_2 = gsl_partBox_2;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_2 <= gsl_vanishView.w) {
    let gsl_partVanish_2 = gsl_vanish(gsl_partLo_2, gsl_partHi_2);
    gsl_part_2 = 1e10;
    if (gsl_partVanish_2.x > 0.0) { gsl_part_2 = length(p - vec3<f32>(0.0, 0.0, 0.0)) - 0.5 + gsl_partVanish_2.y; }
    if (abs(gsl_part_2) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_2), gsl_partVanish_2.zw); }
  }
  let focusMarker = gsl_part_2;
  let ellipseQ = vec2<f32>((p.x + focusOffset) / semiMajor, p.y / semiMinor);
  let ellipseRing = (length(ellipseQ) - 1.0) * semiMinor;
  let ellipseTrack = abs(ellipseRing) - 0.04;
  let ellipseSlab = abs(p.z) - 0.04;
  let ellipse = max(ellipseTrack, ellipseSlab);
  let gsl_s1 = gsl_shape(body, 1u);
  let gsl_s2 = gsl_shape(tangent, 2u);
  let gsl_s3 = gsl_shape(acc, 3u);
  let gsl_s4 = gsl_shape(focusMarker, 4u);
  let gsl_s5 = gsl_shape(ellipse, 5u);
  return gsl_min(gsl_min(gsl_min(gsl_min(gsl_s1, gsl_s2), gsl_s3), gsl_s4), gsl_s5);
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
