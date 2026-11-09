fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let gsl_s1 = gsl_shape(length(p - vec3<f32>(0.000000040192344, -1.5, 0.0)) - 0.4, 1u);
  return gsl_s1;
}
