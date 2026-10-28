fn gsl_remap(p: vec3<f32>) -> f32 {
  let gsl_tmp_1 = p.x * 18.849556;
  let gsl_tmp_2 = p.y * 18.849556;
  let gsl_tmp_3 = p.z * 18.849556;
  return abs(sin(gsl_tmp_1) * cos(gsl_tmp_2) + sin(gsl_tmp_2) * cos(gsl_tmp_3) + sin(gsl_tmp_3) * cos(gsl_tmp_1)) / 6.2831855 - 0.03;
}

fn gsl_remap_2(p: vec3<f32>) -> f32 {
  return max(max(abs(p.x) - 5.0, abs(p.y) - 5.0), abs(p.z) - 5.0);
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let gyroid = max(gsl_remap(p), gsl_remap_2(p));
  let gsl_s1 = gsl_shape(gyroid, 1u);
  return gsl_s1;
}
