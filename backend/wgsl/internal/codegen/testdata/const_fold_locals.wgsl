fn wave(ang: f32) -> f32 {
  let ca = cos(ang);
  let sa = sin(ang);
  return ca * ca + 2.0 * sa;
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let runtime = wave(p.x);
  let gsl_s1 = gsl_shape(length(p) - 0.6228967 - runtime * 0.0, 1u);
  return gsl_s1;
}
