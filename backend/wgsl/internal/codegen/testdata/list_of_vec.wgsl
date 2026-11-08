fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let gsl_tmp_1 = p.x;
  let gsl_tmp_2 = p.y;
  let gsl_tmp_3 = p.x + p.y;
  let gsl_tmp_4 = p.x * 2.0;
  let radials = mat4x3<f32>(vec3<f32>(cos(gsl_tmp_1), sin(gsl_tmp_1), 0.0), vec3<f32>(cos(gsl_tmp_2), sin(gsl_tmp_2), 1.0), vec3<f32>(cos(gsl_tmp_3), sin(gsl_tmp_3), 2.0), vec3<f32>(cos(gsl_tmp_4), sin(gsl_tmp_4), 3.0));
  let first = radials[0];
  let gsl_tmp_5 = p.x;
  let gsl_tmp_6 = p.y;
  let gsl_tmp_7 = p.x + p.y;
  let gsl_tmp_8 = p.x * 2.0;
  let sum = vec3<f32>(0.0, 0.0, 0.0) + vec3<f32>(cos(gsl_tmp_5), sin(gsl_tmp_5), 0.0) + vec3<f32>(cos(gsl_tmp_6), sin(gsl_tmp_6), 1.0) + vec3<f32>(cos(gsl_tmp_7), sin(gsl_tmp_7), 2.0) + vec3<f32>(cos(gsl_tmp_8), sin(gsl_tmp_8), 3.0);
  let gsl_s1 = gsl_shape(length(p + first * 0.1 + sum * 0.05 + vec3<f32>(0.04, 0.05, 0.06)) - 2.0, 1u);
  return gsl_s1;
}
