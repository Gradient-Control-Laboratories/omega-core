fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  let v = vec3<f32>(p.x, p.y, p.z);
  let Mv = mat3x3<f32>(vec3<f32>(2.0, 0.0, 1.0), vec3<f32>(1.0, 1.0, 0.0), vec3<f32>(0.0, 1.0, 1.0)) * v;
  let gsl_s1 = gsl_shape(length(Mv) + 7.0 - 1.0, 1u);
  return gsl_s1;
}
