fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let gsl_s1 = gsl_shape(21.0 + length(p) - 2.0, 1u);
  return gsl_s1;
}
