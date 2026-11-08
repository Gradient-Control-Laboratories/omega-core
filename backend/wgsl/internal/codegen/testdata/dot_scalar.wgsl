fn builtin_sphere(p: vec3<f32>, radius: f32) -> f32 {
  return length(p) - radius;
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let runtime = p.x * p.y;
  let mixed = 4.0 * p.z;
  let gsl_s1 = gsl_shape(builtin_sphere(p, 6.0 + runtime + mixed), 1u);
  return gsl_s1;
}
