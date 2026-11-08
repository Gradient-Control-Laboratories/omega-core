fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let gsl_s1 = gsl_shape(min(min(min(min(1000.0, length(p - vec3<f32>(1.5, 0.0, 0.0)) - 0.4), length(p - vec3<f32>(0.00000000000000009184851, -1.5, 0.0)) - 0.4), length(p - vec3<f32>(-1.5, -0.00000000000000018369701, 0.0)) - 0.4), length(p - vec3<f32>(-0.00000000000000027554554, 1.5, 0.0)) - 0.4), 1u);
  return gsl_s1;
}
