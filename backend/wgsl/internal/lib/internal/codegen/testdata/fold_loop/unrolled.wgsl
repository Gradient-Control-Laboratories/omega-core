fn builtin_sphere(p: vec3<f32>, center: vec3<f32>, radius: f32) -> f32 {
  return length(p - center) - radius;
}

fn few(q: vec3<f32>) -> f32 {
  return 0.0 + length(q - vec3<f32>(0.0, 0.0, 0.0)) + length(q - vec3<f32>(1.0, 0.0, 0.0)) + length(q - vec3<f32>(2.0, 0.0, 0.0));
}

fn indexed(q: vec3<f32>) -> f32 {
  return 0.0 + 1.0 * q.x + 2.0 * q.x + 3.0 * q.x + 4.0 * q.x + 5.0 * q.x + 6.0 * q.x + 7.0 * q.x + 8.0 * q.x + 9.0 * q.x;
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let gsl_s1 = gsl_shape(builtin_sphere(p, p, 10.0 + few(p) + indexed(p)), 1u);
  return gsl_s1;
}
