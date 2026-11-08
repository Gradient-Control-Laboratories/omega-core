fn builtin_plane(p: vec3<f32>, point: vec3<f32>, normal: vec3<f32>) -> f32 {
  return dot(p - point, normalize(normal));
}

fn builtin_intersectRound(d1: f32, d2: f32, radius: f32) -> f32 {
  let inset1 = max(d1 + radius, 0.0);
  let inset2 = max(d2 + radius, 0.0);
  return min(-radius, max(d1, d2)) + length(vec2<f32>(inset1, inset2));
}

fn builtin_unionSharp(d1: f32, d2: f32) -> f32 {
  return min(d1, d2);
}

fn sdShape(p: vec3<f32>, time: f32, gsl_param_3_9: f32) -> gsl_shape {
  let a = p.x;
  let vector_zero = vec3<f32>(0.0, 0.0, 0.0);
  let angle = gsl_param_3_9;
  let b = builtin_plane(p, vector_zero, vec3<f32>(1.0 * cos(angle) + 0.0 * sin(angle), -1.0 * sin(angle) + 0.0 * cos(angle), 0.0));
  let blend = builtin_intersectRound(a, b, 50.0);
  let separation = builtin_unionSharp(a + 50.0, b + 50.0);
  let gsl_s1 = gsl_shape(blend, 1u);
  let gsl_s2 = gsl_shape(separation, 2u);
  return gsl_min(gsl_s1, gsl_s2);
}
