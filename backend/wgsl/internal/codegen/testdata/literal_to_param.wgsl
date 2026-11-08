fn sdShape(p: vec3<f32>, time: f32, gsl_param_2_16: f32) -> gsl_shape {
  let globalOffset = gsl_param_2_16;
  let gsl_s1 = gsl_shape(length(p + vec3<f32>(8.0, 9.0, 10.0) + globalOffset) - 2.5, 1u);
  return gsl_s1;
}
