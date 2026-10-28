fn builtin_sphere(p: vec3<f32>, center: vec3<f32>, radius: f32) -> f32 {
  return length(p - center) - radius;
}

fn scan_gsl_anon(q: vec3<f32>, base: vec3<f32>, k: f32, best: vec2<f32>, i: f32) -> vec2<f32> {
  let d = length(q - base - vec3<f32>(i, 0.0, 0.0)) * k;
  return select(best, vec2<f32>(d, i), d < best.x);
}

fn scan(q: vec3<f32>, k: f32) -> vec2<f32> {
  let base = floor(q);
  var gsl_fold = vec2<f32>(1000.0, 0.0);
  for (var gsl_i = 0; gsl_i < 10; gsl_i = gsl_i + 1) {
    gsl_fold = scan_gsl_anon(q, base, k, gsl_fold, f32(gsl_i));
  }
  return gsl_fold;
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let s = scan(p / 10.0, 2.0);
  let gsl_s1 = gsl_shape(builtin_sphere(p, p, 10.0 + s.x), 1u);
  return gsl_s1;
}
