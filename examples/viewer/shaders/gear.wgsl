fn sdShape(p: vec3<f32>, time: f32, gsl_param_2_9: f32, gsl_param_6_10: f32, gsl_param_10_8: f32, gsl_param_14_10: f32) -> gsl_shape {
  let r = sqrt(p.x * p.x + p.y * p.y);
  let radius = gsl_param_6_10;
  let angle = atan2(p.y, p.x);
  let teeth = gsl_param_2_9;
  let tooth = 0.25 * sin(angle * teeth);
  let d_outer = r - (radius + tooth);
  let bore = gsl_param_10_8;
  let d_inner = bore - r;
  let d_2d = max(d_outer, d_inner);
  let height = gsl_param_14_10;
  let d_z = abs(p.z) - height * 0.5;
  let gsl_s1 = gsl_shape(max(d_2d, d_z), 1u);
  return gsl_s1;
}