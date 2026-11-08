fn builtin_axisBasis(direction: vec3<f32>) -> gsl_struct {
  let n = normalize(direction);
  let sz = sign(n.z);
  let s = sz + 1.0 - abs(sz);
  let a = -1.0 / (s + n.z);
  let b = n.x * n.y * a;
  var gsl_result : gsl_struct;
  gsl_result.v_1 = vec3<f32>(1.0 + s * n.x * n.x * a, s * b, -(s * n.x));
  gsl_result.v_2 = vec3<f32>(b, s + n.y * n.y * a, -n.y);
  return gsl_result;
}

fn builtin_axisAngle(p: vec3<f32>, center: vec3<f32>, direction: vec3<f32>) -> f32 {
  let d = p - center;
  let gsl_result = builtin_axisBasis(direction);
  let t2 = gsl_result.v_2;
  let t1 = gsl_result.v_1;
  return atan2(dot(d, t2), dot(d, t1));
}

fn builtin_cylinderCapped__u(p: vec3<f32>, pointA: vec3<f32>, pointB: vec3<f32>, radius: f32) -> f32 {
  return builtin_axisAngle(p, pointA, pointB - pointA);
}

fn builtin_sphere(p: vec3<f32>, center: vec3<f32>, radius: f32) -> f32 {
  let form = length(p - center) - radius;
  return form;
}

fn builtin_torus(p: vec3<f32>, majorRadius: f32, minorRadius: f32) -> f32 {
  let radial = length(p.xy) - majorRadius;
  let form = length(vec2<f32>(radial, p.z)) - minorRadius;
  return form;
}

fn gsl_remap(p: vec3<f32>, vector_dirZ: vec3<f32>, vector_dirX: vec3<f32>) -> f32 {
  let gslCarry9 = builtin_torus(p, 4.0, 1.5);
  return gslCarry9;
}

fn builtin_cappedCylinderV(p: vec3<f32>, pointA: vec3<f32>, pointB: vec3<f32>, radius: f32) -> f32 {
  let pa = p - pointA;
  let ba = pointB - pointA;
  let paba = dot(pa, ba);
  let baba = dot(ba, ba);
  let across = length(pa - ba * paba / baba);
  let along = (paba - baba * 0.5) / sqrt(baba);
  let t = clamp(paba / baba, 0.0, 1.0);
  let radialDistance = length(pa - ba * t) - radius;
  let axialDistance = (abs(paba - baba * 0.5) - baba * 0.5) / sqrt(baba);
  return select(across, along, radialDistance >= axialDistance);
}

fn builtin_intersectSharp(fields: vec2<f32>) -> f32 {
  return max(fields.x, fields.y);
}

fn builtin_intersectRound(fields: vec2<f32>, radius: f32) -> f32 {
  let sumOfSquares = 0.0 + max(fields.x + radius, 0.0) * max(fields.x + radius, 0.0) + max(fields.y + radius, 0.0) * max(fields.y + radius, 0.0);
  return min(-radius, builtin_intersectSharp(fields)) + sqrt(sumOfSquares);
}

fn builtin_gcl_closure_intersectRound_3(fields: vec2<f32>) -> f32 {
  return builtin_intersectRound(fields, 0.0);
}

fn builtin_gcl_closure_cylinderCapped_3(p: vec3<f32>, pointA: vec3<f32>, pointB: vec3<f32>, radius: f32) -> f32 {
  let pa = p - pointA;
  let ba = pointB - pointA;
  let paba = dot(pa, ba);
  let baba = dot(ba, ba);
  let t = clamp(paba / baba, 0.0, 1.0);
  let radialDistance = length(pa - ba * t) - radius;
  let axialDistance = (abs(paba - baba * 0.5) - baba * 0.5) / sqrt(baba);
  let form = builtin_gcl_closure_intersectRound_3(vec2<f32>(radialDistance, axialDistance));
  return form;
}

fn builtin_boxU(p: vec3<f32>, center: vec3<f32>, size: vec3<f32>) -> f32 {
  let q = p - center;
  let d = abs(q) - size * 0.5;
  return select(select(q.x, q.z, d.y >= d.z), q.y, d.x >= max(d.y, d.z));
}

