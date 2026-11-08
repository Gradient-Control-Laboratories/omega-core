fn builtin_sphere(p: vec3<f32>, center: vec3<f32>, radius: f32) -> f32 {
  return length(p - center) - radius;
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let gsl_s1 = gsl_shape(builtin_sphere(p, p, 10.0), 1u);
  return gsl_s1;
}
