fn material_withRoughness(m: mat4x4<f32>, rough: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  let surface = m[0];
  return mat4x4<f32>(vec4<f32>(rough, surface.y, surface.z, surface.w), m[1], m[2], m[3]);
}

fn material_withLay(m: mat4x4<f32>, direction: vec3<f32>, anisotropy: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  return mat4x4<f32>(m[0], m[1], m[2], vec4<f32>(direction.x, direction.y, direction.z, anisotropy));
}

fn ease_quadratic(t: f32, noiseTex: texture_3d<f32>) -> f32 {
  return t * t * (3.0 - 2.0 * t);
}

fn texture_mask(field: f32, edge: f32, width: f32, noiseTex: texture_3d<f32>) -> f32 {
  let t = clamp((field - edge) / width * 0.5 + 0.5, 0.0, 1.0);
  return ease_quadratic(t, noiseTex);
}

fn texture_fadeDetail(size: f32, noiseTex: texture_3d<f32>) -> f32 {
  return 1.0 - texture_mask(gsl_pw, 0.375 * size, 0.125 * size, noiseTex);
}

fn material_withDisplacement(m: mat4x4<f32>, height: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  let volume = m[2];
  return mat4x4<f32>(m[0], m[1], vec4<f32>(volume.x, volume.y, height, volume.w), m[3]);
}

