fn gsl_floored(a: f32, n: f32) -> f32 {
  return a - n * floor(a / n);
}

fn builtin_sphere(p: vec3<f32>, radius: f32) -> f32 {
  return length(p) - radius;
}

fn builtin_intersectSharp(fields: vec3<f32>) -> f32 {
  return max(max(fields.x, fields.y), fields.z);
}

fn builtin_firstOf(list: f32) -> f32 {
  return list;
}

fn builtin_intersectChamferEdges(fields: vec3<f32>, setback: f32) -> f32 {
  let face = builtin_intersectSharp(fields);
  return builtin_intersectSharp(vec3<f32>(builtin_intersectSharp(vec3<f32>(select(face, (fields.x + fields.x + setback) / 1.4142135, 0.0 > 0.0), select(face, (fields.x + fields.y + setback) / 1.4142135, 1.0 > 0.0), select(face, (fields.x + fields.z + setback) / 1.4142135, 1.0 > 0.0))), builtin_intersectSharp(vec3<f32>(select(face, (fields.y + fields.x + setback) / 1.4142135, 0.0 > 0.0), select(face, (fields.y + fields.y + setback) / 1.4142135, 0.0 > 0.0), select(face, (fields.y + fields.z + setback) / 1.4142135, 1.0 > 0.0))), builtin_intersectSharp(vec3<f32>(select(face, (fields.z + fields.x + setback) / 1.4142135, 0.0 > 0.0), select(face, (fields.z + fields.y + setback) / 1.4142135, 0.0 > 0.0), select(face, (fields.z + fields.z + setback) / 1.4142135, 0.0 > 0.0)))));
}

fn builtin_lastOf(list: f32) -> f32 {
  return list;
}

fn builtin_intersectChamferVertices(fields: vec3<f32>, setback: f32) -> f32 {
  let face = builtin_intersectSharp(fields);
  return builtin_intersectSharp(vec3<f32>(builtin_intersectSharp(vec3<f32>(builtin_intersectSharp(vec3<f32>(select(face, (fields.x + fields.x + fields.x + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.x + fields.x + fields.y + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.x + fields.x + fields.z + setback) / 1.7320508, 0.0 > 0.0))), builtin_intersectSharp(vec3<f32>(select(face, (fields.x + fields.y + fields.x + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.x + fields.y + fields.y + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.x + fields.y + fields.z + setback) / 1.7320508, 1.0 > 0.0))), builtin_intersectSharp(vec3<f32>(select(face, (fields.x + fields.z + fields.x + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.x + fields.z + fields.y + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.x + fields.z + fields.z + setback) / 1.7320508, 0.0 > 0.0))))), builtin_intersectSharp(vec3<f32>(builtin_intersectSharp(vec3<f32>(select(face, (fields.y + fields.x + fields.x + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.y + fields.x + fields.y + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.y + fields.x + fields.z + setback) / 1.7320508, 0.0 > 0.0))), builtin_intersectSharp(vec3<f32>(select(face, (fields.y + fields.y + fields.x + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.y + fields.y + fields.y + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.y + fields.y + fields.z + setback) / 1.7320508, 0.0 > 0.0))), builtin_intersectSharp(vec3<f32>(select(face, (fields.y + fields.z + fields.x + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.y + fields.z + fields.y + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.y + fields.z + fields.z + setback) / 1.7320508, 0.0 > 0.0))))), builtin_intersectSharp(vec3<f32>(builtin_intersectSharp(vec3<f32>(select(face, (fields.z + fields.x + fields.x + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.z + fields.x + fields.y + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.z + fields.x + fields.z + setback) / 1.7320508, 0.0 > 0.0))), builtin_intersectSharp(vec3<f32>(select(face, (fields.z + fields.y + fields.x + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.z + fields.y + fields.y + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.z + fields.y + fields.z + setback) / 1.7320508, 0.0 > 0.0))), builtin_intersectSharp(vec3<f32>(select(face, (fields.z + fields.z + fields.x + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.z + fields.z + fields.y + setback) / 1.7320508, 0.0 > 0.0), select(face, (fields.z + fields.z + fields.z + setback) / 1.7320508, 0.0 > 0.0)))))));
}

fn builtin_intersectChamfer(fields: vec3<f32>, setback: f32) -> f32 {
  return max(max(builtin_intersectSharp(fields), builtin_intersectChamferEdges(fields, builtin_firstOf(setback))), builtin_intersectChamferVertices(fields, builtin_lastOf(setback)));
}

fn builtin_gcl_closure_intersectChamfer(fields: vec3<f32>) -> f32 {
  return builtin_intersectChamfer(fields, 0.2);
}