fn builtin_boxV(p: vec3<f32>, center: vec3<f32>, size: vec3<f32>) -> f32 {
  let q = p - center;
  let d = abs(q) - size * 0.5;
  return select(select(q.y, q.x, d.y >= d.z), q.z, d.x >= max(d.y, d.z));
}

fn builtin_intersectSharp_2(fields: vec3<f32>) -> f32 {
  return max(max(fields.x, fields.y), fields.z);
}

fn builtin_intersectRound_2(fields: vec3<f32>, radius: f32) -> f32 {
  let sumOfSquares = 0.0 + max(fields.x + radius, 0.0) * max(fields.x + radius, 0.0) + max(fields.y + radius, 0.0) * max(fields.y + radius, 0.0) + max(fields.z + radius, 0.0) * max(fields.z + radius, 0.0);
  return min(-radius, builtin_intersectSharp_2(fields)) + sqrt(sumOfSquares);
}

fn builtin_gcl_closure_intersectRound_6(fields: vec3<f32>) -> f32 {
  return builtin_intersectRound_2(fields, 0.0);
}

fn builtin_gcl_closure_boxCentered_3(p: vec3<f32>, center: vec3<f32>, size: vec3<f32>) -> f32 {
  let halfSize = size * 0.5;
  let d = abs(p - center) - halfSize;
  let solid = builtin_gcl_closure_intersectRound_6(vec3<f32>(d.x, d.y, d.z));
  return solid;
}

fn builtin_sphereAzimuth(p: vec3<f32>, center: vec3<f32>, pole: vec3<f32>, reference: vec3<f32>) -> f32 {
  let d = p - center;
  let n = normalize(pole);
  let e1 = normalize(reference - n * dot(reference, n));
  let e2 = cross(n, e1);
  return atan2(dot(d, e2), dot(d, e1));
}

fn builtin_spherePolar(p: vec3<f32>, center: vec3<f32>, pole: vec3<f32>) -> f32 {
  let d = p - center;
  let n = normalize(pole);
  let along = dot(d, n);
  return atan2(length(d - n * along), along);
}

