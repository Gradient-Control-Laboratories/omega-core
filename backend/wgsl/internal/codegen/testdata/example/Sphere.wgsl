fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let gsl_s1 = gsl_shape(sqrt(p.x * p.x + p.y * p.y + p.z * p.z) - 0.8, 1u);
  return gsl_s1;
}
