fn gsl_remap(p: vec3<f32>) -> f32 {
  let n1 = normalize(p);
  let n2 = normalize(p - vec3<f32>(1.2, 0.0, 0.0));
  let blob = gcl_closure_intersectRoundExact_2(p, n1, n2);
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

fn gsl_gcl_closure_unionRoundExact_3_remap(p: vec3<f32>, normal1: vec3<f32>, normal2: vec3<f32>) -> f32 {
  return -(length(p - vec3<f32>(0.0, 0.0, 0.0)) - 1.0);
}

fn gsl_gcl_closure_unionRoundExact_3_remap_2(p: vec3<f32>, normal1: vec3<f32>, normal2: vec3<f32>) -> f32 {
  return -(length(p - vec3<f32>(1.2, 0.0, 0.0)) - 1.0);
}

fn gcl_closure_unionRoundExact_3(p: vec3<f32>, normal1: vec3<f32>, normal2: vec3<f32>) -> f32 {
  let d1 = -(length(p - vec3<f32>(0.0, 0.0, 0.0)) - 1.0);
  let d2 = -(length(p - vec3<f32>(1.2, 0.0, 0.0)) - 1.0);
  let sumScale = length(normal1 + normal2);
  let sum = (d1 + d2) / sumScale;
  let diffScale = length(normal1 - normal2);
  let diff = (d1 - d2) / diffScale;
  let sharp = min(d1, d2);
  let field1OnSurface2 = gsl_gcl_closure_unionRoundExact_3_remap(p - normal2 * d2, normal1, normal2);
  let field2OnSurface1 = gsl_gcl_closure_unionRoundExact_3_remap_2(p - normal1 * d1, normal1, normal2);
  return select(-length(vec2<f32>(sum, diff)), sharp, max(field1OnSurface2, field2OnSurface1) > 0.0);
}

fn gcl_closure_intersectRoundExact(p: vec3<f32>, normal1: vec3<f32>, normal2: vec3<f32>) -> f32 {
  return -gcl_closure_unionRoundExact_3(p, -normal1, -normal2);
}

fn gsl_gcl_closure_unionRoundExact_6_remap(p: vec3<f32>, normal1: vec3<f32>, normal2: vec3<f32>) -> f32 {
  return -builtin_gcl_closure_boxCentered(p, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.6, 1.6, 1.6));
}

fn gsl_gcl_closure_unionRoundExact_6_remap_2(p: vec3<f32>, normal1: vec3<f32>, normal2: vec3<f32>) -> f32 {
  return -(length(p - vec3<f32>(0.8, 0.0, 0.0)) - 1.0);
}

fn gcl_closure_unionRoundExact_6(p: vec3<f32>, normal1: vec3<f32>, normal2: vec3<f32>) -> f32 {
  let d1 = -builtin_gcl_closure_boxCentered(p, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.6, 1.6, 1.6));
  let d2 = -(length(p - vec3<f32>(0.8, 0.0, 0.0)) - 1.0);
  let sumScale = length(normal1 + normal2);
  let sum = (d1 + d2) / sumScale;
  let diffScale = length(normal1 - normal2);
  let diff = (d1 - d2) / diffScale;
  let sharp = min(d1, d2);
  let field1OnSurface2 = gsl_gcl_closure_unionRoundExact_6_remap(p - normal2 * d2, normal1, normal2);
  let field2OnSurface1 = gsl_gcl_closure_unionRoundExact_6_remap_2(p - normal1 * d1, normal1, normal2);
  return select(-length(vec2<f32>(sum, diff)), sharp, max(field1OnSurface2, field2OnSurface1) > 0.0);
}

fn gcl_closure_intersectRoundExact_2(p: vec3<f32>, normal1: vec3<f32>, normal2: vec3<f32>) -> f32 {
  return -gcl_closure_unionRoundExact_6(p, -normal1, -normal2);
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let n1 = normalize(p);
  let n2 = normalize(p - vec3<f32>(1.2, 0.0, 0.0));
  let lens = gcl_closure_intersectRoundExact(p, n1, n2);
  let gsl_s1 = gsl_shape(builtin_unionSharp(vec2<f32>(lens, gsl_remap(p - vec3<f32>(4.0, 0.0, 0.0)))), 1u);
  return gsl_s1;
}