fn builtin_coneCapped(p: vec3<f32>, pointA: vec3<f32>, pointB: vec3<f32>, radiusA: f32, radiusB: f32, noiseTex: texture_3d<f32>) -> f32 {
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

fn builtin_sphere(p: vec3<f32>, radius: f32, noiseTex: texture_3d<f32>) -> f32 {
  return length(p) - radius;
}

fn noise_cellularF1F2CellLattice(q: vec3<f32>, noiseTex: texture_3d<f32>) -> gsl_struct {
  var gsl_result : gsl_struct;
  gsl_result.v_1 = gsl_cellularScan(q, 1, noiseTex).f1;
  gsl_result.v_2 = gsl_cellularScan(q, 1, noiseTex).f2;
  gsl_result.v_3 = gsl_cellularScan(q, 1, noiseTex).nearCell;
  return gsl_result;
}

fn noise_cellularCell(p: vec3<f32>, size: f32, noiseTex: texture_3d<f32>) -> gsl_struct {
  let gsl_result = noise_cellularF1F2CellLattice(p / size, noiseTex);
  let f1 = gsl_result.v_1;
  let f2 = gsl_result.v_2;
  let cell = gsl_result.v_3;
  var gsl_result_2 : gsl_struct;
  gsl_result_2.v_1 = f1 * size;
  gsl_result_2.v_2 = f2 * size;
  gsl_result_2.v_3 = cell;
  return gsl_result_2;
}

fn builtin_differenceSharp(gsl_var_target: f32, tool: f32, noiseTex: texture_3d<f32>) -> f32 {
  return max(gsl_var_target, -tool);
}

fn gsl_remap(p: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  return builtin_differenceSharp(builtin_gcl_closure_boxCentered(p, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(20.0, 20.0, 20.0), noiseTex) - 2.0, builtin_gcl_closure_cylinderCapped(p, vec3<f32>(0.0, -13.0, 0.0), vec3<f32>(0.0, 13.0, 0.0), 7.0, noiseTex), noiseTex);
}

fn builtin_intersectSharp(fields: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  return max(max(fields.x, fields.y), fields.z);
}

fn builtin_intersectRound(fields: vec3<f32>, radius: f32, noiseTex: texture_3d<f32>) -> f32 {
  let sumOfSquares = 0.0 + max(fields.x + radius, 0.0) * max(fields.x + radius, 0.0) + max(fields.y + radius, 0.0) * max(fields.y + radius, 0.0) + max(fields.z + radius, 0.0) * max(fields.z + radius, 0.0);
  return min(-radius, builtin_intersectSharp(fields, noiseTex)) + sqrt(sumOfSquares);
}

fn builtin_gcl_closure_intersectRound(fields: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  return builtin_intersectRound(fields, 0.0, noiseTex);
}

fn builtin_gcl_closure_boxCentered(p: vec3<f32>, center: vec3<f32>, size: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let halfSize = size * 0.5;
  let d = abs(p - center) - halfSize;
  let solid = builtin_gcl_closure_intersectRound(vec3<f32>(d.x, d.y, d.z), noiseTex);
  return solid;
}

fn builtin_intersectSharp_2(fields: vec2<f32>, noiseTex: texture_3d<f32>) -> f32 {
  return max(fields.x, fields.y);
}

fn builtin_intersectRound_2(fields: vec2<f32>, radius: f32, noiseTex: texture_3d<f32>) -> f32 {
  let sumOfSquares = 0.0 + max(fields.x + radius, 0.0) * max(fields.x + radius, 0.0) + max(fields.y + radius, 0.0) * max(fields.y + radius, 0.0);
  return min(-radius, builtin_intersectSharp_2(fields, noiseTex)) + sqrt(sumOfSquares);
}

fn builtin_gcl_closure_intersectRound_2(fields: vec2<f32>, noiseTex: texture_3d<f32>) -> f32 {
  return builtin_intersectRound_2(fields, 0.0, noiseTex);
}

fn builtin_gcl_closure_cylinderCapped(p: vec3<f32>, pointA: vec3<f32>, pointB: vec3<f32>, radius: f32, noiseTex: texture_3d<f32>) -> f32 {
  let pa = p - pointA;
  let ba = pointB - pointA;
  let paba = dot(pa, ba);
  let baba = dot(ba, ba);
  let t = clamp(paba / baba, 0.0, 1.0);
  let radialDistance = length(pa - ba * t) - radius;
  let axialDistance = (abs(paba - baba * 0.5) - baba * 0.5) / sqrt(baba);
  let form = builtin_gcl_closure_intersectRound_2(vec2<f32>(radialDistance, axialDistance), noiseTex);
  return form;
}

fn material_translucent(m: mat4x4<f32>, transmission: f32, absorptionDistance: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  let volume = m[2];
  return mat4x4<f32>(m[0], m[1], vec4<f32>(transmission, absorptionDistance, volume.z, volume.w), m[3]);
}

fn material_withTransmission(m: mat4x4<f32>, transmission: f32, absorptionDistance: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  return material_translucent(m, transmission, absorptionDistance, noiseTex);
}

fn texture_fadeTo(gsl_var_value: f32, mean: f32, size: f32, noiseTex: texture_3d<f32>) -> f32 {
  return mean + (gsl_var_value - mean) * texture_fadeDetail(size, noiseTex);
}

fn printed_layerLay(p: vec3<f32>, buildDir: vec3<f32>, center: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  return cross(normalize(buildDir), p - center);
}

fn printed_layerFinish(p: vec3<f32>, baseMaterial: mat4x4<f32>, up: vec3<f32>, center: vec3<f32>, layerHeight: f32, relief: f32, edgeAlpha: f32, slope: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  let farAcross = sqrt(edgeAlpha * edgeAlpha + slope * slope);
  let layDir = printed_layerLay(p, up, center, noiseTex);
  let farStretch = (1.0 - clamp(edgeAlpha / farAcross, 0.1, 1.0)) / 0.9;
  let coarse = material_withLay(material_withRoughness(baseMaterial, sqrt(sqrt(edgeAlpha * farAcross)), noiseTex), layDir, farStretch, noiseTex);
  let size = min(layerHeight, 0.15);
  var gsl_detailValue = coarse;
  if (max(gsl_pw, gsl_detailPw) < 0.5 * (size)) { let gsl_detailOuterPw = gsl_pw; gsl_pw = max(gsl_pw, gsl_detailPw);
    let shown = texture_fadeDetail(size, noiseTex) * gsl_dispDrawn;
    let across = sqrt(edgeAlpha * edgeAlpha + (1.0 - shown) * slope * slope);
    let stretch = (1.0 - clamp(edgeAlpha / across, 0.1, 1.0)) / 0.9;
    let fineLay = select(vec3<f32>(0.0, 0.0, 0.0), layDir, stretch > 0.0);
    let fine = material_withDisplacement(material_withLay(material_withRoughness(baseMaterial, sqrt(sqrt(edgeAlpha * across)), noiseTex), fineLay, stretch, noiseTex), relief * texture_fadeDetail(size, noiseTex), noiseTex);
    gsl_detailValue = fine;
    gsl_pw = gsl_detailOuterPw;
  }
  var gsl_simpleValue = coarse;
  if (gsl_simpleDetail > 0.0) {
    let gsl_simpleFull = gsl_detailValue;
    gsl_simpleValue = mat4x4<f32>(mix(gsl_simpleValue[0], gsl_simpleFull[0], gsl_simpleDetail), mix(gsl_simpleValue[1], gsl_simpleFull[1], gsl_simpleDetail), mix(gsl_simpleValue[2], gsl_simpleFull[2], gsl_simpleDetail), mix(gsl_simpleValue[3], gsl_simpleFull[3], gsl_simpleDetail));
  }
  let finish = gsl_simpleValue;
  return finish;
}

fn gsl_shapeParts(p: vec3<f32>, time: f32, noiseTex: texture_3d<f32>) -> array<f32, 8> {
  gsl_partNear = vec3<f32>(1e10, 1e10, 1e10);
  let gsl_partLo = vec3<f32>(-56.4, 8.6, -31.4);
  let gsl_partHi = vec3<f32>(-23.6, 41.4, 31.4);
  let gsl_partBox = gsl_boxDistance(p, gsl_partLo, gsl_partHi);
  var gsl_part = gsl_partBox;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox <= gsl_vanishView.w) {
    let gsl_partVanish = gsl_vanish(gsl_partLo, gsl_partHi);
    gsl_part = 1e10;
    if (gsl_partVanish.x > 0.0) { gsl_part = builtin_coneCapped(p - vec3<f32>(-40.0, 25.0, 0.0), vec3<f32>(0.0, 0.0, -15.0), vec3<f32>(0.0, 0.0, 15.0), 16.4, 0.0, noiseTex) + gsl_partVanish.y; }
    if (abs(gsl_part) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part), gsl_partVanish.zw); }
  }
  let plaCone = gsl_part;
  let gsl_partLo_2 = vec3<f32>(-24.0, 11.0, -14.0);
  let gsl_partHi_2 = vec3<f32>(4.0, 39.0, 14.0);
  let gsl_partBox_2 = gsl_boxDistance(p, gsl_partLo_2, gsl_partHi_2);
  var gsl_part_2 = gsl_partBox_2;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_2 <= gsl_vanishView.w) {
    let gsl_partVanish_2 = gsl_vanish(gsl_partLo_2, gsl_partHi_2);
    gsl_part_2 = 1e10;
    if (gsl_partVanish_2.x > 0.0) { gsl_part_2 = builtin_sphere(p - vec3<f32>(-10.0, 25.0, 0.0), 14.0, noiseTex) + gsl_partVanish_2.y; }
    if (abs(gsl_part_2) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_2), gsl_partVanish_2.zw); }
  }
  let petgBall = gsl_part_2;
  let gsl_partLo_3 = vec3<f32>(8.0, 13.0, -12.0);
  let gsl_partHi_3 = vec3<f32>(32.0, 37.0, 12.0);
  let gsl_partBox_3 = gsl_boxDistance(p, gsl_partLo_3, gsl_partHi_3);
  var gsl_part_3 = gsl_partBox_3;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_3 <= gsl_vanishView.w) {
    let gsl_partVanish_3 = gsl_vanish(gsl_partLo_3, gsl_partHi_3);
    gsl_part_3 = 1e10;
    if (gsl_partVanish_3.x > 0.0) { gsl_part_3 = builtin_sphere(p - vec3<f32>(20.0, 25.0, 0.0), 12.0, noiseTex) + gsl_partVanish_3.y; }
    if (abs(gsl_part_3) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_3), gsl_partVanish_3.zw); }
  }
  let greyBall = gsl_part_3;
  let gsl_partLo_4 = vec3<f32>(36.0, 13.0, -12.0);
  let gsl_partHi_4 = vec3<f32>(60.0, 37.0, 12.0);
  let gsl_partBox_4 = gsl_boxDistance(p, gsl_partLo_4, gsl_partHi_4);
  var gsl_part_4 = gsl_partBox_4;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_4 <= gsl_vanishView.w) {
    let gsl_partVanish_4 = gsl_vanish(gsl_partLo_4, gsl_partHi_4);
    gsl_part_4 = 1e10;
    if (gsl_partVanish_4.x > 0.0) { gsl_part_4 = builtin_sphere(p - vec3<f32>(48.0, 25.0, 0.0), 12.0, noiseTex) + gsl_partVanish_4.y; }
    if (abs(gsl_part_4) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_4), gsl_partVanish_4.zw); }
  }
  let clearBall = gsl_part_4;
  let gsl_partLo_5 = vec3<f32>(-47.0, -37.0, -12.0);
  let gsl_partHi_5 = vec3<f32>(-23.0, -13.0, 12.0);
  let gsl_partBox_5 = gsl_boxDistance(p, gsl_partLo_5, gsl_partHi_5);
  var gsl_part_5 = gsl_partBox_5;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_5 <= gsl_vanishView.w) {
    let gsl_partVanish_5 = gsl_vanish(gsl_partLo_5, gsl_partHi_5);
    gsl_part_5 = 1e10;
    if (gsl_partVanish_5.x > 0.0) { gsl_part_5 = gsl_remap(p - vec3<f32>(-35.0, -25.0, 0.0), noiseTex) + gsl_partVanish_5.y; }
    if (abs(gsl_part_5) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_5), gsl_partVanish_5.zw); }
  }
  let slsBlock = gsl_part_5;
  let gsl_partLo_6 = vec3<f32>(-17.0, -37.0, -12.0);
  let gsl_partHi_6 = vec3<f32>(7.0, -13.0, 12.0);
  let gsl_partBox_6 = gsl_boxDistance(p, gsl_partLo_6, gsl_partHi_6);
  var gsl_part_6 = gsl_partBox_6;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_6 <= gsl_vanishView.w) {
    let gsl_partVanish_6 = gsl_vanish(gsl_partLo_6, gsl_partHi_6);
    gsl_part_6 = 1e10;
    if (gsl_partVanish_6.x > 0.0) { gsl_part_6 = gsl_remap(p - vec3<f32>(-5.0, -25.0, 0.0), noiseTex) + gsl_partVanish_6.y; }
    if (abs(gsl_part_6) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_6), gsl_partVanish_6.zw); }
  }
  let mjfBlock = gsl_part_6;
  let gsl_partLo_7 = vec3<f32>(13.0, -37.0, -12.0);
  let gsl_partHi_7 = vec3<f32>(37.0, -13.0, 12.0);
  let gsl_partBox_7 = gsl_boxDistance(p, gsl_partLo_7, gsl_partHi_7);
  var gsl_part_7 = gsl_partBox_7;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_7 <= gsl_vanishView.w) {
    let gsl_partVanish_7 = gsl_vanish(gsl_partLo_7, gsl_partHi_7);
    gsl_part_7 = 1e10;
    if (gsl_partVanish_7.x > 0.0) { gsl_part_7 = builtin_sphere(p - vec3<f32>(25.0, -25.0, 0.0), 12.0, noiseTex) + gsl_partVanish_7.y; }
    if (abs(gsl_part_7) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_7), gsl_partVanish_7.zw); }
  }
  let tiBall = gsl_part_7;
  let gsl_partLo_8 = vec3<f32>(33.6, -41.4, -31.4);
  let gsl_partHi_8 = vec3<f32>(66.4, -8.6, 31.4);
  let gsl_partBox_8 = gsl_boxDistance(p, gsl_partLo_8, gsl_partHi_8);
  var gsl_part_8 = gsl_partBox_8;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_8 <= gsl_vanishView.w) {
    let gsl_partVanish_8 = gsl_vanish(gsl_partLo_8, gsl_partHi_8);
    gsl_part_8 = 1e10;
    if (gsl_partVanish_8.x > 0.0) { gsl_part_8 = builtin_coneCapped(p - vec3<f32>(50.0, -25.0, 0.0), vec3<f32>(0.0, 0.0, -15.0), vec3<f32>(0.0, 0.0, 15.0), 16.4, 0.0, noiseTex) + gsl_partVanish_8.y; }
    if (abs(gsl_part_8) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_8), gsl_partVanish_8.zw); }
  }
  let tiCone = gsl_part_8;
  return array<f32, 8>(plaCone, petgBall, greyBall, clearBall, slsBlock, mjfBlock, tiBall, tiCone);
}

