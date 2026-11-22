fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let gsl_tmp_1 = p.x;
  let gsl_tmp_2 = p.y;
  let gsl_tmp_3 = p.x + p.y;
  let gsl_tmp_4 = p.x * 2.0;
  let gsl_cse_1 = vec3<f32>(cos(gsl_tmp_1), sin(gsl_tmp_1), 0.0);
  let gsl_cse_2 = vec3<f32>(cos(gsl_tmp_2), sin(gsl_tmp_2), 1.0);
  let gsl_cse_3 = vec3<f32>(cos(gsl_tmp_3), sin(gsl_tmp_3), 2.0);
  let gsl_cse_4 = vec3<f32>(cos(gsl_tmp_4), sin(gsl_tmp_4), 3.0);
  let radials = mat4x3<f32>(gsl_cse_1, gsl_cse_2, gsl_cse_3, gsl_cse_4);
  let first = radials[0];
  let sum = vec3<f32>(0.0, 0.0, 0.0) + gsl_cse_1 + gsl_cse_2 + gsl_cse_3 + gsl_cse_4;
  let gsl_s1 = gsl_shape(length(p + first * 0.1 + sum * 0.05 + vec3<f32>(0.04, 0.05, 0.06)) - 2.0, 1u);
  return gsl_s1;
}