fn material_withLay(m: mat4x4<f32>, direction: vec3<f32>, anisotropy: f32) -> mat4x4<f32> {
  return mat4x4<f32>(m[0], m[1], m[2], vec4<f32>(direction.x, direction.y, direction.z, anisotropy));
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  gsl_partNear = vec3<f32>(1e10, 1e10, 1e10);
  let gsl_partLo = vec3<f32>(-19.0, -4.0, -10.0);
  let gsl_partHi = vec3<f32>(-11.0, 4.0, 10.0);
  let gsl_partBox = gsl_boxDistance(p, gsl_partLo, gsl_partHi);
  var gsl_part = gsl_partBox;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox <= gsl_vanishView.w) {
    let gsl_partVanish = gsl_vanish(gsl_partLo, gsl_partHi);
    gsl_part = 1e10;
    if (gsl_partVanish.x > 0.0) { gsl_part = builtin_gcl_closure_cylinderCapped_3(p, vec3<f32>(-15.0, 0.0, -6.0), vec3<f32>(-15.0, 0.0, 6.0), 4.0) + gsl_partVanish.y; }
    if (abs(gsl_part) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part), gsl_partVanish.zw); }
  }
  let turned = gsl_part;
  let gsl_partLo_2 = vec3<f32>(-9.0, -4.0, -10.0);
  let gsl_partHi_2 = vec3<f32>(-1.0, 4.0, 10.0);
  let gsl_partBox_2 = gsl_boxDistance(p, gsl_partLo_2, gsl_partHi_2);
  var gsl_part_2 = gsl_partBox_2;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_2 <= gsl_vanishView.w) {
    let gsl_partVanish_2 = gsl_vanish(gsl_partLo_2, gsl_partHi_2);
    gsl_part_2 = 1e10;
    if (gsl_partVanish_2.x > 0.0) { gsl_part_2 = builtin_gcl_closure_cylinderCapped_3(p, vec3<f32>(-5.0, 0.0, -6.0), vec3<f32>(-5.0, 0.0, 6.0), 4.0) + gsl_partVanish_2.y; }
    if (abs(gsl_part_2) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_2), gsl_partVanish_2.zw); }
  }
  let brushed = gsl_part_2;
  let gsl_partLo_3 = vec3<f32>(3.0, -4.0, -4.0);
  let gsl_partHi_3 = vec3<f32>(11.0, 4.0, 4.0);
  let gsl_partBox_3 = gsl_boxDistance(p, gsl_partLo_3, gsl_partHi_3);
  var gsl_part_3 = gsl_partBox_3;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_3 <= gsl_vanishView.w) {
    let gsl_partVanish_3 = gsl_vanish(gsl_partLo_3, gsl_partHi_3);
    gsl_part_3 = 1e10;
    if (gsl_partVanish_3.x > 0.0) { gsl_part_3 = builtin_gcl_closure_boxCentered_3(p, vec3<f32>(14.0, 0.0, 0.0) * 0.5, abs(vec3<f32>(8.0, 8.0, 8.0))) + gsl_partVanish_3.y; }
    if (abs(gsl_part_3) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_3), gsl_partVanish_3.zw); }
  }
  let block = gsl_part_3;
  let gsl_partLo_4 = vec3<f32>(16.0, -4.0, -4.0);
  let gsl_partHi_4 = vec3<f32>(24.0, 4.0, 4.0);
  let gsl_partBox_4 = gsl_boxDistance(p, gsl_partLo_4, gsl_partHi_4);
  var gsl_part_4 = gsl_partBox_4;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_4 <= gsl_vanishView.w) {
    let gsl_partVanish_4 = gsl_vanish(gsl_partLo_4, gsl_partHi_4);
    gsl_part_4 = 1e10;
    if (gsl_partVanish_4.x > 0.0) { gsl_part_4 = builtin_sphere(p, vec3<f32>(20.0, 0.0, 0.0), 4.0) + gsl_partVanish_4.y; }
    if (abs(gsl_part_4) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_4), gsl_partVanish_4.zw); }
  }
  let ball = gsl_part_4;
  let vector_dirZ = vec3<f32>(0.0, 0.0, 1.0);
  let vector_dirX = vec3<f32>(1.0, 0.0, 0.0);
  let ring = gsl_remap(p - vec3<f32>(33.0, 0.0, 0.0), vector_dirZ, vector_dirX);
  let gsl_s1 = gsl_shape(turned, 1u);
  let gsl_s2 = gsl_shape(brushed, 2u);
  let gsl_s3 = gsl_shape(block, 3u);
  let gsl_s4 = gsl_shape(ball, 4u);
  let gsl_s5 = gsl_shape(ring, 5u);
  return gsl_min(gsl_min(gsl_min(gsl_min(gsl_s1, gsl_s2), gsl_s3), gsl_s4), gsl_s5);
}

