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

fn builtin_gcl_closure_boxCentered_3(p: vec3<f32>, center: vec3<f32>, size: vec3<f32>) -> f32 {
  let halfSize = size * 0.5;
  let d = abs(p - center) - halfSize;
  let solid = builtin_gcl_closure_intersectChamfer(vec3<f32>(d.x, d.y, d.z));
  return solid;
}

fn builtin_intersectSharp_2(fields: vec2<f32>) -> f32 {
  return max(fields.x, fields.y);
}

fn builtin_intersectRound_2(fields: vec2<f32>, radius: f32) -> f32 {
  let sumOfSquares = 0.0 + max(fields.x + radius, 0.0) * max(fields.x + radius, 0.0) + max(fields.y + radius, 0.0) * max(fields.y + radius, 0.0);
  return min(-radius, builtin_intersectSharp_2(fields)) + sqrt(sumOfSquares);
}

fn builtin_gcl_closure_intersectRound_3(fields: vec2<f32>) -> f32 {
  return builtin_intersectRound_2(fields, 0.3);
}

fn builtin_gcl_closure_rectangleXYCentered(p: vec3<f32>, center: vec2<f32>, size: vec2<f32>) -> f32 {
  let halfSize = size * 0.5;
  let d = abs(p.xy - center) - halfSize;
  return builtin_gcl_closure_intersectRound_3(vec2<f32>(d.x, d.y));
}

fn builtin_gcl_closure_intersectRound_7(fields: vec2<f32>) -> f32 {
  return builtin_intersectRound_2(fields, 0.15);
}

fn builtin_gcl_closure_cylinderCapped(p: vec3<f32>, pointA: vec3<f32>, pointB: vec3<f32>, radius: f32) -> f32 {
  let pa = p - pointA;
  let ba = pointB - pointA;
  let paba = dot(pa, ba);
  let baba = dot(ba, ba);
  let t = clamp(paba / baba, 0.0, 1.0);
  let radialDistance = length(pa - ba * t) - radius;
  let axialDistance = (abs(paba - baba * 0.5) - baba * 0.5) / sqrt(baba);
  let form = builtin_gcl_closure_intersectRound_7(vec2<f32>(radialDistance, axialDistance));
  return form;
}

fn builtin_unionSharp(fields: array<f32, 8>) -> f32 {
  return min(min(min(min(min(min(min(fields[0], fields[1]), fields[2]), fields[3]), fields[4]), fields[5]), fields[6]), fields[7]);
}

fn builtin_intersectSharp(fields: vec3<f32>) -> f32 {
  return max(max(fields.x, fields.y), fields.z);
}

fn builtin_intersectRound(fields: vec3<f32>, radius: f32) -> f32 {
  let sumOfSquares = 0.0 + max(fields.x + radius, 0.0) * max(fields.x + radius, 0.0) + max(fields.y + radius, 0.0) * max(fields.y + radius, 0.0) + max(fields.z + radius, 0.0) * max(fields.z + radius, 0.0);
  return min(-radius, builtin_intersectSharp(fields)) + sqrt(sumOfSquares);
}

fn builtin_gcl_closure_intersectRound(fields: vec3<f32>) -> f32 {
  return builtin_intersectRound(fields, 0.2);
}

fn builtin_gcl_closure_boxCentered(p: vec3<f32>, center: vec3<f32>, size: vec3<f32>) -> f32 {
  let halfSize = size * 0.5;
  let d = abs(p - center) - halfSize;
  let solid = builtin_gcl_closure_intersectRound(vec3<f32>(d.x, d.y, d.z));
  return solid;
}

fn builtin_gcl_closure_intersectRound_2(fields: vec3<f32>) -> f32 {
  return builtin_intersectRound(fields, 0.0);
}

fn builtin_gcl_closure_boxCentered_2(p: vec3<f32>, center: vec3<f32>, size: vec3<f32>) -> f32 {
  let halfSize = size * 0.5;
  let d = abs(p - center) - halfSize;
  let solid = builtin_gcl_closure_intersectRound_2(vec3<f32>(d.x, d.y, d.z));
  return solid;
}

fn builtin_gcl_closure_intersectRound_4(fields: vec2<f32>) -> f32 {
  return builtin_intersectRound_2(fields, 0.0);
}

fn builtin_gcl_closure_rectangleXYCentered_2(p: vec3<f32>, center: vec2<f32>, size: vec2<f32>) -> f32 {
  let halfSize = size * 0.5;
  let d = abs(p.xy - center) - halfSize;
  return builtin_gcl_closure_intersectRound_4(vec2<f32>(d.x, d.y));
}

fn builtin_gcl_closure_intersectRound_5(fields: vec2<f32>) -> f32 {
  return builtin_intersectRound_2(fields, 0.3);
}

fn builtin_gcl_closure_rectangleXYCentered_3(p: vec3<f32>, center: vec2<f32>, size: vec2<f32>) -> f32 {
  let halfSize = size * 0.5;
  let d = abs(p.xy - center) - halfSize;
  return builtin_gcl_closure_intersectRound_5(vec2<f32>(d.x, d.y));
}

fn builtin_gcl_closure_intersectRound_6(fields: vec2<f32>) -> f32 {
  return builtin_intersectRound_2(fields, 0.0);
}

fn builtin_gcl_closure_rectangleXYCentered_4(p: vec3<f32>, center: vec2<f32>, size: vec2<f32>) -> f32 {
  let halfSize = size * 0.5;
  let d = abs(p.xy - center) - halfSize;
  return builtin_gcl_closure_intersectRound_6(vec2<f32>(d.x, d.y));
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let b1 = builtin_gcl_closure_boxCentered(p, vec3<f32>(2.0, 2.0, 2.0) * 0.5, abs(vec3<f32>(2.0, 2.0, 2.0)));
  let b2 = builtin_gcl_closure_boxCentered_2(p, vec3<f32>(8.0, 2.0, 2.0) * 0.5, abs(vec3<f32>(2.0, 2.0, 2.0)));
  let b3 = builtin_gcl_closure_boxCentered_3(p, vec3<f32>(8.0, 0.0, 0.0), vec3<f32>(2.0, 2.0, 2.0));
  let r1 = builtin_gcl_closure_rectangleXYCentered(p, vec2<f32>(0.0, 5.0), vec2<f32>(2.0, 3.0));
  let r2 = builtin_gcl_closure_rectangleXYCentered_2(p, vec2<f32>(5.0, 5.0), vec2<f32>(2.0, 3.0));
  let r3 = builtin_gcl_closure_rectangleXYCentered_3(p, vec2<f32>(18.0, 11.0) * 0.5, abs(vec2<f32>(2.0, 3.0)));
  let r4 = builtin_gcl_closure_rectangleXYCentered_4(p, vec2<f32>(26.0, 11.0) * 0.5, abs(vec2<f32>(2.0, 3.0)));
  let c1 = builtin_gcl_closure_cylinderCapped(p, vec3<f32>(0.0, 0.0, -6.0), vec3<f32>(0.0, 0.0, -3.0), 1.0);
  let gsl_s1 = gsl_shape(builtin_unionSharp(array<f32, 8>(b1, b2, b3, r1, r2, r3, r4, c1)), 1u);
  return gsl_s1;
}
