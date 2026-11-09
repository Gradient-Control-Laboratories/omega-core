fn builtin_sphere(p: vec3<f32>, center: vec3<f32>, radius: f32) -> f32 {
  return length(p - center) - radius;
}

fn gcl_closure_apply(k: f32, a: f32, v: f32) -> f32 {
  return v * k + a;
}

fn gcl_closure_apply_2(k: f32, x: f32) -> f32 {
  return gcl_closure_apply(k, 2.0 * k, x);
}

fn grow(q: f32, k: f32) -> f32 {
  let b = gcl_closure_apply_2(k, q);
  return b;
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let gsl_s1 = gsl_shape(builtin_sphere(p, p, grow(5.0, 0.1 * p.z + 0.5)), 1u);
  return gsl_s1;
}
