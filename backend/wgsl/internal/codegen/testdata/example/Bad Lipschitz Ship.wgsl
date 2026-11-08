fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let gsl_tmp_1 = sqrt(max(0.0, 1.0 - p.x * p.x / 16.0));
  let s = max(length(vec2<f32>(p.y / (1.0 * gsl_tmp_1 + 0.0001), p.z / (0.75 * gsl_tmp_1 + 0.0001))) - 1.0, abs(p.x) - 4.0);
  let gsl_s1 = gsl_shape(s, 1u);
  return gsl_s1;
}
