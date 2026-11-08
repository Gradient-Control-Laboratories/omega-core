fn ease_linear(t: f32) -> f32 {
  return t;
}

fn ease_quadratic(t: f32) -> f32 {
  return t * t * (3.0 - 2.0 * t);
}

fn ease_inQuadratic(t: f32) -> f32 {
  return t * t;
}

fn ease_outQuadratic(t: f32) -> f32 {
  return 1.0 - ease_inQuadratic(1.0 - t);
}

fn ease_cubic(t: f32) -> f32 {
  return t * t * t * (t * (t * 6.0 - 15.0) + 10.0);
}

fn ease_inCubic(t: f32) -> f32 {
  return t * t * t;
}

fn ease_outCubic(t: f32) -> f32 {
  return 1.0 - ease_inCubic(1.0 - t);
}

fn ease_exponential(t: f32) -> f32 {
  let hi = select(1.0 - exp2(-20.0 * t + 10.0) * 0.5, 1.0, t >= 1.0);
  let lo = select(exp2(20.0 * t - 10.0) * 0.5, 0.0, t <= 0.0);
  return select(hi, lo, t < 0.5);
}

fn ease_inExponential(t: f32) -> f32 {
  return select(exp2(10.0 * t - 10.0), 0.0, t <= 0.0);
}

fn ease_outExponential(t: f32) -> f32 {
  return 1.0 - ease_inExponential(1.0 - t);
}

fn ease_circular(t: f32) -> f32 {
  let b = -2.0 * t + 2.0;
  let hi = (sqrt(max(0.0, 1.0 - b * b)) + 1.0) * 0.5;
  let a = 2.0 * t;
  let lo = (1.0 - sqrt(max(0.0, 1.0 - a * a))) * 0.5;
  return select(hi, lo, t < 0.5);
}

fn ease_inCircular(t: f32) -> f32 {
  return 1.0 - sqrt(max(0.0, 1.0 - t * t));
}

fn ease_outCircular(t: f32) -> f32 {
  return 1.0 - ease_inCircular(1.0 - t);
}

fn ease_gcl_closure_interp(gsl_var_value: f32, inMin: f32, inMax: f32, outMin: f32, outMax: f32) -> f32 {
  let t = clamp((gsl_var_value - inMin) / (inMax - inMin), 0.0, 1.0);
  let s = ease_cubic(t);
  return outMin + (outMax - outMin) * s;
}

fn ease_gcl_closure_mix(a: f32, b: f32, t: f32) -> f32 {
  return a + (b - a) * ease_quadratic(t);
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let t = clamp(p.x, 0.0, 1.0);
  let curves = ease_linear(t) + ease_quadratic(t) + ease_inQuadratic(t) + ease_outQuadratic(t) + ease_cubic(t) + ease_inCubic(t) + ease_outCubic(t) + ease_exponential(t) + ease_inExponential(t) + ease_outExponential(t) + ease_circular(t) + ease_inCircular(t) + ease_outCircular(t);
  let remapped = ease_gcl_closure_interp(p.y, 0.0, 10.0, 1.0, 5.0);
  let blended = ease_gcl_closure_mix(1.0, 5.0, t);
  let gsl_s1 = gsl_shape(curves + remapped + blended - length(p), 1u);
  return gsl_s1;
}
