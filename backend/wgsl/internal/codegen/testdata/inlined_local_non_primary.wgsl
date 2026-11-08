fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let bx = vec3<f32>(0.5, 2.0, 3.0) + p;
  let by = p.y;
  let gsl_s1 = gsl_shape(length(bx) + by, 1u);
  return gsl_s1;
}
