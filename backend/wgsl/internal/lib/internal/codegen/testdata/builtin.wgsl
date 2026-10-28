fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let a = clamp(p.x, p.y, p.z);
  let b = cross(vec3<f32>(a), p);
  let c = tanh(b);
  let gsl_s1 = gsl_shape(c, 1u);
  return gsl_s1;
}
