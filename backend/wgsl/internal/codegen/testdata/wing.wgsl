fn gsl_remap(p: vec3<f32>) -> f32 {
  let gsl_tmp_1 = abs(p.y);
  let gsl_tmp_2 = gsl_tmp_1 * 0.83909965;
  let gsl_tmp_3 = 8.0 + gsl_tmp_1 * 0.17632698 - gsl_tmp_2;
  let gsl_tmp_9 = max(-p.x, p.x - gsl_tmp_3);
  let gsl_tmp_4 = clamp(p.x / gsl_tmp_3, 0.0001, 0.9999);
  let gsl_tmp_6 = camberY(gsl_tmp_4) * gsl_tmp_3;
  let gsl_tmp_5 = nacaHalfT(gsl_tmp_4) * gsl_tmp_3;
  let gsl_tmp_7 = cos(atan(camberDY(gsl_tmp_4)));
  let gsl_tmp_8 = max(p.z - (gsl_tmp_6 + gsl_tmp_5 * gsl_tmp_7), gsl_tmp_6 - gsl_tmp_5 * gsl_tmp_7 - p.z);
  let gsl_tmp_10 = max(gsl_tmp_9, gsl_tmp_8);
  return select(length(vec2<f32>(max(gsl_tmp_9, 0.0), max(gsl_tmp_8, 0.0))), gsl_tmp_10, gsl_tmp_10 < 0.0);
}

fn nacaHalfT(xn: f32) -> f32 {
  return 0.6 * (0.2969 * sqrt(abs(xn)) - 0.126 * xn - 0.3516 * xn * xn + 0.2843 * xn * xn * xn - 0.1015 * xn * xn * xn * xn);
}

fn camberY(xn: f32) -> f32 {
  let aft = 0.11111111 * (0.2 + 0.8 * xn - xn * xn);
  let fore = 0.25 * (0.8 * xn - xn * xn);
  return select(aft, fore, xn <= 0.4);
}

fn camberDY(xn: f32) -> f32 {
  let aft = 0.22222222 * (0.4 - xn);
  let fore = 0.5 * (0.4 - xn);
  return select(aft, fore, xn <= 0.4);
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let gsl_tmp_1 = abs(p.y);
  let gsl_tmp_2 = gsl_tmp_1 * 0.83909965;
  let gsl_s1 = gsl_shape(max(gsl_remap(vec3<f32>(p.x - gsl_tmp_2, p.y, p.z)), max(-gsl_tmp_1, gsl_tmp_1 - 6.0)), 1u);
  return gsl_s1;
}