fn gsl_colorField(p: vec3<f32>, time: f32, gsl_colorIdx: u32) -> vec4<f32> {
  if (gsl_colorIdx == 1u) { return vec4<f32>(vec3<f32>(0.8373, 0.8204, 0.7966), 1.0); }
  if (gsl_colorIdx == 2u) { return vec4<f32>(vec3<f32>(0.8373, 0.8204, 0.7966), 1.0); }
  if (gsl_colorIdx == 3u) { return vec4<f32>(vec3<f32>(0.8373, 0.8204, 0.7966), 1.0); }
  if (gsl_colorIdx == 4u) { return vec4<f32>(vec3<f32>(0.8373, 0.8204, 0.7966), 1.0); }
  if (gsl_colorIdx == 5u) { return vec4<f32>(vec3<f32>(0.8373, 0.8204, 0.7966), 1.0); }
  return vec4<f32>(0.0, 0.0, 0.0, 0.0);
}
fn gsl_materialField(p: vec3<f32>, time: f32, gsl_materialIdx: u32) -> gsl_material {
  if (gsl_materialIdx == 1u) {
    let gsl_tmp_31 = dot(p - vec3<f32>(-15.0, 0.0, -6.0), vec3<f32>(1.0, -0.0, -0.0));
    let gsl_tmp_32 = dot(p - vec3<f32>(-15.0, 0.0, -6.0), vec3<f32>(-0.0, 1.0, -0.0));
    let gsl_tmp_33 = dot(p - vec3<f32>(-15.0, 0.0, -6.0), vec3<f32>(1.0, -0.0, -0.0));
    let gsl_tmp_34 = dot(p - vec3<f32>(-15.0, 0.0, -6.0), vec3<f32>(-0.0, 1.0, -0.0));
    let gsl_tmp_35 = dot(p - vec3<f32>(-15.0, 0.0, -6.0), vec3<f32>(1.0, -0.0, -0.0));
    let gsl_tmp_36 = dot(p - vec3<f32>(-15.0, 0.0, -6.0), vec3<f32>(-0.0, 1.0, -0.0));
    let gsl_m = material_withLay(mat4x4<f32>(vec4<f32>(0.35, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), vec3<f32>((gsl_tmp_31 * (dot(p - vec3<f32>(-15.0, 0.0, -6.0), vec3<f32>(0.0, 0.0, -0.0)) + 0.0) - gsl_tmp_32 * (dot(p - vec3<f32>(-15.0, 0.0, -6.0), vec3<f32>(0.0, 0.0, -0.0)) + 1.0)) / (pow(gsl_tmp_31, 2.0) + pow(gsl_tmp_32, 2.0)), (gsl_tmp_33 * (dot(p - vec3<f32>(-15.0, 0.0, -6.0), vec3<f32>(0.0, 0.0, -0.0)) + 1.0) - gsl_tmp_34 * (dot(p - vec3<f32>(-15.0, 0.0, -6.0), vec3<f32>(0.0, 0.0, -0.0)) + 0.0)) / (pow(gsl_tmp_33, 2.0) + pow(gsl_tmp_34, 2.0)), (gsl_tmp_35 * (dot(p - vec3<f32>(-15.0, 0.0, -6.0), vec3<f32>(0.0, 0.0, -0.0)) + 0.0) - gsl_tmp_36 * (dot(p - vec3<f32>(-15.0, 0.0, -6.0), vec3<f32>(0.0, 0.0, -0.0)) + 0.0)) / (pow(gsl_tmp_35, 2.0) + pow(gsl_tmp_36, 2.0))), 0.8); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 2u) {
    let gsl_tmp_37 = dot(p - vec3<f32>(-5.0, 0.0, -6.0), vec3<f32>(0.0, 0.0, 12.0));
    let gsl_tmp_39 = p - vec3<f32>(-5.0, 0.0, -6.0) - vec3<f32>(0.0, 0.0, 12.0) * gsl_tmp_37 / 144.0;
    let gsl_tmp_38 = dot(p - vec3<f32>(-5.0, 0.0, -6.0), vec3<f32>(0.0, 0.0, 0.0)) + 0.0;
    let gsl_tmp_40 = dot(p - vec3<f32>(-5.0, 0.0, -6.0), vec3<f32>(0.0, 0.0, 12.0));
    let gsl_tmp_42 = p - vec3<f32>(-5.0, 0.0, -6.0) - vec3<f32>(0.0, 0.0, 12.0) * gsl_tmp_40 / 144.0;
    let gsl_tmp_41 = dot(p - vec3<f32>(-5.0, 0.0, -6.0), vec3<f32>(0.0, 0.0, 0.0)) + 0.0;
    let gsl_tmp_43 = dot(p - vec3<f32>(-5.0, 0.0, -6.0), vec3<f32>(0.0, 0.0, 12.0));
    let gsl_tmp_45 = p - vec3<f32>(-5.0, 0.0, -6.0) - vec3<f32>(0.0, 0.0, 12.0) * gsl_tmp_43 / 144.0;
    let gsl_tmp_44 = dot(p - vec3<f32>(-5.0, 0.0, -6.0), vec3<f32>(0.0, 0.0, 0.0)) + 12.0;
    let gsl_m = material_withLay(mat4x4<f32>(vec4<f32>(0.35, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), vec3<f32>(select(dot(gsl_tmp_39, vec3<f32>(1.0, 0.0, 0.0) - (vec3<f32>(0.0, 0.0, 0.0) * gsl_tmp_37 / 144.0 + vec3<f32>(0.0, 0.0, 12.0) * (gsl_tmp_38 * 144.0 - gsl_tmp_37 * 0.0) / 20736.0)) / (length(gsl_tmp_39) + 1.0 - select(0.0, 1.0, (length(gsl_tmp_39) > 0.0))), ((gsl_tmp_38 - 0.0) * 12.0 - (gsl_tmp_37 - 72.0) * 0.0) / 144.0, length(p - vec3<f32>(-5.0, 0.0, -6.0) - vec3<f32>(0.0, 0.0, 12.0) * clamp(gsl_tmp_37 / 144.0, 0.0, 1.0)) - 4.0 >= (abs(gsl_tmp_37 - 72.0) - 72.0) / 12.0), select(dot(gsl_tmp_42, vec3<f32>(0.0, 1.0, 0.0) - (vec3<f32>(0.0, 0.0, 0.0) * gsl_tmp_40 / 144.0 + vec3<f32>(0.0, 0.0, 12.0) * (gsl_tmp_41 * 144.0 - gsl_tmp_40 * 0.0) / 20736.0)) / (length(gsl_tmp_42) + 1.0 - select(0.0, 1.0, (length(gsl_tmp_42) > 0.0))), ((gsl_tmp_41 - 0.0) * 12.0 - (gsl_tmp_40 - 72.0) * 0.0) / 144.0, length(p - vec3<f32>(-5.0, 0.0, -6.0) - vec3<f32>(0.0, 0.0, 12.0) * clamp(gsl_tmp_40 / 144.0, 0.0, 1.0)) - 4.0 >= (abs(gsl_tmp_40 - 72.0) - 72.0) / 12.0), select(dot(gsl_tmp_45, vec3<f32>(0.0, 0.0, 1.0) - (vec3<f32>(0.0, 0.0, 0.0) * gsl_tmp_43 / 144.0 + vec3<f32>(0.0, 0.0, 12.0) * (gsl_tmp_44 * 144.0 - gsl_tmp_43 * 0.0) / 20736.0)) / (length(gsl_tmp_45) + 1.0 - select(0.0, 1.0, (length(gsl_tmp_45) > 0.0))), ((gsl_tmp_44 - 0.0) * 12.0 - (gsl_tmp_43 - 72.0) * 0.0) / 144.0, length(p - vec3<f32>(-5.0, 0.0, -6.0) - vec3<f32>(0.0, 0.0, 12.0) * clamp(gsl_tmp_43 / 144.0, 0.0, 1.0)) - 4.0 >= (abs(gsl_tmp_43 - 72.0) - 72.0) / 12.0)), 0.8); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 3u) {
    let gsl_tmp_46 = abs(p - vec3<f32>(7.0, 0.0, 0.0)) - vec3<f32>(4.0, 4.0, 4.0);
    let gsl_tmp_47 = abs(p - vec3<f32>(7.0, 0.0, 0.0)) - vec3<f32>(4.0, 4.0, 4.0);
    let gsl_tmp_48 = abs(p - vec3<f32>(7.0, 0.0, 0.0)) - vec3<f32>(4.0, 4.0, 4.0);
    let gsl_m = material_withLay(mat4x4<f32>(vec4<f32>(0.35, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), vec3<f32>(select(select(1.0, 0.0, gsl_tmp_46.y >= gsl_tmp_46.z), 0.0, gsl_tmp_46.x >= max(gsl_tmp_46.y, gsl_tmp_46.z)), select(select(0.0, 0.0, gsl_tmp_47.y >= gsl_tmp_47.z), 1.0, gsl_tmp_47.x >= max(gsl_tmp_47.y, gsl_tmp_47.z)), select(select(0.0, 1.0, gsl_tmp_48.y >= gsl_tmp_48.z), 0.0, gsl_tmp_48.x >= max(gsl_tmp_48.y, gsl_tmp_48.z))), 0.8); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 4u) {
    let gsl_tmp_49 = dot(p - vec3<f32>(20.0, 0.0, 0.0), vec3<f32>(1.0, 0.0, 0.0));
    let gsl_tmp_50 = dot(p - vec3<f32>(20.0, 0.0, 0.0), vec3<f32>(0.0, 1.0, 0.0));
    let gsl_tmp_51 = dot(p - vec3<f32>(20.0, 0.0, 0.0), vec3<f32>(1.0, 0.0, 0.0));
    let gsl_tmp_52 = dot(p - vec3<f32>(20.0, 0.0, 0.0), vec3<f32>(0.0, 1.0, 0.0));
    let gsl_tmp_53 = dot(p - vec3<f32>(20.0, 0.0, 0.0), vec3<f32>(1.0, 0.0, 0.0));
    let gsl_tmp_54 = dot(p - vec3<f32>(20.0, 0.0, 0.0), vec3<f32>(0.0, 1.0, 0.0));
    let gsl_m = material_withLay(mat4x4<f32>(vec4<f32>(0.35, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), vec3<f32>((gsl_tmp_49 * 0.0 - gsl_tmp_50 * 1.0) / (pow(gsl_tmp_49, 2.0) + pow(gsl_tmp_50, 2.0)), (gsl_tmp_51 * 1.0 - gsl_tmp_52 * 0.0) / (pow(gsl_tmp_51, 2.0) + pow(gsl_tmp_52, 2.0)), (gsl_tmp_53 * 0.0 - gsl_tmp_54 * 0.0) / (pow(gsl_tmp_53, 2.0) + pow(gsl_tmp_54, 2.0))), 0.8); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 5u) {
    let gsl_tmp_55 = (p - vec3<f32>(33.0, 0.0, 0.0)).z;
    let gsl_tmp_56 = (p - vec3<f32>(33.0, 0.0, 0.0)).xy;
    let gsl_tmp_57 = (p - vec3<f32>(33.0, 0.0, 0.0)).z;
    let gsl_tmp_58 = (p - vec3<f32>(33.0, 0.0, 0.0)).xy;
    let gsl_tmp_59 = (p - vec3<f32>(33.0, 0.0, 0.0)).z;
    let gsl_tmp_60 = (p - vec3<f32>(33.0, 0.0, 0.0)).xy;
    let gsl_m = material_withLay(mat4x4<f32>(vec4<f32>(0.35, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), vec3<f32>(1.0 * (gsl_tmp_55 * dot(gsl_tmp_56, vec2<f32>(1.0, 0.0)) / (length(gsl_tmp_56) + 1.0 - select(0.0, 1.0, (length(gsl_tmp_56) > 0.0))) - (length(gsl_tmp_56) - 4.0) * 0.0) / (pow(gsl_tmp_55, 2.0) + pow(length(gsl_tmp_56) - 4.0, 2.0)), 1.0 * (gsl_tmp_57 * dot(gsl_tmp_58, vec2<f32>(0.0, 1.0)) / (length(gsl_tmp_58) + 1.0 - select(0.0, 1.0, (length(gsl_tmp_58) > 0.0))) - (length(gsl_tmp_58) - 4.0) * 0.0) / (pow(gsl_tmp_57, 2.0) + pow(length(gsl_tmp_58) - 4.0, 2.0)), 1.0 * (gsl_tmp_59 * dot(gsl_tmp_60, vec2<f32>(0.0, 0.0)) / (length(gsl_tmp_60) + 1.0 - select(0.0, 1.0, (length(gsl_tmp_60) > 0.0))) - (length(gsl_tmp_60) - 4.0) * 1.0) / (pow(gsl_tmp_59, 2.0) + pow(length(gsl_tmp_60) - 4.0, 2.0))), 0.8); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
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

struct gsl_struct {
  v_1 : vec3<f32>,
  v_2 : vec3<f32>,
}
