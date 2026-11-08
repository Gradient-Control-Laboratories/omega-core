fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let sum = 0.0 + p.x * 2.0 + p.y * 2.0 + p.z * 2.0 + 2.0;
  let single = p.x + 1.0;
  let gsl_s1 = gsl_shape(sum + single, 1u);
  return gsl_s1;
}
