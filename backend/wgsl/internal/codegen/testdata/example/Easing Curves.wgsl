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
  let gsl_tmp_3 = max(abs(gsl_tmp_1 - 0.5) - 0.5, abs(gsl_tmp_2 - 0.5) - 0.5) * 10.0;
  let gsl_tmp_5 = (p.y - -0.0) / 10.0;
  let gsl_tmp_4 = (p.x - 13.0) / 10.0;
  let gsl_tmp_6 = max(abs(gsl_tmp_4 - 0.5) - 0.5, abs(gsl_tmp_5 - 0.5) - 0.5) * 10.0;
  let gsl_tmp_8 = (p.y - -0.0) / 10.0;
  let gsl_tmp_7 = (p.x - 26.0) / 10.0;
  let gsl_tmp_9 = max(abs(gsl_tmp_7 - 0.5) - 0.5, abs(gsl_tmp_8 - 0.5) - 0.5) * 10.0;
  let gsl_tmp_11 = (p.y - -13.0) / 10.0;
  let gsl_tmp_10 = (p.x - 0.0) / 10.0;
  let gsl_tmp_12 = max(abs(gsl_tmp_10 - 0.5) - 0.5, abs(gsl_tmp_11 - 0.5) - 0.5) * 10.0;
  let gsl_tmp_14 = (p.y - -13.0) / 10.0;
  let gsl_tmp_13 = (p.x - 13.0) / 10.0;
  let gsl_tmp_15 = max(abs(gsl_tmp_13 - 0.5) - 0.5, abs(gsl_tmp_14 - 0.5) - 0.5) * 10.0;
  let gsl_tmp_17 = (p.y - -13.0) / 10.0;
  let gsl_tmp_16 = (p.x - 26.0) / 10.0;
  let gsl_tmp_18 = max(abs(gsl_tmp_16 - 0.5) - 0.5, abs(gsl_tmp_17 - 0.5) - 0.5) * 10.0;
  let gsl_tmp_20 = (p.y - -26.0) / 10.0;
  let gsl_tmp_19 = (p.x - 0.0) / 10.0;
  let gsl_tmp_21 = max(abs(gsl_tmp_19 - 0.5) - 0.5, abs(gsl_tmp_20 - 0.5) - 0.5) * 10.0;
  let gsl_tmp_23 = (p.y - -26.0) / 10.0;
  let gsl_tmp_22 = (p.x - 13.0) / 10.0;
  let gsl_tmp_24 = max(abs(gsl_tmp_22 - 0.5) - 0.5, abs(gsl_tmp_23 - 0.5) - 0.5) * 10.0;
  let gsl_tmp_26 = (p.y - -26.0) / 10.0;
  let gsl_tmp_25 = (p.x - 26.0) / 10.0;
  let gsl_tmp_27 = max(abs(gsl_tmp_25 - 0.5) - 0.5, abs(gsl_tmp_26 - 0.5) - 0.5) * 10.0;
  let gsl_tmp_29 = (p.y - -39.0) / 10.0;
  let gsl_tmp_28 = (p.x - 0.0) / 10.0;
  let gsl_tmp_30 = max(abs(gsl_tmp_28 - 0.5) - 0.5, abs(gsl_tmp_29 - 0.5) - 0.5) * 10.0;
  let gsl_tmp_32 = (p.y - -39.0) / 10.0;
  let gsl_tmp_31 = (p.x - 13.0) / 10.0;
  let gsl_tmp_33 = max(abs(gsl_tmp_31 - 0.5) - 0.5, abs(gsl_tmp_32 - 0.5) - 0.5) * 10.0;
  let gsl_tmp_35 = (p.y - -39.0) / 10.0;
  let gsl_tmp_34 = (p.x - 26.0) / 10.0;
  let gsl_tmp_36 = max(abs(gsl_tmp_34 - 0.5) - 0.5, abs(gsl_tmp_35 - 0.5) - 0.5) * 10.0;
  let gsl_tmp_38 = (p.y - -52.0) / 10.0;
  let gsl_tmp_37 = (p.x - 13.0) / 10.0;
  let gsl_tmp_39 = max(abs(gsl_tmp_37 - 0.5) - 0.5, abs(gsl_tmp_38 - 0.5) - 0.5) * 10.0;
  let grid = min(min(min(min(min(min(min(min(min(min(min(min(min(max(abs(gsl_tmp_2 - ease_inQuadratic(clamp(gsl_tmp_1, 0.0, 1.0))) * 10.0 - 0.35, gsl_tmp_3), abs(gsl_tmp_3) - 0.12), min(max(abs(gsl_tmp_5 - ease_outQuadratic(clamp(gsl_tmp_4, 0.0, 1.0))) * 10.0 - 0.35, gsl_tmp_6), abs(gsl_tmp_6) - 0.12)), min(max(abs(gsl_tmp_8 - ease_quadratic(clamp(gsl_tmp_7, 0.0, 1.0))) * 10.0 - 0.35, gsl_tmp_9), abs(gsl_tmp_9) - 0.12)), min(max(abs(gsl_tmp_11 - ease_inCubic(clamp(gsl_tmp_10, 0.0, 1.0))) * 10.0 - 0.35, gsl_tmp_12), abs(gsl_tmp_12) - 0.12)), min(max(abs(gsl_tmp_14 - ease_outCubic(clamp(gsl_tmp_13, 0.0, 1.0))) * 10.0 - 0.35, gsl_tmp_15), abs(gsl_tmp_15) - 0.12)), min(max(abs(gsl_tmp_17 - ease_cubic(clamp(gsl_tmp_16, 0.0, 1.0))) * 10.0 - 0.35, gsl_tmp_18), abs(gsl_tmp_18) - 0.12)), min(max(abs(gsl_tmp_20 - ease_inExponential(clamp(gsl_tmp_19, 0.0, 1.0))) * 10.0 - 0.35, gsl_tmp_21), abs(gsl_tmp_21) - 0.12)), min(max(abs(gsl_tmp_23 - ease_outExponential(clamp(gsl_tmp_22, 0.0, 1.0))) * 10.0 - 0.35, gsl_tmp_24), abs(gsl_tmp_24) - 0.12)), min(max(abs(gsl_tmp_26 - ease_exponential(clamp(gsl_tmp_25, 0.0, 1.0))) * 10.0 - 0.35, gsl_tmp_27), abs(gsl_tmp_27) - 0.12)), min(max(abs(gsl_tmp_29 - ease_inCircular(clamp(gsl_tmp_28, 0.0, 1.0))) * 10.0 - 0.35, gsl_tmp_30), abs(gsl_tmp_30) - 0.12)), min(max(abs(gsl_tmp_32 - ease_outCircular(clamp(gsl_tmp_31, 0.0, 1.0))) * 10.0 - 0.35, gsl_tmp_33), abs(gsl_tmp_33) - 0.12)), min(max(abs(gsl_tmp_35 - ease_circular(clamp(gsl_tmp_34, 0.0, 1.0))) * 10.0 - 0.35, gsl_tmp_36), abs(gsl_tmp_36) - 0.12)), min(max(abs(gsl_tmp_38 - ease_linear(clamp(gsl_tmp_37, 0.0, 1.0))) * 10.0 - 0.35, gsl_tmp_39), abs(gsl_tmp_39) - 0.12));
  let gsl_s1 = gsl_shape(max(grid, abs(p.z) - 0.4), 1u);
  return gsl_s1;
}
