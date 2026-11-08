fn gsl_sphere_remap(p: vec3<f32>) -> f32 {
  return length(p) - 2.5;
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let gsl_s1 = gsl_shape(gsl_sphere_remap(p - vec3<f32>(1.0, 2.0, 3.0)), 1u);
  return gsl_s1;
}