fn builtin_gcl_closure_boxCentered(p: vec3<f32>, center: vec3<f32>, size: vec3<f32>) -> f32 {
  let halfSize = size * 0.5;
  let d = abs(p - center) - halfSize;
  let solid = builtin_gcl_closure_intersectChamfer(vec3<f32>(d.x, d.y, d.z));
  return solid;
}

fn builtin_gcl_closure_box(p: vec3<f32>, pointA: vec3<f32>, pointB: vec3<f32>) -> f32 {
  let center = (pointA + pointB) * 0.5;
  let size = abs(pointB - pointA);
  return builtin_gcl_closure_boxCentered(p, center, size);
}

fn builtin_intersectSharp_2(fields: vec2<f32>) -> f32 {
  return max(fields.x, fields.y);
}

fn builtin_intersectExponential_2(fields: vec2<f32>, width: f32) -> f32 {
  let peak = builtin_intersectSharp_2(fields);
  let scale = width / 2.7182817;
  let total = 0.0 + exp((fields.x - peak) / scale) + exp((fields.y - peak) / scale);
  return peak + scale * log(total);
}

fn builtin_gcl_closure_intersectExponential_2(fields: vec2<f32>) -> f32 {
  return builtin_intersectExponential_2(fields, 0.2);
}

fn builtin_gcl_closure_cylinderCapped(p: vec3<f32>, pointA: vec3<f32>, pointB: vec3<f32>, radius: f32) -> f32 {
  let pa = p - pointA;
  let ba = pointB - pointA;
  let paba = dot(pa, ba);
  let baba = dot(ba, ba);
  let t = clamp(paba / baba, 0.0, 1.0);
  let radialDistance = length(pa - ba * t) - radius;
  let axialDistance = (abs(paba - baba * 0.5) - baba * 0.5) / sqrt(baba);
  let form = builtin_gcl_closure_intersectExponential_2(vec2<f32>(radialDistance, axialDistance));
  return form;
}

fn builtin_intersectRound(fields: vec3<f32>, radius: f32) -> f32 {
  let sumOfSquares = 0.0 + max(fields.x + radius, 0.0) * max(fields.x + radius, 0.0) + max(fields.y + radius, 0.0) * max(fields.y + radius, 0.0) + max(fields.z + radius, 0.0) * max(fields.z + radius, 0.0);
  return min(-radius, builtin_intersectSharp(fields)) + sqrt(sumOfSquares);
}

fn builtin_gcl_closure_intersectRound(fields: vec3<f32>) -> f32 {
  return builtin_intersectRound(fields, 0.2);
}

fn builtin_gcl_closure_boxCentered_2(p: vec3<f32>, center: vec3<f32>, size: vec3<f32>) -> f32 {
  let halfSize = size * 0.5;
  let d = abs(p - center) - halfSize;
  let solid = builtin_gcl_closure_intersectRound(vec3<f32>(d.x, d.y, d.z));
  return solid;
}

fn builtin_intersectExponential(fields: vec3<f32>, width: f32) -> f32 {
  let peak = builtin_intersectSharp(fields);
  let scale = width / 2.7182817;
  let total = 0.0 + exp((fields.x - peak) / scale) + exp((fields.y - peak) / scale) + exp((fields.z - peak) / scale);
  return peak + scale * log(total);
}

fn builtin_gcl_closure_intersectExponential(fields: vec3<f32>) -> f32 {
  return builtin_intersectExponential(fields, 0.5);
}

fn builtin_gcl_closure_boxCentered_3(p: vec3<f32>, center: vec3<f32>, size: vec3<f32>) -> f32 {
  let halfSize = size * 0.5;
  let d = abs(p - center) - halfSize;
  let solid = builtin_gcl_closure_intersectExponential(vec3<f32>(d.x, d.y, d.z));
  return solid;
}

fn builtin_gcl_closure_intersectRound_2(fields: vec3<f32>) -> f32 {
  return builtin_intersectRound(fields, 0.0);
}

