fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let placed = min(min(min(1e30, 1e30), 1e30), 1e30);
  let gsl_s1 = gsl_shape(max(placed, length(p) - 10.0), 1u);
  return gsl_s1;
}
