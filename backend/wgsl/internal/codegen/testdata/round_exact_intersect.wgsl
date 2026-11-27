fn gsl_builtin_unionRoundExact_remap(p: vec3<f32>, a: f32, b: f32, normalA: vec3<f32>, normalB: vec3<f32>, radius: f32) -> f32 {
  return a;
}

fn gsl_builtin_unionRoundExact_remap_2(p: vec3<f32>, a: f32, b: f32, normalA: vec3<f32>, normalB: vec3<f32>, radius: f32) -> f32 {
  return b;
}

fn builtin_unionRoundExact(p: vec3<f32>, a: f32, b: f32, radius: f32, normalA: vec3<f32>, normalB: vec3<f32>) -> f32 {
  let d1 = a - radius;
  let d2 = b - radius;
  let sumScale = length(normalA + normalB);
  let sum = (d1 + d2) / sumScale;
  let diffScale = length(normalA - normalB);
  let diff = (d1 - d2) / diffScale;
  let sharp = min(d1, d2);
  let field1OnSurface2 = gsl_builtin_unionRoundExact_remap(p - normalB * d2, a, b, normalA, normalB, radius) - radius;
  let field2OnSurface1 = gsl_builtin_unionRoundExact_remap_2(p - normalA * d1, a, b, normalA, normalB, radius) - radius;
  return select(-length(vec2<f32>(sum, diff)), sharp, max(field1OnSurface2, field2OnSurface1) > 0.0) + radius;
}

fn builtin_intersectRoundExact(p: vec3<f32>, a: f32, b: f32, radius: f32, normalA: vec3<f32>, normalB: vec3<f32>) -> f32 {
  return -builtin_unionRoundExact(p, -a, -b, radius, -normalA, -normalB);
}

fn gsl_remap(p: vec3<f32>) -> f32 {
  let n1 = normalize(p);
  let n2 = normalize(p - vec3<f32>(1.2, 0.0, 0.0));
  let blob = builtin_intersectRoundExact(p, builtin_gcl_closure_boxCentered(p, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.6, 1.6, 1.6)), length(p - vec3<f32>(0.8, 0.0, 0.0)) - 1.0, 0.0, n1, n2);
  return blob;
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
  return min(sharp, 0.000000000000000000000000000001 - radius) + sqrt(max(sumOfSquares, 0.000000000000000000000000000001));
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

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let n1 = normalize(p);
  let n2 = normalize(p - vec3<f32>(1.2, 0.0, 0.0));
  let lens = builtin_intersectRoundExact(p, length(p - vec3<f32>(0.0, 0.0, 0.0)) - 1.0, length(p - vec3<f32>(1.2, 0.0, 0.0)) - 1.0, 0.0, n1, n2);
  let gsl_s1 = gsl_shape(builtin_unionSharp(vec2<f32>(lens, gsl_remap(p - vec3<f32>(4.0, 0.0, 0.0)))), 1u);
  return gsl_s1;
}