fn sdShape(p: vec3<f32>, time: f32, noiseTex: texture_3d<f32>) -> gsl_shape {
  let gsl_d = gsl_shapeParts(p, time, noiseTex);
  let gsl_s1 = gsl_shape(gsl_d[0], 1u);
  let gsl_s2 = gsl_shape(gsl_d[1], 2u);
  let gsl_s3 = gsl_shape(gsl_d[2], 3u);
  let gsl_s4 = gsl_shape(gsl_d[3], 4u);
  let gsl_s5 = gsl_shape(gsl_d[4], 5u);
  let gsl_s6 = gsl_shape(gsl_d[5], 6u);
  let gsl_s7 = gsl_shape(gsl_d[6], 7u);
  let gsl_s8 = gsl_shape(gsl_d[7], 8u);
  return gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_s1, gsl_s2), gsl_s3), gsl_s4), gsl_s5), gsl_s6), gsl_s7), gsl_s8);
}
fn gsl_colorField(p: vec3<f32>, time: f32, gsl_colorIdx: u32, noiseTex: texture_3d<f32>) -> vec4<f32> {
  if (gsl_colorIdx == 1u) { return vec4<f32>(vec3<f32>(0.85, 0.33, 0.08), 1.0); }
  if (gsl_colorIdx == 2u) { return vec4<f32>(vec3<f32>(0.12, 0.3, 0.75), 1.0); }
  if (gsl_colorIdx == 3u) { return vec4<f32>(vec3<f32>(0.55, 0.57, 0.6), 1.0); }
  if (gsl_colorIdx == 4u) { return vec4<f32>(vec3<f32>(0.9, 0.93, 0.96), 1.0); }
  if (gsl_colorIdx == 5u) {
    var gsl_detailValue = vec2<f32>(0.0, 1.0);
    if (max(gsl_pw, gsl_detailPw) < 0.03) { let gsl_detailOuterPw = gsl_pw; gsl_pw = max(gsl_pw, gsl_detailPw);
      let gsl_result_2 = noise_cellularCell(p, 0.06, noiseTex);
      let grainF1 = gsl_result_2.v_1;
      let grainCell = gsl_result_2.v_3;
      gsl_detailValue = vec2<f32>((clamp(1.0 - grainF1 / 0.03, 0.0, 1.0) * clamp(1.0 - grainF1 / 0.03, 0.0, 1.0) - 0.2) * 0.018 * (1.0 - texture_mask(gsl_pw, 0.0225, 0.0075, noiseTex)), texture_fadeTo(0.9 + 0.2 * fract(textureLoad(noiseTex, vec3<i32>(floor(grainCell + vec3<f32>(23.0, 61.0, 7.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((grainCell + vec3<f32>(23.0, 61.0, 7.0)) / 64.0)) & vec3<i32>(63), 0).rgb).x, 1.0, 0.06, noiseTex));
      gsl_pw = gsl_detailOuterPw;
    }
    var gsl_simpleValue = vec3<f32>(0.86, 0.85, 0.81);
    if (gsl_simpleDetail > 0.0) {
      gsl_simpleValue = mix(gsl_simpleValue, vec3<f32>(0.86, 0.85, 0.81) * gsl_detailValue.y, gsl_simpleDetail);
    }
    let slsColor = gsl_simpleValue;
    return vec4<f32>(slsColor, 1.0);
  }
  if (gsl_colorIdx == 6u) {
    var gsl_detailValue_2 = vec2<f32>(0.0, 1.0);
    if (max(gsl_pw, gsl_detailPw) < 0.03) { let gsl_detailOuterPw_2 = gsl_pw; gsl_pw = max(gsl_pw, gsl_detailPw);
      let gsl_result_3 = noise_cellularCell(p, 0.06, noiseTex);
      let grainF1_2 = gsl_result_3.v_1;
      let grainCell_2 = gsl_result_3.v_3;
      gsl_detailValue_2 = vec2<f32>((clamp(1.0 - grainF1_2 / 0.03, 0.0, 1.0) * clamp(1.0 - grainF1_2 / 0.03, 0.0, 1.0) - 0.2) * 0.018 * (1.0 - texture_mask(gsl_pw, 0.0225, 0.0075, noiseTex)), texture_fadeTo(0.9 + 0.2 * fract(textureLoad(noiseTex, vec3<i32>(floor(grainCell_2 + vec3<f32>(23.0, 61.0, 7.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((grainCell_2 + vec3<f32>(23.0, 61.0, 7.0)) / 64.0)) & vec3<i32>(63), 0).rgb).x, 1.0, 0.06, noiseTex));
      gsl_pw = gsl_detailOuterPw_2;
    }
    var gsl_simpleValue_2 = vec3<f32>(0.07, 0.07, 0.075);
    if (gsl_simpleDetail > 0.0) {
      gsl_simpleValue_2 = mix(gsl_simpleValue_2, vec3<f32>(0.07, 0.07, 0.075) * gsl_detailValue_2.y, gsl_simpleDetail);
    }
    let mjfColor = gsl_simpleValue_2;
    return vec4<f32>(mjfColor, 1.0);
  }
  if (gsl_colorIdx == 7u) {
    var gsl_detailValue_3 = 1.0;
    if (max(gsl_pw, gsl_detailPw) < 0.02) { let gsl_detailOuterPw_3 = gsl_pw; gsl_pw = max(gsl_pw, gsl_detailPw);
      gsl_detailValue_3 = texture_fadeTo(0.88 + 0.24 * fract(textureLoad(noiseTex, vec3<i32>(floor(floor(p / 0.04) + vec3<f32>(31.0, 5.0, 59.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((floor(p / 0.04) + vec3<f32>(31.0, 5.0, 59.0)) / 64.0)) & vec3<i32>(63), 0).rgb).x, 1.0, 0.04, noiseTex);
      gsl_pw = gsl_detailOuterPw_3;
    }
    var gsl_simpleValue_3 = vec3<f32>(0.646, 0.629, 0.595);
    if (gsl_simpleDetail > 0.0) {
      gsl_simpleValue_3 = mix(gsl_simpleValue_3, vec3<f32>(0.76, 0.74, 0.7) * 0.85 * gsl_detailValue_3, gsl_simpleDetail);
    }
    let tiBallColor = gsl_simpleValue_3;
    return vec4<f32>(tiBallColor, 1.0);
  }
  if (gsl_colorIdx == 8u) {
    var gsl_detailValue_4 = 1.0;
    if (max(gsl_pw, gsl_detailPw) < 0.02) { let gsl_detailOuterPw_4 = gsl_pw; gsl_pw = max(gsl_pw, gsl_detailPw);
      gsl_detailValue_4 = texture_fadeTo(0.88 + 0.24 * fract(textureLoad(noiseTex, vec3<i32>(floor(floor(p / 0.04) + vec3<f32>(31.0, 5.0, 59.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((floor(p / 0.04) + vec3<f32>(31.0, 5.0, 59.0)) / 64.0)) & vec3<i32>(63), 0).rgb).x, 1.0, 0.04, noiseTex);
      gsl_pw = gsl_detailOuterPw_4;
    }
    var gsl_simpleValue_4 = vec3<f32>(0.646, 0.629, 0.595);
    if (gsl_simpleDetail > 0.0) {
      gsl_simpleValue_4 = mix(gsl_simpleValue_4, vec3<f32>(0.76, 0.74, 0.7) * 0.85 * gsl_detailValue_4, gsl_simpleDetail);
    }
    let tiConeColor = gsl_simpleValue_4;
    return vec4<f32>(tiConeColor, 1.0);
  }
  return vec4<f32>(0.0, 0.0, 0.0, 0.0);
}
fn gsl_materialField(p: vec3<f32>, time: f32, gsl_materialIdx: u32, noiseTex: texture_3d<f32>) -> gsl_material {
  if (gsl_materialIdx == 1u) {
    let plaMat = printed_layerFinish(p, mat4x4<f32>(vec4<f32>(0.5, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), normalize(vec3<f32>(0.0, 0.0, 1.0)), vec3<f32>(-40.0, 25.0, 0.0), 0.2, 0.1 * (sqrt(max(1.0 - (2.0 * fract(dot(p, normalize(vec3<f32>(0.0, 0.0, 1.0))) / 0.2) - 1.0) * (2.0 * fract(dot(p, normalize(vec3<f32>(0.0, 0.0, 1.0))) / 0.2) - 1.0), 0.0)) - 0.785398), 0.044 * sqrt(100.0 / clamp(40.0, 1.0, 99.0) - 1.0), 0.8, noiseTex);
    let gsl_m = plaMat; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 2u) {
    let petgMat = printed_layerFinish(p, mat4x4<f32>(vec4<f32>(0.5, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), normalize(vec3<f32>(0.0, 0.0, 1.0)), vec3<f32>(-10.0, 25.0, 0.0), 0.2, 0.1 * (sqrt(max(1.0 - (2.0 * fract(dot(p, normalize(vec3<f32>(0.0, 0.0, 1.0))) / 0.2) - 1.0) * (2.0 * fract(dot(p, normalize(vec3<f32>(0.0, 0.0, 1.0))) / 0.2) - 1.0), 0.0)) - 0.785398), 0.044 * sqrt(100.0 / clamp(70.0, 1.0, 99.0) - 1.0), 0.8, noiseTex);
    let gsl_m = petgMat; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 3u) {
    let greyMat = printed_layerFinish(p, material_withTransmission(mat4x4<f32>(vec4<f32>(0.5, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), clamp(0.0, 0.0, 1.0), 20.0, noiseTex), normalize(vec3<f32>(0.0, 0.0, 1.0)), vec3<f32>(20.0, 25.0, 0.0), 0.05, (fract(dot(p, normalize(vec3<f32>(0.0, 0.0, 1.0))) / 0.05) - 0.5) * 0.004, 0.0225, 0.05, noiseTex);
    let gsl_m = greyMat; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 4u) {
    let clearMat = printed_layerFinish(p, material_withTransmission(mat4x4<f32>(vec4<f32>(0.5, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), clamp(0.9, 0.0, 1.0), 20.0, noiseTex), normalize(vec3<f32>(0.0, 0.0, 1.0)), vec3<f32>(48.0, 25.0, 0.0), 0.05, (fract(dot(p, normalize(vec3<f32>(0.0, 0.0, 1.0))) / 0.05) - 0.5) * 0.004, 0.0225, 0.05, noiseTex);
    let gsl_m = clearMat; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 5u) {
    var gsl_detailValue_5 = vec2<f32>(0.0, 1.0);
    if (max(gsl_pw, gsl_detailPw) < 0.03) { let gsl_detailOuterPw_5 = gsl_pw; gsl_pw = max(gsl_pw, gsl_detailPw);
      let gsl_result_2 = noise_cellularCell(p, 0.06, noiseTex);
      let grainF1 = gsl_result_2.v_1;
      let grainCell = gsl_result_2.v_3;
      gsl_detailValue_5 = vec2<f32>((clamp(1.0 - grainF1 / 0.03, 0.0, 1.0) * clamp(1.0 - grainF1 / 0.03, 0.0, 1.0) - 0.2) * 0.018 * (1.0 - texture_mask(gsl_pw, 0.0225, 0.0075, noiseTex)), texture_fadeTo(0.9 + 0.2 * fract(textureLoad(noiseTex, vec3<i32>(floor(grainCell + vec3<f32>(23.0, 61.0, 7.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((grainCell + vec3<f32>(23.0, 61.0, 7.0)) / 64.0)) & vec3<i32>(63), 0).rgb).x, 1.0, 0.06, noiseTex));
      gsl_pw = gsl_detailOuterPw_5;
    }
    var gsl_simpleValue_5 = mat4x4<f32>(vec4<f32>(0.85, 0.0, 1.53, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    if (gsl_simpleDetail > 0.0) {
      let gsl_simpleFull = material_withDisplacement(mat4x4<f32>(vec4<f32>(0.85, 0.0, 1.53, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_detailValue_5.x, noiseTex);
      gsl_simpleValue_5 = mat4x4<f32>(mix(gsl_simpleValue_5[0], gsl_simpleFull[0], gsl_simpleDetail), mix(gsl_simpleValue_5[1], gsl_simpleFull[1], gsl_simpleDetail), mix(gsl_simpleValue_5[2], gsl_simpleFull[2], gsl_simpleDetail), mix(gsl_simpleValue_5[3], gsl_simpleFull[3], gsl_simpleDetail));
    }
    let slsMat = gsl_simpleValue_5;
    let gsl_m = slsMat; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 6u) {
    var gsl_detailValue_6 = vec2<f32>(0.0, 1.0);
    if (max(gsl_pw, gsl_detailPw) < 0.03) { let gsl_detailOuterPw_6 = gsl_pw; gsl_pw = max(gsl_pw, gsl_detailPw);
      let gsl_result_3 = noise_cellularCell(p, 0.06, noiseTex);
      let grainF1_2 = gsl_result_3.v_1;
      let grainCell_2 = gsl_result_3.v_3;
      gsl_detailValue_6 = vec2<f32>((clamp(1.0 - grainF1_2 / 0.03, 0.0, 1.0) * clamp(1.0 - grainF1_2 / 0.03, 0.0, 1.0) - 0.2) * 0.018 * (1.0 - texture_mask(gsl_pw, 0.0225, 0.0075, noiseTex)), texture_fadeTo(0.9 + 0.2 * fract(textureLoad(noiseTex, vec3<i32>(floor(grainCell_2 + vec3<f32>(23.0, 61.0, 7.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((grainCell_2 + vec3<f32>(23.0, 61.0, 7.0)) / 64.0)) & vec3<i32>(63), 0).rgb).x, 1.0, 0.06, noiseTex));
      gsl_pw = gsl_detailOuterPw_6;
    }
    var gsl_simpleValue_6 = mat4x4<f32>(vec4<f32>(0.85, 0.0, 1.53, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    if (gsl_simpleDetail > 0.0) {
      let gsl_simpleFull_2 = material_withDisplacement(mat4x4<f32>(vec4<f32>(0.85, 0.0, 1.53, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_detailValue_6.x, noiseTex);
      gsl_simpleValue_6 = mat4x4<f32>(mix(gsl_simpleValue_6[0], gsl_simpleFull_2[0], gsl_simpleDetail), mix(gsl_simpleValue_6[1], gsl_simpleFull_2[1], gsl_simpleDetail), mix(gsl_simpleValue_6[2], gsl_simpleFull_2[2], gsl_simpleDetail), mix(gsl_simpleValue_6[3], gsl_simpleFull_2[3], gsl_simpleDetail));
    }
    let mjfMat = gsl_simpleValue_6;
    let gsl_m = mjfMat; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 7u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.81903625, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 8u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.81903625, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  return gsl_material(vec4<f32>(0.5, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), 0.0);
}
struct gsl_CellularScan {
  f1: f32,
  f2: f32,
  feature: vec3<f32>,
  farFeature: vec3<f32>,
  nearCell: vec3<f32>,
  farCell: vec3<f32>,
}

fn gsl_cellularHash(cell: vec3<f32>, oct: i32, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let fine = fract(textureLoad(noiseTex, vec3<i32>(floor(cell)) & vec3<i32>(63), 0).rgb);
  if (oct <= 1) { return fine; }
  // fract(fract(a) + b) == fract(a + b): the inner fract only removes an
  // integer, so folding the supercell on top of the already-fracted fine
  // lookup is the same value noiseHash3's two-lookup form produces.
  return fract(fine + textureLoad(noiseTex, vec3<i32>(floor(cell / 64.0)) & vec3<i32>(63), 0).rgb);
}

fn gsl_cellularScan(q: vec3<f32>, oct: i32, noiseTex: texture_3d<f32>) -> gsl_CellularScan {
  let pi = floor(q);
  let pf = q - pi;
  // Seeded with an unreachable distance, squared, because the scan compares
  // squared distances and roots the two winners at the end.
  var best = 64.000000;
  var second = 64.000000;
  var bestCell = pi;
  var bestH = vec3<f32>(0.0);
  var secondCell = pi;
  var secondH = vec3<f32>(0.0);
  for (var i = 0; i < 27; i = i + 1) {
    // The same offset order cellularOffsets() lists: x fastest, then y, then z.
    let o = vec3<f32>(f32(i % 3) - 1.0, f32((i / 3) % 3) - 1.0, f32(i / 9) - 1.0);
    let h = gsl_cellularHash(pi + o, oct, noiseTex);
    let v = pf - (o + h);
    let d2 = dot(v, v);
    if (d2 < best) {
      second = best;
      secondCell = bestCell;
      secondH = bestH;
      best = d2;
      bestCell = pi + o;
      bestH = h;
    } else if (d2 < second) {
      second = d2;
      secondCell = pi + o;
      secondH = h;
    }
  }
  var out: gsl_CellularScan;
  out.f1 = sqrt(best);
  out.f2 = sqrt(second);
  // Both closest points, so both gradients are available without a second hash.
  out.feature = bestCell + bestH;
  out.farFeature = secondCell + secondH;
  out.nearCell = bestCell;
  out.farCell = secondCell;
  return out;
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

var<private> gsl_simpleDetail: f32 = 1.0;

var<private> gsl_dispDrawn: f32 = 0.0;

var<private> gsl_detailPw: f32 = 0.0;

fn gsl_heights(p: vec3<f32>, time: f32, noiseTex: texture_3d<f32>) -> array<f32, 8> {
  let plaMat = printed_layerFinish(p, mat4x4<f32>(vec4<f32>(0.5, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), normalize(vec3<f32>(0.0, 0.0, 1.0)), vec3<f32>(-40.0, 25.0, 0.0), 0.2, 0.1 * (sqrt(max(1.0 - (2.0 * fract(dot(p, normalize(vec3<f32>(0.0, 0.0, 1.0))) / 0.2) - 1.0) * (2.0 * fract(dot(p, normalize(vec3<f32>(0.0, 0.0, 1.0))) / 0.2) - 1.0), 0.0)) - 0.785398), 0.044 * sqrt(100.0 / clamp(40.0, 1.0, 99.0) - 1.0), 0.8, noiseTex);
  let petgMat = printed_layerFinish(p, mat4x4<f32>(vec4<f32>(0.5, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), normalize(vec3<f32>(0.0, 0.0, 1.0)), vec3<f32>(-10.0, 25.0, 0.0), 0.2, 0.1 * (sqrt(max(1.0 - (2.0 * fract(dot(p, normalize(vec3<f32>(0.0, 0.0, 1.0))) / 0.2) - 1.0) * (2.0 * fract(dot(p, normalize(vec3<f32>(0.0, 0.0, 1.0))) / 0.2) - 1.0), 0.0)) - 0.785398), 0.044 * sqrt(100.0 / clamp(70.0, 1.0, 99.0) - 1.0), 0.8, noiseTex);
  let greyMat = printed_layerFinish(p, material_withTransmission(mat4x4<f32>(vec4<f32>(0.5, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), clamp(0.0, 0.0, 1.0), 20.0, noiseTex), normalize(vec3<f32>(0.0, 0.0, 1.0)), vec3<f32>(20.0, 25.0, 0.0), 0.05, (fract(dot(p, normalize(vec3<f32>(0.0, 0.0, 1.0))) / 0.05) - 0.5) * 0.004, 0.0225, 0.05, noiseTex);
  let clearMat = printed_layerFinish(p, material_withTransmission(mat4x4<f32>(vec4<f32>(0.5, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), clamp(0.9, 0.0, 1.0), 20.0, noiseTex), normalize(vec3<f32>(0.0, 0.0, 1.0)), vec3<f32>(48.0, 25.0, 0.0), 0.05, (fract(dot(p, normalize(vec3<f32>(0.0, 0.0, 1.0))) / 0.05) - 0.5) * 0.004, 0.0225, 0.05, noiseTex);
  var gsl_detailValue_7 = vec2<f32>(0.0, 1.0);
  let gsl_result_2 = noise_cellularCell(p, 0.06, noiseTex);
  let grainF1 = gsl_result_2.v_1;
  let grainCell = gsl_result_2.v_3;
  if (max(gsl_pw, gsl_detailPw) < 0.03) { let gsl_detailOuterPw_7 = gsl_pw; gsl_pw = max(gsl_pw, gsl_detailPw);
    gsl_detailValue_7 = vec2<f32>((clamp(1.0 - grainF1 / 0.03, 0.0, 1.0) * clamp(1.0 - grainF1 / 0.03, 0.0, 1.0) - 0.2) * 0.018 * (1.0 - texture_mask(gsl_pw, 0.0225, 0.0075, noiseTex)), texture_fadeTo(0.9 + 0.2 * fract(textureLoad(noiseTex, vec3<i32>(floor(grainCell + vec3<f32>(23.0, 61.0, 7.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((grainCell + vec3<f32>(23.0, 61.0, 7.0)) / 64.0)) & vec3<i32>(63), 0).rgb).x, 1.0, 0.06, noiseTex));
    gsl_pw = gsl_detailOuterPw_7;
  }
  var gsl_simpleValue_7 = mat4x4<f32>(vec4<f32>(0.85, 0.0, 1.53, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
  if (gsl_simpleDetail > 0.0) {
    let gsl_simpleFull_3 = material_withDisplacement(mat4x4<f32>(vec4<f32>(0.85, 0.0, 1.53, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_detailValue_7.x, noiseTex);
    gsl_simpleValue_7 = mat4x4<f32>(mix(gsl_simpleValue_7[0], gsl_simpleFull_3[0], gsl_simpleDetail), mix(gsl_simpleValue_7[1], gsl_simpleFull_3[1], gsl_simpleDetail), mix(gsl_simpleValue_7[2], gsl_simpleFull_3[2], gsl_simpleDetail), mix(gsl_simpleValue_7[3], gsl_simpleFull_3[3], gsl_simpleDetail));
  }
  let slsMat = gsl_simpleValue_7;
  var gsl_detailValue_8 = vec2<f32>(0.0, 1.0);
  let gsl_result_3 = noise_cellularCell(p, 0.06, noiseTex);
  let grainF1_2 = gsl_result_3.v_1;
  let grainCell_2 = gsl_result_3.v_3;
  if (max(gsl_pw, gsl_detailPw) < 0.03) { let gsl_detailOuterPw_8 = gsl_pw; gsl_pw = max(gsl_pw, gsl_detailPw);
    gsl_detailValue_8 = vec2<f32>((clamp(1.0 - grainF1_2 / 0.03, 0.0, 1.0) * clamp(1.0 - grainF1_2 / 0.03, 0.0, 1.0) - 0.2) * 0.018 * (1.0 - texture_mask(gsl_pw, 0.0225, 0.0075, noiseTex)), texture_fadeTo(0.9 + 0.2 * fract(textureLoad(noiseTex, vec3<i32>(floor(grainCell_2 + vec3<f32>(23.0, 61.0, 7.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((grainCell_2 + vec3<f32>(23.0, 61.0, 7.0)) / 64.0)) & vec3<i32>(63), 0).rgb).x, 1.0, 0.06, noiseTex));
    gsl_pw = gsl_detailOuterPw_8;
  }
  var gsl_simpleValue_8 = mat4x4<f32>(vec4<f32>(0.85, 0.0, 1.53, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
  if (gsl_simpleDetail > 0.0) {
    let gsl_simpleFull_4 = material_withDisplacement(mat4x4<f32>(vec4<f32>(0.85, 0.0, 1.53, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_detailValue_8.x, noiseTex);
    gsl_simpleValue_8 = mat4x4<f32>(mix(gsl_simpleValue_8[0], gsl_simpleFull_4[0], gsl_simpleDetail), mix(gsl_simpleValue_8[1], gsl_simpleFull_4[1], gsl_simpleDetail), mix(gsl_simpleValue_8[2], gsl_simpleFull_4[2], gsl_simpleDetail), mix(gsl_simpleValue_8[3], gsl_simpleFull_4[3], gsl_simpleDetail));
  }
  let mjfMat = gsl_simpleValue_8;
  return array<f32, 8>(plaMat[2].z, petgMat[2].z, greyMat[2].z, clearMat[2].z, slsMat[2].z, mjfMat[2].z, 0.0, 0.0);
}

fn gsl_sdShapeDisplaced(p: vec3<f32>, time: f32, gsl_dispK: f32, noiseTex: texture_3d<f32>) -> gsl_shape {
  let gsl_d = gsl_shapeParts(p, time, noiseTex);
  let gsl_h = gsl_heights(p, time, noiseTex);
  let gsl_s1 = gsl_shape(gsl_d[0] - gsl_h[0] * gsl_dispK, 1u);
  let gsl_s2 = gsl_shape(gsl_d[1] - gsl_h[1] * gsl_dispK, 2u);
  let gsl_s3 = gsl_shape(gsl_d[2] - gsl_h[2] * gsl_dispK, 3u);
  let gsl_s4 = gsl_shape(gsl_d[3] - gsl_h[3] * gsl_dispK, 4u);
  let gsl_s5 = gsl_shape(gsl_d[4] - gsl_h[4] * gsl_dispK, 5u);
  let gsl_s6 = gsl_shape(gsl_d[5] - gsl_h[5] * gsl_dispK, 6u);
  let gsl_s7 = gsl_shape(gsl_d[6], 7u);
  let gsl_s8 = gsl_shape(gsl_d[7], 8u);
  return gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_s1, gsl_s2), gsl_s3), gsl_s4), gsl_s5), gsl_s6), gsl_s7), gsl_s8);
}

struct gsl_struct {
  v_1 : f32,
  v_2 : f32,
  v_3 : vec3<f32>,
}
