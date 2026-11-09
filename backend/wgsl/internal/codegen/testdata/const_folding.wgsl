fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let gsl_s1 = gsl_shape(332.2111, 1u);
  let gsl_s2 = gsl_shape(500.0, 2u);
  let gsl_s3 = gsl_shape(505.0, 3u);
  return gsl_min(gsl_min(gsl_s1, gsl_s2), gsl_s3);
}
