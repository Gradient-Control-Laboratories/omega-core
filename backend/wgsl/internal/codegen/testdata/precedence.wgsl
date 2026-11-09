fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let a2 = 6.0 - (3.0 - p.x);
  let b2 = 5.3333335 / (3.0 / p.x);
  let c2 = -1.0 + p.x;
  let d2 = -(-3.0 + p.x);
  let gsl_s1 = gsl_shape(a2, 1u);
  let gsl_s2 = gsl_shape(b2, 2u);
  let gsl_s3 = gsl_shape(c2, 3u);
  let gsl_s4 = gsl_shape(d2, 4u);
  let gsl_s5 = gsl_shape(4.4545455, 5u);
  return gsl_min(gsl_min(gsl_min(gsl_min(gsl_s1, gsl_s2), gsl_s3), gsl_s4), gsl_s5);
}
