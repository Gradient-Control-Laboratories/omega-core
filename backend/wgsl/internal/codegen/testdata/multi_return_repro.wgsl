fn gsl_floored(a: f32, n: f32) -> f32 {
  return a - n * floor(a / n);
}

fn intersectRound2(d1: f32, d2: f32, k: f32) -> f32 {
  let c1 = max(d1 + k, 0.0);
  let c2 = max(d2 + k, 0.0);
  return min(-k, max(d1, d2)) + length(vec2<f32>(c1, c2));
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let gsl_tmp_1 = length(p.xy);
  let gsl_tmp_2 = gsl_floored((atan2(p.x, p.y) + 0.5235988), 1.0471976) - 0.5235988;
  let gsl_tmp_3 = gsl_tmp_1 * sin(gsl_tmp_2);
  let gsl_tmp_4 = length(p.xy);
  let gsl_tmp_5 = gsl_floored((atan2(p.x, p.y) + 0.5235988), 1.0471976) - 0.5235988;
  let gsl_tmp_6 = gsl_tmp_4 * sin(gsl_tmp_5);
  let gsl_s1 = gsl_shape(intersectRound2(abs(length(vec2<f32>(gsl_tmp_3, gsl_tmp_1 * cos(gsl_tmp_2))) - 3.0) - 0.5, abs(gsl_tmp_3) - 0.25, 0.0), 1u);
  let gsl_s2 = gsl_shape(intersectRound2(abs(length(vec2<f32>(gsl_tmp_6, gsl_tmp_4 * cos(gsl_tmp_5))) - 2.0) - 0.4, abs(gsl_tmp_6) - 0.2, 0.0), 2u);
  return gsl_min(gsl_s1, gsl_s2);
}
