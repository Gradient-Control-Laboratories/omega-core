fn blend(d1: f32, d2: f32, k: f32) -> f32 {
  let h = max(k - abs(d1 - d2), 0.0) / k;
  return min(d1, d2) - h * h * k * 0.25;
}

fn gcl_closure_blendWith(a: f32, b: f32) -> f32 {
  return blend(a, b, 0.25);
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let gsl_tmp_1 = abs(p) - 1.0;
  let gsl_s1 = gsl_shape(gcl_closure_blendWith(gcl_closure_blendWith(gsl_tmp_1.x, gsl_tmp_1.y), gsl_tmp_1.z), 1u);
  return gsl_s1;
}
