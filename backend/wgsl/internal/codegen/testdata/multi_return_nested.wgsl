fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let gsl_tmp_1 = p.x;
  let gsl_tmp_2 = p.y;
  let s = gsl_tmp_1 + gsl_tmp_2;
  let d = gsl_tmp_1 - gsl_tmp_2;
  let wa = p.x + p.x;
  let gsl_s1 = gsl_shape(s + d + wa + gsl_tmp_2 + s + gsl_tmp_2, 1u);
  return gsl_s1;
}
