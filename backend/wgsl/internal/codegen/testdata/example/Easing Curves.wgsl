fn ease_inQuadratic(t: f32) -> f32 {
  return t * t;
}

fn ease_outQuadratic(t: f32) -> f32 {
  return 1.0 - ease_inQuadratic(1.0 - t);
}

fn ease_quadratic(t: f32) -> f32 {
  return t * t * (3.0 - 2.0 * t);
}

fn ease_inCubic(t: f32) -> f32 {
  return t * t * t;
}

fn ease_outCubic(t: f32) -> f32 {
  return 1.0 - ease_inCubic(1.0 - t);
}

fn ease_cubic(t: f32) -> f32 {
  return t * t * t * (t * (t * 6.0 - 15.0) + 10.0);
}

fn ease_inExponential(t: f32) -> f32 {
  return select(exp2(10.0 * t - 10.0), 0.0, t <= 0.0);
}

fn ease_outExponential(t: f32) -> f32 {
  return 1.0 - ease_inExponential(1.0 - t);
}

fn ease_exponential(t: f32) -> f32 {
  let hi = select(1.0 - exp2(-20.0 * t + 10.0) * 0.5, 1.0, t >= 1.0);
  let lo = select(exp2(20.0 * t - 10.0) * 0.5, 0.0, t <= 0.0);
  return select(hi, lo, t < 0.5);
}

fn ease_inCircular(t: f32) -> f32 {
  return 1.0 - sqrt(max(0.0, 1.0 - t * t));
}

fn ease_outCircular(t: f32) -> f32 {
  return 1.0 - ease_inCircular(1.0 - t);
}

fn ease_circular(t: f32) -> f32 {
  let b = -2.0 * t + 2.0;
  let hi = (sqrt(max(0.0, 1.0 - b * b)) + 1.0) * 0.5;
  let a = 2.0 * t;
  let lo = (1.0 - sqrt(max(0.0, 1.0 - a * a))) * 0.5;
  return select(hi, lo, t < 0.5);
}

fn ease_linear(t: f32) -> f32 {
  return t;
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let gsl_tmp_2 = (p.y - -0.0) / 10.0;
  let gsl_tmp_1 = (p.x - 0.0) / 10.0;
  let gsl_cse_4 = abs(gsl_tmp_1 - 0.5) - 0.5;
  let gsl_cse_5 = abs(gsl_tmp_2 - 0.5) - 0.5;
  let gsl_tmp_3 = max(gsl_cse_4, gsl_cse_5) * 10.0;
  let gsl_tmp_4 = (p.x - 13.0) / 10.0;
  let gsl_cse_6 = abs(gsl_tmp_4 - 0.5) - 0.5;
  let gsl_tmp_6 = max(gsl_cse_6, gsl_cse_5) * 10.0;
  let gsl_tmp_7 = (p.x - 26.0) / 10.0;
  let gsl_cse_7 = abs(gsl_tmp_7 - 0.5) - 0.5;
  let gsl_tmp_9 = max(gsl_cse_7, gsl_cse_5) * 10.0;
  let gsl_tmp_11 = (p.y - -13.0) / 10.0;
  let gsl_cse_1 = abs(gsl_tmp_11 - 0.5) - 0.5;
  let gsl_tmp_12 = max(gsl_cse_4, gsl_cse_1) * 10.0;
  let gsl_tmp_15 = max(gsl_cse_6, gsl_cse_1) * 10.0;
  let gsl_tmp_18 = max(gsl_cse_7, gsl_cse_1) * 10.0;
  let gsl_tmp_20 = (p.y - -26.0) / 10.0;
  let gsl_cse_2 = abs(gsl_tmp_20 - 0.5) - 0.5;
  let gsl_tmp_21 = max(gsl_cse_4, gsl_cse_2) * 10.0;
  let gsl_tmp_24 = max(gsl_cse_6, gsl_cse_2) * 10.0;
  let gsl_tmp_27 = max(gsl_cse_7, gsl_cse_2) * 10.0;
  let gsl_tmp_29 = (p.y - -39.0) / 10.0;
  let gsl_cse_3 = abs(gsl_tmp_29 - 0.5) - 0.5;
  let gsl_tmp_30 = max(gsl_cse_4, gsl_cse_3) * 10.0;
  let gsl_tmp_33 = max(gsl_cse_6, gsl_cse_3) * 10.0;
  let gsl_tmp_36 = max(gsl_cse_7, gsl_cse_3) * 10.0;
  let gsl_tmp_38 = (p.y - -52.0) / 10.0;
  let gsl_tmp_39 = max(gsl_cse_6, abs(gsl_tmp_38 - 0.5) - 0.5) * 10.0;
  let gsl_cse_8 = clamp(gsl_tmp_1, 0.0, 1.0);
  let gsl_cse_9 = clamp(gsl_tmp_4, 0.0, 1.0);
  let gsl_cse_10 = clamp(gsl_tmp_7, 0.0, 1.0);
  let grid = min(min(min(min(min(min(min(min(min(min(min(min(min(max(abs(gsl_tmp_2 - ease_inQuadratic(gsl_cse_8)) * 10.0 - 0.35, gsl_tmp_3), abs(gsl_tmp_3) - 0.12), min(max(abs(gsl_tmp_2 - ease_outQuadratic(gsl_cse_9)) * 10.0 - 0.35, gsl_tmp_6), abs(gsl_tmp_6) - 0.12)), min(max(abs(gsl_tmp_2 - ease_quadratic(gsl_cse_10)) * 10.0 - 0.35, gsl_tmp_9), abs(gsl_tmp_9) - 0.12)), min(max(abs(gsl_tmp_11 - ease_inCubic(gsl_cse_8)) * 10.0 - 0.35, gsl_tmp_12), abs(gsl_tmp_12) - 0.12)), min(max(abs(gsl_tmp_11 - ease_outCubic(gsl_cse_9)) * 10.0 - 0.35, gsl_tmp_15), abs(gsl_tmp_15) - 0.12)), min(max(abs(gsl_tmp_11 - ease_cubic(gsl_cse_10)) * 10.0 - 0.35, gsl_tmp_18), abs(gsl_tmp_18) - 0.12)), min(max(abs(gsl_tmp_20 - ease_inExponential(gsl_cse_8)) * 10.0 - 0.35, gsl_tmp_21), abs(gsl_tmp_21) - 0.12)), min(max(abs(gsl_tmp_20 - ease_outExponential(gsl_cse_9)) * 10.0 - 0.35, gsl_tmp_24), abs(gsl_tmp_24) - 0.12)), min(max(abs(gsl_tmp_20 - ease_exponential(gsl_cse_10)) * 10.0 - 0.35, gsl_tmp_27), abs(gsl_tmp_27) - 0.12)), min(max(abs(gsl_tmp_29 - ease_inCircular(gsl_cse_8)) * 10.0 - 0.35, gsl_tmp_30), abs(gsl_tmp_30) - 0.12)), min(max(abs(gsl_tmp_29 - ease_outCircular(gsl_cse_9)) * 10.0 - 0.35, gsl_tmp_33), abs(gsl_tmp_33) - 0.12)), min(max(abs(gsl_tmp_29 - ease_circular(gsl_cse_10)) * 10.0 - 0.35, gsl_tmp_36), abs(gsl_tmp_36) - 0.12)), min(max(abs(gsl_tmp_38 - ease_linear(gsl_cse_9)) * 10.0 - 0.35, gsl_tmp_39), abs(gsl_tmp_39) - 0.12));
  let gsl_s1 = gsl_shape(max(grid, abs(p.z) - 0.4), 1u);
  return gsl_s1;
}
