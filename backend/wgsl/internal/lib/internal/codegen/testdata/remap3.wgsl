fn gsl_remap(p: vec3<f32>) -> f32 {
  return sqrt(p.x * p.x + p.y * p.y + p.z * p.z) - 0.8;
}

fn gsl_remap_2(p: vec3<f32>) -> f32 {
  return gsl_remap(vec3<f32>(p.x - 1.0, p.y, p.z));
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let result = gsl_remap_2(vec3<f32>(p.x - 1.0, p.y, p.z));
  let gsl_s1 = gsl_shape(result, 1u);
  return gsl_s1;
}
