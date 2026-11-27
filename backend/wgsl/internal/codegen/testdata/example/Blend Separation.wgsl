fn builtin_plane(p: vec3<f32>, point: vec3<f32>, normal: vec3<f32>) -> f32 {
  return dot(p - point, normalize(normal));
}

fn builtin_intersectRound(a: f32, b: f32, radius: f32) -> f32 {
  let inset1 = max(a + radius, 0.0);
  let inset2 = max(b + radius, 0.0);
  return min(-radius, max(a, b)) + length(vec2<f32>(inset1, inset2));
}

fn builtin_unionSharp(a: f32, b: f32) -> f32 {
  return min(a, b);
}

fn sdShape(p: vec3<f32>, time: f32, gsl_param_3_28: f32) -> gsl_shape {
  let a = p.x;
  let vector_zero = vec3<f32>(0.0, 0.0, 0.0);
  let angle = gsl_param_3_28;
  let b = builtin_plane(p, vector_zero, vec3<f32>(1.0 * cos(angle) + 0.0 * sin(angle), -1.0 * sin(angle) + 0.0 * cos(angle), 0.0));
  let blend = builtin_intersectRound(a, b, 50.0);
  let separation = builtin_unionSharp(a + 50.0, b + 50.0);
  let gsl_s1 = gsl_shape(blend, 1u);
  let gsl_s2 = gsl_shape(separation, 2u);
  return gsl_min(gsl_s1, gsl_s2);
}