fn builtin_gcl_closure_boxCentered_4(p: vec3<f32>, center: vec3<f32>, size: vec3<f32>) -> f32 {
  let halfSize = size * 0.5;
  let d = abs(p - center) - halfSize;
  let solid = builtin_gcl_closure_intersectRound_2(vec3<f32>(d.x, d.y, d.z));
  return solid;
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  gsl_partNear = vec3<f32>(1e10, 1e10, 1e10);
  let lens = builtin_sphere(p - vec3<f32>(-13.5, 0.0, 0.0), 3.5);
  let brassBall = builtin_sphere(p - vec3<f32>(-13.5, 8.0, 0.0), 2.0);
  let slab = builtin_gcl_closure_box(p - vec3<f32>(-4.5, -4.0, 0.0), vec3<f32>(-4.0, -0.75, -4.0), vec3<f32>(4.0, 0.75, 4.0));
  let redBall = builtin_sphere(p - vec3<f32>(-4.5, 4.0, 0.0), 2.5);
  let cube = builtin_gcl_closure_boxCentered_2(p - vec3<f32>(4.5, 0.0, 0.0), vec3<f32>(0.0, 0.0, 0.0) * 0.5, abs(vec3<f32>(6.0, 6.0, 6.0)));
  let block = builtin_gcl_closure_boxCentered_3(p - vec3<f32>(13.5, 0.0, 0.0), vec3<f32>(0.0, 0.0, 0.0) * 0.5, abs(vec3<f32>(7.0, 7.0, 7.0)));
  let pin = builtin_gcl_closure_cylinderCapped(p - vec3<f32>(13.5, 0.0, 0.0), vec3<f32>(0.0, 0.0, -2.5), vec3<f32>(0.0, 0.0, 2.5), 0.6);
  let gsl_partLo = vec3<f32>(-19.0, -8.0, -5.0);
  let gsl_partHi = vec3<f32>(19.0, 12.0, -4.0);
  let gsl_partBox = gsl_boxDistance(p, gsl_partLo, gsl_partHi);
  var gsl_part = gsl_partBox;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox <= gsl_vanishView.w) {
    let gsl_partVanish = gsl_vanish(gsl_partLo, gsl_partHi);
    gsl_part = 1e10;
    if (gsl_partVanish.x > 0.0) { gsl_part = builtin_gcl_closure_boxCentered_4(p, vec3<f32>(0.0, 2.0, -4.5), vec3<f32>(38.0, 20.0, 1.0)) + gsl_partVanish.y; }
    if (abs(gsl_part) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part), gsl_partVanish.zw); }
  }
  let board = gsl_part;
  let gsl_s1 = gsl_shape(lens, 1u);
  let gsl_s2 = gsl_shape(brassBall, 2u);
  let gsl_s3 = gsl_shape(slab, 3u);
  let gsl_s4 = gsl_shape(redBall, 4u);
  let gsl_s5 = gsl_shape(cube, 5u);
  let gsl_s6 = gsl_shape(block, 6u);
  let gsl_s7 = gsl_shape(pin, 7u);
  let gsl_s8 = gsl_shape(board, 8u);
  return gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_s1, gsl_s2), gsl_s3), gsl_s4), gsl_s5), gsl_s6), gsl_s7), gsl_s8);
}

fn gsl_colorField(p: vec3<f32>, time: f32, gsl_colorIdx: u32) -> vec4<f32> {
  if (gsl_colorIdx == 1u) { return vec4<f32>(vec3<f32>(1.0, 1.0, 1.0), 1.0); }
  if (gsl_colorIdx == 2u) { return vec4<f32>(vec3<f32>(0.95, 0.9, 0.69), 1.0); }
  if (gsl_colorIdx == 3u) { return vec4<f32>(vec3<f32>(1.0, 1.0, 1.0), 1.0); }
  if (gsl_colorIdx == 4u) {
    let color_cherry = vec3<f32>(0.871, 0.192, 0.388);
    return vec4<f32>(color_cherry, 1.0);
  }
  if (gsl_colorIdx == 5u) { return vec4<f32>(vec3<f32>(0.35, 0.35, 0.37), 1.0); }
  if (gsl_colorIdx == 6u) { return vec4<f32>(vec3<f32>(0.95, 0.85, 0.6), 1.0); }
  if (gsl_colorIdx == 7u) { return vec4<f32>(vec3<f32>(0.8, 0.81, 0.82), 1.0); }
  if (gsl_colorIdx == 8u) { return vec4<f32>(select(vec3<f32>(0.12, 0.12, 0.12), vec3<f32>(0.85, 0.85, 0.85), 1.0 - 2.0 * gsl_floored((gsl_floored((floor(p.x / 2.0) + floor(p.y / 2.0)), 2.0) + 2.0), 2.0) > 0.0), 1.0); }
  return vec4<f32>(0.0, 0.0, 0.0, 0.0);
}
fn gsl_materialField(p: vec3<f32>, time: f32, gsl_materialIdx: u32) -> gsl_material {
  if (gsl_materialIdx == 1u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.02, 0.0, 1.52, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 2u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.3, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 3u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.5, 0.0, 1.52, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 4u) {
    let material_shinyPlastic = mat4x4<f32>(vec4<f32>(0.1, 0.0, 1.46, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    let gsl_m = material_shinyPlastic; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 5u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.05, 0.0, 1.52, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 6.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 6u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.15, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 8.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 7u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.35, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 8u) {
    let material_matte = mat4x4<f32>(vec4<f32>(1.0, 0.0, 1.45, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    let gsl_m = material_matte; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
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
