fn gsl_floored(a: f32, n: f32) -> f32 {
  return a - n * floor(a / n);
}

fn builtin_sphere(p: vec3<f32>, center: vec3<f32>, radius: f32) -> f32 {
  return length(p - center) - radius;
}

fn ripples_gsl_anon(p: vec3<f32>, base: vec3<f32>, size: f32, period: f32, acc: f32, i: f32) -> f32 {
  let cell = base + vec3<f32>(gsl_floored(i, 3.0) - 1.0, floor(i / 3.0) - 1.0, 0.0);
  return acc + sin(length(vec3<f32>(p.x, p.y, 0.0) - cell * size) / period);
}

fn ripples(p: vec3<f32>, size: f32, period: f32) -> f32 {
  let base = floor(vec3<f32>(p.x / size, p.y / size, 0.0));
  var gsl_fold = 0.0;
  for (var gsl_i = 0; gsl_i < 9; gsl_i = gsl_i + 1) {
    gsl_fold = ripples_gsl_anon(p, base, size, period, gsl_fold, f32(gsl_i));
  }
  return gsl_fold;
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let gsl_s1 = gsl_shape(builtin_sphere(p, p, 10.0 + ripples(p, 4.0, 2.0) + ripples(p, 7.0, 3.0)), 1u);
  return gsl_s1;
}
