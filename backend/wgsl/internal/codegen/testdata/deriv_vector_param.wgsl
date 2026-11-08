fn shell__d(q: vec3<f32>, q__dp0: vec3<f32>, q__dp1: vec3<f32>, q__dp2: vec3<f32>) -> f32 {
  let d = length(q) - 1.0;
  return abs(d) / max(length(vec3<f32>(dot(q, q__dp0) / (length(q) + 1.0 - select(0.0, 1.0, (length(q) > 0.0))), dot(q, q__dp1) / (length(q) + 1.0 - select(0.0, 1.0, (length(q) > 0.0))), dot(q, q__dp2) / (length(q) + 1.0 - select(0.0, 1.0, (length(q) > 0.0))))), 0.5);
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let gsl_s1 = gsl_shape(shell__d(p * 0.5, vec3<f32>(0.5, 0.0, 0.0), vec3<f32>(0.0, 0.5, 0.0), vec3<f32>(0.0, 0.0, 0.5)) - 0.2, 1u);
  return gsl_s1;
}
